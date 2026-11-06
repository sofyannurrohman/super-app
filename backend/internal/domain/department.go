package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Department struct {
	ID        uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CompanyID *uuid.UUID `gorm:"type:uuid" json:"company_id"`
	Name      string     `gorm:"type:varchar(100);not null" json:"name"`
	ParentID  *uuid.UUID `gorm:"type:uuid" json:"parent_id"`
	CreatedAt time.Time  `gorm:"default:now()" json:"created_at"`

	Company          *Company    `gorm:"foreignKey:CompanyID" json:"company,omitempty"`
	ParentDepartment *Department `gorm:"foreignKey:ParentID" json:"parent_department,omitempty"`
}

func (d *Department) BeforeCreate(tx *gorm.DB) (err error) {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	return
}
