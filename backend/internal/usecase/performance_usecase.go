package usecase

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"example.com/hris/backend/internal/repository"
)

type PerformanceUseCase interface {
	GetEmployeeKPIs(month, year int) ([]domain.EmployeeKPI, error)
	UpdateEmployeeKPI(id uuid.UUID, productivityScore float64) error
	FinalizeKPI(id uuid.UUID) error
	EvaluateAndAppraise(req EvaluateAppraisalRequest) error
	GetAppraisals(employeeID uuid.UUID) ([]domain.ManagerialAppraisal, error)

	// Employee Self-View
	GetMyPerformance(employeeID uuid.UUID, month, year int) (map[string]interface{}, error)
	GetMyPerformanceHistory(employeeID uuid.UUID) ([]map[string]interface{}, error)
}

type performanceUseCase struct {
	repo              repository.PerformanceRepository
	attendanceRepo    repository.AttendanceRepository
	employeeShiftRepo repository.EmployeeShiftRepository
	leaveRepo         repository.LeaveRepository
	attendanceUC      AttendanceUseCase
}

func NewPerformanceUseCase(
	repo repository.PerformanceRepository,
	attendanceRepo repository.AttendanceRepository,
	employeeShiftRepo repository.EmployeeShiftRepository,
	leaveRepo repository.LeaveRepository,
	attendanceUC AttendanceUseCase,
) PerformanceUseCase {
	return &performanceUseCase{repo, attendanceRepo, employeeShiftRepo, leaveRepo, attendanceUC}
}

func (u *performanceUseCase) GetEmployeeKPIs(month, year int) ([]domain.EmployeeKPI, error) {
	// First, fetch existing KPIs
	existingKPIs, err := u.repo.GetEmployeeKPIs(month, year)
	if err != nil {
		return nil, err
	}

	// Map existing by EmployeeID
	kpiMap := make(map[uuid.UUID]bool)
	for _, kpi := range existingKPIs {
		kpiMap[kpi.EmployeeID] = true
	}

	// Fetch all regular employees
	employees, err := u.repo.GetRegularEmployees()
	if err != nil {
		return nil, err
	}

	// Auto-generate missing KPIs
	var finalKPIs []domain.EmployeeKPI
	finalKPIs = append(finalKPIs, existingKPIs...)

	for _, emp := range employees {
		// 1. Ensure ALFA logs exist for current month up to yesterday
		now := time.Now()
		start := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
		end := start.AddDate(0, 1, 0).Add(-time.Second)

		// If we are in the same month, only process up to yesterday
		processEnd := end
		if now.Year() == year && int(now.Month()) == month {
			yesterday := now.AddDate(0, 0, -1)
			processEnd = time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 23, 59, 59, 0, time.UTC)
		}

		// Sync ALFAs for the employee (optimizable by doing batch, but this is safe for MVP)
		for d := start; d.Before(processEnd); d = d.AddDate(0, 0, 1) {
			// Small optimization: we only really need to check days that have passed
			_ = u.attendanceUC.ProcessDailyAlpha(d)
		}

		// 2. Fetch all logs and shifts for the month
		logs, _ := u.attendanceRepo.FindHistoryByDateRange(emp.ID, start, end)
		shifts, _ := u.employeeShiftRepo.FindByEmployeeIDAndDateRange(emp.ID, start, end)

		onTimeCount := 0
		lateCount := 0
		alphaCount := 0
		permitCount := 0

		// Use shifts as the base for "scheduled days"
		for _, shift := range shifts {
			if shift.IsOffDay {
				continue
			}

			// Find log for this specific date
			var dayLog *domain.AttendanceLog
			for _, log := range logs {
				if log.ClockInTime != nil && log.ClockInTime.Format("2006-01-02") == shift.Date.Format("2006-01-02") {
					dayLog = &log
					break
				}
			}

			if dayLog != nil {
				switch dayLog.Status {
				case "ON_TIME":
					onTimeCount++
				case "LATE":
					lateCount++
				case "ALFA":
					alphaCount++
				default:
					onTimeCount++ // Fallback
				}
			} else {
				// If no log, check for approved leave
				leaves, _ := u.leaveRepo.FindApprovedByEmployeeAndDateRange(emp.ID, shift.Date, shift.Date.Add(24*time.Hour))
				if len(leaves) > 0 {
					permitCount++
				} else {
					alphaCount++ // Should have been caught by ProcessDailyAlpha, but safeguard
				}
			}
		}

		// 3. Calculate Score
		// Weights: ON_TIME = 1.0, LATE = 0.8, ALPHA = 0.0, PERMIT = 1.0
		totalScheduledDays := onTimeCount + lateCount + alphaCount + permitCount
		var attnScore float64 = 0
		if totalScheduledDays > 0 {
			points := (float64(onTimeCount) * 1.0) + (float64(lateCount) * 0.8) + (float64(permitCount) * 1.0)
			attnScore = (points / float64(totalScheduledDays)) * 100
		}

		if !kpiMap[emp.ID] {
			// Create new entry
			newKPI := domain.EmployeeKPI{
				ID:                uuid.New(),
				EmployeeID:        emp.ID,
				AttendanceScore:   attnScore,
				OnTimeCount:       onTimeCount,
				LateCount:         lateCount,
				AlphaCount:        alphaCount,
				PermitCount:       permitCount,
				ProductivityScore: 0,
				FinalScore:        attnScore * 0.5,
				Status:            "DRAFT",
				PeriodMonth:       month,
				PeriodYear:        year,
				Employee:          &emp,
			}
			err = u.repo.SaveEmployeeKPI(&newKPI)
			if err == nil {
				finalKPIs = append(finalKPIs, newKPI)
			}
		} else {
			// Update existing DRAFT KPI
			for i := range finalKPIs {
				if finalKPIs[i].EmployeeID == emp.ID && finalKPIs[i].Status == "DRAFT" {
					finalKPIs[i].AttendanceScore = attnScore
					finalKPIs[i].OnTimeCount = onTimeCount
					finalKPIs[i].LateCount = lateCount
					finalKPIs[i].AlphaCount = alphaCount
					finalKPIs[i].PermitCount = permitCount
					finalKPIs[i].FinalScore = (attnScore * 0.5) + (finalKPIs[i].ProductivityScore * 0.5)
					u.repo.SaveEmployeeKPI(&finalKPIs[i])
				}
			}
		}
	}

	return finalKPIs, nil
}

func (u *performanceUseCase) UpdateEmployeeKPI(id uuid.UUID, productivityScore float64) error {
	kpi, err := u.repo.GetEmployeeKPIByID(id)
	if err != nil {
		return err
	}
	if kpi == nil {
		return fmt.Errorf("KPI not found")
	}

	if kpi.Status == "FINALIZED" {
		return fmt.Errorf("cannot update finalized KPI")
	}

	kpi.ProductivityScore = productivityScore
	kpi.FinalScore = (kpi.AttendanceScore * 0.5) + (productivityScore * 0.5)

	return u.repo.SaveEmployeeKPI(kpi)
}

func (u *performanceUseCase) FinalizeKPI(id uuid.UUID) error {
	kpi, err := u.repo.GetEmployeeKPIByID(id)
	if err != nil {
		return err
	}
	if kpi == nil {
		return fmt.Errorf("KPI not found")
	}

	kpi.Status = "FINALIZED"
	return u.repo.SaveEmployeeKPI(kpi)
}

type EvaluateAppraisalRequest struct {
	EmployeeID uuid.UUID `json:"employee_id" binding:"required"`
	ManagerID  uuid.UUID `json:"manager_id" binding:"required"`
	Notes      string    `json:"notes" binding:"required"`
	Rating     float64   `json:"rating" binding:"required"`
}

func (u *performanceUseCase) EvaluateAndAppraise(req EvaluateAppraisalRequest) error {
	appraisal := &domain.ManagerialAppraisal{
		ID:          uuid.New(),
		EmployeeID:  req.EmployeeID,
		ManagerID:   req.ManagerID,
		ReviewNotes: req.Notes,
		Rating:      req.Rating,
		ReviewDate:  time.Now(),
	}

	return u.repo.SaveAppraisal(appraisal)
}

func (u *performanceUseCase) GetAppraisals(employeeID uuid.UUID) ([]domain.ManagerialAppraisal, error) {
	return u.repo.GetAppraisalsByEmployee(employeeID)
}

func (u *performanceUseCase) GetMyPerformance(employeeID uuid.UUID, month, year int) (map[string]interface{}, error) {
	// 1. Try to get EmployeeKPI
	regKPI, _ := u.repo.GetEmployeeKPIByEmployeeAndPeriod(employeeID, month, year)

	// 2. Try to get SalesKPI
	salesKPI, _ := u.repo.GetSalesKPIByEmployeeAndPeriod(employeeID, month, year)

	if salesKPI == nil || (salesKPI != nil && salesKPI.TargetOmzet == 0) {
		history, _ := u.repo.GetSalesKPIHistory(employeeID)
		if len(history) > 0 {
			for _, h := range history {
				if h.TargetOmzet > 0 {
					if salesKPI == nil {
						salesKPI = &domain.SalesKPI{
							EmployeeID:    employeeID,
							PeriodMonth:   month,
							PeriodYear:    year,
							AchievedOmzet: 0,
						}
					}
					salesKPI.TargetOmzet = h.TargetOmzet
					salesKPI.TargetNewStores = h.TargetNewStores
					salesKPI.TargetDailyVisits = h.TargetDailyVisits
					salesKPI.TargetDailyNewStores = h.TargetDailyNewStores
					salesKPI.Commission = h.Commission
					break
				}
			}
		}
	}

	// 3. Compose response
	res := map[string]interface{}{
		"period": fmt.Sprintf("%04d-%02d", year, month),
		"status": "DRAFT",
	}

	if salesKPI != nil {
		res["type"] = "SALES"
		var achievementPct float64
		if salesKPI.TargetOmzet > 0 {
			achievementPct = (salesKPI.AchievedOmzet / salesKPI.TargetOmzet) * 100
		}
		res["achievement_percentage"] = achievementPct
		res["final_score"] = achievementPct
		res["bonus"] = salesKPI.EstimatedBonus
		res["commission"] = salesKPI.Commission
		res["items"] = []map[string]interface{}{
			{"label": "Omzet Target", "score": salesKPI.TargetOmzet, "weight": "100%"},
			{"label": "Pencapaian", "score": salesKPI.AchievedOmzet, "weight": ""},
		}
	} else if regKPI != nil {
		res["type"] = "INTERNAL"
		res["final_score"] = regKPI.FinalScore
		res["items"] = []map[string]interface{}{
			{"label": "Kehadiran", "score": regKPI.AttendanceScore, "weight": "50%"},
			{"label": "Produktivitas", "score": regKPI.ProductivityScore, "weight": "50%"},
		}
	}

	todayUnvalidated, todayValidated, _ := u.repo.GetTodayOmzetByEmployee(employeeID)
	res["today_omzet_unvalidated"] = todayUnvalidated
	res["today_omzet_validated"] = todayValidated

	todayTotalOrder, _ := u.repo.GetTodayTotalOrderValue(employeeID)
	res["today_omzet_total"] = todayTotalOrder
	res["today_order_total"] = todayTotalOrder
	res["today_total_order"] = todayTotalOrder
	res["today_nota_total"] = todayTotalOrder

	// Daily Missions Logic
	emp, _ := u.repo.GetEmployeeByID(employeeID)

	targetVisits := 0
	targetNewStores := 0

	if salesKPI != nil {
		targetVisits = salesKPI.TargetDailyVisits
		targetNewStores = salesKPI.TargetDailyNewStores
	}

	if targetVisits == 0 || targetNewStores == 0 {
		if emp != nil && emp.JobPosition != nil {
			title := emp.JobPosition.Title
			if title == "Sales Motoris" || title == "Motoris" {
				if targetVisits == 0 {
					targetVisits = 15
				}
				if targetNewStores == 0 {
					targetNewStores = 2
				}
			} else if title == "Sales Booster" || title == "Booster" {
				if targetVisits == 0 {
					targetVisits = 20
				}
				if targetNewStores == 0 {
					targetNewStores = 3
				}
			} else if title == "Sales Task Order" || title == "Task Order" {
				if targetVisits == 0 {
					targetVisits = 10
				}
				if targetNewStores == 0 {
					targetNewStores = 0
				}
			} else {
				if targetVisits == 0 {
					targetVisits = 10
				}
				if targetNewStores == 0 {
					targetNewStores = 1
				}
			}
		} else {
			if targetVisits == 0 {
				targetVisits = 10
			}
			if targetNewStores == 0 {
				targetNewStores = 1
			}
		}
	}

	todayVisits, _ := u.repo.GetTodayVisitsCount(employeeID)
	todayNewStores, _ := u.repo.GetTodayNewStoresCount(employeeID)

	res["daily_missions"] = map[string]interface{}{
		"target_visits":       targetVisits,
		"achieved_visits":     todayVisits,
		"target_new_stores":   targetNewStores,
		"achieved_new_stores": todayNewStores,
	}

	return res, nil
}

func (u *performanceUseCase) GetMyPerformanceHistory(employeeID uuid.UUID) ([]map[string]interface{}, error) {
	regHistory, _ := u.repo.GetEmployeeKPIHistory(employeeID)
	salesHistory, _ := u.repo.GetSalesKPIHistory(employeeID)

	var history []map[string]interface{}

	for _, k := range regHistory {
		history = append(history, map[string]interface{}{
			"period":      fmt.Sprintf("%04d-%02d", k.PeriodYear, k.PeriodMonth),
			"final_score": k.FinalScore,
			"type":        "INTERNAL",
		})
	}

	for _, k := range salesHistory {
		var finalScore float64
		if k.TargetOmzet > 0 {
			finalScore = (k.AchievedOmzet / k.TargetOmzet) * 100
		}
		history = append(history, map[string]interface{}{
			"period":      fmt.Sprintf("%04d-%02d", k.PeriodYear, k.PeriodMonth),
			"final_score": finalScore,
			"type":        "SALES",
		})
	}

	return history, nil
}
