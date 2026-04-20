package service

import (
	"errors"

	"github.com/wayt/wayt-core/model"
	"github.com/wayt/wayt-core/repository"
)

type PlanService interface {
	Create(name string, maxBranches, maxReservationsPerMonth int, waNotifEnabled bool, warningThresholdPct int, price float64) (*model.Plan, error)
	FindByID(id uint) (*model.Plan, error)
	List() ([]model.Plan, error)
	Update(id uint, name string, maxBranches, maxReservationsPerMonth int, waNotifEnabled bool, warningThresholdPct int, price float64, isActive bool) (*model.Plan, error)
	Delete(id uint) error
}

type planService struct {
	repo repository.PlanRepository
}

func NewPlanService(repo repository.PlanRepository) PlanService {
	return &planService{repo: repo}
}

func (s *planService) Create(name string, maxBranches, maxReservationsPerMonth int, waNotifEnabled bool, warningThresholdPct int, price float64) (*model.Plan, error) {
	if name == "" {
		return nil, errors.New("nama paket wajib diisi")
	}
	p := &model.Plan{
		Name:                    name,
		MaxBranches:             maxBranches,
		MaxReservationsPerMonth: maxReservationsPerMonth,
		WaNotifEnabled:          waNotifEnabled,
		WarningThresholdPct:     warningThresholdPct,
		Price:                   price,
		IsActive:                true,
	}
	if err := s.repo.Create(p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *planService) FindByID(id uint) (*model.Plan, error) {
	p, err := s.repo.FindByID(id)
	if err != nil {
		return nil, errors.New("paket tidak ditemukan")
	}
	return p, nil
}

func (s *planService) List() ([]model.Plan, error) {
	return s.repo.List()
}

func (s *planService) Update(id uint, name string, maxBranches, maxReservationsPerMonth int, waNotifEnabled bool, warningThresholdPct int, price float64, isActive bool) (*model.Plan, error) {
	p, err := s.repo.FindByID(id)
	if err != nil {
		return nil, errors.New("paket tidak ditemukan")
	}
	if name != "" {
		p.Name = name
	}
	p.MaxBranches = maxBranches
	p.MaxReservationsPerMonth = maxReservationsPerMonth
	p.WaNotifEnabled = waNotifEnabled
	p.WarningThresholdPct = warningThresholdPct
	p.Price = price
	p.IsActive = isActive
	if err := s.repo.Update(p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *planService) Delete(id uint) error {
	if _, err := s.repo.FindByID(id); err != nil {
		return errors.New("paket tidak ditemukan")
	}
	return s.repo.Delete(id)
}
