package repository

import (
	"github.com/wayt/wayt-core/model"
	"gorm.io/gorm"
)

type PlanRepository interface {
	Create(p *model.Plan) error
	FindByID(id uint) (*model.Plan, error)
	FindFirst() (*model.Plan, error)
	List() ([]model.Plan, error)
	Update(p *model.Plan) error
	Delete(id uint) error
}

type planRepository struct{ db *gorm.DB }

func NewPlanRepository(db *gorm.DB) PlanRepository {
	return &planRepository{db: db}
}

func (r *planRepository) Create(p *model.Plan) error {
	return r.db.Create(p).Error
}

func (r *planRepository) FindByID(id uint) (*model.Plan, error) {
	var p model.Plan
	return &p, r.db.Where("id = ?", id).First(&p).Error
}

func (r *planRepository) FindFirst() (*model.Plan, error) {
	var p model.Plan
	return &p, r.db.Where("is_active = true").Order("id ASC").First(&p).Error
}

func (r *planRepository) List() ([]model.Plan, error) {
	var list []model.Plan
	return list, r.db.Order("id ASC").Find(&list).Error
}

func (r *planRepository) Update(p *model.Plan) error {
	return r.db.Save(p).Error
}

func (r *planRepository) Delete(id uint) error {
	return r.db.Model(&model.Plan{}).Where("id = ?", id).Update("is_active", false).Error
}
