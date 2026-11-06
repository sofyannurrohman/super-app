package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"example.com/hris/backend/internal/usecase"
)

type DepartmentHandler struct {
	useCase usecase.DepartmentUseCase
}

func NewDepartmentHandler(useCase usecase.DepartmentUseCase) *DepartmentHandler {
	return &DepartmentHandler{useCase: useCase}
}

func (h *DepartmentHandler) SetupRoutes(r *gin.RouterGroup) {
	departments := r.Group("/departments")
	{
		departments.POST("", h.Create)
		departments.GET("", h.GetAll)
		departments.GET("/:id", h.GetByID)
		departments.PUT("/:id", h.Update)
		departments.DELETE("/:id", h.Delete)
	}
}

func (h *DepartmentHandler) Create(c *gin.Context) {
	var req usecase.CreateDepartmentRequest
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

	if err := h.useCase.CreateDepartment(&req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Department created successfully"})
}

func (h *DepartmentHandler) GetAll(c *gin.Context) {
	departments, err := h.useCase.GetDepartments()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": departments})
}

func (h *DepartmentHandler) GetByID(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	department, err := h.useCase.GetDepartmentByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if department == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Department not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": department})
}

func (h *DepartmentHandler) Update(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	var req usecase.UpdateDepartmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.useCase.UpdateDepartment(id, &req); err != nil {
		if err.Error() == "department not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Department updated successfully"})
}

func (h *DepartmentHandler) Delete(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	if err := h.useCase.DeleteDepartment(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Department deleted successfully"})
}
