package repository

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// =============================================================================
// Interface
// =============================================================================

type AccountingRepository interface {
	// Chart of Accounts
	GetChartOfAccounts(companyID uuid.UUID) ([]domain.Account, error)
	GetAccountByCode(companyID uuid.UUID, code string) (*domain.Account, error)
	GetAccountByID(id uuid.UUID) (*domain.Account, error)
	CreateAccount(account *domain.Account) error
	UpdateAccount(account *domain.Account) error

	// Journal Entries
	CreateJournalEntry(entry *domain.JournalEntry) error
	GetJournalEntries(companyID uuid.UUID, month, year, page, limit int) ([]domain.JournalEntry, int64, error)
	GetJournalEntryByRef(referenceType string, referenceID uuid.UUID) (*domain.JournalEntry, error)
	GetJournalEntryByID(id uuid.UUID) (*domain.JournalEntry, error)
	GenerateEntryNumber(companyID uuid.UUID, year, month int) (string, error)

	// Account Balances
	UpsertAccountBalance(accountID uuid.UUID, year, month int, debit, credit float64) error
	GetAccountBalances(companyID uuid.UUID, year, month int) ([]domain.AccountBalance, error)
	GetAccountBalanceSumByType(companyID uuid.UUID, accountType string, year, month int) (float64, error)
	GetAccountBalanceSumBySubtype(companyID uuid.UUID, accountSubtype string, year, month int) (float64, error)
	GetCashBalance(companyID uuid.UUID, year, month int) (float64, error)

	// Accounting Period
	GetOrCreatePeriod(companyID uuid.UUID, year, month int) (*domain.AccountingPeriod, error)
	ClosePeriod(companyID uuid.UUID, year, month int, closedBy uuid.UUID) error

	// Seed default Chart of Accounts (PSAK)
	SeedDefaultAccounts(companyID uuid.UUID) error
}

// =============================================================================
// Implementation
// =============================================================================

type accountingRepository struct {
	db *gorm.DB
}

func NewAccountingRepository(db *gorm.DB) AccountingRepository {
	return &accountingRepository{db: db}
}

// --- Chart of Accounts ---

func (r *accountingRepository) GetChartOfAccounts(companyID uuid.UUID) ([]domain.Account, error) {
	var accounts []domain.Account
	err := r.db.Where("company_id = ? AND is_active = true", companyID).
		Order("code ASC").
		Find(&accounts).Error
	return accounts, err
}

func (r *accountingRepository) GetAccountByCode(companyID uuid.UUID, code string) (*domain.Account, error) {
	var account domain.Account
	err := r.db.Where("company_id = ? AND code = ?", companyID, code).First(&account).Error
	if err != nil {
		return nil, err
	}
	return &account, nil
}

func (r *accountingRepository) GetAccountByID(id uuid.UUID) (*domain.Account, error) {
	var account domain.Account
	err := r.db.First(&account, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &account, nil
}

func (r *accountingRepository) CreateAccount(account *domain.Account) error {
	return r.db.Create(account).Error
}

func (r *accountingRepository) UpdateAccount(account *domain.Account) error {
	return r.db.Save(account).Error
}

// --- Journal Entries ---

func (r *accountingRepository) CreateJournalEntry(entry *domain.JournalEntry) error {
	return r.db.Create(entry).Error
}

func (r *accountingRepository) GetJournalEntries(companyID uuid.UUID, month, year, page, limit int) ([]domain.JournalEntry, int64, error) {
	var entries []domain.JournalEntry
	var total int64

	q := r.db.Model(&domain.JournalEntry{}).
		Where("company_id = ?", companyID)

	if month > 0 && year > 0 {
		q = q.Where("EXTRACT(MONTH FROM entry_date) = ? AND EXTRACT(YEAR FROM entry_date) = ?", month, year)
	}

	q.Count(&total)

	offset := (page - 1) * limit
	err := q.Preload("Lines.Account").
		Order("entry_date DESC, created_at DESC").
		Offset(offset).Limit(limit).
		Find(&entries).Error

	return entries, total, err
}

func (r *accountingRepository) GetJournalEntryByRef(referenceType string, referenceID uuid.UUID) (*domain.JournalEntry, error) {
	var entry domain.JournalEntry
	err := r.db.Preload("Lines.Account").
		Where("reference_type = ? AND reference_id = ?", referenceType, referenceID).
		First(&entry).Error
	if err != nil {
		return nil, err
	}
	return &entry, nil
}

func (r *accountingRepository) GetJournalEntryByID(id uuid.UUID) (*domain.JournalEntry, error) {
	var entry domain.JournalEntry
	err := r.db.Preload("Lines.Account").First(&entry, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &entry, nil
}

func (r *accountingRepository) GenerateEntryNumber(companyID uuid.UUID, year, month int) (string, error) {
	var count int64
	r.db.Model(&domain.JournalEntry{}).
		Where("company_id = ? AND EXTRACT(YEAR FROM entry_date) = ? AND EXTRACT(MONTH FROM entry_date) = ?",
			companyID, year, month).
		Count(&count)
	return fmt.Sprintf("JRN-%d-%02d-%04d", year, month, count+1), nil
}

// --- Account Balances ---

// UpsertAccountBalance menambah debit/credit ke saldo akun untuk periode tertentu.
// Menggunakan ON CONFLICT DO UPDATE (upsert) agar idempotent.
func (r *accountingRepository) UpsertAccountBalance(accountID uuid.UUID, year, month int, debit, credit float64) error {
	// Ambil akun untuk menentukan normal balance
	var account domain.Account
	if err := r.db.First(&account, "id = ?", accountID).Error; err != nil {
		return err
	}

	now := time.Now()

	// Hitung balance berdasarkan normal balance akun
	var balanceDelta float64
	if account.NormalBalance == "DEBIT" {
		balanceDelta = debit - credit
	} else {
		balanceDelta = credit - debit
	}

	// Upsert ke account_balances
	return r.db.Exec(`
		INSERT INTO account_balances (id, account_id, period_year, period_month, debit_total, credit_total, balance, updated_at)
		VALUES (gen_random_uuid(), ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (account_id, period_year, period_month)
		DO UPDATE SET
			debit_total  = account_balances.debit_total  + EXCLUDED.debit_total,
			credit_total = account_balances.credit_total + EXCLUDED.credit_total,
			balance      = account_balances.balance      + ?,
			updated_at   = EXCLUDED.updated_at
	`, accountID, year, month, debit, credit, balanceDelta, now, balanceDelta).Error
}

func (r *accountingRepository) GetAccountBalances(companyID uuid.UUID, year, month int) ([]domain.AccountBalance, error) {
	var balances []domain.AccountBalance
	err := r.db.Joins("JOIN accounts ON accounts.id = account_balances.account_id").
		Where("accounts.company_id = ? AND account_balances.period_year = ? AND account_balances.period_month = ?",
			companyID, year, month).
		Preload("Account").
		Find(&balances).Error
	return balances, err
}

func (r *accountingRepository) GetAccountBalanceSumByType(companyID uuid.UUID, accountType string, year, month int) (float64, error) {
	var total float64
	err := r.db.Table("account_balances ab").
		Joins("JOIN accounts a ON a.id = ab.account_id").
		Where("a.company_id = ? AND a.account_type = ? AND ab.period_year = ? AND ab.period_month = ?",
			companyID, accountType, year, month).
		Select("COALESCE(SUM(ab.balance), 0)").
		Scan(&total).Error
	return total, err
}

func (r *accountingRepository) GetAccountBalanceSumBySubtype(companyID uuid.UUID, accountSubtype string, year, month int) (float64, error) {
	var total float64
	err := r.db.Table("account_balances ab").
		Joins("JOIN accounts a ON a.id = ab.account_id").
		Where("a.company_id = ? AND a.account_subtype = ? AND ab.period_year = ? AND ab.period_month = ?",
			companyID, accountSubtype, year, month).
		Select("COALESCE(SUM(ab.balance), 0)").
		Scan(&total).Error
	return total, err
}

func (r *accountingRepository) GetCashBalance(companyID uuid.UUID, year, month int) (float64, error) {
	var balance float64
	err := r.db.Table("account_balances ab").
		Joins("JOIN accounts a ON a.id = ab.account_id").
		Where("a.company_id = ? AND a.code = '1100' AND ab.period_year = ? AND ab.period_month = ?",
			companyID, year, month).
		Select("COALESCE(SUM(ab.balance), 0)").
		Scan(&balance).Error
	return balance, err
}

// --- Accounting Period ---

func (r *accountingRepository) GetOrCreatePeriod(companyID uuid.UUID, year, month int) (*domain.AccountingPeriod, error) {
	period := domain.AccountingPeriod{
		CompanyID:   companyID,
		PeriodYear:  year,
		PeriodMonth: month,
		Status:      "OPEN",
	}
	result := r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "company_id"}, {Name: "period_year"}, {Name: "period_month"}},
		DoNothing: true,
	}).Create(&period)

	if result.Error != nil {
		return nil, result.Error
	}

	// Ambil periode (mungkin sudah ada sebelumnya)
	var existing domain.AccountingPeriod
	err := r.db.Where("company_id = ? AND period_year = ? AND period_month = ?",
		companyID, year, month).First(&existing).Error
	return &existing, err
}

func (r *accountingRepository) ClosePeriod(companyID uuid.UUID, year, month int, closedBy uuid.UUID) error {
	now := time.Now()
	return r.db.Model(&domain.AccountingPeriod{}).
		Where("company_id = ? AND period_year = ? AND period_month = ?", companyID, year, month).
		Updates(map[string]interface{}{
			"status":    "CLOSED",
			"closed_at": now,
			"closed_by": closedBy,
		}).Error
}

// --- Seed Default Chart of Accounts (PSAK Indonesia) ---

// SeedDefaultAccounts membuat bagan akun standar PSAK untuk perusahaan baru.
// Hanya dieksekusi jika perusahaan belum memiliki akun sama sekali.
func (r *accountingRepository) SeedDefaultAccounts(companyID uuid.UUID) error {
	var count int64
	r.db.Model(&domain.Account{}).Where("company_id = ?", companyID).Count(&count)
	if count > 0 {
		return nil // sudah ada akun, skip
	}

	defaults := []domain.Account{
		// === ASET (1xxx) ===
		{CompanyID: &companyID, Code: "1000", Name: "ASET", AccountType: "ASSET", NormalBalance: "DEBIT", IsSystem: true},
		{CompanyID: &companyID, Code: "1100", Name: "Kas & Bank", AccountType: "ASSET", AccountSubtype: "CURRENT_ASSET", NormalBalance: "DEBIT", IsSystem: true},
		{CompanyID: &companyID, Code: "1110", Name: "Kas Tunai", AccountType: "ASSET", AccountSubtype: "CURRENT_ASSET", NormalBalance: "DEBIT", IsSystem: true},
		{CompanyID: &companyID, Code: "1120", Name: "Bank BCA", AccountType: "ASSET", AccountSubtype: "CURRENT_ASSET", NormalBalance: "DEBIT", IsSystem: false},
		{CompanyID: &companyID, Code: "1130", Name: "Bank Mandiri", AccountType: "ASSET", AccountSubtype: "CURRENT_ASSET", NormalBalance: "DEBIT", IsSystem: false},
		{CompanyID: &companyID, Code: "1200", Name: "Piutang Usaha", AccountType: "ASSET", AccountSubtype: "CURRENT_ASSET", NormalBalance: "DEBIT", IsSystem: true},
		{CompanyID: &companyID, Code: "1300", Name: "Persediaan Barang", AccountType: "ASSET", AccountSubtype: "CURRENT_ASSET", NormalBalance: "DEBIT", IsSystem: true},
		{CompanyID: &companyID, Code: "1400", Name: "Uang Muka & Beban Dibayar Dimuka", AccountType: "ASSET", AccountSubtype: "CURRENT_ASSET", NormalBalance: "DEBIT", IsSystem: false},
		{CompanyID: &companyID, Code: "1500", Name: "Aset Tetap", AccountType: "ASSET", AccountSubtype: "FIXED_ASSET", NormalBalance: "DEBIT", IsSystem: true},
		{CompanyID: &companyID, Code: "1510", Name: "Kendaraan", AccountType: "ASSET", AccountSubtype: "FIXED_ASSET", NormalBalance: "DEBIT", IsSystem: false},
		{CompanyID: &companyID, Code: "1520", Name: "Peralatan & Inventaris", AccountType: "ASSET", AccountSubtype: "FIXED_ASSET", NormalBalance: "DEBIT", IsSystem: false},
		{CompanyID: &companyID, Code: "1590", Name: "Akumulasi Depresiasi", AccountType: "ASSET", AccountSubtype: "FIXED_ASSET", NormalBalance: "CREDIT", IsSystem: false},

		// === KEWAJIBAN (2xxx) ===
		{CompanyID: &companyID, Code: "2000", Name: "KEWAJIBAN", AccountType: "LIABILITY", NormalBalance: "CREDIT", IsSystem: true},
		{CompanyID: &companyID, Code: "2100", Name: "Hutang Usaha", AccountType: "LIABILITY", AccountSubtype: "CURRENT_LIABILITY", NormalBalance: "CREDIT", IsSystem: true},
		{CompanyID: &companyID, Code: "2200", Name: "Hutang Gaji", AccountType: "LIABILITY", AccountSubtype: "CURRENT_LIABILITY", NormalBalance: "CREDIT", IsSystem: true},
		{CompanyID: &companyID, Code: "2300", Name: "Hutang Pajak", AccountType: "LIABILITY", AccountSubtype: "CURRENT_LIABILITY", NormalBalance: "CREDIT", IsSystem: true},
		{CompanyID: &companyID, Code: "2310", Name: "Hutang PPh 21", AccountType: "LIABILITY", AccountSubtype: "CURRENT_LIABILITY", NormalBalance: "CREDIT", IsSystem: true},
		{CompanyID: &companyID, Code: "2320", Name: "Hutang PPN", AccountType: "LIABILITY", AccountSubtype: "CURRENT_LIABILITY", NormalBalance: "CREDIT", IsSystem: false},
		{CompanyID: &companyID, Code: "2400", Name: "Hutang BPJS", AccountType: "LIABILITY", AccountSubtype: "CURRENT_LIABILITY", NormalBalance: "CREDIT", IsSystem: true},
		{CompanyID: &companyID, Code: "2500", Name: "Hutang Jangka Panjang", AccountType: "LIABILITY", AccountSubtype: "NON_CURRENT_LIABILITY", NormalBalance: "CREDIT", IsSystem: false},

		// === EKUITAS (3xxx) ===
		{CompanyID: &companyID, Code: "3000", Name: "EKUITAS", AccountType: "EQUITY", NormalBalance: "CREDIT", IsSystem: true},
		{CompanyID: &companyID, Code: "3100", Name: "Modal Disetor", AccountType: "EQUITY", AccountSubtype: "PAID_IN_CAPITAL", NormalBalance: "CREDIT", IsSystem: true},
		{CompanyID: &companyID, Code: "3200", Name: "Laba Ditahan", AccountType: "EQUITY", AccountSubtype: "RETAINED_EARNINGS", NormalBalance: "CREDIT", IsSystem: true},
		{CompanyID: &companyID, Code: "3300", Name: "Laba/Rugi Tahun Berjalan", AccountType: "EQUITY", AccountSubtype: "CURRENT_EARNINGS", NormalBalance: "CREDIT", IsSystem: true},

		// === PENDAPATAN (4xxx) ===
		{CompanyID: &companyID, Code: "4000", Name: "PENDAPATAN", AccountType: "REVENUE", NormalBalance: "CREDIT", IsSystem: true},
		{CompanyID: &companyID, Code: "4100", Name: "Pendapatan Penjualan", AccountType: "REVENUE", AccountSubtype: "OPERATING_REVENUE", NormalBalance: "CREDIT", IsSystem: true},
		{CompanyID: &companyID, Code: "4200", Name: "Retur & Potongan Penjualan", AccountType: "REVENUE", AccountSubtype: "CONTRA_REVENUE", NormalBalance: "DEBIT", IsSystem: true},
		{CompanyID: &companyID, Code: "4900", Name: "Pendapatan Lain-lain", AccountType: "REVENUE", AccountSubtype: "OTHER_REVENUE", NormalBalance: "CREDIT", IsSystem: false},

		// === BEBAN (5xxx) ===
		{CompanyID: &companyID, Code: "5000", Name: "BEBAN", AccountType: "EXPENSE", NormalBalance: "DEBIT", IsSystem: true},
		{CompanyID: &companyID, Code: "5100", Name: "Harga Pokok Penjualan (HPP)", AccountType: "EXPENSE", AccountSubtype: "COGS", NormalBalance: "DEBIT", IsSystem: true},
		{CompanyID: &companyID, Code: "5200", Name: "Beban Gaji & Tunjangan", AccountType: "EXPENSE", AccountSubtype: "OPERATING_EXPENSE", NormalBalance: "DEBIT", IsSystem: true},
		{CompanyID: &companyID, Code: "5210", Name: "Beban BPJS Perusahaan", AccountType: "EXPENSE", AccountSubtype: "OPERATING_EXPENSE", NormalBalance: "DEBIT", IsSystem: true},
		{CompanyID: &companyID, Code: "5300", Name: "Beban Operasional", AccountType: "EXPENSE", AccountSubtype: "OPERATING_EXPENSE", NormalBalance: "DEBIT", IsSystem: true},
		{CompanyID: &companyID, Code: "5310", Name: "Beban Transportasi & BBM", AccountType: "EXPENSE", AccountSubtype: "OPERATING_EXPENSE", NormalBalance: "DEBIT", IsSystem: false},
		{CompanyID: &companyID, Code: "5320", Name: "Beban Promosi & Pemasaran", AccountType: "EXPENSE", AccountSubtype: "OPERATING_EXPENSE", NormalBalance: "DEBIT", IsSystem: false},
		{CompanyID: &companyID, Code: "5400", Name: "Beban Reimbursement", AccountType: "EXPENSE", AccountSubtype: "OPERATING_EXPENSE", NormalBalance: "DEBIT", IsSystem: true},
		{CompanyID: &companyID, Code: "5500", Name: "Beban Penyusutan", AccountType: "EXPENSE", AccountSubtype: "OPERATING_EXPENSE", NormalBalance: "DEBIT", IsSystem: false},
		{CompanyID: &companyID, Code: "5900", Name: "Beban Lain-lain", AccountType: "EXPENSE", AccountSubtype: "OTHER_EXPENSE", NormalBalance: "DEBIT", IsSystem: false},
	}

	return r.db.Create(&defaults).Error
}
