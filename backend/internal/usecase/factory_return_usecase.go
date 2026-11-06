package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"example.com/hris/backend/internal/repository"
	"example.com/hris/backend/pkg/sse"
	"gorm.io/gorm"
)

type FactoryReturnUsecase interface {
	CreateReturn(req CreateFactoryReturnRequest) (*domain.FactoryReturn, error)
	ReceiveReturn(id uuid.UUID, req *ReceiveFactoryReturnRequest) error
	RejectReturn(id uuid.UUID, rejectorID *uuid.UUID, rejectorName string, reason string) error
	GetByBranch(branchID uuid.UUID) ([]domain.FactoryReturn, error)
	GetByFactory(factoryID uuid.UUID) ([]domain.FactoryReturn, error)
	GetByCompany(companyID uuid.UUID) ([]domain.FactoryReturn, error)
}

type factoryReturnUsecase struct {
	returnRepo repository.FactoryReturnRepository
	db         *gorm.DB
	sseHub     *sse.Hub
}

func NewFactoryReturnUsecase(returnRepo repository.FactoryReturnRepository, db *gorm.DB, sseHub *sse.Hub) FactoryReturnUsecase {
	return &factoryReturnUsecase{returnRepo, db, sseHub}
}

type CreateFactoryReturnRequest struct {
	FromBranchID uuid.UUID              `json:"from_branch_id" binding:"required"`
	ToFactoryID  uuid.UUID              `json:"to_factory_id" binding:"required"`
	ReturnDate   *time.Time             `json:"return_date"`
	Notes        string                 `json:"notes"`
	VehicleID    *uuid.UUID             `json:"vehicle_id"`
	DriverID     *uuid.UUID             `json:"driver_id"`
	Items        []FactoryReturnItemReq `json:"items" binding:"required,min=1"`
}

type FactoryReturnItemReq struct {
	ProductID    uuid.UUID  `json:"product_id" binding:"required"`
	Quantity     int        `json:"quantity" binding:"required,min=1"`
	BatchNo      string     `json:"batch_no"`
	ReturnType   string     `json:"return_type"` // GOOD, DAMAGED
	DamageType   string     `json:"damage_type"` // BOTTLE_PACKAGING, PRODUCT_CONTENT, EXPIRED, TOTAL_DAMAGE, OTHER
	DamageReason string     `json:"damage_reason"`
	ExpiryDate   *time.Time `json:"expiry_date"`
}

type ReceiveFactoryReturnRequest struct {
	ReceiverName    string                        `json:"receiver_name"`
	InspectedByName string                        `json:"inspected_by_name"`
	InspectionNotes string                        `json:"inspection_notes"`
	Items           []ReceiveFactoryReturnItemReq `json:"items"`
}

type ReceiveFactoryReturnItemReq struct {
	ItemID                    uuid.UUID  `json:"item_id"`
	ReceivedGoodQty           int        `json:"received_good_qty"`
	ReceivedBottleDamagedQty  int        `json:"received_bottle_damaged_qty"`
	ReceivedContentDamagedQty int        `json:"received_content_damaged_qty"`
	DispositionAction         string     `json:"disposition_action"` // RESTOCK_GOOD, REPACK_BOTTLE, SCRAP_DISPOSAL, PARTIAL_MIXED
	InspectionNotes           string     `json:"inspection_notes"`
	TargetFactoryWarehouseID  *uuid.UUID `json:"target_factory_warehouse_id"`
}

func (u *factoryReturnUsecase) CreateReturn(req CreateFactoryReturnRequest) (*domain.FactoryReturn, error) {
	var createdReturn *domain.FactoryReturn

	err := u.db.Transaction(func(tx *gorm.DB) error {
		returnRepo := repository.NewFactoryReturnRepository(tx)
		wRepo := repository.NewWarehouseRepository(tx)
		pRepo := repository.NewFactoryRepository(tx)

		shippedDate := time.Now()
		if req.ReturnDate != nil && !req.ReturnDate.IsZero() {
			shippedDate = *req.ReturnDate
		}

		// Generate Return No
		countStr := fmt.Sprintf("%04d", time.Now().Unix()%10000)
		returnNo := fmt.Sprintf("RTF-%s-%s", shippedDate.Format("20060102"), countStr)

		factoryReturn := &domain.FactoryReturn{
			ReturnNo:     returnNo,
			FromBranchID: req.FromBranchID,
			ToFactoryID:  req.ToFactoryID,
			Notes:        req.Notes,
			VehicleID:    req.VehicleID,
			DriverID:     req.DriverID,
			Status:       "SHIPPED",
			ShippedAt:    &shippedDate,
		}

		for _, item := range req.Items {
			if item.ReturnType != "GOOD" && strings.TrimSpace(item.DamageReason) == "" {
				return fmt.Errorf("keterangan kerusakan wajib diisi untuk produk yang rusak")
			}
			batch := item.BatchNo
			if batch == "" {
				batch = "DEFAULT"
			}

			returnType := item.ReturnType
			if returnType != "GOOD" {
				returnType = "DAMAGED"
			}

			damageType := item.DamageType
			if returnType == "GOOD" {
				damageType = "NONE"
			} else if damageType == "" {
				damageType = "OTHER"
			}

			product, err := pRepo.GetProductByID(item.ProductID)
			if err != nil {
				return err
			}

			pcsPerUnit := product.PcsPerUnit
			if pcsPerUnit <= 0 {
				pcsPerUnit = 1
			}

			totalPcs := item.Quantity * pcsPerUnit

			// Validasi dan potong stok produk di gudang cabang
			if returnType == "GOOD" {
				// 1. Coba ambil stok GOOD pada batch spesifik yang diminta
				stock, _ := wRepo.GetStock(req.FromBranchID, item.ProductID, "GOOD", batch)
				availInRequestedBatch := 0
				if stock != nil && stock.Quantity > 0 {
					availInRequestedBatch = stock.Quantity
				}

				// Jika batch spesifik mencukupi
				if availInRequestedBatch >= totalPcs {
					stock.Quantity -= totalPcs
					if err := wRepo.UpdateStock(stock); err != nil {
						return err
					}
					wRepo.CreateLog(&domain.WarehouseLog{
						BranchID:  req.FromBranchID,
						ProductID: item.ProductID,
						Type:      "OUT",
						Source:    fmt.Sprintf("FACTORY_RETURN (GOOD): %s", returnNo),
						Quantity:  totalPcs,
					})
				} else {
					// Jika batch spesifik tidak mencukupi, kumpulkan semua batch GOOD yang tersedia di cabang ini
					availBatches, errBatches := wRepo.GetAvailableBatches(req.FromBranchID, item.ProductID, "GOOD")
					if errBatches != nil {
						return errBatches
					}

					totalAvailable := 0
					for _, b := range availBatches {
						totalAvailable += b.Quantity
					}

					if totalAvailable < totalPcs {
						return fmt.Errorf("stok barang (GOOD) tidak mencukupi untuk %s (tersedia: %d pcs, diminta: %d pcs)", product.Name, totalAvailable, totalPcs)
					}

					remaining := totalPcs
					// Prioritaskan potong dari batch yang dipilih dulu jika ada sisanya
					if availInRequestedBatch > 0 && stock != nil {
						deduct := availInRequestedBatch
						if deduct > remaining {
							deduct = remaining
						}
						stock.Quantity -= deduct
						remaining -= deduct
						if err := wRepo.UpdateStock(stock); err != nil {
							return err
						}
						wRepo.CreateLog(&domain.WarehouseLog{
							BranchID:  req.FromBranchID,
							ProductID: item.ProductID,
							Type:      "OUT",
							Source:    fmt.Sprintf("FACTORY_RETURN (GOOD Batch %s): %s", stock.BatchNo, returnNo),
							Quantity:  deduct,
						})
					}

					// Potong sisa kebutuhan dari batch-batch lain yang tersedia
					for i := range availBatches {
						if remaining <= 0 {
							break
						}
						b := &availBatches[i]
						if stock != nil && b.ID == stock.ID {
							continue
						}
						if b.Quantity <= 0 {
							continue
						}
						deduct := b.Quantity
						if deduct > remaining {
							deduct = remaining
						}
						b.Quantity -= deduct
						remaining -= deduct
						if err := wRepo.UpdateStock(b); err != nil {
							return err
						}
						wRepo.CreateLog(&domain.WarehouseLog{
							BranchID:  req.FromBranchID,
							ProductID: item.ProductID,
							Type:      "OUT",
							Source:    fmt.Sprintf("FACTORY_RETURN (GOOD Batch %s): %s", b.BatchNo, returnNo),
							Quantity:  deduct,
						})
					}
				}
			} else {
				// returnType == "DAMAGED"
				// Cek stok DAMAGED pada batch yang dipilih & batch-batch lain jika kurang
				damagedStock, _ := wRepo.GetStock(req.FromBranchID, item.ProductID, "DAMAGED", batch)
				availDamagedInRequested := 0
				if damagedStock != nil && damagedStock.Quantity > 0 {
					availDamagedInRequested = damagedStock.Quantity
				}

				allDamagedBatches, _ := wRepo.GetAvailableBatches(req.FromBranchID, item.ProductID, "DAMAGED")
				totalDamagedAvailable := 0
				for _, b := range allDamagedBatches {
					totalDamagedAvailable += b.Quantity
				}
				// Pastikan jika ada damagedStock yang belum tercakup di allDamagedBatches
				if availDamagedInRequested > 0 {
					found := false
					for _, b := range allDamagedBatches {
						if b.ID == damagedStock.ID {
							found = true
							break
						}
					}
					if !found {
						totalDamagedAvailable += availDamagedInRequested
					}
				}

				// Cek juga ketersediaan stok GOOD sebagai fallback retur jika rusak
				goodStock, _ := wRepo.GetStock(req.FromBranchID, item.ProductID, "GOOD", batch)
				availGoodInRequested := 0
				if goodStock != nil && goodStock.Quantity > 0 {
					availGoodInRequested = goodStock.Quantity
				}

				allGoodBatches, _ := wRepo.GetAvailableBatches(req.FromBranchID, item.ProductID, "GOOD")
				totalGoodAvailable := 0
				for _, b := range allGoodBatches {
					totalGoodAvailable += b.Quantity
				}
				if availGoodInRequested > 0 {
					found := false
					for _, b := range allGoodBatches {
						if b.ID == goodStock.ID {
							found = true
							break
						}
					}
					if !found {
						totalGoodAvailable += availGoodInRequested
					}
				}

				totalAvailable := totalDamagedAvailable + totalGoodAvailable
				if totalAvailable < totalPcs {
					return fmt.Errorf("stok barang tidak mencukupi untuk %s (tersedia: %d pcs, diminta: %d pcs)", product.Name, totalAvailable, totalPcs)
				}

				remainingToDeduct := totalPcs

				// 1. Potong dari DAMAGED stock pada batch yang dipilih
				if availDamagedInRequested > 0 && damagedStock != nil {
					deductDamaged := availDamagedInRequested
					if deductDamaged > remainingToDeduct {
						deductDamaged = remainingToDeduct
					}
					damagedStock.Quantity -= deductDamaged
					remainingToDeduct -= deductDamaged
					if err := wRepo.UpdateStock(damagedStock); err != nil {
						return err
					}
					wRepo.CreateLog(&domain.WarehouseLog{
						BranchID:  req.FromBranchID,
						ProductID: item.ProductID,
						Type:      "OUT",
						Source:    fmt.Sprintf("FACTORY_RETURN (DAMAGED Batch %s): %s", damagedStock.BatchNo, returnNo),
						Quantity:  deductDamaged,
					})
				}

				// 2. Jika masih sisa, potong dari DAMAGED stock batch lain
				if remainingToDeduct > 0 {
					for i := range allDamagedBatches {
						if remainingToDeduct <= 0 {
							break
						}
						b := &allDamagedBatches[i]
						if damagedStock != nil && b.ID == damagedStock.ID {
							continue
						}
						if b.Quantity <= 0 {
							continue
						}
						deduct := b.Quantity
						if deduct > remainingToDeduct {
							deduct = remainingToDeduct
						}
						b.Quantity -= deduct
						remainingToDeduct -= deduct
						if err := wRepo.UpdateStock(b); err != nil {
							return err
						}
						wRepo.CreateLog(&domain.WarehouseLog{
							BranchID:  req.FromBranchID,
							ProductID: item.ProductID,
							Type:      "OUT",
							Source:    fmt.Sprintf("FACTORY_RETURN (DAMAGED Batch %s): %s", b.BatchNo, returnNo),
							Quantity:  deduct,
						})
					}
				}

				// 3. Jika masih sisa kuantitas, potong dari GOOD stock batch terpilih
				if remainingToDeduct > 0 && availGoodInRequested > 0 && goodStock != nil {
					deductGood := availGoodInRequested
					if deductGood > remainingToDeduct {
						deductGood = remainingToDeduct
					}
					goodStock.Quantity -= deductGood
					remainingToDeduct -= deductGood
					if err := wRepo.UpdateStock(goodStock); err != nil {
						return err
					}
					wRepo.CreateLog(&domain.WarehouseLog{
						BranchID:  req.FromBranchID,
						ProductID: item.ProductID,
						Type:      "OUT",
						Source:    fmt.Sprintf("FACTORY_RETURN (DAMAGED from GOOD Batch %s): %s", goodStock.BatchNo, returnNo),
						Quantity:  deductGood,
					})
				}

				// 4. Jika masih sisa kuantitas, potong dari GOOD stock batch-batch lainnya
				if remainingToDeduct > 0 {
					for i := range allGoodBatches {
						if remainingToDeduct <= 0 {
							break
						}
						b := &allGoodBatches[i]
						if goodStock != nil && b.ID == goodStock.ID {
							continue
						}
						if b.Quantity <= 0 {
							continue
						}
						deduct := b.Quantity
						if deduct > remainingToDeduct {
							deduct = remainingToDeduct
						}
						b.Quantity -= deduct
						remainingToDeduct -= deduct
						if err := wRepo.UpdateStock(b); err != nil {
							return err
						}
						wRepo.CreateLog(&domain.WarehouseLog{
							BranchID:  req.FromBranchID,
							ProductID: item.ProductID,
							Type:      "OUT",
							Source:    fmt.Sprintf("FACTORY_RETURN (DAMAGED from GOOD Batch %s): %s", b.BatchNo, returnNo),
							Quantity:  deduct,
						})
					}
				}
			}

			factoryReturn.Items = append(factoryReturn.Items, domain.FactoryReturnItem{
				ProductID:    item.ProductID,
				BatchNo:      batch,
				Quantity:     item.Quantity,
				Unit:         product.Unit,
				PcsPerUnit:   pcsPerUnit,
				ReturnType:   returnType,
				DamageType:   damageType,
				DamageReason: item.DamageReason,
				ExpiryDate:   item.ExpiryDate,
			})
		}

		if err := returnRepo.Create(factoryReturn); err != nil {
			return err
		}

		createdReturn = factoryReturn
		return nil
	})

	if err != nil {
		return nil, err
	}

	fullReturn, err := u.returnRepo.GetByID(createdReturn.ID)
	if err == nil && u.sseHub != nil && fullReturn != nil {
		go func() {
			branchName := "Cabang"
			if fullReturn.FromBranch != nil {
				branchName = fullReturn.FromBranch.Name
			}
			factoryName := "Pabrik"
			if fullReturn.ToFactory != nil {
				factoryName = fullReturn.ToFactory.Name
			}
			companyID := ""
			if fullReturn.FromBranch != nil && fullReturn.FromBranch.CompanyID != nil {
				companyID = fullReturn.FromBranch.CompanyID.String()
			} else if fullReturn.ToFactory != nil {
				companyID = fullReturn.ToFactory.CompanyID.String()
			}

			event := sse.NotificationEvent{
				ID:        fullReturn.ID.String(),
				Type:      "factory_return",
				Title:     "Retur ke Pabrik Baru",
				Message:   fmt.Sprintf("Cabang %s mengajukan retur barang %s ke pabrik %s", branchName, fullReturn.ReturnNo, factoryName),
				CompanyID: companyID,
				CreatedAt: time.Now().Format("2006-01-02 15:04:05"),
				Data: map[string]interface{}{
					"return_id":    fullReturn.ID.String(),
					"return_no":    fullReturn.ReturnNo,
					"branch_name":  branchName,
					"factory_name": factoryName,
				},
			}
			_ = u.sseHub.PublishNotification(context.Background(), event)
		}()
	}
	return fullReturn, err
}

func (u *factoryReturnUsecase) ReceiveReturn(id uuid.UUID, req *ReceiveFactoryReturnRequest) error {
	return u.db.Transaction(func(tx *gorm.DB) error {
		returnRepo := repository.NewFactoryReturnRepository(tx)
		fRepo := repository.NewFactoryRepository(tx)

		ret, err := returnRepo.GetByID(id)
		if err != nil {
			return err
		}

		if ret.Status == "RECEIVED" {
			return errors.New("retur ini sudah diterima")
		}

		// Create item lookup if req has items
		reqItemMap := make(map[uuid.UUID]ReceiveFactoryReturnItemReq)
		if req != nil && len(req.Items) > 0 {
			for _, itemReq := range req.Items {
				reqItemMap[itemReq.ItemID] = itemReq
			}
		}

		totalReceivedQty := 0
		totalGoodQty := 0
		totalBottleDamagedQty := 0
		totalContentDamagedQty := 0

		for i := range ret.Items {
			item := &ret.Items[i]
			pcsPerUnit := item.PcsPerUnit
			if pcsPerUnit <= 0 {
				pcsPerUnit = 1
			}

			itemReq, hasCustomQC := reqItemMap[item.ID]

			var goodQty, bottleDamagedQty, contentDamagedQty, totalRecQty int
			var disposition, inspectionNotes string
			var targetWarehouseID *uuid.UUID

			if hasCustomQC {
				goodQty = itemReq.ReceivedGoodQty
				bottleDamagedQty = itemReq.ReceivedBottleDamagedQty
				contentDamagedQty = itemReq.ReceivedContentDamagedQty
				totalRecQty = goodQty + bottleDamagedQty + contentDamagedQty
				disposition = itemReq.DispositionAction
				inspectionNotes = itemReq.InspectionNotes
				targetWarehouseID = itemReq.TargetFactoryWarehouseID

				if disposition == "" {
					if goodQty > 0 && bottleDamagedQty == 0 && contentDamagedQty == 0 {
						disposition = "RESTOCK_GOOD"
					} else if bottleDamagedQty > 0 && goodQty == 0 && contentDamagedQty == 0 {
						disposition = "REPACK_BOTTLE"
					} else if contentDamagedQty > 0 && goodQty == 0 && bottleDamagedQty == 0 {
						disposition = "SCRAP_DISPOSAL"
					} else if totalRecQty == 0 {
						disposition = "MISSING_LOST"
					} else {
						disposition = "PARTIAL_MIXED"
					}
				}
				if totalRecQty < 0 {
					return errors.New("jumlah hasil inspeksi retur tidak boleh negatif")
				}
				if totalRecQty > item.Quantity {
					return fmt.Errorf("jumlah hasil inspeksi untuk %s melebihi jumlah retur", item.Product.Name)
				}
				if (bottleDamagedQty > 0 || contentDamagedQty > 0 || totalRecQty < item.Quantity) && strings.TrimSpace(inspectionNotes) == "" && strings.TrimSpace(item.DamageReason) == "" {
					// Beri catatan default jika tidak diisi
					if inspectionNotes == "" {
						inspectionNotes = "Penyesuaian penerimaan fisik retur oleh gudang / QC"
					}
				}
			} else {
				// Default fallback according to original item declaration
				if item.ReturnType == "GOOD" {
					goodQty = item.Quantity
					disposition = "RESTOCK_GOOD"
				} else if item.DamageType == "BOTTLE_PACKAGING" {
					bottleDamagedQty = item.Quantity
					disposition = "REPACK_BOTTLE"
				} else if item.DamageType == "PRODUCT_CONTENT" || item.DamageType == "EXPIRED" || item.DamageType == "TOTAL_DAMAGE" {
					contentDamagedQty = item.Quantity
					disposition = "SCRAP_DISPOSAL"
				} else {
					bottleDamagedQty = item.Quantity
					disposition = "REPACK_BOTTLE"
				}
				totalRecQty = item.Quantity
			}

			item.ReceivedQuantity = totalRecQty
			item.ReceivedGoodQty = goodQty
			item.ReceivedBottleDamagedQty = bottleDamagedQty
			item.ReceivedContentDamagedQty = contentDamagedQty
			item.DispositionAction = disposition
			item.InspectionNotes = inspectionNotes
			item.TargetFactoryWarehouseID = targetWarehouseID

			totalReceivedQty += totalRecQty
			totalGoodQty += goodQty
			totalBottleDamagedQty += bottleDamagedQty
			totalContentDamagedQty += contentDamagedQty

			// Update database for item
			if err := tx.Save(item).Error; err != nil {
				return err
			}

			// 1. Process Good Qty -> Add to finished goods stock
			if goodQty > 0 {
				goodPcs := goodQty * pcsPerUnit
				stock, err := fRepo.GetStock(ret.ToFactoryID, item.ProductID, item.BatchNo)
				if err != nil {
					return err
				}
				stock.Quantity += goodPcs
				if targetWarehouseID != nil {
					stock.FactoryWarehouseID = targetWarehouseID
				}
				if err := fRepo.UpdateStock(stock); err != nil {
					return err
				}

				// Sync to Linked Warehouses
				if err := syncFactoryStockDeltaToLinkedWarehouses(tx, ret.ToFactoryID, item.ProductID, item.BatchNo, item.ExpiryDate, goodPcs, "RETURN_GOOD: "+ret.ReturnNo, "SYSTEM"); err != nil {
					return err
				}

				// Log Factory Inventory
				fRepo.CreateInventoryLog(&domain.FactoryInventoryLog{
					FactoryID: ret.ToFactoryID,
					ProductID: item.ProductID,
					Type:      "IN",
					Source:    "RETURN_GOOD: " + ret.ReturnNo,
					BatchNo:   item.BatchNo,
					Quantity:  goodPcs,
				})
			}

			// 2. Process Bottle Damaged Qty -> Add to factory stock for repack / rework
			if bottleDamagedQty > 0 {
				repackPcs := bottleDamagedQty * pcsPerUnit
				stock, err := fRepo.GetStock(ret.ToFactoryID, item.ProductID, item.BatchNo)
				if err != nil {
					return err
				}
				stock.Quantity += repackPcs
				if targetWarehouseID != nil {
					stock.FactoryWarehouseID = targetWarehouseID
				}
				if err := fRepo.UpdateStock(stock); err != nil {
					return err
				}

				if err := syncFactoryStockDeltaToLinkedWarehouses(tx, ret.ToFactoryID, item.ProductID, item.BatchNo, item.ExpiryDate, repackPcs, "RETURN_REPACK: "+ret.ReturnNo, "SYSTEM"); err != nil {
					return err
				}

				fRepo.CreateInventoryLog(&domain.FactoryInventoryLog{
					FactoryID: ret.ToFactoryID,
					ProductID: item.ProductID,
					Type:      "IN",
					Source:    "RETURN_REPACK: " + ret.ReturnNo,
					BatchNo:   item.BatchNo,
					Quantity:  repackPcs,
				})
			}

			// 3. Process Content Damaged / Expired / Scrap Qty -> Log only for destruction audit
			if contentDamagedQty > 0 {
				scrapPcs := contentDamagedQty * pcsPerUnit
				fRepo.CreateInventoryLog(&domain.FactoryInventoryLog{
					FactoryID: ret.ToFactoryID,
					ProductID: item.ProductID,
					Type:      "OUT",
					Source:    "RETURN_SCRAP_WASTE: " + ret.ReturnNo,
					BatchNo:   item.BatchNo,
					Quantity:  scrapPcs,
				})
			}
		}

		now := time.Now()
		ret.Status = "RECEIVED"
		ret.IsReceived = true
		ret.ReceivedAt = &now
		ret.TotalReceivedQty = totalReceivedQty
		ret.TotalGoodQty = totalGoodQty
		ret.TotalBottleDamagedQty = totalBottleDamagedQty
		ret.TotalContentDamagedQty = totalContentDamagedQty

		if req != nil {
			if req.ReceiverName != "" {
				ret.ReceiverName = req.ReceiverName
				ret.ReceivedBy = &req.ReceiverName
			}
			if req.InspectedByName != "" {
				ret.InspectedByName = req.InspectedByName
			}
			if req.InspectionNotes != "" {
				ret.InspectionNotes = req.InspectionNotes
			}
		}

		return returnRepo.Update(ret)
	})
}

func (u *factoryReturnUsecase) RejectReturn(id uuid.UUID, rejectorID *uuid.UUID, rejectorName string, reason string) error {
	return u.db.Transaction(func(tx *gorm.DB) error {
		returnRepo := repository.NewFactoryReturnRepository(tx)
		wRepo := repository.NewWarehouseRepository(tx)

		ret, err := returnRepo.GetByID(id)
		if err != nil {
			return err
		}

		if ret.Status == "RECEIVED" {
			return errors.New("retur yang sudah diterima tidak dapat ditolak")
		}
		if ret.Status == "REJECTED" {
			return errors.New("retur ini sudah dalam status ditolak")
		}

		// Restore Branch Stock that was deducted upon return creation
		for _, item := range ret.Items {
			pcsPerUnit := item.PcsPerUnit
			if pcsPerUnit <= 0 {
				pcsPerUnit = 1
			}
			totalPcs := item.Quantity * pcsPerUnit
			stock, errStock := wRepo.GetStock(ret.FromBranchID, item.ProductID, item.ReturnType, item.BatchNo)
			if errStock == nil && stock != nil {
				stock.Quantity += totalPcs
				_ = wRepo.UpdateStock(stock)

				wRepo.CreateLog(&domain.WarehouseLog{
					BranchID:  ret.FromBranchID,
					ProductID: item.ProductID,
					Type:      "IN",
					Source:    fmt.Sprintf("FACTORY_RETURN_REJECTED (%s): %s", item.ReturnType, ret.ReturnNo),
					Quantity:  totalPcs,
				})
			}
		}

		now := time.Now()
		ret.Status = "REJECTED"
		ret.RejectionReason = reason
		ret.ApprovedByID = rejectorID
		ret.ApprovedByName = rejectorName
		ret.ApprovedAt = &now

		return returnRepo.Update(ret)
	})
}

func (u *factoryReturnUsecase) GetByBranch(branchID uuid.UUID) ([]domain.FactoryReturn, error) {
	return u.returnRepo.GetByBranch(branchID)
}

func (u *factoryReturnUsecase) GetByFactory(factoryID uuid.UUID) ([]domain.FactoryReturn, error) {
	return u.returnRepo.GetByFactory(factoryID)
}

func (u *factoryReturnUsecase) GetByCompany(companyID uuid.UUID) ([]domain.FactoryReturn, error) {
	return u.returnRepo.GetByCompany(companyID)
}
