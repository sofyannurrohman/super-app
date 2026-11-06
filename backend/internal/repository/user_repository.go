package repository

import (
	"errors"
	"strings"

	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"gorm.io/gorm"
)

type UserRepository interface {
	Create(user *domain.User) error
	CreateWithEmployee(user *domain.User, employee *domain.Employee) error
	FindByEmail(email string) (*domain.User, error)
	FindByID(id uuid.UUID) (*domain.User, error)
	Update(user *domain.User) error
	UpdateMobileSession(userID uuid.UUID, sessionID *string) error
	Delete(id uuid.UUID) error
	GetBranchIDByUserID(userID uuid.UUID) (string, error)
	GetBranchByUserID(userID uuid.UUID) (*domain.Branch, error)
	HasPermission(userID uuid.UUID, permissionName string) (bool, error)
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db}
}

func (r *userRepository) Create(user *domain.User) error {
	return r.db.Create(user).Error
}

func (r *userRepository) CreateWithEmployee(user *domain.User, employee *domain.Employee) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(user).Error; err != nil {
			return err
		}

		employee.UserID = user.ID
		if err := tx.Create(employee).Error; err != nil {
			return err
		}

		return nil
	})
}

func (r *userRepository) FindByEmail(email string) (*domain.User, error) {
	var user domain.User
	cleanEmail := strings.ToLower(strings.TrimSpace(email))
	err := r.db.Preload("Company").Preload("AccessRole").Preload("AccessRole.Menus").
		Where("LOWER(TRIM(email)) = ?", cleanEmail).
		Order("is_active DESC, updated_at DESC").
		First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	// Preload DistributorStore safely only if user is associated with a distributor
	if user.DistributorStoreID != nil {
		_ = r.db.Model(&user).Association("DistributorStore").Find(&user.DistributorStore)
	}
	return &user, nil
}
func (r *userRepository) FindByID(id uuid.UUID) (*domain.User, error) {
	var user domain.User
	err := r.db.Preload("Company").Preload("AccessRole").Preload("AccessRole.Menus").First(&user, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	// Preload DistributorStore safely only if user is associated with a distributor
	if user.DistributorStoreID != nil {
		_ = r.db.Model(&user).Association("DistributorStore").Find(&user.DistributorStore)
	}
	return &user, nil
}

func (r *userRepository) Update(user *domain.User) error {
	return r.db.Save(user).Error
}

func (r *userRepository) UpdateMobileSession(userID uuid.UUID, sessionID *string) error {
	return r.db.Model(&domain.User{}).Where("id = ?", userID).Update("mobile_session_id", sessionID).Error
}

func (r *userRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&domain.User{}, "id = ?", id).Error
}

func (r *userRepository) GetBranchIDByUserID(userID uuid.UUID) (string, error) {
	var emp domain.Employee
	err := r.db.Where("user_id = ?", userID).First(&emp).Error
	if err != nil {
		return "", err
	}
	if emp.BranchID != nil {
		return emp.BranchID.String(), nil
	}
	return "", nil
}

func (r *userRepository) GetBranchByUserID(userID uuid.UUID) (*domain.Branch, error) {
	var emp domain.Employee
	err := r.db.Preload("Branch").Where("user_id = ?", userID).First(&emp).Error
	if err != nil {
		return nil, err
	}
	return emp.Branch, nil
}

func (r *userRepository) HasPermission(userID uuid.UUID, permissionName string) (bool, error) {
	var count int64
	err := r.db.Model(&domain.UserPermission{}).
		Joins("JOIN permissions ON permissions.id = user_permissions.permission_id").
		Where("user_permissions.user_id = ? AND permissions.name = ? AND user_permissions.is_granted = ?", userID, permissionName, true).
		Count(&count).Error
	return count > 0, err
}
