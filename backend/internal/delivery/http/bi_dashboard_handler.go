package http

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"example.com/hris/backend/internal/usecase"
)

type BIDashboardHandler struct {
	uc usecase.BIDashboardUsecase
}

func NewBIDashboardHandler(uc usecase.BIDashboardUsecase) *BIDashboardHandler {
	return &BIDashboardHandler{uc: uc}
}

func (h *BIDashboardHandler) RegisterRoutes(r *gin.RouterGroup) {
	bi := r.Group("/bi")
	bi.Use(RoleMiddleware(
		string(domain.RoleSuperAdmin),
		string(domain.RoleHRAdmin),
		string(domain.RoleManager),
	))
	{
		bi.GET("/summary/hr", h.GetHRSummary)
		bi.GET("/summary/sales", h.GetSalesSummary)
		bi.GET("/summary/payroll", h.GetPayrollSummary)
		bi.GET("/trend/attendance", h.GetAttendanceTrend)
		bi.GET("/trend/sales", h.GetSalesTrend)
		bi.GET("/top-salesperson", h.GetTopSalesperson)
		bi.GET("/revenue-expense", h.GetRevenueExpense)
	}
}

// parseCompanyID extracts company_id from query param or context.
func parseCompanyID(c *gin.Context) uuid.UUID {
	cidStr := c.Query("company_id")
	if cidStr != "" {
		if id, err := uuid.Parse(cidStr); err == nil {
			return id
		}
	}
	// Fall back to company ID in token context
	if cid, ok := c.Get("companyID"); ok {
		if id, ok2 := cid.(uuid.UUID); ok2 {
			return id
		}
	}
	return uuid.Nil
}

// parseMonthYear extracts month/year from query, defaulting to current month/year.
func parseMonthYear(c *gin.Context) (int, int) {
	now := time.Now()
	month, _ := strconv.Atoi(c.DefaultQuery("month", strconv.Itoa(int(now.Month()))))
	year, _ := strconv.Atoi(c.DefaultQuery("year", strconv.Itoa(now.Year())))
	if month < 1 || month > 12 {
		month = int(now.Month())
	}
	if year < 2000 || year > 2100 {
		year = now.Year()
	}
	return month, year
}

// GET /api/v1/bi/summary/hr
func (h *BIDashboardHandler) GetHRSummary(c *gin.Context) {
	companyID := parseCompanyID(c)
	summary, err := h.uc.GetHRSummary(companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": summary})
}

// GET /api/v1/bi/summary/sales?month=8&year=2025
func (h *BIDashboardHandler) GetSalesSummary(c *gin.Context) {
	companyID := parseCompanyID(c)
	month, year := parseMonthYear(c)
	summary, err := h.uc.GetSalesSummary(companyID, month, year)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": summary})
}

// GET /api/v1/bi/summary/payroll?month=8&year=2025
func (h *BIDashboardHandler) GetPayrollSummary(c *gin.Context) {
	companyID := parseCompanyID(c)
	month, year := parseMonthYear(c)
	summary, err := h.uc.GetPayrollSummary(companyID, month, year)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": summary})
}

// GET /api/v1/bi/trend/attendance?months=6
func (h *BIDashboardHandler) GetAttendanceTrend(c *gin.Context) {
	companyID := parseCompanyID(c)
	months, _ := strconv.Atoi(c.DefaultQuery("months", "6"))
	trend, err := h.uc.GetAttendanceTrend(companyID, months)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": trend})
}

// GET /api/v1/bi/trend/sales?months=6
func (h *BIDashboardHandler) GetSalesTrend(c *gin.Context) {
	companyID := parseCompanyID(c)
	months, _ := strconv.Atoi(c.DefaultQuery("months", "6"))
	trend, err := h.uc.GetSalesTrend(companyID, months)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": trend})
}

// GET /api/v1/bi/top-salesperson?month=8&year=2025&limit=5
func (h *BIDashboardHandler) GetTopSalesperson(c *gin.Context) {
	companyID := parseCompanyID(c)
	month, year := parseMonthYear(c)
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "5"))
	ranks, err := h.uc.GetTopSalesperson(companyID, month, year, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": ranks})
}

// GET /api/v1/bi/revenue-expense?months=6
func (h *BIDashboardHandler) GetRevenueExpense(c *gin.Context) {
	companyID := parseCompanyID(c)
	months, _ := strconv.Atoi(c.DefaultQuery("months", "6"))
	points, err := h.uc.GetRevenueExpense(companyID, months)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": points})
}
