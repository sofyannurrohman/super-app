package http

import (
	"encoding/json"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
)

type profitabilityExpense struct {
	ID                uint      `json:"id"`
	BranchID          uuid.UUID `json:"branch_id"`
	Marketplace       string    `json:"marketplace"`
	ExpenseDate       time.Time `json:"expense_date"`
	Category          string    `json:"category"`
	Amount            float64   `json:"amount"`
	AttributedRevenue float64   `json:"attributed_revenue"`
	SellerSKU         string    `json:"seller_sku"`
	Notes             string    `json:"notes"`
	CreatedAt         time.Time `json:"created_at"`
}

func (profitabilityExpense) TableName() string { return "marketplace_profitability_expenses" }

type profitabilityRow struct {
	BranchID          string   `json:"branch_id"`
	Store             string   `json:"store"`
	Marketplace       string   `json:"marketplace"`
	SellerSKU         string   `json:"seller_sku,omitempty"`
	ProductName       string   `json:"product_name,omitempty"`
	Units             int      `json:"units"`
	Orders            int      `json:"orders"`
	GrossRevenue      float64  `json:"gross_revenue"`
	Refunds           float64  `json:"refunds"`
	NetSales          float64  `json:"net_sales"`
	MarketplaceFees   float64  `json:"marketplace_fees"`
	COGS              float64  `json:"cogs"`
	MissingCOGSUnits  int      `json:"missing_cogs_units"`
	Ads               float64  `json:"ads"`
	Fulfillment       float64  `json:"fulfillment"`
	Returns           float64  `json:"returns"`
	Operational       float64  `json:"operational"`
	AttributedRevenue float64  `json:"attributed_revenue"`
	Contribution      float64  `json:"contribution"`
	MarginPercent     float64  `json:"margin_percent"`
	ROAS              *float64 `json:"roas"`
}

func (r *profitabilityRow) finish() {
	r.NetSales = r.GrossRevenue - r.Refunds
	r.Contribution = r.NetSales - r.MarketplaceFees - r.COGS - r.Ads - r.Fulfillment - r.Returns - r.Operational
	if r.NetSales != 0 {
		r.MarginPercent = 100 * r.Contribution / r.NetSales
	}
	if r.Ads > 0 && r.AttributedRevenue > 0 {
		value := r.AttributedRevenue / r.Ads
		r.ROAS = &value
	}
}

func profitKey(branchID, marketplace, sku string) string {
	return branchID + "|" + marketplace + "|" + sku
}

func (h *MarketplaceHandler) GetProfitability(c *gin.Context) {
	from, errFrom := time.Parse("2006-01-02", c.Query("from"))
	to, errTo := time.Parse("2006-01-02", c.Query("to"))
	if errFrom != nil || errTo != nil || to.Before(from) || to.Sub(from) > 366*24*time.Hour {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Periode tidak valid atau lebih dari 366 hari"})
		return
	}
	branch := parseBranchID(c)
	marketplace := strings.ToLower(c.Query("marketplace"))
	if marketplace != "" && marketplace != "tiktok" && marketplace != "shopee" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Marketplace tidak valid"})
		return
	}
	query := h.db.Model(&domain.MarketplaceOrder{}).Where("created_at >= ? AND created_at < ?", from, to.AddDate(0, 0, 1))
	if branch != uuid.Nil {
		query = query.Where("branch_id = ?", branch)
	}
	if marketplace != "" {
		query = query.Where("marketplace = ?", marketplace)
	}
	var lines []domain.MarketplaceOrder
	if err := query.Order("order_id, id").Find(&lines).Error; err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	// Each line stores the order total and finance fields. Count them once, then
	// allocate them by quantity. This is an estimate until item prices are synced.
	orders := map[string][]domain.MarketplaceOrder{}
	for _, line := range lines {
		key := profitKey(branchString(line.BranchID), line.Marketplace, line.OrderID)
		orders[key] = append(orders[key], line)
	}
	var costRows []struct {
		BranchID    string
		Marketplace string
		SkuID       string
		Cost        float64
	}
	err := h.db.Raw(`SELECT m.branch_id::text AS branch_id, '' AS marketplace,
		m.sku_id, SUM(COALESCE(p.cost_price, 0) * GREATEST(m.quantity, 1)) AS cost
		FROM marketplace_product_mappings m
		JOIN marketplace_stocks s ON s.id = m.marketplace_stock_id
		LEFT JOIN products p ON p.id = s.product_id
		GROUP BY m.branch_id, m.sku_id`).Scan(&costRows).Error
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	costs := map[string]float64{}
	for _, row := range costRows {
		costs[profitKey(row.BranchID, row.Marketplace, row.SkuID)] = row.Cost
	}
	var directCosts []struct {
		SKU  string
		Cost float64
	}
	if err := h.db.Raw("SELECT sku, cost_price AS cost FROM products").Scan(&directCosts).Error; err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	for _, row := range directCosts {
		costs["direct|"+row.SKU] = row.Cost
	}
	var branches []struct {
		ID   string
		Name string
	}
	if err := h.db.Raw("SELECT id::text, name FROM branches").Scan(&branches).Error; err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	branchNames := map[string]string{}
	for _, b := range branches {
		branchNames[b.ID] = b.Name
	}
	storeRows := map[string]*profitabilityRow{}
	skuRows := map[string]*profitabilityRow{}
	getRow := func(target map[string]*profitabilityRow, branchID, mp, sku, name string) *profitabilityRow {
		key := profitKey(branchID, mp, sku)
		if target[key] == nil {
			store := branchNames[branchID]
			if store == "" {
				store = "Tanpa cabang"
			}
			target[key] = &profitabilityRow{BranchID: branchID, Store: store, Marketplace: mp, SellerSKU: sku, ProductName: name}
		}
		return target[key]
	}
	for _, items := range orders {
		first := items[0]
		if strings.Contains(strings.ToUpper(first.OrderStatus), "CANCEL") {
			continue
		}
		branchID := branchString(first.BranchID)
		store := getRow(storeRows, branchID, first.Marketplace, "", "")
		units := 0
		for _, item := range items {
			if item.ItemCount > 0 {
				units += item.ItemCount
			}
		}
		if units == 0 {
			units = 1
		}
		gross, refund := first.TotalAmount, first.RefundAmount
		fee := first.FeeAndTaxAmount
		if fee == 0 {
			fee = first.PlatformFee + first.CommissionFee + first.ShippingFeeDeduction
		}
		// Shopee escrow may include an ads wallet top-up. It is prepaid credit,
		// not ad spend, so actual spend is entered separately in this report.
		if first.Marketplace == "shopee" && first.FeeAndTaxAmount > 0 {
			var breakdown struct {
				AdsTopUp float64 `json:"ads_escrow_top_up_fee_or_technical_support_fee"`
			}
			if json.Unmarshal([]byte(first.FeeTaxBreakdown), &breakdown) == nil {
				fee = math.Max(0, fee-breakdown.AdsTopUp)
			}
		}
		store.Orders++
		store.GrossRevenue += gross
		store.Refunds += refund
		store.MarketplaceFees += fee
		for _, item := range items {
			qty := item.ItemCount
			if qty <= 0 && len(items) == 1 {
				qty = 1
			}
			if qty <= 0 {
				continue
			}
			share := float64(qty) / float64(units)
			sku := item.SellerSKU
			if sku == "" {
				sku = item.SkuID
			}
			if sku == "" {
				sku = "(SKU tidak tersedia)"
			}
			row := getRow(skuRows, branchID, item.Marketplace, sku, item.ProductName)
			row.Orders++
			row.Units += qty
			row.GrossRevenue += gross * share
			row.Refunds += refund * share
			row.MarketplaceFees += fee * share
			cost, ok := costs[profitKey(branchID, item.Marketplace, item.SkuID)]
			if !ok {
				cost, ok = costs[profitKey(branchID, "", item.SkuID)]
			}
			if !ok {
				cost, ok = costs["direct|"+item.SellerSKU]
			}
			if !ok || cost <= 0 {
				row.MissingCOGSUnits += qty
				store.MissingCOGSUnits += qty
			} else {
				row.COGS += cost * float64(qty)
				store.COGS += cost * float64(qty)
			}
			store.Units += qty
		}
	}
	queryExpense := h.db.Model(&profitabilityExpense{}).Where("expense_date >= ? AND expense_date <= ?", from, to)
	if branch != uuid.Nil {
		queryExpense = queryExpense.Where("branch_id = ?", branch)
	}
	if marketplace != "" {
		queryExpense = queryExpense.Where("marketplace = ?", marketplace)
	}
	var expenses []profitabilityExpense
	if err := queryExpense.Order("expense_date DESC, id DESC").Find(&expenses).Error; err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	addExpense := func(row *profitabilityRow, e profitabilityExpense) {
		switch e.Category {
		case "ads":
			row.Ads += e.Amount
			row.AttributedRevenue += e.AttributedRevenue
		case "fulfillment":
			row.Fulfillment += e.Amount
		case "return":
			row.Returns += e.Amount
		case "operational":
			row.Operational += e.Amount
		}
	}
	for _, e := range expenses {
		branchID := e.BranchID.String()
		addExpense(getRow(storeRows, branchID, e.Marketplace, "", ""), e)
		if e.SellerSKU != "" {
			addExpense(getRow(skuRows, branchID, e.Marketplace, e.SellerSKU, ""), e)
		}
	}
	stores := make([]*profitabilityRow, 0, len(storeRows))
	for _, row := range storeRows {
		row.finish()
		stores = append(stores, row)
	}
	products := make([]*profitabilityRow, 0, len(skuRows))
	for _, row := range skuRows {
		row.finish()
		products = append(products, row)
	}
	c.JSON(http.StatusOK, gin.H{"stores": stores, "products": products, "expenses": expenses, "method": "order_total_allocated_by_quantity"})
}

func branchString(id *uuid.UUID) string {
	if id == nil {
		return ""
	}
	return id.String()
}

func (h *MarketplaceHandler) CreateProfitabilityExpense(c *gin.Context) {
	var req struct {
		BranchID          string  `json:"branch_id"`
		Marketplace       string  `json:"marketplace"`
		ExpenseDate       string  `json:"expense_date"`
		Category          string  `json:"category"`
		Amount            float64 `json:"amount"`
		AttributedRevenue float64 `json:"attributed_revenue"`
		SellerSKU         string  `json:"seller_sku"`
		Notes             string  `json:"notes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Data biaya tidak valid"})
		return
	}
	branchID, err := uuid.Parse(req.BranchID)
	date, dateErr := time.Parse("2006-01-02", req.ExpenseDate)
	if err != nil || dateErr != nil || (req.Marketplace != "tiktok" && req.Marketplace != "shopee") ||
		(req.Category != "ads" && req.Category != "fulfillment" && req.Category != "return" && req.Category != "operational") ||
		math.IsNaN(req.Amount) || math.IsInf(req.Amount, 0) || req.Amount <= 0 || math.IsNaN(req.AttributedRevenue) || math.IsInf(req.AttributedRevenue, 0) || req.AttributedRevenue < 0 || (req.Category != "ads" && req.AttributedRevenue != 0) {
		c.JSON(400, gin.H{"error": "Cabang, tanggal, kategori, atau nilai biaya tidak valid"})
		return
	}
	if val, ok := c.Get("branch_id"); ok && val != nil && val != "" {
		if assigned, parseErr := uuid.Parse(val.(string)); parseErr == nil && assigned != branchID {
			c.JSON(403, gin.H{"error": "Cabang tidak sesuai akses"})
			return
		}
	}
	e := profitabilityExpense{BranchID: branchID, Marketplace: req.Marketplace, ExpenseDate: date, Category: req.Category, Amount: req.Amount, AttributedRevenue: req.AttributedRevenue, SellerSKU: strings.TrimSpace(req.SellerSKU), Notes: strings.TrimSpace(req.Notes)}
	if err := h.db.Create(&e).Error; err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(201, gin.H{"data": e})
}

func (h *MarketplaceHandler) DeleteProfitabilityExpense(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"error": "ID tidak valid"})
		return
	}
	query := h.db.Where("id = ?", id)
	if branch := parseBranchID(c); branch != uuid.Nil {
		query = query.Where("branch_id = ?", branch)
	}
	result := query.Delete(&profitabilityExpense{})
	if result.Error != nil {
		c.JSON(500, gin.H{"error": result.Error.Error()})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(404, gin.H{"error": "Biaya tidak ditemukan"})
		return
	}
	c.Status(204)
}
