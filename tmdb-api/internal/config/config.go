// internal/config/config.go
package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	TMDBApiKey     string
	ServerPort     string
	DatabaseURL    string
	RedisAddr      string
	JWTSecret      string // NEW: signs and verifies JWT tokens
	JWTExpiryHours int    // NEW: how long tokens are valid
	AdminEmail     string
	AdminPassword  string
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, reading from environment directly")
	} else {
		log.Println("✅ Successfully loaded .env file!")
	}

	cfg := &Config{
		TMDBApiKey:     getEnv("TMDB_API_KEY", ""),
		ServerPort:     getEnv("SERVER_PORT", "8080"),
		DatabaseURL:    getEnv("DATABASE_URL", ""),
		RedisAddr:      getEnv("REDIS_ADDR", "localhost:6379"),
		JWTSecret:      getEnv("JWT_SECRET", ""),
		JWTExpiryHours: getEnvAsInt("JWT_EXPIRY_HOURS", 24),
		AdminEmail:     getEnv("ADMIN_EMAIL", ""),
		AdminPassword:  getEnv("ADMIN_PASSWORD", ""),
	}

	if cfg.TMDBApiKey == "" {
		log.Fatal("FATAL: TMDB_API_KEY is not set")
	}
	if cfg.JWTSecret == "" {
		log.Fatal("FATAL: JWT_SECRET is not set")
	}
	// Enforce a minimum secret length — short secrets are brute-forceable.
	if len(cfg.JWTSecret) < 32 {
		log.Fatal("FATAL: JWT_SECRET must be at least 32 characters")
	}
	// Admin credentials are optional — the app works without them.
	// But if only one is set, that's a configuration mistake we catch early.
	if (cfg.AdminEmail == "") != (cfg.AdminPassword == "") {
		log.Fatal("FATAL: both ADMIN_EMAIL and ADMIN_PASSWORD must be set together")
	}

	return cfg
}

func getEnv(key, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return defaultVal
}

func getEnvAsInt(key string, defaultVal int) int {
	if val, ok := os.LookupEnv(key); ok {
		var i int
		if _, err := fmt.Sscan(val, &i); err == nil {
			return i
		}
	}
	return defaultVal
}
