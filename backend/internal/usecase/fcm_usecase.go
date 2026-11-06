package usecase

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"google.golang.org/api/option"
	"gorm.io/gorm"
)

type FcmUsecase interface {
	StartReminderWorker(ctx context.Context)
	SendPushNotification(userID uuid.UUID, title, body string) error
	SendLeaveNotification(userID uuid.UUID, title, body string, leaveData map[string]string) error
	SendWarningNotification(userID uuid.UUID, title, body string, warningData map[string]string) error
	BroadcastNotification(title, body string) error
	// SendToUsers mengirim notifikasi push ke daftar user tertentu dengan data payload tambahan.
	// data["type"] bisa "popup" atau "notification", data["announcement_id"] berisi UUID pengumuman.
	SendToUsers(userIDs []uuid.UUID, title, body string, data map[string]string) error
	// SendForceLogoutExcept mengirim notifikasi FORCE_LOGOUT ke token user yang berbeda dari excludeToken.
	SendForceLogoutExcept(userID uuid.UUID, excludeToken string, newSessionID string) error
	// SendRemoteSyncTrigger mengirim silent push notification berkategori REMOTE_SYNC untuk memicu sync offline mobile.
	SendRemoteSyncTrigger(userID uuid.UUID) error
	// BroadcastRemoteSyncTrigger mengirim trigger silent sync ke daftar sales user.
	BroadcastRemoteSyncTrigger(userIDs []uuid.UUID) error
	// TriggerAttendanceReminders memicu pengecekan reminder absen secara langsung.
	// Berguna untuk debugging manual tanpa menunggu ticker 1 menit.
	TriggerAttendanceReminders()
}

type fcmUsecase struct {
	db        *gorm.DB
	fcmClient *messaging.Client
	sentMap   sync.Map // Key: "YYYY-MM-DD:userID:checkType", Value: time.Time
}

func NewFcmUsecase(db *gorm.DB, credentialsFile string) FcmUsecase {
	if credentialsFile == "" {
		return &fcmUsecase{db: db, fcmClient: nil}
	}
	opt := option.WithCredentialsFile(credentialsFile)
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		log.Printf("ERROR: Failed to initialize Firebase App: %v. FCM features will be disabled.\n", err)
		return &fcmUsecase{db: db, fcmClient: nil}
	}
	client, err := app.Messaging(context.Background())
	if err != nil {
		log.Printf("ERROR: Failed to get Firebase Messaging client: %v. FCM features will be disabled.\n", err)
		return &fcmUsecase{db: db, fcmClient: nil}
	}
	log.Println("SUCCESS: Firebase FCM client initialized successfully")
	return &fcmUsecase{
		db:        db,
		fcmClient: client,
	}
}

func (u *fcmUsecase) SendPushNotification(userID uuid.UUID, title, body string) error {
	if u.fcmClient == nil {
		return fmt.Errorf("FCM client is not initialized")
	}

	go func() {
		// Fetch token for user
		var tokenRecord domain.UserFcmToken
		err := u.db.Where("user_id = ?", userID).First(&tokenRecord).Error
		if err != nil {
			log.Printf("[FCM] ⚠️ FCM token not found for user %s: %v\n", userID, err)
			return
		}

		// Construct message with high priority, alarm sound, and explicit notification channel
		message := &messaging.Message{
			Token: tokenRecord.FCMToken,
			Notification: &messaging.Notification{
				Title: title,
				Body:  body,
			},
			Data: map[string]string{
				"type":  "reminder",
				"title": title,
				"body":  body,
			},
			Android: &messaging.AndroidConfig{
				Priority: "high",
				Notification: &messaging.AndroidNotification{
					ChannelID:             "shift_alerts_alarm_v2",
					Sound:                 "alarm",
					Priority:              messaging.PriorityMax,
					DefaultVibrateTimings: true,
				},
			},
			APNS: &messaging.APNSConfig{
				Payload: &messaging.APNSPayload{
					Aps: &messaging.Aps{
						Sound:            "alarm.mp3",
						ContentAvailable: true,
					},
				},
			},
		}

		_, err = u.fcmClient.Send(context.Background(), message)
		if err != nil {
			log.Printf("[FCM] ❌ Failed to send FCM message to user %s: %v\n", userID, err)
			return
		}

		log.Printf("[FCM] ✅ SUCCESS: Sent FCM reminder to user %s: %s\n", userID, title)
	}()

	return nil
}

func (u *fcmUsecase) SendLeaveNotification(userID uuid.UUID, title, body string, leaveData map[string]string) error {
	if u.fcmClient == nil {
		return fmt.Errorf("FCM client is not initialized")
	}

	go func() {
		var tokenRecord domain.UserFcmToken
		err := u.db.Where("user_id = ?", userID).First(&tokenRecord).Error
		if err != nil {
			log.Printf("[FCM] ⚠️ FCM token not found for user %s: %v\n", userID, err)
			return
		}

		if leaveData == nil {
			leaveData = make(map[string]string)
		}
		leaveData["type"] = "leave"
		leaveData["title"] = title
		leaveData["body"] = body

		message := &messaging.Message{
			Token: tokenRecord.FCMToken,
			Notification: &messaging.Notification{
				Title: title,
				Body:  body,
			},
			Data: leaveData,
			Android: &messaging.AndroidConfig{
				Priority: "high",
				Notification: &messaging.AndroidNotification{
					ChannelID:             "fcm_alerts_channel_v1",
					Sound:                 "default",
					Priority:              messaging.PriorityHigh,
					DefaultVibrateTimings: true,
				},
			},
			APNS: &messaging.APNSConfig{
				Payload: &messaging.APNSPayload{
					Aps: &messaging.Aps{
						Sound:            "default",
						ContentAvailable: true,
					},
				},
			},
		}

		_, err = u.fcmClient.Send(context.Background(), message)
		if err != nil {
			log.Printf("[FCM] ❌ Failed to send FCM leave message to user %s: %v\n", userID, err)
			return
		}

		log.Printf("[FCM] ✅ SUCCESS: Sent FCM leave notification to user %s: %s\n", userID, title)
	}()

	return nil
}

func (u *fcmUsecase) SendWarningNotification(userID uuid.UUID, title, body string, warningData map[string]string) error {
	if u.fcmClient == nil {
		return fmt.Errorf("FCM client is not initialized")
	}

	go func() {
		var tokenRecord domain.UserFcmToken
		err := u.db.Where("user_id = ?", userID).First(&tokenRecord).Error
		if err != nil {
			log.Printf("[FCM] ⚠️ FCM token not found for user %s: %v\n", userID, err)
			return
		}

		if warningData == nil {
			warningData = make(map[string]string)
		}
		warningData["type"] = "warning"
		warningData["title"] = title
		warningData["body"] = body

		message := &messaging.Message{
			Token: tokenRecord.FCMToken,
			Notification: &messaging.Notification{
				Title: title,
				Body:  body,
			},
			Data: warningData,
			Android: &messaging.AndroidConfig{
				Priority: "high",
				Notification: &messaging.AndroidNotification{
					ChannelID:             "hr_warning_alerts",
					Sound:                 "alarm",
					Priority:              messaging.PriorityMax,
					DefaultVibrateTimings: true,
				},
			},
			APNS: &messaging.APNSConfig{
				Payload: &messaging.APNSPayload{
					Aps: &messaging.Aps{
						Sound:            "alarm.mp3",
						ContentAvailable: true,
					},
				},
			},
		}

		_, err = u.fcmClient.Send(context.Background(), message)
		if err != nil {
			log.Printf("[FCM] ❌ Failed to send FCM warning message to user %s: %v\n", userID, err)
			return
		}

		log.Printf("[FCM] ✅ SUCCESS: Sent FCM warning to user %s: %s\n", userID, title)
	}()

	return nil
}

func (u *fcmUsecase) StartReminderWorker(ctx context.Context) {
	log.Println("[FCM-WORKER] ✅ Attendance reminder worker started (interval: 1 menit)")
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			u.sendAttendanceReminders()
		case <-ctx.Done():
			log.Println("[FCM-WORKER] 🛑 Attendance reminder worker dihentikan.")
			return
		}
	}
}

// TriggerAttendanceReminders adalah public wrapper untuk memanggil
// sendAttendanceReminders secara manual (misal dari HTTP handler admin).
func (u *fcmUsecase) TriggerAttendanceReminders() {
	log.Println("[FCM-WORKER] 🔔 Manual trigger: pengecekan reminder absen dipanggil")
	u.sendAttendanceReminders()
}

type employeeResolvedShift struct {
	EmployeeID uuid.UUID  `gorm:"column:employee_id"`
	UserID     uuid.UUID  `gorm:"column:user_id"`
	FirstName  string     `gorm:"column:first_name"`
	LastName   string     `gorm:"column:last_name"`
	BranchID   *uuid.UUID `gorm:"column:branch_id"`
	BranchName string     `gorm:"column:branch_name"`
	Timezone   string     `gorm:"column:timezone"`
	ShiftID    *uuid.UUID `gorm:"column:shift_id"`
	ShiftName  string     `gorm:"column:shift_name"`
	StartTime  string     `gorm:"column:start_time"`
	EndTime    string     `gorm:"column:end_time"`
	IsOffDay   bool       `gorm:"column:is_off_day"`
}

func (u *fcmUsecase) sendAttendanceReminders() {
	// Query seluruh karyawan aktif dan shift yang berlaku hari ini
	// Prioritas: 1. Roster Harian (employee_shifts) -> 2. Default Shift Karyawan -> 3. Default Shift Cabang
	var employees []employeeResolvedShift
	query := `
		SELECT 
			e.id AS employee_id,
			u.id AS user_id,
			e.first_name,
			COALESCE(e.last_name, '') AS last_name,
			e.branch_id,
			COALESCE(b.name, 'Pusat') AS branch_name,
			COALESCE(NULLIF(b.timezone, ''), 'Asia/Jakarta') AS timezone,
			COALESCE(es.shift_id, e.default_shift_id, b.default_shift_id) AS shift_id,
			COALESCE(es_s.name, def_s.name, br_s.name, 'Shift Reguler') AS shift_name,
			COALESCE(es_s.start_time::text, def_s.start_time::text, br_s.start_time::text, '') AS start_time,
			COALESCE(es_s.end_time::text, def_s.end_time::text, br_s.end_time::text, '') AS end_time,
			COALESCE(es.is_off_day, FALSE) AS is_off_day
		FROM users u
		JOIN employees e ON e.user_id = u.id
		LEFT JOIN branches b ON b.id = e.branch_id
		-- 1. Daily Roster
		LEFT JOIN employee_shifts es ON es.employee_id = e.id AND es.date = CURRENT_DATE
		LEFT JOIN shifts es_s ON es_s.id = es.shift_id
		-- 2. Employee Default Shift
		LEFT JOIN shifts def_s ON def_s.id = e.default_shift_id
		-- 3. Branch Default Shift
		LEFT JOIN shifts br_s ON br_s.id = b.default_shift_id
		WHERE u.is_active = TRUE
		  AND COALESCE(es.is_off_day, FALSE) = FALSE
	`

	if err := u.db.Raw(query).Scan(&employees).Error; err != nil {
		log.Printf("[FCM-WORKER] ❌ Error fetching active employees & shifts: %v\n", err)
		return
	}

	if len(employees) == 0 {
		return
	}

	// Cache locations
	locations := make(map[string]*time.Location)

	for _, emp := range employees {
		if emp.StartTime == "" || emp.StartTime == "-" || emp.EndTime == "" || emp.EndTime == "-" {
			continue
		}

		tzName := emp.Timezone
		if tzName == "" {
			tzName = "Asia/Jakarta"
		}

		loc, ok := locations[tzName]
		if !ok {
			var err error
			loc, err = time.LoadLocation(tzName)
			if err != nil {
				loc = time.FixedZone("WIB", 7*3600)
			}
			locations[tzName] = loc
		}

		now := time.Now().In(loc)
		todayStr := now.Format("2006-01-02")

		startTime := normalizeTimeStr(emp.StartTime)
		endTime := normalizeTimeStr(emp.EndTime)

		var startHour, startMin, endHour, endMin int
		fmt.Sscanf(startTime, "%d:%d", &startHour, &startMin)
		fmt.Sscanf(endTime, "%d:%d", &endHour, &endMin)

		shiftStart := time.Date(now.Year(), now.Month(), now.Day(), startHour, startMin, 0, 0, loc)

		var shiftEnd time.Time
		if startHour*60+startMin > endHour*60+endMin {
			// Night shift ends tomorrow
			nextDay := shiftStart.Add(24 * time.Hour)
			shiftEnd = time.Date(nextDay.Year(), nextDay.Month(), nextDay.Day(), endHour, endMin, 0, 0, loc)
		} else {
			shiftEnd = time.Date(now.Year(), now.Month(), now.Day(), endHour, endMin, 0, 0, loc)
		}

		// ─── 1. CEK REMINDER MASUK (IN) ──────────────────────────────
		// Target: 5 menit sebelum jam masuk (Window: 10 menit s/d 2 menit sebelum start)
		clockInReminderWindowStart := shiftStart.Add(-10 * time.Minute)
		clockInReminderWindowEnd := shiftStart.Add(2 * time.Minute)

		if now.After(clockInReminderWindowStart) && now.Before(clockInReminderWindowEnd) {
			dedupKey := fmt.Sprintf("%s:%s:IN", todayStr, emp.UserID)
			if _, alreadySent := u.sentMap.Load(dedupKey); !alreadySent {
				// Cek apakah user sudah absen masuk hari ini
				todayDayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
				var count int64
				u.db.Model(&domain.AttendanceLog{}).
					Where("employee_id = ? AND clock_in_time >= ?", emp.EmployeeID, todayDayStart).
					Count(&count)

				if count == 0 {
					log.Printf("[FCM-WORKER] 📤 Mengirim reminder MASUK ke %s %s (%s) Shift: %s",
						emp.FirstName, emp.LastName, emp.UserID, startTime)
					u.sentMap.Store(dedupKey, time.Now())
					go u.SendPushNotification(
						emp.UserID,
						"⏰ Persiapan Presensi Masuk",
						fmt.Sprintf("Waktu presensi masuk Anda (%s) segera dimulai. Yuk bersiap melakukan presensi di Karsa HR! 👋", startTime),
					)
				} else {
					u.sentMap.Store(dedupKey, time.Now())
				}
			}
		}

		// ─── 2. CEK REMINDER KELUAR 5 MENIT (OUT) ─────────────────────
		// Target: 5 menit sebelum jam pulang (Window: 8 menit s/d 1 menit sebelum end)
		clockOutReminderWindowStart := shiftEnd.Add(-8 * time.Minute)
		clockOutReminderWindowEnd := shiftEnd.Add(-1 * time.Minute)

		if now.After(clockOutReminderWindowStart) && now.Before(clockOutReminderWindowEnd) {
			dedupKey := fmt.Sprintf("%s:%s:OUT", todayStr, emp.UserID)
			if _, alreadySent := u.sentMap.Load(dedupKey); !alreadySent {
				// Cek apakah user sudah absen masuk dan belum absen keluar
				todayDayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
				var logRecord domain.AttendanceLog
				err := u.db.Where("employee_id = ? AND clock_in_time >= ?", emp.EmployeeID, todayDayStart).
					Order("clock_in_time DESC").
					First(&logRecord).Error

				if err == nil && logRecord.ClockInTime != nil && logRecord.ClockOutTime == nil {
					log.Printf("[FCM-WORKER] 📤 Mengirim reminder KELUAR (5m) ke %s %s (%s) Shift End: %s",
						emp.FirstName, emp.LastName, emp.UserID, endTime)
					u.sentMap.Store(dedupKey, time.Now())
					go u.SendPushNotification(
						emp.UserID,
						"⏰ Sesi Kerja Hampir Berakhir",
						fmt.Sprintf("5 menit lagi waktu kerja Anda berakhir (%s). Jangan lupa lakukan presensi pulang! 🏃", endTime),
					)
				} else {
					u.sentMap.Store(dedupKey, time.Now())
				}
			}
		}

		// ─── 3. CEK REMINDER FINAL KELUAR (FINAL) ─────────────────────
		// Target: Tepat waktu pulang hingga 30 menit setelahnya
		clockOutFinalWindowStart := shiftEnd.Add(-1 * time.Minute)
		clockOutFinalWindowEnd := shiftEnd.Add(30 * time.Minute)

		if now.After(clockOutFinalWindowStart) && now.Before(clockOutFinalWindowEnd) {
			dedupKey := fmt.Sprintf("%s:%s:FINAL", todayStr, emp.UserID)
			if _, alreadySent := u.sentMap.Load(dedupKey); !alreadySent {
				todayDayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
				var logRecord domain.AttendanceLog
				err := u.db.Where("employee_id = ? AND clock_in_time >= ?", emp.EmployeeID, todayDayStart).
					Order("clock_in_time DESC").
					First(&logRecord).Error

				if err == nil && logRecord.ClockInTime != nil && logRecord.ClockOutTime == nil {
					log.Printf("[FCM-WORKER] 📤 Mengirim reminder FINAL ke %s %s (%s) Shift End: %s",
						emp.FirstName, emp.LastName, emp.UserID, endTime)
					u.sentMap.Store(dedupKey, time.Now())
					go u.SendPushNotification(
						emp.UserID,
						"⏰ Waktu Kerja Selesai",
						fmt.Sprintf("Waktu kerja Anda telah berakhir (%s). Terima kasih atas dedikasi Anda hari ini! Silakan lakukan presensi keluar. 🙏", endTime),
					)
				} else {
					u.sentMap.Store(dedupKey, time.Now())
				}
			}
		}
	}
}

func normalizeTimeStr(timeStr string) string {
	timeStr = strings.ReplaceAll(timeStr, ".", ":")
	parts := strings.Split(timeStr, ":")
	if len(parts) >= 2 {
		h := parts[0]
		m := parts[1]
		if len(h) == 1 {
			h = "0" + h
		}
		if len(m) == 1 {
			m = "0" + m
		}
		return h + ":" + m
	}
	return timeStr
}

// SendToUsers mengirim notifikasi push ke sekumpulan user tertentu.
// Mendukung data payload tambahan (misal: type=popup, announcement_id=uuid).
// Pengiriman dilakukan di background goroutine agar tidak memblokir request.
func (u *fcmUsecase) SendToUsers(userIDs []uuid.UUID, title, body string, data map[string]string) error {
	if u.fcmClient == nil {
		return fmt.Errorf("FCM client is not initialized")
	}
	if len(userIDs) == 0 {
		return nil
	}

	go func(ids []uuid.UUID) {
		// Ambil FCM tokens untuk user-user yang ditarget (hanya user & karyawan aktif)
		var tokens []string
		err := u.db.Table("user_fcm_tokens").
			Joins("JOIN users ON users.id = user_fcm_tokens.user_id").
			Joins("LEFT JOIN employees ON employees.user_id = users.id").
			Where("user_fcm_tokens.user_id IN ?", ids).
			Where("users.is_active = TRUE").
			Where("(employees.id IS NULL OR COALESCE(employees.employment_status, 'Active') NOT IN ('Inactive', 'INACTIVE', 'Resigned', 'RESIGNED', 'Terminated', 'TERMINATED'))").
			Pluck("user_fcm_tokens.fcm_token", &tokens).Error
		if err != nil {
			log.Printf("[FCM] ❌ Gagal fetch tokens untuk targeted notif: %v\n", err)
			return
		}
		if len(tokens) == 0 {
			log.Printf("[FCM] ⚠️  Tidak ada FCM token aktif untuk %d user yang ditarget\n", len(ids))
			return
		}

		chunks := chunkTokens(tokens, 500)
		for _, chunk := range chunks {
			message := &messaging.MulticastMessage{
				Tokens: chunk,
				Notification: &messaging.Notification{
					Title: title,
					Body:  body,
				},
				Data: data,
				Android: &messaging.AndroidConfig{
					Priority: "high",
				},
				APNS: &messaging.APNSConfig{
					Payload: &messaging.APNSPayload{
						Aps: &messaging.Aps{
							Sound:            "default",
							ContentAvailable: true,
						},
					},
				},
			}

			response, err := u.fcmClient.SendEachForMulticast(context.Background(), message)
			if err != nil {
				log.Printf("[FCM] ❌ Gagal kirim targeted multicast: %v\n", err)
				continue
			}
			log.Printf("[FCM] ✅ Targeted notif terkirim. Success: %d, Failure: %d\n",
				response.SuccessCount, response.FailureCount)
		}
	}(userIDs)

	return nil
}

func (u *fcmUsecase) SendForceLogoutExcept(userID uuid.UUID, excludeToken string, newSessionID string) error {
	if u.fcmClient == nil {
		return fmt.Errorf("FCM client is not initialized")
	}

	go func() {
		var oldTokens []string
		query := u.db.Model(&domain.UserFcmToken{}).Where("user_id = ?", userID)
		if excludeToken != "" {
			query = query.Where("fcm_token != ?", excludeToken)
		}
		err := query.Pluck("fcm_token", &oldTokens).Error
		if err != nil || len(oldTokens) == 0 {
			// Tidak ada device lama yang perlu di-kick
			return
		}

		title := "Sesi Berakhir"
		body := "Akun Anda telah login di perangkat lain. Anda telah keluar otomatis dari perangkat ini."
		data := map[string]string{
			"type":           "FORCE_LOGOUT",
			"title":          title,
			"body":           body,
			"new_session_id": newSessionID,
		}

		chunks := chunkTokens(oldTokens, 500)
		for _, chunk := range chunks {
			message := &messaging.MulticastMessage{
				Tokens: chunk,
				Notification: &messaging.Notification{
					Title: title,
					Body:  body,
				},
				Data: data,
				Android: &messaging.AndroidConfig{
					Priority: "high",
				},
			}

			response, err := u.fcmClient.SendEachForMulticast(context.Background(), message)
			if err != nil {
				log.Printf("[FCM] ❌ Gagal kirim force logout multicast: %v\n", err)
				continue
			}
			log.Printf("[FCM] ✅ Force logout notif terkirim ke %d device lama. Success: %d, Failure: %d\n",
				len(chunk), response.SuccessCount, response.FailureCount)
		}
	}()

	return nil
}

func (u *fcmUsecase) BroadcastNotification(title, body string) error {

	if u.fcmClient == nil {
		return fmt.Errorf("FCM client is not initialized")
	}

	// Fetch all FCM tokens for active users & active employees only
	var tokens []string
	err := u.db.Table("user_fcm_tokens").
		Joins("JOIN users ON users.id = user_fcm_tokens.user_id").
		Joins("LEFT JOIN employees ON employees.user_id = users.id").
		Where("users.is_active = TRUE").
		Where("(employees.id IS NULL OR COALESCE(employees.employment_status, 'Active') NOT IN ('Inactive', 'INACTIVE', 'Resigned', 'RESIGNED', 'Terminated', 'TERMINATED'))").
		Pluck("user_fcm_tokens.fcm_token", &tokens).Error
	if err != nil {
		return fmt.Errorf("failed to fetch FCM tokens: %v", err)
	}

	if len(tokens) == 0 {
		return nil
	}

	// Send in a background goroutine to avoid blocking the main server thread
	go func(tokenList []string) {
		chunks := chunkTokens(tokenList, 500)
		for _, chunk := range chunks {
			message := &messaging.MulticastMessage{
				Tokens: chunk,
				Notification: &messaging.Notification{
					Title: title,
					Body:  body,
				},
				Android: &messaging.AndroidConfig{
					Priority: "high",
				},
			}

			response, err := u.fcmClient.SendEachForMulticast(context.Background(), message)
			if err != nil {
				log.Printf("ERROR: Failed to send multicast broadcast FCM: %v\n", err)
				continue
			}
			log.Printf("SUCCESS: Dispatched broadcast FCM. Success: %d, Failure: %d\n", response.SuccessCount, response.FailureCount)
		}
	}(tokens)

	return nil
}

func (u *fcmUsecase) SendRemoteSyncTrigger(userID uuid.UUID) error {
	if u.fcmClient == nil {
		return fmt.Errorf("FCM client is not initialized")
	}

	var tokenRecord domain.UserFcmToken
	err := u.db.Where("user_id = ?", userID).First(&tokenRecord).Error
	if err != nil {
		return fmt.Errorf("FCM token not found for user %s: %v", userID, err)
	}

	go func() {
		message := &messaging.Message{
			Token: tokenRecord.FCMToken,
			Data: map[string]string{
				"type":      "REMOTE_SYNC",
				"action":    "FORCE_SYNC_TRANSACTIONS",
				"timestamp": time.Now().UTC().Format(time.RFC3339),
			},
			Android: &messaging.AndroidConfig{
				Priority: "high",
			},
			APNS: &messaging.APNSConfig{
				Payload: &messaging.APNSPayload{
					Aps: &messaging.Aps{
						ContentAvailable: true,
					},
				},
			},
		}

		res, err := u.fcmClient.Send(context.Background(), message)
		if err != nil {
			log.Printf("[FCM] ❌ Failed to send REMOTE_SYNC to user %s: %v\n", userID, err)
			return
		}
		log.Printf("[FCM] 🚀 Sent REMOTE_SYNC signal to user %s (messageID: %s)\n", userID, res)
	}()

	return nil
}

func (u *fcmUsecase) BroadcastRemoteSyncTrigger(userIDs []uuid.UUID) error {
	if u.fcmClient == nil {
		return fmt.Errorf("FCM client is not initialized")
	}

	query := u.db.Table("user_fcm_tokens").
		Joins("JOIN users ON users.id = user_fcm_tokens.user_id").
		Where("users.is_active = TRUE")

	if len(userIDs) > 0 {
		query = query.Where("user_fcm_tokens.user_id IN ?", userIDs)
	}

	var tokens []string
	if err := query.Pluck("user_fcm_tokens.fcm_token", &tokens).Error; err != nil {
		return fmt.Errorf("failed to fetch tokens: %v", err)
	}

	if len(tokens) == 0 {
		return nil
	}

	go func(tokenList []string) {
		chunks := chunkTokens(tokenList, 500)
		for _, chunk := range chunks {
			message := &messaging.MulticastMessage{
				Tokens: chunk,
				Data: map[string]string{
					"type":      "REMOTE_SYNC",
					"action":    "FORCE_SYNC_TRANSACTIONS",
					"timestamp": time.Now().UTC().Format(time.RFC3339),
				},
				Android: &messaging.AndroidConfig{
					Priority: "high",
				},
				APNS: &messaging.APNSConfig{
					Payload: &messaging.APNSPayload{
						Aps: &messaging.Aps{
							ContentAvailable: true,
						},
					},
				},
			}

			response, err := u.fcmClient.SendEachForMulticast(context.Background(), message)
			if err != nil {
				log.Printf("[FCM] ❌ Failed to send multicast REMOTE_SYNC: %v\n", err)
				continue
			}
			log.Printf("[FCM] 🚀 Broadcast REMOTE_SYNC dispatched. Success: %d, Failure: %d\n", response.SuccessCount, response.FailureCount)
		}
	}(tokens)

	return nil
}

func chunkTokens(tokens []string, chunkSize int) [][]string {
	var chunks [][]string
	for i := 0; i < len(tokens); i += chunkSize {
		end := i + chunkSize
		if end > len(tokens) {
			end = len(tokens)
		}
		chunks = append(chunks, tokens[i:end])
	}
	return chunks
}
