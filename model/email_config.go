package model

import "time"

type EmailConfig struct {
	ID             uint      `gorm:"primaryKey"               json:"id"`
	HeaderImageURL string    `gorm:"type:text;default:''"     json:"header_image_url"`
	LogoURL        string    `gorm:"type:text;default:''"     json:"logo_url"`
	InstagramURL   string    `gorm:"type:text;default:''"     json:"instagram_url"`
	FacebookURL    string    `gorm:"type:text;default:''"     json:"facebook_url"`
	TiktokURL      string    `gorm:"type:text;default:''"     json:"tiktok_url"`
	WebsiteURL     string    `gorm:"type:text;default:'https://wayt.fun'" json:"website_url"`
	SupportEmail   string    `gorm:"size:150;default:'support@wayt.fun'"  json:"support_email"`
	FooterBgURL    string    `gorm:"type:text;default:''"     json:"footer_bg_url"`
	FooterNote     string    `gorm:"type:text;default:'Email ini dikirim secara otomatis, mohon tidak membalas email ini.'" json:"footer_note"`
	Copyright      string    `gorm:"type:text;default:'© 2026 Wayt. All rights reserved.'" json:"copyright"`
	UpdatedAt      time.Time `                                json:"updated_at"`
}

func (EmailConfig) TableName() string { return "tabl_email_config" }
