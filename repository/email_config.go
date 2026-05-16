package repository

import (
	"github.com/wayt-app/wayt-core/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type EmailConfigRepository interface {
	Get() (*model.EmailConfig, error)
	Upsert(cfg *model.EmailConfig) error
}

type emailConfigRepository struct{ db *gorm.DB }

func NewEmailConfigRepository(db *gorm.DB) EmailConfigRepository {
	return &emailConfigRepository{db: db}
}

func (r *emailConfigRepository) Get() (*model.EmailConfig, error) {
	var cfg model.EmailConfig
	err := r.db.FirstOrCreate(&cfg, model.EmailConfig{ID: 1}).Error
	return &cfg, err
}

func (r *emailConfigRepository) Upsert(cfg *model.EmailConfig) error {
	cfg.ID = 1
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"header_image_url", "logo_url",
			"instagram_url", "facebook_url", "tiktok_url",
			"website_url", "support_email",
			"footer_note", "copyright", "updated_at",
		}),
	}).Create(cfg).Error
}
