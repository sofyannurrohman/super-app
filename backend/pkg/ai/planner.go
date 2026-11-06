package ai

import (
	"encoding/json"
	"fmt"
	"strings"
)

type Plan struct {
	Intent       string         `json:"intent"`
	NeedsClarify bool           `json:"needs_clarification"`
	Question     string         `json:"clarification_question,omitempty"`
	Tool         string         `json:"tool,omitempty"`
	Arguments    map[string]any `json:"arguments,omitempty"`
}

type ChatMessage struct {
	Role    string `json:"role"` // "user" | "assistant"
	Content string `json:"content"`
}

// PlannerSystemPrompt membatasi model hanya memilih tool terdaftar atau "none" untuk general/knowledge base.
func PlannerSystemPrompt() string {
	data, _ := json.Marshal(OperationalTools())
	return `Anda adalah planner AI operasional Super App & HRIS Karsa HR. Jangan menjawab pertanyaan langsung ke pengguna.
Tugas Anda:
1. Pahami maksud pengguna dan tentukan apakah butuh query data real-time database dari tool operasional atau tidak.
2. Jika pertanyaan terkait pengetahuan umum, matematika, coding, motivasi kerja, bisnis umum, sapaan/basa-basi, panduan pitching sales, SOP input data, atau troubleshooting aplikasi, gunakan tool: "none".
3. PENTING DI LINGKUNGAN KERJA PERUSAHAAN:
   - Semua pertanyaan yang menanyakan nama orang, identitas seseorang, atau profil seseorang (contoh: "siapa itu Sofyan", "siapa Sofyan", "Sofyan siapa", "siapa sih Budi", "kenal dengan Budi?", "siapa karyawan Sofyan", "cari kontak Pak Joko", "profil staf X") ADALAH kueri pencarian personil/karyawan/akun manajemen internal perusahaan, BUKAN tokoh publik luar. WAJIB menggunakan tool "search_employee" dengan arguments: {"keyword": "[nama orang yang dicari]"}.
   - Kueri presensi/kehadiran harian (contoh: "Siapa saja yang hadir hari ini?", "Daftar kehadiran hari ini", "Siapa yang terlambat hari ini?") WAJIB menggunakan tool "get_employee_attendance".
   - Kueri produk perusahaan, lini produk pabrik, katalog barang jadi, maupun keunggulan produk (contoh: "Produk perusahaan & keunggulannya", "Daftar produk pabrik", "Apa saja produk yang diproduksi?") WAJIB menggunakan tool "get_products_catalog" dengan arguments: {}.
   - Kueri antrean approval / reimbursement / cuti pending WAJIB menggunakan tool "get_pending_approvals".
   - Kueri daftar cabang perusahaan, lokasi gudang, atau divisi departemen (contoh: "daftar cabang perusahaan", "kantor perusahaan ada di mana saja", "ada departemen apa saja") WAJIB menggunakan tool "get_departments_and_branches".
   - Kueri pengiriman / Surat Jalan / Delivery Order (contoh: "status pengiriman hari ini", "cek surat jalan", "status DO driver") WAJIB menggunakan tool "get_delivery_orders".
   - Kueri pesanan penjualan / Sales Order / PO (contoh: "daftar order sales", "cek pesanan toko X") WAJIB menggunakan tool "get_sales_orders".
   - Kueri kendaraan / armada truk / mobil kanvas (contoh: "daftar mobil kanvas", "cek plat kendaraan operasional") WAJIB menggunakan tool "get_vehicles_status".
   - Kueri stok fisik di gudang barang jadi WAJIB menggunakan tool "get_inventory_stock".
   - Kueri stok bahan baku pabrik (contoh: "stok kedelai", "stok botol/kardus di pabrik") WAJIB menggunakan tool "get_raw_material_stock".
   - Kueri kas operasional pabrik / kas kecil WAJIB menggunakan tool "get_kas_pabrik_summary".
   - Kueri ringkasan penjualan / omzet (contoh: "Ringkasan penjualan bulan ini", "penjualan cabang solo bulan ini", "penjualan modern trade vs kanvas", "omzet hari ini", "performa sales per cabang") WAJIB menggunakan tool "get_sales_summary" dengan arguments: {"branch_name": "[nama cabang jika ada]", "sales_type": "[CANVAS/MOTORIS/MODERN_TRADE/ALL jika ada]", "period": "[hari_ini/bulan_ini/bulan_lalu/tahun_ini jika ada]"}.
   - Kueri marketplace / toko online e-commerce (contoh: "Ringkasan marketplace", "omzet tiktok shop bulan ini", "penjualan shopee cabang solo", "marketplace per cabang", "produk terlaris marketplace") WAJIB menggunakan tool "get_marketplace_summary" dengan arguments: {"branch_name": "[nama cabang/gudang jika ada]", "marketplace": "[tiktok/shopee/tokopedia/all jika ada]", "period": "[hari_ini/bulan_ini/bulan_lalu/tahun_ini jika ada]"}.
   - Kueri riwayat penjualan per toko/pelanggan WAJIB menggunakan "get_customer_sales_summary".
   - Kueri cari toko/pelanggan WAJIB menggunakan "search_customer".
4. Jika entity penting masih kurang/ambigu pada kueri data operasional, set needs_clarification=true dan isi clarification_question.
5. Keluarkan HANYA JSON valid tanpa markdown dengan bentuk:
{"intent":"...","needs_clarification":false,"tool":"none","arguments":{}}
Jangan membuat UUID, SQL, atau nama tool baru di luar daftar berikut.
Tool yang tersedia: ` + string(data)
}

// BuildPlannerPrompt menyusun prompt untuk planner dengan riwayat percakapan sebelumnya.
func BuildPlannerPrompt(history []ChatMessage, currentMessage string) string {
	var sb strings.Builder
	sb.WriteString(PlannerSystemPrompt())
	if len(history) > 0 {
		sb.WriteString("\n\n=== RIWAYAT PERCAKAPAN SEBELUMNYA ===\n")
		for _, h := range history {
			role := "User"
			if h.Role == "assistant" {
				role = "AI Assistant"
			}
			content := strings.TrimSpace(h.Content)
			if content != "" {
				sb.WriteString(fmt.Sprintf("%s: %s\n", role, content))
			}
		}
		sb.WriteString("=====================================\n")
		sb.WriteString("Gunakan riwayat percakapan di atas untuk memahami konteks jika pesan terbaru pengguna adalah jawaban atas pertanyaan klarifikasi AI sebelumnya.\n")
	}
	sb.WriteString("\nPertanyaan/Pesan Terbaru Pengguna: " + strings.TrimSpace(currentMessage))
	return sb.String()
}

func ParsePlan(raw string) (Plan, error) {
	raw = strings.TrimSpace(raw)
	raw = strings.TrimPrefix(raw, "```json")
	raw = strings.TrimPrefix(raw, "```")
	raw = strings.TrimSuffix(raw, "```")
	raw = strings.TrimSpace(raw)

	// Extract JSON object if surrounded by extra text
	if start := strings.Index(raw, "{"); start != -1 {
		if end := strings.LastIndex(raw, "}"); end != -1 && end > start {
			raw = raw[start : end+1]
		}
	}

	var plan Plan
	if err := json.Unmarshal([]byte(raw), &plan); err != nil {
		return Plan{}, fmt.Errorf("planner menghasilkan JSON tidak valid: %w", err)
	}
	if strings.TrimSpace(plan.Intent) == "" {
		return Plan{}, fmt.Errorf("intent planner kosong")
	}
	if plan.NeedsClarify {
		if strings.TrimSpace(plan.Question) == "" {
			return Plan{}, fmt.Errorf("pertanyaan klarifikasi kosong")
		}
		return plan, nil
	}
	if plan.Tool == "" || strings.EqualFold(plan.Tool, "none") || strings.EqualFold(plan.Tool, "null") {
		plan.Tool = "none"
		return plan, nil
	}
	if err := ValidateToolCall(plan.Tool, plan.Arguments); err != nil {
		return Plan{}, err
	}
	return plan, nil
}

// CleanPersonSearchKeyword membersihkan kata sambung, sapaan, dan tanda baca dari query nama/orang
func CleanPersonSearchKeyword(raw string) string {
	s := strings.TrimSpace(raw)
	s = strings.Map(func(r rune) rune {
		if strings.ContainsRune("?!.,\"'()[]{}:;-_`", r) {
			return ' '
		}
		return r
	}, s)
	s = strings.Join(strings.Fields(s), " ")

	stopWords := map[string]bool{
		"siapa": true, "itu": true, "si": true, "sih": true, "adalah": true,
		"karyawan": true, "karyawannya": true, "staf": true, "staff": true, "pegawai": true,
		"pak": true, "bapak": true, "bu": true, "ibu": true, "mas": true, "mba": true, "mbak": true,
		"profil": true, "biodata": true, "kontak": true, "tentang": true, "kenal": true, "dengan": true,
		"sama": true, "apakah": true, "ada": true, "tau": true, "tahu": true, "tolong": true, "cari": true,
		"carikan": true, "info": true, "informasi": true, "data": true, "orang": true, "nama": true,
		"nomor": true, "hp": true, "telepon": true, "dong": true, "ya": true, "kah": true,
	}

	words := strings.Fields(s)
	var filtered []string
	for _, w := range words {
		if !stopWords[strings.ToLower(w)] {
			filtered = append(filtered, w)
		}
	}

	if len(filtered) > 0 {
		return strings.Join(filtered, " ")
	}
	return strings.TrimSpace(raw)
}

// FastMatchPlan melakukan pencocokan intent deterministik berlatensi 0ms untuk kueri-kueri operasional umum
func FastMatchPlan(message string) (Plan, bool) {
	msg := strings.TrimSpace(message)
	if msg == "" {
		return Plan{}, false
	}
	lower := strings.ToLower(msg)

	// 1. Presensi / Kehadiran
	if (strings.Contains(lower, "siapa") || strings.Contains(lower, "daftar") || strings.Contains(lower, "cek") || strings.Contains(lower, "rekap")) &&
		(strings.Contains(lower, "hadir") || strings.Contains(lower, "kehadiran") || strings.Contains(lower, "presensi") || strings.Contains(lower, "absen") || strings.Contains(lower, "terlambat") || strings.Contains(lower, "telat") || strings.Contains(lower, "cuti")) {
		return Plan{
			Intent:    "attendance_check",
			Tool:      "get_employee_attendance",
			Arguments: map[string]any{},
		}, true
	}

	// 2. Produk Pabrik & Keunggulannya
	if strings.Contains(lower, "katalog produk") || strings.Contains(lower, "daftar produk") || strings.Contains(lower, "apa saja produk") || strings.Contains(lower, "lini produk") {
		return Plan{
			Intent:    "product_catalog",
			Tool:      "get_products_catalog",
			Arguments: map[string]any{},
		}, true
	}

	// 3. Antrean Approval
	if strings.Contains(lower, "approval") || strings.Contains(lower, "antrean pending") || strings.Contains(lower, "pengajuan pending") {
		return Plan{
			Intent:    "pending_approvals",
			Tool:      "get_pending_approvals",
			Arguments: map[string]any{},
		}, true
	}

	// 4. Cabang & Departemen
	if strings.Contains(lower, "daftar cabang") || strings.Contains(lower, "lokasi cabang") || strings.Contains(lower, "cabang perusahaan") || strings.Contains(lower, "daftar departemen") || strings.Contains(lower, "daftar divisi") {
		return Plan{
			Intent:    "departments_and_branches",
			Tool:      "get_departments_and_branches",
			Arguments: map[string]any{},
		}, true
	}

	// 5. Kas Operasional Pabrik
	if strings.Contains(lower, "kas pabrik") || strings.Contains(lower, "kas operasional") || strings.Contains(lower, "saldo kas") {
		return Plan{
			Intent:    "kas_pabrik",
			Tool:      "get_kas_pabrik_summary",
			Arguments: map[string]any{},
		}, true
	}

	// 6. Status Pengiriman / Surat Jalan / DO
	if strings.Contains(lower, "status pengiriman") || strings.Contains(lower, "surat jalan") || strings.Contains(lower, "status do") || strings.Contains(lower, "delivery order") {
		return Plan{
			Intent:    "delivery_orders",
			Tool:      "get_delivery_orders",
			Arguments: map[string]any{},
		}, true
	}

	// 7. Armada / Kendaraan
	if strings.Contains(lower, "daftar kendaraan") || strings.Contains(lower, "mobil kanvas") || strings.Contains(lower, "armada truk") || strings.Contains(lower, "plat kendaraan") {
		return Plan{
			Intent:    "vehicles_status",
			Tool:      "get_vehicles_status",
			Arguments: map[string]any{},
		}, true
	}

	// 8. Pencarian Karyawan / Personil
	isPersonQuery := false
	if strings.HasPrefix(lower, "siapa ") || strings.HasSuffix(lower, " siapa") || strings.Contains(lower, " itu siapa") ||
		strings.HasPrefix(lower, "kenal ") || strings.HasPrefix(lower, "kenal dengan ") || strings.HasPrefix(lower, "kenalkah ") ||
		strings.HasPrefix(lower, "cari karyawan ") || strings.HasPrefix(lower, "cari staf ") || strings.HasPrefix(lower, "cari staff ") ||
		strings.HasPrefix(lower, "profil ") || strings.HasPrefix(lower, "biodata ") || strings.HasPrefix(lower, "kontak ") {
		isPersonQuery = true
	}

	if isPersonQuery {
		cleanKeyword := CleanPersonSearchKeyword(msg)
		if cleanKeyword != "" && len(cleanKeyword) >= 2 {
			return Plan{
				Intent: "employee_search",
				Tool:   "search_employee",
				Arguments: map[string]any{
					"keyword": cleanKeyword,
				},
			}, true
		}
	}

	return Plan{}, false
}
