package domain

import (
	"time"

	"github.com/google/uuid"
)

type CashierFinancialReport struct {
	ID         uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	BranchID   uuid.UUID  `gorm:"type:uuid;not null" json:"branch_id"`
	CompanyID  *uuid.UUID `gorm:"type:uuid" json:"company_id,omitempty"`
	ReportDate time.Time  `gorm:"type:date;not null" json:"report_date"`

	TotalCashSales     float64 `gorm:"type:decimal(15,2);default:0" json:"total_cash_sales"`
	TotalTransferSales float64 `gorm:"type:decimal(15,2);default:0" json:"total_transfer_sales"`

	// Petty Cash (Kas Kecil) Allocation
	AllocatedToPettyCash float64 `gorm:"type:decimal(15,2);default:0" json:"allocated_to_petty_cash"`

	// Bank / Kas Pabrik Deposit Info
	DepositedAmount        float64    `gorm:"type:decimal(15,2);default:0" json:"deposited_amount"`
	DepositDestinationType string     `gorm:"type:varchar(30);default:'BRANCH_BANK'" json:"deposit_destination_type"` // BRANCH_BANK, FACTORY_KAS, OTHER
	TargetBankAccountID    *uuid.UUID `gorm:"type:uuid" json:"target_bank_account_id,omitempty"`
	TargetKasPabrikID      *uuid.UUID `gorm:"type:uuid" json:"target_kas_pabrik_id,omitempty"`
	TargetBankName         string     `gorm:"type:varchar(100)" json:"target_bank_name"`
	TargetAccountNumber    string     `gorm:"type:varchar(50)" json:"target_account_number"`
	TargetAccountName      string     `gorm:"type:varchar(200)" json:"target_account_name"`
	DepositProofURL        string     `gorm:"type:text" json:"deposit_proof_url"`

	Status string `gorm:"type:varchar(20);default:'DRAFT'" json:"status"` // DRAFT, DEPOSITED
	Notes  string `gorm:"type:text" json:"notes"`

	CreatedAt time.Time `gorm:"default:now()" json:"created_at"`
	UpdatedAt time.Time `gorm:"default:now()" json:"updated_at"`

	Branch            *Branch             `gorm:"foreignKey:BranchID" json:"branch,omitempty"`
	Company           *Company            `gorm:"foreignKey:CompanyID" json:"company,omitempty"`
	TargetBankAccount *CompanyBankAccount `gorm:"foreignKey:TargetBankAccountID" json:"target_bank_account,omitempty"`
	TargetKasPabrik   *KasPabrik          `gorm:"foreignKey:TargetKasPabrikID" json:"target_kas_pabrik,omitempty"`
}
