package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Factory struct {
	ID            uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CompanyID     uuid.UUID `gorm:"type:uuid;not null" json:"company_id"`
	BranchID      uuid.UUID `gorm:"type:uuid;not null" json:"branch_id"`
	Name          string    `gorm:"type:varchar(255);not null" json:"name"`
	WarehouseCode string    `gorm:"-:migration;type:varchar(50);uniqueIndex" json:"warehouse_code"` // Kode Gudang unik, misal: GD-A1, GD-B2
	Location      string    `gorm:"type:text" json:"location"`
	Description   string    `gorm:"type:text" json:"description"`
	CreatedAt     time.Time `gorm:"default:now()" json:"created_at"`
	UpdatedAt     time.Time `gorm:"default:now()" json:"updated_at"`

	Branch  *Branch  `gorm:"foreignKey:BranchID" json:"branch,omitempty"`
	Company *Company `gorm:"foreignKey:CompanyID" json:"company,omitempty"`
}

type Product struct {
	ID                      uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id" form:"id"`
	CompanyID               uuid.UUID  `gorm:"type:uuid;not null" json:"company_id" form:"company_id"`
	Name                    string     `gorm:"type:varchar(255);not null" json:"name" form:"name"`
	SKU                     string     `gorm:"-:migration;type:varchar(50);uniqueIndex;not null" json:"sku" form:"sku"`
	Unit                    string     `gorm:"type:varchar(20);not null" json:"unit" form:"unit"` // e.g., PCS, BOX, KG
	PcsPerUnit              int        `gorm:"type:int;default:1" json:"pcs_per_unit" form:"pcs_per_unit"`
	Category                string     `gorm:"type:varchar(100)" json:"category" form:"category"`
	Brand                   string     `gorm:"type:varchar(100)" json:"brand" form:"brand"`
	Weight                  float64    `gorm:"type:decimal(10,2);default:0" json:"weight" form:"weight"`
	WeightUnit              string     `gorm:"type:varchar(10);default:'KG'" json:"weight_unit" form:"weight_unit"`                            // GR, KG, TON
	Length                  float64    `gorm:"type:decimal(10,2);default:0" json:"length" form:"length"`                                       // cm
	Width                   float64    `gorm:"type:decimal(10,2);default:0" json:"width" form:"width"`                                         // cm
	Height                  float64    `gorm:"type:decimal(10,2);default:0" json:"height" form:"height"`                                       // cm
	Volume                  float64    `gorm:"type:decimal(10,4);default:0" json:"volume" form:"volume"`                                       // m³ (CBM) per kemasan/dus
	CostPrice               float64    `gorm:"type:decimal(15,2);default:0" json:"cost_price" form:"cost_price"`                               // HPP (Harga Pokok Penjualan)
	FactoryPrice            float64    `gorm:"type:decimal(15,2);default:0" json:"factory_price" form:"factory_price"`                         // Harga Pabrik ke Cabang (Transfer Pricing)
	MarketplaceFactoryPrice float64    `gorm:"type:decimal(15,2);default:0" json:"marketplace_factory_price" form:"marketplace_factory_price"` // Harga Pabrik ke Marketplace
	SellingPrice            float64    `gorm:"type:decimal(15,2);default:0" json:"selling_price" form:"selling_price"`                         // Harga Jual Cabang ke Retail (Default HET / CBP)
	DBP                     float64    `gorm:"type:decimal(15,2);default:0" json:"dbp" form:"dbp"`                                             // Distributor Buying Price
	WBP                     float64    `gorm:"type:decimal(15,2);default:0" json:"wbp" form:"wbp"`                                             // Wholesaler Buying Price
	RBP                     float64    `gorm:"type:decimal(15,2);default:0" json:"rbp" form:"rbp"`                                             // Retailer Buying Price
	Description             string     `gorm:"type:text" json:"description" form:"description"`
	Specs                   string     `gorm:"type:text" json:"specs" form:"specs"` // Detailed technical specs
	ImageURL                string     `gorm:"type:text" json:"image_url" form:"image_url"`
	ShelfLifeDays           int        `gorm:"type:int;default:0" json:"shelf_life_days" form:"shelf_life_days"`
	BPOMNumber              string     `gorm:"type:varchar(100)" json:"bpom_number" form:"bpom_number"`
	BPOMExpiryDate          *time.Time `gorm:"type:timestamptz" json:"bpom_expiry_date" form:"bpom_expiry_date"`
	HalalNumber             string     `gorm:"type:varchar(100)" json:"halal_number" form:"halal_number"`
	HalalExpiryDate         *time.Time `gorm:"type:timestamptz" json:"halal_expiry_date" form:"halal_expiry_date"`
	WarehouseStock          int        `gorm:"-" json:"warehouse_stock"`
	CreatedAt               time.Time  `gorm:"default:now()" json:"created_at"`
	UpdatedAt               time.Time  `gorm:"default:now()" json:"updated_at"`
}

type BranchProductPrice struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	BranchID  uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_branch_product" json:"branch_id"`
	ProductID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_branch_product" json:"product_id"`
	HET       float64   `gorm:"type:decimal(15,2);not null" json:"het"` // Harga Eceran Tertinggi untuk Cabang ini
	CreatedAt time.Time `gorm:"default:now()" json:"created_at"`
	UpdatedAt time.Time `gorm:"default:now()" json:"updated_at"`

	Branch  *Branch  `gorm:"foreignKey:BranchID" json:"branch,omitempty"`
	Product *Product `gorm:"foreignKey:ProductID" json:"product,omitempty"`
}

func (bpp *BranchProductPrice) BeforeCreate(tx *gorm.DB) (err error) {
	if bpp.ID == uuid.Nil {
		bpp.ID = uuid.New()
	}
	return
}

func (p *Product) GetWarehouseNumber() int {
	nameLower := strings.ToLower(p.Name)
	catLower := strings.ToLower(p.Category)

	if strings.Contains(nameLower, "kecap") || strings.Contains(catLower, "kecap") {
		return 1
	}
	if strings.Contains(nameLower, "saus") || strings.Contains(catLower, "saus") {
		return 2
	}
	if strings.Contains(nameLower, "minyak") || strings.Contains(catLower, "minyak") {
		return 3
	}
	if strings.Contains(nameLower, "sambal") || strings.Contains(catLower, "sambal") {
		return 4
	}
	if strings.Contains(nameLower, "bumbu") || strings.Contains(catLower, "bumbu") {
		return 5
	}
	if strings.Contains(nameLower, "mie") || strings.Contains(nameLower, "noodle") || strings.Contains(catLower, "mie") {
		return 6
	}
	if strings.Contains(nameLower, "tepung") || strings.Contains(catLower, "tepung") {
		return 7
	}
	if strings.Contains(nameLower, "gula") || strings.Contains(catLower, "gula") {
		return 8
	}
	if strings.Contains(nameLower, "beras") || strings.Contains(catLower, "beras") {
		return 9
	}
	if strings.Contains(nameLower, "garam") || strings.Contains(catLower, "garam") {
		return 10
	}
	if strings.Contains(nameLower, "penyedap") || strings.Contains(catLower, "penyedap") {
		return 11
	}
	return 12
}

type FactoryWarehouse struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	FactoryID   uuid.UUID `gorm:"type:uuid;not null" json:"factory_id"`
	Name        string    `gorm:"type:varchar(255);not null" json:"name"`
	Code        string    `gorm:"-:migration;type:varchar(50);uniqueIndex" json:"code"`
	Type        string    `gorm:"type:varchar(50)" json:"type"` // RAW_MATERIAL, FINISHED_GOODS, REJECT
	Description string    `gorm:"type:text" json:"description"`
	CreatedAt   time.Time `gorm:"default:now()" json:"created_at"`
	UpdatedAt   time.Time `gorm:"default:now()" json:"updated_at"`

	Factory *Factory `gorm:"foreignKey:FactoryID" json:"factory,omitempty"`
}

type FactoryStock struct {
	ID                 uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	FactoryID          uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex:idx_factory_product_batch" json:"factory_id"`
	ProductID          uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex:idx_factory_product_batch" json:"product_id"`
	FactoryWarehouseID *uuid.UUID `gorm:"type:uuid" json:"factory_warehouse_id"`
	BatchNo            string     `gorm:"type:varchar(50);not null;default:'DEFAULT';uniqueIndex:idx_factory_product_batch" json:"batch_no"`
	ExpiryDate         *time.Time `json:"expiry_date"`
	Quantity           int        `gorm:"type:int;default:0" json:"quantity"`
	WarehouseNumber    int        `gorm:"type:int;default:12" json:"warehouse_number"` // Deprecated: Use FactoryWarehouseID
	CreatedAt          time.Time  `gorm:"default:now()" json:"created_at"`
	UpdatedAt          time.Time  `gorm:"default:now()" json:"updated_at"`

	Factory          *Factory          `gorm:"foreignKey:FactoryID" json:"factory,omitempty"`
	FactoryWarehouse *FactoryWarehouse `gorm:"foreignKey:FactoryWarehouseID" json:"factory_warehouse,omitempty"`
	Product          *Product          `gorm:"foreignKey:ProductID" json:"product,omitempty"`
}

type ProductionLog struct {
	ID                uuid.UUID               `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	FactoryID         uuid.UUID               `gorm:"type:uuid;not null" json:"factory_id"`
	ProductID         uuid.UUID               `gorm:"type:uuid;not null" json:"product_id"`
	TargetWarehouseID *uuid.UUID              `gorm:"type:uuid" json:"target_warehouse_id"`
	BatchNo           string                  `gorm:"type:varchar(50);not null" json:"batch_no"`
	ExpiryDate        *time.Time              `json:"expiry_date"`
	EmployeeID        *uuid.UUID              `gorm:"type:uuid" json:"employee_id,omitempty"`
	OperatorName      string                  `gorm:"type:text" json:"operator_name"`
	Quantity          int                     `gorm:"type:int;not null" json:"quantity"`
	CartonCount       int                     `gorm:"type:int" json:"carton_count"`
	PiecesPerCarton   int                     `gorm:"type:int" json:"pieces_per_carton"`
	ProductionDate    time.Time               `gorm:"type:timestamptz;default:now()" json:"production_date"`
	Notes             string                  `gorm:"type:text" json:"notes"`
	BatchGroupID      *uuid.UUID              `gorm:"type:uuid;index" json:"batch_group_id,omitempty"`
	WarehouseNumber   int                     `gorm:"type:int;default:12" json:"warehouse_number"` // Deprecated
	CreatedAt         time.Time               `gorm:"default:now()" json:"created_at"`

	Factory          *Factory                `gorm:"foreignKey:FactoryID" json:"factory,omitempty"`
	FactoryWarehouse *FactoryWarehouse       `gorm:"foreignKey:TargetWarehouseID" json:"factory_warehouse,omitempty"`
	Product          *Product                `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	Employee         *Employee               `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
	Materials        []ProductionLogMaterial `gorm:"foreignKey:ProductionLogID;constraint:OnDelete:CASCADE" json:"materials,omitempty"`
	Operators        []ProductionLogOperator `gorm:"foreignKey:ProductionLogID;constraint:OnDelete:CASCADE" json:"operators,omitempty"`
}

type ProductionLogOperator struct {
	ID              uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ProductionLogID uuid.UUID `gorm:"type:uuid;not null;index" json:"production_log_id"`
	EmployeeID      *uuid.UUID `gorm:"type:uuid;index" json:"employee_id,omitempty"`
	OperatorName    string    `gorm:"type:varchar(255);not null" json:"operator_name"`
	IsManual        bool      `gorm:"type:boolean;default:false" json:"is_manual"`
	CreatedAt       time.Time `gorm:"default:now()" json:"created_at"`

	Employee *Employee `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
}

type ProductionLogMaterial struct {
	ID              uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ProductionLogID uuid.UUID `gorm:"type:uuid;not null;index" json:"production_log_id"`
	ProductID       uuid.UUID `gorm:"type:uuid;not null" json:"product_id"`
	BatchNo         string    `gorm:"type:varchar(50);not null;default:'DEFAULT'" json:"batch_no"`
	Quantity        int       `gorm:"type:int;not null" json:"quantity"`
	Notes           string    `gorm:"type:text" json:"notes"`
	CreatedAt       time.Time `gorm:"default:now()" json:"created_at"`

	Product *Product `gorm:"foreignKey:ProductID" json:"product,omitempty"`
}

type ProductTransfer struct {
	ID                      uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	FromFactoryID           uuid.UUID  `gorm:"type:uuid;not null" json:"from_factory_id"`
	ToBranchID              uuid.UUID  `gorm:"type:uuid;not null" json:"to_branch_id"`
	ProductID               uuid.UUID  `gorm:"type:uuid;not null" json:"product_id"`
	FulfillmentBranchID     *uuid.UUID `gorm:"type:uuid" json:"fulfillment_branch_id"`
	IsStockBorrowed         bool       `gorm:"default:false" json:"is_stock_borrowed"`
	ReplenishmentTransferID *uuid.UUID `gorm:"type:uuid" json:"replenishment_transfer_id"`
	BatchNo                 string     `gorm:"type:varchar(50);not null;default:'DEFAULT'" json:"batch_no"`
	ExpiryDate              *time.Time `json:"expiry_date"`
	Quantity                int        `gorm:"type:int;not null" json:"quantity"`
	TotalWeight             float64    `gorm:"type:decimal(10,2)" json:"total_weight"`
	TotalVolume             float64    `gorm:"type:decimal(10,4);default:0" json:"total_volume"`
	Status                  string     `gorm:"type:varchar(20);default:'REQUESTED'" json:"status"` // REQUESTED, APPROVED, SHIPPED, ARRIVED, RECEIVED, REJECTED
	DeliveryOrderNo         string     `gorm:"type:varchar(100)" json:"delivery_order_no"`
	Notes                   string     `gorm:"type:text" json:"notes"`
	ShippedAt               *time.Time `json:"shipped_at,omitempty"`
	ReceivedAt              *time.Time `json:"received_at,omitempty"`
	EstimatedArrival        *time.Time `json:"estimated_arrival,omitempty"`
	TargetShipmentDate      *time.Time `json:"target_shipment_date,omitempty"`
	ActualReceivedQuantity  int        `gorm:"type:int;default:0" json:"actual_received_quantity"`
	DamagedQuantity         int        `gorm:"type:int;default:0" json:"damaged_quantity"`
	InitiatedBy             string     `gorm:"type:varchar(20);default:'FACTORY'" json:"initiated_by"`
	Unit                    string     `gorm:"type:varchar(20);default:'PCS'" json:"unit"`
	PcsPerUnit              int        `gorm:"type:int;default:1" json:"pcs_per_unit"`
	RejectionReason         string     `gorm:"type:text" json:"rejection_reason"`
	PaymentMethod           string     `gorm:"type:varchar(50);default:'NET'" json:"payment_method"`    // NET, TEMPO, INTERNAL
	PaymentStatus           string     `gorm:"type:varchar(20);default:'UNPAID'" json:"payment_status"` // UNPAID, PAID, EXEMPT
	TotalAmount             float64    `gorm:"type:decimal(15,2);default:0" json:"total_amount"`        // Quantity * Product.FactoryPrice
	PaymentDueAt            *time.Time `json:"payment_due_at,omitempty"`
	TransferType            string     `gorm:"type:varchar(30);default:'COMMERCIAL'" json:"transfer_type"` // COMMERCIAL, INTERNAL_TRANSFER
	IsBillable              bool       `gorm:"default:true" json:"is_billable"`
	VehicleID               *uuid.UUID `gorm:"type:uuid" json:"vehicle_id"`
	DriverID                *uuid.UUID `gorm:"type:uuid" json:"driver_id"`
	ApprovedByID            *uuid.UUID `gorm:"type:uuid" json:"approved_by_id,omitempty"`
	ApprovedByName          string     `gorm:"type:varchar(100)" json:"approved_by_name,omitempty"`
	ApprovedByPosition      string     `gorm:"type:varchar(100)" json:"approved_by_position,omitempty"`
	ApprovedAt              *time.Time `json:"approved_at,omitempty"`
	CreatedAt               time.Time  `gorm:"default:now()" json:"created_at"`
	UpdatedAt               time.Time  `gorm:"default:now()" json:"updated_at"`

	FromFactory       *Factory                    `gorm:"foreignKey:FromFactoryID" json:"from_factory,omitempty"`
	ToBranch          *Branch                     `gorm:"foreignKey:ToBranchID" json:"to_branch,omitempty"`
	FulfillmentBranch *Branch                     `gorm:"foreignKey:FulfillmentBranchID" json:"fulfillment_branch,omitempty"`
	Product           *Product                    `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	Vehicle           *Vehicle                    `gorm:"foreignKey:VehicleID" json:"vehicle,omitempty"`
	Driver            *Employee                   `gorm:"foreignKey:DriverID;constraint:OnDelete:SET NULL;" json:"driver,omitempty"`
	ApprovedBy        *Employee                   `gorm:"foreignKey:ApprovedByID" json:"approved_by,omitempty"`
	Corrections       []ProductTransferCorrection `gorm:"foreignKey:ProductTransferID" json:"corrections,omitempty"`
}

type ProductTransferItem struct {
	ID         *uuid.UUID `json:"id,omitempty"`
	ProductID  uuid.UUID  `json:"product_id"`
	BatchNo    string     `json:"batch_no"`
	Quantity   int        `json:"quantity"`
	Unit       string     `json:"unit"`
	PcsPerUnit int        `json:"pcs_per_unit"`
}

type WarehouseStock struct {
	ID               uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	BranchID         uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex:idx_branch_product_type_batch" json:"branch_id"`
	ProductID        uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex:idx_branch_product_type_batch" json:"product_id"`
	StockType        string     `gorm:"type:varchar(20);default:'GOOD';uniqueIndex:idx_branch_product_type_batch" json:"stock_type"` // 'GOOD', 'QUARANTINE'
	BatchNo          string     `gorm:"type:varchar(50);not null;default:'DEFAULT';uniqueIndex:idx_branch_product_type_batch" json:"batch_no"`
	ExpiryDate       *time.Time `json:"expiry_date"`
	Quantity         int        `gorm:"type:int;default:0" json:"quantity"`
	ReservedQuantity int        `gorm:"type:int;default:0" json:"reserved_quantity"` // Stok ter-reserve oleh SO CONFIRMED
	MinLimit         int        `gorm:"type:int;default:0" json:"min_limit"`
	CreatedAt        time.Time  `gorm:"default:now()" json:"created_at"`
	UpdatedAt        time.Time  `gorm:"default:now()" json:"updated_at"`

	Branch  *Branch  `gorm:"foreignKey:BranchID" json:"branch,omitempty"`
	Product *Product `gorm:"foreignKey:ProductID" json:"product,omitempty"`
}

type WarehouseLog struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	BranchID    uuid.UUID `gorm:"type:uuid;not null" json:"branch_id"`
	ProductID   uuid.UUID `gorm:"type:uuid;not null" json:"product_id"`
	Type        string    `gorm:"type:varchar(20);not null" json:"type"` // IN, OUT
	Source      string    `gorm:"type:varchar(50)" json:"source"`        // FACTORY_TRANSFER, SALES, OPNAME
	BatchNo     string    `gorm:"type:varchar(50)" json:"batch_no"`
	Quantity    int       `gorm:"type:int;not null" json:"quantity"`
	ProcessedBy string    `gorm:"type:varchar(100)" json:"processed_by"` // Admin gudang yang memproses/bongkar muatan
	ApprovedBy  string    `gorm:"type:varchar(100)" json:"approved_by"`  // Admin yang approve (jika ada)
	CreatedAt   time.Time `gorm:"default:now()" json:"created_at"`

	Branch  *Branch  `gorm:"foreignKey:BranchID" json:"branch,omitempty"`
	Product *Product `gorm:"foreignKey:ProductID" json:"product,omitempty"`
}

type FactoryInventoryLog struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	FactoryID uuid.UUID `gorm:"type:uuid;not null" json:"factory_id"`
	ProductID uuid.UUID `gorm:"type:uuid;not null" json:"product_id"`
	Type      string    `gorm:"type:varchar(20);not null" json:"type"`  // IN, OUT
	Source    string    `gorm:"type:varchar(255)" json:"source"`        // PRODUCTION, TRANSFER, ADJUSTMENT, DIRECT_SALE
	BatchNo   string    `gorm:"type:varchar(50)" json:"batch_no"`
	Quantity  int       `gorm:"type:int;not null" json:"quantity"`
	CreatedAt time.Time `gorm:"default:now()" json:"created_at"`

	Factory *Factory `gorm:"foreignKey:FactoryID" json:"factory,omitempty"`
	Product *Product `gorm:"foreignKey:ProductID" json:"product,omitempty"`
}

// ProductionRecipe mendefinisikan resep/BOM untuk sebuah produk jadi.
type ProductionRecipe struct {
	ID                uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	FinishedProductID uuid.UUID `gorm:"-:migration;type:uuid;not null;uniqueIndex" json:"finished_product_id"`
	Description       string    `gorm:"type:text" json:"description"`
	CreatedAt         time.Time `gorm:"default:now()" json:"created_at"`
	UpdatedAt         time.Time `gorm:"default:now()" json:"updated_at"`

	FinishedProduct *Product               `gorm:"foreignKey:FinishedProductID" json:"finished_product,omitempty"`
	Items           []ProductionRecipeItem `gorm:"foreignKey:RecipeID" json:"items,omitempty"`
}

// ProductionRecipeItem mendefinisikan bahan baku yang dibutuhkan dalam resep.
type ProductionRecipeItem struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	RecipeID     uuid.UUID `gorm:"type:uuid;not null" json:"recipe_id"`
	RawProductID uuid.UUID `gorm:"type:uuid;not null" json:"raw_product_id"`
	Quantity     float64   `gorm:"type:decimal(10,4);not null" json:"quantity"` // Jumlah bahan per 1 unit barang jadi

	RawProduct *Product `gorm:"foreignKey:RawProductID" json:"raw_product,omitempty"`
}

type BranchPaymentLog struct {
	ID                uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ProductTransferID uuid.UUID `gorm:"type:uuid;not null" json:"product_transfer_id"`
	BranchID          uuid.UUID `gorm:"type:uuid;not null" json:"branch_id"`
	AmountPaid        float64   `gorm:"type:decimal(15,2);not null" json:"amount_paid"`
	PaymentDate       time.Time `gorm:"type:timestamptz;default:now()" json:"payment_date"`
	SourceOfFunds     string    `gorm:"type:varchar(50);default:'OMZET'" json:"source_of_funds"` // OMZET, CASH, TRANSFER
	Notes             string    `gorm:"type:text" json:"notes"`
	CreatedAt         time.Time `gorm:"default:now()" json:"created_at"`

	ProductTransfer *ProductTransfer `gorm:"foreignKey:ProductTransferID" json:"product_transfer,omitempty"`
	Branch          *Branch          `gorm:"foreignKey:BranchID" json:"branch,omitempty"`
}

type FactoryDashboardStats struct {
	TotalFactories   int               `json:"total_factories"`
	TotalProducts    int               `json:"total_products"`
	PendingShipments int               `json:"pending_shipments"`
	TodayProduction  int               `json:"today_production"`
	RecentTransfers  []ProductTransfer `json:"recent_transfers"`
}

type BranchReceivableSummary struct {
	BranchID                  uuid.UUID `json:"branch_id"`
	BranchName                string    `json:"branch_name"`
	TotalGoodInvoiced         float64   `json:"total_good_invoiced"`
	TotalDamagedDeduction     float64   `json:"total_damaged_deduction"`
	TotalReturnDeduction      float64   `json:"total_return_deduction"`
	TotalPaid                 float64   `json:"total_paid"`
	GoodReceivable            float64   `json:"good_receivable"`
	FinancialReceivable       float64   `json:"financial_receivable"`
	TotalReceivable           float64   `json:"total_receivable"`
	GoodPaymentThisMonth      float64   `json:"good_payment_this_month"`
	FinancialPaymentThisMonth float64   `json:"financial_payment_this_month"`
	TotalPaymentThisMonth     float64   `json:"total_payment_this_month"`
}

type FactoryReceivablesResponse struct {
	Summary struct {
		TotalReceivables          float64 `json:"total_receivables"`
		TotalGoodReceivables      float64 `json:"total_good_receivables"`
		TotalGoodInvoiced         float64 `json:"total_good_invoiced"`
		TotalDamagedDeductions    float64 `json:"total_damaged_deductions"`
		TotalReturnDeductions     float64 `json:"total_return_deductions"`
		TotalFinancialReceivables float64 `json:"total_financial_receivables"`
		TotalPaymentsThisMonth    float64 `json:"total_payments_this_month"`
		UnpaidBranchesCount       int     `json:"unpaid_branches_count"`
	} `json:"summary"`
	Branches []BranchReceivableSummary `json:"branches"`
}

func (f *Factory) BeforeCreate(tx *gorm.DB) (err error) {
	if f.ID == uuid.Nil {
		f.ID = uuid.New()
	}
	return
}

func (fw *FactoryWarehouse) BeforeCreate(tx *gorm.DB) (err error) {
	if fw.ID == uuid.Nil {
		fw.ID = uuid.New()
	}
	return
}

func (p *Product) BeforeCreate(tx *gorm.DB) (err error) {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return
}

func (fs *FactoryStock) BeforeCreate(tx *gorm.DB) (err error) {
	if fs.ID == uuid.Nil {
		fs.ID = uuid.New()
	}
	return
}

func (pl *ProductionLog) BeforeCreate(tx *gorm.DB) (err error) {
	if pl.ID == uuid.Nil {
		pl.ID = uuid.New()
	}
	return
}

func (pt *ProductTransfer) BeforeCreate(tx *gorm.DB) (err error) {
	if pt.ID == uuid.Nil {
		pt.ID = uuid.New()
	}
	return
}

func (ws *WarehouseStock) BeforeCreate(tx *gorm.DB) (err error) {
	if ws.ID == uuid.Nil {
		ws.ID = uuid.New()
	}
	return
}

func (wl *WarehouseLog) BeforeCreate(tx *gorm.DB) (err error) {
	if wl.ID == uuid.Nil {
		wl.ID = uuid.New()
	}
	return
}

func (fil *FactoryInventoryLog) BeforeCreate(tx *gorm.DB) (err error) {
	if fil.ID == uuid.Nil {
		fil.ID = uuid.New()
	}
	return
}

func (pr *ProductionRecipe) BeforeCreate(tx *gorm.DB) (err error) {
	if pr.ID == uuid.Nil {
		pr.ID = uuid.New()
	}
	return
}

func (pri *ProductionRecipeItem) BeforeCreate(tx *gorm.DB) (err error) {
	if pri.ID == uuid.Nil {
		pri.ID = uuid.New()
	}
	return
}

func (bpl *BranchPaymentLog) BeforeCreate(tx *gorm.DB) (err error) {
	if bpl.ID == uuid.Nil {
		bpl.ID = uuid.New()
	}
	return
}

type LowStockAlert struct {
	WarehouseStockID uuid.UUID `json:"warehouse_stock_id"`
	BranchID         uuid.UUID `json:"branch_id"`
	BranchName       string    `json:"branch_name"`
	ProductID        uuid.UUID `json:"product_id"`
	ProductName      string    `json:"product_name"`
	ProductSKU       string    `json:"product_sku"`
	ProductUnit      string    `json:"product_unit"`
	PcsPerUnit       int       `json:"pcs_per_unit"`
	Quantity         int       `json:"quantity"`
	MinLimit         int       `json:"min_limit"`
}

type FactoryReturn struct {
	ID                     uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ReturnNo               string     `gorm:"-:migration;type:varchar(100);column:return_no" json:"return_no"`
	FromBranchID           uuid.UUID  `gorm:"type:uuid;not null" json:"from_branch_id"`
	ToFactoryID            uuid.UUID  `gorm:"type:uuid;not null" json:"to_factory_id"`
	Status                 string     `gorm:"type:varchar(20);default:'SHIPPED'" json:"status"` // REQUESTED, APPROVED, SHIPPED, RECEIVED
	Notes                  string     `gorm:"type:text" json:"notes"`
	IsReceived             bool       `gorm:"default:false" json:"is_received"`
	ReceivedBy             *string    `gorm:"type:varchar(255)" json:"received_by"`
	ReceivedAt             *time.Time `json:"received_at"`
	ReceiverName           string     `gorm:"type:varchar(100)" json:"receiver_name,omitempty"`
	InspectedByName        string     `gorm:"type:varchar(100)" json:"inspected_by_name,omitempty"`
	InspectionNotes        string     `gorm:"type:text" json:"inspection_notes,omitempty"`
	TotalReceivedQty       int        `gorm:"type:int;default:0" json:"total_received_qty"`
	TotalGoodQty           int        `gorm:"type:int;default:0" json:"total_good_qty"`
	TotalBottleDamagedQty  int        `gorm:"type:int;default:0" json:"total_bottle_damaged_qty"`
	TotalContentDamagedQty int        `gorm:"type:int;default:0" json:"total_content_damaged_qty"`
	// Mileage Tracking
	DepartureMileage       *int       `gorm:"type:int" json:"departure_mileage,omitempty"`
	ReturnMileage          *int       `gorm:"type:int" json:"return_mileage,omitempty"`
	VehicleID              *uuid.UUID `gorm:"type:uuid" json:"vehicle_id"`
	DriverID               *uuid.UUID `gorm:"type:uuid" json:"driver_id"`
	ShippedAt              *time.Time `json:"shipped_at"`
	ApprovedByID           *uuid.UUID `gorm:"type:uuid" json:"approved_by_id,omitempty"`
	ApprovedByName         string     `gorm:"type:varchar(100)" json:"approved_by_name,omitempty"`
	ApprovedAt             *time.Time `json:"approved_at,omitempty"`
	RejectionReason        string     `gorm:"type:text" json:"rejection_reason,omitempty"`
	CreatedAt              time.Time  `gorm:"default:now()" json:"created_at"`
	UpdatedAt              time.Time  `gorm:"default:now()" json:"updated_at"`

	FromBranch *Branch             `gorm:"foreignKey:FromBranchID" json:"from_branch,omitempty"`
	ToFactory  *Factory            `gorm:"foreignKey:ToFactoryID" json:"to_factory,omitempty"`
	Vehicle    *Vehicle            `gorm:"foreignKey:VehicleID" json:"vehicle,omitempty"`
	Driver     *Employee           `gorm:"foreignKey:DriverID" json:"driver,omitempty"`
	ApprovedBy *Employee           `gorm:"foreignKey:ApprovedByID" json:"approved_by,omitempty"`
	Items      []FactoryReturnItem `gorm:"foreignKey:FactoryReturnID" json:"items,omitempty"`
}

type FactoryReturnItem struct {
	ID                         uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	FactoryReturnID            uuid.UUID  `gorm:"type:uuid;not null;index" json:"factory_return_id"`
	ProductID                  uuid.UUID  `gorm:"type:uuid;not null" json:"product_id"`
	BatchNo                    string     `gorm:"type:varchar(50);not null;default:'DEFAULT'" json:"batch_no"`
	Quantity                   int        `gorm:"type:int;not null" json:"quantity"`
	Unit                       string     `gorm:"type:varchar(20);default:'PCS'" json:"unit"`
	PcsPerUnit                 int        `gorm:"type:int;default:1" json:"pcs_per_unit"`
	ReturnType                 string     `gorm:"type:varchar(30);default:'DAMAGED'" json:"return_type"` // GOOD, DAMAGED
	DamageType                 string     `gorm:"type:varchar(50);default:'OTHER'" json:"damage_type"`   // BOTTLE_PACKAGING, PRODUCT_CONTENT, EXPIRED, TOTAL_DAMAGE, OTHER
	DamageReason               string     `gorm:"type:text" json:"damage_reason,omitempty"`
	ExpiryDate                 *time.Time `json:"expiry_date,omitempty"`
	ReceivedQuantity           int        `gorm:"type:int;default:0" json:"received_quantity"`
	ReceivedGoodQty            int        `gorm:"type:int;default:0" json:"received_good_qty"`
	ReceivedBottleDamagedQty   int        `gorm:"type:int;default:0" json:"received_bottle_damaged_qty"`
	ReceivedContentDamagedQty  int        `gorm:"type:int;default:0" json:"received_content_damaged_qty"`
	DispositionAction          string     `gorm:"type:varchar(50);default:'PENDING'" json:"disposition_action"` // PENDING, RESTOCK_GOOD, REPACK_BOTTLE, SCRAP_DISPOSAL, PARTIAL_MIXED
	InspectionNotes            string     `gorm:"type:text" json:"inspection_notes,omitempty"`
	TargetFactoryWarehouseID   *uuid.UUID `gorm:"type:uuid" json:"target_factory_warehouse_id,omitempty"`

	Product          *Product          `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	FactoryWarehouse *FactoryWarehouse `gorm:"foreignKey:TargetFactoryWarehouseID" json:"factory_warehouse,omitempty"`
}

func (fr *FactoryReturn) BeforeCreate(tx *gorm.DB) (err error) {
	if fr.ID == uuid.Nil {
		fr.ID = uuid.New()
	}
	return
}

func (fri *FactoryReturnItem) BeforeCreate(tx *gorm.DB) (err error) {
	if fri.ID == uuid.Nil {
		fri.ID = uuid.New()
	}
	return
}

type ProductTransferCorrection struct {
	ID                 uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ProductTransferID  uuid.UUID `gorm:"type:uuid;not null;index" json:"product_transfer_id"`
	CorrectedBy        string    `gorm:"type:varchar(100);not null" json:"corrected_by"`
	OldActualQuantity  int       `gorm:"type:int;not null" json:"old_actual_quantity"`
	NewActualQuantity  int       `gorm:"type:int;not null" json:"new_actual_quantity"`
	OldDamagedQuantity int       `gorm:"type:int;not null" json:"old_damaged_quantity"`
	NewDamagedQuantity int       `gorm:"type:int;not null" json:"new_damaged_quantity"`
	CorrectionReason   string    `gorm:"type:text;not null" json:"correction_reason"`
	CreatedAt          time.Time `gorm:"default:now()" json:"created_at"`

	ProductTransfer *ProductTransfer `gorm:"foreignKey:ProductTransferID" json:"product_transfer,omitempty"`
}

func (ptc *ProductTransferCorrection) BeforeCreate(tx *gorm.DB) (err error) {
	if ptc.ID == uuid.Nil {
		ptc.ID = uuid.New()
	}
	return
}

type FactoryQcSample struct {
	ID              uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	FactoryID       uuid.UUID  `gorm:"type:uuid;not null;index" json:"factory_id"`
	ProductID       uuid.UUID  `gorm:"type:uuid;not null;index" json:"product_id"`
	ProductionLogID *uuid.UUID `gorm:"type:uuid;index" json:"production_log_id,omitempty"`
	BatchNo         string     `gorm:"type:varchar(50);not null" json:"batch_no"`
	ExpiryDate      *time.Time `json:"expiry_date,omitempty"`
	ProductionDate  *time.Time `json:"production_date,omitempty"`
	Quantity        int        `gorm:"type:int;not null;default:1" json:"quantity"`
	SampleType      string     `gorm:"type:varchar(50);default:'LAB_QC'" json:"sample_type"` // LAB_QC, RETAINED_SAMPLE, STABILITY_TEST
	Purpose         string     `gorm:"type:text" json:"purpose"`
	Status          string     `gorm:"type:varchar(20);default:'PENDING'" json:"status"` // PENDING, APPROVED, REJECTED
	RequestedByID   uuid.UUID  `gorm:"type:uuid;not null" json:"requested_by_id"`
	RequestedByName string     `gorm:"type:varchar(100);not null" json:"requested_by_name"`
	RequestedAt     time.Time  `gorm:"default:now()" json:"requested_at"`
	ApprovedByID    *uuid.UUID `gorm:"type:uuid" json:"approved_by_id,omitempty"`
	ApprovedByName  string     `gorm:"type:varchar(100)" json:"approved_by_name,omitempty"`
	ApprovedAt      *time.Time `json:"approved_at,omitempty"`
	RejectionReason string     `gorm:"type:text" json:"rejection_reason,omitempty"`
	Notes           string     `gorm:"type:text" json:"notes,omitempty"`
	CreatedAt       time.Time  `gorm:"default:now()" json:"created_at"`
	UpdatedAt       time.Time  `gorm:"default:now()" json:"updated_at"`

	Factory       *Factory       `gorm:"foreignKey:FactoryID" json:"factory,omitempty"`
	Product       *Product       `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	ProductionLog *ProductionLog `gorm:"foreignKey:ProductionLogID" json:"production_log,omitempty"`
}

func (s *FactoryQcSample) BeforeCreate(tx *gorm.DB) (err error) {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	if s.RequestedAt.IsZero() {
		s.RequestedAt = time.Now()
	}
	return
}

