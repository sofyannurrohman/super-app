package repository

import (
	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"gorm.io/gorm"
)

type BannerOrderRepository interface {
	Create(order *domain.BannerOrder) error
	FindAll() ([]domain.BannerOrder, error)
	FindByID(id uuid.UUID) (*domain.BannerOrder, error)
	Update(order *domain.BannerOrder) error
	Delete(id uuid.UUID) error
	FindByCompanyID(companyID uuid.UUID) ([]domain.BannerOrder, error)
}

type bannerOrderRepository struct {
	db *gorm.DB
}

func NewBannerOrderRepository(db *gorm.DB) BannerOrderRepository {
	return &bannerOrderRepository{db: db}
}

func (r *bannerOrderRepository) Create(order *domain.BannerOrder) error {
	return r.db.Create(order).Error
}

func (r *bannerOrderRepository) FindAll() ([]domain.BannerOrder, error) {
	var orders []domain.BannerOrder
	err := r.db.Preload("Employee").Preload("Designer").Preload("Installer").Find(&orders).Error
	return orders, err
}

func (r *bannerOrderRepository) FindByID(id uuid.UUID) (*domain.BannerOrder, error) {
	var order domain.BannerOrder
	err := r.db.Preload("Employee").Preload("Designer").Preload("Installer").First(&order, "id = ?", id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &order, nil
}

func (r *bannerOrderRepository) Update(order *domain.BannerOrder) error {
	return r.db.Save(order).Error
}

func (r *bannerOrderRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&domain.BannerOrder{}, "id = ?", id).Error
}

func (r *bannerOrderRepository) FindByCompanyID(companyID uuid.UUID) ([]domain.BannerOrder, error) {
	var orders []domain.BannerOrder
	err := r.db.Preload("Employee").Preload("Designer").Preload("Installer").Where("company_id = ?", companyID).Find(&orders).Error
	return orders, err
}
