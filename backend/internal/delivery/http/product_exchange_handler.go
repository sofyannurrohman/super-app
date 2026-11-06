package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"example.com/hris/backend/internal/usecase"
)

type ProductExchangeHandler struct {
	exchangeUsecase usecase.ProductExchangeUsecase
	employeeUseCase usecase.EmployeeUsecase
}

func NewProductExchangeHandler(eu usecase.ProductExchangeUsecase, empUsecase usecase.EmployeeUsecase) *ProductExchangeHandler {
	return &ProductExchangeHandler{eu, empUsecase}
}

func (h *ProductExchangeHandler) RegisterRoutes(router *gin.RouterGroup) {
	router.GET("/admin/product-exchanges", h.GetByBranch)
	router.GET("/admin/product-exchanges/:id", h.GetByID)
	router.POST("/admin/product-exchanges", h.CreateExchange)
	router.PATCH("/admin/product-exchanges/:id/approve", h.ApproveExchange)
}

func (h *ProductExchangeHandler) CreateExchange(c *gin.Context) {
	var req usecase.CreateExchangeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Resolve EmployeeID from JWT token
	if userIDStr, exists := c.Get("userID"); exists {
		userID := userIDStr.(uuid.UUID)
		emp, err := h.employeeUseCase.GetEmployeeByUserID(userID)
		if err == nil && emp != nil {
			req.EmployeeID = emp.ID
		}
	}

	pe, err := h.exchangeUsecase.CreateExchange(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": pe})
}

func (h *ProductExchangeHandler) ApproveExchange(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	var approvedByID uuid.UUID
	if userIDStr, exists := c.Get("userID"); exists {
		userID := userIDStr.(uuid.UUID)
		emp, err := h.employeeUseCase.GetEmployeeByUserID(userID)
		if err == nil && emp != nil {
			approvedByID = emp.ID
		}
	}

	if err := h.exchangeUsecase.ApproveExchange(id, approvedByID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Klaim garansi/tukar barang disetujui"})
}

func (h *ProductExchangeHandler) GetByBranch(c *gin.Context) {
	branchID, _ := uuid.Parse(c.Query("branch_id"))
	status := c.Query("status")
	exchanges, err := h.exchangeUsecase.GetByBranch(branchID, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": exchanges})
}

func (h *ProductExchangeHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}
	pe, err := h.exchangeUsecase.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Klaim garansi/tukar barang tidak ditemukan"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": pe})
}
