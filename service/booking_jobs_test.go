package service_test

import (
	"errors"
	"testing"
	"time"

	"github.com/wayt-app/wayt-core/internal/testmock"
	"github.com/wayt-app/wayt-core/model"
)

// --- ProcessNoShows ---

func TestProcessNoShows_RepoError(t *testing.T) {
	svc := newBookingSvc(
		&testmock.BookingRepo{
			FindNoShowCandidatesFn: func(graceMinutes int) ([]model.Booking, error) {
				return nil, errors.New("db error")
			},
		},
		&testmock.BranchRepo{}, &testmock.TableTypeRepo{},
		&testmock.BookingTableRepo{}, nil, nil,
	)
	if err := svc.ProcessNoShows(); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestProcessNoShows_MarksNoShow(t *testing.T) {
	var markedIDs []uint
	svc := newBookingSvc(
		&testmock.BookingRepo{
			FindNoShowCandidatesFn: func(graceMinutes int) ([]model.Booking, error) {
				return []model.Booking{
					{ID: 10, BranchID: 1, Status: model.BookingStatusConfirmed},
					{ID: 11, BranchID: 1, Status: model.BookingStatusConfirmed},
				}, nil
			},
			UpdateStatusFn: func(id uint, status model.BookingStatus) error {
				if status == model.BookingStatusNoShow {
					markedIDs = append(markedIDs, id)
				}
				return nil
			},
			FindWaitingListForSlotFn: func(branchID uint, date time.Time, startTime, endTime string) ([]model.Booking, error) {
				return nil, nil // no waiters
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
	if err := svc.ProcessNoShows(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(markedIDs) != 2 {
		t.Errorf("expected 2 bookings marked no_show, got %d", len(markedIDs))
	}
}

func TestProcessNoShows_SkipsOnUpdateError(t *testing.T) {
	var markedIDs []uint
	svc := newBookingSvc(
		&testmock.BookingRepo{
			FindNoShowCandidatesFn: func(graceMinutes int) ([]model.Booking, error) {
				return []model.Booking{
					{ID: 10, Status: model.BookingStatusConfirmed},
					{ID: 11, Status: model.BookingStatusConfirmed},
				}, nil
			},
			UpdateStatusFn: func(id uint, status model.BookingStatus) error {
				if id == 10 {
					return errors.New("db error") // fail first
				}
				markedIDs = append(markedIDs, id)
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
	if err := svc.ProcessNoShows(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// ID 10 failed → skipped, ID 11 succeeded
	if len(markedIDs) != 1 || markedIDs[0] != 11 {
		t.Errorf("expected only booking #11 marked, got %v", markedIDs)
	}
}

// --- ProcessReminders ---

func TestProcessReminders_RepoError(t *testing.T) {
	svc := newBookingSvc(
		&testmock.BookingRepo{
			FindReminderCandidatesFn: func() ([]model.Booking, error) {
				return nil, errors.New("db error")
			},
		},
		&testmock.BranchRepo{}, &testmock.TableTypeRepo{},
		&testmock.BookingTableRepo{}, nil, nil,
	)
	if err := svc.ProcessReminders(); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestProcessReminders_MarksReminderSent(t *testing.T) {
	var markedIDs []uint
	svc := newBookingSvc(
		&testmock.BookingRepo{
			FindReminderCandidatesFn: func() ([]model.Booking, error) {
				return []model.Booking{
					{ID: 20, CustomerID: 1, BranchID: 1},
					{ID: 21, CustomerID: 2, BranchID: 1},
				}, nil
			},
			MarkReminderSentFn: func(id uint) error {
				markedIDs = append(markedIDs, id)
				return nil
			},
		},
		&testmock.BranchRepo{
			FindByIDFn: func(id uint) (*model.Branch, error) { return activeBranch(id), nil },
		},
		&testmock.TableTypeRepo{}, &testmock.BookingTableRepo{}, nil, nil,
	)
	if err := svc.ProcessReminders(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(markedIDs) != 2 {
		t.Errorf("expected 2 reminder-sent marks, got %d", len(markedIDs))
	}
}

func TestProcessReminders_ContinuesOnMarkError(t *testing.T) {
	// MarkReminderSent fails for one booking → should not stop loop
	var markedIDs []uint
	svc := newBookingSvc(
		&testmock.BookingRepo{
			FindReminderCandidatesFn: func() ([]model.Booking, error) {
				return []model.Booking{
					{ID: 30, CustomerID: 1, BranchID: 1},
					{ID: 31, CustomerID: 2, BranchID: 1},
				}, nil
			},
			MarkReminderSentFn: func(id uint) error {
				if id == 30 {
					return errors.New("mark failed")
				}
				markedIDs = append(markedIDs, id)
				return nil
			},
		},
		&testmock.BranchRepo{
			FindByIDFn: func(id uint) (*model.Branch, error) { return activeBranch(id), nil },
		},
		&testmock.TableTypeRepo{}, &testmock.BookingTableRepo{}, nil, nil,
	)
	if err := svc.ProcessReminders(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// ID 30 failed silently, ID 31 succeeded
	if len(markedIDs) != 1 || markedIDs[0] != 31 {
		t.Errorf("expected booking #31 marked, got %v", markedIDs)
	}
}

func TestProcessReminders_EmptyCandidates(t *testing.T) {
	svc := newBookingSvc(
		&testmock.BookingRepo{
			FindReminderCandidatesFn: func() ([]model.Booking, error) {
				return []model.Booking{}, nil
			},
		},
		&testmock.BranchRepo{}, &testmock.TableTypeRepo{},
		&testmock.BookingTableRepo{}, nil, nil,
	)
	if err := svc.ProcessReminders(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
