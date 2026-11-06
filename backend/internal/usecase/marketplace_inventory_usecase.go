package usecase

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"example.com/hris/backend/internal/repository"
)

type MarketplaceInventoryUsecase interface {
	CreateInventory(inventory *domain.MarketplaceInventory, actorID uuid.UUID, ip string) error
	GetInventoryByID(id uuid.UUID) (*domain.MarketplaceInventory, error)
	GetAllInventories(branchID uuid.UUID) ([]domain.MarketplaceInventory, error)
	UpdateInventory(inventory *domain.MarketplaceInventory, actorID uuid.UUID, ip string) error
	DeleteInventory(id uuid.UUID, actorID uuid.UUID, actorName, ip string) error

	// Stock Movements
	StockIn(inventoryID uuid.UUID, quantity int, reference, notes, createdBy string, actorID uuid.UUID, ip string) error
	StockOut(inventoryID uuid.UUID, quantity int, reference, notes, createdBy string, actorID uuid.UUID, ip string) error

	// Opname
	PerformOpname(inventoryID uuid.UUID, actualQuantity int, notes, createdBy string, actorID uuid.UUID, ip string) error

	// History
	GetLogs(inventoryID *uuid.UUID, from, to *time.Time) ([]domain.MarketplaceInventoryLog, error)
	GetOpnames(inventoryID *uuid.UUID, from, to *time.Time, limit, offset int) ([]domain.MarketplaceInventoryOpname, int64, error)
}

type marketplaceInventoryUsecase struct {
	repo         repository.MarketplaceInventoryRepository
	auditUsecase AuditLogUsecase
	companyID    uuid.UUID
}

func NewMarketplaceInventoryUsecase(repo repository.MarketplaceInventoryRepository, audit AuditLogUsecase, companyID string) MarketplaceInventoryUsecase {
	cid, _ := uuid.Parse(companyID)
	return &marketplaceInventoryUsecase{
		repo:         repo,
		auditUsecase: audit,
		companyID:    cid,
	}
}

func (u *marketplaceInventoryUsecase) CreateInventory(inventory *domain.MarketplaceInventory, actorID uuid.UUID, ip string) error {
	if err := u.repo.CreateInventory(inventory); err != nil {
		return err
	}
	u.auditUsecase.Log(u.companyID, actorID, inventory.CreatedBy, "CREATE", "MarketplaceInventory", inventory.ID.String(),
		"Tambah inventaris: "+inventory.Name, nil, inventory, ip)
	return nil
}

func (u *marketplaceInventoryUsecase) GetInventoryByID(id uuid.UUID) (*domain.MarketplaceInventory, error) {
	return u.repo.GetInventoryByID(id)
}

func (u *marketplaceInventoryUsecase) GetAllInventories(branchID uuid.UUID) ([]domain.MarketplaceInventory, error) {
	return u.repo.GetAllInventories(branchID)
}

func (u *marketplaceInventoryUsecase) UpdateInventory(inventory *domain.MarketplaceInventory, actorID uuid.UUID, ip string) error {
	existing, err := u.repo.GetInventoryByID(inventory.ID)
	if err != nil {
		return err
	}
	old := *existing
	inventory.Quantity = existing.Quantity // Prevent manual quantity update through standard update
	inventory.CreatedAt = existing.CreatedAt
	inventory.UpdatedAt = time.Now()

	if err := u.repo.UpdateInventory(inventory); err != nil {
		return err
	}
	u.auditUsecase.Log(u.companyID, actorID, "System", "UPDATE", "MarketplaceInventory", inventory.ID.String(),
		"Update inventaris: "+inventory.Name, old, inventory, ip)
	return nil
}

func (u *marketplaceInventoryUsecase) DeleteInventory(id uuid.UUID, actorID uuid.UUID, actorName, ip string) error {
	existing, err := u.repo.GetInventoryByID(id)
	if err != nil {
		return err
	}
	if err := u.repo.DeleteInventory(id); err != nil {
		return err
	}
	u.auditUsecase.Log(u.companyID, actorID, actorName, "DELETE", "MarketplaceInventory", id.String(),
		"Hapus inventaris: "+existing.Name, existing, nil, ip)
	return nil
}

func (u *marketplaceInventoryUsecase) StockIn(inventoryID uuid.UUID, quantity int, reference, notes, createdBy string, actorID uuid.UUID, ip string) error {
	if quantity <= 0 {
		return errors.New("quantity must be greater than zero")
	}

	inventory, err := u.repo.GetInventoryByID(inventoryID)
	if err != nil {
		return err
	}

	quantityBefore := inventory.Quantity
	inventory.Quantity += quantity

	// Update inventory
	if err := u.repo.UpdateInventory(inventory); err != nil {
		return err
	}

	// Create Log
	log := &domain.MarketplaceInventoryLog{
		MarketplaceInventoryID: inventory.ID,
		Type:                   domain.InventoryIn,
		Quantity:               quantity,
		QuantityBefore:         quantityBefore,
		QuantityAfter:          inventory.Quantity,
		Reference:              reference,
		Notes:                  notes,
		CreatedBy:              createdBy,
	}

	if err := u.repo.CreateLog(log); err != nil {
		return err
	}

	u.auditUsecase.Log(u.companyID, actorID, createdBy, "STOCK_IN", "MarketplaceInventory", inventory.ID.String(),
		"Stok Masuk: "+inventory.Name, nil, log, ip)
	return nil
}

func (u *marketplaceInventoryUsecase) StockOut(inventoryID uuid.UUID, quantity int, reference, notes, createdBy string, actorID uuid.UUID, ip string) error {
	if quantity <= 0 {
		return errors.New("quantity must be greater than zero")
	}

	inventory, err := u.repo.GetInventoryByID(inventoryID)
	if err != nil {
		return err
	}

	if inventory.Quantity < quantity {
		return errors.New("insufficient stock")
	}

	quantityBefore := inventory.Quantity
	inventory.Quantity -= quantity

	// Update inventory
	if err := u.repo.UpdateInventory(inventory); err != nil {
		return err
	}

	// Create Log
	log := &domain.MarketplaceInventoryLog{
		MarketplaceInventoryID: inventory.ID,
		Type:                   domain.InventoryOut,
		Quantity:               quantity,
		QuantityBefore:         quantityBefore,
		QuantityAfter:          inventory.Quantity,
		Reference:              reference,
		Notes:                  notes,
		CreatedBy:              createdBy,
	}

	if err := u.repo.CreateLog(log); err != nil {
		return err
	}

	u.auditUsecase.Log(u.companyID, actorID, createdBy, "STOCK_OUT", "MarketplaceInventory", inventory.ID.String(),
		"Stok Keluar: "+inventory.Name, nil, log, ip)
	return nil
}

func (u *marketplaceInventoryUsecase) PerformOpname(inventoryID uuid.UUID, actualQuantity int, notes, createdBy string, actorID uuid.UUID, ip string) error {
	if actualQuantity < 0 {
		return errors.New("actual quantity cannot be negative")
	}

	inventory, err := u.repo.GetInventoryByID(inventoryID)
	if err != nil {
		return err
	}

	quantitySystem := inventory.Quantity
	difference := actualQuantity - quantitySystem

	if difference == 0 {
		return errors.New("actual quantity is same as system quantity, no adjustment needed")
	}

	// Update stock
	inventory.Quantity = actualQuantity
	if err := u.repo.UpdateInventory(inventory); err != nil {
		return err
	}

	// Log Opname
	opname := &domain.MarketplaceInventoryOpname{
		MarketplaceInventoryID: inventory.ID,
		QuantitySystem:         quantitySystem,
		QuantityActual:         actualQuantity,
		Difference:             difference,
		Notes:                  notes,
		CreatedBy:              createdBy,
	}

	if err := u.repo.CreateOpname(opname); err != nil {
		return err
	}

	// Log Adjustment in Audit Trail
	adjQuantity := difference
	if adjQuantity < 0 {
		adjQuantity = -adjQuantity // absolute value for log
	}

	log := &domain.MarketplaceInventoryLog{
		MarketplaceInventoryID: inventory.ID,
		Type:                   domain.InventoryAdj,
		Quantity:               adjQuantity,
		QuantityBefore:         quantitySystem,
		QuantityAfter:          actualQuantity,
		Reference:              "OPNAME",
		Notes:                  notes,
		CreatedBy:              createdBy,
	}

	if err := u.repo.CreateLog(log); err != nil {
		return err
	}

	u.auditUsecase.Log(u.companyID, actorID, createdBy, "ADJUSTMENT", "MarketplaceInventory", inventory.ID.String(),
		"Opname stok: "+inventory.Name, nil, opname, ip)
	return nil
}

func (u *marketplaceInventoryUsecase) GetLogs(inventoryID *uuid.UUID, from, to *time.Time) ([]domain.MarketplaceInventoryLog, error) {
	return u.repo.GetLogs(inventoryID, from, to)
}

func (u *marketplaceInventoryUsecase) GetOpnames(inventoryID *uuid.UUID, from, to *time.Time, limit, offset int) ([]domain.MarketplaceInventoryOpname, int64, error) {
	return u.repo.GetOpnames(inventoryID, from, to, limit, offset)
}
