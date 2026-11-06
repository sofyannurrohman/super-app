package domain

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SalesVisitPlanProjection struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	VisitPlanID uuid.UUID `gorm:"type:uuid;not null;index" json:"visit_plan_id"`
	ProductID   uuid.UUID `gorm:"type:uuid;not null" json:"product_id"`
	Quantity    int       `gorm:"not null" json:"quantity"`

	Product Product `gorm:"foreignKey:ProductID" json:"product"`
}

func (s *SalesVisitPlanProjection) BeforeCreate(tx *gorm.DB) (err error) {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return
}
