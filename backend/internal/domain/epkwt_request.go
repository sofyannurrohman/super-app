package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type EpkwtRequest struct {
	ID             uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	EmployeeID     uuid.UUID `gorm:"type:uuid;not null" json:"employee_id"`
	DocumentNumber *string   `gorm:"type:varchar(100)" json:"document_number"`

	// Personal Data
	IdentityNumber     string     `gorm:"type:varchar(50);not null" json:"identity_number"`
	FirstName          string     `gorm:"type:varchar(100);not null" json:"first_name"`
	LastName           *string    `gorm:"type:varchar(100)" json:"last_name"`
	PhoneNumber        *string    `gorm:"type:varchar(20)" json:"phone_number"`
	Gender             *string    `gorm:"type:varchar(10)" json:"gender"`
	BirthDate          *time.Time `gorm:"type:date" json:"birth_date"`
	BirthPlace         *string    `gorm:"type:varchar(100)" json:"birth_place"`
	AddressKtp         *string    `gorm:"type:text" json:"address_ktp"`
	AddressResidential *string    `gorm:"type:text" json:"address_residential"`
	MaritalStatus      *string    `gorm:"type:varchar(20)" json:"marital_status"`

	// Job Data
	JobPositionID *uuid.UUID `gorm:"type:uuid" json:"job_position_id"`
	DepartmentID  *uuid.UUID `gorm:"type:uuid" json:"department_id"`
	Salary        float64    `gorm:"type:decimal(15,2);default:0" json:"salary"`
	StartDate     *time.Time `gorm:"type:date" json:"start_date"`
	EndDate       *time.Time `gorm:"type:date" json:"end_date"`
	AgreementDate *time.Time `gorm:"type:date" json:"agreement_date"`

	// Signatures & Status
	EmployeeSignatureURL *string `gorm:"type:text" json:"employee_signature_url"`
	HrSignatureURL       *string `gorm:"type:text" json:"hr_signature_url"`
	Status               string  `gorm:"type:varchar(50);default:'PENDING_HR'" json:"status"`
	ProbationMonths      *int    `gorm:"type:int;default:0" json:"probation_months"`

	// Enhancements: Sequence, Document Upload & Notes
	ContractSequence int     `gorm:"type:int;default:1" json:"contract_sequence"`
	DocumentURL      *string `gorm:"type:text" json:"document_url"`
	Notes            *string `gorm:"type:text" json:"notes"`

	CreatedAt time.Time `gorm:"default:now()" json:"created_at"`
	UpdatedAt time.Time `gorm:"default:now()" json:"updated_at"`

	Employee    *Employee    `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
	JobPosition *JobPosition `gorm:"foreignKey:JobPositionID" json:"job_position,omitempty"`
	Department  *Department  `gorm:"foreignKey:DepartmentID" json:"department,omitempty"`
}

func (r *EpkwtRequest) BeforeCreate(tx *gorm.DB) (err error) {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return
}
