package repository

import (
	"time"

	"github.com/wayt/wayt-core/model"
	"gorm.io/gorm"
)

type SubscriptionRepository interface {
	Create(s *model.Subscription) error
	FindByOwnerID(ownerID uint) (*model.Subscription, error)
	FindByID(id uint) (*model.Subscription, error)
	UpdateStatus(id uint, status model.SubscriptionStatus, notes string) error
	IncrementReservations(id uint) error
	ResetMonthlyCount(id uint) error
	FindTrialExpiring(within time.Duration) ([]model.Subscription, error)
	FindTrialExpired() ([]model.Subscription, error)
	FindNeedingReset() ([]model.Subscription, error)
	FindAll() ([]model.Subscription, error)
	Update(s *model.Subscription) error
	FindByRestaurantID(restaurantID uint) (*model.Subscription, error)
}

type subscriptionRepository struct{ db *gorm.DB }

func NewSubscriptionRepository(db *gorm.DB) SubscriptionRepository {
	return &subscriptionRepository{db: db}
}

func (r *subscriptionRepository) Create(s *model.Subscription) error {
	return r.db.Create(s).Error
}

func (r *subscriptionRepository) FindByOwnerID(ownerID uint) (*model.Subscription, error) {
	var s model.Subscription
	err := r.db.Preload("Plan").
		Where("business_owner_id = ?", ownerID).
		Order("id DESC").
		First(&s).Error
	return &s, err
}

func (r *subscriptionRepository) FindByID(id uint) (*model.Subscription, error) {
	var s model.Subscription
	err := r.db.Preload("Plan").Preload("BusinessOwner").
		Where("id = ?", id).First(&s).Error
	return &s, err
}

func (r *subscriptionRepository) UpdateStatus(id uint, status model.SubscriptionStatus, notes string) error {
	updates := map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}
	if notes != "" {
		updates["notes"] = notes
	}
	if status == model.SubscriptionStatusActive {
		updates["activated_at"] = time.Now()
	}
	return r.db.Model(&model.Subscription{}).Where("id = ?", id).Updates(updates).Error
}

func (r *subscriptionRepository) IncrementReservations(id uint) error {
	return r.db.Exec(
		"UPDATE tabl_subscriptions SET reservations_this_month = reservations_this_month + 1, updated_at = NOW() WHERE id = ?",
		id,
	).Error
}

func (r *subscriptionRepository) ResetMonthlyCount(id uint) error {
	return r.db.Exec(
		"UPDATE tabl_subscriptions SET reservations_this_month = 0, last_reset_at = CURRENT_DATE, warning_sent = FALSE, updated_at = NOW() WHERE id = ?",
		id,
	).Error
}

// FindTrialExpiring returns trial subscriptions where trial_ends_at is within the next `within` duration
// and warning has NOT been sent yet.
func (r *subscriptionRepository) FindTrialExpiring(within time.Duration) ([]model.Subscription, error) {
	var list []model.Subscription
	now := time.Now()
	future := now.Add(within)
	err := r.db.Preload("Plan").Preload("BusinessOwner").
		Where("status = 'trial' AND warning_sent = FALSE AND trial_ends_at >= ? AND trial_ends_at <= ?", now, future).
		Find(&list).Error
	return list, err
}

// FindTrialExpired returns trial subscriptions that have passed their trial_ends_at.
func (r *subscriptionRepository) FindTrialExpired() ([]model.Subscription, error) {
	var list []model.Subscription
	err := r.db.Preload("Plan").Preload("BusinessOwner").
		Where("status = 'trial' AND trial_ends_at < ?", time.Now()).
		Find(&list).Error
	return list, err
}

// FindNeedingReset returns subscriptions where last_reset_at is before the start of the current month.
func (r *subscriptionRepository) FindNeedingReset() ([]model.Subscription, error) {
	var list []model.Subscription
	now := time.Now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local)
	err := r.db.Where("last_reset_at < ?", monthStart.Format("2006-01-02")).Find(&list).Error
	return list, err
}

func (r *subscriptionRepository) FindAll() ([]model.Subscription, error) {
	var list []model.Subscription
	err := r.db.Preload("Plan").Preload("BusinessOwner").
		Order("id DESC").Find(&list).Error
	return list, err
}

func (r *subscriptionRepository) Update(s *model.Subscription) error {
	return r.db.Model(s).Omit("Plan", "BusinessOwner").Save(s).Error
}

// FindByRestaurantID finds the active/trial subscription for the restaurant's business owner.
func (r *subscriptionRepository) FindByRestaurantID(restaurantID uint) (*model.Subscription, error) {
	var s model.Subscription
	err := r.db.Preload("Plan").
		Joins("JOIN tabl_business_owners bo ON tabl_subscriptions.business_owner_id = bo.id").
		Joins("JOIN tabl_restaurants rest ON rest.business_owner_id = bo.id").
		Where("rest.id = ? AND rest.deleted_at IS NULL", restaurantID).
		Where("tabl_subscriptions.status IN ('active','trial')").
		Order("tabl_subscriptions.id DESC").
		First(&s).Error
	return &s, err
}
