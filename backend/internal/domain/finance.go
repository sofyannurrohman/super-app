package domain

import (
	"time"

	"github.com/google/uuid"
)

type ProfitLossReport struct {
	CompanyID   uuid.UUID `json:"company_id"`
	CompanyName string    `json:"company_name"`
	PeriodMonth int       `json:"period_month"`
	PeriodYear  int       `json:"period_year"`
	StartDate   string    `json:"start_date,omitempty"`
	EndDate     string    `json:"end_date,omitempty"`

	// Revenue
	TotalRevenue float64 `json:"total_revenue"`

	// Expenses
	TotalCOGS     float64 `json:"total_cogs"`     // Cost of Goods Sold
	TotalPayroll  float64 `json:"total_payroll"`  // Gaji Karyawan
	TotalExpenses float64 `json:"total_expenses"` // Other operational costs

	// Profit
	GrossProfit float64 `json:"gross_profit"` // Revenue - COGS
	NetProfit   float64 `json:"net_profit"`   // Gross - (Payroll + Expenses)
}

type BranchProfitLoss struct {
	BranchID      uuid.UUID `json:"branch_id"`
	BranchName    string    `json:"branch_name"`
	TotalRevenue  float64   `json:"total_revenue"`
	TotalCOGS     float64   `json:"total_cogs"`
	TotalPayroll  float64   `json:"total_payroll"`
	TotalExpenses float64   `json:"total_expenses"`
	GrossProfit   float64   `json:"gross_profit"`
	NetProfit     float64   `json:"net_profit"`
}

type ProfitLossDetailReport struct {
	CompanyID     uuid.UUID          `json:"company_id"`
	CompanyName   string             `json:"company_name"`
	PeriodMonth   int                `json:"period_month"`
	PeriodYear    int                `json:"period_year"`
	StartDate     string             `json:"start_date,omitempty"`
	EndDate       string             `json:"end_date,omitempty"`
	TotalRevenue  float64            `json:"total_revenue"`
	TotalCOGS     float64            `json:"total_cogs"`
	TotalPayroll  float64            `json:"total_payroll"`
	TotalExpenses float64            `json:"total_expenses"`
	GrossProfit   float64            `json:"gross_profit"`
	NetProfit     float64            `json:"net_profit"`
	Branches      []BranchProfitLoss `json:"branches"`
}

type IncomeBreakdown struct {
	ProductSales     float64 `json:"product_sales"`
	InterestIncome   float64 `json:"interest_income"`
	AmpasIncome      float64 `json:"ampas_income"`
	RosokIncome      float64 `json:"rosok_income"`
	ClaimPenalty     float64 `json:"claim_penalty"`
	OtherIncome      float64 `json:"other_income"`
	TotalOtherIncome float64 `json:"total_other_income"`
	TotalRevenue     float64 `json:"total_revenue"`
}

type ExpenseAccountItem struct {
	CategoryKey string  `json:"category_key"`
	Name        string  `json:"name"`
	Group       string  `json:"group"` // "LABOR", "UTILITIES", "FLEET", "TAX", "GENERAL"
	Amount      float64 `json:"amount"`
}

type ReceivableEntityItem struct {
	Code           string  `json:"code"`
	Name           string  `json:"name"`
	Category       string  `json:"category"` // "CABANG", "BOOSTER", "MARKETPLACE", "KARYAWAN", "PROYEK"
	InitialBalance float64 `json:"initial_balance"`
	PeriodAddition float64 `json:"period_addition"`
	PeriodPayment  float64 `json:"period_payment"`
	EndingBalance  float64 `json:"ending_balance"`
	Status         string  `json:"status"` // "LUNAS", "BERJALAN", "OVERDUE"
}

type CashierProfitLossReport struct {
	PeriodMonth         int                    `json:"period_month"`
	PeriodYear          int                    `json:"period_year"`
	StartDate           string                 `json:"start_date,omitempty"`
	EndDate             string                 `json:"end_date,omitempty"`
	BranchID            *uuid.UUID             `json:"branch_id,omitempty"`
	BranchName          string                 `json:"branch_name,omitempty"`
	Income              IncomeBreakdown        `json:"income"`
	TotalCOGS           float64                `json:"total_cogs"`
	GrossProfit         float64                `json:"gross_profit"`
	Expenses            []ExpenseAccountItem   `json:"expenses"`
	TotalExpenses       float64                `json:"total_expenses"`
	TotalPayrollDirect  float64                `json:"total_payroll_direct"`
	NetProfit           float64                `json:"net_profit"`
	ProfitMargin        float64                `json:"profit_margin"`
	Receivables         []ReceivableEntityItem `json:"receivables"`
	TotalReceivablesSum float64                `json:"total_receivables_sum"`
}

type BranchFinancialReceivableItem struct {
	ID                   uuid.UUID             `json:"id"`
	KasTransactionID     uuid.UUID             `json:"kas_transaction_id"`
	BranchID             uuid.UUID             `json:"branch_id"`
	BranchName           string                `json:"branch_name"`
	KasPabrikID          uuid.UUID             `json:"kas_pabrik_id"`
	KasName              string                `json:"kas_name"`
	TransactionType      string                `json:"transaction_type"` // OUT (Talangan) / IN (Pelunasan)
	Amount               float64               `json:"amount"`
	ExpenseCategory      string                `json:"expense_category,omitempty"`
	IncomeSource         string                `json:"income_source,omitempty"`
	TransactionDate      string                `json:"transaction_date"`
	Description          string                `json:"description"`
	ProofURL             string                `json:"proof_url,omitempty"`
	AdvanceStatus        string                `json:"advance_status,omitempty"`
	IsSettled            bool                  `json:"is_settled"`
	Settlement           *KasAdvanceSettlement `json:"settlement,omitempty"`
	RelatedTransactionID *uuid.UUID            `json:"related_transaction_id,omitempty"`
	CreatedAt            time.Time             `json:"created_at"`
}

type RepayBranchFinancialReceivableRequest struct {
	BranchID             uuid.UUID  `json:"branch_id" binding:"required"`
	KasPabrikID          uuid.UUID  `json:"kas_pabrik_id" binding:"required"`
	Amount               float64    `json:"amount" binding:"required,gt=0"`
	TransactionDate      string     `json:"transaction_date"`
	Description          string     `json:"description"`
	ProofURL             string     `json:"proof_url"`
	TargetTransactionID  *uuid.UUID `json:"target_transaction_id"`
}

type SettleFinancialReceivableTxRequest struct {
	ActualExpenseAmount  float64 `json:"actual_expense_amount" binding:"gte=0"`
	Notes                string  `json:"notes"`
	ProofURL             string  `json:"proof_url"`
}

type BranchFinancialSummary struct {
	BranchID             uuid.UUID                       `json:"branch_id"`
	BranchName           string                          `json:"branch_name"`
	TotalDisbursement    float64                         `json:"total_disbursement"`     // Total Talangan Keluar
	TotalRepayment       float64                         `json:"total_repayment"`        // Total Pelunasan Diterima
	OutstandingBalance   float64                         `json:"outstanding_balance"`    // Sisa Saldo Piutang Keuangan
	TotalRepaymentMonth  float64                         `json:"total_repayment_month"`  // Pelunasan Bulan Ini
	Status               string                          `json:"status"`                 // LUNAS / BELUM LUNAS
	Transactions         []BranchFinancialReceivableItem `json:"transactions,omitempty"`
}

type FinancialReceivablesResponse struct {
	TotalOutstanding float64                  `json:"total_outstanding"`
	TotalDisbursed   float64                  `json:"total_disbursed"`
	TotalRepaid      float64                  `json:"total_repaid"`
	Branches         []BranchFinancialSummary `json:"branches"`
}

type IncomeDetailItem struct {
	ID              uuid.UUID  `json:"id"`
	TransactionType string     `json:"transaction_type"` // "SALES_INVOICE", "SALES_ORDER", "KAS_IN"
	ReferenceNo     string     `json:"reference_no"`     // InvoiceNo, SONo, or Kas Transaction ID
	TransactionDate string     `json:"transaction_date"`
	EntityName      string     `json:"entity_name"`      // Customer/Store name, or Kas Name
	PaymentMethod   string     `json:"payment_method,omitempty"`
	Description     string     `json:"description"`
	Amount          float64    `json:"amount"`
	Status          string     `json:"status"`
	CreatedByName   string     `json:"created_by_name,omitempty"`
	ProofURL        string     `json:"proof_url,omitempty"`
}

type IncomeDetailResponse struct {
	IncomeType       string             `json:"income_type"`
	IncomeTitle      string             `json:"income_title"`
	PeriodMonth      int                `json:"period_month"`
	PeriodYear       int                `json:"period_year"`
	StartDate        string             `json:"start_date,omitempty"`
	EndDate          string             `json:"end_date,omitempty"`
	BranchID         *uuid.UUID         `json:"branch_id,omitempty"`
	TotalAmount      float64            `json:"total_amount"`
	TotalCount       int                `json:"total_count"`
	Items            []IncomeDetailItem `json:"items"`
}

type ExpenseDetailItem struct {
	ID              uuid.UUID `json:"id"`
	ReferenceNo     string    `json:"reference_no"`
	TransactionDate string    `json:"transaction_date"`
	KasName         string    `json:"kas_name"`
	CategoryKey     string    `json:"category_key"`
	CategoryName    string    `json:"category_name"`
	RecipientName   string    `json:"recipient_name"`
	Description     string    `json:"description"`
	Amount          float64   `json:"amount"`
	Status          string    `json:"status"`
	ProofURL        string    `json:"proof_url,omitempty"`
}

type ExpenseDetailResponse struct {
	CategoryKey  string              `json:"category_key"`
	CategoryName string              `json:"category_name"`
	Group        string              `json:"group"`
	PeriodMonth  int                 `json:"period_month"`
	PeriodYear   int                 `json:"period_year"`
	StartDate    string              `json:"start_date,omitempty"`
	EndDate      string              `json:"end_date,omitempty"`
	BranchID     *uuid.UUID          `json:"branch_id,omitempty"`
	TotalAmount  float64             `json:"total_amount"`
	TotalCount   int                 `json:"total_count"`
	Items        []ExpenseDetailItem `json:"items"`
}

type ReceivableDetailItem struct {
	ID              uuid.UUID `json:"id"`
	TransactionType string    `json:"transaction_type"` // "DEBET" | "KREDIT"
	ReferenceNo     string    `json:"reference_no"`
	TransactionDate string    `json:"transaction_date"`
	KasName         string    `json:"kas_name"`
	Description     string    `json:"description"`
	Amount          float64   `json:"amount"`
	Status          string    `json:"status"`
	ProofURL        string    `json:"proof_url,omitempty"`
}

type ReceivableDetailResponse struct {
	EntityCode     string                 `json:"entity_code"`
	EntityName     string                 `json:"entity_name"`
	Category       string                 `json:"category"`
	PeriodMonth    int                    `json:"period_month"`
	PeriodYear     int                    `json:"period_year"`
	StartDate      string                 `json:"start_date,omitempty"`
	EndDate        string                 `json:"end_date,omitempty"`
	BranchID       *uuid.UUID             `json:"branch_id,omitempty"`
	InitialBalance float64                `json:"initial_balance"`
	PeriodAddition float64                `json:"period_addition"`
	PeriodPayment  float64                `json:"period_payment"`
	EndingBalance  float64                `json:"ending_balance"`
	Status         string                 `json:"status"`
	TotalCount     int                    `json:"total_count"`
	Items          []ReceivableDetailItem `json:"items"`
}

type DailySalesIncomeSummary struct {
	Date              string  `json:"date"`
	TotalCash         float64 `json:"total_cash"`
	TotalTransfer     float64 `json:"total_transfer"`
	GrandTotal        float64 `json:"grand_total"`
	TotalTransactions int     `json:"total_transactions"`
}

type DailySalesIncomeBranch struct {
	BranchID         string  `json:"branch_id"`
	BranchName       string  `json:"branch_name"`
	CashAmount       float64 `json:"cash_amount"`
	TransferAmount   float64 `json:"transfer_amount"`
	TotalAmount      float64 `json:"total_amount"`
	TransactionCount int     `json:"transaction_count"`
}

type DailySalesIncomeCompany struct {
	CompanyID        string                   `json:"company_id"`
	CompanyName      string                   `json:"company_name"`
	CompanyCode      string                   `json:"company_code"`
	TotalCash        float64                  `json:"total_cash"`
	TotalTransfer    float64                  `json:"total_transfer"`
	TotalAmount      float64                  `json:"total_amount"`
	TransactionCount int                      `json:"transaction_count"`
	Branches         []DailySalesIncomeBranch `json:"branches"`
}

type DailySalesIncomePaymentItem struct {
	ID               uuid.UUID  `json:"id"`
	PaymentDate      time.Time  `json:"payment_date"`
	ApprovedAt       *time.Time `json:"approved_at"`
	ReferenceNo      string     `json:"reference_no"`
	CompanyID        string     `json:"company_id"`
	CompanyName      string     `json:"company_name"`
	BranchID         string     `json:"branch_id"`
	BranchName       string     `json:"branch_name"`
	StoreID          string     `json:"store_id"`
	StoreName        string     `json:"store_name"`
	SalesmanID       string     `json:"salesman_id"`
	SalesmanName     string     `json:"salesman_name"`
	ApproverID       string     `json:"approver_id"`
	ApproverName     string     `json:"approver_name"`
	PaymentMethod    string     `json:"payment_method"`
	Amount           float64    `json:"amount"`
	PaymentStatus    string     `json:"payment_status"`
	Notes            string     `json:"notes"`
	TransferProofURL *string    `json:"transfer_proof_url"`
}

type DailySalesIncomeReport struct {
	Summary   DailySalesIncomeSummary       `json:"summary"`
	Companies []DailySalesIncomeCompany     `json:"companies"`
	Payments  []DailySalesIncomePaymentItem `json:"payments"`
}


