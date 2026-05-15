package repository

import (
	"time"

	"github.com/wayt-app/wayt-core/model"
	"gorm.io/gorm"
)

type PromoLinkRepository interface {
	Create(p *model.PromoLink) error
	ListByRestaurant(restaurantID uint) ([]model.PromoLink, error)
	FindByToken(token string) (*model.PromoLink, error)
	UpdateLabel(id uint, label string) error
	RegenerateToken(id uint, token string) error
	IncrementVisit(id uint) error
	Delete(id uint) error
}

type promoLinkRepository struct{ db *gorm.DB }

func NewPromoLinkRepository(db *gorm.DB) PromoLinkRepository {
	return &promoLinkRepository{db: db}
}

func (r *promoLinkRepository) Create(p *model.PromoLink) error {
	return r.db.Create(p).Error
}

func (r *promoLinkRepository) ListByRestaurant(restaurantID uint) ([]model.PromoLink, error) {
	var list []model.PromoLink
	err := r.db.Where("restaurant_id = ? AND deleted_at IS NULL", restaurantID).
		Order("created_at ASC").Find(&list).Error
	return list, err
}

func (r *promoLinkRepository) FindByToken(token string) (*model.PromoLink, error) {
	var p model.PromoLink
	err := r.db.Where("token = ? AND deleted_at IS NULL", token).First(&p).Error
	return &p, err
}

func (r *promoLinkRepository) UpdateLabel(id uint, label string) error {
	return r.db.Model(&model.PromoLink{}).Where("id = ?", id).
		Updates(map[string]interface{}{"label": label, "updated_at": time.Now()}).Error
}

func (r *promoLinkRepository) RegenerateToken(id uint, token string) error {
	return r.db.Model(&model.PromoLink{}).Where("id = ?", id).
		Updates(map[string]interface{}{"token": token, "updated_at": time.Now()}).Error
}

func (r *promoLinkRepository) IncrementVisit(id uint) error {
	return r.db.Exec("UPDATE tabl_promo_links SET visit_count = visit_count + 1, updated_at = NOW() WHERE id = ?", id).Error
}

func (r *promoLinkRepository) Delete(id uint) error {
	return r.db.Exec("UPDATE tabl_promo_links SET deleted_at = NOW() WHERE id = ?", id).Error
}
