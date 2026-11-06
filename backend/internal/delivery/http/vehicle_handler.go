package http

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"example.com/hris/backend/internal/usecase"
	"example.com/hris/backend/pkg/utils"
)

type VehicleHandler struct {
	usecase usecase.VehicleUsecase
}

func NewVehicleHandler(u usecase.VehicleUsecase) *VehicleHandler {
	return &VehicleHandler{u}
}

func (h *VehicleHandler) RegisterRoutes(r *gin.RouterGroup) {
	vehicles := r.Group("/vehicles")
	{
		vehicles.GET("", h.GetAllVehicles)
		vehicles.POST("", h.CreateVehicle)
		vehicles.GET("/:id", h.GetVehicleDetail)
		vehicles.PUT("/:id", h.UpdateVehicle)
		vehicles.DELETE("/:id", h.DeleteVehicle)
		vehicles.GET("/:id/logs", h.GetVehicleLogs)
		vehicles.GET("/:id/active-task", h.GetActiveTask)
		vehicles.POST("/:id/maintenance", h.RecordMaintenance)
		vehicles.GET("/traxcenter/:imei", h.GetTraxCenterLive)
	}
}

func (h *VehicleHandler) GetAllVehicles(c *gin.Context) {
	companyIDStr := c.Query("company_id")
	var companyID *uuid.UUID

	if companyIDStr != "" && companyIDStr != "all" && companyIDStr != "ALL" && companyIDStr != "ALL_COMPANIES" {
		parsed, err := uuid.Parse(companyIDStr)
		if err == nil && parsed != uuid.Nil {
			companyID = &parsed
		}
	}

	roleVal, existsRole := c.Get("role")
	roleStr := ""
	if existsRole {
		roleStr = roleVal.(string)
	}

	isCentralOrAdminRole := roleStr == string(domain.RoleSuperAdmin) ||
		roleStr == string(domain.RoleCashierPusat) ||
		roleStr == string(domain.RoleAdminKasir) ||
		roleStr == string(domain.RoleAdminPabrik) ||
		roleStr == string(domain.RoleAuditor) ||
		roleStr == string(domain.RoleAdminNota) ||
		roleStr == string(domain.RoleSupervisorSales) ||
		roleStr == string(domain.RoleAdminOperasional)

	bypassCompany := c.Query("bypass_company") == "true" || c.Query("all_companies") == "true" || c.Query("all") == "true" || isCentralOrAdminRole

	// Fallback to token if query is empty and not explicitly bypassed
	if companyID == nil && !bypassCompany {
		if val, exists := c.Get("companyID"); exists {
			cid := val.(uuid.UUID)
			if cid != uuid.Nil {
				companyID = &cid
			}
		}
	}

	branchIDStr := c.Query("branch_id")
	var branchID *uuid.UUID

	// Check if the user is a warehouse admin (admin_gudang or supervisor_gudang) or central role to bypass branch filter
	isWarehouseAdmin := roleStr == string(domain.RoleAdminGudang) || roleStr == string(domain.RoleSupervisorGudang)

	// Bypass branch filter if explicit query parameters are set or if role is admin_gudang / central role
	bypassBranch := c.Query("bypass_branch") == "true" || c.Query("all_branches") == "true" || c.Query("all") == "true" || isWarehouseAdmin || isCentralOrAdminRole

	if !bypassBranch && branchIDStr != "" && branchIDStr != "all" && branchIDStr != "ALL" && branchIDStr != "ALL_BRANCHES" {
		parsed, err := uuid.Parse(branchIDStr)
		if err == nil && parsed != uuid.Nil {
			branchID = &parsed
		}
	}

	vehicles, err := h.usecase.GetVehicles(companyID, branchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, vehicles)
}

func (h *VehicleHandler) CreateVehicle(c *gin.Context) {
	var req VehicleRequest
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

	if compJSON := c.PostForm("components_json"); compJSON != "" {
		importJson := []VehicleComponentRequest{}
		if err := json.Unmarshal([]byte(compJSON), &importJson); err == nil {
			req.Components = importJson
		}
	}

	var vehicle domain.Vehicle
	vehicle.Name = req.Name
	vehicle.LicensePlate = req.LicensePlate
	vehicle.Model = req.Model
	vehicle.Type = req.Type
	vehicle.Status = req.Status
	vehicle.Year = req.Year
	vehicle.Mileage = req.Mileage
	vehicle.Capacity = req.Capacity
	vehicle.BedLength = req.BedLength
	vehicle.BedWidth = req.BedWidth
	vehicle.BedHeight = req.BedHeight
	vehicle.BedVolume = req.BedVolume
	vehicle.GpsIMEI = strings.TrimSpace(req.GpsIMEI)

	if req.BranchID != "" {
		if bid, err := uuid.Parse(req.BranchID); err == nil {
			vehicle.BranchID = &bid
		}
	}

	if req.TaxExpiryDate != "" {
		t, err := time.Parse("2006-01-02", req.TaxExpiryDate)
		if err == nil {
			vehicle.TaxExpiryDate = &t
		}
	}

	// Handle Image Uploads
	photoFront, err := h.handleFileUpload(c, "photo_front")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload photo_front: " + err.Error()})
		return
	}
	if photoFront != "" {
		vehicle.PhotoFront = photoFront
		vehicle.ImageURL = photoFront // for backward compatibility
	}

	photoLeft, _ := h.handleFileUpload(c, "photo_left")
	if photoLeft != "" {
		vehicle.PhotoLeft = photoLeft
	}

	photoRight, _ := h.handleFileUpload(c, "photo_right")
	if photoRight != "" {
		vehicle.PhotoRight = photoRight
	}

	photoBack, _ := h.handleFileUpload(c, "photo_back")
	if photoBack != "" {
		vehicle.PhotoBack = photoBack
	}

	companyID, err := h.getCompanyID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	vehicle.CompanyID = companyID

	var components []domain.VehicleComponent
	for _, compReq := range req.Components {
		components = append(components, domain.VehicleComponent{
			Name:       compReq.Name,
			MaxMileage: compReq.MaxMileage,
		})
	}

	if err := h.usecase.CreateVehicle(&vehicle, components); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, vehicle)
}

func (h *VehicleHandler) GetVehicleDetail(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid vehicle ID format"})
		return
	}
	vehicle, err := h.usecase.GetVehicleDetail(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Vehicle not found"})
		return
	}
	c.JSON(http.StatusOK, vehicle)
}

func (h *VehicleHandler) UpdateVehicle(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid vehicle ID format"})
		return
	}

	var req VehicleRequest
	var bindErr error
	if strings.Contains(c.GetHeader("Content-Type"), "multipart/form-data") {
		bindErr = c.ShouldBindWith(&req, binding.FormMultipart)
	} else {
		bindErr = c.ShouldBindJSON(&req)
	}

	if bindErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": bindErr.Error()})
		return
	}

	if compJSON := c.PostForm("components_json"); compJSON != "" {
		importJson := []VehicleComponentRequest{}
		if err := json.Unmarshal([]byte(compJSON), &importJson); err == nil {
			req.Components = importJson
		}
	}

	// Retrieve existing vehicle to preserve fields not sent
	existing, err := h.usecase.GetVehicleDetail(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Vehicle not found"})
		return
	}

	existing.Name = req.Name
	existing.LicensePlate = req.LicensePlate
	existing.Model = req.Model
	existing.Type = req.Type
	existing.Status = req.Status
	existing.Year = req.Year
	existing.Mileage = req.Mileage
	existing.Capacity = req.Capacity
	existing.BedLength = req.BedLength
	existing.BedWidth = req.BedWidth
	existing.BedHeight = req.BedHeight
	existing.BedVolume = req.BedVolume
	existing.GpsIMEI = strings.TrimSpace(req.GpsIMEI)

	if req.BranchID != "" {
		if bid, err := uuid.Parse(req.BranchID); err == nil {
			existing.BranchID = &bid
		}
	} else {
		existing.BranchID = nil
	}

	if req.TaxExpiryDate != "" {
		t, err := time.Parse("2006-01-02", req.TaxExpiryDate)
		if err == nil {
			existing.TaxExpiryDate = &t
		}
	} else {
		existing.TaxExpiryDate = nil
	}

	// Handle Image Uploads
	photoFront, err := h.handleFileUpload(c, "photo_front")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload photo_front: " + err.Error()})
		return
	}
	if photoFront != "" {
		existing.PhotoFront = photoFront
		existing.ImageURL = photoFront // for backward compatibility
	}

	photoLeft, _ := h.handleFileUpload(c, "photo_left")
	if photoLeft != "" {
		existing.PhotoLeft = photoLeft
	}

	photoRight, _ := h.handleFileUpload(c, "photo_right")
	if photoRight != "" {
		existing.PhotoRight = photoRight
	}

	photoBack, _ := h.handleFileUpload(c, "photo_back")
	if photoBack != "" {
		existing.PhotoBack = photoBack
	}

	companyID, err := h.getCompanyID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	existing.CompanyID = companyID

	var components []domain.VehicleComponent
	for _, compReq := range req.Components {
		comp := domain.VehicleComponent{
			Name:       compReq.Name,
			MaxMileage: compReq.MaxMileage,
		}
		if compReq.ID != "" {
			if id, err := uuid.Parse(compReq.ID); err == nil {
				comp.ID = id
			}
		}
		components = append(components, comp)
	}

	if err := h.usecase.UpdateVehicle(existing, components); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, existing)
}

func (h *VehicleHandler) DeleteVehicle(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid vehicle ID format"})
		return
	}
	if err := h.usecase.DeleteVehicle(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Vehicle deleted"})
}

func (h *VehicleHandler) GetVehicleLogs(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid vehicle ID format"})
		return
	}
	logs, err := h.usecase.GetVehicleLogs(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, logs)
}

func (h *VehicleHandler) GetActiveTask(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid vehicle ID format"})
		return
	}
	task, err := h.usecase.GetActiveAssignment(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if task == nil {
		c.JSON(http.StatusOK, gin.H{"active": false, "task": nil})
		return
	}
	c.JSON(http.StatusOK, gin.H{"active": true, "task": task})
}

func (h *VehicleHandler) handleFileUpload(c *gin.Context, fieldName string) (string, error) {
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

	// Limit 5MB
	if file.Size > 5*1024*1024 {
		return "", fmt.Errorf("file size exceeds 5MB limit")
	}

	filename := uuid.New().String() + filepath.Ext(file.Filename)
	uploadDir := "uploads/vehicles"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return "", err
	}

	savePath := filepath.Join(uploadDir, filename)
	if err := utils.SaveUploadedFile(file, savePath); err != nil {
		return "", err
	}

	return "/uploads/vehicles/" + filename, nil
}

func (h *VehicleHandler) RecordMaintenance(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid vehicle ID format"})
		return
	}

	var req struct {
		ComponentID     string `json:"component_id" binding:"required"`
		MaintenanceDate string `json:"maintenance_date" binding:"required"` // Format: YYYY-MM-DD
		VehicleMileage  int    `json:"vehicle_mileage" binding:"required"`
		Notes           string `json:"notes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	compID, err := uuid.Parse(req.ComponentID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid component ID format"})
		return
	}

	// Simple string parsing or use time.Parse
	parsedDate, err := time.Parse("2006-01-02", req.MaintenanceDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format, use YYYY-MM-DD"})
		return
	}

	maintenance := domain.VehicleMaintenance{
		VehicleID:          id,
		VehicleComponentID: compID,
		MaintenanceDate:    parsedDate,
		VehicleMileage:     req.VehicleMileage,
		Notes:              req.Notes,
	}

	if err := h.usecase.RecordMaintenance(&maintenance); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, maintenance)
}

type VehicleComponentRequest struct {
	ID         string `form:"id" json:"id"`
	Name       string `form:"name" json:"name" binding:"required"`
	MaxMileage int    `form:"max_mileage" json:"max_mileage" binding:"required,gt=0"`
}

type VehicleRequest struct {
	Name         string                    `form:"name" json:"name" binding:"required"`
	LicensePlate string                    `form:"license_plate" json:"license_plate" binding:"required"`
	Model        string                    `form:"model" json:"model"`
	Type         string                    `form:"type" json:"type"`
	Status       string                    `form:"status" json:"status"`
	Year         int                       `form:"year" json:"year"`
	Mileage      int                       `form:"mileage" json:"mileage"`
	Capacity     float64                   `form:"capacity" json:"capacity"`
	BedLength    float64                   `form:"bed_length" json:"bed_length"`
	BedWidth     float64                   `form:"bed_width" json:"bed_width"`
	BedHeight    float64                   `form:"bed_height" json:"bed_height"`
	BedVolume    float64                   `form:"bed_volume" json:"bed_volume"`
	BranchID     string                    `form:"branch_id" json:"branch_id"`
	TaxExpiryDate string                   `form:"tax_expiry_date" json:"tax_expiry_date"`
	GpsIMEI      string                    `form:"gps_imei" json:"gps_imei"`
	Components   []VehicleComponentRequest `form:"components" json:"components"`
}

func (h *VehicleHandler) getCompanyID(c *gin.Context) (uuid.UUID, error) {
	val, exists := c.Get("companyID")
	if !exists || val == nil {
		return uuid.Nil, fmt.Errorf("company ID not found in context")
	}

	if cid, ok := val.(uuid.UUID); ok {
		return cid, nil
	}

	if cidStr, ok := val.(string); ok {
		parsed, err := uuid.Parse(cidStr)
		if err == nil {
			return parsed, nil
		}
	}

	return uuid.Nil, fmt.Errorf("invalid company ID format in context")
}

func (h *VehicleHandler) GetTraxCenterLive(c *gin.Context) {
	imei := strings.TrimSpace(c.Param("imei"))
	if imei == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "IMEI is required"})
		return
	}

	url := fmt.Sprintf("https://trax.center/api/?%s", imei)
	client := http.Client{
		Timeout: 7 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Failed to connect to TraxCenter API: " + err.Error()})
		return
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read TraxCenter response: " + err.Error()})
		return
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &parsed); err != nil {
		c.Data(resp.StatusCode, "application/json", bodyBytes)
		return
	}

	c.JSON(http.StatusOK, parsed)
}

