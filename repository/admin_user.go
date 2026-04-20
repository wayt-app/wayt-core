package repository

import (
	"time"

	"github.com/wayt/wayt-core/model"
	"gorm.io/gorm"
)

type AdminUserRepository interface {
	Create(user *model.AdminUser) error
	FindByUsername(username string) (*model.AdminUser, error)
	FindByID(id uint) (*model.AdminUser, error)
	FindAll() ([]model.AdminUser, error)
	Update(user *model.AdminUser) error
	Delete(id uint) error
	ExistsAny() (bool, error)
	SetResetToken(id uint, token string, expiresAt time.Time) error
	FindByResetToken(token string) (*model.AdminUser, error)
	ClearResetToken(id uint) error
	FindTokenVersion(id uint) (int, error)
	IncrementTokenVersion(id uint) error
}

type adminUserRepository struct{ db *gorm.DB }

func NewAdminUserRepository(db *gorm.DB) AdminUserRepository {
	return &adminUserRepository{db: db}
}

func (r *adminUserRepository) Create(user *model.AdminUser) error {
	return r.db.Create(user).Error
}

func (r *adminUserRepository) FindByUsername(username string) (*model.AdminUser, error) {
	var u model.AdminUser
	return &u, r.db.Where("username = ?", username).First(&u).Error
}

func (r *adminUserRepository) FindByID(id uint) (*model.AdminUser, error) {
	var u model.AdminUser
	return &u, r.db.Where("id = ?", id).First(&u).Error
}

func (r *adminUserRepository) FindAll() ([]model.AdminUser, error) {
	var users []model.AdminUser
	err := r.db.Order("created_at ASC").Find(&users).Error
	return users, err
}

func (r *adminUserRepository) Update(user *model.AdminUser) error {
	return r.db.Save(user).Error
}

func (r *adminUserRepository) Delete(id uint) error {
	return r.db.Delete(&model.AdminUser{}, id).Error
}

func (r *adminUserRepository) ExistsAny() (bool, error) {
	var count int64
	err := r.db.Model(&model.AdminUser{}).Count(&count).Error
	return count > 0, err
}

func (r *adminUserRepository) SetResetToken(id uint, token string, expiresAt time.Time) error {
	return r.db.Model(&model.AdminUser{}).Where("id = ?", id).
		Updates(map[string]interface{}{"reset_token": token, "reset_token_expires_at": expiresAt}).Error
}

func (r *adminUserRepository) FindByResetToken(token string) (*model.AdminUser, error) {
	var u model.AdminUser
	return &u, r.db.Where("reset_token = ?", token).First(&u).Error
}

func (r *adminUserRepository) ClearResetToken(id uint) error {
	return r.db.Model(&model.AdminUser{}).Where("id = ?", id).
		Updates(map[string]interface{}{"reset_token": nil, "reset_token_expires_at": nil}).Error
}

func (r *adminUserRepository) FindTokenVersion(id uint) (int, error) {
	var v int
	err := r.db.Raw("SELECT token_version FROM tabl_admin_users WHERE id = ?", id).Scan(&v).Error
	return v, err
}

func (r *adminUserRepository) IncrementTokenVersion(id uint) error {
	return r.db.Exec("UPDATE tabl_admin_users SET token_version = token_version + 1 WHERE id = ?", id).Error
}
