package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// DailyKpiReport represents the header of daily sales performance report
type DailyKpiReport struct {
	ID                 uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ReportDate         time.Time  `gorm:"type:date;not null;index" json:"report_date"`
	ReporterEmployeeID uuid.UUID  `gorm:"type:uuid;not null;index" json:"reporter_employee_id"`
	ReporterRole       string     `gorm:"type:varchar(50);not null;default:'SPV'" json:"reporter_role"` // SPV, ASM, NSM, GM
	BranchID           *uuid.UUID `gorm:"type:uuid;index" json:"branch_id,omitempty"`
	AreaName           string     `gorm:"type:varchar(150)" json:"area_name"`
	ParentReportID     *uuid.UUID `gorm:"type:uuid;index" json:"parent_report_id,omitempty"`

	Status             string     `gorm:"type:varchar(30);not null;default:'DRAFT'" json:"status"` // DRAFT, SUBMITTED, REVISED, APPROVED
	SubmittedAt        *time.Time `json:"submitted_at,omitempty"`
	ReviewedAt         *time.Time `json:"reviewed_at,omitempty"`
	ReviewerEmployeeID *uuid.UUID `gorm:"type:uuid;index" json:"reviewer_employee_id,omitempty"`
	ReviewerFeedback   string     `gorm:"type:text" json:"reviewer_feedback"`

	// Summary Tim
	SummaryTargetCall              int     `gorm:"default:0" json:"summary_target_call"`
	SummaryRealizationCall         int     `gorm:"default:0" json:"summary_realization_call"`
	SummaryAchievementCallPercent  float64 `gorm:"type:decimal(5,2);default:0" json:"summary_achievement_call_percent"`

	SummaryTargetEC                int     `gorm:"default:0" json:"summary_target_ec"`
	SummaryRealizationEC           int     `gorm:"default:0" json:"summary_realization_ec"`
	SummaryAchievementECPercent   float64 `gorm:"type:decimal(5,2);default:0" json:"summary_achievement_ec_percent"`

	SummaryTargetSales             float64 `gorm:"type:decimal(15,2);default:0" json:"summary_target_sales"`
	SummaryRealizationSales        float64 `gorm:"type:decimal(15,2);default:0" json:"summary_realization_sales"`
	SummaryAchievementSalesPercent float64 `gorm:"type:decimal(5,2);default:0" json:"summary_achievement_sales_percent"`

	// Bagian C: Catatan & Validasi
	TomorrowPriority string `gorm:"type:text" json:"tomorrow_priority"`
	Notes            string `gorm:"type:text" json:"notes"`

	CreatedAt time.Time `gorm:"default:now()" json:"created_at"`
	UpdatedAt time.Time `gorm:"default:now()" json:"updated_at"`

	// Relations
	ReporterEmployee *Employee                `gorm:"foreignKey:ReporterEmployeeID" json:"reporter_employee,omitempty"`
	ReviewerEmployee *Employee                `gorm:"foreignKey:ReviewerEmployeeID" json:"reviewer_employee,omitempty"`
	Branch           *Branch                  `gorm:"foreignKey:BranchID" json:"branch,omitempty"`
	Items            []DailyKpiReportItem     `gorm:"foreignKey:DailyKpiReportID;constraint:OnDelete:CASCADE" json:"items,omitempty"`
	ActionPlans      []DailyKpiReportActionPlan `gorm:"foreignKey:DailyKpiReportID;constraint:OnDelete:CASCADE" json:"action_plans,omitempty"`
}

func (r *DailyKpiReport) BeforeCreate(tx *gorm.DB) (err error) {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return
}

// DailyKpiReportItem represents Section A: Target and Realization per Salesman
type DailyKpiReportItem struct {
	ID               uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	DailyKpiReportID uuid.UUID `gorm:"type:uuid;not null;index" json:"daily_kpi_report_id"`
	EmployeeID       uuid.UUID `gorm:"type:uuid;not null;index" json:"employee_id"`
	EmployeeName     string    `gorm:"type:varchar(150);not null" json:"employee_name"`
	AreaRoute        string    `gorm:"type:varchar(150)" json:"area_route"`

	TargetCall             int     `gorm:"default:0" json:"target_call"`
	RealizationCall        int     `gorm:"default:0" json:"realization_call"`
	AchievementCallPercent float64 `gorm:"type:decimal(5,2);default:0" json:"achievement_call_percent"`

	TargetEC             int     `gorm:"default:0" json:"target_ec"`
	RealizationEC        int     `gorm:"default:0" json:"realization_ec"`
	AchievementECPercent float64 `gorm:"type:decimal(5,2);default:0" json:"achievement_ec_percent"`

	TargetSales             float64 `gorm:"type:decimal(15,2);default:0" json:"target_sales"`
	RealizationSales        float64 `gorm:"type:decimal(15,2);default:0" json:"realization_sales"`
	AchievementSalesPercent float64 `gorm:"type:decimal(5,2);default:0" json:"achievement_sales_percent"`
	DiffSales               float64 `gorm:"type:decimal(15,2);default:0" json:"diff_sales"`

	CreatedAt time.Time `gorm:"default:now()" json:"created_at"`
	UpdatedAt time.Time `gorm:"default:now()" json:"updated_at"`

	Employee *Employee `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
}

func (i *DailyKpiReportItem) BeforeCreate(tx *gorm.DB) (err error) {
	if i.ID == uuid.Nil {
		i.ID = uuid.New()
	}
	return
}

// DailyKpiReportActionPlan represents Section B: Evaluation and Action Plan
type DailyKpiReportActionPlan struct {
	ID                uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	DailyKpiReportID  uuid.UUID  `gorm:"type:uuid;not null;index" json:"daily_kpi_report_id"`
	EmployeeID        *uuid.UUID `gorm:"type:uuid;index" json:"employee_id,omitempty"`
	EmployeeName      string     `gorm:"type:varchar(150);not null" json:"employee_name"`
	ReasonUnachieved  string     `gorm:"type:text" json:"reason_unachieved"`
	FieldObstacle     string     `gorm:"type:text" json:"field_obstacle"`
	ActionFollowUp    string     `gorm:"type:text" json:"action_follow_up"`
	PicName           string     `gorm:"type:varchar(100)" json:"pic_name"`
	DueDate           *time.Time `gorm:"type:date" json:"due_date,omitempty"`
	Status            string     `gorm:"type:varchar(30);default:'OPEN'" json:"status"` // OPEN, IN_PROGRESS, DONE, CANCELLED
	SupportNeeded     string     `gorm:"type:text" json:"support_needed"`

	CreatedAt time.Time `gorm:"default:now()" json:"created_at"`
	UpdatedAt time.Time `gorm:"default:now()" json:"updated_at"`

	Employee *Employee `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
}

func (a *DailyKpiReportActionPlan) BeforeCreate(tx *gorm.DB) (err error) {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return
}
