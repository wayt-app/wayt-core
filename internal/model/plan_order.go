package model

import "time"

type PlanOrderStatus string

const (
	PlanOrderStatusPending   PlanOrderStatus = "pending"
	PlanOrderStatusCompleted PlanOrderStatus = "completed"
	PlanOrderStatusFailed    PlanOrderStatus = "failed"
)

type PlanOrder struct {
	ID              uint            `gorm:"primaryKey;autoIncrement" json:"id"`
	BusinessOwnerID uint            `gorm:"not null;index"           json:"business_owner_id"`
	PlanID          uint            `gorm:"not null"                 json:"plan_id"`
	Status          PlanOrderStatus `gorm:"not null;default:'pending'" json:"status"`
	ProcessAt       time.Time       `gorm:"not null"                 json:"process_at"`
	CreatedAt       time.Time       `                                json:"created_at"`
	UpdatedAt       time.Time       `                                json:"updated_at"`

	Plan *Plan `gorm:"foreignKey:PlanID" json:"plan,omitempty"`
}

func (PlanOrder) TableName() string { return "tabl_plan_orders" }
