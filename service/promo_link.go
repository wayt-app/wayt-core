package service

import (
	"crypto/rand"
	"errors"
	"strings"

	"github.com/wayt-app/wayt-core/model"
	"github.com/wayt-app/wayt-core/repository"
)

type PromoLinkService interface {
	Create(restaurantID uint, label string) (*model.PromoLink, error)
	List(restaurantID uint) ([]model.PromoLink, error)
	FindByToken(token string) (*model.PromoLink, error)
	UpdateLabel(id, restaurantID uint, label string) error
	RegenerateToken(id, restaurantID uint) (*model.PromoLink, error)
	IncrementVisit(token string) error
	Delete(id, restaurantID uint) error
}

type promoLinkService struct{ repo repository.PromoLinkRepository }

func NewPromoLinkService(repo repository.PromoLinkRepository) PromoLinkService {
	return &promoLinkService{repo: repo}
}

func (s *promoLinkService) Create(restaurantID uint, label string) (*model.PromoLink, error) {
	token, err := generatePromoLinkToken()
	if err != nil {
		return nil, errors.New("gagal membuat token")
	}
	p := &model.PromoLink{
		RestaurantID: restaurantID,
		Label:        strings.TrimSpace(label),
		Token:        token,
	}
	return p, s.repo.Create(p)
}

func (s *promoLinkService) List(restaurantID uint) ([]model.PromoLink, error) {
	return s.repo.ListByRestaurant(restaurantID)
}

func (s *promoLinkService) FindByToken(token string) (*model.PromoLink, error) {
	return s.repo.FindByToken(token)
}

func (s *promoLinkService) UpdateLabel(id, restaurantID uint, label string) error {
	return s.repo.UpdateLabel(id, strings.TrimSpace(label))
}

func (s *promoLinkService) RegenerateToken(id, restaurantID uint) (*model.PromoLink, error) {
	token, err := generatePromoLinkToken()
	if err != nil {
		return nil, errors.New("gagal membuat token baru")
	}
	if err := s.repo.RegenerateToken(id, token); err != nil {
		return nil, err
	}
	return s.repo.FindByToken(token)
}

func (s *promoLinkService) IncrementVisit(token string) error {
	p, err := s.repo.FindByToken(token)
	if err != nil {
		return nil
	}
	return s.repo.IncrementVisit(p.ID)
}

func (s *promoLinkService) Delete(id, restaurantID uint) error {
	return s.repo.Delete(id)
}

func generatePromoLinkToken() (string, error) {
	const chars = "abcdefghjkmnpqrstuvwxyz23456789"
	b := make([]byte, 12)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	for i := range b {
		b[i] = chars[int(b[i])%len(chars)]
	}
	return string(b), nil
}
