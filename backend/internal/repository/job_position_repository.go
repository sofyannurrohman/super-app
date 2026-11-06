package repository

import (
	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"gorm.io/gorm"
)

type JobPositionRepository interface {
	Create(jobPosition *domain.JobPosition) error
	FindAll() ([]domain.JobPosition, error)
	FindByID(id uuid.UUID) (*domain.JobPosition, error)
	Update(jobPosition *domain.JobPosition) error
	Delete(id uuid.UUID) error
}

type jobPositionRepository struct {
	db *gorm.DB
}

func NewJobPositionRepository(db *gorm.DB) JobPositionRepository {
	return &jobPositionRepository{db}
}

func (r *jobPositionRepository) Create(jobPosition *domain.JobPosition) error {
	return r.db.Create(jobPosition).Error
}

func (r *jobPositionRepository) FindAll() ([]domain.JobPosition, error) {
	var positions []domain.JobPosition
	// Only fetch positions, optionally preload company if needed, but not strictly required
	if err := r.db.Find(&positions).Error; err != nil {
		return nil, err
	}
	return positions, nil
}

func (r *jobPositionRepository) FindByID(id uuid.UUID) (*domain.JobPosition, error) {
	var position domain.JobPosition
	if err := r.db.Where("id = ?", id).First(&position).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &position, nil
}

func (r *jobPositionRepository) Update(jobPosition *domain.JobPosition) error {
	return r.db.Save(jobPosition).Error
}

func (r *jobPositionRepository) Delete(id uuid.UUID) error {
	return r.db.Where("id = ?", id).Delete(&domain.JobPosition{}).Error
}
