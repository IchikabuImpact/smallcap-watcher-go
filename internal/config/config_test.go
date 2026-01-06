package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadUsesConfigFileWhenPresent(t *testing.T) {
	resetEnv(t)
	workDir := t.TempDir()
	configPath := filepath.Join(workDir, "env.config")

	content := []byte("DB_HOST=filehost:3306\nDB_USER=fileuser\nDB_PASSWORD=filepass\nDB_NAME=filename\n")
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
}

func TestLoadEnvOverridesConfigFile(t *testing.T) {
	resetEnv(t)
	workDir := t.TempDir()
	configPath := filepath.Join(workDir, "env.config")

	content := []byte("DB_HOST=filehost:3306\nDB_USER=fileuser\nDB_PASSWORD=filepass\nDB_NAME=filename\n")
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
}

func resetEnv(t *testing.T) {
	t.Helper()
	keys := []string{"DB_HOST", "DB_USER", "DB_PASSWORD", "DB_NAME"}
	for _, key := range keys {
		if err := os.Unsetenv(key); err != nil {
			t.Fatalf("unset %s: %v", key, err)
		}
	}
}
