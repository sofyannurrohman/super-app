package http

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"example.com/hris/backend/internal/domain"
	"example.com/hris/backend/internal/repository"
	"example.com/hris/backend/internal/usecase"
	"example.com/hris/backend/pkg/redisclient"
	"example.com/hris/backend/pkg/sse"
	"example.com/hris/backend/pkg/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SalesHandler struct {
	salesUsecase    usecase.SalesUsecase
	employeeUseCase usecase.EmployeeUsecase
	attendanceRepo  repository.AttendanceRepository
	auditLogUseCase usecase.AuditLogUsecase
	sseHub          *sse.Hub
	fcmUseCase      usecase.FcmUsecase
	db              *gorm.DB
}

func NewSalesHandler(salesUsecase usecase.SalesUsecase, employeeUseCase usecase.EmployeeUsecase, attendanceRepo repository.AttendanceRepository, auditLogUseCase usecase.AuditLogUsecase, sseHub *sse.Hub, fcmUseCase usecase.FcmUsecase, db *gorm.DB) *SalesHandler {
	h := &SalesHandler{
		salesUsecase:    salesUsecase,
		employeeUseCase: employeeUseCase,
		attendanceRepo:  attendanceRepo,
		auditLogUseCase: auditLogUseCase,
		sseHub:          sseHub,
		fcmUseCase:      fcmUseCase,
		db:              db,
	}
	return h
}

func (h *SalesHandler) RegisterRoutes(router *gin.RouterGroup) {
	sales := router.Group("/admin/sales")
	sales.Use(RoleMiddleware(
		string(domain.RoleSuperAdmin),
		string(domain.RoleHRAdmin),
		string(domain.RoleManager),
		string(domain.RoleAdminNota),
		string(domain.RoleSupervisorSales),
		string(domain.RoleSupervisorSalesMT),
		string(domain.RoleCashier),
		string(domain.RoleCashierPusat),
		string(domain.RoleAdminKasir),
	))
	{
		sales.POST("/manual-entry", h.ManualEntry)
		sales.POST("/direct-sales", h.CreateDirectSale)
		sales.GET("/direct-sales/next-invoice-no", h.GetNextDirectSaleInvoiceNo)
		sales.GET("/transactions/all-pending", h.GetAllPending)
		sales.GET("/transactions/outstanding", h.GetAllOutstandingInvoices)
		sales.GET("/transactions/all", h.GetAllTransactions)
		sales.GET("/transactions/next-receipt-no", h.GetNextReceiptNo)
		sales.GET("/delivery-pending", h.GetTransactions)
		sales.GET("/visits", h.GetVisitHistory)
		sales.POST("/visits", h.CreateVisit)
		sales.POST("/visits/batch", h.CreateBatchVisits)
		sales.PUT("/visits/:id", h.UpdateVisit)
		sales.DELETE("/visits/:id", h.DeleteVisit)
		sales.PUT("/transactions/:id", h.UpdateTransaction)
		sales.POST("/transactions/:id/payments", h.RecordAdminPayment)
		sales.POST("/transactions/:id/verify-penerimaan", h.VerifyPenerimaan)
		sales.GET("/penerimaan/tempo-payments", h.GetPenerimaanTempoPayments)
		sales.POST("/penerimaan/tempo-payments/:id/verify", h.VerifyPenerimaanTempoPayment)
		sales.POST("/transactions/:id/redeliver", h.RedeliverTransaction)
		sales.POST("/transactions/:id/cancel-failed-delivery", RoleMiddleware(string(domain.RoleSuperAdmin), string(domain.RoleManager), string(domain.RoleAdminNota)), h.CancelFailedDeliveryTransaction)
		sales.POST("/transactions/:id/restore", h.RestoreTransaction)
		sales.POST("/transactions/:id/convert-to-so", h.ConvertToSalesOrder)
		sales.DELETE("/transactions/:id", h.DeleteTransaction)
		sales.POST("/reconcile-returns", h.ReconcileReturns)
		sales.POST("/targets", h.SetKPITarget)
		sales.DELETE("/targets/:id", h.DeleteKPITarget)
		sales.GET("/transactions/export", h.ExportTransactionsExcel)
		sales.GET("/reports/excel", h.ExportExcel)
		sales.GET("/reports/summary", h.GetSummary)
		sales.GET("/reports/performance", h.GetPerformance)
		sales.GET("/reports/performance/:employee_id/product-recap", h.GetEmployeeProductRecap)
		sales.GET("/reports/product-distribution", h.GetProductDistribution)
		sales.GET("/reports/top-products", h.GetTopProducts)
		sales.GET("/reports/city-revenue", h.GetCityRevenueReport)
		sales.GET("/reports/city-revenue/export", h.ExportCityRevenueExcel)
		sales.GET("/kasbon", h.GetSalesKasbon)
		sales.POST("/upload", h.UploadPhoto)
		sales.GET("/visit-plans", h.GetAdminVisitPlans)
		sales.GET("/live-locations", h.GetLiveLocations)
		sales.GET("/tracking/noo-ec", h.GetNooEcTracking)
		sales.GET("/tracking/noo-ec/export", h.ExportNooEcTrackingExcel)
		sales.POST("/remote-sync/:employee_id", h.TriggerRemoteSync)
		sales.POST("/remote-sync-all", h.TriggerRemoteSyncAll)
		sales.POST("/import-offline-backup", h.ImportOfflineBackup)
	}
}

func (h *SalesHandler) SetupMobileRoutes(router *gin.RouterGroup) {
	sales := router.Group("/sales")
	{
		sales.POST("/transactions", h.CreateTransaction)
		sales.PUT("/transactions/:id", h.UpdateTransactionMobile)
		sales.GET("/transactions/pending", h.GetPendingTransactions)
		sales.GET("/transactions/history", h.GetHistoryTransactions)
		sales.DELETE("/transactions/:id/pending", h.DeletePendingTransactionMobile)
		sales.GET("/visits/pending", h.GetPendingVisits)
		sales.GET("/transactions/settlement-queue", h.GetSettlementQueue)
		sales.PATCH("/transactions/:id/verify", h.VerifyTransaction)
		sales.POST("/transactions/:id/settle-omzet", h.SettleOmzet)
		sales.PATCH("/transactions/:id/reject", h.RejectTransaction)
		sales.POST("/visit-plans/manual", h.CreateManualVisitPlan)
		sales.GET("/visit-plans/builder", h.GetVisitPlanBuilder)
		sales.GET("/visit-plans", h.GetVisitPlan)
		sales.DELETE("/visit-plans/:store_id", h.DeleteManualVisitPlan)
		sales.GET("/transactions/receipt/:receipt_no", h.GetByReceipt)
		sales.POST("/transactions/:id/payments", h.RecordPayment)
		sales.POST("/transactions/:id/regenerate-payment", h.RegeneratePayment)
		sales.GET("/stores/:id/outstanding", h.GetOutstandingByStore)
		sales.GET("/transactions/due-date", h.GetByDueDate)
		sales.GET("/transactions/status/:status", h.GetTransactions)
		sales.POST("/attendance", h.RecordAttendance)
		sales.POST("/upload", h.UploadPhoto)
		sales.POST("/live-location", h.UpdateLiveLocation)
		sales.GET("/live-locations", h.GetLiveLocations)
	}
}

func (h *SalesHandler) RecordAttendance(c *gin.Context) {
	fmt.Println("DEBUG: SalesHandler.RecordAttendance hit")
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}

	// Restore body
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	// Pre-process JSON to fix empty strings for UUID fields from old app
	var rawMap map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &rawMap); err == nil {
		fieldsToClean := []string{"store_id", "employee_id", "latitude", "longitude"}
		changed := false
		for _, field := range fieldsToClean {
			if val, ok := rawMap[field]; ok && val == "" {
				rawMap[field] = nil
				changed = true
			}
		}
		if changed {
			if newBytes, err := json.Marshal(rawMap); err == nil {
				c.Request.Body = io.NopCloser(bytes.NewBuffer(newBytes))
			}
		}
	}

	var req usecase.RecordVisitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format: " + err.Error()})
		return
	}

	// 1. Try to get EmployeeID directly from context (set by EmployeeMiddleware)
	if empID, exists := c.Get("employeeID"); exists {
		req.EmployeeID = empID.(uuid.UUID)
		fmt.Printf("DEBUG: Resolved EmployeeID from context: %v\n", req.EmployeeID)
	}

	// 2. Fallback to resolution from userID if still empty
	if req.EmployeeID == uuid.Nil {
		if userIDStr, exists := c.Get("userID"); exists {
			userID := userIDStr.(uuid.UUID)
			employee, err := h.employeeUseCase.GetEmployeeByUserID(userID)
			if err == nil && employee != nil {
				req.EmployeeID = employee.ID
				fmt.Printf("DEBUG: Resolved EmployeeID from userID: %v\n", req.EmployeeID)
			} else {
				fmt.Printf("DEBUG: Failed to resolve employee from userID %v: %v\n", userID, err)
			}
		}
	}

	// 3. Final check
	if req.EmployeeID == uuid.Nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "User is not associated with an employee record"})
		return
	}

	if req.CheckTime.IsZero() {
		req.CheckTime = time.Now()
	}

	if err := h.salesUsecase.RecordVisit(req); err != nil {
		fmt.Printf("DEBUG: RecordVisit failed: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record visit: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Visit recorded successfully"})
}

func (h *SalesHandler) GetPendingVisits(c *gin.Context) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userID := userIDStr.(uuid.UUID)
	employee, err := h.employeeUseCase.GetEmployeeByUserID(userID)
	if err != nil || employee == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get employee profile"})
		return
	}

	visits, err := h.salesUsecase.GetPendingVisits(employee.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Success",
		"data":    visits,
	})
}

func (h *SalesHandler) GetVisitHistory(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if c.Query("all") == "true" {
		limit = -1
		offset = 0
	}

	var empID *uuid.UUID
	if idStr := c.Query("employee_id"); idStr != "" {
		if id, err := uuid.Parse(idStr); err == nil {
			empID = &id
		}
	}

	var branchID *uuid.UUID
	if idStr := c.Query("branch_id"); idStr != "" {
		if id, err := uuid.Parse(idStr); err == nil {
			branchID = &id
		}
	}

	var start *time.Time
	if sStr := c.Query("start_date"); sStr != "" {
		if t, err := time.Parse("2006-01-02", sStr); err == nil {
			jakartaLoc, locErr := time.LoadLocation("Asia/Jakarta")
			if locErr != nil {
				jakartaLoc = time.Local
			}
			sDate := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, jakartaLoc)
			start = &sDate
		}
	}

	var end *time.Time
	if eStr := c.Query("end_date"); eStr != "" {
		if t, err := time.Parse("2006-01-02", eStr); err == nil {
			jakartaLoc, locErr := time.LoadLocation("Asia/Jakarta")
			if locErr != nil {
				jakartaLoc = time.Local
			}
			// Set to end of day in WIB
			eDate := time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 0, jakartaLoc)
			end = &eDate
		}
	}

	salesType := c.Query("sales_type")
	summary := c.Query("summary") == "true"
	search := c.Query("search")
	if search == "" {
		search = c.Query("q")
	}

	visits, total, err := h.salesUsecase.GetVisitHistory(limit, offset, empID, branchID, salesType, search, start, end, summary)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Success",
		"data":    visits,
		"total":   total,
	})
}

func (h *SalesHandler) CreateVisit(c *gin.Context) {
	var req usecase.RecordVisitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	visit, err := h.salesUsecase.CreateVisit(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Visit created successfully",
		"data":    visit,
	})
}

type BatchRecordVisitRequest struct {
	Visits []usecase.RecordVisitRequest `json:"visits" binding:"required"`
}

func (h *SalesHandler) CreateBatchVisits(c *gin.Context) {
	var req BatchRecordVisitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(req.Visits) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tidak ada kunjungan yang dikirim"})
		return
	}

	visits, err := h.salesUsecase.CreateBatchVisits(req.Visits)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": fmt.Sprintf("%d kunjungan berhasil dibuat", len(visits)),
		"data":    visits,
	})
}

func (h *SalesHandler) UpdateVisit(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid visit id"})
		return
	}

	var req usecase.RecordVisitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.salesUsecase.UpdateVisit(id, req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Visit updated successfully"})
}

func (h *SalesHandler) DeleteVisit(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid visit id"})
		return
	}

	if err := h.salesUsecase.DeleteVisit(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Visit deleted successfully"})
}

func (h *SalesHandler) SetupPublicRoutes(router *gin.RouterGroup) {
	sales := router.Group("/sales")
	{
		sales.POST("/payments/notification", h.HandleMidtransNotification)
		sales.GET("/invoice/detail/*receiptNo", h.GetPublicInvoiceDetail)
		sales.GET("/invoice-detail", h.GetPublicInvoiceDetail)
		sales.GET("/qris-image", h.QrisImageProxy)
	}
}

func (h *SalesHandler) GetPublicInvoiceDetail(c *gin.Context) {
	receiptNo := strings.TrimPrefix(c.Param("receiptNo"), "/")
	if receiptNo == "" {
		receiptNo = c.Query("receipt_no")
	}
	if receiptNo == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nomor faktur tidak boleh kosong"})
		return
	}

	trx, err := h.salesUsecase.GetTransactionByReceipt(receiptNo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if trx == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Faktur tidak ditemukan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Faktur berhasil ditemukan",
		"data":    trx,
	})
}

func (h *SalesHandler) ManualEntry(c *gin.Context) {
	var req usecase.ManualEntryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if role, _ := c.Get("role"); role == string(domain.RoleAdminNota) {
		req.DeferPaymentToCashier = true
	}

	trx, err := h.salesUsecase.ManualEntry(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Manual entry successful",
		"data":    trx,
	})
}

func (h *SalesHandler) CreateDirectSale(c *gin.Context) {
	var req usecase.DirectSaleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if role, _ := c.Get("role"); role == string(domain.RoleAdminNota) {
		req.DeferPaymentToCashier = true
	}

	// Resolve cashier if not provided
	if req.CashierEmployeeID == nil || *req.CashierEmployeeID == uuid.Nil {
		if empID, exists := c.Get("employeeID"); exists {
			if id, ok := empID.(uuid.UUID); ok {
				req.CashierEmployeeID = &id
			}
		}
	}

	trx, err := h.salesUsecase.CreateDirectSale(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Faktur penjualan langsung berhasil dibuat",
		"data":    trx,
	})
}

func (h *SalesHandler) GetNextDirectSaleInvoiceNo(c *gin.Context) {
	branchIDStr := c.Query("branch_id")
	var branchID *uuid.UUID
	if branchIDStr != "" && branchIDStr != "ALL" {
		if id, err := uuid.Parse(branchIDStr); err == nil {
			branchID = &id
		}
	}

	dateStr := c.Query("date")
	var date time.Time
	if dateStr != "" {
		date, _ = time.Parse("2006-01-02", dateStr)
	} else {
		date = time.Now()
	}

	invoiceNo, err := h.salesUsecase.GetNextDirectSaleInvoiceNo(branchID, date)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"invoice_no": invoiceNo})
}

func (h *SalesHandler) ExportExcel(c *gin.Context) {
	monthStr := c.DefaultQuery("month", "1")
	yearStr := c.DefaultQuery("year", "2024")
	month, _ := strconv.Atoi(monthStr)
	year, _ := strconv.Atoi(yearStr)

	branchIDs := utils.ParseUUIDs(c.Query("branch_id"))

	data, err := h.salesUsecase.GenerateKPIReportExcel(month, year, branchIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Header("Content-Disposition", "attachment; filename=sales_kpi_report.csv")
	c.Data(http.StatusOK, "text/csv", data)
}

func (h *SalesHandler) GetSummary(c *gin.Context) {
	now := time.Now()
	monthStr := c.DefaultQuery("month", strconv.Itoa(int(now.Month())))
	yearStr := c.DefaultQuery("year", strconv.Itoa(now.Year()))
	month, _ := strconv.Atoi(monthStr)
	year, _ := strconv.Atoi(yearStr)

	periodType := c.DefaultQuery("period_type", "daily")
	dateStr := c.Query("date")
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")
	salesType := c.Query("sales_type")

	if periodType == "custom" && startDateStr != "" {
		if endDateStr != "" && endDateStr != startDateStr {
			dateStr = fmt.Sprintf("%s,%s", startDateStr, endDateStr)
		} else {
			dateStr = startDateStr
		}
	}

	branchIDs := utils.ParseUUIDs(c.Query("branch_id"))

	summary, err := h.salesUsecase.GetSummaryKPI(periodType, dateStr, month, year, branchIDs, salesType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Success retrieving summary KPI",
		"data":    summary,
	})
}

func (h *SalesHandler) GetSalesKasbon(c *gin.Context) {
	companyIDStr := c.Query("company_id")
	var companyID uuid.UUID
	if companyIDStr != "" && companyIDStr != "ALL" {
		parsed, err := uuid.Parse(companyIDStr)
		if err == nil {
			companyID = parsed
		}
	}

	kasbon, err := h.salesUsecase.GetSalesKasbon(companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Success",
		"data":    kasbon,
	})
}

func (h *SalesHandler) GetSettlementQueue(c *gin.Context) {
	queue, err := h.salesUsecase.GetSettlementQueue()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Success retrieving settlement queue",
		"data":    queue,
	})
}

// SettleOmzet digunakan Admin Nota untuk mensetujui settlement MT.
func (h *SalesHandler) SettleOmzet(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID Transaksi tidak valid"})
		return
	}

	var verifiedByID *uuid.UUID
	if userIDStr, exists := c.Get("userID"); exists {
		userID := userIDStr.(uuid.UUID)
		emp, err := h.employeeUseCase.GetEmployeeByUserID(userID)
		if err == nil && emp != nil {
			verifiedByID = &emp.ID
		}
	}

	// 1. Dapatkan transaction
	trx, err := h.salesUsecase.SettleOmzetMT(id, verifiedByID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal update transaksi: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Omzet berhasil disettle", "data": trx})
}

func (h *SalesHandler) GetPerformance(c *gin.Context) {
	loc, errLoc := time.LoadLocation("Asia/Jakarta")
	if errLoc != nil {
		loc = time.Local
	}
	now := time.Now().In(loc)
	monthStr := c.DefaultQuery("month", strconv.Itoa(int(now.Month())))
	yearStr := c.DefaultQuery("year", strconv.Itoa(now.Year()))
	month, _ := strconv.Atoi(monthStr)
	year, _ := strconv.Atoi(yearStr)

	branchIDs := utils.ParseUUIDs(c.Query("branch_id"))

	periodType := c.DefaultQuery("period_type", "")
	dateStr := c.Query("date")
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	var performance []domain.SalesKPI
	var err error

	if periodType == "custom" && (startDateStr != "" || dateStr != "") {
		sStr := startDateStr
		if sStr == "" {
			sStr = dateStr
		}
		eStr := endDateStr
		if eStr == "" {
			eStr = sStr
		}

		sTime, errS := time.ParseInLocation("2006-01-02", sStr, loc)
		eTime, errE := time.ParseInLocation("2006-01-02", eStr, loc)
		if errS == nil && errE == nil {
			if sTime.After(eTime) {
				sTime, eTime = eTime, sTime
			}
			performance, err = h.salesUsecase.GetPerformanceByDateRange(sTime, eTime, branchIDs)
		} else {
			performance, err = h.salesUsecase.GetDailyPerformanceList(now, branchIDs)
		}
	} else if periodType == "daily" || (dateStr != "" && periodType != "monthly" && periodType != "yearly") {
		targetDate := now
		if dateStr != "" {
			if parsed, errParse := time.ParseInLocation("2006-01-02", dateStr, loc); errParse == nil {
				targetDate = parsed
			} else {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format, use YYYY-MM-DD"})
				return
			}
		}
		performance, err = h.salesUsecase.GetDailyPerformanceList(targetDate, branchIDs)
	} else if periodType == "yearly" {
		performance, err = h.salesUsecase.GetYearlyPerformanceList(year, branchIDs)
	} else {
		// default monthly
		performance, err = h.salesUsecase.GetPerformanceList(month, year, branchIDs)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	salesType := c.Query("sales_type")
	if salesType != "" && salesType != "ALL_SALES" && salesType != "ALL" {
		var filtered []domain.SalesKPI
		for _, kpi := range performance {
			if kpi.Employee != nil && kpi.Employee.JobPosition != nil {
				jpTitle := strings.ToLower(kpi.Employee.JobPosition.Title)

				// Skip operational roles
				if strings.Contains(jpTitle, "operator") || strings.Contains(jpTitle, "driver") || strings.Contains(jpTitle, "auditor") || strings.Contains(jpTitle, "editor") || strings.Contains(jpTitle, "customer service") || strings.Contains(jpTitle, "validator") {
					continue
				}

				if salesType == "sales_mt" {
					if strings.Contains(jpTitle, "sales mt") || strings.Contains(jpTitle, "modern trade") || strings.Contains(jpTitle, "kam") || strings.Contains(jpTitle, "key account") || jpTitle == "mt" || strings.HasPrefix(jpTitle, "mt ") || strings.HasSuffix(jpTitle, " mt") {
						filtered = append(filtered, kpi)
					}
				} else if salesType == "sales_to" {
					if strings.Contains(jpTitle, "take order") || strings.Contains(jpTitle, "task order") || strings.Contains(jpTitle, "sales to") || jpTitle == "to" || strings.HasPrefix(jpTitle, "to ") || strings.HasSuffix(jpTitle, " to") {
						filtered = append(filtered, kpi)
					}
				} else if salesType == "sales_motoris_booster" {
					if strings.Contains(jpTitle, "motoris") || strings.Contains(jpTitle, "booster") {
						filtered = append(filtered, kpi)
					}
				}
			}
		}
		performance = filtered
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Success retrieving performance list",
		"data":    performance,
	})
}

func (h *SalesHandler) GetEmployeeProductRecap(c *gin.Context) {
	employeeIDParam := c.Param("employee_id")
	employeeID, err := uuid.Parse(employeeIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid employee ID format"})
		return
	}

	monthStr := c.DefaultQuery("month", "1")
	yearStr := c.DefaultQuery("year", "2024")
	month, _ := strconv.Atoi(monthStr)
	year, _ := strconv.Atoi(yearStr)

	recaps, err := h.salesUsecase.GetEmployeeProductRecap(employeeID, month, year)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Success retrieving product recap",
		"data":    recaps,
	})
}

func (h *SalesHandler) GetAllPending(c *gin.Context) {
	companyID := c.Query("company_id")
	branchID := c.Query("branch_id")
	salesType := c.Query("sales_type")

	// Hardcode PENDING status and reasonable limits to prevent database lockups
	transactions, _, _, err := h.salesUsecase.GetAllTransactionsPaginated(1, 1000, companyID, "PENDING", "", "", "", "", "", salesType, branchID, "", "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Success retrieving all pending transactions",
		"data":    transactions,
	})
}

func (h *SalesHandler) GetAllTransactions(c *gin.Context) {
	pageStr := c.Query("page")
	limitStr := c.Query("limit")
	companyID := c.Query("company_id")
	status := c.Query("status")
	search := c.Query("search")
	salesman := c.Query("salesman")
	month := c.Query("month")
	year := c.Query("year")
	date := c.Query("date")
	branchID := c.Query("branch_id")
	salesType := c.Query("sales_type")
	paymentMethod := c.Query("payment_method")
	paymentStatus := c.Query("payment_status")
	targetEntity := c.Query("target_entity")

	page := 1
	if pageStr != "" {
		p, err := strconv.Atoi(pageStr)
		if err == nil && p > 0 {
			page = p
		}
	}

	limit := 1000 // default reasonable limit instead of full database scan
	if limitStr != "" {
		if limitStr == "all" || limitStr == "-1" {
			limit = -1
		} else {
			l, err := strconv.Atoi(limitStr)
			if err == nil {
				limit = l
			}
		}
	}

	trxs, total, totalAmount, err := h.salesUsecase.GetAllTransactionsPaginated(page, limit, companyID, status, search, salesman, month, year, date, salesType, branchID, paymentMethod, paymentStatus, targetEntity)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Success",
		"data":         trxs,
		"total":        total,
		"total_amount": totalAmount,
		"page":         page,
		"limit":        limit,
	})
}

func (h *SalesHandler) GetAllOutstandingInvoices(c *gin.Context) {
	companyIDStr := c.Query("company_id")
	branchIDStr := c.Query("branch_id")
	salesType := c.Query("sales_type")

	trxs, err := h.salesUsecase.GetAllOutstandingInvoices(companyIDStr, branchIDStr, salesType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Success",
		"data":    trxs,
	})
}

func (h *SalesHandler) ExportTransactionsExcel(c *gin.Context) {
	companyID := c.Query("company_id")
	status := c.Query("status")
	search := c.Query("search")
	salesman := c.Query("salesman")
	month := c.Query("month")
	year := c.Query("year")
	date := c.Query("date")
	branchID := c.Query("branch_id")
	salesType := c.Query("sales_type")
	paymentMethod := c.Query("payment_method")
	paymentStatus := c.Query("payment_status")
	targetEntity := c.Query("target_entity")

	excelBytes, err := h.salesUsecase.ExportTransactionsExcel(companyID, status, search, salesman, month, year, date, salesType, branchID, paymentMethod, paymentStatus, targetEntity)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate Excel: " + err.Error()})
		return
	}

	c.Header("Content-Disposition", `attachment; filename="faktur_penjualan.xlsx"`)
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", excelBytes)
}

func (h *SalesHandler) CreateTransaction(c *gin.Context) {
	startTime := time.Now()
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}
	payloadSize := len(bodyBytes)

	// Restore body
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	// Pre-process JSON to fix common empty string issues from old app
	var rawMap map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &rawMap); err == nil {
		fieldsToClean := []string{
			"payment_due_date", "visit_id", "company_bank_account_id", "factory_company_id",
			"latitude", "longitude", "company_id", "employee_id", "store_id", "branch_id",
		}
		changed := false
		for _, field := range fieldsToClean {
			if val, ok := rawMap[field]; ok {
				if val == "" || val == "00000000-0000-0000-0000-000000000000" {
					rawMap[field] = nil
					changed = true
				}
			}
		}
		// Sanitize items array: remove items with empty or nil product_id
		if itemsRaw, ok := rawMap["items"].([]interface{}); ok {
			var cleanedItems []interface{}
			for _, itemRaw := range itemsRaw {
				if itemMap, isMap := itemRaw.(map[string]interface{}); isMap {
					prodID, _ := itemMap["product_id"].(string)
					if prodID != "" && prodID != "00000000-0000-0000-0000-000000000000" {
						cleanedItems = append(cleanedItems, itemMap)
					}
				}
			}
			rawMap["items"] = cleanedItems
			changed = true
		}

		if changed {
			if newBytes, err := json.Marshal(rawMap); err == nil {
				c.Request.Body = io.NopCloser(bytes.NewBuffer(newBytes))
			}
		}
	}

	var req usecase.CreateTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fmt.Printf("DEBUG: CreateTransaction binding error: %v\nRaw Body: %s\n", err, string(bodyBytes))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format: " + err.Error()})
		return
	}

	fmt.Printf("[POST /sales/transactions] req_size=%dB client_id=%s receipt_no=%s method=%s store=%v items=%d\n",
		payloadSize, req.ClientID, req.ReceiptNo, req.PaymentMethod, req.StoreID, len(req.Items))

	// Auto-fill EmployeeID from token if empty or nil
	var salesmanName string
	if req.EmployeeID.String() == "00000000-0000-0000-0000-000000000000" || req.EmployeeID == uuid.Nil {
		if userIDStr, exists := c.Get("userID"); exists {
			userID := userIDStr.(uuid.UUID)
			employee, err := h.employeeUseCase.GetEmployeeByUserID(userID)
			if err == nil && employee != nil {
				req.EmployeeID = employee.ID
				if employee.CompanyID != nil && req.CompanyID == uuid.Nil {
					req.CompanyID = *employee.CompanyID
				}
				salesmanName = employee.FirstName + " " + employee.LastName
			}
		}
	}

	trx, err := h.salesUsecase.CreateTransaction(req)
	duration := time.Since(startTime)
	fmt.Printf("[POST /sales/transactions] completed in %v, client_id=%s\n", duration, req.ClientID)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "duplicate") {
			// Idempotent duplicate: instead of 409 Conflict which breaks old mobile clients,
			// respond with 200 OK so old mobile versions mark it as synced.
			c.JSON(http.StatusOK, gin.H{
				"message": "Transaksi sudah tercatat sebelumnya (idempotent duplicate)",
				"code":    "DUPLICATE_TRANSACTION",
				"data": gin.H{
					"receipt_no":     req.ReceiptNo,
					"payment_method": req.PaymentMethod,
				},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Jika transaksi memiliki foto nota fisik, publish SSE event ke admin
	// Note: untuk SALES_ORDER sudah ditrigger dari CreateSO
	if trx != nil && trx.ReceiptImageURL != "" && trx.PaymentMethod != "SALES_ORDER" && h.sseHub != nil {
		go func() {
			var targetBranchIDs []string
			if trx.BranchID != nil {
				targetBranchIDs = []string{trx.BranchID.String()}
			}
			event := sse.NotaFisikEvent{
				TransactionID:   trx.ID.String(),
				ReceiptNo:       trx.ReceiptNo,
				Salesman:        salesmanName,
				ImageURL:        trx.ReceiptImageURL,
				CompanyID:       trx.CompanyID.String(),
				CreatedAt:       trx.CreatedAt.Format("2006-01-02 15:04:05"),
				TargetBranchIDs: targetBranchIDs,
				TargetRoles: []string{
					string(domain.RoleSuperAdmin),
					string(domain.RoleCashier),
					string(domain.RoleCashierPusat),
					string(domain.RoleAdminKasir),
					string(domain.RoleManager),
					string(domain.RoleAdminNota),
				},
			}
			if trx.Store != nil {
				event.StoreName = trx.Store.Name
			}
			if pubErr := h.sseHub.PublishNotaFisik(context.Background(), event); pubErr != nil {
				fmt.Printf("[SSE] Failed to publish nota_fisik event: %v\n", pubErr)
			}
		}()
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Transaction created successfully",
		"data":    trx,
	})
}

func (h *SalesHandler) GetPendingTransactions(c *gin.Context) {
	var employeeID uuid.UUID
	employeeIDStr := c.Query("employee_id")

	if employeeIDStr != "" {
		parsed, err := uuid.Parse(employeeIDStr)
		if err == nil {
			employeeID = parsed
		}
	} else {
		// Resolve from token
		if userIDStr, exists := c.Get("userID"); exists {
			userID := userIDStr.(uuid.UUID)
			employee, err := h.employeeUseCase.GetEmployeeByUserID(userID)
			if err == nil && employee != nil {
				employeeID = employee.ID
			}
		}
	}

	if employeeID == uuid.Nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "employee_id is required or token must be associated with an employee"})
		return
	}

	pending, err := h.salesUsecase.GetPendingTransactions(employeeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Success",
		"data":    pending,
	})
}

func (h *SalesHandler) GetHistoryTransactions(c *gin.Context) {
	// --- Branch filter (used by cashier view) ---
	var branchID *uuid.UUID
	if branchIDStr := c.Query("branch_id"); branchIDStr != "" && branchIDStr != "ALL_BRANCHES" && branchIDStr != "ALL" {
		if parsed, err := uuid.Parse(branchIDStr); err == nil {
			branchID = &parsed
		}
	}

	var employeeID uuid.UUID
	employeeIDStr := c.Query("employee_id")

	if employeeIDStr != "" {
		parsed, err := uuid.Parse(employeeIDStr)
		if err == nil {
			employeeID = parsed
		}
	} else if branchID == nil {
		// Resolve from token only if no branch filter and no explicit employee_id
		if userIDStr, exists := c.Get("userID"); exists {
			userID := userIDStr.(uuid.UUID)
			employee, err := h.employeeUseCase.GetEmployeeByUserID(userID)
			if err == nil && employee != nil {
				employeeID = employee.ID
			}
		}
	}

	if employeeID == uuid.Nil && branchID == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "employee_id or branch_id is required, or token must be associated with an employee"})
		return
	}

	history, err := h.salesUsecase.GetHistoryTransactions(employeeID, branchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Success",
		"data":    history,
	})
}

func (h *SalesHandler) VerifyTransaction(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid transaction id"})
		return
	}

	var req usecase.VerifyTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var approverID *uuid.UUID
	approverInfo := GetApproverInfo(c, nil)
	if approverInfo != nil && approverInfo.EmployeeID != nil {
		approverID = approverInfo.EmployeeID
	} else if approverInfo != nil && approverInfo.UserID != uuid.Nil {
		employee, err := h.employeeUseCase.GetEmployeeByUserID(approverInfo.UserID)
		if err == nil && employee != nil {
			approverID = &employee.ID
		}
	}

	err = h.salesUsecase.VerifyTransaction(id, approverID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Transaction verified successfully",
	})
}

func (h *SalesHandler) VerifyPenerimaan(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid transaction id"})
		return
	}

	var req usecase.VerifyPenerimaanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Resolve Admin Nota ID (from auth context)
	var adminNotaID uuid.UUID
	if empID, exists := c.Get("employeeID"); exists {
		adminNotaID = empID.(uuid.UUID)
	}
	if adminNotaID == uuid.Nil {
		if userIDStr, exists := c.Get("userID"); exists {
			userID := userIDStr.(uuid.UUID)
			employee, err := h.employeeUseCase.GetEmployeeByUserID(userID)
			if err == nil && employee != nil {
				adminNotaID = employee.ID
			}
		}
	}

	err = h.salesUsecase.VerifyPenerimaan(id, adminNotaID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Penerimaan verified successfully",
	})
}

func (h *SalesHandler) GetPenerimaanTempoPayments(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	branchID := c.Query("branch_id")
	status := c.DefaultQuery("status", "PENDING")
	search := c.Query("search")
	salesman := c.Query("salesman")
	date := c.Query("date")

	payments, total, totalAmount, err := h.salesUsecase.GetPenerimaanTempoPayments(page, limit, branchID, status, search, salesman, date)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":         payments,
		"total":        total,
		"total_amount": totalAmount,
		"page":         page,
		"limit":        limit,
	})
}

type VerifyPenerimaanTempoPaymentRequest struct {
	Notes string `json:"notes"`
}

func (h *SalesHandler) VerifyPenerimaanTempoPayment(c *gin.Context) {
	paymentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payment ID"})
		return
	}

	adminID := uuid.Nil
	if empID, exists := c.Get("employeeID"); exists {
		if uid, ok := empID.(uuid.UUID); ok {
			adminID = uid
		}
	}
	if adminID == uuid.Nil {
		if adminUserID, exists := c.Get("userID"); exists {
			if uid, ok := adminUserID.(uuid.UUID); ok {
				emp, _ := h.employeeUseCase.GetEmployeeByUserID(uid)
				if emp != nil {
					adminID = emp.ID
				}
			}
		}
	}

	var req VerifyPenerimaanTempoPaymentRequest
	_ = c.ShouldBindJSON(&req)

	if err := h.salesUsecase.VerifyPenerimaanTempoPayment(paymentID, adminID, req.Notes); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Penerimaan pembayaran tempo berhasil diverifikasi dan omzet sales telah diperbarui",
	})
}

func (h *SalesHandler) RedeliverTransaction(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid transaction id"})
		return
	}
	var req struct {
		Reason string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil && err != io.EOF {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	// Resolve Admin Nota ID (from auth context)
	var adminNotaID uuid.UUID
	if empID, exists := c.Get("employeeID"); exists {
		adminNotaID = empID.(uuid.UUID)
	}
	if adminNotaID == uuid.Nil {
		if userIDStr, exists := c.Get("userID"); exists {
			userID := userIDStr.(uuid.UUID)
			employee, err := h.employeeUseCase.GetEmployeeByUserID(userID)
			if err == nil && employee != nil {
				adminNotaID = employee.ID
			}
		}
	}

	err = h.salesUsecase.RedeliverTransaction(id, adminNotaID, strings.TrimSpace(req.Reason))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Faktur berhasil dibatalkan dan dikembalikan ke tahap Surat Jalan",
	})
}

func (h *SalesHandler) CancelFailedDeliveryTransaction(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID faktur penjualan tidak valid"})
		return
	}
	var req struct {
		Reason string `json:"reason" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.Reason) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "alasan pembatalan wajib diisi"})
		return
	}

	existingTrx, _ := h.salesUsecase.GetTransactionByID(id)
	if err := h.salesUsecase.CancelFailedDeliveryTransaction(id, req.Reason); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var userID, companyID uuid.UUID
	var userName string
	if value, exists := c.Get("userID"); exists {
		userID = value.(uuid.UUID)
		if employee, err := h.employeeUseCase.GetEmployeeByUserID(userID); err == nil && employee != nil {
			userName = employee.FirstName + " " + employee.LastName
			if employee.CompanyID != nil {
				companyID = *employee.CompanyID
			}
		}
	}
	updatedTrx, _ := h.salesUsecase.GetTransactionByID(id)
	h.auditLogUseCase.Log(companyID, userID, userName, "CANCEL_FAILED_DELIVERY", "SalesTransaction", id.String(), "Faktur gagal kirim dibatalkan: "+strings.TrimSpace(req.Reason), existingTrx, updatedTrx, c.ClientIP())
	c.JSON(http.StatusOK, gin.H{"message": "Faktur gagal kirim berhasil dibatalkan"})
}

func (h *SalesHandler) UpdateTransaction(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid transaction id"})
		return
	}

	var req usecase.ManualEntryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var userName string
	var companyID uuid.UUID
	var userID uuid.UUID
	if userIDStr, exists := c.Get("userID"); exists {
		userID = userIDStr.(uuid.UUID)
		employee, err := h.employeeUseCase.GetEmployeeByUserID(userID)
		if err == nil && employee != nil {
			userName = employee.FirstName + " " + employee.LastName
			if employee.CompanyID != nil {
				companyID = *employee.CompanyID
			}
		}
	}
	existingTx, _ := h.salesUsecase.GetTransactionByID(id)
	if role, _ := c.Get("role"); role == string(domain.RoleAdminNota) && existingTx != nil {
		req.PaidAmount = existingTx.PaidAmount
		req.PaymentStatus = existingTx.PaymentStatus
		if existingTx.PaidAmount > 0 {
			req.PaymentMethod = existingTx.PaymentMethod
		}
		req.DeferPaymentToCashier = true
	}

	// Keep old state
	var oldTx domain.SalesTransaction
	if existingTx != nil {
		oldTxBytes, _ := json.Marshal(existingTx)
		_ = json.Unmarshal(oldTxBytes, &oldTx)
	}

	err = h.salesUsecase.UpdateTransaction(id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	newTx, _ := h.salesUsecase.GetTransactionByID(id)
	h.auditLogUseCase.Log(companyID, userID, userName, "UPDATE", "SalesTransaction", id.String(), "Updated Sales Transaction: "+newTx.ReceiptNo, oldTx, newTx, c.ClientIP())

	c.JSON(http.StatusOK, gin.H{"message": "Transaction updated successfully"})
}

func (h *SalesHandler) UpdateTransactionMobile(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid transaction id"})
		return
	}

	var req usecase.MobileUpdateTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var userName string
	var companyID uuid.UUID
	var userID uuid.UUID
	if userIDStr, exists := c.Get("userID"); exists {
		userID = userIDStr.(uuid.UUID)
		employee, err := h.employeeUseCase.GetEmployeeByUserID(userID)
		if err == nil && employee != nil {
			userName = employee.FirstName + " " + employee.LastName
			if employee.CompanyID != nil {
				companyID = *employee.CompanyID
			}
		}
	}
	existingTx, _ := h.salesUsecase.GetTransactionByID(id)

	// Keep old state for audit
	var oldTx domain.SalesTransaction
	if existingTx != nil {
		oldTxBytes, _ := json.Marshal(existingTx)
		_ = json.Unmarshal(oldTxBytes, &oldTx)
	}

	err = h.salesUsecase.UpdateTransactionMobile(id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	newTx, _ := h.salesUsecase.GetTransactionByID(id)
	h.auditLogUseCase.Log(companyID, userID, userName, "UPDATE", "SalesTransaction", id.String(), "Updated Sales Transaction Mobile: "+newTx.ReceiptNo, oldTx, newTx, c.ClientIP())

	c.JSON(http.StatusOK, gin.H{"message": "Transaction updated successfully"})
}

func (h *SalesHandler) DeleteTransaction(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid transaction id"})
		return
	}

	var req struct {
		Reason string `json:"reason" form:"reason"`
	}
	_ = c.ShouldBind(&req)

	reason := strings.TrimSpace(req.Reason)
	if reason == "" {
		reason = c.Query("reason")
	}
	if reason == "" {
		reason = "Dihapus oleh admin/user"
	}

	var userName string
	var companyID uuid.UUID
	var userID uuid.UUID
	var empID uuid.UUID
	if userIDStr, exists := c.Get("userID"); exists {
		userID = userIDStr.(uuid.UUID)
		employee, err := h.employeeUseCase.GetEmployeeByUserID(userID)
		if err == nil && employee != nil {
			empID = employee.ID
			userName = employee.FirstName + " " + employee.LastName
			if employee.CompanyID != nil {
				companyID = *employee.CompanyID
			}
		}
	}
	existingTx, _ := h.salesUsecase.GetTransactionByID(id)

	err = h.salesUsecase.SoftDeleteTransaction(id, empID, reason)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if existingTx != nil {
		h.auditLogUseCase.Log(companyID, userID, userName, "DELETE", "SalesTransaction", id.String(), "Deleted Sales Transaction: "+existingTx.ReceiptNo+" (Alasan: "+reason+")", existingTx, nil, c.ClientIP())
	}

	c.JSON(http.StatusOK, gin.H{"message": "Faktur penjualan berhasil dihapus"})
}

func (h *SalesHandler) DeletePendingTransactionMobile(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID transaksi tidak valid"})
		return
	}
	userValue, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Identitas sales tidak ditemukan"})
		return
	}
	userID := userValue.(uuid.UUID)
	employee, err := h.employeeUseCase.GetEmployeeByUserID(userID)
	if err != nil || employee == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Identitas sales tidak ditemukan"})
		return
	}
	existing, _ := h.salesUsecase.GetTransactionByID(id)
	if err := h.salesUsecase.DeletePendingTransactionBySales(id, employee.ID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if existing != nil {
		companyID := uuid.Nil
		if employee.CompanyID != nil {
			companyID = *employee.CompanyID
		}
		h.auditLogUseCase.Log(companyID, userID, employee.FirstName+" "+employee.LastName,
			"DELETE", "SalesTransaction", id.String(), "Sales deleted pending transaction: "+existing.ReceiptNo,
			existing, nil, c.ClientIP())
	}
	c.JSON(http.StatusOK, gin.H{"message": "Transaksi menunggu validasi berhasil dihapus"})
}

func (h *SalesHandler) RestoreTransaction(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid transaction id"})
		return
	}

	var userName string
	var companyID uuid.UUID
	var userID uuid.UUID
	if userIDStr, exists := c.Get("userID"); exists {
		userID = userIDStr.(uuid.UUID)
		employee, err := h.employeeUseCase.GetEmployeeByUserID(userID)
		if err == nil && employee != nil {
			userName = employee.FirstName + " " + employee.LastName
			if employee.CompanyID != nil {
				companyID = *employee.CompanyID
			}
		}
	}

	err = h.salesUsecase.RestoreTransaction(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.auditLogUseCase.Log(companyID, userID, userName, "RESTORE", "SalesTransaction", id.String(), "Restored Sales Transaction: "+id.String(), nil, nil, c.ClientIP())

	c.JSON(http.StatusOK, gin.H{"message": "Faktur penjualan berhasil dipulihkan"})
}

func (h *SalesHandler) ConvertToSalesOrder(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID faktur penjualan tidak valid"})
		return
	}

	var adminID uuid.UUID
	var userName string
	var companyID uuid.UUID
	if userIDStr, exists := c.Get("userID"); exists {
		userID := userIDStr.(uuid.UUID)
		employee, err := h.employeeUseCase.GetEmployeeByUserID(userID)
		if err == nil && employee != nil {
			adminID = employee.ID
			userName = employee.FirstName + " " + employee.LastName
			if employee.CompanyID != nil {
				companyID = *employee.CompanyID
			}
		}
	}

	newSO, err := h.salesUsecase.ConvertToSalesOrder(id, adminID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.auditLogUseCase.Log(companyID, adminID, userName, "CONVERT_TO_SO", "SalesTransaction", id.String(), fmt.Sprintf("Faktur penjualan dikonversi menjadi Pesanan SO: %s", newSO.SONumber), nil, newSO, c.ClientIP())

	c.JSON(http.StatusOK, gin.H{
		"message": "Faktur penjualan berhasil diubah menjadi Pesanan SO",
		"data":    newSO,
	})
}

func (h *SalesHandler) SetKPITarget(c *gin.Context) {
	var req struct {
		EmployeeID             uuid.UUID `json:"employee_id" binding:"required"`
		Month                  int       `json:"month" binding:"required"`
		Year                   int       `json:"year" binding:"required"`
		TargetOmzet            float64   `json:"target_omzet"`
		TargetNewStores        int       `json:"target_new_stores"`
		TargetDailyVisits      int       `json:"target_daily_visits"`
		TargetDailyNewStores   int       `json:"target_daily_new_stores"`
		WorkingTerritory       string    `json:"working_territory"`
		ReceivableLimitPercent float64   `json:"receivable_limit_percent"`
		MaxReceivableLimit     float64   `json:"max_receivable_limit"`
		Commission             float64   `json:"commission"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.salesUsecase.SetKPITarget(
		req.EmployeeID,
		req.Month,
		req.Year,
		req.TargetOmzet,
		req.TargetNewStores,
		req.TargetDailyVisits,
		req.TargetDailyNewStores,
		req.WorkingTerritory,
		req.ReceivableLimitPercent,
		req.MaxReceivableLimit,
		req.Commission,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "KPI target set successfully"})
}

func (h *SalesHandler) DeleteKPITarget(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid target id"})
		return
	}

	err = h.salesUsecase.DeleteKPITarget(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "KPI target deleted successfully"})
}

func (h *SalesHandler) UploadPhoto(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file is received"})
		return
	}

	// Save and auto-compress image to JPEG with 75% quality and max 1024px dimension
	fileURL, err := utils.SaveUploadedImageCompressed(file, "./uploads/sales", 1024, 75)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to save the file: " + err.Error()})
		return
	}

	if strings.HasPrefix(fileURL, "/uploads/") {
		fileURL = "/api/v1" + fileURL
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "File uploaded successfully",
		"url":     fileURL,
	})
}

func (h *SalesHandler) GetVisitPlan(c *gin.Context) {
	var employeeID uuid.UUID
	if userIDStr, exists := c.Get("userID"); exists {
		userID := userIDStr.(uuid.UUID)
		employee, err := h.employeeUseCase.GetEmployeeByUserID(userID)
		if err == nil && employee != nil {
			employeeID = employee.ID
		}
	}

	if employeeID == uuid.Nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unauthorized employee"})
		return
	}

	dateStr := c.Query("date")
	loc, _ := time.LoadLocation("Asia/Jakarta")
	date := time.Now().In(loc)
	if dateStr != "" {
		if parsed, err := time.Parse("2006-01-02", dateStr); err == nil {
			date = parsed.In(loc)
		}
	}

	plan, err := h.salesUsecase.GetVisitPlan(employeeID, date)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Success",
		"data":    plan,
	})
}

// GetVisitPlanBuilder fetches all stores with history to build a plan
func (h *SalesHandler) GetVisitPlanBuilder(c *gin.Context) {
	var employeeID uuid.UUID
	var currentEmployee *domain.Employee
	if userIDStr, exists := c.Get("userID"); exists {
		userID := userIDStr.(uuid.UUID)
		employee, err := h.employeeUseCase.GetEmployeeByUserID(userID)
		if err == nil && employee != nil {
			employeeID = employee.ID
			currentEmployee = employee
		}
	}

	if employeeID == uuid.Nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unauthorized employee"})
		return
	}

	var targetSalesmanID *uuid.UUID
	if salesmanStr := c.Query("salesman_id"); salesmanStr != "" {
		if parsed, err := uuid.Parse(salesmanStr); err == nil {
			targetSalesmanID = &parsed
		}
	}

	var branchID *uuid.UUID
	if branchStr := c.Query("branch_id"); branchStr != "" && branchStr != "ALL_BRANCHES" {
		if parsed, err := uuid.Parse(branchStr); err == nil {
			branchID = &parsed
		}
	} else if branchStr == "" && currentEmployee != nil && currentEmployee.BranchID != nil {
		branchID = currentEmployee.BranchID
	}

	plans, err := h.salesUsecase.GetVisitPlanBuilder(employeeID, targetSalesmanID, branchID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"message": "Visit plan builder data fetched successfully",
		"data":    plans,
	})
}

func (h *SalesHandler) GetByReceipt(c *gin.Context) {
	receiptNo := c.Param("receipt_no")
	if receiptNo == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "receipt_no is required"})
		return
	}

	trx, err := h.salesUsecase.GetTransactionByReceipt(receiptNo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if trx == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Transaction not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Success",
		"data":    trx,
	})
}

func (h *SalesHandler) RecordAdminPayment(c *gin.Context) {
	trxID := uuid.MustParse(c.Param("id"))
	if role, _ := c.Get("role"); role == string(domain.RoleAdminNota) {
		trx, err := h.salesUsecase.GetTransactionByID(trxID)
		if err != nil || trx == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Faktur penjualan tidak ditemukan"})
			return
		}
		if trx.PaymentMethod != "TEMPO" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Admin Nota hanya dapat mencatat pembayaran faktur tempo"})
			return
		}
	}
	h.recordPayment(c, trxID)
}

func (h *SalesHandler) RecordPayment(c *gin.Context) {
	if role, _ := c.Get("role"); role == string(domain.RoleAdminNota) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Pembayaran faktur dicatat oleh Admin Kasir melalui menu Keuangan"})
		return
	}
	h.recordPayment(c, uuid.MustParse(c.Param("id")))
}

func (h *SalesHandler) recordPayment(c *gin.Context, trxID uuid.UUID) {
	var req usecase.RecordPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.TransactionID = trxID

	// Resolve EmployeeID from token if empty
	if req.EmployeeID == uuid.Nil {
		if userIDStr, exists := c.Get("userID"); exists {
			userID := userIDStr.(uuid.UUID)
			employee, err := h.employeeUseCase.GetEmployeeByUserID(userID)
			if err == nil && employee != nil {
				req.EmployeeID = employee.ID
			}
		}
	}

	if err := h.salesUsecase.RecordPayment(req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Pembayaran berhasil dicatat"})
}

func (h *SalesHandler) GetOutstandingByStore(c *gin.Context) {
	storeID := uuid.MustParse(c.Param("id"))

	trxs, err := h.salesUsecase.GetOutstandingTransactionsByStore(storeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Success",
		"data":    trxs,
	})
}

func (h *SalesHandler) GetByDueDate(c *gin.Context) {
	dateStr := c.Query("date")

	var employeeID uuid.UUID
	if userIDStr, exists := c.Get("userID"); exists {
		userID := userIDStr.(uuid.UUID)
		employee, err := h.employeeUseCase.GetEmployeeByUserID(userID)
		if err == nil && employee != nil {
			employeeID = employee.ID
		}
	}

	var trxs []domain.SalesTransaction
	var err error

	if dateStr == "" || dateStr == "all" {
		trxs, err = h.salesUsecase.GetTransactionsByDueDate(time.Time{}, employeeID)
	} else {
		var date time.Time
		if parsed, err := time.Parse("2006-01-02", dateStr); err == nil {
			date = parsed
		} else {
			date = time.Now()
		}
		trxs, err = h.salesUsecase.GetTransactionsByDueDate(date, employeeID)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Success",
		"data":    trxs,
	})
}
func (h *SalesHandler) GetTransactions(c *gin.Context) {
	status := c.Param("status")
	if status == "" {
		status = c.Query("status")
	}
	companyIDStr := c.Query("company_id")

	var companyID uuid.UUID
	if companyIDStr != "" {
		parsed, err := uuid.Parse(companyIDStr)
		if err == nil {
			companyID = parsed
		}
	}

	var trxs []domain.SalesTransaction
	var err error

	if status == "VERIFIED" {
		trxs, err = h.salesUsecase.GetDeliveryPending(companyID)
	} else {
		trxs, err = h.salesUsecase.GetTransactionsByStatusAndCompany(status, companyID)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Success",
		"data":    trxs,
	})
}

func (h *SalesHandler) RejectTransaction(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid transaction id"})
		return
	}

	var req struct {
		Notes string `json:"notes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.salesUsecase.RejectTransaction(id, req.Notes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Transaction rejected successfully",
	})
}

func (h *SalesHandler) GetNextReceiptNo(c *gin.Context) {
	fmt.Println("DEBUG: GetNextReceiptNo hit")
	companyIDStr := c.Query("company_id")
	companyID, _ := uuid.Parse(companyIDStr)
	dateStr := c.Query("date")

	var date time.Time
	if dateStr != "" {
		date, _ = time.Parse("2006-01-02", dateStr)
	} else {
		date = time.Now()
	}

	count, err := h.salesUsecase.CountByCompanyAndDate(companyID, date)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	receiptNo := fmt.Sprintf("INV/%s/%03d", date.Format("20060102"), count+1)
	c.JSON(http.StatusOK, gin.H{"receipt_no": receiptNo})
}

func (h *SalesHandler) HandleMidtransNotification(c *gin.Context) {
	var notificationPayload map[string]interface{}
	if err := c.ShouldBindJSON(&notificationPayload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.salesUsecase.HandleMidtransNotification(notificationPayload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Notification handled successfully"})
}

func (h *SalesHandler) GetProductDistribution(c *gin.Context) {
	productIDStr := c.Query("product_id")
	productIDsStr := c.Query("product_ids")

	var productIDs []uuid.UUID
	if productIDsStr != "" {
		for _, s := range strings.Split(productIDsStr, ",") {
			s = strings.TrimSpace(s)
			if id, err := uuid.Parse(s); err == nil && id != uuid.Nil {
				productIDs = append(productIDs, id)
			}
		}
	} else if productIDStr != "" && productIDStr != "ALL" {
		for _, s := range strings.Split(productIDStr, ",") {
			s = strings.TrimSpace(s)
			if id, err := uuid.Parse(s); err == nil && id != uuid.Nil {
				productIDs = append(productIDs, id)
			}
		}
	}

	if len(productIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "product_id or product_ids is required"})
		return
	}

	companyIDStr := c.Query("company_id")
	companyID := uuid.Nil
	if companyIDStr != "" {
		companyID, _ = uuid.Parse(companyIDStr)
	}

	distribution, err := h.salesUsecase.GetProductSalesDistribution(productIDs, companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Success retrieving product sales distribution",
		"data":    distribution,
	})
}

func (h *SalesHandler) GetTopProducts(c *gin.Context) {
	companyIDStr := c.Query("company_id")
	var companyID uuid.UUID
	if companyIDStr != "" {
		parsed, err := uuid.Parse(companyIDStr)
		if err == nil {
			companyID = parsed
		}
	}

	products, err := h.salesUsecase.GetTopProducts(companyID, 10)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Success retrieving top products",
		"data":    products,
	})
}

type CreateManualVisitPlanRequest struct {
	StoreIDs      []uuid.UUID                `json:"store_ids"`
	StoreIDsCamel []uuid.UUID                `json:"storeIds"`
	Stores        []usecase.StorePlanRequest `json:"stores"`
	PlanDate      time.Time                  `json:"plan_date" binding:"required"`
}

func (h *SalesHandler) CreateManualVisitPlan(c *gin.Context) {
	var employeeID uuid.UUID
	if userIDStr, exists := c.Get("userID"); exists {
		userID := userIDStr.(uuid.UUID)
		employee, err := h.employeeUseCase.GetEmployeeByUserID(userID)
		if err == nil && employee != nil {
			employeeID = employee.ID
		}
	}

	if employeeID == uuid.Nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unauthorized employee"})
		return
	}

	var req CreateManualVisitPlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if len(req.StoreIDsCamel) > 0 {
		req.StoreIDs = append(req.StoreIDs, req.StoreIDsCamel...)
	}

	storesToPlan := req.Stores
	if len(storesToPlan) == 0 && len(req.StoreIDs) > 0 {
		for _, id := range req.StoreIDs {
			storesToPlan = append(storesToPlan, usecase.StorePlanRequest{StoreID: id})
		}
	}

	if len(storesToPlan) == 0 {
		c.JSON(400, gin.H{"error": "store_ids or stores must be provided"})
		return
	}

	if err := h.salesUsecase.CreateManualVisitPlan(employeeID, storesToPlan, req.PlanDate); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "Manual visit plan created successfully"})
}

func (h *SalesHandler) RegeneratePayment(c *gin.Context) {
	transactionID := c.Param("id")
	if transactionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "transaction ID required"})
		return
	}

	trx, err := h.salesUsecase.RegeneratePayment(transactionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, trx)
}

// QrisImageProxy proxies the Midtrans QR code image with proper Basic Auth.
// Mobile apps call this endpoint instead of calling Midtrans directly (which requires server key auth).
// Usage: GET /api/v1/sales/qris-image?url=https://api.sandbox.midtrans.com/v2/qris/{id}/qr-code
func (h *SalesHandler) QrisImageProxy(c *gin.Context) {
	qrURL := c.Query("url")
	if qrURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "url parameter is required"})
		return
	}

	serverKey := os.Getenv("MIDTRANS_SERVER_KEY")
	if serverKey == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server key not configured"})
		return
	}

	// Midtrans uses Basic Auth: base64(serverKey + ":")
	credentials := base64.StdEncoding.EncodeToString([]byte(serverKey + ":"))

	req, err := http.NewRequest("GET", qrURL, nil)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid url"})
		return
	}
	req.Header.Set("Authorization", "Basic "+credentials)
	req.Header.Set("Accept", "image/png")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "failed to fetch QR image: " + err.Error()})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusBadGateway, gin.H{"error": fmt.Sprintf("midtrans returned status %d", resp.StatusCode)})
		return
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "image/png"
	}

	c.Header("Content-Type", contentType)
	c.Header("Cache-Control", "public, max-age=300")
	c.Status(http.StatusOK)
	io.Copy(c.Writer, resp.Body)
}

func (h *SalesHandler) GetAdminVisitPlans(c *gin.Context) {
	employeeIDStr := c.Query("employee_id")
	if employeeIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "employee_id is required"})
		return
	}

	employeeID, err := uuid.Parse(employeeIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid employee_id"})
		return
	}

	dateStr := c.Query("date")
	loc, _ := time.LoadLocation("Asia/Jakarta")
	date := time.Now().In(loc)
	if dateStr != "" {
		if parsed, err := time.Parse("2006-01-02", dateStr); err == nil {
			date = parsed.In(loc)
		}
	}

	plan, err := h.salesUsecase.GetVisitPlan(employeeID, date)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Success",
		"data":    plan,
	})
}

func (h *SalesHandler) DeleteManualVisitPlan(c *gin.Context) {
	var employeeID uuid.UUID
	if userIDStr, exists := c.Get("userID"); exists {
		userID := userIDStr.(uuid.UUID)
		employee, err := h.employeeUseCase.GetEmployeeByUserID(userID)
		if err == nil && employee != nil {
			employeeID = employee.ID
		}
	}

	if employeeID == uuid.Nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unauthorized employee"})
		return
	}

	storeIDStr := c.Param("store_id")
	storeID, err := uuid.Parse(storeIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid store_id"})
		return
	}

	dateStr := c.Query("date")
	date := time.Now()
	if dateStr != "" {
		if parsed, err := time.Parse("2006-01-02", dateStr); err == nil {
			date = parsed
		}
	}

	if err := h.salesUsecase.DeleteSpecificPlan(employeeID, storeID, date); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Visit plan deleted successfully"})
}

func (h *SalesHandler) UpdateLiveLocation(c *gin.Context) {
	var req struct {
		Latitude   float64    `json:"latitude" binding:"required"`
		Longitude  float64    `json:"longitude" binding:"required"`
		Accuracy   float64    `json:"accuracy"`
		RecordedAt *time.Time `json:"recorded_at"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	if req.Latitude < -90 || req.Latitude > 90 || req.Longitude < -180 || req.Longitude > 180 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid latitude or longitude"})
		return
	}
	if req.Accuracy < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid location accuracy"})
		return
	}

	employeeIDVal, exists := c.Get("employeeID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Employee not identified"})
		return
	}
	employeeID := employeeIDVal.(uuid.UUID)

	isOnDuty := false
	if h.attendanceRepo != nil {
		attendance, _ := h.attendanceRepo.FindTodayByUserID(employeeID)
		if attendance == nil {
			latest, latestErr := h.attendanceRepo.FindLatestByEmployeeID(employeeID)
			if latestErr == nil && latest != nil && latest.ClockInTime != nil && latest.ClockOutTime == nil && time.Since(*latest.ClockInTime) < 24*time.Hour {
				attendance = latest
			}
		}
		if attendance != nil && attendance.ClockInTime != nil && attendance.ClockOutTime == nil {
			isOnDuty = true
		}
	}

	companyIDVal, exists := c.Get("companyID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Company not identified"})
		return
	}
	companyID := companyIDVal.(uuid.UUID)

	employee, err := h.employeeUseCase.GetEmployeeByID(employeeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch employee details"})
		return
	}
	if !isLiveMonitoringEligible(employee) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Live location is only available for Field Staff (Sales, Marketing, Driver)"})
		return
	}

	jobTitle := ""
	if employee.JobPosition != nil {
		jobTitle = employee.JobPosition.Title
	}
	deptName := ""
	if employee.Department != nil {
		deptName = employee.Department.Name
	}
	branchIDStr := ""
	branchName := ""
	if employee.BranchID != nil {
		branchIDStr = employee.BranchID.String()
	}
	if employee.Branch != nil {
		branchName = employee.Branch.Name
	}
	operatingBranchIDStr := ""
	if employee.OperatingBranchID != nil {
		operatingBranchIDStr = employee.OperatingBranchID.String()
	}

	updatedAt := time.Now().UTC()
	if req.RecordedAt != nil {
		updatedAt = req.RecordedAt.UTC()
	}
	payload := map[string]interface{}{
		"type":                "salesman_location",
		"employee_id":         employeeID.String(),
		"employee_name":       employee.FirstName + " " + employee.LastName,
		"job_title":           jobTitle,
		"department":          deptName,
		"category":            getStaffCategory(employee),
		"branch_id":           branchIDStr,
		"operating_branch_id": operatingBranchIDStr,
		"branch_name":         branchName,
		"latitude":            req.Latitude,
		"longitude":           req.Longitude,
		"accuracy":            req.Accuracy,
		"is_on_duty":          isOnDuty,
		"company_id":          companyID.String(),
		"recorded_at":         updatedAt.Format(time.RFC3339),
		"received_at":         time.Now().UTC().Format(time.RFC3339),
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to serialize payload"})
		return
	}

	ctx := context.Background()

	// 1. Save in String key with 24-hour TTL (so location is never wiped prematurely)
	if redisclient.Client != nil {
		redisKey := fmt.Sprintf("salesman:location:%s:%s", companyID.String(), employeeID.String())
		_ = redisclient.Client.Set(ctx, redisKey, payloadBytes, 24*time.Hour).Err()

		// 2. Geospatial Index (GEOADD) for proximity queries
		geoKey := fmt.Sprintf("salesman:geo:%s", companyID.String())
		_ = redisclient.Client.GeoAdd(ctx, geoKey, &redis.GeoLocation{
			Name:      employeeID.String(),
			Longitude: req.Longitude,
			Latitude:  req.Latitude,
		}).Err()

		// 3. Publish to channel for SSE real-time streaming
		_ = redisclient.Client.Publish(ctx, "salesman_location_updates", payloadBytes).Err()
	}

	// 4. Persist to PostgreSQL database
	if h.db != nil {
		liveLoc := domain.SalesmanLiveLocation{
			EmployeeID: employeeID,
			CompanyID:  companyID,
			BranchID:   employee.BranchID,
			Latitude:   req.Latitude,
			Longitude:  req.Longitude,
			Accuracy:   req.Accuracy,
			IsOnDuty:   isOnDuty,
			RecordedAt: updatedAt,
			UpdatedAt:  time.Now().UTC(),
		}
		_ = h.db.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "employee_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"latitude", "longitude", "accuracy", "is_on_duty", "recorded_at", "updated_at", "branch_id", "company_id"}),
		}).Create(&liveLoc).Error
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Location updated successfully"})
}

func (h *SalesHandler) GetLiveLocations(c *gin.Context) {
	companyIDVal, exists := c.Get("companyID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Company not identified"})
		return
	}
	companyID := companyIDVal.(uuid.UUID)
	branchFilter := c.Query("branch_id")

	ctx := context.Background()
	pattern := fmt.Sprintf("salesman:location:%s:*", companyID.String())

	locationsMap := make(map[string]map[string]interface{})

	wib := time.FixedZone("WIB", 7*3600)
	nowWIB := time.Now().In(wib)
	startOfDayWIB := time.Date(nowWIB.Year(), nowWIB.Month(), nowWIB.Day(), 0, 0, 0, 0, wib).UTC()

	// 1. Try fetching from Redis first
	if redisclient.Client != nil {
		var keys []string
		var cursor uint64
		for {
			batch, nextCursor, scanErr := redisclient.Client.Scan(ctx, cursor, pattern, 100).Result()
			if scanErr != nil {
				break
			}
			keys = append(keys, batch...)
			cursor = nextCursor
			if cursor == 0 {
				break
			}
		}

		if len(keys) > 0 {
			values, err := redisclient.Client.MGet(ctx, keys...).Result()
			if err == nil {
				for _, val := range values {
					if val == nil {
						continue
					}
					var loc map[string]interface{}
					if strVal, ok := val.(string); ok {
						if json.Unmarshal([]byte(strVal), &loc) == nil {
							if empID, ok := loc["employee_id"].(string); ok && empID != "" {
								// In-memory branch filtering
								if branchFilter != "" && branchFilter != "ALL" && branchFilter != "ALL_BRANCHES" {
									locBranch := fmt.Sprint(loc["branch_id"])
									locOpBranch := fmt.Sprint(loc["operating_branch_id"])
									if locBranch != branchFilter && locOpBranch != branchFilter {
										continue
									}
								}

								// Check if recorded_at or received_at is from today (WIB)
								var locTime time.Time
								if recStr, ok := loc["recorded_at"].(string); ok && recStr != "" {
									if t, parseErr := time.Parse(time.RFC3339, recStr); parseErr == nil {
										locTime = t
									}
								}
								if locTime.IsZero() {
									if recStr, ok := loc["received_at"].(string); ok && recStr != "" {
										if t, parseErr := time.Parse(time.RFC3339, recStr); parseErr == nil {
											locTime = t
										}
									}
								}

								// If location is from previous days, discard it from live map
								if !locTime.IsZero() && locTime.Before(startOfDayWIB) {
									continue
								}

								// Calculate freshness metadata
								lastSeenSeconds := 0
								if !locTime.IsZero() {
									lastSeenSeconds = int(time.Since(locTime).Seconds())
									if lastSeenSeconds < 0 {
										lastSeenSeconds = 0
									}
								}

								isOnDuty, _ := loc["is_on_duty"].(bool)
								freshnessTier := "STALE"
								if !isOnDuty {
									freshnessTier = "OFFLINE"
								} else if lastSeenSeconds < 600 {
									freshnessTier = "LIVE"
								} else if lastSeenSeconds < 3600 {
									freshnessTier = "STANDBY"
								}

								loc["last_seen_seconds"] = lastSeenSeconds
								loc["is_today"] = true
								loc["freshness_tier"] = freshnessTier

								locationsMap[empID] = loc
							}
						}
					}
				}
			}
		}
	}

	// 2. Query PostgreSQL database for today's locations to ensure complete coverage even if Redis is cold/expired
	if h.db != nil {
		var dbLocs []domain.SalesmanLiveLocation
		query := h.db.Preload("Employee.JobPosition").
			Preload("Employee.Department").
			Preload("Employee.Branch").
			Preload("Employee.OperatingBranch").
			Joins("LEFT JOIN employees e_loc ON e_loc.id = salesman_live_locations.employee_id").
			Where("salesman_live_locations.company_id = ? AND salesman_live_locations.updated_at >= ?", companyID, startOfDayWIB)

		if branchFilter != "" && branchFilter != "ALL" && branchFilter != "ALL_BRANCHES" {
			if bid, err := uuid.Parse(branchFilter); err == nil {
				query = query.Where("(salesman_live_locations.branch_id = ? OR e_loc.branch_id = ? OR e_loc.operating_branch_id = ?)", bid, bid, bid)
			}
		}

		if err := query.Find(&dbLocs).Error; err == nil {
			for _, dLoc := range dbLocs {
				empIDStr := dLoc.EmployeeID.String()
				if _, exists := locationsMap[empIDStr]; !exists {
					// Validate time is from today
					locTime := dLoc.RecordedAt
					if locTime.IsZero() {
						locTime = dLoc.UpdatedAt
					}
					if locTime.Before(startOfDayWIB) {
						continue
					}

					empName := "Petugas Lapangan"
					jobTitle := ""
					deptName := ""
					category := "FIELD_STAFF"
					branchIDStr := ""
					branchName := ""

					if dLoc.Employee != nil {
						empName = strings.TrimSpace(dLoc.Employee.FirstName + " " + dLoc.Employee.LastName)
						if dLoc.Employee.JobPosition != nil {
							jobTitle = dLoc.Employee.JobPosition.Title
						}
						if dLoc.Employee.Department != nil {
							deptName = dLoc.Employee.Department.Name
						}
						category = getStaffCategory(dLoc.Employee)
						if dLoc.Employee.Branch != nil {
							branchName = dLoc.Employee.Branch.Name
						}
					}
					if dLoc.BranchID != nil {
						branchIDStr = dLoc.BranchID.String()
					}

					lastSeenSeconds := int(time.Since(locTime).Seconds())
					if lastSeenSeconds < 0 {
						lastSeenSeconds = 0
					}

					freshnessTier := "STALE"
					if !dLoc.IsOnDuty {
						freshnessTier = "OFFLINE"
					} else if lastSeenSeconds < 600 {
						freshnessTier = "LIVE"
					} else if lastSeenSeconds < 3600 {
						freshnessTier = "STANDBY"
					}

					locationsMap[empIDStr] = map[string]interface{}{
						"type":              "salesman_location",
						"employee_id":       empIDStr,
						"employee_name":     empName,
						"job_title":         jobTitle,
						"department":        deptName,
						"category":          category,
						"branch_id":         branchIDStr,
						"branch_name":       branchName,
						"latitude":          dLoc.Latitude,
						"longitude":         dLoc.Longitude,
						"accuracy":          dLoc.Accuracy,
						"is_on_duty":        dLoc.IsOnDuty,
						"company_id":        companyID.String(),
						"recorded_at":       dLoc.RecordedAt.Format(time.RFC3339),
						"received_at":       dLoc.UpdatedAt.Format(time.RFC3339),
						"last_seen_seconds": lastSeenSeconds,
						"is_today":          true,
						"freshness_tier":    freshnessTier,
					}
				}
			}
		}
	}

	locations := make([]map[string]interface{}, 0, len(locationsMap))
	for _, loc := range locationsMap {
		locations = append(locations, loc)
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "data": locations})
}

// isLiveMonitoringEligible restricts live tracking to Field Staff (Sales, Marketing, Driver).
func isLiveMonitoringEligible(employee *domain.Employee) bool {
	if employee == nil {
		return false
	}
	var deptName, posName string
	if employee.Department != nil {
		deptName = strings.ToLower(strings.TrimSpace(employee.Department.Name))
	}
	if employee.JobPosition != nil {
		posName = strings.ToLower(strings.TrimSpace(employee.JobPosition.Title))
	}

	// 1. Explicit blacklist check for office/internal-only roles
	blacklist := []string{
		"accounting", "finance", "keuangan", "kasir", "cashier", "pabrik", "factory",
		"operator mesin", "security", "satpam", "cleaning", "ob", "hrd", "legal",
		"direktur", "director", "owner",
	}
	for _, b := range blacklist {
		if strings.Contains(posName, b) || strings.Contains(deptName, b) {
			return false
		}
	}

	// 2. Check Driver keywords
	isDriver := strings.Contains(posName, "driver") || strings.Contains(posName, "supir") ||
		strings.Contains(posName, "sopir") || strings.Contains(posName, "pengemudi") ||
		strings.Contains(posName, "kurir") || strings.Contains(posName, "delivery") ||
		strings.Contains(deptName, "driver") || strings.Contains(deptName, "logistik") ||
		strings.Contains(deptName, "delivery") || strings.Contains(deptName, "pengiriman")

	// 3. Check Sales & Marketing keywords
	isSalesOrMarketing := strings.Contains(posName, "sales") || strings.Contains(posName, "marketing") ||
		strings.Contains(posName, "motoris") || strings.Contains(posName, "booster") ||
		strings.Contains(posName, "take order") || strings.Contains(posName, "task order") ||
		strings.Contains(posName, "modern trade") || strings.Contains(posName, "kam") ||
		strings.Contains(posName, "key account") || strings.Contains(posName, "kanvas") ||
		strings.Contains(posName, "canvas") || strings.Contains(posName, "spanduk") ||
		strings.Contains(posName, "promosi") || strings.Contains(posName, "branding") ||
		strings.Contains(deptName, "sales") || strings.Contains(deptName, "marketing") ||
		strings.Contains(deptName, "promosi") || strings.Contains(deptName, "branding")

	// Check regex word boundary for short codes "mt", "to", "spv"
	words := strings.FieldsFunc(posName, func(r rune) bool {
		return r == ' ' || r == '/' || r == '-' || r == '_' || r == '.'
	})
	for _, w := range words {
		if w == "to" || w == "mt" || w == "spv" {
			isSalesOrMarketing = true
			break
		}
	}

	return isDriver || isSalesOrMarketing
}

func getStaffCategory(employee *domain.Employee) string {
	if employee == nil {
		return "FIELD_STAFF"
	}
	var deptName, posName string
	if employee.Department != nil {
		deptName = strings.ToLower(strings.TrimSpace(employee.Department.Name))
	}
	if employee.JobPosition != nil {
		posName = strings.ToLower(strings.TrimSpace(employee.JobPosition.Title))
	}

	if strings.Contains(posName, "driver") || strings.Contains(posName, "supir") ||
		strings.Contains(posName, "sopir") || strings.Contains(posName, "pengemudi") ||
		strings.Contains(posName, "kurir") || strings.Contains(posName, "delivery") ||
		strings.Contains(deptName, "driver") || strings.Contains(deptName, "logistik") ||
		strings.Contains(deptName, "delivery") {
		return "DRIVER"
	}

	if strings.Contains(posName, "marketing") || strings.Contains(posName, "promosi") ||
		strings.Contains(posName, "branding") || strings.Contains(posName, "spanduk") ||
		strings.Contains(deptName, "marketing") || strings.Contains(deptName, "promosi") {
		return "MARKETING"
	}

	return "SALES"
}

func (h *SalesHandler) GetNooEcTracking(c *gin.Context) {
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	now := time.Now()
	startDate := now
	endDate := now

	if startDateStr != "" {
		if parsed, err := time.Parse("2006-01-02", startDateStr); err == nil {
			startDate = parsed
		}
	}
	if endDateStr != "" {
		if parsed, err := time.Parse("2006-01-02", endDateStr); err == nil {
			endDate = parsed
		}
	}

	var branchID *uuid.UUID
	if idStr := c.Query("branch_id"); idStr != "" && idStr != "ALL" && idStr != "ALL_BRANCHES" {
		if id, err := uuid.Parse(idStr); err == nil {
			branchID = &id
		}
	}

	var employeeID *uuid.UUID
	if idStr := c.Query("employee_id"); idStr != "" && idStr != "ALL" {
		if id, err := uuid.Parse(idStr); err == nil {
			employeeID = &id
		}
	}

	salesType := c.Query("sales_type")

	res, err := h.salesUsecase.GetNooEcTracking(startDate, endDate, branchID, employeeID, salesType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memuat data tracking NOO & EC: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Success",
		"data":    res,
	})
}

func (h *SalesHandler) ExportNooEcTrackingExcel(c *gin.Context) {
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	now := time.Now()
	startDate := now
	endDate := now

	if startDateStr != "" {
		if parsed, err := time.Parse("2006-01-02", startDateStr); err == nil {
			startDate = parsed
		}
	}
	if endDateStr != "" {
		if parsed, err := time.Parse("2006-01-02", endDateStr); err == nil {
			endDate = parsed
		}
	}

	var branchID *uuid.UUID
	if idStr := c.Query("branch_id"); idStr != "" && idStr != "ALL" && idStr != "ALL_BRANCHES" {
		if id, err := uuid.Parse(idStr); err == nil {
			branchID = &id
		}
	}

	var employeeID *uuid.UUID
	if idStr := c.Query("employee_id"); idStr != "" && idStr != "ALL" {
		if id, err := uuid.Parse(idStr); err == nil {
			employeeID = &id
		}
	}

	salesType := c.Query("sales_type")

	xlsxBytes, err := h.salesUsecase.ExportNooEcTrackingExcel(startDate, endDate, branchID, employeeID, salesType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat file excel tracking NOO & EC: " + err.Error()})
		return
	}

	filename := fmt.Sprintf("tracking_noo_ec_%s_%s.xlsx", startDate.Format("20060102"), endDate.Format("20060102"))
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", xlsxBytes)
}

// TriggerRemoteSync mengirimkan sinyal silent push notification ke smartphone sales agar segera mengunggah transaksi offline.
func (h *SalesHandler) TriggerRemoteSync(c *gin.Context) {
	empIDStr := c.Param("employee_id")
	empID, err := uuid.Parse(empIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID Karyawan/Sales tidak valid"})
		return
	}

	emp, err := h.employeeUseCase.GetEmployeeByID(empID)
	if err != nil || emp == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Data karyawan sales tidak ditemukan"})
		return
	}

	if emp.UserID == uuid.Nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Sales ini belum memiliki akun pengguna/login di aplikasi mobile"})
		return
	}

	if h.fcmUseCase == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Layanan Firebase Cloud Messaging belum terkonfigurasi"})
		return
	}

	if err := h.fcmUseCase.SendRemoteSyncTrigger(emp.UserID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengirimkan sinyal remote sync: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Sinyal Remote Sync berhasil dikirim ke perangkat %s %s. Aplikasi mobile akan otomatis mengunggah data.", emp.FirstName, emp.LastName),
	})
}

// TriggerRemoteSyncAll mengirimkan sinyal silent sync ke seluruh sales aktif.
func (h *SalesHandler) TriggerRemoteSyncAll(c *gin.Context) {
	if h.fcmUseCase == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Layanan Firebase Cloud Messaging belum terkonfigurasi"})
		return
	}

	if err := h.fcmUseCase.BroadcastRemoteSyncTrigger(nil); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengirim broadcast remote sync: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Sinyal Remote Sync berhasil dikirim ke seluruh perangkat sales aktif.",
	})
}

// ImportOfflineBackup menerima file cadangan offline JSON dan memasukkannya ke database dengan proteksi idempotency.
func (h *SalesHandler) ImportOfflineBackup(c *gin.Context) {
	var req struct {
		Transactions []usecase.CreateTransactionRequest `json:"transactions"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format data cadangan JSON tidak valid: " + err.Error()})
		return
	}

	if len(req.Transactions) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tidak ada data transaksi di dalam file cadangan"})
		return
	}

	successCount := 0
	duplicateCount := 0
	failCount := 0
	var errorsList []string

	for idx, tx := range req.Transactions {
		_, err := h.salesUsecase.CreateTransaction(tx)
		if err != nil {
			if strings.Contains(err.Error(), "duplicate") {
				duplicateCount++
			} else {
				failCount++
				errorsList = append(errorsList, fmt.Sprintf("Transaksi #%d (Nota: %s, Toko: %s): %v", idx+1, tx.ReceiptNo, tx.StoreID, err))
			}
		} else {
			successCount++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Impor data cadangan offline selesai: %d berhasil baru, %d sudah ada sebelumnya (duplikat), %d gagal.", successCount, duplicateCount, failCount),
		"data": gin.H{
			"total":       len(req.Transactions),
			"success":     successCount,
			"duplicates":  duplicateCount,
			"failed":      failCount,
			"error_items": errorsList,
		},
	})
}

func parseCityReportPeriod(c *gin.Context) (time.Time, time.Time) {
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		loc = time.Local
	}
	now := time.Now().In(loc)

	periodType := c.DefaultQuery("period_type", "monthly")
	dateStr := c.Query("date")
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")
	monthStr := c.Query("month")
	yearStr := c.Query("year")

	if dateStr == "" && startDateStr != "" {
		dateStr = startDateStr
	}

	var startDate, endDate time.Time

	switch periodType {
	case "daily":
		targetDate := now
		if dateStr != "" {
			if parsed, err := time.ParseInLocation("2006-01-02", dateStr, loc); err == nil {
				targetDate = parsed
			}
		}
		startDate = time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), 0, 0, 0, 0, loc)
		endDate = time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), 23, 59, 59, 999999999, loc)

	case "yearly":
		year := now.Year()
		if yearStr != "" {
			if y, err := strconv.Atoi(yearStr); err == nil {
				year = y
			}
		}
		startDate = time.Date(year, 1, 1, 0, 0, 0, 0, loc)
		endDate = time.Date(year, 12, 31, 23, 59, 59, 999999999, loc)

	case "custom":
		startDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc)
		endDate = time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 999999999, loc)
		if startDateStr != "" {
			if parsed, err := time.ParseInLocation("2006-01-02", startDateStr, loc); err == nil {
				startDate = time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 0, 0, 0, 0, loc)
			}
		}
		if endDateStr != "" {
			if parsed, err := time.ParseInLocation("2006-01-02", endDateStr, loc); err == nil {
				endDate = time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 23, 59, 59, 999999999, loc)
			}
		} else if startDateStr != "" {
			// If end_date not provided, default to end of startDate day
			endDate = time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 23, 59, 59, 999999999, loc)
		}
		if startDate.After(endDate) {
			startDate, endDate = endDate, startDate
		}

	case "monthly":
		fallthrough
	default:
		year := now.Year()
		month := int(now.Month())
		if yearStr != "" {
			if y, err := strconv.Atoi(yearStr); err == nil {
				year = y
			}
		}
		if monthStr != "" {
			if m, err := strconv.Atoi(monthStr); err == nil {
				month = m
			}
		}
		startDate = time.Date(year, time.Month(month), 1, 0, 0, 0, 0, loc)
		endDate = startDate.AddDate(0, 1, 0).Add(-time.Nanosecond)
	}

	return startDate, endDate
}

func (h *SalesHandler) GetCityRevenueReport(c *gin.Context) {
	startDate, endDate := parseCityReportPeriod(c)

	companyIDStr := c.Query("company_id")
	branchIDStr := c.Query("branch_id")
	salesType := c.Query("sales_type")
	productIDStr := c.Query("product_id")
	productIDsStr := c.Query("product_ids")

	var companyID uuid.UUID
	if companyIDStr != "" && companyIDStr != "ALL" {
		if id, err := uuid.Parse(companyIDStr); err == nil {
			companyID = id
		}
	}

	branchIDs := utils.ParseUUIDs(branchIDStr)

	var productIDs []uuid.UUID
	if productIDsStr != "" {
		for _, s := range strings.Split(productIDsStr, ",") {
			s = strings.TrimSpace(s)
			if id, err := uuid.Parse(s); err == nil && id != uuid.Nil {
				productIDs = append(productIDs, id)
			}
		}
	} else if productIDStr != "" && productIDStr != "ALL" {
		for _, s := range strings.Split(productIDStr, ",") {
			s = strings.TrimSpace(s)
			if id, err := uuid.Parse(s); err == nil && id != uuid.Nil {
				productIDs = append(productIDs, id)
			}
		}
	}

	res, err := h.salesUsecase.GetCityRevenueReport(startDate, endDate, companyID, branchIDs, salesType, productIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memuat rekapitulasi omzet per kota: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Success",
		"data":    res,
	})
}

func (h *SalesHandler) ExportCityRevenueExcel(c *gin.Context) {
	startDate, endDate := parseCityReportPeriod(c)

	companyIDStr := c.Query("company_id")
	branchIDStr := c.Query("branch_id")
	salesType := c.Query("sales_type")
	productIDStr := c.Query("product_id")
	productIDsStr := c.Query("product_ids")

	var companyID uuid.UUID
	if companyIDStr != "" && companyIDStr != "ALL" {
		if id, err := uuid.Parse(companyIDStr); err == nil {
			companyID = id
		}
	}

	branchIDs := utils.ParseUUIDs(branchIDStr)

	var productIDs []uuid.UUID
	if productIDsStr != "" {
		for _, s := range strings.Split(productIDsStr, ",") {
			s = strings.TrimSpace(s)
			if id, err := uuid.Parse(s); err == nil && id != uuid.Nil {
				productIDs = append(productIDs, id)
			}
		}
	} else if productIDStr != "" && productIDStr != "ALL" {
		for _, s := range strings.Split(productIDStr, ",") {
			s = strings.TrimSpace(s)
			if id, err := uuid.Parse(s); err == nil && id != uuid.Nil {
				productIDs = append(productIDs, id)
			}
		}
	}

	xlsxBytes, err := h.salesUsecase.ExportCityRevenueExcel(startDate, endDate, companyID, branchIDs, salesType, productIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengekspor laporan omzet per kota: " + err.Error()})
		return
	}

	filename := fmt.Sprintf("rekap_omzet_per_kota_%s_%s.xlsx", startDate.Format("20060102"), endDate.Format("20060102"))
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", xlsxBytes)
}

func (h *SalesHandler) ReconcileReturns(c *gin.Context) {
	count, err := h.salesUsecase.ReconcileAllNotaReturnItems()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Rekonsiliasi retur nota berhasil diproses",
		"count":   count,
	})
}
