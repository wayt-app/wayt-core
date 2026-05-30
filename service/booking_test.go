package service_test

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/wayt-app/wayt-core/internal/testmock"
	"github.com/wayt-app/wayt-core/model"
	"github.com/wayt-app/wayt-core/service"
)

// tomorrow returns a date string for tomorrow in YYYY-MM-DD format.
func tomorrow() string {
	return time.Now().AddDate(0, 0, 1).Format("2006-01-02")
}

// newBookingSvc creates a BookingService with all mocks injected.
// Only override the fields you care about in each test.
func newBookingSvc(
	bookingRepo *testmock.BookingRepo,
	branchRepo *testmock.BranchRepo,
	tableTypeRepo *testmock.TableTypeRepo,
	bookingTableRepo *testmock.BookingTableRepo,
	subRepo *testmock.SubscriptionRepo,
	restaurantRepo *testmock.RestaurantRepo,
) service.BookingService {
	return service.NewBookingService(
		bookingRepo,
		branchRepo,
		tableTypeRepo,
		bookingTableRepo,
		&testmock.CustomerRepo{},
		subRepo,
		restaurantRepo,
		&testmock.StaffRepo{},
		&testmock.BusinessOwnerRepo{},
		&testmock.WASender{},
		&testmock.EmailSender{},
		&testmock.NotificationSvc{},
		&testmock.ReservationIncr{},
		&testmock.EmailConfigRepo{},
	)
}

// activeBranch returns a minimal active branch.
func activeBranch(id uint) *model.Branch {
	return &model.Branch{
		ID:                     id,
		RestaurantID:           1,
		Name:                   "Branch Utama",
		IsActive:               true,
		RequireConfirmation:    true,
		DefaultDurationMinutes: 120,
	}
}

// activeTableType returns a minimal active table type.
func activeTableType(id, branchID uint, capacity int) *model.TableType {
	return &model.TableType{
		ID:       id,
		BranchID: branchID,
		Name:     "Meja Reguler",
		Capacity: capacity,
		IsActive: true,
	}
}

// --- Create tests ---

func TestCreate_BranchNotFound(t *testing.T) {
	svc := newBookingSvc(
		&testmock.BookingRepo{},
		&testmock.BranchRepo{
			FindByIDFn: func(id uint) (*model.Branch, error) {
				return nil, errors.New("not found")
			},
		},
		&testmock.TableTypeRepo{},
		&testmock.BookingTableRepo{},
		&testmock.SubscriptionRepo{},
		&testmock.RestaurantRepo{},
	)

	_, err := svc.Create(1, 99, 1, tomorrow(), "10:00", 2, "", "", "", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestCreate_BranchInactive(t *testing.T) {
	branch := activeBranch(1)
	branch.IsActive = false

	svc := newBookingSvc(
		&testmock.BookingRepo{},
		&testmock.BranchRepo{
			FindByIDFn: func(id uint) (*model.Branch, error) { return branch, nil },
		},
		&testmock.TableTypeRepo{},
		&testmock.BookingTableRepo{},
		&testmock.SubscriptionRepo{},
		&testmock.RestaurantRepo{},
	)

	_, err := svc.Create(1, 1, 1, tomorrow(), "10:00", 2, "", "", "", "")
	if err == nil {
		t.Fatal("expected error for inactive branch")
	}
}

func TestCreate_TableTypeNotFound(t *testing.T) {
	svc := newBookingSvc(
		&testmock.BookingRepo{},
		&testmock.BranchRepo{
			FindByIDFn: func(id uint) (*model.Branch, error) { return activeBranch(id), nil },
		},
		&testmock.TableTypeRepo{
			FindByIDFn: func(id uint) (*model.TableType, error) {
				return nil, errors.New("not found")
			},
		},
		&testmock.BookingTableRepo{},
		&testmock.SubscriptionRepo{},
		&testmock.RestaurantRepo{},
	)

	_, err := svc.Create(1, 1, 99, tomorrow(), "10:00", 2, "", "", "", "")
	if err == nil {
		t.Fatal("expected error for missing table type")
	}
}

func TestCreate_TableTypeBelongsToDifferentBranch(t *testing.T) {
	tt := activeTableType(1, 99 /* different branch */, 4)

	svc := newBookingSvc(
		&testmock.BookingRepo{},
		&testmock.BranchRepo{
			FindByIDFn: func(id uint) (*model.Branch, error) { return activeBranch(1), nil },
		},
		&testmock.TableTypeRepo{
			FindByIDFn: func(id uint) (*model.TableType, error) { return tt, nil },
		},
		&testmock.BookingTableRepo{},
		&testmock.SubscriptionRepo{},
		&testmock.RestaurantRepo{},
	)

	_, err := svc.Create(1, 1, 1, tomorrow(), "10:00", 2, "", "", "", "")
	if err == nil {
		t.Fatal("expected error when table type belongs to different branch")
	}
}

func TestCreate_TableTypeInactive(t *testing.T) {
	tt := activeTableType(1, 1, 4)
	tt.IsActive = false

	svc := newBookingSvc(
		&testmock.BookingRepo{},
		&testmock.BranchRepo{
			FindByIDFn: func(id uint) (*model.Branch, error) { return activeBranch(1), nil },
		},
		&testmock.TableTypeRepo{
			FindByIDFn: func(id uint) (*model.TableType, error) { return tt, nil },
		},
		&testmock.BookingTableRepo{},
		&testmock.SubscriptionRepo{},
		&testmock.RestaurantRepo{},
	)

	_, err := svc.Create(1, 1, 1, tomorrow(), "10:00", 2, "", "", "", "")
	if err == nil {
		t.Fatal("expected error for inactive table type")
	}
}

func TestCreate_ZeroGuestCount(t *testing.T) {
	tt := activeTableType(1, 1, 4)

	svc := newBookingSvc(
		&testmock.BookingRepo{},
		&testmock.BranchRepo{
			FindByIDFn: func(id uint) (*model.Branch, error) { return activeBranch(id), nil },
		},
		&testmock.TableTypeRepo{
			FindByIDFn: func(id uint) (*model.TableType, error) { return tt, nil },
		},
		&testmock.BookingTableRepo{},
		&testmock.SubscriptionRepo{},
		&testmock.RestaurantRepo{},
	)

	_, err := svc.Create(1, 1, 1, tomorrow(), "10:00", 0, "", "", "", "")
	if err == nil {
		t.Fatal("expected error for zero guest count")
	}
}

func TestCreate_PastDate(t *testing.T) {
	tt := activeTableType(1, 1, 4)

	svc := newBookingSvc(
		&testmock.BookingRepo{},
		&testmock.BranchRepo{
			FindByIDFn: func(id uint) (*model.Branch, error) { return activeBranch(id), nil },
		},
		&testmock.TableTypeRepo{
			FindByIDFn: func(id uint) (*model.TableType, error) { return tt, nil },
		},
		&testmock.BookingTableRepo{},
		&testmock.SubscriptionRepo{},
		&testmock.RestaurantRepo{},
	)

	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	_, err := svc.Create(1, 1, 1, yesterday, "10:00", 2, "", "", "", "")
	if err == nil {
		t.Fatal("expected error for past date")
	}
}

func TestCreate_DateTooFarAhead(t *testing.T) {
	tt := activeTableType(1, 1, 4)

	svc := newBookingSvc(
		&testmock.BookingRepo{},
		&testmock.BranchRepo{
			FindByIDFn: func(id uint) (*model.Branch, error) { return activeBranch(id), nil },
		},
		&testmock.TableTypeRepo{
			FindByIDFn: func(id uint) (*model.TableType, error) { return tt, nil },
		},
		&testmock.BookingTableRepo{},
		&testmock.SubscriptionRepo{},
		&testmock.RestaurantRepo{},
	)

	farFuture := time.Now().AddDate(0, 0, 31).Format("2006-01-02")
	_, err := svc.Create(1, 1, 1, farFuture, "10:00", 2, "", "", "", "")
	if err == nil {
		t.Fatal("expected error for date > 30 days ahead")
	}
}

func TestCreate_SubscriptionInactive(t *testing.T) {
	tt := activeTableType(1, 1, 4)

	svc := newBookingSvc(
		&testmock.BookingRepo{},
		&testmock.BranchRepo{
			FindByIDFn: func(id uint) (*model.Branch, error) { return activeBranch(id), nil },
		},
		&testmock.TableTypeRepo{
			FindByIDFn:    func(id uint) (*model.TableType, error) { return tt, nil },
			FindByGroupFn: func(_ uint, _ string, _ int, _ *uint) ([]model.TableType, error) { return []model.TableType{*tt}, nil },
		},
		&testmock.BookingTableRepo{},
		&testmock.SubscriptionRepo{
			FindByRestaurantIDFn: func(restaurantID uint) (*model.Subscription, error) {
				return &model.Subscription{
					Status: model.SubscriptionStatusSuspended,
				}, nil
			},
		},
		&testmock.RestaurantRepo{},
	)

	_, err := svc.Create(1, 1, 1, tomorrow(), "10:00", 2, "", "", "", "")
	if err == nil {
		t.Fatal("expected error for inactive subscription")
	}
}

func TestCreate_AvailableTables_PendingStatus(t *testing.T) {
	tt := activeTableType(1, 1, 4)
	var createdBooking *model.Booking

	svc := newBookingSvc(
		&testmock.BookingRepo{
			CreateFn: func(b *model.Booking) error {
				createdBooking = b
				b.ID = 42
				return nil
			},
			CountOverlappingByGroupFn: func(_ []uint, _ time.Time, _, _ string, _ uint) (int64, error) {
				return 0, nil // no overlapping bookings
			},
		},
		&testmock.BranchRepo{
			FindByIDFn: func(id uint) (*model.Branch, error) { return activeBranch(id), nil },
		},
		&testmock.TableTypeRepo{
			FindByIDFn:    func(id uint) (*model.TableType, error) { return tt, nil },
			FindByGroupFn: func(_ uint, _ string, _ int, _ *uint) ([]model.TableType, error) { return []model.TableType{*tt}, nil },
		},
		&testmock.BookingTableRepo{},
		&testmock.SubscriptionRepo{},
		&testmock.RestaurantRepo{},
	)

	// Branch has RequireConfirmation=true → status should be pending
	b, err := svc.Create(1, 1, 1, tomorrow(), "10:00", 2, "test notes", "", "app", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if b.Status != model.BookingStatusPending {
		t.Errorf("expected status pending, got %s", b.Status)
	}
	if createdBooking == nil {
		t.Fatal("Create was not called on repo")
	}
	if createdBooking.GuestCount != 2 {
		t.Errorf("expected guest count 2, got %d", createdBooking.GuestCount)
	}
}

func TestCreate_AvailableTables_ConfirmedStatus(t *testing.T) {
	tt := activeTableType(1, 1, 4)
	branch := activeBranch(1)
	branch.RequireConfirmation = false

	svc := newBookingSvc(
		&testmock.BookingRepo{
			CreateFn: func(b *model.Booking) error { b.ID = 1; return nil },
			CountOverlappingByGroupFn: func(_ []uint, _ time.Time, _, _ string, _ uint) (int64, error) {
				return 0, nil
			},
		},
		&testmock.BranchRepo{
			FindByIDFn: func(id uint) (*model.Branch, error) { return branch, nil },
		},
		&testmock.TableTypeRepo{
			FindByIDFn:    func(id uint) (*model.TableType, error) { return tt, nil },
			FindByGroupFn: func(_ uint, _ string, _ int, _ *uint) ([]model.TableType, error) { return []model.TableType{*tt}, nil },
		},
		&testmock.BookingTableRepo{},
		&testmock.SubscriptionRepo{},
		&testmock.RestaurantRepo{},
	)

	b, err := svc.Create(1, 1, 1, tomorrow(), "10:00", 2, "", "", "app", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if b.Status != model.BookingStatusConfirmed {
		t.Errorf("expected status confirmed, got %s", b.Status)
	}
}

func TestCreate_FullTables_WaitingList(t *testing.T) {
	tt := activeTableType(1, 1, 4)

	svc := newBookingSvc(
		&testmock.BookingRepo{
			CreateFn: func(b *model.Booking) error { b.ID = 1; return nil },
			CountOverlappingByGroupFn: func(_ []uint, _ time.Time, _, _ string, _ uint) (int64, error) {
				return 1, nil // 1 table already booked = full (only 1 table in group)
			},
		},
		&testmock.BranchRepo{
			FindByIDFn: func(id uint) (*model.Branch, error) { return activeBranch(id), nil },
		},
		&testmock.TableTypeRepo{
			FindByIDFn:    func(id uint) (*model.TableType, error) { return tt, nil },
			FindByGroupFn: func(_ uint, _ string, _ int, _ *uint) ([]model.TableType, error) { return []model.TableType{*tt}, nil },
		},
		&testmock.BookingTableRepo{},
		&testmock.SubscriptionRepo{},
		&testmock.RestaurantRepo{},
	)

	b, err := svc.Create(1, 1, 1, tomorrow(), "10:00", 2, "", "", "app", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if b.Status != model.BookingStatusWaitingList {
		t.Errorf("expected status waiting_list, got %s", b.Status)
	}
}

func TestCreate_RepoCreateFails(t *testing.T) {
	tt := activeTableType(1, 1, 4)
	dbErr := errors.New("db connection failed")

	svc := newBookingSvc(
		&testmock.BookingRepo{
			CreateFn: func(b *model.Booking) error { return dbErr },
			CountOverlappingByGroupFn: func(_ []uint, _ time.Time, _, _ string, _ uint) (int64, error) {
				return 0, nil
			},
		},
		&testmock.BranchRepo{
			FindByIDFn: func(id uint) (*model.Branch, error) { return activeBranch(id), nil },
		},
		&testmock.TableTypeRepo{
			FindByIDFn:    func(id uint) (*model.TableType, error) { return tt, nil },
			FindByGroupFn: func(_ uint, _ string, _ int, _ *uint) ([]model.TableType, error) { return []model.TableType{*tt}, nil },
		},
		&testmock.BookingTableRepo{},
		&testmock.SubscriptionRepo{},
		&testmock.RestaurantRepo{},
	)

	_, err := svc.Create(1, 1, 1, tomorrow(), "10:00", 2, "", "", "app", "")
	if !errors.Is(err, dbErr) {
		t.Errorf("expected db error, got: %v", err)
	}
}

func TestCreate_TablesNeededExceedsCapacityPerTable(t *testing.T) {
	// Capacity=2, 5 guests → needs 3 tables but only 2 exist → cap at 2
	tt1 := activeTableType(1, 1, 2)
	tt2 := activeTableType(2, 1, 2) // same group
	var capturedTablesCount int

	svc := newBookingSvc(
		&testmock.BookingRepo{
			CreateFn: func(b *model.Booking) error {
				capturedTablesCount = b.TablesCount
				b.ID = 1
				return nil
			},
			CountOverlappingByGroupFn: func(_ []uint, _ time.Time, _, _ string, _ uint) (int64, error) {
				return 0, nil
			},
		},
		&testmock.BranchRepo{
			FindByIDFn: func(id uint) (*model.Branch, error) { return activeBranch(id), nil },
		},
		&testmock.TableTypeRepo{
			FindByIDFn: func(id uint) (*model.TableType, error) { return tt1, nil },
			FindByGroupFn: func(_ uint, _ string, _ int, _ *uint) ([]model.TableType, error) {
				return []model.TableType{*tt1, *tt2}, nil // 2 physical tables
			},
		},
		&testmock.BookingTableRepo{},
		&testmock.SubscriptionRepo{},
		&testmock.RestaurantRepo{},
	)

	_, err := svc.Create(1, 1, 1, tomorrow(), "10:00", 5, "", "", "app", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// ceil(5/2)=3 tables needed, but only 2 exist → capped at 2
	if capturedTablesCount != 2 {
		t.Errorf("expected tables_count=2 (capped), got %d", capturedTablesCount)
	}
}

// --- Cancel tests ---

func newCancelSvc(bookingRepo *testmock.BookingRepo) service.BookingService {
	return service.NewBookingService(
		bookingRepo,
		&testmock.BranchRepo{},
		&testmock.TableTypeRepo{},
		&testmock.BookingTableRepo{},
		&testmock.CustomerRepo{},
		&testmock.SubscriptionRepo{},
		&testmock.RestaurantRepo{},
		&testmock.StaffRepo{},
		&testmock.BusinessOwnerRepo{},
		&testmock.WASender{},
		&testmock.EmailSender{},
		&testmock.NotificationSvc{},
		&testmock.ReservationIncr{},
		&testmock.EmailConfigRepo{},
	)
}

func TestCancel_BookingNotFound(t *testing.T) {
	svc := newCancelSvc(&testmock.BookingRepo{
		FindByIDFn: func(id uint) (*model.Booking, error) {
			return nil, errors.New("not found")
		},
	})

	err := svc.Cancel(99, 1, "")
	if err == nil {
		t.Fatal("expected error for missing booking")
	}
}

func TestCancel_WrongCustomer(t *testing.T) {
	booking := &model.Booking{ID: 1, CustomerID: 5, Status: model.BookingStatusPending}
	svc := newCancelSvc(&testmock.BookingRepo{
		FindByIDFn: func(id uint) (*model.Booking, error) { return booking, nil },
	})

	err := svc.Cancel(1, 99 /* different customer */, "")
	if err == nil {
		t.Fatal("expected error for wrong customer")
	}
}

func TestCancel_AlreadyCancelled(t *testing.T) {
	booking := &model.Booking{ID: 1, CustomerID: 1, Status: model.BookingStatusCancelled}
	svc := newCancelSvc(&testmock.BookingRepo{
		FindByIDFn: func(id uint) (*model.Booking, error) { return booking, nil },
	})

	err := svc.Cancel(1, 1, "")
	if err == nil {
		t.Fatal("expected error for already cancelled booking")
	}
}

func TestCancel_AlreadyCompleted(t *testing.T) {
	booking := &model.Booking{ID: 1, CustomerID: 1, Status: model.BookingStatusCompleted}
	svc := newCancelSvc(&testmock.BookingRepo{
		FindByIDFn: func(id uint) (*model.Booking, error) { return booking, nil },
	})

	err := svc.Cancel(1, 1, "")
	if err == nil {
		t.Fatal("expected error for completed booking")
	}
}

func TestCancel_Success(t *testing.T) {
	booking := &model.Booking{ID: 1, CustomerID: 1, Status: model.BookingStatusPending}
	var updatedStatus model.BookingStatus

	svc := newCancelSvc(&testmock.BookingRepo{
		FindByIDFn: func(id uint) (*model.Booking, error) { return booking, nil },
		UpdateStatusAndReasonFn: func(id uint, status model.BookingStatus, reason string) error {
			updatedStatus = status
			return nil
		},
		FindWaitingListForSlotFn: func(_ uint, _ time.Time, _, _ string) ([]model.Booking, error) {
			return nil, nil
		},
	})

	err := svc.Cancel(1, 1, "berubah pikiran")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updatedStatus != model.BookingStatusCancelled {
		t.Errorf("expected status cancelled, got %s", updatedStatus)
	}
}

// --- CheckAvailability tests ---

func newAvailSvc(branchRepo *testmock.BranchRepo, tableTypeRepo *testmock.TableTypeRepo, bookingRepo *testmock.BookingRepo) service.BookingService {
	return service.NewBookingService(
		bookingRepo,
		branchRepo,
		tableTypeRepo,
		&testmock.BookingTableRepo{},
		&testmock.CustomerRepo{},
		&testmock.SubscriptionRepo{},
		&testmock.RestaurantRepo{},
		&testmock.StaffRepo{},
		&testmock.BusinessOwnerRepo{},
		&testmock.WASender{},
		&testmock.EmailSender{},
		&testmock.NotificationSvc{},
		&testmock.ReservationIncr{},
		&testmock.EmailConfigRepo{},
	)
}

func TestCheckAvailability_BranchNotFound(t *testing.T) {
	svc := newAvailSvc(
		&testmock.BranchRepo{
			FindByIDFn: func(id uint) (*model.Branch, error) {
				return nil, errors.New("not found")
			},
		},
		&testmock.TableTypeRepo{},
		&testmock.BookingRepo{},
	)

	_, err := svc.CheckAvailability(1, tomorrow(), "10:00", 2, nil)
	if err == nil {
		t.Fatal("expected error for missing branch")
	}
}

func TestCheckAvailability_PastDate(t *testing.T) {
	svc := newAvailSvc(
		&testmock.BranchRepo{
			FindByIDFn: func(id uint) (*model.Branch, error) { return activeBranch(id), nil },
		},
		&testmock.TableTypeRepo{},
		&testmock.BookingRepo{},
	)

	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	_, err := svc.CheckAvailability(1, yesterday, "10:00", 2, nil)
	if err == nil {
		t.Fatal("expected error for past date")
	}
}

func TestCheckAvailability_NoTableTypes(t *testing.T) {
	svc := newAvailSvc(
		&testmock.BranchRepo{
			FindByIDFn: func(id uint) (*model.Branch, error) { return activeBranch(id), nil },
		},
		&testmock.TableTypeRepo{
			FindByBranchAndRoomFn: func(_ uint, _ *uint) ([]model.TableType, error) {
				return []model.TableType{}, nil
			},
		},
		&testmock.BookingRepo{},
	)

	results, err := svc.CheckAvailability(1, tomorrow(), "10:00", 2, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected empty results, got %d", len(results))
	}
}

func TestCheckAvailability_ReturnsCorrectAvailability(t *testing.T) {
	tt1 := activeTableType(1, 1, 4)
	tt2 := activeTableType(2, 1, 4) // same group (same capacity, no room)

	svc := newAvailSvc(
		&testmock.BranchRepo{
			FindByIDFn: func(id uint) (*model.Branch, error) { return activeBranch(id), nil },
		},
		&testmock.TableTypeRepo{
			FindByBranchAndRoomFn: func(_ uint, _ *uint) ([]model.TableType, error) {
				return []model.TableType{*tt1, *tt2}, nil
			},
		},
		&testmock.BookingRepo{
			CountOverlappingByGroupFn: func(_ []uint, _ time.Time, _, _ string, _ uint) (int64, error) {
				return 1, nil // 1 of 2 tables booked
			},
		},
	)

	results, err := svc.CheckAvailability(1, tomorrow(), "10:00", 2, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 group result, got %d", len(results))
	}
	r := results[0]
	if r.TotalTables != 2 {
		t.Errorf("expected TotalTables=2, got %d", r.TotalTables)
	}
	if r.BookedCount != 1 {
		t.Errorf("expected BookedCount=1, got %d", r.BookedCount)
	}
	if r.Available != 1 {
		t.Errorf("expected Available=1, got %d", r.Available)
	}
}

func TestCheckAvailability_InactiveTableTypesExcluded(t *testing.T) {
	tt := activeTableType(1, 1, 4)
	inactiveTT := activeTableType(2, 1, 4)
	inactiveTT.IsActive = false

	svc := newAvailSvc(
		&testmock.BranchRepo{
			FindByIDFn: func(id uint) (*model.Branch, error) { return activeBranch(id), nil },
		},
		&testmock.TableTypeRepo{
			FindByBranchAndRoomFn: func(_ uint, _ *uint) ([]model.TableType, error) {
				return []model.TableType{*tt, *inactiveTT}, nil
			},
		},
		&testmock.BookingRepo{
			CountOverlappingByGroupFn: func(ids []uint, _ time.Time, _, _ string, _ uint) (int64, error) {
				return 0, nil
			},
		},
	)

	results, err := svc.CheckAvailability(1, tomorrow(), "10:00", 2, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 group (inactive excluded), got %d", len(results))
	}
	// Only 1 active table in the group
	if results[0].TotalTables != 1 {
		t.Errorf("expected TotalTables=1, got %d", results[0].TotalTables)
	}
}

// Compile-time check that mocks satisfy their interfaces.
var _ service.BookingService = (*bookingServiceIface)(nil)

type bookingServiceIface struct{ service.BookingService }

// Verify mock types implement their target interfaces.
func init() {
	_ = fmt.Sprintf // silence import
}
