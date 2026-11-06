package http

import (
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"example.com/hris/backend/pkg/utils"
)

type campaignVideoInput struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	VideoURL    string `json:"video_url"`
	ProductName string `json:"product_name"`
	PurchaseURL string `json:"purchase_url"`
	IsActive    *bool  `json:"is_active"`
}

func validDriveVideoURL(raw string) bool {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "https" || strings.ToLower(u.Hostname()) != "drive.google.com" {
		return false
	}
	id := utils.ExtractGDriveFileID(raw)
	if id == "" || strings.ContainsAny(id, "/?&# ") {
		return false
	}
	return strings.Contains(u.Path, "/file/d/") || (u.Path == "/open" && u.Query().Get("id") == id)
}

func validPurchaseURL(raw string) bool {
	u, err := url.Parse(raw)
	return err == nil && u.Scheme == "https" && u.Hostname() != "" && u.User == nil
}

func (h *BrandingProxyHandler) GetCampaignVideos(c *gin.Context) {
	var videos []domain.BrandingCampaignVideo
	query := h.localDB.Order("created_at DESC")
	if c.Query("active") == "true" {
		query = query.Where("is_active = ?", true)
	}
	if err := query.Find(&videos).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil video kampanye"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": videos})
}

func (h *BrandingProxyHandler) saveCampaignVideo(c *gin.Context, existing *domain.BrandingCampaignVideo) {
	var input campaignVideoInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Data video tidak valid"})
		return
	}
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)
	input.VideoURL = strings.TrimSpace(input.VideoURL)
	input.ProductName = strings.TrimSpace(input.ProductName)
	input.PurchaseURL = strings.TrimSpace(input.PurchaseURL)
	if input.Title == "" || len(input.Title) > 200 || len(input.ProductName) > 200 || input.ProductName == "" || !validDriveVideoURL(input.VideoURL) || !validPurchaseURL(input.PurchaseURL) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Isi judul dan produk, gunakan link video Google Drive serta link pembelian HTTPS yang valid"})
		return
	}
	active := true
	if input.IsActive != nil {
		active = *input.IsActive
	}
	video := domain.BrandingCampaignVideo{
		Title: input.Title, Description: input.Description, VideoURL: input.VideoURL,
		ProductName: input.ProductName, PurchaseURL: input.PurchaseURL,
		IsActive: active, UpdatedAt: time.Now(),
	}
	if existing == nil {
		video.ID = uuid.New()
		if err := h.localDB.Create(&video).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan video kampanye"})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"data": video})
		return
	}
	if err := h.localDB.Model(existing).Updates(map[string]interface{}{
		"title": video.Title, "description": video.Description, "video_url": video.VideoURL,
		"product_name": video.ProductName, "purchase_url": video.PurchaseURL,
		"is_active": video.IsActive, "updated_at": video.UpdatedAt,
	}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memperbarui video kampanye"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": video})
}

func (h *BrandingProxyHandler) CreateCampaignVideo(c *gin.Context) {
	h.saveCampaignVideo(c, nil)
}

func (h *BrandingProxyHandler) UpdateCampaignVideo(c *gin.Context) {
	var video domain.BrandingCampaignVideo
	if err := h.localDB.First(&video, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Video tidak ditemukan"})
		return
	}
	h.saveCampaignVideo(c, &video)
}

func (h *BrandingProxyHandler) DeleteCampaignVideo(c *gin.Context) {
	result := h.localDB.Delete(&domain.BrandingCampaignVideo{}, "id = ?", c.Param("id"))
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus video kampanye"})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Video tidak ditemukan"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Video kampanye dihapus"})
}

func (h *BrandingProxyHandler) UploadCampaignVideo(c *gin.Context) {
	if h.gdrive == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Google Drive belum tersedia. Tempel link video Drive secara manual."})
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 96<<20)
	header, err := c.FormFile("video")
	if err != nil || header.Size == 0 || header.Size > 95<<20 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Pilih video berukuran maksimal 95 MB"})
		return
	}
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext != ".mp4" && ext != ".mov" && ext != ".webm" && ext != ".m4v" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format video harus MP4, MOV, WebM, atau M4V"})
		return
	}
	file, err := header.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Video tidak dapat dibaca"})
		return
	}
	defer file.Close()
	link, err := h.gdrive.UploadFileWithName(file, "kampanye_"+uuid.NewString()+ext)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Gagal mengunggah ke Google Drive"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"video_url": link})
}
