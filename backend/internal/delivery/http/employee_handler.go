package http

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"example.com/hris/backend/internal/usecase"
	"example.com/hris/backend/pkg/redisclient"
	"example.com/hris/backend/pkg/utils"
	"gorm.io/gorm"
)

type EmployeeHandler struct {
	employeeUsecase     usecase.EmployeeUsecase
	autoInactiveUsecase usecase.AutoInactiveUsecase
	db                  *gorm.DB
}

func NewEmployeeHandler(employeeUsecase usecase.EmployeeUsecase, autoInactiveUsecase usecase.AutoInactiveUsecase, db *gorm.DB) *EmployeeHandler {
	return &EmployeeHandler{
		employeeUsecase:     employeeUsecase,
		autoInactiveUsecase: autoInactiveUsecase,
		db:                  db,
	}
}

func (h *EmployeeHandler) SetupRoutes(router *gin.RouterGroup) {
	// Group under /employees
	employees := router.Group("/employees")

	// Authenticated but common endpoints
	employees.GET("/profile", h.GetProfile)
	employees.PUT("/profile", h.UpdateProfile)
	employees.POST("/avatar", h.UploadAvatar)
	employees.POST("/upload-avatar", h.UploadAvatarFile)
	employees.GET("/directory", h.GetDirectory)

	// Read-Only Admin Endpoints
	readOnlyAdmin := employees.Group("")
	readOnlyAdmin.Use(RoleMiddleware(
		string(domain.RoleSuperAdmin),
		string(domain.RoleHRAdmin),
		string(domain.RoleManager),
		string(domain.RoleAdminNota),
		string(domain.RoleSupervisorSales),
		string(domain.RoleCashier),
		string(domain.RoleCashierPusat),
		string(domain.RoleAdminKasir),
		string(domain.RoleAdminDesigner),
		string(domain.RoleValidatorDesain),
		string(domain.RoleValidatorToko),
		string(domain.RoleAdminPurchasing),
		string(domain.RoleOperator),
		string(domain.RoleAdminPabrik),
		string(domain.RoleSupervisorGudang),
		string(domain.RoleAdminGudang),
		string(domain.RoleAdminMarketplace),
		string(domain.RoleAdminOperasional),
	))
	{
		readOnlyAdmin.GET("", h.GetEmployees)
		readOnlyAdmin.GET("/export-csv", h.ExportCSV)
	}

	// Open Admin Endpoints
	admin := employees.Group("")
	admin.Use(RoleMiddleware(string(domain.RoleSuperAdmin), string(domain.RoleHRAdmin)))
	{
		admin.POST("/import-daily", h.ImportDailyWorkers)
		admin.PUT("/:id", h.UpdateEmployee)
		admin.DELETE("/:id", h.DeleteEmployee)
		admin.POST("/auto-inactive/run", h.RunAutoInactive)
		admin.GET("/auto-inactive/candidates", h.GetAutoInactiveCandidates)
		admin.POST("/:id/reactivate", h.ReactivateEmployee)
	}

	// Create Employee Endpoint
	createEmployee := employees.Group("")
	createEmployee.Use(RoleMiddleware(string(domain.RoleSuperAdmin), string(domain.RoleHRAdmin), string(domain.RoleCashier), string(domain.RoleAdminKasir)))
	{
		createEmployee.POST("", h.CreateEmployee)
	}
}

func (h *EmployeeHandler) GetEmployees(c *gin.Context) {
	limitStr := c.Query("limit")
	limit := 10 // default
	if limitStr != "" {
		if limitStr == "all" {
			limit = 0 // Fetch all
		} else if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	pageStr := c.Query("page")
	page := 1
	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil {
			page = p
		}
	}

	search := c.Query("search")
	branchIDs := utils.ParseUUIDs(c.Query("branch_id"))

	var jobPositionID *uuid.UUID
	jobPositionIDStr := c.Query("job_position_id")
	if jobPositionIDStr != "" && jobPositionIDStr != "all" {
		if id, err := uuid.Parse(jobPositionIDStr); err == nil {
			jobPositionID = &id
		}
	}

	isSales := c.Query("is_sales") == "true"
	salesType := c.Query("sales_type")
	employmentType := c.Query("employment_type")
	status := c.Query("status")
	branchMatch := c.Query("branch_match")
	res, total, err := h.employeeUsecase.GetEmployeesPaginated(page, limit, branchIDs, search, isSales, salesType, employmentType, status, jobPositionID, branchMatch)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get employees: "+err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Employees fetched successfully",
		"data":    res,
		"meta": gin.H{
			"total": total,
			"page":  page,
			"limit": limit,
		},
	})
}

func (h *EmployeeHandler) CreateEmployee(c *gin.Context) {
	var req usecase.CreateEmployeeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// Role Cashier restriction: can only create "Harian" employees
	userRole, existsRole := c.Get("role")
	if existsRole && (userRole.(string) == string(domain.RoleCashier) || userRole.(string) == string(domain.RoleAdminKasir)) {
		if req.EmploymentStatus != "Harian" {
			utils.ErrorResponse(c, http.StatusForbidden, "Kasir hanya diizinkan menambahkan pekerja dengan status Harian")
			return
		}
	}

	// Automatically use company ID from JWT claims
	if companyID, exists := c.Get("companyID"); exists {
		cid := companyID.(uuid.UUID)
		if cid != uuid.Nil {
			req.CompanyID = &cid
		}
	}

	if err := h.employeeUsecase.CreateEmployee(&req); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create employee: "+err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Employee created successfully", nil)
}

func (h *EmployeeHandler) UpdateEmployee(c *gin.Context) {
	idParam := c.Param("id")
	employeeID, err := uuid.Parse(idParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid Employee ID")
		return
	}

	var req usecase.UpdateEmployeeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.employeeUsecase.UpdateEmployee(employeeID, &req); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update employee: "+err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Employee updated successfully", nil)
}

func (h *EmployeeHandler) DeleteEmployee(c *gin.Context) {
	idParam := c.Param("id")
	employeeID, err := uuid.Parse(idParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid Employee ID")
		return
	}

	if err := h.employeeUsecase.DeleteEmployee(employeeID); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete employee: "+err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Employee deleted successfully", nil)
}

func (h *EmployeeHandler) GetProfile(c *gin.Context) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized")
		return
	}
	userID := userIDStr.(uuid.UUID)

	employee, err := h.employeeUsecase.GetEmployeeByUserID(userID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get profile: "+err.Error())
		return
	}
	if employee == nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Employee record not found")
		return
	}

	// Fetch active warning if exists
	var warning domain.EmployeeWarning
	if err := h.db.Where("employee_id = ? AND is_active = ?", employee.ID, true).
		Order("created_at desc").
		First(&warning).Error; err == nil {
		employee.ActiveWarning = &warning
	}

	utils.SuccessResponse(c, http.StatusOK, "Profile fetched successfully", employee)
}

func (h *EmployeeHandler) UpdateProfile(c *gin.Context) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized")
		return
	}
	userID := userIDStr.(uuid.UUID)

	var req usecase.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.employeeUsecase.UpdateProfile(userID, &req); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update profile: "+err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Profile updated successfully", nil)
}

func (h *EmployeeHandler) ExportCSV(c *gin.Context) {
	branchIDs := utils.ParseUUIDs(c.Query("branch_id"))

	var jobPositionID *uuid.UUID
	jobPositionIDStr := c.Query("job_position_id")
	if jobPositionIDStr != "" && jobPositionIDStr != "all" {
		if id, err := uuid.Parse(jobPositionIDStr); err == nil {
			jobPositionID = &id
		}
	}

	status := c.Query("status")
	search := c.Query("search")
	employmentType := c.Query("employment_type")
	branchMatch := c.Query("branch_match")

	// Fetch all records (limit=0 = no limit)
	employees, _, err := h.employeeUsecase.GetEmployeesPaginated(1, 0, branchIDs, search, false, "", employmentType, status, jobPositionID, branchMatch)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to export employees")
		return
	}

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Header
	_ = writer.Write([]string{"No", "ID Karyawan", "Nama", "Email", "No. Telepon", "Departemen", "Jabatan", "Cabang", "Status", "Tanggal Bergabung", "Gaji Pokok", "Bank", "No. Rekening"})

	for i, emp := range employees {
		deptName := "-"
		if emp.Department != nil {
			deptName = emp.Department.Name
		}
		posName := "-"
		if emp.JobPosition != nil {
			posName = emp.JobPosition.Title
		}
		branchName := "-"
		if emp.Branch != nil {
			branchName = emp.Branch.Name
		}
		email := "-"
		if emp.User != nil {
			email = emp.User.Email
		}

		salaryStr := "0"
		if emp.Salary != nil {
			salaryStr = fmt.Sprintf("%.0f", *emp.Salary)
		}

		_ = writer.Write([]string{
			strconv.Itoa(i + 1),
			emp.EmployeeIDNumber,
			emp.FirstName,
			email,
			emp.PhoneNumber,
			deptName,
			posName,
			branchName,
			emp.EmploymentStatus,
			emp.JoinDate.Format("02/01/2006"),
			salaryStr,
			emp.BankName,
			emp.BankAccountNumber,
		})
	}

	writer.Flush()

	filename := "data_karyawan.csv"
	if branchName := c.Query("branch_name"); branchName != "" {
		filename = fmt.Sprintf("data_karyawan_%s.csv", branchName)
	}

	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Data(http.StatusOK, "text/csv; charset=utf-8", buf.Bytes())
}

func (h *EmployeeHandler) GetDirectory(c *gin.Context) {
	query := c.Query("q")
	branchIDs := utils.ParseUUIDs(c.Query("branch_id"))

	res, err := h.employeeUsecase.GetDirectory(query, branchIDs)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get directory: "+err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Directory fetched successfully", res)
}

func (h *EmployeeHandler) ImportDailyWorkers(c *gin.Context) {
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "File is required")
		return
	}
	defer file.Close()

	if err := h.employeeUsecase.ImportDailyWorkers(file); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Daily workers imported successfully", nil)
}

func (h *EmployeeHandler) RunAutoInactive(c *gin.Context) {
	dryRunStr := c.DefaultQuery("dry_run", "false")
	dryRun := dryRunStr == "true" || dryRunStr == "1"

	summary, err := h.autoInactiveUsecase.EvaluateAndDeactivateInactiveEmployees(c.Request.Context(), dryRun)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal mengevaluasi auto-inactive karyawan: "+err.Error())
		return
	}

	msg := fmt.Sprintf("Evaluasi selesai. %d karyawan dievaluasi, %d karyawan dinonaktifkan.", summary.EvaluatedCount, summary.DeactivatedCount)
	if dryRun {
		msg = fmt.Sprintf("[DRY RUN] %d karyawan dievaluasi, %d karyawan terdeteksi mangkir 10 hari.", summary.EvaluatedCount, summary.DeactivatedCount)
	}

	utils.SuccessResponse(c, http.StatusOK, msg, summary)
}

func (h *EmployeeHandler) GetAutoInactiveCandidates(c *gin.Context) {
	minDaysStr := c.DefaultQuery("min_days", "7")
	minDays, err := strconv.Atoi(minDaysStr)
	if err != nil || minDays <= 0 {
		minDays = 7
	}

	candidates, err := h.autoInactiveUsecase.GetAutoInactiveCandidates(c.Request.Context(), minDays)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal memuat kandidat auto-inactive: "+err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Kandidat auto-inactive berhasil dimuat", candidates)
}

type ReactivateEmployeeRequest struct {
	Reason string `json:"reason" binding:"required"`
}

func (h *EmployeeHandler) ReactivateEmployee(c *gin.Context) {
	idStr := c.Param("id")
	empID, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "ID karyawan tidak valid")
		return
	}

	var req ReactivateEmployeeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Alasan reaktivasi harus diisi")
		return
	}

	adminUserID := uuid.Nil
	if val, exists := c.Get("userID"); exists {
		if uidStr, ok := val.(string); ok {
			adminUserID, _ = uuid.Parse(uidStr)
		} else if uid, ok := val.(uuid.UUID); ok {
			adminUserID = uid
		}
	}

	adminName := "ADMIN"
	if val, exists := c.Get("name"); exists && val != nil && val.(string) != "" {
		adminName = val.(string)
	} else if val, exists := c.Get("username"); exists && val != nil && val.(string) != "" {
		adminName = val.(string)
	}

	if err := h.autoInactiveUsecase.ReactivateEmployee(c.Request.Context(), empID, adminUserID, adminName, req.Reason); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal mengaktifkan kembali karyawan: "+err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Karyawan berhasil diaktifkan kembali", nil)
}

func (h *EmployeeHandler) UploadAvatar(c *gin.Context) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized")
		return
	}
	userID := userIDStr.(uuid.UUID)

	employee, err := h.employeeUsecase.GetEmployeeByUserID(userID)
	if err != nil || employee == nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Data karyawan tidak ditemukan")
		return
	}

	file, _, err := c.Request.FormFile("avatar")
	if err != nil {
		file, _, err = c.Request.FormFile("photo")
		if err != nil {
			file, _, err = c.Request.FormFile("file")
			if err != nil {
				utils.ErrorResponse(c, http.StatusBadRequest, "File foto tidak ditemukan dalam request")
				return
			}
		}
	}
	defer file.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Gagal membaca file foto")
		return
	}

	compressedData, err := utils.CompressImage(fileBytes, 800, 75)
	if err != nil {
		compressedData = fileBytes
	}

	uploadDir := "uploads/faces"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal membuat direktori upload")
		return
	}

	filename := fmt.Sprintf("%s_%d.jpg", employee.ID.String(), time.Now().Unix())
	filePath := filepath.Join(uploadDir, filename)

	if err := os.WriteFile(filePath, compressedData, 0644); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal menyimpan foto")
		return
	}

	avatarURL := "/" + filePath
	if err := h.db.Model(&domain.Employee{}).Where("id = ?", employee.ID).Update("face_reference_url", avatarURL).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal memperbarui foto profil karyawan: "+err.Error())
		return
	}

	if redisclient.Client != nil {
		ctx := context.Background()
		_ = redisclient.Client.Del(ctx, fmt.Sprintf("employee:id:%s", employee.ID.String())).Err()
		_ = redisclient.Client.Del(ctx, fmt.Sprintf("employee:userid:%s", userID.String())).Err()
	}

	utils.SuccessResponse(c, http.StatusOK, "Foto profil berhasil diperbarui", gin.H{
		"face_reference_url": avatarURL,
	})
}

func (h *EmployeeHandler) UploadAvatarFile(c *gin.Context) {
	file, _, err := c.Request.FormFile("avatar")
	if err != nil {
		file, _, err = c.Request.FormFile("photo")
		if err != nil {
			file, _, err = c.Request.FormFile("file")
			if err != nil {
				utils.ErrorResponse(c, http.StatusBadRequest, "File foto tidak ditemukan dalam request")
				return
			}
		}
	}
	defer file.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Gagal membaca file foto")
		return
	}

	compressedData, err := utils.CompressImage(fileBytes, 800, 75)
	if err != nil {
		compressedData = fileBytes
	}

	uploadDir := "uploads/faces"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal membuat direktori upload")
		return
	}

	employeeIDStr := c.PostForm("employee_id")
	if employeeIDStr == "" {
		employeeIDStr = c.Query("employee_id")
	}

	var filename string
	if employeeIDStr != "" {
		filename = fmt.Sprintf("%s_%d.jpg", employeeIDStr, time.Now().Unix())
	} else {
		filename = fmt.Sprintf("%s_%d.jpg", uuid.New().String(), time.Now().Unix())
	}

	filePath := filepath.Join(uploadDir, filename)
	if err := os.WriteFile(filePath, compressedData, 0644); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal menyimpan foto")
		return
	}

	avatarURL := "/" + filePath

	if employeeIDStr != "" {
		if empUUID, parseErr := uuid.Parse(employeeIDStr); parseErr == nil {
			_ = h.db.Model(&domain.Employee{}).Where("id = ?", empUUID).Update("face_reference_url", avatarURL).Error
			if redisclient.Client != nil {
				ctx := context.Background()
				_ = redisclient.Client.Del(ctx, fmt.Sprintf("employee:id:%s", empUUID.String())).Err()
			}
		}
	}

	utils.SuccessResponse(c, http.StatusOK, "Foto profil berhasil diunggah", gin.H{
		"url":                avatarURL,
		"face_reference_url": avatarURL,
	})
}


