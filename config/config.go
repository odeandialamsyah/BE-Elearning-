package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort      string
	DBUser       string
	DBPass       string
	DBHost       string
	DBPort       string
	DBName       string
	PasetoSecret string
}

func LoadConfig() *Config {
	// Load .env hanya untuk local development
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️ No .env file found, using system env")
	}

	return &Config{
		AppPort:      getEnv("PORT", "8080"),

		// MySQL (Railway / Docker)
		DBUser: getEnv("DB_USER", "root"),
		DBPass: getEnv("DB_PASS", ""),
		DBHost: getEnv("DB_HOST", "127.0.0.1"),
		DBPort: getEnv("DB_PORT", "3306"),
		DBName: getEnv("DB_NAME", "elearning_db"),

		PasetoSecret: getEnv("PASETO_SECRET_KEY", ""),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}
	return fallback
}
