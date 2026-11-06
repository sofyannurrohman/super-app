package repository

import (
	"time"

	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"gorm.io/gorm"
)

type MarketplaceInventoryRepository interface {
	CreateInventory(inventory *domain.MarketplaceInventory) error
	GetInventoryByID(id uuid.UUID) (*domain.MarketplaceInventory, error)
	GetAllInventories(branchID uuid.UUID) ([]domain.MarketplaceInventory, error)
	UpdateInventory(inventory *domain.MarketplaceInventory) error
	DeleteInventory(id uuid.UUID) error

	CreateLog(log *domain.MarketplaceInventoryLog) error
	GetLogs(inventoryID *uuid.UUID, from, to *time.Time) ([]domain.MarketplaceInventoryLog, error)

	CreateOpname(opname *domain.MarketplaceInventoryOpname) error
	GetOpnames(inventoryID *uuid.UUID, from, to *time.Time, limit, offset int) ([]domain.MarketplaceInventoryOpname, int64, error)
}

type marketplaceInventoryRepository struct {
	db *gorm.DB
}

func NewMarketplaceInventoryRepository(db *gorm.DB) MarketplaceInventoryRepository {
	return &marketplaceInventoryRepository{db}
}

func (r *marketplaceInventoryRepository) CreateInventory(inventory *domain.MarketplaceInventory) error {
	return r.db.Create(inventory).Error
}

func (r *marketplaceInventoryRepository) GetInventoryByID(id uuid.UUID) (*domain.MarketplaceInventory, error) {
	var inventory domain.MarketplaceInventory
	err := r.db.First(&inventory, "id = ?", id).Error
	return &inventory, err
}

func (r *marketplaceInventoryRepository) GetAllInventories(branchID uuid.UUID) ([]domain.MarketplaceInventory, error) {
	var inventories []domain.MarketplaceInventory
	q := r.db.Order("name asc")
	if branchID != uuid.Nil {
		q = q.Where("branch_id = ?", branchID)
	}
	err := q.Find(&inventories).Error
	return inventories, err
}

func (r *marketplaceInventoryRepository) UpdateInventory(inventory *domain.MarketplaceInventory) error {
	return r.db.Save(inventory).Error
}

func (r *marketplaceInventoryRepository) DeleteInventory(id uuid.UUID) error {
	return r.db.Delete(&domain.MarketplaceInventory{}, "id = ?", id).Error
}

func (r *marketplaceInventoryRepository) CreateLog(log *domain.MarketplaceInventoryLog) error {
	return r.db.Create(log).Error
}

func (r *marketplaceInventoryRepository) GetLogs(inventoryID *uuid.UUID, from, to *time.Time) ([]domain.MarketplaceInventoryLog, error) {
	var logs []domain.MarketplaceInventoryLog
	q := r.db.Preload("MarketplaceInventory").Order("created_at desc")
	if inventoryID != nil {
		q = q.Where("marketplace_inventory_id = ?", *inventoryID)
	}
	if from != nil {
		q = q.Where("created_at >= ?", *from)
	}
	if to != nil {
		q = q.Where("created_at <= ?", *to)
	}
	err := q.Find(&logs).Error
	return logs, err
}

func (r *marketplaceInventoryRepository) CreateOpname(opname *domain.MarketplaceInventoryOpname) error {
	return r.db.Create(opname).Error
}

func (r *marketplaceInventoryRepository) GetOpnames(inventoryID *uuid.UUID, from, to *time.Time, limit, offset int) ([]domain.MarketplaceInventoryOpname, int64, error) {
	var opnames []domain.MarketplaceInventoryOpname
	var total int64
	q := r.db.Model(&domain.MarketplaceInventoryOpname{})
	if inventoryID != nil {
		q = q.Where("marketplace_inventory_id = ?", *inventoryID)
	}
	if from != nil {
		q = q.Where("created_at >= ?", *from)
	}
	if to != nil {
		q = q.Where("created_at <= ?", *to)
	}
	err := q.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	err = q.Preload("MarketplaceInventory").Order("created_at desc").Limit(limit).Offset(offset).Find(&opnames).Error
	return opnames, total, err
}
