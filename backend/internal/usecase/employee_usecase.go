package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"example.com/hris/backend/internal/repository"
	"example.com/hris/backend/pkg/redisclient"
	"github.com/xuri/excelize/v2"
	"golang.org/x/crypto/bcrypt"
	"mime/multipart"
)

type EmployeeUsecase interface {
	CreateEmployee(req *CreateEmployeeRequest) error
	GetEmployees(limit int, branchIDs []uuid.UUID, jobPositionID *uuid.UUID) ([]domain.Employee, error)
	GetEmployeesPaginated(page, limit int, branchIDs []uuid.UUID, search string, isSales bool, salesType string, employmentType string, status string, jobPositionID *uuid.UUID, branchMatch ...string) ([]domain.Employee, int64, error)
	GetEmployeeByID(id uuid.UUID) (*domain.Employee, error)
	GetEmployeeByUserID(userID uuid.UUID) (*domain.Employee, error)
	UpdateEmployee(id uuid.UUID, req *UpdateEmployeeRequest) error
	UpdateProfile(userID uuid.UUID, req *UpdateProfileRequest) error
	DeleteEmployee(id uuid.UUID) error
	GetDirectory(query string, branchIDs []uuid.UUID) ([]DirectoryResponse, error)
	ImportDailyWorkers(file multipart.File) error
}

type employeeUsecase struct {
	employeeRepo repository.EmployeeRepository
	userRepo     repository.UserRepository
}

// Request Data Transfer Objects
type CreateEmployeeRequest struct {
	Name              string     `json:"name" binding:"required"`
	Email             string     `json:"email" binding:"required,email"`
	EmployeeIDNumber  string     `json:"employeeIDNumber"`
	BankName          string     `json:"bankName"`
	BankAccountNumber string     `json:"bankAccountNumber"`
	AccountHolderName string     `json:"accountHolderName"`
	Password          string     `json:"password"`
	CompanyID         *uuid.UUID `json:"companyId"`
	BranchID          *uuid.UUID `json:"branchId"`
	OperatingBranchID *uuid.UUID `json:"operatingBranchId"`
	DepartmentID      *uuid.UUID `json:"departmentId"`
	JobPositionID     *uuid.UUID `json:"jobPositionId"`
	EmploymentStatus  string     `json:"employmentStatus"`
	JoinDate          string     `json:"joinDate"` // Format: YYYY-MM-DD
	Salary            *float64   `json:"salary"`
	PhoneNumber       string     `json:"phoneNumber"`

	// New Fields
	BirthPlace                string `json:"birthPlace"`
	BirthDate                 string `json:"birthDate"`
	Gender                    string `json:"gender"`
	MaritalStatus             string `json:"maritalStatus"`
	BpjsKesehatanNumber       string `json:"bpjsKesehatanNumber"`
	BpjsKetenagakerjaanNumber string `json:"bpjsKetenagakerjaanNumber"`
	IdentityNumber            string `json:"identityNumber"`
	NpwpNumber                string `json:"npwpNumber"`

	// More Fields
	LastName           string     `json:"lastName"`
	AddressKTP         string     `json:"addressKTP"`
	AddressResidential string     `json:"addressResidential"`
	EmergencyContact   string     `json:"emergencyContact"`
	PtkpStatus         string     `json:"ptkpStatus"`
	ManagerID          *uuid.UUID `json:"managerId"`
	DefaultShiftID     *uuid.UUID `json:"defaultShiftId"`
	FaceReferenceURL   string     `json:"face_reference_url"`

	// Distributor Fields
	IsDistributor      bool       `json:"is_distributor"`
	DistributorStoreID *uuid.UUID `json:"distributor_store_id"`
}

type UpdateEmployeeRequest struct {
	Name              string     `json:"name" binding:"required"`
	Email             string     `json:"email" binding:"required,email"`
	EmployeeIDNumber  string     `json:"employeeIDNumber"`
	BankName          string     `json:"bankName"`
	BankAccountNumber string     `json:"bankAccountNumber"`
	AccountHolderName string     `json:"accountHolderName"`
	DepartmentID      *uuid.UUID `json:"departmentId"`
	JobPositionID     *uuid.UUID `json:"jobPositionId"`
	BranchID          *uuid.UUID `json:"branchId"`
	OperatingBranchID *uuid.UUID `json:"operatingBranchId"`
	EmploymentStatus  string     `json:"employmentStatus"`
	JoinDate          string     `json:"joinDate"`
	Salary            *float64   `json:"salary"`
	PhoneNumber       string     `json:"phoneNumber"`
	CompanyID         *uuid.UUID `json:"companyId"`

	// New Fields
	BirthPlace                string `json:"birthPlace"`
	BirthDate                 string `json:"birthDate"`
	Gender                    string `json:"gender"`
	MaritalStatus             string `json:"maritalStatus"`
	BpjsKesehatanNumber       string `json:"bpjsKesehatanNumber"`
	BpjsKetenagakerjaanNumber string `json:"bpjsKetenagakerjaanNumber"`
	IdentityNumber            string `json:"identityNumber"`
	NpwpNumber                string `json:"npwpNumber"`

	// More Fields
	LastName           string     `json:"lastName"`
	AddressKTP         string     `json:"addressKTP"`
	AddressResidential string     `json:"addressResidential"`
	EmergencyContact   string     `json:"emergencyContact"`
	PtkpStatus         string     `json:"ptkpStatus"`
	ManagerID          *uuid.UUID `json:"managerId"`
	DefaultShiftID     *uuid.UUID `json:"defaultShiftId"`
	FaceReferenceURL   string     `json:"face_reference_url"`

	// Distributor Fields
	IsDistributor      bool       `json:"is_distributor"`
	DistributorStoreID *uuid.UUID `json:"distributor_store_id"`
}

type UpdateProfileRequest struct {
	FirstName          string `json:"firstName"`
	LastName           string `json:"lastName"`
	BirthDate          string `json:"birthDate"` // YYYY-MM-DD
	BirthPlace         string `json:"birthPlace"`
	Gender             string `json:"gender"`
	MaritalStatus      string `json:"maritalStatus"`
	PhoneNumber        string `json:"phoneNumber"`
	AddressKTP         string `json:"addressKTP"`
	AddressResidential string `json:"addressResidential"`
	EmergencyContact   string `json:"emergencyContact"`
	IdentityNumber     string `json:"identityNumber"`
	NpwpNumber         string `json:"npwpNumber"`
	BankName           string `json:"bankName"`
	BankAccountNumber  string `json:"bankAccountNumber"`
	AccountHolderName  string `json:"accountHolderName"`
}

type DirectoryResponse struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Position    string  `json:"position"`
	Department  string  `json:"department"`
	PhoneNumber string  `json:"phone_number"`
	ProfileURL  *string `json:"profile_url"`
	BranchID    *string `json:"branch_id,omitempty"`
	BranchName  string  `json:"branch_name,omitempty"`
}

func NewEmployeeUsecase(employeeRepo repository.EmployeeRepository, userRepo repository.UserRepository) EmployeeUsecase {
	return &employeeUsecase{
		employeeRepo: employeeRepo,
		userRepo:     userRepo,
	}
}

func (u *employeeUsecase) CreateEmployee(req *CreateEmployeeRequest) error {
	// Check if user email already exists
	existingUser, _ := u.userRepo.FindByEmail(req.Email)
	if existingUser != nil {
		return errors.New("email already exists")
	}

	// Prepare Default Password if not provided
	password := req.Password
	if password == "" {
		password = os.Getenv("TRIAL_STAFF_PASSWORD")
		if password == "" {
			return errors.New("password wajib diisi")
		}
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("failed to hash password")
	}

	// Create User Record
	newUser := &domain.User{
		Email:              req.Email,
		PasswordHash:       string(hashedPassword),
		Role:               domain.RoleEmployee,
		IsActive:           true,
		DistributorStoreID: req.DistributorStoreID,
	}

	if err := u.userRepo.Create(newUser); err != nil {
		return err
	}

	// Create Associated Employee Record
	empID := req.EmployeeIDNumber
	if empID == "" {
		empID = "EMP-" + uuid.New().String()[:5]
	}

	joinDate := time.Now()
	if req.JoinDate != "" {
		if t, err := time.Parse("2006-01-02", req.JoinDate); err == nil {
			joinDate = t
		}
	}

	newEmployee := &domain.Employee{
		UserID:                    newUser.ID,
		FirstName:                 req.Name,
		BankName:                  req.BankName,
		BankAccountNumber:         req.BankAccountNumber,
		AccountHolderName:         req.AccountHolderName,
		EmployeeIDNumber:          empID,
		CompanyID:                 req.CompanyID,
		BranchID:                  req.BranchID,
		OperatingBranchID:         req.OperatingBranchID,
		DepartmentID:              req.DepartmentID,
		JobPositionID:             req.JobPositionID,
		EmploymentStatus:          req.EmploymentStatus,
		JoinDate:                  joinDate,
		Salary:                    req.Salary,
		PhoneNumber:               req.PhoneNumber,
		BirthPlace:                req.BirthPlace,
		Gender:                    req.Gender,
		MaritalStatus:             req.MaritalStatus,
		BpjsKesehatanNumber:       req.BpjsKesehatanNumber,
		BpjsKetenagakerjaanNumber: req.BpjsKetenagakerjaanNumber,
		IdentityNumber:            req.IdentityNumber,
		NpwpNumber:                req.NpwpNumber,
		LastName:                  req.LastName,
		AddressKTP:                req.AddressKTP,
		AddressResidential:        req.AddressResidential,
		EmergencyContact:          req.EmergencyContact,
		PtkpStatus:                req.PtkpStatus,
		ManagerID:                 req.ManagerID,
		DefaultShiftID:            req.DefaultShiftID,
		FaceReferenceURL:          req.FaceReferenceURL,
		IsDistributor:             req.IsDistributor,
		DistributorStoreID:        req.DistributorStoreID,
	}

	if req.BirthDate != "" {
		if t, err := time.Parse("2006-01-02", req.BirthDate); err == nil {
			newEmployee.BirthDate = &t
		}
	}

	return u.employeeRepo.Create(newEmployee)
}

func (u *employeeUsecase) GetEmployees(limit int, branchIDs []uuid.UUID, jobPositionID *uuid.UUID) ([]domain.Employee, error) {
	return u.employeeRepo.FindAll(limit, branchIDs, jobPositionID)
}

func (u *employeeUsecase) GetEmployeesPaginated(page, limit int, branchIDs []uuid.UUID, search string, isSales bool, salesType string, employmentType string, status string, jobPositionID *uuid.UUID, branchMatch ...string) ([]domain.Employee, int64, error) {
	return u.employeeRepo.FindPaginated(page, limit, branchIDs, search, isSales, salesType, employmentType, status, jobPositionID, branchMatch...)
}

func (u *employeeUsecase) GetEmployeeByID(id uuid.UUID) (*domain.Employee, error) {
	ctx := context.Background()
	cacheKey := fmt.Sprintf("employee:id:%s", id.String())

	if redisclient.Client != nil {
		if cached, err := redisclient.Client.Get(ctx, cacheKey).Result(); err == nil {
			var emp domain.Employee
			if json.Unmarshal([]byte(cached), &emp) == nil {
				return &emp, nil
			}
		}
	}

	emp, err := u.employeeRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if emp == nil {
		return nil, nil
	}

	if redisclient.Client != nil {
		if bytes, err := json.Marshal(emp); err == nil {
			_ = redisclient.Client.Set(ctx, cacheKey, bytes, 15*time.Minute).Err()
		}
	}

	return emp, nil
}

func (u *employeeUsecase) GetEmployeeByUserID(userID uuid.UUID) (*domain.Employee, error) {
	ctx := context.Background()
	cacheKey := fmt.Sprintf("employee:userid:%s", userID.String())

	if redisclient.Client != nil {
		if cached, err := redisclient.Client.Get(ctx, cacheKey).Result(); err == nil {
			var emp domain.Employee
			if json.Unmarshal([]byte(cached), &emp) == nil {
				return &emp, nil
			}
		}
	}

	emp, err := u.employeeRepo.FindByUserID(userID)
	if err != nil {
		return nil, err
	}
	if emp == nil {
		return nil, nil
	}

	if redisclient.Client != nil {
		if bytes, err := json.Marshal(emp); err == nil {
			_ = redisclient.Client.Set(ctx, cacheKey, bytes, 15*time.Minute).Err()
		}
	}

	return emp, nil
}

func invalidateEmployeeCache(ctx context.Context, empID uuid.UUID, userID uuid.UUID) {
	if redisclient.Client == nil {
		return
	}
	if empID != uuid.Nil {
		_ = redisclient.Client.Del(ctx, fmt.Sprintf("employee:id:%s", empID.String())).Err()
	}
	if userID != uuid.Nil {
		_ = redisclient.Client.Del(ctx, fmt.Sprintf("employee:userid:%s", userID.String())).Err()
	}
}

func (u *employeeUsecase) UpdateEmployee(id uuid.UUID, req *UpdateEmployeeRequest) error {
	employee, err := u.employeeRepo.FindByID(id)
	if err != nil {
		return err
	}
	if employee == nil {
		return errors.New("employee not found")
	}

	// Update User Email if changed
	if req.Email != "" && employee.User != nil && employee.User.Email != req.Email {
		existingUser, err := u.userRepo.FindByEmail(req.Email)
		if err == nil && existingUser != nil && existingUser.ID != employee.UserID {
			return errors.New("email is already registered by another user")
		}

		employee.User.Email = req.Email
		if err := u.userRepo.Update(employee.User); err != nil {
			return err
		}
	}

	// Update Fields
	employee.FirstName = req.Name
	employee.BankName = req.BankName
	employee.BankAccountNumber = req.BankAccountNumber
	employee.AccountHolderName = req.AccountHolderName
	employee.EmployeeIDNumber = req.EmployeeIDNumber
	employee.DepartmentID = req.DepartmentID
	employee.JobPositionID = req.JobPositionID
	employee.BranchID = req.BranchID
	employee.OperatingBranchID = req.OperatingBranchID
	employee.EmploymentStatus = req.EmploymentStatus
	employee.CompanyID = req.CompanyID

	if req.JoinDate != "" {
		if t, err := time.Parse("2006-01-02", req.JoinDate); err == nil {
			employee.JoinDate = t
		}
	}

	if req.Salary != nil {
		employee.Salary = req.Salary
	}

	employee.PhoneNumber = req.PhoneNumber
	employee.BirthPlace = req.BirthPlace
	employee.Gender = req.Gender
	employee.MaritalStatus = req.MaritalStatus
	employee.BpjsKesehatanNumber = req.BpjsKesehatanNumber
	employee.BpjsKetenagakerjaanNumber = req.BpjsKetenagakerjaanNumber
	employee.IdentityNumber = req.IdentityNumber
	employee.NpwpNumber = req.NpwpNumber
	employee.LastName = req.LastName
	employee.AddressKTP = req.AddressKTP
	employee.AddressResidential = req.AddressResidential
	employee.EmergencyContact = req.EmergencyContact
	employee.PtkpStatus = req.PtkpStatus
	employee.ManagerID = req.ManagerID
	employee.DefaultShiftID = req.DefaultShiftID
	employee.FaceReferenceURL = req.FaceReferenceURL
	employee.IsDistributor = req.IsDistributor
	employee.DistributorStoreID = req.DistributorStoreID

	if req.BirthDate != "" {
		if t, err := time.Parse("2006-01-02", req.BirthDate); err == nil {
			employee.BirthDate = &t
		}
	}

	// Sync User IsActive status with EmploymentStatus
	if employee.User != nil && req.EmploymentStatus != "" {
		isInactive := false
		switch strings.ToUpper(strings.TrimSpace(req.EmploymentStatus)) {
		case "INACTIVE", "RESIGNED", "TERMINATED":
			isInactive = true
		}
		newIsActive := !isInactive
		if employee.User.IsActive != newIsActive {
			employee.User.IsActive = newIsActive
			_ = u.userRepo.Update(employee.User)
		}
	}

	// Sync User DistributorStoreID
	if employee.User != nil {
		if employee.User.DistributorStoreID != req.DistributorStoreID {
			employee.User.DistributorStoreID = req.DistributorStoreID
			_ = u.userRepo.Update(employee.User)
		}
	}

	if err := u.employeeRepo.Update(employee); err != nil {
		return err
	}
	invalidateEmployeeCache(context.Background(), employee.ID, employee.UserID)
	return nil
}

func (u *employeeUsecase) UpdateProfile(userID uuid.UUID, req *UpdateProfileRequest) error {
	employee, err := u.employeeRepo.FindByUserID(userID)
	if err != nil {
		return err
	}
	if employee == nil {
		return errors.New("employee not found")
	}

	// Update Fields
	employee.FirstName = req.FirstName
	employee.LastName = req.LastName
	employee.BirthPlace = req.BirthPlace
	employee.Gender = req.Gender
	employee.MaritalStatus = req.MaritalStatus
	employee.PhoneNumber = req.PhoneNumber
	employee.AddressKTP = req.AddressKTP
	employee.AddressResidential = req.AddressResidential
	employee.EmergencyContact = req.EmergencyContact
	employee.IdentityNumber = req.IdentityNumber
	employee.NpwpNumber = req.NpwpNumber
	employee.BankName = req.BankName
	employee.BankAccountNumber = req.BankAccountNumber
	employee.AccountHolderName = req.AccountHolderName

	if req.BirthDate != "" {
		if t, err := time.Parse("2006-01-02", req.BirthDate); err == nil {
			employee.BirthDate = &t
		}
	}

	if err := u.employeeRepo.Update(employee); err != nil {
		return err
	}
	invalidateEmployeeCache(context.Background(), employee.ID, employee.UserID)
	return nil
}

func (u *employeeUsecase) DeleteEmployee(id uuid.UUID) error {
	employee, err := u.employeeRepo.FindByID(id)
	if err != nil {
		return err
	}
	if employee == nil {
		return errors.New("employee not found")
	}

	userID := employee.UserID

	// 1. Delete Employee (Cascading deletes related HR data like attendance, payroll, etc. due to DB constraints)
	if err := u.employeeRepo.Delete(id); err != nil {
		return err
	}

	// 2. Delete User account (security clean up)
	if userID != uuid.Nil {
		if err := u.userRepo.Delete(userID); err != nil {
			// We only log if user deletion fails as the main employee data is already gone
			// and this might be an orphaned employee without a user record.
			fmt.Printf("Warning: Failed to delete user record %s for employee %s: %v\n", userID, id, err)
		}
	}

	invalidateEmployeeCache(context.Background(), id, userID)
	return nil
}

func (u *employeeUsecase) GetDirectory(query string, branchIDs []uuid.UUID) ([]DirectoryResponse, error) {
	employees, err := u.employeeRepo.FindAll(0, branchIDs, nil)
	if err != nil {
		return nil, err
	}

	query = strings.ToLower(query)
	var resp []DirectoryResponse
	for _, e := range employees {
		// Filter out inactive employees
		if e.User != nil && !e.User.IsActive {
			continue
		}
		statusLower := strings.ToLower(strings.TrimSpace(e.EmploymentStatus))
		if statusLower == "inactive" ||
			statusLower == "resigned" ||
			statusLower == "terminated" ||
			statusLower == "keluar" ||
			statusLower == "nonaktif" ||
			statusLower == "non-aktif" {
			continue
		}

		fullName := strings.TrimSpace(e.FirstName + " " + e.LastName)
		deptName := ""
		if e.Department != nil {
			deptName = e.Department.Name
		}
		posName := ""
		if e.JobPosition != nil {
			posName = e.JobPosition.Title
		}

		// Filter out daily workers (pekerja harian)
		if strings.Contains(statusLower, "harian") ||
			strings.Contains(statusLower, "daily") ||
			strings.Contains(strings.ToLower(posName), "harian") ||
			strings.Contains(strings.ToLower(deptName), "harian") {
			continue
		}
		branchName := ""
		if e.Branch != nil && e.Branch.Name != "" {
			branchName = e.Branch.Name
		} else if e.OperatingBranch != nil && e.OperatingBranch.Name != "" {
			branchName = e.OperatingBranch.Name
		}

		var branchIDStr *string
		if e.BranchID != nil {
			str := e.BranchID.String()
			branchIDStr = &str
		} else if e.OperatingBranchID != nil {
			str := e.OperatingBranchID.String()
			branchIDStr = &str
		}

		var profileURL *string
		if e.FaceReferenceURL != "" {
			url := e.FaceReferenceURL
			profileURL = &url
		}

		if query == "" ||
			strings.Contains(strings.ToLower(fullName), query) ||
			strings.Contains(strings.ToLower(deptName), query) ||
			strings.Contains(strings.ToLower(posName), query) ||
			strings.Contains(strings.ToLower(branchName), query) {
			resp = append(resp, DirectoryResponse{
				ID:          e.ID.String(),
				Name:        fullName,
				Position:    posName,
				Department:  deptName,
				PhoneNumber: e.PhoneNumber,
				ProfileURL:  profileURL,
				BranchID:    branchIDStr,
				BranchName:  branchName,
			})
		}
	}
	return resp, nil
}

func (u *employeeUsecase) ImportDailyWorkers(file multipart.File) error {
	f, err := excelize.OpenReader(file)
	if err != nil {
		return errors.New("gagal membaca file excel")
	}
	defer f.Close()

	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return errors.New("file excel kosong")
	}

	rows, err := f.GetRows(sheets[0])
	if err != nil {
		return err
	}

	if len(rows) <= 1 {
		return errors.New("tidak ada data yang ditemukan")
	}

	for i, row := range rows {
		if i == 0 {
			continue // skip header
		}
		if len(row) < 3 {
			return fmt.Errorf("baris %d: data tidak lengkap (harus ada No, Nama, dan Branch ID)", i+1)
		}

		name := strings.TrimSpace(row[1])
		branchIDStr := strings.TrimSpace(row[2])
		if name == "" || branchIDStr == "" {
			return fmt.Errorf("baris %d: nama dan branch ID tidak boleh kosong", i+1)
		}

		parsedBranchID, err := uuid.Parse(branchIDStr)
		if err != nil {
			return fmt.Errorf("baris %d: format branch ID tidak valid (harus berupa UUID)", i+1)
		}

		uniqueId := uuid.New().String()[:8]
		email := fmt.Sprintf("harian_%s@demo.local", uniqueId)

		req := &CreateEmployeeRequest{
			Name:             name,
			Email:            email,
			BranchID:         &parsedBranchID,
			EmploymentStatus: "Harian",
		}

		if err := u.CreateEmployee(req); err != nil {
			return fmt.Errorf("baris %d: gagal menyimpan data pekerja (%v)", i+1, err)
		}
	}

	return nil
}
