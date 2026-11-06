package usecase

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"example.com/hris/backend/internal/repository"
	"example.com/hris/backend/pkg/sse"
	"gorm.io/gorm"
)

type ProductionRawMaterialUsage struct {
	ProductID          uuid.UUID  `json:"product_id" binding:"required"`
	BatchNo            string     `json:"batch_no"`
	FactoryWarehouseID *uuid.UUID `json:"factory_warehouse_id"`
	Quantity           int        `json:"quantity" binding:"required,min=1"`
	Unit               string     `json:"unit"`
	Notes              string     `json:"notes"`
}

type BatchProductionItemInput struct {
	ProductID         uuid.UUID  `json:"product_id"`
	TargetWarehouseID *uuid.UUID `json:"target_warehouse_id"`
	BatchNo           string     `json:"batch_no"`
	ExpiryDate        *time.Time `json:"expiry_date"`
	Quantity          int        `json:"quantity"`
	CartonCount       int        `json:"carton_count"`
	PiecesPerCarton   int        `json:"pieces_per_carton"`
	Notes             string     `json:"notes"`
}

type FactoryUsecase interface {
	// Factory
	CreateFactory(factory *domain.Factory) error
	GetAllFactories(companyID uuid.UUID) ([]domain.Factory, error)
	GetFactoryDetail(id uuid.UUID) (*domain.Factory, error)
	UpdateFactory(factory *domain.Factory) error
	DeleteFactory(id uuid.UUID) error

	// Factory Warehouse
	CreateFactoryWarehouse(warehouse *domain.FactoryWarehouse) error
	GetFactoryWarehouses(factoryID uuid.UUID) ([]domain.FactoryWarehouse, error)
	GetFactoryWarehouseDetail(id uuid.UUID) (*domain.FactoryWarehouse, error)
	UpdateFactoryWarehouse(warehouse *domain.FactoryWarehouse) error
	DeleteFactoryWarehouse(id uuid.UUID) error

	// Product
	CreateProduct(product *domain.Product) error
	GetProducts(companyID uuid.UUID, branchID uuid.UUID) ([]domain.Product, error)
	UpdateProduct(product *domain.Product) error
	DeleteProduct(id uuid.UUID) error

	// Production
	LogProduction(factoryID, productID uuid.UUID, employeeID *uuid.UUID, targetWarehouseID *uuid.UUID, quantity, cartonCount, piecesPerCarton int, notes string, batchNo string, expiryDate *time.Time, rawMaterials []ProductionRawMaterialUsage, operators []ProductionOperatorInput) error
	LogBatchProduction(factoryID uuid.UUID, employeeID *uuid.UUID, notes string, items []BatchProductionItemInput, rawMaterials []ProductionRawMaterialUsage, operators []ProductionOperatorInput) error
	GetProductionHistory(factoryID uuid.UUID) ([]domain.ProductionLog, error)
	GetAllProductionHistory(companyID uuid.UUID) ([]domain.ProductionLog, error)
	UpdateProductionLog(id uuid.UUID, productID *uuid.UUID, targetWarehouseID *uuid.UUID, batchNo *string, expiryDate *time.Time, quantity, cartonCount, piecesPerCarton int, notes string, employeeID *uuid.UUID, operators []ProductionOperatorInput, rawMaterials []ProductionRawMaterialUsage) error
	DeleteProductionLog(id uuid.UUID) error

	// Inventory
	GetFactoryInventory(factoryID uuid.UUID) ([]domain.FactoryStock, error)
	GetAllInventory(companyID uuid.UUID) ([]domain.FactoryStock, error)
	GetInventoryLogs(factoryID uuid.UUID) ([]domain.FactoryInventoryLog, error)
	GetAllInventoryLogs(companyID uuid.UUID) ([]domain.FactoryInventoryLog, error)
	AdjustStock(factoryID, productID uuid.UUID, quantity int, reason string, batchNo string, factoryWarehouseID *uuid.UUID) error
	SyncFactoryWarehouseStock(factoryID uuid.UUID) error
	SyncAllFactoriesWarehouseStock(companyID uuid.UUID) error
	DeleteStock(id uuid.UUID) error
	DeleteInventoryLog(id uuid.UUID) error
	GetRawMaterialStock(factoryID uuid.UUID) ([]domain.FactoryStock, error)
	GetRawMaterialLogs(factoryID uuid.UUID, productID *uuid.UUID) ([]domain.FactoryInventoryLog, error)
	MutateRawMaterial(factoryID uuid.UUID, productID uuid.UUID, mutationType string, quantity int, notes string, batchNo string, targetWarehouseID *uuid.UUID) error

	// Transfer & Logistics
	RequestShipment(fromFactoryID, toBranchID uuid.UUID, items []domain.ProductTransferItem, notes string, targetShipmentDate, estimatedArrival *time.Time, initiatedBy string, vehicleID, driverID *uuid.UUID, paymentMethod string, paymentDueAt *time.Time, transferType string, isBillable *bool) error
	ApproveTransfer(transferID uuid.UUID, approverID *uuid.UUID, approverName, approverPosition string) error
	ExecuteApprovedShipment(transferID uuid.UUID) error
	GetTransferHistory(factoryID uuid.UUID) ([]domain.ProductTransfer, error)
	GetAllTransfers(companyID uuid.UUID, productID string, batchNo string) ([]domain.ProductTransfer, error)
	GetAllTransfersPaginated(page, limit int, companyID uuid.UUID, branchID, status, search string) ([]domain.ProductTransfer, int64, float64, map[string]int64, error)
	GetDelegatedShipments(companyID uuid.UUID) ([]domain.ProductTransfer, []DelegatedShipmentSummary, error)
	UpdateTransfer(id uuid.UUID, data map[string]interface{}) error
	UpdateTransferByDO(doNo string, input UpdateTransferDOInput) error
	DeleteTransfer(id uuid.UUID) error
	CreateTransfer(transfer *domain.ProductTransfer) error

	// Recipe
	CreateRecipe(recipe *domain.ProductionRecipe) error
	GetRecipes(companyID uuid.UUID) ([]domain.ProductionRecipe, error)
	GetRecipeByProduct(productID uuid.UUID) (*domain.ProductionRecipe, error)
	DeleteRecipe(id uuid.UUID) error

	// Dashboard Demand
	GetBackorderDemand(companyID uuid.UUID) ([]BackorderDemand, error)
	GetDashboardStats(companyID uuid.UUID) (*domain.FactoryDashboardStats, error)
	GetLowStockAlerts(companyID uuid.UUID) ([]domain.LowStockAlert, error)

	// Receivables
	GetBranchReceivables(companyID uuid.UUID) (*domain.FactoryReceivablesResponse, error)

	// QC Sample
	RequestQcSample(input RequestQcSampleInput) (*domain.FactoryQcSample, error)
	GetQcSamples(factoryID uuid.UUID, status string) ([]domain.FactoryQcSample, error)
	GetAllQcSamples(companyID uuid.UUID, status string) ([]domain.FactoryQcSample, error)
	GetQcSampleDetail(id uuid.UUID) (*domain.FactoryQcSample, error)
	ApproveQcSample(id uuid.UUID, approverID *uuid.UUID, approverName string) error
	RejectQcSample(id uuid.UUID, approverID *uuid.UUID, approverName string, reason string) error
}

type RequestQcSampleInput struct {
	FactoryID       uuid.UUID  `json:"factory_id" binding:"required"`
	ProductID       uuid.UUID  `json:"product_id" binding:"required"`
	ProductionLogID *uuid.UUID `json:"production_log_id"`
	BatchNo         string     `json:"batch_no" binding:"required"`
	Quantity        int        `json:"quantity" binding:"required,min=1"`
	SampleType      string     `json:"sample_type"`
	Purpose         string     `json:"purpose"`
	Notes           string     `json:"notes"`
	RequestedByID   uuid.UUID  `json:"requested_by_id"`
	RequestedByName string     `json:"requested_by_name"`
}

type BackorderDemand struct {
	ProductID   uuid.UUID `json:"product_id"`
	ProductName string    `json:"product_name"`
	TotalQty    int       `json:"total_qty"`
	OrderCount  int       `json:"order_count"`
}

type ProductionOperatorInput struct {
	EmployeeID   *uuid.UUID `json:"employee_id"`
	OperatorName string     `json:"operator_name"`
	IsManual     bool       `json:"is_manual"`
}

func resolveProductionOperators(tx *gorm.DB, fallbackEmpID *uuid.UUID, fallbackOpName string, inputOps []ProductionOperatorInput) (*uuid.UUID, string, []domain.ProductionLogOperator) {
	var primaryEmpID *uuid.UUID = fallbackEmpID
	var opNames []string
	var logOperators []domain.ProductionLogOperator

	if len(inputOps) > 0 {
		for _, op := range inputOps {
			name := strings.TrimSpace(op.OperatorName)
			var empID *uuid.UUID = op.EmployeeID

			if empID != nil && *empID != uuid.Nil {
				if primaryEmpID == nil || *primaryEmpID == uuid.Nil {
					primaryEmpID = empID
				}
				if name == "" {
					var emp domain.Employee
					if err := tx.Select("first_name, last_name").First(&emp, "id = ?", *empID).Error; err == nil {
						name = strings.TrimSpace(emp.FirstName + " " + emp.LastName)
					}
				}
			}
			if name == "" {
				continue
			}
			opNames = append(opNames, name)
			logOperators = append(logOperators, domain.ProductionLogOperator{
				EmployeeID:   empID,
				OperatorName: name,
				IsManual:     op.IsManual || (empID == nil || *empID == uuid.Nil),
			})
		}
	}

	if len(logOperators) == 0 {
		if primaryEmpID != nil && *primaryEmpID != uuid.Nil {
			var emp domain.Employee
			name := ""
			if err := tx.Select("first_name, last_name").First(&emp, "id = ?", *primaryEmpID).Error; err == nil {
				name = strings.TrimSpace(emp.FirstName + " " + emp.LastName)
			}
			if name != "" {
				opNames = append(opNames, name)
				logOperators = append(logOperators, domain.ProductionLogOperator{
					EmployeeID:   primaryEmpID,
					OperatorName: name,
					IsManual:     false,
				})
			}
		}
		if len(logOperators) == 0 && fallbackOpName != "" {
			trimmed := strings.TrimSpace(fallbackOpName)
			opNames = append(opNames, trimmed)
			logOperators = append(logOperators, domain.ProductionLogOperator{
				EmployeeID:   nil,
				OperatorName: trimmed,
				IsManual:     true,
			})
		}
	}

	consolidatedName := strings.Join(opNames, ", ")
	if consolidatedName == "" && fallbackOpName != "" {
		consolidatedName = fallbackOpName
	}

	return primaryEmpID, consolidatedName, logOperators
}

type factoryUsecase struct {
	repo   repository.FactoryRepository
	db     *gorm.DB
	sseHub *sse.Hub
}

func NewFactoryUsecase(repo repository.FactoryRepository, db *gorm.DB, sseHub *sse.Hub) FactoryUsecase {
	return &factoryUsecase{repo, db, sseHub}
}

func (u *factoryUsecase) CreateFactory(factory *domain.Factory) error {
	return u.repo.CreateFactory(factory)
}

func (u *factoryUsecase) GetAllFactories(companyID uuid.UUID) ([]domain.Factory, error) {
	return u.repo.GetFactoriesByCompanyID(companyID)
}

func (u *factoryUsecase) GetFactoryDetail(id uuid.UUID) (*domain.Factory, error) {
	return u.repo.GetFactoryByID(id)
}

func (u *factoryUsecase) UpdateFactory(factory *domain.Factory) error {
	return u.repo.UpdateFactory(factory)
}

func (u *factoryUsecase) DeleteFactory(id uuid.UUID) error {
	return u.repo.DeleteFactory(id)
}

func (u *factoryUsecase) CreateFactoryWarehouse(warehouse *domain.FactoryWarehouse) error {
	return u.repo.CreateFactoryWarehouse(warehouse)
}

func (u *factoryUsecase) GetFactoryWarehouses(factoryID uuid.UUID) ([]domain.FactoryWarehouse, error) {
	return u.repo.GetFactoryWarehousesByFactoryID(factoryID)
}

func (u *factoryUsecase) GetFactoryWarehouseDetail(id uuid.UUID) (*domain.FactoryWarehouse, error) {
	return u.repo.GetFactoryWarehouseByID(id)
}

func (u *factoryUsecase) UpdateFactoryWarehouse(warehouse *domain.FactoryWarehouse) error {
	return u.repo.UpdateFactoryWarehouse(warehouse)
}

func (u *factoryUsecase) DeleteFactoryWarehouse(id uuid.UUID) error {
	return u.repo.DeleteFactoryWarehouse(id)
}

func (u *factoryUsecase) CreateProduct(product *domain.Product) error {
	return u.repo.CreateProduct(product)
}

func (u *factoryUsecase) GetProducts(companyID uuid.UUID, branchID uuid.UUID) ([]domain.Product, error) {
	return u.repo.GetProductsByCompanyID(companyID, branchID)
}

func (u *factoryUsecase) UpdateProduct(product *domain.Product) error {
	return u.repo.UpdateProduct(product)
}

func (u *factoryUsecase) DeleteProduct(id uuid.UUID) error {
	return u.repo.DeleteProduct(id)
}

func (u *factoryUsecase) LogProduction(factoryID, productID uuid.UUID, employeeID *uuid.UUID, targetWarehouseID *uuid.UUID, quantity, cartonCount, piecesPerCarton int, notes string, batchNo string, expiryDate *time.Time, rawMaterials []ProductionRawMaterialUsage, operators []ProductionOperatorInput) error {
	if batchNo == "" {
		batchNo = generateInventoryBatch("FG", productID, time.Now())
	}
	return u.db.Transaction(func(tx *gorm.DB) error {
		repo := repository.NewFactoryRepository(tx)

		// 0. Operator resolution
		primaryEmpID, consolidatedOpName, logOperators := resolveProductionOperators(tx, employeeID, "", operators)

		// Anti-double submission check (Idempotency window 15s)
		var recentLog domain.ProductionLog
		chkQuery := tx.Where("factory_id = ? AND product_id = ? AND batch_no = ? AND quantity = ? AND created_at >= ?",
			factoryID, productID, batchNo, quantity, time.Now().Add(-15*time.Second))
		if primaryEmpID != nil && *primaryEmpID != uuid.Nil {
			chkQuery = chkQuery.Where("employee_id = ?", *primaryEmpID)
		}
		if err := chkQuery.First(&recentLog).Error; err == nil && recentLog.ID != uuid.Nil {
			// Already recorded recently, return gracefully to prevent duplicate stock & logs
			return nil
		}

		// 1. Kumpulkan bahan baku yang terpakai untuk dicatat ke ProductionLog
		var usedMaterials []domain.ProductionLogMaterial

		if len(rawMaterials) > 0 {
			for _, rm := range rawMaterials {
				if rm.Quantity <= 0 {
					continue
				}

				prod, _ := repo.GetProductByID(rm.ProductID)
				pName := "Bahan Baku"
				pUnit := "PCS"
				pcsPerUnit := 1
				if prod != nil {
					pName = prod.Name
					pUnit = prod.Unit
					pcsPerUnit = prod.PcsPerUnit
				}

				// Normalize quantity to product base unit if unit was specified
				neededQty := rm.Quantity
				if rm.Unit != "" && prod != nil {
					neededQty = convertQuantityByUnit(rm.Quantity, rm.Unit, pUnit, pcsPerUnit)
				}
				if neededQty <= 0 {
					neededQty = rm.Quantity
				}

				// Jika batch_no ditentukan spesifik dan bukan "AUTO"
				if rm.BatchNo != "" && rm.BatchNo != "AUTO" {
					stock, err := repo.GetStock(factoryID, rm.ProductID, rm.BatchNo)
					if err != nil {
						return err
					}
					if stock == nil || stock.Quantity < neededQty {
						available := 0
						if stock != nil {
							available = stock.Quantity
						}
						return fmt.Errorf("stok %s pada batch %s tidak mencukupi (Dibutuhkan: %d %s, Tersedia: %d %s)", pName, rm.BatchNo, neededQty, pUnit, available, pUnit)
					}

					stock.Quantity -= neededQty
					if err := repo.UpdateStock(stock); err != nil {
						return err
					}

					// Catat Log Mutasi Bahan Baku (OUT)
					if err := repo.CreateInventoryLog(&domain.FactoryInventoryLog{
						FactoryID: factoryID,
						ProductID: rm.ProductID,
						Type:      "OUT",
						Source:    "PRODUCTION_CONSUMPTION",
						BatchNo:   stock.BatchNo,
						Quantity:  neededQty,
					}); err != nil {
						return err
					}

					usedMaterials = append(usedMaterials, domain.ProductionLogMaterial{
						ProductID: rm.ProductID,
						BatchNo:   stock.BatchNo,
						Quantity:  neededQty,
						Notes:     rm.Notes,
					})
				} else {
					// Fallback ke FEFO otomatis
					stocks, err := repo.GetAvailableBatches(factoryID, rm.ProductID)
					if err != nil {
						return err
					}

					totalAvailable := 0
					for _, s := range stocks {
						totalAvailable += s.Quantity
					}

					if totalAvailable < neededQty {
						return fmt.Errorf("stok %s tidak mencukupi (Dibutuhkan: %d %s, Tersedia: %d %s)", pName, neededQty, pUnit, totalAvailable, pUnit)
					}

					remainingToDeduct := neededQty
					for i := range stocks {
						s := &stocks[i]
						if remainingToDeduct <= 0 {
							break
						}

						deduct := s.Quantity
						if deduct > remainingToDeduct {
							deduct = remainingToDeduct
						}

						s.Quantity -= deduct
						remainingToDeduct -= deduct

						if err := repo.UpdateStock(s); err != nil {
							return err
						}

						// Catat Log Mutasi Bahan Baku (OUT)
						if err := repo.CreateInventoryLog(&domain.FactoryInventoryLog{
							FactoryID: factoryID,
							ProductID: rm.ProductID,
							Type:      "OUT",
							Source:    "PRODUCTION_CONSUMPTION",
							BatchNo:   s.BatchNo,
							Quantity:  deduct,
						}); err != nil {
							return err
						}

						usedMaterials = append(usedMaterials, domain.ProductionLogMaterial{
							ProductID: rm.ProductID,
							BatchNo:   s.BatchNo,
							Quantity:  deduct,
							Notes:     rm.Notes,
						})
					}
				}
			}
		} else {
			// 2. Fallback: Cek apakah ada resep/BOM yang terdaftar
			recipe, err := repo.GetRecipeByFinishedProduct(productID)
			if err == nil && recipe != nil {
				for _, item := range recipe.Items {
					neededQty := float64(quantity) * item.Quantity

					stocks, err := repo.GetAvailableBatches(factoryID, item.RawProductID)
					if err != nil {
						return err
					}

					totalAvailable := 0
					for _, s := range stocks {
						totalAvailable += s.Quantity
					}

					if float64(totalAvailable) < neededQty {
						return fmt.Errorf("stok bahan baku %s tidak mencukupi (Butuh: %.2f, Ada: %d)", item.RawProduct.Name, neededQty, totalAvailable)
					}

					remainingToDeduct := int(neededQty)
					for i := range stocks {
						s := &stocks[i]
						if remainingToDeduct <= 0 {
							break
						}

						deduct := s.Quantity
						if deduct > remainingToDeduct {
							deduct = remainingToDeduct
						}

						s.Quantity -= deduct
						remainingToDeduct -= deduct

						if err := repo.UpdateStock(s); err != nil {
							return err
						}

						// Catat Log Mutasi Bahan Baku (OUT)
						if err := repo.CreateInventoryLog(&domain.FactoryInventoryLog{
							FactoryID: factoryID,
							ProductID: item.RawProductID,
							Type:      "OUT",
							Source:    "PRODUCTION_CONSUMPTION",
							BatchNo:   s.BatchNo,
							Quantity:  deduct,
						}); err != nil {
							return err
						}

						usedMaterials = append(usedMaterials, domain.ProductionLogMaterial{
							ProductID: item.RawProductID,
							BatchNo:   s.BatchNo,
							Quantity:  deduct,
						})
					}
				}
			}
		}

		// 3. Tambahkan stok Barang Jadi
		whNo := 12
		product, err := repo.GetProductByID(productID)
		if err == nil && product != nil {
			whNo = product.GetWarehouseNumber()
		}

		productionLog := &domain.ProductionLog{
			FactoryID:         factoryID,
			ProductID:         productID,
			TargetWarehouseID: targetWarehouseID,
			BatchNo:           batchNo,
			ExpiryDate:        expiryDate,
			EmployeeID:        primaryEmpID,
			OperatorName:      consolidatedOpName,
			Quantity:          quantity,
			CartonCount:       cartonCount,
			PiecesPerCarton:   piecesPerCarton,
			ProductionDate:    time.Now(),
			Notes:             notes,
			WarehouseNumber:   whNo,
			Materials:         usedMaterials,
			Operators:         logOperators,
		}
		if err := repo.CreateProductionLog(productionLog); err != nil {
			return err
		}

		stock, err := repo.GetStock(factoryID, productID, batchNo)
		if err != nil && err != gorm.ErrRecordNotFound {
			return err
		}
		stock.Quantity += quantity
		stock.ExpiryDate = expiryDate
		stock.WarehouseNumber = whNo
		if targetWarehouseID != nil {
			stock.FactoryWarehouseID = targetWarehouseID
		}
		if err := repo.UpdateStock(stock); err != nil {
			return err
		}

		// Sync to Warehouse Stock (Gudang Pabrik & Gudang Trenggalek 1:1)
		if err := syncFactoryStockDeltaToLinkedWarehouses(tx, factoryID, productID, batchNo, expiryDate, quantity, "PRODUCTION", "SYSTEM"); err != nil {
			return err
		}

		inventoryLog := &domain.FactoryInventoryLog{
			FactoryID: factoryID,
			ProductID: productID,
			Type:      "IN",
			Source:    "PRODUCTION_OUTPUT",
			BatchNo:   batchNo,
			Quantity:  quantity,
		}
		return repo.CreateInventoryLog(inventoryLog)
	})
}

func (u *factoryUsecase) LogBatchProduction(factoryID uuid.UUID, employeeID *uuid.UUID, notes string, items []BatchProductionItemInput, rawMaterials []ProductionRawMaterialUsage, operators []ProductionOperatorInput) error {
	if len(items) == 0 {
		return errors.New("tidak ada produk yang diproduksi")
	}

	return u.db.Transaction(func(tx *gorm.DB) error {
		repo := repository.NewFactoryRepository(tx)

		// 0. Selesaikan data operator
		primaryEmpID, consolidatedOpName, logOperators := resolveProductionOperators(tx, employeeID, "", operators)

		// 1. Kumpulkan bahan baku yang terpakai untuk dicatat ke ProductionLog & potong stok secara global
		var usedMaterials []domain.ProductionLogMaterial

		if len(rawMaterials) > 0 {
			for _, rm := range rawMaterials {
				if rm.Quantity <= 0 {
					continue
				}

				prod, _ := repo.GetProductByID(rm.ProductID)
				pName := "Bahan Baku"
				pUnit := "PCS"
				pcsPerUnit := 1
				if prod != nil {
					pName = prod.Name
					pUnit = prod.Unit
					pcsPerUnit = prod.PcsPerUnit
				}

				// Normalize quantity to product base unit if unit was specified
				neededQty := rm.Quantity
				if rm.Unit != "" && prod != nil {
					neededQty = convertQuantityByUnit(rm.Quantity, rm.Unit, pUnit, pcsPerUnit)
				}
				if neededQty <= 0 {
					neededQty = rm.Quantity
				}

				// Jika batch_no ditentukan spesifik dan bukan "AUTO"
				if rm.BatchNo != "" && rm.BatchNo != "AUTO" {
					stock, err := repo.GetStock(factoryID, rm.ProductID, rm.BatchNo)
					if err != nil {
						return err
					}
					if stock == nil || stock.Quantity < neededQty {
						available := 0
						if stock != nil {
							available = stock.Quantity
						}
						return fmt.Errorf("stok %s pada batch %s tidak mencukupi (Dibutuhkan: %d %s, Tersedia: %d %s)", pName, rm.BatchNo, neededQty, pUnit, available, pUnit)
					}

					stock.Quantity -= neededQty
					if err := repo.UpdateStock(stock); err != nil {
						return err
					}

					// Catat Log Mutasi Bahan Baku (OUT)
					if err := repo.CreateInventoryLog(&domain.FactoryInventoryLog{
						FactoryID: factoryID,
						ProductID: rm.ProductID,
						Type:      "OUT",
						Source:    "PRODUCTION_CONSUMPTION",
						BatchNo:   stock.BatchNo,
						Quantity:  neededQty,
					}); err != nil {
						return err
					}

					usedMaterials = append(usedMaterials, domain.ProductionLogMaterial{
						ProductID: rm.ProductID,
						BatchNo:   stock.BatchNo,
						Quantity:  neededQty,
						Notes:     rm.Notes,
					})
				} else {
					// Fallback ke FEFO otomatis
					stocks, err := repo.GetAvailableBatches(factoryID, rm.ProductID)
					if err != nil {
						return err
					}

					totalAvailable := 0
					for _, s := range stocks {
						totalAvailable += s.Quantity
					}

					if totalAvailable < neededQty {
						return fmt.Errorf("stok %s tidak mencukupi (Dibutuhkan: %d %s, Tersedia: %d %s)", pName, neededQty, pUnit, totalAvailable, pUnit)
					}

					remainingToDeduct := neededQty
					for i := range stocks {
						s := &stocks[i]
						if remainingToDeduct <= 0 {
							break
						}

						deduct := s.Quantity
						if deduct > remainingToDeduct {
							deduct = remainingToDeduct
						}

						s.Quantity -= deduct
						remainingToDeduct -= deduct

						if err := repo.UpdateStock(s); err != nil {
							return err
						}

						// Catat Log Mutasi Bahan Baku (OUT)
						if err := repo.CreateInventoryLog(&domain.FactoryInventoryLog{
							FactoryID: factoryID,
							ProductID: rm.ProductID,
							Type:      "OUT",
							Source:    "PRODUCTION_CONSUMPTION",
							BatchNo:   s.BatchNo,
							Quantity:  deduct,
						}); err != nil {
							return err
						}

						usedMaterials = append(usedMaterials, domain.ProductionLogMaterial{
							ProductID: rm.ProductID,
							BatchNo:   s.BatchNo,
							Quantity:  deduct,
							Notes:     rm.Notes,
						})
					}
				}
			}
		}

		batchGroupID := uuid.New()

		// 2. Tambahkan stok Barang Jadi untuk setiap item
		for _, item := range items {
			if item.Quantity <= 0 {
				continue
			}

			bNo := item.BatchNo
			if bNo == "" {
				bNo = generateInventoryBatch("FG", item.ProductID, time.Now())
			}

			whNo := 12
			product, err := repo.GetProductByID(item.ProductID)
			if err == nil && product != nil {
				whNo = product.GetWarehouseNumber()
			}

			// Salin materials untuk tiap production log agar tetap dapat dilacak rinciannya
			itemMaterials := make([]domain.ProductionLogMaterial, len(usedMaterials))
			for idx, m := range usedMaterials {
				itemMaterials[idx] = domain.ProductionLogMaterial{
					ProductID: m.ProductID,
					BatchNo:   m.BatchNo,
					Quantity:  m.Quantity,
					Notes:     m.Notes,
				}
			}

			itemNotes := notes
			if item.Notes != "" {
				if itemNotes != "" {
					itemNotes += " - " + item.Notes
				} else {
					itemNotes = item.Notes
				}
			}

			itemOperators := make([]domain.ProductionLogOperator, len(logOperators))
			for idx, op := range logOperators {
				itemOperators[idx] = domain.ProductionLogOperator{
					EmployeeID:   op.EmployeeID,
					OperatorName: op.OperatorName,
					IsManual:     op.IsManual,
				}
			}

			productionLog := &domain.ProductionLog{
				FactoryID:         factoryID,
				ProductID:         item.ProductID,
				TargetWarehouseID: item.TargetWarehouseID,
				BatchNo:           bNo,
				ExpiryDate:        item.ExpiryDate,
				EmployeeID:        primaryEmpID,
				OperatorName:      consolidatedOpName,
				Quantity:          item.Quantity,
				CartonCount:       item.CartonCount,
				PiecesPerCarton:   item.PiecesPerCarton,
				ProductionDate:    time.Now(),
				Notes:             itemNotes,
				BatchGroupID:      &batchGroupID,
				WarehouseNumber:   whNo,
				Materials:         itemMaterials,
				Operators:         itemOperators,
			}
			if err := repo.CreateProductionLog(productionLog); err != nil {
				return err
			}

			stock, err := repo.GetStock(factoryID, item.ProductID, bNo)
			if err != nil && err != gorm.ErrRecordNotFound {
				return err
			}
			stock.Quantity += item.Quantity
			stock.ExpiryDate = item.ExpiryDate
			stock.WarehouseNumber = whNo
			if item.TargetWarehouseID != nil {
				stock.FactoryWarehouseID = item.TargetWarehouseID
			}
			if err := repo.UpdateStock(stock); err != nil {
				return err
			}

			// Sync to Warehouse Stock (Gudang Pabrik & Gudang Trenggalek 1:1)
			if err := syncFactoryStockDeltaToLinkedWarehouses(tx, factoryID, item.ProductID, bNo, item.ExpiryDate, item.Quantity, "PRODUCTION", "SYSTEM"); err != nil {
				return err
			}

			inventoryLog := &domain.FactoryInventoryLog{
				FactoryID: factoryID,
				ProductID: item.ProductID,
				Type:      "IN",
				Source:    "PRODUCTION_OUTPUT",
				BatchNo:   bNo,
				Quantity:  item.Quantity,
			}
			if err := repo.CreateInventoryLog(inventoryLog); err != nil {
				return err
			}
		}

		return nil
	})
}

// --- Recipe Implementations ---

func (u *factoryUsecase) CreateRecipe(recipe *domain.ProductionRecipe) error {
	return u.repo.CreateRecipe(recipe)
}

func (u *factoryUsecase) GetRecipes(companyID uuid.UUID) ([]domain.ProductionRecipe, error) {
	return u.repo.GetRecipesByCompany(companyID)
}

func (u *factoryUsecase) GetRecipeByProduct(productID uuid.UUID) (*domain.ProductionRecipe, error) {
	return u.repo.GetRecipeByFinishedProduct(productID)
}

func (u *factoryUsecase) DeleteRecipe(id uuid.UUID) error {
	return u.repo.DeleteRecipe(id)
}

func (u *factoryUsecase) GetProductionHistory(factoryID uuid.UUID) ([]domain.ProductionLog, error) {
	return u.repo.GetProductionLogsByFactoryID(factoryID)
}

func (u *factoryUsecase) GetAllProductionHistory(companyID uuid.UUID) ([]domain.ProductionLog, error) {
	return u.repo.GetAllProductionLogsByCompanyID(companyID)
}

func (u *factoryUsecase) UpdateProductionLog(id uuid.UUID, productID *uuid.UUID, targetWarehouseID *uuid.UUID, batchNo *string, expiryDate *time.Time, quantity, cartonCount, piecesPerCarton int, notes string, employeeID *uuid.UUID, operators []ProductionOperatorInput, rawMaterials []ProductionRawMaterialUsage) error {
	return u.db.Transaction(func(tx *gorm.DB) error {
		repo := repository.NewFactoryRepository(tx)

		log, err := repo.GetProductionLogByID(id)
		if err != nil {
			return err
		}

		oldProductID := log.ProductID
		oldBatchNo := log.BatchNo
		oldQuantity := log.Quantity
		oldExpiryDate := log.ExpiryDate

		newProductID := oldProductID
		if productID != nil && *productID != uuid.Nil {
			newProductID = *productID
		}

		newBatchNo := oldBatchNo
		if batchNo != nil && strings.TrimSpace(*batchNo) != "" {
			newBatchNo = strings.TrimSpace(*batchNo)
		}

		newExpiryDate := oldExpiryDate
		if expiryDate != nil {
			newExpiryDate = expiryDate
		}

		isProductChanged := newProductID != oldProductID
		isBatchChanged := newBatchNo != oldBatchNo

		if isProductChanged || isBatchChanged {
			// Validasi: Cek apakah batch / produk lama sudah pernah dibuatkan surat jalan / transfer antar gudang
			var transferCount int64
			if err := tx.Model(&domain.ProductTransfer{}).Where("product_id = ? AND batch_no = ?", oldProductID, oldBatchNo).Count(&transferCount).Error; err == nil && transferCount > 0 {
				return fmt.Errorf("produk / batch tidak dapat diubah karena batch %s sudah memiliki %d riwayat pengiriman transfer ke cabang", oldBatchNo, transferCount)
			}
		}

		if isProductChanged || isBatchChanged {
			// 1. Revert stok produk lama
			oldStock, err := repo.GetStock(log.FactoryID, oldProductID, oldBatchNo)
			if err != nil {
				return err
			}
			oldStock.Quantity -= oldQuantity
			if oldStock.Quantity < 0 {
				oldStock.Quantity = 0
			}
			if err := repo.UpdateStock(oldStock); err != nil {
				return err
			}

			// Sinkronisasi pengurangan stok gudang cabang yang terhubung
			if err := syncFactoryStockDeltaToLinkedWarehouses(tx, log.FactoryID, oldProductID, oldBatchNo, oldExpiryDate, -oldQuantity, "UPDATE_PRODUCTION_LOG_OLD_PRODUCT", "SYSTEM"); err != nil {
				return err
			}

			_ = repo.CreateInventoryLog(&domain.FactoryInventoryLog{
				FactoryID: log.FactoryID,
				ProductID: oldProductID,
				Type:      "OUT",
				Source:    "UPDATE_PRODUCTION_LOG_CHANGE",
				BatchNo:   oldBatchNo,
				Quantity:  oldQuantity,
			})

			// 2. Tambahkan stok produk baru
			newStock, err := repo.GetStock(log.FactoryID, newProductID, newBatchNo)
			if err != nil {
				return err
			}
			newStock.Quantity += quantity
			newStock.ExpiryDate = newExpiryDate
			if targetWarehouseID != nil {
				newStock.FactoryWarehouseID = targetWarehouseID
			}
			if err := repo.UpdateStock(newStock); err != nil {
				return err
			}

			// Sinkronisasi penambahan stok gudang cabang yang terhubung
			if err := syncFactoryStockDeltaToLinkedWarehouses(tx, log.FactoryID, newProductID, newBatchNo, newExpiryDate, quantity, "UPDATE_PRODUCTION_LOG_NEW_PRODUCT", "SYSTEM"); err != nil {
				return err
			}

			_ = repo.CreateInventoryLog(&domain.FactoryInventoryLog{
				FactoryID: log.FactoryID,
				ProductID: newProductID,
				Type:      "IN",
				Source:    "UPDATE_PRODUCTION_LOG_CHANGE",
				BatchNo:   newBatchNo,
				Quantity:  quantity,
			})
		} else {
			// Produk dan batch sama: sesuaikan selisih (diff)
			diff := quantity - oldQuantity
			stock, err := repo.GetStock(log.FactoryID, log.ProductID, log.BatchNo)
			if err != nil {
				return err
			}
			stock.Quantity += diff
			if stock.Quantity < 0 {
				stock.Quantity = 0
			}
			stock.ExpiryDate = newExpiryDate
			if targetWarehouseID != nil {
				stock.FactoryWarehouseID = targetWarehouseID
			}
			if err := repo.UpdateStock(stock); err != nil {
				return err
			}

			// Sync to Warehouse Stock (Gudang Pabrik & Gudang Trenggalek 1:1)
			if err := syncFactoryStockDeltaToLinkedWarehouses(tx, log.FactoryID, log.ProductID, log.BatchNo, newExpiryDate, diff, "UPDATE_PRODUCTION_LOG", "SYSTEM"); err != nil {
				return err
			}
		}

		// Reconcile Raw Materials
		// 1. Revert previous raw materials back to FactoryStock
		for _, oldMat := range log.Materials {
			if oldMat.Quantity <= 0 {
				continue
			}
			matStock, err := repo.GetStock(log.FactoryID, oldMat.ProductID, oldMat.BatchNo)
			if err == nil && matStock != nil {
				matStock.Quantity += oldMat.Quantity
				if err := repo.UpdateStock(matStock); err != nil {
					return err
				}
				_ = repo.CreateInventoryLog(&domain.FactoryInventoryLog{
					FactoryID: log.FactoryID,
					ProductID: oldMat.ProductID,
					Type:      "IN",
					Source:    "UPDATE_PRODUCTION_LOG_RESTORE_MATERIAL",
					BatchNo:   oldMat.BatchNo,
					Quantity:  oldMat.Quantity,
				})
			}
		}

		// 2. Remove old materials records
		if err := tx.Where("production_log_id = ?", log.ID).Delete(&domain.ProductionLogMaterial{}).Error; err != nil {
			return err
		}

		// 3. Deduct new raw materials from FactoryStock
		var newUsedMaterials []domain.ProductionLogMaterial
		if len(rawMaterials) > 0 {
			for _, rm := range rawMaterials {
				if rm.Quantity <= 0 {
					continue
				}

				prod, _ := repo.GetProductByID(rm.ProductID)
				pName := "Bahan Baku"
				pUnit := "PCS"
				pcsPerUnit := 1
				if prod != nil {
					pName = prod.Name
					pUnit = prod.Unit
					pcsPerUnit = prod.PcsPerUnit
				}

				neededQty := rm.Quantity
				if rm.Unit != "" && prod != nil {
					neededQty = convertQuantityByUnit(rm.Quantity, rm.Unit, pUnit, pcsPerUnit)
				}
				if neededQty <= 0 {
					neededQty = rm.Quantity
				}

				if rm.BatchNo != "" && rm.BatchNo != "AUTO" {
					stock, err := repo.GetStock(log.FactoryID, rm.ProductID, rm.BatchNo)
					if err != nil {
						return err
					}
					if stock == nil || stock.Quantity < neededQty {
						available := 0
						if stock != nil {
							available = stock.Quantity
						}
						return fmt.Errorf("stok %s pada batch %s tidak mencukupi (Dibutuhkan: %d %s, Tersedia: %d %s)", pName, rm.BatchNo, neededQty, pUnit, available, pUnit)
					}

					stock.Quantity -= neededQty
					if err := repo.UpdateStock(stock); err != nil {
						return err
					}

					if err := repo.CreateInventoryLog(&domain.FactoryInventoryLog{
						FactoryID: log.FactoryID,
						ProductID: rm.ProductID,
						Type:      "OUT",
						Source:    "PRODUCTION_CONSUMPTION",
						BatchNo:   stock.BatchNo,
						Quantity:  neededQty,
					}); err != nil {
						return err
					}

					newUsedMaterials = append(newUsedMaterials, domain.ProductionLogMaterial{
						ProductionLogID: log.ID,
						ProductID:       rm.ProductID,
						BatchNo:         stock.BatchNo,
						Quantity:        neededQty,
						Notes:           rm.Notes,
					})
				} else {
					stocks, err := repo.GetAvailableBatches(log.FactoryID, rm.ProductID)
					if err != nil {
						return err
					}

					totalAvailable := 0
					for _, s := range stocks {
						totalAvailable += s.Quantity
					}

					if totalAvailable < neededQty {
						return fmt.Errorf("stok %s tidak mencukupi (Dibutuhkan: %d %s, Tersedia: %d %s)", pName, neededQty, pUnit, totalAvailable, pUnit)
					}

					remainingToDeduct := neededQty
					for i := range stocks {
						s := &stocks[i]
						if remainingToDeduct <= 0 {
							break
						}

						deduct := s.Quantity
						if deduct > remainingToDeduct {
							deduct = remainingToDeduct
						}

						s.Quantity -= deduct
						remainingToDeduct -= deduct

						if err := repo.UpdateStock(s); err != nil {
							return err
						}

						if err := repo.CreateInventoryLog(&domain.FactoryInventoryLog{
							FactoryID: log.FactoryID,
							ProductID: rm.ProductID,
							Type:      "OUT",
							Source:    "PRODUCTION_CONSUMPTION",
							BatchNo:   s.BatchNo,
							Quantity:  deduct,
						}); err != nil {
							return err
						}

						newUsedMaterials = append(newUsedMaterials, domain.ProductionLogMaterial{
							ProductionLogID: log.ID,
							ProductID:       rm.ProductID,
							BatchNo:         s.BatchNo,
							Quantity:        deduct,
							Notes:           rm.Notes,
						})
					}
				}
			}

			for i := range newUsedMaterials {
				if err := tx.Create(&newUsedMaterials[i]).Error; err != nil {
					return err
				}
			}
		}

		log.ProductID = newProductID
		log.BatchNo = newBatchNo
		log.ExpiryDate = newExpiryDate
		if targetWarehouseID != nil {
			log.TargetWarehouseID = targetWarehouseID
		}
		log.Quantity = quantity
		log.CartonCount = cartonCount
		log.PiecesPerCarton = piecesPerCarton
		log.Notes = notes

		if len(operators) > 0 || (employeeID != nil && *employeeID != uuid.Nil) {
			primaryEmpID, consolidatedOpName, logOperators := resolveProductionOperators(tx, employeeID, "", operators)
			log.EmployeeID = primaryEmpID
			log.OperatorName = consolidatedOpName

			// Replace existing operators
			if err := tx.Where("production_log_id = ?", log.ID).Delete(&domain.ProductionLogOperator{}).Error; err != nil {
				return err
			}
			for i := range logOperators {
				logOperators[i].ProductionLogID = log.ID
				if err := tx.Create(&logOperators[i]).Error; err != nil {
					return err
				}
			}
		}

		return repo.UpdateProductionLog(log)
	})
}

func (u *factoryUsecase) DeleteProductionLog(id uuid.UUID) error {
	return u.db.Transaction(func(tx *gorm.DB) error {
		repo := repository.NewFactoryRepository(tx)

		log, err := repo.GetProductionLogByID(id)
		if err != nil {
			return err
		}

		// Reverse Stock Finished Good
		stock, err := repo.GetStock(log.FactoryID, log.ProductID, log.BatchNo)
		if err != nil {
			return err
		}
		stock.Quantity -= log.Quantity
		if stock.Quantity < 0 {
			stock.Quantity = 0
		}
		if err := repo.UpdateStock(stock); err != nil {
			return err
		}

		// Sync to Warehouse Stock (Gudang Pabrik & Gudang Trenggalek 1:1)
		if err := syncFactoryStockDeltaToLinkedWarehouses(tx, log.FactoryID, log.ProductID, log.BatchNo, log.ExpiryDate, -log.Quantity, "DELETE_PRODUCTION_LOG", "SYSTEM"); err != nil {
			return err
		}

		_ = repo.CreateInventoryLog(&domain.FactoryInventoryLog{
			FactoryID: log.FactoryID,
			ProductID: log.ProductID,
			Type:      "OUT",
			Source:    "DELETE_PRODUCTION_LOG",
			BatchNo:   log.BatchNo,
			Quantity:  log.Quantity,
		})

		// Reverse Raw Materials Stock
		for _, mat := range log.Materials {
			if mat.Quantity <= 0 {
				continue
			}
			matStock, err := repo.GetStock(log.FactoryID, mat.ProductID, mat.BatchNo)
			if err == nil && matStock != nil {
				matStock.Quantity += mat.Quantity
				_ = repo.UpdateStock(matStock)
				_ = repo.CreateInventoryLog(&domain.FactoryInventoryLog{
					FactoryID: log.FactoryID,
					ProductID: mat.ProductID,
					Type:      "IN",
					Source:    "DELETE_PRODUCTION_LOG_RESTORE",
					BatchNo:   mat.BatchNo,
					Quantity:  mat.Quantity,
				})
			}
		}

		return repo.DeleteProductionLog(id)
	})
}

func (u *factoryUsecase) GetFactoryInventory(factoryID uuid.UUID) ([]domain.FactoryStock, error) {
	return u.repo.GetStocksByFactoryID(factoryID)
}

func (u *factoryUsecase) GetAllInventory(companyID uuid.UUID) ([]domain.FactoryStock, error) {
	return u.repo.GetAllStocksByCompanyID(companyID)
}

func (u *factoryUsecase) GetInventoryLogs(factoryID uuid.UUID) ([]domain.FactoryInventoryLog, error) {
	return u.repo.GetInventoryLogsByFactoryID(factoryID)
}

func (u *factoryUsecase) GetAllInventoryLogs(companyID uuid.UUID) ([]domain.FactoryInventoryLog, error) {
	return u.repo.GetAllInventoryLogsByCompanyID(companyID)
}

func (u *factoryUsecase) AdjustStock(factoryID, productID uuid.UUID, quantity int, reason string, batchNo string, factoryWarehouseID *uuid.UUID) error {
	if batchNo == "" {
		batchNo = "DEFAULT"
	}
	return u.db.Transaction(func(tx *gorm.DB) error {
		repo := repository.NewFactoryRepository(tx)

		stock, err := repo.GetStock(factoryID, productID, batchNo)
		if err != nil {
			return err
		}

		diff := quantity - stock.Quantity
		if diff == 0 && factoryWarehouseID == nil {
			return nil
		}

		logType := "IN"
		absDiff := diff
		if diff < 0 {
			logType = "OUT"
			absDiff = -diff
		}

		stock.Quantity = quantity
		if factoryWarehouseID != nil {
			stock.FactoryWarehouseID = factoryWarehouseID
		}

		if err := repo.UpdateStock(stock); err != nil {
			return err
		}

		// Store reason into Source if it's not too long, but prefix it with ADJUSTMENT
		source := "ADJUSTMENT"
		if reason != "" {
			source = "ADJUSTMENT - " + reason
			if len(source) > 50 {
				source = source[:50]
			}
		}

		inventoryLog := &domain.FactoryInventoryLog{
			FactoryID: factoryID,
			ProductID: productID,
			Type:      logType,
			Source:    source,
			BatchNo:   batchNo,
			Quantity:  absDiff,
		}
		if err := repo.CreateInventoryLog(inventoryLog); err != nil {
			return err
		}

		// Sync 1:1 to Warehouse Stock (Gudang Pabrik & Gudang Trenggalek)
		whLogSource := "PABRIK_ADJUSTMENT"
		if reason != "" {
			whLogSource = "PABRIK_ADJUSTMENT: " + reason
			if len(whLogSource) > 50 {
				whLogSource = whLogSource[:50]
			}
		}
		if err := syncFactoryStockAbsoluteToLinkedWarehouses(tx, factoryID, productID, batchNo, stock.ExpiryDate, quantity, whLogSource, "SYSTEM"); err != nil {
			return err
		}

		return nil
	})
}

func (u *factoryUsecase) SyncFactoryWarehouseStock(factoryID uuid.UUID) error {
	return u.db.Transaction(func(tx *gorm.DB) error {
		repo := repository.NewFactoryRepository(tx)
		wRepo := repository.NewWarehouseRepository(tx)

		branchIDs, err := GetFactoryLinkedBranchIDs(tx, factoryID)
		if err != nil {
			return fmt.Errorf("failed to get linked branches: %w", err)
		}

		stocks, err := repo.GetStocksByFactoryID(factoryID)
		if err != nil {
			return err
		}

		for _, fs := range stocks {
			for _, branchID := range branchIDs {
				whStock, _ := wRepo.GetStock(branchID, fs.ProductID, "GOOD", fs.BatchNo)
				if whStock == nil || whStock.ID == uuid.Nil {
					whStock = &domain.WarehouseStock{
						ID:        uuid.New(),
						BranchID:  branchID,
						ProductID: fs.ProductID,
						BatchNo:   fs.BatchNo,
						StockType: "GOOD",
					}
				}
				whStock.Quantity = fs.Quantity
				whStock.ExpiryDate = fs.ExpiryDate
				if err := wRepo.UpdateStock(whStock); err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func (u *factoryUsecase) SyncAllFactoriesWarehouseStock(companyID uuid.UUID) error {
	return u.db.Transaction(func(tx *gorm.DB) error {
		var factories []domain.Factory
		if companyID != uuid.Nil {
			_ = tx.Where("company_id = ?", companyID).Find(&factories).Error
		}
		if len(factories) == 0 {
			if err := tx.Find(&factories).Error; err != nil {
				return err
			}
		}

		repo := repository.NewFactoryRepository(tx)
		wRepo := repository.NewWarehouseRepository(tx)

		for _, factory := range factories {
			branchIDs, err := GetFactoryLinkedBranchIDs(tx, factory.ID)
			if err != nil {
				continue
			}

			stocks, err := repo.GetStocksByFactoryID(factory.ID)
			if err != nil {
				continue
			}

			for _, fs := range stocks {
				for _, branchID := range branchIDs {
					whStock, _ := wRepo.GetStock(branchID, fs.ProductID, "GOOD", fs.BatchNo)
					if whStock == nil || whStock.ID == uuid.Nil {
						whStock = &domain.WarehouseStock{
							ID:        uuid.New(),
							BranchID:  branchID,
							ProductID: fs.ProductID,
							BatchNo:   fs.BatchNo,
							StockType: "GOOD",
						}
					}
					whStock.Quantity = fs.Quantity
					whStock.ExpiryDate = fs.ExpiryDate
					_ = wRepo.UpdateStock(whStock)
				}
			}
		}
		return nil
	})
}

type ShipmentItem struct {
	ProductID uuid.UUID
	Quantity  int
}

func (u *factoryUsecase) RequestShipment(fromFactoryID, toBranchID uuid.UUID, items []domain.ProductTransferItem, notes string, targetShipmentDate, estimatedArrival *time.Time, initiatedBy string, vehicleID, driverID *uuid.UUID, paymentMethod string, paymentDueAt *time.Time, transferType string, isBillable *bool) error {
	var doNo string
	err := u.db.Transaction(func(tx *gorm.DB) error {
		repo := repository.NewFactoryRepository(tx)

		// 1. If Vehicle is assigned, check capacity
		if vehicleID != nil {
			var vehicle domain.Vehicle
			if err := tx.First(&vehicle, "id = ?", *vehicleID).Error; err != nil {
				return fmt.Errorf("kendaraan tidak ditemukan: %w", err)
			}

			totalRequestWeight := 0.0
			for _, item := range items {
				product, err := repo.GetProductByID(item.ProductID)
				if err != nil {
					return err
				}
				pcsPerUnit := item.PcsPerUnit
				if isBaseUnit(item.Unit) {
					pcsPerUnit = 1
				}
				weight := float64(item.Quantity) * float64(pcsPerUnit) * product.Weight
				unit := strings.ToUpper(product.WeightUnit)
				if unit == "GR" || unit == "ML" {
					weight /= 1000
				} else if unit == "TON" {
					weight *= 1000
				}
				totalRequestWeight += weight
			}

			if totalRequestWeight > vehicle.Capacity && vehicle.Capacity > 0 {
				return fmt.Errorf("tonase pengiriman (%.2f KG) melebihi kapasitas armada %s (%.2f KG)", totalRequestWeight, vehicle.Name, vehicle.Capacity)
			}
		}

		// 2. Generate a common DO number for this shipment (SJ - Surat Jalan)
		now := time.Now()
		doNo = fmt.Sprintf("SJ-%s-%s", now.Format("20060102"), strings.ToUpper(uuid.New().String()[:6]))

		// Determine transfer type and billing status
		finalTransferType := "COMMERCIAL"
		finalIsBillable := true
		if transferType == "INTERNAL_TRANSFER" || (isBillable != nil && !*isBillable) || paymentMethod == "INTERNAL" {
			finalTransferType = "INTERNAL_TRANSFER"
			finalIsBillable = false
			paymentMethod = "INTERNAL"
		} else if transferType != "" {
			finalTransferType = transferType
		}

		fmt.Printf("DEBUG: Processing Shipment Request with %d items (Type: %s, Billable: %v)\n", len(items), finalTransferType, finalIsBillable)
		for _, item := range items {
			// Get Product for weight calculation
			product, err := repo.GetProductByID(item.ProductID)
			if err != nil {
				return err
			}

			pcsPerUnit := item.PcsPerUnit
			if isBaseUnit(item.Unit) {
				pcsPerUnit = 1
			}

			totalWeight := float64(item.Quantity) * float64(pcsPerUnit) * product.Weight
			unit := strings.ToUpper(product.WeightUnit)
			if unit == "GR" || unit == "ML" {
				totalWeight /= 1000
			} else if unit == "TON" {
				totalWeight *= 1000
			}

			totalVolume := float64(item.Quantity) * product.Volume
			if totalVolume == 0 && product.Length > 0 && product.Width > 0 && product.Height > 0 {
				totalVolume = float64(item.Quantity) * (product.Length * product.Width * product.Height) / 1000000.0
			}

			status := "REQUESTED"

			batchNo := item.BatchNo
			if batchNo == "" {
				batchNo = "DEFAULT"
			}

			// Get Expiry Date from FactoryStock
			var expiryDate *time.Time
			stock, errStock := repo.GetStock(fromFactoryID, item.ProductID, batchNo)
			if errStock == nil && stock != nil {
				expiryDate = stock.ExpiryDate
			}

			totalAmount := float64(item.Quantity*pcsPerUnit) * product.FactoryPrice
			paymentStatus := "UNPAID"
			if !finalIsBillable {
				totalAmount = 0
				paymentStatus = "EXEMPT"
			}

			transfer := &domain.ProductTransfer{
				FromFactoryID:      fromFactoryID,
				ToBranchID:         toBranchID,
				ProductID:          item.ProductID,
				BatchNo:            batchNo,
				ExpiryDate:         expiryDate,
				Quantity:           item.Quantity,
				PcsPerUnit:         pcsPerUnit,
				TotalWeight:        totalWeight,
				TotalVolume:        totalVolume,
				Status:             status,
				DeliveryOrderNo:    doNo,
				Notes:              notes,
				TargetShipmentDate: targetShipmentDate,
				EstimatedArrival:   estimatedArrival,
				InitiatedBy:        initiatedBy,
				Unit:               item.Unit,
				VehicleID:          vehicleID,
				DriverID:           driverID,
				PaymentMethod:      paymentMethod,
				PaymentStatus:      paymentStatus,
				TotalAmount:        totalAmount,
				PaymentDueAt:       paymentDueAt,
				TransferType:       finalTransferType,
				IsBillable:         finalIsBillable,
			}
			if err := repo.CreateTransfer(transfer); err != nil {
				return err
			}
		}
		return nil
	})

	// Kirim notifikasi realtime SSE ke Tab Pabrik
	if err == nil && u.sseHub != nil {
		go func() {
			var toBranch domain.Branch
			branchName := "Cabang"
			if errB := u.db.Select("name").Where("id = ?", toBranchID).First(&toBranch).Error; errB == nil && toBranch.Name != "" {
				branchName = toBranch.Name
			}

			var factory domain.Factory
			companyIDStr := ""
			if errF := u.db.Select("company_id").Where("id = ?", fromFactoryID).First(&factory).Error; errF == nil {
				companyIDStr = factory.CompanyID.String()
			}

			event := sse.NotificationEvent{
				ID:        uuid.New().String(),
				Type:      "factory_shipment_request",
				Title:     "🏭 Permintaan Kirim Barang Baru",
				Message:   fmt.Sprintf("Gudang %s mengajukan permintaan pengiriman barang (%d item, No. SJ/Req: %s)", branchName, len(items), doNo),
				CompanyID: companyIDStr,
				CreatedAt: time.Now().Format(time.RFC3339),
				Data: map[string]interface{}{
					"from_factory_id": fromFactoryID.String(),
					"to_branch_id":   toBranchID.String(),
					"branch_name":    branchName,
					"do_number":      doNo,
					"item_count":     len(items),
					"notes":          notes,
				},
				TargetRoles: []string{
					string(domain.RoleSuperAdmin),
					string(domain.RoleAdminPabrik),
					string(domain.RoleManager),
				},
			}
			_ = u.sseHub.PublishNotification(context.Background(), event)
		}()
	}

	return err
}

func (u *factoryUsecase) ApproveTransfer(transferID uuid.UUID, approverID *uuid.UUID, approverName, approverPosition string) error {
	err := u.repo.UpdateTransferStatusWithApprover(transferID, "APPROVED", approverID, approverName, approverPosition)
	if err == nil && u.sseHub != nil {
		go func() {
			transfer, errT := u.repo.GetTransferByID(transferID)
			if errT == nil && transfer != nil {
				prodName := "Produk"
				if transfer.Product != nil {
					prodName = transfer.Product.Name
				}
				approverStr := approverName
				if approverStr == "" {
					approverStr = "Pabrik / Supervisor"
				}
				event := sse.NotificationEvent{
					ID:        transfer.ID.String() + "-approved",
					Type:      "warehouse_transfer_approved",
					Title:     "✅ Permintaan Pengiriman Disetujui",
					Message:   fmt.Sprintf("Permintaan pengiriman %s (%s, Qty: %d %s) telah disetujui oleh %s", transfer.DeliveryOrderNo, prodName, transfer.Quantity, transfer.Unit, approverStr),
					CreatedAt: time.Now().Format(time.RFC3339),
					Data: map[string]interface{}{
						"transfer_id": transfer.ID.String(),
						"do_number":   transfer.DeliveryOrderNo,
						"status":      "APPROVED",
						"approved_by": approverStr,
						"product_id":  transfer.ProductID.String(),
						"branch_id":   transfer.ToBranchID.String(),
					},
					TargetBranchIDs: []string{transfer.ToBranchID.String()},
					TargetRoles: []string{
						string(domain.RoleSuperAdmin),
						string(domain.RoleAdminGudang),
						string(domain.RoleSupervisorGudang),
						string(domain.RoleManager),
					},
				}
				_ = u.sseHub.PublishNotification(context.Background(), event)
			}
		}()
	}
	return err
}

func (u *factoryUsecase) ExecuteApprovedShipment(transferID uuid.UUID) error {
	err := u.db.Transaction(func(tx *gorm.DB) error {
		repo := repository.NewFactoryRepository(tx)

		// 1. Get Transfer
		transfer, err := repo.GetTransferByID(transferID)
		if err != nil {
			return err
		}
		if transfer.Status != "APPROVED" {
			return errors.New("only APPROVED transfers can be shipped")
		}

		// 2. Check Stock (Convert to total pieces)
		pcsPerUnit := transfer.PcsPerUnit
		if isBaseUnit(transfer.Unit) {
			pcsPerUnit = 1
		}
		totalPieces := transfer.Quantity * pcsPerUnit
		if transfer.BatchNo == "" {
			transfer.BatchNo = "DEFAULT"
		}

		if transfer.FulfillmentBranchID != nil {
			// Delegated Fulfillment: Deduct from Branch Stock
			wRepo := repository.NewWarehouseRepository(tx)
			stock, err := wRepo.GetStock(*transfer.FulfillmentBranchID, transfer.ProductID, "GOOD", transfer.BatchNo)
			if err != nil {
				return err
			}
			if stock.Quantity < totalPieces && (transfer.BatchNo == "DEFAULT" || transfer.BatchNo == "") {
				availBatches, errBatches := wRepo.GetAvailableBatches(*transfer.FulfillmentBranchID, transfer.ProductID, "GOOD")
				if errBatches == nil && len(availBatches) > 0 {
					for _, b := range availBatches {
						if b.Quantity >= totalPieces {
							stock = &b
							transfer.BatchNo = b.BatchNo
							break
						}
					}
				}
			}
			if stock.Quantity < totalPieces {
				return fmt.Errorf("stok cabang delegasi (Batch %s) tidak mencukupi (Butuh: %d, Tersedia: %d)", transfer.BatchNo, totalPieces, stock.Quantity)
			}
			stock.Quantity -= totalPieces
			if err := wRepo.UpdateStock(stock); err != nil {
				return err
			}

			// Mark as borrowed? No, create payment log instead for piutang transfer
			transfer.IsStockBorrowed = false

			// Create BranchPaymentLog for Piutang Transfer (Kompensasi Hutang)
			if transfer.TotalAmount > 0 {
				paymentLog := &domain.BranchPaymentLog{
					ProductTransferID: transfer.ID,
					BranchID:          *transfer.FulfillmentBranchID,
					AmountPaid:        transfer.TotalAmount,
					SourceOfFunds:     "DELEGATED_STOCK_TRANSFER",
					Notes:             fmt.Sprintf("Kompensasi pengalihan piutang untuk DO %s", transfer.DeliveryOrderNo),
				}
				if err := tx.Create(paymentLog).Error; err != nil {
					return err
				}
			}

			// Log Inventory for Branch
			log := &domain.WarehouseLog{
				BranchID:    *transfer.FulfillmentBranchID,
				ProductID:   transfer.ProductID,
				Type:        "OUT",
				Source:      "DELEGATED FACTORY TRANSFER",
				BatchNo:     transfer.BatchNo,
				Quantity:    totalPieces,
				ProcessedBy: "SISTEM",
			}
			wRepo.CreateLog(log)
		} else {
			// Normal: Deduct from Factory Stock
			stock, err := repo.GetStock(transfer.FromFactoryID, transfer.ProductID, transfer.BatchNo)
			if err != nil {
				return err
			}
			if stock.Quantity < totalPieces {
				product, errProd := repo.GetProductByID(transfer.ProductID)
				if errProd == nil && product != nil {
					return fmt.Errorf("stok produk '%s' tidak mencukupi (Butuh: %d, Tersedia: %d)", product.Name, totalPieces, stock.Quantity)
				}
				return errors.New("insufficient stock in factory for this shipment")
			}
			stock.Quantity -= totalPieces
			if stock.Quantity < 0 {
				stock.Quantity = 0
			}
			if err := repo.UpdateStock(stock); err != nil {
				return err
			}

			// Sync deduction to Warehouse Stock (Gudang Pabrik & Gudang Trenggalek 1:1)
			if err := syncFactoryStockDeltaToLinkedWarehouses(tx, transfer.FromFactoryID, transfer.ProductID, transfer.BatchNo, nil, -totalPieces, "FACTORY_TRANSFER: "+transfer.DeliveryOrderNo, "SYSTEM"); err != nil {
				return err
			}

			toBranchName := "Cabang"
			var toBranch domain.Branch
			if err := tx.First(&toBranch, "id = ?", transfer.ToBranchID).Error; err == nil && toBranch.Name != "" {
				toBranchName = toBranch.Name
			}
			transferSource := fmt.Sprintf("TRANSFER: %s (%s)", transfer.DeliveryOrderNo, toBranchName)
			if len(transferSource) > 255 {
				transferSource = transferSource[:255]
			}

			// Create Inventory Log for Factory
			inventoryLog := &domain.FactoryInventoryLog{
				FactoryID: transfer.FromFactoryID,
				ProductID: transfer.ProductID,
				Type:      "OUT",
				Source:    transferSource,
				BatchNo:   transfer.BatchNo,
				Quantity:  totalPieces,
			}
			repo.CreateInventoryLog(inventoryLog)
		}

		// 4. Update Status to SHIPPED. Keep the existing SJ number if present.
		now := time.Now()
		transfer.Status = "SHIPPED"
		transfer.ShippedAt = &now
		if transfer.DeliveryOrderNo == "" {
			transfer.DeliveryOrderNo = fmt.Sprintf("SJ-%s-%s", now.Format("20060102"), strings.ToUpper(uuid.New().String()[:6]))
		}

		if err := repo.UpdateTransfer(transfer); err != nil {
			return err
		}

		return nil
	})

	if err == nil && u.sseHub != nil {
		go func() {
			transfer, errT := u.repo.GetTransferByID(transferID)
			if errT == nil && transfer != nil {
				prodName := "Produk"
				if transfer.Product != nil {
					prodName = transfer.Product.Name
				}
				event := sse.NotificationEvent{
					ID:        transfer.ID.String() + "-shipped",
					Type:      "warehouse_transfer_shipped",
					Title:     "🚚 Barang Sedang Dikirim dari Pabrik",
					Message:   fmt.Sprintf("Pengiriman %s (%s, Qty: %d %s) sedang dalam perjalanan menuju gudang Anda", transfer.DeliveryOrderNo, prodName, transfer.Quantity, transfer.Unit),
					CreatedAt: time.Now().Format(time.RFC3339),
					Data: map[string]interface{}{
						"transfer_id": transfer.ID.String(),
						"do_number":   transfer.DeliveryOrderNo,
						"status":      "SHIPPED",
						"branch_id":   transfer.ToBranchID.String(),
					},
					TargetBranchIDs: []string{transfer.ToBranchID.String()},
					TargetRoles: []string{
						string(domain.RoleSuperAdmin),
						string(domain.RoleAdminGudang),
						string(domain.RoleSupervisorGudang),
						string(domain.RoleManager),
					},
				}
				_ = u.sseHub.PublishNotification(context.Background(), event)
			}
		}()
	}

	return err
}

func (u *factoryUsecase) GetTransferHistory(factoryID uuid.UUID) ([]domain.ProductTransfer, error) {
	return u.repo.GetTransfersByFactoryID(factoryID)
}

func (u *factoryUsecase) GetAllTransfers(companyID uuid.UUID, productID string, batchNo string) ([]domain.ProductTransfer, error) {
	return u.repo.GetAllTransfersByCompanyID(companyID, productID, batchNo)
}

func (u *factoryUsecase) GetAllTransfersPaginated(page, limit int, companyID uuid.UUID, branchID, status, search string) ([]domain.ProductTransfer, int64, float64, map[string]int64, error) {
	return u.repo.GetAllTransfersPaginated(page, limit, companyID, branchID, status, search)
}

func (u *factoryUsecase) UpdateTransfer(id uuid.UUID, data map[string]interface{}) error {
	transfer, err := u.repo.GetTransferByID(id)
	if err != nil {
		return err
	}
	if transfer.Status != "REQUESTED" && transfer.Status != "APPROVED" && transfer.Status != "LOADING" && transfer.Status != "REJECTED" {
		return errors.New("pengiriman yang sudah diproses tidak dapat diubah")
	}

	// Update only allowed fields directly on the ProductTransfer model
	if err := u.db.Model(&domain.ProductTransfer{}).Where("id = ?", id).Updates(data).Error; err != nil {
		return err
	}

	// If quantity, pcs_per_unit, product, or transfer_type changed, recalculate total_weight and total_amount
	if _, ok := data["quantity"]; ok || data["pcs_per_unit"] != nil || data["unit"] != nil || data["product_id"] != nil || data["transfer_type"] != nil || data["is_billable"] != nil || data["payment_method"] != nil {
		// Re-fetch to get latest values
		updated, _ := u.repo.GetTransferByID(id)
		product, _ := u.repo.GetProductByID(updated.ProductID)

		if isBaseUnit(updated.Unit) {
			updated.PcsPerUnit = 1
		}
		totalWeight := float64(updated.Quantity) * float64(updated.PcsPerUnit) * product.Weight
		unit := strings.ToUpper(product.WeightUnit)
		if unit == "GR" || unit == "ML" {
			totalWeight /= 1000
		} else if unit == "TON" {
			totalWeight *= 1000
		}
		updated.TotalWeight = totalWeight
		if updated.TransferType == "INTERNAL_TRANSFER" || !updated.IsBillable || updated.PaymentMethod == "INTERNAL" {
			updated.TotalAmount = 0
			updated.TransferType = "INTERNAL_TRANSFER"
			updated.IsBillable = false
			updated.PaymentMethod = "INTERNAL"
			updated.PaymentStatus = "EXEMPT"
		} else {
			updated.TotalAmount = float64(updated.Quantity*updated.PcsPerUnit) * product.FactoryPrice
		}
		if errUpdate := u.repo.UpdateTransfer(updated); errUpdate != nil {
			return errUpdate
		}
	}

	// Kirim notifikasi realtime SSE ke Gudang Cabang pemohon bahwa request telah disesuaikan / di-adjust
	if u.sseHub != nil {
		go func() {
			updated, errT := u.repo.GetTransferByID(id)
			if errT == nil && updated != nil {
				prodName := "Produk"
				if updated.Product != nil {
					prodName = updated.Product.Name
				}
				event := sse.NotificationEvent{
					ID:        updated.ID.String() + "-adjusted-" + fmt.Sprint(time.Now().Unix()),
					Type:      "warehouse_transfer_adjusted",
					Title:     "✏️ Permintaan Pengiriman Disesuaikan",
					Message:   fmt.Sprintf("Permintaan pengiriman %s (%s) telah disesuaikan: Qty %d %s", updated.DeliveryOrderNo, prodName, updated.Quantity, updated.Unit),
					CreatedAt: time.Now().Format(time.RFC3339),
					Data: map[string]interface{}{
						"transfer_id": updated.ID.String(),
						"do_number":   updated.DeliveryOrderNo,
						"quantity":    updated.Quantity,
						"unit":        updated.Unit,
						"branch_id":   updated.ToBranchID.String(),
					},
					TargetBranchIDs: []string{updated.ToBranchID.String()},
					TargetRoles: []string{
						string(domain.RoleSuperAdmin),
						string(domain.RoleAdminGudang),
						string(domain.RoleSupervisorGudang),
						string(domain.RoleManager),
					},
				}
				_ = u.sseHub.PublishNotification(context.Background(), event)
			}
		}()
	}

	return nil
}

type UpdateTransferDOInput struct {
	FromFactoryID      uuid.UUID                    `json:"from_factory_id"`
	ToBranchID         uuid.UUID                    `json:"to_branch_id"`
	VehicleID          *uuid.UUID                   `json:"vehicle_id"`
	DriverID           *uuid.UUID                   `json:"driver_id"`
	TargetShipmentDate *time.Time                   `json:"target_shipment_date"`
	EstimatedArrival   *time.Time                   `json:"estimated_arrival"`
	PaymentMethod      string                       `json:"payment_method"`
	PaymentDueAt       *time.Time                   `json:"payment_due_at"`
	TransferType       string                       `json:"transfer_type"`
	IsBillable         *bool                        `json:"is_billable"`
	Notes              string                       `json:"notes"`
	Items              []domain.ProductTransferItem `json:"items" binding:"required"`
}

func (u *factoryUsecase) UpdateTransferByDO(doNo string, input UpdateTransferDOInput) error {
	existingTransfers, err := u.repo.GetTransfersByDO(doNo)
	if err != nil || len(existingTransfers) == 0 {
		return errors.New("surat jalan / faktur pengiriman cabang tidak ditemukan")
	}

	for _, t := range existingTransfers {
		if t.Status != "REQUESTED" && t.Status != "APPROVED" && t.Status != "LOADING" && t.Status != "REJECTED" {
			return errors.New("pengiriman yang sudah diproses atau terkirim tidak dapat diubah")
		}
	}

	if len(input.Items) == 0 {
		return errors.New("faktur pengiriman harus memiliki setidaknya satu item produk")
	}

	return u.db.Transaction(func(tx *gorm.DB) error {
		txRepo := repository.NewFactoryRepository(tx)

		// 1. Capacity check if vehicle is specified
		if input.VehicleID != nil && *input.VehicleID != uuid.Nil {
			var vehicle domain.Vehicle
			if err := tx.First(&vehicle, "id = ?", *input.VehicleID).Error; err == nil && vehicle.Capacity > 0 {
				totalWeight := 0.0
				for _, item := range input.Items {
					product, errP := txRepo.GetProductByID(item.ProductID)
					if errP == nil && product != nil {
						pcsPerUnit := item.PcsPerUnit
						if isBaseUnit(item.Unit) {
							pcsPerUnit = 1
						}
						w := float64(item.Quantity) * float64(pcsPerUnit) * product.Weight
						unit := strings.ToUpper(product.WeightUnit)
						if unit == "GR" || unit == "ML" {
							w /= 1000
						} else if unit == "TON" {
							w *= 1000
						}
						totalWeight += w
					}
				}
				if totalWeight > vehicle.Capacity {
					return fmt.Errorf("tonase pengiriman (%.2f KG) melebihi kapasitas armada %s (%.2f KG)", totalWeight, vehicle.Name, vehicle.Capacity)
				}
			}
		}

		// 2. Determine transfer type and billing status
		finalTransferType := "COMMERCIAL"
		finalIsBillable := true
		if input.TransferType == "INTERNAL_TRANSFER" || (input.IsBillable != nil && !*input.IsBillable) || input.PaymentMethod == "INTERNAL" {
			finalTransferType = "INTERNAL_TRANSFER"
			finalIsBillable = false
			input.PaymentMethod = "INTERNAL"
		} else if input.TransferType != "" {
			finalTransferType = input.TransferType
		}

		paymentStatus := "UNPAID"
		if !finalIsBillable {
			paymentStatus = "EXEMPT"
		}

		commonStatus := existingTransfers[0].Status
		commonInitiatedBy := existingTransfers[0].InitiatedBy

		// Map existing transfers by item ID for quick lookup
		existingMap := make(map[uuid.UUID]*domain.ProductTransfer)
		for i := range existingTransfers {
			existingMap[existingTransfers[i].ID] = &existingTransfers[i]
		}

		var retainedIDs []uuid.UUID

		for _, item := range input.Items {
			product, errP := txRepo.GetProductByID(item.ProductID)
			if errP != nil || product == nil {
				return fmt.Errorf("produk dengan ID %s tidak ditemukan", item.ProductID)
			}

			pcsPerUnit := item.PcsPerUnit
			if isBaseUnit(item.Unit) {
				pcsPerUnit = 1
			}

			totalWeight := float64(item.Quantity) * float64(pcsPerUnit) * product.Weight
			unit := strings.ToUpper(product.WeightUnit)
			if unit == "GR" || unit == "ML" {
				totalWeight /= 1000
			} else if unit == "TON" {
				totalWeight *= 1000
			}

			totalVolume := float64(item.Quantity) * product.Volume
			if totalVolume == 0 && product.Length > 0 && product.Width > 0 && product.Height > 0 {
				totalVolume = float64(item.Quantity) * (product.Length * product.Width * product.Height) / 1000000.0
			}

			batchNo := item.BatchNo
			if batchNo == "" {
				batchNo = "DEFAULT"
			}

			var expiryDate *time.Time
			stock, errStock := txRepo.GetStock(input.FromFactoryID, item.ProductID, batchNo)
			if errStock == nil && stock != nil {
				expiryDate = stock.ExpiryDate
			}

			totalAmount := float64(item.Quantity*pcsPerUnit) * product.FactoryPrice
			if !finalIsBillable {
				totalAmount = 0
			}

			if item.ID != nil && *item.ID != uuid.Nil {
				if existing, exists := existingMap[*item.ID]; exists {
					existing.FromFactoryID = input.FromFactoryID
					existing.ToBranchID = input.ToBranchID
					existing.ProductID = item.ProductID
					existing.BatchNo = batchNo
					existing.ExpiryDate = expiryDate
					existing.Quantity = item.Quantity
					existing.Unit = item.Unit
					existing.PcsPerUnit = pcsPerUnit
					existing.TotalWeight = totalWeight
					existing.TotalVolume = totalVolume
					existing.TotalAmount = totalAmount
					existing.PaymentMethod = input.PaymentMethod
					existing.PaymentStatus = paymentStatus
					existing.PaymentDueAt = input.PaymentDueAt
					existing.TransferType = finalTransferType
					existing.IsBillable = finalIsBillable
					existing.VehicleID = input.VehicleID
					existing.DriverID = input.DriverID
					existing.TargetShipmentDate = input.TargetShipmentDate
					existing.EstimatedArrival = input.EstimatedArrival
					existing.Notes = input.Notes

					if errU := txRepo.UpdateTransfer(existing); errU != nil {
						return errU
					}
					retainedIDs = append(retainedIDs, existing.ID)
					continue
				}
			}

			// New item added during edit
			newTransfer := &domain.ProductTransfer{
				FromFactoryID:      input.FromFactoryID,
				ToBranchID:         input.ToBranchID,
				ProductID:          item.ProductID,
				BatchNo:            batchNo,
				ExpiryDate:         expiryDate,
				Quantity:           item.Quantity,
				PcsPerUnit:         pcsPerUnit,
				TotalWeight:        totalWeight,
				TotalVolume:        totalVolume,
				Status:             commonStatus,
				DeliveryOrderNo:    doNo,
				Notes:              input.Notes,
				TargetShipmentDate: input.TargetShipmentDate,
				EstimatedArrival:   input.EstimatedArrival,
				InitiatedBy:        commonInitiatedBy,
				Unit:               item.Unit,
				VehicleID:          input.VehicleID,
				DriverID:           input.DriverID,
				PaymentMethod:      input.PaymentMethod,
				PaymentStatus:      paymentStatus,
				TotalAmount:        totalAmount,
				PaymentDueAt:       input.PaymentDueAt,
				TransferType:       finalTransferType,
				IsBillable:         finalIsBillable,
			}
			if errC := txRepo.CreateTransfer(newTransfer); errC != nil {
				return errC
			}
			retainedIDs = append(retainedIDs, newTransfer.ID)
		}

		// Delete removed items
		if errD := txRepo.DeleteTransfersByDOAndExcludeIDs(doNo, retainedIDs); errD != nil {
			return errD
		}

		// Send realtime SSE notification
		if u.sseHub != nil {
			go func() {
				event := sse.NotificationEvent{
					ID:        doNo + "-updated-" + fmt.Sprint(time.Now().Unix()),
					Type:      "warehouse_transfer_updated",
					Title:     "✏️ Faktur Pengiriman Cabang Diperbarui",
					Message:   fmt.Sprintf("Faktur pengiriman %s telah diperbarui.", doNo),
					CreatedAt: time.Now().Format(time.RFC3339),
					Data: map[string]interface{}{
						"do_number": doNo,
						"branch_id": input.ToBranchID.String(),
					},
					TargetBranchIDs: []string{input.ToBranchID.String()},
					TargetRoles: []string{
						string(domain.RoleSuperAdmin),
						string(domain.RoleAdminGudang),
						string(domain.RoleSupervisorGudang),
						string(domain.RoleManager),
					},
				}
				_ = u.sseHub.PublishNotification(context.Background(), event)
			}()
		}

		return nil
	})
}

func (u *factoryUsecase) DeleteTransfer(id uuid.UUID) error {
	transfer, err := u.repo.GetTransferByID(id)
	if err != nil {
		return err
	}
	if transfer.Status != "REQUESTED" && transfer.Status != "APPROVED" && transfer.Status != "REJECTED" {
		return errors.New("pengiriman yang sudah diproses tidak dapat dihapus")
	}

	return u.repo.DeleteTransfer(id)
}

func (u *factoryUsecase) CreateTransfer(transfer *domain.ProductTransfer) error {
	product, err := u.repo.GetProductByID(transfer.ProductID)
	if err == nil && product != nil {
		if isBaseUnit(transfer.Unit) {
			transfer.PcsPerUnit = 1
		}

		totalWeight := float64(transfer.Quantity) * float64(transfer.PcsPerUnit) * product.Weight
		unit := strings.ToUpper(product.WeightUnit)
		if unit == "GR" || unit == "ML" {
			totalWeight /= 1000
		} else if unit == "TON" {
			totalWeight *= 1000
		}
		transfer.TotalWeight = totalWeight

		totalVolume := float64(transfer.Quantity) * product.Volume
		if totalVolume == 0 && product.Length > 0 && product.Width > 0 && product.Height > 0 {
			totalVolume = float64(transfer.Quantity) * (product.Length * product.Width * product.Height) / 1000000.0
		}
		transfer.TotalVolume = totalVolume
		if transfer.TransferType == "INTERNAL_TRANSFER" || !transfer.IsBillable || transfer.PaymentMethod == "INTERNAL" {
			transfer.TotalAmount = 0
			transfer.TransferType = "INTERNAL_TRANSFER"
			transfer.IsBillable = false
			transfer.PaymentMethod = "INTERNAL"
			transfer.PaymentStatus = "EXEMPT"
		} else {
			if transfer.TransferType == "" {
				transfer.TransferType = "COMMERCIAL"
			}
			transfer.IsBillable = true
			transfer.TotalAmount = float64(transfer.Quantity*transfer.PcsPerUnit) * product.FactoryPrice
		}
	}

	return u.repo.CreateTransfer(transfer)
}

func (u *factoryUsecase) GetBranchReceivables(companyID uuid.UUID) (*domain.FactoryReceivablesResponse, error) {
	branches, err := u.repo.GetBranchReceivables(companyID)
	if err != nil {
		return nil, err
	}

	response := &domain.FactoryReceivablesResponse{
		Branches: branches,
	}

	for _, b := range branches {
		response.Summary.TotalReceivables += b.TotalReceivable
		response.Summary.TotalGoodReceivables += b.GoodReceivable
		response.Summary.TotalGoodInvoiced += b.TotalGoodInvoiced
		response.Summary.TotalDamagedDeductions += b.TotalDamagedDeduction
		response.Summary.TotalReturnDeductions += b.TotalReturnDeduction
		response.Summary.TotalFinancialReceivables += b.FinancialReceivable
		response.Summary.TotalPaymentsThisMonth += b.TotalPaymentThisMonth
		if b.TotalReceivable > 0 {
			response.Summary.UnpaidBranchesCount++
		}
	}

	return response, nil
}

func (u *factoryUsecase) GetDelegatedShipments(companyID uuid.UUID) ([]domain.ProductTransfer, []DelegatedShipmentSummary, error) {
	transfers, err := u.repo.GetAllTransfersByCompanyID(companyID, "", "")
	if err != nil {
		return nil, nil, err
	}

	var delegated []domain.ProductTransfer
	summaryMap := make(map[uuid.UUID]*DelegatedShipmentSummary)

	for _, t := range transfers {
		if t.FulfillmentBranchID != nil {
			delegated = append(delegated, t)
			if t.IsStockBorrowed && t.ReplenishmentTransferID == nil {
				pcsPerUnit := t.PcsPerUnit
				if isBaseUnit(t.Unit) {
					pcsPerUnit = 1
				}
				totalPieces := t.Quantity * pcsPerUnit

				if s, exists := summaryMap[t.ProductID]; exists {
					s.TotalPieces += totalPieces
				} else {
					productName := "Unknown Product"
					if t.Product != nil {
						productName = t.Product.Name
					}
					summaryMap[t.ProductID] = &DelegatedShipmentSummary{
						ProductID:   t.ProductID,
						ProductName: productName,
						TotalPieces: totalPieces,
					}
				}
			}
		}
	}

	var summary []DelegatedShipmentSummary
	for _, s := range summaryMap {
		summary = append(summary, *s)
	}

	return delegated, summary, nil
}

func (u *factoryUsecase) GetBackorderDemand(companyID uuid.UUID) ([]BackorderDemand, error) {
	var results []BackorderDemand
	err := u.db.Table("sales_order_items").
		Select("sales_order_items.product_id, products.name as product_name, SUM(sales_order_items.quantity) as total_qty, COUNT(DISTINCT sales_order_items.sales_order_id) as order_count").
		Joins("JOIN sales_orders ON sales_orders.id = sales_order_items.sales_order_id").
		Joins("JOIN products ON products.id = sales_order_items.product_id").
		Where("sales_orders.company_id = ? AND sales_orders.status = ?", companyID, domain.SOStatusWaitingStock).
		Group("sales_order_items.product_id, products.name").
		Having("SUM(sales_order_items.quantity) > 0").
		Order("total_qty DESC").
		Scan(&results).Error
	return results, err
}

func (u *factoryUsecase) GetDashboardStats(companyID uuid.UUID) (*domain.FactoryDashboardStats, error) {
	return u.repo.GetDashboardStats(companyID)
}

func (u *factoryUsecase) GetLowStockAlerts(companyID uuid.UUID) ([]domain.LowStockAlert, error) {
	return u.repo.GetLowStockAlerts(companyID)
}

func isBaseUnit(unit string) bool {
	u := strings.ToUpper(strings.TrimSpace(unit))
	return u == "PCS" || u == "BOTOL" || u == "SACHET" || u == "JERIGEN" || u == "PACK" || u == "GRAM" || u == "KG"
}

func convertQuantityByUnit(qty int, fromUnit, toUnit string, pcsPerUnit int) int {
	from := strings.ToUpper(strings.TrimSpace(fromUnit))
	to := strings.ToUpper(strings.TrimSpace(toUnit))
	if from == "" || to == "" || from == to {
		return qty
	}
	// Weight conversions (base: GRAM)
	weightToGram := map[string]float64{
		"MG": 0.001, "G": 1, "GR": 1, "GRAM": 1, "ONS": 100, "HG": 100, "KG": 1000, "KILOGRAM": 1000, "KWINTAL": 100000, "TON": 1000000,
	}
	if fromF, ok1 := weightToGram[from]; ok1 {
		if toF, ok2 := weightToGram[to]; ok2 {
			inGram := float64(qty) * fromF
			return int(math.Round(inGram / toF))
		}
	}
	// Volume conversions (base: ML)
	volumeToMl := map[string]float64{
		"ML": 1, "MILLILITER": 1, "CL": 10, "DL": 100, "L": 1000, "LTR": 1000, "LITER": 1000, "GALON": 19000, "JERIGEN": 20000, "DRUM": 200000,
	}
	if fromF, ok1 := volumeToMl[from]; ok1 {
		if toF, ok2 := volumeToMl[to]; ok2 {
			inMl := float64(qty) * fromF
			return int(math.Round(inMl / toF))
		}
	}
	// Count conversions (base: PCS)
	if pcsPerUnit <= 0 {
		pcsPerUnit = 24
	}
	countToPcs := func(u string) (float64, bool) {
		switch u {
		case "PCS", "BUAH", "BIJI", "BOTOL", "LEMBAR", "TUTUP", "ROLL", "PACK", "BOX":
			return 1, true
		case "LUSIN", "DOZEN":
			return 12, true
		case "KODI":
			return 20, true
		case "GROSS":
			return 144, true
		case "DUS", "KARTON", "BALL", "SAK":
			return float64(pcsPerUnit), true
		default:
			return 1, false
		}
	}
	if fromF, ok1 := countToPcs(from); ok1 {
		if toF, ok2 := countToPcs(to); ok2 {
			inPcs := float64(qty) * fromF
			return int(math.Round(inPcs / toF))
		}
	}
	return qty
}

func (u *factoryUsecase) GetRawMaterialStock(factoryID uuid.UUID) ([]domain.FactoryStock, error) {
	return u.repo.GetRawMaterialStockByFactoryID(factoryID)
}

func (u *factoryUsecase) GetRawMaterialLogs(factoryID uuid.UUID, productID *uuid.UUID) ([]domain.FactoryInventoryLog, error) {
	return u.repo.GetRawMaterialLogsByFactoryID(factoryID, productID)
}

func (u *factoryUsecase) MutateRawMaterial(factoryID uuid.UUID, productID uuid.UUID, mutationType string, quantity int, notes string, batchNo string, targetWarehouseID *uuid.UUID) error {
	batchNo = strings.TrimSpace(batchNo)
	if batchNo == "" {
		batchNo = fmt.Sprintf("RM-%s", time.Now().Format("060102-150405"))
	}

	return u.db.Transaction(func(tx *gorm.DB) error {
		txRepo := repository.NewFactoryRepository(tx)

		// Get or create stock
		stock, err := txRepo.GetStock(factoryID, productID, batchNo)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				stock = &domain.FactoryStock{
					FactoryID:          factoryID,
					ProductID:          productID,
					FactoryWarehouseID: targetWarehouseID,
					BatchNo:            batchNo,
					Quantity:           0,
				}
				if err := tx.Create(stock).Error; err != nil {
					return err
				}
			} else {
				return err
			}
		}

		delta := 0
		if mutationType == "IN" {
			delta = quantity
			stock.Quantity += delta
		} else if mutationType == "OUT" {
			delta = -quantity
			if stock.Quantity < quantity {
				return errors.New("insufficient raw material stock")
			}
			stock.Quantity -= quantity
		} else if mutationType == "ADJUSTMENT" {
			delta = quantity - stock.Quantity
			stock.Quantity = quantity
		} else {
			return errors.New("invalid mutation type")
		}

		if targetWarehouseID != nil {
			stock.FactoryWarehouseID = targetWarehouseID
		}

		if err := txRepo.UpdateStock(stock); err != nil {
			return err
		}

		logEntry := &domain.FactoryInventoryLog{
			FactoryID: factoryID,
			ProductID: productID,
			Type:      mutationType,
			Source:    "RAW_MATERIAL_MUTATION",
			BatchNo:   batchNo,
			Quantity:  delta,
		}

		if err := txRepo.CreateInventoryLog(logEntry); err != nil {
			return err
		}

		return nil
	})
}

func (u *factoryUsecase) DeleteStock(id uuid.UUID) error {
	return u.db.Transaction(func(tx *gorm.DB) error {
		var stock domain.FactoryStock
		if err := tx.First(&stock, "id = ?", id).Error; err != nil {
			return err
		}
		// Preserve an auditable movement before deleting the saldo row.
		if stock.Quantity != 0 {
			if err := tx.Create(&domain.FactoryInventoryLog{
				FactoryID: stock.FactoryID,
				ProductID: stock.ProductID,
				Type:      "OUT",
				Source:    "STOCK_DELETE",
				BatchNo:   stock.BatchNo,
				Quantity:  stock.Quantity,
			}).Error; err != nil {
				return err
			}
		}
		return tx.Delete(&domain.FactoryStock{}, "id = ?", id).Error
	})
}

func (u *factoryUsecase) DeleteInventoryLog(id uuid.UUID) error {
	return u.repo.DeleteInventoryLog(id)
}

// --- QC Sample Implementations ---

func (u *factoryUsecase) RequestQcSample(input RequestQcSampleInput) (*domain.FactoryQcSample, error) {
	if input.Quantity <= 0 {
		return nil, errors.New("jumlah sampel minimal 1 pcs")
	}
	if input.BatchNo == "" {
		return nil, errors.New("nomor batch produksi wajib diisi")
	}

	sample := &domain.FactoryQcSample{
		FactoryID:       input.FactoryID,
		ProductID:       input.ProductID,
		ProductionLogID: input.ProductionLogID,
		BatchNo:         input.BatchNo,
		Quantity:        input.Quantity,
		SampleType:      input.SampleType,
		Purpose:         input.Purpose,
		Notes:           input.Notes,
		Status:          "PENDING",
		RequestedByID:   input.RequestedByID,
		RequestedByName: input.RequestedByName,
		RequestedAt:     time.Now(),
	}

	if sample.SampleType == "" {
		sample.SampleType = "LAB_QC"
	}

	// Ambil metadata tanggal jika ditautkan ke production log
	if input.ProductionLogID != nil && *input.ProductionLogID != uuid.Nil {
		prodLog, err := u.repo.GetProductionLogByID(*input.ProductionLogID)
		if err == nil && prodLog != nil {
			sample.ExpiryDate = prodLog.ExpiryDate
			sample.ProductionDate = &prodLog.ProductionDate
		}
	}

	// Jika belum ada expiry date, cari dari stock
	if sample.ExpiryDate == nil {
		stock, err := u.repo.GetStock(input.FactoryID, input.ProductID, input.BatchNo)
		if err == nil && stock != nil {
			sample.ExpiryDate = stock.ExpiryDate
		}
	}

	if err := u.repo.CreateQcSample(sample); err != nil {
		return nil, err
	}

	return sample, nil
}

func (u *factoryUsecase) GetQcSamples(factoryID uuid.UUID, status string) ([]domain.FactoryQcSample, error) {
	return u.repo.GetQcSamplesByFactoryID(factoryID, status)
}

func (u *factoryUsecase) GetAllQcSamples(companyID uuid.UUID, status string) ([]domain.FactoryQcSample, error) {
	return u.repo.GetAllQcSamplesByCompanyID(companyID, status)
}

func (u *factoryUsecase) GetQcSampleDetail(id uuid.UUID) (*domain.FactoryQcSample, error) {
	return u.repo.GetQcSampleByID(id)
}

func (u *factoryUsecase) ApproveQcSample(id uuid.UUID, approverID *uuid.UUID, approverName string) error {
	return u.db.Transaction(func(tx *gorm.DB) error {
		txRepo := repository.NewFactoryRepository(tx)

		sample, err := txRepo.GetQcSampleByID(id)
		if err != nil {
			return fmt.Errorf("sampel tidak ditemukan: %w", err)
		}

		if sample.Status != "PENDING" {
			return fmt.Errorf("status sampel sudah '%s', tidak dapat di-approve", sample.Status)
		}

		// Ambil stok dari batch terkait
		stock, err := txRepo.GetStock(sample.FactoryID, sample.ProductID, sample.BatchNo)
		if err != nil {
			return fmt.Errorf("stok batch %s tidak ditemukan: %w", sample.BatchNo, err)
		}

		if stock.Quantity < sample.Quantity {
			return fmt.Errorf("stok batch %s tidak mencukupi (Tersedia: %d, Diminta: %d)", sample.BatchNo, stock.Quantity, sample.Quantity)
		}

		// Kurangi stok di pabrik
		stock.Quantity -= sample.Quantity
		if err := txRepo.UpdateStock(stock); err != nil {
			return fmt.Errorf("gagal memotong stok pabrik: %w", err)
		}

		// Sinkronkan ke gudang linked
		sourceDesc := fmt.Sprintf("SAMPLE_QC_LAB (Batch: %s)", sample.BatchNo)
		if err := syncFactoryStockDeltaToLinkedWarehouses(tx, sample.FactoryID, sample.ProductID, sample.BatchNo, stock.ExpiryDate, -sample.Quantity, sourceDesc, approverName); err != nil {
			return fmt.Errorf("gagal sinkronisasi pengurangan stok gudang: %w", err)
		}

		// Catat log mutasi inventori pabrik
		invLog := &domain.FactoryInventoryLog{
			FactoryID: sample.FactoryID,
			ProductID: sample.ProductID,
			Type:      "OUT",
			Source:    sourceDesc,
			BatchNo:   sample.BatchNo,
			Quantity:  sample.Quantity,
		}
		if err := txRepo.CreateInventoryLog(invLog); err != nil {
			return fmt.Errorf("gagal mencatat log mutasi inventori: %w", err)
		}

		// Update data status sampel
		now := time.Now()
		sample.Status = "APPROVED"
		sample.ApprovedByID = approverID
		sample.ApprovedByName = approverName
		sample.ApprovedAt = &now

		if err := txRepo.UpdateQcSample(sample); err != nil {
			return fmt.Errorf("gagal mengupdate status sampel: %w", err)
		}

		return nil
	})
}

func (u *factoryUsecase) RejectQcSample(id uuid.UUID, approverID *uuid.UUID, approverName string, reason string) error {
	return u.db.Transaction(func(tx *gorm.DB) error {
		txRepo := repository.NewFactoryRepository(tx)

		sample, err := txRepo.GetQcSampleByID(id)
		if err != nil {
			return fmt.Errorf("sampel tidak ditemukan: %w", err)
		}

		if sample.Status != "PENDING" {
			return fmt.Errorf("status sampel sudah '%s', tidak dapat ditolak", sample.Status)
		}

		now := time.Now()
		sample.Status = "REJECTED"
		sample.ApprovedByID = approverID
		sample.ApprovedByName = approverName
		sample.ApprovedAt = &now
		sample.RejectionReason = reason

		if err := txRepo.UpdateQcSample(sample); err != nil {
			return fmt.Errorf("gagal mengupdate status sampel: %w", err)
		}

		return nil
	})
}

