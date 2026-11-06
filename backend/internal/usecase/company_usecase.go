package usecase

import (
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"example.com/hris/backend/internal/repository"
)

type CompanyUseCase interface {
	CreateCompany(req *CreateCompanyRequest) error
	GetCompanies() ([]domain.Company, error)
	GetCompanyByID(id uuid.UUID) (*domain.Company, error)
	UpdateCompany(id uuid.UUID, req *UpdateCompanyRequest) error
	DeleteCompany(id uuid.UUID) error
	GetCompanyBranchIDs(id uuid.UUID) ([]uuid.UUID, error)
	SetCompanyBranches(id uuid.UUID, branchIDs []uuid.UUID) error
}

type companyUseCase struct {
	repo repository.CompanyRepository
}

type CreateCompanyRequest struct {
	Name      string      `json:"name" binding:"required"`
	TaxNumber string      `json:"taxNumber"`
	Address   string      `json:"address"`
	LogoURL   string      `json:"logoUrl"`
	BranchIDs []uuid.UUID `json:"branchIds"`
}

type UpdateCompanyRequest struct {
	Name      string      `json:"name" binding:"required"`
	TaxNumber string      `json:"taxNumber"`
	Address   string      `json:"address"`
	LogoURL   string      `json:"logoUrl"`
	BranchIDs []uuid.UUID `json:"branchIds"`
}

func NewCompanyUseCase(repo repository.CompanyRepository) CompanyUseCase {
	return &companyUseCase{repo: repo}
}

func cleanAndUpper(s string) string {
	var res strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			res.WriteRune(r)
		}
	}
	cleaned := strings.ToUpper(res.String())
	// Remove "PT" prefix if present at start to get a nicer code
	if strings.HasPrefix(cleaned, "PT") {
		cleaned = strings.TrimPrefix(cleaned, "PT")
	}
	if cleaned == "" {
		cleaned = "CMP"
	}
	if len(cleaned) > 10 {
		cleaned = cleaned[:10]
	}
	return cleaned
}

func (u *companyUseCase) CreateCompany(req *CreateCompanyRequest) error {
	baseCode := cleanAndUpper(req.Name)
	code := baseCode
	counter := 1
	for {
		existing, err := u.repo.FindByCode(code)
		if err != nil {
			return err
		}
		if existing == nil {
			break
		}
		// Code already exists, append a counter
		suffix := fmt.Sprintf("%d", counter)
		if len(baseCode)+len(suffix) > 20 {
			code = baseCode[:20-len(suffix)] + suffix
		} else {
			code = baseCode + suffix
		}
		counter++
	}

	company := &domain.Company{
		Name:      req.Name,
		Code:      code,
		TaxNumber: req.TaxNumber,
		Address:   req.Address,
		LogoURL:   req.LogoURL,
	}

	if err := u.repo.Create(company); err != nil {
		return err
	}
	return u.repo.SetBranchIDs(company.ID, req.BranchIDs)
}

func (u *companyUseCase) GetCompanies() ([]domain.Company, error) {
	return u.repo.FindAll()
}

func (u *companyUseCase) GetCompanyByID(id uuid.UUID) (*domain.Company, error) {
	return u.repo.FindByID(id)
}

func (u *companyUseCase) GetCompanyBranchIDs(id uuid.UUID) ([]uuid.UUID, error) {
	return u.repo.GetBranchIDs(id)
}

func (u *companyUseCase) SetCompanyBranches(id uuid.UUID, branchIDs []uuid.UUID) error {
	return u.repo.SetBranchIDs(id, branchIDs)
}

func (u *companyUseCase) UpdateCompany(id uuid.UUID, req *UpdateCompanyRequest) error {
	company, err := u.repo.FindByID(id)
	if err != nil {
		return err
	}
	if company == nil {
		return errors.New("company not found")
	}

	// Proactively heal legacy empty codes during updates
	if company.Code == "" {
		baseCode := cleanAndUpper(req.Name)
		code := baseCode
		counter := 1
		for {
			existing, err := u.repo.FindByCode(code)
			if err != nil {
				return err
			}
			if existing == nil {
				break
			}
			suffix := fmt.Sprintf("%d", counter)
			if len(baseCode)+len(suffix) > 20 {
				code = baseCode[:20-len(suffix)] + suffix
			} else {
				code = baseCode + suffix
			}
			counter++
		}
		company.Code = code
	}

	company.Name = req.Name
	company.TaxNumber = req.TaxNumber
	company.Address = req.Address
	company.LogoURL = req.LogoURL

	return u.repo.Update(company)
}

func (u *companyUseCase) DeleteCompany(id uuid.UUID) error {
	return u.repo.Delete(id)
}
