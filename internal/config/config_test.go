package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadUsesConfigFileWhenPresent(t *testing.T) {
	resetEnv(t)
	workDir := t.TempDir()
	configPath := filepath.Join(workDir, "env.config")

	content := []byte("DB_HOST=filehost:3306\nDB_USER=fileuser\nDB_PASSWORD=filepass\nDB_NAME=filename\nSCRAPER_BASE_URL=http://file-scraper:8082\nSCRAPER_REQUEST_INTERVAL=4s\nOUTPUT_DIR=file-output\nINDEX_MAX_AGE=48h\n")
	if err := os.WriteFile(configPath, content, 0o644); err != nil {
		t.Fatalf("write env.config: %v", err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	defer func() {
		_ = os.Chdir(cwd)
	}()

	if err := os.Chdir(workDir); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	cfg := Load()

	if cfg.DBHost != "filehost:3306" {
		t.Fatalf("DBHost = %q", cfg.DBHost)
	}
	if cfg.DBUser != "fileuser" {
		t.Fatalf("DBUser = %q", cfg.DBUser)
	}
	if cfg.DBPassword != "filepass" {
		t.Fatalf("DBPassword = %q", cfg.DBPassword)
	}
	if cfg.DBName != "filename" {
		t.Fatalf("DBName = %q", cfg.DBName)
	}
	if cfg.ScraperBaseURL != "http://file-scraper:8082" {
		t.Fatalf("ScraperBaseURL = %q", cfg.ScraperBaseURL)
	}
	if cfg.ScraperRequestInterval != 4*time.Second {
		t.Fatalf("ScraperRequestInterval = %s", cfg.ScraperRequestInterval)
	}
	if cfg.OutputDir != "file-output" {
		t.Fatalf("OutputDir = %q", cfg.OutputDir)
	}
	if cfg.IndexMaxAge != 48*time.Hour {
		t.Fatalf("IndexMaxAge = %s", cfg.IndexMaxAge)
	}
}

func TestLoadEnvOverridesConfigFile(t *testing.T) {
	resetEnv(t)
	workDir := t.TempDir()
	configPath := filepath.Join(workDir, "env.config")

	content := []byte("DB_HOST=filehost:3306\nDB_USER=fileuser\nDB_PASSWORD=filepass\nDB_NAME=filename\nSCRAPER_BASE_URL=http://file-scraper:8082\nSCRAPER_REQUEST_INTERVAL=4s\nOUTPUT_DIR=file-output\nINDEX_MAX_AGE=48h\n")
	if err := os.WriteFile(configPath, content, 0o644); err != nil {
		t.Fatalf("write env.config: %v", err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	defer func() {
		_ = os.Chdir(cwd)
	}()

	if err := os.Chdir(workDir); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	t.Setenv("DB_HOST", "envhost:3306")
	t.Setenv("DB_USER", "envuser")
	t.Setenv("DB_PASSWORD", "envpass")
	t.Setenv("DB_NAME", "envname")
	t.Setenv("SCRAPER_BASE_URL", "http://env-scraper:8082")
	t.Setenv("SCRAPER_REQUEST_INTERVAL", "5s")
	t.Setenv("OUTPUT_DIR", "env-output")
	t.Setenv("INDEX_MAX_AGE", "72h")

	cfg := Load()

	if cfg.DBHost != "envhost:3306" {
		t.Fatalf("DBHost = %q", cfg.DBHost)
	}
	if cfg.DBUser != "envuser" {
		t.Fatalf("DBUser = %q", cfg.DBUser)
	}
	if cfg.DBPassword != "envpass" {
		t.Fatalf("DBPassword = %q", cfg.DBPassword)
	}
	if cfg.DBName != "envname" {
		t.Fatalf("DBName = %q", cfg.DBName)
	}
	if cfg.ScraperBaseURL != "http://env-scraper:8082" {
		t.Fatalf("ScraperBaseURL = %q", cfg.ScraperBaseURL)
	}
	if cfg.ScraperRequestInterval != 5*time.Second {
		t.Fatalf("ScraperRequestInterval = %s", cfg.ScraperRequestInterval)
	}
	if cfg.OutputDir != "env-output" {
		t.Fatalf("OutputDir = %q", cfg.OutputDir)
	}
	if cfg.IndexMaxAge != 72*time.Hour {
		t.Fatalf("IndexMaxAge = %s", cfg.IndexMaxAge)
	}
}

func TestLoadDefaultsWithoutConfigFile(t *testing.T) {
	resetEnv(t)
	workDir := t.TempDir()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	defer func() {
		_ = os.Chdir(cwd)
	}()

	if err := os.Chdir(workDir); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	cfg := Load()

	if cfg.DBHost != "localhost" {
		t.Fatalf("DBHost = %q", cfg.DBHost)
	}
	if cfg.DBUser != "jpx_user" {
		t.Fatalf("DBUser = %q", cfg.DBUser)
	}
	if cfg.DBPassword != "jpx_password" {
		t.Fatalf("DBPassword = %q", cfg.DBPassword)
	}
	if cfg.DBName != "jpx_data" {
		t.Fatalf("DBName = %q", cfg.DBName)
	}
	if cfg.ScraperBaseURL != "http://host.docker.internal:8085" {
		t.Fatalf("ScraperBaseURL = %q", cfg.ScraperBaseURL)
	}
	if cfg.ScraperRequestInterval != 3*time.Second {
		t.Fatalf("ScraperRequestInterval = %s", cfg.ScraperRequestInterval)
	}
	if cfg.OutputDir != "output" {
		t.Fatalf("OutputDir = %q", cfg.OutputDir)
	}
	if cfg.IndexMaxAge != 36*time.Hour {
		t.Fatalf("IndexMaxAge = %s", cfg.IndexMaxAge)
	}
}

func resetEnv(t *testing.T) {
	t.Helper()
	keys := []string{"DB_HOST", "DB_USER", "DB_PASSWORD", "DB_NAME", "SCRAPER_BASE_URL", "SCRAPER_REQUEST_INTERVAL", "OUTPUT_DIR", "INDEX_MAX_AGE"}
	for _, key := range keys {
		if err := os.Unsetenv(key); err != nil {
			t.Fatalf("unset %s: %v", key, err)
		}
	}
}
