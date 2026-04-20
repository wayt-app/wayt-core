package model

import "time"

type Restaurant struct {
	ID              uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	Name            string     `gorm:"size:150;not null"        json:"name"`
	Description     string     `gorm:"type:text"                json:"description"`
	Address         string     `gorm:"type:text"                json:"address"`
	Phone           string     `gorm:"size:20"                  json:"phone"`
	CuisineType     string     `gorm:"size:50"                  json:"cuisine_type"`
	LogoURL         string     `gorm:"type:text"                json:"logo_url"`
	PromoToken      string     `gorm:"size:32;uniqueIndex"      json:"promo_token,omitempty"`
	IsActive        bool       `gorm:"default:true"             json:"is_active"`
	BusinessOwnerID *uint      `gorm:"index"                    json:"business_owner_id,omitempty"`
	CreatedAt       time.Time  `                                json:"created_at"`
	UpdatedAt       time.Time  `                                json:"updated_at"`
	DeletedAt       *time.Time `gorm:"index"                    json:"deleted_at,omitempty"`
}

func (Restaurant) TableName() string { return "tabl_restaurants" }
