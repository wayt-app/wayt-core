package service

import (
	"encoding/json"
	"fmt"

	"github.com/wayt/wayt-core/internal/model"
	"github.com/wayt/wayt-core/internal/repository"
	"github.com/wayt/wayt-core/pkg/sse"
)

type NotificationService interface {
	Send(userType string, userID uint, title, message string) error
	List(userType string, userID uint) ([]model.Notification, error)
	CountUnread(userType string, userID uint) (int64, error)
	MarkAllRead(userType string, userID uint) error
}

type notificationService struct {
	repo repository.NotificationRepository
	hub  *sse.Hub
}

func NewNotificationService(repo repository.NotificationRepository, hub *sse.Hub) NotificationService {
	return &notificationService{repo: repo, hub: hub}
}

func (s *notificationService) Send(userType string, userID uint, title, message string) error {
	n := &model.Notification{
		UserType: userType,
		UserID:   userID,
		Title:    title,
		Message:  message,
	}
	if err := s.repo.Create(n); err != nil {
		return err
	}

	// Push via SSE
	data, _ := json.Marshal(map[string]any{"type": "notification", "data": n})
	var key string
	switch userType {
	case "customer":
		key = sse.CustomerKey(userID)
	case "owner":
		key = sse.OwnerKey(userID)
	case "staff":
		key = sse.StaffKey(userID)
	default:
		key = fmt.Sprintf("%s:%d", userType, userID)
	}
	s.hub.Publish(key, string(data))
	return nil
}

func (s *notificationService) List(userType string, userID uint) ([]model.Notification, error) {
	return s.repo.FindByUser(userType, userID, 30)
}

func (s *notificationService) CountUnread(userType string, userID uint) (int64, error) {
	return s.repo.CountUnread(userType, userID)
}

func (s *notificationService) MarkAllRead(userType string, userID uint) error {
	return s.repo.MarkAllRead(userType, userID)
}
