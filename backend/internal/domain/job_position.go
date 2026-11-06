package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type JobPosition struct {
	ID        uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CompanyID *uuid.UUID `gorm:"type:uuid" json:"company_id"`
	Title     string     `gorm:"type:varchar(100);not null" json:"title"`
	Level     int        `gorm:"default:1" json:"level"`
	CreatedAt time.Time  `gorm:"default:now()" json:"created_at"`

	Company *Company `gorm:"foreignKey:CompanyID"`
}

func (j *JobPosition) BeforeCreate(tx *gorm.DB) (err error) {
	if j.ID == uuid.Nil {
		j.ID = uuid.New()
	}
	return
}
