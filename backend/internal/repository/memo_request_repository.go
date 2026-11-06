package repository

import (
	"fmt"
	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"gorm.io/gorm"
)

type MemoRequestRepository interface {
	Create(memo *domain.MemoRequest) error
	Update(memo *domain.MemoRequest) error
	FindByID(id uuid.UUID) (*domain.MemoRequest, error)
	FindAll(status string, branchID *uuid.UUID, limit, offset int) ([]domain.MemoRequest, int64, error)
	GenerateMemoNumber(branchName, monthRoman, year string) (string, error)
}

type memoRequestRepository struct {
	db *gorm.DB
}

func NewMemoRequestRepository(db *gorm.DB) MemoRequestRepository {
	return &memoRequestRepository{db}
}

func (r *memoRequestRepository) Create(memo *domain.MemoRequest) error {
	return r.db.Create(memo).Error
}

func (r *memoRequestRepository) Update(memo *domain.MemoRequest) error {
	return r.db.Save(memo).Error
}

func (r *memoRequestRepository) FindByID(id uuid.UUID) (*domain.MemoRequest, error) {
	var memo domain.MemoRequest
	err := r.db.Preload("Branch").
		Preload("Requester").
		Preload("Requester.JobPosition").
		Preload("Approver").
		Preload("Approver.JobPosition").
		Where("id = ?", id).First(&memo).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &memo, nil
}

func (r *memoRequestRepository) FindAll(status string, branchID *uuid.UUID, limit, offset int) ([]domain.MemoRequest, int64, error) {
	var memos []domain.MemoRequest
	var total int64
	query := r.db.Model(&domain.MemoRequest{})

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if branchID != nil && *branchID != uuid.Nil {
		query = query.Where("branch_id = ?", *branchID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Preload("Branch").
		Preload("Requester").
		Preload("Requester.JobPosition").
		Preload("Approver").
		Preload("Approver.JobPosition").
		Find(&memos).Error

	return memos, total, err
}

func (r *memoRequestRepository) GenerateMemoNumber(branchName, monthRoman, year string) (string, error) {
	var seq int
	// Get next value from sequence
	err := r.db.Raw("SELECT nextval('memo_requests_seq')").Scan(&seq).Error
	if err != nil {
		return "", err
	}

	// Format: {Seq}/{BranchName}/{RomanMonth}/{Year}
	// e.g. 300/Cabang-A/vii/2026
	// We'll format it in the usecase to keep it clean, but since we get seq here, we can return just the seq or format it.
	// Returning the formatted string.
	return fmt.Sprintf("%d/%s/%s/%s", seq, branchName, monthRoman, year), nil
}
