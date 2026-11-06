package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"example.com/hris/backend/internal/usecase"
)

type SalesTesterRequestHandler struct {
	usecase         usecase.SalesTesterRequestUsecase
	employeeUsecase usecase.EmployeeUsecase
}

func NewSalesTesterRequestHandler(usecase usecase.SalesTesterRequestUsecase, employeeUsecase usecase.EmployeeUsecase) *SalesTesterRequestHandler {
	return &SalesTesterRequestHandler{usecase, employeeUsecase}
}

func (h *SalesTesterRequestHandler) RegisterRoutes(router *gin.RouterGroup) {
	router.POST("/admin/sales-tester-requests", h.CreateRequest)
	router.GET("/admin/sales-tester-requests", h.ListRequests)
	router.GET("/admin/sales-tester-requests/employee", h.GetEmployeeRequests)
	router.GET("/admin/sales-tester-requests/:id", h.GetRequestByID)
	router.PATCH("/admin/sales-tester-requests/:id/approve", h.ApproveRequest)
	router.PATCH("/admin/sales-tester-requests/:id/reject", h.RejectRequest)
	router.PATCH("/admin/sales-tester-requests/:id/issue-direct", h.IssueDirectly)
	router.DELETE("/admin/sales-tester-requests/:id", h.DeleteRequest)
}

type CreateTesterRequest struct {
	BranchID  uuid.UUID  `json:"branch_id" binding:"required"`
	StoreID   *uuid.UUID `json:"store_id"`
	ProductID uuid.UUID  `json:"product_id" binding:"required"`
	Quantity  int        `json:"quantity" binding:"required,min=1"`
	Notes     string     `json:"notes"`
}

func (h *SalesTesterRequestHandler) CreateRequest(c *gin.Context) {
	var body CreateTesterRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var employeeID uuid.UUID
	if userIDStr, exists := c.Get("userID"); exists {
		userID := userIDStr.(uuid.UUID)
		emp, err := h.employeeUsecase.GetEmployeeByUserID(userID)
		if err == nil && emp != nil {
			employeeID = emp.ID
		}
	}

	if employeeID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized employee"})
		return
	}

	req := &domain.SalesTesterRequest{
		BranchID:   body.BranchID,
		EmployeeID: employeeID,
		StoreID:    body.StoreID,
		ProductID:  body.ProductID,
		Quantity:   body.Quantity,
		Notes:      body.Notes,
	}

	if err := h.usecase.CreateRequest(req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": req})
}

func (h *SalesTesterRequestHandler) ListRequests(c *gin.Context) {
	var branchIDPtr *uuid.UUID
	if branchIDStr := c.Query("branch_id"); branchIDStr != "" {
		if bid, err := uuid.Parse(branchIDStr); err == nil {
			branchIDPtr = &bid
		}
	}
	status := c.Query("status")

	requests, err := h.usecase.ListRequests(branchIDPtr, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": requests})
}

func (h *SalesTesterRequestHandler) GetEmployeeRequests(c *gin.Context) {
	var employeeID uuid.UUID
	if userIDStr, exists := c.Get("userID"); exists {
		userID := userIDStr.(uuid.UUID)
		emp, err := h.employeeUsecase.GetEmployeeByUserID(userID)
		if err == nil && emp != nil {
			employeeID = emp.ID
		}
	}

	if employeeID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized employee"})
		return
	}

	requests, err := h.usecase.GetEmployeeRequests(employeeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": requests})
}

func (h *SalesTesterRequestHandler) GetRequestByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	req, err := h.usecase.GetRequestByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": req})
}

func (h *SalesTesterRequestHandler) ApproveRequest(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	if err := h.usecase.ApproveRequest(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Request approved successfully"})
}

type RejectTesterRequest struct {
	Notes string `json:"notes"`
}

func (h *SalesTesterRequestHandler) RejectRequest(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var body RejectTesterRequest
	_ = c.ShouldBindJSON(&body)

	if err := h.usecase.RejectRequest(id, body.Notes); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Request rejected successfully"})
}

func (h *SalesTesterRequestHandler) IssueDirectly(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	if err := h.usecase.IssueDirectly(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Tester issued directly and stock transferred"})
}

func (h *SalesTesterRequestHandler) DeleteRequest(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	if err := h.usecase.DeleteRequest(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Request deleted successfully"})
}
