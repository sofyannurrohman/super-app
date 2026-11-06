package http

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"example.com/hris/backend/internal/usecase"
	"example.com/hris/backend/pkg/utils"
)

type PerformanceHandler struct {
	performanceUseCase usecase.PerformanceUseCase
	employeeUseCase    usecase.EmployeeUsecase
}

func NewPerformanceHandler(performanceUseCase usecase.PerformanceUseCase, employeeUseCase usecase.EmployeeUsecase) *PerformanceHandler {
	return &PerformanceHandler{
		performanceUseCase: performanceUseCase,
		employeeUseCase:    employeeUseCase,
	}
}

func (h *PerformanceHandler) SetupRoutes(router *gin.RouterGroup) {
	// Root /performance
	performance := router.Group("/performance")
	{
		performance.GET("/my-performance", h.GetMyPerformance)
		performance.GET("/my-history", h.GetMyPerformanceHistory)
	}

	// Admin routes
	admin := performance.Group("")
	admin.Use(RoleMiddleware(string(domain.RoleSuperAdmin), string(domain.RoleHRAdmin), string(domain.RoleManager)))
	{
		admin.GET("/regular", h.GetRegularKPIs)
		admin.PUT("/regular/:id", h.UpdateRegularKPI)
		admin.POST("/regular/:id/finalize", h.FinalizeKPI)
		admin.POST("/appraise", h.CreateAppraisal)
		admin.GET("/appraisals/:employee_id", h.GetAppraisals)
	}
}

func (h *PerformanceHandler) GetRegularKPIs(c *gin.Context) {
	monthStr := c.Query("month")
	yearStr := c.Query("year")

	now := time.Now()
	month := int(now.Month())
	year := now.Year()

	if monthStr != "" {
		if m, err := strconv.Atoi(monthStr); err == nil {
			month = m
		}
	}
	if yearStr != "" {
		if y, err := strconv.Atoi(yearStr); err == nil {
			year = y
		}
	}

	kpis, err := h.performanceUseCase.GetEmployeeKPIs(month, year)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get KPIs: "+err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "KPIs fetched successfully", kpis)
}

type UpdateKPIRequest struct {
	ProductivityScore float64 `json:"productivity_score" binding:"required"`
}

func (h *PerformanceHandler) UpdateRegularKPI(c *gin.Context) {
	idParam := c.Param("id")
	kpiID, err := uuid.Parse(idParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid KPI ID")
		return
	}

	var req UpdateKPIRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	err = h.performanceUseCase.UpdateEmployeeKPI(kpiID, req.ProductivityScore)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update KPI: "+err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "KPI updated successfully", nil)
}

func (h *PerformanceHandler) FinalizeKPI(c *gin.Context) {
	idParam := c.Param("id")
	kpiID, err := uuid.Parse(idParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid KPI ID")
		return
	}

	err = h.performanceUseCase.FinalizeKPI(kpiID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to finalize KPI: "+err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "KPI finalized successfully", nil)
}

func (h *PerformanceHandler) CreateAppraisal(c *gin.Context) {
	var req usecase.EvaluateAppraisalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// Auto inject manager ID from token if not given
	if req.ManagerID == uuid.Nil {
		if userID, exists := c.Get("userID"); exists {
			req.ManagerID = userID.(uuid.UUID)
		}
	}

	err := h.performanceUseCase.EvaluateAndAppraise(req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to save appraisal: "+err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Appraisal saved successfully", nil)
}

func (h *PerformanceHandler) GetAppraisals(c *gin.Context) {
	idParam := c.Param("employee_id")
	empID, err := uuid.Parse(idParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid Employee ID")
		return
	}

	appraisals, err := h.performanceUseCase.GetAppraisals(empID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch appraisals: "+err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Appraisals fetched successfully", appraisals)
}

func (h *PerformanceHandler) GetMyPerformance(c *gin.Context) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized")
		return
	}
	userID := userIDStr.(uuid.UUID)

	employee, err := h.employeeUseCase.GetEmployeeByUserID(userID)
	if err != nil || employee == nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Employee record not found")
		return
	}

	monthStr := c.Query("month")
	yearStr := c.Query("year")
	now := time.Now()
	month := int(now.Month())
	year := now.Year()

	if m, err := strconv.Atoi(monthStr); err == nil {
		month = m
	}
	if y, err := strconv.Atoi(yearStr); err == nil {
		year = y
	}

	res, err := h.performanceUseCase.GetMyPerformance(employee.ID, month, year)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Performance fetched successfully", res)
}

func (h *PerformanceHandler) GetMyPerformanceHistory(c *gin.Context) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized")
		return
	}
	userID := userIDStr.(uuid.UUID)

	employee, err := h.employeeUseCase.GetEmployeeByUserID(userID)
	if err != nil || employee == nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Employee record not found")
		return
	}

	res, err := h.performanceUseCase.GetMyPerformanceHistory(employee.ID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "History fetched successfully", res)
}
