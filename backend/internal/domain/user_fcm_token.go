package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserFcmToken struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID     uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_user_fcm" json:"user_id"`
	FCMToken   string    `gorm:"type:varchar(255);not null" json:"fcm_token"`
	DeviceInfo string    `gorm:"type:varchar(100)" json:"device_info"`
	CreatedAt  time.Time `gorm:"default:now()" json:"created_at"`
	UpdatedAt  time.Time `gorm:"default:now()" json:"updated_at"`

	User *User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE;" json:"user,omitempty"`
}

func (t *UserFcmToken) BeforeCreate(tx *gorm.DB) (err error) {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return
}
