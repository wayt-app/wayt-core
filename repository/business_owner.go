package repository

import (
	"time"

	"github.com/wayt-app/wayt-core/model"
	"gorm.io/gorm"
)

type BusinessOwnerRepository interface {
	Create(o *model.BusinessOwner) error
	FindByEmail(email string) (*model.BusinessOwner, error)
	FindByID(id uint) (*model.BusinessOwner, error)
	UpdatePassword(id uint, hashedPassword string) error
	SetVerificationToken(id uint, token string) error
	FindByVerificationToken(token string) (*model.BusinessOwner, error)
	MarkVerified(id uint) error
	SetResetToken(id uint, token string, expiresAt time.Time) error
	FindByResetToken(token string) (*model.BusinessOwner, error)
	ClearResetToken(id uint) error
	List() ([]model.BusinessOwner, error)
	FindByRestaurantID(restaurantID uint) (*model.BusinessOwner, error)
	FindTokenVersion(id uint) (int, error)
	IncrementTokenVersion(id uint) error
	FindByGoogleID(googleID string) (*model.BusinessOwner, error)
	SetGoogleInfo(id uint, googleID, avatarURL string) error
}

type businessOwnerRepository struct{ db *gorm.DB }

func NewBusinessOwnerRepository(db *gorm.DB) BusinessOwnerRepository {
	return &businessOwnerRepository{db: db}
}

func (r *businessOwnerRepository) Create(o *model.BusinessOwner) error {
	return r.db.Create(o).Error
}

func (r *businessOwnerRepository) FindByEmail(email string) (*model.BusinessOwner, error) {
	var o model.BusinessOwner
	return &o, r.db.Where("email = ?", email).First(&o).Error
}

func (r *businessOwnerRepository) FindByID(id uint) (*model.BusinessOwner, error) {
	var o model.BusinessOwner
	return &o, r.db.Where("id = ?", id).First(&o).Error
}

func (r *businessOwnerRepository) UpdatePassword(id uint, hashedPassword string) error {
	return r.db.Model(&model.BusinessOwner{}).Where("id = ?", id).
		Updates(map[string]interface{}{"password": hashedPassword, "updated_at": time.Now()}).Error
}

func (r *businessOwnerRepository) SetVerificationToken(id uint, token string) error {
	return r.db.Model(&model.BusinessOwner{}).Where("id = ?", id).
		Update("verification_token", token).Error
}

func (r *businessOwnerRepository) FindByVerificationToken(token string) (*model.BusinessOwner, error) {
	var o model.BusinessOwner
	return &o, r.db.Where("verification_token = ?", token).First(&o).Error
}

func (r *businessOwnerRepository) MarkVerified(id uint) error {
	return r.db.Exec(
		"UPDATE tabl_business_owners SET is_verified = TRUE, verification_token = NULL, updated_at = NOW() WHERE id = ?",
		id,
	).Error
}

func (r *businessOwnerRepository) SetResetToken(id uint, token string, expiresAt time.Time) error {
	return r.db.Model(&model.BusinessOwner{}).Where("id = ?", id).
		Updates(map[string]interface{}{"reset_token": token, "reset_token_expires_at": expiresAt}).Error
}

func (r *businessOwnerRepository) FindByResetToken(token string) (*model.BusinessOwner, error) {
	var o model.BusinessOwner
	return &o, r.db.Where("reset_token = ?", token).First(&o).Error
}

func (r *businessOwnerRepository) ClearResetToken(id uint) error {
	return r.db.Model(&model.BusinessOwner{}).Where("id = ?", id).
		Updates(map[string]interface{}{"reset_token": nil, "reset_token_expires_at": nil}).Error
}

func (r *businessOwnerRepository) List() ([]model.BusinessOwner, error) {
	var list []model.BusinessOwner
	return list, r.db.Order("id desc").Find(&list).Error
}

func (r *businessOwnerRepository) FindTokenVersion(id uint) (int, error) {
	var v int
	err := r.db.Raw("SELECT token_version FROM tabl_business_owners WHERE id = ?", id).Scan(&v).Error
	return v, err
}

func (r *businessOwnerRepository) IncrementTokenVersion(id uint) error {
	return r.db.Exec("UPDATE tabl_business_owners SET token_version = token_version + 1 WHERE id = ?", id).Error
}

func (r *businessOwnerRepository) FindByGoogleID(googleID string) (*model.BusinessOwner, error) {
	var o model.BusinessOwner
	return &o, r.db.Where("google_id = ?", googleID).First(&o).Error
}

func (r *businessOwnerRepository) SetGoogleInfo(id uint, googleID, avatarURL string) error {
	return r.db.Model(&model.BusinessOwner{}).Where("id = ?", id).
		Updates(map[string]interface{}{"google_id": googleID, "avatar_url": avatarURL}).Error
}

func (r *businessOwnerRepository) FindByRestaurantID(restaurantID uint) (*model.BusinessOwner, error) {
	var o model.BusinessOwner
	err := r.db.Joins("JOIN tabl_restaurants r ON r.business_owner_id = tabl_business_owners.id").
		Where("r.id = ? AND r.deleted_at IS NULL", restaurantID).
		First(&o).Error
	return &o, err
}
