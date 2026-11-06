/**
 * Master Kamus Wilayah Administratif Republik Indonesia
 * Digunakan untuk normalisasi, validasi, dan penataan hierarki
 * Provinsi -> Kabupaten/Kota -> Kecamatan -> Desa/Kelurahan
 */

// Helper: Ubah teks menjadi Title Case
export const toTitleCase = (str: string): string => {
  if (!str) return ''
  return str
    .toLowerCase()
    .split(' ')
    .filter(Boolean)
    .map(word => word.charAt(0).toUpperCase() + word.slice(1))
    .join(' ')
}

// Regex untuk membersihkan prefix administratif
export const cleanPrefixRegex = /^(KABUPATEN|KOTA\s+ADMINISTRASI|KOTA\s+ADM\.|KOTA|KAB\.|KAB|KODYA|DESA|KELURAHAN|KEL\.|DS\.|KECAMATAN|KEC\.|KEC|PROVINSI|PROP\.|PROV\.)\s+/i

// Master Kamus Alias Provinsi Kanonikal Indonesia (38 Provinsi & Singkatan Populer)
export const CANONICAL_PROVINCE_MAP: Record<string, string> = {
  // DI Yogyakarta
  'DI YOGYAKARTA': 'DI Yogyakarta',
  'D.I. YOGYAKARTA': 'DI Yogyakarta',
  'DAERAH ISTIMEWA YOGYAKARTA': 'DI Yogyakarta',
  'DIY': 'DI Yogyakarta',
  'YOGYAKARTA': 'DI Yogyakarta',
  'JOGJA': 'DI Yogyakarta',
  'DI YOGYA': 'DI Yogyakarta',
  'D.I. YOGYA': 'DI Yogyakarta',
  'YOGYA': 'DI Yogyakarta',

  // DKI Jakarta
  'DKI JAKARTA': 'DKI Jakarta',
  'D.K.I. JAKARTA': 'DKI Jakarta',
  'DAERAH KHUSUS IBUKOTA JAKARTA': 'DKI Jakarta',
  'JAKARTA': 'DKI Jakarta',
  'DKI': 'DKI Jakarta',

  // Pulau Jawa
  'JAWA TENGAH': 'Jawa Tengah',
  'JATENG': 'Jawa Tengah',
  'JAWA TIMUR': 'Jawa Timur',
  'JATIM': 'Jawa Timur',
  'JAWA BARAT': 'Jawa Barat',
  'JABAR': 'Jawa Barat',
  'BANTEN': 'Banten',

  // Bali & Nusa Tenggara
  'BALI': 'Bali',
  'NUSA TENGGARA BARAT': 'Nusa Tenggara Barat',
  'NTB': 'Nusa Tenggara Barat',
  'NUSA TENGGARA TIMUR': 'Nusa Tenggara Timur',
  'NTT': 'Nusa Tenggara Timur',

  // Sumatera
  'SUMATERA UTARA': 'Sumatera Utara',
  'SUMUT': 'Sumatera Utara',
  'SUMATERA BARAT': 'Sumatera Barat',
  'SUMBAR': 'Sumatera Barat',
  'SUMATERA SELATAN': 'Sumatera Selatan',
  'SUMSEL': 'Sumatera Selatan',
  'ACEH': 'Aceh',
  'RIAU': 'Riau',
  'KEPULAUAN RIAU': 'Kepulauan Riau',
  'KEPRI': 'Kepulauan Riau',
  'JAMBI': 'Jambi',
  'BENGKULU': 'Bengkulu',
  'LAMPUNG': 'Lampung',
  'KEPULAUAN BANGKA BELITUNG': 'Kepulauan Bangka Belitung',
  'BANGKA BELITUNG': 'Kepulauan Bangka Belitung',
  'BABEL': 'Kepulauan Bangka Belitung',

  // Kalimantan
  'KALIMANTAN BARAT': 'Kalimantan Barat',
  'KALBAR': 'Kalimantan Barat',
  'KALIMANTAN TENGAH': 'Kalimantan Tengah',
  'KALTENG': 'Kalimantan Tengah',
  'KALIMANTAN SELATAN': 'Kalimantan Selatan',
  'KALSEL': 'Kalimantan Selatan',
  'KALIMANTAN TIMUR': 'Kalimantan Timur',
  'KALTIM': 'Kalimantan Timur',
  'KALIMANTAN UTARA': 'Kalimantan Utara',
  'KALTARA': 'Kalimantan Utara',

  // Sulawesi
  'SULAWESI UTARA': 'Sulawesi Utara',
  'SULUT': 'Sulawesi Utara',
  'SULAWESI TENGAH': 'Sulawesi Tengah',
  'SULTENG': 'Sulawesi Tengah',
  'SULAWESI SELATAN': 'Sulawesi Selatan',
  'SULSEL': 'Sulawesi Selatan',
  'SULAWESI TENGGARA': 'Sulawesi Tenggara',
  'SULTRA': 'Sulawesi Tenggara',
  'GORONTALO': 'Gorontalo',
  'SULAWESI BARAT': 'Sulawesi Barat',
  'SULBAR': 'Sulawesi Barat',

  // Maluku & Papua
  'MALUKU': 'Maluku',
  'MALUKU UTARA': 'Maluku Utara',
  'PAPUA': 'Papua',
  'PAPUA BARAT': 'Papua Barat',
  'PAPUA SELATAN': 'Papua Selatan',
  'PAPUA TENGAH': 'Papua Tengah',
  'PAPUA PEGUNUNGAN': 'Papua Pegunungan',
  'PAPUA BARAT DAYA': 'Papua Barat Daya',
}

// Normalisasi String Menjadi Nama Provinsi Kanonikal Resmi
export const normalizeCanonicalProvince = (raw: string | undefined | null): string | null => {
  if (!raw) return null
  const cleaned = raw.replace(cleanPrefixRegex, '').trim().toUpperCase()
  if (!cleaned || cleaned === 'TIDAK TERDETEKSI' || cleaned === 'LAINNYA') return null

  if (CANONICAL_PROVINCE_MAP[cleaned]) {
    return CANONICAL_PROVINCE_MAP[cleaned]
  }

  // Cek pencocokan frasa
  for (const [key, canonical] of Object.entries(CANONICAL_PROVINCE_MAP)) {
    const regex = new RegExp(`\\b${key}\\b`, 'i')
    if (regex.test(cleaned)) {
      return canonical
    }
  }

  return toTitleCase(cleaned)
}

// 1. Daftar Resmi Kota Otonom (Tanpa kata Kabupaten)
export const OFFICIAL_KOTA_SET = new Set([
  // Jawa & DIY
  'SURAKARTA', 'SOLO', 'SALATIGA', 'SEMARANG', 'MAGELANG', 'PEKALONGAN', 'TEGAL',
  'YOGYAKARTA', 'JOGJA',
  'SURABAYA', 'MALANG', 'BATU', 'KEDIRI', 'BLITAR', 'MADIUN', 'MOJOKERTO', 'PASURUAN', 'PROBOLINGGO',
  'BANDUNG', 'BEKASI', 'BOGOR', 'CIMAHI', 'CIREBON', 'DEPOK', 'SUKABUMI', 'TASIKMALAYA', 'BANJAR',
  'SERANG', 'CILEGON', 'TANGERANG', 'TANGERANG SELATAN',
  'JAKARTA PUSAT', 'JAKARTA UTARA', 'JAKARTA BARAT', 'JAKARTA SELATAN', 'JAKARTA TIMUR',
  // Luar Jawa
  'DENPASAR', 'MATARAM', 'BIMA', 'KUPANG',
  'MEDAN', 'BINJAI', 'TEBING TINGGI', 'PEMATANGSIANTAR', 'TANJUNGBALAI', 'SIBOLGA', 'PADANGSIDIMPUAN', 'GUNUNGSITOLI',
  'PADANG', 'SOLOK', 'SAWAHLUNTO', 'PADANG PANJANG', 'BUKITTINGGI', 'PAYAKUMBUH', 'PARIAMAN',
  'PEKANBARU', 'DUMAI', 'BATAM', 'TANJUNGPINANG', 'JAMBI', 'SUNGAI PENUH',
  'PALEMBANG', 'PRABUMULIH', 'PAGAR ALAM', 'LUBUKLINGGAU', 'PANGKALPINANG',
  'BENGKULU', 'BANDAR LAMPUNG', 'METRO',
  'PONTIANAK', 'SINGKAWANG', 'BANJARMASIN', 'BANJARBARU', 'PALANGKA RAYA', 'SAMARINDA', 'BALIKPAPAN', 'BONTANG', 'TARAKAN',
  'MANADO', 'BITUNG', 'TOMOHON', 'KOTAMOBAGU', 'PALU', 'MAKASSAR', 'PAREPARE', 'PALOPO', 'KENDARI', 'BAUBAU', 'GORONTALO',
  'AMBON', 'TUAL', 'TERNATE', 'TIDORE KEPULAUAN', 'JAYAPURA', 'SORONG'
])

// 2. Kamus Master Seluruh Kabupaten / Kota di Indonesia (Base Name -> Tipe & Provinsi)
export const OFFICIAL_REGENCY_TO_PROVINCE: Record<string, string> = {
  // --- JAWA TENGAH ---
  'BANJARNEGARA': 'Jawa Tengah',
  'BANYUMAS': 'Jawa Tengah',
  'BATANG': 'Jawa Tengah',
  'BLORA': 'Jawa Tengah',
  'BOYOLALI': 'Jawa Tengah',
  'BREBES': 'Jawa Tengah',
  'CILACAP': 'Jawa Tengah',
  'DEMAK': 'Jawa Tengah',
  'GROBOGAN': 'Jawa Tengah',
  'JEPARA': 'Jawa Tengah',
  'KARANGANYAR': 'Jawa Tengah',
  'KEBUMEN': 'Jawa Tengah',
  'KENDAL': 'Jawa Tengah',
  'KLATEN': 'Jawa Tengah',
  'KUDUS': 'Jawa Tengah',
  'MAGELANG': 'Jawa Tengah',
  'PATI': 'Jawa Tengah',
  'PEKALONGAN': 'Jawa Tengah',
  'PEMALANG': 'Jawa Tengah',
  'PURBALINGGA': 'Jawa Tengah',
  'PURWOREJO': 'Jawa Tengah',
  'REMBANG': 'Jawa Tengah',
  'SEMARANG': 'Jawa Tengah',
  'SRAGEN': 'Jawa Tengah',
  'SUKOHARJO': 'Jawa Tengah',
  'TEGAL': 'Jawa Tengah',
  'TEMANGGUNG': 'Jawa Tengah',
  'WONOGIRI': 'Jawa Tengah',
  'WONOSOBO': 'Jawa Tengah',
  'SALATIGA': 'Jawa Tengah',
  'SURAKARTA': 'Jawa Tengah',
  'SOLO': 'Jawa Tengah',

  // --- DI YOGYAKARTA ---
  'BANTUL': 'DI Yogyakarta',
  'GUNUNGKIDUL': 'DI Yogyakarta',
  'GUNUNG KIDUL': 'DI Yogyakarta',
  'KULON PROGO': 'DI Yogyakarta',
  'KULONPROGO': 'DI Yogyakarta',
  'SLEMAN': 'DI Yogyakarta',
  'YOGYAKARTA': 'DI Yogyakarta',
  'JOGJA': 'DI Yogyakarta',
  'DI YOGYAKARTA': 'DI Yogyakarta',
  'D.I. YOGYAKARTA': 'DI Yogyakarta',
  'DIY': 'DI Yogyakarta',
  'DAERAH ISTIMEWA YOGYAKARTA': 'DI Yogyakarta',

  // --- JAWA TIMUR ---
  'BANGKALAN': 'Jawa Timur',
  'BANYUWANGI': 'Jawa Timur',
  'BLITAR': 'Jawa Timur',
  'BOJONEGORO': 'Jawa Timur',
  'BONDOWOSO': 'Jawa Timur',
  'GRESIK': 'Jawa Timur',
  'JEMBER': 'Jawa Timur',
  'JOMBANG': 'Jawa Timur',
  'KEDIRI': 'Jawa Timur',
  'LAMONGAN': 'Jawa Timur',
  'LUMAJANG': 'Jawa Timur',
  'MADIUN': 'Jawa Timur',
  'MAGETAN': 'Jawa Timur',
  'MALANG': 'Jawa Timur',
  'MOJOKERTO': 'Jawa Timur',
  'NGANJUK': 'Jawa Timur',
  'NGAWI': 'Jawa Timur',
  'PACITAN': 'Jawa Timur',
  'PAMEKASAN': 'Jawa Timur',
  'PASURUAN': 'Jawa Timur',
  'PONOROGO': 'Jawa Timur',
  'PROBOLINGGO': 'Jawa Timur',
  'SAMPANG': 'Jawa Timur',
  'SIDOARJO': 'Jawa Timur',
  'SITUBONDO': 'Jawa Timur',
  'SUMENEP': 'Jawa Timur',
  'TRENGGALEK': 'Jawa Timur',
  'TUBAN': 'Jawa Timur',
  'TULUNGAGUNG': 'Jawa Timur',
  'BATU': 'Jawa Timur',
  'SURABAYA': 'Jawa Timur',

  // --- JAWA BARAT ---
  'BANDUNG': 'Jawa Barat',
  'BANDUNG BARAT': 'Jawa Barat',
  'BEKASI': 'Jawa Barat',
  'BOGOR': 'Jawa Barat',
  'CIAMIS': 'Jawa Barat',
  'CIANJUR': 'Jawa Barat',
  'CIREBON': 'Jawa Barat',
  'GARUT': 'Jawa Barat',
  'INDRAMAYU': 'Jawa Barat',
  'KARAWANG': 'Jawa Barat',
  'KUNINGAN': 'Jawa Barat',
  'MAJALENGKA': 'Jawa Barat',
  'PANGANDARAN': 'Jawa Barat',
  'PURWAKARTA': 'Jawa Barat',
  'SUBANG': 'Jawa Barat',
  'SUKABUMI': 'Jawa Barat',
  'SUMEDANG': 'Jawa Barat',
  'TASIKMALAYA': 'Jawa Barat',
  'BANJAR': 'Jawa Barat',
  'CIMAHI': 'Jawa Barat',
  'DEPOK': 'Jawa Barat',

  // --- BANTEN ---
  'LEBAK': 'Banten',
  'PANDEGLANG': 'Banten',
  'SERANG': 'Banten',
  'TANGERANG': 'Banten',
  'CILEGON': 'Banten',
  'TANGERANG SELATAN': 'Banten',

  // --- DKI JAKARTA ---
  'KEPULAUAN SERIBU': 'DKI Jakarta',
  'JAKARTA BARAT': 'DKI Jakarta',
  'JAKARTA PUSAT': 'DKI Jakarta',
  'JAKARTA SELATAN': 'DKI Jakarta',
  'JAKARTA TIMUR': 'DKI Jakarta',
  'JAKARTA UTARA': 'DKI Jakarta',

  // --- BALI & NUSA TENGGARA ---
  'BADUNG': 'Bali',
  'BANGLI': 'Bali',
  'BULELENG': 'Bali',
  'GIANYAR': 'Bali',
  'JEMBRANA': 'Bali',
  'KARANGASEM': 'Bali',
  'KLUNGKUNG': 'Bali',
  'TABANAN': 'Bali',
  'DENPASAR': 'Bali',
  'LOMBOK BARAT': 'Nusa Tenggara Barat',
  'LOMBOK TENGAH': 'Nusa Tenggara Barat',
  'LOMBOK TIMUR': 'Nusa Tenggara Barat',
  'LOMBOK UTARA': 'Nusa Tenggara Barat',
  'SUMBAWA': 'Nusa Tenggara Barat',
  'SUMBAWA BARAT': 'Nusa Tenggara Barat',
  'DOMPU': 'Nusa Tenggara Barat',
  'BIMA': 'Nusa Tenggara Barat',
  'MATARAM': 'Nusa Tenggara Barat',

  // --- LAMPUNG ---
  'LAMPUNG BARAT': 'Lampung',
  'LAMPUNG SELATAN': 'Lampung',
  'LAMPUNG TENGAH': 'Lampung',
  'LAMPUNG TIMUR': 'Lampung',
  'LAMPUNG UTARA': 'Lampung',
  'MESUJI': 'Lampung',
  'PESAWARAN': 'Lampung',
  'PESISIR BARAT': 'Lampung',
  'PRINGSEWU': 'Lampung',
  'TANGGAMUS': 'Lampung',
  'TULANG BAWANG': 'Lampung',
  'TULANG BAWANG BARAT': 'Lampung',
  'WAY KANAN': 'Lampung',
  'BANDAR LAMPUNG': 'Lampung',
  'METRO': 'Lampung'
}

// 3. Kamus Cerdas: Desa / Kelurahan / Kecamatan Populer yang Sering Keliru Jadi Regency
export interface KnownSubRegionInfo {
  regency: string
  district?: string
  village?: string
  level: 'DESA' | 'KECAMATAN'
}

export const KNOWN_SUB_REGIONS_MAP: Record<string, KnownSubRegionInfo> = {
  // Desa di Boyolali (Ngemplak)
  'DONOHUDAN': { regency: 'Kabupaten Boyolali', district: 'Ngemplak', village: 'Donohudan', level: 'DESA' },
  'NGESREP': { regency: 'Kabupaten Boyolali', district: 'Ngemplak', village: 'Ngesrep', level: 'DESA' },
  'SOBOKERTO': { regency: 'Kabupaten Boyolali', district: 'Ngemplak', village: 'Sobokerto', level: 'DESA' },
  'GAGAKSIPAT': { regency: 'Kabupaten Boyolali', district: 'Ngemplak', village: 'Gagaksipat', level: 'DESA' },
  'DIBAL': { regency: 'Kabupaten Boyolali', district: 'Ngemplak', village: 'Dibal', level: 'DESA' },
  'SINDON': { regency: 'Kabupaten Boyolali', district: 'Ngemplak', village: 'Sindon', level: 'DESA' },
  'PANDEYAN': { regency: 'Kabupaten Boyolali', district: 'Ngemplak', village: 'Pandeyan', level: 'DESA' },
  'SAWAHAN': { regency: 'Kabupaten Boyolali', district: 'Ngemplak', village: 'Sawahan', level: 'DESA' },
  'KISMOYOSO': { regency: 'Kabupaten Boyolali', district: 'Ngemplak', village: 'Kismoyoso', level: 'DESA' },
  'GIRIROTO': { regency: 'Kabupaten Boyolali', district: 'Ngemplak', village: 'Giriroto', level: 'DESA' },

  // Kecamatan di Boyolali
  'NGEMPLAK': { regency: 'Kabupaten Boyolali', district: 'Ngemplak', level: 'KECAMATAN' },
  'BANYUDONO': { regency: 'Kabupaten Boyolali', district: 'Banyudono', level: 'KECAMATAN' },
  'SAWIT': { regency: 'Kabupaten Boyolali', district: 'Sawit', level: 'KECAMATAN' },
  'TERAS': { regency: 'Kabupaten Boyolali', district: 'Teras', level: 'KECAMATAN' },
  'MOJOSONGO': { regency: 'Kabupaten Boyolali', district: 'Mojosongo', level: 'KECAMATAN' },
  'AMPEL': { regency: 'Kabupaten Boyolali', district: 'Ampel', level: 'KECAMATAN' },
  'CEPOGO': { regency: 'Kabupaten Boyolali', district: 'Cepogo', level: 'KECAMATAN' },
  'SELO': { regency: 'Kabupaten Boyolali', district: 'Selo', level: 'KECAMATAN' },
  'MUSUK': { regency: 'Kabupaten Boyolali', district: 'Musuk', level: 'KECAMATAN' },
  'ANDONG': { regency: 'Kabupaten Boyolali', district: 'Andong', level: 'KECAMATAN' },
  'KLEGO': { regency: 'Kabupaten Boyolali', district: 'Klego', level: 'KECAMATAN' },
  'SIMO': { regency: 'Kabupaten Boyolali', district: 'Simo', level: 'KECAMATAN' },
  'NOGOSARI': { regency: 'Kabupaten Boyolali', district: 'Nogosari', level: 'KECAMATAN' },
  'KARANGGEDE': { regency: 'Kabupaten Boyolali', district: 'Karanggede', level: 'KECAMATAN' },
  'WONOSEGORO': { regency: 'Kabupaten Boyolali', district: 'Wonosegoro', level: 'KECAMATAN' },
  'JUWANGI': { regency: 'Kabupaten Boyolali', district: 'Juwangi', level: 'KECAMATAN' },

  // Karanganyar
  'COLOMADU': { regency: 'Kabupaten Karanganyar', district: 'Colomadu', level: 'KECAMATAN' },
  'GONDANGREJO': { regency: 'Kabupaten Karanganyar', district: 'Gondangrejo', level: 'KECAMATAN' },
  'JATEN': { regency: 'Kabupaten Karanganyar', district: 'Jaten', level: 'KECAMATAN' },
  'KEBAKKRAMAT': { regency: 'Kabupaten Karanganyar', district: 'Kebakkramat', level: 'KECAMATAN' },
  'TASIKMADU': { regency: 'Kabupaten Karanganyar', district: 'Tasikmadu', level: 'KECAMATAN' },
  'MATESIH': { regency: 'Kabupaten Karanganyar', district: 'Matesih', level: 'KECAMATAN' },
  'KARANGPANDAN': { regency: 'Kabupaten Karanganyar', district: 'Karangpandan', level: 'KECAMATAN' },
  'TAWANGMANGU': { regency: 'Kabupaten Karanganyar', district: 'Tawangmangu', level: 'KECAMATAN' },
  'NGARGOYOSO': { regency: 'Kabupaten Karanganyar', district: 'Ngargoyoso', level: 'KECAMATAN' },
  'MOJOLABAN': { regency: 'Kabupaten Sukoharjo', district: 'Mojolaban', level: 'KECAMATAN' },
  'BAKI': { regency: 'Kabupaten Sukoharjo', district: 'Baki', level: 'KECAMATAN' },
  'GROGOL': { regency: 'Kabupaten Sukoharjo', district: 'Grogol', level: 'KECAMATAN' },
  'KARTASURA': { regency: 'Kabupaten Sukoharjo', district: 'Kartasura', level: 'KECAMATAN' },
  'POLOKARTO': { regency: 'Kabupaten Sukoharjo', district: 'Polokarto', level: 'KECAMATAN' },
  'TAWANGSARI': { regency: 'Kabupaten Sukoharjo', district: 'Tawangsari', level: 'KECAMATAN' },
  'WERU': { regency: 'Kabupaten Sukoharjo', district: 'Weru', level: 'KECAMATAN' },
  'BULU': { regency: 'Kabupaten Sukoharjo', district: 'Bulu', level: 'KECAMATAN' },
  'NGUTER': { regency: 'Kabupaten Sukoharjo', district: 'Nguter', level: 'KECAMATAN' },
  'BENDOSARI': { regency: 'Kabupaten Sukoharjo', district: 'Bendosari', level: 'KECAMATAN' },

  // Kota Surakarta (Solo)
  'BANJARSARI': { regency: 'Kota Surakarta', district: 'Banjarsari', level: 'KECAMATAN' },
  'JEBRES': { regency: 'Kota Surakarta', district: 'Jebres', level: 'KECAMATAN' },
  'LAWEYAN': { regency: 'Kota Surakarta', district: 'Laweyan', level: 'KECAMATAN' },
  'PASAR KLIWON': { regency: 'Kota Surakarta', district: 'Pasar Kliwon', level: 'KECAMATAN' },
  'SERENGAN': { regency: 'Kota Surakarta', district: 'Serengan', level: 'KECAMATAN' },

  // Klaten
  'DELANGGU': { regency: 'Kabupaten Klaten', district: 'Delanggu', level: 'KECAMATAN' },
  'PEDAN': { regency: 'Kabupaten Klaten', district: 'Pedan', level: 'KECAMATAN' },
  'CEPER': { regency: 'Kabupaten Klaten', district: 'Ceper', level: 'KECAMATAN' },
  'PRAMBANAN': { regency: 'Kabupaten Klaten', district: 'Prambanan', level: 'KECAMATAN' },
  'JOGONALAN': { regency: 'Kabupaten Klaten', district: 'Jogonalan', level: 'KECAMATAN' },
  'WEDI': { regency: 'Kabupaten Klaten', district: 'Wedi', level: 'KECAMATAN' },
  'BAYAT': { regency: 'Kabupaten Klaten', district: 'Bayat', level: 'KECAMATAN' },
  'CAWAS': { regency: 'Kabupaten Klaten', district: 'Cawas', level: 'KECAMATAN' },
  'TRUCUK': { regency: 'Kabupaten Klaten', district: 'Trucuk', level: 'KECAMATAN' },
  'POLANHARJO': { regency: 'Kabupaten Klaten', district: 'Polanharjo', level: 'KECAMATAN' },
  'TULUNG': { regency: 'Kabupaten Klaten', district: 'Tulung', level: 'KECAMATAN' },
  'JATINOM': { regency: 'Kabupaten Klaten', district: 'Jatinom', level: 'KECAMATAN' },
  'KARANGANOM': { regency: 'Kabupaten Klaten', district: 'Karanganom', level: 'KECAMATAN' },

  // Sragen
  'GEMOLONG': { regency: 'Kabupaten Sragen', district: 'Gemolong', level: 'KECAMATAN' },
  'MASARAN': { regency: 'Kabupaten Sragen', district: 'Masaran', level: 'KECAMATAN' },
  'SAMBUNGMACAN': { regency: 'Kabupaten Sragen', district: 'Sambungmacan', level: 'KECAMATAN' },
  'GONDANG': { regency: 'Kabupaten Sragen', district: 'Gondang', level: 'KECAMATAN' },
  'TANGEN': { regency: 'Kabupaten Sragen', district: 'Tangen', level: 'KECAMATAN' },
  'KALIJAMBE': { regency: 'Kabupaten Sragen', district: 'Kalijambe', level: 'KECAMATAN' },
  'PLUPUH': { regency: 'Kabupaten Sragen', district: 'Plupuh', level: 'KECAMATAN' },
  'SUMBERLAWANG': { regency: 'Kabupaten Sragen', district: 'Sumberlawang', level: 'KECAMATAN' },
}

/**
 * Mengecek apakah suatu nama adalah Kabupaten/Kota resmi di Indonesia.
 */
export const isValidOfficialRegency = (val: string | null | undefined): boolean => {
  if (!val) return false
  const clean = val.replace(cleanPrefixRegex, '').trim().toUpperCase()
  return Boolean(OFFICIAL_REGENCY_TO_PROVINCE[clean])
}

/**
 * Membersihkan & menstandardisasi nama Kabupaten/Kota
 * Contoh: "BOYOLALI" -> "Kabupaten Boyolali", "SURAKARTA" -> "Kota Surakarta", "SOLO" -> "Kota Surakarta"
 */
export const formatCleanRegencyName = (rawName: string): string => {
  let cleaned = rawName.trim()
  if (!cleaned) return ''

  let isKota = /^(KOTA\s+ADMINISTRASI|KOTA\s+ADM\.|KOTA|KODYA)\s+/i.test(cleaned)
  let isKab = /^(KABUPATEN|KAB\.|KAB)\s+/i.test(cleaned)

  const baseName = cleaned.replace(cleanPrefixRegex, '').trim()
  const baseUpper = baseName.toUpperCase()

  if (baseUpper === 'SOLO' || baseUpper === 'SURAKARTA' || baseUpper === 'KOTA SOLO' || baseUpper === 'KOTA SURAKARTA') return 'Kota Surakarta'
  if (baseUpper === 'JOGJA' || baseUpper === 'YOGYAKARTA' || baseUpper === 'DIY' || baseUpper === 'DI YOGYAKARTA' || baseUpper === 'DAERAH ISTIMEWA YOGYAKARTA') return 'Kota Yogyakarta'
  if (baseUpper === 'GUNUNG KIDUL' || baseUpper === 'GUNUNGKIDUL') return 'Kabupaten Gunungkidul'
  if (baseUpper === 'KULON PROGO' || baseUpper === 'KULONPROGO') return 'Kabupaten Kulon Progo'
  if (baseUpper === 'KEPULAUAN SERIBU' || baseUpper === 'PULAU SERIBU') return 'Kabupaten Kepulauan Seribu'
  if (baseUpper === 'TANGERANG SELATAN' || baseUpper === 'TANGSEL') return 'Kota Tangerang Selatan'

  if (!isKota && !isKab) {
    if (OFFICIAL_KOTA_SET.has(baseUpper)) {
      isKota = true
    } else {
      isKab = true
    }
  }

  if (isKota) {
    return `Kota ${toTitleCase(baseName)}`
  }
  return `Kabupaten ${toTitleCase(baseName)}`
}

/**
 * Ekstraksi Wilayah Cerdas dari string Alamat Toko
 */
export const extractRegionFromAddressString = (
  address: string | undefined | null,
  type: 'PROVINCE' | 'KABUPATEN' | 'KECAMATAN' | 'DESA'
): string | null => {
  if (!address) return null
  const upperAddr = address.toUpperCase()

  if (type === 'PROVINCE') {
    // 1. Cek apakah ada nama regency resmi di alamat untuk mendapatkan provinsi otoritatif
    for (const [regKey, prov] of Object.entries(OFFICIAL_REGENCY_TO_PROVINCE)) {
      const regex = new RegExp(`\\b${regKey}\\b`, 'i')
      if (regex.test(upperAddr)) {
        return prov
      }
    }

    // 2. Cek sub-region map
    for (const [subKey, info] of Object.entries(KNOWN_SUB_REGIONS_MAP)) {
      const regex = new RegExp(`\\b${subKey}\\b`, 'i')
      if (regex.test(upperAddr)) {
        const cleanParent = info.regency.replace(cleanPrefixRegex, '').trim().toUpperCase()
        if (OFFICIAL_REGENCY_TO_PROVINCE[cleanParent]) {
          return OFFICIAL_REGENCY_TO_PROVINCE[cleanParent]
        }
      }
    }

    // 3. Cek kamus kanonikal provinsi
    for (const [provKey, canonical] of Object.entries(CANONICAL_PROVINCE_MAP)) {
      const regex = new RegExp(`\\b${provKey}\\b`, 'i')
      if (regex.test(upperAddr)) {
        return canonical
      }
    }

    return null
  }

  if (type === 'KABUPATEN') {
    // 1. Cek regex eksplisit (Kabupaten/Kota ...)
    const kabMatch = upperAddr.match(/(KABUPATEN|KOTA\s+ADMINISTRASI|KOTA\s+ADM\.|KOTA|KAB\.|KAB|KODYA)\s+([A-Z\s]+?)(?:,|$)/)
    if (kabMatch && kabMatch[1] && kabMatch[2]) {
      const base = kabMatch[2].trim().replace(/^(KECAMATAN|DESA|KELURAHAN)\s+/i, '')
      if (isValidOfficialRegency(base)) {
        return formatCleanRegencyName(base)
      }
    }

    // 2. Cek apakah ada nama kabupaten resmi yang termuat di dalam alamat
    for (const regKey of Object.keys(OFFICIAL_REGENCY_TO_PROVINCE)) {
      // Pastikan pencarian menggunakan batas kata agar tidak salah cocok
      const regex = new RegExp(`\\b${regKey}\\b`, 'i')
      if (regex.test(upperAddr)) {
        return formatCleanRegencyName(regKey)
      }
    }

    // 3. Cek sub-region populer di alamat (misal: "Donohudan" -> Kabupaten Boyolali)
    for (const [subKey, info] of Object.entries(KNOWN_SUB_REGIONS_MAP)) {
      const regex = new RegExp(`\\b${subKey}\\b`, 'i')
      if (regex.test(upperAddr)) {
        return info.regency
      }
    }

    return null
  }

  if (type === 'KECAMATAN') {
    const kecMatch = upperAddr.match(/(?:KECAMATAN|KEC\.|KEC)\s+([A-Z\s]+?)(?:,|$)/)
    if (kecMatch && kecMatch[1]) {
      const cleaned = kecMatch[1].replace(cleanPrefixRegex, '').trim()
      if (cleaned) return toTitleCase(cleaned)
    }

    // Cek sub-region map
    for (const [subKey, info] of Object.entries(KNOWN_SUB_REGIONS_MAP)) {
      const regex = new RegExp(`\\b${subKey}\\b`, 'i')
      if (regex.test(upperAddr) && info.district) {
        return toTitleCase(info.district)
      }
    }
    return null
  }

  if (type === 'DESA') {
    const desaMatch = upperAddr.match(/(?:DESA|KELURAHAN|KEL\.|DS\.)\s+([A-Z\s]+?)(?:,|$)/)
    if (desaMatch && desaMatch[1]) {
      const cleaned = desaMatch[1].replace(cleanPrefixRegex, '').trim()
      if (cleaned) return toTitleCase(cleaned)
    }

    // Cek sub-region map
    for (const [subKey, info] of Object.entries(KNOWN_SUB_REGIONS_MAP)) {
      const regex = new RegExp(`\\b${subKey}\\b`, 'i')
      if (regex.test(upperAddr) && info.level === 'DESA') {
        return toTitleCase(info.village || subKey)
      }
    }
    return null
  }

  return null
}

/**
 * Resolusi Hierarki Administratif Toko yang Akurat & Menyeluruh
 * Memastikan tidak ada desa/kecamatan/kabupaten yang salah tingkat atau terduplikasi.
 */
export const getAccurateStoreRegion = (
  store: {
    province?: string | null
    regency?: string | null
    district?: string | null
    village?: string | null
    address?: string | null
  },
  type: 'PROVINCE' | 'KABUPATEN' | 'KECAMATAN' | 'DESA'
): string | null => {
  const rawRegency = (store.regency || '').trim()
  const rawDistrict = (store.district || '').trim()
  const rawVillage = (store.village || '').trim()
  const rawProvince = (store.province || '').trim()
  const rawAddress = (store.address || '').trim()

  const cleanRegUpper = rawRegency.replace(cleanPrefixRegex, '').trim().toUpperCase()
  const cleanDistUpper = rawDistrict.replace(cleanPrefixRegex, '').trim().toUpperCase()
  const cleanVillUpper = rawVillage.replace(cleanPrefixRegex, '').trim().toUpperCase()

  // ─── TIPE: PROVINSI ──────────────────────────────────────────────────────────
  if (type === 'PROVINCE') {
    // 1. Prioritas 1: Jika field regency valid (resmi) / dikenal, ambil provinsi resminya
    // Ini mencegah regency seperti 'Sleman', 'Kudus', 'Kediri', 'Sidoarjo' bocor ke daftar provinsi
    if (cleanRegUpper && OFFICIAL_REGENCY_TO_PROVINCE[cleanRegUpper]) {
      return OFFICIAL_REGENCY_TO_PROVINCE[cleanRegUpper]
    }
    if (cleanRegUpper && KNOWN_SUB_REGIONS_MAP[cleanRegUpper]) {
      const parentReg = KNOWN_SUB_REGIONS_MAP[cleanRegUpper].regency
      const cleanParentUpper = parentReg.replace(cleanPrefixRegex, '').trim().toUpperCase()
      if (OFFICIAL_REGENCY_TO_PROVINCE[cleanParentUpper]) {
        return OFFICIAL_REGENCY_TO_PROVINCE[cleanParentUpper]
      }
    }
    if (cleanDistUpper && KNOWN_SUB_REGIONS_MAP[cleanDistUpper]) {
      const parentReg = KNOWN_SUB_REGIONS_MAP[cleanDistUpper].regency
      const cleanParentUpper = parentReg.replace(cleanPrefixRegex, '').trim().toUpperCase()
      if (OFFICIAL_REGENCY_TO_PROVINCE[cleanParentUpper]) {
        return OFFICIAL_REGENCY_TO_PROVINCE[cleanParentUpper]
      }
    }
    if (cleanVillUpper && KNOWN_SUB_REGIONS_MAP[cleanVillUpper]) {
      const parentReg = KNOWN_SUB_REGIONS_MAP[cleanVillUpper].regency
      const cleanParentUpper = parentReg.replace(cleanPrefixRegex, '').trim().toUpperCase()
      if (OFFICIAL_REGENCY_TO_PROVINCE[cleanParentUpper]) {
        return OFFICIAL_REGENCY_TO_PROVINCE[cleanParentUpper]
      }
    }

    // 2. Prioritas 2: Normalisasi nilai rawProvince via Kamus Kanonikal (DI Yogyakarta, DKI Jakarta, NTB, dll)
    const canonProv = normalizeCanonicalProvince(rawProvince)
    if (canonProv && canonProv !== 'Tidak Terdeteksi') {
      return canonProv
    }

    // 3. Prioritas 3: Ekstraksi dari alamat
    return extractRegionFromAddressString(rawAddress, 'PROVINCE')
  }

  // ─── TIPE: KABUPATEN / KOTA ──────────────────────────────────────────────────
  if (type === 'KABUPATEN') {
    // 1. Cek apakah field regency valid sebagai Kabupaten/Kota resmi
    if (cleanRegUpper && isValidOfficialRegency(cleanRegUpper)) {
      return formatCleanRegencyName(cleanRegUpper)
    }

    // 2. Jika field regency ternyata nama desa/kecamatan (seperti Donohudan), cari kabupaten induknya
    if (cleanRegUpper && KNOWN_SUB_REGIONS_MAP[cleanRegUpper]) {
      return KNOWN_SUB_REGIONS_MAP[cleanRegUpper].regency
    }
    if (cleanDistUpper && KNOWN_SUB_REGIONS_MAP[cleanDistUpper]) {
      return KNOWN_SUB_REGIONS_MAP[cleanDistUpper].regency
    }
    if (cleanVillUpper && KNOWN_SUB_REGIONS_MAP[cleanVillUpper]) {
      return KNOWN_SUB_REGIONS_MAP[cleanVillUpper].regency
    }

    // 3. Ekstraksi dari alamat
    const fromAddr = extractRegionFromAddressString(rawAddress, 'KABUPATEN')
    if (fromAddr && isValidOfficialRegency(fromAddr)) {
      return fromAddr
    }

    // 4. Jika tetap tidak cocok dengan daftar resmi Kabupaten di Indonesia, kembalikan null (Tidak Terdeteksi)
    return null
  }

  // ─── TIPE: KECAMATAN ────────────────────────────────────────────────────────
  if (type === 'KECAMATAN') {
    // 1. Prioritaskan field district jika valid
    if (rawDistrict && rawDistrict !== 'Tidak Terdeteksi') {
      const clean = rawDistrict.replace(cleanPrefixRegex, '').trim()
      if (clean) return toTitleCase(clean)
    }

    // 2. Jika regency di DB sebenarnya adalah nama kecamatan (misal: "Ngemplak", "Colomadu")
    if (cleanRegUpper && KNOWN_SUB_REGIONS_MAP[cleanRegUpper]?.level === 'KECAMATAN') {
      return toTitleCase(KNOWN_SUB_REGIONS_MAP[cleanRegUpper].district || cleanRegUpper)
    }

    // 3. Jika village di DB punya info kecamatan di kamus
    if (cleanVillUpper && KNOWN_SUB_REGIONS_MAP[cleanVillUpper]?.district) {
      return toTitleCase(KNOWN_SUB_REGIONS_MAP[cleanVillUpper].district!)
    }

    // 4. Fallback ke alamat
    const fromAddr = extractRegionFromAddressString(rawAddress, 'KECAMATAN')
    if (fromAddr) return fromAddr

    return null
  }

  // ─── TIPE: DESA / KELURAHAN ─────────────────────────────────────────────────
  if (type === 'DESA') {
    // 1. Prioritaskan field village
    if (rawVillage && rawVillage !== 'Tidak Terdeteksi') {
      const clean = rawVillage.replace(cleanPrefixRegex, '').trim()
      if (clean) return toTitleCase(clean)
    }

    // 2. Jika regency di DB sebenarnya adalah nama desa (misal: "Donohudan", "Ngesrep", "Sobokerto")
    if (cleanRegUpper && KNOWN_SUB_REGIONS_MAP[cleanRegUpper]?.level === 'DESA') {
      return toTitleCase(KNOWN_SUB_REGIONS_MAP[cleanRegUpper].village || cleanRegUpper)
    }

    // 3. Jika district di DB sebenarnya adalah nama desa
    if (cleanDistUpper && KNOWN_SUB_REGIONS_MAP[cleanDistUpper]?.level === 'DESA') {
      return toTitleCase(KNOWN_SUB_REGIONS_MAP[cleanDistUpper].village || cleanDistUpper)
    }

    // 4. Fallback ke alamat
    const fromAddr = extractRegionFromAddressString(rawAddress, 'DESA')
    if (fromAddr) return fromAddr

    return null
  }

  return null
}
