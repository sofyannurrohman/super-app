package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
	"example.com/hris/backend/internal/domain"
	"example.com/hris/backend/pkg/utils"
	"gorm.io/gorm"
)

type ExpiryAlertPayload struct {
	Type         string `json:"type"`
	Category     string `json:"category"`
	ItemName     string `json:"item_name"`
	Detail       string `json:"detail"`
	DaysLeft     int    `json:"days_left"`
	ExpiryDate   string `json:"expiry_date"`
}

func StartEmailWorker(redisClient *redis.Client, db *gorm.DB, emailSender utils.EmailSender) {
	log.Println("[EmailWorker] Starting redis queue listener for email_alerts...")
	ctx := context.Background()

	for {
		// BLPop blocks until an item is available in the list
		result, err := redisClient.BLPop(ctx, 0, "queue:email_alerts").Result()
		if err != nil {
			log.Printf("[EmailWorker] Redis BLPop error: %v, retrying in 5s...", err)
			time.Sleep(5 * time.Second)
			continue
		}

		if len(result) < 2 {
			continue
		}

		message := result[1]
		var payload ExpiryAlertPayload
		if err := json.Unmarshal([]byte(message), &payload); err != nil {
			log.Printf("[EmailWorker] Failed to parse message: %v", err)
			continue
		}

		if payload.Type == "EXPIRY_ALERT" {
			handleExpiryAlert(db, emailSender, payload)
		}
	}
}

func handleExpiryAlert(db *gorm.DB, emailSender utils.EmailSender, payload ExpiryAlertPayload) {
	// Find Legal Employees
	var legalUsers []domain.User
	if err := db.Joins("JOIN employees ON employees.user_id = users.id").
		Joins("JOIN job_positions ON job_positions.id = employees.job_position_id").
		Where("job_positions.title LIKE ?", "%Legal%").
		Find(&legalUsers).Error; err != nil {
		log.Printf("[EmailWorker] Failed to fetch legal users: %v", err)
		return
	}

	if len(legalUsers) == 0 {
		log.Printf("[EmailWorker] No legal users found to send email to")
		return
	}

	subject := fmt.Sprintf("Peringatan Expired: %s (%s)", payload.Category, payload.ItemName)
	
	statusText := "akan expired"
	if payload.DaysLeft == 0 {
		statusText = "TELAH EXPIRED"
	}

	body := fmt.Sprintf(`
		<h2>Peringatan Masa Aktif Dokumen</h2>
		<p>Halo Tim Legal,</p>
		<p>Sistem mendeteksi bahwa dokumen berikut %s dalam waktu <b>%d hari</b>.</p>
		<br/>
		<table border="1" cellpadding="10" cellspacing="0">
			<tr>
				<td><b>Kategori</b></td>
				<td>%s</td>
			</tr>
			<tr>
				<td><b>Nama Item</b></td>
				<td>%s</td>
			</tr>
			<tr>
				<td><b>Nomor/Detail</b></td>
				<td>%s</td>
			</tr>
			<tr>
				<td><b>Tanggal Expired</b></td>
				<td>%s</td>
			</tr>
		</table>
		<br/>
		<p>Mohon segera lakukan pengecekan dan perpanjangan dokumen terkait.</p>
		<p>Terima kasih,</p>
		<p>Sistem Notifikasi HRIS</p>
	`, statusText, payload.DaysLeft, payload.Category, payload.ItemName, payload.Detail, payload.ExpiryDate)

	for _, user := range legalUsers {
		if user.Email != "" {
			err := emailSender.SendEmail(user.Email, subject, body)
			if err != nil {
				log.Printf("[EmailWorker] Failed to send email to %s: %v", user.Email, err)
			} else {
				log.Printf("[EmailWorker] Successfully sent email to %s for %s", user.Email, payload.ItemName)
			}
		}
	}
}
