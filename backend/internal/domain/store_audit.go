package domain

import (
	"time"

	"github.com/google/uuid"
)

type StoreAudit struct {
	ID            uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	StoreID       uuid.UUID `gorm:"type:uuid;not null" json:"store_id"`
	AuditorID     uuid.UUID `gorm:"type:uuid;not null" json:"auditor_id"`
	Status        string    `gorm:"type:varchar(20);not null" json:"status"` // VERIFIED_REAL, FICTITIOUS
	Notes         string    `gorm:"type:text" json:"notes"`
	PhotoProofURL string    `gorm:"type:text" json:"photo_proof_url"`
	AuditDate     time.Time `gorm:"default:now()" json:"audit_date"`
	CreatedAt     time.Time `gorm:"default:now()" json:"created_at"`
	UpdatedAt     time.Time `gorm:"default:now()" json:"updated_at"`

	Store   *Store    `gorm:"foreignKey:StoreID" json:"store,omitempty"`
	Auditor *Employee `gorm:"foreignKey:AuditorID" json:"auditor,omitempty"`
}

func (s *StoreAudit) TableName() string {
	return "store_audits"
}
