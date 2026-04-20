package service

import (
	"errors"
	"log"
	"time"

	"github.com/wayt/wayt-core/internal/model"
	"github.com/wayt/wayt-core/internal/repository"
)

type PlanOrderService interface {
	PlaceOrder(ownerID, planID uint) (*model.PlanOrder, error)
	GetPendingOrder(ownerID uint) (*model.PlanOrder, error)
	ProcessDueOrders() error
}

type planOrderService struct {
	orderRepo   repository.PlanOrderRepository
	planRepo    repository.PlanRepository
	subRepo     repository.SubscriptionRepository
	bookingRepo repository.BookingRepository
	simulDelay  time.Duration
}

func NewPlanOrderService(
	orderRepo repository.PlanOrderRepository,
	planRepo repository.PlanRepository,
	subRepo repository.SubscriptionRepository,
	bookingRepo repository.BookingRepository,
) PlanOrderService {
	return &planOrderService{
		orderRepo:   orderRepo,
		planRepo:    planRepo,
		subRepo:     subRepo,
		bookingRepo: bookingRepo,
		simulDelay:  5 * time.Minute,
	}
}

func (s *planOrderService) PlaceOrder(ownerID, planID uint) (*model.PlanOrder, error) {
	// Validate plan exists and is active
	plan, err := s.planRepo.FindByID(planID)
	if err != nil || !plan.IsActive {
		return nil, errors.New("paket tidak ditemukan atau tidak aktif")
	}

	// Cancel any existing pending order for this owner
	if existing, err := s.orderRepo.FindPendingByOwnerID(ownerID); err == nil {
		_ = s.orderRepo.UpdateStatus(existing.ID, model.PlanOrderStatusFailed)
	}

	o := &model.PlanOrder{
		BusinessOwnerID: ownerID,
		PlanID:          planID,
		Status:          model.PlanOrderStatusPending,
		ProcessAt:       time.Now().Add(s.simulDelay),
	}
	if err := s.orderRepo.Create(o); err != nil {
		return nil, err
	}
	o.Plan = plan
	return o, nil
}

func (s *planOrderService) GetPendingOrder(ownerID uint) (*model.PlanOrder, error) {
	return s.orderRepo.FindPendingByOwnerID(ownerID)
}

// ProcessDueOrders is called by the background job.
func (s *planOrderService) ProcessDueOrders() error {
	orders, err := s.orderRepo.FindDue()
	if err != nil {
		return err
	}
	for _, o := range orders {
		if err := s.activatePlan(o.BusinessOwnerID, o.PlanID); err != nil {
			log.Printf("[PLAN ORDER] failed to activate plan for owner %d: %v", o.BusinessOwnerID, err)
			_ = s.orderRepo.UpdateStatus(o.ID, model.PlanOrderStatusFailed)
			continue
		}
		_ = s.orderRepo.UpdateStatus(o.ID, model.PlanOrderStatusCompleted)
		log.Printf("[PLAN ORDER] order #%d completed: owner %d → plan %d", o.ID, o.BusinessOwnerID, o.PlanID)
	}
	return nil
}

func (s *planOrderService) activatePlan(ownerID, planID uint) error {
	sub, err := s.subRepo.FindByOwnerID(ownerID)
	if err != nil {
		return err
	}
	plan, err := s.planRepo.FindByID(planID)
	if err != nil {
		return err
	}
	now := time.Now()
	sub.PlanID = planID
	sub.Status = model.SubscriptionStatusActive
	sub.ActivatedAt = &now
	if err := s.subRepo.Update(sub); err != nil {
		return err
	}
	applyOverLimitFlags(s.bookingRepo, ownerID, plan.MaxReservationsPerMonth, sub.ReservationsThisMonth)
	return nil
}

// applyOverLimitFlags clears all is_over_limit flags then, for a downgrade,
// re-marks the active current-month bookings that exceed the new limit.
func applyOverLimitFlags(bookingRepo interface {
	ClearOverLimitByOwner(uint) error
	MarkOverLimitByOwner(uint, int) error
}, ownerID uint, maxPerMonth, reservationsThisMonth int) {
	_ = bookingRepo.ClearOverLimitByOwner(ownerID)
	if maxPerMonth != -1 && reservationsThisMonth > maxPerMonth {
		_ = bookingRepo.MarkOverLimitByOwner(ownerID, maxPerMonth)
	}
}
