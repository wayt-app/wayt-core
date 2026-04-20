package model

import "time"

type Media struct {
	ID           uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	RestaurantID uint      `gorm:"not null;index"           json:"restaurant_id"`
	BranchID     *uint     `gorm:"index"                    json:"branch_id"`
	Type         string    `gorm:"size:20;not null"         json:"type"` // logo, menu
	URL          string    `gorm:"type:text;not null"       json:"url"`
	StoragePath  string    `gorm:"type:text;not null"       json:"-"`
	DisplayOrder int       `gorm:"default:0"                json:"display_order"`
	CreatedAt    time.Time `                                json:"created_at"`
}

func (Media) TableName() string { return "tabl_media" }
