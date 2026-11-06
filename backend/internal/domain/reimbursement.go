package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ReimbursementStatus string

const (
	ReimbursementPending  ReimbursementStatus = "PENDING"
	ReimbursementApproved ReimbursementStatus = "APPROVED"
	ReimbursementRejected ReimbursementStatus = "REJECTED"
)

type Reimbursement struct {
	ID             uuid.UUID           `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	EmployeeID     uuid.UUID           `gorm:"type:uuid;not null" json:"employee_id"`
	Title          string              `gorm:"type:varchar(255);not null" json:"title"`
	Description    string              `gorm:"type:text" json:"description"`
	Amount         float64             `gorm:"type:decimal(15,2);not null" json:"amount"`
	AttachmentURL  string              `gorm:"type:text" json:"attachment_url"`
	Status         ReimbursementStatus `gorm:"type:varchar(20);default:'PENDING'" json:"status"`
	ApprovedBy     *uuid.UUID          `gorm:"type:uuid" json:"approved_by"`
	RejectedReason *string             `gorm:"type:text" json:"rejected_reason"`

	// Reimbursement Category & Banner Installation details
	Category             string              `gorm:"type:varchar(50);default:'GENERAL'" json:"category"` // GENERAL, BANNER_INSTALLATION, BBM
	BannerCount          int                 `gorm:"default:0" json:"banner_count"`
	UnitRate             float64             `gorm:"type:decimal(15,2);default:0" json:"unit_rate"`

	// BBM & Odometer Details for Cashier Validation
	StartKM              *float64            `gorm:"type:decimal(10,2)" json:"start_km"`
	EndKM                *float64            `gorm:"type:decimal(10,2)" json:"end_km"`
	TotalKM              *float64            `gorm:"type:decimal(10,2)" json:"total_km"`
	StartKMPhotoURL      string              `gorm:"type:text" json:"start_km_photo_url"`
	EndKMPhotoURL        string              `gorm:"type:text" json:"end_km_photo_url"`
	VehiclePlateNumber   string              `gorm:"type:varchar(50)" json:"vehicle_plate_number"`
	FuelLiters           *float64            `gorm:"type:decimal(10,2)" json:"fuel_liters"`

	// Disbursement Info
	DisbursementStatus     string     `gorm:"type:varchar(20);default:'UNPAID'" json:"disbursement_status"` // UNPAID, DISBURSED
	DisbursedVia           string     `gorm:"type:varchar(20)" json:"disbursed_via"`                        // PETTY_CASH, BANK_TRANSFER, CASH
	DisbursedAt            *time.Time `json:"disbursed_at"`
	DisbursedBy            *uuid.UUID `gorm:"type:uuid" json:"disbursed_by"`
	DisbursementProofURL   string     `gorm:"type:text" json:"disbursement_proof_url"`
	DisbursementSourceType string     `gorm:"type:varchar(50)" json:"disbursement_source_type"` // KAS_PABRIK, BRANCH_PETTY_CASH, BRANCH_CASH, BANK_TRANSFER
	KasPabrikID            *uuid.UUID `gorm:"type:uuid" json:"kas_pabrik_id"`
	ExpenseAccountID       *uuid.UUID `gorm:"type:uuid" json:"expense_account_id"`
	ExpenseAccountName     string     `gorm:"type:varchar(255)" json:"expense_account_name"`
	BankAccountID          *uuid.UUID `gorm:"type:uuid" json:"bank_account_id"`

	CreatedAt time.Time `gorm:"default:now()" json:"created_at"`
	UpdatedAt time.Time `gorm:"default:now()" json:"updated_at"`

	Employee       *Employee           `gorm:"foreignKey:EmployeeID;constraint:OnDelete:CASCADE;" json:"employee,omitempty"`
	Approver       *Employee           `gorm:"foreignKey:ApprovedBy;constraint:OnDelete:SET NULL;" json:"approver,omitempty"`
	Disburser      *User               `gorm:"foreignKey:DisbursedBy;constraint:OnDelete:SET NULL;" json:"disburser,omitempty"`
	BannerOrders   []BannerOrder       `gorm:"foreignKey:ReimbursementID" json:"banner_orders,omitempty"`
	KasPabrik      *KasPabrik          `gorm:"foreignKey:KasPabrikID" json:"kas_pabrik,omitempty"`
	ExpenseAccount *Account            `gorm:"foreignKey:ExpenseAccountID" json:"expense_account,omitempty"`
	BankAccount    *CompanyBankAccount `gorm:"foreignKey:BankAccountID" json:"bank_account,omitempty"`
}

func (r *Reimbursement) BeforeCreate(tx *gorm.DB) (err error) {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return
}
