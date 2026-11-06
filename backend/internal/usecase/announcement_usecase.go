package usecase

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"example.com/hris/backend/internal/repository"
	"example.com/hris/backend/pkg/sse"
)

type AnnouncementUseCase interface {
	GetAnnouncements(userID *uuid.UUID) ([]domain.Announcement, error)
	GetAnnouncementByID(id uuid.UUID, userID *uuid.UUID) (*domain.Announcement, error)
	AdminListAnnouncements() ([]domain.Announcement, error)
	CreateAnnouncement(req CreateAnnouncementRequest) error
	UpdateAnnouncement(id uuid.UUID, req UpdateAnnouncementRequest) error
	DeleteAnnouncement(id uuid.UUID) error
	// Interaction
	ToggleLike(announcementID, userID uuid.UUID) (bool, int, error)
	GetComments(announcementID uuid.UUID, currentUserID *uuid.UUID, page, limit int) ([]domain.AnnouncementComment, int64, error)
	AddComment(announcementID, userID uuid.UUID, content string) (*domain.AnnouncementComment, error)
	DeleteComment(announcementID, commentID, userID uuid.UUID, isAdmin bool) error
	// Popup acknowledge & pending popup polling
	AcknowledgeAnnouncement(announcementID, userID uuid.UUID) error
	GetPendingPopups(userID uuid.UUID) ([]domain.Announcement, error)
	GetAnnouncementConfirmations(announcementID uuid.UUID) (*domain.AnnouncementConfirmationSummary, error)
	ResendAnnouncementFCM(announcementID uuid.UUID, mode string, targetUserID *uuid.UUID) (int, error)
	RecordActivity(announcementID, userID uuid.UUID, actionType, channel string) error
}

// CreateAnnouncementRequest berisi semua informasi yang dibutuhkan untuk membuat pengumuman baru,
// termasuk konfigurasi targeting, popup, serta media & link sosmed.
type CreateAnnouncementRequest struct {
	Title    string `json:"title" binding:"required"`
	Content  string `json:"content" binding:"required"`
	Category string `json:"category" binding:"required"`

	// Media & Social Media Feed
	MediaType   string   `json:"media_type"`   // image, multi_image, video_url, instagram, tiktok, youtube, none
	ImageURL    string   `json:"image_url"`    // Single image / primary thumbnail
	MediaURLs   []string `json:"media_urls"`   // Multiple images carousel
	ExternalURL string   `json:"external_url"` // Link postingan medsos

	// Activity Tracking Options
	TrackDownloads bool `json:"track_downloads"`
	TrackShares    bool `json:"track_shares"`

	// Targeting
	// TargetType: "broadcast" (default) | "role" | "individual"
	TargetType string `json:"target_type"`

	// Diisi jika TargetType = "role". Bisa lebih dari satu role.
	// Contoh: ["supervisor_sales", "manager"]
	TargetRoles []string `json:"target_roles"`

	// Diisi jika TargetType = "individual". Bisa lebih dari satu user.
	TargetUserIDs []string `json:"target_user_ids"`

	// Popup settings
	IsPopup        bool       `json:"is_popup"`
	PopupExpiresAt *time.Time `json:"popup_expires_at"`
}

type UpdateAnnouncementRequest struct {
	Title          string     `json:"title" binding:"required"`
	Content        string     `json:"content" binding:"required"`
	Category       string     `json:"category" binding:"required"`
	IsActive       bool       `json:"is_active"`
	MediaType      string     `json:"media_type"`
	ImageURL       string     `json:"image_url"`
	MediaURLs      []string   `json:"media_urls"`
	ExternalURL    string     `json:"external_url"`
	TrackDownloads bool       `json:"track_downloads"`
	TrackShares    bool       `json:"track_shares"`
	TargetType     string     `json:"target_type"`
	TargetRoles    []string   `json:"target_roles"`
	TargetUserIDs  []string   `json:"target_user_ids"`
	IsPopup        bool       `json:"is_popup"`
	PopupExpiresAt *time.Time `json:"popup_expires_at"`
}

type announcementUseCase struct {
	announcementRepo repository.AnnouncementRepository
	employeeRepo     repository.EmployeeRepository
	fcmUseCase       FcmUsecase
	sseHub           *sse.Hub
}

func NewAnnouncementUseCase(announcementRepo repository.AnnouncementRepository, employeeRepo repository.EmployeeRepository, fcmUseCase FcmUsecase, sseHub *sse.Hub) AnnouncementUseCase {
	return &announcementUseCase{
		announcementRepo: announcementRepo,
		employeeRepo:     employeeRepo,
		fcmUseCase:       fcmUseCase,
		sseHub:           sseHub,
	}
}

func (u *announcementUseCase) GetAnnouncements(userID *uuid.UUID) ([]domain.Announcement, error) {
	var formal []domain.Announcement
	var err error
	if userID != nil {
		formal, err = u.announcementRepo.FindAllForUser(*userID)
	} else {
		formal, err = u.announcementRepo.FindAll()
	}
	if err != nil {
		return nil, err
	}

	birthdayEmployees, err := u.employeeRepo.FindBirthdaysToday()
	if err != nil {
		fmt.Printf("Error fetching birthdays: %v\n", err)
	}

	for _, emp := range birthdayEmployees {
		birthdayAnn := domain.Announcement{
			ID:        uuid.New(),
			Title:     "Selamat Ulang Tahun!",
			Content:   fmt.Sprintf("Selamat ulang tahun untuk %s! Mari berikan ucapan terbaik dan rayakan hari spesial ini bersama tim.", emp.FirstName),
			Category:  "birthday",
			Author:    "System",
			CreatedAt: time.Now(),
		}
		formal = append([]domain.Announcement{birthdayAnn}, formal...)
	}

	return formal, nil
}

func (u *announcementUseCase) GetAnnouncementByID(id uuid.UUID, userID *uuid.UUID) (*domain.Announcement, error) {
	return u.announcementRepo.FindByIDForUser(id, userID)
}

func (u *announcementUseCase) ToggleLike(announcementID, userID uuid.UUID) (bool, int, error) {
	return u.announcementRepo.ToggleLike(announcementID, userID)
}

func (u *announcementUseCase) GetComments(announcementID uuid.UUID, currentUserID *uuid.UUID, page, limit int) ([]domain.AnnouncementComment, int64, error) {
	return u.announcementRepo.GetComments(announcementID, currentUserID, page, limit)
}

func (u *announcementUseCase) AddComment(announcementID, userID uuid.UUID, content string) (*domain.AnnouncementComment, error) {
	trimmed := strings.TrimSpace(content)
	if trimmed == "" {
		return nil, errors.New("komentar tidak boleh kosong")
	}
	if len(trimmed) > 1000 {
		return nil, errors.New("komentar maksimal 1000 karakter")
	}

	comment := &domain.AnnouncementComment{
		AnnouncementID: announcementID,
		UserID:         userID,
		Content:        trimmed,
	}

	if err := u.announcementRepo.CreateComment(comment); err != nil {
		return nil, err
	}

	return comment, nil
}

func (u *announcementUseCase) DeleteComment(announcementID, commentID, userID uuid.UUID, isAdmin bool) error {
	return u.announcementRepo.DeleteComment(announcementID, commentID, userID, isAdmin)
}

func (u *announcementUseCase) AdminListAnnouncements() ([]domain.Announcement, error) {
	return u.announcementRepo.FindAllAdmin()
}

// CreateAnnouncement membuat pengumuman baru dan mengirim FCM ke target yang ditentukan.
// Mendukung postingan media/gambar, link video/sosmed, targeting broadcast/role/individual.
func (u *announcementUseCase) CreateAnnouncement(req CreateAnnouncementRequest) error {
	// Normalisasi target_type
	targetType := domain.TargetType(req.TargetType)
	if targetType == "" {
		targetType = domain.TargetBroadcast
	}

	primaryImage := req.ImageURL
	mediaURLs := req.MediaURLs
	if primaryImage == "" && len(mediaURLs) > 0 {
		primaryImage = mediaURLs[0]
	}
	if len(mediaURLs) == 0 && primaryImage != "" {
		mediaURLs = []string{primaryImage}
	}

	cleanURL := strings.ToLower(strings.TrimSpace(primaryImage))
	isVideoFile := strings.HasSuffix(cleanURL, ".mp4") || strings.HasSuffix(cleanURL, ".mov") ||
		strings.HasSuffix(cleanURL, ".webm") || strings.HasSuffix(cleanURL, ".m4v") ||
		strings.Contains(cleanURL, "drive.google.com") || strings.Contains(cleanURL, "/file/d/")

	mediaType := req.MediaType
	if mediaType == "" || (mediaType == "image" && isVideoFile) {
		if isVideoFile {
			mediaType = "video"
		} else if len(req.MediaURLs) > 1 {
			mediaType = "multi_image"
		} else if len(req.MediaURLs) == 1 || req.ImageURL != "" {
			mediaType = "image"
		} else if req.ExternalURL != "" {
			mediaType = "social_media"
		} else {
			mediaType = "none"
		}
	}

	announcement := &domain.Announcement{
		Title:          req.Title,
		Content:        req.Content,
		Category:       req.Category,
		IsActive:       true,
		TargetType:     targetType,
		IsPopup:        req.IsPopup,
		PopupExpiresAt: req.PopupExpiresAt,
		MediaType:      mediaType,
		ImageURL:       primaryImage,
		MediaURLs:      mediaURLs,
		ExternalURL:    req.ExternalURL,
		TrackDownloads: req.TrackDownloads,
		TrackShares:    req.TrackShares,
	}

	// Build daftar target rows
	var targets []domain.AnnouncementTarget

	switch targetType {
	case domain.TargetRole:
		for _, role := range req.TargetRoles {
			if role != "" {
				targets = append(targets, domain.AnnouncementTarget{
					TargetRole: role,
				})
			}
		}

	case domain.TargetIndividual:
		for _, idStr := range req.TargetUserIDs {
			uid, err := uuid.Parse(idStr)
			if err != nil {
				continue // skip ID yang tidak valid
			}
			targets = append(targets, domain.AnnouncementTarget{
				TargetUserID: &uid,
			})
		}
	}

	// Simpan pengumuman + targets dalam satu transaksi
	if err := u.announcementRepo.CreateWithTargets(announcement, targets); err != nil {
		return err
	}

	// Resolve daftar user_id yang akan menerima FCM
	targetUserIDs, err := u.resolveTargetUserIDs(announcement.ID, targetType, req)
	if err != nil {
		// Log error tapi jangan gagalkan pembuatan pengumuman
		fmt.Printf("[ANNOUNCEMENT] Warning: gagal resolve target user IDs: %v\n", err)
	}

	// Siapkan body notifikasi (potong jika terlalu panjang)
	body := req.Content
	if len(body) > 100 {
		body = body[:97] + "..."
	}

	fcmTitle := "Pengumuman Baru: " + req.Title
	if req.Category == "urgent_call" || req.IsPopup {
		fcmTitle = "⚠️ " + req.Title
	} else if req.Category == "social_media" || req.Category == "promo" {
		fcmTitle = "📢 Pengumuman Perusahaan: " + req.Title
	}

	// Tentukan tipe notifikasi untuk mobile
	notifType := "notification"
	if req.IsPopup {
		notifType = "popup"
	}

	fcmData := map[string]string{
		"type":            notifType,
		"announcement_id": announcement.ID.String(),
		"category":        req.Category,
		"media_type":      mediaType,
		"image_url":       primaryImage,
		"external_url":    req.ExternalURL,
	}

	// Kirim FCM
	if targetType == domain.TargetBroadcast || len(targetUserIDs) == 0 {
		// Broadcast ke semua
		_ = u.fcmUseCase.BroadcastNotification(fcmTitle, body)
	} else {
		// Targeted
		_ = u.fcmUseCase.SendToUsers(targetUserIDs, fcmTitle, body, fcmData)
	}

	// Publish ke Redis Pub/Sub untuk realtime notification & auto popup di Web Manajemen
	u.publishRealtimeNotification(announcement, fcmTitle, body, req.TargetRoles, targetUserIDs)

	return nil
}

// resolveTargetUserIDs mengumpulkan semua user_id yang menjadi target berdasarkan targeting config.
func (u *announcementUseCase) resolveTargetUserIDs(
	announcementID uuid.UUID,
	targetType domain.TargetType,
	req CreateAnnouncementRequest,
) ([]uuid.UUID, error) {
	var userIDs []uuid.UUID

	switch targetType {
	case domain.TargetBroadcast:
		// Tidak perlu resolve — BroadcastNotification akan menangani sendiri
		return nil, nil

	case domain.TargetRole:
		seen := make(map[uuid.UUID]bool)
		for _, role := range req.TargetRoles {
			ids, err := u.announcementRepo.GetUserIDsByRole(role)
			if err != nil {
				return nil, fmt.Errorf("gagal fetch user IDs untuk role %s: %w", role, err)
			}
			for _, id := range ids {
				if !seen[id] {
					seen[id] = true
					userIDs = append(userIDs, id)
				}
			}
		}

	case domain.TargetIndividual:
		for _, idStr := range req.TargetUserIDs {
			uid, err := uuid.Parse(idStr)
			if err != nil {
				continue
			}
			userIDs = append(userIDs, uid)
		}
	}

	return userIDs, nil
}

func (u *announcementUseCase) UpdateAnnouncement(id uuid.UUID, req UpdateAnnouncementRequest) error {
	announcement, err := u.announcementRepo.FindByID(id)
	if err != nil {
		return err
	}

	oldImageURL := announcement.ImageURL
	oldMediaURLs := append([]string{}, announcement.MediaURLs...)

	targetType := domain.TargetType(req.TargetType)
	if targetType == "" {
		targetType = domain.TargetBroadcast
	}

	primaryImage := req.ImageURL
	mediaURLs := req.MediaURLs
	if primaryImage == "" && len(mediaURLs) > 0 {
		primaryImage = mediaURLs[0]
	}
	if len(mediaURLs) == 0 && primaryImage != "" {
		mediaURLs = []string{primaryImage}
	}

	cleanURL := strings.ToLower(strings.TrimSpace(primaryImage))
	isVideoFile := strings.HasSuffix(cleanURL, ".mp4") || strings.HasSuffix(cleanURL, ".mov") ||
		strings.HasSuffix(cleanURL, ".webm") || strings.HasSuffix(cleanURL, ".m4v") ||
		strings.Contains(cleanURL, "drive.google.com") || strings.Contains(cleanURL, "/file/d/")

	mediaType := req.MediaType
	if mediaType == "" || (mediaType == "image" && isVideoFile) {
		if isVideoFile {
			mediaType = "video"
		} else if len(req.MediaURLs) > 1 {
			mediaType = "multi_image"
		} else if len(req.MediaURLs) == 1 || req.ImageURL != "" {
			mediaType = "image"
		} else if req.ExternalURL != "" {
			mediaType = "social_media"
		} else {
			mediaType = "none"
		}
	}

	announcement.Title = req.Title
	announcement.Content = req.Content
	announcement.Category = req.Category
	announcement.IsActive = req.IsActive
	announcement.TargetType = targetType
	announcement.IsPopup = req.IsPopup
	announcement.PopupExpiresAt = req.PopupExpiresAt
	announcement.MediaType = mediaType
	announcement.ImageURL = primaryImage
	announcement.MediaURLs = mediaURLs
	announcement.ExternalURL = req.ExternalURL
	announcement.TrackDownloads = req.TrackDownloads
	announcement.TrackShares = req.TrackShares

	var targets []domain.AnnouncementTarget
	switch targetType {
	case domain.TargetRole:
		for _, role := range req.TargetRoles {
			if role != "" {
				targets = append(targets, domain.AnnouncementTarget{
					AnnouncementID: announcement.ID,
					TargetRole:     role,
				})
			}
		}
	case domain.TargetIndividual:
		for _, idStr := range req.TargetUserIDs {
			uid, err := uuid.Parse(idStr)
			if err != nil {
				continue
			}
			targets = append(targets, domain.AnnouncementTarget{
				AnnouncementID: announcement.ID,
				TargetUserID:   &uid,
			})
		}
	}

	if err := u.announcementRepo.UpdateWithTargets(announcement, targets); err != nil {
		return err
	}

	// Jika pop-up diaktifkan saat update, kirimkan push notification FCM ke target
	if req.IsPopup {
		targetUserIDs, _ := u.resolveTargetUserIDs(announcement.ID, targetType, CreateAnnouncementRequest{
			TargetType:    req.TargetType,
			TargetRoles:   req.TargetRoles,
			TargetUserIDs: req.TargetUserIDs,
		})

		body := req.Content
		if len(body) > 100 {
			body = body[:97] + "..."
		}
		fcmTitle := "⚠️ " + req.Title
		fcmData := map[string]string{
			"type":            "popup",
			"announcement_id": announcement.ID.String(),
			"category":        req.Category,
		}

		if targetType == domain.TargetBroadcast || len(targetUserIDs) == 0 {
			_ = u.fcmUseCase.BroadcastNotification(fcmTitle, body)
		} else {
			_ = u.fcmUseCase.SendToUsers(targetUserIDs, fcmTitle, body, fcmData)
		}

		u.publishRealtimeNotification(announcement, fcmTitle, body, req.TargetRoles, targetUserIDs)
	}

	// Tangani pembersihan file media lama jika diganti saat update
	var oldMediaFiles []string
	if oldImageURL != "" && oldImageURL != primaryImage {
		oldMediaFiles = append(oldMediaFiles, oldImageURL)
	}
	for _, oldURL := range oldMediaURLs {
		isStillUsed := false
		for _, newURL := range mediaURLs {
			if oldURL == newURL {
				isStillUsed = true
				break
			}
		}
		if !isStillUsed && oldURL != primaryImage {
			oldMediaFiles = append(oldMediaFiles, oldURL)
		}
	}
	if len(oldMediaFiles) > 0 {
		u.deleteLocalMediaFiles(oldMediaFiles)
	}

	return nil
}

// deleteLocalMediaFiles menghapus berkas fisik yang disimpan di disk server lokal (uploads/announcements/...)
func (u *announcementUseCase) deleteLocalMediaFiles(urls []string) {
	for _, uStr := range urls {
		clean := strings.TrimSpace(uStr)
		if clean == "" {
			continue
		}
		// Pastikan hanya menghapus berkas yang memang berada di folder uploads/announcements/
		if strings.HasPrefix(clean, "/uploads/announcements/") || strings.HasPrefix(clean, "uploads/announcements/") {
			localPath := strings.TrimPrefix(clean, "/")
			if err := os.Remove(localPath); err != nil && !os.IsNotExist(err) {
				log.Printf("[CLEANUP] Gagal menghapus file media announcement: %s, err: %v\n", localPath, err)
			} else if err == nil {
				log.Printf("[CLEANUP] Berhasil menghapus file media announcement: %s\n", localPath)
			}
		}
	}
}

func (u *announcementUseCase) DeleteAnnouncement(id uuid.UUID) error {
	// Ambil data pengumuman terlebih dahulu untuk mendeteksi file media lokal yang tersimpan
	ann, err := u.announcementRepo.FindByID(id)
	if err == nil && ann != nil {
		var filesToDelete []string
		if ann.ImageURL != "" {
			filesToDelete = append(filesToDelete, ann.ImageURL)
		}
		for _, mURL := range ann.MediaURLs {
			if mURL != "" && mURL != ann.ImageURL {
				filesToDelete = append(filesToDelete, mURL)
			}
		}

		// Hapus record dari database terlebih dahulu
		if err := u.announcementRepo.Delete(id); err != nil {
			return err
		}

		// Jika database berhasil dihapus, bersihkan berkas fisik di disk server
		if len(filesToDelete) > 0 {
			u.deleteLocalMediaFiles(filesToDelete)
		}
		return nil
	}

	return u.announcementRepo.Delete(id)
}

// AcknowledgeAnnouncement dipanggil ketika user menekan OK pada popup di mobile atau web.
func (u *announcementUseCase) AcknowledgeAnnouncement(announcementID, userID uuid.UUID) error {
	return u.announcementRepo.AcknowledgeByUser(announcementID, userID)
}

// GetPendingPopups mengembalikan daftar popup yang belum di-acknowledge oleh user.
// Dipanggil oleh mobile & web manajemen saat app dibuka (polling) untuk memastikan tidak ada popup yang terlewat.
func (u *announcementUseCase) GetPendingPopups(userID uuid.UUID) ([]domain.Announcement, error) {
	return u.announcementRepo.FindPendingPopupsForUser(userID)
}

func (u *announcementUseCase) GetAnnouncementConfirmations(announcementID uuid.UUID) (*domain.AnnouncementConfirmationSummary, error) {
	return u.announcementRepo.FindConfirmationsByAnnouncementID(announcementID)
}

// ResendAnnouncementFCM mengirim ulang push notification FCM ke target tertentu.
// Mode:
// - "all": kirim ke semua target pengumuman
// - "pending_only": hanya kirim ke target yang belum konfirmasi (default)
// - "single": hanya kirim ke 1 user_id tertentu
func (u *announcementUseCase) ResendAnnouncementFCM(announcementID uuid.UUID, mode string, targetUserID *uuid.UUID) (int, error) {
	ann, err := u.announcementRepo.FindByID(announcementID)
	if err != nil {
		return 0, fmt.Errorf("pengumuman tidak ditemukan: %w", err)
	}

	var targetUserIDs []uuid.UUID

	switch mode {
	case "single":
		if targetUserID == nil {
			return 0, fmt.Errorf("target_user_id wajib diisi untuk mode single")
		}
		targetUserIDs = []uuid.UUID{*targetUserID}

	case "all":
		if ann.TargetType == domain.TargetBroadcast || ann.TargetType == "" {
			// Broadcast mode
			body := ann.Content
			if len(body) > 100 {
				body = body[:97] + "..."
			}
			fcmTitle := "📢 " + ann.Title
			if ann.IsPopup || ann.Category == "urgent_call" {
				fcmTitle = "⚠️ " + ann.Title
			}
			_ = u.fcmUseCase.BroadcastNotification(fcmTitle, body)
			u.publishRealtimeNotification(ann, fcmTitle, body, nil, nil)
			return 0, nil
		}
		targetUserIDs, err = u.announcementRepo.FindAllTargetUserIDsByAnnouncementID(ann.ID)
		if err != nil {
			return 0, err
		}

	default: // "pending_only"
		targetUserIDs, err = u.announcementRepo.FindPendingUserIDsByAnnouncementID(ann.ID)
		if err != nil {
			return 0, err
		}
	}

	if len(targetUserIDs) == 0 {
		return 0, nil
	}

	body := ann.Content
	if len(body) > 100 {
		body = body[:97] + "..."
	}
	fcmTitle := "📢 Pengingat: " + ann.Title
	if ann.IsPopup || ann.Category == "urgent_call" {
		fcmTitle = "⚠️ Panggilan Urgent: " + ann.Title
	}

	fcmData := map[string]string{
		"type":            "popup",
		"announcement_id": ann.ID.String(),
		"category":        ann.Category,
	}

	if err := u.fcmUseCase.SendToUsers(targetUserIDs, fcmTitle, body, fcmData); err != nil {
		return 0, err
	}

	u.publishRealtimeNotification(ann, fcmTitle, body, nil, targetUserIDs)

	return len(targetUserIDs), nil
}

// publishRealtimeNotification mempublikasikan event pengumuman ke Redis Pub/Sub untuk realtime update Web Manajemen
func (u *announcementUseCase) publishRealtimeNotification(
	ann *domain.Announcement,
	title string,
	message string,
	targetRoles []string,
	targetUserIDs []uuid.UUID,
) {
	if u.sseHub == nil || ann == nil {
		return
	}

	var targetUserIDStrings []string
	for _, id := range targetUserIDs {
		targetUserIDStrings = append(targetUserIDStrings, id.String())
	}

	event := sse.NotificationEvent{
		ID:        ann.ID.String(),
		Type:      "announcement",
		Title:     title,
		Message:   message,
		CreatedAt: time.Now().Format("2006-01-02 15:04:05"),
		Data: map[string]interface{}{
			"announcement_id": ann.ID.String(),
			"title":           ann.Title,
			"content":         ann.Content,
			"category":        ann.Category,
			"is_popup":        ann.IsPopup,
			"target_type":     string(ann.TargetType),
			"target_roles":    targetRoles,
			"target_user_ids": targetUserIDStrings,
			"created_at":      ann.CreatedAt.Format(time.RFC3339),
		},
	}

	_ = u.sseHub.PublishNotification(context.Background(), event)
}

func (u *announcementUseCase) RecordActivity(announcementID, userID uuid.UUID, actionType, channel string) error {
	return u.announcementRepo.RecordActivity(announcementID, userID, actionType, channel)
}
