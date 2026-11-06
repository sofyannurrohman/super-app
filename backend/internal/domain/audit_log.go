package domain

import (
	"time"

	"github.com/google/uuid"
)

type AuditLog struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CompanyID   uuid.UUID `gorm:"type:uuid;not null" json:"company_id"`
	UserID      uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	UserName    string    `gorm:"type:varchar(100)" json:"user_name"`
	Action      string    `gorm:"type:varchar(20);not null" json:"action"`    // CREATE, UPDATE, DELETE
	Entity      string    `gorm:"type:varchar(50);not null" json:"entity"`    // SalesOrder, Store, etc.
	EntityID    string    `gorm:"type:varchar(50);not null" json:"entity_id"` // UUID string or key
	Description string    `gorm:"type:text" json:"description"`
	OldValue    string    `gorm:"type:text" json:"old_value"`
	NewValue    string    `gorm:"type:text" json:"new_value"`
	IPAddress   string    `gorm:"type:varchar(50)" json:"ip_address"`
	CreatedAt   time.Time `gorm:"default:now()" json:"created_at"`
}
