package http

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"gorm.io/gorm"
)

const (
	aiQueuedHermes     = "queued_hermes"
	aiHermesProcessing = "hermes_processing"
	aiAwaitingCorel    = "awaiting_corel"
	aiCorelProcessing  = "corel_processing"
	aiCompleted        = "completed"
	aiFailed           = "failed"
)

func ensureBrandingAIDesignJobsTable(db *gorm.DB) {
	if db == nil {
		return
	}
	db.Exec(`
		CREATE TABLE IF NOT EXISTS branding_ai_design_jobs (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			banner_order_id UUID NOT NULL REFERENCES banner_orders(id) ON DELETE CASCADE,
			requested_by UUID REFERENCES employees(id) ON DELETE SET NULL,
			status VARCHAR(40) NOT NULL DEFAULT 'queued_hermes',
			reference_banner_id UUID REFERENCES banner_orders(id) ON DELETE SET NULL,
			reference_image_url TEXT,
			prompt TEXT NOT NULL,
			svg_url TEXT,
			preview_url TEXT,
			cdr_url TEXT,
			cdr_file_name VARCHAR(500),
			error_message TEXT,
			claimed_at TIMESTAMPTZ,
			completed_at TIMESTAMPTZ,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
	`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_branding_ai_jobs_banner_created ON branding_ai_design_jobs (banner_order_id, created_at DESC)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_branding_ai_jobs_status_created ON branding_ai_design_jobs (status, created_at ASC)`)
	db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_branding_ai_jobs_one_active_per_banner
		ON branding_ai_design_jobs (banner_order_id)
		WHERE status IN ('queued_hermes', 'hermes_processing', 'awaiting_corel', 'corel_processing')
	`)
}

func (h *BrandingProxyHandler) findBannerOrder(rawID string) (*domain.BannerOrder, error) {
	var banner domain.BannerOrder
	query := h.localDB.Where("id = ?", rawID)
	if numericID, err := strconv.ParseInt(rawID, 10, 64); err == nil {
		query = h.localDB.Where("id::text = ? OR mysql_banner_id = ?", rawID, numericID)
	}
	if err := query.First(&banner).Error; err != nil {
		return nil, err
	}
	return &banner, nil
}

func canManageAIDraft(c *gin.Context, banner *domain.BannerOrder, employeeID *uuid.UUID) bool {
	role, _ := c.Get("role")
	roleName := strings.ToLower(fmt.Sprint(role))
	if roleName == "superadmin" || roleName == "admin" || roleName == "hr_admin" {
		return true
	}
	return banner.EditorID != nil && employeeID != nil && *banner.EditorID == *employeeID
}

// CreateAIDesignDraft creates a durable job consumed first by Hermes on the Mac mini,
// then by the CorelDRAW bridge on Windows.
func (h *BrandingProxyHandler) CreateAIDesignDraft(c *gin.Context) {
	var request struct {
		Instructions string `json:"instructions" binding:"max=3000"`
	}
	if err := c.ShouldBindJSON(&request); err != nil && err != io.EOF {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Arahan editor harus berupa teks maksimal 3000 karakter"})
		return
	}
	banner, err := h.findBannerOrder(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Request banner tidak ditemukan"})
		return
	}

	userIDValue, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Identitas pengguna tidak ditemukan"})
		return
	}
	userID, ok := userIDValue.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Identitas pengguna tidak valid"})
		return
	}
	var requestedBy *uuid.UUID
	if employee, empErr := h.employeeUseCase.GetEmployeeByUserID(userID); empErr == nil && employee != nil && employee.ID != uuid.Nil {
		requestedBy = &employee.ID
	}
	if !canManageAIDraft(c, banner, requestedBy) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Ambil atau assign tugas ini sebelum membuat draft AI"})
		return
	}

	if banner.SizeLength <= 0 || banner.SizeWidth <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ukuran banner belum lengkap"})
		return
	}

	var activeCount int64
	if err := h.localDB.Model(&domain.BrandingAIDesignJob{}).
		Where("banner_order_id = ? AND status IN ?", banner.ID, []string{aiQueuedHermes, aiHermesProcessing, aiAwaitingCorel, aiCorelProcessing}).
		Count(&activeCount).Error; err != nil {
		log.Printf("[BrandingAI] Schema check triggered on Count error: %v", err)
		ensureBrandingAIDesignJobsTable(h.localDB)
		h.localDB.Model(&domain.BrandingAIDesignJob{}).
			Where("banner_order_id = ? AND status IN ?", banner.ID, []string{aiQueuedHermes, aiHermesProcessing, aiAwaitingCorel, aiCorelProcessing}).
			Count(&activeCount)
	}
	if activeCount > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "Draft AI untuk banner ini masih diproses"})
		return
	}

	type referenceRow struct {
		ID         uuid.UUID `gorm:"column:id"`
		DesignLink string    `gorm:"column:design_link"`
	}
	var reference referenceRow
	ratio := banner.SizeLength / banner.SizeWidth
	h.localDB.Raw(`
		SELECT id, COALESCE(NULLIF(design_link, ''), design_url) AS design_link
		FROM banner_orders
		WHERE id <> ?
		  AND COALESCE(NULLIF(design_link, ''), design_url, '') <> ''
		  AND COALESCE(size_width, 0) > 0
		ORDER BY
		  CASE WHEN LOWER(COALESCE(banner_type, '')) = LOWER(?) THEN 0 ELSE 1 END,
		  ABS((COALESCE(size_length, 0) / NULLIF(size_width, 0)) - ?),
		  CASE WHEN is_design_approved THEN 0 ELSE 1 END,
		  updated_at DESC
		LIMIT 1`, banner.ID, banner.BannerType, ratio).Scan(&reference)

	var referenceID *uuid.UUID
	var referenceURL *string
	if reference.ID != uuid.Nil && reference.DesignLink != "" {
		referenceID = &reference.ID
		referenceURL = &reference.DesignLink
	}
	notes := strings.TrimSpace(valueOrEmpty(banner.Notes))
	prompt := fmt.Sprintf(`Buat draft banner editable berbentuk SVG berdasarkan brief berikut.
Ukuran akhir wajib %.2f cm x %.2f cm (panjang x lebar), viewBox proporsional, tanpa mengubah rasio.
Jenis banner: %s. Nama toko: %s. Jumlah: %d.
Materi dari sales: %s
Gunakan gambar referensi sebagai inspirasi komposisi, warna, dan hierarki. Jangan menyalin teks/nama toko lama dari gambar referensi.
Semua teks baru harus berupa elemen text editable. Pisahkan background, gambar, logo, judul, isi, dan kontak ke group SVG dengan id yang jelas.
Hindari font langka; gunakan Arial, Helvetica, atau sans-serif. Sisakan safe area 2%% dan jangan menaruh teks di luar artboard.
Simpan hasil sebagai draft.svg dan preview sebagai preview.png.`, banner.SizeLength, banner.SizeWidth, banner.BannerType, banner.StoreName, banner.Qty, notes)
	if revision := strings.TrimSpace(valueOrEmpty(banner.RevisionNote)); revision != "" {
		prompt += "\nCatatan revisi yang wajib diterapkan:\n" + revision
	}
	if instructions := strings.TrimSpace(request.Instructions); instructions != "" {
		prompt += "\nArahan tambahan editor:\n" + instructions
	}

	job := domain.BrandingAIDesignJob{
		BannerOrderID:     banner.ID,
		RequestedBy:       requestedBy,
		Status:            aiQueuedHermes,
		ReferenceBannerID: referenceID,
		ReferenceImageURL: referenceURL,
		Prompt:            prompt,
	}
	if err := h.localDB.Create(&job).Error; err != nil {
		log.Printf("[BrandingAI] Gagal membuat antrean draft AI untuk banner %s: %v", banner.ID, err)
		ensureBrandingAIDesignJobsTable(h.localDB)
		if retryErr := h.localDB.Create(&job).Error; retryErr != nil {
			log.Printf("[BrandingAI] Retry gagal membuat antrean draft AI untuk banner %s: %v", banner.ID, retryErr)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat antrean draft AI"})
			return
		}
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Draft AI masuk antrean Hermes", "data": job})
}

func valueOrEmpty(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func (h *BrandingProxyHandler) GetLatestAIDesignDraft(c *gin.Context) {
	banner, err := h.findBannerOrder(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Request banner tidak ditemukan"})
		return
	}
	var job domain.BrandingAIDesignJob
	err = h.localDB.Where("banner_order_id = ?", banner.ID).Order("created_at DESC").First(&job).Error
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"data": nil})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": job})
}

func (h *BrandingProxyHandler) ClaimNextAIDesignJob(c *gin.Context) {
	stage := strings.ToLower(c.DefaultQuery("stage", "hermes"))
	queuedStatus, processingStatus := aiQueuedHermes, aiHermesProcessing
	if stage == "corel" {
		queuedStatus, processingStatus = aiAwaitingCorel, aiCorelProcessing
	} else if stage != "hermes" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "stage harus hermes atau corel"})
		return
	}

	var job domain.BrandingAIDesignJob
	err := h.localDB.Raw(`
		UPDATE branding_ai_design_jobs
		SET status = ?, claimed_at = NOW(), updated_at = NOW(), error_message = NULL
		WHERE id = (
			SELECT id FROM branding_ai_design_jobs
			WHERE status = ? OR (status = ? AND claimed_at < NOW() - INTERVAL '30 minutes')
			ORDER BY created_at ASC
			FOR UPDATE SKIP LOCKED
			LIMIT 1
		)
		RETURNING *`, processingStatus, queuedStatus, processingStatus).Scan(&job).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil antrean AI"})
		return
	}
	if job.ID == uuid.Nil {
		c.Status(http.StatusNoContent)
		return
	}
	var banner domain.BannerOrder
	if err := h.localDB.First(&banner, "id = ?", job.BannerOrderID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Data banner untuk pekerjaan tidak ditemukan"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{
		"job": job,
		"banner": gin.H{
			"id": banner.ID, "store_name": banner.StoreName, "banner_type": banner.BannerType,
			"size_length": banner.SizeLength, "size_width": banner.SizeWidth, "qty": banner.Qty,
			"notes": banner.Notes, "documentation_image_url": banner.DocumentationImageURL,
		},
	}})
}

func validUploadExtension(fileName string, allowed ...string) bool {
	ext := strings.ToLower(filepath.Ext(fileName))
	for _, candidate := range allowed {
		if ext == candidate {
			return true
		}
	}
	return false
}

func (h *BrandingProxyHandler) SubmitHermesDesignResult(c *gin.Context) {
	jobID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID pekerjaan tidak valid"})
		return
	}
	if h.gdrive == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Google Drive tidak tersedia"})
		return
	}
	file, header, err := c.Request.FormFile("svg_file")
	if err != nil || !validUploadExtension(header.Filename, ".svg") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "svg_file berformat SVG wajib dikirim"})
		return
	}
	defer file.Close()
	svgURL, err := h.gdrive.UploadFileWithName(file, fmt.Sprintf("ai_%s_draft.svg", jobID))
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Gagal mengunggah SVG: " + err.Error()})
		return
	}

	updates := map[string]interface{}{"svg_url": svgURL, "status": aiAwaitingCorel, "updated_at": time.Now(), "claimed_at": nil}
	if preview, previewHeader, previewErr := c.Request.FormFile("preview_file"); previewErr == nil {
		defer preview.Close()
		if validUploadExtension(previewHeader.Filename, ".png", ".jpg", ".jpeg") {
			if previewURL, uploadErr := h.gdrive.UploadFileWithName(preview, fmt.Sprintf("ai_%s_preview%s", jobID, strings.ToLower(filepath.Ext(previewHeader.Filename)))); uploadErr == nil {
				updates["preview_url"] = previewURL
			}
		}
	}
	result := h.localDB.Model(&domain.BrandingAIDesignJob{}).
		Where("id = ? AND status = ?", jobID, aiHermesProcessing).Updates(updates)
	if result.Error != nil || result.RowsAffected == 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "Pekerjaan tidak sedang diproses Hermes"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "SVG diterima; menunggu CorelDRAW", "svg_url": svgURL})
}

func (h *BrandingProxyHandler) SubmitCorelDesignResult(c *gin.Context) {
	jobID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID pekerjaan tidak valid"})
		return
	}
	if h.gdrive == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Google Drive tidak tersedia"})
		return
	}
	file, header, err := c.Request.FormFile("cdr_file")
	if err != nil || !validUploadExtension(header.Filename, ".cdr") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cdr_file berformat CDR wajib dikirim"})
		return
	}
	defer file.Close()
	fileName := filepath.Base(header.Filename)
	cdrURL, err := h.gdrive.UploadFileWithName(file, fmt.Sprintf("ai_%s_%s", jobID, fileName))
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Gagal mengunggah CDR: " + err.Error()})
		return
	}
	now := time.Now()
	result := h.localDB.Model(&domain.BrandingAIDesignJob{}).
		Where("id = ? AND status = ?", jobID, aiCorelProcessing).
		Updates(map[string]interface{}{"cdr_url": cdrURL, "cdr_file_name": fileName, "status": aiCompleted, "completed_at": now, "updated_at": now})
	if result.Error != nil || result.RowsAffected == 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "Pekerjaan tidak sedang diproses CorelDRAW"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Draft CDR siap diedit", "cdr_url": cdrURL})
}

func (h *BrandingProxyHandler) FailAIDesignJob(c *gin.Context) {
	jobID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID pekerjaan tidak valid"})
		return
	}
	var request struct {
		Error string `json:"error" binding:"required"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Pesan error wajib diisi"})
		return
	}
	message := request.Error
	if len(message) > 2000 {
		message = message[:2000]
	}
	result := h.localDB.Model(&domain.BrandingAIDesignJob{}).
		Where("id = ? AND status IN ?", jobID, []string{aiHermesProcessing, aiCorelProcessing}).
		Updates(map[string]interface{}{"status": aiFailed, "error_message": message, "updated_at": time.Now()})
	if result.Error != nil || result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Pekerjaan tidak ditemukan"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Pekerjaan ditandai gagal"})
}
