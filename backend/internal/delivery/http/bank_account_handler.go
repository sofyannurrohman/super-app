package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"example.com/hris/backend/internal/usecase"
)

type BankAccountHandler struct {
	usecase usecase.BankAccountUsecase
}

func NewBankAccountHandler(u usecase.BankAccountUsecase) *BankAccountHandler {
	return &BankAccountHandler{u}
}

func (h *BankAccountHandler) RegisterRoutes(r *gin.RouterGroup) {
	ba := r.Group("/finance/bank-accounts")
	{
		// Accessible to all authenticated users (mobile salesmen, drivers, cashiers)
		ba.GET("", h.GetAll)
		ba.GET("/active", h.GetActive)

		// Admin-only routes
		admin := ba.Group("")
		admin.Use(RoleMiddleware(string(domain.RoleSuperAdmin), string(domain.RoleHRAdmin), string(domain.RoleManager), string(domain.RoleCashier), string(domain.RoleCashierPusat), string(domain.RoleAdminKasir)))
		{
			admin.POST("", h.Create)
			admin.PUT("/:id", h.Update)
			admin.DELETE("/:id", h.Delete)
			admin.POST("/:id/set-active", h.SetActive)
		}
	}
}

func (h *BankAccountHandler) GetAll(c *gin.Context) {
	var companyID *uuid.UUID
	if cid := c.Query("company_id"); cid != "" && cid != "ALL" {
		if id, err := uuid.Parse(cid); err == nil {
			companyID = &id
		}
	}
	var branchID *uuid.UUID
	if bid := c.Query("branch_id"); bid != "" && bid != "ALL_BRANCHES" {
		if id, err := uuid.Parse(bid); err == nil {
			branchID = &id
		}
	}
	accounts, err := h.usecase.GetAll(companyID, branchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Success", "data": accounts})
}

func (h *BankAccountHandler) GetActive(c *gin.Context) {
	var companyID *uuid.UUID
	if cid := c.Query("company_id"); cid != "" {
		if id, err := uuid.Parse(cid); err == nil {
			companyID = &id
		}
	}
	account, err := h.usecase.GetActive(companyID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tidak ada rekening bank aktif untuk perusahaan ini"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Success", "data": account})
}

func (h *BankAccountHandler) Create(c *gin.Context) {
	var req usecase.CreateBankAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	account, err := h.usecase.Create(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Rekening bank berhasil ditambahkan", "data": account})
}

func (h *BankAccountHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req usecase.UpdateBankAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	account, err := h.usecase.Update(id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Rekening bank berhasil diperbarui", "data": account})
}

func (h *BankAccountHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.usecase.Delete(id); err != nil {
		if errors.Is(err, domain.ErrBankAccountInUse) {
			c.JSON(http.StatusConflict, gin.H{"error": "Rekening bank tidak dapat dihapus karena sudah digunakan dalam transaksi. Nonaktifkan rekening agar riwayat transaksi tetap tersimpan."})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Rekening bank berhasil dihapus"})
}

func (h *BankAccountHandler) SetActive(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.usecase.SetActive(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Rekening bank berhasil diaktifkan sebagai rekening utama"})
}
