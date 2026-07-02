package model

import "time"

type Notification struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserType  string    `gorm:"size:20;not null;index"   json:"user_type"` // customer, owner, staff
	UserID    uint      `gorm:"not null;index"           json:"user_id"`
	Title     string    `gorm:"size:200;not null"        json:"title"`
	Message   string    `gorm:"type:text;not null"       json:"message"`
	IsRead    bool      `gorm:"default:false"            json:"is_read"`
	BookingID *uint     `gorm:"index"                    json:"booking_id,omitempty"` // referensi booking terkait (nullable)
	CreatedAt time.Time `                                json:"created_at"`
}

func (Notification) TableName() string { return "tabl_notifications" }
