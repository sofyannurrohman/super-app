package usecase

import (
	"time"

	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"example.com/hris/backend/internal/repository"
)

type PayrollConfigUseCase interface {
	GetConfig(companyID uuid.UUID) (*domain.PayrollConfig, error)
	UpdateConfig(companyID uuid.UUID, config *domain.PayrollConfig) error
}

type payrollConfigUseCase struct {
	repo repository.PayrollConfigRepository
}

func NewPayrollConfigUseCase(repo repository.PayrollConfigRepository) PayrollConfigUseCase {
	return &payrollConfigUseCase{repo}
}

func (u *payrollConfigUseCase) GetConfig(companyID uuid.UUID) (*domain.PayrollConfig, error) {
	config, err := u.repo.GetByCompanyID(companyID)
	if err != nil {
		return nil, err
	}

	// If no config exists, return default values from the model
	if config == nil {
		defaultConfig := &domain.PayrollConfig{
			CompanyID: &companyID,
			// Defaults are set in the domain model tags or BeforeCreate
			JHTCompanyPercentage:      3.70,
			JHTEmployeePercentage:     2.00,
			JPCompanyPercentage:       2.00,
			JPEmployeePercentage:      1.00,
			JPMaxWageBase:             10042300,
			JKKCompanyPercentage:      0.24,
			JKMCompanyPercentage:      0.30,
			BPJSKesCompanyPercentage:  4.00,
			BPJSKesEmployeePercentage: 1.00,
			BPJSKesMaxWageBase:        12000000,
			PtkpBaseTK0:               54000000,
			AbsentDeduction:           100000.0,
		}
		return defaultConfig, nil
	}

	return config, nil
}

func (u *payrollConfigUseCase) UpdateConfig(companyID uuid.UUID, config *domain.PayrollConfig) error {
	config.CompanyID = &companyID
	config.UpdatedAt = time.Now()

	// If ID is nil, it's a new config for this company
	if config.ID == uuid.Nil {
		config.ID = uuid.New()
	}

	return u.repo.Upsert(config)
}
