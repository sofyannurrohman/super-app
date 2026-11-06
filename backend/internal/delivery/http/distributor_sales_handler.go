package http

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"example.com/hris/backend/internal/usecase"
	"example.com/hris/backend/pkg/utils"
)

type DistributorSalesHandler struct {
	usecase usecase.DistributorSalesUsecase
}

func NewDistributorSalesHandler(usecase usecase.DistributorSalesUsecase) *DistributorSalesHandler {
	return &DistributorSalesHandler{usecase: usecase}
}

func (h *DistributorSalesHandler) SetupRoutes(r *gin.RouterGroup) {
	dist := r.Group("/distributor")
	{
		dist.GET("/stores", h.GetStores)
		dist.POST("/stores", h.CreateStore)
		dist.POST("/visits", h.RecordVisit)
		dist.GET("/visits", h.GetVisits)
		dist.POST("/sales", h.CreateSale)
		dist.GET("/sales", h.GetSales)
		dist.POST("/sales/import-excel", h.ImportSalesExcel)
		dist.GET("/sales/template-excel", h.DownloadExcelTemplate)
		dist.GET("/summary", h.GetSummary)
	}
}

func (h *DistributorSalesHandler) resolveDistributorID(c *gin.Context) (uuid.UUID, error) {
	// 1. From context (set by AuthMiddleware for distributor users)
	if distIDStr, exists := c.Get("distributor_store_id"); exists && distIDStr != "" {
		if parsed, err := uuid.Parse(distIDStr.(string)); err == nil && parsed != uuid.Nil {
			return parsed, nil
		}
	}
	// 2. From query param (for superadmin or managers viewing distributor reports)
	if q := c.Query("distributor_store_id"); q != "" {
		if parsed, err := uuid.Parse(q); err == nil && parsed != uuid.Nil {
			return parsed, nil
		}
	}
	if q := c.Query("distributor_id"); q != "" {
		if parsed, err := uuid.Parse(q); err == nil && parsed != uuid.Nil {
			return parsed, nil
		}
	}

	return uuid.Nil, fmt.Errorf("identitas distributor tidak ditemukan pada akun Anda atau parameter request")
}

func (h *DistributorSalesHandler) GetStores(c *gin.Context) {
	distID, err := h.resolveDistributorID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	search := c.Query("search")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	stores, total, err := h.usecase.GetDistributorStores(distID, search, page, limit)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Berhasil mengambil daftar toko binaan", gin.H{
		"stores": stores,
		"total":  total,
		"page":   page,
		"limit":  limit,
	})
}

func (h *DistributorSalesHandler) CreateStore(c *gin.Context) {
	distID, err := h.resolveDistributorID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	var req usecase.CreateDistributorStoreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	var userID *uuid.UUID
	if uID, exists := c.Get("userID"); exists {
		parsed := uID.(uuid.UUID)
		userID = &parsed
	}

	store, err := h.usecase.CreateDistributorStore(distID, userID, req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Toko retail berhasil didaftarkan", store)
}

func (h *DistributorSalesHandler) RecordVisit(c *gin.Context) {
	distID, err := h.resolveDistributorID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	var req struct {
		StoreID   uuid.UUID `json:"store_id" binding:"required"`
		Latitude  float64   `json:"latitude"`
		Longitude float64   `json:"longitude"`
		SelfieURL string    `json:"selfie_url"`
		Notes     string    `json:"notes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	var userID *uuid.UUID
	if uID, exists := c.Get("userID"); exists {
		parsed := uID.(uuid.UUID)
		userID = &parsed
	}

	var employeeID *uuid.UUID
	if empIDStr, exists := c.Get("employee_id"); exists && empIDStr != "" {
		if parsed, err := uuid.Parse(empIDStr.(string)); err == nil && parsed != uuid.Nil {
			employeeID = &parsed
		}
	}

	visit, err := h.usecase.RecordDistributorVisit(distID, req.StoreID, userID, employeeID, req.Latitude, req.Longitude, req.SelfieURL, req.Notes)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Kunjungan berhasil dicatat", visit)
}

func (h *DistributorSalesHandler) GetVisits(c *gin.Context) {
	distID, err := h.resolveDistributorID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	var startDate, endDate *time.Time
	if s := c.Query("start_date"); s != "" {
		if t, err := time.Parse("2006-01-02", s); err == nil {
			startDate = &t
		}
	}
	if e := c.Query("end_date"); e != "" {
		if t, err := time.Parse("2006-01-02", e); err == nil {
			end := t.Add(24*time.Hour - time.Nanosecond)
			endDate = &end
		}
	}

	visits, total, err := h.usecase.GetDistributorVisits(distID, startDate, endDate, page, limit)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Berhasil mengambil daftar kunjungan", gin.H{
		"visits": visits,
		"total":  total,
		"page":   page,
		"limit":  limit,
	})
}

func (h *DistributorSalesHandler) CreateSale(c *gin.Context) {
	distID, err := h.resolveDistributorID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	var req usecase.CreateDistributorSaleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	var userID *uuid.UUID
	if uID, exists := c.Get("userID"); exists {
		parsed := uID.(uuid.UUID)
		userID = &parsed
	}

	var employeeID *uuid.UUID
	if empIDStr, exists := c.Get("employee_id"); exists && empIDStr != "" {
		if parsed, err := uuid.Parse(empIDStr.(string)); err == nil && parsed != uuid.Nil {
			employeeID = &parsed
		}
	}

	sale, err := h.usecase.CreateDistributorSale(distID, userID, employeeID, req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Transaksi penjualan distributor berhasil dicatat", sale)
}

func (h *DistributorSalesHandler) GetSales(c *gin.Context) {
	distID, err := h.resolveDistributorID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	var startDate, endDate *time.Time
	if s := c.Query("start_date"); s != "" {
		if t, err := time.Parse("2006-01-02", s); err == nil {
			startDate = &t
		}
	}
	if e := c.Query("end_date"); e != "" {
		if t, err := time.Parse("2006-01-02", e); err == nil {
			end := t.Add(24*time.Hour - time.Nanosecond)
			endDate = &end
		}
	}

	sales, total, err := h.usecase.GetDistributorSales(distID, startDate, endDate, page, limit)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Berhasil mengambil riwayat penjualan distributor", gin.H{
		"sales": sales,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

func (h *DistributorSalesHandler) ImportSalesExcel(c *gin.Context) {
	distID, err := h.resolveDistributorID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	file, _, err := c.Request.FormFile("file")
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "File Excel (.xlsx) wajib diunggah")
		return
	}
	defer file.Close()

	var userID *uuid.UUID
	if uID, exists := c.Get("userID"); exists {
		parsed := uID.(uuid.UUID)
		userID = &parsed
	}

	result, err := h.usecase.ImportSellOutExcel(distID, userID, file)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Proses impor selesai", result)
}

func (h *DistributorSalesHandler) DownloadExcelTemplate(c *gin.Context) {
	bytes, err := h.usecase.GenerateExcelTemplate()
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal membuat template: "+err.Error())
		return
	}

	filename := fmt.Sprintf("Template_Penjualan_Distributor_%s.xlsx", time.Now().Format("20060102"))
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", bytes)
}

func (h *DistributorSalesHandler) GetSummary(c *gin.Context) {
	distID, err := h.resolveDistributorID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	month, _ := strconv.Atoi(c.DefaultQuery("month", strconv.Itoa(int(time.Now().Month()))))
	year, _ := strconv.Atoi(c.DefaultQuery("year", strconv.Itoa(time.Now().Year())))

	summary, err := h.usecase.GetDistributorSummary(distID, month, year)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Berhasil memuat ringkasan distributor", summary)
}
