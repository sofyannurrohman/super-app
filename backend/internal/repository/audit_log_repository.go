package repository

import (
	"strings"

	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"gorm.io/gorm"
)

type AuditLogRepository interface {
	Create(log *domain.AuditLog) error
	FindAll(companyID uuid.UUID, entity, action, search string, page, limit int) ([]domain.AuditLog, int64, error)
}

type auditLogRepository struct {
	db *gorm.DB
}

func NewAuditLogRepository(db *gorm.DB) AuditLogRepository {
	return &auditLogRepository{db}
}

func (r *auditLogRepository) Create(log *domain.AuditLog) error {
	return r.db.Create(log).Error
}

func (r *auditLogRepository) FindAll(companyID uuid.UUID, entity, action, search string, page, limit int) ([]domain.AuditLog, int64, error) {
	var logs []domain.AuditLog
	var total int64

	query := r.db.Model(&domain.AuditLog{}).Where("company_id = ?", companyID)

	if entity != "" && entity != "ALL" {
		entities := strings.Split(entity, ",")
		if len(entities) > 1 {
			query = query.Where("entity IN ?", entities)
		} else {
			query = query.Where("entity = ?", entity)
		}
	}

	if action != "" && action != "ALL" {
		query = query.Where("action = ?", action)
	}

	if search != "" {
		searchLike := "%" + search + "%"
		query = query.Where("(user_name ILIKE ? OR description ILIKE ? OR ip_address ILIKE ?)", searchLike, searchLike, searchLike)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&logs).Error
	return logs, total, err
}
