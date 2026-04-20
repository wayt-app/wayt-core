package model

import "time"

type BusinessOwner struct {
	ID                  uint       `gorm:"primaryKey;autoIncrement"      json:"id"`
	Name                string     `gorm:"size:100;not null"             json:"name"`
	Email               string     `gorm:"size:150;not null;uniqueIndex" json:"email"`
	Phone               string     `gorm:"size:20"                       json:"phone"`
	Password            string     `gorm:"size:255;not null"             json:"-"`
	IsVerified          bool       `gorm:"not null;default:false"        json:"is_verified"`
	TokenVersion        int        `gorm:"not null;default:0"            json:"-"`
	VerificationToken   *string    `gorm:"size:64;index"                 json:"-"`
	ResetToken          *string    `gorm:"size:64;index"                 json:"-"`
	ResetTokenExpiresAt *time.Time `                                     json:"-"`
	GoogleID            *string    `gorm:"size:255;uniqueIndex"          json:"google_id,omitempty"`
	AvatarURL           *string    `gorm:"type:text"                    json:"avatar_url,omitempty"`
	CreatedAt           time.Time  `                                     json:"created_at"`
	UpdatedAt           time.Time  `                                     json:"updated_at"`
}

func (BusinessOwner) TableName() string { return "tabl_business_owners" }
