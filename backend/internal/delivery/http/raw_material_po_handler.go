package http

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"example.com/hris/backend/internal/usecase"
)

type RawMaterialPOHandler struct {
	usecase usecase.RawMaterialPOUsecase
}

func NewRawMaterialPOHandler(u usecase.RawMaterialPOUsecase) *RawMaterialPOHandler {
	return &RawMaterialPOHandler{usecase: u}
}

func (h *RawMaterialPOHandler) RegisterRoutes(r *gin.RouterGroup) {
	group := r.Group("/factory/raw-materials")
	{
		// Purchase Requests (PR) - Permintaan Bahan Baku oleh Admin Gudang
		group.GET("/requests", h.ListPRs)
		group.POST("/requests", h.CreatePR)
		group.GET("/requests/:id", h.GetPRDetail)
		group.PUT("/requests/:id", h.UpdatePR)
		group.POST("/requests/:id/reject", RoleMiddleware(string(domain.RoleSuperAdmin), string(domain.RoleAdminPurchasing), string(domain.RoleManager)), h.RejectPR)
		group.POST("/requests/:id/complete", RoleMiddleware(string(domain.RoleSuperAdmin), string(domain.RoleAdminPurchasing), string(domain.RoleManager)), h.CompletePR)
		group.DELETE("/requests/:id", h.DeletePR)

		// Purchase Orders (PO)
		// Nilai PO dan harga bahan baku adalah informasi purchasing. Pembatasan
		// diletakkan di API (bukan hanya menu frontend) agar tidak bisa dibypass oleh Admin Gudang.
		purchasingOnly := RoleMiddleware(string(domain.RoleSuperAdmin), string(domain.RoleAdminPurchasing), string(domain.RoleManager))
		group.GET("/orders", purchasingOnly, h.ListPOs)
		group.POST("/orders", purchasingOnly, h.CreatePO)
		group.GET("/orders/:id", purchasingOnly, h.GetPODetail)
		group.PUT("/orders/:id", purchasingOnly, h.UpdatePO)
		group.DELETE("/orders/:id", purchasingOnly, h.DeletePO)
		group.POST("/orders/:id/approve-spv", h.ApprovePOBySPV)
		group.POST("/orders/:id/approve-purchasing", purchasingOnly, h.ApprovePOByPurchasing)
		group.PUT("/orders/:id/progress", purchasingOnly, h.UpdatePOProgress)
		group.PUT("/orders/:id/payment", purchasingOnly, h.UpdatePOPaymentStatus)
		group.POST("/orders/:id/reject", purchasingOnly, h.RejectPO)

		// Receipts (Penerimaan Bahan Baku Supplier)
		group.GET("/receipts", h.ListReceipts)
		group.POST("/receipts", h.CreateReceipt)
		group.GET("/receipts/:id", h.GetReceiptDetail)
		group.POST("/receipts/:id/pay", h.PayReceipt)
		group.POST("/receipts/:id/pay/approve", RoleMiddleware(string(domain.RoleSuperAdmin), string(domain.RoleCashierPusat)), h.ApproveReceiptPayment)
	}
}

func (h *RawMaterialPOHandler) PayReceipt(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid receipt id"})
		return
	}
	var req struct {
		Amount float64 `json:"amount" binding:"required,gt=0"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.usecase.PayReceipt(id, req.Amount); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "pembayaran faktur bahan berhasil dicatat"})
}

func (h *RawMaterialPOHandler) ApproveReceiptPayment(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid receipt id"})
		return
	}
	if err := h.usecase.ApproveReceiptPayment(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "pembayaran supplier disetujui dan kas telah dipotong"})
}

func (h *RawMaterialPOHandler) ListPOs(c *gin.Context) {
	companyID := uuid.Nil
	if val, ok := c.Get("companyID"); ok && val != nil {
		if id, ok := val.(uuid.UUID); ok {
			companyID = id
		}
	}

	var factoryID uuid.UUID
	if fStr := c.Query("factory_id"); fStr != "" {
		if fID, err := uuid.Parse(fStr); err == nil {
			factoryID = fID
		}
	}

	status := c.Query("status")

	pos, err := h.usecase.ListPOs(companyID, factoryID, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": pos})
}

func (h *RawMaterialPOHandler) CreatePO(c *gin.Context) {
	var req usecase.CreateRawMaterialPORequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	companyID := uuid.Nil
	if val, ok := c.Get("companyID"); ok && val != nil {
		if id, ok := val.(uuid.UUID); ok {
			companyID = id
		}
	}

	var userID *uuid.UUID
	if val, ok := c.Get("userID"); ok && val != nil {
		if uID, ok := val.(uuid.UUID); ok {
			userID = &uID
		}
	}

	userName := ""
	if val, ok := c.Get("name"); ok && val != nil {
		userName = val.(string)
	}

	po, err := h.usecase.CreatePO(req, companyID, userID, userName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Purchase Order Bahan Baku berhasil dibuat",
		"data":    po,
	})
}

func (h *RawMaterialPOHandler) GetPODetail(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	po, err := h.usecase.GetPOByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Purchase Order tidak ditemukan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": po})
}

func (h *RawMaterialPOHandler) UpdatePO(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	var req usecase.CreateRawMaterialPORequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	po, err := h.usecase.UpdatePO(id, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Purchase Order berhasil diperbarui",
		"data":    po,
	})
}

func (h *RawMaterialPOHandler) DeletePO(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	if err := h.usecase.DeletePO(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Purchase Order berhasil dihapus"})
}

type ApprovalRequest struct {
	Notes string `json:"notes"`
}

func (h *RawMaterialPOHandler) ApprovePOBySPV(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	var req ApprovalRequest
	_ = c.ShouldBindJSON(&req)

	var userID *uuid.UUID
	if val, ok := c.Get("userID"); ok && val != nil {
		if uID, ok := val.(uuid.UUID); ok {
			userID = &uID
		}
	}

	userName := ""
	if val, ok := c.Get("name"); ok && val != nil {
		userName = val.(string)
	}

	userPosition := "Supervisor Gudang"
	if val, ok := c.Get("job_position"); ok && val != nil && val.(string) != "" {
		userPosition = val.(string)
	} else if rVal, ok := c.Get("role"); ok && rVal != nil {
		userPosition = strings.ToUpper(strings.ReplaceAll(rVal.(string), "_", " "))
	}

	po, err := h.usecase.ApprovePOBySPV(id, userID, userName, userPosition, req.Notes)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Purchase Order berhasil disetujui oleh SPV Gudang Pabrik",
		"data":    po,
	})
}

func (h *RawMaterialPOHandler) ApprovePOByPurchasing(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	var req ApprovalRequest
	_ = c.ShouldBindJSON(&req)

	var userID *uuid.UUID
	if val, ok := c.Get("userID"); ok && val != nil {
		if uID, ok := val.(uuid.UUID); ok {
			userID = &uID
		}
	}

	userName := ""
	if val, ok := c.Get("name"); ok && val != nil {
		userName = val.(string)
	}

	userPosition := "Admin Purchasing"
	if val, ok := c.Get("job_position"); ok && val != nil && val.(string) != "" {
		userPosition = val.(string)
	} else if rVal, ok := c.Get("role"); ok && rVal != nil {
		userPosition = strings.ToUpper(strings.ReplaceAll(rVal.(string), "_", " "))
	}

	po, err := h.usecase.ApprovePOByPurchasing(id, userID, userName, userPosition, req.Notes)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Purchase Order resmi disetujui oleh Purchasing",
		"data":    po,
	})
}

func (h *RawMaterialPOHandler) UpdatePOProgress(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	var req usecase.UpdatePOProgressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var userID *uuid.UUID
	if val, ok := c.Get("userID"); ok && val != nil {
		if uID, ok := val.(uuid.UUID); ok {
			userID = &uID
		}
	}

	userName := ""
	if val, ok := c.Get("name"); ok && val != nil {
		userName = val.(string)
	}

	po, err := h.usecase.UpdatePOProgress(id, req, userID, userName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Progress Purchase Order berhasil diperbarui",
		"data":    po,
	})
}

func (h *RawMaterialPOHandler) UpdatePOPaymentStatus(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	var req usecase.UpdatePOPaymentStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	po, err := h.usecase.UpdatePOPaymentStatus(id, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Status pembayaran berhasil diperbarui",
		"data":    po,
	})
}

type RejectionRequest struct {
	Reason string `json:"reason" binding:"required"`
}

func (h *RawMaterialPOHandler) RejectPO(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	var req RejectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Alasan penolakan wajib diisi"})
		return
	}

	var userID *uuid.UUID
	if val, ok := c.Get("userID"); ok && val != nil {
		if uID, ok := val.(uuid.UUID); ok {
			userID = &uID
		}
	}

	userName := ""
	if val, ok := c.Get("name"); ok && val != nil {
		userName = val.(string)
	}

	po, err := h.usecase.RejectPO(id, userID, userName, req.Reason)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Purchase Order telah ditolak",
		"data":    po,
	})
}

func (h *RawMaterialPOHandler) ListReceipts(c *gin.Context) {
	companyID := uuid.Nil
	if val, ok := c.Get("companyID"); ok && val != nil {
		if id, ok := val.(uuid.UUID); ok {
			companyID = id
		}
	}

	var factoryID uuid.UUID
	if fStr := c.Query("factory_id"); fStr != "" {
		if fID, err := uuid.Parse(fStr); err == nil {
			factoryID = fID
		}
	}

	receipts, err := h.usecase.ListReceipts(companyID, factoryID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": receipts})
}

func (h *RawMaterialPOHandler) CreateReceipt(c *gin.Context) {
	var req usecase.CreateRawMaterialReceiptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	companyID := uuid.Nil
	if val, ok := c.Get("companyID"); ok && val != nil {
		if id, ok := val.(uuid.UUID); ok {
			companyID = id
		}
	}

	var userID *uuid.UUID
	if val, ok := c.Get("userID"); ok && val != nil {
		if uID, ok := val.(uuid.UUID); ok {
			userID = &uID
		}
	}

	userName := ""
	if val, ok := c.Get("name"); ok && val != nil {
		userName = val.(string)
	}

	receipt, err := h.usecase.CreateReceipt(req, companyID, userID, userName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Penerimaan Bahan Baku berhasil dicatat dan stok pabrik diperbarui",
		"data":    receipt,
	})
}

func (h *RawMaterialPOHandler) GetReceiptDetail(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	receipt, err := h.usecase.GetReceiptByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Penerimaan tidak ditemukan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": receipt})
}

// Purchase Requests (PR) Handlers

func (h *RawMaterialPOHandler) ListPRs(c *gin.Context) {
	companyID := uuid.Nil
	if val, ok := c.Get("companyID"); ok && val != nil {
		if id, ok := val.(uuid.UUID); ok {
			companyID = id
		}
	}

	var factoryID uuid.UUID
	if fStr := c.Query("factory_id"); fStr != "" {
		if fID, err := uuid.Parse(fStr); err == nil {
			factoryID = fID
		}
	}

	status := c.Query("status")

	prs, err := h.usecase.ListPRs(companyID, factoryID, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if c.Query("mine") == "true" {
		userID, ok := c.Get("userID")
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Pengguna tidak ditemukan"})
			return
		}
		id, ok := userID.(uuid.UUID)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Pengguna tidak ditemukan"})
			return
		}
		mine := make([]domain.RawMaterialPurchaseRequest, 0)
		for _, pr := range prs {
			if pr.CreatedByID != nil && *pr.CreatedByID == id {
				mine = append(mine, pr)
			}
		}
		prs = mine
	}

	c.JSON(http.StatusOK, gin.H{"data": prs})
}

func (h *RawMaterialPOHandler) CompletePR(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID Purchase Request tidak valid"})
		return
	}
	var req struct {
		Notes string `json:"notes" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.Notes) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Catatan tindak lanjut wajib diisi"})
		return
	}
	var userID *uuid.UUID
	if val, ok := c.Get("userID"); ok {
		if id, ok := val.(uuid.UUID); ok {
			userID = &id
		}
	}
	name, _ := c.Get("name")
	userName, _ := name.(string)
	pr, err := h.usecase.CompletePR(id, userID, userName, req.Notes)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": pr})
}

func (h *RawMaterialPOHandler) CreatePR(c *gin.Context) {
	var req usecase.CreateRawMaterialPRRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	companyID := uuid.Nil
	if val, ok := c.Get("companyID"); ok && val != nil {
		if id, ok := val.(uuid.UUID); ok {
			companyID = id
		}
	}

	var userID *uuid.UUID
	if val, ok := c.Get("userID"); ok && val != nil {
		if uID, ok := val.(uuid.UUID); ok {
			userID = &uID
		}
	}

	userName := ""
	if val, ok := c.Get("name"); ok && val != nil {
		userName = val.(string)
	}

	pr, err := h.usecase.CreatePR(req, companyID, userID, userName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Purchase Request bahan baku berhasil diajukan ke Purchasing",
		"data":    pr,
	})
}

func (h *RawMaterialPOHandler) GetPRDetail(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID Purchase Request tidak valid"})
		return
	}

	pr, err := h.usecase.GetPRByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Purchase Request tidak ditemukan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": pr})
}

func (h *RawMaterialPOHandler) UpdatePR(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID Purchase Request tidak valid"})
		return
	}

	var req usecase.CreateRawMaterialPRRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	pr, err := h.usecase.UpdatePR(id, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Purchase Request berhasil diperbarui",
		"data":    pr,
	})
}

func (h *RawMaterialPOHandler) RejectPR(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID Purchase Request tidak valid"})
		return
	}

	var req struct {
		Reason string `json:"reason" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var userID *uuid.UUID
	if val, ok := c.Get("userID"); ok && val != nil {
		if uID, ok := val.(uuid.UUID); ok {
			userID = &uID
		}
	}

	userName := ""
	if val, ok := c.Get("name"); ok && val != nil {
		userName = val.(string)
	}

	pr, err := h.usecase.RejectPR(id, userID, userName, req.Reason)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Purchase Request telah ditolak",
		"data":    pr,
	})
}

func (h *RawMaterialPOHandler) DeletePR(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID Purchase Request tidak valid"})
		return
	}

	if err := h.usecase.DeletePR(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Purchase Request berhasil dihapus"})
}
