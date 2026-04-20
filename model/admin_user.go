package model

import "time"

type AdminRole string

const (
	RoleSuperAdmin AdminRole = "superadmin"
	RoleAdmin      AdminRole = "admin"
)

type AdminUser struct {
	ID                  uint       `gorm:"primaryKey;autoIncrement"      json:"id"`
	Username            string     `gorm:"size:100;not null;uniqueIndex" json:"username"`
	Password            string     `gorm:"size:255;not null"             json:"-"`
	Role                AdminRole  `gorm:"type:tabl_admin_role;default:'admin'" json:"role"`
	RestaurantID        *uint      `gorm:"index"                         json:"restaurant_id,omitempty"`
	TokenVersion        int        `gorm:"not null;default:0"            json:"-"`
	ResetToken          *string    `gorm:"size:64;index"                 json:"-"`
	ResetTokenExpiresAt *time.Time `                                     json:"-"`
	CreatedAt           time.Time  `                                     json:"created_at"`
	UpdatedAt           time.Time  `                                     json:"updated_at"`

	Restaurant *Restaurant `gorm:"foreignKey:RestaurantID" json:"restaurant,omitempty"`
}

func (AdminUser) TableName() string { return "tabl_admin_users" }
