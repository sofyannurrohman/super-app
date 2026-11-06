package domain

import (
	"time"

	"gorm.io/gorm"
)

type Order struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	OrderNumber     string         `gorm:"-:migration;type:varchar(50);uniqueIndex;not null" json:"order_number"`
	Source          string         `gorm:"type:varchar(20);not null" json:"source"` // 'SUPER_APP' atau 'TIKTOK'
	ExternalOrderID string         `gorm:"type:varchar(100)" json:"external_order_id"`
	Status          string         `gorm:"type:varchar(30);default:'PENDING'" json:"status"`
	CreatedAt       time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
	OrderItems      []OrderItem    `gorm:"foreignKey:OrderID" json:"order_items"`
}

type OrderItem struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	OrderID     uint           `gorm:"index" json:"order_id"`
	SKU         string         `gorm:"type:varchar(50);not null" json:"sku"`
	Qty         int            `gorm:"not null" json:"qty"`
	PickedQty   int            `gorm:"default:0" json:"picked_qty"`
	PackedQty   int            `gorm:"default:0" json:"packed_qty"`
	BinLocation string         `gorm:"type:varchar(50)" json:"bin_location"`
	CreatedAt   time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

type WebhookLog struct {
	NotificationID string    `gorm:"type:varchar(100);primaryKey" json:"notification_id"`
	Payload        string    `gorm:"type:jsonb" json:"payload"`
	ProcessedAt    time.Time `gorm:"autoCreateTime" json:"processed_at"`
}
