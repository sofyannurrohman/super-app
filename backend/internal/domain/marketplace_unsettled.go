package domain

import (
	"time"

	"github.com/google/uuid"
)

// MarketplaceUnsettledSnapshot is an independent finance snapshot. It never
// participates in operational order status or settlement updates.
type MarketplaceUnsettledSnapshot struct {
	ID                 uint       `gorm:"primaryKey" json:"id"`
	Marketplace        string     `gorm:"type:varchar(50);not null;index" json:"marketplace"`
	ShopID             string     `gorm:"type:varchar(100);not null;index" json:"shop_id"`
	BranchID           *uuid.UUID `gorm:"type:uuid;index" json:"branch_id,omitempty"`
	CapturedAt         time.Time  `gorm:"not null;index" json:"captured_at"`
	Currency           string     `gorm:"type:varchar(10);not null;default:'IDR'" json:"currency"`
	AwaitingDelivery   int64      `gorm:"not null;default:0" json:"awaiting_delivery"`
	AfterSales         int64      `gorm:"not null;default:0" json:"after_sales"`
	Processing         int64      `gorm:"not null;default:0" json:"processing"`
	Source             string     `gorm:"type:varchar(30);not null" json:"source"`
	TransactionCount   int        `gorm:"not null;default:0" json:"transaction_count"`
	UnclassifiedAmount int64      `gorm:"not null;default:0" json:"unclassified_amount"`
	CreatedAt          time.Time  `json:"created_at"`
}

type MarketplaceUnsettledTransaction struct {
	ID                  uint       `gorm:"primaryKey" json:"id"`
	SnapshotID          uint       `gorm:"not null;index" json:"snapshot_id"`
	TransactionID       string     `gorm:"type:varchar(100);not null" json:"transaction_id"`
	OrderID             string     `gorm:"type:varchar(100);index" json:"order_id"`
	AdjustmentID        string     `gorm:"type:varchar(100)" json:"adjustment_id"`
	TransactionType     string     `gorm:"type:varchar(100)" json:"transaction_type"`
	UnsettledReason     string     `gorm:"type:text" json:"unsettled_reason"`
	Category            string     `gorm:"type:varchar(40);not null" json:"category"`
	EstimatedAmount     int64      `gorm:"not null" json:"estimated_amount"`
	EstimatedSettlement string     `gorm:"type:varchar(100)" json:"estimated_settlement"`
	OrderDeliveryTime   *time.Time `json:"order_delivery_time,omitempty"`
}
