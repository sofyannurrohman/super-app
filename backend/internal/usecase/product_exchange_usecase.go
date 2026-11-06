package usecase

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"example.com/hris/backend/internal/repository"
	"gorm.io/gorm"
)

type ProductExchangeUsecase interface {
	CreateExchange(req CreateExchangeRequest) (*domain.ProductExchange, error)
	ApproveExchange(id uuid.UUID, approvedByID uuid.UUID) error
	GetByBranch(branchID uuid.UUID, status string) ([]domain.ProductExchange, error)
	GetByID(id uuid.UUID) (*domain.ProductExchange, error)
}

type productExchangeUsecase struct {
	repo          repository.ProductExchangeRepository
	warehouseRepo repository.WarehouseRepository
	db            *gorm.DB
}

func NewProductExchangeUsecase(
	repo repository.ProductExchangeRepository,
	warehouseRepo repository.WarehouseRepository,
	db *gorm.DB,
) ProductExchangeUsecase {
	return &productExchangeUsecase{repo, warehouseRepo, db}
}

type CreateExchangeRequest struct {
	StoreID      uuid.UUID             `json:"store_id" binding:"required"`
	BranchID     uuid.UUID             `json:"branch_id" binding:"required"`
	EmployeeID   uuid.UUID             `json:"employee_id"`
	DeliveryType string                `json:"delivery_type" binding:"required"` // DIRECT, DELIVERY
	Notes        string                `json:"notes"`
	Items        []ExchangeItemRequest `json:"items" binding:"required,min=1"`
}

type ExchangeItemRequest struct {
	ProductID uuid.UUID `json:"product_id" binding:"required"`
	Quantity  int       `json:"quantity" binding:"required,min=1"`
	Reason    string    `json:"reason"`
	Condition string    `json:"condition"` // DAMAGED, GOOD
}

func (u *productExchangeUsecase) CreateExchange(req CreateExchangeRequest) (*domain.ProductExchange, error) {
	now := time.Now()
	exchangeNo, err := u.repo.GenerateNextExchangeNumber(now)
	if err != nil {
		return nil, fmt.Errorf("gagal generate nomor tukar barang: %w", err)
	}

	pe := &domain.ProductExchange{
		ExchangeNo:   exchangeNo,
		StoreID:      req.StoreID,
		BranchID:     req.BranchID,
		EmployeeID:   req.EmployeeID,
		Status:       domain.ExchangeStatusPending,
		DeliveryType: req.DeliveryType,
		Notes:        req.Notes,
	}

	for _, item := range req.Items {
		cond := item.Condition
		if cond == "" {
			cond = "DAMAGED"
		}
		pe.Items = append(pe.Items, domain.ProductExchangeItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Reason:    item.Reason,
			Condition: cond,
		})
	}

	var createErr error
	for attempt := 0; attempt < 3; attempt++ {
		createErr = u.repo.Create(pe)
		if createErr == nil {
			break
		}
		if strings.Contains(createErr.Error(), "duplicate key") || strings.Contains(createErr.Error(), "unique constraint") {
			pe.ID = uuid.Nil
			newNo, genErr := u.repo.GenerateNextExchangeNumber(time.Now())
			if genErr == nil {
				pe.ExchangeNo = newNo
				continue
			}
		}
		break
	}
	if createErr != nil {
		return nil, createErr
	}

	return u.repo.FindByID(pe.ID)
}

func (u *productExchangeUsecase) ApproveExchange(id uuid.UUID, approvedByID uuid.UUID) error {
	return u.db.Transaction(func(tx *gorm.DB) error {
		peRepo := repository.NewProductExchangeRepository(tx)
		wRepo := repository.NewWarehouseRepository(tx)

		pe, err := peRepo.FindByID(id)
		if err != nil {
			return err
		}
		if pe.Status != domain.ExchangeStatusPending {
			return errors.New("hanya klaim garansi/tukar barang berstatus PENDING yang bisa disetujui")
		}

		// 1. Masukkan barang retur/rusak ke gudang DAMAGED
		for _, item := range pe.Items {
			stock, err := wRepo.GetStock(pe.BranchID, item.ProductID, "DAMAGED", "DEFAULT")
			if err != nil {
				return err
			}
			stock.Quantity += item.Quantity
			if err := wRepo.UpdateStock(stock); err != nil {
				return err
			}

			wRepo.CreateLog(&domain.WarehouseLog{
				BranchID:  pe.BranchID,
				ProductID: item.ProductID,
				Type:      "IN",
				Source:    "GARANSI/KLAIM: " + pe.ExchangeNo,
				Quantity:  item.Quantity,
			})
		}

		now := time.Now()

		// 2. Proses pengeluaran barang pengganti
		if pe.DeliveryType == "DIRECT" {
			// Jika tukar langsung di tempat, kurangi stok barang bagus (GOOD) dari gudang utama
			for _, item := range pe.Items {
				goodStock, err := wRepo.GetStock(pe.BranchID, item.ProductID, "GOOD", "DEFAULT")
				if err != nil {
					return err
				}
				if goodStock.Quantity < item.Quantity {
					return errors.New("stok barang bagus tidak mencukupi untuk melakukan tukar langsung")
				}
				goodStock.Quantity -= item.Quantity
				if err := wRepo.UpdateStock(goodStock); err != nil {
					return err
				}

				wRepo.CreateLog(&domain.WarehouseLog{
					BranchID:  pe.BranchID,
					ProductID: item.ProductID,
					Type:      "OUT",
					Source:    "GARANSI/KLAIM (PENGGANTI): " + pe.ExchangeNo,
					Quantity:  item.Quantity,
				})
			}
			pe.Status = domain.ExchangeStatusCompleted
		} else {
			// Jika DELIVERY (kirim via Surat Jalan), buat SalesSubsequentItem
			for _, item := range pe.Items {
				subItem := &domain.SalesSubsequentItem{
					ProductExchangeID: &pe.ID,
					ProductID:         item.ProductID,
					Quantity:          item.Quantity,
					Status:            domain.SubsequentStatusPending,
				}
				if err := tx.Create(subItem).Error; err != nil {
					return err
				}
			}
			pe.Status = domain.ExchangeStatusApproved
		}

		pe.ApprovedAt = &now
		pe.ApprovedByID = &approvedByID

		return peRepo.Update(pe)
	})
}

func (u *productExchangeUsecase) GetByBranch(branchID uuid.UUID, status string) ([]domain.ProductExchange, error) {
	return u.repo.FindByBranch(branchID, status)
}

func (u *productExchangeUsecase) GetByID(id uuid.UUID) (*domain.ProductExchange, error) {
	return u.repo.FindByID(id)
}
