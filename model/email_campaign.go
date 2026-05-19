package model

import "time"

type CampaignStatus string

const (
	CampaignStatusScheduled CampaignStatus = "scheduled"
	CampaignStatusSending   CampaignStatus = "sending"
	CampaignStatusSent      CampaignStatus = "sent"
	CampaignStatusFailed    CampaignStatus = "failed"
	CampaignStatusCancelled CampaignStatus = "cancelled"
)

type EmailCampaign struct {
	ID             uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	RestaurantID   uint           `gorm:"not null;index"           json:"restaurant_id"`
	Subject        string         `gorm:"size:200;not null"        json:"subject"`
	Body           string         `gorm:"type:text;not null"       json:"body"`
	FilterBranchID *uint          `gorm:"index"                    json:"filter_branch_id,omitempty"`
	FilterSegment  string         `gorm:"size:20;not null;default:'all'" json:"filter_segment"`
	Status         CampaignStatus `gorm:"type:varchar(20);not null;default:'scheduled'" json:"status"`
	ScheduledAt    *time.Time     `                                json:"scheduled_at,omitempty"`
	SentAt         *time.Time     `                                json:"sent_at,omitempty"`
	RecipientCount int            `gorm:"not null;default:0"       json:"recipient_count"`
	SuccessCount   int            `gorm:"not null;default:0"       json:"success_count"`
	FailCount      int            `gorm:"not null;default:0"       json:"fail_count"`
	CreatedAt      time.Time      `                                json:"created_at"`
	UpdatedAt      time.Time      `                                json:"updated_at"`

	Restaurant *Restaurant `gorm:"foreignKey:RestaurantID" json:"restaurant,omitempty"`
}

func (EmailCampaign) TableName() string { return "tabl_email_campaigns" }
