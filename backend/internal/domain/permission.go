package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Permission struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name        string    `gorm:"type:varchar(100);not null;unique" json:"name"`
	MenuGroup   string    `gorm:"type:varchar(100);not null" json:"menu_group"`
	Description string    `gorm:"type:text" json:"description"`
	CreatedAt   time.Time `gorm:"default:now()" json:"created_at"`
}

func (p *Permission) BeforeCreate(tx *gorm.DB) (err error) {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return
}

type UserPermission struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID       uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_user_permission" json:"user_id"`
	PermissionID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_user_permission" json:"permission_id"`
	IsGranted    bool      `gorm:"default:true" json:"is_granted"`
	CreatedAt    time.Time `gorm:"default:now()" json:"created_at"`

	User       *User       `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE;" json:"user,omitempty"`
	Permission *Permission `gorm:"foreignKey:PermissionID;constraint:OnDelete:CASCADE;" json:"permission,omitempty"`
}

func (up *UserPermission) BeforeCreate(tx *gorm.DB) (err error) {
	if up.ID == uuid.Nil {
		up.ID = uuid.New()
	}
	return
}
