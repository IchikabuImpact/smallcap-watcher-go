package config

import (
	"bufio"
	"os"
	"strings"
)

type Config struct {
	DBHost         string
	DBUser         string
	DBPassword     string
	DBName         string
	ScraperBaseURL string
}

func Load() Config {
	loadConfigFile("env.config")
	return Config{
		DBHost:         getEnv("DB_HOST", "localhost"),
		DBUser:         getEnv("DB_USER", "jpx_user"),
		DBPassword:     getEnv("DB_PASSWORD", "jpx_password"),
		DBName:         getEnv("DB_NAME", "jpx_data"),
		ScraperBaseURL: getEnv("SCRAPER_BASE_URL", "http://127.0.0.1:8085"),
	}
}

func loadConfigFile(path string) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if key == "" {
			continue
		}
		if _, exists := os.LookupEnv(key); !exists {
			_ = os.Setenv(key, value)
		}
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
