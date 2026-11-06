package usecase

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"gorm.io/gorm"
)

type EmployeeWarningUsecase struct {
	db         *gorm.DB
	fcmUsecase FcmUsecase
}

func NewEmployeeWarningUsecase(db *gorm.DB, fcmUsecase FcmUsecase) *EmployeeWarningUsecase {
	return &EmployeeWarningUsecase{
		db:         db,
		fcmUsecase: fcmUsecase,
	}
}

// ProblematicEmployee represents the DTO for the frontend
type ProblematicEmployee struct {
	ID           uuid.UUID `json:"id"`
	EmployeeID   string    `json:"employee_id"`
	Name         string    `json:"name"`
	Role         string    `json:"role"`
	Branch       string    `json:"branch"`
	Issue        string    `json:"issue"`
	LastActivity string    `json:"last_activity"`
	Status       string    `json:"status"`
	HandledBy    *string   `json:"handled_by"`
}

func (u *EmployeeWarningUsecase) GetProblematicEmployees() ([]ProblematicEmployee, error) {
	// Karyawan bermasalah = Sales (punya role sales) yang AKTIF dan tidak ada sales_visits dan sales_transactions di 2 hari terakhir.
	twoDaysAgo := time.Now().AddDate(0, 0, -2)

	var employees []domain.Employee
	err := u.db.Preload("JobPosition").Preload("Branch").
		Joins("JOIN job_positions jp ON jp.id = employees.job_position_id").
		Joins("JOIN users u ON u.id = employees.user_id").
		Where("jp.title ILIKE ?", "%Sales%").
		Where("u.is_active = TRUE").
		Where("COALESCE(employees.employment_status, 'Active') NOT IN ('Inactive', 'INACTIVE', 'Resigned', 'RESIGNED', 'Terminated', 'TERMINATED')").
		Find(&employees).Error

	if err != nil {
		return nil, err
	}

	var results []ProblematicEmployee

	for _, emp := range employees {
		// Cek visit 2 hari terakhir
		var visitCount int64
		u.db.Model(&domain.SalesVisit{}).
			Where("employee_id = ? AND check_in_time >= ?", emp.ID, twoDaysAgo).
			Count(&visitCount)

		// Cek transaksi 2 hari terakhir
		var trxCount int64
		u.db.Model(&domain.SalesTransaction{}).
			Where("employee_id = ? AND created_at >= ?", emp.ID, twoDaysAgo).
			Count(&trxCount)

		if visitCount == 0 && trxCount == 0 {
			// Karyawan ini bermasalah

			// Cek apakah sudah ditangani di tabel employee_warnings (recent warning in last 2 days)
			var warning domain.EmployeeWarning
			warningFound := true
			err := u.db.Preload("HandledBy").
				Where("employee_id = ? AND created_at >= ?", emp.ID, twoDaysAgo).
				Order("created_at desc").
				First(&warning).Error

			if errors.Is(err, gorm.ErrRecordNotFound) {
				warningFound = false
			}

			status := "pending"
			var handledBy *string
			var warningID uuid.UUID

			if warningFound {
				status = warning.WarningType // "teguran", "sp1"
				warningID = warning.ID
				if warning.HandledBy != nil {
					name := warning.HandledBy.Email
					handledBy = &name
				} else {
					defaultName := "Admin HRD"
					handledBy = &defaultName
				}
			} else {
				// We still need an ID for the row in UI, just use employee ID for pending
				warningID = emp.ID
			}

			branchName := "-"
			if emp.Branch != nil {
				branchName = emp.Branch.Name
			}

			roleName := "-"
			if emp.JobPosition != nil {
				roleName = emp.JobPosition.Title
			}

			results = append(results, ProblematicEmployee{
				ID:           warningID, // Can be warning ID or Employee ID depending on state
				EmployeeID:   emp.EmployeeIDNumber,
				Name:         emp.FirstName + " " + emp.LastName,
				Role:         roleName,
				Branch:       branchName,
				Issue:        "Tidak ada kunjungan & transaksi dalam 2 hari terakhir berturut-turut",
				LastActivity: twoDaysAgo.Format("2006-01-02"),
				Status:       status,
				HandledBy:    handledBy,
			})
		}
	}

	return results, nil
}

func (u *EmployeeWarningUsecase) SubmitWarning(employeeIdNumber string, warningType string, notes string, handledByUserID uuid.UUID) error {
	// Find employee by employee_id_number
	var emp domain.Employee
	err := u.db.Where("employee_id_number = ?", employeeIdNumber).First(&emp).Error
	if err != nil {
		return errors.New("employee not found")
	}

	warning := domain.EmployeeWarning{
		EmployeeID:       emp.ID,
		WarningType:      warningType,
		IssueDescription: notes,
		Status:           "handled",
		HandledByID:      &handledByUserID,
	}

	if err := u.db.Create(&warning).Error; err != nil {
		return err
	}

	// Send FCM Notification to User
	displayType := warningType
	switch warningType {
	case "cs_followup":
		displayType = "Tindak Lanjut CS"
	case "spv_followup":
		displayType = "Tindak Lanjut SPV Sales"
	case "teguran":
		displayType = "Teguran"
	case "sp1":
		displayType = "SP 1"
	case "sp2":
		displayType = "SP 2"
	case "sp3":
		displayType = "SP 3"
	}

	title := "Informasi HRD: " + displayType
	body := "Anda menerima " + displayType + " dari HRD / Manajemen. Catatan: " + notes

	warningData := map[string]string{
		"warning_id":   warning.ID.String(),
		"warning_type": warningType,
		"notes":        notes,
		"created_at":   warning.CreatedAt.Format(time.RFC3339),
	}

	// Execute FCM Notification asynchronously to prevent blocking the HTTP response
	go func(uID uuid.UUID, t, b string, wData map[string]string) {
		_ = u.fcmUsecase.SendWarningNotification(uID, t, b, wData)
	}(emp.UserID, title, body, warningData)

	return nil
}

func (u *EmployeeWarningUsecase) GetActiveWarning(employeeID uuid.UUID) (*domain.EmployeeWarning, error) {
	var warning domain.EmployeeWarning
	// Active warning if is_active is true
	err := u.db.Where("employee_id = ? AND is_active = ?", employeeID, true).
		Order("created_at desc").
		First(&warning).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &warning, nil
}

// WarningHistoryResponse DTO
type WarningHistoryResponse struct {
	ID          uuid.UUID `json:"id"`
	EmployeeID  string    `json:"employee_id"`
	Name        string    `json:"name"`
	Role        string    `json:"role"`
	Branch      string    `json:"branch"`
	WarningType string    `json:"warning_type"`
	Notes       string    `json:"notes"`
	IsActive    bool      `json:"is_active"`
	HandledBy   *string   `json:"handled_by"`
	CreatedAt   time.Time `json:"created_at"`
}

func (u *EmployeeWarningUsecase) GetAllWarnings() ([]WarningHistoryResponse, error) {
	var warnings []domain.EmployeeWarning
	err := u.db.Preload("Employee.JobPosition").Preload("Employee.Branch").Preload("HandledBy").
		Order("created_at desc").
		Find(&warnings).Error
	if err != nil {
		return nil, err
	}

	var results []WarningHistoryResponse
	for _, w := range warnings {
		var handledBy *string
		if w.HandledBy != nil {
			name := w.HandledBy.Email
			handledBy = &name
		} else {
			defaultName := "Admin HRD"
			handledBy = &defaultName
		}

		branchName := "-"
		if w.Employee != nil && w.Employee.Branch != nil {
			branchName = w.Employee.Branch.Name
		}

		roleName := "-"
		if w.Employee != nil && w.Employee.JobPosition != nil {
			roleName = w.Employee.JobPosition.Title
		}

		empID := "-"
		name := "-"
		if w.Employee != nil {
			empID = w.Employee.EmployeeIDNumber
			name = w.Employee.FirstName + " " + w.Employee.LastName
		}

		results = append(results, WarningHistoryResponse{
			ID:          w.ID,
			EmployeeID:  empID,
			Name:        name,
			Role:        roleName,
			Branch:      branchName,
			WarningType: w.WarningType,
			Notes:       w.IssueDescription,
			IsActive:    w.IsActive,
			HandledBy:   handledBy,
			CreatedAt:   w.CreatedAt,
		})
	}

	return results, nil
}

func (u *EmployeeWarningUsecase) UpdateWarning(warningID uuid.UUID, warningType string, notes string) error {
	return u.db.Model(&domain.EmployeeWarning{}).
		Where("id = ?", warningID).
		Updates(map[string]interface{}{
			"warning_type":      warningType,
			"issue_description": notes,
		}).Error
}

func (u *EmployeeWarningUsecase) DeleteWarning(warningID uuid.UUID) error {
	return u.db.Where("id = ?", warningID).Delete(&domain.EmployeeWarning{}).Error
}

func (u *EmployeeWarningUsecase) ToggleWarningActive(warningID uuid.UUID) error {
	return u.db.Model(&domain.EmployeeWarning{}).
		Where("id = ?", warningID).
		Update("is_active", gorm.Expr("NOT is_active")).Error
}
