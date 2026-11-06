package repository

import (
	"errors"
	"strings"
	"time"

	"example.com/hris/backend/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FinanceRepository interface {
	GetRevenueByCompany(companyID uuid.UUID, month, year int, startDate, endDate string) (float64, error)
	GetPayrollByCompany(companyID uuid.UUID, month, year int, startDate, endDate string) (float64, error)
	GetReimbursementByCompany(companyID uuid.UUID, month, year int, startDate, endDate string) (float64, error)
	GetCompanyByID(id uuid.UUID) (*domain.Company, error)
	GetAllCompanies() ([]domain.Company, error)
	GetRevenueByBranch(branchID uuid.UUID, month, year int, startDate, endDate string) (float64, error)
	GetPayrollByBranch(branchID uuid.UUID, month, year int, startDate, endDate string) (float64, error)
	GetReimbursementByBranch(branchID uuid.UUID, month, year int, startDate, endDate string) (float64, error)
	GetBranchesByCompany(companyID uuid.UUID) ([]domain.Branch, error)
	GetBranchByID(id uuid.UUID) (*domain.Branch, error)
	GetPendingPayments(companyID, branchID uuid.UUID, page, limit int, search, status, startDate, endDate string) ([]domain.SalesPayment, int64, error)
	GetMidtransPayments(companyID, branchID uuid.UUID, page, limit int, search, status, startDate, endDate string) ([]domain.SalesPayment, int64, error)
	GetKasExpenseBreakdown(month, year int, branchID *uuid.UUID, startDate, endDate string) (map[string]float64, error)
	GetKasIncomeBreakdown(month, year int, branchID *uuid.UUID, startDate, endDate string) (map[string]float64, error)
	GetReceivablesBreakdown(month, year int, branchID *uuid.UUID, startDate, endDate string) ([]domain.ReceivableEntityItem, error)
	GetFinancialReceivablesList(companyID uuid.UUID, branchID *uuid.UUID) (*domain.FinancialReceivablesResponse, error)
	RecordFinancialReceivablePayment(req domain.RepayBranchFinancialReceivableRequest, recordedBy *uuid.UUID) (*domain.KasTransaction, error)
	SettleFinancialReceivableTx(txID uuid.UUID, req domain.SettleFinancialReceivableTxRequest, settledBy *uuid.UUID) (*domain.KasAdvanceSettlement, error)
	GetIncomeTransactionDetails(incomeType string, month, year int, branchID *uuid.UUID, startDate, endDate string) ([]domain.IncomeDetailItem, error)
	GetExpenseTransactionDetails(categoryKey string, month, year int, branchID *uuid.UUID, startDate, endDate string) ([]domain.ExpenseDetailItem, error)
	GetReceivableTransactionDetails(entityCode string, month, year int, branchID *uuid.UUID, startDate, endDate string) (*domain.ReceivableDetailResponse, error)
}

type financeRepository struct {
	db *gorm.DB
}

func NewFinanceRepository(db *gorm.DB) FinanceRepository {
	return &financeRepository{db}
}

func (r *financeRepository) GetRevenueByCompany(companyID uuid.UUID, month, year int, startDate, endDate string) (float64, error) {
	var transactionTotal float64
	txQuery := r.db.Model(&domain.SalesTransaction{}).
		Where("company_id = ? AND status = ?", companyID, "VERIFIED")
	if startDate != "" && endDate != "" {
		txQuery = txQuery.Where("DATE(transaction_date) >= ? AND DATE(transaction_date) <= ?", startDate, endDate)
	} else {
		txQuery = txQuery.Where("period_month = ? AND period_year = ?", month, year)
	}

	err := txQuery.Select("COALESCE(SUM(total_amount), 0)").Scan(&transactionTotal).Error
	if err != nil {
		return 0, err
	}

	// For SalesOrder, ONLY include completed/delivered/paid orders that HAVE NOT been converted to a SalesTransaction invoice (invoice_id IS NULL)
	// to prevent double-counting of invoices.
	var soTotal float64
	soQuery := r.db.Model(&domain.SalesOrder{}).
		Where("company_id = ? AND status IN ? AND invoice_id IS NULL", companyID, []string{string(domain.SOStatusDelivered), string(domain.SOStatusPaid), string(domain.SOStatusShipped)})
	if startDate != "" && endDate != "" {
		soQuery = soQuery.Where("DATE(order_date) >= ? AND DATE(order_date) <= ?", startDate, endDate)
	} else {
		soQuery = soQuery.Where("EXTRACT(MONTH FROM order_date) = ? AND EXTRACT(YEAR FROM order_date) = ?", month, year)
	}

	errSO := soQuery.Select("COALESCE(SUM(total_amount), 0)").Scan(&soTotal).Error
	if errSO != nil {
		return 0, errSO
	}

	return transactionTotal + soTotal, nil
}

func (r *financeRepository) GetPayrollByCompany(companyID uuid.UUID, month, year int, startDate, endDate string) (float64, error) {
	var total float64
	query := r.db.Model(&domain.PayrollRun{}).
		Where("company_id = ? AND status != ?", companyID, "DRAFT")
	if startDate != "" && endDate != "" {
		query = query.Where("DATE(period_start) <= ? AND DATE(period_end) >= ?", endDate, startDate)
	} else {
		query = query.Where("EXTRACT(MONTH FROM period_start) = ? AND EXTRACT(YEAR FROM period_start) = ?", month, year)
	}

	err := query.Select("COALESCE(SUM(total_payout), 0)").Scan(&total).Error
	return total, err
}

func (r *financeRepository) GetReimbursementByCompany(companyID uuid.UUID, month, year int, startDate, endDate string) (float64, error) {
	var total float64
	query := r.db.Model(&domain.Reimbursement{}).
		Joins("JOIN employees ON employees.id = reimbursements.employee_id").
		Where("employees.company_id = ? AND status = ?", companyID, domain.ReimbursementApproved)
	if startDate != "" && endDate != "" {
		query = query.Where("DATE(reimbursements.created_at) >= ? AND DATE(reimbursements.created_at) <= ?", startDate, endDate)
	} else {
		query = query.Where("EXTRACT(MONTH FROM reimbursements.created_at) = ? AND EXTRACT(YEAR FROM reimbursements.created_at) = ?", month, year)
	}

	err := query.Select("COALESCE(SUM(reimbursements.amount), 0)").Scan(&total).Error
	return total, err
}

func (r *financeRepository) GetCompanyByID(id uuid.UUID) (*domain.Company, error) {
	var company domain.Company
	err := r.db.First(&company, id).Error
	if err != nil {
		return nil, err
	}
	return &company, nil
}

func (r *financeRepository) GetAllCompanies() ([]domain.Company, error) {
	var companies []domain.Company
	err := r.db.Order("name ASC").Find(&companies).Error
	return companies, err
}

func (r *financeRepository) GetRevenueByBranch(branchID uuid.UUID, month, year int, startDate, endDate string) (float64, error) {
	var branch domain.Branch
	isHeadOfficeOrHasCompany := false
	var compID uuid.UUID
	if err := r.db.First(&branch, "id = ?", branchID).Error; err == nil {
		if branch.CompanyID != nil && *branch.CompanyID != uuid.Nil {
			isHeadOfficeOrHasCompany = true
			compID = *branch.CompanyID
		}
	} else {
		var comp domain.Company
		if err := r.db.First(&comp, "id = ?", branchID).Error; err == nil {
			isHeadOfficeOrHasCompany = true
			compID = comp.ID
		}
	}

	var transactionTotal float64
	txQuery := r.db.Model(&domain.SalesTransaction{}).
		Where("status = ?", "VERIFIED")
	if isHeadOfficeOrHasCompany {
		txQuery = txQuery.Where("(branch_id = ? OR (branch_id IS NULL AND company_id = ?))", branchID, compID)
	} else {
		txQuery = txQuery.Where("branch_id = ?", branchID)
	}
	if startDate != "" && endDate != "" {
		txQuery = txQuery.Where("DATE(transaction_date) >= ? AND DATE(transaction_date) <= ?", startDate, endDate)
	} else {
		txQuery = txQuery.Where("period_month = ? AND period_year = ?", month, year)
	}

	err := txQuery.Select("COALESCE(SUM(total_amount), 0)").Scan(&transactionTotal).Error
	if err != nil {
		return 0, err
	}

	// For SalesOrder, ONLY include completed/delivered/paid orders that HAVE NOT been converted to a SalesTransaction invoice (invoice_id IS NULL)
	// to prevent double-counting of invoices.
	var soTotal float64
	soQuery := r.db.Model(&domain.SalesOrder{}).
		Where("status IN ? AND invoice_id IS NULL", []string{string(domain.SOStatusDelivered), string(domain.SOStatusPaid), string(domain.SOStatusShipped)})
	if isHeadOfficeOrHasCompany {
		soQuery = soQuery.Where("(branch_id = ? OR (branch_id IS NULL AND company_id = ?))", branchID, compID)
	} else {
		soQuery = soQuery.Where("branch_id = ?", branchID)
	}
	if startDate != "" && endDate != "" {
		soQuery = soQuery.Where("DATE(order_date) >= ? AND DATE(order_date) <= ?", startDate, endDate)
	} else {
		soQuery = soQuery.Where("EXTRACT(MONTH FROM order_date) = ? AND EXTRACT(YEAR FROM order_date) = ?", month, year)
	}

	errSO := soQuery.Select("COALESCE(SUM(total_amount), 0)").Scan(&soTotal).Error
	if errSO != nil {
		return 0, errSO
	}

	return transactionTotal + soTotal, nil
}

func (r *financeRepository) GetPayrollByBranch(branchID uuid.UUID, month, year int, startDate, endDate string) (float64, error) {
	var total float64
	query := r.db.Table("payslips").
		Joins("JOIN payroll_runs ON payroll_runs.id = payslips.payroll_run_id").
		Joins("JOIN employees ON employees.id = payslips.employee_id").
		Where("COALESCE(employees.operating_branch_id, employees.branch_id) = ? AND payroll_runs.status != ?", branchID, "DRAFT")
	if startDate != "" && endDate != "" {
		query = query.Where("DATE(payroll_runs.period_start) <= ? AND DATE(payroll_runs.period_end) >= ?", endDate, startDate)
	} else {
		query = query.Where("EXTRACT(MONTH FROM payroll_runs.period_start) = ? AND EXTRACT(YEAR FROM payroll_runs.period_start) = ?", month, year)
	}

	err := query.Select("COALESCE(SUM(payslips.take_home_pay), 0)").Scan(&total).Error
	return total, err
}

func (r *financeRepository) GetReimbursementByBranch(branchID uuid.UUID, month, year int, startDate, endDate string) (float64, error) {
	var total float64
	query := r.db.Model(&domain.Reimbursement{}).
		Joins("JOIN employees ON employees.id = reimbursements.employee_id").
		Where("COALESCE(employees.operating_branch_id, employees.branch_id) = ? AND status = ?", branchID, domain.ReimbursementApproved)
	if startDate != "" && endDate != "" {
		query = query.Where("DATE(reimbursements.created_at) >= ? AND DATE(reimbursements.created_at) <= ?", startDate, endDate)
	} else {
		query = query.Where("EXTRACT(MONTH FROM reimbursements.created_at) = ? AND EXTRACT(YEAR FROM reimbursements.created_at) = ?", month, year)
	}

	err := query.Select("COALESCE(SUM(reimbursements.amount), 0)").Scan(&total).Error
	return total, err
}

func (r *financeRepository) GetBranchesByCompany(companyID uuid.UUID) ([]domain.Branch, error) {
	var branches []domain.Branch
	err := r.db.Where("company_id = ?", companyID).Order("name ASC").Find(&branches).Error
	return branches, err
}

func (r *financeRepository) GetBranchByID(id uuid.UUID) (*domain.Branch, error) {
	var branch domain.Branch
	err := r.db.First(&branch, id).Error
	if err != nil {
		return nil, err
	}
	return &branch, nil
}

func (r *financeRepository) GetPendingPayments(companyID, branchID uuid.UUID, page, limit int, search, status, startDate, endDate string) ([]domain.SalesPayment, int64, error) {
	var payments []domain.SalesPayment
	var total int64
	db := r.db.Model(&domain.SalesPayment{}).
		Joins("LEFT JOIN sales_orders ON sales_orders.id = sales_payments.sales_order_id").
		Joins("LEFT JOIN sales_transactions ON sales_transactions.id = sales_payments.sales_transaction_id").
		Joins("LEFT JOIN stores s1 ON s1.id = sales_orders.store_id").
		Joins("LEFT JOIN stores s2 ON s2.id = sales_transactions.store_id").
		Joins("LEFT JOIN employees e ON e.id = sales_payments.employee_id").
		Joins("LEFT JOIN employees e_so ON e_so.id = sales_orders.employee_id").
		Joins("LEFT JOIN employees e_st ON e_st.id = sales_transactions.employee_id").
		Where("sales_payments.payment_method IN ('CASH', 'TUNAI') AND sales_payments.amount > 0").
		Where("sales_transactions.id IS NULL OR sales_transactions.status != 'REJECTED'")

	if startDate != "" {
		db = db.Where("COALESCE(DATE(sales_payments.approved_at), DATE(sales_payments.payment_date), DATE(sales_payments.created_at)) >= ?", startDate)
	}
	if endDate != "" {
		db = db.Where("COALESCE(DATE(sales_payments.approved_at), DATE(sales_payments.payment_date), DATE(sales_payments.created_at)) <= ?", endDate)
	}

	if status != "" && status != "all" {
		if status == "success" {
			db = db.Where("sales_payments.payment_status IN ?", []string{"SUCCESS", "PAID", "APPROVED", "CASHIER_APPROVED"})
		} else if status == "rejected" {
			db = db.Where("sales_payments.payment_status IN ?", []string{"REJECTED", "FAILED"})
		} else {
			db = db.Where("sales_payments.payment_status = ?", "PENDING")
		}
	} else {
		// Default to pending if not specified, though frontend will send 'all' for all
		if status == "" {
			db = db.Where("sales_payments.payment_status = ?", "PENDING")
		}
	}

	if companyID != uuid.Nil {
		db = db.Where("sales_orders.company_id = ? OR sales_transactions.company_id = ?", companyID, companyID)
	}

	if branchID != uuid.Nil {
		db = db.Where(`COALESCE(
			sales_transactions.branch_id, sales_orders.branch_id,
			e_st.operating_branch_id, e_st.branch_id,
			e_so.operating_branch_id, e_so.branch_id,
			e.operating_branch_id, e.branch_id
		) = ?`, branchID)
	}

	if search != "" {
		like := "%" + search + "%"
		db = db.Where("s1.name ILIKE ? OR s2.name ILIKE ? OR e.first_name ILIKE ? OR e.last_name ILIKE ? OR CONCAT(e.first_name, ' ', COALESCE(e.last_name, '')) ILIKE ? OR sales_transactions.buyer_name ILIKE ? OR sales_transactions.receipt_no ILIKE ? OR sales_orders.so_number ILIKE ? OR EXISTS (SELECT 1 FROM sales_transactions invoice WHERE invoice.id = sales_orders.invoice_id AND invoice.receipt_no ILIKE ?)", like, like, like, like, like, like, like, like, like)
	}

	err := db.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	if limit == -1 {
		offset = 0
	}
	db = db.Preload("Employee").
		Preload("CompanyBankAccount").
		Preload("Employee.Branch").
		Preload("Employee.OperatingBranch").
		Preload("VerifiedBy").
		Preload("VerifiedBy.JobPosition").
		Preload("SalesOrder").Preload("SalesOrder.Store").Preload("SalesOrder.Employee").
		Preload("SalesOrder.Employee.Branch").Preload("SalesOrder.Employee.OperatingBranch").
		Preload("SalesOrder.Items").Preload("SalesOrder.Items.Product").
		Preload("SalesOrder.Invoice").
		Preload("SalesTransaction").Preload("SalesTransaction.Store").Preload("SalesTransaction.Employee").
		Preload("SalesTransaction.Employee.Branch").Preload("SalesTransaction.Employee.OperatingBranch").
		Preload("SalesTransaction.BuyerEmployee").Preload("SalesTransaction.Branch").
		Preload("SalesTransaction.KasPabrik").
		Preload("SalesTransaction.Items").Preload("SalesTransaction.Items.Product").
		Order("sales_payments.created_at DESC").
		Offset(offset).Limit(limit).
		Find(&payments)

	return payments, total, db.Error
}

func (r *financeRepository) GetMidtransPayments(companyID, branchID uuid.UUID, page, limit int, search, status, startDate, endDate string) ([]domain.SalesPayment, int64, error) {
	var payments []domain.SalesPayment
	var total int64
	db := r.db.Model(&domain.SalesPayment{}).
		Joins("LEFT JOIN sales_orders ON sales_orders.id = sales_payments.sales_order_id").
		Joins("LEFT JOIN sales_transactions ON sales_transactions.id = sales_payments.sales_transaction_id").
		Joins("LEFT JOIN stores s1 ON s1.id = sales_orders.store_id").
		Joins("LEFT JOIN stores s2 ON s2.id = sales_transactions.store_id").
		Joins("LEFT JOIN employees e ON e.id = sales_payments.employee_id").
		Joins("LEFT JOIN employees e_so ON e_so.id = sales_orders.employee_id").
		Joins("LEFT JOIN employees e_st ON e_st.id = sales_transactions.employee_id").
		Where("sales_payments.payment_method IN ('TRANSFER', 'VA', 'QRIS', 'MIDTRANS_QRIS') AND sales_payments.amount > 0")

	if startDate != "" {
		db = db.Where("COALESCE(DATE(sales_payments.approved_at), DATE(sales_payments.payment_date), DATE(sales_payments.created_at)) >= ?", startDate)
	}
	if endDate != "" {
		db = db.Where("COALESCE(DATE(sales_payments.approved_at), DATE(sales_payments.payment_date), DATE(sales_payments.created_at)) <= ?", endDate)
	}

	if status != "" && status != "all" {
		if status == "success" {
			db = db.Where("sales_payments.payment_status IN ?", []string{"SUCCESS", "PAID", "APPROVED", "CASHIER_APPROVED"})
		} else if status == "rejected" {
			db = db.Where("sales_payments.payment_status IN ?", []string{"REJECTED", "FAILED"})
		} else {
			db = db.Where("sales_payments.payment_status = ?", "PENDING")
		}
	} else {
		// Default to pending if not specified, though frontend will send 'all' for all
		if status == "" {
			db = db.Where("sales_payments.payment_status = ?", "PENDING")
		}
	}

	if companyID != uuid.Nil {
		db = db.Where("sales_orders.company_id = ? OR sales_transactions.company_id = ?", companyID, companyID)
	}

	if branchID != uuid.Nil {
		db = db.Where(`COALESCE(
			sales_transactions.branch_id, sales_orders.branch_id,
			e_st.operating_branch_id, e_st.branch_id,
			e_so.operating_branch_id, e_so.branch_id,
			e.operating_branch_id, e.branch_id
		) = ?`, branchID)
	}

	if search != "" {
		like := "%" + search + "%"
		db = db.Where("s1.name ILIKE ? OR s2.name ILIKE ? OR e.first_name ILIKE ? OR e.last_name ILIKE ? OR CONCAT(e.first_name, ' ', COALESCE(e.last_name, '')) ILIKE ? OR sales_transactions.buyer_name ILIKE ? OR sales_transactions.receipt_no ILIKE ? OR sales_orders.so_number ILIKE ? OR EXISTS (SELECT 1 FROM sales_transactions invoice WHERE invoice.id = sales_orders.invoice_id AND invoice.receipt_no ILIKE ?)", like, like, like, like, like, like, like, like, like)
	}

	err := db.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	if limit == -1 {
		offset = 0
	}
	db = db.Preload("Employee").
		Preload("CompanyBankAccount").
		Preload("Employee.Branch").
		Preload("Employee.OperatingBranch").
		Preload("VerifiedBy").
		Preload("VerifiedBy.JobPosition").
		Preload("SalesOrder").Preload("SalesOrder.Store").Preload("SalesOrder.Employee").
		Preload("SalesOrder.Employee.Branch").Preload("SalesOrder.Employee.OperatingBranch").
		Preload("SalesOrder.Items").Preload("SalesOrder.Items.Product").
		Preload("SalesOrder.Invoice").
		Preload("SalesTransaction").Preload("SalesTransaction.Store").Preload("SalesTransaction.Employee").
		Preload("SalesTransaction.Employee.Branch").Preload("SalesTransaction.Employee.OperatingBranch").
		Preload("SalesTransaction.BuyerEmployee").Preload("SalesTransaction.Branch").
		Preload("SalesTransaction.KasPabrik").
		Preload("SalesTransaction.Items").Preload("SalesTransaction.Items.Product").
		Order("sales_payments.created_at DESC").
		Offset(offset).Limit(limit).
		Find(&payments)

	return payments, total, db.Error
}

func (r *financeRepository) GetKasExpenseBreakdown(month, year int, branchID *uuid.UUID, startDate, endDate string) (map[string]float64, error) {
	type Result struct {
		ExpenseCategory string  `gorm:"column:expense_category"`
		Total           float64 `gorm:"column:total"`
	}
	var results []Result

	query := r.db.Table("kas_transactions").
		Select("expense_category, COALESCE(SUM(amount), 0) as total").
		Where("transaction_type = 'OUT'")

	if startDate != "" && endDate != "" {
		query = query.Where("DATE(transaction_date) >= ? AND DATE(transaction_date) <= ?", startDate, endDate)
	} else {
		query = query.Where("EXTRACT(MONTH FROM transaction_date) = ? AND EXTRACT(YEAR FROM transaction_date) = ?", month, year)
	}

	if branchID != nil && *branchID != uuid.Nil {
		query = query.Joins("JOIN kas_pabrik ON kas_pabrik.id = kas_transactions.kas_pabrik_id").
			Where("kas_pabrik.branch_id = ?", *branchID)
	}

	err := query.Group("expense_category").Scan(&results).Error
	if err != nil {
		return nil, err
	}

	resMap := make(map[string]float64)
	for _, res := range results {
		if res.ExpenseCategory != "" {
			resMap[res.ExpenseCategory] = res.Total
		}
	}
	return resMap, nil
}

func normalizeIncomeSourceKey(code, name, rawSrc string) string {
	combined := strings.ToUpper(strings.TrimSpace(code + " " + name + " " + rawSrc))
	if combined == "" {
		return "OTHER"
	}

	switch {
	case strings.Contains(combined, "AMPAS"):
		return "AMPAS_SALE"
	case strings.Contains(combined, "ROSOK") || strings.Contains(combined, "AFVAL") || strings.Contains(combined, "JURIGEN") || strings.Contains(combined, "DRUM"):
		return "ROSOK_SALE"
	case strings.Contains(combined, "INTEREST") || strings.Contains(combined, "BUNGA"):
		return "INTEREST_INCOME"
	case strings.Contains(combined, "CLAIM") || strings.Contains(combined, "KLAIM") || strings.Contains(combined, "PENALTY") || strings.Contains(combined, "DENDA"):
		return "CLAIM_PENALTY"
	case strings.Contains(combined, "PRODUCT_SALE") || strings.Contains(combined, "PRODUK") || strings.Contains(combined, "TELEMARKETING") || strings.Contains(combined, "FAKTUR"):
		return "PRODUCT_SALE"
	case strings.Contains(combined, "ADVANCE_REFUND") || strings.Contains(combined, "KASBON"):
		return "ADVANCE_REFUND"
	case strings.Contains(combined, "BRANCH_RECEIVABLE") || strings.Contains(combined, "TALANGAN"):
		return "BRANCH_RECEIVABLE_PAYMENT"
	case strings.Contains(combined, "BRANCH_DEPOSIT") || strings.Contains(combined, "SETORAN"):
		return "BRANCH_DEPOSIT"
	case strings.Contains(combined, "BANK_DISBURSEMENT") || strings.Contains(combined, "PENCAIRAN"):
		return "BANK_DISBURSEMENT"
	case strings.Contains(combined, "TRANSFER_FROM_KAS") || strings.Contains(combined, "TRANSFER"):
		return "TRANSFER_FROM_KAS"
	default:
		if code != "" {
			return code
		}
		if rawSrc != "" {
			return rawSrc
		}
		return "OTHER"
	}
}

func (r *financeRepository) GetKasIncomeBreakdown(month, year int, branchID *uuid.UUID, startDate, endDate string) (map[string]float64, error) {
	type Result struct {
		RawCode string  `gorm:"column:code"`
		RawName string  `gorm:"column:name"`
		RawSrc  string  `gorm:"column:income_source"`
		Total   float64 `gorm:"column:total"`
	}
	var results []Result

	query := r.db.Table("kas_transactions kt").
		Select("COALESCE(kis.code, '') as code, COALESCE(kis.name, '') as name, COALESCE(kt.income_source, '') as income_source, COALESCE(SUM(kt.amount), 0) as total").
		Joins("LEFT JOIN kas_income_sources kis ON kis.id = kt.income_source_id").
		Where("kt.transaction_type = 'IN'")

	if startDate != "" && endDate != "" {
		query = query.Where("DATE(kt.transaction_date) >= ? AND DATE(kt.transaction_date) <= ?", startDate, endDate)
	} else {
		query = query.Where("EXTRACT(MONTH FROM kt.transaction_date) = ? AND EXTRACT(YEAR FROM kt.transaction_date) = ?", month, year)
	}

	if branchID != nil && *branchID != uuid.Nil {
		query = query.Joins("JOIN kas_pabrik kp ON kp.id = kt.kas_pabrik_id").
			Where("kp.branch_id = ?", *branchID)
	}

	err := query.Group("kis.code, kis.name, kt.income_source").Scan(&results).Error
	if err != nil {
		return nil, err
	}

	resMap := make(map[string]float64)
	for _, res := range results {
		key := normalizeIncomeSourceKey(res.RawCode, res.RawName, res.RawSrc)
		if key != "" {
			resMap[key] += res.Total
		}
	}
	return resMap, nil
}

type receivableDef struct {
	Code     string
	Name     string
	Category string
	BranchKW string
	KW       []string
}

var standardReceivableEntities = []receivableDef{
	// Client branches are configured in the branch master data, not seeded here.

	// 14-18 Booster & Distributor
	{"RCV-14", "Jaringan Booster Area Timur", "BOOSTER", "", []string{"BOOSTER TIMUR", "BOOSTER AREA TIMUR"}},
	{"RCV-15", "Jaringan Booster Area Barat", "BOOSTER", "", []string{"BOOSTER BARAT", "BOOSTER AREA BARAT"}},
	{"RCV-16", "Jaringan Booster Area Tengah", "BOOSTER", "", []string{"BOOSTER TENGAH", "BOOSTER AREA TENGAH", "BOOSTER"}},
	{"RCV-17", "Distributor Utama / Grosir", "DISTRIBUTOR", "", []string{"DISTRIBUTOR", "GROSIR"}},
	{"RCV-18", "Agen & Sub-Distributor", "DISTRIBUTOR", "", []string{"AGEN", "SUB-DISTRIBUTOR", "SUB DISTRIBUTOR"}},

	// 19-21 Marketplace
	{"RCV-19", "Marketplace Shopee", "MARKETPLACE", "", []string{"SHOPEE"}},
	{"RCV-20", "Marketplace Tokopedia", "MARKETPLACE", "", []string{"TOKOPEDIA", "TOPED"}},
	{"RCV-21", "Marketplace TikTok Shop & Lainnya", "MARKETPLACE", "", []string{"TIKTOK", "LAZADA", "MARKETPLACE"}},

	// 22-23 Karyawan / Dinas
	{"RCV-22", "Kasbon Dinas Karyawan (Advance Loan)", "KARYAWAN", "", []string{"KASBON", "ADVANCE"}},
	{"RCV-23", "Piutang Karyawan / Staff Kasir", "KARYAWAN", "", []string{"STAFF", "KARYAWAN"}},

	// 24-28 Proyek & Pos Operasional Khusus
	{"RCV-26", "Talangan Ekspedisi", "PROYEK", "", []string{"EXPEDISI", "PENGIRIMAN"}},
	{"RCV-27", "Talangan Operasional Cabang Lainnya", "PROYEK", "", []string{"OPERASIONAL CABANG"}},
	{"RCV-28", "Piutang Retur & Klaim Pihak Ketiga", "PROYEK", "", []string{"RETUR", "KLAIM", "CLAIM"}},
}

func (r *financeRepository) receivableEntities() []receivableDef {
	entities := append([]receivableDef(nil), standardReceivableEntities...)
	var branches []domain.Branch
	if err := r.db.Select("id, name").Find(&branches).Error; err == nil {
		for _, branch := range branches {
			entities = append(entities, receivableDef{
				Code: "BR-" + branch.ID.String(), Name: branch.Name,
				Category: "CABANG", BranchKW: branch.Name,
			})
		}
	}
	return entities
}

func (r *financeRepository) getBranchMap() map[string]uuid.UUID {
	branchMap := make(map[string]uuid.UUID)
	var branches []domain.Branch
	if err := r.db.Select("id, name").Find(&branches).Error; err == nil {
		for _, b := range branches {
			branchMap[b.Name] = b.ID
		}
	}
	return branchMap
}

func (r *financeRepository) buildReceivableQuery(def receivableDef, branchMap map[string]uuid.UUID, month, year int, branchID *uuid.UUID, startDate, endDate string) *gorm.DB {
	query := r.db.Model(&domain.KasTransaction{})

	if startDate != "" && endDate != "" {
		query = query.Where("DATE(kas_transactions.transaction_date) >= ? AND DATE(kas_transactions.transaction_date) <= ?", startDate, endDate)
	} else {
		query = query.Where("EXTRACT(MONTH FROM kas_transactions.transaction_date) = ? AND EXTRACT(YEAR FROM kas_transactions.transaction_date) = ?", month, year)
	}

	if branchID != nil && *branchID != uuid.Nil {
		query = query.Joins("JOIN kas_pabrik kp ON kp.id = kas_transactions.kas_pabrik_id").
			Where("kp.branch_id = ?", *branchID)
	}

	if def.Category == "CABANG" && def.BranchKW != "" {
		targetBID, exists := branchMap[def.BranchKW]
		if exists && targetBID != uuid.Nil {
			query = query.Where(`(
				kas_transactions.target_branch_id = ? 
				OR (kas_transactions.is_branch_receivable = TRUE AND (kas_transactions.target_branch_id = ? OR kas_transactions.description ILIKE ?))
				OR (kas_transactions.expense_category = 'BIAYA_OPERASIONAL_CABANG' AND (kas_transactions.target_branch_id = ? OR kas_transactions.description ILIKE ?))
				OR (kas_transactions.income_source IN ('BRANCH_RECEIVABLE_PAYMENT', 'BRANCH_SETTLEMENT', 'BRANCH_DEPOSIT') AND (kas_transactions.target_branch_id = ? OR kas_transactions.description ILIKE ?))
			)`, targetBID, targetBID, "%"+def.BranchKW+"%", targetBID, "%"+def.BranchKW+"%", targetBID, "%"+def.BranchKW+"%")
		} else {
			query = query.Where("kas_transactions.description ILIKE ?", "%"+def.BranchKW+"%")
		}
	} else if def.Code == "RCV-22" {
		query = query.Where("(kas_transactions.advance_status IN ('OUTSTANDING', 'SETTLED') OR kas_transactions.expense_category = 'KASBON_DINAS' OR kas_transactions.income_source = 'ADVANCE_REFUND' OR kas_transactions.description ILIKE '%kasbon%')")
	} else if def.Code == "RCV-23" {
		query = query.Where("(kas_transactions.employee_id IS NOT NULL OR kas_transactions.description ILIKE '%karyawan%' OR kas_transactions.description ILIKE '%staff%')")
	} else if def.Code == "RCV-26" {
		query = query.Where("(kas_transactions.expense_category IN ('EXPEDISI', 'EKSPEDISI_OTHER') OR kas_transactions.description ILIKE '%ekspedisi%')")
	} else if def.Code == "RCV-28" {
		query = query.Where("(kas_transactions.expense_category = 'CLAIM_PENALTY' OR kas_transactions.income_source = 'CLAIM_PENALTY' OR kas_transactions.description ILIKE '%klaim%' OR kas_transactions.description ILIKE '%retur%' OR kas_transactions.description ILIKE '%penalty%')")
	} else {
		var orClauses []string
		var args []interface{}
		for _, kw := range def.KW {
			orClauses = append(orClauses, "kas_transactions.description ILIKE ?")
			args = append(args, "%"+kw+"%")
		}
		if len(orClauses) > 0 {
			query = query.Where("("+strings.Join(orClauses, " OR ")+")", args...)
		}
	}

	return query
}

func (r *financeRepository) GetReceivablesBreakdown(month, year int, branchID *uuid.UUID, startDate, endDate string) ([]domain.ReceivableEntityItem, error) {
	branchMap := r.getBranchMap()
	var items []domain.ReceivableEntityItem

	for _, def := range r.receivableEntities() {
		type SumResult struct {
			TransactionType string  `gorm:"column:transaction_type"`
			Total           float64 `gorm:"column:total"`
		}
		var sums []SumResult

		q := r.buildReceivableQuery(def, branchMap, month, year, branchID, startDate, endDate)
		q.Select("kas_transactions.transaction_type, COALESCE(SUM(kas_transactions.amount), 0) as total").
			Group("kas_transactions.transaction_type").
			Scan(&sums)

		var addition, payment float64
		for _, s := range sums {
			if s.TransactionType == "OUT" || s.TransactionType == "TRANSFER_OUT" {
				addition += s.Total
			} else if s.TransactionType == "IN" || s.TransactionType == "TRANSFER_IN" {
				payment += s.Total
			}
		}

		initial := 0.0
		ending := initial + addition - payment
		status := "LUNAS"
		if ending > 0 {
			status = "BERJALAN"
		}

		items = append(items, domain.ReceivableEntityItem{
			Code:           def.Code,
			Name:           def.Name,
			Category:       def.Category,
			InitialBalance: initial,
			PeriodAddition: addition,
			PeriodPayment:  payment,
			EndingBalance:  ending,
			Status:         status,
		})
	}

	return items, nil
}

func (r *financeRepository) GetFinancialReceivablesList(companyID uuid.UUID, branchID *uuid.UUID) (*domain.FinancialReceivablesResponse, error) {
	var branches []domain.Branch

	query := r.db.Model(&domain.Branch{})
	if companyID != uuid.Nil {
		query = query.Where("company_id = ?", companyID)
	}
	if branchID != nil && *branchID != uuid.Nil {
		query = query.Where("id = ?", *branchID)
	}
	if err := query.Order("name ASC").Find(&branches).Error; err != nil {
		return nil, err
	}

	response := &domain.FinancialReceivablesResponse{
		Branches: make([]domain.BranchFinancialSummary, 0),
	}

	for _, b := range branches {
		var txs []domain.KasTransaction
		err := r.db.Preload("KasPabrik").
			Preload("Settlement").
			Preload("RelatedTransaction").
			Where(`
				(target_branch_id = ? AND transaction_type = 'OUT' AND is_branch_receivable = TRUE)
				OR (
					transaction_type = 'IN' AND (
						(target_branch_id = ? AND (income_source IN ('BRANCH_RECEIVABLE_PAYMENT', 'BRANCH_SETTLEMENT', 'BRANCH_DEPOSIT', 'ADVANCE_REFUND', 'TALANGAN_CABANG') OR description ILIKE '%talangan%' OR description ILIKE '%pelunasan%'))
						OR related_transaction_id IN (SELECT id FROM kas_transactions WHERE target_branch_id = ? AND is_branch_receivable = TRUE)
					)
				)
			`, b.ID, b.ID, b.ID).
			Order("transaction_date DESC, created_at DESC").
			Find(&txs).Error
		if err != nil {
			return nil, err
		}

		var totalDisbursed, totalRepaid, totalRepaidMonth float64

		// First pass: Calculate totals
		for _, t := range txs {
			if t.TransactionType == "OUT" {
				totalDisbursed += t.Amount
				// Account for SPJ non-refund settlement if applicable
				if t.Settlement != nil && t.Settlement.ActualExpenseAmount > 0 {
					totalRepaid += t.Settlement.ActualExpenseAmount
				}
			} else if t.TransactionType == "IN" {
				totalRepaid += t.Amount
				if t.TransactionDate.Month() == time.Now().Month() && t.TransactionDate.Year() == time.Now().Year() {
					totalRepaidMonth += t.Amount
				}
			}
		}

		balance := totalDisbursed - totalRepaid
		if balance < 0 {
			balance = 0
		}
		status := "LUNAS"
		if balance > 0 {
			status = "BELUM LUNAS"
		}

		var txItems []domain.BranchFinancialReceivableItem
		for _, t := range txs {
			kasName := ""
			if t.KasPabrik != nil {
				kasName = t.KasPabrik.KasName
			}

			isSettled := false
			if t.TransactionType == "IN" {
				isSettled = true
			} else if t.TransactionType == "OUT" {
				if t.AdvanceStatus == "SETTLED" || t.Settlement != nil || balance <= 0 {
					isSettled = true
				}
			}

			txItems = append(txItems, domain.BranchFinancialReceivableItem{
				ID:                   t.ID,
				KasTransactionID:     t.ID,
				BranchID:             b.ID,
				BranchName:           b.Name,
				KasPabrikID:          t.KasPabrikID,
				KasName:              kasName,
				TransactionType:      string(t.TransactionType),
				Amount:               t.Amount,
				ExpenseCategory:      t.ExpenseCategory,
				IncomeSource:         t.IncomeSource,
				TransactionDate:      t.TransactionDate.Format("2006-01-02"),
				Description:          t.Description,
				ProofURL:             t.ProofURL,
				AdvanceStatus:        t.AdvanceStatus,
				IsSettled:            isSettled,
				Settlement:           t.Settlement,
				RelatedTransactionID: t.RelatedTransactionID,
				CreatedAt:            t.CreatedAt,
			})
		}

		// Include branch in financial summary
		summary := domain.BranchFinancialSummary{
			BranchID:            b.ID,
			BranchName:          b.Name,
			TotalDisbursement:   totalDisbursed,
			TotalRepayment:      totalRepaid,
			OutstandingBalance:  balance,
			TotalRepaymentMonth: totalRepaidMonth,
			Status:              status,
			Transactions:        txItems,
		}

		response.Branches = append(response.Branches, summary)
		response.TotalDisbursed += totalDisbursed
		response.TotalRepaid += totalRepaid
		response.TotalOutstanding += balance
	}

	return response, nil
}

func (r *financeRepository) RecordFinancialReceivablePayment(req domain.RepayBranchFinancialReceivableRequest, recordedBy *uuid.UUID) (*domain.KasTransaction, error) {
	var createdTx *domain.KasTransaction

	err := r.db.Transaction(func(tx *gorm.DB) error {
		var kas domain.KasPabrik
		if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&kas, "id = ?", req.KasPabrikID).Error; err != nil {
			return errors.New("Buku KAS Pabrik tidak ditemukan")
		}

		var branch domain.Branch
		if err := tx.First(&branch, "id = ?", req.BranchID).Error; err != nil {
			return errors.New("Cabang tidak ditemukan")
		}

		txDate := time.Now()
		if req.TransactionDate != "" {
			if parsed, err := time.Parse("2006-01-02", req.TransactionDate); err == nil {
				txDate = parsed
			}
		}

		desc := strings.TrimSpace(req.Description)
		if desc == "" {
			desc = "Pelunasan talangan kas operasional " + branch.Name
		}

		balBefore := kas.Balance
		balAfter := kas.Balance + req.Amount
		kas.Balance = balAfter
		if err := tx.Save(&kas).Error; err != nil {
			return err
		}

		// If TargetTransactionID is provided, mark that OUT transaction as SETTLED
		if req.TargetTransactionID != nil && *req.TargetTransactionID != uuid.Nil {
			var targetTx domain.KasTransaction
			if err := tx.First(&targetTx, "id = ?", *req.TargetTransactionID).Error; err == nil {
				targetTx.AdvanceStatus = "SETTLED"
				_ = tx.Save(&targetTx)
			}
		}

		createdTx = &domain.KasTransaction{
			ID:                   uuid.New(),
			KasPabrikID:          req.KasPabrikID,
			TransactionType:      "IN",
			Amount:               req.Amount,
			BalanceBefore:        balBefore,
			BalanceAfter:         balAfter,
			IncomeSource:         "BRANCH_RECEIVABLE_PAYMENT",
			TargetBranchID:       &req.BranchID,
			IsBranchReceivable:   true,
			RelatedTransactionID: req.TargetTransactionID,
			Description:          desc,
			ProofURL:             req.ProofURL,
			TransactionDate:      txDate,
			CreatedBy:            recordedBy,
			AdvanceStatus:        "NONE",
		}

		if err := tx.Create(createdTx).Error; err != nil {
			return err
		}

		// Recalculate ledger for kas
		var txs []domain.KasTransaction
		if err := tx.Where("kas_pabrik_id = ?", req.KasPabrikID).
			Order("transaction_date ASC, created_at ASC, id ASC").
			Find(&txs).Error; err != nil {
			return err
		}

		runningBalance := 0.0
		for _, t := range txs {
			bBefore := runningBalance
			switch t.TransactionType {
			case "IN", "TRANSFER_IN":
				runningBalance += t.Amount
			case "OUT", "TRANSFER_OUT":
				runningBalance -= t.Amount
			}
			bAfter := runningBalance

			if t.BalanceBefore != bBefore || t.BalanceAfter != bAfter {
				if err := tx.Model(&domain.KasTransaction{}).
					Where("id = ?", t.ID).
					Updates(map[string]interface{}{
						"balance_before": bBefore,
						"balance_after":  bAfter,
					}).Error; err != nil {
					return err
				}
			}
		}

		return tx.Model(&domain.KasPabrik{}).Where("id = ?", req.KasPabrikID).Update("balance", runningBalance).Error
	})

	if err != nil {
		return nil, err
	}
	return createdTx, nil
}

func (r *financeRepository) SettleFinancialReceivableTx(txID uuid.UUID, req domain.SettleFinancialReceivableTxRequest, settledBy *uuid.UUID) (*domain.KasAdvanceSettlement, error) {
	var settlement *domain.KasAdvanceSettlement

	err := r.db.Transaction(func(tx *gorm.DB) error {
		var transaction domain.KasTransaction
		if err := tx.First(&transaction, "id = ?", txID).Error; err != nil {
			return errors.New("Transaksi talangan tidak ditemukan")
		}

		actualAmount := req.ActualExpenseAmount
		if actualAmount <= 0 {
			actualAmount = transaction.Amount
		}

		var refundAmount, shortageAmount float64
		if actualAmount < transaction.Amount {
			refundAmount = transaction.Amount - actualAmount
		} else if actualAmount > transaction.Amount {
			shortageAmount = actualAmount - transaction.Amount
		}

		settlement = &domain.KasAdvanceSettlement{
			ID:                  uuid.New(),
			KasTransactionID:    transaction.ID,
			ActualExpenseAmount: actualAmount,
			RefundAmount:        refundAmount,
			ShortageAmount:      shortageAmount,
			Notes:               req.Notes,
			ProofURL:            req.ProofURL,
			SettledBy:           settledBy,
			SettledAt:           time.Now(),
		}

		if err := tx.Create(settlement).Error; err != nil {
			return err
		}

		transaction.AdvanceStatus = "SETTLED"
		if err := tx.Save(&transaction).Error; err != nil {
			return err
		}

		if refundAmount > 0 {
			var kas domain.KasPabrik
			if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&kas, "id = ?", transaction.KasPabrikID).Error; err != nil {
				return err
			}

			balBefore := kas.Balance
			balAfter := kas.Balance + refundAmount
			kas.Balance = balAfter
			if err := tx.Save(&kas).Error; err != nil {
				return err
			}

			refundTx := domain.KasTransaction{
				ID:                   uuid.New(),
				KasPabrikID:          transaction.KasPabrikID,
				TransactionType:      "IN",
				Amount:               refundAmount,
				BalanceBefore:        balBefore,
				BalanceAfter:         balAfter,
				IncomeSource:         "BRANCH_RECEIVABLE_PAYMENT",
				TargetBranchID:       transaction.TargetBranchID,
				IsBranchReceivable:   true,
				RelatedTransactionID: &transaction.ID,
				Description:          "Pengembalian sisa talangan cabang: " + transaction.Description,
				ProofURL:             req.ProofURL,
				TransactionDate:      time.Now(),
				CreatedBy:            settledBy,
				AdvanceStatus:        "NONE",
			}
			if err := tx.Create(&refundTx).Error; err != nil {
				return err
			}

			// recalculate ledger
			var txs []domain.KasTransaction
			if err := tx.Where("kas_pabrik_id = ?", transaction.KasPabrikID).
				Order("transaction_date ASC, created_at ASC, id ASC").
				Find(&txs).Error; err != nil {
				return err
			}

			runningBalance := 0.0
			for _, t := range txs {
				bBefore := runningBalance
				switch t.TransactionType {
				case "IN", "TRANSFER_IN":
					runningBalance += t.Amount
				case "OUT", "TRANSFER_OUT":
					runningBalance -= t.Amount
				}
				bAfter := runningBalance

				if t.BalanceBefore != bBefore || t.BalanceAfter != bAfter {
					if err := tx.Model(&domain.KasTransaction{}).
						Where("id = ?", t.ID).
						Updates(map[string]interface{}{
							"balance_before": bBefore,
							"balance_after":  bAfter,
						}).Error; err != nil {
						return err
					}
				}
			}

			if err := tx.Model(&domain.KasPabrik{}).Where("id = ?", transaction.KasPabrikID).Update("balance", runningBalance).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}
	return settlement, nil
}

func (r *financeRepository) GetIncomeTransactionDetails(incomeType string, month, year int, branchID *uuid.UUID, startDate, endDate string) ([]domain.IncomeDetailItem, error) {
	var items []domain.IncomeDetailItem
	upperType := strings.ToUpper(strings.TrimSpace(incomeType))

	var branch domain.Branch
	isHeadOfficeOrHasCompany := false
	var compID uuid.UUID
	if branchID != nil && *branchID != uuid.Nil {
		if err := r.db.First(&branch, "id = ?", *branchID).Error; err == nil {
			if branch.CompanyID != nil && *branch.CompanyID != uuid.Nil {
				isHeadOfficeOrHasCompany = true
				compID = *branch.CompanyID
			}
		} else {
			var comp domain.Company
			if err := r.db.First(&comp, "id = ?", *branchID).Error; err == nil {
				isHeadOfficeOrHasCompany = true
				compID = comp.ID
			}
		}
	}

	if upperType == "PRODUCT_SALE" {
		// 1. SalesTransactions (Verified Invoices)
		var salesTxs []domain.SalesTransaction
		txQuery := r.db.Model(&domain.SalesTransaction{}).
			Preload("Store").Preload("BuyerEmployee").
			Where("status = ?", "VERIFIED")

		if startDate != "" && endDate != "" {
			txQuery = txQuery.Where("DATE(transaction_date) >= ? AND DATE(transaction_date) <= ?", startDate, endDate)
		} else {
			txQuery = txQuery.Where("period_month = ? AND period_year = ?", month, year)
		}

		if branchID != nil && *branchID != uuid.Nil {
			if isHeadOfficeOrHasCompany {
				txQuery = txQuery.Where("(branch_id = ? OR (branch_id IS NULL AND company_id = ?))", *branchID, compID)
			} else {
				txQuery = txQuery.Where("branch_id = ?", *branchID)
			}
		}
		if err := txQuery.Order("transaction_date DESC, created_at DESC").Find(&salesTxs).Error; err == nil {
			for _, st := range salesTxs {
				entityName := "Penjualan Umum / Pabrik"
				if st.Store != nil && st.Store.Name != "" {
					entityName = st.Store.Name
				} else if st.BuyerName != "" {
					entityName = st.BuyerName
				} else if st.BuyerEmployee != nil {
					entityName = st.BuyerEmployee.FirstName + " " + st.BuyerEmployee.LastName
				}

				refNo := st.ReceiptNo
				if refNo == "" {
					refNo = "TRX-" + st.ID.String()[:8]
				}

				desc := ""
				if st.Notes != nil {
					desc = *st.Notes
				}

				items = append(items, domain.IncomeDetailItem{
					ID:              st.ID,
					TransactionType: "SALES_INVOICE",
					ReferenceNo:     refNo,
					TransactionDate: st.TransactionDate.Format("2006-01-02"),
					EntityName:      entityName,
					PaymentMethod:   st.PaymentMethod,
					Description:     desc,
					Amount:          st.TotalAmount,
					Status:          st.Status,
					ProofURL:        st.ReceiptImageURL,
				})
			}
		}

		// 2. SalesOrders not yet invoiced (invoice_id IS NULL)
		var salesOrders []domain.SalesOrder
		soQuery := r.db.Model(&domain.SalesOrder{}).
			Preload("Store").
			Where("status IN ? AND invoice_id IS NULL", []string{string(domain.SOStatusDelivered), string(domain.SOStatusPaid), string(domain.SOStatusShipped)})

		if startDate != "" && endDate != "" {
			soQuery = soQuery.Where("DATE(order_date) >= ? AND DATE(order_date) <= ?", startDate, endDate)
		} else {
			soQuery = soQuery.Where("EXTRACT(MONTH FROM order_date) = ? AND EXTRACT(YEAR FROM order_date) = ?", month, year)
		}

		if branchID != nil && *branchID != uuid.Nil {
			if isHeadOfficeOrHasCompany {
				soQuery = soQuery.Where("(branch_id = ? OR (branch_id IS NULL AND company_id = ?))", *branchID, compID)
			} else {
				soQuery = soQuery.Where("branch_id = ?", *branchID)
			}
		}
		if err := soQuery.Order("order_date DESC, created_at DESC").Find(&salesOrders).Error; err == nil {
			for _, so := range salesOrders {
				entityName := "Pesanan Toko / Direct"
				if so.Store.Name != "" {
					entityName = so.Store.Name
				}

				refNo := so.SONumber
				if refNo == "" {
					refNo = "SO-" + so.ID.String()[:8]
				}

				items = append(items, domain.IncomeDetailItem{
					ID:              so.ID,
					TransactionType: "SALES_ORDER",
					ReferenceNo:     refNo,
					TransactionDate: so.OrderDate.Format("2006-01-02"),
					EntityName:      entityName,
					PaymentMethod:   "TERKIRIM (BELUM INVOICE)",
					Description:     so.Notes,
					Amount:          so.TotalAmount,
					Status:          string(so.Status),
				})
			}
		}

		// 3. Kas Transactions with PRODUCT_SALE
		var kasTxs []domain.KasTransaction
		kasQuery := r.db.Model(&domain.KasTransaction{}).
			Preload("KasPabrik").
			Joins("LEFT JOIN kas_income_sources kis ON kis.id = kas_transactions.income_source_id").
			Where("kas_transactions.transaction_type = 'IN'").
			Where("kis.code = 'PRODUCT_SALE' OR UPPER(kas_transactions.income_source) LIKE '%PRODUCT%' OR UPPER(kas_transactions.income_source) LIKE '%PRODUK%' OR UPPER(kas_transactions.income_source) LIKE '%TELEMARKETING%' OR UPPER(kas_transactions.income_source) LIKE '%FAKTUR%'")

		if startDate != "" && endDate != "" {
			kasQuery = kasQuery.Where("DATE(kas_transactions.transaction_date) >= ? AND DATE(kas_transactions.transaction_date) <= ?", startDate, endDate)
		} else {
			kasQuery = kasQuery.Where("EXTRACT(MONTH FROM kas_transactions.transaction_date) = ? AND EXTRACT(YEAR FROM kas_transactions.transaction_date) = ?", month, year)
		}

		if branchID != nil && *branchID != uuid.Nil {
			kasQuery = kasQuery.Joins("JOIN kas_pabrik kp ON kp.id = kas_transactions.kas_pabrik_id")
			if isHeadOfficeOrHasCompany {
				kasQuery = kasQuery.Where("(kp.branch_id = ? OR kp.branch_id IS NULL)", *branchID)
			} else {
				kasQuery = kasQuery.Where("kp.branch_id = ?", *branchID)
			}
		}

		if err := kasQuery.Order("kas_transactions.transaction_date DESC, kas_transactions.created_at DESC").Find(&kasTxs).Error; err == nil {
			for _, kt := range kasTxs {
				kasName := "Kas Masuk Penjualan"
				if kt.KasPabrik != nil && kt.KasPabrik.KasName != "" {
					kasName = kt.KasPabrik.KasName
				}

				refNo := "KAS-IN-" + kt.ID.String()[:8]
				items = append(items, domain.IncomeDetailItem{
					ID:              kt.ID,
					TransactionType: "KAS_IN",
					ReferenceNo:     refNo,
					TransactionDate: kt.TransactionDate.Format("2006-01-02"),
					EntityName:      kasName,
					PaymentMethod:   "KAS / TUNAI",
					Description:     kt.Description,
					Amount:          kt.Amount,
					Status:          "CONFIRMED",
					ProofURL:        kt.ProofURL,
				})
			}
		}

		return items, nil
	}

	// Kas transactions for other income types
	var kasTxs []domain.KasTransaction
	query := r.db.Model(&domain.KasTransaction{}).
		Preload("KasPabrik").
		Joins("LEFT JOIN kas_income_sources kis ON kis.id = kas_transactions.income_source_id").
		Where("kas_transactions.transaction_type = 'IN'")

	if startDate != "" && endDate != "" {
		query = query.Where("DATE(kas_transactions.transaction_date) >= ? AND DATE(kas_transactions.transaction_date) <= ?", startDate, endDate)
	} else {
		query = query.Where("EXTRACT(MONTH FROM kas_transactions.transaction_date) = ? AND EXTRACT(YEAR FROM kas_transactions.transaction_date) = ?", month, year)
	}

	if branchID != nil && *branchID != uuid.Nil {
		query = query.Joins("JOIN kas_pabrik kp ON kp.id = kas_transactions.kas_pabrik_id")
		if isHeadOfficeOrHasCompany {
			query = query.Where("(kp.branch_id = ? OR kp.branch_id IS NULL)", *branchID)
		} else {
			query = query.Where("kp.branch_id = ?", *branchID)
		}
	}

	switch upperType {
	case "AMPAS_SALE":
		query = query.Where("kis.code = 'AMPAS_SALE' OR UPPER(kas_transactions.income_source) LIKE '%AMPAS%'")
	case "ROSOK_SALE":
		query = query.Where("kis.code = 'ROSOK_SALE' OR UPPER(kas_transactions.income_source) LIKE '%ROSOK%' OR UPPER(kas_transactions.income_source) LIKE '%AFVAL%' OR UPPER(kas_transactions.income_source) LIKE '%JURIGEN%' OR UPPER(kas_transactions.income_source) LIKE '%DRUM%'")
	case "INTEREST_INCOME":
		query = query.Where("kis.code = 'INTEREST_INCOME' OR UPPER(kas_transactions.income_source) LIKE '%INTEREST%' OR UPPER(kas_transactions.income_source) LIKE '%BUNGA%'")
	case "CLAIM_PENALTY":
		query = query.Where("kis.code = 'CLAIM_PENALTY' OR UPPER(kas_transactions.income_source) LIKE '%CLAIM%' OR UPPER(kas_transactions.income_source) LIKE '%KLAIM%' OR UPPER(kas_transactions.income_source) LIKE '%PENALTY%' OR UPPER(kas_transactions.income_source) LIKE '%DENDA%'")
	case "OTHER":
		query = query.Where("(kis.code = 'OTHER' OR kas_transactions.income_source = 'OTHER' OR (kis.code IS NULL AND kas_transactions.income_source NOT ILIKE '%AMPAS%' AND kas_transactions.income_source NOT ILIKE '%ROSOK%' AND kas_transactions.income_source NOT ILIKE '%AFVAL%' AND kas_transactions.income_source NOT ILIKE '%BUNGA%' AND kas_transactions.income_source NOT ILIKE '%CLAIM%' AND kas_transactions.income_source NOT ILIKE '%PRODUK%' AND kas_transactions.income_source NOT ILIKE '%FAKTUR%'))")
	}

	if err := query.Order("kas_transactions.transaction_date DESC, kas_transactions.created_at DESC").Find(&kasTxs).Error; err != nil {
		return nil, err
	}

	for _, kt := range kasTxs {
		kasName := "Kas Pabrik"
		if kt.KasPabrik != nil && kt.KasPabrik.KasName != "" {
			kasName = kt.KasPabrik.KasName
		}

		refNo := "KAS-IN-" + kt.ID.String()[:8]
		items = append(items, domain.IncomeDetailItem{
			ID:              kt.ID,
			TransactionType: "KAS_IN",
			ReferenceNo:     refNo,
			TransactionDate: kt.TransactionDate.Format("2006-01-02"),
			EntityName:      kasName,
			Description:     kt.Description,
			Amount:          kt.Amount,
			Status:          "CONFIRMED",
			ProofURL:        kt.ProofURL,
		})
	}

	return items, nil
}

func (r *financeRepository) GetExpenseTransactionDetails(categoryKey string, month, year int, branchID *uuid.UUID, startDate, endDate string) ([]domain.ExpenseDetailItem, error) {
	var kasTxs []domain.KasTransaction
	query := r.db.Model(&domain.KasTransaction{}).
		Preload("KasPabrik").
		Preload("TargetBranch").
		Preload("Vehicle").
		Preload("Employee").
		Preload("ExpenseTypeObj").
		Preload("ExpenseAccountObj").
		Where("kas_transactions.transaction_type = 'OUT'")

	if startDate != "" && endDate != "" {
		query = query.Where("DATE(kas_transactions.transaction_date) >= ? AND DATE(kas_transactions.transaction_date) <= ?", startDate, endDate)
	} else {
		query = query.Where("EXTRACT(MONTH FROM kas_transactions.transaction_date) = ? AND EXTRACT(YEAR FROM kas_transactions.transaction_date) = ?", month, year)
	}

	if branchID != nil && *branchID != uuid.Nil {
		query = query.Joins("JOIN kas_pabrik kp ON kp.id = kas_transactions.kas_pabrik_id").
			Where("kp.branch_id = ?", *branchID)
	}

	if categoryKey != "" && categoryKey != "ALL" {
		query = query.Where("kas_transactions.expense_category = ? OR (kas_transactions.expense_account_id IN (SELECT id FROM kas_expense_accounts WHERE account_code = ?))", categoryKey, categoryKey)
	}

	if err := query.Order("kas_transactions.transaction_date DESC, kas_transactions.created_at DESC").Find(&kasTxs).Error; err != nil {
		return nil, err
	}

	var items []domain.ExpenseDetailItem
	for _, tx := range kasTxs {
		kasName := "Kas Pabrik"
		if tx.KasPabrik != nil && tx.KasPabrik.KasName != "" {
			kasName = tx.KasPabrik.KasName
		}

		catName := tx.ExpenseCategory
		if tx.ExpenseAccountObj != nil && tx.ExpenseAccountObj.AccountName != "" {
			catName = tx.ExpenseAccountObj.AccountName
		}

		recipient := ""
		if tx.TargetBranch != nil && tx.TargetBranch.Name != "" {
			recipient = "Cabang: " + tx.TargetBranch.Name
		} else if tx.Employee != nil {
			recipient = "Karyawan: " + tx.Employee.FirstName + " " + tx.Employee.LastName
		} else if tx.Vehicle != nil {
			recipient = "Armada: " + tx.Vehicle.LicensePlate + " (" + tx.Vehicle.Name + ")"
		}

		status := "CONFIRMED"
		if tx.AdvanceStatus != "" && tx.AdvanceStatus != "NONE" {
			status = tx.AdvanceStatus
		}

		refNo := "KAS-OUT-" + tx.ID.String()[:8]
		items = append(items, domain.ExpenseDetailItem{
			ID:              tx.ID,
			ReferenceNo:     refNo,
			TransactionDate: tx.TransactionDate.Format("2006-01-02"),
			KasName:         kasName,
			CategoryKey:     tx.ExpenseCategory,
			CategoryName:    catName,
			RecipientName:   recipient,
			Description:     tx.Description,
			Amount:          tx.Amount,
			Status:          status,
			ProofURL:        tx.ProofURL,
		})
	}

	return items, nil
}

func (r *financeRepository) GetReceivableTransactionDetails(entityCode string, month, year int, branchID *uuid.UUID, startDate, endDate string) (*domain.ReceivableDetailResponse, error) {
	upperCode := strings.ToUpper(strings.TrimSpace(entityCode))
	var targetDef *receivableDef
	for _, def := range r.receivableEntities() {
		if strings.EqualFold(def.Code, upperCode) {
			targetDef = &def
			break
		}
	}

	if targetDef == nil {
		targetDef = &receivableDef{
			Code:     upperCode,
			Name:     "Entitas Piutang " + upperCode,
			Category: "PROYEK",
			KW:       []string{upperCode},
		}
	}

	branchMap := r.getBranchMap()
	q := r.buildReceivableQuery(*targetDef, branchMap, month, year, branchID, startDate, endDate)

	var txs []domain.KasTransaction
	if err := q.Preload("KasPabrik").
		Preload("TargetBranch").
		Preload("Employee").
		Order("kas_transactions.transaction_date DESC, kas_transactions.created_at DESC").
		Find(&txs).Error; err != nil {
		return nil, err
	}

	var items []domain.ReceivableDetailItem
	var totalAddition, totalPayment float64

	for _, tx := range txs {
		kasName := "Kas Pabrik"
		if tx.KasPabrik != nil && tx.KasPabrik.KasName != "" {
			kasName = tx.KasPabrik.KasName
		}

		txType := "DEBET"
		if tx.TransactionType == "OUT" || tx.TransactionType == "TRANSFER_OUT" {
			txType = "DEBET"
			totalAddition += tx.Amount
		} else {
			txType = "KREDIT"
			totalPayment += tx.Amount
		}

		prefix := "KAS-OUT-"
		if txType == "KREDIT" {
			prefix = "KAS-IN-"
		}
		refNo := prefix + tx.ID.String()[:8]

		status := "CONFIRMED"
		if tx.AdvanceStatus != "" && tx.AdvanceStatus != "NONE" {
			status = tx.AdvanceStatus
		}

		items = append(items, domain.ReceivableDetailItem{
			ID:              tx.ID,
			TransactionType: txType,
			ReferenceNo:     refNo,
			TransactionDate: tx.TransactionDate.Format("2006-01-02"),
			KasName:         kasName,
			Description:     tx.Description,
			Amount:          tx.Amount,
			Status:          status,
			ProofURL:        tx.ProofURL,
		})
	}

	initialBalance := 0.0
	endingBalance := initialBalance + totalAddition - totalPayment
	status := "LUNAS"
	if endingBalance > 0 {
		status = "BERJALAN"
	}

	return &domain.ReceivableDetailResponse{
		EntityCode:     targetDef.Code,
		EntityName:     targetDef.Name,
		Category:       targetDef.Category,
		PeriodMonth:    month,
		PeriodYear:     year,
		StartDate:      startDate,
		EndDate:        endDate,
		BranchID:       branchID,
		InitialBalance: initialBalance,
		PeriodAddition: totalAddition,
		PeriodPayment:  totalPayment,
		EndingBalance:  endingBalance,
		Status:         status,
		TotalCount:     len(items),
		Items:          items,
	}, nil
}
