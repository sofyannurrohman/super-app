package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SalesVisitPlan struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	EmployeeID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_visit_plan_unique" json:"employee_id"`
	StoreID    uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_visit_plan_unique" json:"store_id"`
	PlanDate   time.Time `gorm:"type:date;not null;uniqueIndex:idx_visit_plan_unique" json:"plan_date"`
	CreatedAt  time.Time `gorm:"default:now()" json:"created_at"`
	UpdatedAt  time.Time `gorm:"default:now()" json:"updated_at"`

	Store       Store                      `gorm:"foreignKey:StoreID" json:"store"`
	Projections []SalesVisitPlanProjection `gorm:"foreignKey:VisitPlanID" json:"projections"`
}

func (s *SalesVisitPlan) BeforeCreate(tx *gorm.DB) (err error) {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return
}
