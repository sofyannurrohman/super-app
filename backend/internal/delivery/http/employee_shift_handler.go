package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"example.com/hris/backend/internal/usecase"
)

type EmployeeShiftHandler struct {
	useCase usecase.EmployeeShiftUseCase
}

func NewEmployeeShiftHandler(useCase usecase.EmployeeShiftUseCase) *EmployeeShiftHandler {
	return &EmployeeShiftHandler{useCase: useCase}
}

func (h *EmployeeShiftHandler) SetupRoutes(r *gin.RouterGroup) {
	empShifts := r.Group("/employee-shifts")
	{
		empShifts.GET("", h.GetAll)
		empShifts.POST("", h.Assign)
		empShifts.GET("/employee/:employeeId", h.GetByEmployee)
		empShifts.PUT("/:id", h.Update)
		empShifts.DELETE("/:id", h.Delete)
	}
}

func (h *EmployeeShiftHandler) Assign(c *gin.Context) {
	var req usecase.AssignShiftRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.useCase.AssignShift(&req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Shift assigned successfully"})
}

func (h *EmployeeShiftHandler) GetAll(c *gin.Context) {
	assignments, err := h.useCase.GetAllAssignments()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": assignments})
}

func (h *EmployeeShiftHandler) GetByEmployee(c *gin.Context) {
	idParam := c.Param("employeeId")
	employeeID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid employee ID"})
		return
	}

	assignments, err := h.useCase.GetAssignmentsByEmployee(employeeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": assignments})
}

func (h *EmployeeShiftHandler) Update(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var req usecase.UpdateShiftAssignmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.useCase.UpdateAssignment(id, &req); err != nil {
		if err.Error() == "shift assignment not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Shift assignment updated"})
}

func (h *EmployeeShiftHandler) Delete(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	if err := h.useCase.DeleteAssignment(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Shift assignment deleted"})
}
