package domain

import (
	"time"

	"github.com/google/uuid"
)

type DailyWorker struct {
	ID        uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	BranchID  *uuid.UUID `gorm:"type:uuid;index" json:"branch_id,omitempty"`
	CompanyID *uuid.UUID `gorm:"type:uuid" json:"company_id,omitempty"`
	Name      string     `gorm:"type:varchar(100);not null" json:"name"`
	Type      string     `gorm:"type:varchar(50);not null" json:"type"` // "Kuli", "Driver"
	CreatedAt time.Time  `gorm:"default:now()" json:"created_at"`

	Branch *Branch `gorm:"foreignKey:BranchID" json:"branch,omitempty"`
}

type DailyWorkerSalary struct {
	ID               uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	DailyWorkerID    uuid.UUID  `gorm:"type:uuid;not null" json:"daily_worker_id"`
	BranchID         *uuid.UUID `gorm:"type:uuid;index" json:"branch_id,omitempty"`
	Date             time.Time  `gorm:"type:date;not null" json:"date"`
	TotalPaid        float64    `gorm:"type:decimal(15,2)" json:"total_paid"`         // Gaji / Tarif Pengiriman
	DeliveryBatchID  *uuid.UUID `gorm:"type:uuid" json:"delivery_batch_id,omitempty"` // Nullable, khusus driver
	Notes            string     `gorm:"type:text" json:"notes"`
	ProofURL         string     `gorm:"type:text" json:"proof_url"`                      // Foto bukti hitungan
	Status           string     `gorm:"type:varchar(20);default:'UNPAID'" json:"status"` // UNPAID, PAID
	KasPabrikID      *uuid.UUID `gorm:"type:uuid;index" json:"kas_pabrik_id,omitempty"`
	KasTransactionID *uuid.UUID `gorm:"type:uuid;index" json:"kas_transaction_id,omitempty"`
	PaidAt           *time.Time `json:"paid_at,omitempty"`
	PaidBy           *uuid.UUID `gorm:"type:uuid" json:"paid_by,omitempty"`
	CreatedAt        time.Time  `gorm:"default:now()" json:"created_at"`

	DailyWorker    *DailyWorker    `gorm:"foreignKey:DailyWorkerID" json:"daily_worker,omitempty"`
	Branch         *Branch         `gorm:"foreignKey:BranchID" json:"branch,omitempty"`
	DeliveryBatch  *DeliveryBatch  `gorm:"foreignKey:DeliveryBatchID" json:"delivery_batch,omitempty"`
	KasPabrik      *KasPabrik      `gorm:"foreignKey:KasPabrikID" json:"kas_pabrik,omitempty"`
	KasTransaction *KasTransaction `gorm:"foreignKey:KasTransactionID" json:"kas_transaction,omitempty"`
	Payer          *User           `gorm:"foreignKey:PaidBy" json:"payer,omitempty"`
}
