package testmock

import "github.com/wayt-app/wayt-core/model"

// ---- NotificationService ----

type NotificationSvc struct {
	SendFn        func(userType string, userID uint, title, message string) error
	ListFn        func(userType string, userID uint) ([]model.Notification, error)
	ListUnreadFn  func(userType string, userID uint) ([]model.Notification, error)
	CountUnreadFn func(userType string, userID uint) (int64, error)
	MarkAllReadFn func(userType string, userID uint) error
}

func (m *NotificationSvc) Send(userType string, userID uint, title, message string) error {
	if m.SendFn != nil {
		return m.SendFn(userType, userID, title, message)
	}
	return nil
}
func (m *NotificationSvc) List(userType string, userID uint) ([]model.Notification, error) {
	if m.ListFn != nil {
		return m.ListFn(userType, userID)
	}
	return nil, nil
}
func (m *NotificationSvc) ListUnread(userType string, userID uint) ([]model.Notification, error) {
	if m.ListUnreadFn != nil {
		return m.ListUnreadFn(userType, userID)
	}
	return nil, nil
}
func (m *NotificationSvc) CountUnread(userType string, userID uint) (int64, error) {
	if m.CountUnreadFn != nil {
		return m.CountUnreadFn(userType, userID)
	}
	return 0, nil
}
func (m *NotificationSvc) MarkAllRead(userType string, userID uint) error {
	if m.MarkAllReadFn != nil {
		return m.MarkAllReadFn(userType, userID)
	}
	return nil
}

// ---- ReservationIncrementer ----

type ReservationIncr struct {
	IncrementFn func(ownerID uint) error
}

func (m *ReservationIncr) IncrementReservation(ownerID uint) error {
	if m.IncrementFn != nil {
		return m.IncrementFn(ownerID)
	}
	return nil
}

// ---- email.Sender ----

type EmailSender struct {
	SendFn func(to, subject, html string) error
}

func (m *EmailSender) Send(to, subject, html string) error {
	if m.SendFn != nil {
		return m.SendFn(to, subject, html)
	}
	return nil
}

// ---- whatsapp.Sender ----

type WASender struct {
	SendFn func(phone, message string) error
}

func (m *WASender) Send(phone, message string) error {
	if m.SendFn != nil {
		return m.SendFn(phone, message)
	}
	return nil
}
