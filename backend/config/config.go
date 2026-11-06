package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost               string
	DBUser               string
	DBPassword           string
	DBName               string
	DBPort               string
	ServerPort           string
	JWTSecret            string
	SMTPHost             string
	SMTPPort             string
	SMTPUser             string
	SMTPPass             string
	SMTPFrom             string
	MidtransServerKey    string
	MidtransIsProduction bool
	// Branding App MySQL Connection
	BrandingDBHost     string
	BrandingDBUser     string
	BrandingDBPassword string
	BrandingDBName     string
	BrandingDBPort     string
	GeminiAPIKey       string
	GeminiModel        string
	MacMiniLLMURL      string
	MacMiniLLMModel    string
	MacMiniLLMAPIKey   string
	MacMiniLLMTimeout  string
	// WhatsApp Auto Report (diakses oleh WA service di Mac Mini)
	WASessionDB     string
	WAReportGroupID string
	InternalAPIKey  string
	// Skip GORM AutoMigrate on startup (rekomendasi untuk produksi / zero table lock)
	SkipAutoMigrate bool
}

func LoadConfig() Config {
	err := godotenv.Load(".env")
	if err != nil {
		err = godotenv.Load("../../.env")
	}

	if err != nil {
		log.Println("No .env file found or error loading it, looking at environment variables")
	} else {
		log.Println("Loaded configuration from .env file")
	}

	return Config{
		DBHost:               getEnv("DB_HOST", "localhost"),
		DBUser:               getEnv("DB_USER", "root"),
		DBPassword:           getEnv("DB_PASSWORD", ""),
		DBName:               getEnv("DB_NAME", "hris_db"),
		DBPort:               getEnv("DB_PORT", "5321"),
		ServerPort:           getEnv("SERVER_PORT", "8080"),
		JWTSecret:            getEnv("JWT_SECRET", ""),
		SMTPHost:             getEnv("SMTP_HOST", "smtp.gmail.com"),
		SMTPPort:             getEnv("SMTP_PORT", "587"),
		SMTPUser:             getEnv("SMTP_USER", ""),
		SMTPPass:             getEnv("SMTP_PASS", ""),
		SMTPFrom:             getEnv("SMTP_FROM", ""),
		MidtransServerKey:    getEnv("MIDTRANS_SERVER_KEY", ""),
		MidtransIsProduction: getEnv("MIDTRANS_IS_PRODUCTION", "false") == "true",
		// Branding App MySQL — read from env
		BrandingDBHost:     getEnv("BRANDING_DB_HOST", "localhost"),
		BrandingDBUser:     getEnv("BRANDING_DB_USER", ""),
		BrandingDBPassword: getEnv("BRANDING_DB_PASSWORD", ""),
		BrandingDBName:     getEnv("BRANDING_DB_NAME", ""),
		BrandingDBPort:     getEnv("BRANDING_DB_PORT", "3306"),
		GeminiAPIKey:       getEnv("GEMINI_API_KEY", ""),
		GeminiModel:        getEnv("GEMINI_MODEL", "gemini-2.5-flash"),
		MacMiniLLMURL:      getEnv("MAC_MINI_LLM_URL", ""),
		MacMiniLLMModel:    getEnv("MAC_MINI_LLM_MODEL", "hermes-combo"),
		MacMiniLLMAPIKey:   getEnv("MAC_MINI_LLM_API_KEY", ""),
		MacMiniLLMTimeout:  getEnv("MAC_MINI_LLM_TIMEOUT", "25s"),
		// WA Report Service (Mac Mini)
		WASessionDB:     getEnv("WA_SESSION_DB", "./wa_session.db"),
		WAReportGroupID: getEnv("WA_REPORT_GROUP_ID", ""),
		InternalAPIKey:  getEnv("INTERNAL_API_KEY", ""),
		SkipAutoMigrate: getEnv("SKIP_AUTO_MIGRATE", "false") == "true",
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

// GetEnvOrDefault is the exported version of getEnv for use outside the config package.
func GetEnvOrDefault(key, fallback string) string {
	return getEnv(key, fallback)
}
