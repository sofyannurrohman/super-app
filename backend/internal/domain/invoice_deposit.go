package domain

import (
	"time"

	"github.com/google/uuid"
)

// InvoiceDeposit is a cashier's claim that money for the listed invoices was
// deposited to a registered factory bank account. Central cashier verifies it.
type InvoiceDeposit struct {
	ID            uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	BranchID      uuid.UUID  `gorm:"type:uuid;not null;index" json:"branch_id"`
	CompanyID     uuid.UUID  `gorm:"type:uuid;not null;index" json:"company_id"`
	BankAccountID uuid.UUID  `gorm:"type:uuid;not null;index" json:"bank_account_id"`
	DepositDate   time.Time  `gorm:"type:date;not null" json:"deposit_date"`
	Amount        float64    `gorm:"type:decimal(15,2);not null" json:"amount"`
	ProofURL      string     `gorm:"type:text" json:"proof_url"`
	Notes         string     `gorm:"type:text" json:"notes"`
	Status        string     `gorm:"type:varchar(20);not null;default:'PENDING';index" json:"status"`
	SubmittedByID uuid.UUID  `gorm:"type:uuid;not null" json:"submitted_by_id"`
	ReviewedByID  *uuid.UUID `gorm:"type:uuid" json:"reviewed_by_id,omitempty"`
	ReviewedAt    *time.Time `json:"reviewed_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`

	Branch      *Branch              `gorm:"foreignKey:BranchID" json:"branch,omitempty"`
	Company     *Company             `gorm:"foreignKey:CompanyID" json:"company,omitempty"`
	BankAccount *CompanyBankAccount  `gorm:"foreignKey:BankAccountID" json:"bank_account,omitempty"`
	Items       []InvoiceDepositItem `gorm:"foreignKey:DepositID" json:"items"`
}

type InvoiceDepositItem struct {
	ID                 uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	DepositID          uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_invoice_deposit_item" json:"deposit_id"`
	SalesTransactionID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_invoice_deposit_item;index" json:"sales_transaction_id"`
	Amount             float64   `gorm:"type:decimal(15,2);not null" json:"amount"`
	CreatedAt          time.Time `json:"created_at"`

	SalesTransaction *SalesTransaction `gorm:"foreignKey:SalesTransactionID" json:"sales_transaction,omitempty"`
}
