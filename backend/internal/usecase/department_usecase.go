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

type DepartmentUseCase interface {
	CreateDepartment(req *CreateDepartmentRequest) error
	GetDepartments() ([]domain.Department, error)
	GetDepartmentByID(id uuid.UUID) (*domain.Department, error)
	UpdateDepartment(id uuid.UUID, req *UpdateDepartmentRequest) error
	DeleteDepartment(id uuid.UUID) error
}

type departmentUseCase struct {
	repo repository.DepartmentRepository
}

type CreateDepartmentRequest struct {
	CompanyID *uuid.UUID `json:"companyId"`
	Name      string     `json:"name" binding:"required"`
	ParentID  *uuid.UUID `json:"parentId"`
}

type UpdateDepartmentRequest struct {
	Name     string     `json:"name" binding:"required"`
	ParentID *uuid.UUID `json:"parentId"`
}

func NewDepartmentUseCase(repo repository.DepartmentRepository) DepartmentUseCase {
	return &departmentUseCase{repo: repo}
}

func (u *departmentUseCase) CreateDepartment(req *CreateDepartmentRequest) error {
	department := &domain.Department{
		CompanyID: req.CompanyID,
		Name:      req.Name,
		ParentID:  req.ParentID,
	}

	if err := u.repo.Create(department); err != nil {
		return err
	}
	invalidateDepartmentCache(context.Background(), uuid.Nil)
	return nil
}

func (u *departmentUseCase) GetDepartments() ([]domain.Department, error) {
	ctx := context.Background()
	cacheKey := "master:departments"

	if redisclient.Client != nil {
		if cached, err := redisclient.Client.Get(ctx, cacheKey).Result(); err == nil {
			var depts []domain.Department
			if json.Unmarshal([]byte(cached), &depts) == nil {
				return depts, nil
			}
		}
	}

	depts, err := u.repo.FindAll()
	if err != nil {
		return nil, err
	}

	if redisclient.Client != nil {
		if bytes, err := json.Marshal(depts); err == nil {
			_ = redisclient.Client.Set(ctx, cacheKey, bytes, 24*time.Hour).Err()
		}
	}

	return depts, nil
}

func (u *departmentUseCase) GetDepartmentByID(id uuid.UUID) (*domain.Department, error) {
	ctx := context.Background()
	cacheKey := fmt.Sprintf("master:department:id:%s", id.String())

	if redisclient.Client != nil {
		if cached, err := redisclient.Client.Get(ctx, cacheKey).Result(); err == nil {
			var dept domain.Department
			if json.Unmarshal([]byte(cached), &dept) == nil {
				return &dept, nil
			}
		}
	}

	dept, err := u.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if dept == nil {
		return nil, nil
	}

	if redisclient.Client != nil {
		if bytes, err := json.Marshal(dept); err == nil {
			_ = redisclient.Client.Set(ctx, cacheKey, bytes, 2*time.Hour).Err()
		}
	}

	return dept, nil
}

func invalidateDepartmentCache(ctx context.Context, id uuid.UUID) {
	if redisclient.Client == nil {
		return
	}
	_ = redisclient.Client.Del(ctx, "master:departments").Err()
	if id != uuid.Nil {
		_ = redisclient.Client.Del(ctx, fmt.Sprintf("master:department:id:%s", id.String())).Err()
	}
}

func (u *departmentUseCase) UpdateDepartment(id uuid.UUID, req *UpdateDepartmentRequest) error {
	department, err := u.repo.FindByID(id)
	if err != nil {
		return err
	}
	if department == nil {
		return errors.New("department not found")
	}

	department.Name = req.Name
	// Optional assignment for parent department updates
	if req.ParentID != nil && *req.ParentID != uuid.Nil {
		department.ParentID = req.ParentID
	} else {
		department.ParentID = nil
	}

	if err := u.repo.Update(department); err != nil {
		return err
	}
	invalidateDepartmentCache(context.Background(), id)
	return nil
}

func (u *departmentUseCase) DeleteDepartment(id uuid.UUID) error {
	if err := u.repo.Delete(id); err != nil {
		return err
	}
	invalidateDepartmentCache(context.Background(), id)
	return nil
}
