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

type DeliveryHandler struct {
	usecase         usecase.DeliveryUsecase
	employeeUseCase usecase.EmployeeUsecase
}

func NewDeliveryHandler(u usecase.DeliveryUsecase, e usecase.EmployeeUsecase) *DeliveryHandler {
	return &DeliveryHandler{u, e}
}

func (h *DeliveryHandler) RegisterRoutes(r *gin.RouterGroup) {
	delivery := r.Group("/delivery")
	{
		delivery.POST("/batch", h.CreateBatch)
		delivery.POST("/batch/:id/supervisor-approve", h.SupervisorApprove)
		delivery.POST("/batch/:id/approve", h.ApproveBatch)
		delivery.POST("/batch/:id/assign", h.AssignArmada)
		delivery.PUT("/batch/:id/revise-qty", h.ReviseOrderQuantity)
		delivery.GET("/batches", h.ListBatches)
		delivery.PUT("/batches/:id", h.UpdateBatch)
		delivery.DELETE("/batches/:id", h.DeleteBatch)
		delivery.GET("/driver-load", h.GetDriverLoad)
		delivery.GET("/tasks", h.GetTasks)
		delivery.GET("/history", h.GetHistory)
		delivery.GET("/batch/:id", h.GetDetail)
		delivery.POST("/batch/:id/start", h.StartDelivery)
		delivery.POST("/batch/:id/admin-report", h.AdminReportBatch)
		// Membuat faktur dari SJ COMPLETED milik daily driver
		delivery.POST("/batch/:id/force-invoice", h.ForceInvoiceBatch)
		delivery.POST("/batch/:id/cash", h.UpdateCash)
		delivery.POST("/items/:id/confirm", h.ConfirmItem)
		delivery.POST("/items/confirm-by-receipt", h.ConfirmByReceipt)

		// Gudang: melihat SJ yang perlu disiapkan, scan barcode, konfirmasi ready
		delivery.GET("/batches/warehouse", h.GetWarehouseBatches)
		delivery.GET("/batch/scan", h.ScanBatch) // ?code=SJ-xxx
		delivery.POST("/batch/:id/warehouse-ready", h.WarehouseConfirmReady)
		delivery.POST("/batch/:id/spares", h.UpdateSpares)
		delivery.GET("/batch/:id/eligible-orders", h.GetEligibleOrdersForBatch)
		delivery.POST("/batch/:id/append-orders", h.AppendOrdersToBatch)
		delivery.GET("/subsequent/pending", h.GetPendingSubsequentItems)
		delivery.POST("/subsequent", h.CreateManualSubsequentItem)
		delivery.POST("/batch/:id/subsequent/:sub_id/slip", h.SlipSubsequentItem)
	}
}

func (h *DeliveryHandler) CreateBatch(c *gin.Context) {
	var req usecase.CreateBatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	batch, err := h.usecase.CreateBatch(req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Batch created successfully", batch)
}

func (h *DeliveryHandler) ApproveBatch(c *gin.Context) {
	batchID := uuid.MustParse(c.Param("id"))

	// Resolve employee_id from userID
	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user identity not found"})
		return
	}
	userID := userIDStr.(uuid.UUID)
	employee, err := h.employeeUseCase.GetEmployeeByUserID(userID)
	if err != nil || employee == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "employee identity not found"})
		return
	}
	adminID := employee.ID

	if err := h.usecase.ApproveBatch(batchID, adminID); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Surat jalan berhasil disetujui", nil)
}

func (h *DeliveryHandler) SupervisorApprove(c *gin.Context) {
	batchID := uuid.MustParse(c.Param("id"))

	// Resolve employee_id from userID
	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user identity not found"})
		return
	}
	userID := userIDStr.(uuid.UUID)
	employee, err := h.employeeUseCase.GetEmployeeByUserID(userID)
	if err != nil || employee == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "supervisor identity not found"})
		return
	}
	supervisorID := employee.ID

	if err := h.usecase.SupervisorApproveBatch(batchID, supervisorID); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Batch berhasil disetujui Supervisor", nil)
}

func (h *DeliveryHandler) AssignArmada(c *gin.Context) {
	batchID := uuid.MustParse(c.Param("id"))

	var req struct {
		DriverID  uuid.UUID  `json:"driver_id" binding:"required"`
		HelperID  *uuid.UUID `json:"helper_id"`
		VehicleID uuid.UUID  `json:"vehicle_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Resolve employee_id from userID
	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user identity not found"})
		return
	}
	userID := userIDStr.(uuid.UUID)
	employee, err := h.employeeUseCase.GetEmployeeByUserID(userID)
	if err != nil || employee == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "supervisor identity not found"})
		return
	}
	supervisorID := employee.ID

	if err := h.usecase.AssignArmada(batchID, req.DriverID, req.HelperID, req.VehicleID, supervisorID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Armada berhasil ditugaskan"})
}

func (h *DeliveryHandler) ListBatches(c *gin.Context) {
	branchIDs := utils.ParseUUIDs(c.Query("branch_id"))
	salesType := c.Query("sales_type")
	batches, err := h.usecase.ListBatches(branchIDs, salesType)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Batches fetched successfully", batches)
}

func (h *DeliveryHandler) UpdateBatch(c *gin.Context) {
	id := uuid.MustParse(c.Param("id"))
	var req usecase.CreateBatchRequest // Reuse request struct
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.usecase.UpdateBatchItems(id, req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Batch updated successfully"})
}

func (h *DeliveryHandler) DeleteBatch(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid batch id"})
		return
	}

	err = h.usecase.DeleteBatch(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Batch deleted successfully"})
}

func (h *DeliveryHandler) GetTasks(c *gin.Context) {
	// Resolve employee_id from userID
	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "driver identity not found"})
		return
	}
	userID := userIDStr.(uuid.UUID)
	employee, err := h.employeeUseCase.GetEmployeeByUserID(userID)
	if err != nil || employee == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "driver identity not found"})
		return
	}
	driverID := employee.ID

	tasks, err := h.usecase.GetDriverTasks(driverID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tasks)
}

func (h *DeliveryHandler) GetHistory(c *gin.Context) {
	// Resolve employee_id from userID
	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "driver identity not found"})
		return
	}
	userID := userIDStr.(uuid.UUID)
	employee, err := h.employeeUseCase.GetEmployeeByUserID(userID)
	if err != nil || employee == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "driver identity not found"})
		return
	}
	driverID := employee.ID

	history, err := h.usecase.GetDriverHistory(driverID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, history)
}

func (h *DeliveryHandler) GetDetail(c *gin.Context) {
	idOrNo := c.Param("id")
	fmt.Printf("DEBUG: GetDetail request for [%s]\n", idOrNo)

	var batch *domain.DeliveryBatch
	var err error

	// Try parsing as UUID first
	if id, parseErr := uuid.Parse(idOrNo); parseErr == nil {
		batch, err = h.usecase.GetBatchByID(id)
	} else {
		batch, err = h.usecase.GetBatchDetail(idOrNo)
	}

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Batch atau Surat jalan tidak ditemukan"})
		return
	}

	c.JSON(http.StatusOK, batch)
}

func (h *DeliveryHandler) StartDelivery(c *gin.Context) {
	doNo := c.Param("id")
	if err := h.usecase.StartDelivery(doNo); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Delivery started successfully"})
}

func (h *DeliveryHandler) AdminReportBatch(c *gin.Context) {
	batchID := uuid.MustParse(c.Param("id"))
	var req struct {
		Items            []usecase.AdminReportItemRequest `json:"items" binding:"required"`
		DepartureMileage *int                             `json:"departure_mileage"`
		ReturnMileage    *int                             `json:"return_mileage"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.usecase.AdminReportBatch(batchID, req.Items, req.DepartureMileage, req.ReturnMileage); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Batch reported and completed successfully"})
}

func (h *DeliveryHandler) ConfirmItem(c *gin.Context) {
	itemID := uuid.MustParse(c.Param("id"))
	var req struct {
		Status domain.DeliveryItemStatus `json:"status" binding:"required"`
		Notes  string                    `json:"notes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.usecase.ConfirmItemDelivery(itemID, req.Status, req.Notes); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Item status updated"})
}
func (h *DeliveryHandler) ConfirmByReceipt(c *gin.Context) {
	var req struct {
		ReceiptNo string `json:"receipt_no" binding:"required"`
		Notes     string `json:"notes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.usecase.ConfirmItemByReceipt(req.ReceiptNo, req.Notes); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Item marked as DELIVERED"})
}

func (h *DeliveryHandler) UpdateCash(c *gin.Context) {
	batchID := uuid.MustParse(c.Param("id"))
	var req struct {
		Amount float64 `json:"amount" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.usecase.UpdateBatchCash(batchID, req.Amount); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Cash amount updated"})
}

// =============================================================================
// Gudang: Surat Jalan & Barcode Scan
// =============================================================================

// GetWarehouseBatches mengembalikan SJ yang relevan untuk menu gudang
// (status: SUPERVISOR_APPROVED, PICKING, ON_DELIVERY, COMPLETED).
func (h *DeliveryHandler) GetWarehouseBatches(c *gin.Context) {
	branchIDs := utils.ParseUUIDs(c.Query("branch_id"))
	batches, err := h.usecase.GetWarehouseBatches(branchIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Success", "data": batches})
}

// ScanBatch mencari SJ berdasarkan kode barcode atau nomor SJ.
// GET /delivery/batch/scan?code=SJ-20250519-abc123
func (h *DeliveryHandler) ScanBatch(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "parameter 'code' diperlukan"})
		return
	}

	batch, err := h.usecase.ScanBatch(code)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Surat Jalan tidak ditemukan: " + code})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Surat Jalan ditemukan", "data": batch})
}

// WarehouseConfirmReady: Gudang menandai barang sudah disiapkan -> status PICKING.
func (h *DeliveryHandler) WarehouseConfirmReady(c *gin.Context) {
	batchID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	if err := h.usecase.WarehouseConfirmReady(batchID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Surat Jalan dikonfirmasi siap. Driver bisa mengambil barang."})
}

func (h *DeliveryHandler) SetupPublicRoutes(r *gin.RouterGroup) {
	publicDelivery := r.Group("/public/delivery")
	{
		publicDelivery.GET("/stores", h.PublicGetStores)
		publicDelivery.GET("/track", h.PublicTrackShipment)
	}
}

func (h *DeliveryHandler) PublicGetStores(c *gin.Context) {
	stores, err := h.usecase.GetPublicStores()
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.SuccessResponse(c, http.StatusOK, "Stores retrieved successfully", stores)
}

func (h *DeliveryHandler) PublicTrackShipment(c *gin.Context) {
	storeIDStr := c.Query("store_id")
	if storeIDStr == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "parameter 'store_id' diperlukan")
		return
	}

	storeID, err := uuid.Parse(storeIDStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "ID Toko tidak valid")
		return
	}

	trackingDetail, err := h.usecase.TrackShipment(storeID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Tracking details retrieved successfully", trackingDetail)
}

func (h *DeliveryHandler) UpdateSpares(c *gin.Context) {
	id := uuid.MustParse(c.Param("id"))
	var req struct {
		Spares []usecase.SpareItemRequest `json:"spares" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.usecase.UpdateBatchSpares(id, req.Spares); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Spares updated successfully"})
}

func (h *DeliveryHandler) GetPendingSubsequentItems(c *gin.Context) {
	branchIDStr := c.Query("branch_id")
	var branchID *uuid.UUID
	if branchIDStr != "" {
		bid, err := uuid.Parse(branchIDStr)
		if err == nil {
			branchID = &bid
		}
	}

	items, err := h.usecase.GetPendingSubsequentItems(branchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Pending subsequent items fetched", items)
}

func (h *DeliveryHandler) CreateManualSubsequentItem(c *gin.Context) {
	var req usecase.CreateManualSubsequentItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	item, err := h.usecase.CreateManualSubsequentItem(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Manual subsequent item created successfully", item)
}

func (h *DeliveryHandler) SlipSubsequentItem(c *gin.Context) {
	batchID := uuid.MustParse(c.Param("id"))
	subItemID := uuid.MustParse(c.Param("sub_id"))

	if err := h.usecase.SlipSubsequentItem(batchID, subItemID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Subsequent item slipped into batch successfully"})
}

func (h *DeliveryHandler) GetDriverLoad(c *gin.Context) {
	// Resolve employee_id from userID
	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "driver identity not found"})
		return
	}
	userID := userIDStr.(uuid.UUID)
	employee, err := h.employeeUseCase.GetEmployeeByUserID(userID)
	if err != nil || employee == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "driver identity not found"})
		return
	}

	load, err := h.usecase.GetDriverLoad(employee.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": load})
}

// ForceInvoiceBatch membuat/memperbaiki faktur penjualan (SalesTransaction) dari
// Surat Jalan berstatus COMPLETED yang menggunakan driver harian.
// Admin nota menginputkan payment method & total per nota berdasarkan
// laporan kertas harian yang dibawa driver.
func (h *DeliveryHandler) ForceInvoiceBatch(c *gin.Context) {
	batchID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID Surat Jalan tidak valid"})
		return
	}

	var req struct {
		CompanyID *string                           `json:"company_id"`
		Items     []usecase.ForceInvoiceItemRequest `json:"items"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var parsedCompanyID uuid.UUID
	if req.CompanyID != nil && *req.CompanyID != "" {
		if id, err := uuid.Parse(*req.CompanyID); err == nil {
			parsedCompanyID = id
		}
	}

	if err := h.usecase.ForceInvoiceBatch(batchID, parsedCompanyID, req.Items); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Faktur penjualan berhasil dibuat dan siap untuk persetujuan Admin Nota"})
}

func (h *DeliveryHandler) ReviseOrderQuantity(c *gin.Context) {
	batchID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID Surat Jalan tidak valid"})
		return
	}

	var req usecase.ReviseOrderQuantityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.usecase.ReviseOrderQuantityBeforeDelivery(batchID, req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Kuantitas pesanan berhasil direvisi, SO dan Faktur telah disesuaikan"})
}

func (h *DeliveryHandler) GetEligibleOrdersForBatch(c *gin.Context) {
	batchID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID Surat Jalan tidak valid"})
		return
	}

	search := c.Query("search")
	orders, err := h.usecase.GetEligibleOrdersForBatch(batchID, search)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Daftar pesanan yang dapat disusulkan berhasil dimuat", orders)
}

func (h *DeliveryHandler) AppendOrdersToBatch(c *gin.Context) {
	batchID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID Surat Jalan tidak valid"})
		return
	}

	var req usecase.AppendOrdersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.usecase.AppendOrdersToBatch(batchID, req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Pesanan susulan berhasil ditambahkan ke Surat Jalan", nil)
}

