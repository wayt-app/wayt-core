package model

import "time"

type TableType struct {
	ID          uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	BranchID    uint       `gorm:"not null;index"           json:"branch_id"`
	RoomID      *uint      `gorm:"index"                    json:"room_id,omitempty"`
	Name        string     `gorm:"size:100;not null"        json:"name"`
	Capacity    int        `gorm:"not null"                 json:"capacity"`
	IsActive    bool       `gorm:"default:true"             json:"is_active"`
	CreatedAt   time.Time  `                                json:"created_at"`
	UpdatedAt   time.Time  `                                json:"updated_at"`
	DeletedAt   *time.Time `gorm:"index"                    json:"deleted_at,omitempty"`

	Branch *Branch `gorm:"foreignKey:BranchID" json:"branch,omitempty"`
	Room   *Room   `gorm:"foreignKey:RoomID"   json:"room,omitempty"`
}

func (TableType) TableName() string { return "tabl_table_types" }
