package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"example.com/hris/backend/internal/usecase"
)

// FcmHandler menyediakan endpoint admin untuk:
//   - POST /admin/fcm/broadcast   → kirim notifikasi ke semua user
//   - POST /admin/fcm/send-user   → kirim notifikasi ke 1 user tertentu
//   - POST /admin/fcm/test-reminder → trigger pengecekan reminder absen sekarang
type FcmHandler struct {
	fcmUsecase usecase.FcmUsecase
}

func NewFcmHandler(fcmUsecase usecase.FcmUsecase) *FcmHandler {
	return &FcmHandler{fcmUsecase: fcmUsecase}
}

// RegisterAdminRoutes mendaftarkan route admin (tidak memerlukan JWT karyawan biasa,
// namun tetap di-protect oleh AuthMiddleware — hanya bisa dipakai oleh HR/Admin).
func (h *FcmHandler) RegisterAdminRoutes(rg *gin.RouterGroup) {
	fcm := rg.Group("/admin/fcm")
	{
		// Broadcast ke semua user yang terdaftar FCM token-nya
		fcm.POST("/broadcast", h.broadcast)

		// Kirim notifikasi ke user tertentu berdasarkan user_id
		fcm.POST("/send-user", h.sendToUser)

		// Trigger pengecekan reminder absen manual (berguna untuk debugging)
		fcm.POST("/test-reminder", h.testReminder)
	}
}

// broadcast kirim notifikasi push ke semua device yang terdaftar.
//
// Body JSON:
//
//	{ "title": "Judul", "body": "Isi pesan" }
func (h *FcmHandler) broadcast(c *gin.Context) {
	var req struct {
		Title string `json:"title" binding:"required"`
		Body  string `json:"body"  binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "title dan body wajib diisi"})
		return
	}

	if err := h.fcmUsecase.BroadcastNotification(req.Title, req.Body); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Broadcast FCM berhasil dikirim ke semua device"})
}

// sendToUser kirim notifikasi push ke user tertentu.
//
// Body JSON:
//
//	{ "user_id": "uuid", "title": "Judul", "body": "Isi pesan" }
func (h *FcmHandler) sendToUser(c *gin.Context) {
	var req struct {
		UserID string `json:"user_id" binding:"required"`
		Title  string `json:"title"   binding:"required"`
		Body   string `json:"body"    binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id, title, dan body wajib diisi"})
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id tidak valid (harus UUID)"})
		return
	}

	if err := h.fcmUsecase.SendPushNotification(userID, req.Title, req.Body); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Notifikasi berhasil dikirim",
		"user_id": req.UserID,
	})
}

// testReminder memicu pengecekan reminder absen secara langsung (sync, bukan async).
// Berguna untuk debugging di staging/dev tanpa harus menunggu ticker 1 menit.
func (h *FcmHandler) testReminder(c *gin.Context) {
	// Panggil melalui interface dengan type assertion agar tidak perlu expose method internal
	type reminderTrigger interface {
		TriggerAttendanceReminders()
	}
	if t, ok := h.fcmUsecase.(reminderTrigger); ok {
		t.TriggerAttendanceReminders()
		c.JSON(http.StatusOK, gin.H{"message": "Pengecekan reminder absen dipicu secara manual. Cek server log untuk hasilnya."})
	} else {
		c.JSON(http.StatusNotImplemented, gin.H{"message": "Fitur trigger manual belum tersedia"})
	}
}
