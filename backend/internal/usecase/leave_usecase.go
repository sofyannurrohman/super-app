package usecase

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"example.com/hris/backend/internal/repository"
	"example.com/hris/backend/pkg/sse"
	"gorm.io/gorm"
)

type LeaveUseCase interface {
	SubmitLeave(userID uuid.UUID, req SubmitLeaveRequest) error
	UpdateMyLeave(leaveID, userID uuid.UUID, req SubmitLeaveRequest) error
	DeleteMyLeave(leaveID, userID uuid.UUID) error
	AdminCreateLeave(req AdminLeaveRequest) error
	AdminUpdateLeave(leaveID uuid.UUID, req AdminLeaveRequest) error
	AdminDeleteLeave(leaveID uuid.UUID) error
	GetMyLeaves(userID uuid.UUID, page, limit int) ([]domain.LeaveRequest, error)
	GetMyBalances(userID uuid.UUID) ([]LeaveBalanceResponse, error)
	GetAllLeaves(status, search string, page, limit int, branchIDs []uuid.UUID) ([]domain.LeaveRequest, int64, error)
	GetAllLeaveBalances() ([]*domain.LeaveBalance, error)
	ApproveReturnLeave(leaveID, approverID uuid.UUID, req ApproveLeaveRequest) error
	AdminUpdateBalance(req AdminUpdateBalanceRequest) error
	AdminDeleteBalance(req AdminUpdateBalanceRequest) error
}

type leaveUseCase struct {
	db           *gorm.DB
	repo         repository.LeaveRepository
	employeeRepo repository.EmployeeRepository
	fcmUseCase   FcmUsecase
	sseHub       *sse.Hub
}

func NewLeaveUseCase(db *gorm.DB, repo repository.LeaveRepository, employeeRepo repository.EmployeeRepository, fcmUseCase FcmUsecase, sseHub *sse.Hub) LeaveUseCase {
	return &leaveUseCase{db, repo, employeeRepo, fcmUseCase, sseHub}
}

type SubmitLeaveRequest struct {
	LeaveTypeID   string `json:"leave_type_id" form:"leave_type_id" binding:"required"`
	StartDate     string `json:"start_date" form:"start_date" binding:"required"`
	EndDate       string `json:"end_date" form:"end_date" binding:"required"`
	Reason        string `json:"reason" form:"reason" binding:"required"`
	AttachmentURL string `json:"attachment_url" form:"attachment_url"`
}

type LeaveBalanceResponse struct {
	LeaveTypeID   string `json:"leave_type_id"`
	LeaveTypeName string `json:"leave_type_name"`
	IsPaid        bool   `json:"is_paid"`
	RequiresQuota bool   `json:"requires_quota"`
	Total         int    `json:"total"`
	Used          int    `json:"used"`
	Remaining     int    `json:"remaining"`
}

type ApproveLeaveRequest struct {
	Status       string `json:"status" binding:"required,oneof=APPROVED REJECTED approved rejected"`
	RejectReason string `json:"reject_reason"`
}

type AdminLeaveRequest struct {
	EmployeeID    string `json:"employee_id" form:"employee_id" binding:"required"`
	LeaveTypeID   string `json:"leave_type_id" form:"leave_type_id" binding:"required"`
	StartDate     string `json:"start_date" form:"start_date" binding:"required"`
	EndDate       string `json:"end_date" form:"end_date" binding:"required"`
	Reason        string `json:"reason" form:"reason"`
	AttachmentURL string `json:"attachment_url" form:"attachment_url"`
	Status        string `json:"status" form:"status"` // PENDING, APPROVED, REJECTED
}

type AdminUpdateBalanceRequest struct {
	EmployeeID   string `json:"employee_id" binding:"required"`
	LeaveTypeID  string `json:"leave_type_id" binding:"required"`
	Year         int    `json:"year" binding:"required"`
	BalanceTotal int    `json:"balance_total"`
	BalanceUsed  int    `json:"balance_used"`
}

func (u *leaveUseCase) SubmitLeave(userID uuid.UUID, req SubmitLeaveRequest) error {
	employee, err := u.employeeRepo.FindByUserID(userID)
	if err != nil {
		return err
	}
	if employee == nil {
		return errors.New("employee record not found")
	}

	layout := "2006-01-02"
	start, err := time.Parse(layout, req.StartDate)
	if err != nil {
		return errors.New("invalid start date format, expected YYYY-MM-DD")
	}
	end, err := time.Parse(layout, req.EndDate)
	if err != nil {
		return errors.New("invalid end date format, expected YYYY-MM-DD")
	}

	if end.Before(start) {
		return errors.New("end date cannot be before start date")
	}

	// Calculate total days
	duration := end.Sub(start)
	totalDays := int(duration.Hours()/24) + 1

	leaveTypeID, err := uuid.Parse(req.LeaveTypeID)
	if err != nil {
		return errors.New("invalid leave_type_id format")
	}

	leaveType, err := u.repo.GetLeaveTypeByID(leaveTypeID)
	if err != nil {
		return errors.New("leave type not found")
	}

	// VALIDATION: Sakit must have attachment
	if strings.Contains(strings.ToLower(leaveType.Name), "sakit") && req.AttachmentURL == "" {
		return errors.New("Surat izin dokter wajib dilampirkan untuk izin sakit")
	}

	// VALIDATION: Annual Leave monthly limit (max 1 day per month)
	isAnnualLeave := strings.Contains(strings.ToLower(leaveType.Name), "annual") || strings.Contains(strings.ToLower(leaveType.Name), "tahunan")
	if isAnnualLeave {
		// Divide request into monthly buckets
		tempStart := start
		for tempStart.Before(end) || tempStart.Equal(end) {
			year, month, _ := tempStart.Date()
			monthStart := time.Date(year, month, 1, 0, 0, 0, 0, time.Local)
			monthEnd := monthStart.AddDate(0, 1, -1)

			// Calculate days in THIS month for THIS request
			overlapStart := tempStart
			overlapEnd := end
			if overlapEnd.After(monthEnd) {
				overlapEnd = monthEnd
			}
			requestedDaysInMonth := int(overlapEnd.Sub(overlapStart).Hours()/24) + 1

			// Check existing usage in this month
			existingDays, err := u.repo.CountMonthlyLeaveDays(employee.ID, leaveType.ID, int(month), year, uuid.Nil)
			if err != nil {
				return err
			}

			if existingDays+requestedDaysInMonth > 1 {
				return errors.New("Maksimal pengambilan cuti tahunan adalah 1 hari per bulan. Anda sudah memiliki pengajuan atau jatah yang digunakan di bulan " + month.String())
			}

			// Move to start of next month
			tempStart = monthEnd.AddDate(0, 0, 1)
		}
	}

	var balanceToUpdate *domain.LeaveBalance
	currentYear := time.Now().Year()

	if leaveType.RequiresQuota {
		balance, err := u.repo.GetLeaveBalance(employee.ID, leaveType.ID, currentYear)
		if err != nil {
			// If balance record not found, initialize it with default quota from leave type
			balance = &domain.LeaveBalance{
				EmployeeID:   employee.ID,
				LeaveTypeID:  leaveType.ID,
				Year:         currentYear,
				BalanceTotal: leaveType.DefaultQuota,
				BalanceUsed:  0,
			}
		}

		remaining := balance.BalanceTotal - balance.BalanceUsed
		if remaining < totalDays {
			return errors.New("insufficient leave balance")
		}

		balance.BalanceUsed += totalDays
		balanceToUpdate = balance
	}

	leave := &domain.LeaveRequest{
		EmployeeID:    employee.ID,
		LeaveTypeID:   leaveType.ID,
		StartDate:     start,
		EndDate:       end,
		Reason:        req.Reason,
		AttachmentURL: req.AttachmentURL,
		Status:        "PENDING",
	}

	err = u.repo.CreateLeaveRequestWithBalance(leave, balanceToUpdate)
	if err == nil && u.sseHub != nil {
		go func() {
			event := sse.NotificationEvent{
				ID:        leave.ID.String(),
				Type:      "leave",
				Title:     "Pengajuan Cuti Baru",
				Message:   fmt.Sprintf("%s %s mengajukan cuti %s untuk tanggal %s s/d %s", employee.FirstName, employee.LastName, leaveType.Name, start.Format("2006-01-02"), end.Format("2006-01-02")),
				CompanyID: employee.CompanyID.String(),
				CreatedAt: time.Now().Format("2006-01-02 15:04:05"),
				Data: map[string]interface{}{
					"leave_id": leave.ID.String(),
					"employee": fmt.Sprintf("%s %s", employee.FirstName, employee.LastName),
				},
			}
			_ = u.sseHub.PublishNotification(context.Background(), event)
		}()
	}
	return err
}

func (u *leaveUseCase) UpdateMyLeave(leaveID, userID uuid.UUID, req SubmitLeaveRequest) error {
	employee, err := u.employeeRepo.FindByUserID(userID)
	if err != nil {
		return err
	}
	if employee == nil {
		return errors.New("employee record not found")
	}

	leave, err := u.repo.FindByID(leaveID)
	if err != nil || leave == nil {
		return errors.New("leave request not found")
	}

	if leave.EmployeeID != employee.ID {
		return errors.New("you can only update your own leave requests")
	}

	if leave.Status != "PENDING" {
		return errors.New("only pending leave requests can be updated")
	}

	layout := "2006-01-02"
	newStart, err := time.Parse(layout, req.StartDate)
	if err != nil {
		return errors.New("invalid start date format")
	}
	newEnd, err := time.Parse(layout, req.EndDate)
	if err != nil {
		return errors.New("invalid end date format")
	}
	if newEnd.Before(newStart) {
		return errors.New("end date cannot be before start date")
	}

	newDuration := newEnd.Sub(newStart)
	newTotalDays := int(newDuration.Hours()/24) + 1

	newLeaveTypeID, _ := uuid.Parse(req.LeaveTypeID)
	newLeaveType, _ := u.repo.GetLeaveTypeByID(newLeaveTypeID)
	if newLeaveType == nil {
		return errors.New("leave type not found")
	}

	// VALIDATION: Annual Leave monthly limit (max 1 day per month)
	isAnnualLeave := strings.Contains(strings.ToLower(newLeaveType.Name), "annual") || strings.Contains(strings.ToLower(newLeaveType.Name), "tahunan")
	if isAnnualLeave {
		tempStart := newStart
		for tempStart.Before(newEnd) || tempStart.Equal(newEnd) {
			y, m, _ := tempStart.Date()
			monthStart := time.Date(y, m, 1, 0, 0, 0, 0, time.Local)
			monthEnd := monthStart.AddDate(0, 1, -1)

			overlapStart := tempStart
			overlapEnd := newEnd
			if overlapEnd.After(monthEnd) {
				overlapEnd = monthEnd
			}
			requestedDaysInMonth := int(overlapEnd.Sub(overlapStart).Hours()/24) + 1

			// Exclude the current leave request from the count
			existingDays, err := u.repo.CountMonthlyLeaveDays(employee.ID, newLeaveType.ID, int(m), y, leaveID)
			if err != nil {
				return err
			}

			if existingDays+requestedDaysInMonth > 1 {
				return errors.New("Maksimal pengambilan cuti tahunan adalah 1 hari per bulan. Anda sudah memiliki pengajuan atau jatah yang digunakan di bulan " + m.String())
			}
			tempStart = monthEnd.AddDate(0, 0, 1)
		}
	}

	var balanceToUpdate *domain.LeaveBalance
	currentYear := time.Now().Year()

	if leave.LeaveTypeID != newLeaveTypeID || !leave.StartDate.Equal(newStart) || !leave.EndDate.Equal(newEnd) {
		oldLeaveType, _ := u.repo.GetLeaveTypeByID(leave.LeaveTypeID)
		if oldLeaveType != nil && oldLeaveType.RequiresQuota {
			oldDuration := leave.EndDate.Sub(leave.StartDate)
			oldTotalDays := int(oldDuration.Hours()/24) + 1
			balance, err := u.repo.GetLeaveBalance(employee.ID, leave.LeaveTypeID, leave.StartDate.Year())
			if err == nil {
				balance.BalanceUsed -= oldTotalDays
				balanceToUpdate = balance
			}
		}

		if newLeaveType.RequiresQuota {
			balance := balanceToUpdate
			if balance == nil || balance.LeaveTypeID != newLeaveTypeID {
				balance, err = u.repo.GetLeaveBalance(employee.ID, newLeaveTypeID, currentYear)
				if err != nil {
					// Auto-initialize balance record if missing during update
					balance = &domain.LeaveBalance{
						EmployeeID:   employee.ID,
						LeaveTypeID:  newLeaveTypeID,
						Year:         currentYear,
						BalanceTotal: newLeaveType.DefaultQuota,
						BalanceUsed:  0,
					}
				}
			}

			remaining := (balance.BalanceTotal - balance.BalanceUsed)
			if remaining < newTotalDays {
				return errors.New("insufficient leave balance")
			}
			balance.BalanceUsed += newTotalDays
			balanceToUpdate = balance
		}
	}

	leave.LeaveTypeID = newLeaveTypeID
	leave.StartDate = newStart
	leave.EndDate = newEnd
	leave.Reason = req.Reason
	if req.AttachmentURL != "" {
		leave.AttachmentURL = req.AttachmentURL
	}

	return u.repo.UpdateLeaveRequestStatus(leave, balanceToUpdate)
}

func (u *leaveUseCase) DeleteMyLeave(leaveID, userID uuid.UUID) error {
	employee, err := u.employeeRepo.FindByUserID(userID)
	if err != nil {
		return err
	}
	if employee == nil {
		return errors.New("employee record not found")
	}

	leave, err := u.repo.FindByID(leaveID)
	if err != nil || leave == nil {
		return errors.New("leave request not found")
	}

	if leave.EmployeeID != employee.ID {
		return errors.New("you can only delete your own leave requests")
	}

	if leave.Status != "PENDING" {
		return errors.New("only pending leave requests can be deleted")
	}

	var balanceToRefund *domain.LeaveBalance
	leaveType, _ := u.repo.GetLeaveTypeByID(leave.LeaveTypeID)
	if leaveType != nil && leaveType.RequiresQuota {
		duration := leave.EndDate.Sub(leave.StartDate)
		totalDays := int(duration.Hours()/24) + 1

		currentYear := leave.StartDate.Year()
		balance, err := u.repo.GetLeaveBalance(leave.EmployeeID, leave.LeaveTypeID, currentYear)
		if err == nil {
			balance.BalanceUsed -= totalDays
			balanceToRefund = balance
		}
	}

	leave.Status = "CANCELLED"
	return u.repo.UpdateLeaveRequestStatus(leave, balanceToRefund)
}

func (u *leaveUseCase) AdminCreateLeave(req AdminLeaveRequest) error {
	layout := "2006-01-02"
	start, err := time.Parse(layout, req.StartDate)
	if err != nil {
		return errors.New("invalid start_date format")
	}
	end, err := time.Parse(layout, req.EndDate)
	if err != nil {
		return errors.New("invalid end_date format")
	}
	if end.Before(start) {
		return errors.New("end date cannot be before start date")
	}

	employeeID, err := uuid.Parse(req.EmployeeID)
	if err != nil {
		return errors.New("invalid employee_id")
	}
	leaveTypeID, err := uuid.Parse(req.LeaveTypeID)
	if err != nil {
		return errors.New("invalid leave_type_id")
	}

	leave := &domain.LeaveRequest{
		EmployeeID:    employeeID,
		LeaveTypeID:   leaveTypeID,
		StartDate:     start,
		EndDate:       end,
		Reason:        req.Reason,
		AttachmentURL: req.AttachmentURL,
		Status:        "PENDING",
	}
	return u.repo.Create(leave)
}

func (u *leaveUseCase) AdminUpdateLeave(leaveID uuid.UUID, req AdminLeaveRequest) error {
	leave, err := u.repo.FindByID(leaveID)
	if err != nil {
		return err
	}
	if leave == nil {
		return errors.New("leave request not found")
	}

	layout := "2006-01-02"
	start, err := time.Parse(layout, req.StartDate)
	if err != nil {
		return errors.New("invalid start_date format")
	}
	end, err := time.Parse(layout, req.EndDate)
	if err != nil {
		return errors.New("invalid end_date format")
	}
	if end.Before(start) {
		return errors.New("end date cannot be before start date")
	}
	leaveTypeID, err := uuid.Parse(req.LeaveTypeID)
	if err != nil {
		return errors.New("invalid leave_type_id")
	}
	employeeID, err := uuid.Parse(req.EmployeeID)
	if err != nil {
		return errors.New("invalid employee_id")
	}
	leave.StartDate = start
	leave.EndDate = end
	leave.LeaveTypeID = leaveTypeID
	leave.EmployeeID = employeeID
	leave.Reason = req.Reason
	if req.AttachmentURL != "" {
		leave.AttachmentURL = req.AttachmentURL
	}
	if req.Status != "" {
		leave.Status = req.Status
	}
	return u.repo.Update(leave)
}

func (u *leaveUseCase) AdminDeleteLeave(leaveID uuid.UUID) error {
	return u.repo.Delete(leaveID)
}

func (u *leaveUseCase) GetMyLeaves(userID uuid.UUID, page, limit int) ([]domain.LeaveRequest, error) {
	employee, err := u.employeeRepo.FindByUserID(userID)
	if err != nil {
		return nil, err
	}
	if employee == nil {
		return nil, errors.New("employee record not found")
	}

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	offset := (page - 1) * limit

	return u.repo.FindByUserID(employee.ID, limit, offset)
}

func (u *leaveUseCase) GetMyBalances(userID uuid.UUID) ([]LeaveBalanceResponse, error) {
	employee, err := u.employeeRepo.FindByUserID(userID)
	if err != nil {
		return nil, err
	}
	if employee == nil {
		return nil, errors.New("employee record not found")
	}

	currentYear := time.Now().Year()

	if employee.CompanyID == nil {
		log.Printf("Warning: Employee %s has no company assigned", employee.ID)
		return []LeaveBalanceResponse{}, nil
	}
	leaveTypes, err := u.repo.GetLeaveTypesByCompany(*employee.CompanyID)
	if err != nil {
		return nil, err
	}

	// Fail-safe: Ensure default Izin types are ALWAYS seeded for this company
	// Removing the hasIzin check because if a company had a custom non-quota leave, it would permanently block these default types.
	izinTypes := []domain.LeaveType{
		{CompanyID: employee.CompanyID, Name: "Izin Sakit", IsPaid: true, RequiresQuota: false, DefaultQuota: 0},
		{CompanyID: employee.CompanyID, Name: "Izin Terkena Musibah", IsPaid: true, RequiresQuota: false, DefaultQuota: 0},
		{CompanyID: employee.CompanyID, Name: "Izin Lainnya", IsPaid: false, RequiresQuota: false, DefaultQuota: 0},
	}

	for _, it := range izinTypes {
		tempIT := it // Create a copy for the iteration
		// FirstOrCreate will find existing or create new. Either way, we force the correct attributes.
		if err := u.db.Where("company_id = ? AND name = ?", tempIT.CompanyID, tempIT.Name).FirstOrCreate(&tempIT).Error; err != nil {
			log.Printf("Warning: Failed to seed/find leave type %s: %v", tempIT.Name, err)
			continue
		}

		// Ensure zero-values (false, 0) are correctly saved to DB overriding GORM defaults if tampered
		if err := u.db.Model(&tempIT).Updates(map[string]interface{}{
			"is_paid":        it.IsPaid,
			"requires_quota": it.RequiresQuota,
			"default_quota":  it.DefaultQuota,
		}).Error; err != nil {
			log.Printf("Warning: Failed to update leave type %s attributes: %v", tempIT.Name, err)
		}
	}

	// ALWAYS RE-FETCH after seeding to include the new types and updated parameters
	leaveTypes, err = u.repo.GetLeaveTypesByCompany(*employee.CompanyID)
	if err != nil {
		log.Printf("Error: Failed to re-fetch leave types after seeding: %v", err)
		return nil, err
	}

	// 2. Get existing balances
	balances, err := u.repo.GetAllBalancesByEmployee(employee.ID, currentYear)
	if err != nil {
		return nil, err
	}

	// 3. Create map of existing leave type IDs
	existingTypeMap := make(map[uuid.UUID]*domain.LeaveBalance)
	for _, b := range balances {
		existingTypeMap[b.LeaveTypeID] = b
	}

	// 4. Auto-initialize missing balances
	for _, lt := range leaveTypes {
		if _, exists := existingTypeMap[lt.ID]; !exists {
			newBalance := &domain.LeaveBalance{
				EmployeeID:   employee.ID,
				LeaveTypeID:  lt.ID,
				Year:         currentYear,
				BalanceTotal: lt.DefaultQuota,
				BalanceUsed:  0,
			}
			err := u.repo.SaveBalance(newBalance)
			// Always append to balances so the UI can construct requests.
			// Non-quota leave types (Izin) must be available even if the balance row creation had a DB issue.
			newBalance.LeaveType = &lt
			balances = append(balances, newBalance)
			if err != nil {
				// Log the error silently
			}
		}
	}

	var res []LeaveBalanceResponse
	for _, b := range balances {
		name := ""
		isPaid := true
		requiresQuota := true
		if b.LeaveType != nil {
			name = b.LeaveType.Name
			isPaid = b.LeaveType.IsPaid
			requiresQuota = b.LeaveType.RequiresQuota
		}
		res = append(res, LeaveBalanceResponse{
			LeaveTypeID:   b.LeaveTypeID.String(),
			LeaveTypeName: name,
			IsPaid:        isPaid,
			RequiresQuota: requiresQuota,
			Total:         b.BalanceTotal,
			Used:          b.BalanceUsed,
			Remaining:     b.BalanceTotal - b.BalanceUsed,
		})
	}
	if res == nil {
		res = []LeaveBalanceResponse{}
	}
	return res, nil
}

func (u *leaveUseCase) GetAllLeaveBalances() ([]*domain.LeaveBalance, error) {
	return u.repo.GetAllBalances()
}

func (u *leaveUseCase) GetAllLeaves(status, search string, page, limit int, branchIDs []uuid.UUID) ([]domain.LeaveRequest, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	offset := (page - 1) * limit

	return u.repo.FindAll(status, search, limit, offset, branchIDs)
}

func (u *leaveUseCase) ApproveReturnLeave(leaveID, approverID uuid.UUID, req ApproveLeaveRequest) error {
	req.Status = strings.ToUpper(req.Status)
	leave, err := u.repo.FindByID(leaveID)
	if err != nil {
		return err
	}
	if leave == nil {
		return errors.New("leave request not found")
	}

	if leave.Status != "PENDING" {
		return errors.New("leave request is not in pending state")
	}

	if req.Status == "REJECTED" && len(req.RejectReason) == 0 {
		return errors.New("reject reason is required when status is REJECTED")
	}

	var balanceToRefund *domain.LeaveBalance

	if req.Status == "REJECTED" {
		leaveType, _ := u.repo.GetLeaveTypeByID(leave.LeaveTypeID)
		if leaveType != nil && leaveType.RequiresQuota {
			duration := leave.EndDate.Sub(leave.StartDate)
			totalDays := int(duration.Hours()/24) + 1

			currentYear := leave.StartDate.Year()
			balance, err := u.repo.GetLeaveBalance(leave.EmployeeID, leave.LeaveTypeID, currentYear)
			if err == nil {
				balance.BalanceUsed -= totalDays
				balanceToRefund = balance
			}
		}
		leave.RejectReason = &req.RejectReason
	}

	approverEmp, err := u.employeeRepo.FindByUserID(approverID)
	if err != nil {
		return errors.New("failed to resolve approver identity")
	}
	if approverEmp == nil {
		return errors.New("approver employee record not found")
	}

	leave.Status = req.Status
	leave.ApprovedBy = &approverEmp.ID

	err = u.repo.UpdateLeaveRequestStatus(leave, balanceToRefund)
	if err != nil {
		return err
	}

	// Send push notification to the employee in the background
	go func() {
		emp, err := u.employeeRepo.FindByID(leave.EmployeeID)
		if err == nil && emp != nil && emp.UserID != uuid.Nil {
			leaveTypeName := "Cuti/Izin"
			if leave.LeaveType != nil && leave.LeaveType.Name != "" {
				leaveTypeName = leave.LeaveType.Name
			}

			dateRange := leave.StartDate.Format("02-01-2006")
			if !leave.StartDate.Equal(leave.EndDate) {
				dateRange += " s/d " + leave.EndDate.Format("02-01-2006")
			}

			approverName := approverEmp.FirstName
			if approverEmp.LastName != "" {
				approverName += " " + approverEmp.LastName
			}

			var title, body string
			notificationData := map[string]string{
				"leave_id":      leave.ID.String(),
				"status":        req.Status,
				"leave_type":    leaveTypeName,
				"start_date":    leave.StartDate.Format("2006-01-02"),
				"end_date":      leave.EndDate.Format("2006-01-02"),
				"approver_name": approverName,
			}

			if req.Status == "APPROVED" {
				title = fmt.Sprintf("Pengajuan %s Disetujui ✅", leaveTypeName)
				body = fmt.Sprintf("Pengajuan %s Anda untuk tanggal %s telah disetujui oleh %s.", leaveTypeName, dateRange, approverName)
			} else {
				title = fmt.Sprintf("Pengajuan %s Ditolak ❌", leaveTypeName)
				body = fmt.Sprintf("Pengajuan %s Anda untuk tanggal %s ditolak oleh %s. Alasan: %s", leaveTypeName, dateRange, approverName, req.RejectReason)
				notificationData["reject_reason"] = req.RejectReason
			}

			_ = u.fcmUseCase.SendLeaveNotification(emp.UserID, title, body, notificationData)
		}
	}()

	return nil
}

func (u *leaveUseCase) AdminUpdateBalance(req AdminUpdateBalanceRequest) error {
	empID, err := uuid.Parse(req.EmployeeID)
	if err != nil {
		return errors.New("invalid employee_id")
	}
	ltID, err := uuid.Parse(req.LeaveTypeID)
	if err != nil {
		return errors.New("invalid leave_type_id")
	}

	balance, err := u.repo.GetLeaveBalance(empID, ltID, req.Year)
	if err != nil {
		// Create new if not exists
		balance = &domain.LeaveBalance{
			EmployeeID:   empID,
			LeaveTypeID:  ltID,
			Year:         req.Year,
			BalanceTotal: req.BalanceTotal,
			BalanceUsed:  req.BalanceUsed,
		}
	} else {
		balance.BalanceTotal = req.BalanceTotal
		balance.BalanceUsed = req.BalanceUsed
	}

	return u.repo.SaveBalance(balance)
}

func (u *leaveUseCase) AdminDeleteBalance(req AdminUpdateBalanceRequest) error {
	empID, err := uuid.Parse(req.EmployeeID)
	if err != nil {
		return errors.New("invalid employee_id")
	}
	ltID, err := uuid.Parse(req.LeaveTypeID)
	if err != nil {
		return errors.New("invalid leave_type_id")
	}

	return u.repo.DeleteLeaveBalance(empID, ltID, req.Year)
}
