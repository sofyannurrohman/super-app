package repository

import (
	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"gorm.io/gorm"
)

type FactoryReturnRepository interface {
	Create(factoryReturn *domain.FactoryReturn) error
	GetByID(id uuid.UUID) (*domain.FactoryReturn, error)
	GetByBranch(branchID uuid.UUID) ([]domain.FactoryReturn, error)
	GetByFactory(factoryID uuid.UUID) ([]domain.FactoryReturn, error)
	GetByCompany(companyID uuid.UUID) ([]domain.FactoryReturn, error)
	Update(factoryReturn *domain.FactoryReturn) error
}

type factoryReturnRepository struct {
	db *gorm.DB
}

func NewFactoryReturnRepository(db *gorm.DB) FactoryReturnRepository {
	return &factoryReturnRepository{db}
}

func (r *factoryReturnRepository) Create(factoryReturn *domain.FactoryReturn) error {
	return r.db.Create(factoryReturn).Error
}

func (r *factoryReturnRepository) GetByID(id uuid.UUID) (*domain.FactoryReturn, error) {
	var factoryReturn domain.FactoryReturn
	err := r.db.Preload("Items.Product").
		Preload("Items.FactoryWarehouse").
		Preload("FromBranch").
		Preload("ToFactory").
		Preload("Vehicle").
		Preload("Driver").
		Preload("ApprovedBy").
		Preload("ApprovedBy.JobPosition").
		First(&factoryReturn, "id = ?", id).Error
	return &factoryReturn, err
}

func (r *factoryReturnRepository) GetByBranch(branchID uuid.UUID) ([]domain.FactoryReturn, error) {
	var returns []domain.FactoryReturn
	err := r.db.Preload("Items.Product").
		Preload("Items.FactoryWarehouse").
		Preload("ToFactory").
		Preload("Vehicle").
		Preload("Driver").
		Preload("ApprovedBy").
		Preload("ApprovedBy.JobPosition").
		Where("from_branch_id = ?", branchID).
		Order("created_at desc").
		Find(&returns).Error
	return returns, err
}

func (r *factoryReturnRepository) GetByFactory(factoryID uuid.UUID) ([]domain.FactoryReturn, error) {
	var returns []domain.FactoryReturn
	err := r.db.Preload("Items.Product").
		Preload("Items.FactoryWarehouse").
		Preload("FromBranch").
		Preload("Vehicle").
		Preload("Driver").
		Preload("ApprovedBy").
		Preload("ApprovedBy.JobPosition").
		Where("to_factory_id = ?", factoryID).
		Order("created_at desc").
		Find(&returns).Error
	return returns, err
}

func (r *factoryReturnRepository) GetByCompany(companyID uuid.UUID) ([]domain.FactoryReturn, error) {
	var returns []domain.FactoryReturn
	err := r.db.Select("factory_returns.*").
		Preload("Items.Product").
		Preload("Items.FactoryWarehouse").
		Preload("FromBranch").
		Preload("ToFactory").
		Preload("Vehicle").
		Preload("Driver").
		Preload("ApprovedBy").
		Preload("ApprovedBy.JobPosition").
		Joins("JOIN factories ON factories.id = factory_returns.to_factory_id").
		Where("factories.company_id = ?", companyID).
		Order("factory_returns.created_at desc").
		Find(&returns).Error
	return returns, err
}

func (r *factoryReturnRepository) Update(factoryReturn *domain.FactoryReturn) error {
	return r.db.Save(factoryReturn).Error
}
