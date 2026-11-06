package usecase

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"example.com/hris/backend/internal/repository"
)

type SupplierUsecase interface {
	// Supplier CRUD
	CreateSupplier(req CreateSupplierRequest, companyID uuid.UUID) (*domain.Supplier, error)
	BulkCreateSuppliers(reqs []CreateSupplierRequest, companyID uuid.UUID) ([]domain.Supplier, error)
	GetSupplierByID(id uuid.UUID) (*domain.Supplier, error)
	ListSuppliers(companyID uuid.UUID) ([]domain.Supplier, error)
	UpdateSupplier(id uuid.UUID, req CreateSupplierRequest) (*domain.Supplier, error)
	DeleteSupplier(id uuid.UUID) error

	// Supplier Receipt CRUD
	CreateSupplierReceipt(req CreateSupplierReceiptRequest, companyID uuid.UUID) (*domain.SupplierReceipt, error)
	GetSupplierReceiptByID(id uuid.UUID) (*domain.SupplierReceipt, error)
	ListSupplierReceipts(branchID uuid.UUID) ([]domain.SupplierReceipt, error)
	UpdateSupplierReceipt(id uuid.UUID, req CreateSupplierReceiptRequest) (*domain.SupplierReceipt, error)
	ReceiveSupplierReceipt(id uuid.UUID, req ReceiveSupplierReceiptRequest) error
	DeleteSupplierReceipt(id uuid.UUID) error
	PaySupplier(receiptID uuid.UUID, amount float64, paymentDate time.Time, referenceNo, notes string) error

	// Supplier Stock
	GetSupplierStock(branchID uuid.UUID) ([]domain.SupplierStock, error)
}

// ─── Request DTOs ─────────────────────────────────────────────────────────────

type CreateSupplierRequest struct {
	Name            string `json:"name" binding:"required"`
	ContactName     string `json:"contact_name"`
	Phone           string `json:"phone"`
	Email           string `json:"email"`
	Address         string `json:"address"`
	City            string `json:"city"`
	NPWP            string `json:"npwp"`
	BankName        string `json:"bank_name"`
	BankAccount     string `json:"bank_account"`
	BankAccountName string `json:"bank_account_name"`
	TypicalProducts string `json:"typical_products"`
	Notes           string `json:"notes"`
	IsActive        *bool  `json:"is_active"`
}

type SupplierReceiptItemRequest struct {
	ProductID   uuid.UUID `json:"product_id" binding:"required"`
	ExpectedQty int       `json:"expected_qty"`
	ReceivedQty int       `json:"received_qty"`
	RejectedQty int       `json:"rejected_qty"`
	Unit        string    `json:"unit"`
	UnitPrice   float64   `json:"unit_price"`
	Notes       string    `json:"notes"`
}

type CreateSupplierReceiptRequest struct {
	BranchID          uuid.UUID                    `json:"branch_id" binding:"required"`
	SupplierID        uuid.UUID                    `json:"supplier_id" binding:"required"`
	SupplierInvoiceNo string                       `json:"supplier_invoice_no"`
	InvoiceDate       *time.Time                   `json:"invoice_date,omitempty"`
	PaymentDueDate    *time.Time                   `json:"payment_due_date,omitempty"`
	DeliveryDate      *time.Time                   `json:"delivery_date,omitempty"`
	ReceivedDate      *time.Time                   `json:"received_date,omitempty"`
	ProofImageURL     string                       `json:"proof_image_url"`
	Notes             string                       `json:"notes"`
	ReceivedBy        string                       `json:"received_by"`
	Items             []SupplierReceiptItemRequest `json:"items"`
}

type SupplierPaymentRequest struct {
	Amount      float64    `json:"amount" binding:"required,gt=0"`
	PaymentDate *time.Time `json:"payment_date"`
	ReferenceNo string     `json:"reference_no"`
	Notes       string     `json:"notes"`
}

type ReceiveSupplierReceiptRequest struct {
	Status     domain.SupplierReceiptStatus `json:"status" binding:"required"`
	ReceivedBy string                       `json:"received_by"`
	Notes      string                       `json:"notes"`
	Items      []SupplierReceiptItemRequest `json:"items"`
}

// ─── Implementation ───────────────────────────────────────────────────────────

type supplierUsecase struct {
	repo repository.SupplierRepository
}

func NewSupplierUsecase(repo repository.SupplierRepository) SupplierUsecase {
	return &supplierUsecase{repo}
}

func (u *supplierUsecase) CreateSupplier(req CreateSupplierRequest, companyID uuid.UUID) (*domain.Supplier, error) {
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}
	supplier := &domain.Supplier{
		CompanyID:       companyID,
		Name:            req.Name,
		ContactName:     req.ContactName,
		Phone:           req.Phone,
		Email:           req.Email,
		Address:         req.Address,
		City:            req.City,
		NPWP:            req.NPWP,
		BankName:        req.BankName,
		BankAccount:     req.BankAccount,
		BankAccountName: req.BankAccountName,
		TypicalProducts: req.TypicalProducts,
		Notes:           req.Notes,
		IsActive:        isActive,
	}
	if err := u.repo.CreateSupplier(supplier); err != nil {
		return nil, err
	}
	return supplier, nil
}

func (u *supplierUsecase) BulkCreateSuppliers(reqs []CreateSupplierRequest, companyID uuid.UUID) ([]domain.Supplier, error) {
	if len(reqs) == 0 {
		return nil, errors.New("tidak ada data supplier untuk diimpor")
	}

	var suppliers []domain.Supplier
	for _, req := range reqs {
		if req.Name == "" {
			continue
		}
		isActive := true
		if req.IsActive != nil {
			isActive = *req.IsActive
		}
		supplier := domain.Supplier{
			ID:              uuid.New(),
			CompanyID:       companyID,
			Name:            req.Name,
			ContactName:     req.ContactName,
			Phone:           req.Phone,
			Email:           req.Email,
			Address:         req.Address,
			City:            req.City,
			NPWP:            req.NPWP,
			BankName:        req.BankName,
			BankAccount:     req.BankAccount,
			BankAccountName: req.BankAccountName,
			TypicalProducts: req.TypicalProducts,
			Notes:           req.Notes,
			IsActive:        isActive,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		}
		suppliers = append(suppliers, supplier)
	}

	if len(suppliers) == 0 {
		return nil, errors.New("semua baris tidak valid atau nama supplier kosong")
	}

	if err := u.repo.BulkCreateSuppliers(suppliers); err != nil {
		return nil, err
	}

	return suppliers, nil
}

func (u *supplierUsecase) GetSupplierByID(id uuid.UUID) (*domain.Supplier, error) {
	return u.repo.GetSupplierByID(id)
}

func (u *supplierUsecase) ListSuppliers(companyID uuid.UUID) ([]domain.Supplier, error) {
	return u.repo.ListSuppliers(companyID)
}

func (u *supplierUsecase) UpdateSupplier(id uuid.UUID, req CreateSupplierRequest) (*domain.Supplier, error) {
	supplier, err := u.repo.GetSupplierByID(id)
	if err != nil {
		return nil, err
	}
	supplier.Name = req.Name
	supplier.ContactName = req.ContactName
	supplier.Phone = req.Phone
	supplier.Email = req.Email
	supplier.Address = req.Address
	supplier.City = req.City
	supplier.NPWP = req.NPWP
	supplier.BankName = req.BankName
	supplier.BankAccount = req.BankAccount
	supplier.BankAccountName = req.BankAccountName
	supplier.TypicalProducts = req.TypicalProducts
	supplier.Notes = req.Notes
	if req.IsActive != nil {
		supplier.IsActive = *req.IsActive
	}
	if err := u.repo.UpdateSupplier(supplier); err != nil {
		return nil, err
	}
	return supplier, nil
}

func (u *supplierUsecase) DeleteSupplier(id uuid.UUID) error {
	return u.repo.DeleteSupplier(id)
}

func (u *supplierUsecase) CreateSupplierReceipt(req CreateSupplierReceiptRequest, companyID uuid.UUID) (*domain.SupplierReceipt, error) {
	receiptNo := u.repo.GenerateReceiptNo(req.BranchID)

	items := make([]domain.SupplierReceiptItem, 0, len(req.Items))
	var invoiceTotal float64
	for _, itemReq := range req.Items {
		unit := itemReq.Unit
		if unit == "" {
			unit = "pcs"
		}
		items = append(items, domain.SupplierReceiptItem{
			ProductID:   itemReq.ProductID,
			ExpectedQty: itemReq.ExpectedQty,
			ReceivedQty: itemReq.ReceivedQty,
			RejectedQty: itemReq.RejectedQty,
			Unit:        unit,
			UnitPrice:   itemReq.UnitPrice,
			Notes:       itemReq.Notes,
		})
		invoiceTotal += float64(itemReq.ReceivedQty) * itemReq.UnitPrice
	}

	receipt := &domain.SupplierReceipt{
		CompanyID:         companyID,
		BranchID:          req.BranchID,
		SupplierID:        req.SupplierID,
		ReceiptNo:         receiptNo,
		SupplierInvoiceNo: req.SupplierInvoiceNo,
		InvoiceDate:       req.InvoiceDate,
		PaymentDueDate:    req.PaymentDueDate,
		DeliveryDate:      req.DeliveryDate,
		ReceivedDate:      req.ReceivedDate,
		Status:            domain.SupplierReceiptPending,
		ProofImageURL:     req.ProofImageURL,
		Notes:             req.Notes,
		ReceivedBy:        req.ReceivedBy,
		TotalItems:        len(items),
		InvoiceTotal:      invoiceTotal,
		PaymentStatus:     domain.SupplierInvoiceUnpaid,
		Items:             items,
	}

	if err := u.repo.CreateSupplierReceipt(receipt); err != nil {
		return nil, err
	}
	return receipt, nil
}

func (u *supplierUsecase) GetSupplierReceiptByID(id uuid.UUID) (*domain.SupplierReceipt, error) {
	return u.repo.GetSupplierReceiptByID(id)
}

func (u *supplierUsecase) ListSupplierReceipts(branchID uuid.UUID) ([]domain.SupplierReceipt, error) {
	return u.repo.ListSupplierReceipts(branchID)
}

func (u *supplierUsecase) UpdateSupplierReceipt(id uuid.UUID, req CreateSupplierReceiptRequest) (*domain.SupplierReceipt, error) {
	receipt, err := u.repo.GetSupplierReceiptByID(id)
	if err != nil {
		return nil, err
	}
	if receipt.Status != domain.SupplierReceiptPending {
		return nil, errors.New("hanya penerimaan berstatus PENDING yang bisa diedit")
	}

	items := make([]domain.SupplierReceiptItem, 0, len(req.Items))
	var invoiceTotal float64
	for _, itemReq := range req.Items {
		unit := itemReq.Unit
		if unit == "" {
			unit = "pcs"
		}
		items = append(items, domain.SupplierReceiptItem{
			SupplierReceiptID: receipt.ID,
			ProductID:         itemReq.ProductID,
			ExpectedQty:       itemReq.ExpectedQty,
			ReceivedQty:       itemReq.ReceivedQty,
			RejectedQty:       itemReq.RejectedQty,
			Unit:              unit,
			UnitPrice:         itemReq.UnitPrice,
			Notes:             itemReq.Notes,
		})
		invoiceTotal += float64(itemReq.ReceivedQty) * itemReq.UnitPrice
	}

	receipt.SupplierID = req.SupplierID
	receipt.SupplierInvoiceNo = req.SupplierInvoiceNo
	receipt.InvoiceDate = req.InvoiceDate
	receipt.PaymentDueDate = req.PaymentDueDate
	receipt.DeliveryDate = req.DeliveryDate
	receipt.ReceivedDate = req.ReceivedDate
	receipt.ProofImageURL = req.ProofImageURL
	receipt.Notes = req.Notes
	receipt.ReceivedBy = req.ReceivedBy
	receipt.TotalItems = len(items)
	receipt.InvoiceTotal = invoiceTotal
	receipt.Items = items

	if err := u.repo.UpdateSupplierReceipt(receipt); err != nil {
		return nil, err
	}
	return receipt, nil
}

func (u *supplierUsecase) ReceiveSupplierReceipt(id uuid.UUID, req ReceiveSupplierReceiptRequest) error {
	receipt, err := u.repo.GetSupplierReceiptByID(id)
	if err != nil {
		return err
	}
	if receipt.Status == domain.SupplierReceiptReceived {
		return errors.New("penerimaan ini sudah diproses")
	}

	// Update items if provided
	if len(req.Items) > 0 {
		for i, itemReq := range req.Items {
			if i < len(receipt.Items) {
				receipt.Items[i].ReceivedQty = itemReq.ReceivedQty
				receipt.Items[i].RejectedQty = itemReq.RejectedQty
				receipt.Items[i].Notes = itemReq.Notes
			}
		}
	}

	// Update status
	receipt.Status = req.Status
	if req.ReceivedBy != "" {
		receipt.ReceivedBy = req.ReceivedBy
	}
	if req.Notes != "" {
		receipt.Notes = req.Notes
	}
	now := time.Now()
	receipt.ReceivedDate = &now

	if err := u.repo.UpdateSupplierReceipt(receipt); err != nil {
		return err
	}

	// If RECEIVED or PARTIAL, update supplier stock
	if req.Status == domain.SupplierReceiptReceived || req.Status == domain.SupplierReceiptPartial {
		for _, item := range receipt.Items {
			if item.ReceivedQty > 0 {
				if err := u.repo.UpsertSupplierStock(receipt.BranchID, receipt.SupplierID, item.ProductID, item.ReceivedQty); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (u *supplierUsecase) DeleteSupplierReceipt(id uuid.UUID) error {
	receipt, err := u.repo.GetSupplierReceiptByID(id)
	if err != nil {
		return err
	}
	if receipt.Status == domain.SupplierReceiptReceived {
		return errors.New("penerimaan yang sudah diproses tidak bisa dihapus")
	}
	return u.repo.DeleteSupplierReceipt(id)
}

func (u *supplierUsecase) PaySupplier(receiptID uuid.UUID, amount float64, paymentDate time.Time, referenceNo, notes string) error {
	return u.repo.CreateSupplierPayment(&domain.SupplierPayment{SupplierReceiptID: receiptID, Amount: amount, PaymentDate: paymentDate, ReferenceNo: referenceNo, Notes: notes})
}

func (u *supplierUsecase) GetSupplierStock(branchID uuid.UUID) ([]domain.SupplierStock, error) {
	return u.repo.GetSupplierStock(branchID)
}
