package domain

import (
	"time"

	"github.com/google/uuid"
)

type Notification struct {
	ID        uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CompanyID uuid.UUID  `gorm:"type:uuid;not null" json:"company_id"`
	BranchID  *uuid.UUID `gorm:"type:uuid" json:"branch_id"`
	UserID    *uuid.UUID `gorm:"type:uuid" json:"user_id"` // Optional: targeted to a specific user
	Title     string     `gorm:"type:varchar(255);not null" json:"title"`
	Message   string     `gorm:"type:text;not null" json:"message"`
	Type      string     `gorm:"type:varchar(50)" json:"type"` // e.g., LOW_STOCK, SHIPMENT_REQUEST
	IsRead    bool       `gorm:"type:boolean;default:false" json:"is_read"`
	CreatedAt time.Time  `gorm:"default:now()" json:"created_at"`
}
