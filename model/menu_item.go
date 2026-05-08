package model

import "time"

type MenuCategory string

const (
	MenuCategoryMain    MenuCategory = "main"
	MenuCategorySide    MenuCategory = "side"
	MenuCategoryDessert MenuCategory = "dessert"
	MenuCategoryDrink   MenuCategory = "drink"
	MenuCategorySnack   MenuCategory = "snack"
	MenuCategoryOther   MenuCategory = "other"
)

type MenuItem struct {
	ID                   uint         `gorm:"primaryKey;autoIncrement"    json:"id"`
	BranchID             uint         `gorm:"not null;index"              json:"branch_id"`
	Name                 string       `gorm:"type:varchar(100);not null"  json:"name"`
	Description          string       `gorm:"type:text"                   json:"description"`
	Price                int64        `gorm:"not null;default:0"          json:"price"`
	ImageURL             string       `gorm:"type:text"                   json:"image_url"`
	Category             MenuCategory `gorm:"type:varchar(20);not null"   json:"category"`
	IsAvailable          bool         `gorm:"not null;default:true"       json:"is_available"`
	IsFavorite           bool         `gorm:"not null;default:false"      json:"is_favorite"`
	IsChefRecommendation bool         `gorm:"not null;default:false"      json:"is_chef_recommendation"`
	CreatedAt            time.Time    `                                   json:"created_at"`
	UpdatedAt            time.Time    `                                   json:"updated_at"`
	DeletedAt            *time.Time   `gorm:"index"                       json:"-"`

	Branch *Branch `gorm:"foreignKey:BranchID" json:"branch,omitempty"`
}

func (MenuItem) TableName() string { return "tabl_menu_items" }
