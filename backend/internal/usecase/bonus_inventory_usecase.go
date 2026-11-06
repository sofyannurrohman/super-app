package usecase

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"example.com/hris/backend/internal/domain"
	"example.com/hris/backend/internal/repository"
)

type BonusInventoryUsecase interface {
	Create(ctx context.Context, item *domain.BonusInventory) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.BonusInventory, error)
	GetByBranchID(ctx context.Context, branchID uuid.UUID) ([]domain.BonusInventory, error)
	GetAll(ctx context.Context) ([]domain.BonusInventory, error)
	Update(ctx context.Context, item *domain.BonusInventory) error
	Delete(ctx context.Context, id uuid.UUID) error
	AdjustStock(ctx context.Context, inventoryID uuid.UUID, newQuantity int, user string, notes string) error
	GetLogs(ctx context.Context, inventoryID uuid.UUID) ([]domain.BonusInventoryLog, error)
}

type bonusInventoryUsecase struct {
	repo repository.BonusInventoryRepository
	db   *gorm.DB
}

func NewBonusInventoryUsecase(repo repository.BonusInventoryRepository, db *gorm.DB) BonusInventoryUsecase {
	return &bonusInventoryUsecase{
		repo: repo,
		db:   db,
	}
}

func (u *bonusInventoryUsecase) Create(ctx context.Context, item *domain.BonusInventory) error {
	return u.db.Transaction(func(tx *gorm.DB) error {
		txRepo := u.repo.WithTrx(tx)
		if err := txRepo.Create(ctx, item); err != nil {
			return err
		}

		if item.Quantity > 0 {
			log := &domain.BonusInventoryLog{
				BonusInventoryID: item.ID,
				Type:             domain.BonusInventoryIn,
				Quantity:         item.Quantity,
				QuantityBefore:   0,
				QuantityAfter:    item.Quantity,
				Reference:        "INITIAL_STOCK",
				Notes:            "Initial stock creation",
				CreatedBy:        "SYSTEM", // We can enhance this to take user context if needed
			}
			if err := txRepo.CreateLog(ctx, log); err != nil {
				return err
			}
		}

		return nil
	})
}

func (u *bonusInventoryUsecase) GetByID(ctx context.Context, id uuid.UUID) (*domain.BonusInventory, error) {
	return u.repo.GetByID(ctx, id)
}

func (u *bonusInventoryUsecase) GetByBranchID(ctx context.Context, branchID uuid.UUID) ([]domain.BonusInventory, error) {
	return u.repo.GetByBranchID(ctx, branchID)
}

func (u *bonusInventoryUsecase) GetAll(ctx context.Context) ([]domain.BonusInventory, error) {
	return u.repo.GetAll(ctx)
}

func (u *bonusInventoryUsecase) Update(ctx context.Context, item *domain.BonusInventory) error {
	// For stock updates, we should use AdjustStock instead.
	// Update here is just for Name, Unit, Description
	existing, err := u.repo.GetByID(ctx, item.ID)
	if err != nil {
		return err
	}

	existing.Name = item.Name
	existing.Unit = item.Unit
	existing.Description = item.Description

	return u.repo.Update(ctx, existing)
}

func (u *bonusInventoryUsecase) Delete(ctx context.Context, id uuid.UUID) error {
	return u.repo.Delete(ctx, id)
}

func (u *bonusInventoryUsecase) AdjustStock(ctx context.Context, inventoryID uuid.UUID, newQuantity int, user string, notes string) error {
	return u.db.Transaction(func(tx *gorm.DB) error {
		txRepo := u.repo.WithTrx(tx)

		item, err := txRepo.GetByID(ctx, inventoryID)
		if err != nil {
			return err
		}

		if item.Quantity == newQuantity {
			return errors.New("new quantity is the same as current quantity")
		}

		oldQuantity := item.Quantity
		diff := newQuantity - oldQuantity
		adjType := domain.BonusInventoryAdj
		if diff > 0 {
			adjType = domain.BonusInventoryIn
		} else {
			adjType = domain.BonusInventoryOut
		}

		// Update quantity
		item.Quantity = newQuantity
		if err := txRepo.Update(ctx, item); err != nil {
			return err
		}

		// Create log
		log := &domain.BonusInventoryLog{
			BonusInventoryID: item.ID,
			Type:             adjType,
			Quantity:         diff,
			QuantityBefore:   oldQuantity,
			QuantityAfter:    newQuantity,
			Reference:        "MANUAL_ADJUSTMENT",
			Notes:            notes,
			CreatedBy:        user,
		}

		if err := txRepo.CreateLog(ctx, log); err != nil {
			return err
		}

		return nil
	})
}

func (u *bonusInventoryUsecase) GetLogs(ctx context.Context, inventoryID uuid.UUID) ([]domain.BonusInventoryLog, error) {
	return u.repo.GetLogsByInventoryID(ctx, inventoryID)
}
