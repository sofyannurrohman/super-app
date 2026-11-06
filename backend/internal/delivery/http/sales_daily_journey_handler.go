package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"example.com/hris/backend/internal/usecase"
)

type SalesDailyJourneyHandler struct {
	journeyUseCase usecase.SalesDailyJourneyUseCase
}

func NewSalesDailyJourneyHandler(journeyUseCase usecase.SalesDailyJourneyUseCase) *SalesDailyJourneyHandler {
	return &SalesDailyJourneyHandler{
		journeyUseCase: journeyUseCase,
	}
}

func (h *SalesDailyJourneyHandler) RegisterRoutes(router *gin.RouterGroup) {
	for _, basePath := range []string{"/sales/journey", "/sales/journeys"} {
		journey := router.Group(basePath)
		{
			journey.GET("/today", h.GetToday)
			journey.GET("/history", h.GetHistory)
			journey.POST("/start", h.StartJourney)
			journey.POST("/end", h.EndJourney)
			journey.PATCH("/:id", h.SelfUpdateJourney)
		}
	}

	adminSales := router.Group("/admin/sales/journeys")
	{
		adminSales.GET("", h.GetAllJourneys)
		adminSales.POST("", h.AdminCreateJourney)
		adminSales.PUT("/:id", h.AdminUpdateJourney)
		adminSales.DELETE("/:id", h.AdminDeleteJourney)
		adminSales.POST("/:id/bbm", h.AdminAddBBM)
		adminSales.DELETE("/bbm/:reimbursement_id", h.AdminDeleteBBM)
	}
}

func (h *SalesDailyJourneyHandler) AdminCreateJourney(c *gin.Context) {
	var req usecase.AdminCreateJourneyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	journey, err := h.journeyUseCase.AdminCreateJourney(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status":  "success",
		"message": "Perjalanan berhasil dicatat oleh admin",
		"data":    journey,
	})
}

func (h *SalesDailyJourneyHandler) AdminUpdateJourney(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid journey ID"})
		return
	}

	var req usecase.AdminUpdateJourneyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	journey, err := h.journeyUseCase.AdminUpdateJourney(c.Request.Context(), id, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Perjalanan berhasil diperbarui",
		"data":    journey,
	})
}

func (h *SalesDailyJourneyHandler) AdminDeleteJourney(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid journey ID"})
		return
	}

	if err := h.journeyUseCase.AdminDeleteJourney(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Perjalanan berhasil dihapus",
	})
}

func (h *SalesDailyJourneyHandler) AdminAddBBM(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid journey ID"})
		return
	}

	var req usecase.AdminAddBBMRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	bbm, err := h.journeyUseCase.AdminAddBBM(c.Request.Context(), id, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status":  "success",
		"message": "Klaim BBM berhasil ditambahkan",
		"data":    bbm,
	})
}

func (h *SalesDailyJourneyHandler) AdminDeleteBBM(c *gin.Context) {
	reimbIDStr := c.Param("reimbursement_id")
	reimbID, err := uuid.Parse(reimbIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid reimbursement ID"})
		return
	}

	if err := h.journeyUseCase.AdminDeleteBBM(c.Request.Context(), reimbID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Klaim BBM berhasil dihapus",
	})
}

func (h *SalesDailyJourneyHandler) GetHistory(c *gin.Context) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID := userIDStr.(uuid.UUID)

	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	journeys, err := h.journeyUseCase.GetMyJourneys(c.Request.Context(), userID, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch journey history: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Journey history retrieved successfully",
		"data":    journeys,
	})
}

func (h *SalesDailyJourneyHandler) GetAllJourneys(c *gin.Context) {
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	employeeIDStr := c.Query("employee_id")
	branchIDStr := c.Query("branch_id")

	var employeeID *uuid.UUID
	if employeeIDStr != "" && employeeIDStr != "ALL" {
		if id, err := uuid.Parse(employeeIDStr); err == nil {
			employeeID = &id
		}
	}

	var branchID *uuid.UUID
	if branchIDStr != "" && branchIDStr != "ALL_BRANCHES" {
		if id, err := uuid.Parse(branchIDStr); err == nil {
			branchID = &id
		}
	}

	journeys, err := h.journeyUseCase.GetAllJourneys(c.Request.Context(), startDate, endDate, employeeID, branchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch journeys: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Journeys retrieved successfully",
		"data":    journeys,
	})
}

func (h *SalesDailyJourneyHandler) GetToday(c *gin.Context) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID := userIDStr.(uuid.UUID)

	journey, err := h.journeyUseCase.GetTodayJourney(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Journey not found", "data": nil})
		return
	}

	c.JSON(http.StatusOK, journey)
}

type startJourneyRequest struct {
	StartKM       float64 `json:"start_km" binding:"required"`
	StartPhotoURL string  `json:"start_photo_url"`
	StartLocation string  `json:"start_location"`
}

func (h *SalesDailyJourneyHandler) StartJourney(c *gin.Context) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID := userIDStr.(uuid.UUID)

	var req startJourneyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	journey, err := h.journeyUseCase.StartJourney(c.Request.Context(), userID, req.StartKM, req.StartPhotoURL, req.StartLocation)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "message": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, journey)
}

type endJourneyRequest struct {
	EndKM       float64 `json:"end_km" binding:"required"`
	EndPhotoURL string  `json:"end_photo_url"`
	EndLocation string  `json:"end_location"`
}

func (h *SalesDailyJourneyHandler) EndJourney(c *gin.Context) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID := userIDStr.(uuid.UUID)

	var req endJourneyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	journey, err := h.journeyUseCase.EndJourney(c.Request.Context(), userID, req.EndKM, req.EndPhotoURL, req.EndLocation)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, journey)
}

type selfUpdateJourneyRequest struct {
	StartKM       *float64 `json:"start_km"`
	StartPhotoURL string   `json:"start_photo_url"`
	StartLocation string   `json:"start_location"`
	EndKM         *float64 `json:"end_km"`
	EndPhotoURL   string   `json:"end_photo_url"`
	EndLocation   string   `json:"end_location"`
}

func (h *SalesDailyJourneyHandler) SelfUpdateJourney(c *gin.Context) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID := userIDStr.(uuid.UUID)

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid journey ID"})
		return
	}

	var req selfUpdateJourneyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	journey, err := h.journeyUseCase.SelfUpdateJourney(
		c.Request.Context(),
		userID,
		id,
		req.StartKM,
		req.StartPhotoURL,
		req.StartLocation,
		req.EndKM,
		req.EndPhotoURL,
		req.EndLocation,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Data perjalanan berhasil diperbarui",
		"data":    journey,
	})
}

