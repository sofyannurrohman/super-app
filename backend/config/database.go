package config

import (
	"embed"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"example.com/hris/backend/internal/domain"
	"github.com/golang-migrate/migrate/v4"
	migrate_postgres "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

func ConnectDB(cfg Config) *gorm.DB {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Jakarta",
		cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort)

	gormLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  logger.Warn,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)

	var db *gorm.DB
	var err error
	maxRetries := 5
	for i := 0; i < maxRetries; i++ {
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: gormLogger,
		})
		if err == nil {
			break
		}
		log.Printf("Failed to connect to database (attempt %d/%d). Retrying in 2 seconds... \n", i+1, maxRetries)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		log.Fatal("Failed to connect to database after maximum retries. \n", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("Failed to get DB instance. \n", err)
	}

	// Optimasi connection pool untuk beban tinggi (Web Dashboard + Puluhan Sales Mobile App)
	sqlDB.SetMaxIdleConns(30)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)
	sqlDB.SetConnMaxIdleTime(5 * time.Minute) // bebaskan koneksi idle setelah 5 menit

	log.Println("Connected Successfully to Database")

	// Historical SQL migrations reference vehicles before the seeder's GORM
	// AutoMigrate creates it. Bootstrap the referenced key for fresh installs;
	// AutoMigrate adds the remaining Vehicle fields after migrations finish.
	if err := db.Exec(`CREATE TABLE IF NOT EXISTS vehicles (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid()
	)`).Error; err != nil {
		log.Fatalf("Failed to bootstrap vehicles table: %v", err)
	}

	// Database Migrations using golang-migrate
	log.Println("Running Migrations...")

	d, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		log.Fatal("Failed to load embedded migrations: \n", err)
	}

	driver, err := migrate_postgres.WithInstance(sqlDB, &migrate_postgres.Config{})
	if err != nil {
		log.Fatal("Failed to create migration driver: \n", err)
	}

	m, err := migrate.NewWithInstance(
		"iofs",
		d,
		"postgres",
		driver,
	)
	if err != nil {
		log.Fatal("Failed to initialize migration instance: \n", err)
	}

	// The trial starts from an empty database. Several historical SQL migrations
	// refer to tables that the original installation created with GORM between
	// migration runs. Create those models after the SQL base schema (version 60)
	// and before the later migrations that depend on them. This also repairs a
	// trial database left dirty at version 61 by an earlier seeder attempt.
	if cfg.DBName == "hris_trial" {
		version, dirty, versionErr := m.Version()
		if versionErr != nil && versionErr != migrate.ErrNilVersion {
			log.Fatalf("Failed to read trial migration version: %v", versionErr)
		}
		// Versions 61 and 110 failed inside transactional SQL migrations in
		// this trial rollout; their corrected migrations can be retried.
		if dirty && version != 61 && version != 110 {
			log.Fatalf("Trial migration %d is dirty; inspect it before retrying", version)
		}
		if versionErr == migrate.ErrNilVersion || (version < 60 && !dirty) {
			if err := m.Migrate(60); err != nil && err != migrate.ErrNoChange {
				log.Fatalf("Failed to migrate trial base schema to version 60: %v", err)
			}
		}
		if versionErr == migrate.ErrNilVersion || version <= 61 {
			// Migration 102 adds these columns without IF NOT EXISTS, so create
			// only the key here and let that migration add them.
			if err := db.Exec(`CREATE TABLE IF NOT EXISTS branch_deposits (
				id UUID PRIMARY KEY DEFAULT gen_random_uuid()
			)`).Error; err != nil {
				log.Fatalf("Failed to bootstrap trial branch deposits: %v", err)
			}
			bootstrapDB := db.Session(&gorm.Session{})
			bootstrapDB.Config.IgnoreRelationshipsWhenMigrating = true
			bootstrapDB.Config.DisableForeignKeyConstraintWhenMigrating = true
			if err := bootstrapDB.AutoMigrate(
				&domain.Vehicle{},
				&domain.SalesTransfer{},
				&domain.BonusInventory{},
				&domain.SalesBonusRequest{},
				&domain.SalesItem{},
				&domain.SalesInstallment{},
				&domain.SalesOrderBonusItem{},
				&domain.SalesTransactionBonusItem{},
				&domain.MarketplaceToken{},
				&domain.KasPabrik{},
				&domain.KasTransaction{},
				&domain.CompanyBankAccount{},
				&domain.CashBankMutation{},
			); err != nil {
				log.Fatalf("Failed to bootstrap trial migration prerequisites: %v", err)
			}
		}
	}

	// db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\";") is handled by init_schema.up.sql

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		if strings.Contains(err.Error(), "Dirty database version") {
			log.Println("WARNING: Dirty migration detected. Attempting auto-recovery by forcing previous version...")
			v, _, verErr := m.Version()
			if verErr == nil {
				if forceErr := m.Force(int(v) - 1); forceErr != nil {
					log.Fatalf("Failed to reset dirty migration version: %v", forceErr)
				}
				if retryErr := m.Up(); retryErr != nil && retryErr != migrate.ErrNoChange {
					if cfg.DBName == "hris_trial" {
						log.Fatalf("Trial migration failed after retry: %v", retryErr)
					}
					log.Printf("WARNING: Retry migration returned: %v. Continuing to AutoMigrate...", retryErr)
				} else {
					log.Println("Migrations completed successfully after recovering dirty version.")
				}
			}
		} else {
			if cfg.DBName == "hris_trial" {
				log.Fatalf("Trial migration failed: %v", err)
			}
			log.Printf("WARNING: Migration returned: %v. Continuing to AutoMigrate...", err)
		}
	} else {
		log.Println("Migrations completed")
	}

	// Retur faktur needs these columns even when an earlier dirty migration
	// prevents 000195/000196 from running. AutoMigrate may also be disabled.
	for _, statement := range []string{
		`ALTER TABLE sales_returns ADD COLUMN IF NOT EXISTS return_type VARCHAR(30) DEFAULT 'POTONG_NOTA'`,
		`ALTER TABLE sales_returns ADD COLUMN IF NOT EXISTS invoice_adjusted_at TIMESTAMP`,
		`ALTER TABLE sales_return_items ADD COLUMN IF NOT EXISTS sales_item_id UUID`,
		`ALTER TABLE sales_return_items ADD COLUMN IF NOT EXISTS replacement_sales_item_id UUID`,
		`UPDATE sales_returns
		 SET invoice_adjusted_at = created_at
		 WHERE return_type = 'TUKAR_PRODUK'
		   AND transaction_id IS NOT NULL
		   AND COALESCE(notes, '') NOT LIKE '%[DELIVERY_RETURN]%'
		   AND invoice_adjusted_at IS NULL`,
	} {
		if err := db.Exec(statement).Error; err != nil {
			log.Fatalf("Failed to ensure sales return schema: %v", err)
		}
	}

	// A failed or dirty migration can leave 000188 or 000191 unapplied while
	// startup still continues. Every cash/bank mutation uses these model columns,
	// including ordinary entries that do not set a fund origin.
	for _, statement := range []string{
		`ALTER TABLE cash_bank_mutations
			ADD COLUMN IF NOT EXISTS destination_company_id UUID REFERENCES companies(id),
			ADD COLUMN IF NOT EXISTS status_setoran_note TEXT,
			ADD COLUMN IF NOT EXISTS status_setoran_by UUID,
			ADD COLUMN IF NOT EXISTS status_setoran_at TIMESTAMPTZ,
			ADD COLUMN IF NOT EXISTS fund_origin VARCHAR(40) NOT NULL DEFAULT ''`,
		`CREATE INDEX IF NOT EXISTS idx_cash_bank_mutations_destination_company_id
			ON cash_bank_mutations(destination_company_id)`,
	} {
		if err := db.Exec(statement).Error; err != nil {
			log.Fatalf("Failed to ensure cash/bank mutation schema: %v", err)
		}
	}

	// The migration runner currently logs failures and continues startup. Keep
	// this table available when an earlier dirty migration prevented 000189 from
	// running; the versioned migration remains the normal schema path.
	for _, statement := range []string{
		`CREATE TABLE IF NOT EXISTS marketplace_profitability_expenses (
			id BIGSERIAL PRIMARY KEY,
			branch_id UUID NOT NULL,
			marketplace VARCHAR(50) NOT NULL CHECK (marketplace IN ('tiktok', 'shopee')),
			expense_date DATE NOT NULL,
			category VARCHAR(30) NOT NULL CHECK (category IN ('ads', 'fulfillment', 'return', 'operational')),
			amount DECIMAL(15,2) NOT NULL CHECK (amount >= 0),
			attributed_revenue DECIMAL(15,2) NOT NULL DEFAULT 0 CHECK (attributed_revenue >= 0),
			seller_sku VARCHAR(100) NOT NULL DEFAULT '',
			notes TEXT NOT NULL DEFAULT '',
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_marketplace_profit_expenses_period
			ON marketplace_profitability_expenses (expense_date, branch_id, marketplace)`,
	} {
		if err := db.Exec(statement).Error; err != nil {
			log.Fatalf("Failed to ensure marketplace profitability schema: %v", err)
		}
	}

	// Invoice deposits are needed by the finance page even when deployments set
	// SKIP_AUTO_MIGRATE=true. Keep this idempotent bootstrap as a fallback for
	// installations whose migration runner was skipped or left behind a dirty version.
	for _, statement := range []string{
		`CREATE TABLE IF NOT EXISTS invoice_deposits (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			branch_id UUID NOT NULL,
			company_id UUID NOT NULL,
			bank_account_id UUID NOT NULL,
			deposit_date DATE NOT NULL,
			amount DECIMAL(15,2) NOT NULL,
			proof_url TEXT,
			notes TEXT,
			status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
			submitted_by_id UUID NOT NULL,
			reviewed_by_id UUID,
			reviewed_at TIMESTAMPTZ,
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS invoice_deposit_items (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			deposit_id UUID NOT NULL REFERENCES invoice_deposits(id) ON DELETE CASCADE,
			sales_transaction_id UUID NOT NULL,
			amount DECIMAL(15,2) NOT NULL,
			created_at TIMESTAMPTZ DEFAULT NOW(),
			CONSTRAINT idx_invoice_deposit_item UNIQUE (deposit_id, sales_transaction_id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_invoice_deposits_branch_id ON invoice_deposits(branch_id)`,
		`CREATE INDEX IF NOT EXISTS idx_invoice_deposits_company_id ON invoice_deposits(company_id)`,
		`CREATE INDEX IF NOT EXISTS idx_invoice_deposits_bank_account_id ON invoice_deposits(bank_account_id)`,
		`CREATE INDEX IF NOT EXISTS idx_invoice_deposits_status ON invoice_deposits(status)`,
		`CREATE INDEX IF NOT EXISTS idx_invoice_deposit_items_sales_transaction_id ON invoice_deposit_items(sales_transaction_id)`,
		`CREATE TABLE IF NOT EXISTS cash_bank_mutation_invoices (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			mutation_id UUID NOT NULL REFERENCES cash_bank_mutations(id) ON DELETE CASCADE,
			sales_payment_id UUID NOT NULL REFERENCES sales_payments(id),
			sales_transaction_id UUID NOT NULL REFERENCES sales_transactions(id),
			amount DECIMAL(15,2) NOT NULL,
			created_at TIMESTAMPTZ DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_cbm_invoice_mutation_id ON cash_bank_mutation_invoices(mutation_id)`,
		`CREATE INDEX IF NOT EXISTS idx_cbm_invoice_payment_id ON cash_bank_mutation_invoices(sales_payment_id)`,
		`CREATE INDEX IF NOT EXISTS idx_cbm_invoice_transaction_id ON cash_bank_mutation_invoices(sales_transaction_id)`,
	} {
		if err := db.Exec(statement).Error; err != nil {
			log.Fatalf("Failed to ensure invoice deposit schema: %v", err)
		}
	}

	// Run AutoMigrate only if not explicitly skipped
	if !cfg.SkipAutoMigrate {
		log.Println("Running AutoMigrate for Modern Trade and Sales Models...")

		if err := db.AutoMigrate(
			&domain.User{},
			&domain.Store{},
			&domain.KeyAccount{},
			&domain.MTMerchandising{},
			&domain.KasPabrik{},
			&domain.KasTransaction{},
			&domain.KasAdvanceSettlement{},
			&domain.BonusInventory{},
			&domain.BonusInventoryLog{},
			&domain.SalesKPI{},
			&domain.DailyKpiReport{},
			&domain.DailyKpiReportItem{},
			&domain.DailyKpiReportActionPlan{},
			&domain.SalesOrderBonusItem{},
			&domain.SalesDailyJourney{},
			&domain.Reimbursement{},

			&domain.SalesOrder{},
			&domain.SalesPayment{},
			&domain.LPSReport{},
			&domain.MTSpg{},
			&domain.MTSpgDailyReport{},
			&domain.MTSpgSaleItem{},
			&domain.MTSpgSchedule{},
			&domain.MTContract{},
			&domain.MTPurchaseOrder{},
			&domain.MTPromotion{},
			&domain.MTInventory{},
			&domain.MTInventoryLog{},
			&domain.MTStoreStock{},
			&domain.MTStoreStockLog{},
			&domain.WarehouseDeliveryOrder{},
			&domain.WarehouseDeliveryOrderItem{},
			&domain.WarehouseDeliveryOrderTransfer{},
			&domain.MarketplaceStock{},
			&domain.MarketplaceOpname{},
			&domain.MarketplaceOrder{},
			&domain.MarketplaceUnsettledSnapshot{},
			&domain.MarketplaceUnsettledTransaction{},
			&domain.MarketplaceProduct{},
			&domain.MarketplaceInventory{},
			&domain.MarketplaceProductMapping{},
			&domain.MarketplaceInventoryLog{},
			&domain.MarketplaceInventoryOpname{},
			&domain.MarketplaceScanHistory{},
			&domain.MarketplaceToken{},
			&domain.MarketplaceConfig{},
			&domain.Product{},
			&domain.BranchProductPrice{},
			&domain.PayslipRequest{},
			&domain.DeliveryBatch{}, // support batch lintas company: company_id nullable
			&domain.DeliveryItem{},
			&domain.DeliverySpare{},
			&domain.DeliveryBatchSample{},
			&domain.Order{},
			&domain.OrderItem{},
			&domain.WebhookLog{},
			&domain.EmployeeWarning{},
			&domain.BranchDeposit{},
			&domain.InvoiceDeposit{},
			&domain.InvoiceDepositItem{},
			&domain.DailyWorker{},
			&domain.DailyWorkerSalary{},
			&domain.CashierFinancialReport{},
			&domain.CompanyBankAccount{},
			&domain.CompanyBranch{},
			&domain.CashBankMutation{},
			&domain.CashBankMutationInvoice{},
			&domain.BranchCompanyBalance{},
			&domain.FixedAsset{},
			&domain.GpsDevice{},
			&domain.EpkwtRequest{},
			&domain.VehicleLocation{},
			&domain.SalesmanLiveLocation{},
			&domain.BrandingVendor{},
			&domain.BrandingCampaignVideo{},
			&domain.IndoProvince{},
			&domain.IndoCity{},
			&domain.IndoDistrict{},
			&domain.IndoVillage{},
			&domain.BannerOrder{},
			&domain.BrandingAIDesignJob{},
			&domain.Reimbursement{},
			&domain.RawMaterialPurchaseRequest{},
			&domain.RawMaterialPurchaseRequestItem{},
			&domain.RawMaterialPurchaseOrder{},
			&domain.RawMaterialPurchaseOrderItem{},
			&domain.RawMaterialReceipt{},
			&domain.RawMaterialReceiptItem{},
			&domain.ProductionLog{},
			&domain.ProductionLogMaterial{},
			&domain.FactoryInventoryLog{},
			&domain.FactoryQcSample{},
			&domain.EmployeeRequest{},
			&domain.EmployeeRequestCandidate{},
			&domain.TechnicianItem{},
			&domain.TechnicianStockLog{},
			&domain.Announcement{},
			&domain.AnnouncementTarget{},
			&domain.AnnouncementRead{},
			&domain.AnnouncementLike{},
			&domain.AnnouncementComment{},
		); err != nil {
			log.Printf("⚠️ AutoMigrate failed: %v", err)
		}

		// Ensure announcements columns for media & social media feed exist
		_ = db.Exec(`ALTER TABLE announcements ADD COLUMN IF NOT EXISTS media_type VARCHAR(50) DEFAULT 'none';`).Error
		_ = db.Exec(`ALTER TABLE announcements ADD COLUMN IF NOT EXISTS image_url TEXT;`).Error
		_ = db.Exec(`ALTER TABLE announcements ADD COLUMN IF NOT EXISTS media_urls TEXT[];`).Error
		_ = db.Exec(`ALTER TABLE announcements ADD COLUMN IF NOT EXISTS external_url TEXT;`).Error
		_ = db.Exec(`ALTER TABLE announcements ADD COLUMN IF NOT EXISTS likes_count INT DEFAULT 0;`).Error
		_ = db.Exec(`ALTER TABLE announcements ADD COLUMN IF NOT EXISTS comments_count INT DEFAULT 0;`).Error

		// Ensure announcement_likes table exists
		_ = db.Exec(`
			CREATE TABLE IF NOT EXISTS announcement_likes (
				id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
				announcement_id UUID NOT NULL,
				user_id UUID NOT NULL,
				created_at TIMESTAMPTZ DEFAULT now(),
				CONSTRAINT uq_announcement_user_like UNIQUE (announcement_id, user_id)
			);
		`).Error
		_ = db.Exec(`CREATE INDEX IF NOT EXISTS idx_announcement_likes_announcement_id ON announcement_likes(announcement_id);`).Error
		_ = db.Exec(`CREATE INDEX IF NOT EXISTS idx_announcement_likes_user_id ON announcement_likes(user_id);`).Error

		// Ensure announcement_comments table exists
		_ = db.Exec(`
			CREATE TABLE IF NOT EXISTS announcement_comments (
				id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
				announcement_id UUID NOT NULL,
				user_id UUID NOT NULL,
				content TEXT NOT NULL,
				created_at TIMESTAMPTZ DEFAULT now(),
				updated_at TIMESTAMPTZ DEFAULT now()
			);
		`).Error
		_ = db.Exec(`CREATE INDEX IF NOT EXISTS idx_announcement_comments_announcement_id ON announcement_comments(announcement_id);`).Error
		_ = db.Exec(`CREATE INDEX IF NOT EXISTS idx_announcement_comments_user_id ON announcement_comments(user_id);`).Error
		_ = db.Exec(`CREATE INDEX IF NOT EXISTS idx_announcement_comments_announcement_created ON announcement_comments(announcement_id, created_at ASC);`).Error

		// Ensure factory_inventory_logs source column is long enough
		_ = db.Exec(`ALTER TABLE factory_inventory_logs ALTER COLUMN source TYPE VARCHAR(255);`).Error
		_ = db.Exec(`ALTER TABLE warehouse_logs ALTER COLUMN source TYPE VARCHAR(255);`).Error
		_ = db.Exec(`ALTER TABLE sales_transfers ADD COLUMN IF NOT EXISTS batch_no VARCHAR(50) DEFAULT '';`).Error
		_ = db.Exec(`ALTER TABLE production_logs ADD COLUMN IF NOT EXISTS batch_group_id UUID;`).Error
		_ = db.Exec(`CREATE INDEX IF NOT EXISTS idx_production_logs_batch_group_id ON production_logs(batch_group_id);`).Error

		// Migrate existing mappings from seller_sku to sku_id to prevent mapping loss
		log.Println("Migrating marketplace_product_mappings from seller_sku to sku_id...")
		if err := db.Exec(`
			UPDATE marketplace_product_mappings mpm
			SET sku_id = mp.sku_id
			FROM marketplace_products mp
			WHERE mpm.seller_sku = mp.seller_sku
			  AND mpm.seller_sku IS NOT NULL
			  AND mpm.seller_sku != ''
			  AND (mpm.sku_id IS NULL OR mpm.sku_id = '')
		`).Error; err != nil {
			log.Printf("⚠️ Failed to migrate marketplace_product_mappings: %v", err)
		}

		// Delete orphaned mappings that have empty sku_id after migration
		if err := db.Exec(`
			DELETE FROM marketplace_product_mappings
			WHERE sku_id IS NULL OR sku_id = ''
		`).Error; err != nil {
			log.Printf("⚠️ Failed to cleanup orphaned marketplace_product_mappings: %v", err)
		}

		// Recreate unique indexes yang dikelola manual (bukan oleh GORM) — idempotent
		// delivery_order_no dan delivery_batch_id adalah nullable, pakai partial index
		db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_sales_orders_delivery_order_no ON sales_orders(delivery_order_no) WHERE delivery_order_no IS NOT NULL")
		db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_wdo_delivery_batch_id ON warehouse_delivery_orders(delivery_batch_id) WHERE delivery_batch_id IS NOT NULL")
		db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_factory_returns_return_no ON factory_returns(return_no)")

		// Composite indexes untuk mempercepat kalkulasi store category & idempotency lookup
		db.Exec("CREATE INDEX IF NOT EXISTS idx_sales_orders_store_status ON sales_orders(store_id, status)")
		db.Exec("CREATE INDEX IF NOT EXISTS idx_sales_transactions_store_status ON sales_transactions(store_id, status)")
		db.Exec("CREATE INDEX IF NOT EXISTS idx_sales_transactions_store_date ON sales_transactions(store_id, transaction_date DESC)")
		db.Exec("CREATE INDEX IF NOT EXISTS idx_sales_visits_store_checkin ON sales_visits(store_id, check_in_time DESC)")
		db.Exec("CREATE INDEX IF NOT EXISTS idx_stores_assigned_employee ON stores(assigned_employee_id) WHERE is_active = true")
		db.Exec("CREATE INDEX IF NOT EXISTS idx_stores_company_active ON stores(company_id, is_active)")
		db.Exec("CREATE INDEX IF NOT EXISTS idx_sales_orders_client_id ON sales_orders(client_id)")

		// Ensure mobile_session_id exists on users table
		db.Exec("ALTER TABLE users ADD COLUMN IF NOT EXISTS mobile_session_id VARCHAR(64)")

		// Drop NOT NULL constraint on delivery_batches.company_id agar batch lintas company bisa dibuat
		if err := db.Exec(`ALTER TABLE delivery_batches ALTER COLUMN company_id DROP NOT NULL`).Error; err != nil {
			// Ignore jika constraint sudah tidak ada (idempotent)
			log.Printf("ℹ️ ALTER delivery_batches.company_id (oke jika sudah nullable): %v", err)
		}
	} else {
		log.Println("⏩ Skipping GORM AutoMigrate (SKIP_AUTO_MIGRATE=true) for fast startup & zero table locks.")
	}

	// Always ensure raw material procurement columns exist (idempotent)
	db.Exec("ALTER TABLE raw_material_purchase_orders ADD COLUMN IF NOT EXISTS purchase_request_id UUID")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_rm_po_pr_id ON raw_material_purchase_orders(purchase_request_id)")
	db.Exec("ALTER TABLE raw_material_purchase_orders ADD COLUMN IF NOT EXISTS spv_approved_by_id UUID")
	db.Exec("ALTER TABLE raw_material_purchase_orders ADD COLUMN IF NOT EXISTS spv_approved_by_name VARCHAR(100)")
	db.Exec("ALTER TABLE raw_material_purchase_orders ADD COLUMN IF NOT EXISTS spv_approved_by_position VARCHAR(100)")
	db.Exec("ALTER TABLE raw_material_purchase_orders ADD COLUMN IF NOT EXISTS spv_approved_at TIMESTAMPTZ")
	db.Exec("ALTER TABLE raw_material_purchase_orders ADD COLUMN IF NOT EXISTS spv_notes TEXT")
	db.Exec("ALTER TABLE raw_material_purchase_orders ADD COLUMN IF NOT EXISTS purchasing_approved_by_id UUID")
	db.Exec("ALTER TABLE raw_material_purchase_orders ADD COLUMN IF NOT EXISTS purchasing_approved_by_name VARCHAR(100)")
	db.Exec("ALTER TABLE raw_material_purchase_orders ADD COLUMN IF NOT EXISTS purchasing_approved_by_position VARCHAR(100)")
	db.Exec("ALTER TABLE raw_material_purchase_orders ADD COLUMN IF NOT EXISTS purchasing_approved_at TIMESTAMPTZ")
	db.Exec("ALTER TABLE raw_material_purchase_orders ADD COLUMN IF NOT EXISTS purchasing_notes TEXT")
	db.Exec("ALTER TABLE raw_material_purchase_orders ADD COLUMN IF NOT EXISTS progress_status VARCHAR(50) DEFAULT 'PENDING_APPROVAL'")
	db.Exec("ALTER TABLE raw_material_purchase_orders ADD COLUMN IF NOT EXISTS progress_notes TEXT")
	db.Exec("ALTER TABLE raw_material_purchase_orders ADD COLUMN IF NOT EXISTS progress_updated_by_id UUID")
	db.Exec("ALTER TABLE raw_material_purchase_orders ADD COLUMN IF NOT EXISTS progress_updated_by_name VARCHAR(100)")
	db.Exec("ALTER TABLE raw_material_purchase_orders ADD COLUMN IF NOT EXISTS progress_updated_at TIMESTAMPTZ")
	db.Exec("ALTER TABLE raw_material_purchase_orders ADD COLUMN IF NOT EXISTS payment_status VARCHAR(30) DEFAULT 'UNPAID'")
	db.Exec("ALTER TABLE raw_material_purchase_orders ADD COLUMN IF NOT EXISTS payment_notes TEXT")
	db.Exec("ALTER TABLE raw_material_purchase_orders ADD COLUMN IF NOT EXISTS rejection_reason TEXT")
	db.Exec("ALTER TABLE raw_material_purchase_orders ADD COLUMN IF NOT EXISTS rejected_by_id UUID")
	db.Exec("ALTER TABLE raw_material_purchase_orders ADD COLUMN IF NOT EXISTS rejected_by_name VARCHAR(100)")
	db.Exec("ALTER TABLE raw_material_purchase_orders ADD COLUMN IF NOT EXISTS rejected_at TIMESTAMPTZ")

	db.Exec("ALTER TABLE raw_material_purchase_requests ADD COLUMN IF NOT EXISTS purchase_order_id UUID")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_rm_pr_po_id ON raw_material_purchase_requests(purchase_order_id)")
	db.Exec("ALTER TABLE raw_material_purchase_requests ADD COLUMN IF NOT EXISTS reviewed_by_id UUID")
	db.Exec("ALTER TABLE raw_material_purchase_requests ADD COLUMN IF NOT EXISTS reviewed_by_name VARCHAR(100)")
	db.Exec("ALTER TABLE raw_material_purchase_requests ADD COLUMN IF NOT EXISTS reviewed_at TIMESTAMPTZ")
	db.Exec("ALTER TABLE raw_material_purchase_requests ADD COLUMN IF NOT EXISTS review_notes TEXT")
	db.Exec("ALTER TABLE raw_material_purchase_requests ADD COLUMN IF NOT EXISTS rejection_reason TEXT")

	db.Exec("ALTER TABLE raw_material_receipts ADD COLUMN IF NOT EXISTS is_paid_from_petty_cash BOOLEAN DEFAULT FALSE")
	db.Exec("ALTER TABLE raw_material_receipts ADD COLUMN IF NOT EXISTS porter_fee DECIMAL(15,2) DEFAULT 0")
	db.Exec("ALTER TABLE raw_material_receipts ADD COLUMN IF NOT EXISTS petty_cash_id UUID")
	db.Exec("ALTER TABLE raw_material_receipts ADD COLUMN IF NOT EXISTS petty_cash_status VARCHAR(30)")

	db.Exec("ALTER TABLE raw_material_purchase_order_items ADD COLUMN IF NOT EXISTS factory_warehouse_id UUID")
	db.Exec("ALTER TABLE raw_material_purchase_request_items ADD COLUMN IF NOT EXISTS factory_warehouse_id UUID")
	db.Exec("ALTER TABLE raw_material_receipt_items ADD COLUMN IF NOT EXISTS factory_warehouse_id UUID")

	db.Exec("ALTER TABLE marketplace_orders ADD COLUMN IF NOT EXISTS cancel_reason VARCHAR(255)")
	db.Exec("ALTER TABLE marketplace_orders ADD COLUMN IF NOT EXISTS cancel_user VARCHAR(50)")
	db.Exec("ALTER TABLE marketplace_orders ADD COLUMN IF NOT EXISTS fail_delivery_reason VARCHAR(255)")
	db.Exec("ALTER TABLE marketplace_orders ADD COLUMN IF NOT EXISTS failed_delivery_shipping_fee DECIMAL(15,2) DEFAULT 0")
	db.Exec("ALTER TABLE marketplace_orders ADD COLUMN IF NOT EXISTS cogs_amount DECIMAL(15,2) DEFAULT 0")
	db.Exec("ALTER TABLE marketplace_orders ADD COLUMN IF NOT EXISTS packaging_cost DECIMAL(15,2) DEFAULT 0")
	db.Exec("ALTER TABLE marketplace_orders ADD COLUMN IF NOT EXISTS fee_and_tax_amount DECIMAL(15,2) DEFAULT 0")
	db.Exec("ALTER TABLE marketplace_orders ADD COLUMN IF NOT EXISTS shipping_cost_amount DECIMAL(15,2) DEFAULT 0")
	db.Exec("ALTER TABLE marketplace_orders ADD COLUMN IF NOT EXISTS adjustment_amount DECIMAL(15,2) DEFAULT 0")
	db.Exec("ALTER TABLE marketplace_orders ADD COLUMN IF NOT EXISTS fee_tax_breakdown JSONB")
	db.Exec("ALTER TABLE marketplace_orders ADD COLUMN IF NOT EXISTS hold_reason VARCHAR(255)")
	db.Exec("ALTER TABLE marketplace_orders ADD COLUMN IF NOT EXISTS finance_synced_at TIMESTAMPTZ")
	db.Exec("ALTER TABLE marketplace_orders ADD COLUMN IF NOT EXISTS synced_at TIMESTAMPTZ")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_mp_orders_synced_at ON marketplace_orders(synced_at)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_mp_orders_finance_synced_at ON marketplace_orders(finance_synced_at)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_mp_orders_status_settlement ON marketplace_orders(order_status, settlement_status)")

	// Ensure factory_returns approval and rejection columns exist
	db.Exec("ALTER TABLE factory_returns ADD COLUMN IF NOT EXISTS approved_by_id UUID REFERENCES employees(id) ON DELETE SET NULL")
	db.Exec("ALTER TABLE factory_returns ADD COLUMN IF NOT EXISTS approved_by_name VARCHAR(100)")
	db.Exec("ALTER TABLE factory_returns ADD COLUMN IF NOT EXISTS approved_at TIMESTAMPTZ")
	db.Exec("ALTER TABLE factory_returns ADD COLUMN IF NOT EXISTS rejection_reason TEXT")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_factory_returns_approved_by_id ON factory_returns(approved_by_id)")

	// Ensure Reimbursement Flexible Fund Source & Kas Pabrik link columns exist (idempotent)
	db.Exec("ALTER TABLE reimbursements ADD COLUMN IF NOT EXISTS disbursement_source_type VARCHAR(50)")
	db.Exec("ALTER TABLE reimbursements ADD COLUMN IF NOT EXISTS kas_pabrik_id UUID")
	db.Exec("ALTER TABLE reimbursements ADD COLUMN IF NOT EXISTS expense_account_id UUID")
	db.Exec("ALTER TABLE reimbursements ADD COLUMN IF NOT EXISTS expense_account_name VARCHAR(255)")
	db.Exec("ALTER TABLE reimbursements ADD COLUMN IF NOT EXISTS bank_account_id UUID")
	db.Exec("ALTER TABLE kas_transactions ADD COLUMN IF NOT EXISTS reimbursement_id UUID REFERENCES reimbursements(id) ON DELETE SET NULL")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_kas_transactions_reimbursement_id ON kas_transactions(reimbursement_id)")

	// Ensure branding_ai_design_jobs table and indexes exist (idempotent)
	db.Exec(`
		CREATE TABLE IF NOT EXISTS branding_ai_design_jobs (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			banner_order_id UUID NOT NULL REFERENCES banner_orders(id) ON DELETE CASCADE,
			requested_by UUID REFERENCES employees(id) ON DELETE SET NULL,
			status VARCHAR(40) NOT NULL DEFAULT 'queued_hermes',
			reference_banner_id UUID REFERENCES banner_orders(id) ON DELETE SET NULL,
			reference_image_url TEXT,
			prompt TEXT NOT NULL,
			svg_url TEXT,
			preview_url TEXT,
			cdr_url TEXT,
			cdr_file_name VARCHAR(500),
			error_message TEXT,
			claimed_at TIMESTAMPTZ,
			completed_at TIMESTAMPTZ,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
	`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_branding_ai_jobs_banner_created ON branding_ai_design_jobs (banner_order_id, created_at DESC)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_branding_ai_jobs_status_created ON branding_ai_design_jobs (status, created_at ASC)`)
	db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_branding_ai_jobs_one_active_per_banner
		ON branding_ai_design_jobs (banner_order_id)
		WHERE status IN ('queued_hermes', 'hermes_processing', 'awaiting_corel', 'corel_processing')
	`)

	// Ensure sales_payments penerimaan columns exist (idempotent, runs even when SKIP_AUTO_MIGRATE=true)
	db.Exec("ALTER TABLE sales_payments ADD COLUMN IF NOT EXISTS penerimaan_status VARCHAR(50) DEFAULT 'PENDING'")
	db.Exec("ALTER TABLE sales_payments ADD COLUMN IF NOT EXISTS penerimaan_verified_at TIMESTAMPTZ")
	db.Exec("ALTER TABLE sales_payments ADD COLUMN IF NOT EXISTS penerimaan_verified_by_id UUID REFERENCES employees(id) ON DELETE SET NULL")
	db.Exec("ALTER TABLE sales_payments ADD COLUMN IF NOT EXISTS penerimaan_notes TEXT")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_sales_payments_penerimaan_status ON sales_payments(penerimaan_status)")
	db.Exec("ALTER TABLE sales_transactions ALTER COLUMN payment_method TYPE VARCHAR(50)")

	// Distributor secondary sales & visits tracking schema (idempotent, runs even when SKIP_AUTO_MIGRATE=true)
	db.Exec("ALTER TABLE stores ADD COLUMN IF NOT EXISTS distributor_id UUID REFERENCES stores(id) ON DELETE SET NULL")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_stores_distributor_id ON stores(distributor_id)")
	db.Exec("ALTER TABLE users ADD COLUMN IF NOT EXISTS distributor_store_id UUID REFERENCES stores(id) ON DELETE SET NULL")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_users_distributor_store_id ON users(distributor_store_id)")
	db.Exec("ALTER TABLE employees ADD COLUMN IF NOT EXISTS is_distributor BOOLEAN NOT NULL DEFAULT FALSE")
	db.Exec("ALTER TABLE employees ADD COLUMN IF NOT EXISTS distributor_store_id UUID REFERENCES stores(id) ON DELETE SET NULL")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_employees_distributor ON employees(is_distributor, distributor_store_id)")
	db.Exec("ALTER TABLE sales_visits ALTER COLUMN employee_id DROP NOT NULL")
	db.Exec("ALTER TABLE sales_visits ADD COLUMN IF NOT EXISTS user_id UUID REFERENCES users(id) ON DELETE SET NULL")
	db.Exec("ALTER TABLE sales_visits ADD COLUMN IF NOT EXISTS distributor_store_id UUID REFERENCES stores(id) ON DELETE SET NULL")
	db.Exec("ALTER TABLE sales_visits ADD COLUMN IF NOT EXISTS is_distributor_visit BOOLEAN NOT NULL DEFAULT FALSE")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_sales_visits_distributor ON sales_visits(distributor_store_id, is_distributor_visit)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_sales_visits_user_id ON sales_visits(user_id)")
	db.Exec("ALTER TABLE sales_transactions ALTER COLUMN employee_id DROP NOT NULL")
	db.Exec("ALTER TABLE sales_transactions ADD COLUMN IF NOT EXISTS user_id UUID REFERENCES users(id) ON DELETE SET NULL")
	db.Exec("ALTER TABLE sales_transactions ADD COLUMN IF NOT EXISTS distributor_store_id UUID REFERENCES stores(id) ON DELETE SET NULL")
	db.Exec("ALTER TABLE sales_transactions ADD COLUMN IF NOT EXISTS is_distributor_order BOOLEAN NOT NULL DEFAULT FALSE")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_sales_transactions_distributor ON sales_transactions(distributor_store_id, is_distributor_order)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_sales_transactions_user_id ON sales_transactions(user_id)")
	db.Exec("ALTER TABLE sales_orders ALTER COLUMN employee_id DROP NOT NULL")
	db.Exec("ALTER TABLE sales_orders ADD COLUMN IF NOT EXISTS user_id UUID REFERENCES users(id) ON DELETE SET NULL")
	db.Exec("ALTER TABLE sales_orders ADD COLUMN IF NOT EXISTS distributor_store_id UUID REFERENCES stores(id) ON DELETE SET NULL")
	db.Exec("ALTER TABLE sales_orders ADD COLUMN IF NOT EXISTS is_distributor_order BOOLEAN NOT NULL DEFAULT FALSE")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_sales_orders_distributor ON sales_orders(distributor_store_id, is_distributor_order)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_sales_orders_user_id ON sales_orders(user_id)")

	// Visit Management & Sales Filtering performance indexes (idempotent)
	db.Exec("CREATE INDEX IF NOT EXISTS idx_sales_transactions_visit_id ON sales_transactions(visit_id) WHERE deleted_at IS NULL")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_sales_items_transaction_id ON sales_items(sales_transaction_id) WHERE deleted_at IS NULL")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_sales_orders_invoice_id ON sales_orders(invoice_id) WHERE deleted_at IS NULL")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_sales_visits_check_in_time ON sales_visits(check_in_time DESC)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_sales_visits_store_check_in ON sales_visits(store_id, check_in_time DESC)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_sales_transactions_date ON sales_transactions(transaction_date DESC) WHERE deleted_at IS NULL")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_sales_transactions_store_date ON sales_transactions(store_id, transaction_date DESC) WHERE deleted_at IS NULL")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_sales_orders_employee_order_date ON sales_orders(employee_id, order_date DESC) WHERE deleted_at IS NULL")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_sales_orders_order_date ON sales_orders(order_date DESC) WHERE deleted_at IS NULL")

	// Shopee Auto Boost tables (idempotent, runs even when SKIP_AUTO_MIGRATE=true)
	db.Exec(`CREATE TABLE IF NOT EXISTS shopee_auto_boost_items (
		id BIGSERIAL PRIMARY KEY,
		branch_id UUID REFERENCES branches(id) ON DELETE SET NULL,
		shop_id VARCHAR(100) NOT NULL,
		item_id BIGINT NOT NULL,
		item_name VARCHAR(500) NOT NULL,
		item_sku VARCHAR(100),
		image_url TEXT,
		price_amount NUMERIC(15, 2) DEFAULT 0,
		sequence INT NOT NULL DEFAULT 0,
		is_active BOOLEAN NOT NULL DEFAULT TRUE,
		last_boosted_at TIMESTAMPTZ,
		cooldown_until TIMESTAMPTZ,
		status VARCHAR(50) NOT NULL DEFAULT 'IDLE',
		last_error TEXT,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		CONSTRAINT uq_shopee_auto_boost_shop_item UNIQUE (shop_id, item_id)
	)`)
	db.Exec("CREATE INDEX IF NOT EXISTS idx_shopee_boost_shop_active ON shopee_auto_boost_items(shop_id, is_active)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_shopee_boost_cooldown ON shopee_auto_boost_items(cooldown_until)")
	db.Exec(`CREATE TABLE IF NOT EXISTS shopee_auto_boost_settings (
		id BIGSERIAL PRIMARY KEY,
		branch_id UUID REFERENCES branches(id) ON DELETE SET NULL,
		shop_id VARCHAR(100) NOT NULL UNIQUE,
		is_enabled BOOLEAN NOT NULL DEFAULT FALSE,
		rotation_mode VARCHAR(50) NOT NULL DEFAULT 'ROUND_ROBIN',
		check_interval_minutes INT NOT NULL DEFAULT 5,
		last_run_at TIMESTAMPTZ,
		last_run_status VARCHAR(50),
		last_run_message TEXT,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	)`)
	db.Exec(`CREATE TABLE IF NOT EXISTS shopee_auto_boost_logs (
		id BIGSERIAL PRIMARY KEY,
		shop_id VARCHAR(100) NOT NULL,
		item_id BIGINT NOT NULL,
		item_name VARCHAR(500),
		triggered_by VARCHAR(50) NOT NULL DEFAULT 'WORKER',
		status VARCHAR(50) NOT NULL,
		message TEXT,
		boosted_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	)`)
	db.Exec("CREATE INDEX IF NOT EXISTS idx_shopee_boost_logs_shop ON shopee_auto_boost_logs(shop_id, boosted_at DESC)")

	// Ensure factory_qc_samples table exists (idempotent, runs even when SKIP_AUTO_MIGRATE=true)
	db.Exec(`CREATE TABLE IF NOT EXISTS factory_qc_samples (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		factory_id UUID NOT NULL REFERENCES factories(id) ON DELETE CASCADE,
		product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
		production_log_id UUID REFERENCES production_logs(id) ON DELETE SET NULL,
		batch_no VARCHAR(50) NOT NULL,
		expiry_date TIMESTAMPTZ,
		production_date TIMESTAMPTZ,
		quantity INT NOT NULL DEFAULT 1,
		sample_type VARCHAR(50) NOT NULL DEFAULT 'LAB_QC',
		purpose TEXT,
		status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
		requested_by_id UUID NOT NULL,
		requested_by_name VARCHAR(100) NOT NULL,
		requested_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		approved_by_id UUID,
		approved_by_name VARCHAR(100),
		approved_at TIMESTAMPTZ,
		rejection_reason TEXT,
		notes TEXT,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	)`)
	db.Exec("CREATE INDEX IF NOT EXISTS idx_factory_qc_samples_factory_id ON factory_qc_samples(factory_id)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_factory_qc_samples_product_id ON factory_qc_samples(product_id)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_factory_qc_samples_production_log_id ON factory_qc_samples(production_log_id)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_factory_qc_samples_status ON factory_qc_samples(status)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_factory_qc_samples_created_at ON factory_qc_samples(created_at DESC)")

	return db
}

// ConnectBrandingDB tries to connect to the external catalog MySQL
// database with a hard 8-second timeout so it NEVER blocks server startup.
// Returns nil (graceful offline mode) if the host is unreachable.
func ConnectBrandingDB(cfg Config) *gorm.DB {
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local&timeout=15s&readTimeout=60s&writeTimeout=60s",
		cfg.BrandingDBUser,
		cfg.BrandingDBPassword,
		cfg.BrandingDBHost,
		cfg.BrandingDBPort,
		cfg.BrandingDBName,
	)

	type result struct {
		db  *gorm.DB
		err error
	}

	ch := make(chan result, 1)

	// Run connection attempt in a goroutine so server startup is never blocked.
	go func() {
		db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Warn),
		})
		ch <- result{db, err}
	}()

	// Wait at most 8 seconds for the connection goroutine.
	select {
	case r := <-ch:
		if r.err != nil {
			log.Printf("[BrandingDB] ⚠️  Tidak dapat terhubung ke MySQL Branding: %v", r.err)
			log.Println("[BrandingDB] Mode offline aktif — data Branding tidak tersedia.")
			return nil
		}

		sqlDB, err := r.db.DB()
		if err != nil {
			log.Printf("[BrandingDB] Gagal mendapatkan instance DB: %v", err)
			return nil
		}

		// Conservative pool — external DB, keep connections minimal
		sqlDB.SetMaxIdleConns(2)
		sqlDB.SetMaxOpenConns(10)
		sqlDB.SetConnMaxLifetime(3 * time.Minute)
		sqlDB.SetConnMaxIdleTime(1 * time.Minute)

		log.Println("[BrandingDB] ✅  Terhubung ke MySQL Branding App")
		return r.db

	case <-time.After(20 * time.Second):
		log.Println("[BrandingDB] ⚠️  Koneksi MySQL Branding timeout (20s) — mode offline aktif.")
		log.Println("[BrandingDB] Kemungkinan penyebab: remote MySQL diblokir oleh shared hosting.")
		log.Println("[BrandingDB] Aktifkan Remote MySQL di cPanel/Hostinger untuk mengizinkan IP server ini.")
		return nil
	}
}
