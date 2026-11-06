package http

import (
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"example.com/hris/backend/internal/usecase"
	"example.com/hris/backend/pkg/utils"
)

type MarketplaceInventoryHandler struct {
	usecase usecase.MarketplaceInventoryUsecase
}

func NewMarketplaceInventoryHandler(u usecase.MarketplaceInventoryUsecase) *MarketplaceInventoryHandler {
	return &MarketplaceInventoryHandler{u}
}

func (h *MarketplaceInventoryHandler) RegisterRoutes(r *gin.RouterGroup) {
	inventory := r.Group("/marketplace/inventories")
	inventory.Use(RoleMiddleware("superadmin", "manager", "hr_admin", "admin_gudang", "supervisor_sales", "admin_marketplace"))
	{
		inventory.GET("", h.ListInventories)
		inventory.GET("/:id", h.GetInventory)
		inventory.POST("", h.CreateInventory)
		inventory.PUT("/:id", h.UpdateInventory)
		inventory.DELETE("/:id", h.DeleteInventory)

		// Movements
		inventory.POST("/:id/in", h.StockIn)
		inventory.POST("/:id/out", h.StockOut)
		inventory.POST("/:id/opname", h.DoOpname)

		// History
		inventory.GET("/:id/logs", h.GetLogs)
		inventory.GET("/:id/opnames", h.GetOpnames)
	}
}

func (h *MarketplaceInventoryHandler) ListInventories(c *gin.Context) {
	branchID := parseBranchID(c)
	inventories, err := h.usecase.GetAllInventories(branchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, inventories)
}

func (h *MarketplaceInventoryHandler) GetInventory(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}
	inventory, err := h.usecase.GetInventoryByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Inventaris tidak ditemukan"})
		return
	}
	c.JSON(http.StatusOK, inventory)
}

func (h *MarketplaceInventoryHandler) CreateInventory(c *gin.Context) {
	var req struct {
		Name     string `form:"name" binding:"required"`
		SKU      string `form:"sku"`
		Unit     string `form:"unit" binding:"required"`
		MinStock int    `form:"min_stock"`
		Notes    string `form:"notes"`
	}
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	imageURL, err := h.handleImageUpload(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal upload gambar: " + err.Error()})
		return
	}
	actorID, actorName := getActorInfo(c)
	inventory := &domain.MarketplaceInventory{
		BranchID:  parseBranchID(c),
		Name:      req.Name,
		SKU:       req.SKU,
		Unit:      req.Unit,
		MinStock:  req.MinStock,
		Notes:     req.Notes,
		ImageURL:  imageURL,
		CreatedBy: actorName,
	}
	if err := h.usecase.CreateInventory(inventory, actorID, c.ClientIP()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, inventory)
}

func (h *MarketplaceInventoryHandler) UpdateInventory(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}
	var req struct {
		Name     string `form:"name" binding:"required"`
		SKU      string `form:"sku"`
		Unit     string `form:"unit" binding:"required"`
		MinStock int    `form:"min_stock"`
		Notes    string `form:"notes"`
	}
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	imageURL, err := h.handleImageUpload(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal upload gambar: " + err.Error()})
		return
	}

	var updateData domain.MarketplaceInventory
	updateData.Name = req.Name
	updateData.SKU = req.SKU
	updateData.Unit = req.Unit
	updateData.MinStock = req.MinStock
	updateData.Notes = req.Notes
	if imageURL != "" {
		updateData.ImageURL = imageURL
	}
	updateData.ID = id
	actorID, _ := getActorInfo(c)
	if err := h.usecase.UpdateInventory(&updateData, actorID, c.ClientIP()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Inventaris berhasil diperbarui"})
}

func (h *MarketplaceInventoryHandler) DeleteInventory(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}
	actorID, actorName := getActorInfo(c)
	if err := h.usecase.DeleteInventory(id, actorID, actorName, c.ClientIP()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Inventaris berhasil dihapus"})
}

func (h *MarketplaceInventoryHandler) StockIn(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}
	var req struct {
		Quantity  *int   `json:"quantity" binding:"required"`
		Reference string `json:"reference"`
		Notes     string `json:"notes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	actorID, actorName := getActorInfo(c)
	if err := h.usecase.StockIn(id, *req.Quantity, req.Reference, req.Notes, actorName, actorID, c.ClientIP()); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Stok berhasil ditambahkan"})
}

func (h *MarketplaceInventoryHandler) StockOut(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}
	var req struct {
		Quantity  *int   `json:"quantity" binding:"required"`
		Reference string `json:"reference"`
		Notes     string `json:"notes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	actorID, actorName := getActorInfo(c)
	if err := h.usecase.StockOut(id, *req.Quantity, req.Reference, req.Notes, actorName, actorID, c.ClientIP()); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Stok berhasil dikurangi"})
}

func (h *MarketplaceInventoryHandler) DoOpname(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}
	var req struct {
		Quantity *int   `json:"quantity" binding:"required"`
		Notes    string `json:"notes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	actorID, actorName := getActorInfo(c)
	if err := h.usecase.PerformOpname(id, *req.Quantity, req.Notes, actorName, actorID, c.ClientIP()); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Opname berhasil disimpan"})
}

func (h *MarketplaceInventoryHandler) GetLogs(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	var from, to *time.Time
	if f := c.Query("from"); f != "" {
		t, err := time.Parse("2006-01-02", f)
		if err == nil {
			from = &t
		}
	}
	if t := c.Query("to"); t != "" {
		parsed, err := time.Parse("2006-01-02", t)
		if err == nil {
			to = &parsed
		}
	}

	logs, err := h.usecase.GetLogs(&id, from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, logs)
}

func (h *MarketplaceInventoryHandler) GetOpnames(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	var from, to *time.Time
	if f := c.Query("from"); f != "" {
		t, err := time.Parse("2006-01-02", f)
		if err == nil {
			from = &t
		}
	}
	if t := c.Query("to"); t != "" {
		parsed, err := time.Parse("2006-01-02", t)
		if err == nil {
			to = &parsed
		}
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	offset := (page - 1) * limit

	opnames, total, err := h.usecase.GetOpnames(&id, from, to, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data":  opnames,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

func (h *MarketplaceInventoryHandler) handleImageUpload(c *gin.Context) (string, error) {
	file, err := c.FormFile("image")
	if err != nil {
		if err == http.ErrMissingFile {
			return "", nil
		}
		return "", err
	}
	filename := time.Now().Format("20060102150405") + "_" + file.Filename
	uploadDir := "uploads/inventories"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return "", err
	}
	savePath := filepath.Join(uploadDir, filename)
	if err := utils.CompressAndSaveImage(file, savePath, 1000); err != nil {
		return "", err
	}
	return "/uploads/inventories/" + filename, nil
}
