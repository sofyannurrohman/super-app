package domain

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// FloatArray is a custom type to allow GORM to save float arrays into PostgreSQL JSONB/Text columns
type FloatArray []float32

func (a FloatArray) Value() (driver.Value, error) {
	return json.Marshal(a)
}

func (a *FloatArray) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, &a)
}

type Employee struct {
	ID            uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID        uuid.UUID  `gorm:"type:uuid;not null" json:"user_id"`
	CompanyID         *uuid.UUID `gorm:"type:uuid" json:"company_id"`
	BranchID          *uuid.UUID `gorm:"type:uuid" json:"branch_id"`
	OperatingBranchID *uuid.UUID `gorm:"type:uuid" json:"operating_branch_id"`
	DepartmentID      *uuid.UUID `gorm:"type:uuid" json:"department_id"`
	JobPositionID     *uuid.UUID `gorm:"type:uuid" json:"job_position_id"`

	// Personal Data
	EmployeeIDNumber   string     `gorm:"type:varchar(50);not null" json:"employee_id_number"`
	FirstName          string     `gorm:"type:varchar(100);not null" json:"first_name"`
	LastName           string     `gorm:"type:varchar(100)" json:"last_name"`
	Pin                string     `gorm:"type:varchar(10)" json:"pin"`
	PhoneNumber        string     `gorm:"type:varchar(20)" json:"phone_number"`
	Gender             string     `gorm:"type:varchar(10)" json:"gender"`
	BirthDate          *time.Time `gorm:"type:date" json:"birth_date"`
	BirthPlace         string     `gorm:"type:varchar(100)" json:"birth_place"`
	MaritalStatus      string     `gorm:"type:varchar(20)" json:"marital_status"`
	AddressKTP         string     `gorm:"type:text" json:"address_ktp"`
	AddressResidential string     `gorm:"type:text" json:"address_residential"`
	EmergencyContact   string     `gorm:"type:text" json:"emergency_contact"`
	FaceReferenceURL   string     `gorm:"type:text" json:"face_reference_url"`
	FaceEmbedding      FloatArray `gorm:"type:jsonb" json:"face_embedding"`
	IsFaceRegistered   bool       `json:"is_face_registered"`

	// Government IDs
	IdentityNumber            string `gorm:"type:varchar(50)" json:"identity_number"`
	NpwpNumber                string `gorm:"type:varchar(50)" json:"npwp_number"`
	BpjsKetenagakerjaanNumber string `gorm:"type:varchar(50)" json:"bpjs_ketenagakerjaan_number"`
	BpjsKesehatanNumber       string `gorm:"type:varchar(50)" json:"bpjs_kesehatan_number"`
	PtkpStatus                string `gorm:"type:varchar(10)" json:"ptkp_status"`

	// Bank Information
	BankName          string `gorm:"type:varchar(50)" json:"bank_name"`
	BankAccountNumber string `gorm:"type:varchar(50)" json:"bank_account_number"`
	AccountHolderName string `gorm:"type:varchar(100)" json:"account_holder_name"`

	// Employment Status
	JoinDate         time.Time  `gorm:"type:date;not null" json:"join_date"`
	EndDate          *time.Time `gorm:"type:date" json:"end_date"`
	EmploymentStatus string     `gorm:"type:varchar(50)" json:"employment_status"`
	ManagerID        *uuid.UUID `gorm:"type:uuid" json:"manager_id"`
	DefaultShiftID   *uuid.UUID `gorm:"type:uuid" json:"default_shift_id"`

	TargetDailyInstallations int       `gorm:"column:target_daily_installations;default:0" json:"target_daily_installations"`
	Salary                   *float64  `gorm:"type:numeric(15,2)" json:"salary"`
	IsDistributor            bool       `gorm:"default:false;index" json:"is_distributor"`
	DistributorStoreID       *uuid.UUID `gorm:"type:uuid;index" json:"distributor_store_id,omitempty"`
	CreatedAt                time.Time `gorm:"default:now()" json:"created_at"`
	UpdatedAt                time.Time `gorm:"default:now()" json:"updated_at"`

	User            *User            `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE;" json:"user,omitempty"`
	Company         *Company         `gorm:"foreignKey:CompanyID" json:"company,omitempty"`
	Branch          *Branch          `gorm:"foreignKey:BranchID" json:"branch,omitempty"`
	OperatingBranch *Branch          `gorm:"foreignKey:OperatingBranchID" json:"operating_branch,omitempty"`
	Department      *Department      `gorm:"foreignKey:DepartmentID" json:"department,omitempty"`
	JobPosition     *JobPosition     `gorm:"foreignKey:JobPositionID" json:"job_position,omitempty"`
	DistributorStore *Store          `gorm:"foreignKey:DistributorStoreID" json:"distributor_store,omitempty"`
	Manager        *Employee        `gorm:"foreignKey:ManagerID;constraint:OnDelete:SET NULL;" json:"manager,omitempty"`
	EmployeeShifts []EmployeeShift  `gorm:"foreignKey:EmployeeID" json:"employee_shifts,omitempty"`
	DefaultShift   *Shift           `gorm:"foreignKey:DefaultShiftID" json:"default_shift,omitempty"`
	ActiveWarning  *EmployeeWarning `gorm:"-" json:"active_warning,omitempty"`
}

func (e *Employee) BeforeCreate(tx *gorm.DB) (err error) {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}

	if e.Pin == "" {
		var maxPin string
		tx.Model(&Employee{}).Select("MAX(NULLIF(regexp_replace(pin, '\\D', '', 'g'), ''))").Scan(&maxPin)

		if maxPin == "" {
			e.Pin = "001"
		} else {
			pinInt, _ := strconv.Atoi(maxPin)
			e.Pin = fmt.Sprintf("%03d", pinInt+1)
		}
	}
	return
}
