package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"example.com/hris/backend/internal/domain"
)

type BonusInventoryRepository interface {
	Create(ctx context.Context, item *domain.BonusInventory) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.BonusInventory, error)
	GetByBranchID(ctx context.Context, branchID uuid.UUID) ([]domain.BonusInventory, error)
	GetAll(ctx context.Context) ([]domain.BonusInventory, error)
	Update(ctx context.Context, item *domain.BonusInventory) error
	Delete(ctx context.Context, id uuid.UUID) error
	CreateLog(ctx context.Context, log *domain.BonusInventoryLog) error
	GetLogsByInventoryID(ctx context.Context, inventoryID uuid.UUID) ([]domain.BonusInventoryLog, error)
	WithTrx(trxHandle *gorm.DB) BonusInventoryRepository
}

type bonusInventoryRepository struct {
	db *gorm.DB
}

func NewBonusInventoryRepository(db *gorm.DB) BonusInventoryRepository {
	return &bonusInventoryRepository{
		db: db,
	}
}

func (r *bonusInventoryRepository) WithTrx(trxHandle *gorm.DB) BonusInventoryRepository {
	if trxHandle == nil {
		return r
	}
	return &bonusInventoryRepository{
		db: trxHandle,
	}
}

func (r *bonusInventoryRepository) Create(ctx context.Context, item *domain.BonusInventory) error {
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *bonusInventoryRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.BonusInventory, error) {
	var item domain.BonusInventory
	if err := r.db.WithContext(ctx).Preload("Branch").First(&item, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *bonusInventoryRepository) GetByBranchID(ctx context.Context, branchID uuid.UUID) ([]domain.BonusInventory, error) {
	var items []domain.BonusInventory
	if err := r.db.WithContext(ctx).Preload("Branch").Where("branch_id = ?", branchID).Order("created_at desc").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (r *bonusInventoryRepository) GetAll(ctx context.Context) ([]domain.BonusInventory, error) {
	var items []domain.BonusInventory
	if err := r.db.WithContext(ctx).Preload("Branch").Order("created_at desc").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (r *bonusInventoryRepository) Update(ctx context.Context, item *domain.BonusInventory) error {
	return r.db.WithContext(ctx).Save(item).Error
}

func (r *bonusInventoryRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.BonusInventory{}, "id = ?", id).Error
}

func (r *bonusInventoryRepository) CreateLog(ctx context.Context, log *domain.BonusInventoryLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *bonusInventoryRepository) GetLogsByInventoryID(ctx context.Context, inventoryID uuid.UUID) ([]domain.BonusInventoryLog, error) {
	var logs []domain.BonusInventoryLog
	if err := r.db.WithContext(ctx).Where("bonus_inventory_id = ?", inventoryID).Order("created_at desc").Find(&logs).Error; err != nil {
		return nil, err
	}
	return logs, nil
}
