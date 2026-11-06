package domain

import (
	"time"

	"github.com/google/uuid"
)

type LPSReport struct {
	ID            uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	EmployeeID    uuid.UUID  `gorm:"type:uuid;not null" json:"employee_id"`
	ReportDate    time.Time  `gorm:"type:date;not null" json:"report_date"`
	LPSType       string     `gorm:"type:varchar(30);default:'DELIVERY'" json:"lps_type"`
	TotalCash     float64    `gorm:"type:decimal(15,2);default:0" json:"total_cash"`
	TotalTransfer float64    `gorm:"type:decimal(15,2);default:0" json:"total_transfer"`
	TotalTempo    float64    `gorm:"type:decimal(15,2);default:0" json:"total_tempo"`
	Status        string     `gorm:"type:varchar(50);default:'DRAFT'" json:"status"` // 'DRAFT', 'VERIFIED'
	VerifiedByID  *uuid.UUID `gorm:"type:uuid" json:"verified_by_id,omitempty"`
	VerifiedAt    *time.Time `json:"verified_at,omitempty"`
	SnapshotData  string     `gorm:"type:jsonb" json:"snapshot_data"`
	CreatedAt     time.Time  `gorm:"default:now()" json:"created_at"`
	UpdatedAt     time.Time  `gorm:"default:now()" json:"updated_at"`

	Employee *Employee `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
}
