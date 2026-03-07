package shared

import (
	"os"
	"strconv"
)

// Config uygulama yapılandırması
type Config struct {
	JWTAccessSecret  string
	JWTRefreshSecret string
	AccessTokenTTL   int // seconds
	RefreshTokenTTL  int // seconds
	ServerPort       string
	DatabasePath     string
	Environment      string // development, production, etc.
}

// GetConfig uygulama yapılandırmasını döner
func GetConfig() *Config {
	config := &Config{
		JWTAccessSecret:  getEnv("JWT_ACCESS_SECRET", "super-secret-access-key"),
		JWTRefreshSecret: getEnv("JWT_REFRESH_SECRET", "super-secret-refresh-key"),
		AccessTokenTTL:   getEnvInt("ACCESS_TOKEN_TTL", 900),    // 15 minutes
		RefreshTokenTTL:  getEnvInt("REFRESH_TOKEN_TTL", 604800), // 7 days
		ServerPort:       getEnv("SERVER_PORT", ":8081"),
		DatabasePath:     getEnv("DB_PATH", "people.db"),
		Environment:      getEnv("ENVIRONMENT", "development"),
	}

	// Production ortamında güvenlik kontrolleri
	if config.Environment == "production" {
		if config.JWTAccessSecret == "super-secret-access-key" || 
		   config.JWTRefreshSecret == "super-secret-refresh-key" {
			panic("Production ortamında JWT secret'ları güvenli olmalıdır!")
		}
	}

	return config
}

// getEnv environment variable'ı okur, yoksa default değeri döner
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt environment variable'ı integer olarak okur
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
