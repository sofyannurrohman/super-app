package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"example.com/hris/backend/internal/domain"
	"example.com/hris/backend/internal/repository"
)

type BranchDepositUsecase interface {
	SubmitDeposit(ctx context.Context, req SubmitDepositRequest, submittedBy uuid.UUID) (*domain.BranchDeposit, error)
	GetDeposits(ctx context.Context, branchID *uuid.UUID, status *string, startDate, endDate *time.Time) ([]domain.BranchDeposit, error)
	GetDailyBalance(ctx context.Context, branchID uuid.UUID, date time.Time) (float64, error)
	ApproveDeposit(ctx context.Context, depositID uuid.UUID, approvedBy uuid.UUID) error
	RejectDeposit(ctx context.Context, depositID uuid.UUID, approvedBy uuid.UUID) error
}

type branchDepositUsecase struct {
	repo       repository.BranchDepositRepository
	glPosting  GLPostingUsecase
	branchRepo repository.BranchRepository
	salesRepo  repository.SalesTransactionRepository
}

func NewBranchDepositUsecase(
	repo repository.BranchDepositRepository,
	glPosting GLPostingUsecase,
	branchRepo repository.BranchRepository,
	salesRepo repository.SalesTransactionRepository,
) BranchDepositUsecase {
	return &branchDepositUsecase{
		repo:       repo,
		glPosting:  glPosting,
		branchRepo: branchRepo,
		salesRepo:  salesRepo,
	}
}

type SubmitDepositRequest struct {
	CompanyID     uuid.UUID  `json:"company_id"`
	BranchID      uuid.UUID  `json:"branch_id"`
	DepositDate   time.Time  `json:"deposit_date"`
	Amount        float64    `json:"amount"`
	PaymentMethod string     `json:"payment_method"`
	BankAccountID *uuid.UUID `json:"bank_account_id"`
	ProofURL      string     `json:"proof_url"`
	Notes         string     `json:"notes"`
}

func (u *branchDepositUsecase) SubmitDeposit(ctx context.Context, req SubmitDepositRequest, submittedBy uuid.UUID) (*domain.BranchDeposit, error) {
	if req.Amount <= 0 {
		return nil, errors.New("jumlah setoran harus lebih dari 0")
	}

	method := domain.PaymentMethod(req.PaymentMethod)
	if method != domain.PaymentMethodCash && method != domain.PaymentMethodTransfer {
		return nil, errors.New("metode pembayaran tidak valid")
	}

	if method == domain.PaymentMethodTransfer && (req.BankAccountID == nil || *req.BankAccountID == uuid.Nil) {
		return nil, errors.New("bank_account_id wajib diisi untuk metode transfer")
	}

	startOfDay := time.Date(req.DepositDate.Year(), req.DepositDate.Month(), req.DepositDate.Day(), 0, 0, 0, 0, req.DepositDate.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	var totalAdminNota float64
	var err error
	if method == domain.PaymentMethodCash {
		totalAdminNota, err = u.salesRepo.GetTotalCashValidatedByAdminNota(ctx, req.BranchID, startOfDay, endOfDay)
		if err != nil {
			return nil, err
		}
	}

	deposit := &domain.BranchDeposit{
		CompanyID:     req.CompanyID,
		BranchID:      req.BranchID,
		DepositDate:   req.DepositDate,
		Amount:        req.Amount,
		SystemAmount:  totalAdminNota,
		PaymentMethod: method,
		BankAccountID: req.BankAccountID,
		ProofURL:      req.ProofURL,
		Notes:         req.Notes,
		Status:        domain.DepositStatusPending,
		SubmittedBy:   submittedBy,
	}

	err = u.repo.Create(ctx, deposit)
	if err != nil {
		return nil, err
	}

	return deposit, nil
}

func (u *branchDepositUsecase) GetDailyBalance(ctx context.Context, branchID uuid.UUID, date time.Time) (float64, error) {
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)
	return u.salesRepo.GetTotalCashValidatedByAdminNota(ctx, branchID, startOfDay, endOfDay)
}

func (u *branchDepositUsecase) GetDeposits(ctx context.Context, branchID *uuid.UUID, status *string, startDate, endDate *time.Time) ([]domain.BranchDeposit, error) {
	return u.repo.FindAll(ctx, branchID, status, startDate, endDate)
}

func (u *branchDepositUsecase) ApproveDeposit(ctx context.Context, depositID uuid.UUID, approvedBy uuid.UUID) error {
	deposit, err := u.repo.GetByID(ctx, depositID)
	if err != nil {
		return err
	}

	if deposit.Status != domain.DepositStatusPending {
		return errors.New("hanya setoran berstatus PENDING yang dapat di-approve")
	}

	// Buat Jurnal Akuntansi via GL Posting Usecase
	isCash := deposit.PaymentMethod == domain.PaymentMethodCash
	// Posting jurnal ke GL, sementara kita pass nil untuk ID Jurnal karena implementasi post tidak return UUID
	_, err = u.glPosting.PostBranchDeposit(deposit.CompanyID, deposit.ID, deposit.Amount, isCash, deposit.DepositDate)
	if err != nil {
		return err
	}

	// Update status menjadi APPROVED
	return u.repo.UpdateStatus(ctx, depositID, domain.DepositStatusApproved, approvedBy, nil)
}

func (u *branchDepositUsecase) RejectDeposit(ctx context.Context, depositID uuid.UUID, approvedBy uuid.UUID) error {
	deposit, err := u.repo.GetByID(ctx, depositID)
	if err != nil {
		return err
	}

	if deposit.Status != domain.DepositStatusPending {
		return errors.New("hanya setoran berstatus PENDING yang dapat di-reject")
	}

	// Update status menjadi REJECTED
	return u.repo.UpdateStatus(ctx, depositID, domain.DepositStatusRejected, approvedBy, nil)
}
