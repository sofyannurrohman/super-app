package usecase

import (
	"encoding/json"
	"log"

	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"example.com/hris/backend/internal/repository"
)

type AuditLogUsecase interface {
	Log(companyID, userID uuid.UUID, userName, action, entity, entityID, description string, oldValue, newValue interface{}, ip string)
	GetAuditLogs(companyID uuid.UUID, entity, action, search string, page, limit int) ([]domain.AuditLog, int64, error)
}

type auditLogUsecase struct {
	repo repository.AuditLogRepository
}

func NewAuditLogUsecase(repo repository.AuditLogRepository) AuditLogUsecase {
	return &auditLogUsecase{repo}
}

func (u *auditLogUsecase) Log(companyID, userID uuid.UUID, userName, action, entity, entityID, description string, oldValue, newValue interface{}, ip string) {
	var oldStr, newStr string

	if oldValue != nil {
		if bytes, err := json.Marshal(oldValue); err == nil {
			oldStr = string(bytes)
		}
	}
	if newValue != nil {
		if bytes, err := json.Marshal(newValue); err == nil {
			newStr = string(bytes)
		}
	}

	audit := &domain.AuditLog{
		CompanyID:   companyID,
		UserID:      userID,
		UserName:    userName,
		Action:      action,
		Entity:      entity,
		EntityID:    entityID,
		Description: description,
		OldValue:    oldStr,
		NewValue:    newStr,
		IPAddress:   ip,
	}

	if err := u.repo.Create(audit); err != nil {
		log.Printf("Failed to create audit log: %v", err)
	}
}

func (u *auditLogUsecase) GetAuditLogs(companyID uuid.UUID, entity, action, search string, page, limit int) ([]domain.AuditLog, int64, error) {
	return u.repo.FindAll(companyID, entity, action, search, page, limit)
}
