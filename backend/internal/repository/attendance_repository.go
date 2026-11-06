package repository

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"gorm.io/gorm"
)

type AttendanceRepository interface {
	Create(attendance *domain.AttendanceLog) error
	Update(attendance *domain.AttendanceLog) error
	FindTodayByUserID(userID uuid.UUID) (*domain.AttendanceLog, error)
	FindHistoryByUserID(userID uuid.UUID, limit, offset int) ([]domain.AttendanceLog, error)
	FindHistoryByDateRange(employeeID uuid.UUID, startDate, endDate time.Time) ([]domain.AttendanceLog, error)
	FindAllHistory(limit, offset int, branchIDs []uuid.UUID, employeeID *uuid.UUID, month string, employmentType string) ([]domain.AttendanceLog, error)
	FindByID(id uuid.UUID) (*domain.AttendanceLog, error)
	FindLatestByEmployeeID(employeeID uuid.UUID) (*domain.AttendanceLog, error)
	Delete(id uuid.UUID) error
}

type attendanceRepository struct {
	db *gorm.DB
}

func NewAttendanceRepository(db *gorm.DB) AttendanceRepository {
	return &attendanceRepository{db}
}

func (r *attendanceRepository) Create(attendance *domain.AttendanceLog) error {
	return r.db.Create(attendance).Error
}

func (r *attendanceRepository) Update(attendance *domain.AttendanceLog) error {
	return r.db.Save(attendance).Error
}

func (r *attendanceRepository) FindTodayByUserID(userID uuid.UUID) (*domain.AttendanceLog, error) {
	var attendance domain.AttendanceLog
	// Use WIB (UTC+7) to determine "today" correctly for Indonesian users.
	// Using time.Now().UTC() would shift the date boundary: e.g. 06:00 WIB = 23:00 UTC (previous day),
	// causing yesterday's record to be found and blocking today's check-in.
	wib := time.FixedZone("WIB", 7*3600)
	nowWIB := time.Now().In(wib)
	startOfDayWIB := time.Date(nowWIB.Year(), nowWIB.Month(), nowWIB.Day(), 0, 0, 0, 0, wib)
	endOfDayWIB := startOfDayWIB.Add(24 * time.Hour)

	err := r.db.Where("employee_id = ? AND clock_in_time >= ? AND clock_in_time < ?", userID, startOfDayWIB.UTC(), endOfDayWIB.UTC()).Order("clock_in_time DESC").First(&attendance).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // Return nil, nil if not found today
		}
		return nil, err
	}
	return &attendance, nil
}

func (r *attendanceRepository) FindHistoryByUserID(employeeID uuid.UUID, limit, offset int) ([]domain.AttendanceLog, error) {
	var attendances []domain.AttendanceLog
	err := r.db.Where("employee_id = ?", employeeID).
		Order("clock_in_time DESC").
		Limit(limit).
		Offset(offset).
		Find(&attendances).Error
	return attendances, err
}

func (r *attendanceRepository) FindHistoryByDateRange(employeeID uuid.UUID, startDate, endDate time.Time) ([]domain.AttendanceLog, error) {
	var logs []domain.AttendanceLog
	err := r.db.Where("employee_id = ? AND clock_in_time >= ? AND clock_in_time <= ?", employeeID, startDate, endDate).
		Order("clock_in_time DESC").
		Find(&logs).Error
	return logs, err
}

func (r *attendanceRepository) FindAllHistory(limit, offset int, branchIDs []uuid.UUID, employeeID *uuid.UUID, month string, employmentType string) ([]domain.AttendanceLog, error) {
	var attendances []domain.AttendanceLog
	// The attendance list and XLSX export only use attendance and employee details.
	// Loading every sale, sale item and product for each log makes an all-branch
	// monthly request grow far beyond the number of attendance rows.
	query := r.db.Preload("Employee.User").Preload("Employee.JobPosition").Preload("Employee.Branch")

	if len(branchIDs) > 0 {
		query = query.Where(`(
			attendance_logs.employee_id IN (SELECT id FROM employees WHERE branch_id IN (?))
			OR (
				attendance_logs.is_out_of_town = true 
				AND EXISTS (SELECT 1 FROM branches WHERE id IN (?) AND (name ILIKE '%pabrik%' OR name ILIKE '%head office%' OR name ILIKE '%pusat%'))
			)
		)`, branchIDs, branchIDs)
	}

	if employeeID != nil && *employeeID != uuid.Nil {
		query = query.Where("attendance_logs.employee_id = ?", *employeeID)
	}

	if employmentType != "" {
		if strings.EqualFold(employmentType, "reguler") || strings.EqualFold(employmentType, "regular") {
			query = query.Where("attendance_logs.employee_id IN (SELECT id FROM employees WHERE COALESCE(employment_status, '') NOT ILIKE '%harian%')")
		} else if strings.EqualFold(employmentType, "harian") {
			query = query.Where("attendance_logs.employee_id IN (SELECT id FROM employees WHERE employment_status ILIKE '%harian%')")
		} else {
			query = query.Where("attendance_logs.employee_id IN (SELECT id FROM employees WHERE employment_status = ?)", employmentType)
		}
	}

	if month != "" {
		// Use WIB (UTC+7) for date boundaries to match how dates are displayed in Indonesia
		wib := time.FixedZone("WIB", 7*3600)
		if t, err := time.Parse("2006-01-02", month); err == nil {
			// Single day: start 00:00:00 WIB → end 23:59:59 WIB
			start := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, wib)
			end := start.Add(24*time.Hour - time.Second)
			query = query.Where("attendance_logs.clock_in_time >= ? AND attendance_logs.clock_in_time <= ?", start.UTC(), end.UTC())
		} else if t, err := time.Parse("2006-01", month); err == nil {
			// Month: 1st day 00:00:00 WIB → last day 23:59:59 WIB
			start := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, wib)
			end := start.AddDate(0, 1, 0).Add(-time.Second)
			query = query.Where("attendance_logs.clock_in_time >= ? AND attendance_logs.clock_in_time <= ?", start.UTC(), end.UTC())
		} else if t, err := time.Parse("2006", month); err == nil {
			// Year
			start := time.Date(t.Year(), 1, 1, 0, 0, 0, 0, wib)
			end := start.AddDate(1, 0, 0).Add(-time.Second)
			query = query.Where("attendance_logs.clock_in_time >= ? AND attendance_logs.clock_in_time <= ?", start.UTC(), end.UTC())
		}
	}

	if limit > 0 {
		query = query.Limit(limit).Offset(offset)
	}

	err := query.Order("attendance_logs.clock_in_time DESC").Find(&attendances).Error
	return attendances, err
}

func (r *attendanceRepository) FindByID(id uuid.UUID) (*domain.AttendanceLog, error) {
	var attendance domain.AttendanceLog
	err := r.db.Preload("Employee.User").Where("id = ?", id).First(&attendance).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &attendance, nil
}

func (r *attendanceRepository) Delete(id uuid.UUID) error {
	return r.db.Where("id = ?", id).Delete(&domain.AttendanceLog{}).Error
}

func (r *attendanceRepository) FindLatestByEmployeeID(employeeID uuid.UUID) (*domain.AttendanceLog, error) {
	var attendance domain.AttendanceLog
	err := r.db.Where("employee_id = ?", employeeID).Order("clock_in_time DESC").First(&attendance).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &attendance, nil
}
