package repository

import (
	"time"

	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type AnnouncementRepository interface {
	FindAll() ([]domain.Announcement, error)
	FindAllForUser(userID uuid.UUID) ([]domain.Announcement, error)
	FindAllAdmin() ([]domain.Announcement, error)
	FindByID(id uuid.UUID) (*domain.Announcement, error)
	FindByIDForUser(id uuid.UUID, userID *uuid.UUID) (*domain.Announcement, error)
	Create(announcement *domain.Announcement) error
	CreateWithTargets(announcement *domain.Announcement, targets []domain.AnnouncementTarget) error
	Update(announcement *domain.Announcement) error
	UpdateWithTargets(announcement *domain.Announcement, targets []domain.AnnouncementTarget) error
	Delete(id uuid.UUID) error

	// Like & interaction
	ToggleLike(announcementID, userID uuid.UUID) (bool, int, error)

	// Comments
	GetComments(announcementID uuid.UUID, currentUserID *uuid.UUID, page, limit int) ([]domain.AnnouncementComment, int64, error)
	CreateComment(comment *domain.AnnouncementComment) error
	DeleteComment(announcementID, commentID, userID uuid.UUID, isAdmin bool) error

	// Targeting & popup
	AcknowledgeByUser(announcementID, userID uuid.UUID) error
	FindPendingPopupsForUser(userID uuid.UUID) ([]domain.Announcement, error)
	GetTargetUserIDsByAnnouncement(announcementID uuid.UUID) ([]uuid.UUID, error)
	GetUserIDsByRole(role string) ([]uuid.UUID, error)
	FindConfirmationsByAnnouncementID(announcementID uuid.UUID) (*domain.AnnouncementConfirmationSummary, error)
	FindPendingUserIDsByAnnouncementID(announcementID uuid.UUID) ([]uuid.UUID, error)
	FindAllTargetUserIDsByAnnouncementID(announcementID uuid.UUID) ([]uuid.UUID, error)

	// Activity Tracking
	RecordActivity(announcementID, userID uuid.UUID, actionType, channel string) error
}

type announcementRepository struct {
	db *gorm.DB
}

func NewAnnouncementRepository(db *gorm.DB) AnnouncementRepository {
	// Self-healing: Ensure announcements media columns, announcement_likes, and announcement_comments exist
	_ = db.Exec(`ALTER TABLE announcements ADD COLUMN IF NOT EXISTS media_type VARCHAR(50) DEFAULT 'none';`).Error
	_ = db.Exec(`ALTER TABLE announcements ADD COLUMN IF NOT EXISTS image_url TEXT;`).Error
	_ = db.Exec(`ALTER TABLE announcements ADD COLUMN IF NOT EXISTS media_urls TEXT[];`).Error
	_ = db.Exec(`ALTER TABLE announcements ADD COLUMN IF NOT EXISTS external_url TEXT;`).Error
	_ = db.Exec(`ALTER TABLE announcements ADD COLUMN IF NOT EXISTS likes_count INT DEFAULT 0;`).Error
	_ = db.Exec(`ALTER TABLE announcements ADD COLUMN IF NOT EXISTS comments_count INT DEFAULT 0;`).Error
	_ = db.Exec(`ALTER TABLE announcements ADD COLUMN IF NOT EXISTS track_downloads BOOLEAN DEFAULT false;`).Error
	_ = db.Exec(`ALTER TABLE announcements ADD COLUMN IF NOT EXISTS track_shares BOOLEAN DEFAULT false;`).Error
	_ = db.Exec(`ALTER TABLE announcements ADD COLUMN IF NOT EXISTS downloads_count INT DEFAULT 0;`).Error
	_ = db.Exec(`ALTER TABLE announcements ADD COLUMN IF NOT EXISTS shares_count INT DEFAULT 0;`).Error

	_ = db.Exec(`
		CREATE TABLE IF NOT EXISTS announcement_likes (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			announcement_id UUID NOT NULL,
			user_id UUID NOT NULL,
			created_at TIMESTAMPTZ DEFAULT now(),
			CONSTRAINT uq_announcement_user_like UNIQUE (announcement_id, user_id)
		);
	`).Error
	_ = db.Exec(`CREATE INDEX IF NOT EXISTS idx_announcement_likes_announcement_id ON announcement_likes(announcement_id);`).Error
	_ = db.Exec(`CREATE INDEX IF NOT EXISTS idx_announcement_likes_user_id ON announcement_likes(user_id);`).Error

	_ = db.Exec(`
		CREATE TABLE IF NOT EXISTS announcement_comments (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			announcement_id UUID NOT NULL,
			user_id UUID NOT NULL,
			content TEXT NOT NULL,
			created_at TIMESTAMPTZ DEFAULT now(),
			updated_at TIMESTAMPTZ DEFAULT now()
		);
	`).Error
	_ = db.Exec(`CREATE INDEX IF NOT EXISTS idx_announcement_comments_announcement_id ON announcement_comments(announcement_id);`).Error
	_ = db.Exec(`CREATE INDEX IF NOT EXISTS idx_announcement_comments_user_id ON announcement_comments(user_id);`).Error
	_ = db.Exec(`CREATE INDEX IF NOT EXISTS idx_announcement_comments_announcement_created ON announcement_comments(announcement_id, created_at ASC);`).Error

	_ = db.Exec(`
		CREATE TABLE IF NOT EXISTS announcement_activity_logs (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			announcement_id UUID NOT NULL,
			user_id UUID NOT NULL,
			action_type VARCHAR(50) NOT NULL,
			channel VARCHAR(50),
			created_at TIMESTAMPTZ DEFAULT now()
		);
	`).Error
	_ = db.Exec(`CREATE INDEX IF NOT EXISTS idx_ann_act_announcement_id ON announcement_activity_logs(announcement_id);`).Error
	_ = db.Exec(`CREATE INDEX IF NOT EXISTS idx_ann_act_user_id ON announcement_activity_logs(user_id);`).Error
	_ = db.Exec(`CREATE INDEX IF NOT EXISTS idx_ann_act_ann_action ON announcement_activity_logs(announcement_id, action_type);`).Error

	return &announcementRepository{db}
}

func (r *announcementRepository) FindAll() ([]domain.Announcement, error) {
	var announcements []domain.Announcement
	if err := r.db.Where("is_active = TRUE").Order("created_at desc").Find(&announcements).Error; err != nil {
		return nil, err
	}

	if len(announcements) == 0 {
		return announcements, nil
	}

	announcementIDs := make([]uuid.UUID, len(announcements))
	for i, a := range announcements {
		announcementIDs[i] = a.ID
	}

	type LikeCountResult struct {
		AnnouncementID uuid.UUID `gorm:"column:announcement_id"`
		Count          int       `gorm:"column:count"`
	}
	var likeCounts []LikeCountResult
	_ = r.db.Table("announcement_likes").
		Select("announcement_id, count(*) as count").
		Where("announcement_id IN ?", announcementIDs).
		Group("announcement_id").
		Scan(&likeCounts).Error

	likeMap := make(map[uuid.UUID]int)
	for _, lc := range likeCounts {
		likeMap[lc.AnnouncementID] = lc.Count
	}

	for i := range announcements {
		if count, exists := likeMap[announcements[i].ID]; exists {
			announcements[i].LikesCount = count
		}
		announcements[i].IsLikedByMe = false
	}

	return announcements, nil
}

// FindAllForUser mengembalikan seluruh pengumuman aktif yang relevan bagi user tertentu
// (broadcast, role/jabatan matching, atau targeted individual) beserta status is_liked_by_me.
func (r *announcementRepository) FindAllForUser(userID uuid.UUID) ([]domain.Announcement, error) {
	var announcements []domain.Announcement
	query := `
		SELECT a.*
		FROM announcements a
		WHERE a.is_active = TRUE
		  AND (
		    -- Broadcast: berlaku untuk semua user
		    COALESCE(NULLIF(CAST(a.target_type AS text), ''), 'broadcast') = 'broadcast'
		    OR
		    -- Individual: user ini secara eksplisit ditarget
		    EXISTS (
			      SELECT 1 FROM announcement_targets t
		      WHERE t.announcement_id = a.id
		        AND t.target_user_id = ?
		    )
		    OR
		    -- Role / Jabatan: role user atau jabatan karyawan cocok dengan target
		    EXISTS (
		      SELECT 1 FROM announcement_targets t
		      JOIN users u ON u.id = ?
		      LEFT JOIN employees e ON e.user_id = u.id
		      LEFT JOIN job_positions j ON j.id = e.job_position_id
		      WHERE t.announcement_id = a.id
		        AND (
		          u.role = t.target_role
		          OR j.title = t.target_role
		          OR CAST(j.id AS text) = t.target_role
		        )
		    )
		  )
		ORDER BY a.created_at DESC
	`
	if err := r.db.Raw(query, userID, userID).Scan(&announcements).Error; err != nil {
		return nil, err
	}

	if len(announcements) == 0 {
		return announcements, nil
	}

	announcementIDs := make([]uuid.UUID, len(announcements))
	for i, a := range announcements {
		announcementIDs[i] = a.ID
	}

	type LikeCountResult struct {
		AnnouncementID uuid.UUID `gorm:"column:announcement_id"`
		Count          int       `gorm:"column:count"`
	}
	var likeCounts []LikeCountResult
	_ = r.db.Table("announcement_likes").
		Select("announcement_id, count(*) as count").
		Where("announcement_id IN ?", announcementIDs).
		Group("announcement_id").
		Scan(&likeCounts).Error

	likeMap := make(map[uuid.UUID]int)
	for _, lc := range likeCounts {
		likeMap[lc.AnnouncementID] = lc.Count
	}

	var likedIDs []uuid.UUID
	_ = r.db.Table("announcement_likes").
		Where("announcement_id IN ? AND user_id = ?", announcementIDs, userID).
		Pluck("announcement_id", &likedIDs).Error

	likedMap := make(map[uuid.UUID]bool)
	for _, id := range likedIDs {
		likedMap[id] = true
	}

	for i := range announcements {
		if count, exists := likeMap[announcements[i].ID]; exists {
			announcements[i].LikesCount = count
		}
		announcements[i].IsLikedByMe = likedMap[announcements[i].ID]
	}

	return announcements, nil
}

func (r *announcementRepository) Create(announcement *domain.Announcement) error {
	return r.db.Create(announcement).Error
}

// CreateWithTargets menyimpan pengumuman beserta daftar target dalam satu transaksi.
func (r *announcementRepository) CreateWithTargets(announcement *domain.Announcement, targets []domain.AnnouncementTarget) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(announcement).Error; err != nil {
			return err
		}
		if len(targets) > 0 {
			for i := range targets {
				targets[i].AnnouncementID = announcement.ID
			}
			if err := tx.Create(&targets).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *announcementRepository) FindAllAdmin() ([]domain.Announcement, error) {
	var announcements []domain.Announcement
	if err := r.db.Preload("Targets").Order("created_at desc").Find(&announcements).Error; err != nil {
		return nil, err
	}

	for i := range announcements {
		var confirmedCount int64
		_ = r.db.Model(&domain.AnnouncementRead{}).Where("announcement_id = ?", announcements[i].ID).Count(&confirmedCount).Error
		announcements[i].ConfirmedCount = int(confirmedCount)

		if announcements[i].TargetType == domain.TargetBroadcast {
			var totalUsers int64
			_ = r.db.Table("users").
				Joins("LEFT JOIN employees ON employees.user_id = users.id").
				Where("users.is_active = TRUE").
				Where("(employees.id IS NULL OR COALESCE(employees.employment_status, 'Active') NOT IN ('Inactive', 'INACTIVE', 'Resigned', 'RESIGNED', 'Terminated', 'TERMINATED'))").
				Count(&totalUsers).Error
			announcements[i].TotalTargets = int(totalUsers)
		} else if announcements[i].TargetType == domain.TargetRole {
			var roles []string
			for _, t := range announcements[i].Targets {
				if t.TargetRole != "" {
					roles = append(roles, t.TargetRole)
				}
			}
			var totalUsers int64
			if len(roles) > 0 {
				_ = r.db.Table("users").
					Joins("LEFT JOIN employees ON employees.user_id = users.id").
					Joins("LEFT JOIN job_positions ON job_positions.id = employees.job_position_id").
					Where("users.is_active = TRUE").
					Where("(employees.id IS NULL OR COALESCE(employees.employment_status, 'Active') NOT IN ('Inactive', 'INACTIVE', 'Resigned', 'RESIGNED', 'Terminated', 'TERMINATED'))").
					Where("(users.role IN (?) OR job_positions.title IN (?) OR CAST(job_positions.id AS text) IN (?))", roles, roles, roles).
					Distinct("users.id").
					Count(&totalUsers).Error
			}
			announcements[i].TotalTargets = int(totalUsers)
		} else if announcements[i].TargetType == domain.TargetIndividual {
			var totalUsers int64
			var targetUserIDs []uuid.UUID
			for _, t := range announcements[i].Targets {
				if t.TargetUserID != nil {
					targetUserIDs = append(targetUserIDs, *t.TargetUserID)
				}
			}
			if len(targetUserIDs) > 0 {
				_ = r.db.Table("users").
					Joins("LEFT JOIN employees ON employees.user_id = users.id").
					Where("users.id IN (?) AND users.is_active = TRUE", targetUserIDs).
					Where("(employees.id IS NULL OR COALESCE(employees.employment_status, 'Active') NOT IN ('Inactive', 'INACTIVE', 'Resigned', 'RESIGNED', 'Terminated', 'TERMINATED'))").
					Count(&totalUsers).Error
			}
			announcements[i].TotalTargets = int(totalUsers)
		}
	}

	return announcements, nil
}


func (r *announcementRepository) FindByID(id uuid.UUID) (*domain.Announcement, error) {
	var announcement domain.Announcement
	if err := r.db.First(&announcement, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &announcement, nil
}

func (r *announcementRepository) FindByIDForUser(id uuid.UUID, userID *uuid.UUID) (*domain.Announcement, error) {
	var announcement domain.Announcement
	if err := r.db.First(&announcement, "id = ?", id).Error; err != nil {
		return nil, err
	}

	var totalLikes int64
	_ = r.db.Model(&domain.AnnouncementLike{}).Where("announcement_id = ?", id).Count(&totalLikes).Error
	announcement.LikesCount = int(totalLikes)

	if userID != nil {
		var count int64
		_ = r.db.Model(&domain.AnnouncementLike{}).Where("announcement_id = ? AND user_id = ?", id, *userID).Count(&count).Error
		announcement.IsLikedByMe = count > 0
	}

	return &announcement, nil
}

func (r *announcementRepository) Update(announcement *domain.Announcement) error {
	return r.db.Save(announcement).Error
}

func (r *announcementRepository) UpdateWithTargets(announcement *domain.Announcement, targets []domain.AnnouncementTarget) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		updateMap := map[string]interface{}{
			"title":            announcement.Title,
			"content":          announcement.Content,
			"category":         announcement.Category,
			"is_active":        announcement.IsActive,
			"target_type":      announcement.TargetType,
			"is_popup":         announcement.IsPopup,
			"popup_expires_at": announcement.PopupExpiresAt,
			"media_type":       announcement.MediaType,
			"image_url":        announcement.ImageURL,
			"media_urls":       announcement.MediaURLs,
			"external_url":     announcement.ExternalURL,
			"updated_at":       time.Now(),
		}

		if err := tx.Model(&domain.Announcement{}).Where("id = ?", announcement.ID).Updates(updateMap).Error; err != nil {
			return err
		}

		if err := tx.Where("announcement_id = ?", announcement.ID).Delete(&domain.AnnouncementTarget{}).Error; err != nil {
			return err
		}

		if len(targets) > 0 {
			for i := range targets {
				targets[i].AnnouncementID = announcement.ID
			}
			if err := tx.Create(&targets).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *announcementRepository) Delete(id uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Clean up all child tables defensively to prevent foreign key errors & orphaned records
		_ = tx.Exec("DELETE FROM announcement_comments WHERE announcement_id = ?", id).Error
		_ = tx.Exec("DELETE FROM announcement_likes WHERE announcement_id = ?", id).Error
		_ = tx.Exec("DELETE FROM announcement_reads WHERE announcement_id = ?", id).Error
		_ = tx.Exec("DELETE FROM announcement_targets WHERE announcement_id = ?", id).Error

		if err := tx.Exec("DELETE FROM announcements WHERE id = ?", id).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *announcementRepository) ToggleLike(announcementID, userID uuid.UUID) (bool, int, error) {
	var isLiked bool
	var totalLikes int64

	// Self-healing: ensure table exists defensively
	_ = r.db.Exec(`
		CREATE TABLE IF NOT EXISTS announcement_likes (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			announcement_id UUID NOT NULL,
			user_id UUID NOT NULL,
			created_at TIMESTAMPTZ DEFAULT now(),
			CONSTRAINT uq_announcement_user_like UNIQUE (announcement_id, user_id)
		);
	`).Error

	err := r.db.Transaction(func(tx *gorm.DB) error {
		var existing domain.AnnouncementLike
		err := tx.Table("announcement_likes").
			Where("announcement_id = ? AND user_id = ?", announcementID, userID).
			First(&existing).Error

		if err == nil {
			// Sudah di-like sebelumnya -> Unlike
			if err := tx.Table("announcement_likes").
				Where("announcement_id = ? AND user_id = ?", announcementID, userID).
				Delete(&domain.AnnouncementLike{}).Error; err != nil {
				return err
			}
			isLiked = false
		} else if err == gorm.ErrRecordNotFound {
			// Belum di-like -> Tambahkan Like
			newLike := domain.AnnouncementLike{
				ID:             uuid.New(),
				AnnouncementID: announcementID,
				UserID:         userID,
				CreatedAt:      time.Now(),
			}
			if err := tx.Table("announcement_likes").Create(&newLike).Error; err != nil {
				return err
			}
			isLiked = true
		} else {
			return err
		}

		// Hitung ulang total like
		if err := tx.Table("announcement_likes").
			Where("announcement_id = ?", announcementID).
			Count(&totalLikes).Error; err != nil {
			return err
		}

		// Update cache likes_count pada tabel announcements
		_ = tx.Table("announcements").
			Where("id = ?", announcementID).
			Update("likes_count", totalLikes).Error

		return nil
	})

	return isLiked, int(totalLikes), err
}

func (r *announcementRepository) GetComments(announcementID uuid.UUID, currentUserID *uuid.UUID, page, limit int) ([]domain.AnnouncementComment, int64, error) {
	if limit <= 0 {
		limit = 20
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	var total int64
	_ = r.db.Table("announcement_comments").Where("announcement_id = ?", announcementID).Count(&total).Error

	var comments []domain.AnnouncementComment
	query := `
		SELECT c.id, c.announcement_id, c.user_id, c.content, c.created_at, c.updated_at,
		       COALESCE(NULLIF(TRIM(CONCAT(COALESCE(e.first_name, ''), ' ', COALESCE(e.last_name, ''))), ''), u.email, 'Karyawan') AS user_name,
		       COALESCE(e.face_reference_url, '') AS user_avatar,
		       COALESCE(jp.title, CAST(u.role AS text), 'Karyawan') AS user_role
		FROM announcement_comments c
		LEFT JOIN users u ON u.id = c.user_id
		LEFT JOIN employees e ON e.user_id = u.id
		LEFT JOIN job_positions jp ON jp.id = e.job_position_id
		WHERE c.announcement_id = ?
		ORDER BY c.created_at ASC
		LIMIT ? OFFSET ?
	`
	if err := r.db.Raw(query, announcementID, limit, offset).Scan(&comments).Error; err != nil {
		return nil, 0, err
	}

	for i := range comments {
		if currentUserID != nil && comments[i].UserID == *currentUserID {
			comments[i].IsMine = true
		}
	}

	return comments, total, nil
}

func (r *announcementRepository) CreateComment(comment *domain.AnnouncementComment) error {
	// Self-healing: ensure table exists
	_ = r.db.Exec(`
		CREATE TABLE IF NOT EXISTS announcement_comments (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			announcement_id UUID NOT NULL,
			user_id UUID NOT NULL,
			content TEXT NOT NULL,
			created_at TIMESTAMPTZ DEFAULT now(),
			updated_at TIMESTAMPTZ DEFAULT now()
		);
	`).Error

	return r.db.Transaction(func(tx *gorm.DB) error {
		if comment.ID == uuid.Nil {
			comment.ID = uuid.New()
		}
		if comment.CreatedAt.IsZero() {
			comment.CreatedAt = time.Now()
		}
		if comment.UpdatedAt.IsZero() {
			comment.UpdatedAt = time.Now()
		}

		if err := tx.Table("announcement_comments").Create(comment).Error; err != nil {
			return err
		}

		var count int64
		_ = tx.Table("announcement_comments").Where("announcement_id = ?", comment.AnnouncementID).Count(&count).Error
		_ = tx.Table("announcements").Where("id = ?", comment.AnnouncementID).Update("comments_count", count).Error

		// Hydrate user info
		var userProfile struct {
			UserName   string `gorm:"column:user_name"`
			UserAvatar string `gorm:"column:user_avatar"`
			UserRole   string `gorm:"column:user_role"`
		}
		_ = tx.Raw(`
			SELECT COALESCE(NULLIF(TRIM(CONCAT(COALESCE(e.first_name, ''), ' ', COALESCE(e.last_name, ''))), ''), u.email, 'Karyawan') AS user_name,
			       COALESCE(e.face_reference_url, '') AS user_avatar,
			       COALESCE(jp.title, CAST(u.role AS text), 'Karyawan') AS user_role
			FROM users u
			LEFT JOIN employees e ON e.user_id = u.id
			LEFT JOIN job_positions jp ON jp.id = e.job_position_id
			WHERE u.id = ?
		`, comment.UserID).Scan(&userProfile).Error

		comment.UserName = userProfile.UserName
		comment.UserAvatar = userProfile.UserAvatar
		comment.UserRole = userProfile.UserRole
		comment.IsMine = true

		return nil
	})
}

func (r *announcementRepository) DeleteComment(announcementID, commentID, userID uuid.UUID, isAdmin bool) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var comment domain.AnnouncementComment
		if err := tx.Table("announcement_comments").Where("id = ? AND announcement_id = ?", commentID, announcementID).First(&comment).Error; err != nil {
			return err
		}

		if !isAdmin && comment.UserID != userID {
			return gorm.ErrInvalidData
		}

		if err := tx.Table("announcement_comments").Where("id = ?", commentID).Delete(&domain.AnnouncementComment{}).Error; err != nil {
			return err
		}

		var count int64
		_ = tx.Table("announcement_comments").Where("announcement_id = ?", announcementID).Count(&count).Error
		_ = tx.Table("announcements").Where("id = ?", announcementID).Update("comments_count", count).Error

		return nil
	})
}

// AcknowledgeByUser mencatat bahwa user sudah membaca popup.
// Menggunakan ON CONFLICT DO NOTHING agar idempotent (klik OK berkali-kali tidak error).
func (r *announcementRepository) AcknowledgeByUser(announcementID, userID uuid.UUID) error {
	read := domain.AnnouncementRead{
		AnnouncementID: announcementID,
		UserID:         userID,
	}
	return r.db.Clauses(clause.OnConflict{DoNothing: true}).Create(&read).Error
}

// FindPendingPopupsForUser mengembalikan popup aktif yang belum di-acknowledge oleh user.
// Mendukung semua target_type: broadcast, role/jabatan, dan individual (termasuk legacy NULL/empty target_type).
func (r *announcementRepository) FindPendingPopupsForUser(userID uuid.UUID) ([]domain.Announcement, error) {
	var announcements []domain.Announcement
	err := r.db.Raw(`
		SELECT a.*
		FROM announcements a
		WHERE a.is_popup = TRUE
		  AND a.is_active = TRUE
		  AND (a.popup_expires_at IS NULL OR a.popup_expires_at > NOW())
		  AND (
		    -- Broadcast: berlaku untuk semua user (termasuk legacy NULL atau empty string)
		    COALESCE(NULLIF(CAST(a.target_type AS text), ''), 'broadcast') = 'broadcast'
		    OR
		    -- Individual: user ini secara eksplisit ditarget
		    EXISTS (
		      SELECT 1 FROM announcement_targets t
		      WHERE t.announcement_id = a.id
		        AND t.target_user_id = ?
		    )
		    OR
		    -- Role / Jabatan: role user atau jabatan karyawan cocok dengan target
		    EXISTS (
		      SELECT 1 FROM announcement_targets t
		      JOIN users u ON u.id = ?
		      LEFT JOIN employees e ON e.user_id = u.id
		      LEFT JOIN job_positions j ON j.id = e.job_position_id
		      WHERE t.announcement_id = a.id
		        AND (
		          u.role = t.target_role
		          OR j.title = t.target_role
		          OR CAST(j.id AS text) = t.target_role
		        )
		    )
		  )
		  -- Belum di-acknowledge oleh user ini
		  AND NOT EXISTS (
		    SELECT 1 FROM announcement_reads r
		    WHERE r.announcement_id = a.id
		      AND r.user_id = ?
		  )
		ORDER BY a.created_at DESC
	`, userID, userID, userID).Scan(&announcements).Error

	return announcements, err
}

// GetTargetUserIDsByAnnouncement mengambil semua user_id yang ditarget secara individual
// untuk sebuah pengumuman.
func (r *announcementRepository) GetTargetUserIDsByAnnouncement(announcementID uuid.UUID) ([]uuid.UUID, error) {
	var userIDs []uuid.UUID
	err := r.db.Model(&domain.AnnouncementTarget{}).
		Where("announcement_id = ? AND target_user_id IS NOT NULL", announcementID).
		Pluck("target_user_id", &userIDs).Error
	return userIDs, err
}

// GetUserIDsByRole mengambil semua user_id yang memiliki role atau jabatan tertentu (hanya karyawan aktif).
func (r *announcementRepository) GetUserIDsByRole(role string) ([]uuid.UUID, error) {
	var userIDs []uuid.UUID
	query := `
		SELECT DISTINCT u.id
		FROM users u
		LEFT JOIN employees e ON e.user_id = u.id
		LEFT JOIN job_positions j ON j.id = e.job_position_id
		WHERE u.is_active = TRUE
		  AND (e.id IS NULL OR COALESCE(e.employment_status, 'Active') NOT IN ('Inactive', 'INACTIVE', 'Resigned', 'RESIGNED', 'Terminated', 'TERMINATED'))
		  AND (
		    u.role = ?
		    OR j.title = ?
		    OR CAST(j.id AS text) = ?
		  )
	`
	err := r.db.Raw(query, role, role, role).Scan(&userIDs).Error
	return userIDs, err
}

func (r *announcementRepository) FindConfirmationsByAnnouncementID(announcementID uuid.UUID) (*domain.AnnouncementConfirmationSummary, error) {
	var ann domain.Announcement
	if err := r.db.Preload("Targets").First(&ann, "id = ?", announcementID).Error; err != nil {
		return nil, err
	}

	targetTypeStr := string(ann.TargetType)
	if targetTypeStr == "" {
		targetTypeStr = "broadcast"
	}

	var items []domain.AnnouncementConfirmationItem
	query := `
		SELECT 
			u.id AS user_id,
			e.id AS employee_id,
			COALESCE(NULLIF(TRIM(CONCAT(e.first_name, ' ', COALESCE(e.last_name, ''))), ''), u.email) AS employee_name,
			COALESCE(e.employee_id_number, '-') AS employee_id_number,
			u.email AS email,
			CAST(u.role AS text) AS role,
			COALESCE(d.name, '-') AS department_name,
			COALESCE(j.title, '-') AS job_title,
			COALESCE(e.face_reference_url, '') AS avatar_url,
			(ar.id IS NOT NULL) AS is_confirmed,
			ar.acknowledged_at AS acknowledged_at,
			(dl.user_id IS NOT NULL) AS is_downloaded,
			dl.downloaded_at AS downloaded_at,
			(sh.user_id IS NOT NULL) AS is_shared,
			sh.shared_at AS shared_at
		FROM users u
		LEFT JOIN employees e ON e.user_id = u.id
		LEFT JOIN departments d ON d.id = e.department_id
		LEFT JOIN job_positions j ON j.id = e.job_position_id
		LEFT JOIN announcement_reads ar ON ar.announcement_id = ? AND ar.user_id = u.id
		LEFT JOIN (
			SELECT user_id, MIN(created_at) AS downloaded_at
			FROM announcement_activity_logs
			WHERE announcement_id = ? AND action_type = 'download'
			GROUP BY user_id
		) dl ON dl.user_id = u.id
		LEFT JOIN (
			SELECT user_id, MIN(created_at) AS shared_at
			FROM announcement_activity_logs
			WHERE announcement_id = ? AND action_type = 'share'
			GROUP BY user_id
		) sh ON sh.user_id = u.id
		WHERE u.is_active = TRUE
		  AND (e.id IS NULL OR COALESCE(e.employment_status, 'Active') NOT IN ('Inactive', 'INACTIVE', 'Resigned', 'RESIGNED', 'Terminated', 'TERMINATED'))
		  AND (
		    ? = 'broadcast'
		    OR (? = 'individual' AND u.id IN (
		        SELECT target_user_id FROM announcement_targets WHERE announcement_id = ? AND target_user_id IS NOT NULL
		    ))
		    OR (? = 'role' AND (
		        u.role IN (SELECT target_role FROM announcement_targets WHERE announcement_id = ? AND target_role IS NOT NULL AND target_role != '')
		        OR j.title IN (SELECT target_role FROM announcement_targets WHERE announcement_id = ? AND target_role IS NOT NULL AND target_role != '')
		        OR CAST(j.id AS text) IN (SELECT target_role FROM announcement_targets WHERE announcement_id = ? AND target_role IS NOT NULL AND target_role != '')
		    ))
		  )
		ORDER BY (ar.id IS NOT NULL) DESC, (dl.user_id IS NOT NULL) DESC, (sh.user_id IS NOT NULL) DESC, ar.acknowledged_at DESC, e.first_name ASC, u.email ASC
	`

	if err := r.db.Raw(query, ann.ID, ann.ID, ann.ID, targetTypeStr, targetTypeStr, ann.ID, targetTypeStr, ann.ID, ann.ID, ann.ID).Scan(&items).Error; err != nil {
		return nil, err
	}

	total := len(items)
	confirmed := 0
	downloaded := 0
	shared := 0
	for _, item := range items {
		if item.IsConfirmed {
			confirmed++
		}
		if item.IsDownloaded {
			downloaded++
		}
		if item.IsShared {
			shared++
		}
	}
	pending := total - confirmed
	var rate float64 = 0
	if total > 0 {
		rate = float64(confirmed) / float64(total) * 100
	}

	summary := &domain.AnnouncementConfirmationSummary{
		AnnouncementID:   ann.ID,
		Title:            ann.Title,
		Content:          ann.Content,
		Category:         ann.Category,
		TargetType:       ann.TargetType,
		IsPopup:          ann.IsPopup,
		TrackDownloads:   ann.TrackDownloads,
		TrackShares:      ann.TrackShares,
		PopupExpiresAt:   ann.PopupExpiresAt,
		CreatedAt:        ann.CreatedAt,
		TotalTargets:     total,
		ConfirmedCount:   confirmed,
		PendingCount:     pending,
		ConfirmationRate: rate,
		DownloadedCount:  downloaded,
		SharedCount:      shared,
		Recipients:       items,
	}

	return summary, nil
}

func (r *announcementRepository) FindPendingUserIDsByAnnouncementID(announcementID uuid.UUID) ([]uuid.UUID, error) {
	var ann domain.Announcement
	if err := r.db.First(&ann, "id = ?", announcementID).Error; err != nil {
		return nil, err
	}

	targetTypeStr := string(ann.TargetType)
	if targetTypeStr == "" {
		targetTypeStr = "broadcast"
	}

	var userIDs []uuid.UUID
	query := `
		SELECT u.id
		FROM users u
		LEFT JOIN employees e ON e.user_id = u.id
		LEFT JOIN job_positions j ON j.id = e.job_position_id
		LEFT JOIN announcement_reads ar ON ar.announcement_id = ? AND ar.user_id = u.id
		WHERE u.is_active = TRUE
		  AND (e.id IS NULL OR COALESCE(e.employment_status, 'Active') NOT IN ('Inactive', 'INACTIVE', 'Resigned', 'RESIGNED', 'Terminated', 'TERMINATED'))
		  AND ar.id IS NULL
		  AND (
		    ? = 'broadcast'
		    OR (? = 'individual' AND u.id IN (
		        SELECT target_user_id FROM announcement_targets WHERE announcement_id = ? AND target_user_id IS NOT NULL
		    ))
		    OR (? = 'role' AND (
		        u.role IN (SELECT target_role FROM announcement_targets WHERE announcement_id = ? AND target_role IS NOT NULL AND target_role != '')
		        OR j.title IN (SELECT target_role FROM announcement_targets WHERE announcement_id = ? AND target_role IS NOT NULL AND target_role != '')
		        OR CAST(j.id AS text) IN (SELECT target_role FROM announcement_targets WHERE announcement_id = ? AND target_role IS NOT NULL AND target_role != '')
		    ))
		  )
	`
	err := r.db.Raw(query, ann.ID, targetTypeStr, targetTypeStr, ann.ID, targetTypeStr, ann.ID, ann.ID, ann.ID).Scan(&userIDs).Error
	return userIDs, err
}

func (r *announcementRepository) FindAllTargetUserIDsByAnnouncementID(announcementID uuid.UUID) ([]uuid.UUID, error) {
	var ann domain.Announcement
	if err := r.db.First(&ann, "id = ?", announcementID).Error; err != nil {
		return nil, err
	}

	targetTypeStr := string(ann.TargetType)
	if targetTypeStr == "" {
		targetTypeStr = "broadcast"
	}

	var userIDs []uuid.UUID
	query := `
		SELECT u.id
		FROM users u
		LEFT JOIN employees e ON e.user_id = u.id
		LEFT JOIN job_positions j ON j.id = e.job_position_id
		WHERE u.is_active = TRUE
		  AND (e.id IS NULL OR COALESCE(e.employment_status, 'Active') NOT IN ('Inactive', 'INACTIVE', 'Resigned', 'RESIGNED', 'Terminated', 'TERMINATED'))
		  AND (
		    ? = 'broadcast'
		    OR (? = 'individual' AND u.id IN (
		        SELECT target_user_id FROM announcement_targets WHERE announcement_id = ? AND target_user_id IS NOT NULL
		    ))
		    OR (? = 'role' AND (
		        u.role IN (SELECT target_role FROM announcement_targets WHERE announcement_id = ? AND target_role IS NOT NULL AND target_role != '')
		        OR j.title IN (SELECT target_role FROM announcement_targets WHERE announcement_id = ? AND target_role IS NOT NULL AND target_role != '')
		        OR CAST(j.id AS text) IN (SELECT target_role FROM announcement_targets WHERE announcement_id = ? AND target_role IS NOT NULL AND target_role != '')
		    ))
		  )
	`
	err := r.db.Raw(query, targetTypeStr, targetTypeStr, ann.ID, targetTypeStr, ann.ID, ann.ID, ann.ID).Scan(&userIDs).Error
	return userIDs, err
}

// RecordActivity mencatat log interaksi download atau share oleh karyawan.
func (r *announcementRepository) RecordActivity(announcementID, userID uuid.UUID, actionType, channel string) error {
	log := domain.AnnouncementActivityLog{
		AnnouncementID: announcementID,
		UserID:         userID,
		ActionType:     actionType,
		Channel:        channel,
	}

	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&log).Error; err != nil {
			return err
		}

		if actionType == "download" {
			_ = tx.Model(&domain.Announcement{}).
				Where("id = ?", announcementID).
				UpdateColumn("downloads_count", gorm.Expr("downloads_count + 1")).Error
		} else if actionType == "share" {
			_ = tx.Model(&domain.Announcement{}).
				Where("id = ?", announcementID).
				UpdateColumn("shares_count", gorm.Expr("shares_count + 1")).Error
		}

		return nil
	})
}


