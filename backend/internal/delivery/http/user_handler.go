package http

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"example.com/hris/backend/internal/domain"
	"example.com/hris/backend/internal/repository"
	"example.com/hris/backend/pkg/redisclient"
	"example.com/hris/backend/pkg/utils"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserHandler struct {
	db           *gorm.DB
	employeeRepo repository.EmployeeRepository
}

func NewUserHandler(db *gorm.DB, employeeRepo repository.EmployeeRepository) *UserHandler {
	return &UserHandler{
		db:           db,
		employeeRepo: employeeRepo,
	}
}

func (h *UserHandler) SetupRoutes(router *gin.RouterGroup) {
	users := router.Group("/users")

	// Open to authenticated users initially
	users.POST("/face-register", h.RegisterFace)
	users.POST("/change-password", h.ChangePassword)
	users.POST("/fcm-token", h.SaveFcmToken)
	users.POST("/heartbeat", h.Heartbeat)

	// Admin only
	admin := users.Group("")
	admin.Use(RoleMiddleware(string(domain.RoleSuperAdmin), string(domain.RoleHRAdmin)))
	{
		admin.GET("", h.GetUsers)
		admin.GET("/active-users", h.GetActiveUsers)
		admin.GET("/:id", h.GetUserByID)
		admin.POST("", h.CreateUser)
		admin.PUT("/:id", h.UpdateUser)
		admin.DELETE("/:id", h.DeleteUser)
		admin.POST("/merge", h.MergeUsers)
		admin.GET("/:id/permissions", h.GetUserPermissions)
		admin.PUT("/:id/permissions", h.UpdateUserPermissions)
	}
}

func (h *UserHandler) GetUserPermissions(c *gin.Context) {
	idParam := c.Param("id")
	userID, err := uuid.Parse(idParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid ID format")
		return
	}

	var perms []string
	h.db.Model(&domain.UserPermission{}).
		Select("permissions.name").
		Joins("JOIN permissions ON permissions.id = user_permissions.permission_id").
		Where("user_permissions.user_id = ? AND user_permissions.is_granted = ?", userID, true).
		Pluck("permissions.name", &perms)

	utils.SuccessResponse(c, http.StatusOK, "Permissions fetched", gin.H{"permissions": perms})
}

type UpdatePermissionsRequest struct {
	Permissions []string `json:"permissions"`
}

func (h *UserHandler) UpdateUserPermissions(c *gin.Context) {
	idParam := c.Param("id")
	userID, err := uuid.Parse(idParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid ID format")
		return
	}

	var req UpdatePermissionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// For now, handle 'view_all_branches' specifically or generally
	var perm domain.Permission
	if err := h.db.Where("name = ?", "view_all_branches").First(&perm).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			perm = domain.Permission{
				Name:        "view_all_branches",
				Description: "Allows viewing of all branches globally",
			}
			if err := h.db.Create(&perm).Error; err != nil {
				utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create master permission 'view_all_branches'")
				return
			}
		} else {
			utils.ErrorResponse(c, http.StatusInternalServerError, "Database error finding master permission")
			return
		}
	}

	tx := h.db.Begin()

	// Remove old specific permission
	tx.Where("user_id = ? AND permission_id = ?", userID, perm.ID).Delete(&domain.UserPermission{})

	// Check if view_all_branches is in req
	hasGlobal := false
	for _, p := range req.Permissions {
		if p == "view_all_branches" {
			hasGlobal = true
			break
		}
	}

	if hasGlobal {
		tx.Create(&domain.UserPermission{
			UserID:       userID,
			PermissionID: perm.ID,
			IsGranted:    true,
		})
	}

	tx.Commit()

	utils.SuccessResponse(c, http.StatusOK, "Permissions updated successfully", nil)
}

func (h *UserHandler) GetUsers(c *gin.Context) {
	var users []domain.User
	if err := h.db.Preload("AccessRole").Preload("AccessRole.Menus").Omit("PasswordHash").Order("created_at DESC").Find(&users).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get users")
		return
	}

	var globalUserIDs []uuid.UUID
	h.db.Model(&domain.UserPermission{}).
		Joins("JOIN permissions ON permissions.id = user_permissions.permission_id").
		Where("permissions.name = ? AND user_permissions.is_granted = ?", "view_all_branches", true).
		Pluck("user_permissions.user_id", &globalUserIDs)

	globalMap := make(map[uuid.UUID]bool, len(globalUserIDs))
	for _, uid := range globalUserIDs {
		globalMap[uid] = true
	}

	type UserListItem struct {
		domain.User
		HasGlobalAccess bool `json:"has_global_access"`
	}

	result := make([]UserListItem, len(users))
	for i, u := range users {
		isGlobal := globalMap[u.ID] ||
			u.Role == domain.RoleSuperAdmin ||
			u.Role == domain.RoleCustomerSupport ||
			u.Role == domain.RoleAuditor ||
			u.Role == domain.RoleAdminDesigner ||
			u.Role == domain.RoleValidatorDesain ||
			u.Role == domain.RoleValidatorToko
		result[i] = UserListItem{
			User:            u,
			HasGlobalAccess: isGlobal,
		}
	}

	utils.SuccessResponse(c, http.StatusOK, "Users fetched successfully", result)
}

// CreateUserRequest defines the request body for creating a user
type CreateUserRequest struct {
	Email            string      `json:"email" binding:"required,email"`
	Password         string      `json:"password" binding:"required"`
	Role             domain.Role `json:"role" binding:"required"`
	IsActive         bool        `json:"isActive"`
	CanAbsenAnywhere bool        `json:"canAbsenAnywhere"`
	AccessRoleID     *uuid.UUID  `json:"access_role_id"`
}

func (h *UserHandler) CreateUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	cleanEmail := strings.ToLower(strings.TrimSpace(req.Email))
	user := domain.User{
		Email:            cleanEmail,
		Role:             req.Role,
		IsActive:         req.IsActive,
		CanAbsenAnywhere: req.CanAbsenAnywhere,
		AccessRoleID:     req.AccessRoleID,
	}

	cleanPassword := strings.TrimSpace(req.Password)
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(cleanPassword), bcrypt.DefaultCost)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to hash password")
		return
	}
	user.PasswordHash = string(hashedPassword)

	if err := h.db.Create(&user).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create user")
		return
	}

	user.PasswordHash = "" // hide
	utils.SuccessResponse(c, http.StatusCreated, "User created successfully", gin.H{"data": user})
}

func (h *UserHandler) GetUserByID(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid ID format")
		return
	}

	var user domain.User
	if err := h.db.Preload("AccessRole").Preload("AccessRole.Menus").Omit("PasswordHash").Where("id = ?", id).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "User not found")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch user")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "User fetched successfully", gin.H{"data": user})
}

// UpdateUserRequest defines the request body for updating a user
type UpdateUserRequest struct {
	Email            string      `json:"email" binding:"required,email"`
	Password         string      `json:"password"` // optional
	Role             domain.Role `json:"role" binding:"required"`
	IsActive         bool        `json:"isActive"`
	CanAbsenAnywhere bool        `json:"canAbsenAnywhere"`
	AccessRoleID     *uuid.UUID  `json:"access_role_id"`
}

func (h *UserHandler) UpdateUser(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid ID format")
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	var user domain.User
	if err := h.db.Where("id = ?", id).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "User not found")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch user")
		return
	}

	user.Email = strings.ToLower(strings.TrimSpace(req.Email))
	user.Role = req.Role
	user.IsActive = req.IsActive
	user.CanAbsenAnywhere = req.CanAbsenAnywhere
	user.AccessRoleID = req.AccessRoleID

	cleanPassword := strings.TrimSpace(req.Password)
	if cleanPassword != "" {
		hashed, err := bcrypt.GenerateFromPassword([]byte(cleanPassword), bcrypt.DefaultCost)
		if err != nil {
			utils.ErrorResponse(c, http.StatusInternalServerError, "failed to hash password")
			return
		}
		user.PasswordHash = string(hashed)
	}

	if err := h.db.Save(&user).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update user")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "User updated successfully", gin.H{"data": nil})
}

func (h *UserHandler) DeleteUser(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid ID format")
		return
	}

	var empID uuid.UUID
	err = h.db.Transaction(func(tx *gorm.DB) error {
		// 1. Find the employee associated with this user
		var employee domain.Employee
		if err := tx.Where("user_id = ?", id).First(&employee).Error; err == nil {
			empID = employee.ID
			// Found employee, use the repository's robust delete method
			if err := h.employeeRepo.Delete(employee.ID); err != nil {
				return err
			}
		}

		// 2. Finally delete the user
		if err := tx.Where("id = ?", id).Delete(&domain.User{}).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		fmt.Printf("Error deleting user %s: %v\n", id, err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete user and associated data")
		return
	}

	if redisclient.Client != nil {
		ctx := context.Background()
		if empID != uuid.Nil {
			_ = redisclient.Client.Del(ctx, fmt.Sprintf("employee:id:%s", empID.String())).Err()
		}
		_ = redisclient.Client.Del(ctx, fmt.Sprintf("employee:userid:%s", id.String())).Err()
	}

	utils.SuccessResponse(c, http.StatusOK, "User and all associated data deleted successfully", nil)
}

type RegisterFaceRequest struct {
	Embedding domain.FloatArray `json:"embedding" binding:"required"`
	Selfie    string            `json:"selfie" binding:"required"` // Base64 string
}

func (h *UserHandler) RegisterFace(c *gin.Context) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized")
		return
	}
	userID := userIDStr.(uuid.UUID)

	var req RegisterFaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	employee, err := h.employeeRepo.FindByUserID(userID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get employee record")
		return
	}
	if employee == nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Employee record not found")
		return
	}

	// Process Selfie Upload
	uploadDir := "uploads/faces"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create upload directory")
		return
	}

	imageData, err := base64.StdEncoding.DecodeString(req.Selfie)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid selfie image data")
		return
	}

	compressedData, err := utils.CompressImage(imageData, 800, 75)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to compress image")
		return
	}

	filename := fmt.Sprintf("%s_%d.jpg", employee.ID.String(), time.Now().Unix())
	filePath := filepath.Join(uploadDir, filename)

	if err := os.WriteFile(filePath, compressedData, 0644); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to save image")
		return
	}

	employee.FaceEmbedding = req.Embedding
	employee.FaceReferenceURL = "/" + filePath
	employee.IsFaceRegistered = true

	if err := h.employeeRepo.Update(employee); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to save face embedding")
		return
	}

	if redisclient.Client != nil {
		ctx := context.Background()
		_ = redisclient.Client.Del(ctx, fmt.Sprintf("employee:id:%s", employee.ID.String())).Err()
		_ = redisclient.Client.Del(ctx, fmt.Sprintf("employee:userid:%s", userID.String())).Err()
	}

	utils.SuccessResponse(c, http.StatusOK, "Face registered successfully", nil)
}

type ChangePasswordRequest struct {
	NewPassword string `json:"newPassword" binding:"required,min=6"`
}

func (h *UserHandler) ChangePassword(c *gin.Context) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized")
		return
	}
	userID := userIDStr.(uuid.UUID)

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	var user domain.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to find user")
		return
	}

	// Hash new password
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to hash new password")
		return
	}

	if err := h.db.Model(&user).Update("password_hash", string(hashed)).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update password")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Password berhasil diperbarui", nil)
}

type MergeUsersRequest struct {
	DuplicateInput string `json:"duplicate_input" binding:"required"`
	TargetInput    string `json:"target_input" binding:"required"`
}

func (h *UserHandler) MergeUsers(c *gin.Context) {
	var req MergeUsersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	dupInput := req.DuplicateInput
	targetInput := req.TargetInput

	// Find duplicate and target users
	findUser := func(input string) (uuid.UUID, string, error) {
		var u struct {
			ID    uuid.UUID
			Email string
		}
		var err error
		if _, uuidErr := uuid.Parse(input); uuidErr == nil {
			err = h.db.Table("users").Where("id = ?", input).Take(&u).Error
		} else {
			// Try exact match first to handle duplicate emails with different casings
			err = h.db.Table("users").Where("email = ?", input).Take(&u).Error
			if err != nil {
				// Fallback to case-insensitive
				err = h.db.Table("users").Where("LOWER(email) = ?", strings.ToLower(input)).Take(&u).Error
			}
		}
		return u.ID, u.Email, err
	}

	dupUserID, dupEmail, err := findUser(dupInput)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, fmt.Sprintf("Duplicate user '%s' not found", dupInput))
		return
	}

	targetUserID, targetEmail, err := findUser(targetInput)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, fmt.Sprintf("Target user '%s' not found", targetInput))
		return
	}

	if dupUserID == targetUserID {
		utils.ErrorResponse(c, http.StatusBadRequest, "Duplicate user and Target user are the same")
		return
	}

	// Fetch employees
	var dupEmpIDStr, targetEmpIDStr string
	err = h.db.Table("employees").Select("id").Where("user_id = ?", dupUserID).Take(&dupEmpIDStr).Error
	if err != nil {
		// Non-fatal if duplicate has no employee (just a user without profile)
	}
	err = h.db.Table("employees").Select("id").Where("user_id = ?", targetUserID).Take(&targetEmpIDStr).Error
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Target user has no employee record")
		return
	}

	dupEmpID, _ := uuid.Parse(dupEmpIDStr)
	targetEmpID, _ := uuid.Parse(targetEmpIDStr)

	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if tx.Error != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to start database transaction")
		return
	}

	if dupEmpID != uuid.Nil && targetEmpID != uuid.Nil {
		// Update simple foreign keys referencing employees(id)
		updates := []struct {
			query string
			args  []interface{}
		}{
			{"UPDATE stores SET assigned_employee_id = ? WHERE assigned_employee_id = ?", []interface{}{targetEmpID, dupEmpID}},
			{"UPDATE attendance_logs SET employee_id = ? WHERE employee_id = ?", []interface{}{targetEmpID, dupEmpID}},
			{"UPDATE sales_visits SET employee_id = ? WHERE employee_id = ?", []interface{}{targetEmpID, dupEmpID}},
			{"UPDATE sales_transactions SET employee_id = ? WHERE employee_id = ?", []interface{}{targetEmpID, dupEmpID}},
			{"UPDATE sales_orders SET employee_id = ? WHERE employee_id = ?", []interface{}{targetEmpID, dupEmpID}},
			{"UPDATE sales_orders SET confirmed_by_id = ? WHERE confirmed_by_id = ?", []interface{}{targetEmpID, dupEmpID}},
			{"UPDATE sales_payments SET employee_id = ? WHERE employee_id = ?", []interface{}{targetEmpID, dupEmpID}},
			{"UPDATE sales_returns SET employee_id = ? WHERE employee_id = ?", []interface{}{targetEmpID, dupEmpID}},
			{"UPDATE sales_visit_plans SET employee_id = ? WHERE employee_id = ?", []interface{}{targetEmpID, dupEmpID}},
			{"UPDATE leave_requests SET employee_id = ? WHERE employee_id = ?", []interface{}{targetEmpID, dupEmpID}},
			{"UPDATE leave_requests SET approved_by = ? WHERE approved_by = ?", []interface{}{targetEmpID, dupEmpID}},
			{"UPDATE employee_salary_settings SET employee_id = ? WHERE employee_id = ?", []interface{}{targetEmpID, dupEmpID}},
			{"UPDATE payslips SET employee_id = ? WHERE employee_id = ?", []interface{}{targetEmpID, dupEmpID}},
			{"UPDATE overtimes SET employee_id = ? WHERE employee_id = ?", []interface{}{targetEmpID, dupEmpID}},
			{"UPDATE overtimes SET approved_by = ? WHERE approved_by = ?", []interface{}{targetEmpID, dupEmpID}},
			{"UPDATE reimbursements SET employee_id = ? WHERE employee_id = ?", []interface{}{targetEmpID, dupEmpID}},
			{"UPDATE reimbursements SET approved_by = ? WHERE approved_by = ?", []interface{}{targetEmpID, dupEmpID}},
			{"UPDATE banner_orders SET employee_id = ? WHERE employee_id = ?", []interface{}{targetEmpID, dupEmpID}},
			{"UPDATE delivery_batches SET driver_id = ? WHERE driver_id = ?", []interface{}{targetEmpID, dupEmpID}},
			{"UPDATE production_logs SET employee_id = ? WHERE employee_id = ?", []interface{}{targetEmpID, dupEmpID}},
			{"UPDATE employees SET manager_id = ? WHERE manager_id = ?", []interface{}{targetEmpID, dupEmpID}},
		}

		for _, u := range updates {
			if err := tx.Exec(u.query, u.args...).Error; err != nil {
				tx.Rollback()
				utils.ErrorResponse(c, http.StatusInternalServerError, fmt.Sprintf("Failed to update references: %v", err))
				return
			}
		}

		// Handle Unique Constraints
		// employee_shifts
		if err := tx.Exec(`DELETE FROM employee_shifts 
			WHERE employee_id = ? 
			AND date IN (SELECT date FROM employee_shifts WHERE employee_id = ?)`,
			dupEmpID, targetEmpID).Error; err != nil {
			tx.Rollback()
			utils.ErrorResponse(c, http.StatusInternalServerError, fmt.Sprintf("Failed to resolve shift conflicts: %v", err))
			return
		}
		if err := tx.Exec("UPDATE employee_shifts SET employee_id = ? WHERE employee_id = ?", targetEmpID, dupEmpID).Error; err != nil {
			tx.Rollback()
			utils.ErrorResponse(c, http.StatusInternalServerError, fmt.Sprintf("Failed to update shift references: %v", err))
			return
		}

		// leave_balances
		if err := tx.Exec(`DELETE FROM leave_balances 
			WHERE employee_id = ? 
			AND (leave_type_id, year) IN (SELECT leave_type_id, year FROM leave_balances WHERE employee_id = ?)`,
			dupEmpID, targetEmpID).Error; err != nil {
			tx.Rollback()
			utils.ErrorResponse(c, http.StatusInternalServerError, fmt.Sprintf("Failed to resolve leave balance conflicts: %v", err))
			return
		}
		if err := tx.Exec("UPDATE leave_balances SET employee_id = ? WHERE employee_id = ?", targetEmpID, dupEmpID).Error; err != nil {
			tx.Rollback()
			utils.ErrorResponse(c, http.StatusInternalServerError, fmt.Sprintf("Failed to update leave balance references: %v", err))
			return
		}

		// sales_stocks
		if err := tx.Exec(`UPDATE sales_stocks t 
			SET quantity = t.quantity + d.quantity 
			FROM sales_stocks d 
			WHERE t.employee_id = ? AND d.employee_id = ? AND t.product_id = d.product_id`,
			targetEmpID, dupEmpID).Error; err != nil {
			tx.Rollback()
			utils.ErrorResponse(c, http.StatusInternalServerError, fmt.Sprintf("Failed to merge sales stock: %v", err))
			return
		}
		if err := tx.Exec(`DELETE FROM sales_stocks 
			WHERE employee_id = ? 
			AND product_id IN (SELECT product_id FROM sales_stocks WHERE employee_id = ?)`,
			dupEmpID, targetEmpID).Error; err != nil {
			tx.Rollback()
			utils.ErrorResponse(c, http.StatusInternalServerError, fmt.Sprintf("Failed to clean up sales stock: %v", err))
			return
		}
		if err := tx.Exec("UPDATE sales_stocks SET employee_id = ? WHERE employee_id = ?", targetEmpID, dupEmpID).Error; err != nil {
			tx.Rollback()
			utils.ErrorResponse(c, http.StatusInternalServerError, fmt.Sprintf("Failed to update sales stock references: %v", err))
			return
		}
	}

	// Update notifications
	if err := tx.Exec("UPDATE notifications SET user_id = ? WHERE user_id = ?", targetUserID, dupUserID).Error; err != nil {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusInternalServerError, fmt.Sprintf("Failed to update notifications: %v", err))
		return
	}

	// Delete duplicate employee & user records
	if dupEmpID != uuid.Nil {
		if err := tx.Exec("DELETE FROM employees WHERE id = ?", dupEmpID).Error; err != nil {
			tx.Rollback()
			utils.ErrorResponse(c, http.StatusInternalServerError, fmt.Sprintf("Failed to delete duplicate employee: %v", err))
			return
		}
	}
	if err := tx.Exec("DELETE FROM users WHERE id = ?", dupUserID).Error; err != nil {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusInternalServerError, fmt.Sprintf("Failed to delete duplicate user: %v", err))
		return
	}

	if err := tx.Commit().Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, fmt.Sprintf("Failed to commit merge: %v", err))
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Users merged successfully", gin.H{
		"duplicate_user_email": dupEmail,
		"target_user_email":    targetEmail,
	})
}

type SaveFcmTokenRequest struct {
	FCMToken   string `json:"fcm_token" binding:"required"`
	DeviceInfo string `json:"device_info"`
}

func (h *UserHandler) SaveFcmToken(c *gin.Context) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized")
		return
	}
	userID := userIDStr.(uuid.UUID)

	var req SaveFcmTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request payload")
		return
	}

	token := domain.UserFcmToken{
		UserID:     userID,
		FCMToken:   req.FCMToken,
		DeviceInfo: req.DeviceInfo,
		UpdatedAt:  time.Now(),
	}

	// Upsert: check if token already exists for the user
	var existing domain.UserFcmToken
	err := h.db.Where("user_id = ?", userID).First(&existing).Error
	if err == nil {
		// Update existing
		existing.FCMToken = req.FCMToken
		existing.DeviceInfo = req.DeviceInfo
		existing.UpdatedAt = time.Now()
		if err := h.db.Save(&existing).Error; err != nil {
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to save FCM token")
			return
		}
	} else if err == gorm.ErrRecordNotFound {
		// Check if the same token exists for another user and delete it to prevent duplicate deliveries
		h.db.Where("fcm_token = ?", req.FCMToken).Delete(&domain.UserFcmToken{})

		// Create new
		token.ID = uuid.New()
		token.CreatedAt = time.Now()
		if err := h.db.Create(&token).Error; err != nil {
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create FCM token")
			return
		}
	} else {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Database error")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "FCM token saved successfully",
	})
}

func (h *UserHandler) Heartbeat(c *gin.Context) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized")
		return
	}
	userID := userIDStr.(uuid.UUID)

	companyIDStr, _ := c.Get("companyID")
	var companyID string
	if cid, ok := companyIDStr.(uuid.UUID); ok {
		companyID = cid.String()
	} else if cidStr, ok := companyIDStr.(string); ok {
		companyID = cidStr
	} else {
		companyID = "default"
	}

	roleStr, _ := c.Get("role")
	var role string
	if r, ok := roleStr.(string); ok {
		role = r
	} else if r, ok := roleStr.(domain.Role); ok {
		role = string(r)
	}

	var empName string
	employee, err := h.employeeRepo.FindByUserID(userID)
	if err == nil && employee != nil {
		empName = strings.TrimSpace(employee.FirstName + " " + employee.LastName)
	} else {
		// fallback to user email if employee not found
		var u domain.User
		if err := h.db.Where("id = ?", userID).First(&u).Error; err == nil {
			empName = u.Email
		}
	}

	if redisclient.Client != nil {
		ctx := context.Background()
		now := time.Now().Unix()

		// Add to sorted set for tracking online users (Global)
		redisclient.Client.ZAdd(ctx, "system:active_users", redis.Z{
			Score:  float64(now),
			Member: userID.String(),
		})

		// Save user details for quick retrieval
		userInfoMap := map[string]interface{}{
			"id":          userID.String(),
			"name":        empName,
			"role":        role,
			"company_id":  companyID,
			"last_active": now,
		}
		userInfoBytes, _ := json.Marshal(userInfoMap)
		redisclient.Client.SetEx(ctx, fmt.Sprintf("user:info:%s", userID.String()), string(userInfoBytes), 7*24*time.Hour)
	}

	c.JSON(http.StatusOK, gin.H{"message": "heartbeat received"})
}

func (h *UserHandler) GetActiveUsers(c *gin.Context) {
	if redisclient.Client == nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Redis not configured")
		return
	}

	companyIDStr, _ := c.Get("companyID")
	var companyID string
	if cid, ok := companyIDStr.(uuid.UUID); ok {
		companyID = cid.String()
	} else if cidStr, ok := companyIDStr.(string); ok {
		companyID = cidStr
	} else {
		companyID = "default"
	}

	roleStr, _ := c.Get("role")
	var callerRole string
	if r, ok := roleStr.(string); ok {
		callerRole = r
	} else if r, ok := roleStr.(domain.Role); ok {
		callerRole = string(r)
	}

	ctx := context.Background()
	cutoff := time.Now().Add(-7 * 24 * time.Hour).Unix()
	zsetKey := "system:active_users"

	// Clean up stale users
	redisclient.Client.ZRemRangeByScore(ctx, zsetKey, "-inf", fmt.Sprintf("%d", cutoff))

	// Get active user IDs
	userIds, err := redisclient.Client.ZRevRange(ctx, zsetKey, 0, -1).Result()
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get active users")
		return
	}

	var activeUsers []map[string]interface{}
	if len(userIds) > 0 {
		var keys []string
		for _, id := range userIds {
			keys = append(keys, fmt.Sprintf("user:info:%s", id))
		}

		// Use MGet to fetch all user infos at once
		infos, err := redisclient.Client.MGet(ctx, keys...).Result()
		if err == nil {
			for i, info := range infos {
				if infoStr, ok := info.(string); ok {
					var u map[string]interface{}
					if err := json.Unmarshal([]byte(infoStr), &u); err == nil {
						userCompanyID, _ := u["company_id"].(string)
						// Filter: Superadmin sees all, otherwise only see users from same company
						if callerRole == "superadmin" || userCompanyID == companyID || userCompanyID == "" {
							activeUsers = append(activeUsers, u)
						}
					} else {
						log.Printf("[ActiveUsers] Failed to unmarshal info for user %s: %v", userIds[i], err)
					}
				} else {
					log.Printf("[ActiveUsers] Info for user %s is missing or not a string", userIds[i])
				}
			}
		} else {
			log.Printf("[ActiveUsers] MGet error: %v", err)
		}
	} else {
		log.Printf("[ActiveUsers] No userIds found in zset %s", zsetKey)
	}

	if activeUsers == nil {
		activeUsers = []map[string]interface{}{}
	}
	log.Printf("[ActiveUsers] Returning %d active users", len(activeUsers))

	utils.SuccessResponse(c, http.StatusOK, "Active users fetched", activeUsers)
}
