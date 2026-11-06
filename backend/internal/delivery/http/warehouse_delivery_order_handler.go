package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"example.com/hris/backend/internal/usecase"
)

type WarehouseDeliveryOrderHandler struct {
	usecase usecase.WarehouseDeliveryOrderUsecase
}

func NewWarehouseDeliveryOrderHandler(u usecase.WarehouseDeliveryOrderUsecase) *WarehouseDeliveryOrderHandler {
	return &WarehouseDeliveryOrderHandler{u}
}

func (h *WarehouseDeliveryOrderHandler) RegisterRoutes(router *gin.RouterGroup) {
	// DO Regular (terhubung ke SJ/DeliveryBatch)
	router.POST("/admin/warehouse-delivery-orders/get-or-create", h.CreateOrGetDO)
	router.PATCH("/admin/warehouse-delivery-orders", h.UpdateDO)
	router.POST("/admin/warehouse-delivery-orders/:id/dispatch", h.DispatchDO)
	router.POST("/admin/warehouse-delivery-orders/:id/finalize", h.FinalizeDO)
	router.GET("/admin/warehouse-delivery-orders/batch/:batch_id", h.GetByBatchID)
	router.GET("/admin/warehouse-delivery-orders/id/:id", h.GetByID)

	// DO Motoris (standalone, tanpa SJ)
	router.POST("/admin/warehouse-delivery-orders/motoris", h.CreateMotorissDO)
	router.GET("/admin/warehouse-delivery-orders/motoris/list", h.GetMotorisDOs)
}

func (h *WarehouseDeliveryOrderHandler) CreateOrGetDO(c *gin.Context) {
	var req struct {
		DeliveryBatchID string `json:"delivery_batch_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var wdo *domain.WarehouseDeliveryOrder
	var err error

	batchUUID, parseErr := uuid.Parse(req.DeliveryBatchID)
	if parseErr == nil {
		wdo, err = h.usecase.CreateOrGetDO(batchUUID)
	} else {
		wdo, err = h.usecase.CreateOrGetDelegationDO(req.DeliveryBatchID)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": wdo})
}

// CreateMotorissDO — dipanggil dari frontend ketika klik "Buat DO Motoris"
// di menu Perpindahan Barang. Tidak membuat SJ/DeliveryBatch.
func (h *WarehouseDeliveryOrderHandler) CreateMotorissDO(c *gin.Context) {
	var req struct {
		SalesTransferIDs []uuid.UUID `json:"sales_transfer_ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	wdo, err := h.usecase.CreateMotorissDO(req.SalesTransferIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": wdo, "message": "DO Motoris berhasil dibuat"})
}

// GetMotorisDOs — mengembalikan list DO Motoris untuk halaman Kelola DO Spares
func (h *WarehouseDeliveryOrderHandler) GetMotorisDOs(c *gin.Context) {
	var branchID *uuid.UUID
	if branchIDStr := c.Query("branch_id"); branchIDStr != "" {
		parsed, err := uuid.Parse(branchIDStr)
		if err == nil {
			branchID = &parsed
		}
	}

	wdos, err := h.usecase.GetMotorisDOs(branchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": wdos})
}

func (h *WarehouseDeliveryOrderHandler) UpdateDO(c *gin.Context) {
	var req usecase.UpdateWarehouseDORequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Resolve user info from context
	var userID uuid.UUID
	var userName string
	if val, exists := c.Get("userID"); exists {
		userID = val.(uuid.UUID)
	}
	if val, exists := c.Get("userName"); exists {
		userName = val.(string)
	}

	wdo, err := h.usecase.UpdateDO(req, userID, userName, c.ClientIP())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": wdo})
}

func (h *WarehouseDeliveryOrderHandler) GetByBatchID(c *gin.Context) {
	batchID, err := uuid.Parse(c.Param("batch_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	wdo, err := h.usecase.GetByBatchID(batchID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "DO gudang tidak ditemukan untuk Surat Jalan ini"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": wdo})
}

func (h *WarehouseDeliveryOrderHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	wdo, err := h.usecase.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "DO gudang tidak ditemukan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": wdo})
}

func (h *WarehouseDeliveryOrderHandler) DispatchDO(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	// Resolve user info from context
	var userID uuid.UUID
	var userName string
	if val, exists := c.Get("userID"); exists {
		userID = val.(uuid.UUID)
	}
	if val, exists := c.Get("userName"); exists {
		userName = val.(string)
	}

	err = h.usecase.DispatchDO(id, userID, userName, c.ClientIP())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "DO berhasil di-dispatch, stok dipindahkan ke sales"})
}

func (h *WarehouseDeliveryOrderHandler) FinalizeDO(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	// Resolve user info from context
	var userID uuid.UUID
	var userName string
	if val, exists := c.Get("userID"); exists {
		userID = val.(uuid.UUID)
	}
	if val, exists := c.Get("userName"); exists {
		userName = val.(string)
	}

	err = h.usecase.FinalizeDO(id, userID, userName, c.ClientIP())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "DO berhasil difinalisasi"})
}
