package repository

import (
	"time"

	"github.com/wayt/wayt-core/model"
	"gorm.io/gorm"
)

type StaffRepository interface {
	Create(s *model.Staff) error
	FindByEmail(email string) (*model.Staff, error)
	FindByID(id uint) (*model.Staff, error)
	FindByBranchID(branchID uint) ([]model.Staff, error)
	FindByOwnerID(ownerID uint) ([]model.Staff, error)
	Update(s *model.Staff) error
	Delete(id uint) error
	UpdatePassword(id uint, hashedPassword string) error
	SetResetToken(id uint, token string, expiresAt time.Time) error
	FindByResetToken(token string) (*model.Staff, error)
	ClearResetToken(id uint) error
	FindTokenVersion(id uint) (int, error)
	IncrementTokenVersion(id uint) error
}

type staffRepository struct{ db *gorm.DB }

func NewStaffRepository(db *gorm.DB) StaffRepository {
	return &staffRepository{db: db}
}

func (r *staffRepository) Create(s *model.Staff) error {
	return r.db.Create(s).Error
}

func (r *staffRepository) FindByEmail(email string) (*model.Staff, error) {
	var s model.Staff
	return &s, r.db.Where("email = ?", email).First(&s).Error
}

func (r *staffRepository) FindByID(id uint) (*model.Staff, error) {
	var s model.Staff
	return &s, r.db.Preload("Branch").Where("id = ?", id).First(&s).Error
}

func (r *staffRepository) FindByBranchID(branchID uint) ([]model.Staff, error) {
	var list []model.Staff
	return list, r.db.Where("branch_id = ? AND is_active = true", branchID).Order("id ASC").Find(&list).Error
}

func (r *staffRepository) FindByOwnerID(ownerID uint) ([]model.Staff, error) {
	var list []model.Staff
	return list, r.db.Preload("Branch").Where("business_owner_id = ?", ownerID).Order("id ASC").Find(&list).Error
}

func (r *staffRepository) Update(s *model.Staff) error {
	return r.db.Model(s).Omit("Branch").Save(s).Error
}

func (r *staffRepository) Delete(id uint) error {
	return r.db.Model(&model.Staff{}).Where("id = ?", id).
		Updates(map[string]interface{}{"is_active": false, "updated_at": time.Now()}).Error
}

func (r *staffRepository) UpdatePassword(id uint, hashedPassword string) error {
	return r.db.Model(&model.Staff{}).Where("id = ?", id).
		Updates(map[string]interface{}{"password": hashedPassword, "updated_at": time.Now()}).Error
}

func (r *staffRepository) SetResetToken(id uint, token string, expiresAt time.Time) error {
	return r.db.Model(&model.Staff{}).Where("id = ?", id).
		Updates(map[string]interface{}{"reset_token": token, "reset_token_expires_at": expiresAt, "updated_at": time.Now()}).Error
}

func (r *staffRepository) FindByResetToken(token string) (*model.Staff, error) {
	var s model.Staff
	err := r.db.Where("reset_token = ? AND reset_token_expires_at > ?", token, time.Now()).First(&s).Error
	return &s, err
}

func (r *staffRepository) ClearResetToken(id uint) error {
	return r.db.Model(&model.Staff{}).Where("id = ?", id).
		Updates(map[string]interface{}{"reset_token": nil, "reset_token_expires_at": nil, "updated_at": time.Now()}).Error
}

func (r *staffRepository) FindTokenVersion(id uint) (int, error) {
	var v int
	err := r.db.Raw("SELECT token_version FROM tabl_staff WHERE id = ?", id).Scan(&v).Error
	return v, err
}

func (r *staffRepository) IncrementTokenVersion(id uint) error {
	return r.db.Exec("UPDATE tabl_staff SET token_version = token_version + 1 WHERE id = ?", id).Error
}
