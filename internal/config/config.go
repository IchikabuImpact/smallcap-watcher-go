package config

import "os"

type Config struct {
	DBHost     string
	DBUser     string
	DBPassword string
	DBName     string
}

func Load() Config {
	return Config{
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBUser:     getEnv("DB_USER", "jpx_user"),
		DBPassword: getEnv("DB_PASSWORD", "jpx_password"),
		DBName:     getEnv("DB_NAME", "jpx_data"),
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
