package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SalesPayment menyimpan log transaksi pembayaran parsial atau lunas.
// Model ini mendukung baik alur SalesOrder (baru) maupun SalesTransaction (lama).
type SalesPayment struct {
	ID                    uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	SalesOrderID          *uuid.UUID `gorm:"type:uuid;index" json:"sales_order_id,omitempty"`
	SalesTransactionID    *uuid.UUID `gorm:"type:uuid;index" json:"sales_transaction_id,omitempty"`
	EmployeeID            uuid.UUID  `gorm:"type:uuid" json:"employee_id"` // Siapa yang menginput/menagih
	Amount                float64    `gorm:"type:decimal(15,2);not null" json:"amount"`
	PaymentMethod         string     `gorm:"type:varchar(50)" json:"payment_method"`                   // CASH, MIDTRANS_QRIS, etc.
	PaymentStatus         string     `gorm:"type:varchar(20);default:'SUCCESS'" json:"payment_status"` // PENDING, SUCCESS, FAILED
	MidtransTransactionID *string    `gorm:"type:varchar(100)" json:"midtrans_transaction_id"`
	CompanyBankAccountID  *uuid.UUID `gorm:"type:uuid" json:"company_bank_account_id"`
	TransferProofURL      *string    `gorm:"type:text" json:"transfer_proof_url"`
	IsVerified            bool       `gorm:"type:boolean;default:false" json:"is_verified"`
	VerifiedByID          *uuid.UUID `gorm:"type:uuid" json:"verified_by_id,omitempty"`
	VerifiedBy            *Employee  `gorm:"foreignKey:VerifiedByID" json:"verified_by,omitempty"`
	PaymentDate           time.Time  `gorm:"default:now()" json:"payment_date"`
	Notes                 string     `gorm:"type:text" json:"notes"`
	RejectionReason       *string    `gorm:"type:text" json:"rejection_reason"`
	CollectedBy           string     `gorm:"type:varchar(100)" json:"collected_by"` // Nama supir/admin (string display)
	CreatedAt             time.Time  `gorm:"default:now()" json:"created_at"`
	UpdatedAt             time.Time  `gorm:"default:now()" json:"updated_at"`
	DeletedAt             gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	ApprovedAt            *time.Time `gorm:"type:timestamp" json:"approved_at"`

	PenerimaanStatus       string     `gorm:"type:varchar(50);default:'PENDING'" json:"penerimaan_status"` // 'PENDING', 'VERIFIED'
	PenerimaanVerifiedAt   *time.Time `gorm:"type:timestamp" json:"penerimaan_verified_at,omitempty"`
	PenerimaanVerifiedByID *uuid.UUID `gorm:"type:uuid" json:"penerimaan_verified_by_id,omitempty"`
	PenerimaanVerifiedBy   *Employee  `gorm:"foreignKey:PenerimaanVerifiedByID" json:"penerimaan_verified_by,omitempty"`
	PenerimaanNotes        *string    `gorm:"type:text" json:"penerimaan_notes,omitempty"`

	SalesOrder         *SalesOrder         `gorm:"foreignKey:SalesOrderID" json:"sales_order,omitempty"`
	SalesTransaction   *SalesTransaction   `gorm:"foreignKey:SalesTransactionID" json:"sales_transaction,omitempty"`
	Employee           *Employee           `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
	CompanyBankAccount *CompanyBankAccount `gorm:"foreignKey:CompanyBankAccountID" json:"company_bank_account,omitempty"`
}

func (sp *SalesPayment) BeforeCreate(tx *gorm.DB) (err error) {
	if sp.ID == uuid.Nil {
		sp.ID = uuid.New()
	}
	if sp.PaymentDate.IsZero() {
		sp.PaymentDate = time.Now()
	}
	return
}
