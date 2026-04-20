package repository

import (
	"github.com/wayt/wayt-core/model"
	"gorm.io/gorm"
)

type TableTypeRepository interface {
	Create(t *model.TableType) error
	FindByBranch(branchID uint) ([]model.TableType, error)
	FindByID(id uint) (*model.TableType, error)
	Update(t *model.TableType) error
	Delete(id uint) error
}

type tableTypeRepository struct{ db *gorm.DB }

func NewTableTypeRepository(db *gorm.DB) TableTypeRepository {
	return &tableTypeRepository{db: db}
}

func (r *tableTypeRepository) Create(t *model.TableType) error {
	return r.db.Create(t).Error
}

func (r *tableTypeRepository) FindByBranch(branchID uint) ([]model.TableType, error) {
	var list []model.TableType
	err := r.db.Where("branch_id = ? AND deleted_at IS NULL", branchID).
		Order("id ASC").Find(&list).Error
	return list, err
}

func (r *tableTypeRepository) FindByID(id uint) (*model.TableType, error) {
	var t model.TableType
	err := r.db.Where("id = ? AND deleted_at IS NULL", id).First(&t).Error
	return &t, err
}

func (r *tableTypeRepository) Update(t *model.TableType) error {
	return r.db.Save(t).Error
}

func (r *tableTypeRepository) Delete(id uint) error {
	return r.db.Model(&model.TableType{}).Where("id = ?", id).
		Update("deleted_at", gorm.Expr("NOW()")).Error
}
