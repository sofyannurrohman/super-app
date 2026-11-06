package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PaymentStatus string

const (
	PaymentStatusUnpaid  PaymentStatus = "UNPAID"
	PaymentStatusPartial PaymentStatus = "PARTIAL"
	PaymentStatusPaid    PaymentStatus = "PAID"
)

const (
	SalesChannelCanvas             = "CANVAS"
	SalesChannelMotoris            = "MOTORIS"
	SalesChannelDirectFactory      = "DIRECT_FACTORY"
	SalesChannelDirectBranch       = "DIRECT_BRANCH"
	SalesChannelEmployeeSale       = "EMPLOYEE_SALE"
	SalesChannelDistributorSellOut = "DISTRIBUTOR_SELLOUT"

	BuyerTypeRegisteredStore = "REGISTERED_STORE"
	BuyerTypeGeneralWalkIn   = "GENERAL_WALKIN"
	BuyerTypeEmployee        = "EMPLOYEE"
)

type SalesTransaction struct {
	ID                   uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ClientID             *string    `gorm:"type:varchar(50);uniqueIndex" json:"client_id"` // Idempotency key from mobile
	VisitID              *uuid.UUID `gorm:"type:uuid" json:"visit_id"`                     // Mengikat transaksi ke log check-in
	CompanyID            uuid.UUID  `gorm:"type:uuid;not null;index" json:"company_id"`
	StoreID              *uuid.UUID `gorm:"type:uuid;index" json:"store_id,omitempty"`
	EmployeeID           uuid.UUID  `gorm:"type:uuid;index" json:"employee_id"`
	UserID               *uuid.UUID `gorm:"type:uuid" json:"user_id,omitempty"`
	DistributorStoreID   *uuid.UUID `gorm:"type:uuid;index" json:"distributor_store_id,omitempty"`
	IsDistributorOrder   bool       `gorm:"default:false;index" json:"is_distributor_order"`
	ReceiptNo            string     `gorm:"type:varchar(100)" json:"receipt_no"`
	ReceiptImageURL      string     `gorm:"type:text" json:"receipt_image_url"`
	SignatureImageURL    string     `gorm:"type:text" json:"signature_image_url"`
	TotalAmount          float64    `gorm:"type:decimal(15,2);not null" json:"total_amount"`
	DiscountAmount       float64    `gorm:"type:decimal(15,2);default:0" json:"discount_amount"`
	StoreCategory        string     `gorm:"type:varchar(50)" json:"store_category"` // 'TOKO_LAMA', 'TOKO_BARU'
	PeriodMonth          int        `gorm:"type:int;not null" json:"period_month"`
	PeriodYear           int        `gorm:"type:int;not null" json:"period_year"`
	TransactionDate      time.Time  `gorm:"type:date;not null;index" json:"transaction_date"`
	Status               string     `gorm:"type:varchar(50);default:'PENDING';index" json:"status"` // 'PENDING', 'VERIFIED'
	VerifiedAt           *time.Time `json:"verified_at,omitempty"`
	VerifiedByID         *uuid.UUID `gorm:"type:uuid" json:"verified_by_id,omitempty"`
	PaidAmount           float64    `gorm:"type:decimal(15,2);default:0" json:"paid_amount"`
	PaymentStatus        string     `gorm:"type:varchar(20);default:'UNPAID';index" json:"payment_status"`
	PaymentDueDate       *time.Time `gorm:"type:date" json:"payment_due_date"`
	PaymentMethod        string     `gorm:"type:varchar(20);default:'CASH';index" json:"payment_method"` // 'CASH', 'QRIS', 'TEMPO', 'VA', 'POTONG_GAJI'
	CompanyBankAccountID *uuid.UUID `gorm:"type:uuid" json:"company_bank_account_id"`
	MidtransID           string     `gorm:"type:varchar(100)" json:"midtrans_id"`
	MidtransQRISURL      string     `gorm:"type:text" json:"midtrans_qris_url"`
	MidtransVANumber     *string    `gorm:"type:varchar(50)" json:"midtrans_va_number"`
	MidtransBank         *string    `gorm:"type:varchar(20)" json:"midtrans_bank"`
	MidtransBillKey      *string    `gorm:"type:varchar(50)" json:"midtrans_bill_key"`
	MidtransBillerCode   *string    `gorm:"type:varchar(20)" json:"midtrans_biller_code"`
	Tenor                int        `gorm:"type:int;default:1" json:"tenor"`
	Notes                *string    `gorm:"type:text" json:"notes"`
	Latitude             *float64   `gorm:"type:decimal(10,8)" json:"latitude"`
	Longitude            *float64   `gorm:"type:decimal(11,8)" json:"longitude"`

	// Direct / Walk-in / Factory Sales support
	SalesChannel      string     `gorm:"type:varchar(50);default:'CANVAS'" json:"sales_channel"`
	BuyerType         string     `gorm:"type:varchar(50);default:'REGISTERED_STORE'" json:"buyer_type"`
	BuyerName         string     `gorm:"type:varchar(255)" json:"buyer_name"`
	BuyerPhone        string     `gorm:"type:varchar(50)" json:"buyer_phone"`
	BuyerEmployeeID   *uuid.UUID `gorm:"type:uuid" json:"buyer_employee_id,omitempty"`
	CashierEmployeeID *uuid.UUID `gorm:"type:uuid" json:"cashier_employee_id,omitempty"`
	KasPabrikID       *uuid.UUID `gorm:"type:uuid" json:"kas_pabrik_id,omitempty"`
	BranchID          *uuid.UUID `gorm:"type:uuid;index" json:"branch_id,omitempty"`

	CreatedAt      time.Time      `gorm:"default:now();index" json:"created_at"`
	UpdatedAt      time.Time      `gorm:"default:now()" json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	DeletedByID    *uuid.UUID     `gorm:"type:uuid" json:"deleted_by_id,omitempty"`
	DeletionReason *string        `gorm:"type:text" json:"deletion_reason,omitempty"`

	Company            *Company            `gorm:"foreignKey:CompanyID;constraint:OnDelete:CASCADE;" json:"company,omitempty"`
	Store              *Store              `gorm:"foreignKey:StoreID;constraint:OnDelete:CASCADE;" json:"store,omitempty"`
	Employee           *Employee           `gorm:"foreignKey:EmployeeID;constraint:OnDelete:CASCADE;" json:"employee,omitempty"`
	User               *User               `gorm:"foreignKey:UserID" json:"user,omitempty"`
	DistributorStore   *Store              `gorm:"foreignKey:DistributorStoreID" json:"distributor_store,omitempty"`
	BuyerEmployee      *Employee           `gorm:"foreignKey:BuyerEmployeeID" json:"buyer_employee,omitempty"`
	CashierEmployee    *Employee           `gorm:"foreignKey:CashierEmployeeID" json:"cashier_employee,omitempty"`
	KasPabrik          *KasPabrik          `gorm:"foreignKey:KasPabrikID" json:"kas_pabrik,omitempty"`
	Branch             *Branch             `gorm:"foreignKey:BranchID" json:"branch,omitempty"`
	VerifiedBy         *Employee           `gorm:"foreignKey:VerifiedByID" json:"verified_by,omitempty"`
	DeletedBy          *Employee           `gorm:"foreignKey:DeletedByID" json:"deleted_by,omitempty"`
	Visit              *SalesVisit         `gorm:"foreignKey:VisitID" json:"visit,omitempty"`
	CompanyBankAccount *CompanyBankAccount `gorm:"foreignKey:CompanyBankAccountID" json:"company_bank_account,omitempty"`

	DeliveryItems  []DeliveryItem                  `gorm:"foreignKey:SalesTransactionID;constraint:OnDelete:CASCADE;" json:"delivery_items,omitempty"`
	Items          []SalesItem                     `gorm:"foreignKey:SalesTransactionID;constraint:OnDelete:CASCADE;" json:"items,omitempty"`
	Installments   []SalesInstallment              `gorm:"foreignKey:SalesTransactionID;constraint:OnDelete:CASCADE;" json:"installments,omitempty"`
	AdjustmentLogs []SalesTransactionAdjustmentLog `gorm:"foreignKey:SalesTransactionID;constraint:OnDelete:CASCADE;" json:"adjustment_logs,omitempty"`
	Returns        []SalesReturn                   `gorm:"foreignKey:TransactionID;constraint:OnDelete:CASCADE;" json:"returns,omitempty"`
	BonusItems     []SalesTransactionBonusItem     `gorm:"foreignKey:SalesTransactionID;constraint:OnDelete:CASCADE;" json:"bonus_items,omitempty"`
	Payments       []SalesPayment                  `gorm:"foreignKey:SalesTransactionID;constraint:OnDelete:CASCADE;" json:"payments,omitempty"`

	IsPaymentLog       bool       `gorm:"-" json:"is_payment_log,omitempty"`
	OriginSalesOrderID *uuid.UUID `gorm:"-" json:"origin_sales_order_id,omitempty"`
	DPAmount           float64    `gorm:"-" json:"down_payment,omitempty"`
	DPPaymentMethod    string     `gorm:"type:varchar(20)" json:"dp_payment_method,omitempty"`
}

func (st *SalesTransaction) BeforeCreate(tx *gorm.DB) (err error) {
	if st.ID == uuid.Nil {
		st.ID = uuid.New()
	}
	return
}

type SalesTransactionBonusItem struct {
	ID                 uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	SalesTransactionID uuid.UUID      `gorm:"type:uuid;not null;index" json:"sales_transaction_id"`
	BonusInventoryID   uuid.UUID      `gorm:"type:uuid;not null;index" json:"bonus_inventory_id"`
	Quantity           int            `gorm:"type:int;not null" json:"quantity"`
	Notes              string         `gorm:"type:text" json:"notes"`
	CreatedAt          time.Time      `gorm:"default:now()" json:"created_at"`
	UpdatedAt          time.Time      `gorm:"default:now()" json:"updated_at"`
	DeletedAt          gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	BonusInventory *BonusInventory `gorm:"foreignKey:BonusInventoryID" json:"bonus_inventory,omitempty"`
}

func (s *SalesTransactionBonusItem) BeforeCreate(tx *gorm.DB) (err error) {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return
}

type ProductRecap struct {
	ProductID   uuid.UUID `json:"product_id"`
	SKU         string    `json:"sku"`
	ProductName string    `json:"product_name"`
	TotalQty    int       `json:"total_qty"`
	TotalAmount float64   `json:"total_amount"`
}
