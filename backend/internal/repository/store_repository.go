package repository

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"example.com/hris/backend/pkg/utils"
	"gorm.io/gorm"
)

type StoreRepository interface {
	Create(store *domain.Store) error
	FindCompanyIDByName(name string) (uuid.UUID, error)
	FindAll() ([]domain.Store, error)
	FindByID(id uuid.UUID) (*domain.Store, error)
	Update(store *domain.Store) error
	Delete(id uuid.UUID) error
	FindByVisitDay(day int, employeeID uuid.UUID) ([]domain.Store, error)
	FindByCompanyID(companyID uuid.UUID) ([]domain.Store, error)
	FindPaginated(companyID *uuid.UUID, branchID *uuid.UUID, assignedEmpID *uuid.UUID, category string, search string, status string, sortBy string, limit int, offset int, salesType string, completeness string, includeMetrics ...bool) ([]domain.Store, int64, error)
	GetStoreHistories(storeIDs []uuid.UUID) (map[uuid.UUID]StoreHistory, error)
	GetOverdueStores(employeeID uuid.UUID, days int) ([]domain.Store, error)
	TransferStores(fromEmployeeID uuid.UUID, toEmployeeID uuid.UUID) (int64, error)
	UpdateRegion(id uuid.UUID, province, regency, district, village string) error
	CreateAudit(audit *domain.StoreAudit) error
	GetAuditsByStoreID(storeID uuid.UUID) ([]domain.StoreAudit, error)
	GetAllAudits(companyID *uuid.UUID, branchID *uuid.UUID, limit int, offset int) ([]domain.StoreAudit, int64, error)
	CheckDuplicate(req domain.CheckDuplicateStoreRequest, currentEmpID *uuid.UUID) (*domain.CheckDuplicateStoreResponse, error)
	GetStoreDistribution(companyID *uuid.UUID, branchIDs []uuid.UUID, assignedEmpID *uuid.UUID, salesType string, status string) ([]domain.StoreDistributionItem, error)
}

type StoreHistory struct {
	StoreID               uuid.UUID
	LastVisitDate         *time.Time
	LastTransactionDate   *time.Time
	LastTransactionAmount float64
	LastTransactionItems  string
}

type storeRepository struct {
	db *gorm.DB
}

func NewStoreRepository(db *gorm.DB) StoreRepository {
	return &storeRepository{db}
}

func (r *storeRepository) Create(store *domain.Store) error {
	return r.db.Create(store).Error
}

func (r *storeRepository) FindCompanyIDByName(name string) (uuid.UUID, error) {
	var company domain.Company
	err := r.db.Where("LOWER(name) = LOWER(?)", name).First(&company).Error
	return company.ID, err
}

func (r *storeRepository) UpdateRegion(id uuid.UUID, province, regency, district, village string) error {
	return r.db.Model(&domain.Store{}).Where("id = ?", id).Updates(map[string]interface{}{
		"province": province,
		"regency":  regency,
		"district": district,
		"village":  village,
	}).Error
}

func (r *storeRepository) FindAll() ([]domain.Store, error) {
	var stores []domain.Store
	query := `
		SELECT s.*, 
			COALESCE((SELECT SUM(total_amount - payment_collected_amount) FROM sales_orders WHERE store_id = s.id AND status NOT IN ('PAID', 'CANCELLED', 'REJECTED', 'DRAFT')), 0) as total_receivable,
			(SELECT COUNT(*) FROM sales_orders WHERE store_id = s.id AND status NOT IN ('PAID', 'CANCELLED', 'REJECTED', 'DRAFT') AND received_at < NOW() - INTERVAL '7 days') as overdue_count
		FROM stores s
		WHERE s.is_active = true
	`
	if err := r.db.Raw(query).Preload("Company").Preload("AssignedEmployee").Find(&stores).Error; err != nil {
		return nil, err
	}
	return stores, nil
}

func (r *storeRepository) FindByID(id uuid.UUID) (*domain.Store, error) {
	var store domain.Store
	query := `
		SELECT s.*, 
			COALESCE((SELECT SUM(total_amount - payment_collected_amount) FROM sales_orders WHERE store_id = s.id AND status NOT IN ('PAID', 'CANCELLED', 'REJECTED', 'DRAFT')), 0) as total_receivable,
			(SELECT COUNT(*) FROM sales_orders WHERE store_id = s.id AND status NOT IN ('PAID', 'CANCELLED', 'REJECTED', 'DRAFT') AND received_at < NOW() - INTERVAL '7 days') as overdue_count
		FROM stores s WHERE s.id = ?
	`
	if err := r.db.Raw(query, id).Preload("Company").Preload("AssignedEmployee").First(&store).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &store, nil
}

func (r *storeRepository) Update(store *domain.Store) error {
	return r.db.Save(store).Error
}

func (r *storeRepository) Delete(id uuid.UUID) error {
	err := r.db.Where("id = ?", id).Delete(&domain.Store{}).Error
	if err != nil {
		if strings.Contains(err.Error(), "violates foreign key constraint") || strings.Contains(err.Error(), "SQLSTATE 23503") {
			return errors.New("toko tidak bisa dihapus karena masih terhubung dengan data transaksi/kunjungan. Silakan edit dan nonaktifkan (IsActive) toko ini sebagai gantinya")
		}
		return err
	}
	return nil
}

func (r *storeRepository) FindByVisitDay(day int, employeeID uuid.UUID) ([]domain.Store, error) {
	var stores []domain.Store
	query := `
		SELECT s.*, 
			COALESCE((SELECT SUM(total_amount - payment_collected_amount) FROM sales_orders WHERE store_id = s.id AND status NOT IN ('PAID', 'CANCELLED', 'REJECTED', 'DRAFT')), 0) as total_receivable,
			(SELECT COUNT(*) FROM sales_orders WHERE store_id = s.id AND status NOT IN ('PAID', 'CANCELLED', 'REJECTED', 'DRAFT') AND received_at < NOW() - INTERVAL '7 days') as overdue_count
		FROM stores s WHERE ? = ANY(visit_days) AND assigned_employee_id = ? AND s.is_active = true
	`
	if err := r.db.Raw(query, day, employeeID).Find(&stores).Error; err != nil {
		return nil, err
	}
	return stores, nil
}

func (r *storeRepository) FindByCompanyID(companyID uuid.UUID) ([]domain.Store, error) {
	var stores []domain.Store
	query := `
		SELECT s.*, 
			COALESCE((SELECT SUM(total_amount - payment_collected_amount) FROM sales_orders WHERE store_id = s.id AND status NOT IN ('PAID', 'CANCELLED', 'REJECTED', 'DRAFT')), 0) as total_receivable,
			(SELECT COUNT(*) FROM sales_orders WHERE store_id = s.id AND status NOT IN ('PAID', 'CANCELLED', 'REJECTED', 'DRAFT') AND received_at < NOW() - INTERVAL '7 days') as overdue_count
		FROM stores s WHERE company_id = ? AND s.is_active = true
	`
	if err := r.db.Raw(query, companyID).Preload("Company").Preload("AssignedEmployee").Find(&stores).Error; err != nil {
		return nil, err
	}
	return stores, nil
}

func (r *storeRepository) FindPaginated(companyID *uuid.UUID, branchID *uuid.UUID, assignedEmpID *uuid.UUID, category string, search string, status string, sortBy string, limit int, offset int, salesType string, completeness string, includeMetrics ...bool) ([]domain.Store, int64, error) {
	var stores []domain.Store
	var total int64

	baseQuery := "FROM stores s WHERE 1=1"
	var args []interface{}

	if branchID != nil {
		var branch domain.Branch
		if err := r.db.First(&branch, "id = ?", branchID).Error; err == nil {
			branchNameStr := "%" + branch.Name + "%"
			cleanName := strings.ReplaceAll(strings.ToLower(branch.Name), "kabupaten ", "")
			cleanName = strings.ReplaceAll(cleanName, "kota ", "")
			cleanStr := "%" + strings.TrimSpace(cleanName) + "%"

			if companyID != nil {
				baseQuery += ` AND (
					s.assigned_employee_id IN (
						SELECT e.id FROM employees e WHERE e.branch_id = ? OR e.operating_branch_id = ?
					)
					OR (s.assigned_employee_id IS NULL AND s.company_id = ? AND (s.address ILIKE ? OR s.address ILIKE ?))
				)`
				args = append(args, *branchID, *branchID, *companyID, branchNameStr, cleanStr)
			} else {
				baseQuery += ` AND (
					s.assigned_employee_id IN (
						SELECT e.id FROM employees e WHERE e.branch_id = ? OR e.operating_branch_id = ?
					)
					OR (s.assigned_employee_id IS NULL AND s.company_id IN (SELECT company_id FROM branches WHERE id = ?) AND (s.address ILIKE ? OR s.address ILIKE ?))
				)`
				args = append(args, *branchID, *branchID, *branchID, branchNameStr, cleanStr)
			}
		} else {
			if companyID != nil {
				baseQuery += ` AND (
					s.assigned_employee_id IN (SELECT e.id FROM employees e WHERE COALESCE(e.operating_branch_id, e.branch_id) = ?)
					OR (s.assigned_employee_id IS NULL AND s.company_id = ?)
				)`
				args = append(args, *branchID, *companyID)
			} else {
				baseQuery += ` AND (
					s.assigned_employee_id IN (SELECT e.id FROM employees e WHERE COALESCE(e.operating_branch_id, e.branch_id) = ?)
					OR (s.assigned_employee_id IS NULL AND s.company_id IN (SELECT company_id FROM branches WHERE id = ?))
				)`
				args = append(args, *branchID, *branchID)
			}
		}
	} else if companyID != nil {
		baseQuery += ` AND (
			s.company_id = ? 
			OR s.assigned_employee_id IN (
				SELECT e.id FROM employees e WHERE e.company_id = ? OR e.branch_id IN (SELECT b.id FROM branches b WHERE b.company_id = ?)
			)
		)`
		args = append(args, *companyID, *companyID, *companyID)
	}
	if category != "" && category != "ALL" {
		if category == "TOKO_BARU" || category == "TOKO_LAMA" {
			baseQuery += " AND s.category = ?"
			args = append(args, category)
		} else if category == "NOO_TODAY" {
			baseQuery += " AND s.created_at::date = CURRENT_DATE"
		} else if category == "EC_TODAY" {
			baseQuery += ` AND (
				s.id IN (SELECT store_id FROM sales_orders WHERE created_at::date = CURRENT_DATE AND status NOT IN ('CANCELLED', 'REJECTED', 'DRAFT'))
				OR s.id IN (SELECT store_id FROM sales_transactions WHERE transaction_date::date = CURRENT_DATE AND status NOT IN ('CANCELLED', 'REJECTED'))
			)`
		} else if category == "NON_EC_TODAY" {
			baseQuery += ` AND s.id NOT IN (SELECT store_id FROM sales_orders WHERE created_at::date = CURRENT_DATE AND status NOT IN ('CANCELLED', 'REJECTED', 'DRAFT'))
				AND s.id NOT IN (SELECT store_id FROM sales_transactions WHERE transaction_date::date = CURRENT_DATE AND status NOT IN ('CANCELLED', 'REJECTED'))`
		} else {
			baseQuery += " AND s.customer_type = ?"
			args = append(args, category)
		}
	}
	if assignedEmpID != nil {
		baseQuery += " AND s.assigned_employee_id = ?"
		args = append(args, *assignedEmpID)
	} else if salesType != "" && salesType != "ALL_SALES" {
		var jobFilter string
		if salesType == "sales_mt" {
			jobFilter = "jp.title ILIKE '%sales mt%' OR jp.title ILIKE '%modern trade%' OR jp.title ILIKE '%kam%' OR jp.title ILIKE '%key account%'"
		} else if salesType == "sales_to" {
			jobFilter = "jp.title ~* '(^|[^[:alnum:]])(to|take order|task order)([^[:alnum:]]|$)'"
		} else if salesType == "sales_motoris_booster" {
			jobFilter = "jp.title ILIKE '%motoris%' OR jp.title ILIKE '%booster%'"
		}

		if jobFilter != "" {
			baseQuery += ` AND (
				s.assigned_employee_id IN (SELECT e.id FROM employees e JOIN job_positions jp ON jp.id = e.job_position_id WHERE ` + jobFilter + `)
				OR s.id IN (SELECT st.store_id FROM sales_transactions st JOIN employees e ON e.id = st.employee_id JOIN job_positions jp ON jp.id = e.job_position_id WHERE ` + jobFilter + `)
				OR s.id IN (SELECT so.store_id FROM sales_orders so JOIN employees e ON e.id = so.employee_id JOIN job_positions jp ON jp.id = e.job_position_id WHERE ` + jobFilter + `)
				OR s.branding_shop_id IS NOT NULL
			)`
		}
	}
	if status == "ACTIVE" {
		baseQuery += " AND s.is_active = true"
	} else if status == "INACTIVE" {
		baseQuery += " AND s.is_active = false"
	}
	if search != "" {
		baseQuery += " AND (s.name ILIKE ? OR s.owner_name ILIKE ? OR s.address ILIKE ?)"
		likeSearch := "%" + search + "%"
		args = append(args, likeSearch, likeSearch, likeSearch)
	}

	// Completeness filter
	switch completeness {
	case "INCOMPLETE":
		baseQuery += ` AND (
			s.phone_number IS NULL OR TRIM(s.phone_number) IN ('', '-')
			OR s.photo_url IS NULL OR TRIM(s.photo_url) = ''
			OR (s.latitude = 0 AND s.longitude = 0) OR s.latitude IS NULL OR s.longitude IS NULL
			OR s.owner_name IS NULL OR TRIM(s.owner_name) IN ('', '-')
			OR s.address IS NULL OR TRIM(s.address) = ''
			OR s.assigned_employee_id IS NULL
		)`
	case "COMPLETE":
		baseQuery += ` AND (
			s.phone_number IS NOT NULL AND TRIM(s.phone_number) NOT IN ('', '-')
			AND s.photo_url IS NOT NULL AND TRIM(s.photo_url) != ''
			AND NOT ((s.latitude = 0 AND s.longitude = 0) OR s.latitude IS NULL OR s.longitude IS NULL)
			AND s.owner_name IS NOT NULL AND TRIM(s.owner_name) NOT IN ('', '-')
			AND s.address IS NOT NULL AND TRIM(s.address) != ''
			AND s.assigned_employee_id IS NOT NULL
		)`
	case "MISSING_PHONE":
		baseQuery += ` AND (s.phone_number IS NULL OR TRIM(s.phone_number) IN ('', '-'))`
	case "MISSING_LOCATION":
		baseQuery += ` AND ((s.latitude = 0 AND s.longitude = 0) OR s.latitude IS NULL OR s.longitude IS NULL)`
	case "MISSING_PHOTO":
		baseQuery += ` AND (s.photo_url IS NULL OR TRIM(s.photo_url) = '')`
	case "MISSING_OWNER":
		baseQuery += ` AND (s.owner_name IS NULL OR TRIM(s.owner_name) IN ('', '-'))`
	case "MISSING_SALESMAN":
		baseQuery += ` AND s.assigned_employee_id IS NULL`
	case "MISSING_ADDRESS":
		baseQuery += ` AND (s.address IS NULL OR TRIM(s.address) = '' OR s.regency IS NULL OR TRIM(s.regency) = '')`
	}

	countQuery := "SELECT COUNT(*) " + baseQuery
	if err := r.db.Raw(countQuery, args...).Scan(&total).Error; err != nil {
		return nil, 0, err
	}

	withMetrics := true
	if len(includeMetrics) > 0 {
		withMetrics = includeMetrics[0]
	}

	var selectQuery string
	if !withMetrics {
		if sortBy == "last_visit_asc" {
			selectQuery = `
				SELECT s.*, sv.check_in_time as last_visit_date
				` + baseQuery + `
				LEFT JOIN LATERAL (
					SELECT check_in_time 
					FROM sales_visits 
					WHERE store_id = s.id 
					ORDER BY check_in_time DESC 
					LIMIT 1
				) sv ON true
				ORDER BY sv.check_in_time ASC NULLS FIRST, s.created_at DESC
				LIMIT ? OFFSET ?
			`
		} else {
			selectQuery = `
				SELECT s.*
				` + baseQuery + `
				ORDER BY s.created_at DESC
				LIMIT ? OFFSET ?
			`
		}
	} else if sortBy == "last_visit_asc" {
		selectQuery = `
			WITH paginated_stores AS (
				SELECT s.*, sv.check_in_time as last_visit_date
				` + baseQuery + `
				LEFT JOIN LATERAL (
					SELECT check_in_time 
					FROM sales_visits 
					WHERE store_id = s.id 
					ORDER BY check_in_time DESC 
					LIMIT 1
				) sv ON true
				ORDER BY sv.check_in_time ASC NULLS FIRST, s.created_at DESC
				LIMIT ? OFFSET ?
			)
			SELECT ps.*, 
				COALESCE((SELECT SUM(total_amount - payment_collected_amount) FROM sales_orders WHERE store_id = ps.id AND status NOT IN ('PAID', 'CANCELLED', 'REJECTED', 'DRAFT')), 0) as total_receivable,
				(SELECT COUNT(*) FROM sales_orders WHERE store_id = ps.id AND status NOT IN ('PAID', 'CANCELLED', 'REJECTED', 'DRAFT') AND received_at < NOW() - INTERVAL '7 days') as overdue_count,
				(
					(SELECT COUNT(*) FROM sales_orders WHERE store_id = ps.id AND status NOT IN ('CANCELLED', 'REJECTED', 'DRAFT')) +
					(SELECT COUNT(*) FROM sales_transactions WHERE store_id = ps.id AND status NOT IN ('CANCELLED', 'REJECTED'))
				) as transaction_count,
				(
					SELECT jp.title 
					FROM store_audits sa 
					JOIN employees e ON e.id = sa.auditor_id 
					JOIN job_positions jp ON jp.id = e.job_position_id 
					WHERE sa.store_id = ps.id 
					ORDER BY sa.created_at DESC 
					LIMIT 1
				) as latest_verifier_role,
				(ps.created_at::date = CURRENT_DATE) as is_noo_today,
				(SELECT COUNT(*) FROM sales_orders WHERE store_id = ps.id AND created_at::date = CURRENT_DATE AND status NOT IN ('CANCELLED', 'REJECTED', 'DRAFT')) as today_order_count,
				COALESCE((SELECT SUM(total_amount) FROM sales_orders WHERE store_id = ps.id AND created_at::date = CURRENT_DATE AND status NOT IN ('CANCELLED', 'REJECTED', 'DRAFT')), 0) as today_order_amount,
				(
					(SELECT COUNT(*) FROM sales_orders WHERE store_id = ps.id AND created_at::date = CURRENT_DATE AND status NOT IN ('CANCELLED', 'REJECTED', 'DRAFT')) > 0
					OR (SELECT COUNT(*) FROM sales_transactions WHERE store_id = ps.id AND transaction_date::date = CURRENT_DATE AND status NOT IN ('CANCELLED', 'REJECTED')) > 0
				) as has_order_today,
				(
					(SELECT COUNT(*) FROM sales_orders WHERE store_id = ps.id AND created_at::date = CURRENT_DATE AND status NOT IN ('CANCELLED', 'REJECTED', 'DRAFT')) > 0
					OR (SELECT COUNT(*) FROM sales_transactions WHERE store_id = ps.id AND transaction_date::date = CURRENT_DATE AND status NOT IN ('CANCELLED', 'REJECTED')) > 0
				) as is_effective_call
			FROM paginated_stores ps
			ORDER BY ps.last_visit_date ASC NULLS FIRST, ps.created_at DESC
		`
	} else {
		selectQuery = `
			WITH paginated_stores AS (
				SELECT s.*
				` + baseQuery + `
				ORDER BY s.created_at DESC
				LIMIT ? OFFSET ?
			)
			SELECT ps.*, 
				COALESCE((SELECT SUM(total_amount - payment_collected_amount) FROM sales_orders WHERE store_id = ps.id AND status NOT IN ('PAID', 'CANCELLED', 'REJECTED', 'DRAFT')), 0) as total_receivable,
				(SELECT COUNT(*) FROM sales_orders WHERE store_id = ps.id AND status NOT IN ('PAID', 'CANCELLED', 'REJECTED', 'DRAFT') AND received_at < NOW() - INTERVAL '7 days') as overdue_count,
				(SELECT MAX(check_in_time) FROM sales_visits WHERE store_id = ps.id) as last_visit_date,
				(
					(SELECT COUNT(*) FROM sales_orders WHERE store_id = ps.id AND status NOT IN ('CANCELLED', 'REJECTED', 'DRAFT')) +
					(SELECT COUNT(*) FROM sales_transactions WHERE store_id = ps.id AND status NOT IN ('CANCELLED', 'REJECTED'))
				) as transaction_count,
				(
					SELECT jp.title 
					FROM store_audits sa 
					JOIN employees e ON e.id = sa.auditor_id 
					JOIN job_positions jp ON jp.id = e.job_position_id 
					WHERE sa.store_id = ps.id 
					ORDER BY sa.created_at DESC 
					LIMIT 1
				) as latest_verifier_role,
				(ps.created_at::date = CURRENT_DATE) as is_noo_today,
				(SELECT COUNT(*) FROM sales_orders WHERE store_id = ps.id AND created_at::date = CURRENT_DATE AND status NOT IN ('CANCELLED', 'REJECTED', 'DRAFT')) as today_order_count,
				COALESCE((SELECT SUM(total_amount) FROM sales_orders WHERE store_id = ps.id AND created_at::date = CURRENT_DATE AND status NOT IN ('CANCELLED', 'REJECTED', 'DRAFT')), 0) as today_order_amount,
				(
					(SELECT COUNT(*) FROM sales_orders WHERE store_id = ps.id AND created_at::date = CURRENT_DATE AND status NOT IN ('CANCELLED', 'REJECTED', 'DRAFT')) > 0
					OR (SELECT COUNT(*) FROM sales_transactions WHERE store_id = ps.id AND transaction_date::date = CURRENT_DATE AND status NOT IN ('CANCELLED', 'REJECTED')) > 0
				) as has_order_today,
				(
					(SELECT COUNT(*) FROM sales_orders WHERE store_id = ps.id AND created_at::date = CURRENT_DATE AND status NOT IN ('CANCELLED', 'REJECTED', 'DRAFT')) > 0
					OR (SELECT COUNT(*) FROM sales_transactions WHERE store_id = ps.id AND transaction_date::date = CURRENT_DATE AND status NOT IN ('CANCELLED', 'REJECTED')) > 0
				) as is_effective_call
			FROM paginated_stores ps
			ORDER BY ps.created_at DESC
		`
	}

	args = append(args, limit, offset)

	if err := r.db.Raw(selectQuery, args...).Preload("Company").Preload("AssignedEmployee").Preload("AssignedEmployee.JobPosition").Find(&stores).Error; err != nil {
		return nil, 0, err
	}

	return stores, total, nil
}

func (r *storeRepository) GetStoreHistories(storeIDs []uuid.UUID) (map[uuid.UUID]StoreHistory, error) {
	if len(storeIDs) == 0 {
		return make(map[uuid.UUID]StoreHistory), nil
	}

	query := `
		SELECT 
			s.id as store_id,
			sv.check_in_time as last_visit_date,
			st.transaction_date as last_transaction_date,
			st.total_amount as last_transaction_amount,
			si_agg.last_transaction_items
		FROM stores s
		LEFT JOIN LATERAL (
			SELECT check_in_time 
			FROM sales_visits 
			WHERE store_id = s.id 
			ORDER BY check_in_time DESC 
			LIMIT 1
		) sv ON true
		LEFT JOIN LATERAL (
			SELECT id, transaction_date, total_amount 
			FROM sales_transactions 
			WHERE store_id = s.id 
			ORDER BY transaction_date DESC 
			LIMIT 1
		) st ON true
		LEFT JOIN LATERAL (
			SELECT string_agg(p.name || ' (' || si.quantity || ')', ', ') as last_transaction_items
			FROM sales_items si 
			JOIN products p ON si.product_id = p.id 
			WHERE si.sales_transaction_id = st.id
		) si_agg ON true
		WHERE s.id IN ?
	`
	var results []StoreHistory
	if err := r.db.Raw(query, storeIDs).Scan(&results).Error; err != nil {
		return nil, err
	}

	historyMap := make(map[uuid.UUID]StoreHistory)
	for _, res := range results {
		historyMap[res.StoreID] = res
	}
	return historyMap, nil
}

func (r *storeRepository) GetOverdueStores(employeeID uuid.UUID, days int) ([]domain.Store, error) {
	var stores []domain.Store
	// NOW() - (? * INTERVAL '1 day') works for PostgreSQL parameterization
	query := `
		SELECT s.*
		FROM stores s
		LEFT JOIN LATERAL (
			SELECT check_in_time 
			FROM sales_visits 
			WHERE store_id = s.id 
			ORDER BY check_in_time DESC 
			LIMIT 1
		) v ON true
		WHERE s.assigned_employee_id = ?
		AND (v.check_in_time IS NULL OR v.check_in_time < NOW() - (? * INTERVAL '1 day'))
	`
	if err := r.db.Raw(query, employeeID, days).Find(&stores).Error; err != nil {
		return nil, err
	}
	return stores, nil
}

func (r *storeRepository) TransferStores(fromEmployeeID uuid.UUID, toEmployeeID uuid.UUID) (int64, error) {
	result := r.db.Model(&domain.Store{}).Where("assigned_employee_id = ?", fromEmployeeID).Update("assigned_employee_id", toEmployeeID)
	return result.RowsAffected, result.Error
}

func (r *storeRepository) CreateAudit(audit *domain.StoreAudit) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(audit).Error; err != nil {
			return err
		}

		updates := map[string]interface{}{
			"verification_status": audit.Status,
		}
		if audit.PhotoProofURL != "" {
			updates["photo_url"] = audit.PhotoProofURL
		}

		if err := tx.Model(&domain.Store{}).Where("id = ?", audit.StoreID).Updates(updates).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *storeRepository) GetAuditsByStoreID(storeID uuid.UUID) ([]domain.StoreAudit, error) {
	var audits []domain.StoreAudit
	err := r.db.Where("store_id = ?", storeID).Preload("Auditor").Preload("Auditor.JobPosition").Order("created_at desc").Find(&audits).Error
	return audits, err
}

func (r *storeRepository) GetAllAudits(companyID *uuid.UUID, branchID *uuid.UUID, limit int, offset int) ([]domain.StoreAudit, int64, error) {
	var audits []domain.StoreAudit
	var total int64

	query := r.db.Model(&domain.StoreAudit{}).
		Joins("JOIN stores ON stores.id = store_audits.store_id").
		Joins("JOIN employees ON employees.id = store_audits.auditor_id")

	if branchID != nil {
		query = query.Where("employees.branch_id = ? OR employees.operating_branch_id = ?", branchID, branchID)
	} else if companyID != nil {
		query = query.Where("stores.company_id = ?", companyID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Preload("Store").Preload("Auditor").Preload("Auditor.JobPosition").
		Order("store_audits.created_at desc").
		Limit(limit).Offset(offset).
		Find(&audits).Error

	return audits, total, err
}

func (r *storeRepository) CheckDuplicate(req domain.CheckDuplicateStoreRequest, currentEmpID *uuid.UUID) (*domain.CheckDuplicateStoreResponse, error) {
	sanitizedPhone := utils.SanitizePhoneNumber(req.PhoneNumber)
	normalizedName := utils.NormalizeStoreName(req.Name)

	var candidateStores []domain.Store
	query := r.db.Model(&domain.Store{}).Preload("AssignedEmployee").Preload("AssignedEmployee.Branch")

	if req.ExcludeID != nil && *req.ExcludeID != uuid.Nil {
		query = query.Where("id != ?", *req.ExcludeID)
	}

	// Buat filter OR untuk mengambil kandidat yang relevan
	var conditions []string
	var args []interface{}

	// 1. Kondisi nomor telepon
	if sanitizedPhone != "" {
		conditions = append(conditions, "phone_number != ''")
	}

	// 2. Kondisi koordinat geografis (bounding box ~5km)
	if req.Latitude != 0 && req.Longitude != 0 {
		conditions = append(conditions, "(latitude BETWEEN ? AND ? AND longitude BETWEEN ? AND ?)")
		args = append(args, req.Latitude-0.05, req.Latitude+0.05, req.Longitude-0.05, req.Longitude+0.05)
	}

	// 3. Kondisi wilayah kecamatan/kabupaten
	if req.District != "" && req.Regency != "" {
		conditions = append(conditions, "(LOWER(district) = ? AND LOWER(regency) = ?)")
		args = append(args, strings.ToLower(strings.TrimSpace(req.District)), strings.ToLower(strings.TrimSpace(req.Regency)))
	}

	// 4. Kondisi pencarian nama
	if normalizedName != "" {
		words := strings.Fields(normalizedName)
		if len(words) > 0 {
			// Cari kata yang panjangnya >= 3 karakter
			for _, w := range words {
				if len(w) >= 3 {
					conditions = append(conditions, "LOWER(name) LIKE ?")
					args = append(args, "%"+strings.ToLower(w)+"%")
					break
				}
			}
		}
	}

	if len(conditions) == 0 {
		return &domain.CheckDuplicateStoreResponse{
			HasPotentialDuplicate: false,
			Matches:               []domain.DuplicateStoreMatch{},
		}, nil
	}

	whereClause := strings.Join(conditions, " OR ")
	if err := query.Where(whereClause, args...).Limit(100).Find(&candidateStores).Error; err != nil {
		return nil, err
	}

	var matches []domain.DuplicateStoreMatch
	seenIDs := make(map[uuid.UUID]bool)

	for _, cand := range candidateStores {
		if seenIDs[cand.ID] {
			continue
		}
		seenIDs[cand.ID] = true

		candSanitizedPhone := utils.SanitizePhoneNumber(cand.PhoneNumber)
		nameSim := utils.CalculateSimilarity(req.Name, cand.Name)

		var distMeters float64 = -1
		if req.Latitude != 0 && req.Longitude != 0 && cand.Latitude != 0 && cand.Longitude != 0 {
			distMeters = utils.CalculateDistance(req.Latitude, req.Longitude, cand.Latitude, cand.Longitude)
		}

		phoneScore := 0.0
		distScore := 0.0
		nameScore := 0.0
		var reasons []string

		// Evaluasi Telepon
		if sanitizedPhone != "" && candSanitizedPhone != "" && sanitizedPhone == candSanitizedPhone {
			phoneScore = 1.0
			reasons = append(reasons, "Nomor telepon/WA sama persis")
		}

		// Evaluasi Jarak
		if distMeters >= 0 {
			if distMeters <= 30 {
				distScore = 1.0
				reasons = append(reasons, fmt.Sprintf("Titik lokasi sangat dekat (%.0fm)", distMeters))
			} else if distMeters <= 100 {
				distScore = 0.7
				reasons = append(reasons, fmt.Sprintf("Titik lokasi dalam radius %.0fm", distMeters))
			} else if distMeters <= 300 {
				distScore = 0.4
			}
		}

		// Evaluasi Kemiripan Nama
		if nameSim >= 0.85 {
			nameScore = 1.0
			reasons = append(reasons, fmt.Sprintf("Nama toko sangat mirip (%.0f%%)", nameSim*100))
		} else if nameSim >= 0.70 {
			nameScore = 0.7
			reasons = append(reasons, fmt.Sprintf("Nama toko mirip (%.0f%%)", nameSim*100))
		} else if nameSim >= 0.50 {
			nameScore = 0.3
		}

		// Hitung Composite Score
		totalScore := (phoneScore * 0.40) + (distScore * 0.35) + (nameScore * 0.25)
		if phoneScore == 1.0 {
			if totalScore < 0.85 {
				totalScore = 0.85
			}
		}
		if distScore == 1.0 && nameScore >= 0.7 {
			if totalScore < 0.90 {
				totalScore = 0.90
			}
		}

		// Ambang batas lolos indikasi duplikat
		if totalScore >= 0.45 && len(reasons) > 0 {
			match := domain.DuplicateStoreMatch{
				StoreID:        cand.ID,
				Name:           cand.Name,
				PhoneNumber:    cand.PhoneNumber,
				Address:        cand.Address,
				DistanceMeters: distMeters,
				NameSimilarity: nameSim,
				TotalScore:     totalScore,
				MatchReasons:   reasons,
				PhotoURL:       cand.PhotoURL,
				IsActive:       cand.IsActive,
			}

			if cand.AssignedEmployee != nil {
				match.AssignedEmployeeID = &cand.AssignedEmployee.ID
				match.AssignedEmployeeName = strings.TrimSpace(cand.AssignedEmployee.FirstName + " " + cand.AssignedEmployee.LastName)
				if cand.AssignedEmployee.Branch != nil {
					match.AssignedBranchName = cand.AssignedEmployee.Branch.Name
				}
				if currentEmpID != nil && *currentEmpID == cand.AssignedEmployee.ID {
					match.IsAssignedToMe = true
				}
			}

			matches = append(matches, match)
		}
	}

	// Urutkan dari skor tertinggi
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].TotalScore > matches[j].TotalScore
	})

	// Batasi maksimal 5 rekomendasi
	if len(matches) > 5 {
		matches = matches[:5]
	}

	return &domain.CheckDuplicateStoreResponse{
		HasPotentialDuplicate: len(matches) > 0,
		Matches:               matches,
	}, nil
}

func (r *storeRepository) GetStoreDistribution(companyID *uuid.UUID, branchIDs []uuid.UUID, assignedEmpID *uuid.UUID, salesType string, status string) ([]domain.StoreDistributionItem, error) {
	var items []domain.StoreDistributionItem

	baseQuery := `
		SELECT 
			s.id,
			s.company_id,
			s.name,
			COALESCE(s.owner_name, '') as owner_name,
			COALESCE(s.phone_number, '') as phone_number,
			COALESCE(s.address, '') as address,
			COALESCE(s.latitude, 0) as latitude,
			COALESCE(s.longitude, 0) as longitude,
			s.is_active,
			COALESCE(s.category, 'TOKO_LAMA') as category,
			COALESCE(s.customer_type, 'TOKO') as customer_type,
			COALESCE(s.province, '') as province,
			COALESCE(s.regency, '') as regency,
			COALESCE(s.district, '') as district,
			COALESCE(s.village, '') as village,
			s.assigned_employee_id,
			COALESCE(e.first_name, '') as employee_first_name,
			COALESCE(e.last_name, '') as employee_last_name,
			COALESCE(e.operating_branch_id, e.branch_id) as branch_id,
			COALESCE(b.name, '') as branch_name,
			s.created_at
		FROM stores s
		LEFT JOIN employees e ON e.id = s.assigned_employee_id
		LEFT JOIN branches b ON b.id = COALESCE(e.operating_branch_id, e.branch_id)
		WHERE 1=1
	`
	var args []interface{}

	if len(branchIDs) == 1 {
		branchID := branchIDs[0]
		var branch domain.Branch
		if err := r.db.First(&branch, "id = ?", branchID).Error; err == nil {
			branchNameStr := "%" + branch.Name + "%"
			cleanName := strings.ReplaceAll(strings.ToLower(branch.Name), "kabupaten ", "")
			cleanName = strings.ReplaceAll(cleanName, "kota ", "")
			cleanStr := "%" + strings.TrimSpace(cleanName) + "%"

			if companyID != nil {
				baseQuery += ` AND (
					COALESCE(e.operating_branch_id, e.branch_id) = ?
					OR (s.assigned_employee_id IS NULL AND s.company_id = ? AND (s.address ILIKE ? OR s.address ILIKE ?))
				)`
				args = append(args, branchID, *companyID, branchNameStr, cleanStr)
			} else {
				baseQuery += ` AND (
					COALESCE(e.operating_branch_id, e.branch_id) = ?
					OR (s.assigned_employee_id IS NULL AND s.company_id IN (SELECT company_id FROM branches WHERE id = ?) AND (s.address ILIKE ? OR s.address ILIKE ?))
				)`
				args = append(args, branchID, branchID, branchNameStr, cleanStr)
			}
		} else {
			if companyID != nil {
				baseQuery += ` AND (COALESCE(e.operating_branch_id, e.branch_id) = ? OR (s.assigned_employee_id IS NULL AND s.company_id = ?))`
				args = append(args, branchID, *companyID)
			} else {
				baseQuery += ` AND (COALESCE(e.operating_branch_id, e.branch_id) = ? OR (s.assigned_employee_id IS NULL AND s.company_id IN (SELECT company_id FROM branches WHERE id = ?)))`
				args = append(args, branchID, branchID)
			}
		}
	} else if len(branchIDs) > 1 {
		baseQuery += ` AND (
			COALESCE(e.operating_branch_id, e.branch_id) IN (?)
			OR (s.assigned_employee_id IS NULL AND s.company_id IN (SELECT company_id FROM branches WHERE id IN (?)))
		)`
		args = append(args, branchIDs, branchIDs)
	} else if companyID != nil {
		baseQuery += ` AND (s.company_id = ? OR e.company_id = ? OR e.branch_id IN (SELECT id FROM branches WHERE company_id = ?))`
		args = append(args, *companyID, *companyID, *companyID)
	}

	if assignedEmpID != nil {
		baseQuery += " AND s.assigned_employee_id = ?"
		args = append(args, *assignedEmpID)
	} else if salesType != "" && salesType != "ALL_SALES" {
		var jobFilter string
		if salesType == "sales_mt" {
			jobFilter = "jp.title ILIKE '%sales mt%' OR jp.title ILIKE '%modern trade%' OR jp.title ILIKE '%kam%' OR jp.title ILIKE '%key account%'"
		} else if salesType == "sales_to" {
			jobFilter = "jp.title ~* '(^|[^[:alnum:]])(to|take order|task order)([^[:alnum:]]|$)'"
		} else if salesType == "sales_motoris_booster" {
			jobFilter = "jp.title ILIKE '%motoris%' OR jp.title ILIKE '%booster%'"
		}

		if jobFilter != "" {
			baseQuery += ` AND (
				s.assigned_employee_id IN (SELECT emp.id FROM employees emp JOIN job_positions jp ON jp.id = emp.job_position_id WHERE ` + jobFilter + `)
				OR s.id IN (SELECT st.store_id FROM sales_transactions st JOIN employees emp ON emp.id = st.employee_id JOIN job_positions jp ON jp.id = emp.job_position_id WHERE ` + jobFilter + `)
				OR s.id IN (SELECT so.store_id FROM sales_orders so JOIN employees emp ON emp.id = so.employee_id JOIN job_positions jp ON jp.id = emp.job_position_id WHERE ` + jobFilter + `)
				OR s.branding_shop_id IS NOT NULL
			)`
		}
	}

	if status == "ACTIVE" {
		baseQuery += " AND s.is_active = true"
	} else if status == "INACTIVE" {
		baseQuery += " AND s.is_active = false"
	}

	baseQuery += " ORDER BY s.name ASC"

	if err := r.db.Raw(baseQuery, args...).Scan(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

