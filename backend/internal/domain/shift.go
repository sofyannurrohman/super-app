package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Shift struct {
	ID         uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CompanyID  *uuid.UUID `gorm:"type:uuid" json:"company_id"`
	Name       string     `gorm:"type:varchar(50);not null" json:"name"`
	StartTime  string     `gorm:"type:time;not null" json:"start_time"`
	EndTime    string     `gorm:"type:time;not null" json:"end_time"`
	BreakStart *string    `gorm:"type:time" json:"break_start"`
	BreakEnd   *string    `gorm:"type:time" json:"break_end"`
	BranchID   *uuid.UUID `gorm:"type:uuid" json:"branch_id"`
	IsFlexible bool       `gorm:"default:false" json:"is_flexible"`
	CreatedAt  time.Time  `gorm:"default:now()" json:"created_at"`

	Company *Company `gorm:"foreignKey:CompanyID" json:"company,omitempty"`
	Branch  *Branch  `gorm:"foreignKey:BranchID" json:"branch,omitempty"`
}

func (s *Shift) BeforeCreate(tx *gorm.DB) (err error) {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return
}

type EmployeeShift struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	EmployeeID uuid.UUID `gorm:"type:uuid;not null" json:"employee_id"`
	ShiftID    uuid.UUID `gorm:"type:uuid;not null" json:"shift_id"`
	Date       time.Time `gorm:"type:date;not null;uniqueIndex:idx_emp_shift_date" json:"date"`
	IsOffDay   bool      `gorm:"default:false" json:"is_off_day"`

	Employee *Employee `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
	Shift    *Shift    `gorm:"foreignKey:ShiftID" json:"shift,omitempty"`
}

func (es *EmployeeShift) BeforeCreate(tx *gorm.DB) (err error) {
	if es.ID == uuid.Nil {
		es.ID = uuid.New()
	}
	return
}
