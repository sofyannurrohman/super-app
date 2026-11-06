package repository

import (
	"time"

	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"gorm.io/gorm"
)

var jakartaLoc *time.Location

func init() {
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		loc = time.UTC
	}
	jakartaLoc = loc
}

type SalesVisitPlanRepository interface {
	Create(plan *domain.SalesVisitPlan) error
	GetByEmployeeAndDate(employeeID uuid.UUID, date time.Time) ([]domain.SalesVisitPlan, error)
	DeleteByEmployeeAndDate(employeeID uuid.UUID, date time.Time) error
	DeleteSpecificPlan(employeeID uuid.UUID, storeID uuid.UUID, date time.Time) error
	UpdateProjections(planID uuid.UUID, projections []domain.SalesVisitPlanProjection) error
}

type salesVisitPlanRepository struct {
	db *gorm.DB
}

func NewSalesVisitPlanRepository(db *gorm.DB) SalesVisitPlanRepository {
	return &salesVisitPlanRepository{db}
}

func (r *salesVisitPlanRepository) Create(plan *domain.SalesVisitPlan) error {
	return r.db.Omit("Store", "Projections.Product").Create(plan).Error
}

func (r *salesVisitPlanRepository) GetByEmployeeAndDate(employeeID uuid.UUID, date time.Time) ([]domain.SalesVisitPlan, error) {
	var plans []domain.SalesVisitPlan
	// Always align to Asia/Jakarta so the date range matches what Postgres stores
	date = date.In(jakartaLoc)
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, jakartaLoc)
	endOfDay := startOfDay.Add(24 * time.Hour)
	err := r.db.Preload("Store").
		Preload("Projections").
		Preload("Projections.Product").
		Where("employee_id = ? AND plan_date >= ? AND plan_date < ?", employeeID, startOfDay, endOfDay).
		Find(&plans).Error
	return plans, err
}

func (r *salesVisitPlanRepository) DeleteByEmployeeAndDate(employeeID uuid.UUID, date time.Time) error {
	date = date.In(jakartaLoc)
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, jakartaLoc)
	endOfDay := startOfDay.Add(24 * time.Hour)
	return r.db.Where("employee_id = ? AND plan_date >= ? AND plan_date < ?", employeeID, startOfDay, endOfDay).Delete(&domain.SalesVisitPlan{}).Error
}

func (r *salesVisitPlanRepository) DeleteSpecificPlan(employeeID uuid.UUID, storeID uuid.UUID, date time.Time) error {
	date = date.In(jakartaLoc)
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, jakartaLoc)
	endOfDay := startOfDay.Add(24 * time.Hour)
	return r.db.Where("employee_id = ? AND store_id = ? AND plan_date >= ? AND plan_date < ?", employeeID, storeID, startOfDay, endOfDay).Delete(&domain.SalesVisitPlan{}).Error
}

func (r *salesVisitPlanRepository) UpdateProjections(planID uuid.UUID, projections []domain.SalesVisitPlanProjection) error {
	// First, clear existing projections for this plan
	err := r.db.Where("visit_plan_id = ?", planID).Delete(&domain.SalesVisitPlanProjection{}).Error
	if err != nil {
		return err
	}

	// Insert new projections if any
	if len(projections) > 0 {
		for i := range projections {
			projections[i].VisitPlanID = planID
		}
		return r.db.Create(&projections).Error
	}
	return nil
}
