package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type KeyAccount struct {
	ID              uuid.UUID    `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CompanyID       uuid.UUID    `gorm:"type:uuid;not null" json:"company_id"`
	Name            string       `gorm:"type:varchar(255);not null" json:"name"`
	CompanyName     string       `gorm:"type:varchar(255);not null" json:"company_name"`
	Type            string       `gorm:"type:varchar(50);not null;default:'NATIONAL'" json:"type"` // NATIONAL, REGIONAL
	LogoColor       string       `gorm:"type:varchar(50);default:'bg-indigo-600'" json:"logo_color"`
	HQAddress       string       `gorm:"type:text" json:"hq_address"`
	PicName         string       `gorm:"type:varchar(255)" json:"pic_name"`
	PicPhone        string       `gorm:"type:varchar(50)" json:"pic_phone"`
	PicEmail        string       `gorm:"type:varchar(255)" json:"pic_email"`
	TotalStores     int          `gorm:"type:integer;default:0" json:"total_stores"`
	Status          string       `gorm:"type:varchar(50);default:'ACTIVE'" json:"status"` // ACTIVE, INACTIVE
	ActiveContracts int          `gorm:"-" json:"active_contracts"`
	CreatedAt       time.Time    `gorm:"default:now()" json:"created_at"`
	UpdatedAt       time.Time    `gorm:"default:now()" json:"updated_at"`
	Contracts       []MTContract `gorm:"foreignKey:KeyAccountID;constraint:OnDelete:CASCADE" json:"contracts,omitempty"`
}

func (ka *KeyAccount) BeforeCreate(tx *gorm.DB) (err error) {
	if ka.ID == uuid.Nil {
		ka.ID = uuid.New()
	}
	return
}

type MTContract struct {
	ID                           uuid.UUID   `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	KeyAccountID                 uuid.UUID   `gorm:"type:uuid;not null" json:"key_account_id"`
	StartDate                    time.Time   `gorm:"type:date;not null" json:"start_date"`
	EndDate                      time.Time   `gorm:"type:date;not null" json:"end_date"`
	PaymentTermDays              int         `gorm:"type:integer;default:30" json:"payment_term_days"`
	RebatePercent                float64     `gorm:"type:decimal(5,2);default:0" json:"rebate_percent"`
	MarketingContributionPercent float64     `gorm:"type:decimal(5,2);default:0" json:"marketing_contribution_percent"`
	ListingFeePerSku             float64     `gorm:"type:decimal(15,2);default:0" json:"listing_fee_per_sku"`
	ListedSkusCount              int         `gorm:"type:integer;default:0" json:"listed_skus_count"`
	ServiceLevelTarget           float64     `gorm:"type:decimal(5,2);default:95.0" json:"service_level_target"`
	Notes                        string      `gorm:"type:text" json:"notes"`
	CreatedAt                    time.Time   `gorm:"default:now()" json:"created_at"`
	UpdatedAt                    time.Time   `gorm:"default:now()" json:"updated_at"`
	KeyAccount                   *KeyAccount `gorm:"foreignKey:KeyAccountID" json:"key_account,omitempty"`
}

func (c *MTContract) BeforeCreate(tx *gorm.DB) (err error) {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return
}

type MTPurchaseOrder struct {
	ID               uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	KeyAccountID     uuid.UUID `gorm:"type:uuid;not null" json:"key_account_id"`
	PONumber         string    `gorm:"type:varchar(100);not null" json:"po_number"`
	OrderDate        time.Time `gorm:"type:date;not null" json:"order_date"`
	DeliveryDeadline time.Time `gorm:"type:date;not null" json:"delivery_deadline"`
	TotalAmount      float64   `gorm:"type:decimal(15,2);default:0" json:"total_amount"`
	ItemsOrdered     int       `gorm:"type:integer;default:0" json:"items_ordered"`
	ItemsShipped     int       `gorm:"type:integer;default:0" json:"items_shipped"`
	FillRate         float64   `gorm:"type:decimal(5,2);default:0" json:"fill_rate"`
	Status           string    `gorm:"type:varchar(50);default:'PENDING'" json:"status"` // PENDING, APPROVED, WAREHOUSE_RECEIVED, LOADING, IN_DELIVERY, DELIVERED_AND_UPLOADED, SETTLED
	DCDestination    string    `gorm:"type:varchar(255)" json:"dc_destination"`

	// New requested fields for dynamic unit
	UnitName     string  `gorm:"type:varchar(50);default:'Kardus'" json:"unit_name"`
	PcsPerUnit   int     `gorm:"type:integer;default:1" json:"pcs_per_unit"`
	FactoryPrice float64 `gorm:"type:decimal(15,2);default:0" json:"factory_price"`

	// New requested fields
	POSchedule      *time.Time `gorm:"type:date" json:"po_schedule"`
	InvoicePhotoURL string     `gorm:"type:varchar(255)" json:"invoice_photo_url"`
	ReceiptPhotoURL string     `gorm:"type:varchar(255)" json:"receipt_photo_url"`
	DeliveryNoteURL string     `gorm:"type:varchar(255)" json:"delivery_note_url"`
	BillingSchedule *time.Time `gorm:"type:date" json:"billing_schedule"`
	ReturnSchedule  *time.Time `gorm:"type:date" json:"return_schedule"`
	EventSchedule   *time.Time `gorm:"type:date" json:"event_schedule"`
	SpgSchedule     *time.Time `gorm:"type:date" json:"spg_schedule"`
	StockUpdateLog  string     `gorm:"type:text" json:"stock_update_log"`
	PromoUpdateLog  string     `gorm:"type:text" json:"promo_update_log"`
	SupermarketName string     `gorm:"type:varchar(255)" json:"supermarket_name"`
	MdName          string     `gorm:"type:varchar(255)" json:"md_name"`

	// Payment Term & Billing Info
	PaymentTerm      string     `gorm:"type:varchar(50);default:'TOP_30_DAYS'" json:"payment_term"` // TOP_30_DAYS, CASH_ON_DELIVERY, CONSIGNMENT_SETTLED
	DueDate          *time.Time `gorm:"type:date" json:"due_date"`
	BillingStatus    string     `gorm:"type:varchar(50);default:'UNBILLED'" json:"billing_status"` // UNBILLED, PENDING_SETTLEMENT, PAID_FULL
	PaymentProofURL  string     `gorm:"type:varchar(255)" json:"payment_proof_url"`
	CollectedAt      *time.Time `gorm:"type:timestamp" json:"collected_at,omitempty"`
	CollectedBy      string     `gorm:"type:varchar(255)" json:"collected_by,omitempty"`

	CreatedAt  time.Time   `gorm:"default:now()" json:"created_at"`
	UpdatedAt  time.Time   `gorm:"default:now()" json:"updated_at"`
	KeyAccount *KeyAccount `gorm:"foreignKey:KeyAccountID" json:"key_account,omitempty"`
	SalesOrder *SalesOrder `gorm:"foreignKey:PONumber;references:SONumber" json:"sales_order,omitempty"`
}

func (po *MTPurchaseOrder) BeforeCreate(tx *gorm.DB) (err error) {
	if po.ID == uuid.Nil {
		po.ID = uuid.New()
	}
	return
}

type MTMerchandising struct {
	ID                 uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	KeyAccountID       uuid.UUID `gorm:"type:uuid;not null" json:"key_account_id"`
	MerchandiserName   string    `gorm:"type:varchar(255);not null" json:"merchandiser_name"`
	StoreName          string    `gorm:"type:varchar(255);not null" json:"store_name"`
	VisitTime          time.Time `gorm:"type:timestamp;not null;default:now()" json:"visit_time"`
	PlanogramCompliant bool      `gorm:"type:boolean;default:true" json:"planogram_compliant"`
	ShelfSharePercent  float64   `gorm:"type:decimal(5,2);default:0" json:"shelf_share_percent"`
	OutOfStockSkus     string    `gorm:"type:text" json:"out_of_stock_skus"` // Stored as comma-separated or json string
	PhotoURL           string    `gorm:"type:varchar(255)" json:"photo_url"`
	Status             string    `gorm:"type:varchar(50);default:'COMPLIANT'" json:"status"` // COMPLIANT, NON_COMPLIANT

	// Alignment with requested checklist features
	InvoicePhotoURL string `gorm:"type:varchar(255)" json:"invoice_photo_url"`
	ReceiptPhotoURL string `gorm:"type:varchar(255)" json:"receipt_photo_url"`
	StockUpdateLog  string `gorm:"type:text" json:"stock_update_log"`
	PromoUpdateLog  string `gorm:"type:text" json:"promo_update_log"`
	SupermarketName string `gorm:"type:varchar(255)" json:"supermarket_name"`
	MdName          string `gorm:"type:varchar(255)" json:"md_name"`

	CreatedAt  time.Time   `gorm:"default:now()" json:"created_at"`
	UpdatedAt  time.Time   `gorm:"default:now()" json:"updated_at"`
	KeyAccount *KeyAccount `gorm:"foreignKey:KeyAccountID" json:"key_account,omitempty"`
}

func (m *MTMerchandising) BeforeCreate(tx *gorm.DB) (err error) {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return
}

type MTPromotion struct {
	ID             uuid.UUID   `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	KeyAccountID   uuid.UUID   `gorm:"type:uuid;not null" json:"key_account_id"`
	PromoType      string      `gorm:"type:varchar(50);not null" json:"promo_type"` // ENDCAP, GONDOLA, FLOOR_DISPLAY
	DisplayName    string      `gorm:"type:varchar(255);not null" json:"display_name"`
	StartDate      time.Time   `gorm:"type:date;not null" json:"start_date"`
	EndDate        time.Time   `gorm:"type:date;not null" json:"end_date"`
	Cost           float64     `gorm:"type:decimal(15,2);default:0" json:"cost"`
	SalesTarget    float64     `gorm:"type:decimal(15,2);default:0" json:"sales_target"`
	SalesActual    float64     `gorm:"type:decimal(15,2);default:0" json:"sales_actual"`
	Status         string      `gorm:"type:varchar(50);default:'APPROVED'" json:"status"` // APPROVED, ONGOING, COMPLETED
	OutletCoverage string      `gorm:"type:varchar(255)" json:"outlet_coverage"`
	CreatedAt      time.Time   `gorm:"default:now()" json:"created_at"`
	UpdatedAt      time.Time   `gorm:"default:now()" json:"updated_at"`
	KeyAccount     *KeyAccount `gorm:"foreignKey:KeyAccountID" json:"key_account,omitempty"`
}

func (p *MTPromotion) BeforeCreate(tx *gorm.DB) (err error) {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return
}

type MTSpg struct {
	ID              uuid.UUID   `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	KeyAccountID    uuid.UUID   `gorm:"type:uuid;not null" json:"key_account_id"`
	Name            string      `gorm:"type:varchar(255);not null" json:"name"`
	PhoneNumber     string      `gorm:"type:varchar(50)" json:"phone_number"`
	ShiftSchedule   string      `gorm:"type:varchar(255)" json:"shift_schedule"`
	Status          string      `gorm:"type:varchar(50);default:'ACTIVE'" json:"status"` // ACTIVE, INACTIVE
	TargetSales     float64     `gorm:"type:decimal(15,2);default:0" json:"target_sales"`
	ActualSales     float64     `gorm:"type:decimal(15,2);default:0" json:"actual_sales"`
	PhotoURL        string      `gorm:"type:varchar(255)" json:"photo_url"`
	SupermarketName string      `gorm:"type:varchar(255)" json:"supermarket_name"`
	CreatedAt       time.Time   `gorm:"default:now()" json:"created_at"`
	UpdatedAt       time.Time   `gorm:"default:now()" json:"updated_at"`
	KeyAccount      *KeyAccount `gorm:"foreignKey:KeyAccountID" json:"key_account,omitempty"`
}

func (spg *MTSpg) BeforeCreate(tx *gorm.DB) (err error) {
	if spg.ID == uuid.Nil {
		spg.ID = uuid.New()
	}
	return
}

func (MTSpg) TableName() string {
	return "mt_spgs"
}

type MTSpgDailyReport struct {
	ID                 uuid.UUID       `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ScheduleID         *uuid.UUID      `gorm:"type:uuid;index" json:"schedule_id,omitempty"`
	SpgID              uuid.UUID       `gorm:"type:uuid;not null" json:"spg_id"`
	KeyAccountID       uuid.UUID       `gorm:"type:uuid;not null" json:"key_account_id"`
	SupermarketName    string          `gorm:"type:varchar(255);not null" json:"supermarket_name"`
	ReportDate         time.Time       `gorm:"type:date;not null" json:"report_date"`
	Shift              string          `gorm:"type:varchar(100)" json:"shift"`
	TargetAmount       float64         `gorm:"type:decimal(15,2);default:0" json:"target_amount"`
	TotalAmount        float64         `gorm:"type:decimal(15,2);default:0" json:"total_amount"`
	TotalQty           int             `gorm:"type:integer;default:0" json:"total_qty"`
	BaseFee            float64         `gorm:"type:decimal(15,2);default:200000" json:"base_fee"`
	IncentiveBonus     float64         `gorm:"type:decimal(15,2);default:0" json:"incentive_bonus"`
	TotalSpgPayout     float64         `gorm:"type:decimal(15,2);default:200000" json:"total_spg_payout"`
	IsTargetAchieved   bool            `gorm:"type:boolean;default:false" json:"is_target_achieved"`
	ReceiptPhotoURL    string          `gorm:"type:varchar(255)" json:"receipt_photo_url"`
	AttendancePhotoURL string          `gorm:"type:varchar(255)" json:"attendance_photo_url"`
	Notes              string          `gorm:"type:text" json:"notes"`
	Status             string          `gorm:"type:varchar(50);default:'SUBMITTED'" json:"status"` // SUBMITTED, APPROVED, REJECTED
	VerifiedBy         *uuid.UUID      `gorm:"type:uuid" json:"verified_by,omitempty"`
	VerifiedAt         *time.Time      `gorm:"type:timestamp" json:"verified_at,omitempty"`
	CreatedAt          time.Time       `gorm:"default:now()" json:"created_at"`
	UpdatedAt          time.Time       `gorm:"default:now()" json:"updated_at"`

	Schedule   *MTSpgSchedule  `gorm:"foreignKey:ScheduleID" json:"schedule,omitempty"`
	Spg        *MTSpg          `gorm:"foreignKey:SpgID" json:"spg,omitempty"`
	KeyAccount *KeyAccount     `gorm:"foreignKey:KeyAccountID" json:"key_account,omitempty"`
	Items      []MTSpgSaleItem `gorm:"foreignKey:ReportID;constraint:OnDelete:CASCADE" json:"items,omitempty"`
}

func (r *MTSpgDailyReport) BeforeCreate(tx *gorm.DB) (err error) {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return
}

func (MTSpgDailyReport) TableName() string {
	return "mt_spg_daily_reports"
}

type MTSpgSaleItem struct {
	ID          uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ReportID    uuid.UUID  `gorm:"type:uuid;not null" json:"report_id"`
	ProductID   *uuid.UUID `gorm:"type:uuid" json:"product_id,omitempty"`
	ProductName string     `gorm:"type:varchar(255);not null" json:"product_name"`
	SkuCode     string     `gorm:"type:varchar(100)" json:"sku_code"`
	UnitPrice   float64    `gorm:"type:decimal(15,2);default:0" json:"unit_price"`
	Quantity    int        `gorm:"type:integer;default:1" json:"quantity"`
	Subtotal    float64    `gorm:"type:decimal(15,2);default:0" json:"subtotal"`
	CreatedAt   time.Time  `gorm:"default:now()" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"default:now()" json:"updated_at"`
}

func (i *MTSpgSaleItem) BeforeCreate(tx *gorm.DB) (err error) {
	if i.ID == uuid.Nil {
		i.ID = uuid.New()
	}
	return
}

func (MTSpgSaleItem) TableName() string {
	return "mt_spg_sale_items"
}

type MTSpgSchedule struct {
	ID              uuid.UUID   `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CompanyID       uuid.UUID   `gorm:"type:uuid;not null;index" json:"company_id"`
	SpgID           uuid.UUID   `gorm:"type:uuid;not null;index" json:"spg_id"`
	KeyAccountID    uuid.UUID   `gorm:"type:uuid;not null;index" json:"key_account_id"`
	SupermarketName string      `gorm:"type:varchar(255);not null" json:"supermarket_name"`
	StartDate       time.Time   `gorm:"type:date;not null" json:"start_date"`
	EndDate         time.Time   `gorm:"type:date;not null" json:"end_date"`
	DayPattern      string      `gorm:"type:varchar(50);default:'WEEKEND_FRI_SUN'" json:"day_pattern"` // WEEKEND_FRI_SUN, DAILY_MON_SAT, CUSTOM
	WorkingDaysCount int        `gorm:"type:integer;default:1" json:"working_days_count"`
	DailyFee        float64     `gorm:"type:decimal(15,2);default:0" json:"daily_fee"`
	AllowanceFee    float64     `gorm:"type:decimal(15,2);default:0" json:"allowance_fee"`
	TotalBudget     float64     `gorm:"type:decimal(15,2);default:0" json:"total_budget"`
	TargetSales     float64     `gorm:"type:decimal(15,2);default:0" json:"target_sales"`
	Status          string      `gorm:"type:varchar(50);default:'PLANNED'" json:"status"` // PLANNED, ACTIVE, COMPLETED, CANCELLED
	Notes           string      `gorm:"type:text" json:"notes"`
	CreatedAt       time.Time   `gorm:"default:now()" json:"created_at"`
	UpdatedAt       time.Time   `gorm:"default:now()" json:"updated_at"`

	Spg        *MTSpg      `gorm:"foreignKey:SpgID" json:"spg,omitempty"`
	KeyAccount *KeyAccount `gorm:"foreignKey:KeyAccountID" json:"key_account,omitempty"`
}

func (s *MTSpgSchedule) BeforeCreate(tx *gorm.DB) (err error) {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return
}

func (MTSpgSchedule) TableName() string {
	return "mt_spg_schedules"
}

type MTInventoryCategory string

const (
	MTInventoryCatPromoGift  MTInventoryCategory = "PROMO_GIFT"  // Toples, Mug, Piring Hadiah Promo
	MTInventoryCatDemoTool   MTInventoryCategory = "DEMO_TOOL"   // Solet, Nampan Demo, Booth SPG
	MTInventoryCatPackaging  MTInventoryCategory = "PACKAGING"   // Kardus Khusus, Tas Belanja MT
	MTInventoryCatGeneral    MTInventoryCategory = "GENERAL"     // Lain-lain
)

type MTInventory struct {
	ID            uuid.UUID           `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	BranchID      *uuid.UUID          `gorm:"type:uuid;index" json:"branch_id,omitempty"`
	CompanyID     *uuid.UUID          `gorm:"type:uuid;index" json:"company_id,omitempty"`
	Name          string              `gorm:"type:varchar(255);not null" json:"name"`
	Category      MTInventoryCategory `gorm:"type:varchar(50);default:'PROMO_GIFT'" json:"category"`
	Quantity      int                 `gorm:"type:integer;default:0" json:"quantity"`
	Unit          string              `gorm:"type:varchar(50);default:'PCS'" json:"unit"` // PCS, DUS, LUSIN, SET
	MinStockAlert int                 `gorm:"type:integer;default:10" json:"min_stock_alert"`
	PhotoURL      string              `gorm:"type:varchar(255)" json:"photo_url"`
	Description   string              `gorm:"type:text" json:"description"`
	CreatedAt     time.Time           `gorm:"default:now()" json:"created_at"`
	UpdatedAt     time.Time           `gorm:"default:now()" json:"updated_at"`

	Branch  *Branch  `gorm:"foreignKey:BranchID" json:"branch,omitempty"`
	Company *Company `gorm:"foreignKey:CompanyID" json:"company,omitempty"`
}

func (m *MTInventory) BeforeCreate(tx *gorm.DB) (err error) {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return
}

func (MTInventory) TableName() string {
	return "mt_inventories"
}

type MTInventoryLogType string

const (
	MTInventoryLogIN         MTInventoryLogType = "IN"         // Barang masuk dari suplier/pabrik
	MTInventoryLogOUT        MTInventoryLogType = "OUT"        // Barang keluar (alokasi SPG/Event/Store)
	MTInventoryLogADJUSTMENT MTInventoryLogType = "ADJUSTMENT" // Penyesuaian / Stock opname / Rusak / Hilang
	MTInventoryLogRETURN     MTInventoryLogType = "RETURN"     // Pengembalian sisa dari SPG/Store
)

type MTInventoryLog struct {
	ID              uuid.UUID          `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	MTInventoryID   uuid.UUID          `gorm:"type:uuid;not null;index" json:"mt_inventory_id"`
	Type            MTInventoryLogType `gorm:"type:varchar(50);not null" json:"type"`
	Quantity        int                `gorm:"type:integer;not null" json:"quantity"`
	QuantityBefore  int                `gorm:"type:integer;not null" json:"quantity_before"`
	QuantityAfter   int                `gorm:"type:integer;not null" json:"quantity_after"`
	ReferenceNo     string             `gorm:"type:varchar(100)" json:"reference_no"`
	RecipientType   string             `gorm:"type:varchar(50)" json:"recipient_type"` // SPG, OUTLET_MT, EVENT, GUDANG, LAINNYA
	RecipientName   string             `gorm:"type:varchar(255)" json:"recipient_name"`
	Notes           string             `gorm:"type:text" json:"notes"`
	CreatedBy       string             `gorm:"type:varchar(100)" json:"created_by"`
	CreatedAt       time.Time          `gorm:"default:now()" json:"created_at"`

	MTInventory *MTInventory `gorm:"foreignKey:MTInventoryID" json:"mt_inventory,omitempty"`
}

func (l *MTInventoryLog) BeforeCreate(tx *gorm.DB) (err error) {
	if l.ID == uuid.Nil {
		l.ID = uuid.New()
	}
	return
}

func (MTInventoryLog) TableName() string {
	return "mt_inventory_logs"
}

// ─── MT Store Stock & Consignment Models ─────────────────────────────────────

type MTStoreStock struct {
	ID                uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CompanyID         uuid.UUID  `gorm:"type:uuid;not null;index" json:"company_id"`
	StoreID           uuid.UUID  `gorm:"type:uuid;not null;index" json:"store_id"`
	ProductID         uuid.UUID  `gorm:"type:uuid;not null;index" json:"product_id"`
	Quantity          int        `gorm:"type:integer;default:0" json:"quantity"` // in Pcs
	MinStockThreshold int        `gorm:"type:integer;default:20" json:"min_stock_threshold"` // low stock alert threshold
	LastAdjustedAt    time.Time  `gorm:"default:now()" json:"last_adjusted_at"`
	CreatedAt         time.Time  `gorm:"default:now()" json:"created_at"`
	UpdatedAt         time.Time  `gorm:"default:now()" json:"updated_at"`

	Store   *Store   `gorm:"foreignKey:StoreID" json:"store,omitempty"`
	Product *Product `gorm:"foreignKey:ProductID" json:"product,omitempty"`
}

func (s *MTStoreStock) BeforeCreate(tx *gorm.DB) (err error) {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return
}

func (MTStoreStock) TableName() string {
	return "mt_store_stocks"
}

type MTStoreStockLog struct {
	ID             uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CompanyID      uuid.UUID `gorm:"type:uuid;not null;index" json:"company_id"`
	StoreID        uuid.UUID `gorm:"type:uuid;not null;index" json:"store_id"`
	ProductID      uuid.UUID `gorm:"type:uuid;not null;index" json:"product_id"`
	Type           string    `gorm:"type:varchar(50);not null" json:"type"` // INITIAL, DELIVERY_IN, SPG_SALE_OUT, ADJUSTMENT, RETURN
	Quantity       int       `gorm:"type:integer;not null" json:"quantity"`
	QuantityBefore int       `gorm:"type:integer;not null;default:0" json:"quantity_before"`
	QuantityAfter  int       `gorm:"type:integer;not null;default:0" json:"quantity_after"`
	ReferenceNo    string    `gorm:"type:varchar(100)" json:"reference_no"`
	Notes          string    `gorm:"type:text" json:"notes"`
	CreatedBy      string    `gorm:"type:varchar(255)" json:"created_by"`
	CreatedAt      time.Time `gorm:"default:now()" json:"created_at"`

	Store   *Store   `gorm:"foreignKey:StoreID" json:"store,omitempty"`
	Product *Product `gorm:"foreignKey:ProductID" json:"product,omitempty"`
}

func (l *MTStoreStockLog) BeforeCreate(tx *gorm.DB) (err error) {
	if l.ID == uuid.Nil {
		l.ID = uuid.New()
	}
	return
}

func (MTStoreStockLog) TableName() string {
	return "mt_store_stock_logs"
}



