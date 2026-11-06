package main

import (
	"log"

	"example.com/hris/backend/config"
	"example.com/hris/backend/internal/domain"
	"example.com/hris/backend/internal/repository"
)

func main() {
	// Initialize Config
	cfg := config.LoadConfig()

	// Initialize Database
	db := config.ConnectDB(cfg)

	log.Println("Starting GORM Migrations...")

	// FIX: Ensure missing columns in delivery_items exist BEFORE AutoMigrate
	log.Println("Checking and fixing delivery_items schema...")
	db.Exec("ALTER TABLE delivery_items ADD COLUMN IF NOT EXISTS received_by VARCHAR(100)")
	db.Exec("ALTER TABLE delivery_items ADD COLUMN IF NOT EXISTS pod_image_url TEXT")
	db.Exec("ALTER TABLE delivery_items ADD COLUMN IF NOT EXISTS signature_image_url TEXT")
	db.Exec("ALTER TABLE delivery_items ADD COLUMN IF NOT EXISTS delivered_at TIMESTAMP WITH TIME ZONE")
	db.Exec("ALTER TABLE delivery_items ADD COLUMN IF NOT EXISTS sales_transfer_id UUID")
	db.Exec("ALTER TABLE delivery_items ADD COLUMN IF NOT EXISTS payment_collected BOOLEAN DEFAULT FALSE")
	db.Exec("ALTER TABLE delivery_items ADD COLUMN IF NOT EXISTS payment_amount DECIMAL(15,2) DEFAULT 0")
	db.Exec("ALTER TABLE delivery_items ADD COLUMN IF NOT EXISTS payment_method VARCHAR(50)")
	db.Exec("ALTER TABLE delivery_items ADD COLUMN IF NOT EXISTS payment_collected_at TIMESTAMP WITH TIME ZONE")
	// Kolom retur barang — ditambahkan untuk fitur retur pengiriman
	db.Exec("ALTER TABLE delivery_items ADD COLUMN IF NOT EXISTS return_quantity INT DEFAULT 0")
	db.Exec("ALTER TABLE delivery_items ADD COLUMN IF NOT EXISTS return_reason TEXT")
	db.Exec("ALTER TABLE delivery_items ADD COLUMN IF NOT EXISTS has_return BOOLEAN DEFAULT FALSE")
	db.Exec("ALTER TABLE warehouse_logs ADD COLUMN IF NOT EXISTS batch_no VARCHAR(50)")
	db.Exec("ALTER TABLE production_logs ADD COLUMN IF NOT EXISTS batch_group_id UUID")
	db.Exec("ALTER TABLE production_logs ADD COLUMN IF NOT EXISTS operator_name TEXT")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_production_logs_batch_group_id ON production_logs(batch_group_id)")

	db.Exec("ALTER TABLE sales_orders ADD COLUMN IF NOT EXISTS signature_image_url TEXT")
	db.Exec("ALTER TABLE sales_transactions ADD COLUMN IF NOT EXISTS signature_image_url TEXT")
	db.Exec("ALTER TABLE sales_transactions ADD COLUMN IF NOT EXISTS dp_payment_method VARCHAR(20)")

	// Distributor secondary sales & visits tracking schema
	db.Exec("ALTER TABLE stores ADD COLUMN IF NOT EXISTS distributor_id UUID REFERENCES stores(id) ON DELETE SET NULL")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_stores_distributor_id ON stores(distributor_id)")
	db.Exec("ALTER TABLE users ADD COLUMN IF NOT EXISTS distributor_store_id UUID REFERENCES stores(id) ON DELETE SET NULL")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_users_distributor_store_id ON users(distributor_store_id)")
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
	db.Exec("CREATE INDEX IF NOT EXISTS idx_sales_transactions_created_at ON sales_transactions(created_at DESC)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_sales_transactions_company_id ON sales_transactions(company_id)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_sales_transactions_employee_id ON sales_transactions(employee_id)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_sales_transactions_store_id ON sales_transactions(store_id)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_sales_transactions_trx_date ON sales_transactions(transaction_date)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_sales_transactions_status ON sales_transactions(status)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_sales_transactions_branch_id ON sales_transactions(branch_id)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_sales_transactions_payment ON sales_transactions(payment_method, payment_status)")
	db.Exec("ALTER TABLE sales_orders ALTER COLUMN employee_id DROP NOT NULL")
	db.Exec("ALTER TABLE sales_orders ADD COLUMN IF NOT EXISTS user_id UUID REFERENCES users(id) ON DELETE SET NULL")
	db.Exec("ALTER TABLE sales_orders ADD COLUMN IF NOT EXISTS distributor_store_id UUID REFERENCES stores(id) ON DELETE SET NULL")
	db.Exec("ALTER TABLE sales_orders ADD COLUMN IF NOT EXISTS is_distributor_order BOOLEAN NOT NULL DEFAULT FALSE")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_sales_orders_distributor ON sales_orders(distributor_store_id, is_distributor_order)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_sales_orders_user_id ON sales_orders(user_id)")

	// Products dimensions and marketplace price additions
	db.Exec("ALTER TABLE products ADD COLUMN IF NOT EXISTS length DECIMAL(10,2) DEFAULT 0")
	db.Exec("ALTER TABLE products ADD COLUMN IF NOT EXISTS width DECIMAL(10,2) DEFAULT 0")
	db.Exec("ALTER TABLE products ADD COLUMN IF NOT EXISTS height DECIMAL(10,2) DEFAULT 0")
	db.Exec("ALTER TABLE products ADD COLUMN IF NOT EXISTS volume DECIMAL(10,4) DEFAULT 0")
	db.Exec("ALTER TABLE products ADD COLUMN IF NOT EXISTS marketplace_factory_price DECIMAL(15,2) DEFAULT 0")
	db.Exec("ALTER TABLE product_transfers ADD COLUMN IF NOT EXISTS total_volume DECIMAL(10,4) DEFAULT 0")

	// Raw material procurement schema additions
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
	db.Exec("ALTER TABLE raw_material_purchase_requests ALTER COLUMN factory_id DROP NOT NULL")
	db.Exec("ALTER TABLE raw_material_purchase_request_items ALTER COLUMN product_id DROP NOT NULL")
	db.Exec("ALTER TABLE raw_material_purchase_request_items ADD COLUMN IF NOT EXISTS item_name VARCHAR(255) NOT NULL DEFAULT ''")
	db.Exec("ALTER TABLE raw_material_receipt_items ADD COLUMN IF NOT EXISTS factory_warehouse_id UUID")

	log.Println("Checking and fixing sales_payments schema...")
	db.Exec("ALTER TABLE sales_payments ADD COLUMN IF NOT EXISTS sales_order_id UUID")
	db.Exec("ALTER TABLE sales_payments ADD COLUMN IF NOT EXISTS sales_transaction_id UUID")

	// FIX: Drop NOT NULL constraints if they exist to allow EITHER SalesOrderID OR SalesTransactionID
	db.Exec("ALTER TABLE sales_payments ALTER COLUMN sales_order_id DROP NOT NULL")
	db.Exec("ALTER TABLE sales_payments ALTER COLUMN sales_transaction_id DROP NOT NULL")

	db.Exec("ALTER TABLE sales_payments ADD COLUMN IF NOT EXISTS employee_id UUID")
	db.Exec("ALTER TABLE sales_payments ADD COLUMN IF NOT EXISTS amount DECIMAL(15,2) DEFAULT 0")
	db.Exec("ALTER TABLE sales_payments ADD COLUMN IF NOT EXISTS payment_method VARCHAR(50)")
	db.Exec("ALTER TABLE sales_payments ADD COLUMN IF NOT EXISTS payment_status VARCHAR(20) DEFAULT 'SUCCESS'")
	db.Exec("ALTER TABLE sales_payments ADD COLUMN IF NOT EXISTS midtrans_transaction_id VARCHAR(100)")
	db.Exec("ALTER TABLE sales_payments ADD COLUMN IF NOT EXISTS notes TEXT")
	db.Exec("ALTER TABLE sales_payments ADD COLUMN IF NOT EXISTS collected_by VARCHAR(100)")
	db.Exec("ALTER TABLE sales_payments ADD COLUMN IF NOT EXISTS payment_date TIMESTAMP WITH TIME ZONE DEFAULT NOW()")

	log.Println("Fixing sales_orders schema for Midtrans details...")
	db.Exec("ALTER TABLE sales_orders ADD COLUMN IF NOT EXISTS midtrans_qris_url TEXT")
	db.Exec("ALTER TABLE sales_orders ADD COLUMN IF NOT EXISTS midtrans_va_number VARCHAR(50)")
	db.Exec("ALTER TABLE sales_orders ADD COLUMN IF NOT EXISTS midtrans_bank VARCHAR(20)")
	db.Exec("ALTER TABLE sales_orders ADD COLUMN IF NOT EXISTS midtrans_bill_key VARCHAR(50)")
	db.Exec("ALTER TABLE sales_orders ADD COLUMN IF NOT EXISTS midtrans_biller_code VARCHAR(20)")

	db.Exec("ALTER TABLE delivery_batches ADD COLUMN IF NOT EXISTS total_cash_collected DECIMAL(15,2) DEFAULT 0")
	db.Exec("ALTER TABLE delivery_batches ADD COLUMN IF NOT EXISTS total_transfer_collected DECIMAL(15,2) DEFAULT 0")

	log.Println("Fixing shift time types...")
	db.Exec("ALTER TABLE shifts ALTER COLUMN start_time TYPE TIME USING start_time::TIME")
	db.Exec("ALTER TABLE shifts ALTER COLUMN end_time TYPE TIME USING end_time::TIME")
	db.Exec("ALTER TABLE shifts ALTER COLUMN break_start TYPE TIME USING break_start::TIME")
	db.Exec("ALTER TABLE shifts ALTER COLUMN break_end TYPE TIME USING break_end::TIME")

	log.Println("Ensuring factory_price exists in products...")
	db.Exec("ALTER TABLE products ADD COLUMN IF NOT EXISTS factory_price DECIMAL(15,2) DEFAULT 0")

	log.Println("Schema check completed.")

	// Ensure factory warehouse_code column exists
	db.Exec("ALTER TABLE factories ADD COLUMN IF NOT EXISTS warehouse_code VARCHAR(50)")
	db.Exec("ALTER TABLE factories ADD COLUMN IF NOT EXISTS description TEXT")
	db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_factories_warehouse_code ON factories(warehouse_code) WHERE warehouse_code IS NOT NULL AND warehouse_code != ''")

	// Kolom baru untuk DO Motoris (standalone tanpa SJ)
	db.Exec("ALTER TABLE warehouse_delivery_orders ADD COLUMN IF NOT EXISTS do_type VARCHAR(20) DEFAULT 'REGULAR'")
	db.Exec("ALTER TABLE warehouse_delivery_orders ADD COLUMN IF NOT EXISTS employee_id UUID")
	db.Exec("ALTER TABLE warehouse_delivery_orders ALTER COLUMN delivery_batch_id DROP NOT NULL")
	db.Exec("DROP INDEX IF EXISTS idx_warehouse_delivery_orders_delivery_batch_id")
	db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_wdo_delivery_batch_id ON warehouse_delivery_orders(delivery_batch_id) WHERE delivery_batch_id IS NOT NULL")

	// Petty cash, cashier report, and reimbursement disbursement
	db.Exec("ALTER TABLE banner_orders ADD COLUMN IF NOT EXISTS design_validator_id UUID REFERENCES employees(id) ON DELETE SET NULL")
	db.Exec("ALTER TABLE banner_orders ADD COLUMN IF NOT EXISTS design_validator_name VARCHAR(255)")
	db.Exec("ALTER TABLE banner_orders ADD COLUMN IF NOT EXISTS design_validator_position VARCHAR(255)")
	db.Exec("ALTER TABLE banner_orders ADD COLUMN IF NOT EXISTS design_approved_at TIMESTAMP WITH TIME ZONE")
	db.Exec("ALTER TABLE banner_orders ADD COLUMN IF NOT EXISTS store_validator_id UUID REFERENCES employees(id) ON DELETE SET NULL")
	db.Exec("ALTER TABLE banner_orders ADD COLUMN IF NOT EXISTS store_validator_name VARCHAR(255)")
	db.Exec("ALTER TABLE banner_orders ADD COLUMN IF NOT EXISTS store_validator_position VARCHAR(255)")
	db.Exec("ALTER TABLE banner_orders ADD COLUMN IF NOT EXISTS store_approved_at TIMESTAMP WITH TIME ZONE")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_banner_orders_design_validator_id ON banner_orders(design_validator_id)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_banner_orders_store_validator_id ON banner_orders(store_validator_id)")
	db.Exec(`UPDATE banner_orders
		SET 
			design_validator_id = COALESCE(design_validator_id, validator_id),
			design_validator_name = COALESCE(design_validator_name, validator_name),
			design_validator_position = COALESCE(design_validator_position, validator_position),
			design_approved_at = COALESCE(design_approved_at, validated_at, updated_at)
		WHERE is_design_approved = true AND (design_validator_name IS NULL OR design_validator_name = '')`)
	db.Exec(`UPDATE banner_orders
		SET 
			store_validator_id = COALESCE(store_validator_id, validator_id),
			store_validator_name = COALESCE(store_validator_name, validator_name),
			store_validator_position = COALESCE(store_validator_position, validator_position),
			store_approved_at = COALESCE(store_approved_at, validated_at, updated_at)
		WHERE is_store_approved = true AND (store_validator_name IS NULL OR store_validator_name = '')`)

	db.Exec("ALTER TABLE branches ADD COLUMN IF NOT EXISTS petty_cash_balance DECIMAL(15,2) DEFAULT 0")
	db.Exec("CREATE TABLE IF NOT EXISTS cashier_financial_reports (id UUID PRIMARY KEY DEFAULT gen_random_uuid())")
	db.Exec("ALTER TABLE cashier_financial_reports ADD COLUMN IF NOT EXISTS allocated_to_petty_cash DECIMAL(15,2) DEFAULT 0")
	db.Exec("ALTER TABLE cashier_financial_reports ADD COLUMN IF NOT EXISTS deposit_destination_type VARCHAR(30) DEFAULT 'BRANCH_BANK'")
	db.Exec("ALTER TABLE cashier_financial_reports ADD COLUMN IF NOT EXISTS target_kas_pabrik_id UUID")
	db.Exec("ALTER TABLE reimbursements ADD COLUMN IF NOT EXISTS category VARCHAR(50) DEFAULT 'GENERAL'")
	db.Exec("ALTER TABLE reimbursements ADD COLUMN IF NOT EXISTS banner_count INT DEFAULT 0")
	db.Exec("ALTER TABLE reimbursements ADD COLUMN IF NOT EXISTS unit_rate DECIMAL(15,2) DEFAULT 0")
	db.Exec("ALTER TABLE reimbursements ADD COLUMN IF NOT EXISTS disbursement_status VARCHAR(20) DEFAULT 'UNPAID'")
	db.Exec("ALTER TABLE reimbursements ADD COLUMN IF NOT EXISTS disbursed_via VARCHAR(20)")
	db.Exec("ALTER TABLE reimbursements ADD COLUMN IF NOT EXISTS disbursed_at TIMESTAMP WITH TIME ZONE")
	db.Exec("ALTER TABLE reimbursements ADD COLUMN IF NOT EXISTS disbursed_by UUID")
	db.Exec("ALTER TABLE reimbursements ADD COLUMN IF NOT EXISTS disbursement_proof_url TEXT")
	db.Exec("CREATE TABLE IF NOT EXISTS cash_bank_mutations (id UUID PRIMARY KEY DEFAULT gen_random_uuid())")
	db.Exec("ALTER TABLE cash_bank_mutations ADD COLUMN IF NOT EXISTS category VARCHAR(50)")
	db.Exec("ALTER TABLE cash_bank_mutations ADD COLUMN IF NOT EXISTS reimbursement_id UUID")

	db.Exec("ALTER TABLE reimbursements ADD COLUMN IF NOT EXISTS disbursement_source_type VARCHAR(50)")
	db.Exec("ALTER TABLE reimbursements ADD COLUMN IF NOT EXISTS kas_pabrik_id UUID")
	db.Exec("ALTER TABLE reimbursements ADD COLUMN IF NOT EXISTS expense_account_id UUID")
	db.Exec("ALTER TABLE reimbursements ADD COLUMN IF NOT EXISTS expense_account_name VARCHAR(255)")
	db.Exec("ALTER TABLE reimbursements ADD COLUMN IF NOT EXISTS bank_account_id UUID")

	// Kas Pabrik extended features (piutang cabang, armada kendaraan, kasbon dinas, faktur)
	db.Exec("ALTER TABLE kas_transactions ADD COLUMN IF NOT EXISTS is_branch_receivable BOOLEAN DEFAULT FALSE")
	db.Exec("ALTER TABLE kas_transactions ADD COLUMN IF NOT EXISTS target_branch_id UUID")
	db.Exec("ALTER TABLE kas_transactions ADD COLUMN IF NOT EXISTS vehicle_id UUID")
	db.Exec("ALTER TABLE kas_transactions ADD COLUMN IF NOT EXISTS employee_id UUID")
	db.Exec("ALTER TABLE kas_transactions ADD COLUMN IF NOT EXISTS advance_status VARCHAR(20) DEFAULT 'NONE'")
	db.Exec("ALTER TABLE kas_transactions ADD COLUMN IF NOT EXISTS sales_payment_id UUID")
	db.Exec("ALTER TABLE kas_transactions ADD COLUMN IF NOT EXISTS sales_order_id UUID")
	db.Exec("ALTER TABLE kas_transactions ADD COLUMN IF NOT EXISTS reimbursement_id UUID REFERENCES reimbursements(id) ON DELETE SET NULL")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_kas_transactions_reimbursement_id ON kas_transactions(reimbursement_id)")
	db.Exec(`CREATE TABLE IF NOT EXISTS kas_advance_settlements (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		kas_transaction_id UUID NOT NULL,
		actual_expense_amount DECIMAL(20,2) NOT NULL DEFAULT 0,
		refund_amount DECIMAL(20,2) NOT NULL DEFAULT 0,
		shortage_amount DECIMAL(20,2) NOT NULL DEFAULT 0,
		notes TEXT,
		proof_url TEXT,
		settled_by UUID,
		settled_at TIMESTAMPTZ DEFAULT NOW(),
		created_at TIMESTAMPTZ DEFAULT NOW(),
		updated_at TIMESTAMPTZ DEFAULT NOW()
	)`)

	// ── Modul Akuntansi PSAK — buat tabel jika belum ada ──────────────────────
	log.Println("Setting up accounting tables...")
	db.Exec(`CREATE TABLE IF NOT EXISTS accounts (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		company_id UUID REFERENCES companies(id),
		code VARCHAR(20) NOT NULL,
		name VARCHAR(200) NOT NULL,
		account_type VARCHAR(30) NOT NULL,
		account_subtype VARCHAR(50),
		normal_balance VARCHAR(10) NOT NULL,
		parent_id UUID REFERENCES accounts(id),
		is_active BOOLEAN DEFAULT TRUE,
		is_system BOOLEAN DEFAULT FALSE,
		created_at TIMESTAMPTZ DEFAULT now(),
		UNIQUE(company_id, code)
	)`)
	db.Exec(`CREATE TABLE IF NOT EXISTS journal_entries (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		company_id UUID NOT NULL REFERENCES companies(id),
		entry_number VARCHAR(30) UNIQUE,
		entry_date DATE NOT NULL,
		description TEXT,
		reference_type VARCHAR(50),
		reference_id UUID,
		status VARCHAR(20) DEFAULT 'POSTED',
		created_by UUID REFERENCES users(id),
		created_at TIMESTAMPTZ DEFAULT now(),
		updated_at TIMESTAMPTZ DEFAULT now()
	)`)
	db.Exec(`CREATE TABLE IF NOT EXISTS journal_entry_lines (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		journal_entry_id UUID NOT NULL REFERENCES journal_entries(id) ON DELETE CASCADE,
		account_id UUID NOT NULL REFERENCES accounts(id),
		debit_amount DECIMAL(20,2) DEFAULT 0,
		credit_amount DECIMAL(20,2) DEFAULT 0,
		description TEXT
	)`)
	db.Exec(`CREATE TABLE IF NOT EXISTS accounting_periods (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		company_id UUID NOT NULL REFERENCES companies(id),
		period_year INT NOT NULL,
		period_month INT NOT NULL,
		status VARCHAR(20) DEFAULT 'OPEN',
		closed_at TIMESTAMPTZ,
		closed_by UUID REFERENCES users(id),
		UNIQUE(company_id, period_year, period_month)
	)`)
	db.Exec(`CREATE TABLE IF NOT EXISTS account_balances (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		account_id UUID NOT NULL REFERENCES accounts(id),
		period_year INT NOT NULL,
		period_month INT NOT NULL,
		debit_total DECIMAL(20,2) DEFAULT 0,
		credit_total DECIMAL(20,2) DEFAULT 0,
		balance DECIMAL(20,2) DEFAULT 0,
		updated_at TIMESTAMPTZ DEFAULT now(),
		UNIQUE(account_id, period_year, period_month)
	)`)
	// Index performa untuk laporan keuangan
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_journal_entries_company_date ON journal_entries(company_id, entry_date)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_journal_entry_lines_entry ON journal_entry_lines(journal_entry_id)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_account_balances_account_period ON account_balances(account_id, period_year, period_month)`)
	log.Println("Accounting tables ready.")

	// Run AutoMigrate for critical models to ensure schema sync
	log.Println("Running GORM AutoMigrate...")

	// Force migrate WMS tables first to ensure they exist even if the main AutoMigrate fails
	if err := db.AutoMigrate(&domain.Order{}, &domain.OrderItem{}); err != nil {
		log.Printf("Failed to AutoMigrate Order/OrderItem: %v", err)
	}

	// Force migrate new tables that might get skipped if the large block fails midway
	if err := db.AutoMigrate(&domain.SalesVisitPlanProjection{}); err != nil {
		log.Printf("Failed to AutoMigrate SalesVisitPlanProjection: %v", err)
	}

	if err := db.AutoMigrate(
		&domain.DailyWorker{},
		&domain.DailyWorkerSalary{},
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
		&domain.DailyKpiReport{},
		&domain.DailyKpiReportItem{},
		&domain.DailyKpiReportActionPlan{},
		&domain.BannerOrder{},
		&domain.Factory{},
		&domain.Product{},
		&domain.FactoryStock{},
		&domain.ProductionLog{},
		&domain.ProductionLogOperator{},
		&domain.ProductTransfer{},
		&domain.BranchPaymentLog{},
		&domain.FactoryInventoryLog{},
		&domain.WarehouseStock{},
		&domain.WarehouseLog{},
		&domain.Vehicle{},
		&domain.VehicleComponent{},
		&domain.VehicleMaintenance{},
		&domain.TransportTrip{},
		&domain.TransportBackhaul{},
		&domain.TransportExpenseClaim{},
		&domain.TransportFleetInspection{},
		&domain.TransportDriverAttendance{},
		&domain.Notification{},
		&domain.DeliveryBatch{},
		&domain.DeliveryItem{},
		&domain.DeliverySpare{},
		&domain.SalesPayment{},
		&domain.SalesItem{},
		&domain.SalesStock{},
		&domain.SalesTransfer{},
		&domain.SalesVisit{},
		&domain.LPSReport{},
		&domain.SalesTransaction{},
		&domain.SalesTransactionBonusItem{},
		&domain.SalesOrder{},
		&domain.SalesOrderItem{},
		&domain.SalesOrderItemBatch{},
		&domain.SalesReturn{},
		&domain.SalesReturnItem{},
		&domain.SalesSubsequentItem{},
		&domain.ProductExchange{},
		&domain.ProductExchangeItem{},
		&domain.WarehouseDeliveryOrder{},
		&domain.WarehouseDeliveryOrderItem{},
		&domain.WarehouseDeliveryOrderTransfer{},
		&domain.FactoryReturn{},
		&domain.FactoryReturnItem{},
		&domain.FactoryQcSample{},
		&domain.ProductionRecipe{},
		&domain.ProductionRecipeItem{},
		&domain.Shift{},
		&domain.EmployeeShift{},
		&domain.SalesVisitPlan{},
		&domain.SalesVisitPlanProjection{},
		&domain.SalesInstallment{},
		&domain.UserFcmToken{},
		&domain.CompanyBankAccount{},
		&domain.CashierFinancialReport{},
		&domain.InvoiceDeposit{},
		&domain.InvoiceDepositItem{},
		&domain.CashBankMutation{},
		&domain.CashBankMutationInvoice{},
		&domain.KasPabrik{},
		&domain.KasTransaction{},
		&domain.KasAdvanceSettlement{},
		&domain.Permission{},
		&domain.UserPermission{},
		&domain.Reimbursement{},
		&domain.EpkwtRequest{},
		&domain.AuditLog{},
		&domain.KeyAccount{},
		&domain.MTContract{},
		&domain.MTPurchaseOrder{},
		&domain.MTMerchandising{},
		&domain.MTPromotion{},
		&domain.Supplier{},
		&domain.SupplierReceipt{},
		&domain.SupplierReceiptItem{},
		&domain.SupplierStock{},
		&domain.MTDeliveryOrder{},
		&domain.MTDeliveryOrderItem{},
		&domain.MarketplaceStock{},
		&domain.MarketplaceOpname{},
		&domain.OperatorTarget{},
		&domain.BonusInventory{},
		&domain.BonusInventoryLog{},
		&domain.SalesOrderBonusItem{},
		&domain.SalesBonusRequest{},
		&domain.SalesBonusStock{},
		&domain.MarketplaceToken{},
		&domain.MarketplaceConfig{},
		&domain.MarketplaceOrder{},
		&domain.MarketplaceUnsettledSnapshot{},
		&domain.MarketplaceUnsettledTransaction{},
		&domain.MarketplaceProduct{},
		&domain.MarketplaceProductMapping{},
		&domain.Order{},
		&domain.OrderItem{},
		&domain.Account{},
		&domain.JournalEntry{},
		&domain.JournalEntryLine{},
		&domain.AccountingPeriod{},
		&domain.AccountBalance{},
		// Barang sampingan sales (Krat, Botol, Galon, Mangkok, Piring, Gelas, Retur, dll)
	); err != nil {
		log.Printf("GORM AutoMigrate failed: %v", err)
	} else {
		log.Println("GORM AutoMigrate completed successfully")

		// Seed default permissions if empty
		var count int64
		if err := db.Model(&domain.Permission{}).Count(&count).Error; err == nil && count == 0 {
			log.Println("Seeding default master permissions...")
			defaultPerms := []domain.Permission{
				{Name: "employee:read", MenuGroup: "Employee", Description: "Membaca data karyawan"},
				{Name: "employee:write", MenuGroup: "Employee", Description: "Membuat, mengubah, dan menghapus data karyawan"},
				{Name: "payroll:read", MenuGroup: "Payroll", Description: "Melihat data slip gaji dan laporan penggajian"},
				{Name: "payroll:write", MenuGroup: "Payroll", Description: "Membuat, mengubah, dan menghapus proses penggajian"},
				{Name: "attendance:read", MenuGroup: "Attendance", Description: "Melihat rekap absensi karyawan"},
				{Name: "attendance:write", MenuGroup: "Attendance", Description: "Mengubah atau memverifikasi absensi manual"},
				{Name: "leave:read", MenuGroup: "Leave", Description: "Melihat daftar pengajuan cuti/izin"},
				{Name: "leave:write", MenuGroup: "Leave", Description: "Menyetujui atau menolak pengajuan cuti/izin"},
				{Name: "overtime:read", MenuGroup: "Overtime", Description: "Melihat daftar pengajuan lembur"},
				{Name: "overtime:write", MenuGroup: "Overtime", Description: "Menyetujui atau menolak pengajuan lembur"},
				{Name: "reimbursement:read", MenuGroup: "Reimbursement", Description: "Melihat daftar pengajuan reimbursement"},
				{Name: "reimbursement:write", MenuGroup: "Reimbursement", Description: "Menyetujui atau memproses pembayaran reimbursement"},
				{Name: "sales:read", MenuGroup: "Sales", Description: "Melihat data transaksi sales dan toko"},
				{Name: "sales:write", MenuGroup: "Sales", Description: "Membuat order penjualan atau manipulasi data sales"},
				{Name: "delivery:read", MenuGroup: "Delivery", Description: "Melihat dashboard tugas pengiriman surat jalan"},
				{Name: "delivery:write", MenuGroup: "Delivery", Description: "Membuat, menugaskan, dan memperbarui surat jalan pengiriman"},
				{Name: "warehouse:read", MenuGroup: "Warehouse", Description: "Melihat stok gudang dan mutasi produk"},
				{Name: "warehouse:write", MenuGroup: "Warehouse", Description: "Mengupdate stok gudang dan memproses penerimaan barang"},
				{Name: "announcement:read", MenuGroup: "Announcement", Description: "Melihat pengumuman perusahaan"},
				{Name: "announcement:write", MenuGroup: "Announcement", Description: "Membuat, mengubah, dan mempublikasikan pengumuman"},
				{Name: "view_all_branches", MenuGroup: "System", Description: "Melihat data dari semua cabang tanpa dibatasi cabang sendiri"},
			}
			if err := db.Create(&defaultPerms).Error; err != nil {
				log.Printf("Failed to seed default permissions: %v", err)
			} else {
				log.Println("Default master permissions seeded successfully")
			}
		}
	}

	// Recreate unique indexes yang dikelola manual (bukan oleh GORM) — idempotent
	db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_sales_orders_delivery_order_no ON sales_orders(delivery_order_no) WHERE delivery_order_no IS NOT NULL")
	db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_wdo_delivery_batch_id ON warehouse_delivery_orders(delivery_batch_id) WHERE delivery_batch_id IS NOT NULL")
	db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_factory_returns_return_no ON factory_returns(return_no)")

	// Ensure factory_returns approval and rejection columns exist
	db.Exec("ALTER TABLE factory_returns ADD COLUMN IF NOT EXISTS approved_by_id UUID REFERENCES employees(id) ON DELETE SET NULL")
	db.Exec("ALTER TABLE factory_returns ADD COLUMN IF NOT EXISTS approved_by_name VARCHAR(100)")
	db.Exec("ALTER TABLE factory_returns ADD COLUMN IF NOT EXISTS approved_at TIMESTAMPTZ")
	db.Exec("ALTER TABLE factory_returns ADD COLUMN IF NOT EXISTS rejection_reason TEXT")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_factory_returns_approved_by_id ON factory_returns(approved_by_id)")

	// Performance indexes for banner_orders & stores
	db.Exec("CREATE INDEX IF NOT EXISTS idx_banner_orders_created_at ON banner_orders(created_at DESC)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_banner_orders_status ON banner_orders(status)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_banner_orders_employee_id ON banner_orders(employee_id)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_banner_orders_store_id ON banner_orders(store_id)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_banner_orders_editor_id ON banner_orders(editor_id)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_banner_orders_status_created_at ON banner_orders(status, created_at DESC)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_stores_assigned_employee_id ON stores(assigned_employee_id)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_stores_is_active ON stores(is_active)")

	// Force drop unique constraint if it still exists and replace with non-unique index
	db.Exec("DROP INDEX IF EXISTS idx_product_transfers_delivery_order_no CASCADE")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_product_transfers_delivery_order_no ON product_transfers(delivery_order_no)")

	// Vehicle GPS IMEI (TraxCenter)
	db.Exec("ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS gps_imei VARCHAR(50)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_vehicles_gps_imei ON vehicles(gps_imei)")

	// Seed Chart of Accounts PSAK untuk semua perusahaan yang belum punya akun
	log.Println("Seeding Chart of Accounts (PSAK) untuk perusahaan yang belum punya akun...")
	var companies []domain.Company
	if err := db.Find(&companies).Error; err == nil {
		accountingRepo := repository.NewAccountingRepository(db)
		for _, comp := range companies {
			if seedErr := accountingRepo.SeedDefaultAccounts(comp.ID); seedErr != nil {
				log.Printf("  [SKIP] %s: %v", comp.Name, seedErr)
			} else {
				log.Printf("  [OK] Chart of Accounts untuk %s siap", comp.Name)
			}
		}
	}

	log.Println("GORM Migrations completed successfully.")
}
