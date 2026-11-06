package http

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"example.com/hris/backend/internal/usecase"
)

type GpsHandler struct {
	usecase usecase.GpsUsecase
}

func NewGpsHandler(u usecase.GpsUsecase) *GpsHandler {
	return &GpsHandler{u}
}

func (h *GpsHandler) RegisterRoutes(r *gin.RouterGroup) {
	gps := r.Group("/vehicles")
	{
		gps.GET("/locations/live", h.GetLiveLocations)
		gps.GET("/:id/route-history", h.GetRouteHistory)
		gps.POST("/:id/gps-device", h.RegisterDevice) // Link GPS to Vehicle
	}
}

func (h *GpsHandler) GetLiveLocations(c *gin.Context) {
	// get companyID from context
	val, exists := c.Get("companyID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "company ID not found"})
		return
	}
	companyID := val.(uuid.UUID)

	locations, err := h.usecase.GetLiveLocations(companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, locations)
}

func (h *GpsHandler) GetRouteHistory(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid vehicle ID format"})
		return
	}

	// Default to today if date is not provided
	dateStr := c.Query("date")
	queryDate := time.Now()
	if dateStr != "" {
		parsed, err := time.Parse("2006-01-02", dateStr)
		if err == nil {
			queryDate = parsed
		}
	}

	locations, err := h.usecase.GetRouteHistory(id, queryDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, locations)
}

func (h *GpsHandler) RegisterDevice(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid vehicle ID format"})
		return
	}

	var req struct {
		IMEI      string `json:"imei" binding:"required"`
		SimNumber string `json:"sim_number"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.usecase.RegisterDevice(id, req.IMEI, req.SimNumber); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "GPS device linked successfully"})
}
