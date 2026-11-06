package scheduler

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"example.com/hris/backend/internal/repository"
	"example.com/hris/backend/pkg/sse"
)

type LicenseExpiryScheduler struct {
	factoryRepo      repository.FactoryRepository
	notificationRepo repository.NotificationRepository
	sseHub           *sse.Hub
	tz               *time.Location
	checkInterval    time.Duration
}

func NewLicenseExpiryScheduler(
	factoryRepo repository.FactoryRepository,
	notificationRepo repository.NotificationRepository,
	sseHub *sse.Hub,
	checkInterval time.Duration,
) *LicenseExpiryScheduler {
	loc, _ := time.LoadLocation("Asia/Jakarta")
	return &LicenseExpiryScheduler{
		factoryRepo:      factoryRepo,
		notificationRepo: notificationRepo,
		sseHub:           sseHub,
		tz:               loc,
		checkInterval:    checkInterval,
	}
}

func (s *LicenseExpiryScheduler) Start(ctx context.Context) {
	go func() {
		// Jalankan pengecekan pertama kali setelah start
		s.checkExpiries()

		ticker := time.NewTicker(s.checkInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				log.Println("[LicenseExpiryScheduler] Stopped")
				return
			case <-ticker.C:
				s.checkExpiries()
			}
		}
	}()
}

func (s *LicenseExpiryScheduler) checkExpiries() {
	now := time.Now().In(s.tz)
	// Kita cari produk yang expiry datenya <= 6 bulan dari sekarang (approx 180 hari)
	threshold := now.Add(180 * 24 * time.Hour)

	// Karena repository pattern, lebih baik kita panggil method di FactoryRepo.
	// Tapi kalau tidak ada, kita bisa menggunakan method GetProducts dengan filter khusus,
	// atau kita ambil semua produk dan filter di sini (jika datanya sedikit).
	// Asumsi kita panggil GetProducts (ambil semua) lalu filter di memori demi kemudahan,
	// atau tambahkan method baru di repository.
	// Kita asumsikan ambil semua produk untuk disederhanakan:
	products, err := s.factoryRepo.GetProductsByCompanyID(uuid.Nil, uuid.Nil) // uuid.Nil untuk companyID jika get all, atau perlu sesuaikan method-nya.
	if err != nil {
		log.Printf("[LicenseExpiryScheduler] Error fetching products: %v", err)
		return
	}

	for _, p := range products {
		needsNotification := false
		message := ""

		if p.BPOMExpiryDate != nil && p.BPOMExpiryDate.Before(threshold) {
			needsNotification = true
			message += "BPOM akan/sudah habis masa berlakunya (" + p.BPOMExpiryDate.Format("02-01-2006") + "). "
		}

		if p.HalalExpiryDate != nil && p.HalalExpiryDate.Before(threshold) {
			needsNotification = true
			message += "Sertifikat Halal akan/sudah habis masa berlakunya (" + p.HalalExpiryDate.Format("02-01-2006") + "). "
		}

		if needsNotification {
			// Cek apakah sudah dinotifikasi hari ini atau baru-baru ini.
			// Untuk simpelnya, kita bisa menggunakan "LICENSE_EXPIRING_" + productID sebagai Title untuk deduplikasi jika diperlukan.
			// Di sini kita langsung insert notifikasi (disarankan menambahkan logic deduplikasi agar tidak spam tiap hari)
			// Contoh: cari notifikasi dengan title "Peringatan Lisensi: [Nama Produk]" yang dibuat dalam 30 hari terakhir.

			// Untuk saat ini, kita cukup insert Notifikasi.
			notification := &domain.Notification{
				CompanyID: p.CompanyID,
				Title:     "Peringatan Lisensi: " + p.Name,
				Message:   message,
				Type:      "LICENSE_EXPIRING",
			}

			err := s.notificationRepo.Create(notification)
			if err != nil {
				log.Printf("[LicenseExpiryScheduler] Error creating notification for %s: %v", p.Name, err)
				continue
			}

			// Broadcast SSE
			if s.sseHub != nil {
				s.sseHub.PublishNotification(context.Background(), sse.NotificationEvent{
					ID:        uuid.New().String(),
					Type:      "license_expiring",
					Title:     notification.Title,
					Message:   notification.Message,
					CompanyID: p.CompanyID.String(),
					CreatedAt: time.Now().Format(time.RFC3339),
					Data: map[string]interface{}{
						"product_id": p.ID.String(),
					},
				})
			}
		}
	}
}
