package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MTDeliveryStatus - status pengiriman MT
type MTDeliveryStatus string

const (
	MTDeliveryDraft     MTDeliveryStatus = "DRAFT"     // Dibuat tapi belum di-assign
	MTDeliveryAssigned  MTDeliveryStatus = "ASSIGNED"  // Sudah di-assign ke cabang, menunggu approval
	MTDeliveryApproved  MTDeliveryStatus = "APPROVED"  // Cabang menyetujui, sedang siapkan barang
	MTDeliveryLoading   MTDeliveryStatus = "LOADING"   // Barang sedang dimuat
	MTDeliveryShipped   MTDeliveryStatus = "SHIPPED"   // Barang dalam pengiriman
	MTDeliveryDelivered MTDeliveryStatus = "DELIVERED" // Barang sudah sampai tujuan MT
	MTDeliveryRejected  MTDeliveryStatus = "REJECTED"  // Cabang menolak request
	MTDeliveryCancelled MTDeliveryStatus = "CANCELLED" // Dibatalkan oleh pabrik/pusat
)

// MTDeliverySourceType - asal stok pengiriman
type MTDeliverySourceType string

const (
	MTDeliverySourceFactory MTDeliverySourceType = "FACTORY" // Stok dari pabrik langsung
	MTDeliverySourceBranch  MTDeliverySourceType = "BRANCH"  // Stok diambil dari gudang cabang
	MTDeliverySourceMixed   MTDeliverySourceType = "MIXED"   // Campuran pabrik dan cabang
)

// MTDeliveryOrder - pengiriman ke outlet Modern Trade, dibuat dari pabrik/pusat
type MTDeliveryOrder struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CompanyID    uuid.UUID `gorm:"type:uuid;not null" json:"company_id"`
	DONumber     string    `gorm:"-:migration;type:varchar(100);uniqueIndex" json:"do_number"`
	KeyAccountID uuid.UUID `gorm:"type:uuid;not null" json:"key_account_id"`
	TotalWeight  float64   `gorm:"type:decimal(10,2);default:0" json:"total_weight"` // Opsional

	// Mileage Tracking
	DepartureMileage *int `gorm:"type:int" json:"departure_mileage,omitempty"`
	ReturnMileage    *int `gorm:"type:int" json:"return_mileage,omitempty"`

	PONumber        string               `gorm:"type:varchar(100)" json:"po_number"` // Nomor PO dari MT
	SourceType      MTDeliverySourceType `gorm:"type:varchar(20);default:'BRANCH'" json:"source_type"`
	SourceFactoryID *uuid.UUID           `gorm:"type:uuid" json:"source_factory_id"` // jika sumber dari pabrik
	SourceBranchID  *uuid.UUID           `gorm:"type:uuid" json:"source_branch_id"`  // jika sumber dari cabang (wajib jika BRANCH/MIXED)
	Status          MTDeliveryStatus     `gorm:"type:varchar(20);default:'DRAFT'" json:"status"`
	DCDestination   string               `gorm:"type:varchar(255)" json:"dc_destination"` // Alamat DC / outlet MT
	ScheduledDate   *time.Time           `gorm:"type:date" json:"scheduled_date,omitempty"`
	ShippedAt       *time.Time           `gorm:"type:timestamptz" json:"shipped_at,omitempty"`
	DeliveredAt     *time.Time           `gorm:"type:timestamptz" json:"delivered_at,omitempty"`
	Notes           string               `gorm:"type:text" json:"notes"`
	RejectionReason string               `gorm:"type:text" json:"rejection_reason"`
	ApprovedBy      string               `gorm:"type:varchar(100)" json:"approved_by"` // nama user di cabang yang approve
	PreparedBy      string               `gorm:"type:varchar(100)" json:"prepared_by"` // nama user di cabang yang siapkan
	VehicleID       *uuid.UUID           `gorm:"type:uuid" json:"vehicle_id"`
	DriverID        *uuid.UUID           `gorm:"type:uuid" json:"driver_id"`
	SalesOrderID    *uuid.UUID           `gorm:"type:uuid" json:"sales_order_id"`     // Link ke Sales Order yang terbuat otomatis
	CreatedBy       string               `gorm:"type:varchar(100)" json:"created_by"` // nama user di pusat yang buat
	CreatedAt       time.Time            `gorm:"default:now()" json:"created_at"`
	UpdatedAt       time.Time            `gorm:"default:now()" json:"updated_at"`

	KeyAccount    *KeyAccount           `gorm:"foreignKey:KeyAccountID" json:"key_account,omitempty"`
	SourceFactory *Factory              `gorm:"foreignKey:SourceFactoryID" json:"source_factory,omitempty"`
	SourceBranch  *Branch               `gorm:"foreignKey:SourceBranchID" json:"source_branch,omitempty"`
	Vehicle       *Vehicle              `gorm:"foreignKey:VehicleID" json:"vehicle,omitempty"`
	Driver        *Employee             `gorm:"foreignKey:DriverID" json:"driver,omitempty"`
	SalesOrder    *SalesOrder           `gorm:"foreignKey:SalesOrderID" json:"sales_order,omitempty"`
	Items         []MTDeliveryOrderItem `gorm:"foreignKey:MTDeliveryOrderID" json:"items,omitempty"`
}

// MTDeliveryOrderItem - detail produk dalam pengiriman MT
type MTDeliveryOrderItem struct {
	ID                uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	MTDeliveryOrderID uuid.UUID `gorm:"type:uuid;not null;index" json:"mt_delivery_order_id"`
	ProductID         uuid.UUID `gorm:"type:uuid;not null" json:"product_id"`
	RequestedQty      int       `gorm:"type:int;default:0" json:"requested_qty"` // qty yang diminta oleh pusat
	PreparedQty       int       `gorm:"type:int;default:0" json:"prepared_qty"`  // qty yang disiapkan oleh cabang
	ShippedQty        int       `gorm:"type:int;default:0" json:"shipped_qty"`   // qty yang benar-benar dimuat
	Unit              string    `gorm:"type:varchar(20);default:'pcs'" json:"unit"`
	BatchNo           string    `gorm:"type:varchar(50);default:'DEFAULT'" json:"batch_no"`
	Notes             string    `gorm:"type:text" json:"notes"`
	CreatedAt         time.Time `gorm:"default:now()" json:"created_at"`

	Product *Product `gorm:"foreignKey:ProductID" json:"product,omitempty"`
}

func (o *MTDeliveryOrder) BeforeCreate(tx *gorm.DB) (err error) {
	if o.ID == uuid.Nil {
		o.ID = uuid.New()
	}
	return
}

func (i *MTDeliveryOrderItem) BeforeCreate(tx *gorm.DB) (err error) {
	if i.ID == uuid.Nil {
		i.ID = uuid.New()
	}
	return
}
