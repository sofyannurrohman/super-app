package domain

import (
	"time"

	"github.com/google/uuid"
)

type SalesmanLiveLocation struct {
	EmployeeID uuid.UUID  `gorm:"type:uuid;primaryKey" json:"employee_id"`
	CompanyID  uuid.UUID  `gorm:"type:uuid;index;not null" json:"company_id"`
	BranchID   *uuid.UUID `gorm:"type:uuid;index" json:"branch_id"`
	Latitude   float64    `gorm:"type:decimal(10,7);not null" json:"latitude"`
	Longitude  float64    `gorm:"type:decimal(10,7);not null" json:"longitude"`
	Accuracy   float64    `gorm:"type:decimal(8,2)" json:"accuracy"`
	IsOnDuty   bool       `gorm:"type:boolean;default:false" json:"is_on_duty"`
	RecordedAt time.Time  `gorm:"not null;index" json:"recorded_at"`
	UpdatedAt  time.Time  `gorm:"default:now()" json:"updated_at"`

	Employee *Employee `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
}
