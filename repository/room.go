package repository

import (
	"github.com/wayt-app/wayt-core/model"
	"gorm.io/gorm"
)

type RoomRepository interface {
	Create(r *model.Room) error
	FindByID(id uint) (*model.Room, error)
	FindByBranch(branchID uint) ([]model.Room, error)
	Update(r *model.Room) error
	Delete(id uint) error
	CountActiveBookings(roomID uint) (int64, error)
}

type roomRepository struct{ db *gorm.DB }

func NewRoomRepository(db *gorm.DB) RoomRepository {
	return &roomRepository{db: db}
}

func (r *roomRepository) Create(room *model.Room) error {
	return r.db.Create(room).Error
}

func (r *roomRepository) FindByID(id uint) (*model.Room, error) {
	var room model.Room
	return &room, r.db.Where("id = ?", id).First(&room).Error
}

func (r *roomRepository) FindByBranch(branchID uint) ([]model.Room, error) {
	var list []model.Room
	return list, r.db.Where("branch_id = ?", branchID).Order("is_default DESC, id ASC").Find(&list).Error
}

func (r *roomRepository) Update(room *model.Room) error {
	return r.db.Save(room).Error
}

func (r *roomRepository) Delete(id uint) error {
	return r.db.Where("id = ?", id).Delete(&model.Room{}).Error
}

// CountActiveBookings counts pending/confirmed/checked_in/waiting_list bookings for a room.
func (r *roomRepository) CountActiveBookings(roomID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.Booking{}).
		Where("room_id = ?", roomID).
		Where("status IN ('pending','confirmed','checked_in','waiting_list')").
		Count(&count).Error
	return count, err
}
