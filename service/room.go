package service

import (
	"errors"

	"github.com/wayt-app/wayt-core/model"
	"github.com/wayt-app/wayt-core/repository"
)

type RoomService interface {
	FindByBranch(branchID uint) ([]model.Room, error)
	FindByID(id uint) (*model.Room, error)
	Create(branchID uint, name string, isSmoking, isDefault bool) (*model.Room, error)
	Update(id, branchID uint, name string, isSmoking, isDefault *bool, isActive *bool) (*model.Room, error)
	Delete(id, branchID uint) error
}

type roomService struct {
	repo       repository.RoomRepository
	branchRepo repository.BranchRepository
}

func NewRoomService(repo repository.RoomRepository, branchRepo repository.BranchRepository) RoomService {
	return &roomService{repo: repo, branchRepo: branchRepo}
}

func (s *roomService) FindByBranch(branchID uint) ([]model.Room, error) {
	return s.repo.FindByBranch(branchID)
}

func (s *roomService) FindByID(id uint) (*model.Room, error) {
	r, err := s.repo.FindByID(id)
	if err != nil {
		return nil, errors.New("ruangan tidak ditemukan")
	}
	return r, nil
}

func (s *roomService) Create(branchID uint, name string, isSmoking, isDefault bool) (*model.Room, error) {
	if name == "" {
		return nil, errors.New("nama ruangan wajib diisi")
	}
	if _, err := s.branchRepo.FindByID(branchID); err != nil {
		return nil, errors.New("cabang tidak ditemukan")
	}
	room := &model.Room{
		BranchID:  branchID,
		Name:      name,
		IsSmoking: isSmoking,
		IsDefault: isDefault,
		IsActive:  true,
	}
	if err := s.repo.Create(room); err != nil {
		return nil, err
	}
	return room, nil
}

func (s *roomService) Update(id, branchID uint, name string, isSmoking, isDefault *bool, isActive *bool) (*model.Room, error) {
	room, err := s.repo.FindByID(id)
	if err != nil {
		return nil, errors.New("ruangan tidak ditemukan")
	}
	if room.BranchID != branchID {
		return nil, errors.New("ruangan tidak berada di cabang ini")
	}
	if name != "" {
		room.Name = name
	}
	if isSmoking != nil {
		room.IsSmoking = *isSmoking
	}
	if isDefault != nil {
		room.IsDefault = *isDefault
	}
	if isActive != nil {
		room.IsActive = *isActive
	}
	if err := s.repo.Update(room); err != nil {
		return nil, err
	}
	return room, nil
}

func (s *roomService) Delete(id, branchID uint) error {
	room, err := s.repo.FindByID(id)
	if err != nil {
		return errors.New("ruangan tidak ditemukan")
	}
	if room.BranchID != branchID {
		return errors.New("ruangan tidak berada di cabang ini")
	}
	count, err := s.repo.CountActiveBookings(id)
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.New("ruangan masih memiliki booking aktif, batalkan terlebih dahulu")
	}
	return s.repo.Delete(id)
}
