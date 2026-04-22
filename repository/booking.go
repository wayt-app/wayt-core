package repository

import (
	"time"

	"github.com/wayt-app/wayt-core/model"
	"gorm.io/gorm"
)

type BookingRepository interface {
	Create(b *model.Booking) error
	FindByID(id uint) (*model.Booking, error)
	FindByCustomer(customerID uint) ([]model.Booking, error)
	FindByBranch(branchID uint, date *time.Time, status *model.BookingStatus) ([]model.Booking, error)
	FindActiveByBranchDate(branchID uint, date time.Time) ([]model.Booking, error)
	CountOverlapping(tableTypeID uint, date time.Time, startTime, endTime string, excludeID uint) (int64, error)
	UpdateStatus(id uint, status model.BookingStatus) error
	UpdateStatusAndReason(id uint, status model.BookingStatus, reason string) error
	UpdateTableType(id uint, tableTypeID uint) error
	FindWaitingListForSlot(branchID uint, date time.Time, startTime, endTime string) ([]model.Booking, error)
	CountWaitingListBefore(bookingID uint, branchID uint, date time.Time, startTime string) (int64, error)
	// FindNoShowCandidates returns confirmed bookings whose start_time is more than `graceMinutes` ago.
	FindNoShowCandidates(graceMinutes int) ([]model.Booking, error)
	// ClearOverLimitByOwner resets is_over_limit=false for all active bookings belonging to owner's branches.
	ClearOverLimitByOwner(ownerID uint) error
	// MarkOverLimitByOwner marks active current-month bookings beyond the given limit position
	// (ordered by created_at ASC) as is_over_limit=true. Call after ClearOverLimitByOwner on downgrade.
	MarkOverLimitByOwner(ownerID uint, limit int) error
	// CountByStatusForDate returns a count per status for a branch on a given date.
	CountByStatusForDate(branchID uint, date time.Time) (map[model.BookingStatus]int64, error)
	// CountOverlappingByTableType returns booked count per table type for a given slot.
	CountOverlappingByTableType(branchID uint, date time.Time, startTime, endTime string) (map[uint]int64, error)
	// UpdateSchedule atomically updates booking_date, start_time, end_time, and status.
	UpdateSchedule(id uint, date time.Time, startTime, endTime string, status model.BookingStatus) error
	// FindReminderCandidates returns pending/confirmed bookings scheduled for tomorrow
	// whose reminder has not been sent yet.
	FindReminderCandidates() ([]model.Booking, error)
	// MarkReminderSent sets reminder_sent = true for the given booking.
	MarkReminderSent(id uint) error
	// FindByCustomerPaged returns a paginated booking list for a customer.
	// sortBy: "booking_date" or "created_at"; sortDir: "asc" or "desc".
	FindByCustomerPaged(customerID uint, sortBy, sortDir string, offset, limit int) ([]model.Booking, int64, error)
	// FindByBranchPaged returns a paginated booking list for a branch with optional filters.
	// search filters by customer name or phone (case-insensitive).
	// sortBy: "booking_date" or "created_at"; sortDir: "asc" or "desc".
	FindByBranchPaged(branchID uint, date *time.Time, status *model.BookingStatus, search, sortBy, sortDir string, offset, limit int) ([]model.Booking, int64, error)
}

type bookingRepository struct{ db *gorm.DB }

func NewBookingRepository(db *gorm.DB) BookingRepository {
	return &bookingRepository{db: db}
}

func (r *bookingRepository) Create(b *model.Booking) error {
	return r.db.Create(b).Error
}

func (r *bookingRepository) FindByID(id uint) (*model.Booking, error) {
	var b model.Booking
	err := r.db.Preload("Customer").Preload("Branch").Preload("TableType").
		Where("id = ?", id).First(&b).Error
	return &b, err
}

func (r *bookingRepository) FindByCustomer(customerID uint) ([]model.Booking, error) {
	var list []model.Booking
	err := r.db.Preload("Branch").Preload("TableType").
		Where("customer_id = ?", customerID).
		Order("booking_date DESC, start_time DESC").Find(&list).Error
	return list, err
}

// FindActiveByBranchDate returns all pending/confirmed/checked_in bookings for a branch on a date.
// Used by slot service to compute availability in-memory with a single query.
func (r *bookingRepository) FindActiveByBranchDate(branchID uint, date time.Time) ([]model.Booking, error) {
	var list []model.Booking
	err := r.db.Select("id, table_type_id, start_time, end_time, tables_count").
		Where("branch_id = ?", branchID).
		Where("booking_date = ?", date.Format("2006-01-02")).
		Where("status IN ('pending','confirmed','checked_in')").
		Find(&list).Error
	return list, err
}

func (r *bookingRepository) FindByBranch(branchID uint, date *time.Time, status *model.BookingStatus) ([]model.Booking, error) {
	var list []model.Booking
	q := r.db.Preload("Customer").Preload("TableType").
		Where("branch_id = ?", branchID)
	if date != nil {
		q = q.Where("booking_date = ?", date.Format("2006-01-02"))
	}
	if status != nil {
		q = q.Where("status = ?", *status)
	}
	err := q.Order("booking_date ASC, start_time ASC").Find(&list).Error
	return list, err
}

// CountOverlapping returns the total tables consumed (SUM of tables_count) for overlapping bookings.
// This accounts for combined-table bookings properly.
func (r *bookingRepository) CountOverlapping(tableTypeID uint, date time.Time, startTime, endTime string, excludeID uint) (int64, error) {
	var total *int64
	q := r.db.Model(&model.Booking{}).
		Select("COALESCE(SUM(tables_count), 0)").
		Where("table_type_id = ?", tableTypeID).
		Where("booking_date = ?", date.Format("2006-01-02")).
		Where("status IN ('pending','confirmed','checked_in')").
		Where("start_time < ? AND end_time > ?", endTime, startTime)
	if excludeID > 0 {
		q = q.Where("id != ?", excludeID)
	}
	err := q.Scan(&total).Error
	if total == nil {
		return 0, err
	}
	return *total, err
}

func (r *bookingRepository) UpdateStatus(id uint, status model.BookingStatus) error {
	return r.db.Model(&model.Booking{}).Where("id = ?", id).
		Update("status", status).Error
}

func (r *bookingRepository) UpdateStatusAndReason(id uint, status model.BookingStatus, reason string) error {
	return r.db.Model(&model.Booking{}).Where("id = ?", id).
		Updates(map[string]interface{}{"status": status, "cancel_reason": reason}).Error
}

func (r *bookingRepository) UpdateTableType(id uint, tableTypeID uint) error {
	return r.db.Model(&model.Booking{}).Where("id = ?", id).
		Update("table_type_id", tableTypeID).Error
}

// FindWaitingListForSlot returns waiting_list bookings for a branch that overlap the given slot,
// ordered by created_at ASC (earliest waiter first).
func (r *bookingRepository) FindWaitingListForSlot(branchID uint, date time.Time, startTime, endTime string) ([]model.Booking, error) {
	var list []model.Booking
	err := r.db.Where("branch_id = ?", branchID).
		Where("booking_date = ?", date.Format("2006-01-02")).
		Where("status = 'waiting_list'").
		Where("start_time = ? AND end_time = ?", startTime, endTime).
		Order("created_at ASC").Find(&list).Error
	return list, err
}

// FindNoShowCandidates returns confirmed bookings whose booking_date+start_time is more
// than graceMinutes minutes in the past (i.e., customer never showed up).
func (r *bookingRepository) FindNoShowCandidates(graceMinutes int) ([]model.Booking, error) {
	threshold := time.Now().Add(-time.Duration(graceMinutes) * time.Minute)
	dateStr := threshold.Format("2006-01-02")
	timeStr := threshold.Format("15:04")
	var list []model.Booking
	err := r.db.Where("status = 'confirmed'").
		Where("booking_date < ? OR (booking_date = ? AND start_time <= ?)", dateStr, dateStr, timeStr).
		Find(&list).Error
	return list, err
}

func (r *bookingRepository) ClearOverLimitByOwner(ownerID uint) error {
	return r.db.Exec(`
		UPDATE tabl_bookings SET is_over_limit = FALSE, updated_at = NOW()
		WHERE is_over_limit = TRUE
		AND status IN ('pending','confirmed','waiting_list')
		AND branch_id IN (
			SELECT b.id FROM tabl_branches b
			JOIN tabl_restaurants rest ON rest.id = b.restaurant_id
			WHERE rest.business_owner_id = ? AND b.deleted_at IS NULL AND rest.deleted_at IS NULL
		)`, ownerID).Error
}

func (r *bookingRepository) MarkOverLimitByOwner(ownerID uint, limit int) error {
	return r.db.Exec(`
		UPDATE tabl_bookings SET is_over_limit = TRUE, updated_at = NOW()
		WHERE id IN (
			SELECT id FROM tabl_bookings
			WHERE branch_id IN (
				SELECT b.id FROM tabl_branches b
				JOIN tabl_restaurants rest ON rest.id = b.restaurant_id
				WHERE rest.business_owner_id = ? AND b.deleted_at IS NULL AND rest.deleted_at IS NULL
			)
			AND status IN ('pending','confirmed','waiting_list')
			AND DATE_TRUNC('month', booking_date::timestamp) = DATE_TRUNC('month', NOW())
			ORDER BY created_at ASC
			OFFSET ?
		)`, ownerID, limit).Error
}

// CountByStatusForDate returns a count per status for a branch on a given date.
func (r *bookingRepository) CountByStatusForDate(branchID uint, date time.Time) (map[model.BookingStatus]int64, error) {
	type row struct {
		Status model.BookingStatus
		Count  int64
	}
	var rows []row
	err := r.db.Model(&model.Booking{}).
		Select("status, COUNT(*) as count").
		Where("branch_id = ?", branchID).
		Where("booking_date = ?", date.Format("2006-01-02")).
		Group("status").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	result := make(map[model.BookingStatus]int64)
	for _, rw := range rows {
		result[rw.Status] = rw.Count
	}
	return result, nil
}

// CountOverlappingByTableType returns total tables consumed (SUM of tables_count) per table_type_id for a given slot.
func (r *bookingRepository) CountOverlappingByTableType(branchID uint, date time.Time, startTime, endTime string) (map[uint]int64, error) {
	type row struct {
		TableTypeID uint
		Count       int64
	}
	var rows []row
	err := r.db.Model(&model.Booking{}).
		Select("table_type_id, COALESCE(SUM(tables_count), 0) as count").
		Where("branch_id = ?", branchID).
		Where("booking_date = ?", date.Format("2006-01-02")).
		Where("status IN ('pending','confirmed','checked_in')").
		Where("start_time < ? AND end_time > ?", endTime, startTime).
		Group("table_type_id").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	result := make(map[uint]int64)
	for _, rw := range rows {
		result[rw.TableTypeID] = rw.Count
	}
	return result, nil
}

// CountWaitingListBefore counts how many waiting_list bookings for the same branch/date/start
// were created before this booking (i.e., the position in queue, 0-indexed).
func (r *bookingRepository) CountWaitingListBefore(bookingID uint, branchID uint, date time.Time, startTime string) (int64, error) {
	var b model.Booking
	if err := r.db.Where("id = ?", bookingID).First(&b).Error; err != nil {
		return 0, err
	}
	var count int64
	err := r.db.Model(&model.Booking{}).
		Where("branch_id = ?", branchID).
		Where("booking_date = ?", date.Format("2006-01-02")).
		Where("start_time = ?", startTime).
		Where("status = 'waiting_list'").
		Where("created_at < ?", b.CreatedAt).
		Count(&count).Error
	return count, err
}

func (r *bookingRepository) UpdateSchedule(id uint, date time.Time, startTime, endTime string, status model.BookingStatus) error {
	return r.db.Model(&model.Booking{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"booking_date": date,
			"start_time":   startTime,
			"end_time":     endTime,
			"status":       status,
		}).Error
}

// FindReminderCandidates returns pending/confirmed bookings for tomorrow whose reminder_sent = false.
func (r *bookingRepository) FindReminderCandidates() ([]model.Booking, error) {
	tomorrow := time.Now().AddDate(0, 0, 1).Format("2006-01-02")
	var list []model.Booking
	err := r.db.Preload("Customer").Preload("Branch").
		Where("booking_date = ?", tomorrow).
		Where("status IN ('pending','confirmed')").
		Where("reminder_sent = false").
		Find(&list).Error
	return list, err
}

func (r *bookingRepository) MarkReminderSent(id uint) error {
	return r.db.Model(&model.Booking{}).Where("id = ?", id).
		Update("reminder_sent", true).Error
}

// bookingOrderClause returns a safe ORDER BY clause from whitelisted sort params.
func bookingOrderClause(col, dir, tablePrefix string) string {
	allowedCols := map[string]string{
		"booking_date": tablePrefix + "booking_date",
		"created_at":   tablePrefix + "created_at",
	}
	orderCol, ok := allowedCols[col]
	if !ok {
		orderCol = tablePrefix + "booking_date"
	}
	if dir != "asc" {
		dir = "desc"
	}
	secondary := tablePrefix + "start_time " + dir
	return orderCol + " " + dir + ", " + secondary
}

func (r *bookingRepository) FindByCustomerPaged(customerID uint, sortBy, sortDir string, offset, limit int) ([]model.Booking, int64, error) {
	var list []model.Booking
	var total int64
	q := r.db.Model(&model.Booking{}).Where("customer_id = ?", customerID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := r.db.Preload("Branch").Preload("TableType").
		Where("customer_id = ?", customerID).
		Order(bookingOrderClause(sortBy, sortDir, "")).
		Offset(offset).Limit(limit).Find(&list).Error
	return list, total, err
}

func (r *bookingRepository) FindByBranchPaged(branchID uint, date *time.Time, status *model.BookingStatus, search, sortBy, sortDir string, offset, limit int) ([]model.Booking, int64, error) {
	var list []model.Booking
	var total int64

	applyFilters := func(db *gorm.DB) *gorm.DB {
		db = db.Where("tabl_bookings.branch_id = ?", branchID)
		if date != nil {
			db = db.Where("tabl_bookings.booking_date = ?", date.Format("2006-01-02"))
		}
		if status != nil {
			db = db.Where("tabl_bookings.status = ?", *status)
		}
		if search != "" {
			like := "%" + search + "%"
			db = db.Joins("JOIN tabl_customers ON tabl_customers.id = tabl_bookings.customer_id").
				Where("tabl_customers.name ILIKE ? OR tabl_customers.phone ILIKE ?", like, like)
		}
		return db
	}

	if err := applyFilters(r.db.Model(&model.Booking{})).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	q := r.db.Preload("Customer").Preload("TableType")
	q = applyFilters(q)
	err := q.Order(bookingOrderClause(sortBy, sortDir, "tabl_bookings.")).
		Offset(offset).Limit(limit).Find(&list).Error
	return list, total, err
}
