package repository

import (
	"strings"

	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"gorm.io/gorm"
)

type TechnicianInventoryRepository interface {
	ListItems(companyID, factoryID uuid.UUID, category string, lowStockOnly bool, search string) ([]domain.TechnicianItem, error)
	GetItemByID(id uuid.UUID) (*domain.TechnicianItem, error)
	GetItemByCode(factoryID uuid.UUID, code string) (*domain.TechnicianItem, error)
	CreateItem(item *domain.TechnicianItem) error
	UpdateItem(item *domain.TechnicianItem) error
	DeleteItem(id uuid.UUID) error
	IncrementStock(id uuid.UUID, qty int) error
	CreateStockLog(log *domain.TechnicianStockLog) error
	ListStockLogs(companyID, factoryID uuid.UUID, itemID *uuid.UUID, logType string) ([]domain.TechnicianStockLog, error)
}

type technicianInventoryRepository struct {
	db *gorm.DB
}

func NewTechnicianInventoryRepository(db *gorm.DB) TechnicianInventoryRepository {
	return &technicianInventoryRepository{db: db}
}

func (r *technicianInventoryRepository) ListItems(companyID, factoryID uuid.UUID, category string, lowStockOnly bool, search string) ([]domain.TechnicianItem, error) {
	var items []domain.TechnicianItem
	query := r.db.Model(&domain.TechnicianItem{}).Preload("Factory")

	if factoryID != uuid.Nil {
		query = query.Where("factory_id = ?", factoryID)
	} else if companyID != uuid.Nil {
		query = query.Where("company_id = ?", companyID)
	}

	if category != "" && category != "ALL" {
		query = query.Where("category = ?", category)
	}

	if lowStockOnly {
		query = query.Where("quantity <= min_limit")
	}

	if strings.TrimSpace(search) != "" {
		q := "%" + strings.ToLower(strings.TrimSpace(search)) + "%"
		query = query.Where("LOWER(name) LIKE ? OR LOWER(code) LIKE ? OR LOWER(rack_location) LIKE ?", q, q, q)
	}

	err := query.Order("name ASC").Find(&items).Error
	return items, err
}

func (r *technicianInventoryRepository) GetItemByID(id uuid.UUID) (*domain.TechnicianItem, error) {
	var item domain.TechnicianItem
	err := r.db.Preload("Factory").First(&item, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *technicianInventoryRepository) GetItemByCode(factoryID uuid.UUID, code string) (*domain.TechnicianItem, error) {
	var item domain.TechnicianItem
	err := r.db.First(&item, "factory_id = ? AND LOWER(code) = LOWER(?)", factoryID, strings.TrimSpace(code)).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *technicianInventoryRepository) CreateItem(item *domain.TechnicianItem) error {
	return r.db.Create(item).Error
}

func (r *technicianInventoryRepository) UpdateItem(item *domain.TechnicianItem) error {
	return r.db.Save(item).Error
}

func (r *technicianInventoryRepository) DeleteItem(id uuid.UUID) error {
	return r.db.Delete(&domain.TechnicianItem{}, "id = ?", id).Error
}

func (r *technicianInventoryRepository) IncrementStock(id uuid.UUID, qty int) error {
	return r.db.Model(&domain.TechnicianItem{}).
		Where("id = ?", id).
		UpdateColumn("quantity", gorm.Expr("quantity + ?", qty)).Error
}

func (r *technicianInventoryRepository) CreateStockLog(log *domain.TechnicianStockLog) error {
	return r.db.Create(log).Error
}

func (r *technicianInventoryRepository) ListStockLogs(companyID, factoryID uuid.UUID, itemID *uuid.UUID, logType string) ([]domain.TechnicianStockLog, error) {
	var logs []domain.TechnicianStockLog
	query := r.db.Model(&domain.TechnicianStockLog{}).Preload("TechnicianItem").Preload("Factory")

	if factoryID != uuid.Nil {
		query = query.Where("factory_id = ?", factoryID)
	} else if companyID != uuid.Nil {
		query = query.Where("company_id = ?", companyID)
	}

	if itemID != nil && *itemID != uuid.Nil {
		query = query.Where("technician_item_id = ?", *itemID)
	}

	if logType != "" && logType != "ALL" {
		query = query.Where("type = ?", logType)
	}

	err := query.Order("created_at DESC").Limit(300).Find(&logs).Error
	return logs, err
}
