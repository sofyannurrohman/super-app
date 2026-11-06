package usecase

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"example.com/hris/backend/internal/repository"
	"gorm.io/gorm"
)

type TechnicianInventoryUsecase interface {
	ListItems(companyID, factoryID uuid.UUID, category string, lowStockOnly bool, search string) ([]domain.TechnicianItem, error)
	GetItemByID(id uuid.UUID) (*domain.TechnicianItem, error)
	CreateItem(companyID uuid.UUID, req CreateTechnicianItemRequest) (*domain.TechnicianItem, error)
	UpdateItem(id uuid.UUID, req UpdateTechnicianItemRequest) (*domain.TechnicianItem, error)
	DeleteItem(id uuid.UUID) error
	OpnameStock(id uuid.UUID, req OpnameTechnicianStockRequest, userID *uuid.UUID, userName string) (*domain.TechnicianItem, error)
	BatchOpnameStock(factoryID, companyID uuid.UUID, req BatchOpnameTechnicianStockRequest, userID *uuid.UUID, userName string) error
	RecordUsage(factoryID, companyID uuid.UUID, req RecordTechnicianUsageRequest, userID *uuid.UUID, userName string) error
	ListStockLogs(companyID, factoryID uuid.UUID, itemID *uuid.UUID, logType string) ([]domain.TechnicianStockLog, error)
}

type CreateTechnicianItemRequest struct {
	// FactoryID is populated from the route parameter by the HTTP handler,
	// so it must not be required during JSON binding.
	FactoryID          uuid.UUID `json:"factory_id"`
	Name               string    `json:"name" binding:"required"`
	Code               string    `json:"code"`
	Category           string    `json:"category" binding:"required"`
	Unit               string    `json:"unit" binding:"required"`
	InitialQuantity    int       `json:"initial_quantity"`
	MinLimit           int       `json:"min_limit"`
	RackLocation       string    `json:"rack_location"`
	EstimatedUnitPrice float64   `json:"estimated_unit_price"`
	Description        string    `json:"description"`
}

type UpdateTechnicianItemRequest struct {
	Name               string  `json:"name" binding:"required"`
	Code               string  `json:"code"`
	Category           string  `json:"category" binding:"required"`
	Unit               string  `json:"unit" binding:"required"`
	MinLimit           int     `json:"min_limit"`
	RackLocation       string  `json:"rack_location"`
	EstimatedUnitPrice float64 `json:"estimated_unit_price"`
	Description        string  `json:"description"`
}

type OpnameTechnicianStockRequest struct {
	PhysicalQuantity int    `json:"physical_quantity"`
	Reason           string `json:"reason" binding:"required"` // e.g. Audit bulanan, Koreksi selisih fisik, Kerusakan
	Notes            string `json:"notes"`
}

type BatchOpnameItemRequest struct {
	TechnicianItemID uuid.UUID `json:"technician_item_id" binding:"required"`
	PhysicalQuantity int       `json:"physical_quantity"`
	Reason           string    `json:"reason"`
	Notes            string    `json:"notes"`
}

type BatchOpnameTechnicianStockRequest struct {
	Items []BatchOpnameItemRequest `json:"items" binding:"required,min=1"`
}

type TechnicianUsageItemRequest struct {
	TechnicianItemID uuid.UUID `json:"technician_item_id" binding:"required"`
	Quantity         int       `json:"quantity" binding:"required,min=1"`
	Notes            string    `json:"notes"`
}

type RecordTechnicianUsageRequest struct {
	TechnicianName   string                       `json:"technician_name" binding:"required"`
	MachineOrPurpose string                       `json:"machine_or_purpose" binding:"required"` // e.g. Perbaikan Mesin Sealer Line 2
	UsageDate        string                       `json:"usage_date"`                            // Supports YYYY-MM-DD or RFC3339
	Items            []TechnicianUsageItemRequest `json:"items" binding:"required,min=1"`
	Notes            string                       `json:"notes"`
}

type technicianInventoryUsecase struct {
	repo repository.TechnicianInventoryRepository
	db   *gorm.DB
}

func NewTechnicianInventoryUsecase(repo repository.TechnicianInventoryRepository, db *gorm.DB) TechnicianInventoryUsecase {
	return &technicianInventoryUsecase{
		repo: repo,
		db:   db,
	}
}

func (u *technicianInventoryUsecase) ListItems(companyID, factoryID uuid.UUID, category string, lowStockOnly bool, search string) ([]domain.TechnicianItem, error) {
	return u.repo.ListItems(companyID, factoryID, category, lowStockOnly, search)
}

func (u *technicianInventoryUsecase) GetItemByID(id uuid.UUID) (*domain.TechnicianItem, error) {
	return u.repo.GetItemByID(id)
}

func (u *technicianInventoryUsecase) CreateItem(companyID uuid.UUID, req CreateTechnicianItemRequest) (*domain.TechnicianItem, error) {
	code := strings.TrimSpace(req.Code)
	if code == "" {
		// Auto-generate code e.g. TKN-CAT-UUID
		prefix := "TKN"
		switch req.Category {
		case "MUR_BAUT":
			prefix = "TKN-BT"
		case "SPAREPART_MESIN":
			prefix = "TKN-SP"
		case "KELISTRIKAN":
			prefix = "TKN-EL"
		case "PELUMAS":
			prefix = "TKN-OL"
		case "TOOLS":
			prefix = "TKN-TL"
		}
		code = fmt.Sprintf("%s-%s", prefix, strings.ToUpper(uuid.New().String()[:6]))
	}

	minLimit := req.MinLimit
	if minLimit <= 0 {
		minLimit = 5
	}

	item := &domain.TechnicianItem{
		CompanyID:          companyID,
		FactoryID:          req.FactoryID,
		Name:               req.Name,
		Code:               code,
		Category:           req.Category,
		Unit:               strings.ToUpper(strings.TrimSpace(req.Unit)),
		Quantity:           req.InitialQuantity,
		MinLimit:           minLimit,
		RackLocation:       req.RackLocation,
		EstimatedUnitPrice: req.EstimatedUnitPrice,
		Description:        req.Description,
	}

	if err := u.repo.CreateItem(item); err != nil {
		return nil, err
	}

	if req.InitialQuantity > 0 {
		log := &domain.TechnicianStockLog{
			CompanyID:        companyID,
			FactoryID:        req.FactoryID,
			TechnicianItemID: item.ID,
			Type:             "IN",
			Source:           "INITIAL",
			Quantity:         req.InitialQuantity,
			CurrentStock:     req.InitialQuantity,
			Notes:            "Saldo Awal Master Barang",
		}
		_ = u.repo.CreateStockLog(log)
	}

	return item, nil
}

func (u *technicianInventoryUsecase) UpdateItem(id uuid.UUID, req UpdateTechnicianItemRequest) (*domain.TechnicianItem, error) {
	item, err := u.repo.GetItemByID(id)
	if err != nil {
		return nil, err
	}

	item.Name = req.Name
	if req.Code != "" {
		item.Code = req.Code
	}
	item.Category = req.Category
	item.Unit = strings.ToUpper(strings.TrimSpace(req.Unit))
	item.MinLimit = req.MinLimit
	item.RackLocation = req.RackLocation
	item.EstimatedUnitPrice = req.EstimatedUnitPrice
	item.Description = req.Description

	if err := u.repo.UpdateItem(item); err != nil {
		return nil, err
	}

	return item, nil
}

func (u *technicianInventoryUsecase) DeleteItem(id uuid.UUID) error {
	return u.repo.DeleteItem(id)
}

func (u *technicianInventoryUsecase) OpnameStock(id uuid.UUID, req OpnameTechnicianStockRequest, userID *uuid.UUID, userName string) (*domain.TechnicianItem, error) {
	item, err := u.repo.GetItemByID(id)
	if err != nil {
		return nil, err
	}

	oldQty := item.Quantity
	newQty := req.PhysicalQuantity
	diff := newQty - oldQty

	if diff == 0 {
		return item, nil
	}

	logType := "IN"
	absDiff := diff
	if diff < 0 {
		logType = "OUT"
		absDiff = -diff
	}

	item.Quantity = newQty
	if err := u.repo.UpdateItem(item); err != nil {
		return nil, err
	}

	actor := userName
	if actor == "" {
		actor = "Admin Gudang Teknik"
	}

	source := "OPNAME_ADJUSTMENT"
	if req.Reason != "" {
		source = "OPNAME: " + req.Reason
	}

	log := &domain.TechnicianStockLog{
		CompanyID:        item.CompanyID,
		FactoryID:        item.FactoryID,
		TechnicianItemID: item.ID,
		Type:             logType,
		Source:           source,
		Quantity:         absDiff,
		CurrentStock:     newQty,
		TechnicianName:   actor,
		Notes:            req.Notes,
	}
	_ = u.repo.CreateStockLog(log)

	return item, nil
}

func (u *technicianInventoryUsecase) BatchOpnameStock(factoryID, companyID uuid.UUID, req BatchOpnameTechnicianStockRequest, userID *uuid.UUID, userName string) error {
	return u.db.Transaction(func(tx *gorm.DB) error {
		repo := repository.NewTechnicianInventoryRepository(tx)

		for _, itemReq := range req.Items {
			item, err := repo.GetItemByID(itemReq.TechnicianItemID)
			if err != nil {
				continue
			}

			oldQty := item.Quantity
			newQty := itemReq.PhysicalQuantity
			diff := newQty - oldQty
			if diff == 0 {
				continue
			}

			logType := "IN"
			absDiff := diff
			if diff < 0 {
				logType = "OUT"
				absDiff = -diff
			}

			item.Quantity = newQty
			if err := repo.UpdateItem(item); err != nil {
				return err
			}

			source := "BATCH_OPNAME"
			if itemReq.Reason != "" {
				source = "OPNAME: " + itemReq.Reason
			}

			actor := userName
			if actor == "" {
				actor = "Admin Gudang Teknik"
			}

			log := &domain.TechnicianStockLog{
				CompanyID:        item.CompanyID,
				FactoryID:        item.FactoryID,
				TechnicianItemID: item.ID,
				Type:             logType,
				Source:           source,
				Quantity:         absDiff,
				CurrentStock:     newQty,
				TechnicianName:   actor,
				Notes:            itemReq.Notes,
			}
			if err := repo.CreateStockLog(log); err != nil {
				return err
			}
		}
		return nil
	})
}

func (u *technicianInventoryUsecase) RecordUsage(factoryID, companyID uuid.UUID, req RecordTechnicianUsageRequest, userID *uuid.UUID, userName string) error {
	return u.db.Transaction(func(tx *gorm.DB) error {
		repo := repository.NewTechnicianInventoryRepository(tx)

		usageDate := time.Now()
		if strings.TrimSpace(req.UsageDate) != "" {
			dateStr := strings.TrimSpace(req.UsageDate)
			if parsed, err := time.Parse("2006-01-02", dateStr); err == nil {
				now := time.Now()
				usageDate = time.Date(parsed.Year(), parsed.Month(), parsed.Day(), now.Hour(), now.Minute(), now.Second(), 0, time.Local)
			} else if parsed, err := time.Parse(time.RFC3339, dateStr); err == nil {
				usageDate = parsed
			} else if parsed, err := time.Parse("2006-01-02 15:04:05", dateStr); err == nil {
				usageDate = parsed
			}
		}

		for _, itemReq := range req.Items {
			item, err := repo.GetItemByID(itemReq.TechnicianItemID)
			if err != nil {
				return fmt.Errorf("item teknisi tidak ditemukan: %w", err)
			}

			if item.Quantity < itemReq.Quantity {
				return fmt.Errorf("stok barang '%s' tidak mencukupi (Tersedia: %d %s, Diminta: %d %s)", item.Name, item.Quantity, item.Unit, itemReq.Quantity, item.Unit)
			}

			item.Quantity -= itemReq.Quantity
			if err := repo.UpdateItem(item); err != nil {
				return err
			}

			notes := itemReq.Notes
			if notes == "" && req.Notes != "" {
				notes = req.Notes
			}

			log := &domain.TechnicianStockLog{
				CompanyID:        item.CompanyID,
				FactoryID:        factoryID,
				TechnicianItemID: item.ID,
				Type:             "OUT",
				Source:           "USAGE",
				Quantity:         itemReq.Quantity,
				CurrentStock:     item.Quantity,
				TechnicianName:   req.TechnicianName,
				MachineOrPurpose: req.MachineOrPurpose,
				Notes:            notes,
				CreatedAt:        usageDate,
			}
			if err := repo.CreateStockLog(log); err != nil {
				return err
			}
		}
		return nil
	})
}

func (u *technicianInventoryUsecase) ListStockLogs(companyID, factoryID uuid.UUID, itemID *uuid.UUID, logType string) ([]domain.TechnicianStockLog, error) {
	return u.repo.ListStockLogs(companyID, factoryID, itemID, logType)
}
