package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"example.com/hris/backend/internal/usecase"
)

type SalesBonusHandler struct {
	usecase usecase.SalesBonusUsecase
}

func NewSalesBonusHandler(usecase usecase.SalesBonusUsecase) *SalesBonusHandler {
	return &SalesBonusHandler{
		usecase: usecase,
	}
}

func (h *SalesBonusHandler) SetupRoutes(router *gin.RouterGroup) {
	router.POST("/sales-bonus-requests", h.CreateRequest)
	router.GET("/sales-bonus-requests/pending", h.GetPendingRequests)
	router.PATCH("/sales-bonus-requests/:id/approve", h.ApproveRequest)
	router.PATCH("/sales-bonus-requests/:id/reject", h.RejectRequest)
	router.GET("/sales-bonus-stock/:employeeId", h.GetSalesBonusStock)
}

type CreateSalesBonusRequestPayload struct {
	EmployeeID       uuid.UUID `json:"employee_id" binding:"required"`
	BonusInventoryID uuid.UUID `json:"bonus_inventory_id" binding:"required"`
	Quantity         int       `json:"quantity" binding:"required,min=1"`
	Notes            string    `json:"notes"`
}

func (h *SalesBonusHandler) CreateRequest(c *gin.Context) {
	var payload CreateSalesBonusRequestPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.usecase.CreateRequest(c.Request.Context(), payload.EmployeeID, payload.BonusInventoryID, payload.Quantity, payload.Notes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Request created successfully"})
}

func (h *SalesBonusHandler) GetPendingRequests(c *gin.Context) {
	branchID := c.Query("branch_id")
	reqs, err := h.usecase.GetPendingRequests(c.Request.Context(), branchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": reqs})
}

func (h *SalesBonusHandler) ApproveRequest(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request ID"})
		return
	}

	err = h.usecase.ApproveRequest(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Request approved successfully"})
}

func (h *SalesBonusHandler) RejectRequest(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request ID"})
		return
	}

	err = h.usecase.RejectRequest(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Request rejected successfully"})
}

func (h *SalesBonusHandler) GetSalesBonusStock(c *gin.Context) {
	employeeIdStr := c.Param("employeeId")
	employeeId, err := uuid.Parse(employeeIdStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid employee ID"})
		return
	}

	stocks, err := h.usecase.GetSalesBonusStock(c.Request.Context(), employeeId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": stocks})
}
