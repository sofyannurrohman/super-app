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
	midtransPkg "example.com/hris/backend/pkg/midtrans"
	"example.com/hris/backend/pkg/sse"
	"example.com/hris/backend/pkg/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// SalesOrderUsecase mendefinisikan kontrak bisnis untuk manajemen Pesanan Order (SO).
type SalesOrderUsecase interface {
	// Digunakan salesman (mobile) dan admin (web)
	CreateSO(req CreateSORequest) (*domain.SalesOrder, error)
	UploadNotaFisik(req UploadNotaFisikRequest) (*domain.SalesOrder, error)
	GetSOByEmployee(employeeID uuid.UUID) ([]domain.SalesOrder, error)

	// *** ALUR BARU ***
	// Digunakan Admin Nota (web) — verifikasi / tolak nota dari salesman
	AdminConfirmSO(id uuid.UUID, adminID uuid.UUID) error
	AdminRejectSO(id uuid.UUID, adminID uuid.UUID, notes string) error
	ApproveTaxDocuments(id uuid.UUID, adminID uuid.UUID, billingNote, taxInvoice, deliveryNote string, destinationBranchID uuid.UUID) error
	ApproveMTTaxDocuments(mtPoID uuid.UUID, adminID uuid.UUID, billingNote, taxInvoice, deliveryNote string, destinationBranchID uuid.UUID) error

	// Admin Nota: Adjust pesanan (diskon, qty item) sebelum masuk faktur
	AdjustSO(id uuid.UUID, req AdjustSORequest) (*domain.SalesOrder, error)
	GetStoreDebtHistory(storeID uuid.UUID) ([]domain.SalesTransaction, error)

	// Digunakan Driver (mobile) — konfirmasi pengiriman & tagih pembayaran per nota
	ConfirmDelivery(req ConfirmDeliveryRequest) (*domain.SalesTransaction, error)
	CollectPayment(req CollectPaymentRequest) (*domain.SalesOrder, error)

	// MT Settlement
	GetMTSettlementQueue() ([]domain.SalesPayment, error)
	ApproveMTSettlement(paymentID uuid.UUID, adminID uuid.UUID) error
	UpdateMTOrderStatus(id uuid.UUID, status string) error

	// Digunakan semua role (web)
	GetSOByBranch(branchIDs []uuid.UUID, status string, salesType string, paymentStatus string, startDate string, endDate string, employeeID uuid.UUID, companyID uuid.UUID, search string, page int, limit int) ([]domain.SalesOrder, int64, *repository.SOSummary, error)
	GetSOByStore(storeID uuid.UUID, page int, limit int) ([]domain.SalesOrder, int64, error)
	GetSOByID(id uuid.UUID) (*domain.SalesOrder, error)
	CancelSO(id uuid.UUID) error
	DeleteSO(id uuid.UUID) error
	SoftDeleteSO(id uuid.UUID, deletedByID uuid.UUID, reason string) error
	DeleteDraftSOBySales(id, employeeID uuid.UUID) error
	RestoreSO(id uuid.UUID) error

	// Legacy methods — dipertahankan agar tidak break fitur lama
	ConfirmSO(id uuid.UUID, confirmedByID uuid.UUID) error
	RejectSO(id uuid.UUID, rejectedByID uuid.UUID, notes string) error
	ProcessByWarehouse(req ProcessByWarehouseRequest) error
	ConfirmPOD(req ConfirmPODRequest) error
	ConvertToInvoice(id uuid.UUID, companyID uuid.UUID) (*domain.SalesTransaction, error)
	ConvertDirectToInvoice(id uuid.UUID, adminID uuid.UUID) (*domain.SalesTransaction, error)
	OverrideBackorder(id uuid.UUID) error
	RecheckStock(id uuid.UUID) (*domain.SalesOrder, error)
	SuperAdminFastTrackComplete(id uuid.UUID, superAdminID uuid.UUID) (*domain.SalesOrder, error)
	SuperAdminBulkComplete(req BulkSuperAdminCompleteRequest, superAdminID uuid.UUID) (*BulkCompleteResult, error)
	GetPendingStats(branchID *uuid.UUID, branchIDs []uuid.UUID, companyID *uuid.UUID) (*PendingSOStats, error)
}

type PendingSOStats struct {
	TotalPending       int64   `json:"total_pending"`
	ConfirmedCount     int64   `json:"confirmed_count"`
	WaitingStockCount  int64   `json:"waiting_stock_count"`
	DraftCount         int64   `json:"draft_count"`
	OtherPendingCount  int64   `json:"other_pending_count"`
	TotalPendingAmount float64 `json:"total_pending_amount"`
}

type BulkCompleteResult struct {
	TotalProcessed int      `json:"total_processed"`
	SuccessCount   int      `json:"success_count"`
	FailedCount    int      `json:"failed_count"`
	CompletedIDs   []string `json:"completed_ids"`
	Errors         []string `json:"errors,omitempty"`
}

type BulkSuperAdminCompleteRequest struct {
	BranchID  *uuid.UUID  `json:"branch_id"`
	BranchIDs []uuid.UUID `json:"branch_ids"`
	CompanyID *uuid.UUID  `json:"company_id"`
	SOIDs     []uuid.UUID `json:"so_ids"`
}

// --- Request DTOs (Baru - Mobile POS) ---

type ConfirmDeliveryItem struct {
	ProductID        uuid.UUID `json:"product_id" binding:"required"`
	ReturnedQuantity int       `json:"returned_quantity"` // 0 jika semua terkirim sempurna
	ReturnReason     string    `json:"return_reason"`     // RUSAK / TOLAK_CUSTOMER / KADALUARSA / LAINNYA
	ReturnType       string    `json:"return_type"`       // POTONG_NOTA, TUKAR_LANGSUNG, TUKAR_SUSULAN, TUKAR_PRODUK, GARANSI
	IsGoodCondition  bool      `json:"is_good_condition"` // TRUE jika barang dikembalikan tapi kondisinya masih bagus

	// Khusus TUKAR_PRODUK: produk yang diberikan sebagai pengganti (berbeda dari produk yang diretur)
	ReplacementProductID *uuid.UUID `json:"replacement_product_id"`
	ReplacementQuantity  int        `json:"replacement_quantity"`

	// Khusus GARANSI: nomor klaim resmi & foto bukti kerusakan
	ClaimReferenceNo  string `json:"claim_reference_no"`
	DamageEvidenceURL string `json:"damage_evidence_url"`
}

type ConfirmDeliveryRequest struct {
	SOID              uuid.UUID             `json:"so_id" binding:"required"`
	ReceivedBy        string                `json:"received_by" binding:"required"`
	PODImageURL       string                `json:"pod_image_url"`
	SignatureImageURL string                `json:"signature_image_url"`
	Notes             string                `json:"notes"`
	ReturnedItems     []ConfirmDeliveryItem `json:"returned_items"` // List of items being returned instantly
	RequestItems      []SOItemRequest       `json:"request_items"`  // List of new items requested by customer
	EmployeeID        uuid.UUID             `json:"-"`              // Resolved from token for SalesStock deduction
}

type CollectPaymentRequest struct {
	SOID                 uuid.UUID  `json:"so_id" binding:"required"`
	Amount               float64    `json:"amount"`                            // Diperbolehkan 0 untuk pembayaran TEMPO murni
	PaymentMethod        string     `json:"payment_method" binding:"required"` // CASH, QRIS, VA, TEMPO
	PaymentBank          string     `json:"payment_bank"`                      // For VA: 'bca', 'bni', 'bri'
	CollectedBy          string     `json:"collected_by"`
	EmployeeID           uuid.UUID  `json:"employee_id"` // Resolved from token in handler
	PaymentDueDate       *time.Time `json:"payment_due_date"`
	TransferProofURL     string     `json:"transfer_proof_url"`
	CompanyBankAccountID *uuid.UUID `json:"company_bank_account_id"`
}

// --- Request DTOs (Legacy) ---

type ConfirmPODRequest struct {
	SOID        uuid.UUID `json:"so_id" binding:"required"`
	ReceivedBy  string    `json:"received_by" binding:"required"`
	PODImageURL string    `json:"pod_image_url"`
}

type ProcessByWarehouseRequest struct {
	SOID          uuid.UUID          `json:"so_id" binding:"required"`
	ProcessedByID uuid.UUID          `json:"processed_by_id"`
	Items         []WarehouseItemQty `json:"items" binding:"required,min=1"`
}

type WarehouseItemQty struct {
	ProductID uuid.UUID `json:"product_id" binding:"required"`
	ActualQty int       `json:"actual_qty" binding:"required"`
}

// AdjustSORequest digunakan Admin Nota untuk mengadjust pesanan sebelum faktur.
type AdjustSORequest struct {
	// Diskon dalam Rupiah. Diskon mengurangi TotalAmount.
	DiscountAmount utils.FlexFloat `json:"discount_amount"`
	// Catatan penyesuaian
	AdjustNotes string `json:"adjust_notes"`
	// Item-item pesanan final setelah penyesuaian.
	Items []AdjustSOItem `json:"items"`
	// Item-item bonus
	BonusItems []AdjustSOBonusItem `json:"bonus_items"`
}

type AdjustSOItem struct {
	ProductID       uuid.UUID       `json:"product_id"`
	OrderedQuantity int             `json:"ordered_quantity"`
	Unit            string          `json:"unit"`
	PiecesPerUnit   int             `json:"pieces_per_unit"`
	Price           utils.FlexFloat `json:"price"`
	DiscountPercent utils.FlexFloat `json:"discount_percent"`
}

type AdjustSOBonusItem struct {
	BonusInventoryID uuid.UUID `json:"bonus_inventory_id"`
	Quantity         int       `json:"quantity"`
	Notes            string    `json:"notes"`
}

type salesOrderUsecase struct {
	soRepo         repository.SalesOrderRepository
	warehouseRepo  repository.WarehouseRepository
	deliveryRepo   repository.DeliveryRepository
	salesRepo      repository.SalesTransactionRepository
	db             *gorm.DB
	midtransClient *midtransPkg.MidtransClient
	sseHub         *sse.Hub
	fcmClient      FcmUsecase
}

func NewSalesOrderUsecase(
	soRepo repository.SalesOrderRepository,
	warehouseRepo repository.WarehouseRepository,
	deliveryRepo repository.DeliveryRepository,
	salesRepo repository.SalesTransactionRepository,
	db *gorm.DB,
	midtransClient *midtransPkg.MidtransClient,
	sseHub *sse.Hub,
	fcmClient FcmUsecase,
) SalesOrderUsecase {
	return &salesOrderUsecase{soRepo, warehouseRepo, deliveryRepo, salesRepo, db, midtransClient, sseHub, fcmClient}
}

// --- Request DTOs ---

type CreateSORequest struct {
	BranchID             uuid.UUID       `json:"branch_id" binding:"required"`
	CompanyID            uuid.UUID       `json:"company_id" binding:"required"`
	EmployeeID           uuid.UUID       `json:"employee_id" binding:"required"`
	StoreID              uuid.UUID       `json:"store_id" binding:"required"`
	StoreCategory        string          `json:"store_category"`
	Notes                string          `json:"notes"`
	Items                []SOItemRequest `json:"items" binding:"required,min=1"`
	VisitID              *uuid.UUID      `json:"visit_id"`
	ClientID             *string         `json:"client_id"`
	Latitude             *float64        `json:"latitude"`
	Longitude            *float64        `json:"longitude"`
	PaymentMethod        string          `json:"payment_method"`
	CompanyBankAccountID *uuid.UUID      `json:"company_bank_account_id"`

	// --- Modern Trade (MT) ---
	// IsMT: set true jika SO ini dibuat oleh Sales MT.
	// Faktur dan omzetnya dikelola oleh pusat/pabrik (bukan cabang).
	IsMT bool `json:"is_mt"`
	// FactoryCompanyID: company_id pabrik/pusat (wajib diisi jika IsMT=true)
	FactoryCompanyID *uuid.UUID `json:"factory_company_id"`

	// --- Distributor Order ---
	IsDistributorOrder bool       `json:"is_distributor_order"`
	DistributorStoreID *uuid.UUID `json:"distributor_store_id"`

	IsPhysicalNota       bool                 `json:"is_physical_nota"`
	PhysicalNotaImageURL string               `json:"physical_nota_image_url"`
	TotalAmount          float64              `json:"total_amount"`
	BonusItems           []SOBonusItemRequest `json:"bonus_items"`
	OrderDate            time.Time            `json:"order_date"`
}

type SOBonusItemRequest struct {
	BonusInventoryID uuid.UUID `json:"bonus_inventory_id"`
	Quantity         int       `json:"quantity"`
	Notes            string    `json:"notes"`
}

type UploadNotaFisikRequest struct {
	BranchID             uuid.UUID  `json:"branch_id" binding:"required"`
	CompanyID            uuid.UUID  `json:"company_id" binding:"required"`
	EmployeeID           uuid.UUID  `json:"employee_id" binding:"required"`
	StoreID              uuid.UUID  `json:"store_id" binding:"required"`
	VisitID              *uuid.UUID `json:"visit_id"`
	Latitude             *float64   `json:"latitude"`
	Longitude            *float64   `json:"longitude"`
	Notes                string     `json:"notes"`
	PhysicalNotaImageURL string     `json:"physical_nota_image_url" binding:"required"`
}

type SOItemRequest struct {
	ProductID       uuid.UUID `json:"product_id" binding:"required"`
	OrderedQuantity int       `json:"ordered_quantity" binding:"required,min=1"`
	Unit            string    `json:"unit"`
	PiecesPerUnit   int       `json:"pieces_per_unit"`
	Price           float64   `json:"price"`
	// DiscountPercent: Diskon per item dalam persen (0-100).
	// Sales dapat menerapkan diskon berbeda untuk setiap produk.
	DiscountPercent float64 `json:"discount_percent"`
}

// --- Implementations ---

// CreateSO membuat Sales Order baru dengan status DRAFT.
// Pada tahap ini TIDAK ada perubahan stok sama sekali.
func (u *salesOrderUsecase) CreateSO(req CreateSORequest) (*domain.SalesOrder, error) {
	now := req.OrderDate
	if now.IsZero() {
		now = time.Now()
	} else if now.UTC().Hour() == 0 && now.UTC().Minute() == 0 && now.UTC().Second() == 0 {
		// Jika client (Web / Mobile) hanya mengirimkan tanggal (waktu default 00:00:00 UTC),
		// sisipkan jam, menit, detik saat ini secara lokal untuk menghindari waktu terpotong menjadi 07:00 WIB.
		curr := time.Now()
		now = time.Date(now.Year(), now.Month(), now.Day(), curr.Hour(), curr.Minute(), curr.Second(), curr.Nanosecond(), curr.Location())
	}

	// 1. Cek Branch dari Karyawan (Prioritas Utama: Operating Branch jika ada, atau Home Branch)
	var emp domain.Employee
	if err := u.db.Where("id = ?", req.EmployeeID).First(&emp).Error; err == nil {
		if emp.OperatingBranchID != nil && *emp.OperatingBranchID != uuid.Nil {
			req.BranchID = *emp.OperatingBranchID
		} else if req.BranchID == uuid.Nil && emp.BranchID != nil {
			req.BranchID = *emp.BranchID
		}
	}

	// 2. Fallback ke Branch dari Toko (Hanya jika Karyawan tidak punya cabang)
	if req.BranchID == uuid.Nil {
		var store domain.Store
		if err := u.db.Preload("AssignedEmployee").Where("id = ?", req.StoreID).First(&store).Error; err == nil {
			if store.AssignedEmployee != nil && store.AssignedEmployee.BranchID != nil {
				req.BranchID = *store.AssignedEmployee.BranchID
			}
		}
	}

	// Resolve and Complete Visit
	var visit *domain.SalesVisit
	if req.VisitID != nil {
		visit = &domain.SalesVisit{}
		if err := u.db.First(visit, "id = ?", *req.VisitID).Error; err != nil {
			visit = nil
		}
	} else {
		start := now.Add(-24 * time.Hour)
		end := now.Add(24 * time.Hour)
		visit = &domain.SalesVisit{}
		err := u.db.Where("employee_id = ? AND store_id = ? AND check_in_time >= ? AND check_in_time <= ? AND status != 'SKIPPED' AND type != 'SKIP'", req.EmployeeID, req.StoreID, start, end).Order("check_in_time DESC").First(visit).Error
		if err != nil {
			visit = nil
		}
	}
	if visit != nil {
		req.VisitID = &visit.ID
		checkOutTime := time.Now()
		visit.CheckOutTime = &checkOutTime
		visit.Status = "COMPLETED"
		if visit.Notes == "Deleted from schedule" {
			visit.Notes = "Kunjungan sales & order"
		}
		if visit.Type == "SKIP" {
			visit.Type = "CHECKIN"
		}
		if visit.CheckInTime.After(checkOutTime) {
			visit.CheckInTime = checkOutTime.Add(-15 * time.Minute)
		}
		u.db.Save(visit)
	} else if req.StoreID != uuid.Nil {
		// Auto-create visit for SO created without prior check-in (e.g. outside radius or direct SO)
		// so every SO has a valid visit representation in Kunjungan Sales
		checkInTime := now.Add(-5 * time.Minute)
		checkOutTime := now
		lat := 0.0
		if req.Latitude != nil {
			lat = *req.Latitude
		}
		lng := 0.0
		if req.Longitude != nil {
			lng = *req.Longitude
		}
		newVisit := &domain.SalesVisit{
			ID:           uuid.New(),
			EmployeeID:   req.EmployeeID,
			StoreID:      req.StoreID,
			CheckInTime:  checkInTime,
			CheckOutTime: &checkOutTime,
			Latitude:     lat,
			Longitude:    lng,
			Status:       "COMPLETED",
			Type:         "CHECKIN",
			Notes:        "Order langsung (SO)",
		}
		if errDb := u.db.Create(newVisit).Error; errDb == nil {
			req.VisitID = &newVisit.ID
		}
	}

	var company domain.Company
	prefix := ""
	if err := u.db.First(&company, "id = ?", req.CompanyID).Error; err == nil {
		if company.Code != "" {
			prefix = company.Code
		} else {
			words := strings.Fields(company.Name)
			if len(words) > 0 {
				name := words[len(words)-1] // ambil kata terakhir untuk menghindari "PT"
				if len(name) > 5 {
					prefix = strings.ToUpper(name[:5])
				} else {
					prefix = strings.ToUpper(name)
				}
			}
		}
	}
	var store domain.Store
	if err := u.db.First(&store, "id = ?", req.StoreID).Error; err == nil {
		var soCount int64
		u.db.Model(&domain.SalesOrder{}).Where("store_id = ? AND status != 'CANCELLED' AND status != 'REJECTED'", req.StoreID).Count(&soCount)

		var trxCount int64
		u.db.Model(&domain.SalesTransaction{}).Where("store_id = ? AND status = 'VERIFIED'", req.StoreID).Count(&trxCount)

		totalOrders := soCount + trxCount
		if totalOrders >= 3 {
			req.StoreCategory = "TOKO_LAMA"
			if store.Category != "TOKO_LAMA" {
				u.db.Model(&store).Update("category", "TOKO_LAMA")
			}
		} else {
			req.StoreCategory = "TOKO_BARU"
			if store.Category != "TOKO_BARU" {
				u.db.Model(&store).Update("category", "TOKO_BARU")
			}
		}
	} else {
		// Fallback if not found or error, default to the one from request or TOKO_LAMA
		if req.StoreCategory == "" {
			req.StoreCategory = "TOKO_LAMA"
		}
	}

	paymentMethod := req.PaymentMethod
	if paymentMethod == "" {
		paymentMethod = "SALES_ORDER"
	}

	// Determine effective salesman for omzet / KPI:
	// If the order is created by supervisor / management on a store with an assigned salesman (PIC),
	// the order's EmployeeID is set to the store's assigned salesman so omzet & KPI credit goes to the PIC.
	effectiveSalesmanID := req.EmployeeID
	var creatorEmp domain.Employee
	if err := u.db.Preload("JobPosition").Where("id = ?", req.EmployeeID).First(&creatorEmp).Error; err == nil {
		if creatorEmp.JobPosition != nil {
			title := strings.ToLower(creatorEmp.JobPosition.Title)
			isManagementOrSpv := strings.Contains(title, "direktur") || strings.Contains(title, "director") ||
				strings.Contains(title, "manager") || strings.Contains(title, "owner") ||
				strings.Contains(title, "superadmin") || strings.Contains(title, "spv") ||
				strings.Contains(title, "supervisor")
			if isManagementOrSpv {
				var currentStore domain.Store
				if err := u.db.Select("id, assigned_employee_id").Where("id = ?", req.StoreID).First(&currentStore).Error; err == nil {
					if currentStore.AssignedEmployeeID != nil && *currentStore.AssignedEmployeeID != uuid.Nil {
						effectiveSalesmanID = *currentStore.AssignedEmployeeID
						fmt.Printf("[SO-AUDIT] Order created by SPV/Management %s credited to Store PIC %s\n", req.EmployeeID, effectiveSalesmanID)
					}
				}
			}
		}
	}

	so := &domain.SalesOrder{
		BranchID:             req.BranchID,
		CompanyID:            req.CompanyID,
		EmployeeID:           effectiveSalesmanID,
		StoreID:              req.StoreID,
		StoreCategory:        req.StoreCategory,
		Status:               domain.SOStatusDraft,
		Notes:                req.Notes,
		OrderDate:            now,
		VisitID:              req.VisitID,
		ClientID:             req.ClientID,
		Latitude:             req.Latitude,
		Longitude:            req.Longitude,
		PaymentMethod:        paymentMethod,
		CompanyBankAccountID: req.CompanyBankAccountID,
		// Modern Trade flag
		IsMT:             req.IsMT,
		FactoryCompanyID: req.FactoryCompanyID,
		// Distributor flag
		IsDistributorOrder:   req.IsDistributorOrder,
		DistributorStoreID:   req.DistributorStoreID,
		IsPhysicalNota:       req.IsPhysicalNota,
		PhysicalNotaImageURL: req.PhysicalNotaImageURL,
		PaymentStatus:        "UNPAID",
	}

	var totalAmount float64
	if len(req.Items) == 0 && req.TotalAmount > 0 {
		totalAmount = req.TotalAmount
	} else {
		for _, item := range req.Items {
			ppu := item.PiecesPerUnit
			if ppu <= 0 {
				ppu = 1
			}
			unit := item.Unit
			if unit == "" {
				unit = "PCS"
			}
			totalQty := item.OrderedQuantity * ppu

			// Terapkan diskon per item jika ada
			discountPct := item.DiscountPercent
			if discountPct < 0 {
				discountPct = 0
			} else if discountPct > 100 {
				discountPct = 100
			}
			effectivePrice := item.Price * (1 - discountPct/100)
			subtotal := float64(totalQty) * effectivePrice
			totalAmount += subtotal
			so.Items = append(so.Items, domain.SalesOrderItem{
				ProductID:       item.ProductID,
				Quantity:        totalQty,
				OrderedQuantity: item.OrderedQuantity,
				Unit:            unit,
				PiecesPerUnit:   ppu,
				Price:           item.Price,
				DiscountPercent: discountPct,
				Subtotal:        subtotal,
			})
		}
	}

	for _, bonusReq := range req.BonusItems {
		so.BonusItems = append(so.BonusItems, domain.SalesOrderBonusItem{
			BonusInventoryID: bonusReq.BonusInventoryID,
			Quantity:         bonusReq.Quantity,
			Notes:            bonusReq.Notes,
		})
	}

	// Apply Global Discount Rule from Smart Auto-Discount Engine
	discountEngine := NewDiscountEngine(u.db)
	globalDiscount, err := discountEngine.EvaluateCart(paymentMethod, so.Items)
	if err != nil {
		fmt.Printf("[DiscountEngine] failed to evaluate discount: %v\n", err)
	}

	so.DiscountAmount = globalDiscount
	so.TotalAmount = totalAmount - globalDiscount

	if err := u.soRepo.CreateWithNextSONumber(so, prefix); err != nil {
		return nil, err
	}

	// Trigger SSE if it's a physical nota
	if so.IsPhysicalNota && so.PhysicalNotaImageURL != "" && u.sseHub != nil {
		go func() {
			var salesmanName string
			var employee domain.Employee
			if err := u.db.First(&employee, "id = ?", req.EmployeeID).Error; err == nil {
				salesmanName = employee.FirstName + " " + employee.LastName
			}
			var targetBranchIDs []string
			if so.BranchID != uuid.Nil {
				targetBranchIDs = []string{so.BranchID.String()}
			}
			event := sse.NotaFisikEvent{
				TransactionID:   so.ID.String(),
				ReceiptNo:       so.SONumber,
				Salesman:        salesmanName,
				ImageURL:        so.PhysicalNotaImageURL,
				CompanyID:       so.CompanyID.String(),
				CreatedAt:       so.CreatedAt.Format("2006-01-02 15:04:05"),
				StoreName:       store.Name,
				TargetBranchIDs: targetBranchIDs,
				TargetRoles: []string{
					string(domain.RoleSuperAdmin),
					string(domain.RoleCashier),
					string(domain.RoleCashierPusat),
					string(domain.RoleAdminKasir),
					string(domain.RoleManager),
					string(domain.RoleAdminNota),
				},
			}
			if pubErr := u.sseHub.PublishNotaFisik(context.Background(), event); pubErr != nil {
				fmt.Printf("[SSE] Failed to publish nota_fisik event in CreateSO: %v\n", pubErr)
			}
		}()
	}

	createdSO, err := u.soRepo.FindByID(so.ID)
	if err == nil && u.sseHub != nil && createdSO != nil {
		go func() {
			salesmanName := "Salesman"
			if createdSO.Employee != nil {
				salesmanName = fmt.Sprintf("%s %s", createdSO.Employee.FirstName, createdSO.Employee.LastName)
			}
			storeName := "Toko"
			if createdSO.Store != nil {
				storeName = createdSO.Store.Name
			}
			event := sse.NotificationEvent{
				ID:        createdSO.ID.String(),
				Type:      "sales_order",
				Title:     "Sales Order Baru",
				Message:   fmt.Sprintf("Pesanan baru %s dari toko %s oleh %s sebesar Rp %.2f", createdSO.SONumber, storeName, salesmanName, createdSO.TotalAmount),
				CompanyID: createdSO.CompanyID.String(),
				CreatedAt: time.Now().Format("2006-01-02 15:04:05"),
				Data: map[string]interface{}{
					"so_id":        createdSO.ID.String(),
					"so_number":    createdSO.SONumber,
					"salesman":     salesmanName,
					"store_name":   storeName,
					"total_amount": createdSO.TotalAmount,
					"branch_id":    createdSO.BranchID.String(),
				},
				TargetRoles: []string{
					string(domain.RoleSuperAdmin),
					string(domain.RoleAdminNota),
					string(domain.RoleSupervisorSales),
					string(domain.RoleSupervisorSalesMT),
					string(domain.RoleManager),
				},
			}
			if createdSO.BranchID != uuid.Nil {
				event.TargetBranchIDs = []string{createdSO.BranchID.String()}
			}
			_ = u.sseHub.PublishNotification(context.Background(), event)
		}()
	}
	return createdSO, err
}

// ConfirmSO mengkonfirmasi SO.
// Status berubah menjadi CONFIRMED untuk siap diproses oleh tim gudang/logistik.
func (u *salesOrderUsecase) ConfirmSO(id uuid.UUID, confirmedByID uuid.UUID) error {
	return u.db.Transaction(func(tx *gorm.DB) error {
		soRepo := repository.NewSalesOrderRepository(tx)

		so, err := soRepo.FindByID(id)
		if err != nil {
			return err
		}
		if so.Status != domain.SOStatusDraft {
			return errors.New("hanya SO dengan status DRAFT yang bisa dikonfirmasi")
		}

		now := time.Now()
		so.Status = domain.SOStatusConfirmed
		so.ConfirmedAt = &now
		so.ConfirmedByID = &confirmedByID
		return soRepo.Update(so)
	})
}

// CancelSO membatalkan SO.
func (u *salesOrderUsecase) CancelSO(id uuid.UUID) error {
	return u.db.Transaction(func(tx *gorm.DB) error {
		soRepo := repository.NewSalesOrderRepository(tx)

		so, err := soRepo.FindByID(id)
		if err != nil {
			return err
		}
		if so.Status == domain.SOStatusConverted || so.Status == domain.SOStatusShipped || so.Status == domain.SOStatusDelivered {
			return errors.New("SO yang sudah diproses tidak dapat dibatalkan")
		}

		so.Status = domain.SOStatusCancelled
		return soRepo.Update(so)
	})
}

// RejectSO menolak SO oleh admin gudang.
func (u *salesOrderUsecase) RejectSO(id uuid.UUID, rejectedByID uuid.UUID, notes string) error {
	return u.db.Transaction(func(tx *gorm.DB) error {
		soRepo := repository.NewSalesOrderRepository(tx)

		so, err := soRepo.FindByID(id)
		if err != nil {
			return err
		}
		if so.Status != domain.SOStatusDraft && so.Status != domain.SOStatusWaitingWarehouse && so.Status != domain.SOStatusProcessing && so.Status != domain.SOStatusWaitingStock {
			return errors.New("hanya SO dengan status DRAFT, antrian, atau proses gudang yang bisa ditolak")
		}

		now := time.Now()
		so.Status = domain.SOStatusRejected
		so.RejectedAt = &now
		so.RejectedByID = &rejectedByID
		so.Notes = notes
		return soRepo.Update(so)
	})
}

// ProcessByWarehouse melakukan validasi fisik barang oleh gudang.
// Hasilnya adalah Surat Jalan (DO) dan status berubah menjadi SHIPPED.
func (u *salesOrderUsecase) ProcessByWarehouse(req ProcessByWarehouseRequest) error {
	return u.db.Transaction(func(tx *gorm.DB) error {
		soRepo := repository.NewSalesOrderRepository(tx)
		wRepo := repository.NewWarehouseRepository(tx)

		so, err := soRepo.FindByID(req.SOID)
		if err != nil {
			return err
		}
		if so.Status != domain.SOStatusWaitingWarehouse && so.Status != domain.SOStatusProcessing && so.Status != domain.SOStatusApproved && so.Status != domain.SOStatusConfirmed {
			return errors.New("SO tidak dalam status untuk diproses gudang")
		}

		// Generate Nomor Surat Jalan (DO)
		now := time.Now()
		// Untuk kemudahan, kita pakai format DO-SO_NUMBER
		doNo := "DO-" + so.SONumber

		// Map input qty dari request
		qtyMap := make(map[uuid.UUID]int)
		for _, it := range req.Items {
			qtyMap[it.ProductID] = it.ActualQty
		}

		// Update items dan mutasi stok fisik
		for i, item := range so.Items {
			actualQty, exists := qtyMap[item.ProductID]
			if !exists {
				actualQty = 0 // Jika tidak diinput gudang, anggap 0
			}

			so.Items[i].ActualQuantity = actualQty

			// Potong Stok Fisik permanen dengan FEFO
			if actualQty > 0 {
				batches, err := wRepo.GetAvailableBatches(so.BranchID, item.ProductID, "GOOD")
				if err != nil {
					return err
				}
				remainingToDeduct := actualQty
				for idx := range batches {
					b := &batches[idx]
					if b.Quantity <= 0 {
						continue
					}
					deduct := b.Quantity
					if deduct > remainingToDeduct {
						deduct = remainingToDeduct
					}
					b.Quantity -= deduct
					if err := wRepo.UpdateStock(b); err != nil {
						return err
					}
					remainingToDeduct -= deduct
					if remainingToDeduct <= 0 {
						break
					}
				}

				// Catat Log Mutasi Gudang (Surat Jalan)
				_ = wRepo.CreateLog(&domain.WarehouseLog{
					BranchID:  so.BranchID,
					ProductID: item.ProductID,
					Type:      "OUT",
					Source:    "SURAT_JALAN: " + doNo,
					Quantity:  actualQty,
				})
			}
		}

		if so.IsMT {
			so.Status = "WAREHOUSE_RECEIVED"
		} else {
			so.Status = domain.SOStatusShipped
		}
		so.DeliveryOrderNo = &doNo
		so.ShippedAt = &now

		if err := soRepo.Update(so); err != nil {
			return err
		}

		// Sync ke MT Purchase Order jika ini adalah PO MT
		if so.IsMT && so.SONumber != "" {
			if err := tx.Model(&domain.MTPurchaseOrder{}).Where("po_number = ?", so.SONumber).Update("status", so.Status).Error; err != nil {
				return err
			}

			// Buat record DeliveryBatch khusus untuk MT agar tampil di Surat Jalan Gudang
			batch := domain.DeliveryBatch{
				CompanyID:       &so.CompanyID,
				BranchID:        &so.BranchID,
				DeliveryOrderNo: &doNo,
				Status:          domain.DeliveryBatchPicking,
				Notes:           "Surat Jalan otomatis untuk PO Modern Trade",
			}
			if err := tx.Create(&batch).Error; err != nil {
				return err
			}

			batchItem := domain.DeliveryItem{
				DeliveryBatchID: batch.ID,
				SalesOrderID:    &so.ID,
				Status:          domain.DeliveryItemPending,
				Sequence:        1,
			}
			if err := tx.Create(&batchItem).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

// ConfirmPOD mengkonfirmasi barang telah sampai di tangan pelanggan.
func (u *salesOrderUsecase) ConfirmPOD(req ConfirmPODRequest) error {
	so, err := u.soRepo.FindByID(req.SOID)
	if err != nil {
		return err
	}
	if so.Status != domain.SOStatusShipped {
		return errors.New("hanya pesanan yang sedang dikirim (SHIPPED) yang bisa dikonfirmasi terima")
	}

	now := time.Now()
	so.Status = domain.SOStatusDelivered
	so.ReceivedAt = &now
	so.ReceivedBy = req.ReceivedBy
	if req.PODImageURL != "" {
		so.PODImageURL = &req.PODImageURL
	}
	return u.soRepo.Update(so)
}

// ConvertToInvoice mengkonversi SO yang sudah DELIVERED menjadi SalesTransaction (Invoice).
func (u *salesOrderUsecase) ConvertToInvoice(id uuid.UUID, companyID uuid.UUID) (*domain.SalesTransaction, error) {
	return u.convertToInvoice(id, companyID, false, uuid.Nil)
}

// ConvertDirectToInvoice menerbitkan faktur SO motoris tanpa proses pengiriman.
func (u *salesOrderUsecase) ConvertDirectToInvoice(id uuid.UUID, adminID uuid.UUID) (*domain.SalesTransaction, error) {
	return u.convertToInvoice(id, uuid.Nil, true, adminID)
}

func (u *salesOrderUsecase) convertToInvoice(id uuid.UUID, companyID uuid.UUID, direct bool, adminID uuid.UUID) (*domain.SalesTransaction, error) {
	var resultInvoice *domain.SalesTransaction

	err := u.db.Transaction(func(tx *gorm.DB) error {
		soRepo := repository.NewSalesOrderRepository(tx)
		salesRepo := repository.NewSalesTransactionRepository(tx)

		// Kunci SO agar dua permintaan konversi tidak dapat menerbitkan dua faktur.
		var lockedSO domain.SalesOrder
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&lockedSO, "id = ?", id).Error; err != nil {
			return err
		}
		so, err := soRepo.FindByID(id)
		if err != nil {
			return err
		}
		now := time.Now()
		if direct {
			if (so.Status != domain.SOStatusDraft && so.Status != domain.SOStatusConfirmed && so.Status != domain.SOStatusApproved) || so.InvoiceID != nil || so.DeliveryBatchID != nil || so.IsMT {
				return errors.New("hanya SO DRAFT, CONFIRMED, atau APPROVED tanpa batch pengiriman dan faktur yang dapat langsung diterbitkan")
			}
			var deliveryCount int64
			if err := tx.Model(&domain.DeliveryItem{}).Where("sales_order_id = ?", id).Count(&deliveryCount).Error; err != nil {
				return err
			}
			if deliveryCount > 0 {
				return errors.New("SO sudah terkait pengiriman")
			}
			if len(so.Items) == 0 {
				return errors.New("SO tidak memiliki barang")
			}
			for i := range so.Items {
				item := &so.Items[i]
				if item.OrderedQuantity <= 0 || item.Quantity <= 0 {
					return errors.New("jumlah barang SO tidak valid")
				}
				if err := deductDirectSOProductStock(tx, so, item); err != nil {
					return err
				}
				item.ActualQuantity = item.OrderedQuantity
				if err := tx.Model(item).Updates(map[string]interface{}{
					"actual_quantity": item.ActualQuantity, "reserved_quantity": item.ReservedQuantity,
				}).Error; err != nil {
					return err
				}
			}
			for _, bonus := range so.BonusItems {
				if bonus.Quantity <= 0 {
					continue
				}
				if err := deductDirectSOBonusStock(tx, so, bonus); err != nil {
					return err
				}
			}
			if so.Status == domain.SOStatusDraft {
				so.AdminNotaConfirmedAt = &now
				so.AdminNotaConfirmedByID = &adminID
			}
		} else if so.Status != domain.SOStatusDelivered {
			return errors.New("hanya SO yang sudah diterima pelanggan (DELIVERED) yang bisa diterbitkan fakturnya")
		}
		if companyID != uuid.Nil && companyID != so.CompanyID {
			return errors.New("perusahaan faktur tidak sesuai dengan SO")
		}
		companyID = so.CompanyID

		var invoice *domain.SalesTransaction

		if so.InvoiceID != nil {
			// Re-use existing invoice
			existingTrx, err := salesRepo.FindByID(*so.InvoiceID)
			if err == nil {
				invoice = existingTrx
				// Update total amount and items if needed
				invoice.TotalAmount = so.TotalAmount
				invoice.UpdatedAt = now

				// Clear old items and replace with new ones
				tx.Where("sales_transaction_id = ?", invoice.ID).Delete(&domain.SalesItem{})
				var newItems []domain.SalesItem
				for _, soItem := range so.Items {
					if soItem.ActualQuantity > 0 {
						newItems = append(newItems, domain.SalesItem{
							SalesTransactionID: invoice.ID,
							ProductID:          soItem.ProductID,
							Quantity:           soItem.ActualQuantity,
							PriceAtTransaction: soItem.Price,
							Subtotal:           float64(soItem.ActualQuantity) * soItem.Price,
						})
					}
				}
				if len(newItems) > 0 {
					tx.Create(&newItems)
				}

				if err := tx.Omit(clause.Associations).Save(invoice).Error; err != nil {
					return err
				}
			}
		}

		if invoice == nil {
			// Generate nomor invoice baru
			invoiceCount, _ := salesRepo.CountByCompanyAndDate(companyID, now)
			invoiceNo := fmt.Sprintf("INV-%s-%04d", now.Format("20060102"), invoiceCount+1)

			// Buat SalesTransaction (Invoice) menggunakan total tagihan yang sudah didiskon
			paymentMethod := so.PaymentMethod
			if paymentMethod == "" || (direct && paymentMethod == "SALES_ORDER") {
				paymentMethod = "CASH"
			}
			invoice = &domain.SalesTransaction{
				CompanyID:            companyID,
				StoreID:              &so.StoreID,
				EmployeeID:           so.EmployeeID,
				VisitID:              so.VisitID,
				ReceiptNo:            invoiceNo,
				ReceiptImageURL:      so.PhysicalNotaImageURL,
				TotalAmount:          so.TotalAmount,
				DiscountAmount:       so.DiscountAmount,
				StoreCategory:        so.StoreCategory,
				TransactionDate:      now,
				PeriodMonth:          int(now.Month()),
				PeriodYear:           now.Year(),
				Status:               "PENDING",
				PaymentStatus:        string(domain.PaymentStatusUnpaid),
				PaymentMethod:        paymentMethod,
				PaymentDueDate:       so.PaymentDueDate,
				CompanyBankAccountID: so.CompanyBankAccountID,
				Notes:                &so.Notes,
			}
			if direct {
				invoice.SalesChannel = domain.SalesChannelMotoris
			}

			for _, soItem := range so.Items {
				if soItem.ActualQuantity > 0 {
					quantity := soItem.ActualQuantity
					if direct {
						quantity = soItem.Quantity // stok bronjong dan faktur memakai satuan terkecil
					}
					invoice.Items = append(invoice.Items, domain.SalesItem{
						ProductID:          soItem.ProductID,
						Quantity:           quantity,
						PriceAtTransaction: soItem.Price,
						DiscountPercent:    soItem.DiscountPercent,
						Subtotal:           float64(quantity) * soItem.Price * (1 - soItem.DiscountPercent/100),
					})
				}
			}
			if direct {
				for _, bonus := range so.BonusItems {
					invoice.BonusItems = append(invoice.BonusItems, domain.SalesTransactionBonusItem{
						BonusInventoryID: bonus.BonusInventoryID,
						Quantity:         bonus.Quantity,
						Notes:            bonus.Notes,
					})
				}
			}

			if err := salesRepo.Create(invoice); err != nil {
				return err
			}
		}

		paymentMethod := so.PaymentMethod
		if paymentMethod == "" || (direct && paymentMethod == "SALES_ORDER") {
			paymentMethod = "CASH"
		}

		// Otomatis buat setoran kasir PENDING jika ini nota fisik dan pembayarannya CASH atau SALES_ORDER
		if so.IsPhysicalNota && (paymentMethod == "CASH" || paymentMethod == "SALES_ORDER" || paymentMethod == "TUNAI") && so.TotalAmount > 0 {
			payment := &domain.SalesPayment{
				SalesTransactionID: &invoice.ID,
				EmployeeID:         so.EmployeeID,
				Amount:             so.TotalAmount,
				PaymentMethod:      "CASH",
				PaymentStatus:      "PENDING",
				PaymentDate:        now,
				Notes:              "Setoran nota fisik",
			}
			if err := tx.Create(payment).Error; err != nil {
				return err
			}
		}

		// Update SO ke CONVERTED
		invoiceID := invoice.ID
		so.Status = domain.SOStatusConverted
		so.InvoiceID = &invoiceID
		so.ConvertedAt = &now
		if err := soRepo.Update(so); err != nil {
			return err
		}

		resultInvoice = invoice
		return nil
	})

	return resultInvoice, err
}

// Ambil stok bronjong terlebih dahulu; sisa kebutuhan dipenuhi dari gudang cabang.
// Seluruh mutasi diproses dalam transaksi konversi yang sama.
func deductDirectSOProductStock(tx *gorm.DB, so *domain.SalesOrder, item *domain.SalesOrderItem) error {
	// Reservasi milik SO ini tidak lagi diperlukan ketika barang langsung difakturkan.
	for _, allocation := range item.Batches {
		if allocation.Quantity <= 0 {
			continue
		}
		if err := tx.Model(&domain.WarehouseStock{}).
			Where("branch_id = ? AND product_id = ? AND stock_type = 'GOOD' AND batch_no = ?", so.BranchID, item.ProductID, allocation.BatchNo).
			UpdateColumn("reserved_quantity", gorm.Expr("GREATEST(reserved_quantity - ?, 0)", allocation.Quantity)).Error; err != nil {
			return err
		}
	}
	if len(item.Batches) > 0 {
		if err := tx.Where("sales_order_item_id = ?", item.ID).Delete(&domain.SalesOrderItemBatch{}).Error; err != nil {
			return err
		}
	}
	item.ReservedQuantity = 0
	remaining := item.Quantity
	var salesStock domain.SalesStock
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("employee_id = ? AND product_id = ?", so.EmployeeID, item.ProductID).First(&salesStock).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if err == nil && salesStock.Quantity > 0 {
		used := salesStock.Quantity
		if used > remaining {
			used = remaining
		}
		if err := tx.Model(&salesStock).UpdateColumn("quantity", gorm.Expr("quantity - ?", used)).Error; err != nil {
			return err
		}
		remaining -= used
	}
	if remaining == 0 {
		return nil
	}

	var batches []domain.WarehouseStock
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("branch_id = ? AND product_id = ? AND stock_type = 'GOOD' AND batch_no != 'BACKORDER' AND quantity > reserved_quantity", so.BranchID, item.ProductID).
		Order("expiry_date ASC NULLS LAST, created_at ASC").Find(&batches).Error; err != nil {
		return err
	}
	for _, batch := range batches {
		available := batch.Quantity - batch.ReservedQuantity
		if available <= 0 {
			continue
		}
		used := available
		if used > remaining {
			used = remaining
		}
		if err := tx.Model(&batch).UpdateColumn("quantity", gorm.Expr("quantity - ?", used)).Error; err != nil {
			return err
		}
		source := "SO LANGSUNG: " + so.SONumber
		if len(source) > 50 {
			source = source[:50]
		}
		if err := tx.Create(&domain.WarehouseLog{
			BranchID: so.BranchID, ProductID: item.ProductID, Type: "OUT",
			Source: source, BatchNo: batch.BatchNo, Quantity: used,
		}).Error; err != nil {
			return err
		}
		remaining -= used
		if remaining == 0 {
			return nil
		}
	}
	name := item.ProductID.String()
	if item.Product != nil {
		name = item.Product.Name
	}
	return fmt.Errorf("stok bronjong dan gudang cabang tidak cukup untuk %s (kurang %d pcs)", name, remaining)
}

func deductDirectSOBonusStock(tx *gorm.DB, so *domain.SalesOrder, bonus domain.SalesOrderBonusItem) error {
	remaining := bonus.Quantity
	var salesStock domain.SalesBonusStock
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("employee_id = ? AND bonus_inventory_id = ?", so.EmployeeID, bonus.BonusInventoryID).First(&salesStock).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if err == nil && salesStock.Quantity > 0 {
		used := salesStock.Quantity
		if used > remaining {
			used = remaining
		}
		if err := tx.Model(&salesStock).UpdateColumn("quantity", gorm.Expr("quantity - ?", used)).Error; err != nil {
			return err
		}
		remaining -= used
	}
	if remaining == 0 {
		return nil
	}
	var inventory domain.BonusInventory
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ? AND branch_id = ?", bonus.BonusInventoryID, so.BranchID).First(&inventory).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("stok bonus cabang tidak ditemukan untuk %s", bonus.BonusInventoryID)
		}
		return err
	}
	if inventory.Quantity < remaining {
		return fmt.Errorf("stok bonus bronjong dan gudang cabang tidak cukup untuk %s (kurang %d)", inventory.Name, remaining-inventory.Quantity)
	}
	before := inventory.Quantity
	if err := tx.Model(&inventory).UpdateColumn("quantity", gorm.Expr("quantity - ?", remaining)).Error; err != nil {
		return err
	}
	return tx.Create(&domain.BonusInventoryLog{
		BonusInventoryID: bonus.BonusInventoryID,
		Type:             domain.BonusInventoryOut, Quantity: remaining,
		QuantityBefore: before, QuantityAfter: before - remaining,
		Reference: so.SONumber, Notes: "Faktur langsung tanpa pengiriman",
	}).Error
}

// DeleteSO menghapus SO permanen. Hanya SO dengan status DRAFT yang bisa dihapus.
func (u *salesOrderUsecase) DeleteSO(id uuid.UUID) error {
	so, err := u.soRepo.FindByID(id)
	if err != nil {
		return err
	}
	if so.Status != domain.SOStatusDraft {
		return errors.New("hanya SO dengan status DRAFT yang bisa dihapus")
	}
	return u.soRepo.Delete(id)
}

// SoftDeleteSO menghapus SO dengan mekanisme soft delete
func (u *salesOrderUsecase) SoftDeleteSO(id uuid.UUID, deletedByID uuid.UUID, reason string) error {
	so, err := u.soRepo.FindByID(id)
	if err != nil {
		return err
	}
	if so == nil {
		return errors.New("pesanan tidak ditemukan")
	}

	// Check if this SO is linked to a delivery batch that has been confirmed ready
	var count int64
	dbErr := u.db.Table("delivery_items").
		Joins("JOIN delivery_batches ON delivery_batches.id = delivery_items.delivery_batch_id").
		Where("delivery_items.sales_order_id = ? AND delivery_batches.status IN ?", id,
			[]string{"PICKING", "ON_DELIVERY", "COMPLETED"}).
		Count(&count).Error
	if dbErr == nil && count > 0 {
		return errors.New("tidak bisa menghapus pesanan yang ada di dalam surat jalan yang sudah dikonfirmasi siap, dikirim, atau selesai")
	}

	return u.soRepo.SoftDeleteWithReason(id, deletedByID, reason)
}

// RestoreSO memulihkan SO yang sebelumnya dihapus
func (u *salesOrderUsecase) DeleteDraftSOBySales(id, employeeID uuid.UUID) error {
	return u.db.Transaction(func(tx *gorm.DB) error {
		var so domain.SalesOrder
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&so, "id = ?", id).Error; err != nil {
			return errors.New("pesanan tidak ditemukan")
		}
		if so.EmployeeID != employeeID {
			return errors.New("pesanan bukan milik sales ini")
		}
		if so.Status != domain.SOStatusDraft || so.AdminNotaConfirmedAt != nil || so.InvoiceID != nil {
			return errors.New("hanya SO DRAFT yang belum divalidasi dapat dihapus")
		}
		var linked int64
		if err := tx.Model(&domain.DeliveryItem{}).Where("sales_order_id = ?", id).Count(&linked).Error; err != nil {
			return err
		}
		if linked > 0 {
			return errors.New("pesanan yang sudah masuk pengiriman tidak dapat dihapus")
		}
		return repository.NewSalesOrderRepository(tx).SoftDeleteWithReason(id, employeeID, "Dihapus oleh sales sebelum validasi Admin Nota")
	})
}

// RestoreSO memulihkan SO yang sebelumnya dihapus

func (u *salesOrderUsecase) RestoreSO(id uuid.UUID) error {
	so, err := u.soRepo.FindByIDUnscoped(id)
	if err != nil {
		return err
	}
	if so == nil {
		return errors.New("pesanan tidak ditemukan")
	}

	return u.soRepo.Restore(id)
}

func (u *salesOrderUsecase) UploadNotaFisik(req UploadNotaFisikRequest) (*domain.SalesOrder, error) {
	now := time.Now()

	// 1. Cek Branch & Company dari Karyawan (Prioritas Utama, Behavior Lama)
	if req.BranchID == uuid.Nil || req.CompanyID == uuid.Nil {
		var emp domain.Employee
		if err := u.db.Where("id = ?", req.EmployeeID).First(&emp).Error; err == nil {
			if req.BranchID == uuid.Nil && emp.BranchID != nil {
				req.BranchID = *emp.BranchID
			}
			if req.CompanyID == uuid.Nil && emp.CompanyID != nil {
				req.CompanyID = *emp.CompanyID
			}
		}
	}

	// 2. Fallback BranchID & CompanyID to Store's assigned employee if still empty
	if req.BranchID == uuid.Nil || req.CompanyID == uuid.Nil {
		var store domain.Store
		if err := u.db.Preload("AssignedEmployee").Where("id = ?", req.StoreID).First(&store).Error; err == nil {
			if store.AssignedEmployee != nil {
				if req.BranchID == uuid.Nil && store.AssignedEmployee.BranchID != nil {
					req.BranchID = *store.AssignedEmployee.BranchID
				}
				if req.CompanyID == uuid.Nil && store.AssignedEmployee.CompanyID != nil {
					req.CompanyID = *store.AssignedEmployee.CompanyID
				}
			}
		}
	}

	// Resolve and Complete Visit
	var visit *domain.SalesVisit
	if req.VisitID != nil {
		visit = &domain.SalesVisit{}
		if err := u.db.First(visit, "id = ?", *req.VisitID).Error; err != nil {
			visit = nil
		}
	} else {
		start := now.Add(-24 * time.Hour)
		end := now.Add(24 * time.Hour)
		visit = &domain.SalesVisit{}
		err := u.db.Where("employee_id = ? AND store_id = ? AND check_in_time >= ? AND check_in_time <= ? AND status != 'SKIPPED' AND type != 'SKIP'", req.EmployeeID, req.StoreID, start, end).Order("check_in_time DESC").First(visit).Error
		if err != nil {
			visit = nil
		}
	}
	if visit != nil {
		req.VisitID = &visit.ID
		checkOutTime := time.Now()
		visit.CheckOutTime = &checkOutTime
		visit.Status = "COMPLETED"
		if visit.Notes == "Deleted from schedule" {
			visit.Notes = "Kunjungan sales & order"
		}
		if visit.Type == "SKIP" {
			visit.Type = "CHECKIN"
		}
		if visit.CheckInTime.After(checkOutTime) {
			visit.CheckInTime = checkOutTime.Add(-15 * time.Minute)
		}
		u.db.Save(visit)
	}

	var company domain.Company
	prefix := ""
	if err := u.db.First(&company, "id = ?", req.CompanyID).Error; err == nil {
		if company.Code != "" {
			prefix = company.Code
		} else {
			words := strings.Fields(company.Name)
			if len(words) > 0 {
				name := words[len(words)-1]
				if len(name) > 5 {
					prefix = strings.ToUpper(name[:5])
				} else {
					prefix = strings.ToUpper(name)
				}
			}
		}
	}
	var store domain.Store
	var storeCategory = "TOKO_LAMA"
	if err := u.db.First(&store, "id = ?", req.StoreID).Error; err == nil {
		storeCategory = store.DetermineCategory(now)
	}

	// Determine effective salesman for omzet / KPI:
	// If the nota fisik is uploaded by supervisor / management on a store with an assigned salesman (PIC),
	// the order's EmployeeID is set to the store's assigned salesman so omzet & KPI credit goes to the PIC.
	effectiveSalesmanID := req.EmployeeID
	var creatorEmp domain.Employee
	if err := u.db.Preload("JobPosition").Where("id = ?", req.EmployeeID).First(&creatorEmp).Error; err == nil {
		if creatorEmp.JobPosition != nil {
			title := strings.ToLower(creatorEmp.JobPosition.Title)
			isManagementOrSpv := strings.Contains(title, "direktur") || strings.Contains(title, "director") ||
				strings.Contains(title, "manager") || strings.Contains(title, "owner") ||
				strings.Contains(title, "superadmin") || strings.Contains(title, "spv") ||
				strings.Contains(title, "supervisor")
			if isManagementOrSpv {
				var currentStore domain.Store
				if err := u.db.Select("id, assigned_employee_id").Where("id = ?", req.StoreID).First(&currentStore).Error; err == nil {
					if currentStore.AssignedEmployeeID != nil && *currentStore.AssignedEmployeeID != uuid.Nil {
						effectiveSalesmanID = *currentStore.AssignedEmployeeID
						fmt.Printf("[NOTA-AUDIT] Nota fisik uploaded by SPV/Management %s credited to Store PIC %s\n", req.EmployeeID, effectiveSalesmanID)
					}
				}
			}
		}
	}

	so := &domain.SalesOrder{
		BranchID:             req.BranchID,
		CompanyID:            req.CompanyID,
		EmployeeID:           effectiveSalesmanID,
		StoreID:              req.StoreID,
		StoreCategory:        storeCategory,
		Status:               domain.SOStatusDraft, // Start as Draft
		Notes:                req.Notes,
		OrderDate:            now,
		VisitID:              req.VisitID,
		Latitude:             req.Latitude,
		Longitude:            req.Longitude,
		PaymentMethod:        "SALES_ORDER", // default
		IsPhysicalNota:       true,
		PhysicalNotaImageURL: req.PhysicalNotaImageURL,
		TotalAmount:          0, // Items will be added later by Admin
	}

	// Buat SO kosong (tanpa items)
	if err := u.soRepo.CreateWithNextSONumber(so, prefix); err != nil {
		return nil, err
	}

	// Publish SSE event ke admin nota agar dinotifikasi secara real-time
	if u.sseHub != nil {
		go func() {
			var salesmanName string
			var employee domain.Employee
			if err := u.db.First(&employee, "id = ?", req.EmployeeID).Error; err == nil {
				salesmanName = employee.FirstName + " " + employee.LastName
			}

			var targetBranchIDs []string
			if so.BranchID != uuid.Nil {
				targetBranchIDs = []string{so.BranchID.String()}
			}

			event := sse.NotaFisikEvent{
				TransactionID:   so.ID.String(),
				ReceiptNo:       so.SONumber,
				Salesman:        salesmanName,
				ImageURL:        req.PhysicalNotaImageURL,
				CompanyID:       so.CompanyID.String(),
				CreatedAt:       so.CreatedAt.Format("2006-01-02 15:04:05"),
				StoreName:       store.Name,
				TargetBranchIDs: targetBranchIDs,
				TargetRoles: []string{
					string(domain.RoleSuperAdmin),
					string(domain.RoleCashier),
					string(domain.RoleCashierPusat),
					string(domain.RoleAdminKasir),
					string(domain.RoleManager),
					string(domain.RoleAdminNota),
				},
			}
			if pubErr := u.sseHub.PublishNotaFisik(context.Background(), event); pubErr != nil {
				fmt.Printf("[SSE] Failed to publish nota_fisik event: %v\n", pubErr)
			}
		}()
	}

	return so, nil
}

func (u *salesOrderUsecase) GetSOByEmployee(employeeID uuid.UUID) ([]domain.SalesOrder, error) {
	return u.soRepo.FindByEmployee(employeeID)
}

func (u *salesOrderUsecase) GetSOByBranch(branchIDs []uuid.UUID, status string, salesType string, paymentStatus string, startDate string, endDate string, employeeID uuid.UUID, companyID uuid.UUID, search string, page int, limit int) ([]domain.SalesOrder, int64, *repository.SOSummary, error) {
	return u.soRepo.FindByBranch(branchIDs, status, salesType, paymentStatus, startDate, endDate, employeeID, companyID, search, page, limit)
}

func (u *salesOrderUsecase) GetSOByStore(storeID uuid.UUID, page int, limit int) ([]domain.SalesOrder, int64, error) {
	return u.soRepo.FindByStore(storeID, page, limit)
}

func (u *salesOrderUsecase) GetSOByID(id uuid.UUID) (*domain.SalesOrder, error) {
	return u.soRepo.FindByID(id)
}

func (u *salesOrderUsecase) OverrideBackorder(id uuid.UUID) error {
	so, err := u.soRepo.FindByID(id)
	if err != nil {
		return err
	}
	// Ubah ke CONFIRMED agar bisa masuk antrean pengiriman / gudang
	so.Status = domain.SOStatusConfirmed
	return u.soRepo.Update(so)
}

func (u *salesOrderUsecase) RecheckStock(id uuid.UUID) (*domain.SalesOrder, error) {
	var updatedSO *domain.SalesOrder
	err := u.db.Transaction(func(tx *gorm.DB) error {
		soRepo := repository.NewSalesOrderRepository(tx)

		so, err := soRepo.FindByID(id)
		if err != nil {
			return err
		}
		if so.Status == domain.SOStatusWaitingStock {
			so.Status = domain.SOStatusConfirmed
			if err := soRepo.Update(so); err != nil {
				return err
			}
		}

		reloaded, err := soRepo.FindByID(id)
		if err != nil {
			return err
		}
		updatedSO = reloaded
		return nil
	})

	return updatedSO, err
}

// =============================================================================
// ALUR BARU: Admin Nota, Driver, & Pembayaran
// =============================================================================

// AdminConfirmSO digunakan Admin Nota untuk menyetujui nota pesanan dari salesman.
// Status langsung menjadi CONFIRMED (siap masuk antrean pengiriman Logistics Dashboard).
func (u *salesOrderUsecase) AdminConfirmSO(id uuid.UUID, adminID uuid.UUID) error {
	var confirmedSO *domain.SalesOrder
	err := u.db.Transaction(func(tx *gorm.DB) error {
		soRepo := repository.NewSalesOrderRepository(tx)

		so, err := soRepo.FindByID(id)
		if err != nil {
			return err
		}
		if so.Status != domain.SOStatusDraft {
			return errors.New("hanya pesanan dengan status DRAFT yang bisa diverifikasi")
		}

		now := time.Now()
		so.Status = domain.SOStatusConfirmed
		so.AdminNotaConfirmedAt = &now
		so.AdminNotaConfirmedByID = &adminID
		confirmedSO = so
		if err := soRepo.Update(so); err != nil {
			return err
		}

		// Bila pembayaran kasir sudah selesai lebih dulu, finalisasi status
		// invoice sekarang juga. Sebelumnya status invoice tetap PENDING.
		if so.InvoiceID != nil {
			var invoice domain.SalesTransaction
			if err := tx.First(&invoice, "id = ?", *so.InvoiceID).Error; err == nil && invoice.PaymentStatus == "PAID" {
				if err := tx.Model(&invoice).Update("status", "APPROVED").Error; err != nil {
					return err
				}
			}
		}
		return nil
	})

	if err == nil && confirmedSO != nil && u.fcmClient != nil {
		var employee domain.Employee
		if errDb := u.db.First(&employee, "id = ?", confirmedSO.EmployeeID).Error; errDb == nil {
			title := "Pesanan SO Disetujui"
			body := fmt.Sprintf("Pesanan SO %s Anda telah disetujui oleh Admin.", confirmedSO.SONumber)
			go func() {
				_ = u.fcmClient.SendPushNotification(employee.UserID, title, body)
			}()
		}
	}

	return err
}

// AdminRejectSO digunakan Admin Nota untuk menolak nota pesanan dari salesman.
func (u *salesOrderUsecase) AdminRejectSO(id uuid.UUID, adminID uuid.UUID, notes string) error {
	so, err := u.soRepo.FindByID(id)
	if err != nil {
		return err
	}
	if so.Status != domain.SOStatusDraft {
		return errors.New("hanya pesanan dengan status DRAFT yang bisa ditolak oleh Admin Nota")
	}

	now := time.Now()
	so.Status = domain.SOStatusRejected
	so.RejectedAt = &now
	so.RejectedByID = &adminID
	so.AdminNotaRejectNotes = notes
	err = u.soRepo.Update(so)

	if err == nil && u.fcmClient != nil {
		var employee domain.Employee
		if errDb := u.db.First(&employee, "id = ?", so.EmployeeID).Error; errDb == nil {
			title := "Pesanan Ditolak"
			body := fmt.Sprintf("Pesanan %s ditolak oleh Admin Nota. Catatan: %s", so.SONumber, notes)
			go func() {
				_ = u.fcmClient.SendPushNotification(employee.UserID, title, body)
			}()
		}
	}

	return err
}
func (u *salesOrderUsecase) ApproveTaxDocuments(id uuid.UUID, adminID uuid.UUID, billingNote, taxInvoice, deliveryNote string, destinationBranchID uuid.UUID) error {
	err := u.db.Transaction(func(tx *gorm.DB) error {
		soRepo := repository.NewSalesOrderRepository(tx)

		so, err := soRepo.FindByID(id)
		if err != nil {
			return err
		}
		if so.Status != domain.SOStatusDraft && so.Status != "WAITING_TAX_DOCS" && so.Status != domain.SOStatusConfirmed {
			return errors.New("hanya pesanan DRAFT, WAITING_TAX_DOCS, atau CONFIRMED yang dapat diproses dokumen pajaknya")
		}

		now := time.Now()
		so.BillingNoteURL = billingNote
		so.TaxInvoiceURL = taxInvoice
		so.DeliveryNoteURL = deliveryNote
		so.DestinationBranchID = &destinationBranchID
		so.BranchID = destinationBranchID
		so.AdminNotaConfirmedAt = &now
		so.AdminNotaConfirmedByID = &adminID
		so.Status = domain.SOStatusConfirmed

		if err := soRepo.Update(so); err != nil {
			return err
		}

		return nil
	})

	if err == nil && u.fcmClient != nil {
		// Notifikasi FCM untuk Sales MT dan Admin Gudang
		var employee domain.Employee
		soRepo := repository.NewSalesOrderRepository(u.db)
		if so, errSo := soRepo.FindByID(id); errSo == nil {
			go func() {
				// 1. Notifikasi ke Sales MT
				if errDb := u.db.First(&employee, "id = ?", so.EmployeeID).Error; errDb == nil {
					_ = u.fcmClient.SendPushNotification(employee.UserID, "PO MT Disetujui", fmt.Sprintf("Dokumen Pajak PO MT (%s) Anda telah lengkap. Silakan cetak ketiga dokumen & kirim barang.", so.SONumber))
				}

				// 2. Notifikasi ke Admin Gudang (Cabang Pengirim)
				var gudangUsers []domain.User
				u.db.Joins("JOIN employees ON employees.user_id = users.id").
					Where("users.role = ? AND employees.branch_id = ?", domain.RoleAdminGudang, destinationBranchID).
					Find(&gudangUsers)

				for _, gu := range gudangUsers {
					_ = u.fcmClient.SendPushNotification(gu.ID, "Tugas Pengeluaran Stok PO MT", fmt.Sprintf("PO MT (%s) telah disetujui Pajak. Segera siapkan dan catat pengeluaran fisik barang dari gudang.", so.SONumber))
				}

				// 3. SSE Notification (Logistics Dashboard)
				if u.sseHub != nil {
					storeName := "Toko"
					if so.Store != nil {
						storeName = so.Store.Name
					}
					event := sse.NotificationEvent{
						ID:        so.ID.String(),
						Type:      "sales_order",
						Title:     "PO MT Siap Diproses",
						Message:   fmt.Sprintf("PO MT %s (%s) siap diproses gudang", so.SONumber, storeName),
						CompanyID: so.CompanyID.String(),
						CreatedAt: time.Now().Format("2006-01-02 15:04:05"),
						Data: map[string]interface{}{
							"so_id":     so.ID.String(),
							"so_number": so.SONumber,
							"status":    so.Status,
						},
					}
					_ = u.sseHub.PublishNotification(context.Background(), event)
				}
			}()
		}
	}

	return err
}

func (u *salesOrderUsecase) ApproveMTTaxDocuments(mtPoID uuid.UUID, adminID uuid.UUID, billingNote, taxInvoice, deliveryNote string, destinationBranchID uuid.UUID) error {
	var mtPO domain.MTPurchaseOrder
	if err := u.db.First(&mtPO, "id = ?", mtPoID).Error; err != nil {
		return errors.New("MT Purchase Order tidak ditemukan")
	}

	var so domain.SalesOrder
	if err := u.db.Preload("Items").First(&so, "so_number = ?", mtPO.PONumber).Error; err != nil {
		return errors.New("Sales Order terkait tidak ditemukan")
	}

	err := u.ApproveTaxDocuments(so.ID, adminID, billingNote, taxInvoice, deliveryNote, destinationBranchID)
	if err != nil {
		return err
	}

	// Update MT PO tax documents and status to prevent double approval
	u.db.Model(&mtPO).Updates(map[string]interface{}{
		"status":            "APPROVED",
		"invoice_photo_url": taxInvoice,
		"receipt_photo_url": billingNote,
		"delivery_note_url": deliveryNote,
		"dc_destination":    destinationBranchID.String(),
	})

	// Generate MTDeliveryOrder for the warehouse to process
	var admin domain.Employee
	u.db.First(&admin, "id = ?", adminID)
	createdBy := admin.FirstName + " " + admin.LastName

	mtDO := domain.MTDeliveryOrder{
		CompanyID:      so.CompanyID,
		DONumber:       fmt.Sprintf("DO-MT-%d", time.Now().Unix()),
		KeyAccountID:   mtPO.KeyAccountID,
		PONumber:       mtPO.PONumber,
		SourceType:     domain.MTDeliverySourceBranch,
		SourceBranchID: &destinationBranchID,
		Status:         domain.MTDeliveryAssigned,
		DCDestination:  destinationBranchID.String(),
		CreatedBy:      createdBy,
		SalesOrderID:   &so.ID,
	}

	for _, item := range so.Items {
		mtDO.Items = append(mtDO.Items, domain.MTDeliveryOrderItem{
			ProductID:    item.ProductID,
			RequestedQty: item.Quantity,
			Unit:         "Kardus",
		})
	}

	return u.db.Create(&mtDO).Error
}

func (u *salesOrderUsecase) ConfirmDelivery(req ConfirmDeliveryRequest) (*domain.SalesTransaction, error) {
	var resultInvoice *domain.SalesTransaction
	isTokoTutup := (req.ReceivedBy == "TOKO TUTUP")
	var so *domain.SalesOrder

	err := u.db.Transaction(func(tx *gorm.DB) error {
		soRepo := repository.NewSalesOrderRepository(tx)

		var err error
		so, err = soRepo.FindByID(req.SOID)
		if err != nil {
			return err
		}
		if so.Status != domain.SOStatusInDelivery && so.Status != domain.SOStatusConfirmed && so.Status != domain.SOStatusPaid {
			return fmt.Errorf("pesanan tidak dalam status pengiriman aktif (status saat ini: %s)", so.Status)
		}

		// 1. Bangun map retur: productID → {qty, reason, returnType}
		type returnInfo struct {
			qty        int
			reason     string
			returnType string
		}
		returnedMap := make(map[uuid.UUID]returnInfo)
		for _, ri := range req.ReturnedItems {
			rType := ri.ReturnType
			if rType == "" {
				rType = "POTONG_NOTA"
			}
			returnedMap[ri.ProductID] = returnInfo{qty: ri.ReturnedQuantity, reason: ri.ReturnReason, returnType: rType}
		}

		now := time.Now()
		newTotalAmount := so.TotalAmount

		var adjustmentNotes []string
		for i := range so.Items {
			item := &so.Items[i]
			ri := returnedMap[item.ProductID]

			if isTokoTutup {
				// Biarkan ActualQuantity = 0 karena tidak ada yang dikirim,
				// TAPI jangan nolkan Subtotal agar data SO tidak rusak (tetap sesuai pesanan awal).
				item.ActualQuantity = 0

				prodName := "Produk"
				if item.Product != nil {
					prodName = item.Product.Name
				} else {
					var prod domain.Product
					if errDb := tx.Select("name").Where("id = ?", item.ProductID).First(&prod).Error; errDb == nil {
						prodName = prod.Name
					}
				}
				adjustmentNotes = append(adjustmentNotes, fmt.Sprintf("- %s: Gagal kirim karena Toko Tutup", prodName))
			} else {
				retQty := ri.qty
				rType := ri.returnType

				if retQty > 0 && rType == "POTONG_NOTA" {
					// Hitung qty aktual yang benar-benar sampai
					item.ActualQuantity = item.OrderedQuantity - retQty
					if item.ActualQuantity < 0 {
						item.ActualQuantity = 0
					}

					// Kurangi total tagihan sebesar nilai retur saja (diskon tetap aman)
					returnedValue := float64(retQty*item.PiecesPerUnit) * item.Price
					newTotalAmount -= returnedValue
					if newTotalAmount < 0 {
						newTotalAmount = 0
					}
				} else {
					// Jika Tukar Langsung atau Tukar Susulan, tetap ditagih penuh
					item.ActualQuantity = item.OrderedQuantity
				}

				item.Subtotal = float64(item.ActualQuantity*item.PiecesPerUnit) * item.Price

				if retQty > 0 {
					prodName := "Produk"
					if item.Product != nil {
						prodName = item.Product.Name
					} else {
						var prod domain.Product
						if errDb := tx.Select("name").Where("id = ?", item.ProductID).First(&prod).Error; errDb == nil {
							prodName = prod.Name
						}
					}
					adjustmentNotes = append(adjustmentNotes, fmt.Sprintf("- %s: dipesan %d %s, terkirim %d %s, tidak terkirim %d %s (Alasan: %s, Tipe Retur: %s)", prodName, item.OrderedQuantity, item.Unit, item.ActualQuantity, item.Unit, retQty, item.Unit, ri.reason, rType))
				}
			}

			// Save the sales order item explicitly to persist actual_quantity and subtotal updates without touching associations.
			if err := tx.Model(&domain.SalesOrderItem{}).Where("id = ?", item.ID).Updates(map[string]interface{}{
				"actual_quantity": item.ActualQuantity,
				"subtotal":        item.Subtotal,
			}).Error; err != nil {
				return fmt.Errorf("gagal memperbarui item pesanan: %w", err)
			}

			// Deduct from Driver's SalesStock (Gudang Berjalan)
			if req.EmployeeID != uuid.Nil {
				salesTransferRepo := repository.NewSalesTransferRepository(tx)
				stock, stockErr := salesTransferRepo.GetSalesStock(req.EmployeeID, item.ProductID)
				if stockErr == nil && stock != nil {
					deductQty := item.ActualQuantity
					if ri.qty > 0 {
						if ri.returnType == "TUKAR_SUSULAN" {
							deductQty = item.OrderedQuantity - ri.qty
						} else if ri.returnType == "TUKAR_PRODUK" {
							deductQty = item.OrderedQuantity - ri.qty
						} else if ri.returnType == "TUKAR_LANGSUNG" {
							deductQty = item.OrderedQuantity
						}
					}
					actualQtyPcs := deductQty * item.PiecesPerUnit
					stock.Quantity -= actualQtyPcs
					stock.UpdatedAt = time.Now()
					if err := salesTransferRepo.UpdateSalesStock(stock); err != nil {
						return fmt.Errorf("gagal memperbarui stok driver: %w", err)
					}
				}
			}
		}

		so.TotalAmount = newTotalAmount

		// 2. Update SO Status
		if isTokoTutup {
			so.Status = domain.SOStatusCancelled
		} else {
			so.Status = domain.SOStatusDelivered
		}
		so.ReceivedAt = &now
		so.ReceivedBy = req.ReceivedBy
		if req.PODImageURL != "" {
			so.PODImageURL = &req.PODImageURL
		}
		if req.SignatureImageURL != "" {
			so.SignatureImageURL = &req.SignatureImageURL
		}
		if req.Notes != "" {
			if so.Notes != "" {
				so.Notes += "\n" + req.Notes
			} else {
				so.Notes = req.Notes
			}
		}

		// Hitung total unit retur untuk SO ini
		totalReturnQty := 0
		for _, ri := range req.ReturnedItems {
			totalReturnQty += ri.ReturnedQuantity
		}

		var invoiceID *uuid.UUID
		// 3. Generate Invoice (SalesTransaction)
		salesRepo := repository.NewSalesTransactionRepository(tx)

		paymentMethod := so.PaymentMethod
		if paymentMethod == "" {
			paymentMethod = "CASH"
		}

		var invoiceNotes string
		if len(adjustmentNotes) > 0 {
			invoiceNotes = "[PENYESUAIAN PENGIRIMAN SEBAGIAN]\n" + strings.Join(adjustmentNotes, "\n")
			if so.Notes != "" {
				invoiceNotes = invoiceNotes + "\n\nCatatan Awal:\n" + so.Notes
			}
		} else {
			invoiceNotes = so.Notes
		}

		var invoice *domain.SalesTransaction

		if so.InvoiceID != nil {
			invoiceID = so.InvoiceID
			var existingInvoice domain.SalesTransaction
			if err := tx.Preload("Items").First(&existingInvoice, "id = ?", *invoiceID).Error; err == nil {
				invoice = &existingInvoice
				invoice.TotalAmount = so.TotalAmount
				invoice.ReceiptImageURL = req.PODImageURL
				invoice.SignatureImageURL = req.SignatureImageURL
				invoice.Notes = &invoiceNotes

				if invoice.TotalAmount <= 0 && !isTokoTutup {
					invoice.Status = "VERIFIED"
					invoice.PaymentStatus = string(domain.PaymentStatusPaid)
				}

				// Delete old items and recreate
				if err := tx.Where("sales_transaction_id = ?", invoice.ID).Delete(&domain.SalesItem{}).Error; err != nil {
					return fmt.Errorf("gagal menghapus rincian faktur lama: %w", err)
				}
				var newItems []domain.SalesItem
				for _, soItem := range so.Items {
					if soItem.ActualQuantity > 0 {
						newItems = append(newItems, domain.SalesItem{
							SalesTransactionID: invoice.ID,
							ProductID:          soItem.ProductID,
							Quantity:           soItem.ActualQuantity,
							PriceAtTransaction: soItem.Price,
							Subtotal:           float64(soItem.ActualQuantity) * soItem.Price,
						})
					}
				}
				invoice.Items = newItems
				if err := tx.Omit(clause.Associations).Save(invoice).Error; err != nil {
					return fmt.Errorf("gagal memperbarui faktur penjualan: %w", err)
				}
				if len(newItems) > 0 {
					if err := tx.Create(&newItems).Error; err != nil {
						return fmt.Errorf("gagal membuat rincian faktur baru: %w", err)
					}
				}
			}
		} else {
			invoiceCount, _ := salesRepo.CountByCompanyAndDate(so.CompanyID, now)
			invoiceNo := fmt.Sprintf("INV-%s-%04d", now.Format("20060102"), invoiceCount+1)

			// Ensure unique receipt_no
			for attempt := 1; attempt <= 20; attempt++ {
				var count int64
				tx.Model(&domain.SalesTransaction{}).Where("receipt_no = ?", invoiceNo).Count(&count)
				if count == 0 {
					break
				}
				invoiceNo = fmt.Sprintf("INV-%s-%04d", now.Format("20060102"), invoiceCount+1+int64(attempt))
			}

			initialTrxStatus := "APPROVED"
			initialTrxPaymentStatus := string(domain.PaymentStatusUnpaid)
			if so.TotalAmount <= 0 && !isTokoTutup {
				initialTrxStatus = "VERIFIED"
				initialTrxPaymentStatus = string(domain.PaymentStatusPaid)
			}

			invoice = &domain.SalesTransaction{
				CompanyID:         so.CompanyID,
				StoreID:           &so.StoreID,
				EmployeeID:        so.EmployeeID,
				ReceiptNo:         invoiceNo,
				ReceiptImageURL:   req.PODImageURL,
				SignatureImageURL: req.SignatureImageURL,
				TotalAmount:       so.TotalAmount,
				DiscountAmount:    so.DiscountAmount,
				StoreCategory:     so.StoreCategory,
				TransactionDate:   now,
				PeriodMonth:       int(now.Month()),
				PeriodYear:        now.Year(),
				Status:            initialTrxStatus,
				PaymentStatus:     initialTrxPaymentStatus,
				PaymentMethod:     paymentMethod,
				PaymentDueDate:    so.PaymentDueDate,
				Notes:             &invoiceNotes,
			}

			for _, soItem := range so.Items {
				if soItem.ActualQuantity > 0 {
					invoice.Items = append(invoice.Items, domain.SalesItem{
						ProductID:          soItem.ProductID,
						Quantity:           soItem.ActualQuantity,
						PriceAtTransaction: soItem.Price,
						DiscountPercent:    soItem.DiscountPercent,
						Subtotal:           float64(soItem.ActualQuantity) * soItem.Price * (1 - soItem.DiscountPercent/100),
					})
				}
			}

			if err := salesRepo.Create(invoice); err != nil {
				return fmt.Errorf("gagal membuat faktur penjualan: %w", err)
			}
			invoiceID = &invoice.ID
		}
		if isTokoTutup {
			so.Status = domain.SOStatusCancelled
		} else {
			so.Status = domain.SOStatusConverted
		}
		resultInvoice = invoice

		so.InvoiceID = invoiceID
		so.ConvertedAt = &now

		// 3. Update DeliveryItem (simpan data retur & POD & Tanda Tangan)
		var di domain.DeliveryItem
		if err := tx.Where("sales_order_id = ?", so.ID).First(&di).Error; err == nil {
			if isTokoTutup {
				di.Status = domain.DeliveryItemFailed
			} else {
				di.Status = domain.DeliveryItemDelivered
			}
			di.DeliveredAt = &now
			di.ReceivedBy = req.ReceivedBy
			di.SalesTransactionID = invoiceID
			if req.PODImageURL != "" {
				di.PODImageURL = req.PODImageURL
			}
			if req.SignatureImageURL != "" {
				di.SignatureImageURL = req.SignatureImageURL
			}
			di.Notes = req.Notes
			// Simpan data retur ke DeliveryItem
			di.ReturnQuantity = totalReturnQty
			di.HasReturn = totalReturnQty > 0
			if totalReturnQty > 0 && len(req.ReturnedItems) > 0 {
				di.ReturnReason = req.ReturnedItems[0].ReturnReason // ambil alasan pertama sebagai ringkasan
			}
			if err := tx.Model(&domain.DeliveryItem{}).Where("id = ?", di.ID).Updates(map[string]interface{}{
				"status":               di.Status,
				"delivered_at":         di.DeliveredAt,
				"received_by":          di.ReceivedBy,
				"sales_transaction_id": di.SalesTransactionID,
				"pod_image_url":        di.PODImageURL,
				"signature_image_url":  di.SignatureImageURL,
				"notes":                di.Notes,
				"return_quantity":      di.ReturnQuantity,
				"has_return":           di.HasReturn,
				"return_reason":        di.ReturnReason,
			}).Error; err != nil {
				return fmt.Errorf("gagal memperbarui status delivery item: %w", err)
			}

			// Sinkronisasi antrean Persetujuan Kasir (SalesPayment PENDING)
			if isTokoTutup {
				pDelQuery := tx.Where("payment_status = ?", "PENDING")
				if invoiceID != nil {
					pDelQuery = pDelQuery.Where("sales_order_id = ? OR sales_transaction_id = ?", so.ID, *invoiceID)
				} else {
					pDelQuery = pDelQuery.Where("sales_order_id = ?", so.ID)
				}
				if err := pDelQuery.Delete(&domain.SalesPayment{}).Error; err != nil {
					return fmt.Errorf("gagal menghapus payment pending toko tutup: %w", err)
				}
			} else {
				var pendingPayments []domain.SalesPayment
				pQuery := tx.Where("payment_status = ?", "PENDING")
				if invoiceID != nil {
					pQuery = pQuery.Where("sales_order_id = ? OR sales_transaction_id = ?", so.ID, *invoiceID)
				} else {
					pQuery = pQuery.Where("sales_order_id = ?", so.ID)
				}
				if err := pQuery.Find(&pendingPayments).Error; err == nil {
					for _, p := range pendingPayments {
						updatedNote := p.Notes
						if totalReturnQty > 0 && !strings.Contains(updatedNote, "[ADJUSTED]") {
							if updatedNote != "" {
								updatedNote = "[ADJUSTED] " + updatedNote
							} else {
								updatedNote = "[ADJUSTED] Penyesuaian Pengiriman/Retur Toko"
							}
						}
						if err := tx.Model(&domain.SalesPayment{}).Where("id = ?", p.ID).Updates(map[string]interface{}{
							"amount": so.TotalAmount,
							"notes":  updatedNote,
						}).Error; err != nil {
							return fmt.Errorf("gagal memperbarui payment pending: %w", err)
						}
					}
				}
			}

			// Update Batch Status
			var batch domain.DeliveryBatch
			if err := tx.Preload("Items").Preload("Samples").First(&batch, di.DeliveryBatchID).Error; err == nil {
				if batch.Status == domain.DeliveryBatchSupervisorApproved || batch.Status == domain.DeliveryBatchPicking {
					batch.Status = domain.DeliveryBatchOnDelivery
					if batch.StartedAt == nil {
						batch.StartedAt = &now
					}
				}
				// Cek apakah semua item selesai
				allCompleted := true
				for _, bItem := range batch.Items {
					if bItem.ID == di.ID {
						continue
					}
					if bItem.Status == domain.DeliveryItemPending {
						allCompleted = false
						break
					}
				}
				if allCompleted {
					batch.Status = domain.DeliveryBatchCompleted
					batch.FinishedAt = &now
					if batch.VehicleID != nil {
						if err := tx.Model(&domain.Vehicle{}).Where("id = ?", *batch.VehicleID).Update("status", "AVAILABLE").Error; err != nil {
							return fmt.Errorf("gagal memperbarui status armada: %w", err)
						}
					}

					// Process samples when batch is completed!
					if err := completeSamplesInTransaction(tx, &batch, now); err != nil {
						return err
					}
				}
				if err := tx.Model(&domain.DeliveryBatch{}).Where("id = ?", batch.ID).Updates(map[string]interface{}{
					"status":      batch.Status,
					"started_at":  batch.StartedAt,
					"finished_at": batch.FinishedAt,
				}).Error; err != nil {
					return fmt.Errorf("gagal memperbarui status batch pengiriman: %w", err)
				}
			}
		}

		// 4. Proses Retur jika ada
		if totalReturnQty > 0 && !isTokoTutup { // Don't process returns if Toko Tutup, they are not returns, just failed delivery
			returnRepo := repository.NewSalesReturnRepository(tx)

			type returnGroup struct {
				items            []domain.SalesReturnItem
				totalAmount      float64
				priceDifference  float64
				claimReferenceNo string
			}
			groups := make(map[string]*returnGroup)

			for _, ri := range req.ReturnedItems {
				rType := ri.ReturnType
				if rType == "" {
					rType = "POTONG_NOTA"
				}

				if groups[rType] == nil {
					groups[rType] = &returnGroup{}
				}

				price := 0.0
				var item *domain.SalesOrderItem
				for i := range so.Items {
					if so.Items[i].ProductID == ri.ProductID {
						item = &so.Items[i]
						price = item.Price
						break
					}
				}
				if price == 0 {
					var prod domain.Product
					if errDb := tx.Select("selling_price").Where("id = ?", ri.ProductID).First(&prod).Error; errDb == nil {
						price = prod.SellingPrice
					}
				}

				cond := "DAMAGED"
				if ri.IsGoodCondition {
					cond = "GOOD"
				}

				retItem := domain.SalesReturnItem{
					ProductID: ri.ProductID,
					Quantity:  ri.ReturnedQuantity,
					Price:     price,
					Reason:    ri.ReturnReason,
					Condition: cond,
				}

				// Khusus TUKAR_PRODUK (Tukar Langsung di tempat)
				if rType == "TUKAR_PRODUK" {
					retItem.Condition = "DEFECTIVE_REPLACED" // Barang yang ditarik dari toko biasanya rusak/cacat
					if ri.ReplacementProductID != nil {
						retItem.ReplacementProductID = ri.ReplacementProductID
						retItem.ReplacementQuantity = ri.ReplacementQuantity

						// Jika produk beda, hitung selisih harga
						if *ri.ReplacementProductID != ri.ProductID {
							var repProd domain.Product
							if errDb := tx.Select("selling_price").Where("id = ?", *ri.ReplacementProductID).First(&repProd).Error; errDb == nil {
								diff := (float64(ri.ReplacementQuantity) * repProd.SellingPrice) - (float64(ri.ReturnedQuantity) * price)
								groups[rType].priceDifference += diff
							}
						}
					}
				}

				// Khusus GARANSI: simpan bukti kerusakan dan nomor klaim
				if rType == "GARANSI" {
					retItem.DamageEvidenceURL = ri.DamageEvidenceURL
					retItem.Condition = "DAMAGED"
					if ri.ClaimReferenceNo != "" {
						groups[rType].claimReferenceNo = ri.ClaimReferenceNo
					}
				}

				groups[rType].items = append(groups[rType].items, retItem)
				groups[rType].totalAmount += float64(ri.ReturnedQuantity) * price
			}

			for rType, grp := range groups {
				returnNo, errNo := returnRepo.GenerateNextReturnNumber(now)
				if errNo != nil {
					return fmt.Errorf("gagal generate nomor retur: %w", errNo)
				}
				notesStr := fmt.Sprintf("[DELIVERY_RETURN] Retur %s dari Pengiriman (Driver Drop Off)", rType)
				ret := &domain.SalesReturn{
					ReturnNo:         returnNo,
					SalesOrderID:     &so.ID,
					TransactionID:    invoiceID,
					BranchID:         so.BranchID,
					EmployeeID:       req.EmployeeID,
					ReturnType:       rType,
					Status:           domain.ReturnStatusPending,
					Notes:            notesStr,
					Items:            grp.items,
					TotalAmount:      grp.totalAmount,
					PriceDifference:  grp.priceDifference,
					ClaimReferenceNo: grp.claimReferenceNo,
				}
				if err := returnRepo.Create(ret); err != nil {
					return fmt.Errorf("gagal menyimpan data retur: %w", err)
				}
			}
		}

		// 5. Create new Sales Order if there are request items (pesanan baru)
		if len(req.RequestItems) > 0 {
			var store domain.Store
			if errStore := tx.First(&store, "id = ?", so.StoreID).Error; errStore == nil {
				salesmanID := so.EmployeeID
				if store.AssignedEmployeeID != nil {
					salesmanID = *store.AssignedEmployeeID
				}

				var company domain.Company
				prefix := ""
				if err := tx.First(&company, "id = ?", so.CompanyID).Error; err == nil {
					if company.Code != "" {
						prefix = company.Code
					} else {
						words := strings.Fields(company.Name)
						if len(words) > 0 {
							name := words[len(words)-1]
							if len(name) > 5 {
								prefix = strings.ToUpper(name[:5])
							} else {
								prefix = strings.ToUpper(name)
							}
						}
					}
				}
				newSO := &domain.SalesOrder{
					BranchID:      so.BranchID,
					CompanyID:     so.CompanyID,
					EmployeeID:    salesmanID,
					StoreID:       so.StoreID,
					StoreCategory: store.DetermineCategory(now),
					Status:        domain.SOStatusDraft,
					Notes:         "Pesanan baru diajukan oleh Driver saat pengiriman.",
					OrderDate:     now,
					PaymentMethod: "SALES_ORDER",
				}

				var totalAmt float64
				for _, reqItem := range req.RequestItems {
					var prod domain.Product
					if errProd := tx.First(&prod, "id = ?", reqItem.ProductID).Error; errProd == nil {
						ppu := reqItem.PiecesPerUnit
						if ppu <= 0 {
							ppu = 1
						}
						subtotal := float64(reqItem.OrderedQuantity) * float64(ppu) * prod.SellingPrice
						newSO.Items = append(newSO.Items, domain.SalesOrderItem{
							ProductID:       reqItem.ProductID,
							OrderedQuantity: reqItem.OrderedQuantity,
							Unit:            prod.Unit,
							PiecesPerUnit:   ppu,
							Price:           prod.SellingPrice,
						})
						totalAmt += subtotal
					}
				}

				newSO.TotalAmount = totalAmt
				if len(newSO.Items) > 0 {
					if errCreate := repository.NewSalesOrderRepository(tx).CreateWithNextSONumber(newSO, prefix); errCreate != nil {
						return fmt.Errorf("gagal membuat pesanan baru dari driver: %w", errCreate)
					}
				}
			}
		}

		if err := tx.Model(&domain.SalesOrder{}).Where("id = ?", so.ID).Updates(map[string]interface{}{
			"status":              so.Status,
			"total_amount":        so.TotalAmount,
			"received_at":         so.ReceivedAt,
			"received_by":         so.ReceivedBy,
			"pod_image_url":       so.PODImageURL,
			"signature_image_url": so.SignatureImageURL,
			"notes":               so.Notes,
			"invoice_id":          so.InvoiceID,
			"converted_at":        so.ConvertedAt,
		}).Error; err != nil {
			return fmt.Errorf("gagal memperbarui sales order: %w", err)
		}
		return nil
	})

	if err == nil && resultInvoice != nil {
		go func() {
			storeName := "Toko"
			if so.Store != nil {
				storeName = so.Store.Name
			} else {
				var store domain.Store
				if errDb := u.db.First(&store, "id = ?", so.StoreID).Error; errDb == nil {
					storeName = store.Name
				}
			}
			driverName := "Driver"
			var driver domain.Employee
			if errDb := u.db.First(&driver, "id = ?", req.EmployeeID).Error; errDb == nil {
				driverName = fmt.Sprintf("%s %s", driver.FirstName, driver.LastName)
			}

			if u.sseHub != nil {
				title := "Pengiriman Selesai (POD)"
				msg := fmt.Sprintf("Driver %s telah menyelesaikan pengiriman nota %s ke toko %s", driverName, so.SONumber, storeName)
				statusText := "DELIVERED"
				if isTokoTutup {
					title = "Pengiriman Gagal (Toko Tutup)"
					msg = fmt.Sprintf("Driver %s melaporkan pengiriman nota %s ke toko %s GAGAL karena toko tutup", driverName, so.SONumber, storeName)
					statusText = "FAILED"
				}

				event := sse.NotificationEvent{
					ID:        resultInvoice.ID.String(),
					Type:      "pod",
					Title:     title,
					Message:   msg,
					CompanyID: so.CompanyID.String(),
					CreatedAt: time.Now().Format("2006-01-02 15:04:05"),
					Data: map[string]interface{}{
						"item_id":     resultInvoice.ID.String(),
						"receipt_no":  so.SONumber,
						"store_name":  storeName,
						"driver_name": driverName,
						"status":      statusText,
					},
				}
				_ = u.sseHub.PublishNotification(context.Background(), event)
			}

			if u.fcmClient != nil {
				var salesman domain.Employee
				if errDb := u.db.First(&salesman, "id = ?", so.EmployeeID).Error; errDb == nil {
					title := "Pengiriman Selesai (POD)"
					body := fmt.Sprintf("Pesanan %s untuk toko %s telah selesai dikirim oleh driver %s.", so.SONumber, storeName, driverName)
					if isTokoTutup {
						title = "Pengiriman Gagal (Toko Tutup)"
						body = fmt.Sprintf("Pengiriman pesanan %s untuk toko %s gagal karena toko tutup (Driver: %s).", so.SONumber, storeName, driverName)
					}
					_ = u.fcmClient.SendPushNotification(salesman.UserID, title, body)
				}
			}
		}()
	}

	return resultInvoice, err
}

// CollectPayment digunakan Driver untuk mencatat tagihan pembayaran yang diterima dari customer.
// Status SO berubah DELIVERED → PAID (jika full) atau tetap DELIVERED dengan PaymentStatus PARTIAL.
func (u *salesOrderUsecase) CollectPayment(req CollectPaymentRequest) (*domain.SalesOrder, error) {
	so, err := u.soRepo.FindByID(req.SOID)
	if err != nil {
		return nil, err
	}
	if so.Status != domain.SOStatusDelivered && so.Status != domain.SOStatusPaid && so.Status != domain.SOStatusInDelivery && so.Status != domain.SOStatusShipped && so.Status != domain.SOStatusConverted {
		return nil, errors.New("pembayaran hanya bisa dicatat setelah barang dikonfirmasi diterima customer atau saat pengiriman")
	}

	now := time.Now()
	resData := make(map[string]interface{})

	// 1. Midtrans Logic (QRIS / VA) - Skip if there is a manual transfer proof
	if (req.PaymentMethod == "QRIS" || req.PaymentMethod == "VA") && req.TransferProofURL == "" && u.midtransClient != nil {
		orderID := fmt.Sprintf("PAY-SO-%s-%d", so.SONumber, now.Unix())

		if req.PaymentMethod == "QRIS" {
			resp, err := u.midtransClient.CreateQRIS(orderID, int64(req.Amount))
			if err == nil && resp != nil {
				qrisData := midtransPkg.ExtractQRISData(resp)
				if qrisData != "" {
					resData["midtrans_qris_url"] = qrisData
					so.MidtransQRISURL = qrisData
				}
				so.MidtransTransactionID = &resp.TransactionID
			}
		} else if req.PaymentMethod == "VA" && req.PaymentBank != "" {
			resp, err := u.midtransClient.CreateBankTransfer(orderID, int64(req.Amount), req.PaymentBank)
			if err == nil && resp != nil {
				so.MidtransTransactionID = &resp.TransactionID
				if len(resp.VaNumbers) > 0 {
					vaNum := resp.VaNumbers[0].VANumber
					bank := resp.VaNumbers[0].Bank
					resData["midtrans_va_number"] = vaNum
					resData["midtrans_bank"] = bank
					so.MidtransVANumber = &vaNum
					so.MidtransBank = &bank
				} else if resp.BillKey != "" {
					billKey := resp.BillKey
					billerCode := resp.BillerCode
					bank := "mandiri"
					resData["midtrans_bill_key"] = billKey
					resData["midtrans_biller_code"] = billerCode
					resData["midtrans_bank"] = bank
					so.MidtransBillKey = &billKey
					so.MidtransBillerCode = &billerCode
					so.MidtransBank = &bank
				}
			}
		}
	}

	// 2. Record Payment Log
	initialStatus := "SUCCESS"
	if (req.PaymentMethod == "CASH" || req.PaymentMethod == "VA") && req.Amount > 0 {
		initialStatus = "PENDING"
	}

	payment := domain.SalesPayment{
		SalesOrderID:         &so.ID,
		Amount:               req.Amount,
		PaymentMethod:        req.PaymentMethod,
		CollectedBy:          req.CollectedBy,
		EmployeeID:           req.EmployeeID,
		PaymentStatus:        initialStatus,
		PaymentDate:          now,
		CompanyBankAccountID: req.CompanyBankAccountID,
	}
	if req.TransferProofURL != "" {
		payment.TransferProofURL = &req.TransferProofURL
	}
	if err := u.soRepo.AddPayment(&payment); err != nil {
		return nil, err
	}

	// Allow CASH to update SO totals so the delivery/salesman app reflects the collected payment
	// Only the payment log remains PENDING for finance verification.

	// 3. Update SO Totals
	if req.Amount > 0 {
		so.PaymentCollectedAmount += req.Amount
	}
	so.PaymentCollectedAt = &now
	so.PaymentMethod = req.PaymentMethod

	if req.PaymentMethod == "TEMPO" {
		so.PaymentDueDate = req.PaymentDueDate
		if so.PaymentCollectedAmount >= so.TotalAmount && so.TotalAmount > 0 {
			so.PaymentStatus = "PAID"
			so.Status = domain.SOStatusPaid
		} else {
			so.PaymentStatus = "TEMPO"
		}
	} else {
		if so.PaymentCollectedAmount >= so.TotalAmount {
			so.PaymentStatus = "PAID"
			so.Status = domain.SOStatusPaid
		} else {
			so.PaymentStatus = "PARTIAL"
		}
	}

	if err := u.soRepo.Update(so); err != nil {
		return nil, err
	}

	// Update corresponding SalesTransaction (Invoice) if it exists
	if so.InvoiceID != nil {
		var st domain.SalesTransaction
		if err := u.db.First(&st, *so.InvoiceID).Error; err == nil {
			st.PaymentMethod = req.PaymentMethod
			st.PaymentDueDate = req.PaymentDueDate

			if req.Amount > 0 {
				st.PaidAmount = st.PaidAmount + req.Amount
			}

			if req.PaymentMethod == "TEMPO" {
				if st.PaidAmount >= st.TotalAmount && st.TotalAmount > 0 {
					st.PaymentStatus = string(domain.PaymentStatusPaid)
				} else {
					st.PaymentStatus = "TEMPO"
				}
			} else {
				if st.TotalAmount <= 0 {
					st.PaymentStatus = string(domain.PaymentStatusPaid)
					st.Status = "VERIFIED"
				} else {
					// Recalculate based on SUCCESS payments only
					var successPaidSum float64
					u.db.Model(&domain.SalesPayment{}).
						Where("(sales_transaction_id = ? OR sales_order_id = ?) AND payment_status = 'SUCCESS'", st.ID, so.ID).
						Select("COALESCE(SUM(amount), 0)").Scan(&successPaidSum)

					if successPaidSum >= st.TotalAmount {
						st.PaymentStatus = string(domain.PaymentStatusPaid)
					} else if successPaidSum > 0 || st.PaidAmount > 0 {
						st.PaymentStatus = string(domain.PaymentStatusPartial)
					} else {
						st.PaymentStatus = string(domain.PaymentStatusUnpaid)
					}
				}
			}
			_ = u.db.Save(&st)
		}
	}

	// 4. Update Delivery Item & Batch Totals (if linked)
	var di domain.DeliveryItem
	if err := u.db.Where("sales_order_id = ?", so.ID).First(&di).Error; err == nil {
		di.PaymentCollected = true
		di.PaymentAmount += req.Amount
		di.PaymentMethod = req.PaymentMethod
		di.PaymentCollectedAt = &now
		_ = u.db.Save(&di)

		// Update Batch Totals
		var batch domain.DeliveryBatch
		if err := u.db.First(&batch, di.DeliveryBatchID).Error; err == nil {
			if req.PaymentMethod == "CASH" {
				batch.TotalCashCollected += req.Amount
			} else if req.PaymentMethod == "QRIS" || req.PaymentMethod == "VA" {
				batch.TotalTransferCollected += req.Amount
			} else if req.PaymentMethod == "TEMPO" {
				tempoAmount := so.TotalAmount - req.Amount
				if tempoAmount < 0 {
					tempoAmount = 0
				}
				batch.TotalTempoCollected += tempoAmount
			}
			_ = u.db.Save(&batch)
		}
	}

	if u.sseHub != nil {
		go func() {
			methodText := req.PaymentMethod
			event := sse.NotificationEvent{
				ID:        so.ID.String(),
				Type:      "payment",
				Title:     "Pembayaran Baru Diterima",
				Message:   fmt.Sprintf("Pembayaran tagihan %s sebesar Rp %.2f via %s diterima.", so.SONumber, req.Amount, methodText),
				CompanyID: so.CompanyID.String(),
				CreatedAt: time.Now().Format("2006-01-02 15:04:05"),
				Data: map[string]interface{}{
					"so_id":     so.ID.String(),
					"so_number": so.SONumber,
					"amount":    req.Amount,
					"method":    methodText,
				},
			}
			_ = u.sseHub.PublishNotification(context.Background(), event)
		}()
	}

	return so, nil
}

// =============================================================================
// Admin Nota: Penyesuaian Pesanan (Diskon + Adjust Qty Item)
// =============================================================================

// AdjustSO memungkinkan Admin Nota untuk:
// 1. Memberikan diskon (dalam Rupiah) pada total pesanan
// 2. Menambah, mengubah, atau menghapus item pesanan (diskon per-item dalam persen)
// 3. Menyesuaikan reservasi stok di gudang (melepas reservasi lama & mereservasi ulang)
// 4. Menambahkan catatan penyesuaian
// Hanya bisa dilakukan pada SO dengan status DRAFT, CONFIRMED, atau WAITING_STOCK.
func (u *salesOrderUsecase) AdjustSO(id uuid.UUID, req AdjustSORequest) (*domain.SalesOrder, error) {
	err := u.db.Transaction(func(tx *gorm.DB) error {
		soRepo := repository.NewSalesOrderRepository(tx)

		so, err := soRepo.FindByID(id)
		if err != nil {
			return err
		}
		if so.Status != domain.SOStatusDraft && so.Status != domain.SOStatusConfirmed && so.Status != domain.SOStatusWaitingStock {
			return errors.New("penyesuaian hanya bisa dilakukan pada SO berstatus DRAFT, CONFIRMED, atau WAITING_STOCK")
		}

		// 1. Hapus batch item SO yang lama dan items lama
		_ = tx.Where("sales_order_item_id IN (SELECT id FROM sales_order_items WHERE sales_order_id = ?)", so.ID).Delete(&domain.SalesOrderItemBatch{}).Error
		if err := tx.Where("sales_order_id = ?", so.ID).Delete(&domain.SalesOrderItem{}).Error; err != nil {
			return err
		}

		// 2. Hapus semua item bonus lama
		if err := tx.Where("sales_order_id = ?", so.ID).Delete(&domain.SalesOrderBonusItem{}).Error; err != nil {
			return err
		}

		// 3. Buat item baru & hitung total amount
		var totalAmount float64

		for _, reqItem := range req.Items {
			ppu := reqItem.PiecesPerUnit
			if ppu <= 0 {
				ppu = 1
			}
			qty := reqItem.OrderedQuantity * ppu

			price := reqItem.Price.Float64()
			discountPercent := reqItem.DiscountPercent.Float64()

			// Subtotal per item: qty * price * (1 - discount_percent/100)
			subtotal := float64(qty) * price * (1.0 - discountPercent/100.0)
			totalAmount += subtotal

			item := domain.SalesOrderItem{
				SalesOrderID:    so.ID,
				ProductID:       reqItem.ProductID,
				Quantity:        qty,
				OrderedQuantity: reqItem.OrderedQuantity,
				Unit:            reqItem.Unit,
				PiecesPerUnit:   ppu,
				Price:           price,
				DiscountPercent: discountPercent,
				Subtotal:        subtotal,
			}

			if err := tx.Create(&item).Error; err != nil {
				return err
			}
		}

		// 3.5. Buat item bonus baru
		for _, reqBonus := range req.BonusItems {
			bItem := domain.SalesOrderBonusItem{
				SalesOrderID:     so.ID,
				BonusInventoryID: reqBonus.BonusInventoryID,
				Quantity:         reqBonus.Quantity,
				Notes:            reqBonus.Notes,
			}
			if err := tx.Create(&bItem).Error; err != nil {
				return err
			}
		}

		// 4. Kurangi diskon global (Rupiah) jika ada
		discAmount := req.DiscountAmount.Float64()
		if discAmount > 0 {
			so.DiscountAmount = discAmount
			totalAmount -= discAmount
			if totalAmount < 0 {
				totalAmount = 0
			}
		} else {
			so.DiscountAmount = 0
		}
		so.TotalAmount = totalAmount

		// 5. Update Status SO
		if so.Status == domain.SOStatusWaitingStock {
			so.Status = domain.SOStatusConfirmed
		}
		if so.AdminNotaConfirmedAt == nil && so.Status == domain.SOStatusConfirmed {
			now := time.Now()
			so.AdminNotaConfirmedAt = &now
		}

		// 6. Catatan Penyesuaian
		if req.AdjustNotes != "" {
			if so.Notes != "" {
				so.Notes += "\n[ADJUST] " + req.AdjustNotes
			} else {
				so.Notes = "[ADJUST] " + req.AdjustNotes
			}
		}

		// Update SO without automatically saving associations
		// since we already manually deleted and recreated Items and BonusItems.
		if err := tx.Omit("Items", "BonusItems", "Batches", "Returns").Save(so).Error; err != nil {
			return err
		}

		// 7. Sinkronisasi ke linked invoice (jika sudah ada) dan antrean Persetujuan Kasir (SalesPayment PENDING)
		if so.InvoiceID != nil {
			tx.Model(&domain.SalesTransaction{}).Where("id = ?", *so.InvoiceID).Update("total_amount", so.TotalAmount)
		}

		var pendingPayments []domain.SalesPayment
		pQuery := tx.Where("payment_status = ?", "PENDING")
		if so.InvoiceID != nil {
			pQuery = pQuery.Where("sales_order_id = ? OR sales_transaction_id = ?", so.ID, *so.InvoiceID)
		} else {
			pQuery = pQuery.Where("sales_order_id = ?", so.ID)
		}
		if err := pQuery.Find(&pendingPayments).Error; err == nil {
			for _, p := range pendingPayments {
				updatedNote := p.Notes
				if !strings.Contains(updatedNote, "[ADJUSTED]") {
					if updatedNote != "" {
						updatedNote = "[ADJUSTED] " + updatedNote
					} else {
						updatedNote = "[ADJUSTED] Penyesuaian SO oleh Admin Nota"
					}
				}
				tx.Model(&domain.SalesPayment{}).Where("id = ?", p.ID).Updates(map[string]interface{}{
					"amount": so.TotalAmount,
					"notes":  updatedNote,
				})
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	return u.soRepo.FindByID(id)
}

// GetStoreDebtHistory mengambil riwayat transaksi yang belum lunas untuk sebuah toko.
func (u *salesOrderUsecase) GetStoreDebtHistory(storeID uuid.UUID) ([]domain.SalesTransaction, error) {
	return u.salesRepo.GetOutstandingByStore(storeID)
}

// GetMTSettlementQueue mengembalikan daftar pembayaran penagihan MT
func (u *salesOrderUsecase) GetMTSettlementQueue() ([]domain.SalesPayment, error) {
	return u.soRepo.GetMTSettlementQueue()
}

// ApproveMTSettlement memverifikasi pembayaran penagihan MT
func (u *salesOrderUsecase) ApproveMTSettlement(paymentID uuid.UUID, adminID uuid.UUID) error {
	payment, err := u.soRepo.GetPaymentByID(paymentID)
	if err != nil {
		return err
	}
	if payment.IsVerified {
		return errors.New("pembayaran sudah diverifikasi sebelumnya")
	}

	payment.IsVerified = true
	payment.VerifiedByID = &adminID
	payment.PaymentStatus = "SUCCESS"
	now := time.Now()
	payment.ApprovedAt = &now
	payment.UpdatedAt = now
	if err := u.soRepo.UpdatePayment(payment); err != nil {
		return err
	}

	// Update SO status jika lunas
	so, err := u.soRepo.FindByID(*payment.SalesOrderID)
	if err != nil {
		return err
	}

	if so.PaymentCollectedAmount >= so.TotalAmount {
		so.PaymentStatus = "PAID"
		so.Status = domain.SOStatusPaid
		u.soRepo.Update(so)
	}
	return nil
}

// UpdateMTOrderStatus meng-update status SO dan mensinkronisasikan ke PO MT.
func (u *salesOrderUsecase) UpdateMTOrderStatus(id uuid.UUID, status string) error {
	var so domain.SalesOrder
	if err := u.db.Preload("Items.Batches").First(&so, "id = ?", id).Error; err != nil {
		return err
	}

	err := u.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&so).Update("status", status).Error; err != nil {
			return err
		}
		if so.SONumber != "" {
			if err := tx.Model(&domain.MTPurchaseOrder{}).Where("po_number = ?", so.SONumber).Update("status", status).Error; err != nil {
				return err
			}
		}

		// Mutasi stok saat Sales MT konfirmasi muatan dan mulai pengiriman (IN_DELIVERY)
		if status == "IN_DELIVERY" && so.IsMT {
			wRepo := repository.NewWarehouseRepository(tx)
			for _, item := range so.Items {
				actualQty := item.OrderedQuantity
				if actualQty > 0 {
					batches, err := wRepo.GetAvailableBatches(so.BranchID, item.ProductID, "GOOD")
					if err != nil {
						return err
					}
					remainingToDeduct := actualQty
					for idx := range batches {
						b := &batches[idx]
						if b.Quantity <= 0 {
							continue
						}
						deduct := b.Quantity
						if deduct > remainingToDeduct {
							deduct = remainingToDeduct
						}
						b.Quantity -= deduct
						if err := wRepo.UpdateStock(b); err != nil {
							return err
						}
						remainingToDeduct -= deduct
						if remainingToDeduct <= 0 {
							break
						}
					}

					doStr := "DO-MT-" + so.SONumber
					if so.DeliveryOrderNo != nil && *so.DeliveryOrderNo != "" {
						doStr = *so.DeliveryOrderNo
					}
					_ = wRepo.CreateLog(&domain.WarehouseLog{
						BranchID:  so.BranchID,
						ProductID: item.ProductID,
						Type:      "OUT",
						Source:    "MUTASI STOK MT: " + doStr,
						Quantity:  actualQty,
					})
				}
			}
		}

		return nil
	})
	return err
}

// SuperAdminFastTrackComplete mengeksekusi penyelesaian satu-klik SO lama/pending
// mencakup persetujuan nota, pengiriman, faktur verified, persetujuan kasir, dan pengesahan omzet.
func (u *salesOrderUsecase) SuperAdminFastTrackComplete(id uuid.UUID, superAdminID uuid.UUID) (*domain.SalesOrder, error) {
	var resultSO *domain.SalesOrder

	err := u.db.Transaction(func(tx *gorm.DB) error {
		soRepo := repository.NewSalesOrderRepository(tx)
		salesRepo := repository.NewSalesTransactionRepository(tx)

		so, err := soRepo.FindByID(id)
		if err != nil {
			return fmt.Errorf("pesanan SO tidak ditemukan: %w", err)
		}
		if so.DeletedAt.Valid {
			return errors.New("pesanan SO yang telah dihapus tidak dapat diproses")
		}
		if so.Status == domain.SOStatusCancelled || so.Status == domain.SOStatusRejected {
			return fmt.Errorf("pesanan SO dalam status %s tidak dapat diproses", so.Status)
		}

		now := time.Now()

		// Temukan employee_id yang valid untuk Super Admin (jika ada)
		var adminEmpID *uuid.UUID
		var adminEmp domain.Employee
		if err := tx.Select("id").Where("id = ? OR user_id = ?", superAdminID, superAdminID).First(&adminEmp).Error; err == nil {
			adminEmpID = &adminEmp.ID
		}

		// 1. Verifikasi Admin Nota
		if so.AdminNotaConfirmedAt == nil {
			so.AdminNotaConfirmedAt = &now
			so.AdminNotaConfirmedByID = adminEmpID
		}
		if so.ConfirmedAt == nil {
			so.ConfirmedAt = &now
			so.ConfirmedByID = adminEmpID
		}

		// 2. Realisasi Item: set actual_quantity = ordered_quantity, reserved_quantity = 0
		var calculatedTotal float64
		for i := range so.Items {
			item := &so.Items[i]
			if item.ActualQuantity <= 0 {
				item.ActualQuantity = item.OrderedQuantity
			}
			ppu := item.PiecesPerUnit
			if ppu <= 0 {
				ppu = 1
			}
			item.PiecesPerUnit = ppu
			item.ReservedQuantity = 0
			item.Subtotal = float64(item.ActualQuantity*ppu) * item.Price
			if item.DiscountPercent > 0 {
				item.Subtotal -= (item.Subtotal * item.DiscountPercent / 100.0)
			}
			calculatedTotal += item.Subtotal

			if err := tx.Model(&domain.SalesOrderItem{}).Where("id = ?", item.ID).Updates(map[string]interface{}{
				"actual_quantity":   item.ActualQuantity,
				"pieces_per_unit":   ppu,
				"reserved_quantity": 0,
				"subtotal":          item.Subtotal,
			}).Error; err != nil {
				return fmt.Errorf("gagal memperbarui item pesanan: %w", err)
			}
		}

		if so.TotalAmount <= 0 && calculatedTotal > 0 {
			so.TotalAmount = calculatedTotal - so.DiscountAmount
			if so.TotalAmount < 0 {
				so.TotalAmount = 0
			}
		}

		// 3. Konfirmasi Pengiriman (Delivery & POD)
		if so.ShippedAt == nil {
			so.ShippedAt = &now
		}
		if so.ReceivedAt == nil {
			so.ReceivedAt = &now
		}
		if so.ReceivedBy == "" {
			if so.Store != nil && so.Store.OwnerName != "" {
				so.ReceivedBy = so.Store.OwnerName
			} else if so.Store != nil && so.Store.Name != "" {
				so.ReceivedBy = so.Store.Name
			} else {
				so.ReceivedBy = "Super Admin (Selesai Otomatis)"
			}
		}

		// 4. Generate atau Update Faktur (SalesTransaction)
		var invoice *domain.SalesTransaction
		paymentMethod := so.PaymentMethod
		if paymentMethod == "" {
			paymentMethod = "CASH"
		}

		invoiceNotes := "[FAST-TRACK SUPER ADMIN] Selesai otomatis oleh Super Admin"
		if so.Notes != "" {
			invoiceNotes = invoiceNotes + "\n\nCatatan Awal:\n" + so.Notes
		}

		if so.InvoiceID != nil {
			var existingInvoice domain.SalesTransaction
			if err := tx.Preload("Items").First(&existingInvoice, "id = ?", *so.InvoiceID).Error; err == nil {
				invoice = &existingInvoice
				invoice.TotalAmount = so.TotalAmount
				invoice.DiscountAmount = so.DiscountAmount
				invoice.Status = "VERIFIED"
				invoice.VerifiedAt = &now
				invoice.VerifiedByID = adminEmpID
				invoice.PaymentStatus = string(domain.PaymentStatusPaid)
				invoice.PaidAmount = so.TotalAmount
				invoice.PaymentMethod = paymentMethod
				invoice.PaymentDueDate = nil
				invoice.Notes = &invoiceNotes
				invoice.UpdatedAt = now

				// Hapus items lama dan ganti dengan yang sinkron
				_ = tx.Where("sales_transaction_id = ?", invoice.ID).Delete(&domain.SalesItem{}).Error
				var newItems []domain.SalesItem
				for _, soItem := range so.Items {
					if soItem.ActualQuantity > 0 {
						newItems = append(newItems, domain.SalesItem{
							SalesTransactionID: invoice.ID,
							ProductID:          soItem.ProductID,
							Quantity:           soItem.ActualQuantity,
							PriceAtTransaction: soItem.Price,
							Subtotal:           soItem.Subtotal,
						})
					}
				}
				invoice.Items = newItems
				if err := tx.Omit(clause.Associations).Save(invoice).Error; err != nil {
					return fmt.Errorf("gagal memperbarui faktur penjualan: %w", err)
				}
				if len(newItems) > 0 {
					if err := tx.Create(&newItems).Error; err != nil {
						return fmt.Errorf("gagal memperbarui rincian faktur: %w", err)
					}
				}
			}
		} else {
			invoiceCount, _ := salesRepo.CountByCompanyAndDate(so.CompanyID, now)
			invoiceNo := fmt.Sprintf("INV-%s-%04d", now.Format("20060102"), invoiceCount+1)

			for attempt := 1; attempt <= 50; attempt++ {
				var count int64
				tx.Model(&domain.SalesTransaction{}).Where("receipt_no = ?", invoiceNo).Count(&count)
				if count == 0 {
					break
				}
				invoiceNo = fmt.Sprintf("INV-%s-%04d-%d", now.Format("20060102"), invoiceCount+int64(attempt), time.Now().UnixNano()%10000)
			}

			trxDate := so.OrderDate
			if trxDate.IsZero() {
				trxDate = now
			}

			invoice = &domain.SalesTransaction{
				CompanyID:       so.CompanyID,
				StoreID:         &so.StoreID,
				EmployeeID:      so.EmployeeID,
				ReceiptNo:       invoiceNo,
				TotalAmount:     so.TotalAmount,
				DiscountAmount:  so.DiscountAmount,
				StoreCategory:   so.StoreCategory,
				TransactionDate: trxDate,
				PeriodMonth:     int(trxDate.Month()),
				PeriodYear:      trxDate.Year(),
				Status:          "VERIFIED",
				VerifiedAt:      &now,
				VerifiedByID:    adminEmpID,
				PaymentStatus:   string(domain.PaymentStatusPaid),
				PaidAmount:      so.TotalAmount,
				PaymentMethod:   paymentMethod,
				Notes:           &invoiceNotes,
			}

			for _, soItem := range so.Items {
				if soItem.ActualQuantity > 0 {
					invoice.Items = append(invoice.Items, domain.SalesItem{
						ProductID:          soItem.ProductID,
						Quantity:           soItem.ActualQuantity,
						PriceAtTransaction: soItem.Price,
						DiscountPercent:    soItem.DiscountPercent,
						Subtotal:           soItem.Subtotal,
					})
				}
			}

			if err := salesRepo.Create(invoice); err != nil {
				return fmt.Errorf("gagal membuat faktur penjualan: %w", err)
			}
			so.InvoiceID = &invoice.ID
		}

		// 5. Update status DeliveryItem dan DeliveryBatch jika ada
		var di domain.DeliveryItem
		if err := tx.Where("sales_order_id = ?", so.ID).First(&di).Error; err == nil {
			di.Status = domain.DeliveryItemDelivered
			di.DeliveredAt = &now
			di.ReceivedBy = so.ReceivedBy
			di.SalesTransactionID = so.InvoiceID
			di.PaymentCollected = true
			di.PaymentAmount = so.TotalAmount
			di.PaymentMethod = paymentMethod
			di.PaymentCollectedAt = &now

			_ = tx.Model(&domain.DeliveryItem{}).Where("id = ?", di.ID).Updates(map[string]interface{}{
				"status":               di.Status,
				"delivered_at":         di.DeliveredAt,
				"received_by":          di.ReceivedBy,
				"sales_transaction_id": di.SalesTransactionID,
				"payment_collected":    true,
				"payment_amount":       di.PaymentAmount,
				"payment_method":       di.PaymentMethod,
				"payment_collected_at": di.PaymentCollectedAt,
			}).Error

			// Update batch jika ada
			var batch domain.DeliveryBatch
			if err := tx.Preload("Items").First(&batch, di.DeliveryBatchID).Error; err == nil {
				allCompleted := true
				for _, bItem := range batch.Items {
					if bItem.ID == di.ID {
						continue
					}
					if bItem.Status == domain.DeliveryItemPending {
						allCompleted = false
						break
					}
				}
				if allCompleted {
					batch.Status = domain.DeliveryBatchCompleted
					batch.FinishedAt = &now
					if batch.VehicleID != nil {
						_ = tx.Model(&domain.Vehicle{}).Where("id = ?", *batch.VehicleID).Update("status", "AVAILABLE").Error
					}
				}
				_ = tx.Model(&domain.DeliveryBatch{}).Where("id = ?", batch.ID).Updates(map[string]interface{}{
					"status":      batch.Status,
					"finished_at": batch.FinishedAt,
				}).Error
			}
		}

		// 6. Persetujuan Kasir (SalesPayment SUCCESS)
		paymentEmployeeID := so.EmployeeID
		if adminEmpID != nil {
			paymentEmployeeID = *adminEmpID
		}

		var payments []domain.SalesPayment
		pQuery := tx.Where("sales_order_id = ?", so.ID)
		if so.InvoiceID != nil {
			pQuery = tx.Where("sales_order_id = ? OR sales_transaction_id = ?", so.ID, *so.InvoiceID)
		}
		if err := pQuery.Find(&payments).Error; err == nil && len(payments) > 0 {
			for _, p := range payments {
				note := p.Notes
				if !strings.Contains(note, "[APPROVED]") {
					if note != "" {
						note = "[APPROVED] " + note
					} else {
						note = "[APPROVED] Disetujui Super Admin"
					}
				}
				_ = tx.Model(&domain.SalesPayment{}).Where("id = ?", p.ID).Updates(map[string]interface{}{
					"payment_status": "SUCCESS",
					"is_verified":    true,
					"verified_by_id": adminEmpID,
					"approved_at":    &now,
					"amount":         so.TotalAmount,
					"notes":          note,
				}).Error
			}
		} else {
			// Buat entri SalesPayment berstatus SUCCESS
			newPayment := domain.SalesPayment{
				SalesOrderID:       &so.ID,
				SalesTransactionID: so.InvoiceID,
				EmployeeID:         paymentEmployeeID,
				Amount:             so.TotalAmount,
				PaymentMethod:      paymentMethod,
				PaymentStatus:      "SUCCESS",
				IsVerified:         true,
				VerifiedByID:       adminEmpID,
				ApprovedAt:         &now,
				PaymentDate:        now,
				CollectedBy:        "Super Admin",
				Notes:              "Penyelesaian otomatis oleh Super Admin (Persetujuan Kasir)",
			}
			if err := tx.Create(&newPayment).Error; err != nil {
				return fmt.Errorf("gagal membuat data pembayaran kasir: %w", err)
			}
		}

		// 7. Update SO Final Status & Metadata
		soUpdates := map[string]interface{}{
			"status":                     domain.SOStatusConverted,
			"converted_at":               now,
			"payment_status":             "PAID",
			"payment_collected_amount":   so.TotalAmount,
			"payment_collected_at":       now,
			"payment_method":             paymentMethod,
			"shipped_at":                 so.ShippedAt,
			"received_at":                so.ReceivedAt,
			"received_by":                so.ReceivedBy,
			"total_amount":               so.TotalAmount,
			"invoice_id":                 so.InvoiceID,
			"admin_nota_confirmed_at":    so.AdminNotaConfirmedAt,
			"admin_nota_confirmed_by_id": so.AdminNotaConfirmedByID,
			"confirmed_at":               so.ConfirmedAt,
			"confirmed_by_id":            so.ConfirmedByID,
		}

		if err := tx.Model(&domain.SalesOrder{}).Where("id = ?", so.ID).Updates(soUpdates).Error; err != nil {
			return fmt.Errorf("gagal menyimpan pembaruan pesanan: %w", err)
		}

		so.Status = domain.SOStatusConverted
		so.ConvertedAt = &now
		so.PaymentStatus = "PAID"
		so.PaymentCollectedAmount = so.TotalAmount
		so.PaymentCollectedAt = &now
		so.PaymentMethod = paymentMethod

		resultSO = so
		return nil
	})

	if err != nil {
		return nil, err
	}

	// Invalidate report cache
	if resultSO != nil {
		invalidateReportsCache(context.Background())
	}

	return resultSO, nil
}

// SuperAdminBulkComplete menyelesaikan semua pesanan SO yang tertahan secara massal
// Mendukung pemrosesan per cabang spesifik, daftar cabang, maupun seluruh cabang sekaligus (ALL_BRANCHES).
func (u *salesOrderUsecase) SuperAdminBulkComplete(req BulkSuperAdminCompleteRequest, superAdminID uuid.UUID) (*BulkCompleteResult, error) {
	q := u.db.Model(&domain.SalesOrder{}).Where("deleted_at IS NULL")

	if req.BranchID != nil && *req.BranchID != uuid.Nil {
		q = q.Where("branch_id = ?", *req.BranchID)
	} else if len(req.BranchIDs) > 0 {
		q = q.Where("branch_id IN (?)", req.BranchIDs)
	}

	if req.CompanyID != nil && *req.CompanyID != uuid.Nil {
		q = q.Where("company_id = ?", *req.CompanyID)
	}

	if len(req.SOIDs) > 0 {
		q = q.Where("id IN (?)", req.SOIDs)
	}

	// Hanya proses SO yang berstatus CONFIRMED (bukan DRAFT, WAITING_STOCK, dll)
	q = q.Where("status = ?", domain.SOStatusConfirmed)

	var pendingOrders []domain.SalesOrder
	if err := q.Select("id, so_number, branch_id, company_id, status").Find(&pendingOrders).Error; err != nil {
		return nil, fmt.Errorf("gagal mengambil daftar pesanan tertahan: %w", err)
	}

	result := &BulkCompleteResult{
		TotalProcessed: len(pendingOrders),
		CompletedIDs:   make([]string, 0, len(pendingOrders)),
		Errors:         make([]string, 0),
	}

	for _, order := range pendingOrders {
		_, err := u.SuperAdminFastTrackComplete(order.ID, superAdminID)
		if err != nil {
			result.FailedCount++
			result.Errors = append(result.Errors, fmt.Sprintf("SO %s (%s): %v", order.SONumber, order.ID, err))
		} else {
			result.SuccessCount++
			result.CompletedIDs = append(result.CompletedIDs, order.ID.String())
		}
	}

	// Invalidate report cache setelah batch pemrosesan selesai
	invalidateReportsCache(context.Background())

	return result, nil
}

// GetPendingStats menghitung statistik riil seluruh pesanan SO yang tertahan di database
// tanpa batasan pagination, untuk keperluan pemantauan dan penyelesaian massal Super Admin.
func (u *salesOrderUsecase) GetPendingStats(branchID *uuid.UUID, branchIDs []uuid.UUID, companyID *uuid.UUID) (*PendingSOStats, error) {
	q := u.db.Model(&domain.SalesOrder{}).Where("deleted_at IS NULL")

	if branchID != nil && *branchID != uuid.Nil {
		q = q.Where("branch_id = ?", *branchID)
	} else if len(branchIDs) > 0 {
		q = q.Where("branch_id IN (?)", branchIDs)
	}

	if companyID != nil && *companyID != uuid.Nil {
		q = q.Where("company_id = ?", *companyID)
	}

	// Status yang dianggap tertahan / belum selesai
	q = q.Where("status NOT IN (?)", []domain.SalesOrderStatus{
		domain.SOStatusConverted,
		domain.SOStatusCancelled,
		domain.SOStatusRejected,
	})

	type StatusStat struct {
		Status string
		Count  int64
		Amount float64
	}

	var stats []StatusStat
	if err := q.Select("status, COUNT(*) as count, COALESCE(SUM(total_amount), 0) as amount").Group("status").Scan(&stats).Error; err != nil {
		return nil, fmt.Errorf("gagal menghitung statistik pesanan tertahan: %w", err)
	}

	result := &PendingSOStats{}
	for _, s := range stats {
		result.TotalPending += s.Count
		result.TotalPendingAmount += s.Amount

		switch domain.SalesOrderStatus(s.Status) {
		case domain.SOStatusConfirmed:
			result.ConfirmedCount += s.Count
		case domain.SOStatusWaitingStock:
			result.WaitingStockCount += s.Count
		case domain.SOStatusDraft:
			result.DraftCount += s.Count
		default:
			result.OtherPendingCount += s.Count
		}
	}

	return result, nil
}
