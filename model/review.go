package model

import "time"

type Review struct {
	ID           uint      `gorm:"primaryKey;autoIncrement"     json:"id"`
	CustomerID   uint      `gorm:"not null;index"               json:"customer_id"`
	RestaurantID uint      `gorm:"not null;index"               json:"restaurant_id"`
	BranchID     uint      `gorm:"not null;index"               json:"branch_id"`
	BookingID    uint      `gorm:"not null;uniqueIndex"         json:"booking_id"`
	Rating       int       `gorm:"not null"                     json:"rating"` // 1–5
	Comment      *string   `gorm:"type:text"                    json:"comment,omitempty"`
	Customer     *Customer `gorm:"foreignKey:CustomerID"        json:"customer,omitempty"`
	CreatedAt    time.Time `                                     json:"created_at"`
	UpdatedAt    time.Time `                                     json:"updated_at"`
}

func (Review) TableName() string { return "tabl_reviews" }
