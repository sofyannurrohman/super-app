package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ErrSupplierInUse indicates that historical transactions still reference the supplier.
var ErrSupplierInUse = errors.New("supplier masih digunakan oleh transaksi")

// Supplier - data supplier yang mengirimkan barang ke gudang cabang
type Supplier struct {
	ID              uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CompanyID       uuid.UUID `gorm:"type:uuid;not null" json:"company_id"`
	Name            string    `gorm:"type:varchar(255);not null" json:"name"`
	ContactName     string    `gorm:"type:varchar(100)" json:"contact_name"`
	Phone           string    `gorm:"type:varchar(30)" json:"phone"`
	Email           string    `gorm:"type:varchar(100)" json:"email"`
	Address         string    `gorm:"type:text" json:"address"`
	City            string    `gorm:"type:varchar(100)" json:"city"`
	NPWP            string    `gorm:"type:varchar(50)" json:"npwp"`
	BankName        string    `gorm:"type:varchar(100)" json:"bank_name"`
	BankAccount     string    `gorm:"type:varchar(50)" json:"bank_account"`
	BankAccountName string    `gorm:"type:varchar(100)" json:"bank_account_name"`
	TypicalProducts string    `gorm:"type:text" json:"typical_products"` // comma-separated product names/categories
	Notes           string    `gorm:"type:text" json:"notes"`
	IsActive        bool      `gorm:"default:true" json:"is_active"`
	CreatedAt       time.Time `gorm:"default:now()" json:"created_at"`
	UpdatedAt       time.Time `gorm:"default:now()" json:"updated_at"`
}

// SupplierReceiptStatus - status penerimaan barang dari supplier
type SupplierReceiptStatus string

const (
	SupplierReceiptPending  SupplierReceiptStatus = "PENDING"
	SupplierReceiptReceived SupplierReceiptStatus = "RECEIVED" // barang sudah diterima, stok sudah ditambah
	SupplierReceiptRejected SupplierReceiptStatus = "REJECTED" // barang ditolak
	SupplierReceiptPartial  SupplierReceiptStatus = "PARTIAL"  // sebagian diterima
)

type SupplierInvoicePaymentStatus string

const (
	SupplierInvoiceUnpaid  SupplierInvoicePaymentStatus = "UNPAID"
	SupplierInvoicePartial SupplierInvoicePaymentStatus = "PARTIAL"
	SupplierInvoicePaid    SupplierInvoicePaymentStatus = "PAID"
)

// SupplierReceipt - penerimaan barang dari supplier ke gudang cabang
type SupplierReceipt struct {
	ID                uuid.UUID                    `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CompanyID         uuid.UUID                    `gorm:"type:uuid;not null" json:"company_id"`
	BranchID          uuid.UUID                    `gorm:"type:uuid;not null" json:"branch_id"`
	SupplierID        uuid.UUID                    `gorm:"type:uuid;not null" json:"supplier_id"`
	ReceiptNo         string                       `gorm:"-:migration;type:varchar(100);uniqueIndex" json:"receipt_no"`
	SupplierInvoiceNo string                       `gorm:"type:varchar(100)" json:"supplier_invoice_no"`
	InvoiceDate       *time.Time                   `gorm:"type:date" json:"invoice_date,omitempty"`
	PaymentDueDate    *time.Time                   `gorm:"type:date" json:"payment_due_date,omitempty"`
	InvoiceTotal      float64                      `gorm:"type:decimal(15,2);default:0" json:"invoice_total"`
	PaidAmount        float64                      `gorm:"type:decimal(15,2);default:0" json:"paid_amount"`
	PaymentStatus     SupplierInvoicePaymentStatus `gorm:"type:varchar(20);default:'UNPAID'" json:"payment_status"`
	DeliveryDate      *time.Time                   `gorm:"type:date" json:"delivery_date,omitempty"`
	ReceivedDate      *time.Time                   `gorm:"type:date" json:"received_date,omitempty"`
	Status            SupplierReceiptStatus        `gorm:"type:varchar(20);default:'PENDING'" json:"status"`
	ProofImageURL     string                       `gorm:"type:text" json:"proof_image_url"`
	Notes             string                       `gorm:"type:text" json:"notes"`
	ReceivedBy        string                       `gorm:"type:varchar(100)" json:"received_by"`
	TotalItems        int                          `gorm:"type:int;default:0" json:"total_items"`
	CreatedAt         time.Time                    `gorm:"default:now()" json:"created_at"`
	UpdatedAt         time.Time                    `gorm:"default:now()" json:"updated_at"`

	Supplier *Supplier             `gorm:"foreignKey:SupplierID" json:"supplier,omitempty"`
	Branch   *Branch               `gorm:"foreignKey:BranchID" json:"branch,omitempty"`
	Items    []SupplierReceiptItem `gorm:"foreignKey:SupplierReceiptID" json:"items,omitempty"`
	Payments []SupplierPayment     `gorm:"foreignKey:SupplierReceiptID" json:"payments,omitempty"`
}

type SupplierPayment struct {
	ID                uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	SupplierReceiptID uuid.UUID `gorm:"type:uuid;not null;index" json:"supplier_receipt_id"`
	Amount            float64   `gorm:"type:decimal(15,2);not null" json:"amount"`
	PaymentDate       time.Time `gorm:"type:date;not null" json:"payment_date"`
	ReferenceNo       string    `gorm:"type:varchar(100)" json:"reference_no"`
	Notes             string    `gorm:"type:text" json:"notes"`
	CreatedAt         time.Time `gorm:"default:now()" json:"created_at"`
}

// SupplierReceiptItem - item per produk dalam penerimaan
type SupplierReceiptItem struct {
	ID                uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	SupplierReceiptID uuid.UUID `gorm:"type:uuid;not null" json:"supplier_receipt_id"`
	ProductID         uuid.UUID `gorm:"type:uuid;not null" json:"product_id"`
	ExpectedQty       int       `gorm:"type:int;default:0" json:"expected_qty"`
	ReceivedQty       int       `gorm:"type:int;default:0" json:"received_qty"`
	RejectedQty       int       `gorm:"type:int;default:0" json:"rejected_qty"`
	Unit              string    `gorm:"type:varchar(20);default:'pcs'" json:"unit"`
	UnitPrice         float64   `gorm:"type:decimal(15,2);default:0" json:"unit_price"`
	Notes             string    `gorm:"type:text" json:"notes"`
	CreatedAt         time.Time `gorm:"default:now()" json:"created_at"`

	Product *Product `gorm:"foreignKey:ProductID" json:"product,omitempty"`
}

// SupplierStock - stok gudang dari supplier (pool terpisah dari WarehouseStock pabrik)
type SupplierStock struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	BranchID   uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_supplier_stock_branch_product_supplier" json:"branch_id"`
	SupplierID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_supplier_stock_branch_product_supplier" json:"supplier_id"`
	ProductID  uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_supplier_stock_branch_product_supplier" json:"product_id"`
	Quantity   int       `gorm:"type:int;default:0" json:"quantity"`
	UpdatedAt  time.Time `gorm:"default:now()" json:"updated_at"`

	Branch   *Branch   `gorm:"foreignKey:BranchID" json:"branch,omitempty"`
	Supplier *Supplier `gorm:"foreignKey:SupplierID" json:"supplier,omitempty"`
	Product  *Product  `gorm:"foreignKey:ProductID" json:"product,omitempty"`
}

func (s *Supplier) BeforeCreate(tx *gorm.DB) (err error) {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return
}

func (s *SupplierReceipt) BeforeCreate(tx *gorm.DB) (err error) {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return
}

func (s *SupplierReceiptItem) BeforeCreate(tx *gorm.DB) (err error) {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return
}

func (s *SupplierStock) BeforeCreate(tx *gorm.DB) (err error) {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return
}

func (s *SupplierPayment) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}
