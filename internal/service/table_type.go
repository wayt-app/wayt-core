package service

import (
	"errors"

	"github.com/wayt/wayt-core/internal/model"
	"github.com/wayt/wayt-core/internal/repository"
)

type TableTypeService interface {
	Create(branchID uint, name string, capacity, totalTables int) (*model.TableType, error)
	ListByBranch(branchID uint) ([]model.TableType, error)
	FindByID(id uint) (*model.TableType, error)
	Update(id uint, name string, capacity, totalTables int, isActive bool) (*model.TableType, error)
	Delete(id uint) error
}

type tableTypeService struct {
	repo       repository.TableTypeRepository
	branchRepo repository.BranchRepository
}

func NewTableTypeService(repo repository.TableTypeRepository, branchRepo repository.BranchRepository) TableTypeService {
	return &tableTypeService{repo: repo, branchRepo: branchRepo}
}

func (s *tableTypeService) Create(branchID uint, name string, capacity, totalTables int) (*model.TableType, error) {
	if name == "" {
		return nil, errors.New("nama tipe meja wajib diisi")
	}
	if capacity <= 0 {
		return nil, errors.New("kapasitas kursi harus lebih dari 0")
	}
	if totalTables <= 0 {
		totalTables = 1
	}
	if _, err := s.branchRepo.FindByID(branchID); err != nil {
		return nil, errors.New("cabang tidak ditemukan")
	}
	t := &model.TableType{
		BranchID:    branchID,
		Name:        name,
		Capacity:    capacity,
		TotalTables: totalTables,
		IsActive:    true,
	}
	if err := s.repo.Create(t); err != nil {
		return nil, err
	}
	return t, nil
}

func (s *tableTypeService) ListByBranch(branchID uint) ([]model.TableType, error) {
	return s.repo.FindByBranch(branchID)
}

func (s *tableTypeService) FindByID(id uint) (*model.TableType, error) {
	t, err := s.repo.FindByID(id)
	if err != nil {
		return nil, errors.New("tipe meja tidak ditemukan")
	}
	return t, nil
}

func (s *tableTypeService) Update(id uint, name string, capacity, totalTables int, isActive bool) (*model.TableType, error) {
	t, err := s.repo.FindByID(id)
	if err != nil {
		return nil, errors.New("tipe meja tidak ditemukan")
	}
	if name != "" {
		t.Name = name
	}
	if capacity > 0 {
		t.Capacity = capacity
	}
	if totalTables > 0 {
		t.TotalTables = totalTables
	}
	t.IsActive = isActive
	if err := s.repo.Update(t); err != nil {
		return nil, err
	}
	return t, nil
}

func (s *tableTypeService) Delete(id uint) error {
	if _, err := s.repo.FindByID(id); err != nil {
		return errors.New("tipe meja tidak ditemukan")
	}
	return s.repo.Delete(id)
}
