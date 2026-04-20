package model

import "time"

type Staff struct {
	ID              uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	BusinessOwnerID uint      `gorm:"not null;index"           json:"business_owner_id"`
	BranchID        uint      `gorm:"not null;index"           json:"branch_id"`
	Name                string     `gorm:"size:100;not null"             json:"name"`
	Email               string     `gorm:"size:150;not null;uniqueIndex" json:"email"`
	Password            string     `gorm:"size:255;not null"             json:"-"`
	IsActive            bool       `gorm:"not null;default:true"         json:"is_active"`
	TokenVersion        int        `gorm:"not null;default:0"            json:"-"`
	ResetToken          *string    `gorm:"size:64;index"                 json:"-"`
	ResetTokenExpiresAt *time.Time `                                     json:"-"`
	CreatedAt           time.Time  `                                     json:"created_at"`
	UpdatedAt           time.Time  `                                     json:"updated_at"`

	Branch *Branch `gorm:"foreignKey:BranchID" json:"branch,omitempty"`
}

func (Staff) TableName() string { return "tabl_staff" }
