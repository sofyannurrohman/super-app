package repository

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/lib/pq"
	"example.com/hris/backend/internal/domain"
	"gorm.io/gorm"
)

type SalesOrderRepository interface {
	Create(so *domain.SalesOrder) error
	CreateWithNextSONumber(so *domain.SalesOrder, prefix string) error
	Update(so *domain.SalesOrder) error
	FindByID(id uuid.UUID) (*domain.SalesOrder, error)
	FindBySONumber(soNumber string) (*domain.SalesOrder, error)
	FindByBranch(branchIDs []uuid.UUID, status string, salesType string, paymentStatus string, startDate string, endDate string, employeeID uuid.UUID, companyID uuid.UUID, search string, page int, limit int) ([]domain.SalesOrder, int64, *SOSummary, error)
	FindByEmployee(employeeID uuid.UUID) ([]domain.SalesOrder, error)
	FindByStore(storeID uuid.UUID, page int, limit int) ([]domain.SalesOrder, int64, error)
	Delete(id uuid.UUID) error
	SoftDeleteWithReason(id uuid.UUID, deletedByID uuid.UUID, reason string) error
	Restore(id uuid.UUID) error
	FindByIDUnscoped(id uuid.UUID) (*domain.SalesOrder, error)
	FindWaitingStockByProduct(branchID, productID uuid.UUID) ([]domain.SalesOrder, error)

	// Payment Methods
	AddPayment(payment *domain.SalesPayment) error
	GetPaymentsBySO(soID uuid.UUID) ([]domain.SalesPayment, error)
	GetPaymentByID(id uuid.UUID) (*domain.SalesPayment, error)
	UpdatePayment(payment *domain.SalesPayment) error
	GetMTSettlementQueue() ([]domain.SalesPayment, error)
}

type SOSummary struct {
	TotalApprovedAmount     float64 `json:"total_approved_amount"`
	TotalApprovedCount      int64   `json:"total_approved_count"`
	TotalPendingAmount      float64 `json:"total_pending_amount"`
	TotalPendingCount       int64   `json:"total_pending_count"`
	TotalRejectedAmount     float64 `json:"total_rejected_amount"`
	TotalRejectedCount      int64   `json:"total_rejected_count"`
	TotalWaitingStockAmount float64 `json:"total_waiting_stock_amount"`
	TotalWaitingStockCount  int64   `json:"total_waiting_stock_count"`
	TotalFilteredAmount     float64 `json:"total_filtered_amount"`
	TotalFilteredCount      int64   `json:"total_filtered_count"`
}

type salesOrderRepository struct {
	db *gorm.DB
}

func NewSalesOrderRepository(db *gorm.DB) SalesOrderRepository {
	return &salesOrderRepository{db}
}

func (r *salesOrderRepository) Create(so *domain.SalesOrder) error {
	return r.db.Create(so).Error
}

// CreateWithNextSONumber allocates the number and inserts the order in one transaction.
// The lock is keyed by the visible number prefix, not company ID: the unique
// constraint on so_number is global and company codes can overlap.
func (r *salesOrderRepository) CreateWithNextSONumber(so *domain.SalesOrder, prefix string) error {
	base := strings.TrimSuffix(GenerateSONumber(prefix, so.OrderDate, 0), "0001")
	var err error
	for attempt := 0; attempt < 3; attempt++ {
		err = r.db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Exec("SELECT pg_advisory_xact_lock(hashtext(?))", base).Error; err != nil {
				return err
			}

			// Unscoped is essential: soft-deleted orders still own their SO numbers.
			// Read the largest suffix rather than the row count, so gaps and old
			// numbering mistakes cannot make a previously used number repeat.
			var last int64
			if err := tx.Unscoped().Model(&domain.SalesOrder{}).
				Select("COALESCE(MAX(CAST(SUBSTRING(so_number FROM CHAR_LENGTH(?) + 1) AS BIGINT)), 0)", base).
				Where("LEFT(so_number, CHAR_LENGTH(?)) = ? AND SUBSTRING(so_number FROM CHAR_LENGTH(?) + 1) ~ '^[0-9]+$'", base, base, base).
				Scan(&last).Error; err != nil {
				return err
			}
			so.SONumber = fmt.Sprintf("%s%04d", base, last+1)
			return tx.Create(so).Error
		})
		if err == nil || !isSONumberConflict(err) {
			return err
		}
	}
	return err
}

func isSONumberConflict(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505" && pgErr.ConstraintName == "sales_orders_so_number_key"
	}
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && string(pqErr.Code) == "23505" && pqErr.Constraint == "sales_orders_so_number_key"
}

func (r *salesOrderRepository) Update(so *domain.SalesOrder) error {
	return r.db.Save(so).Error
}

func (r *salesOrderRepository) FindByID(id uuid.UUID) (*domain.SalesOrder, error) {
	var so domain.SalesOrder
	err := r.db.
		Preload("Items.Product").
		Preload("Items.Batches").
		Preload("BonusItems").
		Preload("BonusItems.BonusInventory").
		Preload("Employee").
		Preload("Employee.JobPosition").
		Preload("Store").
		Preload("Branch").
		Preload("Company").
		Preload("ConfirmedBy").
		Preload("ConfirmedBy.JobPosition").
		Preload("AdminNotaConfirmedBy").
		Preload("AdminNotaConfirmedBy.JobPosition").
		Preload("DeletedBy").
		Preload("DeletedBy.JobPosition").
		First(&so, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &so, nil
}

func (r *salesOrderRepository) FindByIDUnscoped(id uuid.UUID) (*domain.SalesOrder, error) {
	var so domain.SalesOrder
	err := r.db.Unscoped().
		Preload("Items.Product").
		Preload("Items.Batches").
		Preload("BonusItems").
		Preload("BonusItems.BonusInventory").
		Preload("Employee").
		Preload("Employee.JobPosition").
		Preload("Store").
		Preload("Branch").
		Preload("Company").
		Preload("ConfirmedBy").
		Preload("ConfirmedBy.JobPosition").
		Preload("AdminNotaConfirmedBy").
		Preload("AdminNotaConfirmedBy.JobPosition").
		Preload("DeletedBy").
		Preload("DeletedBy.JobPosition").
		First(&so, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &so, nil
}

func (r *salesOrderRepository) FindBySONumber(soNumber string) (*domain.SalesOrder, error) {
	var so domain.SalesOrder
	err := r.db.
		Preload("Items.Product").
		Preload("Items.Batches").
		Preload("BonusItems").
		Preload("BonusItems.BonusInventory").
		Preload("Employee").
		Preload("Employee.JobPosition").
		Preload("Store").
		Preload("ConfirmedBy").
		Preload("ConfirmedBy.JobPosition").
		Preload("AdminNotaConfirmedBy").
		Preload("AdminNotaConfirmedBy.JobPosition").
		First(&so, "so_number = ?", soNumber).Error
	if err != nil {
		return nil, err
	}
	return &so, nil
}

func (r *salesOrderRepository) FindByBranch(branchIDs []uuid.UUID, status string, salesType string, paymentStatus string, startDate string, endDate string, employeeID uuid.UUID, companyID uuid.UUID, search string, page int, limit int) ([]domain.SalesOrder, int64, *SOSummary, error) {
	var orders []domain.SalesOrder
	var total int64

	var q *gorm.DB
	if status == "DELETED" || status == "TRASH" {
		q = r.db.Unscoped().Model(&domain.SalesOrder{}).Where("sales_orders.deleted_at IS NOT NULL")
	} else {
		q = r.db.Model(&domain.SalesOrder{})
	}

	loc := jakartaLoc
	if loc == nil {
		loc = time.FixedZone("WIB", 7*3600)
	}
	if startDate != "" {
		if parsedStart, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(startDate), loc); err == nil {
			startOfDay := time.Date(parsedStart.Year(), parsedStart.Month(), parsedStart.Day(), 0, 0, 0, 0, loc)
			q = q.Where("sales_orders.order_date >= ?", startOfDay)
		} else {
			q = q.Where("DATE(sales_orders.order_date AT TIME ZONE 'Asia/Jakarta') >= ?", startDate)
		}
	}
	if endDate != "" {
		if parsedEnd, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(endDate), loc); err == nil {
			endOfDayExclusive := time.Date(parsedEnd.Year(), parsedEnd.Month(), parsedEnd.Day(), 0, 0, 0, 0, loc).AddDate(0, 0, 1)
			q = q.Where("sales_orders.order_date < ?", endOfDayExclusive)
		} else {
			q = q.Where("DATE(sales_orders.order_date AT TIME ZONE 'Asia/Jakarta') <= ?", endDate)
		}
	}

	joinedStores := false
	joinedEmployees := false
	joinedJobPositions := false

	if len(branchIDs) > 0 {
		q = q.Joins("LEFT JOIN stores s ON s.id = sales_orders.store_id").
			Joins("LEFT JOIN employees se ON se.id = s.assigned_employee_id").
			Joins("LEFT JOIN employees so_emp ON so_emp.id = sales_orders.employee_id").
			Where("COALESCE(so_emp.operating_branch_id, so_emp.branch_id, se.operating_branch_id, se.branch_id, sales_orders.branch_id) IN (?)", branchIDs)
		joinedStores = true
	}

	if employeeID != uuid.Nil {
		q = q.Where("sales_orders.employee_id = ?", employeeID)
	}

	if companyID != uuid.Nil {
		q = q.Where("sales_orders.company_id = ?", companyID)
	}

	if status != "" && status != "ALL" && status != "DELETED" && status != "TRASH" {
		q = q.Where("sales_orders.status = ?", status)
		if status == string(domain.SOStatusConfirmed) {
			q = q.Where("sales_orders.delivery_batch_id IS NULL")
		}
	}

	if paymentStatus != "" && paymentStatus != "ALL" {
		// allow comma separated values like UNPAID,PARTIAL
		statuses := strings.Split(paymentStatus, ",")
		if len(statuses) > 1 {
			q = q.Where("sales_orders.payment_status IN ?", statuses)
		} else {
			q = q.Where("sales_orders.payment_status = ?", paymentStatus)
		}
	}

	if salesType == "sales_mt" {
		q = q.Where("sales_orders.is_mt = ?", true)
	} else if salesType == "sales_to" {
		if !joinedEmployees {
			q = q.Joins("LEFT JOIN employees e ON e.id = sales_orders.employee_id")
			joinedEmployees = true
		}
		if !joinedJobPositions {
			q = q.Joins("LEFT JOIN job_positions jp ON jp.id = e.job_position_id")
			joinedJobPositions = true
		}
		q = q.Where("sales_orders.is_mt = ? AND jp.title ~* '(^|[^[:alnum:]])(to|take order|task order)([^[:alnum:]]|$)'", false)
	} else if salesType == "sales_motoris_booster" {
		if !joinedEmployees {
			q = q.Joins("LEFT JOIN employees e ON e.id = sales_orders.employee_id")
			joinedEmployees = true
		}
		if !joinedJobPositions {
			q = q.Joins("LEFT JOIN job_positions jp ON jp.id = e.job_position_id")
			joinedJobPositions = true
		}
		q = q.Where("sales_orders.is_mt = ? AND (jp.title ILIKE '%motoris%' OR jp.title ILIKE '%booster%')", false)
	}

	if strings.TrimSpace(search) != "" {
		if !joinedStores {
			q = q.Joins("LEFT JOIN stores s ON s.id = sales_orders.store_id")
			joinedStores = true
		}
		if !joinedEmployees {
			q = q.Joins("LEFT JOIN employees e ON e.id = sales_orders.employee_id")
			joinedEmployees = true
		}
		searchPattern := "%" + strings.ToLower(strings.TrimSpace(search)) + "%"
		q = q.Where("(LOWER(sales_orders.so_number) LIKE ? OR LOWER(s.name) LIKE ? OR LOWER(e.first_name) LIKE ? OR LOWER(e.last_name) LIKE ? OR LOWER(CONCAT(e.first_name, ' ', COALESCE(e.last_name, ''))) LIKE ?)",
			searchPattern, searchPattern, searchPattern, searchPattern, searchPattern)
	}

	err := q.Count(&total).Error
	if err != nil {
		return nil, 0, nil, err
	}

	// Calculate unpaginated summary across entire filtered set
	summary := &SOSummary{
		TotalFilteredCount: total,
	}

	type summaryRow struct {
		TotalAmount float64
		Status      string
		HasAdmin    bool
	}
	var rows []summaryRow
	_ = q.Session(&gorm.Session{}).
		Select("sales_orders.total_amount, sales_orders.status, (sales_orders.admin_nota_confirmed_at IS NOT NULL) as has_admin").
		Scan(&rows).Error

	for _, rItem := range rows {
		amt := rItem.TotalAmount
		summary.TotalFilteredAmount += amt

		isRejectedOrCancelled := rItem.Status == "REJECTED" || rItem.Status == "CANCELLED"
		isWaitingStock := rItem.Status == "WAITING_STOCK"
		isApproved := !isRejectedOrCancelled && rItem.Status != "DELETED" && (rItem.HasAdmin ||
			rItem.Status == "CONFIRMED" || rItem.Status == "APPROVED" || rItem.Status == "IN_DELIVERY" ||
			rItem.Status == "DELIVERED" || rItem.Status == "PAID" || rItem.Status == "SHIPPED" || rItem.Status == "CONVERTED")
		isPending := !rItem.HasAdmin && (rItem.Status == "DRAFT" || rItem.Status == "WAITING_TAX_DOCS" ||
			rItem.Status == "WAITING_WAREHOUSE" || rItem.Status == "WAITING_STOCK" || rItem.Status == "PROCESSING")

		if isApproved {
			summary.TotalApprovedAmount += amt
			summary.TotalApprovedCount++
		}
		if isPending {
			summary.TotalPendingAmount += amt
			summary.TotalPendingCount++
		}
		if isRejectedOrCancelled {
			summary.TotalRejectedAmount += amt
			summary.TotalRejectedCount++
		}
		if isWaitingStock {
			summary.TotalWaitingStockAmount += amt
			summary.TotalWaitingStockCount++
		}
	}

	if limit > 0 {
		offset := (page - 1) * limit
		q = q.Limit(limit).Offset(offset)
	}

	q = q.
		Preload("Items.Product").
		Preload("Items.Batches").
		Preload("BonusItems").
		Preload("BonusItems.BonusInventory").
		Preload("Employee").
		Preload("Employee.Branch").
		Preload("Employee.OperatingBranch").
		Preload("Employee.JobPosition").
		Preload("Store").
		Preload("Branch").
		Preload("ConfirmedBy").
		Preload("ConfirmedBy.JobPosition").
		Preload("AdminNotaConfirmedBy").
		Preload("AdminNotaConfirmedBy.JobPosition").
		Preload("DeletedBy").
		Preload("DeletedBy.JobPosition")

	err = q.Order("sales_orders.created_at desc").Find(&orders).Error
	return orders, total, summary, err
}

func (r *salesOrderRepository) FindByEmployee(employeeID uuid.UUID) ([]domain.SalesOrder, error) {
	var orders []domain.SalesOrder
	err := r.db.
		Preload("Items.Product").
		Preload("BonusItems").
		Preload("BonusItems.BonusInventory").
		Preload("Store").
		Preload("Company").
		Preload("Employee").
		Where("employee_id = ?", employeeID).
		Order("created_at desc").
		Find(&orders).Error
	return orders, err
}

func (r *salesOrderRepository) FindByStore(storeID uuid.UUID, page int, limit int) ([]domain.SalesOrder, int64, error) {
	var orders []domain.SalesOrder
	var total int64

	q := r.db.Model(&domain.SalesOrder{}).Where("store_id = ?", storeID)

	err := q.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	if limit > 0 {
		offset := (page - 1) * limit
		q = q.Limit(limit).Offset(offset)
	}

	q = q.
		Preload("Items.Product").
		Preload("BonusItems.BonusInventory").
		Preload("Employee").
		Preload("Store")

	err = q.Order("created_at desc").Find(&orders).Error
	return orders, total, err
}

func (r *salesOrderRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&domain.SalesOrder{}, "id = ?", id).Error
}

func (r *salesOrderRepository) SoftDeleteWithReason(id uuid.UUID, deletedByID uuid.UUID, reason string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		now := time.Now()
		// 1. Update sales_order with deleted_at, deleted_by_id, deletion_reason
		updates := map[string]interface{}{
			"deleted_at":      now,
			"deleted_by_id":   deletedByID,
			"deletion_reason": reason,
		}
		if err := tx.Model(&domain.SalesOrder{}).Where("id = ?", id).Updates(updates).Error; err != nil {
			return err
		}
		// 2. Soft delete items
		if err := tx.Model(&domain.SalesOrderItem{}).Where("sales_order_id = ?", id).Update("deleted_at", now).Error; err != nil {
			return err
		}
		// 3. Soft delete bonus items
		if err := tx.Model(&domain.SalesOrderBonusItem{}).Where("sales_order_id = ?", id).Update("deleted_at", now).Error; err != nil {
			return err
		}
		// 4. Soft delete payments
		if err := tx.Model(&domain.SalesPayment{}).Where("sales_order_id = ?", id).Update("deleted_at", now).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *salesOrderRepository) Restore(id uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 1. Restore sales_order
		updates := map[string]interface{}{
			"deleted_at":      nil,
			"deleted_by_id":   nil,
			"deletion_reason": nil,
		}
		if err := tx.Unscoped().Model(&domain.SalesOrder{}).Where("id = ?", id).Updates(updates).Error; err != nil {
			return err
		}
		// 2. Restore items
		if err := tx.Unscoped().Model(&domain.SalesOrderItem{}).Where("sales_order_id = ?", id).Update("deleted_at", nil).Error; err != nil {
			return err
		}
		// 3. Restore bonus items
		if err := tx.Unscoped().Model(&domain.SalesOrderBonusItem{}).Where("sales_order_id = ?", id).Update("deleted_at", nil).Error; err != nil {
			return err
		}
		// 4. Restore payments
		if err := tx.Unscoped().Model(&domain.SalesPayment{}).Where("sales_order_id = ?", id).Update("deleted_at", nil).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *salesOrderRepository) FindWaitingStockByProduct(branchID, productID uuid.UUID) ([]domain.SalesOrder, error) {
	var orders []domain.SalesOrder
	err := r.db.
		Preload("Items").
		Joins("JOIN sales_order_items ON sales_order_items.sales_order_id = sales_orders.id").
		Where("sales_orders.branch_id = ? AND sales_orders.status = ? AND sales_order_items.product_id = ?",
			branchID, domain.SOStatusWaitingStock, productID).
		Order("sales_orders.created_at asc").
		Distinct().
		Find(&orders).Error
	return orders, err
}

// GenerateSONumber generates a unique SO number: SO-PREFIX-YYYYMMDD-NNNN
func GenerateSONumber(prefix string, date time.Time, count int64) string {
	if prefix == "" {
		prefix = "SO"
	} else {
		prefix = "SO-" + prefix
	}
	return fmt.Sprintf("%s-%s-%04d", prefix, date.Format("20060102"), count+1)
}

func (r *salesOrderRepository) AddPayment(payment *domain.SalesPayment) error {
	return r.db.Create(payment).Error
}

func (r *salesOrderRepository) GetPaymentsBySO(soID uuid.UUID) ([]domain.SalesPayment, error) {
	var payments []domain.SalesPayment
	err := r.db.Where("sales_order_id = ?", soID).Find(&payments).Error
	return payments, err
}

func (r *salesOrderRepository) GetPaymentByID(id uuid.UUID) (*domain.SalesPayment, error) {
	var payment domain.SalesPayment
	err := r.db.Preload("SalesOrder").Preload("SalesOrder.Store").Preload("Employee").First(&payment, "id = ?", id).Error
	return &payment, err
}

func (r *salesOrderRepository) UpdatePayment(payment *domain.SalesPayment) error {
	return r.db.Save(payment).Error
}

func (r *salesOrderRepository) GetMTSettlementQueue() ([]domain.SalesPayment, error) {
	var payments []domain.SalesPayment
	err := r.db.
		Joins("JOIN sales_orders ON sales_orders.id = sales_payments.sales_order_id").
		Preload("SalesOrder").
		Preload("SalesOrder.Store").
		Preload("Employee").
		Where("sales_orders.is_mt = ? AND sales_payments.payment_status IN ?", true, []string{"PENDING", "SUCCESS"}).
		Order("sales_payments.created_at DESC").
		Find(&payments).Error
	return payments, err
}
