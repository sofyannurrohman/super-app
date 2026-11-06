package http

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"example.com/hris/backend/internal/usecase"
	"example.com/hris/backend/pkg/utils"
)

type AttendanceHandler struct {
	attendanceUseCase usecase.AttendanceUseCase
}

func NewAttendanceHandler(attendanceUseCase usecase.AttendanceUseCase) *AttendanceHandler {
	return &AttendanceHandler{attendanceUseCase}
}

func (h *AttendanceHandler) SetupPublicRoutes(router *gin.RouterGroup) {
	attendance := router.Group("/attendance")
	{
		attendance.POST("/tablet", h.TabletAttendance)
		attendance.GET("/tablet/check-pin", h.CheckPinTablet)
	}
}

func (h *AttendanceHandler) SetupRoutes(router *gin.RouterGroup) {
	attendance := router.Group("/attendance")
	{
		attendance.POST("/checkin", h.CheckIn)
		attendance.POST("/checkout", h.CheckOut)
		attendance.GET("/history", h.GetHistory)
		attendance.GET("/stats", h.GetStats)
		attendance.PUT("/my-notes/:id", h.UpdateMyNotes)

		readOnlyAdmin := attendance.Group("")
		readOnlyAdmin.Use(RoleMiddleware(string(domain.RoleSuperAdmin), string(domain.RoleHRAdmin), string(domain.RoleManager), string(domain.RoleAdminNota), string(domain.RoleSupervisorSales), string(domain.RoleSupervisorSalesMT), string(domain.RoleAdminKasir), string(domain.RoleAdminOperasional)))
		{
			readOnlyAdmin.GET("/all", h.GetAllHistory)
			readOnlyAdmin.GET("/export-xlsx", h.ExportXLSX)
		}

		admin := attendance.Group("")
		admin.Use(RoleMiddleware(string(domain.RoleSuperAdmin), string(domain.RoleHRAdmin), string(domain.RoleManager), string(domain.RoleAdminKasir)))
		{
			admin.POST("/manual", h.CreateManual)
			admin.PUT("/:id", h.UpdateAttendance)
			admin.DELETE("/:id", h.DeleteAttendance)
		}
	}
}

func (h *AttendanceHandler) CheckIn(c *gin.Context) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized")
		return
	}
	userID := userIDStr.(uuid.UUID)

	var req usecase.CheckInRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	att, err := h.attendanceUseCase.CheckIn(userID, req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Check-in successful", att)
}

func (h *AttendanceHandler) CheckOut(c *gin.Context) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized")
		return
	}
	userID := userIDStr.(uuid.UUID)

	var req usecase.CheckOutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	att, err := h.attendanceUseCase.CheckOut(userID, req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Check-out successful", att)
}

func (h *AttendanceHandler) TabletAttendance(c *gin.Context) {
	var req usecase.TabletAttendanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	att, err := h.attendanceUseCase.TabletAttendance(req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Absensi berhasil dicatat", att)
}

func (h *AttendanceHandler) CheckPinTablet(c *gin.Context) {
	pin := c.Query("pin")
	branchID := c.Query("branch_id")

	if pin == "" || branchID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "PIN dan ID Cabang wajib diisi")
		return
	}

	res, err := h.attendanceUseCase.CheckPinTablet(pin, branchID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Pekerja ditemukan", map[string]interface{}{
		"id":              res.Employee.ID,
		"first_name":      res.Employee.FirstName,
		"last_name":       res.Employee.LastName,
		"has_checked_in":  res.HasCheckedIn,
		"has_clocked_out": res.HasClockedOut,
	})
}

func (h *AttendanceHandler) GetHistory(c *gin.Context) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized")
		return
	}
	userID := userIDStr.(uuid.UUID)

	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	history, err := h.attendanceUseCase.GetHistory(userID, startDate, endDate)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get history")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Attendance history fetched", history)
}

func (h *AttendanceHandler) GetStats(c *gin.Context) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized")
		return
	}
	userID := userIDStr.(uuid.UUID)

	stats, err := h.attendanceUseCase.GetStats(userID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get stats: "+err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Attendance stats fetched", stats)
}

type UpdateMyNotesRequest struct {
	Notes string `json:"notes" binding:"required"`
}

func (h *AttendanceHandler) UpdateMyNotes(c *gin.Context) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized")
		return
	}
	userID := userIDStr.(uuid.UUID)

	idParam := c.Param("id")
	attendanceID, err := uuid.Parse(idParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid Attendance ID")
		return
	}

	var req UpdateMyNotesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Catatan alasan wajib diisi")
		return
	}

	if err := h.attendanceUseCase.UpdateMyNotes(userID, attendanceID, req.Notes); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Catatan kehadiran berhasil diperbarui", nil)
}

func (h *AttendanceHandler) GetAllHistory(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")

	page, _ := strconv.Atoi(pageStr)
	limit, _ := strconv.Atoi(limitStr)

	branchIDs := utils.ParseUUIDs(c.Query("branch_id"))

	month := c.Query("month") // format: YYYY-MM

	employeeIDStr := c.Query("employee_id")
	var employeeID *uuid.UUID
	if employeeIDStr != "" {
		id, err := uuid.Parse(employeeIDStr)
		if err == nil {
			employeeID = &id
		}
	}

	employmentType := c.Query("employment_type")

	history, err := h.attendanceUseCase.GetAllHistory(page, limit, branchIDs, employeeID, month, employmentType)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get all attendance history")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "All attendance history fetched", history)
}

func (h *AttendanceHandler) ExportXLSX(c *gin.Context) {
	branchIDs := utils.ParseUUIDs(c.Query("branch_id"))

	month := c.Query("month")
	if month == "" {
		month = time.Now().Format("2006-01")
	}

	employmentType := c.Query("employment_type")

	xlsxBytes, err := h.attendanceUseCase.ExportAttendanceXLSX(branchIDs, month, employmentType)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to export attendance: "+err.Error())
		return
	}

	filename := fmt.Sprintf("kehadiran_%s.xlsx", month)
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", xlsxBytes)
}

func (h *AttendanceHandler) CreateManual(c *gin.Context) {
	var req usecase.ManualAttendanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	att, err := h.attendanceUseCase.CreateManual(req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create manual attendance: "+err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Manual attendance created successfully", att)
}

func (h *AttendanceHandler) UpdateAttendance(c *gin.Context) {
	idParam := c.Param("id")
	attendanceID, err := uuid.Parse(idParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid Attendance ID")
		return
	}

	var req usecase.UpdateAttendanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.attendanceUseCase.UpdateAttendance(attendanceID, req); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update attendance: "+err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Attendance updated successfully", nil)
}

func (h *AttendanceHandler) DeleteAttendance(c *gin.Context) {
	idParam := c.Param("id")
	attendanceID, err := uuid.Parse(idParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid Attendance ID")
		return
	}

	if err := h.attendanceUseCase.DeleteAttendance(attendanceID); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete attendance: "+err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Attendance deleted successfully", nil)
}
