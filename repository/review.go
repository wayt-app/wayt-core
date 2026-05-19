package repository

import (
	"github.com/wayt-app/wayt-core/model"
	"gorm.io/gorm"
)

type ReviewRepository interface {
	Create(r *model.Review) error
	FindByBookingID(bookingID uint) (*model.Review, error)
	FindByRestaurantID(restaurantID uint, limit, offset int) ([]model.Review, int64, error)
	StatsByRestaurantID(restaurantID uint) (avgRating float64, total int64, err error)
	FindByBranchID(branchID, restaurantID uint, limit, offset int) ([]model.Review, int64, error)
	StatsByBranchID(branchID, restaurantID uint) (avgRating float64, total int64, err error)
}

type reviewRepository struct{ db *gorm.DB }

func NewReviewRepository(db *gorm.DB) ReviewRepository {
	return &reviewRepository{db: db}
}

func (r *reviewRepository) Create(rv *model.Review) error {
	return r.db.Create(rv).Error
}

func (r *reviewRepository) FindByBookingID(bookingID uint) (*model.Review, error) {
	var rv model.Review
	return &rv, r.db.Where("booking_id = ?", bookingID).First(&rv).Error
}

func (r *reviewRepository) FindByRestaurantID(restaurantID uint, limit, offset int) ([]model.Review, int64, error) {
	return r.findPaged("restaurant_id = ?", []interface{}{restaurantID}, limit, offset)
}

func (r *reviewRepository) StatsByRestaurantID(restaurantID uint) (float64, int64, error) {
	var result struct {
		Avg   float64
		Total int64
	}
	err := r.db.Raw(
		`SELECT COALESCE(AVG(rating), 0) AS avg, COUNT(*) AS total FROM tabl_reviews WHERE restaurant_id = ?`,
		restaurantID,
	).Scan(&result).Error
	return result.Avg, result.Total, err
}

func (r *reviewRepository) FindByBranchID(branchID, restaurantID uint, limit, offset int) ([]model.Review, int64, error) {
	return r.findPaged("branch_id = ? AND restaurant_id = ?", []interface{}{branchID, restaurantID}, limit, offset)
}

// findPaged uses count(*) OVER() to get the total count alongside the page rows in a single
// DB round-trip, then preloads associations with one IN query each (same as GORM Preload).
// Net result: 3 queries instead of 4 (no separate COUNT query).
func (r *reviewRepository) findPaged(where string, args []interface{}, limit, offset int) ([]model.Review, int64, error) {
	type row struct {
		model.Review
		TotalCount int64 `gorm:"column:total_count"`
	}
	query := "SELECT *, count(*) OVER() AS total_count FROM tabl_reviews WHERE " +
		where + " ORDER BY id DESC LIMIT ? OFFSET ?"
	params := append(args, limit, offset)
	var rows []row
	if err := r.db.Raw(query, params...).Scan(&rows).Error; err != nil {
		return nil, 0, err
	}
	if len(rows) == 0 {
		return nil, 0, nil
	}
	total := rows[0].TotalCount
	ids := make([]uint, len(rows))
	list := make([]model.Review, len(rows))
	for i, rw := range rows {
		list[i] = rw.Review
		ids[i] = rw.ID
	}
	if err := r.db.Preload("Customer").Preload("Branch").
		Where("id IN ?", ids).Order("id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *reviewRepository) StatsByBranchID(branchID, restaurantID uint) (float64, int64, error) {
	var result struct {
		Avg   float64
		Total int64
	}
	err := r.db.Raw(
		`SELECT COALESCE(AVG(rating), 0) AS avg, COUNT(*) AS total FROM tabl_reviews WHERE branch_id = ? AND restaurant_id = ?`,
		branchID, restaurantID,
	).Scan(&result).Error
	return result.Avg, result.Total, err
}
