package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LeaveType struct {
	ID            uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CompanyID     *uuid.UUID `gorm:"type:uuid" json:"company_id"`
	Name          string     `gorm:"type:varchar(50);not null" json:"name"`
	IsPaid        bool       `gorm:"default:true" json:"is_paid"`
	RequiresQuota bool       `gorm:"default:true" json:"requires_quota"`
	DefaultQuota  int        `gorm:"default:12" json:"default_quota"`
	CreatedAt     time.Time  `gorm:"default:now()" json:"created_at"`

	Company *Company `gorm:"foreignKey:CompanyID" json:"company,omitempty"`
}

func (lt *LeaveType) BeforeCreate(tx *gorm.DB) (err error) {
	if lt.ID == uuid.Nil {
		lt.ID = uuid.New()
	}
	return
}

type LeaveBalance struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	EmployeeID   uuid.UUID `gorm:"type:uuid;not null" json:"employee_id"`
	LeaveTypeID  uuid.UUID `gorm:"type:uuid;not null" json:"leave_type_id"`
	Year         int       `gorm:"not null" json:"year"`
	BalanceTotal int       `gorm:"default:0" json:"balance_total"`
	BalanceUsed  int       `gorm:"default:0" json:"balance_used"`

	Employee  *Employee  `gorm:"foreignKey:EmployeeID;constraint:OnDelete:CASCADE;" json:"employee,omitempty"`
	LeaveType *LeaveType `gorm:"foreignKey:LeaveTypeID" json:"leave_type,omitempty"`
}

func (lb *LeaveBalance) BeforeCreate(tx *gorm.DB) (err error) {
	if lb.ID == uuid.Nil {
		lb.ID = uuid.New()
	}
	return
}

type LeaveRequest struct {
	ID            uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	EmployeeID    uuid.UUID  `gorm:"type:uuid;not null" json:"employee_id"`
	LeaveTypeID   uuid.UUID  `gorm:"type:uuid;not null" json:"leave_type_id"`
	StartDate     time.Time  `gorm:"type:date;not null" json:"start_date"`
	EndDate       time.Time  `gorm:"type:date;not null" json:"end_date"`
	Reason        string     `gorm:"type:text" json:"reason"`
	AttachmentURL string     `gorm:"type:text" json:"attachment_url"`
	Status        string     `gorm:"type:varchar(20);default:'PENDING'" json:"status"`
	ApprovedBy    *uuid.UUID `gorm:"type:uuid" json:"approved_by"`
	RejectReason  *string    `gorm:"type:text" json:"reject_reason"`
	CreatedAt     time.Time  `gorm:"default:now()" json:"created_at"`

	Employee  *Employee  `gorm:"foreignKey:EmployeeID;constraint:OnDelete:CASCADE;" json:"employee,omitempty"`
	LeaveType *LeaveType `gorm:"foreignKey:LeaveTypeID" json:"leave_type,omitempty"`
	Approver  *Employee  `gorm:"foreignKey:ApprovedBy;constraint:OnDelete:SET NULL;" json:"approver,omitempty"`
}

func (lr *LeaveRequest) BeforeCreate(tx *gorm.DB) (err error) {
	if lr.ID == uuid.Nil {
		lr.ID = uuid.New()
	}
	return
}
