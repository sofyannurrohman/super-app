package usecase

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"example.com/hris/backend/internal/repository"
	"gorm.io/gorm"
)

// =============================================================================
// Interface
// =============================================================================

type GLPostingUsecase interface {
	// Auto-posting dari transaksi bisnis yang ada
	// Semua dipanggil sebagai goroutine (non-blocking) dari usecase lain
	PostSalesTransaction(companyID uuid.UUID, refID uuid.UUID, amount float64, cogs float64, date time.Time) error
	PostSalesReturn(companyID uuid.UUID, refID uuid.UUID, returnAmount float64, date time.Time) error
	PostPayrollRun(companyID uuid.UUID, payrollRunID uuid.UUID, totalPayout float64, pph21Total float64, period time.Time) error
	PostReimbursement(companyID uuid.UUID, reimbID uuid.UUID, amount float64, date time.Time) error
	PostBranchDeposit(companyID uuid.UUID, refID uuid.UUID, amount float64, isCash bool, date time.Time) (*uuid.UUID, error)
	PostCashierReport(companyID uuid.UUID, refID uuid.UUID, amount float64, date time.Time) error

	// Manual Journal Entry (Admin)
	CreateManualJournalEntry(companyID uuid.UUID, createdBy uuid.UUID, req domain.CreateJournalEntryRequest) (*domain.JournalEntry, error)

	// Reporting
	GetJournalEntries(companyID uuid.UUID, month, year, page, limit int) ([]domain.JournalEntry, int64, error)
	GetJournalEntryByID(id uuid.UUID) (*domain.JournalEntry, error)
	GetBalanceSheet(companyID uuid.UUID, year, month int) (*domain.BalanceSheetReport, error)
	GetProfitLossReport(companyID uuid.UUID, month, year int) (*domain.ProfitLossReport, error)
	GetCashFlowReport(companyID uuid.UUID, month, year int) (*domain.CashFlowReport, error)
	GetTrialBalance(companyID uuid.UUID, month, year int) ([]domain.TrialBalanceLine, error)

	// Journal Status Management
	PostJournalEntry(id uuid.UUID) error
	CancelJournalEntry(id uuid.UUID) error

	// Chart of Accounts
	GetChartOfAccounts(companyID uuid.UUID) ([]domain.Account, error)
	CreateAccount(companyID uuid.UUID, req CreateAccountRequest) (*domain.Account, error)
	UpdateAccount(id uuid.UUID, req UpdateAccountRequest) (*domain.Account, error)
	SeedDefaultAccounts(companyID uuid.UUID) error

	// Period Management
	ClosePeriod(companyID uuid.UUID, year, month int, closedBy uuid.UUID) error
}

type CreateAccountRequest struct {
	Code           string     `json:"code"`
	Name           string     `json:"name"`
	AccountType    string     `json:"account_type"`
	AccountSubtype string     `json:"account_subtype"`
	NormalBalance  string     `json:"normal_balance"`
	ParentID       *uuid.UUID `json:"parent_id"`
}

type UpdateAccountRequest struct {
	Name           string `json:"name"`
	AccountSubtype string `json:"account_subtype"`
	IsActive       bool   `json:"is_active"`
}

// =============================================================================
// Implementation
// =============================================================================

type glPostingUsecase struct {
	repo        repository.AccountingRepository
	db          *gorm.DB
	companyRepo repository.CompanyRepository
}

func NewGLPostingUsecase(repo repository.AccountingRepository, companyRepo repository.CompanyRepository, db *gorm.DB) GLPostingUsecase {
	return &glPostingUsecase{
		repo:        repo,
		db:          db,
		companyRepo: companyRepo,
	}
}

// =============================================================================
// Auto-posting dari transaksi bisnis
// =============================================================================

// PostSalesTransaction — dipanggil saat pembayaran penjualan di-approve.
// Jurnal PSAK:
//
//	Dr. Kas & Bank (1100)             amount
//	Cr. Pendapatan Penjualan (4100)               amount
//	Dr. Harga Pokok Penjualan (5100)  cogs    (jika cogs > 0)
//	Cr. Persediaan Barang (1300)                  cogs
func (u *glPostingUsecase) PostSalesTransaction(companyID uuid.UUID, refID uuid.UUID, amount float64, cogs float64, date time.Time) error {
	// Cegah posting duplikat
	if existing, _ := u.repo.GetJournalEntryByRef("SALES_TRANSACTION", refID); existing != nil {
		return nil
	}

	kasAccount, err := u.repo.GetAccountByCode(companyID, "1100")
	if err != nil {
		return fmt.Errorf("akun Kas & Bank (1100) tidak ditemukan: %w", err)
	}
	pendapatanAccount, err := u.repo.GetAccountByCode(companyID, "4100")
	if err != nil {
		return fmt.Errorf("akun Pendapatan (4100) tidak ditemukan: %w", err)
	}

	lines := []domain.JournalEntryLine{
		{AccountID: kasAccount.ID, DebitAmount: amount, Description: "Penerimaan kas dari penjualan"},
		{AccountID: pendapatanAccount.ID, CreditAmount: amount, Description: "Pendapatan penjualan"},
	}

	if cogs > 0 {
		hppAccount, err := u.repo.GetAccountByCode(companyID, "5100")
		if err == nil {
			persediaanAccount, err2 := u.repo.GetAccountByCode(companyID, "1300")
			if err2 == nil {
				lines = append(lines,
					domain.JournalEntryLine{AccountID: hppAccount.ID, DebitAmount: cogs, Description: "Harga Pokok Penjualan"},
					domain.JournalEntryLine{AccountID: persediaanAccount.ID, CreditAmount: cogs, Description: "Pengurangan persediaan"},
				)
			}
		}
	}

	return u.createAndPostEntry(companyID, "SALES_TRANSACTION", &refID, date,
		fmt.Sprintf("Penjualan Rp %.0f", amount), lines)
}

// PostSalesReturn — dipanggil saat retur penjualan diproses.
// Jurnal PSAK:
//
//	Dr. Retur & Potongan Penjualan (4200)  amount
//	Cr. Kas & Bank / Piutang (1100)                   amount
func (u *glPostingUsecase) PostSalesReturn(companyID uuid.UUID, refID uuid.UUID, returnAmount float64, date time.Time) error {
	if existing, _ := u.repo.GetJournalEntryByRef("SALES_RETURN", refID); existing != nil {
		return nil
	}

	returAccount, err := u.repo.GetAccountByCode(companyID, "4200")
	if err != nil {
		return fmt.Errorf("akun Retur Penjualan (4200) tidak ditemukan: %w", err)
	}
	kasAccount, err := u.repo.GetAccountByCode(companyID, "1100")
	if err != nil {
		return fmt.Errorf("akun Kas & Bank (1100) tidak ditemukan: %w", err)
	}

	lines := []domain.JournalEntryLine{
		{AccountID: returAccount.ID, DebitAmount: returnAmount, Description: "Retur penjualan"},
		{AccountID: kasAccount.ID, CreditAmount: returnAmount, Description: "Pengembalian kas ke pelanggan"},
	}

	return u.createAndPostEntry(companyID, "SALES_RETURN", &refID, date,
		fmt.Sprintf("Retur penjualan Rp %.0f", returnAmount), lines)
}

// PostPayrollRun — dipanggil saat payroll run difinalisasi (status → PAID).
// Jurnal PSAK:
//
//	Dr. Beban Gaji & Tunjangan (5200)  gaji bersih
//	Dr. Hutang PPh 21 (2310)           pph21 (jika ada)
//	Cr. Kas & Bank (1100)                         total payout
func (u *glPostingUsecase) PostPayrollRun(companyID uuid.UUID, payrollRunID uuid.UUID, totalPayout float64, pph21Total float64, period time.Time) error {
	if existing, _ := u.repo.GetJournalEntryByRef("PAYROLL_RUN", payrollRunID); existing != nil {
		return nil
	}

	gajiAccount, err := u.repo.GetAccountByCode(companyID, "5200")
	if err != nil {
		return fmt.Errorf("akun Beban Gaji (5200) tidak ditemukan: %w", err)
	}
	kasAccount, err := u.repo.GetAccountByCode(companyID, "1100")
	if err != nil {
		return fmt.Errorf("akun Kas & Bank (1100) tidak ditemukan: %w", err)
	}

	// Beban gaji = totalPayout (yang diterima karyawan, sebelum PPh dipotong oleh perusahaan)
	gajiBeban := totalPayout
	lines := []domain.JournalEntryLine{
		{AccountID: gajiAccount.ID, DebitAmount: gajiBeban, Description: "Beban gaji karyawan"},
		{AccountID: kasAccount.ID, CreditAmount: totalPayout, Description: "Pembayaran gaji via bank"},
	}

	// Hutang PPh 21 yang dipotong dari karyawan
	if pph21Total > 0 {
		pph21Account, err := u.repo.GetAccountByCode(companyID, "2310")
		if err == nil {
			lines = append(lines, domain.JournalEntryLine{
				AccountID:    pph21Account.ID,
				CreditAmount: pph21Total,
				Description:  "Hutang PPh 21 karyawan",
			})
			// Koreksi: Kas yang keluar = totalPayout - pph21 (PPh 21 belum disetor ke DJP)
			// Namun untuk simplisitas, totalPayout = gaji bersih yang sudah dipotong PPh
			// Beban gaji gross = totalPayout + pph21 (jika PPh ditanggung perusahaan)
		}
	}

	desc := fmt.Sprintf("Payroll %s %d — total Rp %.0f", monthName(int(period.Month())), period.Year(), totalPayout)
	return u.createAndPostEntry(companyID, "PAYROLL_RUN", &payrollRunID, period, desc, lines)
}

// PostReimbursement — dipanggil saat reimbursement di-approve.
// Jurnal PSAK:
//
//	Dr. Beban Reimbursement (5400)  amount
//	Cr. Kas & Bank (1100)                      amount
func (u *glPostingUsecase) PostReimbursement(companyID uuid.UUID, reimbID uuid.UUID, amount float64, date time.Time) error {
	if existing, _ := u.repo.GetJournalEntryByRef("REIMBURSEMENT", reimbID); existing != nil {
		return nil
	}

	bebanAccount, err := u.repo.GetAccountByCode(companyID, "5400")
	if err != nil {
		return fmt.Errorf("akun Beban Reimbursement (5400) tidak ditemukan: %w", err)
	}
	kasAccount, err := u.repo.GetAccountByCode(companyID, "1100")
	if err != nil {
		return fmt.Errorf("akun Kas & Bank (1100) tidak ditemukan: %w", err)
	}

	lines := []domain.JournalEntryLine{
		{AccountID: bebanAccount.ID, DebitAmount: amount, Description: "Beban reimbursement karyawan"},
		{AccountID: kasAccount.ID, CreditAmount: amount, Description: "Pembayaran reimbursement"},
	}

	return u.createAndPostEntry(companyID, "REIMBURSEMENT", &reimbID, date,
		fmt.Sprintf("Reimbursement Rp %.0f", amount), lines)
}

func (u *glPostingUsecase) PostBranchDeposit(companyID uuid.UUID, refID uuid.UUID, amount float64, isCash bool, date time.Time) (*uuid.UUID, error) {
	// Ambil akun untuk Debit (Kas atau Bank)
	var debitAccCode string
	if isCash {
		debitAccCode = "1110" // Kas di Tangan (sesuaikan dengan CoA)
	} else {
		debitAccCode = "1120" // Bank Pusat (sesuaikan dengan CoA)
	}
	debitAcc, err := u.repo.GetAccountByCode(companyID, debitAccCode)
	if err != nil {
		return nil, fmt.Errorf("akun debit %s tidak ditemukan: %w", debitAccCode, err)
	}

	// Ambil akun untuk Kredit (Piutang Cabang atau Kas Transit Cabang)
	creditAcc, err := u.repo.GetAccountByCode(companyID, "1140") // Contoh: Piutang Karyawan/Cabang, sesuaikan
	if err != nil {
		return nil, fmt.Errorf("akun piutang cabang (1140) tidak ditemukan: %w", err)
	}

	lines := []domain.JournalEntryLine{
		{
			AccountID:   debitAcc.ID,
			DebitAmount: amount,
			Description: "Penerimaan Setoran Cabang",
		},
		{
			AccountID:    creditAcc.ID,
			CreditAmount: amount,
			Description:  "Penerimaan Setoran Cabang",
		},
	}

	err = u.createAndPostEntry(companyID, "BRANCH_DEPOSIT", &refID, date,
		fmt.Sprintf("Setoran Cabang Rp %.0f", amount), lines)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// PostCashierReport - dipanggil saat kasir cabang menyetorkan dana tunai ke bank (Fitur Pelaporan Kasir)
func (u *glPostingUsecase) PostCashierReport(companyID uuid.UUID, refID uuid.UUID, amount float64, date time.Time) error {
	if existing, _ := u.repo.GetJournalEntryByRef("CASHIER_REPORT", refID); existing != nil {
		return nil
	}

	// Debit: Bank Pusat
	debitAcc, err := u.repo.GetAccountByCode(companyID, "1120") // Bank Pusat
	if err != nil {
		return fmt.Errorf("akun Bank Pusat (1120) tidak ditemukan: %w", err)
	}

	// Kredit: Kas Cabang (Kas di Tangan)
	creditAcc, err := u.repo.GetAccountByCode(companyID, "1110") // Kas Cabang / Tangan
	if err != nil {
		return fmt.Errorf("akun Kas Cabang (1110) tidak ditemukan: %w", err)
	}

	lines := []domain.JournalEntryLine{
		{AccountID: debitAcc.ID, DebitAmount: amount, Description: "Setoran tunai kasir ke bank"},
		{AccountID: creditAcc.ID, CreditAmount: amount, Description: "Pengeluaran fisik kas cabang"},
	}

	return u.createAndPostEntry(companyID, "CASHIER_REPORT", &refID, date,
		fmt.Sprintf("Setoran Kasir ke Bank Rp %.0f", amount), lines)
}

// =============================================================================
// Manual Journal Entry
// =============================================================================

func (u *glPostingUsecase) CreateManualJournalEntry(companyID uuid.UUID, createdBy uuid.UUID, req domain.CreateJournalEntryRequest) (*domain.JournalEntry, error) {
	if len(req.Lines) < 2 {
		return nil, errors.New("jurnal harus memiliki minimal 2 baris (debit dan kredit)")
	}

	// Validasi: total debit harus = total kredit (double-entry balance)
	var totalDebit, totalCredit float64
	for _, l := range req.Lines {
		totalDebit += l.DebitAmount
		totalCredit += l.CreditAmount
	}
	if fmt.Sprintf("%.2f", totalDebit) != fmt.Sprintf("%.2f", totalCredit) {
		return nil, fmt.Errorf("jurnal tidak balance: total debit Rp %.2f ≠ total kredit Rp %.2f", totalDebit, totalCredit)
	}

	var lines []domain.JournalEntryLine
	for _, l := range req.Lines {
		if l.AccountID == uuid.Nil {
			return nil, errors.New("account_id tidak boleh kosong")
		}
		lines = append(lines, domain.JournalEntryLine{
			AccountID:    l.AccountID,
			DebitAmount:  l.DebitAmount,
			CreditAmount: l.CreditAmount,
			Description:  l.Description,
		})
	}

	entryNumber, err := u.repo.GenerateEntryNumber(companyID, req.EntryDate.Year(), int(req.EntryDate.Month()))
	if err != nil {
		return nil, err
	}

	entry := &domain.JournalEntry{
		CompanyID:     companyID,
		EntryNumber:   entryNumber,
		EntryDate:     req.EntryDate,
		Description:   req.Description,
		ReferenceType: "MANUAL",
		Status:        "POSTED",
		CreatedBy:     &createdBy,
		Lines:         lines,
	}

	if err := u.repo.CreateJournalEntry(entry); err != nil {
		return nil, err
	}

	// Update account balances
	year, month := req.EntryDate.Year(), int(req.EntryDate.Month())
	for _, l := range lines {
		_ = u.repo.UpsertAccountBalance(l.AccountID, year, month, l.DebitAmount, l.CreditAmount)
	}

	return entry, nil
}

// =============================================================================
// Reporting
// =============================================================================

func (u *glPostingUsecase) GetJournalEntries(companyID uuid.UUID, month, year, page, limit int) ([]domain.JournalEntry, int64, error) {
	return u.repo.GetJournalEntries(companyID, month, year, page, limit)
}

func (u *glPostingUsecase) GetJournalEntryByID(id uuid.UUID) (*domain.JournalEntry, error) {
	return u.repo.GetJournalEntryByID(id)
}

// GetBalanceSheet menghitung Neraca kumulatif s.d. akhir bulan/tahun yang dipilih.
// Neraca = saldo akumulasi dari awal sampai periode ini (bukan hanya periode ini).
func (u *glPostingUsecase) GetBalanceSheet(companyID uuid.UUID, year, month int) (*domain.BalanceSheetReport, error) {
	company, err := u.companyRepo.FindByID(companyID)
	if err != nil {
		return nil, err
	}

	// Ambil semua saldo akun s.d. periode ini
	var allBalances []domain.AccountBalance
	err = u.db.Table("account_balances ab").
		Joins("JOIN accounts a ON a.id = ab.account_id").
		Where("a.company_id = ? AND (ab.period_year < ? OR (ab.period_year = ? AND ab.period_month <= ?))",
			companyID, year, year, month).
		Select("ab.account_id, SUM(ab.balance) as balance, ab.period_year, ab.period_month").
		Group("ab.account_id, ab.period_year, ab.period_month").
		Preload("Account").
		Find(&allBalances).Error
	if err != nil {
		return nil, err
	}

	// Aggregate saldo per akun
	accountNetBalance := make(map[uuid.UUID]float64)
	accountInfo := make(map[uuid.UUID]domain.Account)
	for _, b := range allBalances {
		if b.Account != nil {
			accountNetBalance[b.AccountID] += b.Balance
			accountInfo[b.AccountID] = *b.Account
		}
	}

	report := &domain.BalanceSheetReport{
		CompanyID:   companyID,
		CompanyName: company.Name,
		AsOfDate:    fmt.Sprintf("31 %s %d", monthName(month), year),
	}

	// Klasifikasikan ke tipe akun
	for accID, bal := range accountNetBalance {
		acc := accountInfo[accID]
		if bal == 0 {
			continue
		}
		summary := domain.AccountSummary{
			Code:    acc.Code,
			Name:    acc.Name,
			Balance: bal,
		}
		switch acc.AccountType {
		case "ASSET":
			report.Assets.Accounts = append(report.Assets.Accounts, summary)
			report.Assets.Total += bal
		case "LIABILITY":
			report.Liabilities.Accounts = append(report.Liabilities.Accounts, summary)
			report.Liabilities.Total += bal
		case "EQUITY":
			report.Equity.Accounts = append(report.Equity.Accounts, summary)
			report.Equity.Total += bal
		}
	}
	report.TotalLE = report.Liabilities.Total + report.Equity.Total

	return report, nil
}

// GetCashFlowReport menghitung Laporan Arus Kas untuk periode bulan/tahun.
// Menggunakan metode langsung: melihat mutasi akun Kas & Bank (1100)
// dan mengklasifikasikan berdasarkan reference_type jurnal.
func (u *glPostingUsecase) GetCashFlowReport(companyID uuid.UUID, month, year int) (*domain.CashFlowReport, error) {
	company, err := u.companyRepo.FindByID(companyID)
	if err != nil {
		return nil, err
	}

	// Ambil semua baris jurnal yang menyentuh akun Kas periode ini
	kasAccount, err := u.repo.GetAccountByCode(companyID, "1100")
	if err != nil {
		return nil, fmt.Errorf("akun Kas (1100) tidak ditemukan")
	}

	type cashLine struct {
		ReferenceType string
		DebitAmount   float64
		CreditAmount  float64
		Description   string
	}
	var cashLines []cashLine

	err = u.db.Table("journal_entry_lines jl").
		Joins("JOIN journal_entries je ON je.id = jl.journal_entry_id").
		Where("je.company_id = ? AND jl.account_id = ? AND EXTRACT(MONTH FROM je.entry_date) = ? AND EXTRACT(YEAR FROM je.entry_date) = ? AND je.status = 'POSTED'",
			companyID, kasAccount.ID, month, year).
		Select("je.reference_type, jl.debit_amount, jl.credit_amount, je.description").
		Scan(&cashLines).Error
	if err != nil {
		return nil, err
	}

	report := &domain.CashFlowReport{
		CompanyID:   companyID,
		CompanyName: company.Name,
		PeriodMonth: month,
		PeriodYear:  year,
	}

	// Klasifikasikan arus kas
	for _, cl := range cashLines {
		netCash := cl.DebitAmount - cl.CreditAmount // positif = masuk, negatif = keluar
		item := domain.CashFlowItem{
			Description: cl.Description,
			Amount:      netCash,
		}

		switch cl.ReferenceType {
		case "SALES_TRANSACTION":
			report.Operating.Items = append(report.Operating.Items, item)
			report.Operating.Total += netCash
		case "PAYROLL_RUN", "REIMBURSEMENT":
			report.Operating.Items = append(report.Operating.Items, item)
			report.Operating.Total += netCash
		case "MANUAL":
			// Manual entries masuk ke operasional by default
			report.Operating.Items = append(report.Operating.Items, item)
			report.Operating.Total += netCash
		default:
			report.Operating.Items = append(report.Operating.Items, item)
			report.Operating.Total += netCash
		}
	}

	report.NetCashChange = report.Operating.Total + report.Investing.Total + report.Financing.Total

	// Saldo kas awal bulan (saldo periode sebelumnya)
	prevMonth, prevYear := month-1, year
	if prevMonth == 0 {
		prevMonth = 12
		prevYear--
	}
	report.BeginningCash, _ = u.repo.GetCashBalance(companyID, prevYear, prevMonth)
	report.EndingCash = report.BeginningCash + report.NetCashChange

	return report, nil
}

func (u *glPostingUsecase) GetTrialBalance(companyID uuid.UUID, month, year int) ([]domain.TrialBalanceLine, error) {
	balances, err := u.repo.GetAccountBalances(companyID, year, month)
	if err != nil {
		return nil, err
	}

	var lines []domain.TrialBalanceLine
	for _, b := range balances {
		if b.Account == nil {
			continue
		}
		lines = append(lines, domain.TrialBalanceLine{
			AccountCode: b.Account.Code,
			AccountName: b.Account.Name,
			AccountType: b.Account.AccountType,
			DebitTotal:  b.DebitTotal,
			CreditTotal: b.CreditTotal,
			Balance:     b.Balance,
		})
	}
	return lines, nil
}

// =============================================================================
// Journal Status Management
// =============================================================================

// PostJournalEntry — ubah status jurnal dari DRAFT ke POSTED dan update account_balances.
func (u *glPostingUsecase) PostJournalEntry(id uuid.UUID) error {
	entry, err := u.repo.GetJournalEntryByID(id)
	if err != nil {
		return fmt.Errorf("jurnal tidak ditemukan: %w", err)
	}
	if entry.Status != "DRAFT" {
		return fmt.Errorf("hanya jurnal berstatus DRAFT yang bisa diposting (status saat ini: %s)", entry.Status)
	}

	// Update status ke POSTED
	if err := u.db.Model(&domain.JournalEntry{}).Where("id = ?", id).Update("status", "POSTED").Error; err != nil {
		return err
	}

	// Update account_balances untuk setiap baris
	year, month := entry.EntryDate.Year(), int(entry.EntryDate.Month())
	for _, line := range entry.Lines {
		if err := u.repo.UpsertAccountBalance(line.AccountID, year, month, line.DebitAmount, line.CreditAmount); err != nil {
			fmt.Printf("[GL] warning: gagal update account_balance akun %s: %v\n", line.AccountID, err)
		}
	}
	return nil
}

// CancelJournalEntry — batalkan jurnal (DRAFT → CANCELLED). Jurnal POSTED tidak bisa dibatalkan.
func (u *glPostingUsecase) CancelJournalEntry(id uuid.UUID) error {
	entry, err := u.repo.GetJournalEntryByID(id)
	if err != nil {
		return fmt.Errorf("jurnal tidak ditemukan: %w", err)
	}
	if entry.Status == "POSTED" {
		return errors.New("jurnal yang sudah POSTED tidak bisa dibatalkan — buat jurnal koreksi (reversing entry)")
	}
	if entry.Status == "CANCELLED" {
		return errors.New("jurnal sudah berstatus CANCELLED")
	}
	return u.db.Model(&domain.JournalEntry{}).Where("id = ?", id).Update("status", "CANCELLED").Error
}

// GetProfitLossReport menghasilkan Laporan Laba Rugi untuk periode bulan/tahun.
// Menggunakan domain.ProfitLossReport dari finance.go (flat structure).
func (u *glPostingUsecase) GetProfitLossReport(companyID uuid.UUID, month, year int) (*domain.ProfitLossReport, error) {
	company, err := u.companyRepo.FindByID(companyID)
	if err != nil {
		return nil, err
	}

	// Ambil saldo akun REVENUE dan EXPENSE untuk periode ini
	var balances []domain.AccountBalance
	err = u.db.Table("account_balances ab").
		Joins("JOIN accounts a ON a.id = ab.account_id").
		Where("a.company_id = ? AND ab.period_year = ? AND ab.period_month = ? AND a.account_type IN ('REVENUE','EXPENSE')",
			companyID, year, month).
		Preload("Account").
		Find(&balances).Error
	if err != nil {
		return nil, err
	}

	report := &domain.ProfitLossReport{
		CompanyID:   companyID,
		CompanyName: company.Name,
		PeriodMonth: month,
		PeriodYear:  year,
	}

	for _, b := range balances {
		if b.Account == nil {
			continue
		}
		switch b.Account.AccountType {
		case "REVENUE":
			report.TotalRevenue += b.Balance
		case "EXPENSE":
			// Pisahkan COGS (akun 5100) dari beban lainnya
			if b.Account.Code == "5100" {
				report.TotalCOGS += b.Balance
			} else {
				report.TotalExpenses += b.Balance
			}
		}
	}

	// Juga ambil total payroll dari journal entries bulan ini
	var payrollTotal float64
	u.db.Table("journal_entries je").
		Joins("JOIN journal_entry_lines jel ON jel.journal_entry_id = je.id").
		Joins("JOIN accounts a ON a.id = jel.account_id").
		Where("je.company_id = ? AND je.reference_type = 'PAYROLL_RUN' AND EXTRACT(MONTH FROM je.entry_date) = ? AND EXTRACT(YEAR FROM je.entry_date) = ? AND je.status = 'POSTED' AND a.code = '5200'",
			companyID, month, year).
		Select("COALESCE(SUM(jel.debit_amount), 0)").
		Scan(&payrollTotal)
	report.TotalPayroll = payrollTotal

	report.GrossProfit = report.TotalRevenue - report.TotalCOGS
	report.NetProfit = report.GrossProfit - report.TotalPayroll - report.TotalExpenses
	return report, nil
}

// =============================================================================
// Chart of Accounts Management
// =============================================================================

func (u *glPostingUsecase) GetChartOfAccounts(companyID uuid.UUID) ([]domain.Account, error) {
	return u.repo.GetChartOfAccounts(companyID)
}

func (u *glPostingUsecase) CreateAccount(companyID uuid.UUID, req CreateAccountRequest) (*domain.Account, error) {
	// Validasi tipe akun
	validTypes := map[string]bool{"ASSET": true, "LIABILITY": true, "EQUITY": true, "REVENUE": true, "EXPENSE": true}
	if !validTypes[req.AccountType] {
		return nil, fmt.Errorf("account_type tidak valid: %s", req.AccountType)
	}
	if req.NormalBalance != "DEBIT" && req.NormalBalance != "CREDIT" {
		return nil, errors.New("normal_balance harus DEBIT atau CREDIT")
	}

	account := &domain.Account{
		CompanyID:      &companyID,
		Code:           req.Code,
		Name:           req.Name,
		AccountType:    req.AccountType,
		AccountSubtype: req.AccountSubtype,
		NormalBalance:  req.NormalBalance,
		ParentID:       req.ParentID,
		IsActive:       true,
		IsSystem:       false,
	}
	if err := u.repo.CreateAccount(account); err != nil {
		return nil, err
	}
	return account, nil
}

func (u *glPostingUsecase) UpdateAccount(id uuid.UUID, req UpdateAccountRequest) (*domain.Account, error) {
	account, err := u.repo.GetAccountByID(id)
	if err != nil {
		return nil, err
	}
	if account.IsSystem {
		return nil, errors.New("akun sistem tidak bisa diubah")
	}
	account.Name = req.Name
	account.AccountSubtype = req.AccountSubtype
	account.IsActive = req.IsActive
	if err := u.repo.UpdateAccount(account); err != nil {
		return nil, err
	}
	return account, nil
}

func (u *glPostingUsecase) SeedDefaultAccounts(companyID uuid.UUID) error {
	return u.repo.SeedDefaultAccounts(companyID)
}

func (u *glPostingUsecase) ClosePeriod(companyID uuid.UUID, year, month int, closedBy uuid.UUID) error {
	return u.repo.ClosePeriod(companyID, year, month, closedBy)
}

// =============================================================================
// Private helpers
// =============================================================================

// createAndPostEntry membuat journal entry + lines, lalu update account_balances.
// Memastikan double-entry balance sebelum menyimpan.
func (u *glPostingUsecase) createAndPostEntry(
	companyID uuid.UUID,
	refType string,
	refID *uuid.UUID,
	date time.Time,
	description string,
	lines []domain.JournalEntryLine,
) error {
	// Validasi double-entry balance
	var totalDebit, totalCredit float64
	for _, l := range lines {
		totalDebit += l.DebitAmount
		totalCredit += l.CreditAmount
	}
	if fmt.Sprintf("%.2f", totalDebit) != fmt.Sprintf("%.2f", totalCredit) {
		return fmt.Errorf("internal error: jurnal tidak balance (Dr %.2f ≠ Cr %.2f)", totalDebit, totalCredit)
	}

	year, month := date.Year(), int(date.Month())

	// Generate nomor jurnal
	entryNumber, err := u.repo.GenerateEntryNumber(companyID, year, month)
	if err != nil {
		return err
	}

	entry := &domain.JournalEntry{
		CompanyID:     companyID,
		EntryNumber:   entryNumber,
		EntryDate:     date,
		Description:   description,
		ReferenceType: refType,
		ReferenceID:   refID,
		Status:        "POSTED",
		Lines:         lines,
	}

	if err := u.repo.CreateJournalEntry(entry); err != nil {
		return err
	}

	// Update saldo akun
	for _, l := range lines {
		if err := u.repo.UpsertAccountBalance(l.AccountID, year, month, l.DebitAmount, l.CreditAmount); err != nil {
			// Non-fatal: log tapi jangan batalkan transaksi
			fmt.Printf("[GL] warning: gagal update account_balance untuk akun %s: %v\n", l.AccountID, err)
		}
	}

	return nil
}

// monthName mengonversi nomor bulan ke nama bulan dalam Bahasa Indonesia.
func monthName(month int) string {
	months := []string{"", "Januari", "Februari", "Maret", "April", "Mei", "Juni",
		"Juli", "Agustus", "September", "Oktober", "November", "Desember"}
	if month < 1 || month > 12 {
		return "Unknown"
	}
	return months[month]
}
