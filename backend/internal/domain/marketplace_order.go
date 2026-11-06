package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MarketplaceOrder struct {
	ID                    uint       `gorm:"primaryKey" json:"id"`
	BranchID              *uuid.UUID `gorm:"type:uuid;index" json:"branch_id"`
	OrderID               string     `gorm:"type:varchar(100);index" json:"order_id"`
	OrderStatus           string     `gorm:"type:varchar(50)" json:"order_status"`
	OrderSubstatus        string     `gorm:"type:varchar(50)" json:"order_substatus"`
	CancelationReturnType string     `gorm:"type:varchar(50)" json:"cancelation_return_type"`
	SkuID                 string     `gorm:"type:varchar(100)" json:"sku_id"`
	LineItemID            string     `gorm:"type:varchar(100)" json:"line_item_id"`
	SellerSKU             string     `gorm:"type:varchar(100)" json:"seller_sku"`
	ProductName           string     `gorm:"type:varchar(500)" json:"product_name"`
	WarehouseName         string     `gorm:"type:varchar(200)" json:"warehouse_name"`
	TrackingID            string     `gorm:"type:varchar(100)" json:"tracking_id"`
	Marketplace           string     `gorm:"type:varchar(50);default:'tiktok'" json:"marketplace"`
	BuyerUsername         string     `gorm:"type:varchar(200)" json:"buyer_username"`
	PaymentMethod         string     `gorm:"type:varchar(100)" json:"payment_method"`
	TotalAmount           float64    `gorm:"type:decimal(15,2);default:0" json:"total_amount"`
	Currency              string     `gorm:"type:varchar(10);default:'IDR'" json:"currency"`
	ItemCount             int        `gorm:"default:1" json:"item_count"`
	ShippingProvider      string     `gorm:"type:varchar(100)" json:"shipping_provider"`
	CreateTimeRaw         int64      `gorm:"default:0" json:"create_time_raw"`
	PickedBy              string     `gorm:"type:varchar(200)" json:"picked_by"`
	PackedBy              string     `gorm:"type:varchar(200)" json:"packed_by"`

	// Finansial & Settlement Tracking (TikTok / Shopee)
	SettlementStatus        string     `gorm:"type:varchar(50);default:'PENDING_SETTLEMENT'" json:"settlement_status"`
	SettlementDate          *time.Time `json:"settlement_date"`
	EstimatedSettlementDate *time.Time `json:"estimated_settlement_date"`
	SettlementAmount        float64    `gorm:"type:decimal(15,2);default:0" json:"settlement_amount"`
	PlatformFee             float64    `gorm:"type:decimal(15,2);default:0" json:"platform_fee"`
	CommissionFee           float64    `gorm:"type:decimal(15,2);default:0" json:"commission_fee"`
	ShippingFeeDeduction    float64    `gorm:"type:decimal(15,2);default:0" json:"shipping_fee_deduction"`
	FeeAndTaxAmount         float64    `gorm:"type:decimal(15,2);default:0" json:"fee_and_tax_amount"`
	ShippingCostAmount      float64    `gorm:"type:decimal(15,2);default:0" json:"shipping_cost_amount"`
	AdjustmentAmount        float64    `gorm:"type:decimal(15,2);default:0" json:"adjustment_amount"`
	FeeTaxBreakdown         string     `gorm:"type:jsonb" json:"fee_tax_breakdown,omitempty"`
	HoldReason              string     `gorm:"type:varchar(255)" json:"hold_reason,omitempty"`
	FinanceSyncedAt         *time.Time `gorm:"type:timestamptz" json:"finance_synced_at,omitempty"`

	// Refund & Retur Fisik Barang
	RefundAmount    float64    `gorm:"type:decimal(15,2);default:0" json:"refund_amount"`
	RefundReason    string     `gorm:"type:varchar(255)" json:"refund_reason"`
	ReturnStatus    string     `gorm:"type:varchar(50);default:'NO_RETURN'" json:"return_status"`
	ReturnItemCount int        `gorm:"default:0" json:"return_item_count"`
	RestockedAt     *time.Time `json:"restocked_at"`
	RestockedBy     string     `gorm:"type:varchar(200)" json:"restocked_by"`
	OpnameID        *uuid.UUID `gorm:"type:uuid" json:"opname_id"`

	SyncedAt  *time.Time `gorm:"type:timestamptz;index" json:"synced_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	Branch    *Branch    `gorm:"foreignKey:BranchID" json:"branch,omitempty"`
}

// New orders do not have finance details yet. An empty Go string is not valid JSONB.
func (m *MarketplaceOrder) BeforeCreate(_ *gorm.DB) error {
	if m.FeeTaxBreakdown == "" {
		m.FeeTaxBreakdown = "{}"
	}
	return nil
}

type MarketplaceScanHistory struct {
	ID            uint       `gorm:"primaryKey" json:"id"`
	BranchID      *uuid.UUID `gorm:"type:uuid;index" json:"branch_id"`
	OrderID       string     `gorm:"type:varchar(100);index" json:"order_id"`
	Action        string     `gorm:"type:varchar(20)" json:"action"` // "PICK" or "PACK"
	Actor         string     `gorm:"type:varchar(200)" json:"actor"`
	SellerSKU     string     `gorm:"type:varchar(100)" json:"seller_sku"`
	ProductName   string     `gorm:"type:text" json:"product_name"`
	WarehouseName string     `gorm:"type:varchar(200)" json:"warehouse_name"`
	CreatedAt     time.Time  `gorm:"autoCreateTime;index" json:"created_at"`
	Branch        *Branch    `gorm:"foreignKey:BranchID" json:"branch,omitempty"`
}

type OrderFiltersResponse struct {
	Warehouses        []string `json:"warehouses"`
	Products          []string `json:"products"`
	ShippingProviders []string `json:"shipping_providers"`
}

type MarketplaceOrderGroup struct {
	OrderID                 string             `json:"order_id"`
	BranchID                *uuid.UUID         `json:"branch_id,omitempty"`
	OrderStatus             string             `json:"order_status"`
	WarehouseName           string             `json:"warehouse_name"`
	TrackingID              string             `json:"tracking_id"`
	Marketplace             string             `json:"marketplace"`
	BuyerUsername           string             `json:"buyer_username"`
	PaymentMethod           string             `json:"payment_method"`
	TotalAmount             float64            `json:"total_amount"`
	Currency                string             `json:"currency"`
	ShippingProvider        string             `json:"shipping_provider"`
	PickedBy                string             `json:"picked_by"`
	PackedBy                string             `json:"packed_by"`
	IsPacked                bool               `json:"is_packed"`
	PackingStatus           string             `json:"packing_status"`        // "PACKED" / "UNPACKED"
	ReconciliationStatus    string             `json:"reconciliation_status"` // "MATCH_SETTLED", "IN_TRANSIT_HOLD", "RETURN_DISPATCHED", "RETURN_UNPACKED", "RESTOCKED", "ANOMALY_SETTLED_UNPACKED"
	SettlementStatus        string             `json:"settlement_status"`
	SettlementDate          *time.Time         `json:"settlement_date"`
	EstimatedSettlementDate *time.Time         `json:"estimated_settlement_date"`
	SettlementAmount        float64            `json:"settlement_amount"`
	PlatformFee             float64            `json:"platform_fee"`
	CommissionFee           float64            `json:"commission_fee"`
	ShippingFeeDeduction    float64            `json:"shipping_fee_deduction"`
	FeeAndTaxAmount         float64            `json:"fee_and_tax_amount"`
	ShippingCostAmount      float64            `json:"shipping_cost_amount"`
	AdjustmentAmount        float64            `json:"adjustment_amount"`
	FeeTaxBreakdown         string             `json:"fee_tax_breakdown,omitempty"`
	HoldReason              string             `json:"hold_reason,omitempty"`
	FinanceSyncedAt         *time.Time         `json:"finance_synced_at,omitempty"`
	RefundAmount            float64            `json:"refund_amount"`
	RefundReason            string             `json:"refund_reason"`
	ReturnStatus            string             `json:"return_status"`
	ReturnItemCount         int                `json:"return_item_count"`
	RestockedAt             *time.Time         `json:"restocked_at"`
	RestockedBy             string             `json:"restocked_by"`
	SyncedAt                *time.Time         `json:"synced_at,omitempty"`
	CreatedAt               time.Time          `json:"created_at"`
	Items                   []MarketplaceOrder `json:"items"`
}

type AdditionalItemReq struct {
	StockID string `json:"stock_id"`
	Qty     int    `json:"qty"`
}

type PackOrderReq struct {
	BoxStockID      *uuid.UUID          `json:"box_stock_id"`
	BoxQty          int                 `json:"box_qty"`
	AdditionalItems []AdditionalItemReq `json:"additional_items"`
	SkipDefaultBox  bool                `json:"skip_default_box"`
}

type SettlementSummaryResponse struct {
	TotalSettledAmount          float64 `json:"total_settled_amount"`
	TotalSettledOrders          int64   `json:"total_settled_orders"`
	TotalPendingAmount          float64 `json:"total_pending_amount"`
	TotalPendingOrders          int64   `json:"total_pending_orders"`
	TotalRefundAmount           float64 `json:"total_refund_amount"`
	TotalRefundOrders           int64   `json:"total_refund_orders"`
	TotalPlatformFees           float64 `json:"total_platform_fees"`
	TotalPendingReturnItems     int64   `json:"total_pending_return_items"`
	TotalRestockedItems         int64   `json:"total_restocked_items"`
	TotalDispatchedReturnOrders int64   `json:"total_dispatched_return_orders"`
	TotalDispatchedReturnValue  float64 `json:"total_dispatched_return_value"`
	TotalRestockedReturnOrders  int64   `json:"total_restocked_return_orders"`
	TotalRestockedReturnValue   float64 `json:"total_restocked_return_value"`
}

type RestockOrderReq struct {
	OrderID string `json:"order_id" binding:"required"`
	Notes   string `json:"notes"`
}

type BatchSettlementImportReq struct {
	Marketplace string                   `json:"marketplace"`
	Records     []SettlementImportRecord `json:"records" binding:"required"`
}

type SettlementImportRecord struct {
	OrderID          string    `json:"order_id" binding:"required"`
	SettlementDate   time.Time `json:"settlement_date"`
	SettlementAmount float64   `json:"settlement_amount"`
	PlatformFee      float64   `json:"platform_fee"`
	CommissionFee    float64   `json:"commission_fee"`
	GrossAmount      float64   `json:"gross_amount"`
	Status           string    `json:"status"` // SETTLED / REFUNDED
}

type UpdateSettlementOrderReq struct {
	SettlementStatus   *string    `json:"settlement_status"`
	SettlementDate     *time.Time `json:"settlement_date"`
	SettlementAmount   *float64   `json:"settlement_amount"`
	PlatformFee        *float64   `json:"platform_fee"`
	CommissionFee      *float64   `json:"commission_fee"`
	FeeAndTaxAmount    *float64   `json:"fee_and_tax_amount"`
	ShippingCostAmount *float64   `json:"shipping_cost_amount"`
	AdjustmentAmount   *float64   `json:"adjustment_amount"`
	FeeTaxBreakdown    *string    `json:"fee_tax_breakdown"`
	RefundAmount       *float64   `json:"refund_amount"`
	RefundReason       *string    `json:"refund_reason"`
	ReturnStatus       *string    `json:"return_status"`
}

type FinanceTransactionRecord struct {
	OrderID            string     `json:"order_id"`
	Marketplace        string     `json:"marketplace,omitempty"`
	BranchID           *uuid.UUID `json:"branch_id,omitempty"`
	SettlementAmount   float64    `json:"settlement_amount"`
	SettlementProvided bool       `json:"-"`
	FeeAndTaxAmount    float64    `json:"fee_and_tax_amount"`
	FeeAndTaxProvided  bool       `json:"-"`
	ShippingCostAmount float64    `json:"shipping_cost_amount"`
	AdjustmentAmount   float64    `json:"adjustment_amount"`
	FeeTaxBreakdown    string     `json:"fee_tax_breakdown"`
	SettlementStatus   string     `json:"settlement_status"`
	SettlementDate     *time.Time `json:"settlement_date"`
	StatementID        string     `json:"statement_id"`
	HoldReason         string     `json:"hold_reason"`
}
