package usecase

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"example.com/hris/backend/internal/repository"
	"gorm.io/gorm"
)

type SalesTesterRequestUsecase interface {
	CreateRequest(req *domain.SalesTesterRequest) error
	GetRequestByID(id uuid.UUID) (*domain.SalesTesterRequest, error)
	ListRequests(branchID *uuid.UUID, status string) ([]domain.SalesTesterRequest, error)
	GetEmployeeRequests(employeeID uuid.UUID) ([]domain.SalesTesterRequest, error)
	ApproveRequest(id uuid.UUID) error
	RejectRequest(id uuid.UUID, notes string) error
	IssueDirectly(id uuid.UUID) error
	DeleteRequest(id uuid.UUID) error
}

type salesTesterRequestUsecase struct {
	repo repository.SalesTesterRequestRepository
	db   *gorm.DB
}

func NewSalesTesterRequestUsecase(repo repository.SalesTesterRequestRepository, db *gorm.DB) SalesTesterRequestUsecase {
	return &salesTesterRequestUsecase{repo, db}
}

func (u *salesTesterRequestUsecase) CreateRequest(req *domain.SalesTesterRequest) error {
	req.Status = domain.TesterStatusPending
	return u.repo.Create(req)
}

func (u *salesTesterRequestUsecase) GetRequestByID(id uuid.UUID) (*domain.SalesTesterRequest, error) {
	return u.repo.FindByID(id)
}

func (u *salesTesterRequestUsecase) ListRequests(branchID *uuid.UUID, status string) ([]domain.SalesTesterRequest, error) {
	return u.repo.FindAll(branchID, status)
}

func (u *salesTesterRequestUsecase) GetEmployeeRequests(employeeID uuid.UUID) ([]domain.SalesTesterRequest, error) {
	return u.repo.FindByEmployee(employeeID)
}

func (u *salesTesterRequestUsecase) ApproveRequest(id uuid.UUID) error {
	req, err := u.repo.FindByID(id)
	if err != nil {
		return err
	}
	if req.Status != domain.TesterStatusPending {
		return errors.New("request status must be PENDING to be approved")
	}
	req.Status = domain.TesterStatusApproved
	return u.repo.Update(req)
}

func (u *salesTesterRequestUsecase) RejectRequest(id uuid.UUID, notes string) error {
	req, err := u.repo.FindByID(id)
	if err != nil {
		return err
	}
	if req.Status != domain.TesterStatusPending {
		return errors.New("request status must be PENDING to be rejected")
	}
	req.Status = domain.TesterStatusRejected
	if notes != "" {
		req.Notes = req.Notes + "\nRejection Note: " + notes
	}
	return u.repo.Update(req)
}

func (u *salesTesterRequestUsecase) IssueDirectly(id uuid.UUID) error {
	return u.db.Transaction(func(tx *gorm.DB) error {
		reqRepo := repository.NewSalesTesterRequestRepository(tx)
		req, err := reqRepo.FindByID(id)
		if err != nil {
			return err
		}
		if req.Status != domain.TesterStatusApproved && req.Status != domain.TesterStatusPending {
			return errors.New("request status must be PENDING or APPROVED to be issued directly")
		}

		// Deduct stock from warehouse
		wRepo := repository.NewWarehouseRepository(tx)
		stock, err := wRepo.GetStock(req.BranchID, req.ProductID, "GOOD", "DEFAULT")
		if err != nil || stock == nil {
			return errors.New("insufficient stock in main warehouse")
		}
		if stock.Quantity < req.Quantity {
			return fmt.Errorf("insufficient stock: requested %d, available %d", req.Quantity, stock.Quantity)
		}

		stock.Quantity -= req.Quantity
		if err := wRepo.UpdateStock(stock); err != nil {
			return err
		}

		// Log warehouse mutation
		wRepo.CreateLog(&domain.WarehouseLog{
			BranchID:  req.BranchID,
			ProductID: req.ProductID,
			Type:      "OUT",
			Source:    fmt.Sprintf("AMBIL TESTER LANGSUNG: REQUEST-%s", req.ID.String()[:8]),
			Quantity:  req.Quantity,
		})

		// Add stock to salesman's SalesStock (Gudang Berjalan)
		salesTransferRepo := repository.NewSalesTransferRepository(tx)
		sStock, err := salesTransferRepo.GetSalesStock(req.EmployeeID, req.ProductID)
		now := time.Now()
		if err != nil {
			sStock = &domain.SalesStock{
				EmployeeID: req.EmployeeID,
				ProductID:  req.ProductID,
				Quantity:   req.Quantity,
				UpdatedAt:  now,
			}
			if err := tx.Create(sStock).Error; err != nil {
				return err
			}
		} else {
			sStock.Quantity += req.Quantity
			sStock.UpdatedAt = now
			if err := salesTransferRepo.UpdateSalesStock(sStock); err != nil {
				return err
			}
		}

		req.Status = domain.TesterStatusDelivered
		return reqRepo.Update(req)
	})
}

func (u *salesTesterRequestUsecase) DeleteRequest(id uuid.UUID) error {
	req, err := u.repo.FindByID(id)
	if err != nil {
		return err
	}
	if req.Status != domain.TesterStatusPending {
		return errors.New("cannot delete request that is not PENDING")
	}
	return u.repo.Delete(id)
}
