package repository

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"gorm.io/gorm"
)

// MarketplaceOrderFilter holds filter parameters for querying marketplace orders.
type MarketplaceOrderFilter struct {
	BranchID         *uuid.UUID
	Marketplace      string
	Status           string
	Keyword          string
	ProductName      string
	WarehouseName    string
	ShippingProvider string
	From             *time.Time
	To               *time.Time
	Page             int
	PageSize         int
}

type MarketplaceOrderStatusStats struct {
	AwaitingCollection int64 `json:"awaiting_collection"`
	Packed             int64 `json:"packed"`
	Canceled           int64 `json:"canceled"`
}

// MarketplaceSettlementFilter holds filter parameters for querying settlement & finance orders.
type MarketplaceSettlementFilter struct {
	BranchID         *uuid.UUID
	Marketplace      string
	Tab              string // "ALL", "SETTLED", "HOLD", "REFUND_RETURN"
	SettlementStatus string
	ReturnStatus     string
	PackingStatus    string // "ALL", "PACKED", "UNPACKED", "NEEDS_RESTOCK", "RESTOCKED"
	Keyword          string
	ProductName      string
	WarehouseName    string
	ShippingProvider string
	DateType         string // "settlement_date" or "created_at"
	From             *time.Time
	To               *time.Time
	Page             int
	PageSize         int
}

// Older code could set SETTLED from order age alone. These rows have no finance evidence.
const verifiedMarketplaceSettlementSQL = "(finance_synced_at IS NOT NULL OR COALESCE(fee_and_tax_amount, 0) > 0 OR COALESCE(platform_fee, 0) > 0)"
const marketplaceSettlementAmountSQL = "CASE WHEN marketplace = 'tiktok' THEN COALESCE(settlement_amount, 0) WHEN settlement_amount > 0 THEN settlement_amount ELSE total_amount - COALESCE(NULLIF(fee_and_tax_amount, 0), platform_fee, 0) END"

// ScanHistoryFilter holds filter parameters for querying scan histories.
type ScanHistoryFilter struct {
	BranchID      *uuid.UUID
	From          *time.Time
	To            *time.Time
	Action        string
	WarehouseName string
}

type MarketplaceOpnameSummary struct {
	TotalIn           int64   `json:"total_in"`
	TotalOut          int64   `json:"total_out"`
	TotalAdj          int64   `json:"total_adj"`
	TotalTransactions int64   `json:"total_transactions"`
	TotalOmzet        float64 `json:"total_omzet"`
}

type MarketplaceRepository interface {
	CreateStock(stock *domain.MarketplaceStock) error
	GetStockByID(id uuid.UUID) (*domain.MarketplaceStock, error)
	GetAllStocks(branchID uuid.UUID) ([]domain.MarketplaceStock, error)
	UpdateStock(stock *domain.MarketplaceStock) error
	DeleteStock(id uuid.UUID) error
	CreateOpname(opname *domain.MarketplaceOpname) error
	GetOpnameByID(id uuid.UUID) (*domain.MarketplaceOpname, error)
	UpdateOpname(opname *domain.MarketplaceOpname) error
	GetOpnames(branchID uuid.UUID, stockID *uuid.UUID, opnameType string, from, to *time.Time, search string, limit, offset int, sourceType string) ([]domain.MarketplaceOpname, int64, MarketplaceOpnameSummary, error)
	GetStockStats(branchID uuid.UUID, period string, dateFilter string) (totalSKU int64, criticalCount int64, periodIn int64, periodOut int64, periodOmzet float64, chartData []map[string]interface{}, err error)
	GetFactoryProducts(companyID uuid.UUID) ([]domain.Product, error)
	// Marketplace order & product sync methods
	GetMarketplaceOrders(params MarketplaceOrderFilter) ([]domain.MarketplaceOrderGroup, int64, float64, int64, int64, error)
	GetMarketplaceOrderStatusStats(params MarketplaceOrderFilter) (MarketplaceOrderStatusStats, error)
	GetOrderFilters(branchID *uuid.UUID) (*domain.OrderFiltersResponse, error)
	UpsertMarketplaceOrder(order *domain.MarketplaceOrder) error
	GetMarketplaceProducts(branchID *uuid.UUID) ([]domain.MarketplaceProduct, error)
	UpsertMarketplaceProduct(product *domain.MarketplaceProduct) error
	PackOrder(orderID string, adminUsername string, req domain.PackOrderReq) error
	VerifyOrder(orderID string, adminUsername string) error
	GetProductMappings(skuID string, branchID *uuid.UUID) ([]domain.MarketplaceProductMapping, error)
	SaveProductMappings(skuID string, branchID *uuid.UUID, mappings []domain.MarketplaceProductMapping) error
	GetScanHistories(filter ScanHistoryFilter) ([]domain.MarketplaceScanHistory, error)
	GetScanHistoryWarehouses(branchID *uuid.UUID) ([]string, error)

	// Marketplace Settlement & Finance methods
	GetSettlementOrders(params MarketplaceSettlementFilter) ([]domain.MarketplaceOrderGroup, int64, float64, float64, float64, error)
	GetSettlementSummary(branchID *uuid.UUID, from, to *time.Time, marketplace string, dateType string) (*domain.SettlementSummaryResponse, error)
	RestockOrder(orderID string, adminUsername, notes string) error
	BatchImportSettlements(records []domain.SettlementImportRecord, adminUsername string) (int, error)
	SaveFinanceTransactions(records []domain.FinanceTransactionRecord) ([]string, error)
	UpdateSettlementOrder(orderID string, req domain.UpdateSettlementOrderReq) error
	AutoClassifySettlements(branchID *uuid.UUID) (int64, error)
	CleanNoiseAndRestoreScannedOrders(branchID *uuid.UUID) error
}

type marketplaceRepository struct {
	db *gorm.DB
}

func NewMarketplaceRepository(db *gorm.DB) MarketplaceRepository {
	return &marketplaceRepository{db}
}

func (r *marketplaceRepository) CreateStock(stock *domain.MarketplaceStock) error {
	return r.db.Create(stock).Error
}

func (r *marketplaceRepository) GetStockByID(id uuid.UUID) (*domain.MarketplaceStock, error) {
	var stock domain.MarketplaceStock
	err := r.db.Preload("Product").First(&stock, "id = ?", id).Error
	return &stock, err
}

func (r *marketplaceRepository) GetAllStocks(branchID uuid.UUID) ([]domain.MarketplaceStock, error) {
	var stocks []domain.MarketplaceStock
	q := r.db.Where("is_active = true")
	if branchID != uuid.Nil {
		q = q.Where("branch_id = ?", branchID)
	}
	err := q.Preload("Product").Order("created_at desc").Find(&stocks).Error
	return stocks, err
}

func (r *marketplaceRepository) GetFactoryProducts(companyID uuid.UUID) ([]domain.Product, error) {
	var products []domain.Product
	q := r.db.Order("name asc")
	if companyID != uuid.Nil {
		q = q.Where("company_id = ?", companyID)
	}
	err := q.Find(&products).Error
	return products, err
}

func (r *marketplaceRepository) UpdateStock(stock *domain.MarketplaceStock) error {
	return r.db.Save(stock).Error
}

func (r *marketplaceRepository) DeleteStock(id uuid.UUID) error {
	return r.db.Model(&domain.MarketplaceStock{}).Where("id = ?", id).Update("is_active", false).Error
}

func (r *marketplaceRepository) CreateOpname(opname *domain.MarketplaceOpname) error {
	return r.db.Create(opname).Error
}

func (r *marketplaceRepository) GetOpnameByID(id uuid.UUID) (*domain.MarketplaceOpname, error) {
	var opname domain.MarketplaceOpname
	err := r.db.Preload("MarketplaceStock").Where("id = ?", id).First(&opname).Error
	if err != nil {
		return nil, err
	}
	return &opname, nil
}

func (r *marketplaceRepository) UpdateOpname(opname *domain.MarketplaceOpname) error {
	return r.db.Save(opname).Error
}

func isShopeeOrderID(orderID string) bool {
	id := strings.TrimSpace(orderID)
	if strings.HasPrefix(strings.ToUpper(id), "RETUR-") {
		id = id[6:]
	}
	if len(id) < 12 || len(id) > 20 {
		return false
	}
	hasDigitPrefix := len(id) >= 6
	for i := 0; i < 6 && i < len(id); i++ {
		if id[i] < '0' || id[i] > '9' {
			hasDigitPrefix = false
			break
		}
	}
	hasLetter := false
	for _, ch := range id {
		if (ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z') {
			hasLetter = true
			break
		}
	}
	return hasDigitPrefix && hasLetter
}

func (r *marketplaceRepository) GetOpnames(branchID uuid.UUID, stockID *uuid.UUID, opnameType string, from, to *time.Time, search string, limit, offset int, sourceType string) ([]domain.MarketplaceOpname, int64, MarketplaceOpnameSummary, error) {
	var opnames []domain.MarketplaceOpname
	var total int64
	var summary MarketplaceOpnameSummary

	search = strings.TrimSpace(search)
	q := r.db.Model(&domain.MarketplaceOpname{})
	if branchID != uuid.Nil || search != "" {
		q = q.Joins("JOIN marketplace_stocks ms ON ms.id = marketplace_opnames.marketplace_stock_id")
		if branchID != uuid.Nil {
			q = q.Where("ms.branch_id = ?", branchID)
		}
		if search != "" {
			searchPattern := "%" + search + "%"
			q = q.Where("ms.product_name ILIKE ? OR ms.sku ILIKE ? OR marketplace_opnames.reference ILIKE ? OR marketplace_opnames.notes ILIKE ? OR marketplace_opnames.created_by ILIKE ?",
				searchPattern, searchPattern, searchPattern, searchPattern, searchPattern)
		}
	}
	if stockID != nil {
		q = q.Where("marketplace_opnames.marketplace_stock_id = ?", *stockID)
	}
	if opnameType != "" && opnameType != "ALL" {
		q = q.Where("marketplace_opnames.type = ?", opnameType)
	}
	if sourceType != "" && sourceType != "ALL" {
		switch strings.ToUpper(sourceType) {
		case "TIKTOK":
			q = q.Where(`(
				UPPER(marketplace_opnames.source_type) = 'TIKTOK' 
				OR (
					marketplace_opnames.marketplace_order_id IS NOT NULL 
					AND marketplace_opnames.marketplace_order_id <> '' 
					AND EXISTS (
						SELECT 1 FROM marketplace_orders mo 
						WHERE mo.order_id = marketplace_opnames.marketplace_order_id 
						  AND (LOWER(mo.marketplace) = 'tiktok' OR mo.marketplace IS NULL OR mo.marketplace = '')
					)
				)
			)
			AND NOT (
				marketplace_opnames.reference ~ '^[0-9]{6}[0-9A-Za-z]{6,14}$' 
				AND marketplace_opnames.reference ~ '[A-Za-z]'
			)
			AND NOT (
				marketplace_opnames.marketplace_order_id ~ '^[0-9]{6}[0-9A-Za-z]{6,14}$' 
				AND marketplace_opnames.marketplace_order_id ~ '[A-Za-z]'
			)
			AND NOT (
				marketplace_opnames.marketplace_order_id IS NOT NULL
				AND EXISTS (
					SELECT 1 FROM marketplace_orders mo 
					WHERE mo.order_id = marketplace_opnames.marketplace_order_id 
					  AND LOWER(mo.marketplace) = 'shopee'
				)
			)`)
		case "SHOPEE":
			q = q.Where(`(
				UPPER(marketplace_opnames.source_type) = 'SHOPEE' 
				OR (
					marketplace_opnames.marketplace_order_id IS NOT NULL 
					AND marketplace_opnames.marketplace_order_id <> '' 
					AND EXISTS (
						SELECT 1 FROM marketplace_orders mo 
						WHERE mo.order_id = marketplace_opnames.marketplace_order_id 
						  AND LOWER(mo.marketplace) = 'shopee'
					)
				)
				OR (
					marketplace_opnames.reference ~ '^[0-9]{6}[0-9A-Za-z]{6,14}$' 
					AND marketplace_opnames.reference ~ '[A-Za-z]'
				)
				OR (
					marketplace_opnames.marketplace_order_id ~ '^[0-9]{6}[0-9A-Za-z]{6,14}$' 
					AND marketplace_opnames.marketplace_order_id ~ '[A-Za-z]'
				)
				OR (
					marketplace_opnames.reference ~ '^RETUR-[0-9]{6}[0-9A-Za-z]{6,14}$' 
					AND marketplace_opnames.reference ~ '[A-Za-z]'
				)
			)`)
		case "MANUAL":
			q = q.Where("(UPPER(marketplace_opnames.source_type) = 'MANUAL' OR marketplace_opnames.source_type IS NULL OR marketplace_opnames.source_type = '' OR marketplace_opnames.marketplace_order_id IS NULL OR marketplace_opnames.marketplace_order_id = '') AND UPPER(COALESCE(marketplace_opnames.source_type, '')) NOT IN ('TIKTOK', 'SHOPEE')")
		default:
			q = q.Where("marketplace_opnames.source_type = ?", sourceType)
		}
	}
	if from != nil {
		q = q.Where("marketplace_opnames.created_at >= ?", *from)
	}
	if to != nil {
		q = q.Where("marketplace_opnames.created_at <= ?", *to)
	}

	err := q.Count(&total).Error
	if err != nil {
		return nil, 0, summary, err
	}
	summary.TotalTransactions = total

	// Summary calculation across all filtered data (regardless of pagination limit/offset)
	type summaryRow struct {
		Type       domain.MarketplaceOpnameType
		TotalQty   int64
		TotalCount int64
		TotalOmzet float64
	}
	var rows []summaryRow
	qSum := q.Session(&gorm.Session{})
	errSum := qSum.Select("marketplace_opnames.type, COALESCE(SUM(marketplace_opnames.quantity), 0) as total_qty, COUNT(*) as total_count, COALESCE(SUM(marketplace_opnames.total_value), 0) as total_omzet").Group("marketplace_opnames.type").Scan(&rows).Error
	if errSum == nil {
		for _, row := range rows {
			switch row.Type {
			case domain.OpnameStockIn:
				summary.TotalIn = row.TotalQty
			case domain.OpnameStockOut:
				summary.TotalOut = row.TotalQty
				summary.TotalOmzet = row.TotalOmzet
			case domain.OpnameAdjustment:
				summary.TotalAdj = row.TotalCount
			}
		}
	}

	err = q.Preload("MarketplaceStock").Preload("MarketplaceStock.Product").Order("marketplace_opnames.created_at desc").Limit(limit).Offset(offset).Find(&opnames).Error
	return opnames, total, summary, err
}

func (r *marketplaceRepository) GetStockStats(branchID uuid.UUID, period string, dateFilter string) (totalSKU int64, criticalCount int64, periodIn int64, periodOut int64, periodOmzet float64, chartData []map[string]interface{}, err error) {
	q := r.db.Model(&domain.MarketplaceStock{}).Where("is_active = true")
	if branchID != uuid.Nil {
		q = q.Where("branch_id = ?", branchID)
	}
	q.Count(&totalSKU)

	qc := r.db.Model(&domain.MarketplaceStock{}).Where("is_active = true AND quantity <= min_stock")
	if branchID != uuid.Nil {
		qc = qc.Where("branch_id = ?", branchID)
	}
	qc.Count(&criticalCount)

	loc, errLoc := time.LoadLocation("Asia/Jakarta")
	if errLoc != nil {
		loc = time.FixedZone("WIB", 7*3600)
	}
	now := time.Now().In(loc)
	var startTime time.Time
	var endTime time.Time

	if period == "monthly" {
		if dateFilter != "" {
			// If dateFilter is provided (e.g. YYYY-MM), use it, otherwise use current month
			parsedDate, err := time.ParseInLocation("2006-01-02", dateFilter, loc)
			if err == nil {
				startTime = time.Date(parsedDate.Year(), parsedDate.Month(), 1, 0, 0, 0, 0, loc)
			} else {
				startTime = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc)
			}
		} else {
			startTime = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc)
		}
		endTime = startTime.AddDate(0, 1, 0)
	} else {
		if dateFilter != "" {
			parsedDate, err := time.ParseInLocation("2006-01-02", dateFilter, loc)
			if err == nil {
				startTime = time.Date(parsedDate.Year(), parsedDate.Month(), parsedDate.Day(), 0, 0, 0, 0, loc)
			} else {
				startTime = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
			}
		} else {
			startTime = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
		}
		endTime = startTime.AddDate(0, 0, 1)
	}

	qIn := r.db.Model(&domain.MarketplaceOpname{}).Where("marketplace_opnames.type = ? AND marketplace_opnames.created_at >= ? AND marketplace_opnames.created_at < ?", domain.OpnameStockIn, startTime, endTime)
	if branchID != uuid.Nil {
		qIn = qIn.Joins("JOIN marketplace_stocks ms ON ms.id = marketplace_opnames.marketplace_stock_id").Where("ms.branch_id = ?", branchID)
	}
	qIn.Count(&periodIn)

	qOut := r.db.Model(&domain.MarketplaceOpname{}).Where("marketplace_opnames.type = ? AND marketplace_opnames.created_at >= ? AND marketplace_opnames.created_at < ?", domain.OpnameStockOut, startTime, endTime)
	if branchID != uuid.Nil {
		qOut = qOut.Joins("JOIN marketplace_stocks ms ON ms.id = marketplace_opnames.marketplace_stock_id").Where("ms.branch_id = ?", branchID)
	}
	qOut.Count(&periodOut)

	// Calculate Omzet with fallback to ms.price and products.selling_price
	var omzet float64
	qOmzet := r.db.Table("marketplace_opnames").
		Select("COALESCE(SUM(CASE WHEN marketplace_opnames.total_value > 0 THEN marketplace_opnames.total_value WHEN ms.price > 0 THEN marketplace_opnames.quantity * ms.price WHEN p.selling_price > 0 THEN marketplace_opnames.quantity * p.selling_price ELSE 0 END), 0)").
		Joins("JOIN marketplace_stocks ms ON ms.id = marketplace_opnames.marketplace_stock_id").
		Joins("LEFT JOIN products p ON p.id = ms.product_id").
		Where("marketplace_opnames.type = ? AND marketplace_opnames.created_at >= ? AND marketplace_opnames.created_at < ?", domain.OpnameStockOut, startTime, endTime)

	if branchID != uuid.Nil {
		qOmzet = qOmzet.Where("ms.branch_id = ?", branchID)
	}
	qOmzet.Scan(&omzet)
	periodOmzet = omzet

	// Generate chart data if period is monthly
	if period == "monthly" {
		chartData = []map[string]interface{}{}

		qChart := r.db.Table("marketplace_opnames").
			Select("DATE(marketplace_opnames.created_at) as date, COALESCE(SUM(CASE WHEN marketplace_opnames.total_value > 0 THEN marketplace_opnames.total_value WHEN ms.price > 0 THEN marketplace_opnames.quantity * ms.price WHEN p.selling_price > 0 THEN marketplace_opnames.quantity * p.selling_price ELSE 0 END), 0) as omzet").
			Joins("JOIN marketplace_stocks ms ON ms.id = marketplace_opnames.marketplace_stock_id").
			Joins("LEFT JOIN products p ON p.id = ms.product_id").
			Where("marketplace_opnames.type = ? AND marketplace_opnames.created_at >= ? AND marketplace_opnames.created_at < ?", domain.OpnameStockOut, startTime, endTime)

		if branchID != uuid.Nil {
			qChart = qChart.Where("ms.branch_id = ?", branchID)
		}

		rows, errChart := qChart.Group("DATE(marketplace_opnames.created_at)").Order("DATE(marketplace_opnames.created_at) ASC").Rows()
		if errChart == nil {
			defer rows.Close()
			for rows.Next() {
				var dateStr string
				var dailyOmzet float64
				if err := rows.Scan(&dateStr, &dailyOmzet); err == nil {
					// Extract the first 10 characters (YYYY-MM-DD) just in case it returns full timestamp string
					if len(dateStr) >= 10 {
						dateStr = dateStr[:10]
					}
					chartData = append(chartData, map[string]interface{}{
						"date":  dateStr,
						"omzet": dailyOmzet,
					})
				}
			}
		}
	}

	return
}

func (r *marketplaceRepository) marketplaceOrdersQuery(params MarketplaceOrderFilter) *gorm.DB {
	q := r.db.Model(&domain.MarketplaceOrder{})

	if params.BranchID != nil && *params.BranchID != uuid.Nil {
		q = q.Where("branch_id = ?", *params.BranchID)
	}
	if params.Marketplace != "" {
		q = q.Where("LOWER(marketplace) = ?", strings.ToLower(params.Marketplace))
	}
	if params.Status != "" {
		if strings.Contains(params.Status, ",") {
			parts := strings.Split(params.Status, ",")
			var cleanParts []string
			for _, p := range parts {
				if t := strings.TrimSpace(p); t != "" {
					cleanParts = append(cleanParts, t)
				}
			}
			if len(cleanParts) > 0 {
				q = q.Where("order_status IN ?", cleanParts)
			}
		} else {
			if params.Status == "PACKED" {
				q = q.Where("order_status IN ('PACKED', 'VERIFIED', 'COMPLETED')")
			} else if params.Status == "CANCELED" || params.Status == "CANCELLED" {
				q = q.Where("order_status IN ('CANCELED', 'CANCELLED')")
			} else {
				q = q.Where("order_status = ?", params.Status)
			}
		}
	}
	if params.Keyword != "" {
		kw := "%" + strings.ToLower(params.Keyword) + "%"
		q = q.Where(
			"LOWER(order_id) LIKE ? OR LOWER(product_name) LIKE ? OR LOWER(seller_sku) LIKE ? OR LOWER(warehouse_name) LIKE ? OR LOWER(shipping_provider) LIKE ? OR LOWER(tracking_id) LIKE ?",
			kw, kw, kw, kw, kw, kw,
		)
	}
	if params.ProductName != "" {
		q = q.Where("product_name ILIKE ?", "%"+params.ProductName+"%")
	}
	if params.WarehouseName != "" {
		q = q.Where("warehouse_name ILIKE ?", "%"+params.WarehouseName+"%")
	}
	if params.ShippingProvider != "" {
		q = q.Where("shipping_provider ILIKE ?", "%"+params.ShippingProvider+"%")
	}
	if params.From != nil && params.To != nil {
		q = q.Where(`(
			(created_at >= ? AND created_at <= ?)
			OR (synced_at >= ? AND synced_at <= ?)
			OR order_id IN (
				SELECT order_id FROM marketplace_scan_histories
				WHERE created_at >= ? AND created_at <= ?
			)
		)`, *params.From, *params.To, *params.From, *params.To, *params.From, *params.To)
	} else if params.From != nil {
		q = q.Where(`(
			created_at >= ? 
			OR synced_at >= ?
			OR order_id IN (
				SELECT order_id FROM marketplace_scan_histories
				WHERE created_at >= ?
			)
		)`, *params.From, *params.From, *params.From)
	} else if params.To != nil {
		q = q.Where(`(
			created_at <= ? 
			OR synced_at <= ?
			OR order_id IN (
				SELECT order_id FROM marketplace_scan_histories
				WHERE created_at <= ?
			)
		)`, *params.To, *params.To, *params.To)
	}
	return q
}

// Status cards use the same filters as the list except the selected status.
// Count all three groups in one query instead of loading three more order pages.
func (r *marketplaceRepository) GetMarketplaceOrderStatusStats(params MarketplaceOrderFilter) (MarketplaceOrderStatusStats, error) {
	params.Status = ""
	var stats MarketplaceOrderStatusStats
	err := r.marketplaceOrdersQuery(params).Select(`
		COUNT(DISTINCT CASE WHEN order_status IN ('AWAITING_COLLECTION', 'AWAITING_SHIPMENT', 'VERIFIED') THEN order_id END) AS awaiting_collection,
		COUNT(DISTINCT CASE WHEN order_status IN ('COMPLETED', 'DELIVERED', 'PACKED') THEN order_id END) AS packed,
		COUNT(DISTINCT CASE WHEN order_status IN ('CANCELED', 'CANCELLED', 'RETURN_SUCCESS') THEN order_id END) AS canceled
	`).Scan(&stats).Error
	return stats, err
}

// GetMarketplaceOrders queries marketplace_orders with optional status/keyword/date/branch filters and pagination.
func (r *marketplaceRepository) GetMarketplaceOrders(params MarketplaceOrderFilter) ([]domain.MarketplaceOrderGroup, int64, float64, int64, int64, error) {
	// Aggregate once per order. The previous count, omzet and platform queries
	// each scanned the same filtered rows separately.
	var summary struct {
		Total       int64   `gorm:"column:total"`
		TotalOmzet  float64 `gorm:"column:total_omzet"`
		TotalTiktok int64   `gorm:"column:total_tiktok"`
		TotalShopee int64   `gorm:"column:total_shopee"`
	}
	perOrder := r.marketplaceOrdersQuery(params).Select(`
		order_id, MAX(total_amount) AS amount,
		BOOL_OR(LOWER(COALESCE(marketplace, 'tiktok')) = 'tiktok') AS is_tiktok,
		BOOL_OR(LOWER(marketplace) = 'shopee') AS is_shopee
	`).Group("order_id")
	if err := r.db.Table("(?) AS orders", perOrder).Select(`
		COUNT(*) AS total,
		COALESCE(SUM(amount), 0) AS total_omzet,
		COUNT(*) FILTER (WHERE is_tiktok) AS total_tiktok,
		COUNT(*) FILTER (WHERE is_shopee) AS total_shopee
	`).Scan(&summary).Error; err != nil {
		return nil, 0, 0, 0, 0, err
	}

	page := params.Page
	if page < 1 {
		page = 1
	}
	pageSize := params.PageSize
	if pageSize < 1 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	// Get distinct order_ids ordered by max(created_at) desc
	var orderIDs []string
	if err := r.marketplaceOrdersQuery(params).Select("order_id").Group("order_id").Order("MAX(created_at) desc").Limit(pageSize).Offset(offset).Pluck("order_id", &orderIDs).Error; err != nil {
		return nil, 0, 0, 0, 0, err
	}

	if len(orderIDs) == 0 {
		return []domain.MarketplaceOrderGroup{}, summary.Total, summary.TotalOmzet, summary.TotalTiktok, summary.TotalShopee, nil
	}

	// Fetch all line items for these order_ids
	var lineItems []domain.MarketplaceOrder
	if err := r.db.Where("order_id IN ?", orderIDs).Order("created_at desc").Find(&lineItems).Error; err != nil {
		return nil, 0, 0, 0, 0, err
	}

	// Group line items by order_id
	groupsMap := make(map[string]*domain.MarketplaceOrderGroup)
	for _, item := range lineItems {
		if _, exists := groupsMap[item.OrderID]; !exists {
			groupsMap[item.OrderID] = &domain.MarketplaceOrderGroup{
				OrderID:                 item.OrderID,
				BranchID:                item.BranchID,
				OrderStatus:             item.OrderStatus,
				WarehouseName:           item.WarehouseName,
				TrackingID:              item.TrackingID,
				Marketplace:             item.Marketplace,
				BuyerUsername:           item.BuyerUsername,
				PaymentMethod:           item.PaymentMethod,
				TotalAmount:             item.TotalAmount,
				Currency:                item.Currency,
				ShippingProvider:        item.ShippingProvider,
				PickedBy:                item.PickedBy,
				PackedBy:                item.PackedBy,
				SettlementStatus:        item.SettlementStatus,
				SettlementDate:          item.SettlementDate,
				EstimatedSettlementDate: item.EstimatedSettlementDate,
				SettlementAmount:        item.SettlementAmount,
				PlatformFee:             item.PlatformFee,
				CommissionFee:           item.CommissionFee,
				ShippingFeeDeduction:    item.ShippingFeeDeduction,
				RefundAmount:            item.RefundAmount,
				RefundReason:            item.RefundReason,
				ReturnStatus:            item.ReturnStatus,
				ReturnItemCount:         item.ReturnItemCount,
				RestockedAt:             item.RestockedAt,
				RestockedBy:             item.RestockedBy,
				SyncedAt:                item.SyncedAt,
				CreatedAt:               item.CreatedAt,
				Items:                   []domain.MarketplaceOrder{},
			}
		} else {
			if groupsMap[item.OrderID].TrackingID == "" && item.TrackingID != "" {
				groupsMap[item.OrderID].TrackingID = item.TrackingID
			}
			if groupsMap[item.OrderID].ShippingProvider == "" && item.ShippingProvider != "" {
				groupsMap[item.OrderID].ShippingProvider = item.ShippingProvider
			}
			if groupsMap[item.OrderID].WarehouseName == "" && item.WarehouseName != "" {
				groupsMap[item.OrderID].WarehouseName = item.WarehouseName
			}
		}
		groupsMap[item.OrderID].Items = append(groupsMap[item.OrderID].Items, item)
	}

	// Reorder groups based on the original orderIDs order
	var groups []domain.MarketplaceOrderGroup
	for _, id := range orderIDs {
		if g, exists := groupsMap[id]; exists {
			groups = append(groups, *g)
		}
	}

	return groups, summary.Total, summary.TotalOmzet, summary.TotalTiktok, summary.TotalShopee, nil
}

func (r *marketplaceRepository) GetOrderFilters(branchID *uuid.UUID) (*domain.OrderFiltersResponse, error) {
	q := r.db.Model(&domain.MarketplaceOrder{})
	if branchID != nil && *branchID != uuid.Nil {
		q = q.Where("branch_id = ?", *branchID)
	}

	var warehouses []string
	if err := q.Where("warehouse_name IS NOT NULL AND warehouse_name != ''").
		Distinct("warehouse_name").
		Pluck("warehouse_name", &warehouses).Error; err != nil {
		return nil, err
	}

	var products []string
	qProd := r.db.Model(&domain.MarketplaceOrder{})
	if branchID != nil && *branchID != uuid.Nil {
		qProd = qProd.Where("branch_id = ?", *branchID)
	}
	if err := qProd.Where("product_name IS NOT NULL AND product_name != ''").
		Distinct("product_name").
		Pluck("product_name", &products).Error; err != nil {
		return nil, err
	}

	var rawShippingProviders []string
	qShip := r.db.Model(&domain.MarketplaceOrder{})
	if branchID != nil && *branchID != uuid.Nil {
		qShip = qShip.Where("branch_id = ?", *branchID)
	}
	if err := qShip.Where("shipping_provider IS NOT NULL AND shipping_provider != '' AND shipping_provider != '-'").
		Distinct("shipping_provider").
		Order("shipping_provider ASC").
		Pluck("shipping_provider", &rawShippingProviders).Error; err != nil {
		return nil, err
	}

	var shippingProviders []string
	for _, p := range rawShippingProviders {
		pClean := strings.TrimSpace(p)
		pLower := strings.ToLower(pClean)
		if pClean == "" || pClean == "-" ||
			strings.Contains(pLower, "standard") ||
			strings.Contains(pLower, "economical") ||
			strings.Contains(pLower, "nextday") ||
			strings.Contains(pLower, "next day") ||
			strings.Contains(pLower, "same day") ||
			strings.Contains(pLower, "instant") {
			continue
		}
		shippingProviders = append(shippingProviders, pClean)
	}

	return &domain.OrderFiltersResponse{
		Warehouses:        warehouses,
		Products:          products,
		ShippingProviders: shippingProviders,
	}, nil
}

// UpsertMarketplaceOrder inserts or updates a marketplace order line item
// identified by (order_id, sku_id, branch_id)
func (r *marketplaceRepository) UpsertMarketplaceOrder(order *domain.MarketplaceOrder) error {
	var existing domain.MarketplaceOrder

	query := r.db.Where("order_id = ?", order.OrderID)
	if order.LineItemID != "" {
		query = query.Where("line_item_id = ?", order.LineItemID)
	} else {
		query = query.Where("sku_id = ?", order.SkuID)
	}

	result := query.First(&existing)
	if result.Error != nil {
		// Set created_at to actual order creation time from marketplace if available
		if order.CreateTimeRaw > 0 {
			order.CreatedAt = time.Unix(order.CreateTimeRaw, 0)
		}

		if order.SyncedAt == nil {
			now := time.Now()
			order.SyncedAt = &now
		}

		// Order fulfillment alone does not prove that the marketplace paid the seller.
		if order.OrderStatus == "CANCELED" || order.OrderStatus == "CANCELLED" || order.OrderStatus == "RETURN_SUCCESS" || order.OrderStatus == "FAILED_DELIVERY" {
			order.SettlementStatus = "REFUNDED"
			if order.RefundAmount == 0 {
				order.RefundAmount = order.TotalAmount
			}
		} else if order.SettlementStatus == "" {
			order.SettlementStatus = "PENDING_SETTLEMENT"
		}
		return r.db.Create(order).Error
	}

	// Preserve PACKED status and packer if order was already packed locally
	if existing.OrderStatus == "PACKED" || existing.PackedBy != "" {
		order.OrderStatus = "PACKED"
		if existing.PackedBy != "" {
			order.PackedBy = existing.PackedBy
		}
		if existing.PickedBy != "" {
			order.PickedBy = existing.PickedBy
		}
	} else if existing.OrderStatus == "VERIFIED" {
		if order.OrderStatus == "PENDING" || order.OrderStatus == "UNPAID" || order.OrderStatus == "AWAITING_SHIPMENT" || order.OrderStatus == "AWAITING_COLLECTION" || order.OrderStatus == "VERIFIED" {
			order.OrderStatus = existing.OrderStatus
		}
	}

	// Auto classify settlement status on upsert if not explicitly set
	settlementStatus := existing.SettlementStatus
	settlementAmount := existing.SettlementAmount
	settlementDate := existing.SettlementDate
	refundAmount := existing.RefundAmount

	// Keep existing settlement details intact if already set by finance API or import
	if order.OrderStatus == "CANCELED" || order.OrderStatus == "CANCELLED" || order.OrderStatus == "RETURN_SUCCESS" || order.OrderStatus == "FAILED_DELIVERY" {
		if settlementStatus == "" || settlementStatus == "PENDING_SETTLEMENT" || settlementStatus == "HOLD" || settlementStatus == "ESCROW" {
			settlementStatus = "REFUNDED"
			if refundAmount == 0 {
				refundAmount = order.TotalAmount
			}
		}
	} else if settlementStatus == "" {
		settlementStatus = "PENDING_SETTLEMENT"
	}

	// Record exists — update all mutable fields explicitly via a map
	updates := map[string]interface{}{
		"order_status":      order.OrderStatus,
		"order_substatus":   order.OrderSubstatus,
		"buyer_username":    order.BuyerUsername,
		"total_amount":      order.TotalAmount,
		"currency":          order.Currency,
		"payment_method":    order.PaymentMethod,
		"create_time_raw":   order.CreateTimeRaw,
		"item_count":        order.ItemCount,
		"marketplace":       order.Marketplace,
		"seller_sku":        order.SellerSKU,
		"product_name":      order.ProductName,
		"warehouse_name":    order.WarehouseName,
		"line_item_id":      order.LineItemID,
		"settlement_status": settlementStatus,
		"settlement_amount": settlementAmount,
		"settlement_date":   settlementDate,
		"refund_amount":     refundAmount,
		"synced_at":         time.Now(),
	}
	if order.TrackingID != "" && order.TrackingID != "-" {
		updates["tracking_id"] = order.TrackingID
	}
	if order.ShippingProvider != "" && order.ShippingProvider != "-" {
		updates["shipping_provider"] = order.ShippingProvider
	}
	if existing.PackedBy != "" {
		updates["packed_by"] = existing.PackedBy
	}
	if existing.PickedBy != "" {
		updates["picked_by"] = existing.PickedBy
	}
	if order.CreateTimeRaw > 0 {
		updates["created_at"] = time.Unix(order.CreateTimeRaw, 0)
	}
	if order.BranchID != nil && *order.BranchID != uuid.Nil {
		updates["branch_id"] = order.BranchID
	}

	return r.db.Model(&existing).Updates(updates).Error
}

// GetMarketplaceProducts returns all marketplace products ordered by product_name asc.
func (r *marketplaceRepository) GetMarketplaceProducts(branchID *uuid.UUID) ([]domain.MarketplaceProduct, error) {
	var products []domain.MarketplaceProduct
	q := r.db.Order("product_name asc")
	if branchID != nil && *branchID != uuid.Nil {
		q = q.Where("branch_id = ?", *branchID)
	}
	err := q.Find(&products).Error
	return products, err
}

// UpsertMarketplaceProduct inserts or updates a marketplace product identified by sku_id and branch_id.
func (r *marketplaceRepository) UpsertMarketplaceProduct(product *domain.MarketplaceProduct) error {
	q := r.db.Where("sku_id = ? AND marketplace = ?", product.SkuID, product.Marketplace)
	if product.BranchID != nil && *product.BranchID != uuid.Nil {
		q = q.Where("branch_id = ?", *product.BranchID)
	}
	return q.Assign(*product).FirstOrCreate(product).Error
}

func (r *marketplaceRepository) PackOrder(orderID string, adminUsername string, req domain.PackOrderReq) error {
	tx := r.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	var orders []domain.MarketplaceOrder
	if err := tx.Where("order_id = ?", orderID).Find(&orders).Error; err != nil {
		tx.Rollback()
		return err
	}

	if len(orders) == 0 {
		tx.Rollback()
		return errors.New("order not found")
	}

	// ── Guard: Validasi semua sku_id punya mapping sebelum pack ──────────
	seenSkuIDs := map[string]bool{}
	var unmappedLabels []string
	for _, order := range orders {
		if order.SkuID == "" || seenSkuIDs[order.SkuID] {
			continue
		}
		seenSkuIDs[order.SkuID] = true
		var count int64
		qCheck := tx.Model(&domain.MarketplaceProductMapping{}).Where("sku_id = ?", order.SkuID)
		if order.BranchID != nil && *order.BranchID != uuid.Nil {
			qCheck = qCheck.Where("branch_id = ?", *order.BranchID)
		}
		qCheck.Count(&count)
		if count == 0 {
			unmappedLabels = append(unmappedLabels, order.SellerSKU+" ("+order.SkuID+")")
		}
	}
	if len(unmappedLabels) > 0 {
		tx.Rollback()
		return fmt.Errorf("sku_belum_mapping:%s", strings.Join(unmappedLabels, "|"))
	}
	// ─────────────────────────────────────────────────────────────────────

	// Jika ada pilihan kardus dari pop-up ATAU eksplisit skip_default_box (misal pilih tanpa kardus di modal)
	skipMappingCardboard := req.SkipDefaultBox || (req.BoxStockID != nil && *req.BoxStockID != uuid.Nil && req.BoxQty > 0)

	orderMarketplace := "TIKTOK"
	if len(orders) > 0 && orders[0].Marketplace != "" {
		orderMarketplace = strings.ToUpper(orders[0].Marketplace)
	} else if isShopeeOrderID(orderID) {
		orderMarketplace = "SHOPEE"
	}

	for _, order := range orders {
		if order.OrderStatus == "PACKED" {
			continue
		}

		if err := tx.Model(&order).Updates(map[string]interface{}{
			"order_status": "PACKED",
			"packed_by":    adminUsername,
		}).Error; err != nil {
			tx.Rollback()
			return err
		}

		if order.SkuID == "" {
			continue // Skip mapping deduction if SkuID is empty
		}

		var mappings []domain.MarketplaceProductMapping
		qMap := tx.Where("sku_id = ?", order.SkuID)
		if order.BranchID != nil && *order.BranchID != uuid.Nil {
			qMap = qMap.Where("branch_id = ?", *order.BranchID)
		}
		if err := qMap.Find(&mappings).Error; err != nil {
			tx.Rollback()
			return err
		}

		for _, mapping := range mappings {
			totalDeduction := mapping.Quantity * order.ItemCount

			if mapping.MarketplaceStockID != nil && *mapping.MarketplaceStockID != uuid.Nil {
				var stock domain.MarketplaceStock
				if err := tx.Where("id = ?", *mapping.MarketplaceStockID).First(&stock).Error; err != nil {
					tx.Rollback()
					return err
				}

				if skipMappingCardboard {
					nameUpper := strings.ToUpper(stock.ProductName)
					skuUpper := strings.ToUpper(stock.SKU)
					if strings.Contains(nameUpper, "KARDUS") || strings.Contains(skuUpper, "STOCK - K -") || strings.Contains(skuUpper, "KARDUS") || strings.Contains(skuUpper, "BOX") {
						continue // Skip default cardboard box mapping
					}
				}

				oldQuantity := stock.Quantity
				newQuantity := oldQuantity - totalDeduction
				stock.Quantity = newQuantity

				if err := tx.Save(&stock).Error; err != nil {
					tx.Rollback()
					return err
				}

				effectivePrice := stock.Price
				if effectivePrice == 0 && stock.ProductID != nil && *stock.ProductID != uuid.Nil {
					var p domain.Product
					if err := tx.Where("id = ?", *stock.ProductID).First(&p).Error; err == nil {
						effectivePrice = p.SellingPrice
					}
				}

				log := domain.MarketplaceOpname{
					MarketplaceStockID: stock.ID,
					Type:               domain.OpnameStockOut,
					Quantity:           totalDeduction,
					QuantityBefore:     oldQuantity,
					QuantityAfter:      newQuantity,
					Price:              effectivePrice,
					TotalValue:         float64(totalDeduction) * effectivePrice,
					Reference:          orderID,
					SourceType:         orderMarketplace,
					MarketplaceOrderID: orderID,
					Notes:              "Packing order " + orderID,
					CreatedBy:          adminUsername,
				}

				if err := tx.Create(&log).Error; err != nil {
					tx.Rollback()
					return err
				}
			} else if mapping.MarketplaceInventoryID != nil && *mapping.MarketplaceInventoryID != uuid.Nil {
				var inv domain.MarketplaceInventory
				if err := tx.Where("id = ?", *mapping.MarketplaceInventoryID).First(&inv).Error; err != nil {
					tx.Rollback()
					return err
				}

				if skipMappingCardboard {
					nameUpper := strings.ToUpper(inv.Name)
					skuUpper := strings.ToUpper(inv.SKU)
					if strings.Contains(nameUpper, "KARDUS") || strings.Contains(skuUpper, "STOCK - K -") || strings.Contains(skuUpper, "KARDUS") || strings.Contains(skuUpper, "BOX") {
						continue // Skip default cardboard box mapping
					}
				}

				oldQuantity := inv.Quantity
				newQuantity := oldQuantity - totalDeduction
				inv.Quantity = newQuantity
				if err := tx.Save(&inv).Error; err != nil {
					tx.Rollback()
					return err
				}

				log := domain.MarketplaceInventoryLog{
					MarketplaceInventoryID: inv.ID,
					Type:                   domain.InventoryOut,
					Quantity:               totalDeduction,
					QuantityBefore:         oldQuantity,
					QuantityAfter:          newQuantity,
					Reference:              orderID,
					Notes:                  "Packing order " + orderID,
					CreatedBy:              adminUsername,
				}
				if err := tx.Create(&log).Error; err != nil {
					tx.Rollback()
					return err
				}
			}
		}

		// Insert scan history
		history := domain.MarketplaceScanHistory{
			OrderID:       order.OrderID,
			BranchID:      order.BranchID,
			Action:        "PACK",
			Actor:         adminUsername,
			SellerSKU:     order.SellerSKU,
			ProductName:   order.ProductName,
			WarehouseName: order.WarehouseName,
		}
		if err := tx.Create(&history).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	// ── Potong stok kardus (opsional) ────────────────────────────────────
	if req.BoxStockID != nil && *req.BoxStockID != uuid.Nil && req.BoxQty > 0 {
		var boxStock domain.MarketplaceStock
		if err := tx.Where("id = ?", *req.BoxStockID).First(&boxStock).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("stok kardus tidak ditemukan: %w", err)
		}
		oldQty := boxStock.Quantity
		newQty := oldQty - req.BoxQty
		boxStock.Quantity = newQty
		if err := tx.Save(&boxStock).Error; err != nil {
			tx.Rollback()
			return err
		}
		boxOpname := domain.MarketplaceOpname{
			MarketplaceStockID: boxStock.ID,
			Type:               domain.OpnameStockOut,
			Quantity:           req.BoxQty,
			QuantityBefore:     oldQty,
			QuantityAfter:      newQty,
			Price:              boxStock.Price,
			TotalValue:         float64(req.BoxQty) * boxStock.Price,
			Reference:          orderID,
			SourceType:         orderMarketplace,
			MarketplaceOrderID: orderID,
			Notes:              "Kardus packing order " + orderID,
			CreatedBy:          adminUsername,
		}
		if err := tx.Create(&boxOpname).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	// ─────────────────────────────────────────────────────────────────────

	// ── Potong stok bonus/barang tambahan ────────────────────────────────
	for _, bonus := range req.AdditionalItems {
		if bonus.StockID == "" || bonus.Qty <= 0 {
			continue
		}

		var bonusStock domain.MarketplaceStock
		if err := tx.Where("id = ?", bonus.StockID).First(&bonusStock).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("stok bonus tidak ditemukan: %w", err)
		}

		oldQty := bonusStock.Quantity
		newQty := oldQty - bonus.Qty
		bonusStock.Quantity = newQty
		if err := tx.Save(&bonusStock).Error; err != nil {
			tx.Rollback()
			return err
		}

		bonusOpname := domain.MarketplaceOpname{
			MarketplaceStockID: bonusStock.ID,
			Type:               domain.OpnameStockOut,
			Quantity:           bonus.Qty,
			QuantityBefore:     oldQty,
			QuantityAfter:      newQty,
			Price:              bonusStock.Price,
			TotalValue:         float64(bonus.Qty) * bonusStock.Price,
			Reference:          orderID,
			SourceType:         orderMarketplace,
			MarketplaceOrderID: orderID,
			Notes:              "Barang tambahan/bonus packing order " + orderID,
			CreatedBy:          adminUsername,
		}
		if err := tx.Create(&bonusOpname).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	// ─────────────────────────────────────────────────────────────────────

	return tx.Commit().Error
}

func (r *marketplaceRepository) VerifyOrder(orderID string, adminUsername string) error {
	var orders []domain.MarketplaceOrder
	if err := r.db.Where("order_id = ?", orderID).Find(&orders).Error; err != nil {
		return err
	}

	if len(orders) == 0 {
		return errors.New("order not found")
	}

	for _, order := range orders {
		if order.OrderStatus == "VERIFIED" || order.OrderStatus == "PACKED" {
			continue
		}

		if err := r.db.Model(&order).Updates(map[string]interface{}{
			"order_status": "VERIFIED",
			"picked_by":    adminUsername,
		}).Error; err != nil {
			return err
		}

		// Insert scan history
		history := domain.MarketplaceScanHistory{
			OrderID:       order.OrderID,
			BranchID:      order.BranchID,
			Action:        "PICK",
			Actor:         adminUsername,
			SellerSKU:     order.SellerSKU,
			ProductName:   order.ProductName,
			WarehouseName: order.WarehouseName,
		}
		if err := r.db.Create(&history).Error; err != nil {
			return err
		}
	}

	return nil
}

func (r *marketplaceRepository) GetProductMappings(skuID string, branchID *uuid.UUID) ([]domain.MarketplaceProductMapping, error) {
	var mappings []domain.MarketplaceProductMapping
	q := r.db.Preload("MarketplaceStock").Where("sku_id = ?", skuID)
	if branchID != nil && *branchID != uuid.Nil {
		q = q.Where("branch_id = ?", *branchID)
	}
	if err := q.Find(&mappings).Error; err != nil {
		return nil, err
	}
	return mappings, nil
}

func (r *marketplaceRepository) SaveProductMappings(skuID string, branchID *uuid.UUID, mappings []domain.MarketplaceProductMapping) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Hapus mapping lama untuk SKU dan branch tersebut
		qDel := tx.Where("sku_id = ?", skuID)
		if branchID != nil && *branchID != uuid.Nil {
			qDel = qDel.Where("branch_id = ?", *branchID)
		}
		if err := qDel.Delete(&domain.MarketplaceProductMapping{}).Error; err != nil {
			return err
		}

		if len(mappings) > 0 {
			// Set sku_id and branch_id
			for i := range mappings {
				mappings[i].SkuID = skuID
				if branchID != nil && *branchID != uuid.Nil {
					mappings[i].BranchID = branchID
				}
			}
			if err := tx.Create(&mappings).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *marketplaceRepository) GetScanHistories(filter ScanHistoryFilter) ([]domain.MarketplaceScanHistory, error) {
	var histories []domain.MarketplaceScanHistory

	q := r.db.Table("marketplace_scan_histories as msh").
		Select("msh.id, msh.branch_id, msh.order_id, msh.action, msh.actor, msh.seller_sku, msh.product_name, msh.created_at, COALESCE(NULLIF(msh.warehouse_name, ''), mo.warehouse_name, '') as warehouse_name").
		Joins("LEFT JOIN marketplace_orders as mo ON msh.order_id = mo.order_id AND msh.seller_sku = mo.seller_sku").
		Group("msh.id, msh.branch_id, msh.order_id, msh.action, msh.actor, msh.seller_sku, msh.product_name, msh.created_at, mo.warehouse_name")

	if filter.BranchID != nil && *filter.BranchID != uuid.Nil {
		q = q.Where("msh.branch_id = ?", *filter.BranchID)
	}
	if filter.From != nil {
		q = q.Where("msh.created_at >= ?", *filter.From)
	}
	if filter.To != nil {
		q = q.Where("msh.created_at <= ?", *filter.To)
	}
	if filter.Action != "" && filter.Action != "ALL" {
		q = q.Where("msh.action = ?", filter.Action)
	}
	if filter.WarehouseName != "" {
		q = q.Where("COALESCE(NULLIF(msh.warehouse_name, ''), mo.warehouse_name, '') ILIKE ?", "%"+filter.WarehouseName+"%")
	}

	err := q.Order("msh.created_at desc").Find(&histories).Error
	return histories, err
}

func (r *marketplaceRepository) GetScanHistoryWarehouses(branchID *uuid.UUID) ([]string, error) {
	var warehouses []string
	q := r.db.Table("marketplace_scan_histories as msh").
		Joins("LEFT JOIN marketplace_orders as mo ON msh.order_id = mo.order_id AND msh.seller_sku = mo.seller_sku")

	if branchID != nil && *branchID != uuid.Nil {
		q = q.Where("msh.branch_id = ?", *branchID)
	}

	err := q.Select("DISTINCT COALESCE(NULLIF(msh.warehouse_name, ''), mo.warehouse_name, '')").
		Pluck("COALESCE(NULLIF(msh.warehouse_name, ''), mo.warehouse_name, '')", &warehouses).Error
	if err != nil {
		return nil, err
	}
	// Filter out empty strings
	result := make([]string, 0, len(warehouses))
	for _, w := range warehouses {
		if w != "" {
			result = append(result, w)
		}
	}
	return result, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// MARKETPLACE SETTLEMENT & FINANCE IMPLEMENTATIONS
// ─────────────────────────────────────────────────────────────────────────────

func (r *marketplaceRepository) GetSettlementOrders(params MarketplaceSettlementFilter) ([]domain.MarketplaceOrderGroup, int64, float64, float64, float64, error) {
	var total int64
	q := r.db.Model(&domain.MarketplaceOrder{})

	if params.BranchID != nil && *params.BranchID != uuid.Nil {
		q = q.Where("branch_id = ?", *params.BranchID)
	}

	if params.Marketplace != "" && params.Marketplace != "ALL" {
		q = q.Where("marketplace = ?", strings.ToLower(params.Marketplace))
	}

	// Filter based on Tab
	switch strings.ToUpper(params.Tab) {
	case "SETTLED":
		q = q.Where("settlement_status = 'SETTLED' AND " + verifiedMarketplaceSettlementSQL)
	case "HOLD", "ESCROW":
		q = q.Where("(settlement_status IN ('HOLD', 'PENDING_SETTLEMENT', 'ESCROW', '') OR (settlement_status = 'SETTLED' AND NOT " + verifiedMarketplaceSettlementSQL + ")) AND order_status NOT IN ('CANCELED', 'CANCELLED', 'RETURN_SUCCESS', 'FAILED_DELIVERY')")
	case "REFUND_RETURN", "REFUND", "RETURN":
		q = q.Where("settlement_status IN ('REFUNDED', 'CANCELLED', 'FAILED_DELIVERY') OR order_status IN ('CANCELED', 'CANCELLED', 'RETURN_SUCCESS', 'FAILED_DELIVERY') OR return_status IN ('PENDING_INSPECTION', 'RESTOCKED', 'DAMAGED_LOSS')")
	}

	// Filter based on Packing / Physical Reconciliation Status
	switch strings.ToUpper(params.PackingStatus) {
	case "PACKED":
		q = q.Where("packed_by IS NOT NULL AND packed_by != ''")
	case "UNPACKED":
		q = q.Where("packed_by IS NULL OR packed_by = ''")
	case "NEEDS_RESTOCK", "DISPATCHED_RETURN":
		// Order is cancelled/returned after being packed and not yet restocked
		q = q.Where("(packed_by IS NOT NULL AND packed_by != '') AND (order_status IN ('CANCELED', 'CANCELLED', 'RETURN_SUCCESS', 'FAILED_DELIVERY') OR settlement_status IN ('REFUNDED', 'CANCELLED', 'FAILED_DELIVERY')) AND restocked_at IS NULL")
	case "RESTOCKED":
		q = q.Where("restocked_at IS NOT NULL OR return_status = 'RESTOCKED'")
	}

	if params.SettlementStatus != "" && params.SettlementStatus != "ALL" {
		q = q.Where("settlement_status = ?", params.SettlementStatus)
	}

	if params.ReturnStatus != "" && params.ReturnStatus != "ALL" {
		q = q.Where("return_status = ?", params.ReturnStatus)
	}

	if params.Keyword != "" {
		kw := "%" + params.Keyword + "%"
		q = q.Where(
			"order_id ILIKE ? OR seller_sku ILIKE ? OR tracking_id ILIKE ? OR buyer_username ILIKE ? OR product_name ILIKE ?",
			kw, kw, kw, kw, kw,
		)
	}
	if params.ProductName != "" {
		q = q.Where("product_name ILIKE ?", "%"+params.ProductName+"%")
	}
	if params.WarehouseName != "" {
		q = q.Where("warehouse_name ILIKE ?", "%"+params.WarehouseName+"%")
	}
	if params.ShippingProvider != "" {
		q = q.Where("shipping_provider ILIKE ?", "%"+params.ShippingProvider+"%")
	}

	dateCol := "created_at"
	if params.DateType == "settlement_date" {
		dateCol = "COALESCE(settlement_date, created_at)"
	}

	if params.From != nil {
		q = q.Where(fmt.Sprintf("%s >= ?", dateCol), *params.From)
	}
	if params.To != nil {
		q = q.Where(fmt.Sprintf("%s <= ?", dateCol), *params.To)
	}

	// One grouped pass supplies both pagination count and totals.
	var totalSettledAmount, totalPlatformFee, totalRefundAmount float64
	subQ := q.Session(&gorm.Session{}).
		Select("order_id, MAX(CASE WHEN settlement_status = 'SETTLED' AND " + verifiedMarketplaceSettlementSQL + " THEN " + marketplaceSettlementAmountSQL + " ELSE 0 END) as settled, MAX(COALESCE(NULLIF(fee_and_tax_amount, 0), platform_fee, 0)) as fee, MAX(CASE WHEN refund_amount > 0 THEN refund_amount WHEN order_status IN ('CANCELED', 'CANCELLED', 'RETURN_SUCCESS', 'FAILED_DELIVERY') THEN total_amount ELSE 0 END) as refund").
		Group("order_id")
	row := r.db.Table("(?) as sub", subQ).
		Select("COUNT(*), COALESCE(SUM(settled), 0), COALESCE(SUM(fee), 0), COALESCE(SUM(refund), 0)").Row()
	if err := row.Scan(&total, &totalSettledAmount, &totalPlatformFee, &totalRefundAmount); err != nil {
		return nil, 0, 0, 0, 0, err
	}

	page := params.Page
	if page < 1 {
		page = 1
	}
	pageSize := params.PageSize
	if pageSize < 1 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	// Get distinct order_ids ordered by max(created_at) desc or max(settlement_date) desc
	orderCol := fmt.Sprintf("MAX(%s) desc", dateCol)
	var orderIDs []string
	if err := q.Select("order_id").Group("order_id").Order(orderCol).Limit(pageSize).Offset(offset).Pluck("order_id", &orderIDs).Error; err != nil {
		return nil, 0, 0, 0, 0, err
	}

	if len(orderIDs) == 0 {
		return []domain.MarketplaceOrderGroup{}, total, totalSettledAmount, totalPlatformFee, totalRefundAmount, nil
	}

	// Fetch all line items for these order_ids
	var lineItems []domain.MarketplaceOrder
	if err := r.db.Where("order_id IN ?", orderIDs).Order("created_at desc").Find(&lineItems).Error; err != nil {
		return nil, 0, 0, 0, 0, err
	}

	// Group line items by order_id
	groupsMap := make(map[string]*domain.MarketplaceOrderGroup)
	for _, item := range lineItems {
		if _, exists := groupsMap[item.OrderID]; !exists {
			isPacked := (item.PackedBy != "" || item.OrderStatus == "PACKED" || item.OrderStatus == "VERIFIED")
			packingStatus := "UNPACKED"
			if isPacked {
				packingStatus = "PACKED"
			}

			settleStatus := item.SettlementStatus
			settleAmount := item.SettlementAmount
			settleDate := item.SettlementDate
			refundAmount := item.RefundAmount

			// Compute Reconciliation Status
			feeVal := item.FeeAndTaxAmount
			if feeVal == 0 {
				feeVal = item.PlatformFee
			}

			// TikTok finance figures can arrive before its statement is paid.
			// Never turn those figures into a paid status without a PAID statement.
			if item.Marketplace != "tiktok" && (item.OrderStatus == "COMPLETED" || item.OrderStatus == "DELIVERED") && (settleStatus == "" || settleStatus == "PENDING_SETTLEMENT" || settleStatus == "HOLD" || settleStatus == "ESCROW") && (settleAmount > 0 || feeVal > 0) {
				settleStatus = "SETTLED"
				if settleAmount == 0 {
					settleAmount = item.TotalAmount - feeVal
				}
				if settleDate == nil {
					settleDate = &item.CreatedAt
				}
			} else if (item.OrderStatus == "CANCELED" || item.OrderStatus == "CANCELLED" || item.OrderStatus == "RETURN_SUCCESS" || item.OrderStatus == "FAILED_DELIVERY") && (settleStatus == "" || settleStatus == "PENDING_SETTLEMENT" || settleStatus == "HOLD" || settleStatus == "ESCROW") {
				settleStatus = "REFUNDED"
				if refundAmount == 0 {
					refundAmount = item.TotalAmount
				}
			}

			// Compute Reconciliation Status
			recStatus := "IN_TRANSIT_HOLD"
			isCancelledOrReturn := item.OrderStatus == "CANCELED" || item.OrderStatus == "CANCELLED" ||
				item.OrderStatus == "RETURN_SUCCESS" || item.OrderStatus == "FAILED_DELIVERY" ||
				settleStatus == "REFUNDED" || settleStatus == "CANCELLED" || settleStatus == "FAILED_DELIVERY"

			if item.RestockedAt != nil || item.ReturnStatus == "RESTOCKED" {
				recStatus = "RESTOCKED"
			} else if isCancelledOrReturn {
				if isPacked {
					recStatus = "RETURN_DISPATCHED" // Critical alert: Dispatched from warehouse but cancelled/returned!
				} else {
					recStatus = "RETURN_UNPACKED" // Cancelled before being packed
				}
			} else if settleStatus == "SETTLED" {
				if isPacked {
					recStatus = "MATCH_SETTLED" // Normal: Settled & Packed
				} else {
					recStatus = "ANOMALY_SETTLED_UNPACKED" // Settled without app pack
				}
			} else {
				if isPacked {
					recStatus = "IN_TRANSIT_HOLD"
				} else {
					recStatus = "AWAITING_PACK_HOLD"
				}
			}

			groupsMap[item.OrderID] = &domain.MarketplaceOrderGroup{
				OrderID:                 item.OrderID,
				BranchID:                item.BranchID,
				OrderStatus:             item.OrderStatus,
				WarehouseName:           item.WarehouseName,
				TrackingID:              item.TrackingID,
				Marketplace:             item.Marketplace,
				BuyerUsername:           item.BuyerUsername,
				PaymentMethod:           item.PaymentMethod,
				TotalAmount:             item.TotalAmount,
				Currency:                item.Currency,
				ShippingProvider:        item.ShippingProvider,
				PickedBy:                item.PickedBy,
				PackedBy:                item.PackedBy,
				IsPacked:                isPacked,
				PackingStatus:           packingStatus,
				ReconciliationStatus:    recStatus,
				SettlementStatus:        settleStatus,
				SettlementDate:          settleDate,
				EstimatedSettlementDate: item.EstimatedSettlementDate,
				SettlementAmount:        settleAmount,
				PlatformFee:             item.PlatformFee,
				CommissionFee:           item.CommissionFee,
				ShippingFeeDeduction:    item.ShippingFeeDeduction,
				FeeAndTaxAmount:         item.FeeAndTaxAmount,
				ShippingCostAmount:      item.ShippingCostAmount,
				AdjustmentAmount:        item.AdjustmentAmount,
				FeeTaxBreakdown:         item.FeeTaxBreakdown,
				HoldReason:              item.HoldReason,
				FinanceSyncedAt:         item.FinanceSyncedAt,
				RefundAmount:            refundAmount,
				RefundReason:            item.RefundReason,
				ReturnStatus:            item.ReturnStatus,
				ReturnItemCount:         item.ReturnItemCount,
				RestockedAt:             item.RestockedAt,
				RestockedBy:             item.RestockedBy,
				CreatedAt:               item.CreatedAt,
				Items:                   []domain.MarketplaceOrder{},
			}
		} else {
			if groupsMap[item.OrderID].TrackingID == "" && item.TrackingID != "" {
				groupsMap[item.OrderID].TrackingID = item.TrackingID
			}
			if groupsMap[item.OrderID].ShippingProvider == "" && item.ShippingProvider != "" {
				groupsMap[item.OrderID].ShippingProvider = item.ShippingProvider
			}
			if groupsMap[item.OrderID].WarehouseName == "" && item.WarehouseName != "" {
				groupsMap[item.OrderID].WarehouseName = item.WarehouseName
			}
		}
		groupsMap[item.OrderID].Items = append(groupsMap[item.OrderID].Items, item)
	}

	// Reorder groups based on original orderIDs order
	var groups []domain.MarketplaceOrderGroup
	for _, id := range orderIDs {
		if g, exists := groupsMap[id]; exists {
			groups = append(groups, *g)
		}
	}

	return groups, total, totalSettledAmount, totalPlatformFee, totalRefundAmount, nil
}

func (r *marketplaceRepository) GetSettlementSummary(branchID *uuid.UUID, from, to *time.Time, marketplace string, dateType string) (*domain.SettlementSummaryResponse, error) {
	summary := &domain.SettlementSummaryResponse{}

	baseQ := r.db.Model(&domain.MarketplaceOrder{})
	if branchID != nil && *branchID != uuid.Nil {
		baseQ = baseQ.Where("branch_id = ?", *branchID)
	}
	if marketplace != "" && marketplace != "ALL" {
		baseQ = baseQ.Where("marketplace = ?", strings.ToLower(marketplace))
	}
	dateCol := "created_at"
	if dateType == "settlement_date" {
		dateCol = "COALESCE(settlement_date, created_at)"
	}
	if from != nil {
		baseQ = baseQ.Where(fmt.Sprintf("%s >= ?", dateCol), *from)
	}
	if to != nil {
		baseQ = baseQ.Where(fmt.Sprintf("%s <= ?", dateCol), *to)
	}

	settled := "settlement_status = 'SETTLED' AND " + verifiedMarketplaceSettlementSQL
	pending := "(settlement_status IN ('HOLD', 'PENDING_SETTLEMENT', 'ESCROW', '') OR (settlement_status = 'SETTLED' AND NOT " + verifiedMarketplaceSettlementSQL + ")) AND order_status NOT IN ('CANCELED', 'CANCELLED', 'RETURN_SUCCESS', 'FAILED_DELIVERY')"
	refunded := "settlement_status IN ('REFUNDED', 'CANCELLED', 'FAILED_DELIVERY') OR order_status IN ('CANCELED', 'CANCELLED', 'RETURN_SUCCESS', 'FAILED_DELIVERY')"
	pendingReturn := "return_status = 'PENDING_INSPECTION' OR (order_status IN ('CANCELED', 'RETURN_SUCCESS', 'FAILED_DELIVERY') AND (return_status = 'NO_RETURN' OR return_status = '' OR return_status IS NULL))"
	restocked := "return_status = 'RESTOCKED' OR restocked_at IS NOT NULL"
	dispatchedReturn := "(packed_by IS NOT NULL AND packed_by != '') AND (order_status IN ('CANCELED', 'CANCELLED', 'RETURN_SUCCESS', 'FAILED_DELIVERY') OR settlement_status IN ('REFUNDED', 'CANCELLED', 'FAILED_DELIVERY')) AND restocked_at IS NULL"

	// Aggregate each order once, then calculate every summary card in a single table pass.
	perOrder := baseQ.Session(&gorm.Session{}).Select(strings.Join([]string{
		"order_id",
		"MAX(CASE WHEN " + settled + " THEN 1 ELSE 0 END) AS settled_order",
		"MAX(CASE WHEN " + settled + " THEN " + marketplaceSettlementAmountSQL + " END) AS settled_amount",
		"MAX(CASE WHEN " + settled + " THEN COALESCE(NULLIF(fee_and_tax_amount, 0), platform_fee, 0) END) AS platform_fee_amount",
		"MAX(CASE WHEN " + pending + " THEN 1 ELSE 0 END) AS pending_order",
		"MAX(CASE WHEN " + pending + " THEN total_amount END) AS pending_amount",
		"MAX(CASE WHEN " + refunded + " THEN 1 ELSE 0 END) AS refund_order",
		"MAX(CASE WHEN " + refunded + " THEN CASE WHEN refund_amount > 0 THEN refund_amount ELSE total_amount END END) AS refund_amount_value",
		"SUM(CASE WHEN " + pendingReturn + " THEN COALESCE(item_count, 0) ELSE 0 END) AS pending_return_items",
		"SUM(CASE WHEN " + restocked + " THEN COALESCE(item_count, 0) ELSE 0 END) AS restocked_items",
		"MAX(CASE WHEN " + dispatchedReturn + " THEN 1 ELSE 0 END) AS dispatched_return_order",
		"MAX(CASE WHEN " + dispatchedReturn + " THEN total_amount END) AS dispatched_return_value",
		"MAX(CASE WHEN " + restocked + " THEN 1 ELSE 0 END) AS restocked_return_order",
		"MAX(CASE WHEN " + restocked + " THEN total_amount END) AS restocked_return_value",
	}, ", ")).Group("order_id")

	err := r.db.Table("(?) AS grouped", perOrder).Select(strings.Join([]string{
		"COALESCE(SUM(settled_order), 0) AS total_settled_orders",
		"COALESCE(SUM(settled_amount), 0) AS total_settled_amount",
		"COALESCE(SUM(platform_fee_amount), 0) AS total_platform_fees",
		"COALESCE(SUM(pending_order), 0) AS total_pending_orders",
		"COALESCE(SUM(pending_amount), 0) AS total_pending_amount",
		"COALESCE(SUM(refund_order), 0) AS total_refund_orders",
		"COALESCE(SUM(refund_amount_value), 0) AS total_refund_amount",
		"COALESCE(SUM(pending_return_items), 0) AS total_pending_return_items",
		"COALESCE(SUM(restocked_items), 0) AS total_restocked_items",
		"COALESCE(SUM(dispatched_return_order), 0) AS total_dispatched_return_orders",
		"COALESCE(SUM(dispatched_return_value), 0) AS total_dispatched_return_value",
		"COALESCE(SUM(restocked_return_order), 0) AS total_restocked_return_orders",
		"COALESCE(SUM(restocked_return_value), 0) AS total_restocked_return_value",
	}, ", ")).Scan(summary).Error
	if err != nil {
		return nil, err
	}
	return summary, nil
}

func (r *marketplaceRepository) RestockOrder(orderID string, adminUsername, notes string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var orders []domain.MarketplaceOrder
		if err := tx.Where("order_id = ?", orderID).Find(&orders).Error; err != nil {
			return err
		}
		if len(orders) == 0 {
			return errors.New("pesanan tidak ditemukan")
		}

		for _, item := range orders {
			if item.ReturnStatus == "RESTOCKED" {
				return errors.New("pesanan ini sudah pernah direstock ke gudang")
			}
		}

		now := time.Now()
		totalRestockedQty := 0
		orderMarketplace := "TIKTOK"
		if len(orders) > 0 && orders[0].Marketplace != "" {
			orderMarketplace = strings.ToUpper(orders[0].Marketplace)
		} else if isShopeeOrderID(orderID) {
			orderMarketplace = "SHOPEE"
		}

		for _, item := range orders {
			qty := item.ItemCount
			if qty <= 0 {
				qty = 1
			}

			// 1. Coba cari mapping produk
			var mappings []domain.MarketplaceProductMapping
			qMap := tx.Where("sku_id = ?", item.SkuID)
			if item.BranchID != nil && *item.BranchID != uuid.Nil {
				qMap = qMap.Where("branch_id = ?", *item.BranchID)
			}
			_ = qMap.Find(&mappings).Error

			if len(mappings) > 0 {
				for _, m := range mappings {
					var stock domain.MarketplaceStock
					if err := tx.First(&stock, "id = ?", m.MarketplaceStockID).Error; err != nil {
						continue
					}

					mappingQty := m.Quantity
					if mappingQty <= 0 {
						mappingQty = 1
					}
					itemDeduction := qty * mappingQty
					qtyBefore := stock.Quantity
					stock.Quantity += itemDeduction

					if err := tx.Save(&stock).Error; err != nil {
						return err
					}

					opnameNotes := fmt.Sprintf("Restock pesanan batal/retur %s (SKU: %s)", orderID, item.SellerSKU)
					if notes != "" {
						opnameNotes += " - " + notes
					}

					opname := &domain.MarketplaceOpname{
						MarketplaceStockID: stock.ID,
						Type:               domain.OpnameStockIn,
						Quantity:           itemDeduction,
						QuantityBefore:     qtyBefore,
						QuantityAfter:      stock.Quantity,
						Price:              stock.Price,
						TotalValue:         float64(itemDeduction) * stock.Price,
						Reference:          fmt.Sprintf("RETUR-%s", orderID),
						SourceType:         orderMarketplace,
						MarketplaceOrderID: orderID,
						Notes:              opnameNotes,
						CreatedBy:          adminUsername,
						CreatedAt:          now,
					}
					if err := tx.Create(opname).Error; err != nil {
						return err
					}
					totalRestockedQty += itemDeduction
				}
			} else {
				// 2. Fallback cari direct MarketplaceStock dengan SKU yang sama
				var stock domain.MarketplaceStock
				qStock := tx.Where("sku = ?", item.SellerSKU)
				if item.BranchID != nil && *item.BranchID != uuid.Nil {
					qStock = qStock.Where("branch_id = ?", *item.BranchID)
				}
				if err := qStock.First(&stock).Error; err == nil {
					qtyBefore := stock.Quantity
					stock.Quantity += qty
					if err := tx.Save(&stock).Error; err != nil {
						return err
					}

					opnameNotes := fmt.Sprintf("Restock pesanan batal/retur %s", orderID)
					if notes != "" {
						opnameNotes += " - " + notes
					}

					opname := &domain.MarketplaceOpname{
						MarketplaceStockID: stock.ID,
						Type:               domain.OpnameStockIn,
						Quantity:           qty,
						QuantityBefore:     qtyBefore,
						QuantityAfter:      stock.Quantity,
						Price:              stock.Price,
						TotalValue:         float64(qty) * stock.Price,
						Reference:          fmt.Sprintf("RETUR-%s", orderID),
						SourceType:         orderMarketplace,
						MarketplaceOrderID: orderID,
						Notes:              opnameNotes,
						CreatedBy:          adminUsername,
						CreatedAt:          now,
					}
					if err := tx.Create(opname).Error; err != nil {
						return err
					}
					totalRestockedQty += qty
				}
			}
		}

		// Update all orders items
		updates := map[string]interface{}{
			"return_status":     "RESTOCKED",
			"restocked_at":      now,
			"restocked_by":      adminUsername,
			"return_item_count": totalRestockedQty,
		}
		if err := tx.Model(&domain.MarketplaceOrder{}).Where("order_id = ?", orderID).Updates(updates).Error; err != nil {
			return err
		}

		return nil
	})
}

func (r *marketplaceRepository) BatchImportSettlements(records []domain.SettlementImportRecord, adminUsername string) (int, error) {
	updatedCount := 0
	err := r.db.Transaction(func(tx *gorm.DB) error {
		for _, rec := range records {
			if strings.TrimSpace(rec.OrderID) == "" {
				continue
			}

			status := rec.Status
			if status == "" {
				status = "SETTLED"
			}

			updates := map[string]interface{}{
				"settlement_status": status,
				"settlement_date":   rec.SettlementDate,
				"settlement_amount": rec.SettlementAmount,
				"platform_fee":      rec.PlatformFee,
				"commission_fee":    rec.CommissionFee,
				"finance_synced_at": time.Now(),
			}
			if rec.GrossAmount > 0 {
				updates["total_amount"] = rec.GrossAmount
			}

			res := tx.Model(&domain.MarketplaceOrder{}).Where("order_id = ?", rec.OrderID).Updates(updates)
			if res.Error != nil {
				return res.Error
			}
			if res.RowsAffected > 0 {
				updatedCount++
			}
		}
		return nil
	})
	return updatedCount, err
}

func (r *marketplaceRepository) SaveFinanceTransactions(records []domain.FinanceTransactionRecord) ([]string, error) {
	if len(records) == 0 {
		return nil, nil
	}

	updatedIDs := make(map[string]struct{})
	batchSize := 100
	now := time.Now()

	for i := 0; i < len(records); i += batchSize {
		end := i + batchSize
		if end > len(records) {
			end = len(records)
		}
		chunk := records[i:end]

		err := r.db.Transaction(func(tx *gorm.DB) error {
			for _, rec := range chunk {
				if strings.TrimSpace(rec.OrderID) == "" {
					continue
				}

				// STRICT ISOLATION: Hanya update kolom finansial dan tracking settlement.
				// DILARANG menyertakan order_status, created_at, packed_by, picked_by, atau tracking_id!
				updates := map[string]interface{}{
					"finance_synced_at":    now,
					"shipping_cost_amount": math.Abs(rec.ShippingCostAmount),
					"adjustment_amount":    rec.AdjustmentAmount,
				}
				if rec.SettlementProvided {
					updates["settlement_amount"] = rec.SettlementAmount
				}
				if rec.FeeAndTaxProvided {
					fee := math.Abs(rec.FeeAndTaxAmount)
					updates["fee_and_tax_amount"] = fee
					updates["platform_fee"] = fee
				}
				if rec.FeeTaxBreakdown != "" {
					updates["fee_tax_breakdown"] = rec.FeeTaxBreakdown
				}
				if rec.SettlementStatus != "" {
					updates["settlement_status"] = rec.SettlementStatus
					updates["settlement_date"] = rec.SettlementDate
					updates["hold_reason"] = rec.HoldReason
				}

				marketplace := rec.Marketplace
				if marketplace == "" {
					marketplace = "tiktok"
				}
				q := tx.Model(&domain.MarketplaceOrder{}).Where("marketplace = ? AND order_id = ?", marketplace, rec.OrderID)
				if rec.BranchID != nil {
					q = q.Where("branch_id = ?", *rec.BranchID)
				}
				res := q.Updates(updates)
				if res.Error != nil {
					return res.Error
				}
				if res.RowsAffected > 0 {
					updatedIDs[rec.OrderID] = struct{}{}
				}
			}
			return nil
		})

		if err != nil {
			return nil, err
		}
	}
	ids := make([]string, 0, len(updatedIDs))
	for id := range updatedIDs {
		ids = append(ids, id)
	}
	return ids, nil
}

func (r *marketplaceRepository) UpdateSettlementOrder(orderID string, req domain.UpdateSettlementOrderReq) error {
	updates := make(map[string]interface{})
	if req.SettlementStatus != nil {
		updates["settlement_status"] = *req.SettlementStatus
	}
	if req.SettlementDate != nil {
		updates["settlement_date"] = *req.SettlementDate
	}
	if req.SettlementAmount != nil {
		updates["settlement_amount"] = *req.SettlementAmount
	}
	if req.PlatformFee != nil {
		updates["platform_fee"] = *req.PlatformFee
	}
	if req.CommissionFee != nil {
		updates["commission_fee"] = *req.CommissionFee
	}
	if req.FeeAndTaxAmount != nil {
		updates["fee_and_tax_amount"] = *req.FeeAndTaxAmount
	}
	if req.ShippingCostAmount != nil {
		updates["shipping_cost_amount"] = *req.ShippingCostAmount
	}
	if req.AdjustmentAmount != nil {
		updates["adjustment_amount"] = *req.AdjustmentAmount
	}
	if req.FeeTaxBreakdown != nil {
		updates["fee_tax_breakdown"] = *req.FeeTaxBreakdown
	}
	if req.RefundAmount != nil {
		updates["refund_amount"] = *req.RefundAmount
	}
	if req.RefundReason != nil {
		updates["refund_reason"] = *req.RefundReason
	}
	if req.ReturnStatus != nil {
		updates["return_status"] = *req.ReturnStatus
	}

	if len(updates) == 0 {
		return nil
	}

	return r.db.Model(&domain.MarketplaceOrder{}).Where("order_id = ?", orderID).Updates(updates).Error
}

func (r *marketplaceRepository) AutoClassifySettlements(branchID *uuid.UUID) (int64, error) {
	var totalUpdated int64

	// 1. Auto-classify COMPLETED / DELIVERED orders that have verified settlement figures into SETTLED
	qSettled := r.db.Model(&domain.MarketplaceOrder{}).
		Where("marketplace <> 'tiktok'").
		Where("order_status NOT IN ('CANCELED', 'CANCELLED', 'RETURN_SUCCESS', 'FAILED_DELIVERY')").
		Where("order_status IN ('COMPLETED', 'DELIVERED')").
		Where("settlement_status IS NULL OR settlement_status = '' OR settlement_status IN ('PENDING_SETTLEMENT', 'HOLD', 'ESCROW')").
		Where("settlement_amount > 0 OR fee_and_tax_amount > 0 OR platform_fee > 0")
	if branchID != nil && *branchID != uuid.Nil {
		qSettled = qSettled.Where("branch_id = ?", *branchID)
	}
	resSettled := qSettled.Updates(map[string]interface{}{
		"settlement_status": "SETTLED",
		"settlement_amount": gorm.Expr("CASE WHEN settlement_amount = 0 OR settlement_amount IS NULL THEN (total_amount - COALESCE(NULLIF(fee_and_tax_amount, 0), platform_fee, 0)) ELSE settlement_amount END"),
		"settlement_date":   gorm.Expr("COALESCE(settlement_date, created_at)"),
	})
	if resSettled.Error != nil {
		return 0, resSettled.Error
	}
	totalUpdated += resSettled.RowsAffected

	// 2. Auto-classify CANCELED / RETURN_SUCCESS / FAILED_DELIVERY orders into REFUNDED
	qRefund := r.db.Model(&domain.MarketplaceOrder{}).
		Where("order_status IN ('CANCELED', 'CANCELLED', 'RETURN_SUCCESS', 'FAILED_DELIVERY')").
		Where("settlement_status IS NULL OR settlement_status = '' OR settlement_status IN ('PENDING_SETTLEMENT', 'HOLD', 'ESCROW')")
	if branchID != nil && *branchID != uuid.Nil {
		qRefund = qRefund.Where("branch_id = ?", *branchID)
	}
	resRefund := qRefund.Updates(map[string]interface{}{
		"settlement_status": "REFUNDED",
		"refund_amount":     gorm.Expr("CASE WHEN refund_amount = 0 OR refund_amount IS NULL THEN total_amount ELSE refund_amount END"),
	})
	if resRefund.Error != nil {
		return totalUpdated, resRefund.Error
	}
	totalUpdated += resRefund.RowsAffected

	// 3. Ensure any other remaining empty settlement_status defaults to PENDING_SETTLEMENT
	qPending := r.db.Model(&domain.MarketplaceOrder{}).
		Where("settlement_status IS NULL OR settlement_status = ''")
	if branchID != nil && *branchID != uuid.Nil {
		qPending = qPending.Where("branch_id = ?", *branchID)
	}
	resPending := qPending.Update("settlement_status", "PENDING_SETTLEMENT")
	if resPending.Error != nil {
		return totalUpdated, resPending.Error
	}
	totalUpdated += resPending.RowsAffected

	return totalUpdated, nil
}

func (r *marketplaceRepository) CleanNoiseAndRestoreScannedOrders(branchID *uuid.UUID) error {
	// 1. Restore all orders that have scan history in marketplace_scan_histories or packed_by to PACKED
	qRestore := r.db.Exec(`
		UPDATE marketplace_orders 
		SET order_status = 'PACKED', 
		    packed_by = COALESCE(NULLIF(packed_by, ''), (
		        SELECT actor FROM marketplace_scan_histories 
		        WHERE marketplace_scan_histories.order_id = marketplace_orders.order_id 
		          AND marketplace_scan_histories.action = 'PACK' 
		        ORDER BY created_at DESC LIMIT 1
		    ), 'Admin')
		WHERE (order_status IS DISTINCT FROM 'PACKED' OR packed_by IS NULL OR packed_by = '')
		  AND (order_id IN (
		    SELECT order_id FROM marketplace_scan_histories WHERE action = 'PACK'
		  ) OR (packed_by IS NOT NULL AND packed_by != ''))
	`)
	if qRestore.Error != nil {
		return qRestore.Error
	}

	// 2. Delete noise cancelled orders that have never been packed/scanned
	qDeleteCancelled := r.db.Exec(`
		DELETE FROM marketplace_orders 
		WHERE order_status IN ('CANCELLED', 'CANCELED') 
		  AND (packed_by IS NULL OR packed_by = '') 
		  AND order_id NOT IN (SELECT order_id FROM marketplace_scan_histories)
	`)
	return qDeleteCancelled.Error
}
