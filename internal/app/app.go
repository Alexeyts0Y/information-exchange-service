package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/AlexeytsOY/information-exchange-service/internal/config"
	"github.com/AlexeytsOY/information-exchange-service/internal/database"
	threatapi "github.com/AlexeytsOY/information-exchange-service/internal/gen/threat"
	"github.com/AlexeytsOY/information-exchange-service/internal/handler"
	"github.com/AlexeytsOY/information-exchange-service/internal/middleware"
	"github.com/AlexeytsOY/information-exchange-service/internal/repository"
	"github.com/AlexeytsOY/information-exchange-service/internal/service"
	"github.com/gin-gonic/gin"
	ginopenapiui "github.com/openapi-ui/go-openapi-ui/pkg/middleware/gin"
	"gorm.io/gorm"
)

type App struct {
	cfg *config.Config
}

func New(cfg *config.Config) *App {
	return &App{cfg: cfg}
}

func (app *App) Run() error {
	db, err := database.New(context.Background(), app.cfg.DB)
	if err != nil {
		return fmt.Errorf("app: database: %w", err)
	}
	defer func() {
		if sqlDB, err := db.DB(); err == nil {
			_ = sqlDB.Close()
		}
	}()

	if err := database.Migrate(db); err != nil {
		return fmt.Errorf("app: migrate: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("app: sql.DB: %w", err)
	}

	router := gin.New()
	app.setupMiddlewares(router)
	app.registerHandlers(router, db, sqlDB)

	srv := &http.Server{
		Addr:         ":" + app.cfg.App.Port,
		Handler:      router,
		ReadTimeout:  app.cfg.App.ReadTimeout,
		WriteTimeout: app.cfg.App.WriteTimeout,
		IdleTimeout:  app.cfg.App.IdleTimeout,
	}

	go app.gracefulShutdown(srv)

	log.Printf("app: listening on %s", srv.Addr)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("app: server: %w", err)
	}
	return nil
}

func (app *App) gracefulShutdown(srv *http.Server) {
	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt, syscall.SIGTERM)

	<-exit
	log.Println("app: shutdown signal received")

	timeoutCtx, cancel := context.WithTimeout(context.Background(), app.cfg.App.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(timeoutCtx); err != nil {
		log.Printf("app: graceful shutdown failed: %v", err)
	}
}

func (app *App) setupMiddlewares(router *gin.Engine) {
	router.Use(middleware.Logger(), gin.Recovery())
	if err := router.SetTrustedProxies(nil); err != nil {
		log.Printf("app: trusted proxies: %v", err)
	}
}

func (app *App) registerHandlers(router *gin.Engine, db *gorm.DB, sqlDB *sql.DB) {
	threatRepository := repository.NewThreatRepository(db)
	threatService := service.NewThreatService(threatRepository)
	threatHandler := handler.NewThreatHandler(threatService)

	threatapi.RegisterHandlersWithOptions(router, threatHandler, threatapi.GinServerOptions{
		ErrorHandler: handler.ErrorHandler,
	})

	router.GET("/health", handler.NewHealthHandler(sqlDB, app.cfg.App.HealthTimeout).IsHealthy)

	router.Use(ginopenapiui.New(app.cfg.Doc))
}
