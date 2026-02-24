package config

import "os"

type Config struct {
	AppEnv        string
	AdminPort     string
	WebPort       string
	DBPath        string
	SessionSecret string
	UploadDir     string
}

func Load() *Config {
	return &Config{
		AppEnv:        getEnv("APP_ENV", "development"),
		AdminPort:     getEnv("ADMIN_PORT", "18600"),
		WebPort:       getEnv("WEB_PORT", "8600"),
		DBPath:        getEnv("DB_PATH", "data/shoop.db"),
		SessionSecret: getEnv("SESSION_SECRET", "shoop-secret-key-change-in-production"),
		UploadDir:     getEnv("UPLOAD_DIR", "uploads"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
