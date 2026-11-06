package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"example.com/hris/backend/internal/usecase"
)

type EmployeeWarningHandler struct {
	usecase *usecase.EmployeeWarningUsecase
}

func NewEmployeeWarningHandler(uc *usecase.EmployeeWarningUsecase) *EmployeeWarningHandler {
	return &EmployeeWarningHandler{
		usecase: uc,
	}
}

func (h *EmployeeWarningHandler) RegisterRoutes(r *gin.RouterGroup) {
	hr := r.Group("/hr")
	{
		hr.GET("/problematic-employees", h.GetProblematicEmployees)
		hr.GET("/warnings", h.GetAllWarnings)
		hr.POST("/warnings", h.SubmitWarning)
		hr.PUT("/warnings/:id", h.UpdateWarning)
		hr.PUT("/warnings/:id/toggle", h.ToggleWarningActive)
		hr.DELETE("/warnings/:id", h.DeleteWarning)
	}
}

func (h *EmployeeWarningHandler) GetProblematicEmployees(c *gin.Context) {
	employees, err := h.usecase.GetProblematicEmployees()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "success",
		"data":    employees,
	})
}

func (h *EmployeeWarningHandler) SubmitWarning(c *gin.Context) {
	var req struct {
		EmployeeID string `json:"employee_id" binding:"required"`
		ActionType string `json:"action_type" binding:"required"` // teguran / sp1
		Notes      string `json:"notes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Ambil userID dari JWT middleware (biasanya di-set sebagai "user_id")
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	handledByUserID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid User ID"})
		return
	}

	err = h.usecase.SubmitWarning(req.EmployeeID, req.ActionType, req.Notes, handledByUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Warning submitted successfully",
	})
}

func (h *EmployeeWarningHandler) GetAllWarnings(c *gin.Context) {
	warnings, err := h.usecase.GetAllWarnings()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "success",
		"data":    warnings,
	})
}

func (h *EmployeeWarningHandler) UpdateWarning(c *gin.Context) {
	idParam := c.Param("id")
	warningID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid warning ID"})
		return
	}

	var req struct {
		ActionType string `json:"action_type" binding:"required"`
		Notes      string `json:"notes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.usecase.UpdateWarning(warningID, req.ActionType, req.Notes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Warning updated successfully",
	})
}

func (h *EmployeeWarningHandler) DeleteWarning(c *gin.Context) {
	idParam := c.Param("id")
	warningID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid warning ID"})
		return
	}

	err = h.usecase.DeleteWarning(warningID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Warning deleted successfully",
	})
}

func (h *EmployeeWarningHandler) ToggleWarningActive(c *gin.Context) {
	idParam := c.Param("id")
	warningID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid warning ID"})
		return
	}

	err = h.usecase.ToggleWarningActive(warningID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to toggle warning status", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Warning status toggled successfully",
	})
}
