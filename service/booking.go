package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/wayt-app/wayt-core/model"
	"github.com/wayt-app/wayt-core/pkg/email"
	"github.com/wayt-app/wayt-core/pkg/whatsapp"
	"github.com/wayt-app/wayt-core/repository"
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
	Date        string           `json:"date"`
	BranchID    uint             `json:"branch_id"`
	Counts      map[string]int64 `json:"counts"`
	TotalToday  int64            `json:"total_today"`
	NoShowCount int64            `json:"no_show_count"`
}

type BookingPage struct {
	Data       []model.Booking `json:"data"`
	Total      int64           `json:"total"`
	Page       int             `json:"page"`
	Limit      int             `json:"limit"`
	TotalPages int64           `json:"total_pages"`
}

type AvailabilityResult struct {
	TableTypeID  uint   `json:"table_type_id"`
	Name         string `json:"name"`
	Capacity     int    `json:"capacity"`
	TotalTables  int    `json:"total_tables"`
	BookedCount  int64  `json:"booked_count"`
	Available    int64  `json:"available"`
	TablesNeeded int    `json:"tables_needed"`
	StartTime    string `json:"start_time"`
	EndTime      string `json:"end_time"`
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

type TableOption struct {
	TableTypeID uint   `json:"table_type_id"`
	Name        string `json:"name"`
	Capacity    int    `json:"capacity"`
	RoomID      *uint  `json:"room_id,omitempty"`
	RoomName    string `json:"room_name,omitempty"`
	IsAvailable bool   `json:"is_available"`
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
	// AdminChangeTableType reassigns the table type for an active booking (single-type).
	AdminChangeTableType(bookingID, tableTypeID uint) (*model.Booking, error)
	// GetTableChangeOptions returns all individual tables for a booking's branch with availability at the booking's slot.
	GetTableChangeOptions(bookingID uint) ([]TableOption, error)
	// AdminChangeTables manually reassigns a booking to the given set of table type IDs.
	AdminChangeTables(bookingID uint, tableTypeIDs []uint) (*model.Booking, error)
	// MyBookingsPaged returns a paginated list of a customer's bookings.
	MyBookingsPaged(customerID uint, sortBy, sortDir string, page, limit int) (*BookingPage, error)
	// ProcessReminders sends H-1 reminder notifications for tomorrow's bookings.
	ProcessReminders() error
	// ListCustomersByRestaurant returns aggregated customer data for a restaurant for retention purposes.
	// Sorted by last_visit ASC (longest-absent first).
	ListCustomersByRestaurant(restaurantID uint) ([]CustomerSummary, error)
	// UpdateOrderStatus changes the pre-order kitchen status (new/prepare/ready/done).
	UpdateOrderStatus(bookingID, branchID uint, status string) error
	// UploadPaymentProof saves a payment proof URL for a customer booking.
	UploadPaymentProof(bookingID, customerID uint, proofURL string) error
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

var validOrderStatuses = map[string]bool{"new": true, "prepare": true, "ready": true, "done": true}

var jakartaLoc = func() *time.Location {
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		return time.UTC
	}
	return loc
}()

// --- helpers ---

// checkMinBookingHours returns an error if the booking slot is within minHours from now.
func checkMinBookingHours(date time.Time, startTime string, minHours int) error {
	if minHours <= 0 {
		return nil
	}
	now := nowWIB()
	h, m := 0, 0
	fmt.Sscanf(startTime, "%d:%d", &h, &m)
	slotTime := time.Date(date.Year(), date.Month(), date.Day(), h, m, 0, 0, now.Location())
	if !slotTime.After(now.Add(time.Duration(minHours) * time.Hour)) {
		return fmt.Errorf("reservasi minimal %d jam dari sekarang", minHours)
	}
	return nil
}

func parseDate(s string) (time.Time, error) {
	return time.ParseInLocation("2006-01-02", s, jakartaLoc)
}

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

// addMinutes adds minutes to a time string ("HH:MM" or "HH:MM:SS") and returns "HH:MM".
func addMinutes(t string, minutes int) string {
	if len(t) > 5 {
		t = t[:5] // normalize "HH:MM:SS" → "HH:MM"
	}
	base, _ := time.Parse("15:04", t)
	return base.Add(time.Duration(minutes) * time.Minute).Format("15:04")
}
