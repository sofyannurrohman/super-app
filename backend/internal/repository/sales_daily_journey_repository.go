package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"example.com/hris/backend/internal/domain"
)

type SalesDailyJourneyRepository interface {
	GetTodayJourneyByUserID(ctx context.Context, userID uuid.UUID) (*domain.SalesDailyJourney, error)
	GetActiveJourneyByUserID(ctx context.Context, userID uuid.UUID) (*domain.SalesDailyJourney, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.SalesDailyJourney, error)
	Create(ctx context.Context, journey *domain.SalesDailyJourney) error
	Update(ctx context.Context, journey *domain.SalesDailyJourney) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetAllJourneys(ctx context.Context, startDate, endDate string, employeeID, branchID *uuid.UUID) ([]domain.SalesDailyJourney, error)
	GetJourneysByUserID(ctx context.Context, userID uuid.UUID, startDate, endDate string) ([]domain.SalesDailyJourney, error)
}

type salesDailyJourneyRepository struct {
	db *gorm.DB
}

func NewSalesDailyJourneyRepository(db *gorm.DB) SalesDailyJourneyRepository {
	return &salesDailyJourneyRepository{db: db}
}

func (r *salesDailyJourneyRepository) GetTodayJourneyByUserID(ctx context.Context, userID uuid.UUID) (*domain.SalesDailyJourney, error) {
	var journey domain.SalesDailyJourney
	wib := time.FixedZone("WIB", 7*3600)
	todayWIB := time.Now().In(wib).Format("2006-01-02")

	err := r.db.WithContext(ctx).
		Where("user_id = ? AND date = ?", userID, todayWIB).
		Order("created_at DESC").
		First(&journey).Error

	if err != nil {
		return nil, err
	}
	return &journey, nil
}

func (r *salesDailyJourneyRepository) GetActiveJourneyByUserID(ctx context.Context, userID uuid.UUID) (*domain.SalesDailyJourney, error) {
	var journey domain.SalesDailyJourney
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND status = ?", userID, "STARTED").
		Order("created_at DESC").
		First(&journey).Error

	if err != nil {
		return nil, err
	}
	return &journey, nil
}

func (r *salesDailyJourneyRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.SalesDailyJourney, error) {
	var journey domain.SalesDailyJourney
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("User.Employee").
		Where("id = ?", id).
		First(&journey).Error
	if err != nil {
		return nil, err
	}
	return &journey, nil
}

func (r *salesDailyJourneyRepository) Create(ctx context.Context, journey *domain.SalesDailyJourney) error {
	return r.db.WithContext(ctx).Create(journey).Error
}

func (r *salesDailyJourneyRepository) Update(ctx context.Context, journey *domain.SalesDailyJourney) error {
	return r.db.WithContext(ctx).Save(journey).Error
}

func (r *salesDailyJourneyRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&domain.SalesDailyJourney{}).Error
}

func (r *salesDailyJourneyRepository) GetAllJourneys(ctx context.Context, startDate, endDate string, employeeID, branchID *uuid.UUID) ([]domain.SalesDailyJourney, error) {
	var journeys []domain.SalesDailyJourney
	query := r.db.WithContext(ctx).
		Preload("User").
		Preload("User.Employee").
		Preload("User.Employee.Branch").
		Preload("User.Employee.JobPosition")

	if startDate != "" {
		query = query.Where("date >= ?", startDate)
	}
	if endDate != "" {
		query = query.Where("date <= ?", endDate)
	}

	if employeeID != nil {
		query = query.Joins("JOIN employees ON employees.user_id = sales_daily_journeys.user_id").
			Where("employees.id = ?", *employeeID)
	} else if branchID != nil {
		query = query.Joins("JOIN employees ON employees.user_id = sales_daily_journeys.user_id").
			Where("COALESCE(employees.operating_branch_id, employees.branch_id) = ?", *branchID)
	}

	err := query.Order("date DESC, created_at DESC").Find(&journeys).Error
	return journeys, err
}

func (r *salesDailyJourneyRepository) GetJourneysByUserID(ctx context.Context, userID uuid.UUID, startDate, endDate string) ([]domain.SalesDailyJourney, error) {
	var journeys []domain.SalesDailyJourney
	query := r.db.WithContext(ctx).
		Preload("User").
		Preload("User.Employee").
		Where("sales_daily_journeys.user_id = ?", userID)

	if startDate != "" {
		query = query.Where("date >= ?", startDate)
	}
	if endDate != "" {
		query = query.Where("date <= ?", endDate)
	}

	err := query.Order("date DESC, created_at DESC").Find(&journeys).Error
	return journeys, err
}
