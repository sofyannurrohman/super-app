package usecase

import (
	"errors"

	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"example.com/hris/backend/internal/repository"
)

type BankAccountUsecase interface {
	GetAll(companyID, branchID *uuid.UUID) ([]domain.CompanyBankAccount, error)
	GetByID(id uuid.UUID) (*domain.CompanyBankAccount, error)
	GetActive(companyID *uuid.UUID) (*domain.CompanyBankAccount, error)
	Create(req CreateBankAccountRequest) (*domain.CompanyBankAccount, error)
	Update(id uuid.UUID, req UpdateBankAccountRequest) (*domain.CompanyBankAccount, error)
	Delete(id uuid.UUID) error
	SetActive(id uuid.UUID) error
}

type CreateBankAccountRequest struct {
	CompanyID     *uuid.UUID `json:"company_id"`
	BranchID      *uuid.UUID `json:"branch_id"`
	BankName      string     `json:"bank_name" binding:"required"`
	AccountNumber string     `json:"account_number" binding:"required"`
	AccountHolder string     `json:"account_holder" binding:"required"`
	BankLogoURL   string     `json:"bank_logo_url"`
	ManualBalance float64    `json:"manual_balance"`
	Notes         string     `json:"notes"`
	IsActive      bool       `json:"is_active"`
}

type UpdateBankAccountRequest struct {
	CompanyID     *uuid.UUID `json:"company_id"`
	BranchID      *uuid.UUID `json:"branch_id"`
	BankName      string     `json:"bank_name"`
	AccountNumber string     `json:"account_number"`
	AccountHolder string     `json:"account_holder"`
	BankLogoURL   string     `json:"bank_logo_url"`
	ManualBalance *float64   `json:"manual_balance"`
	Notes         string     `json:"notes"`
	IsActive      *bool      `json:"is_active"`
}

type bankAccountUsecase struct {
	repo repository.BankAccountRepository
}

func NewBankAccountUsecase(repo repository.BankAccountRepository) BankAccountUsecase {
	return &bankAccountUsecase{repo}
}

func (u *bankAccountUsecase) GetAll(companyID, branchID *uuid.UUID) ([]domain.CompanyBankAccount, error) {
	return u.repo.GetAll(companyID, branchID)
}

func (u *bankAccountUsecase) GetByID(id uuid.UUID) (*domain.CompanyBankAccount, error) {
	return u.repo.GetByID(id)
}

func (u *bankAccountUsecase) GetActive(companyID *uuid.UUID) (*domain.CompanyBankAccount, error) {
	return u.repo.GetActive(companyID)
}

func (u *bankAccountUsecase) Create(req CreateBankAccountRequest) (*domain.CompanyBankAccount, error) {
	if req.BankName == "" || req.AccountNumber == "" || req.AccountHolder == "" {
		return nil, errors.New("bank_name, account_number, dan account_holder wajib diisi")
	}

	account := &domain.CompanyBankAccount{
		CompanyID:     req.CompanyID,
		BranchID:      req.BranchID,
		BankName:      req.BankName,
		AccountNumber: req.AccountNumber,
		AccountHolder: req.AccountHolder,
		BankLogoURL:   req.BankLogoURL,
		ManualBalance: req.ManualBalance,
		Notes:         req.Notes,
		IsActive:      false,
	}

	if err := u.repo.Create(account); err != nil {
		return nil, err
	}

	if req.IsActive {
		if err := u.repo.SetActive(account.ID, account.CompanyID); err != nil {
			return nil, err
		}
		account.IsActive = true
	}

	return account, nil
}

func (u *bankAccountUsecase) Update(id uuid.UUID, req UpdateBankAccountRequest) (*domain.CompanyBankAccount, error) {
	account, err := u.repo.GetByID(id)
	if err != nil {
		return nil, errors.New("rekening bank tidak ditemukan")
	}

	// Because BranchID could be set to null explicitly (optional), we might need a way to clear it.
	// We'll update it directly since it's a pointer.
	account.CompanyID = req.CompanyID
	account.BranchID = req.BranchID
	if req.BankName != "" {
		account.BankName = req.BankName
	}
	if req.AccountNumber != "" {
		account.AccountNumber = req.AccountNumber
	}
	if req.AccountHolder != "" {
		account.AccountHolder = req.AccountHolder
	}
	if req.BankLogoURL != "" {
		account.BankLogoURL = req.BankLogoURL
	}
	if req.ManualBalance != nil {
		account.ManualBalance = *req.ManualBalance
	}
	account.Notes = req.Notes

	if req.IsActive != nil {
		account.IsActive = *req.IsActive
	}

	if err := u.repo.Update(account); err != nil {
		return nil, err
	}

	return account, nil
}

func (u *bankAccountUsecase) Delete(id uuid.UUID) error {
	_, err := u.repo.GetByID(id)
	if err != nil {
		return errors.New("rekening bank tidak ditemukan")
	}
	return u.repo.Delete(id)
}

func (u *bankAccountUsecase) SetActive(id uuid.UUID) error {
	account, err := u.repo.GetByID(id)
	if err != nil {
		return errors.New("rekening bank tidak ditemukan")
	}
	return u.repo.SetActive(id, account.CompanyID)
}
