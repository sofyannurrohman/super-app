package usecase

import (
	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"example.com/hris/backend/internal/repository"
)

type BannerOrderUseCase interface {
	CreateOrder(order *domain.BannerOrder) error
	GetAllOrders() ([]domain.BannerOrder, error)
	UpdateOrder(order *domain.BannerOrder) error
	UpdateStatus(id uuid.UUID, status string) error
	DeleteOrder(id uuid.UUID) error
	GetOrdersByCompanyID(companyID uuid.UUID) ([]domain.BannerOrder, error)
}

type bannerOrderUseCase struct {
	bannerOrderRepo repository.BannerOrderRepository
}

func NewBannerOrderUseCase(bannerOrderRepo repository.BannerOrderRepository) BannerOrderUseCase {
	return &bannerOrderUseCase{
		bannerOrderRepo: bannerOrderRepo,
	}
}

func (u *bannerOrderUseCase) CreateOrder(order *domain.BannerOrder) error {
	order.Status = "PENDING"
	return u.bannerOrderRepo.Create(order)
}

func (u *bannerOrderUseCase) GetAllOrders() ([]domain.BannerOrder, error) {
	return u.bannerOrderRepo.FindAll()
}

func (u *bannerOrderUseCase) UpdateOrder(order *domain.BannerOrder) error {
	return u.bannerOrderRepo.Update(order)
}

func (u *bannerOrderUseCase) UpdateStatus(id uuid.UUID, status string) error {
	order, err := u.bannerOrderRepo.FindByID(id)
	if err != nil {
		return err
	}
	if order == nil {
		return nil
	}

	order.Status = status
	return u.bannerOrderRepo.Update(order)
}

func (u *bannerOrderUseCase) DeleteOrder(id uuid.UUID) error {
	return u.bannerOrderRepo.Delete(id)
}

func (u *bannerOrderUseCase) GetOrdersByCompanyID(companyID uuid.UUID) ([]domain.BannerOrder, error) {
	return u.bannerOrderRepo.FindByCompanyID(companyID)
}
