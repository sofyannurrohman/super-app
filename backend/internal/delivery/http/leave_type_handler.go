package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"example.com/hris/backend/internal/usecase"
)

type LeaveTypeHandler struct {
	useCase usecase.LeaveTypeUseCase
}

func NewLeaveTypeHandler(useCase usecase.LeaveTypeUseCase) *LeaveTypeHandler {
	return &LeaveTypeHandler{useCase: useCase}
}

func (h *LeaveTypeHandler) SetupRoutes(r *gin.RouterGroup) {
	leaveTypes := r.Group("/leave-types")
	{
		leaveTypes.POST("", h.Create)
		leaveTypes.GET("", h.GetAll)
		leaveTypes.GET("/:id", h.GetByID)
		leaveTypes.PUT("/:id", h.Update)
		leaveTypes.DELETE("/:id", h.Delete)
	}
}

func (h *LeaveTypeHandler) Create(c *gin.Context) {
	var req usecase.CreateLeaveTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Automatically use company ID from JWT claims
	if companyID, exists := c.Get("companyID"); exists {
		cid := companyID.(uuid.UUID)
		if cid != uuid.Nil {
			req.CompanyID = &cid
		}
	}

	if err := h.useCase.CreateLeaveType(&req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Leave type created successfully"})
}

func (h *LeaveTypeHandler) GetAll(c *gin.Context) {
	leaveTypes, err := h.useCase.GetLeaveTypes()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": leaveTypes})
}

func (h *LeaveTypeHandler) GetByID(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	leaveType, err := h.useCase.GetLeaveTypeByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if leaveType == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Leave type not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": leaveType})
}

func (h *LeaveTypeHandler) Update(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	var req usecase.UpdateLeaveTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.useCase.UpdateLeaveType(id, &req); err != nil {
		if err.Error() == "leave type not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Leave type updated successfully"})
}

func (h *LeaveTypeHandler) Delete(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	if err := h.useCase.DeleteLeaveType(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Leave type deleted successfully"})
}
