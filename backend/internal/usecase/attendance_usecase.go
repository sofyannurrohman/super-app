package usecase

import (
	"encoding/base64"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"example.com/hris/backend/internal/repository"
	"example.com/hris/backend/pkg/utils"
	"github.com/xuri/excelize/v2"
)

type AttendanceUseCase interface {
	CheckIn(userID uuid.UUID, req CheckInRequest) (*AttendanceResponse, error)
	CheckOut(userID uuid.UUID, req CheckOutRequest) (*AttendanceResponse, error)
	GetHistory(userID uuid.UUID, startDate, endDate string) ([]AttendanceHistoryResponse, error)
	GetStats(userID uuid.UUID) (*AttendanceStatsResponse, error)
	GetAllHistory(page, limit int, branchIDs []uuid.UUID, employeeID *uuid.UUID, month string, employmentType string) ([]domain.AttendanceLog, error)
	ExportAttendanceXLSX(branchIDs []uuid.UUID, month string, employmentType string) ([]byte, error)
	CreateManual(req ManualAttendanceRequest) (*domain.AttendanceLog, error)
	UpdateAttendance(id uuid.UUID, req UpdateAttendanceRequest) error
	UpdateMyNotes(userID uuid.UUID, attendanceID uuid.UUID, notes string) error
	DeleteAttendance(id uuid.UUID) error
	ProcessDailyAlpha(date time.Time) error
	ProcessEmployeeAlpha(employeeID uuid.UUID, date time.Time) error
	TabletAttendance(req TabletAttendanceRequest) (*AttendanceResponse, error)
	CheckPinTablet(pin string, branchID string) (*CheckPinTabletResponse, error)
}

type attendanceUseCase struct {
	repo              repository.AttendanceRepository
	employeeRepo      repository.EmployeeRepository
	employeeShiftRepo repository.EmployeeShiftRepository
	leaveRepo         repository.LeaveRepository
}

func NewAttendanceUseCase(
	repo repository.AttendanceRepository,
	employeeRepo repository.EmployeeRepository,
	employeeShiftRepo repository.EmployeeShiftRepository,
	leaveRepo repository.LeaveRepository,
) AttendanceUseCase {
	return &attendanceUseCase{
		repo:              repo,
		employeeRepo:      employeeRepo,
		employeeShiftRepo: employeeShiftRepo,
		leaveRepo:         leaveRepo,
	}
}

type CheckInRequest struct {
	Latitude      float64   `json:"latitude" binding:"required"`
	Longitude     float64   `json:"longitude" binding:"required"`
	Selfie        string    `json:"selfie"` // Base64 string
	FaceEmbedding []float32 `json:"face_embedding"`
	IsOutOfTown   bool      `json:"is_out_of_town"`
	Destination   string    `json:"destination"`
}

type CheckOutRequest struct {
	Latitude      float64   `json:"latitude" binding:"required"`
	Longitude     float64   `json:"longitude" binding:"required"`
	Selfie        string    `json:"selfie"`
	FaceEmbedding []float32 `json:"face_embedding"`
	IsOutOfTown   bool      `json:"is_out_of_town"`
	Destination   string    `json:"destination"`
}

type AttendanceHistoryResponse struct {
	ID               string     `json:"id"`
	UserID           string     `json:"user_id"`
	CheckIn          time.Time  `json:"check_in"`
	CheckOut         *time.Time `json:"check_out"`
	Status           string     `json:"status"`
	Notes            string     `json:"notes"`
	SelfieURL        string     `json:"selfie_url"`
	ClockInPhotoURL  string     `json:"clock_in_photo_url"`
	ClockOutPhotoURL string     `json:"clock_out_photo_url"`
	WorkDuration     int        `json:"work_duration"`
	IsOutOfTown      bool       `json:"is_out_of_town"`
	Destination      string     `json:"destination"`
}

type AttendanceResponse struct {
	ID               string     `json:"id"`
	UserID           string     `json:"user_id"`
	CheckIn          *time.Time `json:"check_in"`
	CheckOut         *time.Time `json:"check_out"`
	Status           string     `json:"status"`
	Notes            string     `json:"notes"`
	SelfieURL        string     `json:"selfie_url"`
	ClockInPhotoURL  string     `json:"clock_in_photo_url"`
	ClockOutPhotoURL string     `json:"clock_out_photo_url"`
	WorkDuration     int        `json:"work_duration"`
	IsOutOfTown      bool       `json:"is_out_of_town"`
	Destination      string     `json:"destination"`
}

type AttendanceStatsResponse struct {
	TotalHours     float64 `json:"total_hours"`
	OvertimeHours  float64 `json:"overtime_hours"`
	AttendanceDays int     `json:"attendance_days"`
	LeaveDays      int     `json:"leave_days"`
}

type ManualAttendanceRequest struct {
	EmployeeID   string `json:"employee_id" binding:"required"`
	ClockInTime  string `json:"clock_in_time" binding:"required"` // Format: YYYY-MM-DDTHH:mm:ssZ
	ClockOutTime string `json:"clock_out_time"`
	Status       string `json:"status" binding:"required"`
	Notes        string `json:"notes"`
	IsOutOfTown  bool   `json:"is_out_of_town"`
	Destination  string `json:"destination"`
}

type UpdateAttendanceRequest struct {
	ClockInTime  string `json:"clock_in_time"`
	ClockOutTime string `json:"clock_out_time"`
	Status       string `json:"status"`
	Notes        string `json:"notes"`
	IsOutOfTown  *bool  `json:"is_out_of_town"`
	Destination  string `json:"destination"`
}

type TabletAttendanceRequest struct {
	BranchID       string `json:"branch_id" binding:"required"`
	EmployeeID     string `json:"employee_id"`
	Pin            string `json:"pin"`
	AttendanceType string `json:"attendance_type" binding:"required"` // "check_in" or "clock_out"
	PhotoBase64    string `json:"photo_base64"`
}

func (u *attendanceUseCase) TabletAttendance(req TabletAttendanceRequest) (*AttendanceResponse, error) {
	var employee *domain.Employee
	var err error

	if req.EmployeeID != "" {
		empID, parseErr := uuid.Parse(req.EmployeeID)
		if parseErr != nil {
			return nil, errors.New("invalid employee_id format")
		}
		employee, err = u.employeeRepo.FindByID(empID)
		if err != nil {
			return nil, err
		}
	} else if req.Pin != "" {
		branchID, parseErr := uuid.Parse(req.BranchID)
		if parseErr != nil {
			return nil, errors.New("invalid branch_id format")
		}
		employee, err = u.employeeRepo.FindByPin(req.Pin, &branchID)
		if err != nil {
			return nil, err
		}
	}

	if employee == nil {
		return nil, errors.New("employee not found with the given PIN")
	}

	var selfieURL string
	if req.PhotoBase64 != "" {
		uploadDir := "uploads/selfies_kiosk"
		if err := os.MkdirAll(uploadDir, 0755); err != nil {
			return nil, err
		}

		parts := strings.Split(req.PhotoBase64, ",")
		base64Data := req.PhotoBase64
		if len(parts) > 1 {
			base64Data = parts[1]
		}

		imageData, err := base64.StdEncoding.DecodeString(base64Data)
		if err != nil {
			return nil, errors.New("invalid photo base64 data")
		}

		compressedData, err := utils.CompressImage(imageData, 800, 75)
		if err != nil {
			return nil, errors.New("failed to compress image")
		}

		filename := fmt.Sprintf("%s_%d_kiosk.jpg", employee.ID.String(), time.Now().Unix())
		filePath := filepath.Join(uploadDir, filename)

		if err := os.WriteFile(filePath, compressedData, 0644); err != nil {
			return nil, err
		}
		selfieURL = "/" + filePath
	}

	now := time.Now().UTC()
	var response *AttendanceResponse

	if req.AttendanceType == "check_in" {
		existing, err := u.repo.FindTodayByUserID(employee.ID)
		if err != nil {
			return nil, err
		}
		if existing != nil {
			return nil, errors.New("hari ini sudah absen masuk")
		}

		attendance := &domain.AttendanceLog{
			EmployeeID:      employee.ID,
			ClockInTime:     &now,
			Status:          "ON_TIME", // Simple logic for kiosk
			ClockInPhotoURL: selfieURL,
		}

		if err := u.repo.Create(attendance); err != nil {
			return nil, err
		}

		t := attendance.ClockInTime.UTC()
		response = &AttendanceResponse{
			ID:               attendance.ID.String(),
			UserID:           employee.ID.String(),
			CheckIn:          &t,
			Status:           attendance.Status,
			SelfieURL:        attendance.ClockInPhotoURL,
			ClockInPhotoURL:  attendance.ClockInPhotoURL,
			ClockOutPhotoURL: "",
		}
	} else if req.AttendanceType == "clock_out" {
		existing, err := u.repo.FindLatestByEmployeeID(employee.ID)
		if err != nil {
			return nil, err
		}
		if existing == nil {
			return nil, errors.New("anda belum melakukan absen masuk (clock in) hari ini")
		}
		if existing.ClockOutTime != nil {
			return nil, errors.New("anda sudah melakukan absen pulang hari ini")
		}

		existing.ClockOutTime = &now
		existing.ClockOutPhotoURL = selfieURL

		if err := u.repo.Update(existing); err != nil {
			return nil, errors.New("gagal menyimpan data absen pulang")
		}

		duration := 0
		if existing.ClockInTime != nil {
			duration = int(now.Sub(*existing.ClockInTime).Minutes())
		}
		var checkInUTC *time.Time
		if existing.ClockInTime != nil {
			t := existing.ClockInTime.UTC()
			checkInUTC = &t
		}
		var checkOutUTC *time.Time
		if existing.ClockOutTime != nil {
			t := existing.ClockOutTime.UTC()
			checkOutUTC = &t
		}

		response = &AttendanceResponse{
			ID:               existing.ID.String(),
			UserID:           employee.ID.String(),
			CheckIn:          checkInUTC,
			CheckOut:         checkOutUTC,
			Status:           existing.Status,
			SelfieURL:        existing.ClockOutPhotoURL,
			ClockInPhotoURL:  existing.ClockInPhotoURL,
			ClockOutPhotoURL: existing.ClockOutPhotoURL,
			WorkDuration:     duration,
		}
	} else {
		return nil, errors.New("Tipe absensi tidak valid")
	}

	return response, nil
}

type CheckPinTabletResponse struct {
	Employee      *domain.Employee
	HasCheckedIn  bool
	HasClockedOut bool
}

func (u *attendanceUseCase) CheckPinTablet(pin string, branchID string) (*CheckPinTabletResponse, error) {
	parsedBranchID, err := uuid.Parse(branchID)
	if err != nil {
		return nil, errors.New("Format ID cabang tidak valid")
	}

	employee, err := u.employeeRepo.FindByPin(pin, &parsedBranchID)
	if err != nil {
		return nil, err
	}

	if employee == nil {
		return nil, errors.New("Pekerja tidak ditemukan dengan PIN tersebut")
	}

	existing, _ := u.repo.FindTodayByUserID(employee.ID)

	res := &CheckPinTabletResponse{
		Employee:      employee,
		HasCheckedIn:  existing != nil,
		HasClockedOut: existing != nil && existing.ClockOutTime != nil,
	}

	return res, nil
}

func (u *attendanceUseCase) CheckIn(userID uuid.UUID, req CheckInRequest) (*AttendanceResponse, error) {
	// 1. Resolve Employee from UserID
	employee, err := u.employeeRepo.FindByUserID(userID)
	if err != nil {
		return nil, err
	}
	if employee == nil {
		return nil, errors.New("employee record not found for this user")
	}

	// 2. Validate Geofencing (Skip if out of town / dinas luar kota or user has CanAbsenAnywhere privilege)
	canAbsenAnywhere := employee.User != nil && employee.User.CanAbsenAnywhere
	if !req.IsOutOfTown && !canAbsenAnywhere {
		if employee.Branch == nil {
			return nil, errors.New("employee branch not found")
		}

		distance := utils.CalculateDistance(req.Latitude, req.Longitude, employee.Branch.Latitude, employee.Branch.Longitude)
		if int(distance) > employee.Branch.RadiusMeter {
			return nil, fmt.Errorf("outside office radius. Distance: %dm, Allowed: %dm", int(distance), employee.Branch.RadiusMeter)
		}
	}

	// 3. Validate AI Face Match (Only if face embedding is provided, e.g. from mobile app)
	if len(req.FaceEmbedding) > 0 {
		if !employee.IsFaceRegistered || len(employee.FaceEmbedding) == 0 {
			return nil, errors.New("wajah Anda belum teregistrasi di sistem")
		}

		faceDistance, err := utils.CalculateEuclideanDistance(employee.FaceEmbedding, req.FaceEmbedding)
		if err != nil {
			return nil, errors.New("gagal memvalidasi data pengenalan wajah")
		}

		// Threshold MLKit embeddings (generally distances > 1.7 indicates a mismatch)
		if faceDistance > 1.7 {
			return nil, fmt.Errorf("wajah tidak cocok (distance: %.2f)", faceDistance)
		}
	}

	// 4. Check if already checked in today
	existing, err := u.repo.FindTodayByUserID(employee.ID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("hari ini sudah absen masuk")
	}

	// 4. Handle Selfie Upload if provided
	var selfieURL string
	if req.Selfie != "" {
		// Ensure directory exists
		uploadDir := "uploads/selfies"
		if err := os.MkdirAll(uploadDir, 0755); err != nil {
			return nil, err
		}

		// Decode base64
		imageData, err := base64.StdEncoding.DecodeString(req.Selfie)
		if err != nil {
			return nil, errors.New("data gambar selfi tidak valid ")
		}

		// Compress image
		compressedData, err := utils.CompressImage(imageData, 800, 75)
		if err != nil {
			return nil, errors.New("failed to compress image")
		}

		// Generate filename
		filename := fmt.Sprintf("%s_%d.jpg", employee.ID.String(), time.Now().Unix())
		filePath := filepath.Join(uploadDir, filename)

		// Save to file
		if err := os.WriteFile(filePath, compressedData, 0644); err != nil {
			return nil, err
		}

		selfieURL = "/" + filePath
	}

	// 5. Evaluate 'late' status based on applicable shift
	// Priority: Specific Date Shift -> Employee Default Shift -> Branch Default Shift
	var applicableShift *domain.Shift
	if len(employee.EmployeeShifts) > 0 {
		applicableShift = employee.EmployeeShifts[0].Shift
	} else if employee.DefaultShift != nil {
		applicableShift = employee.DefaultShift
	} else if employee.Branch != nil && employee.Branch.DefaultShift != nil {
		applicableShift = employee.Branch.DefaultShift
	}

	status := "ON_TIME"
	now := time.Now().UTC()
	// Convert 'now' to Asia/Jakarta for comparison since shift times are usually local
	loc, _ := time.LoadLocation("Asia/Jakarta")
	nowLocal := now.In(loc)

	if applicableShift != nil {
		// Clean and parse shift StartTime robustly (handles "07:30:00", "07:30", and GORM's ISO-8601 string "0000-01-01T07:30:00+07:07")
		tStr := applicableShift.StartTime
		if len(tStr) > 11 && tStr[10] == 'T' {
			tStr = tStr[11:19]
		}

		var shiftTime time.Time
		var err error

		shiftTime, err = time.Parse("15:04:05", tStr)
		if err != nil {
			shiftTime, err = time.Parse("15:04", tStr)
		}

		if err == nil {
			// Create a comparison time for today
			compareTime := time.Date(nowLocal.Year(), nowLocal.Month(), nowLocal.Day(), shiftTime.Hour(), shiftTime.Minute(), shiftTime.Second(), 0, loc)

			// Late if check-in is after shift start time (with a 15-min grace period)
			if nowLocal.After(compareTime.Add(15 * time.Minute)) {
				status = "LATE"
			}
		}
	} else {
		// Fallback to default 08:15 logic if no shift found
		cutoff := time.Date(nowLocal.Year(), nowLocal.Month(), nowLocal.Day(), 8, 15, 0, 0, loc)
		if nowLocal.After(cutoff) {
			status = "LATE"
		}
	}

	var shiftID *uuid.UUID
	if applicableShift != nil {
		shiftID = &applicableShift.ID
	}

	isOutOfTown := req.IsOutOfTown
	destination := req.Destination
	if !isOutOfTown && employee.OperatingBranchID != nil && employee.BranchID != nil && *employee.OperatingBranchID != *employee.BranchID {
		isOutOfTown = true
		if destination == "" {
			if employee.OperatingBranch != nil && employee.OperatingBranch.Name != "" {
				destination = employee.OperatingBranch.Name
			} else {
				destination = "Dinas Luar Kota"
			}
		}
	}

	locStatus := ""
	if isOutOfTown {
		locStatus = "OUT_OF_TOWN"
	}

	attendance := &domain.AttendanceLog{
		EmployeeID:            employee.ID,
		ShiftID:               shiftID,
		ClockInTime:           &now,
		ClockInLat:            req.Latitude,
		ClockInLong:           req.Longitude,
		ClockInLocationStatus: locStatus,
		Status:                status,
		ClockInPhotoURL:       selfieURL,
		IsOutOfTown:           isOutOfTown,
		Destination:           destination,
	}

	if err := u.repo.Create(attendance); err != nil {
		return nil, err
	}

	// Normalisasi ke UTC agar client (mobile/web) dapat mengkonversi ke timezone lokal dengan benar.
	var checkInUTC *time.Time
	if attendance.ClockInTime != nil {
		t := attendance.ClockInTime.UTC()
		checkInUTC = &t
	}
	return &AttendanceResponse{
		ID:               attendance.ID.String(),
		UserID:           employee.ID.String(),
		CheckIn:          checkInUTC,
		Status:           attendance.Status,
		Notes:            attendance.Notes,
		SelfieURL:        attendance.ClockInPhotoURL,
		ClockInPhotoURL:  attendance.ClockInPhotoURL,
		ClockOutPhotoURL: "",
		WorkDuration:     0,
		IsOutOfTown:      attendance.IsOutOfTown,
		Destination:      attendance.Destination,
	}, nil
}

func (u *attendanceUseCase) CheckOut(userID uuid.UUID, req CheckOutRequest) (*AttendanceResponse, error) {
	// 1. Resolve Employee from UserID
	employee, err := u.employeeRepo.FindByUserID(userID)
	if err != nil {
		return nil, err
	}
	if employee == nil {
		return nil, errors.New("employee record not found for this user")
	}

	// 2. Find the latest attendance record for this employee
	existing, err := u.repo.FindLatestByEmployeeID(employee.ID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, errors.New("anda belum melakukan absen masuk (clock in) hari ini")
	}
	if existing.ClockOutTime != nil {
		return nil, errors.New("anda sudah melakukan absen pulang hari ini")
	}

	// 3. Validate Geofencing (Skip if out of town or user has CanAbsenAnywhere privilege)
	isOutOfTown := existing.IsOutOfTown || req.IsOutOfTown
	canAbsenAnywhere := employee.User != nil && employee.User.CanAbsenAnywhere
	if !isOutOfTown && !canAbsenAnywhere {
		if employee.Branch == nil {
			return nil, errors.New("employee branch not found")
		}

		distance := utils.CalculateDistance(req.Latitude, req.Longitude, employee.Branch.Latitude, employee.Branch.Longitude)
		if int(distance) > employee.Branch.RadiusMeter {
			return nil, fmt.Errorf("berada di luar radius kantor. Jarak: %dm, Diizinkan: %dm", int(distance), employee.Branch.RadiusMeter)
		}
	}

	// 4. Validate AI Face Match (Only if face embedding is provided, e.g. from mobile app)
	if len(req.FaceEmbedding) > 0 {
		if !employee.IsFaceRegistered || len(employee.FaceEmbedding) == 0 {
			return nil, errors.New("wajah Anda belum teregistrasi di sistem")
		}

		faceDistance, err := utils.CalculateEuclideanDistance(employee.FaceEmbedding, req.FaceEmbedding)
		if err != nil {
			return nil, errors.New("gagal memvalidasi data pengenalan wajah")
		}

		// Threshold MLKit embeddings (generally distances > 1.7 indicates a mismatch)
		if faceDistance > 1.7 {
			return nil, fmt.Errorf("wajah tidak cocok (distance: %.2f)", faceDistance)
		}
	}

	// 5. Handle Selfie Upload for CheckOut
	var selfieURL string
	if req.Selfie != "" {
		uploadDir := "uploads/selfies"
		if err := os.MkdirAll(uploadDir, 0755); err != nil {
			return nil, err
		}

		imageData, err := base64.StdEncoding.DecodeString(req.Selfie)
		if err != nil {
			return nil, errors.New("data gambar selfie pulang tidak valid")
		}

		compressedData, err := utils.CompressImage(imageData, 800, 75)
		if err != nil {
			return nil, errors.New("failed to compress image")
		}

		filename := fmt.Sprintf("%s_%d_out.jpg", employee.ID.String(), time.Now().Unix())
		filePath := filepath.Join(uploadDir, filename)

		if err := os.WriteFile(filePath, compressedData, 0644); err != nil {
			return nil, err
		}

		selfieURL = "/" + filePath
	} else {
		return nil, errors.New("foto absen pulang wajib dilampirkan")
	}

	now := time.Now().UTC()

	existing.ClockOutTime = &now
	existing.ClockOutLat = req.Latitude
	existing.ClockOutLong = req.Longitude
	existing.ClockOutPhotoURL = selfieURL
	if req.IsOutOfTown {
		existing.IsOutOfTown = true
	}
	if req.Destination != "" {
		existing.Destination = req.Destination
	}

	if err := u.repo.Update(existing); err != nil {
		return nil, errors.New("gagal menyimpan data absen pulang")
	}

	duration := 0
	if existing.ClockInTime != nil {
		duration = int(now.Sub(*existing.ClockInTime).Minutes())
	}

	// Normalisasi ke UTC agar client (mobile/web) dapat mengkonversi ke timezone lokal dengan benar.
	var checkInUTC, checkOutUTC *time.Time
	if existing.ClockInTime != nil {
		t := existing.ClockInTime.UTC()
		checkInUTC = &t
	}
	if existing.ClockOutTime != nil {
		t := existing.ClockOutTime.UTC()
		checkOutUTC = &t
	}
	return &AttendanceResponse{
		ID:               existing.ID.String(),
		UserID:           employee.ID.String(),
		CheckIn:          checkInUTC,
		CheckOut:         checkOutUTC,
		Status:           existing.Status,
		Notes:            existing.Notes,
		SelfieURL:        existing.ClockOutPhotoURL,
		ClockInPhotoURL:  existing.ClockInPhotoURL,
		ClockOutPhotoURL: existing.ClockOutPhotoURL,
		WorkDuration:     duration,
		IsOutOfTown:      existing.IsOutOfTown,
		Destination:      existing.Destination,
	}, nil
}

func (u *attendanceUseCase) GetHistory(userID uuid.UUID, startDate, endDate string) ([]AttendanceHistoryResponse, error) {
	employee, err := u.employeeRepo.FindByUserID(userID)
	if err != nil {
		return nil, err
	}
	if employee == nil {
		return nil, errors.New("employee record not found")
	}

	// 1. Set default date range jika kosong (Bulan Berjalan) dalam zona WIB (UTC+7)
	wib := time.FixedZone("WIB", 7*3600)
	now := time.Now().In(wib)
	var start, end time.Time

	if startDate == "" {
		start = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, wib)
	} else {
		parsedStart, err := time.Parse("2006-01-02", startDate)
		if err == nil {
			start = time.Date(parsedStart.Year(), parsedStart.Month(), parsedStart.Day(), 0, 0, 0, 0, wib)
		} else {
			start = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, wib)
		}
	}

	if endDate == "" {
		end = time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, int(time.Second-time.Nanosecond), wib)
	} else {
		parsedEnd, err := time.Parse("2006-01-02", endDate)
		if err == nil {
			end = time.Date(parsedEnd.Year(), parsedEnd.Month(), parsedEnd.Day(), 23, 59, 59, int(time.Second-time.Nanosecond), wib)
		} else {
			end = time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, int(time.Second-time.Nanosecond), wib)
		}
	}

	// 1.5. Trigger ALFA check for yesterday if within range
	yesterday := now.AddDate(0, 0, -1)
	if (yesterday.After(start) || yesterday.Equal(start)) && (yesterday.Before(end) || yesterday.Equal(end)) {
		_ = u.ProcessEmployeeAlpha(employee.ID, yesterday)
	}

	// 2. Ambil data dari Repository (konversi ke UTC untuk query database)
	logs, err := u.repo.FindHistoryByDateRange(employee.ID, start.UTC(), end.UTC())
	if err != nil {
		return nil, errors.New("gagal mengambil riwayat absensi")
	}

	// 3. Mapping data dari Entitas Database (AttendanceLog) ke DTO (AttendanceHistoryResponse)
	var history []AttendanceHistoryResponse
	for _, log := range logs {
		checkIn := time.Time{}
		if log.ClockInTime != nil {
			// Normalize ke UTC agar Flutter tidak double-convert berdasarkan timezone perangkat.
			// GORM dengan DSN TimeZone=Asia/Jakarta bisa mengembalikan time dengan offset +07:00,
			// sehingga Flutter yang menggunakan .toLocal() di device UTC+8 (WIT/WITA) akan
			// menampilkan waktu +1 jam lebih lambat dari yang seharusnya.
			checkIn = log.ClockInTime.UTC()
		}
		var checkOutUTC *time.Time
		if log.ClockOutTime != nil {
			t := log.ClockOutTime.UTC()
			checkOutUTC = &t
		}
		duration := 0
		if log.ClockInTime != nil && log.ClockOutTime != nil {
			duration = int(log.ClockOutTime.Sub(*log.ClockInTime).Minutes())
		}
		selfieURL := log.ClockInPhotoURL
		if selfieURL == "" && log.ClockOutPhotoURL != "" {
			selfieURL = log.ClockOutPhotoURL
		}
		history = append(history, AttendanceHistoryResponse{
			ID:               log.ID.String(),
			UserID:           log.EmployeeID.String(),
			CheckIn:          checkIn,
			CheckOut:         checkOutUTC,
			Status:           log.Status,
			Notes:            log.Notes,
			SelfieURL:        selfieURL,
			ClockInPhotoURL:  log.ClockInPhotoURL,
			ClockOutPhotoURL: log.ClockOutPhotoURL,
			WorkDuration:     duration,
			IsOutOfTown:      log.IsOutOfTown,
			Destination:      log.Destination,
		})
	}

	// Jika kosong, kembalikan array kosong (bukan null) agar Flutter mudah memprosesnya
	if history == nil {
		history = []AttendanceHistoryResponse{}
	}

	return history, nil
}

func (u *attendanceUseCase) GetStats(userID uuid.UUID) (*AttendanceStatsResponse, error) {
	employee, err := u.employeeRepo.FindByUserID(userID)
	if err != nil {
		return nil, err
	}
	if employee == nil {
		return nil, errors.New("employee record not found")
	}

	// For simplicity, get current month stats in WIB
	wib := time.FixedZone("WIB", 7*3600)
	now := time.Now().In(wib)
	start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, wib)
	end := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, int(time.Second-time.Nanosecond), wib)

	// Sync ALFA for yesterday
	yesterday := now.AddDate(0, 0, -1)
	_ = u.ProcessEmployeeAlpha(employee.ID, yesterday)

	logs, err := u.repo.FindHistoryByDateRange(employee.ID, start.UTC(), end.UTC())
	if err != nil {
		return nil, err
	}

	stats := &AttendanceStatsResponse{
		AttendanceDays: len(logs),
	}

	var totalDuration time.Duration
	for _, log := range logs {
		if log.ClockInTime != nil && log.ClockOutTime != nil {
			duration := log.ClockOutTime.Sub(*log.ClockInTime)
			totalDuration += duration
		}
	}

	stats.TotalHours = totalDuration.Hours()
	// Dummy for now until Leave/Overtime integrated correctly
	stats.OvertimeHours = 0
	stats.LeaveDays = 0

	return stats, nil
}

func (u *attendanceUseCase) GetAllHistory(page, limit int, branchIDs []uuid.UUID, employeeID *uuid.UUID, month string, employmentType string) ([]domain.AttendanceLog, error) {
	var offset int
	if limit > 0 {
		if page < 1 {
			page = 1
		}
		offset = (page - 1) * limit
	}
	return u.repo.FindAllHistory(limit, offset, branchIDs, employeeID, month, employmentType)
}

func convertToJakarta(t *time.Time) time.Time {
	if t == nil {
		return time.Time{}
	}
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		loc = time.FixedZone("WIB", 7*3600)
	}
	// Always convert via UTC first to avoid double-conversion issues
	// Regardless of what timezone the Go time.Time value carries,
	// .UTC() normalizes it, then .In(loc) converts to Jakarta time correctly.
	return t.UTC().In(loc)
}

func (u *attendanceUseCase) ExportAttendanceXLSX(branchIDs []uuid.UUID, month string, employmentType string) ([]byte, error) {
	// 1. Fetch all employees in this branch (or all globally if branchIDs is empty)
	allEmployees, err := u.employeeRepo.FindAll(0, branchIDs, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch employees: %v", err)
	}

	// Filter out inactive employees and apply employmentType filter
	var employees []domain.Employee
	for _, emp := range allEmployees {
		// Exclude inactive employees
		if emp.User != nil && !emp.User.IsActive {
			continue
		}
		empStatus := strings.ToUpper(strings.TrimSpace(emp.EmploymentStatus))
		if empStatus == "INACTIVE" || empStatus == "RESIGNED" || empStatus == "TERMINATED" || empStatus == "NON-AKTIF" {
			continue
		}

		// Filter by employmentType if provided
		if employmentType != "" {
			isHarian := strings.Contains(strings.ToLower(emp.EmploymentStatus), "harian")
			if strings.EqualFold(employmentType, "reguler") || strings.EqualFold(employmentType, "regular") {
				if isHarian {
					continue
				}
			} else if strings.EqualFold(employmentType, "harian") {
				if !isHarian {
					continue
				}
			} else {
				if !strings.EqualFold(emp.EmploymentStatus, employmentType) {
					continue
				}
			}
		}

		employees = append(employees, emp)
	}

	// 2. Fetch all logs for this branch/period
	logs, err := u.repo.FindAllHistory(0, 0, branchIDs, nil, month, employmentType)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch attendance logs: %v", err)
	}

	// 3. Determine the date range based on month parameter (daily, monthly, yearly, or current month default)
	var startDate, endDate time.Time
	now := time.Now()
	if month == "" {
		month = now.Format("2006-01")
	}

	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		loc = time.FixedZone("WIB", 7*3600)
	}

	if t, err := time.Parse("2006-01-02", month); err == nil {
		startDate = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, loc)
		endDate = startDate.Add(24 * time.Hour).Add(-time.Second)
	} else if t, err := time.Parse("2006-01", month); err == nil {
		startDate = time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, loc)
		endDate = startDate.AddDate(0, 1, 0).Add(-time.Second)
	} else if t, err := time.Parse("2006", month); err == nil {
		startDate = time.Date(t.Year(), 1, 1, 0, 0, 0, 0, loc)
		endDate = startDate.AddDate(1, 0, 0).Add(-time.Second)
	} else {
		startDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc)
		endDate = startDate.AddDate(0, 1, 0).Add(-time.Second)
	}

	var dates []time.Time
	for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
		dates = append(dates, d)
	}

	// 4. Map leaves for each employee
	type LeaveKey struct {
		EmployeeID uuid.UUID
		DateStr    string // YYYY-MM-DD
	}
	leaveMap := make(map[LeaveKey]string) // Value: "LEAVE", "SICK", "PERMISSION"

	for _, emp := range employees {
		leaves, err := u.leaveRepo.FindApprovedByEmployeeAndDateRange(emp.ID, startDate, endDate)
		if err == nil {
			for _, l := range leaves {
				lStart := l.StartDate.In(loc)
				lEnd := l.EndDate.In(loc)
				for d := lStart; !d.After(lEnd); d = d.AddDate(0, 0, 1) {
					key := LeaveKey{
						EmployeeID: emp.ID,
						DateStr:    d.Format("2006-01-02"),
					}
					leaveType := "LEAVE"
					if l.LeaveType != nil {
						leaveType = l.LeaveType.Name
					}
					leaveMap[key] = leaveType
				}
			}
		}
	}

	// 5. Map attendance logs by employee ID and date
	type LogKey struct {
		EmployeeID uuid.UUID
		DateStr    string // YYYY-MM-DD
	}
	logMap := make(map[LogKey]domain.AttendanceLog)
	for _, l := range logs {
		if l.ClockInTime != nil {
			localIn := convertToJakarta(l.ClockInTime)
			key := LogKey{
				EmployeeID: l.EmployeeID,
				DateStr:    localIn.Format("2006-01-02"),
			}
			logMap[key] = l
		}
	}

	// 6. Create Excel file using excelize
	f := excelize.NewFile()
	defer f.Close()

	sheetName := "Laporan Kehadiran"
	f.SetSheetName("Sheet1", sheetName)

	// Set Page Grid lines visible
	showGridLines := true
	f.SetSheetView("Laporan Kehadiran", 0, &excelize.ViewOptions{
		ShowGridLines: &showGridLines,
	})

	// Styles
	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 16, Color: "1F4E78"},
	})
	subtitleStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Italic: true, Size: 11, Color: "595959"},
	})

	headerStyle, _ := f.NewStyle(&excelize.Style{
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"2F3542"}, Pattern: 1},
		Font:      &excelize.Font{Bold: true, Color: "FFFFFF", Size: 11},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
		Border: []excelize.Border{
			{Type: "top", Color: "D9D9D9", Style: 1},
			{Type: "bottom", Color: "D9D9D9", Style: 1},
			{Type: "left", Color: "D9D9D9", Style: 1},
			{Type: "right", Color: "D9D9D9", Style: 1},
		},
	})

	dateStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 10, Color: "2F3542"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
		Border: []excelize.Border{
			{Type: "top", Color: "D9D9D9", Style: 1},
			{Type: "bottom", Color: "D9D9D9", Style: 1},
			{Type: "left", Color: "D9D9D9", Style: 1},
			{Type: "right", Color: "D9D9D9", Style: 1},
		},
	})

	presentStyle, _ := f.NewStyle(&excelize.Style{
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"C6EFCE"}, Pattern: 1},
		Font:      &excelize.Font{Color: "006100", Size: 9, Bold: true},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
		Border: []excelize.Border{
			{Type: "top", Color: "D9D9D9", Style: 1},
			{Type: "bottom", Color: "D9D9D9", Style: 1},
			{Type: "left", Color: "D9D9D9", Style: 1},
			{Type: "right", Color: "D9D9D9", Style: 1},
		},
	})

	absentStyle, _ := f.NewStyle(&excelize.Style{
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"FFC7CE"}, Pattern: 1},
		Font:      &excelize.Font{Color: "9C0006", Size: 9, Bold: true},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
		Border: []excelize.Border{
			{Type: "top", Color: "D9D9D9", Style: 1},
			{Type: "bottom", Color: "D9D9D9", Style: 1},
			{Type: "left", Color: "D9D9D9", Style: 1},
			{Type: "right", Color: "D9D9D9", Style: 1},
		},
	})

	leaveStyle, _ := f.NewStyle(&excelize.Style{
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"FFEB9C"}, Pattern: 1},
		Font:      &excelize.Font{Color: "9C5700", Size: 9, Bold: true},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
		Border: []excelize.Border{
			{Type: "top", Color: "D9D9D9", Style: 1},
			{Type: "bottom", Color: "D9D9D9", Style: 1},
			{Type: "left", Color: "D9D9D9", Style: 1},
			{Type: "right", Color: "D9D9D9", Style: 1},
		},
	})

	sickStyle, _ := f.NewStyle(&excelize.Style{
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"DCE6F1"}, Pattern: 1},
		Font:      &excelize.Font{Color: "203764", Size: 9, Bold: true},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
		Border: []excelize.Border{
			{Type: "top", Color: "D9D9D9", Style: 1},
			{Type: "bottom", Color: "D9D9D9", Style: 1},
			{Type: "left", Color: "D9D9D9", Style: 1},
			{Type: "right", Color: "D9D9D9", Style: 1},
		},
	})

	halfDayStyle, _ := f.NewStyle(&excelize.Style{
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"E4DFEC"}, Pattern: 1},
		Font:      &excelize.Font{Color: "5C4284", Size: 9, Bold: true},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
		Border: []excelize.Border{
			{Type: "top", Color: "D9D9D9", Style: 1},
			{Type: "bottom", Color: "D9D9D9", Style: 1},
			{Type: "left", Color: "D9D9D9", Style: 1},
			{Type: "right", Color: "D9D9D9", Style: 1},
		},
	})

	weekendStyle, _ := f.NewStyle(&excelize.Style{
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"F2F2F2"}, Pattern: 1},
		Font:      &excelize.Font{Color: "7F7F7F", Size: 9, Bold: true},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
		Border: []excelize.Border{
			{Type: "top", Color: "D9D9D9", Style: 1},
			{Type: "bottom", Color: "D9D9D9", Style: 1},
			{Type: "left", Color: "D9D9D9", Style: 1},
			{Type: "right", Color: "D9D9D9", Style: 1},
		},
	})

	businessTripStyle, _ := f.NewStyle(&excelize.Style{
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"D9E1F2"}, Pattern: 1}, // Soft Light Blue
		Font:      &excelize.Font{Color: "1B365D", Size: 9, Bold: true},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
		Border: []excelize.Border{
			{Type: "top", Color: "D9D9D9", Style: 1},
			{Type: "bottom", Color: "D9D9D9", Style: 1},
			{Type: "left", Color: "D9D9D9", Style: 1},
			{Type: "right", Color: "D9D9D9", Style: 1},
		},
	})

	// Add Title & Subtitle
	f.SetCellValue(sheetName, "A1", "LAPORAN KEHADIRAN KARYAWAN")
	f.SetCellStyle(sheetName, "A1", "A1", titleStyle)

	subStr := fmt.Sprintf("Periode: %s", month)
	if len(employees) > 0 && employees[0].Branch != nil {
		subStr += fmt.Sprintf(" | Cabang: %s", employees[0].Branch.Name)
	}
	f.SetCellValue(sheetName, "A2", subStr)
	f.SetCellStyle(sheetName, "A2", "A2", subtitleStyle)

	// Set Row Heights
	f.SetRowHeight(sheetName, 1, 24)
	f.SetRowHeight(sheetName, 2, 18)
	f.SetRowHeight(sheetName, 4, 30) // Header Row

	// Table Headers
	f.SetCellValue(sheetName, "A4", "TANGGAL")
	f.SetCellStyle(sheetName, "A4", "A4", headerStyle)
	f.SetColWidth(sheetName, "A", "A", 22)

	// Employee Name Headers starting from column B
	for i, emp := range employees {
		colNum := i + 2
		colName, _ := excelize.ColumnNumberToName(colNum)
		fullName := fmt.Sprintf("%s %s", emp.FirstName, emp.LastName)
		if emp.LastName == "" {
			fullName = emp.FirstName
		}
		f.SetCellValue(sheetName, colName+"4", fullName)
		f.SetCellStyle(sheetName, colName+"4", colName+"4", headerStyle)
		f.SetColWidth(sheetName, colName, colName, 20)
	}

	// Translate day names to Indonesian
	indoDays := map[string]string{
		"Sunday":    "Minggu",
		"Monday":    "Senin",
		"Tuesday":   "Selasa",
		"Wednesday": "Rabu",
		"Thursday":  "Kamis",
		"Friday":    "Jumat",
		"Saturday":  "Sabtu",
	}

	// Write Date Rows
	startRow := 5
	for rIdx, d := range dates {
		rowNum := startRow + rIdx
		f.SetRowHeight(sheetName, rowNum, 45) // Tall enough for 3 lines

		dayEng := d.Format("Monday")
		dayIndo := indoDays[dayEng]
		dateStr := fmt.Sprintf("%s\n%s", d.Format("02/01/2006"), dayIndo)

		f.SetCellValue(sheetName, fmt.Sprintf("A%d", rowNum), dateStr)
		f.SetCellStyle(sheetName, fmt.Sprintf("A%d", rowNum), fmt.Sprintf("A%d", rowNum), dateStyle)

		dKey := d.Format("2006-01-02")
		isWeekend := d.Weekday() == time.Sunday

		// Write data for each employee
		for cIdx, emp := range employees {
			colNum := cIdx + 2
			colName, _ := excelize.ColumnNumberToName(colNum)
			cellRef := fmt.Sprintf("%s%d", colName, rowNum)

			lKey := LogKey{EmployeeID: emp.ID, DateStr: dKey}
			lvKey := LeaveKey{EmployeeID: emp.ID, DateStr: dKey}

			if log, ok := logMap[lKey]; ok && (log.IsOutOfTown || log.Status == "BUSINESS_TRIP") {
				// Dinas Luar Kota (Soft Light Blue)
				inTime := "-"
				outTime := "-"
				totTime := "-"
				if log.ClockInTime != nil {
					localIn := convertToJakarta(log.ClockInTime)
					inTime = localIn.Format("15:04:05")
					if log.ClockOutTime != nil {
						localOut := convertToJakarta(log.ClockOutTime)
						outTime = localOut.Format("15:04:05")
						dur := localOut.Sub(localIn)
						h := int(dur.Hours())
						m := int(dur.Minutes()) % 60
						totTime = fmt.Sprintf("%dj %dm", h, m)
					}
				}
				dest := log.Destination
				if dest == "" && log.Notes != "" {
					dest = log.Notes
				}
				if dest == "" {
					dest = "Luar Kota"
				}
				cellVal := fmt.Sprintf("DINAS LUAR KOTA\n(%s)\nMasuk: %s | Keluar: %s\nTotal: %s", dest, inTime, outTime, totTime)
				f.SetCellValue(sheetName, cellRef, cellVal)
				f.SetCellStyle(sheetName, cellRef, cellRef, businessTripStyle)
			} else if log, ok := logMap[lKey]; ok && (log.Status == "ON_TIME" || log.Status == "PRESENT" || log.Status == "LATE") {
				// Hadir (Green)
				inTime := "-"
				outTime := "-"
				totTime := "-"
				if log.ClockInTime != nil {
					localIn := convertToJakarta(log.ClockInTime)
					inTime = localIn.Format("15:04:05")
					if log.ClockOutTime != nil {
						localOut := convertToJakarta(log.ClockOutTime)
						outTime = localOut.Format("15:04:05")
						dur := localOut.Sub(localIn)
						h := int(dur.Hours())
						m := int(dur.Minutes()) % 60
						totTime = fmt.Sprintf("%dj %dm", h, m)
					}
				}
				// Cek apakah ada izin/cuti/sakit setengah hari di hari yang sama
				if lvType, hasLeave := leaveMap[lvKey]; hasLeave {
					lvTypeLower := strings.ToLower(lvType)
					labelPrefix := "IZIN SETENGAH HARI"
					if strings.Contains(lvTypeLower, "cuti") {
						labelPrefix = "CUTI SETENGAH HARI"
					} else if strings.Contains(lvTypeLower, "sakit") {
						labelPrefix = "SAKIT SETENGAH HARI"
					}
					cellVal := fmt.Sprintf("Masuk: %s\nKeluar: %s\nTotal: %s\n---\n%s\n(%s)", inTime, outTime, totTime, labelPrefix, lvType)
					f.SetCellValue(sheetName, cellRef, cellVal)
					f.SetCellStyle(sheetName, cellRef, cellRef, halfDayStyle)
				} else {
					cellVal := fmt.Sprintf("Masuk: %s\nKeluar: %s\nTotal: %s", inTime, outTime, totTime)
					f.SetCellValue(sheetName, cellRef, cellVal)
					f.SetCellStyle(sheetName, cellRef, cellRef, presentStyle)
				}
			} else if lvType, ok := leaveMap[lvKey]; ok {
				// Cuti/Izin/Sakit (Full Day or Half Day without clock-in)
				lvTypeLower := strings.ToLower(lvType)
				if strings.Contains(lvTypeLower, "1/2") || strings.Contains(lvTypeLower, "setengah") || strings.Contains(lvTypeLower, "half") {
					labelPrefix := "IZIN SETENGAH HARI"
					if strings.Contains(lvTypeLower, "cuti") {
						labelPrefix = "CUTI SETENGAH HARI"
					} else if strings.Contains(lvTypeLower, "sakit") {
						labelPrefix = "SAKIT SETENGAH HARI"
					}
					f.SetCellValue(sheetName, cellRef, fmt.Sprintf("%s\n(%s)", labelPrefix, lvType))
					f.SetCellStyle(sheetName, cellRef, cellRef, halfDayStyle)
				} else if strings.Contains(lvTypeLower, "sakit") {
					f.SetCellValue(sheetName, cellRef, fmt.Sprintf("SAKIT\n(%s)", lvType))
					f.SetCellStyle(sheetName, cellRef, cellRef, sickStyle)
				} else if strings.Contains(lvTypeLower, "cuti") {
					f.SetCellValue(sheetName, cellRef, fmt.Sprintf("CUTI\n(%s)", lvType))
					f.SetCellStyle(sheetName, cellRef, cellRef, leaveStyle)
				} else {
					f.SetCellValue(sheetName, cellRef, fmt.Sprintf("IZIN\n(%s)", lvType))
					f.SetCellStyle(sheetName, cellRef, cellRef, leaveStyle)
				}
			} else if log, ok := logMap[lKey]; ok && (log.Status == "ABSENT" || log.Status == "ALFA") {
				// Absent/Alfa explicitly (Red)
				f.SetCellValue(sheetName, cellRef, "TIDAK ABSEN\n(MANGKIR)")
				f.SetCellStyle(sheetName, cellRef, cellRef, absentStyle)
			} else if isWeekend {
				// Weekend Off Day (Gray)
				f.SetCellValue(sheetName, cellRef, "HARI LIBUR\n(OFF)")
				f.SetCellStyle(sheetName, cellRef, cellRef, weekendStyle)
			} else {
				// Unexcused missing workday (Red)
				f.SetCellValue(sheetName, cellRef, "TIDAK ABSEN\n(ALFA)")
				f.SetCellStyle(sheetName, cellRef, cellRef, absentStyle)
			}

			if log, ok := logMap[lKey]; ok && log.Notes != "" {
				_ = f.AddComment(sheetName, excelize.Comment{
					Cell:   cellRef,
					Author: "Catatan:\n",
					Paragraph: []excelize.RichTextRun{
						{Text: log.Notes},
					},
				})
			}
		}
	}

	// -------------------------------------------------------------
	// 5.5. Calculate Summaries per Employee
	// -------------------------------------------------------------
	isHalfDayLeave := func(name string) bool {
		lower := strings.ToLower(name)
		return strings.Contains(lower, "1/2") || strings.Contains(lower, "setengah") || strings.Contains(lower, "half")
	}

	type EmployeeSummaryCounters struct {
		Hadir             float64
		Dinas             int
		Telat             int
		Sakit             float64
		Cuti              float64
		Izin              float64
		Alfa              int
		IzinHalfBreakdown map[string]int // Map nama jenis izin setengah hari -> jumlah kali
		IzinFullBreakdown map[string]int // Map nama jenis izin seharian -> jumlah hari
	}

	empSummaryMap := make(map[uuid.UUID]*EmployeeSummaryCounters)
	for _, emp := range employees {
		summary := &EmployeeSummaryCounters{
			IzinFullBreakdown: make(map[string]int),
			IzinHalfBreakdown: make(map[string]int),
		}
		for _, d := range dates {
			dKey := d.Format("2006-01-02")
			lKey := LogKey{EmployeeID: emp.ID, DateStr: dKey}
			lvKey := LeaveKey{EmployeeID: emp.ID, DateStr: dKey}
			isWeekend := d.Weekday() == time.Sunday

			log, hasLog := logMap[lKey]
			lvType, hasLeave := leaveMap[lvKey]
			isLeaveHalfDay := hasLeave && isHalfDayLeave(lvType)

			isHadirOrTelat := hasLog && (log.Status == "ON_TIME" || log.Status == "PRESENT" || log.Status == "LATE")

			if hasLog && (log.IsOutOfTown || log.Status == "BUSINESS_TRIP") {
				summary.Dinas++
				summary.Hadir += 1.0
				if log.Status == "LATE" {
					summary.Telat++
				}
			} else if isHadirOrTelat {
				if hasLeave {
					// Karyawan hadir DAN memiliki pengajuan izin/cuti/sakit yang disetujui pada hari yang sama (Setengah Hari)
					summary.Hadir += 0.5
					lvTypeLower := strings.ToLower(lvType)
					typeName := strings.TrimSpace(lvType)
					if strings.Contains(lvTypeLower, "sakit") {
						summary.Sakit += 0.5
					} else if strings.Contains(lvTypeLower, "cuti") {
						summary.Cuti += 0.5
					} else {
						summary.Izin += 0.5
						if typeName == "" {
							typeName = "Izin Setengah Hari"
						}
						summary.IzinHalfBreakdown[typeName]++
					}
				} else {
					summary.Hadir += 1.0
				}
				if log.Status == "LATE" {
					summary.Telat++
				}
			} else if hasLeave {
				lvTypeLower := strings.ToLower(lvType)
				typeName := strings.TrimSpace(lvType)
				if isLeaveHalfDay {
					// Pengajuan izin/cuti/sakit setengah hari secara eksplisit (bobot 0.5)
					if strings.Contains(lvTypeLower, "sakit") {
						summary.Sakit += 0.5
					} else if strings.Contains(lvTypeLower, "cuti") {
						summary.Cuti += 0.5
					} else {
						summary.Izin += 0.5
						if typeName == "" {
							typeName = "Izin Setengah Hari"
						}
						summary.IzinHalfBreakdown[typeName]++
					}
				} else {
					// Cuti / Sakit / Izin seharian penuh (bobot 1.0)
					if strings.Contains(lvTypeLower, "sakit") {
						summary.Sakit += 1.0
					} else if strings.Contains(lvTypeLower, "cuti") {
						summary.Cuti += 1.0
					} else {
						summary.Izin += 1.0
						if typeName == "" {
							typeName = "Izin Lainnya"
						}
						summary.IzinFullBreakdown[typeName]++
					}
				}
			} else if hasLog && (log.Status == "ABSENT" || log.Status == "ALFA") {
				summary.Alfa++
			} else if !isWeekend {
				summary.Alfa++
			}
		}
		empSummaryMap[emp.ID] = summary
	}

	// Helper to format integer vs float counts nicely in Excel
	formatCount := func(v float64) interface{} {
		if v == math.Floor(v) {
			return int(v)
		}
		return v
	}

	// Helper to build detailed Izin breakdown note / comment for Excel
	buildIzinComment := func(s *EmployeeSummaryCounters) string {
		if s == nil || s.Izin <= 0 {
			return ""
		}
		var lines []string
		lines = append(lines, "Rincian Izin:")
		for name, count := range s.IzinFullBreakdown {
			lines = append(lines, fmt.Sprintf("• %s: %d hari", name, count))
		}
		for name, count := range s.IzinHalfBreakdown {
			halfDays := float64(count) * 0.5
			lines = append(lines, fmt.Sprintf("• %s: %d kali (%v hari)", name, count, formatCount(halfDays)))
		}
		lines = append(lines, fmt.Sprintf("Total: %v hari", formatCount(s.Izin)))
		return strings.Join(lines, "\n")
	}

	// -------------------------------------------------------------
	// 5.6. Add Summary Block at the Bottom of "Laporan Kehadiran"
	// -------------------------------------------------------------
	summaryHeaderStyle, _ := f.NewStyle(&excelize.Style{
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"1E293B"}, Pattern: 1}, // Slate-800
		Font:      &excelize.Font{Bold: true, Color: "FFFFFF", Size: 11},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "top", Color: "CBD5E1", Style: 1},
			{Type: "bottom", Color: "CBD5E1", Style: 1},
			{Type: "left", Color: "CBD5E1", Style: 1},
			{Type: "right", Color: "CBD5E1", Style: 1},
		},
	})

	summaryLabelStyle, _ := f.NewStyle(&excelize.Style{
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"F1F5F9"}, Pattern: 1}, // Slate-100
		Font:      &excelize.Font{Bold: true, Color: "1E293B", Size: 10},
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center", Indent: 1},
		Border: []excelize.Border{
			{Type: "top", Color: "CBD5E1", Style: 1},
			{Type: "bottom", Color: "CBD5E1", Style: 1},
			{Type: "left", Color: "CBD5E1", Style: 1},
			{Type: "right", Color: "CBD5E1", Style: 1},
		},
	})

	summaryValPresentStyle, _ := f.NewStyle(&excelize.Style{
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"DCFCE7"}, Pattern: 1}, // Green
		Font:      &excelize.Font{Bold: true, Color: "166534", Size: 11},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "top", Color: "CBD5E1", Style: 1},
			{Type: "bottom", Color: "CBD5E1", Style: 1},
			{Type: "left", Color: "CBD5E1", Style: 1},
			{Type: "right", Color: "CBD5E1", Style: 1},
		},
	})

	summaryValTripStyle, _ := f.NewStyle(&excelize.Style{
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"E0E7FF"}, Pattern: 1}, // Indigo
		Font:      &excelize.Font{Bold: true, Color: "3730A3", Size: 11},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "top", Color: "CBD5E1", Style: 1},
			{Type: "bottom", Color: "CBD5E1", Style: 1},
			{Type: "left", Color: "CBD5E1", Style: 1},
			{Type: "right", Color: "CBD5E1", Style: 1},
		},
	})

	summaryValLateStyle, _ := f.NewStyle(&excelize.Style{
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"FFEDD5"}, Pattern: 1}, // Orange
		Font:      &excelize.Font{Bold: true, Color: "C2410C", Size: 11},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "top", Color: "CBD5E1", Style: 1},
			{Type: "bottom", Color: "CBD5E1", Style: 1},
			{Type: "left", Color: "CBD5E1", Style: 1},
			{Type: "right", Color: "CBD5E1", Style: 1},
		},
	})

	summaryValSickStyle, _ := f.NewStyle(&excelize.Style{
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"E2E8F0"}, Pattern: 1}, // Blue-gray
		Font:      &excelize.Font{Bold: true, Color: "334155", Size: 11},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "top", Color: "CBD5E1", Style: 1},
			{Type: "bottom", Color: "CBD5E1", Style: 1},
			{Type: "left", Color: "CBD5E1", Style: 1},
			{Type: "right", Color: "CBD5E1", Style: 1},
		},
	})

	summaryValLeaveStyle, _ := f.NewStyle(&excelize.Style{
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"FEF3C7"}, Pattern: 1}, // Yellow/Amber
		Font:      &excelize.Font{Bold: true, Color: "92400E", Size: 11},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "top", Color: "CBD5E1", Style: 1},
			{Type: "bottom", Color: "CBD5E1", Style: 1},
			{Type: "left", Color: "CBD5E1", Style: 1},
			{Type: "right", Color: "CBD5E1", Style: 1},
		},
	})

	summaryValAbsentStyle, _ := f.NewStyle(&excelize.Style{
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"FEE2E2"}, Pattern: 1}, // Red
		Font:      &excelize.Font{Bold: true, Color: "991B1B", Size: 11},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "top", Color: "CBD5E1", Style: 1},
			{Type: "bottom", Color: "CBD5E1", Style: 1},
			{Type: "left", Color: "CBD5E1", Style: 1},
			{Type: "right", Color: "CBD5E1", Style: 1},
		},
	})

	summaryValPaidStyle, _ := f.NewStyle(&excelize.Style{
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"E0F2FE"}, Pattern: 1}, // Sky-100
		Font:      &excelize.Font{Bold: true, Color: "0369A1", Size: 11},                 // Sky-700
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "top", Color: "CBD5E1", Style: 1},
			{Type: "bottom", Color: "CBD5E1", Style: 1},
			{Type: "left", Color: "CBD5E1", Style: 1},
			{Type: "right", Color: "CBD5E1", Style: 1},
		},
	})

	curRow := startRow + len(dates) + 1 // +1 empty gap row
	f.SetRowHeight(sheetName, curRow, 28)
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", curRow), "REKAPAN KEHADIRAN")
	f.SetCellStyle(sheetName, fmt.Sprintf("A%d", curRow), fmt.Sprintf("A%d", curRow), summaryHeaderStyle)
	for cIdx := range employees {
		colNum := cIdx + 2
		colName, _ := excelize.ColumnNumberToName(colNum)
		cellRef := fmt.Sprintf("%s%d", colName, curRow)
		f.SetCellValue(sheetName, cellRef, "TOTAL HARI")
		f.SetCellStyle(sheetName, cellRef, cellRef, summaryHeaderStyle)
	}

	type SummaryMetric struct {
		Label string
		Field func(s *EmployeeSummaryCounters) float64
		Style int
	}

	metrics := []SummaryMetric{
		{Label: "TOTAL MASUK (HADIR)", Field: func(s *EmployeeSummaryCounters) float64 { return s.Hadir }, Style: summaryValPresentStyle},
		{Label: "DINAS LUAR KOTA", Field: func(s *EmployeeSummaryCounters) float64 { return float64(s.Dinas) }, Style: summaryValTripStyle},
		{Label: "TERLAMBAT", Field: func(s *EmployeeSummaryCounters) float64 { return float64(s.Telat) }, Style: summaryValLateStyle},
		{Label: "SAKIT", Field: func(s *EmployeeSummaryCounters) float64 { return s.Sakit }, Style: summaryValSickStyle},
		{Label: "CUTI", Field: func(s *EmployeeSummaryCounters) float64 { return s.Cuti }, Style: summaryValLeaveStyle},
		{Label: "IZIN", Field: func(s *EmployeeSummaryCounters) float64 { return s.Izin }, Style: summaryValLeaveStyle},
		{Label: "TOTAL HARI DIBAYAR (HADIR + CUTI + SAKIT)", Field: func(s *EmployeeSummaryCounters) float64 { return s.Hadir + s.Cuti + s.Sakit }, Style: summaryValPaidStyle},
		{Label: "ALFA / TIDAK ABSEN", Field: func(s *EmployeeSummaryCounters) float64 { return float64(s.Alfa) }, Style: summaryValAbsentStyle},
	}

	for _, m := range metrics {
		curRow++
		f.SetRowHeight(sheetName, curRow, 22)
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", curRow), m.Label)
		f.SetCellStyle(sheetName, fmt.Sprintf("A%d", curRow), fmt.Sprintf("A%d", curRow), summaryLabelStyle)

		for cIdx, emp := range employees {
			colNum := cIdx + 2
			colName, _ := excelize.ColumnNumberToName(colNum)
			cellRef := fmt.Sprintf("%s%d", colName, curRow)
			sum := empSummaryMap[emp.ID]
			var val float64
			if sum != nil {
				val = m.Field(sum)
			}
			f.SetCellValue(sheetName, cellRef, formatCount(val))
			f.SetCellStyle(sheetName, cellRef, cellRef, m.Style)

			if m.Label == "IZIN" && sum != nil && sum.Izin > 0 {
				commentText := buildIzinComment(sum)
				if commentText != "" {
					_ = f.AddComment(sheetName, excelize.Comment{
						Cell:   cellRef,
						Author: "Karsa HR:\n",
						Paragraph: []excelize.RichTextRun{
							{Text: commentText},
						},
					})
				}
			}
		}
	}

	// -------------------------------------------------------------
	// 6. Generate Summary Sheet ("Rekapan Absensi")
	// -------------------------------------------------------------
	summarySheet := "Rekapan Absensi"
	f.NewSheet(summarySheet)

	// Set Page Grid lines visible
	f.SetSheetView(summarySheet, 0, &excelize.ViewOptions{
		ShowGridLines: &showGridLines,
	})

	// Write Header for summary sheet
	f.SetCellValue(summarySheet, "A1", "NAMA KARYAWAN")
	f.SetCellValue(summarySheet, "B1", "PERIODE")
	f.SetCellValue(summarySheet, "C1", "HADIR")
	f.SetCellValue(summarySheet, "D1", "DINAS")
	f.SetCellValue(summarySheet, "E1", "TERLAMBAT")
	f.SetCellValue(summarySheet, "F1", "SAKIT")
	f.SetCellValue(summarySheet, "G1", "CUTI")
	f.SetCellValue(summarySheet, "H1", "IZIN")
	f.SetCellValue(summarySheet, "I1", "TOTAL DIBAYAR")
	f.SetCellValue(summarySheet, "J1", "ALFA")

	f.SetCellStyle(summarySheet, "A1", "J1", headerStyle)
	f.SetRowHeight(summarySheet, 1, 28)
	f.SetColWidth(summarySheet, "A", "A", 28)
	f.SetColWidth(summarySheet, "B", "B", 20)
	f.SetColWidth(summarySheet, "C", "H", 14)
	f.SetColWidth(summarySheet, "I", "I", 18)
	f.SetColWidth(summarySheet, "J", "J", 14)

	rowStyle, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Vertical: "center"},
		Border: []excelize.Border{
			{Type: "top", Color: "E2E8F0", Style: 1},
			{Type: "bottom", Color: "E2E8F0", Style: 1},
			{Type: "left", Color: "E2E8F0", Style: 1},
			{Type: "right", Color: "E2E8F0", Style: 1},
		},
	})
	nameStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "1E293B"},
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center", Indent: 1},
		Border: []excelize.Border{
			{Type: "top", Color: "E2E8F0", Style: 1},
			{Type: "bottom", Color: "E2E8F0", Style: 1},
			{Type: "left", Color: "E2E8F0", Style: 1},
			{Type: "right", Color: "E2E8F0", Style: 1},
		},
	})
	centerNumberStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "334155"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "top", Color: "E2E8F0", Style: 1},
			{Type: "bottom", Color: "E2E8F0", Style: 1},
			{Type: "left", Color: "E2E8F0", Style: 1},
			{Type: "right", Color: "E2E8F0", Style: 1},
		},
	})

	sumRow := 2
	for _, emp := range employees {
		fullName := fmt.Sprintf("%s %s", emp.FirstName, emp.LastName)
		if emp.LastName == "" {
			fullName = emp.FirstName
		}

		summary := empSummaryMap[emp.ID]
		if summary == nil {
			summary = &EmployeeSummaryCounters{}
		}

		f.SetRowHeight(summarySheet, sumRow, 22)
		f.SetCellValue(summarySheet, fmt.Sprintf("A%d", sumRow), fullName)
		f.SetCellStyle(summarySheet, fmt.Sprintf("A%d", sumRow), fmt.Sprintf("A%d", sumRow), nameStyle)

		f.SetCellValue(summarySheet, fmt.Sprintf("B%d", sumRow), "Total Bulan Ini")
		f.SetCellStyle(summarySheet, fmt.Sprintf("B%d", sumRow), fmt.Sprintf("B%d", sumRow), rowStyle)

		totalPaid := summary.Hadir + summary.Cuti + summary.Sakit

		f.SetCellValue(summarySheet, fmt.Sprintf("C%d", sumRow), formatCount(summary.Hadir))
		f.SetCellValue(summarySheet, fmt.Sprintf("D%d", sumRow), summary.Dinas)
		f.SetCellValue(summarySheet, fmt.Sprintf("E%d", sumRow), summary.Telat)
		f.SetCellValue(summarySheet, fmt.Sprintf("F%d", sumRow), formatCount(summary.Sakit))
		f.SetCellValue(summarySheet, fmt.Sprintf("G%d", sumRow), formatCount(summary.Cuti))
		f.SetCellValue(summarySheet, fmt.Sprintf("H%d", sumRow), formatCount(summary.Izin))
		f.SetCellValue(summarySheet, fmt.Sprintf("I%d", sumRow), formatCount(totalPaid))
		f.SetCellValue(summarySheet, fmt.Sprintf("J%d", sumRow), summary.Alfa)

		f.SetCellStyle(summarySheet, fmt.Sprintf("C%d", sumRow), fmt.Sprintf("J%d", sumRow), centerNumberStyle)

		if summary.Izin > 0 {
			commentText := buildIzinComment(summary)
			if commentText != "" {
				_ = f.AddComment(summarySheet, excelize.Comment{
					Cell:   fmt.Sprintf("H%d", sumRow),
					Author: "Karsa HR:\n",
					Paragraph: []excelize.RichTextRun{
						{Text: commentText},
					},
				})
			}
		}

		sumRow++
	}

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, fmt.Errorf("failed to generate xlsx buffer: %v", err)
	}

	return buf.Bytes(), nil
}

func (u *attendanceUseCase) CreateManual(req ManualAttendanceRequest) (*domain.AttendanceLog, error) {
	employeeID, err := uuid.Parse(req.EmployeeID)
	if err != nil {
		return nil, errors.New("invalid employee ID")
	}

	clockIn, err := time.Parse(time.RFC3339, req.ClockInTime)
	if err != nil {
		return nil, errors.New("invalid clock_in_time format. Use RFC3339 (e.g. 2026-03-06T08:00:00Z)")
	}

	var clockOut *time.Time
	if req.ClockOutTime != "" {
		parsedOut, err := time.Parse(time.RFC3339, req.ClockOutTime)
		if err != nil {
			return nil, errors.New("invalid clock_out_time format")
		}
		clockOut = &parsedOut
	}

	log := &domain.AttendanceLog{
		EmployeeID:   employeeID,
		ClockInTime:  &clockIn,
		ClockOutTime: clockOut,
		Status:       req.Status,
		Notes:        req.Notes,
		IsOutOfTown:  req.IsOutOfTown,
		Destination:  req.Destination,
	}

	if err := u.repo.Create(log); err != nil {
		return nil, err
	}
	return log, nil
}

func (u *attendanceUseCase) UpdateAttendance(id uuid.UUID, req UpdateAttendanceRequest) error {
	log, err := u.repo.FindByID(id)
	if err != nil {
		return err
	}
	if log == nil {
		return errors.New("attendance log not found")
	}

	if req.ClockInTime != "" {
		clockIn, err := time.Parse(time.RFC3339, req.ClockInTime)
		if err != nil {
			return errors.New("invalid clock_in_time format")
		}
		log.ClockInTime = &clockIn
	}

	if req.ClockOutTime != "" {
		clockOut, err := time.Parse(time.RFC3339, req.ClockOutTime)
		if err != nil {
			return errors.New("invalid clock_out_time format")
		}
		log.ClockOutTime = &clockOut
	} else if req.ClockOutTime == "null" {
		log.ClockOutTime = nil
	}

	if req.Status != "" {
		log.Status = req.Status
	}
	if req.Notes != "" {
		log.Notes = req.Notes
	}
	if req.IsOutOfTown != nil {
		log.IsOutOfTown = *req.IsOutOfTown
	}
	if req.Destination != "" {
		log.Destination = req.Destination
	}

	return u.repo.Update(log)
}

func (u *attendanceUseCase) UpdateMyNotes(userID uuid.UUID, attendanceID uuid.UUID, notes string) error {
	employee, err := u.employeeRepo.FindByUserID(userID)
	if err != nil {
		return err
	}
	if employee == nil {
		return errors.New("employee record not found")
	}

	log, err := u.repo.FindByID(attendanceID)
	if err != nil {
		return err
	}
	if log == nil {
		return errors.New("attendance log not found")
	}

	// Security: Verify that the attendance log belongs to this employee
	if log.EmployeeID != employee.ID {
		return errors.New("unauthorized to update this attendance record")
	}

	log.Notes = strings.TrimSpace(notes)
	return u.repo.Update(log)
}

func (u *attendanceUseCase) DeleteAttendance(id uuid.UUID) error {
	return u.repo.Delete(id)
}

func (u *attendanceUseCase) ProcessDailyAlpha(date time.Time) error {
	// 1. Fetch all active employees
	employees, err := u.employeeRepo.FindAll(0, nil, nil) // Getting all
	if err != nil {
		return err
	}

	for _, emp := range employees {
		_ = u.ProcessEmployeeAlpha(emp.ID, date)
	}

	return nil
}

func (u *attendanceUseCase) ProcessEmployeeAlpha(employeeID uuid.UUID, date time.Time) error {
	// Normalize date to start of day
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	endOfDay := startOfDay.Add(24 * time.Hour).Add(-time.Second)

	// 1. Check if employee has a scheduled shift for this date
	shifts, err := u.employeeShiftRepo.FindByEmployeeIDAndDateRange(employeeID, startOfDay, endOfDay)
	if err != nil || len(shifts) == 0 || shifts[0].IsOffDay {
		return nil // No shift or off day, skip
	}

	// 2. Check if employee already has an attendance log for this date
	logs, err := u.repo.FindHistoryByDateRange(employeeID, startOfDay, endOfDay)
	if err != nil {
		return err
	}
	if len(logs) > 0 {
		return nil // Already has attendance (Poin Present/Late/ALFA)
	}

	// 3. Check if employee has an approved leave for this date
	leaves, err := u.leaveRepo.FindApprovedByEmployeeAndDateRange(employeeID, startOfDay, endOfDay)
	if err != nil {
		return err
	}
	if len(leaves) > 0 {
		return nil // Has approved leave
	}

	// 4. No log and no leave -> Create ALFA log
	alfaLog := &domain.AttendanceLog{
		ID:          uuid.New(),
		EmployeeID:  employeeID,
		ClockInTime: &startOfDay, // Use start of day as reference
		Status:      "ALFA",
		Notes:       "Sistem generated: Tidak absen masuk",
	}

	return u.repo.Create(alfaLog)
}
