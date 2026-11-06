package repository

import (
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/lib/pq"
	"example.com/hris/backend/internal/domain"
	"gorm.io/gorm"
)

type BankAccountRepository interface {
	GetAll(companyID, branchID *uuid.UUID) ([]domain.CompanyBankAccount, error)
	GetByID(id uuid.UUID) (*domain.CompanyBankAccount, error)
	GetActive(companyID *uuid.UUID) (*domain.CompanyBankAccount, error)
	Create(account *domain.CompanyBankAccount) error
	Update(account *domain.CompanyBankAccount) error
	Delete(id uuid.UUID) error
	SetActive(id uuid.UUID, companyID *uuid.UUID) error
}

type bankAccountRepository struct {
	db *gorm.DB
}

func NewBankAccountRepository(db *gorm.DB) BankAccountRepository {
	return &bankAccountRepository{db}
}

func (r *bankAccountRepository) GetAll(companyID, branchID *uuid.UUID) ([]domain.CompanyBankAccount, error) {
	var accounts []domain.CompanyBankAccount
	query := r.db.Preload("Company").Preload("Branch").Order("is_active DESC, created_at DESC")
	if companyID != nil {
		query = query.Where("company_id = ? OR company_id IS NULL", companyID)
	}
	if branchID != nil {
		query = query.Where("branch_id = ? OR branch_id IS NULL", branchID)
	}
	err := query.Find(&accounts).Error
	return accounts, err
}

func (r *bankAccountRepository) GetByID(id uuid.UUID) (*domain.CompanyBankAccount, error) {
	var account domain.CompanyBankAccount
	err := r.db.First(&account, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &account, nil
}

func (r *bankAccountRepository) GetActive(companyID *uuid.UUID) (*domain.CompanyBankAccount, error) {
	var account domain.CompanyBankAccount

	if companyID != nil {
		err := r.db.Where("is_active = true AND company_id = ?", companyID).First(&account).Error
		if err == nil {
			return &account, nil
		}
		if err != gorm.ErrRecordNotFound {
			return nil, err
		}
	}

	err := r.db.Where("is_active = true AND company_id IS NULL").First(&account).Error
	if err != nil {
		return nil, err
	}

	return &account, nil
}

func (r *bankAccountRepository) Create(account *domain.CompanyBankAccount) error {
	return r.db.Create(account).Error
}

func (r *bankAccountRepository) Update(account *domain.CompanyBankAccount) error {
	return r.db.Save(account).Error
}

func (r *bankAccountRepository) Delete(id uuid.UUID) error {
	return bankAccountDeleteError(r.db.Delete(&domain.CompanyBankAccount{}, "id = ?", id).Error)
}

func bankAccountDeleteError(err error) error {
	if err == nil {
		return nil
	}
	var pgxErr *pgconn.PgError
	var pgErr *pq.Error
	if errors.Is(err, gorm.ErrForeignKeyViolated) ||
		(errors.As(err, &pgxErr) && pgxErr.Code == "23503") ||
		(errors.As(err, &pgErr) && pgErr.Code == "23503") {
		return domain.ErrBankAccountInUse
	}
	return err
}

func (r *bankAccountRepository) SetActive(id uuid.UUID, companyID *uuid.UUID) error {
	// Hanya mengaktifkan/menonaktifkan rekening ini (tidak lagi mematikan yang lain)
	// Kita baca dulu status terkininya
	account, err := r.GetByID(id)
	if err != nil {
		return err
	}

	newStatus := !account.IsActive
	return r.db.Model(&domain.CompanyBankAccount{}).Where("id = ?", id).Update("is_active", newStatus).Error
}
