package http

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"example.com/hris/backend/internal/usecase"
)

// FinanceExtendedHandler — Handler untuk AR/AP Aging & Fixed Asset Management
type FinanceExtendedHandler struct {
	arAP       usecase.ARAPUsecase
	fixedAsset usecase.FixedAssetUsecase
}

func NewFinanceExtendedHandler(arAP usecase.ARAPUsecase, fixedAsset usecase.FixedAssetUsecase) *FinanceExtendedHandler {
	return &FinanceExtendedHandler{arAP: arAP, fixedAsset: fixedAsset}
}

func (h *FinanceExtendedHandler) RegisterRoutes(r *gin.RouterGroup) {
	finance := r.Group("/finance")
	finance.Use(RoleMiddleware(
		string(domain.RoleSuperAdmin),
		string(domain.RoleHRAdmin),
		string(domain.RoleManager),
		string(domain.RoleCashier),
		string(domain.RoleAdminKasir),
	))
	{
		// AR/AP Aging
		finance.GET("/ar-aging", h.GetARAgingReport)
		finance.GET("/ap-aging", h.GetAPAgingReport)

		// Fixed Assets
		finance.GET("/fixed-assets", h.ListFixedAssets)
		finance.POST("/fixed-assets", h.CreateFixedAsset)
		finance.PUT("/fixed-assets/:id", h.UpdateFixedAsset)
		finance.DELETE("/fixed-assets/:id", h.DeleteFixedAsset)
		finance.GET("/fixed-assets/:id/schedule", h.GetDepreciationSchedule)
		finance.POST("/fixed-assets/:id/depreciate", h.RunMonthlyDepreciation)
	}
}

// ---- AR/AP Aging Handlers ----

// GET /api/v1/finance/ar-aging?as_of=2025-08-12
func (h *FinanceExtendedHandler) GetARAgingReport(c *gin.Context) {
	companyID := parseFinanceCompanyID(c)
	asOf := parseAsOfDate(c)

	report, err := h.arAP.GetARAgingReport(companyID, asOf)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": report})
}

// GET /api/v1/finance/ap-aging?as_of=2025-08-12
func (h *FinanceExtendedHandler) GetAPAgingReport(c *gin.Context) {
	companyID := parseFinanceCompanyID(c)
	asOf := parseAsOfDate(c)

	report, err := h.arAP.GetAPAgingReport(companyID, asOf)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": report})
}

// ---- Fixed Asset Handlers ----

// GET /api/v1/finance/fixed-assets
func (h *FinanceExtendedHandler) ListFixedAssets(c *gin.Context) {
	companyID := parseFinanceCompanyID(c)
	assets, summary, err := h.fixedAsset.ListFixedAssets(companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": assets, "summary": summary})
}

// POST /api/v1/finance/fixed-assets
func (h *FinanceExtendedHandler) CreateFixedAsset(c *gin.Context) {
	var asset domain.FixedAsset
	if err := c.ShouldBindJSON(&asset); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	asset.CompanyID = parseFinanceCompanyID(c)
	if err := h.fixedAsset.CreateFixedAsset(&asset); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	asset.ComputeDerived()
	c.JSON(http.StatusCreated, gin.H{"data": asset, "message": "Aset tetap berhasil ditambahkan"})
}

// PUT /api/v1/finance/fixed-assets/:id
func (h *FinanceExtendedHandler) UpdateFixedAsset(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID aset tidak valid"})
		return
	}
	var asset domain.FixedAsset
	if err := c.ShouldBindJSON(&asset); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	asset.ID = id
	if err := h.fixedAsset.UpdateFixedAsset(&asset); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Aset berhasil diperbarui"})
}

// DELETE /api/v1/finance/fixed-assets/:id
func (h *FinanceExtendedHandler) DeleteFixedAsset(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID aset tidak valid"})
		return
	}
	if err := h.fixedAsset.DeleteFixedAsset(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Aset berhasil dihapus"})
}

// GET /api/v1/finance/fixed-assets/:id/schedule
func (h *FinanceExtendedHandler) GetDepreciationSchedule(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID aset tidak valid"})
		return
	}
	schedule, err := h.fixedAsset.GetDepreciationSchedule(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": schedule})
}

// POST /api/v1/finance/fixed-assets/:id/depreciate?period=2025-08
func (h *FinanceExtendedHandler) RunMonthlyDepreciation(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID aset tidak valid"})
		return
	}
	periodStr := c.DefaultQuery("period", time.Now().Format("2006-01"))
	period, err := time.Parse("2006-01", periodStr)
	if err != nil {
		period = time.Now()
	}
	if err := h.fixedAsset.RunMonthlyDepreciation(id, period); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Depresiasi bulan " + periodStr + " berhasil dijalankan"})
}

// ---- Helpers ----

func parseFinanceCompanyID(c *gin.Context) uuid.UUID {
	cidStr := c.Query("company_id")
	if cidStr != "" {
		if id, err := uuid.Parse(cidStr); err == nil {
			return id
		}
	}
	if cid, ok := c.Get("companyID"); ok {
		if id, ok2 := cid.(uuid.UUID); ok2 {
			return id
		}
	}
	return uuid.Nil
}

func parseAsOfDate(c *gin.Context) time.Time {
	asOfStr := c.DefaultQuery("as_of", time.Now().Format("2006-01-02"))
	t, err := time.Parse("2006-01-02", asOfStr)
	if err != nil {
		return time.Now()
	}
	return t
}
