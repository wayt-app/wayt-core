package repository

import (
	"time"

	"github.com/wayt/wayt-core/internal/model"
	"gorm.io/gorm"
)

type CustomerRepository interface {
	Create(c *model.Customer) error
	FindByEmail(email string) (*model.Customer, error)
	FindByID(id uint) (*model.Customer, error)
	UpdatePassword(id uint, hashedPassword string) error
	List() ([]model.Customer, error)
	SetResetToken(id uint, token string, expiresAt time.Time) error
	FindByResetToken(token string) (*model.Customer, error)
	ClearResetToken(id uint) error
	SetVerificationToken(id uint, token string) error
	FindByVerificationToken(token string) (*model.Customer, error)
	MarkVerified(id uint) error
	FindTokenVersion(id uint) (int, error)
	IncrementTokenVersion(id uint) error
	UpdateProfile(id uint, name, phone string) error
	FindByGoogleID(googleID string) (*model.Customer, error)
	SetGoogleInfo(id uint, googleID, avatarURL string) error
}

type customerRepository struct{ db *gorm.DB }

func NewCustomerRepository(db *gorm.DB) CustomerRepository {
	return &customerRepository{db: db}
}

func (r *customerRepository) Create(c *model.Customer) error {
	return r.db.Create(c).Error
}

func (r *customerRepository) FindByEmail(email string) (*model.Customer, error) {
	var c model.Customer
	return &c, r.db.Where("email = ?", email).First(&c).Error
}

func (r *customerRepository) FindByID(id uint) (*model.Customer, error) {
	var c model.Customer
	return &c, r.db.Where("id = ?", id).First(&c).Error
}

func (r *customerRepository) UpdatePassword(id uint, hashedPassword string) error {
	return r.db.Model(&model.Customer{}).Where("id = ?", id).Update("password", hashedPassword).Error
}

func (r *customerRepository) List() ([]model.Customer, error) {
	var list []model.Customer
	return list, r.db.Order("id desc").Find(&list).Error
}

func (r *customerRepository) SetResetToken(id uint, token string, expiresAt time.Time) error {
	return r.db.Model(&model.Customer{}).Where("id = ?", id).
		Updates(map[string]interface{}{"reset_token": token, "reset_token_expires_at": expiresAt}).Error
}

func (r *customerRepository) FindByResetToken(token string) (*model.Customer, error) {
	var c model.Customer
	return &c, r.db.Where("reset_token = ?", token).First(&c).Error
}

func (r *customerRepository) ClearResetToken(id uint) error {
	return r.db.Model(&model.Customer{}).Where("id = ?", id).
		Updates(map[string]interface{}{"reset_token": nil, "reset_token_expires_at": nil}).Error
}

func (r *customerRepository) SetVerificationToken(id uint, token string) error {
	return r.db.Model(&model.Customer{}).Where("id = ?", id).
		Update("verification_token", token).Error
}

func (r *customerRepository) FindByVerificationToken(token string) (*model.Customer, error) {
	var c model.Customer
	return &c, r.db.Where("verification_token = ?", token).First(&c).Error
}

func (r *customerRepository) MarkVerified(id uint) error {
	return r.db.Exec(
		"UPDATE tabl_customers SET is_verified = TRUE, verification_token = NULL, updated_at = NOW() WHERE id = ?",
		id,
	).Error
}

func (r *customerRepository) UpdateProfile(id uint, name, phone string) error {
	return r.db.Model(&model.Customer{}).Where("id = ?", id).
		Updates(map[string]interface{}{"name": name, "phone": phone}).Error
}

func (r *customerRepository) FindTokenVersion(id uint) (int, error) {
	var v int
	err := r.db.Raw("SELECT token_version FROM tabl_customers WHERE id = ?", id).Scan(&v).Error
	return v, err
}

func (r *customerRepository) IncrementTokenVersion(id uint) error {
	return r.db.Exec("UPDATE tabl_customers SET token_version = token_version + 1 WHERE id = ?", id).Error
}

func (r *customerRepository) FindByGoogleID(googleID string) (*model.Customer, error) {
	var c model.Customer
	return &c, r.db.Where("google_id = ?", googleID).First(&c).Error
}

func (r *customerRepository) SetGoogleInfo(id uint, googleID, avatarURL string) error {
	return r.db.Model(&model.Customer{}).Where("id = ?", id).
		Updates(map[string]interface{}{"google_id": googleID, "avatar_url": avatarURL}).Error
}
