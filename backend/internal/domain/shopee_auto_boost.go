package domain

import (
	"time"

	"github.com/google/uuid"
)

// ShopeeAutoBoostItem merepresentasikan produk Shopee yang didaftarkan ke antrean boost
type ShopeeAutoBoostItem struct {
	ID            uint       `gorm:"primaryKey" json:"id"`
	BranchID      *uuid.UUID `gorm:"type:uuid;index" json:"branch_id"`
	ShopID        string     `gorm:"type:varchar(100);not null" json:"shop_id"`
	ItemID        int64      `gorm:"not null" json:"item_id"`
	ItemName      string     `gorm:"type:varchar(500);not null" json:"item_name"`
	ItemSKU       string     `gorm:"type:varchar(100)" json:"item_sku"`
	ImageURL      string     `gorm:"type:text" json:"image_url"`
	PriceAmount   float64    `gorm:"type:numeric(15,2);default:0" json:"price_amount"`
	Sequence      int        `gorm:"default:0" json:"sequence"`
	IsActive      bool       `gorm:"default:true" json:"is_active"`
	LastBoostedAt *time.Time `json:"last_boosted_at"`
	CooldownUntil *time.Time `json:"cooldown_until"`
	Status        string     `gorm:"type:varchar(50);default:'IDLE'" json:"status"` // 'IDLE', 'BOOSTED', 'FAILED'
	LastError     string     `gorm:"type:text" json:"last_error,omitempty"`
	CurrentStock  int        `gorm:"-" json:"current_stock"`
	ProductStatus string     `gorm:"-" json:"product_status"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`

	Branch *Branch `gorm:"foreignKey:BranchID" json:"branch,omitempty"`
}

func (ShopeeAutoBoostItem) TableName() string {
	return "shopee_auto_boost_items"
}

// ShopeeAutoBoostSetting merepresentasikan pengaturan auto-boost per toko Shopee
type ShopeeAutoBoostSetting struct {
	ID                   uint       `gorm:"primaryKey" json:"id"`
	BranchID             *uuid.UUID `gorm:"type:uuid;index" json:"branch_id"`
	ShopID               string     `gorm:"type:varchar(100);not null;uniqueIndex" json:"shop_id"`
	IsEnabled            bool       `gorm:"default:false" json:"is_enabled"`
	RotationMode         string     `gorm:"type:varchar(50);default:'ROUND_ROBIN'" json:"rotation_mode"`
	CheckIntervalMinutes int        `gorm:"default:5" json:"check_interval_minutes"`
	LastRunAt            *time.Time `json:"last_run_at"`
	LastRunStatus        string     `gorm:"type:varchar(50)" json:"last_run_status"`
	LastRunMessage       string     `gorm:"type:text" json:"last_run_message"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`

	Branch *Branch `gorm:"foreignKey:BranchID" json:"branch,omitempty"`
}

func (ShopeeAutoBoostSetting) TableName() string {
	return "shopee_auto_boost_settings"
}

// ShopeeAutoBoostLog merepresentasikan riwayat tindakan boost
type ShopeeAutoBoostLog struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	ShopID      string    `gorm:"type:varchar(100);not null;index" json:"shop_id"`
	ItemID      int64     `gorm:"not null" json:"item_id"`
	ItemName    string    `gorm:"type:varchar(500)" json:"item_name"`
	TriggeredBy string    `gorm:"type:varchar(50);default:'WORKER'" json:"triggered_by"` // 'WORKER', 'MANUAL'
	Status      string    `gorm:"type:varchar(50);not null" json:"status"`               // 'SUCCESS', 'FAILED'
	Message     string    `gorm:"type:text" json:"message"`
	BoostedAt   time.Time `gorm:"not null" json:"boosted_at"`
	CreatedAt   time.Time `json:"created_at"`
}

func (ShopeeAutoBoostLog) TableName() string {
	return "shopee_auto_boost_logs"
}

// DTOs
type ShopeeBoostSlotInfo struct {
	SlotIndex       int       `json:"slot_index"` // 1 to 5
	ItemID          int64     `json:"item_id"`
	ItemName        string    `json:"item_name"`
	ImageURL        string    `json:"image_url"`
	CooldownSeconds int       `json:"cooldown_seconds"`
	CooldownUntil   time.Time `json:"cooldown_until"`
	IsOccupied      bool      `json:"is_occupied"`
}

type ShopeeBoostOverviewResponse struct {
	ShopID           string                 `json:"shop_id"`
	IsEnabled        bool                   `json:"is_enabled"`
	MaxSlots         int                    `json:"max_slots"`
	OccupiedSlots    int                    `json:"occupied_slots"`
	AvailableSlots   int                    `json:"available_slots"`
	TotalInQueue     int64                  `json:"total_in_queue"`
	ActiveInQueue    int64                  `json:"active_in_queue"`
	Slots            []ShopeeBoostSlotInfo  `json:"slots"`
	Settings         ShopeeAutoBoostSetting `json:"settings"`
	RecentLogs       []ShopeeAutoBoostLog   `json:"recent_logs"`
}

type AddShopeeBoostItemsRequest struct {
	ItemIDs []int64 `json:"item_ids" binding:"required"`
}

type ReorderShopeeBoostItemsRequest struct {
	ItemIDs []uint `json:"item_ids" binding:"required"` // IDs of ShopeeAutoBoostItem in sequence
}
