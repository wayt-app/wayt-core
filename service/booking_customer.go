package service

import (
	"errors"

	"github.com/wayt-app/wayt-core/model"
)

func (s *bookingService) WaitingListPosition(bookingID uint) (int64, error) {
	b, err := s.repo.FindByID(bookingID)
	if err != nil {
		return 0, errors.New("booking tidak ditemukan")
	}
	if b.Status != model.BookingStatusWaitingList {
		return 0, errors.New("booking tidak dalam waiting list")
	}
	pos, err := s.repo.CountWaitingListBefore(bookingID, b.BranchID, b.BookingDate, b.StartTime)
	if err != nil {
		return 0, err
	}
	return pos + 1, nil // 1-indexed position
}

func (s *bookingService) MyBookings(customerID uint) ([]model.Booking, error) {
	return s.repo.FindByCustomer(customerID)
}

func (s *bookingService) MyBookingsPaged(customerID uint, sortBy, sortDir string, page, limit int) (*BookingPage, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	offset := (page - 1) * limit
	data, total, err := s.repo.FindByCustomerPaged(customerID, sortBy, sortDir, offset, limit)
	if err != nil {
		return nil, err
	}
	totalPages := (total + int64(limit) - 1) / int64(limit)
	return &BookingPage{Data: data, Total: total, Page: page, Limit: limit, TotalPages: totalPages}, nil
}

func (s *bookingService) GetByID(id uint) (*model.Booking, error) {
	b, err := s.repo.FindByID(id)
	if err != nil {
		return nil, errors.New("booking tidak ditemukan")
	}
	return b, nil
}

func (s *bookingService) Cancel(id uint, customerID uint, reason string) error {
	b, err := s.repo.FindByID(id)
	if err != nil {
		return errors.New("booking tidak ditemukan")
	}
	if b.CustomerID != customerID {
		return errors.New("tidak diizinkan membatalkan booking ini")
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
	// Trigger auto-promote for waiting list
	if b.Status == model.BookingStatusPending || b.Status == model.BookingStatusConfirmed || b.Status == model.BookingStatusCheckedIn {
		_ = s.autoPromote(b)
	}
	return nil
}

func (s *bookingService) Reschedule(bookingID, customerID uint, dateStr, startTime string) (*model.Booking, error) {
	b, err := s.repo.FindByID(bookingID)
	if err != nil {
		return nil, errors.New("booking tidak ditemukan")
	}
	if b.CustomerID != customerID {
		return nil, errors.New("tidak diizinkan mengubah booking ini")
	}
	if b.Status != model.BookingStatusPending {
		return nil, errors.New("hanya booking dengan status pending yang dapat diubah jadwalnya")
	}

	// H-1: booking date must be at least tomorrow
	if !b.BookingDate.After(today()) {
		return nil, errors.New("jadwal tidak dapat diubah, sudah melewati batas H-1")
	}

	newDate, err := parseDate(dateStr)
	if err != nil {
		return nil, errors.New("format tanggal tidak valid, gunakan YYYY-MM-DD")
	}
	// New date must be tomorrow at minimum (H-1 dari sekarang)
	if !newDate.After(today()) {
		return nil, errors.New("tanggal baru harus minimal H-1 (besok atau lebih)")
	}
	if newDate.After(today().AddDate(0, 0, 30)) {
		return nil, errors.New("booking maksimal 30 hari ke depan")
	}
	if err := validateTime(startTime); err != nil {
		return nil, err
	}

	branch, err := s.branchRepo.FindByID(b.BranchID)
	if err != nil {
		return nil, errors.New("cabang tidak ditemukan")
	}
	endTime := addMinutes(startTime, branch.DefaultDurationMinutes)

	if len(b.BookingTables) > 0 {
		// Multi-table booking: check each assigned table type independently.
		for _, bt := range b.BookingTables {
			btt, err := s.tableTypeRepo.FindByID(bt.TableTypeID)
			if err != nil {
				return nil, errors.New("tipe meja tidak ditemukan")
			}
			if !btt.IsActive {
				return nil, errors.New("tipe meja tidak lagi aktif, silakan hubungi restoran untuk bantuan lebih lanjut")
			}
			grpTables, _ := s.tableTypeRepo.FindByGroup(btt.BranchID, btt.Name, btt.Capacity, btt.RoomID)
			grpIDs := make([]uint, len(grpTables))
			for i, t := range grpTables {
				grpIDs[i] = t.ID
			}
			grpBooked, _ := s.repo.CountOverlappingByGroup(grpIDs, newDate, startTime, endTime, bookingID)
			if grpBooked+int64(bt.Count) > int64(len(grpTables)) {
				return nil, errors.New("slot baru tidak tersedia, silakan pilih waktu lain")
			}
		}
	} else {
		// Single table type booking: check anchor group.
		tt, err := s.tableTypeRepo.FindByID(b.TableTypeID)
		if err != nil {
			return nil, errors.New("tipe meja tidak ditemukan")
		}
		if !tt.IsActive {
			return nil, errors.New("tipe meja tidak lagi aktif, silakan hubungi restoran untuk bantuan lebih lanjut")
		}
		rGrpTables, _ := s.tableTypeRepo.FindByGroup(tt.BranchID, tt.Name, tt.Capacity, tt.RoomID)
		rGrpIDs := make([]uint, len(rGrpTables))
		for i, t := range rGrpTables {
			rGrpIDs[i] = t.ID
		}
		rBooked, _ := s.repo.CountOverlappingByGroup(rGrpIDs, newDate, startTime, endTime, bookingID)
		if rBooked+int64(b.TablesCount) > int64(len(rGrpTables)) {
			return nil, errors.New("slot baru tidak tersedia, silakan pilih waktu lain")
		}
	}

	if err := s.repo.UpdateSchedule(bookingID, newDate, startTime, endTime, model.BookingStatusPending); err != nil {
		return nil, err
	}

	// Reload booking with associations for notifications
	updated, _ := s.repo.FindByID(bookingID)
	if updated == nil {
		updated = b
		updated.BookingDate = newDate
		updated.StartTime = startTime
		updated.EndTime = endTime
		updated.Status = model.BookingStatusPending
	}

	go s.sendRescheduleNotif(updated, b.BookingDate, b.StartTime)
	return updated, nil
}

func (s *bookingService) UploadPaymentProof(bookingID, customerID uint, proofURL string) error {
	b, err := s.repo.FindByID(bookingID)
	if err != nil {
		return errors.New("booking tidak ditemukan")
	}
	if b.CustomerID != customerID {
		return errors.New("akses tidak diizinkan")
	}
	if proofURL == "" {
		return errors.New("URL bukti transfer tidak boleh kosong")
	}
	return s.repo.UpdatePaymentProof(bookingID, proofURL)
}
