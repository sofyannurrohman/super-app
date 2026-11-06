package main

import (
	"context"
	"log"
	nethttp "net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
	"example.com/hris/backend/config"
	"example.com/hris/backend/internal/delivery/http"
	"example.com/hris/backend/internal/delivery/tcp"
	"example.com/hris/backend/internal/repository"
	"example.com/hris/backend/internal/scheduler"
	"example.com/hris/backend/internal/usecase"
	"example.com/hris/backend/internal/worker"
	aiPkg "example.com/hris/backend/pkg/ai"
	"example.com/hris/backend/pkg/gemini"
	"example.com/hris/backend/pkg/midtrans"
	"example.com/hris/backend/pkg/redisclient"
	"example.com/hris/backend/pkg/sse"
	"example.com/hris/backend/pkg/utils"
)

func main() {
	// Initialize Config
	cfg := config.LoadConfig()

	// Initialize Database (PostgreSQL — main HRIS)
	db := config.ConnectDB(cfg)

	// Initialize Redis
	rdb := redisclient.Connect()

	// Initialize SSE Hub (Redis Pub/Sub → connected admin browsers)
	sseHub := sse.NewHub(rdb)

	// Resolve the default company from this installation's seeded tenant.
	defaultCompanyID := os.Getenv("DEFAULT_COMPANY_ID")
	if defaultCompanyID == "" {
		var company struct{ ID string }
		if err := db.Table("companies").Select("id").Where("name = ?", os.Getenv("TRIAL_COMPANY_NAME")).Take(&company).Error; err != nil || company.ID == "" {
			log.Fatal("perusahaan trial belum ditemukan; jalankan seeder atau atur DEFAULT_COMPANY_ID")
		}
		defaultCompanyID = company.ID
	}

	// Setup Repositories
	userRepo := repository.NewUserRepository(db)
	employeeRepo := repository.NewEmployeeRepository(db)
	attendanceRepo := repository.NewAttendanceRepository(db)
	leaveRepo := repository.NewLeaveRepository(db)
	overtimeRepo := repository.NewOvertimeRepository(db)
	payrollRepo := repository.NewPayrollRepository(db)
	jobPositionRepo := repository.NewJobPositionRepository(db)
	payrollComponentRepo := repository.NewPayrollComponentRepository(db)
	leaveTypeRepo := repository.NewLeaveTypeRepository(db)
	shiftRepo := repository.NewShiftRepository(db)
	departmentRepo := repository.NewDepartmentRepository(db)
	companyRepo := repository.NewCompanyRepository(db)
	branchRepo := repository.NewBranchRepository(db)
	employeeShiftRepo := repository.NewEmployeeShiftRepository(db)
	reimbursementRepo := repository.NewReimbursementRepository(db)
	performanceRepo := repository.NewPerformanceRepository(db)
	payrollConfigRepo := repository.NewPayrollConfigRepository(db)
	announcementRepo := repository.NewAnnouncementRepository(db)
	storeRepo := repository.NewStoreRepository(db)
	salesDailyJourneyRepo := repository.NewSalesDailyJourneyRepository(db)

	salesRepo := repository.NewSalesTransactionRepository(db)
	bannerOrderRepo := repository.NewBannerOrderRepository(db)
	factoryRepo := repository.NewFactoryRepository(db)
	warehouseRepo := repository.NewWarehouseRepository(db)
	vehicleRepo := repository.NewVehicleRepository(db)
	notificationRepo := repository.NewNotificationRepository(db)
	financeRepo := repository.NewFinanceRepository(db)
	bankAccountRepo := repository.NewBankAccountRepository(db)
	deliveryRepo := repository.NewDeliveryRepository(db)
	salesTransferRepo := repository.NewSalesTransferRepository(db)
	salesOrderRepo := repository.NewSalesOrderRepository(db)
	salesReturnRepo := repository.NewSalesReturnRepository(db)
	salesVisitRepo := repository.NewSalesVisitRepository(db)
	salesVisitPlanRepo := repository.NewSalesVisitPlanRepository(db)
	factoryReturnRepo := repository.NewFactoryReturnRepository(db)
	auditLogRepo := repository.NewAuditLogRepository(db)
	epkwtRepo := repository.NewEpkwtRepository(db)
	productExchangeRepo := repository.NewProductExchangeRepository(db)
	warehouseDeliveryOrderRepo := repository.NewWarehouseDeliveryOrderRepository(db)
	salesTesterRequestRepo := repository.NewSalesTesterRequestRepository(db)
	supplierRepo := repository.NewSupplierRepository(db)
	mtDeliveryRepo := repository.NewMTDeliveryRepository(db)
	marketplaceRepo := repository.NewMarketplaceRepository(db)
	marketplaceInventoryRepo := repository.NewMarketplaceInventoryRepository(db)
	payslipRequestRepo := repository.NewPayslipRequestRepository(db)
	memoRequestRepo := repository.NewMemoRequestRepository(db)
	employeeRequestRepo := repository.NewEmployeeRequestRepository(db)
	accessRoleRepo := repository.NewAccessRoleRepository(db)
	bonusInventoryRepo := repository.NewBonusInventoryRepository(db)
	salesBonusRepo := repository.NewSalesBonusRepository(db)
	accountingRepo := repository.NewAccountingRepository(db)
	branchDepositRepo := repository.NewBranchDepositRepository(db)
	kasPabrikRepo := repository.NewKasPabrikRepository(db)
	rawMaterialPORepo := repository.NewRawMaterialPORepository(db)
	technicianInventoryRepo := repository.NewTechnicianInventoryRepository(db)
	transportRepo := repository.NewTransportRepository(db)

	// Start License Expiry Scheduler
	licenseExpiryScheduler := scheduler.NewLicenseExpiryScheduler(
		factoryRepo,
		notificationRepo,
		sseHub,
		24*time.Hour,
	)
	if moduleEnabled("factory") {
		licenseExpiryScheduler.Start(context.Background())
	}

	// Setup Utils
	emailSender := utils.NewEmailSender(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUser, cfg.SMTPPass, cfg.SMTPFrom)
	midtransClient := midtrans.NewMidtransClient()

	// Start Redis Background Email Worker for Expiry Alerts
	go worker.StartEmailWorker(rdb, db, emailSender)

	// Initialize and Start Cron Jobs for Expirations
	c := cron.New()
	notificationUC := usecase.NewNotificationUsecase(db, rdb)
	// Run every day at 08:00
	_, err := c.AddFunc("0 8 * * *", func() {
		log.Println("[Cron] Running CheckAndQueueExpirations...")
		notificationUC.CheckAndQueueExpirations()
	})
	if err != nil {
		log.Printf("Failed to schedule CheckAndQueueExpirations: %v", err)
	}

	// Shopee Auto Boost Worker — Run every 5 minutes
	shopeeBoostHandler := http.NewShopeeBoostHandler(db)
	if moduleEnabled("marketplace") {
		_, boostCronErr := c.AddFunc("*/5 * * * *", func() {
			shopeeBoostHandler.RunAutoBoostForAllShops()
		})
		if boostCronErr != nil {
			log.Printf("Failed to schedule ShopeeAutoBoost: %v", boostCronErr)
		}
	}

	c.Start()

	// Setup UseCases
	fcmUseCase := usecase.NewFcmUsecase(db, os.Getenv("FCM_CREDENTIALS_FILE"))
	go fcmUseCase.StartReminderWorker(context.Background())

	authUseCase := usecase.NewAuthUseCase(userRepo, companyRepo, employeeShiftRepo, emailSender, fcmUseCase)
	employeeUseCase := usecase.NewEmployeeUsecase(employeeRepo, userRepo)
	attendanceUseCase := usecase.NewAttendanceUseCase(attendanceRepo, employeeRepo, employeeShiftRepo, leaveRepo)
	leaveUseCase := usecase.NewLeaveUseCase(db, leaveRepo, employeeRepo, fcmUseCase, sseHub)
	overtimeUseCase := usecase.NewOvertimeUseCase(overtimeRepo, employeeRepo, fcmUseCase, sseHub)
	payrollUseCase := usecase.NewPayrollUseCase(payrollRepo, employeeRepo, payrollConfigRepo, performanceRepo)
	jobPositionUseCase := usecase.NewJobPositionUseCase(jobPositionRepo)
	payrollComponentUseCase := usecase.NewPayrollComponentUseCase(payrollComponentRepo)
	leaveTypeUseCase := usecase.NewLeaveTypeUseCase(leaveTypeRepo)
	shiftUseCase := usecase.NewShiftUseCase(shiftRepo)
	departmentUseCase := usecase.NewDepartmentUseCase(departmentRepo)
	companyUseCase := usecase.NewCompanyUseCase(companyRepo)
	branchUseCase := usecase.NewBranchUseCase(branchRepo)
	employeeShiftUseCase := usecase.NewEmployeeShiftUseCase(employeeShiftRepo, employeeRepo, shiftRepo)
	reimbursementUseCase := usecase.NewReimbursementUseCase(reimbursementRepo, employeeRepo, fcmUseCase, sseHub)
	performanceUseCase := usecase.NewPerformanceUseCase(performanceRepo, attendanceRepo, employeeShiftRepo, leaveRepo, attendanceUseCase)
	payrollConfigUseCase := usecase.NewPayrollConfigUseCase(payrollConfigRepo)
	announcementUseCase := usecase.NewAnnouncementUseCase(announcementRepo, employeeRepo, fcmUseCase, sseHub)
	salesOrderUsecase := usecase.NewSalesOrderUsecase(salesOrderRepo, warehouseRepo, deliveryRepo, salesRepo, db, midtransClient, sseHub, fcmUseCase)
	salesDailyJourneyUseCase := usecase.NewSalesDailyJourneyUseCase(salesDailyJourneyRepo, db)

	salesUseCase := usecase.NewSalesUsecase(salesRepo, performanceRepo, storeRepo, attendanceRepo, companyRepo, employeeRepo, salesOrderUsecase, salesVisitRepo, salesVisitPlanRepo, salesTransferRepo, midtransClient, db, sseHub)
	bannerOrderUseCase := usecase.NewBannerOrderUseCase(bannerOrderRepo)
	factoryUseCase := usecase.NewFactoryUsecase(factoryRepo, db, sseHub)
	warehouseUseCase := usecase.NewWarehouseUsecase(warehouseRepo, notificationRepo, salesRepo, salesOrderRepo, db, sseHub)
	vehicleUseCase := usecase.NewVehicleUsecase(vehicleRepo)
	transportUsecase := usecase.NewTransportUsecase(transportRepo, vehicleRepo, db)
	financeUsecase := usecase.NewFinanceUsecase(financeRepo, salesUseCase, db, fcmUseCase)
	bankAccountUsecase := usecase.NewBankAccountUsecase(bankAccountRepo)
	deliveryUsecase := usecase.NewDeliveryUsecase(deliveryRepo, salesRepo, salesOrderRepo, branchRepo, warehouseRepo, db, sseHub, fcmUseCase)
	auditLogUseCase := usecase.NewAuditLogUsecase(auditLogRepo)
	warehouseDeliveryOrderUsecase := usecase.NewWarehouseDeliveryOrderUsecase(warehouseDeliveryOrderRepo, deliveryRepo, auditLogUseCase, db)
	salesTransferUsecase := usecase.NewSalesTransferUsecase(salesTransferRepo, warehouseRepo, db, fcmUseCase, warehouseDeliveryOrderUsecase)
	salesReturnUsecase := usecase.NewSalesReturnUsecase(salesReturnRepo, salesOrderRepo, salesRepo, warehouseRepo, db, sseHub)
	factoryReturnUsecase := usecase.NewFactoryReturnUsecase(factoryReturnRepo, db, sseHub)
	geminiProvider := aiPkg.GeminiProvider{Client: gemini.NewClient(cfg.GeminiAPIKey, cfg.GeminiModel)}
	var macProvider aiPkg.Provider
	if cfg.MacMiniLLMURL != "" {
		timeout, _ := time.ParseDuration(cfg.MacMiniLLMTimeout)
		if timeout <= 0 {
			timeout = 25 * time.Second
		}
		macProvider = aiPkg.MacMiniProvider{
			URL:     cfg.MacMiniLLMURL,
			Model:   cfg.MacMiniLLMModel,
			APIKey:  cfg.MacMiniLLMAPIKey,
			Timeout: timeout,
		}
		log.Printf("[AI] provider utama: Mac Mini (%s, model %s)", cfg.MacMiniLLMURL, cfg.MacMiniLLMModel)
	} else {
		log.Printf("[AI] MAC_MINI_LLM_URL kosong: semua permintaan AI memakai Gemini")
	}
	aiHandler := http.NewAIHandler(aiPkg.NewRouter(macProvider, geminiProvider))
	epkwtUseCase := usecase.NewEpkwtUseCase(epkwtRepo, employeeRepo)
	productExchangeUsecase := usecase.NewProductExchangeUsecase(productExchangeRepo, warehouseRepo, db)
	salesTesterRequestUsecase := usecase.NewSalesTesterRequestUsecase(salesTesterRequestRepo, db)
	supplierUsecase := usecase.NewSupplierUsecase(supplierRepo)
	rawMaterialPOUsecase := usecase.NewRawMaterialPOUsecase(rawMaterialPORepo, db)
	mtDeliveryUsecase := usecase.NewMTDeliveryUsecase(mtDeliveryRepo, db, fcmUseCase)
	marketplaceUsecase := usecase.NewMarketplaceUsecase(marketplaceRepo, auditLogUseCase, defaultCompanyID)
	marketplaceInventoryUsecase := usecase.NewMarketplaceInventoryUsecase(marketplaceInventoryRepo, auditLogUseCase, defaultCompanyID)
	payslipRequestUseCase := usecase.NewPayslipRequestUsecase(payslipRequestRepo, employeeRepo, payrollConfigRepo, payrollRepo, db)
	memoRequestUseCase := usecase.NewMemoRequestUseCase(memoRequestRepo)
	employeeRequestUseCase := usecase.NewEmployeeRequestUseCase(employeeRequestRepo, employeeRepo)
	accessRoleUseCase := usecase.NewAccessRoleUsecase(accessRoleRepo)
	bonusInventoryUsecase := usecase.NewBonusInventoryUsecase(bonusInventoryRepo, db)
	salesBonusUsecase := usecase.NewSalesBonusUsecase(salesBonusRepo, bonusInventoryRepo, db)
	glPostingUsecase := usecase.NewGLPostingUsecase(accountingRepo, companyRepo, db)
	branchDepositUsecase := usecase.NewBranchDepositUsecase(branchDepositRepo, glPostingUsecase, branchRepo, salesRepo)
	taxUsecase := usecase.NewTaxUsecase(db)
	biDashboardUsecase := usecase.NewBIDashboardUsecase(db)
	arAPUsecase := usecase.NewARAPUsecase(db)
	fixedAssetUsecase := usecase.NewFixedAssetUsecase(db)
	kasPabrikUsecase := usecase.NewKasPabrikUsecase(kasPabrikRepo, db)
	technicianInventoryUsecase := usecase.NewTechnicianInventoryUsecase(technicianInventoryRepo, db)
	autoInactiveUseCase := usecase.NewAutoInactiveUsecase(db, employeeRepo, userRepo, attendanceRepo, leaveRepo, auditLogRepo, notificationRepo, fcmUseCase)

	// Schedule Auto-Inactive Check every day at 01:00 WIB
	_, err = c.AddFunc("0 1 * * *", func() {
		log.Println("[Cron] Running Auto-Inactive Employee Evaluation (10 consecutive days missing)...")
		summary, err := autoInactiveUseCase.EvaluateAndDeactivateInactiveEmployees(context.Background(), false)
		if err != nil {
			log.Printf("[Cron Error] Auto-Inactive Evaluation failed: %v", err)
		} else {
			log.Printf("[Cron Success] Auto-Inactive Evaluation completed. Evaluated: %d, Deactivated: %d",
				summary.EvaluatedCount, summary.DeactivatedCount)
		}
	})
	if err != nil {
		log.Printf("Failed to schedule Auto-Inactive Employee Evaluation: %v", err)
	}

	// Inject GL posting ke reimbursement usecase untuk auto-posting saat approve
	if ru, ok := reimbursementUseCase.(usecase.GLSettable); ok {
		ru.SetGLPosting(glPostingUsecase)
	}

	// FCM Handler untuk admin (broadcast, kirim ke user tertentu, trigger reminder)
	fcmHandler := http.NewFcmHandler(fcmUseCase)

	visitPlanScheduler := scheduler.NewVisitPlanScheduler(db, fcmUseCase)
	if moduleEnabled("sales") {
		visitPlanScheduler.Start(context.Background())
	}

	// Initialize Gin

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.MaxMultipartMemory = 32 << 20 // 32 MiB (RAM safe: larger payloads stream to disk temp files)

	// CORS Middleware
	r.Use(cors.New(cors.Config{
		AllowOriginFunc: func(origin string) bool {
			return true // Allow all origins dynamically to support AllowCredentials
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Gzip Middleware for slow network optimization
	r.Use(http.GzipMiddleware())

	// Static Media Cache-Control Middleware untuk optimasi Cloudflare Edge CDN & byte-range streaming
	staticCacheMiddleware := func(c *gin.Context) {
		c.Header("Cache-Control", "public, max-age=2592000, immutable")
		c.Header("Accept-Ranges", "bytes")
		c.Next()
	}

	// Serve Static Files
	r.Group("/uploads", staticCacheMiddleware).Static("", "./uploads")

	// Initialize Handlers
	authHandler := http.NewAuthHandler(authUseCase, cfg)
	employeeHandler := http.NewEmployeeHandler(employeeUseCase, autoInactiveUseCase, db)
	attendanceHandler := http.NewAttendanceHandler(attendanceUseCase)
	leaveHandler := http.NewLeaveHandler(leaveUseCase)
	userHandler := http.NewUserHandler(db, employeeRepo)
	overtimeHandler := http.NewOvertimeHandler(overtimeUseCase)
	payrollHandler := http.NewPayrollHandler(payrollUseCase)
	jobPositionHandler := http.NewJobPositionHandler(jobPositionUseCase)
	payrollComponentHandler := http.NewPayrollComponentHandler(payrollComponentUseCase)
	leaveTypeHandler := http.NewLeaveTypeHandler(leaveTypeUseCase)
	shiftHandler := http.NewShiftHandler(shiftUseCase)
	departmentHandler := http.NewDepartmentHandler(departmentUseCase)
	companyHandler := http.NewCompanyHandler(companyUseCase)
	branchHandler := http.NewBranchHandler(branchUseCase)
	employeeShiftHandler := http.NewEmployeeShiftHandler(employeeShiftUseCase)
	reimbursementHandler := http.NewReimbursementHandler(reimbursementUseCase)
	performanceHandler := http.NewPerformanceHandler(performanceUseCase, employeeUseCase)
	payrollConfigHandler := http.NewPayrollConfigHandler(payrollConfigUseCase)
	geminiClient := gemini.NewClient(cfg.GeminiAPIKey, cfg.GeminiModel)

	// Initialize GDrive Service untuk unggah media pengumuman & banner storage
	var gdriveService *utils.GDriveService
	if credentialsFile, folderID := os.Getenv("GDRIVE_CREDENTIALS_FILE"), os.Getenv("GDRIVE_FOLDER_ID"); credentialsFile != "" && folderID != "" {
		var gdriveErr error
		gdriveService, gdriveErr = utils.NewGDriveService(credentialsFile, folderID)
		if gdriveErr != nil {
			log.Printf("[Main] GDrive Service gagal diinisialisasi: %v", gdriveErr)
		} else {
			log.Println("[Main] GDrive Service berhasil diinisialisasi")
		}
	}

	announcementHandler := http.NewAnnouncementHandler(announcementUseCase, gdriveService)
	storeHandler := http.NewStoreHandler(storeRepo, employeeUseCase, auditLogUseCase, geminiClient)
	salesHandler := http.NewSalesHandler(salesUseCase, employeeUseCase, attendanceRepo, auditLogUseCase, sseHub, fcmUseCase, db)
	dailyKpiReportRepo := repository.NewDailyKpiReportRepository(db)
	dailyKpiReportUsecase := usecase.NewDailyKpiReportUsecase(dailyKpiReportRepo, employeeRepo, db)
	dailyKpiReportHandler := http.NewDailyKpiReportHandler(dailyKpiReportUsecase)
	bannerOrderHandler := http.NewBannerOrderHandler(bannerOrderUseCase, employeeUseCase)
	factoryHandler := http.NewFactoryHandler(factoryUseCase)
	rawMaterialPOHandler := http.NewRawMaterialPOHandler(rawMaterialPOUsecase)
	warehouseHandler := http.NewWarehouseHandler(warehouseUseCase, db)
	vehicleHandler := http.NewVehicleHandler(vehicleUseCase)
	transportHandler := http.NewTransportHandler(transportUsecase)

	gpsRepo := repository.NewGpsRepository(db)
	gpsUseCase := usecase.NewGpsUsecase(gpsRepo)
	gpsHandler := http.NewGpsHandler(gpsUseCase)

	gpsCleanerScheduler := scheduler.NewGpsCleanerScheduler(gpsUseCase, 24*time.Hour)
	if moduleEnabled("rajatransport") {
		gpsCleanerScheduler.Start(context.Background())
		go func() {
			gt06Server := tcp.NewGT06Server("9000", gpsUseCase)
			if err := gt06Server.Start(); err != nil {
				log.Printf("Failed to start GT06 TCP server: %v", err)
			}
		}()
	}

	notificationHandler := http.NewNotificationHandler(notificationRepo)
	sseHandler := http.NewSSEHandler(sseHub)
	financeHandler := http.NewFinanceHandler(financeUsecase)
	bankAccountHandler := http.NewBankAccountHandler(bankAccountUsecase)
	deliveryHandler := http.NewDeliveryHandler(deliveryUsecase, employeeUseCase)
	salesDailyJourneyHandler := http.NewSalesDailyJourneyHandler(salesDailyJourneyUseCase)

	salesTransferHandler := http.NewSalesTransferHandler(salesTransferUsecase, salesUseCase)
	salesOrderHandler := http.NewSalesOrderHandler(salesOrderUsecase, employeeUseCase, auditLogUseCase)
	salesReturnHandler := http.NewSalesReturnHandler(salesReturnUsecase, employeeUseCase)
	factoryReturnHandler := http.NewFactoryReturnHandler(factoryReturnUsecase)
	productExchangeHandler := http.NewProductExchangeHandler(productExchangeUsecase, employeeUseCase)
	warehouseDeliveryOrderHandler := http.NewWarehouseDeliveryOrderHandler(warehouseDeliveryOrderUsecase)
	salesTesterRequestHandler := http.NewSalesTesterRequestHandler(salesTesterRequestUsecase, employeeUseCase)
	// Branding app integration — 100% PostgreSQL independent
	brandingProxyHandler := http.NewBrandingProxyHandler(nil, db, employeeUseCase, sseHub)
	auditLogHandler := http.NewAuditLogHandler(auditLogUseCase)
	epkwtHandler := http.NewEpkwtHandler(epkwtUseCase)
	modernTradeHandler := http.NewModernTradeHandler(db)
	supplierHandler := http.NewSupplierHandler(supplierUsecase)
	mtDeliveryHandler := http.NewMTDeliveryHandler(mtDeliveryUsecase)
	marketplaceHandler := http.NewMarketplaceHandler(marketplaceUsecase, db)
	marketplaceInventoryHandler := http.NewMarketplaceInventoryHandler(marketplaceInventoryUsecase)
	payslipRequestHandler := http.NewPayslipRequestHandler(payslipRequestUseCase)
	memoRequestHandler := http.NewMemoRequestHandler(memoRequestUseCase)
	employeeRequestHandler := http.NewEmployeeRequestHandler(employeeRequestUseCase, employeeRepo)
	accessRoleHandler := http.NewAccessRoleHandler(accessRoleUseCase)
	bonusInventoryHandler := http.NewBonusInventoryHandler(bonusInventoryUsecase)
	salesBonusHandler := http.NewSalesBonusHandler(salesBonusUsecase)
	dailyWorkerHandler := http.NewDailyWorkerHandler(db)
	accountingHandler := http.NewAccountingHandler(glPostingUsecase)
	branchDepositHandler := http.NewBranchDepositHandler(branchDepositUsecase)
	taxHandler := http.NewTaxHandler(taxUsecase)
	biDashboardHandler := http.NewBIDashboardHandler(biDashboardUsecase)
	financeExtendedHandler := http.NewFinanceExtendedHandler(arAPUsecase, fixedAssetUsecase)
	proyeksiOmzetHandler := http.NewProyeksiOmzetHandler(db)
	directorDashboardHandler := http.NewDirectorDashboardHandler(db)
	cashierReportHandler := http.NewCashierReportHandler(db, glPostingUsecase)
	cashBankMutationHandler := http.NewCashBankMutationHandler(db)
	kasPabrikHandler := http.NewKasPabrikHandler(kasPabrikUsecase)
	technicianInventoryHandler := http.NewTechnicianInventoryHandler(technicianInventoryUsecase)

	employeeWarningUsecase := usecase.NewEmployeeWarningUsecase(db, fcmUseCase)
	employeeWarningHandler := http.NewEmployeeWarningHandler(employeeWarningUsecase)

	// Internal Report Handler — diakses oleh WA Report Service di Mac Mini M4
	// Diproteksi oleh X-Internal-Key header, BUKAN JWT karyawan
	internalReportHandler := http.NewInternalReportHandler(db)

	// LPS Handler — Lembar Penjualan Sales (laporan harian untuk admin gudang, nota, dan kasir)
	lpsHandler := http.NewLPSHandler(db, salesUseCase)

	marketplaceIntegrationHandler := http.NewMarketplaceIntegrationHandler(db)
	tiktokWebhookHandler := http.NewTikTokWebhookHandler(db)
	wmsIngestionHandler := http.NewWMSIngestionHandler(db)
	wmsFulfillmentHandler := http.NewWMSFulfillmentHandler(db)
	tiktokSyncHandler := http.NewTikTokSyncHandler(db)
	imageProxyHandler := http.NewImageProxyHandler()
	hrDashboardHandler := http.NewHRDashboardHandler(db)

	distributorSalesUsecase := usecase.NewDistributorSalesUsecase(db)
	distributorSalesHandler := http.NewDistributorSalesHandler(distributorSalesUsecase)

	// API v1 Routes
	v1 := r.Group("/api/v1")
	{
		// Public Routes
		authHandler.SetupRoutes(v1)
		branchHandler.SetupPublicRoutes(v1)
		attendanceHandler.SetupPublicRoutes(v1)
		jobPositionHandler.SetupPublicRoutes(v1)
		shiftHandler.SetupPublicRoutes(v1)
		salesHandler.SetupPublicRoutes(v1)
		deliveryHandler.SetupPublicRoutes(v1)
		marketplaceIntegrationHandler.RegisterRoutes(v1)
		tiktokWebhookHandler.RegisterRoutes(v1)
		imageProxyHandler.SetupPublicRoutes(v1)
		v1.Group("/uploads", staticCacheMiddleware).Static("", "./uploads")

		// Internal Routes — hanya untuk service-to-service (WA Report di Mac Mini)
		internalReportHandler.RegisterInternalRoutes(v1)

		// Protected Routes
		protected := v1.Group("")
		protected.Use(http.AuthMiddleware(cfg.JWTSecret, db))
		protected.Use(http.EmployeeMiddleware(employeeUseCase))
		protected.Use(http.BranchFilterMiddleware(db))
		wmsIngestionHandler.RegisterRoutes(protected)
		wmsFulfillmentHandler.RegisterRoutes(protected)
		tiktokSyncHandler.RegisterRoutes(protected)
		{
			attendanceHandler.SetupRoutes(protected)
			employeeHandler.SetupRoutes(protected)
			leaveHandler.SetupRoutes(protected)
			userHandler.SetupRoutes(protected)
			overtimeHandler.SetupRoutes(protected)
			payrollHandler.SetupRoutes(protected)
			jobPositionHandler.SetupRoutes(protected)
			payrollComponentHandler.SetupRoutes(protected)
			leaveTypeHandler.SetupRoutes(protected)
			shiftHandler.SetupRoutes(protected)
			departmentHandler.SetupRoutes(protected)
			companyHandler.SetupRoutes(protected)
			branchHandler.SetupRoutes(protected)
			employeeShiftHandler.SetupRoutes(protected)
			reimbursementHandler.SetupRoutes(protected)
			performanceHandler.SetupRoutes(protected)
			taxHandler.SetupRoutes(protected)
			hrDashboardHandler.RegisterRoutes(protected)
			biDashboardHandler.RegisterRoutes(protected)
			financeExtendedHandler.RegisterRoutes(protected)
			payrollConfigHandler.SetupRoutes(protected)
			announcementHandler.SetupRoutes(protected)
			storeHandler.SetupRoutes(protected)
			salesHandler.SetupMobileRoutes(protected)
			salesHandler.RegisterRoutes(protected)
			proyeksiOmzetHandler.RegisterRoutes(protected)
			directorDashboardHandler.RegisterRoutes(protected)
			bannerOrderHandler.SetupRoutes(protected)
			factoryHandler.RegisterRoutes(protected)
			rawMaterialPOHandler.RegisterRoutes(protected)
			technicianInventoryHandler.RegisterRoutes(protected)
			warehouseHandler.RegisterRoutes(protected)
			vehicleHandler.RegisterRoutes(protected)
			transportHandler.RegisterRoutes(protected)
			gpsHandler.RegisterRoutes(protected)
			notificationHandler.RegisterRoutes(protected)
			sseHandler.RegisterRoutes(protected)
			financeHandler.RegisterRoutes(protected)
			bankAccountHandler.RegisterRoutes(protected)
			deliveryHandler.RegisterRoutes(protected)
			salesDailyJourneyHandler.RegisterRoutes(protected)

			salesTransferHandler.RegisterRoutes(protected)
			salesOrderHandler.RegisterRoutes(protected)
			salesOrderHandler.SetupMobileRoutes(protected)
			salesReturnHandler.RegisterRoutes(protected)
			factoryReturnHandler.RegisterRoutes(protected)
			productExchangeHandler.RegisterRoutes(protected)
			warehouseDeliveryOrderHandler.RegisterRoutes(protected)
			salesTesterRequestHandler.RegisterRoutes(protected)
			brandingProxyHandler.RegisterRoutes(protected)
			aiHandler.RegisterRoutes(protected)
			auditLogHandler.SetupRoutes(protected)
			epkwtHandler.RegisterRoutes(protected)
			modernTradeHandler.RegisterRoutes(protected)
			supplierHandler.RegisterRoutes(protected)
			mtDeliveryHandler.RegisterRoutes(protected)
			marketplaceHandler.RegisterRoutes(protected)
			marketplaceInventoryHandler.RegisterRoutes(protected)
			shopeeBoostHandler.RegisterRoutes(protected)
			payslipRequestHandler.SetupRoutes(protected)
			memoRequestHandler.SetupRoutes(protected)
			employeeRequestHandler.SetupRoutes(protected)
			accessRoleHandler.SetupRoutes(protected)
			bonusInventoryHandler.SetupRoutes(protected)
			salesBonusHandler.SetupRoutes(protected)
			lpsHandler.RegisterRoutes(protected)
			fcmHandler.RegisterAdminRoutes(protected)
			dailyWorkerHandler.SetupRoutes(protected)
			employeeWarningHandler.RegisterRoutes(protected)
			accountingHandler.RegisterRoutes(protected)
			branchDepositHandler.RegisterRoutes(protected)
			cashierReportHandler.SetupRoutes(protected)
			cashBankMutationHandler.SetupRoutes(protected)
			kasPabrikHandler.RegisterRoutes(protected)
			distributorSalesHandler.SetupRoutes(protected)
			dailyKpiReportHandler.SetupRoutes(protected)
		}
	}

	// Health Check
	healthHandler := func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	}
	r.GET("/api/v1/health", healthHandler)
	r.HEAD("/api/v1/health", healthHandler)

	log.Printf("Server starting on port %s", cfg.ServerPort)

	srv := &nethttp.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != nethttp.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with a timeout of 10 seconds.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
}

func moduleEnabled(name string) bool {
	modules := os.Getenv("ENABLED_MODULES")
	if modules == "" {
		modules = os.Getenv("VITE_ENABLED_MODULES")
	}
	if modules == "" {
		modules = "hris"
	}
	for _, module := range strings.Split(modules, ",") {
		if strings.EqualFold(strings.TrimSpace(module), name) {
			return true
		}
	}
	return false
}
