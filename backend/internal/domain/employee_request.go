package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type EmployeeRequest struct {
	ID                  uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	RequestNumber       string     `gorm:"type:varchar(100);not null;unique" json:"request_number"`
	RequesterID         *uuid.UUID `gorm:"type:uuid" json:"requester_id"`
	RequesterName       string     `gorm:"type:varchar(255);not null" json:"requester_name"`
	RequesterJobTitle   string     `gorm:"type:varchar(255)" json:"requester_job_title"`
	DepartmentID        *uuid.UUID `gorm:"type:uuid" json:"department_id"`
	BranchID            *uuid.UUID `gorm:"type:uuid" json:"branch_id"`
	CustomBranchName    string     `gorm:"type:varchar(255)" json:"custom_branch_name"`
	JobPositionID       *uuid.UUID `gorm:"type:uuid" json:"job_position_id"`
	CustomPositionTitle string     `gorm:"type:varchar(255)" json:"custom_position_title"`
	EmploymentType      string     `gorm:"type:varchar(50);not null;default:'FULL_TIME'" json:"employment_type"` // FULL_TIME, CONTRACT, DAILY, INTERNSHIP
	RequestReason       string     `gorm:"type:varchar(100);not null;default:'NEW_HEADCOUNT'" json:"request_reason"` // NEW_HEADCOUNT, REPLACEMENT, PROJECT_BASED, WORKLOAD_INCREASE, OTHER
	ReplacementForName  string     `gorm:"type:varchar(255)" json:"replacement_for_name"`
	HeadcountNeeded     int        `gorm:"not null;default:1" json:"headcount_needed"`
	HeadcountFulfilled  int        `gorm:"not null;default:0" json:"headcount_fulfilled"`
	MinEducation        string     `gorm:"type:varchar(100)" json:"min_education"`
	GenderPreference    string     `gorm:"type:varchar(20);default:'ANY'" json:"gender_preference"` // ANY, MALE, FEMALE
	MinAge              int        `gorm:"default:18" json:"min_age"`
	MaxAge              int        `gorm:"default:45" json:"max_age"`
	MinExperienceYears  int        `gorm:"default:0" json:"min_experience_years"`
	RequiredSkills      string     `gorm:"type:text" json:"required_skills"`
	JobDescription      string     `gorm:"type:text" json:"job_description"`
	Notes               string     `gorm:"type:text" json:"notes"`
	TargetDate          time.Time  `gorm:"type:date;not null" json:"target_date"`
	Status              string     `gorm:"type:varchar(50);not null;default:'PENDING'" json:"status"` // PENDING, WAITING_DIRECTOR_APPROVAL, WAITING_HR_APPROVAL, APPROVED, REJECTED, IN_PROGRESS, FULFILLED, CANCELLED

	// Legacy / Direct Approver (backward-compatible)
	ApproverID          *uuid.UUID `gorm:"type:uuid" json:"approver_id"`
	ApprovalDate        *time.Time `json:"approval_date"`
	ApprovalNotes       string     `gorm:"type:text" json:"approval_notes"`

	// Persetujuan HR Manager
	HRApproverID        *uuid.UUID `gorm:"type:uuid" json:"hr_approver_id"`
	HRApprovalDate      *time.Time `json:"hr_approval_date"`
	HRApprovalNotes     string     `gorm:"type:text" json:"hr_approval_notes"`
	HRStatus            string     `gorm:"type:varchar(50);default:'PENDING'" json:"hr_status"` // PENDING, APPROVED, REJECTED

	// Persetujuan Direktur
	DirectorApproverID    *uuid.UUID `gorm:"type:uuid" json:"director_approver_id"`
	DirectorApprovalDate  *time.Time `json:"director_approval_date"`
	DirectorApprovalNotes string     `gorm:"type:text" json:"director_approval_notes"`
	DirectorStatus        string     `gorm:"type:varchar(50);default:'PENDING'" json:"director_status"` // PENDING, APPROVED, REJECTED

	RejectedByRole      string     `gorm:"type:varchar(50)" json:"rejected_by_role"` // HR_MANAGER, DIRECTOR
	RejectReason        string     `gorm:"type:text" json:"reject_reason"`
	RecruitmentNotes    string     `gorm:"type:text" json:"recruitment_notes"`
	CreatedAt           time.Time  `gorm:"default:now()" json:"created_at"`
	UpdatedAt           time.Time  `gorm:"default:now()" json:"updated_at"`

	Requester        *Employee    `gorm:"foreignKey:RequesterID;constraint:OnDelete:SET NULL;" json:"requester,omitempty"`
	Approver         *Employee    `gorm:"foreignKey:ApproverID;constraint:OnDelete:SET NULL;" json:"approver,omitempty"`
	HRApprover       *Employee    `gorm:"foreignKey:HRApproverID;constraint:OnDelete:SET NULL;" json:"hr_approver,omitempty"`
	DirectorApprover *Employee    `gorm:"foreignKey:DirectorApproverID;constraint:OnDelete:SET NULL;" json:"director_approver,omitempty"`
	Department       *Department  `gorm:"foreignKey:DepartmentID;constraint:OnDelete:SET NULL;" json:"department,omitempty"`
	Branch           *Branch      `gorm:"foreignKey:BranchID;constraint:OnDelete:SET NULL;" json:"branch,omitempty"`
	JobPosition      *JobPosition `gorm:"foreignKey:JobPositionID;constraint:OnDelete:SET NULL;" json:"job_position,omitempty"`
	Candidates       []EmployeeRequestCandidate `gorm:"foreignKey:EmployeeRequestID;constraint:OnDelete:CASCADE;" json:"candidates,omitempty"`
}

func (e *EmployeeRequest) BeforeCreate(tx *gorm.DB) (err error) {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	return
}
