package service

import (
	"errors"

	"github.com/wayt-app/wayt-core/model"
	"github.com/wayt-app/wayt-core/repository"
)

type MenuItemService interface {
	Create(branchID uint, name, description string, price int64, category model.MenuCategory) (*model.MenuItem, error)
	ListByBranch(branchID uint) ([]model.MenuItem, error)
	FindByID(id uint) (*model.MenuItem, error)
	Update(id uint, name, description string, price int64, category model.MenuCategory, isAvailable, isFavorite, isChefRecommendation bool) (*model.MenuItem, error)
	UpdateImage(id uint, imageURL string) error
	Delete(id uint) error
}

type menuItemService struct {
	repo       repository.MenuItemRepository
	branchRepo repository.BranchRepository
}

func NewMenuItemService(repo repository.MenuItemRepository, branchRepo repository.BranchRepository) MenuItemService {
	return &menuItemService{repo: repo, branchRepo: branchRepo}
}

func (s *menuItemService) Create(branchID uint, name, description string, price int64, category model.MenuCategory) (*model.MenuItem, error) {
	if name == "" {
		return nil, errors.New("nama menu wajib diisi")
	}
	if price < 0 {
		return nil, errors.New("harga tidak boleh negatif")
	}
	if !validCategory(category) {
		return nil, errors.New("kategori tidak valid")
	}
	if _, err := s.branchRepo.FindByID(branchID); err != nil {
		return nil, errors.New("cabang tidak ditemukan")
	}
	m := &model.MenuItem{
		BranchID:    branchID,
		Name:        name,
		Description: description,
		Price:       price,
		Category:    category,
		IsAvailable: true,
	}
	if err := s.repo.Create(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (s *menuItemService) ListByBranch(branchID uint) ([]model.MenuItem, error) {
	return s.repo.FindByBranch(branchID)
}

func (s *menuItemService) FindByID(id uint) (*model.MenuItem, error) {
	m, err := s.repo.FindByID(id)
	if err != nil {
		return nil, errors.New("menu tidak ditemukan")
	}
	return m, nil
}

func (s *menuItemService) Update(id uint, name, description string, price int64, category model.MenuCategory, isAvailable, isFavorite, isChefRecommendation bool) (*model.MenuItem, error) {
	m, err := s.repo.FindByID(id)
	if err != nil {
		return nil, errors.New("menu tidak ditemukan")
	}
	if name != "" {
		m.Name = name
	}
	if price >= 0 {
		m.Price = price
	}
	if validCategory(category) {
		m.Category = category
	}
	m.Description = description
	m.IsAvailable = isAvailable
	m.IsFavorite = isFavorite
	m.IsChefRecommendation = isChefRecommendation
	if err := s.repo.Update(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (s *menuItemService) UpdateImage(id uint, imageURL string) error {
	m, err := s.repo.FindByID(id)
	if err != nil {
		return errors.New("menu tidak ditemukan")
	}
	m.ImageURL = imageURL
	return s.repo.Update(m)
}

func (s *menuItemService) Delete(id uint) error {
	if _, err := s.repo.FindByID(id); err != nil {
		return errors.New("menu tidak ditemukan")
	}
	return s.repo.Delete(id)
}

func validCategory(c model.MenuCategory) bool {
	switch c {
	case model.MenuCategoryMain, model.MenuCategorySide, model.MenuCategoryDessert,
		model.MenuCategoryDrink, model.MenuCategorySnack, model.MenuCategoryOther:
		return true
	}
	return false
}
