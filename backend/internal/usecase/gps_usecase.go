package usecase

import (
	"time"

	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"example.com/hris/backend/internal/repository"
)

type GpsUsecase interface {
	RegisterDevice(vehicleID uuid.UUID, imei, simNumber string) error
	GetDeviceByIMEI(imei string) (*domain.GpsDevice, error)
	ProcessLocationData(imei string, latitude, longitude, speed, course float64, isAccOn bool, timestamp time.Time) error
	GetLiveLocations(companyID uuid.UUID) ([]domain.VehicleLocation, error)
	GetRouteHistory(vehicleID uuid.UUID, date time.Time) ([]domain.VehicleLocation, error)
	ClearOldLocations() error
}

type gpsUsecase struct {
	repo repository.GpsRepository
}

func NewGpsUsecase(repo repository.GpsRepository) GpsUsecase {
	return &gpsUsecase{repo}
}

func (u *gpsUsecase) RegisterDevice(vehicleID uuid.UUID, imei, simNumber string) error {
	device := &domain.GpsDevice{
		VehicleID: vehicleID,
		IMEI:      imei,
		SimNumber: simNumber,
		Status:    "ACTIVE",
	}
	return u.repo.RegisterDevice(device)
}

func (u *gpsUsecase) GetDeviceByIMEI(imei string) (*domain.GpsDevice, error) {
	return u.repo.GetDeviceByIMEI(imei)
}

func (u *gpsUsecase) ProcessLocationData(imei string, latitude, longitude, speed, course float64, isAccOn bool, timestamp time.Time) error {
	device, err := u.repo.GetDeviceByIMEI(imei)
	if err != nil {
		return err // device not found or inactive
	}

	loc := &domain.VehicleLocation{
		VehicleID: device.VehicleID,
		Latitude:  latitude,
		Longitude: longitude,
		Speed:     speed,
		Course:    course,
		IsAccOn:   isAccOn,
		Timestamp: timestamp,
	}

	return u.repo.SaveLocation(loc)
}

func (u *gpsUsecase) GetLiveLocations(companyID uuid.UUID) ([]domain.VehicleLocation, error) {
	return u.repo.GetLatestLocations(companyID)
}

func (u *gpsUsecase) GetRouteHistory(vehicleID uuid.UUID, date time.Time) ([]domain.VehicleLocation, error) {
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour).Add(-1 * time.Second)
	return u.repo.GetRouteHistory(vehicleID, startOfDay, endOfDay)
}

func (u *gpsUsecase) ClearOldLocations() error {
	// Clear locations before today
	now := time.Now()
	startOfToday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	return u.repo.ClearOldLocations(startOfToday)
}
