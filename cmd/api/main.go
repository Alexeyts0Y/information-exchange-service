package main

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"

	"github.com/AlexeytsOY/information-exchange-service/internal/config"
	"github.com/AlexeytsOY/information-exchange-service/internal/database"
	"github.com/AlexeytsOY/information-exchange-service/internal/handler"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("app: %v", err)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	db, err := database.NewPostgres(context.Background(), cfg.DB)
	if err != nil {
		return err
	}

	if err := database.Migrate(db); err != nil {
		return err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	router := gin.Default()
	router.GET("/health", handler.NewHealthHandler(sqlDB).IsHealthy)

	return router.Run(":8080")
}
