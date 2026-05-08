package repository

import (
	"github.com/wayt-app/wayt-core/model"
	"gorm.io/gorm"
)

type MenuItemRepository interface {
	Create(m *model.MenuItem) error
	FindByBranch(branchID uint) ([]model.MenuItem, error)
	FindByID(id uint) (*model.MenuItem, error)
	Update(m *model.MenuItem) error
	Delete(id uint) error
}

type menuItemRepository struct{ db *gorm.DB }

func NewMenuItemRepository(db *gorm.DB) MenuItemRepository {
	return &menuItemRepository{db: db}
}

func (r *menuItemRepository) Create(m *model.MenuItem) error {
	return r.db.Create(m).Error
}

func (r *menuItemRepository) FindByBranch(branchID uint) ([]model.MenuItem, error) {
	var list []model.MenuItem
	err := r.db.Where("branch_id = ? AND deleted_at IS NULL", branchID).
		Order("category ASC, name ASC").Find(&list).Error
	return list, err
}

func (r *menuItemRepository) FindByID(id uint) (*model.MenuItem, error) {
	var m model.MenuItem
	err := r.db.Where("id = ? AND deleted_at IS NULL", id).First(&m).Error
	return &m, err
}

func (r *menuItemRepository) Update(m *model.MenuItem) error {
	return r.db.Save(m).Error
}

func (r *menuItemRepository) Delete(id uint) error {
	return r.db.Model(&model.MenuItem{}).Where("id = ?", id).
		Update("deleted_at", gorm.Expr("NOW()")).Error
}
