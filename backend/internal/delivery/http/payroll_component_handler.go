package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"example.com/hris/backend/internal/usecase"
)

type PayrollComponentHandler struct {
	useCase usecase.PayrollComponentUseCase
}

func NewPayrollComponentHandler(useCase usecase.PayrollComponentUseCase) *PayrollComponentHandler {
	return &PayrollComponentHandler{useCase: useCase}
}

func (h *PayrollComponentHandler) SetupRoutes(r *gin.RouterGroup) {
	components := r.Group("/payroll-components")
	{
		components.POST("", h.Create)
		components.GET("", h.GetAll)
		components.GET("/:id", h.GetByID)
		components.PUT("/:id", h.Update)
		components.DELETE("/:id", h.Delete)
	}
}

func (h *PayrollComponentHandler) Create(c *gin.Context) {
	var req usecase.CreatePayrollComponentRequest
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

	if err := h.useCase.CreateComponent(&req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Payroll component created successfully"})
}

func (h *PayrollComponentHandler) GetAll(c *gin.Context) {
	components, err := h.useCase.GetComponents()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": components})
}

func (h *PayrollComponentHandler) GetByID(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	component, err := h.useCase.GetComponentByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if component == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Payroll component not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": component})
}

func (h *PayrollComponentHandler) Update(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	var req usecase.UpdatePayrollComponentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.useCase.UpdateComponent(id, &req); err != nil {
		if err.Error() == "payroll component not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Payroll component updated successfully"})
}

func (h *PayrollComponentHandler) Delete(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	if err := h.useCase.DeleteComponent(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Payroll component deleted successfully"})
}
