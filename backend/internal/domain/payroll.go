package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PayrollComponent struct {
	ID        uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CompanyID *uuid.UUID `gorm:"type:uuid" json:"company_id"`
	Name      string     `gorm:"type:varchar(100);not null" json:"name"`
	Type      string     `gorm:"type:varchar(20);not null" json:"type"` // EARNING, DEDUCTION, BENEFIT
	IsTaxable bool       `gorm:"default:true" json:"is_taxable"`
	CreatedAt time.Time  `gorm:"default:now()" json:"created_at"`

	Company *Company `gorm:"foreignKey:CompanyID" json:"company,omitempty"`
}

type MyPayslipResponse struct {
	ID             string              `json:"id"`
	Period         string              `json:"period"`       // "01 Mar - 31 Mar 2026"
	PaymentDate    string              `json:"payment_date"` // YYYY-MM-DD
	JobTitle       string              `json:"job_title"`
	BasicSalary    float64             `json:"basic_salary"`
	TotalAllowance float64             `json:"total_allowance"`
	TotalDeduction float64             `json:"total_deduction"`
	TakeHomePay    float64             `json:"take_home_pay"`
	Earnings       []PayslipItemDetail `json:"earnings"`
	Deductions     []PayslipItemDetail `json:"deductions"`
}

type PayslipItemDetail struct {
	Name   string  `json:"name"`
	Amount float64 `json:"amount"`
}

func (pc *PayrollComponent) BeforeCreate(tx *gorm.DB) (err error) {
	if pc.ID == uuid.Nil {
		pc.ID = uuid.New()
	}
	return
}

type EmployeeSalarySetting struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	EmployeeID  uuid.UUID `gorm:"type:uuid;not null" json:"employee_id"`
	ComponentID uuid.UUID `gorm:"type:uuid;not null" json:"component_id"`
	Amount      float64   `gorm:"type:decimal(15,2);not null" json:"amount"`
	CreatedAt   time.Time `gorm:"default:now()" json:"created_at"`

	Employee  *Employee         `gorm:"foreignKey:EmployeeID;constraint:OnDelete:CASCADE;" json:"employee,omitempty"`
	Component *PayrollComponent `gorm:"foreignKey:ComponentID" json:"component,omitempty"`
}

func (ess *EmployeeSalarySetting) BeforeCreate(tx *gorm.DB) (err error) {
	if ess.ID == uuid.Nil {
		ess.ID = uuid.New()
	}
	return
}

type PayrollRun struct {
	ID              uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	BranchID        *uuid.UUID `gorm:"type:uuid" json:"branch_id"`
	CompanyID       *uuid.UUID `gorm:"type:uuid" json:"company_id"`
	PeriodStart     time.Time  `gorm:"type:date;not null" json:"period_start"`
	PeriodEnd       time.Time  `gorm:"type:date;not null" json:"period_end"`
	PaymentSchedule *time.Time `gorm:"type:date" json:"payment_schedule"`
	Status          string     `gorm:"type:varchar(20);default:'DRAFT'" json:"status"`
	TotalPayout     *float64   `gorm:"type:decimal(20,2)" json:"total_payout"`
	CreatedAt       time.Time  `gorm:"default:now()" json:"created_at"`

	Branch  *Branch  `gorm:"foreignKey:BranchID" json:"branch,omitempty"`
	Company *Company `gorm:"foreignKey:CompanyID" json:"company,omitempty"`
}

func (pr *PayrollRun) BeforeCreate(tx *gorm.DB) (err error) {
	if pr.ID == uuid.Nil {
		pr.ID = uuid.New()
	}
	return
}

type Payslip struct {
	ID                 uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	PayrollRunID       uuid.UUID `gorm:"type:uuid;not null" json:"payroll_run_id"`
	EmployeeID         uuid.UUID `gorm:"type:uuid;not null" json:"employee_id"`
	SnapshotJobTitle   string    `gorm:"type:varchar(100)" json:"snapshot_job_title"`
	SnapshotPtkpStatus string    `gorm:"type:varchar(10)" json:"snapshot_ptkp_status"`
	BasicSalary        float64   `gorm:"type:decimal(15,2)" json:"basic_salary"`
	TotalAllowance     float64   `gorm:"type:decimal(15,2)" json:"total_allowance"`
	TotalDeduction     float64   `gorm:"type:decimal(15,2)" json:"total_deduction"`
	TakeHomePay        float64   `gorm:"type:decimal(15,2)" json:"take_home_pay"`
	Pph21Amount        float64   `gorm:"type:decimal(15,2)" json:"pph21_amount"`
	CreatedAt          time.Time `gorm:"default:now()" json:"created_at"`

	PayrollRun *PayrollRun   `gorm:"foreignKey:PayrollRunID" json:"payroll_run,omitempty"`
	Employee   *Employee     `gorm:"foreignKey:EmployeeID;constraint:OnDelete:CASCADE;" json:"employee,omitempty"`
	Items      []PayslipItem `gorm:"foreignKey:PayslipID" json:"items,omitempty"`
}

func (p *Payslip) BeforeCreate(tx *gorm.DB) (err error) {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return
}

type PayslipItem struct {
	ID            uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	PayslipID     uuid.UUID `gorm:"type:uuid;not null" json:"payslip_id"`
	ComponentName string    `gorm:"type:varchar(100);not null" json:"component_name"`
	Amount        float64   `gorm:"type:decimal(15,2);not null" json:"amount"`
	Type          string    `gorm:"type:varchar(20);not null" json:"type"` // EARNING, DEDUCTION
	CreatedAt     time.Time `gorm:"default:now()" json:"created_at"`

	Payslip *Payslip `gorm:"foreignKey:PayslipID" json:"payslip,omitempty"`
}

func (pi *PayslipItem) BeforeCreate(tx *gorm.DB) (err error) {
	if pi.ID == uuid.Nil {
		pi.ID = uuid.New()
	}
	return
}

// BankTransferGroup — Ringkasan karyawan per bank untuk 1 payroll run
type BankTransferGroup struct {
	BankName     string  `json:"bank_name"`
	TotalRecords int     `json:"total_records"`
	TotalAmount  float64 `json:"total_amount"`
	HasMissing   bool    `json:"has_missing"` // true jika ada karyawan tanpa rekening
}

// PayrollTransferSummary — Response untuk modal summary sebelum download
type PayrollTransferSummary struct {
	PayrollRunID string              `json:"payroll_run_id"`
	Period       string              `json:"period"`
	TotalRecords int                 `json:"total_records"`
	TotalAmount  float64             `json:"total_amount"`
	MissingBank  int                 `json:"missing_bank"` // karyawan tanpa data rekening
	Banks        []BankTransferGroup `json:"banks"`
}
