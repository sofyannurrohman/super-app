package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

// TargetType defines who should receive an announcement notification.
type TargetType string

const (
	TargetBroadcast  TargetType = "broadcast"  // semua user
	TargetRole       TargetType = "role"        // berdasarkan role tertentu
	TargetIndividual TargetType = "individual"  // user-user tertentu yang dipilih
)

type Announcement struct {
	ID             uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Title          string         `gorm:"type:varchar(255);not null" json:"title"`
	Content        string         `gorm:"type:text;not null" json:"content"`
	Category       string         `gorm:"type:varchar(50);default:'policy'" json:"category"` // policy, birthday, event, urgent_call, social_media, promo, operational
	Author         string         `gorm:"type:varchar(100)" json:"author"`
	IsActive       bool           `gorm:"default:true" json:"is_active"`
	TargetType     TargetType     `gorm:"type:varchar(50);default:'broadcast'" json:"target_type"`
	IsPopup        bool           `gorm:"default:false" json:"is_popup"`
	PopupExpiresAt *time.Time     `gorm:"type:timestamptz" json:"popup_expires_at"`

	// Media & Social Media Feed Extensions
	MediaType      string         `gorm:"type:varchar(50);default:'none'" json:"media_type"` // none, image, multi_image, video_url, instagram, tiktok, youtube
	ImageURL       string         `gorm:"type:text" json:"image_url"`
	MediaURLs      pq.StringArray `gorm:"type:text[]" json:"media_urls"`
	ExternalURL    string         `gorm:"type:text" json:"external_url"` // Link postingan asli medsos (Instagram, TikTok, YouTube, dll.)
	LikesCount     int            `gorm:"default:0" json:"likes_count"`
	CommentsCount  int            `gorm:"default:0" json:"comments_count"`
	IsLikedByMe    bool           `gorm:"-" json:"is_liked_by_me"`

	// Activity Tracking Options (ON/OFF)
	TrackDownloads bool           `gorm:"default:false" json:"track_downloads"`
	TrackShares    bool           `gorm:"default:false" json:"track_shares"`
	DownloadsCount int            `gorm:"default:0" json:"downloads_count"`
	SharesCount    int            `gorm:"default:0" json:"shares_count"`

	CreatedAt      time.Time      `gorm:"default:now()" json:"created_at"`
	UpdatedAt      time.Time      `gorm:"default:now()" json:"updated_at"`

	Targets        []AnnouncementTarget `gorm:"foreignKey:AnnouncementID" json:"targets,omitempty"`
	TotalTargets   int                  `gorm:"-" json:"total_targets"`
	ConfirmedCount int                  `gorm:"-" json:"confirmed_count"`
	DownloadedCount int                 `gorm:"-" json:"downloaded_count"`
	SharedCount     int                 `gorm:"-" json:"shared_count"`
}

// AnnouncementActivityLog mencatat aktivitas download atau share oleh user.
type AnnouncementActivityLog struct {
	ID             uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	AnnouncementID uuid.UUID `gorm:"type:uuid;not null;index:idx_ann_act_announcement_id" json:"announcement_id"`
	UserID         uuid.UUID `gorm:"type:uuid;not null;index:idx_ann_act_user_id" json:"user_id"`
	ActionType     string    `gorm:"type:varchar(50);not null" json:"action_type"` // 'download', 'share'
	Channel        string    `gorm:"type:varchar(50)" json:"channel"`             // 'whatsapp', 'general', 'media_file'
	CreatedAt      time.Time `gorm:"default:now()" json:"created_at"`
}

func (l *AnnouncementActivityLog) BeforeCreate(tx *gorm.DB) (err error) {
	if l.ID == uuid.Nil {
		l.ID = uuid.New()
	}
	return
}

func (a *Announcement) BeforeCreate(tx *gorm.DB) (err error) {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return
}

// AnnouncementComment menyimpan komentar user pada postingan/pengumuman.
type AnnouncementComment struct {
	ID             uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	AnnouncementID uuid.UUID `gorm:"type:uuid;not null;index:idx_announcement_comments_announcement_id" json:"announcement_id"`
	UserID         uuid.UUID `gorm:"type:uuid;not null;index:idx_announcement_comments_user_id" json:"user_id"`
	Content        string    `gorm:"type:text;not null" json:"content"`
	CreatedAt      time.Time `gorm:"default:now()" json:"created_at"`
	UpdatedAt      time.Time `gorm:"default:now()" json:"updated_at"`

	// Joined/Hydrated user information (read-only for GORM scan)
	UserName   string `gorm:"->" json:"user_name"`
	UserAvatar string `gorm:"->" json:"user_avatar"`
	UserRole   string `gorm:"->" json:"user_role"`
	IsMine     bool   `gorm:"-" json:"is_mine"`
}

func (c *AnnouncementComment) BeforeCreate(tx *gorm.DB) (err error) {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return
}

// AnnouncementLike mencatat user yang menyukai (like) pengumuman / feed post.
type AnnouncementLike struct {
	ID             uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	AnnouncementID uuid.UUID `gorm:"type:uuid;not null;index:idx_announcement_user_like,unique" json:"announcement_id"`
	UserID         uuid.UUID `gorm:"type:uuid;not null;index:idx_announcement_user_like,unique" json:"user_id"`
	CreatedAt      time.Time `gorm:"default:now()" json:"created_at"`
}

func (l *AnnouncementLike) BeforeCreate(tx *gorm.DB) (err error) {
	if l.ID == uuid.Nil {
		l.ID = uuid.New()
	}
	return
}

// AnnouncementTarget menyimpan target spesifik (user individu atau role) untuk notifikasi.
// Hanya satu dari target_user_id atau target_role yang boleh diisi per baris.
type AnnouncementTarget struct {
	ID               uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	AnnouncementID   uuid.UUID  `gorm:"type:uuid;not null" json:"announcement_id"`
	TargetUserID     *uuid.UUID `gorm:"type:uuid" json:"target_user_id"`
	TargetRole       string     `gorm:"type:varchar(50)" json:"target_role"`
	CreatedAt        time.Time  `gorm:"default:now()" json:"created_at"`
}

func (t *AnnouncementTarget) BeforeCreate(tx *gorm.DB) (err error) {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return
}

// AnnouncementRead mencatat bahwa user sudah membaca/mengakui popup.
type AnnouncementRead struct {
	ID             uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	AnnouncementID uuid.UUID `gorm:"type:uuid;not null" json:"announcement_id"`
	UserID         uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	AcknowledgedAt time.Time `gorm:"default:now()" json:"acknowledged_at"`
}

func (r *AnnouncementRead) BeforeCreate(tx *gorm.DB) (err error) {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return
}

// AnnouncementConfirmationItem detail per penerima
type AnnouncementConfirmationItem struct {
	UserID           uuid.UUID  `json:"user_id"`
	EmployeeID       *uuid.UUID `json:"employee_id,omitempty"`
	EmployeeName     string     `json:"employee_name"`
	EmployeeIDNumber string     `json:"employee_id_number"`
	Email            string     `json:"email"`
	Role             string     `json:"role"`
	DepartmentName   string     `json:"department_name"`
	JobTitle         string     `json:"job_title"`
	AvatarURL        string     `json:"avatar_url"`
	IsConfirmed      bool       `json:"is_confirmed"`
	AcknowledgedAt   *time.Time `json:"acknowledged_at"`

	// Activity Tracking Status
	IsDownloaded     bool       `json:"is_downloaded"`
	DownloadedAt     *time.Time `json:"downloaded_at"`
	IsShared         bool       `json:"is_shared"`
	SharedAt         *time.Time `json:"shared_at"`
}

// AnnouncementConfirmationSummary rekap tracking konfirmasi pengumuman
type AnnouncementConfirmationSummary struct {
	AnnouncementID   uuid.UUID                      `json:"announcement_id"`
	Title            string                         `json:"title"`
	Content          string                         `json:"content"`
	Category         string                         `json:"category"`
	TargetType       TargetType                     `json:"target_type"`
	IsPopup          bool                           `json:"is_popup"`
	TrackDownloads   bool                           `json:"track_downloads"`
	TrackShares      bool                           `json:"track_shares"`
	PopupExpiresAt   *time.Time                     `json:"popup_expires_at"`
	CreatedAt        time.Time                      `json:"created_at"`
	TotalTargets     int                            `json:"total_targets"`
	ConfirmedCount   int                            `json:"confirmed_count"`
	PendingCount     int                            `json:"pending_count"`
	ConfirmationRate float64                        `json:"confirmation_rate"`
	DownloadedCount  int                            `json:"downloaded_count"`
	SharedCount      int                            `json:"shared_count"`
	Recipients       []AnnouncementConfirmationItem `json:"recipients"`
}

