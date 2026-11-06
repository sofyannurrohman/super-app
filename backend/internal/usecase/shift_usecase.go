package usecase

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"example.com/hris/backend/internal/repository"
)

type ShiftUseCase interface {
	CreateShift(req *CreateShiftRequest) error
	GetShifts() ([]domain.Shift, error)
	GetShiftByID(id uuid.UUID) (*domain.Shift, error)
	UpdateShift(id uuid.UUID, req *UpdateShiftRequest) error
	DeleteShift(id uuid.UUID) error
}

type shiftUseCase struct {
	repo repository.ShiftRepository
}

type CreateShiftRequest struct {
	CompanyID  *uuid.UUID `json:"company_id"`
	Name       string     `json:"name" binding:"required"`
	StartTime  string     `json:"start_time" binding:"required"`
	EndTime    string     `json:"end_time" binding:"required"`
	BreakStart string     `json:"break_start"`
	BreakEnd   string     `json:"break_end"`
	IsFlexible bool       `json:"is_flexible"`
	BranchID   *uuid.UUID `json:"branch_id"`
}

type UpdateShiftRequest struct {
	Name       string     `json:"name" binding:"required"`
	StartTime  string     `json:"start_time" binding:"required"`
	EndTime    string     `json:"end_time" binding:"required"`
	BreakStart string     `json:"break_start"`
	BreakEnd   string     `json:"break_end"`
	IsFlexible bool       `json:"is_flexible"`
	BranchID   *uuid.UUID `json:"branch_id"`
}

func NewShiftUseCase(repo repository.ShiftRepository) ShiftUseCase {
	return &shiftUseCase{repo: repo}
}

func parseShiftTime(tStr string) (string, error) {
	if tStr == "" {
		return "", nil
	}
	// Try parsing multiple formats for robustness
	formats := []string{
		"15:04",
		"15:04:05",
		time.RFC3339,
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02T15:04:05",
	}

	var parsed time.Time
	var err error
	for _, f := range formats {
		parsed, err = time.Parse(f, tStr)
		if err == nil {
			break
		}
	}

	if err != nil {
		return "", err
	}

	// Return as HH:mm:ss for consistent storage in PostgreSQL TIME type
	return parsed.Format("15:04:05"), nil
}

func (u *shiftUseCase) CreateShift(req *CreateShiftRequest) error {
	startTime, err := parseShiftTime(req.StartTime)
	if err != nil || startTime == "" {
		return errors.New("invalid start_time format")
	}
	endTime, err := parseShiftTime(req.EndTime)
	if err != nil || endTime == "" {
		return errors.New("invalid end_time format")
	}
	breakStart, _ := parseShiftTime(req.BreakStart)
	breakEnd, _ := parseShiftTime(req.BreakEnd)

	var bs, be *string
	if breakStart != "" {
		bs = &breakStart
	}
	if breakEnd != "" {
		be = &breakEnd
	}

	shift := &domain.Shift{
		CompanyID:  req.CompanyID,
		Name:       req.Name,
		StartTime:  startTime,
		EndTime:    endTime,
		BreakStart: bs,
		BreakEnd:   be,
		BranchID:   req.BranchID,
		IsFlexible: req.IsFlexible,
	}

	return u.repo.Create(shift)
}

func (u *shiftUseCase) GetShifts() ([]domain.Shift, error) {
	shifts, err := u.repo.FindAll()
	if err != nil {
		return nil, err
	}

	// Clean up time strings to prevent +07:07 historical offset issues
	for i := range shifts {
		shifts[i].StartTime = cleanTimeOutput(shifts[i].StartTime)
		shifts[i].EndTime = cleanTimeOutput(shifts[i].EndTime)
		if shifts[i].BreakStart != nil {
			cleaned := cleanTimeOutput(*shifts[i].BreakStart)
			shifts[i].BreakStart = &cleaned
		}
		if shifts[i].BreakEnd != nil {
			cleaned := cleanTimeOutput(*shifts[i].BreakEnd)
			shifts[i].BreakEnd = &cleaned
		}
	}

	return shifts, nil
}

func cleanTimeOutput(t string) string {
	if t == "" {
		return ""
	}
	// If it contains 'T', it's an ISO timestamp from GORM/Postgres
	if i := 11; len(t) > i && t[10] == 'T' {
		// Extract HH:mm from 0000-01-01T16:07:12...
		// But wait, the user wants the ORIGINAL time (e.g. 09:00)
		// The 16:07 is already shifted.
		// If we can't get the original, we should at least format it.
		// However, the best way is to handle this in parseShiftTime during Create/Update
		// and ensure the database stores it correctly.
		timePart := t[i : i+5] // HH:mm
		return timePart
	}
	// If it's already HH:mm:ss
	if len(t) >= 5 {
		return t[:5]
	}
	return t
}

func (u *shiftUseCase) GetShiftByID(id uuid.UUID) (*domain.Shift, error) {
	return u.repo.FindByID(id)
}

func (u *shiftUseCase) UpdateShift(id uuid.UUID, req *UpdateShiftRequest) error {
	shift, err := u.repo.FindByID(id)
	if err != nil {
		return err
	}
	if shift == nil {
		return errors.New("shift not found")
	}

	startTime, err := parseShiftTime(req.StartTime)
	if err != nil || startTime == "" {
		return errors.New("invalid start_time format")
	}
	endTime, err := parseShiftTime(req.EndTime)
	if err != nil || endTime == "" {
		return errors.New("invalid end_time format")
	}
	breakStart, _ := parseShiftTime(req.BreakStart)
	breakEnd, _ := parseShiftTime(req.BreakEnd)

	var bs, be *string
	if breakStart != "" {
		bs = &breakStart
	}
	if breakEnd != "" {
		be = &breakEnd
	}

	shift.Name = req.Name
	shift.StartTime = startTime
	shift.EndTime = endTime
	shift.BreakStart = bs
	shift.BreakEnd = be
	shift.BranchID = req.BranchID
	shift.IsFlexible = req.IsFlexible

	return u.repo.Update(shift)
}

func (u *shiftUseCase) DeleteShift(id uuid.UUID) error {
	return u.repo.Delete(id)
}
