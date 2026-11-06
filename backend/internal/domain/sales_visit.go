package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SalesVisit struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ClientID   *string   `gorm:"type:varchar(50);uniqueIndex" json:"client_id"` // Idempotency key from mobile
	EmployeeID uuid.UUID  `gorm:"type:uuid" json:"employee_id"`
	UserID     *uuid.UUID `gorm:"type:uuid" json:"user_id,omitempty"`
	StoreID    uuid.UUID  `gorm:"type:uuid;not null" json:"store_id"`
	DistributorStoreID *uuid.UUID `gorm:"type:uuid;index" json:"distributor_store_id,omitempty"`
	IsDistributorVisit bool       `gorm:"default:false;index" json:"is_distributor_visit"`

	CheckInTime  time.Time  `gorm:"not null" json:"check_in_time"`
	CheckOutTime *time.Time `json:"check_out_time"`

	Latitude  float64 `gorm:"type:decimal(10,8)" json:"latitude"`
	Longitude float64 `gorm:"type:decimal(11,8)" json:"longitude"`
	SelfieURL string  `gorm:"type:text" json:"selfie_url"`

	Notes  string `gorm:"type:text" json:"notes"`
	Type   string `gorm:"type:varchar(20);default:'CHECKIN'" json:"type"`               // 'CHECKIN', 'CHECKOUT'
	Status string `gorm:"type:varchar(30);default:'PENDING_TRANSACTION'" json:"status"` // PENDING_TRANSACTION, COMPLETED, NO_ORDER

	CreatedAt time.Time `gorm:"default:now()" json:"created_at"`
	UpdatedAt time.Time `gorm:"default:now()" json:"updated_at"`

	Employee         *Employee          `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
	User             *User              `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Store            *Store             `gorm:"foreignKey:StoreID" json:"store,omitempty"`
	DistributorStore *Store             `gorm:"foreignKey:DistributorStoreID" json:"distributor_store,omitempty"`
	Transactions     []SalesTransaction `gorm:"foreignKey:VisitID;constraint:OnDelete:CASCADE;" json:"transactions,omitempty"`
	SalesOrders      []SalesOrder       `gorm:"foreignKey:VisitID;constraint:OnDelete:CASCADE;" json:"sales_orders,omitempty"`
}

func (s *SalesVisit) BeforeCreate(tx *gorm.DB) (err error) {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return
}
