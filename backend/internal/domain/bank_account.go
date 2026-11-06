package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var ErrBankAccountInUse = errors.New("rekening bank sudah digunakan dalam transaksi")

// CompanyBankAccount menyimpan data rekening bank perusahaan untuk transaksi ke pelanggan.
type CompanyBankAccount struct {
	ID                uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CompanyID         *uuid.UUID `gorm:"type:uuid" json:"company_id"`
	BranchID          *uuid.UUID `gorm:"type:uuid" json:"branch_id"`
	BankName          string     `gorm:"type:varchar(100);not null" json:"bank_name"`
	AccountNumber     string     `gorm:"type:varchar(50);not null" json:"account_number"`
	AccountHolder     string     `gorm:"type:varchar(200);not null" json:"account_holder"`
	BankLogoURL       string     `gorm:"type:text" json:"bank_logo_url"`
	IsActive          bool       `gorm:"default:false" json:"is_active"`
	Notes             string     `gorm:"type:text" json:"notes"`
	ManualBalance     float64    `gorm:"type:decimal(15,2);default:0" json:"manual_balance"`
	LastBalanceUpdate *time.Time `json:"last_balance_update"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`

	Company *Company `gorm:"foreignKey:CompanyID" json:"company,omitempty"`
	Branch  *Branch  `gorm:"foreignKey:BranchID" json:"branch,omitempty"`
}

func (b *CompanyBankAccount) BeforeCreate(tx *gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return nil
}
