package http

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"example.com/hris/backend/internal/repository"
)

type ApproverInfo struct {
	UserID      uuid.UUID
	EmployeeID  *uuid.UUID
	FullName    string
	JobPosition string
	Role        string
}

// GetApproverInfo extracts the approver's UserID, EmployeeID, Full Name, and Job Position Title from Gin Context or DB.
func GetApproverInfo(c *gin.Context, empRepo repository.EmployeeRepository) *ApproverInfo {
	info := &ApproverInfo{}

	// 1. Get UserID
	if uidVal, exists := c.Get("userID"); exists {
		if uid, ok := uidVal.(uuid.UUID); ok {
			info.UserID = uid
		}
	}
	if info.UserID == uuid.Nil {
		if uidStr, exists := c.Get("user_id"); exists {
			if str, ok := uidStr.(string); ok {
				if parsed, err := uuid.Parse(str); err == nil {
					info.UserID = parsed
				}
			}
		}
	}

	// 2. Get Role
	if roleVal, exists := c.Get("role"); exists {
		if r, ok := roleVal.(string); ok {
			info.Role = r
		}
	}

	// 3. Get Name from context if present
	if nameVal, exists := c.Get("name"); exists {
		if n, ok := nameVal.(string); ok {
			info.FullName = strings.TrimSpace(n)
		}
	}

	// 4. Get EmployeeID from context if present
	if empIDVal, exists := c.Get("employeeID"); exists {
		if empID, ok := empIDVal.(uuid.UUID); ok && empID != uuid.Nil {
			info.EmployeeID = &empID
		}
	} else if empIDStr, exists := c.Get("employee_id"); exists {
		if str, ok := empIDStr.(string); ok {
			if parsed, err := uuid.Parse(str); err == nil && parsed != uuid.Nil {
				info.EmployeeID = &parsed
			}
		}
	}

	// 5. Query employee repository to resolve full details and JobPosition if needed
	if empRepo != nil && info.UserID != uuid.Nil {
		var emp *domain.Employee
		var err error
		if info.EmployeeID != nil {
			emp, err = empRepo.FindByID(*info.EmployeeID)
		}
		if emp == nil || err != nil {
			emp, err = empRepo.FindByUserID(info.UserID)
		}

		if err == nil && emp != nil {
			info.EmployeeID = &emp.ID
			fullName := strings.TrimSpace(emp.FirstName + " " + emp.LastName)
			if fullName != "" {
				info.FullName = fullName
			}
			if emp.JobPosition != nil && emp.JobPosition.Title != "" {
				info.JobPosition = emp.JobPosition.Title
			}
		}
	}

	// Fallback for JobPosition if not resolved
	if info.JobPosition == "" {
		if info.Role != "" {
			info.JobPosition = strings.ToUpper(strings.ReplaceAll(info.Role, "_", " "))
		} else {
			info.JobPosition = "ADMIN"
		}
	}

	if info.FullName == "" {
		info.FullName = "Admin / Approver"
	}

	return info
}
