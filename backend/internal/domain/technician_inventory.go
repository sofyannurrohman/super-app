package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TechnicianCategory defines standard categories for maintenance & technician goods
type TechnicianCategory string

const (
	TechnicianCatMurBaut     TechnicianCategory = "MUR_BAUT"
	TechnicianCatSparepart   TechnicianCategory = "SPAREPART_MESIN"
	TechnicianCatKelistrikan TechnicianCategory = "KELISTRIKAN"
	TechnicianCatPelumas     TechnicianCategory = "PELUMAS"
	TechnicianCatTools       TechnicianCategory = "TOOLS"
	TechnicianCatPneumatik   TechnicianCategory = "PNEUMATIK_HIDROLIK"
	TechnicianCatLainnya     TechnicianCategory = "LAINNYA"
)

type TechnicianItem struct {
	ID                 uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CompanyID          uuid.UUID `gorm:"type:uuid;not null;index" json:"company_id"`
	FactoryID          uuid.UUID `gorm:"type:uuid;not null;index" json:"factory_id"`
	Name               string    `gorm:"type:varchar(255);not null" json:"name"`
	Code               string    `gorm:"type:varchar(100);not null;index" json:"code"` // SKU / Kode Part
	Category           string    `gorm:"type:varchar(100);not null;default:'SPAREPART_MESIN'" json:"category"`
	Unit               string    `gorm:"type:varchar(50);not null;default:'PCS'" json:"unit"`
	Quantity           int       `gorm:"type:int;not null;default:0" json:"quantity"`
	MinLimit           int       `gorm:"type:int;default:5" json:"min_limit"`
	RackLocation       string    `gorm:"type:varchar(100)" json:"rack_location"` // e.g., Rak A-02, Kotak Baut 3
	EstimatedUnitPrice float64   `gorm:"type:decimal(15,2);default:0" json:"estimated_unit_price"`
	Description        string    `gorm:"type:text" json:"description"`
	CreatedAt          time.Time `gorm:"default:now()" json:"created_at"`
	UpdatedAt          time.Time `gorm:"default:now()" json:"updated_at"`

	Factory *Factory `gorm:"foreignKey:FactoryID" json:"factory,omitempty"`
	Company *Company `gorm:"foreignKey:CompanyID" json:"company,omitempty"`
}

type TechnicianStockLog struct {
	ID               uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CompanyID        uuid.UUID `gorm:"type:uuid;not null;index" json:"company_id"`
	FactoryID        uuid.UUID `gorm:"type:uuid;not null;index" json:"factory_id"`
	TechnicianItemID uuid.UUID `gorm:"type:uuid;not null;index" json:"technician_item_id"`
	Type             string    `gorm:"type:varchar(20);not null" json:"type"` // IN, OUT, OPNAME
	Source           string    `gorm:"type:varchar(255);not null" json:"source"` // SUPPLIER_RECEIPT, USAGE, OPNAME_ADJUSTMENT, INITIAL
	Quantity         int       `gorm:"type:int;not null" json:"quantity"` // delta change
	CurrentStock     int       `gorm:"type:int;not null" json:"current_stock"`
	TechnicianName   string    `gorm:"type:varchar(100)" json:"technician_name"`
	MachineOrPurpose string    `gorm:"type:varchar(255)" json:"machine_or_purpose"` // e.g., Perbaikan Mesin Sealer Line 2
	ReferenceNo      string    `gorm:"type:varchar(100)" json:"reference_no"` // No SJ Supplier atau No Opname
	Notes            string    `gorm:"type:text" json:"notes"`
	CreatedAt        time.Time `gorm:"default:now()" json:"created_at"`

	TechnicianItem *TechnicianItem `gorm:"foreignKey:TechnicianItemID" json:"technician_item,omitempty"`
	Factory        *Factory        `gorm:"foreignKey:FactoryID" json:"factory,omitempty"`
}

func (t *TechnicianItem) BeforeCreate(tx *gorm.DB) (err error) {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return
}

func (t *TechnicianStockLog) BeforeCreate(tx *gorm.DB) (err error) {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return
}
