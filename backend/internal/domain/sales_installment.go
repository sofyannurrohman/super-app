package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SalesInstallment struct {
	ID                 uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	SalesTransactionID uuid.UUID `gorm:"type:uuid;not null;index" json:"sales_transaction_id"`
	TermNumber         int       `gorm:"type:int;not null" json:"term_number"` // 1, 2, 3
	DueDate            time.Time `gorm:"type:date;not null" json:"due_date"`
	Amount             float64   `gorm:"type:decimal(15,2);not null" json:"amount"`
	PaidAmount         float64   `gorm:"type:decimal(15,2);default:0" json:"paid_amount"`
	Status             string    `gorm:"type:varchar(20);default:'UNPAID'" json:"status"` // UNPAID, PAID
	CreatedAt          time.Time `gorm:"default:now()" json:"created_at"`
	UpdatedAt          time.Time `gorm:"default:now()" json:"updated_at"`
	DeletedAt          gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	SalesTransaction *SalesTransaction `gorm:"foreignKey:SalesTransactionID" json:"-"`
}

func (si *SalesInstallment) BeforeCreate(tx *gorm.DB) (err error) {
	if si.ID == uuid.Nil {
		si.ID = uuid.New()
	}
	return
}
