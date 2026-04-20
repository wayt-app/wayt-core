package model

import "time"

type BookingStatus string

const (
	BookingStatusPending     BookingStatus = "pending"
	BookingStatusConfirmed   BookingStatus = "confirmed"
	BookingStatusCompleted   BookingStatus = "completed"
	BookingStatusCancelled   BookingStatus = "cancelled"
	BookingStatusWaitingList BookingStatus = "waiting_list"
	BookingStatusNoShow      BookingStatus = "no_show"
	BookingStatusCheckedIn   BookingStatus = "checked_in"
)

type Booking struct {
	ID          uint          `gorm:"primaryKey;autoIncrement"           json:"id"`
	CustomerID  uint          `gorm:"not null;index"                     json:"customer_id"`
	BranchID    uint          `gorm:"not null;index"                     json:"branch_id"`
	TableTypeID uint          `gorm:"not null;index"                     json:"table_type_id"`
	BookingDate time.Time     `gorm:"type:date;not null"                 json:"booking_date"`
	StartTime   string        `gorm:"type:time;not null"                 json:"start_time"` // "HH:MM"
	EndTime     string        `gorm:"type:time;not null"                 json:"end_time"`
	GuestCount  int           `gorm:"not null"                           json:"guest_count"`
	TablesCount int           `gorm:"not null;default:1"                 json:"tables_count"`
	Status      BookingStatus `gorm:"type:tabl_booking_status;default:'pending'" json:"status"`
	Notes        string        `gorm:"type:text"                          json:"notes"`
	CancelReason string        `gorm:"type:text"                          json:"cancel_reason,omitempty"`
	IsOverLimit  bool          `gorm:"not null;default:false"             json:"is_over_limit"`
	ReminderSent bool          `gorm:"not null;default:false"             json:"-"`
	CreatedAt    time.Time     `                                          json:"created_at"`
	UpdatedAt   time.Time     `                                          json:"updated_at"`

	Customer  *Customer  `gorm:"foreignKey:CustomerID"  json:"customer,omitempty"`
	Branch    *Branch    `gorm:"foreignKey:BranchID"    json:"branch,omitempty"`
	TableType *TableType `gorm:"foreignKey:TableTypeID" json:"table_type,omitempty"`
}

func (Booking) TableName() string { return "tabl_bookings" }
