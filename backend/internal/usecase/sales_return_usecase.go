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
	"gorm.io/gorm/clause"
)

type SalesReturnUsecase interface {
	CreateReturn(req CreateReturnRequest) (*domain.SalesReturn, error)
	UpdateReturn(id uuid.UUID, req UpdateReturnRequest) (*domain.SalesReturn, error)
	ApproveReturnByNota(id uuid.UUID, approvedByID uuid.UUID) error
	ApproveReturnByWarehouse(id uuid.UUID, approvedByID uuid.UUID) error
	GetByBranch(branchID uuid.UUID, status string) ([]domain.SalesReturn, error)
	GetByID(id uuid.UUID) (*domain.SalesReturn, error)
	DeleteReturn(id uuid.UUID) error
}

type salesReturnUsecase struct {
	returnRepo    repository.SalesReturnRepository
	soRepo        repository.SalesOrderRepository
	salesRepo     repository.SalesTransactionRepository
	warehouseRepo repository.WarehouseRepository
	db            *gorm.DB
	sseHub        *sse.Hub
}

func NewSalesReturnUsecase(
	returnRepo repository.SalesReturnRepository,
	soRepo repository.SalesOrderRepository,
	salesRepo repository.SalesTransactionRepository,
	warehouseRepo repository.WarehouseRepository,
	db *gorm.DB,
	sseHub *sse.Hub,
) SalesReturnUsecase {
	return &salesReturnUsecase{returnRepo, soRepo, salesRepo, warehouseRepo, db, sseHub}
}

type CreateReturnRequest struct {
	SOID                 *uuid.UUID          `json:"so_id"`
	TransactionID        *uuid.UUID          `json:"transaction_id"`
	BranchID             *uuid.UUID          `json:"branch_id"`
	EmployeeID           uuid.UUID           `json:"employee_id"`
	ReturnType           string              `json:"return_type"` // POTONG_NOTA, TUKAR_LANGSUNG, TUKAR_SUSULAN, TUKAR_PRODUK, GARANSI
	Notes                string              `json:"notes"`
	DeliveryBatchID      *uuid.UUID          `json:"delivery_batch_id"`
	AutoApproveWarehouse bool                `json:"auto_approve_warehouse"`
	Items                []ReturnItemRequest `json:"items" binding:"required,min=1"`
}

type ReturnItemRequest struct {
	ProductID            uuid.UUID  `json:"product_id" binding:"required"`
	SalesItemID          *uuid.UUID `json:"sales_item_id"`
	Quantity             int        `json:"quantity" binding:"required,min=1"`
	Price                float64    `json:"price"`
	Reason               string     `json:"reason"`
	Condition            string     `json:"condition"` // 'GOOD', 'DAMAGED'
	ReplacementProductID *uuid.UUID `json:"replacement_product_id"`
	ReplacementQuantity  int        `json:"replacement_quantity"`
	ReplacementPrice     float64    `json:"replacement_price"`
}

type UpdateReturnRequest struct {
	Notes string              `json:"notes"`
	Items []ReturnItemRequest `json:"items" binding:"required"`
}

func (u *salesReturnUsecase) CreateReturn(req CreateReturnRequest) (*domain.SalesReturn, error) {
	var branchID uuid.UUID
	var soID *uuid.UUID
	var trxID *uuid.UUID
	var initialStatus domain.SalesReturnStatus = domain.ReturnStatusPending
	var storeName string = "Pelanggan"
	var defaultNotes = req.Notes
	var deliveryBatchID *uuid.UUID = req.DeliveryBatchID
	var companyID uuid.UUID

	if req.SOID != nil && *req.SOID != uuid.Nil {
		so, err := u.soRepo.FindByID(*req.SOID)
		if err != nil {
			return nil, err
		}
		if so == nil {
			return nil, errors.New("sales order tidak ditemukan")
		}
		branchID = so.BranchID
		soID = &so.ID
		trxID = so.InvoiceID
		companyID = so.CompanyID
		if so.Store != nil {
			storeName = so.Store.Name
		}
		if deliveryBatchID == nil && so.DeliveryBatchID != nil {
			deliveryBatchID = so.DeliveryBatchID
		}
		// Jika employee_id belum ditentukan (atau default user login admin), ambil driver pengiriman jika ada
		if deliveryBatchID != nil {
			var batch domain.DeliveryBatch
			if errBatch := u.db.Select("id, driver_id, daily_driver_id, delivery_order_no").First(&batch, "id = ?", *deliveryBatchID).Error; errBatch == nil {
				if req.EmployeeID == uuid.Nil && batch.DriverID != nil {
					req.EmployeeID = *batch.DriverID
				}
				if batch.DeliveryOrderNo != nil && *batch.DeliveryOrderNo != "" && !strings.Contains(defaultNotes, *batch.DeliveryOrderNo) {
					if defaultNotes != "" {
						defaultNotes = fmt.Sprintf("[DELIVERY_RETURN: %s] %s", *batch.DeliveryOrderNo, defaultNotes)
					} else {
						defaultNotes = fmt.Sprintf("[DELIVERY_RETURN: %s]", *batch.DeliveryOrderNo)
					}
				}
			}
		}
		if !strings.Contains(defaultNotes, "[DELIVERY_RETURN]") {
			if defaultNotes != "" {
				defaultNotes = "[DELIVERY_RETURN] " + defaultNotes
			} else {
				defaultNotes = "[DELIVERY_RETURN]"
			}
		}
		// Input retur pengiriman langsung siap diperiksa/diterima fisik gudang
		initialStatus = domain.ReturnStatusPendingWarehouse
	} else if req.TransactionID != nil && *req.TransactionID != uuid.Nil {
		trx, err := u.salesRepo.FindByID(*req.TransactionID)
		if err != nil {
			return nil, err
		}
		if trx == nil {
			return nil, errors.New("faktur penjualan tidak ditemukan")
		}

		// Prioritas resolusi Branch ID:
		if req.BranchID != nil && *req.BranchID != uuid.Nil {
			branchID = *req.BranchID
		} else if trx.BranchID != nil && *trx.BranchID != uuid.Nil {
			branchID = *trx.BranchID
		} else if trx.Employee != nil && trx.Employee.BranchID != nil && *trx.Employee.BranchID != uuid.Nil {
			branchID = *trx.Employee.BranchID
		} else if trx.CashierEmployee != nil && trx.CashierEmployee.BranchID != nil && *trx.CashierEmployee.BranchID != uuid.Nil {
			branchID = *trx.CashierEmployee.BranchID
		} else if trx.BuyerEmployee != nil && trx.BuyerEmployee.BranchID != nil && *trx.BuyerEmployee.BranchID != uuid.Nil {
			branchID = *trx.BuyerEmployee.BranchID
		} else if req.EmployeeID != uuid.Nil {
			var emp domain.Employee
			if err := u.db.Select("branch_id").First(&emp, "id = ?", req.EmployeeID).Error; err == nil && emp.BranchID != nil && *emp.BranchID != uuid.Nil {
				branchID = *emp.BranchID
			}
		}

		// Fallback ke cabang aktif milik perusahaan faktur atau cabang aktif pertama
		if branchID == uuid.Nil {
			var fallbackBranch domain.Branch
			q := u.db.Model(&domain.Branch{}).Where("is_active = ?", true)
			if trx.CompanyID != uuid.Nil {
				q = q.Where("company_id = ?", trx.CompanyID)
			}
			if err := q.First(&fallbackBranch).Error; err == nil && fallbackBranch.ID != uuid.Nil {
				branchID = fallbackBranch.ID
			} else {
				var anyBranch domain.Branch
				if err := u.db.Model(&domain.Branch{}).First(&anyBranch).Error; err == nil && anyBranch.ID != uuid.Nil {
					branchID = anyBranch.ID
				}
			}
		}

		trxID = &trx.ID
		companyID = trx.CompanyID
		if trx.Store != nil {
			storeName = trx.Store.Name
		} else if trx.BuyerName != "" {
			storeName = trx.BuyerName
		}

		// Cari SalesOrder terkait jika faktur berasal dari pesanan SO
		var linkedSO domain.SalesOrder
		if err := u.db.Where("invoice_id = ?", trx.ID).First(&linkedSO).Error; err == nil && linkedSO.ID != uuid.Nil {
			soID = &linkedSO.ID
			if branchID == uuid.Nil && linkedSO.BranchID != uuid.Nil {
				branchID = linkedSO.BranchID
			}
		}

		// Retur penjualan yang diinput oleh Admin Nota langsung berstatus PENDING_WAREHOUSE
		initialStatus = domain.ReturnStatusPendingWarehouse
		if soID == nil && !strings.Contains(defaultNotes, "[DIRECT_SALE_RETURN]") {
			if defaultNotes != "" {
				defaultNotes = "[DIRECT_SALE_RETURN] " + defaultNotes
			} else {
				defaultNotes = "[DIRECT_SALE_RETURN]"
			}
		}
	} else if req.BranchID != nil && *req.BranchID != uuid.Nil {
		branchID = *req.BranchID
	} else {
		return nil, errors.New("sales_order_id atau transaction_id wajib disertakan")
	}

	if branchID == uuid.Nil {
		return nil, errors.New("cabang tujuan retur (branch_id) tidak valid atau tidak ditemukan")
	}

	now := time.Now()
	returnNo, err := u.returnRepo.GenerateNextReturnNumber(now)
	if err != nil {
		return nil, fmt.Errorf("gagal generate nomor retur: %w", err)
	}

	rType := req.ReturnType
	if rType == "" {
		rType = "POTONG_NOTA"
	}
	var swapInvoice *domain.SalesTransaction
	if rType == "TUKAR_PRODUK" && req.TransactionID != nil {
		swapInvoice, err = u.salesRepo.FindByID(*req.TransactionID)
		if err != nil || swapInvoice == nil {
			return nil, errors.New("faktur untuk tukar produk tidak ditemukan")
		}
		if swapInvoice.BranchID != nil && *swapInvoice.BranchID != branchID {
			return nil, errors.New("cabang retur tukar produk harus sama dengan cabang faktur")
		}
	}

	ret := &domain.SalesReturn{
		ReturnNo:      returnNo,
		SalesOrderID:  soID,
		TransactionID: trxID,
		BranchID:      branchID,
		EmployeeID:    req.EmployeeID,
		ReturnType:    rType,
		Status:        initialStatus,
		Notes:         defaultNotes,
	}

	var totalAmount float64
	returnedQuantity := make(map[uuid.UUID]int)
	for _, item := range req.Items {
		var price float64 = item.Price
		if swapInvoice != nil {
			if item.Quantity <= 0 {
				return nil, errors.New("jumlah barang retur harus lebih dari nol")
			}
			var original *domain.SalesItem
			for i := range swapInvoice.Items {
				if item.SalesItemID != nil && swapInvoice.Items[i].ID == *item.SalesItemID && swapInvoice.Items[i].ProductID == item.ProductID {
					original = &swapInvoice.Items[i]
					break
				}
				if item.SalesItemID == nil && swapInvoice.Items[i].ProductID == item.ProductID {
					if original != nil {
						return nil, errors.New("produk muncul lebih dari sekali pada faktur; pilih baris faktur yang tepat")
					}
					original = &swapInvoice.Items[i]
				}
			}
			if original == nil {
				return nil, errors.New("barang retur harus ada pada faktur")
			}
			returnedQuantity[original.ID] += item.Quantity
			if returnedQuantity[original.ID] > original.Quantity {
				return nil, errors.New("jumlah retur tidak boleh melebihi jumlah barang pada faktur")
			}
			if item.ReplacementProductID == nil || *item.ReplacementProductID == uuid.Nil || *item.ReplacementProductID == item.ProductID || item.ReplacementQuantity <= 0 {
				return nil, errors.New("pilih barang pengganti yang berbeda dan jumlah pengganti yang valid")
			}
			var replacement domain.Product
			if err := u.db.First(&replacement, "id = ?", *item.ReplacementProductID).Error; err != nil || replacement.SellingPrice <= 0 {
				return nil, errors.New("barang pengganti tidak ditemukan atau belum memiliki harga jual")
			}
			price = original.PriceAtTransaction * (1 - original.DiscountPercent/100)
			item.SalesItemID = &original.ID
			item.ReplacementPrice = replacement.SellingPrice
			ret.PriceDifference += float64(item.ReplacementQuantity)*item.ReplacementPrice - float64(item.Quantity)*price
		}
		if price <= 0 {
			if soID != nil {
				so, _ := u.soRepo.FindByID(*soID)
				if so != nil {
					for _, soItem := range so.Items {
						if soItem.ProductID == item.ProductID {
							price = soItem.Price
							break
						}
					}
				}
			} else if trxID != nil {
				trx, _ := u.salesRepo.FindByID(*trxID)
				if trx != nil {
					for _, trxItem := range trx.Items {
						if trxItem.ProductID == item.ProductID {
							price = trxItem.PriceAtTransaction
							break
						}
					}
				}
			}
		}

		retItem := domain.SalesReturnItem{
			ProductID:            item.ProductID,
			SalesItemID:          item.SalesItemID,
			Quantity:             item.Quantity,
			Price:                price,
			Reason:               item.Reason,
			Condition:            item.Condition,
			ReplacementProductID: item.ReplacementProductID,
			ReplacementQuantity:  item.ReplacementQuantity,
			ReplacementPrice:     item.ReplacementPrice,
		}
		ret.Items = append(ret.Items, retItem)
		totalAmount += float64(item.Quantity) * price
	}
	ret.TotalAmount = totalAmount
	if swapInvoice != nil && swapInvoice.TotalAmount+ret.PriceDifference < 0 {
		return nil, errors.New("nilai barang yang diretur melebihi nilai faktur")
	}

	var createErr error
	for attempt := 0; attempt < 3; attempt++ {
		if swapInvoice != nil {
			createErr = u.db.Transaction(func(tx *gorm.DB) error {
				if err := repository.NewSalesReturnRepository(tx).Create(ret); err != nil {
					return err
				}
				if err := applyProductExchangeToInvoice(tx, ret); err != nil {
					return err
				}
				return tx.Model(&domain.SalesReturn{}).Where("id = ?", ret.ID).
					Update("invoice_adjusted_at", ret.InvoiceAdjustedAt).Error
			})
		} else {
			createErr = u.returnRepo.Create(ret)
		}
		if createErr == nil {
			break
		}
		if strings.Contains(createErr.Error(), "duplicate key") || strings.Contains(createErr.Error(), "unique constraint") {
			ret.ID = uuid.Nil
			for i := range ret.Items {
				ret.Items[i].ID = uuid.Nil
				ret.Items[i].SalesReturnID = uuid.Nil
			}
			newNo, genErr := u.returnRepo.GenerateNextReturnNumber(time.Now())
			if genErr == nil {
				ret.ReturnNo = newNo
				continue
			}
		}
		break
	}
	if createErr != nil {
		return nil, createErr
	}
	if swapInvoice != nil {
		invalidateReportsCache(context.Background())
	}

	// Jika ini retur terkait TransactionID (Faktur Penjualan)
	if req.TransactionID != nil {
		inv, err := u.salesRepo.FindByID(*req.TransactionID)
		if err == nil && inv != nil {
			newNotes := ""
			if inv.Notes != nil {
				newNotes = *inv.Notes + "\n"
			}

			if rType == "POTONG_NOTA" {
				// Potong total tagihan dan sesuaikan paid_amount
				inv.TotalAmount -= ret.TotalAmount
				if inv.TotalAmount < 0 {
					inv.TotalAmount = 0
				}
				if inv.PaidAmount > inv.TotalAmount {
					inv.PaidAmount = inv.TotalAmount
				}

				// Kurangi kuantitas item barang pada faktur
				for _, retItem := range req.Items {
					for j := range inv.Items {
						if inv.Items[j].ProductID == retItem.ProductID {
							inv.Items[j].Quantity -= retItem.Quantity
							if inv.Items[j].Quantity < 0 {
								inv.Items[j].Quantity = 0
							}
							disc := inv.Items[j].DiscountPercent / 100
							inv.Items[j].Subtotal = float64(inv.Items[j].Quantity) * inv.Items[j].PriceAtTransaction * (1 - disc)
							_ = u.db.Save(&inv.Items[j])
						}
					}
				}

				// Sinkronisasi SalesOrder terkait jika ada
				if soID != nil {
					u.db.Model(&domain.SalesOrder{}).Where("id = ?", *soID).Updates(map[string]interface{}{
						"total_amount": inv.TotalAmount,
					})
				}

				// Sinkronisasi antrean Persetujuan Kasir (SalesPayment PENDING)
				var pendingPayments []domain.SalesPayment
				pQuery := u.db.Where("payment_status = ?", "PENDING")
				if soID != nil {
					pQuery = pQuery.Where("sales_transaction_id = ? OR sales_order_id = ?", inv.ID, *soID)
				} else {
					pQuery = pQuery.Where("sales_transaction_id = ?", inv.ID)
				}
				if err := pQuery.Find(&pendingPayments).Error; err == nil {
					for _, p := range pendingPayments {
						updatedNote := p.Notes
						if !strings.Contains(updatedNote, "[ADJUSTED]") {
							if updatedNote != "" {
								updatedNote = "[ADJUSTED] " + updatedNote
							} else {
								updatedNote = "[ADJUSTED] Potong Nota Retur"
							}
						}
						u.db.Model(&domain.SalesPayment{}).Where("id = ?", p.ID).Updates(map[string]interface{}{
							"amount": inv.TotalAmount,
							"notes":  updatedNote,
						})
					}
				}

				newNotes += fmt.Sprintf("[KREDIT NOTA RETUR: %s senilai %v]", ret.ReturnNo, ret.TotalAmount)
			} else if rType == "TUKAR_PRODUK" {
				// Faktur sudah disesuaikan dalam transaksi pengajuan; stok menunggu gudang.
			} else if rType == "TUKAR_LANGSUNG" {
				// Kuantitas & total tagihan faktur tetap sama (barang langsung diganti unit baru yang bagus)
				newNotes += fmt.Sprintf("[RETUR TUKAR BARANG: %s]", ret.ReturnNo)
			} else if rType == "GARANSI" {
				// Kuantitas & total tagihan faktur tetap sama (klaim garansi unit diganti baru)
				newNotes += fmt.Sprintf("[KLAIM GARANSI: %s]", ret.ReturnNo)
			}

			if rType != "TUKAR_PRODUK" {
				inv.Notes = &newNotes
				_ = u.salesRepo.Update(inv)
				invalidateReportsCache(context.Background())
			}
		}
	}

	// Jika Admin Gudang mencentang "Langsung terima fisik ke stok gudang", langsung eksekusi penerimaan fisik
	if req.AutoApproveWarehouse && ret.ID != uuid.Nil {
		_ = u.ApproveReturnByWarehouse(ret.ID, req.EmployeeID)
	}

	createdReturn, err := u.returnRepo.FindByID(ret.ID)
	if err == nil && u.sseHub != nil && createdReturn != nil {
		go func() {
			salesmanName := "Admin Nota"
			if createdReturn.Employee != nil {
				salesmanName = fmt.Sprintf("%s %s", createdReturn.Employee.FirstName, createdReturn.Employee.LastName)
			}
			event := sse.NotificationEvent{
				ID:        createdReturn.ID.String(),
				Type:      "sales_return",
				Title:     "Retur Penjualan Baru",
				Message:   fmt.Sprintf("Retur baru %s dari %s oleh %s sebesar Rp %.2f", createdReturn.ReturnNo, storeName, salesmanName, createdReturn.TotalAmount),
				CompanyID: branchID.String(),
				CreatedAt: time.Now().Format("2006-01-02 15:04:05"),
				Data: map[string]interface{}{
					"return_id":    createdReturn.ID.String(),
					"return_no":    createdReturn.ReturnNo,
					"salesman":     salesmanName,
					"store_name":   storeName,
					"total_amount": createdReturn.TotalAmount,
					"branch_id":    branchID.String(),
				},
			}
			if companyID != uuid.Nil {
				event.CompanyID = companyID.String()
			}
			if branchID != uuid.Nil {
				event.TargetBranchIDs = []string{branchID.String()}
				event.TargetRoles = []string{
					string(domain.RoleSuperAdmin),
					string(domain.RoleAdminGudang),
					string(domain.RoleSupervisorGudang),
					string(domain.RoleAdminNota),
					string(domain.RoleSupervisorSales),
					string(domain.RoleManager),
				}
			}
			_ = u.sseHub.PublishNotification(context.Background(), event)
		}()
	}
	return createdReturn, err
}

func (u *salesReturnUsecase) ApproveReturnByNota(id uuid.UUID, approvedByID uuid.UUID) error {
	return u.db.Transaction(func(tx *gorm.DB) error {
		returnRepo := repository.NewSalesReturnRepository(tx)
		salesRepo := repository.NewSalesTransactionRepository(tx)

		ret, err := returnRepo.FindByID(id)
		if err != nil {
			return err
		}
		if ret.Status != domain.ReturnStatusPending {
			return errors.New("hanya retur status PENDING yang bisa diproses oleh Admin Nota")
		}

		// 1. Jika ada Invoice, potong nominal Tagihan (Credit Note)
		// Namun jika ini adalah retur dari pengiriman (Delivery Drop Off), nominal invoice SUDAH terpotong di awal.
		// Dan potong tagihan HANYA jika return_type adalah POTONG_NOTA.
		if ret.TransactionID != nil && !strings.Contains(ret.Notes, "[DELIVERY_RETURN]") && ret.ReturnType == "POTONG_NOTA" {
			inv, err := salesRepo.FindByID(*ret.TransactionID)
			if err == nil && inv != nil {
				inv.TotalAmount -= ret.TotalAmount
				// Tambahkan catatan nota kredit ke notes invoice
				newNotes := ""
				if inv.Notes != nil {
					newNotes = *inv.Notes + "\n"
				}
				newNotes += fmt.Sprintf("[KREDIT NOTA: %s senilai %v]", ret.ReturnNo, ret.TotalAmount)
				inv.Notes = &newNotes

				if err := salesRepo.Update(inv); err != nil {
					return err
				}
			}
		}

		// 2. Jika ini adalah TUKAR_SUSULAN / TUKAR_PRODUK_SUSULAN, buat data sales_subsequent_items (leftover) agar bisa dikirim di Surat Jalan mendatang.
		if ret.ReturnType == "TUKAR_SUSULAN" || ret.ReturnType == "TUKAR_PRODUK_SUSULAN" {
			for _, item := range ret.Items {
				prodID := item.ProductID
				qty := item.Quantity
				if ret.ReturnType == "TUKAR_PRODUK_SUSULAN" && item.ReplacementProductID != nil {
					prodID = *item.ReplacementProductID
					qty = item.ReplacementQuantity
				}

				subsequentItem := &domain.SalesSubsequentItem{
					SalesOrderID:  ret.SalesOrderID,
					SalesReturnID: &ret.ID,
					ProductID:     prodID,
					Quantity:      qty,
					Status:        domain.SubsequentStatusPending,
				}
				if err := tx.Create(subsequentItem).Error; err != nil {
					return err
				}
			}
		}

		// Ubah status ke PENDING_WAREHOUSE agar Gudang bisa memverifikasi fisik barang
		ret.Status = domain.ReturnStatusPendingWarehouse
		return returnRepo.Update(ret)
	})
}

func (u *salesReturnUsecase) ApproveReturnByWarehouse(id uuid.UUID, approvedByID uuid.UUID) error {
	err := u.db.Transaction(func(tx *gorm.DB) error {
		returnRepo := repository.NewSalesReturnRepository(tx)
		wRepo := repository.NewWarehouseRepository(tx)

		ret, err := returnRepo.FindByID(id)
		if err != nil {
			return err
		}
		if ret.Status != domain.ReturnStatusPendingWarehouse {
			return errors.New("hanya retur status PENDING_WAREHOUSE yang bisa diterima oleh Gudang")
		}
		// Pengajuan lama dari sebelum perubahan alur ini masih disesuaikan di sini.
		if ret.ReturnType == "TUKAR_PRODUK" && ret.TransactionID != nil && !strings.Contains(ret.Notes, "[DELIVERY_RETURN]") && ret.InvoiceAdjustedAt == nil {
			if err := applyProductExchangeToInvoice(tx, ret); err != nil {
				return err
			}
		}

		// Cek apakah ini retur dari pengiriman armada
		isDeliveryReturn := strings.Contains(ret.Notes, "[DELIVERY_RETURN]")

		// 1. Proses pengembalian barang fisik ke gudang
		for _, item := range ret.Items {
			// Jika barang dikembalikan dalam kondisi GOOD dan ini dari pengiriman (supir),
			// barang tersebut TETAP berada di muatan armada (Gudang Berjalan) milik supir.
			// Oleh karena itu, kita tidak perlu memindahkannya ke gudang utama
			// dan tidak perlu memotongnya dari Gudang Berjalan.
			keepInDriverStock := item.Condition == "GOOD" && isDeliveryReturn

			if !keepInDriverStock {
				// Masukkan barang ke Gudang Utama (bisa DAMAGED atau GOOD)
				targetCondition := item.Condition
				if targetCondition == "" {
					targetCondition = "DAMAGED"
				}

				stock, err := wRepo.GetStock(ret.BranchID, item.ProductID, targetCondition, "DEFAULT")
				if err != nil {
					return err
				}
				stock.Quantity += item.Quantity
				if err := wRepo.UpdateStock(stock); err != nil {
					return err
				}

				// Log mutasi
				wRepo.CreateLog(&domain.WarehouseLog{
					BranchID:  ret.BranchID,
					ProductID: item.ProductID,
					Type:      "IN",
					Source:    "RETUR: " + ret.ReturnNo,
					Quantity:  item.Quantity,
				})

				// 1.1 Jika retur berasal dari delivery driver dan fisik dipindah ke gudang, potong stok Gudang Berjalannya
				if isDeliveryReturn && ret.EmployeeID != uuid.Nil {
					salesTransferRepo := repository.NewSalesTransferRepository(tx)
					salesStock, errStock := salesTransferRepo.GetSalesStock(ret.EmployeeID, item.ProductID)
					if errStock == nil && salesStock != nil {
						if salesStock.Quantity >= item.Quantity {
							salesStock.Quantity -= item.Quantity
						} else {
							salesStock.Quantity = 0
						}
						salesStock.UpdatedAt = time.Now()
						_ = salesTransferRepo.UpdateSalesStock(salesStock)
					}
				}
			}

			// 1.2 Jika ini adalah TUKAR_LANGSUNG / TUKAR_PRODUK / GARANSI, kita juga harus mengurangi stok barang bagus (GOOD) sebagai pengganti
			if ret.ReturnType == "TUKAR_LANGSUNG" || ret.ReturnType == "TUKAR_PRODUK" || ret.ReturnType == "GARANSI" {
				if isDeliveryReturn {
					// Jika berasal dari pengiriman, potong dari stok gudang berjalan supir (dari barang cadangan yang dibawa)
					if ret.EmployeeID != uuid.Nil {
						// KHUSUS TUKAR_LANGSUNG: Stok pengganti (produk yang sama) SUDAH terpotong di awal saat ConfirmDelivery (karena ActualQuantity diset sama dengan OrderedQuantity).
						// Jadi kita HANYA perlu memotong stok pengganti JIKA ini adalah TUKAR_PRODUK (karena produk penggantinya beda dan belum dipotong di ConfirmDelivery).
						if ret.ReturnType == "TUKAR_PRODUK" {
							salesTransferRepo := repository.NewSalesTransferRepository(tx)
							prodID := item.ProductID
							qty := item.Quantity
							if item.ReplacementProductID != nil {
								prodID = *item.ReplacementProductID
								qty = item.ReplacementQuantity
							}

							sStock, errStock := salesTransferRepo.GetSalesStock(ret.EmployeeID, prodID)
							if ret.ReturnType == "TUKAR_PRODUK" && (errStock != nil || sStock == nil || sStock.Quantity < qty) {
								return errors.New("stok barang pengganti di armada tidak mencukupi")
							}
							if errStock == nil && sStock != nil {
								if sStock.Quantity >= qty {
									sStock.Quantity -= qty
								} else {
									sStock.Quantity = 0
								}
								sStock.UpdatedAt = time.Now()
								_ = salesTransferRepo.UpdateSalesStock(sStock)
							}
						}
					}
				} else {
					// Jika bukan dari pengiriman (potong stok gudang utama)
					prodID := item.ProductID
					qty := item.Quantity
					if ret.ReturnType == "TUKAR_PRODUK" && item.ReplacementProductID != nil {
						prodID = *item.ReplacementProductID
						qty = item.ReplacementQuantity
					}

					goodStock, err := wRepo.GetStock(ret.BranchID, prodID, "GOOD", "DEFAULT")
					if ret.ReturnType == "TUKAR_PRODUK" && err == nil && goodStock != nil && goodStock.ID != uuid.Nil {
						err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(goodStock, "id = ?", goodStock.ID).Error
					}
					if ret.ReturnType == "TUKAR_PRODUK" && (err != nil || goodStock == nil || goodStock.Quantity < qty) {
						return errors.New("stok barang pengganti di gudang tidak mencukupi")
					}
					if err == nil && goodStock != nil {
						if goodStock.Quantity >= qty {
							goodStock.Quantity -= qty
						} else {
							goodStock.Quantity = 0
						}
						_ = wRepo.UpdateStock(goodStock)

						// Log mutasi keluar barang bagus
						wRepo.CreateLog(&domain.WarehouseLog{
							BranchID:  ret.BranchID,
							ProductID: prodID,
							Type:      "OUT",
							Source:    fmt.Sprintf("%s (PENGGANTI): %s", ret.ReturnType, ret.ReturnNo),
							Quantity:  qty,
						})
					}
				}
			}
		}

		now := time.Now()
		ret.Status = domain.ReturnStatusApproved
		ret.ApprovedAt = &now
		ret.ApprovedByID = &approvedByID
		if ret.ReturnType == "TUKAR_PRODUK" && ret.TransactionID != nil {
			return tx.Model(&domain.SalesReturn{}).Where("id = ?", ret.ID).Updates(map[string]interface{}{
				"status": ret.Status, "approved_at": ret.ApprovedAt,
				"approved_by_id": ret.ApprovedByID, "invoice_adjusted_at": ret.InvoiceAdjustedAt,
			}).Error
		}
		return returnRepo.Update(ret)
	})
	if err == nil {
		invalidateReportsCache(context.Background())
	}
	return err
}

// applyProductExchangeToInvoice berjalan dalam transaksi pembuatan retur.
// Pembayaran yang sudah masuk tidak berubah; stok fisik diproses oleh gudang kemudian.
func applyProductExchangeToInvoice(tx *gorm.DB, ret *domain.SalesReturn) error {
	var inv domain.SalesTransaction
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Preload("Items").First(&inv, "id = ?", *ret.TransactionID).Error; err != nil {
		return fmt.Errorf("faktur tukar produk tidak ditemukan: %w", err)
	}
	var pendingCount int64
	pendingQuery := tx.Model(&domain.SalesPayment{}).Where("payment_status = ?", "PENDING")
	if ret.SalesOrderID != nil {
		pendingQuery = pendingQuery.Where("sales_transaction_id = ? OR sales_order_id = ?", inv.ID, *ret.SalesOrderID)
	} else {
		pendingQuery = pendingQuery.Where("sales_transaction_id = ?", inv.ID)
	}
	if err := pendingQuery.Count(&pendingCount).Error; err != nil {
		return err
	}
	if pendingCount > 0 {
		return errors.New("selesaikan pembayaran faktur yang masih menunggu persetujuan sebelum mengajukan tukar produk")
	}

	for _, item := range ret.Items {
		if item.SalesItemID == nil || item.ReplacementProductID == nil || item.Quantity <= 0 || item.ReplacementQuantity <= 0 {
			return errors.New("rincian barang asal atau pengganti tukar produk tidak lengkap")
		}
		var original *domain.SalesItem
		for i := range inv.Items {
			if inv.Items[i].ID == *item.SalesItemID && inv.Items[i].ProductID == item.ProductID {
				original = &inv.Items[i]
				break
			}
		}
		if original == nil || original.Quantity < item.Quantity {
			return errors.New("jumlah barang asal pada faktur tidak lagi mencukupi; periksa retur lain")
		}
		originalNetPrice := original.PriceAtTransaction * (1 - original.DiscountPercent/100)
		if math.Abs(originalNetPrice-item.Price) > 0.01 {
			return errors.New("harga barang asal pada faktur telah berubah; ajukan ulang tukar produk")
		}
		original.Quantity -= item.Quantity
		original.Subtotal = float64(original.Quantity) * originalNetPrice
		if err := tx.Model(&domain.SalesItem{}).Where("id = ?", original.ID).Updates(map[string]interface{}{
			"quantity": original.Quantity, "subtotal": original.Subtotal,
		}).Error; err != nil {
			return err
		}
		replacement := domain.SalesItem{
			SalesTransactionID: inv.ID,
			ProductID:          *item.ReplacementProductID,
			Quantity:           item.ReplacementQuantity,
			PriceAtTransaction: item.ReplacementPrice,
			Subtotal:           float64(item.ReplacementQuantity) * item.ReplacementPrice,
		}
		if err := tx.Create(&replacement).Error; err != nil {
			return err
		}
		if err := tx.Model(&domain.SalesReturnItem{}).Where("id = ?", item.ID).
			Update("replacement_sales_item_id", replacement.ID).Error; err != nil {
			return err
		}
	}

	newTotal, status, err := exchangeInvoiceBalance(inv.TotalAmount, inv.PaidAmount, ret.PriceDifference)
	if err != nil {
		return err
	}
	notes := ""
	if inv.Notes != nil && *inv.Notes != "" {
		notes = *inv.Notes + "\n"
	}
	notes += fmt.Sprintf("[TUKAR PRODUK: %s, selisih %.2f]", ret.ReturnNo, ret.PriceDifference)
	if err := tx.Model(&domain.SalesTransaction{}).Where("id = ?", inv.ID).Updates(map[string]interface{}{
		"total_amount": newTotal, "payment_status": status, "notes": notes,
	}).Error; err != nil {
		return err
	}
	if ret.SalesOrderID != nil {
		if err := tx.Model(&domain.SalesOrder{}).Where("id = ?", *ret.SalesOrderID).Update("total_amount", newTotal).Error; err != nil {
			return err
		}
	}
	now := time.Now()
	ret.InvoiceAdjustedAt = &now
	return nil
}

func exchangeInvoiceBalance(total, paid, difference float64) (float64, string, error) {
	newTotal := math.Round((total+difference)*100) / 100
	if newTotal < 0 {
		return 0, "", errors.New("nilai barang yang diretur melebihi nilai faktur")
	}
	status := string(domain.PaymentStatusUnpaid)
	if paid >= newTotal {
		status = string(domain.PaymentStatusPaid)
	} else if paid > 0 {
		status = string(domain.PaymentStatusPartial)
	}
	return newTotal, status, nil
}

func reverseProductExchangeInvoice(tx *gorm.DB, ret *domain.SalesReturn) error {
	var inv domain.SalesTransaction
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Preload("Items").First(&inv, "id = ?", *ret.TransactionID).Error; err != nil {
		return fmt.Errorf("faktur terkait retur tidak ditemukan: %w", err)
	}
	for _, item := range ret.Items {
		var original, replacement *domain.SalesItem
		for i := range inv.Items {
			if inv.Items[i].ID == *item.SalesItemID {
				original = &inv.Items[i]
			}
			if inv.Items[i].ID == *item.ReplacementSalesItemID {
				replacement = &inv.Items[i]
			}
		}
		if original == nil || replacement == nil || replacement.Quantity != item.ReplacementQuantity || replacement.ProductID != *item.ReplacementProductID {
			return errors.New("barang pengganti sudah berubah; pengajuan tukar produk tidak dapat dibatalkan otomatis")
		}
		var dependentReturns int64
		if err := tx.Model(&domain.SalesReturnItem{}).
			Where("sales_item_id = ? AND sales_return_id <> ?", replacement.ID, ret.ID).
			Count(&dependentReturns).Error; err != nil {
			return err
		}
		if dependentReturns > 0 {
			return errors.New("barang pengganti sudah dipakai oleh retur lain; pengajuan ini tidak dapat dibatalkan")
		}
		original.Quantity += item.Quantity
		original.Subtotal = float64(original.Quantity) * original.PriceAtTransaction * (1 - original.DiscountPercent/100)
		if err := tx.Model(&domain.SalesItem{}).Where("id = ?", original.ID).Updates(map[string]interface{}{
			"quantity": original.Quantity, "subtotal": original.Subtotal,
		}).Error; err != nil {
			return err
		}
		if err := tx.Unscoped().Delete(&domain.SalesItem{}, "id = ?", replacement.ID).Error; err != nil {
			return err
		}
	}
	newTotal, status, err := exchangeInvoiceBalance(inv.TotalAmount, inv.PaidAmount, -ret.PriceDifference)
	if err != nil {
		return err
	}
	notes := ""
	if inv.Notes != nil && *inv.Notes != "" {
		notes = *inv.Notes + "\n"
	}
	notes += fmt.Sprintf("[PEMBATALAN TUKAR PRODUK: %s, selisih %.2f]", ret.ReturnNo, ret.PriceDifference)
	if err := tx.Model(&domain.SalesTransaction{}).Where("id = ?", inv.ID).Updates(map[string]interface{}{
		"total_amount": newTotal, "payment_status": status, "notes": notes,
	}).Error; err != nil {
		return err
	}
	if ret.SalesOrderID != nil {
		return tx.Model(&domain.SalesOrder{}).Where("id = ?", *ret.SalesOrderID).Update("total_amount", newTotal).Error
	}
	return nil
}

func (u *salesReturnUsecase) GetByBranch(branchID uuid.UUID, status string) ([]domain.SalesReturn, error) {
	return u.returnRepo.FindByBranch(branchID, status)
}

func (u *salesReturnUsecase) GetByID(id uuid.UUID) (*domain.SalesReturn, error) {
	return u.returnRepo.FindByID(id)
}

func (u *salesReturnUsecase) UpdateReturn(id uuid.UUID, req UpdateReturnRequest) (*domain.SalesReturn, error) {
	var updatedReturn *domain.SalesReturn
	err := u.db.Transaction(func(tx *gorm.DB) error {
		returnRepo := repository.NewSalesReturnRepository(tx)
		soRepo := repository.NewSalesOrderRepository(tx)
		salesRepo := repository.NewSalesTransactionRepository(tx)

		ret, err := returnRepo.FindByID(id)
		if err != nil {
			return err
		}
		if ret.Status != domain.ReturnStatusPending && ret.Status != domain.ReturnStatusPendingWarehouse {
			return errors.New("hanya retur dengan status PENDING atau PENDING_WAREHOUSE yang bisa diubah")
		}
		if ret.ReturnType == "TUKAR_PRODUK" {
			return errors.New("pengajuan tukar produk tidak dapat diedit; hapus pengajuan lalu buat ulang agar harga dan stok tetap sesuai")
		}

		// Hapus item lama
		if err := tx.Where("sales_return_id = ?", ret.ID).Delete(&domain.SalesReturnItem{}).Error; err != nil {
			return err
		}

		ret.Notes = req.Notes

		// Map item baru dan hitung total amount
		var returnItems []domain.SalesReturnItem
		var totalAmount float64
		newReturnMap := make(map[uuid.UUID]int)

		var so *domain.SalesOrder
		if ret.SalesOrderID != nil && *ret.SalesOrderID != uuid.Nil {
			so, _ = soRepo.FindByID(*ret.SalesOrderID)
		}

		var inv *domain.SalesTransaction
		if ret.TransactionID != nil && *ret.TransactionID != uuid.Nil {
			inv, _ = salesRepo.FindByID(*ret.TransactionID)
		}

		for _, itemReq := range req.Items {
			var price float64 = itemReq.Price
			if price <= 0 {
				if so != nil {
					for _, soItem := range so.Items {
						if soItem.ProductID == itemReq.ProductID {
							price = soItem.Price
							break
						}
					}
				} else if inv != nil {
					for _, invItem := range inv.Items {
						if invItem.ProductID == itemReq.ProductID {
							price = invItem.PriceAtTransaction
							break
						}
					}
				}
			}

			returnItems = append(returnItems, domain.SalesReturnItem{
				SalesReturnID: ret.ID,
				ProductID:     itemReq.ProductID,
				Quantity:      itemReq.Quantity,
				Price:         price,
				Reason:        itemReq.Reason,
				Condition:     itemReq.Condition,
			})
			totalAmount += float64(itemReq.Quantity) * price
			newReturnMap[itemReq.ProductID] = itemReq.Quantity
		}

		ret.Items = returnItems
		ret.TotalAmount = totalAmount

		// Jika retur bertipe POTONG_NOTA dan memiliki Invoice terkait, perbarui nominal Invoice dan SO Actual Quantity
		if ret.TransactionID != nil && ret.ReturnType == "POTONG_NOTA" {
			if inv != nil {
				var adjustmentNotes []string

				// Update SalesOrder items actual quantities and subtotals if SO exists
				if so != nil {
					for i := range so.Items {
						soItem := &so.Items[i]
						retQty := newReturnMap[soItem.ProductID]

						soItem.ActualQuantity = soItem.OrderedQuantity - retQty
						if soItem.ActualQuantity < 0 {
							soItem.ActualQuantity = 0
						}
						soItem.Subtotal = float64(soItem.ActualQuantity*soItem.PiecesPerUnit) * soItem.Price
						_ = tx.Save(soItem)

						// Update invoice item quantities
						for j := range inv.Items {
							invItem := &inv.Items[j]
							if invItem.ProductID == soItem.ProductID {
								invItem.Quantity = soItem.ActualQuantity
								invItem.Subtotal = float64(soItem.ActualQuantity) * invItem.PriceAtTransaction
							}
						}

						if retQty > 0 {
							var prod domain.Product
							prodName := "Produk"
							if errDb := tx.Select("name").Where("id = ?", soItem.ProductID).First(&prod).Error; errDb == nil {
								prodName = prod.Name
							}
							adjustmentNotes = append(adjustmentNotes, fmt.Sprintf("- %s: dipesan %d %s, terkirim %d %s, tidak terkirim %d %s (Disesuaikan Admin Gudang)",
								prodName, soItem.OrderedQuantity, soItem.Unit, soItem.ActualQuantity, soItem.Unit, retQty, soItem.Unit))
						}
					}
				}

				// Re-calculate invoice total amount
				var newInvTotal float64
				for _, invItem := range inv.Items {
					newInvTotal += invItem.Subtotal
					_ = tx.Save(&invItem)
				}
				inv.TotalAmount = newInvTotal - inv.DiscountAmount
				if inv.TotalAmount < 0 {
					inv.TotalAmount = 0
				}

				if len(adjustmentNotes) > 0 {
					notes := "[PENYESUAIAN PENGIRIMAN SEBAGIAN - DIEDIT ADMIN GUDANG]\n" + strings.Join(adjustmentNotes, "\n")
					inv.Notes = &notes
				}

				if err := salesRepo.Update(inv); err != nil {
					return err
				}

				// Sinkronisasi antrean Persetujuan Kasir (SalesPayment PENDING)
				var pendingPayments []domain.SalesPayment
				pQuery := tx.Where("payment_status = ?", "PENDING")
				if ret.SalesOrderID != nil {
					pQuery = pQuery.Where("sales_transaction_id = ? OR sales_order_id = ?", inv.ID, *ret.SalesOrderID)
				} else {
					pQuery = pQuery.Where("sales_transaction_id = ?", inv.ID)
				}
				if err := pQuery.Find(&pendingPayments).Error; err == nil {
					for _, p := range pendingPayments {
						updatedNote := p.Notes
						if !strings.Contains(updatedNote, "[ADJUSTED]") {
							if updatedNote != "" {
								updatedNote = "[ADJUSTED] " + updatedNote
							} else {
								updatedNote = "[ADJUSTED] Penyesuaian Pengiriman Sebagian"
							}
						}
						tx.Model(&domain.SalesPayment{}).Where("id = ?", p.ID).Updates(map[string]interface{}{
							"amount": inv.TotalAmount,
							"notes":  updatedNote,
						})
					}
				}
			}
		}

		if err := returnRepo.Update(ret); err != nil {
			return err
		}

		updatedReturn, err = returnRepo.FindByID(ret.ID)
		return err
	})

	if err != nil {
		return nil, err
	}
	return updatedReturn, nil
}

func (u *salesReturnUsecase) DeleteReturn(id uuid.UUID) error {
	ret, err := u.returnRepo.FindByID(id)
	if err != nil {
		return err
	}
	if ret.Status != domain.ReturnStatusPending && ret.Status != domain.ReturnStatusPendingWarehouse {
		return errors.New("hanya retur dengan status PENDING atau PENDING_WAREHOUSE yang dapat dihapus")
	}

	err = u.db.Transaction(func(tx *gorm.DB) error {
		salesRepo := repository.NewSalesTransactionRepository(tx)
		returnRepo := repository.NewSalesReturnRepository(tx)

		if ret.TransactionID != nil && ret.ReturnType == "POTONG_NOTA" {
			inv, err := salesRepo.FindByID(*ret.TransactionID)
			if err == nil && inv != nil {
				inv.TotalAmount += ret.TotalAmount
				if inv.PaymentStatus == "PAID" {
					inv.PaidAmount = inv.TotalAmount
				}

				// Kembalikan kuantitas item barang pada faktur
				for _, retItem := range ret.Items {
					for j := range inv.Items {
						if inv.Items[j].ProductID == retItem.ProductID {
							inv.Items[j].Quantity += retItem.Quantity
							disc := inv.Items[j].DiscountPercent / 100
							inv.Items[j].Subtotal = float64(inv.Items[j].Quantity) * inv.Items[j].PriceAtTransaction * (1 - disc)
							_ = tx.Save(&inv.Items[j])
						}
					}
				}

				// Sinkronisasi SalesOrder terkait jika ada
				if ret.SalesOrderID != nil {
					tx.Model(&domain.SalesOrder{}).Where("id = ?", *ret.SalesOrderID).Updates(map[string]interface{}{
						"total_amount": inv.TotalAmount,
					})
				}

				newNotes := ""
				if inv.Notes != nil {
					newNotes = *inv.Notes + "\n"
				}
				newNotes += fmt.Sprintf("[PEMBATALAN KREDIT NOTA RETUR: %s senilai %v]", ret.ReturnNo, ret.TotalAmount)
				inv.Notes = &newNotes
				_ = salesRepo.Update(inv)

				invalidateReportsCache(context.Background())
			}
		}
		if ret.TransactionID != nil && ret.ReturnType == "TUKAR_PRODUK" && ret.InvoiceAdjustedAt != nil {
			tracked := len(ret.Items) > 0
			for _, item := range ret.Items {
				tracked = tracked && item.SalesItemID != nil && item.ReplacementSalesItemID != nil
			}
			if tracked {
				if err := reverseProductExchangeInvoice(tx, ret); err != nil {
					return err
				}
			} else {
				inv, err := salesRepo.FindByID(*ret.TransactionID)
				if err != nil || inv == nil {
					return errors.New("faktur terkait retur tidak ditemukan")
				}
				for _, item := range ret.Items {
					for i := range inv.Items {
						if inv.Items[i].ProductID == item.ProductID {
							inv.Items[i].Quantity += item.Quantity
							inv.Items[i].Subtotal = float64(inv.Items[i].Quantity) * inv.Items[i].PriceAtTransaction * (1 - inv.Items[i].DiscountPercent/100)
							break
						}
					}
					if item.ReplacementProductID != nil {
						for i := range inv.Items {
							if inv.Items[i].ProductID == *item.ReplacementProductID && inv.Items[i].PriceAtTransaction == item.ReplacementPrice && inv.Items[i].Quantity >= item.ReplacementQuantity {
								inv.Items[i].Quantity -= item.ReplacementQuantity
								inv.Items[i].Subtotal = float64(inv.Items[i].Quantity) * inv.Items[i].PriceAtTransaction * (1 - inv.Items[i].DiscountPercent/100)
								break
							}
						}
					}
				}
				inv.TotalAmount -= ret.PriceDifference
				if inv.TotalAmount < 0 {
					inv.TotalAmount = 0
				}
				if inv.PaidAmount >= inv.TotalAmount {
					inv.PaymentStatus = "PAID"
				}
				if ret.SalesOrderID != nil {
					if err := tx.Model(&domain.SalesOrder{}).Where("id = ?", *ret.SalesOrderID).Update("total_amount", inv.TotalAmount).Error; err != nil {
						return err
					}
				}
				if err := salesRepo.Update(inv); err != nil {
					return err
				}
				if err := tx.Model(&domain.SalesPayment{}).
					Where("sales_transaction_id = ? AND payment_status = ?", inv.ID, "PENDING").
					Update("amount", inv.TotalAmount).Error; err != nil {
					return err
				}
			}
		}

		return returnRepo.Delete(id)
	})
	if err == nil {
		invalidateReportsCache(context.Background())
	}
	return err
}
