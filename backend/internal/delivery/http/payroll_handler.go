package http

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"example.com/hris/backend/internal/usecase"
	"example.com/hris/backend/pkg/utils"
)

type PayrollHandler struct {
	payrollUseCase usecase.PayrollUseCase
}

func NewPayrollHandler(payrollUseCase usecase.PayrollUseCase) *PayrollHandler {
	return &PayrollHandler{payrollUseCase}
}

func (h *PayrollHandler) SetupRoutes(router *gin.RouterGroup) {
	payroll := router.Group("/payroll")
	// For general Employee Self Service (ESS)
	payroll.GET("/my-payslip", h.GetMyPayslips)

	// Only HR Admins, SuperAdmins, and Cashier Pusat can run the payroll batch engine
	adminPayroll := payroll.Group("")
	adminPayroll.Use(RoleMiddleware(string(domain.RoleSuperAdmin), string(domain.RoleHRAdmin), string(domain.RoleCashierPusat)))
	{
		adminPayroll.GET("/all", h.GetAllPayrollRuns)
		adminPayroll.POST("/generate", h.RunPayrollBatch)
		adminPayroll.PUT("/:id/approve", h.ApprovePayrollRun)
		adminPayroll.PUT("/payslips/:id/adjust", h.AdjustPayslip)
		adminPayroll.GET("/:id/export", h.ExportPayrollRunCSV)
		adminPayroll.GET("/:id/export-excel", h.DownloadTransferExcel)
		adminPayroll.GET("/:id/transfer-summary", h.GetTransferSummary)
		adminPayroll.GET("/:id/transfer-file", h.DownloadTransferFile)
		adminPayroll.DELETE("/:id", h.DeletePayrollRun)
		adminPayroll.GET("/:id/details", h.GetPayrollRunDetails)

		// Salary Settings
		adminPayroll.GET("/settings/:employee_id", h.ListEmployeeSalarySettings)
		adminPayroll.POST("/settings", h.SaveEmployeeSalarySetting)
		adminPayroll.DELETE("/settings/:id", h.DeleteEmployeeSalarySetting)
	}
}

func (h *PayrollHandler) RunPayrollBatch(c *gin.Context) {
	var req usecase.GeneratePayrollRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid input for payroll period: "+err.Error())
		return
	}

	// Because only admins can execute this, we use the company ID from context if multitenant,
	// Assuming mono-tenant for now, we'll try to extract the user's company ID (if available in token claims).
	companyIDVal, exists := c.Get("companyID")
	var companyID uuid.UUID
	if exists {
		if cid, ok := companyIDVal.(uuid.UUID); ok {
			companyID = cid
		}
	} else {
		companyID = uuid.Nil
	}

	result, err := h.payrollUseCase.GeneratePayroll(companyID, req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Payroll successfully processed",
		"data": map[string]interface{}{
			"payroll_run_id": result.ID,
			"total_payout":   result.TotalPayout,
			"status":         result.Status,
		},
	})
}

func (h *PayrollHandler) GetMyPayslips(c *gin.Context) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized")
		return
	}
	userID := userIDStr.(uuid.UUID)

	result, err := h.payrollUseCase.GetMyPayslipHistory(userID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Berhasil mengambil riwayat slip gaji",
		"data":    result,
	})
}

func (h *PayrollHandler) ExportPayrollRunCSV(c *gin.Context) {
	runIDStr := c.Param("id")
	runID, err := uuid.Parse(runIDStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid Run ID format")
		return
	}

	csvData, err := h.payrollUseCase.ExportPayrollRunCSV(runID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", fmt.Sprintf("attachment;filename=payroll_export_%s.csv", runIDStr))
	c.Data(http.StatusOK, "text/csv", csvData)
}

func (h *PayrollHandler) GetAllPayrollRuns(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")
	branchIDStr := c.Query("branch_id")

	var page, limit int
	fmt.Sscanf(pageStr, "%d", &page)
	fmt.Sscanf(limitStr, "%d", &limit)

	var branchUUID *uuid.UUID
	if branchIDStr != "" && branchIDStr != "ALL_BRANCHES" {
		parsed, err := uuid.Parse(branchIDStr)
		if err == nil && parsed != uuid.Nil {
			branchUUID = &parsed
		}
	}

	runs, err := h.payrollUseCase.GetAllPayrollRuns(page, limit, branchUUID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get payroll history: "+err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Payroll history fetched successfully", runs)
}

func (h *PayrollHandler) DeletePayrollRun(c *gin.Context) {
	idParam := c.Param("id")
	runID, err := uuid.Parse(idParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid Payroll Run ID")
		return
	}

	if err := h.payrollUseCase.DeletePayrollRun(runID); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete payroll run: "+err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Payroll run deleted successfully", nil)
}

func (h *PayrollHandler) ListEmployeeSalarySettings(c *gin.Context) {
	empIDStr := c.Param("employee_id")
	empID, err := uuid.Parse(empIDStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid Employee ID")
		return
	}

	settings, err := h.payrollUseCase.ListEmployeeSalarySettings(empID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Success", settings)
}

func (h *PayrollHandler) SaveEmployeeSalarySetting(c *gin.Context) {
	var req domain.EmployeeSalarySetting
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.payrollUseCase.SaveEmployeeSalarySetting(&req); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Salary setting saved", req)
}

func (h *PayrollHandler) DeleteEmployeeSalarySetting(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid ID")
		return
	}

	if err := h.payrollUseCase.DeleteEmployeeSalarySetting(id); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Salary setting deleted", nil)
}

func (h *PayrollHandler) GetPayrollRunDetails(c *gin.Context) {
	runIDStr := c.Param("id")
	runID, err := uuid.Parse(runIDStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid Run ID")
		return
	}

	payslips, err := h.payrollUseCase.GetPayrollRunDetails(runID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Success", payslips)
}

// GetTransferSummary — GET /api/v1/payroll/runs/:id/transfer-summary
// Mengembalikan ringkasan karyawan per bank untuk ditampilkan di modal sebelum download.
func (h *PayrollHandler) GetTransferSummary(c *gin.Context) {
	runIDStr := c.Param("id")
	runID, err := uuid.Parse(runIDStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid Payroll Run ID")
		return
	}

	summary, err := h.payrollUseCase.GetTransferSummary(runID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Transfer summary retrieved", summary)
}

// DownloadTransferFile — GET /api/v1/payroll/runs/:id/transfer-file?bank=BCA&format=BCA
// Mendownload file CSV transfer gaji untuk bank tertentu.
// Query params:
//   - bank: "BCA", "MANDIRI", "BNI", "ALL" (default: ALL)
//   - format: "BCA", "MANDIRI", "GENERIC" (default: GENERIC)
func (h *PayrollHandler) DownloadTransferFile(c *gin.Context) {
	runIDStr := c.Param("id")
	runID, err := uuid.Parse(runIDStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid Payroll Run ID")
		return
	}

	bank := c.DefaultQuery("bank", "ALL")
	format := c.DefaultQuery("format", "GENERIC")

	csvData, filename, err := h.payrollUseCase.GenerateBankTransferFile(runID, bank, format)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Header("Content-Type", "text/csv")
	c.Data(http.StatusOK, "text/csv", csvData)
}

// ApprovePayrollRun — PUT /payroll/:id/approve
func (h *PayrollHandler) ApprovePayrollRun(c *gin.Context) {
	runIDStr := c.Param("id")
	runID, err := uuid.Parse(runIDStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid Payroll Run ID")
		return
	}

	run, err := h.payrollUseCase.ApprovePayrollRun(runID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Batch penggajian berhasil disetujui", run)
}

// AdjustPayslip — PUT /payroll/payslips/:id/adjust
func (h *PayrollHandler) AdjustPayslip(c *gin.Context) {
	slipIDStr := c.Param("id")
	slipID, err := uuid.Parse(slipIDStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid Payslip ID")
		return
	}

	var req usecase.AdjustPayslipRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid adjustment payload: "+err.Error())
		return
	}

	slip, err := h.payrollUseCase.AdjustPayslip(slipID, req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Rincian slip gaji berhasil disesuaikan", slip)
}

// DownloadTransferExcel — GET /payroll/:id/export-excel?bank=BCA&format=BCA
func (h *PayrollHandler) DownloadTransferExcel(c *gin.Context) {
	runIDStr := c.Param("id")
	runID, err := uuid.Parse(runIDStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid Payroll Run ID")
		return
	}

	bank := c.DefaultQuery("bank", "ALL")
	format := c.DefaultQuery("format", "GENERIC")

	excelData, filename, err := h.payrollUseCase.GenerateBankTransferExcel(runID, bank, format)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", excelData)
}
