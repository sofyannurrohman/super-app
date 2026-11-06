package http

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"example.com/hris/backend/internal/repository"
	"example.com/hris/backend/internal/usecase"
	"example.com/hris/backend/pkg/utils"
)

type DailyKpiReportHandler struct {
	usecase usecase.DailyKpiReportUsecase
}

func NewDailyKpiReportHandler(usecase usecase.DailyKpiReportUsecase) *DailyKpiReportHandler {
	return &DailyKpiReportHandler{usecase: usecase}
}

func (h *DailyKpiReportHandler) SetupRoutes(router *gin.RouterGroup) {
	reports := router.Group("/sales/daily-kpi-reports")
	{
		reports.GET("/prefill", h.GetPrefill)
		reports.GET("", h.GetReports)
		reports.POST("", h.CreateReport)
		reports.GET("/:id", h.GetReportByID)
		reports.PUT("/:id", h.UpdateReport)
		reports.DELETE("/:id", h.DeleteReport)
		reports.POST("/:id/submit", h.SubmitReport)
		reports.POST("/:id/review", h.ReviewReport)
		reports.GET("/:id/export", h.ExportExcel)
	}
}

func (h *DailyKpiReportHandler) GetPrefill(c *gin.Context) {
	empIDVal, exists := c.Get("employeeID")
	if !exists {
		utils.ErrorResponse(c, http.StatusForbidden, "Profil karyawan tidak ditemukan untuk user ini")
		return
	}
	empID := empIDVal.(uuid.UUID)

	dateStr := c.DefaultQuery("date", time.Now().Format("2006-01-02"))
	targetDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		targetDate = time.Now()
	}

	var branchID *uuid.UUID
	branchStr := c.Query("branch_id")
	if branchStr != "" && branchStr != "ALL" && branchStr != "ALL_BRANCHES" {
		if bid, errBid := uuid.Parse(branchStr); errBid == nil {
			branchID = &bid
		}
	}

	res, err := h.usecase.GetPrefillData(empID, branchID, targetDate)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Prefill data retrieved successfully", res)
}

func (h *DailyKpiReportHandler) GetReports(c *gin.Context) {
	filter := repository.DailyKpiFilter{}

	dateStr := c.Query("date")
	if dateStr != "" {
		if parsed, err := time.Parse("2006-01-02", dateStr); err == nil {
			filter.Date = &parsed
		}
	}

	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")
	if startDateStr != "" && endDateStr != "" {
		sTime, errS := time.Parse("2006-01-02", startDateStr)
		eTime, errE := time.Parse("2006-01-02", endDateStr)
		if errS == nil && errE == nil {
			filter.StartDate = &sTime
			filter.EndDate = &eTime
		}
	}

	branchStr := c.Query("branch_id")
	if branchStr != "" && branchStr != "ALL" && branchStr != "ALL_BRANCHES" {
		if bid, errBid := uuid.Parse(branchStr); errBid == nil {
			filter.BranchID = &bid
		}
	}

	repEmpStr := c.Query("reporter_employee_id")
	if repEmpStr != "" {
		if rid, errRid := uuid.Parse(repEmpStr); errRid == nil {
			filter.ReporterEmployeeID = &rid
		}
	}

	revEmpStr := c.Query("reviewer_employee_id")
	if revEmpStr != "" {
		if rvid, errRvid := uuid.Parse(revEmpStr); errRvid == nil {
			filter.ReviewerEmployeeID = &rvid
		}
	}

	filter.ReporterRole = c.Query("reporter_role")
	filter.Status = c.Query("status")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	filter.Page = page
	filter.Limit = limit

	reports, total, err := h.usecase.GetReports(filter)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "Reports retrieved successfully",
		"data":    reports,
		"meta": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

func (h *DailyKpiReportHandler) GetReportByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid report ID")
		return
	}

	report, err := h.usecase.GetReportByID(id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Report not found")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Report retrieved successfully", report)
}

func (h *DailyKpiReportHandler) CreateReport(c *gin.Context) {
	empIDVal, exists := c.Get("employeeID")
	if !exists {
		utils.ErrorResponse(c, http.StatusForbidden, "Profil karyawan tidak ditemukan untuk user ini")
		return
	}
	empID := empIDVal.(uuid.UUID)

	userRole := ""
	if r, ok := c.Get("role"); ok {
		userRole = r.(string)
	}

	var req usecase.CreateDailyKpiReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	report, err := h.usecase.CreateReport(req, empID, userRole)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Laporan kinerja harian berhasil disimpan", report)
}

func (h *DailyKpiReportHandler) UpdateReport(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid report ID")
		return
	}

	empIDVal, exists := c.Get("employeeID")
	if !exists {
		utils.ErrorResponse(c, http.StatusForbidden, "Profil karyawan tidak ditemukan")
		return
	}
	empID := empIDVal.(uuid.UUID)

	userRole := ""
	if r, ok := c.Get("role"); ok {
		userRole = r.(string)
	}
	isSuperadmin := strings.ToLower(userRole) == "superadmin"

	var req usecase.CreateDailyKpiReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	report, err := h.usecase.UpdateReport(id, req, empID, isSuperadmin)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Laporan kinerja harian berhasil diperbarui", report)
}

func (h *DailyKpiReportHandler) SubmitReport(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid report ID")
		return
	}

	empIDVal, exists := c.Get("employeeID")
	if !exists {
		utils.ErrorResponse(c, http.StatusForbidden, "Profil karyawan tidak ditemukan")
		return
	}
	empID := empIDVal.(uuid.UUID)

	if err := h.usecase.SubmitReport(id, empID); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Laporan berhasil dikirim ke atasan", nil)
}

type ReviewReportRequest struct {
	Status   string `json:"status" binding:"required"` // APPROVED or REVISED
	Feedback string `json:"feedback"`
}

func (h *DailyKpiReportHandler) ReviewReport(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid report ID")
		return
	}

	empIDVal, exists := c.Get("employeeID")
	if !exists {
		utils.ErrorResponse(c, http.StatusForbidden, "Profil karyawan tidak ditemukan")
		return
	}
	reviewerID := empIDVal.(uuid.UUID)

	var req ReviewReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.usecase.ReviewReport(id, reviewerID, req.Status, req.Feedback); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	actionWord := "disetujui"
	if strings.ToUpper(req.Status) == "REVISED" {
		actionWord = "dikembalikan untuk revisi"
	}

	utils.SuccessResponse(c, http.StatusOK, "Laporan berhasil "+actionWord, nil)
}

func (h *DailyKpiReportHandler) DeleteReport(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid report ID")
		return
	}

	empIDVal, exists := c.Get("employeeID")
	if !exists {
		utils.ErrorResponse(c, http.StatusForbidden, "Profil karyawan tidak ditemukan")
		return
	}
	empID := empIDVal.(uuid.UUID)

	userRole := ""
	if r, ok := c.Get("role"); ok {
		userRole = r.(string)
	}
	isSuperadmin := strings.ToLower(userRole) == "superadmin"

	if err := h.usecase.DeleteReport(id, empID, isSuperadmin); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Laporan draft berhasil dihapus", nil)
}

func (h *DailyKpiReportHandler) ExportExcel(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid report ID")
		return
	}

	fileBytes, filename, err := h.usecase.ExportExcel(id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", fileBytes)
}
