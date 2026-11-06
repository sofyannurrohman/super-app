package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// =============================================================================
// Chart of Accounts — Bagan Akun (PSAK)
// Tipe akun mengikuti klasifikasi PSAK:
//   ASSET (Aset), LIABILITY (Kewajiban), EQUITY (Ekuitas),
//   REVENUE (Pendapatan), EXPENSE (Beban)
// =============================================================================

type Account struct {
	ID             uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CompanyID      *uuid.UUID `gorm:"type:uuid;index" json:"company_id"`
	Code           string     `gorm:"type:varchar(20);not null" json:"code"`
	Name           string     `gorm:"type:varchar(200);not null" json:"name"`
	AccountType    string     `gorm:"type:varchar(30);not null" json:"account_type"`   // ASSET, LIABILITY, EQUITY, REVENUE, EXPENSE
	AccountSubtype string     `gorm:"type:varchar(50)" json:"account_subtype"`         // e.g. CURRENT_ASSET, FIXED_ASSET, COGS
	NormalBalance  string     `gorm:"type:varchar(10);not null" json:"normal_balance"` // DEBIT atau CREDIT
	ParentID       *uuid.UUID `gorm:"type:uuid" json:"parent_id"`
	IsActive       bool       `gorm:"default:true" json:"is_active"`
	IsSystem       bool       `gorm:"default:false" json:"is_system"` // akun sistem tidak bisa dihapus user
	CreatedAt      time.Time  `gorm:"default:now()" json:"created_at"`

	Company  *Company  `gorm:"foreignKey:CompanyID" json:"company,omitempty"`
	Parent   *Account  `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
	Children []Account `gorm:"foreignKey:ParentID" json:"children,omitempty"`
}

func (a *Account) TableName() string { return "accounts" }

func (a *Account) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}

// =============================================================================
// Journal Entry — Jurnal Umum (double-entry)
// =============================================================================

type JournalEntry struct {
	ID            uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CompanyID     uuid.UUID  `gorm:"type:uuid;not null;index" json:"company_id"`
	EntryNumber   string     `gorm:"type:varchar(30);unique" json:"entry_number"` // e.g. JRN-2025-08-0001
	EntryDate     time.Time  `gorm:"type:date;not null" json:"entry_date"`
	Description   string     `gorm:"type:text" json:"description"`
	ReferenceType string     `gorm:"type:varchar(50)" json:"reference_type"` // SALES_TRANSACTION, PAYROLL_RUN, REIMBURSEMENT, MANUAL
	ReferenceID   *uuid.UUID `gorm:"type:uuid;index" json:"reference_id"`
	Status        string     `gorm:"type:varchar(20);default:'POSTED'" json:"status"` // DRAFT, POSTED, VOID
	CreatedBy     *uuid.UUID `gorm:"type:uuid" json:"created_by"`
	CreatedAt     time.Time  `gorm:"default:now()" json:"created_at"`
	UpdatedAt     time.Time  `gorm:"default:now()" json:"updated_at"`

	Company *Company           `gorm:"foreignKey:CompanyID" json:"company,omitempty"`
	Lines   []JournalEntryLine `gorm:"foreignKey:JournalEntryID;constraint:OnDelete:CASCADE" json:"lines,omitempty"`
}

func (j *JournalEntry) TableName() string { return "journal_entries" }

func (j *JournalEntry) BeforeCreate(tx *gorm.DB) error {
	if j.ID == uuid.Nil {
		j.ID = uuid.New()
	}
	return nil
}

// =============================================================================
// Journal Entry Line — Baris Debit / Kredit
// Constraint: hanya salah satu dari debit_amount atau credit_amount yang > 0
// =============================================================================

type JournalEntryLine struct {
	ID             uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	JournalEntryID uuid.UUID `gorm:"type:uuid;not null;index" json:"journal_entry_id"`
	AccountID      uuid.UUID `gorm:"type:uuid;not null" json:"account_id"`
	DebitAmount    float64   `gorm:"type:decimal(20,2);default:0" json:"debit_amount"`
	CreditAmount   float64   `gorm:"type:decimal(20,2);default:0" json:"credit_amount"`
	Description    string    `gorm:"type:text" json:"description"`

	JournalEntry *JournalEntry `gorm:"foreignKey:JournalEntryID" json:"journal_entry,omitempty"`
	Account      *Account      `gorm:"foreignKey:AccountID" json:"account,omitempty"`
}

func (l *JournalEntryLine) TableName() string { return "journal_entry_lines" }

func (l *JournalEntryLine) BeforeCreate(tx *gorm.DB) error {
	if l.ID == uuid.Nil {
		l.ID = uuid.New()
	}
	return nil
}

// =============================================================================
// Accounting Period — Periode Akuntansi
// =============================================================================

type AccountingPeriod struct {
	ID          uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CompanyID   uuid.UUID  `gorm:"type:uuid;not null;index" json:"company_id"`
	PeriodYear  int        `gorm:"not null" json:"period_year"`
	PeriodMonth int        `gorm:"not null" json:"period_month"`                  // 1–12
	Status      string     `gorm:"type:varchar(20);default:'OPEN'" json:"status"` // OPEN, CLOSED
	ClosedAt    *time.Time `json:"closed_at"`
	ClosedBy    *uuid.UUID `gorm:"type:uuid" json:"closed_by"`
}

func (p *AccountingPeriod) TableName() string { return "accounting_periods" }

func (p *AccountingPeriod) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

// =============================================================================
// Account Balance — Cache saldo per akun per periode (untuk performa query)
// =============================================================================

type AccountBalance struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	AccountID   uuid.UUID `gorm:"type:uuid;not null;index" json:"account_id"`
	PeriodYear  int       `gorm:"not null" json:"period_year"`
	PeriodMonth int       `gorm:"not null" json:"period_month"`
	DebitTotal  float64   `gorm:"type:decimal(20,2);default:0" json:"debit_total"`
	CreditTotal float64   `gorm:"type:decimal(20,2);default:0" json:"credit_total"`
	// Balance = DebitTotal - CreditTotal untuk akun normal Debit
	//         = CreditTotal - DebitTotal untuk akun normal Kredit
	Balance   float64   `gorm:"type:decimal(20,2);default:0" json:"balance"`
	UpdatedAt time.Time `gorm:"default:now()" json:"updated_at"`

	Account *Account `gorm:"foreignKey:AccountID" json:"account,omitempty"`
}

func (b *AccountBalance) TableName() string { return "account_balances" }

func (b *AccountBalance) BeforeCreate(tx *gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return nil
}

// =============================================================================
// DTO — Report Structures (tidak disimpan ke DB)
// =============================================================================

// BalanceSheetReport — Neraca (PSAK)
type BalanceSheetReport struct {
	CompanyID   uuid.UUID         `json:"company_id"`
	CompanyName string            `json:"company_name"`
	AsOfDate    string            `json:"as_of_date"` // "31 Juli 2025"
	Assets      BalanceSheetGroup `json:"assets"`
	Liabilities BalanceSheetGroup `json:"liabilities"`
	Equity      BalanceSheetGroup `json:"equity"`
	TotalLE     float64           `json:"total_liabilities_equity"` // harus = Total Aset
}

type BalanceSheetGroup struct {
	Total    float64          `json:"total"`
	Accounts []AccountSummary `json:"accounts"`
}

type AccountSummary struct {
	Code    string  `json:"code"`
	Name    string  `json:"name"`
	Balance float64 `json:"balance"`
}

// CashFlowReport — Laporan Arus Kas (metode tidak langsung, PSAK 2)
type CashFlowReport struct {
	CompanyID     uuid.UUID       `json:"company_id"`
	CompanyName   string          `json:"company_name"`
	PeriodMonth   int             `json:"period_month"`
	PeriodYear    int             `json:"period_year"`
	Operating     CashFlowSection `json:"operating"` // arus kas dari aktivitas operasi
	Investing     CashFlowSection `json:"investing"` // arus kas dari aktivitas investasi
	Financing     CashFlowSection `json:"financing"` // arus kas dari aktivitas pendanaan
	NetCashChange float64         `json:"net_cash_change"`
	BeginningCash float64         `json:"beginning_cash"`
	EndingCash    float64         `json:"ending_cash"`
}

type CashFlowSection struct {
	Total float64        `json:"total"`
	Items []CashFlowItem `json:"items"`
}

type CashFlowItem struct {
	Description string  `json:"description"`
	Amount      float64 `json:"amount"` // positif = masuk, negatif = keluar
}

// TrialBalanceLine — Neraca Saldo
type TrialBalanceLine struct {
	AccountCode string  `json:"account_code"`
	AccountName string  `json:"account_name"`
	AccountType string  `json:"account_type"`
	DebitTotal  float64 `json:"debit_total"`
	CreditTotal float64 `json:"credit_total"`
	Balance     float64 `json:"balance"`
}

// CreateJournalEntryRequest — Request untuk manual journal entry
type CreateJournalEntryRequest struct {
	EntryDate   time.Time               `json:"entry_date"`
	Description string                  `json:"description"`
	Lines       []JournalEntryLineInput `json:"lines"`
}

type JournalEntryLineInput struct {
	AccountID    uuid.UUID `json:"account_id"`
	DebitAmount  float64   `json:"debit_amount"`
	CreditAmount float64   `json:"credit_amount"`
	Description  string    `json:"description"`
}
