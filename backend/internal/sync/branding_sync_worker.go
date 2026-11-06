package sync

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ─────────────────────────────────────────────────────────────────────────────
// Struct sementara untuk membaca data MySQL Branding
// ─────────────────────────────────────────────────────────────────────────────

type brandingShopRow struct {
	ID            int        `gorm:"column:id"`
	SalesID       *int       `gorm:"column:sales_id"`
	Name          string     `gorm:"column:name"`
	OwnerName     string     `gorm:"column:owner_name"`
	Phone         string     `gorm:"column:phone"`
	AddressDetail string     `gorm:"column:address_detail"`
	ProvinceName  string     `gorm:"column:province_name"`
	CityName      string     `gorm:"column:city_name"`
	DistrictName  string     `gorm:"column:district_name"`
	VillageName   string     `gorm:"column:village_name"`
	Latitude      string     `gorm:"column:latitude"`
	Longitude     string     `gorm:"column:longitude"`
	Photo         *string    `gorm:"column:photo"`
	CreatedAt     *time.Time `gorm:"column:created_at"`
	UpdatedAt     *time.Time `gorm:"column:updated_at"`
}

func (brandingShopRow) TableName() string { return "shops" }

type brandingUserRow struct {
	ID       int     `gorm:"column:id"`
	Name     string  `gorm:"column:name"`
	Email    string  `gorm:"column:email"`
	Phone    *string `gorm:"column:phone"`
	Role     string  `gorm:"column:role"`
	AreaCode *string `gorm:"column:area_code"`
	IsActive bool    `gorm:"column:is_active"`
}

func (brandingUserRow) TableName() string { return "users" }

// ─────────────────────────────────────────────────────────────────────────────
// BrandingSyncWorker
// ─────────────────────────────────────────────────────────────────────────────

// BrandingSyncWorker menyinkronkan data toko & salesman dari MySQL Branding
// ke database PostgreSQL Super App secara berkala (cronjob).
type BrandingSyncWorker struct {
	superDB    *gorm.DB // PostgreSQL Super App
	brandingDB *gorm.DB // MySQL Branding App (opsional, nil jika offline)
	companyID  uuid.UUID
	interval   time.Duration
}

// NewBrandingSyncWorker membuat instance sync worker baru.
// companyIDStr adalah UUID perusahaan di Super App yang akan menjadi pemilik toko hasil impor.
func NewBrandingSyncWorker(superDB, brandingDB *gorm.DB, companyIDStr string, interval time.Duration) *BrandingSyncWorker {
	cid, err := uuid.Parse(companyIDStr)
	if err != nil {
		log.Printf("[BrandingSync] ⚠️  CompanyID tidak valid: %v — gunakan default kosong", err)
	}
	return &BrandingSyncWorker{
		superDB:    superDB,
		brandingDB: brandingDB,
		companyID:  cid,
		interval:   interval,
	}
}

// Start menjalankan cronjob di goroutine terpisah (non-blocking).
func (w *BrandingSyncWorker) Start() {
	if w.brandingDB == nil {
		log.Println("[BrandingSync] MySQL Branding tidak tersambung — Cronjob tidak dijalankan.")
		return
	}
	log.Printf("[BrandingSync] ✅  Cronjob dimulai, interval: %s", w.interval)
	go func() {
		// Langsung jalankan satu kali saat startup
		w.runSync()
		ticker := time.NewTicker(w.interval)
		defer ticker.Stop()
		for range ticker.C {
			w.runSync()
		}
	}()
}

// ─────────────────────────────────────────────────────────────────────────────
// Core sync logic
// ─────────────────────────────────────────────────────────────────────────────

func (w *BrandingSyncWorker) runSync() {
	start := time.Now()
	log.Println("[BrandingSync] 🔄  Memulai sinkronisasi toko dari Branding App...")

	// Resolve CompanyID dynamically if the hardcoded one doesn't exist
	var company domain.Company
	if err := w.superDB.First(&company, "id = ?", w.companyID).Error; err != nil {
		if err := w.superDB.First(&company).Error; err == nil {
			w.companyID = company.ID
			log.Printf("[BrandingSync] ✅  Auto-resolve CompanyID: %s (%s)", company.ID, company.Name)
		} else {
			log.Printf("[BrandingSync] ❌  Tabel companies kosong atau error: %v. Menunggu seeder...", err)
			return
		}
	}

	// 1. Baca semua toko dari MySQL Branding
	var shops []brandingShopRow
	err := w.brandingDB.Raw(`
		SELECT s.*, 
		       p.name as province_name, 
		       c.name as city_name,
		       d.name as district_name,
		       v.name as village_name
		FROM shops s 
		LEFT JOIN indonesia_provinces p ON s.province_id = p.id 
		LEFT JOIN indonesia_cities c ON s.city_id = c.id
		LEFT JOIN indonesia_districts d ON s.district_id = d.id
		LEFT JOIN indonesia_villages v ON s.village_id = v.id
	`).Scan(&shops).Error
	if err != nil {
		log.Printf("[BrandingSync] ❌  Gagal membaca tabel shops: %v", err)
		return
	}

	if len(shops) == 0 {
		log.Println("[BrandingSync] ℹ️  Tidak ada data toko di MySQL Branding.")
		return
	}

	// 2. Kumpulkan semua unique sales_id yang dibutuhkan
	salesIDs := make([]int, 0)
	seen := map[int]bool{}
	for _, s := range shops {
		if s.SalesID != nil && !seen[*s.SalesID] {
			salesIDs = append(salesIDs, *s.SalesID)
			seen[*s.SalesID] = true
		}
	}

	// 3. Baca SEMUA data user (semua role) dari MySQL Branding
	brandingAgentMap := map[int]brandingUserRow{}
	var agents []brandingUserRow
	if err := w.brandingDB.Find(&agents).Error; err != nil {
		log.Printf("[BrandingSync] ⚠️  Gagal membaca users Branding: %v", err)
	}
	for _, a := range agents {
		brandingAgentMap[a.ID] = a
	}

	// 4. Resolve Employees (Preload for efficiency)
	emails := []string{}
	for _, agent := range brandingAgentMap {
		email := strings.ToLower(strings.TrimSpace(agent.Email))
		if email == "" {
			safeRole := "user"
			if agent.Role != "" {
				safeRole = agent.Role
			}
			email = fmt.Sprintf("branding.%s.%d@demo.local", safeRole, agent.ID)
		}
		emails = append(emails, email)
	}

	userByEmail := map[string]domain.User{}
	existingUserIDs := []uuid.UUID{}
	if len(emails) > 0 {
		var existingUsers []domain.User
		lowerEmails := make([]string, len(emails))
		for i, e := range emails {
			lowerEmails[i] = strings.ToLower(e)
		}
		w.superDB.Where("LOWER(email) IN ?", lowerEmails).Find(&existingUsers)
		for _, u := range existingUsers {
			userByEmail[strings.ToLower(u.Email)] = u
			existingUserIDs = append(existingUserIDs, u.ID)
		}
	}

	empByUserID := map[uuid.UUID]domain.Employee{}
	if len(existingUserIDs) > 0 {
		var existingEmps []domain.Employee
		w.superDB.Where("user_id IN ?", existingUserIDs).Find(&existingEmps)
		for _, e := range existingEmps {
			empByUserID[e.UserID] = e
		}
	}

	employeeMap := map[int]uuid.UUID{}
	for brandingID, agent := range brandingAgentMap {
		email := strings.ToLower(strings.TrimSpace(agent.Email))
		if email == "" {
			safeRole := "user"
			if agent.Role != "" {
				safeRole = agent.Role
			}
			email = fmt.Sprintf("branding.%s.%d@demo.local", safeRole, agent.ID)
			agent.Email = email // normalize
		}

		if user, ok := userByEmail[email]; ok {
			// Super App is the Source of Truth for user roles.
			// Never overwrite an existing user's role if already customized.
			// Only upgrade from generic 'employee' if branding role specifies a specific role.
			mappedRole := mapBrandingRole(agent.Role)
			if user.Role == domain.RoleEmployee && mappedRole != domain.RoleEmployee {
				w.superDB.Model(&user).Update("role", mappedRole)
			}

			if emp, eOk := empByUserID[user.ID]; eOk {
				employeeMap[brandingID] = emp.ID
				// Update JobPosition ONLY if currently nil to avoid overwriting specific sales types (Motoris, TO, Booster) set by admin
				if emp.JobPositionID == nil {
					jobPosID := w.resolveJobPosition(agent.Role)
					w.superDB.Model(&emp).Update("job_position_id", jobPosID)
				}
				// Update BranchID if currently nil and area_code exists
				if emp.BranchID == nil && agent.AreaCode != nil {
					branchID := w.resolveBranchID(*agent.AreaCode)
					if branchID != nil {
						w.superDB.Model(&emp).Update("branch_id", branchID)
					}
				}
			} else {
				empID, err := w.createEmployee(user.ID, agent)
				if err == nil {
					employeeMap[brandingID] = empID
				}
			}
		} else {
			// User not found, create sequentially
			empID, err := w.ensureEmployeeExistsSingle(agent)
			if err != nil {
				log.Printf("[BrandingSync] ⚠️  Gagal membuat employee untuk sales [%s]: %v", agent.Name, err)
				continue
			}
			employeeMap[brandingID] = empID
		}
	}

	// 5. Bulk Upsert toko ke PostgreSQL Super App
	var newStores []domain.Store
	for _, shop := range shops {
		var assignedEmpID *uuid.UUID
		if shop.SalesID != nil {
			if eid, ok := employeeMap[*shop.SalesID]; ok {
				assignedEmpID = &eid
			}
		}

		lat, lon := parseCoord(shop.Latitude), parseCoord(shop.Longitude)
		photoURL := ""
		if shop.Photo != nil && *shop.Photo != "" {
			rawPhoto := strings.TrimSpace(*shop.Photo)
			if strings.HasPrefix(rawPhoto, "http") {
				photoURL = rawPhoto
			} else if strings.HasPrefix(rawPhoto, "/uploads/") || strings.HasPrefix(rawPhoto, "uploads/") {
				if strings.HasPrefix(rawPhoto, "/") {
					photoURL = rawPhoto
				} else {
					photoURL = "/" + rawPhoto
				}
			} else {
				photoURL = "/uploads/" + strings.TrimPrefix(rawPhoto, "/")
			}
		}

		brandingID := shop.ID
		newStore := domain.Store{
			ID:                 uuid.New(),
			CompanyID:          w.companyID,
			AssignedEmployeeID: assignedEmpID,
			Name:               shop.Name,
			OwnerName:          shop.OwnerName,
			PhoneNumber:        shop.Phone,
			Address: func() string {
				var addrParts []string
				if detail := strings.TrimSpace(shop.AddressDetail); detail != "" {
					addrParts = append(addrParts, detail)
				}
				if vill := strings.TrimSpace(shop.VillageName); vill != "" {
					addrParts = append(addrParts, vill)
				}
				if dist := strings.TrimSpace(shop.DistrictName); dist != "" {
					addrParts = append(addrParts, dist)
				}
				if city := strings.TrimSpace(shop.CityName); city != "" {
					addrParts = append(addrParts, city)
				}
				if prov := strings.TrimSpace(shop.ProvinceName); prov != "" {
					addrParts = append(addrParts, prov)
				}
				return strings.Join(addrParts, ", ")
			}(),
			Latitude:       lat,
			Longitude:      lon,
			IsActive:       true,
			PhotoURL:       photoURL,
			BrandingShopID: &brandingID,
		}
		newStores = append(newStores, newStore)
	}

	if len(newStores) > 0 {
		log.Printf("[BrandingSync] ⚡ Memulai Bulk Upsert untuk %d toko...", len(newStores))
		importErr := w.superDB.Clauses(clause.OnConflict{
			Columns:     []clause.Column{{Name: "branding_shop_id"}},
			TargetWhere: clause.Where{Exprs: []clause.Expression{clause.Expr{SQL: "branding_shop_id IS NOT NULL"}}},
			DoNothing:   true,
		}).CreateInBatches(&newStores, 1000).Error

		if importErr != nil {
			log.Printf("[BrandingSync] ❌  Gagal melakukan bulk upsert toko: %v", importErr)
		}
	}

	elapsed := time.Since(start).Round(time.Millisecond)
	log.Printf("[BrandingSync] ✅  Selesai dalam %s — Total toko diproses: %d", elapsed, len(newStores))
}

// ─────────────────────────────────────────────────────────────────────────────
// ensureEmployeeExists — buat User + Employee jika belum ada di Super App
// ─────────────────────────────────────────────────────────────────────────────

func (w *BrandingSyncWorker) ensureEmployeeExistsSingle(agent brandingUserRow) (uuid.UUID, error) {
	// Cek apakah user dengan email ini sudah ada di Super App
	email := strings.ToLower(strings.TrimSpace(agent.Email))
	if email == "" {
		safeRole := "user"
		if agent.Role != "" {
			safeRole = agent.Role
		}
		email = fmt.Sprintf("branding.%s.%d@demo.local", safeRole, agent.ID)
	}

	mappedRole := mapBrandingRole(agent.Role)

	// Fallback lookup: Sebelum membuat user baru atau mencari berdasarkan email,
	// cek apakah ada karyawan (employee) yang memiliki nomor telepon atau nama yang sama.
	// Ini mencegah duplikasi jika sales sudah dibuat di Super App dengan email asli mereka.
	var fallbackEmp domain.Employee

	// 1. Cek berdasarkan nomor telepon
	if agent.Phone != nil && *agent.Phone != "" {
		if err := w.superDB.Where("phone_number = ?", *agent.Phone).First(&fallbackEmp).Error; err == nil {
			if fallbackEmp.JobPositionID == nil {
				jobPosID := w.resolveJobPosition(agent.Role)
				w.superDB.Model(&fallbackEmp).Update("job_position_id", jobPosID)
			}
			if fallbackEmp.BranchID == nil && agent.AreaCode != nil {
				branchID := w.resolveBranchID(*agent.AreaCode)
				if branchID != nil {
					w.superDB.Model(&fallbackEmp).Update("branch_id", branchID)
				}
			}
			return fallbackEmp.ID, nil
		}
	}

	// 2. Cek berdasarkan nama depan & belakang (case-insensitive)
	firstName, lastName := splitName(agent.Name)
	if err := w.superDB.Where("LOWER(first_name) = ? AND LOWER(last_name) = ?", strings.ToLower(firstName), strings.ToLower(lastName)).First(&fallbackEmp).Error; err == nil {
		if fallbackEmp.JobPositionID == nil {
			jobPosID := w.resolveJobPosition(agent.Role)
			w.superDB.Model(&fallbackEmp).Update("job_position_id", jobPosID)
		}
		if fallbackEmp.BranchID == nil && agent.AreaCode != nil {
			branchID := w.resolveBranchID(*agent.AreaCode)
			if branchID != nil {
				w.superDB.Model(&fallbackEmp).Update("branch_id", branchID)
			}
		}
		return fallbackEmp.ID, nil
	}

	var existingUser domain.User
	err := w.superDB.Where("LOWER(email) = ?", strings.ToLower(email)).First(&existingUser).Error
	if err == nil {
		// Super App is the Source of Truth for user roles.
		// Never overwrite an existing user's role if already customized.
		// Only upgrade from generic 'employee' if branding role specifies a specific role.
		if existingUser.Role == domain.RoleEmployee && mappedRole != domain.RoleEmployee {
			w.superDB.Model(&existingUser).Update("role", mappedRole)
		}
		// User sudah ada, cari employee-nya
		var existingEmp domain.Employee
		if err2 := w.superDB.Where("user_id = ?", existingUser.ID).First(&existingEmp).Error; err2 == nil {
			if existingEmp.JobPositionID == nil {
				jobPosID := w.resolveJobPosition(agent.Role)
				w.superDB.Model(&existingEmp).Update("job_position_id", jobPosID)
			}
			if existingEmp.BranchID == nil && agent.AreaCode != nil {
				branchID := w.resolveBranchID(*agent.AreaCode)
				if branchID != nil {
					w.superDB.Model(&existingEmp).Update("branch_id", branchID)
				}
			}
			return existingEmp.ID, nil
		}
		// User ada tapi belum punya employee record → buat employee
		return w.createEmployee(existingUser.ID, agent)
	}

	// User belum ada → buat User baru
	log.Printf("[BrandingSync] 📝  Membuat user baru untuk %s: %s (%s)", agent.Role, agent.Name, email)

	// Generate password acak (sales harus reset via admin)
	tempPassword := generateTempPassword()
	hash, err := bcrypt.GenerateFromPassword([]byte(tempPassword), bcrypt.DefaultCost)
	if err != nil {
		return uuid.Nil, fmt.Errorf("bcrypt gagal: %w", err)
	}

	newUser := domain.User{
		Email:        email,
		PasswordHash: string(hash),
		Role:         mappedRole,
		IsActive:     agent.IsActive,
		CompanyID:    &w.companyID,
	}

	if err := w.superDB.Create(&newUser).Error; err != nil {
		return uuid.Nil, fmt.Errorf("gagal create user: %w", err)
	}

	log.Printf("[BrandingSync]    → User dibuat: %s (temp password: %s)", email, tempPassword)
	return w.createEmployee(newUser.ID, agent)
}

func (w *BrandingSyncWorker) createEmployee(userID uuid.UUID, agent brandingUserRow) (uuid.UUID, error) {
	// Split nama menjadi first + last
	firstName, lastName := splitName(agent.Name)

	phone := ""
	if agent.Phone != nil {
		phone = *agent.Phone
	}

	joinDate := time.Now()
	empStatus := "PERMANENT"
	jobPosID := w.resolveJobPosition(agent.Role)

	safeRole := "USER"
	if agent.Role != "" {
		safeRole = strings.ToUpper(agent.Role)
	}

	var branchID *uuid.UUID
	if agent.AreaCode != nil {
		branchID = w.resolveBranchID(*agent.AreaCode)
	}

	newEmp := domain.Employee{
		UserID:           userID,
		CompanyID:        &w.companyID,
		BranchID:         branchID,
		EmployeeIDNumber: fmt.Sprintf("BRANDING-%s-%d", safeRole, agent.ID),
		FirstName:        firstName,
		LastName:         lastName,
		PhoneNumber:      phone,
		JoinDate:         joinDate,
		EmploymentStatus: empStatus,
		JobPositionID:    &jobPosID,
	}

	if err := w.superDB.Create(&newEmp).Error; err != nil {
		return uuid.Nil, fmt.Errorf("gagal create employee: %w", err)
	}

	log.Printf("[BrandingSync]    → Employee dibuat: %s %s (ID: %s)", firstName, lastName, newEmp.ID)
	return newEmp.ID, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────────────────────────────────────

func splitName(fullName string) (first, last string) {
	parts := strings.Fields(fullName)
	if len(parts) == 0 {
		return "Sales", "Branding"
	}
	if len(parts) == 1 {
		return parts[0], ""
	}
	return parts[0], strings.Join(parts[1:], " ")
}

func parseCoord(s string) float64 {
	if s == "" {
		return 0
	}
	var f float64
	fmt.Sscanf(s, "%f", &f)
	return f
}

func generateTempPassword() string {
	return os.Getenv("TRIAL_STAFF_PASSWORD")
}

func mapBrandingRole(brandingRole string) domain.Role {
	roleLower := strings.ToLower(strings.TrimSpace(brandingRole))
	switch roleLower {
	case "master":
		return domain.RoleSuperAdmin
	case "admin", "admin_nota":
		return domain.RoleHRAdmin
	case "spv", "manager":
		return domain.RoleSupervisorSales
	case "editor":
		return domain.RoleAdminDesigner
	case "operator":
		return domain.RoleOperator
	case "sales":
		return domain.RoleEmployee
	default:
		return domain.RoleEmployee
	}
}

func (w *BrandingSyncWorker) resolveJobPosition(brandingRole string) uuid.UUID {
	roleLower := strings.ToLower(strings.TrimSpace(brandingRole))
	jobTitle := "Karyawan"
	if roleLower != "" {
		jobTitle = strings.ToUpper(roleLower[:1]) + roleLower[1:]
	}

	var jobPos domain.JobPosition
	err := w.superDB.Where("LOWER(title) = ? AND company_id = ?", strings.ToLower(jobTitle), w.companyID).First(&jobPos).Error
	if err != nil {
		jobPos = domain.JobPosition{
			CompanyID: &w.companyID,
			Title:     jobTitle,
			Level:     1,
		}
		w.superDB.Create(&jobPos)
	}
	return jobPos.ID
}

func (w *BrandingSyncWorker) resolveBranchID(areaCode string) *uuid.UUID {
	if areaCode == "" {
		return nil
	}

	searchStr := "%" + strings.ToLower(strings.TrimSpace(areaCode)) + "%"
	var branch domain.Branch
	err := w.superDB.Where("company_id = ? AND LOWER(name) LIKE ?", w.companyID, searchStr).First(&branch).Error
	if err == nil {
		return &branch.ID
	}

	cleanArea := strings.ReplaceAll(strings.ToLower(areaCode), "kabupaten ", "")
	cleanArea = strings.ReplaceAll(cleanArea, "kota ", "")
	searchStr2 := "%" + strings.TrimSpace(cleanArea) + "%"
	err = w.superDB.Where("company_id = ? AND LOWER(name) LIKE ?", w.companyID, searchStr2).First(&branch).Error
	if err == nil {
		return &branch.ID
	}

	return nil
}
