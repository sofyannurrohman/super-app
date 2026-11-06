package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"example.com/hris/backend/internal/domain"
)

type BranchDepositRepository interface {
	Create(ctx context.Context, deposit *domain.BranchDeposit) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.BranchDeposit, error)
	FindAll(ctx context.Context, branchID *uuid.UUID, status *string, startDate, endDate *time.Time) ([]domain.BranchDeposit, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.BranchDepositStatus, approvedBy uuid.UUID, journalEntryID *uuid.UUID) error
}

type branchDepositRepository struct {
	db *gorm.DB
}

func NewBranchDepositRepository(db *gorm.DB) BranchDepositRepository {
	return &branchDepositRepository{db: db}
}

func (r *branchDepositRepository) Create(ctx context.Context, deposit *domain.BranchDeposit) error {
	return r.db.WithContext(ctx).Create(deposit).Error
}

func (r *branchDepositRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.BranchDeposit, error) {
	var deposit domain.BranchDeposit
	err := r.db.WithContext(ctx).
		Preload("Branch").
		Preload("BankAccount").
		Preload("Submitter").
		Preload("Submitter.Employee").
		Preload("Submitter.Employee.JobPosition").
		Preload("Approver").
		Preload("Approver.Employee").
		Preload("Approver.Employee.JobPosition").
		Where("id = ?", id).First(&deposit).Error
	if err != nil {
		return nil, err
	}
	return &deposit, nil
}

func (r *branchDepositRepository) FindAll(ctx context.Context, branchID *uuid.UUID, status *string, startDate, endDate *time.Time) ([]domain.BranchDeposit, error) {
	var deposits []domain.BranchDeposit
	query := r.db.WithContext(ctx).
		Preload("Branch").
		Preload("BankAccount").
		Preload("Submitter").
		Preload("Submitter.Employee").
		Preload("Submitter.Employee.JobPosition").
		Preload("Approver").
		Preload("Approver.Employee").
		Preload("Approver.Employee.JobPosition").
		Order("created_at desc")

	if branchID != nil && *branchID != uuid.Nil {
		query = query.Where("branch_id = ?", *branchID)
	}
	if status != nil && *status != "" {
		query = query.Where("status = ?", *status)
	}
	if startDate != nil && endDate != nil {
		query = query.Where("deposit_date BETWEEN ? AND ?", *startDate, *endDate)
	}

	err := query.Find(&deposits).Error
	return deposits, err
}

func (r *branchDepositRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.BranchDepositStatus, approvedBy uuid.UUID, journalEntryID *uuid.UUID) error {
	updates := map[string]interface{}{
		"status":      status,
		"approved_by": approvedBy,
		"updated_at":  time.Now(),
	}
	if journalEntryID != nil && *journalEntryID != uuid.Nil {
		updates["journal_entry_id"] = *journalEntryID
	}

	return r.db.WithContext(ctx).Model(&domain.BranchDeposit{}).Where("id = ?", id).Updates(updates).Error
}
