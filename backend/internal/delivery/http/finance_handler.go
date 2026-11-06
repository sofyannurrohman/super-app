package http

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"example.com/hris/backend/internal/usecase"
)

type FinanceHandler struct {
	usecase usecase.FinanceUsecase
}

func NewFinanceHandler(u usecase.FinanceUsecase) *FinanceHandler {
	return &FinanceHandler{u}
}

func (h *FinanceHandler) RegisterRoutes(r *gin.RouterGroup) {
	finance := r.Group("/finance")
	finance.Use(RoleMiddleware(
		string(domain.RoleSuperAdmin),
		string(domain.RoleHRAdmin),
		string(domain.RoleManager),
		string(domain.RoleCashier),
		string(domain.RoleCashierPusat),
		string(domain.RoleAdminKasir),
		string(domain.RoleAdminNota),
		string(domain.RoleSupervisorSales),
		string(domain.RoleSupervisorSalesMT),
	))
	{
		finance.GET("/profit-loss", h.GetProfitLoss)
		finance.GET("/profit-loss/all", h.GetProfitLossAll)
		finance.GET("/profit-loss/detail", h.GetProfitLossDetail)
		finance.GET("/profit-loss/branch", h.GetProfitLossByBranch)
		finance.GET("/profit-loss/cashier", h.GetCashierProfitLoss)
		finance.GET("/profit-loss/income-detail", h.GetIncomeTransactionDetails)
		finance.GET("/profit-loss/expense-detail", h.GetExpenseTransactionDetails)
		finance.GET("/profit-loss/receivable-detail", h.GetReceivableTransactionDetails)
		finance.GET("/receivables/financial", h.GetFinancialReceivables)
		finance.POST("/receivables/financial/payment", h.RecordFinancialReceivablePayment)
		finance.POST("/receivables/financial/transactions/:id/settle", h.SettleFinancialReceivableTx)
		finance.GET("/payments/pending", h.GetPendingPayments)
		finance.GET("/payments/midtrans", h.GetMidtransPayments)
		finance.GET("/sales-income/daily", h.GetDailySalesIncome)
		finance.POST("/payments/:id/approve", h.ApprovePayment)
		finance.POST("/payments/:id/reject", h.RejectPayment)
		finance.POST("/payments/:id/discrepancy", h.ReportDiscrepancy)
		finance.GET("/export/csv", h.ExportFinancialDataCSV)
		invoiceDeposits := finance.Group("/invoice-deposits")
		invoiceDeposits.Use(RoleMiddleware(string(domain.RoleSuperAdmin), string(domain.RoleManager), string(domain.RoleCashier), string(domain.RoleAdminKasir), string(domain.RoleCashierPusat)))
		{
			invoiceDeposits.GET("", h.ListInvoiceDeposits)
			invoiceDeposits.GET("/available-invoices", h.GetAvailableDepositInvoices)
			invoiceDeposits.POST("", RoleMiddleware(string(domain.RoleSuperAdmin), string(domain.RoleManager), string(domain.RoleCashier), string(domain.RoleAdminKasir)), h.SubmitInvoiceDeposit)
			invoiceDeposits.POST("/:id/approve", RoleMiddleware(string(domain.RoleSuperAdmin), string(domain.RoleManager), string(domain.RoleCashierPusat)), h.ApproveInvoiceDeposit)
			invoiceDeposits.POST("/:id/reject", RoleMiddleware(string(domain.RoleSuperAdmin), string(domain.RoleManager), string(domain.RoleCashierPusat)), h.RejectInvoiceDeposit)
		}

		// Delivery / Surat Jalan Reconciliation
		finance.GET("/deliveries/summary", h.GetDeliverySummary)
		finance.POST("/deliveries/approve", h.ApproveDeliveryBatch)
		finance.POST("/deliveries/resolve-crosscheck", h.ResolveCrosscheck)
		finance.GET("/deliveries/today", h.GetTodayCompletedDeliveries)
	}
}

func (h *FinanceHandler) GetProfitLoss(c *gin.Context) {
	companyIDStr := c.Query("company_id")
	if companyIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "company_id is required"})
		return
	}

	companyID, err := uuid.Parse(companyIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid company_id"})
		return
	}

	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	month, _ := strconv.Atoi(c.Query("month"))
	year, _ := strconv.Atoi(c.Query("year"))

	if startDate == "" && (month == 0 || year == 0) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "month and year, or start_date and end_date are required"})
		return
	}

	report, err := h.usecase.GetProfitLossReport(companyID, month, year, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Success",
		"data":    report,
	})
}

func (h *FinanceHandler) GetProfitLossAll(c *gin.Context) {
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	month, _ := strconv.Atoi(c.Query("month"))
	year, _ := strconv.Atoi(c.Query("year"))

	if startDate == "" && (month == 0 || year == 0) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "month and year, or start_date and end_date are required"})
		return
	}

	reports, err := h.usecase.GetProfitLossAllCompanies(month, year, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Success",
		"data":    reports,
	})
}

func (h *FinanceHandler) GetProfitLossDetail(c *gin.Context) {
	companyIDStr := c.Query("company_id")
	if companyIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "company_id is required"})
		return
	}

	companyID, err := uuid.Parse(companyIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid company_id"})
		return
	}

	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	month, _ := strconv.Atoi(c.Query("month"))
	year, _ := strconv.Atoi(c.Query("year"))

	if startDate == "" && (month == 0 || year == 0) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "month and year, or start_date and end_date are required"})
		return
	}

	report, err := h.usecase.GetProfitLossDetail(companyID, month, year, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Success",
		"data":    report,
	})
}

func (h *FinanceHandler) GetProfitLossByBranch(c *gin.Context) {
	branchIDStr := c.Query("branch_id")
	if branchIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "branch_id is required"})
		return
	}

	branchID, err := uuid.Parse(branchIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid branch_id"})
		return
	}

	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	month, _ := strconv.Atoi(c.Query("month"))
	year, _ := strconv.Atoi(c.Query("year"))

	if startDate == "" && (month == 0 || year == 0) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "month and year, or start_date and end_date are required"})
		return
	}

	report, err := h.usecase.GetProfitLossByBranch(branchID, month, year, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Success",
		"data":    report,
	})
}

func (h *FinanceHandler) GetCashierProfitLoss(c *gin.Context) {
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	month, _ := strconv.Atoi(c.Query("month"))
	year, _ := strconv.Atoi(c.Query("year"))

	if startDate == "" && (month == 0 || year == 0) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "month and year, or start_date and end_date are required"})
		return
	}

	branchIDStr := c.Query("branch_id")
	var branchID *uuid.UUID
	if branchIDStr != "" && branchIDStr != "ALL" && branchIDStr != "ALL_BRANCHES" {
		if bid, err := uuid.Parse(branchIDStr); err == nil && bid != uuid.Nil {
			branchID = &bid
		}
	}

	report, err := h.usecase.GetCashierProfitLossReport(month, year, branchID, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Success",
		"data":    report,
	})
}

func (h *FinanceHandler) GetIncomeTransactionDetails(c *gin.Context) {
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	month, _ := strconv.Atoi(c.Query("month"))
	year, _ := strconv.Atoi(c.Query("year"))
	incomeType := c.Query("income_type")

	if (startDate == "" && (month == 0 || year == 0)) || incomeType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "income_type and (month, year or start_date, end_date) are required"})
		return
	}

	branchIDStr := c.Query("branch_id")
	var branchID *uuid.UUID
	if branchIDStr != "" && branchIDStr != "ALL" && branchIDStr != "ALL_BRANCHES" {
		if bid, err := uuid.Parse(branchIDStr); err == nil && bid != uuid.Nil {
			branchID = &bid
		}
	}

	report, err := h.usecase.GetIncomeTransactionDetails(incomeType, month, year, branchID, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Success",
		"data":    report,
	})
}

func (h *FinanceHandler) GetExpenseTransactionDetails(c *gin.Context) {
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	month, _ := strconv.Atoi(c.Query("month"))
	year, _ := strconv.Atoi(c.Query("year"))
	categoryKey := c.Query("category_key")

	if (startDate == "" && (month == 0 || year == 0)) || categoryKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "category_key and (month, year or start_date, end_date) are required"})
		return
	}

	branchIDStr := c.Query("branch_id")
	var branchID *uuid.UUID
	if branchIDStr != "" && branchIDStr != "ALL" && branchIDStr != "ALL_BRANCHES" {
		if bid, err := uuid.Parse(branchIDStr); err == nil && bid != uuid.Nil {
			branchID = &bid
		}
	}

	report, err := h.usecase.GetExpenseTransactionDetails(categoryKey, month, year, branchID, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Success",
		"data":    report,
	})
}

func (h *FinanceHandler) GetReceivableTransactionDetails(c *gin.Context) {
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	month, _ := strconv.Atoi(c.Query("month"))
	year, _ := strconv.Atoi(c.Query("year"))
	entityCode := c.Query("entity_code")

	if (startDate == "" && (month == 0 || year == 0)) || entityCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "entity_code and (month, year or start_date, end_date) are required"})
		return
	}

	branchIDStr := c.Query("branch_id")
	var branchID *uuid.UUID
	if branchIDStr != "" && branchIDStr != "ALL" && branchIDStr != "ALL_BRANCHES" {
		if bid, err := uuid.Parse(branchIDStr); err == nil && bid != uuid.Nil {
			branchID = &bid
		}
	}

	report, err := h.usecase.GetReceivableTransactionDetails(entityCode, month, year, branchID, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Success",
		"data":    report,
	})
}

func (h *FinanceHandler) GetPendingPayments(c *gin.Context) {
	companyIDStr := c.Query("company_id")
	var companyID uuid.UUID
	if companyIDStr != "" && companyIDStr != "ALL_BRANCHES" {
		var err error
		companyID, err = uuid.Parse(companyIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid company_id"})
			return
		}
	}

	branchIDStr := c.Query("branch_id")
	var branchID uuid.UUID
	if branchIDStr != "" && branchIDStr != "ALL_BRANCHES" {
		var err error
		branchID, err = uuid.Parse(branchIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid branch_id"})
			return
		}
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if limit < 1 && limit != -1 {
		limit = 10
	}
	search := c.Query("search")
	status := c.Query("status")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	payments, total, err := h.usecase.GetPendingPayments(companyID, branchID, page, limit, search, status, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Success",
		"data":    payments,
		"meta": gin.H{
			"total": total,
			"page":  page,
			"limit": limit,
		},
	})
}

func (h *FinanceHandler) GetMidtransPayments(c *gin.Context) {
	companyIDStr := c.Query("company_id")
	var companyID uuid.UUID
	if companyIDStr != "" && companyIDStr != "ALL_BRANCHES" {
		var err error
		companyID, err = uuid.Parse(companyIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid company_id"})
			return
		}
	}

	branchIDStr := c.Query("branch_id")
	var branchID uuid.UUID
	if branchIDStr != "" && branchIDStr != "ALL_BRANCHES" {
		var err error
		branchID, err = uuid.Parse(branchIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid branch_id"})
			return
		}
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if limit < 1 && limit != -1 {
		limit = 10
	}
	search := c.Query("search")
	status := c.Query("status")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	payments, total, err := h.usecase.GetMidtransPayments(companyID, branchID, page, limit, search, status, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Success",
		"data":    payments,
		"meta": gin.H{
			"total": total,
			"page":  page,
			"limit": limit,
		},
	})
}

func (h *FinanceHandler) ApprovePayment(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payment id"})
		return
	}

	var req struct {
		TargetKasPabrikID    string `json:"target_kas_pabrik_id"`
		CompanyBankAccountID string `json:"company_bank_account_id"`
	}
	_ = c.ShouldBindJSON(&req)
	var bankAccountID *uuid.UUID
	if req.CompanyBankAccountID != "" {
		parsed, parseErr := uuid.Parse(req.CompanyBankAccountID)
		if parseErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ID rekening bank tidak valid"})
			return
		}
		bankAccountID = &parsed
	}

	var approverIDPtr *uuid.UUID
	if empIDVal, exists := c.Get("employeeID"); exists {
		if empID, ok := empIDVal.(uuid.UUID); ok && empID != uuid.Nil {
			approverIDPtr = &empID
		}
	}
	if approverIDPtr == nil {
		if uidVal, exists := c.Get("userID"); exists {
			if uid, ok := uidVal.(uuid.UUID); ok && uid != uuid.Nil {
				approverIDPtr = &uid
			}
		}
	}
	if approverIDPtr == nil {
		if uidStr, exists := c.Get("user_id"); exists {
			if str, ok := uidStr.(string); ok {
				if parsed, err := uuid.Parse(str); err == nil && parsed != uuid.Nil {
					approverIDPtr = &parsed
				}
			}
		}
	}

	err = h.usecase.ApprovePayment(id, req.TargetKasPabrikID, bankAccountID, approverIDPtr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Pembayaran berhasil disetujui kasir",
	})
}

func (h *FinanceHandler) RejectPayment(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payment id"})
		return
	}

	var req struct {
		RejectionReason string `json:"rejection_reason"`
	}
	_ = c.ShouldBindJSON(&req)

	err = h.usecase.RejectPayment(id, req.RejectionReason)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Pembayaran tunai berhasil ditolak kasir",
	})
}

func (h *FinanceHandler) ExportFinancialDataCSV(c *gin.Context) {
	companyIDStr := c.Query("company_id")
	if companyIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "company_id is required"})
		return
	}

	companyID, err := uuid.Parse(companyIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid company_id"})
		return
	}

	month, _ := strconv.Atoi(c.Query("month"))
	year, _ := strconv.Atoi(c.Query("year"))

	if month == 0 || year == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "month and year are required"})
		return
	}

	csvBytes, filename, err := h.usecase.ExportFinancialDataCSV(companyID, month, year)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Data(http.StatusOK, "text/csv", csvBytes)
}

func (h *FinanceHandler) GetDeliverySummary(c *gin.Context) {
	doNo := c.Query("do_no")
	if doNo == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "do_no is required"})
		return
	}

	summary, err := h.usecase.GetDeliverySummary(doNo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": summary})
}

func (h *FinanceHandler) ApproveDeliveryBatch(c *gin.Context) {
	var req struct {
		BatchID      string  `json:"batch_id" binding:"required"`
		PhysicalCash float64 `json:"physical_cash"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	batchID, err := uuid.Parse(req.BatchID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid batch_id"})
		return
	}

	if err := h.usecase.ApproveDeliveryBatch(batchID, req.PhysicalCash); err != nil {
		// If error string contains cross-check, treat it as a success response with specific status
		if err.Error() == "terdapat selisih uang fisik, surat jalan dipindahkan ke tahap Cross-check" {
			c.JSON(http.StatusOK, gin.H{"message": err.Error(), "status": "WAITING_CROSSCHECK"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Semua pembayaran untuk Surat Jalan berhasil disetujui", "status": "COMPLETED"})
}

func (h *FinanceHandler) ResolveCrosscheck(c *gin.Context) {
	var req struct {
		BatchID string `json:"batch_id" binding:"required"`
		Notes   string `json:"notes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	batchID, err := uuid.Parse(req.BatchID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid batch_id"})
		return
	}

	if err := h.usecase.ResolveCrosscheck(batchID, req.Notes); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Cross-check selesai. Semua pembayaran untuk Surat Jalan berhasil disetujui"})
}

func (h *FinanceHandler) GetTodayCompletedDeliveries(c *gin.Context) {
	var branchID *uuid.UUID
	if branchIDStr := c.Query("branch_id"); branchIDStr != "" && branchIDStr != "ALL_BRANCHES" {
		parsed, err := uuid.Parse(branchIDStr)
		if err == nil {
			branchID = &parsed
		}
	}

	deliveries, err := h.usecase.GetTodayCompletedDeliveries(branchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "success",
		"data":    deliveries,
	})
}

func (h *FinanceHandler) ReportDiscrepancy(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payment id"})
		return
	}

	var req struct {
		Notes string `json:"notes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.usecase.ReportDiscrepancy(id, req.Notes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Discrepancy reported successfully"})
}

func (h *FinanceHandler) GetFinancialReceivables(c *gin.Context) {
	companyIDStr := c.Query("company_id")
	var companyID uuid.UUID
	if companyIDStr != "" {
		companyID, _ = uuid.Parse(companyIDStr)
	}

	branchIDStr := c.Query("branch_id")
	var branchIDPtr *uuid.UUID
	if branchIDStr != "" {
		if bID, err := uuid.Parse(branchIDStr); err == nil {
			branchIDPtr = &bID
		}
	}

	res, err := h.usecase.GetFinancialReceivablesList(companyID, branchIDPtr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Financial receivables retrieved successfully",
		"data":    res,
	})
}

func (h *FinanceHandler) RecordFinancialReceivablePayment(c *gin.Context) {
	var req domain.RepayBranchFinancialReceivableRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var userIDPtr *uuid.UUID
	if uidStr, exists := c.Get("user_id"); exists {
		if uid, ok := uidStr.(uuid.UUID); ok {
			userIDPtr = &uid
		} else if uidStrStr, ok := uidStr.(string); ok {
			if uid, err := uuid.Parse(uidStrStr); err == nil {
				userIDPtr = &uid
			}
		}
	}

	tx, err := h.usecase.RecordFinancialReceivablePayment(req, userIDPtr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Pelunasan talangan cabang berhasil dicatat",
		"data":    tx,
	})
}

func (h *FinanceHandler) SettleFinancialReceivableTx(c *gin.Context) {
	txIDStr := c.Param("id")
	txID, err := uuid.Parse(txIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID transaksi tidak valid"})
		return
	}

	var req domain.SettleFinancialReceivableTxRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var userIDPtr *uuid.UUID
	if uidStr, exists := c.Get("user_id"); exists {
		if uid, ok := uidStr.(uuid.UUID); ok {
			userIDPtr = &uid
		} else if uidStrStr, ok := uidStr.(string); ok {
			if uid, err := uuid.Parse(uidStrStr); err == nil {
				userIDPtr = &uid
			}
		}
	}

	settlement, err := h.usecase.SettleFinancialReceivableTx(txID, req, userIDPtr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Transaksi talangan berhasil diselesaikan (lunas)",
		"data":    settlement,
	})
}

func (h *FinanceHandler) GetDailySalesIncome(c *gin.Context) {
	date := c.Query("date")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	companyIDStr := c.Query("company_id")
	branchIDStr := c.Query("branch_id")

	report, err := h.usecase.GetDailySalesIncomeReport(date, startDate, endDate, companyIDStr, branchIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Success",
		"data":    report,
	})
}
