package repository

import (
	"github.com/wayt-app/wayt-core/model"
	"gorm.io/gorm"
)

type MediaRepository interface {
	Create(m *model.Media) error
	FindByRestaurant(restaurantID uint, mediaType string) ([]model.Media, error)
	FindMenuForBranch(restaurantID uint, branchID uint) ([]model.Media, error)
	FindByID(id uint) (*model.Media, error)
	DeleteByID(id uint) error
	DeleteLogoByRestaurant(restaurantID uint) ([]model.Media, error)
	DeleteBannerByRestaurant(restaurantID uint) ([]model.Media, error)
}

type mediaRepository struct{ db *gorm.DB }

func NewMediaRepository(db *gorm.DB) MediaRepository {
	return &mediaRepository{db: db}
}

func (r *mediaRepository) Create(m *model.Media) error {
	return r.db.Create(m).Error
}

func (r *mediaRepository) FindByRestaurant(restaurantID uint, mediaType string) ([]model.Media, error) {
	var list []model.Media
	q := r.db.Where("restaurant_id = ?", restaurantID)
	if mediaType != "" {
		q = q.Where("type = ?", mediaType)
	}
	err := q.Order("display_order ASC, id ASC").Find(&list).Error
	return list, err
}

// FindMenuForBranch returns menu images for a branch: images scoped to that branch + global images (branch_id IS NULL).
func (r *mediaRepository) FindMenuForBranch(restaurantID uint, branchID uint) ([]model.Media, error) {
	var list []model.Media
	err := r.db.Where("restaurant_id = ? AND type = 'menu' AND (branch_id = ? OR branch_id IS NULL)", restaurantID, branchID).
		Order("display_order ASC, id ASC").Find(&list).Error
	return list, err
}

func (r *mediaRepository) FindByID(id uint) (*model.Media, error) {
	var m model.Media
	err := r.db.First(&m, id).Error
	return &m, err
}

func (r *mediaRepository) DeleteByID(id uint) error {
	return r.db.Delete(&model.Media{}, id).Error
}

// DeleteLogoByRestaurant deletes existing logos and returns them (so storage paths can be cleaned up).
func (r *mediaRepository) DeleteLogoByRestaurant(restaurantID uint) ([]model.Media, error) {
	var existing []model.Media
	r.db.Where("restaurant_id = ? AND type = 'logo'", restaurantID).Find(&existing)
	if len(existing) > 0 {
		r.db.Where("restaurant_id = ? AND type = 'logo'", restaurantID).Delete(&model.Media{})
	}
	return existing, nil
}

func (r *mediaRepository) DeleteBannerByRestaurant(restaurantID uint) ([]model.Media, error) {
	var existing []model.Media
	r.db.Where("restaurant_id = ? AND type = 'banner'", restaurantID).Find(&existing)
	if len(existing) > 0 {
		r.db.Where("restaurant_id = ? AND type = 'banner'", restaurantID).Delete(&model.Media{})
	}
	return existing, nil
}
