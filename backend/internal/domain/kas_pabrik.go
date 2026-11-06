package domain

import (
	"time"

	"github.com/google/uuid"
	"example.com/hris/backend/pkg/utils"
	"gorm.io/gorm"
)

// KasIncomeSource — Master Sumber Penerimaan Kas Pabrik
type KasIncomeSource struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Code        string    `gorm:"type:varchar(50);not null;uniqueIndex" json:"code"`
	Name        string    `gorm:"type:varchar(200);not null" json:"name"`
	AccountCode string    `gorm:"type:varchar(50)" json:"account_code"`
	AccountName string    `gorm:"type:varchar(200)" json:"account_name"`
	Description string    `gorm:"type:text" json:"description"`
	IsActive    bool      `gorm:"default:true" json:"is_active"`
	SortOrder   int       `gorm:"default:0" json:"sort_order"`
	CreatedAt   time.Time `gorm:"default:now()" json:"created_at"`
	UpdatedAt   time.Time `gorm:"default:now()" json:"updated_at"`
}

func (KasIncomeSource) TableName() string { return "kas_income_sources" }
func (s *KasIncomeSource) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

// KasExpenseType — Master Tipe Pengeluaran (Kategori Induk) Kas Pabrik
type KasExpenseType struct {
	ID          uuid.UUID           `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Code        string              `gorm:"type:varchar(50);not null;uniqueIndex" json:"code"`
	Name        string              `gorm:"type:varchar(200);not null" json:"name"`
	Description string              `gorm:"type:text" json:"description"`
	IsActive    bool                `gorm:"default:true" json:"is_active"`
	SortOrder   int                 `gorm:"default:0" json:"sort_order"`
	CreatedAt   time.Time           `gorm:"default:now()" json:"created_at"`
	UpdatedAt   time.Time           `gorm:"default:now()" json:"updated_at"`
	Accounts    []KasExpenseAccount `gorm:"foreignKey:ExpenseTypeID" json:"accounts,omitempty"`
}

func (KasExpenseType) TableName() string { return "kas_expense_types" }
func (t *KasExpenseType) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}

// KasExpenseAccount — Master Akun Pengeluaran (Berelasi dengan Tipe Pengeluaran & COA Accurate)
type KasExpenseAccount struct {
	ID                uuid.UUID       `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ExpenseTypeID     uuid.UUID       `gorm:"type:uuid;not null;index" json:"expense_type_id"`
	AccountCode       string          `gorm:"type:varchar(50);not null;uniqueIndex" json:"account_code"`
	AccountName       string          `gorm:"type:varchar(200);not null" json:"account_name"`
	AccurateAccountID string          `gorm:"type:varchar(100)" json:"accurate_account_id"`
	Description       string          `gorm:"type:text" json:"description"`
	IsActive          bool            `gorm:"default:true" json:"is_active"`
	SortOrder         int             `gorm:"default:0" json:"sort_order"`
	CreatedAt         time.Time       `gorm:"default:now()" json:"created_at"`
	UpdatedAt         time.Time       `gorm:"default:now()" json:"updated_at"`
	ExpenseType       *KasExpenseType `gorm:"foreignKey:ExpenseTypeID" json:"expense_type,omitempty"`
}

func (KasExpenseAccount) TableName() string { return "kas_expense_accounts" }
func (a *KasExpenseAccount) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}

// KasPabrik — Master data KAS Kasir Pabrik
type KasPabrik struct {
	ID            uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	BranchID      uuid.UUID `gorm:"type:uuid;not null;index" json:"branch_id"`
	KasCode       string    `gorm:"type:varchar(50);not null;uniqueIndex" json:"kas_code"`
	KasName       string    `gorm:"type:varchar(200);not null" json:"kas_name"`
	Description   string    `gorm:"type:text" json:"description"`
	BankName      string    `gorm:"type:varchar(100)" json:"bank_name"`
	AccountNumber string    `gorm:"type:varchar(50)" json:"account_number"`
	Balance       float64   `gorm:"type:decimal(20,2);default:0;not null" json:"balance"`
	IsActive      bool      `gorm:"default:true" json:"is_active"`
	SortOrder     int       `gorm:"default:0" json:"sort_order"`
	CreatedAt     time.Time `gorm:"default:now()" json:"created_at"`
	UpdatedAt     time.Time `gorm:"default:now()" json:"updated_at"`

	Branch       *Branch          `gorm:"foreignKey:BranchID" json:"branch,omitempty"`
	Transactions []KasTransaction `gorm:"foreignKey:KasPabrikID" json:"transactions,omitempty"`
}

func (k *KasPabrik) TableName() string { return "kas_pabrik" }
func (k *KasPabrik) BeforeCreate(tx *gorm.DB) error {
	if k.ID == uuid.Nil {
		k.ID = uuid.New()
	}
	return nil
}

// KasTransaction — Riwayat transaksi setiap KAS
type KasTransaction struct {
	ID                   uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	KasPabrikID          uuid.UUID  `gorm:"type:uuid;not null;index" json:"kas_pabrik_id"`
	TransactionType      string     `gorm:"type:varchar(20);not null" json:"transaction_type"` // IN, OUT, TRANSFER_IN, TRANSFER_OUT
	Amount               float64    `gorm:"type:decimal(20,2);not null" json:"amount"`
	BalanceBefore        float64    `gorm:"type:decimal(20,2);not null" json:"balance_before"`
	BalanceAfter         float64    `gorm:"type:decimal(20,2);not null" json:"balance_after"`
	IncomeSourceID       *uuid.UUID `gorm:"type:uuid;index" json:"income_source_id"`
	ExpenseTypeID        *uuid.UUID `gorm:"type:uuid;index" json:"expense_type_id"`
	ExpenseAccountID     *uuid.UUID `gorm:"type:uuid;index" json:"expense_account_id"`
	IncomeSource         string     `gorm:"type:varchar(255)" json:"income_source"`        // Untuk IN (fallback string)
	ExpenseCategory      string     `gorm:"type:varchar(255)" json:"expense_category"`     // Untuk OUT (fallback string)
	RelatedKasPabrikID   *uuid.UUID `gorm:"type:uuid;index" json:"related_kas_pabrik_id"` // Untuk TRANSFER
	RelatedTransactionID *uuid.UUID `gorm:"type:uuid;index" json:"related_transaction_id"`
	Description          string     `gorm:"type:text" json:"description"`
	ProofURL             string     `gorm:"type:text" json:"proof_url"`
	TransactionDate      time.Time  `gorm:"type:date;not null" json:"transaction_date"`
	CreatedBy            *uuid.UUID `gorm:"type:uuid" json:"created_by"`
	CreatedAt            time.Time  `gorm:"default:now()" json:"created_at"`
	UpdatedAt            time.Time  `gorm:"default:now()" json:"updated_at"`

	KasPabrik          *KasPabrik            `gorm:"foreignKey:KasPabrikID" json:"kas_pabrik,omitempty"`
	RelatedKasPabrik   *KasPabrik            `gorm:"foreignKey:RelatedKasPabrikID" json:"related_kas_pabrik,omitempty"`
	RelatedTransaction *KasTransaction       `gorm:"foreignKey:RelatedTransactionID" json:"related_transaction,omitempty"`
	TargetBranch       *Branch               `gorm:"foreignKey:TargetBranchID" json:"target_branch,omitempty"`
	Vehicle            *Vehicle              `gorm:"foreignKey:VehicleID" json:"vehicle,omitempty"`
	Employee           *Employee             `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
	Settlement         *KasAdvanceSettlement `gorm:"foreignKey:KasTransactionID;constraint:OnDelete:CASCADE;" json:"settlement,omitempty"`
	IncomeSourceObj    *KasIncomeSource      `gorm:"foreignKey:IncomeSourceID" json:"income_source_obj,omitempty"`
	ExpenseTypeObj     *KasExpenseType       `gorm:"foreignKey:ExpenseTypeID" json:"expense_type_obj,omitempty"`
	ExpenseAccountObj  *KasExpenseAccount    `gorm:"foreignKey:ExpenseAccountID" json:"expense_account_obj,omitempty"`

	IsBranchReceivable bool       `gorm:"default:false" json:"is_branch_receivable"`
	TargetBranchID     *uuid.UUID `gorm:"type:uuid;index" json:"target_branch_id"`
	VehicleID          *uuid.UUID `gorm:"type:uuid;index" json:"vehicle_id"`
	EmployeeID         *uuid.UUID `gorm:"type:uuid;index" json:"employee_id"`
	AdvanceStatus      string     `gorm:"type:varchar(20);default:'NONE'" json:"advance_status"` // NONE, OUTSTANDING, SETTLED
	SalesPaymentID     *uuid.UUID `gorm:"type:uuid" json:"sales_payment_id"`
	SalesOrderID       *uuid.UUID `gorm:"type:uuid" json:"sales_order_id"`
	ReimbursementID    *uuid.UUID `gorm:"type:uuid;index" json:"reimbursement_id"`
	Reimbursement      *Reimbursement `gorm:"foreignKey:ReimbursementID" json:"reimbursement,omitempty"`
}

func (t *KasTransaction) TableName() string { return "kas_transactions" }
func (t *KasTransaction) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}

// KasAdvanceSettlement — Audit Trail Penyelesaian Kasbon / Uang Muka Dinas
type KasAdvanceSettlement struct {
	ID                  uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	KasTransactionID    uuid.UUID  `gorm:"type:uuid;not null;index" json:"kas_transaction_id"`
	ActualExpenseAmount float64    `gorm:"type:decimal(20,2);default:0;not null" json:"actual_expense_amount"`
	RefundAmount        float64    `gorm:"type:decimal(20,2);default:0;not null" json:"refund_amount"`
	ShortageAmount      float64    `gorm:"type:decimal(20,2);default:0;not null" json:"shortage_amount"`
	Notes               string     `gorm:"type:text" json:"notes"`
	ProofURL            string     `gorm:"type:text" json:"proof_url"`
	SettledBy           *uuid.UUID `gorm:"type:uuid" json:"settled_by"`
	SettledAt           time.Time  `gorm:"default:now()" json:"settled_at"`
	CreatedAt           time.Time  `gorm:"default:now()" json:"created_at"`
	UpdatedAt           time.Time  `gorm:"default:now()" json:"updated_at"`

	KasTransaction *KasTransaction `gorm:"foreignKey:KasTransactionID;constraint:OnDelete:CASCADE;" json:"kas_transaction,omitempty"`
	Settler        *User           `gorm:"foreignKey:SettledBy;constraint:OnDelete:SET NULL;" json:"settler,omitempty"`
}

func (KasAdvanceSettlement) TableName() string { return "kas_advance_settlements" }
func (s *KasAdvanceSettlement) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

// DTOs Master Data
type CreateKasIncomeSourceRequest struct {
	Code        string `json:"code" binding:"required"`
	Name        string `json:"name" binding:"required"`
	AccountCode string `json:"account_code"`
	AccountName string `json:"account_name"`
	Description string `json:"description"`
	SortOrder   int    `json:"sort_order"`
	IsActive    *bool  `json:"is_active"`
}

type UpdateKasIncomeSourceRequest struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	AccountCode string `json:"account_code"`
	AccountName string `json:"account_name"`
	Description string `json:"description"`
	SortOrder   int    `json:"sort_order"`
	IsActive    *bool  `json:"is_active"`
}

type CreateKasExpenseTypeRequest struct {
	Code        string `json:"code" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	SortOrder   int    `json:"sort_order"`
	IsActive    *bool  `json:"is_active"`
}

type UpdateKasExpenseTypeRequest struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
	SortOrder   int    `json:"sort_order"`
	IsActive    *bool  `json:"is_active"`
}

type CreateKasExpenseAccountRequest struct {
	ExpenseTypeID     string `json:"expense_type_id" binding:"required"`
	AccountCode       string `json:"account_code" binding:"required"`
	AccountName       string `json:"account_name" binding:"required"`
	AccurateAccountID string `json:"accurate_account_id"`
	Description       string `json:"description"`
	SortOrder         int    `json:"sort_order"`
	IsActive          *bool  `json:"is_active"`
}

type UpdateKasExpenseAccountRequest struct {
	ExpenseTypeID     string `json:"expense_type_id"`
	AccountCode       string `json:"account_code"`
	AccountName       string `json:"account_name"`
	AccurateAccountID string `json:"accurate_account_id"`
	Description       string `json:"description"`
	SortOrder         int    `json:"sort_order"`
	IsActive          *bool  `json:"is_active"`
}

// DTOs Kas Pabrik
type CreateKasPabrikRequest struct {
	BranchID      string `json:"branch_id" binding:"required"`
	KasCode       string `json:"kas_code" binding:"required"`
	KasName       string `json:"kas_name" binding:"required"`
	Description   string `json:"description"`
	BankName      string `json:"bank_name"`
	AccountNumber string `json:"account_number"`
	SortOrder     int    `json:"sort_order"`
}

type CreateKasTransactionRequest struct {
	TransactionType    string          `json:"transaction_type" binding:"required"` // IN or OUT
	Amount             utils.FlexFloat `json:"amount" binding:"required"`
	IncomeSourceID     string          `json:"income_source_id"`
	ExpenseTypeID      string          `json:"expense_type_id"`
	ExpenseAccountID   string          `json:"expense_account_id"`
	IncomeSource       string          `json:"income_source"`
	ExpenseCategory    string          `json:"expense_category"`
	Description        string          `json:"description"`
	ProofURL           string          `json:"proof_url"`
	TransactionDate    string          `json:"transaction_date"` // format: 2006-01-02
	IsBranchReceivable bool            `json:"is_branch_receivable"`
	TargetBranchID     string          `json:"target_branch_id"`
	VehicleID          string          `json:"vehicle_id"`
	EmployeeID         string          `json:"employee_id"`
	IsAdvance          bool            `json:"is_advance"` // Jika true, advance_status diisi 'OUTSTANDING'
	SalesPaymentID     string          `json:"sales_payment_id"`
	SalesOrderID       string          `json:"sales_order_id"`
}

type SettleKasAdvanceRequest struct {
	ActualExpenseAmount utils.FlexFloat `json:"actual_expense_amount" binding:"gte=0"`
	Notes               string          `json:"notes"`
	ProofURL            string          `json:"proof_url"`
}

type TransferKasRequest struct {
	FromKasID       string          `json:"from_kas_id" binding:"required"`
	ToKasID         string          `json:"to_kas_id" binding:"required"`
	Amount          utils.FlexFloat `json:"amount" binding:"required"`
	Description     string          `json:"description"`
	TransactionDate string          `json:"transaction_date"`
}

type ListKasTransactionsFilter struct {
	BranchID         string
	KasPabrikID      string
	Type             string
	DateFrom         string
	DateTo           string
	AdvanceStatus    string
	TargetBranchID   string
	EmployeeID       string
	ExpenseTypeID    string
	ExpenseAccountID string
	IncomeSourceID   string
	Page             int
	Limit            int
}

type UpdateKasPabrikRequest struct {
	KasCode       string `json:"kas_code"`
	KasName       string `json:"kas_name"`
	Description   string `json:"description"`
	BankName      string `json:"bank_name"`
	AccountNumber string `json:"account_number"`
	SortOrder     int    `json:"sort_order"`
	IsActive      *bool  `json:"is_active"`
}

type UpdateKasTransactionRequest struct {
	Amount           utils.FlexFloat `json:"amount"`
	IncomeSourceID   string          `json:"income_source_id"`
	ExpenseTypeID    string          `json:"expense_type_id"`
	ExpenseAccountID string          `json:"expense_account_id"`
	IncomeSource     string          `json:"income_source"`
	ExpenseCategory  string          `json:"expense_category"`
	Description      string          `json:"description"`
	ProofURL         string          `json:"proof_url"`
	TransactionDate  string          `json:"transaction_date"`
}
