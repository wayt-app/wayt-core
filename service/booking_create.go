package service

import (
	"errors"
	"fmt"

	"github.com/wayt-app/wayt-core/model"
)

func (s *bookingService) Create(customerID, branchID, tableTypeID uint, dateStr, startTime string, guestCount int, notes, menuOrder, source, guestEmail string) (*model.Booking, error) {
	branch, err := s.branchRepo.FindByID(branchID)
	if err != nil {
		return nil, errors.New("cabang tidak ditemukan")
	}
	if !branch.IsActive {
		return nil, errors.New("cabang tidak aktif")
	}

	tt, err := s.tableTypeRepo.FindByID(tableTypeID)
	if err != nil {
		return nil, errors.New("tipe meja tidak ditemukan")
	}
	if tt.BranchID != branchID {
		return nil, errors.New("tipe meja tidak tersedia di cabang ini")
	}
	if !tt.IsActive {
		return nil, errors.New("tipe meja tidak aktif")
	}
	if guestCount <= 0 {
		return nil, errors.New("jumlah tamu harus lebih dari 0")
	}
	// Calculate how many tables are needed.
	tablesCount := (guestCount + tt.Capacity - 1) / tt.Capacity // ceil division

	// Load all physical tables in the same group (same name+capacity+room).
	groupTables, err := s.tableTypeRepo.FindByGroup(tt.BranchID, tt.Name, tt.Capacity, tt.RoomID)
	if err != nil || len(groupTables) == 0 {
		return nil, errors.New("tipe meja tidak ditemukan")
	}
	groupIDs := make([]uint, len(groupTables))
	for i, t := range groupTables {
		groupIDs[i] = t.ID
	}
	groupTotal := len(groupTables)
	if tablesCount > groupTotal {
		tablesCount = groupTotal
	}

	date, err := parseDate(dateStr)
	if err != nil {
		return nil, errors.New("format tanggal tidak valid, gunakan YYYY-MM-DD")
	}
	if date.Before(today()) {
		return nil, errors.New("tanggal tidak boleh di masa lalu")
	}
	if date.After(today().AddDate(0, 0, 30)) {
		return nil, errors.New("booking maksimal 30 hari ke depan")
	}
	if err := validateTime(startTime); err != nil {
		return nil, err
	}
	if err := checkMinBookingHours(date, startTime, branch.MinBookingHours); err != nil {
		return nil, err
	}

	endTime := addMinutes(startTime, branch.DefaultDurationMinutes)

	// Check subscription limit (only for restaurants owned by a business owner)
	isOverLimit := false
	if s.subRepo != nil {
		sub, subErr := s.subRepo.FindByRestaurantID(branch.RestaurantID)
		if subErr == nil && sub != nil {
			if sub.Status != model.SubscriptionStatusActive && sub.Status != model.SubscriptionStatusTrial {
				return nil, errors.New("restoran tidak memiliki langganan aktif")
			}
			if sub.Plan != nil && sub.Plan.MaxReservationsPerMonth != -1 &&
				sub.ReservationsThisMonth >= sub.Plan.MaxReservationsPerMonth {
				isOverLimit = true
			}
		}
	}

	// Check availability against the full group
	booked, err := s.repo.CountOverlappingByGroup(groupIDs, date, startTime, endTime, 0)
	if err != nil {
		return nil, err
	}

	var status model.BookingStatus
	if booked+int64(tablesCount) > int64(groupTotal) {
		// Not enough tables available — put customer on waiting list
		status = model.BookingStatusWaitingList
	} else {
		// Determine initial status based on branch config
		status = model.BookingStatusConfirmed
		if branch.RequireConfirmation {
			status = model.BookingStatusPending
		}
	}

	if source == "" {
		source = "app"
	}
	b := &model.Booking{
		CustomerID:  customerID,
		BranchID:    branchID,
		TableTypeID: tableTypeID,
		RoomID:      tt.RoomID,
		BookingDate: date,
		StartTime:   startTime,
		EndTime:     endTime,
		GuestCount:  guestCount,
		TablesCount: tablesCount,
		Status:      status,
		Notes:       notes,
		MenuOrder:   menuOrder,
		IsOverLimit: isOverLimit,
		Source:      source,
		GuestEmail:  guestEmail,
	}
	if err := s.repo.Create(b); err != nil {
		return nil, err
	}
	// Record assignment in booking_tables (single-type)
	if err := s.bookingTableRepo.CreateBatch([]model.BookingTable{
		{BookingID: b.ID, TableTypeID: tableTypeID, Count: tablesCount},
	}); err != nil {
		_ = s.repo.Delete(b.ID)
		return nil, fmt.Errorf("gagal mencatat alokasi meja: %w", err)
	}
	go s.sendBookingNotif(b, string(status))
	go s.sendBookingEmail(b, string(status))
	go s.sendOwnerBookingEmail(b, string(status))
	go s.sendInAppNotif(b, string(status))
	// Increment reservation count + check warning threshold (best-effort)
	if s.reservIncr != nil {
		go func(restaurantID uint) {
			rest, err := s.restaurantRepo.FindByID(restaurantID)
			if err == nil && rest != nil && rest.BusinessOwnerID != nil {
				_ = s.reservIncr.IncrementReservation(*rest.BusinessOwnerID)
			}
		}(branch.RestaurantID)
	}
	return b, nil
}

func (s *bookingService) CreateFromRoom(customerID, branchID uint, roomID *uint, dateStr, startTime string, guestCount int, notes, menuOrder, source, guestEmail string) (*model.Booking, error) {
	branch, err := s.branchRepo.FindByID(branchID)
	if err != nil {
		return nil, errors.New("cabang tidak ditemukan")
	}
	if !branch.IsActive {
		return nil, errors.New("cabang tidak aktif")
	}
	if guestCount <= 0 {
		return nil, errors.New("jumlah tamu harus lebih dari 0")
	}

	date, err := parseDate(dateStr)
	if err != nil {
		return nil, errors.New("format tanggal tidak valid, gunakan YYYY-MM-DD")
	}
	if date.Before(today()) {
		return nil, errors.New("tanggal tidak boleh di masa lalu")
	}
	if date.After(today().AddDate(0, 0, 30)) {
		return nil, errors.New("booking maksimal 30 hari ke depan")
	}
	if err := validateTime(startTime); err != nil {
		return nil, err
	}
	if err := checkMinBookingHours(date, startTime, branch.MinBookingHours); err != nil {
		return nil, err
	}
	endTime := addMinutes(startTime, branch.DefaultDurationMinutes)

	// Load all active tables in the room
	roomTables, err := s.tableTypeRepo.FindByBranchAndRoom(branchID, roomID)
	if err != nil || len(roomTables) == 0 {
		return nil, errors.New("tidak ada meja tersedia di ruangan ini")
	}
	var activeTables []model.TableType
	for _, t := range roomTables {
		if t.IsActive {
			activeTables = append(activeTables, t)
		}
	}

	// Get booked counts per table type for this slot
	allIDs := make([]uint, len(activeTables))
	for i, t := range activeTables {
		allIDs[i] = t.ID
	}
	bookedMap, err := s.bookingTableRepo.SumByTypeIDsAndSlot(allIDs, date, startTime, endTime, 0)
	if err != nil {
		bookedMap = map[uint]int64{}
	}

	// Build availability per table type
	var avails []typeAvail
	for _, t := range activeTables {
		free := 1 - int(bookedMap[t.ID])
		if free > 0 {
			avails = append(avails, typeAvail{id: t.ID, name: t.Name, capacity: t.Capacity, count: free})
		}
	}

	// Compute optimal assignment
	assign := computeOptimalAssignment(guestCount, avails)

	// Subscription limit check
	isOverLimit := false
	if s.subRepo != nil {
		sub, subErr := s.subRepo.FindByRestaurantID(branch.RestaurantID)
		if subErr == nil && sub != nil {
			if sub.Status != model.SubscriptionStatusActive && sub.Status != model.SubscriptionStatusTrial {
				return nil, errors.New("restoran tidak memiliki langganan aktif")
			}
			if sub.Plan != nil && sub.Plan.MaxReservationsPerMonth != -1 &&
				sub.ReservationsThisMonth >= sub.Plan.MaxReservationsPerMonth {
				isOverLimit = true
			}
		}
	}

	var status model.BookingStatus
	var anchorTableTypeID uint
	var totalTables int
	var bookingTables []model.BookingTable

	if assign == nil {
		// No optimal assignment possible → waiting list using smallest available table
		status = model.BookingStatusWaitingList
		if len(activeTables) > 0 {
			anchorTableTypeID = activeTables[0].ID
			totalTables = 1
			bookingTables = []model.BookingTable{{TableTypeID: anchorTableTypeID, Count: 1}}
		} else {
			return nil, errors.New("tidak ada meja tersedia")
		}
	} else {
		status = model.BookingStatusConfirmed
		if branch.RequireConfirmation {
			status = model.BookingStatusPending
		}
		anchorTableTypeID = assign[0].id
		for _, a := range assign {
			totalTables += a.count
			bookingTables = append(bookingTables, model.BookingTable{TableTypeID: a.id, Count: a.count})
		}
	}

	if source == "" {
		source = "app"
	}
	b := &model.Booking{
		CustomerID:  customerID,
		BranchID:    branchID,
		TableTypeID: anchorTableTypeID,
		RoomID:      roomID,
		BookingDate: date,
		StartTime:   startTime,
		EndTime:     endTime,
		GuestCount:  guestCount,
		TablesCount: totalTables,
		Status:      status,
		Notes:       notes,
		MenuOrder:   menuOrder,
		IsOverLimit: isOverLimit,
		Source:      source,
		GuestEmail:  guestEmail,
	}
	if err := s.repo.Create(b); err != nil {
		return nil, err
	}
	for i := range bookingTables {
		bookingTables[i].BookingID = b.ID
	}
	if err := s.bookingTableRepo.CreateBatch(bookingTables); err != nil {
		_ = s.repo.Delete(b.ID)
		return nil, fmt.Errorf("gagal mencatat alokasi meja: %w", err)
	}

	go s.sendBookingNotif(b, string(status))
	go s.sendBookingEmail(b, string(status))
	go s.sendOwnerBookingEmail(b, string(status))
	go s.sendInAppNotif(b, string(status))
	if s.reservIncr != nil {
		go func(restaurantID uint) {
			rest, err := s.restaurantRepo.FindByID(restaurantID)
			if err != nil || rest == nil || rest.BusinessOwnerID == nil {
				return
			}
			_ = s.reservIncr.IncrementReservation(*rest.BusinessOwnerID)
		}(branch.RestaurantID)
	}
	return b, nil
}
