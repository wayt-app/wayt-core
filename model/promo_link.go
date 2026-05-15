package model

import "time"

type PromoLink struct {
	ID           uint       `gorm:"primaryKey;autoIncrement"        json:"id"`
	RestaurantID uint       `gorm:"not null;index"                  json:"restaurant_id"`
	Label        string     `gorm:"size:100;not null;default:''"    json:"label"`
	Token        string     `gorm:"size:32;not null;uniqueIndex"    json:"token"`
	VisitCount   int64      `gorm:"not null;default:0"              json:"visit_count"`
	CreatedAt    time.Time  `                                        json:"created_at"`
	UpdatedAt    time.Time  `                                        json:"updated_at"`
	DeletedAt    *time.Time `gorm:"index"                           json:"-"`
}

func (PromoLink) TableName() string { return "tabl_promo_links" }
