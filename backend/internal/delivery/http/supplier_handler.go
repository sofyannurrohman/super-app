package http

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"example.com/hris/backend/internal/usecase"
)

type SupplierHandler struct {
	usecase usecase.SupplierUsecase
}

func NewSupplierHandler(u usecase.SupplierUsecase) *SupplierHandler {
	return &SupplierHandler{u}
}

func (h *SupplierHandler) RegisterRoutes(r *gin.RouterGroup) {
	adminRoles := RoleMiddleware(
		string(domain.RoleSuperAdmin),
		string(domain.RoleHRAdmin),
		string(domain.RoleManager),
		string(domain.RoleAdminGudang),
	)

	purchasingRoles := RoleMiddleware(
		string(domain.RoleSuperAdmin),
		string(domain.RoleAdminPurchasing),
	)

	// ── Supplier CRUD (Khusus Purchasing & Superadmin) ─────────────────────────
	suppliers := r.Group("/warehouse/suppliers")
	suppliers.Use(purchasingRoles)
	{
		suppliers.GET("", h.ListSuppliers)
		suppliers.POST("", h.CreateSupplier)
		suppliers.POST("/bulk", h.BulkCreateSuppliers)
		suppliers.GET("/:id", h.GetSupplier)
		suppliers.PUT("/:id", h.UpdateSupplier)
		suppliers.DELETE("/:id", h.DeleteSupplier)
	}

	// ── Supplier Receipts ─────────────────────────────────────────────────────
	receipts := r.Group("/warehouse/supplier-receipts")
	receipts.Use(adminRoles)
	{
		receipts.GET("", h.ListSupplierReceipts)
		receipts.POST("", h.CreateSupplierReceipt)
		receipts.GET("/:id", h.GetSupplierReceipt)
		receipts.PUT("/:id", h.UpdateSupplierReceipt)
		receipts.POST("/:id/receive", h.ReceiveSupplierReceipt)
		receipts.POST("/:id/pay", h.PaySupplier)
		receipts.DELETE("/:id", h.DeleteSupplierReceipt)
	}

	// ── Supplier Stock ────────────────────────────────────────────────────────
	stock := r.Group("/warehouse/supplier-stock")
	stock.Use(adminRoles)
	{
		stock.GET("", h.GetSupplierStock)
	}
}

func (h *SupplierHandler) PaySupplier(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid receipt id"})
		return
	}
	var req usecase.SupplierPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	date := time.Now()
	if req.PaymentDate != nil {
		date = *req.PaymentDate
	}
	if err := h.usecase.PaySupplier(id, req.Amount, date, req.ReferenceNo, req.Notes); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "pembayaran supplier berhasil dicatat"})
}

// ─── Supplier Handlers ────────────────────────────────────────────────────────

func (h *SupplierHandler) ListSuppliers(c *gin.Context) {
	companyIDStr, _ := c.Get("company_id")
	companyID := uuid.Nil
	if companyIDStr != nil {
		if id, err := uuid.Parse(companyIDStr.(string)); err == nil {
			companyID = id
		}
	}

	suppliers, err := h.usecase.ListSuppliers(companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": suppliers})
}

func (h *SupplierHandler) GetSupplier(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}
	supplier, err := h.usecase.GetSupplierByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Supplier tidak ditemukan"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": supplier})
}

func (h *SupplierHandler) CreateSupplier(c *gin.Context) {
	var req usecase.CreateSupplierRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	companyIDStr, _ := c.Get("company_id")
	companyID := uuid.Nil
	if companyIDStr != nil {
		if id, err := uuid.Parse(companyIDStr.(string)); err == nil {
			companyID = id
		}
	}

	supplier, err := h.usecase.CreateSupplier(req, companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Supplier berhasil ditambahkan", "data": supplier})
}

func (h *SupplierHandler) BulkCreateSuppliers(c *gin.Context) {
	var req []usecase.CreateSupplierRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	companyIDStr, _ := c.Get("company_id")
	companyID := uuid.Nil
	if companyIDStr != nil {
		if id, err := uuid.Parse(companyIDStr.(string)); err == nil {
			companyID = id
		}
	}

	suppliers, err := h.usecase.BulkCreateSuppliers(req, companyID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Data supplier berhasil diimpor",
		"count":   len(suppliers),
		"data":    suppliers,
	})
}

func (h *SupplierHandler) UpdateSupplier(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}
	var req usecase.CreateSupplierRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	supplier, err := h.usecase.UpdateSupplier(id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Supplier berhasil diperbarui", "data": supplier})
}

func (h *SupplierHandler) DeleteSupplier(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}
	if err := h.usecase.DeleteSupplier(id); err != nil {
		if errors.Is(err, domain.ErrSupplierInUse) {
			c.JSON(http.StatusConflict, gin.H{"error": "Supplier tidak dapat dihapus karena masih digunakan pada Purchase Order bahan baku"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Supplier berhasil dihapus"})
}

// ─── Supplier Receipt Handlers ────────────────────────────────────────────────

func (h *SupplierHandler) ListSupplierReceipts(c *gin.Context) {
	branchID := parseBranchID(c)
	receipts, err := h.usecase.ListSupplierReceipts(branchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": receipts})
}

func (h *SupplierHandler) GetSupplierReceipt(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}
	receipt, err := h.usecase.GetSupplierReceiptByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Penerimaan tidak ditemukan"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": receipt})
}

func (h *SupplierHandler) CreateSupplierReceipt(c *gin.Context) {
	var req usecase.CreateSupplierReceiptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	companyIDStr, _ := c.Get("company_id")
	companyID := uuid.Nil
	if companyIDStr != nil {
		if id, err := uuid.Parse(companyIDStr.(string)); err == nil {
			companyID = id
		}
	}

	receipt, err := h.usecase.CreateSupplierReceipt(req, companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Penerimaan supplier berhasil dibuat", "data": receipt})
}

func (h *SupplierHandler) UpdateSupplierReceipt(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}
	var req usecase.CreateSupplierReceiptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	receipt, err := h.usecase.UpdateSupplierReceipt(id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Penerimaan berhasil diperbarui", "data": receipt})
}

func (h *SupplierHandler) ReceiveSupplierReceipt(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}
	var req usecase.ReceiveSupplierReceiptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.usecase.ReceiveSupplierReceipt(id, req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Status penerimaan berhasil diperbarui dan stok telah disesuaikan"})
}

func (h *SupplierHandler) DeleteSupplierReceipt(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}
	if err := h.usecase.DeleteSupplierReceipt(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Penerimaan berhasil dihapus"})
}

// ─── Supplier Stock Handlers ──────────────────────────────────────────────────

func (h *SupplierHandler) GetSupplierStock(c *gin.Context) {
	branchID := parseBranchID(c)
	stocks, err := h.usecase.GetSupplierStock(branchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": stocks})
}
