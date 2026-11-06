package usecase

import (
	"bytes"
	"context"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"

	"github.com/midtrans/midtrans-go/coreapi"
	"example.com/hris/backend/internal/domain"
	"example.com/hris/backend/internal/repository"
	midtransPkg "example.com/hris/backend/pkg/midtrans"
	"example.com/hris/backend/pkg/redisclient"
	"example.com/hris/backend/pkg/sse"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SalesUsecase interface {
	ManualEntry(req ManualEntryRequest) (*domain.SalesTransaction, error)
	CreateDirectSale(req DirectSaleRequest) (*domain.SalesTransaction, error)
	GetNextDirectSaleInvoiceNo(branchID *uuid.UUID, date time.Time) (string, error)
	CalculateKPI(employeeID uuid.UUID, month, year int) error
	GenerateKPIReportExcel(month, year int, branchIDs []uuid.UUID) ([]byte, error)
	GetSummaryKPI(periodType, dateStr string, month, year int, branchIDs []uuid.UUID, salesType string) (map[string]interface{}, error)
	ReconcileAllNotaReturnItems() (int, error)

	// Mobile App Endpoints
	CreateTransaction(req CreateTransactionRequest) (*domain.SalesTransaction, error)
	SettleOmzetMT(id uuid.UUID, adminID *uuid.UUID) (*domain.SalesTransaction, error)
	GetPendingTransactions(employeeID uuid.UUID) ([]domain.SalesTransaction, error)
	GetHistoryTransactions(employeeID uuid.UUID, branchID *uuid.UUID) ([]domain.SalesTransaction, error)
	VerifyTransaction(id uuid.UUID, approverID *uuid.UUID, req VerifyTransactionRequest) error
	VerifyPenerimaan(id uuid.UUID, adminNotaID uuid.UUID, req VerifyPenerimaanRequest) error
	GetPenerimaanTempoPayments(page, limit int, branchID string, status string, search string, salesman string, date string) ([]domain.SalesPayment, int64, float64, error)
	VerifyPenerimaanTempoPayment(paymentID, adminID uuid.UUID, notes string) error
	RedeliverTransaction(id uuid.UUID, adminNotaID uuid.UUID, reason string) error
	CancelFailedDeliveryTransaction(id uuid.UUID, reason string) error
	GetPerformanceList(month, year int, branchIDs []uuid.UUID) ([]domain.SalesKPI, error)
	GetYearlyPerformanceList(year int, branchIDs []uuid.UUID) ([]domain.SalesKPI, error)
	GetDailyPerformanceList(date time.Time, branchIDs []uuid.UUID) ([]domain.SalesKPI, error)
	GetPerformanceByDateRange(startDate, endDate time.Time, branchIDs []uuid.UUID) ([]domain.SalesKPI, error)
	GetNooEcTracking(startDate, endDate time.Time, branchID, employeeID *uuid.UUID, salesType string) (*domain.NooEcTrackingResponse, error)
	ExportNooEcTrackingExcel(startDate, endDate time.Time, branchID, employeeID *uuid.UUID, salesType string) ([]byte, error)
	GetEmployeeProductRecap(employeeID uuid.UUID, month, year int) ([]domain.ProductRecap, error)
	GetAllTransactionsPaginated(page, limit int, companyID, status, search, salesman, month, year, date, salesType, branchID, paymentMethod, paymentStatus string, targetEntity ...string) ([]domain.SalesTransaction, int64, float64, error)
	GetAllOutstandingInvoices(companyID, branchID, salesType string) ([]domain.SalesTransaction, error)
	ExportTransactionsExcel(companyID, status, search, salesman, month, year, date, salesType, branchID, paymentMethod, paymentStatus string, targetEntity ...string) ([]byte, error)
	GetTransactionByID(id uuid.UUID) (*domain.SalesTransaction, error)
	UpdateTransaction(id uuid.UUID, req ManualEntryRequest) error
	UpdateTransactionMobile(id uuid.UUID, req MobileUpdateTransactionRequest) error
	DeleteTransaction(id uuid.UUID) error
	SoftDeleteTransaction(id uuid.UUID, deletedByID uuid.UUID, reason string) error
	DeletePendingTransactionBySales(id, employeeID uuid.UUID) error
	RestoreTransaction(id uuid.UUID) error
	SetKPITarget(employeeID uuid.UUID, month, year int, targetOmzet float64, targetNewStores int, targetDailyVisits int, targetDailyNewStores int, workingTerritory string, receivableLimitPercent float64, maxReceivableLimit float64, commission float64) error
	DeleteKPITarget(id uuid.UUID) error
	GetVisitPlan(employeeID uuid.UUID, date time.Time) ([]VisitPlanItem, error)
	GetTransactionByReceipt(receiptNo string) (*domain.SalesTransaction, error)

	// Payment & Debt Management
	RecordPayment(req RecordPaymentRequest) error
	GetOutstandingTransactionsByStore(storeID uuid.UUID) ([]domain.SalesTransaction, error)
	GetTransactionsByDueDate(date time.Time, employeeID uuid.UUID) ([]domain.SalesTransaction, error)
	GetTransactionsByStatusAndCompany(status string, companyID uuid.UUID) ([]domain.SalesTransaction, error)
	GetDeliveryPending(companyID uuid.UUID) ([]domain.SalesTransaction, error)
	RejectTransaction(id uuid.UUID, notes string) error
	CountByCompanyAndDate(companyID uuid.UUID, date time.Time) (int64, error)
	HandleMidtransNotification(notificationPayload map[string]interface{}) error
	GetProductSalesDistribution(productIDs []uuid.UUID, companyID uuid.UUID) ([]map[string]interface{}, error)
	ConvertToSalesOrder(transactionID uuid.UUID, adminID uuid.UUID) (*domain.SalesOrder, error)

	// Visit Management
	RecordVisit(req RecordVisitRequest) error
	GetVisitHistory(limit, offset int, employeeID, branchID *uuid.UUID, salesType string, search string, startDate, endDate *time.Time, summary bool) ([]domain.SalesVisit, int64, error)
	CreateVisit(req RecordVisitRequest) (*domain.SalesVisit, error)
	CreateBatchVisits(reqs []RecordVisitRequest) ([]*domain.SalesVisit, error)
	UpdateVisit(id uuid.UUID, req RecordVisitRequest) error
	DeleteVisit(id uuid.UUID) error
	GetSalesKasbon(companyID uuid.UUID) ([]map[string]interface{}, error)
	GetSettlementQueue() ([]domain.SalesTransaction, error)
	GetPendingVisits(employeeID uuid.UUID) ([]domain.SalesVisit, error)

	GetVisitPlanBuilder(employeeID uuid.UUID, targetSalesmanID *uuid.UUID, branchID *uuid.UUID) ([]VisitPlanItem, error)

	GetTopProducts(companyID uuid.UUID, limit int) ([]map[string]interface{}, error)

	CreateManualVisitPlan(employeeID uuid.UUID, stores []StorePlanRequest, planDate time.Time) error
	DeleteSpecificPlan(employeeID uuid.UUID, storeID uuid.UUID, planDate time.Time) error

	RegeneratePayment(transactionID string) (*domain.SalesTransaction, error)
	GetCityRevenueReport(startDate, endDate time.Time, companyID uuid.UUID, branchIDs []uuid.UUID, salesType string, productIDs []uuid.UUID) (*domain.CityRevenueReportResponse, error)
	ExportCityRevenueExcel(startDate, endDate time.Time, companyID uuid.UUID, branchIDs []uuid.UUID, salesType string, productIDs []uuid.UUID) ([]byte, error)
}

type StorePlanRequest struct {
	StoreID      uuid.UUID           `json:"store_id"`
	StoreIDCamel uuid.UUID           `json:"storeId"` // Fallback
	Projections  []ProjectionRequest `json:"projections"`
}

type ProjectionRequest struct {
	ProductID uuid.UUID `json:"product_id"`
	Quantity  int       `json:"quantity"`
}

type VisitPlanItem struct {
	StoreID               uuid.UUID                 `json:"store_id"`
	StoreName             string                    `json:"store_name"`
	Address               string                    `json:"address"`
	Status                string                    `json:"status"` // PLANNED, COMPLETED, EXTRA
	IsExtra               bool                      `json:"is_extra"`
	VisitID               *uuid.UUID                `json:"visit_id,omitempty"`
	LastVisitDate         *time.Time                `json:"last_visit_date"`
	LastTransactionDate   *time.Time                `json:"last_transaction_date"`
	LastTransactionAmount float64                   `json:"last_transaction_amount"`
	LastTransactionItems  string                    `json:"last_transaction_items"`
	IsPriority            bool                      `json:"is_priority"`
	Latitude              float64                   `json:"latitude"`
	Longitude             float64                   `json:"longitude"`
	AssignedEmployeeID    *uuid.UUID                `json:"assigned_employee_id,omitempty"`
	AssignedEmployeeName  string                    `json:"assigned_employee_name,omitempty"`
	Projections           []VisitPlanProjectionItem `json:"projections"`
	ProjectedOmzet        float64                   `json:"projected_omzet"`
}

type VisitPlanProjectionItem struct {
	ProductID   uuid.UUID `json:"product_id"`
	ProductName string    `json:"product_name"`
	Quantity    int       `json:"quantity"`
	Price       float64   `json:"price"`
}

type CreateTransactionRequest struct {
	ClientID             string             `json:"client_id"`
	CompanyID            uuid.UUID          `json:"company_id"`
	StoreID              uuid.UUID          `json:"store_id"`
	EmployeeID           uuid.UUID          `json:"employee_id"`
	ReceiptNo            string             `json:"receipt_no"`
	ReceiptImageURL      string             `json:"receipt_image_url"`
	PhysicalNotaImageURL string             `json:"physical_nota_image_url"`
	SignatureImageURL    string             `json:"signature_image_url"`
	TotalAmount          float64            `json:"total_amount"`
	PaidAmount           float64            `json:"paid_amount"`
	PaymentDueDate       *time.Time         `json:"payment_due_date"`
	StoreCategory        string             `json:"store_category"`
	TransactionDate      time.Time          `json:"transaction_date"`
	PaymentMethod        string             `json:"payment_method"`
	DPPaymentMethod      string             `json:"dp_payment_method"` // For TEMPO: 'CASH', 'QRIS', 'VA'
	Bank                 string             `json:"bank"`              // For VA: 'bca', 'bni', 'bri'
	Tenor                int                `json:"tenor"`             // For TEMPO: number of installment terms
	Notes                string             `json:"notes"`
	CompanyBankAccountID *uuid.UUID         `json:"company_bank_account_id"`
	VisitID              *uuid.UUID         `json:"visit_id"`
	Latitude             *float64           `json:"latitude"`
	Longitude            *float64           `json:"longitude"`
	Items                []SalesItemRequest `json:"items"`

	// --- Barang Sampingan (Krat, Botol, Galon, Mangkok, Piring, Gelas, Retur, dll) ---
	// Dikirim bersamaan dengan items produk utama.
	// Nilai retur secara otomatis menjadi negatif di backend.

	// --- Modern Trade (MT) ---
	// IsMT: true jika transaksi ini dibuat oleh Sales MT.
	// Saat PaymentMethod=SALES_ORDER dan IsMT=true, SO yang dibuat akan diklaim
	// oleh pusat/pabrik (FactoryCompanyID), bukan cabang.
	IsMT             bool       `json:"is_mt"`
	FactoryCompanyID *uuid.UUID `json:"factory_company_id"`

	// --- Distributor Order ---
	IsDistributorOrder bool       `json:"is_distributor_order"`
	DistributorStoreID *uuid.UUID `json:"distributor_store_id"`

	BonusItems []SalesBonusItemRequest `json:"bonus_items"`
}

type SalesBonusItemRequest struct {
	BonusInventoryID uuid.UUID `json:"bonus_inventory_id"`
	Quantity         int       `json:"quantity"`
	Notes            string    `json:"notes"`
}

type SalesItemRequest struct {
	ProductID       uuid.UUID `json:"product_id"`
	Quantity        int       `json:"quantity"`         // Total pieces
	OrderedQuantity int       `json:"ordered_quantity"` // e.g., 2 (Kartons)
	Unit            string    `json:"unit"`             // e.g., 'KRT'
	PiecesPerUnit   int       `json:"pieces_per_unit"`  // e.g., 24
	Price           float64   `json:"price"`
	DiscountPercent float64   `json:"discount_percent"`
}

type RecordPaymentRequest struct {
	TransactionID        uuid.UUID  `json:"transaction_id"`
	EmployeeID           uuid.UUID  `json:"employee_id"`
	Amount               float64    `json:"amount" binding:"required"`
	PaymentMethod        string     `json:"payment_method"`
	PaymentDate          time.Time  `json:"payment_date"`
	NextDueDate          *time.Time `json:"next_due_date"`
	Notes                string     `json:"notes"`
	TransferProofURL     string     `json:"transfer_proof_url"`
	CompanyBankAccountID *uuid.UUID `json:"company_bank_account_id"`
}

type VerifyTransactionRequest struct {
	ReceiptNo   string  `json:"receipt_no"`
	TotalAmount float64 `json:"total_amount"`
	Notes       string  `json:"notes"` // Catatan dari finalisasi
}

type VerifyPenerimaanItemRequest struct {
	ProductID uuid.UUID `json:"product_id" binding:"required"`
	Quantity  int       `json:"quantity"`
	Notes     string    `json:"notes"`
}

type VerifyPenerimaanRequest struct {
	Items []VerifyPenerimaanItemRequest `json:"items"`
	Notes string                        `json:"notes"`
}

type RecordVisitRequest struct {
	ClientID   string    `json:"client_id"`
	EmployeeID uuid.UUID `json:"employee_id"`
	StoreID    uuid.UUID `json:"store_id"`
	Latitude   float64   `json:"latitude"`
	Longitude  float64   `json:"longitude"`
	SelfieURL  string    `json:"selfie_url"`
	Type       string    `json:"type"` // CHECKIN, CHECKOUT
	CheckTime  time.Time `json:"check_time"`
	Notes      string    `json:"notes"`
}

type salesUsecase struct {
	salesRepo         repository.SalesTransactionRepository
	performanceRepo   repository.PerformanceRepository
	storeRepo         repository.StoreRepository
	attendanceRepo    repository.AttendanceRepository
	companyRepo       repository.CompanyRepository
	employeeRepo      repository.EmployeeRepository
	soUsecase         SalesOrderUsecase
	visitRepo         repository.SalesVisitRepository
	visitPlanRepo     repository.SalesVisitPlanRepository
	salesTransferRepo repository.SalesTransferRepository
	midtransClient    *midtransPkg.MidtransClient
	db                *gorm.DB
	sseHub            *sse.Hub
}

func NewSalesUsecase(
	salesRepo repository.SalesTransactionRepository,
	performanceRepo repository.PerformanceRepository,
	storeRepo repository.StoreRepository,
	attendanceRepo repository.AttendanceRepository,
	companyRepo repository.CompanyRepository,
	employeeRepo repository.EmployeeRepository,
	soUsecase SalesOrderUsecase,
	visitRepo repository.SalesVisitRepository,
	visitPlanRepo repository.SalesVisitPlanRepository,
	salesTransferRepo repository.SalesTransferRepository,
	midtransClient *midtransPkg.MidtransClient,
	db *gorm.DB,
	sseHub *sse.Hub,
) SalesUsecase {
	return &salesUsecase{
		salesRepo:         salesRepo,
		performanceRepo:   performanceRepo,
		storeRepo:         storeRepo,
		attendanceRepo:    attendanceRepo,
		companyRepo:       companyRepo,
		employeeRepo:      employeeRepo,
		soUsecase:         soUsecase,
		visitRepo:         visitRepo,
		visitPlanRepo:     visitPlanRepo,
		salesTransferRepo: salesTransferRepo,
		midtransClient:    midtransClient,
		db:                db,
		sseHub:            sseHub,
	}
}

type ManualEntryRequest struct {
	DeferPaymentToCashier bool   `json:"-"`
	ClientID              string `json:"client_id"`

	CompanyID        *uuid.UUID              `json:"company_id"`
	StoreID          *uuid.UUID              `json:"store_id"`
	EmployeeID       *uuid.UUID              `json:"employee_id" binding:"required"`
	ReceiptNo        string                  `json:"receipt_no"`
	ReceiptImageURL  string                  `json:"receipt_image_url"`
	TransferProofURL string                  `json:"transfer_proof_url"`
	TotalAmount      float64                 `json:"total_amount" binding:"required"`
	DiscountAmount   float64                 `json:"discount_amount"`
	StoreCategory    string                  `json:"store_category" binding:"required"`   // 'TOKO_LAMA' or 'TOKO_BARU'
	TransactionDate  string                  `json:"transaction_date" binding:"required"` // YYYY-MM-DD
	Status           string                  `json:"status"`                              // 'PENDING', 'VERIFIED', 'REJECTED'
	PaymentStatus    string                  `json:"payment_status"`                      // 'UNPAID', 'PARTIAL', 'PAID'
	PaymentMethod    string                  `json:"payment_method"`                      // 'CASH', 'TEMPO', 'TRANSFER', 'QRIS', 'POTONG_GAJI'
	DPPaymentMethod  string                  `json:"dp_payment_method"`                   // For TEMPO: 'CASH', 'QRIS', 'VA', 'TRANSFER'
	PaidAmount       float64                 `json:"paid_amount"`
	PaymentDueDate   *time.Time              `json:"payment_due_date"`
	Tenor            int                     `json:"tenor"`
	Items            []SalesItemRequest      `json:"items"`
	BonusItems       []SalesBonusItemRequest `json:"bonus_items"`
	Notes            string                  `json:"notes"`
}

type DirectSaleRequest struct {
	DeferPaymentToCashier bool                    `json:"-"`
	DistributorRequestID  *uuid.UUID              `json:"distributor_request_id"`
	ClientID              string                  `json:"client_id"`
	CompanyID             *uuid.UUID              `json:"company_id"`
	BranchID              uuid.UUID               `json:"branch_id" binding:"required"`
	SalesChannel          string                  `json:"sales_channel"` // DIRECT_FACTORY, DIRECT_BRANCH, EMPLOYEE_SALE
	BuyerType             string                  `json:"buyer_type"`    // GENERAL_WALKIN, REGISTERED_STORE, EMPLOYEE
	BuyerName             string                  `json:"buyer_name"`
	BuyerPhone            string                  `json:"buyer_phone"`
	BuyerEmployeeID       *uuid.UUID              `json:"buyer_employee_id"`
	StoreID               *uuid.UUID              `json:"store_id"`
	CashierEmployeeID     *uuid.UUID              `json:"cashier_employee_id"`
	KasPabrikID           *uuid.UUID              `json:"kas_pabrik_id"`
	ReceiptNo             string                  `json:"receipt_no"`
	TransactionDate       string                  `json:"transaction_date" binding:"required"` // YYYY-MM-DD
	PaymentDueDate        *time.Time              `json:"payment_due_date"`                    // YYYY-MM-DD
	PaymentMethod         string                  `json:"payment_method"`                      // CASH, QRIS, TRANSFER, POTONG_GAJI, TEMPO
	DPPaymentMethod       string                  `json:"dp_payment_method"`
	PaymentStatus         string                  `json:"payment_status"` // PAID, UNPAID, PARTIAL
	PaidAmount            float64                 `json:"paid_amount"`
	TotalAmount           float64                 `json:"total_amount" binding:"required"`
	DiscountAmount        float64                 `json:"discount_amount"`
	Notes                 string                  `json:"notes"`
	Items                 []SalesItemRequest      `json:"items" binding:"required"`
	BonusItems            []SalesBonusItemRequest `json:"bonus_items"`
}

type MobileUpdateTransactionRequest struct {
	TotalAmount float64            `json:"total_amount" binding:"required"`
	Items       []SalesItemRequest `json:"items"`
}

func (u *salesUsecase) resolveCompanyID(companyID *uuid.UUID, employeeID *uuid.UUID, storeID *uuid.UUID) uuid.UUID {
	if companyID != nil && *companyID != uuid.Nil {
		return *companyID
	}
	if employeeID != nil && *employeeID != uuid.Nil {
		var emp domain.Employee
		if err := u.db.Where("id = ?", *employeeID).First(&emp).Error; err == nil && emp.CompanyID != nil && *emp.CompanyID != uuid.Nil {
			return *emp.CompanyID
		}
	}
	if storeID != nil && *storeID != uuid.Nil {
		var store domain.Store
		if err := u.db.Preload("AssignedEmployee").Where("id = ?", *storeID).First(&store).Error; err == nil {
			if store.CompanyID != uuid.Nil {
				return store.CompanyID
			}
			if store.AssignedEmployee != nil && store.AssignedEmployee.CompanyID != nil && *store.AssignedEmployee.CompanyID != uuid.Nil {
				return *store.AssignedEmployee.CompanyID
			}
		}
	}
	var firstComp domain.Company
	if err := u.db.Order("created_at ASC").First(&firstComp).Error; err == nil && firstComp.ID != uuid.Nil {
		return firstComp.ID
	}
	return uuid.Nil
}

func (u *salesUsecase) ManualEntry(req ManualEntryRequest) (*domain.SalesTransaction, error) {
	trxDate, err := time.Parse("2006-01-02", req.TransactionDate)
	if err != nil {
		return nil, errors.New("invalid transaction date format, expected YYYY-MM-DD")
	}

	companyID := u.resolveCompanyID(req.CompanyID, req.EmployeeID, req.StoreID)
	if companyID == uuid.Nil {
		return nil, errors.New("gagal menentukan company_id untuk transaksi")
	}
	var employeeID uuid.UUID
	if req.EmployeeID != nil {
		employeeID = *req.EmployeeID
	}

	if req.ReceiptNo == "" {
		count, _ := u.salesRepo.CountByCompanyAndDate(companyID, trxDate)
		receiptNo := fmt.Sprintf("INV/%s/%03d", trxDate.Format("20060102"), count+1)
		req.ReceiptNo = receiptNo
	}

	var clientID *string
	if req.ClientID != "" {
		clientID = &req.ClientID
	}

	paymentMethod := req.PaymentMethod
	if paymentMethod == "" {
		paymentMethod = "CASH"
	}

	paymentStatus := req.PaymentStatus
	paidAmount := req.PaidAmount
	if paymentStatus == "" {
		if paymentMethod == "POTONG_GAJI" {
			paymentStatus = "UNPAID"
			paidAmount = 0
		} else if paymentMethod == "TEMPO" {
			if paidAmount > 0 && paidAmount < req.TotalAmount {
				paymentStatus = "PARTIAL"
			} else if paidAmount >= req.TotalAmount && req.TotalAmount > 0 {
				paymentStatus = "PAID"
			} else {
				paymentStatus = "UNPAID"
				paidAmount = 0
			}
		} else {
			paymentStatus = "PAID"
			if paidAmount <= 0 {
				paidAmount = req.TotalAmount
			}
		}
	} else if paymentMethod == "TEMPO" && paidAmount == 0 && paymentStatus == "PAID" {
		paidAmount = req.TotalAmount
	}
	if req.DeferPaymentToCashier {
		paymentStatus = "UNPAID"
		paidAmount = 0
		req.PaidAmount = 0
	}

	var trxBranchID *uuid.UUID
	if employeeID != uuid.Nil {
		var emp domain.Employee
		if err := u.db.Where("id = ?", employeeID).First(&emp).Error; err == nil {
			if emp.OperatingBranchID != nil && *emp.OperatingBranchID != uuid.Nil {
				trxBranchID = emp.OperatingBranchID
			} else if emp.BranchID != nil {
				trxBranchID = emp.BranchID
			}
		}
	}
	if trxBranchID == nil && req.StoreID != nil && *req.StoreID != uuid.Nil {
		var store domain.Store
		if err := u.db.Preload("AssignedEmployee").Where("id = ?", *req.StoreID).First(&store).Error; err == nil {
			if store.AssignedEmployee != nil {
				if store.AssignedEmployee.OperatingBranchID != nil && *store.AssignedEmployee.OperatingBranchID != uuid.Nil {
					trxBranchID = store.AssignedEmployee.OperatingBranchID
				} else if store.AssignedEmployee.BranchID != nil {
					trxBranchID = store.AssignedEmployee.BranchID
				}
			}
		}
	}

	trx := &domain.SalesTransaction{
		CompanyID:       companyID,
		BranchID:        trxBranchID,
		StoreID:         req.StoreID,
		EmployeeID:      employeeID,
		ReceiptNo:       req.ReceiptNo,
		ReceiptImageURL: req.ReceiptImageURL,
		TotalAmount:     req.TotalAmount,
		DiscountAmount:  req.DiscountAmount,
		StoreCategory:   req.StoreCategory,
		TransactionDate: trxDate,
		PaymentDueDate:  req.PaymentDueDate,
		PaymentMethod:   paymentMethod,
		PaymentStatus:   paymentStatus,
		PaidAmount:      paidAmount,
		Tenor: func() int {
			if req.Tenor <= 0 {
				return 1
			}
			return req.Tenor
		}(),
		PeriodMonth: int(trxDate.Month()),
		PeriodYear:  trxDate.Year(),
		Status:      req.Status,
		Notes:       &req.Notes,
		ClientID:    clientID,
	}

	for _, itemReq := range req.Items {
		trx.Items = append(trx.Items, domain.SalesItem{
			ProductID:          itemReq.ProductID,
			Quantity:           itemReq.Quantity,
			PriceAtTransaction: itemReq.Price,
			DiscountPercent:    itemReq.DiscountPercent,
			Subtotal:           float64(itemReq.Quantity) * itemReq.Price * (1 - itemReq.DiscountPercent/100),
		})
	}

	for _, bonusReq := range req.BonusItems {
		trx.BonusItems = append(trx.BonusItems, domain.SalesTransactionBonusItem{
			BonusInventoryID: bonusReq.BonusInventoryID,
			Quantity:         bonusReq.Quantity,
			Notes:            bonusReq.Notes,
		})
	}

	// Generate installments if TEMPO
	if paymentMethod == "TEMPO" && trx.Tenor > 0 {
		remainingBalance := trx.TotalAmount - paidAmount
		if remainingBalance > 0 {
			installmentAmount := remainingBalance / float64(trx.Tenor)
			installmentAmount = float64(int(installmentAmount*100)) / 100.0

			baseDueDate := trxDate.AddDate(0, 0, 14)
			if req.PaymentDueDate != nil {
				baseDueDate = *req.PaymentDueDate
			}

			intervalDays := 14
			if req.PaymentDueDate != nil && trx.Tenor > 1 {
				totalDays := int(req.PaymentDueDate.Sub(trxDate).Hours() / 24)
				if totalDays >= trx.Tenor*7 {
					intervalDays = totalDays / trx.Tenor
				}
			}

			for i := 1; i <= trx.Tenor; i++ {
				dueDate := trxDate.AddDate(0, 0, intervalDays*i)
				if i == trx.Tenor && req.PaymentDueDate != nil {
					dueDate = *req.PaymentDueDate
				} else if trx.Tenor == 1 && req.PaymentDueDate != nil {
					dueDate = baseDueDate
				}
				amount := installmentAmount
				if i == trx.Tenor {
					amount = remainingBalance - (installmentAmount * float64(trx.Tenor-1))
				}
				trx.Installments = append(trx.Installments, domain.SalesInstallment{
					TermNumber: i,
					DueDate:    dueDate,
					Amount:     amount,
					PaidAmount: 0,
					Status:     "UNPAID",
				})
			}
		}
	}

	err = u.salesRepo.Create(trx)
	if err != nil {
		return nil, err
	}

	// Sinkronisasi retur barang sampingan (wadah/jerigen/botol bekas) dari catatan nota ke stok gudang & sales_returns
	if req.Notes != "" {
		var processedName string
		if req.EmployeeID != nil {
			var emp domain.Employee
			if err := u.db.Where("id = ?", *req.EmployeeID).First(&emp).Error; err == nil {
				processedName = strings.TrimSpace(emp.FirstName + " " + emp.LastName)
			}
		}
		if updatedNotes, errSync := u.syncNotaReturnItems(u.db, trx, req.Notes, processedName); errSync == nil && updatedNotes != "" {
			trx.Notes = &updatedNotes
			_ = u.salesRepo.UpdateTransactionOnly(trx)
		}
	}

	// Catat SalesPayment PENDING untuk persetujuan kasir (jika CASH, TRANSFER, QRIS, VA, atau TEMPO dengan DP > 0)
	if !req.DeferPaymentToCashier && (paymentMethod == "CASH" || paymentMethod == "TRANSFER" || paymentMethod == "QRIS" || paymentMethod == "VA") && req.TotalAmount > 0 {
		paymentAmount := req.TotalAmount
		if req.PaidAmount > 0 {
			paymentAmount = req.PaidAmount
		}
		var proof *string
		if req.TransferProofURL != "" {
			proof = &req.TransferProofURL
		}
		salesPayment := &domain.SalesPayment{
			SalesTransactionID: &trx.ID,
			EmployeeID:         employeeID,
			Amount:             paymentAmount,
			PaymentMethod:      paymentMethod,
			PaymentStatus:      "PENDING",
			PaymentDate:        trxDate,
			TransferProofURL:   proof,
			Notes:              fmt.Sprintf("Setoran faktur penjualan (%s - %s)", trx.ReceiptNo, paymentMethod),
		}
		_ = u.db.Create(salesPayment).Error
	} else if !req.DeferPaymentToCashier && paymentMethod == "TEMPO" && (req.PaidAmount > 0 || trx.DPAmount > 0) {
		dpAmount := req.PaidAmount
		if dpAmount <= 0 {
			dpAmount = trx.DPAmount
		}
		dpMethod := req.DPPaymentMethod
		if dpMethod == "" {
			dpMethod = "CASH"
		}
		var proof *string
		if req.TransferProofURL != "" {
			proof = &req.TransferProofURL
		}
		salesPayment := &domain.SalesPayment{
			SalesTransactionID: &trx.ID,
			EmployeeID:         employeeID,
			Amount:             dpAmount,
			PaymentMethod:      dpMethod,
			PaymentStatus:      "PENDING",
			PaymentDate:        trxDate,
			TransferProofURL:   proof,
			Notes:              fmt.Sprintf("Uang Muka (DP) Transaksi Tempo (%s - %s)", trx.ReceiptNo, dpMethod),
		}
		_ = u.db.Create(salesPayment).Error
	}

	// Recalculate KPI for this employee if already verified/approved
	if trx.Status == "VERIFIED" || trx.Status == "APPROVED" {
		err = u.CalculateKPI(employeeID, int(trxDate.Month()), trxDate.Year())
		if err != nil {
			// Log the error but don't fail the transaction entry
		}
	}

	invalidateReportsCache(context.Background())
	return trx, nil
}

func (u *salesUsecase) GetNextDirectSaleInvoiceNo(branchID *uuid.UUID, date time.Time) (string, error) {
	prefix := "FP-DIR"
	if branchID != nil && *branchID != uuid.Nil {
		var branch domain.Branch
		if err := u.db.Select("name").Where("id = ?", *branchID).First(&branch).Error; err == nil {
			upperName := strings.ToUpper(branch.Name)
			if strings.Contains(upperName, "PABRIK") || strings.Contains(upperName, "PUSAT") {
				prefix = "FP-PBR"
			} else {
				prefix = "FP-CAB"
			}
		}
	}

	dateStr := date.Format("20060102")
	var count int64
	u.db.Model(&domain.SalesTransaction{}).
		Where("receipt_no LIKE ?", prefix+"/"+dateStr+"/%").
		Count(&count)

	return fmt.Sprintf("%s/%s/%04d", prefix, dateStr, count+1), nil
}

func (u *salesUsecase) CreateDirectSale(req DirectSaleRequest) (*domain.SalesTransaction, error) {
	trxDate, err := time.Parse("2006-01-02", req.TransactionDate)
	if err != nil {
		trxDate = time.Now()
	}

	if req.BranchID == uuid.Nil && req.CashierEmployeeID != nil && *req.CashierEmployeeID != uuid.Nil {
		var cashierEmp domain.Employee
		if err := u.db.Where("id = ?", *req.CashierEmployeeID).First(&cashierEmp).Error; err == nil {
			if cashierEmp.OperatingBranchID != nil && *cashierEmp.OperatingBranchID != uuid.Nil {
				req.BranchID = *cashierEmp.OperatingBranchID
			} else if cashierEmp.BranchID != nil {
				req.BranchID = *cashierEmp.BranchID
			}
		}
	}

	companyID := u.resolveCompanyID(req.CompanyID, req.CashierEmployeeID, req.StoreID)
	if companyID == uuid.Nil && req.BranchID != uuid.Nil {
		var branch domain.Branch
		if err := u.db.Where("id = ?", req.BranchID).First(&branch).Error; err == nil && branch.CompanyID != nil && *branch.CompanyID != uuid.Nil {
			companyID = *branch.CompanyID
		}
	}
	if companyID == uuid.Nil {
		var firstComp domain.Company
		if err := u.db.Order("created_at ASC").First(&firstComp).Error; err == nil {
			companyID = firstComp.ID
		}
	}

	salesChannel := req.SalesChannel
	if salesChannel == "" {
		salesChannel = domain.SalesChannelDirectFactory
	}
	var distributorRequest *domain.SalesTransfer
	if req.DistributorRequestID != nil && *req.DistributorRequestID != uuid.Nil {
		if err := u.db.First(&distributorRequest, "id = ?", *req.DistributorRequestID).Error; err != nil {
			return nil, errors.New("request distributor tidak ditemukan")
		}
		if distributorRequest.RequestSource != "DISTRIBUTOR" || distributorRequest.Status != domain.SalesTransferPending {
			return nil, errors.New("request distributor belum disetujui atau sudah diproses")
		}
		if req.StoreID == nil || distributorRequest.DistributorStoreID == nil || *req.StoreID != *distributorRequest.DistributorStoreID {
			return nil, errors.New("distributor pada faktur tidak sesuai dengan request")
		}
	}
	// Distributor purchases are direct sales, never branch-transfer documents.
	// Keep the source branch explicit so stock is deducted from the selected
	// factory/branch (BranchID), while the distributor remains the buyer.
	if req.StoreID != nil && *req.StoreID != uuid.Nil {
		var store domain.Store
		if err := u.db.Select("id", "customer_type", "name").First(&store, "id = ?", *req.StoreID).Error; err != nil {
			return nil, fmt.Errorf("distributor/customer tidak ditemukan")
		}
		if strings.EqualFold(store.CustomerType, "DISTRIBUTOR") && salesChannel != domain.SalesChannelDirectFactory && salesChannel != domain.SalesChannelDirectBranch {
			return nil, errors.New("penjualan distributor harus menggunakan kanal penjualan langsung")
		}
	}

	buyerType := req.BuyerType
	if buyerType == "" {
		if req.BuyerEmployeeID != nil && *req.BuyerEmployeeID != uuid.Nil {
			buyerType = domain.BuyerTypeEmployee
		} else if req.StoreID != nil && *req.StoreID != uuid.Nil {
			buyerType = domain.BuyerTypeRegisteredStore
		} else {
			buyerType = domain.BuyerTypeGeneralWalkIn
		}
	}

	buyerName := strings.TrimSpace(req.BuyerName)
	if buyerType == domain.BuyerTypeEmployee && req.BuyerEmployeeID != nil && *req.BuyerEmployeeID != uuid.Nil {
		var emp domain.Employee
		if err := u.db.Where("id = ?", *req.BuyerEmployeeID).First(&emp).Error; err == nil {
			if buyerName == "" {
				buyerName = strings.TrimSpace(emp.FirstName + " " + emp.LastName)
			}
			if req.BuyerPhone == "" {
				req.BuyerPhone = emp.PhoneNumber
			}
		}
	} else if buyerType == domain.BuyerTypeRegisteredStore && req.StoreID != nil && *req.StoreID != uuid.Nil {
		var store domain.Store
		if err := u.db.Where("id = ?", *req.StoreID).First(&store).Error; err == nil {
			if buyerName == "" {
				buyerName = store.Name
			}
			if req.BuyerPhone == "" {
				req.BuyerPhone = store.PhoneNumber
			}
		}
	}
	if buyerName == "" {
		buyerName = "Pelanggan Luar / Walk-in"
	}

	receiptNo := strings.TrimSpace(req.ReceiptNo)
	if receiptNo == "" {
		generatedNo, err := u.GetNextDirectSaleInvoiceNo(&req.BranchID, trxDate)
		if err == nil {
			receiptNo = generatedNo
		} else {
			receiptNo = fmt.Sprintf("FP-DIR/%s/%04d", trxDate.Format("20060102"), time.Now().Unix()%10000)
		}
	}

	paymentStatus := req.PaymentStatus
	paidAmount := req.PaidAmount
	if paymentStatus == "" {
		if req.PaymentMethod == "POTONG_GAJI" {
			paymentStatus = "UNPAID"
			paidAmount = 0
		} else if req.PaymentMethod == "TEMPO" {
			if paidAmount > 0 && paidAmount < req.TotalAmount {
				paymentStatus = "PARTIAL"
			} else if paidAmount >= req.TotalAmount && req.TotalAmount > 0 {
				paymentStatus = "PAID"
			} else {
				paymentStatus = "UNPAID"
				paidAmount = 0
			}
		} else {
			paymentStatus = "PAID"
			if paidAmount <= 0 {
				paidAmount = req.TotalAmount
			}
		}
	}
	if req.DeferPaymentToCashier {
		paymentStatus = "UNPAID"
		paidAmount = 0
		req.PaidAmount = 0
	}

	var assignedEmployeeID uuid.UUID
	if req.CashierEmployeeID != nil && *req.CashierEmployeeID != uuid.Nil {
		assignedEmployeeID = *req.CashierEmployeeID
	} else {
		var emp domain.Employee
		if err := u.db.Where("branch_id = ?", req.BranchID).Order("created_at ASC").First(&emp).Error; err == nil {
			assignedEmployeeID = emp.ID
		} else if err := u.db.Where("company_id = ?", companyID).Order("created_at ASC").First(&emp).Error; err == nil {
			assignedEmployeeID = emp.ID
		}
	}

	var clientID *string
	if req.ClientID != "" {
		clientID = &req.ClientID
	}

	var createdTrx *domain.SalesTransaction

	err = u.db.Transaction(func(tx *gorm.DB) error {
		trx := &domain.SalesTransaction{
			CompanyID:         companyID,
			BranchID:          &req.BranchID,
			StoreID:           req.StoreID,
			EmployeeID:        assignedEmployeeID,
			CashierEmployeeID: req.CashierEmployeeID,
			BuyerEmployeeID:   req.BuyerEmployeeID,
			SalesChannel:      salesChannel,
			BuyerType:         buyerType,
			BuyerName:         buyerName,
			BuyerPhone:        req.BuyerPhone,
			ReceiptNo:         receiptNo,
			TotalAmount:       req.TotalAmount,
			DiscountAmount:    req.DiscountAmount,
			StoreCategory:     "WALK_IN",
			TransactionDate:   trxDate,
			PaymentDueDate:    req.PaymentDueDate,
			PeriodMonth:       int(trxDate.Month()),
			PeriodYear:        trxDate.Year(),
			Status:            "PENDING",
			PaymentStatus:     paymentStatus,
			PaidAmount:        paidAmount,
			PaymentMethod:     req.PaymentMethod,
			KasPabrikID:       req.KasPabrikID,
			Notes:             &req.Notes,
			ClientID:          clientID,
		}

		for _, itemReq := range req.Items {
			subtotal := float64(itemReq.Quantity) * itemReq.Price * (1 - itemReq.DiscountPercent/100)
			trx.Items = append(trx.Items, domain.SalesItem{
				ProductID:          itemReq.ProductID,
				Quantity:           itemReq.Quantity,
				PriceAtTransaction: itemReq.Price,
				DiscountPercent:    itemReq.DiscountPercent,
				Subtotal:           subtotal,
			})
		}

		for _, bonusReq := range req.BonusItems {
			trx.BonusItems = append(trx.BonusItems, domain.SalesTransactionBonusItem{
				BonusInventoryID: bonusReq.BonusInventoryID,
				Quantity:         bonusReq.Quantity,
				Notes:            bonusReq.Notes,
			})
		}

		if err := tx.Create(trx).Error; err != nil {
			return fmt.Errorf("gagal membuat transaksi penjualan langsung: %w", err)
		}

		// Sinkronisasi retur barang sampingan (wadah/jerigen/botol bekas) dari catatan nota ke stok gudang & sales_returns
		if req.Notes != "" {
			if updatedNotes, errSync := u.syncNotaReturnItems(tx, trx, req.Notes, buyerName); errSync == nil && updatedNotes != "" {
				trx.Notes = &updatedNotes
				_ = tx.Model(trx).Update("notes", updatedNotes)
			}
		}
		if distributorRequest != nil {
			if err := tx.Model(&domain.SalesTransfer{}).
				Where("id = ? AND status = ?", distributorRequest.ID, domain.SalesTransferPending).
				Updates(map[string]interface{}{"status": "INVOICED", "sales_transaction_id": trx.ID}).Error; err != nil {
				return fmt.Errorf("gagal mengaitkan request distributor ke faktur: %w", err)
			}
		}

		// Potong stok barang bonus jika ada
		bonusRepo := repository.NewBonusInventoryRepository(tx)
		for _, bonusReq := range req.BonusItems {
			if bonusReq.BonusInventoryID != uuid.Nil && bonusReq.Quantity > 0 {
				inv, err := bonusRepo.GetByID(tx.Statement.Context, bonusReq.BonusInventoryID)
				if err == nil && inv != nil {
					inv.Quantity -= bonusReq.Quantity
					_ = bonusRepo.Update(tx.Statement.Context, inv)
					_ = tx.Create(&domain.BonusInventoryLog{
						BonusInventoryID: inv.ID,
						Type:             domain.BonusInventoryOut,
						Quantity:         bonusReq.Quantity,
						Notes:            fmt.Sprintf("Bonus Penjualan Langsung (%s - %s)", trx.ReceiptNo, buyerName),
					})
				}
			}
		}

		// 1. Potong stok gudang barang jadi cabang terpilih (WarehouseStock & WarehouseLog)
		wRepo := repository.NewWarehouseRepository(tx)
		for _, item := range req.Items {
			remainingToDeduct := item.Quantity
			batches, err := wRepo.GetAvailableBatches(req.BranchID, item.ProductID, "GOOD")
			if err == nil && len(batches) > 0 {
				for i := range batches {
					if remainingToDeduct <= 0 {
						break
					}
					deductQty := batches[i].Quantity
					if deductQty > remainingToDeduct {
						deductQty = remainingToDeduct
					}
					batches[i].Quantity -= deductQty
					if err := wRepo.UpdateStock(&batches[i]); err != nil {
						return fmt.Errorf("gagal memperbarui stok gudang cabang: %w", err)
					}
					remainingToDeduct -= deductQty

					logSource := fmt.Sprintf("PENJUALAN_LANGSUNG: %s (%s)", trx.ReceiptNo, buyerName)
					if len(logSource) > 250 {
						logSource = logSource[:250]
					}
					_ = wRepo.CreateLog(&domain.WarehouseLog{
						BranchID:  req.BranchID,
						ProductID: item.ProductID,
						Type:      "OUT",
						Source:    logSource,
						BatchNo:   batches[i].BatchNo,
						Quantity:  deductQty,
					})
				}
			}

			if remainingToDeduct > 0 {
				defStock, err := wRepo.GetStock(req.BranchID, item.ProductID, "GOOD", "DEFAULT")
				if err != nil || defStock == nil {
					defStock = &domain.WarehouseStock{
						BranchID:  req.BranchID,
						ProductID: item.ProductID,
						StockType: "GOOD",
						BatchNo:   "DEFAULT",
						Quantity:  0,
					}
				}
				defStock.Quantity -= remainingToDeduct
				if err := wRepo.UpdateStock(defStock); err != nil {
					return fmt.Errorf("gagal memperbarui stok gudang cabang: %w", err)
				}

				logSource := fmt.Sprintf("PENJUALAN_LANGSUNG: %s (%s)", trx.ReceiptNo, buyerName)
				if len(logSource) > 250 {
					logSource = logSource[:250]
				}
				_ = wRepo.CreateLog(&domain.WarehouseLog{
					BranchID:  req.BranchID,
					ProductID: item.ProductID,
					Type:      "OUT",
					Source:    logSource,
					BatchNo:   "DEFAULT",
					Quantity:  remainingToDeduct,
				})
			}
		}

		// 2. Potong stok barang jadi pabrik jika ada pabrik untuk cabang ini atau untuk kanal penjualan langsung pabrik
		var targetFactory *domain.Factory
		var directFactory domain.Factory
		if err := tx.Where("branch_id = ?", req.BranchID).First(&directFactory).Error; err == nil && directFactory.ID != uuid.Nil {
			targetFactory = &directFactory
		} else {
			var allFactories []domain.Factory
			if err := tx.Where("company_id = ?", companyID).Find(&allFactories).Error; err == nil && len(allFactories) > 0 {
				for _, f := range allFactories {
					linkedBranchIDs, err := GetFactoryLinkedBranchIDs(tx, f.ID)
					if err == nil {
						for _, bID := range linkedBranchIDs {
							if bID == req.BranchID {
								factoryCopy := f
								targetFactory = &factoryCopy
								break
							}
						}
					}
					if targetFactory != nil {
						break
					}
				}
				if targetFactory == nil && salesChannel == domain.SalesChannelDirectFactory {
					factoryCopy := allFactories[0]
					targetFactory = &factoryCopy
				}
			}
		}

		if targetFactory != nil {
			fRepo := repository.NewFactoryRepository(tx)
			for _, item := range req.Items {
				remainingFactoryQty := item.Quantity
				fBatches, err := fRepo.GetAvailableBatches(targetFactory.ID, item.ProductID)
				if err == nil && len(fBatches) > 0 {
					for i := range fBatches {
						if remainingFactoryQty <= 0 {
							break
						}
						deductQty := fBatches[i].Quantity
						if deductQty > remainingFactoryQty {
							deductQty = remainingFactoryQty
						}
						fBatches[i].Quantity -= deductQty
						if err := fRepo.UpdateStock(&fBatches[i]); err != nil {
							return fmt.Errorf("gagal memperbarui stok pabrik: %w", err)
						}
						remainingFactoryQty -= deductQty

						logSource := fmt.Sprintf("PENJUALAN LANGSUNG KASIR (%s - %s)", trx.ReceiptNo, buyerName)
						if len(logSource) > 250 {
							logSource = logSource[:250]
						}
						_ = fRepo.CreateInventoryLog(&domain.FactoryInventoryLog{
							FactoryID: targetFactory.ID,
							ProductID: item.ProductID,
							Type:      "OUT",
							Source:    logSource,
							BatchNo:   fBatches[i].BatchNo,
							Quantity:  deductQty,
						})
					}
				}

				if remainingFactoryQty > 0 {
					defStock, err := fRepo.GetStock(targetFactory.ID, item.ProductID, "DEFAULT")
					if err != nil || defStock == nil {
						defStock = &domain.FactoryStock{
							FactoryID: targetFactory.ID,
							ProductID: item.ProductID,
							BatchNo:   "DEFAULT",
							Quantity:  0,
						}
					}
					defStock.Quantity -= remainingFactoryQty
					if err := fRepo.UpdateStock(defStock); err != nil {
						return fmt.Errorf("gagal memperbarui stok pabrik: %w", err)
					}

					logSource := fmt.Sprintf("PENJUALAN LANGSUNG KASIR (%s - %s)", trx.ReceiptNo, buyerName)
					if len(logSource) > 250 {
						logSource = logSource[:250]
					}
					_ = fRepo.CreateInventoryLog(&domain.FactoryInventoryLog{
						FactoryID: targetFactory.ID,
						ProductID: item.ProductID,
						Type:      "OUT",
						Source:    logSource,
						BatchNo:   "DEFAULT",
						Quantity:  remainingFactoryQty,
					})
				}
			}
		}

		// Catat SalesPayment PENDING untuk persetujuan kasir (jika CASH / TRANSFER / VA atau TEMPO dengan DP > 0)
		if (req.PaymentMethod == "CASH" || req.PaymentMethod == "TRANSFER" || req.PaymentMethod == "VA") && req.TotalAmount > 0 {
			paymentAmount := req.TotalAmount
			if req.PaidAmount > 0 {
				paymentAmount = req.PaidAmount
			}
			trxID := trx.ID
			salesPayment := &domain.SalesPayment{
				SalesTransactionID: &trxID,
				EmployeeID:         assignedEmployeeID,
				Amount:             paymentAmount,
				PaymentMethod:      req.PaymentMethod,
				PaymentStatus:      "PENDING",
				PaymentDate:        trxDate,
				Notes:              fmt.Sprintf("Penjualan Langsung Pabrik (%s - %s)", trx.ReceiptNo, buyerName),
			}
			if err := tx.Create(salesPayment).Error; err != nil {
				return fmt.Errorf("gagal membuat antrean persetujuan kasir: %w", err)
			}
		} else if !req.DeferPaymentToCashier && req.PaymentMethod == "TEMPO" && req.PaidAmount > 0 {
			trxID := trx.ID
			salesPayment := &domain.SalesPayment{
				SalesTransactionID: &trxID,
				EmployeeID:         assignedEmployeeID,
				Amount:             req.PaidAmount,
				PaymentMethod:      "CASH",
				PaymentStatus:      "PENDING",
				PaymentDate:        trxDate,
				Notes:              fmt.Sprintf("Uang Muka (DP) Penjualan Langsung Pabrik (%s - %s)", trx.ReceiptNo, buyerName),
			}
			if err := tx.Create(salesPayment).Error; err != nil {
				return fmt.Errorf("gagal membuat antrean persetujuan kasir: %w", err)
			}
		}

		createdTrx = trx
		return nil
	})

	if err != nil {
		return nil, err
	}

	fullTrx, err := u.salesRepo.FindByID(createdTrx.ID)
	if err == nil && fullTrx != nil {
		createdTrx = fullTrx
	}

	invalidateReportsCache(context.Background())
	return createdTrx, nil
}

func (u *salesUsecase) CalculateKPI(employeeID uuid.UUID, month, year int) error {
	kpiReports, err := u.performanceRepo.GetSalesKPIReportByMonth(month, year, nil)
	if err != nil {
		return err
	}

	// DEBUG: print all rows returned by KPI query for this period
	fmt.Printf("[KPI-DEBUG] CalculateKPI called for emp=%s period=%04d-%02d — query returned %d rows\n",
		employeeID.String()[:8], year, month, len(kpiReports))
	for _, r := range kpiReports {
		fmt.Printf("[KPI-DEBUG]   row: emp=%s totalOmzet=%.0f omzetLama=%.0f omzetBaru=%.0f\n",
			r.EmployeeID.String()[:8], r.TotalOmzet, r.OmzetLama, r.OmzetBaru)
	}

	var report *repository.SalesKPIReport
	for _, r := range kpiReports {
		if r.EmployeeID == employeeID {
			report = &r
			break
		}
	}

	if report == nil {
		fmt.Printf("[KPI-DEBUG] No report found for emp=%s, skipping\n", employeeID.String()[:8])
		return nil // No transactions found
	}

	totalOmzet := report.TotalOmzet
	fmt.Printf("[KPI-DEBUG] Setting AchievedOmzet=%.0f for emp=%s\n", totalOmzet, employeeID.String()[:8])

	salesKpi, err := u.performanceRepo.GetSalesKPIByEmployeeAndPeriod(employeeID, month, year)
	if err != nil {
		return err
	}

	if salesKpi == nil {
		// Create new if not exists
		salesKpi = &domain.SalesKPI{
			EmployeeID:      employeeID,
			TargetOmzet:     0, // Needs manual setup later
			TargetNewStores: 0,
			PeriodMonth:     month,
			PeriodYear:      year,
		}
	}

	fmt.Printf("[KPI-DEBUG] KPI before save: AchievedOmzet=%.0f → %.0f\n", salesKpi.AchievedOmzet, totalOmzet)
	salesKpi.AchievedOmzet = totalOmzet
	salesKpi.AchievedOmzetLama = report.OmzetLama
	salesKpi.AchievedOmzetBaru = report.OmzetBaru
	salesKpi.AchievedNewStores = report.TotalTokoBaru
	salesKpi.TotalVisits = report.TotalVisits

	return u.performanceRepo.SaveSalesKPI(salesKpi)
}

func (u *salesUsecase) GenerateKPIReportExcel(month, year int, branchIDs []uuid.UUID) ([]byte, error) {
	// 1. Fetch data from PerformanceRepository
	reports, err := u.performanceRepo.GetSalesKPIReportByMonth(month, year, branchIDs)
	if err != nil {
		return nil, err
	}

	// 2. Mocking Excel generation process
	// Usually this would use a library like `github.com/xuri/excelize/v2`
	// Here we simply return a mocked byte array mimicking CSV/Excel data
	csvData := "EmployeeID,OmzetLama,OmzetBaru,TotalTokoBaru\n"
	for _, r := range reports {
		csvData += r.EmployeeID.String() + "," +
			// Formatting float simply for demo
			"" + "," +
			"" + "," +
			"0" + "\n"
	}

	return []byte(csvData), nil
}

func (u *salesUsecase) GetSummaryKPI(periodType, dateStr string, month, year int, branchIDs []uuid.UUID, salesType string) (map[string]interface{}, error) {
	ctx := context.Background()
	branchStr := "ALL"
	if len(branchIDs) > 0 {
		var strParts []string
		for _, id := range branchIDs {
			strParts = append(strParts, id.String())
		}
		branchStr = strings.Join(strParts, ",")
	}
	cacheKey := fmt.Sprintf("reports:summary:type:%s:date:%s:month:%d:year:%d:branch:%s:sales:%s", periodType, dateStr, month, year, branchStr, salesType)

	if redisclient.Client != nil {
		if cached, err := redisclient.Client.Get(ctx, cacheKey).Result(); err == nil {
			var summary map[string]interface{}
			if json.Unmarshal([]byte(cached), &summary) == nil {
				return summary, nil
			}
		}
	}

	reports, err := u.performanceRepo.GetSalesKPIReportByPeriod(periodType, dateStr, month, year, branchIDs)
	if err != nil {
		return nil, err
	}

	var totalOmzetLama float64
	var totalOmzetBaru float64
	var totalTokoBaru int

	for _, r := range reports {
		totalOmzetLama += r.OmzetLama
		totalOmzetBaru += r.OmzetBaru
		totalTokoBaru += r.TotalTokoBaru
	}

	result := map[string]interface{}{
		"period_type":      periodType,
		"date":             dateStr,
		"period_month":     month,
		"period_year":      year,
		"total_omzet_lama": totalOmzetLama,
		"total_omzet_baru": totalOmzetBaru,
		"total_omzet_all":  totalOmzetLama + totalOmzetBaru,
		"total_toko_baru":  totalTokoBaru,
	}

	if redisclient.Client != nil {
		if bytes, err := json.Marshal(result); err == nil {
			_ = redisclient.Client.Set(ctx, cacheKey, bytes, 10*time.Minute).Err()
		}
	}

	return result, nil
}

func invalidateReportsCache(ctx context.Context) {
	if redisclient.Client == nil {
		return
	}

	// Jalankan secara asynchronous agar tidak memblokir response HTTP
	go func() {
		// Gunakan context baru yang terpisah dengan timeout
		bgCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var cursor uint64
		var iterKeys []string

		// Gunakan SCAN alih-alih KEYS untuk mencegah blokir pada Redis server
		for {
			var err error
			iterKeys, cursor, err = redisclient.Client.Scan(bgCtx, cursor, "reports:summary:*", 100).Result()
			if err != nil {
				fmt.Printf("DEBUG: Redis SCAN error: %v\n", err)
				break
			}

			if len(iterKeys) > 0 {
				_ = redisclient.Client.Del(bgCtx, iterKeys...).Err()
			}

			if cursor == 0 {
				break
			}
		}
	}()
}

func (u *salesUsecase) CreateTransaction(req CreateTransactionRequest) (*domain.SalesTransaction, error) {
	// Idempotency Check
	if req.ClientID != "" {
		// 1. Cek di tabel SalesTransaction (Reguler)
		var existingTrx domain.SalesTransaction
		if err := u.db.Where("client_id = ?", req.ClientID).First(&existingTrx).Error; err == nil {
			// Found duplicate request, return success immediately (Idempotent)
			return &existingTrx, nil
		}

		// 2. Cek di tabel SalesOrder (MT Mode)
		var existingSO domain.SalesOrder
		if err := u.db.Where("client_id = ?", req.ClientID).First(&existingSO).Error; err == nil {
			// Jika ini adalah retry dari transaksi MT, kembalikan fake transaction agar HP menganggapnya sukses
			return &domain.SalesTransaction{
				ID:              existingSO.ID,
				ReceiptNo:       existingSO.SONumber,
				TotalAmount:     existingSO.TotalAmount,
				TransactionDate: existingSO.OrderDate,
				Status:          string(existingSO.Status),
				PaymentMethod:   "SALES_ORDER",
				PaymentStatus:   existingSO.PaymentStatus,
				CompanyID:       existingSO.CompanyID,
			}, nil
		}
	}

	if req.ReceiptNo != "" {
		if !lockReceiptNo(req.ReceiptNo) {
			return nil, fmt.Errorf("duplicate: transaction with receipt_no %s is already being processed", req.ReceiptNo)
		}
		defer unlockReceiptNo(req.ReceiptNo)
	}

	// Offline backup tidak selalu memiliki employee_id (versi aplikasi lama).
	// Gunakan PIC toko sebagai fallback agar import dari web tetap dapat
	// memproses transaksi tanpa mengatribusikannya ke akun admin pengimpor.
	if req.EmployeeID == uuid.Nil && req.StoreID != uuid.Nil {
		var storeOwner struct{ AssignedEmployeeID *uuid.UUID }
		if err := u.db.Table("stores").Select("assigned_employee_id").Where("id = ?", req.StoreID).First(&storeOwner).Error; err == nil && storeOwner.AssignedEmployeeID != nil {
			req.EmployeeID = *storeOwner.AssignedEmployeeID
		}
	}

	// Pre-resolve Employee to check job title and branch
	emp, err := u.employeeRepo.FindByID(req.EmployeeID)
	if err != nil || emp == nil {
		if req.StoreID != uuid.Nil {
			var store domain.Store
			if errStore := u.db.Preload("AssignedEmployee").Where("id = ?", req.StoreID).First(&store).Error; errStore == nil && store.AssignedEmployee != nil {
				emp = store.AssignedEmployee
				req.EmployeeID = emp.ID
			}
		}
		if emp == nil {
			var fallbackEmp domain.Employee
			if req.CompanyID != uuid.Nil {
				if errEmp := u.db.Preload("JobPosition").Where("company_id = ?", req.CompanyID).Order("created_at ASC").First(&fallbackEmp).Error; errEmp == nil {
					emp = &fallbackEmp
					req.EmployeeID = emp.ID
				}
			}
			if emp == nil {
				if errEmp := u.db.Preload("JobPosition").Order("created_at ASC").First(&fallbackEmp).Error; errEmp == nil {
					emp = &fallbackEmp
					req.EmployeeID = emp.ID
				}
			}
		}
		if emp == nil {
			return nil, fmt.Errorf("failed to resolve employee: %v", err)
		}
	}

	// Determine effective salesman for omzet / KPI:
	// If the transaction is created by management (Director/Manager/Owner/Superadmin) on a store with an assigned salesman (PIC),
	// the order's EmployeeID is set to the store's assigned salesman so omzet & KPI credit goes to the PIC.
	effectiveSalesmanID := emp.ID
	if emp.JobPosition != nil {
		title := strings.ToLower(emp.JobPosition.Title)
		isManagementUser := strings.Contains(title, "direktur") || strings.Contains(title, "director") ||
			strings.Contains(title, "manager") || strings.Contains(title, "owner") ||
			strings.Contains(title, "superadmin") || strings.Contains(title, "spv") ||
			strings.Contains(title, "supervisor")

		if isManagementUser {
			var currentStore domain.Store
			if err := u.db.Select("id, assigned_employee_id").Where("id = ?", req.StoreID).First(&currentStore).Error; err == nil {
				if currentStore.AssignedEmployeeID != nil && *currentStore.AssignedEmployeeID != uuid.Nil {
					effectiveSalesmanID = *currentStore.AssignedEmployeeID
					fmt.Printf("[PROD-AUDIT] Management order created by %s credited to Store PIC %s\n", emp.ID, effectiveSalesmanID)
				}
			}
		}
	}

	if req.CompanyID == uuid.Nil {
		req.CompanyID = u.resolveCompanyID(nil, &req.EmployeeID, &req.StoreID)
	}

	// Force SALES_ORDER payment method if the employee is Sales TO (Task Order / Take Order).
	// PENTING: Pastikan "motoris" tidak ter-match oleh pengecekan "to" karena
	// kata "motoris" mengandung substring "to" (mo-TO-ris).
	// Sales Motoris menggunakan nota fisik dan harus langsung masuk ke Faktur Penjualan.
	// KHUSUS DISTRIBUTOR: Akun distributor fleksibel dapat melakukan mekanisme Sales TO maupun Motoris (Faktur Langsung)
	// terlepas dari nama jabatan resminya (Sales TO maupun Motoris), sehingga TO-guard dibypass jika isDistributor == true.
	isDistributor := emp.IsDistributor || (emp.DistributorStoreID != nil && *emp.DistributorStoreID != uuid.Nil) || req.IsDistributorOrder
	if isDistributor {
		req.IsDistributorOrder = true
		if req.DistributorStoreID == nil && emp.DistributorStoreID != nil && *emp.DistributorStoreID != uuid.Nil {
			req.DistributorStoreID = emp.DistributorStoreID
		}
	}

	if emp.JobPosition != nil {
		title := strings.ToLower(strings.TrimSpace(emp.JobPosition.Title))
		isMotoris := strings.Contains(title, "motoris")
		isTO := !isMotoris && (strings.Contains(title, "take order") ||
			strings.Contains(title, "task order") ||
			strings.Contains(title, "sales to") ||
			strings.HasSuffix(title, " to") ||
			strings.HasPrefix(title, "to ") ||
			title == "to")
		if isTO && !isDistributor {
			if req.PaymentMethod != "SALES_ORDER" {
				fmt.Printf("[TO-GUARD] Employee %s %s (Job: %s) request method '%s' forced to SALES_ORDER\n",
					emp.FirstName, emp.LastName, emp.JobPosition.Title, req.PaymentMethod)
			}
			req.PaymentMethod = "SALES_ORDER"
		}
	}

	if req.PaymentMethod == "SALES_ORDER" {
		var effectiveBranchID *uuid.UUID
		if emp.OperatingBranchID != nil && *emp.OperatingBranchID != uuid.Nil {
			effectiveBranchID = emp.OperatingBranchID
		} else if emp.BranchID != nil {
			effectiveBranchID = emp.BranchID
		}
		var effectiveCompanyID *uuid.UUID
		if emp.CompanyID != nil {
			effectiveCompanyID = emp.CompanyID
		}

		if effectiveBranchID == nil || effectiveCompanyID == nil {
			var store domain.Store
			if err := u.db.Preload("AssignedEmployee").Where("id = ?", req.StoreID).First(&store).Error; err == nil {
				if store.AssignedEmployee != nil {
					if effectiveBranchID == nil && store.AssignedEmployee.BranchID != nil {
						effectiveBranchID = store.AssignedEmployee.BranchID
					}
					if effectiveCompanyID == nil && store.AssignedEmployee.CompanyID != nil {
						effectiveCompanyID = store.AssignedEmployee.CompanyID
					}
				}
			}
		}

		if effectiveBranchID == nil {
			// Fallback 1: Ambil branch pertama yang aktif dari company karyawan/toko
			var fallbackBranch domain.Branch
			if effectiveCompanyID != nil && *effectiveCompanyID != uuid.Nil {
				if err := u.db.Where("company_id = ?", *effectiveCompanyID).Order("created_at ASC").First(&fallbackBranch).Error; err == nil {
					effectiveBranchID = &fallbackBranch.ID
				}
			}
			// Fallback 2: Ambil sembarang branch pertama yang ada di sistem jika belum ditemukan
			if effectiveBranchID == nil {
				if err := u.db.Order("created_at ASC").First(&fallbackBranch).Error; err == nil {
					effectiveBranchID = &fallbackBranch.ID
				}
			}
		}

		if effectiveBranchID == nil {
			return nil, errors.New("employee must be assigned to a branch or store must belong to a branch to create sales orders")
		}

		// Resolve VisitID to link this SO to the visit record
		var visitID *uuid.UUID
		// Use req.TransactionDate if provided, otherwise fallback to now
		targetDate := req.TransactionDate
		if targetDate.IsZero() {
			targetDate = time.Now()
		}

		// Pengecekan Duplikasi Ekstra: Cegah transaksi ganda (SO) jika data sama (store dan total amount) dan dalam waktu berdekatan (30 detik)
		recentTime := time.Now().Add(-30 * time.Second)
		var existingSO domain.SalesOrder
		if err := u.db.Where("employee_id = ? AND store_id = ? AND total_amount = ? AND created_at >= ?", req.EmployeeID, req.StoreID, req.TotalAmount, recentTime).First(&existingSO).Error; err == nil {
			return nil, fmt.Errorf("duplicate: sales order for store %s with same amount already exists recently", req.StoreID)
		}
		var visit *domain.SalesVisit
		if req.VisitID != nil {
			visit, _ = u.visitRepo.FindByID(*req.VisitID)
		} else {
			visit, _ = u.visitRepo.FindLatestByEmployeeStoreAndDate(req.EmployeeID, req.StoreID, targetDate)
		}
		if visit != nil {
			visitID = &visit.ID
			fmt.Printf("DEBUG: Found visit %v for SALES_ORDER transaction at %v\n", visit.ID, targetDate)
			now := time.Now()
			visit.CheckOutTime = &now
			visit.Status = "COMPLETED"
			if visit.Notes == "Deleted from schedule" {
				visit.Notes = "Kunjungan sales & order"
			}
			if visit.Type == "SKIP" {
				visit.Type = "CHECKIN"
			}
			if visit.CheckInTime.After(now) {
				visit.CheckInTime = now.Add(-15 * time.Minute)
			}
			_ = u.visitRepo.Update(visit)
		} else {
			fmt.Printf("DEBUG: No visit found for SALES_ORDER transaction at %v, store %v, emp %v. Proceeding without visit.\n", targetDate, req.StoreID, req.EmployeeID)
		}

		// 2. Map to CreateSORequest
		// Use BranchID and CompanyID from employee record for maximum reliability
		imgURL := req.PhysicalNotaImageURL
		if imgURL == "" {
			imgURL = req.ReceiptImageURL
		}

		soReq := CreateSORequest{
			BranchID:           *effectiveBranchID,
			CompanyID:          *effectiveCompanyID,
			EmployeeID:         effectiveSalesmanID,
			StoreID:            req.StoreID,
			StoreCategory:      req.StoreCategory,
			Notes:              req.Notes,
			VisitID:            visitID,
			Latitude:           req.Latitude,
			Longitude:          req.Longitude,
			PaymentMethod:      "SALES_ORDER",
			IsMT:               req.IsMT,
			FactoryCompanyID:   req.FactoryCompanyID,
			IsDistributorOrder: req.IsDistributorOrder,
			DistributorStoreID: req.DistributorStoreID,
			TotalAmount:        req.TotalAmount,
			OrderDate:          targetDate,
		}
		if req.ClientID != "" {
			soReq.ClientID = &req.ClientID
		}
		if imgURL != "" {
			soReq.IsPhysicalNota = true
			soReq.PhysicalNotaImageURL = imgURL
		}

		for _, it := range req.Items {
			soReq.Items = append(soReq.Items, SOItemRequest{
				ProductID:       it.ProductID,
				OrderedQuantity: it.OrderedQuantity,
				Unit:            it.Unit,
				PiecesPerUnit:   it.PiecesPerUnit,
				Price:           it.Price,
				DiscountPercent: it.DiscountPercent,
			})
		}

		for _, bit := range req.BonusItems {
			soReq.BonusItems = append(soReq.BonusItems, SOBonusItemRequest{
				BonusInventoryID: bit.BonusInventoryID,
				Quantity:         bit.Quantity,
				Notes:            bit.Notes,
			})
		}

		// 3. Call SO Usecase
		so, err := u.soUsecase.CreateSO(soReq)
		if err != nil {
			return nil, fmt.Errorf("failed to create sales order: %v", err)
		}

		if req.IsMT {
			// Find store details
			var store domain.Store
			storeName := ""
			if errStore := u.db.First(&store, "id = ?", req.StoreID).Error; errStore == nil {
				storeName = store.Name
			}

			// Auto-create / Link MT Purchase Order and MT Delivery Order to show in the web PO and MT Delivery views
			var ka domain.KeyAccount
			// Try 1: Match by store.OwnerName or store.Name with KeyAccount name
			errKA := u.db.Where("LOWER(name) = LOWER(?) OR LOWER(company_name) = LOWER(?)", store.OwnerName, store.OwnerName).First(&ka).Error
			if errKA != nil && store.Name != "" {
				errKA = u.db.Where("LOWER(name) LIKE LOWER(?) OR LOWER(company_name) LIKE LOWER(?)", "%"+store.Name+"%", "%"+store.Name+"%").First(&ka).Error
			}
			// Try 2: Match by company_id if factory company ID or transaction company ID matches
			if errKA != nil && req.FactoryCompanyID != nil && *req.FactoryCompanyID != uuid.Nil {
				errKA = u.db.Where("company_id = ?", *req.FactoryCompanyID).First(&ka).Error
			}
			if errKA != nil {
				errKA = u.db.Where("company_id = ?", req.CompanyID).First(&ka).Error
			}
			// Try 3: Any active key account
			if errKA != nil {
				errKA = u.db.Where("status = ?", "ACTIVE").First(&ka).Error
			}
			if errKA != nil {
				// Create default Key Account if none exists
				ka = domain.KeyAccount{
					ID:          uuid.New(),
					CompanyID:   req.CompanyID,
					Name:        "Ritel Modern Default",
					CompanyName: "PT Ritel Modern Default",
					Type:        "NATIONAL",
					LogoColor:   "bg-indigo-600",
					Status:      "ACTIVE",
				}
				_ = u.db.Create(&ka)
			}

			itemsOrdered := 0
			for _, it := range req.Items {
				itemsOrdered += it.OrderedQuantity
			}

			// Find employee name
			empName := ""
			if emp != nil {
				empName = emp.FirstName + " " + emp.LastName
			}

			po := domain.MTPurchaseOrder{
				ID:               uuid.New(),
				KeyAccountID:     ka.ID,
				PONumber:         so.SONumber, // Use SO Number as PO Number
				OrderDate:        so.OrderDate,
				DeliveryDeadline: so.OrderDate.AddDate(0, 0, 7), // default 7 days
				TotalAmount:      so.TotalAmount,
				ItemsOrdered:     itemsOrdered,
				ItemsShipped:     0,
				Status:           "PENDING",
				MdName:           empName,
				SupermarketName:  storeName,
			}
			_ = u.db.Create(&po)

			// Also auto-create MTDeliveryOrder so it appears in MT Delivery View
			var deliveryItems []domain.MTDeliveryOrderItem
			for _, it := range req.Items {
				unit := it.Unit
				if unit == "" {
					unit = "pcs"
				}
				deliveryItems = append(deliveryItems, domain.MTDeliveryOrderItem{
					ID:           uuid.New(),
					ProductID:    it.ProductID,
					RequestedQty: it.OrderedQuantity,
					Unit:         unit,
					BatchNo:      "DEFAULT",
					Notes:        "",
				})
			}

			now := time.Now()
			doNumber := fmt.Sprintf("MT-DO-%s-%04d", now.Format("20060102"), now.Unix()%10000)
			deliveryOrder := domain.MTDeliveryOrder{
				ID:             uuid.New(),
				CompanyID:      req.CompanyID,
				DONumber:       doNumber,
				KeyAccountID:   ka.ID,
				PONumber:       so.SONumber,
				SourceType:     domain.MTDeliverySourceBranch,
				SourceBranchID: effectiveBranchID,
				Status:         domain.MTDeliveryDraft,
				DCDestination:  storeName,
				Notes:          so.Notes,
				CreatedBy:      empName,
				SalesOrderID:   &so.ID,
				Items:          deliveryItems,
			}
			_ = u.db.Create(&deliveryOrder)
		}

		// 4. Return "fake" transaction to satisfy mobile app sync
		return &domain.SalesTransaction{
			ID:              so.ID,
			ReceiptNo:       so.SONumber,
			TotalAmount:     so.TotalAmount,
			TransactionDate: so.OrderDate,
			Status:          string(so.Status),
			PaymentMethod:   "SALES_ORDER",
			PaymentStatus:   so.PaymentStatus,
			CompanyID:       so.CompanyID,
			ReceiptImageURL: imgURL,
			CreatedAt:       so.CreatedAt,
		}, nil
	}

	paymentMethod := req.PaymentMethod
	if paymentMethod == "" {
		paymentMethod = "CASH"
	}

	dpPaymentMethod := req.DPPaymentMethod
	if paymentMethod == "TEMPO" && dpPaymentMethod == "" && req.PaidAmount > 0 {
		dpPaymentMethod = "CASH"
	}

	paymentStatus := string(domain.PaymentStatusUnpaid)
	isDigitalPayment := paymentMethod == "QRIS" || paymentMethod == "VA"
	isDigitalDP := paymentMethod == "TEMPO" && (dpPaymentMethod == "QRIS" || dpPaymentMethod == "VA")
	if !isDigitalPayment && !isDigitalDP {
		if paymentMethod != "VA" {
			if req.PaidAmount >= req.TotalAmount {
				paymentStatus = string(domain.PaymentStatusPaid)
			} else if req.PaidAmount > 0 {
				paymentStatus = string(domain.PaymentStatusPartial)
			}
		}
	}

	if req.ReceiptNo != "" {
		// Pengecekan Duplikasi Berdasarkan Receipt No
		var existing domain.SalesTransaction
		if err := u.db.Where("receipt_no = ? AND company_id = ?", req.ReceiptNo, req.CompanyID).First(&existing).Error; err == nil {
			// Append random suffix to make it unique and allow it to save
			randomSuffix := strings.ToUpper(uuid.New().String()[:5])
			req.ReceiptNo = fmt.Sprintf("%s-%s", req.ReceiptNo, randomSuffix)
		}
	} else {
		count, _ := u.salesRepo.CountByCompanyAndDate(req.CompanyID, req.TransactionDate)

		// Fetch company code
		comp, _ := u.companyRepo.FindByID(req.CompanyID)
		compCode := "WOW"
		if comp != nil && comp.Code != "" {
			compCode = comp.Code
		}

		req.ReceiptNo = fmt.Sprintf("INV-%s-%s-%04d", compCode, req.TransactionDate.Format("20060102"), count+1)
	}

	// Duplicate check by store and date removed to allow multiple transactions per store per day

	// Try to find actual visit to link it
	// Use req.TransactionDate if provided, otherwise fallback to now
	targetDate := req.TransactionDate
	if targetDate.IsZero() {
		targetDate = time.Now()
	}
	var visit *domain.SalesVisit
	if req.VisitID != nil {
		visit, _ = u.visitRepo.FindByID(*req.VisitID)
	} else {
		visit, _ = u.visitRepo.FindLatestByEmployeeStoreAndDate(req.EmployeeID, req.StoreID, targetDate)
	}
	var visitID *uuid.UUID
	if visit != nil {
		visitID = &visit.ID
		fmt.Printf("DEBUG: Found visit %v for transaction at %v\n", visit.ID, targetDate)
		now := time.Now()
		visit.CheckOutTime = &now
		visit.Status = "COMPLETED"
		if visit.Notes == "Deleted from schedule" {
			visit.Notes = "Kunjungan sales & order"
		}
		if visit.Type == "SKIP" {
			visit.Type = "CHECKIN"
		}
		if visit.CheckInTime.After(now) {
			visit.CheckInTime = now.Add(-15 * time.Minute)
		}
		_ = u.visitRepo.Update(visit)
	} else {
		fmt.Printf("DEBUG: No visit found for transaction at %v, store %v, emp %v. Proceeding without visit.\n", targetDate, req.StoreID, req.EmployeeID)
	}

	receiptNo := req.ReceiptNo
	if receiptNo == "" {
		// Format: WOW-[YYYYMMDD]-[RANDOM]
		now := time.Now()
		randomPart := uuid.New().String()[:4]
		receiptNo = fmt.Sprintf("WOW-%s-%s", now.Format("20060102"), randomPart)
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
		if req.StoreCategory == "" {
			req.StoreCategory = "TOKO_LAMA"
		}
	}

	var salesChannel string = domain.SalesChannelCanvas
	if emp.JobPosition != nil {
		title := strings.ToLower(emp.JobPosition.Title)
		if strings.Contains(title, "motoris") || strings.Contains(title, "booster") {
			salesChannel = domain.SalesChannelMotoris
		}
	}

	effectiveTrxBranchID := emp.BranchID
	if emp.OperatingBranchID != nil && *emp.OperatingBranchID != uuid.Nil {
		effectiveTrxBranchID = emp.OperatingBranchID
	}

	var storeIDPtr *uuid.UUID
	if req.StoreID != uuid.Nil && store.ID != uuid.Nil {
		storeIDPtr = &store.ID
	}

	trx := &domain.SalesTransaction{
		VisitID:              visitID,
		Latitude:             req.Latitude,
		Longitude:            req.Longitude,
		CompanyID:            req.CompanyID,
		BranchID:             effectiveTrxBranchID,
		SalesChannel:         salesChannel,
		StoreID:              storeIDPtr,
		EmployeeID:           effectiveSalesmanID,
		DistributorStoreID:   req.DistributorStoreID,
		IsDistributorOrder:   req.IsDistributorOrder,
		ReceiptNo:            receiptNo,
		ReceiptImageURL:      req.ReceiptImageURL,
		SignatureImageURL:    req.SignatureImageURL,
		TotalAmount:          req.TotalAmount,
		PaidAmount:           0,
		PaymentStatus:        paymentStatus,
		PaymentDueDate:       req.PaymentDueDate,
		StoreCategory:        req.StoreCategory,
		TransactionDate:      req.TransactionDate,
		PeriodMonth:          int(req.TransactionDate.Month()),
		PeriodYear:           req.TransactionDate.Year(),
		Status:               "PENDING",
		PaymentMethod:        paymentMethod,
		DPPaymentMethod:      dpPaymentMethod,
		CompanyBankAccountID: req.CompanyBankAccountID,
		Tenor: func() int {
			if req.Tenor <= 0 {
				return 1
			}
			return req.Tenor
		}(),
		Notes: &req.Notes,
	}
	if req.ClientID != "" {
		trx.ClientID = &req.ClientID
	}
	trx.PaidAmount = req.PaidAmount

	for _, itemReq := range req.Items {
		if itemReq.ProductID == uuid.Nil || itemReq.ProductID.String() == "00000000-0000-0000-0000-000000000000" || itemReq.Quantity <= 0 {
			continue
		}
		trx.Items = append(trx.Items, domain.SalesItem{
			ProductID:          itemReq.ProductID,
			Quantity:           itemReq.Quantity,
			PriceAtTransaction: itemReq.Price,
			DiscountPercent:    itemReq.DiscountPercent,
			Subtotal:           float64(itemReq.Quantity) * itemReq.Price * (1 - itemReq.DiscountPercent/100),
		})
	}

	for _, bonusReq := range req.BonusItems {
		trx.BonusItems = append(trx.BonusItems, domain.SalesTransactionBonusItem{
			BonusInventoryID: bonusReq.BonusInventoryID,
			Quantity:         bonusReq.Quantity,
			Notes:            bonusReq.Notes,
		})
	}

	// Generate installments if TEMPO
	if paymentMethod == "TEMPO" && trx.Tenor > 0 {
		remainingBalance := trx.TotalAmount - req.PaidAmount
		if remainingBalance > 0 {
			installmentAmount := remainingBalance / float64(trx.Tenor)
			// Format to 2 decimal places
			installmentAmount = float64(int(installmentAmount*100)) / 100.0

			for i := 1; i <= trx.Tenor; i++ {
				dueDate := req.TransactionDate.AddDate(0, 0, i*7)
				amount := installmentAmount
				if i == trx.Tenor {
					amount = remainingBalance - (installmentAmount * float64(trx.Tenor-1))
				}
				trx.Installments = append(trx.Installments, domain.SalesInstallment{
					TermNumber: i,
					DueDate:    dueDate,
					Amount:     amount,
					PaidAmount: 0,
					Status:     "UNPAID",
				})
			}
		}
	}

	err = u.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(trx).Error; err != nil {
			return err
		}

		// Deduct from Sales Stock (Bronjong)
		for _, itemReq := range req.Items {
			var stock domain.SalesStock
			if stockErr := tx.Where("employee_id = ? AND product_id = ?", req.EmployeeID, itemReq.ProductID).First(&stock).Error; stockErr == nil {
				stock.Quantity -= itemReq.Quantity
				stock.UpdatedAt = time.Now()
				_ = tx.Save(&stock).Error
			} else {
				fmt.Printf("DEBUG: Failed to get/update sales stock for employee %v, product %v: %v\n", req.EmployeeID, itemReq.ProductID, stockErr)
			}
		}

		for _, bonusReq := range req.BonusItems {
			tx.Model(&domain.SalesBonusStock{}).
				Where("employee_id = ? AND bonus_inventory_id = ?", req.EmployeeID, bonusReq.BonusInventoryID).
				UpdateColumn("quantity", gorm.Expr("quantity - ?", bonusReq.Quantity))
		}

		// Create initial payment record if paid
		if req.PaidAmount > 0 {
			initialPaymentStatus := "SUCCESS"
			actualPaymentMethod := paymentMethod
			if paymentMethod == "TEMPO" {
				actualPaymentMethod = dpPaymentMethod
			}
			if actualPaymentMethod == "VA" || actualPaymentMethod == "CASH" || actualPaymentMethod == "TRANSFER" {
				initialPaymentStatus = "PENDING"
			}
			notes := "Pembayaran awal saat transaksi"
			if paymentMethod == "TEMPO" {
				notes = fmt.Sprintf("Uang Muka (DP) Transaksi Tempo (%s - %s)", trx.ReceiptNo, actualPaymentMethod)
			}
			payment := &domain.SalesPayment{
				SalesTransactionID: &trx.ID,
				EmployeeID:         req.EmployeeID,
				Amount:             req.PaidAmount,
				PaymentMethod:      actualPaymentMethod,
				PaymentStatus:      initialPaymentStatus,
				PaymentDate:        req.TransactionDate,
				Notes:              notes,
			}
			if (actualPaymentMethod == "VA" || actualPaymentMethod == "TRANSFER") && req.ReceiptImageURL != "" {
				payment.TransferProofURL = &req.ReceiptImageURL
			}
			if err := tx.Create(payment).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Midtrans QRIS Generation
	if trx.PaymentMethod == "QRIS" && u.midtransClient != nil {
		resp, err := u.midtransClient.CreateQRIS(trx.ReceiptNo, int64(trx.TotalAmount))
		if err == nil && resp != nil {
			trx.MidtransID = resp.TransactionID
			qrisData := midtransPkg.ExtractQRISData(resp)
			if qrisData != "" {
				trx.MidtransQRISURL = qrisData
				_ = u.salesRepo.UpdateTransactionOnly(trx)
			}
		} else {
			fmt.Printf("DEBUG: Midtrans QRIS Error for %s: %v\n", trx.ReceiptNo, err)
		}
	}

	// Midtrans VA Generation - Skip if there is a manual transfer proof
	if trx.PaymentMethod == "VA" && req.ReceiptImageURL == "" && u.midtransClient != nil {
		resp, err := u.midtransClient.CreateBankTransfer(trx.ReceiptNo, int64(trx.TotalAmount), req.Bank)
		if err == nil && resp != nil {
			// Deep inspection of the response
			respJSON, _ := json.Marshal(resp)
			fmt.Printf("DEBUG MIDTRANS FULL Response: %s\n", string(respJSON))

			if len(resp.VaNumbers) > 0 {
				vaNum := resp.VaNumbers[0].VANumber
				bank := resp.VaNumbers[0].Bank
				trx.MidtransVANumber = &vaNum
				trx.MidtransBank = &bank
				trx.MidtransID = resp.TransactionID
				errUpdate := u.salesRepo.Update(trx)
				if errUpdate != nil {
					fmt.Printf("ERROR updating VA info to DB: %v\n", errUpdate)
				} else {
					fmt.Printf("SUCCESS: Saved VA Number %s to DB\n", *trx.MidtransVANumber)
				}
			} else if resp.BillKey != "" {
				// Mandiri Bill (E-Channel)
				fmt.Printf("DEBUG: Found Mandiri BillKey: %s, BillerCode: %s\n", resp.BillKey, resp.BillerCode)
				billKey := resp.BillKey
				billerCode := resp.BillerCode
				bankName := "mandiri"
				trx.MidtransBillKey = &billKey
				trx.MidtransBillerCode = &billerCode
				trx.MidtransBank = &bankName
				trx.MidtransID = resp.TransactionID
				errUpdate := u.salesRepo.Update(trx)
				if errUpdate != nil {
					fmt.Printf("ERROR updating Mandiri VA info to DB: %v\n", errUpdate)
				} else {
					fmt.Printf("SUCCESS: Saved Mandiri BillKey %s to DB\n", *trx.MidtransBillKey)
				}
			} else {
				fmt.Printf("DEBUG: Midtrans VA Response has no VA Numbers or Bill Key. Status: %s\n", resp.TransactionStatus)
			}
		} else {
			fmt.Printf("DEBUG: Midtrans VA Error for %s (%s): %v\n", trx.ReceiptNo, req.Bank, err)
		}
	}

	// Fetch Store info to include in response
	if trx.StoreID != nil {
		storeInfo, _ := u.storeRepo.FindByID(*trx.StoreID)
		if storeInfo != nil {
			trx.Store = storeInfo
		}
	}

	invalidateReportsCache(context.Background())
	return trx, nil
}

func (u *salesUsecase) SettleOmzetMT(id uuid.UUID, adminID *uuid.UUID) (*domain.SalesTransaction, error) {
	trx, err := u.salesRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if trx == nil {
		return nil, errors.New("transaksi tidak ditemukan")
	}

	trx.Status = "SETTLED"
	trx.PaymentStatus = string(domain.PaymentStatusPaid)
	now := time.Now()
	trx.VerifiedAt = &now
	trx.VerifiedByID = adminID

	if err := u.salesRepo.UpdateTransactionOnly(trx); err != nil {
		return nil, err
	}

	return trx, nil
}

func (u *salesUsecase) RecordPayment(req RecordPaymentRequest) error {
	trx, err := u.salesRepo.FindByID(req.TransactionID)
	if err != nil {
		return err
	}
	if trx == nil {
		return errors.New("transaksi tidak ditemukan")
	}
	if trx.Status == "CANCELLED" {
		return errors.New("faktur yang dibatalkan tidak dapat menerima pembayaran")
	}
	if req.CompanyBankAccountID != nil {
		var bank domain.CompanyBankAccount
		if err := u.db.First(&bank, "id = ? AND is_active = ?", *req.CompanyBankAccountID, true).Error; err != nil {
			return errors.New("rekening bank penerima tidak ditemukan atau tidak aktif")
		}
	}

	paymentMethod := "CASH"
	if req.PaymentMethod != "" {
		paymentMethod = req.PaymentMethod
	} else if trx.PaymentMethod != "" && trx.PaymentMethod != "TEMPO" {
		paymentMethod = trx.PaymentMethod
	}
	if paymentMethod != "TRANSFER" && req.CompanyBankAccountID != nil {
		return errors.New("rekening bank penerima hanya untuk pembayaran transfer")
	}

	paymentDate := req.PaymentDate
	if paymentDate.IsZero() {
		paymentDate = time.Now()
	}

	// For all payment methods, calculate new paid amount to reflect on the UI
	newPaidAmount := trx.PaidAmount + req.Amount
	if newPaidAmount > trx.TotalAmount {
		return errors.New("jumlah pembayaran melebihi total tagihan")
	}

	initialPaymentStatus := "SUCCESS"
	if paymentMethod == "VA" || paymentMethod == "CASH" || paymentMethod == "TRANSFER" {
		initialPaymentStatus = "PENDING"
	}

	// For CASH: guard against duplicate PENDING payments.
	// If an identical PENDING CASH payment already exists for this trx+amount,
	// skip creating another one to prevent double-approving.
	if paymentMethod == "CASH" {
		var existing domain.SalesPayment
		if err := u.salesRepo.FindPendingPaymentByTrxAndAmount(req.TransactionID, req.Amount, &existing); err == nil {
			// Already a PENDING CASH payment with same amount → idempotent, skip
			fmt.Printf("[RecordPayment] Skipping duplicate PENDING CASH payment trx=%s amount=%.0f\n",
				trx.ReceiptNo, req.Amount)
			return nil
		}
	}

	// Distribute payment across installments immediately to reflect on mobile app
	remainingPayment := req.Amount
	for i := range trx.Installments {
		inst := &trx.Installments[i]
		if inst.Status != "PAID" {
			unpaidAmount := inst.Amount - inst.PaidAmount
			if remainingPayment >= unpaidAmount {
				remainingPayment -= unpaidAmount
				inst.PaidAmount = inst.Amount
				inst.Status = "PAID"
				inst.UpdatedAt = time.Now()
				_ = u.salesRepo.UpdateInstallment(inst)
			} else if remainingPayment > 0 {
				inst.PaidAmount += remainingPayment
				remainingPayment = 0
				inst.UpdatedAt = time.Now()
				_ = u.salesRepo.UpdateInstallment(inst)
				break
			}
		}
	}

	// If payment exceeds predefined installments (flexible payments exceeding target tenor), record dynamic installment
	if remainingPayment > 0 {
		nextTerm := len(trx.Installments) + 1
		extraInst := domain.SalesInstallment{
			SalesTransactionID: trx.ID,
			TermNumber:         nextTerm,
			Amount:             remainingPayment,
			PaidAmount:         remainingPayment,
			DueDate:            paymentDate,
			Status:             "PAID",
		}
		_ = u.db.Create(&extraInst)
		trx.Installments = append(trx.Installments, extraInst)
	}

	var successPaidSum float64
	u.db.Model(&domain.SalesPayment{}).
		Where("sales_transaction_id = ? AND payment_status = 'SUCCESS'", trx.ID).
		Select("COALESCE(SUM(amount), 0)").Scan(&successPaidSum)

	if initialPaymentStatus == "SUCCESS" {
		successPaidSum += req.Amount
	}

	paymentStatus := string(domain.PaymentStatusUnpaid)
	if successPaidSum >= trx.TotalAmount {
		paymentStatus = string(domain.PaymentStatusPaid)
		trx.PaymentDueDate = nil
	} else if successPaidSum > 0 || trx.PaidAmount > 0 {
		paymentStatus = string(domain.PaymentStatusPartial)
		if req.NextDueDate != nil {
			trx.PaymentDueDate = req.NextDueDate
		}
	} else if req.NextDueDate != nil {
		trx.PaymentDueDate = req.NextDueDate
	}

	trx.PaidAmount = newPaidAmount
	trx.PaymentStatus = paymentStatus

	if err := u.salesRepo.Update(trx); err != nil {
		return err
	}

	paymentNotes := req.Notes
	if paymentNotes == "" {
		paymentNotes = fmt.Sprintf("Pembayaran Cicilan Tempo (%s - %s)", trx.ReceiptNo, paymentMethod)
	}

	payment := &domain.SalesPayment{
		SalesTransactionID:   &trx.ID,
		EmployeeID:           req.EmployeeID,
		Amount:               req.Amount,
		PaymentMethod:        paymentMethod,
		PaymentStatus:        initialPaymentStatus,
		PaymentDate:          paymentDate,
		CompanyBankAccountID: req.CompanyBankAccountID,
		Notes:                paymentNotes,
	}
	if req.TransferProofURL != "" {
		payment.TransferProofURL = &req.TransferProofURL
	}

	err = u.salesRepo.CreatePayment(payment)
	if err == nil {
		invalidateReportsCache(context.Background())
	}
	return err
}

func (u *salesUsecase) GetOutstandingTransactionsByStore(storeID uuid.UUID) ([]domain.SalesTransaction, error) {
	return u.salesRepo.GetOutstandingByStore(storeID)
}

func (u *salesUsecase) GetTransactionsByDueDate(date time.Time, employeeID uuid.UUID) ([]domain.SalesTransaction, error) {
	if date.IsZero() {
		return u.salesRepo.GetByDueDate("", employeeID)
	}
	dateStr := date.Format("2006-01-02")
	return u.salesRepo.GetByDueDate(dateStr, employeeID)
}

func (u *salesUsecase) GetPendingTransactions(employeeID uuid.UUID) ([]domain.SalesTransaction, error) {
	pending, err := u.salesRepo.FindPendingByEmployeeID(employeeID)
	if err != nil {
		return nil, err
	}

	// Include Sales Orders (Draft, Waiting Warehouse/Stock, Processing)
	orders, _ := u.soUsecase.GetSOByEmployee(employeeID)
	for _, so := range orders {
		if so.Status == domain.SOStatusDraft ||
			so.Status == domain.SOStatusWaitingWarehouse ||
			so.Status == domain.SOStatusWaitingStock ||
			so.Status == domain.SOStatusProcessing {
			pending = append(pending, u.mapSOToTransaction(so))
		}
	}

	return pending, nil
}

func (u *salesUsecase) GetHistoryTransactions(employeeID uuid.UUID, branchID *uuid.UUID) ([]domain.SalesTransaction, error) {
	all, err := u.salesRepo.FindAll()
	if err != nil {
		return nil, err
	}

	var history []domain.SalesTransaction
	for _, t := range all {
		// If branchID filter is given, filter by employee's branch
		if branchID != nil {
			if t.Employee.BranchID == nil || *t.Employee.BranchID != *branchID {
				continue
			}
		} else if employeeID != uuid.Nil {
			if t.EmployeeID != employeeID {
				continue
			}
		}
		history = append(history, t)
	}

	// Include Sales Orders
	orders, _ := u.soUsecase.GetSOByEmployee(employeeID)
	for _, so := range orders {
		history = append(history, u.mapSOToTransaction(so))
	}

	// Include Sales Payments
	payments, _ := u.salesRepo.FindPaymentsByEmployee(employeeID)
	for _, p := range payments {
		history = append(history, u.mapPaymentToTransaction(p))
	}

	// Sort by date descending
	sort.Slice(history, func(i, j int) bool {
		return history[i].CreatedAt.After(history[j].CreatedAt)
	})

	return history, nil
}

func (u *salesUsecase) mapSOToTransaction(so domain.SalesOrder) domain.SalesTransaction {
	var mappedItems []domain.SalesItem
	for _, item := range so.Items {
		mappedItems = append(mappedItems, domain.SalesItem{
			ID:                 item.ID,
			ProductID:          item.ProductID,
			Quantity:           item.Quantity,
			PriceAtTransaction: item.Price,
			DiscountPercent:    item.DiscountPercent,
			Subtotal:           item.Subtotal,
			Product:            item.Product,
		})
	}

	sigURL := ""
	if so.SignatureImageURL != nil {
		sigURL = *so.SignatureImageURL
	}

	return domain.SalesTransaction{
		ID:                so.ID,
		ReceiptNo:         so.SONumber,
		TotalAmount:       so.TotalAmount,
		TransactionDate:   so.OrderDate,
		Status:            string(so.Status),
		PaymentMethod:     "SALES_ORDER",
		CreatedAt:         so.CreatedAt,
		UpdatedAt:         so.UpdatedAt,
		EmployeeID:        so.EmployeeID,
		StoreID:           &so.StoreID,
		Store:             so.Store,
		CompanyID:         so.CompanyID,
		Company:           so.Company,
		Items:             mappedItems,
		Notes:             &so.Notes,
		PaymentStatus:     so.PaymentStatus,
		SignatureImageURL: sigURL,
	}
}

func (u *salesUsecase) mapPaymentToTransaction(payment domain.SalesPayment) domain.SalesTransaction {
	var trx domain.SalesTransaction

	if payment.SalesTransaction != nil {
		trx = *payment.SalesTransaction
	} else if payment.SalesOrder != nil {
		trx = u.mapSOToTransaction(*payment.SalesOrder)
	} else {
		// Fallback if no parent transaction
		return domain.SalesTransaction{
			ID:              payment.ID,
			ReceiptNo:       "PEMBAYARAN TEMPO",
			TotalAmount:     payment.Amount,
			TransactionDate: payment.PaymentDate,
			Status:          "LUNAS",
			PaymentMethod:   "TEMPO",
			CreatedAt:       payment.CreatedAt,
			UpdatedAt:       payment.UpdatedAt,
			EmployeeID:      payment.EmployeeID,
			IsPaymentLog:    true,
			Notes:           &payment.Notes,
		}
	}

	// Override specific fields for the history timeline
	trx.CreatedAt = payment.CreatedAt
	trx.UpdatedAt = payment.UpdatedAt
	trx.TransactionDate = payment.PaymentDate
	trx.IsPaymentLog = true

	// Force the status to appear as PAID/LUNAS for the timeline event
	trx.Status = "LUNAS"
	trx.PaymentStatus = "PAID"
	if payment.PaymentMethod != "" {
		trx.PaymentMethod = payment.PaymentMethod
	}

	return trx
}

func (u *salesUsecase) VerifyTransaction(id uuid.UUID, approverID *uuid.UUID, req VerifyTransactionRequest) error {
	trx, err := u.salesRepo.FindByID(id)
	if err != nil {
		return err
	}
	if trx == nil {
		return errors.New("transaction not found")
	}
	if trx.Status != "PENDING" {
		return errors.New("hanya faktur PENDING yang dapat diverifikasi")
	}

	now := time.Now()
	if req.ReceiptNo != "" {
		trx.ReceiptNo = req.ReceiptNo
	}
	if req.TotalAmount > 0 {
		trx.TotalAmount = req.TotalAmount

		// Sync pending cashier approval payment amount
		var pendingPayments []domain.SalesPayment
		if err := u.db.Where("(sales_transaction_id = ? OR sales_order_id IN (SELECT id FROM sales_orders WHERE invoice_id = ?)) AND payment_status = ?", id, id, "PENDING").Find(&pendingPayments).Error; err == nil {
			for _, p := range pendingPayments {
				updatedNote := p.Notes
				if !strings.Contains(updatedNote, "[ADJUSTED]") {
					if updatedNote != "" {
						updatedNote = "[ADJUSTED] " + updatedNote
					} else {
						updatedNote = "[ADJUSTED] Penyesuaian oleh Admin Nota"
					}
				}
				u.db.Model(&domain.SalesPayment{}).Where("id = ?", p.ID).Updates(map[string]interface{}{
					"amount": req.TotalAmount,
					"notes":  updatedNote,
				})
			}
		}
	}
	trx.Status = "VERIFIED"
	trx.VerifiedAt = &now
	if approverID != nil && *approverID != uuid.Nil {
		trx.VerifiedByID = approverID
	}
	if req.Notes != "" {
		trx.Notes = &req.Notes
	}

	err = u.salesRepo.UpdateTransactionOnly(trx)
	if err != nil {
		return err
	}

	// Recalculate KPI after verification
	_ = u.CalculateKPI(trx.EmployeeID, trx.PeriodMonth, trx.PeriodYear)

	invalidateReportsCache(context.Background())
	return nil
}

func (u *salesUsecase) VerifyPenerimaan(id uuid.UUID, adminNotaID uuid.UUID, req VerifyPenerimaanRequest) error {
	return u.db.Transaction(func(tx *gorm.DB) error {
		// 1. Get the SalesTransaction (Invoice)
		var trx domain.SalesTransaction
		if err := tx.Preload("Items").Preload("Items.Product").Where("id = ?", id).First(&trx).Error; err != nil {
			return fmt.Errorf("transaction not found: %v", err)
		}

		if trx.Status != "PENDING" {
			return errors.New("hanya faktur PENDING yang dapat diverifikasi")
		}

		// Find parent sales order if exists to check linked payments
		var so domain.SalesOrder
		var soID *uuid.UUID
		if err := tx.Select("id").Where("invoice_id = ?", trx.ID).First(&so).Error; err == nil {
			soID = &so.ID
		}

		if trx.PaymentStatus != string(domain.PaymentStatusPaid) {
			// Check if there are any PENDING cashier deposit approvals
			var pendingPaymentsCount int64
			q := tx.Model(&domain.SalesPayment{}).Where("payment_status = 'PENDING'")
			if soID != nil {
				q = q.Where("sales_transaction_id = ? OR sales_order_id = ?", trx.ID, *soID)
			} else {
				q = q.Where("sales_transaction_id = ?", trx.ID)
			}

			if err := q.Count(&pendingPaymentsCount).Error; err == nil && pendingPaymentsCount > 0 {
				return fmt.Errorf("faktur penjualan %s tidak bisa diverifikasi karena setoran uang belum disetujui oleh Kasir", trx.ReceiptNo)
			}
		}

		// Map to verify requested items
		reqItemMap := make(map[uuid.UUID]VerifyPenerimaanItemRequest)
		for _, it := range req.Items {
			reqItemMap[it.ProductID] = it
		}

		now := time.Now()
		var newTotalAmount float64

		// 2. Process each SalesItem only if items exist
		if len(trx.Items) > 0 {
			for i := range trx.Items {
				item := &trx.Items[i]
				reqItem, exists := reqItemMap[item.ProductID]

				originalQty := item.Quantity
				finalQty := originalQty // default to original if not provided/changed

				if exists {
					finalQty = reqItem.Quantity
				}

				// Calculate difference
				diff := originalQty - finalQty

				if diff != 0 {
					// Create adjustment log
					adjLog := &domain.SalesTransactionAdjustmentLog{
						SalesTransactionID: trx.ID,
						ProductID:          item.ProductID,
						OriginalQuantity:   originalQty,
						AdjustedQuantity:   finalQty,
						Difference:         diff,
						Notes:              reqItem.Notes,
						AdminNotaID:        adminNotaID,
						CreatedAt:          now,
					}
					if err := tx.Create(adjLog).Error; err != nil {
						return fmt.Errorf("failed to create adjustment log: %v", err)
					}

					// Update SalesItem
					item.Quantity = finalQty
					item.Subtotal = float64(finalQty) * item.PriceAtTransaction
					if err := tx.Save(item).Error; err != nil {
						return fmt.Errorf("failed to update sales item: %v", err)
					}

					// Adjust salesman/driver stock
					// A. Resolve stock employee (Driver if TO with Delivery, Salesman if direct Booster/Motoris)
					var stockEmployeeID uuid.UUID
					var parentSO domain.SalesOrder
					// Check if there is an associated SalesOrder pointing to this invoice
					if err := tx.Preload("DeliveryBatch").Where("invoice_id = ?", trx.ID).First(&parentSO).Error; err == nil {
						if parentSO.DeliveryBatch != nil && parentSO.DeliveryBatch.DriverID != nil {
							stockEmployeeID = *parentSO.DeliveryBatch.DriverID
						} else {
							stockEmployeeID = trx.EmployeeID
						}
					} else {
						stockEmployeeID = trx.EmployeeID
					}

					// B. Update stock
					if stockEmployeeID != uuid.Nil {
						salesTransferRepo := repository.NewSalesTransferRepository(tx)
						stock, stockErr := salesTransferRepo.GetSalesStock(stockEmployeeID, item.ProductID)
						if stockErr == nil && stock != nil {
							// diff > 0 means customer received LESS than originally calculated.
							// So the stock is returned back to the employee's vehicle/brojong.
							// diff < 0 means customer received MORE (very rare).
							stock.Quantity += diff
							stock.UpdatedAt = now
							if err := salesTransferRepo.UpdateSalesStock(stock); err != nil {
								return fmt.Errorf("failed to update sales stock: %v", err)
							}
						}
					}
				}

				newTotalAmount += item.Subtotal
			}

			// Recalculate transaction amount (minus discount)
			trx.TotalAmount = newTotalAmount
			// Make sure TotalAmount is non-negative
			if trx.TotalAmount < 0 {
				trx.TotalAmount = 0
			}
		}

		// Update transaction status & audit fields
		trx.Status = "VERIFIED"
		trx.VerifiedAt = &now
		trx.VerifiedByID = &adminNotaID
		if req.Notes != "" {
			if trx.Notes != nil && *trx.Notes != "" {
				updatedNotes := *trx.Notes + "\nVerifikasi Admin Nota: " + req.Notes
				trx.Notes = &updatedNotes
			} else {
				trx.Notes = &req.Notes
			}
		}

		if err := tx.Save(&trx).Error; err != nil {
			return fmt.Errorf("failed to verify transaction: %v", err)
		}

		if soID != nil {
			tx.Model(&domain.SalesOrder{}).Where("id = ?", *soID).Update("total_amount", trx.TotalAmount)
		}

		// Sinkronisasi antrean Persetujuan Kasir (SalesPayment PENDING)
		var pendingPayments []domain.SalesPayment
		pQuery := tx.Where("payment_status = ?", "PENDING")
		if soID != nil {
			pQuery = pQuery.Where("sales_transaction_id = ? OR sales_order_id = ?", trx.ID, *soID)
		} else {
			pQuery = pQuery.Where("sales_transaction_id = ?", trx.ID)
		}
		if err := pQuery.Find(&pendingPayments).Error; err == nil {
			for _, p := range pendingPayments {
				updatedNote := p.Notes
				if !strings.Contains(updatedNote, "[ADJUSTED]") {
					if updatedNote != "" {
						updatedNote = "[ADJUSTED] " + updatedNote
					} else {
						updatedNote = "[ADJUSTED] Penyesuaian oleh Admin Nota"
					}
				}
				tx.Model(&domain.SalesPayment{}).Where("id = ?", p.ID).Updates(map[string]interface{}{
					"amount": trx.TotalAmount,
					"notes":  updatedNote,
				})
			}
		}

		// Recalculate salesman KPI/omzet
		_ = u.CalculateKPI(trx.EmployeeID, trx.PeriodMonth, trx.PeriodYear)

		invalidateReportsCache(context.Background())
		return nil
	})
}

func (u *salesUsecase) GetPenerimaanTempoPayments(page, limit int, branchID string, status string, search string, salesman string, date string) ([]domain.SalesPayment, int64, float64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit

	query := u.db.Model(&domain.SalesPayment{}).
		Joins("JOIN sales_transactions st ON st.id = sales_payments.sales_transaction_id").
		Joins("LEFT JOIN stores s ON s.id = st.store_id").
		Joins("LEFT JOIN employees emp ON emp.id = COALESCE(st.employee_id, sales_payments.employee_id)").
		Where("sales_payments.payment_status = 'SUCCESS'").
		Where("(st.payment_method = 'TEMPO' OR sales_payments.notes ILIKE '%tempo%' OR sales_payments.notes ILIKE '%cicilan%')")

	if status != "" && status != "ALL" {
		query = query.Where("sales_payments.penerimaan_status = ?", status)
	} else if status == "" {
		query = query.Where("sales_payments.penerimaan_status = 'PENDING'")
	}

	if branchID != "" && branchID != "ALL_BRANCHES" && branchID != "ALL" {
		query = query.Where("COALESCE(emp.operating_branch_id, emp.branch_id, st.branch_id) = ?", branchID)
	}

	if search != "" {
		query = query.Where("st.receipt_no ILIKE ? OR s.name ILIKE ? OR emp.first_name ILIKE ? OR emp.last_name ILIKE ? OR CONCAT(emp.first_name, ' ', COALESCE(emp.last_name, '')) ILIKE ?",
			"%"+search+"%", "%"+search+"%", "%"+search+"%", "%"+search+"%", "%"+search+"%")
	}

	if salesman != "" {
		if empUUID, err := uuid.Parse(salesman); err == nil && empUUID != uuid.Nil {
			query = query.Where("COALESCE(st.employee_id, sales_payments.employee_id) = ?", empUUID)
		} else {
			query = query.Where("CONCAT(emp.first_name, ' ', COALESCE(emp.last_name, '')) ILIKE ?", "%"+salesman+"%")
		}
	}

	if date != "" {
		query = query.Where("DATE(sales_payments.payment_date) = ?", date)
	}

	var total int64
	var totalAmount float64

	// Count total rows
	if err := query.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, 0, 0, err
	}

	// Scan uses GORM's row callback, which leaves JOINs in its statement.
	// Keep it isolated so the following Find does not add the same JOINs twice.
	if err := query.Session(&gorm.Session{}).
		Select("COALESCE(SUM(sales_payments.amount), 0)").
		Scan(&totalAmount).Error; err != nil {
		return nil, 0, 0, err
	}

	var payments []domain.SalesPayment
	err := query.Session(&gorm.Session{}).Select("sales_payments.*").
		Preload("SalesTransaction").
		Preload("SalesTransaction.Store").
		Preload("SalesTransaction.Employee").
		Preload("SalesTransaction.Employee.JobPosition").
		Preload("SalesTransaction.Installments").
		Preload("Employee").
		Preload("Employee.JobPosition").
		Preload("VerifiedBy").
		Preload("PenerimaanVerifiedBy").
		Order("sales_payments.approved_at DESC NULLS LAST, sales_payments.created_at DESC").
		Offset(offset).Limit(limit).
		Find(&payments).Error

	return payments, total, totalAmount, err
}

func (u *salesUsecase) VerifyPenerimaanTempoPayment(paymentID, adminID uuid.UUID, notes string) error {
	return u.db.Transaction(func(tx *gorm.DB) error {
		var payment domain.SalesPayment
		if err := tx.Preload("SalesTransaction").Where("id = ?", paymentID).First(&payment).Error; err != nil {
			return fmt.Errorf("pembayaran tidak ditemukan: %v", err)
		}

		if payment.PaymentStatus != "SUCCESS" {
			return errors.New("hanya pembayaran yang telah disetujui Kasir yang dapat diverifikasi")
		}

		if payment.PenerimaanStatus == "VERIFIED" {
			return errors.New("pembayaran tempo ini sudah diverifikasi sebelumnya")
		}

		now := time.Now()
		payment.PenerimaanStatus = "VERIFIED"
		payment.PenerimaanVerifiedAt = &now
		if adminID != uuid.Nil {
			payment.PenerimaanVerifiedByID = &adminID
		}
		if notes != "" {
			payment.PenerimaanNotes = &notes
		}
		payment.UpdatedAt = now

		if err := tx.Save(&payment).Error; err != nil {
			return fmt.Errorf("gagal memverifikasi penerimaan pembayaran tempo: %v", err)
		}

		// Recalculate salesman KPI and Omzet
		var salesmanID uuid.UUID
		var periodMonth, periodYear int

		if payment.SalesTransaction != nil {
			salesmanID = payment.SalesTransaction.EmployeeID
			periodMonth = payment.SalesTransaction.PeriodMonth
			periodYear = payment.SalesTransaction.PeriodYear
		} else if payment.EmployeeID != uuid.Nil {
			salesmanID = payment.EmployeeID
		}

		if periodMonth == 0 || periodYear == 0 {
			periodMonth = int(payment.PaymentDate.Month())
			periodYear = payment.PaymentDate.Year()
		}

		if salesmanID != uuid.Nil {
			_ = u.CalculateKPI(salesmanID, periodMonth, periodYear)
			// Juga update periode payment date jika berbeda dengan periode nota
			payMonth := int(payment.PaymentDate.Month())
			payYear := payment.PaymentDate.Year()
			if payMonth != periodMonth || payYear != periodYear {
				_ = u.CalculateKPI(salesmanID, payMonth, payYear)
			}
		}

		invalidateReportsCache(context.Background())
		return nil
	})
}

func (u *salesUsecase) RedeliverTransaction(id uuid.UUID, adminNotaID uuid.UUID, reason string) error {
	return u.db.Transaction(func(tx *gorm.DB) error {
		// 1. Get the SalesTransaction (Invoice)
		var trx domain.SalesTransaction
		if err := tx.Preload("Items").Preload("Items.Product").Where("id = ?", id).First(&trx).Error; err != nil {
			return fmt.Errorf("transaction not found: %v", err)
		}

		if trx.Status != "PENDING" {
			return errors.New("hanya faktur PENDING yang dapat dikembalikan ke surat jalan")
		}
		var failedItems int64
		if err := tx.Model(&domain.DeliveryItem{}).
			Where("sales_transaction_id = ? AND status = ?", id, domain.DeliveryItemFailed).
			Count(&failedItems).Error; err != nil {
			return err
		}
		if failedItems == 0 {
			return errors.New("hanya faktur dengan pengiriman gagal yang dapat dikirim ulang")
		}

		// 2. Find the associated SalesOrder
		var so domain.SalesOrder
		if err := tx.Preload("Items").Where("invoice_id = ?", trx.ID).First(&so).Error; err != nil {
			return fmt.Errorf("sales order terkait tidak ditemukan: %v", err)
		}

		// 3. Restore warehouse stock
		// When the driver started delivery, the items were deducted from the branch's GOOD stock.
		// Since we are resetting the SO to be delivered again, we must return these items back to the branch's GOOD stock.
		wRepo := repository.NewWarehouseRepository(tx)
		for _, item := range so.Items {
			// Find the batch allocation for this SO item if exists, otherwise DEFAULT
			batchNo := "DEFAULT"
			var soiBatch domain.SalesOrderItemBatch
			if err := tx.Where("sales_order_item_id = ?", item.ID).First(&soiBatch).Error; err == nil && soiBatch.BatchNo != "" {
				batchNo = soiBatch.BatchNo
			}

			// Get GOOD stock for this product in the branch
			stock, err := wRepo.GetStock(so.BranchID, item.ProductID, "GOOD", batchNo)
			if err != nil {
				// If stock doesn't exist, create it
				stock = &domain.WarehouseStock{
					BranchID:  so.BranchID,
					ProductID: item.ProductID,
					StockType: "GOOD",
					BatchNo:   batchNo,
					Quantity:  0,
				}
				if err := tx.Create(stock).Error; err != nil {
					return fmt.Errorf("failed to create warehouse stock: %v", err)
				}
			}

			// Add the quantity back to GOOD stock
			qtyToRestore := item.OrderedQuantity
			if qtyToRestore <= 0 {
				qtyToRestore = item.Quantity
			}
			stock.Quantity += qtyToRestore
			if err := wRepo.UpdateStock(stock); err != nil {
				return fmt.Errorf("failed to restore warehouse stock: %v", err)
			}

			// Log warehouse mutation
			wRepo.CreateLog(&domain.WarehouseLog{
				BranchID:  so.BranchID,
				ProductID: item.ProductID,
				Type:      "IN",
				Source:    fmt.Sprintf("BATAL_KIRIM: %s (Redeliver)", so.SONumber),
				Quantity:  qtyToRestore,
				BatchNo:   batchNo,
			})
		}

		// 4. Delete the associated SalesReturn (if exists)
		if err := tx.Where("sales_order_id = ? AND transaction_id = ?", so.ID, trx.ID).Delete(&domain.SalesReturn{}).Error; err != nil {
			fmt.Printf("DEBUG Redeliver: failed to delete sales return or none existed: %v\n", err)
		}

		// 5. Delete/Disconnect DeliveryItem
		if err := tx.Where("sales_order_id = ?", so.ID).Delete(&domain.DeliveryItem{}).Error; err != nil {
			fmt.Printf("DEBUG Redeliver: failed to delete delivery item: %v\n", err)
		}

		// 6. Reset SalesOrder fields so it can be batched again
		so.Status = domain.SOStatusConfirmed // Back to CONFIRMED (ready for delivery batch)
		if reason != "" {
			so.Notes = strings.TrimSpace(so.Notes + "\nGagal kirim: " + reason)
		}
		so.DeliveryBatchID = nil
		so.DeliveryOrderNo = nil
		// so.InvoiceID = nil // [REMOVED] DO NOT delete the invoice link, reuse it
		so.ShippedAt = nil
		so.ReceivedAt = nil
		so.ReceivedBy = ""
		so.PODImageURL = nil
		so.SignatureImageURL = nil

		// Reset actual quantities and restore original subtotals on SO items
		var originalTotalAmount float64
		for i := range so.Items {
			so.Items[i].ActualQuantity = 0
			ppu := so.Items[i].PiecesPerUnit
			if ppu <= 0 {
				ppu = 1
			}
			qty := so.Items[i].OrderedQuantity * ppu
			if qty <= 0 {
				qty = so.Items[i].Quantity
			}
			so.Items[i].Subtotal = float64(qty) * so.Items[i].Price * (1.0 - so.Items[i].DiscountPercent/100.0)
			originalTotalAmount += so.Items[i].Subtotal

			if err := tx.Save(&so.Items[i]).Error; err != nil {
				return fmt.Errorf("failed to reset SO item: %v", err)
			}
		}

		so.TotalAmount = originalTotalAmount - so.DiscountAmount
		if so.TotalAmount < 0 {
			so.TotalAmount = 0
		}

		if err := tx.Save(&so).Error; err != nil {
			return fmt.Errorf("failed to reset sales order: %v", err)
		}

		// Hapus antrean setoran kasir PENDING karena pengiriman dibatalkan/akan dikirim ulang
		tx.Where("(sales_order_id = ? OR sales_transaction_id = ?) AND payment_status = ?", so.ID, trx.ID, "PENDING").Delete(&domain.SalesPayment{})

		// 7. [REMOVED] Do not delete SalesTransaction, reuse the old one
		// if err := tx.Delete(&trx).Error; err != nil {
		// 	return fmt.Errorf("failed to delete sales transaction: %v", err)
		// }

		invalidateReportsCache(context.Background())
		return nil
	})
}

// CancelFailedDeliveryTransaction closes a failed delivery without scheduling it again.
// The physical return transfer remains pending until warehouse staff receive the goods.
func (u *salesUsecase) CancelFailedDeliveryTransaction(id uuid.UUID, reason string) error {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return errors.New("alasan pembatalan wajib diisi")
	}

	var employeeID uuid.UUID
	var periodMonth, periodYear int
	err := u.db.Transaction(func(tx *gorm.DB) error {
		var trx domain.SalesTransaction
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&trx, "id = ?", id).Error; err != nil {
			return errors.New("faktur penjualan tidak ditemukan")
		}
		if trx.Status != "PENDING" || trx.VerifiedAt != nil {
			return errors.New("hanya faktur PENDING yang belum diverifikasi dapat dibatalkan")
		}
		if trx.PaidAmount > 0 || (trx.PaymentStatus != "" && trx.PaymentStatus != string(domain.PaymentStatusUnpaid)) {
			return errors.New("faktur dengan pembayaran tercatat tidak dapat dibatalkan")
		}

		var so domain.SalesOrder
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("invoice_id = ?", id).First(&so).Error; err != nil {
			return errors.New("SO terkait faktur tidak ditemukan")
		}
		if so.Status != domain.SOStatusCancelled {
			return errors.New("hanya faktur dari SO gagal kirim yang dapat dibatalkan")
		}
		var failedItems int64
		if err := tx.Model(&domain.DeliveryItem{}).
			Where("sales_order_id = ? AND sales_transaction_id = ? AND status = ?", so.ID, id, domain.DeliveryItemFailed).
			Count(&failedItems).Error; err != nil {
			return err
		}
		if failedItems == 0 {
			return errors.New("pengiriman gagal untuk faktur ini tidak ditemukan")
		}

		var settledPayments int64
		if err := tx.Model(&domain.SalesPayment{}).
			Where("(sales_order_id = ? OR sales_transaction_id = ?) AND payment_status <> ?", so.ID, id, "PENDING").
			Count(&settledPayments).Error; err != nil {
			return err
		}
		if settledPayments > 0 {
			return errors.New("faktur dengan pembayaran tercatat tidak dapat dibatalkan")
		}
		if err := tx.Where("(sales_order_id = ? OR sales_transaction_id = ?) AND payment_status = ?", so.ID, id, "PENDING").
			Delete(&domain.SalesPayment{}).Error; err != nil {
			return err
		}

		invoiceNotes := "Dibatalkan karena gagal kirim: " + reason
		if trx.Notes != nil && strings.TrimSpace(*trx.Notes) != "" {
			invoiceNotes = strings.TrimSpace(*trx.Notes) + "\n" + invoiceNotes
		}
		if err := tx.Model(&trx).Updates(map[string]interface{}{"status": "CANCELLED", "notes": invoiceNotes}).Error; err != nil {
			return err
		}
		if err := tx.Model(&so).Update("notes", strings.TrimSpace(so.Notes+"\nFaktur dibatalkan karena gagal kirim: "+reason)).Error; err != nil {
			return err
		}
		employeeID, periodMonth, periodYear = trx.EmployeeID, trx.PeriodMonth, trx.PeriodYear
		return nil
	})
	if err != nil {
		return err
	}
	_ = u.CalculateKPI(employeeID, periodMonth, periodYear)
	invalidateReportsCache(context.Background())
	return nil
}

func (u *salesUsecase) GetPerformanceList(month, year int, branchIDs []uuid.UUID) ([]domain.SalesKPI, error) {
	return u.performanceRepo.GetSalesKPIsByMonth(month, year, branchIDs)
}

func (u *salesUsecase) GetYearlyPerformanceList(year int, branchIDs []uuid.UUID) ([]domain.SalesKPI, error) {
	return u.performanceRepo.GetSalesKPIsByYear(year, branchIDs)
}

func (u *salesUsecase) GetDailyPerformanceList(date time.Time, branchIDs []uuid.UUID) ([]domain.SalesKPI, error) {
	return u.performanceRepo.GetSalesDailyKPIs(date, branchIDs)
}

func (u *salesUsecase) GetPerformanceByDateRange(startDate, endDate time.Time, branchIDs []uuid.UUID) ([]domain.SalesKPI, error) {
	return u.performanceRepo.GetSalesKPIsByDateRange(startDate, endDate, branchIDs)
}

func (u *salesUsecase) GetTransactionByID(id uuid.UUID) (*domain.SalesTransaction, error) {
	return u.salesRepo.FindByID(id)
}

func (u *salesUsecase) GetAllTransactionsPaginated(page, limit int, companyID, status, search, salesman, month, year, date, salesType, branchID, paymentMethod, paymentStatus string, targetEntity ...string) ([]domain.SalesTransaction, int64, float64, error) {
	return u.salesRepo.FindAllPaginated(page, limit, companyID, status, search, salesman, month, year, date, salesType, branchID, paymentMethod, paymentStatus, targetEntity...)
}

// GetAllOutstandingInvoices returns all invoices with TEMPO payment method that are not yet fully paid.
// Used by Admin Nota to pick invoices for Surat Jalan collection.
func (u *salesUsecase) GetAllOutstandingInvoices(companyID, branchID, salesType string) ([]domain.SalesTransaction, error) {
	query := u.db.Model(&domain.SalesTransaction{}).
		Preload("Employee").Preload("Employee.JobPosition").Preload("Store").
		Preload("Payments").Preload("Installments").
		Where("payment_method = 'TEMPO' AND payment_status IN ('UNPAID','PARTIAL') AND status != 'CANCELLED'")

	if companyID != "" {
		query = query.Where("company_id = ?", companyID)
	}
	if branchID != "" && branchID != "ALL_BRANCHES" && branchID != "ALL" {
		query = query.Joins("LEFT JOIN employees emp_oi ON emp_oi.id = sales_transactions.employee_id").
			Where("emp_oi.branch_id = ?", branchID)
	}
	if salesType != "" && salesType != "ALL" {
		query = query.Joins("LEFT JOIN employees emp_st ON emp_st.id = sales_transactions.employee_id").
			Joins("LEFT JOIN job_positions jp_st ON jp_st.id = emp_st.job_position_id").
			Where("jp_st.title ILIKE ?", "%"+salesType+"%")
	}

	var transactions []domain.SalesTransaction
	if err := query.Order("payment_due_date ASC").Find(&transactions).Error; err != nil {
		return nil, err
	}
	return transactions, nil
}

func (u *salesUsecase) ExportTransactionsExcel(companyID, status, search, salesman, month, year, date, salesType, branchID, paymentMethod, paymentStatus string, targetEntity ...string) ([]byte, error) {
	// Ambil semua data sesuai filter (limit = -1)
	transactions, _, _, err := u.salesRepo.FindAllPaginated(1, -1, companyID, status, search, salesman, month, year, date, salesType, branchID, paymentMethod, paymentStatus, targetEntity...)
	if err != nil {
		return nil, err
	}

	f := excelize.NewFile()
	defer f.Close()
	sheet := "Sheet1"

	// Headers
	headers := []string{"Tanggal", "No. Nota", "Nama Toko", "Kategori Toko", "Salesman", "Metode Pembayaran", "Status Pembayaran", "Jatuh Tempo", "Total Tagihan", "Total Terbayar", "Sisa Tagihan", "Status Faktur"}
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, header)
	}

	// Data rows
	for i, trx := range transactions {
		row := i + 2
		dateStr := trx.TransactionDate.Format("2006-01-02 15:04:05")
		storeName := ""
		if trx.Store != nil {
			storeName = trx.Store.Name
		}
		salesmanName := ""
		if trx.Employee != nil {
			salesmanName = trx.Employee.FirstName + " " + trx.Employee.LastName
		}
		dueDateStr := "-"
		if trx.PaymentDueDate != nil {
			dueDateStr = trx.PaymentDueDate.Format("2006-01-02")
		}
		sisaTagihan := trx.TotalAmount - trx.PaidAmount
		if sisaTagihan < 0 {
			sisaTagihan = 0
		}

		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), dateStr)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", row), trx.ReceiptNo)
		f.SetCellValue(sheet, fmt.Sprintf("C%d", row), storeName)
		f.SetCellValue(sheet, fmt.Sprintf("D%d", row), trx.StoreCategory)
		f.SetCellValue(sheet, fmt.Sprintf("E%d", row), salesmanName)
		f.SetCellValue(sheet, fmt.Sprintf("F%d", row), trx.PaymentMethod)
		f.SetCellValue(sheet, fmt.Sprintf("G%d", row), trx.PaymentStatus)
		f.SetCellValue(sheet, fmt.Sprintf("H%d", row), dueDateStr)
		f.SetCellValue(sheet, fmt.Sprintf("I%d", row), trx.TotalAmount)
		f.SetCellValue(sheet, fmt.Sprintf("J%d", row), trx.PaidAmount)
		f.SetCellValue(sheet, fmt.Sprintf("K%d", row), sisaTagihan)
		f.SetCellValue(sheet, fmt.Sprintf("L%d", row), trx.Status)
	}

	var b bytes.Buffer
	if err := f.Write(&b); err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

func (u *salesUsecase) UpdateTransaction(id uuid.UUID, req ManualEntryRequest) error {
	trx, err := u.salesRepo.FindByID(id)
	if err != nil {
		return err
	}
	if trx == nil {
		return errors.New("transaction not found")
	}

	// Check if this transaction or its parent SalesOrder is linked to a delivery batch that is actively on the road (ON_DELIVERY)
	var count int64
	dbErr := u.db.Table("delivery_items").
		Joins("JOIN delivery_batches ON delivery_batches.id = delivery_items.delivery_batch_id").
		Joins("LEFT JOIN sales_orders ON sales_orders.id = delivery_items.sales_order_id").
		Where("(delivery_items.sales_transaction_id = ? OR sales_orders.invoice_id = ?) AND delivery_batches.status = ?", id, id, "ON_DELIVERY").
		Count(&count).Error
	if dbErr == nil && count > 0 {
		return errors.New("tidak bisa mengubah faktur yang ada di dalam surat jalan yang sedang dalam proses pengiriman (ON_DELIVERY)")
	}

	trxDate, err := time.Parse("2006-01-02", req.TransactionDate)
	if err != nil {
		return errors.New("invalid transaction date format, expected YYYY-MM-DD")
	}

	// Store old values for KPI recalculation if period or employee changes
	oldEmployeeID := trx.EmployeeID
	oldMonth := trx.PeriodMonth
	oldYear := trx.PeriodYear
	oldTotalAmount := trx.TotalAmount

	var employeeID uuid.UUID
	if req.EmployeeID != nil {
		employeeID = *req.EmployeeID
	}

	trx.StoreID = req.StoreID
	trx.EmployeeID = employeeID
	resolvedCompanyID := u.resolveCompanyID(req.CompanyID, req.EmployeeID, req.StoreID)
	if resolvedCompanyID != uuid.Nil {
		trx.CompanyID = resolvedCompanyID
	} else if trx.CompanyID == uuid.Nil {
		trx.CompanyID = u.resolveCompanyID(nil, &trx.EmployeeID, trx.StoreID)
	}
	trx.ReceiptNo = req.ReceiptNo
	trx.ReceiptImageURL = req.ReceiptImageURL
	trx.TotalAmount = req.TotalAmount
	trx.DiscountAmount = req.DiscountAmount
	trx.StoreCategory = req.StoreCategory
	trx.TransactionDate = trxDate
	trx.PeriodMonth = int(trxDate.Month())
	trx.PeriodYear = trxDate.Year()
	if req.PaymentMethod != "" {
		trx.PaymentMethod = req.PaymentMethod
	}
	if req.PaymentDueDate != nil {
		trx.PaymentDueDate = req.PaymentDueDate
	}
	if req.Tenor > 0 {
		trx.Tenor = req.Tenor
	}
	if req.PaidAmount >= 0 {
		trx.PaidAmount = req.PaidAmount
	}
	if req.PaymentStatus != "" {
		trx.PaymentStatus = req.PaymentStatus
	}
	if req.DeferPaymentToCashier && trx.TotalAmount != oldTotalAmount {
		// Admin Nota may correct the invoice total, but only the cashier changes
		// recorded payments. Derive the status from the preserved paid amount.
		switch {
		case trx.PaidAmount <= 0:
			trx.PaymentStatus = string(domain.PaymentStatusUnpaid)
		case trx.PaidAmount >= trx.TotalAmount:
			trx.PaymentStatus = string(domain.PaymentStatusPaid)
		default:
			trx.PaymentStatus = string(domain.PaymentStatusPartial)
		}
	}
	trx.Notes = &req.Notes
	if trx.ClientID != nil && *trx.ClientID == "" {
		trx.ClientID = nil
	}
	if req.ClientID != "" {
		trx.ClientID = &req.ClientID
	}

	// Update items
	trx.Items = []domain.SalesItem{} // Clear existing slice in memory
	for _, itemReq := range req.Items {
		trx.Items = append(trx.Items, domain.SalesItem{
			SalesTransactionID: trx.ID,
			ProductID:          itemReq.ProductID,
			Quantity:           itemReq.Quantity,
			PriceAtTransaction: itemReq.Price,
			DiscountPercent:    itemReq.DiscountPercent,
			Subtotal:           float64(itemReq.Quantity) * itemReq.Price * (1 - itemReq.DiscountPercent/100),
		})
	}

	// Delete old bonus items
	u.db.Where("sales_transaction_id = ?", trx.ID).Delete(&domain.SalesTransactionBonusItem{})
	trx.BonusItems = nil

	for _, bonusReq := range req.BonusItems {
		trx.BonusItems = append(trx.BonusItems, domain.SalesTransactionBonusItem{
			SalesTransactionID: trx.ID,
			BonusInventoryID:   bonusReq.BonusInventoryID,
			Quantity:           bonusReq.Quantity,
			Notes:              bonusReq.Notes,
		})
	}

	err = u.salesRepo.Update(trx)
	if err != nil {
		return err
	}

	// Sinkronisasi retur barang sampingan (wadah/jerigen/botol bekas) dari catatan nota ke stok gudang & sales_returns
	if req.Notes != "" {
		var processedName string
		if req.EmployeeID != nil {
			var emp domain.Employee
			if err := u.db.Where("id = ?", *req.EmployeeID).First(&emp).Error; err == nil {
				processedName = strings.TrimSpace(emp.FirstName + " " + emp.LastName)
			}
		}
		if updatedNotes, errSync := u.syncNotaReturnItems(u.db, trx, req.Notes, processedName); errSync == nil && updatedNotes != "" {
			trx.Notes = &updatedNotes
			_ = u.salesRepo.UpdateTransactionOnly(trx)
		}
	}

	// Update installments if TEMPO and no verified payments yet
	if trx.PaymentMethod == "TEMPO" && trx.Tenor > 0 {
		var paymentCount int64
		u.db.Model(&domain.SalesPayment{}).Where("sales_transaction_id = ? AND payment_status = 'SUCCESS'", trx.ID).Count(&paymentCount)
		if paymentCount == 0 {
			u.db.Where("sales_transaction_id = ?", trx.ID).Delete(&domain.SalesInstallment{})
			remainingBalance := trx.TotalAmount - trx.PaidAmount
			if remainingBalance > 0 {
				installmentAmount := remainingBalance / float64(trx.Tenor)
				installmentAmount = float64(int(installmentAmount*100)) / 100.0

				baseDueDate := trxDate.AddDate(0, 0, 14)
				if trx.PaymentDueDate != nil {
					baseDueDate = *trx.PaymentDueDate
				}

				intervalDays := 14
				if trx.PaymentDueDate != nil && trx.Tenor > 1 {
					totalDays := int(trx.PaymentDueDate.Sub(trxDate).Hours() / 24)
					if totalDays >= trx.Tenor*7 {
						intervalDays = totalDays / trx.Tenor
					}
				}

				for i := 1; i <= trx.Tenor; i++ {
					dueDate := trxDate.AddDate(0, 0, intervalDays*i)
					if i == trx.Tenor && trx.PaymentDueDate != nil {
						dueDate = *trx.PaymentDueDate
					} else if trx.Tenor == 1 && trx.PaymentDueDate != nil {
						dueDate = baseDueDate
					}
					amount := installmentAmount
					if i == trx.Tenor {
						amount = remainingBalance - (installmentAmount * float64(trx.Tenor-1))
					}
					newInst := domain.SalesInstallment{
						SalesTransactionID: trx.ID,
						TermNumber:         i,
						DueDate:            dueDate,
						Amount:             amount,
						PaidAmount:         0,
						Status:             "UNPAID",
					}
					_ = u.db.Create(&newInst).Error
				}
			}
		} else {
			if req.DeferPaymentToCashier && trx.TotalAmount != oldTotalAmount {
				// Keep paid installments intact and spread the corrected balance
				// across terms that are still open.
				var installments []domain.SalesInstallment
				if err := u.db.Where("sales_transaction_id = ?", trx.ID).Order("term_number ASC").Find(&installments).Error; err != nil {
					return fmt.Errorf("gagal membaca cicilan faktur: %w", err)
				}
				var openTerms []domain.SalesInstallment
				maxTerm := 0
				for _, inst := range installments {
					if inst.TermNumber > maxTerm {
						maxTerm = inst.TermNumber
					}
					if inst.Status != "PAID" {
						openTerms = append(openTerms, inst)
					}
				}
				remaining := math.Max(0, trx.TotalAmount-trx.PaidAmount)
				if len(openTerms) == 0 && remaining > 0 {
					dueDate := trxDate.AddDate(0, 0, 14)
					if trx.PaymentDueDate != nil {
						dueDate = *trx.PaymentDueDate
					}
					newInst := domain.SalesInstallment{
						SalesTransactionID: trx.ID,
						TermNumber:         maxTerm + 1,
						DueDate:            dueDate,
						Amount:             remaining,
						Status:             "UNPAID",
					}
					if err := u.db.Create(&newInst).Error; err != nil {
						return fmt.Errorf("gagal membuat cicilan sisa faktur: %w", err)
					}
				} else if len(openTerms) > 0 {
					perTerm := math.Floor(remaining/float64(len(openTerms))*100) / 100
					for i, inst := range openTerms {
						balance := perTerm
						if i == len(openTerms)-1 {
							balance = remaining - perTerm*float64(i)
						}
						status := "UNPAID"
						if balance <= 0 {
							status = "PAID"
						}
						if err := u.db.Model(&domain.SalesInstallment{}).Where("id = ?", inst.ID).
							Updates(map[string]interface{}{"amount": inst.PaidAmount + balance, "status": status}).Error; err != nil {
							return fmt.Errorf("gagal menyesuaikan cicilan faktur: %w", err)
						}
					}
				}
			}
			if trx.PaymentDueDate != nil {
				// Keep paid terms intact, but move the remaining due dates when the
				// customer changes the payment schedule after the first payment.
				intervalDays := int(trx.PaymentDueDate.Sub(trxDate).Hours()/24) / trx.Tenor
				for term := 1; term <= trx.Tenor; term++ {
					dueDate := trxDate.AddDate(0, 0, intervalDays*term)
					if term == trx.Tenor {
						dueDate = *trx.PaymentDueDate
					}
					if err := u.db.Model(&domain.SalesInstallment{}).
						Where("sales_transaction_id = ? AND term_number = ? AND status <> ?", trx.ID, term, "PAID").
						Update("due_date", dueDate).Error; err != nil {
						return fmt.Errorf("gagal memperbarui jadwal cicilan: %w", err)
					}
				}
			}
		}
	}

	// Synchronize linked SalesOrder & delivery items if exists
	var linkedSO domain.SalesOrder
	if err := u.db.Where("invoice_id = ?", trx.ID).First(&linkedSO).Error; err == nil {
		if err := u.db.Model(&linkedSO).Updates(map[string]interface{}{
			"total_amount":     req.TotalAmount,
			"store_id":         req.StoreID,
			"store_category":   req.StoreCategory,
			"notes":            req.Notes,
			"company_id":       trx.CompanyID,
			"payment_method":   trx.PaymentMethod,
			"payment_due_date": trx.PaymentDueDate,
		}).Error; err != nil {
			return fmt.Errorf("gagal memperbarui jatuh tempo Sales Order: %w", err)
		}

		// Replace SalesOrderItem
		u.db.Where("sales_order_id = ?", linkedSO.ID).Delete(&domain.SalesOrderItemBatch{})
		u.db.Where("sales_order_id = ?", linkedSO.ID).Delete(&domain.SalesOrderItem{})
		for _, itemReq := range req.Items {
			soItem := domain.SalesOrderItem{
				SalesOrderID:    linkedSO.ID,
				ProductID:       itemReq.ProductID,
				Quantity:        itemReq.Quantity,
				OrderedQuantity: itemReq.Quantity,
				Unit:            "PCS",
				PiecesPerUnit:   1,
				Price:           itemReq.Price,
				DiscountPercent: itemReq.DiscountPercent,
				Subtotal:        float64(itemReq.Quantity) * itemReq.Price * (1 - itemReq.DiscountPercent/100),
			}
			u.db.Create(&soItem)
		}

		// Replace SalesOrderBonusItem
		u.db.Where("sales_order_id = ?", linkedSO.ID).Delete(&domain.SalesOrderBonusItem{})
		for _, bonusReq := range req.BonusItems {
			bItem := domain.SalesOrderBonusItem{
				SalesOrderID:     linkedSO.ID,
				BonusInventoryID: bonusReq.BonusInventoryID,
				Quantity:         bonusReq.Quantity,
				Notes:            bonusReq.Notes,
			}
			u.db.Create(&bItem)
		}

		// Update pending payment for SalesOrder
		var soPendingPayments []domain.SalesPayment
		if !req.DeferPaymentToCashier {
			if err := u.db.Where("sales_order_id = ? AND payment_status = ?", linkedSO.ID, "PENDING").Find(&soPendingPayments).Error; err == nil {
				for _, p := range soPendingPayments {
					updatedNote := p.Notes
					if !strings.Contains(updatedNote, "[ADJUSTED]") {
						if updatedNote != "" {
							updatedNote = "[ADJUSTED] " + updatedNote
						} else {
							updatedNote = "[ADJUSTED] Penyesuaian oleh Admin Nota"
						}
					}
					u.db.Model(&domain.SalesPayment{}).Where("id = ?", p.ID).Updates(map[string]interface{}{
						"amount":       req.TotalAmount,
						"employee_id":  employeeID,
						"payment_date": trxDate,
						"notes":        updatedNote,
					})
				}
			}
		}
	}

	// If changed to pure TEMPO without down payment, remove old pending payment deposit
	if !req.DeferPaymentToCashier && req.PaymentMethod == "TEMPO" && req.PaidAmount <= 0 {
		u.db.Where("sales_transaction_id = ? AND payment_status = ?", id, "PENDING").Delete(&domain.SalesPayment{})
	}

	paymentUpdates := map[string]interface{}{
		"amount":         req.TotalAmount,
		"employee_id":    employeeID,
		"payment_date":   trxDate,
		"payment_method": trx.PaymentMethod,
	}
	if req.PaidAmount > 0 {
		paymentUpdates["amount"] = req.PaidAmount
	}
	if req.TransferProofURL != "" {
		paymentUpdates["transfer_proof_url"] = req.TransferProofURL
	}

	// Update pending SalesPayment amount, method, proof, and references to match the new transaction
	var trxPendingPayments []domain.SalesPayment
	if !req.DeferPaymentToCashier {
		u.db.Where("sales_transaction_id = ? AND payment_status = ?", id, "PENDING").Find(&trxPendingPayments)
	}
	for _, p := range trxPendingPayments {
		updatedNote := p.Notes
		if !strings.Contains(updatedNote, "[ADJUSTED]") {
			if updatedNote != "" {
				updatedNote = "[ADJUSTED] " + updatedNote
			} else {
				updatedNote = "[ADJUSTED] Penyesuaian oleh Admin Nota"
			}
		}
		paymentUpdates["notes"] = updatedNote
		u.db.Model(&domain.SalesPayment{}).Where("id = ?", p.ID).Updates(paymentUpdates)
	}

	// If no pending payment exists but amount is > 0 and method needs payment, create it
	if !req.DeferPaymentToCashier && len(trxPendingPayments) == 0 && req.TotalAmount > 0 {
		actualPaymentMethod := trx.PaymentMethod
		if actualPaymentMethod == "" {
			actualPaymentMethod = "CASH"
		}

		if actualPaymentMethod == "VA" || actualPaymentMethod == "CASH" || actualPaymentMethod == "TRANSFER" || actualPaymentMethod == "QRIS" {
			paymentStatus := "PENDING"
			paymentAmt := req.TotalAmount
			if req.PaidAmount > 0 {
				paymentAmt = req.PaidAmount
			}
			var proof *string
			if req.TransferProofURL != "" {
				proof = &req.TransferProofURL
			}
			payment := &domain.SalesPayment{
				SalesTransactionID: &trx.ID,
				EmployeeID:         employeeID,
				Amount:             paymentAmt,
				PaymentMethod:      actualPaymentMethod,
				PaymentStatus:      paymentStatus,
				PaymentDate:        trxDate,
				TransferProofURL:   proof,
				Notes:              fmt.Sprintf("Setoran nota (diperbarui dari edit faktur: %s)", actualPaymentMethod),
			}
			_ = u.db.Create(payment).Error
		} else if actualPaymentMethod == "TEMPO" && req.PaidAmount > 0 {
			dpMethod := req.DPPaymentMethod
			if dpMethod == "" {
				dpMethod = "CASH"
			}
			var proof *string
			if req.TransferProofURL != "" {
				proof = &req.TransferProofURL
			}
			payment := &domain.SalesPayment{
				SalesTransactionID: &trx.ID,
				EmployeeID:         employeeID,
				Amount:             req.PaidAmount,
				PaymentMethod:      dpMethod,
				PaymentStatus:      "PENDING",
				PaymentDate:        trxDate,
				TransferProofURL:   proof,
				Notes:              fmt.Sprintf("Uang Muka (DP) Transaksi Tempo (%s - %s)", trx.ReceiptNo, dpMethod),
			}
			_ = u.db.Create(payment).Error
		}
	}

	// Recalculate KPI for current (new) period/employee
	_ = u.CalculateKPI(trx.EmployeeID, trx.PeriodMonth, trx.PeriodYear)

	// If period or employee changed, recalculate the old one too
	if oldEmployeeID != trx.EmployeeID || oldMonth != trx.PeriodMonth || oldYear != trx.PeriodYear {
		_ = u.CalculateKPI(oldEmployeeID, oldMonth, oldYear)
	}

	invalidateReportsCache(context.Background())
	return nil
}

func (u *salesUsecase) UpdateTransactionMobile(id uuid.UUID, req MobileUpdateTransactionRequest) error {
	trx, err := u.salesRepo.FindByID(id)
	if err != nil {
		return err
	}
	if trx == nil {
		return errors.New("transaction not found")
	}

	// Check if this transaction or its parent SalesOrder is linked to a delivery batch that is actively on the road (ON_DELIVERY)
	var count int64
	dbErr := u.db.Table("delivery_items").
		Joins("JOIN delivery_batches ON delivery_batches.id = delivery_items.delivery_batch_id").
		Joins("LEFT JOIN sales_orders ON sales_orders.id = delivery_items.sales_order_id").
		Where("(delivery_items.sales_transaction_id = ? OR sales_orders.invoice_id = ?) AND delivery_batches.status = ?", id, id, "ON_DELIVERY").
		Count(&count).Error
	if dbErr == nil && count > 0 {
		return errors.New("tidak bisa mengubah faktur yang ada di dalam surat jalan yang sedang dalam proses pengiriman (ON_DELIVERY)")
	}

	trx.TotalAmount = req.TotalAmount

	// Update items
	trx.Items = []domain.SalesItem{} // Clear existing slice in memory
	for _, itemReq := range req.Items {
		trx.Items = append(trx.Items, domain.SalesItem{
			SalesTransactionID: trx.ID,
			ProductID:          itemReq.ProductID,
			Quantity:           itemReq.Quantity,
			PriceAtTransaction: itemReq.Price,
			DiscountPercent:    itemReq.DiscountPercent,
			Subtotal:           float64(itemReq.Quantity) * itemReq.Price * (1 - itemReq.DiscountPercent/100),
		})
	}

	err = u.salesRepo.Update(trx)
	if err != nil {
		return err
	}

	// Synchronize linked SalesOrder & delivery items if exists
	var linkedSO domain.SalesOrder
	if err := u.db.Where("invoice_id = ?", trx.ID).First(&linkedSO).Error; err == nil {
		u.db.Model(&linkedSO).Updates(map[string]interface{}{
			"total_amount": req.TotalAmount,
		})

		u.db.Where("sales_order_id = ?", linkedSO.ID).Delete(&domain.SalesOrderItemBatch{})
		u.db.Where("sales_order_id = ?", linkedSO.ID).Delete(&domain.SalesOrderItem{})
		for _, itemReq := range req.Items {
			soItem := domain.SalesOrderItem{
				SalesOrderID:    linkedSO.ID,
				ProductID:       itemReq.ProductID,
				Quantity:        itemReq.Quantity,
				OrderedQuantity: itemReq.Quantity,
				Unit:            "PCS",
				PiecesPerUnit:   1,
				Price:           itemReq.Price,
				DiscountPercent: itemReq.DiscountPercent,
				Subtotal:        float64(itemReq.Quantity) * itemReq.Price * (1 - itemReq.DiscountPercent/100),
			}
			u.db.Create(&soItem)
		}

		u.db.Model(&domain.SalesPayment{}).
			Where("sales_order_id = ? AND payment_status = ?", linkedSO.ID, "PENDING").
			Updates(map[string]interface{}{
				"amount": req.TotalAmount,
			})
	}

	// Update pending SalesPayment amount to match the new transaction amount
	result := u.db.Model(&domain.SalesPayment{}).
		Where("sales_transaction_id = ? AND payment_status = ?", id, "PENDING").
		Updates(map[string]interface{}{
			"amount": req.TotalAmount,
		})

	// If no pending payment exists but amount is > 0 and method needs payment, create it
	if result.RowsAffected == 0 && req.TotalAmount > 0 {
		actualPaymentMethod := trx.PaymentMethod
		if actualPaymentMethod == "" {
			actualPaymentMethod = "CASH"
		}

		if actualPaymentMethod == "VA" || actualPaymentMethod == "CASH" {
			paymentStatus := "SUCCESS"
			if actualPaymentMethod == "VA" || actualPaymentMethod == "CASH" {
				paymentStatus = "PENDING"
			}
			payment := &domain.SalesPayment{
				SalesTransactionID: &trx.ID,
				EmployeeID:         trx.EmployeeID,
				Amount:             req.TotalAmount,
				PaymentMethod:      actualPaymentMethod,
				PaymentStatus:      paymentStatus,
				PaymentDate:        trx.TransactionDate,
				Notes:              "Setoran nota (diperbarui dari edit pesanan mobile)",
			}
			_ = u.db.Create(payment).Error
		}
	}

	// Recalculate KPI for current period/employee
	_ = u.CalculateKPI(trx.EmployeeID, trx.PeriodMonth, trx.PeriodYear)

	invalidateReportsCache(context.Background())
	return nil
}

func (u *salesUsecase) DeleteTransaction(id uuid.UUID) error {
	trx, err := u.salesRepo.FindByID(id)
	if err != nil {
		return err
	}
	if trx == nil {
		return errors.New("transaction not found")
	}

	// Check if this transaction or its parent SalesOrder is linked to a delivery batch that has been confirmed ready (PICKING, ON_DELIVERY, COMPLETED)
	var count int64
	dbErr := u.db.Table("delivery_items").
		Joins("JOIN delivery_batches ON delivery_batches.id = delivery_items.delivery_batch_id").
		Joins("LEFT JOIN sales_orders ON sales_orders.id = delivery_items.sales_order_id").
		Where("(delivery_items.sales_transaction_id = ? OR sales_orders.invoice_id = ?) AND delivery_batches.status IN ?", id, id,
			[]string{"PICKING", "ON_DELIVERY", "COMPLETED"}).
		Count(&count).Error
	if dbErr == nil && count > 0 {
		return errors.New("tidak bisa menghapus faktur yang ada di dalam surat jalan yang sudah dikonfirmasi siap, dikirim, atau selesai")
	}

	err = u.salesRepo.Delete(id)
	if err != nil {
		return err
	}

	// Recalculate KPI after deletion
	_ = u.CalculateKPI(trx.EmployeeID, trx.PeriodMonth, trx.PeriodYear)

	invalidateReportsCache(context.Background())
	return nil
}

func (u *salesUsecase) SoftDeleteTransaction(id uuid.UUID, deletedByID uuid.UUID, reason string) error {
	trx, err := u.salesRepo.FindByID(id)
	if err != nil {
		return err
	}
	if trx == nil {
		return errors.New("transaction not found")
	}

	// Check if this transaction or its parent SalesOrder is linked to a delivery batch that has been confirmed ready (PICKING, ON_DELIVERY, COMPLETED)
	var count int64
	dbErr := u.db.Table("delivery_items").
		Joins("JOIN delivery_batches ON delivery_batches.id = delivery_items.delivery_batch_id").
		Joins("LEFT JOIN sales_orders ON sales_orders.id = delivery_items.sales_order_id").
		Where("(delivery_items.sales_transaction_id = ? OR sales_orders.invoice_id = ?) AND delivery_batches.status IN ?", id, id,
			[]string{"PICKING", "ON_DELIVERY", "COMPLETED"}).
		Count(&count).Error
	if dbErr == nil && count > 0 {
		return errors.New("tidak bisa menghapus faktur yang ada di dalam surat jalan yang sudah dikonfirmasi siap, dikirim, atau selesai")
	}

	err = u.salesRepo.SoftDeleteWithReason(id, deletedByID, reason)
	if err != nil {
		return err
	}

	// Recalculate KPI after deletion
	_ = u.CalculateKPI(trx.EmployeeID, trx.PeriodMonth, trx.PeriodYear)

	invalidateReportsCache(context.Background())
	return nil
}

func (u *salesUsecase) DeletePendingTransactionBySales(id, employeeID uuid.UUID) error {
	var periodMonth, periodYear int
	err := u.db.Transaction(func(tx *gorm.DB) error {
		var trx domain.SalesTransaction
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Preload("Items").Preload("BonusItems").First(&trx, "id = ?", id).Error; err != nil {
			return errors.New("transaksi tidak ditemukan")
		}
		if trx.EmployeeID != employeeID {
			return errors.New("transaksi bukan milik sales ini")
		}
		if trx.Status != "PENDING" || trx.VerifiedAt != nil || trx.VerifiedByID != nil {
			return errors.New("hanya transaksi PENDING yang belum divalidasi dapat dihapus")
		}
		if trx.PaidAmount > 0 || trx.PaymentStatus == string(domain.PaymentStatusPaid) || trx.PaymentStatus == string(domain.PaymentStatusPartial) {
			return errors.New("transaksi dengan pembayaran tercatat tidak dapat dihapus oleh sales")
		}
		var related int64
		if err := tx.Model(&domain.SalesOrder{}).Where("invoice_id = ?", id).Count(&related).Error; err != nil {
			return err
		}
		if related > 0 {
			return errors.New("transaksi yang terkait Sales Order tidak dapat dihapus di sini")
		}
		if err := tx.Model(&domain.SalesPayment{}).Where("sales_transaction_id = ?", id).Count(&related).Error; err != nil {
			return err
		}
		if related > 0 {
			return errors.New("transaksi dengan pembayaran tercatat tidak dapat dihapus oleh sales")
		}
		if err := tx.Model(&domain.DeliveryItem{}).Where("sales_transaction_id = ?", id).Count(&related).Error; err != nil {
			return err
		}
		if related > 0 {
			return errors.New("transaksi yang sudah masuk pengiriman tidak dapat dihapus")
		}
		if err := tx.Model(&domain.SalesReturn{}).Where("transaction_id = ?", id).Count(&related).Error; err != nil {
			return err
		}
		if related > 0 {
			return errors.New("transaksi dengan retur tidak dapat dihapus")
		}

		for _, item := range trx.Items {
			result := tx.Model(&domain.SalesStock{}).
				Where("employee_id = ? AND product_id = ?", employeeID, item.ProductID).
				UpdateColumn("quantity", gorm.Expr("quantity + ?", item.Quantity))
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected == 0 {
				return errors.New("stok sales untuk salah satu produk tidak ditemukan")
			}
		}
		for _, bonus := range trx.BonusItems {
			result := tx.Model(&domain.SalesBonusStock{}).
				Where("employee_id = ? AND bonus_inventory_id = ?", employeeID, bonus.BonusInventoryID).
				UpdateColumn("quantity", gorm.Expr("quantity + ?", bonus.Quantity))
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected == 0 {
				return errors.New("stok bonus sales tidak ditemukan")
			}
		}
		periodMonth, periodYear = trx.PeriodMonth, trx.PeriodYear
		return repository.NewSalesTransactionRepository(tx).SoftDeleteWithReason(id, employeeID, "Dihapus oleh sales sebelum validasi Admin Nota")
	})
	if err != nil {
		return err
	}
	_ = u.CalculateKPI(employeeID, periodMonth, periodYear)
	invalidateReportsCache(context.Background())
	return nil
}

func (u *salesUsecase) RestoreTransaction(id uuid.UUID) error {
	trx, err := u.salesRepo.FindByIDUnscoped(id)
	if err != nil {
		return err
	}
	if trx == nil {
		return errors.New("transaction not found")
	}

	err = u.salesRepo.Restore(id)
	if err != nil {
		return err
	}

	// Recalculate KPI after restoration
	_ = u.CalculateKPI(trx.EmployeeID, trx.PeriodMonth, trx.PeriodYear)

	invalidateReportsCache(context.Background())
	return nil
}

func (u *salesUsecase) SetKPITarget(employeeID uuid.UUID, month, year int, targetOmzet float64, targetNewStores int, targetDailyVisits int, targetDailyNewStores int, workingTerritory string, receivableLimitPercent float64, maxReceivableLimit float64, commission float64) error {
	kpi, err := u.performanceRepo.GetSalesKPIByEmployeeAndPeriod(employeeID, month, year)
	if err != nil {
		return err
	}

	if kpi == nil {
		kpi = &domain.SalesKPI{
			EmployeeID:  employeeID,
			PeriodMonth: month,
			PeriodYear:  year,
		}
	}

	kpi.TargetOmzet = targetOmzet
	kpi.TargetNewStores = targetNewStores
	kpi.TargetDailyVisits = targetDailyVisits
	kpi.TargetDailyNewStores = targetDailyNewStores
	kpi.WorkingTerritory = workingTerritory
	kpi.ReceivableLimitPercent = receivableLimitPercent
	kpi.MaxReceivableLimit = maxReceivableLimit
	kpi.Commission = commission

	return u.performanceRepo.SaveSalesKPI(kpi)
}

func (u *salesUsecase) DeleteKPITarget(id uuid.UUID) error {
	return u.performanceRepo.DeleteSalesKPI(id)
}

func (u *salesUsecase) GetVisitPlan(employeeID uuid.UUID, date time.Time) ([]VisitPlanItem, error) {
	// Ensure we align the reference date to Asia/Jakarta timezone location
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err == nil {
		date = date.In(loc)
	}

	// 1. Get Day Number (Monday=1, Sunday=7)
	day := int(date.Weekday())
	if day == 0 {
		day = 7
	}

	// 2. Fetch Scheduled Stores
	scheduledStores, err := u.storeRepo.FindByVisitDay(day, employeeID)
	if err != nil {
		return nil, err
	}

	// 3. Fetch Actual Visits for today
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	actualVisits, err := u.visitRepo.FindAllHistory(100, 0, &employeeID, nil, "", "", &startOfDay, &endOfDay, true)
	if err != nil {
		return nil, err
	}

	// FindAllHistory returns visits in DESC order (newest first).
	// Only keep the FIRST (= most recent) visit seen per store.
	visitMap := make(map[uuid.UUID]domain.SalesVisit) // StoreID -> latest SalesVisit
	for _, v := range actualVisits {
		if _, alreadyMapped := visitMap[v.StoreID]; !alreadyMapped {
			visitMap[v.StoreID] = v
		}
	}

	var result []VisitPlanItem

	// Add manual visit plans
	manualPlans, _ := u.visitPlanRepo.GetByEmployeeAndDate(employeeID, date)

	// Create a map for manual plans to get projections easily
	manualPlanMap := make(map[uuid.UUID]domain.SalesVisitPlan)
	for _, plan := range manualPlans {
		manualPlanMap[plan.StoreID] = plan
	}

	// Add Scheduled Stores
	scheduledMap := make(map[uuid.UUID]bool)
	for _, store := range scheduledStores {
		status := "PLANNED"
		var visitID *uuid.UUID
		if v, exists := visitMap[store.ID]; exists {
			if v.Status == "SKIPPED" {
				continue // Do not include skipped stores
			}
			if v.CheckOutTime != nil {
				status = "COMPLETED"
			} else {
				status = "CHECKED_IN"
			}
			visitID = &v.ID
		}

		var projections []VisitPlanProjectionItem
		var projectedOmzet float64
		if plan, exists := manualPlanMap[store.ID]; exists {
			for _, proj := range plan.Projections {
				price := proj.Product.SellingPrice
				if proj.Product.FactoryPrice > 0 {
					price = proj.Product.FactoryPrice // Using FactoryPrice or regular Price? Let's use Price
				}
				projections = append(projections, VisitPlanProjectionItem{
					ProductID:   proj.ProductID,
					ProductName: proj.Product.Name,
					Quantity:    proj.Quantity,
					Price:       price,
				})
				projectedOmzet += float64(proj.Quantity) * price
			}
		}

		result = append(result, VisitPlanItem{
			StoreID:        store.ID,
			StoreName:      store.Name,
			Address:        store.Address,
			Status:         status,
			IsExtra:        false,
			VisitID:        visitID,
			Projections:    projections,
			ProjectedOmzet: projectedOmzet,
		})
		scheduledMap[store.ID] = true
	}

	for _, plan := range manualPlans {
		if scheduledMap[plan.StoreID] {
			continue // Already added from scheduled stores
		}

		status := "PLANNED"
		var visitID *uuid.UUID
		if v, exists := visitMap[plan.StoreID]; exists {
			if v.Status == "SKIPPED" {
				continue
			}
			if v.CheckOutTime != nil {
				status = "COMPLETED"
			} else {
				status = "CHECKED_IN"
			}
			visitID = &v.ID
		}

		var projections []VisitPlanProjectionItem
		var projectedOmzet float64
		for _, proj := range plan.Projections {
			price := proj.Product.SellingPrice
			if proj.Product.FactoryPrice > 0 {
				price = proj.Product.FactoryPrice
			}
			projections = append(projections, VisitPlanProjectionItem{
				ProductID:   proj.ProductID,
				ProductName: proj.Product.Name,
				Quantity:    proj.Quantity,
				Price:       price,
			})
			projectedOmzet += float64(proj.Quantity) * price
		}

		result = append(result, VisitPlanItem{
			StoreID:        plan.StoreID,
			StoreName:      plan.Store.Name,
			Address:        plan.Store.Address,
			Status:         status,
			IsExtra:        false, // Manual plans are considered planned
			VisitID:        visitID,
			Projections:    projections,
			ProjectedOmzet: projectedOmzet,
		})
		scheduledMap[plan.StoreID] = true
	}

	// Add Extra Visits (Visited but not in schedule or manual plans).
	// Use visitMap to avoid duplicates — it already holds the latest visit per store.
	for storeID, v := range visitMap {
		if !scheduledMap[storeID] {
			if v.Status == "SKIPPED" {
				continue
			}
			status := "EXTRA"
			if v.CheckOutTime == nil {
				status = "CHECKED_IN"
			}
			visitIDCopy := v.ID
			result = append(result, VisitPlanItem{
				StoreID:   v.StoreID,
				StoreName: v.Store.Name,
				Address:   v.Store.Address,
				Status:    status,
				IsExtra:   true,
				VisitID:   &visitIDCopy,
			})
		}
	}

	// Fetch store histories to populate last visit & transaction data
	var storeIDs []uuid.UUID
	for _, item := range result {
		storeIDs = append(storeIDs, item.StoreID)
	}
	if len(storeIDs) > 0 {
		if storeHistories, err := u.storeRepo.GetStoreHistories(storeIDs); err == nil {
			for idx, item := range result {
				if hist, exists := storeHistories[item.StoreID]; exists {
					result[idx].LastVisitDate = hist.LastVisitDate
					result[idx].LastTransactionDate = hist.LastTransactionDate
					result[idx].LastTransactionAmount = hist.LastTransactionAmount
					result[idx].LastTransactionItems = hist.LastTransactionItems
				}
			}
		}
	}

	return result, nil
}

func (u *salesUsecase) GetTransactionByReceipt(receiptNo string) (*domain.SalesTransaction, error) {
	trx, err := u.salesRepo.FindByReceiptNo(receiptNo)
	if err != nil {
		return nil, err
	}
	if trx == nil {
		return nil, nil
	}

	// Active Status Polling: If transaction is digital and still PENDING, check directly with Midtrans
	if (trx.PaymentMethod == "QRIS" || trx.PaymentMethod == "VA") && trx.Status == "PENDING" && u.midtransClient != nil {
		fmt.Printf("DEBUG ACTIVE POLLING: Checking status directly with Midtrans for %s\n", receiptNo)
		statusResp, checkErr := u.midtransClient.CheckTransaction(receiptNo)
		if checkErr == nil && statusResp != nil {
			status := statusResp.TransactionStatus
			fraud := statusResp.FraudStatus

			isSettled := (status == "capture" && fraud == "accept") || status == "settlement"
			isFailed := status == "deny" || status == "expire" || status == "cancel"

			if isSettled {
				trx.PaymentStatus = string(domain.PaymentStatusPaid)
				trx.PaidAmount = trx.TotalAmount
				trx.Status = "VERIFIED"

				_ = u.salesRepo.UpdateTransactionOnly(trx)
				fmt.Printf("DEBUG ACTIVE POLLING SUCCESS: Transaction %s is settled! Status updated to VERIFIED/PAID in DB.\n", receiptNo)
			} else if isFailed {
				trx.PaymentStatus = "FAILED"
				_ = u.salesRepo.UpdateTransactionOnly(trx)
				fmt.Printf("DEBUG ACTIVE POLLING FAILED: Transaction %s is failed! Status updated to FAILED in DB.\n", receiptNo)
			}
		} else {
			fmt.Printf("DEBUG ACTIVE POLLING WARNING: Failed to check status directly with Midtrans for %s: %v\n", receiptNo, checkErr)
		}
	}

	return trx, nil
}

func (u *salesUsecase) GetTransactionsByStatusAndCompany(status string, companyID uuid.UUID) ([]domain.SalesTransaction, error) {
	return u.salesRepo.FindByStatusAndCompany(status, companyID)
}
func (u *salesUsecase) GetDeliveryPending(companyID uuid.UUID) ([]domain.SalesTransaction, error) {
	return u.salesRepo.FindDeliveryPending(companyID)
}

func (u *salesUsecase) RejectTransaction(id uuid.UUID, notes string) error {
	var empID uuid.UUID
	var pMonth, pYear int

	err := u.db.Transaction(func(tx *gorm.DB) error {
		// 1. Get the SalesTransaction (Invoice)
		var trx domain.SalesTransaction
		if err := tx.Preload("Items").Where("id = ?", id).First(&trx).Error; err != nil {
			return fmt.Errorf("transaction not found: %w", err)
		}

		if trx.Status != "PENDING" {
			return errors.New("hanya faktur PENDING yang dapat ditolak")
		}

		empID = trx.EmployeeID
		pMonth = trx.PeriodMonth
		pYear = trx.PeriodYear

		// 2. Find the associated SalesOrder (if any)
		var so domain.SalesOrder
		hasSO := false
		if err := tx.Preload("Items").Where("invoice_id = ?", trx.ID).First(&so).Error; err == nil {
			hasSO = true
		} else if trx.ReceiptNo != "" {
			if errSo := tx.Preload("Items").Where("so_number = ?", trx.ReceiptNo).First(&so).Error; errSo == nil {
				hasSO = true
			}
		}

		wRepo := repository.NewWarehouseRepository(tx)

		// 3. Restore warehouse stock
		if hasSO {
			for _, item := range so.Items {
				batchNo := "DEFAULT"
				var soiBatch domain.SalesOrderItemBatch
				if err := tx.Where("sales_order_item_id = ?", item.ID).First(&soiBatch).Error; err == nil && soiBatch.BatchNo != "" {
					batchNo = soiBatch.BatchNo
				}

				stock, err := wRepo.GetStock(so.BranchID, item.ProductID, "GOOD", batchNo)
				if err != nil {
					stock = &domain.WarehouseStock{
						BranchID:  so.BranchID,
						ProductID: item.ProductID,
						StockType: "GOOD",
						BatchNo:   batchNo,
						Quantity:  0,
					}
					if err := tx.Create(stock).Error; err != nil {
						return fmt.Errorf("failed to create warehouse stock: %w", err)
					}
				}

				qtyToRestore := item.OrderedQuantity
				if qtyToRestore <= 0 {
					qtyToRestore = item.Quantity
				}
				stock.Quantity += qtyToRestore
				if err := wRepo.UpdateStock(stock); err != nil {
					return fmt.Errorf("failed to restore warehouse stock: %w", err)
				}

				wRepo.CreateLog(&domain.WarehouseLog{
					BranchID:  so.BranchID,
					ProductID: item.ProductID,
					Type:      "IN",
					Source:    fmt.Sprintf("BATAL_PESANAN: %s (Ditolak)", so.SONumber),
					Quantity:  qtyToRestore,
					BatchNo:   batchNo,
				})
			}

			// Clean up pending SalesTransfer (RTN-MTN) created during AdminReportBatch for this SO to prevent double-restoring
			tx.Where("branch_id = ? AND type = ? AND status = ? AND notes LIKE ?",
				so.BranchID, "RETURN", "PENDING", "%Pengembalian Muatan (Gagal Kirim)%").Delete(&domain.SalesTransfer{})

			// Update SO status to CANCELLED
			so.Status = domain.SOStatusCancelled
			if err := tx.Save(&so).Error; err != nil {
				return fmt.Errorf("failed to update sales order status: %w", err)
			}

			// Hapus antrean setoran kasir PENDING
			tx.Where("(sales_order_id = ? OR sales_transaction_id = ?) AND payment_status = ?", so.ID, trx.ID, "PENDING").Delete(&domain.SalesPayment{})
		} else if trx.SalesChannel == domain.SalesChannelDirectFactory || trx.SalesChannel == domain.SalesChannelDirectBranch || trx.SalesChannel == domain.SalesChannelEmployeeSale {
			// Direct sales also deduct warehouse stock on creation
			if trx.BranchID != nil && *trx.BranchID != uuid.Nil {
				for _, item := range trx.Items {
					stock, err := wRepo.GetStock(*trx.BranchID, item.ProductID, "GOOD", "DEFAULT")
					if err == nil && stock != nil {
						stock.Quantity += item.Quantity
						_ = wRepo.UpdateStock(stock)

						wRepo.CreateLog(&domain.WarehouseLog{
							BranchID:  *trx.BranchID,
							ProductID: item.ProductID,
							Type:      "IN",
							Source:    fmt.Sprintf("BATAL_PENJUALAN_LANGSUNG: %s (Ditolak)", trx.ReceiptNo),
							Quantity:  item.Quantity,
							BatchNo:   "DEFAULT",
						})
					}
				}
			}
			tx.Where("sales_transaction_id = ? AND payment_status = ?", trx.ID, "PENDING").Delete(&domain.SalesPayment{})
		} else {
			tx.Where("sales_transaction_id = ? AND payment_status = ?", trx.ID, "PENDING").Delete(&domain.SalesPayment{})
		}

		// 4. Update transaction status to REJECTED
		trx.Status = "REJECTED"
		if notes != "" {
			oldNotes := ""
			if trx.Notes != nil {
				oldNotes = *trx.Notes
			}
			var newNotes string
			if oldNotes != "" {
				newNotes = fmt.Sprintf("%s | Rejected: %s", oldNotes, notes)
			} else {
				newNotes = fmt.Sprintf("Rejected: %s", notes)
			}
			trx.Notes = &newNotes
		}

		if err := tx.Save(&trx).Error; err != nil {
			return fmt.Errorf("failed to save transaction: %w", err)
		}

		return nil
	})

	if err != nil {
		return err
	}

	// Recalculate KPI for current period/employee
	if empID != uuid.Nil && pMonth > 0 && pYear > 0 {
		_ = u.CalculateKPI(empID, pMonth, pYear)
	}

	invalidateReportsCache(context.Background())
	return nil
}

func (u *salesUsecase) CountByCompanyAndDate(companyID uuid.UUID, date time.Time) (int64, error) {
	return u.salesRepo.CountByCompanyAndDate(companyID, date)
}

func (u *salesUsecase) HandleMidtransNotification(payload map[string]interface{}) error {
	orderID, _ := payload["order_id"].(string)
	baseOrderID := orderID
	if idx := strings.LastIndex(orderID, "-R"); idx != -1 {
		baseOrderID = orderID[:idx]
	}
	transactionStatus, _ := payload["transaction_status"].(string)
	fraudStatus, _ := payload["fraud_status"].(string)

	fmt.Printf("DEBUG MIDTRANS CALLBACK: order_id=%s, status=%s\n", orderID, transactionStatus)

	// 1. Signature Key Verification (Keamanan Verifikasi Status)
	serverKey := os.Getenv("MIDTRANS_SERVER_KEY")
	if serverKey != "" {
		sig, _ := payload["signature_key"].(string)
		var statusCode string
		if sc, ok := payload["status_code"].(string); ok {
			statusCode = sc
		} else if scNum, ok := payload["status_code"].(float64); ok {
			statusCode = fmt.Sprintf("%.0f", scNum)
		}

		var grossAmount string
		if ga, ok := payload["gross_amount"].(string); ok {
			grossAmount = ga
		} else if gaNum, ok := payload["gross_amount"].(float64); ok {
			grossAmount = fmt.Sprintf("%.2f", gaNum)
		}

		// Calculate hash
		input := orderID + statusCode + grossAmount + serverKey
		hasher := sha512.New()
		hasher.Write([]byte(input))
		calculatedSig := hex.EncodeToString(hasher.Sum(nil))

		isSignatureValid := sig == calculatedSig
		if !isSignatureValid && strings.Contains(grossAmount, ".") {
			// Try fallback without decimals (in case gross_amount is "10000" instead of "10000.00")
			parts := strings.Split(grossAmount, ".")
			input2 := orderID + statusCode + parts[0] + serverKey
			hasher2 := sha512.New()
			hasher2.Write([]byte(input2))
			calculatedSig2 := hex.EncodeToString(hasher2.Sum(nil))
			isSignatureValid = sig == calculatedSig2
		}

		if !isSignatureValid {
			fmt.Printf("WARNING MIDTRANS CALLBACK: Signature verification failed for order_id=%s. Calculated signature mismatch.\n", orderID)
			// Return error in production to abort unauthorized notifications
			if os.Getenv("MIDTRANS_IS_PRODUCTION") == "true" {
				return errors.New("invalid midtrans signature key")
			}
		} else {
			fmt.Printf("DEBUG MIDTRANS CALLBACK: Signature verified successfully for order_id=%s\n", orderID)
		}
	}

	// Check if orderID starts with "PAY-SO-"
	if strings.HasPrefix(orderID, "PAY-SO-") {
		// Format: PAY-SO-SO_NUMBER-TIMESTAMP
		// Let's parse out the SO number.
		parts := strings.Split(orderID, "-")
		if len(parts) >= 4 {
			// e.g. "PAY", "SO", "SO-2026-0001", "1710928131"
			// The SO number could contain hyphens.
			// Let's extract everything between parts[2] and parts[len(parts)-1]
			soNumber := strings.Join(parts[2:len(parts)-1], "-")
			fmt.Printf("DEBUG MIDTRANS CALLBACK: resolved SO number=%s\n", soNumber)

			db := u.salesRepo.GetDB()
			var so domain.SalesOrder
			if err := db.Where("so_number = ?", soNumber).First(&so).Error; err != nil {
				return fmt.Errorf("sales order not found for number %s: %w", soNumber, err)
			}

			// Idempotency: skip if already paid
			if so.PaymentStatus == "PAID" {
				fmt.Printf("DEBUG MIDTRANS CALLBACK: SalesOrder %s already processed (PAID). Skipping duplicate settlement.\n", soNumber)
				return nil
			}

			isSettled := (transactionStatus == "capture" && fraudStatus == "accept") || transactionStatus == "settlement"
			isFailed := transactionStatus == "deny" || transactionStatus == "expire" || transactionStatus == "cancel"

			if isSettled {
				so.PaymentStatus = "PAID"
				so.Status = domain.SOStatusPaid
				so.PaymentCollectedAmount = so.TotalAmount
				now := time.Now()
				so.PaymentCollectedAt = &now
				db.Save(&so)

				if u.sseHub != nil {
					go func() {
						event := sse.NotificationEvent{
							ID:        uuid.New().String(),
							Type:      "payment",
							Title:     "Pembayaran Sukses (Midtrans)",
							Message:   fmt.Sprintf("Pembayaran Midtrans untuk Sales Order %s sebesar Rp %.2f sukses.", so.SONumber, so.TotalAmount),
							CompanyID: so.CompanyID.String(),
							CreatedAt: time.Now().Format("2006-01-02 15:04:05"),
							Data: map[string]interface{}{
								"so_id":        so.ID.String(),
								"so_number":    so.SONumber,
								"total_amount": so.TotalAmount,
								"method":       "MIDTRANS",
							},
						}
						_ = u.sseHub.PublishNotification(context.Background(), event)
					}()
				}

				// Settle delivery item cash totals if linked
				var di domain.DeliveryItem
				if err := db.Where("sales_order_id = ?", so.ID).First(&di).Error; err == nil {
					di.PaymentCollected = true
					di.PaymentAmount = so.TotalAmount
					di.PaymentMethod = "QRIS"
					di.PaymentCollectedAt = &now
					db.Save(&di)

					// Update Batch Totals
					var batch domain.DeliveryBatch
					if err := db.First(&batch, di.DeliveryBatchID).Error; err == nil {
						batch.TotalTransferCollected += so.TotalAmount
						db.Save(&batch)
					}
				}

				// Update corresponding invoice if exists
				if so.InvoiceID != nil {
					var st domain.SalesTransaction
					if err := db.First(&st, *so.InvoiceID).Error; err == nil {
						st.PaymentStatus = string(domain.PaymentStatusPaid)
						st.PaidAmount = st.TotalAmount
						st.Status = "VERIFIED" // Lunas (Final)
						st.PaymentMethod = "QRIS"
						db.Save(&st)
						go func(empID uuid.UUID, m, y int) {
							_ = u.CalculateKPI(empID, m, y)
						}(st.EmployeeID, st.PeriodMonth, st.PeriodYear)
					}
				} else {
					if so.EmployeeID != uuid.Nil {
						go func(empID uuid.UUID, m, y int) {
							_ = u.CalculateKPI(empID, m, y)
						}(so.EmployeeID, int(so.OrderDate.Month()), so.OrderDate.Year())
					}
				}
			} else if isFailed {
				so.PaymentStatus = "FAILED"
				db.Save(&so)

				if so.InvoiceID != nil {
					var st domain.SalesTransaction
					if err := db.First(&st, *so.InvoiceID).Error; err == nil {
						st.PaymentStatus = "FAILED"
						db.Save(&st)
					}
				}
			}
			return nil
		}
	}

	trx, err := u.salesRepo.FindByReceiptNo(baseOrderID)
	if err != nil || trx == nil {
		return errors.New("transaction not found for midtrans notification")
	}

	// Idempotency: skip if already verified and paid
	if trx.Status == "VERIFIED" && trx.PaymentStatus == string(domain.PaymentStatusPaid) {
		fmt.Printf("DEBUG MIDTRANS CALLBACK: Transaction %s already processed. Skipping duplicate settlement.\n", orderID)
		return nil
	}

	if (transactionStatus == "capture" && fraudStatus == "accept") || transactionStatus == "settlement" {
		trx.PaymentStatus = string(domain.PaymentStatusPaid)
		trx.PaidAmount = trx.TotalAmount
		trx.Status = "VERIFIED"
	} else if transactionStatus == "deny" || transactionStatus == "expire" || transactionStatus == "cancel" {
		trx.PaymentStatus = "FAILED"
	}

	err = u.salesRepo.UpdateTransactionOnly(trx)
	if err == nil {
		invalidateReportsCache(context.Background())
		if trx.Status == "VERIFIED" {
			go func(empID uuid.UUID, m, y int) {
				_ = u.CalculateKPI(empID, m, y)
			}(trx.EmployeeID, trx.PeriodMonth, trx.PeriodYear)
		}
	}

	return err
}

func (u *salesUsecase) GetProductSalesDistribution(productIDs []uuid.UUID, companyID uuid.UUID) ([]map[string]interface{}, error) {
	return u.salesRepo.GetProductSalesDistribution(productIDs, companyID)
}

func (u *salesUsecase) RecordVisit(req RecordVisitRequest) error {
	fmt.Printf("DEBUG: SalesUsecase.RecordVisit called for employee %v and store %v\n", req.EmployeeID, req.StoreID)

	if req.ClientID != "" {
		var existing domain.SalesVisit
		if err := u.db.Where("client_id = ?", req.ClientID).First(&existing).Error; err == nil {
			fmt.Printf("DEBUG: Prevented duplicate RecordVisit for ClientID %s\n", req.ClientID)
			return nil
		}
	}

	lat := req.Latitude
	lng := req.Longitude
	if lat == 0 && lng == 0 && req.StoreID != uuid.Nil {
		var store domain.Store
		if err := u.db.Select("latitude, longitude").First(&store, "id = ?", req.StoreID).Error; err == nil {
			if store.Latitude != 0 || store.Longitude != 0 {
				lat = store.Latitude
				lng = store.Longitude
				fmt.Printf("DEBUG: Fallback visit coordinates to store coordinates (%v, %v) for store %v\n", lat, lng, req.StoreID)
			}
		}
	}

	visit := &domain.SalesVisit{
		EmployeeID:  req.EmployeeID,
		StoreID:     req.StoreID,
		CheckInTime: req.CheckTime,
		Latitude:    lat,
		Longitude:   lng,
		SelfieURL:   req.SelfieURL,
		Type:        req.Type,
		Notes:       req.Notes,
	}
	if req.ClientID != "" {
		visit.ClientID = &req.ClientID
	}

	if req.Type == "CHECKOUT" {
		existing, _ := u.visitRepo.FindDuplicateVisit(req.EmployeeID, req.StoreID, req.CheckTime, 1440)
		if existing != nil {
			existing.CheckOutTime = &req.CheckTime
			existing.Type = "CHECKOUT"
			if existing.Status == "PENDING_TRANSACTION" {
				existing.Status = "NO_ORDER"
			}
			if req.Notes != "" {
				existing.Notes = req.Notes
			}
			// Evaluasi peringatan kunjungan berulang < 2 hari tanpa order sebelumnya
			// hanya jika kunjungan ini berakhir NO_ORDER dan bukan mencocokkan dirinya sendiri
			if existing.Status == "NO_ORDER" && req.StoreID != uuid.Nil && req.EmployeeID != uuid.Nil {
				u.flagFrequentVisitWithoutOrder(&existing.Notes, req.EmployeeID, req.StoreID, existing.CheckInTime, existing.ID)
			}
			err := u.visitRepo.Update(existing)
			if err == nil {
				u.publishVisitNotification(existing)
				go u.autoReconcileMTMerchandising(req.EmployeeID, req.StoreID, req.CheckTime, req.Notes, req.SelfieURL)
			}
			return err
		}
	} else if req.Type == "CHECKIN" || req.Type == "" {
		existing, _ := u.visitRepo.FindDuplicateVisit(req.EmployeeID, req.StoreID, req.CheckTime, 5)
		if existing != nil {
			fmt.Printf("DEBUG: Prevented duplicate visit sync for employee %v and store %v\n", req.EmployeeID, req.StoreID)
			return nil
		}
	}

	// Jika CHECKOUT berdiri sendiri (tanpa record CHECKIN terpisah sebelumnya)
	if visit.Type == "CHECKOUT" {
		if visit.Status == "PENDING_TRANSACTION" || visit.Status == "" {
			visit.Status = "NO_ORDER"
		}
		if visit.Status == "NO_ORDER" && req.StoreID != uuid.Nil && req.EmployeeID != uuid.Nil {
			u.flagFrequentVisitWithoutOrder(&visit.Notes, req.EmployeeID, req.StoreID, visit.CheckInTime, uuid.Nil)
		}
	}

	err := u.visitRepo.Create(visit)
	if err == nil {
		u.publishVisitNotification(visit)
		if req.Type == "CHECKOUT" {
			go u.autoReconcileMTMerchandising(req.EmployeeID, req.StoreID, req.CheckTime, req.Notes, req.SelfieURL)
		}
	}
	return err
}

func (u *salesUsecase) autoReconcileMTMerchandising(employeeID, storeID uuid.UUID, checkTime time.Time, notes, selfieURL string) {
	if !strings.Contains(notes, "Merchandising Selesai") || storeID == uuid.Nil {
		return
	}

	var store domain.Store
	if errStore := u.db.First(&store, "id = ?", storeID).Error; errStore != nil {
		return
	}

	var emp domain.Employee
	empName := "Sales MT"
	if errEmp := u.db.First(&emp, "id = ?", employeeID).Error; errEmp == nil {
		fullName := strings.TrimSpace(emp.FirstName + " " + emp.LastName)
		if fullName != "" {
			empName = fullName
		}
	}

	// Find or create KeyAccount
	var ka domain.KeyAccount
	// Keep the automatically reconciled audit in the same tenant as the
	// visited store. A global name lookup can select another company's Key
	// Account, after which the company-scoped merchandising list hides it.
	if errKA := u.db.Where("company_id = ? AND (LOWER(name) = LOWER(?) OR LOWER(company_name) = LOWER(?))", store.CompanyID, store.Name, store.Name).First(&ka).Error; errKA != nil {
		if errKA2 := u.db.Where("company_id = ?", store.CompanyID).First(&ka).Error; errKA2 != nil {
			ka = domain.KeyAccount{
				ID:          uuid.New(),
				CompanyID:   store.CompanyID,
				Name:        store.Name,
				CompanyName: store.Name,
				Type:        "NATIONAL",
				Status:      "ACTIVE",
			}
			_ = u.db.Create(&ka)
		}
	}

	// Check if already created today for this store
	todayStart := time.Date(checkTime.Year(), checkTime.Month(), checkTime.Day(), 0, 0, 0, 0, checkTime.Location())
	todayEnd := todayStart.Add(24 * time.Hour)
	var existingMD domain.MTMerchandising
	if errMD := u.db.Where("(store_name = ? OR supermarket_name = ?) AND visit_time >= ? AND visit_time < ?", store.Name, store.Name, todayStart, todayEnd).First(&existingMD).Error; errMD != nil {
		photo := selfieURL
		if photo == "" {
			photo = "https://images.unsplash.com/photo-1542838132-92c53300491e?auto=format&fit=crop&q=80&w=400"
		}
		newMD := domain.MTMerchandising{
			ID:                 uuid.New(),
			KeyAccountID:       ka.ID,
			MerchandiserName:   empName,
			MdName:             empName,
			StoreName:          store.Name,
			SupermarketName:    store.Name,
			VisitTime:          checkTime,
			PlanogramCompliant: true,
			ShelfSharePercent:  100.0,
			OutOfStockSkus:     "",
			PhotoURL:           photo,
			Status:             "COMPLIANT",
			StockUpdateLog:     notes,
		}
		_ = u.db.Create(&newMD)
	}
}

func (u *salesUsecase) GetVisitHistory(limit, offset int, employeeID, branchID *uuid.UUID, salesType string, search string, startDate, endDate *time.Time, summary bool) ([]domain.SalesVisit, int64, error) {
	visits, err := u.visitRepo.FindAllHistory(limit, offset, employeeID, branchID, salesType, search, startDate, endDate, summary)
	if err != nil {
		return nil, 0, err
	}
	if limit <= 0 || (offset == 0 && len(visits) < limit) {
		return visits, int64(len(visits)), nil
	}
	total, err := u.visitRepo.CountHistory(employeeID, branchID, salesType, search, startDate, endDate)
	if err != nil {
		return nil, 0, err
	}
	return visits, total, nil
}

func (u *salesUsecase) GetPendingVisits(employeeID uuid.UUID) ([]domain.SalesVisit, error) {
	var visits []domain.SalesVisit
	err := u.db.Preload("Store").Where("employee_id = ? AND status = ?", employeeID, "PENDING_TRANSACTION").Order("check_in_time DESC").Find(&visits).Error
	return visits, err
}

func (u *salesUsecase) CreateVisit(req RecordVisitRequest) (*domain.SalesVisit, error) {
	if req.ClientID != "" {
		var existingVisit domain.SalesVisit
		if err := u.db.Where("client_id = ?", req.ClientID).First(&existingVisit).Error; err == nil {
			return &existingVisit, nil
		}
	}

	// Evaluasi peringatan kunjungan berulang hanya jika dibuat langsung sebagai CHECKOUT
	if req.Type == "CHECKOUT" && req.StoreID != uuid.Nil && req.EmployeeID != uuid.Nil {
		u.flagFrequentVisitWithoutOrder(&req.Notes, req.EmployeeID, req.StoreID, req.CheckTime, uuid.Nil)
	}

	visit := &domain.SalesVisit{
		EmployeeID:  req.EmployeeID,
		StoreID:     req.StoreID,
		Latitude:    req.Latitude,
		Longitude:   req.Longitude,
		SelfieURL:   req.SelfieURL,
		Type:        req.Type,
		Status:      "PENDING_TRANSACTION",
		CheckInTime: req.CheckTime,
		Notes:       req.Notes,
	}
	if req.ClientID != "" {
		visit.ClientID = &req.ClientID
	}

	if req.Type == "CHECKIN" || req.Type == "" {
		existing, _ := u.visitRepo.FindDuplicateVisit(req.EmployeeID, req.StoreID, req.CheckTime, 5)
		if existing != nil {
			return existing, nil
		}
	}

	err := u.visitRepo.Create(visit)
	if err == nil {
		u.publishVisitNotification(visit)
	}
	return visit, err
}

func (u *salesUsecase) CreateBatchVisits(reqs []RecordVisitRequest) ([]*domain.SalesVisit, error) {
	var results []*domain.SalesVisit
	for _, req := range reqs {
		visit, err := u.CreateVisit(req)
		if err != nil {
			return results, err
		}
		if visit != nil {
			results = append(results, visit)
		}
	}
	return results, nil
}

// flagFrequentVisitWithoutOrder memeriksa apakah ada kunjungan SEBELUMNYA ke toko yang sama
// dalam kurun waktu < 2 hari kalender (hari yang sama atau H-1) yang TIDAK menghasilkan order/transaksi.
// Jika ya, sistem otomatis menyematkan catatan peringatan agar terpantau oleh supervisor.
func (u *salesUsecase) flagFrequentVisitWithoutOrder(notes *string, employeeID, storeID uuid.UUID, currentCheckInTime time.Time, excludeVisitID uuid.UUID) {
	if notes == nil || storeID == uuid.Nil || employeeID == uuid.Nil {
		return
	}
	flagNotice := "[SISTEM: Kunjungan berulang < 2 hari tanpa order sebelumnya]"
	if strings.Contains(*notes, "[SISTEM: Kunjungan berulang") {
		return
	}

	if currentCheckInTime.IsZero() {
		currentCheckInTime = time.Now()
	}

	// Hitung awal hari untuk kunjungan saat ini dalam zona waktu lokal
	loc := currentCheckInTime.Location()
	currYear, currMonth, currDay := currentCheckInTime.Date()
	currDate := time.Date(currYear, currMonth, currDay, 0, 0, 0, 0, loc)

	// Batas awal pencarian: 2 hari kalender sebelumnya (awal hari H-2)
	windowStart := currDate.AddDate(0, 0, -2)

	var lastVisit domain.SalesVisit
	query := u.db.Where("employee_id = ? AND store_id = ? AND check_in_time >= ? AND status != 'SKIPPED' AND type != 'SKIP'",
		employeeID, storeID, windowStart)

	if excludeVisitID != uuid.Nil {
		query = query.Where("id != ?", excludeVisitID)
	}

	// Kunjungan sebelumnya harus dimulai sebelum sesi kunjungan saat ini (buffer 5 menit untuk mencegah bentrok)
	query = query.Where("check_in_time < ?", currentCheckInTime.Add(-5*time.Minute))

	err := query.Order("check_in_time DESC").First(&lastVisit).Error
	if err != nil || lastVisit.ID == uuid.Nil {
		// Tidak ada kunjungan sebelumnya dalam rentang waktu tersebut
		return
	}

	// Hitung selisih hari kalender antara kunjungan saat ini dan kunjungan sebelumnya
	lastYear, lastMonth, lastDay := lastVisit.CheckInTime.Date()
	lastDate := time.Date(lastYear, lastMonth, lastDay, 0, 0, 0, 0, loc)
	daysDiff := int(currDate.Sub(lastDate) / (24 * time.Hour))

	// Jika selisih hari kalender >= 2 (misal Selasa ke Kamis, atau Senin ke Rabu),
	// ini adalah siklus kunjungan rutin normal (bukan < 2 hari)
	if daysDiff >= 2 {
		return
	}

	// Cek apakah kunjungan sebelumnya menghasilkan order / transaksi
	if lastVisit.Status == "COMPLETED" {
		return
	}

	var trxCount int64
	u.db.Model(&domain.SalesTransaction{}).
		Where("store_id = ? AND (visit_id = ? OR (transaction_date >= ? AND transaction_date <= ?))",
			storeID, lastVisit.ID, lastVisit.CheckInTime.Add(-1*time.Hour), lastVisit.CheckInTime.Add(24*time.Hour)).
		Count(&trxCount)
	if trxCount > 0 {
		return
	}

	var orderCount int64
	u.db.Model(&domain.SalesOrder{}).
		Where("store_id = ? AND (visit_id = ? OR (order_date >= ? AND order_date <= ?))",
			storeID, lastVisit.ID, lastVisit.CheckInTime.Add(-1*time.Hour), lastVisit.CheckInTime.Add(24*time.Hour)).
		Count(&orderCount)
	if orderCount > 0 {
		return
	}

	// Flag hanya jika kunjungan sebelumnya benar-benar tidak menghasilkan transaksi atau order
	cleanNotes := strings.TrimSpace(*notes)
	if cleanNotes != "" {
		*notes = flagNotice + " " + cleanNotes
	} else {
		*notes = flagNotice
	}
}

func (u *salesUsecase) UpdateVisit(id uuid.UUID, req RecordVisitRequest) error {
	visit, err := u.visitRepo.FindByID(id)
	if err != nil {
		return err
	}
	if visit == nil {
		return errors.New("visit not found")
	}

	visit.EmployeeID = req.EmployeeID
	visit.StoreID = req.StoreID
	visit.Latitude = req.Latitude
	visit.Longitude = req.Longitude
	if req.SelfieURL != "" {
		visit.SelfieURL = req.SelfieURL
	}
	visit.Type = req.Type
	visit.CheckInTime = req.CheckTime
	visit.Notes = req.Notes

	return u.visitRepo.Update(visit)
}

func (u *salesUsecase) DeleteVisit(id uuid.UUID) error {
	return u.visitRepo.Delete(id)
}

func (u *salesUsecase) RegeneratePayment(transactionID string) (*domain.SalesTransaction, error) {
	id, err := uuid.Parse(transactionID)
	if err != nil {
		return nil, errors.New("invalid transaction ID")
	}

	trx, err := u.salesRepo.FindByID(id)
	if err != nil || trx == nil {
		return nil, errors.New("transaction not found")
	}

	if trx.PaymentMethod != "QRIS" && trx.PaymentMethod != "VA" && trx.DPPaymentMethod != "QRIS" && trx.DPPaymentMethod != "VA" {
		return nil, errors.New("payment method does not support regeneration")
	}

	// Generate a new Midtrans order ID by appending -R<timestamp> to the original ReceiptNo
	newOrderID := fmt.Sprintf("%s-R%d", trx.ReceiptNo, time.Now().Unix())

	var midtransResp *coreapi.ChargeResponse

	// Determine the amount to charge
	amount := int64(trx.TotalAmount)
	if trx.PaymentMethod == "TEMPO" && trx.DPAmount > 0 {
		amount = int64(trx.DPAmount)
	}

	activePaymentMethod := trx.PaymentMethod
	if trx.PaymentMethod == "TEMPO" {
		activePaymentMethod = trx.DPPaymentMethod
	}

	if activePaymentMethod == "QRIS" && u.midtransClient != nil {
		resp, err := u.midtransClient.CreateQRIS(newOrderID, amount)
		if err != nil {
			return nil, fmt.Errorf("failed to regenerate QRIS: %v", err)
		}
		midtransResp = resp
	} else if activePaymentMethod == "VA" && u.midtransClient != nil {
		bankStr := "bca" // Default fallback
		if trx.MidtransBank != nil && *trx.MidtransBank != "" {
			bankStr = *trx.MidtransBank
		}
		resp, err := u.midtransClient.CreateBankTransfer(newOrderID, amount, bankStr)
		if err != nil {
			return nil, fmt.Errorf("failed to regenerate VA: %v", err)
		}
		midtransResp = resp
	}

	if midtransResp != nil {
		if activePaymentMethod == "QRIS" {
			qrisData := midtransPkg.ExtractQRISData(midtransResp)
			if qrisData != "" {
				trx.MidtransQRISURL = qrisData
			}
		} else if activePaymentMethod == "VA" {
			if len(midtransResp.VaNumbers) > 0 {
				trx.MidtransVANumber = &midtransResp.VaNumbers[0].VANumber
				trx.MidtransBank = &midtransResp.VaNumbers[0].Bank
			} else if midtransResp.BillerCode != "" && midtransResp.BillKey != "" {
				trx.MidtransBillerCode = &midtransResp.BillerCode
				trx.MidtransBillKey = &midtransResp.BillKey
				mandiriStr := "mandiri"
				trx.MidtransBank = &mandiriStr
			}
		}

		if err := u.salesRepo.UpdateTransactionOnly(trx); err != nil {
			return nil, err
		}
	}

	return trx, nil
}
func (u *salesUsecase) GetSalesKasbon(companyID uuid.UUID) ([]map[string]interface{}, error) {
	return u.salesRepo.GetSalesKasbon(companyID)
}

func (u *salesUsecase) GetSettlementQueue() ([]domain.SalesTransaction, error) {
	return u.salesRepo.GetSettlementQueue()
}

func (u *salesUsecase) CreateManualVisitPlan(employeeID uuid.UUID, stores []StorePlanRequest, planDate time.Time) error {
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err == nil {
		planDate = planDate.In(loc)
	}

	existingPlans, err := u.visitPlanRepo.GetByEmployeeAndDate(employeeID, planDate)
	if err != nil {
		return err
	}

	// 1. Map all requested (checked) stores
	requestedStoreMap := make(map[uuid.UUID]bool)
	for _, req := range stores {
		storeID := req.StoreID
		if storeID == uuid.Nil {
			storeID = req.StoreIDCamel
		}
		if storeID != uuid.Nil {
			requestedStoreMap[storeID] = true
		}
	}

	// 2. Identify stores that were PREVIOUSLY planned today, but are NO LONGER requested
	currentPlans, _ := u.GetVisitPlan(employeeID, planDate)
	for _, currentPlan := range currentPlans {
		if !requestedStoreMap[currentPlan.StoreID] {
			// Delete them (removes from manual DB and/or creates SKIPPED record)
			_ = u.DeleteSpecificPlan(employeeID, currentPlan.StoreID, planDate)
		}
	}

	// 3. For all requested stores, make sure any SKIPPED visit is deleted so they reappear if they were previously deleted
	startOfDay := time.Date(planDate.Year(), planDate.Month(), planDate.Day(), 0, 0, 0, 0, planDate.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)
	for storeID := range requestedStoreMap {
		u.db.Where("employee_id = ? AND store_id = ? AND status = 'SKIPPED' AND check_in_time >= ? AND check_in_time < ?", employeeID, storeID, startOfDay, endOfDay).Delete(&domain.SalesVisit{})
	}

	existingMap := make(map[uuid.UUID]*domain.SalesVisitPlan)
	for i, p := range existingPlans {
		existingMap[p.StoreID] = &existingPlans[i]
	}

	for _, storeReq := range stores {
		storeID := storeReq.StoreID
		if storeID == uuid.Nil {
			storeID = storeReq.StoreIDCamel
		}
		if storeID == uuid.Nil {
			continue // Skip gracefully to avoid FK constraint errors for old apps
		}

		projections := make([]domain.SalesVisitPlanProjection, 0)
		for _, proj := range storeReq.Projections {
			projections = append(projections, domain.SalesVisitPlanProjection{
				ProductID: proj.ProductID,
				Quantity:  proj.Quantity,
			})
		}

		if existingPlan, exists := existingMap[storeID]; exists {
			// Update existing plan projections
			if err := u.visitPlanRepo.UpdateProjections(existingPlan.ID, projections); err != nil {
				return err
			}
			continue
		}

		plan := &domain.SalesVisitPlan{
			EmployeeID:  employeeID,
			StoreID:     storeID,
			PlanDate:    planDate,
			Projections: projections,
		}
		if err := u.visitPlanRepo.Create(plan); err != nil {
			// Gracefully handle duplicate key — treat as "already exists" and skip.
			// This preserves backward compatibility with older mobile app versions that
			// may re-send the same store IDs after a partial failure or network retry.
			if strings.Contains(err.Error(), "duplicate key") ||
				strings.Contains(err.Error(), "unique constraint") ||
				strings.Contains(err.Error(), "idx_visit_plan_unique") {
				// Since we now update existing, this duplicate key is a rare race condition. Just ignore.
				continue
			}
			return err
		}
	}
	return nil
}

func (u *salesUsecase) DeleteSpecificPlan(employeeID uuid.UUID, storeID uuid.UUID, planDate time.Time) error {
	err := u.visitPlanRepo.DeleteSpecificPlan(employeeID, storeID, planDate)
	if err != nil {
		return err
	}

	loc, _ := time.LoadLocation("Asia/Jakarta")
	if loc != nil {
		planDate = planDate.In(loc)
	}

	// Cek apakah sudah ada kunjungan nyata di hari ini
	startOfDay := time.Date(planDate.Year(), planDate.Month(), planDate.Day(), 0, 0, 0, 0, planDate.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)
	existingVisits, errHistory := u.visitRepo.FindAllHistory(100, 0, &employeeID, nil, "", "", &startOfDay, &endOfDay, true)

	if errHistory == nil {
		for _, v := range existingVisits {
			if v.StoreID == storeID && v.Status != "SKIPPED" {
				// Sudah ada kunjungan asli. Jangan timpa dengan SKIPPED.
				return nil
			}
		}
	}

	// Gunakan planDate jam 12 siang agar selalu jatuh tepat pada tanggal rencana kunjungan yang dituju
	skippedTime := time.Date(planDate.Year(), planDate.Month(), planDate.Day(), 12, 0, 0, 0, planDate.Location())

	visit := domain.SalesVisit{
		EmployeeID:   employeeID,
		StoreID:      storeID,
		CheckInTime:  skippedTime,
		CheckOutTime: &skippedTime,
		Type:         "SKIP",
		Status:       "SKIPPED",
		Notes:        "Deleted from schedule",
	}
	err = u.visitRepo.Create(&visit)
	if err != nil {
		fmt.Println("Error creating skipped visit:", err)
		return err
	}

	return nil
}

func (u *salesUsecase) GetVisitPlanBuilder(employeeID uuid.UUID, targetSalesmanID *uuid.UUID, branchID *uuid.UUID) ([]VisitPlanItem, error) {
	var filterEmpID *uuid.UUID = &employeeID

	// If targetSalesmanID is provided from the frontend, directly use targetSalesmanID
	if targetSalesmanID != nil && *targetSalesmanID != uuid.Nil {
		filterEmpID = targetSalesmanID
	} else {
		// Check if caller is management
		emp, _ := u.employeeRepo.FindByID(employeeID)
		isManagement := false
		if emp != nil && emp.JobPosition != nil {
			title := strings.ToLower(emp.JobPosition.Title)
			if strings.Contains(title, "direktur") || strings.Contains(title, "director") ||
				strings.Contains(title, "manager") || strings.Contains(title, "owner") ||
				strings.Contains(title, "superadmin") || strings.Contains(title, "spv") ||
				strings.Contains(title, "supervisor") {
				isManagement = true
			}
		}

		if isManagement {
			if branchID == nil && emp != nil && emp.BranchID != nil {
				branchID = emp.BranchID
			}
			if branchID == nil {
				return []VisitPlanItem{}, nil
			}
			filterEmpID = nil
		}
	}

	// Ambil stores tanpa subquery berat karena histories diambil terpisah via GetStoreHistories
	stores, _, err := u.storeRepo.FindPaginated(nil, branchID, filterEmpID, "", "", "ACTIVE", "", 1000, 0, "", "", false)
	if err != nil {
		return nil, err
	}
	var storeIDs []uuid.UUID
	for _, s := range stores {
		storeIDs = append(storeIDs, s.ID)
	}
	storeHistories, _ := u.storeRepo.GetStoreHistories(storeIDs)
	var result []VisitPlanItem
	for _, store := range stores {
		hist := storeHistories[store.ID]
		isPriority := false
		if hist.LastVisitDate == nil {
			isPriority = true
		} else {
			daysSinceVisit := time.Since(*hist.LastVisitDate).Hours() / 24
			if daysSinceVisit > 7 {
				isPriority = true
			}
		}
		empName := ""
		if store.AssignedEmployee != nil {
			empName = strings.TrimSpace(store.AssignedEmployee.FirstName + " " + store.AssignedEmployee.LastName)
		}
		result = append(result, VisitPlanItem{
			StoreID:               store.ID,
			StoreName:             store.Name,
			Address:               store.Address,
			Status:                "PLANNED",
			IsExtra:               true,
			LastVisitDate:         hist.LastVisitDate,
			LastTransactionDate:   hist.LastTransactionDate,
			LastTransactionAmount: hist.LastTransactionAmount,
			LastTransactionItems:  hist.LastTransactionItems,
			IsPriority:            isPriority,
			Latitude:              store.Latitude,
			Longitude:             store.Longitude,
			AssignedEmployeeID:    store.AssignedEmployeeID,
			AssignedEmployeeName:  empName,
		})
	}
	return result, nil
}

func (u *salesUsecase) GetTopProducts(companyID uuid.UUID, limit int) ([]map[string]interface{}, error) {
	return u.salesRepo.GetTopProducts(companyID, limit)
}

func (u *salesUsecase) publishVisitNotification(visit *domain.SalesVisit) {
	if u.sseHub == nil {
		return
	}
	go func() {
		emp, err := u.employeeRepo.FindByID(visit.EmployeeID)
		if err != nil || emp == nil {
			return
		}
		st, err := u.storeRepo.FindByID(visit.StoreID)
		if err != nil || st == nil {
			return
		}

		title := "Sales Check-In"
		msg := fmt.Sprintf("Salesman %s %s melakukan check-in di toko %s", emp.FirstName, emp.LastName, st.Name)
		if visit.Type == "CHECKOUT" {
			title = "Sales Check-Out"
			msg = fmt.Sprintf("Salesman %s %s melakukan check-out dari toko %s", emp.FirstName, emp.LastName, st.Name)
		}

		event := sse.NotificationEvent{
			ID:        visit.ID.String(),
			Type:      "sales_visit",
			Title:     title,
			Message:   msg,
			CompanyID: emp.CompanyID.String(),
			CreatedAt: time.Now().Format("2006-01-02 15:04:05"),
			Data: map[string]interface{}{
				"visit_id":   visit.ID.String(),
				"salesman":   fmt.Sprintf("%s %s", emp.FirstName, emp.LastName),
				"store_name": st.Name,
				"type":       visit.Type,
			},
		}
		_ = u.sseHub.PublishNotification(context.Background(), event)
	}()
}

func (u *salesUsecase) GetEmployeeProductRecap(employeeID uuid.UUID, month, year int) ([]domain.ProductRecap, error) {
	return u.salesRepo.GetProductRecapByEmployeeAndPeriod(employeeID, month, year)
}

func (u *salesUsecase) GetNooEcTracking(startDate, endDate time.Time, branchID, employeeID *uuid.UUID, salesType string) (*domain.NooEcTrackingResponse, error) {
	start := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, startDate.Location())
	end := time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 999999999, endDate.Location())

	// 1. Ambil daftar Salesman yang relevan (hanya jabatan sales/marketing dan aktif)
	salesEmpQuery := u.db.Table("employees e").
		Select(`
			e.id, 
			COALESCE(e.employee_id_number, '-') as employee_id_number,
			e.first_name, 
			e.last_name, 
			jp.title as job_title, 
			COALESCE(b.name, '-') as branch_name,
			COALESCE(ob.name, '') as operating_branch_name,
			COALESCE(sk.working_territory, '') as working_territory
		`).
		Joins("JOIN job_positions jp ON jp.id = e.job_position_id").
		Joins("LEFT JOIN branches b ON b.id = e.branch_id").
		Joins("LEFT JOIN branches ob ON ob.id = e.operating_branch_id").
		Joins(`LEFT JOIN LATERAL (
			SELECT working_territory FROM sales_kpis 
			WHERE employee_id = e.id AND working_territory IS NOT NULL AND TRIM(working_territory) != '' 
			ORDER BY (period_year = ? AND period_month = ?) DESC, created_at DESC LIMIT 1
		) sk ON true`, startDate.Year(), int(startDate.Month())).
		Where("(e.employment_status IS NULL OR e.employment_status NOT IN ('Inactive', 'INACTIVE', 'Resigned', 'RESIGNED', 'Terminated', 'TERMINATED'))").
		Where("e.first_name NOT ILIKE '%resign%' AND e.last_name NOT ILIKE '%resign%' AND e.first_name NOT ILIKE '%issued out%' AND e.last_name NOT ILIKE '%issued out%'").
		Where("(jp.title ILIKE '%sales%' OR jp.title ILIKE '%motoris%' OR jp.title ILIKE '%booster%' OR jp.title ILIKE '%take order%' OR jp.title ILIKE '%task order%' OR jp.title ~* '\\mto\\M' OR jp.title ~* '\\mmt\\M' OR jp.title ILIKE '%modern trade%' OR jp.title ILIKE '%kam%' OR jp.title ILIKE '%key account%')").
		Where("jp.title NOT ILIKE '%operator%' AND jp.title NOT ILIKE '%driver%' AND jp.title NOT ILIKE '%auditor%' AND jp.title NOT ILIKE '%editor%' AND jp.title NOT ILIKE '%customer service%' AND jp.title NOT ILIKE '%validator%' AND jp.title NOT ILIKE '%gudang%' AND jp.title NOT ILIKE '%admin%' AND jp.title NOT ILIKE '%security%'")

	if branchID != nil && *branchID != uuid.Nil {
		salesEmpQuery = salesEmpQuery.Where("e.branch_id = ?", *branchID)
	}
	if employeeID != nil && *employeeID != uuid.Nil {
		salesEmpQuery = salesEmpQuery.Where("e.id = ?", *employeeID)
	}
	if salesType != "" && salesType != "ALL" && salesType != "ALL_SALES" {
		if salesType == "sales_mt" {
			salesEmpQuery = salesEmpQuery.Where("jp.title ILIKE '%sales mt%' OR jp.title ILIKE '%modern trade%' OR jp.title ILIKE '%kam%' OR jp.title ILIKE '%key account%' OR jp.title ~* '\\mmt\\M'")
		} else if salesType == "sales_to" {
			salesEmpQuery = salesEmpQuery.Where("jp.title ILIKE '%take order%' OR jp.title ILIKE '%task order%' OR jp.title ILIKE '%sales to%' OR jp.title ~* '\\mto\\M'")
		} else if salesType == "sales_motoris_booster" {
			salesEmpQuery = salesEmpQuery.Where("jp.title ILIKE '%motoris%' OR jp.title ILIKE '%booster%'")
		}
	}

	type salesEmpRow struct {
		ID                  uuid.UUID
		EmployeeIDNumber    string
		FirstName           string
		LastName            string
		JobTitle            string
		BranchName          string
		OperatingBranchName string
		WorkingTerritory    string
	}
	var salesEmployees []salesEmpRow
	if err := salesEmpQuery.Order("b.name ASC, e.first_name ASC").Scan(&salesEmployees).Error; err != nil {
		return nil, err
	}

	// 2. Ambil Matriks Performa per Salesman
	salesmanMatrix := make([]domain.SalesmanNooEcSummary, 0, len(salesEmployees))
	var totalOverallCalls int
	var totalOverallNoo int

	for _, emp := range salesEmployees {
		empName := strings.TrimSpace(emp.FirstName + " " + emp.LastName)

		salesArea := strings.TrimSpace(emp.WorkingTerritory)
		if salesArea == "" || salesArea == "-" {
			if emp.OperatingBranchName != "" && emp.OperatingBranchName != "-" {
				salesArea = emp.OperatingBranchName
			} else if emp.BranchName != "" && emp.BranchName != "-" {
				salesArea = emp.BranchName
			} else {
				salesArea = "Pusat"
			}
		}

		// Total Toko yang di-assign ke salesman ini
		var totalAssigned int64
		_ = u.db.Table("stores").Where("assigned_employee_id = ?", emp.ID).Count(&totalAssigned).Error

		// Total Kunjungan (Visits / Call) pada periode
		var totalVisits int64
		_ = u.db.Table("sales_visits").
			Where("employee_id = ? AND check_in_time BETWEEN ? AND ?", emp.ID, start, end).
			Count(&totalVisits).Error

		// NOO (Toko Baru yang dibuat pada periode)
		var nooCount int64
		_ = u.db.Table("stores").
			Where("assigned_employee_id = ? AND created_at BETWEEN ? AND ?", emp.ID, start, end).
			Count(&nooCount).Error

		// Effective Calls: Kunjungan atau toko dengan order/transaksi pada periode untuk salesman ini
		type ecAggRow struct {
			EffectiveStoreCount int64
			TotalOrderCount     int64
			TotalAmount         float64
		}
		var ecAgg ecAggRow
		_ = u.db.Raw(`
			SELECT 
				COUNT(DISTINCT store_id) as effective_store_count,
				COUNT(id) as total_order_count,
				COALESCE(SUM(total_amount), 0) as total_amount
			FROM (
				SELECT id, store_id, total_amount FROM sales_orders 
				WHERE employee_id = ? AND created_at BETWEEN ? AND ? AND status NOT IN ('CANCELLED', 'REJECTED', 'DRAFT')
				UNION ALL
				SELECT id, store_id, total_amount FROM sales_transactions 
				WHERE employee_id = ? AND transaction_date BETWEEN ? AND ? AND status NOT IN ('CANCELLED', 'REJECTED')
			) combined_orders
		`, emp.ID, start, end, emp.ID, start, end).Scan(&ecAgg).Error

		var ecRatio float64
		if totalVisits > 0 {
			ecRatio = (float64(ecAgg.EffectiveStoreCount) / float64(totalVisits)) * 100.0
			if ecRatio > 100 {
				ecRatio = 100
			}
		} else if totalAssigned > 0 && ecAgg.EffectiveStoreCount > 0 {
			ecRatio = (float64(ecAgg.EffectiveStoreCount) / float64(totalAssigned)) * 100.0
		}

		salesmanMatrix = append(salesmanMatrix, domain.SalesmanNooEcSummary{
			EmployeeID:          emp.ID,
			EmployeeIDNumber:    emp.EmployeeIDNumber,
			EmployeeName:        empName,
			JobTitle:            emp.JobTitle,
			BranchName:          emp.BranchName,
			OperatingBranchName: emp.OperatingBranchName,
			SalesArea:           salesArea,
			TotalAssignedStores: int(totalAssigned),
			TotalVisits:         int(totalVisits),
			EffectiveCalls:      int(ecAgg.EffectiveStoreCount),
			EcRatio:             float64(int(ecRatio*10)) / 10,
			NooCount:            int(nooCount),
			TotalOrderCount:     int(ecAgg.TotalOrderCount),
			TotalOrderAmount:    ecAgg.TotalAmount,
		})

		totalOverallCalls += int(totalVisits)
		totalOverallNoo += int(nooCount)
	}

	// 3. Ambil Log Detail Toko pada Periode Ini (Toko yang di-NOO, atau dikunjungi, atau ada order pada periode)
	var storeLogs []domain.StoreNooEcDetail

	storeQuery := `
		SELECT 
			s.id as store_id,
			s.name as store_name,
			COALESCE(s.owner_name, '-') as owner_name,
			COALESCE(s.phone_number, '-') as phone_number,
			COALESCE(s.address, '-') as address,
			COALESCE(s.regency, '') as city,
			COALESCE(s.district, '') as district,
			s.created_at,
			(s.created_at BETWEEN ? AND ?) as is_noo_in_period,
			s.assigned_employee_id as salesman_id,
			COALESCE(e.employee_id_number, '-') as salesman_id_number,
			COALESCE(TRIM(CONCAT(e.first_name, ' ', e.last_name)), 'Belum Ditugaskan') as salesman_name,
			COALESCE(
				NULLIF(TRIM(sk.working_territory), ''),
				NULLIF(TRIM(ob.name), ''),
				b.name,
				'Area Belum Diatur'
			) as salesman_area,
			s.latitude,
			s.longitude,
			sv_agg.last_visit_time,
			(sv_agg.visit_count > 0) as is_visited_in_period,
			COALESCE(ord_agg.order_count, 0) as order_count,
			COALESCE(ord_agg.total_amount, 0) as order_amount,
			COALESCE(ord_agg.latest_so_number, '-') as latest_order_number,
			(COALESCE(ord_agg.order_count, 0) > 0) as is_effective_call
		FROM stores s
		LEFT JOIN employees e ON e.id = s.assigned_employee_id
		LEFT JOIN branches b ON b.id = e.branch_id
		LEFT JOIN branches ob ON ob.id = e.operating_branch_id
		LEFT JOIN LATERAL (
			SELECT working_territory FROM sales_kpis 
			WHERE employee_id = e.id AND working_territory IS NOT NULL AND TRIM(working_territory) != '' 
			ORDER BY (period_year = ? AND period_month = ?) DESC, created_at DESC LIMIT 1
		) sk ON true
		LEFT JOIN LATERAL (
			SELECT 
				COUNT(*) as visit_count,
				MAX(check_in_time) as last_visit_time
			FROM sales_visits sv
			WHERE sv.store_id = s.id AND sv.check_in_time BETWEEN ? AND ?
		) sv_agg ON true
		LEFT JOIN LATERAL (
			SELECT 
				COUNT(*) as order_count,
				SUM(total_amount) as total_amount,
				MAX(doc_no) as latest_so_number
			FROM (
				SELECT so_number as doc_no, total_amount FROM sales_orders 
				WHERE store_id = s.id AND created_at BETWEEN ? AND ? AND status NOT IN ('CANCELLED', 'REJECTED', 'DRAFT')
				UNION ALL
				SELECT receipt_no as doc_no, total_amount FROM sales_transactions 
				WHERE store_id = s.id AND transaction_date BETWEEN ? AND ? AND status NOT IN ('CANCELLED', 'REJECTED')
			) o
		) ord_agg ON true
		WHERE s.is_active = true AND (
			(s.created_at BETWEEN ? AND ?)
			OR (sv_agg.visit_count > 0)
			OR (COALESCE(ord_agg.order_count, 0) > 0)
		)
	`
	var storeArgs []interface{}
	storeArgs = append(storeArgs, start, end, startDate.Year(), int(startDate.Month()), start, end, start, end, start, end, start, end)

	if branchID != nil && *branchID != uuid.Nil {
		storeQuery += " AND (COALESCE(e.operating_branch_id, e.branch_id) = ? OR s.assigned_employee_id IN (SELECT id FROM employees WHERE COALESCE(operating_branch_id, branch_id) = ?))"
		storeArgs = append(storeArgs, *branchID, *branchID)
	}
	if employeeID != nil && *employeeID != uuid.Nil {
		storeQuery += " AND s.assigned_employee_id = ?"
		storeArgs = append(storeArgs, *employeeID)
	}

	storeQuery += " ORDER BY s.created_at DESC, ord_agg.total_amount DESC LIMIT 500"

	if err := u.db.Raw(storeQuery, storeArgs...).Scan(&storeLogs).Error; err != nil {
		return nil, err
	}

	// 4. Hitung Executive Summary
	var totalOverallEc int
	var totalOverallOrderAmount float64
	for _, sm := range salesmanMatrix {
		totalOverallEc += sm.EffectiveCalls
		totalOverallOrderAmount += sm.TotalOrderAmount
	}

	var overallEcRatio float64
	if totalOverallCalls > 0 {
		overallEcRatio = (float64(totalOverallEc) / float64(totalOverallCalls)) * 100.0
		if overallEcRatio > 100 {
			overallEcRatio = 100
		}
	}

	return &domain.NooEcTrackingResponse{
		StartDate:        startDate.Format("2006-01-02"),
		EndDate:          endDate.Format("2006-01-02"),
		TotalCalls:       totalOverallCalls,
		TotalEc:          totalOverallEc,
		EcRatio:          float64(int(overallEcRatio*10)) / 10,
		TotalNoo:         totalOverallNoo,
		TotalOrderAmount: totalOverallOrderAmount,
		SalesmanMatrix:   salesmanMatrix,
		StoreLogs:        storeLogs,
	}, nil
}

func (u *salesUsecase) ExportNooEcTrackingExcel(startDate, endDate time.Time, branchID, employeeID *uuid.UUID, salesType string) ([]byte, error) {
	data, err := u.GetNooEcTracking(startDate, endDate, branchID, employeeID, salesType)
	if err != nil {
		return nil, err
	}

	f := excelize.NewFile()
	defer f.Close()

	// Sheet 1: Performa Salesman
	sheet1 := "Performa Salesman"
	f.SetSheetName("Sheet1", sheet1)

	// Styles
	borderThin := []excelize.Border{
		{Type: "top", Color: "D0D5DD", Style: 1},
		{Type: "bottom", Color: "D0D5DD", Style: 1},
		{Type: "left", Color: "D0D5DD", Style: 1},
		{Type: "right", Color: "D0D5DD", Style: 1},
	}
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"1E293B"}, Pattern: 1},
		Font:      &excelize.Font{Bold: true, Color: "FFFFFF", Size: 10},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
		Border:    borderThin,
	})
	cellStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: 10, Color: "1E293B"},
		Alignment: &excelize.Alignment{Vertical: "center"},
		Border:    borderThin,
	})
	centerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: 10, Color: "1E293B"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Border:    borderThin,
	})
	currencyStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 10, Color: "0F766E"},
		Alignment: &excelize.Alignment{Horizontal: "right", Vertical: "center"},
		Border:    borderThin,
		NumFmt:    3,
	})

	// Header Sheet 1
	f.SetCellValue(sheet1, "A1", fmt.Sprintf("LAPORAN TRACKING NOO & EC (PERIODE: %s s/d %s)", data.StartDate, data.EndDate))
	f.SetRowHeight(sheet1, 1, 24)

	headers1 := []string{"No.", "ID / NIK Sales", "Nama Salesman", "Area Sales Bekerja", "Jabatan", "Cabang Penempatan", "Total Toko", "Total Kunjungan (Call)", "Effective Call (EC)", "Rasio EC (%)", "Toko Baru (NOO)", "Total Omzet Pesanan (Rp)"}
	for colIdx, h := range headers1 {
		cell, _ := excelize.CoordinatesToCellName(colIdx+1, 3)
		f.SetCellValue(sheet1, cell, h)
		f.SetCellStyle(sheet1, cell, cell, headerStyle)
	}

	for i, sm := range data.SalesmanMatrix {
		row := i + 4
		f.SetCellValue(sheet1, fmt.Sprintf("A%d", row), i+1)
		f.SetCellStyle(sheet1, fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), centerStyle)
		f.SetCellValue(sheet1, fmt.Sprintf("B%d", row), sm.EmployeeIDNumber)
		f.SetCellStyle(sheet1, fmt.Sprintf("B%d", row), fmt.Sprintf("B%d", row), centerStyle)
		f.SetCellValue(sheet1, fmt.Sprintf("C%d", row), sm.EmployeeName)
		f.SetCellStyle(sheet1, fmt.Sprintf("C%d", row), fmt.Sprintf("C%d", row), cellStyle)
		f.SetCellValue(sheet1, fmt.Sprintf("D%d", row), sm.SalesArea)
		f.SetCellStyle(sheet1, fmt.Sprintf("D%d", row), fmt.Sprintf("D%d", row), cellStyle)
		f.SetCellValue(sheet1, fmt.Sprintf("E%d", row), sm.JobTitle)
		f.SetCellStyle(sheet1, fmt.Sprintf("E%d", row), fmt.Sprintf("E%d", row), cellStyle)
		f.SetCellValue(sheet1, fmt.Sprintf("F%d", row), sm.BranchName)
		f.SetCellStyle(sheet1, fmt.Sprintf("F%d", row), fmt.Sprintf("F%d", row), cellStyle)
		f.SetCellValue(sheet1, fmt.Sprintf("G%d", row), sm.TotalAssignedStores)
		f.SetCellStyle(sheet1, fmt.Sprintf("G%d", row), fmt.Sprintf("G%d", row), centerStyle)
		f.SetCellValue(sheet1, fmt.Sprintf("H%d", row), sm.TotalVisits)
		f.SetCellStyle(sheet1, fmt.Sprintf("H%d", row), fmt.Sprintf("H%d", row), centerStyle)
		f.SetCellValue(sheet1, fmt.Sprintf("I%d", row), sm.EffectiveCalls)
		f.SetCellStyle(sheet1, fmt.Sprintf("I%d", row), fmt.Sprintf("I%d", row), centerStyle)
		f.SetCellValue(sheet1, fmt.Sprintf("J%d", row), fmt.Sprintf("%.1f%%", sm.EcRatio))
		f.SetCellStyle(sheet1, fmt.Sprintf("J%d", row), fmt.Sprintf("J%d", row), centerStyle)
		f.SetCellValue(sheet1, fmt.Sprintf("K%d", row), sm.NooCount)
		f.SetCellStyle(sheet1, fmt.Sprintf("K%d", row), fmt.Sprintf("K%d", row), centerStyle)
		f.SetCellValue(sheet1, fmt.Sprintf("L%d", row), sm.TotalOrderAmount)
		f.SetCellStyle(sheet1, fmt.Sprintf("L%d", row), fmt.Sprintf("L%d", row), currencyStyle)
	}

	// Sheet 2: Log Outlet & Order
	sheet2 := "Log Toko & Pesanan"
	_, _ = f.NewSheet(sheet2)

	f.SetCellValue(sheet2, "A1", fmt.Sprintf("DETAIL AKTIVITAS OUTLET, NOO & EC (PERIODE: %s s/d %s)", data.StartDate, data.EndDate))
	f.SetRowHeight(sheet2, 1, 24)

	headers2 := []string{"No.", "ID Toko", "Nama Toko", "Pemilik", "No. Telepon", "Waktu Registrasi Toko", "Status NOO", "ID / NIK Sales", "Salesman PIC", "Area Sales Bekerja", "Status Kunjungan", "Status EC", "No. Pesanan Terakhir", "Nilai Pesanan (Rp)", "Kota / Wilayah", "Alamat"}
	for colIdx, h := range headers2 {
		cell, _ := excelize.CoordinatesToCellName(colIdx+1, 3)
		f.SetCellValue(sheet2, cell, h)
		f.SetCellStyle(sheet2, cell, cell, headerStyle)
	}

	for i, st := range data.StoreLogs {
		row := i + 4
		nooStr := "Bukan NOO"
		if st.IsNooInPeriod {
			nooStr = "NOO (Toko Baru)"
		}

		visitStr := "Belum Dikunjungi"
		if st.IsVisitedInPeriod && st.LastVisitTime != nil {
			visitStr = st.LastVisitTime.Format("02/01/2006 15:04")
		}

		ecStr := "Non-EC"
		if st.IsEffectiveCall {
			ecStr = fmt.Sprintf("EC (%d Order)", st.OrderCount)
		}

		f.SetCellValue(sheet2, fmt.Sprintf("A%d", row), i+1)
		f.SetCellStyle(sheet2, fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), centerStyle)
		f.SetCellValue(sheet2, fmt.Sprintf("B%d", row), st.StoreID.String())
		f.SetCellStyle(sheet2, fmt.Sprintf("B%d", row), fmt.Sprintf("B%d", row), centerStyle)
		f.SetCellValue(sheet2, fmt.Sprintf("C%d", row), st.StoreName)
		f.SetCellStyle(sheet2, fmt.Sprintf("C%d", row), fmt.Sprintf("C%d", row), cellStyle)
		f.SetCellValue(sheet2, fmt.Sprintf("D%d", row), st.OwnerName)
		f.SetCellStyle(sheet2, fmt.Sprintf("D%d", row), fmt.Sprintf("D%d", row), cellStyle)
		f.SetCellValue(sheet2, fmt.Sprintf("E%d", row), st.PhoneNumber)
		f.SetCellStyle(sheet2, fmt.Sprintf("E%d", row), fmt.Sprintf("E%d", row), cellStyle)
		f.SetCellValue(sheet2, fmt.Sprintf("F%d", row), st.CreatedAt.Format("02/01/2006 15:04 WIB"))
		f.SetCellStyle(sheet2, fmt.Sprintf("F%d", row), fmt.Sprintf("F%d", row), centerStyle)
		f.SetCellValue(sheet2, fmt.Sprintf("G%d", row), nooStr)
		f.SetCellStyle(sheet2, fmt.Sprintf("G%d", row), fmt.Sprintf("G%d", row), centerStyle)
		f.SetCellValue(sheet2, fmt.Sprintf("H%d", row), st.SalesmanIDNumber)
		f.SetCellStyle(sheet2, fmt.Sprintf("H%d", row), fmt.Sprintf("H%d", row), centerStyle)
		f.SetCellValue(sheet2, fmt.Sprintf("I%d", row), st.SalesmanName)
		f.SetCellStyle(sheet2, fmt.Sprintf("I%d", row), fmt.Sprintf("I%d", row), cellStyle)
		f.SetCellValue(sheet2, fmt.Sprintf("J%d", row), st.SalesmanArea)
		f.SetCellStyle(sheet2, fmt.Sprintf("J%d", row), fmt.Sprintf("J%d", row), cellStyle)
		f.SetCellValue(sheet2, fmt.Sprintf("K%d", row), visitStr)
		f.SetCellStyle(sheet2, fmt.Sprintf("K%d", row), fmt.Sprintf("K%d", row), centerStyle)
		f.SetCellValue(sheet2, fmt.Sprintf("L%d", row), ecStr)
		f.SetCellStyle(sheet2, fmt.Sprintf("L%d", row), fmt.Sprintf("L%d", row), centerStyle)
		f.SetCellValue(sheet2, fmt.Sprintf("M%d", row), st.LatestOrderNumber)
		f.SetCellStyle(sheet2, fmt.Sprintf("M%d", row), fmt.Sprintf("M%d", row), centerStyle)
		f.SetCellValue(sheet2, fmt.Sprintf("N%d", row), st.OrderAmount)
		f.SetCellStyle(sheet2, fmt.Sprintf("N%d", row), fmt.Sprintf("N%d", row), currencyStyle)
		f.SetCellValue(sheet2, fmt.Sprintf("O%d", row), st.City)
		f.SetCellStyle(sheet2, fmt.Sprintf("O%d", row), fmt.Sprintf("O%d", row), cellStyle)
		f.SetCellValue(sheet2, fmt.Sprintf("P%d", row), st.Address)
		f.SetCellStyle(sheet2, fmt.Sprintf("P%d", row), fmt.Sprintf("P%d", row), cellStyle)
	}

	var b bytes.Buffer
	if err := f.Write(&b); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

// Regional normalization logic and Indonesian administrative region dictionary
// are maintained in region_dictionary.go

func (u *salesUsecase) GetCityRevenueReport(startDate, endDate time.Time, companyID uuid.UUID, branchIDs []uuid.UUID, salesType string, productIDs []uuid.UUID) (*domain.CityRevenueReportResponse, error) {
	type storeCityRow struct {
		StoreID    uuid.UUID `gorm:"column:store_id"`
		StoreName  string    `gorm:"column:store_name"`
		RawRegency string    `gorm:"column:regency"`
		RawDist    string    `gorm:"column:distribution_city"`
		RawProv    string    `gorm:"column:province"`
		RawAddress string    `gorm:"column:address"`
		BranchName string    `gorm:"column:branch_name"`
		IsActive   bool      `gorm:"column:is_active"`
	}

	storeQuery := `
		SELECT 
			s.id as store_id,
			s.name as store_name,
			COALESCE(s.regency, '') as regency,
			COALESCE(s.distribution_city, '') as distribution_city,
			COALESCE(s.province, '') as province,
			COALESCE(s.address, '') as address,
			COALESCE(b.name, 'Pusat') as branch_name,
			s.is_active
		FROM stores s
		LEFT JOIN employees e ON e.id = s.assigned_employee_id
		LEFT JOIN branches b ON b.id = e.branch_id
		WHERE 1=1
	`
	var storeArgs []interface{}
	if companyID != uuid.Nil {
		storeQuery += " AND (s.company_id = ? OR e.company_id = ?)"
		storeArgs = append(storeArgs, companyID, companyID)
	}
	if len(branchIDs) == 1 {
		storeQuery += " AND (COALESCE(e.operating_branch_id, e.branch_id) = ? OR s.assigned_employee_id IN (SELECT id FROM employees WHERE COALESCE(operating_branch_id, branch_id) = ?))"
		storeArgs = append(storeArgs, branchIDs[0], branchIDs[0])
	} else if len(branchIDs) > 1 {
		storeQuery += " AND (COALESCE(e.operating_branch_id, e.branch_id) IN (?) OR s.assigned_employee_id IN (SELECT id FROM employees WHERE COALESCE(operating_branch_id, branch_id) IN (?)))"
		storeArgs = append(storeArgs, branchIDs, branchIDs)
	}

	var storeRows []storeCityRow
	if err := u.db.Raw(storeQuery, storeArgs...).Scan(&storeRows).Error; err != nil {
		return nil, err
	}

	type storeInfo struct {
		City       string
		Province   string
		BranchName string
	}

	storeCityMap := make(map[uuid.UUID]storeInfo)
	cityStoresCount := make(map[string]int)
	cityProvinceMap := make(map[string]string)
	cityBranchMap := make(map[string]string)
	monitoredProvinces := make(map[string]bool)
	defaultBranchName := "Pusat"

	for _, sr := range storeRows {
		rawCity := sr.RawRegency
		if rawCity == "" {
			rawCity = sr.RawDist
		}
		cleanCity, cleanProv := NormalizeCityAndProvince(rawCity, sr.RawProv, sr.RawAddress)

		storeCityMap[sr.StoreID] = storeInfo{
			City:       cleanCity,
			Province:   cleanProv,
			BranchName: sr.BranchName,
		}

		if sr.IsActive {
			cityStoresCount[cleanCity]++
		}
		if cleanProv != "" && cleanProv != "Lainnya" && cleanProv != "Tidak Terdeteksi" {
			monitoredProvinces[cleanProv] = true
			if cityProvinceMap[cleanCity] == "" {
				cityProvinceMap[cleanCity] = cleanProv
			}
		}
		if sr.BranchName != "" {
			defaultBranchName = sr.BranchName
			if cityBranchMap[cleanCity] == "" {
				cityBranchMap[cleanCity] = sr.BranchName
			}
		}
	}

	type orderItemRow struct {
		StoreID      uuid.UUID `gorm:"column:store_id"`
		SalesmanName string    `gorm:"column:salesman_name"`
		ProductName  string    `gorm:"column:product_name"`
		Quantity     float64   `gorm:"column:quantity"`
		TotalAmount  float64   `gorm:"column:total_amount"`
	}

	var orderRows []orderItemRow

	// Modern Trade (sales_orders)
	if salesType == "" || salesType == "ALL" || salesType == "ALL_SALES" || salesType == "sales_mt" {
		soQuery := `
			SELECT 
				so.store_id,
				COALESCE(TRIM(CONCAT(e.first_name, ' ', e.last_name)), 'Salesman') as salesman_name,
				COALESCE(p.name, 'Produk') as product_name,
				COALESCE(soi.quantity, 1) as quantity,
				COALESCE(NULLIF(soi.subtotal, 0), soi.total_price, so.total_amount) as total_amount
			FROM sales_orders so
			LEFT JOIN sales_order_items soi ON soi.sales_order_id = so.id
			LEFT JOIN products p ON p.id = soi.product_id
			LEFT JOIN employees e ON e.id = so.employee_id
			WHERE so.created_at BETWEEN ? AND ? 
			  AND so.status NOT IN ('CANCELLED', 'REJECTED', 'DRAFT')
		`
		soArgs := []interface{}{startDate, endDate}
		if companyID != uuid.Nil {
			soQuery += " AND so.company_id = ?"
			soArgs = append(soArgs, companyID)
		}
		if len(branchIDs) == 1 {
			soQuery += " AND (so.branch_id = ? OR COALESCE(e.operating_branch_id, e.branch_id) = ?)"
			soArgs = append(soArgs, branchIDs[0], branchIDs[0])
		} else if len(branchIDs) > 1 {
			soQuery += " AND (so.branch_id IN (?) OR COALESCE(e.operating_branch_id, e.branch_id) IN (?))"
			soArgs = append(soArgs, branchIDs, branchIDs)
		}
		if len(productIDs) > 0 {
			soQuery += " AND soi.product_id IN (?)"
			soArgs = append(soArgs, productIDs)
		}

		var mtRows []orderItemRow
		if err := u.db.Raw(soQuery, soArgs...).Scan(&mtRows).Error; err == nil {
			orderRows = append(orderRows, mtRows...)
		}
	}

	// Taking Order / Motoris (sales_transactions)
	if salesType == "" || salesType == "ALL" || salesType == "ALL_SALES" || salesType != "sales_mt" {
		stQuery := `
			SELECT 
				st.store_id,
				COALESCE(TRIM(CONCAT(e.first_name, ' ', e.last_name)), 'Salesman') as salesman_name,
				COALESCE(p.name, 'Produk') as product_name,
				COALESCE(si.quantity, 1) as quantity,
				COALESCE(NULLIF(si.subtotal, 0), st.total_amount) as total_amount
			FROM sales_transactions st
			LEFT JOIN sales_items si ON si.sales_transaction_id = st.id
			LEFT JOIN products p ON p.id = si.product_id
			LEFT JOIN employees e ON e.id = st.employee_id
			WHERE st.transaction_date BETWEEN ? AND ? 
			  AND st.status NOT IN ('CANCELLED', 'REJECTED')
		`
		stArgs := []interface{}{startDate, endDate}
		if companyID != uuid.Nil {
			stQuery += " AND st.company_id = ?"
			stArgs = append(stArgs, companyID)
		}
		if len(branchIDs) == 1 {
			stQuery += " AND (COALESCE(e.operating_branch_id, e.branch_id) = ? OR st.employee_id IN (SELECT id FROM employees WHERE COALESCE(operating_branch_id, branch_id) = ?))"
			stArgs = append(stArgs, branchIDs[0], branchIDs[0])
		} else if len(branchIDs) > 1 {
			stQuery += " AND (COALESCE(e.operating_branch_id, e.branch_id) IN (?) OR st.employee_id IN (SELECT id FROM employees WHERE COALESCE(operating_branch_id, branch_id) IN (?)))"
			stArgs = append(stArgs, branchIDs, branchIDs)
		}
		if len(productIDs) > 0 {
			stQuery += " AND si.product_id IN (?)"
			stArgs = append(stArgs, productIDs)
		}

		var toRows []orderItemRow
		if err := u.db.Raw(stQuery, stArgs...).Scan(&toRows).Error; err == nil {
			orderRows = append(orderRows, toRows...)
		}
	}

	type cityAgg struct {
		City            string
		Province        string
		BranchName      string
		ActiveStoresMap map[uuid.UUID]bool
		TotalQuantity   float64
		TotalRevenue    float64
		SalesmanRevMap  map[string]float64
		ProductQtyMap   map[string]float64
	}

	cityAggMap := make(map[string]*cityAgg)

	for city := range cityStoresCount {
		cityAggMap[city] = &cityAgg{
			City:            city,
			Province:        cityProvinceMap[city],
			BranchName:      cityBranchMap[city],
			ActiveStoresMap: make(map[uuid.UUID]bool),
			SalesmanRevMap:  make(map[string]float64),
			ProductQtyMap:   make(map[string]float64),
		}
	}

	// Masukkan master Kabupaten/Kota untuk seluruh provinsi operasional yang dipantau.
	// Jika belum ada provinsi terdeteksi dari toko (misal data kosong), pantau provinsi utama.
	if len(monitoredProvinces) == 0 {
		monitoredProvinces["Jawa Tengah"] = true
		monitoredProvinces["Banten"] = true
		monitoredProvinces["Jawa Timur"] = true
		monitoredProvinces["Jawa Barat"] = true
		monitoredProvinces["DI Yogyakarta"] = true
		monitoredProvinces["DKI Jakarta"] = true
	}
	for provName := range monitoredProvinces {
		if masterList, ok := ProvinceMasterRegencies[provName]; ok {
			for _, masterCity := range masterList {
				if _, exists := cityAggMap[masterCity]; !exists {
					cityAggMap[masterCity] = &cityAgg{
						City:            masterCity,
						Province:        provName,
						BranchName:      defaultBranchName,
						ActiveStoresMap: make(map[uuid.UUID]bool),
						SalesmanRevMap:  make(map[string]float64),
						ProductQtyMap:   make(map[string]float64),
					}
				}
			}
		}
	}

	var overallTotalRevenue float64
	var overallTotalQuantity float64

	for _, ord := range orderRows {
		city := "Tidak Terdeteksi"
		prov := "Jawa Tengah"
		branch := defaultBranchName
		if sInfo, ok := storeCityMap[ord.StoreID]; ok {
			city = sInfo.City
			prov = sInfo.Province
			branch = sInfo.BranchName
		}

		agg, exists := cityAggMap[city]
		if !exists {
			agg = &cityAgg{
				City:            city,
				Province:        prov,
				BranchName:      branch,
				ActiveStoresMap: make(map[uuid.UUID]bool),
				SalesmanRevMap:  make(map[string]float64),
				ProductQtyMap:   make(map[string]float64),
			}
			cityAggMap[city] = agg
		}

		if ord.StoreID != uuid.Nil {
			agg.ActiveStoresMap[ord.StoreID] = true
		}
		agg.TotalQuantity += ord.Quantity
		agg.TotalRevenue += ord.TotalAmount
		if ord.SalesmanName != "" {
			agg.SalesmanRevMap[ord.SalesmanName] += ord.TotalAmount
		}
		if ord.ProductName != "" {
			agg.ProductQtyMap[ord.ProductName] += ord.Quantity
		}

		overallTotalRevenue += ord.TotalAmount
		overallTotalQuantity += ord.Quantity
	}

	var cityItems []domain.CityRevenueReportItem
	penetratedCount := 0
	unpenetratedCount := 0

	for city, agg := range cityAggMap {
		totalStores := cityStoresCount[city]
		status := "UNPENETRATED"
		// Wilayah berstatus PENETRATED (Sudah Masuk) jika sudah memiliki toko terdaftar atau ada transaksi
		if totalStores > 0 || agg.TotalRevenue > 0 {
			status = "PENETRATED"
			penetratedCount++
		} else {
			unpenetratedCount++
		}

		topSalesman := "-"
		var maxSalesmanRev float64
		for sName, rev := range agg.SalesmanRevMap {
			if rev > maxSalesmanRev {
				maxSalesmanRev = rev
				topSalesman = sName
			}
		}

		topProduct := "-"
		var maxProdQty float64
		for pName, pQty := range agg.ProductQtyMap {
			if pQty > maxProdQty {
				maxProdQty = pQty
				topProduct = pName
			}
		}

		sharePct := 0.0
		if overallTotalRevenue > 0 {
			sharePct = (agg.TotalRevenue / overallTotalRevenue) * 100.0
		}

		cityItems = append(cityItems, domain.CityRevenueReportItem{
			City:              city,
			Province:          agg.Province,
			Status:            status,
			TotalStores:       totalStores,
			ActiveStoresCount: len(agg.ActiveStoresMap),
			TotalQuantity:     agg.TotalQuantity,
			TotalRevenue:      agg.TotalRevenue,
			RevenueSharePct:   float64(int(sharePct*10)) / 10.0,
			BranchName:        agg.BranchName,
			TopSalesmanName:   topSalesman,
			TopProductName:    topProduct,
		})
	}

	sort.Slice(cityItems, func(i, j int) bool {
		if cityItems[i].TotalRevenue != cityItems[j].TotalRevenue {
			return cityItems[i].TotalRevenue > cityItems[j].TotalRevenue
		}
		if cityItems[i].TotalStores != cityItems[j].TotalStores {
			return cityItems[i].TotalStores > cityItems[j].TotalStores
		}
		if cityItems[i].Status != cityItems[j].Status {
			return cityItems[i].Status == "PENETRATED"
		}
		return cityItems[i].City < cityItems[j].City
	})

	topCityName := "-"
	topCityRevenue := 0.0
	if len(cityItems) > 0 && cityItems[0].TotalRevenue > 0 {
		topCityName = cityItems[0].City
		topCityRevenue = cityItems[0].TotalRevenue
	}

	totalMonitored := len(cityItems)
	penetrationRate := 0.0
	if totalMonitored > 0 {
		penetrationRate = (float64(penetratedCount) / float64(totalMonitored)) * 100.0
	}

	return &domain.CityRevenueReportResponse{
		StartDate:               startDate.Format("2006-01-02"),
		EndDate:                 endDate.Format("2006-01-02"),
		TotalCitiesMonitored:    totalMonitored,
		PenetratedCitiesCount:   penetratedCount,
		UnpenetratedCitiesCount: unpenetratedCount,
		PenetrationRate:         float64(int(penetrationRate*10)) / 10.0,
		TotalRevenue:            overallTotalRevenue,
		TotalQuantity:           overallTotalQuantity,
		TopCityName:             topCityName,
		TopCityRevenue:          topCityRevenue,
		Cities:                  cityItems,
	}, nil
}

func (u *salesUsecase) ExportCityRevenueExcel(startDate, endDate time.Time, companyID uuid.UUID, branchIDs []uuid.UUID, salesType string, productIDs []uuid.UUID) ([]byte, error) {
	data, err := u.GetCityRevenueReport(startDate, endDate, companyID, branchIDs, salesType, productIDs)
	if err != nil {
		return nil, err
	}

	f := excelize.NewFile()
	defer f.Close()

	sheet := "Rekap Omzet Kota"
	f.SetSheetName("Sheet1", sheet)

	borderThin := []excelize.Border{
		{Type: "top", Color: "D0D5DD", Style: 1},
		{Type: "bottom", Color: "D0D5DD", Style: 1},
		{Type: "left", Color: "D0D5DD", Style: 1},
		{Type: "right", Color: "D0D5DD", Style: 1},
	}
	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Color: "0F172A", Size: 14},
	})
	kpiHeaderStyle, _ := f.NewStyle(&excelize.Style{
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"F1F5F9"}, Pattern: 1},
		Font:      &excelize.Font{Bold: true, Color: "475569", Size: 9},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Border:    borderThin,
	})
	kpiValStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "0F172A", Size: 11},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Border:    borderThin,
	})
	tableHeaderStyle, _ := f.NewStyle(&excelize.Style{
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"1E293B"}, Pattern: 1},
		Font:      &excelize.Font{Bold: true, Color: "FFFFFF", Size: 10},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
		Border:    borderThin,
	})
	cellStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: 10, Color: "1E293B"},
		Alignment: &excelize.Alignment{Vertical: "center"},
		Border:    borderThin,
	})
	centerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: 10, Color: "1E293B"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Border:    borderThin,
	})
	statusPenetratedStyle, _ := f.NewStyle(&excelize.Style{
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"DCFCE7"}, Pattern: 1},
		Font:      &excelize.Font{Bold: true, Color: "166534", Size: 9},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Border:    borderThin,
	})
	statusUnpenetratedStyle, _ := f.NewStyle(&excelize.Style{
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"F1F5F9"}, Pattern: 1},
		Font:      &excelize.Font{Bold: true, Color: "64748B", Size: 9},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Border:    borderThin,
	})
	currencyStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 10, Color: "0F766E"},
		Alignment: &excelize.Alignment{Horizontal: "right", Vertical: "center"},
		Border:    borderThin,
		NumFmt:    3,
	})

	f.SetCellValue(sheet, "A1", "LAPORAN REKAPITULASI OMZET & PENETRASI WILAYAH PER KOTA / KABUPATEN")
	f.SetCellStyle(sheet, "A1", "A1", titleStyle)
	f.SetCellValue(sheet, "A2", fmt.Sprintf("Periode: %s s/d %s | Waktu Export: %s WIB", data.StartDate, data.EndDate, time.Now().Format("02/01/2006 15:04")))

	f.SetCellValue(sheet, "A4", "TOTAL KOTA DIPANTAU")
	f.SetCellStyle(sheet, "A4", "A4", kpiHeaderStyle)
	f.SetCellValue(sheet, "A5", data.TotalCitiesMonitored)
	f.SetCellStyle(sheet, "A5", "A5", kpiValStyle)

	f.SetCellValue(sheet, "B4", "KOTA SUDAH MASUK (AKTIF)")
	f.SetCellStyle(sheet, "B4", "B4", kpiHeaderStyle)
	f.SetCellValue(sheet, "B5", data.PenetratedCitiesCount)
	f.SetCellStyle(sheet, "B5", "B5", kpiValStyle)

	f.SetCellValue(sheet, "C4", "KOTA BELUM MASUK (PELUANG)")
	f.SetCellStyle(sheet, "C4", "C4", kpiHeaderStyle)
	f.SetCellValue(sheet, "C5", data.UnpenetratedCitiesCount)
	f.SetCellStyle(sheet, "C5", "C5", kpiValStyle)

	f.SetCellValue(sheet, "D4", "TINGKAT PENETRASI")
	f.SetCellStyle(sheet, "D4", "D4", kpiHeaderStyle)
	f.SetCellValue(sheet, "D5", fmt.Sprintf("%.1f%%", data.PenetrationRate))
	f.SetCellStyle(sheet, "D5", "D5", kpiValStyle)

	f.SetCellValue(sheet, "E4", "TOTAL OMZET NASIONAL/WILAYAH")
	f.SetCellStyle(sheet, "E4", "E4", kpiHeaderStyle)
	f.SetCellValue(sheet, "E5", data.TotalRevenue)
	f.SetCellStyle(sheet, "E5", "E5", currencyStyle)

	headers := []string{
		"No.", "Kota / Kabupaten", "Provinsi", "Status Penetrasi", "Total Toko",
		"Toko Beli (EC)", "Volume Terjual (PCS)", "Total Omzet (Rp)", "% Kontribusi Omzet",
		"Cabang Pembina", "Top Salesman", "Top Produk Terlaris",
	}

	headerRow := 7
	for colIdx, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(colIdx+1, headerRow)
		f.SetCellValue(sheet, cell, h)
		f.SetCellStyle(sheet, cell, cell, tableHeaderStyle)
	}
	f.SetRowHeight(sheet, headerRow, 26)

	for i, c := range data.Cities {
		row := headerRow + 1 + i
		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), i+1)
		f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), centerStyle)

		f.SetCellValue(sheet, fmt.Sprintf("B%d", row), c.City)
		f.SetCellStyle(sheet, fmt.Sprintf("B%d", row), fmt.Sprintf("B%d", row), cellStyle)

		f.SetCellValue(sheet, fmt.Sprintf("C%d", row), c.Province)
		f.SetCellStyle(sheet, fmt.Sprintf("C%d", row), fmt.Sprintf("C%d", row), cellStyle)

		statusText := "Sudah Masuk"
		statusStyle := statusPenetratedStyle
		if c.Status != "PENETRATED" {
			statusText = "Belum Masuk"
			statusStyle = statusUnpenetratedStyle
		}
		f.SetCellValue(sheet, fmt.Sprintf("D%d", row), statusText)
		f.SetCellStyle(sheet, fmt.Sprintf("D%d", row), fmt.Sprintf("D%d", row), statusStyle)

		f.SetCellValue(sheet, fmt.Sprintf("E%d", row), c.TotalStores)
		f.SetCellStyle(sheet, fmt.Sprintf("E%d", row), fmt.Sprintf("E%d", row), centerStyle)

		f.SetCellValue(sheet, fmt.Sprintf("F%d", row), c.ActiveStoresCount)
		f.SetCellStyle(sheet, fmt.Sprintf("F%d", row), fmt.Sprintf("F%d", row), centerStyle)

		f.SetCellValue(sheet, fmt.Sprintf("G%d", row), c.TotalQuantity)
		f.SetCellStyle(sheet, fmt.Sprintf("G%d", row), fmt.Sprintf("G%d", row), centerStyle)

		f.SetCellValue(sheet, fmt.Sprintf("H%d", row), c.TotalRevenue)
		f.SetCellStyle(sheet, fmt.Sprintf("H%d", row), fmt.Sprintf("H%d", row), currencyStyle)

		f.SetCellValue(sheet, fmt.Sprintf("I%d", row), fmt.Sprintf("%.1f%%", c.RevenueSharePct))
		f.SetCellStyle(sheet, fmt.Sprintf("I%d", row), fmt.Sprintf("I%d", row), centerStyle)

		f.SetCellValue(sheet, fmt.Sprintf("J%d", row), c.BranchName)
		f.SetCellStyle(sheet, fmt.Sprintf("J%d", row), fmt.Sprintf("J%d", row), cellStyle)

		f.SetCellValue(sheet, fmt.Sprintf("K%d", row), c.TopSalesmanName)
		f.SetCellStyle(sheet, fmt.Sprintf("K%d", row), fmt.Sprintf("K%d", row), cellStyle)

		f.SetCellValue(sheet, fmt.Sprintf("L%d", row), c.TopProductName)
		f.SetCellStyle(sheet, fmt.Sprintf("L%d", row), fmt.Sprintf("L%d", row), cellStyle)
	}

	f.SetColWidth(sheet, "A", "A", 6)
	f.SetColWidth(sheet, "B", "B", 24)
	f.SetColWidth(sheet, "C", "C", 20)
	f.SetColWidth(sheet, "D", "D", 16)
	f.SetColWidth(sheet, "E", "G", 14)
	f.SetColWidth(sheet, "H", "H", 20)
	f.SetColWidth(sheet, "I", "I", 16)
	f.SetColWidth(sheet, "J", "L", 22)

	var b bytes.Buffer
	if err := f.Write(&b); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

// ConvertToSalesOrder mengkonversi SalesTransaction yang berstatus PENDING menjadi SalesOrder (SO),
// lalu membatalkan/menghapus SalesTransaction asal agar tidak terjadi duplikasi dan stok tidak terpotong langsung.
func (u *salesUsecase) ConvertToSalesOrder(transactionID uuid.UUID, adminID uuid.UUID) (*domain.SalesOrder, error) {
	var createdSO *domain.SalesOrder

	err := u.db.Transaction(func(tx *gorm.DB) error {
		var trx domain.SalesTransaction
		if err := tx.Preload("Items.Product").Preload("BonusItems").Where("id = ?", transactionID).First(&trx).Error; err != nil {
			return fmt.Errorf("faktur penjualan tidak ditemukan: %v", err)
		}

		if trx.Status != "PENDING" {
			return fmt.Errorf("hanya faktur penjualan berstatus PENDING yang dapat diubah menjadi pesanan SO (status saat ini: %s)", trx.Status)
		}

		if trx.StoreID == nil || *trx.StoreID == uuid.Nil {
			return errors.New("faktur tidak memiliki data toko pelanggan yang valid untuk diubah ke SO")
		}

		// 1. Resolve Branch ID & Company ID
		var effectiveBranchID *uuid.UUID
		var emp domain.Employee
		if err := tx.Preload("JobPosition").Where("id = ?", trx.EmployeeID).First(&emp).Error; err == nil {
			if emp.OperatingBranchID != nil && *emp.OperatingBranchID != uuid.Nil {
				effectiveBranchID = emp.OperatingBranchID
			} else if emp.BranchID != nil && *emp.BranchID != uuid.Nil {
				effectiveBranchID = emp.BranchID
			}
		}

		if effectiveBranchID == nil {
			var store domain.Store
			if err := tx.Preload("AssignedEmployee").Where("id = ?", *trx.StoreID).First(&store).Error; err == nil {
				if store.AssignedEmployee != nil && store.AssignedEmployee.BranchID != nil {
					effectiveBranchID = store.AssignedEmployee.BranchID
				}
			}
		}

		if effectiveBranchID == nil {
			var fallbackBranch domain.Branch
			if err := tx.Where("company_id = ?", trx.CompanyID).Order("created_at ASC").First(&fallbackBranch).Error; err == nil {
				effectiveBranchID = &fallbackBranch.ID
			} else if err := tx.Order("created_at ASC").First(&fallbackBranch).Error; err == nil {
				effectiveBranchID = &fallbackBranch.ID
			}
		}

		if effectiveBranchID == nil {
			return errors.New("tidak dapat menentukan cabang operasional untuk pembuatan SO")
		}

		// 2. Siapkan CreateSORequest
		notes := ""
		if trx.ReceiptNo != "" {
			notes = fmt.Sprintf("Dialihkan dari Faktur: %s", trx.ReceiptNo)
		}
		if trx.Notes != nil && *trx.Notes != "" {
			if notes != "" {
				notes += " | " + *trx.Notes
			} else {
				notes = *trx.Notes
			}
		}

		orderDate := trx.TransactionDate
		if orderDate.IsZero() {
			orderDate = time.Now()
		}

		soReq := CreateSORequest{
			BranchID:             *effectiveBranchID,
			CompanyID:            trx.CompanyID,
			EmployeeID:           trx.EmployeeID,
			StoreID:              *trx.StoreID,
			StoreCategory:        trx.StoreCategory,
			Notes:                notes,
			VisitID:              trx.VisitID,
			PaymentMethod:        "SALES_ORDER",
			TotalAmount:          trx.TotalAmount,
			OrderDate:            orderDate,
			CompanyBankAccountID: trx.CompanyBankAccountID,
		}

		if trx.ReceiptImageURL != "" {
			soReq.IsPhysicalNota = true
			soReq.PhysicalNotaImageURL = trx.ReceiptImageURL
		}

		for _, item := range trx.Items {
			unit := "PCS"
			pcsPerUnit := 1
			if item.Product != nil {
				if item.Product.Unit != "" {
					unit = item.Product.Unit
				}
				if item.Product.PcsPerUnit > 0 {
					pcsPerUnit = item.Product.PcsPerUnit
				}
			}

			soReq.Items = append(soReq.Items, SOItemRequest{
				ProductID:       item.ProductID,
				OrderedQuantity: item.Quantity,
				Unit:            unit,
				PiecesPerUnit:   pcsPerUnit,
				Price:           item.PriceAtTransaction,
				DiscountPercent: item.DiscountPercent,
			})
		}

		for _, bonus := range trx.BonusItems {
			soReq.BonusItems = append(soReq.BonusItems, SOBonusItemRequest{
				BonusInventoryID: bonus.BonusInventoryID,
				Quantity:         bonus.Quantity,
				Notes:            bonus.Notes,
			})
		}

		if len(soReq.Items) == 0 {
			return errors.New("faktur tidak memiliki rincian barang untuk dikonversi menjadi SO")
		}

		// 3. Buat Sales Order via SO Usecase
		newSO, err := u.soUsecase.CreateSO(soReq)
		if err != nil {
			return fmt.Errorf("gagal membuat pesanan SO: %v", err)
		}

		// 4. Update status transaksi asal menjadi soft-deleted dengan keterangan terkonversi ke SO
		now := time.Now()
		reason := fmt.Sprintf("Dikonversi ke Pesanan SO (%s)", newSO.SONumber)
		trx.Status = "DELETED"
		trx.DeletedByID = &adminID
		trx.DeletionReason = &reason
		trx.DeletedAt = gorm.DeletedAt{Time: now, Valid: true}
		if err := tx.Omit(clause.Associations).Save(&trx).Error; err != nil {
			return fmt.Errorf("gagal memperbarui status faktur asal: %v", err)
		}

		// Hapus item-item transaksi agar omzet/stok tidak dobel jika di-query tanpa filter deleted
		_ = tx.Where("sales_transaction_id = ?", trx.ID).Delete(&domain.SalesPayment{})

		createdSO = newSO
		return nil
	})

	if err != nil {
		return nil, err
	}

	return createdSO, nil
}

type notaReturnItemJSON struct {
	ProductID   *uuid.UUID `json:"product_id"`
	ProductName string     `json:"product_name"`
	Quantity    int        `json:"quantity"`
	Unit        string     `json:"unit"`
	Price       float64    `json:"price"`
}

func (u *salesUsecase) syncNotaReturnItems(tx *gorm.DB, trx *domain.SalesTransaction, rawNotes string, processedByName string) (string, error) {
	if rawNotes == "" || trx == nil {
		return rawNotes, nil
	}

	var notesMap map[string]interface{}
	if err := json.Unmarshal([]byte(rawNotes), &notesMap); err != nil {
		return rawNotes, nil
	}

	rawReturnItems, ok := notesMap["return_items"]
	if !ok || rawReturnItems == nil {
		return rawNotes, nil
	}

	returnBytes, err := json.Marshal(rawReturnItems)
	if err != nil {
		return rawNotes, nil
	}

	var returnItems []notaReturnItemJSON
	if err := json.Unmarshal(returnBytes, &returnItems); err != nil {
		return rawNotes, nil
	}

	if len(returnItems) == 0 {
		return rawNotes, nil
	}

	// 1. Resolve Target Branch ID
	var targetBranchID uuid.UUID
	if trx.BranchID != nil && *trx.BranchID != uuid.Nil {
		targetBranchID = *trx.BranchID
	}
	if targetBranchID == uuid.Nil && trx.EmployeeID != uuid.Nil {
		var emp domain.Employee
		if err := tx.Where("id = ?", trx.EmployeeID).First(&emp).Error; err == nil {
			if emp.OperatingBranchID != nil && *emp.OperatingBranchID != uuid.Nil {
				targetBranchID = *emp.OperatingBranchID
			} else if emp.BranchID != nil && *emp.BranchID != uuid.Nil {
				targetBranchID = *emp.BranchID
			}
		}
	}
	if targetBranchID == uuid.Nil && trx.StoreID != nil && *trx.StoreID != uuid.Nil {
		var store domain.Store
		if err := tx.Preload("AssignedEmployee").Where("id = ?", *trx.StoreID).First(&store).Error; err == nil && store.AssignedEmployee != nil {
			if store.AssignedEmployee.OperatingBranchID != nil && *store.AssignedEmployee.OperatingBranchID != uuid.Nil {
				targetBranchID = *store.AssignedEmployee.OperatingBranchID
			} else if store.AssignedEmployee.BranchID != nil && *store.AssignedEmployee.BranchID != uuid.Nil {
				targetBranchID = *store.AssignedEmployee.BranchID
			}
		}
	}
	if targetBranchID == uuid.Nil {
		var pabrikBranch domain.Branch
		if err := tx.Where("name ILIKE '%Pabrik%'").First(&pabrikBranch).Error; err == nil {
			targetBranchID = pabrikBranch.ID
		}
	}

	// 2. Resolve Product IDs for each return item (support fuzzy & whitespace-insensitive match)
	notesUpdated := false
	type resolvedItem struct {
		productID   uuid.UUID
		productName string
		quantity    int
		price       float64
		unit        string
	}
	var resolvedItems []resolvedItem
	totalReturnAmount := 0.0

	for i := range returnItems {
		item := &returnItems[i]
		var prodID uuid.UUID
		if item.ProductID != nil && *item.ProductID != uuid.Nil {
			prodID = *item.ProductID
		}

		var prod domain.Product
		if prodID != uuid.Nil {
			if err := tx.Where("id = ?", prodID).First(&prod).Error; err != nil {
				prodID = uuid.Nil
			}
		}

		if prodID == uuid.Nil && strings.TrimSpace(item.ProductName) != "" {
			trimmedName := strings.TrimSpace(item.ProductName)
			normalizedName := strings.ReplaceAll(strings.ToLower(trimmedName), " ", "")

			// a. Exact match
			if err := tx.Where("LOWER(TRIM(name)) = LOWER(?)", trimmedName).First(&prod).Error; err == nil {
				prodID = prod.ID
			} else if err := tx.Where("REPLACE(LOWER(name), ' ', '') = ?", normalizedName).First(&prod).Error; err == nil {
				// b. Whitespace-insensitive match
				prodID = prod.ID
			} else {
				// c. Keyword match
				cleanKw := strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(trimmedName, "BEKAS", ""), "KOSONG", ""))
				if cleanKw != "" {
					_ = tx.Where("name ILIKE ?", "%"+cleanKw+"%").Order("name ASC").First(&prod).Error
					if prod.ID != uuid.Nil {
						prodID = prod.ID
					}
				}
			}

			if prodID != uuid.Nil {
				item.ProductID = &prodID
				item.ProductName = prod.Name
				if item.Unit == "" && prod.Unit != "" {
					item.Unit = prod.Unit
				}
				notesUpdated = true
			}
		}

		if prodID != uuid.Nil && item.Quantity > 0 {
			resolvedItems = append(resolvedItems, resolvedItem{
				productID:   prodID,
				productName: item.ProductName,
				quantity:    item.Quantity,
				price:       item.Price,
				unit:        item.Unit,
			})
			totalReturnAmount += float64(item.Quantity) * item.Price
		}
	}

	if notesUpdated {
		notesMap["return_items"] = returnItems
		if updatedJSON, err := json.Marshal(notesMap); err == nil {
			rawNotes = string(updatedJSON)
		}
	}

	if len(resolvedItems) == 0 || targetBranchID == uuid.Nil {
		return rawNotes, nil
	}

	// 3. Create or Update formal SalesReturn record
	empID := trx.EmployeeID
	if empID == uuid.Nil {
		var firstEmp domain.Employee
		if err := tx.First(&firstEmp).Error; err == nil {
			empID = firstEmp.ID
		}
	}

	now := time.Now()
	var salesRet domain.SalesReturn
	errFindRet := tx.Where("transaction_id = ? AND notes LIKE ?", trx.ID, "%[RETUR_BARANG_SAMPINGAN]%").First(&salesRet).Error

	if errFindRet == gorm.ErrRecordNotFound {
		var count int64
		datePrefix := trx.TransactionDate.Format("20060102")
		tx.Model(&domain.SalesReturn{}).Where("return_no LIKE ?", "RET-SMP/"+datePrefix+"%").Count(&count)
		returnNo := fmt.Sprintf("RET-SMP/%s/%03d", datePrefix, count+1)

		salesRet = domain.SalesReturn{
			ReturnNo:      returnNo,
			TransactionID: &trx.ID,
			BranchID:      targetBranchID,
			EmployeeID:    empID,
			ReturnType:    "POTONG_NOTA",
			Status:        domain.ReturnStatusApproved,
			TotalAmount:   totalReturnAmount,
			Notes:         fmt.Sprintf("[RETUR_BARANG_SAMPINGAN] Retur wadah/sampingan faktur %s", trx.ReceiptNo),
			ApprovedAt:    &now,
			ApprovedByID:  &empID,
			CreatedAt:     trx.TransactionDate,
			UpdatedAt:     now,
		}
		for _, ri := range resolvedItems {
			salesRet.Items = append(salesRet.Items, domain.SalesReturnItem{
				ProductID: ri.productID,
				Quantity:  ri.quantity,
				Price:     ri.price,
				Condition: "GOOD",
				Reason:    "Retur Barang Sampingan / Wadah",
			})
		}
		_ = tx.Create(&salesRet).Error
	} else if errFindRet == nil {
		salesRet.TotalAmount = totalReturnAmount
		salesRet.Status = domain.ReturnStatusApproved
		salesRet.UpdatedAt = now
		_ = tx.Save(&salesRet).Error

		_ = tx.Where("sales_return_id = ?", salesRet.ID).Delete(&domain.SalesReturnItem{}).Error
		for _, ri := range resolvedItems {
			_ = tx.Create(&domain.SalesReturnItem{
				SalesReturnID: salesRet.ID,
				ProductID:     ri.productID,
				Quantity:      ri.quantity,
				Price:         ri.price,
				Condition:     "GOOD",
				Reason:        "Retur Barang Sampingan / Wadah",
			}).Error
		}
	}

	// 4. Update warehouse stock & log mutasi using Delta calculation
	wRepo := repository.NewWarehouseRepository(tx)
	logSource := fmt.Sprintf("RETUR SAMPINGAN: %s", trx.ReceiptNo)
	var existingLogs []domain.WarehouseLog
	_ = tx.Where("branch_id = ? AND (source = ? OR source = ?)", targetBranchID, logSource, "KOREKSI "+logSource).Find(&existingLogs).Error

	prevLogged := make(map[uuid.UUID]int)
	for _, l := range existingLogs {
		if l.Type == "IN" {
			prevLogged[l.ProductID] += l.Quantity
		} else if l.Type == "OUT" {
			prevLogged[l.ProductID] -= l.Quantity
		}
	}

	newQuantities := make(map[uuid.UUID]int)
	for _, ri := range resolvedItems {
		newQuantities[ri.productID] += ri.quantity
	}

	allProductIDs := make(map[uuid.UUID]bool)
	for pID := range newQuantities {
		allProductIDs[pID] = true
	}
	for pID := range prevLogged {
		allProductIDs[pID] = true
	}

	pName := processedByName
	if pName == "" {
		pName = "Admin Nota"
	}

	for pID := range allProductIDs {
		newQ := newQuantities[pID]
		prevQ := prevLogged[pID]
		diff := newQ - prevQ

		if diff == 0 {
			continue
		}

		stock, err := wRepo.GetStock(targetBranchID, pID, "GOOD", "DEFAULT")
		if err != nil || stock == nil {
			stock = &domain.WarehouseStock{
				BranchID:  targetBranchID,
				ProductID: pID,
				StockType: "GOOD",
				BatchNo:   "DEFAULT",
				Quantity:  0,
			}
		}

		stock.Quantity += diff
		if stock.Quantity < 0 {
			stock.Quantity = 0
		}
		_ = wRepo.UpdateStock(stock)

		logType := "IN"
		logQty := diff
		currentSource := logSource
		if diff < 0 {
			logType = "OUT"
			logQty = -diff
			currentSource = fmt.Sprintf("KOREKSI RETUR SAMPINGAN: %s", trx.ReceiptNo)
		}

		_ = wRepo.CreateLog(&domain.WarehouseLog{
			BranchID:    targetBranchID,
			ProductID:   pID,
			Type:        logType,
			Source:      currentSource,
			BatchNo:     "DEFAULT",
			Quantity:    logQty,
			ProcessedBy: pName,
		})
	}

	return rawNotes, nil
}

func (u *salesUsecase) ReconcileAllNotaReturnItems() (int, error) {
	var trxs []domain.SalesTransaction
	err := u.db.Where("notes ILIKE '%return_items%' AND notes NOT LIKE '%\"return_items\":[]%'").
		Order("transaction_date ASC").
		Find(&trxs).Error
	if err != nil {
		return 0, err
	}

	count := 0
	for i := range trxs {
		trx := &trxs[i]
		if trx.Notes == nil || *trx.Notes == "" {
			continue
		}

		var emp domain.Employee
		var empName string
		if trx.EmployeeID != uuid.Nil {
			if err := u.db.Where("id = ?", trx.EmployeeID).First(&emp).Error; err == nil {
				empName = strings.TrimSpace(emp.FirstName + " " + emp.LastName)
			}
		}

		updatedNotes, err := u.syncNotaReturnItems(u.db, trx, *trx.Notes, empName)
		if err == nil {
			if updatedNotes != *trx.Notes {
				trx.Notes = &updatedNotes
				_ = u.db.Model(trx).Update("notes", updatedNotes)
			}
			count++
		}
	}
	return count, nil
}
