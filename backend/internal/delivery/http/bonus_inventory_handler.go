package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"example.com/hris/backend/internal/domain"
	"example.com/hris/backend/internal/usecase"
)

type BonusInventoryHandler struct {
	usecase usecase.BonusInventoryUsecase
}

func NewBonusInventoryHandler(u usecase.BonusInventoryUsecase) *BonusInventoryHandler {
	return &BonusInventoryHandler{usecase: u}
}

func (h *BonusInventoryHandler) SetupRoutes(r *gin.RouterGroup) {
	bonusGroup := r.Group("/bonus-inventory")
	{
		bonusGroup.POST("", h.Create)
		bonusGroup.GET("/branch/:branch_id", h.GetByBranchID)
		bonusGroup.GET("/:id", h.GetByID)
		bonusGroup.PUT("/:id", h.Update)
		bonusGroup.DELETE("/:id", h.Delete)
		bonusGroup.POST("/:id/adjust", h.AdjustStock)
		bonusGroup.GET("/:id/logs", h.GetLogs)
	}
}

func (h *BonusInventoryHandler) Create(c *gin.Context) {
	var item domain.BonusInventory
	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.usecase.Create(c.Request.Context(), &item); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, item)
}

func (h *BonusInventoryHandler) GetByBranchID(c *gin.Context) {
	branchIDStr := c.Param("branch_id")

	if branchIDStr == "ALL_BRANCHES" {
		items, err := h.usecase.GetAll(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, items)
		return
	}

	branchID, err := uuid.Parse(branchIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid branch id"})
		return
	}

	items, err := h.usecase.GetByBranchID(c.Request.Context(), branchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, items)
}

func (h *BonusInventoryHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	item, err := h.usecase.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, item)
}

func (h *BonusInventoryHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var item domain.BonusInventory
	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	item.ID = id
	if err := h.usecase.Update(c.Request.Context(), &item); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "updated successfully"})
}

func (h *BonusInventoryHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.usecase.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "deleted successfully"})
}

type AdjustStockReq struct {
	Quantity int    `json:"quantity"`
	Notes    string `json:"notes"`
}

func (h *BonusInventoryHandler) AdjustStock(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req AdjustStockReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// In real app, get user from context
	user := "ADMIN"

	if err := h.usecase.AdjustStock(c.Request.Context(), id, req.Quantity, user, req.Notes); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "stock adjusted successfully"})
}

func (h *BonusInventoryHandler) GetLogs(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	logs, err := h.usecase.GetLogs(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, logs)
}
