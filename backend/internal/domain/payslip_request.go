package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PayslipRequestStatus string

const (
	PayslipRequestStatusPending  PayslipRequestStatus = "pending"
	PayslipRequestStatusApproved PayslipRequestStatus = "approved"
	PayslipRequestStatusRejected PayslipRequestStatus = "rejected"
)

type PayslipRequest struct {
	ID           uuid.UUID            `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID       uuid.UUID            `gorm:"type:uuid;not null" json:"user_id"`
	PeriodMonth  int                  `gorm:"type:int;not null" json:"period_month"`
	PeriodYear   int                  `gorm:"type:int;not null" json:"period_year"`
	Reason       *string              `gorm:"type:text" json:"reason"`
	Status       PayslipRequestStatus `gorm:"type:varchar(20);default:'pending'" json:"status"`
	Nominal      *float64             `gorm:"type:decimal(15,2)" json:"nominal"`
	FileURL      *string              `gorm:"type:varchar(255)" json:"file_url"`
	RejectReason *string              `gorm:"type:text" json:"reject_reason"`
	PayslipData  *string              `gorm:"type:jsonb" json:"payslip_data"`
	ApprovedBy   *uuid.UUID           `gorm:"type:uuid" json:"approved_by"`
	ApprovedAt   *time.Time           `json:"approved_at"`
	CreatedAt    time.Time            `gorm:"default:now()" json:"created_at"`
	UpdatedAt    time.Time            `gorm:"default:now()" json:"updated_at"`

	User     *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Approver *User     `gorm:"foreignKey:ApprovedBy" json:"approver,omitempty"`
}

func (pr *PayslipRequest) BeforeCreate(tx *gorm.DB) (err error) {
	if pr.ID == uuid.Nil {
		pr.ID = uuid.New()
	}
	return
}

// Structs for Auto-Lookup Result
type PayslipLookupEmployee struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	EmployeeIDNumber  string `json:"employee_id_number"`
	JobPosition       string `json:"job_position"`
	Department        string `json:"department"`
	PtkpStatus        string `json:"ptkp_status"`
	BankName          string `json:"bank_name"`
	BankAccountNumber string `json:"bank_account_number"`
}

type PayslipLookupAttendance struct {
	PresentDays     int     `json:"present_days"`
	LateDays        int     `json:"late_days"`
	AbsentDays      int     `json:"absent_days"`
	AbsentDeduction float64 `json:"absent_deduction"`
}

type PayslipLookupLeave struct {
	PaidLeaveDays   int     `json:"paid_leave_days"`
	UnpaidLeaveDays int     `json:"unpaid_leave_days"`
	UnpaidDeduction float64 `json:"unpaid_deduction"`
}

type PayslipLookupBPJS struct {
	JHTEmployee float64 `json:"jht_employee"`
	JPEmployee  float64 `json:"jp_employee"`
	KesEmployee float64 `json:"kes_employee"`
	TotalBPJS   float64 `json:"total_bpjs"`
}

type PayslipItemSummary struct {
	Name   string  `json:"name"`
	Amount float64 `json:"amount"`
	Type   string  `json:"type"` // EARNING, DEDUCTION
}

type PayslipLookupResult struct {
	RequestID       string                   `json:"request_id"`
	PeriodMonth     int                      `json:"period_month"`
	PeriodYear      int                      `json:"period_year"`
	PeriodLabel     string                   `json:"period_label"`
	Employee        PayslipLookupEmployee    `json:"employee"`
	BasicSalary     float64                  `json:"basic_salary"`
	Allowances      []PayslipItemSummary     `json:"allowances"`
	TotalAllowance  float64                  `json:"total_allowance"`
	TotalEarning    float64                  `json:"total_earning"`
	Attendance      PayslipLookupAttendance  `json:"attendance"`
	Leaves          PayslipLookupLeave       `json:"leaves"`
	BPJS            PayslipLookupBPJS        `json:"bpjs"`
	Deductions      []PayslipItemSummary     `json:"deductions"`
	TotalDeduction  float64                  `json:"total_deduction"`
	TakeHomePay     float64                  `json:"take_home_pay"`
}
