package repository

import (
	"github.com/wayt/wayt-core/internal/model"
	"gorm.io/gorm"
)

type NotificationRepository interface {
	Create(n *model.Notification) error
	FindByUser(userType string, userID uint, limit int) ([]model.Notification, error)
	CountUnread(userType string, userID uint) (int64, error)
	MarkAllRead(userType string, userID uint) error
	MarkRead(id uint) error
}

type notificationRepository struct{ db *gorm.DB }

func NewNotificationRepository(db *gorm.DB) NotificationRepository {
	return &notificationRepository{db: db}
}

func (r *notificationRepository) Create(n *model.Notification) error {
	return r.db.Create(n).Error
}

func (r *notificationRepository) FindByUser(userType string, userID uint, limit int) ([]model.Notification, error) {
	var list []model.Notification
	q := r.db.Where("user_type = ? AND user_id = ?", userType, userID).Order("created_at DESC")
	if limit > 0 {
		q = q.Limit(limit)
	}
	err := q.Find(&list).Error
	return list, err
}

func (r *notificationRepository) CountUnread(userType string, userID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.Notification{}).
		Where("user_type = ? AND user_id = ? AND is_read = false", userType, userID).
		Count(&count).Error
	return count, err
}

func (r *notificationRepository) MarkAllRead(userType string, userID uint) error {
	return r.db.Model(&model.Notification{}).
		Where("user_type = ? AND user_id = ? AND is_read = false", userType, userID).
		Update("is_read", true).Error
}

func (r *notificationRepository) MarkRead(id uint) error {
	return r.db.Model(&model.Notification{}).Where("id = ?", id).Update("is_read", true).Error
}
