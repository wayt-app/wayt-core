package service

import (
	"errors"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/wayt-app/wayt-core/model"
	"github.com/wayt-app/wayt-core/repository"
	"github.com/wayt-app/wayt-core/pkg/email"
	"github.com/wayt-app/wayt-core/pkg/whatsapp"
)

type TableTypeStatus struct {
	TableTypeID uint   `json:"table_type_id"`
	Name        string `json:"name"`
	Capacity    int    `json:"capacity"`
	TotalTables int    `json:"total_tables"`
	Booked      int64  `json:"booked"`
	Available   int64  `json:"available"`
}

type TableStatusResult struct {
	Date      string            `json:"date"`
	StartTime string            `json:"start_time"`
	EndTime   string            `json:"end_time"`
	Tables    []TableTypeStatus `json:"tables"`
}

type BranchDashboard struct {
	BranchID   uint             `json:"branch_id"`
	BranchName string           `json:"branch_name"`
	Counts     map[string]int64 `json:"counts"`
	Total      int64            `json:"total"`
}

type RestaurantDashboard struct {
	Date     string            `json:"date"`
	Branches []BranchDashboard `json:"branches"`
}

type DashboardStats struct {
	Date        string            `json:"date"`
	BranchID    uint              `json:"branch_id"`
	Counts      map[string]int64  `json:"counts"`
	TotalToday  int64             `json:"total_today"`
	NoShowCount int64             `json:"no_show_count"`
}

type BookingPage struct {
	Data       []model.Booking `json:"data"`
	Total      int64           `json:"total"`
	Page       int             `json:"page"`
	Limit      int             `json:"limit"`
	TotalPages int64           `json:"total_pages"`
}

type AvailabilityResult struct {
	TableTypeID   uint   `json:"table_type_id"`
	Name          string `json:"name"`
	Capacity      int    `json:"capacity"`
	TotalTables   int    `json:"total_tables"`
	BookedCount   int64  `json:"booked_count"`
	Available     int64  `json:"available"`
	TablesNeeded  int    `json:"tables_needed"`
	StartTime     string `json:"start_time"`
	EndTime       string `json:"end_time"`
}

type CustomerSummary struct {
	CustomerID  uint      `json:"customer_id"`
	Name        string    `json:"name"`
	Email       string    `json:"email"`
	Phone       string    `json:"phone"`
	TotalVisits int64     `json:"total_visits"`
	LastVisit   time.Time `json:"last_visit"`
	TotalSpend  int64     `json:"total_spend"`
	Segment     string    `json:"segment"` // "new", "regular", "loyal"
}

type BookingService interface {
	CheckAvailability(branchID uint, dateStr, startTime string, guests int, roomID *uint) ([]AvailabilityResult, error)
	Create(customerID, branchID, tableTypeID uint, dateStr, startTime string, guestCount int, notes, menuOrder, source, guestEmail string) (*model.Booking, error)
	// CreateFromRoom books by choosing a room; the system computes the optimal multi-type assignment.
	CreateFromRoom(customerID, branchID uint, roomID *uint, dateStr, startTime string, guestCount int, notes, menuOrder, source, guestEmail string) (*model.Booking, error)
	MyBookings(customerID uint) ([]model.Booking, error)
	GetByID(id uint) (*model.Booking, error)
	Cancel(id uint, customerID uint, reason string) error
	WaitingListPosition(bookingID uint) (int64, error)
	ProcessNoShows() error
	// Dashboard
	GetDashboardStats(branchID uint, dateStr string) (*DashboardStats, error)
	GetTableStatus(branchID uint, dateStr, startTime string) (*TableStatusResult, error)
	GetRestaurantDashboard(restaurantID uint, dateStr string) (*RestaurantDashboard, error)
	// Check-in
	CheckIn(id uint) error
	// Reschedule changes date/time for a pending or confirmed booking (customer only, H-1).
	Reschedule(bookingID, customerID uint, dateStr, startTime string) (*model.Booking, error)
	// Admin actions
	ListByBranch(branchID uint, dateStr string, status *model.BookingStatus) ([]model.Booking, error)
	ListByBranchPaged(branchID uint, dateStr string, status *model.BookingStatus, search, sortBy, sortDir string, page, limit int) (*BookingPage, error)
	Confirm(id uint) error
	Complete(id uint, notes string, totalBill int64) error
	AdminCancel(id uint, reason string) error
	// AdminUpdate edits booking date/time/guest_count/notes for active bookings.
	AdminUpdate(id uint, dateStr, startTime, notes string, guestCount int) (*model.Booking, error)
	// AdminChangeTableType reassigns the table type for an active booking.
	AdminChangeTableType(bookingID, tableTypeID uint) (*model.Booking, error)
	// MyBookingsPaged returns a paginated list of a customer's bookings.
	MyBookingsPaged(customerID uint, sortBy, sortDir string, page, limit int) (*BookingPage, error)
	// ProcessReminders sends H-1 reminder notifications for tomorrow's bookings.
	ProcessReminders() error
	// ListCustomersByRestaurant returns aggregated customer data for a restaurant for retention purposes.
	// Sorted by last_visit ASC (longest-absent first).
	ListCustomersByRestaurant(restaurantID uint) ([]CustomerSummary, error)
	// UpdateOrderStatus changes the pre-order kitchen status (new/prepare/ready/done).
	UpdateOrderStatus(bookingID, branchID uint, status string) error
}

// ReservationIncrementer is a narrow interface so bookingService can trigger
// the full increment-with-warning-check without importing businessOwnerService directly.
type ReservationIncrementer interface {
	IncrementReservation(ownerID uint) error
}

type bookingService struct {
	repo             repository.BookingRepository
	branchRepo       repository.BranchRepository
	tableTypeRepo    repository.TableTypeRepository
	bookingTableRepo repository.BookingTableRepository
	customerRepo     repository.CustomerRepository
	subRepo          repository.SubscriptionRepository
	restaurantRepo   repository.RestaurantRepository
	staffRepo        repository.StaffRepository
	ownerRepo        repository.BusinessOwnerRepository
	waSender         whatsapp.Sender
	emailSender      email.Sender
	notifSvc         NotificationService
	reservIncr       ReservationIncrementer
	emailConfigRepo  repository.EmailConfigRepository
}

func NewBookingService(
	repo repository.BookingRepository,
	branchRepo repository.BranchRepository,
	tableTypeRepo repository.TableTypeRepository,
	bookingTableRepo repository.BookingTableRepository,
	customerRepo repository.CustomerRepository,
	subRepo repository.SubscriptionRepository,
	restaurantRepo repository.RestaurantRepository,
	staffRepo repository.StaffRepository,
	ownerRepo repository.BusinessOwnerRepository,
	waSender whatsapp.Sender,
	emailSender email.Sender,
	notifSvc NotificationService,
	reservIncr ReservationIncrementer,
	emailConfigRepo repository.EmailConfigRepository,
) BookingService {
	return &bookingService{
		repo:             repo,
		branchRepo:       branchRepo,
		tableTypeRepo:    tableTypeRepo,
		bookingTableRepo: bookingTableRepo,
		customerRepo:     customerRepo,
		subRepo:          subRepo,
		restaurantRepo:   restaurantRepo,
		staffRepo:        staffRepo,
		ownerRepo:        ownerRepo,
		waSender:         waSender,
		emailSender:      emailSender,
		notifSvc:         notifSvc,
		reservIncr:       reservIncr,
		emailConfigRepo:  emailConfigRepo,
	}
}

func (s *bookingService) CheckAvailability(branchID uint, dateStr, startTime string, guests int, roomID *uint) ([]AvailabilityResult, error) {
	branch, err := s.branchRepo.FindByID(branchID)
	if err != nil {
		return nil, errors.New("cabang tidak ditemukan")
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

	endTime := addMinutes(startTime, branch.DefaultDurationMinutes)
	tableTypes, err := s.tableTypeRepo.FindByBranchAndRoom(branchID, roomID)
	if err != nil {
		return nil, err
	}

	// Group physical tables by (name, capacity, room_id)
	type grp struct {
		rep      uint
		name     string
		capacity int
		roomID   *uint
		ids      []uint
	}
	groups := map[string]*grp{}
	var groupOrder []string
	for _, tt := range tableTypes {
		if !tt.IsActive {
			continue
		}
		key := tableGroupKey(tt.Capacity, tt.RoomID)
		if _, ok := groups[key]; !ok {
			groups[key] = &grp{rep: tt.ID, name: tt.Name, capacity: tt.Capacity, roomID: tt.RoomID}
			groupOrder = append(groupOrder, key)
		}
		groups[key].ids = append(groups[key].ids, tt.ID)
	}

	var results []AvailabilityResult
	for _, key := range groupOrder {
		g := groups[key]
		totalTables := len(g.ids)
		tablesNeeded := 1
		if guests > 0 {
			tablesNeeded = (guests + g.capacity - 1) / g.capacity
		}
		booked, err := s.repo.CountOverlappingByGroup(g.ids, date, startTime, endTime, 0)
		if err != nil {
			booked = 0
		}
		available := int64(totalTables) - booked
		results = append(results, AvailabilityResult{
			TableTypeID:  g.rep,
			Name:         g.name,
			Capacity:     g.capacity,
			TotalTables:  totalTables,
			BookedCount:  booked,
			Available:    available,
			TablesNeeded: tablesNeeded,
			StartTime:    startTime,
			EndTime:      endTime,
		})
	}
	return results, nil
}

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
	_ = s.bookingTableRepo.CreateBatch([]model.BookingTable{
		{BookingID: b.ID, TableTypeID: tableTypeID, Count: tablesCount},
	})
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
	_ = s.bookingTableRepo.CreateBatch(bookingTables)

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
	branch, err := s.branchRepo.FindByID(b.BranchID)
	if err != nil {
		return nil, errors.New("cabang tidak ditemukan")
	}
	endTime := addMinutes(b.StartTime, branch.DefaultDurationMinutes)
	grpTables2, _ := s.tableTypeRepo.FindByGroup(tt.BranchID, tt.Name, tt.Capacity, tt.RoomID)
	grpIDs2 := make([]uint, len(grpTables2))
	for i, t := range grpTables2 { grpIDs2[i] = t.ID }
	if tablesCount > len(grpTables2) { tablesCount = len(grpTables2) }
	booked, err := s.repo.CountOverlappingByGroup(grpIDs2, b.BookingDate, b.StartTime, endTime, bookingID)
	if err != nil {
		return nil, err
	}
	if booked+int64(tablesCount) > int64(len(grpTables2)) {
		return nil, errors.New("tipe meja yang dipilih sudah penuh di slot ini")
	}
	if err := s.repo.UpdateTableType(bookingID, tableTypeID); err != nil {
		return nil, err
	}
	return s.repo.FindByID(bookingID)
}

// ProcessNoShows marks confirmed bookings as no_show if they started more than 15 minutes ago,
// then triggers auto-promote for each freed slot.
func (s *bookingService) ProcessNoShows() error {
	candidates, err := s.repo.FindNoShowCandidates(15)
	if err != nil {
		return err
	}
	for _, b := range candidates {
		if err := s.repo.UpdateStatus(b.ID, model.BookingStatusNoShow); err != nil {
			continue
		}
		_ = s.autoPromote(&b)
	}
	return nil
}

// autoPromote promotes the earliest waiting list entry for the same slot when a booking is cancelled.
func (s *bookingService) autoPromote(cancelled *model.Booking) error {
	branch, err := s.branchRepo.FindByID(cancelled.BranchID)
	if err != nil {
		return nil
	}
	waiters, err := s.repo.FindWaitingListForSlot(cancelled.BranchID, cancelled.BookingDate, cancelled.StartTime, cancelled.EndTime)
	if err != nil || len(waiters) == 0 {
		return nil
	}
	tableTypes, err := s.tableTypeRepo.FindByBranch(cancelled.BranchID)
	if err != nil {
		return nil
	}
	sort.Slice(tableTypes, func(i, j int) bool {
		return tableTypes[i].Capacity < tableTypes[j].Capacity
	})

	for _, waiter := range waiters {
		tablesNeeded := waiter.TablesCount
		if tablesNeeded <= 0 {
			// Fallback for old bookings without tables_count
			tablesNeeded = 1
		}
		// Find a table type group that can fit this waiter
		for _, tt := range tableTypes {
			if !tt.IsActive {
				continue
			}
			needed := (waiter.GuestCount + tt.Capacity - 1) / tt.Capacity
			grpT, _ := s.tableTypeRepo.FindByGroup(tt.BranchID, tt.Name, tt.Capacity, tt.RoomID)
			if len(grpT) == 0 || needed > len(grpT) {
				continue
			}
			grpI := make([]uint, len(grpT))
			for i, t := range grpT { grpI[i] = t.ID }
			booked, _ := s.repo.CountOverlappingByGroup(grpI, cancelled.BookingDate, cancelled.StartTime, cancelled.EndTime, waiter.ID)
			if booked+int64(needed) <= int64(len(grpT)) {
				// Assign this table type and promote
				_ = s.repo.UpdateTableType(waiter.ID, tt.ID)
				status := model.BookingStatusConfirmed
				if branch.RequireConfirmation {
					status = model.BookingStatusPending
				}
				_ = s.repo.UpdateStatus(waiter.ID, status)
				waiter.TableTypeID = tt.ID
				waiter.EndTime = cancelled.EndTime
				go s.sendBookingNotif(&waiter, "promoted")
				go s.sendBookingEmail(&waiter, "promoted")
				go s.sendInAppNotif(&waiter, "promoted")
				return nil // promote one at a time
			}
		}
	}
	return nil
}

func (s *bookingService) GetTableStatus(branchID uint, dateStr, startTime string) (*TableStatusResult, error) {
	branch, err := s.branchRepo.FindByID(branchID)
	if err != nil {
		return nil, errors.New("cabang tidak ditemukan")
	}
	if dateStr == "" {
		dateStr = today().Format("2006-01-02")
	}
	date, err := parseDate(dateStr)
	if err != nil {
		return nil, errors.New("format tanggal tidak valid")
	}
	if err := validateTime(startTime); err != nil {
		return nil, err
	}
	endTime := addMinutes(startTime, branch.DefaultDurationMinutes)

	bookedMap, err := s.repo.CountOverlappingByTableType(branchID, date, startTime, endTime)
	if err != nil {
		return nil, err
	}

	tableTypes, err := s.tableTypeRepo.FindByBranch(branchID)
	if err != nil {
		return nil, err
	}

	// Group physical tables; sum booked per group
	type tsGrp struct {
		rep      uint
		name     string
		capacity int
		ids      []uint
	}
	tsGroups := map[string]*tsGrp{}
	var tsOrder []string
	for _, tt := range tableTypes {
		if !tt.IsActive {
			continue
		}
		key := tableGroupKey(tt.Capacity, tt.RoomID)
		if _, ok := tsGroups[key]; !ok {
			tsGroups[key] = &tsGrp{rep: tt.ID, name: tt.Name, capacity: tt.Capacity}
			tsOrder = append(tsOrder, key)
		}
		tsGroups[key].ids = append(tsGroups[key].ids, tt.ID)
	}
	var tables []TableTypeStatus
	for _, key := range tsOrder {
		g := tsGroups[key]
		var booked int64
		for _, id := range g.ids {
			booked += bookedMap[id]
		}
		totalTables := int64(len(g.ids))
		available := totalTables - booked
		if available < 0 {
			available = 0
		}
		tables = append(tables, TableTypeStatus{
			TableTypeID: g.rep,
			Name:        g.name,
			Capacity:    g.capacity,
			TotalTables: int(totalTables),
			Booked:      booked,
			Available:   available,
		})
	}

	return &TableStatusResult{
		Date:      dateStr,
		StartTime: startTime,
		EndTime:   endTime,
		Tables:    tables,
	}, nil
}

func (s *bookingService) GetRestaurantDashboard(restaurantID uint, dateStr string) (*RestaurantDashboard, error) {
	if dateStr == "" {
		dateStr = today().Format("2006-01-02")
	}
	date, err := parseDate(dateStr)
	if err != nil {
		return nil, errors.New("format tanggal tidak valid")
	}

	branches, err := s.branchRepo.FindByRestaurant(restaurantID)
	if err != nil {
		return nil, err
	}

	var result []BranchDashboard
	for _, b := range branches {
		counts, err := s.repo.CountByStatusForDate(b.ID, date)
		if err != nil {
			continue
		}
		mapped := make(map[string]int64)
		var total int64
		for status, count := range counts {
			mapped[string(status)] = count
			total += count
		}
		result = append(result, BranchDashboard{
			BranchID:   b.ID,
			BranchName: b.Name,
			Counts:     mapped,
			Total:      total,
		})
	}

	return &RestaurantDashboard{Date: dateStr, Branches: result}, nil
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

func (s *bookingService) GetDashboardStats(branchID uint, dateStr string) (*DashboardStats, error) {
	if dateStr == "" {
		dateStr = today().Format("2006-01-02")
	}
	date, err := parseDate(dateStr)
	if err != nil {
		return nil, errors.New("format tanggal tidak valid")
	}
	counts, err := s.repo.CountByStatusForDate(branchID, date)
	if err != nil {
		return nil, err
	}
	mapped := make(map[string]int64)
	var total int64
	for status, count := range counts {
		mapped[string(status)] = count
		total += count
	}
	noShow := mapped[string(model.BookingStatusNoShow)]
	return &DashboardStats{
		Date:        dateStr,
		BranchID:    branchID,
		Counts:      mapped,
		TotalToday:  total,
		NoShowCount: noShow,
	}, nil
}

// sendBookingNotif sends a WhatsApp notification to the customer for a booking event.
// Errors are ignored — notifications are best-effort.
func (s *bookingService) sendBookingNotif(b *model.Booking, event string) {
	customer, err := s.customerRepo.FindByID(b.CustomerID)
	if err != nil || customer.Phone == "" {
		return
	}
	branch, err := s.branchRepo.FindByID(b.BranchID)
	if err != nil {
		return
	}

	dateStr := b.BookingDate.Format("02 Jan 2006")
	var msg string
	switch event {
	case "confirmed":
		msg = fmt.Sprintf(
			"Halo *%s*! 🎉\n\nBooking Anda di *%s* telah *dikonfirmasi*.\n\n📅 %s\n⏰ %s – %s\n👥 %d tamu\n🔖 ID: #%d\n\nSampai jumpa!",
			customer.Name, branch.Name, dateStr, b.StartTime, b.EndTime, b.GuestCount, b.ID,
		)
	case "pending":
		msg = fmt.Sprintf(
			"Halo *%s*! 📋\n\nBooking Anda di *%s* sedang *menunggu konfirmasi* admin.\n\n📅 %s\n⏰ %s – %s\n👥 %d tamu\n🔖 ID: #%d\n\nKami akan segera mengonfirmasi.",
			customer.Name, branch.Name, dateStr, b.StartTime, b.EndTime, b.GuestCount, b.ID,
		)
	case "waiting_list":
		msg = fmt.Sprintf(
			"Halo *%s*! ⏳\n\nMaaf, slot penuh. Anda telah masuk *waiting list* di *%s*.\n\n📅 %s\n⏰ %s – %s\n👥 %d tamu\n🔖 ID: #%d\n\nKami akan memberitahu jika ada slot kosong.",
			customer.Name, branch.Name, dateStr, b.StartTime, b.EndTime, b.GuestCount, b.ID,
		)
	case "promoted":
		msg = fmt.Sprintf(
			"Halo *%s*! 🎊\n\nKabar baik! Slot tersedia dan booking Anda di *%s* telah *dikonfirmasi*.\n\n📅 %s\n⏰ %s – %s\n👥 %d tamu\n🔖 ID: #%d\n\nSampai jumpa!",
			customer.Name, branch.Name, dateStr, b.StartTime, b.EndTime, b.GuestCount, b.ID,
		)
	case "cancelled":
		msg = fmt.Sprintf(
			"Halo *%s*,\n\nBooking Anda #%d di *%s* pada %s pukul %s telah *dibatalkan*.\n\nTerima kasih.",
			customer.Name, b.ID, branch.Name, dateStr, b.StartTime,
		)
	default:
		return
	}

	if err := s.waSender.Send(customer.Phone, msg); err != nil {
		log.Printf("[WA ERROR] booking #%d ke %s: %v", b.ID, customer.Phone, err)
	}
}

// sendBookingEmail sends an HTML email notification for a booking event.
// Errors are ignored — notifications are best-effort.
func (s *bookingService) sendBookingEmail(b *model.Booking, event string) {
	customer, err := s.customerRepo.FindByID(b.CustomerID)
	if err != nil {
		return
	}
	recipientEmail := customer.Email
	if b.GuestEmail != "" {
		recipientEmail = b.GuestEmail
	}
	if recipientEmail == "" {
		return
	}

	var emailCfg *model.EmailConfig
	if s.emailConfigRepo != nil {
		emailCfg, _ = s.emailConfigRepo.Get()
	}
	branch, err := s.branchRepo.FindByID(b.BranchID)
	if err != nil {
		return
	}

	dateStr := b.BookingDate.Format("02 Jan 2006")

	var subject, body string
	switch event {
	case "confirmed":
		subject = fmt.Sprintf("Booking #%d Dikonfirmasi — %s", b.ID, branch.Name)
		body = fmt.Sprintf(`
<p>Halo <strong>%s</strong>,</p>
<p>Booking Anda telah <strong style="color:#16a34a">dikonfirmasi</strong>. Berikut detailnya:</p>
<table style="border-collapse:collapse;margin:16px 0">
  <tr><td style="padding:4px 12px 4px 0;color:#6b7280">Restoran</td><td><strong>%s</strong></td></tr>
  <tr><td style="padding:4px 12px 4px 0;color:#6b7280">Tanggal</td><td>%s</td></tr>
  <tr><td style="padding:4px 12px 4px 0;color:#6b7280">Waktu</td><td>%s – %s</td></tr>
  <tr><td style="padding:4px 12px 4px 0;color:#6b7280">Tamu</td><td>%d orang</td></tr>
  <tr><td style="padding:4px 12px 4px 0;color:#6b7280">ID Booking</td><td>#%d</td></tr>
</table>
<p>Sampai jumpa!</p>`,
			customer.Name, branch.Name, dateStr, b.StartTime, b.EndTime, b.GuestCount, b.ID)

	case "pending":
		subject = fmt.Sprintf("Booking #%d Menunggu Konfirmasi — %s", b.ID, branch.Name)
		body = fmt.Sprintf(`
<p>Halo <strong>%s</strong>,</p>
<p>Booking Anda sedang <strong>menunggu konfirmasi</strong> dari admin. Kami akan segera mengonfirmasi.</p>
<table style="border-collapse:collapse;margin:16px 0">
  <tr><td style="padding:4px 12px 4px 0;color:#6b7280">Restoran</td><td><strong>%s</strong></td></tr>
  <tr><td style="padding:4px 12px 4px 0;color:#6b7280">Tanggal</td><td>%s</td></tr>
  <tr><td style="padding:4px 12px 4px 0;color:#6b7280">Waktu</td><td>%s – %s</td></tr>
  <tr><td style="padding:4px 12px 4px 0;color:#6b7280">Tamu</td><td>%d orang</td></tr>
  <tr><td style="padding:4px 12px 4px 0;color:#6b7280">ID Booking</td><td>#%d</td></tr>
</table>`,
			customer.Name, branch.Name, dateStr, b.StartTime, b.EndTime, b.GuestCount, b.ID)

	case "waiting_list":
		subject = fmt.Sprintf("Booking #%d Masuk Waiting List — %s", b.ID, branch.Name)
		body = fmt.Sprintf(`
<p>Halo <strong>%s</strong>,</p>
<p>Maaf, slot yang Anda pilih sedang penuh. Anda telah masuk <strong>waiting list</strong>.</p>
<table style="border-collapse:collapse;margin:16px 0">
  <tr><td style="padding:4px 12px 4px 0;color:#6b7280">Restoran</td><td><strong>%s</strong></td></tr>
  <tr><td style="padding:4px 12px 4px 0;color:#6b7280">Tanggal</td><td>%s</td></tr>
  <tr><td style="padding:4px 12px 4px 0;color:#6b7280">Waktu</td><td>%s – %s</td></tr>
  <tr><td style="padding:4px 12px 4px 0;color:#6b7280">Tamu</td><td>%d orang</td></tr>
  <tr><td style="padding:4px 12px 4px 0;color:#6b7280">ID Booking</td><td>#%d</td></tr>
</table>
<p>Kami akan mengirim email jika ada slot kosong untuk Anda.</p>`,
			customer.Name, branch.Name, dateStr, b.StartTime, b.EndTime, b.GuestCount, b.ID)

	case "promoted":
		subject = fmt.Sprintf("Booking #%d Dikonfirmasi dari Waiting List — %s", b.ID, branch.Name)
		body = fmt.Sprintf(`
<p>Halo <strong>%s</strong>,</p>
<p>Kabar baik! Slot tersedia dan booking Anda dari waiting list telah <strong style="color:#16a34a">dikonfirmasi</strong>.</p>
<table style="border-collapse:collapse;margin:16px 0">
  <tr><td style="padding:4px 12px 4px 0;color:#6b7280">Restoran</td><td><strong>%s</strong></td></tr>
  <tr><td style="padding:4px 12px 4px 0;color:#6b7280">Tanggal</td><td>%s</td></tr>
  <tr><td style="padding:4px 12px 4px 0;color:#6b7280">Waktu</td><td>%s – %s</td></tr>
  <tr><td style="padding:4px 12px 4px 0;color:#6b7280">Tamu</td><td>%d orang</td></tr>
  <tr><td style="padding:4px 12px 4px 0;color:#6b7280">ID Booking</td><td>#%d</td></tr>
</table>
<p>Sampai jumpa!</p>`,
			customer.Name, branch.Name, dateStr, b.StartTime, b.EndTime, b.GuestCount, b.ID)

	case "checked_in":
		subject = fmt.Sprintf("Check-in Berhasil — %s", branch.Name)
		body = fmt.Sprintf(`
<p>Halo <strong>%s</strong>,</p>
<p>Anda telah berhasil <strong style="color:#7c3aed">check-in</strong> di <strong>%s</strong>. Selamat menikmati!</p>
<table style="border-collapse:collapse;margin:16px 0">
  <tr><td style="padding:4px 12px 4px 0;color:#6b7280">Restoran</td><td><strong>%s</strong></td></tr>
  <tr><td style="padding:4px 12px 4px 0;color:#6b7280">Tanggal</td><td>%s</td></tr>
  <tr><td style="padding:4px 12px 4px 0;color:#6b7280">Waktu</td><td>%s – %s</td></tr>
  <tr><td style="padding:4px 12px 4px 0;color:#6b7280">Tamu</td><td>%d orang</td></tr>
  <tr><td style="padding:4px 12px 4px 0;color:#6b7280">ID Booking</td><td>#%d</td></tr>
</table>
<p>Terima kasih telah memilih kami. Semoga pengalaman makan Anda menyenangkan! 🍽️</p>`,
			customer.Name, branch.Name, branch.Name, dateStr, b.StartTime, b.EndTime, b.GuestCount, b.ID)

	case "cancelled":
		subject = fmt.Sprintf("Booking #%d Dibatalkan — %s", b.ID, branch.Name)
		reasonRow := ""
		if b.CancelReason != "" {
			reasonRow = fmt.Sprintf(`<tr><td style="padding:4px 12px 4px 0;color:#6b7280">Alasan</td><td>%s</td></tr>`, b.CancelReason)
		}
		// Determine contact phone: prefer restaurant phone, fall back to branch phone
		contactPhone := branch.Phone
		if restaurant, err := s.restaurantRepo.FindByID(branch.RestaurantID); err == nil && restaurant.Phone != "" {
			contactPhone = restaurant.Phone
		}
		phoneInfo := ""
		if contactPhone != "" {
			phoneInfo = fmt.Sprintf(`<p>Jika ada pertanyaan, silakan hubungi kami di <strong>%s</strong>.</p>`, contactPhone)
		} else {
			phoneInfo = `<p>Jika ada pertanyaan, hubungi kami langsung.</p>`
		}
		body = fmt.Sprintf(`
<p>Halo <strong>%s</strong>,</p>
<p>Booking Anda telah <strong style="color:#dc2626">dibatalkan</strong> oleh restoran.</p>
<table style="border-collapse:collapse;margin:16px 0">
  <tr><td style="padding:4px 12px 4px 0;color:#6b7280">Restoran</td><td><strong>%s</strong></td></tr>
  <tr><td style="padding:4px 12px 4px 0;color:#6b7280">Tanggal</td><td>%s</td></tr>
  <tr><td style="padding:4px 12px 4px 0;color:#6b7280">Waktu</td><td>%s</td></tr>
  <tr><td style="padding:4px 12px 4px 0;color:#6b7280">ID Booking</td><td>#%d</td></tr>
  %s
</table>
%s`,
			customer.Name, branch.Name, dateStr, b.StartTime, b.ID, reasonRow, phoneInfo)

	default:
		return
	}

	var restLogoURL string
	if rest, rerr := s.restaurantRepo.FindByID(branch.RestaurantID); rerr == nil {
		restLogoURL = rest.LogoURL
	}
	wrapped := wrapEmailHTML(body, emailCfg, customer.Name, branch.Name, restLogoURL)
	if err := s.emailSender.Send(recipientEmail, subject, wrapped); err != nil {
		log.Printf("[EMAIL ERROR] booking #%d ke %s: %v", b.ID, recipientEmail, err)
	}
}

// sendOwnerBookingEmail notifies the restaurant owner when a customer creates a new reservation.
// Only fires for confirmed, pending, and waiting_list events. Errors are ignored — best-effort.
func (s *bookingService) sendOwnerBookingEmail(b *model.Booking, event string) {
	if event != string(model.BookingStatusConfirmed) &&
		event != string(model.BookingStatusPending) &&
		event != string(model.BookingStatusWaitingList) {
		return
	}
	if s.ownerRepo == nil {
		return
	}

	branch, err := s.branchRepo.FindByID(b.BranchID)
	if err != nil {
		return
	}
	restaurant, err := s.restaurantRepo.FindByID(branch.RestaurantID)
	if err != nil || restaurant.BusinessOwnerID == nil {
		return
	}
	owner, err := s.ownerRepo.FindByID(*restaurant.BusinessOwnerID)
	if err != nil || owner.Email == "" {
		return
	}
	customer, err := s.customerRepo.FindByID(b.CustomerID)
	if err != nil {
		return
	}

	dateStr := b.BookingDate.Format("02 Jan 2006")

	var statusLabel string
	switch event {
	case string(model.BookingStatusConfirmed):
		statusLabel = `<span style="color:#16a34a;font-weight:bold">Terkonfirmasi Otomatis</span>`
	case string(model.BookingStatusPending):
		statusLabel = `<span style="color:#d97706;font-weight:bold">Menunggu Konfirmasi Anda</span>`
	case string(model.BookingStatusWaitingList):
		statusLabel = `<span style="color:#6b7280;font-weight:bold">Waiting List</span>`
	}

	notesRow := ""
	if b.Notes != "" {
		notesRow = fmt.Sprintf(`<tr><td style="padding:4px 12px 4px 0;color:#6b7280">Catatan</td><td>%s</td></tr>`, b.Notes)
	}

	subject := fmt.Sprintf("Reservasi Baru #%d — %s", b.ID, branch.Name)
	body := fmt.Sprintf(`
<p>Halo <strong>%s</strong>,</p>
<p>Ada reservasi baru masuk di cabang <strong>%s</strong>. Status: %s</p>
<table style="border-collapse:collapse;margin:16px 0">
  <tr><td style="padding:4px 12px 4px 0;color:#6b7280">Pelanggan</td><td><strong>%s</strong></td></tr>
  <tr><td style="padding:4px 12px 4px 0;color:#6b7280">No. HP</td><td>%s</td></tr>
  <tr><td style="padding:4px 12px 4px 0;color:#6b7280">Tanggal</td><td>%s</td></tr>
  <tr><td style="padding:4px 12px 4px 0;color:#6b7280">Waktu</td><td>%s – %s</td></tr>
  <tr><td style="padding:4px 12px 4px 0;color:#6b7280">Tamu</td><td>%d orang</td></tr>
  <tr><td style="padding:4px 12px 4px 0;color:#6b7280">ID Booking</td><td>#%d</td></tr>
  %s
</table>`,
		owner.Name, branch.Name, statusLabel,
		customer.Name, customer.Phone,
		dateStr, b.StartTime, b.EndTime, b.GuestCount, b.ID,
		notesRow)

	var emailCfg *model.EmailConfig
	if s.emailConfigRepo != nil {
		emailCfg, _ = s.emailConfigRepo.Get()
	}
	var restLogoURL string
	if rest, rerr := s.restaurantRepo.FindByID(branch.RestaurantID); rerr == nil {
		restLogoURL = rest.LogoURL
	}
	wrapped := wrapEmailHTML(body, emailCfg, owner.Name, branch.Name, restLogoURL)
	if err := s.emailSender.Send(owner.Email, subject, wrapped); err != nil {
		log.Printf("[EMAIL ERROR] notif owner booking #%d ke %s: %v", b.ID, owner.Email, err)
	}
}

// sendInAppNotif creates a DB notification and pushes it via SSE.
func (s *bookingService) sendInAppNotif(b *model.Booking, event string) {
	if s.notifSvc == nil {
		return
	}
	branch, _ := s.branchRepo.FindByID(b.BranchID)
	branchName := ""
	if branch != nil {
		branchName = branch.Name
	}
	dateStr := b.BookingDate.Format("02 Jan 2006")

	var customerTitle, customerMsg string
	switch event {
	case string(model.BookingStatusPending):
		customerTitle = "Booking Menunggu Konfirmasi"
		customerMsg = fmt.Sprintf("Booking #%d di %s pada %s pukul %s menunggu konfirmasi.", b.ID, branchName, dateStr, b.StartTime)
	case string(model.BookingStatusConfirmed):
		customerTitle = "Booking Dikonfirmasi"
		customerMsg = fmt.Sprintf("Booking #%d di %s pada %s pukul %s telah dikonfirmasi.", b.ID, branchName, dateStr, b.StartTime)
	case string(model.BookingStatusWaitingList):
		customerTitle = "Masuk Waiting List"
		customerMsg = fmt.Sprintf("Booking #%d di %s masuk waiting list untuk %s pukul %s.", b.ID, branchName, dateStr, b.StartTime)
	case "promoted":
		customerTitle = "Naik dari Waiting List"
		customerMsg = fmt.Sprintf("Booking #%d di %s berhasil mendapat tempat untuk %s pukul %s.", b.ID, branchName, dateStr, b.StartTime)
	case "cancelled":
		customerTitle = "Booking Dibatalkan"
		customerMsg = fmt.Sprintf("Booking #%d di %s pada %s pukul %s telah dibatalkan.", b.ID, branchName, dateStr, b.StartTime)
	default:
		return
	}

	// Notify customer
	_ = s.notifSvc.Send("customer", b.CustomerID, customerTitle, customerMsg)

	// Notify owner: new booking or cancellation
	if event == string(model.BookingStatusPending) || event == string(model.BookingStatusConfirmed) || event == string(model.BookingStatusWaitingList) {
		customer, _ := s.customerRepo.FindByID(b.CustomerID)
		customerName := ""
		if customer != nil {
			customerName = customer.Name
		}
		if branch != nil {
			rest, _ := s.restaurantRepo.FindByBranchID(branch.RestaurantID)
			if rest != nil && rest.BusinessOwnerID != nil {
				ownerMsg := fmt.Sprintf("Booking baru #%d dari %s di %s pada %s pukul %s.", b.ID, customerName, branchName, dateStr, b.StartTime)
				_ = s.notifSvc.Send("owner", *rest.BusinessOwnerID, "Booking Baru", ownerMsg)
			}
		}
	}
	if event == "cancelled" {
		customer, _ := s.customerRepo.FindByID(b.CustomerID)
		customerName := ""
		if customer != nil {
			customerName = customer.Name
		}
		if branch != nil {
			rest, _ := s.restaurantRepo.FindByBranchID(branch.RestaurantID)
			if rest != nil && rest.BusinessOwnerID != nil {
				ownerMsg := fmt.Sprintf("Booking #%d dari %s di %s pada %s dibatalkan.", b.ID, customerName, branchName, dateStr)
				_ = s.notifSvc.Send("owner", *rest.BusinessOwnerID, "Booking Dibatalkan", ownerMsg)
			}
		}
	}
}

func (s *bookingService) Reschedule(bookingID, customerID uint, dateStr, startTime string) (*model.Booking, error) {
	b, err := s.repo.FindByID(bookingID)
	if err != nil {
		return nil, errors.New("booking tidak ditemukan")
	}
	if b.CustomerID != customerID {
		return nil, errors.New("tidak diizinkan mengubah booking ini")
	}
	if b.Status != model.BookingStatusPending && b.Status != model.BookingStatusConfirmed {
		return nil, errors.New("hanya booking dengan status pending atau confirmed yang dapat diubah jadwalnya")
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

	tt, err := s.tableTypeRepo.FindByID(b.TableTypeID)
	if err != nil {
		return nil, errors.New("tipe meja tidak ditemukan")
	}
	if !tt.IsActive {
		return nil, errors.New("tipe meja tidak lagi aktif, silakan hubungi restoran untuk bantuan lebih lanjut")
	}
	rGrpTables, _ := s.tableTypeRepo.FindByGroup(tt.BranchID, tt.Name, tt.Capacity, tt.RoomID)
	rGrpIDs := make([]uint, len(rGrpTables))
	for i, t := range rGrpTables { rGrpIDs[i] = t.ID }
	rBooked, _ := s.repo.CountOverlappingByGroup(rGrpIDs, newDate, startTime, endTime, bookingID)
	if rBooked+int64(b.TablesCount) > int64(len(rGrpTables)) {
		return nil, errors.New("slot baru tidak tersedia, silakan pilih waktu lain")
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

// sendRescheduleNotif notifies customer (in-app, WA, email), owner, and branch staff about a reschedule.
func (s *bookingService) sendRescheduleNotif(b *model.Booking, oldDate time.Time, oldStartTime string) {
	branch, _ := s.branchRepo.FindByID(b.BranchID)
	branchName := ""
	if branch != nil {
		branchName = branch.Name
	}
	customer, _ := s.customerRepo.FindByID(b.CustomerID)
	customerName := ""
	if customer != nil {
		customerName = customer.Name
	}

	oldDateStr := oldDate.Format("02 Jan 2006")
	newDateStr := b.BookingDate.Format("02 Jan 2006")

	customerMsg := fmt.Sprintf("Booking #%d di %s berhasil dijadwalkan ulang dari %s %s ke %s %s. Menunggu konfirmasi.", b.ID, branchName, oldDateStr, oldStartTime, newDateStr, b.StartTime)
	ownerMsg := fmt.Sprintf("Booking #%d dari %s di %s diubah jadwal dari %s %s ke %s %s.", b.ID, customerName, branchName, oldDateStr, oldStartTime, newDateStr, b.StartTime)

	// In-app notification to customer
	if s.notifSvc != nil {
		_ = s.notifSvc.Send("customer", b.CustomerID, "Jadwal Booking Diubah", customerMsg)
	}

	// WhatsApp to customer
	if s.waSender != nil && customer != nil && customer.Phone != "" {
		waMsg := fmt.Sprintf(
			"Halo *%s*! 🗓️\n\nJadwal booking Anda telah diubah.\n\n📍 %s\n🔖 ID: #%d\n\n🕐 Jadwal Lama: %s %s\n✅ Jadwal Baru: %s %s\n⏰ %s – %s\n👥 %d tamu\n\nStatus kembali ke *menunggu konfirmasi*.",
			customer.Name, branchName, b.ID, oldDateStr, oldStartTime, newDateStr, b.StartTime, b.StartTime, b.EndTime, b.GuestCount,
		)
		if err := s.waSender.Send(customer.Phone, waMsg); err != nil {
			log.Printf("[WA ERROR] reschedule booking #%d ke %s: %v", b.ID, customer.Phone, err)
		}
	}

	// Email to customer
	if s.emailSender != nil && customer != nil && customer.Email != "" {
		subject := fmt.Sprintf("Jadwal Booking #%d Diubah — %s", b.ID, branchName)
		body := fmt.Sprintf(`
<p>Halo <strong>%s</strong>,</p>
<p>Jadwal booking Anda telah berhasil diubah. Berikut detailnya:</p>
<table style="border-collapse:collapse;margin:16px 0">
  <tr><td style="padding:4px 12px 4px 0;color:#6b7280">Restoran</td><td><strong>%s</strong></td></tr>
  <tr><td style="padding:4px 12px 4px 0;color:#6b7280">ID Booking</td><td>#%d</td></tr>
  <tr><td style="padding:4px 12px 4px 0;color:#6b7280">Jadwal Lama</td><td>%s pukul %s</td></tr>
  <tr><td style="padding:4px 12px 4px 0;color:#6b7280">Jadwal Baru</td><td>%s pukul %s – %s</td></tr>
  <tr><td style="padding:4px 12px 4px 0;color:#6b7280">Tamu</td><td>%d orang</td></tr>
  <tr><td style="padding:4px 12px 4px 0;color:#6b7280">Status</td><td>Menunggu konfirmasi</td></tr>
</table>
<p>Kami akan segera mengonfirmasi jadwal baru Anda.</p>`,
			customer.Name, branchName, b.ID, oldDateStr, oldStartTime, newDateStr, b.StartTime, b.EndTime, b.GuestCount)
		if err := s.emailSender.Send(customer.Email, subject, body); err != nil {
			log.Printf("[EMAIL ERROR] reschedule booking #%d ke %s: %v", b.ID, customer.Email, err)
		}
	}

	// In-app notification to owner
	if s.notifSvc != nil && branch != nil {
		rest, _ := s.restaurantRepo.FindByBranchID(branch.RestaurantID)
		if rest != nil && rest.BusinessOwnerID != nil {
			_ = s.notifSvc.Send("owner", *rest.BusinessOwnerID, "Jadwal Booking Diubah", ownerMsg)
		}
	}

	// In-app notification to all active staff in the branch
	if s.notifSvc != nil && s.staffRepo != nil && branch != nil {
		staffList, _ := s.staffRepo.FindByBranchID(b.BranchID)
		for _, st := range staffList {
			_ = s.notifSvc.Send("staff", st.ID, "Jadwal Booking Diubah", ownerMsg)
		}
	}
}

func (s *bookingService) ListCustomersByRestaurant(restaurantID uint) ([]CustomerSummary, error) {
	rows, err := s.repo.ListCustomerSummaryByRestaurant(restaurantID)
	if err != nil {
		return nil, err
	}
	result := make([]CustomerSummary, 0, len(rows))
	for _, r := range rows {
		segment := "new"
		if r.TotalVisits >= 5 {
			segment = "loyal"
		} else if r.TotalVisits >= 2 {
			segment = "regular"
		}
		result = append(result, CustomerSummary{
			CustomerID:  r.CustomerID,
			Name:        r.Name,
			Email:       r.Email,
			Phone:       r.Phone,
			TotalVisits: r.TotalVisits,
			LastVisit:   r.LastVisit,
			TotalSpend:  r.TotalSpend,
			Segment:     segment,
		})
	}
	return result, nil
}

var validOrderStatuses = map[string]bool{"new": true, "prepare": true, "ready": true, "done": true}

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

// ProcessReminders sends H-1 reminders for tomorrow's bookings (pending or confirmed).
// Designed to be called by a background job (e.g., every hour).
func (s *bookingService) ProcessReminders() error {
	candidates, err := s.repo.FindReminderCandidates()
	if err != nil {
		return err
	}
	for _, b := range candidates {
		b := b // capture loop variable
		s.sendReminderNotif(&b)
		_ = s.repo.MarkReminderSent(b.ID)
	}
	return nil
}

func (s *bookingService) sendReminderNotif(b *model.Booking) {
	customer, _ := s.customerRepo.FindByID(b.CustomerID)
	branch, _ := s.branchRepo.FindByID(b.BranchID)
	if customer == nil || branch == nil {
		return
	}
	dateStr := b.BookingDate.Format("02 Jan 2006")

	// In-app
	if s.notifSvc != nil {
		_ = s.notifSvc.Send("customer", b.CustomerID,
			"Pengingat Booking Besok",
			fmt.Sprintf("Booking #%d di %s besok (%s) pukul %s. Jangan lupa hadir ya!", b.ID, branch.Name, dateStr, b.StartTime),
		)
	}

	// WhatsApp
	if s.waSender != nil && customer.Phone != "" {
		msg := fmt.Sprintf(
			"Halo *%s*! 🔔\n\nIni pengingat booking Anda *besok*.\n\n📍 %s\n📅 %s\n⏰ %s – %s\n👥 %d tamu\n🔖 ID: #%d\n\nSampai jumpa!",
			customer.Name, branch.Name, dateStr, b.StartTime, b.EndTime, b.GuestCount, b.ID,
		)
		if err := s.waSender.Send(customer.Phone, msg); err != nil {
			log.Printf("[WA ERROR] reminder booking #%d ke %s: %v", b.ID, customer.Phone, err)
		}
	}

	// Email
	if s.emailSender != nil && customer.Email != "" {
		subject := fmt.Sprintf("Pengingat Booking Besok — %s", branch.Name)
		body := fmt.Sprintf(`
<p>Halo <strong>%s</strong>,</p>
<p>Ini pengingat bahwa Anda memiliki booking <strong>besok</strong>!</p>
<table style="border-collapse:collapse;margin:16px 0">
  <tr><td style="padding:4px 12px 4px 0;color:#6b7280">Restoran</td><td><strong>%s</strong></td></tr>
  <tr><td style="padding:4px 12px 4px 0;color:#6b7280">Tanggal</td><td>%s</td></tr>
  <tr><td style="padding:4px 12px 4px 0;color:#6b7280">Waktu</td><td>%s – %s</td></tr>
  <tr><td style="padding:4px 12px 4px 0;color:#6b7280">Tamu</td><td>%d orang</td></tr>
  <tr><td style="padding:4px 12px 4px 0;color:#6b7280">ID Booking</td><td>#%d</td></tr>
</table>
<p>Sampai jumpa besok! 🎉</p>`,
			customer.Name, branch.Name, dateStr, b.StartTime, b.EndTime, b.GuestCount, b.ID)
		if err := s.emailSender.Send(customer.Email, subject, body); err != nil {
			log.Printf("[EMAIL ERROR] reminder booking #%d ke %s: %v", b.ID, customer.Email, err)
		}
	}
}

// --- helpers ---

// checkMinBookingHours returns an error if startTime is within branch.MinBookingHours from now (today only).
func checkMinBookingHours(date time.Time, startTime string, minHours int) error {
	if minHours <= 0 || !date.Equal(today()) {
		return nil
	}
	nowMins := timeToMinutes(nowWIB().Format("15:04"))
	startMins := timeToMinutes(startTime)
	if startMins <= nowMins+minHours*60 {
		return fmt.Errorf("reservasi minimal %d jam dari sekarang", minHours)
	}
	return nil
}

func parseDate(s string) (time.Time, error) {
	return time.ParseInLocation("2006-01-02", s, jakartaLoc)
}

var jakartaLoc = func() *time.Location {
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		return time.UTC
	}
	return loc
}()

func today() time.Time {
	y, m, d := time.Now().In(jakartaLoc).Date()
	return time.Date(y, m, d, 0, 0, 0, 0, jakartaLoc)
}

func nowWIB() time.Time {
	return time.Now().In(jakartaLoc)
}

func validateTime(t string) error {
	parts := strings.Split(t, ":")
	if len(parts) != 2 || len(parts[0]) != 2 || len(parts[1]) != 2 {
		return errors.New("format waktu tidak valid, gunakan HH:MM")
	}
	return nil
}

// addMinutes adds minutes to a "HH:MM" string and returns "HH:MM".
func addMinutes(t string, minutes int) string {
	base, _ := time.Parse("15:04", t)
	result := base.Add(time.Duration(minutes) * time.Minute)
	return result.Format("15:04")
}
