package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"example.com/hris/backend/internal/domain"
	"example.com/hris/backend/internal/repository"
)

type AdminCreateJourneyRequest struct {
	EmployeeID    uuid.UUID `json:"employee_id" binding:"required"`
	Date          string    `json:"date" binding:"required"` // YYYY-MM-DD
	StartKM       float64   `json:"start_km"`
	StartTime     string    `json:"start_time"` // HH:mm or RFC3339
	StartLocation string    `json:"start_location"`
	StartPhotoURL string    `json:"start_photo_url"`
	EndKM         *float64  `json:"end_km"`
	EndTime       *string   `json:"end_time"` // HH:mm or RFC3339
	EndLocation   string    `json:"end_location"`
	EndPhotoURL   string    `json:"end_photo_url"`
	Status        string    `json:"status"` // STARTED or ENDED
	BBMAmount     *float64  `json:"bbm_amount"`
	BBMProofURL   string    `json:"bbm_proof_url"`
}

type AdminUpdateJourneyRequest struct {
	Date          string   `json:"date"`
	StartKM       float64  `json:"start_km"`
	StartTime     string   `json:"start_time"`
	StartLocation string   `json:"start_location"`
	StartPhotoURL string   `json:"start_photo_url"`
	EndKM         *float64 `json:"end_km"`
	EndTime       *string  `json:"end_time"`
	EndLocation   string   `json:"end_location"`
	EndPhotoURL   string   `json:"end_photo_url"`
	Status        string   `json:"status"` // STARTED or ENDED
}

type AdminAddBBMRequest struct {
	Amount        float64 `json:"amount" binding:"required"`
	AttachmentURL string  `json:"attachment_url"`
	Description   string  `json:"description"`
}

type SalesDailyJourneyUseCase interface {
	GetTodayJourney(ctx context.Context, userID uuid.UUID) (*domain.SalesDailyJourney, error)
	StartJourney(ctx context.Context, userID uuid.UUID, startKM float64, startPhotoURL, startLocation string) (*domain.SalesDailyJourney, error)
	EndJourney(ctx context.Context, userID uuid.UUID, endKM float64, endPhotoURL, endLocation string) (*domain.SalesDailyJourney, error)
	GetAllJourneys(ctx context.Context, startDate, endDate string, employeeID, branchID *uuid.UUID) ([]map[string]interface{}, error)
	GetMyJourneys(ctx context.Context, userID uuid.UUID, startDate, endDate string) ([]map[string]interface{}, error)
	AdminCreateJourney(ctx context.Context, req AdminCreateJourneyRequest) (*domain.SalesDailyJourney, error)
	AdminUpdateJourney(ctx context.Context, id uuid.UUID, req AdminUpdateJourneyRequest) (*domain.SalesDailyJourney, error)
	AdminDeleteJourney(ctx context.Context, id uuid.UUID) error
	AdminAddBBM(ctx context.Context, journeyID uuid.UUID, req AdminAddBBMRequest) (*domain.Reimbursement, error)
	AdminDeleteBBM(ctx context.Context, reimbursementID uuid.UUID) error
	SelfUpdateJourney(ctx context.Context, userID uuid.UUID, journeyID uuid.UUID, startKM *float64, startPhotoURL, startLocation string, endKM *float64, endPhotoURL, endLocation string) (*domain.SalesDailyJourney, error)
}

type salesDailyJourneyUseCase struct {
	journeyRepo repository.SalesDailyJourneyRepository
	db          *gorm.DB
}

func NewSalesDailyJourneyUseCase(journeyRepo repository.SalesDailyJourneyRepository, db *gorm.DB) SalesDailyJourneyUseCase {
	return &salesDailyJourneyUseCase{
		journeyRepo: journeyRepo,
		db:          db,
	}
}

func (u *salesDailyJourneyUseCase) GetTodayJourney(ctx context.Context, userID uuid.UUID) (*domain.SalesDailyJourney, error) {
	wib := time.FixedZone("WIB", 7*3600)
	todayWIB := time.Now().In(wib).Format("2006-01-02")

	// 1. Check if there is an active started journey
	active, err := u.journeyRepo.GetActiveJourneyByUserID(ctx, userID)
	if err == nil && active != nil {
		activeDate := active.Date.In(wib).Format("2006-01-02")
		// If the active journey was from a past date, auto-close it so it doesn't block daily reset
		if activeDate != todayWIB {
			active.Status = "ENDED"
			if active.EndKM == 0 {
				active.EndKM = active.StartKM
			}
			active.TotalKM = active.EndKM - active.StartKM
			_ = u.journeyRepo.Update(ctx, active)
		} else {
			return active, nil
		}
	}

	// 2. Return today's journey record (STARTED or ENDED)
	return u.journeyRepo.GetTodayJourneyByUserID(ctx, userID)
}

func (u *salesDailyJourneyUseCase) StartJourney(ctx context.Context, userID uuid.UUID, startKM float64, startPhotoURL, startLocation string) (*domain.SalesDailyJourney, error) {
	wib := time.FixedZone("WIB", 7*3600)
	nowWIB := time.Now().In(wib)
	todayWIB := nowWIB.Format("2006-01-02")

	active, _ := u.journeyRepo.GetActiveJourneyByUserID(ctx, userID)
	if active != nil {
		activeDate := active.Date.In(wib).Format("2006-01-02")
		if activeDate != todayWIB {
			// Auto-close stale journey from previous date
			active.Status = "ENDED"
			if active.EndKM == 0 {
				active.EndKM = active.StartKM
			}
			active.TotalKM = active.EndKM - active.StartKM
			_ = u.journeyRepo.Update(ctx, active)
		} else {
			return nil, fmt.Errorf("perjalanan hari ini sudah dimulai pada pukul %s", active.StartTime.In(wib).Format("15:04"))
		}
	}

	existing, err := u.journeyRepo.GetTodayJourneyByUserID(ctx, userID)
	if err == nil && existing != nil {
		return nil, errors.New("perjalanan hari ini sudah tercatat sebelumnya")
	}

	dateWIB := time.Date(nowWIB.Year(), nowWIB.Month(), nowWIB.Day(), 0, 0, 0, 0, time.UTC)

	journey := &domain.SalesDailyJourney{
		UserID:        userID,
		Date:          dateWIB,
		StartKM:       startKM,
		StartPhotoURL: startPhotoURL,
		StartLocation: startLocation,
		StartTime:     nowWIB,
		Status:        "STARTED",
	}

	err = u.journeyRepo.Create(ctx, journey)
	if err != nil {
		return nil, err
	}

	return journey, nil
}

func (u *salesDailyJourneyUseCase) EndJourney(ctx context.Context, userID uuid.UUID, endKM float64, endPhotoURL, endLocation string) (*domain.SalesDailyJourney, error) {
	wib := time.FixedZone("WIB", 7*3600)
	nowWIB := time.Now().In(wib)
	todayWIB := nowWIB.Format("2006-01-02")

	// First search for the active started journey
	journey, err := u.journeyRepo.GetActiveJourneyByUserID(ctx, userID)
	if err != nil || journey == nil {
		// If not active, check if today's journey was already ended
		todayJourney, errToday := u.journeyRepo.GetTodayJourneyByUserID(ctx, userID)
		if errToday == nil && todayJourney != nil && todayJourney.Status == "ENDED" {
			return nil, errors.New("perjalanan hari ini sudah diselesaikan sebelumnya")
		}
		return nil, errors.New("belum ada pencatatan KM Berangkat. Silakan catat KM Berangkat terlebih dahulu")
	}

	journeyDate := journey.Date.In(wib).Format("2006-01-02")
	if journeyDate != todayWIB {
		// Stale journey from past date: close it and inform user
		journey.Status = "ENDED"
		if journey.EndKM == 0 {
			journey.EndKM = journey.StartKM
		}
		journey.TotalKM = journey.EndKM - journey.StartKM
		_ = u.journeyRepo.Update(ctx, journey)
		return nil, errors.New("perjalanan hari sebelumnya sudah otomatis diselesaikan. Silakan catat KM Berangkat baru untuk hari ini")
	}

	if journey.Status == "ENDED" {
		return nil, errors.New("perjalanan hari ini sudah diselesaikan")
	}

	if endKM < journey.StartKM {
		return nil, fmt.Errorf("angka KM Pulang (%.1f) tidak boleh lebih kecil dari KM Berangkat (%.1f)", endKM, journey.StartKM)
	}

	journey.EndKM = endKM
	journey.EndPhotoURL = endPhotoURL
	journey.EndLocation = endLocation
	journey.EndTime = &nowWIB
	journey.TotalKM = endKM - journey.StartKM
	journey.Status = "ENDED"

	err = u.journeyRepo.Update(ctx, journey)
	if err != nil {
		return nil, err
	}

	return journey, nil
}

func (u *salesDailyJourneyUseCase) GetAllJourneys(ctx context.Context, startDate, endDate string, employeeID, branchID *uuid.UUID) ([]map[string]interface{}, error) {
	journeys, err := u.journeyRepo.GetAllJourneys(ctx, startDate, endDate, employeeID, branchID)
	if err != nil {
		return nil, err
	}
	return u.enrichJourneys(ctx, journeys)
}

func (u *salesDailyJourneyUseCase) GetMyJourneys(ctx context.Context, userID uuid.UUID, startDate, endDate string) ([]map[string]interface{}, error) {
	journeys, err := u.journeyRepo.GetJourneysByUserID(ctx, userID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	return u.enrichJourneys(ctx, journeys)
}

func (u *salesDailyJourneyUseCase) enrichJourneys(ctx context.Context, journeys []domain.SalesDailyJourney) ([]map[string]interface{}, error) {
	var results []map[string]interface{}
	for _, j := range journeys {
		dateStr := j.Date.Format("2006-01-02")
		item := map[string]interface{}{
			"id":              j.ID,
			"user_id":         j.UserID,
			"date":            dateStr,
			"start_km":        j.StartKM,
			"start_photo_url": j.StartPhotoURL,
			"start_location":  j.StartLocation,
			"start_time":      j.StartTime,
			"end_km":          j.EndKM,
			"end_photo_url":   j.EndPhotoURL,
			"end_location":    j.EndLocation,
			"end_time":        j.EndTime,
			"total_km":        j.TotalKM,
			"status":          j.Status,
			"created_at":      j.CreatedAt,
			"user":            j.User,
		}

		// Enrich employee info if exists
		var empID *uuid.UUID
		if j.User != nil && j.User.Employee != nil {
			empID = &j.User.Employee.ID
			item["employee"] = j.User.Employee
		}

		// Enrich BBM Reimbursements for this employee on this date
		if empID != nil && u.db != nil {
			var reimbursements []domain.Reimbursement
			startOfDay := j.Date
			endOfDay := j.Date.Add(24*time.Hour - time.Nanosecond)

			u.db.WithContext(ctx).
				Where("employee_id = ? AND created_at >= ? AND created_at <= ?", *empID, startOfDay, endOfDay).
				Find(&reimbursements)

			var totalBBM float64
			var bbmList []map[string]interface{}
			for _, r := range reimbursements {
				totalBBM += r.Amount
				bbmList = append(bbmList, map[string]interface{}{
					"id":             r.ID,
					"title":          r.Title,
					"amount":         r.Amount,
					"status":         r.Status,
					"attachment_url": r.AttachmentURL,
					"description":    r.Description,
				})
			}
			item["bbm_reimbursements"] = bbmList
			item["total_bbm_amount"] = totalBBM

			// Enrich visit counts for this employee on this date
			var visitCount int64
			wib := time.FixedZone("WIB", 7*3600)
			visitStart := time.Date(j.Date.Year(), j.Date.Month(), j.Date.Day(), 0, 0, 0, 0, wib)
			visitEnd := visitStart.Add(24 * time.Hour)
			u.db.WithContext(ctx).Table("sales_visits").
				Where("employee_id = ? AND check_in_time >= ? AND check_in_time < ? AND status != 'SKIPPED' AND type != 'SKIP'", *empID, visitStart, visitEnd).
				Count(&visitCount)
			item["total_visits"] = visitCount
		}

		results = append(results, item)
	}

	if results == nil {
		results = []map[string]interface{}{}
	}

	return results, nil
}

func (u *salesDailyJourneyUseCase) AdminCreateJourney(ctx context.Context, req AdminCreateJourneyRequest) (*domain.SalesDailyJourney, error) {
	// Look up employee and user_id
	var emp domain.Employee
	if err := u.db.WithContext(ctx).First(&emp, "id = ?", req.EmployeeID).Error; err != nil {
		return nil, fmt.Errorf("karyawan tidak ditemukan: %w", err)
	}
	if emp.UserID == uuid.Nil {
		return nil, errors.New("karyawan ini belum memiliki akun pengguna (user account)")
	}

	tDate, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return nil, fmt.Errorf("format tanggal tidak valid (harus YYYY-MM-DD): %w", err)
	}
	dateUTC := time.Date(tDate.Year(), tDate.Month(), tDate.Day(), 0, 0, 0, 0, time.UTC)

	// Determine start time
	wib := time.FixedZone("WIB", 7*3600)
	var startTime time.Time
	if req.StartTime != "" {
		if parsedT, errT := time.Parse("15:04", req.StartTime); errT == nil {
			startTime = time.Date(tDate.Year(), tDate.Month(), tDate.Day(), parsedT.Hour(), parsedT.Minute(), 0, 0, wib)
		} else if parsedRFC, errRFC := time.Parse(time.RFC3339, req.StartTime); errRFC == nil {
			startTime = parsedRFC
		} else {
			startTime = time.Date(tDate.Year(), tDate.Month(), tDate.Day(), 8, 0, 0, 0, wib)
		}
	} else {
		startTime = time.Date(tDate.Year(), tDate.Month(), tDate.Day(), 8, 0, 0, 0, wib)
	}

	status := "STARTED"
	if req.Status != "" {
		status = req.Status
	}

	var endKM float64
	var endTime *time.Time
	var totalKM float64

	if req.EndKM != nil {
		endKM = *req.EndKM
		totalKM = endKM - req.StartKM
		if totalKM < 0 {
			totalKM = 0
		}
	}

	if req.EndTime != nil && *req.EndTime != "" {
		if parsedT, errT := time.Parse("15:04", *req.EndTime); errT == nil {
			et := time.Date(tDate.Year(), tDate.Month(), tDate.Day(), parsedT.Hour(), parsedT.Minute(), 0, 0, wib)
			endTime = &et
		} else if parsedRFC, errRFC := time.Parse(time.RFC3339, *req.EndTime); errRFC == nil {
			endTime = &parsedRFC
		}
	}

	journey := &domain.SalesDailyJourney{
		UserID:        emp.UserID,
		Date:          dateUTC,
		StartKM:       req.StartKM,
		StartPhotoURL: req.StartPhotoURL,
		StartLocation: req.StartLocation,
		StartTime:     startTime,
		EndKM:         endKM,
		EndPhotoURL:   req.EndPhotoURL,
		EndLocation:   req.EndLocation,
		EndTime:       endTime,
		TotalKM:       totalKM,
		Status:        status,
	}

	if err := u.journeyRepo.Create(ctx, journey); err != nil {
		return nil, err
	}

	// Optionally create BBM reimbursement if provided
	if req.BBMAmount != nil && *req.BBMAmount > 0 {
		var totalKM *float64
		var endKMVal *float64
		if req.EndKM != nil {
			diff := *req.EndKM - req.StartKM
			totalKM = &diff
			endKMVal = req.EndKM
		}
		startKMVal := req.StartKM

		reimbursement := &domain.Reimbursement{
			EmployeeID:      emp.ID,
			Category:        "BBM",
			Title:           fmt.Sprintf("BBM Sales - %s", req.Date),
			Amount:          *req.BBMAmount,
			Status:          domain.ReimbursementPending,
			AttachmentURL:   req.BBMProofURL,
			Description:     "Pencatatan BBM Sales Manual oleh Admin",
			StartKM:         &startKMVal,
			EndKM:           endKMVal,
			TotalKM:         totalKM,
			StartKMPhotoURL: req.StartPhotoURL,
			EndKMPhotoURL:   req.EndPhotoURL,
			CreatedAt:       startTime,
			UpdatedAt:       startTime,
		}
		_ = u.db.WithContext(ctx).Create(reimbursement).Error
	}

	return journey, nil
}

func (u *salesDailyJourneyUseCase) AdminUpdateJourney(ctx context.Context, id uuid.UUID, req AdminUpdateJourneyRequest) (*domain.SalesDailyJourney, error) {
	journey, err := u.journeyRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("perjalanan tidak ditemukan: %w", err)
	}

	wib := time.FixedZone("WIB", 7*3600)

	if req.Date != "" {
		if tDate, errT := time.Parse("2006-01-02", req.Date); errT == nil {
			journey.Date = time.Date(tDate.Year(), tDate.Month(), tDate.Day(), 0, 0, 0, 0, time.UTC)
		}
	}

	journey.StartKM = req.StartKM
	if req.StartLocation != "" {
		journey.StartLocation = req.StartLocation
	}
	if req.StartPhotoURL != "" {
		journey.StartPhotoURL = req.StartPhotoURL
	}
	if req.StartTime != "" {
		if parsedT, errT := time.Parse("15:04", req.StartTime); errT == nil {
			journey.StartTime = time.Date(journey.Date.Year(), journey.Date.Month(), journey.Date.Day(), parsedT.Hour(), parsedT.Minute(), 0, 0, wib)
		} else if parsedRFC, errRFC := time.Parse(time.RFC3339, req.StartTime); errRFC == nil {
			journey.StartTime = parsedRFC
		}
	}

	if req.EndKM != nil {
		journey.EndKM = *req.EndKM
		journey.TotalKM = journey.EndKM - journey.StartKM
		if journey.TotalKM < 0 {
			journey.TotalKM = 0
		}
	}
	if req.EndLocation != "" {
		journey.EndLocation = req.EndLocation
	}
	if req.EndPhotoURL != "" {
		journey.EndPhotoURL = req.EndPhotoURL
	}
	if req.EndTime != nil && *req.EndTime != "" {
		if parsedT, errT := time.Parse("15:04", *req.EndTime); errT == nil {
			et := time.Date(journey.Date.Year(), journey.Date.Month(), journey.Date.Day(), parsedT.Hour(), parsedT.Minute(), 0, 0, wib)
			journey.EndTime = &et
		} else if parsedRFC, errRFC := time.Parse(time.RFC3339, *req.EndTime); errRFC == nil {
			journey.EndTime = &parsedRFC
		}
	}
	if req.Status != "" {
		journey.Status = req.Status
	}

	if err := u.journeyRepo.Update(ctx, journey); err != nil {
		return nil, err
	}

	return journey, nil
}

func (u *salesDailyJourneyUseCase) AdminDeleteJourney(ctx context.Context, id uuid.UUID) error {
	return u.journeyRepo.Delete(ctx, id)
}

func (u *salesDailyJourneyUseCase) AdminAddBBM(ctx context.Context, journeyID uuid.UUID, req AdminAddBBMRequest) (*domain.Reimbursement, error) {
	journey, err := u.journeyRepo.GetByID(ctx, journeyID)
	if err != nil {
		return nil, fmt.Errorf("perjalanan tidak ditemukan: %w", err)
	}

	var emp domain.Employee
	if err := u.db.WithContext(ctx).First(&emp, "user_id = ?", journey.UserID).Error; err != nil {
		return nil, fmt.Errorf("data karyawan tidak ditemukan untuk user perjalanan ini: %w", err)
	}

	dateStr := journey.Date.Format("2006-01-02")
	desc := req.Description
	if desc == "" {
		desc = "BBM Sales (Admin Entry)"
	}

	var totalKM *float64
	var endKMVal *float64
	if journey.EndKM > 0 {
		diff := journey.EndKM - journey.StartKM
		totalKM = &diff
		endKM := journey.EndKM
		endKMVal = &endKM
	}
	startKMVal := journey.StartKM

	reimbursement := &domain.Reimbursement{
		EmployeeID:      emp.ID,
		Category:        "BBM",
		Title:           fmt.Sprintf("BBM Sales - %s", dateStr),
		Amount:          req.Amount,
		Status:          domain.ReimbursementPending,
		AttachmentURL:   req.AttachmentURL,
		Description:     desc,
		StartKM:         &startKMVal,
		EndKM:           endKMVal,
		TotalKM:         totalKM,
		StartKMPhotoURL: journey.StartPhotoURL,
		EndKMPhotoURL:   journey.EndPhotoURL,
		CreatedAt:       journey.StartTime,
		UpdatedAt:       journey.StartTime,
	}

	if err := u.db.WithContext(ctx).Create(reimbursement).Error; err != nil {
		return nil, err
	}

	return reimbursement, nil
}

func (u *salesDailyJourneyUseCase) AdminDeleteBBM(ctx context.Context, reimbursementID uuid.UUID) error {
	return u.db.WithContext(ctx).Delete(&domain.Reimbursement{}, "id = ?", reimbursementID).Error
}

func (u *salesDailyJourneyUseCase) SelfUpdateJourney(
	ctx context.Context,
	userID uuid.UUID,
	journeyID uuid.UUID,
	startKM *float64,
	startPhotoURL, startLocation string,
	endKM *float64,
	endPhotoURL, endLocation string,
) (*domain.SalesDailyJourney, error) {
	journey, err := u.journeyRepo.GetByID(ctx, journeyID)
	if err != nil {
		return nil, errors.New("data perjalanan tidak ditemukan")
	}

	// Pastikan journey milik user yang login
	if journey.UserID != userID {
		return nil, errors.New("anda tidak berhak mengubah data perjalanan ini")
	}

	// Update start fields
	if startKM != nil {
		journey.StartKM = *startKM
	}
	if startPhotoURL != "" {
		journey.StartPhotoURL = startPhotoURL
	}
	if startLocation != "" {
		journey.StartLocation = startLocation
	}

	// Update end fields
	if endKM != nil {
		if *endKM < journey.StartKM {
			return nil, fmt.Errorf("angka KM Pulang (%.1f) tidak boleh lebih kecil dari KM Berangkat (%.1f)", *endKM, journey.StartKM)
		}
		journey.EndKM = *endKM
		journey.TotalKM = journey.EndKM - journey.StartKM
		if journey.TotalKM < 0 {
			journey.TotalKM = 0
		}
	} else if startKM != nil && journey.Status == "ENDED" {
		if journey.EndKM < journey.StartKM {
			return nil, fmt.Errorf("angka KM Berangkat baru (%.1f) tidak boleh lebih besar dari KM Pulang (%.1f)", journey.StartKM, journey.EndKM)
		}
		// Recalculate TotalKM if startKM changed while endKM is known
		journey.TotalKM = journey.EndKM - journey.StartKM
		if journey.TotalKM < 0 {
			journey.TotalKM = 0
		}
	}

	if endPhotoURL != "" {
		journey.EndPhotoURL = endPhotoURL
	}
	if endLocation != "" {
		journey.EndLocation = endLocation
	}

	if err := u.journeyRepo.Update(ctx, journey); err != nil {
		return nil, err
	}

	return journey, nil
}


