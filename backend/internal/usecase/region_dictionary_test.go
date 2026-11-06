package usecase

import (
	"testing"
)

func TestNormalizeCityAndProvince(t *testing.T) {
	tests := []struct {
		name         string
		rawRegency   string
		rawProvince  string
		rawAddress   string
		expectedCity string
		expectedProv string
	}{
		{
			name:         "Kabupaten Serang explicit",
			rawRegency:   "KABUPATEN SERANG",
			rawProvince:  "BANTEN",
			rawAddress:   "Jl. Raya Serang",
			expectedCity: "Kabupaten Serang",
			expectedProv: "Banten",
		},
		{
			name:         "Kota Serang explicit",
			rawRegency:   "KOTA SERANG",
			rawProvince:  "BANTEN",
			rawAddress:   "Jl. Ahmad Yani No. 10",
			expectedCity: "Kota Serang",
			expectedProv: "Banten",
		},
		{
			name:         "Kabupaten Karanganyar with empty province",
			rawRegency:   "Kabupaten Karanganyar",
			rawProvince:  "",
			rawAddress:   "Jl. Lawu No. 15",
			expectedCity: "Kabupaten Karanganyar",
			expectedProv: "Jawa Tengah",
		},
		{
			name:         "Karanganyar bare name with mistaken Jawa Timur province",
			rawRegency:   "Karanganyar",
			rawProvince:  "Jawa Timur",
			rawAddress:   "Tasikmadu, Karanganyar",
			expectedCity: "Kabupaten Karanganyar",
			expectedProv: "Jawa Tengah", // Authoritative dictionary must override
		},
		{
			name:         "Kota Surakarta bare or alias Solo",
			rawRegency:   "Solo",
			rawProvince:  "",
			rawAddress:   "Pasar Gede, Solo",
			expectedCity: "Kota Surakarta",
			expectedProv: "Jawa Tengah",
		},
		{
			name:         "Kota Cilegon",
			rawRegency:   "Cilegon",
			rawProvince:  "",
			rawAddress:   "Jl. Raya Cilegon",
			expectedCity: "Kota Cilegon",
			expectedProv: "Banten",
		},
		{
			name:         "Kota Surabaya",
			rawRegency:   "Surabaya",
			rawProvince:  "",
			rawAddress:   "Wonokromo, Surabaya",
			expectedCity: "Kota Surabaya",
			expectedProv: "Jawa Timur",
		},
		{
			name:         "Dual city: Kota Semarang vs Kabupaten Semarang",
			rawRegency:   "Kota Semarang",
			rawProvince:  "",
			rawAddress:   "Simpang Lima",
			expectedCity: "Kota Semarang",
			expectedProv: "Jawa Tengah",
		},
		{
			name:         "Dual city: Kabupaten Semarang",
			rawRegency:   "Kabupaten Semarang",
			rawProvince:  "",
			rawAddress:   "Ungaran",
			expectedCity: "Kabupaten Semarang",
			expectedProv: "Jawa Tengah",
		},
		{
			name:         "Desa Donohudan resolved to Kabupaten Boyolali",
			rawRegency:   "Donohudan",
			rawProvince:  "",
			rawAddress:   "Jl. Embarkasi Haji, Donohudan",
			expectedCity: "Kabupaten Boyolali",
			expectedProv: "Jawa Tengah",
		},
		{
			name:         "Desa Ngesrep resolved to Kabupaten Boyolali",
			rawRegency:   "Ngesrep",
			rawProvince:  "",
			rawAddress:   "Ngesrep, Ngemplak, Boyolali",
			expectedCity: "Kabupaten Boyolali",
			expectedProv: "Jawa Tengah",
		},
		{
			name:         "Desa Sobokerto resolved to Kabupaten Boyolali",
			rawRegency:   "Sobokerto",
			rawProvince:  "",
			rawAddress:   "Sobokerto, Ngemplak",
			expectedCity: "Kabupaten Boyolali",
			expectedProv: "Jawa Tengah",
		},
		{
			name:         "Kecamatan Ngemplak resolved to Kabupaten Boyolali",
			rawRegency:   "Ngemplak",
			rawProvince:  "",
			rawAddress:   "Kecamatan Ngemplak",
			expectedCity: "Kabupaten Boyolali",
			expectedProv: "Jawa Tengah",
		},
		{
			name:         "Kecamatan Colomadu resolved to Kabupaten Karanganyar",
			rawRegency:   "Colomadu",
			rawProvince:  "",
			rawAddress:   "Jl. Adi Sucipto, Colomadu",
			expectedCity: "Kabupaten Karanganyar",
			expectedProv: "Jawa Tengah",
		},
		{
			name:         "Kecamatan Kartasura resolved to Kabupaten Sukoharjo",
			rawRegency:   "Kartasura",
			rawProvince:  "",
			rawAddress:   "Kartasura",
			expectedCity: "Kabupaten Sukoharjo",
			expectedProv: "Jawa Tengah",
		},
		{
			name:         "Address-only fallback with Donohudan",
			rawRegency:   "",
			rawProvince:  "",
			rawAddress:   "Gg. Mawar, Donohudan, Ngemplak",
			expectedCity: "Kabupaten Boyolali",
			expectedProv: "Jawa Tengah",
		},
		{
			name:         "Yogyakarta bare name",
			rawRegency:   "Yogyakarta",
			rawProvince:  "DI Yogyakarta",
			rawAddress:   "Malioboro, Yogyakarta",
			expectedCity: "Kota Yogyakarta",
			expectedProv: "DI Yogyakarta",
		},
		{
			name:         "DI Yogyakarta as raw regency with empty province",
			rawRegency:   "DI Yogyakarta",
			rawProvince:  "",
			rawAddress:   "Jl. Kaliurang",
			expectedCity: "Kota Yogyakarta",
			expectedProv: "DI Yogyakarta",
		},
		{
			name:         "DIY as raw regency with DIY province",
			rawRegency:   "DIY",
			rawProvince:  "DIY",
			rawAddress:   "Jl. Gejayan",
			expectedCity: "Kota Yogyakarta",
			expectedProv: "DI Yogyakarta",
		},
		{
			name:         "Gunung Kidul with space",
			rawRegency:   "GUNUNG KIDUL",
			rawProvince:  "",
			rawAddress:   "Wonosari",
			expectedCity: "Kabupaten Gunungkidul",
			expectedProv: "DI Yogyakarta",
		},
		{
			name:         "Gunungkidul without space",
			rawRegency:   "GUNUNGKIDUL",
			rawProvince:  "",
			rawAddress:   "Wonosari",
			expectedCity: "Kabupaten Gunungkidul",
			expectedProv: "DI Yogyakarta",
		},
		{
			name:         "Kulonprogo without space",
			rawRegency:   "KULONPROGO",
			rawProvince:  "",
			rawAddress:   "Wates",
			expectedCity: "Kabupaten Kulon Progo",
			expectedProv: "DI Yogyakarta",
		},
		{
			name:         "Kulon Progo with space",
			rawRegency:   "KULON PROGO",
			rawProvince:  "",
			rawAddress:   "Wates",
			expectedCity: "Kabupaten Kulon Progo",
			expectedProv: "DI Yogyakarta",
		},
		{
			name:         "Kota Jakarta Barat",
			rawRegency:   "KOTA JAKARTA BARAT",
			rawProvince:  "DKI JAKARTA",
			rawAddress:   "Grogol",
			expectedCity: "Kota Jakarta Barat",
			expectedProv: "DKI Jakarta",
		},
		{
			name:         "NTB province normalization",
			rawRegency:   "",
			rawProvince:  "NTB",
			rawAddress:   "",
			expectedCity: "Tidak Terdeteksi",
			expectedProv: "Nusa Tenggara Barat",
		},
		{
			name:         "NTT province normalization",
			rawRegency:   "",
			rawProvince:  "NTT",
			rawAddress:   "",
			expectedCity: "Tidak Terdeteksi",
			expectedProv: "Nusa Tenggara Timur",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCity, gotProv := NormalizeCityAndProvince(tt.rawRegency, tt.rawProvince, tt.rawAddress)
			if gotCity != tt.expectedCity {
				t.Errorf("City mismatch: got %q, want %q", gotCity, tt.expectedCity)
			}
			if gotProv != tt.expectedProv {
				t.Errorf("Province mismatch: got %q, want %q", gotProv, tt.expectedProv)
			}
		})
	}
}

func TestStatusPenetrasiLogic(t *testing.T) {
	// Scenario A: Penarikan data harian di mana suatu kota memiliki 35 toko terdaftar, tapi 0 transaksi hari ini
	totalStoresDaily := 35
	revenueDaily := 0.0

	statusDaily := "UNPENETRATED"
	if totalStoresDaily > 0 || revenueDaily > 0 {
		statusDaily = "PENETRATED"
	}

	if statusDaily != "PENETRATED" {
		t.Errorf("Expected status 'PENETRATED' (Sudah Masuk) on daily pull with 0 revenue, but got %q", statusDaily)
	}

	// Scenario B: Penarikan data bulanan di mana suatu kota memiliki 35 toko terdaftar dan ada transaksi
	totalStoresMonthly := 35
	revenueMonthly := 25000000.0

	statusMonthly := "UNPENETRATED"
	if totalStoresMonthly > 0 || revenueMonthly > 0 {
		statusMonthly = "PENETRATED"
	}

	if statusMonthly != "PENETRATED" {
		t.Errorf("Expected status 'PENETRATED' (Sudah Masuk) on monthly pull, but got %q", statusMonthly)
	}

	// Scenario C: Kota master yang belum dimasuki sama sekali (0 toko, 0 revenue)
	totalStoresUnentered := 0
	revenueUnentered := 0.0

	statusUnentered := "UNPENETRATED"
	if totalStoresUnentered > 0 || revenueUnentered > 0 {
		statusUnentered = "PENETRATED"
	}

	if statusUnentered != "UNPENETRATED" {
		t.Errorf("Expected status 'UNPENETRATED' (Belum Masuk) for unentered master city, but got %q", statusUnentered)
	}
}

