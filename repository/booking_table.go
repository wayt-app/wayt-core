package repository

import (
	"time"

	"github.com/wayt-app/wayt-core/model"
	"gorm.io/gorm"
)

type BookingTableRepository interface {
	CreateBatch(tables []model.BookingTable) error
	FindByBooking(bookingID uint) ([]model.BookingTable, error)
	DeleteByBooking(bookingID uint) error
	// SumByTypeIDsAndSlot returns total count per table_type_id for overlapping active bookings.
	SumByTypeIDsAndSlot(tableTypeIDs []uint, date time.Time, startTime, endTime string, excludeBookingID uint) (map[uint]int64, error)
}

type bookingTableRepository struct{ db *gorm.DB }

func NewBookingTableRepository(db *gorm.DB) BookingTableRepository {
	return &bookingTableRepository{db: db}
}

func (r *bookingTableRepository) CreateBatch(tables []model.BookingTable) error {
	if len(tables) == 0 {
		return nil
	}
	return r.db.Create(&tables).Error
}

func (r *bookingTableRepository) FindByBooking(bookingID uint) ([]model.BookingTable, error) {
	var list []model.BookingTable
	err := r.db.Where("booking_id = ?", bookingID).
		Preload("TableType").Find(&list).Error
	return list, err
}

func (r *bookingTableRepository) DeleteByBooking(bookingID uint) error {
	return r.db.Where("booking_id = ?", bookingID).Delete(&model.BookingTable{}).Error
}

func (r *bookingTableRepository) SumByTypeIDsAndSlot(
	tableTypeIDs []uint,
	date time.Time,
	startTime, endTime string,
	excludeBookingID uint,
) (map[uint]int64, error) {
	if len(tableTypeIDs) == 0 {
		return map[uint]int64{}, nil
	}
	type row struct {
		TableTypeID uint  `gorm:"column:table_type_id"`
		Total       int64 `gorm:"column:total"`
	}
	var rows []row
	q := r.db.Table("tabl_booking_tables bt").
		Select("bt.table_type_id, COALESCE(SUM(bt.count), 0) AS total").
		Joins("JOIN tabl_bookings b ON b.id = bt.booking_id").
		Where("bt.table_type_id IN ?", tableTypeIDs).
		Where("b.booking_date = ?", date.Format("2006-01-02")).
		Where("b.status IN ('pending','confirmed','checked_in')").
		Where("b.start_time < ? AND b.end_time > ?", endTime, startTime).
		Group("bt.table_type_id")
	if excludeBookingID > 0 {
		q = q.Where("b.id != ?", excludeBookingID)
	}
	if err := q.Scan(&rows).Error; err != nil {
		return nil, err
	}
	result := make(map[uint]int64, len(rows))
	for _, r := range rows {
		result[r.TableTypeID] = r.Total
	}
	return result, nil
}
