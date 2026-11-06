package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"example.com/hris/backend/internal/usecase"
	"example.com/hris/backend/pkg/utils"
)

type SalesOrderHandler struct {
	soUsecase       usecase.SalesOrderUsecase
	employeeUseCase usecase.EmployeeUsecase
	auditLogUseCase usecase.AuditLogUsecase
}

func NewSalesOrderHandler(soUsecase usecase.SalesOrderUsecase, employeeUseCase usecase.EmployeeUsecase, auditLogUseCase usecase.AuditLogUsecase) *SalesOrderHandler {
	return &SalesOrderHandler{soUsecase, employeeUseCase, auditLogUseCase}
}

// RegisterRoutes mendaftarkan semua route SO untuk admin (web dashboard).
func (h *SalesOrderHandler) RegisterRoutes(router *gin.RouterGroup) {
	so := router.Group("/admin/sales-orders")
	so.Use(RoleMiddleware(string(domain.RoleSuperAdmin), string(domain.RoleHRAdmin), string(domain.RoleManager), string(domain.RoleAdminNota), string(domain.RoleSupervisorSales), string(domain.RoleSupervisorSalesMT), string(domain.RoleAdminGudang)))
	{
		so.GET("", h.GetByBranch)
		so.GET("/:id", h.GetByID)
		so.POST("/manual", h.CreateSOManual)

		// *** ALUR BARU: Admin Nota ***
		so.PATCH("/:id/admin-confirm", h.AdminConfirmSO)
		so.PATCH("/:id/admin-reject", h.AdminRejectSO)
		so.PATCH("/:id/tax-approve", h.ApproveTaxDocuments)
		so.PATCH("/mt-approve-tax/:id", h.ApproveMTTaxDocuments)

		// Admin Nota: Adjust pesanan (diskon + qty item) + riwayat pesanan toko
		so.PATCH("/:id/adjust", h.AdjustSO)
		so.GET("/store/:store_id/orders", h.GetByStore)
		so.GET("/store/:store_id/debt", h.GetStoreDebtHistory)

		// Driver routes (dipanggil dari mobile)
		so.PATCH("/:id/confirm-delivery", h.ConfirmDelivery)
		so.PATCH("/:id/collect-payment", h.CollectPayment)

		// Legacy routes
		so.PATCH("/:id/confirm", h.ConfirmSO)
		so.PATCH("/:id/reject", h.RejectSO)
		so.PATCH("/:id/cancel", h.CancelSO)
		so.PATCH("/:id/process-warehouse", h.ProcessByWarehouse)
		so.PATCH("/:id/override-backorder", h.OverrideBackorder)
		so.POST("/:id/recheck-stock", h.RecheckStock)
		so.PATCH("/:id/confirm-pod", h.ConfirmPOD)
		so.POST("/:id/convert", h.ConvertToInvoice)
		so.POST("/:id/convert-direct", h.ConvertDirectToInvoice)
		so.POST("/:id/restore", h.RestoreSO)
		so.POST("/:id/superadmin-complete", h.SuperAdminCompleteSO)
		so.POST("/bulk-superadmin-complete", h.SuperAdminBulkCompleteSO)
		so.GET("/pending-stats", h.GetPendingStats)
		so.DELETE("/:id", h.DeleteSO)
	}
}

// SetupMobileRoutes mendaftarkan route SO untuk salesman (mobile app).
func (h *SalesOrderHandler) SetupMobileRoutes(router *gin.RouterGroup) {
	so := router.Group("/sales-orders") // This matches mobile AppConstants.baseUrl + 'sales-orders/'
	{
		so.POST("/confirm-delivery", h.ConfirmDelivery)
		so.POST("/collect-payment", h.CollectPayment)
	}

	orders := router.Group("/sales/orders")
	{
		orders.POST("", h.CreateSO)
		orders.POST("/upload-nota", h.UploadNotaFisik)
		orders.GET("", h.GetByEmployee)
		orders.DELETE("/:id", h.CancelSO)
		orders.DELETE("/:id/draft", h.DeleteDraftSOBySales)
		// Sales bisa mengedit SO yang masih DRAFT (sebelum diverifikasi Admin Nota)
		orders.PATCH("/:id/edit", h.EditSOBySales)
	}

	stores := router.Group("/sales/stores")
	{
		stores.GET("/:id/orders", h.GetByStore)
	}

	mtOrders := router.Group("/sales/modern-trade")
	{
		mtOrders.GET("/orders", h.GetMTOrdersByEmployee)
		mtOrders.PUT("/orders/:id/status", h.UpdateMTOrderStatus)
		mtOrders.POST("/orders/:id/submit-settlement", h.SubmitMTSettlement)
		mtOrders.GET("/settlement-queue", h.GetMTSettlementQueue)
		mtOrders.POST("/settlements/:id/approve", h.ApproveMTSettlement)
	}
}

// --- Admin & Mobile Handlers ---

func (h *SalesOrderHandler) GetByBranch(c *gin.Context) {
	branchIDStr := c.Query("branch_id")
	status := c.DefaultQuery("status", "ALL")
	salesType := c.Query("sales_type")
	paymentStatus := c.Query("payment_status")

	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	date := c.Query("date")
	month := c.Query("month")

	if date != "" {
		startDate = date
		endDate = date
	} else if month != "" {
		// month format: YYYY-MM
		startDate = month + "-01"
		parsedMonth, _ := time.Parse("2006-01", month)
		lastDay := parsedMonth.AddDate(0, 1, -1).Format("2006-01-02")
		endDate = lastDay
	}

	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "0") // 0 means all

	page, _ := strconv.Atoi(pageStr)
	limit, _ := strconv.Atoi(limitStr)

	branchIDs := utils.ParseUUIDs(branchIDStr)

	employeeIDStr := c.Query("employee_id")
	companyIDStr := c.Query("company_id")
	search := c.Query("search")

	var employeeID uuid.UUID
	if employeeIDStr != "" && employeeIDStr != "ALL" {
		var err error
		employeeID, err = uuid.Parse(employeeIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "employee_id tidak valid"})
			return
		}
	}

	var companyID uuid.UUID
	if companyIDStr != "" && companyIDStr != "ALL" {
		var err error
		companyID, err = uuid.Parse(companyIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "company_id tidak valid"})
			return
		}
	}

	orders, total, summary, err := h.soUsecase.GetSOByBranch(branchIDs, status, salesType, paymentStatus, startDate, endDate, employeeID, companyID, search, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Success",
		"data":    orders,
		"meta": gin.H{
			"total":   total,
			"page":    page,
			"limit":   limit,
			"summary": summary,
		},
	})
}

func (h *SalesOrderHandler) GetByStore(c *gin.Context) {
	storeIDStr := c.Param("store_id")
	if storeIDStr == "" {
		storeIDStr = c.Param("id")
	}
	storeID, err := uuid.Parse(storeIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid store ID"})
		return
	}

	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")

	page, _ := strconv.Atoi(pageStr)
	limit, _ := strconv.Atoi(limitStr)

	orders, total, err := h.soUsecase.GetSOByStore(storeID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Success",
		"data":    orders,
		"meta": gin.H{
			"total": total,
			"page":  page,
			"limit": limit,
		},
	})
}

func (h *SalesOrderHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}
	so, err := h.soUsecase.GetSOByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "SO tidak ditemukan"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Success", "data": so})
}

// CreateSO digunakan oleh salesman (mobile) untuk membuat SO baru.
func (h *SalesOrderHandler) CreateSO(c *gin.Context) {
	var req usecase.CreateSORequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Resolve EmployeeID dari token jika tidak disertakan
	if req.EmployeeID == uuid.Nil {
		if userIDStr, exists := c.Get("userID"); exists {
			userID := userIDStr.(uuid.UUID)
			emp, err := h.employeeUseCase.GetEmployeeByUserID(userID)
			if err == nil && emp != nil {
				req.EmployeeID = emp.ID
				if emp.CompanyID != nil {
					req.CompanyID = *emp.CompanyID
				}
			}
		}
	}

	so, err := h.soUsecase.CreateSO(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Pesanan berhasil dibuat", "data": so})
}

// UploadNotaFisik digunakan oleh salesman (mobile) untuk upload foto nota fisik dan buat draft SO.
func (h *SalesOrderHandler) UploadNotaFisik(c *gin.Context) {
	var req usecase.UploadNotaFisikRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Auto-fill EmployeeID if not provided but exists in token
	if req.EmployeeID.String() == "00000000-0000-0000-0000-000000000000" {
		if userIDStr, exists := c.Get("userID"); exists {
			userID := userIDStr.(uuid.UUID)
			employee, err := h.employeeUseCase.GetEmployeeByUserID(userID)
			if err == nil && employee != nil {
				req.EmployeeID = employee.ID
				if employee.CompanyID != nil {
					req.CompanyID = *employee.CompanyID
				}
			}
		}
	}

	so, err := h.soUsecase.UploadNotaFisik(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Nota fisik berhasil diupload, Pesanan draft dibuat", "data": so})
}

// CreateSOManual digunakan oleh admin (web) untuk membuat SO secara manual.
func (h *SalesOrderHandler) CreateSOManual(c *gin.Context) {
	var req usecase.CreateSORequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	so, err := h.soUsecase.CreateSO(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Pesanan berhasil dibuat", "data": so})
}

func (h *SalesOrderHandler) GetByEmployee(c *gin.Context) {
	var employeeID uuid.UUID
	empIDStr := c.Query("employee_id")

	if empIDStr != "" {
		employeeID, _ = uuid.Parse(empIDStr)
	} else if userIDStr, exists := c.Get("userID"); exists {
		userID := userIDStr.(uuid.UUID)
		emp, err := h.employeeUseCase.GetEmployeeByUserID(userID)
		if err == nil && emp != nil {
			employeeID = emp.ID
		}
	}

	if employeeID == uuid.Nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "employee_id diperlukan"})
		return
	}

	orders, err := h.soUsecase.GetSOByEmployee(employeeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Success", "data": orders})
}

func (h *SalesOrderHandler) GetMTOrdersByEmployee(c *gin.Context) {
	var employeeID uuid.UUID
	empIDStr := c.Query("employee_id")

	if empIDStr != "" {
		employeeID, _ = uuid.Parse(empIDStr)
	} else if userIDStr, exists := c.Get("userID"); exists {
		userID := userIDStr.(uuid.UUID)
		emp, err := h.employeeUseCase.GetEmployeeByUserID(userID)
		if err == nil && emp != nil {
			employeeID = emp.ID
		}
	}

	if employeeID == uuid.Nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "employee_id diperlukan"})
		return
	}

	orders, err := h.soUsecase.GetSOByEmployee(employeeID) // For now, we reuse GetSOByEmployee and filter or we should create GetMTOrdersByEmployee in usecase.
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Filter for MT orders only
	var mtOrders []domain.SalesOrder
	for _, order := range orders {
		if order.IsMT {
			mtOrders = append(mtOrders, order)
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Success", "data": mtOrders})
}

// UpdateMTOrderStatus mengubah status SO dan PO
func (h *SalesOrderHandler) UpdateMTOrderStatus(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID SO tidak valid"})
		return
	}
	var req struct {
		Status string `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.soUsecase.UpdateMTOrderStatus(id, req.Status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Status berhasil diperbarui"})
}

// SubmitMTSettlement endpoint untuk submit bukti tagihan MT
func (h *SalesOrderHandler) SubmitMTSettlement(c *gin.Context) {
	soIDStr := c.Param("id")
	soID, err := uuid.Parse(soIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID SO tidak valid"})
		return
	}

	var req usecase.CollectPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.SOID = soID

	// Resolve EmployeeID from token
	if userIDStr, exists := c.Get("userID"); exists {
		userID := userIDStr.(uuid.UUID)
		emp, err := h.employeeUseCase.GetEmployeeByUserID(userID)
		if err == nil && emp != nil {
			req.EmployeeID = emp.ID
		}
	}

	so, err := h.soUsecase.CollectPayment(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal submit settlement: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Bukti transfer berhasil dikirim", "data": so})
}

// GetMTSettlementQueue digunakan oleh admin nota untuk mengambil daftar settlement MT yang perlu diverifikasi
func (h *SalesOrderHandler) GetMTSettlementQueue(c *gin.Context) {
	queue, err := h.soUsecase.GetMTSettlementQueue()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Success",
		"data":    queue,
	})
}

// ApproveMTSettlement memverifikasi pembayaran MT dan mengubah omzet
func (h *SalesOrderHandler) ApproveMTSettlement(c *gin.Context) {
	paymentIDStr := c.Param("id")
	paymentID, err := uuid.Parse(paymentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	var adminID uuid.UUID
	if userIDStr, exists := c.Get("userID"); exists {
		userID := userIDStr.(uuid.UUID)
		emp, err := h.employeeUseCase.GetEmployeeByUserID(userID)
		if err == nil && emp != nil {
			adminID = emp.ID
		}
	}

	if err := h.soUsecase.ApproveMTSettlement(paymentID, adminID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Omzet berhasil di-settle"})
}

// ConfirmSO mengkonfirmasi SO dan me-reserve stok di gudang.
func (h *SalesOrderHandler) ConfirmSO(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	var confirmedByID uuid.UUID
	if userIDStr, exists := c.Get("userID"); exists {
		userID := userIDStr.(uuid.UUID)
		emp, err := h.employeeUseCase.GetEmployeeByUserID(userID)
		if err == nil && emp != nil {
			confirmedByID = emp.ID
		}
	}

	if err := h.soUsecase.ConfirmSO(id, confirmedByID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Pesanan dikonfirmasi dan stok berhasil di-reserve"})
}

// RejectSO menolak SO.
func (h *SalesOrderHandler) RejectSO(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	var req struct {
		Notes string `json:"notes"`
	}
	c.ShouldBindJSON(&req)

	var rejectedByID uuid.UUID
	if userIDStr, exists := c.Get("userID"); exists {
		userID := userIDStr.(uuid.UUID)
		emp, err := h.employeeUseCase.GetEmployeeByUserID(userID)
		if err == nil && emp != nil {
			rejectedByID = emp.ID
		}
	}

	if err := h.soUsecase.RejectSO(id, rejectedByID, req.Notes); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Pesanan ditolak"})
}

// CancelSO membatalkan SO dan melepas reserve stok jika sudah dikonfirmasi.
func (h *SalesOrderHandler) CancelSO(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	if err := h.soUsecase.CancelSO(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Pesanan dibatalkan dan reserve stok dilepas"})
}

// ConvertToInvoice mengkonversi SO ke Faktur Penjualan (SalesTransaction).
// Ini memotong stok secara permanen.
func (h *SalesOrderHandler) ConvertToInvoice(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	var req struct {
		CompanyID uuid.UUID `json:"company_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	invoice, err := h.soUsecase.ConvertToInvoice(id, req.CompanyID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"message": "Faktur Penjualan berhasil diterbitkan dan stok sudah dipotong",
		"data":    invoice,
	})
}

// ConvertDirectToInvoice menerbitkan faktur motoris tanpa surat jalan.
func (h *SalesOrderHandler) ConvertDirectToInvoice(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}
	adminID, err := h.resolveEmployeeID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Identitas admin tidak ditemukan"})
		return
	}
	oldSO, _ := h.soUsecase.GetSOByID(id)
	invoice, err := h.soUsecase.ConvertDirectToInvoice(id, adminID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if oldSO != nil {
		userID, _ := c.Get("userID")
		if actorID, ok := userID.(uuid.UUID); ok {
			newSO, _ := h.soUsecase.GetSOByID(id)
			name := ""
			if employee, err := h.employeeUseCase.GetEmployeeByUserID(actorID); err == nil && employee != nil {
				name = employee.FirstName + " " + employee.LastName
			}
			h.auditLogUseCase.Log(oldSO.CompanyID, actorID, name, "CONVERT_DIRECT", "SalesOrder", id.String(),
				"SO langsung menjadi faktur tanpa pengiriman: "+invoice.ReceiptNo, oldSO, newSO, c.ClientIP())
		}
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Faktur Penjualan langsung berhasil diterbitkan", "data": invoice})
}

// ProcessByWarehouse digunakan oleh tim gudang untuk validasi fisik barang.
func (h *SalesOrderHandler) ProcessByWarehouse(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	var req usecase.ProcessByWarehouseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.SOID = id

	// Resolve ProcessedByID dari token
	if userIDStr, exists := c.Get("userID"); exists {
		userID := userIDStr.(uuid.UUID)
		emp, err := h.employeeUseCase.GetEmployeeByUserID(userID)
		if err == nil && emp != nil {
			req.ProcessedByID = emp.ID
		}
	}

	if err := h.soUsecase.ProcessByWarehouse(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Barang berhasil diproses gudang dan Surat Jalan telah diterbitkan"})
}

// ConfirmPOD konfirmasi penerimaan barang (Proof of Delivery).
func (h *SalesOrderHandler) ConfirmPOD(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	var req usecase.ConfirmPODRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.SOID = id

	if err := h.soUsecase.ConfirmPOD(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Penerimaan barang berhasil dikonfirmasi (POD)"})
}

// DeleteSO menghapus SO dengan mekanisme soft delete.
func (h *SalesOrderHandler) DeleteSO(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
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
	existingSO, _ := h.soUsecase.GetSOByID(id)

	if err := h.soUsecase.SoftDeleteSO(id, empID, reason); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if existingSO != nil {
		h.auditLogUseCase.Log(companyID, userID, userName, "DELETE", "SalesOrder", id.String(), "Deleted Sales Order: "+existingSO.SONumber+" (Alasan: "+reason+")", existingSO, nil, c.ClientIP())
	}

	c.JSON(http.StatusOK, gin.H{"message": "Pesanan berhasil dihapus"})
}

func (h *SalesOrderHandler) DeleteDraftSOBySales(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}
	employeeID, err := h.resolveEmployeeID(c)
	if err != nil || employeeID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Identitas sales tidak ditemukan"})
		return
	}
	existing, _ := h.soUsecase.GetSOByID(id)
	if err := h.soUsecase.DeleteDraftSOBySales(id, employeeID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if existing != nil {
		if userValue, ok := c.Get("userID"); ok {
			userID := userValue.(uuid.UUID)
			h.auditLogUseCase.Log(existing.CompanyID, userID, "Sales", "DELETE", "SalesOrder", id.String(),
				"Sales deleted draft SO: "+existing.SONumber, existing, nil, c.ClientIP())
		}
	}
	c.JSON(http.StatusOK, gin.H{"message": "Pesanan DRAFT berhasil dihapus"})
}

// RestoreSO memulihkan SO yang terhapus
func (h *SalesOrderHandler) RestoreSO(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
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

	if err := h.soUsecase.RestoreSO(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.auditLogUseCase.Log(companyID, userID, userName, "RESTORE", "SalesOrder", id.String(), "Restored Sales Order: "+id.String(), nil, nil, c.ClientIP())

	c.JSON(http.StatusOK, gin.H{"message": "Pesanan berhasil dipulihkan"})
}
func (h *SalesOrderHandler) OverrideBackorder(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	if err := h.soUsecase.OverrideBackorder(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Pesanan berhasil di-override ke antrean kirim"})
}

func (h *SalesOrderHandler) RecheckStock(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	so, err := h.soUsecase.RecheckStock(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	msg := "Pengecekan stok berhasil."
	if so.Status == domain.SOStatusConfirmed {
		msg = "Semua stok telah terpenuhi! Status pesanan berubah menjadi CONFIRMED (Siap Kirim)."
	} else {
		msg = "Pengecekan selesai. Sebagian item masih menunggu pasokan stok di gudang."
	}

	c.JSON(http.StatusOK, gin.H{
		"message": msg,
		"data":    so,
	})
}

// =============================================================================
// ALUR BARU: Admin Nota
// =============================================================================

func (h *SalesOrderHandler) AdminConfirmSO(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	// Resolve admin employee ID dari token
	adminID, err := h.resolveEmployeeID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Identitas admin tidak ditemukan"})
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
	existingSO, _ := h.soUsecase.GetSOByID(id)

	// Keep old state
	var oldSO domain.SalesOrder
	if existingSO != nil {
		oldSOBytes, _ := json.Marshal(existingSO)
		_ = json.Unmarshal(oldSOBytes, &oldSO)
	}

	if err := h.soUsecase.AdminConfirmSO(id, adminID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	newSO, _ := h.soUsecase.GetSOByID(id)
	h.auditLogUseCase.Log(companyID, userID, userName, "UPDATE", "SalesOrder", id.String(), "Admin Nota Confirmed Sales Order: "+newSO.SONumber, oldSO, newSO, c.ClientIP())

	c.JSON(http.StatusOK, gin.H{"message": "Nota berhasil diverifikasi oleh Admin Nota"})
}

func (h *SalesOrderHandler) ApproveTaxDocuments(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	adminID, err := h.resolveEmployeeID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Identitas admin tidak ditemukan"})
		return
	}

	var req struct {
		BillingNoteURL      string    `json:"billing_note_url" binding:"required"`
		TaxInvoiceURL       string    `json:"tax_invoice_url" binding:"required"`
		DeliveryNoteURL     string    `json:"delivery_note_url" binding:"required"`
		DestinationBranchID uuid.UUID `json:"destination_branch_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Semua berkas pajak & cabang pengirim wajib diisi"})
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
	existingSO, _ := h.soUsecase.GetSOByID(id)

	var oldSO domain.SalesOrder
	if existingSO != nil {
		oldSOBytes, _ := json.Marshal(existingSO)
		_ = json.Unmarshal(oldSOBytes, &oldSO)
	}

	if err := h.soUsecase.ApproveTaxDocuments(id, adminID, req.BillingNoteURL, req.TaxInvoiceURL, req.DeliveryNoteURL, req.DestinationBranchID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	newSO, _ := h.soUsecase.GetSOByID(id)
	h.auditLogUseCase.Log(companyID, userID, userName, "UPDATE", "SalesOrder", id.String(), "Tax Approved & Assigned Branch for PO MT: "+newSO.SONumber, oldSO, newSO, c.ClientIP())

	c.JSON(http.StatusOK, gin.H{"message": "Dokumen pajak PO MT berhasil disetujui"})
}

func (h *SalesOrderHandler) ApproveMTTaxDocuments(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID MT PO tidak valid"})
		return
	}

	adminID, err := h.resolveEmployeeID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Identitas admin tidak ditemukan"})
		return
	}

	// Parse multipart form
	if err := c.Request.ParseMultipartForm(20 << 20); err != nil { // 20 MB max
		c.JSON(http.StatusBadRequest, gin.H{"error": "Gagal memproses form data"})
		return
	}

	destinationBranchIDStr := c.PostForm("destination_branch_id")
	destinationBranchID, err := uuid.Parse(destinationBranchIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID cabang pengirim tidak valid"})
		return
	}

	uploadFile := func(field string, destDir string) (string, error) {
		file, err := c.FormFile(field)
		if err != nil {
			if err == http.ErrMissingFile {
				return "", fmt.Errorf("berkas %s wajib diunggah", field)
			}
			return "", err
		}

		if err := os.MkdirAll(destDir, os.ModePerm); err != nil {
			return "", err
		}

		filename := uuid.New().String() + filepath.Ext(file.Filename)
		path := filepath.Join(destDir, filename)
		if err := c.SaveUploadedFile(file, path); err != nil {
			return "", err
		}

		return "/" + path, nil
	}

	uploadDir := "uploads/mt_tax"
	billingNoteURL, err := uploadFile("billing_note_file", uploadDir)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	taxInvoiceURL, err := uploadFile("tax_invoice_file", uploadDir)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	deliveryNoteURL, err := uploadFile("delivery_note_file", uploadDir)
	if err != nil {
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

	if err := h.soUsecase.ApproveMTTaxDocuments(id, adminID, billingNoteURL, taxInvoiceURL, deliveryNoteURL, destinationBranchID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.auditLogUseCase.Log(companyID, userID, userName, "UPDATE", "MTPurchaseOrder", id.String(), "Tax Approved & Assigned Branch for PO MT via MT API", nil, nil, c.ClientIP())

	c.JSON(http.StatusOK, gin.H{"message": "Dokumen pajak PO MT berhasil disetujui"})
}

func (h *SalesOrderHandler) AdminRejectSO(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	adminID, err := h.resolveEmployeeID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Identitas admin tidak ditemukan"})
		return
	}

	var req struct {
		Notes string `json:"notes"`
	}
	_ = c.ShouldBindJSON(&req)

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
	existingSO, _ := h.soUsecase.GetSOByID(id)

	// Keep old state
	var oldSO domain.SalesOrder
	if existingSO != nil {
		oldSOBytes, _ := json.Marshal(existingSO)
		_ = json.Unmarshal(oldSOBytes, &oldSO)
	}

	if err := h.soUsecase.AdminRejectSO(id, adminID, req.Notes); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	newSO, _ := h.soUsecase.GetSOByID(id)
	h.auditLogUseCase.Log(companyID, userID, userName, "UPDATE", "SalesOrder", id.String(), "Admin Nota Rejected Sales Order: "+newSO.SONumber+" with notes: "+req.Notes, oldSO, newSO, c.ClientIP())

	c.JSON(http.StatusOK, gin.H{"message": "Nota berhasil ditolak"})
}

// =============================================================================
// ALUR BARU: Driver — Konfirmasi Pengiriman & Tagih Pembayaran
// =============================================================================

func (h *SalesOrderHandler) ConfirmDelivery(c *gin.Context) {
	var req usecase.ConfirmDeliveryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// If ID is in param (PATCH /admin/sales-orders/:id/confirm-delivery), override body so_id
	idStr := c.Param("id")
	if idStr != "" {
		id, err := uuid.Parse(idStr)
		if err == nil {
			req.SOID = id
		}
	}

	// Resolve employee ID from token for SalesStock deduction
	employeeID, _ := h.resolveEmployeeID(c)
	req.EmployeeID = employeeID

	invoice, err := h.soUsecase.ConfirmDelivery(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Pengiriman berhasil dikonfirmasi",
		"invoice": invoice,
	})
}

func (h *SalesOrderHandler) CollectPayment(c *gin.Context) {
	var req usecase.CollectPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// If ID is in param (PATCH /admin/sales-orders/:id/collect-payment), override body so_id
	idStr := c.Param("id")
	if idStr != "" {
		id, err := uuid.Parse(idStr)
		if err == nil {
			req.SOID = id
		}
	}

	// Resolve employee ID from token
	employeeID, _ := h.resolveEmployeeID(c)
	req.EmployeeID = employeeID

	so, err := h.soUsecase.CollectPayment(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Pembayaran berhasil dicatat",
		"data":    so,
	})
}

// =============================================================================
// Mobile: Sales Edit SO (hanya status DRAFT)
// =============================================================================

// EditSOBySales memungkinkan sales/salesman mengedit SO yang masih berstatus DRAFT.
// Ini adalah mekanisme koreksi sebelum nota diverifikasi oleh Admin Nota.
// Hanya bisa dilakukan pada SO milik sales tersebut dengan status DRAFT.
func (h *SalesOrderHandler) EditSOBySales(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	// Resolve employee ID dari token
	employeeID, err := h.resolveEmployeeID(c)
	if err != nil || employeeID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Identitas sales tidak ditemukan"})
		return
	}

	var req usecase.AdjustSORequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Pastikan SO milik employee ini dan berstatus DRAFT
	existingSO, err := h.soUsecase.GetSOByID(id)
	if err != nil || existingSO == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "SO tidak ditemukan"})
		return
	}
	if existingSO.EmployeeID != employeeID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Anda tidak memiliki akses untuk mengedit SO ini"})
		return
	}
	if existingSO.Status != "DRAFT" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Pesanan hanya dapat diedit saat berstatus DRAFT (sebelum diverifikasi Admin Nota)"})
		return
	}

	so, err := h.soUsecase.AdjustSO(id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var userID uuid.UUID
	var userName string
	var companyID uuid.UUID
	if userIDStr, exists := c.Get("userID"); exists {
		userID = userIDStr.(uuid.UUID)
		if employee, empErr := h.employeeUseCase.GetEmployeeByUserID(userID); empErr == nil && employee != nil {
			userName = employee.FirstName + " " + employee.LastName
			if employee.CompanyID != nil {
				companyID = *employee.CompanyID
			}
		}
	}
	h.auditLogUseCase.Log(companyID, userID, userName, "UPDATE", "SalesOrder", id.String(),
		"Sales Edit Draft SO: "+so.SONumber, existingSO, so, c.ClientIP())

	c.JSON(http.StatusOK, gin.H{"message": "Pesanan berhasil diperbarui", "data": so})
}

// resolveEmployeeID adalah helper untuk mendapatkan employee ID dari JWT token.
func (h *SalesOrderHandler) resolveEmployeeID(c *gin.Context) (uuid.UUID, error) {
	userIDRaw, exists := c.Get("userID")
	if !exists {
		return uuid.Nil, gin.Error{}
	}
	userID := userIDRaw.(uuid.UUID)
	employee, err := h.employeeUseCase.GetEmployeeByUserID(userID)
	if err != nil || employee == nil {
		return uuid.Nil, err
	}
	return employee.ID, nil
}

// =============================================================================
// Admin Nota: Adjust Pesanan
// =============================================================================

// AdjustSO memungkinkan admin mengubah diskon dan jumlah item SO sebelum masuk faktur.
func (h *SalesOrderHandler) AdjustSO(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	var req usecase.AdjustSORequest
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
	existingSO, _ := h.soUsecase.GetSOByID(id)

	// Keep old state
	var oldSO domain.SalesOrder
	if existingSO != nil {
		oldSOBytes, _ := json.Marshal(existingSO)
		_ = json.Unmarshal(oldSOBytes, &oldSO)
	}

	so, err := h.soUsecase.AdjustSO(id, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.auditLogUseCase.Log(companyID, userID, userName, "UPDATE", "SalesOrder", id.String(), "Adjusted Sales Order: "+so.SONumber, oldSO, so, c.ClientIP())

	c.JSON(http.StatusOK, gin.H{"message": "Pesanan berhasil disesuaikan", "data": so})
}

// GetStoreDebtHistory mengambil riwayat tagihan belum lunas untuk sebuah toko.
func (h *SalesOrderHandler) GetStoreDebtHistory(c *gin.Context) {
	storeID, err := uuid.Parse(c.Param("store_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "store_id tidak valid"})
		return
	}

	history, err := h.soUsecase.GetStoreDebtHistory(storeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Success", "data": history})
}

// SuperAdminCompleteSO memproses penyelesaian satu-klik SO oleh Super Admin
func (h *SalesOrderHandler) SuperAdminCompleteSO(c *gin.Context) {
	roleRaw, _ := c.Get("role")
	roleStr, _ := roleRaw.(string)
	if roleStr != string(domain.RoleSuperAdmin) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Fitur ini hanya dapat diakses oleh Super Admin"})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID pesanan tidak valid"})
		return
	}

	adminID, err := h.resolveEmployeeID(c)
	if err != nil || adminID == uuid.Nil {
		userIDRaw, exists := c.Get("userID")
		if exists {
			adminID = userIDRaw.(uuid.UUID)
		}
	}

	existingSO, _ := h.soUsecase.GetSOByID(id)
	var oldSO map[string]interface{}
	if existingSO != nil {
		oldSOBytes, _ := json.Marshal(existingSO)
		_ = json.Unmarshal(oldSOBytes, &oldSO)
	}

	updatedSO, err := h.soUsecase.SuperAdminFastTrackComplete(id, adminID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Audit Log
	var userID uuid.UUID
	var userName string = "Super Admin"
	var companyID uuid.UUID
	if existingSO != nil {
		companyID = existingSO.CompanyID
	}
	if userIDStr, exists := c.Get("userID"); exists {
		userID = userIDStr.(uuid.UUID)
		if employee, empErr := h.employeeUseCase.GetEmployeeByUserID(userID); empErr == nil && employee != nil {
			userName = employee.FirstName + " " + employee.LastName
			if employee.CompanyID != nil {
				companyID = *employee.CompanyID
			}
		}
	}

	h.auditLogUseCase.Log(companyID, userID, userName, "COMPLETE", "SalesOrder", id.String(), "Super Admin Fast-Track Complete Sales Order: "+updatedSO.SONumber, oldSO, updatedSO, c.ClientIP())

	c.JSON(http.StatusOK, gin.H{
		"message": "Pesanan SO berhasil diselesaikan secara otomatis (Pengiriman, Faktur, Kasir, dan Omzet Terverifikasi)",
		"data":    updatedSO,
	})
}

// SuperAdminBulkCompleteSO memproses penyelesaian massal SO oleh Super Admin
func (h *SalesOrderHandler) SuperAdminBulkCompleteSO(c *gin.Context) {
	roleRaw, _ := c.Get("role")
	roleStr, _ := roleRaw.(string)
	if roleStr != string(domain.RoleSuperAdmin) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Fitur ini hanya dapat diakses oleh Super Admin"})
		return
	}

	var req usecase.BulkSuperAdminCompleteRequest
	if err := c.ShouldBindJSON(&req); err != nil && err.Error() != "EOF" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format request tidak valid: " + err.Error()})
		return
	}

	adminID, err := h.resolveEmployeeID(c)
	if err != nil || adminID == uuid.Nil {
		userIDRaw, exists := c.Get("userID")
		if exists {
			adminID = userIDRaw.(uuid.UUID)
		}
	}

	result, err := h.soUsecase.SuperAdminBulkComplete(req, adminID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Audit Log
	var userID uuid.UUID
	var userName string = "Super Admin"
	var companyID uuid.UUID
	if req.CompanyID != nil {
		companyID = *req.CompanyID
	}
	if userIDStr, exists := c.Get("userID"); exists {
		userID = userIDStr.(uuid.UUID)
		if employee, empErr := h.employeeUseCase.GetEmployeeByUserID(userID); empErr == nil && employee != nil {
			userName = employee.FirstName + " " + employee.LastName
			if companyID == uuid.Nil && employee.CompanyID != nil {
				companyID = *employee.CompanyID
			}
		}
	}

	h.auditLogUseCase.Log(companyID, userID, userName, "BULK_COMPLETE", "SalesOrder", "BULK", fmt.Sprintf("Super Admin Bulk Complete SO: %d sukses, %d gagal dari total %d pesanan", result.SuccessCount, result.FailedCount, result.TotalProcessed), nil, result, c.ClientIP())

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Penyelesaian massal berhasil diproses: %d selesai, %d gagal dari %d pesanan", result.SuccessCount, result.FailedCount, result.TotalProcessed),
		"data":    result,
	})
}

// GetPendingStats mengambil statistik jumlah dan nominal pesanan SO tertahan di DB
func (h *SalesOrderHandler) GetPendingStats(c *gin.Context) {
	branchIDStr := c.Query("branch_id")
	companyIDStr := c.Query("company_id")

	var branchID *uuid.UUID
	var branchIDs []uuid.UUID
	if branchIDStr != "" && branchIDStr != "ALL_BRANCHES" {
		parsedIDs := utils.ParseUUIDs(branchIDStr)
		if len(parsedIDs) == 1 {
			branchID = &parsedIDs[0]
		} else if len(parsedIDs) > 1 {
			branchIDs = parsedIDs
		}
	}

	var companyID *uuid.UUID
	if companyIDStr != "" && companyIDStr != "ALL" {
		if cid, err := uuid.Parse(companyIDStr); err == nil {
			companyID = &cid
		}
	}

	stats, err := h.soUsecase.GetPendingStats(branchID, branchIDs, companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Success",
		"data":    stats,
	})
}
