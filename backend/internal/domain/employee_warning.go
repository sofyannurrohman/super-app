package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type EmployeeWarning struct {
	ID               uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	EmployeeID       uuid.UUID  `gorm:"type:uuid;not null" json:"employee_id"`
	WarningType      string     `gorm:"type:varchar(50);not null" json:"warning_type"`
	IssueDescription string     `gorm:"type:text" json:"issue_description"`
	Status           string     `gorm:"type:varchar(50);not null;default:'pending'" json:"status"`
	IsActive         bool       `gorm:"type:boolean;not null;default:true" json:"is_active"`
	HandledByID      *uuid.UUID `gorm:"type:uuid" json:"handled_by_id"`
	CreatedAt        time.Time  `gorm:"default:now()" json:"created_at"`
	UpdatedAt        time.Time  `gorm:"default:now()" json:"updated_at"`

	Employee  *Employee `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
	HandledBy *User     `gorm:"foreignKey:HandledByID" json:"handled_by,omitempty"`
}

func (w *EmployeeWarning) BeforeCreate(tx *gorm.DB) (err error) {
	if w.ID == uuid.Nil {
		w.ID = uuid.New()
	}
	return
}
