package repository

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"example.com/hris/backend/pkg/utils"
	"gorm.io/gorm"
)

type SalesTransactionRepository interface {
	Create(transaction *domain.SalesTransaction) error
	FindAll() ([]domain.SalesTransaction, error)
	FindAllPaginated(page, limit int, companyID, status, search, salesman, month, year, date, salesType, branchID, paymentMethod, paymentStatus string, targetEntity ...string) ([]domain.SalesTransaction, int64, float64, error)
	FindByID(id uuid.UUID) (*domain.SalesTransaction, error)
	Update(transaction *domain.SalesTransaction) error
	UpdateTransactionOnly(transaction *domain.SalesTransaction) error
	Delete(id uuid.UUID) error
	SoftDeleteWithReason(id uuid.UUID, deletedByID uuid.UUID, reason string) error
	Restore(id uuid.UUID) error
	FindByIDUnscoped(id uuid.UUID) (*domain.SalesTransaction, error)
	GetTransactionsByEmployeeAndPeriod(employeeID uuid.UUID, month, year int) ([]domain.SalesTransaction, error)
	FindByReceiptNo(receiptNo string) (*domain.SalesTransaction, error)
	CreatePayment(payment *domain.SalesPayment) error
	FindPendingPaymentByTrxAndAmount(trxID uuid.UUID, amount float64, out *domain.SalesPayment) error
	FindPaymentsByEmployee(employeeID uuid.UUID) ([]domain.SalesPayment, error)
	GetOutstandingByStore(storeID uuid.UUID) ([]domain.SalesTransaction, error)
	GetByDueDate(date string, employeeID uuid.UUID) ([]domain.SalesTransaction, error)
	FindPendingByEmployeeID(employeeID uuid.UUID) ([]domain.SalesTransaction, error)
	FindByStatusAndCompany(status string, companyID uuid.UUID) ([]domain.SalesTransaction, error)
	FindDeliveryPending(companyID uuid.UUID) ([]domain.SalesTransaction, error)
	CountByCompany(companyID uuid.UUID) (int64, error)
	CountByCompanyAndDate(companyID uuid.UUID, date time.Time) (int64, error)
	GetProductSalesDistribution(productIDs []uuid.UUID, companyID uuid.UUID) ([]map[string]interface{}, error)
	GetTopProducts(companyID uuid.UUID, limit int) ([]map[string]interface{}, error)
	GetSalesKasbon(companyID uuid.UUID) ([]map[string]interface{}, error)
	UpdateInstallment(installment *domain.SalesInstallment) error
	GetSettlementQueue() ([]domain.SalesTransaction, error)
	GetTotalCashValidatedByAdminNota(ctx context.Context, branchID uuid.UUID, start time.Time, end time.Time) (float64, error)
	GetProductRecapByEmployeeAndPeriod(employeeID uuid.UUID, month, year int) ([]domain.ProductRecap, error)
	GetDB() *gorm.DB
}

type salesTransactionRepository struct {
	db *gorm.DB
}

func NewSalesTransactionRepository(db *gorm.DB) SalesTransactionRepository {
	return &salesTransactionRepository{db}
}

func (r *salesTransactionRepository) GetDB() *gorm.DB {
	return r.db
}

func (r *salesTransactionRepository) GetTotalCashValidatedByAdminNota(ctx context.Context, branchID uuid.UUID, start time.Time, end time.Time) (float64, error) {
	var total float64
	err := r.db.WithContext(ctx).
		Model(&domain.SalesTransaction{}).
		Joins("JOIN sales_orders ON sales_orders.invoice_id = sales_transactions.id").
		Where("sales_orders.branch_id = ?", branchID).
		Where("sales_transactions.payment_method = ?", "CASH").
		Where("sales_orders.admin_nota_confirmed_at >= ? AND sales_orders.admin_nota_confirmed_at < ?", start, end).
		Select("COALESCE(SUM(sales_transactions.total_amount), 0)").
		Scan(&total).Error

	return total, err
}

func (r *salesTransactionRepository) GetSettlementQueue() ([]domain.SalesTransaction, error) {
	var transactions []domain.SalesTransaction
	err := r.db.Preload("Store").Preload("Employee").
		Where("status IN ?", []string{"CASHIER_APPROVED", "SETTLED"}).
		Order("transaction_date DESC").
		Find(&transactions).Error
	return transactions, err
}

func (r *salesTransactionRepository) Create(transaction *domain.SalesTransaction) error {
	return r.db.Create(transaction).Error
}

func (r *salesTransactionRepository) FindAll() ([]domain.SalesTransaction, error) {
	var transactions []domain.SalesTransaction
	if err := r.db.Preload("Company").Preload("Store").Preload("Employee").
		Preload("Employee.Branch").Preload("Employee.OperatingBranch").
		Preload("Employee.Company").Preload("Employee.JobPosition").
		Preload("VerifiedBy").Preload("VerifiedBy.JobPosition").
		Preload("DeletedBy").Preload("DeletedBy.JobPosition").
		Preload("DeliveryItems").Preload("DeliveryItems.DeliveryBatch").Preload("DeliveryItems.DeliveryBatch.Driver").
		Preload("Visit").
		Preload("Items").Preload("Items.Product").Preload("BonusItems").Preload("BonusItems.BonusInventory").
		Order("created_at desc").Find(&transactions).Error; err != nil {
		return nil, err
	}

	ptrTransactions := make([]*domain.SalesTransaction, len(transactions))
	for i := range transactions {
		ptrTransactions[i] = &transactions[i]
	}
	r.injectBonusProducts(ptrTransactions)

	return transactions, nil
}

var branchAcronymAliases = map[string][]string{
	"sbt":            {"SANKE BERSINAR TERANG", "SBT"},
	"sls":            {"SHOWA LANCAR SEJAHTERA", "SLS"},
	"pas":            {"PREDATOR ALAM SEMESTA", "PAS"},
	"mob":            {"MUKASI OGON BERKILAU", "MOB"},
	"mui":            {"MURAMASA UNIVERSAL", "MUI"},
	"cbi":            {"CUAN BESAR INDONESIA", "CBI"},
	"cbi_kudus":      {"CUAN BESAR INDONESIA KUDUS", "CBI KUDUS"},
	"cbi_serang":     {"CUAN BESAR INDONESIA CABANG SERANG", "CBI SERANG"},
	"cbi_bandung":    {"CUAN BESAR INDONESIA BANDUNG", "CBI BANDUNG"},
	"cbi_bogor":      {"CUAN BESAR INDONESIA CABANG BOGOR", "CBI BOGOR"},
	"latubo":         {"BOOSTER LATUBO", "LATUBO"},
	"booster_latubo": {"BOOSTER LATUBO", "LATUBO"},
	"sby":            {"BOOSTER SURABAYA", "SURABAYA"},
	"surabaya":       {"BOOSTER SURABAYA", "SURABAYA"},
	"booster_sby":    {"BOOSTER SURABAYA"},
	"ksa":            {"KOHAKU SUMBER ABADI", "KSA"},
	"bks":            {"BKS"},
	"gja":            {"GJA"},
}

func (r *salesTransactionRepository) FindAllPaginated(page, limit int, companyID, status, search, salesman, month, year, date, salesType, branchID, paymentMethod, paymentStatus string, targetEntity ...string) ([]domain.SalesTransaction, int64, float64, error) {
	var transactions []domain.SalesTransaction
	var total int64
	var totalAmount float64

	var query *gorm.DB
	if status == "DELETED" || status == "TRASH" {
		query = r.db.Unscoped().Model(&domain.SalesTransaction{}).Where("sales_transactions.deleted_at IS NOT NULL")
	} else {
		query = r.db.Model(&domain.SalesTransaction{})
	}
	// An invoice kept for redelivery belongs to its SO, but it is not an active
	// sales invoice until that SO has been put on a new delivery note.
	query = query.Where(`NOT EXISTS (
		SELECT 1 FROM sales_orders so
		WHERE so.invoice_id = sales_transactions.id
		  AND so.deleted_at IS NULL
		  AND so.status = ?
		  AND so.delivery_batch_id IS NULL
	)`, domain.SOStatusConfirmed)

	if salesType == "sales_mt" {
		query = query.Joins("LEFT JOIN employees emt ON emt.id = sales_transactions.employee_id").
			Joins("LEFT JOIN job_positions jpmt ON jpmt.id = emt.job_position_id").
			Where("jpmt.title ILIKE '%mt%' OR jpmt.title ILIKE '%modern trade%' OR jpmt.title ILIKE '%kam%' OR jpmt.title ILIKE '%key account%'")
	} else if salesType == "sales_to" {
		query = query.Joins("LEFT JOIN employees eto ON eto.id = sales_transactions.employee_id").
			Joins("LEFT JOIN job_positions jpto ON jpto.id = eto.job_position_id").
			Where("jpto.title ~* '(^|[^[:alnum:]])(to|take order|task order)([^[:alnum:]]|$)' AND (sales_transactions.sales_channel IS NULL OR sales_transactions.sales_channel = 'CANVAS')")
	} else if salesType == "sales_motoris_booster" {
		query = query.Joins("LEFT JOIN employees emb ON emb.id = sales_transactions.employee_id").
			Joins("LEFT JOIN job_positions jpmb ON jpmb.id = emb.job_position_id").
			Where("jpmb.title ILIKE '%motoris%' OR jpmb.title ILIKE '%booster%'")
	} else if salesType == "direct_sale" || salesType == "direct_factory" || salesType == "DIRECT_FACTORY" {
		query = query.Where("(sales_transactions.sales_channel IN ('DIRECT_FACTORY', 'DIRECT_BRANCH') OR sales_transactions.receipt_no LIKE 'FP-PBR%' OR sales_transactions.receipt_no LIKE 'FP-CAB%' OR sales_transactions.receipt_no LIKE 'FP-DIR%')")
	} else if salesType == "regular_sales" || salesType == "REGULAR_SALES" {
		query = query.Where("(sales_transactions.sales_channel IS NULL OR sales_transactions.sales_channel IN ('CANVAS', 'MOTORIS') OR (sales_transactions.sales_channel NOT IN ('DIRECT_FACTORY', 'DIRECT_BRANCH', 'EMPLOYEE_SALE') AND sales_transactions.receipt_no NOT LIKE 'FP-PBR%' AND sales_transactions.receipt_no NOT LIKE 'FP-CAB%'))")
	} else if salesType == "employee_sale" || salesType == "EMPLOYEE_SALE" {
		query = query.Where("sales_transactions.sales_channel = 'EMPLOYEE_SALE' OR sales_transactions.buyer_type = 'EMPLOYEE'")
	}

	if companyID != "" {
		query = query.Where("sales_transactions.company_id = ?", companyID)
	}

	if branchID != "" && branchID != "ALL_BRANCHES" && branchID != "ALL" {
		branchUUIDs := utils.ParseUUIDs(branchID)
		if len(branchUUIDs) > 1 {
			query = query.
				Joins("LEFT JOIN employees emp_branch_filter ON emp_branch_filter.id = sales_transactions.employee_id").
				Joins("LEFT JOIN stores store_branch_filter ON store_branch_filter.id = sales_transactions.store_id").
				Joins("LEFT JOIN employees store_emp_branch_filter ON store_emp_branch_filter.id = store_branch_filter.assigned_employee_id").
				Where("COALESCE(emp_branch_filter.operating_branch_id, emp_branch_filter.branch_id, store_emp_branch_filter.operating_branch_id, store_emp_branch_filter.branch_id, sales_transactions.branch_id) IN (?)", branchUUIDs)
		} else {
			query = query.
				Joins("LEFT JOIN employees emp_branch_filter ON emp_branch_filter.id = sales_transactions.employee_id").
				Joins("LEFT JOIN stores store_branch_filter ON store_branch_filter.id = sales_transactions.store_id").
				Joins("LEFT JOIN employees store_emp_branch_filter ON store_emp_branch_filter.id = store_branch_filter.assigned_employee_id").
				Where("COALESCE(emp_branch_filter.operating_branch_id, emp_branch_filter.branch_id, store_emp_branch_filter.operating_branch_id, store_emp_branch_filter.branch_id, sales_transactions.branch_id) = ?", branchID)
		}
	}
	if status != "" && status != "ALL" && status != "DELETED" && status != "TRASH" {
		// TEMPO_UNPAID: Faktur tempo yang masih memiliki sisa piutang / belum lunas
		if status == "TEMPO_UNPAID" {
			query = query.Where("sales_transactions.payment_method = 'TEMPO' AND sales_transactions.payment_status IN ('UNPAID', 'PARTIAL')")
		} else if status == "APPROVED" {
			// Penerimaan menggunakan status APPROVED sebagai status bisnis: faktur
			// sudah dikonfirmasi Admin Nota dan pembayarannya sudah disetujui kasir.
			// Approval tersebut disimpan di sales_orders/sales_payments, bukan selalu
			// menulis sales_transactions.status = 'APPROVED'.
			query = query.Where(`
				sales_transactions.status = 'APPROVED'
				OR (
					sales_transactions.status = 'PENDING'
					AND sales_transactions.payment_status = 'PAID'
					AND EXISTS (
						SELECT 1 FROM sales_orders so
						WHERE so.invoice_id = sales_transactions.id
						  AND so.admin_nota_confirmed_at IS NOT NULL
					)
				)`)
		} else {
			query = query.Where("sales_transactions.status = ?", status)
		}
	}

	hasJoinedStores := false
	joinStoresEmployees := func() {
		if !hasJoinedStores {
			query = query.Joins("LEFT JOIN stores ON stores.id = sales_transactions.store_id").
				Joins("LEFT JOIN employees ON employees.id = sales_transactions.employee_id").
				Joins("LEFT JOIN employees buyer_emp ON buyer_emp.id = sales_transactions.buyer_employee_id")
			hasJoinedStores = true
		}
	}

	var targetEntityVal string
	if len(targetEntity) > 0 && targetEntity[0] != "" && targetEntity[0] != "ALL" {
		targetEntityVal = strings.TrimSpace(targetEntity[0])
	}

	if targetEntityVal != "" {
		joinStoresEmployees()
		if targetUUID, err := uuid.Parse(targetEntityVal); err == nil && targetUUID != uuid.Nil {
			query = query.Where("(sales_transactions.store_id = ? OR sales_transactions.branch_id = ?)", targetUUID, targetUUID)
		} else {
			cleanKey := strings.ToLower(strings.TrimSpace(targetEntityVal))
			if aliases, ok := branchAcronymAliases[cleanKey]; ok {
				var conds []string
				var args []interface{}
				for _, alias := range aliases {
					conds = append(conds, "stores.name ILIKE ?", "sales_transactions.buyer_name ILIKE ?")
					args = append(args, "%"+alias+"%", "%"+alias+"%")
				}
				query = query.Where("("+strings.Join(conds, " OR ")+")", args...)
			} else {
				query = query.Where("(stores.name ILIKE ? OR sales_transactions.buyer_name ILIKE ?)", "%"+targetEntityVal+"%", "%"+targetEntityVal+"%")
			}
		}
	}

	if search != "" {
		joinStoresEmployees()
		lowerSearch := strings.ToLower(strings.TrimSpace(search))
		conds := []string{
			"sales_transactions.receipt_no ILIKE ?",
			"stores.name ILIKE ?",
			"stores.owner_name ILIKE ?",
			"sales_transactions.buyer_name ILIKE ?",
			"employees.first_name ILIKE ?",
			"employees.last_name ILIKE ?",
			"CONCAT(employees.first_name, ' ', COALESCE(employees.last_name, '')) ILIKE ?",
			"buyer_emp.first_name ILIKE ?",
			"buyer_emp.last_name ILIKE ?",
			"CONCAT(buyer_emp.first_name, ' ', COALESCE(buyer_emp.last_name, '')) ILIKE ?",
		}
		searchTerm := "%" + search + "%"
		args := []interface{}{
			searchTerm, searchTerm, searchTerm, searchTerm, searchTerm,
			searchTerm, searchTerm, searchTerm, searchTerm, searchTerm,
		}

		// Smart Acronym matching (e.g. searching "SBT" also matches "PT SANKE BERSINAR TERANG")
		for acronym, aliases := range branchAcronymAliases {
			if strings.Contains(lowerSearch, acronym) || lowerSearch == acronym {
				for _, alias := range aliases {
					conds = append(conds, "stores.name ILIKE ?", "sales_transactions.buyer_name ILIKE ?")
					args = append(args, "%"+alias+"%", "%"+alias+"%")
				}
			}
		}

		query = query.Where("("+strings.Join(conds, " OR ")+")", args...)
	}
	if salesman != "" {
		if empUUID, err := uuid.Parse(salesman); err == nil && empUUID != uuid.Nil {
			query = query.Where("sales_transactions.employee_id = ?", empUUID)
		} else {
			query = query.Joins("LEFT JOIN employees e2 ON e2.id = sales_transactions.employee_id").
				Where("e2.first_name || ' ' || e2.last_name = ?", salesman)
		}
	}
	if month != "" && month != "ALL" {
		if strings.Contains(month, "-") {
			if parsedMonth, err := time.Parse("2006-01", strings.TrimSpace(month)); err == nil {
				startMonth := parsedMonth.Format("2006-01-02")
				nextMonth := parsedMonth.AddDate(0, 1, 0).Format("2006-01-02")
				query = query.Where("sales_transactions.transaction_date >= ? AND sales_transactions.transaction_date < ?", startMonth, nextMonth)
			} else {
				parts := strings.Split(month, "-")
				if len(parts) == 2 {
					query = query.Where("EXTRACT(YEAR FROM sales_transactions.transaction_date) = ? AND EXTRACT(MONTH FROM sales_transactions.transaction_date) = ?", parts[0], parts[1])
				}
			}
		} else if year != "" && year != "ALL" {
			if parsedMY, err := time.Parse("2006-1", strings.TrimSpace(year)+"-"+strings.TrimSpace(month)); err == nil {
				startMonth := parsedMY.Format("2006-01-02")
				nextMonth := parsedMY.AddDate(0, 1, 0).Format("2006-01-02")
				query = query.Where("sales_transactions.transaction_date >= ? AND sales_transactions.transaction_date < ?", startMonth, nextMonth)
			} else {
				query = query.Where("EXTRACT(YEAR FROM sales_transactions.transaction_date) = ? AND EXTRACT(MONTH FROM sales_transactions.transaction_date) = ?", year, month)
			}
		} else {
			query = query.Where("EXTRACT(MONTH FROM sales_transactions.transaction_date) = ?", month)
		}
	} else if year != "" && year != "ALL" {
		if parsedYear, err := time.Parse("2006", strings.TrimSpace(year)); err == nil {
			startYear := parsedYear.Format("2006-01-02")
			nextYear := parsedYear.AddDate(1, 0, 0).Format("2006-01-02")
			query = query.Where("sales_transactions.transaction_date >= ? AND sales_transactions.transaction_date < ?", startYear, nextYear)
		} else {
			query = query.Where("EXTRACT(YEAR FROM sales_transactions.transaction_date) = ?", year)
		}
	}
	if date != "" {
		if parsedDate, err := time.Parse("2006-01-02", strings.TrimSpace(date)); err == nil {
			startDay := parsedDate.Format("2006-01-02")
			nextDay := parsedDate.AddDate(0, 0, 1).Format("2006-01-02")
			query = query.Where("sales_transactions.transaction_date >= ? AND sales_transactions.transaction_date < ?", startDay, nextDay)
		} else {
			query = query.Where("DATE(sales_transactions.transaction_date) = ?", date)
		}
	}

	// Filter Metode Pembayaran
	if paymentMethod != "" && paymentMethod != "ALL" {
		switch paymentMethod {
		case "CASH":
			query = query.Where("sales_transactions.payment_method = 'CASH'")
		case "TRANSFER":
			query = query.Where("sales_transactions.payment_method = 'TRANSFER'")
		case "TEMPO":
			query = query.Where("sales_transactions.payment_method = 'TEMPO'")
		case "TEMPO_CASH":
			query = query.Where(`sales_transactions.payment_method = 'TEMPO' AND (
				EXISTS (
					SELECT 1 FROM sales_payments sp 
					WHERE sp.sales_transaction_id = sales_transactions.id 
					  AND sp.payment_method = 'CASH' 
					  AND sp.deleted_at IS NULL
				)
				OR sales_transactions.dp_payment_method = 'CASH'
				OR (
					sales_transactions.dp_payment_method IS NULL 
					AND NOT EXISTS (
						SELECT 1 FROM sales_payments sp2 
						WHERE sp2.sales_transaction_id = sales_transactions.id 
						  AND sp2.payment_method != 'CASH' 
						  AND sp2.deleted_at IS NULL
					)
				)
			)`)
		case "TEMPO_TRANSFER":
			query = query.Where(`sales_transactions.payment_method = 'TEMPO' AND (
				EXISTS (
					SELECT 1 FROM sales_payments sp 
					WHERE sp.sales_transaction_id = sales_transactions.id 
					  AND (sp.payment_method = 'TRANSFER' OR sp.payment_method = 'VA') 
					  AND sp.deleted_at IS NULL
				)
				OR sales_transactions.dp_payment_method = 'TRANSFER'
				OR sales_transactions.dp_payment_method = 'VA'
			)`)
		case "QRIS":
			query = query.Where("sales_transactions.payment_method IN ('QRIS', 'MIDTRANS_QRIS')")
		case "POTONG_GAJI":
			query = query.Where("sales_transactions.payment_method = 'POTONG_GAJI'")
		default:
			query = query.Where("sales_transactions.payment_method = ?", paymentMethod)
		}
	}

	// Filter Status Pembayaran
	if paymentStatus != "" && paymentStatus != "ALL" {
		switch paymentStatus {
		case "OVERDUE":
			query = query.Where("sales_transactions.payment_method = 'TEMPO' AND sales_transactions.payment_status != 'PAID' AND sales_transactions.payment_due_date IS NOT NULL AND sales_transactions.payment_due_date < CURRENT_DATE")
		case "UNPAID_OR_PARTIAL", "OUTSTANDING":
			query = query.Where("sales_transactions.payment_status IN ('UNPAID', 'PARTIAL')")
		case "UNPAID":
			query = query.Where("sales_transactions.payment_status = 'UNPAID'")
		case "PARTIAL":
			query = query.Where("sales_transactions.payment_status = 'PARTIAL'")
		case "PAID":
			query = query.Where("sales_transactions.payment_status = 'PAID'")
		default:
			query = query.Where("sales_transactions.payment_status = ?", paymentStatus)
		}
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, 0, err
	}

	// Calculate total amount for the filtered query
	// We need to copy the query before adding limit/offset
	amountQuery := query.Session(&gorm.Session{})
	if status != "CANCELLED" {
		amountQuery = amountQuery.Where("sales_transactions.status != 'CANCELLED'")
	}
	var result struct {
		TotalAmount float64
	}
	amountQuery.Select("COALESCE(SUM(total_amount), 0) as total_amount").Scan(&result)
	totalAmount = result.TotalAmount

	offset := (page - 1) * limit
	if limit == -1 {
		offset = -1
	}

	if err := query.Preload("Company").Preload("Store").Preload("Employee").
		Preload("Employee.Branch").Preload("Employee.OperatingBranch").
		Preload("Employee.Company").Preload("Employee.JobPosition").
		Preload("BuyerEmployee").Preload("CashierEmployee").Preload("KasPabrik").Preload("Branch").
		Preload("DeliveryItems").Preload("DeliveryItems.DeliveryBatch").
		Preload("DeliveryItems.DeliveryBatch.Driver").
		Preload("VerifiedBy").Preload("VerifiedBy.JobPosition").
		Preload("DeletedBy").Preload("DeletedBy.JobPosition").
		Preload("AdjustmentLogs").Preload("AdjustmentLogs.Product").
		Preload("Returns").Preload("Returns.Items").Preload("Returns.Items.Product").
		Preload("Visit").
		Preload("Items").Preload("Items.Product").Preload("BonusItems").Preload("BonusItems.BonusInventory").Preload("Installments").Preload("Payments").Preload("Payments.CompanyBankAccount").
		Order("sales_transactions.transaction_date desc, sales_transactions.created_at desc").
		Limit(limit).Offset(offset).
		Find(&transactions).Error; err != nil {
		return nil, 0, 0, err
	}

	// A converted invoice may be dated later than its SO. Mark it even when the
	// SO falls outside the requested transaction period, so visit summaries do
	// not count the same order again on the invoice date.
	if len(transactions) > 0 {
		transactionIDs := make([]uuid.UUID, 0, len(transactions))
		for _, transaction := range transactions {
			transactionIDs = append(transactionIDs, transaction.ID)
		}
		var linkedOrders []struct {
			ID        uuid.UUID
			InvoiceID uuid.UUID
		}
		if err := r.db.Model(&domain.SalesOrder{}).
			Select("id, invoice_id").
			Where("invoice_id IN ? AND status NOT IN ?", transactionIDs, []string{"CANCELLED", "REJECTED", "DELETED"}).
			Scan(&linkedOrders).Error; err != nil {
			return nil, 0, 0, err
		}
		orderByInvoice := make(map[uuid.UUID]uuid.UUID, len(linkedOrders))
		for _, order := range linkedOrders {
			orderByInvoice[order.InvoiceID] = order.ID
		}
		for i := range transactions {
			if orderID, ok := orderByInvoice[transactions[i].ID]; ok {
				transactions[i].OriginSalesOrderID = &orderID
			}
		}
	}

	ptrTransactions := make([]*domain.SalesTransaction, len(transactions))
	for i := range transactions {
		ptrTransactions[i] = &transactions[i]
	}
	r.injectBonusProducts(ptrTransactions)

	return transactions, total, totalAmount, nil
}

func (r *salesTransactionRepository) FindByID(id uuid.UUID) (*domain.SalesTransaction, error) {
	var transaction domain.SalesTransaction
	if err := r.db.Preload("Company").Preload("Store").Preload("Employee").
		Preload("Employee.Branch").Preload("Employee.OperatingBranch").
		Preload("Employee.Company").Preload("Employee.JobPosition").
		Preload("BuyerEmployee").Preload("CashierEmployee").Preload("KasPabrik").Preload("Branch").
		Preload("Visit").
		Preload("Items").Preload("Items.Product").Preload("BonusItems").Preload("BonusItems.BonusInventory").Preload("Installments").Preload("Payments").Preload("Payments.CompanyBankAccount").
		Preload("VerifiedBy").Preload("VerifiedBy.JobPosition").
		Preload("DeletedBy").Preload("DeletedBy.JobPosition").
		Preload("AdjustmentLogs").Preload("AdjustmentLogs.Product").
		Preload("DeliveryItems").Preload("DeliveryItems.DeliveryBatch").Preload("DeliveryItems.DeliveryBatch.Driver").
		Preload("Returns").Preload("Returns.Items").Preload("Returns.Items.Product").
		Where("id = ?", id).First(&transaction).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &transaction, nil
}

func (r *salesTransactionRepository) FindByIDUnscoped(id uuid.UUID) (*domain.SalesTransaction, error) {
	var transaction domain.SalesTransaction
	if err := r.db.Unscoped().Preload("Company").Preload("Store").Preload("Employee").
		Preload("Employee.Company").Preload("Employee.JobPosition").
		Preload("Visit").
		Preload("Items").Preload("Items.Product").Preload("BonusItems").Preload("BonusItems.BonusInventory").Preload("Installments").Preload("Payments").Preload("Payments.CompanyBankAccount").
		Preload("VerifiedBy").Preload("VerifiedBy.JobPosition").
		Preload("DeletedBy").Preload("DeletedBy.JobPosition").
		Preload("AdjustmentLogs").Preload("AdjustmentLogs.Product").
		Preload("DeliveryItems").Preload("DeliveryItems.DeliveryBatch").Preload("DeliveryItems.DeliveryBatch.Driver").
		Preload("Returns").Preload("Returns.Items").Preload("Returns.Items.Product").
		Where("id = ?", id).First(&transaction).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &transaction, nil
}

func (r *salesTransactionRepository) Update(transaction *domain.SalesTransaction) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Delete existing items first
		if err := tx.Unscoped().Where("sales_transaction_id = ?", transaction.ID).Delete(&domain.SalesItem{}).Error; err != nil {
			return err
		}
		// Reset IDs for new items to force clean insertion
		for i := range transaction.Items {
			transaction.Items[i].ID = uuid.New()
			transaction.Items[i].SalesTransactionID = transaction.ID
		}
		// Save the transaction and its new items (installments will be cascaded)
		// Omit preloaded associations to prevent GORM from attempting cascading inserts on preloaded relations
		return tx.Omit("Company", "Store", "Employee", "Visit", "VerifiedBy", "DeletedBy", "AdjustmentLogs", "DeliveryItems", "Returns").Save(transaction).Error
	})
}

func (r *salesTransactionRepository) UpdateTransactionOnly(transaction *domain.SalesTransaction) error {
	return r.db.Omit("Items", "Installments", "DeliveryItems").Save(transaction).Error
}

func (r *salesTransactionRepository) UpdateInstallment(installment *domain.SalesInstallment) error {
	return r.db.Save(installment).Error
}

func (r *salesTransactionRepository) Delete(id uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 1. Delete sales_items
		if err := tx.Where("sales_transaction_id = ?", id).Delete(&domain.SalesItem{}).Error; err != nil {
			return err
		}
		// 2. Delete sales_installments
		if err := tx.Where("sales_transaction_id = ?", id).Delete(&domain.SalesInstallment{}).Error; err != nil {
			return err
		}
		// 3. Delete sales_payments
		if err := tx.Where("sales_transaction_id = ?", id).Delete(&domain.SalesPayment{}).Error; err != nil {
			return err
		}
		// 4. Update sales_orders pointing to this transaction (nullify invoice_id and restore status to confirmed)
		var so domain.SalesOrder
		if err := tx.Where("invoice_id = ?", id).First(&so).Error; err == nil {
			so.InvoiceID = nil
			so.Status = domain.SOStatusConfirmed
			_ = tx.Save(&so)
		}
		// 5. Update delivery_items pointing to this transaction
		if err := tx.Table("delivery_items").Where("sales_transaction_id = ?", id).Update("sales_transaction_id", nil).Error; err != nil {
			return err
		}
		// 6. Update sales_returns pointing to this transaction
		if err := tx.Table("sales_returns").Where("transaction_id = ?", id).Update("transaction_id", nil).Error; err != nil {
			return err
		}
		// 7. Delete sales_transaction itself
		return tx.Where("id = ?", id).Delete(&domain.SalesTransaction{}).Error
	})
}

func (r *salesTransactionRepository) SoftDeleteWithReason(id uuid.UUID, deletedByID uuid.UUID, reason string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		now := time.Now()
		// 1. Soft delete items
		if err := tx.Model(&domain.SalesItem{}).Where("sales_transaction_id = ?", id).Update("deleted_at", now).Error; err != nil {
			return err
		}
		// 2. Soft delete installments
		if err := tx.Model(&domain.SalesInstallment{}).Where("sales_transaction_id = ?", id).Update("deleted_at", now).Error; err != nil {
			return err
		}
		// 3. Soft delete payments
		if err := tx.Model(&domain.SalesPayment{}).Where("sales_transaction_id = ?", id).Update("deleted_at", now).Error; err != nil {
			return err
		}
		// 4. Soft delete bonus items
		if err := tx.Model(&domain.SalesTransactionBonusItem{}).Where("sales_transaction_id = ?", id).Update("deleted_at", now).Error; err != nil {
			return err
		}
		// 5. Update sales_orders pointing to this transaction (nullify invoice_id and restore status to confirmed)
		var so domain.SalesOrder
		if err := tx.Where("invoice_id = ?", id).First(&so).Error; err == nil {
			so.InvoiceID = nil
			so.Status = domain.SOStatusConfirmed
			_ = tx.Save(&so)
		}
		// 6. Update delivery_items pointing to this transaction
		if err := tx.Table("delivery_items").Where("sales_transaction_id = ?", id).Update("sales_transaction_id", nil).Error; err != nil {
			return err
		}
		// 7. Update sales_returns pointing to this transaction
		if err := tx.Table("sales_returns").Where("transaction_id = ?", id).Update("transaction_id", nil).Error; err != nil {
			return err
		}
		// 8. Soft delete sales_transaction itself
		updates := map[string]interface{}{
			"deleted_at":      now,
			"deleted_by_id":   deletedByID,
			"deletion_reason": reason,
		}
		return tx.Model(&domain.SalesTransaction{}).Where("id = ?", id).Updates(updates).Error
	})
}

func (r *salesTransactionRepository) Restore(id uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 1. Restore items
		if err := tx.Unscoped().Model(&domain.SalesItem{}).Where("sales_transaction_id = ?", id).Update("deleted_at", nil).Error; err != nil {
			return err
		}
		// 2. Restore installments
		if err := tx.Unscoped().Model(&domain.SalesInstallment{}).Where("sales_transaction_id = ?", id).Update("deleted_at", nil).Error; err != nil {
			return err
		}
		// 3. Restore payments
		if err := tx.Unscoped().Model(&domain.SalesPayment{}).Where("sales_transaction_id = ?", id).Update("deleted_at", nil).Error; err != nil {
			return err
		}
		// 4. Restore bonus items
		if err := tx.Unscoped().Model(&domain.SalesTransactionBonusItem{}).Where("sales_transaction_id = ?", id).Update("deleted_at", nil).Error; err != nil {
			return err
		}
		// 5. Restore sales_transaction itself
		updates := map[string]interface{}{
			"deleted_at":      nil,
			"deleted_by_id":   nil,
			"deletion_reason": nil,
		}
		return tx.Unscoped().Model(&domain.SalesTransaction{}).Where("id = ?", id).Updates(updates).Error
	})
}

func (r *salesTransactionRepository) GetTransactionsByEmployeeAndPeriod(employeeID uuid.UUID, month, year int) ([]domain.SalesTransaction, error) {
	var transactions []domain.SalesTransaction
	if err := r.db.Where("employee_id = ? AND period_month = ? AND period_year = ?", employeeID, month, year).Order("created_at desc").Find(&transactions).Error; err != nil {
		return nil, err
	}
	return transactions, nil
}

func (r *salesTransactionRepository) FindByReceiptNo(receiptNo string) (*domain.SalesTransaction, error) {
	var transaction domain.SalesTransaction
	if err := r.db.Preload("Company").Preload("Store").Preload("Employee").
		Preload("Employee.Company").Preload("Employee.JobPosition").
		Preload("CashierEmployee").Preload("BuyerEmployee").Preload("Branch").
		Preload("Items").Preload("Items.Product").Preload("BonusItems").Preload("BonusItems.BonusInventory").Preload("Installments").
		Preload("DeliveryItems").Preload("DeliveryItems.DeliveryBatch").Preload("DeliveryItems.DeliveryBatch.Driver").
		Where("receipt_no = ?", receiptNo).First(&transaction).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &transaction, nil
}

func (r *salesTransactionRepository) CreatePayment(payment *domain.SalesPayment) error {
	return r.db.Create(payment).Error
}

// FindPendingPaymentByTrxAndAmount checks if a PENDING CASH payment already exists
// for the given transaction and amount. Used as an idempotency guard.
func (r *salesTransactionRepository) FindPendingPaymentByTrxAndAmount(trxID uuid.UUID, amount float64, out *domain.SalesPayment) error {
	return r.db.Where(
		"sales_transaction_id = ? AND amount = ? AND payment_status = ? AND payment_method = ?",
		trxID, amount, "PENDING", "CASH",
	).First(out).Error
}

func (r *salesTransactionRepository) FindPaymentsByEmployee(employeeID uuid.UUID) ([]domain.SalesPayment, error) {
	var payments []domain.SalesPayment
	err := r.db.
		Preload("SalesTransaction").
		Preload("SalesTransaction.Store").
		Preload("SalesTransaction.Company").
		Preload("SalesTransaction.Items").
		Preload("SalesTransaction.Items.Product").
		Preload("SalesOrder").
		Preload("SalesOrder.Store").
		Preload("SalesOrder.Company").
		Preload("SalesOrder.Items").
		Preload("SalesOrder.Items.Product").
		Where("employee_id = ?", employeeID).Find(&payments).Error
	return payments, err
}

func (r *salesTransactionRepository) GetOutstandingByStore(storeID uuid.UUID) ([]domain.SalesTransaction, error) {
	var transactions []domain.SalesTransaction
	err := r.db.Preload("Store").Preload("Employee").Preload("Items").Preload("Installments").
		Where("store_id = ? AND payment_status != ? AND status != 'CANCELLED'", storeID, domain.PaymentStatusPaid).
		Order("transaction_date ASC").Order("created_at desc").Find(&transactions).Error
	return transactions, err
}

func (r *salesTransactionRepository) GetByDueDate(date string, employeeID uuid.UUID) ([]domain.SalesTransaction, error) {
	var transactions []domain.SalesTransaction
	db := r.db.Preload("Store").Preload("Employee").Preload("Items").Preload("Installments")
	if date != "" {
		db = db.Where("employee_id = ? AND payment_due_date = ? AND payment_status != ? AND status != 'CANCELLED'", employeeID, date, domain.PaymentStatusPaid)
	} else {
		db = db.Where("employee_id = ? AND payment_due_date IS NOT NULL AND payment_status != ? AND status != 'CANCELLED'", employeeID, domain.PaymentStatusPaid)
	}
	err := db.Order("payment_due_date ASC").Order("created_at desc").Find(&transactions).Error
	return transactions, err
}

func (r *salesTransactionRepository) FindPendingByEmployeeID(employeeID uuid.UUID) ([]domain.SalesTransaction, error) {
	var transactions []domain.SalesTransaction
	err := r.db.Preload("Store").
		Preload("Items").Preload("Items.Product").
		Preload("BonusItems").Preload("BonusItems.BonusInventory").
		Preload("DeliveryItems").Preload("DeliveryItems.DeliveryBatch").Preload("DeliveryItems.DeliveryBatch.Driver").
		Where("employee_id = ? AND status = ?", employeeID, "PENDING").
		Order("transaction_date DESC").Order("created_at DESC").
		Find(&transactions).Error
	if err != nil {
		return nil, err
	}

	ptrTransactions := make([]*domain.SalesTransaction, len(transactions))
	for i := range transactions {
		ptrTransactions[i] = &transactions[i]
	}
	r.injectBonusProducts(ptrTransactions)

	return transactions, nil
}

func (r *salesTransactionRepository) FindByStatusAndCompany(status string, companyID uuid.UUID) ([]domain.SalesTransaction, error) {
	var transactions []domain.SalesTransaction
	db := r.db.Preload("Store").Preload("Employee").Preload("Company").
		Preload("Employee.Company").Preload("Employee.JobPosition").
		Preload("DeliveryItems").Preload("DeliveryItems.DeliveryBatch").
		Preload("Items").Preload("Items.Product").Preload("BonusItems").Preload("BonusItems.BonusInventory")
	if status != "" {
		db = db.Where("status = ?", status)
	}
	if companyID != uuid.Nil {
		db = db.Joins("LEFT JOIN employees emp_filter ON emp_filter.id = sales_transactions.employee_id").
			Where("emp_filter.company_id = ?", companyID)
	}
	err := db.Order("transaction_date DESC").Order("created_at desc").Find(&transactions).Error

	if err == nil {
		ptrTransactions := make([]*domain.SalesTransaction, len(transactions))
		for i := range transactions {
			ptrTransactions[i] = &transactions[i]
		}
		r.injectBonusProducts(ptrTransactions)
	}

	return transactions, err
}

func (r *salesTransactionRepository) FindDeliveryPending(companyID uuid.UUID) ([]domain.SalesTransaction, error) {
	var transactions []domain.SalesTransaction

	// Query transactions with status VERIFIED that are NOT in any delivery_items
	db := r.db.Preload("Store").Preload("Employee").Preload("Company").
		Preload("Employee.Company").Preload("Employee.JobPosition").
		Where("status = ?", "VERIFIED").
		Where("id NOT IN (SELECT sales_transaction_id FROM delivery_items)")

	if companyID != uuid.Nil {
		db = db.Joins("LEFT JOIN employees emp_filter ON emp_filter.id = sales_transactions.employee_id").
			Where("emp_filter.company_id = ?", companyID)
	}

	err := db.Order("transaction_date DESC").Order("created_at desc").Find(&transactions).Error

	if err == nil {
		ptrTransactions := make([]*domain.SalesTransaction, len(transactions))
		for i := range transactions {
			ptrTransactions[i] = &transactions[i]
		}
		r.injectBonusProducts(ptrTransactions)
	}

	return transactions, err
}

func (r *salesTransactionRepository) CountByCompany(companyID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&domain.SalesTransaction{}).
		Joins("LEFT JOIN employees emp_filter ON emp_filter.id = sales_transactions.employee_id").
		Where("emp_filter.company_id = ?", companyID).Count(&count).Error
	return count, err
}

func (r *salesTransactionRepository) CountByCompanyAndDate(companyID uuid.UUID, date time.Time) (int64, error) {
	var count int64
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)
	err := r.db.Model(&domain.SalesTransaction{}).
		Joins("LEFT JOIN employees emp_filter ON emp_filter.id = sales_transactions.employee_id").
		Where("emp_filter.company_id = ? AND transaction_date >= ? AND transaction_date < ?", companyID, startOfDay, endOfDay).
		Count(&count).Error
	return count, err
}

func (r *salesTransactionRepository) GetProductSalesDistribution(productIDs []uuid.UUID, companyID uuid.UUID) ([]map[string]interface{}, error) {
	var results []map[string]interface{}
	if len(productIDs) == 0 {
		return results, nil
	}

	companyFilterMT := ""
	companyFilterTO := ""
	args := []interface{}{productIDs}

	if companyID != uuid.Nil {
		companyFilterMT = "AND st.company_id = ?"
		companyFilterTO = "AND st.company_id = ?"
		args = append(args, companyID) // For MT
	}

	// Add TO args
	args = append(args, productIDs)
	if companyID != uuid.Nil {
		args = append(args, companyID) // For TO
	}

	query := `
		SELECT store_id, store_name, latitude, longitude, SUM(total_quantity) as total_quantity
		FROM (
			SELECT s.id as store_id, s.name as store_name, s.latitude, s.longitude, SUM(si.quantity) as total_quantity
			FROM sales_order_items si
			JOIN sales_orders st ON si.sales_order_id = st.id
			JOIN stores s ON st.store_id = s.id
			WHERE si.product_id IN (?) AND st.status NOT IN ('CANCELLED', 'REJECTED') ` + companyFilterMT + `
			GROUP BY s.id, s.name, s.latitude, s.longitude

			UNION ALL

			SELECT s.id as store_id, s.name as store_name, s.latitude, s.longitude, SUM(si.quantity) as total_quantity
			FROM sales_items si
			JOIN sales_transactions st ON si.sales_transaction_id = st.id
			JOIN stores s ON st.store_id = s.id
			WHERE si.product_id IN (?) AND st.status != 'REJECTED' ` + companyFilterTO + `
			GROUP BY s.id, s.name, s.latitude, s.longitude
		) AS combined
		GROUP BY store_id, store_name, latitude, longitude
	`

	err := r.db.Raw(query, args...).Scan(&results).Error
	return results, err
}

func (r *salesTransactionRepository) GetTopProducts(companyID uuid.UUID, limit int) ([]map[string]interface{}, error) {
	var results []map[string]interface{}
	query := r.db.Table("sales_items si").
		Select("p.id, p.name, p.image_url as photo_url, SUM(si.quantity) as total_sold").
		Joins("JOIN products p ON si.product_id = p.id").
		Joins("JOIN sales_transactions st ON si.sales_transaction_id = st.id").
		Where("st.status = 'VERIFIED'")

	if companyID != uuid.Nil {
		query = query.Joins("JOIN employees e_filter ON st.employee_id = e_filter.id").
			Where("e_filter.company_id = ?", companyID)
	}

	query = query.Group("p.id, p.name, p.image_url").
		Order("total_sold DESC").
		Limit(limit)

	err := query.Find(&results).Error
	return results, err
}

func (r *salesTransactionRepository) GetSalesKasbon(companyID uuid.UUID) ([]map[string]interface{}, error) {
	var results []map[string]interface{}

	query := r.db.Table("sales_transactions st").
		Select("e.id as employee_id, e.first_name as employee_name, COUNT(st.id) as total_invoices, SUM(st.total_amount - st.paid_amount) as total_kasbon").
		Joins("JOIN employees e ON st.employee_id = e.id").
		Where("st.payment_method = ? AND st.status != 'CANCELLED'", "TEMPO").
		Where("st.payment_status != ?", domain.PaymentStatusPaid).
		Where("st.payment_due_date < CURRENT_DATE")

	if companyID != uuid.Nil {
		query = query.Where("e.company_id = ?", companyID)
	}

	err := query.Group("e.id, e.first_name").Having("SUM(st.total_amount - st.paid_amount) > 0").Scan(&results).Error
	return results, err
}

func (r *salesTransactionRepository) GetProductRecapByEmployeeAndPeriod(employeeID uuid.UUID, month, year int) ([]domain.ProductRecap, error) {
	var recaps []domain.ProductRecap
	err := r.db.Table("sales_items").
		Select("sales_items.product_id, products.sku, products.name as product_name, SUM(sales_items.quantity) as total_qty, SUM(sales_items.subtotal) as total_amount").
		Joins("JOIN sales_transactions ON sales_transactions.id = sales_items.sales_transaction_id").
		Joins("JOIN products ON products.id = sales_items.product_id").
		Where("sales_transactions.employee_id = ? AND sales_transactions.period_month = ? AND sales_transactions.period_year = ? AND sales_transactions.status != 'CANCELLED'", employeeID, month, year).
		Group("sales_items.product_id, products.sku, products.name").
		Scan(&recaps).Error
	return recaps, err
}
