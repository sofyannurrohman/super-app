package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"example.com/hris/backend/internal/repository"
	"gorm.io/gorm"
)

type SalesBonusUsecase interface {
	CreateRequest(ctx context.Context, employeeID, bonusInventoryID uuid.UUID, quantity int, notes string) error
	ApproveRequest(ctx context.Context, requestID uuid.UUID) error
	RejectRequest(ctx context.Context, requestID uuid.UUID) error
	GetPendingRequests(ctx context.Context, branchID string) ([]domain.SalesBonusRequest, error)
	GetSalesBonusStock(ctx context.Context, employeeID uuid.UUID) ([]domain.SalesBonusStock, error)
}

type salesBonusUsecase struct {
	repo      repository.SalesBonusRepository
	bonusRepo repository.BonusInventoryRepository
	db        *gorm.DB
}

func NewSalesBonusUsecase(repo repository.SalesBonusRepository, bonusRepo repository.BonusInventoryRepository, db *gorm.DB) SalesBonusUsecase {
	return &salesBonusUsecase{
		repo:      repo,
		bonusRepo: bonusRepo,
		db:        db,
	}
}

func (u *salesBonusUsecase) CreateRequest(ctx context.Context, employeeID, bonusInventoryID uuid.UUID, quantity int, notes string) error {
	if quantity <= 0 {
		return errors.New("quantity must be greater than 0")
	}

	req := &domain.SalesBonusRequest{
		EmployeeID:       employeeID,
		BonusInventoryID: bonusInventoryID,
		Quantity:         quantity,
		Notes:            notes,
		Status:           domain.SalesBonusRequestPending,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	return u.repo.CreateRequest(ctx, req)
}

func (u *salesBonusUsecase) ApproveRequest(ctx context.Context, requestID uuid.UUID) error {
	return u.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txRepo := u.repo.WithTrx(tx)
		txBonusRepo := repository.NewBonusInventoryRepository(tx)

		req, err := txRepo.GetRequestByID(ctx, requestID)
		if err != nil {
			return err
		}

		if req.Status != domain.SalesBonusRequestPending {
			return errors.New("request is not pending")
		}

		// Check warehouse stock
		inv, err := txBonusRepo.GetByID(ctx, req.BonusInventoryID)
		if err != nil {
			return err
		}

		if inv.Quantity < req.Quantity {
			return errors.New("insufficient bonus inventory stock in warehouse")
		}

		// Deduct from warehouse
		inv.Quantity -= req.Quantity
		if err := txBonusRepo.Update(ctx, inv); err != nil {
			return err
		}

		// Create Log
		log := &domain.BonusInventoryLog{
			BonusInventoryID: inv.ID,
			Type:             domain.BonusInventoryOut,
			Quantity:         req.Quantity,
			Notes:            fmt.Sprintf("Approved Sales Request from Employee %s", req.EmployeeID),
		}
		if err := txBonusRepo.CreateLog(ctx, log); err != nil {
			return err
		}

		// Update Request Status
		req.Status = domain.SalesBonusRequestApproved
		req.UpdatedAt = time.Now()
		if err := txRepo.UpdateRequest(ctx, req); err != nil {
			return err
		}

		// Increment Sales Bonus Stock
		stock, err := txRepo.GetSalesBonusStockItem(ctx, req.EmployeeID, req.BonusInventoryID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				stock = &domain.SalesBonusStock{
					EmployeeID:       req.EmployeeID,
					BonusInventoryID: req.BonusInventoryID,
					Quantity:         req.Quantity,
					UpdatedAt:        time.Now(),
				}
				if err := txRepo.CreateSalesBonusStock(ctx, stock); err != nil {
					return err
				}
			} else {
				return err
			}
		} else {
			stock.Quantity += req.Quantity
			stock.UpdatedAt = time.Now()
			if err := txRepo.UpdateSalesBonusStock(ctx, stock); err != nil {
				return err
			}
		}

		return nil
	})
}

func (u *salesBonusUsecase) RejectRequest(ctx context.Context, requestID uuid.UUID) error {
	req, err := u.repo.GetRequestByID(ctx, requestID)
	if err != nil {
		return err
	}

	if req.Status != domain.SalesBonusRequestPending {
		return errors.New("request is not pending")
	}

	req.Status = domain.SalesBonusRequestRejected
	req.UpdatedAt = time.Now()

	return u.repo.UpdateRequest(ctx, req)
}

func (u *salesBonusUsecase) GetPendingRequests(ctx context.Context, branchID string) ([]domain.SalesBonusRequest, error) {
	var parsedBranchID *uuid.UUID
	if branchID != "" {
		id, err := uuid.Parse(branchID)
		if err == nil {
			parsedBranchID = &id
		}
	}
	return u.repo.GetPendingRequests(ctx, parsedBranchID)
}

func (u *salesBonusUsecase) GetSalesBonusStock(ctx context.Context, employeeID uuid.UUID) ([]domain.SalesBonusStock, error) {
	return u.repo.GetSalesBonusStock(ctx, employeeID)
}
