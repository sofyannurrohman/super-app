package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type EmployeeRequestCandidate struct {
	ID                uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	EmployeeRequestID uuid.UUID  `gorm:"type:uuid;not null;index" json:"employee_request_id"`
	CandidateName     string     `gorm:"type:varchar(255);not null" json:"candidate_name"`
	CandidatePhone    string     `gorm:"type:varchar(50)" json:"candidate_phone"`
	CandidateEmail    string     `gorm:"type:varchar(255)" json:"candidate_email"`
	JoinDate          time.Time  `gorm:"type:date;not null" json:"join_date"`
	DocumentURL       string     `gorm:"type:text" json:"document_url"`
	DocumentName      string     `gorm:"type:varchar(255)" json:"document_name"`
	DocumentType      string     `gorm:"type:varchar(50)" json:"document_type"` // PDF, IMAGE, OTHER
	Notes             string     `gorm:"type:text" json:"notes"`
	CreatedAt         time.Time  `gorm:"default:now()" json:"created_at"`
	UpdatedAt         time.Time  `gorm:"default:now()" json:"updated_at"`

	EmployeeRequest   *EmployeeRequest `gorm:"foreignKey:EmployeeRequestID;constraint:OnDelete:CASCADE;" json:"employee_request,omitempty"`
}

func (c *EmployeeRequestCandidate) BeforeCreate(tx *gorm.DB) (err error) {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return
}
