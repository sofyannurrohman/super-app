package http

import (
	"fmt"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"example.com/hris/backend/internal/usecase"
	"example.com/hris/backend/pkg/utils"
)

type FactoryHandler struct {
	usecase usecase.FactoryUsecase
}

func NewFactoryHandler(u usecase.FactoryUsecase) *FactoryHandler {
	return &FactoryHandler{u}
}

func (h *FactoryHandler) RegisterRoutes(r *gin.RouterGroup) {
	factory := r.Group("/factory")
	{
		factory.GET("", h.GetAllFactories)
		factory.POST("", h.CreateFactory)
		factory.GET("/:id", h.GetFactoryDetail)
		factory.PUT("/:id", h.UpdateFactory)
		factory.DELETE("/:id", h.DeleteFactory)

		factory.GET("/products", h.GetProducts)
		factory.GET("/stock/inventory", h.GetAllInventory)
		factory.GET("/stock/inventory/logs", h.GetAllInventoryLogs)
		factory.GET("/stock/transfers", h.GetAllTransfers)
		factory.GET("/stock/production", h.GetAllProductionHistory)
		factory.GET("/demand", h.GetBackorderDemand)
		factory.GET("/delegated-shipments", h.GetDelegatedShipments)
		factory.GET("/dashboard/stats", h.GetDashboardStats)
		factory.GET("/receivables", h.GetBranchReceivables)
		factory.GET("/stock/low-alerts", h.GetLowStockAlerts)
		factory.POST("/products", h.CreateProduct)
		factory.PUT("/products/:id", h.UpdateProduct)
		factory.DELETE("/products/:id", h.DeleteProduct)

		// Factory Warehouses
		factory.POST("/:id/warehouses", h.CreateFactoryWarehouse)
		factory.GET("/:id/warehouses", h.GetFactoryWarehouses)
		factory.GET("/warehouses/:wid", h.GetFactoryWarehouseDetail)
		factory.PUT("/warehouses/:wid", h.UpdateFactoryWarehouse)
		factory.DELETE("/warehouses/:wid", h.DeleteFactoryWarehouse)

		factory.GET("/:id/inventory", h.GetInventory)
		factory.GET("/:id/inventory/logs", h.GetInventoryLogs)
		factory.GET("/:id/raw-materials", h.GetRawMaterialStock)
		factory.GET("/:id/raw-materials/logs", h.GetRawMaterialLogs)
		factory.POST("/:id/raw-materials/mutate", h.MutateRawMaterial)
		factory.DELETE("/raw-materials/logs/:id", h.DeleteRawMaterialLog)
		factory.DELETE("/stocks/:id", h.DeleteStock)
		factory.POST("/:id/production", h.LogProduction)
		factory.POST("/:id/production/batch", h.LogBatchProduction)
		factory.GET("/:id/production", h.GetProductionHistory)
		factory.PUT("/production/:id", h.UpdateProductionLog)
		factory.DELETE("/production/:id", h.DeleteProductionLog)

		// QC Sample Requests & Approvals
		factory.POST("/:id/qc-samples", h.RequestQcSample)
		factory.GET("/:id/qc-samples", h.GetQcSamples)
		factory.GET("/qc-samples", h.GetAllQcSamples)
		factory.GET("/qc-samples/:id", h.GetQcSampleDetail)
		factory.POST("/qc-samples/:id/approve", h.ApproveQcSample)
		factory.POST("/qc-samples/:id/reject", h.RejectQcSample)

		factory.POST("/:id/inventory/adjust", h.AdjustStock)
		factory.POST("/:id/inventory/sync-warehouse", h.SyncWarehouseStock)
		factory.POST("/sync-all-warehouses", h.SyncAllFactoriesWarehouseStock)
		factory.POST("/:id/transfer/request", h.RequestShipment)
		factory.POST("/transfer/:id/ship", h.ExecuteShipment)
		factory.PUT("/transfer/:id/approve", h.ApproveTransfer)
		factory.PUT("/transfer/:id", h.UpdateTransfer)
		factory.PUT("/transfer/do/:do_no", h.UpdateTransferByDO)
		factory.DELETE("/transfer/:id", h.DeleteTransfer)
		factory.POST("/transfer", h.CreateTransfer)
		factory.GET("/:id/transfer", h.GetTransferHistory)

		factory.GET("/recipes", h.GetRecipes)
		factory.GET("/recipes/:product_id", h.GetRecipeByProduct)
		factory.POST("/recipes", h.CreateRecipe)
		factory.DELETE("/recipes/:id", h.DeleteRecipe)
	}
}

func (h *FactoryHandler) GetAllFactories(c *gin.Context) {
	companyID := uuid.Nil

	// Skip companyID filter dari JWT agar manajemen bisa melihat semua pabrik (universal view)
	// Hal yang sama diterapkan pada branch_handler.go
	/*
	if val, ok := c.Get("companyID"); ok && val != nil {
		if id, ok := val.(uuid.UUID); ok {
			companyID = id
		}
	}
	*/

	if qID := c.Query("company_id"); qID != "" {
		if id, err := uuid.Parse(qID); err == nil {
			companyID = id
		} else if qID == "all" {
			companyID = uuid.Nil
		}
	}

	factories, err := h.usecase.GetAllFactories(companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, factories)
}

func (h *FactoryHandler) CreateFactory(c *gin.Context) {
	var factory domain.Factory
	if err := c.ShouldBindJSON(&factory); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	val, _ := c.Get("companyID")
	factory.CompanyID = val.(uuid.UUID)
	if err := h.usecase.CreateFactory(&factory); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, factory)
}

func (h *FactoryHandler) GetFactoryDetail(c *gin.Context) {
	id := uuid.MustParse(c.Param("id"))
	factory, err := h.usecase.GetFactoryDetail(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Factory not found"})
		return
	}
	c.JSON(http.StatusOK, factory)
}

func (h *FactoryHandler) UpdateFactory(c *gin.Context) {
	id := uuid.MustParse(c.Param("id"))
	var factory domain.Factory
	if err := c.ShouldBindJSON(&factory); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	factory.ID = id
	if err := h.usecase.UpdateFactory(&factory); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, factory)
}

func (h *FactoryHandler) DeleteFactory(c *gin.Context) {
	id := uuid.MustParse(c.Param("id"))
	if err := h.usecase.DeleteFactory(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Factory deleted"})
}

func (h *FactoryHandler) CreateFactoryWarehouse(c *gin.Context) {
	factoryID := uuid.MustParse(c.Param("id"))
	var warehouse domain.FactoryWarehouse
	if err := c.ShouldBindJSON(&warehouse); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	warehouse.FactoryID = factoryID
	if err := h.usecase.CreateFactoryWarehouse(&warehouse); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, warehouse)
}

func (h *FactoryHandler) GetFactoryWarehouses(c *gin.Context) {
	factoryID := uuid.MustParse(c.Param("id"))
	warehouses, err := h.usecase.GetFactoryWarehouses(factoryID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, warehouses)
}

func (h *FactoryHandler) GetFactoryWarehouseDetail(c *gin.Context) {
	id := uuid.MustParse(c.Param("wid"))
	warehouse, err := h.usecase.GetFactoryWarehouseDetail(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Factory warehouse not found"})
		return
	}
	c.JSON(http.StatusOK, warehouse)
}

func (h *FactoryHandler) UpdateFactoryWarehouse(c *gin.Context) {
	id := uuid.MustParse(c.Param("wid"))
	var warehouse domain.FactoryWarehouse
	if err := c.ShouldBindJSON(&warehouse); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	warehouse.ID = id
	if err := h.usecase.UpdateFactoryWarehouse(&warehouse); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, warehouse)
}

func (h *FactoryHandler) DeleteFactoryWarehouse(c *gin.Context) {
	id := uuid.MustParse(c.Param("wid"))
	if err := h.usecase.DeleteFactoryWarehouse(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Factory warehouse deleted"})
}

func (h *FactoryHandler) CreateProduct(c *gin.Context) {
	var product domain.Product
	if err := c.ShouldBind(&product); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Handle Image Upload
	imageURL, err := h.handleImageUpload(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload image: " + err.Error()})
		return
	}
	if imageURL != "" {
		product.ImageURL = imageURL
	}

	val, _ := c.Get("companyID")
	product.CompanyID = val.(uuid.UUID)
	if err := h.usecase.CreateProduct(&product); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, product)
}

func (h *FactoryHandler) UpdateProduct(c *gin.Context) {
	id := uuid.MustParse(c.Param("id"))
	var product domain.Product
	if err := c.ShouldBind(&product); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Handle Image Upload
	imageURL, err := h.handleImageUpload(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload image: " + err.Error()})
		return
	}
	if imageURL != "" {
		product.ImageURL = imageURL
	}

	product.ID = id
	if err := h.usecase.UpdateProduct(&product); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, product)
}

func (h *FactoryHandler) DeleteProduct(c *gin.Context) {
	id := uuid.MustParse(c.Param("id"))
	if err := h.usecase.DeleteProduct(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Product deleted"})
}

func isAncillaryOrGift(cat, n string) bool {
	return strings.Contains(cat, "SAMPINGAN") ||
		strings.Contains(cat, "HADIAH") ||
		strings.Contains(cat, "MERCHANDISE") ||
		strings.Contains(cat, "PROMOSI") ||
		strings.Contains(cat, "MANGKOK") ||
		strings.Contains(cat, "PIRING") ||
		strings.Contains(cat, "GELAS") ||
		strings.Contains(n, "MANGKOK") ||
		strings.Contains(n, "PIRING") ||
		strings.Contains(n, "GELAS") ||
		strings.Contains(n, "SENDOK") ||
		strings.Contains(n, "GARPU")
}

func isProcurementProduct(category, name string) bool {
	cat := strings.ToUpper(strings.TrimSpace(category))
	n := strings.ToUpper(strings.TrimSpace(name))

	// Kategori pengadaan tetap berlaku walaupun nama menyerupai barang hadiah.
	if strings.Contains(cat, "BAHAN") || strings.Contains(cat, "RAW") ||
		strings.Contains(cat, "BUMBU") || strings.Contains(cat, "KEMASAN") ||
		strings.Contains(cat, "PACKAGING") || strings.Contains(cat, "WADAH") ||
		strings.HasPrefix(n, "STIKER ") || strings.HasPrefix(n, "STICKER ") {
		return true
	}

	// Hadiah, Merchandise, Barang Sampingan (Mangkok, Piring, Gelas, dll.) are commercial/sellable products
	if isAncillaryOrGift(cat, n) {
		return false
	}

	// 1. Raw materials & ingredients
	if strings.Contains(cat, "BAHAN") ||
		strings.Contains(cat, "RAW") ||
		strings.Contains(cat, "BUMBU") ||
		strings.Contains(cat, "KEDELAI") ||
		strings.Contains(cat, "GULA") ||
		strings.Contains(cat, "GARAM") ||
		strings.Contains(n, "KEDELAI") ||
		strings.Contains(n, "GULA") ||
		strings.Contains(n, "GARAM") ||
		strings.Contains(n, "TEPUNG") {
		return true
	}

	// 2. Packaging / Containers
	if strings.Contains(cat, "KEMASAN") ||
		strings.Contains(cat, "BOTOL") ||
		strings.Contains(cat, "WADAH") ||
		strings.Contains(cat, "KARTON") ||
		strings.Contains(cat, "DUS") ||
		strings.Contains(cat, "LABEL") ||
		strings.Contains(cat, "TUTUP") ||
		strings.Contains(cat, "PACKAGING") ||
		strings.Contains(cat, "JERIGEN") ||
		strings.Contains(n, "BOTOL KOSONG") ||
		strings.Contains(n, "KARTON") ||
		strings.Contains(n, "TUTUP BOTOL") ||
		strings.Contains(n, "LABEL") {
		return true
	}

	return false
}

func (h *FactoryHandler) GetProducts(c *gin.Context) {
	companyID := uuid.Nil
	if val, ok := c.Get("companyID"); ok {
		companyID = val.(uuid.UUID)
	}

	// Allow override via query param for specific company lookup
	if qID := c.Query("company_id"); qID != "" {
		if id, err := uuid.Parse(qID); err == nil {
			companyID = id
		} else if qID == "all" {
			companyID = uuid.Nil
		}
	}

	branchID := uuid.Nil
	// Check query param for branch_id
	if qBranchID := c.Query("branch_id"); qBranchID != "" {
		if id, err := uuid.Parse(qBranchID); err == nil {
			branchID = id
		}
	}
	// Fallback to branchID set in GIN context by EmployeeMiddleware
	if branchID == uuid.Nil {
		if val, ok := c.Get("branchID"); ok {
			if bid, ok := val.(uuid.UUID); ok {
				branchID = bid
			}
		}
	}

	products, err := h.usecase.GetProducts(companyID, branchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	userRole, exists := c.Get("role")
	roleStr := ""
	if exists {
		if rs, ok := userRole.(string); ok {
			roleStr = strings.ToLower(rs)
		}
	}

	scope := strings.ToLower(c.Query("scope"))
	typeParam := strings.ToLower(c.Query("type"))

	// Filter based on scope / type / role
	// If role is sales-related, non-factory manager, or client requested scope=sales / type=finished_goods,
	// strictly filter OUT any raw materials, ingredients, packaging, or procurement items.
	isFactoryManager := roleStr == "superadmin" || roleStr == "manager" || roleStr == "admin_pabrik" || roleStr == "admin_gudang" || roleStr == "supervisor_gudang" || roleStr == "admin_purchasing" || roleStr == "hr_admin" || roleStr == "admin_nota"
	isSalesRole := roleStr == "sales" || strings.Contains(roleStr, "sales") || strings.Contains(roleStr, "driver") || strings.Contains(roleStr, "kolektor") || strings.Contains(roleStr, "customer") || strings.Contains(roleStr, "toko") || strings.Contains(roleStr, "motoris") || strings.Contains(roleStr, "canvas")

	isFinishedGoodsOnly := scope == "sales" || scope == "commercial" || typeParam == "finished_goods" || (isSalesRole && typeParam != "all" && scope != "all") || (!isFactoryManager && typeParam != "raw_material" && scope != "raw_material" && typeParam != "all" && scope != "all")
	isRawMaterialsOnly := scope == "raw_material" || scope == "procurement" || typeParam == "raw_material" || typeParam == "raw_materials"

	if isFinishedGoodsOnly {
		var filtered []domain.Product
		for _, p := range products {
			if !isProcurementProduct(p.Category, p.Name) {
				filtered = append(filtered, p)
			}
		}
		products = filtered
	} else if isRawMaterialsOnly {
		var filtered []domain.Product
		for _, p := range products {
			if isProcurementProduct(p.Category, p.Name) {
				filtered = append(filtered, p)
			}
		}
		products = filtered
	}

	// Strict Price Masking:
	// Raw material and procurement prices (HPP/CostPrice) are strictly confidential to Purchasing & Management.
	// Admin Gudang, Supervisor Gudang, Admin Pabrik, Operator, etc. MUST NOT see raw material prices.
	canViewRawMaterialPrices := roleStr == "superadmin" || roleStr == "admin_purchasing" || roleStr == "manager"
	canViewFinishedGoodsPrices := roleStr == "superadmin" || roleStr == "manager" || roleStr == "hr_admin" || roleStr == "admin_gudang" || roleStr == "admin_pabrik" || roleStr == "supervisor_gudang" || roleStr == "admin_purchasing" || roleStr == "admin_nota"

	for i := range products {
		if isProcurementProduct(products[i].Category, products[i].Name) {
			if !canViewRawMaterialPrices {
				products[i].CostPrice = 0
				products[i].FactoryPrice = 0
				products[i].SellingPrice = 0
			}
		} else {
			if !canViewFinishedGoodsPrices {
				products[i].CostPrice = 0
				products[i].FactoryPrice = 0
			}
		}
	}

	c.JSON(http.StatusOK, products)
}

func (h *FactoryHandler) GetInventory(c *gin.Context) {
	id := uuid.MustParse(c.Param("id"))
	inventory, err := h.usecase.GetFactoryInventory(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, inventory)
}

func (h *FactoryHandler) GetAllInventory(c *gin.Context) {
	val, _ := c.Get("companyID")
	companyID := val.(uuid.UUID)
	inventory, err := h.usecase.GetAllInventory(companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, inventory)
}

func (h *FactoryHandler) GetAllInventoryLogs(c *gin.Context) {
	val, exists := c.Get("companyID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	companyID, ok := val.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid company ID"})
		return
	}
	logs, err := h.usecase.GetAllInventoryLogs(companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, logs)
}

func (h *FactoryHandler) GetInventoryLogs(c *gin.Context) {
	id := uuid.MustParse(c.Param("id"))
	logs, err := h.usecase.GetInventoryLogs(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, logs)
}

func (h *FactoryHandler) LogProduction(c *gin.Context) {
	factoryID := uuid.MustParse(c.Param("id"))
	var req struct {
		ProductID         uuid.UUID                            `json:"product_id" binding:"required"`
		EmployeeID        *uuid.UUID                           `json:"employee_id"`
		OperatorName      string                               `json:"operator_name"`
		Operators         []usecase.ProductionOperatorInput    `json:"operators"`
		TargetWarehouseID *uuid.UUID                           `json:"target_warehouse_id"`
		Quantity          *int                                 `json:"quantity" binding:"required"`
		CartonCount       int                                  `json:"carton_count"`
		PiecesPerCarton   int                                  `json:"pieces_per_carton"`
		Notes             string                               `json:"notes"`
		BatchNo           string                               `json:"batch_no"`
		ExpiryDate        *string                              `json:"expiry_date"`
		RawMaterials      []usecase.ProductionRawMaterialUsage `json:"raw_materials"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if (req.EmployeeID == nil || *req.EmployeeID == uuid.Nil) && strings.TrimSpace(req.OperatorName) == "" && len(req.Operators) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "minimal 1 operator produksi harus dipilih atau diisi secara manual"})
		return
	}

	var parsedExpiryDate *time.Time
	if req.ExpiryDate != nil && *req.ExpiryDate != "" {
		t, err := time.Parse("2006-01-02", *req.ExpiryDate)
		if err == nil {
			parsedExpiryDate = &t
		} else {
			if t2, err2 := time.Parse(time.RFC3339, *req.ExpiryDate); err2 == nil {
				parsedExpiryDate = &t2
			} else {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid expiry_date format"})
				return
			}
		}
	}

	if err := h.usecase.LogProduction(factoryID, req.ProductID, req.EmployeeID, req.TargetWarehouseID, *req.Quantity, req.CartonCount, req.PiecesPerCarton, req.Notes, req.BatchNo, parsedExpiryDate, req.RawMaterials, req.Operators); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Production logged successfully"})
}

func (h *FactoryHandler) LogBatchProduction(c *gin.Context) {
	factoryID := uuid.MustParse(c.Param("id"))
	var req struct {
		EmployeeID   *uuid.UUID                        `json:"employee_id"`
		OperatorName string                            `json:"operator_name"`
		Operators    []usecase.ProductionOperatorInput `json:"operators"`
		Notes        string                            `json:"notes"`
		Items        []struct {
			ProductID         uuid.UUID  `json:"product_id" binding:"required"`
			TargetWarehouseID *uuid.UUID `json:"target_warehouse_id"`
			BatchNo           string     `json:"batch_no"`
			ExpiryDate        *string    `json:"expiry_date"`
			Quantity          *int       `json:"quantity" binding:"required"`
			CartonCount       int        `json:"carton_count"`
			PiecesPerCarton   int        `json:"pieces_per_carton"`
			Notes             string     `json:"notes"`
		} `json:"items" binding:"required,min=1"`
		RawMaterials []usecase.ProductionRawMaterialUsage `json:"raw_materials"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if (req.EmployeeID == nil || *req.EmployeeID == uuid.Nil) && strings.TrimSpace(req.OperatorName) == "" && len(req.Operators) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "minimal 1 operator produksi harus dipilih atau diisi secara manual"})
		return
	}

	items := make([]usecase.BatchProductionItemInput, 0, len(req.Items))
	for _, it := range req.Items {
		if it.Quantity == nil || *it.Quantity <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "quantity produk harus lebih dari 0"})
			return
		}

		var parsedExpiryDate *time.Time
		if it.ExpiryDate != nil && *it.ExpiryDate != "" {
			t, err := time.Parse("2006-01-02", *it.ExpiryDate)
			if err == nil {
				parsedExpiryDate = &t
			} else {
				if t2, err2 := time.Parse(time.RFC3339, *it.ExpiryDate); err2 == nil {
					parsedExpiryDate = &t2
				} else {
					c.JSON(http.StatusBadRequest, gin.H{"error": "format expiry_date tidak valid"})
					return
				}
			}
		}

		items = append(items, usecase.BatchProductionItemInput{
			ProductID:         it.ProductID,
			TargetWarehouseID: it.TargetWarehouseID,
			BatchNo:           it.BatchNo,
			ExpiryDate:        parsedExpiryDate,
			Quantity:          *it.Quantity,
			CartonCount:       it.CartonCount,
			PiecesPerCarton:   it.PiecesPerCarton,
			Notes:             it.Notes,
		})
	}

	if err := h.usecase.LogBatchProduction(factoryID, req.EmployeeID, req.Notes, items, req.RawMaterials, req.Operators); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Batch production logged successfully"})
}

func (h *FactoryHandler) canViewProductionMaterials(c *gin.Context) bool {
	roleVal, _ := c.Get("role")
	role := strings.ToLower(strings.TrimSpace(fmt.Sprintf("%v", roleVal)))
	if role == "superadmin" || role == "admin" || role == "supervisor_gudang" {
		return true
	}

	emailVal, exists := c.Get("userEmail")
	if !exists || emailVal == nil {
		emailVal, _ = c.Get("email")
	}
	email := strings.ToLower(strings.TrimSpace(fmt.Sprintf("%v", emailVal)))

	if (role == "admin_pabrik" || strings.Contains(role, "pabrik")) && email == "rizkikusumaww@gmail.com" {
		return true
	}

	return false
}

func (h *FactoryHandler) GetProductionHistory(c *gin.Context) {
	id := uuid.MustParse(c.Param("id"))
	history, err := h.usecase.GetProductionHistory(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !h.canViewProductionMaterials(c) {
		for i := range history {
			history[i].Materials = nil
		}
	}
	c.JSON(http.StatusOK, history)
}

func (h *FactoryHandler) GetAllProductionHistory(c *gin.Context) {
	val, _ := c.Get("companyID")
	companyID := val.(uuid.UUID)
	history, err := h.usecase.GetAllProductionHistory(companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !h.canViewProductionMaterials(c) {
		for i := range history {
			history[i].Materials = nil
		}
	}
	c.JSON(http.StatusOK, history)
}

func (h *FactoryHandler) RequestShipment(c *gin.Context) {
	factoryID := uuid.MustParse(c.Param("id"))
	var req struct {
		ToBranchID         uuid.UUID                    `json:"to_branch_id" binding:"required"`
		Items              []domain.ProductTransferItem `json:"items" binding:"required"`
		Notes              string                       `json:"notes"`
		EstimatedArrival   *string                      `json:"estimated_arrival"`
		TargetShipmentDate *string                      `json:"target_shipment_date"`
		InitiatedBy        string                       `json:"initiated_by"`
		PaymentMethod      string                       `json:"payment_method"`
		PaymentDueAt       *string                      `json:"payment_due_at"`
		VehicleID          *uuid.UUID                   `json:"vehicle_id"`
		DriverID           *uuid.UUID                   `json:"driver_id"`
		TransferType       string                       `json:"transfer_type"`
		IsBillable         *bool                        `json:"is_billable"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	initiatedBy := req.InitiatedBy
	if initiatedBy == "" {
		initiatedBy = "FACTORY"
	}

	paymentMethod := req.PaymentMethod
	if paymentMethod == "" {
		paymentMethod = "NET"
	}

	var parsedEstimatedArrival *time.Time
	if req.EstimatedArrival != nil && *req.EstimatedArrival != "" {
		if t, err := time.Parse("2006-01-02T15:04", *req.EstimatedArrival); err == nil {
			parsedEstimatedArrival = &t
		} else if t2, err2 := time.Parse(time.RFC3339, *req.EstimatedArrival); err2 == nil {
			parsedEstimatedArrival = &t2
		} else if t3, err3 := time.Parse("2006-01-02", *req.EstimatedArrival); err3 == nil {
			parsedEstimatedArrival = &t3
		}
	}

	var parsedTargetShipmentDate *time.Time
	if req.TargetShipmentDate != nil && *req.TargetShipmentDate != "" {
		if t, err := time.Parse("2006-01-02T15:04", *req.TargetShipmentDate); err == nil {
			parsedTargetShipmentDate = &t
		} else if t2, err2 := time.Parse(time.RFC3339, *req.TargetShipmentDate); err2 == nil {
			parsedTargetShipmentDate = &t2
		} else if t3, err3 := time.Parse("2006-01-02", *req.TargetShipmentDate); err3 == nil {
			parsedTargetShipmentDate = &t3
		}
	}

	var parsedPaymentDueAt *time.Time
	if req.PaymentDueAt != nil && *req.PaymentDueAt != "" {
		if t, err := time.Parse("2006-01-02", *req.PaymentDueAt); err == nil {
			parsedPaymentDueAt = &t
		} else if t2, err2 := time.Parse(time.RFC3339, *req.PaymentDueAt); err2 == nil {
			parsedPaymentDueAt = &t2
		}
	}

	if err := h.usecase.RequestShipment(factoryID, req.ToBranchID, req.Items, req.Notes, parsedTargetShipmentDate, parsedEstimatedArrival, initiatedBy, req.VehicleID, req.DriverID, paymentMethod, parsedPaymentDueAt, req.TransferType, req.IsBillable); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Shipment request created"})
}

func (h *FactoryHandler) ExecuteShipment(c *gin.Context) {
	transferID := uuid.MustParse(c.Param("id"))
	if err := h.usecase.ExecuteApprovedShipment(transferID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Shipment executed and factory stock reduced"})
}

func (h *FactoryHandler) ApproveTransfer(c *gin.Context) {
	transferID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid transfer ID"})
		return
	}

	var approverID *uuid.UUID
	if val, ok := c.Get("employeeID"); ok && val != nil {
		if empID, ok := val.(uuid.UUID); ok && empID != uuid.Nil {
			approverID = &empID
		}
	}
	if approverID == nil {
		if val, ok := c.Get("userID"); ok && val != nil {
			if uid, ok := val.(uuid.UUID); ok && uid != uuid.Nil {
				approverID = &uid
			}
		}
	}

	approverName := ""
	if val, ok := c.Get("name"); ok && val != nil {
		approverName = val.(string)
	}

	approverPosition := "Admin Pabrik"
	if val, ok := c.Get("job_position"); ok && val != nil && val.(string) != "" {
		approverPosition = val.(string)
	} else if rVal, ok := c.Get("role"); ok && rVal != nil {
		approverPosition = strings.ToUpper(strings.ReplaceAll(rVal.(string), "_", " "))
	}

	if err := h.usecase.ApproveTransfer(transferID, approverID, approverName, approverPosition); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Transfer approved"})
}

func (h *FactoryHandler) UpdateProductionLog(c *gin.Context) {
	id := uuid.MustParse(c.Param("id"))
	var req struct {
		ProductID         *uuid.UUID                        `json:"product_id"`
		TargetWarehouseID *uuid.UUID                        `json:"target_warehouse_id"`
		BatchNo           *string                           `json:"batch_no"`
		ExpiryDate        *string                           `json:"expiry_date"`
		Quantity          *int                              `json:"quantity" binding:"required"`
		CartonCount       int                               `json:"carton_count"`
		PiecesPerCarton   int                               `json:"pieces_per_carton"`
		Notes             string                            `json:"notes"`
		EmployeeID        *uuid.UUID                           `json:"employee_id"`
		Operators         []usecase.ProductionOperatorInput    `json:"operators"`
		RawMaterials      []usecase.ProductionRawMaterialUsage `json:"raw_materials"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var parsedExpiryDate *time.Time
	if req.ExpiryDate != nil && strings.TrimSpace(*req.ExpiryDate) != "" {
		dateStr := strings.TrimSpace(*req.ExpiryDate)
		if t, err := time.Parse("2006-01-02", dateStr); err == nil {
			parsedExpiryDate = &t
		} else if t2, err2 := time.Parse(time.RFC3339, dateStr); err2 == nil {
			parsedExpiryDate = &t2
		} else if t3, err3 := time.Parse("2006-01-02T15:04", dateStr); err3 == nil {
			parsedExpiryDate = &t3
		} else if t4, err4 := time.Parse("2006-01-02 15:04:05", dateStr); err4 == nil {
			parsedExpiryDate = &t4
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "format expiry_date tidak valid (gunakan format YYYY-MM-DD)"})
			return
		}
	}

	if err := h.usecase.UpdateProductionLog(id, req.ProductID, req.TargetWarehouseID, req.BatchNo, parsedExpiryDate, *req.Quantity, req.CartonCount, req.PiecesPerCarton, req.Notes, req.EmployeeID, req.Operators, req.RawMaterials); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Production log updated successfully"})
}

func (h *FactoryHandler) DeleteProductionLog(c *gin.Context) {
	id := uuid.MustParse(c.Param("id"))
	if err := h.usecase.DeleteProductionLog(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Production log deleted and stock reversed"})
}

func (h *FactoryHandler) AdjustStock(c *gin.Context) {
	factoryID := uuid.MustParse(c.Param("id"))
	var req struct {
		ProductID          uuid.UUID  `json:"product_id" binding:"required"`
		Quantity           *int       `json:"quantity" binding:"required"`
		Reason             string     `json:"reason"`
		BatchNo            string     `json:"batch_no"`
		FactoryWarehouseID *uuid.UUID `json:"factory_warehouse_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.usecase.AdjustStock(factoryID, req.ProductID, *req.Quantity, req.Reason, req.BatchNo, req.FactoryWarehouseID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Stock adjusted successfully"})
}

func (h *FactoryHandler) SyncWarehouseStock(c *gin.Context) {
	factoryID := uuid.MustParse(c.Param("id"))
	if err := h.usecase.SyncFactoryWarehouseStock(factoryID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Stok pabrik dan stok gudang berhasil disinkronkan 1:1"})
}

func (h *FactoryHandler) SyncAllFactoriesWarehouseStock(c *gin.Context) {
	var companyID uuid.UUID
	if cid, exists := c.Get("company_id"); exists && cid != nil {
		if cidUUID, ok := cid.(uuid.UUID); ok {
			companyID = cidUUID
		}
	}
	if err := h.usecase.SyncAllFactoriesWarehouseStock(companyID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Seluruh stok pabrik dan stok gudang berhasil disinkronkan 1:1"})
}

func (h *FactoryHandler) GetTransferHistory(c *gin.Context) {
	id := uuid.MustParse(c.Param("id"))
	history, err := h.usecase.GetTransferHistory(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, history)
}

func (h *FactoryHandler) UpdateTransfer(c *gin.Context) {
	id := uuid.MustParse(c.Param("id"))
	var data map[string]interface{}
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.usecase.UpdateTransfer(id, data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Transfer updated"})
}

func (h *FactoryHandler) UpdateTransferByDO(c *gin.Context) {
	doNo := c.Param("do_no")
	if doNo == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nomor surat jalan (DO) wajib diisi"})
		return
	}

	var req struct {
		FromFactoryID      uuid.UUID                    `json:"from_factory_id" binding:"required"`
		ToBranchID         uuid.UUID                    `json:"to_branch_id" binding:"required"`
		VehicleID          *uuid.UUID                   `json:"vehicle_id"`
		DriverID           *uuid.UUID                   `json:"driver_id"`
		TargetShipmentDate *string                      `json:"target_shipment_date"`
		EstimatedArrival   *string                      `json:"estimated_arrival"`
		PaymentMethod      string                       `json:"payment_method"`
		PaymentDueAt       *string                      `json:"payment_due_at"`
		TransferType       string                       `json:"transfer_type"`
		IsBillable         *bool                        `json:"is_billable"`
		Notes              string                       `json:"notes"`
		Items              []domain.ProductTransferItem `json:"items" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var parsedTargetShipmentDate *time.Time
	if req.TargetShipmentDate != nil && *req.TargetShipmentDate != "" {
		if t, err := time.Parse("2006-01-02T15:04", *req.TargetShipmentDate); err == nil {
			parsedTargetShipmentDate = &t
		} else if t2, err2 := time.Parse(time.RFC3339, *req.TargetShipmentDate); err2 == nil {
			parsedTargetShipmentDate = &t2
		} else if t3, err3 := time.Parse("2006-01-02", *req.TargetShipmentDate); err3 == nil {
			parsedTargetShipmentDate = &t3
		}
	}

	var parsedEstimatedArrival *time.Time
	if req.EstimatedArrival != nil && *req.EstimatedArrival != "" {
		if t, err := time.Parse("2006-01-02T15:04", *req.EstimatedArrival); err == nil {
			parsedEstimatedArrival = &t
		} else if t2, err2 := time.Parse(time.RFC3339, *req.EstimatedArrival); err2 == nil {
			parsedEstimatedArrival = &t2
		} else if t3, err3 := time.Parse("2006-01-02", *req.EstimatedArrival); err3 == nil {
			parsedEstimatedArrival = &t3
		}
	}

	var parsedPaymentDueAt *time.Time
	if req.PaymentDueAt != nil && *req.PaymentDueAt != "" {
		if t, err := time.Parse("2006-01-02", *req.PaymentDueAt); err == nil {
			parsedPaymentDueAt = &t
		} else if t2, err2 := time.Parse(time.RFC3339, *req.PaymentDueAt); err2 == nil {
			parsedPaymentDueAt = &t2
		}
	}

	paymentMethod := req.PaymentMethod
	if paymentMethod == "" {
		paymentMethod = "NET"
	}

	input := usecase.UpdateTransferDOInput{
		FromFactoryID:      req.FromFactoryID,
		ToBranchID:         req.ToBranchID,
		VehicleID:          req.VehicleID,
		DriverID:           req.DriverID,
		TargetShipmentDate: parsedTargetShipmentDate,
		EstimatedArrival:   parsedEstimatedArrival,
		PaymentMethod:      paymentMethod,
		PaymentDueAt:       parsedPaymentDueAt,
		TransferType:       req.TransferType,
		IsBillable:         req.IsBillable,
		Notes:              req.Notes,
		Items:              req.Items,
	}

	if err := h.usecase.UpdateTransferByDO(doNo, input); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Faktur pengiriman cabang berhasil diperbarui"})
}

func (h *FactoryHandler) DeleteTransfer(c *gin.Context) {
	id := uuid.MustParse(c.Param("id"))
	if err := h.usecase.DeleteTransfer(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Transfer deleted"})
}

func (h *FactoryHandler) CreateTransfer(c *gin.Context) {
	var transfer domain.ProductTransfer
	if err := c.ShouldBindJSON(&transfer); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.usecase.CreateTransfer(&transfer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, transfer)
}

func (h *FactoryHandler) GetDelegatedShipments(c *gin.Context) {
	val, _ := c.Get("companyID")
	companyID := val.(uuid.UUID)

	shipments, summary, err := h.usecase.GetDelegatedShipments(companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"shipments": shipments,
		"summary":   summary,
	})
}

func (h *FactoryHandler) GetAllTransfers(c *gin.Context) {
	companyID := uuid.Nil
	if val, ok := c.Get("companyID"); ok && val != nil {
		if id, ok := val.(uuid.UUID); ok {
			companyID = id
		}
	}
	if qID := c.Query("company_id"); qID != "" {
		if id, err := uuid.Parse(qID); err == nil {
			companyID = id
		} else if qID == "all" || qID == "ALL" {
			companyID = uuid.Nil
		}
	}

	pageStr := c.Query("page")
	limitStr := c.Query("limit")
	branchID := c.Query("branch_id")
	status := c.Query("status")
	search := c.Query("search")
	productID := c.Query("product_id")
	batchNo := c.Query("batch_no")

	if pageStr != "" || limitStr != "" || branchID != "" || (status != "" && status != "ALL") || search != "" {
		page, _ := strconv.Atoi(pageStr)
		limit, _ := strconv.Atoi(limitStr)
		if page < 1 && limit > 0 {
			page = 1
		}

		if page > 0 && limit > 0 {
			transfers, total, totalAmount, stats, err := h.usecase.GetAllTransfersPaginated(page, limit, companyID, branchID, status, search)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			totalPages := 0
			if limit > 0 {
				totalPages = int(math.Ceil(float64(total) / float64(limit)))
			}
			c.JSON(http.StatusOK, gin.H{
				"data":         transfers,
				"total":        total,
				"total_amount": totalAmount,
				"page":         page,
				"limit":        limit,
				"total_pages":  totalPages,
				"stats":        stats,
			})
			return
		}
	}

	history, err := h.usecase.GetAllTransfers(companyID, productID, batchNo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, history)
}

func (h *FactoryHandler) handleImageUpload(c *gin.Context) (string, error) {
	contentType := c.GetHeader("Content-Type")
	if !strings.HasPrefix(contentType, "multipart/form-data") {
		return "", nil
	}

	file, err := c.FormFile("image")
	if err != nil {
		if err == http.ErrMissingFile {
			return "", nil
		}
		return "", err
	}

	// Limit 5MB
	if file.Size > 5*1024*1024 {
		return "", fmt.Errorf("file size exceeds 5MB limit")
	}

	filename := uuid.New().String() + filepath.Ext(file.Filename)
	uploadDir := "uploads/products"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return "", err
	}

	savePath := filepath.Join(uploadDir, filename)
	if err := utils.CompressAndSaveImage(file, savePath, 1000); err != nil {
		return "", err
	}

	return "/uploads/products/" + filename, nil
}

func (h *FactoryHandler) GetRecipes(c *gin.Context) {
	val, _ := c.Get("companyID")
	companyID := val.(uuid.UUID)
	recipes, err := h.usecase.GetRecipes(companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, recipes)
}

func (h *FactoryHandler) GetRecipeByProduct(c *gin.Context) {
	productID := uuid.MustParse(c.Param("product_id"))
	recipe, err := h.usecase.GetRecipeByProduct(productID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if recipe == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Recipe not found"})
		return
	}
	c.JSON(http.StatusOK, recipe)
}

func (h *FactoryHandler) CreateRecipe(c *gin.Context) {
	var recipe domain.ProductionRecipe
	if err := c.ShouldBindJSON(&recipe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.usecase.CreateRecipe(&recipe); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, recipe)
}

func (h *FactoryHandler) DeleteRecipe(c *gin.Context) {
	id := uuid.MustParse(c.Param("id"))
	if err := h.usecase.DeleteRecipe(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Recipe deleted"})
}

func (h *FactoryHandler) GetBackorderDemand(c *gin.Context) {
	companyID, _ := uuid.Parse(c.Query("company_id"))
	demand, err := h.usecase.GetBackorderDemand(companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": demand})
}

func (h *FactoryHandler) GetBranchReceivables(c *gin.Context) {
	companyIDStr := c.GetString("company_id")
	if companyIDStr == "" {
		utils.ErrorResponse(c, http.StatusUnauthorized, "company_id not found")
		return
	}
	companyID, err := uuid.Parse(companyIDStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid company_id")
		return
	}

	response, err := h.usecase.GetBranchReceivables(companyID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": response})
}

func (h *FactoryHandler) GetDashboardStats(c *gin.Context) {
	val, _ := c.Get("companyID")
	companyID := val.(uuid.UUID)

	stats, err := h.usecase.GetDashboardStats(companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, stats)
}

func (h *FactoryHandler) GetLowStockAlerts(c *gin.Context) {
	val, ok := c.Get("companyID")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "companyID not found in session"})
		return
	}
	companyID := val.(uuid.UUID)

	alerts, err := h.usecase.GetLowStockAlerts(companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, alerts)
}

func (h *FactoryHandler) GetRawMaterialStock(c *gin.Context) {
	id := uuid.MustParse(c.Param("id"))
	stock, err := h.usecase.GetRawMaterialStock(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, stock)
}

func (h *FactoryHandler) GetRawMaterialLogs(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid factory ID"})
		return
	}

	var productID *uuid.UUID
	productIDStr := strings.TrimSpace(c.Query("product_id"))
	if productIDStr != "" {
		pID, err := uuid.Parse(productIDStr)
		if err == nil {
			productID = &pID
		}
	}

	logs, err := h.usecase.GetRawMaterialLogs(id, productID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, logs)
}

func (h *FactoryHandler) MutateRawMaterial(c *gin.Context) {
	id := uuid.MustParse(c.Param("id"))
	var req struct {
		ProductID         uuid.UUID  `json:"product_id" binding:"required"`
		Type              string     `json:"type" binding:"required"` // IN, OUT, ADJUSTMENT
		Quantity          int        `json:"quantity" binding:"required"`
		Notes             string     `json:"notes"`
		BatchNo           string     `json:"batch_no"`
		TargetWarehouseID *uuid.UUID `json:"target_warehouse_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.BatchNo == "" {
		req.BatchNo = "DEFAULT"
	}

	if err := h.usecase.MutateRawMaterial(id, req.ProductID, req.Type, req.Quantity, req.Notes, req.BatchNo, req.TargetWarehouseID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Raw material mutated successfully"})
}

func (h *FactoryHandler) DeleteStock(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid stock ID"})
		return
	}
	if err := h.usecase.DeleteStock(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Stock deleted successfully"})
}

func (h *FactoryHandler) DeleteRawMaterialLog(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid log ID"})
		return
	}
	if err := h.usecase.DeleteInventoryLog(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Log deleted successfully"})
}

func (h *FactoryHandler) RequestQcSample(c *gin.Context) {
	factoryID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid factory ID"})
		return
	}

	var req struct {
		ProductID       uuid.UUID  `json:"product_id" binding:"required"`
		ProductionLogID *uuid.UUID `json:"production_log_id"`
		BatchNo         string     `json:"batch_no" binding:"required"`
		Quantity        int        `json:"quantity" binding:"required,min=1"`
		SampleType      string     `json:"sample_type"`
		Purpose         string     `json:"purpose"`
		Notes           string     `json:"notes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var requestedByID uuid.UUID
	if val, ok := c.Get("employeeID"); ok && val != nil {
		if eid, ok := val.(uuid.UUID); ok {
			requestedByID = eid
		}
	} else if val, ok := c.Get("userID"); ok && val != nil {
		if uidStr, ok := val.(string); ok {
			requestedByID, _ = uuid.Parse(uidStr)
		} else if uid, ok := val.(uuid.UUID); ok {
			requestedByID = uid
		}
	}

	requestedByName := "Admin Pabrik"
	if val, ok := c.Get("employeeName"); ok && val != nil && val.(string) != "" {
		requestedByName = val.(string)
	} else if val, ok := c.Get("username"); ok && val != nil && val.(string) != "" {
		requestedByName = val.(string)
	}

	sample, err := h.usecase.RequestQcSample(usecase.RequestQcSampleInput{
		FactoryID:       factoryID,
		ProductID:       req.ProductID,
		ProductionLogID: req.ProductionLogID,
		BatchNo:         req.BatchNo,
		Quantity:        req.Quantity,
		SampleType:      req.SampleType,
		Purpose:         req.Purpose,
		Notes:           req.Notes,
		RequestedByID:   requestedByID,
		RequestedByName: requestedByName,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Request sampel QC berhasil dibuat", "data": sample})
}

func (h *FactoryHandler) GetQcSamples(c *gin.Context) {
	factoryID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid factory ID"})
		return
	}
	status := c.Query("status")
	samples, err := h.usecase.GetQcSamples(factoryID, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, samples)
}

func (h *FactoryHandler) GetAllQcSamples(c *gin.Context) {
	companyID, _ := uuid.Parse(c.Query("company_id"))
	status := c.Query("status")
	samples, err := h.usecase.GetAllQcSamples(companyID, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, samples)
}

func (h *FactoryHandler) GetQcSampleDetail(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sample ID"})
		return
	}
	sample, err := h.usecase.GetQcSampleDetail(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, sample)
}

func (h *FactoryHandler) ApproveQcSample(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sample ID"})
		return
	}

	var approverID *uuid.UUID
	if val, ok := c.Get("employeeID"); ok && val != nil {
		if eid, ok := val.(uuid.UUID); ok {
			approverID = &eid
		}
	} else if val, ok := c.Get("userID"); ok && val != nil {
		if uidStr, ok := val.(string); ok {
			if uid, err := uuid.Parse(uidStr); err == nil {
				approverID = &uid
			}
		} else if uid, ok := val.(uuid.UUID); ok {
			approverID = &uid
		}
	}

	approverName := "Admin Gudang Pabrik"
	if val, ok := c.Get("employeeName"); ok && val != nil && val.(string) != "" {
		approverName = val.(string)
	} else if val, ok := c.Get("username"); ok && val != nil && val.(string) != "" {
		approverName = val.(string)
	}

	if err := h.usecase.ApproveQcSample(id, approverID, approverName); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Sampel QC berhasil disetujui dan stok gudang terpotong"})
}

func (h *FactoryHandler) RejectQcSample(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sample ID"})
		return
	}

	var req struct {
		Reason string `json:"reason" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Alasan penolakan wajib diisi"})
		return
	}

	var approverID *uuid.UUID
	if val, ok := c.Get("employeeID"); ok && val != nil {
		if eid, ok := val.(uuid.UUID); ok {
			approverID = &eid
		}
	} else if val, ok := c.Get("userID"); ok && val != nil {
		if uidStr, ok := val.(string); ok {
			if uid, err := uuid.Parse(uidStr); err == nil {
				approverID = &uid
			}
		} else if uid, ok := val.(uuid.UUID); ok {
			approverID = &uid
		}
	}

	approverName := "Admin Gudang Pabrik"
	if val, ok := c.Get("employeeName"); ok && val != nil && val.(string) != "" {
		approverName = val.(string)
	} else if val, ok := c.Get("username"); ok && val != nil && val.(string) != "" {
		approverName = val.(string)
	}

	if err := h.usecase.RejectQcSample(id, approverID, approverName, req.Reason); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Request sampel QC telah ditolak"})
}

