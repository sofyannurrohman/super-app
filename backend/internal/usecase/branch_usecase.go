package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"example.com/hris/backend/internal/repository"
	"example.com/hris/backend/pkg/redisclient"
)

type BranchUseCase interface {
	CreateBranch(req *CreateBranchRequest) error
	GetBranches() ([]domain.Branch, error)
	GetBranchesByCompany(companyID uuid.UUID) ([]domain.Branch, error)
	GetBranchByID(id uuid.UUID) (*domain.Branch, error)
	UpdateBranch(id uuid.UUID, req *UpdateBranchRequest) error
	DeleteBranch(id uuid.UUID) error
	SetProductPrice(branchID, productID uuid.UUID, het float64) error
	GetProductPrices(branchID uuid.UUID) ([]domain.BranchProductPrice, error)
}

type branchUseCase struct {
	repo repository.BranchRepository
}

type CreateBranchRequest struct {
	CompanyID   *uuid.UUID `json:"companyId" binding:"required"`
	Name        string     `json:"name" binding:"required"`
	Address     string     `json:"address"`
	Timezone    string     `json:"timezone"`
	Latitude    float64    `json:"latitude"`
	Longitude   float64    `json:"longitude"`
	RadiusMeter int        `json:"radiusMeter"`
}

type UpdateBranchRequest struct {
	CompanyID   string  `json:"companyId" binding:"omitempty"`
	Name        string  `json:"name" binding:"required"`
	Address     string  `json:"address"`
	Timezone    string  `json:"timezone"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	RadiusMeter int     `json:"radiusMeter"`
}

func NewBranchUseCase(repo repository.BranchRepository) BranchUseCase {
	return &branchUseCase{repo: repo}
}

func (u *branchUseCase) CreateBranch(req *CreateBranchRequest) error {
	timezone := req.Timezone
	if timezone == "" {
		timezone = "Asia/Jakarta"
	}
	radiusMeter := req.RadiusMeter
	if radiusMeter == 0 {
		radiusMeter = 100
	}

	branch := &domain.Branch{
		CompanyID:   req.CompanyID,
		Name:        req.Name,
		Address:     req.Address,
		Timezone:    timezone,
		Latitude:    req.Latitude,
		Longitude:   req.Longitude,
		RadiusMeter: radiusMeter,
	}

	return u.repo.Create(branch)
}

func (u *branchUseCase) GetBranches() ([]domain.Branch, error) {
	return u.repo.FindAll()
}

func (u *branchUseCase) GetBranchesByCompany(companyID uuid.UUID) ([]domain.Branch, error) {
	return u.repo.FindByCompanyID(companyID)
}

func (u *branchUseCase) GetBranchByID(id uuid.UUID) (*domain.Branch, error) {
	ctx := context.Background()
	cacheKey := fmt.Sprintf("branch:id:%s", id.String())

	if redisclient.Client != nil {
		if cached, err := redisclient.Client.Get(ctx, cacheKey).Result(); err == nil {
			var branch domain.Branch
			if json.Unmarshal([]byte(cached), &branch) == nil {
				return &branch, nil
			}
		}
	}

	branch, err := u.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if branch == nil {
		return nil, nil
	}

	if redisclient.Client != nil {
		if bytes, err := json.Marshal(branch); err == nil {
			_ = redisclient.Client.Set(ctx, cacheKey, bytes, 2*time.Hour).Err()
		}
	}

	return branch, nil
}

func invalidateBranchCache(ctx context.Context, id uuid.UUID) {
	if redisclient.Client == nil {
		return
	}
	_ = redisclient.Client.Del(ctx, fmt.Sprintf("branch:id:%s", id.String())).Err()
}

func (u *branchUseCase) UpdateBranch(id uuid.UUID, req *UpdateBranchRequest) error {
	branch, err := u.repo.FindByID(id)
	if err != nil {
		return err
	}
	if branch == nil {
		return errors.New("branch not found")
	}

	// Update CompanyID if provided
	if req.CompanyID != "" {
		parsedID, err := uuid.Parse(req.CompanyID)
		if err != nil {
			return err
		}
		branch.CompanyID = &parsedID
	} else {
		branch.CompanyID = nil
	}
	branch.Name = req.Name
	branch.Address = req.Address
	branch.Timezone = req.Timezone
	branch.Latitude = req.Latitude
	branch.Longitude = req.Longitude
	branch.RadiusMeter = req.RadiusMeter

	if err := u.repo.Update(branch); err != nil {
		return err
	}
	invalidateBranchCache(context.Background(), id)
	return nil
}

func (u *branchUseCase) DeleteBranch(id uuid.UUID) error {
	if err := u.repo.Delete(id); err != nil {
		return err
	}
	invalidateBranchCache(context.Background(), id)
	return nil
}

func (u *branchUseCase) SetProductPrice(branchID, productID uuid.UUID, het float64) error {
	return u.repo.SetProductPrice(branchID, productID, het)
}

func (u *branchUseCase) GetProductPrices(branchID uuid.UUID) ([]domain.BranchProductPrice, error) {
	return u.repo.GetProductPrices(branchID)
}
