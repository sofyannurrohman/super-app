package http

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"example.com/hris/backend/internal/usecase"
	"example.com/hris/backend/pkg/utils"
)

type TransportHandler struct {
	usecase usecase.TransportUsecase
}

func NewTransportHandler(u usecase.TransportUsecase) *TransportHandler {
	return &TransportHandler{usecase: u}
}

func (h *TransportHandler) RegisterRoutes(r *gin.RouterGroup) {
	t := r.Group("/transport")
	{
		t.GET("/dashboard/stats", h.GetDashboardStats)

		// Trips
		t.GET("/trips", h.GetTrips)
		t.POST("/trips", h.CreateTrip)
		t.GET("/trips/:id", h.GetTripByID)
		t.PUT("/trips/:id", h.UpdateTrip)
		t.POST("/trips/:id/complete", h.CompleteTrip)
		t.DELETE("/trips/:id", h.DeleteTrip)

		// Backhauls (Pengiriman Kembali)
		t.GET("/backhauls", h.GetBackhauls)
		t.POST("/backhauls", h.CreateBackhaul)
		t.GET("/backhauls/:id", h.GetBackhaulByID)
		t.PUT("/backhauls/:id/approval", h.ApprovalBackhaul)
		t.DELETE("/backhauls/:id", h.DeleteBackhaul)

		// Expense Claims (Struk BBM & Part)
		t.GET("/claims", h.GetExpenseClaims)
		t.POST("/claims", h.CreateExpenseClaim)
		t.GET("/claims/:id", h.GetExpenseClaimByID)
		t.PUT("/claims/:id/approval", h.ApprovalExpenseClaim)
		t.PUT("/claims/:id/disburse", h.DisburseExpenseClaim)
		t.DELETE("/claims/:id", h.DeleteExpenseClaim)

		// Fleet Inspections
		t.GET("/inspections", h.GetFleetInspections)
		t.POST("/inspections", h.CreateFleetInspection)
		t.GET("/inspections/:id", h.GetFleetInspectionByID)
		t.PUT("/inspections/:id/follow-up", h.UpdateInspectionFollowUp)
		t.DELETE("/inspections/:id", h.DeleteFleetInspection)

		// Driver Attendances
		t.GET("/attendances", h.GetDriverAttendances)
		t.POST("/attendances", h.RecordDriverAttendance)
	}
}

// ──────────────── DASHBOARD STATS ────────────────
func (h *TransportHandler) GetDashboardStats(c *gin.Context) {
	var companyID *uuid.UUID
	if cidStr := c.Query("company_id"); cidStr != "" && cidStr != "all" && cidStr != "ALL" {
		if cid, err := uuid.Parse(cidStr); err == nil {
			companyID = &cid
		}
	}
	var branchID *uuid.UUID
	if bidStr := c.Query("branch_id"); bidStr != "" && bidStr != "all" && bidStr != "ALL" && bidStr != "ALL_BRANCHES" {
		if bid, err := uuid.Parse(bidStr); err == nil {
			branchID = &bid
		}
	}

	stats, err := h.usecase.GetDashboardStats(companyID, branchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, stats)
}

// ──────────────── TRIPS HANDLERS ────────────────

type CreateTripRequest struct {
	CompanyID           string                `json:"company_id" form:"company_id"`
	BranchID            string                `json:"branch_id" form:"branch_id"`
	VehicleID           string                `json:"vehicle_id" form:"vehicle_id" binding:"required"`
	DriverID            string                `json:"driver_id" form:"driver_id" binding:"required"`
	HelperID            string                `json:"helper_id" form:"helper_id"`
	OriginCity          string                `json:"origin_city" form:"origin_city" binding:"required"`
	OriginProvince      string                `json:"origin_province" form:"origin_province" binding:"required"`
	DestinationCity     string                `json:"destination_city" form:"destination_city" binding:"required"`
	DestinationProvince string                `json:"destination_province" form:"destination_province" binding:"required"`
	RegionType          domain.TripRegionType `json:"region_type" form:"region_type"`
	DepartureDate       string                `json:"departure_date" form:"departure_date" binding:"required"`
	EstimatedReturnDate string                `json:"estimated_return_date" form:"estimated_return_date" binding:"required"`
	EstimatedDays       int                   `json:"estimated_days" form:"estimated_days"`
	StartOdometer       int                   `json:"start_odometer" form:"start_odometer"`
	CargoDescription    string                `json:"cargo_description" form:"cargo_description"`
	Notes               string                `json:"notes" form:"notes"`
}

func (h *TransportHandler) CreateTrip(c *gin.Context) {
	var req CreateTripRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	vehID, err := uuid.Parse(req.VehicleID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid vehicle_id"})
		return
	}
	drvID, err := uuid.Parse(req.DriverID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid driver_id"})
		return
	}

	depDate, err := time.Parse("2006-01-02", req.DepartureDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid departure_date (YYYY-MM-DD)"})
		return
	}
	estDate, err := time.Parse("2006-01-02", req.EstimatedReturnDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid estimated_return_date (YYYY-MM-DD)"})
		return
	}

	var compID uuid.UUID
	if req.CompanyID != "" {
		compID, _ = uuid.Parse(req.CompanyID)
	}
	if compID == uuid.Nil {
		if val, exists := c.Get("companyID"); exists {
			compID = val.(uuid.UUID)
		}
	}

	trip := domain.TransportTrip{
		CompanyID:           compID,
		VehicleID:           vehID,
		DriverID:            drvID,
		OriginCity:          req.OriginCity,
		OriginProvince:      req.OriginProvince,
		DestinationCity:     req.DestinationCity,
		DestinationProvince: req.DestinationProvince,
		RegionType:          req.RegionType,
		DepartureDate:       depDate,
		EstimatedReturnDate: estDate,
		EstimatedDays:       req.EstimatedDays,
		StartOdometer:       req.StartOdometer,
		CargoDescription:    req.CargoDescription,
		Notes:               req.Notes,
		Status:              domain.TripStatusScheduled,
	}

	if req.BranchID != "" {
		if bid, err := uuid.Parse(req.BranchID); err == nil {
			trip.BranchID = &bid
		}
	}
	if req.HelperID != "" {
		if hid, err := uuid.Parse(req.HelperID); err == nil {
			trip.HelperID = &hid
		}
	}

	if err := h.usecase.CreateTrip(&trip); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, trip)
}

func (h *TransportHandler) GetTrips(c *gin.Context) {
	var companyID, branchID, driverID, vehicleID *uuid.UUID
	if val := c.Query("company_id"); val != "" && val != "ALL" && val != "all" {
		if parsed, err := uuid.Parse(val); err == nil {
			companyID = &parsed
		}
	}
	if val := c.Query("branch_id"); val != "" && val != "ALL" && val != "all" && val != "ALL_BRANCHES" {
		if parsed, err := uuid.Parse(val); err == nil {
			branchID = &parsed
		}
	}
	if val := c.Query("driver_id"); val != "" && val != "ALL" {
		if parsed, err := uuid.Parse(val); err == nil {
			driverID = &parsed
		}
	}
	if val := c.Query("vehicle_id"); val != "" && val != "ALL" {
		if parsed, err := uuid.Parse(val); err == nil {
			vehicleID = &parsed
		}
	}

	var startDate, endDate *time.Time
	if s := c.Query("start_date"); s != "" {
		if t, err := time.Parse("2006-01-02", s); err == nil {
			startDate = &t
		}
	}
	if e := c.Query("end_date"); e != "" {
		if t, err := time.Parse("2006-01-02", e); err == nil {
			endDate = &t
		}
	}

	status := c.Query("status")
	regionType := c.Query("region_type")

	trips, err := h.usecase.GetTrips(companyID, branchID, driverID, vehicleID, status, regionType, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, trips)
}

func (h *TransportHandler) GetTripByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid trip ID"})
		return
	}
	trip, err := h.usecase.GetTripByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Trip not found"})
		return
	}
	c.JSON(http.StatusOK, trip)
}

func (h *TransportHandler) UpdateTrip(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid trip ID"})
		return
	}
	existing, err := h.usecase.GetTripByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Trip not found"})
		return
	}

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if st, ok := req["status"].(string); ok && st != "" {
		existing.Status = domain.TripStatus(st)
	}
	if notes, ok := req["notes"].(string); ok {
		existing.Notes = notes
	}
	if cargo, ok := req["cargo_description"].(string); ok {
		existing.CargoDescription = cargo
	}

	if err := h.usecase.UpdateTrip(existing); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, existing)
}

func (h *TransportHandler) CompleteTrip(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid trip ID"})
		return
	}

	var req struct {
		ActualReturnDate string `json:"actual_return_date" binding:"required"`
		EndOdometer      int    `json:"end_odometer" binding:"required"`
		Notes            string `json:"notes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	retDate, err := time.Parse("2006-01-02", req.ActualReturnDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid actual_return_date (YYYY-MM-DD)"})
		return
	}

	completed, err := h.usecase.CompleteTrip(id, retDate, req.EndOdometer, req.Notes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, completed)
}

func (h *TransportHandler) DeleteTrip(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid trip ID"})
		return
	}
	if err := h.usecase.DeleteTrip(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Trip deleted successfully"})
}

// ──────────────── BACKHAULS HANDLERS ────────────────

type CreateBackhaulRequest struct {
	TripID              string  `json:"trip_id" form:"trip_id"`
	DriverID            string  `json:"driver_id" form:"driver_id" binding:"required"`
	VehicleID           string  `json:"vehicle_id" form:"vehicle_id" binding:"required"`
	OriginLocation      string  `json:"origin_location" form:"origin_location" binding:"required"`
	DestinationLocation string  `json:"destination_location" form:"destination_location" binding:"required"`
	CustomerName        string  `json:"customer_name" form:"customer_name" binding:"required"`
	CustomerPhone       string  `json:"customer_phone" form:"customer_phone"`
	CargoName           string  `json:"cargo_name" form:"cargo_name" binding:"required"`
	CargoWeightKg       float64 `json:"cargo_weight_kg" form:"cargo_weight_kg"`
	FreightCost         float64 `json:"freight_cost" form:"freight_cost" binding:"required"`
	FeePercentage       float64 `json:"fee_percentage" form:"fee_percentage"`
	AttachmentProofURL  string  `json:"attachment_proof_url" form:"attachment_proof_url"`
	Notes               string  `json:"notes" form:"notes"`
}

func (h *TransportHandler) CreateBackhaul(c *gin.Context) {
	var req CreateBackhaulRequest
	var err error
	if strings.Contains(c.GetHeader("Content-Type"), "multipart/form-data") {
		err = c.ShouldBindWith(&req, binding.FormMultipart)
	} else {
		err = c.ShouldBindJSON(&req)
	}

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	drvID, err := uuid.Parse(req.DriverID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid driver_id"})
		return
	}
	vehID, err := uuid.Parse(req.VehicleID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid vehicle_id"})
		return
	}

	// Handle upload file proof if any
	proofURL, _ := h.handleFileUpload(c, "attachment_proof", "uploads/transport/backhauls")
	if proofURL != "" {
		req.AttachmentProofURL = proofURL
	}

	backhaul := domain.TransportBackhaul{
		DriverID:            drvID,
		VehicleID:           vehID,
		OriginLocation:      req.OriginLocation,
		DestinationLocation: req.DestinationLocation,
		CustomerName:        req.CustomerName,
		CustomerPhone:       req.CustomerPhone,
		CargoName:           req.CargoName,
		CargoWeightKg:       req.CargoWeightKg,
		FreightCost:         req.FreightCost,
		FeePercentage:       req.FeePercentage,
		AttachmentProofURL:  req.AttachmentProofURL,
		Notes:               req.Notes,
	}

	if req.TripID != "" {
		if tid, err := uuid.Parse(req.TripID); err == nil {
			backhaul.TripID = &tid
		}
	}

	if err := h.usecase.CreateBackhaul(&backhaul); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, backhaul)
}

func (h *TransportHandler) GetBackhauls(c *gin.Context) {
	var tripID, driverID, vehicleID *uuid.UUID
	if val := c.Query("trip_id"); val != "" && val != "ALL" {
		if parsed, err := uuid.Parse(val); err == nil {
			tripID = &parsed
		}
	}
	if val := c.Query("driver_id"); val != "" && val != "ALL" {
		if parsed, err := uuid.Parse(val); err == nil {
			driverID = &parsed
		}
	}
	if val := c.Query("vehicle_id"); val != "" && val != "ALL" {
		if parsed, err := uuid.Parse(val); err == nil {
			vehicleID = &parsed
		}
	}

	status := c.Query("status")
	var startDate, endDate *time.Time
	if s := c.Query("start_date"); s != "" {
		if t, err := time.Parse("2006-01-02", s); err == nil {
			startDate = &t
		}
	}
	if e := c.Query("end_date"); e != "" {
		if t, err := time.Parse("2006-01-02", e); err == nil {
			endDate = &t
		}
	}

	backhauls, err := h.usecase.GetBackhauls(tripID, driverID, vehicleID, status, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, backhauls)
}

func (h *TransportHandler) GetBackhaulByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid backhaul ID"})
		return
	}
	backhaul, err := h.usecase.GetBackhaulByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Backhaul not found"})
		return
	}
	c.JSON(http.StatusOK, backhaul)
}

func (h *TransportHandler) ApprovalBackhaul(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid backhaul ID"})
		return
	}

	var req struct {
		Action domain.TransportApprovalStatus `json:"action" binding:"required"` // APPROVED or REJECTED
		Notes  string                         `json:"notes"`
		Reason string                         `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var approverID uuid.UUID
	if val, exists := c.Get("employeeID"); exists {
		approverID = val.(uuid.UUID)
	} else if val, exists := c.Get("userID"); exists {
		approverID = val.(uuid.UUID)
	}

	if req.Action == domain.TransportApprovalApproved {
		updated, err := h.usecase.ApproveBackhaul(id, approverID, req.Notes)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, updated)
	} else if req.Action == domain.TransportApprovalRejected {
		updated, err := h.usecase.RejectBackhaul(id, approverID, req.Reason)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, updated)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid action, must be APPROVED or REJECTED"})
	}
}

func (h *TransportHandler) DeleteBackhaul(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid backhaul ID"})
		return
	}
	if err := h.usecase.DeleteBackhaul(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Backhaul deleted successfully"})
}

// ──────────────── EXPENSE CLAIMS HANDLERS ────────────────

type CreateExpenseClaimRequest struct {
	TripID          string                 `json:"trip_id" form:"trip_id"`
	VehicleID       string                 `json:"vehicle_id" form:"vehicle_id" binding:"required"`
	DriverID        string                 `json:"driver_id" form:"driver_id" binding:"required"`
	Category        domain.ExpenseCategory `json:"category" form:"category" binding:"required"`
	MerchantName    string                 `json:"merchant_name" form:"merchant_name"`
	FuelLiter       *float64               `json:"fuel_liter" form:"fuel_liter"`
	CurrentOdometer int                    `json:"current_odometer" form:"current_odometer"`
	Amount          float64                `json:"amount" form:"amount" binding:"required"`
	ReceiptImageURL string                 `json:"receipt_image_url" form:"receipt_image_url"`
	Description     string                 `json:"description" form:"description"`
}

func (h *TransportHandler) CreateExpenseClaim(c *gin.Context) {
	var req CreateExpenseClaimRequest
	var err error
	if strings.Contains(c.GetHeader("Content-Type"), "multipart/form-data") {
		err = c.ShouldBindWith(&req, binding.FormMultipart)
	} else {
		err = c.ShouldBindJSON(&req)
	}

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	drvID, err := uuid.Parse(req.DriverID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid driver_id"})
		return
	}
	vehID, err := uuid.Parse(req.VehicleID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid vehicle_id"})
		return
	}

	// Upload receipt photo if provided
	receiptURL, _ := h.handleFileUpload(c, "receipt_image", "uploads/transport/claims")
	if receiptURL != "" {
		req.ReceiptImageURL = receiptURL
	}

	claim := domain.TransportExpenseClaim{
		VehicleID:       vehID,
		DriverID:        drvID,
		Category:        req.Category,
		MerchantName:    req.MerchantName,
		FuelLiter:       req.FuelLiter,
		CurrentOdometer: req.CurrentOdometer,
		Amount:          req.Amount,
		ReceiptImageURL: req.ReceiptImageURL,
		Description:     req.Description,
	}

	if req.TripID != "" {
		if tid, err := uuid.Parse(req.TripID); err == nil {
			claim.TripID = &tid
		}
	}

	if err := h.usecase.CreateExpenseClaim(&claim); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, claim)
}

func (h *TransportHandler) GetExpenseClaims(c *gin.Context) {
	var tripID, vehicleID, driverID *uuid.UUID
	if val := c.Query("trip_id"); val != "" && val != "ALL" {
		if parsed, err := uuid.Parse(val); err == nil {
			tripID = &parsed
		}
	}
	if val := c.Query("vehicle_id"); val != "" && val != "ALL" {
		if parsed, err := uuid.Parse(val); err == nil {
			vehicleID = &parsed
		}
	}
	if val := c.Query("driver_id"); val != "" && val != "ALL" {
		if parsed, err := uuid.Parse(val); err == nil {
			driverID = &parsed
		}
	}

	category := c.Query("category")
	status := c.Query("status")

	var startDate, endDate *time.Time
	if s := c.Query("start_date"); s != "" {
		if t, err := time.Parse("2006-01-02", s); err == nil {
			startDate = &t
		}
	}
	if e := c.Query("end_date"); e != "" {
		if t, err := time.Parse("2006-01-02", e); err == nil {
			endDate = &t
		}
	}

	claims, err := h.usecase.GetExpenseClaims(tripID, vehicleID, driverID, category, status, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, claims)
}

func (h *TransportHandler) GetExpenseClaimByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid claim ID"})
		return
	}
	claim, err := h.usecase.GetExpenseClaimByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Claim not found"})
		return
	}
	c.JSON(http.StatusOK, claim)
}

func (h *TransportHandler) ApprovalExpenseClaim(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid claim ID"})
		return
	}

	var req struct {
		Action domain.TransportApprovalStatus `json:"action" binding:"required"` // APPROVED or REJECTED
		Reason string                         `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var approverID uuid.UUID
	if val, exists := c.Get("employeeID"); exists {
		approverID = val.(uuid.UUID)
	} else if val, exists := c.Get("userID"); exists {
		approverID = val.(uuid.UUID)
	}

	if req.Action == domain.TransportApprovalApproved {
		updated, err := h.usecase.ApproveExpenseClaim(id, approverID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, updated)
	} else if req.Action == domain.TransportApprovalRejected {
		updated, err := h.usecase.RejectExpenseClaim(id, approverID, req.Reason)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, updated)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid action, must be APPROVED or REJECTED"})
	}
}

func (h *TransportHandler) DisburseExpenseClaim(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid claim ID"})
		return
	}

	var disburserID uuid.UUID
	if val, exists := c.Get("userID"); exists {
		disburserID = val.(uuid.UUID)
	}

	updated, err := h.usecase.DisburseExpenseClaim(id, disburserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, updated)
}

func (h *TransportHandler) DeleteExpenseClaim(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid claim ID"})
		return
	}
	if err := h.usecase.DeleteExpenseClaim(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Claim deleted successfully"})
}

// ──────────────── FLEET INSPECTIONS HANDLERS ────────────────

type CreateFleetInspectionRequest struct {
	TripID             string                `json:"trip_id" form:"trip_id"`
	VehicleID          string                `json:"vehicle_id" form:"vehicle_id" binding:"required"`
	InspectorDriverID  string                `json:"inspector_driver_id" form:"inspector_driver_id" binding:"required"`
	InspectionDate     string                `json:"inspection_date" form:"inspection_date" binding:"required"`
	ReturnOdometer     int                   `json:"return_odometer" form:"return_odometer" binding:"required"`
	BrakeCondition     string                `json:"brake_condition" form:"brake_condition"`
	TireCondition      string                `json:"tire_condition" form:"tire_condition"`
	EngineCondition    string                `json:"engine_condition" form:"engine_condition"`
	OilCondition       string                `json:"oil_condition" form:"oil_condition"`
	BatteryCondition   string                `json:"battery_condition" form:"battery_condition"`
	ElectricalLights   string                `json:"electrical_lights" form:"electrical_lights"`
	BodyCondition      string                `json:"body_condition" form:"body_condition"`
	DriverRemarks      string                `json:"driver_remarks" form:"driver_remarks"`
	PhotoConditionURLs string                `json:"photo_condition_urls" form:"photo_condition_urls"`
	FollowUpAction     domain.FollowUpAction `json:"follow_up_action" form:"follow_up_action"`
	PartsToReplace     string                `json:"parts_to_replace" form:"parts_to_replace"`
	WorkshopName       string                `json:"workshop_name" form:"workshop_name"`
	EstimatedCost      float64               `json:"estimated_cost" form:"estimated_cost"`
}

func (h *TransportHandler) CreateFleetInspection(c *gin.Context) {
	var req CreateFleetInspectionRequest
	var err error
	if strings.Contains(c.GetHeader("Content-Type"), "multipart/form-data") {
		err = c.ShouldBindWith(&req, binding.FormMultipart)
	} else {
		err = c.ShouldBindJSON(&req)
	}

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	vehID, err := uuid.Parse(req.VehicleID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid vehicle_id"})
		return
	}
	drvID, err := uuid.Parse(req.InspectorDriverID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid inspector_driver_id"})
		return
	}

	insDate, err := time.Parse("2006-01-02", req.InspectionDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid inspection_date (YYYY-MM-DD)"})
		return
	}

	photoURL, _ := h.handleFileUpload(c, "photo_condition", "uploads/transport/inspections")
	if photoURL != "" {
		req.PhotoConditionURLs = photoURL
	}

	inspection := domain.TransportFleetInspection{
		VehicleID:          vehID,
		InspectorDriverID:  drvID,
		InspectionDate:     insDate,
		ReturnOdometer:     req.ReturnOdometer,
		BrakeCondition:     req.BrakeCondition,
		TireCondition:      req.TireCondition,
		EngineCondition:    req.EngineCondition,
		OilCondition:       req.OilCondition,
		BatteryCondition:   req.BatteryCondition,
		ElectricalLights:   req.ElectricalLights,
		BodyCondition:      req.BodyCondition,
		DriverRemarks:      req.DriverRemarks,
		PhotoConditionURLs: req.PhotoConditionURLs,
		FollowUpAction:     req.FollowUpAction,
		PartsToReplace:     req.PartsToReplace,
		WorkshopName:       req.WorkshopName,
		EstimatedCost:      req.EstimatedCost,
	}

	if req.TripID != "" {
		if tid, err := uuid.Parse(req.TripID); err == nil {
			inspection.TripID = &tid
		}
	}

	if err := h.usecase.CreateFleetInspection(&inspection); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, inspection)
}

func (h *TransportHandler) GetFleetInspections(c *gin.Context) {
	var vehicleID, tripID *uuid.UUID
	if val := c.Query("vehicle_id"); val != "" && val != "ALL" {
		if parsed, err := uuid.Parse(val); err == nil {
			vehicleID = &parsed
		}
	}
	if val := c.Query("trip_id"); val != "" && val != "ALL" {
		if parsed, err := uuid.Parse(val); err == nil {
			tripID = &parsed
		}
	}

	followUpAction := c.Query("follow_up_action")
	followUpStatus := c.Query("follow_up_status")

	var startDate, endDate *time.Time
	if s := c.Query("start_date"); s != "" {
		if t, err := time.Parse("2006-01-02", s); err == nil {
			startDate = &t
		}
	}
	if e := c.Query("end_date"); e != "" {
		if t, err := time.Parse("2006-01-02", e); err == nil {
			endDate = &t
		}
	}

	inspections, err := h.usecase.GetFleetInspections(vehicleID, tripID, followUpAction, followUpStatus, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, inspections)
}

func (h *TransportHandler) GetFleetInspectionByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid inspection ID"})
		return
	}
	inspection, err := h.usecase.GetFleetInspectionByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Inspection not found"})
		return
	}
	c.JSON(http.StatusOK, inspection)
}

func (h *TransportHandler) UpdateInspectionFollowUp(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid inspection ID"})
		return
	}

	var req struct {
		WorkshopName   string                           `json:"workshop_name"`
		PartsReplaced  string                           `json:"parts_replaced"`
		ActualCost     float64                          `json:"actual_cost"`
		FollowUpStatus domain.MaintenanceFollowUpStatus `json:"follow_up_status" binding:"required"`
		RepairNotes    string                           `json:"repair_notes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updated, err := h.usecase.UpdateInspectionFollowUp(id, req.WorkshopName, req.PartsReplaced, req.ActualCost, req.FollowUpStatus, req.RepairNotes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, updated)
}

func (h *TransportHandler) DeleteFleetInspection(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid inspection ID"})
		return
	}
	if err := h.usecase.DeleteFleetInspection(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Inspection deleted successfully"})
}

// ──────────────── DRIVER ATTENDANCE HANDLERS ────────────────

type RecordDriverAttendanceRequest struct {
	TripID         string  `json:"trip_id" form:"trip_id" binding:"required"`
	DriverID       string  `json:"driver_id" form:"driver_id" binding:"required"`
	AttendanceDate string  `json:"attendance_date" form:"attendance_date" binding:"required"`
	Status         string  `json:"status" form:"status"`
	CurrentCity    string  `json:"current_city" form:"current_city"`
	CheckinLat     float64 `json:"checkin_lat" form:"checkin_lat"`
	CheckinLong    float64 `json:"checkin_long" form:"checkin_long"`
	PhotoURL       string  `json:"photo_url" form:"photo_url"`
	DailyAllowance float64 `json:"daily_allowance" form:"daily_allowance"`
	Notes          string  `json:"notes" form:"notes"`
}

func (h *TransportHandler) RecordDriverAttendance(c *gin.Context) {
	var req RecordDriverAttendanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tripID, err := uuid.Parse(req.TripID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid trip_id"})
		return
	}
	drvID, err := uuid.Parse(req.DriverID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid driver_id"})
		return
	}

	attDate, err := time.Parse("2006-01-02", req.AttendanceDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid attendance_date (YYYY-MM-DD)"})
		return
	}

	att := domain.TransportDriverAttendance{
		TripID:         tripID,
		DriverID:       drvID,
		AttendanceDate: attDate,
		Status:         req.Status,
		CurrentCity:    req.CurrentCity,
		CheckinLat:     req.CheckinLat,
		CheckinLong:    req.CheckinLong,
		PhotoURL:       req.PhotoURL,
		DailyAllowance: req.DailyAllowance,
		Notes:          req.Notes,
	}

	if err := h.usecase.RecordDriverAttendance(&att); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, att)
}

func (h *TransportHandler) GetDriverAttendances(c *gin.Context) {
	var tripID, driverID *uuid.UUID
	if val := c.Query("trip_id"); val != "" && val != "ALL" {
		if parsed, err := uuid.Parse(val); err == nil {
			tripID = &parsed
		}
	}
	if val := c.Query("driver_id"); val != "" && val != "ALL" {
		if parsed, err := uuid.Parse(val); err == nil {
			driverID = &parsed
		}
	}

	var startDate, endDate *time.Time
	if s := c.Query("start_date"); s != "" {
		if t, err := time.Parse("2006-01-02", s); err == nil {
			startDate = &t
		}
	}
	if e := c.Query("end_date"); e != "" {
		if t, err := time.Parse("2006-01-02", e); err == nil {
			endDate = &t
		}
	}

	attendances, err := h.usecase.GetDriverAttendances(tripID, driverID, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, attendances)
}

// ──────────────── HELPER FOR UPLOADS ────────────────

func (h *TransportHandler) handleFileUpload(c *gin.Context, fieldName string, targetDir string) (string, error) {
	contentType := c.GetHeader("Content-Type")
	if !strings.HasPrefix(contentType, "multipart/form-data") {
		return "", nil
	}

	file, err := c.FormFile(fieldName)
	if err != nil {
		if err == http.ErrMissingFile {
			return "", nil
		}
		return "", err
	}

	// Max 10MB
	if file.Size > 10*1024*1024 {
		return "", fmt.Errorf("file size exceeds 10MB")
	}

	filename := fmt.Sprintf("%s_%d%s", uuid.New().String(), time.Now().UnixNano(), filepath.Ext(file.Filename))
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return "", err
	}

	savePath := filepath.Join(targetDir, filename)
	if err := utils.SaveUploadedFile(file, savePath); err != nil {
		return "", err
	}

	return "/" + targetDir + "/" + filename, nil
}

// Helper int query param parser
func parseQueryInt(val string, defaultVal int) int {
	if val == "" {
		return defaultVal
	}
	if i, err := strconv.Atoi(val); err == nil {
		return i
	}
	return defaultVal
}
