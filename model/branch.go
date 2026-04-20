package model

import "time"

type Branch struct {
	ID                     uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	RestaurantID           uint       `gorm:"not null;index"           json:"restaurant_id"`
	Name                   string     `gorm:"size:150;not null"        json:"name"`
	Address                string     `gorm:"type:text"                json:"address"`
	Phone                  string     `gorm:"size:20"                  json:"phone"`
	OpeningHours           string     `gorm:"type:text"                json:"opening_hours"` // human-readable display text
	OpenFrom               string     `gorm:"size:5"                   json:"open_from"`     // "HH:MM" for slot engine
	OpenTo                 string     `gorm:"size:5"                   json:"open_to"`       // "HH:MM" for slot engine
	SlotIntervalMinutes    int        `gorm:"default:30"               json:"slot_interval_minutes"`
	DefaultDurationMinutes int        `gorm:"default:120"              json:"default_duration_minutes"`
	RequireConfirmation    bool       `gorm:"default:true"             json:"require_confirmation"`
	IsActive               bool       `gorm:"default:true"             json:"is_active"`
	Latitude               float64    `gorm:"default:0"                json:"latitude"`
	Longitude              float64    `gorm:"default:0"                json:"longitude"`
	CreatedAt              time.Time  `                                json:"created_at"`
	UpdatedAt              time.Time  `                                json:"updated_at"`
	DeletedAt              *time.Time `gorm:"index"                    json:"deleted_at,omitempty"`

	Restaurant *Restaurant `gorm:"foreignKey:RestaurantID" json:"restaurant,omitempty"`
}

func (Branch) TableName() string { return "tabl_branches" }
