package http

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"example.com/hris/backend/internal/repository"
	"example.com/hris/backend/internal/usecase"
	geminiPkg "example.com/hris/backend/pkg/gemini"
	"example.com/hris/backend/pkg/utils"
	"github.com/xuri/excelize/v2"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type StoreHandler struct {
	storeRepo       repository.StoreRepository
	employeeUseCase usecase.EmployeeUsecase
	auditLogUseCase usecase.AuditLogUsecase
	geminiClient    *geminiPkg.Client
}

func NewStoreHandler(storeRepo repository.StoreRepository, employeeUseCase usecase.EmployeeUsecase, auditLogUseCase usecase.AuditLogUsecase, geminiClient *geminiPkg.Client) *StoreHandler {
	return &StoreHandler{
		storeRepo:       storeRepo,
		employeeUseCase: employeeUseCase,
		auditLogUseCase: auditLogUseCase,
		geminiClient:    geminiClient,
	}
}

func (h *StoreHandler) SetupRoutes(r *gin.RouterGroup) {
	stores := r.Group("/stores")
	{
		stores.GET("", h.GetAll)
		stores.GET("/distribution-map", h.GetDistributionMap)
		stores.GET("/export", h.ExportExcel)
		stores.POST("/check-duplicate", h.CheckDuplicate)
		stores.POST("/import/preview", h.ImportPreview)
		stores.POST("/import", h.ImportProcess)
		stores.GET("/:id", h.GetByID)
		stores.POST("", h.Create)
		stores.PUT("/:id", h.Update)
		stores.DELETE("/:id", h.Delete)
		stores.POST("/transfer", h.TransferStores)
		stores.POST("/sync-regions", h.SyncRegions)
		stores.POST("/:id/audit", h.AuditStore)
		stores.GET("/audits/all", h.GetAllAudits)
		stores.GET("/:id/audits", h.GetStoreAudits)
	}
}

// GET /api/v1/stores?assigned_employee_id=<uuid>&company_id=<uuid>&page=1&limit=10&search=...&status=ALL
func (h *StoreHandler) GetAll(c *gin.Context) {
	// Parse pagination & filter
	page := utils.StringToInt(c.Query("page"), 1)
	limit := utils.StringToInt(c.Query("limit"), 1000) // Default besar untuk backward compatibility mobile app
	search := c.Query("search")
	status := c.Query("status")
	if status == "" {
		status = "ACTIVE"
	}
	offset := (page - 1) * limit

	var companyID *uuid.UUID
	companyIDStr := c.Query("company_id")
	if companyIDStr != "" {
		if parsed, err := uuid.Parse(companyIDStr); err == nil {
			companyID = &parsed
		}
	}

	// 1. Resolve User & Employee Profile
	var currentEmployee *domain.Employee
	role, _ := c.Get("role")
	roleStr := ""
	if r, ok := role.(string); ok {
		roleStr = r
	} else if r, ok := role.(domain.Role); ok {
		roleStr = string(r)
	}

	if userIDStr, exists := c.Get("userID"); exists {
		userID := userIDStr.(uuid.UUID)
		if emp, err := h.employeeUseCase.GetEmployeeByUserID(userID); err == nil && emp != nil {
			currentEmployee = emp
		}
	}

	// 2. Determine permissions (Top Management vs Supervisor vs Ordinary Sales vs Admin/Office)
	isTopManagement := roleStr == string(domain.RoleSuperAdmin) || roleStr == "owner" || roleStr == "direktur" || roleStr == "director"
	isManagementOrSpv := isTopManagement || roleStr == string(domain.RoleManager) || roleStr == string(domain.RoleSupervisorSales) || roleStr == string(domain.RoleSupervisorSalesMT)
	isAdminOrOffice := isManagementOrSpv ||
		roleStr == string(domain.RoleAdminNota) ||
		roleStr == string(domain.RoleHRAdmin) ||
		roleStr == string(domain.RoleAdminKasir) ||
		roleStr == string(domain.RoleCashier) ||
		roleStr == string(domain.RoleCashierPusat) ||
		roleStr == string(domain.RoleAdminGudang) ||
		roleStr == string(domain.RoleSupervisorGudang) ||
		roleStr == string(domain.RoleAdminDesigner) ||
		roleStr == string(domain.RoleValidatorDesain) ||
		roleStr == string(domain.RoleValidatorToko) ||
		roleStr == string(domain.RoleAdminPurchasing) ||
		roleStr == string(domain.RoleAdminMarketplace) ||
		roleStr == string(domain.RoleAdminOperasional) ||
		roleStr == string(domain.RoleAdminPabrik) ||
		roleStr == string(domain.RoleAuditor) ||
		roleStr == string(domain.RoleCustomerSupport)

	if currentEmployee != nil && currentEmployee.JobPosition != nil {
		title := strings.ToLower(currentEmployee.JobPosition.Title)
		if strings.Contains(title, "direktur") || strings.Contains(title, "director") || strings.Contains(title, "owner") || strings.Contains(title, "superadmin") {
			isTopManagement = true
		}
		if isTopManagement || strings.Contains(title, "manager") || strings.Contains(title, "spv") || strings.Contains(title, "supervisor") || strings.Contains(title, "head") {
			isManagementOrSpv = true
			isAdminOrOffice = true
		}
		if strings.Contains(title, "admin") || strings.Contains(title, "nota") || strings.Contains(title, "kasir") ||
			strings.Contains(title, "cashier") || strings.Contains(title, "gudang") || strings.Contains(title, "finance") ||
			strings.Contains(title, "accounting") || strings.Contains(title, "operasional") || strings.Contains(title, "purchasing") ||
			strings.Contains(title, "designer") || strings.Contains(title, "validator") || strings.Contains(title, "auditor") ||
			strings.Contains(title, "hr") || strings.Contains(title, "operator") || strings.Contains(title, "support") {
			isAdminOrOffice = true
		}
	}

	// 3. Resolve Branch ID Filter
	var branchID *uuid.UUID
	branchIDStr := c.Query("branch_id")
	if branchIDStr != "" && branchIDStr != "ALL_BRANCHES" {
		if parsed, err := uuid.Parse(branchIDStr); err == nil {
			branchID = &parsed
		}
	} else if branchIDStr == "" && !isTopManagement && currentEmployee != nil && currentEmployee.BranchID != nil {
		// Auto-scope to supervisor's / employee's branch when not explicitly specified
		branchID = currentEmployee.BranchID
	}

	// 4. Resolve Assigned Employee ID Filter
	var assignedEmpID *uuid.UUID
	assignedEmpIDStr := c.Query("assigned_employee_id")
	if assignedEmpIDStr != "" && assignedEmpIDStr != "ALL" {
		if parsed, err := uuid.Parse(assignedEmpIDStr); err == nil {
			assignedEmpID = &parsed
		}
	} else if assignedEmpIDStr == "" && !isAdminOrOffice && currentEmployee != nil {
		// Ordinary sales: restrict to their own stores
		assignedEmpID = &currentEmployee.ID
	}

	category := c.Query("category")
	sortBy := c.Query("sort_by")
	salesType := c.Query("sales_type")
	completeness := c.Query("completeness")
	minimal := c.Query("minimal") == "true" || c.Query("simple") == "true" || c.Query("include_metrics") == "false"
	stores, total, err := h.storeRepo.FindPaginated(companyID, branchID, assignedEmpID, category, search, status, sortBy, limit, offset, salesType, completeness, !minimal)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": stores,
		"meta": gin.H{
			"total": total,
			"page":  page,
			"limit": limit,
		},
	})
}

// GET /api/v1/stores/distribution-map?company_id=<uuid>&branch_id=<uuid>&sales_type=...&status=ACTIVE
func (h *StoreHandler) GetDistributionMap(c *gin.Context) {
	var companyID *uuid.UUID
	companyIDStr := c.Query("company_id")
	if companyIDStr != "" {
		if parsed, err := uuid.Parse(companyIDStr); err == nil {
			companyID = &parsed
		}
	}

	role, _ := c.Get("role")
	roleStr := ""
	if r, ok := role.(string); ok {
		roleStr = r
	} else if r, ok := role.(domain.Role); ok {
		roleStr = string(r)
	}

	var currentEmployee *domain.Employee
	if userIDStr, exists := c.Get("userID"); exists {
		userID := userIDStr.(uuid.UUID)
		if emp, err := h.employeeUseCase.GetEmployeeByUserID(userID); err == nil && emp != nil {
			currentEmployee = emp
		}
	}

	isTopManagement := roleStr == string(domain.RoleSuperAdmin) || roleStr == "owner" || roleStr == "direktur" || roleStr == "director"
	isManagementOrSpv := isTopManagement || roleStr == string(domain.RoleManager) || roleStr == string(domain.RoleSupervisorSales) || roleStr == string(domain.RoleSupervisorSalesMT)
	isAdminOrOffice := isManagementOrSpv ||
		roleStr == string(domain.RoleAdminNota) ||
		roleStr == string(domain.RoleHRAdmin) ||
		roleStr == string(domain.RoleAdminKasir) ||
		roleStr == string(domain.RoleCashier) ||
		roleStr == string(domain.RoleCashierPusat) ||
		roleStr == string(domain.RoleAdminGudang) ||
		roleStr == string(domain.RoleSupervisorGudang) ||
		roleStr == string(domain.RoleAdminDesigner) ||
		roleStr == string(domain.RoleValidatorDesain) ||
		roleStr == string(domain.RoleValidatorToko) ||
		roleStr == string(domain.RoleAdminPurchasing) ||
		roleStr == string(domain.RoleAdminMarketplace) ||
		roleStr == string(domain.RoleAdminOperasional) ||
		roleStr == string(domain.RoleAdminPabrik) ||
		roleStr == string(domain.RoleAuditor) ||
		roleStr == string(domain.RoleCustomerSupport)

	if currentEmployee != nil && currentEmployee.JobPosition != nil {
		title := strings.ToLower(currentEmployee.JobPosition.Title)
		if strings.Contains(title, "direktur") || strings.Contains(title, "director") || strings.Contains(title, "owner") || strings.Contains(title, "superadmin") {
			isTopManagement = true
		}
		if isTopManagement || strings.Contains(title, "manager") || strings.Contains(title, "spv") || strings.Contains(title, "supervisor") || strings.Contains(title, "head") {
			isManagementOrSpv = true
			isAdminOrOffice = true
		}
		if strings.Contains(title, "admin") || strings.Contains(title, "nota") || strings.Contains(title, "kasir") ||
			strings.Contains(title, "cashier") || strings.Contains(title, "gudang") || strings.Contains(title, "finance") ||
			strings.Contains(title, "accounting") || strings.Contains(title, "operasional") || strings.Contains(title, "purchasing") ||
			strings.Contains(title, "designer") || strings.Contains(title, "validator") || strings.Contains(title, "auditor") ||
			strings.Contains(title, "hr") || strings.Contains(title, "operator") || strings.Contains(title, "support") {
			isAdminOrOffice = true
		}
	}

	var branchIDs []uuid.UUID
	branchIDStr := c.Query("branch_id")
	if branchIDStr != "" && branchIDStr != "ALL_BRANCHES" {
		branchIDs = utils.ParseUUIDs(branchIDStr)
	} else if branchIDStr == "" && !isTopManagement && currentEmployee != nil && currentEmployee.BranchID != nil {
		branchIDs = []uuid.UUID{*currentEmployee.BranchID}
	}

	var assignedEmpID *uuid.UUID
	assignedEmpIDStr := c.Query("assigned_employee_id")
	if assignedEmpIDStr != "" && assignedEmpIDStr != "ALL" {
		if parsed, err := uuid.Parse(assignedEmpIDStr); err == nil {
			assignedEmpID = &parsed
		}
	} else if assignedEmpIDStr == "" && !isAdminOrOffice && currentEmployee != nil {
		assignedEmpID = &currentEmployee.ID
	}

	status := c.Query("status")
	if status == "" {
		status = "ACTIVE"
	}
	salesType := c.Query("sales_type")

	items, err := h.storeRepo.GetStoreDistribution(companyID, branchIDs, assignedEmpID, salesType, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	mappedCount := 0
	for _, item := range items {
		if item.Latitude != 0 && item.Longitude != 0 {
			mappedCount++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Success retrieving store distribution map",
		"total":        len(items),
		"mapped_total": mappedCount,
		"data":         items,
	})
}

// GET /api/v1/stores/:id
func (h *StoreHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid store ID"})
		return
	}

	store, err := h.storeRepo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if store == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Store not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Success", "data": store})
}

// POST /api/v1/stores/check-duplicate
func (h *StoreHandler) CheckDuplicate(c *gin.Context) {
	var req domain.CheckDuplicateStoreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Name = c.PostForm("name")
		req.PhoneNumber = c.PostForm("phone_number")
		req.District = c.PostForm("district")
		req.Regency = c.PostForm("regency")
		req.Province = c.PostForm("province")
		if latStr := c.PostForm("latitude"); latStr != "" {
			req.Latitude = utils.StringToFloat(latStr, 0)
		}
		if lonStr := c.PostForm("longitude"); lonStr != "" {
			req.Longitude = utils.StringToFloat(lonStr, 0)
		}
		if exclStr := c.PostForm("exclude_id"); exclStr != "" {
			if parsed, err := uuid.Parse(exclStr); err == nil {
				req.ExcludeID = &parsed
			}
		}
	}

	var currentEmpID *uuid.UUID
	if userIDStr, exists := c.Get("userID"); exists {
		userID := userIDStr.(uuid.UUID)
		employee, err := h.employeeUseCase.GetEmployeeByUserID(userID)
		if err == nil && employee != nil {
			currentEmpID = &employee.ID
		}
	}

	res, err := h.storeRepo.CheckDuplicate(req, currentEmpID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check duplicate stores: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Duplicate check completed",
		"data":    res,
	})
}

// POST /api/v1/stores
func (h *StoreHandler) Create(c *gin.Context) {
	var store domain.Store

	// Handle multipart form manual binding
	h.manualBindStore(c, &store)

	// Fallback to JSON binding
	if store.Name == "" {
		if err := c.ShouldBindJSON(&store); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	// Validate required fields
	if store.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}

	// Support client-generated ID / Idempotency for offline sync
	if store.ID != uuid.Nil {
		if existing, err := h.storeRepo.FindByID(store.ID); err == nil && existing != nil {
			c.JSON(http.StatusOK, gin.H{"message": "Store already exists (idempotent)", "data": existing})
			return
		}
	}

	// Ensure company_id is cleaned if bind failed
	cID := h.getCleanFormValue(c, "company_id")
	if cID != "" {
		if parsed, err := uuid.Parse(cID); err == nil {
			store.CompanyID = parsed
		}
	}

	// Handle Image Upload
	photoURL, err := h.handleImageUpload(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload image: " + err.Error()})
		return
	}
	if photoURL != "" {
		store.PhotoURL = photoURL
	}

	// Auto assign to current employee if logged in
	var userName string
	var companyID uuid.UUID
	var userID uuid.UUID
	if userIDStr, exists := c.Get("userID"); exists {
		userID = userIDStr.(uuid.UUID)
		employee, err := h.employeeUseCase.GetEmployeeByUserID(userID)
		if err == nil && employee != nil {
			store.CreatedByEmployeeID = &employee.ID
			userName = employee.FirstName + " " + employee.LastName
			if employee.CompanyID != nil {
				companyID = *employee.CompanyID
				if store.CompanyID == uuid.Nil {
					store.CompanyID = *employee.CompanyID
				}
			}
			if store.AssignedEmployeeID == nil || *store.AssignedEmployeeID == uuid.Nil {
				store.AssignedEmployeeID = &employee.ID
			}
			// Hanya manajemen yang boleh menunjuk PIC lain. Sales biasa selalu
			// menjadi PIC dirinya sendiri, walaupun client mengirim UUID lain.
			management := false
			if employee.JobPosition != nil {
				title := strings.ToLower(employee.JobPosition.Title)
				management = strings.Contains(title, "direktur") || strings.Contains(title, "director") ||
					strings.Contains(title, "manager") || strings.Contains(title, "supervisor") ||
					strings.Contains(title, "spv") || strings.Contains(title, "owner")
			}
			if !management {
				store.AssignedEmployeeID = &employee.ID
			}
		}
	}

	// company_id boleh kosong dari client (misalnya saat memilih Semua Cabang).
	// Setelah fallback ke company user, gunakan company default yang dikonfigurasi
	// untuk toko yang dibuat dari menu manajemen toko.
	if store.CompanyID == uuid.Nil {
		fallbackCompanyID, lookupErr := h.storeRepo.FindCompanyIDByName(os.Getenv("TRIAL_COMPANY_NAME"))
		if lookupErr != nil || fallbackCompanyID == uuid.Nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Perusahaan demo tidak ditemukan"})
			return
		}
		store.CompanyID = fallbackCompanyID
	}

	if err := h.storeRepo.Create(&store); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if store.Regency == "" && store.District == "" && store.Village == "" {
		if store.Latitude != 0 && store.Longitude != 0 {
			go h.backgroundGeocodeStore(store.ID, store.Latitude, store.Longitude)
		} else if store.Address != "" {
			go h.backgroundGeocodeStoreByAddress(store.ID, store.Address)
		}
	}
	h.auditLogUseCase.Log(companyID, userID, userName, "CREATE", "Store", store.ID.String(), "Created Store: "+store.Name, nil, store, c.ClientIP())

	c.JSON(http.StatusCreated, gin.H{"message": "Store created successfully", "data": store})
}

// PUT /api/v1/stores/:id
func (h *StoreHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid store ID"})
		return
	}

	existing, err := h.storeRepo.FindByID(id)
	if err != nil || existing == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Store not found"})
		return
	}

	// Deep copy existing object to log old value
	var oldStore domain.Store
	oldStoreBytes, _ := json.Marshal(existing)
	_ = json.Unmarshal(oldStoreBytes, &oldStore)

	if err := c.ShouldBind(existing); err != nil {
		// Fallback for anomaly
		h.manualBindStore(c, existing)
	}

	// Ensure company_id is cleaned if bind failed
	cID := h.getCleanFormValue(c, "company_id")
	if cID != "" {
		parsed, err := uuid.Parse(cID)
		if err == nil {
			existing.CompanyID = parsed
			existing.Company = nil
		}
	}

	isForm := strings.Contains(c.GetHeader("Content-Type"), "form")

	// Ensure assigned_employee_id is updated only if explicitly provided in form
	if isForm {
		_ = c.Request.ParseMultipartForm(32 << 20)
		hasAssignedKey := false
		if c.Request.PostForm != nil {
			_, hasAssignedKey = c.Request.PostForm["assigned_employee_id"]
		}
		if !hasAssignedKey && c.Request.MultipartForm != nil && c.Request.MultipartForm.Value != nil {
			_, hasAssignedKey = c.Request.MultipartForm.Value["assigned_employee_id"]
		}

		if hasAssignedKey {
			eID := h.getCleanFormValue(c, "assigned_employee_id")
			if eID != "" && eID != "null" && eID != "unassigned" {
				if parsed, err := uuid.Parse(eID); err == nil {
					existing.AssignedEmployeeID = &parsed
					existing.AssignedEmployee = nil
				}
			} else {
				// Explicitly set to nil if empty string / unassigned sent via form
				existing.AssignedEmployeeID = nil
				existing.AssignedEmployee = nil
			}
		}
	}

	// Handle Image Upload
	photoURL, err := h.handleImageUpload(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload image: " + err.Error()})
		return
	}
	if photoURL != "" {
		existing.PhotoURL = photoURL
	}
	existing.ID = id // ensure ID not overwritten

	if err := h.storeRepo.Update(existing); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Geocode if location fields changed and region is empty
	if existing.Latitude != oldStore.Latitude || existing.Longitude != oldStore.Longitude || existing.Address != oldStore.Address {
		if existing.Regency == "" && existing.District == "" && existing.Village == "" {
			if existing.Latitude != 0 && existing.Longitude != 0 {
				go h.backgroundGeocodeStore(existing.ID, existing.Latitude, existing.Longitude)
			} else if existing.Address != "" {
				go h.backgroundGeocodeStoreByAddress(existing.ID, existing.Address)
			}
		}
	}

	var userName string
	var companyID uuid.UUID
	var userID uuid.UUID
	if userIDStr, exists := c.Get("userID"); exists {
		userID = userIDStr.(uuid.UUID)
		employee, err := h.employeeUseCase.GetEmployeeByUserID(userID)
		if err == nil && employee != nil {
			userName = employee.FirstName + " " + employee.LastName
			if employee.CompanyID != nil {
				companyID = *employee.CompanyID
			}
		}
	}

	h.auditLogUseCase.Log(companyID, userID, userName, "UPDATE", "Store", existing.ID.String(), "Updated Store: "+existing.Name, oldStore, existing, c.ClientIP())

	c.JSON(http.StatusOK, gin.H{"message": "Store updated successfully", "data": existing})
}

// DELETE /api/v1/stores/:id
func (h *StoreHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid store ID"})
		return
	}

	existing, err := h.storeRepo.FindByID(id)
	if err != nil || existing == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Store not found"})
		return
	}

	if err := h.storeRepo.Delete(id); err != nil {
		if strings.Contains(err.Error(), "toko tidak bisa dihapus") {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	var userName string
	var companyID uuid.UUID
	var userID uuid.UUID
	if userIDStr, exists := c.Get("userID"); exists {
		userID = userIDStr.(uuid.UUID)
		employee, err := h.employeeUseCase.GetEmployeeByUserID(userID)
		if err == nil && employee != nil {
			userName = employee.FirstName + " " + employee.LastName
			if employee.CompanyID != nil {
				companyID = *employee.CompanyID
			}
		}
	}

	h.auditLogUseCase.Log(companyID, userID, userName, "DELETE", "Store", id.String(), "Deleted Store: "+existing.Name, existing, nil, c.ClientIP())

	c.JSON(http.StatusOK, gin.H{"message": "Store deleted successfully"})
}

func (h *StoreHandler) handleImageUpload(c *gin.Context) (string, error) {
	contentType := c.GetHeader("Content-Type")
	if !strings.HasPrefix(contentType, "multipart/form-data") {
		return "", nil
	}

	file, err := c.FormFile("photo")
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
	uploadDir := "uploads/stores"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return "", err
	}

	savePath := filepath.Join(uploadDir, filename)
	if err := utils.SaveUploadedFile(file, savePath); err != nil {
		return "", err
	}

	return "/uploads/stores/" + filename, nil
}

func (h *StoreHandler) getCleanFormValue(c *gin.Context, key string) string {
	val := c.Request.PostFormValue(key)
	if val == "" {
		return ""
	}
	// Remove brackets and quotes if present
	val = strings.TrimSpace(val)
	val = strings.TrimPrefix(val, "[")
	val = strings.TrimSuffix(val, "]")
	val = strings.Trim(val, "\"")
	return val
}

func (h *StoreHandler) manualBindStore(c *gin.Context, store *domain.Store) {
	// Manual extraction for problematic fields when using multipart/form-data
	if idStr := h.getCleanFormValue(c, "id"); idStr != "" {
		if parsedID, err := uuid.Parse(idStr); err == nil && parsedID != uuid.Nil {
			store.ID = parsedID
		}
	} else if clientID := h.getCleanFormValue(c, "client_id"); clientID != "" {
		if parsedID, err := uuid.Parse(clientID); err == nil && parsedID != uuid.Nil {
			store.ID = parsedID
		}
	}
	if name := c.Request.PostFormValue("name"); name != "" {
		store.Name = name
	}
	if owner := c.Request.PostFormValue("owner_name"); owner != "" {
		store.OwnerName = owner
	}
	if phone := c.Request.PostFormValue("phone_number"); phone != "" {
		store.PhoneNumber = phone
	}
	if addr := c.Request.PostFormValue("address"); addr != "" {
		store.Address = addr
	}
	if lat := c.Request.PostFormValue("latitude"); lat != "" {
		fmt.Sscanf(lat, "%f", &store.Latitude)
	}
	if lng := c.Request.PostFormValue("longitude"); lng != "" {
		fmt.Sscanf(lng, "%f", &store.Longitude)
	}
	if active := c.Request.PostFormValue("is_active"); active != "" {
		store.IsActive = active == "true"
	}
	if prov := c.Request.PostFormValue("province"); prov != "" {
		store.Province = prov
	}
	if reg := c.Request.PostFormValue("regency"); reg != "" {
		store.Regency = reg
	}
	if dist := c.Request.PostFormValue("district"); dist != "" {
		store.District = dist
	}
	if vil := c.Request.PostFormValue("village"); vil != "" {
		store.Village = vil
	}
	if freq := c.Request.PostFormValue("visit_frequency"); freq != "" {
		store.VisitFrequency = freq
	}
	if category := c.Request.PostFormValue("category"); category != "" {
		store.Category = category
	}
	if custType := c.Request.PostFormValue("customer_type"); custType != "" {
		store.CustomerType = custType
	}
	if subCustType := c.Request.PostFormValue("sub_customer_type"); subCustType != "" {
		store.SubCustomerType = subCustType
	}
	if distCity := c.Request.PostFormValue("distribution_city"); distCity != "" {
		store.DistributionCity = distCity
	}
	if assignedEmp := h.getCleanFormValue(c, "assigned_employee_id"); assignedEmp != "" && assignedEmp != "null" && assignedEmp != "unassigned" {
		if parsed, err := uuid.Parse(assignedEmp); err == nil {
			store.AssignedEmployeeID = &parsed
		}
	}
	if npwp := h.getCleanFormValue(c, "npwp"); npwp != "" {
		store.NPWP = &npwp
	}
	if nik := h.getCleanFormValue(c, "nik"); nik != "" {
		store.NIK = &nik
	}
	if taxName := h.getCleanFormValue(c, "tax_name"); taxName != "" {
		store.TaxName = &taxName
	}

	// Parse visit_days array from form fields
	days := c.Request.PostForm["visit_days"]
	if len(days) == 0 {
		days = c.Request.PostForm["visit_days[]"]
	}
	if len(days) > 0 {
		var intDays []int64
		for _, dayStr := range days {
			dayStr = strings.Trim(dayStr, "[]")
			parts := strings.Split(dayStr, ",")
			for _, part := range parts {
				part = strings.TrimSpace(part)
				if part != "" {
					var d int64
					if _, err := fmt.Sscanf(part, "%d", &d); err == nil {
						intDays = append(intDays, d)
					}
				}
			}
		}
		store.VisitDays = intDays
	}
}

type TransferStoresRequest struct {
	FromEmployeeID uuid.UUID `json:"from_employee_id" binding:"required"`
	ToEmployeeID   uuid.UUID `json:"to_employee_id" binding:"required"`
}

// POST /api/v1/stores/transfer
func (h *StoreHandler) TransferStores(c *gin.Context) {
	var req TransferStoresRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	rowsAffected, err := h.storeRepo.TransferStores(req.FromEmployeeID, req.ToEmployeeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal melakukan transfer toko: " + err.Error()})
		return
	}

	// Log audit trail
	var userName string
	var companyID uuid.UUID
	var userID uuid.UUID
	if userIDStr, exists := c.Get("userID"); exists {
		userID = userIDStr.(uuid.UUID)
		employee, err := h.employeeUseCase.GetEmployeeByUserID(userID)
		if err == nil && employee != nil {
			userName = employee.FirstName + " " + employee.LastName
			if employee.CompanyID != nil {
				companyID = *employee.CompanyID
			}
		}
	}

	h.auditLogUseCase.Log(companyID, userID, userName, "TRANSFER", "Store", req.FromEmployeeID.String(),
		fmt.Sprintf("Transferred %d stores to employee %s", rowsAffected, req.ToEmployeeID.String()),
		nil, nil, c.ClientIP())

	c.JSON(http.StatusOK, gin.H{
		"message": "Transfer toko berhasil",
		"data": gin.H{
			"stores_transferred": rowsAffected,
		},
	})
}

func (h *StoreHandler) backgroundGeocodeStore(storeID uuid.UUID, lat, lon float64) {
	url := fmt.Sprintf("https://nominatim.openstreetmap.org/reverse?format=json&lat=%f&lon=%f&zoom=18&addressdetails=1", lat, lon)

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "KarsaHR/1.0")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	var result struct {
		Address map[string]string `json:"address"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return
	}

	addr := result.Address

	regency := addr["city"]
	if regency == "" {
		regency = addr["county"]
	}
	if regency == "" {
		regency = addr["state_district"]
	}

	district := addr["suburb"]
	if district == "" {
		district = addr["city_district"]
	}

	village := addr["village"]
	if village == "" {
		village = addr["neighbourhood"]
	}

	if regency == "" || district == "" || village == "" {
		if h.geminiClient != nil {
			// Fallback to Gemini AI for missing parts
			prompt := fmt.Sprintf(`Berdasarkan koordinat latitude %f dan longitude %f (di Indonesia), sebutkan Kabupaten/Kota, Kecamatan, dan Desa/Kelurahannya. 
Jika kamu tidak yakin, tebaklah yang paling mendekati berdasarkan data geografismu.
Balas HANYA dengan format JSON persis seperti ini tanpa tambahan karakter apapun:
{"regency": "Nama Kabupaten/Kota", "district": "Nama Kecamatan", "village": "Nama Desa"}`, lat, lon)

			aiResponse, err := h.geminiClient.GenerateText(prompt)
			if err != nil {
				fmt.Printf("[StoreHandler] Gemini API Error for lat %f, lon %f: %v\n", lat, lon, err)
			} else if aiResponse != "" {
				cleanRes := strings.TrimSpace(aiResponse)
				cleanRes = strings.TrimPrefix(cleanRes, "```json")
				cleanRes = strings.TrimPrefix(cleanRes, "```")
				cleanRes = strings.TrimSuffix(cleanRes, "```")
				cleanRes = strings.TrimSpace(cleanRes)

				var regionData struct {
					Regency  string `json:"regency"`
					District string `json:"district"`
					Village  string `json:"village"`
				}

				if err := json.Unmarshal([]byte(cleanRes), &regionData); err == nil {
					if regency == "" {
						regency = regionData.Regency
					}
					if district == "" {
						district = regionData.District
					}
					if village == "" {
						village = regionData.Village
					}
				} else {
					fmt.Printf("[StoreHandler] Failed to unmarshal Gemini JSON. Error: %v. Raw Response: %s\n", err, cleanRes)
				}
			}
		}
	}

	if regency != "" || district != "" || village != "" {
		regency, district, village = sanitizeStoreHierarchy(regency, district, village, "")
		_ = h.storeRepo.UpdateRegion(storeID, "", regency, district, village)
	}
}

func sanitizeStoreHierarchy(rawRegency, rawDistrict, rawVillage, address string) (string, string, string) {
	reg := strings.TrimSpace(rawRegency)
	dist := strings.TrimSpace(rawDistrict)
	vill := strings.TrimSpace(rawVillage)

	cleanRegUpper := strings.ToUpper(strings.TrimSpace(reg))
	cleanDistUpper := strings.ToUpper(strings.TrimSpace(dist))
	cleanVillUpper := strings.ToUpper(strings.TrimSpace(vill))

	// 1. Check if regency is actually a known village/subdistrict (e.g. Donohudan, Ngesrep, Sobokerto)
	if subInfo, ok := usecase.KnownSubRegionsMap[cleanRegUpper]; ok {
		if vill == "" && subInfo.Village != "" {
			vill = subInfo.Village
		}
		if dist == "" && subInfo.District != "" {
			dist = subInfo.District
		}
		reg = subInfo.Regency
	} else if subInfo, ok := usecase.KnownSubRegionsMap[cleanDistUpper]; ok {
		if reg == "" || !usecase.IsValidRegency(reg) {
			reg = subInfo.Regency
		}
		if vill == "" && subInfo.Village != "" {
			vill = subInfo.Village
		}
		if subInfo.District != "" {
			dist = subInfo.District
		}
	} else if subInfo, ok := usecase.KnownSubRegionsMap[cleanVillUpper]; ok {
		if reg == "" || !usecase.IsValidRegency(reg) {
			reg = subInfo.Regency
		}
		if dist == "" && subInfo.District != "" {
			dist = subInfo.District
		}
	}

	// 2. Validate regency against official list
	if reg != "" && !usecase.IsValidRegency(reg) {
		cleanCity, _ := usecase.NormalizeCityAndProvince(reg, "", address)
		if cleanCity != "Tidak Terdeteksi" {
			reg = cleanCity
		}
	}

	return reg, dist, vill
}

func (h *StoreHandler) backgroundGeocodeStoreByAddress(storeID uuid.UUID, address string) {
	if address == "" {
		return
	}

	url := fmt.Sprintf("https://nominatim.openstreetmap.org/search?q=%s&format=json&addressdetails=1&limit=1", strings.ReplaceAll(address, " ", "+"))
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "KarsaHR/1.0")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	var results []struct {
		Address map[string]string `json:"address"`
	}
	if err := json.Unmarshal(body, &results); err != nil || len(results) == 0 {
		return
	}

	addr := results[0].Address

	regency := addr["city"]
	if regency == "" {
		regency = addr["county"]
	}
	if regency == "" {
		regency = addr["state_district"]
	}

	district := addr["suburb"]
	if district == "" {
		district = addr["city_district"]
	}

	village := addr["village"]
	if village == "" {
		village = addr["neighbourhood"]
	}

	if regency != "" || district != "" || village != "" {
		regency, district, village = sanitizeStoreHierarchy(regency, district, village, address)
		_ = h.storeRepo.UpdateRegion(storeID, "", regency, district, village)
	}
}

func (h *StoreHandler) SyncRegions(c *gin.Context) {
	// Find all stores where region is empty (lightweight query without financial subqueries)
	stores, _, err := h.storeRepo.FindPaginated(nil, nil, nil, "", "", "ALL", "", 10000, 0, "", "", false)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch stores"})
		return
	}

	var storesToSync []domain.Store
	for _, s := range stores {
		if s.Regency == "" || s.District == "" || s.Village == "" {
			storesToSync = append(storesToSync, s)
		}
	}

	if len(storesToSync) == 0 {
		c.JSON(http.StatusOK, gin.H{"message": "Semua toko sudah terdeteksi wilayahnya."})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Memulai sinkronisasi wilayah untuk %d toko di latar belakang...", len(storesToSync)),
		"count":   len(storesToSync),
	})

	go func() {
		for _, s := range storesToSync {
			if s.Latitude != 0 && s.Longitude != 0 {
				h.backgroundGeocodeStore(s.ID, s.Latitude, s.Longitude)
				time.Sleep(1100 * time.Millisecond) // Hormati Rate Limit OSM Nominatim
			} else if s.Address != "" {
				h.backgroundGeocodeStoreByAddress(s.ID, s.Address)
				time.Sleep(1100 * time.Millisecond)
			}
		}
	}()
}

// GET /api/v1/stores/export
func (h *StoreHandler) ExportExcel(c *gin.Context) {
	// ── 1. Parse filter parameter yang sama dengan GetAll ───────────────────
	search := c.Query("search")
	status := c.Query("status")
	if status == "" {
		status = "ALL"
	}

	var companyID *uuid.UUID
	if cIDStr := c.Query("company_id"); cIDStr != "" {
		if parsed, err := uuid.Parse(cIDStr); err == nil {
			companyID = &parsed
		}
	}

	var branchID *uuid.UUID
	if bIDStr := c.Query("branch_id"); bIDStr != "" {
		if parsed, err := uuid.Parse(bIDStr); err == nil {
			branchID = &parsed
		}
	}

	var assignedEmpID *uuid.UUID
	if eIDStr := c.Query("assigned_employee_id"); eIDStr != "" && eIDStr != "ALL" {
		if parsed, err := uuid.Parse(eIDStr); err == nil {
			assignedEmpID = &parsed
		}
	}

	category := c.Query("category")
	salesType := c.Query("sales_type")
	completeness := c.Query("completeness")

	// ── 2. Fetch semua data sesuai filter (limit besar, tanpa pagination) ────
	stores, total, err := h.storeRepo.FindPaginated(companyID, branchID, assignedEmpID, category, search, status, "", 10000, 0, salesType, completeness)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data toko: " + err.Error()})
		return
	}

	// ── 3. Bangun file Excel ──────────────────────────────────────────────────
	xlsxBytes, err := h.buildStoreExcel(stores, total, search, status, category, assignedEmpID, salesType, completeness)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat file Excel: " + err.Error()})
		return
	}

	filename := fmt.Sprintf("kelola_toko_%s.xlsx", time.Now().Format("20060102_150405"))
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", xlsxBytes)
}

func (h *StoreHandler) buildStoreExcel(stores []domain.Store, total int64, search, status, category string, assignedEmpID *uuid.UUID, salesType string, completeness string) ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()

	const sheet = "Data Toko"
	f.SetSheetName("Sheet1", sheet)

	showGridLines := true
	_ = f.SetSheetView(sheet, 0, &excelize.ViewOptions{ShowGridLines: &showGridLines})

	// ── Styles ────────────────────────────────────────────────────────────────
	borderThin := []excelize.Border{
		{Type: "top", Color: "D0D5DD", Style: 1},
		{Type: "bottom", Color: "D0D5DD", Style: 1},
		{Type: "left", Color: "D0D5DD", Style: 1},
		{Type: "right", Color: "D0D5DD", Style: 1},
	}

	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 18, Color: "1F3864"},
		Alignment: &excelize.Alignment{Vertical: "center"},
	})
	subtitleStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Italic: true, Size: 11, Color: "595959"},
		Alignment: &excelize.Alignment{Vertical: "center"},
	})
	metaStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: 10, Color: "808080"},
		Alignment: &excelize.Alignment{Vertical: "center"},
	})
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"2F3542"}, Pattern: 1},
		Font:      &excelize.Font{Bold: true, Color: "FFFFFF", Size: 10},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
		Border:    borderThin,
	})
	rowOddStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: 10, Color: "1A1A2E"},
		Alignment: &excelize.Alignment{Vertical: "center", WrapText: true},
		Border:    borderThin,
	})
	rowEvenStyle, _ := f.NewStyle(&excelize.Style{
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"F8F9FA"}, Pattern: 1},
		Font:      &excelize.Font{Size: 10, Color: "1A1A2E"},
		Alignment: &excelize.Alignment{Vertical: "center", WrapText: true},
		Border:    borderThin,
	})
	piutangStyle, _ := f.NewStyle(&excelize.Style{
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"FFF5F5"}, Pattern: 1},
		Font:      &excelize.Font{Bold: true, Size: 10, Color: "C0392B"},
		Alignment: &excelize.Alignment{Horizontal: "right", Vertical: "center"},
		Border:    borderThin,
		NumFmt:    3, // #,##0 untuk angka dengan separator ribuan
	})
	piutangZeroStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: 10, Color: "27AE60"},
		Alignment: &excelize.Alignment{Horizontal: "right", Vertical: "center"},
		Border:    borderThin,
		NumFmt:    3,
	})
	overdueStyle, _ := f.NewStyle(&excelize.Style{
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"FFF0F0"}, Pattern: 1},
		Font:      &excelize.Font{Bold: true, Size: 10, Color: "E74C3C"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Border:    borderThin,
	})
	centerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: 10, Color: "1A1A2E"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Border:    borderThin,
	})
	centerEvenStyle, _ := f.NewStyle(&excelize.Style{
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"F8F9FA"}, Pattern: 1},
		Font:      &excelize.Font{Size: 10, Color: "1A1A2E"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Border:    borderThin,
	})
	summaryLabelStyle, _ := f.NewStyle(&excelize.Style{
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"EEF2FF"}, Pattern: 1},
		Font:      &excelize.Font{Bold: true, Size: 10, Color: "1F4E78"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Border:    borderThin,
	})
	summaryValueStyle, _ := f.NewStyle(&excelize.Style{
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"EEF2FF"}, Pattern: 1},
		Font:      &excelize.Font{Bold: true, Size: 10, Color: "1F4E78"},
		Alignment: &excelize.Alignment{Horizontal: "right", Vertical: "center"},
		Border:    borderThin,
		NumFmt:    3,
	})

	// ── Row 1: Judul ──────────────────────────────────────────────────────────
	judulText := "DATA TOKO CUSTOMER"
	if salesType == "sales_mt" {
		judulText = "DATA MITRA MODERN TRADE (MT)"
	}
	_ = f.SetCellValue(sheet, "A1", judulText)
	_ = f.SetCellStyle(sheet, "A1", "A1", titleStyle)
	_ = f.MergeCell(sheet, "A1", "R1")
	_ = f.SetRowHeight(sheet, 1, 32)

	// ── Row 2: Info Filter ────────────────────────────────────────────────────
	filterParts := []string{}
	statusLabel := map[string]string{"ALL": "Semua", "ACTIVE": "Aktif", "INACTIVE": "Non-Aktif"}[status]
	if statusLabel == "" {
		statusLabel = status
	}
	filterParts = append(filterParts, "Status: "+statusLabel)

	if category != "" && category != "ALL" {
		catLabel := map[string]string{"TOKO_BARU": "Toko Baru", "TOKO_LAMA": "Toko Lama"}[category]
		if catLabel == "" {
			catLabel = category
		}
		filterParts = append(filterParts, "Kategori: "+catLabel)
	} else {
		filterParts = append(filterParts, "Kategori: Semua")
	}

	if search != "" {
		filterParts = append(filterParts, "Pencarian: \""+search+"\"")
	}
	if assignedEmpID != nil {
		filterParts = append(filterParts, "Salesman: Terfilter")
	} else {
		filterParts = append(filterParts, "Salesman: Semua")
	}

	if completeness != "" && completeness != "ALL" {
		complMap := map[string]string{
			"INCOMPLETE":       "Data Kurang Lengkap",
			"COMPLETE":         "Data Lengkap",
			"MISSING_PHONE":    "Tanpa No. HP",
			"MISSING_LOCATION": "Tanpa Lokasi GPS",
			"MISSING_PHOTO":    "Tanpa Foto Toko",
			"MISSING_OWNER":    "Tanpa Nama Pemilik",
			"MISSING_SALESMAN": "Tanpa Salesman PIC",
			"MISSING_ADDRESS":  "Tanpa Alamat/Wilayah",
		}
		complLabel := complMap[completeness]
		if complLabel == "" {
			complLabel = completeness
		}
		filterParts = append(filterParts, "Kelengkapan: "+complLabel)
	}

	filterStr := strings.Join(filterParts, "  |  ")
	_ = f.SetCellValue(sheet, "A2", filterStr)
	_ = f.SetCellStyle(sheet, "A2", "A2", subtitleStyle)
	_ = f.MergeCell(sheet, "A2", "R2")
	_ = f.SetRowHeight(sheet, 2, 22)

	// ── Row 3: Meta info ──────────────────────────────────────────────────────
	metaText := fmt.Sprintf("Diekspor pada: %s  |  Total data: %d toko",
		time.Now().Format("02 January 2006, 15:04 WIB"), total)
	_ = f.SetCellValue(sheet, "A3", metaText)
	_ = f.SetCellStyle(sheet, "A3", "A3", metaStyle)
	_ = f.MergeCell(sheet, "A3", "R3")
	_ = f.SetRowHeight(sheet, 3, 18)

	// ── Row 4: Kosong sebagai spacer ──────────────────────────────────────────
	_ = f.SetRowHeight(sheet, 4, 8)

	// ── Row 5: Header Kolom ───────────────────────────────────────────────────
	headers := []struct {
		col   string
		label string
		width float64
	}{
		{"A", "No.", 5},
		{"B", "Nama Toko", 30},
		{"C", "Jenis Pelanggan", 20},
		{"D", "Pemilik", 24},
		{"E", "No. HP", 17},
		{"F", "Salesman", 24},
		{"G", "Jabatan Salesman", 20},
		{"H", "Tipe Status", 13},
		{"I", "Status Aktif", 12},
		{"J", "Kab./Kota", 20},
		{"K", "Kecamatan", 20},
		{"L", "Desa/Kel.", 20},
		{"M", "Alamat", 38},
		{"N", "Total Piutang (Rp)", 20},
		{"O", "Nota Jatuh Tempo", 16},
		{"P", "NPWP", 20},
		{"Q", "NIK", 20},
		{"R", "Waktu Registrasi", 24},
		{"S", "NOO Hari Ini", 15},
		{"T", "Status EC Hari Ini", 22},
		{"U", "Pesanan Hari Ini (Rp)", 22},
		{"V", "Total Transaksi", 16},
	}

	_ = f.SetRowHeight(sheet, 5, 36)
	for _, h := range headers {
		cell := h.col + "5"
		_ = f.SetCellValue(sheet, cell, h.label)
		_ = f.SetCellStyle(sheet, cell, cell, headerStyle)
		_ = f.SetColWidth(sheet, h.col, h.col, h.width)
	}

	// ── Row 6+: Data Baris ────────────────────────────────────────────────────
	var totalPiutang float64
	for i, s := range stores {
		rowNum := i + 6
		isEven := i%2 == 1

		baseStyle := rowOddStyle
		baseCenterStyle := centerStyle
		if isEven {
			baseStyle = rowEvenStyle
			baseCenterStyle = centerEvenStyle
		}

		_ = f.SetRowHeight(sheet, rowNum, 20)

		// Nama salesman & jabatan
		salesmanName := "Belum Ditugaskan"
		salesmanTitle := "-"
		if s.AssignedEmployee != nil {
			salesmanName = strings.TrimSpace(s.AssignedEmployee.FirstName + " " + s.AssignedEmployee.LastName)
			if s.AssignedEmployee.JobPosition != nil {
				salesmanTitle = s.AssignedEmployee.JobPosition.Title
			}
		}

		statusText := "Non-Aktif"
		if s.IsActive {
			statusText = "Aktif"
		}
		categoryText := map[string]string{"TOKO_BARU": "Toko Baru", "TOKO_LAMA": "Toko Lama"}[s.Category]
		if categoryText == "" {
			categoryText = s.Category
		}

		custTypeLabel := s.CustomerType
		if custTypeLabel == "" {
			custTypeLabel = "TOKO"
		}
		if s.CustomerType == "HOREKA" && s.SubCustomerType != "" {
			custTypeLabel = fmt.Sprintf("HOREKA (%s)", s.SubCustomerType)
		} else if s.CustomerType == "DISTRIBUTOR" && s.DistributionCity != "" {
			custTypeLabel = fmt.Sprintf("DISTRIBUTOR (%s)", s.DistributionCity)
		}

		npwp := ""
		if s.NPWP != nil {
			npwp = *s.NPWP
		}
		nik := ""
		if s.NIK != nil {
			nik = *s.NIK
		}

		createdAt := s.CreatedAt.Format("02/01/2006 15:04 WIB")
		nooText := "Bukan"
		if s.IsNooToday {
			nooText = "NOO Hari Ini"
		}

		ecText := "Non-EC (Belum Order)"
		if s.IsEffectiveCall {
			if s.TodayOrderCount > 0 {
				ecText = fmt.Sprintf("EC (%d Pesanan)", s.TodayOrderCount)
			} else {
				ecText = "EC (Ada Transaksi)"
			}
		}

		rowStr := fmt.Sprintf("%d", rowNum)
		_ = f.SetCellValue(sheet, "A"+rowStr, i+1)
		_ = f.SetCellStyle(sheet, "A"+rowStr, "A"+rowStr, baseCenterStyle)
		_ = f.SetCellValue(sheet, "B"+rowStr, s.Name)
		_ = f.SetCellStyle(sheet, "B"+rowStr, "B"+rowStr, baseStyle)
		_ = f.SetCellValue(sheet, "C"+rowStr, custTypeLabel)
		_ = f.SetCellStyle(sheet, "C"+rowStr, "C"+rowStr, baseCenterStyle)
		_ = f.SetCellValue(sheet, "D"+rowStr, s.OwnerName)
		_ = f.SetCellStyle(sheet, "D"+rowStr, "D"+rowStr, baseStyle)
		_ = f.SetCellValue(sheet, "E"+rowStr, s.PhoneNumber)
		_ = f.SetCellStyle(sheet, "E"+rowStr, "E"+rowStr, baseStyle)
		_ = f.SetCellValue(sheet, "F"+rowStr, salesmanName)
		_ = f.SetCellStyle(sheet, "F"+rowStr, "F"+rowStr, baseStyle)
		_ = f.SetCellValue(sheet, "G"+rowStr, salesmanTitle)
		_ = f.SetCellStyle(sheet, "G"+rowStr, "G"+rowStr, baseStyle)
		_ = f.SetCellValue(sheet, "H"+rowStr, categoryText)
		_ = f.SetCellStyle(sheet, "H"+rowStr, "H"+rowStr, baseCenterStyle)
		_ = f.SetCellValue(sheet, "I"+rowStr, statusText)
		_ = f.SetCellStyle(sheet, "I"+rowStr, "I"+rowStr, baseCenterStyle)
		_ = f.SetCellValue(sheet, "J"+rowStr, s.Regency)
		_ = f.SetCellStyle(sheet, "J"+rowStr, "J"+rowStr, baseStyle)
		_ = f.SetCellValue(sheet, "K"+rowStr, s.District)
		_ = f.SetCellStyle(sheet, "K"+rowStr, "K"+rowStr, baseStyle)
		_ = f.SetCellValue(sheet, "L"+rowStr, s.Village)
		_ = f.SetCellStyle(sheet, "L"+rowStr, "L"+rowStr, baseStyle)
		_ = f.SetCellValue(sheet, "M"+rowStr, s.Address)
		_ = f.SetCellStyle(sheet, "M"+rowStr, "M"+rowStr, baseStyle)

		// Kolom Piutang — gunakan angka numerik agar Excel bisa sort/filter
		totalPiutang += s.TotalReceivable
		_ = f.SetCellValue(sheet, "N"+rowStr, s.TotalReceivable)
		if s.TotalReceivable > 0 {
			_ = f.SetCellStyle(sheet, "N"+rowStr, "N"+rowStr, piutangStyle)
		} else {
			_ = f.SetCellStyle(sheet, "N"+rowStr, "N"+rowStr, piutangZeroStyle)
		}

		// Kolom Nota Jatuh Tempo
		if s.OverdueCount > 0 {
			_ = f.SetCellValue(sheet, "O"+rowStr, fmt.Sprintf("%d nota", s.OverdueCount))
			_ = f.SetCellStyle(sheet, "O"+rowStr, "O"+rowStr, overdueStyle)
		} else {
			_ = f.SetCellValue(sheet, "O"+rowStr, "-")
			_ = f.SetCellStyle(sheet, "O"+rowStr, "O"+rowStr, baseCenterStyle)
		}

		_ = f.SetCellValue(sheet, "P"+rowStr, npwp)
		_ = f.SetCellStyle(sheet, "P"+rowStr, "P"+rowStr, baseStyle)

		// Kolom NIK — simpan sebagai teks agar leading zero tidak hilang
		if nik != "" {
			_ = f.SetCellValue(sheet, "Q"+rowStr, "'"+nik)
		}
		_ = f.SetCellStyle(sheet, "Q"+rowStr, "Q"+rowStr, baseStyle)

		// Waktu Registrasi (Tgl & Jam)
		_ = f.SetCellValue(sheet, "R"+rowStr, createdAt)
		_ = f.SetCellStyle(sheet, "R"+rowStr, "R"+rowStr, baseCenterStyle)

		// NOO Hari Ini
		_ = f.SetCellValue(sheet, "S"+rowStr, nooText)
		_ = f.SetCellStyle(sheet, "S"+rowStr, "S"+rowStr, baseCenterStyle)

		// Status EC Hari Ini
		_ = f.SetCellValue(sheet, "T"+rowStr, ecText)
		_ = f.SetCellStyle(sheet, "T"+rowStr, "T"+rowStr, baseCenterStyle)

		// Pesanan Hari Ini (Rp)
		_ = f.SetCellValue(sheet, "U"+rowStr, s.TodayOrderAmount)
		if s.TodayOrderAmount > 0 {
			_ = f.SetCellStyle(sheet, "U"+rowStr, "U"+rowStr, piutangZeroStyle)
		} else {
			_ = f.SetCellStyle(sheet, "U"+rowStr, "U"+rowStr, baseCenterStyle)
		}

		// Total Transaksi
		_ = f.SetCellValue(sheet, "V"+rowStr, s.TransactionCount)
		_ = f.SetCellStyle(sheet, "V"+rowStr, "V"+rowStr, baseCenterStyle)
	}

	// ── Baris Summary ─────────────────────────────────────────────────────────
	if len(stores) > 0 {
		summaryRow := len(stores) + 7 // 1 baris kosong setelah data terakhir
		_ = f.SetRowHeight(sheet, summaryRow, 22)

		_ = f.SetCellValue(sheet, fmt.Sprintf("A%d", summaryRow), "TOTAL")
		_ = f.MergeCell(sheet, fmt.Sprintf("A%d", summaryRow), fmt.Sprintf("M%d", summaryRow))
		_ = f.SetCellStyle(sheet, fmt.Sprintf("A%d", summaryRow), fmt.Sprintf("M%d", summaryRow), summaryLabelStyle)

		_ = f.SetCellValue(sheet, fmt.Sprintf("N%d", summaryRow), totalPiutang)
		_ = f.SetCellStyle(sheet, fmt.Sprintf("N%d", summaryRow), fmt.Sprintf("N%d", summaryRow), summaryValueStyle)

		// Total toko
		totalRow := summaryRow + 1
		_ = f.SetRowHeight(sheet, totalRow, 20)
		_ = f.SetCellValue(sheet, fmt.Sprintf("A%d", totalRow), fmt.Sprintf("Total Toko: %d", len(stores)))
		_ = f.MergeCell(sheet, fmt.Sprintf("A%d", totalRow), fmt.Sprintf("R%d", totalRow))
		_ = f.SetCellStyle(sheet, fmt.Sprintf("A%d", totalRow), fmt.Sprintf("R%d", totalRow), metaStyle)
	}

	// ── Sheet 2: Ringkasan per Salesman ───────────────────────────────────────
	const sheetRingkasan = "Ringkasan per Salesman"
	_, _ = f.NewSheet(sheetRingkasan)

	// Header ringkasan
	ringkasanHeaders := []struct {
		col   string
		label string
		width float64
	}{
		{"A", "No.", 5},
		{"B", "Nama Salesman", 28},
		{"C", "Jabatan", 22},
		{"D", "Jml. Toko Aktif", 16},
		{"E", "Jml. Toko Non-Aktif", 18},
		{"F", "Total Toko", 13},
		{"G", "Total Piutang (Rp)", 22},
	}

	_ = f.SetCellValue(sheetRingkasan, "A1", "RINGKASAN PER SALESMAN")
	_ = f.MergeCell(sheetRingkasan, "A1", "G1")
	_ = f.SetCellStyle(sheetRingkasan, "A1", "A1", titleStyle)
	_ = f.SetRowHeight(sheetRingkasan, 1, 28)

	_ = f.SetCellValue(sheetRingkasan, "A2", metaText)
	_ = f.MergeCell(sheetRingkasan, "A2", "G2")
	_ = f.SetCellStyle(sheetRingkasan, "A2", "A2", metaStyle)
	_ = f.SetRowHeight(sheetRingkasan, 2, 18)
	_ = f.SetRowHeight(sheetRingkasan, 3, 8) // spacer

	_ = f.SetRowHeight(sheetRingkasan, 4, 32)
	for _, h := range ringkasanHeaders {
		cell := h.col + "4"
		_ = f.SetCellValue(sheetRingkasan, cell, h.label)
		_ = f.SetCellStyle(sheetRingkasan, cell, cell, headerStyle)
		_ = f.SetColWidth(sheetRingkasan, h.col, h.col, h.width)
	}

	// Agregasi per salesman
	type salesmanSummary struct {
		Name     string
		Title    string
		Active   int
		Inactive int
		Piutang  float64
	}
	salesmanMap := make(map[string]*salesmanSummary)
	salesmanOrder := []string{}

	for _, s := range stores {
		key := "Belum Ditugaskan"
		name := "Belum Ditugaskan"
		title := "-"
		if s.AssignedEmployee != nil {
			key = s.AssignedEmployee.ID.String()
			name = strings.TrimSpace(s.AssignedEmployee.FirstName + " " + s.AssignedEmployee.LastName)
			if s.AssignedEmployee.JobPosition != nil {
				title = s.AssignedEmployee.JobPosition.Title
			}
		}

		if _, exists := salesmanMap[key]; !exists {
			salesmanMap[key] = &salesmanSummary{Name: name, Title: title}
			salesmanOrder = append(salesmanOrder, key)
		}
		if s.IsActive {
			salesmanMap[key].Active++
		} else {
			salesmanMap[key].Inactive++
		}
		salesmanMap[key].Piutang += s.TotalReceivable
	}

	ringkasanRow := 5
	for idx, key := range salesmanOrder {
		sm := salesmanMap[key]
		isEven := idx%2 == 1

		baseStyle := rowOddStyle
		baseCenterStyle := centerStyle
		if isEven {
			baseStyle = rowEvenStyle
			baseCenterStyle = centerEvenStyle
		}

		rowStr := fmt.Sprintf("%d", ringkasanRow)
		_ = f.SetRowHeight(sheetRingkasan, ringkasanRow, 20)
		_ = f.SetCellValue(sheetRingkasan, "A"+rowStr, idx+1)
		_ = f.SetCellStyle(sheetRingkasan, "A"+rowStr, "A"+rowStr, baseCenterStyle)
		_ = f.SetCellValue(sheetRingkasan, "B"+rowStr, sm.Name)
		_ = f.SetCellStyle(sheetRingkasan, "B"+rowStr, "B"+rowStr, baseStyle)
		_ = f.SetCellValue(sheetRingkasan, "C"+rowStr, sm.Title)
		_ = f.SetCellStyle(sheetRingkasan, "C"+rowStr, "C"+rowStr, baseStyle)
		_ = f.SetCellValue(sheetRingkasan, "D"+rowStr, sm.Active)
		_ = f.SetCellStyle(sheetRingkasan, "D"+rowStr, "D"+rowStr, baseCenterStyle)
		_ = f.SetCellValue(sheetRingkasan, "E"+rowStr, sm.Inactive)
		_ = f.SetCellStyle(sheetRingkasan, "E"+rowStr, "E"+rowStr, baseCenterStyle)
		_ = f.SetCellValue(sheetRingkasan, "F"+rowStr, sm.Active+sm.Inactive)
		_ = f.SetCellStyle(sheetRingkasan, "F"+rowStr, "F"+rowStr, baseCenterStyle)
		_ = f.SetCellValue(sheetRingkasan, "G"+rowStr, sm.Piutang)
		if sm.Piutang > 0 {
			_ = f.SetCellStyle(sheetRingkasan, "G"+rowStr, "G"+rowStr, piutangStyle)
		} else {
			_ = f.SetCellStyle(sheetRingkasan, "G"+rowStr, "G"+rowStr, piutangZeroStyle)
		}
		ringkasanRow++
	}

	// Summary baris akhir ringkasan
	if len(salesmanOrder) > 0 {
		_ = f.SetRowHeight(sheetRingkasan, ringkasanRow+1, 22)
		_ = f.SetCellValue(sheetRingkasan, fmt.Sprintf("A%d", ringkasanRow+1), "TOTAL")
		_ = f.MergeCell(sheetRingkasan, fmt.Sprintf("A%d", ringkasanRow+1), fmt.Sprintf("F%d", ringkasanRow+1))
		_ = f.SetCellStyle(sheetRingkasan, fmt.Sprintf("A%d", ringkasanRow+1), fmt.Sprintf("F%d", ringkasanRow+1), summaryLabelStyle)
		_ = f.SetCellValue(sheetRingkasan, fmt.Sprintf("G%d", ringkasanRow+1), totalPiutang)
		_ = f.SetCellStyle(sheetRingkasan, fmt.Sprintf("G%d", ringkasanRow+1), fmt.Sprintf("G%d", ringkasanRow+1), summaryValueStyle)
	}

	// Aktifkan sheet pertama
	idx, _ := f.GetSheetIndex(sheet)
	f.SetActiveSheet(idx)

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, fmt.Errorf("gagal menulis buffer excel: %v", err)
	}
	return buf.Bytes(), nil
}

type AuditStoreRequest struct {
	Status        string `json:"status" binding:"required,oneof=VERIFIED_REAL FICTITIOUS"`
	Notes         string `json:"notes" binding:"required"`
	PhotoProofURL string `json:"photo_proof_url"`
}

func (h *StoreHandler) AuditStore(c *gin.Context) {
	storeIDParam := c.Param("id")
	storeID, err := uuid.Parse(storeIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid store ID format"})
		return
	}

	var req AuditStoreRequest

	contentType := c.GetHeader("Content-Type")
	if strings.HasPrefix(contentType, "multipart/form-data") {
		req.Status = c.PostForm("status")
		req.Notes = c.PostForm("notes")

		photoURL, err := h.handleImageUpload(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload image: " + err.Error()})
			return
		}
		req.PhotoProofURL = photoURL

		if req.Status != "VERIFIED_REAL" && req.Status != "FICTITIOUS" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status"})
			return
		}
		if req.Notes == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Notes is required"})
			return
		}
	} else {
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	var auditorID uuid.UUID
	if userIDStr, exists := c.Get("userID"); exists {
		userID := userIDStr.(uuid.UUID)
		employee, err := h.employeeUseCase.GetEmployeeByUserID(userID)
		if err == nil && employee != nil {
			auditorID = employee.ID
		}
	}

	if auditorID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: auditor not found"})
		return
	}

	audit := &domain.StoreAudit{
		StoreID:       storeID,
		AuditorID:     auditorID,
		Status:        req.Status,
		Notes:         req.Notes,
		PhotoProofURL: req.PhotoProofURL,
	}

	if err := h.storeRepo.CreateAudit(audit); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save audit: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Store verification saved successfully",
		"data":    audit,
	})
}

// GET /api/v1/stores/:id/audits
func (h *StoreHandler) GetStoreAudits(c *gin.Context) {
	storeIDParam := c.Param("id")
	storeID, err := uuid.Parse(storeIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid store ID format"})
		return
	}

	audits, err := h.storeRepo.GetAuditsByStoreID(storeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch audits: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": audits})
}

// GET /api/v1/stores/audits/all?company_id=...&branch_id=...&page=1&limit=10
func (h *StoreHandler) GetAllAudits(c *gin.Context) {
	page := utils.StringToInt(c.Query("page"), 1)
	limit := utils.StringToInt(c.Query("limit"), 50)
	offset := (page - 1) * limit

	var companyID *uuid.UUID
	if cid := c.Query("company_id"); cid != "" {
		if parsed, err := uuid.Parse(cid); err == nil {
			companyID = &parsed
		}
	}

	var branchID *uuid.UUID
	if bid := c.Query("branch_id"); bid != "" && bid != "ALL_BRANCHES" {
		if parsed, err := uuid.Parse(bid); err == nil {
			branchID = &parsed
		}
	}

	audits, total, err := h.storeRepo.GetAllAudits(companyID, branchID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch all audits: " + err.Error()})
		return
	}

	totalPages := (int(total) + limit - 1) / limit

	c.JSON(http.StatusOK, gin.H{
		"data": audits,
		"meta": gin.H{
			"total":       total,
			"page":        page,
			"limit":       limit,
			"total_pages": totalPages,
		},
	})
}
