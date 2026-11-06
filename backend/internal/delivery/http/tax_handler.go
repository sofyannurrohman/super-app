package http

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"example.com/hris/backend/internal/usecase"
)

type TaxHandler struct {
	taxUsecase usecase.TaxUsecase
}

func NewTaxHandler(taxUsecase usecase.TaxUsecase) *TaxHandler {
	return &TaxHandler{taxUsecase: taxUsecase}
}

func (h *TaxHandler) SetupRoutes(r *gin.RouterGroup) {
	taxGroup := r.Group("/tax")
	{
		taxGroup.POST("/invoice", h.SetTaxInvoice)
		taxGroup.GET("/efaktur/export", h.ExportEfaktur)
	}
}

func (h *TaxHandler) SetTaxInvoice(c *gin.Context) {
	var req domain.SetTaxInvoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.taxUsecase.SetTaxInvoice(req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Tax invoice details updated successfully"})
}

func (h *TaxHandler) ExportEfaktur(c *gin.Context) {
	var req domain.ExportEfakturRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	companyID, err := uuid.Parse(req.CompanyID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid company_id"})
		return
	}

	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start_date format, expected YYYY-MM-DD"})
		return
	}

	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end_date format, expected YYYY-MM-DD"})
		return
	}

	// Adjust end date to cover the whole day
	endDate = endDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)

	csvData, err := h.taxUsecase.GenerateEfakturCSV(companyID, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate CSV: " + err.Error()})
		return
	}

	// Format filename: efaktur_YYYYMMDD_YYYYMMDD.csv
	filename := "efaktur_" + startDate.Format("20060102") + "_" + endDate.Format("20060102") + ".csv"

	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Type", "text/csv")
	c.Header("Content-Length", string(rune(len(csvData)))) // Gin doesn't need this strictly but good practice

	c.Data(http.StatusOK, "text/csv", csvData)
}
