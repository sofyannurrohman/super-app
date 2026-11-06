package usecase

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"example.com/hris/backend/internal/domain"
	"example.com/hris/backend/internal/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type FinanceUsecase interface {
	GetProfitLossReport(companyID uuid.UUID, month, year int, startDate, endDate string) (*domain.ProfitLossReport, error)
	GetProfitLossAllCompanies(month, year int, startDate, endDate string) ([]*domain.ProfitLossReport, error)
	GetProfitLossDetail(companyID uuid.UUID, month, year int, startDate, endDate string) (*domain.ProfitLossDetailReport, error)
	GetProfitLossByBranch(branchID uuid.UUID, month, year int, startDate, endDate string) (*domain.ProfitLossReport, error)
	GetCashierProfitLossReport(month, year int, branchID *uuid.UUID, startDate, endDate string) (*domain.CashierProfitLossReport, error)
	GetIncomeTransactionDetails(incomeType string, month, year int, branchID *uuid.UUID, startDate, endDate string) (*domain.IncomeDetailResponse, error)
	GetExpenseTransactionDetails(categoryKey string, month, year int, branchID *uuid.UUID, startDate, endDate string) (*domain.ExpenseDetailResponse, error)
	GetReceivableTransactionDetails(entityCode string, month, year int, branchID *uuid.UUID, startDate, endDate string) (*domain.ReceivableDetailResponse, error)
	GetFinancialReceivablesList(companyID uuid.UUID, branchID *uuid.UUID) (*domain.FinancialReceivablesResponse, error)
	RecordFinancialReceivablePayment(req domain.RepayBranchFinancialReceivableRequest, recordedBy *uuid.UUID) (*domain.KasTransaction, error)
	SettleFinancialReceivableTx(txID uuid.UUID, req domain.SettleFinancialReceivableTxRequest, settledBy *uuid.UUID) (*domain.KasAdvanceSettlement, error)
	GetPendingPayments(companyID, branchID uuid.UUID, page, limit int, search, status, startDate, endDate string) ([]domain.SalesPayment, int64, error)
	GetMidtransPayments(companyID, branchID uuid.UUID, page, limit int, search, status, startDate, endDate string) ([]domain.SalesPayment, int64, error)
	ApprovePayment(id uuid.UUID, targetKasPabrikID string, bankAccountID *uuid.UUID, approvedBy *uuid.UUID) error
	GetAvailableDepositInvoices(branchID, companyID *uuid.UUID, search string) ([]InvoiceDepositAvailable, error)
	GetInvoiceDepositBranch(userID uuid.UUID, role string) (*uuid.UUID, error)
	ListInvoiceDeposits(branchID *uuid.UUID) ([]domain.InvoiceDeposit, error)
	SubmitInvoiceDeposit(req SubmitInvoiceDepositRequest, submittedBy uuid.UUID, role string) (*domain.InvoiceDeposit, error)
	ReviewInvoiceDeposit(id, reviewedBy uuid.UUID, approve bool) error
	ReportDiscrepancy(id uuid.UUID, notes string) error
	RejectPayment(id uuid.UUID, reason string) error
	ExportFinancialDataCSV(companyID uuid.UUID, month, year int) ([]byte, string, error)

	GetDeliverySummary(doNo string) (map[string]interface{}, error)
	ApproveDeliveryBatch(batchID uuid.UUID, physicalCash float64) error
	ResolveCrosscheck(batchID uuid.UUID, notes string) error
	GetTodayCompletedDeliveries(branchID *uuid.UUID) ([]map[string]interface{}, error)
	GetDailySalesIncomeReport(date, startDate, endDate, companyIDStr, branchIDStr string) (*domain.DailySalesIncomeReport, error)
}

type financeUsecase struct {
	repo         repository.FinanceRepository
	salesUsecase SalesUsecase
	db           *gorm.DB
	fcmClient    FcmUsecase
}

func NewFinanceUsecase(repo repository.FinanceRepository, salesUsecase SalesUsecase, db *gorm.DB, fcmClient FcmUsecase) FinanceUsecase {
	return &financeUsecase{
		repo:         repo,
		salesUsecase: salesUsecase,
		db:           db,
		fcmClient:    fcmClient,
	}
}

func (u *financeUsecase) GetProfitLossReport(companyID uuid.UUID, month, year int, startDate, endDate string) (*domain.ProfitLossReport, error) {
	company, err := u.repo.GetCompanyByID(companyID)
	if err != nil {
		return nil, err
	}

	revenue, err := u.repo.GetRevenueByCompany(companyID, month, year, startDate, endDate)
	if err != nil {
		return nil, err
	}

	payroll, err := u.repo.GetPayrollByCompany(companyID, month, year, startDate, endDate)
	if err != nil {
		return nil, err
	}

	reimbursements, err := u.repo.GetReimbursementByCompany(companyID, month, year, startDate, endDate)
	if err != nil {
		return nil, err
	}

	// For simple P&L, assume 70% of revenue is COGS if not tracked individually
	cogs := revenue * 0.70

	report := &domain.ProfitLossReport{
		CompanyID:     companyID,
		CompanyName:   company.Name,
		PeriodMonth:   month,
		PeriodYear:    year,
		StartDate:     startDate,
		EndDate:       endDate,
		TotalRevenue:  revenue,
		TotalCOGS:     cogs,
		TotalPayroll:  payroll,
		TotalExpenses: reimbursements,
		GrossProfit:   revenue - cogs,
		NetProfit:     (revenue - cogs) - (payroll + reimbursements),
	}

	return report, nil
}

func (u *financeUsecase) GetProfitLossAllCompanies(month, year int, startDate, endDate string) ([]*domain.ProfitLossReport, error) {
	companies, err := u.repo.GetAllCompanies()
	if err != nil {
		return nil, err
	}

	var reports []*domain.ProfitLossReport
	for _, comp := range companies {
		report, err := u.GetProfitLossReport(comp.ID, month, year, startDate, endDate)
		if err != nil {
			continue
		}
		reports = append(reports, report)
	}
	return reports, nil
}

func (u *financeUsecase) GetProfitLossDetail(companyID uuid.UUID, month, year int, startDate, endDate string) (*domain.ProfitLossDetailReport, error) {
	summary, err := u.GetProfitLossReport(companyID, month, year, startDate, endDate)
	if err != nil {
		return nil, err
	}

	branches, err := u.repo.GetBranchesByCompany(companyID)
	if err != nil {
		return nil, err
	}

	var branchReports []domain.BranchProfitLoss
	for _, branch := range branches {
		rev, _ := u.repo.GetRevenueByBranch(branch.ID, month, year, startDate, endDate)
		payroll, _ := u.repo.GetPayrollByBranch(branch.ID, month, year, startDate, endDate)
		reimb, _ := u.repo.GetReimbursementByBranch(branch.ID, month, year, startDate, endDate)
		cogs := rev * 0.70

		branchReports = append(branchReports, domain.BranchProfitLoss{
			BranchID:      branch.ID,
			BranchName:    branch.Name,
			TotalRevenue:  rev,
			TotalCOGS:     cogs,
			TotalPayroll:  payroll,
			TotalExpenses: reimb,
			GrossProfit:   rev - cogs,
			NetProfit:     (rev - cogs) - (payroll + reimb),
		})
	}

	return &domain.ProfitLossDetailReport{
		CompanyID:     companyID,
		CompanyName:   summary.CompanyName,
		PeriodMonth:   month,
		PeriodYear:    year,
		StartDate:     startDate,
		EndDate:       endDate,
		TotalRevenue:  summary.TotalRevenue,
		TotalCOGS:     summary.TotalCOGS,
		TotalPayroll:  summary.TotalPayroll,
		TotalExpenses: summary.TotalExpenses,
		GrossProfit:   summary.GrossProfit,
		NetProfit:     summary.NetProfit,
		Branches:      branchReports,
	}, nil
}

func (u *financeUsecase) GetProfitLossByBranch(branchID uuid.UUID, month, year int, startDate, endDate string) (*domain.ProfitLossReport, error) {
	branch, err := u.repo.GetBranchByID(branchID)
	if err != nil {
		return nil, err
	}

	revenue, err := u.repo.GetRevenueByBranch(branchID, month, year, startDate, endDate)
	if err != nil {
		return nil, err
	}

	payroll, err := u.repo.GetPayrollByBranch(branchID, month, year, startDate, endDate)
	if err != nil {
		return nil, err
	}

	reimbursements, err := u.repo.GetReimbursementByBranch(branchID, month, year, startDate, endDate)
	if err != nil {
		return nil, err
	}

	cogs := revenue * 0.70
	companyID := uuid.Nil
	if branch.CompanyID != nil {
		companyID = *branch.CompanyID
	}

	report := &domain.ProfitLossReport{
		CompanyID:     companyID,
		CompanyName:   branch.Name,
		PeriodMonth:   month,
		PeriodYear:    year,
		StartDate:     startDate,
		EndDate:       endDate,
		TotalRevenue:  revenue,
		TotalCOGS:     cogs,
		TotalPayroll:  payroll,
		TotalExpenses: reimbursements,
		GrossProfit:   revenue - cogs,
		NetProfit:     (revenue - cogs) - (payroll + reimbursements),
	}

	return report, nil
}

func (u *financeUsecase) GetCashierProfitLossReport(month, year int, branchID *uuid.UUID, startDate, endDate string) (*domain.CashierProfitLossReport, error) {
	// 1. Get Income Breakdown from Kas
	kasIncomeMap, err := u.repo.GetKasIncomeBreakdown(month, year, branchID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	// 2. Get Sales Revenue from Orders/Transactions
	var productSales float64
	if branchID != nil && *branchID != uuid.Nil {
		productSales, _ = u.repo.GetRevenueByBranch(*branchID, month, year, startDate, endDate)
	} else {
		companies, _ := u.repo.GetAllCompanies()
		for _, comp := range companies {
			rev, _ := u.repo.GetRevenueByCompany(comp.ID, month, year, startDate, endDate)
			productSales += rev
		}
	}

	// Fallback if no transactions in SalesOrder yet, check PRODUCT_SALE in Kas
	if kasProductSale, ok := kasIncomeMap["PRODUCT_SALE"]; ok && kasProductSale > productSales {
		productSales = kasProductSale
	}

	interestIncome := kasIncomeMap["INTEREST_INCOME"]
	ampasIncome := kasIncomeMap["AMPAS_SALE"]
	rosokIncome := kasIncomeMap["ROSOK_SALE"]
	claimPenalty := kasIncomeMap["CLAIM_PENALTY"]
	otherIncome := kasIncomeMap["OTHER"]
	totalOtherIncome := interestIncome + ampasIncome + rosokIncome + claimPenalty + otherIncome
	totalRevenue := productSales + totalOtherIncome

	// 3. Get Expense Breakdown from Kas
	kasExpenseMap, err := u.repo.GetKasExpenseBreakdown(month, year, branchID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	// Standard Expense Accounts Definition
	type CatDef struct {
		Key   string
		Name  string
		Group string
	}
	allCatDefs := []CatDef{
		// Group: LABOR (Beban Tenaga Kerja & Borongan)
		{"GAJI_TUKANG_PABRIK", "Biaya Gaji Tukang Pabrik", "LABOR"},
		{"GAJI_BORONGAN_PACKING_KECAP", "Biaya Gaji Borongan Packing Kecap", "LABOR"},
		{"GAJI_BORONGAN_PACKING_SAOS", "Biaya Gaji Borongan Packing Saos", "LABOR"},
		{"GAJI_BORONGAN_OVEN", "Biaya Gaji Borongan Oven", "LABOR"},
		{"GAJI_PRODUKSI_KECAP", "Biaya Gaji Produksi Kecap", "LABOR"},
		{"GAJI_PRODUKSI_SAOS", "Biaya Gaji Produksi Saos", "LABOR"},
		{"BIAYA_BORONGAN_MUAT_KECAP", "Biaya Borongan Muat Kecap", "LABOR"},
		{"BIAYA_BORONGAN_MUAT_SAOS", "Biaya Borongan Muat Saos", "LABOR"},
		{"BIAYA_BORONGAN_MUAT_BIHUN", "Biaya Borongan Muat Bihun", "LABOR"},
		{"BIAYA_BORONGAN_MUAT_CUKA", "Biaya Borongan Muat Cuka", "LABOR"},
		{"BIAYA_BORONGAN_TRONTON", "Biaya Borongan dan Muat Tronton", "LABOR"},
		{"KULI", "Gaji Kuli Muat / Bongkar", "LABOR"},
		{"DAILY_WORKER", "Gaji Harian Mingguan", "LABOR"},
		{"LEMBUR", "Biaya Lembur", "LABOR"},
		{"LEMBUR_TEKNISI", "Biaya Lembur Teknisi", "LABOR"},
		{"BIAYA_THR", "Biaya THR", "LABOR"},
		{"BIAYA_KOMISI_BONUS", "Biaya Komisi & Bonus Karyawan", "LABOR"},
		{"BIAYA_KEHADIRAN_KARYAWAN", "Biaya Kehadiran Karyawan", "LABOR"},
		{"BIAYA_BPJS_KETENAGAKERJAAN", "Biaya BPJS Ketenagakerjaan", "LABOR"},
		{"BIAYA_BPJS_KESEHATAN", "Biaya BPJS Kesehatan", "LABOR"},

		// Group: UTILITIES (Beban Utilitas & Operasional Produksi)
		{"BIAYA_SOLAR", "Biaya Solar", "UTILITIES"},
		{"BIAYA_GAS_LPG", "Biaya Gas LPG", "UTILITIES"},
		{"DROP_GULA", "Biaya Drop Gula", "UTILITIES"},
		{"CUCI_BOTOL_LINGKUNGAN", "Biaya Cuci Botol Lingkungan", "UTILITIES"},
		{"CUCI_BOTOL_GUNUNGKEBO", "Biaya Cuci Botol Gunungkebo", "UTILITIES"},
		{"PEMBERSIHAN_PRODUKSI", "Biaya Pembersihan Tempat Produksi", "UTILITIES"},
		{"BIAYA_PERALATAN_KEBERSIHAN", "Biaya Peralatan & Perlengkapan Kebersihan", "UTILITIES"},
		{"PERLENGKAPAN_PABRIK", "Biaya Perlengkapan dan Peralatan Pabrik", "UTILITIES"},
		{"PERLENGKAPAN_TEKNISI", "Biaya Perlengkapan dan Peralatan Teknisi", "UTILITIES"},
		{"BIAYA_SEWA_TANAH_LIMBAH_SAMBIREJO", "Biaya Sewa Tanah Limbah Sambirejo", "UTILITIES"},
		{"BIAYA_PLN_PABRIK", "Biaya Listrik PLN Pabrik", "UTILITIES"},
		{"BIAYA_PLN_MESS", "Biaya Listrik PLN Mess", "UTILITIES"},
		{"BIAYA_PDAM_PABRIK", "Biaya PDAM Pabrik", "UTILITIES"},
		{"BIAYA_PAJAK_AIR_TANAH", "Biaya Pajak Air Tanah", "UTILITIES"},
		{"BIAYA_TELPON_PABRIK", "Biaya Telpon Pabrik", "UTILITIES"},
		{"BIAYA_PULSA_HP", "Biaya Pulsa HP", "UTILITIES"},
		{"BIAYA_INTERNET", "Biaya Internet", "UTILITIES"},

		// Group: FLEET (Beban Pemeliharaan, Armada & Ekspedisi)
		{"BIAYA_PEMELIHARAAN_MESIN", "Biaya Pemeliharaan Mesin", "FLEET"},
		{"BIAYA_PEMELIHARAAN_BANGUNAN_PABRIK", "Biaya Pemeliharaan Bangunan Pabrik", "FLEET"},
		{"BIAYA_PEMELIHARAAN_INVENTARIS", "Biaya Pemeliharaan Inventaris", "FLEET"},
		{"INVENTARIS", "Inventaris Pabrik", "FLEET"},
		{"BIAYA_PEMELIHARAAN_KENDARAAN", "Pemeliharaan & Servis Armada Kendaraan", "FLEET"},
		{"BIAYA_BBM_KENDARAAN", "BBM & Tol Armada Kendaraan", "FLEET"},
		{"EXPEDISI", "Biaya Ekspedisi", "FLEET"},
		{"EKSPEDISI_OTHER", "Ekspedisi Eksternal (JNT/POS/JNE, dsb)", "FLEET"},

		// Group: TAX (Beban Pajak, Legal & Perijinan)
		{"BIAYA_PPH_21", "Biaya PPh Pasal 21", "TAX"},
		{"BIAYA_PPH_22", "Biaya PPh Pasal 22", "TAX"},
		{"BIAYA_PPH_23", "Biaya PPh Pasal 23", "TAX"},
		{"BIAYA_PPH_25", "Biaya PPh Pasal 25", "TAX"},
		{"BIAYA_PPH_FINAL", "Biaya PPh Final", "TAX"},
		{"BIAYA_PPN_KELUARAN", "Biaya PPN Keluaran", "TAX"},
		{"PAJAK_PPN", "Biaya Pajak PPN", "TAX"},
		{"PAJAK_PPH", "Biaya Pajak PPH", "TAX"},
		{"BIAYA_PBB", "Biaya PBB", "TAX"},
		{"BIAYA_PERIJINAN", "Biaya Perijinan & Legal", "TAX"},
		{"BIAYA_RETRIBUSI_PARKIR", "Biaya Retribusi Parkir", "TAX"},

		// Group: GENERAL (Beban Umum & Administrasi)
		{"PERLENGKAPAN_KANTOR", "Biaya Perlengkapan Kantor", "GENERAL"},
		{"BIAYA_FOTOCOPY", "Biaya Fotocopy", "GENERAL"},
		{"BIAYA_MATERAI", "Biaya Materai", "GENERAL"},
		{"BIAYA_BUNGA_BANK", "Biaya Bunga Bank", "GENERAL"},
		{"ADMIN_BANK", "Biaya Admin Bank", "GENERAL"},
		{"PROMOSI", "Biaya Promosi", "GENERAL"},
		{"BIAYA_BANNER_MMT", "Biaya Banner MMT", "GENERAL"},
		{"BIAYA_SAMPLE", "Biaya Sample", "GENERAL"},
		{"BIAYA_KONSUMSI", "Biaya Konsumsi", "GENERAL"},
		{"UANG_MAKAN_TEKNISI", "Biaya Uang Makan Teknisi", "GENERAL"},
		{"BIAYA_PERAWATAN_MEDIS", "Biaya Perawatan Medis", "GENERAL"},
		{"BIAYA_ASURANSI", "Biaya Asuransi", "GENERAL"},
		{"PERJALANAN_DINAS", "Biaya Perjalanan Dinas Umum", "GENERAL"},
		{"BIAYA_BBM_PERJALANAN_DINAS", "Biaya BBM Perjalanan Dinas", "GENERAL"},
		{"BIAYA_PERJALANAN_DINAS_AUDIT", "Biaya Perjalanan Dinas Audit", "GENERAL"},
		{"ENTERTAINT", "Biaya Entertaint", "GENERAL"},
		{"CSR", "Sumbangan / CSR", "GENERAL"},
		{"BIAYA_REIMBURSE", "Biaya Reimburse", "GENERAL"},
		{"BEBAN_LAIN_PRODUKSI", "Beban Lain-Lain Produksi", "GENERAL"},
		{"BEBAN_LAIN", "Beban Lain-Lain", "GENERAL"},
		{"OTHER", "Beban Pengeluaran Lainnya", "GENERAL"},
	}

	var expenseItems []domain.ExpenseAccountItem
	var totalExpenses float64
	for _, def := range allCatDefs {
		amt := kasExpenseMap[def.Key]
		expenseItems = append(expenseItems, domain.ExpenseAccountItem{
			CategoryKey: def.Key,
			Name:        def.Name,
			Group:       def.Group,
			Amount:      amt,
		})
		totalExpenses += amt
	}

	// 4. Receivables Matrix (28 entities)
	receivables, err := u.repo.GetReceivablesBreakdown(month, year, branchID, startDate, endDate)
	if err != nil {
		receivables = []domain.ReceivableEntityItem{}
	}
	var totalReceivablesSum float64
	for _, r := range receivables {
		totalReceivablesSum += r.EndingBalance
	}

	// 5. Calculations
	cogs := productSales * 0.70 // Estimated direct material COGS
	grossProfit := totalRevenue - cogs
	netProfit := grossProfit - totalExpenses
	var profitMargin float64
	if totalRevenue > 0 {
		profitMargin = (netProfit / totalRevenue) * 100
	}

	branchName := "Seluruh Pabrik & Unit"
	if branchID != nil && *branchID != uuid.Nil {
		b, _ := u.repo.GetBranchByID(*branchID)
		if b != nil {
			branchName = b.Name
		}
	}

	return &domain.CashierProfitLossReport{
		PeriodMonth: month,
		PeriodYear:  year,
		StartDate:   startDate,
		EndDate:     endDate,
		BranchID:    branchID,
		BranchName:  branchName,
		Income: domain.IncomeBreakdown{
			ProductSales:     productSales,
			InterestIncome:   interestIncome,
			AmpasIncome:      ampasIncome,
			RosokIncome:      rosokIncome,
			ClaimPenalty:     claimPenalty,
			OtherIncome:      otherIncome,
			TotalOtherIncome: totalOtherIncome,
			TotalRevenue:     totalRevenue,
		},
		TotalCOGS:           cogs,
		GrossProfit:         grossProfit,
		Expenses:            expenseItems,
		TotalExpenses:       totalExpenses,
		TotalPayrollDirect:  0,
		NetProfit:           netProfit,
		ProfitMargin:        profitMargin,
		Receivables:         receivables,
		TotalReceivablesSum: totalReceivablesSum,
	}, nil
}

func (u *financeUsecase) GetIncomeTransactionDetails(incomeType string, month, year int, branchID *uuid.UUID, startDate, endDate string) (*domain.IncomeDetailResponse, error) {
	items, err := u.repo.GetIncomeTransactionDetails(incomeType, month, year, branchID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	titleMap := map[string]string{
		"PRODUCT_SALE":    "1.1 Pendapatan Penjualan Produk Pabrik (Faktur / Telemarketing)",
		"INTEREST_INCOME": "1.2 Pendapatan Bunga Bank",
		"AMPAS_SALE":      "1.3 Pendapatan Ampas (Penjualan Ampas Pabrik)",
		"ROSOK_SALE":      "1.4 Penjualan Rosok / Afval (Jurigen, Drum, dll)",
		"CLAIM_PENALTY":   "1.5 Penerimaan Klaim / Denda Penalty",
		"OTHER":           "1.6 Pendapatan Lain-Lain Umum",
	}

	upperType := strings.ToUpper(strings.TrimSpace(incomeType))
	title := titleMap[upperType]
	if title == "" {
		title = "Rincian Transaksi Pendapatan"
	}

	var totalAmt float64
	for _, it := range items {
		totalAmt += it.Amount
	}

	return &domain.IncomeDetailResponse{
		IncomeType:  upperType,
		IncomeTitle: title,
		PeriodMonth: month,
		PeriodYear:  year,
		StartDate:   startDate,
		EndDate:     endDate,
		BranchID:    branchID,
		TotalAmount: totalAmt,
		TotalCount:  len(items),
		Items:       items,
	}, nil
}

func (u *financeUsecase) GetExpenseTransactionDetails(categoryKey string, month, year int, branchID *uuid.UUID, startDate, endDate string) (*domain.ExpenseDetailResponse, error) {
	items, err := u.repo.GetExpenseTransactionDetails(categoryKey, month, year, branchID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	expenseMap := map[string]struct {
		Name  string
		Group string
	}{
		"GAJI_TUKANG_PABRIK":                 {"Biaya Gaji Tukang Pabrik", "LABOR"},
		"GAJI_BORONGAN_PACKING_KECAP":        {"Biaya Gaji Borongan Packing Kecap", "LABOR"},
		"GAJI_BORONGAN_PACKING_SAOS":         {"Biaya Gaji Borongan Packing Saos", "LABOR"},
		"GAJI_BORONGAN_OVEN":                 {"Biaya Gaji Borongan Oven", "LABOR"},
		"GAJI_PRODUKSI_KECAP":                {"Biaya Gaji Produksi Kecap", "LABOR"},
		"GAJI_PRODUKSI_SAOS":                 {"Biaya Gaji Produksi Saos", "LABOR"},
		"BIAYA_BORONGAN_MUAT_KECAP":          {"Biaya Borongan Muat Kecap", "LABOR"},
		"BIAYA_BORONGAN_MUAT_SAOS":           {"Biaya Borongan Muat Saos", "LABOR"},
		"BIAYA_BORONGAN_MUAT_BIHUN":          {"Biaya Borongan Muat Bihun", "LABOR"},
		"BIAYA_BORONGAN_MUAT_CUKA":           {"Biaya Borongan Muat Cuka", "LABOR"},
		"BIAYA_BORONGAN_TRONTON":             {"Biaya Borongan dan Muat Tronton", "LABOR"},
		"KULI":                               {"Gaji Kuli Muat / Bongkar", "LABOR"},
		"DAILY_WORKER":                       {"Gaji Harian Mingguan", "LABOR"},
		"LEMBUR":                             {"Biaya Lembur", "LABOR"},
		"LEMBUR_TEKNISI":                     {"Biaya Lembur Teknisi", "LABOR"},
		"BIAYA_THR":                          {"Biaya THR", "LABOR"},
		"BIAYA_KOMISI_BONUS":                 {"Biaya Komisi & Bonus Karyawan", "LABOR"},
		"BIAYA_KEHADIRAN_KARYAWAN":           {"Biaya Kehadiran Karyawan", "LABOR"},
		"BIAYA_BPJS_KETENAGAKERJAAN":         {"Biaya BPJS Ketenagakerjaan", "LABOR"},
		"BIAYA_BPJS_KESEHATAN":               {"Biaya BPJS Kesehatan", "LABOR"},
		"BIAYA_SOLAR":                        {"Biaya Solar", "UTILITIES"},
		"BIAYA_GAS_LPG":                      {"Biaya Gas LPG", "UTILITIES"},
		"DROP_GULA":                          {"Biaya Drop Gula", "UTILITIES"},
		"CUCI_BOTOL_LINGKUNGAN":              {"Biaya Cuci Botol Lingkungan", "UTILITIES"},
		"CUCI_BOTOL_GUNUNGKEBO":              {"Biaya Cuci Botol Gunungkebo", "UTILITIES"},
		"PEMBERSIHAN_PRODUKSI":               {"Biaya Pembersihan Tempat Produksi", "UTILITIES"},
		"BIAYA_PERALATAN_KEBERSIHAN":         {"Biaya Peralatan & Perlengkapan Kebersihan", "UTILITIES"},
		"PERLENGKAPAN_PABRIK":                {"Biaya Perlengkapan dan Peralatan Pabrik", "UTILITIES"},
		"PERLENGKAPAN_TEKNISI":               {"Biaya Perlengkapan dan Peralatan Teknisi", "UTILITIES"},
		"BIAYA_SEWA_TANAH_LIMBAH_SAMBIREJO":  {"Biaya Sewa Tanah Limbah Sambirejo", "UTILITIES"},
		"BIAYA_PLN_PABRIK":                   {"Biaya Listrik PLN Pabrik", "UTILITIES"},
		"BIAYA_PLN_MESS":                     {"Biaya Listrik PLN Mess", "UTILITIES"},
		"BIAYA_PDAM_PABRIK":                  {"Biaya PDAM Pabrik", "UTILITIES"},
		"BIAYA_PAJAK_AIR_TANAH":              {"Biaya Pajak Air Tanah", "UTILITIES"},
		"BIAYA_TELPON_PABRIK":                {"Biaya Telpon Pabrik", "UTILITIES"},
		"BIAYA_PULSA_HP":                     {"Biaya Pulsa HP", "UTILITIES"},
		"BIAYA_INTERNET":                     {"Biaya Internet", "UTILITIES"},
		"BIAYA_PEMELIHARAAN_MESIN":           {"Biaya Pemeliharaan Mesin", "FLEET"},
		"BIAYA_PEMELIHARAAN_BANGUNAN_PABRIK": {"Biaya Pemeliharaan Bangunan Pabrik", "FLEET"},
		"BIAYA_PEMELIHARAAN_INVENTARIS":      {"Biaya Pemeliharaan Inventaris", "FLEET"},
		"INVENTARIS":                         {"Inventaris Pabrik", "FLEET"},
		"BIAYA_PEMELIHARAAN_KENDARAAN":       {"Pemeliharaan & Servis Armada Kendaraan", "FLEET"},
		"BIAYA_BBM_KENDARAAN":                {"BBM & Tol Armada Kendaraan", "FLEET"},
		"EXPEDISI":                           {"Biaya Ekspedisi", "FLEET"},
		"EKSPEDISI_OTHER":                    {"Ekspedisi Eksternal (JNT/POS/JNE, dsb)", "FLEET"},
		"BIAYA_PPH_21":                       {"Biaya PPh Pasal 21", "TAX"},
		"BIAYA_PPH_22":                       {"Biaya PPh Pasal 22", "TAX"},
		"BIAYA_PPH_23":                       {"Biaya PPh Pasal 23", "TAX"},
		"BIAYA_PPH_25":                       {"Biaya PPh Pasal 25", "TAX"},
		"BIAYA_PPH_FINAL":                    {"Biaya PPh Final", "TAX"},
		"BIAYA_PPN_KELUARAN":                 {"Biaya PPN Keluaran", "TAX"},
		"PAJAK_PPN":                          {"Biaya Pajak PPN", "TAX"},
		"PAJAK_PPH":                          {"Biaya Pajak PPH", "TAX"},
		"BIAYA_PBB":                          {"Biaya PBB", "TAX"},
		"BIAYA_PERIJINAN":                    {"Biaya Perijinan & Legal", "TAX"},
		"BIAYA_RETRIBUSI_PARKIR":             {"Biaya Retribusi Parkir", "TAX"},
		"PERLENGKAPAN_KANTOR":                {"Biaya Perlengkapan Kantor", "GENERAL"},
		"BIAYA_FOTOCOPY":                     {"Biaya Fotocopy", "GENERAL"},
		"BIAYA_MATERAI":                      {"Biaya Materai", "GENERAL"},
		"BIAYA_BUNGA_BANK":                   {"Biaya Bunga Bank", "GENERAL"},
		"ADMIN_BANK":                         {"Biaya Admin Bank", "GENERAL"},
		"PROMOSI":                            {"Biaya Promosi", "GENERAL"},
		"BIAYA_BANNER_MMT":                   {"Biaya Banner MMT", "GENERAL"},
		"BIAYA_SAMPLE":                       {"Biaya Sample", "GENERAL"},
		"BIAYA_KONSUMSI":                     {"Biaya Konsumsi", "GENERAL"},
		"UANG_MAKAN_TEKNISI":                 {"Biaya Uang Makan Teknisi", "GENERAL"},
		"BIAYA_PERAWATAN_MEDIS":              {"Biaya Perawatan Medis", "GENERAL"},
		"BIAYA_ASURANSI":                     {"Biaya Asuransi", "GENERAL"},
		"PERJALANAN_DINAS":                   {"Biaya Perjalanan Dinas Umum", "GENERAL"},
		"BIAYA_BBM_PERJALANAN_DINAS":         {"Biaya BBM Perjalanan Dinas", "GENERAL"},
		"BIAYA_PERJALANAN_DINAS_AUDIT":       {"Biaya Perjalanan Dinas Audit", "GENERAL"},
		"ENTERTAINT":                         {"Biaya Entertaint", "GENERAL"},
		"CSR":                                {"Sumbangan / CSR", "GENERAL"},
		"BIAYA_REIMBURSE":                    {"Biaya Reimburse", "GENERAL"},
		"BEBAN_LAIN_PRODUKSI":                {"Beban Lain-Lain Produksi", "GENERAL"},
		"BEBAN_LAIN":                         {"Beban Lain-Lain", "GENERAL"},
		"OTHER":                              {"Beban Pengeluaran Lainnya", "GENERAL"},
	}

	catName := categoryKey
	group := "GENERAL"
	if info, ok := expenseMap[categoryKey]; ok {
		catName = info.Name
		group = info.Group
	}

	var totalAmt float64
	for _, it := range items {
		totalAmt += it.Amount
	}

	return &domain.ExpenseDetailResponse{
		CategoryKey:  categoryKey,
		CategoryName: catName,
		Group:        group,
		PeriodMonth:  month,
		PeriodYear:   year,
		StartDate:    startDate,
		EndDate:      endDate,
		BranchID:     branchID,
		TotalAmount:  totalAmt,
		TotalCount:   len(items),
		Items:        items,
	}, nil
}

func (u *financeUsecase) GetReceivableTransactionDetails(entityCode string, month, year int, branchID *uuid.UUID, startDate, endDate string) (*domain.ReceivableDetailResponse, error) {
	return u.repo.GetReceivableTransactionDetails(entityCode, month, year, branchID, startDate, endDate)
}

func (u *financeUsecase) GetPendingPayments(companyID, branchID uuid.UUID, page, limit int, search, status, startDate, endDate string) ([]domain.SalesPayment, int64, error) {
	payments, total, err := u.repo.GetPendingPayments(companyID, branchID, page, limit, search, status, startDate, endDate)
	if err != nil {
		return nil, 0, err
	}

	// Auto-heal / synchronize pending payment amount with linked verified/adjusted invoice or sales order
	for i := range payments {
		p := &payments[i]
		if p.PaymentStatus == "PENDING" {
			var expectedAmount *float64
			if p.SalesTransaction != nil && p.SalesTransaction.PaymentMethod != "TEMPO" && p.SalesTransaction.TotalAmount > 0 {
				expectedAmount = &p.SalesTransaction.TotalAmount
			} else if p.SalesOrder != nil && p.SalesOrder.PaymentMethod != "TEMPO" && p.SalesOrder.TotalAmount > 0 {
				expectedAmount = &p.SalesOrder.TotalAmount
			}

			// If expected amount is different and payment is CASH or TRANSFER
			if expectedAmount != nil && *expectedAmount != p.Amount {
				if p.PaymentMethod == "CASH" || p.PaymentMethod == "TRANSFER" {
					newAmt := *expectedAmount
					p.Amount = newAmt
					_ = u.db.Model(&domain.SalesPayment{}).Where("id = ?", p.ID).Update("amount", newAmt)
				}
			}
		}
	}

	return payments, total, nil
}

func (u *financeUsecase) GetMidtransPayments(companyID, branchID uuid.UUID, page, limit int, search, status, startDate, endDate string) ([]domain.SalesPayment, int64, error) {
	return u.repo.GetMidtransPayments(companyID, branchID, page, limit, search, status, startDate, endDate)
}

func (u *financeUsecase) ApprovePayment(id uuid.UUID, targetKasPabrikID string, bankAccountID *uuid.UUID, approvedBy *uuid.UUID) error {
	var salesmanID uuid.UUID
	var transactionMonth int
	var transactionYear int
	var approvedAmount float64

	err := u.db.Transaction(func(tx *gorm.DB) error {
		// 1. Get payment
		var payment domain.SalesPayment
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&payment, "id = ?", id).Error; err != nil {
			return err
		}
		if payment.PaymentStatus != "PENDING" {
			return errors.New("pembayaran tidak dalam status PENDING")
		}
		if payment.PaymentMethod == "TRANSFER" {
			if bankAccountID != nil {
				payment.CompanyBankAccountID = bankAccountID
			}
			if payment.CompanyBankAccountID == nil {
				return errors.New("rekening bank penerima wajib dipilih")
			}
			var bank domain.CompanyBankAccount
			if err := tx.First(&bank, "id = ? AND is_active = ?", *payment.CompanyBankAccountID, true).Error; err != nil {
				return errors.New("rekening bank penerima tidak ditemukan atau tidak aktif")
			}
		} else if bankAccountID != nil {
			return errors.New("rekening bank penerima hanya untuk pembayaran transfer")
		}

		approvedAmount = payment.Amount

		// 2. Update payment status to SUCCESS
		payment.PaymentStatus = "SUCCESS"
		payment.IsVerified = true
		if approvedBy != nil {
			payment.VerifiedByID = approvedBy
		}
		now := time.Now()
		payment.UpdatedAt = now
		payment.ApprovedAt = &now
		if payment.PenerimaanStatus == "" {
			payment.PenerimaanStatus = "PENDING"
		}
		if err := tx.Save(&payment).Error; err != nil {
			return err
		}
		if payment.PaymentMethod == "TRANSFER" && payment.CompanyBankAccountID != nil {
			if err := tx.Model(&domain.CompanyBankAccount{}).
				Where("id = ?", *payment.CompanyBankAccountID).
				Updates(map[string]interface{}{
					"manual_balance":      gorm.Expr("COALESCE(manual_balance, 0) + ?", payment.Amount),
					"last_balance_update": now,
				}).Error; err != nil {
				return err
			}
		}

		// 3. Update SalesTransaction or SalesOrder Paid Amount
		if payment.SalesOrderID != nil {
			var so domain.SalesOrder
			if err := tx.First(&so, "id = ?", *payment.SalesOrderID).Error; err != nil {
				return err
			}

			// Check if invoice has been generated
			var trx domain.SalesTransaction
			var hasInvoice bool
			if so.InvoiceID != nil {
				if err := tx.First(&trx, "id = ?", *so.InvoiceID).Error; err == nil {
					hasInvoice = true
				}
			}

			salesmanID = so.EmployeeID // Extract the actual salesman ID from the SO
			// Use the order's actual period for KPI calculation
			transactionMonth = int(so.OrderDate.Month())
			transactionYear = so.OrderDate.Year()

			// Payment collected amount and status are now updated immediately in CollectPayment
			// to reflect on the driver app instantly. Thus, we skip incrementing it here.
			so.PaymentMethod = payment.PaymentMethod
			now := time.Now()
			so.PaymentCollectedAt = &now

			if err := tx.Save(&so).Error; err != nil {
				return err
			}

			// Update corresponding SalesTransaction (Invoice) status and paid amount if exists
			if hasInvoice {
				var successPaidSum float64
				q := tx.Model(&domain.SalesPayment{}).Where("payment_status = 'SUCCESS'")
				q = q.Where("sales_transaction_id = ? OR sales_order_id = ?", trx.ID, so.ID)
				q.Select("COALESCE(SUM(amount), 0)").Scan(&successPaidSum)

				paymentStatus := string(domain.PaymentStatusPartial)
				if successPaidSum >= trx.TotalAmount {
					paymentStatus = string(domain.PaymentStatusPaid)
					trx.PaymentDueDate = nil
				}
				trx.PaymentStatus = paymentStatus
				trx.PaidAmount = successPaidSum
				trx.PaymentMethod = payment.PaymentMethod
				trx.UpdatedAt = time.Now()

				if err := tx.Omit("Items", "Installments", "DeliveryItems").Save(&trx).Error; err != nil {
					return err
				}
				if so.AdminNotaConfirmedAt != nil && trx.PaymentStatus == string(domain.PaymentStatusPaid) {
					if err := tx.Model(&trx).Update("status", "APPROVED").Error; err != nil {
						return err
					}
				}
			}

			// Update DeliveryItem status to PAID if all delivered
			var di domain.DeliveryItem
			if err := tx.Where("sales_order_id = ?", so.ID).First(&di).Error; err == nil {
				di.PaymentCollected = true
				// Note: di.PaymentAmount and batch.TotalCashCollected are already updated immediately in CollectPayment.
				// We update the method and timestamp here but skip incrementing to prevent double-counting.
				di.PaymentMethod = payment.PaymentMethod
				di.PaymentCollectedAt = &now
				_ = tx.Save(&di)
			}
		}

		if payment.SalesTransactionID != nil {
			var trx domain.SalesTransaction
			if err := tx.Preload("Installments").First(&trx, "id = ?", *payment.SalesTransactionID).Error; err != nil {
				return err
			}

			salesmanID = trx.EmployeeID
			transactionMonth = trx.PeriodMonth
			transactionYear = trx.PeriodYear

			fmt.Printf("[APPROVE-DEBUG] paymentID=%s method=%s amount=%.0f trxID=%s paid_before=%.0f total=%.0f\n",
				payment.ID.String()[:8], payment.PaymentMethod, payment.Amount,
				trx.ID.String()[:8], trx.PaidAmount, trx.TotalAmount)

			// Both initial DP and subsequent cash payments are now updated in the DB immediately
			// during CreateTransaction and RecordPayment. Thus, we skip incrementing PaidAmount
			// and preloading/distributing installments here to prevent double-counting.
			fmt.Printf("[APPROVE-DEBUG] Skipping PaidAmount and installment increment since it's already updated immediately\n")

			fmt.Printf("[APPROVE-DEBUG] paid_after=%.0f status=%s->%s\n",
				trx.PaidAmount, trx.PaymentStatus, func() string {
					if trx.PaidAmount >= trx.TotalAmount {
						return "PAID"
					}
					return "PARTIAL"
				}())

			// Recalculate based on SUCCESS payments only
			var successPaidSum float64
			var soID *uuid.UUID
			var so domain.SalesOrder
			if err := tx.Select("id").Where("invoice_id = ?", trx.ID).First(&so).Error; err == nil {
				soID = &so.ID
			}

			q := tx.Model(&domain.SalesPayment{}).Where("payment_status = 'SUCCESS'")
			if soID != nil {
				q = q.Where("sales_transaction_id = ? OR sales_order_id = ?", trx.ID, *soID)
			} else {
				q = q.Where("sales_transaction_id = ?", trx.ID)
			}
			q.Select("COALESCE(SUM(amount), 0)").Scan(&successPaidSum)

			trx.PaidAmount = successPaidSum
			paymentStatus := string(domain.PaymentStatusPartial)
			if successPaidSum >= trx.TotalAmount && trx.TotalAmount > 0 {
				paymentStatus = string(domain.PaymentStatusPaid)
				trx.PaymentDueDate = nil
			}
			trx.PaymentStatus = paymentStatus
			trx.UpdatedAt = time.Now()

			if err := tx.Omit("Items", "Installments", "DeliveryItems").Save(&trx).Error; err != nil {
				return err
			}
		}

		// Check for delivery batch cash settlement
		var batchID *uuid.UUID
		if payment.SalesOrderID != nil {
			var so domain.SalesOrder
			if err := tx.First(&so, "id = ?", *payment.SalesOrderID).Error; err == nil {
				batchID = so.DeliveryBatchID
			}
		} else if payment.SalesTransactionID != nil {
			var di domain.DeliveryItem
			if err := tx.Where("sales_transaction_id = ?", *payment.SalesTransactionID).First(&di).Error; err == nil {
				batchID = &di.DeliveryBatchID
			}
		}

		if batchID != nil {
			var pendingCount int64
			tx.Table("sales_payments").
				Joins("join sales_orders on sales_orders.id = sales_payments.sales_order_id").
				Where("sales_orders.delivery_batch_id = ? AND sales_payments.payment_status = 'PENDING' AND sales_payments.payment_method = 'CASH'", *batchID).
				Count(&pendingCount)

			var pendingTrxCount int64
			tx.Table("sales_payments").
				Joins("join delivery_items on delivery_items.sales_transaction_id = sales_payments.sales_transaction_id").
				Where("delivery_items.delivery_batch_id = ? AND sales_payments.payment_status = 'PENDING' AND sales_payments.payment_method = 'CASH'", *batchID).
				Count(&pendingTrxCount)

			if pendingCount == 0 && pendingTrxCount == 0 {
				var batch domain.DeliveryBatch
				if err := tx.First(&batch, *batchID).Error; err == nil {
					if batch.CashSettledAt == nil {
						now := time.Now()
						batch.CashSettledAt = &now
						_ = tx.Save(&batch)

						if batch.DriverID != nil && u.fcmClient != nil {
							var driver domain.Employee
							if errDb := tx.First(&driver, "id = ?", *batch.DriverID).Error; errDb == nil {
								doNo := "N/A"
								if batch.DeliveryOrderNo != nil {
									doNo = *batch.DeliveryOrderNo
								}
								go func() {
									_ = u.fcmClient.SendPushNotification(driver.UserID, "Setoran Pengiriman Selesai", "Setoran uang fisik untuk tugas pengiriman "+doNo+" telah diverifikasi dan disetujui oleh Kasir.")
								}()
							}
						}
					}
				}
			}
		}

		// 4. Automatic Kas Pabrik Balance Integration
		if targetKasPabrikID != "" {
			if kasUUID, err := uuid.Parse(targetKasPabrikID); err == nil {
				var kas domain.KasPabrik
				if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&kas, "id = ?", kasUUID).Error; err == nil {
					balBefore := kas.Balance
					balAfter := kas.Balance + payment.Amount
					kas.Balance = balAfter
					_ = tx.Save(&kas)

					proofURL := ""
					if payment.TransferProofURL != nil {
						proofURL = *payment.TransferProofURL
					}

					desc := fmt.Sprintf("Penerimaan pembayaran penjualan (%s) - Rp %.0f", payment.PaymentMethod, payment.Amount)
					if payment.Notes != "" {
						desc += " (" + payment.Notes + ")"
					}

					kasTx := domain.KasTransaction{
						KasPabrikID:     kas.ID,
						TransactionType: "IN",
						Amount:          payment.Amount,
						BalanceBefore:   balBefore,
						BalanceAfter:    balAfter,
						IncomeSource:    "PRODUCT_SALE",
						SalesPaymentID:  &payment.ID,
						SalesOrderID:    payment.SalesOrderID,
						Description:     desc,
						ProofURL:        proofURL,
						TransactionDate: time.Now(),
						CreatedBy:       approvedBy,
					}
					_ = tx.Create(&kasTx)
				}
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	// 4. Trigger sales KPI & target achievement progress update!
	if salesmanID != uuid.Nil && transactionMonth != 0 && transactionYear != 0 {
		go func() {
			err := u.salesUsecase.CalculateKPI(salesmanID, transactionMonth, transactionYear)
			if err == nil && u.fcmClient != nil {
				var emp domain.Employee
				if errEmp := u.db.First(&emp, "id = ?", salesmanID).Error; errEmp == nil {
					var kpi domain.SalesKPI
					totalOmzetBulanIni := approvedAmount
					if errKpi := u.db.Where("employee_id = ? AND period_month = ? AND period_year = ?", salesmanID, transactionMonth, transactionYear).First(&kpi).Error; errKpi == nil {
						totalOmzetBulanIni = kpi.AchievedOmzet
					}

					title := "Faktur Penjualan Disetujui"
					body := fmt.Sprintf("Faktur Anda sebesar Rp %.0f telah disetujui oleh Kasir. Total omzet Anda bulan ini menjadi Rp %.0f.", approvedAmount, totalOmzetBulanIni)
					_ = u.fcmClient.SendPushNotification(emp.UserID, title, body)
				}
			}
		}()
	}

	return nil
}

func (u *financeUsecase) RejectPayment(id uuid.UUID, reason string) error {
	var payment domain.SalesPayment
	if err := u.db.Preload("SalesTransaction.Installments").Preload("SalesOrder").First(&payment, "id = ?", id).Error; err != nil {
		return err
	}
	if payment.PaymentStatus != "PENDING" {
		return errors.New("hanya pembayaran PENDING yang bisa ditolak")
	}

	payment.PaymentStatus = "FAILED"
	if reason != "" {
		payment.RejectionReason = &reason
	}
	payment.UpdatedAt = time.Now()

	err := u.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&payment).Error; err != nil {
			return err
		}

		if payment.SalesTransactionID != nil && payment.SalesTransaction != nil {
			trx := payment.SalesTransaction
			if payment.PaymentMethod == "CASH" {
				trx.PaidAmount -= payment.Amount
				if trx.PaidAmount < 0 {
					trx.PaidAmount = 0
				}

				// Recalculate based on SUCCESS payments only
				var successPaidSum float64
				var soID *uuid.UUID
				var so domain.SalesOrder
				if err := tx.Select("id").Where("invoice_id = ?", trx.ID).First(&so).Error; err == nil {
					soID = &so.ID
				}

				q := tx.Model(&domain.SalesPayment{}).Where("payment_status = 'SUCCESS'")
				if soID != nil {
					q = q.Where("sales_transaction_id = ? OR sales_order_id = ?", trx.ID, *soID)
				} else {
					q = q.Where("sales_transaction_id = ?", trx.ID)
				}
				q.Select("COALESCE(SUM(amount), 0)").Scan(&successPaidSum)

				paymentStatus := string(domain.PaymentStatusUnpaid)
				if successPaidSum >= trx.TotalAmount {
					paymentStatus = string(domain.PaymentStatusPaid)
				} else if successPaidSum > 0 || trx.PaidAmount > 0 {
					paymentStatus = string(domain.PaymentStatusPartial)
				}
				trx.PaymentStatus = paymentStatus
				trx.UpdatedAt = time.Now()

				// Revert payment from installments
				remainingRevert := payment.Amount
				for i := len(trx.Installments) - 1; i >= 0; i-- {
					inst := &trx.Installments[i]
					if inst.PaidAmount > 0 {
						if remainingRevert >= inst.PaidAmount {
							remainingRevert -= inst.PaidAmount
							inst.PaidAmount = 0
							inst.Status = "UNPAID"
							inst.UpdatedAt = time.Now()
							_ = tx.Save(inst)
						} else if remainingRevert > 0 {
							inst.PaidAmount -= remainingRevert
							remainingRevert = 0
							inst.Status = "UNPAID"
							inst.UpdatedAt = time.Now()
							_ = tx.Save(inst)
							break
						}
					}
				}

				if err := tx.Omit("Items", "Installments", "DeliveryItems").Save(trx).Error; err != nil {
					return err
				}
			}
		}

		if payment.SalesOrderID != nil && payment.SalesOrder != nil {
			so := payment.SalesOrder
			if payment.PaymentMethod == "CASH" {
				so.PaymentCollectedAmount -= payment.Amount
				if so.PaymentCollectedAmount < 0 {
					so.PaymentCollectedAmount = 0
				}

				if so.PaymentCollectedAmount >= so.TotalAmount {
					so.PaymentStatus = "PAID"
					so.Status = domain.SOStatusPaid
				} else if so.PaymentCollectedAmount > 0 {
					so.PaymentStatus = "PARTIAL"
					so.Status = domain.SOStatusDelivered
				} else {
					so.PaymentStatus = "UNPAID"
					so.Status = domain.SOStatusDelivered
				}
				so.PaymentCollectedAt = nil
				so.UpdatedAt = time.Now()

				if err := tx.Save(so).Error; err != nil {
					return err
				}

				// Revert corresponding SalesTransaction (Invoice) status and paid amount if exists
				if so.InvoiceID != nil {
					var trx domain.SalesTransaction
					if err := tx.First(&trx, "id = ?", *so.InvoiceID).Error; err == nil {
						var successPaidSum float64
						q := tx.Model(&domain.SalesPayment{}).Where("payment_status = 'SUCCESS'")
						q = q.Where("sales_transaction_id = ? OR sales_order_id = ?", trx.ID, so.ID)
						q.Select("COALESCE(SUM(amount), 0)").Scan(&successPaidSum)

						paymentStatus := string(domain.PaymentStatusUnpaid)
						if successPaidSum >= trx.TotalAmount {
							paymentStatus = string(domain.PaymentStatusPaid)
						} else if successPaidSum > 0 {
							paymentStatus = string(domain.PaymentStatusPartial)
						}
						trx.PaymentStatus = paymentStatus
						trx.PaidAmount = successPaidSum
						trx.UpdatedAt = time.Now()

						if err := tx.Omit("Items", "Installments", "DeliveryItems").Save(&trx).Error; err != nil {
							return err
						}
					}
				}
			}
		}

		return nil
	})

	if err == nil && u.fcmClient != nil {
		var emp domain.Employee
		if e := u.db.First(&emp, "id = ?", payment.EmployeeID).Error; e == nil && emp.UserID != uuid.Nil {
			title := "Setoran Ditolak"
			body := "Setoran Anda sebesar Rp" + fmt.Sprintf("%.0f", payment.Amount) + " telah ditolak oleh Kasir."
			if reason != "" {
				body += " Alasan: " + reason
			}
			_ = u.fcmClient.SendPushNotification(emp.UserID, title, body)
		}
	}

	return err
}

func (u *financeUsecase) ExportFinancialDataCSV(companyID uuid.UUID, month, year int) ([]byte, string, error) {
	company, err := u.repo.GetCompanyByID(companyID)
	if err != nil {
		return nil, "", err
	}

	// 1. Get transaction list (revenues) for the period
	var transactions []domain.SalesTransaction
	err = u.db.Preload("Store").Preload("Employee").
		Where("company_id = ? AND period_month = ? AND period_year = ? AND status = ?", companyID, month, year, "VERIFIED").
		Find(&transactions).Error
	if err != nil {
		return nil, "", err
	}

	// 1b. Get sales orders (revenues) for the period
	var salesOrders []domain.SalesOrder
	err = u.db.Preload("Store").Preload("Employee").
		Where("company_id = ? AND EXTRACT(MONTH FROM order_date) = ? AND EXTRACT(YEAR FROM order_date) = ? AND status IN ?", companyID, month, year, []string{string(domain.SOStatusDelivered), string(domain.SOStatusPaid), string(domain.SOStatusConverted), string(domain.SOStatusShipped)}).
		Find(&salesOrders).Error
	if err != nil {
		return nil, "", err
	}

	// 2. Get payslips for the period
	var payslips []domain.Payslip
	err = u.db.Preload("Employee").Preload("PayrollRun").
		Joins("JOIN payroll_runs ON payroll_runs.id = payslips.payroll_run_id").
		Where("payroll_runs.company_id = ? AND EXTRACT(MONTH FROM payroll_runs.period_start) = ? AND EXTRACT(YEAR FROM payroll_runs.period_start) = ? AND payroll_runs.status != ?", companyID, month, year, "DRAFT").
		Find(&payslips).Error
	if err != nil {
		return nil, "", err
	}

	// 3. Get reimbursements
	var reimbursements []domain.Reimbursement
	err = u.db.Preload("Employee").
		Joins("JOIN employees ON employees.id = reimbursements.employee_id").
		Where("employees.company_id = ? AND status = ? AND EXTRACT(MONTH FROM reimbursements.created_at) = ? AND EXTRACT(YEAR FROM reimbursements.created_at) = ?",
			companyID, domain.ReimbursementApproved, month, year).
		Find(&reimbursements).Error
	if err != nil {
		return nil, "", err
	}

	// Generate CSV
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Write metadata header
	writer.Write([]string{"LAPORAN KEUANGAN EKSPOR"})
	writer.Write([]string{"Perusahaan", company.Name})
	writer.Write([]string{"Periode", fmt.Sprintf("%02d/%d", month, year)})
	writer.Write([]string{""})

	// Section 1: Pendapatan (Revenue)
	writer.Write([]string{"BAGIAN 1: PENDAPATAN (SALES REVENUE)"})
	writer.Write([]string{"Tanggal", "No Nota / Invoice", "Salesman", "Toko / Customer", "Tipe Toko", "Metode Pembayaran", "Jumlah (IDR)"})
	for _, t := range transactions {
		salesmanName := ""
		if t.Employee != nil {
			salesmanName = fmt.Sprintf("%s %s", t.Employee.FirstName, t.Employee.LastName)
		}
		storeName := ""
		if t.Store != nil {
			storeName = t.Store.Name
		}
		writer.Write([]string{
			t.TransactionDate.Format("2006-01-02"),
			t.ReceiptNo,
			salesmanName,
			storeName,
			t.StoreCategory,
			t.PaymentMethod,
			fmt.Sprintf("%.2f", t.TotalAmount),
		})
	}
	for _, so := range salesOrders {
		salesmanName := ""
		if so.Employee != nil {
			salesmanName = fmt.Sprintf("%s %s", so.Employee.FirstName, so.Employee.LastName)
		}
		storeName := ""
		if so.Store != nil {
			storeName = so.Store.Name
		}
		writer.Write([]string{
			so.OrderDate.Format("2006-01-02"),
			so.SONumber,
			salesmanName,
			storeName,
			so.StoreCategory,
			so.PaymentMethod,
			fmt.Sprintf("%.2f", so.TotalAmount),
		})
	}
	writer.Write([]string{""})

	// Section 2: Beban Gaji (Payroll)
	writer.Write([]string{"BAGIAN 2: BEBAN GAJI KARYAWAN (PAYROLL)"})
	writer.Write([]string{"Karyawan", "Status Payroll", "Tanggal Dibuat", "Total Dibayarkan (IDR)"})
	for _, p := range payslips {
		empName := ""
		if p.Employee != nil {
			empName = fmt.Sprintf("%s %s", p.Employee.FirstName, p.Employee.LastName)
		}
		status := "DRAFT"
		if p.PayrollRun != nil {
			status = p.PayrollRun.Status
		}
		writer.Write([]string{
			empName,
			status,
			p.CreatedAt.Format("2006-01-02"),
			fmt.Sprintf("%.2f", p.TakeHomePay),
		})
	}
	writer.Write([]string{""})

	// Section 3: Beban Reimbursement (Expenses)
	writer.Write([]string{"BAGIAN 3: BEBAN OPERASIONAL REIMBURSEMENT"})
	writer.Write([]string{"Tanggal Diajukan", "Karyawan", "Deskripsi", "Catatan/Tujuan", "Jumlah (IDR)"})
	for _, r := range reimbursements {
		empName := ""
		if r.Employee != nil {
			empName = fmt.Sprintf("%s %s", r.Employee.FirstName, r.Employee.LastName)
		}
		writer.Write([]string{
			r.CreatedAt.Format("2006-01-02"),
			empName,
			r.Title,
			r.Description,
			fmt.Sprintf("%.2f", r.Amount),
		})
	}

	writer.Flush()

	writer.Flush()

	filename := fmt.Sprintf("keuangan_%s_%d_%02d.csv", company.Code, year, month)
	return buf.Bytes(), filename, nil
}

func (u *financeUsecase) GetDeliverySummary(doNo string) (map[string]interface{}, error) {
	var batch domain.DeliveryBatch
	// Try parsing as UUID first
	if id, err := uuid.Parse(doNo); err == nil {
		if err := u.db.Preload("Driver").First(&batch, id).Error; err != nil {
			return nil, errors.New("surat jalan tidak ditemukan")
		}
	} else {
		if err := u.db.Preload("Driver").Where("delivery_order_no = ?", doNo).First(&batch).Error; err != nil {
			return nil, errors.New("surat jalan tidak ditemukan")
		}
	}

	// Calculate totals by scanning sales_payments
	var payments []domain.SalesPayment
	// We join with delivery_items via sales_transaction_id or sales_order_id to find payments for this batch
	err := u.db.Table("sales_payments").
		Select("sales_payments.*").
		Joins("LEFT JOIN sales_transactions st ON st.id = sales_payments.sales_transaction_id").
		Joins("LEFT JOIN delivery_items di1 ON di1.sales_transaction_id = st.id").
		Joins("LEFT JOIN sales_orders so ON so.id = sales_payments.sales_order_id").
		Joins("LEFT JOIN delivery_items di2 ON di2.sales_order_id = so.id").
		Where("(di1.delivery_batch_id = ? OR di2.delivery_batch_id = ?)", batch.ID, batch.ID).
		Find(&payments).Error

	if err != nil {
		return nil, err
	}

	totalCash := 0.0
	totalTransfer := 0.0
	var pendingPaymentIDs []uuid.UUID
	var allPaymentIDs []uuid.UUID

	for _, p := range payments {
		allPaymentIDs = append(allPaymentIDs, p.ID)
		if p.PaymentStatus == "PENDING" {
			pendingPaymentIDs = append(pendingPaymentIDs, p.ID)
		}

		// Only sum successful or pending cash/transfer collected by driver
		if p.PaymentStatus != "FAILED" && p.PaymentStatus != "REJECTED" {
			if p.PaymentMethod == "CASH" || p.PaymentMethod == "TUNAI" {
				totalCash += p.Amount
			} else if p.PaymentMethod == "TRANSFER" || p.PaymentMethod == "MIDTRANS" {
				totalTransfer += p.Amount
			}
		}
	}

	// FALLBACK: Jika Admin Nota belum membuat tagihan/faktur (len(payments) == 0),
	// maka Kasir bisa langsung melihat ekspektasi nominal dari Surat Jalan (SalesOrder / SalesTransaction)
	fmt.Printf("[DEBUG] len(payments) = %d\n", len(payments))
	if len(payments) == 0 {
		var deliveryItems []domain.DeliveryItem
		if err := u.db.Preload("SalesOrder").Preload("SalesTransaction").Where("delivery_batch_id = ?", batch.ID).Find(&deliveryItems).Error; err == nil {
			fmt.Printf("[DEBUG] deliveryItems count = %d\n", len(deliveryItems))
			processedSOs := make(map[uuid.UUID]bool)
			processedTrxs := make(map[uuid.UUID]bool)
			for _, di := range deliveryItems {
				fmt.Printf("[DEBUG] di.ID = %s, SO_ID = %v, Trx_ID = %v\n", di.ID, di.SalesOrderID, di.SalesTransactionID)
				if di.SalesOrder != nil {
					fmt.Printf("[DEBUG] di.SalesOrder.PaymentMethod = '%s', TotalAmount = %f\n", di.SalesOrder.PaymentMethod, di.SalesOrder.TotalAmount)
				}
				if di.SalesTransaction != nil {
					fmt.Printf("[DEBUG] di.SalesTransaction.PaymentMethod = '%s', TotalAmount = %f\n", di.SalesTransaction.PaymentMethod, di.SalesTransaction.TotalAmount)
				}

				if di.SalesTransaction != nil && di.SalesTransactionID != nil {
					trxID := *di.SalesTransactionID
					if !processedTrxs[trxID] {
						processedTrxs[trxID] = true
						pm := di.SalesTransaction.PaymentMethod

						// Jika metode adalah SALES_ORDER (Taking Order belum bayar), asumsikan ekspektasi CASH
						if pm == "SALES_ORDER" {
							pm = "CASH"
						}

						if pm == "CASH" || pm == "TUNAI" || pm == "" {
							totalCash += di.SalesTransaction.TotalAmount
						} else if pm == "TRANSFER" || pm == "MIDTRANS" || pm == "bank_transfer" {
							totalTransfer += di.SalesTransaction.TotalAmount
						}
					}
				} else if di.SalesOrder != nil && di.SalesOrderID != nil {
					soID := *di.SalesOrderID
					if !processedSOs[soID] {
						processedSOs[soID] = true
						pm := di.SalesOrder.PaymentMethod

						// Jika metode adalah SALES_ORDER, asumsikan ekspektasi CASH
						if pm == "SALES_ORDER" {
							pm = "CASH"
						}

						if pm == "CASH" || pm == "TUNAI" || pm == "" {
							totalCash += di.SalesOrder.TotalAmount
						} else if pm == "TRANSFER" || pm == "MIDTRANS" || pm == "bank_transfer" {
							totalTransfer += di.SalesOrder.TotalAmount
						}
					}
				}
			}
		}
	}

	driverName := "-"
	if batch.Driver != nil {
		driverName = batch.Driver.FirstName + " " + batch.Driver.LastName
	}

	return map[string]interface{}{
		"batch_id":            batch.ID,
		"delivery_order_no":   batch.DeliveryOrderNo,
		"driver_name":         driverName,
		"delivery_date":       batch.DeliveryDate,
		"status":              batch.Status,
		"total_cash":          totalCash,
		"total_transfer":      totalTransfer,
		"pending_payment_ids": pendingPaymentIDs,
		"payment_ids":         allPaymentIDs,
		"cash_settled_at":     batch.CashSettledAt,
	}, nil
}

func (u *financeUsecase) ApproveDeliveryBatch(batchID uuid.UUID, physicalCash float64) error {
	var batch domain.DeliveryBatch
	if err := u.db.First(&batch, batchID).Error; err != nil {
		return errors.New("surat jalan tidak ditemukan")
	}

	if batch.CashSettledAt != nil {
		return errors.New("surat jalan sudah ditutup bukunya")
	}

	// Calculate total cash collected for this batch first
	var allPayments []domain.SalesPayment
	err := u.db.Table("sales_payments").
		Select("sales_payments.*").
		Joins("LEFT JOIN sales_transactions st ON st.id = sales_payments.sales_transaction_id").
		Joins("LEFT JOIN delivery_items di1 ON di1.sales_transaction_id = st.id").
		Joins("LEFT JOIN sales_orders so ON so.id = sales_payments.sales_order_id").
		Joins("LEFT JOIN delivery_items di2 ON di2.sales_order_id = so.id").
		Where("(di1.delivery_batch_id = ? OR di2.delivery_batch_id = ?)", batch.ID, batch.ID).
		Where("sales_payments.payment_status IN ('PENDING', 'SUCCESS')").
		Find(&allPayments).Error

	if err != nil {
		return err
	}

	var totalCash float64
	var pendingPayments []domain.SalesPayment
	for _, p := range allPayments {
		if p.PaymentMethod == "CASH" || p.PaymentMethod == "TUNAI" {
			totalCash += p.Amount
		}
		if p.PaymentStatus == "PENDING" {
			pendingPayments = append(pendingPayments, p)
		}
	}

	// Update CashSettledAt
	now := time.Now()

	// Cross-check mechanism: if physical cash doesn't match, send to WAITING_CROSSCHECK
	if physicalCash != totalCash {
		batch.Status = domain.DeliveryBatchWaitingCrosscheck
		batch.ReportedCashCollected = &physicalCash
		batch.CashSettledAt = &now
		if err := u.db.Save(&batch).Error; err != nil {
			return err
		}
		return errors.New("terdapat selisih uang fisik, surat jalan dipindahkan ke tahap Cross-check")
	}

	// Approve all pending payments ONLY IF there's no discrepancy
	for _, p := range pendingPayments {
		if err := u.ApprovePayment(p.ID, "", nil, nil); err != nil {
			fmt.Printf("[FinanceUsecase] Failed to approve payment %v: %v\n", p.ID, err)
		}
	}

	// If match, determine if we can complete the batch
	var pendingItemsCount int64
	u.db.Model(&domain.DeliveryItem{}).Where("delivery_batch_id = ? AND status = ?", batch.ID, domain.DeliveryItemPending).Count(&pendingItemsCount)

	updates := map[string]interface{}{
		"cash_settled_at": &now,
	}

	if pendingItemsCount == 0 {
		updates["status"] = domain.DeliveryBatchCompleted
	}

	u.db.Model(&batch).Updates(updates)

	return nil
}

func (u *financeUsecase) ResolveCrosscheck(batchID uuid.UUID, notes string) error {
	var batch domain.DeliveryBatch
	if err := u.db.First(&batch, batchID).Error; err != nil {
		return errors.New("surat jalan tidak ditemukan")
	}

	if batch.Status != domain.DeliveryBatchWaitingCrosscheck {
		return errors.New("surat jalan tidak dalam status cross-check")
	}

	// Get all pending payments for this batch
	var pendingPayments []domain.SalesPayment
	err := u.db.Table("sales_payments").
		Select("sales_payments.id").
		Joins("LEFT JOIN sales_transactions st ON st.id = sales_payments.sales_transaction_id").
		Joins("LEFT JOIN delivery_items di1 ON di1.sales_transaction_id = st.id").
		Joins("LEFT JOIN sales_orders so ON so.id = sales_payments.sales_order_id").
		Joins("LEFT JOIN delivery_items di2 ON di2.sales_order_id = so.id").
		Where("(di1.delivery_batch_id = ? OR di2.delivery_batch_id = ?)", batch.ID, batch.ID).
		Where("sales_payments.payment_status = 'PENDING'").
		Find(&pendingPayments).Error

	if err != nil {
		return err
	}

	// Approve them all
	for _, p := range pendingPayments {
		if err := u.ApprovePayment(p.ID, "", nil, nil); err != nil {
			fmt.Printf("[FinanceUsecase] Failed to approve payment %v: %v\n", p.ID, err)
		}
	}

	// Finish batch
	now := time.Now()

	// Determine if we can complete the batch
	var pendingItemsCount int64
	u.db.Model(&domain.DeliveryItem{}).Where("delivery_batch_id = ? AND status = ?", batch.ID, domain.DeliveryItemPending).Count(&pendingItemsCount)

	updates := map[string]interface{}{
		"cash_settled_at":  &now,
		"crosscheck_notes": notes,
	}

	if pendingItemsCount == 0 {
		updates["status"] = domain.DeliveryBatchCompleted
	}

	u.db.Model(&batch).Updates(updates)

	return nil
}

func (u *financeUsecase) GetTodayCompletedDeliveries(branchID *uuid.UUID) ([]map[string]interface{}, error) {
	var batches []domain.DeliveryBatch

	// We want to show DOs that Kasir might want to reconcile.
	// This means DOs that are already published (SUPERVISOR_APPROVED and beyond).
	// Sorted by newest created_at first.
	query := u.db.Preload("Driver").
		Where("status NOT IN ?", []domain.DeliveryBatchStatus{
			domain.DeliveryBatchWaitingApproval,
			domain.DeliveryBatchWaitingAssignment, // If still waiting assignment, driver hasn't gone yet, but Kasir might want to see it? Actually, published is fine.
		}).
		// Menampilkan DO 7 hari terakhir agar DO kemarin yang belum disetor masih muncul
		Where("created_at >= ?", time.Now().AddDate(0, 0, -7))

	if branchID != nil {
		query = query.Where("branch_id = ?", *branchID)
	}

	if err := query.Order("created_at DESC").Find(&batches).Error; err != nil {
		return nil, err
	}

	var results []map[string]interface{}
	for _, batch := range batches {
		driverName := "-"
		if batch.Driver != nil {
			driverName = batch.Driver.FirstName + " " + batch.Driver.LastName
		}
		results = append(results, map[string]interface{}{
			"id":                batch.ID,
			"delivery_order_no": batch.DeliveryOrderNo,
			"driver_name":       driverName,
			"status":            batch.Status,
			"updated_at":        batch.UpdatedAt,
		})
	}

	return results, nil
}

func (u *financeUsecase) ReportDiscrepancy(id uuid.UUID, notes string) error {
	return u.db.Transaction(func(tx *gorm.DB) error {
		var payment domain.SalesPayment
		if err := tx.First(&payment, "id = ?", id).Error; err != nil {
			return err
		}
		if payment.PaymentStatus != "PENDING" {
			return errors.New("pembayaran tidak dalam status PENDING")
		}

		// Add notes indicating discrepancy
		payment.Notes = "[SELISIH/CROSSCHECK]: " + notes
		if err := tx.Save(&payment).Error; err != nil {
			return err
		}

		// Also update the DeliveryBatch to WAITING_CROSSCHECK
		var batchID *uuid.UUID
		if payment.SalesOrderID != nil {
			var so domain.SalesOrder
			if err := tx.First(&so, "id = ?", *payment.SalesOrderID).Error; err == nil {
				if so.DeliveryBatchID != nil {
					batchID = so.DeliveryBatchID
				} else {
					var di domain.DeliveryItem
					if err := tx.First(&di, "sales_order_id = ?", so.ID).Error; err == nil {
						bID := di.DeliveryBatchID
						batchID = &bID
					}
				}
			}
		}
		if batchID == nil && payment.SalesTransactionID != nil {
			// Find via delivery_items
			var di domain.DeliveryItem
			if err := tx.First(&di, "sales_transaction_id = ?", *payment.SalesTransactionID).Error; err == nil {
				bID := di.DeliveryBatchID
				batchID = &bID
			}
		}

		if batchID != nil {
			var batch domain.DeliveryBatch
			if err := tx.First(&batch, "id = ?", *batchID).Error; err == nil {
				now := time.Now()
				batchNotes := batch.Notes
				if batchNotes != "" {
					batchNotes += " | "
				}
				batchNotes += "[SELISIH KASIR]: " + notes
				tx.Model(&batch).Updates(map[string]interface{}{
					"status":          domain.DeliveryBatchWaitingCrosscheck,
					"cash_settled_at": &now,
					"notes":           batchNotes,
				})
			}
		}

		return nil
	})
}

func (u *financeUsecase) GetFinancialReceivablesList(companyID uuid.UUID, branchID *uuid.UUID) (*domain.FinancialReceivablesResponse, error) {
	return u.repo.GetFinancialReceivablesList(companyID, branchID)
}

func (u *financeUsecase) RecordFinancialReceivablePayment(req domain.RepayBranchFinancialReceivableRequest, recordedBy *uuid.UUID) (*domain.KasTransaction, error) {
	return u.repo.RecordFinancialReceivablePayment(req, recordedBy)
}

func (u *financeUsecase) SettleFinancialReceivableTx(txID uuid.UUID, req domain.SettleFinancialReceivableTxRequest, settledBy *uuid.UUID) (*domain.KasAdvanceSettlement, error) {
	return u.repo.SettleFinancialReceivableTx(txID, req, settledBy)
}

func (u *financeUsecase) GetDailySalesIncomeReport(date, startDate, endDate, companyIDStr, branchIDStr string) (*domain.DailySalesIncomeReport, error) {
	loc := time.FixedZone("WIB", 7*3600)
	var startOfDay, endOfDay time.Time
	var displayDate string

	if startDate != "" && endDate != "" {
		s, err := time.ParseInLocation("2006-01-02", startDate, loc)
		if err != nil {
			s = time.Now().In(loc).Truncate(24 * time.Hour)
		}
		e, err := time.ParseInLocation("2006-01-02", endDate, loc)
		if err != nil {
			e = s
		}
		startOfDay = s
		endOfDay = e.Add(24 * time.Hour)
		displayDate = fmt.Sprintf("%s s/d %s", startDate, endDate)
	} else {
		if date == "" {
			date = time.Now().In(loc).Format("2006-01-02")
		}
		t, err := time.ParseInLocation("2006-01-02", date, loc)
		if err != nil {
			t = time.Now().In(loc).Truncate(24 * time.Hour)
			date = t.Format("2006-01-02")
		}
		startOfDay = t
		endOfDay = startOfDay.Add(24 * time.Hour)
		displayDate = date
	}

	report := &domain.DailySalesIncomeReport{
		Summary: domain.DailySalesIncomeSummary{
			Date:              displayDate,
			TotalCash:         0,
			TotalTransfer:     0,
			GrandTotal:        0,
			TotalTransactions: 0,
		},
		Companies: []domain.DailySalesIncomeCompany{},
		Payments:  []domain.DailySalesIncomePaymentItem{},
	}

	type QueryRow struct {
		ID               uuid.UUID  `gorm:"column:id"`
		Amount           float64    `gorm:"column:amount"`
		PaymentMethod    string     `gorm:"column:payment_method"`
		PaymentStatus    string     `gorm:"column:payment_status"`
		PaymentDate      time.Time  `gorm:"column:payment_date"`
		ApprovedAt       *time.Time `gorm:"column:approved_at"`
		Notes            string     `gorm:"column:notes"`
		TransferProofURL *string    `gorm:"column:transfer_proof_url"`
		CompanyID        *uuid.UUID `gorm:"column:company_id"`
		CompanyName      string     `gorm:"column:company_name"`
		CompanyCode      string     `gorm:"column:company_code"`
		BranchID         *uuid.UUID `gorm:"column:branch_id"`
		BranchName       string     `gorm:"column:branch_name"`
		StoreID          *uuid.UUID `gorm:"column:store_id"`
		StoreName        string     `gorm:"column:store_name"`
		SalesmanID       *uuid.UUID `gorm:"column:salesman_id"`
		SalesmanName     string     `gorm:"column:salesman_name"`
		ApproverID       *uuid.UUID `gorm:"column:approver_id"`
		ApproverName     string     `gorm:"column:approver_name"`
		ReferenceNo      string     `gorm:"column:reference_no"`
	}

	query := `
		SELECT
			sp.id,
			sp.amount,
			sp.payment_method,
			sp.payment_status,
			sp.payment_date,
			sp.approved_at,
			COALESCE(sp.notes, '') AS notes,
			sp.transfer_proof_url,
			COALESCE(so.company_id, st.company_id, e.company_id, b.company_id) AS company_id,
			COALESCE(c_so.name, c_st.name, c_e.name, c_b.name, 'Tanpa PT') AS company_name,
			COALESCE(c_so.code, c_st.code, c_e.code, c_b.code, '') AS company_code,
			COALESCE(e.operating_branch_id, e.branch_id, so.branch_id, st.branch_id) AS branch_id,
			COALESCE(b_op.name, b_e.name, b_so.name, b_st.name, 'Tanpa Cabang') AS branch_name,
			COALESCE(s_so.id, s_st.id) AS store_id,
			COALESCE(s_so.name, s_st.name, '-') AS store_name,
			COALESCE(e_so.id, e_st.id, e.id) AS salesman_id,
			COALESCE(
				NULLIF(TRIM(CONCAT(e_so.first_name, ' ', COALESCE(e_so.last_name, ''))), ''),
				NULLIF(TRIM(CONCAT(e_st.first_name, ' ', COALESCE(e_st.last_name, ''))), ''),
				NULLIF(TRIM(CONCAT(e.first_name, ' ', COALESCE(e.last_name, ''))), ''),
				'-'
			) AS salesman_name,
			sp.verified_by_id AS approver_id,
			COALESCE(NULLIF(TRIM(CONCAT(v.first_name, ' ', COALESCE(v.last_name, ''))), ''), '-') AS approver_name,
			COALESCE(so.so_number, st.receipt_no, '-') AS reference_no
		FROM sales_payments sp
		LEFT JOIN sales_orders so ON so.id = sp.sales_order_id
		LEFT JOIN sales_transactions st ON st.id = sp.sales_transaction_id
		LEFT JOIN stores s_so ON s_so.id = so.store_id
		LEFT JOIN stores s_st ON s_st.id = st.store_id
		LEFT JOIN employees e ON e.id = sp.employee_id
		LEFT JOIN employees e_so ON e_so.id = so.employee_id
		LEFT JOIN employees e_st ON e_st.id = st.employee_id
		LEFT JOIN employees v ON v.id = sp.verified_by_id
		LEFT JOIN branches b ON b.id = COALESCE(e.operating_branch_id, e.branch_id, so.branch_id, st.branch_id)
		LEFT JOIN branches b_op ON b_op.id = e.operating_branch_id
		LEFT JOIN branches b_e ON b_e.id = e.branch_id
		LEFT JOIN branches b_so ON b_so.id = so.branch_id
		LEFT JOIN branches b_st ON b_st.id = st.branch_id
		LEFT JOIN companies c_so ON c_so.id = so.company_id
		LEFT JOIN companies c_st ON c_st.id = st.company_id
		LEFT JOIN companies c_e ON c_e.id = e.company_id
		LEFT JOIN companies c_b ON c_b.id = b.company_id
		WHERE sp.payment_status IN ('PAID', 'APPROVED', 'CASHIER_APPROVED', 'SUCCESS')
		  AND COALESCE(sp.approved_at, sp.payment_date, sp.updated_at) >= ?
		  AND COALESCE(sp.approved_at, sp.payment_date, sp.updated_at) < ?
	`
	args := []interface{}{startOfDay, endOfDay}

	if companyIDStr != "" && companyIDStr != "ALL" {
		query += ` AND (so.company_id = ? OR st.company_id = ? OR e.company_id = ? OR b.company_id = ?)`
		args = append(args, companyIDStr, companyIDStr, companyIDStr, companyIDStr)
	}

	if branchIDStr != "" && branchIDStr != "ALL_BRANCHES" {
		query += ` AND (COALESCE(e.operating_branch_id, e.branch_id, so.branch_id, st.branch_id) = ?)`
		args = append(args, branchIDStr)
	}

	query += ` ORDER BY COALESCE(sp.approved_at, sp.payment_date) DESC, sp.created_at DESC`

	var rows []QueryRow
	if err := u.db.Raw(query, args...).Scan(&rows).Error; err != nil {
		return nil, err
	}

	// Pre-load all registered companies to ensure every PT is present
	companiesList, _ := u.repo.GetAllCompanies()
	compMap := make(map[string]*domain.DailySalesIncomeCompany)
	compOrder := make([]string, 0)

	for _, comp := range companiesList {
		cID := comp.ID.String()
		compMap[cID] = &domain.DailySalesIncomeCompany{
			CompanyID:        cID,
			CompanyName:      comp.Name,
			CompanyCode:      comp.Code,
			TotalCash:        0,
			TotalTransfer:    0,
			TotalAmount:      0,
			TransactionCount: 0,
			Branches:         []domain.DailySalesIncomeBranch{},
		}
		compOrder = append(compOrder, cID)
	}

	// Helper for branch map per company: companyID -> map[branchID]*domain.DailySalesIncomeBranch
	branchMap := make(map[string]map[string]*domain.DailySalesIncomeBranch)
	for cID := range compMap {
		branchMap[cID] = make(map[string]*domain.DailySalesIncomeBranch)
	}

	for _, row := range rows {
		isCash := strings.EqualFold(row.PaymentMethod, "CASH") || strings.EqualFold(row.PaymentMethod, "TUNAI")
		if isCash {
			report.Summary.TotalCash += row.Amount
		} else {
			report.Summary.TotalTransfer += row.Amount
		}
		report.Summary.GrandTotal += row.Amount
		report.Summary.TotalTransactions++

		// Resolve company ID
		cID := ""
		cName := row.CompanyName
		cCode := row.CompanyCode
		if row.CompanyID != nil && *row.CompanyID != uuid.Nil {
			cID = row.CompanyID.String()
		} else {
			cID = "OTHER"
		}

		if _, exists := compMap[cID]; !exists {
			compMap[cID] = &domain.DailySalesIncomeCompany{
				CompanyID:        cID,
				CompanyName:      cName,
				CompanyCode:      cCode,
				TotalCash:        0,
				TotalTransfer:    0,
				TotalAmount:      0,
				TransactionCount: 0,
				Branches:         []domain.DailySalesIncomeBranch{},
			}
			compOrder = append(compOrder, cID)
			branchMap[cID] = make(map[string]*domain.DailySalesIncomeBranch)
		}

		comp := compMap[cID]
		comp.TransactionCount++
		comp.TotalAmount += row.Amount
		if isCash {
			comp.TotalCash += row.Amount
		} else {
			comp.TotalTransfer += row.Amount
		}

		// Resolve branch
		bID := ""
		bName := row.BranchName
		if row.BranchID != nil && *row.BranchID != uuid.Nil {
			bID = row.BranchID.String()
		} else {
			bID = "OTHER_BRANCH"
		}

		bMap := branchMap[cID]
		if _, exists := bMap[bID]; !exists {
			bMap[bID] = &domain.DailySalesIncomeBranch{
				BranchID:         bID,
				BranchName:       bName,
				CashAmount:       0,
				TransferAmount:   0,
				TotalAmount:      0,
				TransactionCount: 0,
			}
		}
		bItem := bMap[bID]
		bItem.TransactionCount++
		bItem.TotalAmount += row.Amount
		if isCash {
			bItem.CashAmount += row.Amount
		} else {
			bItem.TransferAmount += row.Amount
		}

		// Itemized payment detail
		pItem := domain.DailySalesIncomePaymentItem{
			ID:               row.ID,
			PaymentDate:      row.PaymentDate,
			ApprovedAt:       row.ApprovedAt,
			ReferenceNo:      row.ReferenceNo,
			CompanyID:        cID,
			CompanyName:      cName,
			BranchID:         bID,
			BranchName:       bName,
			PaymentMethod:    row.PaymentMethod,
			Amount:           row.Amount,
			PaymentStatus:    row.PaymentStatus,
			Notes:            row.Notes,
			TransferProofURL: row.TransferProofURL,
		}
		if row.StoreID != nil {
			pItem.StoreID = row.StoreID.String()
		}
		pItem.StoreName = row.StoreName
		if row.SalesmanID != nil {
			pItem.SalesmanID = row.SalesmanID.String()
		}
		pItem.SalesmanName = row.SalesmanName
		if row.ApproverID != nil {
			pItem.ApproverID = row.ApproverID.String()
		}
		pItem.ApproverName = row.ApproverName

		report.Payments = append(report.Payments, pItem)
	}

	// Assemble companies and their branches
	for _, cID := range compOrder {
		comp := compMap[cID]
		bSlice := make([]domain.DailySalesIncomeBranch, 0)
		for _, bItem := range branchMap[cID] {
			bSlice = append(bSlice, *bItem)
		}
		sort.Slice(bSlice, func(i, j int) bool {
			return bSlice[i].TotalAmount > bSlice[j].TotalAmount
		})
		comp.Branches = bSlice

		if companyIDStr != "" && companyIDStr != "ALL" {
			if comp.CompanyID == companyIDStr {
				report.Companies = append(report.Companies, *comp)
			}
		} else {
			report.Companies = append(report.Companies, *comp)
		}
	}

	sort.Slice(report.Companies, func(i, j int) bool {
		return report.Companies[i].TotalAmount > report.Companies[j].TotalAmount
	})

	return report, nil
}
