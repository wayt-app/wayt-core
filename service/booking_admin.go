package service

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/wayt-app/wayt-core/model"
)

func (s *bookingService) ListByBranch(branchID uint, dateStr string, status *model.BookingStatus) ([]model.Booking, error) {
	var date *time.Time
	if dateStr != "" {
		d, err := parseDate(dateStr)
		if err != nil {
			return nil, errors.New("format tanggal tidak valid")
		}
		date = &d
	}
	return s.repo.FindByBranch(branchID, date, status)
}

func (s *bookingService) ListByBranchPaged(branchID uint, dateStr string, status *model.BookingStatus, search, sortBy, sortDir string, page, limit int) (*BookingPage, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	var date *time.Time
	if dateStr != "" {
		d, err := parseDate(dateStr)
		if err != nil {
			return nil, errors.New("format tanggal tidak valid")
		}
		date = &d
	}
	offset := (page - 1) * limit
	data, total, err := s.repo.FindByBranchPaged(branchID, date, status, search, sortBy, sortDir, offset, limit)
	if err != nil {
		return nil, err
	}
	totalPages := (total + int64(limit) - 1) / int64(limit)
	return &BookingPage{Data: data, Total: total, Page: page, Limit: limit, TotalPages: totalPages}, nil
}

func (s *bookingService) Confirm(id uint) error {
	b, err := s.repo.FindByID(id)
	if err != nil {
		return errors.New("booking tidak ditemukan")
	}
	if b.Status != model.BookingStatusPending {
		return errors.New("hanya booking dengan status pending yang bisa dikonfirmasi")
	}
	if err := s.repo.UpdateStatus(id, model.BookingStatusConfirmed); err != nil {
		return err
	}
	go s.sendBookingNotif(b, "confirmed")
	go s.sendBookingEmail(b, "confirmed")
	go s.sendInAppNotif(b, "confirmed")
	return nil
}

func (s *bookingService) Complete(id uint, notes string, totalBill int64) error {
	b, err := s.repo.FindByID(id)
	if err != nil {
		return errors.New("booking tidak ditemukan")
	}
	if b.Status != model.BookingStatusConfirmed && b.Status != model.BookingStatusCheckedIn {
		return errors.New("hanya booking confirmed atau checked_in yang bisa diselesaikan")
	}
	if totalBill < 0 {
		return errors.New("total biaya tidak boleh negatif")
	}
	return s.repo.CompleteWithDetails(id, strings.TrimSpace(notes), totalBill)
}

func (s *bookingService) AdminCancel(id uint, reason string) error {
	b, err := s.repo.FindByID(id)
	if err != nil {
		return errors.New("booking tidak ditemukan")
	}
	if b.Status == model.BookingStatusCompleted || b.Status == model.BookingStatusCancelled {
		return errors.New("booking sudah tidak bisa dibatalkan")
	}
	if err := s.repo.UpdateStatusAndReason(id, model.BookingStatusCancelled, reason); err != nil {
		return err
	}
	b.CancelReason = reason
	go s.sendBookingNotif(b, "cancelled")
	go s.sendBookingEmail(b, "cancelled")
	go s.sendInAppNotif(b, "cancelled")
	if b.Status == model.BookingStatusPending || b.Status == model.BookingStatusConfirmed || b.Status == model.BookingStatusCheckedIn {
		_ = s.autoPromote(b)
	}
	return nil
}

func (s *bookingService) AdminUpdate(id uint, dateStr, startTime, notes string, guestCount int) (*model.Booking, error) {
	b, err := s.repo.FindByID(id)
	if err != nil {
		return nil, errors.New("booking tidak ditemukan")
	}
	if b.Status != model.BookingStatusPending && b.Status != model.BookingStatusConfirmed && b.Status != model.BookingStatusCheckedIn && b.Status != model.BookingStatusWaitingList {
		return nil, errors.New("hanya booking yang aktif yang bisa diedit")
	}
	newDate, err := parseDate(dateStr)
	if err != nil {
		return nil, errors.New("format tanggal tidak valid, gunakan YYYY-MM-DD")
	}
	if err := validateTime(startTime); err != nil {
		return nil, err
	}
	if guestCount <= 0 {
		return nil, errors.New("jumlah tamu harus lebih dari 0")
	}
	branch, err := s.branchRepo.FindByID(b.BranchID)
	if err != nil {
		return nil, errors.New("cabang tidak ditemukan")
	}
	endTime := addMinutes(startTime, branch.DefaultDurationMinutes)
	tt, err := s.tableTypeRepo.FindByID(b.TableTypeID)
	if err != nil {
		return nil, errors.New("tipe meja tidak ditemukan")
	}
	tablesCount := (guestCount + tt.Capacity - 1) / tt.Capacity
	grpTables, _ := s.tableTypeRepo.FindByGroup(tt.BranchID, tt.Name, tt.Capacity, tt.RoomID)
	grpIDs := make([]uint, len(grpTables))
	for i, t := range grpTables { grpIDs[i] = t.ID }
	if tablesCount > len(grpTables) { tablesCount = len(grpTables) }
	booked, err := s.repo.CountOverlappingByGroup(grpIDs, newDate, startTime, endTime, id)
	if err != nil {
		return nil, err
	}
	if booked+int64(tablesCount) > int64(len(grpTables)) {
		return nil, errors.New("slot yang dipilih sudah penuh")
	}
	if err := s.repo.UpdateDetails(id, newDate, startTime, endTime, guestCount, notes); err != nil {
		return nil, err
	}
	return s.repo.FindByID(id)
}

func (s *bookingService) AdminChangeTableType(bookingID, tableTypeID uint) (*model.Booking, error) {
	b, err := s.repo.FindByID(bookingID)
	if err != nil {
		return nil, errors.New("booking tidak ditemukan")
	}
	if b.Status != model.BookingStatusPending && b.Status != model.BookingStatusConfirmed && b.Status != model.BookingStatusCheckedIn && b.Status != model.BookingStatusWaitingList {
		return nil, errors.New("hanya booking yang aktif yang bisa diubah mejanya")
	}
	tt, err := s.tableTypeRepo.FindByID(tableTypeID)
	if err != nil {
		return nil, errors.New("tipe meja tidak ditemukan")
	}
	if tt.BranchID != b.BranchID {
		return nil, errors.New("tipe meja tidak tersedia di cabang ini")
	}
	if !tt.IsActive {
		return nil, errors.New("tipe meja tidak aktif")
	}
	tablesCount := (b.GuestCount + tt.Capacity - 1) / tt.Capacity
	grpTables2, _ := s.tableTypeRepo.FindByGroup(tt.BranchID, tt.Name, tt.Capacity, tt.RoomID)
	grpIDs2 := make([]uint, len(grpTables2))
	for i, t := range grpTables2 { grpIDs2[i] = t.ID }
	if tablesCount > len(grpTables2) { tablesCount = len(grpTables2) }
	booked, err := s.repo.CountOverlappingByGroup(grpIDs2, b.BookingDate, b.StartTime, b.EndTime, bookingID)
	if err != nil {
		return nil, err
	}
	if booked+int64(tablesCount) > int64(len(grpTables2)) {
		return nil, errors.New("tipe meja yang dipilih sudah penuh di slot ini")
	}
	if err := s.repo.UpdateTableTypeWithCount(bookingID, tableTypeID, tablesCount, tt.RoomID); err != nil {
		return nil, err
	}
	// Rebuild booking_tables: replace old assignment with new single-type assignment
	if err := s.bookingTableRepo.DeleteByBooking(bookingID); err != nil {
		log.Printf("[WARN] ChangeTable booking #%d: gagal hapus booking_tables lama: %v", bookingID, err)
	}
	if err := s.bookingTableRepo.CreateBatch([]model.BookingTable{
		{BookingID: bookingID, TableTypeID: tableTypeID, Count: tablesCount},
	}); err != nil {
		return nil, fmt.Errorf("gagal memperbarui alokasi meja: %w", err)
	}
	return s.repo.FindByID(bookingID)
}

func (s *bookingService) GetTableChangeOptions(bookingID uint) ([]TableOption, error) {
	b, err := s.repo.FindByID(bookingID)
	if err != nil {
		return nil, errors.New("booking tidak ditemukan")
	}
	allTypes, err := s.tableTypeRepo.FindByBranchWithRoom(b.BranchID)
	if err != nil {
		return nil, err
	}

	var activeIDs []uint
	for _, tt := range allTypes {
		if tt.IsActive {
			activeIDs = append(activeIDs, tt.ID)
		}
	}
	bookedMap, _ := s.bookingTableRepo.SumByTypeIDsAndSlot(activeIDs, b.BookingDate, b.StartTime, b.EndTime, bookingID)

	var result []TableOption
	for _, tt := range allTypes {
		if !tt.IsActive {
			continue
		}
		opt := TableOption{
			TableTypeID: tt.ID,
			Name:        tt.Name,
			Capacity:    tt.Capacity,
			RoomID:      tt.RoomID,
			IsAvailable: bookedMap[tt.ID] == 0,
		}
		if tt.Room != nil {
			opt.RoomName = tt.Room.Name
		}
		result = append(result, opt)
	}
	return result, nil
}

func (s *bookingService) AdminChangeTables(bookingID uint, tableTypeIDs []uint) (*model.Booking, error) {
	if len(tableTypeIDs) == 0 {
		return nil, errors.New("pilih minimal 1 meja")
	}
	b, err := s.repo.FindByID(bookingID)
	if err != nil {
		return nil, errors.New("booking tidak ditemukan")
	}
	if b.Status != model.BookingStatusPending && b.Status != model.BookingStatusConfirmed &&
		b.Status != model.BookingStatusCheckedIn && b.Status != model.BookingStatusWaitingList {
		return nil, errors.New("hanya booking yang aktif yang bisa diubah mejanya")
	}
	bookedMap, _ := s.bookingTableRepo.SumByTypeIDsAndSlot(tableTypeIDs, b.BookingDate, b.StartTime, b.EndTime, bookingID)

	var anchorID uint
	var anchorRoomID *uint
	for i, id := range tableTypeIDs {
		tt, err := s.tableTypeRepo.FindByID(id)
		if err != nil {
			return nil, fmt.Errorf("meja ID %d tidak ditemukan", id)
		}
		if tt.BranchID != b.BranchID {
			return nil, errors.New("meja tidak tersedia di cabang ini")
		}
		if !tt.IsActive {
			return nil, fmt.Errorf("meja '%s' tidak aktif", tt.Name)
		}
		if bookedMap[id] > 0 {
			return nil, fmt.Errorf("meja '%s' sudah dipesan di jam ini", tt.Name)
		}
		if i == 0 {
			anchorID = id
			anchorRoomID = tt.RoomID
		}
	}
	bookingTables := make([]model.BookingTable, len(tableTypeIDs))
	for i, id := range tableTypeIDs {
		bookingTables[i] = model.BookingTable{BookingID: bookingID, TableTypeID: id, Count: 1}
	}

	if err := s.repo.UpdateTableTypeWithCount(bookingID, anchorID, len(tableTypeIDs), anchorRoomID); err != nil {
		return nil, err
	}
	if err := s.bookingTableRepo.DeleteByBooking(bookingID); err != nil {
		log.Printf("[WARN] ChangeTableGroup booking #%d: gagal hapus booking_tables lama: %v", bookingID, err)
	}
	if err := s.bookingTableRepo.CreateBatch(bookingTables); err != nil {
		return nil, fmt.Errorf("gagal memperbarui alokasi meja: %w", err)
	}
	return s.repo.FindByID(bookingID)
}

func (s *bookingService) CheckIn(id uint) error {
	b, err := s.repo.FindByID(id)
	if err != nil {
		return errors.New("booking tidak ditemukan")
	}
	if b.Status != model.BookingStatusConfirmed {
		return errors.New("hanya booking confirmed yang bisa check-in")
	}
	if err := s.repo.UpdateStatus(id, model.BookingStatusCheckedIn); err != nil {
		return err
	}
	go s.sendBookingEmail(b, "checked_in")
	return nil
}

func (s *bookingService) UpdateOrderStatus(bookingID, branchID uint, status string) error {
	if !validOrderStatuses[status] {
		return errors.New("status tidak valid, gunakan: new, prepare, ready, done")
	}
	b, err := s.repo.FindByID(bookingID)
	if err != nil {
		return errors.New("booking tidak ditemukan")
	}
	if b.BranchID != branchID {
		return errors.New("akses tidak diizinkan")
	}
	return s.repo.UpdateOrderStatus(bookingID, status)
}
