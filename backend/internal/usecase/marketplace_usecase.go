package usecase

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"example.com/hris/backend/internal/repository"
)

type UpdateOpnamePayload struct {
	Date      *string `json:"date"`
	Quantity  *int    `json:"quantity"`
	Reference *string `json:"reference"`
	Notes     *string `json:"notes"`
}

type MarketplaceUsecase interface {
	CreateStock(stock *domain.MarketplaceStock, actorID uuid.UUID, actorName, ip string) error
	GetStockByID(id uuid.UUID) (*domain.MarketplaceStock, error)
	GetAllStocks(branchID uuid.UUID) ([]domain.MarketplaceStock, error)
	UpdateStock(id uuid.UUID, updates *domain.MarketplaceStock, actorID uuid.UUID, actorName, ip string) error
	DeleteStock(id uuid.UUID, actorID uuid.UUID, actorName, ip string) error
	DoOpname(stockID uuid.UUID, opnameType domain.MarketplaceOpnameType, qty int, reference, notes string, opnameDate *time.Time, actorID uuid.UUID, actorName, ip string) (*domain.MarketplaceOpname, error)
	UpdateOpname(id uuid.UUID, req UpdateOpnamePayload, actorID uuid.UUID, actorName, ip string) (*domain.MarketplaceOpname, error)
	GetOpnames(branchID uuid.UUID, stockID *uuid.UUID, opnameType string, from, to *time.Time, search string, limit, offset int, sourceType string) ([]domain.MarketplaceOpname, int64, repository.MarketplaceOpnameSummary, error)
	GetDashboardStats(branchID uuid.UUID, period string, dateFilter string) (map[string]interface{}, error)
	GetFactoryProducts(companyID uuid.UUID) ([]domain.Product, error)
	GetAuditLogs(entity string) ([]domain.AuditLog, error)
	// Marketplace order & product sync
	GetMarketplaceOrders(params repository.MarketplaceOrderFilter) ([]domain.MarketplaceOrderGroup, int64, float64, int64, int64, error)
	GetMarketplaceOrderStatusStats(params repository.MarketplaceOrderFilter) (repository.MarketplaceOrderStatusStats, error)
	GetOrderFilters(branchID *uuid.UUID) (*domain.OrderFiltersResponse, error)
	SaveMarketplaceOrder(order *domain.MarketplaceOrder) error
	GetMarketplaceProducts(branchID *uuid.UUID) ([]domain.MarketplaceProduct, error)
	SaveMarketplaceProduct(product *domain.MarketplaceProduct) error
	PackOrder(orderID string, adminUsername string, req domain.PackOrderReq) error
	VerifyOrder(orderID string, adminUsername string) error
	GetProductMappings(skuID string, branchID *uuid.UUID) ([]domain.MarketplaceProductMapping, error)
	SaveProductMappings(skuID string, branchID *uuid.UUID, mappings []domain.MarketplaceProductMapping) error
	GetScanHistories(filter repository.ScanHistoryFilter) ([]domain.MarketplaceScanHistory, error)
	GetScanHistoryWarehouses(branchID *uuid.UUID) ([]string, error)

	// Marketplace Settlement & Finance methods
	GetSettlementOrders(params repository.MarketplaceSettlementFilter) ([]domain.MarketplaceOrderGroup, int64, float64, float64, float64, error)
	GetSettlementSummary(branchID *uuid.UUID, from, to *time.Time, marketplace string, dateType string) (*domain.SettlementSummaryResponse, error)
	RestockOrder(orderID string, adminUsername, notes string) error
	BatchImportSettlements(records []domain.SettlementImportRecord, adminUsername string) (int, error)
	SaveFinanceTransactions(records []domain.FinanceTransactionRecord) ([]string, error)
	UpdateSettlementOrder(orderID string, req domain.UpdateSettlementOrderReq) error
	AutoClassifySettlements(branchID *uuid.UUID) (int64, error)
	CleanNoiseAndRestoreScannedOrders(branchID *uuid.UUID) error
}

type marketplaceUsecase struct {
	repo         repository.MarketplaceRepository
	auditUsecase AuditLogUsecase
	companyID    uuid.UUID
}

func NewMarketplaceUsecase(repo repository.MarketplaceRepository, audit AuditLogUsecase, companyID string) MarketplaceUsecase {
	cid, _ := uuid.Parse(companyID)
	return &marketplaceUsecase{repo: repo, auditUsecase: audit, companyID: cid}
}

func (u *marketplaceUsecase) CreateStock(stock *domain.MarketplaceStock, actorID uuid.UUID, actorName, ip string) error {
	// Auto-snapshot nama dan SKU dari produk pabrik jika belum diisi dan ProductID diberikan
	if stock.ProductID != nil && *stock.ProductID != uuid.Nil {
		products, err := u.repo.GetFactoryProducts(u.companyID)
		if err == nil {
			for _, p := range products {
				if p.ID == *stock.ProductID {
					if strings.TrimSpace(stock.ProductName) == "" {
						stock.ProductName = p.Name
					}
					if strings.TrimSpace(stock.SKU) == "" {
						stock.SKU = p.SKU
					}
					// Set harga pabrik sebagai harga default jika belum diisi
					if stock.Price == 0 {
						stock.Price = p.SellingPrice
					}
					break
				}
			}
		}
	}
	stock.ProductName = strings.TrimSpace(stock.ProductName)
	stock.SKU = strings.TrimSpace(stock.SKU)
	if err := u.repo.CreateStock(stock); err != nil {
		return err
	}
	u.auditUsecase.Log(u.companyID, actorID, actorName, "CREATE", "MarketplaceStock", stock.ID.String(),
		fmt.Sprintf("Tambah produk %s (SKU: %s)", stock.ProductName, stock.SKU), nil, stock, ip)
	return nil
}

func (u *marketplaceUsecase) GetStockByID(id uuid.UUID) (*domain.MarketplaceStock, error) {
	return u.repo.GetStockByID(id)
}

func (u *marketplaceUsecase) GetAllStocks(branchID uuid.UUID) ([]domain.MarketplaceStock, error) {
	return u.repo.GetAllStocks(branchID)
}

func (u *marketplaceUsecase) UpdateStock(id uuid.UUID, updates *domain.MarketplaceStock, actorID uuid.UUID, actorName, ip string) error {
	existing, err := u.repo.GetStockByID(id)
	if err != nil {
		return err
	}
	old := *existing
	existing.ProductID = updates.ProductID
	if strings.TrimSpace(updates.ProductName) != "" {
		existing.ProductName = strings.TrimSpace(updates.ProductName)
	}
	if strings.TrimSpace(updates.SKU) != "" {
		existing.SKU = strings.TrimSpace(updates.SKU)
	}
	existing.MinStock = updates.MinStock
	existing.Price = updates.Price
	existing.Notes = updates.Notes
	if updates.ImageURL != "" {
		existing.ImageURL = updates.ImageURL
	}
	if err := u.repo.UpdateStock(existing); err != nil {
		return err
	}
	u.auditUsecase.Log(u.companyID, actorID, actorName, "UPDATE", "MarketplaceStock", id.String(),
		fmt.Sprintf("Update produk %s (SKU: %s)", existing.ProductName, existing.SKU), old, existing, ip)
	return nil
}

func (u *marketplaceUsecase) DeleteStock(id uuid.UUID, actorID uuid.UUID, actorName, ip string) error {
	existing, err := u.repo.GetStockByID(id)
	if err != nil {
		return err
	}
	if err := u.repo.DeleteStock(id); err != nil {
		return err
	}
	u.auditUsecase.Log(u.companyID, actorID, actorName, "DELETE", "MarketplaceStock", id.String(),
		fmt.Sprintf("Hapus produk %s", existing.ProductName), existing, nil, ip)
	return nil
}

func (u *marketplaceUsecase) DoOpname(stockID uuid.UUID, opnameType domain.MarketplaceOpnameType, qty int, reference, notes string, opnameDate *time.Time, actorID uuid.UUID, actorName, ip string) (*domain.MarketplaceOpname, error) {
	stock, err := u.repo.GetStockByID(stockID)
	if err != nil {
		return nil, errors.New("produk tidak ditemukan")
	}

	qtyBefore := stock.Quantity

	switch opnameType {
	case domain.OpnameStockIn:
		stock.Quantity += qty
	case domain.OpnameStockOut:
		if stock.Quantity < qty {
			return nil, errors.New("stok tidak mencukupi")
		}
		stock.Quantity -= qty
	case domain.OpnameAdjustment:
		stock.Quantity = qty
	}

	qtyAfter := stock.Quantity

	if err := u.repo.UpdateStock(stock); err != nil {
		return nil, err
	}

	createdAt := time.Now()
	if opnameDate != nil && !opnameDate.IsZero() {
		createdAt = *opnameDate
	}

	effectivePrice := stock.Price
	if effectivePrice == 0 && stock.Product != nil {
		effectivePrice = stock.Product.SellingPrice
	}

	opname := &domain.MarketplaceOpname{
		MarketplaceStockID: stockID,
		Type:               opnameType,
		Quantity:           qty,
		QuantityBefore:     qtyBefore,
		QuantityAfter:      qtyAfter,
		Price:              effectivePrice,
		TotalValue:         float64(qty) * effectivePrice,
		Reference:          reference,
		SourceType:         "MANUAL",
		Notes:              notes,
		CreatedBy:          actorName,
		CreatedAt:          createdAt,
	}
	if err := u.repo.CreateOpname(opname); err != nil {
		return nil, err
	}
	u.auditUsecase.Log(u.companyID, actorID, actorName, string(opnameType), "MarketplaceStock", stockID.String(),
		fmt.Sprintf("Opname %s: %d → %d (qty: %d) [Waktu: %s]", opnameType, qtyBefore, qtyAfter, qty, createdAt.Format("2006-01-02 15:04:05")), nil, opname, ip)
	return opname, nil
}

func (u *marketplaceUsecase) UpdateOpname(id uuid.UUID, req UpdateOpnamePayload, actorID uuid.UUID, actorName, ip string) (*domain.MarketplaceOpname, error) {
	opname, err := u.repo.GetOpnameByID(id)
	if err != nil {
		return nil, errors.New("riwayat pergerakan stok tidak ditemukan")
	}

	stock, err := u.repo.GetStockByID(opname.MarketplaceStockID)
	if err != nil {
		return nil, errors.New("stok produk tidak ditemukan")
	}

	oldOpname := *opname

	// 1. Update Tanggal jika dikirim
	if req.Date != nil && strings.TrimSpace(*req.Date) != "" {
		dateStr := strings.TrimSpace(*req.Date)
		formats := []string{
			time.RFC3339,
			"2006-01-02T15:04:05Z07:00",
			"2006-01-02T15:04:05",
			"2006-01-02T15:04",
			"2006-01-02 15:04:05",
			"2006-01-02 15:04",
			"2006-01-02",
		}
		for _, f := range formats {
			if parsed, err := time.ParseInLocation(f, dateStr, time.Local); err == nil {
				opname.CreatedAt = parsed
				break
			}
		}
	}

	// 2. Update No. Referensi
	if req.Reference != nil {
		opname.Reference = strings.TrimSpace(*req.Reference)
	}

	// 3. Update Catatan
	if req.Notes != nil {
		opname.Notes = strings.TrimSpace(*req.Notes)
	}

	// 4. Update Quantity & sesuaikan saldo stok
	if req.Quantity != nil && *req.Quantity >= 0 && *req.Quantity != opname.Quantity {
		newQty := *req.Quantity
		delta := newQty - opname.Quantity

		switch opname.Type {
		case domain.OpnameStockIn:
			if stock.Quantity+delta < 0 {
				return nil, errors.New("perubahan jumlah menyebabkan saldo stok fisik menjadi minus")
			}
			stock.Quantity += delta
		case domain.OpnameStockOut:
			if stock.Quantity-delta < 0 {
				return nil, errors.New("stok tidak mencukupi untuk penambahan jumlah keluar")
			}
			stock.Quantity -= delta
		case domain.OpnameAdjustment:
			if stock.Quantity+delta < 0 {
				return nil, errors.New("perubahan penyesuaian menyebabkan saldo stok fisik menjadi minus")
			}
			stock.Quantity += delta
		}

		opname.Quantity = newQty
		effectivePrice := opname.Price
		if effectivePrice == 0 {
			effectivePrice = stock.Price
		}
		if effectivePrice == 0 && stock.Product != nil {
			effectivePrice = stock.Product.SellingPrice
		}
		opname.Price = effectivePrice
		opname.TotalValue = float64(newQty) * effectivePrice
		opname.QuantityAfter = opname.QuantityBefore + delta

		if err := u.repo.UpdateStock(stock); err != nil {
			return nil, err
		}
	}

	if err := u.repo.UpdateOpname(opname); err != nil {
		return nil, err
	}

	u.auditUsecase.Log(u.companyID, actorID, actorName, "UPDATE_OPNAME", "MarketplaceStock", opname.MarketplaceStockID.String(),
		fmt.Sprintf("Edit riwayat mutasi %s: ID %s [Waktu: %s, Qty: %d, Ref: %s]",
			opname.Type, opname.ID.String(), opname.CreatedAt.Format("2006-01-02 15:04"), opname.Quantity, opname.Reference),
		oldOpname, opname, ip)

	return opname, nil
}

func (u *marketplaceUsecase) GetOpnames(branchID uuid.UUID, stockID *uuid.UUID, opnameType string, from, to *time.Time, search string, limit, offset int, sourceType string) ([]domain.MarketplaceOpname, int64, repository.MarketplaceOpnameSummary, error) {
	return u.repo.GetOpnames(branchID, stockID, opnameType, from, to, search, limit, offset, sourceType)
}

func (u *marketplaceUsecase) GetDashboardStats(branchID uuid.UUID, period string, dateFilter string) (map[string]interface{}, error) {
	totalSKU, criticalCount, todayIn, todayOut, todayOmzet, chartData, err := u.repo.GetStockStats(branchID, period, dateFilter)
	if err != nil {
		return nil, err
	}

	result := map[string]interface{}{
		"total_sku":      totalSKU,
		"critical_count": criticalCount,
		"today_in":       todayIn,
		"today_out":      todayOut,
		"today_omzet":    todayOmzet,
	}

	if chartData != nil {
		result["chart_data"] = chartData
	}

	return result, nil
}

func (u *marketplaceUsecase) GetFactoryProducts(companyID uuid.UUID) ([]domain.Product, error) {
	if companyID == uuid.Nil {
		companyID = u.companyID
	}
	return u.repo.GetFactoryProducts(companyID)
}

func (u *marketplaceUsecase) GetAuditLogs(entity string) ([]domain.AuditLog, error) {
	logs, _, err := u.auditUsecase.GetAuditLogs(u.companyID, entity, "", "", 1, 1000)
	return logs, err
}

// GetMarketplaceOrders delegates to the repository with filter and pagination.
func (u *marketplaceUsecase) GetMarketplaceOrders(params repository.MarketplaceOrderFilter) ([]domain.MarketplaceOrderGroup, int64, float64, int64, int64, error) {
	return u.repo.GetMarketplaceOrders(params)
}

func (u *marketplaceUsecase) GetMarketplaceOrderStatusStats(params repository.MarketplaceOrderFilter) (repository.MarketplaceOrderStatusStats, error) {
	return u.repo.GetMarketplaceOrderStatusStats(params)
}

func (u *marketplaceUsecase) GetOrderFilters(branchID *uuid.UUID) (*domain.OrderFiltersResponse, error) {
	return u.repo.GetOrderFilters(branchID)
}

// SaveMarketplaceOrder upserts a marketplace order via the repository.
func (u *marketplaceUsecase) SaveMarketplaceOrder(order *domain.MarketplaceOrder) error {
	return u.repo.UpsertMarketplaceOrder(order)
}

// GetMarketplaceProducts returns all marketplace products from the repository.
func (u *marketplaceUsecase) GetMarketplaceProducts(branchID *uuid.UUID) ([]domain.MarketplaceProduct, error) {
	return u.repo.GetMarketplaceProducts(branchID)
}

// SaveMarketplaceProduct upserts a marketplace product via the repository.
func (u *marketplaceUsecase) SaveMarketplaceProduct(product *domain.MarketplaceProduct) error {
	return u.repo.UpsertMarketplaceProduct(product)
}

func (u *marketplaceUsecase) PackOrder(orderID string, adminUsername string, req domain.PackOrderReq) error {
	return u.repo.PackOrder(orderID, adminUsername, req)
}

func (u *marketplaceUsecase) VerifyOrder(orderID string, adminUsername string) error {
	return u.repo.VerifyOrder(orderID, adminUsername)
}

func (u *marketplaceUsecase) GetProductMappings(skuID string, branchID *uuid.UUID) ([]domain.MarketplaceProductMapping, error) {
	return u.repo.GetProductMappings(skuID, branchID)
}

func (u *marketplaceUsecase) SaveProductMappings(skuID string, branchID *uuid.UUID, mappings []domain.MarketplaceProductMapping) error {
	return u.repo.SaveProductMappings(skuID, branchID, mappings)
}

func (u *marketplaceUsecase) GetScanHistories(filter repository.ScanHistoryFilter) ([]domain.MarketplaceScanHistory, error) {
	return u.repo.GetScanHistories(filter)
}

func (u *marketplaceUsecase) GetScanHistoryWarehouses(branchID *uuid.UUID) ([]string, error) {
	return u.repo.GetScanHistoryWarehouses(branchID)
}

func (u *marketplaceUsecase) GetSettlementOrders(params repository.MarketplaceSettlementFilter) ([]domain.MarketplaceOrderGroup, int64, float64, float64, float64, error) {
	return u.repo.GetSettlementOrders(params)
}

func (u *marketplaceUsecase) GetSettlementSummary(branchID *uuid.UUID, from, to *time.Time, marketplace string, dateType string) (*domain.SettlementSummaryResponse, error) {
	return u.repo.GetSettlementSummary(branchID, from, to, marketplace, dateType)
}

func (u *marketplaceUsecase) RestockOrder(orderID string, adminUsername, notes string) error {
	return u.repo.RestockOrder(orderID, adminUsername, notes)
}

func (u *marketplaceUsecase) BatchImportSettlements(records []domain.SettlementImportRecord, adminUsername string) (int, error) {
	return u.repo.BatchImportSettlements(records, adminUsername)
}

func (u *marketplaceUsecase) SaveFinanceTransactions(records []domain.FinanceTransactionRecord) ([]string, error) {
	return u.repo.SaveFinanceTransactions(records)
}

func (u *marketplaceUsecase) UpdateSettlementOrder(orderID string, req domain.UpdateSettlementOrderReq) error {
	return u.repo.UpdateSettlementOrder(orderID, req)
}

func (u *marketplaceUsecase) AutoClassifySettlements(branchID *uuid.UUID) (int64, error) {
	return u.repo.AutoClassifySettlements(branchID)
}

func (u *marketplaceUsecase) CleanNoiseAndRestoreScannedOrders(branchID *uuid.UUID) error {
	return u.repo.CleanNoiseAndRestoreScannedOrders(branchID)
}
