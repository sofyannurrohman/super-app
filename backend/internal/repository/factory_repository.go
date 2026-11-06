package repository

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"gorm.io/gorm"
)

type FactoryRepository interface {
	// Factory
	CreateFactory(factory *domain.Factory) error
	GetFactoryByID(id uuid.UUID) (*domain.Factory, error)
	GetFactoriesByCompanyID(companyID uuid.UUID) ([]domain.Factory, error)
	UpdateFactory(factory *domain.Factory) error
	DeleteFactory(id uuid.UUID) error

	// Factory Warehouse
	CreateFactoryWarehouse(warehouse *domain.FactoryWarehouse) error
	GetFactoryWarehouseByID(id uuid.UUID) (*domain.FactoryWarehouse, error)
	GetFactoryWarehousesByFactoryID(factoryID uuid.UUID) ([]domain.FactoryWarehouse, error)
	UpdateFactoryWarehouse(warehouse *domain.FactoryWarehouse) error
	DeleteFactoryWarehouse(id uuid.UUID) error

	// Product
	CreateProduct(product *domain.Product) error
	GetProductsByCompanyID(companyID uuid.UUID, branchID uuid.UUID) ([]domain.Product, error)
	GetProductByID(id uuid.UUID) (*domain.Product, error)
	UpdateProduct(product *domain.Product) error
	DeleteProduct(id uuid.UUID) error

	// Stock
	GetStock(factoryID, productID uuid.UUID, batchNo string) (*domain.FactoryStock, error)
	UpdateStock(stock *domain.FactoryStock) error
	DeleteStock(id uuid.UUID) error
	GetStocksByFactoryID(factoryID uuid.UUID) ([]domain.FactoryStock, error)
	GetAllStocksByCompanyID(companyID uuid.UUID) ([]domain.FactoryStock, error)
	GetAvailableBatches(factoryID, productID uuid.UUID) ([]domain.FactoryStock, error)
	GetRawMaterialStockByFactoryID(factoryID uuid.UUID) ([]domain.FactoryStock, error)
	GetRawMaterialLogsByFactoryID(factoryID uuid.UUID, productID *uuid.UUID) ([]domain.FactoryInventoryLog, error)

	// Logs
	CreateProductionLog(log *domain.ProductionLog) error
	GetProductionLogByID(id uuid.UUID) (*domain.ProductionLog, error)
	GetProductionLogsByFactoryID(factoryID uuid.UUID) ([]domain.ProductionLog, error)
	GetAllProductionLogsByCompanyID(companyID uuid.UUID) ([]domain.ProductionLog, error)
	UpdateProductionLog(log *domain.ProductionLog) error
	DeleteProductionLog(id uuid.UUID) error
	CreateInventoryLog(log *domain.FactoryInventoryLog) error
	DeleteInventoryLog(id uuid.UUID) error
	GetInventoryLogsByFactoryID(factoryID uuid.UUID) ([]domain.FactoryInventoryLog, error)
	GetAllInventoryLogsByCompanyID(companyID uuid.UUID) ([]domain.FactoryInventoryLog, error)

	// Transfer
	CreateTransfer(transfer *domain.ProductTransfer) error
	GetTransfersByFactoryID(factoryID uuid.UUID) ([]domain.ProductTransfer, error)
	GetAllTransfersByCompanyID(companyID uuid.UUID, productID string, batchNo string) ([]domain.ProductTransfer, error)
	GetAllTransfersPaginated(page, limit int, companyID uuid.UUID, branchID, status, search string) ([]domain.ProductTransfer, int64, float64, map[string]int64, error)
	GetTransferByID(id uuid.UUID) (*domain.ProductTransfer, error)
	UpdateTransfer(transfer *domain.ProductTransfer) error
	UpdateTransferStatus(id uuid.UUID, status string) error
	UpdateTransferStatusWithApprover(id uuid.UUID, status string, approverID *uuid.UUID, approverName, approverPosition string) error
	DeleteTransfer(id uuid.UUID) error
	GetTransfersByDO(doNo string) ([]domain.ProductTransfer, error)
	DeleteTransfersByDOAndExcludeIDs(doNo string, excludeIDs []uuid.UUID) error

	// Receivables
	GetBranchReceivables(companyID uuid.UUID) ([]domain.BranchReceivableSummary, error)

	// Recipe
	CreateRecipe(recipe *domain.ProductionRecipe) error
	GetRecipeByFinishedProduct(productID uuid.UUID) (*domain.ProductionRecipe, error)
	GetRecipesByCompany(companyID uuid.UUID) ([]domain.ProductionRecipe, error)
	DeleteRecipe(id uuid.UUID) error
	GetDashboardStats(companyID uuid.UUID) (*domain.FactoryDashboardStats, error)
	GetLowStockAlerts(companyID uuid.UUID) ([]domain.LowStockAlert, error)

	// QC Sample
	CreateQcSample(sample *domain.FactoryQcSample) error
	GetQcSamplesByFactoryID(factoryID uuid.UUID, status string) ([]domain.FactoryQcSample, error)
	GetAllQcSamplesByCompanyID(companyID uuid.UUID, status string) ([]domain.FactoryQcSample, error)
	GetQcSampleByID(id uuid.UUID) (*domain.FactoryQcSample, error)
	UpdateQcSample(sample *domain.FactoryQcSample) error
}

type factoryRepository struct {
	db *gorm.DB
}

func NewFactoryRepository(db *gorm.DB) FactoryRepository {
	return &factoryRepository{db}
}

func (r *factoryRepository) CreateFactory(factory *domain.Factory) error {
	return r.db.Create(factory).Error
}

func (r *factoryRepository) GetFactoryByID(id uuid.UUID) (*domain.Factory, error) {
	var factory domain.Factory
	err := r.db.Preload("Branch").First(&factory, "id = ?", id).Error
	return &factory, err
}

func (r *factoryRepository) GetFactoriesByCompanyID(companyID uuid.UUID) ([]domain.Factory, error) {
	var factories []domain.Factory
	query := r.db.Preload("Branch")
	if companyID != uuid.Nil {
		query = query.Where("company_id = ?", companyID)
	}
	err := query.Find(&factories).Error
	return factories, err
}

func (r *factoryRepository) UpdateFactory(factory *domain.Factory) error {
	return r.db.Model(factory).Select("Name", "Location", "BranchID").Updates(factory).Error
}

func (r *factoryRepository) DeleteFactory(id uuid.UUID) error {
	return r.db.Delete(&domain.Factory{}, "id = ?", id).Error
}

func (r *factoryRepository) CreateFactoryWarehouse(warehouse *domain.FactoryWarehouse) error {
	return r.db.Create(warehouse).Error
}

func (r *factoryRepository) GetFactoryWarehouseByID(id uuid.UUID) (*domain.FactoryWarehouse, error) {
	var warehouse domain.FactoryWarehouse
	err := r.db.First(&warehouse, "id = ?", id).Error
	return &warehouse, err
}

func (r *factoryRepository) GetFactoryWarehousesByFactoryID(factoryID uuid.UUID) ([]domain.FactoryWarehouse, error) {
	var warehouses []domain.FactoryWarehouse
	err := r.db.Find(&warehouses, "factory_id = ?", factoryID).Error
	return warehouses, err
}

func (r *factoryRepository) UpdateFactoryWarehouse(warehouse *domain.FactoryWarehouse) error {
	return r.db.Model(warehouse).Select("Name", "Code", "Type", "Description").Updates(warehouse).Error
}

func (r *factoryRepository) DeleteFactoryWarehouse(id uuid.UUID) error {
	return r.db.Delete(&domain.FactoryWarehouse{}, "id = ?", id).Error
}

func (r *factoryRepository) CreateProduct(product *domain.Product) error {
	return r.db.Create(product).Error
}

func (r *factoryRepository) GetProductsByCompanyID(companyID uuid.UUID, branchID uuid.UUID) ([]domain.Product, error) {
	var products []domain.Product
	err := r.db.Find(&products, "company_id = ?", companyID).Error
	if err != nil {
		return nil, err
	}

	// Fallback: If 0 products found, this might be a distributor/driver company requesting the factory catalog.
	// Since distributors sell factory products, we fetch all products if the initial query is empty.
	if len(products) == 0 {
		err = r.db.Find(&products).Error
		if err != nil {
			return nil, err
		}
	}

	if len(products) == 0 {
		return products, nil
	}

	productIDs := make([]uuid.UUID, len(products))
	for i := range products {
		productIDs[i] = products[i].ID
	}
	var stockTotals []struct {
		ProductID uuid.UUID `gorm:"column:product_id"`
		Quantity  int       `gorm:"column:quantity"`
	}
	stockQuery := r.db.Model(&domain.WarehouseStock{}).
		Select("product_id, COALESCE(SUM(quantity), 0) AS quantity").
		Where("product_id IN ? AND stock_type = ?", productIDs, "GOOD")
	if branchID != uuid.Nil {
		stockQuery = stockQuery.Where("branch_id = ?", branchID)
	}
	if err := stockQuery.Group("product_id").Scan(&stockTotals).Error; err != nil {
		return nil, err
	}
	stockByProduct := make(map[uuid.UUID]int, len(stockTotals))
	for _, stock := range stockTotals {
		stockByProduct[stock.ProductID] = stock.Quantity
	}
	for i := range products {
		products[i].WarehouseStock = stockByProduct[products[i].ID]
	}

	return products, nil
}

func (r *factoryRepository) GetProductByID(id uuid.UUID) (*domain.Product, error) {
	var product domain.Product
	err := r.db.First(&product, "id = ?", id).Error
	return &product, err
}

func (r *factoryRepository) UpdateProduct(product *domain.Product) error {
	return r.db.Model(product).Select("Name", "SKU", "Unit", "PcsPerUnit", "Category", "Brand", "Weight", "WeightUnit", "Length", "Width", "Height", "Volume", "CostPrice", "FactoryPrice", "MarketplaceFactoryPrice", "SellingPrice", "DBP", "WBP", "RBP", "Description", "Specs", "ImageURL", "ShelfLifeDays", "BPOMNumber", "BPOMExpiryDate", "HalalNumber", "HalalExpiryDate").Updates(product).Error
}

func (r *factoryRepository) DeleteProduct(id uuid.UUID) error {
	return r.db.Delete(&domain.Product{}, "id = ?", id).Error
}

func (r *factoryRepository) GetStock(factoryID, productID uuid.UUID, batchNo string) (*domain.FactoryStock, error) {
	var stock domain.FactoryStock
	err := r.db.Where("factory_id = ? AND product_id = ? AND batch_no = ?", factoryID, productID, batchNo).First(&stock).Error
	if err == gorm.ErrRecordNotFound {
		return &domain.FactoryStock{FactoryID: factoryID, ProductID: productID, BatchNo: batchNo, Quantity: 0}, nil
	}
	return &stock, err
}

func (r *factoryRepository) UpdateStock(stock *domain.FactoryStock) error {
	if stock.ID == uuid.Nil {
		return r.db.Create(stock).Error
	}
	return r.db.Omit("Product", "Factory", "FactoryWarehouse").Save(stock).Error
}

func (r *factoryRepository) DeleteStock(id uuid.UUID) error {
	return r.db.Delete(&domain.FactoryStock{}, "id = ?", id).Error
}

func (r *factoryRepository) GetStocksByFactoryID(factoryID uuid.UUID) ([]domain.FactoryStock, error) {
	var stocks []domain.FactoryStock
	err := r.db.Preload("Product").Where("factory_id = ?", factoryID).Find(&stocks).Error
	return stocks, err
}

func (r *factoryRepository) GetAllStocksByCompanyID(companyID uuid.UUID) ([]domain.FactoryStock, error) {
	var stocks []domain.FactoryStock
	query := r.db.Select("factory_stocks.*").Preload("Product").Preload("Factory").
		Joins("JOIN factories ON factories.id = factory_stocks.factory_id")
	if companyID != uuid.Nil {
		query = query.Where("factories.company_id = ?", companyID)
	}
	err := query.Find(&stocks).Error
	return stocks, err
}

func (r *factoryRepository) GetAvailableBatches(factoryID, productID uuid.UUID) ([]domain.FactoryStock, error) {
	var stocks []domain.FactoryStock
	err := r.db.Where("factory_id = ? AND product_id = ? AND quantity > 0 AND batch_no != 'BACKORDER'", factoryID, productID).
		Order("expiry_date ASC").
		Find(&stocks).Error
	return stocks, err
}

func (r *factoryRepository) CreateProductionLog(log *domain.ProductionLog) error {
	return r.db.Create(log).Error
}

func (r *factoryRepository) GetProductionLogByID(id uuid.UUID) (*domain.ProductionLog, error) {
	var log domain.ProductionLog
	err := r.db.Preload("Product").Preload("Employee").Preload("Factory").Preload("Materials.Product").Preload("Operators.Employee").Preload("Operators").First(&log, "id = ?", id).Error
	return &log, err
}

func (r *factoryRepository) GetProductionLogsByFactoryID(factoryID uuid.UUID) ([]domain.ProductionLog, error) {
	var logs []domain.ProductionLog
	err := r.db.Preload("Product").Preload("Employee").Preload("Materials.Product").Preload("Operators.Employee").Preload("Operators").Where("factory_id = ?", factoryID).Order("production_date desc").Find(&logs).Error
	return logs, err
}

func (r *factoryRepository) GetAllProductionLogsByCompanyID(companyID uuid.UUID) ([]domain.ProductionLog, error) {
	var logs []domain.ProductionLog
	query := r.db.Select("production_logs.*").Preload("Product").Preload("Employee").Preload("Factory").Preload("Materials.Product").Preload("Operators.Employee").Preload("Operators").
		Joins("JOIN factories ON factories.id = production_logs.factory_id")
	if companyID != uuid.Nil {
		query = query.Where("factories.company_id = ?", companyID)
	}
	err := query.Order("production_date desc").Find(&logs).Error
	return logs, err
}

func (r *factoryRepository) UpdateProductionLog(log *domain.ProductionLog) error {
	return r.db.Model(&domain.ProductionLog{}).
		Where("id = ?", log.ID).
		Select("product_id", "target_warehouse_id", "batch_no", "expiry_date", "quantity", "carton_count", "pieces_per_carton", "notes", "employee_id", "operator_name").
		Updates(map[string]interface{}{
			"product_id":          log.ProductID,
			"target_warehouse_id": log.TargetWarehouseID,
			"batch_no":           log.BatchNo,
			"expiry_date":        log.ExpiryDate,
			"quantity":           log.Quantity,
			"carton_count":       log.CartonCount,
			"pieces_per_carton":   log.PiecesPerCarton,
			"notes":               log.Notes,
			"employee_id":         log.EmployeeID,
			"operator_name":       log.OperatorName,
		}).Error
}

func (r *factoryRepository) DeleteProductionLog(id uuid.UUID) error {
	return r.db.Delete(&domain.ProductionLog{}, "id = ?", id).Error
}

func (r *factoryRepository) CreateInventoryLog(log *domain.FactoryInventoryLog) error {
	return r.db.Create(log).Error
}

func (r *factoryRepository) DeleteInventoryLog(id uuid.UUID) error {
	return r.db.Delete(&domain.FactoryInventoryLog{}, "id = ?", id).Error
}

func (r *factoryRepository) GetInventoryLogsByFactoryID(factoryID uuid.UUID) ([]domain.FactoryInventoryLog, error) {
	var logs []domain.FactoryInventoryLog
	err := r.db.Preload("Product").Where("factory_id = ?", factoryID).Order("created_at desc").Find(&logs).Error
	return logs, err
}

func (r *factoryRepository) GetAllInventoryLogsByCompanyID(companyID uuid.UUID) ([]domain.FactoryInventoryLog, error) {
	var logs []domain.FactoryInventoryLog
	query := r.db.Select("factory_inventory_logs.*").
		Joins("JOIN factories ON factories.id = factory_inventory_logs.factory_id").
		Preload("Product").
		Preload("Factory")
	if companyID != uuid.Nil {
		query = query.Where("factories.company_id = ?", companyID)
	}
	err := query.Order("factory_inventory_logs.created_at desc").Find(&logs).Error
	return logs, err
}

func (r *factoryRepository) CreateTransfer(transfer *domain.ProductTransfer) error {
	return r.db.Create(transfer).Error
}

func (r *factoryRepository) GetTransfersByFactoryID(factoryID uuid.UUID) ([]domain.ProductTransfer, error) {
	var transfers []domain.ProductTransfer
	err := r.db.Preload("Product").
		Preload("ToBranch").
		Preload("Vehicle").
		Preload("Driver").
		Preload("ApprovedBy").
		Preload("ApprovedBy.JobPosition").
		Where("from_factory_id = ?", factoryID).Order("created_at desc").Find(&transfers).Error
	return transfers, err
}

func (r *factoryRepository) GetAllTransfersByCompanyID(companyID uuid.UUID, productID string, batchNo string) ([]domain.ProductTransfer, error) {
	var transfers []domain.ProductTransfer
	query := r.db.Select("product_transfers.*").
		Preload("Product").
		Preload("ToBranch").
		Preload("FromFactory").
		Preload("Vehicle").
		Preload("Driver").
		Preload("ApprovedBy").
		Preload("ApprovedBy.JobPosition").
		Joins("JOIN factories ON factories.id = product_transfers.from_factory_id")
	if companyID != uuid.Nil {
		query = query.Where("factories.company_id = ?", companyID)
	}
	if productID != "" {
		query = query.Where("product_transfers.product_id = ?", productID)
	}
	if batchNo != "" {
		query = query.Where("product_transfers.batch_no = ?", batchNo)
	}
	err := query.Order("product_transfers.created_at desc").Find(&transfers).Error
	return transfers, err
}

func (r *factoryRepository) GetAllTransfersPaginated(page, limit int, companyID uuid.UUID, branchID, status, search string) ([]domain.ProductTransfer, int64, float64, map[string]int64, error) {
	var transfers []domain.ProductTransfer
	var totalDOs int64
	var totalAmount float64
	stats := map[string]int64{
		"total_invoices": 0,
		"pending_count":  0,
		"shipped_count":  0,
		"received_count": 0,
	}

	// Base query
	query := r.db.Model(&domain.ProductTransfer{}).
		Joins("JOIN factories ON factories.id = product_transfers.from_factory_id").
		Joins("LEFT JOIN branches ON branches.id = product_transfers.to_branch_id").
		Joins("LEFT JOIN products ON products.id = product_transfers.product_id")

	if companyID != uuid.Nil {
		query = query.Where("factories.company_id = ?", companyID)
	}
	if branchID != "" && branchID != "ALL" && branchID != "ALL_BRANCHES" {
		if bUUID, err := uuid.Parse(branchID); err == nil {
			query = query.Where("product_transfers.to_branch_id = ?", bUUID)
		}
	}
	if status != "" && status != "ALL" {
		query = query.Where("product_transfers.status = ?", status)
	}
	if search != "" {
		searchPattern := "%" + strings.ToLower(search) + "%"
		query = query.Where(
			"(LOWER(product_transfers.delivery_order_no) LIKE ? OR LOWER(branches.name) LIKE ? OR LOWER(factories.name) LIKE ? OR LOWER(products.name) LIKE ?)",
			searchPattern, searchPattern, searchPattern, searchPattern,
		)
	}

	// 1. Overall stats and total count by distinct DO key
	type StatResult struct {
		Status string
		Count  int64
	}
	var statResults []StatResult
	statsQuery := query.Session(&gorm.Session{}).
		Select("product_transfers.status, COUNT(DISTINCT COALESCE(NULLIF(product_transfers.delivery_order_no, ''), product_transfers.id::text)) AS count").
		Group("product_transfers.status")
	if err := statsQuery.Scan(&statResults).Error; err != nil {
		return nil, 0, 0, stats, err
	}

	for _, sr := range statResults {
		stats["total_invoices"] += sr.Count
		switch sr.Status {
		case "REQUESTED", "APPROVED":
			stats["pending_count"] += sr.Count
		case "SHIPPED", "ARRIVED":
			stats["shipped_count"] += sr.Count
		case "RECEIVED":
			stats["received_count"] += sr.Count
		}
	}
	totalDOs = stats["total_invoices"]

	// Total amount
	amountQuery := query.Session(&gorm.Session{}).
		Select("COALESCE(SUM(product_transfers.total_amount), 0)")
	if err := amountQuery.Scan(&totalAmount).Error; err != nil {
		return nil, 0, 0, stats, err
	}

	// 2. Paginated DO keys
	if limit > 0 {
		if page < 1 {
			page = 1
		}
		offset := (page - 1) * limit
		var rows []struct {
			DOKey string
		}
		pageDOQuery := query.Session(&gorm.Session{}).
			Select("COALESCE(NULLIF(product_transfers.delivery_order_no, ''), product_transfers.id::text) AS do_key, MAX(product_transfers.created_at) AS max_created_at").
			Group("COALESCE(NULLIF(product_transfers.delivery_order_no, ''), product_transfers.id::text)").
			Order("max_created_at DESC").
			Limit(limit).Offset(offset)
		if err := pageDOQuery.Scan(&rows).Error; err != nil {
			return nil, 0, 0, stats, err
		}

		if len(rows) == 0 {
			return []domain.ProductTransfer{}, totalDOs, totalAmount, stats, nil
		}

		var doKeys []string
		for _, r := range rows {
			doKeys = append(doKeys, r.DOKey)
		}

		err := r.db.Preload("Product").
			Preload("ToBranch").
			Preload("FromFactory").
			Preload("Vehicle").
			Preload("Driver").
			Preload("ApprovedBy").
			Preload("ApprovedBy.JobPosition").
			Where("COALESCE(NULLIF(delivery_order_no, ''), id::text) IN ?", doKeys).
			Order("created_at desc").
			Find(&transfers).Error

		return transfers, totalDOs, totalAmount, stats, err
	}

	// If unpaginated (limit <= 0)
	err := query.Preload("Product").
		Preload("ToBranch").
		Preload("FromFactory").
		Preload("Vehicle").
		Preload("Driver").
		Preload("ApprovedBy").
		Preload("ApprovedBy.JobPosition").
		Order("product_transfers.created_at desc").
		Find(&transfers).Error

	return transfers, totalDOs, totalAmount, stats, err
}

func (r *factoryRepository) GetTransferByID(id uuid.UUID) (*domain.ProductTransfer, error) {
	var transfer domain.ProductTransfer
	err := r.db.Preload("Product").
		Preload("ToBranch").
		Preload("FromFactory").
		Preload("Vehicle").
		Preload("Driver").
		Preload("ApprovedBy").
		Preload("ApprovedBy.JobPosition").
		First(&transfer, "id = ?", id).Error
	return &transfer, err
}

func (r *factoryRepository) UpdateTransfer(transfer *domain.ProductTransfer) error {
	return r.db.Omit("Product", "ToBranch", "FromFactory", "FulfillmentBranch", "Vehicle", "Driver", "ApprovedBy", "Corrections").Save(transfer).Error
}

func (r *factoryRepository) UpdateTransferStatus(id uuid.UUID, status string) error {
	return r.db.Model(&domain.ProductTransfer{}).Where("id = ?", id).Update("status", status).Error
}

func (r *factoryRepository) UpdateTransferStatusWithApprover(id uuid.UUID, status string, approverID *uuid.UUID, approverName, approverPosition string) error {
	updates := map[string]interface{}{
		"status": status,
	}
	if status == "APPROVED" {
		now := time.Now()
		updates["approved_at"] = &now
		if approverID != nil {
			updates["approved_by_id"] = approverID
		}
		if approverName != "" {
			updates["approved_by_name"] = approverName
		}
		if approverPosition != "" {
			updates["approved_by_position"] = approverPosition
		}
	}
	return r.db.Model(&domain.ProductTransfer{}).Where("id = ?", id).Updates(updates).Error
}

func (r *factoryRepository) DeleteTransfer(id uuid.UUID) error {
	return r.db.Delete(&domain.ProductTransfer{}, "id = ?", id).Error
}

func (r *factoryRepository) GetTransfersByDO(doNo string) ([]domain.ProductTransfer, error) {
	var transfers []domain.ProductTransfer
	err := r.db.Preload("Product").
		Preload("ToBranch").
		Preload("FromFactory").
		Preload("Vehicle").
		Preload("Driver").
		Preload("ApprovedBy").
		Preload("ApprovedBy.JobPosition").
		Where("delivery_order_no = ?", doNo).Find(&transfers).Error
	return transfers, err
}

func (r *factoryRepository) DeleteTransfersByDOAndExcludeIDs(doNo string, excludeIDs []uuid.UUID) error {
	query := r.db.Where("delivery_order_no = ?", doNo)
	if len(excludeIDs) > 0 {
		query = query.Where("id NOT IN ?", excludeIDs)
	}
	return query.Delete(&domain.ProductTransfer{}).Error
}

func (r *factoryRepository) CreateRecipe(recipe *domain.ProductionRecipe) error {
	return r.db.Create(recipe).Error
}

func (r *factoryRepository) GetRecipeByFinishedProduct(productID uuid.UUID) (*domain.ProductionRecipe, error) {
	var recipe domain.ProductionRecipe
	err := r.db.Preload("Items.RawProduct").Where("finished_product_id = ?", productID).First(&recipe).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &recipe, err
}

func (r *factoryRepository) GetRecipesByCompany(companyID uuid.UUID) ([]domain.ProductionRecipe, error) {
	var recipes []domain.ProductionRecipe
	query := r.db.Preload("FinishedProduct").
		Joins("JOIN products ON products.id = production_recipes.finished_product_id")
	if companyID != uuid.Nil {
		query = query.Where("products.company_id = ?", companyID)
	}
	err := query.Find(&recipes).Error
	return recipes, err
}

func (r *factoryRepository) DeleteRecipe(id uuid.UUID) error {
	return r.db.Delete(&domain.ProductionRecipe{}, "id = ?", id).Error
}

func (r *factoryRepository) GetDashboardStats(companyID uuid.UUID) (*domain.FactoryDashboardStats, error) {
	var stats domain.FactoryDashboardStats
	var totalFactories int64
	r.db.Model(&domain.Factory{}).Where("company_id = ?", companyID).Count(&totalFactories)
	stats.TotalFactories = int(totalFactories)

	var totalProducts int64
	r.db.Model(&domain.Product{}).Where("company_id = ?", companyID).Count(&totalProducts)
	stats.TotalProducts = int(totalProducts)

	var pendingTransfers int64
	r.db.Model(&domain.ProductTransfer{}).
		Joins("JOIN factories ON factories.id = product_transfers.from_factory_id").
		Where("factories.company_id = ? AND product_transfers.status IN (?)", companyID, []string{"REQUESTED", "APPROVED"}).
		Count(&pendingTransfers)
	stats.PendingShipments = int(pendingTransfers)

	today := time.Now().Format("2006-01-02")
	var todayProd int64
	r.db.Model(&domain.ProductionLog{}).
		Joins("JOIN factories ON factories.id = production_logs.factory_id").
		Where("factories.company_id = ? AND DATE(production_logs.created_at) = ?", companyID, today).
		Select("COALESCE(SUM(quantity), 0)").
		Scan(&todayProd)
	stats.TodayProduction = int(todayProd)

	var recentTransfers []domain.ProductTransfer
	r.db.Preload("Product").Preload("ToBranch").Preload("FromFactory").
		Joins("JOIN factories ON factories.id = product_transfers.from_factory_id").
		Where("factories.company_id = ?", companyID).
		Order("product_transfers.created_at desc").
		Limit(5).
		Find(&recentTransfers)
	stats.RecentTransfers = recentTransfers

	return &stats, nil
}

func (r *factoryRepository) GetLowStockAlerts(companyID uuid.UUID) ([]domain.LowStockAlert, error) {
	var alerts []domain.LowStockAlert
	err := r.db.Raw(`
		SELECT 
			fs.id as stock_id,
			f.name as factory_name,
			p.id as product_id,
			p.name as product_name,
			fs.quantity as current_stock,
			fs.min_limit as min_limit,
			(fs.min_limit - fs.quantity) as deficit
		FROM factory_stocks fs
		JOIN factories f ON f.id = fs.factory_id
		JOIN products p ON p.id = fs.product_id
		WHERE f.company_id = ? AND fs.quantity <= fs.min_limit
		ORDER BY deficit DESC
	`, companyID).Scan(&alerts).Error
	return alerts, err
}

func (r *factoryRepository) GetBranchReceivables(companyID uuid.UUID) ([]domain.BranchReceivableSummary, error) {
	var results []domain.BranchReceivableSummary

	query := `
		SELECT 
			b.id as branch_id,
			b.name as branch_name,
			COALESCE((
				SELECT SUM(pt.total_amount) 
				FROM product_transfers pt 
				WHERE pt.to_branch_id = b.id 
					AND pt.status IN ('SHIPPED', 'ARRIVED', 'RECEIVED') 
					AND COALESCE(pt.is_billable, true) = TRUE
					AND pt.payment_method != 'INTERNAL'
					AND COALESCE(pt.transfer_type, 'COMMERCIAL') != 'INTERNAL_TRANSFER'
			), 0) as total_good_invoiced,
			COALESCE((
				SELECT SUM(pt2.damaged_quantity * COALESCE(p2.factory_price, 0))
				FROM product_transfers pt2
				LEFT JOIN products p2 ON p2.id = pt2.product_id
				WHERE pt2.to_branch_id = b.id
					AND pt2.status = 'RECEIVED'
					AND COALESCE(pt2.is_billable, true) = TRUE
					AND pt2.payment_method != 'INTERNAL'
					AND COALESCE(pt2.transfer_type, 'COMMERCIAL') != 'INTERNAL_TRANSFER'
			), 0) as total_damaged_deduction,
			COALESCE((
				SELECT SUM(COALESCE(fri.received_quantity, fri.quantity) * COALESCE(p.factory_price, 0))
				FROM factory_returns fr
				JOIN factory_return_items fri ON fri.factory_return_id = fr.id
				LEFT JOIN products p ON p.id = fri.product_id
				WHERE fr.from_branch_id = b.id AND (fr.is_received = true OR fr.status = 'RECEIVED')
			), 0) as total_return_deduction,
			(
				COALESCE((
					SELECT SUM(bpl.amount_paid) 
					FROM branch_payment_logs bpl 
					WHERE bpl.branch_id = b.id 
				), 0) + COALESCE((
					SELECT SUM(bd.amount)
					FROM branch_deposits bd
					WHERE bd.branch_id = b.id AND bd.status = 'APPROVED'
				), 0)
			) as total_paid,
			COALESCE((
				SELECT SUM(kt.amount)
				FROM kas_transactions kt
				WHERE kt.target_branch_id = b.id
					AND kt.transaction_type = 'OUT'
					AND kt.is_branch_receivable = TRUE
			), 0) - (
				COALESCE((
					SELECT SUM(kt2.amount)
					FROM kas_transactions kt2
					WHERE kt2.transaction_type = 'IN'
						AND (
							(kt2.target_branch_id = b.id AND (kt2.income_source IN ('BRANCH_RECEIVABLE_PAYMENT', 'BRANCH_SETTLEMENT', 'BRANCH_DEPOSIT', 'ADVANCE_REFUND', 'TALANGAN_CABANG') OR kt2.description ILIKE '%talangan%' OR kt2.description ILIKE '%pelunasan%'))
							OR kt2.related_transaction_id IN (SELECT id FROM kas_transactions WHERE target_branch_id = b.id AND is_branch_receivable = TRUE)
						)
				), 0) + COALESCE((
					SELECT SUM(kas.actual_expense_amount)
					FROM kas_advance_settlements kas
					JOIN kas_transactions kt_parent ON kas.kas_transaction_id = kt_parent.id
					WHERE kt_parent.target_branch_id = b.id AND kt_parent.is_branch_receivable = TRUE
				), 0)
			) as financial_receivable,
			COALESCE((
				SELECT SUM(bpl.amount_paid) 
				FROM branch_payment_logs bpl 
				WHERE bpl.branch_id = b.id 
					AND EXTRACT(MONTH FROM bpl.payment_date) = EXTRACT(MONTH FROM CURRENT_DATE)
					AND EXTRACT(YEAR FROM bpl.payment_date) = EXTRACT(YEAR FROM CURRENT_DATE)
			), 0) + COALESCE((
				SELECT SUM(bd.amount)
				FROM branch_deposits bd
				WHERE bd.branch_id = b.id AND bd.status = 'APPROVED'
					AND EXTRACT(MONTH FROM bd.deposit_date) = EXTRACT(MONTH FROM CURRENT_DATE)
					AND EXTRACT(YEAR FROM bd.deposit_date) = EXTRACT(YEAR FROM CURRENT_DATE)
			), 0) as good_payment_this_month,
			COALESCE((
				SELECT SUM(kt3.amount)
				FROM kas_transactions kt3
				WHERE kt3.transaction_type = 'IN'
					AND (
						(kt3.target_branch_id = b.id AND (kt3.income_source IN ('BRANCH_RECEIVABLE_PAYMENT', 'BRANCH_SETTLEMENT', 'BRANCH_DEPOSIT', 'ADVANCE_REFUND', 'TALANGAN_CABANG') OR kt3.description ILIKE '%talangan%' OR kt3.description ILIKE '%pelunasan%'))
						OR kt3.related_transaction_id IN (SELECT id FROM kas_transactions WHERE target_branch_id = b.id AND is_branch_receivable = TRUE)
					)
					AND EXTRACT(MONTH FROM kt3.transaction_date) = EXTRACT(MONTH FROM CURRENT_DATE)
					AND EXTRACT(YEAR FROM kt3.transaction_date) = EXTRACT(YEAR FROM CURRENT_DATE)
			), 0) as financial_payment_this_month
		FROM branches b
		WHERE b.company_id = ?
		ORDER BY b.name ASC
	`

	err := r.db.Raw(query, companyID).Scan(&results).Error
	if err != nil {
		return nil, err
	}

	for i := range results {
		goodRec := results[i].TotalGoodInvoiced - results[i].TotalDamagedDeduction - results[i].TotalReturnDeduction - results[i].TotalPaid
		if goodRec < 0 {
			goodRec = 0
		}
		results[i].GoodReceivable = goodRec
		results[i].TotalReceivable = results[i].GoodReceivable + results[i].FinancialReceivable
		results[i].TotalPaymentThisMonth = results[i].GoodPaymentThisMonth + results[i].FinancialPaymentThisMonth
	}

	return results, nil
}

func (r *factoryRepository) GetRawMaterialStockByFactoryID(factoryID uuid.UUID) ([]domain.FactoryStock, error) {
	var stocks []domain.FactoryStock
	err := r.db.Preload("Product").Preload("FactoryWarehouse").
		Joins("JOIN products ON products.id = factory_stocks.product_id").
		Where(`factory_stocks.factory_id = ? AND (
			products.category ILIKE '%Bahan%' OR 
			products.category ILIKE '%RAW%' OR 
			products.category ILIKE '%BUMBU%' OR 
			products.category ILIKE '%KEMASAN%' OR 
			products.category ILIKE '%WADAH%' OR 
			products.category ILIKE '%PACKAGING%' OR 
			products.category ILIKE '%PRODUK_PELENGKAP%' OR 
			products.category ILIKE '%TRADING%' OR 
			products.category ILIKE '%HADIAH%' OR 
			products.category ILIKE '%BOTOL%' OR 
			products.category ILIKE '%KARTON%' OR 
			products.category ILIKE '%DUS%' OR 
			products.category ILIKE '%LABEL%' OR 
			products.category ILIKE '%TUTUP%'
		)`, factoryID).
		Find(&stocks).Error
	return stocks, err
}

func (r *factoryRepository) GetRawMaterialLogsByFactoryID(factoryID uuid.UUID, productID *uuid.UUID) ([]domain.FactoryInventoryLog, error) {
	var logs []domain.FactoryInventoryLog
	query := r.db.Preload("Product").
		Joins("JOIN products ON products.id = factory_inventory_logs.product_id").
		Where(`factory_inventory_logs.factory_id = ? AND (
			products.category ILIKE '%Bahan%' OR 
			products.category ILIKE '%RAW%' OR 
			products.category ILIKE '%BUMBU%' OR 
			products.category ILIKE '%KEMASAN%' OR 
			products.category ILIKE '%WADAH%' OR 
			products.category ILIKE '%PACKAGING%' OR 
			products.category ILIKE '%PRODUK_PELENGKAP%' OR 
			products.category ILIKE '%TRADING%' OR 
			products.category ILIKE '%HADIAH%' OR 
			products.category ILIKE '%BOTOL%' OR 
			products.category ILIKE '%KARTON%' OR 
			products.category ILIKE '%DUS%' OR 
			products.category ILIKE '%LABEL%' OR 
			products.category ILIKE '%TUTUP%'
		)`, factoryID)

	if productID != nil && *productID != uuid.Nil {
		query = query.Where("factory_inventory_logs.product_id = ?", *productID)
	}

	err := query.Order("factory_inventory_logs.created_at desc").
		Find(&logs).Error
	return logs, err
}

func (r *factoryRepository) CreateQcSample(sample *domain.FactoryQcSample) error {
	return r.db.Create(sample).Error
}

func (r *factoryRepository) GetQcSamplesByFactoryID(factoryID uuid.UUID, status string) ([]domain.FactoryQcSample, error) {
	var samples []domain.FactoryQcSample
	query := r.db.Preload("Product").Preload("Factory").Preload("ProductionLog").
		Where("factory_id = ?", factoryID)
	if status != "" {
		query = query.Where("status = ?", status)
	}
	err := query.Order("created_at desc").Find(&samples).Error
	return samples, err
}

func (r *factoryRepository) GetAllQcSamplesByCompanyID(companyID uuid.UUID, status string) ([]domain.FactoryQcSample, error) {
	var samples []domain.FactoryQcSample
	query := r.db.Preload("Product").Preload("Factory").Preload("ProductionLog").
		Joins("JOIN factories ON factories.id = factory_qc_samples.factory_id")
	if companyID != uuid.Nil {
		query = query.Where("factories.company_id = ?", companyID)
	}
	if status != "" {
		query = query.Where("factory_qc_samples.status = ?", status)
	}
	err := query.Order("factory_qc_samples.created_at desc").Find(&samples).Error
	return samples, err
}

func (r *factoryRepository) GetQcSampleByID(id uuid.UUID) (*domain.FactoryQcSample, error) {
	var sample domain.FactoryQcSample
	err := r.db.Preload("Product").Preload("Factory").Preload("ProductionLog").
		Where("id = ?", id).First(&sample).Error
	if err != nil {
		return nil, err
	}
	return &sample, nil
}

func (r *factoryRepository) UpdateQcSample(sample *domain.FactoryQcSample) error {
	return r.db.Save(sample).Error
}

