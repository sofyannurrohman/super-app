package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"example.com/hris/backend/internal/repository"
	"example.com/hris/backend/pkg/redisclient"
)

type JobPositionUseCase interface {
	CreateJobPosition(req *CreateJobPositionRequest) error
	GetJobPositions() ([]domain.JobPosition, error)
	GetJobPositionByID(id uuid.UUID) (*domain.JobPosition, error)
	UpdateJobPosition(id uuid.UUID, req *UpdateJobPositionRequest) error
	DeleteJobPosition(id uuid.UUID) error
}

type jobPositionUseCase struct {
	repo repository.JobPositionRepository
}

type CreateJobPositionRequest struct {
	Title     string     `json:"title" binding:"required"`
	Level     int        `json:"level"`
	CompanyID *uuid.UUID `json:"companyId"`
}

type UpdateJobPositionRequest struct {
	Title string `json:"title" binding:"required"`
	Level int    `json:"level"`
}

func NewJobPositionUseCase(repo repository.JobPositionRepository) JobPositionUseCase {
	return &jobPositionUseCase{repo: repo}
}

func (u *jobPositionUseCase) CreateJobPosition(req *CreateJobPositionRequest) error {
	position := &domain.JobPosition{
		Title:     req.Title,
		Level:     req.Level,
		CompanyID: req.CompanyID,
	}

	if err := u.repo.Create(position); err != nil {
		return err
	}
	invalidateJobPositionCache(context.Background(), uuid.Nil)
	return nil
}

func (u *jobPositionUseCase) GetJobPositions() ([]domain.JobPosition, error) {
	ctx := context.Background()
	cacheKey := "master:jobpositions"

	if redisclient.Client != nil {
		if cached, err := redisclient.Client.Get(ctx, cacheKey).Result(); err == nil {
			var positions []domain.JobPosition
			if json.Unmarshal([]byte(cached), &positions) == nil {
				return positions, nil
			}
		}
	}

	positions, err := u.repo.FindAll()
	if err != nil {
		return nil, err
	}

	if redisclient.Client != nil {
		if bytes, err := json.Marshal(positions); err == nil {
			_ = redisclient.Client.Set(ctx, cacheKey, bytes, 24*time.Hour).Err()
		}
	}

	return positions, nil
}

func (u *jobPositionUseCase) GetJobPositionByID(id uuid.UUID) (*domain.JobPosition, error) {
	ctx := context.Background()
	cacheKey := fmt.Sprintf("master:jobposition:id:%s", id.String())

	if redisclient.Client != nil {
		if cached, err := redisclient.Client.Get(ctx, cacheKey).Result(); err == nil {
			var pos domain.JobPosition
			if json.Unmarshal([]byte(cached), &pos) == nil {
				return &pos, nil
			}
		}
	}

	pos, err := u.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if pos == nil {
		return nil, nil
	}

	if redisclient.Client != nil {
		if bytes, err := json.Marshal(pos); err == nil {
			_ = redisclient.Client.Set(ctx, cacheKey, bytes, 2*time.Hour).Err()
		}
	}

	return pos, nil
}

func invalidateJobPositionCache(ctx context.Context, id uuid.UUID) {
	if redisclient.Client == nil {
		return
	}
	_ = redisclient.Client.Del(ctx, "master:jobpositions").Err()
	if id != uuid.Nil {
		_ = redisclient.Client.Del(ctx, fmt.Sprintf("master:jobposition:id:%s", id.String())).Err()
	}
}

func (u *jobPositionUseCase) UpdateJobPosition(id uuid.UUID, req *UpdateJobPositionRequest) error {
	position, err := u.repo.FindByID(id)
	if err != nil {
		return err
	}
	if position == nil {
		return errors.New("job position not found")
	}

	position.Title = req.Title
	position.Level = req.Level

	if err := u.repo.Update(position); err != nil {
		return err
	}
	invalidateJobPositionCache(context.Background(), id)
	return nil
}

func (u *jobPositionUseCase) DeleteJobPosition(id uuid.UUID) error {
	if err := u.repo.Delete(id); err != nil {
		return err
	}
	invalidateJobPositionCache(context.Background(), id)
	return nil
}
