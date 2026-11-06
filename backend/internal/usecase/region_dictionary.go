package usecase

import (
	"regexp"
	"strings"
)

var (
	kabPrefixRegex = regexp.MustCompile(`(?i)^(KABUPATEN|KAB\.|KAB)\s+`)
	kotaPrefixRegex = regexp.MustCompile(`(?i)^(KOTA\s+ADMINISTRASI|KOTA\s+ADM\.|KOTA|KODYA)\s+`)
	cleanPrefixRegex = regexp.MustCompile(`(?i)^(KABUPATEN|KOTA\s+ADMINISTRASI|KOTA\s+ADM\.|KOTA|KAB\.|KAB|KODYA|DESA|KELURAHAN|KEL\.|DS\.|KECAMATAN|KEC\.|KEC)\s+`)
	regencyAddressExtractor = regexp.MustCompile(`(?i)(KABUPATEN|KOTA\s+ADMINISTRASI|KOTA\s+ADM\.|KOTA|KAB\.|KAB|KODYA)\s+([a-zA-Z\s]+?)(?:,|$)`)
)

func toTitleCase(s string) string {
	words := strings.Fields(s)
	for i, w := range words {
		wLower := strings.ToLower(w)
		if len(wLower) > 0 {
			words[i] = strings.ToUpper(wLower[:1]) + wLower[1:]
		}
	}
	return strings.Join(words, " ")
}

// CanonicalProvincesMap maps province aliases and acronyms to their official names.
var CanonicalProvincesMap = map[string]string{
	// DI Yogyakarta
	"DI YOGYAKARTA":              "DI Yogyakarta",
	"D.I. YOGYAKARTA":            "DI Yogyakarta",
	"DAERAH ISTIMEWA YOGYAKARTA": "DI Yogyakarta",
	"DIY":                        "DI Yogyakarta",
	"YOGYAKARTA":                 "DI Yogyakarta",
	"JOGJA":                      "DI Yogyakarta",
	"DI YOGYA":                   "DI Yogyakarta",
	"D.I. YOGYA":                 "DI Yogyakarta",
	"YOGYA":                      "DI Yogyakarta",

	// DKI Jakarta
	"DKI JAKARTA":                  "DKI Jakarta",
	"D.K.I. JAKARTA":                "DKI Jakarta",
	"DAERAH KHUSUS IBUKOTA JAKARTA": "DKI Jakarta",
	"JAKARTA":                      "DKI Jakarta",
	"DKI":                          "DKI Jakarta",

	// Pulau Jawa
	"JAWA TENGAH": "Jawa Tengah",
	"JATENG":      "Jawa Tengah",
	"JAWA TIMUR":  "Jawa Timur",
	"JATIM":       "Jawa Timur",
	"JAWA BARAT":  "Jawa Barat",
	"JABAR":       "Jawa Barat",
	"BANTEN":      "Banten",

	// Bali & Nusa Tenggara
	"BALI":                "Bali",
	"NUSA TENGGARA BARAT": "Nusa Tenggara Barat",
	"NTB":                 "Nusa Tenggara Barat",
	"NUSA TENGGARA TIMUR": "Nusa Tenggara Timur",
	"NTT":                 "Nusa Tenggara Timur",

	// Sumatera
	"SUMATERA UTARA":            "Sumatera Utara",
	"SUMUT":                     "Sumatera Utara",
	"SUMATERA BARAT":            "Sumatera Barat",
	"SUMBAR":                    "Sumatera Barat",
	"SUMATERA SELATAN":          "Sumatera Selatan",
	"SUMSEL":                    "Sumatera Selatan",
	"ACEH":                      "Aceh",
	"RIAU":                      "Riau",
	"KEPULAUAN RIAU":            "Kepulauan Riau",
	"KEPRI":                     "Kepulauan Riau",
	"JAMBI":                     "Jambi",
	"BENGKULU":                  "Bengkulu",
	"LAMPUNG":                   "Lampung",
	"KEPULAUAN BANGKA BELITUNG": "Kepulauan Bangka Belitung",
	"BANGKA BELITUNG":           "Kepulauan Bangka Belitung",
	"BABEL":                     "Kepulauan Bangka Belitung",

	// Kalimantan
	"KALIMANTAN BARAT":   "Kalimantan Barat",
	"KALBAR":             "Kalimantan Barat",
	"KALIMANTAN TENGAH":  "Kalimantan Tengah",
	"KALTENG":            "Kalimantan Tengah",
	"KALIMANTAN SELATAN": "Kalimantan Selatan",
	"KALSEL":             "Kalimantan Selatan",
	"KALIMANTAN TIMUR":   "Kalimantan Timur",
	"KALTIM":             "Kalimantan Timur",
	"KALIMANTAN UTARA":   "Kalimantan Utara",
	"KALTARA":            "Kalimantan Utara",

	// Sulawesi
	"SULAWESI UTARA":    "Sulawesi Utara",
	"SULUT":             "Sulawesi Utara",
	"SULAWESI TENGAH":   "Sulawesi Tengah",
	"SULTENG":           "Sulawesi Tengah",
	"SULAWESI SELATAN":  "Sulawesi Selatan",
	"SULSEL":            "Sulawesi Selatan",
	"SULAWESI TENGGARA": "Sulawesi Tenggara",
	"SULTRA":            "Sulawesi Tenggara",
	"GORONTALO":         "Gorontalo",
	"SULAWESI BARAT":    "Sulawesi Barat",
	"SULBAR":            "Sulawesi Barat",

	// Maluku & Papua
	"MALUKU":           "Maluku",
	"MALUKU UTARA":     "Maluku Utara",
	"PAPUA":            "Papua",
	"PAPUA BARAT":      "Papua Barat",
	"PAPUA SELATAN":    "Papua Selatan",
	"PAPUA TENGAH":     "Papua Tengah",
	"PAPUA PEGUNUNGAN": "Papua Pegunungan",
	"PAPUA BARAT DAYA": "Papua Barat Daya",
}

// NormalizeCanonicalProvince standardizes province strings to their official names.
func NormalizeCanonicalProvince(raw string) string {
	cleaned := strings.TrimSpace(cleanPrefixRegex.ReplaceAllString(raw, ""))
	cleanedUpper := strings.ToUpper(cleaned)
	if cleanedUpper == "" || cleanedUpper == "TIDAK TERDETEKSI" || cleanedUpper == "LAINNYA" {
		return ""
	}
	if canon, ok := CanonicalProvincesMap[cleanedUpper]; ok {
		return canon
	}
	for key, canon := range CanonicalProvincesMap {
		if strings.Contains(cleanedUpper, key) {
			return canon
		}
	}
	return toTitleCase(cleaned)
}

// RegencyToProvince maps base regency/city names to their official Indonesian province.
var RegencyToProvince = map[string]string{
	// Jawa Tengah
	"KARANGANYAR":  "Jawa Tengah",
	"SURAKARTA":    "Jawa Tengah",
	"SOLO":         "Jawa Tengah",
	"BOYOLALI":     "Jawa Tengah",
	"SUKOHARJO":    "Jawa Tengah",
	"KLATEN":       "Jawa Tengah",
	"SRAGEN":       "Jawa Tengah",
	"WONOGIRI":     "Jawa Tengah",
	"SEMARANG":     "Jawa Tengah",
	"SALATIGA":     "Jawa Tengah",
	"MAGELANG":     "Jawa Tengah",
	"KUDUS":        "Jawa Tengah",
	"PATI":         "Jawa Tengah",
	"JEPARA":       "Jawa Tengah",
	"DEMAK":        "Jawa Tengah",
	"KENDAL":       "Jawa Tengah",
	"BATANG":       "Jawa Tengah",
	"PEKALONGAN":   "Jawa Tengah",
	"PEMALANG":     "Jawa Tengah",
	"TEGAL":        "Jawa Tengah",
	"BREBES":       "Jawa Tengah",
	"BANYUMAS":     "Jawa Tengah",
	"CILACAP":      "Jawa Tengah",
	"PURBALINGGA":  "Jawa Tengah",
	"BANJARNEGARA": "Jawa Tengah",
	"KEBUMEN":      "Jawa Tengah",
	"PURWOREJO":    "Jawa Tengah",
	"WONOSOBO":     "Jawa Tengah",
	"TEMANGGUNG":   "Jawa Tengah",
	"GROBOGAN":     "Jawa Tengah",
	"BLORA":        "Jawa Tengah",
	"REMBANG":      "Jawa Tengah",

	// Banten
	"SERANG":            "Banten",
	"CILEGON":           "Banten",
	"TANGERANG":         "Banten",
	"TANGERANG SELATAN": "Banten",
	"LEBAK":             "Banten",
	"PANDEGLANG":        "Banten",

	// DI Yogyakarta
	"YOGYAKARTA":                 "DI Yogyakarta",
	"JOGJA":                      "DI Yogyakarta",
	"DI YOGYAKARTA":              "DI Yogyakarta",
	"D.I. YOGYAKARTA":            "DI Yogyakarta",
	"DIY":                        "DI Yogyakarta",
	"DAERAH ISTIMEWA YOGYAKARTA": "DI Yogyakarta",
	"SLEMAN":                     "DI Yogyakarta",
	"BANTUL":                     "DI Yogyakarta",
	"GUNUNGKIDUL":                "DI Yogyakarta",
	"GUNUNG KIDUL":               "DI Yogyakarta",
	"KULON PROGO":                "DI Yogyakarta",
	"KULONPROGO":                 "DI Yogyakarta",

	// Jawa Timur
	"SURABAYA":    "Jawa Timur",
	"SIDOARJO":    "Jawa Timur",
	"GRESIK":      "Jawa Timur",
	"MOJOKERTO":   "Jawa Timur",
	"JOMBANG":     "Jawa Timur",
	"BOJONEGORO":  "Jawa Timur",
	"TUBAN":       "Jawa Timur",
	"LAMONGAN":    "Jawa Timur",
	"MADIUN":      "Jawa Timur",
	"MAGETAN":     "Jawa Timur",
	"NGAWI":       "Jawa Timur",
	"PONOROGO":    "Jawa Timur",
	"PACITAN":     "Jawa Timur",
	"KEDIRI":      "Jawa Timur",
	"NGANJUK":     "Jawa Timur",
	"BLITAR":      "Jawa Timur",
	"TULUNGAGUNG": "Jawa Timur",
	"TRENGGALEK":  "Jawa Timur",
	"MALANG":      "Jawa Timur",
	"BATU":        "Jawa Timur",
	"PASURUAN":    "Jawa Timur",
	"PROBOLINGGO": "Jawa Timur",
	"LUMAJANG":    "Jawa Timur",
	"JEMBER":      "Jawa Timur",
	"BONDOWOSO":   "Jawa Timur",
	"SITUBONDO":   "Jawa Timur",
	"BANYUWANGI":  "Jawa Timur",
	"BANGKALAN":   "Jawa Timur",
	"SAMPANG":     "Jawa Timur",
	"PAMEKASAN":   "Jawa Timur",
	"SUMENEP":     "Jawa Timur",

	// Jawa Barat
	"BANDUNG":       "Jawa Barat",
	"BANDUNG BARAT": "Jawa Barat",
	"BEKASI":        "Jawa Barat",
	"BOGOR":         "Jawa Barat",
	"CIMAHI":        "Jawa Barat",
	"CIREBON":       "Jawa Barat",
	"DEPOK":         "Jawa Barat",
	"SUKABUMI":      "Jawa Barat",
	"TASIKMALAYA":   "Jawa Barat",
	"BANJAR":        "Jawa Barat",
	"CIAMIS":        "Jawa Barat",
	"CIANJUR":       "Jawa Barat",
	"GARUT":         "Jawa Barat",
	"INDRAMAYU":     "Jawa Barat",
	"KARAWANG":      "Jawa Barat",
	"KUNINGAN":      "Jawa Barat",
	"MAJALENGKA":    "Jawa Barat",
	"PANGANDARAN":   "Jawa Barat",
	"PURWAKARTA":    "Jawa Barat",
	"SUBANG":        "Jawa Barat",
	"SUMEDANG":      "Jawa Barat",

	// DKI Jakarta
	"JAKARTA PUSAT":   "DKI Jakarta",
	"JAKARTA UTARA":   "DKI Jakarta",
	"JAKARTA BARAT":   "DKI Jakarta",
	"JAKARTA SELATAN": "DKI Jakarta",
	"JAKARTA TIMUR":   "DKI Jakarta",
	"KEPULAUAN SERIBU": "DKI Jakarta",
}

// knownKotaNames lists regions that are strictly "Kota" when no prefix is specified.
var knownKotaNames = map[string]bool{
	"SURAKARTA":         true,
	"SOLO":              true,
	"SALATIGA":          true,
	"CILEGON":           true,
	"TANGERANG SELATAN": true,
	"CIMAHI":            true,
	"DEPOK":             true,
	"BANJAR":            true,
	"BATU":              true,
	"SURABAYA":          true,
	"YOGYAKARTA":        true,
	"JOGJA":             true,
	"JAKARTA PUSAT":     true,
	"JAKARTA UTARA":     true,
	"JAKARTA BARAT":     true,
	"JAKARTA SELATAN":   true,
	"JAKARTA TIMUR":     true,
	"DENPASAR":          true,
	"BANDAR LAMPUNG":    true,
	"METRO":             true,
}

// ProvinceMasterRegencies provides the standard list of regencies & cities for monitored operating provinces.
var ProvinceMasterRegencies = map[string][]string{
	"Jawa Tengah": {
		"Kabupaten Banjarnegara", "Kabupaten Banyumas", "Kabupaten Batang", "Kabupaten Blora",
		"Kabupaten Boyolali", "Kabupaten Brebes", "Kabupaten Cilacap", "Kabupaten Demak",
		"Kabupaten Grobogan", "Kabupaten Jepara", "Kabupaten Karanganyar", "Kabupaten Kebumen",
		"Kabupaten Kendal", "Kabupaten Klaten", "Kabupaten Kudus", "Kabupaten Magelang",
		"Kabupaten Pati", "Kabupaten Pekalongan", "Kabupaten Pemalang", "Kabupaten Purbalingga",
		"Kabupaten Purworejo", "Kabupaten Rembang", "Kabupaten Semarang", "Kabupaten Sragen",
		"Kabupaten Sukoharjo", "Kabupaten Tegal", "Kabupaten Temanggung", "Kabupaten Wonogiri",
		"Kabupaten Wonosobo", "Kota Magelang", "Kota Pekalongan", "Kota Salatiga",
		"Kota Semarang", "Kota Surakarta", "Kota Tegal",
	},
	"Banten": {
		"Kabupaten Lebak", "Kabupaten Pandeglang", "Kabupaten Serang", "Kabupaten Tangerang",
		"Kota Cilegon", "Kota Serang", "Kota Tangerang", "Kota Tangerang Selatan",
	},
	"DI Yogyakarta": {
		"Kabupaten Bantul", "Kabupaten Gunungkidul", "Kabupaten Kulon Progo",
		"Kabupaten Sleman", "Kota Yogyakarta",
	},
	"Jawa Timur": {
		"Kabupaten Bangkalan", "Kabupaten Banyuwangi", "Kabupaten Blitar", "Kabupaten Bojonegoro",
		"Kabupaten Bondowoso", "Kabupaten Gresik", "Kabupaten Jember", "Kabupaten Jombang",
		"Kabupaten Kediri", "Kabupaten Lamongan", "Kabupaten Lumajang", "Kabupaten Madiun",
		"Kabupaten Magetan", "Kabupaten Malang", "Kabupaten Mojokerto", "Kabupaten Nganjuk",
		"Kabupaten Ngawi", "Kabupaten Pacitan", "Kabupaten Pamekasan", "Kabupaten Pasuruan",
		"Kabupaten Ponorogo", "Kabupaten Probolinggo", "Kabupaten Sampang", "Kabupaten Sidoarjo",
		"Kabupaten Situbondo", "Kabupaten Sumenep", "Kabupaten Trenggalek", "Kabupaten Tuban",
		"Kabupaten Tulungagung", "Kota Batu", "Kota Blitar", "Kota Kediri",
		"Kota Madiun", "Kota Malang", "Kota Mojokerto", "Kota Pasuruan",
		"Kota Probolinggo", "Kota Surabaya",
	},
	"Jawa Barat": {
		"Kabupaten Bandung", "Kabupaten Bandung Barat", "Kabupaten Bekasi", "Kabupaten Bogor",
		"Kabupaten Ciamis", "Kabupaten Cianjur", "Kabupaten Cirebon", "Kabupaten Garut",
		"Kabupaten Indramayu", "Kabupaten Karawang", "Kabupaten Kuningan", "Kabupaten Majalengka",
		"Kabupaten Pangandaran", "Kabupaten Purwakarta", "Kabupaten Subang", "Kabupaten Sukabumi",
		"Kabupaten Sumedang", "Kabupaten Tasikmalaya", "Kota Bandung", "Kota Banjar",
		"Kota Bekasi", "Kota Bogor", "Kota Cimahi", "Kota Cirebon",
		"Kota Depok", "Kota Sukabumi", "Kota Tasikmalaya",
	},
	"DKI Jakarta": {
		"Kabupaten Kepulauan Seribu", "Kota Jakarta Barat", "Kota Jakarta Pusat",
		"Kota Jakarta Selatan", "Kota Jakarta Timur", "Kota Jakarta Utara",
	},
}

// KnownSubRegionsMap maps popular subdistricts/villages to their official parent Regency & Province
var KnownSubRegionsMap = map[string]struct {
	Regency  string
	District string
	Village  string
	Province string
}{
	// Desa di Ngemplak, Boyolali
	"DONOHUDAN":  {Regency: "Kabupaten Boyolali", District: "Ngemplak", Village: "Donohudan", Province: "Jawa Tengah"},
	"NGESREP":    {Regency: "Kabupaten Boyolali", District: "Ngemplak", Village: "Ngesrep", Province: "Jawa Tengah"},
	"SOBOKERTO":  {Regency: "Kabupaten Boyolali", District: "Ngemplak", Village: "Sobokerto", Province: "Jawa Tengah"},
	"GAGAKSIPAT": {Regency: "Kabupaten Boyolali", District: "Ngemplak", Village: "Gagaksipat", Province: "Jawa Tengah"},
	"DIBAL":      {Regency: "Kabupaten Boyolali", District: "Ngemplak", Village: "Dibal", Province: "Jawa Tengah"},
	"SINDON":     {Regency: "Kabupaten Boyolali", District: "Ngemplak", Village: "Sindon", Province: "Jawa Tengah"},
	"PANDEYAN":   {Regency: "Kabupaten Boyolali", District: "Ngemplak", Village: "Pandeyan", Province: "Jawa Tengah"},
	"SAWAHAN":    {Regency: "Kabupaten Boyolali", District: "Ngemplak", Village: "Sawahan", Province: "Jawa Tengah"},
	"KISMOYOSO":  {Regency: "Kabupaten Boyolali", District: "Ngemplak", Village: "Kismoyoso", Province: "Jawa Tengah"},
	"GIRIROTO":   {Regency: "Kabupaten Boyolali", District: "Ngemplak", Village: "Giriroto", Province: "Jawa Tengah"},

	// Kecamatan di Boyolali
	"NGEMPLAK":    {Regency: "Kabupaten Boyolali", District: "Ngemplak", Province: "Jawa Tengah"},
	"BANYUDONO":   {Regency: "Kabupaten Boyolali", District: "Banyudono", Province: "Jawa Tengah"},
	"SAWIT":       {Regency: "Kabupaten Boyolali", District: "Sawit", Province: "Jawa Tengah"},
	"TERAS":       {Regency: "Kabupaten Boyolali", District: "Teras", Province: "Jawa Tengah"},
	"MOJOSONGO":   {Regency: "Kabupaten Boyolali", District: "Mojosongo", Province: "Jawa Tengah"},
	"AMPEL":       {Regency: "Kabupaten Boyolali", District: "Ampel", Province: "Jawa Tengah"},
	"CEPOGO":      {Regency: "Kabupaten Boyolali", District: "Cepogo", Province: "Jawa Tengah"},
	"SELO":        {Regency: "Kabupaten Boyolali", District: "Selo", Province: "Jawa Tengah"},
	"MUSUK":       {Regency: "Kabupaten Boyolali", District: "Musuk", Province: "Jawa Tengah"},
	"ANDONG":      {Regency: "Kabupaten Boyolali", District: "Andong", Province: "Jawa Tengah"},
	"KLEGO":       {Regency: "Kabupaten Boyolali", District: "Klego", Province: "Jawa Tengah"},
	"SIMO":        {Regency: "Kabupaten Boyolali", District: "Simo", Province: "Jawa Tengah"},
	"NOGOSARI":    {Regency: "Kabupaten Boyolali", District: "Nogosari", Province: "Jawa Tengah"},
	"KARANGGEDE":  {Regency: "Kabupaten Boyolali", District: "Karanggede", Province: "Jawa Tengah"},
	"WONOSEGORO":  {Regency: "Kabupaten Boyolali", District: "Wonosegoro", Province: "Jawa Tengah"},
	"JUWANGI":     {Regency: "Kabupaten Boyolali", District: "Juwangi", Province: "Jawa Tengah"},

	// Karanganyar & Sukoharjo & Klaten
	"COLOMADU":    {Regency: "Kabupaten Karanganyar", District: "Colomadu", Province: "Jawa Tengah"},
	"GONDANGREJO": {Regency: "Kabupaten Karanganyar", District: "Gondangrejo", Province: "Jawa Tengah"},
	"JATEN":       {Regency: "Kabupaten Karanganyar", District: "Jaten", Province: "Jawa Tengah"},
	"KEBAKKRAMAT": {Regency: "Kabupaten Karanganyar", District: "Kebakkramat", Province: "Jawa Tengah"},
	"TASIKMADU":   {Regency: "Kabupaten Karanganyar", District: "Tasikmadu", Province: "Jawa Tengah"},
	"KARTASURA":   {Regency: "Kabupaten Sukoharjo", District: "Kartasura", Province: "Jawa Tengah"},
	"BAKI":        {Regency: "Kabupaten Sukoharjo", District: "Baki", Province: "Jawa Tengah"},
	"GROGOL":      {Regency: "Kabupaten Sukoharjo", District: "Grogol", Province: "Jawa Tengah"},
	"MOJOLABAN":   {Regency: "Kabupaten Sukoharjo", District: "Mojolaban", Province: "Jawa Tengah"},
	"DELANGGU":    {Regency: "Kabupaten Klaten", District: "Delanggu", Province: "Jawa Tengah"},
	"PEDAN":       {Regency: "Kabupaten Klaten", District: "Pedan", Province: "Jawa Tengah"},

	// Kecamatan Solo
	"BANJARSARI":   {Regency: "Kota Surakarta", District: "Banjarsari", Province: "Jawa Tengah"},
	"JEBRES":       {Regency: "Kota Surakarta", District: "Jebres", Province: "Jawa Tengah"},
	"LAWEYAN":      {Regency: "Kota Surakarta", District: "Laweyan", Province: "Jawa Tengah"},
	"PASAR KLIWON": {Regency: "Kota Surakarta", District: "Pasar Kliwon", Province: "Jawa Tengah"},
	"SERENGAN":     {Regency: "Kota Surakarta", District: "Serengan", Province: "Jawa Tengah"},
}

// IsValidRegency checks if the given name corresponds to an official Indonesian regency or city
func IsValidRegency(name string) bool {
	clean := strings.TrimSpace(cleanPrefixRegex.ReplaceAllString(name, ""))
	cleanUpper := strings.ToUpper(clean)
	_, exists := RegencyToProvince[cleanUpper]
	return exists
}

// NormalizeCityAndProvince parses and standardizes city and province names, preserving
// the distinction between "Kabupaten" and "Kota" (e.g. Kabupaten Serang vs Kota Serang)
// and enforcing authoritative province mapping (e.g. Karanganyar -> Jawa Tengah).
func NormalizeCityAndProvince(rawRegency, rawProvince, rawAddress string) (string, string) {
	cityInput := strings.TrimSpace(rawRegency)
	provInput := strings.TrimSpace(rawProvince)

	// Clean candidate
	cleanBase := strings.TrimSpace(cleanPrefixRegex.ReplaceAllString(cityInput, ""))
	cleanBaseUpper := strings.ToUpper(cleanBase)

	// 1. Check if cityInput is a known subdistrict/village (like Donohudan)
	if subInfo, ok := KnownSubRegionsMap[cleanBaseUpper]; ok {
		return subInfo.Regency, subInfo.Province
	}

	// 2. Check if cleanBaseUpper is a valid official Regency
	if _, ok := RegencyToProvince[cleanBaseUpper]; ok {
		// Valid regency found directly
		return formatRegencyName(cleanBaseUpper, cityInput, rawAddress), RegencyToProvince[cleanBaseUpper]
	}

	// 3. Fallback to address extraction if city input is empty, "Tidak Terdeteksi", or invalid
	if rawAddress != "" {
		upperAddr := strings.ToUpper(rawAddress)

		// A. Check for sub-region in address
		for subKey, subInfo := range KnownSubRegionsMap {
			regex := regexp.MustCompile(`\b` + regexp.QuoteMeta(subKey) + `\b`)
			if regex.MatchString(upperAddr) {
				return subInfo.Regency, subInfo.Province
			}
		}

		// B. Check for explicit prefix matches in address
		matches := regencyAddressExtractor.FindStringSubmatch(rawAddress)
		if len(matches) > 2 && strings.TrimSpace(matches[2]) != "" {
			extractedBase := strings.TrimSpace(cleanPrefixRegex.ReplaceAllString(matches[2], ""))
			extractedUpper := strings.ToUpper(extractedBase)
			if authProv, ok := RegencyToProvince[extractedUpper]; ok {
				return formatRegencyName(extractedUpper, matches[0], rawAddress), authProv
			}
		}

		// C. Check word-by-word against official regencies
		for regKey, authProv := range RegencyToProvince {
			regex := regexp.MustCompile(`\b` + regexp.QuoteMeta(regKey) + `\b`)
			if regex.MatchString(upperAddr) {
				return formatRegencyName(regKey, regKey, rawAddress), authProv
			}
		}
	}

	// 4. If provInput is provided, normalize via canonical province map
	if provInput != "" {
		canon := NormalizeCanonicalProvince(provInput)
		if canon != "" {
			return "Tidak Terdeteksi", canon
		}
	}

	return "Tidak Terdeteksi", "Jawa Tengah"
}

func formatRegencyName(baseUpper, originalInput, rawAddress string) string {
	if baseUpper == "SOLO" || baseUpper == "SURAKARTA" || baseUpper == "KOTA SOLO" || baseUpper == "KOTA SURAKARTA" {
		return "Kota Surakarta"
	}
	if baseUpper == "JOGJA" || baseUpper == "YOGYAKARTA" || baseUpper == "DIY" || baseUpper == "DI YOGYAKARTA" || baseUpper == "DAERAH ISTIMEWA YOGYAKARTA" {
		return "Kota Yogyakarta"
	}
	if baseUpper == "GUNUNG KIDUL" || baseUpper == "GUNUNGKIDUL" {
		return "Kabupaten Gunungkidul"
	}
	if baseUpper == "KULON PROGO" || baseUpper == "KULONPROGO" {
		return "Kabupaten Kulon Progo"
	}
	if baseUpper == "KEPULAUAN SERIBU" || baseUpper == "PULAU SERIBU" {
		return "Kabupaten Kepulauan Seribu"
	}
	if baseUpper == "TANGERANG SELATAN" || baseUpper == "TANGSEL" {
		return "Kota Tangerang Selatan"
	}

	isKota := false
	if knownKotaNames[baseUpper] {
		isKota = true
	} else if kotaPrefixRegex.MatchString(originalInput) {
		isKota = true
	} else if rawAddress != "" {
		upperAddr := strings.ToUpper(rawAddress)
		if strings.Contains(upperAddr, "KOTA "+baseUpper) || strings.Contains(upperAddr, "KOTA ADM. "+baseUpper) {
			isKota = true
		}
	}

	if isKota {
		return "Kota " + toTitleCase(baseUpper)
	}
	return "Kabupaten " + toTitleCase(baseUpper)
}
