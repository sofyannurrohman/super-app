package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DeliveryBatchStatus string

const (
	DeliveryBatchWaitingApproval    DeliveryBatchStatus = "WAITING_APPROVAL"    // Menunggu persetujuan Supervisor
	DeliveryBatchSupervisorApproved DeliveryBatchStatus = "SUPERVISOR_APPROVED" // Disetujui Supervisor, SJ terbit
	DeliveryBatchPicking            DeliveryBatchStatus = "PICKING"             // Driver sedang ambil barang di Gudang
	DeliveryBatchOnDelivery         DeliveryBatchStatus = "ON_DELIVERY"         // Barang sedang dikirim ke Customer
	DeliveryBatchCompleted          DeliveryBatchStatus = "COMPLETED"           // Semua nota sudah selesai
	DeliveryBatchWaitingCrosscheck  DeliveryBatchStatus = "WAITING_CROSSCHECK"  // Menunggu pengecekan selisih setoran

	// Legacy statuses
	DeliveryBatchWaitingAssignment DeliveryBatchStatus = "WAITING_ASSIGNMENT"
	DeliveryBatchPending           DeliveryBatchStatus = "PENDING"
)

type DeliveryBatch struct {
	ID              uuid.UUID           `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CompanyID       *uuid.UUID          `gorm:"type:uuid" json:"company_id"` // nullable: batch bisa lintas company
	BranchID        *uuid.UUID          `gorm:"type:uuid" json:"branch_id"`
	DriverID        *uuid.UUID          `gorm:"type:uuid" json:"driver_id"`
	DailyDriverID   *uuid.UUID          `gorm:"type:uuid" json:"daily_driver_id"` // Eksternal Daily Driver
	HelperID        *uuid.UUID          `gorm:"type:uuid" json:"helper_id"`
	DailyHelperID   *uuid.UUID          `gorm:"type:uuid" json:"daily_helper_id"` // Eksternal Daily Helper
	VehicleID       *uuid.UUID          `gorm:"type:uuid" json:"vehicle_id"`
	DeliveryOrderNo *string             `gorm:"-:migration;type:varchar(100);uniqueIndex" json:"delivery_order_no"`
	Status          DeliveryBatchStatus `gorm:"type:varchar(30);default:'WAITING_APPROVAL'" json:"status"`
	Notes           string              `gorm:"type:text" json:"notes"`
	DeliveryDate    *time.Time          `gorm:"type:date" json:"delivery_date,omitempty"`
	DeliveryFee     float64             `gorm:"type:decimal(15,2);default:0" json:"delivery_fee"`

	// QR Code / Barcode untuk Surat Jalan digital
	BarcodeData string `gorm:"type:varchar(255)" json:"barcode_data"`

	// Mileage Tracking
	DepartureMileage *int `gorm:"type:int" json:"departure_mileage,omitempty"`
	ReturnMileage    *int `gorm:"type:int" json:"return_mileage,omitempty"`

	// Approval & Assignment
	AdminNotaID  *uuid.UUID `gorm:"type:uuid" json:"admin_nota_id"`
	SupervisorID *uuid.UUID `gorm:"type:uuid" json:"supervisor_id"`
	ApprovedAt   *time.Time `json:"approved_at,omitempty"`
	AssignedAt   *time.Time `json:"assigned_at,omitempty"`

	// Timeline
	PickingStartedAt *time.Time `json:"picking_started_at,omitempty"` // Driver tiba di Gudang
	StartedAt        *time.Time `json:"started_at,omitempty"`         // Driver mulai kirim
	FinishedAt       *time.Time `json:"finished_at,omitempty"`        // Semua nota selesai

	// Rekonsiliasi Cash & Digital
	TotalCashCollected     float64    `gorm:"type:decimal(15,2);default:0" json:"total_cash_collected"`
	ReportedCashCollected  *float64   `gorm:"type:decimal(15,2)" json:"reported_cash_collected"`
	CrosscheckNotes        string     `gorm:"type:text" json:"crosscheck_notes"`
	TotalTransferCollected float64    `gorm:"type:decimal(15,2);default:0" json:"total_transfer_collected"`
	TotalTempoCollected    float64    `gorm:"type:decimal(15,2);default:0" json:"total_tempo_collected"`
	CashSettledAt          *time.Time `json:"cash_settled_at,omitempty"`

	CreatedAt time.Time `gorm:"default:now()" json:"created_at"`
	UpdatedAt time.Time `gorm:"default:now()" json:"updated_at"`

	Driver      *Employee             `gorm:"foreignKey:DriverID" json:"driver,omitempty"`
	DailyDriver *DailyWorker          `gorm:"foreignKey:DailyDriverID" json:"daily_driver,omitempty"`
	Helper      *Employee             `gorm:"foreignKey:HelperID" json:"helper,omitempty"`
	DailyHelper *DailyWorker          `gorm:"foreignKey:DailyHelperID" json:"daily_helper,omitempty"`
	Vehicle     *Vehicle              `gorm:"foreignKey:VehicleID" json:"vehicle,omitempty"`
	AdminNota   *Employee             `gorm:"foreignKey:AdminNotaID" json:"admin_nota,omitempty"`
	Supervisor  *Employee             `gorm:"foreignKey:SupervisorID" json:"supervisor,omitempty"`
	Items       []DeliveryItem        `gorm:"foreignKey:DeliveryBatchID" json:"items,omitempty"`
	Spares      []DeliverySpare       `gorm:"foreignKey:DeliveryBatchID" json:"spares,omitempty"`
	Samples     []DeliveryBatchSample `gorm:"foreignKey:DeliveryBatchID" json:"samples,omitempty"`
}

type DeliveryItemStatus string

const (
	DeliveryItemPending   DeliveryItemStatus = "PENDING"
	DeliveryItemDelivered DeliveryItemStatus = "DELIVERED"
	DeliveryItemFailed    DeliveryItemStatus = "FAILED"
)

// DeliveryItem sekarang merujuk ke SalesOrder, bukan SalesTransaction.
// Invoice (SalesTransaction) baru dibuat SETELAH barang diterima customer dan pembayaran ditagih.
type DeliveryItem struct {
	ID                    uuid.UUID          `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	DeliveryBatchID       uuid.UUID          `gorm:"type:uuid;not null" json:"delivery_batch_id"`
	SalesOrderID          *uuid.UUID         `gorm:"type:uuid" json:"sales_order_id"`
	IsCollectionOnly      bool               `gorm:"default:false" json:"is_collection_only"`
	IsSubsequent          bool               `gorm:"default:false" json:"is_subsequent"`
	SalesSubsequentItemID *uuid.UUID         `gorm:"type:uuid" json:"sales_subsequent_item_id"`
	SalesTransferID       *uuid.UUID         `gorm:"type:uuid" json:"sales_transfer_id,omitempty"`
	Sequence              int                `gorm:"type:int;not null" json:"sequence"`
	Status                DeliveryItemStatus `gorm:"type:varchar(20);default:'PENDING'" json:"status"`
	Notes                 string             `gorm:"type:text" json:"notes"`

	// Barcode per nota (bisa di-scan driver saat serahkan ke customer)
	BarcodeData string `gorm:"type:varchar(255)" json:"barcode_data"`

	// Konfirmasi pengiriman
	DeliveredAt       *time.Time `json:"delivered_at,omitempty"`
	ReceivedBy        string     `gorm:"type:varchar(100)" json:"received_by"`
	PODImageURL       string     `gorm:"type:text" json:"pod_image_url"`
	SignatureImageURL string     `gorm:"type:text" json:"signature_image_url"`

	// Penagihan pembayaran oleh Driver
	PaymentCollected   bool       `gorm:"default:false" json:"payment_collected"`
	PaymentAmount      float64    `gorm:"type:decimal(15,2);default:0" json:"payment_amount"`
	PaymentMethod      string     `gorm:"type:varchar(50)" json:"payment_method"`
	PaymentCollectedAt *time.Time `json:"payment_collected_at,omitempty"`

	// Legacy field — dijaga agar tidak break kode yang sudah ada
	SalesTransactionID *uuid.UUID `gorm:"type:uuid" json:"sales_transaction_id"`

	// Retur Barang — dicatat saat driver melakukan konfirmasi pengiriman
	ReturnQuantity int    `gorm:"type:int;default:0" json:"return_quantity"` // Total unit yang diretur
	ReturnReason   string `gorm:"type:text" json:"return_reason"`            // Alasan: RUSAK / TOLAK_CUSTOMER / LAINNYA
	HasReturn      bool   `gorm:"default:false" json:"has_return"`           // Shortcut flag untuk filter

	SalesOrder          *SalesOrder          `gorm:"foreignKey:SalesOrderID" json:"sales_order,omitempty"`
	SalesSubsequentItem *SalesSubsequentItem `gorm:"foreignKey:SalesSubsequentItemID" json:"sales_subsequent_item,omitempty"`
	SalesTransfer       *SalesTransfer       `gorm:"foreignKey:SalesTransferID" json:"sales_transfer,omitempty"`
	SalesTransaction    *SalesTransaction    `gorm:"foreignKey:SalesTransactionID" json:"sales_transaction,omitempty"`
	DeliveryBatch       *DeliveryBatch       `gorm:"foreignKey:DeliveryBatchID" json:"delivery_batch,omitempty"`
}

func (db *DeliveryBatch) BeforeCreate(tx *gorm.DB) (err error) {
	if db.ID == uuid.Nil {
		db.ID = uuid.New()
	}
	return
}

func (di *DeliveryItem) BeforeCreate(tx *gorm.DB) (err error) {
	if di.ID == uuid.Nil {
		di.ID = uuid.New()
	}
	return
}
