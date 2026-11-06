package scheduler

import (
	"context"
	"fmt"
	"log"
	"time"

	"example.com/hris/backend/internal/domain"
	"example.com/hris/backend/internal/usecase"
	"gorm.io/gorm"
)

type VisitPlanScheduler struct {
	db        *gorm.DB
	fcmClient usecase.FcmUsecase
	tz        *time.Location
}

func NewVisitPlanScheduler(db *gorm.DB, fcmClient usecase.FcmUsecase) *VisitPlanScheduler {
	loc, _ := time.LoadLocation("Asia/Jakarta")
	return &VisitPlanScheduler{
		db:        db,
		fcmClient: fcmClient,
		tz:        loc,
	}
}

func (s *VisitPlanScheduler) Start(ctx context.Context) {
	go func() {
		for {
			now := time.Now().In(s.tz)

			// Target jam: 07:00:00 WIB hari ini
			next := time.Date(now.Year(), now.Month(), now.Day(), 7, 0, 0, 0, s.tz)
			if !now.Before(next) {
				// Sudah lewat jam 07:00 hari ini → jadwalkan besok
				next = next.Add(24 * time.Hour)
			}

			duration := next.Sub(now)
			log.Printf("[VisitPlanScheduler] ⏰ Pengingat kunjungan berikutnya: %s (dalam %s)",
				next.Format("02 Jan 2006 15:04:05 WIB"),
				duration.Round(time.Second).String())

			select {
			case <-ctx.Done():
				log.Println("[VisitPlanScheduler] 🛑 Dihentikan.")
				return
			case <-time.After(duration):
				if err := s.run(); err != nil {
					log.Printf("[VisitPlanScheduler] ❌ Error: %v", err)
				}
			}
		}
	}()
}

func (s *VisitPlanScheduler) run() error {
	log.Println("[VisitPlanScheduler] 🚀 Memulai pengecekan rencana kunjungan hari ini...")

	now := time.Now().In(s.tz)
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, s.tz)
	endOfDay := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, s.tz)

	var plans []domain.SalesVisitPlan
	if err := s.db.Where("plan_date >= ? AND plan_date <= ?", startOfDay, endOfDay).Find(&plans).Error; err != nil {
		return fmt.Errorf("gagal query visit plans: %w", err)
	}

	// Group by employee
	employeePlans := make(map[string]int)
	for _, p := range plans {
		employeePlans[p.EmployeeID.String()]++
	}

	for empID, count := range employeePlans {
		var emp domain.Employee
		if err := s.db.Where("id = ?", empID).First(&emp).Error; err != nil {
			log.Printf("[VisitPlanScheduler] ❌ Gagal mencari employee %s: %v", empID, err)
			continue
		}

		title := "Jadwal Kunjungan Hari Ini"
		body := fmt.Sprintf("Ada %d toko yang direncanakan untuk dikunjungi hari ini. Semangat bekerja!", count)

		if err := s.fcmClient.SendPushNotification(emp.UserID, title, body); err != nil {
			log.Printf("[VisitPlanScheduler] ❌ Gagal kirim notifikasi ke user %s: %v", emp.UserID, err)
		} else {
			log.Printf("[VisitPlanScheduler] ✅ Terkirim ke user %s (%d toko)", emp.UserID, count)
		}
	}

	return nil
}
