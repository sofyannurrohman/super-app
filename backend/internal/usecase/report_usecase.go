package usecase

import (
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"example.com/hris/backend/internal/domain"
	geminiPkg "example.com/hris/backend/pkg/gemini"
)

// ReportUsecase bertanggung jawab untuk:
// 1. Mengumpulkan data agregat dari PostgreSQL (AggregateDaily)
// 2. Membuat narasi laporan via Gemini AI (GenerateAIReport)
// 3. Menyediakan fallback jika Gemini tidak tersedia (BuildFallbackReport)
type ReportUsecase struct {
	db           *gorm.DB
	geminiClient *geminiPkg.Client
}

func NewReportUsecase(db *gorm.DB, geminiClient *geminiPkg.Client) *ReportUsecase {
	return &ReportUsecase{db: db, geminiClient: geminiClient}
}

// AggregateDaily mengumpulkan semua data dari DB untuk hari ini (WIB).
func (r *ReportUsecase) AggregateDaily() (*domain.DailyReportData, error) {
	loc, _ := time.LoadLocation("Asia/Jakarta")
	now := time.Now().In(loc)
	todayStr := now.Format("2006-01-02")

	data := &domain.DailyReportData{
		Date:         now,
		ReportPeriod: now.Format("Monday, 02 January 2006"),
	}

	// ── Karyawan Aktif ─────────────────────────────────────────────────────────
	r.db.Raw(`SELECT COUNT(*) FROM employees WHERE status = 'active'`).Scan(&data.TotalEmployees)

	// ── Absensi Hari Ini ───────────────────────────────────────────────────────
	r.db.Raw(`
		SELECT COUNT(*) FROM attendances
		WHERE DATE(check_in AT TIME ZONE 'Asia/Jakarta') = ?
		  AND status IN ('present', 'on_time', 'ON_TIME')
	`, todayStr).Scan(&data.PresentToday)

	r.db.Raw(`
		SELECT COUNT(*) FROM attendances
		WHERE DATE(check_in AT TIME ZONE 'Asia/Jakarta') = ?
		  AND status IN ('late', 'LATE')
	`, todayStr).Scan(&data.LateToday)

	// ── Cuti Aktif Hari Ini ────────────────────────────────────────────────────
	r.db.Raw(`
		SELECT COUNT(*) FROM leaves
		WHERE status = 'approved'
		  AND ? BETWEEN DATE(start_date) AND DATE(end_date)
	`, todayStr).Scan(&data.OnLeaveToday)

	// ── Pending Approvals ──────────────────────────────────────────────────────
	r.db.Raw(`SELECT COUNT(*) FROM leaves WHERE status = 'pending'`).Scan(&data.PendingLeaveReqs)
	r.db.Raw(`SELECT COUNT(*) FROM overtimes WHERE status = 'pending'`).Scan(&data.PendingOvertimeReqs)
	r.db.Raw(`SELECT COUNT(*) FROM reimbursements WHERE status = 'pending'`).Scan(&data.PendingReimbursement)

	// ── Sales Hari Ini ─────────────────────────────────────────────────────────
	r.db.Raw(`
		SELECT COALESCE(SUM(total_amount), 0) FROM sales_transactions
		WHERE DATE(created_at AT TIME ZONE 'Asia/Jakarta') = ?
	`, todayStr).Scan(&data.TotalSalesToday)

	r.db.Raw(`
		SELECT COUNT(*) FROM sales_transactions
		WHERE DATE(created_at AT TIME ZONE 'Asia/Jakarta') = ?
	`, todayStr).Scan(&data.TotalTransactions)

	// ── Top Salesman Hari Ini ──────────────────────────────────────────────────
	row := r.db.Raw(`
		SELECT
			CONCAT(e.first_name, ' ', e.last_name) AS full_name,
			COALESCE(SUM(st.total_amount), 0)       AS total
		FROM sales_transactions st
		JOIN employees e ON e.id = st.employee_id
		WHERE DATE(st.created_at AT TIME ZONE 'Asia/Jakarta') = ?
		GROUP BY e.first_name, e.last_name
		ORDER BY total DESC
		LIMIT 1
	`, todayStr).Row()

	if row != nil {
		_ = row.Scan(&data.TopSalesmanName, &data.TopSalesmanAmount)
	}

	// ── Delivery ───────────────────────────────────────────────────────────────
	r.db.Raw(`
		SELECT COUNT(*) FROM delivery_batches
		WHERE DATE(created_at AT TIME ZONE 'Asia/Jakarta') = ?
		  AND status = 'completed'
	`, todayStr).Scan(&data.CompletedDeliveries)

	r.db.Raw(`
		SELECT COUNT(*) FROM delivery_batches
		WHERE DATE(created_at AT TIME ZONE 'Asia/Jakarta') = ?
	`, todayStr).Scan(&data.TotalDeliveries)

	// ── Pending Sales Orders ───────────────────────────────────────────────────
	r.db.Raw(`
		SELECT COUNT(*) FROM sales_orders
		WHERE status IN ('pending', 'confirmed')
		  AND DATE(created_at AT TIME ZONE 'Asia/Jakarta') = ?
	`, todayStr).Scan(&data.PendingOrders)

	// Hitung karyawan tidak hadir (estimasi)
	data.AbsentToday = data.TotalEmployees - data.PresentToday - data.LateToday - data.OnLeaveToday
	if data.AbsentToday < 0 {
		data.AbsentToday = 0
	}

	return data, nil
}

// GenerateAIReport membuat narasi laporan menggunakan Gemini Flash.
func (r *ReportUsecase) GenerateAIReport(data *domain.DailyReportData) (string, error) {
	attendanceRate := 0.0
	if data.TotalEmployees > 0 {
		attendanceRate = float64(data.PresentToday+data.LateToday) / float64(data.TotalEmployees) * 100
	}

	totalPending := data.PendingLeaveReqs + data.PendingOvertimeReqs + data.PendingReimbursement
	pendingWarning := ""
	if totalPending > 3 {
		pendingWarning = fmt.Sprintf("PERHATIAN: Ada %d total approval yang menunggu, mohon segera diproses.", totalPending)
	}

	prompt := fmt.Sprintf(`Tugasmu adalah membuat Laporan Harian lengkap untuk grup WhatsApp Managerial Karsa HR.
Berikan HANYA teks laporan lengkap (jangan hanya judulnya) berdasarkan data di bawah ini.

DATA HARI INI (%s):

KEHADIRAN:
- Total Karyawan Aktif: %d orang
- Hadir Tepat Waktu: %d orang
- Terlambat: %d orang
- Tingkat Kehadiran: %.1f%%
- Sedang Cuti/Izin: %d orang
- Tidak Hadir: %d orang

PENDING APPROVALS:
- Pengajuan Cuti: %d
- Pengajuan Lembur: %d
- Reimbursement: %d
%s

PENJUALAN:
- Total Penjualan: Rp %s
- Jumlah Transaksi: %d
- Sales Terbaik: %s (Rp %s)
- Order Pending: %d
- Pengiriman Selesai: %d dari %d

INSTRUKSI:
- Gunakan bahasa Indonesia profesional namun hangat.
- Buat laporan UTUH sesuai struktur di bawah ini. Jangan berhenti di tengah jalan!
- Gunakan format *tebal* dan emoji.

STRUKTUR LAPORAN YANG DIHARAPKAN:
📊 *Laporan Harian*
📅 [Tanggal sesuai data] | [Jam saat ini] WIB

[1-2 kalimat ringkasan singkat tentang performa kehadiran dan penjualan hari ini]

👥 *Kehadiran*
• ✅ Hadir: [Jumlah] orang ([Persentase]%%)
• ⏰ Terlambat: [Jumlah] orang
• 🏖️ Cuti/Izin: [Jumlah] orang
• ❌ Tidak Hadir: [Jumlah] orang

⏳ *Pending Approval*
• Cuti: [Jumlah] | Lembur: [Jumlah] | Reimburse: [Jumlah]
[Tampilkan peringatan pending jika ada]

💰 *Penjualan Hari Ini*
• Total: *Rp [Jumlah]*
• Transaksi: *[Jumlah] transaksi*
• 🏆 Terbaik: [Nama]
• 🚚 Pengiriman: [Jumlah]/[Jumlah] selesai

[1 kalimat motivasi singkat untuk tim]

_Laporan otomatis Karsa HR • %s WIB_`,
		data.ReportPeriod,
		data.TotalEmployees,
		data.PresentToday,
		data.LateToday,
		attendanceRate,
		data.OnLeaveToday,
		data.AbsentToday,
		data.PendingLeaveReqs,
		data.PendingOvertimeReqs,
		data.PendingReimbursement,
		pendingWarning,
		formatRupiah(data.TotalSalesToday),
		data.TotalTransactions,
		data.TopSalesmanName, formatRupiah(data.TopSalesmanAmount),
		data.PendingOrders,
		data.CompletedDeliveries, data.TotalDeliveries,
		data.Date.Format("15:04"),
	)

	return r.geminiClient.GenerateText(prompt)
}

// BuildFallbackReport membuat laporan teks sederhana tanpa AI.
// Digunakan ketika Gemini tidak tersedia atau gagal.
func (r *ReportUsecase) BuildFallbackReport(data *domain.DailyReportData) string {
	totalPending := data.PendingLeaveReqs + data.PendingOvertimeReqs + data.PendingReimbursement
	warningLine := ""
	if totalPending > 3 {
		warningLine = fmt.Sprintf("\n⚠️ *%d approval menunggu tindakan!*\n", totalPending)
	}

	topSales := "-"
	if data.TopSalesmanName != "" {
		topSales = fmt.Sprintf("%s (Rp %s)", data.TopSalesmanName, formatRupiah(data.TopSalesmanAmount))
	}

	return strings.TrimSpace(fmt.Sprintf(
		"📊 *Laporan Harian*\n"+
			"📅 %s | %s WIB\n\n"+
			"👥 *Kehadiran* (%d karyawan aktif)\n"+
			"• ✅ Hadir: *%d orang* (%.0f%%)\n"+
			"• ⏰ Terlambat: %d orang\n"+
			"• 🏖️ Cuti/Izin: %d orang\n"+
			"• ❌ Tidak Hadir: %d orang\n"+
			"\n⏳ *Pending Approval*\n"+
			"• Cuti: %d | Lembur: %d | Reimburse: %d\n"+
			"%s\n"+
			"💰 *Penjualan Hari Ini*\n"+
			"• Total: *Rp %s*\n"+
			"• Transaksi: *%d transaksi*\n"+
			"• 🏆 Terbaik: %s\n"+
			"• 🚚 Pengiriman: %d/%d selesai\n\n"+
			"_Laporan otomatis Karsa HR • %s WIB_",
		data.ReportPeriod,
		data.Date.Format("15:04"),
		data.TotalEmployees,
		data.PresentToday,
		func() float64 {
			if data.TotalEmployees == 0 {
				return 0
			}
			return float64(data.PresentToday) / float64(data.TotalEmployees) * 100
		}(),
		data.LateToday,
		data.OnLeaveToday,
		data.AbsentToday,
		data.PendingLeaveReqs, data.PendingOvertimeReqs, data.PendingReimbursement,
		warningLine,
		formatRupiah(data.TotalSalesToday),
		data.TotalTransactions,
		topSales,
		data.CompletedDeliveries, data.TotalDeliveries,
		data.Date.Format("15:04"),
	))
}

// formatRupiah memformat angka float64 ke format Rupiah Indonesia (titik pemisah ribuan).
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
