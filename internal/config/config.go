package config

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	App App
	DB  DB
}

type App struct {
	Env             string
	Port            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
	HealthTimeout   time.Duration
}

type DB struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string

	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
	PingTimeout     time.Duration
	LogLevel        string
}

func (d DB) GetDsn() (string, error) {
	var missing []string
	if d.Host == "" {
		missing = append(missing, "DB_HOST")
	}
	if d.User == "" {
		missing = append(missing, "DB_USER")
	}
	if d.Name == "" {
		missing = append(missing, "DB_NAME")
	}
	if len(missing) > 0 {
		return "", fmt.Errorf("config: missing: %v", missing)
	}

	u := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(d.User, d.Password),
		Host:   net.JoinHostPort(d.Host, d.Port),
		Path:   d.Name,
	}
	q := u.Query()
	q.Set("sslmode", d.SSLMode)
	u.RawQuery = q.Encode()
	return u.String(), nil
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		App: App{
			Env:             getenv("APP_ENV", "development"),
			Port:            getenv("APP_PORT", "8080"),
			ReadTimeout:     mustDuration("APP_READ_TIMEOUT", 5*time.Second),
			WriteTimeout:    mustDuration("APP_WRITE_TIMEOUT", 10*time.Second),
			IdleTimeout:     mustDuration("APP_IDLE_TIMEOUT", 60*time.Second),
			ShutdownTimeout: mustDuration("APP_SHUTDOWN_TIMEOUT", 5*time.Second),
			HealthTimeout:   mustDuration("APP_HEALTH_TIMEOUT", 2*time.Second),
		},
		DB: DB{
			Host:            getenv("DB_HOST", "localhost"),
			Port:            getenv("DB_PORT", "5432"),
			User:            getenv("DB_USER", ""),
			Password:        getenv("DB_PASSWORD", ""),
			Name:            getenv("DB_NAME", ""),
			SSLMode:         getenv("DB_SSLMODE", "disable"),
			MaxOpenConns:    mustInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    mustInt("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: mustDuration("DB_CONN_MAX_LIFETIME", 30*time.Minute),
			ConnMaxIdleTime: mustDuration("DB_CONN_MAX_IDLE_TIME", 5*time.Minute),
			PingTimeout:     mustDuration("DB_PING_TIMEOUT", 5*time.Second),
			LogLevel:        getenv("DB_LOG_LEVEL", "warn"),
		},
	}

	if _, err := cfg.DB.GetDsn(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func getenv(key, def string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return def
}

func mustInt(key string, def int) int {
	if v, ok := os.LookupEnv(key); !ok || v == "" {
		return def
	} else if n, err := strconv.Atoi(v); err == nil {
		return n
	}
	return def
}

func mustDuration(key string, def time.Duration) time.Duration {
	if v, ok := os.LookupEnv(key); !ok || v == "" {
		return def
	} else if d, err := time.ParseDuration(v); err == nil {
		return d
	}
	return def
}
