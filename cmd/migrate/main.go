package main

import (
	"context"
	"log"

	"github.com/AlexeytsOY/information-exchange-service/internal/config"
	"github.com/AlexeytsOY/information-exchange-service/internal/database"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("migrate: %v", err)
	}
	log.Println("migrate: done")
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	ctx := context.Background()
	db, err := database.NewPostgres(ctx, cfg.DB)
	if err != nil {
		return err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	defer func() { _ = sqlDB.Close() }()

	return database.Migrate(db)
}
