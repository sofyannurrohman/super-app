package repository

import (
	"context"

	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"gorm.io/gorm"
)

type SalesBonusRepository interface {
	CreateRequest(ctx context.Context, req *domain.SalesBonusRequest) error
	GetRequestByID(ctx context.Context, id uuid.UUID) (*domain.SalesBonusRequest, error)
	GetPendingRequests(ctx context.Context, branchID *uuid.UUID) ([]domain.SalesBonusRequest, error)
	UpdateRequest(ctx context.Context, req *domain.SalesBonusRequest) error
	GetSalesBonusStock(ctx context.Context, employeeID uuid.UUID) ([]domain.SalesBonusStock, error)
	GetSalesBonusStockItem(ctx context.Context, employeeID uuid.UUID, bonusID uuid.UUID) (*domain.SalesBonusStock, error)
	CreateSalesBonusStock(ctx context.Context, stock *domain.SalesBonusStock) error
	UpdateSalesBonusStock(ctx context.Context, stock *domain.SalesBonusStock) error
	WithTrx(tx *gorm.DB) SalesBonusRepository
}

type salesBonusRepository struct {
	db *gorm.DB
}

func NewSalesBonusRepository(db *gorm.DB) SalesBonusRepository {
	return &salesBonusRepository{db: db}
}

func (r *salesBonusRepository) WithTrx(tx *gorm.DB) SalesBonusRepository {
	return &salesBonusRepository{db: tx}
}

func (r *salesBonusRepository) CreateRequest(ctx context.Context, req *domain.SalesBonusRequest) error {
	return r.db.WithContext(ctx).Create(req).Error
}

func (r *salesBonusRepository) GetRequestByID(ctx context.Context, id uuid.UUID) (*domain.SalesBonusRequest, error) {
	var req domain.SalesBonusRequest
	err := r.db.WithContext(ctx).Preload("Employee").Preload("BonusInventory").First(&req, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &req, nil
}

func (r *salesBonusRepository) GetPendingRequests(ctx context.Context, branchID *uuid.UUID) ([]domain.SalesBonusRequest, error) {
	var reqs []domain.SalesBonusRequest
	query := r.db.WithContext(ctx).Preload("Employee").Preload("Employee.Branch").Preload("Employee.OperatingBranch").Preload("BonusInventory").Where("sales_bonus_requests.status = ?", domain.SalesBonusRequestPending)

	if branchID != nil {
		query = query.Joins("JOIN employees ON employees.id = sales_bonus_requests.employee_id").
			Where("COALESCE(employees.operating_branch_id, employees.branch_id) = ?", branchID)
	}

	err := query.Order("sales_bonus_requests.created_at desc").Find(&reqs).Error
	return reqs, err
}

func (r *salesBonusRepository) UpdateRequest(ctx context.Context, req *domain.SalesBonusRequest) error {
	return r.db.WithContext(ctx).Save(req).Error
}

func (r *salesBonusRepository) GetSalesBonusStock(ctx context.Context, employeeID uuid.UUID) ([]domain.SalesBonusStock, error) {
	var stocks []domain.SalesBonusStock
	err := r.db.WithContext(ctx).Preload("BonusInventory").Where("employee_id = ?", employeeID).Find(&stocks).Error
	return stocks, err
}

func (r *salesBonusRepository) GetSalesBonusStockItem(ctx context.Context, employeeID uuid.UUID, bonusID uuid.UUID) (*domain.SalesBonusStock, error) {
	var stock domain.SalesBonusStock
	err := r.db.WithContext(ctx).Where("employee_id = ? AND bonus_inventory_id = ?", employeeID, bonusID).First(&stock).Error
	if err != nil {
		return nil, err
	}
	return &stock, nil
}

func (r *salesBonusRepository) CreateSalesBonusStock(ctx context.Context, stock *domain.SalesBonusStock) error {
	return r.db.WithContext(ctx).Create(stock).Error
}

func (r *salesBonusRepository) UpdateSalesBonusStock(ctx context.Context, stock *domain.SalesBonusStock) error {
	return r.db.WithContext(ctx).Save(stock).Error
}
