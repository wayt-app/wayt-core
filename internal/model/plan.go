package model

import "time"

type Plan struct {
	ID                       uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Name                     string    `gorm:"size:100;not null"        json:"name"`
	MaxBranches              int       `gorm:"not null;default:1"       json:"max_branches"`
	MaxReservationsPerMonth  int       `gorm:"not null;default:15"      json:"max_reservations_per_month"`
	WaNotifEnabled           bool      `gorm:"not null;default:false"   json:"wa_notif_enabled"`
	WarningThresholdPct      int       `gorm:"not null;default:80"      json:"warning_threshold_pct"`
	Price                    float64   `gorm:"type:numeric(12,2);not null;default:0" json:"price"`
	IsActive                 bool      `gorm:"not null;default:true"    json:"is_active"`
	CreatedAt                time.Time `                                json:"created_at"`
	UpdatedAt                time.Time `                                json:"updated_at"`
}

func (Plan) TableName() string { return "tabl_plans" }
