package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SalesDailyJourney struct {
	ID            uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID        uuid.UUID  `gorm:"type:uuid;not null" json:"user_id"`
	Date          time.Time  `gorm:"type:date;not null" json:"date"`
	StartKM       float64    `gorm:"type:numeric" json:"start_km"`
	StartPhotoURL string     `gorm:"type:text" json:"start_photo_url"`
	StartLocation string     `gorm:"type:text" json:"start_location"`
	StartTime     time.Time  `gorm:"not null" json:"start_time"`
	EndKM         float64    `gorm:"type:numeric" json:"end_km"`
	EndPhotoURL   string     `gorm:"type:text" json:"end_photo_url"`
	EndLocation   string     `gorm:"type:text" json:"end_location"`
	EndTime       *time.Time `json:"end_time"`
	TotalKM       float64    `gorm:"type:numeric" json:"total_km"`
	Status        string     `gorm:"type:varchar(20);not null" json:"status"` // STARTED, ENDED
	CreatedAt     time.Time  `gorm:"default:now()" json:"created_at"`
	UpdatedAt     time.Time  `gorm:"default:now()" json:"updated_at"`

	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (s *SalesDailyJourney) BeforeCreate(tx *gorm.DB) (err error) {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return
}
