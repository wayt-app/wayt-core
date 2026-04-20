package model

import "time"

type SubscriptionStatus string

const (
	SubscriptionStatusTrial           SubscriptionStatus = "trial"
	SubscriptionStatusPendingApproval SubscriptionStatus = "pending_approval"
	SubscriptionStatusActive          SubscriptionStatus = "active"
	SubscriptionStatusSuspended       SubscriptionStatus = "suspended"
	SubscriptionStatusExpired         SubscriptionStatus = "expired"
)

type Subscription struct {
	ID              uint               `gorm:"primaryKey;autoIncrement" json:"id"`
	BusinessOwnerID uint               `gorm:"not null;index"           json:"business_owner_id"`
	PlanID          uint               `gorm:"not null"                 json:"plan_id"`
	Status          SubscriptionStatus `gorm:"not null;default:'trial'" json:"status"`
	TrialStartedAt  *time.Time         `                                json:"trial_started_at"`
	TrialEndsAt     *time.Time         `                                json:"trial_ends_at"`
	ActivatedAt     *time.Time         `                                json:"activated_at"`
	Notes           string             `gorm:"type:text"                json:"notes"`
	ReservationsThisMonth int          `gorm:"not null;default:0"       json:"reservations_this_month"`
	LastResetAt     time.Time          `gorm:"type:date;not null"       json:"last_reset_at"`
	WarningSent     bool               `gorm:"not null;default:false"   json:"warning_sent"`
	CreatedAt       time.Time          `                                json:"created_at"`
	UpdatedAt       time.Time          `                                json:"updated_at"`

	Plan          *Plan          `gorm:"foreignKey:PlanID"          json:"plan,omitempty"`
	BusinessOwner *BusinessOwner `gorm:"foreignKey:BusinessOwnerID" json:"business_owner,omitempty"`
}

func (Subscription) TableName() string { return "tabl_subscriptions" }
