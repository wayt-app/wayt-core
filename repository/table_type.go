package repository

import (
	"github.com/wayt-app/wayt-core/model"
	"gorm.io/gorm"
)

type TableTypeRepository interface {
	Create(t *model.TableType) error
	FindByBranch(branchID uint) ([]model.TableType, error)
	// FindByBranchWithRoom returns all table types for a branch with Room preloaded.
	FindByBranchWithRoom(branchID uint) ([]model.TableType, error)
	// FindByBranchAndRoom returns table types for a branch filtered by room.
	// roomID == nil → all table types regardless of room assignment.
	// roomID != nil → only table types with room_id = *roomID.
	FindByBranchAndRoom(branchID uint, roomID *uint) ([]model.TableType, error)
	// FindByGroup returns all physical table rows sharing the same name+capacity+room_id
	// within a branch (i.e. all rows in the same "table group").
	FindByGroup(branchID uint, name string, capacity int, roomID *uint) ([]model.TableType, error)
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

func (r *tableTypeRepository) FindByBranchWithRoom(branchID uint) ([]model.TableType, error) {
	var list []model.TableType
	err := r.db.Where("branch_id = ? AND deleted_at IS NULL", branchID).
		Preload("Room").
		Order("id ASC").Find(&list).Error
	return list, err
}

func (r *tableTypeRepository) FindByBranchAndRoom(branchID uint, roomID *uint) ([]model.TableType, error) {
	var list []model.TableType
	q := r.db.Where("branch_id = ? AND deleted_at IS NULL", branchID)
	if roomID != nil {
		q = q.Where("room_id = ?", *roomID)
	}
	err := q.Order("id ASC").Find(&list).Error
	return list, err
}

func (r *tableTypeRepository) FindByGroup(branchID uint, name string, capacity int, roomID *uint) ([]model.TableType, error) {
	var list []model.TableType
	// Group by capacity + room only; name is cosmetic.
	q := r.db.Where("branch_id = ? AND capacity = ? AND deleted_at IS NULL", branchID, capacity)
	if roomID == nil {
		q = q.Where("room_id IS NULL")
	} else {
		q = q.Where("room_id = ?", *roomID)
	}
	err := q.Order("id ASC").Find(&list).Error
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
