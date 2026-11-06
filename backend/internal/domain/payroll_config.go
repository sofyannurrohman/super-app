package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PayrollConfig represents the global variables used in payroll calculations
// like tax rates thresholds and insurance percentages.
type PayrollConfig struct {
	ID        uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CompanyID *uuid.UUID `gorm:"type:uuid" json:"company_id"`

	// JHT (Jaminan Hari Tua)
	JHTCompanyPercentage  float64 `gorm:"type:decimal(5,2);default:3.70" json:"jht_company_percentage"`
	JHTEmployeePercentage float64 `gorm:"type:decimal(5,2);default:2.00" json:"jht_employee_percentage"`

	// JP (Jaminan Pensiun)
	JPCompanyPercentage  float64 `gorm:"type:decimal(5,2);default:2.00" json:"jp_company_percentage"`
	JPEmployeePercentage float64 `gorm:"type:decimal(5,2);default:1.00" json:"jp_employee_percentage"`
	JPMaxWageBase        float64 `gorm:"type:decimal(15,2);default:10042300" json:"jp_max_wage_base"`

	// JKK (Kecelakaan Kerja) & JKM (Kematian) usually only company pays
	JKKCompanyPercentage float64 `gorm:"type:decimal(5,2);default:0.24" json:"jkk_company_percentage"`
	JKMCompanyPercentage float64 `gorm:"type:decimal(5,2);default:0.30" json:"jkm_company_percentage"`

	// BPJS Kesehatan
	BPJSKesCompanyPercentage  float64 `gorm:"type:decimal(5,2);default:4.00" json:"bpjs_kes_company_percentage"`
	BPJSKesEmployeePercentage float64 `gorm:"type:decimal(5,2);default:1.00" json:"bpjs_kes_employee_percentage"`
	BPJSKesMaxWageBase        float64 `gorm:"type:decimal(15,2);default:12000000" json:"bpjs_kes_max_wage_base"`

	// Tax (PPh21 TER)
	PtkpBaseTK0 float64 `gorm:"type:decimal(15,2);default:54000000" json:"ptkp_base_tk0"`

	// Deductions
	AbsentDeduction float64 `gorm:"type:decimal(15,2);default:100000" json:"absent_deduction"`
	KPIBonusAmount  float64 `gorm:"type:decimal(15,2);default:500000" json:"kpi_bonus_amount"`

	UpdatedAt time.Time `gorm:"default:now()" json:"updated_at"`
	CreatedAt time.Time `gorm:"default:now()" json:"created_at"`

	Company *Company `gorm:"foreignKey:CompanyID" json:"company,omitempty"`
}

func (pc *PayrollConfig) BeforeCreate(tx *gorm.DB) (err error) {
	if pc.ID == uuid.Nil {
		pc.ID = uuid.New()
	}
	return
}
