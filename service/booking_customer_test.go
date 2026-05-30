package service_test

import (
	"errors"
	"testing"
	"time"

	"github.com/wayt-app/wayt-core/internal/testmock"
	"github.com/wayt-app/wayt-core/model"
)

// --- GetByID ---

func TestGetByID_NotFound(t *testing.T) {
	svc := newBookingSvc(
		&testmock.BookingRepo{
			FindByIDFn: func(id uint) (*model.Booking, error) {
				return nil, errors.New("not found")
			},
		},
		&testmock.BranchRepo{}, &testmock.TableTypeRepo{},
		&testmock.BookingTableRepo{}, nil, nil,
	)
	_, err := svc.GetByID(99)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetByID_Success(t *testing.T) {
	svc := newBookingSvc(
		&testmock.BookingRepo{
			FindByIDFn: func(id uint) (*model.Booking, error) {
				return &model.Booking{ID: id, Status: model.BookingStatusConfirmed}, nil
			},
		},
		&testmock.BranchRepo{}, &testmock.TableTypeRepo{},
		&testmock.BookingTableRepo{}, nil, nil,
	)
	b, err := svc.GetByID(5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if b.ID != 5 {
		t.Fatalf("expected ID 5, got %d", b.ID)
	}
}

// --- WaitingListPosition ---

func TestWaitingListPosition_BookingNotFound(t *testing.T) {
	svc := newBookingSvc(
		&testmock.BookingRepo{
			FindByIDFn: func(id uint) (*model.Booking, error) {
				return nil, errors.New("not found")
			},
		},
		&testmock.BranchRepo{}, &testmock.TableTypeRepo{},
		&testmock.BookingTableRepo{}, nil, nil,
	)
	_, err := svc.WaitingListPosition(1)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestWaitingListPosition_NotInWaitingList(t *testing.T) {
	svc := newBookingSvc(
		&testmock.BookingRepo{
			FindByIDFn: func(id uint) (*model.Booking, error) {
				return &model.Booking{ID: id, Status: model.BookingStatusConfirmed}, nil
			},
		},
		&testmock.BranchRepo{}, &testmock.TableTypeRepo{},
		&testmock.BookingTableRepo{}, nil, nil,
	)
	_, err := svc.WaitingListPosition(1)
	if err == nil {
		t.Fatal("expected error for non-waiting-list booking")
	}
}

func TestWaitingListPosition_ReturnsOneIndexed(t *testing.T) {
	svc := newBookingSvc(
		&testmock.BookingRepo{
			FindByIDFn: func(id uint) (*model.Booking, error) {
				return &model.Booking{ID: id, BranchID: 1, Status: model.BookingStatusWaitingList, StartTime: "12:00"}, nil
			},
			CountWaitingListBeforeFn: func(bookingID uint, branchID uint, date time.Time, startTime string) (int64, error) {
				return 2, nil // 2 waiters before this one → position 3
			},
		},
		&testmock.BranchRepo{}, &testmock.TableTypeRepo{},
		&testmock.BookingTableRepo{}, nil, nil,
	)
	pos, err := svc.WaitingListPosition(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pos != 3 {
		t.Fatalf("expected position 3, got %d", pos)
	}
}

// --- MyBookingsPaged ---

func TestMyBookingsPaged_DefaultsOnBadInput(t *testing.T) {
	var gotOffset, gotLimit int
	svc := newBookingSvc(
		&testmock.BookingRepo{
			FindByCustomerPagedFn: func(customerID uint, sortBy, sortDir string, offset, limit int) ([]model.Booking, int64, error) {
				gotOffset = offset
				gotLimit = limit
				return []model.Booking{}, 0, nil
			},
		},
		&testmock.BranchRepo{}, &testmock.TableTypeRepo{},
		&testmock.BookingTableRepo{}, nil, nil,
	)
	// page=0 → clamped to 1, limit=200 → clamped to 10
	_, err := svc.MyBookingsPaged(1, "", "desc", 0, 200)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotOffset != 0 {
		t.Errorf("expected offset 0, got %d", gotOffset)
	}
	if gotLimit != 10 {
		t.Errorf("expected limit 10, got %d", gotLimit)
	}
}

func TestMyBookingsPaged_TotalPagesCalculation(t *testing.T) {
	svc := newBookingSvc(
		&testmock.BookingRepo{
			FindByCustomerPagedFn: func(customerID uint, sortBy, sortDir string, offset, limit int) ([]model.Booking, int64, error) {
				return make([]model.Booking, 5), 25, nil
			},
		},
		&testmock.BranchRepo{}, &testmock.TableTypeRepo{},
		&testmock.BookingTableRepo{}, nil, nil,
	)
	page, err := svc.MyBookingsPaged(1, "", "desc", 1, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if page.TotalPages != 3 {
		t.Errorf("expected 3 total pages for 25 items at limit 10, got %d", page.TotalPages)
	}
	if page.Total != 25 {
		t.Errorf("expected total 25, got %d", page.Total)
	}
}

// --- UploadPaymentProof ---

func TestUploadPaymentProof_BookingNotFound(t *testing.T) {
	svc := newBookingSvc(
		&testmock.BookingRepo{
			FindByIDFn: func(id uint) (*model.Booking, error) { return nil, errors.New("not found") },
		},
		&testmock.BranchRepo{}, &testmock.TableTypeRepo{},
		&testmock.BookingTableRepo{}, nil, nil,
	)
	err := svc.UploadPaymentProof(1, 1, "https://example.com/proof.jpg")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestUploadPaymentProof_WrongCustomer(t *testing.T) {
	svc := newBookingSvc(
		&testmock.BookingRepo{
			FindByIDFn: func(id uint) (*model.Booking, error) {
				return &model.Booking{ID: id, CustomerID: 10}, nil
			},
		},
		&testmock.BranchRepo{}, &testmock.TableTypeRepo{},
		&testmock.BookingTableRepo{}, nil, nil,
	)
	err := svc.UploadPaymentProof(1, 99, "https://example.com/proof.jpg")
	if err == nil {
		t.Fatal("expected error for wrong customer")
	}
}

func TestUploadPaymentProof_EmptyURL(t *testing.T) {
	svc := newBookingSvc(
		&testmock.BookingRepo{
			FindByIDFn: func(id uint) (*model.Booking, error) {
				return &model.Booking{ID: id, CustomerID: 1}, nil
			},
		},
		&testmock.BranchRepo{}, &testmock.TableTypeRepo{},
		&testmock.BookingTableRepo{}, nil, nil,
	)
	err := svc.UploadPaymentProof(1, 1, "")
	if err == nil {
		t.Fatal("expected error for empty URL")
	}
}

func TestUploadPaymentProof_Success(t *testing.T) {
	var savedURL string
	svc := newBookingSvc(
		&testmock.BookingRepo{
			FindByIDFn: func(id uint) (*model.Booking, error) {
				return &model.Booking{ID: id, CustomerID: 1}, nil
			},
			UpdatePaymentProofFn: func(id uint, url string) error {
				savedURL = url
				return nil
			},
		},
		&testmock.BranchRepo{}, &testmock.TableTypeRepo{},
		&testmock.BookingTableRepo{}, nil, nil,
	)
	err := svc.UploadPaymentProof(1, 1, "https://example.com/proof.jpg")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if savedURL != "https://example.com/proof.jpg" {
		t.Errorf("expected URL to be saved, got %q", savedURL)
	}
}

// --- Reschedule ---

func TestReschedule_BookingNotFound(t *testing.T) {
	svc := newBookingSvc(
		&testmock.BookingRepo{
			FindByIDFn: func(id uint) (*model.Booking, error) { return nil, errors.New("not found") },
		},
		&testmock.BranchRepo{}, &testmock.TableTypeRepo{},
		&testmock.BookingTableRepo{}, nil, nil,
	)
	_, err := svc.Reschedule(1, 1, tomorrow(), "12:00")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestReschedule_WrongCustomer(t *testing.T) {
	svc := newBookingSvc(
		&testmock.BookingRepo{
			FindByIDFn: func(id uint) (*model.Booking, error) {
				return &model.Booking{ID: id, CustomerID: 10, Status: model.BookingStatusPending,
					BookingDate: time.Now().AddDate(0, 0, 2)}, nil
			},
		},
		&testmock.BranchRepo{}, &testmock.TableTypeRepo{},
		&testmock.BookingTableRepo{}, nil, nil,
	)
	_, err := svc.Reschedule(1, 99, tomorrow(), "12:00")
	if err == nil {
		t.Fatal("expected error for wrong customer")
	}
}

func TestReschedule_NotPending(t *testing.T) {
	svc := newBookingSvc(
		&testmock.BookingRepo{
			FindByIDFn: func(id uint) (*model.Booking, error) {
				return &model.Booking{ID: id, CustomerID: 1, Status: model.BookingStatusConfirmed,
					BookingDate: time.Now().AddDate(0, 0, 2)}, nil
			},
		},
		&testmock.BranchRepo{}, &testmock.TableTypeRepo{},
		&testmock.BookingTableRepo{}, nil, nil,
	)
	_, err := svc.Reschedule(1, 1, tomorrow(), "12:00")
	if err == nil {
		t.Fatal("expected error for non-pending booking")
	}
}

func TestReschedule_BookingDateTodayOrPast_CannotReschedule(t *testing.T) {
	// Booking is yesterday → past H-1 cutoff (booking date is not After today())
	yesterday := time.Now().AddDate(0, 0, -1)
	svc := newBookingSvc(
		&testmock.BookingRepo{
			FindByIDFn: func(id uint) (*model.Booking, error) {
				return &model.Booking{ID: id, CustomerID: 1, Status: model.BookingStatusPending,
					BookingDate: yesterday}, nil
			},
		},
		&testmock.BranchRepo{}, &testmock.TableTypeRepo{},
		&testmock.BookingTableRepo{}, nil, nil,
	)
	_, err := svc.Reschedule(1, 1, tomorrow(), "12:00")
	if err == nil {
		t.Fatal("expected error when booking date is in the past (past H-1)")
	}
}

func TestReschedule_NewDateInPast(t *testing.T) {
	svc := newBookingSvc(
		&testmock.BookingRepo{
			FindByIDFn: func(id uint) (*model.Booking, error) {
				return &model.Booking{ID: id, CustomerID: 1, Status: model.BookingStatusPending,
					BookingDate: time.Now().AddDate(0, 0, 3)}, nil
			},
		},
		&testmock.BranchRepo{
			FindByIDFn: func(id uint) (*model.Branch, error) { return activeBranch(id), nil },
		},
		&testmock.TableTypeRepo{}, &testmock.BookingTableRepo{}, nil, nil,
	)
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	_, err := svc.Reschedule(1, 1, yesterday, "12:00")
	if err == nil {
		t.Fatal("expected error for new date in the past")
	}
}

func TestReschedule_NewDateTooFarAhead(t *testing.T) {
	svc := newBookingSvc(
		&testmock.BookingRepo{
			FindByIDFn: func(id uint) (*model.Booking, error) {
				return &model.Booking{ID: id, CustomerID: 1, Status: model.BookingStatusPending,
					BookingDate: time.Now().AddDate(0, 0, 3)}, nil
			},
		},
		&testmock.BranchRepo{
			FindByIDFn: func(id uint) (*model.Branch, error) { return activeBranch(id), nil },
		},
		&testmock.TableTypeRepo{}, &testmock.BookingTableRepo{}, nil, nil,
	)
	farFuture := time.Now().AddDate(0, 0, 31).Format("2006-01-02")
	_, err := svc.Reschedule(1, 1, farFuture, "12:00")
	if err == nil {
		t.Fatal("expected error for date beyond 30 days")
	}
}

func TestReschedule_SlotNotAvailable(t *testing.T) {
	tt := activeTableType(1, 1, 4)
	svc := newBookingSvc(
		&testmock.BookingRepo{
			FindByIDFn: func(id uint) (*model.Booking, error) {
				return &model.Booking{
					ID: id, CustomerID: 1, BranchID: 1,
					Status: model.BookingStatusPending, TableTypeID: 1, TablesCount: 1,
					BookingDate: time.Now().AddDate(0, 0, 3),
				}, nil
			},
			CountOverlappingByGroupFn: func(ids []uint, date time.Time, start, end string, excludeID uint) (int64, error) {
				return 1, nil // fully booked
			},
		},
		&testmock.BranchRepo{
			FindByIDFn: func(id uint) (*model.Branch, error) { return activeBranch(id), nil },
		},
		&testmock.TableTypeRepo{
			FindByIDFn: func(id uint) (*model.TableType, error) { return tt, nil },
			FindByGroupFn: func(branchID uint, name string, capacity int, roomID *uint) ([]model.TableType, error) {
				return []model.TableType{*tt}, nil // only 1 table total, already booked
			},
		},
		&testmock.BookingTableRepo{}, nil, nil,
	)
	_, err := svc.Reschedule(1, 1, tomorrow(), "12:00")
	if err == nil {
		t.Fatal("expected error when slot is not available")
	}
}

func TestReschedule_Success(t *testing.T) {
	tt := activeTableType(1, 1, 4)
	var updatedDate time.Time
	svc := newBookingSvc(
		&testmock.BookingRepo{
			FindByIDFn: func(id uint) (*model.Booking, error) {
				return &model.Booking{
					ID: id, CustomerID: 1, BranchID: 1,
					Status: model.BookingStatusPending, TableTypeID: 1, TablesCount: 1,
					BookingDate: time.Now().AddDate(0, 0, 3),
				}, nil
			},
			CountOverlappingByGroupFn: func(ids []uint, date time.Time, start, end string, excludeID uint) (int64, error) {
				return 0, nil // slot free
			},
			UpdateScheduleFn: func(id uint, date time.Time, startTime, endTime string, status model.BookingStatus) error {
				updatedDate = date
				return nil
			},
		},
		&testmock.BranchRepo{
			FindByIDFn: func(id uint) (*model.Branch, error) { return activeBranch(id), nil },
		},
		&testmock.TableTypeRepo{
			FindByIDFn: func(id uint) (*model.TableType, error) { return tt, nil },
			FindByGroupFn: func(branchID uint, name string, capacity int, roomID *uint) ([]model.TableType, error) {
				return []model.TableType{*tt, *tt}, nil // 2 tables available
			},
		},
		&testmock.BookingTableRepo{}, nil, nil,
	)
	newDateStr := tomorrow()
	_, err := svc.Reschedule(1, 1, newDateStr, "14:00")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected, _ := time.ParseInLocation("2006-01-02", newDateStr, time.Local)
	if updatedDate.Format("2006-01-02") != expected.Format("2006-01-02") {
		t.Errorf("expected UpdateSchedule called with %s, got %s", newDateStr, updatedDate.Format("2006-01-02"))
	}
}
