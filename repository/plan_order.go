package repository

import (
	"time"

	"github.com/wayt-app/wayt-core/model"
	"gorm.io/gorm"
)

type PlanOrderRepository interface {
	Create(o *model.PlanOrder) error
	FindPendingByOwnerID(ownerID uint) (*model.PlanOrder, error)
	FindDue() ([]model.PlanOrder, error)
	UpdateStatus(id uint, status model.PlanOrderStatus) error
}

type planOrderRepository struct{ db *gorm.DB }

func NewPlanOrderRepository(db *gorm.DB) PlanOrderRepository {
	return &planOrderRepository{db: db}
}

func (r *planOrderRepository) Create(o *model.PlanOrder) error {
	return r.db.Create(o).Error
}

func (r *planOrderRepository) FindPendingByOwnerID(ownerID uint) (*model.PlanOrder, error) {
	var o model.PlanOrder
	err := r.db.Preload("Plan").
		Where("business_owner_id = ? AND status = 'pending'", ownerID).
		Order("id DESC").First(&o).Error
	return &o, err
}

func (r *planOrderRepository) FindDue() ([]model.PlanOrder, error) {
	var list []model.PlanOrder
	err := r.db.Where("status = 'pending' AND process_at <= ?", time.Now()).Find(&list).Error
	return list, err
}

func (r *planOrderRepository) UpdateStatus(id uint, status model.PlanOrderStatus) error {
	return r.db.Model(&model.PlanOrder{}).Where("id = ?", id).
		Updates(map[string]interface{}{"status": status, "updated_at": time.Now()}).Error
}
