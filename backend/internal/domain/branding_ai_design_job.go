package domain

import (
	"time"

	"github.com/google/uuid"
)

// BrandingAIDesignJob tracks a draft from Hermes through the Windows CorelDRAW bridge.
// The generated CDR remains a draft; the editor still uploads the approved export through
// the existing branding validation flow.
type BrandingAIDesignJob struct {
	ID                uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	BannerOrderID     uuid.UUID  `gorm:"type:uuid;not null;index" json:"banner_order_id"`
	RequestedBy       *uuid.UUID `gorm:"type:uuid" json:"requested_by,omitempty"`
	Status            string     `gorm:"type:varchar(40);not null;default:'queued_hermes';index" json:"status"`
	ReferenceBannerID *uuid.UUID `gorm:"type:uuid" json:"reference_banner_id,omitempty"`
	ReferenceImageURL *string    `gorm:"type:text" json:"reference_image_url,omitempty"`
	Prompt            string     `gorm:"type:text;not null" json:"prompt"`
	SVGURL            *string    `gorm:"type:text" json:"svg_url,omitempty"`
	PreviewURL        *string    `gorm:"type:text" json:"preview_url,omitempty"`
	CDRURL            *string    `gorm:"type:text" json:"cdr_url,omitempty"`
	CDRFileName       *string    `gorm:"type:varchar(500)" json:"cdr_file_name,omitempty"`
	ErrorMessage      *string    `gorm:"type:text" json:"error_message,omitempty"`
	ClaimedAt         *time.Time `json:"claimed_at,omitempty"`
	CompletedAt       *time.Time `json:"completed_at,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

func (BrandingAIDesignJob) TableName() string { return "branding_ai_design_jobs" }
