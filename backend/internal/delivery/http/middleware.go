package http

import (
	"compress/gzip"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"example.com/hris/backend/internal/usecase"
	"example.com/hris/backend/pkg/utils"
	"gorm.io/gorm"
)

func AuthMiddleware(secret string, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		token := ""

		if authHeader != "" {
			parts := strings.Split(authHeader, " ")
			if len(parts) == 2 && parts[0] == "Bearer" {
				token = parts[1]
			}
		}

		// Fallback to query parameter for window.open/exports
		if token == "" {
			token = c.Query("token")
		}

		if token == "" {
			utils.ErrorResponse(c, http.StatusUnauthorized, "Authorization token required")
			c.Abort()
			return
		}

		claims, err := utils.ValidateToken(token, secret)
		if err != nil {
			// Log the actual error internally for debugging
			log.Printf("Token validation failed: %v", err)
			utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid or expired token")
			c.Abort()
			return
		}

		// Jika token berasal dari client mobile (memiliki claim MobileSessionID), pastikan masih aktif di database
		if claims.MobileSessionID != "" && db != nil {
			var currentSession *string
			err := db.Model(&domain.User{}).
				Select("mobile_session_id").
				Where("id = ?", claims.UserID).
				Scan(&currentSession).Error
			if err != nil || currentSession == nil || *currentSession != claims.MobileSessionID {
				c.JSON(http.StatusUnauthorized, gin.H{
					"status":     false,
					"error_code": "DUPLICATE_LOGIN",
					"message":    "Akun Anda telah login di perangkat lain.",
				})
				c.Abort()
				return
			}
		}

		c.Set("userID", claims.UserID)
		c.Set("user_id", claims.UserID.String()) // Compatibility
		c.Set("companyID", claims.CompanyID)
		c.Set("company_id", claims.CompanyID.String()) // Compatibility

		// Role changes must take effect immediately. The role claim is a snapshot
		// from the time the JWT was issued and can be stale after an admin changes
		// the user's role. Read the current role from the database for every
		// authenticated request so permissions cannot depend on an old token.
		role := claims.Role
		if db != nil {
			var currentUser domain.User
			// Try querying with distributor_store_id; fallback if column does not exist yet
			if err := db.Select("role, email, is_active, distributor_store_id").Where("id = ?", claims.UserID).First(&currentUser).Error; err != nil {
				if errFallback := db.Select("role, email, is_active").Where("id = ?", claims.UserID).First(&currentUser).Error; errFallback != nil || !currentUser.IsActive {
					utils.ErrorResponse(c, http.StatusUnauthorized, "User tidak aktif atau tidak ditemukan")
					c.Abort()
					return
				}
			} else if !currentUser.IsActive {
				utils.ErrorResponse(c, http.StatusUnauthorized, "User tidak aktif atau tidak ditemukan")
				c.Abort()
				return
			}
			role = string(currentUser.Role)
			c.Set("userEmail", currentUser.Email)
			c.Set("email", currentUser.Email)
			if currentUser.DistributorStoreID != nil {
				c.Set("distributor_store_id", currentUser.DistributorStoreID.String())
				c.Set("distributorStoreID", *currentUser.DistributorStoreID)
			}
		}
		c.Set("role", role)
		c.Next()
	}
}

func EmployeeMiddleware(employeeUseCase usecase.EmployeeUsecase) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDStr, exists := c.Get("user_id")
		if !exists {
			c.Next()
			return
		}

		userID := uuid.MustParse(userIDStr.(string))
		employee, err := employeeUseCase.GetEmployeeByUserID(userID)
		if err == nil && employee != nil {
			c.Set("employeeID", employee.ID)
			c.Set("employee_id", employee.ID.String())

			name := employee.FirstName
			if employee.LastName != "" {
				name += " " + employee.LastName
			}
			c.Set("name", name)

			if employee.BranchID != nil {
				c.Set("branchID", *employee.BranchID)
				c.Set("branch_id", employee.BranchID.String())
			}
			if employee.Branch != nil {
				c.Set("branch_name", employee.Branch.Name)
				isBooster := strings.Contains(strings.ToLower(employee.Branch.Name), "booster")
				c.Set("is_booster_branch", isBooster)
			}
			if employee.CompanyID != nil {
				c.Set("companyID", *employee.CompanyID)
				c.Set("company_id", employee.CompanyID.String())
			}

			// Dynamic role promotion for HR/Manager employees
			if roleVal, ok := c.Get("role"); ok && strings.ToLower(roleVal.(string)) == string(domain.RoleEmployee) {
				if employee.JobPosition != nil {
					title := strings.ToLower(employee.JobPosition.Title)
					if strings.Contains(title, "hr") || strings.Contains(title, "human resource") || strings.Contains(title, "recruiter") || strings.Contains(title, "personalia") || strings.Contains(title, "admin") {
						c.Set("role", string(domain.RoleHRAdmin))
					} else if strings.Contains(title, "manager") || strings.Contains(title, "supervisor") || strings.Contains(title, "kepala") {
						c.Set("role", string(domain.RoleManager))
					}
				}
			}
		}

		c.Next()
	}
}

func RoleMiddleware(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("role")
		if !exists {
			utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized")
			c.Abort()
			return
		}

		userRoleStr, _ := userRole.(string)

		// SuperAdmin always has full system access
		if userRoleStr == string(domain.RoleSuperAdmin) {
			c.Next()
			return
		}

		// Auditor and Customer Support roles can access any GET request for read-only access
		if (userRoleStr == string(domain.RoleAuditor) || userRoleStr == string(domain.RoleCustomerSupport)) && c.Request.Method == "GET" {
			c.Next()
			return
		}

		valid := false
		for _, role := range roles {
			if role == userRoleStr {
				valid = true
				break
			}

			// supervisor_sales and supervisor_sales_mt cross-compatibility
			if (userRoleStr == string(domain.RoleSupervisorSales) || userRoleStr == string(domain.RoleSupervisorSalesMT) || userRoleStr == "supervisor" || userRoleStr == "spv") &&
				(role == string(domain.RoleSupervisorSales) || role == string(domain.RoleSupervisorSalesMT) || role == "supervisor" || role == "spv") {
				valid = true
				break
			}

			// Manager and Director/Owner inherit all supervisor and admin department permissions
			if (userRoleStr == string(domain.RoleManager) || userRoleStr == "direktur" || userRoleStr == "director" || userRoleStr == "owner") &&
				(role == string(domain.RoleSupervisorSales) || role == string(domain.RoleSupervisorSalesMT) || role == string(domain.RoleSupervisorGudang) || role == string(domain.RoleAdminNota) || role == string(domain.RoleAdminGudang)) {
				valid = true
				break
			}

			// admin_operasional inherits permissions for gudang, nota, and kasir
			if userRoleStr == string(domain.RoleAdminOperasional) {
				if role == string(domain.RoleAdminGudang) || role == string(domain.RoleAdminNota) || role == string(domain.RoleAdminKasir) || role == string(domain.RoleCashier) {
					valid = true
					break
				}
			}

			// Auditor on Booster Branch inherits operational permissions for gudang, nota, and kasir
			if userRoleStr == string(domain.RoleAuditor) {
				if isBoosterVal, ok := c.Get("is_booster_branch"); ok && isBoosterVal.(bool) {
					if role == string(domain.RoleAdminGudang) || role == string(domain.RoleAdminNota) || role == string(domain.RoleAdminKasir) || role == string(domain.RoleCashier) {
						valid = true
						break
					}
				}
			}

			// supervisor_gudang inherits permissions for gudang and pabrik
			if userRoleStr == string(domain.RoleSupervisorGudang) {
				if role == string(domain.RoleAdminGudang) || role == string(domain.RoleAdminPabrik) {
					valid = true
					break
				}
			}

			// admin_pabrik inherits permissions for gudang
			if userRoleStr == string(domain.RoleAdminPabrik) {
				if role == string(domain.RoleAdminGudang) {
					valid = true
					break
				}
			}

			// cashier_pusat inherits permissions for cashier
			if userRoleStr == string(domain.RoleCashierPusat) {
				if role == string(domain.RoleCashier) {
					valid = true
					break
				}
			}

			// admin_purchasing inherits permissions for cashier, cashier_pusat, pabrik, and gudang
			if userRoleStr == string(domain.RoleAdminPurchasing) {
				if role == string(domain.RoleCashier) || role == string(domain.RoleCashierPusat) || role == string(domain.RoleAdminPabrik) || role == string(domain.RoleAdminGudang) {
					valid = true
					break
				}
			}
		}

		if !valid {
			utils.ErrorResponse(c, http.StatusForbidden, "Forbidden: insufficient permissions")
			c.Abort()
			return
		}

		c.Next()
	}
}

// DirectorRoleMiddleware also allows users whose job position identifies them
// as a director. This is needed because the user role and the employee job
// position are stored separately.
func DirectorRoleMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, _ := c.Get("role")
		roleString, _ := role.(string)
		if roleString == string(domain.RoleSuperAdmin) || roleString == "owner" ||
			roleString == "direktur" || roleString == string(domain.RoleManager) {
			c.Next()
			return
		}

		userID, exists := c.Get("userID")
		if !exists || db == nil {
			utils.ErrorResponse(c, http.StatusForbidden, "Forbidden")
			c.Abort()
			return
		}

		var count int64
		db.Table("employees").
			Joins("JOIN job_positions ON job_positions.id = employees.job_position_id").
			Where("employees.user_id = ?", userID).
			Where("LOWER(job_positions.title) LIKE ? OR LOWER(job_positions.title) LIKE ?", "%direktur%", "%director%").
			Count(&count)
		if count > 0 {
			c.Set("role", "direktur")
			c.Next()
			return
		}

		utils.ErrorResponse(c, http.StatusForbidden, "Forbidden")
		c.Abort()
	}
}

func BranchFilterMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			c.Next()
			return
		}

		roleStr := role.(string)

		isBooster := false
		if isBoosterVal, ok := c.Get("is_booster_branch"); ok && isBoosterVal.(bool) {
			isBooster = true
		}

		// SuperAdmin, Customer Support, Non-Booster Auditor, Admin Designer, Validator Desain, Validator Toko, and Admin Purchasing selalu bisa melihat semua
		if roleStr == string(domain.RoleSuperAdmin) || roleStr == string(domain.RoleCustomerSupport) || (roleStr == string(domain.RoleAuditor) && !isBooster) || roleStr == string(domain.RoleAdminDesigner) || roleStr == string(domain.RoleValidatorDesain) || roleStr == string(domain.RoleValidatorToko) || roleStr == string(domain.RoleAdminPurchasing) {
			c.Next()
			return
		}

		// Antrian tugas pemasangan spanduk & jadwal tugas operator adalah daftar kerja lintas cabang:
		// Operator lapangan harus bisa melihat semua cabang pada tab Semua Wilayah dan spanduk yang ditugaskan kepada mereka.
		// Filter cabang eksplisit tetap ditangani oleh query parameter jika disediakan oleh klien.
		if c.Request.Method == http.MethodGet &&
			strings.HasSuffix(c.FullPath(), "/branding/banners") {
			c.Next()
			return
		}

		hasGlobalAccess := false

		// Cek permission spesifik 'view_all_branches' untuk HR/User lain
		userIDStr, existsUser := c.Get("user_id")
		if existsUser && userIDStr != "" {
			var count int64
			db.Model(&domain.UserPermission{}).
				Joins("JOIN permissions ON permissions.id = user_permissions.permission_id").
				Where("user_permissions.user_id = ? AND permissions.name = ? AND user_permissions.is_granted = ?", userIDStr, "view_all_branches", true).
				Count(&count)
			if count > 0 {
				hasGlobalAccess = true
			}
		}

		if !hasGlobalAccess {
			// Allow bypassing branch filter if explicit query parameter is set (hanya jika memang dia sudah di-filter di level selanjutnya/memang dibolehkan)
			// Namun karena ini middleware pembatasan, jika tidak punya global access, HARUS dibatasi ke branch_id dia
			branchIDVal, existsBranch := c.Get("branch_id")
			if existsBranch && branchIDVal != "" {
				branchIDStr := branchIDVal.(string)
				// Override query parameter "branch_id"
				q := c.Request.URL.Query()
				q.Set("branch_id", branchIDStr)
				c.Request.URL.RawQuery = q.Encode()
			} else {
				// Fallback jika employee tidak punya branch (misal kantor pusat/super), tapi ini jarang jika bukan global
			}
		}
		c.Next()
	}
}

type gzipWriter struct {
	gin.ResponseWriter
	writer io.Writer
}

func (g *gzipWriter) Write(data []byte) (int, error) {
	return g.writer.Write(data)
}

func (g *gzipWriter) WriteString(s string) (int, error) {
	return g.writer.Write([]byte(s))
}

// gzipWriterPool menghindari alokasi gzip.Writer baru setiap request
var gzipWriterPool = sync.Pool{
	New: func() interface{} {
		gz, _ := gzip.NewWriterLevel(io.Discard, gzip.BestSpeed)
		return gz
	},
}

func GzipMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !strings.Contains(c.GetHeader("Accept-Encoding"), "gzip") {
			c.Next()
			return
		}

		// Skip compression for SSE paths based on URL
		reqPath := strings.ToLower(c.Request.URL.Path)
		if strings.Contains(reqPath, "/sse") {
			c.Next()
			return
		}

		// Skip compression for static uploads and pre-compressed media to preserve HTTP 206 Range requests and save CPU
		if strings.HasPrefix(reqPath, "/uploads") || strings.HasPrefix(reqPath, "/api/v1/uploads") ||
			strings.HasSuffix(reqPath, ".mp4") || strings.HasSuffix(reqPath, ".mov") ||
			strings.HasSuffix(reqPath, ".webm") || strings.HasSuffix(reqPath, ".m4v") ||
			strings.HasSuffix(reqPath, ".jpg") || strings.HasSuffix(reqPath, ".jpeg") ||
			strings.HasSuffix(reqPath, ".png") || strings.HasSuffix(reqPath, ".webp") ||
			strings.HasSuffix(reqPath, ".gif") || strings.HasSuffix(reqPath, ".pdf") ||
			strings.HasSuffix(reqPath, ".zip") {
			c.Next()
			return
		}

		// Skip compression for byte range requests (HTTP 206 Partial Content / video seeking)
		if c.GetHeader("Range") != "" {
			c.Next()
			return
		}

		// Skip compression for assets/already compressed files or SSE
		contentType := c.Writer.Header().Get("Content-Type")
		if strings.Contains(contentType, "image/") || strings.Contains(contentType, "audio/") || strings.Contains(contentType, "video/") || strings.Contains(contentType, "text/event-stream") {
			c.Next()
			return
		}

		gz := gzipWriterPool.Get().(*gzip.Writer)
		gz.Reset(c.Writer)
		defer func() {
			gz.Close()
			gzipWriterPool.Put(gz)
		}()

		c.Header("Content-Encoding", "gzip")
		c.Header("Vary", "Accept-Encoding")
		c.Writer.Header().Del("Content-Length")

		c.Writer = &gzipWriter{ResponseWriter: c.Writer, writer: gz}
		c.Next()
	}
}
