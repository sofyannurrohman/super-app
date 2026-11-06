package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"example.com/hris/backend/internal/usecase"
	"example.com/hris/backend/pkg/utils"
)

type PayrollConfigHandler struct {
	useCase usecase.PayrollConfigUseCase
}

func NewPayrollConfigHandler(useCase usecase.PayrollConfigUseCase) *PayrollConfigHandler {
	return &PayrollConfigHandler{useCase}
}

func (h *PayrollConfigHandler) SetupRoutes(router *gin.RouterGroup) {
	config := router.Group("/payroll/config")
	// Only HR Admins and SuperAdmins can manage global payroll settings
	config.Use(RoleMiddleware(string(domain.RoleSuperAdmin), string(domain.RoleHRAdmin)))
	{
		config.GET("", h.GetConfig)
		config.PUT("", h.UpdateConfig)
	}
}

func (h *PayrollConfigHandler) GetConfig(c *gin.Context) {
	companyIDVal, exists := c.Get("companyID")
	var companyID uuid.UUID
	if exists {
		if cid, ok := companyIDVal.(uuid.UUID); ok {
			companyID = cid
		}
	} else {
		companyID = uuid.Nil
	}

	result, err := h.useCase.GetConfig(companyID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get payroll config: "+err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Payroll config fetched successfully", result)
}

func (h *PayrollConfigHandler) UpdateConfig(c *gin.Context) {
	companyIDVal, exists := c.Get("companyID")
	var companyID uuid.UUID
	if exists {
		if cid, ok := companyIDVal.(uuid.UUID); ok {
			companyID = cid
		}
	} else {
		companyID = uuid.Nil
	}

	var req domain.PayrollConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid input: "+err.Error())
		return
	}

	if err := h.useCase.UpdateConfig(companyID, &req); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update payroll config: "+err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Payroll config updated successfully", nil)
}
