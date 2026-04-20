package repository

import (
	"github.com/wayt/wayt-core/internal/model"
	"gorm.io/gorm"
)

type BranchRepository interface {
	Create(b *model.Branch) error
	FindAll() ([]model.Branch, error)
	FindByRestaurant(restaurantID uint) ([]model.Branch, error)
	FindActiveByRestaurant(restaurantID uint) ([]model.Branch, error)
	FindByID(id uint) (*model.Branch, error)
	Update(b *model.Branch) error
	Delete(id uint) error
}

type branchRepository struct{ db *gorm.DB }

func NewBranchRepository(db *gorm.DB) BranchRepository {
	return &branchRepository{db: db}
}

func (r *branchRepository) Create(b *model.Branch) error {
	return r.db.Create(b).Error
}

func (r *branchRepository) FindAll() ([]model.Branch, error) {
	var list []model.Branch
	err := r.db.Where("deleted_at IS NULL").Order("restaurant_id ASC, id ASC").Find(&list).Error
	return list, err
}

func (r *branchRepository) FindByRestaurant(restaurantID uint) ([]model.Branch, error) {
	var list []model.Branch
	err := r.db.Where("restaurant_id = ? AND deleted_at IS NULL", restaurantID).
		Order("id ASC").Find(&list).Error
	return list, err
}

func (r *branchRepository) FindActiveByRestaurant(restaurantID uint) ([]model.Branch, error) {
	var list []model.Branch
	err := r.db.Where("restaurant_id = ? AND deleted_at IS NULL AND is_active = true", restaurantID).
		Order("id ASC").Find(&list).Error
	return list, err
}

func (r *branchRepository) FindByID(id uint) (*model.Branch, error) {
	var b model.Branch
	err := r.db.Where("id = ? AND deleted_at IS NULL", id).First(&b).Error
	return &b, err
}

func (r *branchRepository) Update(b *model.Branch) error {
	return r.db.Save(b).Error
}

func (r *branchRepository) Delete(id uint) error {
	return r.db.Model(&model.Branch{}).Where("id = ?", id).
		Update("deleted_at", gorm.Expr("NOW()")).Error
}
