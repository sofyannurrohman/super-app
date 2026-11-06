package sync

import (
	"log"
	"strings"
	gosync "sync"
	"time"

	"example.com/hris/backend/internal/domain"
	"gorm.io/gorm"
)

var (
	bannerStatusSyncMutex gosync.Mutex
)

// BannerStatusSyncWorker menyinkronkan perubahan status banner dari MySQL Branding
// ke PostgreSQL Super App secara ringan (hanya field status).
// Ini adalah worker yang berjalan lebih sering dari BrandingSyncWorker.
// PostgreSQL Super App bertindak sebagai SOURCE OF TRUTH untuk masa transisi.
type BannerStatusSyncWorker struct {
	superDB    *gorm.DB
	brandingDB *gorm.DB
	interval   time.Duration
}

// NewBannerStatusSyncWorker membuat instance sync worker status banner.
func NewBannerStatusSyncWorker(superDB, brandingDB *gorm.DB, interval time.Duration) *BannerStatusSyncWorker {
	return &BannerStatusSyncWorker{
		superDB:    superDB,
		brandingDB: brandingDB,
		interval:   interval,
	}
}

// Start menjalankan cronjob sync status di goroutine terpisah (non-blocking).
func (w *BannerStatusSyncWorker) Start() {
	if w.brandingDB == nil {
		log.Println("[BannerStatusSync] MySQL Branding tidak tersambung — Banner status sync tidak dijalankan.")
		return
	}
	log.Printf("[BannerStatusSync] ✅ Cronjob dimulai, interval: %s", w.interval)
	go func() {
		// Langsung jalankan satu kali saat startup
		SyncBannerStatusesFromMySQL(w.superDB, w.brandingDB)
		ticker := time.NewTicker(w.interval)
		defer ticker.Stop()
		for range ticker.C {
			SyncBannerStatusesFromMySQL(w.superDB, w.brandingDB)
		}
	}()
}

// SyncBannerStatusesFromMySQL — worker ringan yang hanya menyinkronkan perubahan
// status banner dari MySQL Branding ke PostgreSQL Super App.
//
// Aturan Source of Truth:
//   - PostgreSQL Super App adalah source of truth.
//   - Status yang sudah diatur oleh operator Super App (diterima_operator, installed)
//     TIDAK akan ditimpa oleh MySQL Branding — karena MySQL Branding tidak tahu
//     tentang proses operasional lapangan.
//   - Status di fase awal (request → editing → validation → ready_for_operator)
//     diikuti dari MySQL Branding jika MySQL punya versi lebih baru.
func SyncBannerStatusesFromMySQL(localDB *gorm.DB, mysqlDB *gorm.DB) {
	if mysqlDB == nil {
		return
	}

	// Gunakan mutex agar tidak ada concurrent sync yang tumpang tindih
	bannerStatusSyncMutex.Lock()
	defer bannerStatusSyncMutex.Unlock()

	// Hanya ambil banner yang diupdate dalam 15 menit terakhir (sedikit overlap
	// dari interval 5 menit untuk menghindari missed updates)
	since := time.Now().Add(-15 * time.Minute)

	type MySQLStatusRow struct {
		ID             int64      `gorm:"column:id"`
		Status         string     `gorm:"column:status"`
		UpdatedAt      time.Time  `gorm:"column:updated_at"`
		InstalledAt    *time.Time `gorm:"column:installed_at"`
		PhotoInstalled *string    `gorm:"column:photo_installed"`
		EditorName     *string    `gorm:"column:editor_name"`
	}

	var rows []MySQLStatusRow
	err := mysqlDB.Table("banners").
		Select("banners.id, banners.status, banners.updated_at, banners.installed_at, banners.photo_installed, editors.name as editor_name").
		Joins("LEFT JOIN users editors ON editors.id = banners.editor_id").
		Where("banners.updated_at > ?", since).
		Find(&rows).Error

	if err != nil {
		log.Printf("[BannerStatusSync] ❌ Gagal membaca status terbaru dari MySQL: %v", err)
		return
	}

	if len(rows) == 0 {
		return
	}

	log.Printf("[BannerStatusSync] 🔄 Memproses %d perubahan status dari MySQL Branding...", len(rows))

	updated := 0
	skipped := 0

	for _, row := range rows {
		var existing domain.BannerOrder
		if err := localDB.Where("mysql_banner_id = ?", row.ID).First(&existing).Error; err != nil {
			// Banner belum diimport ke PG — tidak perlu diproses di sini,
			// akan dihandle oleh SyncBannersFromMySQL (full sync)
			continue
		}

		pgStatus := strings.ToLower(existing.Status)

		// STATUS TERMINAL — diatur oleh Super App, jangan ditimpa MySQL
		// Ini adalah status yang hanya bisa terjadi setelah operator lapangan bekerja
		if isSuperAppTerminalStatus(pgStatus) {
			skipped++
			continue
		}

		// Jika mysql_updated_at di PG sudah sama dengan UpdatedAt MySQL, skip
		// (data sudah sinkron, tidak ada perubahan baru)
		if existing.MySQLUpdatedAt != nil && !existing.MySQLUpdatedAt.Before(row.UpdatedAt) {
			skipped++
			continue
		}

		// Build update payload
		updates := map[string]interface{}{
			"mysql_updated_at": row.UpdatedAt,
		}

		// Jangan menimpa status 'editing' di Super App jika MySQL belum sinkron, kecuali MySQL sudah berstatus installed
		rowStatusLower := strings.ToLower(strings.TrimSpace(row.Status))
		if (pgStatus != "editing" && pgStatus != "in_design") || rowStatusLower == "installed" || rowStatusLower == "terpasang" {
			updates["status"] = row.Status
		}

		// Ikut-update photo_installed jika MySQL punya data (foto dari Branding App)
		if row.PhotoInstalled != nil && *row.PhotoInstalled != "" {
			photoUrl := *row.PhotoInstalled
			photoUrl = externalBrandingPhotoURL(photoUrl)
			updates["photo_installed"] = photoUrl
		}

		// Ikut-update installed_at jika MySQL punya data
		if row.InstalledAt != nil {
			updates["installed_at"] = row.InstalledAt
		}

		// Ikut-update mysql_editor_name jika ada data baru
		if row.EditorName != nil && *row.EditorName != "" {
			updates["mysql_editor_name"] = row.EditorName
		}

		if err := localDB.Model(&existing).Updates(updates).Error; err != nil {
			log.Printf("[BannerStatusSync] ⚠️ Gagal update status banner (mysql_id=%d): %v", row.ID, err)
		} else {
			updated++
		}
	}

	if updated > 0 || skipped > 0 {
		log.Printf("[BannerStatusSync] ✅ Status sync selesai — diupdate: %d, dilewati (terminal/tidak berubah): %d", updated, skipped)
	}
}

// isSuperAppTerminalStatus mengembalikan true jika status ini sudah diatur
// oleh operator Super App dan tidak boleh ditimpa oleh data dari MySQL Branding.
// Status ini menandakan banner sudah berada di tangan operator lapangan.
func isSuperAppTerminalStatus(pgStatus string) bool {
	switch pgStatus {
	case "diterima_operator", // Operator sudah ambil dari gudang
		"installed",        // Sudah dipasang (via Super App)
		"terpasang",        // Alias installed
		"selesai_dipasang": // Alias installed
		return true
	}
	return false
}
