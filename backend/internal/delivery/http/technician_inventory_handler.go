package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"example.com/hris/backend/internal/usecase"
)

type TechnicianInventoryHandler struct {
	usecase usecase.TechnicianInventoryUsecase
}

func NewTechnicianInventoryHandler(usecase usecase.TechnicianInventoryUsecase) *TechnicianInventoryHandler {
	return &TechnicianInventoryHandler{usecase: usecase}
}

func (h *TechnicianInventoryHandler) RegisterRoutes(r *gin.RouterGroup) {
	group := r.Group("/factory/:id/technician-items")
	{
		group.GET("", h.ListItems)
		group.POST("", h.CreateItem)
		group.GET("/logs", h.ListLogs)
		group.POST("/usage", h.RecordUsage)
		group.POST("/batch-opname", h.BatchOpname)
		group.GET("/:itemId", h.GetItem)
		group.PUT("/:itemId", h.UpdateItem)
		group.DELETE("/:itemId", h.DeleteItem)
		group.POST("/:itemId/opname", h.OpnameItem)
	}
}

func (h *TechnicianInventoryHandler) ListItems(c *gin.Context) {
	val, exists := c.Get("companyID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: missing company context"})
		return
	}
	companyID := val.(uuid.UUID)

	factoryIDStr := c.Param("id")
	var factoryID uuid.UUID
	if factoryIDStr != "" && factoryIDStr != "all" {
		var err error
		factoryID, err = uuid.Parse(factoryIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid factory ID"})
			return
		}
	}

	category := c.Query("category")
	lowStockOnly := c.Query("low_stock") == "true"
	search := c.Query("search")

	items, err := h.usecase.ListItems(companyID, factoryID, category, lowStockOnly, search)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": items})
}

func (h *TechnicianInventoryHandler) GetItem(c *gin.Context) {
	itemIDStr := c.Param("itemId")
	itemID, err := uuid.Parse(itemIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid item ID"})
		return
	}

	item, err := h.usecase.GetItemByID(itemID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Item not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": item})
}

func (h *TechnicianInventoryHandler) CreateItem(c *gin.Context) {
	val, exists := c.Get("companyID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: missing company context"})
		return
	}
	companyID := val.(uuid.UUID)

	factoryIDStr := c.Param("id")
	factoryID, err := uuid.Parse(factoryIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid factory ID"})
		return
	}

	var req usecase.CreateTechnicianItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.FactoryID = factoryID

	item, err := h.usecase.CreateItem(companyID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": item, "message": "Barang teknisi berhasil ditambahkan"})
}

func (h *TechnicianInventoryHandler) UpdateItem(c *gin.Context) {
	itemIDStr := c.Param("itemId")
	itemID, err := uuid.Parse(itemIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid item ID"})
		return
	}

	var req usecase.UpdateTechnicianItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	item, err := h.usecase.UpdateItem(itemID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": item, "message": "Barang teknisi berhasil diperbarui"})
}

func (h *TechnicianInventoryHandler) DeleteItem(c *gin.Context) {
	itemIDStr := c.Param("itemId")
	itemID, err := uuid.Parse(itemIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid item ID"})
		return
	}

	if err := h.usecase.DeleteItem(itemID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Barang teknisi berhasil dihapus"})
}

func (h *TechnicianInventoryHandler) OpnameItem(c *gin.Context) {
	itemIDStr := c.Param("itemId")
	itemID, err := uuid.Parse(itemIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid item ID"})
		return
	}

	var req usecase.OpnameTechnicianStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var userID *uuid.UUID
	userName := "Admin Pabrik"
	if uidVal, exists := c.Get("userID"); exists {
		parsedUID := uidVal.(uuid.UUID)
		userID = &parsedUID
	}
	if nameVal, exists := c.Get("userName"); exists && nameVal.(string) != "" {
		userName = nameVal.(string)
	}

	item, err := h.usecase.OpnameStock(itemID, req, userID, userName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": item, "message": "Stock Opname barang teknisi berhasil disimpan"})
}

func (h *TechnicianInventoryHandler) BatchOpname(c *gin.Context) {
	val, exists := c.Get("companyID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: missing company context"})
		return
	}
	companyID := val.(uuid.UUID)

	factoryIDStr := c.Param("id")
	factoryID, err := uuid.Parse(factoryIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid factory ID"})
		return
	}

	var req usecase.BatchOpnameTechnicianStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var userID *uuid.UUID
	userName := "Admin Pabrik"
	if uidVal, exists := c.Get("userID"); exists {
		parsedUID := uidVal.(uuid.UUID)
		userID = &parsedUID
	}
	if nameVal, exists := c.Get("userName"); exists && nameVal.(string) != "" {
		userName = nameVal.(string)
	}

	if err := h.usecase.BatchOpnameStock(factoryID, companyID, req, userID, userName); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Batch Stock Opname berhasil disimpan"})
}

func (h *TechnicianInventoryHandler) RecordUsage(c *gin.Context) {
	val, exists := c.Get("companyID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: missing company context"})
		return
	}
	companyID := val.(uuid.UUID)

	factoryIDStr := c.Param("id")
	factoryID, err := uuid.Parse(factoryIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid factory ID"})
		return
	}

	var req usecase.RecordTechnicianUsageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var userID *uuid.UUID
	userName := "Admin Pabrik"
	if uidVal, exists := c.Get("userID"); exists {
		parsedUID := uidVal.(uuid.UUID)
		userID = &parsedUID
	}
	if nameVal, exists := c.Get("userName"); exists && nameVal.(string) != "" {
		userName = nameVal.(string)
	}

	if err := h.usecase.RecordUsage(factoryID, companyID, req, userID, userName); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Pemakaian barang teknisi berhasil dicatat"})
}

func (h *TechnicianInventoryHandler) ListLogs(c *gin.Context) {
	val, exists := c.Get("companyID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: missing company context"})
		return
	}
	companyID := val.(uuid.UUID)

	factoryIDStr := c.Param("id")
	var factoryID uuid.UUID
	if factoryIDStr != "" && factoryIDStr != "all" {
		var err error
		factoryID, err = uuid.Parse(factoryIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid factory ID"})
			return
		}
	}

	var itemID *uuid.UUID
	if itemIDStr := c.Query("item_id"); itemIDStr != "" {
		if parsed, err := uuid.Parse(itemIDStr); err == nil {
			itemID = &parsed
		}
	}

	logType := c.Query("type")

	logs, err := h.usecase.ListStockLogs(companyID, factoryID, itemID, logType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": logs})
}
