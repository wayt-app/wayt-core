package model

import "time"

type Room struct {
	ID         uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	BranchID   uint      `gorm:"not null;index"           json:"branch_id"`
	Name       string    `gorm:"size:100;not null"        json:"name"`
	IsSmoking  bool      `gorm:"not null;default:false"   json:"is_smoking"`
	IsDefault  bool      `gorm:"not null;default:false"   json:"is_default"`
	IsActive   bool      `gorm:"not null;default:true"    json:"is_active"`
	CreatedAt  time.Time `                                json:"created_at"`
	UpdatedAt  time.Time `                                json:"updated_at"`

	Branch *Branch `gorm:"foreignKey:BranchID" json:"branch,omitempty"`
}

func (Room) TableName() string { return "tabl_rooms" }
