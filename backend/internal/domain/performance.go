package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SalesKPI tracks Omzet/Sales targets for sales employees
type SalesKPI struct {
	ID                     uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	EmployeeID             uuid.UUID `gorm:"type:uuid;not null"`
	TargetOmzet            float64   `gorm:"type:decimal(15,2);not null"`
	AchievedOmzet          float64   `gorm:"type:decimal(15,2);default:0"`
	AchievedOmzetLama      float64   `gorm:"type:decimal(15,2);default:0"`
	AchievedOmzetBaru      float64   `gorm:"type:decimal(15,2);default:0"`
	TargetNewStores        int       `gorm:"type:int;default:0"`
	AchievedNewStores      int       `gorm:"type:int;default:0"`
	TotalVisits            int       `gorm:"type:int;default:0"`
	TargetDailyVisits      int       `gorm:"type:int;default:0" json:"target_daily_visits"`
	TargetDailyNewStores   int       `gorm:"type:int;default:0" json:"target_daily_new_stores"`
	Commission             float64   `gorm:"type:decimal(15,2);default:0" json:"commission"`
	EstimatedBonus         float64   `gorm:"type:decimal(15,2);default:0"`
	PeriodMonth            int       `gorm:"type:int;not null" json:"period_month"`
	PeriodYear             int       `gorm:"type:int;not null" json:"period_year"`
	WorkingTerritory       string    `gorm:"type:varchar(200)" json:"working_territory"`
	ReceivableLimitPercent float64   `gorm:"type:decimal(5,2);default:0" json:"receivable_limit_percent"`
	MaxReceivableLimit     float64   `gorm:"type:decimal(15,2);default:0" json:"max_receivable_limit"`
	CreatedAt              time.Time `gorm:"default:now()"`
	UpdatedAt              time.Time `gorm:"default:now()"`

	Employee *Employee `gorm:"foreignKey:EmployeeID;constraint:OnDelete:CASCADE;"`
}

func (s *SalesKPI) BeforeCreate(tx *gorm.DB) (err error) {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return
}

// EmployeeKPI tracks Attendance, Productivity, and general metrics for regular employees
type EmployeeKPI struct {
	ID                uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	EmployeeID        uuid.UUID `gorm:"type:uuid;not null" json:"employee_id"`
	AttendanceScore   float64   `gorm:"type:decimal(5,2);default:0" json:"attendance_score"`   // 0-100 scale
	ProductivityScore float64   `gorm:"type:decimal(5,2);default:0" json:"productivity_score"` // System generated 0-100 scale
	FinalScore        float64   `gorm:"type:decimal(5,2);default:0" json:"final_score"`        // Aggregate score

	OnTimeCount int `gorm:"default:0" json:"on_time_count"`
	LateCount   int `gorm:"default:0" json:"late_count"`
	AlphaCount  int `gorm:"default:0" json:"alpha_count"`
	PermitCount int `gorm:"default:0" json:"permit_count"`

	Status      string    `gorm:"type:varchar(20);default:'DRAFT'" json:"status"` // DRAFT, FINALIZED
	PeriodMonth int       `gorm:"type:int;not null" json:"period_month"`
	PeriodYear  int       `gorm:"type:int;not null" json:"period_year"`
	CreatedAt   time.Time `gorm:"default:now()" json:"created_at"`
	UpdatedAt   time.Time `gorm:"default:now()" json:"updated_at"`

	Employee *Employee `gorm:"foreignKey:EmployeeID;constraint:OnDelete:CASCADE;" json:"employee,omitempty"`
}

func (e *EmployeeKPI) BeforeCreate(tx *gorm.DB) (err error) {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	return
}

// OperatorTarget tracks daily installation targets for banner operators
type OperatorTarget struct {
	ID                       uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	EmployeeID               uuid.UUID `gorm:"type:uuid;not null;unique" json:"employee_id"`
	TargetDailyInstallations int       `gorm:"type:int;default:0" json:"target_daily_installations"`
	CreatedAt                time.Time `gorm:"default:now()" json:"created_at"`
	UpdatedAt                time.Time `gorm:"default:now()" json:"updated_at"`

	Employee *Employee `gorm:"foreignKey:EmployeeID;constraint:OnDelete:CASCADE;" json:"employee,omitempty"`
}

func (s *OperatorTarget) BeforeCreate(tx *gorm.DB) (err error) {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return
}

// ManagerialAppraisal tracks qualitative feedback and rating from a manager
type ManagerialAppraisal struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	EmployeeID  uuid.UUID `gorm:"type:uuid;not null"`
	ManagerID   uuid.UUID `gorm:"type:uuid;not null"`
	ReviewNotes string    `gorm:"type:text"`
	Rating      float64   `gorm:"type:decimal(3,2);not null"` // 1-5 scale
	ReviewDate  time.Time `gorm:"type:date;not null"`
	CreatedAt   time.Time `gorm:"default:now()"`

	Employee *Employee `gorm:"foreignKey:EmployeeID;constraint:OnDelete:CASCADE;"`
	Manager  *Employee `gorm:"foreignKey:ManagerID;constraint:OnDelete:CASCADE;"`
}

func (m *ManagerialAppraisal) BeforeCreate(tx *gorm.DB) (err error) {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return
}
