package repository

import (
	"time"

	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"gorm.io/gorm"
)

type GpsRepository interface {
	GetDeviceByIMEI(imei string) (*domain.GpsDevice, error)
	RegisterDevice(device *domain.GpsDevice) error
	SaveLocation(location *domain.VehicleLocation) error
	GetLatestLocations(companyID uuid.UUID) ([]domain.VehicleLocation, error)
	GetRouteHistory(vehicleID uuid.UUID, startDate, endDate time.Time) ([]domain.VehicleLocation, error)
	ClearOldLocations(before time.Time) error
}

type gpsRepository struct {
	db *gorm.DB
}

func NewGpsRepository(db *gorm.DB) GpsRepository {
	return &gpsRepository{db}
}

func (r *gpsRepository) GetDeviceByIMEI(imei string) (*domain.GpsDevice, error) {
	var device domain.GpsDevice
	err := r.db.Where("imei = ? AND status = ?", imei, "ACTIVE").First(&device).Error
	return &device, err
}

func (r *gpsRepository) RegisterDevice(device *domain.GpsDevice) error {
	return r.db.Save(device).Error
}

func (r *gpsRepository) SaveLocation(location *domain.VehicleLocation) error {
	// Throttle: Check if there's already a record within the last 1 minute
	var count int64
	oneMinuteAgo := location.Timestamp.Add(-1 * time.Minute)
	r.db.Model(&domain.VehicleLocation{}).
		Where("vehicle_id = ? AND timestamp > ?", location.VehicleID, oneMinuteAgo).
		Count(&count)

	if count > 0 {
		return nil // skip saving to throttle (1 record per minute)
	}

	return r.db.Create(location).Error
}

func (r *gpsRepository) GetLatestLocations(companyID uuid.UUID) ([]domain.VehicleLocation, error) {
	// Using a subquery to get the latest location per vehicle for a company
	var locations []domain.VehicleLocation

	// Join with vehicles to filter by companyID
	subQuery := r.db.Table("vehicle_locations").
		Select("MAX(timestamp) as max_time, vehicle_id").
		Group("vehicle_id")

	err := r.db.Joins("JOIN (?) q ON vehicle_locations.vehicle_id = q.vehicle_id AND vehicle_locations.timestamp = q.max_time", subQuery).
		Joins("JOIN vehicles v ON v.id = vehicle_locations.vehicle_id").
		Where("v.company_id = ?", companyID).
		Preload("Vehicle").
		Find(&locations).Error

	return locations, err
}

func (r *gpsRepository) GetRouteHistory(vehicleID uuid.UUID, startDate, endDate time.Time) ([]domain.VehicleLocation, error) {
	var locations []domain.VehicleLocation
	err := r.db.Where("vehicle_id = ? AND timestamp BETWEEN ? AND ?", vehicleID, startDate, endDate).
		Order("timestamp asc").
		Find(&locations).Error
	return locations, err
}

func (r *gpsRepository) ClearOldLocations(before time.Time) error {
	return r.db.Where("timestamp < ?", before).Delete(&domain.VehicleLocation{}).Error
}
