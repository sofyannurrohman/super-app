package usecase

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"example.com/hris/backend/internal/repository"
)

type MemoRequestUseCase interface {
	CreateMemoRequest(req *CreateMemoRequestPayload) error
	GetMemoRequests(status string, branchID *uuid.UUID, limit, offset int) ([]domain.MemoRequest, int64, error)
	GetMemoRequestByID(id uuid.UUID) (*domain.MemoRequest, error)
	ApproveMemoRequest(id uuid.UUID, approverID uuid.UUID) error
}

type memoRequestUseCase struct {
	repo repository.MemoRequestRepository
}

type CreateMemoRequestPayload struct {
	Type          string     `json:"type" binding:"required"`
	BranchID      *uuid.UUID `json:"branch_id"`
	Destination   string     `json:"destination" binding:"required"`
	Subject       string     `json:"subject" binding:"required"`
	RequesterName string     `json:"requester_name" binding:"required"`
	RequesterID   *uuid.UUID `json:"requester_id"`
	ApproverID    *uuid.UUID `json:"approver_id"`
}

func NewMemoRequestUseCase(repo repository.MemoRequestRepository) MemoRequestUseCase {
	return &memoRequestUseCase{repo: repo}
}

func (u *memoRequestUseCase) CreateMemoRequest(req *CreateMemoRequestPayload) error {
	memo := &domain.MemoRequest{
		Type:          req.Type,
		BranchID:      req.BranchID,
		Destination:   req.Destination,
		Subject:       req.Subject,
		RequesterName: req.RequesterName,
		RequesterID:   req.RequesterID,
		ApproverID:    req.ApproverID,
		Status:        "PENDING",
	}

	return u.repo.Create(memo)
}

func (u *memoRequestUseCase) GetMemoRequests(status string, branchID *uuid.UUID, limit, offset int) ([]domain.MemoRequest, int64, error) {
	return u.repo.FindAll(status, branchID, limit, offset)
}

func (u *memoRequestUseCase) GetMemoRequestByID(id uuid.UUID) (*domain.MemoRequest, error) {
	return u.repo.FindByID(id)
}

func (u *memoRequestUseCase) ApproveMemoRequest(id uuid.UUID, approverID uuid.UUID) error {
	memo, err := u.repo.FindByID(id)
	if err != nil {
		return err
	}
	if memo == nil {
		return errors.New("memo request not found")
	}
	if memo.Status == "APPROVED" {
		return errors.New("memo request is already approved")
	}

	// If a branch is selected, format it, otherwise default
	branchName := "Perusahaan Demo"
	if memo.Branch != nil && memo.Branch.Name != "" {
		branchName = memo.Branch.Name
	}

	now := time.Now()
	romanMonth := intToRomanMonth(now.Month())
	year := fmt.Sprintf("%d", now.Year())

	memoNumber, err := u.repo.GenerateMemoNumber(branchName, romanMonth, year)
	if err != nil {
		return fmt.Errorf("failed to generate memo number: %v", err)
	}

	memo.Status = "APPROVED"
	memo.MemoNumber = &memoNumber
	// memo.ApproverID = &approverID // if we want to overwrite, but usually it's set on creation

	return u.repo.Update(memo)
}

func intToRomanMonth(month time.Month) string {
	romans := []string{"i", "ii", "iii", "iv", "v", "vi", "vii", "viii", "ix", "x", "xi", "xii"}
	if month < 1 || month > 12 {
		return "i"
	}
	return romans[month-1]
}
