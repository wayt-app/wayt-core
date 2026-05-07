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
	var list []model.Review
	var total int64
	q := r.db.Model(&model.Review{}).Where("restaurant_id = ?", restaurantID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Preload("Customer").Order("id desc").Limit(limit).Offset(offset).Find(&list).Error
	return list, total, err
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
