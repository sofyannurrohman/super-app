package domain

import (
	"time"

	"github.com/google/uuid"
)

type CashBankMutation struct {
	ID                   uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	BranchID             uuid.UUID  `gorm:"type:uuid;not null" json:"branch_id"`
	CompanyID            *uuid.UUID `gorm:"type:uuid" json:"company_id,omitempty"`
	DestinationCompanyID *uuid.UUID `gorm:"type:uuid;index" json:"destination_company_id,omitempty"`
	MutationType         string     `gorm:"type:varchar(20);not null" json:"mutation_type"` // IN, OUT
	FundType             string     `gorm:"type:varchar(20);not null" json:"fund_type"`     // CASH, PETTY_CASH, BANK
	BankAccountID        *uuid.UUID `gorm:"type:uuid" json:"bank_account_id,omitempty"`     // null if FundType != BANK

	Amount       float64   `gorm:"type:decimal(15,2);not null;default:0" json:"amount"`
	Notes        string    `gorm:"type:text" json:"notes"`
	ProofURL     string    `gorm:"type:text" json:"proof_url"`
	FundOrigin   string    `gorm:"type:varchar(40)" json:"fund_origin,omitempty"`
	MutationDate time.Time `gorm:"type:date;not null" json:"mutation_date"`

	TargetKasPabrikID *uuid.UUID `gorm:"type:uuid" json:"target_kas_pabrik_id,omitempty"`
	StatusSetoran     string     `gorm:"type:varchar(20);default:'PENDING'" json:"status_setoran,omitempty"`
	StatusSetoranNote string     `gorm:"type:text" json:"status_setoran_note,omitempty"`
	StatusSetoranBy   *uuid.UUID `gorm:"type:uuid" json:"status_setoran_by,omitempty"`
	StatusSetoranAt   *time.Time `json:"status_setoran_at,omitempty"`

	Category        string                    `gorm:"type:varchar(50)" json:"category,omitempty"` // SALES_DEPOSIT, PETTY_CASH_TOPUP, OPERATIONAL_EXPENSE, REIMBURSEMENT, OTHER
	ReimbursementID *uuid.UUID                `gorm:"type:uuid" json:"reimbursement_id,omitempty"`
	CashierReportID *uuid.UUID                `gorm:"type:uuid;index" json:"cashier_report_id,omitempty"`
	InvoiceItems    []CashBankMutationInvoice `gorm:"foreignKey:MutationID" json:"invoice_items,omitempty"`

	CreatedAt time.Time `gorm:"default:now()" json:"created_at"`
	UpdatedAt time.Time `gorm:"default:now()" json:"updated_at"`

	Branch             *Branch                 `gorm:"foreignKey:BranchID" json:"branch,omitempty"`
	Company            *Company                `gorm:"foreignKey:CompanyID" json:"company,omitempty"`
	DestinationCompany *Company                `gorm:"foreignKey:DestinationCompanyID" json:"destination_company,omitempty"`
	BankAccount        *CompanyBankAccount     `gorm:"foreignKey:BankAccountID" json:"bank_account,omitempty"`
	TargetKas          *KasPabrik              `gorm:"foreignKey:TargetKasPabrikID" json:"target_kas,omitempty"`
	Reimbursement      *Reimbursement          `gorm:"foreignKey:ReimbursementID" json:"reimbursement,omitempty"`
	CashierReport      *CashierFinancialReport `gorm:"foreignKey:CashierReportID" json:"cashier_report,omitempty"`
}

// CashBankMutationInvoice allocates an approved customer cash payment to a
// branch-to-factory bank deposit mutation.
type CashBankMutationInvoice struct {
	ID                 uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	MutationID         uuid.UUID `gorm:"type:uuid;not null;index" json:"mutation_id"`
	SalesPaymentID     uuid.UUID `gorm:"type:uuid;not null;index" json:"sales_payment_id"`
	SalesTransactionID uuid.UUID `gorm:"type:uuid;not null;index" json:"sales_transaction_id"`
	Amount             float64   `gorm:"type:decimal(15,2);not null" json:"amount"`
	CreatedAt          time.Time `json:"created_at"`

	SalesPayment     *SalesPayment     `gorm:"foreignKey:SalesPaymentID" json:"sales_payment,omitempty"`
	SalesTransaction *SalesTransaction `gorm:"foreignKey:SalesTransactionID" json:"sales_transaction,omitempty"`
}
