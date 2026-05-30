package service

import (
	"errors"

	"github.com/wayt-app/wayt-core/model"
)

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
