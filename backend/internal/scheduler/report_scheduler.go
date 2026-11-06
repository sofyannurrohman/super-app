package scheduler

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"example.com/hris/backend/internal/usecase"
	"example.com/hris/backend/pkg/whatsapp"
)

// ReportScheduler menjalankan background goroutine yang:
//   - Menghitung waktu mundur ke 17:30 WIB hari ini (atau besok jika sudah lewat)
//   - Memanggil AggregateDaily → GenerateAIReport → SendToGroup
//   - Loop setiap hari tanpa perlu restart server
type ReportScheduler struct {
	reportUC *usecase.ReportUsecase
	waClient *whatsapp.Client
	groupID  string
	tz       *time.Location
}

// NewReportScheduler membuat scheduler baru.
// Group ID diambil dari env var WA_REPORT_GROUP_ID.
func NewReportScheduler(reportUC *usecase.ReportUsecase, waClient *whatsapp.Client) *ReportScheduler {
	loc, _ := time.LoadLocation("Asia/Jakarta")
	return &ReportScheduler{
		reportUC: reportUC,
		waClient: waClient,
		groupID:  os.Getenv("WA_REPORT_GROUP_ID"),
		tz:       loc,
	}
}

// Start menjalankan scheduler di background goroutine.
// Aman dipanggil sekali saat server startup.
func (s *ReportScheduler) Start(ctx context.Context) {
	go func() {
		for {
			now := time.Now().In(s.tz)

			// Target jam: 17:30:00 WIB hari ini
			next := time.Date(now.Year(), now.Month(), now.Day(), 17, 30, 0, 0, s.tz)
			if !now.Before(next) {
				// Sudah lewat jam 17:30 hari ini → jadwalkan besok
				next = next.Add(24 * time.Hour)
			}

			duration := next.Sub(now)
			log.Printf("[ReportScheduler] ⏰ Laporan berikutnya: %s (dalam %s)",
				next.Format("02 Jan 2006 15:04:05 WIB"),
				duration.Round(time.Second).String())

			select {
			case <-ctx.Done():
				log.Println("[ReportScheduler] 🛑 Dihentikan.")
				return
			case <-time.After(duration):
				if err := s.run(ctx); err != nil {
					log.Printf("[ReportScheduler] ❌ Error: %v", err)
				}
			}
		}
	}()
}

// TriggerNow memicu pengiriman laporan secara manual (untuk testing).
// Dapat dipanggil dari HTTP handler.
func (s *ReportScheduler) TriggerNow(ctx context.Context) error {
	return s.run(ctx)
}

// run adalah fungsi inti: kumpul data → buat narasi → kirim WA.
func (s *ReportScheduler) run(ctx context.Context) error {
	log.Println("[ReportScheduler] 🚀 Memulai pembuatan laporan harian...")

	// 1. Kumpulkan data dari DB
	data, err := s.reportUC.AggregateDaily()
	if err != nil {
		return fmt.Errorf("gagal agregasi data: %w", err)
	}
	log.Printf("[ReportScheduler] 📊 Data terkumpul: %d karyawan, Rp %s penjualan",
		data.TotalEmployees, formatRupiah(data.TotalSalesToday))

	// 2. Buat narasi via Gemini (dengan fallback)
	var message string
	aiText, err := s.reportUC.GenerateAIReport(data)
	if err != nil {
		log.Printf("[ReportScheduler] ⚠️  Gemini gagal (%v), menggunakan fallback", err)
		message = s.reportUC.BuildFallbackReport(data)
	} else {
		message = aiText
		log.Println("[ReportScheduler] 🤖 Narasi AI berhasil dibuat")
	}

	// 3. Kirim ke WhatsApp group (atau print ke log jika Group ID belum diset)
	if s.groupID == "" {
		log.Println("[ReportScheduler] ℹ️  WA_REPORT_GROUP_ID belum diset — preview laporan:")
		log.Println("─────────────────────────────────────────")
		fmt.Println(message)
		log.Println("─────────────────────────────────────────")
		return nil
	}

	if err := s.waClient.SendToGroup(s.groupID, message); err != nil {
		return fmt.Errorf("gagal kirim ke WA group %s: %w", s.groupID, err)
	}

	log.Printf("[ReportScheduler] ✅ Laporan berhasil dikirim ke group: %s", s.groupID)
	return nil
}

func formatRupiah(amount float64) string {
	s := fmt.Sprintf("%.0f", amount)
	n := len(s)
	if n <= 3 {
		return s
	}
	var result []byte
	for i, c := range s {
		if i > 0 && (n-i)%3 == 0 {
			result = append(result, '.')
		}
		result = append(result, byte(c))
	}
	return string(result)
}
