//go:build integration

package testutil

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/AlexeytsOY/information-exchange-service/internal/config"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

func StartPostgres() (cfg config.DB, cleanup func(), err error) {
	ctx := context.Background()

	c, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("threats_test"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		postgres.BasicWaitStrategies(),
	)
	if err != nil {
		return config.DB{}, nil, fmt.Errorf("testutil: start postgres: %w", err)
	}

	dsn, err := c.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		_ = c.Terminate(ctx)
		return config.DB{}, nil, fmt.Errorf("testutil: connection string: %w", err)
	}

	cfg, err = parseDSN(dsn)
	if err != nil {
		_ = c.Terminate(ctx)
		return config.DB{}, nil, err
	}

	cleanup = func() {
		_ = c.Terminate(context.Background())
	}
	return cfg, cleanup, nil
}

func parseDSN(dsn string) (config.DB, error) {
	u, err := url.Parse(dsn)
	if err != nil {
		return config.DB{}, fmt.Errorf("testutil: parse dsn: %w", err)
	}
	pw, _ := u.User.Password()
	return config.DB{
		Host:            u.Hostname(),
		Port:            u.Port(),
		User:            u.User.Username(),
		Password:        pw,
		Name:            strings.TrimPrefix(u.Path, "/"),
		SSLMode:         u.Query().Get("sslmode"),
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: time.Minute,
		ConnMaxIdleTime: time.Minute,
		PingTimeout:     10 * time.Second,
		LogLevel:        "silent",
	}, nil
}
