package repository

import (
	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"gorm.io/gorm"
)

type KasPabrikRepository interface {
	GetAll() ([]domain.KasPabrik, error)
	GetAllByBranch(branchID uuid.UUID) ([]domain.KasPabrik, error)
	GetByID(id uuid.UUID) (*domain.KasPabrik, error)
	Create(kas *domain.KasPabrik) error
	ListTransactions(filter domain.ListKasTransactionsFilter) ([]domain.KasTransaction, int64, error)
	GetTransactionByID(id uuid.UUID) (*domain.KasTransaction, error)
	DeleteTransaction(id uuid.UUID) error
	Update(kas *domain.KasPabrik) error
	Delete(id uuid.UUID) error
	UpdateTransaction(tx *domain.KasTransaction) error
	CreateSettlement(settlement *domain.KasAdvanceSettlement) error
	GetSettlementByTransactionID(txID uuid.UUID) (*domain.KasAdvanceSettlement, error)

	// Master Income Sources
	ListIncomeSources(onlyActive bool) ([]domain.KasIncomeSource, error)
	GetIncomeSourceByID(id uuid.UUID) (*domain.KasIncomeSource, error)
	CreateIncomeSource(src *domain.KasIncomeSource) error
	UpdateIncomeSource(src *domain.KasIncomeSource) error
	DeleteIncomeSource(id uuid.UUID) error

	// Master Expense Types
	ListExpenseTypes(onlyActive bool) ([]domain.KasExpenseType, error)
	GetExpenseTypeByID(id uuid.UUID) (*domain.KasExpenseType, error)
	CreateExpenseType(et *domain.KasExpenseType) error
	UpdateExpenseType(et *domain.KasExpenseType) error
	DeleteExpenseType(id uuid.UUID) error

	// Master Expense Accounts
	ListExpenseAccounts(typeID *uuid.UUID, onlyActive bool) ([]domain.KasExpenseAccount, error)
	GetExpenseAccountByID(id uuid.UUID) (*domain.KasExpenseAccount, error)
	CreateExpenseAccount(acc *domain.KasExpenseAccount) error
	UpdateExpenseAccount(acc *domain.KasExpenseAccount) error
	DeleteExpenseAccount(id uuid.UUID) error
}

type kasPabrikRepository struct {
	db *gorm.DB
}

func NewKasPabrikRepository(db *gorm.DB) KasPabrikRepository {
	return &kasPabrikRepository{db: db}
}

func (r *kasPabrikRepository) GetAll() ([]domain.KasPabrik, error) {
	var kas []domain.KasPabrik
	err := r.db.Order("sort_order asc, created_at desc").Find(&kas).Error
	return kas, err
}

func (r *kasPabrikRepository) GetAllByBranch(branchID uuid.UUID) ([]domain.KasPabrik, error) {
	var kas []domain.KasPabrik
	err := r.db.Where("branch_id = ?", branchID).Order("sort_order asc, created_at desc").Find(&kas).Error
	return kas, err
}

func (r *kasPabrikRepository) GetByID(id uuid.UUID) (*domain.KasPabrik, error) {
	var kas domain.KasPabrik
	err := r.db.First(&kas, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &kas, nil
}

func (r *kasPabrikRepository) Create(kas *domain.KasPabrik) error {
	return r.db.Create(kas).Error
}

func (r *kasPabrikRepository) ListTransactions(filter domain.ListKasTransactionsFilter) ([]domain.KasTransaction, int64, error) {
	var transactions []domain.KasTransaction
	var total int64

	query := r.db.Model(&domain.KasTransaction{})

	if filter.BranchID != "" {
		if bid, err := uuid.Parse(filter.BranchID); err == nil {
			query = query.Joins("JOIN kas_pabrik ON kas_pabrik.id = kas_transactions.kas_pabrik_id").
				Where("kas_pabrik.branch_id = ?", bid)
		}
	}
	if filter.KasPabrikID != "" {
		if kid, err := uuid.Parse(filter.KasPabrikID); err == nil {
			query = query.Where("kas_transactions.kas_pabrik_id = ?", kid)
		}
	}
	if filter.Type != "" {
		query = query.Where("kas_transactions.transaction_type = ?", filter.Type)
	}
	if filter.DateFrom != "" {
		query = query.Where("kas_transactions.transaction_date >= ?", filter.DateFrom)
	}
	if filter.DateTo != "" {
		query = query.Where("kas_transactions.transaction_date <= ?", filter.DateTo)
	}
	if filter.AdvanceStatus != "" {
		query = query.Where("kas_transactions.advance_status = ?", filter.AdvanceStatus)
	}
	if filter.TargetBranchID != "" {
		if tbid, err := uuid.Parse(filter.TargetBranchID); err == nil {
			query = query.Where("kas_transactions.target_branch_id = ?", tbid)
		}
	}
	if filter.EmployeeID != "" {
		if eid, err := uuid.Parse(filter.EmployeeID); err == nil {
			query = query.Where("kas_transactions.employee_id = ?", eid)
		}
	}
	if filter.ExpenseTypeID != "" {
		if etid, err := uuid.Parse(filter.ExpenseTypeID); err == nil {
			query = query.Where("kas_transactions.expense_type_id = ?", etid)
		}
	}
	if filter.ExpenseAccountID != "" {
		if eaid, err := uuid.Parse(filter.ExpenseAccountID); err == nil {
			query = query.Where("kas_transactions.expense_account_id = ?", eaid)
		}
	}
	if filter.IncomeSourceID != "" {
		if isid, err := uuid.Parse(filter.IncomeSourceID); err == nil {
			query = query.Where("kas_transactions.income_source_id = ?", isid)
		}
	}

	query.Count(&total)

	query = query.Preload("KasPabrik").
		Preload("RelatedKasPabrik").
		Preload("TargetBranch").
		Preload("Vehicle").
		Preload("Employee").
		Preload("Settlement").
		Preload("Settlement.Settler").
		Preload("IncomeSourceObj").
		Preload("ExpenseTypeObj").
		Preload("ExpenseAccountObj").
		Order("kas_transactions.transaction_date desc, kas_transactions.created_at desc")

	if filter.Limit > 0 {
		page := filter.Page
		if page < 1 {
			page = 1
		}
		offset := (page - 1) * filter.Limit
		query = query.Offset(offset).Limit(filter.Limit)
	}

	err := query.Find(&transactions).Error
	return transactions, total, err
}

func (r *kasPabrikRepository) GetTransactionByID(id uuid.UUID) (*domain.KasTransaction, error) {
	var tx domain.KasTransaction
	err := r.db.Preload("KasPabrik").
		Preload("RelatedKasPabrik").
		Preload("TargetBranch").
		Preload("Vehicle").
		Preload("Employee").
		Preload("Settlement").
		Preload("Settlement.Settler").
		Preload("IncomeSourceObj").
		Preload("ExpenseTypeObj").
		Preload("ExpenseAccountObj").
		First(&tx, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &tx, nil
}

func (r *kasPabrikRepository) DeleteTransaction(id uuid.UUID) error {
	return r.db.Delete(&domain.KasTransaction{}, "id = ?", id).Error
}

func (r *kasPabrikRepository) Update(kas *domain.KasPabrik) error {
	return r.db.Save(kas).Error
}

func (r *kasPabrikRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&domain.KasPabrik{}, "id = ?", id).Error
}

func (r *kasPabrikRepository) UpdateTransaction(tx *domain.KasTransaction) error {
	return r.db.Save(tx).Error
}

func (r *kasPabrikRepository) CreateSettlement(settlement *domain.KasAdvanceSettlement) error {
	return r.db.Create(settlement).Error
}

func (r *kasPabrikRepository) GetSettlementByTransactionID(txID uuid.UUID) (*domain.KasAdvanceSettlement, error) {
	var s domain.KasAdvanceSettlement
	err := r.db.Preload("Settler").First(&s, "kas_transaction_id = ?", txID).Error
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// Master Income Sources
func (r *kasPabrikRepository) ListIncomeSources(onlyActive bool) ([]domain.KasIncomeSource, error) {
	var list []domain.KasIncomeSource
	q := r.db.Model(&domain.KasIncomeSource{})
	if onlyActive {
		q = q.Where("is_active = ?", true)
	}
	err := q.Order("sort_order asc, created_at asc").Find(&list).Error
	return list, err
}

func (r *kasPabrikRepository) GetIncomeSourceByID(id uuid.UUID) (*domain.KasIncomeSource, error) {
	var src domain.KasIncomeSource
	err := r.db.First(&src, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &src, nil
}

func (r *kasPabrikRepository) CreateIncomeSource(src *domain.KasIncomeSource) error {
	return r.db.Create(src).Error
}

func (r *kasPabrikRepository) UpdateIncomeSource(src *domain.KasIncomeSource) error {
	return r.db.Save(src).Error
}

func (r *kasPabrikRepository) DeleteIncomeSource(id uuid.UUID) error {
	return r.db.Delete(&domain.KasIncomeSource{}, "id = ?", id).Error
}

// Master Expense Types
func (r *kasPabrikRepository) ListExpenseTypes(onlyActive bool) ([]domain.KasExpenseType, error) {
	var list []domain.KasExpenseType
	q := r.db.Model(&domain.KasExpenseType{})
	if onlyActive {
		q = q.Where("is_active = ?", true)
	}
	err := q.Preload("Accounts", func(db *gorm.DB) *gorm.DB {
		if onlyActive {
			return db.Where("is_active = ?", true).Order("sort_order asc, created_at asc")
		}
		return db.Order("sort_order asc, created_at asc")
	}).Order("sort_order asc, created_at asc").Find(&list).Error
	return list, err
}

func (r *kasPabrikRepository) GetExpenseTypeByID(id uuid.UUID) (*domain.KasExpenseType, error) {
	var et domain.KasExpenseType
	err := r.db.Preload("Accounts").First(&et, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &et, nil
}

func (r *kasPabrikRepository) CreateExpenseType(et *domain.KasExpenseType) error {
	return r.db.Create(et).Error
}

func (r *kasPabrikRepository) UpdateExpenseType(et *domain.KasExpenseType) error {
	return r.db.Save(et).Error
}

func (r *kasPabrikRepository) DeleteExpenseType(id uuid.UUID) error {
	return r.db.Delete(&domain.KasExpenseType{}, "id = ?", id).Error
}

// Master Expense Accounts
func (r *kasPabrikRepository) ListExpenseAccounts(typeID *uuid.UUID, onlyActive bool) ([]domain.KasExpenseAccount, error) {
	var list []domain.KasExpenseAccount
	q := r.db.Model(&domain.KasExpenseAccount{})
	if typeID != nil && *typeID != uuid.Nil {
		q = q.Where("expense_type_id = ?", *typeID)
	}
	if onlyActive {
		q = q.Where("is_active = ?", true)
	}
	err := q.Preload("ExpenseType").Order("sort_order asc, created_at asc").Find(&list).Error
	return list, err
}

func (r *kasPabrikRepository) GetExpenseAccountByID(id uuid.UUID) (*domain.KasExpenseAccount, error) {
	var acc domain.KasExpenseAccount
	err := r.db.Preload("ExpenseType").First(&acc, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &acc, nil
}

func (r *kasPabrikRepository) CreateExpenseAccount(acc *domain.KasExpenseAccount) error {
	return r.db.Create(acc).Error
}

func (r *kasPabrikRepository) UpdateExpenseAccount(acc *domain.KasExpenseAccount) error {
	return r.db.Omit("ExpenseType").Save(acc).Error
}

func (r *kasPabrikRepository) DeleteExpenseAccount(id uuid.UUID) error {
	return r.db.Delete(&domain.KasExpenseAccount{}, "id = ?", id).Error
}
