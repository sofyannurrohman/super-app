package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SalesOrderStatus string

const (
	SOStatusDraft      SalesOrderStatus = "DRAFT"       // Baru dibuat salesman
	SOStatusConfirmed  SalesOrderStatus = "CONFIRMED"   // Diverifikasi Admin Nota (siap dibatch)
	SOStatusApproved   SalesOrderStatus = "APPROVED"    // Disetujui (Gudang MT)
	SOStatusInDelivery SalesOrderStatus = "IN_DELIVERY" // Sudah masuk Batch Pengiriman
	SOStatusDelivered  SalesOrderStatus = "DELIVERED"   // Barang sampai di pelanggan
	SOStatusPaid       SalesOrderStatus = "PAID"        // Pembayaran sudah diterima
	SOStatusCancelled  SalesOrderStatus = "CANCELLED"   // Dibatalkan
	SOStatusRejected   SalesOrderStatus = "REJECTED"    // Ditolak Admin Nota

	// Legacy statuses — dipertahankan agar data lama tidak rusak
	SOStatusWaitingWarehouse SalesOrderStatus = "WAITING_WAREHOUSE"
	SOStatusWaitingStock     SalesOrderStatus = "WAITING_STOCK"
	SOStatusProcessing       SalesOrderStatus = "PROCESSING"
	SOStatusShipped          SalesOrderStatus = "SHIPPED"
	SOStatusConverted        SalesOrderStatus = "CONVERTED"
)

// SalesOrder adalah dokumen Pesanan Order (PO) yang dibuat salesman.
type SalesOrder struct {
	ID             uuid.UUID        `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	VisitID        *uuid.UUID       `gorm:"type:uuid" json:"visit_id"` // Link ke log check-in
	ClientID       *string          `gorm:"type:varchar(50)" json:"client_id"` // Idempotency key dari mobile app
	SONumber       string           `gorm:"type:varchar(50);unique;not null" json:"so_number"`
	BranchID           uuid.UUID        `gorm:"type:uuid;not null" json:"branch_id"`
	CompanyID          uuid.UUID        `gorm:"type:uuid;not null" json:"company_id"`
	EmployeeID         uuid.UUID        `gorm:"type:uuid" json:"employee_id"`
	UserID             *uuid.UUID       `gorm:"type:uuid" json:"user_id,omitempty"`
	DistributorStoreID *uuid.UUID       `gorm:"type:uuid;index" json:"distributor_store_id,omitempty"`
	IsDistributorOrder bool             `gorm:"default:false;index" json:"is_distributor_order"`
	StoreID            uuid.UUID        `gorm:"type:uuid;not null" json:"store_id"`
	StoreCategory      string           `gorm:"type:varchar(50)" json:"store_category"`
	Status             SalesOrderStatus `gorm:"type:varchar(30);default:'DRAFT'" json:"status"`
	TotalAmount    float64          `gorm:"type:decimal(15,2);default:0" json:"total_amount"`
	DiscountAmount float64          `gorm:"type:decimal(15,2);default:0" json:"discount_amount"`
	Notes          string           `gorm:"type:text" json:"notes"`
	Latitude       *float64         `gorm:"type:decimal(10,8)" json:"latitude"`
	Longitude      *float64         `gorm:"type:decimal(11,8)" json:"longitude"`


	// --- Fields untuk e-Faktur Pajak ---
	IsTaxable        bool       `gorm:"default:false" json:"is_taxable"`
	TaxInvoiceNumber *string    `gorm:"type:varchar(50)" json:"tax_invoice_number"` // NSFP manual input admin
	TaxInvoiceDate   *time.Time `gorm:"type:date" json:"tax_invoice_date"`
	TaxRate          float64    `gorm:"type:decimal(5,2);default:11.00" json:"tax_rate"`
	DPPAmount        float64    `gorm:"type:decimal(15,2);default:0" json:"dpp_amount"`
	PPNAmount        float64    `gorm:"type:decimal(15,2);default:0" json:"ppn_amount"`

	// --- Modern Trade (MT) fields ---
	// IsMT: true jika SO ini dibuat oleh Sales MT.
	// Omzet dan faktur SO ini menjadi tanggung jawab pusat/pabrik (FactoryCompanyID).
	IsMT bool `gorm:"type:boolean;default:false" json:"is_mt"`
	// FactoryCompanyID: company_id pabrik/pusat yang mengelola omzet MT ini.
	FactoryCompanyID    *uuid.UUID `gorm:"type:uuid" json:"factory_company_id,omitempty"`
	BillingNoteURL      string     `gorm:"type:varchar(255)" json:"billing_note_url"`
	TaxInvoiceURL       string     `gorm:"type:varchar(255)" json:"tax_invoice_url"`
	DeliveryNoteURL     string     `gorm:"type:varchar(255)" json:"delivery_note_url"`
	DestinationBranchID *uuid.UUID `gorm:"type:uuid" json:"destination_branch_id,omitempty"`
	DestinationBranch   *Branch    `gorm:"foreignKey:DestinationBranchID" json:"destination_branch,omitempty"`

	// --- Nota Fisik Sales TO ---
	IsPhysicalNota       bool   `gorm:"type:boolean;default:false" json:"is_physical_nota"`
	PhysicalNotaImageURL string `gorm:"type:varchar(255)" json:"physical_nota_image_url"`

	// --- Verifikasi Admin Nota ---
	AdminNotaConfirmedAt   *time.Time `json:"admin_nota_confirmed_at,omitempty"`
	AdminNotaConfirmedByID *uuid.UUID `gorm:"type:uuid" json:"admin_nota_confirmed_by_id"`
	AdminNotaRejectNotes   string     `gorm:"type:text" json:"admin_nota_reject_notes"`

	// --- Batch Pengiriman ---
	DeliveryBatchID *uuid.UUID `gorm:"type:uuid" json:"delivery_batch_id"`
	DeliveryOrderNo *string    `gorm:"-:migration;type:varchar(100);column:delivery_order_no" json:"delivery_order_no"`
	ShippedAt       *time.Time `json:"shipped_at,omitempty"`

	// --- Proof of Delivery (POD) ---
	PODImageURL       *string    `gorm:"type:text" json:"pod_image_url"`
	SignatureImageURL *string    `gorm:"type:text" json:"signature_image_url"`
	ReceivedAt        *time.Time `json:"received_at,omitempty"`
	ReceivedBy        string     `gorm:"type:varchar(100)" json:"received_by"`

	// --- Pembayaran & Midtrans ---
	PaymentStatus          string     `gorm:"type:varchar(20);default:'UNPAID'" json:"payment_status"` // UNPAID, PARTIAL, PAID
	PaymentCollectedAt     *time.Time `json:"payment_collected_at,omitempty"`
	PaymentCollectedAmount float64    `gorm:"type:decimal(15,2);default:0" json:"payment_collected_amount"`
	PaymentMethod          string     `gorm:"type:varchar(50)" json:"payment_method"` // CASH, MIDTRANS_QRIS, MIDTRANS_VA
	PaymentDueDate         *time.Time `json:"payment_due_date,omitempty"`
	CompanyBankAccountID   *uuid.UUID `gorm:"type:uuid" json:"company_bank_account_id"`
	MidtransTransactionID  *string    `gorm:"type:varchar(100)" json:"midtrans_transaction_id"`
	MidtransQRISURL        string     `gorm:"type:text" json:"midtrans_qris_url"`
	MidtransVANumber       *string    `gorm:"type:varchar(50)" json:"midtrans_va_number"`
	MidtransBank           *string    `gorm:"type:varchar(20)" json:"midtrans_bank"`
	MidtransBillKey        *string    `gorm:"type:varchar(50)" json:"midtrans_bill_key"`
	MidtransBillerCode     *string    `gorm:"type:varchar(20)" json:"midtrans_biller_code"`

	// --- Link ke Invoice setelah lunas ---
	InvoiceID *uuid.UUID `gorm:"type:uuid" json:"invoice_id"`

	// --- Legacy / Approval metadata ---
	ConfirmedAt   *time.Time `json:"confirmed_at,omitempty"`
	ConfirmedByID *uuid.UUID `gorm:"type:uuid" json:"confirmed_by_id"`
	ConvertedAt   *time.Time `json:"converted_at,omitempty"`
	RejectedAt    *time.Time `json:"rejected_at,omitempty"`
	RejectedByID  *uuid.UUID `gorm:"type:uuid" json:"rejected_by_id"`
	RejectNotes   string     `gorm:"type:text" json:"reject_notes"`

	OrderDate time.Time `gorm:"default:now()" json:"order_date"`
	CreatedAt time.Time `gorm:"default:now()" json:"created_at"`
	UpdatedAt time.Time `gorm:"default:now()" json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	DeletedByID    *uuid.UUID     `gorm:"type:uuid" json:"deleted_by_id,omitempty"`
	DeletionReason *string        `gorm:"type:text" json:"deletion_reason,omitempty"`

	Branch               *Branch               `gorm:"foreignKey:BranchID" json:"branch,omitempty"`
	Company              *Company              `gorm:"foreignKey:CompanyID" json:"company,omitempty"`
	FactoryCompany       *Company              `gorm:"foreignKey:FactoryCompanyID" json:"factory_company,omitempty"`
	Employee             *Employee             `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
	User                 *User                 `gorm:"foreignKey:UserID" json:"user,omitempty"`
	DistributorStore     *Store                `gorm:"foreignKey:DistributorStoreID" json:"distributor_store,omitempty"`
	Store                *Store                `gorm:"foreignKey:StoreID" json:"store,omitempty"`
	ConfirmedBy          *Employee             `gorm:"foreignKey:ConfirmedByID" json:"confirmed_by,omitempty"`
	AdminNotaConfirmedBy *Employee             `gorm:"foreignKey:AdminNotaConfirmedByID" json:"admin_nota_confirmed_by,omitempty"`
	DeletedBy            *Employee             `gorm:"foreignKey:DeletedByID" json:"deleted_by,omitempty"`
	DeliveryBatch        *DeliveryBatch        `gorm:"foreignKey:DeliveryBatchID" json:"delivery_batch,omitempty"`
	Invoice              *SalesTransaction     `gorm:"foreignKey:InvoiceID" json:"invoice,omitempty"`
	CompanyBankAccount   *CompanyBankAccount   `gorm:"foreignKey:CompanyBankAccountID" json:"company_bank_account,omitempty"`
	Items                []SalesOrderItem      `gorm:"foreignKey:SalesOrderID;constraint:OnDelete:CASCADE;" json:"items,omitempty"`
	Returns              []SalesReturn         `gorm:"foreignKey:SalesOrderID;constraint:OnDelete:CASCADE;" json:"returns,omitempty"`
	BonusItems           []SalesOrderBonusItem `gorm:"foreignKey:SalesOrderID;constraint:OnDelete:CASCADE;" json:"bonus_items,omitempty"`
	Payments             []SalesPayment        `gorm:"foreignKey:SalesOrderID;constraint:OnDelete:CASCADE;" json:"payments,omitempty"`
}

// SalesOrderItem adalah item produk dalam sebuah Sales Order.
type SalesOrderItem struct {
	ID               uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	SalesOrderID     uuid.UUID `gorm:"type:uuid;not null" json:"sales_order_id"`
	ProductID        uuid.UUID `gorm:"type:uuid;not null" json:"product_id"`
	Quantity         int       `gorm:"type:int;not null" json:"quantity"`                   // Total Qty dalam satuan terkecil (Base Unit)
	OrderedQuantity  int       `gorm:"type:int;not null;default:0" json:"ordered_quantity"` // Qty yang diinput (misal: 3)
	Unit             string    `gorm:"type:varchar(20);not null;default:'PCS'" json:"unit"` // Satuan yang dipilih (misal: Karton)
	PiecesPerUnit    int       `gorm:"type:int;not null;default:1" json:"pieces_per_unit"`  // Isi per satuan (misal: 24)
	ReservedQuantity int       `gorm:"type:int;default:0" json:"reserved_quantity"`         // Stok ter-lock (Auto-Allocation)
	ActualQuantity   int       `gorm:"type:int;default:0" json:"actual_quantity"`           // Qty Riil yang dikirim Gudang
	Price            float64   `gorm:"type:decimal(15,2);default:0" json:"price"`
	DiscountPercent  float64   `gorm:"type:decimal(5,2);default:0" json:"discount_percent"`
	Subtotal         float64   `gorm:"type:decimal(15,2);default:0" json:"subtotal"`
	CreatedAt        time.Time `gorm:"default:now()" json:"created_at"`
	UpdatedAt        time.Time `gorm:"default:now()" json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	Product *Product              `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	Batches []SalesOrderItemBatch `gorm:"foreignKey:SalesOrderItemID" json:"batches,omitempty"`
}

// SalesOrderItemBatch menyimpan rincian batch mana saja yang ter-reserve untuk item ini.
type SalesOrderItemBatch struct {
	ID               uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	SalesOrderItemID uuid.UUID `gorm:"type:uuid;not null;index" json:"sales_order_item_id"`
	BatchNo          string    `gorm:"type:varchar(50);not null" json:"batch_no"`
	Quantity         int       `gorm:"type:int;not null" json:"quantity"`
	CreatedAt        time.Time `gorm:"default:now()" json:"created_at"`
	UpdatedAt        time.Time `gorm:"default:now()" json:"updated_at"`
}

// SalesOrderBonusItem stores bonus items given with a Sales Order.
type SalesOrderBonusItem struct {
	ID               uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	SalesOrderID     uuid.UUID `gorm:"type:uuid;not null;index" json:"sales_order_id"`
	BonusInventoryID uuid.UUID `gorm:"type:uuid;not null;index" json:"bonus_inventory_id"`
	Quantity         int       `gorm:"type:int;not null" json:"quantity"`
	Notes            string    `gorm:"type:text" json:"notes"`
	CreatedAt        time.Time `gorm:"default:now()" json:"created_at"`
	UpdatedAt        time.Time `gorm:"default:now()" json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	BonusInventory *BonusInventory `gorm:"foreignKey:BonusInventoryID" json:"bonus_inventory,omitempty"`
}

func (so *SalesOrder) BeforeCreate(tx *gorm.DB) (err error) {
	if so.ID == uuid.Nil {
		so.ID = uuid.New()
	}
	return
}

func (soi *SalesOrderItem) BeforeCreate(tx *gorm.DB) (err error) {
	if soi.ID == uuid.Nil {
		soi.ID = uuid.New()
	}
	return
}

func (sobi *SalesOrderBonusItem) BeforeCreate(tx *gorm.DB) (err error) {
	if sobi.ID == uuid.Nil {
		sobi.ID = uuid.New()
	}
	return
}
