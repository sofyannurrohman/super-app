package repository

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"gorm.io/gorm"
)

type SalesVisitRepository interface {
	Create(visit *domain.SalesVisit) error
	Update(visit *domain.SalesVisit) error
	FindByID(id uuid.UUID) (*domain.SalesVisit, error)
	FindTodayByEmployeeAndStore(employeeID, storeID uuid.UUID) (*domain.SalesVisit, error)
	FindLatestByEmployeeStoreAndDate(employeeID, storeID uuid.UUID, date time.Time) (*domain.SalesVisit, error)
	FindDuplicateVisit(employeeID, storeID uuid.UUID, checkInTime time.Time, windowMinutes int) (*domain.SalesVisit, error)
	FindAllHistory(limit, offset int, employeeID *uuid.UUID, branchID *uuid.UUID, salesType string, search string, startDate, endDate *time.Time, summary bool) ([]domain.SalesVisit, error)
	CountHistory(employeeID *uuid.UUID, branchID *uuid.UUID, salesType string, search string, startDate, endDate *time.Time) (int64, error)
	Delete(id uuid.UUID) error
}

type salesVisitRepository struct {
	db *gorm.DB
}

func NewSalesVisitRepository(db *gorm.DB) SalesVisitRepository {
	return &salesVisitRepository{db}
}

func (r *salesVisitRepository) Create(visit *domain.SalesVisit) error {
	return r.db.Create(visit).Error
}

func (r *salesVisitRepository) Update(visit *domain.SalesVisit) error {
	return r.db.Save(visit).Error
}

func (r *salesVisitRepository) FindByID(id uuid.UUID) (*domain.SalesVisit, error) {
	var visit domain.SalesVisit
	err := r.db.Preload("Employee.User").Preload("Store").Preload("Transactions").Preload("SalesOrders").Where("id = ?", id).First(&visit).Error
	if err != nil {
		return nil, err
	}
	return &visit, nil
}

func (r *salesVisitRepository) FindTodayByEmployeeAndStore(employeeID, storeID uuid.UUID) (*domain.SalesVisit, error) {
	var visit domain.SalesVisit
	now := time.Now()
	// Look back 24 hours to find the latest visit for this store/employee, allowing for slight client clock drift
	start := now.Add(-24 * time.Hour)
	end := now.Add(2 * time.Hour)
	err := r.db.Where("employee_id = ? AND store_id = ? AND check_in_time >= ? AND check_in_time <= ? AND status != 'SKIPPED' AND type != 'SKIP'", employeeID, storeID, start, end).
		Order("check_in_time DESC").
		First(&visit).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &visit, nil
}

func (r *salesVisitRepository) FindDuplicateVisit(employeeID, storeID uuid.UUID, checkInTime time.Time, windowMinutes int) (*domain.SalesVisit, error) {
	var visit domain.SalesVisit
	if windowMinutes <= 0 {
		windowMinutes = 5
	}
	start := checkInTime.Add(-time.Duration(windowMinutes) * time.Minute)
	end := checkInTime.Add(time.Duration(windowMinutes) * time.Minute)
	err := r.db.Where("employee_id = ? AND store_id = ? AND check_in_time >= ? AND check_in_time <= ? AND status != 'SKIPPED' AND type != 'SKIP'", employeeID, storeID, start, end).
		Order("check_in_time DESC").
		First(&visit).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &visit, nil
}

func (r *salesVisitRepository) FindLatestByEmployeeStoreAndDate(employeeID, storeID uuid.UUID, date time.Time) (*domain.SalesVisit, error) {
	var visit domain.SalesVisit
	start := date.Add(-24 * time.Hour)
	end := date.Add(24 * time.Hour)

	err := r.db.Where("employee_id = ? AND store_id = ? AND check_in_time >= ? AND check_in_time <= ? AND status != 'SKIPPED' AND type != 'SKIP'", employeeID, storeID, start, end).
		Order("check_in_time DESC").
		First(&visit).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &visit, nil
}

func (r *salesVisitRepository) applyHistoryFilters(query *gorm.DB, employeeID *uuid.UUID, branchID *uuid.UUID, salesType string, search string, startDate, endDate *time.Time) *gorm.DB {
	if employeeID != nil && *employeeID != uuid.Nil {
		query = query.Where("employee_id = ?", *employeeID)
	}

	if branchID != nil && *branchID != uuid.Nil {
		query = query.Where("employee_id IN (SELECT id FROM employees WHERE COALESCE(operating_branch_id, branch_id) = ?)", *branchID)
	}

	if salesType != "" && salesType != "ALL_SALES" {
		if salesType == "sales_to" {
			query = query.Where("employee_id IN (SELECT employees.id FROM employees JOIN job_positions jp ON jp.id = employees.job_position_id WHERE jp.title ~* '(^|[^[:alnum:]])(to|take order|task order)([^[:alnum:]]|$)')")
		}
		var likeQueries []string
		if salesType == "sales_mt" {
			likeQueries = []string{"%mt%", "%modern trade%", "%kam%", "%key account%"}
		} else if salesType == "sales_motoris_booster" {
			likeQueries = []string{"%motoris%", "%booster%"}
		}

		if len(likeQueries) > 0 {
			var conditions []string
			var values []interface{}
			for _, q := range likeQueries {
				conditions = append(conditions, "LOWER(jp.title) LIKE ?")
				values = append(values, q)
			}

			subQuery := r.db.Model(&domain.Employee{}).Select("employees.id").
				Joins("JOIN job_positions jp ON jp.id = employees.job_position_id").
				Where(strings.Join(conditions, " OR "), values...)

			query = query.Where("employee_id IN (?)", subQuery)
		}
	}

	if search = strings.TrimSpace(search); search != "" {
		searchTerm := "%" + strings.ToLower(search) + "%"
		query = query.Where(`
			sales_visits.employee_id IN (
				SELECT id FROM employees WHERE LOWER(first_name) LIKE ? OR LOWER(last_name) LIKE ? OR LOWER(first_name || ' ' || last_name) LIKE ?
			)
			OR sales_visits.store_id IN (
				SELECT id FROM stores WHERE LOWER(name) LIKE ? OR LOWER(address) LIKE ?
			)
			OR LOWER(sales_visits.notes) LIKE ?
		`, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm)
	}

	if startDate != nil {
		query = query.Where("check_in_time >= ?", *startDate)
	}
	if endDate != nil {
		query = query.Where("check_in_time <= ?", *endDate)
	}

	return query.Where("status != ?", "SKIPPED")
}

func (r *salesVisitRepository) FindAllHistory(limit, offset int, employeeID *uuid.UUID, branchID *uuid.UUID, salesType string, search string, startDate, endDate *time.Time, summary bool) ([]domain.SalesVisit, error) {
	var visits []domain.SalesVisit
	query := r.db.Model(&domain.SalesVisit{})

	if summary {
		query = query.
			Select("id, employee_id, store_id, check_in_time, check_out_time, type, status, created_at").
			Preload("Store", func(db *gorm.DB) *gorm.DB {
				return db.Select("id, name, address, category")
			})
	} else {
		query = query.
			Preload("Employee.User").
			Preload("Employee.Branch").
			Preload("Employee.OperatingBranch").
			Preload("Store").
			Preload("Transactions").
			Preload("Transactions.Items.Product").
			Preload("SalesOrders").
			Preload("SalesOrders.Items.Product")
	}

	query = r.applyHistoryFilters(query, employeeID, branchID, salesType, search, startDate, endDate)

	err := query.Order("check_in_time DESC").Limit(limit).Offset(offset).Find(&visits).Error
	return visits, err
}

func (r *salesVisitRepository) CountHistory(employeeID *uuid.UUID, branchID *uuid.UUID, salesType string, search string, startDate, endDate *time.Time) (int64, error) {
	var count int64
	query := r.applyHistoryFilters(r.db.Model(&domain.SalesVisit{}), employeeID, branchID, salesType, search, startDate, endDate)

	err := query.Count(&count).Error
	return count, err
}

func (r *salesVisitRepository) Delete(id uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 1. Ambil data SalesTransaction dan bersihkan anak-anaknya yang tidak tercascade
		var trxs []domain.SalesTransaction
		if err := tx.Where("visit_id = ?", id).Select("id").Find(&trxs).Error; err == nil && len(trxs) > 0 {
			var trxIDs []uuid.UUID
			for _, t := range trxs {
				trxIDs = append(trxIDs, t.ID)
			}
			// Hapus Bonus Items
			if err := tx.Where("sales_transaction_id IN ?", trxIDs).Delete(&domain.SalesTransactionBonusItem{}).Error; err != nil {
				return err
			}
			// Hapus Payments
			if err := tx.Where("sales_transaction_id IN ?", trxIDs).Delete(&domain.SalesPayment{}).Error; err != nil {
				return err
			}
			// Hapus Sales Return Items & Sales Returns
			var returnTrxs []domain.SalesReturn
			if err := tx.Where("transaction_id IN ?", trxIDs).Select("id").Find(&returnTrxs).Error; err == nil && len(returnTrxs) > 0 {
				var retIDs []uuid.UUID
				for _, r := range returnTrxs {
					retIDs = append(retIDs, r.ID)
				}
				if err := tx.Where("sales_return_id IN ?", retIDs).Delete(&domain.SalesReturnItem{}).Error; err != nil {
					return err
				}
				if err := tx.Where("id IN ?", retIDs).Delete(&domain.SalesReturn{}).Error; err != nil {
					return err
				}
			}
		}

		// 2. Ambil data SalesOrder dan bersihkan anak-anaknya yang tidak tercascade
		var orders []domain.SalesOrder
		if err := tx.Where("visit_id = ?", id).Select("id", "so_number").Find(&orders).Error; err == nil && len(orders) > 0 {
			var orderIDs []uuid.UUID
			var soNumbers []string
			for _, o := range orders {
				orderIDs = append(orderIDs, o.ID)
				if o.SONumber != "" {
					soNumbers = append(soNumbers, o.SONumber)
				}
			}

			// Hapus MT Purchase Orders
			if len(soNumbers) > 0 {
				if err := tx.Where("po_number IN ?", soNumbers).Delete(&domain.MTPurchaseOrder{}).Error; err != nil {
					return err
				}
			}

			// Hapus Bonus Items
			if err := tx.Where("sales_order_id IN ?", orderIDs).Delete(&domain.SalesOrderBonusItem{}).Error; err != nil {
				return err
			}
			// Hapus Payments
			if err := tx.Where("sales_order_id IN ?", orderIDs).Delete(&domain.SalesPayment{}).Error; err != nil {
				return err
			}
			// Hapus Sales Return Items & Sales Returns
			var returnOrders []domain.SalesReturn
			if err := tx.Where("sales_order_id IN ?", orderIDs).Select("id").Find(&returnOrders).Error; err == nil && len(returnOrders) > 0 {
				var retIDs []uuid.UUID
				for _, r := range returnOrders {
					retIDs = append(retIDs, r.ID)
				}
				if err := tx.Where("sales_return_id IN ?", retIDs).Delete(&domain.SalesReturnItem{}).Error; err != nil {
					return err
				}
				if err := tx.Where("id IN ?", retIDs).Delete(&domain.SalesReturn{}).Error; err != nil {
					return err
				}
			}
		}

		// 3. Setelah anak-anak protektif dihapus, hapus visit dengan aman
		return tx.Delete(&domain.SalesVisit{}, id).Error
	})
}
