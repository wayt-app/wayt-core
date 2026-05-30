package service_test

import (
	"errors"
	"testing"
	"time"

	"github.com/wayt-app/wayt-core/internal/testmock"
	"github.com/wayt-app/wayt-core/model"
)

// --- Confirm ---

func TestConfirm_BookingNotFound(t *testing.T) {
	svc := newBookingSvc(
		&testmock.BookingRepo{
			FindByIDFn: func(id uint) (*model.Booking, error) { return nil, errors.New("not found") },
		},
		&testmock.BranchRepo{}, &testmock.TableTypeRepo{},
		&testmock.BookingTableRepo{}, nil, nil,
	)
	if err := svc.Confirm(99); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestConfirm_NotPending(t *testing.T) {
	svc := newBookingSvc(
		&testmock.BookingRepo{
			FindByIDFn: func(id uint) (*model.Booking, error) {
				return &model.Booking{ID: id, Status: model.BookingStatusConfirmed}, nil
			},
		},
		&testmock.BranchRepo{}, &testmock.TableTypeRepo{},
		&testmock.BookingTableRepo{}, nil, nil,
	)
	if err := svc.Confirm(1); err == nil {
		t.Fatal("expected error for non-pending booking")
	}
}

func TestConfirm_Success(t *testing.T) {
	var updatedID uint
	var updatedStatus model.BookingStatus
	svc := newBookingSvc(
		&testmock.BookingRepo{
			FindByIDFn: func(id uint) (*model.Booking, error) {
				return &model.Booking{ID: id, Status: model.BookingStatusPending}, nil
			},
			UpdateStatusFn: func(id uint, status model.BookingStatus) error {
				updatedID = id
				updatedStatus = status
				return nil
			},
		},
		&testmock.BranchRepo{}, &testmock.TableTypeRepo{},
		&testmock.BookingTableRepo{}, nil, nil,
	)
	if err := svc.Confirm(5); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updatedID != 5 || updatedStatus != model.BookingStatusConfirmed {
		t.Errorf("UpdateStatus not called correctly: id=%d status=%s", updatedID, updatedStatus)
	}
}

// --- Complete ---

func TestComplete_BookingNotFound(t *testing.T) {
	svc := newBookingSvc(
		&testmock.BookingRepo{
			FindByIDFn: func(id uint) (*model.Booking, error) { return nil, errors.New("not found") },
		},
		&testmock.BranchRepo{}, &testmock.TableTypeRepo{},
		&testmock.BookingTableRepo{}, nil, nil,
	)
	if err := svc.Complete(1, "", 0); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestComplete_WrongStatus(t *testing.T) {
	svc := newBookingSvc(
		&testmock.BookingRepo{
			FindByIDFn: func(id uint) (*model.Booking, error) {
				return &model.Booking{ID: id, Status: model.BookingStatusPending}, nil
			},
		},
		&testmock.BranchRepo{}, &testmock.TableTypeRepo{},
		&testmock.BookingTableRepo{}, nil, nil,
	)
	if err := svc.Complete(1, "", 0); err == nil {
		t.Fatal("expected error for wrong status")
	}
}

func TestComplete_NegativeBill(t *testing.T) {
	svc := newBookingSvc(
		&testmock.BookingRepo{
			FindByIDFn: func(id uint) (*model.Booking, error) {
				return &model.Booking{ID: id, Status: model.BookingStatusConfirmed}, nil
			},
		},
		&testmock.BranchRepo{}, &testmock.TableTypeRepo{},
		&testmock.BookingTableRepo{}, nil, nil,
	)
	if err := svc.Complete(1, "", -100); err == nil {
		t.Fatal("expected error for negative bill")
	}
}

func TestComplete_Success(t *testing.T) {
	var calledWith int64
	svc := newBookingSvc(
		&testmock.BookingRepo{
			FindByIDFn: func(id uint) (*model.Booking, error) {
				return &model.Booking{ID: id, Status: model.BookingStatusCheckedIn}, nil
			},
			CompleteWithDetailsFn: func(id uint, notes string, totalBill int64) error {
				calledWith = totalBill
				return nil
			},
		},
		&testmock.BranchRepo{}, &testmock.TableTypeRepo{},
		&testmock.BookingTableRepo{}, nil, nil,
	)
	if err := svc.Complete(1, "test notes", 150000); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if calledWith != 150000 {
		t.Errorf("expected totalBill 150000, got %d", calledWith)
	}
}

// --- AdminCancel ---

func TestAdminCancel_BookingNotFound(t *testing.T) {
	svc := newBookingSvc(
		&testmock.BookingRepo{
			FindByIDFn: func(id uint) (*model.Booking, error) { return nil, errors.New("not found") },
		},
		&testmock.BranchRepo{}, &testmock.TableTypeRepo{},
		&testmock.BookingTableRepo{}, nil, nil,
	)
	if err := svc.AdminCancel(1, ""); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestAdminCancel_AlreadyCompleted(t *testing.T) {
	svc := newBookingSvc(
		&testmock.BookingRepo{
			FindByIDFn: func(id uint) (*model.Booking, error) {
				return &model.Booking{ID: id, Status: model.BookingStatusCompleted}, nil
			},
		},
		&testmock.BranchRepo{}, &testmock.TableTypeRepo{},
		&testmock.BookingTableRepo{}, nil, nil,
	)
	if err := svc.AdminCancel(1, "reason"); err == nil {
		t.Fatal("expected error for completed booking")
	}
}

func TestAdminCancel_Success(t *testing.T) {
	var gotStatus model.BookingStatus
	svc := newBookingSvc(
		&testmock.BookingRepo{
			FindByIDFn: func(id uint) (*model.Booking, error) {
				return &model.Booking{ID: id, Status: model.BookingStatusConfirmed}, nil
			},
			UpdateStatusAndReasonFn: func(id uint, status model.BookingStatus, reason string) error {
				gotStatus = status
				return nil
			},
			FindWaitingListForSlotFn: func(branchID uint, date time.Time, startTime, endTime string) ([]model.Booking, error) {
				return nil, nil
			},
		},
		&testmock.BranchRepo{
			FindByIDFn: func(id uint) (*model.Branch, error) { return activeBranch(id), nil },
		},
		&testmock.TableTypeRepo{
			FindByBranchFn: func(branchID uint) ([]model.TableType, error) { return nil, nil },
		},
		&testmock.BookingTableRepo{}, nil, nil,
	)
	if err := svc.AdminCancel(1, "closed"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotStatus != model.BookingStatusCancelled {
		t.Errorf("expected status cancelled, got %s", gotStatus)
	}
}

// --- CheckIn ---

func TestCheckIn_BookingNotFound(t *testing.T) {
	svc := newBookingSvc(
		&testmock.BookingRepo{
			FindByIDFn: func(id uint) (*model.Booking, error) { return nil, errors.New("not found") },
		},
		&testmock.BranchRepo{}, &testmock.TableTypeRepo{},
		&testmock.BookingTableRepo{}, nil, nil,
	)
	if err := svc.CheckIn(1); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestCheckIn_NotConfirmed(t *testing.T) {
	svc := newBookingSvc(
		&testmock.BookingRepo{
			FindByIDFn: func(id uint) (*model.Booking, error) {
				return &model.Booking{ID: id, Status: model.BookingStatusPending}, nil
			},
		},
		&testmock.BranchRepo{}, &testmock.TableTypeRepo{},
		&testmock.BookingTableRepo{}, nil, nil,
	)
	if err := svc.CheckIn(1); err == nil {
		t.Fatal("expected error for non-confirmed booking")
	}
}

func TestCheckIn_Success(t *testing.T) {
	var gotStatus model.BookingStatus
	svc := newBookingSvc(
		&testmock.BookingRepo{
			FindByIDFn: func(id uint) (*model.Booking, error) {
				return &model.Booking{ID: id, Status: model.BookingStatusConfirmed}, nil
			},
			UpdateStatusFn: func(id uint, status model.BookingStatus) error {
				gotStatus = status
				return nil
			},
		},
		&testmock.BranchRepo{}, &testmock.TableTypeRepo{},
		&testmock.BookingTableRepo{}, nil, nil,
	)
	if err := svc.CheckIn(3); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotStatus != model.BookingStatusCheckedIn {
		t.Errorf("expected checked_in, got %s", gotStatus)
	}
}

// --- UpdateOrderStatus ---

func TestUpdateOrderStatus_InvalidStatus(t *testing.T) {
	svc := newBookingSvc(
		&testmock.BookingRepo{}, &testmock.BranchRepo{},
		&testmock.TableTypeRepo{}, &testmock.BookingTableRepo{}, nil, nil,
	)
	if err := svc.UpdateOrderStatus(1, 1, "unknown"); err == nil {
		t.Fatal("expected error for invalid order status")
	}
}

func TestUpdateOrderStatus_BookingNotFound(t *testing.T) {
	svc := newBookingSvc(
		&testmock.BookingRepo{
			FindByIDFn: func(id uint) (*model.Booking, error) { return nil, errors.New("not found") },
		},
		&testmock.BranchRepo{}, &testmock.TableTypeRepo{},
		&testmock.BookingTableRepo{}, nil, nil,
	)
	if err := svc.UpdateOrderStatus(1, 1, "prepare"); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestUpdateOrderStatus_WrongBranch(t *testing.T) {
	svc := newBookingSvc(
		&testmock.BookingRepo{
			FindByIDFn: func(id uint) (*model.Booking, error) {
				return &model.Booking{ID: id, BranchID: 5}, nil
			},
		},
		&testmock.BranchRepo{}, &testmock.TableTypeRepo{},
		&testmock.BookingTableRepo{}, nil, nil,
	)
	if err := svc.UpdateOrderStatus(1, 99, "prepare"); err == nil {
		t.Fatal("expected error for wrong branch")
	}
}

func TestUpdateOrderStatus_ValidStatuses(t *testing.T) {
	for _, status := range []string{"new", "prepare", "ready", "done"} {
		svc := newBookingSvc(
			&testmock.BookingRepo{
				FindByIDFn: func(id uint) (*model.Booking, error) {
					return &model.Booking{ID: id, BranchID: 1}, nil
				},
			},
			&testmock.BranchRepo{}, &testmock.TableTypeRepo{},
			&testmock.BookingTableRepo{}, nil, nil,
		)
		if err := svc.UpdateOrderStatus(1, 1, status); err != nil {
			t.Errorf("status %q should be valid, got error: %v", status, err)
		}
	}
}

// --- AdminUpdate ---

func TestAdminUpdate_BookingNotFound(t *testing.T) {
	svc := newBookingSvc(
		&testmock.BookingRepo{
			FindByIDFn: func(id uint) (*model.Booking, error) { return nil, errors.New("not found") },
		},
		&testmock.BranchRepo{}, &testmock.TableTypeRepo{},
		&testmock.BookingTableRepo{}, nil, nil,
	)
	_, err := svc.AdminUpdate(1, tomorrow(), "12:00", "", 2)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestAdminUpdate_WrongStatus(t *testing.T) {
	svc := newBookingSvc(
		&testmock.BookingRepo{
			FindByIDFn: func(id uint) (*model.Booking, error) {
				return &model.Booking{ID: id, Status: model.BookingStatusCompleted}, nil
			},
		},
		&testmock.BranchRepo{}, &testmock.TableTypeRepo{},
		&testmock.BookingTableRepo{}, nil, nil,
	)
	_, err := svc.AdminUpdate(1, tomorrow(), "12:00", "", 2)
	if err == nil {
		t.Fatal("expected error for completed booking")
	}
}

func TestAdminUpdate_ZeroGuestCount(t *testing.T) {
	svc := newBookingSvc(
		&testmock.BookingRepo{
			FindByIDFn: func(id uint) (*model.Booking, error) {
				return &model.Booking{ID: id, BranchID: 1, Status: model.BookingStatusPending, TableTypeID: 1}, nil
			},
		},
		&testmock.BranchRepo{
			FindByIDFn: func(id uint) (*model.Branch, error) { return activeBranch(id), nil },
		},
		&testmock.TableTypeRepo{}, &testmock.BookingTableRepo{}, nil, nil,
	)
	_, err := svc.AdminUpdate(1, tomorrow(), "12:00", "", 0)
	if err == nil {
		t.Fatal("expected error for zero guest count")
	}
}

func TestAdminUpdate_SlotFull(t *testing.T) {
	tt := activeTableType(1, 1, 4)
	svc := newBookingSvc(
		&testmock.BookingRepo{
			FindByIDFn: func(id uint) (*model.Booking, error) {
				return &model.Booking{ID: id, BranchID: 1, Status: model.BookingStatusPending, TableTypeID: 1}, nil
			},
			CountOverlappingByGroupFn: func(ids []uint, date time.Time, start, end string, excludeID uint) (int64, error) {
				return 1, nil // 1 already booked, only 1 table total → full
			},
		},
		&testmock.BranchRepo{
			FindByIDFn: func(id uint) (*model.Branch, error) { return activeBranch(id), nil },
		},
		&testmock.TableTypeRepo{
			FindByIDFn: func(id uint) (*model.TableType, error) { return tt, nil },
			FindByGroupFn: func(branchID uint, name string, capacity int, roomID *uint) ([]model.TableType, error) {
				return []model.TableType{*tt}, nil
			},
		},
		&testmock.BookingTableRepo{}, nil, nil,
	)
	_, err := svc.AdminUpdate(1, tomorrow(), "12:00", "", 2)
	if err == nil {
		t.Fatal("expected error when slot is full")
	}
}

func TestAdminUpdate_Success(t *testing.T) {
	tt := activeTableType(1, 1, 4)
	svc := newBookingSvc(
		&testmock.BookingRepo{
			FindByIDFn: func(id uint) (*model.Booking, error) {
				return &model.Booking{ID: id, BranchID: 1, Status: model.BookingStatusPending, TableTypeID: 1}, nil
			},
			CountOverlappingByGroupFn: func(ids []uint, date time.Time, start, end string, excludeID uint) (int64, error) {
				return 0, nil
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
	_, err := svc.AdminUpdate(1, tomorrow(), "12:00", "notes", 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
