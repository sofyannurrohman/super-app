package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AttendanceLog struct {
	ID         uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	EmployeeID uuid.UUID  `gorm:"type:uuid;not null" json:"employee_id"`
	ShiftID    *uuid.UUID `gorm:"type:uuid" json:"shift_id"`

	ClockInTime  *time.Time `json:"clock_in_time"`
	ClockOutTime *time.Time `json:"clock_out_time"`

	ClockInLat            float64 `gorm:"type:decimal(10,8)" json:"clock_in_lat"`
	ClockInLong           float64 `gorm:"type:decimal(11,8)" json:"clock_in_long"`
	ClockInPhotoURL       string  `gorm:"type:text" json:"clock_in_photo_url"`
	ClockInDeviceID       string  `gorm:"type:varchar(100)" json:"clock_in_device_id"`
	ClockInLocationStatus string  `gorm:"type:varchar(20)" json:"clock_in_location_status"`

	ClockOutLat      float64 `gorm:"type:decimal(10,8)" json:"clock_out_lat"`
	ClockOutLong     float64 `gorm:"type:decimal(11,8)" json:"clock_out_long"`
	ClockOutPhotoURL string  `gorm:"type:text" json:"clock_out_photo_url"`

	Status      string `gorm:"type:varchar(20)" json:"status"`
	Notes       string `gorm:"type:text" json:"notes"`
	IsOutOfTown bool   `gorm:"default:false" json:"is_out_of_town"`
	Destination string `gorm:"type:varchar(255)" json:"destination"`

	CreatedAt time.Time `gorm:"default:now()" json:"created_at"`
	UpdatedAt time.Time `gorm:"default:now()" json:"updated_at"`

	Employee          *Employee          `gorm:"foreignKey:EmployeeID;constraint:OnDelete:CASCADE;" json:"employee,omitempty"`
	Shift             *Shift             `gorm:"foreignKey:ShiftID" json:"shift,omitempty"`
	SalesTransactions []SalesTransaction `gorm:"foreignKey:VisitID" json:"sales_transactions,omitempty"`
}

func (a *AttendanceLog) BeforeCreate(tx *gorm.DB) (err error) {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return
}
