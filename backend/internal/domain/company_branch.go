package domain

import (
	"time"

	"github.com/google/uuid"
)

// CompanyBranch menyatakan bahwa sebuah PT diizinkan beroperasi pada cabang
// tertentu. Satu cabang dapat mengelola banyak PT dan satu PT dapat memiliki
// banyak cabang.
type CompanyBranch struct {
	CompanyID uuid.UUID `gorm:"type:uuid;primaryKey;not null" json:"company_id"`
	BranchID  uuid.UUID `gorm:"type:uuid;primaryKey;not null" json:"branch_id"`
	IsActive  bool      `gorm:"default:true" json:"is_active"`
	CreatedAt time.Time `gorm:"default:now()" json:"created_at"`
	UpdatedAt time.Time `gorm:"default:now()" json:"updated_at"`

	Company *Company `gorm:"foreignKey:CompanyID" json:"company,omitempty"`
	Branch  *Branch  `gorm:"foreignKey:BranchID" json:"branch,omitempty"`
}
