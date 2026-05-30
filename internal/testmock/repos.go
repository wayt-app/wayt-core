// Package testmock provides lightweight mock implementations of repository and service
// interfaces for use in unit tests. Each mock has optional function fields; unset fields
// return zero values and nil errors so tests only override what they need.
package testmock

import (
	"time"

	"github.com/wayt-app/wayt-core/model"
	"github.com/wayt-app/wayt-core/repository"
)

// ---- BookingRepository ----

type BookingRepo struct {
	CreateFn                      func(b *model.Booking) error
	FindByIDFn                    func(id uint) (*model.Booking, error)
	FindByCustomerFn              func(customerID uint) ([]model.Booking, error)
	FindByBranchFn                func(branchID uint, date *time.Time, status *model.BookingStatus) ([]model.Booking, error)
	FindActiveByBranchDateFn      func(branchID uint, date time.Time) ([]model.Booking, error)
	CountOverlappingFn            func(tableTypeID uint, date time.Time, startTime, endTime string, excludeID uint) (int64, error)
	UpdateStatusFn                func(id uint, status model.BookingStatus) error
	UpdateStatusAndReasonFn       func(id uint, status model.BookingStatus, reason string) error
	UpdateTableTypeFn             func(id uint, tableTypeID uint) error
	UpdateTableTypeWithCountFn    func(id, tableTypeID uint, tablesCount int, roomID *uint) error
	FindWaitingListForSlotFn      func(branchID uint, date time.Time, startTime, endTime string) ([]model.Booking, error)
	CountWaitingListBeforeFn      func(bookingID uint, branchID uint, date time.Time, startTime string) (int64, error)
	FindNoShowCandidatesFn        func(graceMinutes int) ([]model.Booking, error)
	ClearOverLimitByOwnerFn       func(ownerID uint) error
	MarkOverLimitByOwnerFn        func(ownerID uint, limit int) error
	CountByStatusForDateFn        func(branchID uint, date time.Time) (map[model.BookingStatus]int64, error)
	CountOverlappingByTableTypeFn func(branchID uint, date time.Time, startTime, endTime string) (map[uint]int64, error)
	UpdateScheduleFn              func(id uint, date time.Time, startTime, endTime string, status model.BookingStatus) error
	UpdateDetailsFn               func(id uint, date time.Time, startTime, endTime string, guestCount int, notes string) error
	FindReminderCandidatesFn      func() ([]model.Booking, error)
	MarkReminderSentFn            func(id uint) error
	FindByCustomerPagedFn         func(customerID uint, sortBy, sortDir string, offset, limit int) ([]model.Booking, int64, error)
	FindByBranchPagedFn           func(branchID uint, date *time.Time, status *model.BookingStatus, search, sortBy, sortDir string, offset, limit int) ([]model.Booking, int64, error)
	CompleteWithDetailsFn         func(id uint, notes string, totalBill int64) error
	ListCustomerSummaryFn         func(restaurantID uint) ([]repository.CustomerSummaryRow, error)
	UpdateOrderStatusFn           func(id uint, status string) error
	UpdatePaymentProofFn          func(id uint, url string) error
	CountOverlappingByGroupFn     func(tableTypeIDs []uint, date time.Time, startTime, endTime string, excludeID uint) (int64, error)
	DeleteFn                      func(id uint) error
}

func (m *BookingRepo) Create(b *model.Booking) error {
	if m.CreateFn != nil {
		return m.CreateFn(b)
	}
	return nil
}
func (m *BookingRepo) FindByID(id uint) (*model.Booking, error) {
	if m.FindByIDFn != nil {
		return m.FindByIDFn(id)
	}
	return nil, nil
}
func (m *BookingRepo) FindByCustomer(customerID uint) ([]model.Booking, error) {
	if m.FindByCustomerFn != nil {
		return m.FindByCustomerFn(customerID)
	}
	return nil, nil
}
func (m *BookingRepo) FindByBranch(branchID uint, date *time.Time, status *model.BookingStatus) ([]model.Booking, error) {
	if m.FindByBranchFn != nil {
		return m.FindByBranchFn(branchID, date, status)
	}
	return nil, nil
}
func (m *BookingRepo) FindActiveByBranchDate(branchID uint, date time.Time) ([]model.Booking, error) {
	if m.FindActiveByBranchDateFn != nil {
		return m.FindActiveByBranchDateFn(branchID, date)
	}
	return nil, nil
}
func (m *BookingRepo) CountOverlapping(tableTypeID uint, date time.Time, startTime, endTime string, excludeID uint) (int64, error) {
	if m.CountOverlappingFn != nil {
		return m.CountOverlappingFn(tableTypeID, date, startTime, endTime, excludeID)
	}
	return 0, nil
}
func (m *BookingRepo) UpdateStatus(id uint, status model.BookingStatus) error {
	if m.UpdateStatusFn != nil {
		return m.UpdateStatusFn(id, status)
	}
	return nil
}
func (m *BookingRepo) UpdateStatusAndReason(id uint, status model.BookingStatus, reason string) error {
	if m.UpdateStatusAndReasonFn != nil {
		return m.UpdateStatusAndReasonFn(id, status, reason)
	}
	return nil
}
func (m *BookingRepo) UpdateTableType(id uint, tableTypeID uint) error {
	if m.UpdateTableTypeFn != nil {
		return m.UpdateTableTypeFn(id, tableTypeID)
	}
	return nil
}
func (m *BookingRepo) UpdateTableTypeWithCount(id, tableTypeID uint, tablesCount int, roomID *uint) error {
	if m.UpdateTableTypeWithCountFn != nil {
		return m.UpdateTableTypeWithCountFn(id, tableTypeID, tablesCount, roomID)
	}
	return nil
}
func (m *BookingRepo) FindWaitingListForSlot(branchID uint, date time.Time, startTime, endTime string) ([]model.Booking, error) {
	if m.FindWaitingListForSlotFn != nil {
		return m.FindWaitingListForSlotFn(branchID, date, startTime, endTime)
	}
	return nil, nil
}
func (m *BookingRepo) CountWaitingListBefore(bookingID uint, branchID uint, date time.Time, startTime string) (int64, error) {
	if m.CountWaitingListBeforeFn != nil {
		return m.CountWaitingListBeforeFn(bookingID, branchID, date, startTime)
	}
	return 0, nil
}
func (m *BookingRepo) FindNoShowCandidates(graceMinutes int) ([]model.Booking, error) {
	if m.FindNoShowCandidatesFn != nil {
		return m.FindNoShowCandidatesFn(graceMinutes)
	}
	return nil, nil
}
func (m *BookingRepo) ClearOverLimitByOwner(ownerID uint) error {
	if m.ClearOverLimitByOwnerFn != nil {
		return m.ClearOverLimitByOwnerFn(ownerID)
	}
	return nil
}
func (m *BookingRepo) MarkOverLimitByOwner(ownerID uint, limit int) error {
	if m.MarkOverLimitByOwnerFn != nil {
		return m.MarkOverLimitByOwnerFn(ownerID, limit)
	}
	return nil
}
func (m *BookingRepo) CountByStatusForDate(branchID uint, date time.Time) (map[model.BookingStatus]int64, error) {
	if m.CountByStatusForDateFn != nil {
		return m.CountByStatusForDateFn(branchID, date)
	}
	return map[model.BookingStatus]int64{}, nil
}
func (m *BookingRepo) CountOverlappingByTableType(branchID uint, date time.Time, startTime, endTime string) (map[uint]int64, error) {
	if m.CountOverlappingByTableTypeFn != nil {
		return m.CountOverlappingByTableTypeFn(branchID, date, startTime, endTime)
	}
	return map[uint]int64{}, nil
}
func (m *BookingRepo) UpdateSchedule(id uint, date time.Time, startTime, endTime string, status model.BookingStatus) error {
	if m.UpdateScheduleFn != nil {
		return m.UpdateScheduleFn(id, date, startTime, endTime, status)
	}
	return nil
}
func (m *BookingRepo) UpdateDetails(id uint, date time.Time, startTime, endTime string, guestCount int, notes string) error {
	if m.UpdateDetailsFn != nil {
		return m.UpdateDetailsFn(id, date, startTime, endTime, guestCount, notes)
	}
	return nil
}
func (m *BookingRepo) FindReminderCandidates() ([]model.Booking, error) {
	if m.FindReminderCandidatesFn != nil {
		return m.FindReminderCandidatesFn()
	}
	return nil, nil
}
func (m *BookingRepo) MarkReminderSent(id uint) error {
	if m.MarkReminderSentFn != nil {
		return m.MarkReminderSentFn(id)
	}
	return nil
}
func (m *BookingRepo) FindByCustomerPaged(customerID uint, sortBy, sortDir string, offset, limit int) ([]model.Booking, int64, error) {
	if m.FindByCustomerPagedFn != nil {
		return m.FindByCustomerPagedFn(customerID, sortBy, sortDir, offset, limit)
	}
	return nil, 0, nil
}
func (m *BookingRepo) FindByBranchPaged(branchID uint, date *time.Time, status *model.BookingStatus, search, sortBy, sortDir string, offset, limit int) ([]model.Booking, int64, error) {
	if m.FindByBranchPagedFn != nil {
		return m.FindByBranchPagedFn(branchID, date, status, search, sortBy, sortDir, offset, limit)
	}
	return nil, 0, nil
}
func (m *BookingRepo) CompleteWithDetails(id uint, notes string, totalBill int64) error {
	if m.CompleteWithDetailsFn != nil {
		return m.CompleteWithDetailsFn(id, notes, totalBill)
	}
	return nil
}
func (m *BookingRepo) ListCustomerSummaryByRestaurant(restaurantID uint) ([]repository.CustomerSummaryRow, error) {
	if m.ListCustomerSummaryFn != nil {
		return m.ListCustomerSummaryFn(restaurantID)
	}
	return nil, nil
}
func (m *BookingRepo) UpdateOrderStatus(id uint, status string) error {
	if m.UpdateOrderStatusFn != nil {
		return m.UpdateOrderStatusFn(id, status)
	}
	return nil
}
func (m *BookingRepo) UpdatePaymentProof(id uint, url string) error {
	if m.UpdatePaymentProofFn != nil {
		return m.UpdatePaymentProofFn(id, url)
	}
	return nil
}
func (m *BookingRepo) CountOverlappingByGroup(tableTypeIDs []uint, date time.Time, startTime, endTime string, excludeID uint) (int64, error) {
	if m.CountOverlappingByGroupFn != nil {
		return m.CountOverlappingByGroupFn(tableTypeIDs, date, startTime, endTime, excludeID)
	}
	return 0, nil
}
func (m *BookingRepo) Delete(id uint) error {
	if m.DeleteFn != nil {
		return m.DeleteFn(id)
	}
	return nil
}

// ---- BranchRepository ----

type BranchRepo struct {
	CreateFn               func(b *model.Branch) error
	FindAllFn              func() ([]model.Branch, error)
	FindByRestaurantFn     func(restaurantID uint) ([]model.Branch, error)
	FindActiveByRestaurantFn func(restaurantID uint) ([]model.Branch, error)
	FindByIDFn             func(id uint) (*model.Branch, error)
	UpdateFn               func(b *model.Branch) error
	DeleteFn               func(id uint) error
}

func (m *BranchRepo) Create(b *model.Branch) error {
	if m.CreateFn != nil {
		return m.CreateFn(b)
	}
	return nil
}
func (m *BranchRepo) FindAll() ([]model.Branch, error) {
	if m.FindAllFn != nil {
		return m.FindAllFn()
	}
	return nil, nil
}
func (m *BranchRepo) FindByRestaurant(restaurantID uint) ([]model.Branch, error) {
	if m.FindByRestaurantFn != nil {
		return m.FindByRestaurantFn(restaurantID)
	}
	return nil, nil
}
func (m *BranchRepo) FindActiveByRestaurant(restaurantID uint) ([]model.Branch, error) {
	if m.FindActiveByRestaurantFn != nil {
		return m.FindActiveByRestaurantFn(restaurantID)
	}
	return nil, nil
}
func (m *BranchRepo) FindByID(id uint) (*model.Branch, error) {
	if m.FindByIDFn != nil {
		return m.FindByIDFn(id)
	}
	return nil, nil
}
func (m *BranchRepo) Update(b *model.Branch) error {
	if m.UpdateFn != nil {
		return m.UpdateFn(b)
	}
	return nil
}
func (m *BranchRepo) Delete(id uint) error {
	if m.DeleteFn != nil {
		return m.DeleteFn(id)
	}
	return nil
}

// ---- TableTypeRepository ----

type TableTypeRepo struct {
	CreateFn             func(t *model.TableType) error
	FindByBranchFn       func(branchID uint) ([]model.TableType, error)
	FindByBranchWithRoomFn func(branchID uint) ([]model.TableType, error)
	FindByBranchAndRoomFn func(branchID uint, roomID *uint) ([]model.TableType, error)
	FindByGroupFn        func(branchID uint, name string, capacity int, roomID *uint) ([]model.TableType, error)
	FindByIDFn           func(id uint) (*model.TableType, error)
	UpdateFn             func(t *model.TableType) error
	DeleteFn             func(id uint) error
}

func (m *TableTypeRepo) Create(t *model.TableType) error {
	if m.CreateFn != nil {
		return m.CreateFn(t)
	}
	return nil
}
func (m *TableTypeRepo) FindByBranch(branchID uint) ([]model.TableType, error) {
	if m.FindByBranchFn != nil {
		return m.FindByBranchFn(branchID)
	}
	return nil, nil
}
func (m *TableTypeRepo) FindByBranchWithRoom(branchID uint) ([]model.TableType, error) {
	if m.FindByBranchWithRoomFn != nil {
		return m.FindByBranchWithRoomFn(branchID)
	}
	return nil, nil
}
func (m *TableTypeRepo) FindByBranchAndRoom(branchID uint, roomID *uint) ([]model.TableType, error) {
	if m.FindByBranchAndRoomFn != nil {
		return m.FindByBranchAndRoomFn(branchID, roomID)
	}
	return nil, nil
}
func (m *TableTypeRepo) FindByGroup(branchID uint, name string, capacity int, roomID *uint) ([]model.TableType, error) {
	if m.FindByGroupFn != nil {
		return m.FindByGroupFn(branchID, name, capacity, roomID)
	}
	return nil, nil
}
func (m *TableTypeRepo) FindByID(id uint) (*model.TableType, error) {
	if m.FindByIDFn != nil {
		return m.FindByIDFn(id)
	}
	return nil, nil
}
func (m *TableTypeRepo) Update(t *model.TableType) error {
	if m.UpdateFn != nil {
		return m.UpdateFn(t)
	}
	return nil
}
func (m *TableTypeRepo) Delete(id uint) error {
	if m.DeleteFn != nil {
		return m.DeleteFn(id)
	}
	return nil
}

// ---- BookingTableRepository ----

type BookingTableRepo struct {
	CreateBatchFn        func(tables []model.BookingTable) error
	FindByBookingFn      func(bookingID uint) ([]model.BookingTable, error)
	DeleteByBookingFn    func(bookingID uint) error
	SumByTypeIDsAndSlotFn func(tableTypeIDs []uint, date time.Time, startTime, endTime string, excludeBookingID uint) (map[uint]int64, error)
}

func (m *BookingTableRepo) CreateBatch(tables []model.BookingTable) error {
	if m.CreateBatchFn != nil {
		return m.CreateBatchFn(tables)
	}
	return nil
}
func (m *BookingTableRepo) FindByBooking(bookingID uint) ([]model.BookingTable, error) {
	if m.FindByBookingFn != nil {
		return m.FindByBookingFn(bookingID)
	}
	return nil, nil
}
func (m *BookingTableRepo) DeleteByBooking(bookingID uint) error {
	if m.DeleteByBookingFn != nil {
		return m.DeleteByBookingFn(bookingID)
	}
	return nil
}
func (m *BookingTableRepo) SumByTypeIDsAndSlot(tableTypeIDs []uint, date time.Time, startTime, endTime string, excludeBookingID uint) (map[uint]int64, error) {
	if m.SumByTypeIDsAndSlotFn != nil {
		return m.SumByTypeIDsAndSlotFn(tableTypeIDs, date, startTime, endTime, excludeBookingID)
	}
	return map[uint]int64{}, nil
}

// ---- CustomerRepository ----

type CustomerRepo struct {
	CreateFn                  func(c *model.Customer) error
	FindByEmailFn             func(email string) (*model.Customer, error)
	FindByIDFn                func(id uint) (*model.Customer, error)
	UpdatePasswordFn          func(id uint, hashedPassword string) error
	ListFn                    func() ([]model.Customer, error)
	ListPagedFn               func(search string, page, limit int) ([]model.Customer, int64, error)
	SetResetTokenFn           func(id uint, token string, expiresAt time.Time) error
	FindByResetTokenFn        func(token string) (*model.Customer, error)
	ClearResetTokenFn         func(id uint) error
	SetVerificationTokenFn    func(id uint, token string) error
	FindByVerificationTokenFn func(token string) (*model.Customer, error)
	MarkVerifiedFn            func(id uint) error
	FindTokenVersionFn        func(id uint) (int, error)
	IncrementTokenVersionFn   func(id uint) error
	UpdateProfileFn           func(id uint, name, phone, address string) error
	UpdateAvatarURLFn         func(id uint, url string) error
	FindByGoogleIDFn          func(googleID string) (*model.Customer, error)
	SetGoogleInfoFn           func(id uint, googleID, avatarURL string) error
	FindByPhoneFn             func(phone string) (*model.Customer, error)
}

func (m *CustomerRepo) Create(c *model.Customer) error {
	if m.CreateFn != nil {
		return m.CreateFn(c)
	}
	return nil
}
func (m *CustomerRepo) FindByEmail(email string) (*model.Customer, error) {
	if m.FindByEmailFn != nil {
		return m.FindByEmailFn(email)
	}
	return nil, nil
}
func (m *CustomerRepo) FindByID(id uint) (*model.Customer, error) {
	if m.FindByIDFn != nil {
		return m.FindByIDFn(id)
	}
	return nil, nil
}
func (m *CustomerRepo) UpdatePassword(id uint, hashedPassword string) error {
	if m.UpdatePasswordFn != nil {
		return m.UpdatePasswordFn(id, hashedPassword)
	}
	return nil
}
func (m *CustomerRepo) List() ([]model.Customer, error) {
	if m.ListFn != nil {
		return m.ListFn()
	}
	return nil, nil
}
func (m *CustomerRepo) ListPaged(search string, page, limit int) ([]model.Customer, int64, error) {
	if m.ListPagedFn != nil {
		return m.ListPagedFn(search, page, limit)
	}
	return nil, 0, nil
}
func (m *CustomerRepo) SetResetToken(id uint, token string, expiresAt time.Time) error {
	if m.SetResetTokenFn != nil {
		return m.SetResetTokenFn(id, token, expiresAt)
	}
	return nil
}
func (m *CustomerRepo) FindByResetToken(token string) (*model.Customer, error) {
	if m.FindByResetTokenFn != nil {
		return m.FindByResetTokenFn(token)
	}
	return nil, nil
}
func (m *CustomerRepo) ClearResetToken(id uint) error {
	if m.ClearResetTokenFn != nil {
		return m.ClearResetTokenFn(id)
	}
	return nil
}
func (m *CustomerRepo) SetVerificationToken(id uint, token string) error {
	if m.SetVerificationTokenFn != nil {
		return m.SetVerificationTokenFn(id, token)
	}
	return nil
}
func (m *CustomerRepo) FindByVerificationToken(token string) (*model.Customer, error) {
	if m.FindByVerificationTokenFn != nil {
		return m.FindByVerificationTokenFn(token)
	}
	return nil, nil
}
func (m *CustomerRepo) MarkVerified(id uint) error {
	if m.MarkVerifiedFn != nil {
		return m.MarkVerifiedFn(id)
	}
	return nil
}
func (m *CustomerRepo) FindTokenVersion(id uint) (int, error) {
	if m.FindTokenVersionFn != nil {
		return m.FindTokenVersionFn(id)
	}
	return 0, nil
}
func (m *CustomerRepo) IncrementTokenVersion(id uint) error {
	if m.IncrementTokenVersionFn != nil {
		return m.IncrementTokenVersionFn(id)
	}
	return nil
}
func (m *CustomerRepo) UpdateProfile(id uint, name, phone, address string) error {
	if m.UpdateProfileFn != nil {
		return m.UpdateProfileFn(id, name, phone, address)
	}
	return nil
}
func (m *CustomerRepo) UpdateAvatarURL(id uint, url string) error {
	if m.UpdateAvatarURLFn != nil {
		return m.UpdateAvatarURLFn(id, url)
	}
	return nil
}
func (m *CustomerRepo) FindByGoogleID(googleID string) (*model.Customer, error) {
	if m.FindByGoogleIDFn != nil {
		return m.FindByGoogleIDFn(googleID)
	}
	return nil, nil
}
func (m *CustomerRepo) SetGoogleInfo(id uint, googleID, avatarURL string) error {
	if m.SetGoogleInfoFn != nil {
		return m.SetGoogleInfoFn(id, googleID, avatarURL)
	}
	return nil
}
func (m *CustomerRepo) FindByPhone(phone string) (*model.Customer, error) {
	if m.FindByPhoneFn != nil {
		return m.FindByPhoneFn(phone)
	}
	return nil, nil
}

// ---- SubscriptionRepository ----

type SubscriptionRepo struct {
	CreateFn                func(s *model.Subscription) error
	FindByOwnerIDFn         func(ownerID uint) (*model.Subscription, error)
	FindByOwnerIDsFn        func(ownerIDs []uint) (map[uint]*model.Subscription, error)
	FindByIDFn              func(id uint) (*model.Subscription, error)
	UpdateStatusFn          func(id uint, status model.SubscriptionStatus, notes string) error
	IncrementReservationsFn func(id uint) error
	IncrementCampaignsFn    func(id uint) error
	ResetMonthlyCountFn     func(id uint) error
	FindTrialExpiringFn     func(within time.Duration) ([]model.Subscription, error)
	FindTrialExpiredFn      func() ([]model.Subscription, error)
	FindNeedingResetFn      func() ([]model.Subscription, error)
	FindAllFn               func() ([]model.Subscription, error)
	UpdateFn                func(s *model.Subscription) error
	FindByRestaurantIDFn    func(restaurantID uint) (*model.Subscription, error)
}

func (m *SubscriptionRepo) Create(s *model.Subscription) error {
	if m.CreateFn != nil {
		return m.CreateFn(s)
	}
	return nil
}
func (m *SubscriptionRepo) FindByOwnerID(ownerID uint) (*model.Subscription, error) {
	if m.FindByOwnerIDFn != nil {
		return m.FindByOwnerIDFn(ownerID)
	}
	return nil, nil
}
func (m *SubscriptionRepo) FindByOwnerIDs(ownerIDs []uint) (map[uint]*model.Subscription, error) {
	if m.FindByOwnerIDsFn != nil {
		return m.FindByOwnerIDsFn(ownerIDs)
	}
	return nil, nil
}
func (m *SubscriptionRepo) FindByID(id uint) (*model.Subscription, error) {
	if m.FindByIDFn != nil {
		return m.FindByIDFn(id)
	}
	return nil, nil
}
func (m *SubscriptionRepo) UpdateStatus(id uint, status model.SubscriptionStatus, notes string) error {
	if m.UpdateStatusFn != nil {
		return m.UpdateStatusFn(id, status, notes)
	}
	return nil
}
func (m *SubscriptionRepo) IncrementReservations(id uint) error {
	if m.IncrementReservationsFn != nil {
		return m.IncrementReservationsFn(id)
	}
	return nil
}
func (m *SubscriptionRepo) IncrementCampaigns(id uint) error {
	if m.IncrementCampaignsFn != nil {
		return m.IncrementCampaignsFn(id)
	}
	return nil
}
func (m *SubscriptionRepo) ResetMonthlyCount(id uint) error {
	if m.ResetMonthlyCountFn != nil {
		return m.ResetMonthlyCountFn(id)
	}
	return nil
}
func (m *SubscriptionRepo) FindTrialExpiring(within time.Duration) ([]model.Subscription, error) {
	if m.FindTrialExpiringFn != nil {
		return m.FindTrialExpiringFn(within)
	}
	return nil, nil
}
func (m *SubscriptionRepo) FindTrialExpired() ([]model.Subscription, error) {
	if m.FindTrialExpiredFn != nil {
		return m.FindTrialExpiredFn()
	}
	return nil, nil
}
func (m *SubscriptionRepo) FindNeedingReset() ([]model.Subscription, error) {
	if m.FindNeedingResetFn != nil {
		return m.FindNeedingResetFn()
	}
	return nil, nil
}
func (m *SubscriptionRepo) FindAll() ([]model.Subscription, error) {
	if m.FindAllFn != nil {
		return m.FindAllFn()
	}
	return nil, nil
}
func (m *SubscriptionRepo) Update(s *model.Subscription) error {
	if m.UpdateFn != nil {
		return m.UpdateFn(s)
	}
	return nil
}
func (m *SubscriptionRepo) FindByRestaurantID(restaurantID uint) (*model.Subscription, error) {
	if m.FindByRestaurantIDFn != nil {
		return m.FindByRestaurantIDFn(restaurantID)
	}
	return nil, nil
}

// ---- RestaurantRepository ----

type RestaurantRepo struct {
	CreateFn                      func(r *model.Restaurant) error
	FindAllFn                     func() ([]model.Restaurant, error)
	FindAllActiveFn               func() ([]model.Restaurant, error)
	FindAllActiveWithBranchCoordsFn func() ([]model.RestaurantWithCoords, error)
	FindByIDFn                    func(id uint) (*model.Restaurant, error)
	FindByOwnerIDFn               func(ownerID uint) (*model.Restaurant, error)
	FindByBranchIDFn              func(branchID uint) (*model.Restaurant, error)
	FindByPromoTokenFn            func(token string) (*model.Restaurant, error)
	UpdateFn                      func(r *model.Restaurant) error
	UpdateLogoURLFn               func(id uint, logoURL string) error
	UpdateBannerURLFn             func(id uint, bannerURL string) error
	DeleteFn                      func(id uint) error
}

func (m *RestaurantRepo) Create(r *model.Restaurant) error {
	if m.CreateFn != nil {
		return m.CreateFn(r)
	}
	return nil
}
func (m *RestaurantRepo) FindAll() ([]model.Restaurant, error) {
	if m.FindAllFn != nil {
		return m.FindAllFn()
	}
	return nil, nil
}
func (m *RestaurantRepo) FindAllActive() ([]model.Restaurant, error) {
	if m.FindAllActiveFn != nil {
		return m.FindAllActiveFn()
	}
	return nil, nil
}
func (m *RestaurantRepo) FindAllActiveWithBranchCoords() ([]model.RestaurantWithCoords, error) {
	if m.FindAllActiveWithBranchCoordsFn != nil {
		return m.FindAllActiveWithBranchCoordsFn()
	}
	return nil, nil
}
func (m *RestaurantRepo) FindByID(id uint) (*model.Restaurant, error) {
	if m.FindByIDFn != nil {
		return m.FindByIDFn(id)
	}
	return nil, nil
}
func (m *RestaurantRepo) FindByOwnerID(ownerID uint) (*model.Restaurant, error) {
	if m.FindByOwnerIDFn != nil {
		return m.FindByOwnerIDFn(ownerID)
	}
	return nil, nil
}
func (m *RestaurantRepo) FindByBranchID(branchID uint) (*model.Restaurant, error) {
	if m.FindByBranchIDFn != nil {
		return m.FindByBranchIDFn(branchID)
	}
	return nil, nil
}
func (m *RestaurantRepo) FindByPromoToken(token string) (*model.Restaurant, error) {
	if m.FindByPromoTokenFn != nil {
		return m.FindByPromoTokenFn(token)
	}
	return nil, nil
}
func (m *RestaurantRepo) Update(r *model.Restaurant) error {
	if m.UpdateFn != nil {
		return m.UpdateFn(r)
	}
	return nil
}
func (m *RestaurantRepo) UpdateLogoURL(id uint, logoURL string) error {
	if m.UpdateLogoURLFn != nil {
		return m.UpdateLogoURLFn(id, logoURL)
	}
	return nil
}
func (m *RestaurantRepo) UpdateBannerURL(id uint, bannerURL string) error {
	if m.UpdateBannerURLFn != nil {
		return m.UpdateBannerURLFn(id, bannerURL)
	}
	return nil
}
func (m *RestaurantRepo) Delete(id uint) error {
	if m.DeleteFn != nil {
		return m.DeleteFn(id)
	}
	return nil
}

// ---- StaffRepository ----

type StaffRepo struct {
	CreateFn                func(s *model.Staff) error
	FindByEmailFn           func(email string) (*model.Staff, error)
	FindByIDFn              func(id uint) (*model.Staff, error)
	FindByBranchIDFn        func(branchID uint) ([]model.Staff, error)
	FindByOwnerIDFn         func(ownerID uint) ([]model.Staff, error)
	UpdateFn                func(s *model.Staff) error
	DeleteFn                func(id uint) error
	UpdatePasswordFn        func(id uint, hashedPassword string) error
	SetResetTokenFn         func(id uint, token string, expiresAt time.Time) error
	FindByResetTokenFn      func(token string) (*model.Staff, error)
	ClearResetTokenFn       func(id uint) error
	FindTokenVersionFn      func(id uint) (int, error)
	IncrementTokenVersionFn func(id uint) error
}

func (m *StaffRepo) Create(s *model.Staff) error {
	if m.CreateFn != nil {
		return m.CreateFn(s)
	}
	return nil
}
func (m *StaffRepo) FindByEmail(email string) (*model.Staff, error) {
	if m.FindByEmailFn != nil {
		return m.FindByEmailFn(email)
	}
	return nil, nil
}
func (m *StaffRepo) FindByID(id uint) (*model.Staff, error) {
	if m.FindByIDFn != nil {
		return m.FindByIDFn(id)
	}
	return nil, nil
}
func (m *StaffRepo) FindByBranchID(branchID uint) ([]model.Staff, error) {
	if m.FindByBranchIDFn != nil {
		return m.FindByBranchIDFn(branchID)
	}
	return nil, nil
}
func (m *StaffRepo) FindByOwnerID(ownerID uint) ([]model.Staff, error) {
	if m.FindByOwnerIDFn != nil {
		return m.FindByOwnerIDFn(ownerID)
	}
	return nil, nil
}
func (m *StaffRepo) Update(s *model.Staff) error {
	if m.UpdateFn != nil {
		return m.UpdateFn(s)
	}
	return nil
}
func (m *StaffRepo) Delete(id uint) error {
	if m.DeleteFn != nil {
		return m.DeleteFn(id)
	}
	return nil
}
func (m *StaffRepo) UpdatePassword(id uint, hashedPassword string) error {
	if m.UpdatePasswordFn != nil {
		return m.UpdatePasswordFn(id, hashedPassword)
	}
	return nil
}
func (m *StaffRepo) SetResetToken(id uint, token string, expiresAt time.Time) error {
	if m.SetResetTokenFn != nil {
		return m.SetResetTokenFn(id, token, expiresAt)
	}
	return nil
}
func (m *StaffRepo) FindByResetToken(token string) (*model.Staff, error) {
	if m.FindByResetTokenFn != nil {
		return m.FindByResetTokenFn(token)
	}
	return nil, nil
}
func (m *StaffRepo) ClearResetToken(id uint) error {
	if m.ClearResetTokenFn != nil {
		return m.ClearResetTokenFn(id)
	}
	return nil
}
func (m *StaffRepo) FindTokenVersion(id uint) (int, error) {
	if m.FindTokenVersionFn != nil {
		return m.FindTokenVersionFn(id)
	}
	return 0, nil
}
func (m *StaffRepo) IncrementTokenVersion(id uint) error {
	if m.IncrementTokenVersionFn != nil {
		return m.IncrementTokenVersionFn(id)
	}
	return nil
}

// ---- BusinessOwnerRepository ----

type BusinessOwnerRepo struct {
	CreateFn                  func(o *model.BusinessOwner) error
	FindByEmailFn             func(email string) (*model.BusinessOwner, error)
	FindByIDFn                func(id uint) (*model.BusinessOwner, error)
	UpdatePasswordFn          func(id uint, hashedPassword string) error
	SetVerificationTokenFn    func(id uint, token string) error
	FindByVerificationTokenFn func(token string) (*model.BusinessOwner, error)
	MarkVerifiedFn            func(id uint) error
	SetResetTokenFn           func(id uint, token string, expiresAt time.Time) error
	FindByResetTokenFn        func(token string) (*model.BusinessOwner, error)
	ClearResetTokenFn         func(id uint) error
	ListFn                    func() ([]model.BusinessOwner, error)
	ListPagedFn               func(search string, page, limit int) ([]model.BusinessOwner, int64, error)
	FindByRestaurantIDFn      func(restaurantID uint) (*model.BusinessOwner, error)
	FindTokenVersionFn        func(id uint) (int, error)
	IncrementTokenVersionFn   func(id uint) error
	FindByGoogleIDFn          func(googleID string) (*model.BusinessOwner, error)
	SetGoogleInfoFn           func(id uint, googleID, avatarURL string) error
}

func (m *BusinessOwnerRepo) Create(o *model.BusinessOwner) error {
	if m.CreateFn != nil {
		return m.CreateFn(o)
	}
	return nil
}
func (m *BusinessOwnerRepo) FindByEmail(email string) (*model.BusinessOwner, error) {
	if m.FindByEmailFn != nil {
		return m.FindByEmailFn(email)
	}
	return nil, nil
}
func (m *BusinessOwnerRepo) FindByID(id uint) (*model.BusinessOwner, error) {
	if m.FindByIDFn != nil {
		return m.FindByIDFn(id)
	}
	return nil, nil
}
func (m *BusinessOwnerRepo) UpdatePassword(id uint, hashedPassword string) error {
	if m.UpdatePasswordFn != nil {
		return m.UpdatePasswordFn(id, hashedPassword)
	}
	return nil
}
func (m *BusinessOwnerRepo) SetVerificationToken(id uint, token string) error {
	if m.SetVerificationTokenFn != nil {
		return m.SetVerificationTokenFn(id, token)
	}
	return nil
}
func (m *BusinessOwnerRepo) FindByVerificationToken(token string) (*model.BusinessOwner, error) {
	if m.FindByVerificationTokenFn != nil {
		return m.FindByVerificationTokenFn(token)
	}
	return nil, nil
}
func (m *BusinessOwnerRepo) MarkVerified(id uint) error {
	if m.MarkVerifiedFn != nil {
		return m.MarkVerifiedFn(id)
	}
	return nil
}
func (m *BusinessOwnerRepo) SetResetToken(id uint, token string, expiresAt time.Time) error {
	if m.SetResetTokenFn != nil {
		return m.SetResetTokenFn(id, token, expiresAt)
	}
	return nil
}
func (m *BusinessOwnerRepo) FindByResetToken(token string) (*model.BusinessOwner, error) {
	if m.FindByResetTokenFn != nil {
		return m.FindByResetTokenFn(token)
	}
	return nil, nil
}
func (m *BusinessOwnerRepo) ClearResetToken(id uint) error {
	if m.ClearResetTokenFn != nil {
		return m.ClearResetTokenFn(id)
	}
	return nil
}
func (m *BusinessOwnerRepo) List() ([]model.BusinessOwner, error) {
	if m.ListFn != nil {
		return m.ListFn()
	}
	return nil, nil
}
func (m *BusinessOwnerRepo) ListPaged(search string, page, limit int) ([]model.BusinessOwner, int64, error) {
	if m.ListPagedFn != nil {
		return m.ListPagedFn(search, page, limit)
	}
	return nil, 0, nil
}
func (m *BusinessOwnerRepo) FindByRestaurantID(restaurantID uint) (*model.BusinessOwner, error) {
	if m.FindByRestaurantIDFn != nil {
		return m.FindByRestaurantIDFn(restaurantID)
	}
	return nil, nil
}
func (m *BusinessOwnerRepo) FindTokenVersion(id uint) (int, error) {
	if m.FindTokenVersionFn != nil {
		return m.FindTokenVersionFn(id)
	}
	return 0, nil
}
func (m *BusinessOwnerRepo) IncrementTokenVersion(id uint) error {
	if m.IncrementTokenVersionFn != nil {
		return m.IncrementTokenVersionFn(id)
	}
	return nil
}
func (m *BusinessOwnerRepo) FindByGoogleID(googleID string) (*model.BusinessOwner, error) {
	if m.FindByGoogleIDFn != nil {
		return m.FindByGoogleIDFn(googleID)
	}
	return nil, nil
}
func (m *BusinessOwnerRepo) SetGoogleInfo(id uint, googleID, avatarURL string) error {
	if m.SetGoogleInfoFn != nil {
		return m.SetGoogleInfoFn(id, googleID, avatarURL)
	}
	return nil
}

// ---- EmailConfigRepository ----

type EmailConfigRepo struct {
	GetFn    func() (*model.EmailConfig, error)
	UpsertFn func(cfg *model.EmailConfig) error
}

func (m *EmailConfigRepo) Get() (*model.EmailConfig, error) {
	if m.GetFn != nil {
		return m.GetFn()
	}
	return nil, nil
}
func (m *EmailConfigRepo) Upsert(cfg *model.EmailConfig) error {
	if m.UpsertFn != nil {
		return m.UpsertFn(cfg)
	}
	return nil
}
