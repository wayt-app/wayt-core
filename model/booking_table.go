package model

import "time"

// BookingTable records which physical table types (and how many) are assigned
// to a booking. One booking can span multiple table types within the same room.
type BookingTable struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	BookingID   uint      `gorm:"not null;index"           json:"booking_id"`
	TableTypeID uint      `gorm:"not null;index"           json:"table_type_id"`
	Count       int       `gorm:"not null;default:1"       json:"count"`
	CreatedAt   time.Time `                                json:"created_at"`

	TableType *TableType `gorm:"foreignKey:TableTypeID" json:"table_type,omitempty"`
}

func (BookingTable) TableName() string { return "tabl_booking_tables" }
