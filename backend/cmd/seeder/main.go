package main

import (
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"

	"example.com/hris/backend/config"
	"example.com/hris/backend/internal/domain"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func main() {
	log.Println("Starting database seeder...")

	cfg := config.LoadConfig()
	db := config.ConnectDB(cfg)

	// Run GORM AutoMigrate first to ensure all table columns (like paid_amount) exist
	log.Println("Running GORM AutoMigrate from seeder...")
	if err := db.AutoMigrate(
		&domain.AccessRole{},
		&domain.AppModule{},
		&domain.AppMenu{},
		&domain.Company{},
		&domain.Branch{},
		&domain.Department{},
		&domain.JobPosition{},
		&domain.Employee{},
		&domain.Store{},
		&domain.User{},
		&domain.SalesKPI{},
		&domain.BannerOrder{},
		&domain.Factory{},
		&domain.Product{},
		&domain.FactoryStock{},
		&domain.ProductionLog{},
		&domain.ProductTransfer{},
		&domain.FactoryInventoryLog{},
		&domain.WarehouseStock{},
		&domain.WarehouseLog{},
		&domain.Vehicle{},
		&domain.Notification{},
		&domain.DeliveryBatch{},
		&domain.DeliveryItem{},
		&domain.SalesPayment{},
		&domain.SalesItem{},
		&domain.SalesStock{},
		&domain.SalesTransfer{},
		&domain.SalesTransferCorrection{},
		&domain.SalesVisit{},
		&domain.SalesTransaction{},
		&domain.SalesOrder{},
		&domain.SalesOrderItem{},
		&domain.SalesOrderItemBatch{},
		&domain.SalesReturn{},
		&domain.SalesReturnItem{},
		&domain.Reimbursement{},
		&domain.ProductionRecipe{},
		&domain.ProductionRecipeItem{},
		&domain.Shift{},
		&domain.EmployeeShift{},
		&domain.KeyAccount{},
		&domain.MTContract{},
		&domain.MTPurchaseOrder{},
		&domain.MTMerchandising{},
		&domain.MTPromotion{},
		&domain.Supplier{},
		&domain.SupplierReceipt{},
		&domain.SupplierReceiptItem{},
		&domain.SupplierStock{},
	); err != nil {
		log.Fatalf("AutoMigrate failed in seeder: %v", err)
	}

	rand.Seed(time.Now().UnixNano())

	// 1. Create Company
	company := domain.Company{
		Name:      requiredEnv("TRIAL_COMPANY_NAME"),
		Code:      "DEMO",
		TaxNumber: "",
		Address:   "Kantor Demo",
	}
	if err := db.FirstOrCreate(&company, domain.Company{Name: company.Name}).Error; err != nil {
		log.Fatalf("Failed to seed company: %v", err)
	}
	log.Printf("Company seeded: %s\n", company.ID)

	// 2. Create Branch
	branch := domain.Branch{
		CompanyID: &company.ID,
		Name:      "Kantor Pusat Demo",
		Address:   "Kantor Demo",
	}
	if err := db.FirstOrCreate(&branch, domain.Branch{Name: branch.Name}).Error; err != nil {
		log.Fatalf("Failed to seed branch: %v", err)
	}
	log.Printf("Branch seeded: %s\n", branch.ID)

	// 3. Create Departments
	departments := []domain.Department{
		{CompanyID: &company.ID, Name: "Engineering"},
		{CompanyID: &company.ID, Name: "Product"},
		{CompanyID: &company.ID, Name: "Human Resources"},
	}
	for i := range departments {
		if err := db.FirstOrCreate(&departments[i], domain.Department{Name: departments[i].Name}).Error; err != nil {
			log.Fatalf("Failed to seed department %s: %v", departments[i].Name, err)
		}
	}
	log.Println("Departments seeded.")

	// 4. Create Job Positions
	positions := []domain.JobPosition{
		{CompanyID: &company.ID, Title: "CTO", Level: 9},
		{CompanyID: &company.ID, Title: "Software Engineer", Level: 3},
		{CompanyID: &company.ID, Title: "HR Manager", Level: 5},
		{CompanyID: &company.ID, Title: "Sales Motoris", Level: 2},
		{CompanyID: &company.ID, Title: "Sales Task Order", Level: 2},
		{CompanyID: &company.ID, Title: "Sales Booster", Level: 2},
		{CompanyID: &company.ID, Title: "Kurir Pengiriman", Level: 2},
		{CompanyID: &company.ID, Title: "Admin Gudang", Level: 2},
	}
	for i := range positions {
		if err := db.FirstOrCreate(&positions[i], domain.JobPosition{Title: positions[i].Title, CompanyID: &company.ID}).Error; err != nil {
			log.Fatalf("Failed to seed job position %s: %v", positions[i].Title, err)
		}
	}
	log.Println("Job Positions seeded.")

	// Hash password
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(requiredEnv("TRIAL_ADMIN_PASSWORD")), bcrypt.DefaultCost)
	staffPassword, _ := bcrypt.GenerateFromPassword([]byte(requiredEnv("TRIAL_STAFF_PASSWORD")), bcrypt.DefaultCost)

	// 5. Create Users and Employees (Admin)
	adminUser := domain.User{
		Email:        requiredEnv("TRIAL_ADMIN_EMAIL"),
		PasswordHash: string(hashedPassword),
		Role:         domain.RoleSuperAdmin,
		IsActive:     true,
		CompanyID:    &company.ID,
	}
	if err := db.Where(domain.User{Email: adminUser.Email}).FirstOrCreate(&adminUser).Error; err != nil {
		log.Fatalf("Failed to seed admin user: %v", err)
	}
	// Force SuperAdmin role in case it was modified
	if adminUser.Role != domain.RoleSuperAdmin {
		adminUser.Role = domain.RoleSuperAdmin
		if err := db.Model(&adminUser).Update("role", domain.RoleSuperAdmin).Error; err != nil {
			log.Fatalf("Failed to force superadmin role for admin user: %v", err)
		}
	}

	adminEmployee := domain.Employee{
		UserID:           adminUser.ID,
		CompanyID:        &company.ID,
		BranchID:         &branch.ID,
		DepartmentID:     &departments[0].ID,
		JobPositionID:    &positions[0].ID,
		EmployeeIDNumber: "EMP-001",
		FirstName:        "Admin",
		LastName:         "Demo",
		JoinDate:         time.Now().AddDate(-2, 0, 0),
		EmploymentStatus: "PERMANENT",
	}
	if err := db.Where(domain.Employee{EmployeeIDNumber: adminEmployee.EmployeeIDNumber}).FirstOrCreate(&adminEmployee).Error; err != nil {
		log.Fatalf("Failed to seed admin employee: %v", err)
	}
	log.Printf("Admin seeded: %s\n", adminEmployee.ID)

	// 6. Create Users and Employees (Staff)
	staffUser := domain.User{
		Email:        "staff@demo.local",
		PasswordHash: string(staffPassword),
		Role:         domain.RoleEmployee,
		IsActive:     true,
		CompanyID:    &company.ID,
	}
	if err := db.Where(domain.User{Email: staffUser.Email}).FirstOrCreate(&staffUser).Error; err != nil {
		log.Fatalf("Failed to seed staff user: %v", err)
	}

	staffEmployee := domain.Employee{
		UserID:           staffUser.ID,
		CompanyID:        &company.ID,
		BranchID:         &branch.ID,
		DepartmentID:     &departments[0].ID,
		JobPositionID:    &positions[1].ID,
		EmployeeIDNumber: "EMP-002",
		FirstName:        "Budi",
		LastName:         "Pratama",
		JoinDate:         time.Now().AddDate(-1, 0, 0),
		EmploymentStatus: "PERMANENT",
		ManagerID:        &adminEmployee.ID,
	}
	if err := db.Where(domain.Employee{EmployeeIDNumber: staffEmployee.EmployeeIDNumber}).FirstOrCreate(&staffEmployee).Error; err != nil {
		log.Fatalf("Failed to seed staff employee: %v", err)
	}
	log.Printf("Staff seeded: %s\n", staffEmployee.ID)

	allEmployees := []domain.Employee{adminEmployee, staffEmployee}

	// Additional fictional HR staff make employee lists and approval flows useful.
	hrStaff := []struct {
		email, first, last, number string
		role                       domain.Role
		department, position       int
	}{
		{"hr@demo.local", "Nadia", "Putri", "EMP-003", domain.RoleHRAdmin, 2, 2},
		{"manager@demo.local", "Rizky", "Mahendra", "EMP-004", domain.RoleManager, 1, 0},
		{"employee1@demo.local", "Dina", "Lestari", "EMP-005", domain.RoleEmployee, 2, 1},
		{"employee2@demo.local", "Fajar", "Hidayat", "EMP-006", domain.RoleEmployee, 1, 1},
		{"employee3@demo.local", "Maya", "Sari", "EMP-007", domain.RoleEmployee, 0, 1},
	}
	for _, sample := range hrStaff {
		user := domain.User{Email: sample.email, PasswordHash: string(staffPassword), Role: sample.role, IsActive: true, CompanyID: &company.ID}
		if err := db.Where(domain.User{Email: sample.email}).FirstOrCreate(&user).Error; err != nil {
			log.Fatalf("Failed to seed sample user %s: %v", sample.email, err)
		}
		emp := domain.Employee{
			UserID: user.ID, CompanyID: &company.ID, BranchID: &branch.ID,
			DepartmentID:     &departments[sample.department].ID,
			JobPositionID:    &positions[sample.position].ID,
			EmployeeIDNumber: sample.number, FirstName: sample.first, LastName: sample.last,
			JoinDate: time.Now().AddDate(-1, 0, 0), EmploymentStatus: "PERMANENT",
			ManagerID: &adminEmployee.ID,
		}
		if err := db.Where(domain.Employee{EmployeeIDNumber: sample.number}).FirstOrCreate(&emp).Error; err != nil {
			log.Fatalf("Failed to seed sample employee %s: %v", sample.number, err)
		}
		allEmployees = append(allEmployees, emp)
	}

	// 7. Create Shifts and EmployeeShifts
	shift := domain.Shift{
		CompanyID: &company.ID,
		Name:      "Regular Shift",
		StartTime: "09:00:00",
		EndTime:   "18:00:00",
	}
	if err := db.FirstOrCreate(&shift, domain.Shift{Name: shift.Name}).Error; err != nil {
		log.Fatalf("Failed to seed shift: %v", err)
	}

	for _, emp := range allEmployees {
		db.FirstOrCreate(&domain.EmployeeShift{
			EmployeeID: emp.ID,
			ShiftID:    shift.ID,
			Date:       time.Now(),
		})
	}

	// 8. Create Random Attendances
	log.Println("Seeding attendances...")
	for _, emp := range allEmployees {
		for d := 1; d <= 20; d++ {
			date := time.Now().AddDate(0, 0, -d)
			if date.Weekday() == time.Saturday || date.Weekday() == time.Sunday {
				continue
			}

			checkInHour := 8
			checkInMin := rand.Intn(90)
			if checkInMin >= 60 {
				checkInHour = 9
				checkInMin -= 60
			}
			checkInTime := time.Date(date.Year(), date.Month(), date.Day(), checkInHour, checkInMin, 0, 0, time.Local)
			checkOutTime := checkInTime.Add(time.Duration(8+rand.Intn(2)) * time.Hour)

			status := "PRESENT"
			if checkInHour == 9 {
				status = "LATE"
			}

			// Check if attendance already exists for this day
			var existingAtt domain.AttendanceLog
			startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.Local)
			endOfDay := startOfDay.AddDate(0, 0, 1)

			err := db.Where("employee_id = ? AND clock_in_time >= ? AND clock_in_time < ?", emp.ID, startOfDay, endOfDay).First(&existingAtt).Error
			if err != nil {
				att := domain.AttendanceLog{
					EmployeeID:   emp.ID,
					ShiftID:      &shift.ID,
					ClockInTime:  &checkInTime,
					ClockOutTime: &checkOutTime,
					ClockInLat:   -6.2088,
					ClockInLong:  106.8456,
					Status:       status,
				}
				if err := db.Create(&att).Error; err != nil {
					log.Fatalf("Failed to seed attendance: %v", err)
				}
			}
		}
	}

	// 9. Leave Types & Leaves
	leaveType := domain.LeaveType{
		CompanyID: &company.ID,
		Name:      "Annual Leave",
	}
	if err := db.FirstOrCreate(&leaveType, domain.LeaveType{Name: leaveType.Name}).Error; err != nil {
		log.Fatalf("Failed to seed leave type: %v", err)
	}

	// 9b. Add Izin Types
	izinTypes := []domain.LeaveType{
		{CompanyID: &company.ID, Name: "Izin Sakit", IsPaid: true, RequiresQuota: false, DefaultQuota: 0},
		{CompanyID: &company.ID, Name: "Izin Terkena Musibah", IsPaid: true, RequiresQuota: false, DefaultQuota: 0},
		{CompanyID: &company.ID, Name: "Izin Lainnya", IsPaid: false, RequiresQuota: false, DefaultQuota: 0},
	}
	for _, it := range izinTypes {
		db.FirstOrCreate(&domain.LeaveType{}, it)
	}

	log.Println("Seeding leaves...")
	leaves := []domain.LeaveRequest{
		{EmployeeID: staffEmployee.ID, LeaveTypeID: leaveType.ID, StartDate: time.Now().AddDate(0, 0, 5), EndDate: time.Now().AddDate(0, 0, 8), Reason: "Vacation", Status: "PENDING"},
		{EmployeeID: staffEmployee.ID, LeaveTypeID: leaveType.ID, StartDate: time.Now().AddDate(0, 0, -2), EndDate: time.Now().AddDate(0, 0, -2), Reason: "Sick", Status: "APPROVED", ApprovedBy: &adminEmployee.ID},
	}
	for i := range leaves {
		db.Where(domain.LeaveRequest{
			EmployeeID: leaves[i].EmployeeID,
			StartDate:  leaves[i].StartDate,
			EndDate:    leaves[i].EndDate,
		}).FirstOrCreate(&leaves[i])
	}

	// 10. Create Overtimes
	log.Println("Seeding overtimes...")
	overtimes := []domain.Overtime{
		{EmployeeID: staffEmployee.ID, Date: time.Now().AddDate(0, 0, -1), StartTime: time.Now().AddDate(0, 0, -1).Add(18 * time.Hour), EndTime: time.Now().AddDate(0, 0, -1).Add(20 * time.Hour), DurationMinutes: 120, Reason: "Project deadline", Status: domain.OvertimePending},
	}
	for i := range overtimes {
		db.Where(domain.Overtime{
			EmployeeID: overtimes[i].EmployeeID,
			Date:       overtimes[i].Date,
			Reason:     overtimes[i].Reason,
		}).FirstOrCreate(&overtimes[i])
	}

	// 11. Initialize Leave Balances for all employees and types
	log.Println("Initializing leave balances...")
	var allEmployeesForBalances []domain.Employee
	db.Find(&allEmployeesForBalances)
	var allLeaveTypes []domain.LeaveType
	db.Find(&allLeaveTypes)

	currentYear := time.Now().Year()
	for _, emp := range allEmployeesForBalances {
		for _, lt := range allLeaveTypes {
			db.FirstOrCreate(&domain.LeaveBalance{}, domain.LeaveBalance{
				EmployeeID:  emp.ID,
				LeaveTypeID: lt.ID,
				Year:        currentYear,
			}).Updates(domain.LeaveBalance{
				BalanceTotal: lt.DefaultQuota,
			})
		}
	}

	// 12. Create Sales Accounts and Data
	log.Println("Seeding sales operational data...")
	salesRoles := []struct {
		Email     string
		FirstName string
		Position  string
		EmpID     string
	}{
		{"motoris@demo.local", "Andi", "Sales Motoris", "SLS-MOT-001"},
		{"to@demo.local", "Siti", "Sales Task Order", "SLS-TO-001"},
		{"booster@demo.local", "Eko", "Sales Booster", "SLS-BST-001"},
		{"delivery@demo.local", "Dedi", "Kurir Pengiriman", "DEL-001"},
		{"gudang@demo.local", "Budi", "Admin Gudang", "WH-001"},
	}

	if os.Getenv("TRIAL_SEED_OPERATIONS") == "true" {
		for _, sr := range salesRoles {
			user := domain.User{
				Email:        sr.Email,
				PasswordHash: string(staffPassword),
				Role:         domain.RoleEmployee,
				IsActive:     true,
				CompanyID:    &company.ID,
			}
			db.FirstOrCreate(&user, domain.User{Email: user.Email})

			var pos domain.JobPosition
			db.Where("title = ? AND company_id = ?", sr.Position, company.ID).First(&pos)

			emp := domain.Employee{
				UserID:           user.ID,
				CompanyID:        &company.ID,
				BranchID:         &branch.ID,
				DepartmentID:     &departments[0].ID, // Assuming first dept
				JobPositionID:    &pos.ID,
				EmployeeIDNumber: sr.EmpID,
				FirstName:        sr.FirstName,
				LastName:         "Demo",
				JoinDate:         time.Now().AddDate(0, -6, 0),
				EmploymentStatus: "PERMANENT",
			}
			db.FirstOrCreate(&emp, domain.Employee{EmployeeIDNumber: emp.EmployeeIDNumber})

			// Seed KPI for this month
			kpi := domain.SalesKPI{
				EmployeeID:      emp.ID,
				TargetOmzet:     50000000,
				TargetNewStores: 10,
				PeriodMonth:     int(time.Now().Month()),
				PeriodYear:      time.Now().Year(),
			}
			db.FirstOrCreate(&kpi, domain.SalesKPI{EmployeeID: emp.ID, PeriodMonth: kpi.PeriodMonth, PeriodYear: kpi.PeriodYear})

			// Seed some Stores for this sales
			if sr.Position != "Kurir Pengiriman" {
				for i := 1; i <= 3; i++ {
					store := domain.Store{
						CompanyID:          company.ID,
						AssignedEmployeeID: &emp.ID,
						Name:               "Toko " + sr.FirstName + " " + strconv.Itoa(i),
						OwnerName:          "Owner " + strconv.Itoa(i),
						Address:            "Alamat Toko " + strconv.Itoa(i),
						IsActive:           true,
					}
					db.FirstOrCreate(&store, domain.Store{Name: store.Name, CompanyID: store.CompanyID})

					// Seed a transaction for one of the stores
					if i == 1 {
						trx := domain.SalesTransaction{
							CompanyID:       company.ID,
							StoreID:         &store.ID,
							EmployeeID:      emp.ID,
							TotalAmount:     1500000,
							StoreCategory:   "TOKO_LAMA",
							PeriodMonth:     int(time.Now().Month()),
							PeriodYear:      time.Now().Year(),
							TransactionDate: time.Now(),
							Status:          "VERIFIED",
						}
						var existing int64
						db.Model(&domain.SalesTransaction{}).Where("company_id = ? AND employee_id = ? AND store_id = ? AND period_month = ? AND period_year = ?", company.ID, emp.ID, store.ID, trx.PeriodMonth, trx.PeriodYear).Count(&existing)
						if existing == 0 {
							if err := db.Create(&trx).Error; err != nil {
								log.Fatalf("Failed to seed transaction: %v", err)
							}
						}
					}
				}
			}
		}

	}

	// 14. Seed Access Control Modules and Menus
	seedAccessControl(db)

	log.Println("Database seeded successfully!")
}
func seedAccessControl(db *gorm.DB) {
	log.Println("Seeding access control modules and menus...")

	type MenuItem struct {
		Path string
		Name string
	}

	modules := []struct {
		Name        string
		Description string
		Menus       []MenuItem
	}{
		{
			Name:        "HRIS",
			Description: "Human Resource Information System",
			Menus: []MenuItem{
				{Path: "/", Name: "Dashboard Utama"},
				{Path: "/employees", Name: "Data Karyawan"},
				{Path: "/attendance", Name: "Kehadiran / Presensi"},
				{Path: "/leaves", Name: "Pengajuan Cuti"},
				{Path: "/overtimes", Name: "Lembur Karyawan"},
				{Path: "/payroll", Name: "Penggajian (Payroll)"},
				{Path: "/pkwt", Name: "Manajemen Kontrak PKWT"},
				{Path: "/analytics", Name: "Analytics & Laporan HR"},
				{Path: "/users", Name: "Manajemen Pengguna"},
				{Path: "/leave-types", Name: "Tipe Cuti"},
				{Path: "/organization", Name: "Struktur Organisasi"},
				{Path: "/job-positions", Name: "Jabatan & Posisi"},
				{Path: "/shifts", Name: "Pengaturan Shift"},
				{Path: "/announcements", Name: "Pengumuman Internal"},
				{Path: "/system-modules", Name: "Manajemen Modul"},
				{Path: "/access-roles", Name: "Hak Akses (Role Dinamis)"},
			},
		},
		{
			Name:        "Sales",
			Description: "Operasional Sales",
			Menus: []MenuItem{
				{Path: "/sales", Name: "Dashboard Sales"},
				{Path: "/sales/spv-approval", Name: "Persetujuan SPV"},
				{Path: "/sales/kunjungan", Name: "Kunjungan Salesman"},
				{Path: "/sales/pesanan", Name: "Pesanan Penjualan (SO)"},
				{Path: "/sales/delivery", Name: "Pengiriman Sales"},
				{Path: "/sales/surat-jalan", Name: "Surat Jalan Sales"},
				{Path: "/sales/penjualan", Name: "Transaksi Penjualan"},
				{Path: "/sales/penerimaan", Name: "Penerimaan Pembayaran"},
				{Path: "/sales/salesman", Name: "Data Salesman"},
				{Path: "/sales/customers", Name: "Data Pelanggan Toko"},
				{Path: "/sales/monitoring", Name: "Monitoring Sales"},
				{Path: "/sales/tracking-noo-ec", Name: "Tracking NOO & EC"},
				{Path: "/sales/kpi", Name: "Performa KPI Sales"},
				{Path: "/sales/audit-logs", Name: "Audit Log Sales"},
			},
		},
		{
			Name:        "Sales MT",
			Description: "Modern Trade",
			Menus: []MenuItem{
				{Path: "/sales/modern-trade/kunjungan", Name: "Kunjungan Toko MT"},
				{Path: "/sales/modern-trade/po", Name: "Purchase Order MT"},
				{Path: "/sales/modern-trade/merchandising", Name: "Merchandising & Display"},
				{Path: "/sales/modern-trade/settlement", Name: "Settlement Pembayaran MT"},
				{Path: "/sales/modern-trade/spg", Name: "Manajemen SPG"},
				{Path: "/sales/modern-trade/spg-reports", Name: "Laporan Penjualan SPG"},
				{Path: "/sales/modern-trade/kam", Name: "Key Account Management"},
			},
		},
		{
			Name:        "Branding",
			Description: "Branding & Marketing",
			Menus: []MenuItem{
				{Path: "/sales/spanduk", Name: "Dashboard Banner & Spanduk"},
				{Path: "/sales/operator", Name: "Operator Cetak & Produksi"},
				{Path: "/sales/editor", Name: "Desainer & Editor Banner"},
				{Path: "/sales/spanduk-editor", Name: "Workspace Editor Spanduk"},
				{Path: "/sales/spanduk-validator", Name: "Validasi Desain"},
				{Path: "/sales/spanduk-vendor", Name: "Manajemen Vendor Cetak"},
				{Path: "/sales/spanduk-pemasangan", Name: "Bukti Tayang & Pemasangan"},
			},
		},
		{
			Name:        "Pabrik",
			Description: "Operasional Pabrik",
			Menus: []MenuItem{
				{Path: "/factory", Name: "Dashboard Pabrik"},
				{Path: "/factory/stock", Name: "Manajemen Stok Barang Jadi"},
				{Path: "/factory/transfers", Name: "Manajemen Pengiriman Cabang"},
				{Path: "/factory/returns", Name: "Barang Returan Pabrik"},
				{Path: "/factory/raw-materials", Name: "Stok Bahan Baku"},
				{Path: "/factory/raw-materials/orders", Name: "PO Pengadaan Bahan"},
				{Path: "/factory/raw-materials/receipts", Name: "Penerimaan Supplier"},
				{Path: "/factory/management", Name: "Manajemen Pabrik"},
				{Path: "/factory/warehouses", Name: "Gudang Pabrik"},
				{Path: "/factory/products", Name: "Produk"},
				{Path: "/factory/suppliers", Name: "Data Supplier"},
				{Path: "/vehicles", Name: "Manajemen Kendaraan"},
			},
		},
		{
			Name:        "Gudang",
			Description: "Gudang Cabang",
			Menus: []MenuItem{
				{Path: "/warehouse", Name: "Dashboard Gudang"},
				{Path: "/warehouse/stock", Name: "Persediaan Gudang"},
				{Path: "/warehouse/het-pricing", Name: "Harga HET Pelanggan"},
				{Path: "/warehouse/sales-transfers", Name: "Pindah Barang Antar Cabang"},
				{Path: "/warehouse/surat-jalan", Name: "Data Surat Jalan"},
				{Path: "/warehouse/incoming-orders", Name: "SJ Masuk Gudang"},
				{Path: "/warehouse/do-management", Name: "Kelola DO & Cadangan"},
				{Path: "/warehouse/leftover", Name: "Sisa Stok DO"},
				{Path: "/warehouse/retur", Name: "Retur dari Pengiriman"},
				{Path: "/warehouse/tagihan-pabrik", Name: "Tagihan Pabrik"},
				{Path: "/warehouse/logs", Name: "Log Mutasi Stok"},
				{Path: "/warehouse/supplier-receipts", Name: "Penerimaan Supplier Cabang"},
			},
		},
		{
			Name:        "Keuangan",
			Description: "Keuangan",
			Menus: []MenuItem{
				{Path: "/finance", Name: "Dashboard Keuangan"},
				{Path: "/reimbursements", Name: "Reimbursement"},
				{Path: "/finance/approval", Name: "Approval Keuangan"},
				{Path: "/finance/setoran-faktur", Name: "Setoran Faktur Pabrik"},
				{Path: "/finance/kpi-sales", Name: "KPI Omzet Sales"},
				{Path: "/finance/tracking-noo-ec", Name: "Tracking NOO & EC"},
				{Path: "/finance/profit-loss", Name: "Laba Rugi"},
				{Path: "/finance/export", Name: "Export Data Keuangan"},
				{Path: "/finance/bank-accounts", Name: "Rekening Bank"},
			},
		},
		{
			Name:        "Marketplace",
			Description: "Marketplace",
			Menus: []MenuItem{
				{Path: "/marketplace", Name: "Dashboard Marketplace"},
				{Path: "/marketplace/stock", Name: "Stok Marketplace"},
				{Path: "/marketplace/inventories", Name: "Kelola Inventori Marketplace"},
				{Path: "/marketplace/pricing", Name: "Penetapan Harga Marketplace"},
				{Path: "/marketplace/audit-trail", Name: "Audit Trail Marketplace"},
			},
		},
	}

	for _, m := range modules {
		mod := domain.AppModule{
			Name:        m.Name,
			Description: m.Description,
		}
		if err := db.Where(domain.AppModule{Name: m.Name}).FirstOrCreate(&mod).Error; err != nil {
			log.Printf("Error seeding module %s: %v\n", m.Name, err)
			continue
		}

		for _, item := range m.Menus {
			menu := domain.AppMenu{
				ModuleID: mod.ID,
				Name:     item.Name,
				Path:     item.Path,
				Level:    1,
			}
			// Update name if menu already exists
			if err := db.Where(domain.AppMenu{Path: item.Path, ModuleID: mod.ID}).Assign(domain.AppMenu{Name: item.Name}).FirstOrCreate(&menu).Error; err != nil {
				log.Printf("Error seeding menu %s: %v\n", item.Path, err)
			}
		}
	}

	// Seed Preset Access Roles for Pabrik
	seedPresetPabrikRoles(db)

	log.Println("Finished seeding access control modules and menus.")
}

func seedPresetPabrikRoles(db *gorm.DB) {
	presetRoles := []struct {
		Name        string
		Description string
		MenuPaths   []string
	}{
		{
			Name:        "Admin Gudang Bahan Jadi (Stok Only)",
			Description: "Hak akses manajemen stok barang jadi pabrik dan monitoring request distributor.",
			MenuPaths:   []string{"/factory/stock", "/warehouse/distributor-requests"},
		},
		{
			Name:        "Admin Gudang Bahan Baku & PO",
			Description: "Hak akses stok bahan baku, input/pemantauan PO bahan pengadaan, dan penerimaan supplier.",
			MenuPaths:   []string{"/factory/raw-materials", "/factory/raw-materials/orders", "/factory/raw-materials/receipts"},
		},
		{
			Name:        "Admin Gudang Bahan Baku & Retur",
			Description: "Hak akses stok bahan baku, PO pengadaan, penerimaan supplier, dan administrasi sortir returan.",
			MenuPaths:   []string{"/factory/raw-materials", "/factory/raw-materials/orders", "/factory/raw-materials/receipts", "/factory/returns"},
		},
		{
			Name:        "SPV Gudang Pabrik (Full)",
			Description: "Hak akses menyeluruh untuk seluruh fitur, stok, alur pengiriman, dan manajemen operasional pabrik.",
			MenuPaths: []string{
				"/factory",
				"/factory/stock",
				"/factory/transfers",
				"/factory/returns",
				"/factory/raw-materials",
				"/factory/raw-materials/orders",
				"/factory/raw-materials/receipts",
				"/factory/management",
				"/factory/warehouses",
				"/factory/products",
				"/vehicles",
			},
		},
	}

	for _, pr := range presetRoles {
		var role domain.AccessRole
		if err := db.Where("name = ?", pr.Name).First(&role).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				role = domain.AccessRole{
					Name:        pr.Name,
					Description: pr.Description,
				}
				if err := db.Create(&role).Error; err != nil {
					log.Printf("Error creating preset role %s: %v\n", pr.Name, err)
					continue
				}
			} else {
				continue
			}
		}

		var menus []domain.AppMenu
		if err := db.Where("path IN ?", pr.MenuPaths).Find(&menus).Error; err == nil && len(menus) > 0 {
			db.Model(&role).Association("Menus").Replace(menus)
		}
	}
}

func requiredEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("%s is required for trial seeding", key)
	}
	return value
}
