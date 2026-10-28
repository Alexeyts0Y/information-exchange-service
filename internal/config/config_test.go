package config

import (
	"strings"
	"testing"
)

func setValidEnv(t *testing.T) {
	t.Helper()
	t.Setenv("DB_HOST", "db.local")
	t.Setenv("DB_PORT", "5432")
	t.Setenv("DB_USER", "u")
	t.Setenv("DB_PASSWORD", "p")
	t.Setenv("DB_NAME", "n")
	t.Setenv("DB_SSLMODE", "require")
}

func TestLoad_Defaults(t *testing.T) {
	setValidEnv(t)
	cfg, err := Load()
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if cfg.App.Port != "8080" {
		t.Errorf("port=%q", cfg.App.Port)
	}
	if cfg.App.Env != "development" {
		t.Errorf("env=%q", cfg.App.Env)
	}
	if cfg.DB.Host != "db.local" || cfg.DB.User != "u" || cfg.DB.Name != "n" {
		t.Errorf("db=%+v", cfg.DB)
	}
}

func TestLoad_OverrideEnv(t *testing.T) {
	setValidEnv(t)
	t.Setenv("APP_PORT", "9090")
	t.Setenv("APP_ENV", "prod")
	t.Setenv("DB_MAX_OPEN_CONNS", "50")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if cfg.App.Port != "9090" || cfg.App.Env != "prod" || cfg.DB.MaxOpenConns != 50 {
		t.Errorf("cfg=%+v", cfg)
	}
}

func TestLoad_MissingUser(t *testing.T) {
	setValidEnv(t)
	t.Setenv("DB_USER", "")
	if _, err := Load(); err == nil {
		t.Fatal("expected error")
	}
}

func TestLoad_MissingName(t *testing.T) {
	setValidEnv(t)
	t.Setenv("DB_NAME", "")
	if _, err := Load(); err == nil {
		t.Fatal("expected error")
	}
}

func TestGetDsn(t *testing.T) {
	db := DB{
		Host: "h", Port: "5432", User: "u", Password: "p",
		Name: "n", SSLMode: "disable",
	}
	dsn, err := db.GetDsn()
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if !strings.HasPrefix(dsn, "postgres://") {
		t.Errorf("dsn=%q", dsn)
	}
	if !strings.Contains(dsn, "sslmode=disable") {
		t.Errorf("dsn=%q", dsn)
	}
}

func TestGetDsn_Missing(t *testing.T) {
	db := DB{}
	if _, err := db.GetDsn(); err == nil {
		t.Fatal("expected error")
	}
}
