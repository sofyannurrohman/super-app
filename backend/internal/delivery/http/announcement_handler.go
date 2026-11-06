package http

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"example.com/hris/backend/internal/usecase"
	"example.com/hris/backend/pkg/utils"
)

type AnnouncementHandler struct {
	usecase usecase.AnnouncementUseCase
	gdrive  *utils.GDriveService
}

func NewAnnouncementHandler(usecase usecase.AnnouncementUseCase, gdrive *utils.GDriveService) *AnnouncementHandler {
	return &AnnouncementHandler{
		usecase: usecase,
		gdrive:  gdrive,
	}
}

func (h *AnnouncementHandler) SetupRoutes(router *gin.RouterGroup) {
	router.GET("/announcements", h.GetAnnouncements)
	router.GET("/announcements/:id", h.GetAnnouncementByID)
	router.POST("/announcements/:id/like", h.ToggleLike)
	router.GET("/announcements/:id/comments", h.GetComments)
	router.POST("/announcements/:id/comments", h.AddComment)
	router.DELETE("/announcements/:id/comments/:comment_id", h.DeleteComment)
	router.GET("/admin/announcements", h.AdminListAnnouncements)
	router.POST("/admin/announcements/upload", h.UploadMedia)
	router.POST("/admin/announcements/upload-chunk", h.UploadMediaChunk)
	router.GET("/admin/announcements/:id/confirmations", h.GetAnnouncementConfirmations)
	router.POST("/admin/announcements/:id/resend-fcm", h.ResendAnnouncementFCM)
	router.POST("/announcements", h.CreateAnnouncement)
	router.PUT("/announcements/:id", h.UpdateAnnouncement)
	router.DELETE("/announcements/:id", h.DeleteAnnouncement)

	// Popup endpoints untuk mobile
	// GET  /announcements/pending-popup  → cek popup yang belum di-acknowledge
	// POST /announcements/:id/acknowledge → user klik OK pada popup
	router.GET("/announcements/pending-popup", h.GetPendingPopups)
	router.POST("/announcements/:id/acknowledge", h.AcknowledgeAnnouncement)
	router.POST("/announcements/:id/activity", h.RecordActivity)
}

func (h *AnnouncementHandler) GetAnnouncements(c *gin.Context) {
	var targetUID *uuid.UUID
	if userIDVal, exists := c.Get("user_id"); exists {
		if uid, err := uuid.Parse(userIDVal.(string)); err == nil {
			targetUID = &uid
		}
	}

	announcements, err := h.usecase.GetAnnouncements(targetUID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Announcements fetched successfully", announcements)
}

func (h *AnnouncementHandler) GetAnnouncementByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Format ID pengumuman tidak valid")
		return
	}

	var targetUID *uuid.UUID
	if userIDVal, exists := c.Get("user_id"); exists {
		if uid, err := uuid.Parse(userIDVal.(string)); err == nil {
			targetUID = &uid
		}
	}

	announcement, err := h.usecase.GetAnnouncementByID(id, targetUID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Pengumuman tidak ditemukan")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Announcement fetched successfully", announcement)
}

func (h *AnnouncementHandler) ToggleLike(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Format ID pengumuman tidak valid")
		return
	}

	var userID uuid.UUID
	if uidVal, exists := c.Get("userID"); exists {
		if uid, ok := uidVal.(uuid.UUID); ok {
			userID = uid
		}
	}
	if userID == uuid.Nil {
		if uidVal, exists := c.Get("user_id"); exists {
			if uidStr, ok := uidVal.(string); ok {
				userID, _ = uuid.Parse(uidStr)
			}
		}
	}

	if userID == uuid.Nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User tidak terautentikasi")
		return
	}

	isLiked, totalLikes, err := h.usecase.ToggleLike(id, userID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	actionMsg := "Disukai"
	if !isLiked {
		actionMsg = "Batal disukai"
	}

	utils.SuccessResponse(c, http.StatusOK, actionMsg, gin.H{
		"announcement_id": id,
		"is_liked":        isLiked,
		"likes_count":     totalLikes,
	})
}

type AddCommentRequest struct {
	Content string `json:"content" binding:"required"`
}

func (h *AnnouncementHandler) GetComments(c *gin.Context) {
	idStr := c.Param("id")
	announcementID, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Format ID pengumuman tidak valid")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	var currentUID *uuid.UUID
	if uidVal, exists := c.Get("userID"); exists {
		if uid, ok := uidVal.(uuid.UUID); ok {
			currentUID = &uid
		}
	}
	if currentUID == nil {
		if uidVal, exists := c.Get("user_id"); exists {
			if uidStr, ok := uidVal.(string); ok {
				if uid, err := uuid.Parse(uidStr); err == nil {
					currentUID = &uid
				}
			}
		}
	}

	comments, total, err := h.usecase.GetComments(announcementID, currentUID, page, limit)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Comments fetched successfully", gin.H{
		"comments": comments,
		"total":    total,
		"page":     page,
		"limit":    limit,
	})
}

func (h *AnnouncementHandler) AddComment(c *gin.Context) {
	idStr := c.Param("id")
	announcementID, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Format ID pengumuman tidak valid")
		return
	}

	var userID uuid.UUID
	if uidVal, exists := c.Get("userID"); exists {
		if uid, ok := uidVal.(uuid.UUID); ok {
			userID = uid
		}
	}
	if userID == uuid.Nil {
		if uidVal, exists := c.Get("user_id"); exists {
			if uidStr, ok := uidVal.(string); ok {
				userID, _ = uuid.Parse(uidStr)
			}
		}
	}

	if userID == uuid.Nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User tidak terautentikasi")
		return
	}

	var req AddCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Isi komentar wajib diisi")
		return
	}

	comment, err := h.usecase.AddComment(announcementID, userID, req.Content)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Komentar berhasil dikirim", comment)
}

func (h *AnnouncementHandler) DeleteComment(c *gin.Context) {
	idStr := c.Param("id")
	announcementID, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Format ID pengumuman tidak valid")
		return
	}

	commentIDStr := c.Param("comment_id")
	commentID, err := uuid.Parse(commentIDStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Format ID komentar tidak valid")
		return
	}

	var userID uuid.UUID
	if uidVal, exists := c.Get("userID"); exists {
		if uid, ok := uidVal.(uuid.UUID); ok {
			userID = uid
		}
	}
	if userID == uuid.Nil {
		if uidVal, exists := c.Get("user_id"); exists {
			if uidStr, ok := uidVal.(string); ok {
				userID, _ = uuid.Parse(uidStr)
			}
		}
	}

	if userID == uuid.Nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User tidak terautentikasi")
		return
	}

	userRole := ""
	if rVal, exists := c.Get("role"); exists {
		if rStr, ok := rVal.(string); ok {
			userRole = rStr
		}
	}
	isAdmin := userRole == "superadmin" || userRole == "hr_admin" || userRole == "manager"

	if err := h.usecase.DeleteComment(announcementID, commentID, userID, isAdmin); err != nil {
		utils.ErrorResponse(c, http.StatusForbidden, "Tidak memiliki izin menghapus komentar ini")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Komentar berhasil dihapus", nil)
}

func (h *AnnouncementHandler) AdminListAnnouncements(c *gin.Context) {
	announcements, err := h.usecase.AdminListAnnouncements()
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Admin announcements fetched successfully", announcements)
}

// CreateAnnouncement membuat pengumuman baru dengan konfigurasi targeting dan popup opsional.
//
// Contoh body JSON:
//
// Broadcast biasa:
//
//	{ "title": "Info", "content": "...", "category": "policy" }
//
// Kirim ke role tertentu:
//
//	{ "title": "Info Sales", "content": "...", "category": "event",
//	  "target_type": "role", "target_roles": ["supervisor_sales", "manager"] }
//
// Kirim ke individu tertentu:
//
//	{ "title": "Panggilan", "content": "Harap menghadap HRD", "category": "urgent_call",
//	  "target_type": "individual", "target_user_ids": ["uuid1", "uuid2"],
//	  "is_popup": true }
func (h *AnnouncementHandler) CreateAnnouncement(c *gin.Context) {
	var req usecase.CreateAnnouncementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.usecase.CreateAnnouncement(req); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Announcement created successfully", nil)
}

func (h *AnnouncementHandler) UpdateAnnouncement(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid ID format")
		return
	}

	var req usecase.UpdateAnnouncementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.usecase.UpdateAnnouncement(id, req); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Announcement updated successfully", nil)
}

func (h *AnnouncementHandler) DeleteAnnouncement(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid ID format")
		return
	}

	if err := h.usecase.DeleteAnnouncement(id); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Announcement deleted successfully", nil)
}

// GetPendingPopups mengembalikan daftar popup yang belum di-acknowledge oleh user yang sedang login.
// Mobile memanggil endpoint ini saat app dibuka untuk menampilkan popup yang mungkin terlewat.
//
// Membutuhkan JWT claim "user_id".
//
// Response: array Announcement dengan is_popup=true yang belum di-acknowledge.
func (h *AnnouncementHandler) GetPendingPopups(c *gin.Context) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User tidak terautentikasi")
		return
	}

	userID, err := uuid.Parse(userIDVal.(string))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "User ID tidak valid")
		return
	}

	popups, err := h.usecase.GetPendingPopups(userID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Pending popups fetched successfully", popups)
}

// AcknowledgeAnnouncement dipanggil ketika user menekan tombol OK pada popup di mobile.
// Mencatat bahwa user sudah membaca pengumuman tersebut sehingga tidak akan muncul lagi.
//
// Membutuhkan JWT claim "user_id".
func (h *AnnouncementHandler) AcknowledgeAnnouncement(c *gin.Context) {
	idStr := c.Param("id")
	announcementID, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid announcement ID format")
		return
	}

	userIDVal, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User tidak terautentikasi")
		return
	}

	userID, err := uuid.Parse(userIDVal.(string))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "User ID tidak valid")
		return
	}

	if err := h.usecase.AcknowledgeAnnouncement(announcementID, userID); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Announcement acknowledged", nil)
}

// GetAnnouncementConfirmations mengambil rekap data konfirmasi penerima target dari pengumuman.
func (h *AnnouncementHandler) GetAnnouncementConfirmations(c *gin.Context) {
	idStr := c.Param("id")
	announcementID, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid announcement ID format")
		return
	}

	summary, err := h.usecase.GetAnnouncementConfirmations(announcementID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Announcement confirmations fetched successfully", summary)
}

type ResendAnnouncementFCMRequest struct {
	Mode         string  `json:"mode"`           // "all", "pending_only", "single"
	TargetUserID *string `json:"target_user_id"` // jika mode="single"
}

// ResendAnnouncementFCM memicu pengiriman ulang notifikasi FCM untuk sebuah pengumuman.
func (h *AnnouncementHandler) ResendAnnouncementFCM(c *gin.Context) {
	idStr := c.Param("id")
	announcementID, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid announcement ID format")
		return
	}

	var req ResendAnnouncementFCMRequest
	_ = c.ShouldBindJSON(&req)

	mode := req.Mode
	if mode == "" {
		mode = "pending_only"
	}

	var targetUID *uuid.UUID
	if req.TargetUserID != nil && *req.TargetUserID != "" {
		uid, err := uuid.Parse(*req.TargetUserID)
		if err == nil {
			targetUID = &uid
		}
	}

	count, err := h.usecase.ResendAnnouncementFCM(announcementID, mode, targetUID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Notifikasi FCM berhasil dikirim ulang", gin.H{
		"recipients_count": count,
		"mode":             mode,
	})
}

// uploadSemaphore membatasi maksimal 2 proses upload media bersamaan agar I/O & resource server
// tetap terjaga stabil dan tidak mengganggu proses transaksi operasional lainnya.
var uploadSemaphore = make(chan struct{}, 2)

// UploadMedia mengunggah file gambar atau video pengumuman ke server lokal.
// File disimpan di ./uploads/announcements/images/ atau ./uploads/announcements/videos/.
func (h *AnnouncementHandler) UploadMedia(c *gin.Context) {
	select {
	case uploadSemaphore <- struct{}{}:
		defer func() { <-uploadSemaphore }()
	default:
		utils.ErrorResponse(c, http.StatusTooManyRequests, "Server sedang memproses unggahan media lain. Silakan coba sesaat lagi agar performa sistem tetap stabil.")
		return
	}

	file, err := c.FormFile("media")
	if err != nil {
		file, err = c.FormFile("file")
	}
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "File media wajib diunggah (field 'media' atau 'file')")
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	contentType := file.Header.Get("Content-Type")

	isImage := ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".webp" ||
		strings.HasPrefix(contentType, "image/")
	isVideo := ext == ".mp4" || ext == ".mov" || ext == ".webm" || ext == ".m4v" ||
		strings.HasPrefix(contentType, "video/")

	if !isImage && !isVideo {
		utils.ErrorResponse(c, http.StatusBadRequest, "Format file tidak didukung. Harap unggah gambar (JPG, PNG, WebP) atau video (MP4, MOV).")
		return
	}

	// Batasan ukuran untuk single upload: Gambar max 15MB, Video max 95MB (batas aman body Cloudflare)
	if isImage && file.Size > 15*1024*1024 {
		utils.ErrorResponse(c, http.StatusBadRequest, "Ukuran file gambar melebihi batas maksimum 15MB")
		return
	}
	if isVideo && file.Size > 95*1024*1024 {
		utils.ErrorResponse(c, http.StatusBadRequest, "Ukuran file video melebihi batas 95MB untuk upload langsung. Silakan gunakan fitur unggah video bertahap (chunked upload) hingga 500MB atau masukkan link Google Drive.")
		return
	}

	var fileURL string
	var mediaType string
	isGDrive := false

	if isImage {
		mediaType = "image"
		targetDir := "./uploads/announcements/images"
		if err := os.MkdirAll(targetDir, 0755); err != nil {
			utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal menyiapkan direktori penyimpanan gambar")
			return
		}

		// Kompresi gambar otomatis ke WebP/JPEG terkompresi
		fileURL, err = utils.SaveUploadedImageCompressed(file, targetDir, 1600, 80)
		if err != nil {
			utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal mengompresi dan menyimpan gambar: "+err.Error())
			return
		}
	} else {
		mediaType = "video"
		cleanExt := ext
		if cleanExt == "" {
			cleanExt = ".mp4"
		}

		// Jika file video berukuran > 40MB dan GDrive service aktif, simpan langsung ke Google Drive
		// untuk menghemat disk VPS dan membebaskan beban bandwidth server saat video diputar.
		if h.gdrive != nil && file.Size > 40*1024*1024 {
			tempDir := "./uploads/announcements/temp"
			_ = os.MkdirAll(tempDir, 0755)
			tempPath := filepath.Join(tempDir, fmt.Sprintf("temp_vid_%d%s", time.Now().UnixNano(), cleanExt))
			if err := utils.SaveUploadedFile(file, tempPath); err == nil {
				customName := fmt.Sprintf("pengumuman_%d%s", time.Now().Unix(), cleanExt)
				gdriveURL, gErr := h.gdrive.UploadLocalFile(tempPath, customName)
				_ = os.Remove(tempPath)
				if gErr == nil && gdriveURL != "" {
					fileURL = gdriveURL
					isGDrive = true
				} else {
					log.Printf("[AnnouncementUpload] Warning: GDrive upload failed (%v), falling back to local storage", gErr)
				}
			}
		}

		if fileURL == "" {
			targetDir := "./uploads/announcements/videos"
			if err := os.MkdirAll(targetDir, 0755); err != nil {
				utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal menyiapkan direktori penyimpanan video")
				return
			}

			filename := fmt.Sprintf("announcement_vid_%d%s", time.Now().UnixNano(), cleanExt)
			savePath := filepath.Join(targetDir, filename)

			if err := utils.SaveUploadedFile(file, savePath); err != nil {
				utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal menyimpan file video: "+err.Error())
				return
			}
			fileURL = "/uploads/announcements/videos/" + filename
		}
	}

	utils.SuccessResponse(c, http.StatusOK, "File media berhasil diunggah", gin.H{
		"url":        fileURL,
		"file_name":  file.Filename,
		"file_size":  file.Size,
		"media_type": mediaType,
		"is_gdrive":  isGDrive,
	})
}

// UploadMediaChunk menangani pengunggahan berkas besar secara bertahap (chunked upload, e.g. 20MB per chunk)
// agar dapat menembus limit 100MB Cloudflare dan mengunggah video hingga 500MB ke Google Drive atau server lokal.
func (h *AnnouncementHandler) UploadMediaChunk(c *gin.Context) {
	select {
	case uploadSemaphore <- struct{}{}:
		defer func() { <-uploadSemaphore }()
	default:
		utils.ErrorResponse(c, http.StatusTooManyRequests, "Server sedang memproses unggahan media lain. Silakan coba sesaat lagi.")
		return
	}

	uploadID := strings.TrimSpace(c.PostForm("upload_id"))
	chunkIndexStr := strings.TrimSpace(c.PostForm("chunk_index"))
	totalChunksStr := strings.TrimSpace(c.PostForm("total_chunks"))
	fileName := strings.TrimSpace(c.PostForm("file_name"))

	if uploadID == "" || chunkIndexStr == "" || totalChunksStr == "" || fileName == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Parameter upload_id, chunk_index, total_chunks, dan file_name wajib disertakan")
		return
	}

	// Sanitasi uploadID agar aman dari path traversal
	for _, ch := range uploadID {
		if !((ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '-' || ch == '_') {
			utils.ErrorResponse(c, http.StatusBadRequest, "Format upload_id tidak valid")
			return
		}
	}

	chunkIndex, err := strconv.Atoi(chunkIndexStr)
	if err != nil || chunkIndex < 0 {
		utils.ErrorResponse(c, http.StatusBadRequest, "chunk_index tidak valid")
		return
	}

	totalChunks, err := strconv.Atoi(totalChunksStr)
	if err != nil || totalChunks <= 0 || totalChunks > 100 { // Max 100 chunks * 20MB = 2GB
		utils.ErrorResponse(c, http.StatusBadRequest, "total_chunks tidak valid")
		return
	}

	file, err := c.FormFile("chunk")
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "File chunk wajib disertakan")
		return
	}

	// Direktori sementara chunk: ./uploads/announcements/chunks/<upload_id>/
	chunkDir := filepath.Join("./uploads/announcements/chunks", uploadID)
	if err := os.MkdirAll(chunkDir, 0755); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal membuat direktori chunk sementara")
		return
	}

	chunkPath := filepath.Join(chunkDir, fmt.Sprintf("chunk_%05d", chunkIndex))
	if err := utils.SaveUploadedFile(file, chunkPath); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal menyimpan potongan chunk: "+err.Error())
		return
	}

	// Jika belum chunk terakhir, kembalikan konfirmasi penerimaan
	if chunkIndex < totalChunks-1 {
		utils.SuccessResponse(c, http.StatusOK, fmt.Sprintf("Chunk %d/%d berhasil diterima", chunkIndex+1, totalChunks), gin.H{
			"status":       "chunk_received",
			"chunk_index":  chunkIndex,
			"total_chunks": totalChunks,
		})
		return
	}

	// Jika ini chunk terakhir, pastikan semua file chunk 0 s/d totalChunks-1 ada
	for i := 0; i < totalChunks; i++ {
		p := filepath.Join(chunkDir, fmt.Sprintf("chunk_%05d", i))
		if _, err := os.Stat(p); os.IsNotExist(err) {
			utils.ErrorResponse(c, http.StatusBadRequest, fmt.Sprintf("Potongan chunk %d belum lengkap di server", i))
			return
		}
	}

	// Gabungkan seluruh potongan chunk menjadi satu file utuh
	cleanExt := strings.ToLower(filepath.Ext(fileName))
	if cleanExt == "" {
		cleanExt = ".mp4"
	}
	assembledFilename := fmt.Sprintf("assembled_%d%s", time.Now().UnixNano(), cleanExt)
	assembledPath := filepath.Join(chunkDir, assembledFilename)

	assembledFile, err := os.Create(assembledPath)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal menyiapkan penggabungan file: "+err.Error())
		return
	}

	for i := 0; i < totalChunks; i++ {
		p := filepath.Join(chunkDir, fmt.Sprintf("chunk_%05d", i))
		chunkData, err := os.ReadFile(p)
		if err != nil {
			assembledFile.Close()
			_ = os.Remove(assembledPath)
			utils.ErrorResponse(c, http.StatusInternalServerError, fmt.Sprintf("Gagal membaca chunk %d: %v", i, err))
			return
		}
		if _, err := assembledFile.Write(chunkData); err != nil {
			assembledFile.Close()
			_ = os.Remove(assembledPath)
			utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal menulis chunk ke file akhir: "+err.Error())
			return
		}
	}
	assembledFile.Close()

	// Periksa ukuran total file (maksimum 500MB)
	assembledStat, err := os.Stat(assembledPath)
	if err != nil {
		_ = os.RemoveAll(chunkDir)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal memeriksa file hasil penggabungan")
		return
	}

	if assembledStat.Size() > 500*1024*1024 {
		_ = os.RemoveAll(chunkDir)
		utils.ErrorResponse(c, http.StatusBadRequest, "Total ukuran file melebihi batas maksimum 500MB")
		return
	}

	isVideo := cleanExt == ".mp4" || cleanExt == ".mov" || cleanExt == ".webm" || cleanExt == ".m4v"
	var fileURL string
	var mediaType string
	isGDrive := false

	if isVideo {
		mediaType = "video"
		// Jika Google Drive tersedia, unggah file besar langsung ke Google Drive
		// agar VPS tidak terbebani penyimpanan dan streaming video ditangani Google CDN.
		if h.gdrive != nil {
			customName := fmt.Sprintf("pengumuman_%d%s", time.Now().Unix(), cleanExt)
			gdriveURL, err := h.gdrive.UploadLocalFile(assembledPath, customName)
			if err == nil && gdriveURL != "" {
				fileURL = gdriveURL
				isGDrive = true
			} else {
				log.Printf("[AnnouncementChunked] Warning: GDrive upload failed (%v), falling back to local storage", err)
			}
		}

		// Fallback ke penyimpanan lokal jika GDrive belum aktif atau gagal
		if fileURL == "" {
			targetDir := "./uploads/announcements/videos"
			if err := os.MkdirAll(targetDir, 0755); err != nil {
				_ = os.RemoveAll(chunkDir)
				utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal menyiapkan direktori video")
				return
			}
			finalName := fmt.Sprintf("announcement_vid_%d%s", time.Now().UnixNano(), cleanExt)
			finalPath := filepath.Join(targetDir, finalName)

			if err := copyLocalFile(assembledPath, finalPath); err != nil {
				_ = os.RemoveAll(chunkDir)
				utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal memindahkan file video: "+err.Error())
				return
			}
			fileURL = "/uploads/announcements/videos/" + finalName
		}
	} else {
		mediaType = "image"
		targetDir := "./uploads/announcements/images"
		if err := os.MkdirAll(targetDir, 0755); err != nil {
			_ = os.RemoveAll(chunkDir)
			utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal menyiapkan direktori gambar")
			return
		}
		finalName := fmt.Sprintf("announcement_img_%d%s", time.Now().UnixNano(), cleanExt)
		finalPath := filepath.Join(targetDir, finalName)
		if err := copyLocalFile(assembledPath, finalPath); err != nil {
			_ = os.RemoveAll(chunkDir)
			utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal menyimpan gambar: "+err.Error())
			return
		}
		fileURL = "/uploads/announcements/images/" + finalName
	}

	// Bersihkan direktori temporary chunk setelah penggabungan & upload selesai
	_ = os.RemoveAll(chunkDir)

	utils.SuccessResponse(c, http.StatusOK, "File media berhasil digabungkan dan disimpan", gin.H{
		"url":        fileURL,
		"file_name":  fileName,
		"file_size":  assembledStat.Size(),
		"media_type": mediaType,
		"is_gdrive":  isGDrive,
	})
}

// copyLocalFile menyalin berkas fisik dari src ke dst
func copyLocalFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

type RecordAnnouncementActivityRequest struct {
	ActionType string `json:"action_type" binding:"required"` // 'download', 'share'
	Channel    string `json:"channel"`                         // 'whatsapp', 'general', 'media_file'
}

func (h *AnnouncementHandler) RecordActivity(c *gin.Context) {
	idStr := c.Param("id")
	announcementID, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Format ID pengumuman tidak valid")
		return
	}

	var userID uuid.UUID
	if uidVal, exists := c.Get("userID"); exists {
		if uid, ok := uidVal.(uuid.UUID); ok {
			userID = uid
		}
	}
	if userID == uuid.Nil {
		if uidVal, exists := c.Get("user_id"); exists {
			if uidStr, ok := uidVal.(string); ok {
				userID, _ = uuid.Parse(uidStr)
			}
		}
	}
	if userID == uuid.Nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User tidak terautentikasi")
		return
	}

	var req RecordAnnouncementActivityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "ActionType wajib diisi ('download' atau 'share')")
		return
	}

	if err := h.usecase.RecordActivity(announcementID, userID, req.ActionType, req.Channel); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Aktivitas berhasil dicatat", nil)
}


