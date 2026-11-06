package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"example.com/hris/backend/internal/usecase"
)

type BannerOrderHandler struct {
	bannerOrderUseCase usecase.BannerOrderUseCase
	employeeUseCase    usecase.EmployeeUsecase
}

func NewBannerOrderHandler(bannerOrderUseCase usecase.BannerOrderUseCase, employeeUseCase usecase.EmployeeUsecase) *BannerOrderHandler {
	return &BannerOrderHandler{
		bannerOrderUseCase: bannerOrderUseCase,
		employeeUseCase:    employeeUseCase,
	}
}

func (h *BannerOrderHandler) SetupRoutes(r *gin.RouterGroup) {
	bannerOrders := r.Group("/banner-orders")
	{
		bannerOrders.POST("", h.CreateOrder)
		bannerOrders.GET("", h.GetAllOrders)
		bannerOrders.PUT("/:id", h.UpdateOrder)
		bannerOrders.PATCH("/:id/status", h.UpdateStatus)
		bannerOrders.DELETE("/:id", h.DeleteOrder)
	}
}

func (h *BannerOrderHandler) CreateOrder(c *gin.Context) {
	var req domain.BannerOrder
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Auto assign company_id and employee_id from token
	if userIDStr, exists := c.Get("userID"); exists {
		userID := userIDStr.(uuid.UUID)
		employee, err := h.employeeUseCase.GetEmployeeByUserID(userID)
		if err == nil && employee != nil {
			if employee.CompanyID != nil && *employee.CompanyID != uuid.Nil {
				req.CompanyID = *employee.CompanyID
			}
			req.EmployeeID = employee.ID
		}
	}

	if req.EmployeeID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized or employee not found"})
		return
	}

	if err := h.bannerOrderUseCase.CreateOrder(&req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Banner order created successfully",
		"data":    req,
	})
}

func (h *BannerOrderHandler) GetAllOrders(c *gin.Context) {
	companyIDStr := c.Query("company_id")
	var orders []domain.BannerOrder
	var err error

	if companyIDStr != "" {
		companyID, parseErr := uuid.Parse(companyIDStr)
		if parseErr == nil {
			orders, err = h.bannerOrderUseCase.GetOrdersByCompanyID(companyID)
		} else {
			orders, err = h.bannerOrderUseCase.GetAllOrders()
		}
	} else {
		orders, err = h.bannerOrderUseCase.GetAllOrders()
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Success",
		"data":    orders,
	})
}

func (h *BannerOrderHandler) UpdateStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.bannerOrderUseCase.UpdateStatus(id, req.Status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Status updated successfully"})
}

func (h *BannerOrderHandler) UpdateOrder(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
		return
	}

	var req domain.BannerOrder
	if err := c.ShouldBindJSON(&req); err != nil {
		// If it's a UUID error, we might want to manually check some fields
	}
	req.ID = id

	// Manual sanitize for empty UUID strings in JSON
	if req.StoreID != nil && *req.StoreID == uuid.Nil {
		req.StoreID = nil
	}
	if req.DesignerID != nil && *req.DesignerID == uuid.Nil {
		req.DesignerID = nil
	}
	if req.InstallerID != nil && *req.InstallerID == uuid.Nil {
		req.InstallerID = nil
	}

	if err := h.bannerOrderUseCase.UpdateOrder(&req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Order updated successfully", "data": req})
}

func (h *BannerOrderHandler) DeleteOrder(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
		return
	}

	if err := h.bannerOrderUseCase.DeleteOrder(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Order deleted successfully"})
}
