package http

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"gorm.io/gorm"
)

type HRDashboardHandler struct {
	db *gorm.DB
}

func NewHRDashboardHandler(db *gorm.DB) *HRDashboardHandler {
	return &HRDashboardHandler{db: db}
}

func (h *HRDashboardHandler) RegisterRoutes(r *gin.RouterGroup) {
	group := r.Group("/hr/dashboard")
	group.Use(RoleMiddleware(
		string(domain.RoleSuperAdmin),
		string(domain.RoleHRAdmin),
		string(domain.RoleManager),
		string(domain.RoleAdminKasir),
		string(domain.RoleAdminOperasional),
		string(domain.RoleAdminNota),
	))
	{
		group.GET("/overview", h.GetOverview)
	}
}

// Response structs
type HRDashboardOverviewResponse struct {
	Summary                    HRDashboardSummary            `json:"summary"`
	AttendanceTrend            []HRAttendanceTrendPoint      `json:"attendance_trend"`
	DepartmentDistribution     []HRDepartmentDistribution    `json:"department_distribution"`
	EmploymentDistribution     []HREmploymentDistribution    `json:"employment_distribution"`
	ExpiringContracts          []HRExpiringContract          `json:"expiring_contracts"`
	RecentAttendance           []HRRecentAttendance          `json:"recent_attendance"`
	PendingLeaves              []HRPendingLeave              `json:"pending_leaves"`
	PendingOvertimes           []HRPendingOvertime           `json:"pending_overtimes"`
}

type HRDashboardSummary struct {
	TotalEmployees             int     `json:"total_employees"`
	PresentToday               int     `json:"present_today"`
	OnTimeToday                int     `json:"on_time_today"`
	LateToday                  int     `json:"late_today"`
	OnLeaveToday               int     `json:"on_leave_today"`
	NotCheckedInToday          int     `json:"not_checked_in_today"`
	AttendanceRate             float64 `json:"attendance_rate"`
	PendingLeavesCount         int     `json:"pending_leaves_count"`
	PendingOvertimesCount      int     `json:"pending_overtimes_count"`
	PendingReimbursementsCount int     `json:"pending_reimbursements_count"`
	ExpiringContractsCount     int     `json:"expiring_contracts_count"`
}

type HRAttendanceTrendPoint struct {
	Date           string  `json:"date"`
	Label          string  `json:"label"`
	Present        int     `json:"present"`
	OnTime         int     `json:"on_time"`
	Late           int     `json:"late"`
	Leave          int     `json:"leave"`
	AttendanceRate float64 `json:"attendance_rate"`
}

type HRDepartmentDistribution struct {
	Department string  `json:"department"`
	Count      int     `json:"count"`
	Percentage float64 `json:"percentage"`
}

type HREmploymentDistribution struct {
	Status string `json:"status"`
	Count  int    `json:"count"`
}

type HRExpiringContract struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	EmployeeIDNumber string `json:"employee_id_number"`
	Department       string `json:"department"`
	JobPosition      string `json:"job_position"`
	EndDate          string `json:"end_date"`
	DaysRemaining    int    `json:"days_remaining"`
}

type HRRecentAttendance struct {
	ID          string `json:"id"`
	EmployeeID  string `json:"employee_id"`
	Name        string `json:"name"`
	JobPosition string `json:"job_position"`
	Department  string `json:"department"`
	ClockInTime string `json:"clock_in_time"`
	Status      string `json:"status"`
	Notes       string `json:"notes"`
	AvatarURL   string `json:"avatar_url"`
}

type HRPendingLeave struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
	Days      int    `json:"days"`
	Reason    string `json:"reason"`
	CreatedAt string `json:"created_at"`
	TimeAgo   string `json:"time_ago"`
}

type HRPendingOvertime struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Date        string  `json:"date"`
	DurationHrs float64 `json:"duration_hrs"`
	Reason      string  `json:"reason"`
	CreatedAt   string  `json:"created_at"`
	TimeAgo     string  `json:"time_ago"`
}

func (h *HRDashboardHandler) GetOverview(c *gin.Context) {
	loc, _ := time.LoadLocation("Asia/Jakarta")
	if loc == nil {
		loc = time.Local
	}
	now := time.Now().In(loc)
	todayStr := now.Format("2006-01-02")

	// Filter parameters
	branchIDStr := c.Query("branch_id")
	rangeParam := c.DefaultQuery("range", "7d") // 7d, 30d, month, year

	var branchID *uuid.UUID
	if branchIDStr != "" && branchIDStr != "ALL_BRANCHES" {
		if bid, err := uuid.Parse(branchIDStr); err == nil {
			branchID = &bid
		}
	}

	companyID := parseCompanyID(c)

	resp := HRDashboardOverviewResponse{
		AttendanceTrend:        make([]HRAttendanceTrendPoint, 0),
		DepartmentDistribution: make([]HRDepartmentDistribution, 0),
		EmploymentDistribution: make([]HREmploymentDistribution, 0),
		ExpiringContracts:      make([]HRExpiringContract, 0),
		RecentAttendance:       make([]HRRecentAttendance, 0),
		PendingLeaves:          make([]HRPendingLeave, 0),
		PendingOvertimes:       make([]HRPendingOvertime, 0),
	}

	// 1. Total Active Employees
	empQuery := h.db.Table("employees").
		Where("employment_status NOT IN ('RESIGNED', 'TERMINATED', 'INACTIVE')")
	if branchID != nil {
		empQuery = empQuery.Where("branch_id = ?", *branchID)
	}
	if companyID != uuid.Nil {
		empQuery = empQuery.Where("company_id = ?", companyID)
	}

	var totalEmployees int64
	empQuery.Count(&totalEmployees)
	resp.Summary.TotalEmployees = int(totalEmployees)

	// 2. Attendance Today
	type attTodayStat struct {
		TotalPresent int
		OnTime       int
		Late         int
	}
	var attStat attTodayStat
	attTodayQ := h.db.Table("attendance_logs al").
		Select(`
			COUNT(DISTINCT al.id) as total_present,
			COUNT(DISTINCT CASE WHEN al.status = 'ON_TIME' THEN al.id END) as on_time,
			COUNT(DISTINCT CASE WHEN al.status = 'LATE' THEN al.id END) as late
		`).
		Joins("JOIN employees e ON al.employee_id = e.id").
		Where("e.employment_status NOT IN ('RESIGNED', 'TERMINATED', 'INACTIVE')").
		Where("DATE(al.clock_in_time AT TIME ZONE 'Asia/Jakarta') = ?", todayStr)

	if branchID != nil {
		attTodayQ = attTodayQ.Where("e.branch_id = ?", *branchID)
	}
	if companyID != uuid.Nil {
		attTodayQ = attTodayQ.Where("e.company_id = ?", companyID)
	}
	attTodayQ.Scan(&attStat)

	resp.Summary.PresentToday = attStat.TotalPresent
	resp.Summary.OnTimeToday = attStat.OnTime
	resp.Summary.LateToday = attStat.Late

	// 3. On Leave Today
	var onLeaveToday int64
	leaveTodayQ := h.db.Table("leave_requests lr").
		Joins("JOIN employees e ON lr.employee_id = e.id").
		Where("e.employment_status NOT IN ('RESIGNED', 'TERMINATED', 'INACTIVE')").
		Where("lr.status = 'APPROVED'").
		Where("? BETWEEN lr.start_date AND lr.end_date", todayStr)
	if branchID != nil {
		leaveTodayQ = leaveTodayQ.Where("e.branch_id = ?", *branchID)
	}
	if companyID != uuid.Nil {
		leaveTodayQ = leaveTodayQ.Where("e.company_id = ?", companyID)
	}
	leaveTodayQ.Count(&onLeaveToday)
	resp.Summary.OnLeaveToday = int(onLeaveToday)

	// Not checked in calculation
	notCheckedIn := int(totalEmployees) - (attStat.TotalPresent + int(onLeaveToday))
	if notCheckedIn < 0 {
		notCheckedIn = 0
	}
	resp.Summary.NotCheckedInToday = notCheckedIn

	if totalEmployees > 0 {
		resp.Summary.AttendanceRate = float64(attStat.TotalPresent) / float64(totalEmployees) * 100
		if resp.Summary.AttendanceRate > 100 {
			resp.Summary.AttendanceRate = 100
		}
	}

	// 4. Pending Requests Counts
	var pendingLeavesCount int64
	plQ := h.db.Table("leave_requests lr").
		Joins("JOIN employees e ON lr.employee_id = e.id").
		Where("lr.status = 'PENDING'")
	if branchID != nil {
		plQ = plQ.Where("e.branch_id = ?", *branchID)
	}
	if companyID != uuid.Nil {
		plQ = plQ.Where("e.company_id = ?", companyID)
	}
	plQ.Count(&pendingLeavesCount)
	resp.Summary.PendingLeavesCount = int(pendingLeavesCount)

	var pendingOvertimesCount int64
	poQ := h.db.Table("overtimes o").
		Joins("JOIN employees e ON o.employee_id = e.id").
		Where("o.status = 'PENDING'")
	if branchID != nil {
		poQ = poQ.Where("e.branch_id = ?", *branchID)
	}
	if companyID != uuid.Nil {
		poQ = poQ.Where("e.company_id = ?", companyID)
	}
	poQ.Count(&pendingOvertimesCount)
	resp.Summary.PendingOvertimesCount = int(pendingOvertimesCount)

	var pendingReimburseCount int64
	prQ := h.db.Table("reimbursements r").
		Joins("JOIN employees e ON r.employee_id = e.id").
		Where("r.status = 'PENDING'")
	if branchID != nil {
		prQ = prQ.Where("e.branch_id = ?", *branchID)
	}
	if companyID != uuid.Nil {
		prQ = prQ.Where("e.company_id = ?", companyID)
	}
	prQ.Count(&pendingReimburseCount)
	resp.Summary.PendingReimbursementsCount = int(pendingReimburseCount)

	// 5. Expiring Contracts (< 30 days)
	in30Days := now.AddDate(0, 0, 30).Format("2006-01-02")
	var expContractsCount int64
	ecQ := h.db.Table("employees e").
		Where("e.employment_status NOT IN ('RESIGNED', 'TERMINATED', 'INACTIVE')").
		Where("e.end_date IS NOT NULL AND e.end_date >= ? AND e.end_date <= ?", todayStr, in30Days)
	if branchID != nil {
		ecQ = ecQ.Where("e.branch_id = ?", *branchID)
	}
	if companyID != uuid.Nil {
		ecQ = ecQ.Where("e.company_id = ?", companyID)
	}
	ecQ.Count(&expContractsCount)
	resp.Summary.ExpiringContractsCount = int(expContractsCount)

	// 6. Attendance Trend Calculation
	daysToFetch := 7
	if rangeParam == "30d" || rangeParam == "month" {
		daysToFetch = 30
	} else if rangeParam == "year" {
		daysToFetch = 365
	}

	if rangeParam == "year" {
		// Monthly aggregation for 12 months
		type monthlyRow struct {
			MonthYear string
			MonthKey  string
			Present   int
			OnTime    int
			Late      int
		}
		var mRows []monthlyRow
		mTrendQ := h.db.Raw(`
			SELECT 
				TO_CHAR(DATE_TRUNC('month', al.clock_in_time AT TIME ZONE 'Asia/Jakarta'), 'Mon YYYY') AS month_year,
				TO_CHAR(DATE_TRUNC('month', al.clock_in_time AT TIME ZONE 'Asia/Jakarta'), 'YYYY-MM') AS month_key,
				COUNT(DISTINCT al.id) AS present,
				COUNT(DISTINCT CASE WHEN al.status = 'ON_TIME' THEN al.id END) AS on_time,
				COUNT(DISTINCT CASE WHEN al.status = 'LATE' THEN al.id END) AS late
			FROM attendance_logs al
			JOIN employees e ON al.employee_id = e.id
			WHERE al.clock_in_time >= (NOW() AT TIME ZONE 'Asia/Jakarta') - INTERVAL '12 months'
			  AND e.employment_status NOT IN ('RESIGNED', 'TERMINATED', 'INACTIVE')
			  `+hrDashboardBranchFilter(branchID, companyID)+`
			GROUP BY DATE_TRUNC('month', al.clock_in_time AT TIME ZONE 'Asia/Jakarta')
			ORDER BY DATE_TRUNC('month', al.clock_in_time AT TIME ZONE 'Asia/Jakarta') ASC
		`)
		mTrendQ.Scan(&mRows)

		for _, mr := range mRows {
			rate := 0.0
			if totalEmployees > 0 {
				rate = float64(mr.Present) / float64(totalEmployees) * 100
				if rate > 100 {
					rate = 100
				}
			}
			resp.AttendanceTrend = append(resp.AttendanceTrend, HRAttendanceTrendPoint{
				Date:           mr.MonthKey,
				Label:          mr.MonthYear,
				Present:        mr.Present,
				OnTime:         mr.OnTime,
				Late:           mr.Late,
				Leave:          0,
				AttendanceRate: rate,
			})
		}
	} else {
		// Daily aggregation
		startDate := now.AddDate(0, 0, -(daysToFetch - 1))
		startDateStr := startDate.Format("2006-01-02")

		type dailyAttRow struct {
			LogDate string
			Present int
			OnTime  int
			Late    int
		}
		var dailyRows []dailyAttRow
		dTrendQ := h.db.Raw(`
			SELECT 
				TO_CHAR(al.clock_in_time AT TIME ZONE 'Asia/Jakarta', 'YYYY-MM-DD') AS log_date,
				COUNT(DISTINCT al.id) AS present,
				COUNT(DISTINCT CASE WHEN al.status = 'ON_TIME' THEN al.id END) AS on_time,
				COUNT(DISTINCT CASE WHEN al.status = 'LATE' THEN al.id END) AS late
			FROM attendance_logs al
			JOIN employees e ON al.employee_id = e.id
			WHERE DATE(al.clock_in_time AT TIME ZONE 'Asia/Jakarta') >= ?
			  AND e.employment_status NOT IN ('RESIGNED', 'TERMINATED', 'INACTIVE')
			  `+hrDashboardBranchFilter(branchID, companyID)+`
			GROUP BY TO_CHAR(al.clock_in_time AT TIME ZONE 'Asia/Jakarta', 'YYYY-MM-DD')
			ORDER BY log_date ASC
		`, startDateStr)
		dTrendQ.Scan(&dailyRows)

		dailyMap := make(map[string]dailyAttRow)
		for _, r := range dailyRows {
			dailyMap[r.LogDate] = r
		}

		// Daily Leaves query
		type dailyLeaveRow struct {
			LogDate    string
			LeaveCount int
		}
		var leaveRows []dailyLeaveRow
		h.db.Raw(`
			SELECT 
				d::date::text AS log_date,
				COUNT(DISTINCT lr.id) AS leave_count
			FROM generate_series(?::date, ?::date, '1 day'::interval) d
			JOIN leave_requests lr ON d::date BETWEEN lr.start_date AND lr.end_date AND lr.status = 'APPROVED'
			JOIN employees e ON lr.employee_id = e.id
			WHERE e.employment_status NOT IN ('RESIGNED', 'TERMINATED', 'INACTIVE')
			  `+hrDashboardBranchFilter(branchID, companyID)+`
			GROUP BY d::date
		`, startDateStr, todayStr).Scan(&leaveRows)

		leaveMap := make(map[string]int)
		for _, lr := range leaveRows {
			leaveMap[lr.LogDate] = lr.LeaveCount
		}

		dayNamesIndo := map[time.Weekday]string{
			time.Sunday:    "Min",
			time.Monday:    "Sen",
			time.Tuesday:   "Sel",
			time.Wednesday: "Rab",
			time.Thursday:  "Kam",
			time.Friday:    "Jum",
			time.Saturday:  "Sab",
		}

		monthNamesIndo := map[time.Month]string{
			time.January:   "Jan",
			time.February:  "Feb",
			time.March:     "Mar",
			time.April:     "Apr",
			time.May:       "Mei",
			time.June:      "Jun",
			time.July:      "Jul",
			time.August:    "Agu",
			time.September: "Sep",
			time.October:   "Okt",
			time.November:  "Nov",
			time.December:  "Des",
		}

		for i := 0; i < daysToFetch; i++ {
			curD := startDate.AddDate(0, 0, i)
			dKey := curD.Format("2006-01-02")
			label := fmt.Sprintf("%s, %02d %s", dayNamesIndo[curD.Weekday()], curD.Day(), monthNamesIndo[curD.Month()])

			att := dailyMap[dKey]
			lCnt := leaveMap[dKey]

			rate := 0.0
			if totalEmployees > 0 {
				rate = float64(att.Present) / float64(totalEmployees) * 100
				if rate > 100 {
					rate = 100
				}
			}

			resp.AttendanceTrend = append(resp.AttendanceTrend, HRAttendanceTrendPoint{
				Date:           dKey,
				Label:          label,
				Present:        att.Present,
				OnTime:         att.OnTime,
				Late:           att.Late,
				Leave:          lCnt,
				AttendanceRate: rate,
			})
		}
	}

	// 7. Department Distribution
	type deptRow struct {
		Name  string
		Count int
	}
	var deptRows []deptRow
	deptQ := h.db.Table("employees e").
		Select("COALESCE(d.name, 'Umum') AS name, COUNT(e.id) AS count").
		Joins("LEFT JOIN departments d ON e.department_id = d.id").
		Where("e.employment_status NOT IN ('RESIGNED', 'TERMINATED', 'INACTIVE')")
	if branchID != nil {
		deptQ = deptQ.Where("e.branch_id = ?", *branchID)
	}
	if companyID != uuid.Nil {
		deptQ = deptQ.Where("e.company_id = ?", companyID)
	}
	deptQ.Group("COALESCE(d.name, 'Umum')").Order("count DESC").Scan(&deptRows)

	for _, dr := range deptRows {
		pct := 0.0
		if totalEmployees > 0 {
			pct = float64(dr.Count) / float64(totalEmployees) * 100
		}
		resp.DepartmentDistribution = append(resp.DepartmentDistribution, HRDepartmentDistribution{
			Department: dr.Name,
			Count:      dr.Count,
			Percentage: pct,
		})
	}

	// 8. Employment Status Distribution
	type statusRow struct {
		Status string
		Count  int
	}
	var stRows []statusRow
	stQ := h.db.Table("employees e").
		Select("COALESCE(NULLIF(TRIM(e.employment_status), ''), 'Tetap') AS status, COUNT(e.id) AS count").
		Where("e.employment_status NOT IN ('RESIGNED', 'TERMINATED', 'INACTIVE')")
	if branchID != nil {
		stQ = stQ.Where("e.branch_id = ?", *branchID)
	}
	if companyID != uuid.Nil {
		stQ = stQ.Where("e.company_id = ?", companyID)
	}
	stQ.Group("COALESCE(NULLIF(TRIM(e.employment_status), ''), 'Tetap')").Order("count DESC").Scan(&stRows)

	for _, sr := range stRows {
		statusLabel := formatEmploymentStatus(sr.Status)
		resp.EmploymentDistribution = append(resp.EmploymentDistribution, HREmploymentDistribution{
			Status: statusLabel,
			Count:  sr.Count,
		})
	}

	// 9. Expiring Contracts List (Next 60 Days)
	in60Days := now.AddDate(0, 0, 60).Format("2006-01-02")
	type expRow struct {
		ID               string
		FirstName        string
		LastName         string
		EmployeeIDNumber string
		DepartmentName   string
		JobPositionTitle string
		EndDate          time.Time
	}
	var expRows []expRow
	h.db.Table("employees e").
		Select(`
			e.id,
			e.first_name,
			COALESCE(e.last_name, '') as last_name,
			e.employee_id_number,
			COALESCE(d.name, '-') as department_name,
			COALESCE(jp.title, '-') as job_position_title,
			e.end_date
		`).
		Joins("LEFT JOIN departments d ON e.department_id = d.id").
		Joins("LEFT JOIN job_positions jp ON e.job_position_id = jp.id").
		Where("e.employment_status NOT IN ('RESIGNED', 'TERMINATED', 'INACTIVE')").
		Where("e.end_date IS NOT NULL AND e.end_date >= ? AND e.end_date <= ?", todayStr, in60Days).
		Where(func() *gorm.DB {
			if branchID != nil {
				return h.db.Where("e.branch_id = ?", *branchID)
			}
			return h.db
		}()).
		Where(func() *gorm.DB {
			if companyID != uuid.Nil {
				return h.db.Where("e.company_id = ?", companyID)
			}
			return h.db
		}()).
		Order("e.end_date ASC").
		Limit(10).
		Scan(&expRows)

	for _, er := range expRows {
		daysLeft := int(er.EndDate.Sub(now).Hours() / 24)
		if daysLeft < 0 {
			daysLeft = 0
		}
		fullName := strings.TrimSpace(er.FirstName + " " + er.LastName)
		resp.ExpiringContracts = append(resp.ExpiringContracts, HRExpiringContract{
			ID:               er.ID,
			Name:             fullName,
			EmployeeIDNumber: er.EmployeeIDNumber,
			Department:       er.DepartmentName,
			JobPosition:      er.JobPositionTitle,
			EndDate:          er.EndDate.Format("02 Jan 2006"),
			DaysRemaining:    daysLeft,
		})
	}

	// 10. Recent Attendance Logs Today
	type attRecRow struct {
		ID          string
		EmployeeID  string
		FirstName   string
		LastName    string
		JobPosition string
		Department  string
		ClockInTime time.Time
		Status      string
		Notes       string
		FacePhoto   string
	}
	var attRecRows []attRecRow
	h.db.Table("attendance_logs al").
		Select(`
			al.id,
			al.employee_id,
			e.first_name,
			COALESCE(e.last_name, '') as last_name,
			COALESCE(jp.title, '-') as job_position,
			COALESCE(d.name, '-') as department,
			al.clock_in_time,
			al.status,
			COALESCE(al.notes, '') as notes,
			COALESCE(e.face_reference_url, '') as face_photo
		`).
		Joins("JOIN employees e ON al.employee_id = e.id").
		Joins("LEFT JOIN departments d ON e.department_id = d.id").
		Joins("LEFT JOIN job_positions jp ON e.job_position_id = jp.id").
		Where("DATE(al.clock_in_time AT TIME ZONE 'Asia/Jakarta') = ?", todayStr).
		Where(func() *gorm.DB {
			if branchID != nil {
				return h.db.Where("e.branch_id = ?", *branchID)
			}
			return h.db
		}()).
		Where(func() *gorm.DB {
			if companyID != uuid.Nil {
				return h.db.Where("e.company_id = ?", companyID)
			}
			return h.db
		}()).
		Order("al.clock_in_time DESC").
		Limit(10).
		Scan(&attRecRows)

	for _, ar := range attRecRows {
		fullName := strings.TrimSpace(ar.FirstName + " " + ar.LastName)
		clockStr := ar.ClockInTime.In(loc).Format("15:04")
		statusText := "Tepat Waktu"
		if ar.Status == "LATE" {
			statusText = "Terlambat"
		}
		resp.RecentAttendance = append(resp.RecentAttendance, HRRecentAttendance{
			ID:          ar.ID,
			EmployeeID:  ar.EmployeeID,
			Name:        fullName,
			JobPosition: ar.JobPosition,
			Department:  ar.Department,
			ClockInTime: clockStr,
			Status:      statusText,
			Notes:       ar.Notes,
			AvatarURL:   ar.FacePhoto,
		})
	}

	// 11. Pending Leaves (Top 5)
	type pLeaveRow struct {
		ID        string
		FirstName string
		LastName  string
		LeaveType string
		StartDate time.Time
		EndDate   time.Time
		Reason    string
		CreatedAt time.Time
	}
	var pLeaveRows []pLeaveRow
	h.db.Table("leave_requests lr").
		Select(`
			lr.id,
			e.first_name,
			COALESCE(e.last_name, '') as last_name,
			COALESCE(lt.name, 'Cuti') as leave_type,
			lr.start_date,
			lr.end_date,
			COALESCE(lr.reason, '') as reason,
			lr.created_at
		`).
		Joins("JOIN employees e ON lr.employee_id = e.id").
		Joins("LEFT JOIN leave_types lt ON lr.leave_type_id = lt.id").
		Where("lr.status = 'PENDING'").
		Where(func() *gorm.DB {
			if branchID != nil {
				return h.db.Where("e.branch_id = ?", *branchID)
			}
			return h.db
		}()).
		Where(func() *gorm.DB {
			if companyID != uuid.Nil {
				return h.db.Where("e.company_id = ?", companyID)
			}
			return h.db
		}()).
		Order("lr.created_at DESC").
		Limit(5).
		Scan(&pLeaveRows)

	for _, plr := range pLeaveRows {
		fullName := strings.TrimSpace(plr.FirstName + " " + plr.LastName)
		days := int(plr.EndDate.Sub(plr.StartDate).Hours()/24) + 1
		if days < 1 {
			days = 1
		}
		diffH := int(now.Sub(plr.CreatedAt).Hours())
		timeAgo := "Baru saja"
		if diffH >= 24 {
			timeAgo = fmt.Sprintf("%d hari lalu", diffH/24)
		} else if diffH >= 1 {
			timeAgo = fmt.Sprintf("%dj lalu", diffH)
		}

		resp.PendingLeaves = append(resp.PendingLeaves, HRPendingLeave{
			ID:        plr.ID,
			Name:      fullName,
			Type:      plr.LeaveType,
			StartDate: plr.StartDate.Format("02 Jan 2006"),
			EndDate:   plr.EndDate.Format("02 Jan 2006"),
			Days:      days,
			Reason:    plr.Reason,
			CreatedAt: plr.CreatedAt.Format("02 Jan 2006 15:04"),
			TimeAgo:   timeAgo,
		})
	}

	// 12. Pending Overtimes (Top 5)
	type pOtRow struct {
		ID        string
		FirstName string
		LastName  string
		Date      time.Time
		Duration  float64
		Reason    string
		CreatedAt time.Time
	}
	var pOtRows []pOtRow
	h.db.Table("overtimes o").
		Select(`
			o.id,
			e.first_name,
			COALESCE(e.last_name, '') as last_name,
			o.date,
			COALESCE(o.duration, 0) as duration,
			COALESCE(o.reason, '') as reason,
			o.created_at
		`).
		Joins("JOIN employees e ON o.employee_id = e.id").
		Where("o.status = 'PENDING'").
		Where(func() *gorm.DB {
			if branchID != nil {
				return h.db.Where("e.branch_id = ?", *branchID)
			}
			return h.db
		}()).
		Where(func() *gorm.DB {
			if companyID != uuid.Nil {
				return h.db.Where("e.company_id = ?", companyID)
			}
			return h.db
		}()).
		Order("o.created_at DESC").
		Limit(5).
		Scan(&pOtRows)

	for _, por := range pOtRows {
		fullName := strings.TrimSpace(por.FirstName + " " + por.LastName)
		diffH := int(now.Sub(por.CreatedAt).Hours())
		timeAgo := "Baru saja"
		if diffH >= 24 {
			timeAgo = fmt.Sprintf("%d hari lalu", diffH/24)
		} else if diffH >= 1 {
			timeAgo = fmt.Sprintf("%dj lalu", diffH)
		}

		resp.PendingOvertimes = append(resp.PendingOvertimes, HRPendingOvertime{
			ID:          por.ID,
			Name:        fullName,
			Date:        por.Date.Format("02 Jan 2006"),
			DurationHrs: por.Duration,
			Reason:      por.Reason,
			CreatedAt:   por.CreatedAt.Format("02 Jan 2006 15:04"),
			TimeAgo:     timeAgo,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"data": resp,
	})
}

func hrDashboardBranchFilter(branchID *uuid.UUID, companyID uuid.UUID) string {
	var conds []string
	if branchID != nil {
		conds = append(conds, fmt.Sprintf("e.branch_id = '%s'", branchID.String()))
	}
	if companyID != uuid.Nil {
		conds = append(conds, fmt.Sprintf("e.company_id = '%s'", companyID.String()))
	}
	if len(conds) == 0 {
		return ""
	}
	return "AND " + strings.Join(conds, " AND ")
}

func formatEmploymentStatus(status string) string {
	s := strings.ToUpper(strings.TrimSpace(status))
	switch s {
	case "PERMANENT", "TETAP":
		return "Karyawan Tetap"
	case "CONTRACT", "PKWT", "KONTRAK":
		return "Kontrak / PKWT"
	case "DAILY", "HARIAN", "PEKERJA_HARIAN", "HARIAN_LEPAS":
		return "Harian Lepas"
	case "PROBATION", "MAGANG", "INTERN":
		return "Probation / Magang"
	default:
		return strings.Title(strings.ToLower(s))
	}
}
