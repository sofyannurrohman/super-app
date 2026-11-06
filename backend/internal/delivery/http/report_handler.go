package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"example.com/hris/backend/internal/scheduler"
	"example.com/hris/backend/pkg/whatsapp"
)

// ReportHandler menyediakan endpoint HTTP untuk:
//   - GET  /reports/groups    → list semua group WA (bantu cari Group ID)
//   - POST /reports/send-now  → trigger laporan manual (untuk testing)
type ReportHandler struct {
	scheduler *scheduler.ReportScheduler
	waClient  *whatsapp.Client
}

func NewReportHandler(sch *scheduler.ReportScheduler, waClient *whatsapp.Client) *ReportHandler {
	return &ReportHandler{scheduler: sch, waClient: waClient}
}

func (h *ReportHandler) RegisterRoutes(r *gin.RouterGroup) {
	reports := r.Group("/reports")
	{
		// Daftar semua group WA yang diikuti — gunakan ini untuk cari Group JID target
		reports.GET("/groups", h.GetGroups)

		// Trigger laporan sekarang tanpa menunggu 17:30 — hanya untuk testing
		reports.POST("/send-now", h.SendNow)
	}
}

// GetGroups mengembalikan daftar group WA.
// Gunakan endpoint ini setelah scan QR untuk mendapatkan Group JID
// yang kemudian diisi ke env var WA_REPORT_GROUP_ID.
func (h *ReportHandler) GetGroups(c *gin.Context) {
	groups, err := h.waClient.GetGroups(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Gagal mengambil daftar group",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total":  len(groups),
		"groups": groups,
		"tip":    "Salin JID yang sesuai ke env var WA_REPORT_GROUP_ID",
	})
}

// SendNow memicu pembuatan dan pengiriman laporan secara manual.
// Berguna untuk test tanpa menunggu jam 17:30.
func (h *ReportHandler) SendNow(c *gin.Context) {
	if err := h.scheduler.TriggerNow(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Gagal mengirim laporan",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "✅ Laporan berhasil dikirim ke WhatsApp group",
	})
}
