package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/wayt-app/wayt-core/model"
	"github.com/wayt-app/wayt-core/repository"
	"github.com/wayt-app/wayt-core/pkg/storage"
)

type MediaService interface {
	UploadLogo(restaurantID uint, data []byte, contentType, ext string) (*model.Media, error)
	UploadMenu(restaurantID uint, branchID *uint, data []byte, contentType, ext string) (*model.Media, error)
	ListByRestaurant(restaurantID uint, mediaType string) ([]model.Media, error)
	MenuForBranch(restaurantID, branchID uint) ([]model.Media, error)
	Delete(id, restaurantID uint) error
}

type mediaService struct {
	repo        repository.MediaRepository
	restaurantR repository.RestaurantRepository
	storage     *storage.Client
}

func NewMediaService(repo repository.MediaRepository, restaurantRepo repository.RestaurantRepository, storageClient *storage.Client) MediaService {
	return &mediaService{repo: repo, restaurantR: restaurantRepo, storage: storageClient}
}

func (s *mediaService) UploadLogo(restaurantID uint, data []byte, contentType, ext string) (*model.Media, error) {
	// Delete existing logo (storage + DB)
	existing, _ := s.repo.DeleteLogoByRestaurant(restaurantID)
	for _, old := range existing {
		_ = s.storage.Delete([]string{old.StoragePath})
	}

	path := fmt.Sprintf("logos/%d%s", restaurantID, ext)
	url, err := s.storage.Upload(path, data, contentType)
	if err != nil {
		return nil, errors.New("gagal upload ke storage: " + err.Error())
	}

	m := &model.Media{
		RestaurantID: restaurantID,
		Type:         "logo",
		URL:          url,
		StoragePath:  path,
	}
	if err := s.repo.Create(m); err != nil {
		return nil, err
	}
	_ = s.restaurantR.UpdateLogoURL(restaurantID, url)
	return m, nil
}

func (s *mediaService) UploadMenu(restaurantID uint, branchID *uint, data []byte, contentType, ext string) (*model.Media, error) {
	ts := time.Now().UnixMilli()
	var path string
	if branchID != nil {
		path = fmt.Sprintf("menus/%d/%d/%d%s", restaurantID, *branchID, ts, ext)
	} else {
		path = fmt.Sprintf("menus/%d/all/%d%s", restaurantID, ts, ext)
	}

	url, err := s.storage.Upload(path, data, contentType)
	if err != nil {
		return nil, errors.New("gagal upload ke storage: " + err.Error())
	}

	m := &model.Media{
		RestaurantID: restaurantID,
		BranchID:     branchID,
		Type:         "menu",
		URL:          url,
		StoragePath:  path,
	}
	if err := s.repo.Create(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (s *mediaService) ListByRestaurant(restaurantID uint, mediaType string) ([]model.Media, error) {
	return s.repo.FindByRestaurant(restaurantID, mediaType)
}

func (s *mediaService) MenuForBranch(restaurantID, branchID uint) ([]model.Media, error) {
	return s.repo.FindMenuForBranch(restaurantID, branchID)
}

func (s *mediaService) Delete(id, restaurantID uint) error {
	m, err := s.repo.FindByID(id)
	if err != nil {
		return errors.New("media tidak ditemukan")
	}
	if m.RestaurantID != restaurantID {
		return errors.New("tidak diizinkan menghapus media ini")
	}
	_ = s.storage.Delete([]string{m.StoragePath})
	if err := s.repo.DeleteByID(id); err != nil {
		return err
	}
	if m.Type == "logo" {
		_ = s.restaurantR.UpdateLogoURL(restaurantID, "")
	}
	return nil
}
