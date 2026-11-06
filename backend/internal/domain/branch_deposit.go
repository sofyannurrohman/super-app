package domain

import (
	"time"

	"github.com/google/uuid"
)

type BranchDepositStatus string

const (
	DepositStatusPending  BranchDepositStatus = "PENDING"
	DepositStatusApproved BranchDepositStatus = "APPROVED"
	DepositStatusRejected BranchDepositStatus = "REJECTED"
)

type PaymentMethod string

const (
	PaymentMethodCash     PaymentMethod = "CASH"
	PaymentMethodTransfer PaymentMethod = "TRANSFER"
)

type BranchDeposit struct {
	ID               uuid.UUID           `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CompanyID        uuid.UUID           `gorm:"type:uuid;not null" json:"company_id"`
	BranchID         uuid.UUID           `gorm:"type:uuid;not null" json:"branch_id"`
	DepositDate      time.Time           `gorm:"type:date;not null" json:"deposit_date"`
	Amount           float64             `gorm:"type:decimal(20,2);not null" json:"amount"`
	SystemAmount     float64             `gorm:"type:decimal(20,2);default:0" json:"system_amount"`
	PaymentMethod    PaymentMethod       `gorm:"type:varchar(20);not null" json:"payment_method"`
	BankAccountID    *uuid.UUID          `gorm:"type:uuid" json:"bank_account_id"` // Jika metode TRANSFER
	ProofURL         string              `gorm:"type:text" json:"proof_url"`
	Notes            string              `gorm:"type:text" json:"notes"`
	DiscrepancyNotes string              `gorm:"type:text" json:"discrepancy_notes"`
	Status           BranchDepositStatus `gorm:"type:varchar(20);default:'PENDING'" json:"status"`

	SubmittedBy    uuid.UUID  `gorm:"type:uuid;not null" json:"submitted_by"`
	ApprovedBy     *uuid.UUID `gorm:"type:uuid" json:"approved_by"`
	JournalEntryID *uuid.UUID `gorm:"type:uuid" json:"journal_entry_id"` // Ref ke JournalEntry setelah approve

	CreatedAt time.Time `gorm:"default:now()" json:"created_at"`
	UpdatedAt time.Time `gorm:"default:now()" json:"updated_at"`

	Branch       *Branch             `gorm:"foreignKey:BranchID" json:"branch,omitempty"`
	BankAccount  *CompanyBankAccount `gorm:"foreignKey:BankAccountID" json:"bank_account,omitempty"`
	JournalEntry *JournalEntry       `gorm:"foreignKey:JournalEntryID" json:"journal_entry,omitempty"`
	Submitter    *User               `gorm:"foreignKey:SubmittedBy" json:"submitter,omitempty"`
	Approver     *User               `gorm:"foreignKey:ApprovedBy" json:"approver,omitempty"`
}

func (b *BranchDeposit) TableName() string { return "branch_deposits" }
