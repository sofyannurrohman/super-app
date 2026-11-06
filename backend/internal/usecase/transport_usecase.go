package usecase

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"example.com/hris/backend/internal/repository"
	"gorm.io/gorm"
)

type TransportUsecase interface {
	// Trips
	CreateTrip(trip *domain.TransportTrip) error
	GetTrips(companyID *uuid.UUID, branchID *uuid.UUID, driverID *uuid.UUID, vehicleID *uuid.UUID, status string, regionType string, startDate, endDate *time.Time) ([]domain.TransportTrip, error)
	GetTripByID(id uuid.UUID) (*domain.TransportTrip, error)
	UpdateTrip(trip *domain.TransportTrip) error
	CompleteTrip(id uuid.UUID, actualReturnDate time.Time, endOdometer int, notes string) (*domain.TransportTrip, error)
	DeleteTrip(id uuid.UUID) error

	// Backhauls
	CreateBackhaul(backhaul *domain.TransportBackhaul) error
	GetBackhauls(tripID *uuid.UUID, driverID *uuid.UUID, vehicleID *uuid.UUID, status string, startDate, endDate *time.Time) ([]domain.TransportBackhaul, error)
	GetBackhaulByID(id uuid.UUID) (*domain.TransportBackhaul, error)
	ApproveBackhaul(id uuid.UUID, approverID uuid.UUID, notes string) (*domain.TransportBackhaul, error)
	RejectBackhaul(id uuid.UUID, approverID uuid.UUID, reason string) (*domain.TransportBackhaul, error)
	DeleteBackhaul(id uuid.UUID) error

	// Expense Claims
	CreateExpenseClaim(claim *domain.TransportExpenseClaim) error
	GetExpenseClaims(tripID *uuid.UUID, vehicleID *uuid.UUID, driverID *uuid.UUID, category string, status string, startDate, endDate *time.Time) ([]domain.TransportExpenseClaim, error)
	GetExpenseClaimByID(id uuid.UUID) (*domain.TransportExpenseClaim, error)
	ApproveExpenseClaim(id uuid.UUID, approverID uuid.UUID) (*domain.TransportExpenseClaim, error)
	RejectExpenseClaim(id uuid.UUID, approverID uuid.UUID, reason string) (*domain.TransportExpenseClaim, error)
	DisburseExpenseClaim(id uuid.UUID, disburserID uuid.UUID) (*domain.TransportExpenseClaim, error)
	DeleteExpenseClaim(id uuid.UUID) error

	// Fleet Inspections
	CreateFleetInspection(inspection *domain.TransportFleetInspection) error
	GetFleetInspections(vehicleID *uuid.UUID, tripID *uuid.UUID, followUpAction string, followUpStatus string, startDate, endDate *time.Time) ([]domain.TransportFleetInspection, error)
	GetFleetInspectionByID(id uuid.UUID) (*domain.TransportFleetInspection, error)
	UpdateInspectionFollowUp(id uuid.UUID, workshopName string, partsReplaced string, actualCost float64, status domain.MaintenanceFollowUpStatus, repairNotes string) (*domain.TransportFleetInspection, error)
	DeleteFleetInspection(id uuid.UUID) error

	// Driver Attendance
	RecordDriverAttendance(att *domain.TransportDriverAttendance) error
	GetDriverAttendances(tripID *uuid.UUID, driverID *uuid.UUID, startDate, endDate *time.Time) ([]domain.TransportDriverAttendance, error)

	// Dashboard Stats
	GetDashboardStats(companyID *uuid.UUID, branchID *uuid.UUID) (map[string]interface{}, error)
}

type transportUsecase struct {
	transportRepo repository.TransportRepository
	vehicleRepo   repository.VehicleRepository
	db            *gorm.DB
}

func NewTransportUsecase(transportRepo repository.TransportRepository, vehicleRepo repository.VehicleRepository, db *gorm.DB) TransportUsecase {
	return &transportUsecase{
		transportRepo: transportRepo,
		vehicleRepo:   vehicleRepo,
		db:            db,
	}
}

// ──────────────── TRIPS IMPLEMENTATION ────────────────

func (u *transportUsecase) CreateTrip(trip *domain.TransportTrip) error {
	// 1. Auto generate trip number
	if trip.TripNumber == "" {
		trip.TripNumber = fmt.Sprintf("TRIP-%s-%04d", time.Now().Format("20060102"), rand.Intn(9000)+1000)
	}

	// 2. Evaluasi Region Type & Daily Allowance Rate
	// Luar Provinsi (Sumatera, Bali, Antar Provinsi jauh) = Rp 150.000 / hari
	// Dalam Provinsi = Rp 100.000 / hari
	if trip.RegionType == "" {
		if strings.EqualFold(strings.TrimSpace(trip.OriginProvince), strings.TrimSpace(trip.DestinationProvince)) && trip.OriginProvince != "" {
			trip.RegionType = domain.TripRegionInProvince
		} else {
			trip.RegionType = domain.TripRegionOutProvince
		}
	}

	if trip.RegionType == domain.TripRegionOutProvince {
		trip.DailyAllowanceRate = 150000
	} else {
		trip.DailyAllowanceRate = 100000
	}

	// 3. Hitung estimasi hari jika belum diset
	if trip.EstimatedDays <= 0 {
		days := int(trip.EstimatedReturnDate.Sub(trip.DepartureDate).Hours()/24) + 1
		if days <= 0 {
			days = 1
		}
		trip.EstimatedDays = days
	}

	// 4. Hitung Total Uang Jalan Driver
	trip.TotalDriverFee = float64(trip.EstimatedDays) * trip.DailyAllowanceRate

	// 5. Simpan Trip
	if err := u.transportRepo.CreateTrip(trip); err != nil {
		return err
	}

	// 6. Update Status Kendaraan ke IN_USE
	if trip.VehicleID != uuid.Nil {
		_ = u.db.Model(&domain.Vehicle{}).Where("id = ?", trip.VehicleID).Update("status", "IN_USE").Error
	}

	return nil
}

func (u *transportUsecase) GetTrips(companyID *uuid.UUID, branchID *uuid.UUID, driverID *uuid.UUID, vehicleID *uuid.UUID, status string, regionType string, startDate, endDate *time.Time) ([]domain.TransportTrip, error) {
	return u.transportRepo.GetTrips(companyID, branchID, driverID, vehicleID, status, regionType, startDate, endDate)
}

func (u *transportUsecase) GetTripByID(id uuid.UUID) (*domain.TransportTrip, error) {
	return u.transportRepo.GetTripByID(id)
}

func (u *transportUsecase) UpdateTrip(trip *domain.TransportTrip) error {
	// Recalculate fees if dates changed
	if trip.RegionType == domain.TripRegionOutProvince {
		trip.DailyAllowanceRate = 150000
	} else {
		trip.DailyAllowanceRate = 100000
	}

	if trip.EstimatedDays > 0 {
		trip.TotalDriverFee = float64(trip.EstimatedDays) * trip.DailyAllowanceRate
	}

	return u.transportRepo.UpdateTrip(trip)
}

func (u *transportUsecase) CompleteTrip(id uuid.UUID, actualReturnDate time.Time, endOdometer int, notes string) (*domain.TransportTrip, error) {
	trip, err := u.transportRepo.GetTripByID(id)
	if err != nil {
		return nil, err
	}

	trip.Status = domain.TripStatusCompleted
	trip.ActualReturnDate = &actualReturnDate
	trip.EndOdometer = &endOdometer
	if notes != "" {
		if trip.Notes != "" {
			trip.Notes = trip.Notes + " | " + notes
		} else {
			trip.Notes = notes
		}
	}

	// Hitung durasi aktual hari
	actualDays := int(actualReturnDate.Sub(trip.DepartureDate).Hours()/24) + 1
	if actualDays <= 0 {
		actualDays = 1
	}
	trip.ActualDays = &actualDays
	trip.TotalDriverFee = float64(actualDays) * trip.DailyAllowanceRate

	if err := u.transportRepo.UpdateTrip(trip); err != nil {
		return nil, err
	}

	// Update mileage kendaraan dan ubah status ke AVAILABLE
	if trip.VehicleID != uuid.Nil {
		_ = u.db.Model(&domain.Vehicle{}).Where("id = ?", trip.VehicleID).Updates(map[string]interface{}{
			"status":  "AVAILABLE",
			"mileage": endOdometer,
		}).Error
	}

	return trip, nil
}

func (u *transportUsecase) DeleteTrip(id uuid.UUID) error {
	return u.transportRepo.DeleteTrip(id)
}

// ──────────────── BACKHAULS (PENGIRIMAN KEMBALI) ────────────────

func (u *transportUsecase) CreateBackhaul(backhaul *domain.TransportBackhaul) error {
	if backhaul.BackhaulNumber == "" {
		backhaul.BackhaulNumber = fmt.Sprintf("BKH-%s-%04d", time.Now().Format("20060102"), rand.Intn(9000)+1000)
	}

	// Fee 10% Calculation
	if backhaul.FeePercentage <= 0 {
		backhaul.FeePercentage = 10.0
	}
	backhaul.FeeAmount = (backhaul.FeePercentage / 100.0) * backhaul.FreightCost
	backhaul.NetAmount = backhaul.FreightCost - backhaul.FeeAmount
	backhaul.Status = domain.TransportApprovalPending

	return u.transportRepo.CreateBackhaul(backhaul)
}

func (u *transportUsecase) GetBackhauls(tripID *uuid.UUID, driverID *uuid.UUID, vehicleID *uuid.UUID, status string, startDate, endDate *time.Time) ([]domain.TransportBackhaul, error) {
	return u.transportRepo.GetBackhauls(tripID, driverID, vehicleID, status, startDate, endDate)
}

func (u *transportUsecase) GetBackhaulByID(id uuid.UUID) (*domain.TransportBackhaul, error) {
	return u.transportRepo.GetBackhaulByID(id)
}

func (u *transportUsecase) ApproveBackhaul(id uuid.UUID, approverID uuid.UUID, notes string) (*domain.TransportBackhaul, error) {
	b, err := u.transportRepo.GetBackhaulByID(id)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	b.Status = domain.TransportApprovalApproved
	b.ApprovedBy = &approverID
	b.ApprovedAt = &now
	if notes != "" {
		b.Notes = b.Notes + " [Approved: " + notes + "]"
	}

	if err := u.transportRepo.UpdateBackhaul(b); err != nil {
		return nil, err
	}
	return b, nil
}

func (u *transportUsecase) RejectBackhaul(id uuid.UUID, approverID uuid.UUID, reason string) (*domain.TransportBackhaul, error) {
	b, err := u.transportRepo.GetBackhaulByID(id)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	b.Status = domain.TransportApprovalRejected
	b.ApprovedBy = &approverID
	b.ApprovedAt = &now
	b.RejectionReason = reason

	if err := u.transportRepo.UpdateBackhaul(b); err != nil {
		return nil, err
	}
	return b, nil
}

func (u *transportUsecase) DeleteBackhaul(id uuid.UUID) error {
	return u.transportRepo.DeleteBackhaul(id)
}

// ──────────────── EXPENSE CLAIMS (STRUK BBM & PART) ────────────────

func (u *transportUsecase) CreateExpenseClaim(claim *domain.TransportExpenseClaim) error {
	if claim.ClaimNumber == "" {
		claim.ClaimNumber = fmt.Sprintf("CLM-%s-%04d", time.Now().Format("20060102"), rand.Intn(9000)+1000)
	}
	claim.Status = domain.TransportApprovalPending
	claim.DisbursementStatus = "UNPAID"

	return u.transportRepo.CreateExpenseClaim(claim)
}

func (u *transportUsecase) GetExpenseClaims(tripID *uuid.UUID, vehicleID *uuid.UUID, driverID *uuid.UUID, category string, status string, startDate, endDate *time.Time) ([]domain.TransportExpenseClaim, error) {
	return u.transportRepo.GetExpenseClaims(tripID, vehicleID, driverID, category, status, startDate, endDate)
}

func (u *transportUsecase) GetExpenseClaimByID(id uuid.UUID) (*domain.TransportExpenseClaim, error) {
	return u.transportRepo.GetExpenseClaimByID(id)
}

func (u *transportUsecase) ApproveExpenseClaim(id uuid.UUID, approverID uuid.UUID) (*domain.TransportExpenseClaim, error) {
	claim, err := u.transportRepo.GetExpenseClaimByID(id)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	claim.Status = domain.TransportApprovalApproved
	claim.ApprovedBy = &approverID
	claim.ApprovedAt = &now

	if err := u.transportRepo.UpdateExpenseClaim(claim); err != nil {
		return nil, err
	}
	return claim, nil
}

func (u *transportUsecase) RejectExpenseClaim(id uuid.UUID, approverID uuid.UUID, reason string) (*domain.TransportExpenseClaim, error) {
	claim, err := u.transportRepo.GetExpenseClaimByID(id)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	claim.Status = domain.TransportApprovalRejected
	claim.ApprovedBy = &approverID
	claim.ApprovedAt = &now
	claim.RejectionReason = reason

	if err := u.transportRepo.UpdateExpenseClaim(claim); err != nil {
		return nil, err
	}
	return claim, nil
}

func (u *transportUsecase) DisburseExpenseClaim(id uuid.UUID, disburserID uuid.UUID) (*domain.TransportExpenseClaim, error) {
	claim, err := u.transportRepo.GetExpenseClaimByID(id)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	claim.DisbursementStatus = "PAID"
	claim.DisbursedAt = &now
	claim.DisbursedBy = &disburserID

	if err := u.transportRepo.UpdateExpenseClaim(claim); err != nil {
		return nil, err
	}
	return claim, nil
}

func (u *transportUsecase) DeleteExpenseClaim(id uuid.UUID) error {
	return u.transportRepo.DeleteExpenseClaim(id)
}

// ──────────────── FLEET INSPECTIONS ────────────────

func (u *transportUsecase) CreateFleetInspection(inspection *domain.TransportFleetInspection) error {
	if inspection.FollowUpAction == "" {
		inspection.FollowUpAction = domain.FollowUpNone
	}

	if inspection.FollowUpAction != domain.FollowUpNone {
		inspection.FollowUpStatus = domain.FollowUpStatusPending
		// Set vehicle status to MAINTENANCE if repairs are required
		if inspection.VehicleID != uuid.Nil {
			_ = u.db.Model(&domain.Vehicle{}).Where("id = ?", inspection.VehicleID).Update("status", "MAINTENANCE").Error
		}
	} else {
		inspection.FollowUpStatus = domain.FollowUpStatusCompleted
	}

	return u.transportRepo.CreateFleetInspection(inspection)
}

func (u *transportUsecase) GetFleetInspections(vehicleID *uuid.UUID, tripID *uuid.UUID, followUpAction string, followUpStatus string, startDate, endDate *time.Time) ([]domain.TransportFleetInspection, error) {
	return u.transportRepo.GetFleetInspections(vehicleID, tripID, followUpAction, followUpStatus, startDate, endDate)
}

func (u *transportUsecase) GetFleetInspectionByID(id uuid.UUID) (*domain.TransportFleetInspection, error) {
	return u.transportRepo.GetFleetInspectionByID(id)
}

func (u *transportUsecase) UpdateInspectionFollowUp(id uuid.UUID, workshopName string, partsReplaced string, actualCost float64, status domain.MaintenanceFollowUpStatus, repairNotes string) (*domain.TransportFleetInspection, error) {
	ins, err := u.transportRepo.GetFleetInspectionByID(id)
	if err != nil {
		return nil, err
	}

	ins.WorkshopName = workshopName
	ins.PartsToReplace = partsReplaced
	ins.ActualCost = actualCost
	ins.FollowUpStatus = status
	ins.RepairNotes = repairNotes

	if status == domain.FollowUpStatusCompleted {
		now := time.Now()
		ins.RepairedAt = &now

		// Pulihkan status kendaraan ke AVAILABLE jika sudah selesai diservis
		if ins.VehicleID != uuid.Nil {
			_ = u.db.Model(&domain.Vehicle{}).Where("id = ?", ins.VehicleID).Update("status", "AVAILABLE").Error
		}
	}

	if err := u.transportRepo.UpdateFleetInspection(ins); err != nil {
		return nil, err
	}
	return ins, nil
}

func (u *transportUsecase) DeleteFleetInspection(id uuid.UUID) error {
	return u.transportRepo.DeleteFleetInspection(id)
}

// ──────────────── DRIVER ATTENDANCES ────────────────

func (u *transportUsecase) RecordDriverAttendance(att *domain.TransportDriverAttendance) error {
	if att.DailyAllowance <= 0 {
		// Dapatkan trip untuk menentukan allowance
		if att.TripID != uuid.Nil {
			if trip, err := u.transportRepo.GetTripByID(att.TripID); err == nil && trip != nil {
				att.DailyAllowance = trip.DailyAllowanceRate
			}
		}
		if att.DailyAllowance <= 0 {
			att.DailyAllowance = 100000 // default fallback
		}
	}
	return u.transportRepo.CreateDriverAttendance(att)
}

func (u *transportUsecase) GetDriverAttendances(tripID *uuid.UUID, driverID *uuid.UUID, startDate, endDate *time.Time) ([]domain.TransportDriverAttendance, error) {
	return u.transportRepo.GetDriverAttendances(tripID, driverID, startDate, endDate)
}

// ──────────────── DASHBOARD STATS ────────────────

func (u *transportUsecase) GetDashboardStats(companyID *uuid.UUID, branchID *uuid.UUID) (map[string]interface{}, error) {
	return u.transportRepo.GetDashboardStats(companyID, branchID)
}
