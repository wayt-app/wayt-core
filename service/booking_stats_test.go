package service_test

import (
	"errors"
	"testing"
	"time"

	"github.com/wayt-app/wayt-core/internal/testmock"
	"github.com/wayt-app/wayt-core/model"
	"github.com/wayt-app/wayt-core/repository"
)

// --- GetDashboardStats ---

func TestGetDashboardStats_InvalidDate(t *testing.T) {
	svc := newBookingSvc(
		&testmock.BookingRepo{}, &testmock.BranchRepo{},
		&testmock.TableTypeRepo{}, &testmock.BookingTableRepo{}, nil, nil,
	)
	_, err := svc.GetDashboardStats(1, "not-a-date")
	if err == nil {
		t.Fatal("expected error for invalid date")
	}
}

func TestGetDashboardStats_DefaultsToToday(t *testing.T) {
	var calledDate time.Time
	svc := newBookingSvc(
		&testmock.BookingRepo{
			CountByStatusForDateFn: func(branchID uint, date time.Time) (map[model.BookingStatus]int64, error) {
				calledDate = date
				return map[model.BookingStatus]int64{}, nil
			},
		},
		&testmock.BranchRepo{}, &testmock.TableTypeRepo{},
		&testmock.BookingTableRepo{}, nil, nil,
	)
	_, err := svc.GetDashboardStats(1, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	today := time.Now().Format("2006-01-02")
	if calledDate.Format("2006-01-02") != today {
		t.Errorf("expected today %s, got %s", today, calledDate.Format("2006-01-02"))
	}
}

func TestGetDashboardStats_CountsAndTotals(t *testing.T) {
	svc := newBookingSvc(
		&testmock.BookingRepo{
			CountByStatusForDateFn: func(branchID uint, date time.Time) (map[model.BookingStatus]int64, error) {
				return map[model.BookingStatus]int64{
					model.BookingStatusConfirmed: 3,
					model.BookingStatusNoShow:    1,
				}, nil
			},
		},
		&testmock.BranchRepo{}, &testmock.TableTypeRepo{},
		&testmock.BookingTableRepo{}, nil, nil,
	)
	stats, err := svc.GetDashboardStats(1, "2026-06-01")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats.TotalToday != 4 {
		t.Errorf("expected total 4, got %d", stats.TotalToday)
	}
	if stats.NoShowCount != 1 {
		t.Errorf("expected no_show 1, got %d", stats.NoShowCount)
	}
}

// --- GetRestaurantDashboard ---

func TestGetRestaurantDashboard_InvalidDate(t *testing.T) {
	svc := newBookingSvc(
		&testmock.BookingRepo{}, &testmock.BranchRepo{},
		&testmock.TableTypeRepo{}, &testmock.BookingTableRepo{}, nil, nil,
	)
	_, err := svc.GetRestaurantDashboard(1, "bad-date")
	if err == nil {
		t.Fatal("expected error for invalid date")
	}
}

func TestGetRestaurantDashboard_AggregatesBranches(t *testing.T) {
	svc := newBookingSvc(
		&testmock.BookingRepo{
			CountByStatusForDateFn: func(branchID uint, date time.Time) (map[model.BookingStatus]int64, error) {
				return map[model.BookingStatus]int64{
					model.BookingStatusConfirmed: int64(branchID), // branch 1 → 1, branch 2 → 2
				}, nil
			},
		},
		&testmock.BranchRepo{
			FindByRestaurantFn: func(restaurantID uint) ([]model.Branch, error) {
				return []model.Branch{
					{ID: 1, Name: "Cabang A"},
					{ID: 2, Name: "Cabang B"},
				}, nil
			},
		},
		&testmock.TableTypeRepo{}, &testmock.BookingTableRepo{}, nil, nil,
	)
	dash, err := svc.GetRestaurantDashboard(10, "2026-06-01")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(dash.Branches) != 2 {
		t.Fatalf("expected 2 branches, got %d", len(dash.Branches))
	}
	if dash.Branches[0].Total != 1 {
		t.Errorf("expected branch A total 1, got %d", dash.Branches[0].Total)
	}
	if dash.Branches[1].Total != 2 {
		t.Errorf("expected branch B total 2, got %d", dash.Branches[1].Total)
	}
}

// --- ListCustomersByRestaurant ---

func TestListCustomersByRestaurant_RepoError(t *testing.T) {
	svc := newBookingSvc(
		&testmock.BookingRepo{
			ListCustomerSummaryFn: func(restaurantID uint) ([]repository.CustomerSummaryRow, error) {
				return nil, errors.New("db error")
			},
		},
		&testmock.BranchRepo{}, &testmock.TableTypeRepo{},
		&testmock.BookingTableRepo{}, nil, nil,
	)
	_, err := svc.ListCustomersByRestaurant(1)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestListCustomersByRestaurant_SegmentClassification(t *testing.T) {
	svc := newBookingSvc(
		&testmock.BookingRepo{
			ListCustomerSummaryFn: func(restaurantID uint) ([]repository.CustomerSummaryRow, error) {
				return []repository.CustomerSummaryRow{
					{CustomerID: 1, Name: "Alice", TotalVisits: 1},  // new
					{CustomerID: 2, Name: "Bob", TotalVisits: 3},    // regular
					{CustomerID: 3, Name: "Charlie", TotalVisits: 7}, // loyal
				}, nil
			},
		},
		&testmock.BranchRepo{}, &testmock.TableTypeRepo{},
		&testmock.BookingTableRepo{}, nil, nil,
	)
	summaries, err := svc.ListCustomersByRestaurant(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(summaries) != 3 {
		t.Fatalf("expected 3 summaries, got %d", len(summaries))
	}
	cases := []struct {
		name    string
		segment string
	}{
		{"Alice", "new"},
		{"Bob", "regular"},
		{"Charlie", "loyal"},
	}
	for i, c := range cases {
		if summaries[i].Segment != c.segment {
			t.Errorf("%s: expected segment %q, got %q", c.name, c.segment, summaries[i].Segment)
		}
	}
}
