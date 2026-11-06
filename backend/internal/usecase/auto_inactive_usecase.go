package usecase

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"example.com/hris/backend/internal/repository"
	"gorm.io/gorm"
)

type AutoInactiveSummary struct {
	EvaluatedCount   int                 `json:"evaluated_count"`
	DeactivatedCount int                 `json:"deactivated_count"`
	DeactivatedList  []DeactivatedDetail `json:"deactivated_list"`
	ExecutionTime    time.Time           `json:"execution_time"`
}

type DeactivatedDetail struct {
	EmployeeID       uuid.UUID  `json:"employee_id"`
	UserID           uuid.UUID  `json:"user_id"`
	EmployeeIDNumber string     `json:"employee_id_number"`
	FullName         string     `json:"full_name"`
	Email            string     `json:"email"`
	BranchID         *uuid.UUID `json:"branch_id"`
	BranchName       string     `json:"branch_name"`
	DepartmentName   string     `json:"department_name"`
	ConsecutiveDays  int        `json:"consecutive_days"`
	PeriodStart      string     `json:"period_start"`
	PeriodEnd        string     `json:"period_end"`
}

type CandidateDetail struct {
	EmployeeID       uuid.UUID `json:"employee_id"`
	UserID           uuid.UUID `json:"user_id"`
	EmployeeIDNumber string    `json:"employee_id_number"`
	FullName         string    `json:"full_name"`
	Email            string    `json:"email"`
	BranchName       string    `json:"branch_name"`
	DepartmentName   string    `json:"department_name"`
	JobPositionName  string    `json:"job_position_name"`
	ConsecutiveDays  int       `json:"consecutive_days"`
	PeriodStart      string    `json:"period_start"`
	PeriodEnd        string    `json:"period_end"`
	JoinDate         string    `json:"join_date"`
	LastClockInDate  *string   `json:"last_clock_in_date"`
	IsSales          bool      `json:"is_sales"`
	LastVisitDate    *string   `json:"last_visit_date,omitempty"`
	ActivePlansCount int       `json:"active_plans_count,omitempty"`
}

type AutoInactiveUsecase interface {
	EvaluateAndDeactivateInactiveEmployees(ctx context.Context, dryRun bool) (*AutoInactiveSummary, error)
	GetAutoInactiveCandidates(ctx context.Context, minConsecutiveDays int) ([]CandidateDetail, error)
	ReactivateEmployee(ctx context.Context, empID uuid.UUID, adminUserID uuid.UUID, adminName string, reason string) error
}

type autoInactiveUsecase struct {
	db               *gorm.DB
	employeeRepo     repository.EmployeeRepository
	userRepo         repository.UserRepository
	attendanceRepo   repository.AttendanceRepository
	leaveRepo        repository.LeaveRepository
	auditLogRepo     repository.AuditLogRepository
	notificationRepo repository.NotificationRepository
	fcmUseCase       FcmUsecase
}

func NewAutoInactiveUsecase(
	db *gorm.DB,
	employeeRepo repository.EmployeeRepository,
	userRepo repository.UserRepository,
	attendanceRepo repository.AttendanceRepository,
	leaveRepo repository.LeaveRepository,
	auditLogRepo repository.AuditLogRepository,
	notificationRepo repository.NotificationRepository,
	fcmUseCase FcmUsecase,
) AutoInactiveUsecase {
	return &autoInactiveUsecase{
		db:               db,
		employeeRepo:     employeeRepo,
		userRepo:         userRepo,
		attendanceRepo:   attendanceRepo,
		leaveRepo:        leaveRepo,
		auditLogRepo:     auditLogRepo,
		notificationRepo: notificationRepo,
		fcmUseCase:       fcmUseCase,
	}
}

// EvaluateAndDeactivateInactiveEmployees evaluates active employees who have no clock-ins, no approved leaves, and no sales visits for 10 consecutive days.
func (u *autoInactiveUsecase) EvaluateAndDeactivateInactiveEmployees(ctx context.Context, dryRun bool) (*AutoInactiveSummary, error) {
	wib := time.FixedZone("WIB", 7*3600)
	nowWIB := time.Now().In(wib)

	// Past 10 calendar days: H-10 to H-1 (excluding today because today's work shift is still in progress)
	yesterdayWIB := nowWIB.AddDate(0, 0, -1)
	endPeriod := time.Date(yesterdayWIB.Year(), yesterdayWIB.Month(), yesterdayWIB.Day(), 23, 59, 59, 999999999, wib)
	startPeriodWIB := yesterdayWIB.AddDate(0, 0, -9) // 10 days total: from (H-10) to (H-1)
	startPeriod := time.Date(startPeriodWIB.Year(), startPeriodWIB.Month(), startPeriodWIB.Day(), 0, 0, 0, 0, wib)
	todayStartWIB := time.Date(nowWIB.Year(), nowWIB.Month(), nowWIB.Day(), 0, 0, 0, 0, wib)

	summary := &AutoInactiveSummary{
		DeactivatedList: make([]DeactivatedDetail, 0),
		ExecutionTime:   nowWIB,
	}

	// Fetch all active employees (excluding Super Admin, management, daily workers, and attendance-exempt accounts)
	var employees []domain.Employee
	err := u.db.Preload("User").Preload("Branch").Preload("Department").Preload("JobPosition").
		Where("COALESCE(employees.employment_status, 'Active') NOT IN ('Inactive', 'INACTIVE', 'Resigned', 'RESIGNED', 'Terminated', 'TERMINATED')").
		Where("COALESCE(employees.employment_status, '') NOT ILIKE '%harian%'").
		Joins("JOIN users ON users.id = employees.user_id").
		Where("users.is_active = true").
		Where("LOWER(users.role) NOT IN ('superadmin', 'hr_admin', 'manager', 'auditor')").
		Where("users.can_absen_anywhere = false OR users.can_absen_anywhere IS NULL").
		Find(&employees).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch active employees: %w", err)
	}

	summary.EvaluatedCount = len(employees)

	// Evaluate each employee
	for _, emp := range employees {
		// 0. Double-check exemption in memory
		if isExemptFromAutoInactive(&emp) {
			continue
		}

		// 1. Skip if employee joined recently (less than 10 days ago)
		if !emp.JoinDate.IsZero() && emp.JoinDate.After(startPeriod) {
			continue
		}

		// 2. If employee has active / future sales visit plans scheduled for today or upcoming dates, skip deactivation
		var activePlansCount int64
		u.db.Model(&domain.SalesVisitPlan{}).
			Where("employee_id = ? AND plan_date >= ?", emp.ID, todayStartWIB.Format("2006-01-02")).
			Count(&activePlansCount)
		if activePlansCount > 0 {
			continue
		}

		isEligibleForDeactivation, consecutiveDays := u.checkConsecutiveAbsence(&emp, startPeriod, endPeriod, 10, wib)
		if !isEligibleForDeactivation {
			continue
		}

		branchName := "-"
		if emp.Branch != nil {
			branchName = emp.Branch.Name
		}
		deptName := "-"
		if emp.Department != nil {
			deptName = emp.Department.Name
		}
		fullName := emp.FirstName
		if emp.LastName != "" {
			fullName += " " + emp.LastName
		}
		email := ""
		if emp.User != nil {
			email = emp.User.Email
		}

		detail := DeactivatedDetail{
			EmployeeID:       emp.ID,
			UserID:           emp.UserID,
			EmployeeIDNumber: emp.EmployeeIDNumber,
			FullName:         fullName,
			Email:            email,
			BranchID:         emp.BranchID,
			BranchName:       branchName,
			DepartmentName:   deptName,
			ConsecutiveDays:  consecutiveDays,
			PeriodStart:      startPeriod.Format("2006-01-02"),
			PeriodEnd:        endPeriod.Format("2006-01-02"),
		}

		if !dryRun {
			// Execute Deactivation in a DB Transaction
			txErr := u.db.Transaction(func(tx *gorm.DB) error {
				// 1. Update Employee status
				if err := tx.Model(&domain.Employee{}).Where("id = ?", emp.ID).Update("employment_status", "Inactive").Error; err != nil {
					return err
				}

				// 2. Update User account
				if err := tx.Model(&domain.User{}).Where("id = ?", emp.UserID).Update("is_active", false).Error; err != nil {
					return err
				}

				// 3. Create Audit Log
				companyID := uuid.Nil
				if emp.CompanyID != nil {
					companyID = *emp.CompanyID
				}
				auditDesc := fmt.Sprintf("Auto-Inactive: Karyawan dinonaktifkan otomatis oleh sistem karena mangkir %d hari berturut-turut (%s s/d %s) tanpa absensi dan tanpa cuti.",
					consecutiveDays, startPeriod.Format("02-01-2006"), endPeriod.Format("02-01-2006"))
				
				audit := &domain.AuditLog{
					CompanyID:   companyID,
					UserID:      emp.UserID,
					UserName:    "SYSTEM (Auto-Inactive)",
					Action:      "AUTO_DEACTIVATE",
					Entity:      "Employee",
					EntityID:    emp.ID.String(),
					Description: auditDesc,
					OldValue:    fmt.Sprintf("employment_status: %s, is_active: true", emp.EmploymentStatus),
					NewValue:    "employment_status: Inactive, is_active: false",
					CreatedAt:   time.Now(),
				}
				if err := tx.Create(audit).Error; err != nil {
					log.Printf("[AutoInactive] Warning: failed to save audit log: %v", err)
				}

				// 4. Create Notification for HR
				if companyID != uuid.Nil {
					notif := &domain.Notification{
						CompanyID: companyID,
						BranchID:  emp.BranchID,
						Title:     "Karyawan Dinonaktifkan Otomatis",
						Message:   fmt.Sprintf("Karyawan %s (%s) dinonaktifkan otomatis karena tidak hadir dan tidak ada cuti selama %d hari berturut-turut.", fullName, emp.EmployeeIDNumber, consecutiveDays),
						Type:      "EMPLOYEE_AUTO_INACTIVE",
						CreatedAt: time.Now(),
					}
					_ = tx.Create(notif).Error
				}

				return nil
			})

			if txErr != nil {
				log.Printf("[AutoInactive] Error deactivating employee %s (%s): %v", fullName, emp.ID, txErr)
				continue
			}

			// Invalidate Redis cache
			invalidateEmployeeCache(context.Background(), emp.ID, emp.UserID)

			// Kick mobile session
			if u.fcmUseCase != nil {
				_ = u.fcmUseCase.SendForceLogoutExcept(emp.UserID, "", "auto-inactive")
			}
		}

		summary.DeactivatedList = append(summary.DeactivatedList, detail)
		summary.DeactivatedCount++
	}

	return summary, nil
}

// GetAutoInactiveCandidates retrieves employees who have consecutive absence days >= minConsecutiveDays (useful for monitoring / early warning)
func (u *autoInactiveUsecase) GetAutoInactiveCandidates(ctx context.Context, minConsecutiveDays int) ([]CandidateDetail, error) {
	if minConsecutiveDays <= 0 {
		minConsecutiveDays = 7 // default to 7 days warning threshold
	}

	wib := time.FixedZone("WIB", 7*3600)
	nowWIB := time.Now().In(wib)
	yesterdayWIB := nowWIB.AddDate(0, 0, -1)
	endPeriod := time.Date(yesterdayWIB.Year(), yesterdayWIB.Month(), yesterdayWIB.Day(), 23, 59, 59, 999999999, wib)
	startPeriodWIB := yesterdayWIB.AddDate(0, 0, -(minConsecutiveDays - 1))
	startPeriod := time.Date(startPeriodWIB.Year(), startPeriodWIB.Month(), startPeriodWIB.Day(), 0, 0, 0, 0, wib)

	var employees []domain.Employee
	err := u.db.Preload("User").Preload("Branch").Preload("Department").Preload("JobPosition").
		Where("COALESCE(employees.employment_status, 'Active') NOT IN ('Inactive', 'INACTIVE', 'Resigned', 'RESIGNED', 'Terminated', 'TERMINATED')").
		Where("COALESCE(employees.employment_status, '') NOT ILIKE '%harian%'").
		Joins("JOIN users ON users.id = employees.user_id").
		Where("users.is_active = true").
		Where("LOWER(users.role) NOT IN ('superadmin', 'hr_admin', 'manager', 'auditor')").
		Where("users.can_absen_anywhere = false OR users.can_absen_anywhere IS NULL").
		Find(&employees).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch active employees: %w", err)
	}

	candidates := make([]CandidateDetail, 0)

	for _, emp := range employees {
		// 0. Double-check exemption in memory
		if isExemptFromAutoInactive(&emp) {
			continue
		}

		if !emp.JoinDate.IsZero() && emp.JoinDate.After(startPeriod) {
			continue
		}

		// Calculate consecutive missing days backwards from yesterday
		consecutiveDays := u.countConsecutiveAbsenceDaysBackwards(&emp, yesterdayWIB, 30, wib)
		if consecutiveDays < minConsecutiveDays {
			continue
		}

		branchName := "-"
		if emp.Branch != nil {
			branchName = emp.Branch.Name
		}
		deptName := "-"
		if emp.Department != nil {
			deptName = emp.Department.Name
		}
		jobPosName := "-"
		if emp.JobPosition != nil {
			jobPosName = emp.JobPosition.Title
		}
		fullName := emp.FirstName
		if emp.LastName != "" {
			fullName += " " + emp.LastName
		}
		email := ""
		if emp.User != nil {
			email = emp.User.Email
		}

		// Find last valid clock in date
		var lastAttendance domain.AttendanceLog
		var lastClockInDate *string
		if err := u.db.Where("employee_id = ? AND clock_in_time IS NOT NULL AND status != 'ALFA'", emp.ID).
			Order("clock_in_time DESC").First(&lastAttendance).Error; err == nil && lastAttendance.ClockInTime != nil {
			str := lastAttendance.ClockInTime.In(wib).Format("2006-01-02 15:04")
			lastClockInDate = &str
		}

		// Sales-specific metadata
		isSales := isSalesEmployee(&emp)
		var lastVisitDate *string
		var activePlansCount int64

		if isSales {
			var lastVisit domain.SalesVisit
			if err := u.db.Where("employee_id = ? AND (status IS NULL OR status != 'SKIPPED')", emp.ID).
				Order("check_in_time DESC").First(&lastVisit).Error; err == nil && !lastVisit.CheckInTime.IsZero() {
				str := lastVisit.CheckInTime.In(wib).Format("2006-01-02 15:04")
				lastVisitDate = &str
			}

			u.db.Model(&domain.SalesVisitPlan{}).
				Where("employee_id = ? AND plan_date >= ?", emp.ID, nowWIB.Format("2006-01-02")).
				Count(&activePlansCount)
		}

		candidates = append(candidates, CandidateDetail{
			EmployeeID:       emp.ID,
			UserID:           emp.UserID,
			EmployeeIDNumber: emp.EmployeeIDNumber,
			FullName:         fullName,
			Email:            email,
			BranchName:       branchName,
			DepartmentName:   deptName,
			JobPositionName:  jobPosName,
			ConsecutiveDays:  consecutiveDays,
			PeriodStart:      startPeriod.Format("2006-01-02"),
			PeriodEnd:        endPeriod.Format("2006-01-02"),
			JoinDate:         emp.JoinDate.Format("2006-01-02"),
			LastClockInDate:  lastClockInDate,
			IsSales:          isSales,
			LastVisitDate:    lastVisitDate,
			ActivePlansCount: int(activePlansCount),
		})
	}

	return candidates, nil
}

// ReactivateEmployee re-activates an inactive employee and logs the action
func (u *autoInactiveUsecase) ReactivateEmployee(ctx context.Context, empID uuid.UUID, adminUserID uuid.UUID, adminName string, reason string) error {
	var emp domain.Employee
	if err := u.db.Preload("User").Where("id = ?", empID).First(&emp).Error; err != nil {
		return fmt.Errorf("employee not found: %w", err)
	}

	oldStatus := emp.EmploymentStatus

	err := u.db.Transaction(func(tx *gorm.DB) error {
		// 1. Update Employee status to Active
		if err := tx.Model(&domain.Employee{}).Where("id = ?", empID).Update("employment_status", "Active").Error; err != nil {
			return err
		}

		// 2. Update User account to Active
		if err := tx.Model(&domain.User{}).Where("id = ?", emp.UserID).Update("is_active", true).Error; err != nil {
			return err
		}

		// 3. Record Audit Log
		companyID := uuid.Nil
		if emp.CompanyID != nil {
			companyID = *emp.CompanyID
		}
		if adminName == "" {
			adminName = "ADMIN"
		}
		desc := fmt.Sprintf("Reaktivasi Karyawan oleh %s. Alasan: %s", adminName, reason)

		audit := &domain.AuditLog{
			CompanyID:   companyID,
			UserID:      adminUserID,
			UserName:    adminName,
			Action:      "REACTIVATE",
			Entity:      "Employee",
			EntityID:    emp.ID.String(),
			Description: desc,
			OldValue:    fmt.Sprintf("employment_status: %s, is_active: false", oldStatus),
			NewValue:    "employment_status: Active, is_active: true",
			CreatedAt:   time.Now(),
		}
		if err := tx.Create(audit).Error; err != nil {
			log.Printf("[ReactivateEmployee] Warning: failed to save audit log: %v", err)
		}

		return nil
	})

	if err != nil {
		return err
	}

	invalidateEmployeeCache(context.Background(), emp.ID, emp.UserID)
	return nil
}

// checkConsecutiveAbsence checks if an employee has NO clock-ins, NO approved leaves, and NO sales visits/orders for N consecutive days between start and end.
func (u *autoInactiveUsecase) checkConsecutiveAbsence(emp *domain.Employee, startPeriod, endPeriod time.Time, requiredDays int, loc *time.Location) (bool, int) {
	// 1. Query all attendance logs in period with clock_in_time and status != 'ALFA'
	var attendances []domain.AttendanceLog
	u.db.Where("employee_id = ? AND clock_in_time >= ? AND clock_in_time <= ? AND status != 'ALFA'",
		emp.ID, startPeriod.UTC(), endPeriod.UTC()).Find(&attendances)

	attendanceDateMap := make(map[string]bool)
	for _, att := range attendances {
		if att.ClockInTime != nil {
			dateStr := att.ClockInTime.In(loc).Format("2006-01-02")
			attendanceDateMap[dateStr] = true
		}
	}

	// 2. Query approved leaves overlapping the period
	var leaves []domain.LeaveRequest
	u.db.Where("employee_id = ? AND status = 'APPROVED'", emp.ID).
		Where("((start_date <= ? AND end_date >= ?) OR (start_date >= ? AND start_date <= ?))",
			endPeriod, startPeriod, startPeriod, endPeriod).
		Find(&leaves)

	leaveDateMap := make(map[string]bool)
	for _, l := range leaves {
		cur := l.StartDate
		for !cur.After(l.EndDate) {
			leaveDateMap[cur.Format("2006-01-02")] = true
			cur = cur.AddDate(0, 0, 1)
		}
	}

	// 3. Query sales activities (Sales Visits, Sales Transactions, Sales Orders)
	salesDateMap := make(map[string]bool)

	var salesVisits []domain.SalesVisit
	u.db.Where("employee_id = ? AND check_in_time >= ? AND check_in_time <= ? AND (status IS NULL OR status != 'SKIPPED')",
		emp.ID, startPeriod.UTC(), endPeriod.UTC()).Find(&salesVisits)
	for _, sv := range salesVisits {
		dateStr := sv.CheckInTime.In(loc).Format("2006-01-02")
		salesDateMap[dateStr] = true
	}

	var salesTransactions []domain.SalesTransaction
	u.db.Where("employee_id = ? AND created_at >= ? AND created_at <= ?",
		emp.ID, startPeriod.UTC(), endPeriod.UTC()).Find(&salesTransactions)
	for _, st := range salesTransactions {
		dateStr := st.CreatedAt.In(loc).Format("2006-01-02")
		salesDateMap[dateStr] = true
	}

	var salesOrders []domain.SalesOrder
	u.db.Where("employee_id = ? AND created_at >= ? AND created_at <= ?",
		emp.ID, startPeriod.UTC(), endPeriod.UTC()).Find(&salesOrders)
	for _, so := range salesOrders {
		dateStr := so.CreatedAt.In(loc).Format("2006-01-02")
		salesDateMap[dateStr] = true
	}

	// Iterate each of the evaluated days
	missingDays := 0
	curr := startPeriod
	for !curr.After(endPeriod) {
		dateStr := curr.Format("2006-01-02")
		hasAttendance := attendanceDateMap[dateStr]
		hasLeave := leaveDateMap[dateStr]
		hasSalesActivity := salesDateMap[dateStr]

		// If employee attended, had approved leave, or conducted sales visits/transactions on this day, the consecutive absence streak is broken!
		if hasAttendance || hasLeave || hasSalesActivity {
			return false, missingDays
		}

		missingDays++
		curr = curr.AddDate(0, 0, 1)
	}

	return missingDays >= requiredDays, missingDays
}

// countConsecutiveAbsenceDaysBackwards counts consecutive days backwards from yesterday where no clock-in, no approved leave, and no sales visits exist
func (u *autoInactiveUsecase) countConsecutiveAbsenceDaysBackwards(emp *domain.Employee, yesterday time.Time, maxDays int, loc *time.Location) int {
	// If employee has active or future visit plans today onwards, they are actively working
	todayStart := time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day()+1, 0, 0, 0, 0, loc)
	var futurePlansCount int64
	u.db.Model(&domain.SalesVisitPlan{}).
		Where("employee_id = ? AND plan_date >= ?", emp.ID, todayStart.Format("2006-01-02")).
		Count(&futurePlansCount)
	if futurePlansCount > 0 {
		return 0
	}

	startDate := yesterday.AddDate(0, 0, -(maxDays - 1))
	startPeriod := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, loc)
	endPeriod := time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 23, 59, 59, 999999999, loc)

	var attendances []domain.AttendanceLog
	u.db.Where("employee_id = ? AND clock_in_time >= ? AND clock_in_time <= ? AND status != 'ALFA'",
		emp.ID, startPeriod.UTC(), endPeriod.UTC()).Find(&attendances)

	attendanceDateMap := make(map[string]bool)
	for _, att := range attendances {
		if att.ClockInTime != nil {
			dateStr := att.ClockInTime.In(loc).Format("2006-01-02")
			attendanceDateMap[dateStr] = true
		}
	}

	var leaves []domain.LeaveRequest
	u.db.Where("employee_id = ? AND status = 'APPROVED'", emp.ID).
		Where("((start_date <= ? AND end_date >= ?) OR (start_date >= ? AND start_date <= ?))",
			endPeriod, startPeriod, startPeriod, endPeriod).
		Find(&leaves)

	leaveDateMap := make(map[string]bool)
	for _, l := range leaves {
		cur := l.StartDate
		for !cur.After(l.EndDate) {
			leaveDateMap[cur.Format("2006-01-02")] = true
			cur = cur.AddDate(0, 0, 1)
		}
	}

	salesDateMap := make(map[string]bool)

	var salesVisits []domain.SalesVisit
	u.db.Where("employee_id = ? AND check_in_time >= ? AND check_in_time <= ? AND (status IS NULL OR status != 'SKIPPED')",
		emp.ID, startPeriod.UTC(), endPeriod.UTC()).Find(&salesVisits)
	for _, sv := range salesVisits {
		dateStr := sv.CheckInTime.In(loc).Format("2006-01-02")
		salesDateMap[dateStr] = true
	}

	var salesTransactions []domain.SalesTransaction
	u.db.Where("employee_id = ? AND created_at >= ? AND created_at <= ?",
		emp.ID, startPeriod.UTC(), endPeriod.UTC()).Find(&salesTransactions)
	for _, st := range salesTransactions {
		dateStr := st.CreatedAt.In(loc).Format("2006-01-02")
		salesDateMap[dateStr] = true
	}

	var salesOrders []domain.SalesOrder
	u.db.Where("employee_id = ? AND created_at >= ? AND created_at <= ?",
		emp.ID, startPeriod.UTC(), endPeriod.UTC()).Find(&salesOrders)
	for _, so := range salesOrders {
		dateStr := so.CreatedAt.In(loc).Format("2006-01-02")
		salesDateMap[dateStr] = true
	}

	consecutiveDays := 0
	curr := yesterday
	for i := 0; i < maxDays; i++ {
		dateStr := curr.Format("2006-01-02")
		hasAttendance := attendanceDateMap[dateStr]
		hasLeave := leaveDateMap[dateStr]
		hasSalesActivity := salesDateMap[dateStr]

		if hasAttendance || hasLeave || hasSalesActivity {
			break
		}

		consecutiveDays++
		curr = curr.AddDate(0, 0, -1)
	}

	return consecutiveDays
}

// isSalesEmployee determines if an employee belongs to the Sales team
func isSalesEmployee(emp *domain.Employee) bool {
	if emp == nil {
		return false
	}
	if emp.User != nil {
		role := strings.ToLower(string(emp.User.Role))
		if strings.Contains(role, "sales") {
			return true
		}
	}
	if emp.JobPosition != nil {
		title := strings.ToLower(emp.JobPosition.Title)
		if strings.Contains(title, "sales") ||
			strings.Contains(title, "marketing") ||
			strings.Contains(title, "motoris") ||
			strings.Contains(title, "booster") ||
			strings.Contains(title, "take order") ||
			strings.Contains(title, "task order") ||
			strings.Contains(title, "modern trade") ||
			strings.Contains(title, "key account") ||
			strings.Contains(title, "kam") ||
			strings.Contains(title, "canvas") ||
			strings.Contains(title, "kolektor") ||
			strings.Contains(title, "to") ||
			strings.Contains(title, "mt") {
			// exclude false positives like "operator", "driver", "auditor", "gudang", "security"
			if !strings.Contains(title, "operator") &&
				!strings.Contains(title, "driver") &&
				!strings.Contains(title, "auditor") &&
				!strings.Contains(title, "gudang") &&
				!strings.Contains(title, "security") {
				return true
			}
		}
	}
	if emp.Department != nil {
		dept := strings.ToLower(emp.Department.Name)
		if strings.Contains(dept, "sales") || strings.Contains(dept, "marketing") || strings.Contains(dept, "modern trade") || strings.Contains(dept, "take order") {
			return true
		}
	}
	return false
}

// isExemptFromAutoInactive returns true if the employee/user should never be automatically deactivated due to missing attendance
func isExemptFromAutoInactive(emp *domain.Employee) bool {
	if emp == nil {
		return false
	}
	// Pekerja harian tidak menggunakan absensi reguler, kebal dari auto-inactivation
	empStatus := strings.ToLower(emp.EmploymentStatus)
	if strings.Contains(empStatus, "harian") || strings.Contains(empStatus, "daily") {
		return true
	}
	if emp.User == nil {
		return false
	}
	role := strings.ToLower(string(emp.User.Role))
	// Roles exempt from auto-inactivation: superadmin, hr_admin, manager, auditor
	if role == string(domain.RoleSuperAdmin) ||
		role == string(domain.RoleHRAdmin) ||
		role == string(domain.RoleManager) ||
		role == string(domain.RoleAuditor) {
		return true
	}
	// Users with explicit permission to absen anywhere / executive mobile exemption
	if emp.User.CanAbsenAnywhere {
		return true
	}
	return false
}
