package service

import (
	"errors"

	"github.com/wayt/wayt-core/internal/model"
	"github.com/wayt/wayt-core/internal/repository"
)

type BranchService interface {
	Create(restaurantID uint, name, address, phone, openingHours, openFrom, openTo string, slotInterval, durationMinutes int, requireConfirmation bool, latitude, longitude float64) (*model.Branch, error)
	ListByRestaurant(restaurantID uint) ([]model.Branch, error)
	ListActiveByRestaurant(restaurantID uint) ([]model.Branch, error)
	FindByID(id uint) (*model.Branch, error)
	Update(id uint, name, address, phone, openingHours, openFrom, openTo string, slotInterval, durationMinutes int, requireConfirmation, isActive bool, latitude, longitude float64) (*model.Branch, error)
	Delete(id uint) error
}

type branchService struct {
	repo           repository.BranchRepository
	restaurantRepo repository.RestaurantRepository
}

func NewBranchService(repo repository.BranchRepository, restaurantRepo repository.RestaurantRepository) BranchService {
	return &branchService{repo: repo, restaurantRepo: restaurantRepo}
}

func (s *branchService) Create(restaurantID uint, name, address, phone, openingHours, openFrom, openTo string, slotInterval, durationMinutes int, requireConfirmation bool, latitude, longitude float64) (*model.Branch, error) {
	if name == "" {
		return nil, errors.New("nama cabang wajib diisi")
	}
	if _, err := s.restaurantRepo.FindByID(restaurantID); err != nil {
		return nil, errors.New("restoran tidak ditemukan")
	}
	if durationMinutes <= 0 {
		durationMinutes = 120
	}
	if slotInterval <= 0 {
		slotInterval = 30
	}
	b := &model.Branch{
		RestaurantID:           restaurantID,
		Name:                   name,
		Address:                address,
		Phone:                  phone,
		OpeningHours:           openingHours,
		OpenFrom:               openFrom,
		OpenTo:                 openTo,
		SlotIntervalMinutes:    slotInterval,
		DefaultDurationMinutes: durationMinutes,
		RequireConfirmation:    requireConfirmation,
		IsActive:               true,
		Latitude:               latitude,
		Longitude:              longitude,
	}
	if err := s.repo.Create(b); err != nil {
		return nil, err
	}
	return b, nil
}

func (s *branchService) ListByRestaurant(restaurantID uint) ([]model.Branch, error) {
	return s.repo.FindByRestaurant(restaurantID)
}

func (s *branchService) ListActiveByRestaurant(restaurantID uint) ([]model.Branch, error) {
	return s.repo.FindActiveByRestaurant(restaurantID)
}

func (s *branchService) FindByID(id uint) (*model.Branch, error) {
	b, err := s.repo.FindByID(id)
	if err != nil {
		return nil, errors.New("cabang tidak ditemukan")
	}
	return b, nil
}

func (s *branchService) Update(id uint, name, address, phone, openingHours, openFrom, openTo string, slotInterval, durationMinutes int, requireConfirmation, isActive bool, latitude, longitude float64) (*model.Branch, error) {
	b, err := s.repo.FindByID(id)
	if err != nil {
		return nil, errors.New("cabang tidak ditemukan")
	}
	if name != "" {
		b.Name = name
	}
	b.Address = address
	b.Phone = phone
	b.OpeningHours = openingHours
	b.OpenFrom = openFrom
	b.OpenTo = openTo
	if slotInterval > 0 {
		b.SlotIntervalMinutes = slotInterval
	}
	if durationMinutes > 0 {
		b.DefaultDurationMinutes = durationMinutes
	}
	b.RequireConfirmation = requireConfirmation
	b.IsActive = isActive
	b.Latitude = latitude
	b.Longitude = longitude
	if err := s.repo.Update(b); err != nil {
		return nil, err
	}
	return b, nil
}

func (s *branchService) Delete(id uint) error {
	if _, err := s.repo.FindByID(id); err != nil {
		return errors.New("cabang tidak ditemukan")
	}
	return s.repo.Delete(id)
}
