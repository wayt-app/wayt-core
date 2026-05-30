package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort  string
	AppEnv   string
	DB       DBConfig
	Auth     AuthConfig
	Email    EmailConfig
	WhatsApp WhatsAppConfig
	Supabase SupabaseConfig
	Google   GoogleOAuthConfig
}

type GoogleOAuthConfig struct {
	ClientID            string
	ClientSecret        string
	RedirectURL         string
	CustomerRedirectURL string
}

type SupabaseConfig struct {
	URL        string
	ServiceKey string
	Bucket     string
}

type EmailConfig struct {
	ResendAPIKey        string
	FromEmail           string
	AppBaseURL          string
	NotifBookingEnabled bool
}

type WhatsAppConfig struct {
	FonnteToken         string
	NotifBookingEnabled bool
}

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

type AuthConfig struct {
	JWTSecret     string
	AdminUsername string
	AdminPassword string
}

func Load() (*Config, error) {
	_ = godotenv.Load()
	return &Config{
		AppPort: getEnv("APP_PORT", "8080"),
		AppEnv:  getEnv("APP_ENV", "development"),
		DB: DBConfig{
			Host:     getEnv("DB_HOST", "127.0.0.1"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "root"),
			Password: getEnv("DB_PASSWORD", ""),
			Name:     getEnv("DB_NAME", "tabl"),
			SSLMode:  getEnv("DB_SSL_MODE", "disable"),
		},
		Auth: AuthConfig{
			JWTSecret:     requireEnv("JWT_SECRET"),
			AdminUsername: getEnv("ADMIN_USERNAME", "superadmin"),
			AdminPassword: getEnv("ADMIN_PASSWORD", ""),
		},
		Email: EmailConfig{
			ResendAPIKey:        getEnv("RESEND_API_KEY", ""),
			FromEmail:           getEnv("FROM_EMAIL", "onboarding@resend.dev"),
			AppBaseURL:          getEnv("APP_BASE_URL", "http://localhost:8080"),
			NotifBookingEnabled: getEnv("EMAIL_NOTIF_BOOKING", "true") == "true",
		},
		WhatsApp: WhatsAppConfig{
			FonnteToken:         getEnv("FONNTE_TOKEN", ""),
			NotifBookingEnabled: getEnv("WA_NOTIF_BOOKING", "true") == "true",
		},
		Supabase: SupabaseConfig{
			URL:        getEnv("SUPABASE_URL", ""),
			ServiceKey: getEnv("SUPABASE_SERVICE_KEY", ""),
			Bucket:     getEnv("SUPABASE_BUCKET", "tabl-media"),
		},
		Google: GoogleOAuthConfig{
			ClientID:            getEnv("GOOGLE_CLIENT_ID", ""),
			ClientSecret:        getEnv("GOOGLE_CLIENT_SECRET", ""),
			RedirectURL:         getEnv("GOOGLE_REDIRECT_URL", ""),
			CustomerRedirectURL: getEnv("GOOGLE_CUSTOMER_REDIRECT_URL", ""),
		},
	}, nil
}

func (c *DBConfig) DSN() string {
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=UTC",
		c.Host, c.User, c.Password, c.Name, c.Port, c.SSLMode)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// requireEnv returns the env var value or exits with a fatal log if not set.
func requireEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("FATAL: required environment variable %s is not set", key)
	}
	return v
}
