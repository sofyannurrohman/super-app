package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AppModule represents a major module in the app (e.g., HR, Sales, Inventory)
type AppModule struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name        string    `gorm:"type:varchar(100);not null;unique" json:"name"`
	Description string    `gorm:"type:text" json:"description"`
	CreatedAt   time.Time `gorm:"default:now()" json:"created_at"`

	Menus []AppMenu `gorm:"foreignKey:ModuleID" json:"menus,omitempty"`
}

func (m *AppModule) BeforeCreate(tx *gorm.DB) (err error) {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return
}

// AppMenu represents a specific menu inside a module
type AppMenu struct {
	ID        uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ModuleID  uuid.UUID  `gorm:"type:uuid;not null" json:"module_id"`
	Name      string     `gorm:"type:varchar(100);not null" json:"name"`
	Path      string     `gorm:"type:varchar(150)" json:"path"`
	Level     int        `gorm:"default:1" json:"level"`
	ParentID  *uuid.UUID `gorm:"type:uuid" json:"parent_id"`
	CreatedAt time.Time  `gorm:"default:now()" json:"created_at"`

	Module *AppModule `gorm:"foreignKey:ModuleID" json:"module,omitempty"`
	Parent *AppMenu   `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
}

func (m *AppMenu) BeforeCreate(tx *gorm.DB) (err error) {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return
}

// AccessRole represents the dynamic user role
type AccessRole struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name        string    `gorm:"type:varchar(100);not null;unique" json:"name"`
	Description string    `gorm:"type:text" json:"description"`
	CreatedAt   time.Time `gorm:"default:now()" json:"created_at"`
	UpdatedAt   time.Time `gorm:"default:now()" json:"updated_at"`

	Menus []AppMenu `gorm:"many2many:access_role_menus;" json:"menus,omitempty"`
}

func (r *AccessRole) BeforeCreate(tx *gorm.DB) (err error) {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return
}
