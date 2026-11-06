package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BranchCompanyBalance melacak saldo Kas Besar dan Kas Kecil secara terpisah per Cabang dan per Perusahaan (PT).
type BranchCompanyBalance struct {
	ID                uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	BranchID          uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex:idx_branch_company" json:"branch_id"`
	CompanyID         uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex:idx_branch_company" json:"company_id"`
	ManualCashBalance float64    `gorm:"type:decimal(15,2);default:0" json:"manual_cash_balance"` // Kas Besar per PT
	PettyCashBalance  float64    `gorm:"type:decimal(15,2);default:0" json:"petty_cash_balance"`  // Kas Kecil per PT
	CreatedAt         time.Time  `gorm:"default:now()" json:"created_at"`
	UpdatedAt         time.Time  `gorm:"default:now()" json:"updated_at"`

	Branch  *Branch  `gorm:"foreignKey:BranchID" json:"branch,omitempty"`
	Company *Company `gorm:"foreignKey:CompanyID" json:"company,omitempty"`
}

func (b *BranchCompanyBalance) BeforeCreate(tx *gorm.DB) (err error) {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return
}
