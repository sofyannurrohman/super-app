package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OvertimeStatus string

const (
	OvertimePending  OvertimeStatus = "pending"
	OvertimeApproved OvertimeStatus = "approved"
	OvertimeRejected OvertimeStatus = "rejected"
)

type Overtime struct {
	ID              uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	EmployeeID      uuid.UUID      `gorm:"type:uuid;not null;index" json:"employee_id"`
	Date            time.Time      `gorm:"type:date;not null" json:"date"`
	StartTime       time.Time      `gorm:"not null" json:"start_time"`
	EndTime         time.Time      `gorm:"not null" json:"end_time"`
	DurationMinutes int            `gorm:"not null" json:"duration_minutes"`
	Type            string         `gorm:"type:varchar(50);default:'working_day'" json:"type"`
	Reason          string         `gorm:"type:text;not null" json:"reason"`
	Status          OvertimeStatus `gorm:"type:varchar(20);default:'pending'" json:"status"`
	AttachmentUrl   *string        `gorm:"type:text" json:"attachment_url,omitempty"`
	ApprovedBy      *uuid.UUID     `gorm:"type:uuid" json:"approved_by,omitempty"`
	RejectReason    *string        `gorm:"type:text" json:"reject_reason,omitempty"`
	CreatedAt       time.Time      `gorm:"default:now()" json:"created_at"`

	Employee *Employee `gorm:"foreignKey:EmployeeID;constraint:OnDelete:CASCADE;" json:"employee,omitempty"`
	Approver *Employee `gorm:"foreignKey:ApprovedBy;constraint:OnDelete:SET NULL;" json:"approver,omitempty"`
}

func (o *Overtime) BeforeCreate(tx *gorm.DB) (err error) {
	if o.ID == uuid.Nil {
		o.ID = uuid.New()
	}
	return
}
