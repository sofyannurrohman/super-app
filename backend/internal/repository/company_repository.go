package repository

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"gorm.io/gorm"
)

type CompanyRepository interface {
	Create(company *domain.Company) error
	FindAll() ([]domain.Company, error)
	FindByID(id uuid.UUID) (*domain.Company, error)
	Update(company *domain.Company) error
	Delete(id uuid.UUID) error
	FindByName(name string) (*domain.Company, error)
	FindByCode(code string) (*domain.Company, error)
	GetBranchIDs(companyID uuid.UUID) ([]uuid.UUID, error)
	SetBranchIDs(companyID uuid.UUID, branchIDs []uuid.UUID) error
}

func (r *companyRepository) GetBranchIDs(companyID uuid.UUID) ([]uuid.UUID, error) {
	var ids []uuid.UUID
	err := r.db.Model(&domain.CompanyBranch{}).Where("company_id = ? AND is_active = true", companyID).Pluck("branch_id", &ids).Error
	return ids, err
}

func (r *companyRepository) SetBranchIDs(companyID uuid.UUID, branchIDs []uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("company_id = ?", companyID).Delete(&domain.CompanyBranch{}).Error; err != nil {
			return err
		}
		for _, branchID := range branchIDs {
			if branchID == uuid.Nil {
				continue
			}
			if err := tx.Create(&domain.CompanyBranch{CompanyID: companyID, BranchID: branchID, IsActive: true}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

type companyRepository struct {
	db *gorm.DB
}

func NewCompanyRepository(db *gorm.DB) CompanyRepository {
	return &companyRepository{db}
}

func (r *companyRepository) Create(company *domain.Company) error {
	return r.db.Create(company).Error
}

func (r *companyRepository) FindAll() ([]domain.Company, error) {
	var companies []domain.Company
	if err := r.db.Find(&companies).Error; err != nil {
		return nil, err
	}
	return companies, nil
}

func (r *companyRepository) FindByID(id uuid.UUID) (*domain.Company, error) {
	var company domain.Company
	if err := r.db.Where("id = ?", id).First(&company).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &company, nil
}

func (r *companyRepository) Update(company *domain.Company) error {
	return r.db.Save(company).Error
}

func (r *companyRepository) Delete(id uuid.UUID) error {
	// Check for associated branches
	var branchCount int64
	r.db.Model(&domain.Branch{}).Where("company_id = ?", id).Count(&branchCount)
	if branchCount > 0 {
		return fmt.Errorf("tidak dapat menghapus perusahaan: terdapat %d cabang yang masih terhubung. Hapus cabang terlebih dahulu", branchCount)
	}

	// Check for associated employees
	var employeeCount int64
	r.db.Model(&domain.Employee{}).Where("company_id = ?", id).Count(&employeeCount)
	if employeeCount > 0 {
		return fmt.Errorf("tidak dapat menghapus perusahaan: terdapat %d karyawan yang masih terdaftar. Hapus atau pindahkan karyawan terlebih dahulu", employeeCount)
	}

	// Check for associated departments
	var deptCount int64
	r.db.Model(&domain.Department{}).Where("company_id = ?", id).Count(&deptCount)
	if deptCount > 0 {
		return fmt.Errorf("tidak dapat menghapus perusahaan: terdapat %d departemen yang masih terhubung. Hapus departemen terlebih dahulu", deptCount)
	}

	// Check for associated job positions
	var jobCount int64
	r.db.Model(&domain.JobPosition{}).Where("company_id = ?", id).Count(&jobCount)
	if jobCount > 0 {
		return fmt.Errorf("tidak dapat menghapus perusahaan: terdapat %d posisi jabatan yang masih terhubung. Hapus posisi jabatan terlebih dahulu", jobCount)
	}

	result := r.db.Where("id = ?", id).Delete(&domain.Company{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("perusahaan tidak ditemukan")
	}
	return nil
}

func (r *companyRepository) FindByName(name string) (*domain.Company, error) {
	var company domain.Company
	if err := r.db.Where("name = ?", name).First(&company).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &company, nil
}

func (r *companyRepository) FindByCode(code string) (*domain.Company, error) {
	var company domain.Company
	if err := r.db.Where("code = ?", code).First(&company).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &company, nil
}
