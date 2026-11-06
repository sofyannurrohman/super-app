package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"example.com/hris/backend/internal/usecase"
)

type MTDeliveryHandler struct {
	uc usecase.MTDeliveryUsecase
}

func NewMTDeliveryHandler(uc usecase.MTDeliveryUsecase) *MTDeliveryHandler {
	return &MTDeliveryHandler{uc}
}

func (h *MTDeliveryHandler) RegisterRoutes(r *gin.RouterGroup) {
	factoryRoles := RoleMiddleware(
		string(domain.RoleSuperAdmin),
		string(domain.RoleHRAdmin),
		string(domain.RoleManager),
		string(domain.RoleAdminGudang),
	)

	mt := r.Group("/mt-delivery")
	mt.Use(factoryRoles)
	{
		// ── Dibuat & dikelola dari PABRIK/PUSAT ──────────────────────────────
		mt.GET("", h.ListByCompany)              // List semua DO perusahaan (view pabrik)
		mt.POST("", h.CreateDO)                  // Buat DO baru
		mt.GET("/:id", h.GetDO)                  // Detail DO
		mt.PUT("/:id", h.UpdateDO)               // Edit DO (DRAFT/ASSIGNED only)
		mt.POST("/:id/assign", h.AssignToBranch) // Assign ke cabang
		mt.POST("/:id/cancel", h.CancelDO)       // Batalkan DO
		mt.DELETE("/:id", h.DeleteDO)            // Hapus DO

		// ── Diakses dari CABANG ───────────────────────────────────────────────
		mt.GET("/branch/inbox", h.ListByBranch)    // List DO yang di-assign ke cabang ini
		mt.POST("/:id/approve", h.ApproveDO)       // Cabang approve
		mt.POST("/:id/reject", h.RejectDO)         // Cabang reject
		mt.POST("/:id/loading", h.StartLoading)    // Mulai muat barang
		mt.POST("/:id/ship", h.ShipDO)             // Kirim barang
		mt.POST("/:id/delivered", h.MarkDelivered) // Tandai sudah diterima MT
	}
}

// ─── Factory / Pusat Handlers ─────────────────────────────────────────────────

func (h *MTDeliveryHandler) ListByCompany(c *gin.Context) {
	companyID := getCompanyID(c)
	orders, err := h.uc.ListByCompany(companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": orders})
}

func (h *MTDeliveryHandler) GetDO(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}
	order, err := h.uc.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "DO tidak ditemukan"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": order})
}

func (h *MTDeliveryHandler) CreateDO(c *gin.Context) {
	var req usecase.CreateMTDeliveryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	companyID := getCompanyID(c)
	createdBy := getEmployeeName(c)

	order, err := h.uc.CreateDO(req, companyID, createdBy)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "MT Delivery Order berhasil dibuat", "data": order})
}

func (h *MTDeliveryHandler) UpdateDO(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}
	var req usecase.CreateMTDeliveryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	order, err := h.uc.UpdateDO(id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "MT DO berhasil diperbarui", "data": order})
}

func (h *MTDeliveryHandler) AssignToBranch(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}
	var body struct {
		BranchID uuid.UUID `json:"branch_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.uc.AssignToBranch(id, body.BranchID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "DO berhasil di-assign ke cabang"})
}

func (h *MTDeliveryHandler) CancelDO(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}
	var body struct {
		Reason string `json:"reason"`
	}
	c.ShouldBindJSON(&body)
	if err := h.uc.CancelDO(id, body.Reason); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "DO berhasil dibatalkan"})
}

func (h *MTDeliveryHandler) DeleteDO(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}
	if err := h.uc.DeleteDO(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "DO berhasil dihapus"})
}

// ─── Branch / Cabang Handlers ─────────────────────────────────────────────────

func (h *MTDeliveryHandler) ListByBranch(c *gin.Context) {
	branchID := parseBranchID(c)
	orders, err := h.uc.ListByBranch(branchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": orders})
}

func (h *MTDeliveryHandler) ApproveDO(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}
	approvedBy := getEmployeeName(c)
	if err := h.uc.ApproveDO(id, approvedBy); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "DO berhasil di-approve, siapkan barang"})
}

func (h *MTDeliveryHandler) RejectDO(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}
	var body struct {
		Reason string `json:"reason" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Alasan penolakan wajib diisi"})
		return
	}
	if err := h.uc.RejectDO(id, body.Reason); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "DO berhasil ditolak"})
}

func (h *MTDeliveryHandler) StartLoading(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}
	var body struct {
		PreparedBy string                        `json:"prepared_by"`
		Items      []usecase.PreparedItemRequest `json:"items"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if body.PreparedBy == "" {
		body.PreparedBy = getEmployeeName(c)
	}
	if err := h.uc.StartLoading(id, body.PreparedBy, body.Items); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Status muat barang berhasil diperbarui"})
}

func (h *MTDeliveryHandler) ShipDO(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}
	var body struct {
		VehicleID *uuid.UUID `json:"vehicle_id"`
		DriverID  *uuid.UUID `json:"driver_id"`
		Notes     string     `json:"notes"`
	}
	c.ShouldBindJSON(&body)
	if err := h.uc.ShipDO(id, body.VehicleID, body.DriverID, body.Notes); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "DO berhasil dikirim"})
}

func (h *MTDeliveryHandler) MarkDelivered(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}
	if err := h.uc.MarkDelivered(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "DO ditandai sudah diterima MT"})
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

func getCompanyID(c *gin.Context) uuid.UUID {
	val, _ := c.Get("companyID")
	if val != nil {
		if id, ok := val.(uuid.UUID); ok {
			return id
		}
	}
	// fallback: try string
	strVal, _ := c.Get("company_id")
	if strVal != nil {
		if id, err := uuid.Parse(strVal.(string)); err == nil {
			return id
		}
	}
	return uuid.Nil
}

func getEmployeeName(c *gin.Context) string {
	val, _ := c.Get("employeeName")
	if val != nil {
		if name, ok := val.(string); ok {
			return name
		}
	}
	return ""
}
