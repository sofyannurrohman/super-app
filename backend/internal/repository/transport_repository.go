package repository

import (
	"time"

	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"gorm.io/gorm"
)

type TransportRepository interface {
	// Trips
	CreateTrip(trip *domain.TransportTrip) error
	GetTrips(companyID *uuid.UUID, branchID *uuid.UUID, driverID *uuid.UUID, vehicleID *uuid.UUID, status string, regionType string, startDate, endDate *time.Time) ([]domain.TransportTrip, error)
	GetTripByID(id uuid.UUID) (*domain.TransportTrip, error)
	UpdateTrip(trip *domain.TransportTrip) error
	DeleteTrip(id uuid.UUID) error

	// Backhauls
	CreateBackhaul(backhaul *domain.TransportBackhaul) error
	GetBackhauls(tripID *uuid.UUID, driverID *uuid.UUID, vehicleID *uuid.UUID, status string, startDate, endDate *time.Time) ([]domain.TransportBackhaul, error)
	GetBackhaulByID(id uuid.UUID) (*domain.TransportBackhaul, error)
	UpdateBackhaul(backhaul *domain.TransportBackhaul) error
	DeleteBackhaul(id uuid.UUID) error

	// Expense Claims
	CreateExpenseClaim(claim *domain.TransportExpenseClaim) error
	GetExpenseClaims(tripID *uuid.UUID, vehicleID *uuid.UUID, driverID *uuid.UUID, category string, status string, startDate, endDate *time.Time) ([]domain.TransportExpenseClaim, error)
	GetExpenseClaimByID(id uuid.UUID) (*domain.TransportExpenseClaim, error)
	UpdateExpenseClaim(claim *domain.TransportExpenseClaim) error
	DeleteExpenseClaim(id uuid.UUID) error

	// Fleet Inspections
	CreateFleetInspection(inspection *domain.TransportFleetInspection) error
	GetFleetInspections(vehicleID *uuid.UUID, tripID *uuid.UUID, followUpAction string, followUpStatus string, startDate, endDate *time.Time) ([]domain.TransportFleetInspection, error)
	GetFleetInspectionByID(id uuid.UUID) (*domain.TransportFleetInspection, error)
	UpdateFleetInspection(inspection *domain.TransportFleetInspection) error
	DeleteFleetInspection(id uuid.UUID) error

	// Driver Attendances
	CreateDriverAttendance(att *domain.TransportDriverAttendance) error
	GetDriverAttendances(tripID *uuid.UUID, driverID *uuid.UUID, startDate, endDate *time.Time) ([]domain.TransportDriverAttendance, error)

	// Dashboard Stats
	GetDashboardStats(companyID *uuid.UUID, branchID *uuid.UUID) (map[string]interface{}, error)
}

type transportRepository struct {
	db *gorm.DB
}

func NewTransportRepository(db *gorm.DB) TransportRepository {
	return &transportRepository{db: db}
}

// ──────────────── TRIPS ────────────────
func (r *transportRepository) CreateTrip(trip *domain.TransportTrip) error {
	return r.db.Create(trip).Error
}

func (r *transportRepository) GetTrips(companyID *uuid.UUID, branchID *uuid.UUID, driverID *uuid.UUID, vehicleID *uuid.UUID, status string, regionType string, startDate, endDate *time.Time) ([]domain.TransportTrip, error) {
	var trips []domain.TransportTrip
	query := r.db.Model(&domain.TransportTrip{}).
		Preload("Vehicle").
		Preload("Driver").
		Preload("Helper").
		Preload("Branch").
		Preload("Backhauls").
		Preload("ExpenseClaims").
		Preload("Inspection").
		Order("departure_date DESC, created_at DESC")

	if companyID != nil && *companyID != uuid.Nil {
		query = query.Where("company_id = ?", *companyID)
	}
	if branchID != nil && *branchID != uuid.Nil {
		query = query.Where("branch_id = ?", *branchID)
	}
	if driverID != nil && *driverID != uuid.Nil {
		query = query.Where("driver_id = ?", *driverID)
	}
	if vehicleID != nil && *vehicleID != uuid.Nil {
		query = query.Where("vehicle_id = ?", *vehicleID)
	}
	if status != "" && status != "ALL" {
		query = query.Where("status = ?", status)
	}
	if regionType != "" && regionType != "ALL" {
		query = query.Where("region_type = ?", regionType)
	}
	if startDate != nil {
		query = query.Where("departure_date >= ?", *startDate)
	}
	if endDate != nil {
		query = query.Where("departure_date <= ?", *endDate)
	}

	err := query.Find(&trips).Error
	return trips, err
}

func (r *transportRepository) GetTripByID(id uuid.UUID) (*domain.TransportTrip, error) {
	var trip domain.TransportTrip
	err := r.db.Preload("Vehicle").
		Preload("Driver").
		Preload("Helper").
		Preload("Branch").
		Preload("Backhauls.Approver").
		Preload("ExpenseClaims.Approver").
		Preload("Inspection").
		First(&trip, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &trip, nil
}

func (r *transportRepository) UpdateTrip(trip *domain.TransportTrip) error {
	return r.db.Save(trip).Error
}

func (r *transportRepository) DeleteTrip(id uuid.UUID) error {
	return r.db.Delete(&domain.TransportTrip{}, "id = ?", id).Error
}

// ──────────────── BACKHAULS ────────────────
func (r *transportRepository) CreateBackhaul(backhaul *domain.TransportBackhaul) error {
	return r.db.Create(backhaul).Error
}

func (r *transportRepository) GetBackhauls(tripID *uuid.UUID, driverID *uuid.UUID, vehicleID *uuid.UUID, status string, startDate, endDate *time.Time) ([]domain.TransportBackhaul, error) {
	var backhauls []domain.TransportBackhaul
	query := r.db.Model(&domain.TransportBackhaul{}).
		Preload("Trip").
		Preload("Driver").
		Preload("Vehicle").
		Preload("Approver").
		Order("created_at DESC")

	if tripID != nil && *tripID != uuid.Nil {
		query = query.Where("trip_id = ?", *tripID)
	}
	if driverID != nil && *driverID != uuid.Nil {
		query = query.Where("driver_id = ?", *driverID)
	}
	if vehicleID != nil && *vehicleID != uuid.Nil {
		query = query.Where("vehicle_id = ?", *vehicleID)
	}
	if status != "" && status != "ALL" {
		query = query.Where("status = ?", status)
	}
	if startDate != nil {
		query = query.Where("created_at >= ?", *startDate)
	}
	if endDate != nil {
		query = query.Where("created_at <= ?", *endDate)
	}

	err := query.Find(&backhauls).Error
	return backhauls, err
}

func (r *transportRepository) GetBackhaulByID(id uuid.UUID) (*domain.TransportBackhaul, error) {
	var backhaul domain.TransportBackhaul
	err := r.db.Preload("Trip").
		Preload("Driver").
		Preload("Vehicle").
		Preload("Approver").
		First(&backhaul, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &backhaul, nil
}

func (r *transportRepository) UpdateBackhaul(backhaul *domain.TransportBackhaul) error {
	return r.db.Save(backhaul).Error
}

func (r *transportRepository) DeleteBackhaul(id uuid.UUID) error {
	return r.db.Delete(&domain.TransportBackhaul{}, "id = ?", id).Error
}

// ──────────────── EXPENSE CLAIMS ────────────────
func (r *transportRepository) CreateExpenseClaim(claim *domain.TransportExpenseClaim) error {
	return r.db.Create(claim).Error
}

func (r *transportRepository) GetExpenseClaims(tripID *uuid.UUID, vehicleID *uuid.UUID, driverID *uuid.UUID, category string, status string, startDate, endDate *time.Time) ([]domain.TransportExpenseClaim, error) {
	var claims []domain.TransportExpenseClaim
	query := r.db.Model(&domain.TransportExpenseClaim{}).
		Preload("Trip").
		Preload("Vehicle").
		Preload("Driver").
		Preload("Approver").
		Order("created_at DESC")

	if tripID != nil && *tripID != uuid.Nil {
		query = query.Where("trip_id = ?", *tripID)
	}
	if vehicleID != nil && *vehicleID != uuid.Nil {
		query = query.Where("vehicle_id = ?", *vehicleID)
	}
	if driverID != nil && *driverID != uuid.Nil {
		query = query.Where("driver_id = ?", *driverID)
	}
	if category != "" && category != "ALL" {
		query = query.Where("category = ?", category)
	}
	if status != "" && status != "ALL" {
		query = query.Where("status = ?", status)
	}
	if startDate != nil {
		query = query.Where("created_at >= ?", *startDate)
	}
	if endDate != nil {
		query = query.Where("created_at <= ?", *endDate)
	}

	err := query.Find(&claims).Error
	return claims, err
}

func (r *transportRepository) GetExpenseClaimByID(id uuid.UUID) (*domain.TransportExpenseClaim, error) {
	var claim domain.TransportExpenseClaim
	err := r.db.Preload("Trip").
		Preload("Vehicle").
		Preload("Driver").
		Preload("Approver").
		First(&claim, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &claim, nil
}

func (r *transportRepository) UpdateExpenseClaim(claim *domain.TransportExpenseClaim) error {
	return r.db.Save(claim).Error
}

func (r *transportRepository) DeleteExpenseClaim(id uuid.UUID) error {
	return r.db.Delete(&domain.TransportExpenseClaim{}, "id = ?", id).Error
}

// ──────────────── FLEET INSPECTIONS ────────────────
func (r *transportRepository) CreateFleetInspection(inspection *domain.TransportFleetInspection) error {
	return r.db.Create(inspection).Error
}

func (r *transportRepository) GetFleetInspections(vehicleID *uuid.UUID, tripID *uuid.UUID, followUpAction string, followUpStatus string, startDate, endDate *time.Time) ([]domain.TransportFleetInspection, error) {
	var inspections []domain.TransportFleetInspection
	query := r.db.Model(&domain.TransportFleetInspection{}).
		Preload("Vehicle").
		Preload("Driver").
		Preload("Trip").
		Order("inspection_date DESC, created_at DESC")

	if vehicleID != nil && *vehicleID != uuid.Nil {
		query = query.Where("vehicle_id = ?", *vehicleID)
	}
	if tripID != nil && *tripID != uuid.Nil {
		query = query.Where("trip_id = ?", *tripID)
	}
	if followUpAction != "" && followUpAction != "ALL" {
		query = query.Where("follow_up_action = ?", followUpAction)
	}
	if followUpStatus != "" && followUpStatus != "ALL" {
		query = query.Where("follow_up_status = ?", followUpStatus)
	}
	if startDate != nil {
		query = query.Where("inspection_date >= ?", *startDate)
	}
	if endDate != nil {
		query = query.Where("inspection_date <= ?", *endDate)
	}

	err := query.Find(&inspections).Error
	return inspections, err
}

func (r *transportRepository) GetFleetInspectionByID(id uuid.UUID) (*domain.TransportFleetInspection, error) {
	var inspection domain.TransportFleetInspection
	err := r.db.Preload("Vehicle").
		Preload("Driver").
		Preload("Trip").
		First(&inspection, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &inspection, nil
}

func (r *transportRepository) UpdateFleetInspection(inspection *domain.TransportFleetInspection) error {
	return r.db.Save(inspection).Error
}

func (r *transportRepository) DeleteFleetInspection(id uuid.UUID) error {
	return r.db.Delete(&domain.TransportFleetInspection{}, "id = ?", id).Error
}

// ──────────────── DRIVER ATTENDANCES ────────────────
func (r *transportRepository) CreateDriverAttendance(att *domain.TransportDriverAttendance) error {
	return r.db.Create(att).Error
}

func (r *transportRepository) GetDriverAttendances(tripID *uuid.UUID, driverID *uuid.UUID, startDate, endDate *time.Time) ([]domain.TransportDriverAttendance, error) {
	var list []domain.TransportDriverAttendance
	query := r.db.Model(&domain.TransportDriverAttendance{}).
		Preload("Trip").
		Preload("Driver").
		Order("attendance_date DESC, created_at DESC")

	if tripID != nil && *tripID != uuid.Nil {
		query = query.Where("trip_id = ?", *tripID)
	}
	if driverID != nil && *driverID != uuid.Nil {
		query = query.Where("driver_id = ?", *driverID)
	}
	if startDate != nil {
		query = query.Where("attendance_date >= ?", *startDate)
	}
	if endDate != nil {
		query = query.Where("attendance_date <= ?", *endDate)
	}

	err := query.Find(&list).Error
	return list, err
}

// ──────────────── DASHBOARD METRICS ────────────────
func (r *transportRepository) GetDashboardStats(companyID *uuid.UUID, branchID *uuid.UUID) (map[string]interface{}, error) {
	var activeTripsCount int64
	var completedTripsCount int64
	var pendingBackhaulsCount int64
	var approvedBackhaulFeeTotal float64
	var pendingClaimsCount int64
	var approvedClaimsAmountTotal float64
	var fleetsInMaintenanceCount int64

	// Active Trips
	qTrip := r.db.Model(&domain.TransportTrip{}).Where("status IN ?", []string{string(domain.TripStatusInTransit), string(domain.TripStatusScheduled), string(domain.TripStatusArrived), string(domain.TripStatusReturning)})
	if companyID != nil && *companyID != uuid.Nil {
		qTrip = qTrip.Where("company_id = ?", *companyID)
	}
	qTrip.Count(&activeTripsCount)

	// Completed Trips
	r.db.Model(&domain.TransportTrip{}).Where("status = ?", domain.TripStatusCompleted).Count(&completedTripsCount)

	// Backhauls Pending & Total Fee
	r.db.Model(&domain.TransportBackhaul{}).Where("status = ?", domain.TransportApprovalPending).Count(&pendingBackhaulsCount)
	r.db.Model(&domain.TransportBackhaul{}).Where("status = ?", domain.TransportApprovalApproved).Select("COALESCE(SUM(fee_amount), 0)").Scan(&approvedBackhaulFeeTotal)

	// Claims Pending & Total Amount
	r.db.Model(&domain.TransportExpenseClaim{}).Where("status = ?", domain.TransportApprovalPending).Count(&pendingClaimsCount)
	r.db.Model(&domain.TransportExpenseClaim{}).Where("status = ?", domain.TransportApprovalApproved).Select("COALESCE(SUM(amount), 0)").Scan(&approvedClaimsAmountTotal)

	// Fleet maintenance follow ups pending
	r.db.Model(&domain.TransportFleetInspection{}).Where("follow_up_action != ? AND follow_up_status != ?", domain.FollowUpNone, domain.FollowUpStatusCompleted).Count(&fleetsInMaintenanceCount)

	return map[string]interface{}{
		"active_trips_count":            activeTripsCount,
		"completed_trips_count":         completedTripsCount,
		"pending_backhauls_count":       pendingBackhaulsCount,
		"approved_backhaul_fee_total":   approvedBackhaulFeeTotal,
		"pending_claims_count":          pendingClaimsCount,
		"approved_claims_amount_total": approvedClaimsAmountTotal,
		"fleets_in_maintenance_count":   fleetsInMaintenanceCount,
	}, nil
}
