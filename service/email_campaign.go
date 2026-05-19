package service

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/wayt-app/wayt-core/model"
	"github.com/wayt-app/wayt-core/repository"
	"github.com/wayt-app/wayt-core/pkg/email"
)

type EmailCampaignService interface {
	Create(ownerID uint, subject, body string, filterBranchID *uint, filterSegment string, scheduledAt *time.Time) (*model.EmailCampaign, error)
	List(restaurantID uint) ([]model.EmailCampaign, error)
	GetByID(id, restaurantID uint) (*model.EmailCampaign, error)
	Cancel(id, restaurantID uint) error
	PreviewCount(restaurantID uint, branchID *uint, segment string) (int64, error)
	ProcessScheduled() error
}

type emailCampaignService struct {
	repo           repository.EmailCampaignRepository
	subRepo        repository.SubscriptionRepository
	restaurantRepo repository.RestaurantRepository
	emailConfigRepo repository.EmailConfigRepository
	emailSender    email.Sender
}

func NewEmailCampaignService(
	repo repository.EmailCampaignRepository,
	subRepo repository.SubscriptionRepository,
	restaurantRepo repository.RestaurantRepository,
	emailConfigRepo repository.EmailConfigRepository,
	emailSender email.Sender,
) EmailCampaignService {
	return &emailCampaignService{
		repo:            repo,
		subRepo:         subRepo,
		restaurantRepo:  restaurantRepo,
		emailConfigRepo: emailConfigRepo,
		emailSender:     emailSender,
	}
}

func (s *emailCampaignService) Create(ownerID uint, subject, body string, filterBranchID *uint, filterSegment string, scheduledAt *time.Time) (*model.EmailCampaign, error) {
	if subject == "" {
		return nil, errors.New("subject email wajib diisi")
	}
	if body == "" {
		return nil, errors.New("isi email wajib diisi")
	}
	if filterSegment == "" {
		filterSegment = "all"
	}

	// Look up restaurant for this owner
	rest, err := s.restaurantRepo.FindByOwnerID(ownerID)
	if err != nil || rest == nil {
		return nil, errors.New("restoran tidak ditemukan")
	}

	// Quota check
	sub, subErr := s.subRepo.FindByRestaurantID(rest.ID)
	if subErr == nil && sub != nil {
		if sub.Status != model.SubscriptionStatusActive && sub.Status != model.SubscriptionStatusTrial {
			return nil, errors.New("restoran tidak memiliki langganan aktif")
		}
		if sub.Plan != nil && sub.Plan.MaxCampaignsPerMonth != -1 {
			if sub.Plan.MaxCampaignsPerMonth == 0 {
				return nil, errors.New("paket Anda tidak mendukung fitur blast email")
			}
			if sub.CampaignsThisMonth >= sub.Plan.MaxCampaignsPerMonth {
				return nil, fmt.Errorf("kuota kampanye bulan ini sudah habis (%d/%d)", sub.CampaignsThisMonth, sub.Plan.MaxCampaignsPerMonth)
			}
		}
	}

	// Default: send now
	if scheduledAt == nil {
		now := time.Now()
		scheduledAt = &now
	}

	c := &model.EmailCampaign{
		RestaurantID:   rest.ID,
		Subject:        subject,
		Body:           body,
		FilterBranchID: filterBranchID,
		FilterSegment:  filterSegment,
		Status:         model.CampaignStatusScheduled,
		ScheduledAt:    scheduledAt,
	}
	if err := s.repo.Create(c); err != nil {
		return nil, err
	}
	// Increment quota counter
	if sub != nil {
		_ = s.subRepo.IncrementCampaigns(sub.ID)
	}
	return c, nil
}

func (s *emailCampaignService) List(restaurantID uint) ([]model.EmailCampaign, error) {
	return s.repo.FindByRestaurant(restaurantID)
}

func (s *emailCampaignService) GetByID(id, restaurantID uint) (*model.EmailCampaign, error) {
	c, err := s.repo.FindByID(id)
	if err != nil || c.RestaurantID != restaurantID {
		return nil, errors.New("kampanye tidak ditemukan")
	}
	return c, nil
}

func (s *emailCampaignService) Cancel(id, restaurantID uint) error {
	c, err := s.repo.FindByID(id)
	if err != nil || c.RestaurantID != restaurantID {
		return errors.New("kampanye tidak ditemukan")
	}
	if c.Status != model.CampaignStatusScheduled {
		return errors.New("hanya kampanye terjadwal yang bisa dibatalkan")
	}
	return s.repo.UpdateStatus(id, model.CampaignStatusCancelled)
}

func (s *emailCampaignService) PreviewCount(restaurantID uint, branchID *uint, segment string) (int64, error) {
	if segment == "" {
		segment = "all"
	}
	recipients, err := s.repo.FindRecipients(restaurantID, branchID, segment)
	return int64(len(recipients)), err
}

// ProcessScheduled is called by the background worker every minute.
func (s *emailCampaignService) ProcessScheduled() error {
	due, err := s.repo.FindDue()
	if err != nil || len(due) == 0 {
		return err
	}
	for _, c := range due {
		c := c
		go s.sendCampaign(&c)
	}
	return nil
}

func (s *emailCampaignService) sendCampaign(c *model.EmailCampaign) {
	if err := s.repo.UpdateStatus(c.ID, model.CampaignStatusSending); err != nil {
		return
	}

	recipients, err := s.repo.FindRecipients(c.RestaurantID, c.FilterBranchID, c.FilterSegment)
	if err != nil {
		_ = s.repo.UpdateStatus(c.ID, model.CampaignStatusFailed)
		return
	}
	if len(recipients) == 0 {
		_ = s.repo.UpdateSentStats(c.ID, 0, 0, 0, time.Now())
		return
	}

	rest, _ := s.restaurantRepo.FindByID(c.RestaurantID)
	restName := ""
	restLogoURL := ""
	if rest != nil {
		restName = rest.Name
		restLogoURL = rest.LogoURL
	}

	var emailCfg *model.EmailConfig
	if s.emailConfigRepo != nil {
		emailCfg, _ = s.emailConfigRepo.Get()
	}

	successCount, failCount := 0, 0
	for _, r := range recipients {
		if r.Email == "" {
			continue
		}
		wrapped := wrapEmailHTML(c.Body, emailCfg, r.Name, restName, restLogoURL)
		if err := s.emailSender.Send(r.Email, c.Subject, wrapped); err != nil {
			log.Printf("[CAMPAIGN %d] gagal kirim ke %s: %v", c.ID, r.Email, err)
			failCount++
		} else {
			successCount++
		}
	}

	_ = s.repo.UpdateSentStats(c.ID, len(recipients), successCount, failCount, time.Now())
}
