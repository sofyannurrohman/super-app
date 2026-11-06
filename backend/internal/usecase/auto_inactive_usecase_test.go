package usecase

import (
	"testing"

	"example.com/hris/backend/internal/domain"
)

func TestIsSalesEmployee(t *testing.T) {
	tests := []struct {
		name     string
		employee *domain.Employee
		expected bool
	}{
		{
			name:     "Nil Employee",
			employee: nil,
			expected: false,
		},
		{
			name: "Sales TO by JobPosition",
			employee: &domain.Employee{
				JobPosition: &domain.JobPosition{
					Title: "Sales Take Order (TO)",
				},
			},
			expected: true,
		},
		{
			name: "Sales MT by JobPosition",
			employee: &domain.Employee{
				JobPosition: &domain.JobPosition{
					Title: "Sales Modern Trade (MT)",
				},
			},
			expected: true,
		},
		{
			name: "Sales Motoris by JobPosition",
			employee: &domain.Employee{
				JobPosition: &domain.JobPosition{
					Title: "Motoris Luar Kota",
				},
			},
			expected: true,
		},
		{
			name: "Sales by Department Name",
			employee: &domain.Employee{
				JobPosition: &domain.JobPosition{
					Title: "Staff Lapangan",
				},
				Department: &domain.Department{
					Name: "Departemen Sales & Marketing",
				},
			},
			expected: true,
		},
		{
			name: "Supervisor Sales by User Role",
			employee: &domain.Employee{
				User: &domain.User{
					Role: domain.RoleSupervisorSales,
				},
			},
			expected: true,
		},
		{
			name: "Regular Staff (Accounting/HR)",
			employee: &domain.Employee{
				JobPosition: &domain.JobPosition{
					Title: "Staff Accounting",
				},
				Department: &domain.Department{
					Name: "Finance & HR",
				},
				User: &domain.User{
					Role: domain.RoleEmployee,
				},
			},
			expected: false,
		},
		{
			name: "Auditor/Gudang with sales substring (e.g. Auditor Sales)",
			employee: &domain.Employee{
				JobPosition: &domain.JobPosition{
					Title: "Auditor Sales",
				},
				Department: &domain.Department{
					Name: "Internal Audit",
				},
			},
			expected: false, // filtered out by auditor keyword
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isSalesEmployee(tt.employee)
			if result != tt.expected {
				t.Errorf("isSalesEmployee() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestIsExemptFromAutoInactive(t *testing.T) {
	tests := []struct {
		name     string
		employee *domain.Employee
		expected bool
	}{
		{
			name:     "Nil Employee",
			employee: nil,
			expected: false,
		},
		{
			name: "Superadmin is exempt",
			employee: &domain.Employee{
				User: &domain.User{
					Role: domain.RoleSuperAdmin,
				},
			},
			expected: true,
		},
		{
			name: "HR Admin is exempt",
			employee: &domain.Employee{
				User: &domain.User{
					Role: domain.RoleHRAdmin,
				},
			},
			expected: true,
		},
		{
			name: "CanAbsenAnywhere user is exempt",
			employee: &domain.Employee{
				User: &domain.User{
					Role:             domain.RoleEmployee,
					CanAbsenAnywhere: true,
				},
			},
			expected: true,
		},
		{
			name: "Standard employee is not exempt",
			employee: &domain.Employee{
				EmploymentStatus: "Tetap",
				User: &domain.User{
					Role:             domain.RoleEmployee,
					CanAbsenAnywhere: false,
				},
			},
			expected: false,
		},
		{
			name: "Harian employee is exempt",
			employee: &domain.Employee{
				EmploymentStatus: "Harian",
				User: &domain.User{
					Role:             domain.RoleEmployee,
					CanAbsenAnywhere: false,
				},
			},
			expected: true,
		},
		{
			name: "Daily worker employee is exempt",
			employee: &domain.Employee{
				EmploymentStatus: "daily",
				User: &domain.User{
					Role:             domain.RoleEmployee,
					CanAbsenAnywhere: false,
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isExemptFromAutoInactive(tt.employee)
			if result != tt.expected {
				t.Errorf("isExemptFromAutoInactive() = %v, expected %v", result, tt.expected)
			}
		})
	}
}
