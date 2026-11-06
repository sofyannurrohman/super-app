package repository

import (
	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"gorm.io/gorm"
)

type NotificationRepository interface {
	Create(notification *domain.Notification) error
	GetByBranch(companyID, branchID uuid.UUID) ([]domain.Notification, error)
	MarkAsRead(id uuid.UUID) error
}

type notificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) NotificationRepository {
	return &notificationRepository{db}
}

func (r *notificationRepository) Create(notification *domain.Notification) error {
	return r.db.Create(notification).Error
}

func (r *notificationRepository) GetByBranch(companyID, branchID uuid.UUID) ([]domain.Notification, error) {
	var notifications []domain.Notification
	err := r.db.Where("company_id = ? AND (branch_id = ? OR branch_id IS NULL)", companyID, branchID).
		Order("created_at desc").Limit(50).Find(&notifications).Error
	return notifications, err
}

func (r *notificationRepository) MarkAsRead(id uuid.UUID) error {
	return r.db.Model(&domain.Notification{}).Where("id = ?", id).Update("is_read", true).Error
}
