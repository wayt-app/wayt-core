package repository

import (
	"time"

	"github.com/wayt-app/wayt-core/model"
	"gorm.io/gorm"
)

// CampaignRecipient is a lightweight struct for email sending.
type CampaignRecipient struct {
	CustomerID uint   `gorm:"column:customer_id"`
	Name       string `gorm:"column:name"`
	Email      string `gorm:"column:email"`
}

type EmailCampaignRepository interface {
	Create(c *model.EmailCampaign) error
	FindByID(id uint) (*model.EmailCampaign, error)
	FindByRestaurant(restaurantID uint) ([]model.EmailCampaign, error)
	UpdateStatus(id uint, status model.CampaignStatus) error
	UpdateSentStats(id uint, recipientCount, successCount, failCount int, sentAt time.Time) error
	FindDue() ([]model.EmailCampaign, error)
	Delete(id uint) error
	// FindRecipients returns unique customers for a restaurant matching the given branch + segment filters.
	FindRecipients(restaurantID uint, branchID *uint, segment string) ([]CampaignRecipient, error)
}

type emailCampaignRepository struct{ db *gorm.DB }

func NewEmailCampaignRepository(db *gorm.DB) EmailCampaignRepository {
	return &emailCampaignRepository{db: db}
}

func (r *emailCampaignRepository) Create(c *model.EmailCampaign) error {
	return r.db.Create(c).Error
}

func (r *emailCampaignRepository) FindByID(id uint) (*model.EmailCampaign, error) {
	var c model.EmailCampaign
	return &c, r.db.Where("id = ?", id).First(&c).Error
}

func (r *emailCampaignRepository) FindByRestaurant(restaurantID uint) ([]model.EmailCampaign, error) {
	var list []model.EmailCampaign
	return list, r.db.Where("restaurant_id = ?", restaurantID).
		Order("created_at DESC").Find(&list).Error
}

func (r *emailCampaignRepository) UpdateStatus(id uint, status model.CampaignStatus) error {
	return r.db.Model(&model.EmailCampaign{}).Where("id = ?", id).
		Update("status", status).Error
}

func (r *emailCampaignRepository) UpdateSentStats(id uint, recipientCount, successCount, failCount int, sentAt time.Time) error {
	return r.db.Model(&model.EmailCampaign{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":          model.CampaignStatusSent,
			"recipient_count": recipientCount,
			"success_count":   successCount,
			"fail_count":      failCount,
			"sent_at":         sentAt,
		}).Error
}

// FindDue returns scheduled campaigns whose scheduled_at is in the past.
func (r *emailCampaignRepository) FindDue() ([]model.EmailCampaign, error) {
	var list []model.EmailCampaign
	return list, r.db.Where("status = ? AND scheduled_at <= ?", model.CampaignStatusScheduled, time.Now()).
		Find(&list).Error
}

func (r *emailCampaignRepository) Delete(id uint) error {
	return r.db.Where("id = ?", id).Delete(&model.EmailCampaign{}).Error
}

func (r *emailCampaignRepository) FindRecipients(restaurantID uint, branchID *uint, segment string) ([]CampaignRecipient, error) {
	var rows []CampaignRecipient

	q := `
		SELECT c.id AS customer_id, c.name, c.email
		FROM tabl_customers c
		INNER JOIN tabl_bookings b   ON b.customer_id = c.id
		INNER JOIN tabl_branches br  ON b.branch_id = br.id AND br.deleted_at IS NULL
		INNER JOIN tabl_restaurants r ON br.restaurant_id = r.id AND r.deleted_at IS NULL
		WHERE r.id = ?
		  AND b.status = 'completed'
		  AND c.email != ''`

	args := []interface{}{restaurantID}

	if branchID != nil {
		q += " AND br.id = ?"
		args = append(args, *branchID)
	}

	q += " GROUP BY c.id, c.name, c.email"

	switch segment {
	case "new":
		q += " HAVING COUNT(b.id) = 1"
	case "regular":
		q += " HAVING COUNT(b.id) BETWEEN 2 AND 4"
	case "loyal":
		q += " HAVING COUNT(b.id) >= 5"
	}

	return rows, r.db.Raw(q, args...).Scan(&rows).Error
}
