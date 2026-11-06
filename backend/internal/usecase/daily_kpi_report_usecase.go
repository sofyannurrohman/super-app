package usecase

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"example.com/hris/backend/internal/repository"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

type CreateDailyKpiReportRequest struct {
	ReportDate time.Time `json:"report_date"`
	BranchID   *uuid.UUID `json:"branch_id"`
	AreaName   string    `json:"area_name"`
	IsDraft    bool      `json:"is_draft"`

	// Summary Tim (bisa diisi atau di-calculate otomatis dari items)
	SummaryTargetCall      int     `json:"summary_target_call"`
	SummaryRealizationCall int     `json:"summary_realization_call"`
	SummaryTargetEC        int     `json:"summary_target_ec"`
	SummaryRealizationEC   int     `json:"summary_realization_ec"`
	SummaryTargetSales     float64 `json:"summary_target_sales"`
	SummaryRealizationSales float64 `json:"summary_realization_sales"`

	TomorrowPriority string `json:"tomorrow_priority"`
	Notes            string `json:"notes"`

	Items       []DailyKpiItemInput       `json:"items"`
	ActionPlans []DailyKpiActionPlanInput `json:"action_plans"`
}

type DailyKpiItemInput struct {
	EmployeeID       uuid.UUID `json:"employee_id"`
	EmployeeName     string    `json:"employee_name"`
	AreaRoute        string    `json:"area_route"`
	TargetCall       int       `json:"target_call"`
	RealizationCall  int       `json:"realization_call"`
	TargetEC         int       `json:"target_ec"`
	RealizationEC    int       `json:"realization_ec"`
	TargetSales      float64   `json:"target_sales"`
	RealizationSales float64   `json:"realization_sales"`
}

type DailyKpiActionPlanInput struct {
	EmployeeID       *uuid.UUID `json:"employee_id"`
	EmployeeName     string     `json:"employee_name"`
	ReasonUnachieved string     `json:"reason_unachieved"`
	FieldObstacle    string     `json:"field_obstacle"`
	ActionFollowUp   string     `json:"action_follow_up"`
	PicName          string     `json:"pic_name"`
	DueDate          *time.Time `json:"due_date"`
	Status           string     `json:"status"`
	SupportNeeded    string     `json:"support_needed"`
}

type PrefillResponse struct {
	ReportDate             string                            `json:"report_date"`
	ReporterName           string                            `json:"reporter_name"`
	AreaName               string                            `json:"area_name"`
	SummaryTargetCall      int                               `json:"summary_target_call"`
	SummaryRealizationCall int                               `json:"summary_realization_call"`
	SummaryTargetEC        int                               `json:"summary_target_ec"`
	SummaryRealizationEC   int                               `json:"summary_realization_ec"`
	SummaryTargetSales     float64                           `json:"summary_target_sales"`
	SummaryRealizationSales float64                           `json:"summary_realization_sales"`
	Items                  []repository.SubordinateActualKPI `json:"items"`
}

type DailyKpiReportUsecase interface {
	GetPrefillData(reporterID uuid.UUID, branchID *uuid.UUID, date time.Time) (*PrefillResponse, error)
	CreateReport(req CreateDailyKpiReportRequest, reporterEmployeeID uuid.UUID, userRole string) (*domain.DailyKpiReport, error)
	UpdateReport(id uuid.UUID, req CreateDailyKpiReportRequest, currentEmployeeID uuid.UUID, isSuperadmin bool) (*domain.DailyKpiReport, error)
	GetReportByID(id uuid.UUID) (*domain.DailyKpiReport, error)
	GetReports(filter repository.DailyKpiFilter) ([]domain.DailyKpiReport, int64, error)
	SubmitReport(id uuid.UUID, currentEmployeeID uuid.UUID) error
	ReviewReport(id uuid.UUID, reviewerEmployeeID uuid.UUID, status string, feedback string) error
	DeleteReport(id uuid.UUID, currentEmployeeID uuid.UUID, isSuperadmin bool) error
	ExportExcel(id uuid.UUID) ([]byte, string, error)
}

type dailyKpiReportUsecase struct {
	repo         repository.DailyKpiReportRepository
	employeeRepo repository.EmployeeRepository
	db           *gorm.DB
}

func NewDailyKpiReportUsecase(
	repo repository.DailyKpiReportRepository,
	employeeRepo repository.EmployeeRepository,
	db *gorm.DB,
) DailyKpiReportUsecase {
	return &dailyKpiReportUsecase{
		repo:         repo,
		employeeRepo: employeeRepo,
		db:           db,
	}
}

func (u *dailyKpiReportUsecase) GetPrefillData(reporterID uuid.UUID, branchID *uuid.UUID, date time.Time) (*PrefillResponse, error) {
	reporter, err := u.employeeRepo.FindByID(reporterID)
	if err != nil || reporter == nil {
		return nil, errors.New("reporter employee not found")
	}

	reporterName := strings.TrimSpace(reporter.FirstName + " " + reporter.LastName)
	areaName := ""
	if reporter.Branch != nil {
		areaName = reporter.Branch.Name
	}
	if reporter.OperatingBranch != nil && areaName == "" {
		areaName = reporter.OperatingBranch.Name
	}

	activeBranchID := branchID
	if activeBranchID == nil {
		if reporter.OperatingBranchID != nil {
			activeBranchID = reporter.OperatingBranchID
		} else {
			activeBranchID = reporter.BranchID
		}
	}

	items, err := u.repo.GetSubordinateActualData(reporterID, activeBranchID, date)
	if err != nil {
		return nil, err
	}

	sumTargetCall := 0
	sumRealCall := 0
	sumTargetEC := 0
	sumRealEC := 0
	sumTargetSales := 0.0
	sumRealSales := 0.0

	for _, it := range items {
		sumTargetCall += it.TargetCall
		sumRealCall += it.RealizationCall
		sumTargetEC += it.TargetEC
		sumRealEC += it.RealizationEC
		sumTargetSales += it.TargetSales
		sumRealSales += it.RealizationSales
	}

	return &PrefillResponse{
		ReportDate:             date.Format("2006-01-02"),
		ReporterName:           reporterName,
		AreaName:               areaName,
		SummaryTargetCall:      sumTargetCall,
		SummaryRealizationCall: sumRealCall,
		SummaryTargetEC:        sumTargetEC,
		SummaryRealizationEC:   sumRealEC,
		SummaryTargetSales:     sumTargetSales,
		SummaryRealizationSales: sumRealSales,
		Items:                  items,
	}, nil
}

func (u *dailyKpiReportUsecase) CreateReport(req CreateDailyKpiReportRequest, reporterEmployeeID uuid.UUID, userRole string) (*domain.DailyKpiReport, error) {
	reporter, err := u.employeeRepo.FindByID(reporterEmployeeID)
	if err != nil || reporter == nil {
		return nil, errors.New("reporter employee not found")
	}

	reporterRole := "SPV"
	lowerRole := strings.ToLower(userRole)
	if strings.Contains(lowerRole, "asm") || strings.Contains(strings.ToLower(reporter.EmploymentStatus), "asm") {
		reporterRole = "ASM"
	} else if strings.Contains(lowerRole, "nsm") {
		reporterRole = "NSM"
	} else if strings.Contains(lowerRole, "gm") || strings.Contains(lowerRole, "manager") {
		reporterRole = "GM"
	}

	status := "SUBMITTED"
	var submittedAt *time.Time
	if req.IsDraft {
		status = "DRAFT"
	} else {
		now := time.Now()
		submittedAt = &now
	}

	// Tentukan reviewer secara otomatis jika pelapor memiliki manager_id
	var reviewerID *uuid.UUID
	if reporter.ManagerID != nil {
		reviewerID = reporter.ManagerID
	}

	// Hitung kalkulasi item & summary
	sumTargetCall := 0
	sumRealCall := 0
	sumTargetEC := 0
	sumRealEC := 0
	sumTargetSales := 0.0
	sumRealSales := 0.0

	domainItems := make([]domain.DailyKpiReportItem, 0, len(req.Items))
	for _, it := range req.Items {
		capaiCall := 0.0
		if it.TargetCall > 0 {
			capaiCall = (float64(it.RealizationCall) / float64(it.TargetCall)) * 100.0
		}
		capaiEC := 0.0
		if it.TargetEC > 0 {
			capaiEC = (float64(it.RealizationEC) / float64(it.TargetEC)) * 100.0
		}
		capaiSales := 0.0
		if it.TargetSales > 0 {
			capaiSales = (it.RealizationSales / it.TargetSales) * 100.0
		}
		diffSales := it.RealizationSales - it.TargetSales

		sumTargetCall += it.TargetCall
		sumRealCall += it.RealizationCall
		sumTargetEC += it.TargetEC
		sumRealEC += it.RealizationEC
		sumTargetSales += it.TargetSales
		sumRealSales += it.RealizationSales

		domainItems = append(domainItems, domain.DailyKpiReportItem{
			EmployeeID:             it.EmployeeID,
			EmployeeName:           it.EmployeeName,
			AreaRoute:              it.AreaRoute,
			TargetCall:             it.TargetCall,
			RealizationCall:        it.RealizationCall,
			AchievementCallPercent: capaiCall,
			TargetEC:               it.TargetEC,
			RealizationEC:          it.RealizationEC,
			AchievementECPercent:   capaiEC,
			TargetSales:            it.TargetSales,
			RealizationSales:       it.RealizationSales,
			AchievementSalesPercent: capaiSales,
			DiffSales:              diffSales,
		})
	}

	if req.SummaryTargetCall > 0 {
		sumTargetCall = req.SummaryTargetCall
	}
	if req.SummaryRealizationCall > 0 {
		sumRealCall = req.SummaryRealizationCall
	}
	if req.SummaryTargetEC > 0 {
		sumTargetEC = req.SummaryTargetEC
	}
	if req.SummaryRealizationEC > 0 {
		sumRealEC = req.SummaryRealizationEC
	}
	if req.SummaryTargetSales > 0 {
		sumTargetSales = req.SummaryTargetSales
	}
	if req.SummaryRealizationSales > 0 {
		sumRealSales = req.SummaryRealizationSales
	}

	sumAchieveCall := 0.0
	if sumTargetCall > 0 {
		sumAchieveCall = (float64(sumRealCall) / float64(sumTargetCall)) * 100.0
	}
	sumAchieveEC := 0.0
	if sumTargetEC > 0 {
		sumAchieveEC = (float64(sumRealEC) / float64(sumTargetEC)) * 100.0
	}
	sumAchieveSales := 0.0
	if sumTargetSales > 0 {
		sumAchieveSales = (sumRealSales / sumTargetSales) * 100.0
	}

	domainActionPlans := make([]domain.DailyKpiReportActionPlan, 0, len(req.ActionPlans))
	for _, ap := range req.ActionPlans {
		statusPlan := ap.Status
		if statusPlan == "" {
			statusPlan = "OPEN"
		}
		domainActionPlans = append(domainActionPlans, domain.DailyKpiReportActionPlan{
			EmployeeID:       ap.EmployeeID,
			EmployeeName:     ap.EmployeeName,
			ReasonUnachieved: ap.ReasonUnachieved,
			FieldObstacle:    ap.FieldObstacle,
			ActionFollowUp:   ap.ActionFollowUp,
			PicName:          ap.PicName,
			DueDate:          ap.DueDate,
			Status:           statusPlan,
			SupportNeeded:    ap.SupportNeeded,
		})
	}

	report := &domain.DailyKpiReport{
		ReportDate:                     req.ReportDate,
		ReporterEmployeeID:             reporterEmployeeID,
		ReporterRole:                   reporterRole,
		BranchID:                       req.BranchID,
		AreaName:                       req.AreaName,
		Status:                         status,
		SubmittedAt:                    submittedAt,
		ReviewerEmployeeID:             reviewerID,
		SummaryTargetCall:              sumTargetCall,
		SummaryRealizationCall:         sumRealCall,
		SummaryAchievementCallPercent:  sumAchieveCall,
		SummaryTargetEC:                sumTargetEC,
		SummaryRealizationEC:           sumRealEC,
		SummaryAchievementECPercent:   sumAchieveEC,
		SummaryTargetSales:             sumTargetSales,
		SummaryRealizationSales:        sumRealSales,
		SummaryAchievementSalesPercent: sumAchieveSales,
		TomorrowPriority:               req.TomorrowPriority,
		Notes:                          req.Notes,
		Items:                          domainItems,
		ActionPlans:                    domainActionPlans,
	}

	if err := u.repo.CreateReport(report); err != nil {
		return nil, err
	}

	return u.GetReportByID(report.ID)
}

func (u *dailyKpiReportUsecase) UpdateReport(id uuid.UUID, req CreateDailyKpiReportRequest, currentEmployeeID uuid.UUID, isSuperadmin bool) (*domain.DailyKpiReport, error) {
	existing, err := u.repo.GetReportByID(id)
	if err != nil || existing == nil {
		return nil, errors.New("report not found")
	}

	if !isSuperadmin && existing.ReporterEmployeeID != currentEmployeeID {
		return nil, errors.New("unauthorized: only the reporter or superadmin can edit this report")
	}

	if existing.Status == "APPROVED" && !isSuperadmin {
		return nil, errors.New("approved reports cannot be modified")
	}

	if req.IsDraft {
		existing.Status = "DRAFT"
	} else if existing.Status == "DRAFT" || existing.Status == "REVISED" {
		existing.Status = "SUBMITTED"
		now := time.Now()
		existing.SubmittedAt = &now
	}

	existing.ReportDate = req.ReportDate
	if req.BranchID != nil {
		existing.BranchID = req.BranchID
	}
	if req.AreaName != "" {
		existing.AreaName = req.AreaName
	}
	existing.TomorrowPriority = req.TomorrowPriority
	existing.Notes = req.Notes

	sumTargetCall := 0
	sumRealCall := 0
	sumTargetEC := 0
	sumRealEC := 0
	sumTargetSales := 0.0
	sumRealSales := 0.0

	domainItems := make([]domain.DailyKpiReportItem, 0, len(req.Items))
	for _, it := range req.Items {
		capaiCall := 0.0
		if it.TargetCall > 0 {
			capaiCall = (float64(it.RealizationCall) / float64(it.TargetCall)) * 100.0
		}
		capaiEC := 0.0
		if it.TargetEC > 0 {
			capaiEC = (float64(it.RealizationEC) / float64(it.TargetEC)) * 100.0
		}
		capaiSales := 0.0
		if it.TargetSales > 0 {
			capaiSales = (it.RealizationSales / it.TargetSales) * 100.0
		}
		diffSales := it.RealizationSales - it.TargetSales

		sumTargetCall += it.TargetCall
		sumRealCall += it.RealizationCall
		sumTargetEC += it.TargetEC
		sumRealEC += it.RealizationEC
		sumTargetSales += it.TargetSales
		sumRealSales += it.RealizationSales

		domainItems = append(domainItems, domain.DailyKpiReportItem{
			DailyKpiReportID:       existing.ID,
			EmployeeID:             it.EmployeeID,
			EmployeeName:           it.EmployeeName,
			AreaRoute:              it.AreaRoute,
			TargetCall:             it.TargetCall,
			RealizationCall:        it.RealizationCall,
			AchievementCallPercent: capaiCall,
			TargetEC:               it.TargetEC,
			RealizationEC:          it.RealizationEC,
			AchievementECPercent:   capaiEC,
			TargetSales:            it.TargetSales,
			RealizationSales:       it.RealizationSales,
			AchievementSalesPercent: capaiSales,
			DiffSales:              diffSales,
		})
	}

	if req.SummaryTargetCall > 0 {
		sumTargetCall = req.SummaryTargetCall
	}
	if req.SummaryRealizationCall > 0 {
		sumRealCall = req.SummaryRealizationCall
	}
	if req.SummaryTargetEC > 0 {
		sumTargetEC = req.SummaryTargetEC
	}
	if req.SummaryRealizationEC > 0 {
		sumRealEC = req.SummaryRealizationEC
	}
	if req.SummaryTargetSales > 0 {
		sumTargetSales = req.SummaryTargetSales
	}
	if req.SummaryRealizationSales > 0 {
		sumRealSales = req.SummaryRealizationSales
	}

	existing.SummaryTargetCall = sumTargetCall
	existing.SummaryRealizationCall = sumRealCall
	if sumTargetCall > 0 {
		existing.SummaryAchievementCallPercent = (float64(sumRealCall) / float64(sumTargetCall)) * 100.0
	} else {
		existing.SummaryAchievementCallPercent = 0
	}

	existing.SummaryTargetEC = sumTargetEC
	existing.SummaryRealizationEC = sumRealEC
	if sumTargetEC > 0 {
		existing.SummaryAchievementECPercent = (float64(sumRealEC) / float64(sumTargetEC)) * 100.0
	} else {
		existing.SummaryAchievementECPercent = 0
	}

	existing.SummaryTargetSales = sumTargetSales
	existing.SummaryRealizationSales = sumRealSales
	if sumTargetSales > 0 {
		existing.SummaryAchievementSalesPercent = (sumRealSales / sumTargetSales) * 100.0
	} else {
		existing.SummaryAchievementSalesPercent = 0
	}

	domainActionPlans := make([]domain.DailyKpiReportActionPlan, 0, len(req.ActionPlans))
	for _, ap := range req.ActionPlans {
		statusPlan := ap.Status
		if statusPlan == "" {
			statusPlan = "OPEN"
		}
		domainActionPlans = append(domainActionPlans, domain.DailyKpiReportActionPlan{
			DailyKpiReportID: existing.ID,
			EmployeeID:       ap.EmployeeID,
			EmployeeName:     ap.EmployeeName,
			ReasonUnachieved: ap.ReasonUnachieved,
			FieldObstacle:    ap.FieldObstacle,
			ActionFollowUp:   ap.ActionFollowUp,
			PicName:          ap.PicName,
			DueDate:          ap.DueDate,
			Status:           statusPlan,
			SupportNeeded:    ap.SupportNeeded,
		})
	}

	existing.Items = domainItems
	existing.ActionPlans = domainActionPlans

	if err := u.repo.UpdateReport(existing); err != nil {
		return nil, err
	}

	return u.GetReportByID(existing.ID)
}

func (u *dailyKpiReportUsecase) GetReportByID(id uuid.UUID) (*domain.DailyKpiReport, error) {
	return u.repo.GetReportByID(id)
}

func (u *dailyKpiReportUsecase) GetReports(filter repository.DailyKpiFilter) ([]domain.DailyKpiReport, int64, error) {
	return u.repo.GetReports(filter)
}

func (u *dailyKpiReportUsecase) SubmitReport(id uuid.UUID, currentEmployeeID uuid.UUID) error {
	report, err := u.repo.GetReportByID(id)
	if err != nil || report == nil {
		return errors.New("report not found")
	}

	if report.ReporterEmployeeID != currentEmployeeID {
		return errors.New("unauthorized: only the reporter can submit this report")
	}

	report.Status = "SUBMITTED"
	now := time.Now()
	report.SubmittedAt = &now

	return u.repo.UpdateReport(report)
}

func (u *dailyKpiReportUsecase) ReviewReport(id uuid.UUID, reviewerEmployeeID uuid.UUID, status string, feedback string) error {
	report, err := u.repo.GetReportByID(id)
	if err != nil || report == nil {
		return errors.New("report not found")
	}

	cleanStatus := strings.ToUpper(strings.TrimSpace(status))
	if cleanStatus != "APPROVED" && cleanStatus != "REVISED" {
		return errors.New("invalid status: must be APPROVED or REVISED")
	}

	report.Status = cleanStatus
	report.ReviewerEmployeeID = &reviewerEmployeeID
	report.ReviewerFeedback = feedback
	now := time.Now()
	report.ReviewedAt = &now

	return u.repo.UpdateReport(report)
}

func (u *dailyKpiReportUsecase) DeleteReport(id uuid.UUID, currentEmployeeID uuid.UUID, isSuperadmin bool) error {
	report, err := u.repo.GetReportByID(id)
	if err != nil || report == nil {
		return errors.New("report not found")
	}

	if !isSuperadmin && report.ReporterEmployeeID != currentEmployeeID {
		return errors.New("unauthorized: only reporter or superadmin can delete")
	}

	if !isSuperadmin && report.Status != "DRAFT" {
		return errors.New("only draft reports can be deleted")
	}

	return u.repo.DeleteReport(id)
}

func (u *dailyKpiReportUsecase) ExportExcel(id uuid.UUID) ([]byte, string, error) {
	report, err := u.repo.GetReportByID(id)
	if err != nil || report == nil {
		return nil, "", errors.New("report not found")
	}

	f := excelize.NewFile()
	sheet := "Laporan KPI Sales"
	f.SetSheetName("Sheet1", sheet)

	// Styles
	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 14, Color: "0F172A"},
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center"},
	})
	headerMetaLabelStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 10, Color: "334155"},
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center"},
	})
	headerMetaValStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 10, Color: "0284C7"},
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center"},
	})
	tableHeaderDarkStyle, _ := f.NewStyle(&excelize.Style{
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"1E293B"}, Pattern: 1},
		Font:      &excelize.Font{Bold: true, Color: "FFFFFF", Size: 10},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
		Border: []excelize.Border{
			{Type: "top", Style: 1, Color: "94A3B8"},
			{Type: "bottom", Style: 1, Color: "94A3B8"},
			{Type: "left", Style: 1, Color: "94A3B8"},
			{Type: "right", Style: 1, Color: "94A3B8"},
		},
	})
	yellowInputStyle, _ := f.NewStyle(&excelize.Style{
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"FEF9C3"}, Pattern: 1}, // Kuning muda (diisi SPV)
		Font:      &excelize.Font{Size: 10, Color: "1E293B"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "top", Style: 1, Color: "CBD5E1"},
			{Type: "bottom", Style: 1, Color: "CBD5E1"},
			{Type: "left", Style: 1, Color: "CBD5E1"},
			{Type: "right", Style: 1, Color: "CBD5E1"},
		},
	})
	grayFormulaStyle, _ := f.NewStyle(&excelize.Style{
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"F1F5F9"}, Pattern: 1}, // Abu muda (rumus otomatis)
		Font:      &excelize.Font{Bold: true, Size: 10, Color: "1E293B"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "top", Style: 1, Color: "CBD5E1"},
			{Type: "bottom", Style: 1, Color: "CBD5E1"},
			{Type: "left", Style: 1, Color: "CBD5E1"},
			{Type: "right", Style: 1, Color: "CBD5E1"},
		},
	})
	currencyYellowStyle, _ := f.NewStyle(&excelize.Style{
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"FEF9C3"}, Pattern: 1},
		Font:      &excelize.Font{Size: 10, Color: "1E293B"},
		Alignment: &excelize.Alignment{Horizontal: "right", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "top", Style: 1, Color: "CBD5E1"},
			{Type: "bottom", Style: 1, Color: "CBD5E1"},
			{Type: "left", Style: 1, Color: "CBD5E1"},
			{Type: "right", Style: 1, Color: "CBD5E1"},
		},
	})
	currencyGrayStyle, _ := f.NewStyle(&excelize.Style{
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"F1F5F9"}, Pattern: 1},
		Font:      &excelize.Font{Bold: true, Size: 10, Color: "0F766E"},
		Alignment: &excelize.Alignment{Horizontal: "right", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "top", Style: 1, Color: "CBD5E1"},
			{Type: "bottom", Style: 1, Color: "CBD5E1"},
			{Type: "left", Style: 1, Color: "CBD5E1"},
			{Type: "right", Style: 1, Color: "CBD5E1"},
		},
	})
	cellBorderNormal, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: 10, Color: "1E293B"},
		Alignment: &excelize.Alignment{Vertical: "center"},
		Border: []excelize.Border{
			{Type: "top", Style: 1, Color: "E2E8F0"},
			{Type: "bottom", Style: 1, Color: "E2E8F0"},
			{Type: "left", Style: 1, Color: "E2E8F0"},
			{Type: "right", Style: 1, Color: "E2E8F0"},
		},
	})
	legendStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Italic: true, Size: 9, Color: "64748B"},
		Alignment: &excelize.Alignment{Vertical: "center"},
	})

	// 1. Title
	f.SetCellValue(sheet, "A1", "LAPORAN KPI SALES HARIAN")
	f.SetCellStyle(sheet, "A1", "A1", titleStyle)

	// 2. Meta
	f.SetCellValue(sheet, "A3", "TANGGAL LAPORAN")
	f.SetCellValue(sheet, "B3", report.ReportDate.Format("02/01/2006"))
	f.SetCellValue(sheet, "D3", "NAMA SPV / PELAPOR")
	reporterName := "-"
	if report.ReporterEmployee != nil {
		reporterName = strings.TrimSpace(report.ReporterEmployee.FirstName + " " + report.ReporterEmployee.LastName)
	}
	f.SetCellValue(sheet, "E3", reporterName)
	f.SetCellValue(sheet, "G3", "CABANG / AREA")
	areaName := report.AreaName
	if areaName == "" && report.Branch != nil {
		areaName = report.Branch.Name
	}
	f.SetCellValue(sheet, "H3", areaName)

	f.SetCellStyle(sheet, "A3", "A3", headerMetaLabelStyle)
	f.SetCellStyle(sheet, "B3", "B3", headerMetaValStyle)
	f.SetCellStyle(sheet, "D3", "D3", headerMetaLabelStyle)
	f.SetCellStyle(sheet, "E3", "E3", headerMetaValStyle)
	f.SetCellStyle(sheet, "G3", "G3", headerMetaLabelStyle)
	f.SetCellStyle(sheet, "H3", "H3", headerMetaValStyle)

	// 3. Ringkasan Tim
	f.SetCellValue(sheet, "A5", "RINGKASAN TIM")
	f.SetCellStyle(sheet, "A5", "A5", headerMetaLabelStyle)

	f.SetCellValue(sheet, "A6", "KPI")
	f.SetCellValue(sheet, "B6", "TARGET")
	f.SetCellValue(sheet, "C6", "REALISASI")
	f.SetCellValue(sheet, "D6", "PENCAPAIAN")
	f.SetCellStyle(sheet, "A6", "D6", tableHeaderDarkStyle)

	// Petunjuk / Legend di kanan
	f.SetCellValue(sheet, "F5", "KUNING MUDA: DIISI SPV / EDITABLE")
	f.SetCellValue(sheet, "F6", "ABU MUDA: RUMUS OTOMATIS")
	f.SetCellValue(sheet, "F7", "ISI 0 JIKA REALISASI NIHIL. KOSONG = BELUM DILAPORKAN.")
	f.SetCellStyle(sheet, "F5", "F7", legendStyle)

	// Rows Ringkasan
	f.SetCellValue(sheet, "A7", "CALL (KUNJUNGAN)")
	f.SetCellValue(sheet, "B7", report.SummaryTargetCall)
	f.SetCellValue(sheet, "C7", report.SummaryRealizationCall)
	f.SetCellValue(sheet, "D7", fmt.Sprintf("%.1f%%", report.SummaryAchievementCallPercent))
	f.SetCellStyle(sheet, "A7", "A7", cellBorderNormal)
	f.SetCellStyle(sheet, "B7", "B7", yellowInputStyle)
	f.SetCellStyle(sheet, "C7", "C7", yellowInputStyle)
	f.SetCellStyle(sheet, "D7", "D7", grayFormulaStyle)

	f.SetCellValue(sheet, "A8", "EC (OUTLET ORDER)")
	f.SetCellValue(sheet, "B8", report.SummaryTargetEC)
	f.SetCellValue(sheet, "C8", report.SummaryRealizationEC)
	f.SetCellValue(sheet, "D8", fmt.Sprintf("%.1f%%", report.SummaryAchievementECPercent))
	f.SetCellStyle(sheet, "A8", "A8", cellBorderNormal)
	f.SetCellStyle(sheet, "B8", "B8", yellowInputStyle)
	f.SetCellStyle(sheet, "C8", "C8", yellowInputStyle)
	f.SetCellStyle(sheet, "D8", "D8", grayFormulaStyle)

	f.SetCellValue(sheet, "A9", "PENJUALAN (RP)")
	f.SetCellValue(sheet, "B9", fmt.Sprintf("Rp %s", formatDailyKpiRupiah(report.SummaryTargetSales)))
	f.SetCellValue(sheet, "C9", fmt.Sprintf("Rp %s", formatDailyKpiRupiah(report.SummaryRealizationSales)))
	f.SetCellValue(sheet, "D9", fmt.Sprintf("%.1f%%", report.SummaryAchievementSalesPercent))
	f.SetCellStyle(sheet, "A9", "A9", cellBorderNormal)
	f.SetCellStyle(sheet, "B9", "B9", currencyYellowStyle)
	f.SetCellStyle(sheet, "C9", "C9", currencyYellowStyle)
	f.SetCellStyle(sheet, "D9", "D9", grayFormulaStyle)

	// 4. Bagian A: Target dan Realisasi per Sales
	f.SetCellValue(sheet, "A11", "A. TARGET DAN REALISASI PER SALES")
	f.SetCellStyle(sheet, "A11", "A11", headerMetaLabelStyle)

	headersA := []string{
		"NAMA SALES", "AREA / RUTE",
		"TARGET CALL", "REALISASI CALL", "CAPAI CALL %",
		"TARGET EC", "REALISASI EC", "CAPAI EC %",
		"TARGET JUAL (RP)", "REALISASI (RP)", "CAPAI JUAL %", "SELISIH JUAL (RP)",
	}
	for i, h := range headersA {
		colName, _ := excelize.ColumnNumberToName(i + 1)
		f.SetCellValue(sheet, fmt.Sprintf("%s12", colName), h)
	}
	lastColA, _ := excelize.ColumnNumberToName(len(headersA))
	f.SetCellStyle(sheet, "A12", fmt.Sprintf("%s12", lastColA), tableHeaderDarkStyle)

	currentRow := 13
	for _, item := range report.Items {
		f.SetCellValue(sheet, fmt.Sprintf("A%d", currentRow), item.EmployeeName)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", currentRow), item.AreaRoute)
		f.SetCellValue(sheet, fmt.Sprintf("C%d", currentRow), item.TargetCall)
		f.SetCellValue(sheet, fmt.Sprintf("D%d", currentRow), item.RealizationCall)
		f.SetCellValue(sheet, fmt.Sprintf("E%d", currentRow), fmt.Sprintf("%.1f%%", item.AchievementCallPercent))
		f.SetCellValue(sheet, fmt.Sprintf("F%d", currentRow), item.TargetEC)
		f.SetCellValue(sheet, fmt.Sprintf("G%d", currentRow), item.RealizationEC)
		f.SetCellValue(sheet, fmt.Sprintf("H%d", currentRow), fmt.Sprintf("%.1f%%", item.AchievementECPercent))
		f.SetCellValue(sheet, fmt.Sprintf("I%d", currentRow), fmt.Sprintf("Rp %s", formatDailyKpiRupiah(item.TargetSales)))
		f.SetCellValue(sheet, fmt.Sprintf("J%d", currentRow), fmt.Sprintf("Rp %s", formatDailyKpiRupiah(item.RealizationSales)))
		f.SetCellValue(sheet, fmt.Sprintf("K%d", currentRow), fmt.Sprintf("%.1f%%", item.AchievementSalesPercent))
		f.SetCellValue(sheet, fmt.Sprintf("L%d", currentRow), fmt.Sprintf("Rp %s", formatDailyKpiRupiah(item.DiffSales)))

		f.SetCellStyle(sheet, fmt.Sprintf("A%d", currentRow), fmt.Sprintf("B%d", currentRow), cellBorderNormal)
		f.SetCellStyle(sheet, fmt.Sprintf("C%d", currentRow), fmt.Sprintf("D%d", currentRow), yellowInputStyle)
		f.SetCellStyle(sheet, fmt.Sprintf("E%d", currentRow), fmt.Sprintf("E%d", currentRow), grayFormulaStyle)
		f.SetCellStyle(sheet, fmt.Sprintf("F%d", currentRow), fmt.Sprintf("G%d", currentRow), yellowInputStyle)
		f.SetCellStyle(sheet, fmt.Sprintf("H%d", currentRow), fmt.Sprintf("H%d", currentRow), grayFormulaStyle)
		f.SetCellStyle(sheet, fmt.Sprintf("I%d", currentRow), fmt.Sprintf("J%d", currentRow), currencyYellowStyle)
		f.SetCellStyle(sheet, fmt.Sprintf("K%d", currentRow), fmt.Sprintf("L%d", currentRow), currencyGrayStyle)

		currentRow++
	}

	// 5. Bagian B: Evaluasi dan Tindak Lanjut SPV
	currentRow += 2
	f.SetCellValue(sheet, fmt.Sprintf("A%d", currentRow), "B. EVALUASI DAN TINDAK LANJUT SPV")
	f.SetCellStyle(sheet, fmt.Sprintf("A%d", currentRow), fmt.Sprintf("A%d", currentRow), headerMetaLabelStyle)

	currentRow++
	headersB := []string{
		"NAMA SALES", "ALASAN TARGET BELUM TERCAPAI", "KENDALA LAPANGAN",
		"TINDAK LANJUT SPV", "PIC", "BATAS WAKTU", "STATUS", "DUKUNGAN DIPERLUKAN",
	}
	for i, h := range headersB {
		colName, _ := excelize.ColumnNumberToName(i + 1)
		f.SetCellValue(sheet, fmt.Sprintf("%s%d", colName, currentRow), h)
	}
	lastColB, _ := excelize.ColumnNumberToName(len(headersB))
	f.SetCellStyle(sheet, fmt.Sprintf("A%d", currentRow), fmt.Sprintf("%s%d", lastColB, currentRow), tableHeaderDarkStyle)

	currentRow++
	if len(report.ActionPlans) == 0 {
		f.SetCellValue(sheet, fmt.Sprintf("A%d", currentRow), "-")
		f.SetCellValue(sheet, fmt.Sprintf("B%d", currentRow), "Seluruh target tercapai / belum ada evaluasi kendala")
		f.SetCellStyle(sheet, fmt.Sprintf("A%d", currentRow), fmt.Sprintf("%s%d", lastColB, currentRow), cellBorderNormal)
		currentRow++
	} else {
		for _, ap := range report.ActionPlans {
			dueDateStr := "-"
			if ap.DueDate != nil {
				dueDateStr = ap.DueDate.Format("02/01/2006")
			}
			f.SetCellValue(sheet, fmt.Sprintf("A%d", currentRow), ap.EmployeeName)
			f.SetCellValue(sheet, fmt.Sprintf("B%d", currentRow), ap.ReasonUnachieved)
			f.SetCellValue(sheet, fmt.Sprintf("C%d", currentRow), ap.FieldObstacle)
			f.SetCellValue(sheet, fmt.Sprintf("D%d", currentRow), ap.ActionFollowUp)
			f.SetCellValue(sheet, fmt.Sprintf("E%d", currentRow), ap.PicName)
			f.SetCellValue(sheet, fmt.Sprintf("F%d", currentRow), dueDateStr)
			f.SetCellValue(sheet, fmt.Sprintf("G%d", currentRow), ap.Status)
			f.SetCellValue(sheet, fmt.Sprintf("H%d", currentRow), ap.SupportNeeded)

			f.SetCellStyle(sheet, fmt.Sprintf("A%d", currentRow), fmt.Sprintf("%s%d", lastColB, currentRow), cellBorderNormal)
			currentRow++
		}
	}

	// 6. Bagian C: Catatan & Validasi SPV
	currentRow += 2
	f.SetCellValue(sheet, fmt.Sprintf("A%d", currentRow), "C. CATATAN DAN VALIDASI SPV")
	f.SetCellStyle(sheet, fmt.Sprintf("A%d", currentRow), fmt.Sprintf("A%d", currentRow), headerMetaLabelStyle)

	currentRow++
	f.SetCellValue(sheet, fmt.Sprintf("A%d", currentRow), "PRIORITAS BESOK:")
	f.SetCellStyle(sheet, fmt.Sprintf("A%d", currentRow), fmt.Sprintf("A%d", currentRow), headerMetaLabelStyle)
	currentRow++
	priorityText := report.TomorrowPriority
	if priorityText == "" {
		priorityText = "-"
	}
	f.SetCellValue(sheet, fmt.Sprintf("A%d", currentRow), priorityText)
	f.SetCellStyle(sheet, fmt.Sprintf("A%d", currentRow), fmt.Sprintf("A%d", currentRow), cellBorderNormal)

	currentRow += 2
	reviewerName := "-"
	if report.ReviewerEmployee != nil {
		reviewerName = strings.TrimSpace(report.ReviewerEmployee.FirstName + " " + report.ReviewerEmployee.LastName)
	}
	f.SetCellValue(sheet, fmt.Sprintf("B%d", currentRow), "DIBUAT OLEH (SPV / PELAPOR)")
	f.SetCellValue(sheet, fmt.Sprintf("G%d", currentRow), "DIPERIKSA OLEH (ATASAN)")
	f.SetCellStyle(sheet, fmt.Sprintf("B%d", currentRow), fmt.Sprintf("B%d", currentRow), headerMetaLabelStyle)
	f.SetCellStyle(sheet, fmt.Sprintf("G%d", currentRow), fmt.Sprintf("G%d", currentRow), headerMetaLabelStyle)

	currentRow += 3
	f.SetCellValue(sheet, fmt.Sprintf("B%d", currentRow), fmt.Sprintf("( %s )", reporterName))
	f.SetCellValue(sheet, fmt.Sprintf("G%d", currentRow), fmt.Sprintf("( %s )", reviewerName))
	f.SetCellStyle(sheet, fmt.Sprintf("B%d", currentRow), fmt.Sprintf("B%d", currentRow), headerMetaLabelStyle)
	f.SetCellStyle(sheet, fmt.Sprintf("G%d", currentRow), fmt.Sprintf("G%d", currentRow), headerMetaLabelStyle)

	// Footnotes
	currentRow += 2
	footnotes := []string{
		"CALL = JUMLAH OUTLET YANG DIKUNJUNGI. EC = JUMLAH OUTLET DIKUNJUNGI YANG MENGHASILKAN ORDER, MAKSIMAL 1 EC PER OUTLET/HARI.",
		"PENJUALAN = NILAI ORDER BERSIH SETELAH DISKON HARI ITU. SELISIH = REALISASI DIKURANG TARGET. PENCAPAIAN = REALISASI / TARGET.",
		"ALASAN: JELASKAN PENYEBAB TERUKUR. KENDALA: HAMBATAN LAPANGAN. TINDAK LANJUT: TULIS AKSI, PIC, DAN BATAS WAKTU.",
		"CONTOH: 4 OUTLET TIDAK ORDER KARENA STOK MASIH PENUH. KENDALA: STOK PRODUK FAST MOVING KOSONG. AKSI: SPV KOORDINASI GUDANG.",
	}
	for _, fn := range footnotes {
		f.SetCellValue(sheet, fmt.Sprintf("A%d", currentRow), fn)
		f.SetCellStyle(sheet, fmt.Sprintf("A%d", currentRow), fmt.Sprintf("A%d", currentRow), legendStyle)
		currentRow++
	}

	// Auto column widths
	_ = f.SetColWidth(sheet, "A", "A", 25)
	_ = f.SetColWidth(sheet, "B", "B", 20)
	_ = f.SetColWidth(sheet, "C", "C", 16)
	_ = f.SetColWidth(sheet, "D", "D", 16)
	_ = f.SetColWidth(sheet, "E", "E", 16)
	_ = f.SetColWidth(sheet, "F", "F", 16)
	_ = f.SetColWidth(sheet, "G", "G", 16)
	_ = f.SetColWidth(sheet, "H", "H", 20)
	_ = f.SetColWidth(sheet, "I", "I", 20)
	_ = f.SetColWidth(sheet, "J", "J", 20)
	_ = f.SetColWidth(sheet, "K", "K", 16)
	_ = f.SetColWidth(sheet, "L", "L", 20)

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, "", err
	}

	filename := fmt.Sprintf("Laporan_KPI_Sales_Harian_%s_%s.xlsx", report.ReportDate.Format("2006-01-02"), strings.ReplaceAll(areaName, " ", "_"))
	return buf.Bytes(), filename, nil
}

func formatDailyKpiRupiah(amount float64) string {
	valStr := fmt.Sprintf("%.0f", amount)
	n := len(valStr)
	if n <= 3 {
		return valStr
	}
	var res []byte
	rem := n % 3
	if rem > 0 {
		res = append(res, valStr[:rem]...)
		if n > rem {
			res = append(res, '.')
		}
	}
	for i := rem; i < n; i += 3 {
		res = append(res, valStr[i:i+3]...)
		if i+3 < n {
			res = append(res, '.')
		}
	}
	return string(res)
}
