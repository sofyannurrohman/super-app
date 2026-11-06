package usecase

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"example.com/hris/backend/internal/repository"
)

type EpkwtUseCase interface {
	CreateRequest(employeeID uuid.UUID, req *CreateEpkwtRequestParams) (*domain.EpkwtRequest, error)
	UpdateRequest(id uuid.UUID, req *UpdateEpkwtRequestParams) error
	SignRequest(id uuid.UUID, signatureURL string) error
	UploadDocument(id uuid.UUID, documentURL string) error
	GetByID(id uuid.UUID) (*domain.EpkwtRequest, error)
	GetAll(employeeID *uuid.UUID) ([]domain.EpkwtRequest, error)
}

type epkwtUseCase struct {
	repo         repository.EpkwtRepository
	employeeRepo repository.EmployeeRepository
}

func NewEpkwtUseCase(repo repository.EpkwtRepository, employeeRepo repository.EmployeeRepository) EpkwtUseCase {
	return &epkwtUseCase{
		repo:         repo,
		employeeRepo: employeeRepo,
	}
}

func generateEpkwtNumber(count int64, date time.Time) string {
	romanMonths := []string{"", "I", "II", "III", "IV", "V", "VI", "VII", "VIII", "IX", "X", "XI", "XII"}
	month := int(date.Month())
	romanMonth := ""
	if month >= 1 && month <= 12 {
		romanMonth = romanMonths[month]
	}
	year := date.Year()

	return fmt.Sprintf("%03d/DEMO-PKWT/%s/%d", count+1, romanMonth, year)
}

type CreateEpkwtRequestParams struct {
	IdentityNumber       string     `json:"identity_number" binding:"required"`
	FirstName            string     `json:"first_name" binding:"required"`
	LastName             *string    `json:"last_name"`
	PhoneNumber          *string    `json:"phone_number"`
	Gender               *string    `json:"gender"`
	BirthDate            *time.Time `json:"birth_date"`
	BirthPlace           *string    `json:"birth_place"`
	AddressKtp           *string    `json:"address_ktp"`
	AddressResidential   *string    `json:"address_residential"`
	MaritalStatus        *string    `json:"marital_status"`
	EmployeeSignatureURL *string    `json:"employee_signature_url"`
	ContractSequence     *int       `json:"contract_sequence"`
	DocumentURL          *string    `json:"document_url"`
	Notes                *string    `json:"notes"`
}

type UpdateEpkwtRequestParams struct {
	DocumentNumber   *string    `json:"document_number"`
	JobPositionID    *uuid.UUID `json:"job_position_id"`
	DepartmentID     *uuid.UUID `json:"department_id"`
	Salary           float64    `json:"salary"`
	StartDate        *time.Time `json:"start_date"`
	EndDate          *time.Time `json:"end_date"`
	AgreementDate    *time.Time `json:"agreement_date"`
	ProbationMonths  *int       `json:"probation_months"`
	HrSignatureURL   *string    `json:"hr_signature_url"`
	Status           string     `json:"status"` // PENDING_HR, APPROVED_DRAFT, ACTIVE, REJECTED
	ContractSequence *int       `json:"contract_sequence"`
	DocumentURL      *string    `json:"document_url"`
	Notes            *string    `json:"notes"`
}

func (u *epkwtUseCase) CreateRequest(employeeID uuid.UUID, req *CreateEpkwtRequestParams) (*domain.EpkwtRequest, error) {
	// Verify employee exists
	emp, err := u.employeeRepo.FindByID(employeeID)
	if err != nil || emp == nil {
		return nil, errors.New("employee profile not found")
	}

	seq := 1
	if req.ContractSequence != nil && *req.ContractSequence > 0 {
		seq = *req.ContractSequence
	}

	epkwt := &domain.EpkwtRequest{
		EmployeeID:           employeeID,
		IdentityNumber:       req.IdentityNumber,
		FirstName:            req.FirstName,
		LastName:             req.LastName,
		PhoneNumber:          req.PhoneNumber,
		Gender:               req.Gender,
		BirthDate:            req.BirthDate,
		BirthPlace:           req.BirthPlace,
		AddressKtp:           req.AddressKtp,
		AddressResidential:   req.AddressResidential,
		MaritalStatus:        req.MaritalStatus,
		EmployeeSignatureURL: req.EmployeeSignatureURL,
		ContractSequence:     seq,
		DocumentURL:          req.DocumentURL,
		Notes:                req.Notes,
		Status:               "PENDING_HR",
	}

	if err := u.repo.Create(epkwt); err != nil {
		return nil, err
	}
	return epkwt, nil
}

func (u *epkwtUseCase) UpdateRequest(id uuid.UUID, req *UpdateEpkwtRequestParams) error {
	epkwt, err := u.repo.FindByID(id)
	if err != nil || epkwt == nil {
		return errors.New("EPKWT request not found")
	}

	if req.DocumentNumber == nil || *req.DocumentNumber == "" {
		if epkwt.DocumentNumber == nil || *epkwt.DocumentNumber == "" {
			now := time.Now()
			if req.AgreementDate != nil {
				now = *req.AgreementDate
			} else if epkwt.AgreementDate != nil {
				now = *epkwt.AgreementDate
			}

			count, err := u.repo.CountByYear(now.Year())
			if err == nil {
				docNum := generateEpkwtNumber(count, now)
				epkwt.DocumentNumber = &docNum
			}
		}
	} else {
		epkwt.DocumentNumber = req.DocumentNumber
	}

	if req.JobPositionID != nil {
		epkwt.JobPositionID = req.JobPositionID
	}
	if req.DepartmentID != nil {
		epkwt.DepartmentID = req.DepartmentID
	}
	if req.Salary > 0 {
		epkwt.Salary = req.Salary
	}
	if req.StartDate != nil {
		epkwt.StartDate = req.StartDate
	}
	if req.EndDate != nil {
		epkwt.EndDate = req.EndDate
	}
	if req.AgreementDate != nil {
		epkwt.AgreementDate = req.AgreementDate
	}
	if req.ProbationMonths != nil {
		epkwt.ProbationMonths = req.ProbationMonths
	}
	if req.HrSignatureURL != nil {
		epkwt.HrSignatureURL = req.HrSignatureURL
	}
	if req.Status != "" {
		epkwt.Status = req.Status
	}
	if req.ContractSequence != nil && *req.ContractSequence > 0 {
		epkwt.ContractSequence = *req.ContractSequence
	}
	if req.DocumentURL != nil {
		epkwt.DocumentURL = req.DocumentURL
	}
	if req.Notes != nil {
		epkwt.Notes = req.Notes
	}

	return u.repo.Update(epkwt)
}

func (u *epkwtUseCase) SignRequest(id uuid.UUID, signatureURL string) error {
	epkwt, err := u.repo.FindByID(id)
	if err != nil || epkwt == nil {
		return errors.New("EPKWT request not found")
	}

	epkwt.EmployeeSignatureURL = &signatureURL
	epkwt.Status = "ACTIVE"

	return u.repo.Update(epkwt)
}

func (u *epkwtUseCase) UploadDocument(id uuid.UUID, documentURL string) error {
	epkwt, err := u.repo.FindByID(id)
	if err != nil || epkwt == nil {
		return errors.New("EPKWT request not found")
	}

	epkwt.DocumentURL = &documentURL
	return u.repo.Update(epkwt)
}

func (u *epkwtUseCase) GetByID(id uuid.UUID) (*domain.EpkwtRequest, error) {
	return u.repo.FindByID(id)
}

func (u *epkwtUseCase) GetAll(employeeID *uuid.UUID) ([]domain.EpkwtRequest, error) {
	return u.repo.FindAll(employeeID)
}
