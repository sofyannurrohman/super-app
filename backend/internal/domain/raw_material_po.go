package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RawMaterialPOStatus string

const (
	RawMaterialPODraft             RawMaterialPOStatus = "DRAFT"
	RawMaterialPOPendingSPV        RawMaterialPOStatus = "PENDING_SPV"
	RawMaterialPOPendingPurchasing RawMaterialPOStatus = "PENDING_PURCHASING"
	RawMaterialPOApproved          RawMaterialPOStatus = "APPROVED"
	RawMaterialPOPartialReceived   RawMaterialPOStatus = "PARTIAL_RECEIVED"
	RawMaterialPOReceived          RawMaterialPOStatus = "RECEIVED"
	RawMaterialPORejected          RawMaterialPOStatus = "REJECTED"
	RawMaterialPOCancelled         RawMaterialPOStatus = "CANCELLED"
)

type RawMaterialPurchaseOrder struct {
	ID                   uuid.UUID           `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CompanyID            uuid.UUID           `gorm:"type:uuid;not null" json:"company_id"`
	FactoryID            uuid.UUID           `gorm:"type:uuid;not null" json:"factory_id"`
	SupplierID           uuid.UUID           `gorm:"type:uuid;not null" json:"supplier_id"`
	PONumber             string              `gorm:"type:varchar(100);not null;index" json:"po_number"`
	OrderDate            time.Time           `gorm:"type:date;not null" json:"order_date"`
	EstimatedArrival     *time.Time          `gorm:"type:date" json:"estimated_arrival,omitempty"`
	Status               RawMaterialPOStatus `gorm:"type:varchar(30);default:'PENDING_SPV'" json:"status"`
	TotalEstimatedAmount float64             `gorm:"type:decimal(15,2);default:0" json:"total_estimated_amount"`
	Notes                string              `gorm:"type:text" json:"notes"`
	CreatedByID          *uuid.UUID          `gorm:"type:uuid" json:"created_by_id,omitempty"`
	CreatedByName        string              `gorm:"type:varchar(100)" json:"created_by_name"`

	// Approval SPV Gudang
	SPVApprovedByID       *uuid.UUID `gorm:"type:uuid" json:"spv_approved_by_id,omitempty"`
	SPVApprovedByName     string     `gorm:"type:varchar(100)" json:"spv_approved_by_name,omitempty"`
	SPVApprovedByPosition string     `gorm:"type:varchar(100)" json:"spv_approved_by_position,omitempty"`
	SPVApprovedAt         *time.Time `json:"spv_approved_at,omitempty"`
	SPVNotes              string     `gorm:"type:text" json:"spv_notes,omitempty"`

	// Approval Purchasing
	PurchasingApprovedByID       *uuid.UUID `gorm:"type:uuid" json:"purchasing_approved_by_id,omitempty"`
	PurchasingApprovedByName     string     `gorm:"type:varchar(100)" json:"purchasing_approved_by_name,omitempty"`
	PurchasingApprovedByPosition string     `gorm:"type:varchar(100)" json:"purchasing_approved_by_position,omitempty"`
	PurchasingApprovedAt         *time.Time `json:"purchasing_approved_at,omitempty"`
	PurchasingNotes              string     `gorm:"type:text" json:"purchasing_notes,omitempty"`

	// Purchasing Milestone / Progress Tracking
	ProgressStatus        string     `gorm:"type:varchar(50);default:'PENDING_APPROVAL'" json:"progress_status"`
	ProgressNotes         string     `gorm:"type:text" json:"progress_notes,omitempty"`
	ProgressUpdatedByID   *uuid.UUID `gorm:"type:uuid" json:"progress_updated_by_id,omitempty"`
	ProgressUpdatedByName string     `gorm:"type:varchar(100)" json:"progress_updated_by_name,omitempty"`
	ProgressUpdatedAt     *time.Time `json:"progress_updated_at,omitempty"`

	// Finance & Payment Status
	PaymentStatus string `gorm:"type:varchar(30);default:'UNPAID'" json:"payment_status"`
	PaymentNotes  string `gorm:"type:text" json:"payment_notes,omitempty"`

	RejectionReason string     `gorm:"type:text" json:"rejection_reason,omitempty"`
	RejectedByID    *uuid.UUID `gorm:"type:uuid" json:"rejected_by_id,omitempty"`
	RejectedByName  string     `gorm:"type:varchar(100)" json:"rejected_by_name,omitempty"`
	RejectedAt      *time.Time `json:"rejected_at,omitempty"`

	PurchaseRequestID    *uuid.UUID          `gorm:"type:uuid;index" json:"purchase_request_id,omitempty"`
	CreatedAt time.Time `gorm:"default:now()" json:"created_at"`
	UpdatedAt time.Time `gorm:"default:now()" json:"updated_at"`

	Company         *Company                       `gorm:"foreignKey:CompanyID" json:"company,omitempty"`
	Factory         *Factory                       `gorm:"foreignKey:FactoryID" json:"factory,omitempty"`
	Supplier        *Supplier                      `gorm:"foreignKey:SupplierID" json:"supplier,omitempty"`
	PurchaseRequest *RawMaterialPurchaseRequest    `gorm:"foreignKey:PurchaseRequestID" json:"purchase_request,omitempty"`
	Items           []RawMaterialPurchaseOrderItem `gorm:"foreignKey:PurchaseOrderID" json:"items,omitempty"`
	Receipts        []RawMaterialReceipt           `gorm:"foreignKey:PurchaseOrderID" json:"receipts,omitempty"`
}

type RawMaterialPurchaseOrderItem struct {
	ID                 uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	PurchaseOrderID    uuid.UUID  `gorm:"type:uuid;not null;index" json:"purchase_order_id"`
	ProductID          uuid.UUID  `gorm:"type:uuid;not null" json:"product_id"`
	FactoryWarehouseID *uuid.UUID `gorm:"type:uuid" json:"factory_warehouse_id,omitempty"`
	RequestedQty       int        `gorm:"type:int;not null;default:0" json:"requested_qty"`
	ReceivedQty        int        `gorm:"type:int;default:0" json:"received_qty"`
	Unit               string     `gorm:"type:varchar(20);default:'KG'" json:"unit"`
	EstimatedUnitPrice float64    `gorm:"type:decimal(15,2);default:0" json:"estimated_unit_price"`
	Subtotal           float64    `gorm:"type:decimal(15,2);default:0" json:"subtotal"`
	Notes              string     `gorm:"type:text" json:"notes"`
	CreatedAt          time.Time  `gorm:"default:now()" json:"created_at"`
	UpdatedAt          time.Time  `gorm:"default:now()" json:"updated_at"`

	Product          *Product          `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	FactoryWarehouse *FactoryWarehouse `gorm:"foreignKey:FactoryWarehouseID" json:"factory_warehouse,omitempty"`
}

type RawMaterialReceipt struct {
	ID                     uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CompanyID              uuid.UUID  `gorm:"type:uuid;not null" json:"company_id"`
	FactoryID              uuid.UUID  `gorm:"type:uuid;not null" json:"factory_id"`
	PurchaseOrderID        *uuid.UUID `gorm:"type:uuid;index" json:"purchase_order_id,omitempty"`
	SupplierID             *uuid.UUID `gorm:"type:uuid;index" json:"supplier_id,omitempty"`
	SupplierName           string     `gorm:"type:varchar(255)" json:"supplier_name,omitempty"`            // Nama toko / supplier bebas untuk pembelian langsung tanpa PO
	ReceiptType            string     `gorm:"type:varchar(50);default:'RAW_MATERIAL'" json:"receipt_type"` // RAW_MATERIAL, TECHNICIAN_ITEM, MIXED
	ReceiptNo              string     `gorm:"type:varchar(100);not null;index" json:"receipt_no"`
	SupplierDeliveryNoteNo string     `gorm:"type:varchar(100)" json:"supplier_delivery_note_no"` // No Surat Jalan Supplier
	SupplierInvoiceNo      string     `gorm:"type:varchar(100)" json:"supplier_invoice_no,omitempty"`
	InvoiceDate            *time.Time `gorm:"type:date" json:"invoice_date,omitempty"`
	PaymentDueDate         *time.Time `gorm:"type:date" json:"payment_due_date,omitempty"`
	InvoiceTotal           float64    `gorm:"type:decimal(15,2);default:0" json:"invoice_total"`
	PaidAmount             float64    `gorm:"type:decimal(15,2);default:0" json:"paid_amount"`
	PaymentStatus          string     `gorm:"type:varchar(20);default:'UNPAID'" json:"payment_status"`
	PendingPaymentAmount   float64    `gorm:"type:decimal(15,2);default:0" json:"pending_payment_amount"`
	PaymentApprovalStatus  string     `gorm:"type:varchar(30);default:'NONE'" json:"payment_approval_status"`
	DeliveryDate           *time.Time `gorm:"type:date" json:"delivery_date,omitempty"`
	ReceivedDate           time.Time  `gorm:"type:date;not null" json:"received_date"`
	ReceivedByID           *uuid.UUID `gorm:"type:uuid" json:"received_by_id,omitempty"`
	ReceivedByName         string     `gorm:"type:varchar(100)" json:"received_by_name"`
	ProofImageURL          string     `gorm:"type:text" json:"proof_image_url,omitempty"`
	Notes                  string     `gorm:"type:text" json:"notes"`

	// Unloading / Ongkos Bongkar Kuli
	UnloadingCost    float64    `gorm:"type:decimal(15,2);default:0" json:"unloading_cost"`
	UnloadingNotes   string     `gorm:"type:text" json:"unloading_notes,omitempty"`
	KasPabrikID      *uuid.UUID `gorm:"type:uuid;index" json:"kas_pabrik_id,omitempty"`
	KasTransactionID *uuid.UUID `gorm:"type:uuid;index" json:"kas_transaction_id,omitempty"`

	CreatedAt time.Time `gorm:"default:now()" json:"created_at"`
	UpdatedAt time.Time `gorm:"default:now()" json:"updated_at"`

	Company       *Company                  `gorm:"foreignKey:CompanyID" json:"company,omitempty"`
	Factory       *Factory                  `gorm:"foreignKey:FactoryID" json:"factory,omitempty"`
	Supplier      *Supplier                 `gorm:"foreignKey:SupplierID" json:"supplier,omitempty"`
	PurchaseOrder *RawMaterialPurchaseOrder `gorm:"foreignKey:PurchaseOrderID" json:"purchase_order,omitempty"`
	KasPabrik     *KasPabrik                `gorm:"foreignKey:KasPabrikID" json:"kas_pabrik,omitempty"`
	Items         []RawMaterialReceiptItem  `gorm:"foreignKey:ReceiptID" json:"items,omitempty"`
}

type RawMaterialReceiptItem struct {
	ID                 uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ReceiptID          uuid.UUID  `gorm:"type:uuid;not null;index" json:"receipt_id"`
	POItemID           *uuid.UUID `gorm:"type:uuid;index" json:"po_item_id,omitempty"`
	ProductID          *uuid.UUID `gorm:"type:uuid;index" json:"product_id,omitempty"`
	TechnicianItemID   *uuid.UUID `gorm:"type:uuid;index" json:"technician_item_id,omitempty"`
	ItemType           string     `gorm:"type:varchar(50);default:'RAW_MATERIAL'" json:"item_type"` // RAW_MATERIAL, TECHNICIAN_ITEM
	FactoryWarehouseID *uuid.UUID `gorm:"type:uuid" json:"factory_warehouse_id,omitempty"`
	BatchNo            string     `gorm:"type:varchar(50);not null;default:'DEFAULT'" json:"batch_no"`
	ExpiryDate         *time.Time `json:"expiry_date,omitempty"`
	ReceivedQty        int        `gorm:"type:int;not null;default:0" json:"received_qty"`
	RejectedQty        int        `gorm:"type:int;default:0" json:"rejected_qty"`
	Unit               string     `gorm:"type:varchar(20);default:'KG'" json:"unit"`
	UnitPrice          float64    `gorm:"type:decimal(15,2);default:0" json:"unit_price"`
	Notes              string     `gorm:"type:text" json:"notes"`
	CreatedAt          time.Time  `gorm:"default:now()" json:"created_at"`

	Product          *Product          `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	TechnicianItem   *TechnicianItem   `gorm:"foreignKey:TechnicianItemID" json:"technician_item,omitempty"`
	FactoryWarehouse *FactoryWarehouse `gorm:"foreignKey:FactoryWarehouseID" json:"factory_warehouse,omitempty"`
}

func (r *RawMaterialPurchaseOrder) BeforeCreate(tx *gorm.DB) (err error) {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return
}

func (r *RawMaterialPurchaseOrderItem) BeforeCreate(tx *gorm.DB) (err error) {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return
}

func (r *RawMaterialReceipt) BeforeCreate(tx *gorm.DB) (err error) {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return
}

func (r *RawMaterialReceiptItem) BeforeCreate(tx *gorm.DB) (err error) {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return
}
