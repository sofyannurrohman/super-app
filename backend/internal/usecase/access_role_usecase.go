package usecase

import (
	"context"

	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"example.com/hris/backend/internal/repository"
)

type AccessRoleUsecase interface {
	CreateRole(ctx context.Context, role *domain.AccessRole, menuIDs []uuid.UUID) error
	GetRoleByID(ctx context.Context, id uuid.UUID) (*domain.AccessRole, error)
	GetAllRoles(ctx context.Context) ([]domain.AccessRole, error)
	UpdateRole(ctx context.Context, role *domain.AccessRole, menuIDs []uuid.UUID) error
	DeleteRole(ctx context.Context, id uuid.UUID) error

	CreateModule(ctx context.Context, module *domain.AppModule) error
	GetAllModules(ctx context.Context) ([]domain.AppModule, error)

	CreateMenu(ctx context.Context, menu *domain.AppMenu) error
	GetAllMenus(ctx context.Context) ([]domain.AppMenu, error)
}

type accessRoleUsecase struct {
	repo repository.AccessRoleRepository
}

func NewAccessRoleUsecase(repo repository.AccessRoleRepository) AccessRoleUsecase {
	return &accessRoleUsecase{repo}
}

func (u *accessRoleUsecase) CreateRole(ctx context.Context, role *domain.AccessRole, menuIDs []uuid.UUID) error {
	var menus []domain.AppMenu
	for _, id := range menuIDs {
		menus = append(menus, domain.AppMenu{ID: id})
	}
	role.Menus = menus
	return u.repo.CreateRole(ctx, role)
}

func (u *accessRoleUsecase) GetRoleByID(ctx context.Context, id uuid.UUID) (*domain.AccessRole, error) {
	return u.repo.GetRoleByID(ctx, id)
}

func (u *accessRoleUsecase) GetAllRoles(ctx context.Context) ([]domain.AccessRole, error) {
	return u.repo.GetAllRoles(ctx)
}

func (u *accessRoleUsecase) UpdateRole(ctx context.Context, role *domain.AccessRole, menuIDs []uuid.UUID) error {
	var menus []domain.AppMenu
	for _, id := range menuIDs {
		menus = append(menus, domain.AppMenu{ID: id})
	}
	role.Menus = menus
	return u.repo.UpdateRole(ctx, role)
}

func (u *accessRoleUsecase) DeleteRole(ctx context.Context, id uuid.UUID) error {
	return u.repo.DeleteRole(ctx, id)
}

func (u *accessRoleUsecase) CreateModule(ctx context.Context, module *domain.AppModule) error {
	return u.repo.CreateModule(ctx, module)
}

func (u *accessRoleUsecase) GetAllModules(ctx context.Context) ([]domain.AppModule, error) {
	return u.repo.GetAllModules(ctx)
}

func (u *accessRoleUsecase) CreateMenu(ctx context.Context, menu *domain.AppMenu) error {
	return u.repo.CreateMenu(ctx, menu)
}

func (u *accessRoleUsecase) GetAllMenus(ctx context.Context) ([]domain.AppMenu, error) {
	return u.repo.GetAllMenus(ctx)
}
