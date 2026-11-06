package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MemoRequest struct {
	ID            uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Type          string     `gorm:"type:varchar(50);not null" json:"type"` // 'internal' or 'external'
	BranchID      *uuid.UUID `gorm:"type:uuid" json:"branch_id"`
	Destination   string     `gorm:"type:varchar(255);not null" json:"destination"`
	Subject       string     `gorm:"type:varchar(255);not null" json:"subject"`
	RequesterName string     `gorm:"type:varchar(255);not null" json:"requester_name"`
	RequesterID   *uuid.UUID `gorm:"type:uuid" json:"requester_id"`
	ApproverID    *uuid.UUID `gorm:"type:uuid" json:"approver_id"`
	Status        string     `gorm:"type:varchar(50);not null;default:'PENDING'" json:"status"`
	MemoNumber    *string    `gorm:"type:varchar(100)" json:"memo_number"`
	CreatedAt     time.Time  `gorm:"default:now()" json:"created_at"`
	UpdatedAt     time.Time  `gorm:"default:now()" json:"updated_at"`

	Branch    *Branch   `gorm:"foreignKey:BranchID;constraint:OnDelete:SET NULL;" json:"branch,omitempty"`
	Requester *Employee `gorm:"foreignKey:RequesterID;constraint:OnDelete:SET NULL;" json:"requester,omitempty"`
	Approver  *Employee `gorm:"foreignKey:ApproverID;constraint:OnDelete:SET NULL;" json:"approver,omitempty"`
}

func (m *MemoRequest) BeforeCreate(tx *gorm.DB) (err error) {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return
}
