package ai

import (
	"fmt"
	"strings"
)

// ToolDefinition adalah kontrak yang boleh dipilih planner. Tool tidak pernah
// menerima SQL dari model; backend tetap menentukan implementasi query-nya.
type ToolDefinition struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	ReadOnly    bool              `json:"read_only"`
	Arguments   map[string]string `json:"arguments"`
}

var operationalTools = []ToolDefinition{
	{Name: "search_customer", Description: "Cari toko/outlet/pelanggan eksternal berdasarkan nama toko, pemilik, kota, atau alamat", ReadOnly: true, Arguments: map[string]string{"query": "string"}},
	{Name: "get_customer_sales_summary", Description: "Ambil total penjualan customer pada periode tertentu", ReadOnly: true, Arguments: map[string]string{"customer_id": "uuid", "start_date": "YYYY-MM-DD", "end_date": "YYYY-MM-DD"}},
	{Name: "get_sales_summary", Description: "Ambil ringkasan omzet penjualan, breakdown antar cabang, channel (GT/MT/Direct), top 5 salesman, dan top 5 produk terlaris pada periode tertentu", ReadOnly: true, Arguments: map[string]string{"branch_name": "string?", "branch_id": "uuid?", "sales_type": "string?", "start_date": "YYYY-MM-DD?", "end_date": "YYYY-MM-DD?", "period": "string?"}},
	{Name: "get_inventory_stock", Description: "Cek kuantitas/jumlah stok fisik produk jadi di gudang cabang atau pabrik", ReadOnly: true, Arguments: map[string]string{"product_query": "string?", "warehouse_name": "string?", "branch_id": "uuid?"}},
	{Name: "get_pending_visits", Description: "Cari rencana kunjungan sales yang belum selesai atau sedang berjalan", ReadOnly: true, Arguments: map[string]string{"date": "YYYY-MM-DD?", "branch_id": "uuid?"}},
	{Name: "get_employee_attendance", Description: "Cek siapa saja yang hadir/masuk hari ini, daftar kehadiran/presensi karyawan, jam clock-in, keterlambatan, dan cuti", ReadOnly: true, Arguments: map[string]string{"date": "YYYY-MM-DD?", "branch_name": "string?", "branch_id": "uuid?"}},
	{Name: "get_products_catalog", Description: "Ambil daftar master produk resmi perusahaan dari pabrik/database (nama produk, SKU, kategori, kemasan, harga, spesifikasi)", ReadOnly: true, Arguments: map[string]string{"category": "string?", "query": "string?"}},
	{Name: "search_employee", Description: "Cari data profil karyawan internal (nama lengkap, NIP/NIK, jabatan, divisi/departemen, cabang, nomor telepon, status)", ReadOnly: true, Arguments: map[string]string{"keyword": "string"}},
	{Name: "get_pending_approvals", Description: "Ambil antrean pengajuan yang menunggu persetujuan (cuti, lembur, reimbursement, memo, e-PKWT)", ReadOnly: true, Arguments: map[string]string{"module_type": "string?"}},
	{Name: "get_departments_and_branches", Description: "Ambil daftar cabang operasional perusahaan, lokasi gudang, dan divisi departemen yang ada", ReadOnly: true, Arguments: map[string]string{"type": "string?"}},
	{Name: "get_delivery_orders", Description: "Cek status pengiriman Surat Jalan / Delivery Order (DO), nama driver, nomor plat kendaraan, toko tujuan, dan status pengiriman", ReadOnly: true, Arguments: map[string]string{"status": "string?", "date": "YYYY-MM-DD?", "search": "string?"}},
	{Name: "get_sales_orders", Description: "Cek daftar pesanan penjualan (Sales Order / PO), toko pemesan, salesman, status pembayaran, dan status order", ReadOnly: true, Arguments: map[string]string{"status": "string?", "search": "string?"}},
	{Name: "get_vehicles_status", Description: "Cek daftar armada kendaraan operasional (truk, mobil kanvas, pick-up), nomor plat, tipe kendaraan, dan status", ReadOnly: true, Arguments: map[string]string{"search": "string?"}},
	{Name: "get_raw_material_stock", Description: "Cek stok bahan baku dan packaging di pabrik (kedelai, gula, botol, tutup botol, kardus, stiker label)", ReadOnly: true, Arguments: map[string]string{"query": "string?"}},
	{Name: "get_kas_pabrik_summary", Description: "Cek ringkasan kas operasional pabrik / kas kecil dan laporan mutasi kasir terbaru", ReadOnly: true, Arguments: map[string]string{"date": "YYYY-MM-DD?"}},
	{Name: "get_marketplace_summary", Description: "Cek ringkasan pesanan & omzet marketplace (TikTok Shop, Shopee, Tokopedia), breakdown per cabang/gudang, breakdown per platform, dan top 5 produk terlaris", ReadOnly: true, Arguments: map[string]string{"branch_name": "string?", "branch_id": "uuid?", "marketplace": "string?", "start_date": "YYYY-MM-DD?", "end_date": "YYYY-MM-DD?", "period": "string?"}},
}

func OperationalTools() []ToolDefinition { return append([]ToolDefinition(nil), operationalTools...) }

// ValidateToolCall melakukan validasi awal sebelum permission dan eksekusi.
// Validasi entity (UUID benar-benar ada) tetap dilakukan oleh service tool.
func ValidateToolCall(name string, args map[string]any) error {
	var definition *ToolDefinition
	for i := range operationalTools {
		if operationalTools[i].Name == name {
			definition = &operationalTools[i]
			break
		}
	}
	if definition == nil {
		return fmt.Errorf("tool tidak diizinkan: %s", name)
	}
	if args == nil {
		return fmt.Errorf("arguments untuk %s wajib diisi", name)
	}
	for key, kind := range definition.Arguments {
		if strings.HasSuffix(kind, "?") {
			continue
		}
		value, ok := args[key]
		if !ok || strings.TrimSpace(fmt.Sprint(value)) == "" {
			return fmt.Errorf("argument %s wajib untuk %s", key, name)
		}
	}
	return nil
}
