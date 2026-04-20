package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"

	"github.com/wayt/wayt-core/model"
	"github.com/wayt/wayt-core/repository"
)

func generatePromoToken() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		// fallback: won't happen in practice
		return hex.EncodeToString([]byte("fallback-token"))
	}
	return hex.EncodeToString(b)
}

type RestaurantService interface {
	Create(name, description, address, phone, cuisineType string) (*model.Restaurant, error)
	CreateForOwner(ownerID uint, name, description, address, phone, cuisineType string) (*model.Restaurant, error)
	List() ([]model.Restaurant, error)
	ListActive() ([]model.Restaurant, error)
	FindByID(id uint) (*model.Restaurant, error)
	FindByOwnerID(ownerID uint) (*model.Restaurant, error)
	FindByPromoToken(token string) (*model.Restaurant, error)
	Update(id uint, name, description, address, phone string, isActive bool, cuisineType string) (*model.Restaurant, error)
	UpdateForOwner(id, ownerID uint, name, description, address, phone string, isActive bool, cuisineType string) (*model.Restaurant, error)
	Delete(id uint) error
}

type restaurantService struct {
	repo repository.RestaurantRepository
}

func NewRestaurantService(repo repository.RestaurantRepository) RestaurantService {
	return &restaurantService{repo: repo}
}

func (s *restaurantService) Create(name, description, address, phone, cuisineType string) (*model.Restaurant, error) {
	if name == "" {
		return nil, errors.New("nama restoran wajib diisi")
	}
	r := &model.Restaurant{Name: name, Description: description, Address: address, Phone: phone, CuisineType: cuisineType, IsActive: true, PromoToken: generatePromoToken()}
	if err := s.repo.Create(r); err != nil {
		return nil, err
	}
	return r, nil
}

func (s *restaurantService) CreateForOwner(ownerID uint, name, description, address, phone, cuisineType string) (*model.Restaurant, error) {
	if name == "" {
		return nil, errors.New("nama restoran wajib diisi")
	}
	// Check if owner already has a restaurant
	if _, err := s.repo.FindByOwnerID(ownerID); err == nil {
		return nil, errors.New("Anda sudah memiliki restoran")
	}
	r := &model.Restaurant{
		Name:            name,
		Description:     description,
		Address:         address,
		Phone:           phone,
		CuisineType:     cuisineType,
		IsActive:        true,
		BusinessOwnerID: &ownerID,
		PromoToken:      generatePromoToken(),
	}
	if err := s.repo.Create(r); err != nil {
		return nil, err
	}
	return r, nil
}

func (s *restaurantService) FindByOwnerID(ownerID uint) (*model.Restaurant, error) {
	r, err := s.repo.FindByOwnerID(ownerID)
	if err != nil {
		return nil, errors.New("restoran tidak ditemukan")
	}
	return r, nil
}

func (s *restaurantService) UpdateForOwner(id, ownerID uint, name, description, address, phone string, isActive bool, cuisineType string) (*model.Restaurant, error) {
	r, err := s.repo.FindByID(id)
	if err != nil {
		return nil, errors.New("restoran tidak ditemukan")
	}
	if r.BusinessOwnerID == nil || *r.BusinessOwnerID != ownerID {
		return nil, errors.New("tidak diizinkan mengubah restoran ini")
	}
	if name != "" {
		r.Name = name
	}
	r.Description = description
	r.Address = address
	r.Phone = phone
	r.CuisineType = cuisineType
	r.IsActive = isActive
	if err := s.repo.Update(r); err != nil {
		return nil, err
	}
	return r, nil
}

func (s *restaurantService) List() ([]model.Restaurant, error) {
	return s.repo.FindAll()
}

func (s *restaurantService) ListActive() ([]model.Restaurant, error) {
	return s.repo.FindAllActive()
}

func (s *restaurantService) FindByID(id uint) (*model.Restaurant, error) {
	r, err := s.repo.FindByID(id)
	if err != nil {
		return nil, errors.New("restoran tidak ditemukan")
	}
	return r, nil
}

func (s *restaurantService) Update(id uint, name, description, address, phone string, isActive bool, cuisineType string) (*model.Restaurant, error) {
	r, err := s.repo.FindByID(id)
	if err != nil {
		return nil, errors.New("restoran tidak ditemukan")
	}
	if name != "" {
		r.Name = name
	}
	r.Description = description
	r.Address = address
	r.Phone = phone
	r.CuisineType = cuisineType
	r.IsActive = isActive
	if err := s.repo.Update(r); err != nil {
		return nil, err
	}
	return r, nil
}

func (s *restaurantService) FindByPromoToken(token string) (*model.Restaurant, error) {
	r, err := s.repo.FindByPromoToken(token)
	if err != nil {
		return nil, errors.New("restoran tidak ditemukan")
	}
	return r, nil
}

func (s *restaurantService) Delete(id uint) error {
	if _, err := s.repo.FindByID(id); err != nil {
		return errors.New("restoran tidak ditemukan")
	}
	return s.repo.Delete(id)
}
