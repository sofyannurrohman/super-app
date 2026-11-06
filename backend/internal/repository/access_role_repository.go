package repository

import (
	"context"

	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"gorm.io/gorm"
)

type AccessRoleRepository interface {
	CreateRole(ctx context.Context, role *domain.AccessRole) error
	GetRoleByID(ctx context.Context, id uuid.UUID) (*domain.AccessRole, error)
	GetAllRoles(ctx context.Context) ([]domain.AccessRole, error)
	UpdateRole(ctx context.Context, role *domain.AccessRole) error
	DeleteRole(ctx context.Context, id uuid.UUID) error

	CreateModule(ctx context.Context, module *domain.AppModule) error
	GetAllModules(ctx context.Context) ([]domain.AppModule, error)

	CreateMenu(ctx context.Context, menu *domain.AppMenu) error
	GetAllMenus(ctx context.Context) ([]domain.AppMenu, error)
	GetMenusByModule(ctx context.Context, moduleID uuid.UUID) ([]domain.AppMenu, error)
}

type accessRoleRepository struct {
	db *gorm.DB
}

func NewAccessRoleRepository(db *gorm.DB) AccessRoleRepository {
	return &accessRoleRepository{db}
}

func (r *accessRoleRepository) CreateRole(ctx context.Context, role *domain.AccessRole) error {
	return r.db.WithContext(ctx).Create(role).Error
}

func (r *accessRoleRepository) GetRoleByID(ctx context.Context, id uuid.UUID) (*domain.AccessRole, error) {
	var role domain.AccessRole
	err := r.db.WithContext(ctx).Preload("Menus").First(&role, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *accessRoleRepository) GetAllRoles(ctx context.Context) ([]domain.AccessRole, error) {
	var roles []domain.AccessRole
	err := r.db.WithContext(ctx).Preload("Menus").Find(&roles).Error
	return roles, err
}

func (r *accessRoleRepository) UpdateRole(ctx context.Context, role *domain.AccessRole) error {
	// Update menus association
	err := r.db.WithContext(ctx).Model(role).Association("Menus").Replace(role.Menus)
	if err != nil {
		return err
	}
	return r.db.WithContext(ctx).Save(role).Error
}

func (r *accessRoleRepository) DeleteRole(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.AccessRole{}, "id = ?", id).Error
}

func (r *accessRoleRepository) CreateModule(ctx context.Context, module *domain.AppModule) error {
	return r.db.WithContext(ctx).Create(module).Error
}

func (r *accessRoleRepository) GetAllModules(ctx context.Context) ([]domain.AppModule, error) {
	var modules []domain.AppModule
	err := r.db.WithContext(ctx).Preload("Menus").Find(&modules).Error
	return modules, err
}

func (r *accessRoleRepository) CreateMenu(ctx context.Context, menu *domain.AppMenu) error {
	return r.db.WithContext(ctx).Create(menu).Error
}

func (r *accessRoleRepository) GetAllMenus(ctx context.Context) ([]domain.AppMenu, error) {
	var menus []domain.AppMenu
	err := r.db.WithContext(ctx).Find(&menus).Error
	return menus, err
}

func (r *accessRoleRepository) GetMenusByModule(ctx context.Context, moduleID uuid.UUID) ([]domain.AppMenu, error) {
	var menus []domain.AppMenu
	err := r.db.WithContext(ctx).Where("module_id = ?", moduleID).Find(&menus).Error
	return menus, err
}
