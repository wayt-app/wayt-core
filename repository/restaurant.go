package repository

import (
	"github.com/wayt-app/wayt-core/model"
	"gorm.io/gorm"
)

type RestaurantRepository interface {
	Create(r *model.Restaurant) error
	FindAll() ([]model.Restaurant, error)
	FindAllActive() ([]model.Restaurant, error)
	FindAllActiveWithBranchCoords() ([]model.RestaurantWithCoords, error)
	FindByID(id uint) (*model.Restaurant, error)
	FindByOwnerID(ownerID uint) (*model.Restaurant, error)
	FindByBranchID(restaurantID uint) (*model.Restaurant, error)
	FindByPromoToken(token string) (*model.Restaurant, error)
	Update(r *model.Restaurant) error
	UpdateLogoURL(id uint, logoURL string) error
	UpdateBannerURL(id uint, bannerURL string) error
	Delete(id uint) error
}

type restaurantRepository struct{ db *gorm.DB }

func NewRestaurantRepository(db *gorm.DB) RestaurantRepository {
	return &restaurantRepository{db: db}
}

func (r *restaurantRepository) Create(rest *model.Restaurant) error {
	return r.db.Create(rest).Error
}

func (r *restaurantRepository) FindAll() ([]model.Restaurant, error) {
	var list []model.Restaurant
	err := r.db.Where("deleted_at IS NULL").Order("created_at ASC").Find(&list).Error
	return list, err
}

func (r *restaurantRepository) FindAllActive() ([]model.Restaurant, error) {
	var list []model.Restaurant
	err := r.db.Where("deleted_at IS NULL AND is_active = true").Order("name ASC").Find(&list).Error
	return list, err
}

func (r *restaurantRepository) FindAllActiveWithBranchCoords() ([]model.RestaurantWithCoords, error) {
	var list []model.RestaurantWithCoords
	err := r.db.Raw(`
		SELECT r.*,
			COALESCE((
				SELECT b.latitude FROM tabl_branches b
				WHERE b.restaurant_id = r.id AND b.is_active = true
				  AND b.deleted_at IS NULL AND b.latitude <> 0
				LIMIT 1
			), 0) AS nearest_lat,
			COALESCE((
				SELECT b.longitude FROM tabl_branches b
				WHERE b.restaurant_id = r.id AND b.is_active = true
				  AND b.deleted_at IS NULL AND b.latitude <> 0
				LIMIT 1
			), 0) AS nearest_lng
		FROM tabl_restaurants r
		WHERE r.deleted_at IS NULL AND r.is_active = true
		ORDER BY r.name ASC
	`).Scan(&list).Error
	return list, err
}

func (r *restaurantRepository) FindByID(id uint) (*model.Restaurant, error) {
	var rest model.Restaurant
	err := r.db.Where("id = ? AND deleted_at IS NULL", id).First(&rest).Error
	return &rest, err
}

func (r *restaurantRepository) FindByOwnerID(ownerID uint) (*model.Restaurant, error) {
	var rest model.Restaurant
	err := r.db.Where("business_owner_id = ? AND deleted_at IS NULL", ownerID).First(&rest).Error
	return &rest, err
}

func (r *restaurantRepository) Update(rest *model.Restaurant) error {
	return r.db.Save(rest).Error
}

func (r *restaurantRepository) FindByBranchID(restaurantID uint) (*model.Restaurant, error) {
	return r.FindByID(restaurantID)
}

func (r *restaurantRepository) FindByPromoToken(token string) (*model.Restaurant, error) {
	var rest model.Restaurant
	err := r.db.Where("promo_token = ? AND deleted_at IS NULL AND is_active = true", token).First(&rest).Error
	return &rest, err
}

func (r *restaurantRepository) UpdateLogoURL(id uint, logoURL string) error {
	return r.db.Model(&model.Restaurant{}).Where("id = ?", id).Update("logo_url", logoURL).Error
}

func (r *restaurantRepository) UpdateBannerURL(id uint, bannerURL string) error {
	return r.db.Model(&model.Restaurant{}).Where("id = ?", id).Update("banner_url", bannerURL).Error
}

func (r *restaurantRepository) Delete(id uint) error {
	return r.db.Model(&model.Restaurant{}).Where("id = ?", id).
		Update("deleted_at", gorm.Expr("NOW()")).Error
}
