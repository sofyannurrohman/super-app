package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// FixedAsset — Entitas aset tetap perusahaan (PSAK 16)
type FixedAsset struct {
	ID                      uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	CompanyID               uuid.UUID      `gorm:"type:uuid;not null;index" json:"company_id"`
	AssetCode               string         `gorm:"size:50;not null;uniqueIndex" json:"asset_code"` // FA-001
	Name                    string         `gorm:"size:255;not null" json:"name"`
	Category                string         `gorm:"size:100;not null" json:"category"` // MACHINERY, VEHICLE, BUILDING, IT_EQUIPMENT, FURNITURE, OTHER
	Description             string         `gorm:"type:text" json:"description"`
	Location                string         `gorm:"size:255" json:"location"`
	AcquisitionDate         time.Time      `gorm:"not null" json:"acquisition_date"`
	AcquisitionCost         float64        `gorm:"not null" json:"acquisition_cost"`
	UsefulLifeYears         int            `gorm:"not null" json:"useful_life_years"`
	ResidualValue           float64        `gorm:"default:0" json:"residual_value"`
	DepreciationMethod      string         `gorm:"size:30;default:'STRAIGHT_LINE'" json:"depreciation_method"` // STRAIGHT_LINE, DECLINING_BALANCE
	Status                  string         `gorm:"size:20;default:'ACTIVE'" json:"status"`                     // ACTIVE, DISPOSED, SOLD
	AccumulatedDepreciation float64        `gorm:"default:0" json:"accumulated_depreciation"`
	LastDepreciatedMonth    *time.Time     `json:"last_depreciated_month"`
	Notes                   string         `gorm:"type:text" json:"notes"`
	CreatedAt               time.Time      `json:"created_at"`
	UpdatedAt               time.Time      `json:"updated_at"`
	DeletedAt               gorm.DeletedAt `gorm:"index" json:"-"`

	// Computed (not stored)
	BookValue           float64 `gorm:"-" json:"book_value"`
	MonthlyDepreciation float64 `gorm:"-" json:"monthly_depreciation"`
}

func (fa *FixedAsset) BeforeCreate(tx *gorm.DB) error {
	if fa.ID == uuid.Nil {
		fa.ID = uuid.New()
	}
	return nil
}

// ComputeDerived mengisi field-field computed dari data yang ada
func (fa *FixedAsset) ComputeDerived() {
	fa.BookValue = fa.AcquisitionCost - fa.AccumulatedDepreciation
	if fa.BookValue < fa.ResidualValue {
		fa.BookValue = fa.ResidualValue
	}
	depreciableAmount := fa.AcquisitionCost - fa.ResidualValue
	if fa.UsefulLifeYears > 0 {
		fa.MonthlyDepreciation = depreciableAmount / float64(fa.UsefulLifeYears*12)
	}
}

// DepreciationScheduleLine — Baris jadwal depresiasi per bulan
type DepreciationScheduleLine struct {
	Period             string  `json:"period"` // "Jan 2025"
	OpeningBookValue   float64 `json:"opening_book_value"`
	DepreciationAmount float64 `json:"depreciation_amount"`
	ClosingBookValue   float64 `json:"closing_book_value"`
	AccumulatedToDate  float64 `json:"accumulated_to_date"`
}

// FixedAssetSummary — Ringkasan nilai aset untuk KPI cards
type FixedAssetSummary struct {
	TotalAssets          int     `json:"total_assets"`
	TotalAcquisitionCost float64 `json:"total_acquisition_cost"`
	TotalAccumulated     float64 `json:"total_accumulated_depreciation"`
	TotalBookValue       float64 `json:"total_book_value"`
}
