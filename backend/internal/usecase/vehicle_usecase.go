package usecase

import (
	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"example.com/hris/backend/internal/repository"
)

type VehicleUsecase interface {
	CreateVehicle(vehicle *domain.Vehicle, components []domain.VehicleComponent) error
	GetVehicles(companyID *uuid.UUID, branchID *uuid.UUID) ([]domain.Vehicle, error)
	GetVehicleDetail(id uuid.UUID) (*domain.Vehicle, error)
	UpdateVehicle(vehicle *domain.Vehicle, components []domain.VehicleComponent) error
	DeleteVehicle(id uuid.UUID) error
	GetVehicleLogs(vehicleID uuid.UUID) ([]domain.DeliveryBatch, error)
	GetActiveAssignment(vehicleID uuid.UUID) (*domain.ActiveDeliveryAssignment, error)
	RecordMaintenance(maintenance *domain.VehicleMaintenance) error
}

type vehicleUsecase struct {
	repo repository.VehicleRepository
}

func NewVehicleUsecase(repo repository.VehicleRepository) VehicleUsecase {
	return &vehicleUsecase{repo}
}

func (u *vehicleUsecase) CreateVehicle(vehicle *domain.Vehicle, components []domain.VehicleComponent) error {
	return u.repo.CreateWithComponents(vehicle, components)
}

func (u *vehicleUsecase) GetVehicles(companyID *uuid.UUID, branchID *uuid.UUID) ([]domain.Vehicle, error) {
	return u.repo.GetAll(companyID, branchID)
}

func (u *vehicleUsecase) GetVehicleDetail(id uuid.UUID) (*domain.Vehicle, error) {
	return u.repo.GetByID(id)
}

func (u *vehicleUsecase) UpdateVehicle(vehicle *domain.Vehicle, components []domain.VehicleComponent) error {
	return u.repo.UpdateWithComponents(vehicle, components)
}

func (u *vehicleUsecase) DeleteVehicle(id uuid.UUID) error {
	return u.repo.Delete(id)
}

func (u *vehicleUsecase) GetVehicleLogs(vehicleID uuid.UUID) ([]domain.DeliveryBatch, error) {
	return u.repo.GetLogs(vehicleID)
}

func (u *vehicleUsecase) GetActiveAssignment(vehicleID uuid.UUID) (*domain.ActiveDeliveryAssignment, error) {
	return u.repo.GetActiveAssignment(vehicleID)
}

func (u *vehicleUsecase) RecordMaintenance(maintenance *domain.VehicleMaintenance) error {
	return u.repo.RecordMaintenance(maintenance)
}
