package repository

import (
	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"gorm.io/gorm"
)

type BranchRepository interface {
	Create(branch *domain.Branch) error
	FindAll() ([]domain.Branch, error)
	FindByCompanyID(companyID uuid.UUID) ([]domain.Branch, error)
	FindByID(id uuid.UUID) (*domain.Branch, error)
	Update(branch *domain.Branch) error
	Delete(id uuid.UUID) error
	SetProductPrice(branchID, productID uuid.UUID, het float64) error
	GetProductPrices(branchID uuid.UUID) ([]domain.BranchProductPrice, error)
}

type branchRepository struct {
	db *gorm.DB
}

func NewBranchRepository(db *gorm.DB) BranchRepository {
	return &branchRepository{db}
}

func (r *branchRepository) Create(branch *domain.Branch) error {
	return r.db.Create(branch).Error
}

func (r *branchRepository) FindAll() ([]domain.Branch, error) {
	var branches []domain.Branch
	if err := r.db.Preload("Company").Find(&branches).Error; err != nil {
		return nil, err
	}
	return branches, nil
}

func (r *branchRepository) FindByCompanyID(companyID uuid.UUID) ([]domain.Branch, error) {
	var branches []domain.Branch
	if err := r.db.Preload("Company").Where("company_id = ?", companyID).Find(&branches).Error; err != nil {
		return nil, err
	}
	return branches, nil
}

func (r *branchRepository) FindByID(id uuid.UUID) (*domain.Branch, error) {
	var branch domain.Branch
	if err := r.db.Preload("Company").Where("id = ?", id).First(&branch).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &branch, nil
}

func (r *branchRepository) Update(branch *domain.Branch) error {
	return r.db.Omit("Company").Save(branch).Error
}

func (r *branchRepository) Delete(id uuid.UUID) error {
	return r.db.Where("id = ?", id).Delete(&domain.Branch{}).Error
}

func (r *branchRepository) SetProductPrice(branchID, productID uuid.UUID, het float64) error {
	var price domain.BranchProductPrice
	err := r.db.Where("branch_id = ? AND product_id = ?", branchID, productID).First(&price).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}
	if err == gorm.ErrRecordNotFound {
		price = domain.BranchProductPrice{
			BranchID:  branchID,
			ProductID: productID,
			HET:       het,
		}
		return r.db.Create(&price).Error
	}
	price.HET = het
	return r.db.Save(&price).Error
}

func (r *branchRepository) GetProductPrices(branchID uuid.UUID) ([]domain.BranchProductPrice, error) {
	var prices []domain.BranchProductPrice
	err := r.db.Where("branch_id = ?", branchID).Preload("Product").Find(&prices).Error
	return prices, err
}
