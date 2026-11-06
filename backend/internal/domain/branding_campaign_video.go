package domain

import (
	"time"

	"github.com/google/uuid"
)

// BrandingCampaignVideo stores only metadata; the video itself lives in Google Drive.
type BrandingCampaignVideo struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Title       string    `gorm:"type:varchar(200);not null" json:"title"`
	Description string    `gorm:"type:text" json:"description"`
	VideoURL    string    `gorm:"type:text;not null" json:"video_url"`
	ProductName string    `gorm:"type:varchar(200);not null" json:"product_name"`
	PurchaseURL string    `gorm:"type:text;not null" json:"purchase_url"`
	IsActive    bool      `gorm:"not null;default:true" json:"is_active"`
	CreatedAt   time.Time `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt   time.Time `gorm:"not null;default:now()" json:"updated_at"`
}
