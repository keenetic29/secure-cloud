package config

import (
	"log"
	"os"
	
	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort string
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	JWTSecret  string
	YandexDisk YandexDiskConfig
}

type YandexDiskConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
}

func Load() *Config {
	// Загружаем .env файл
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}
	
	return &Config{
		ServerPort: getEnv("SERVER_PORT", "8080"),
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "password"),
		DBName:     getEnv("DB_NAME", "cloud_storage"),
		JWTSecret:  getEnv("JWT_SECRET", "fallback-secret-key"),
		YandexDisk: YandexDiskConfig{
			ClientID:     getEnv("YANDEX_DISK_CLIENT_ID", ""),
			ClientSecret: getEnv("YANDEX_DISK_CLIENT_SECRET", ""),
			RedirectURI:  getEnv("YANDEX_DISK_REDIRECT_URI", "http://localhost:8080/api/v1/storage/yandex/callback"),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}