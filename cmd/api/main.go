package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dip-roy/go-backend/internal/config"
	"github.com/dip-roy/go-backend/internal/handler"
	"github.com/dip-roy/go-backend/internal/model"
	"github.com/dip-roy/go-backend/internal/repository"
	"github.com/dip-roy/go-backend/internal/server"
	"github.com/dip-roy/go-backend/internal/service"
	"github.com/dip-roy/go-backend/pkg/logger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		// Use standard log before logger is initialized
		panic("config error: " + err.Error())
	}

	logger.Init(cfg.App.LogLevel, cfg.App.PrettyLog)
	log := logger.Get()

	db, err := setupDB(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to database")
	}

	if err := runMigrations(db); err != nil {
		log.Fatal().Err(err).Msg("failed to run migrations")
	}

	// Repositories
	userRepo := repository.NewUserRepository(db)
	refreshTokenRepo := repository.NewRefreshTokenRepository(db)

	// Services
	authSvc := service.NewAuthService(
		userRepo,
		refreshTokenRepo,
		cfg.JWT.Secret,
		cfg.JWT.AccessExpiry,
		cfg.JWT.RefreshExpiry,
	)
	userSvc := service.NewUserService(userRepo)

	// Handlers
	h := handler.New(authSvc, userSvc)
	healthH := handler.NewHealthHandler(db)

	// Server
	r := server.New(cfg, db, h, healthH)
	srv := &http.Server{
		Addr:         server.Addr(cfg),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server
	go func() {
		log.Info().Str("addr", srv.Addr).Str("env", cfg.App.Env).Msg("server starting")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("server failed")
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("shutting down server")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("forced shutdown")
	}
	log.Info().Msg("server stopped")
}

func setupDB(cfg *config.Config) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.DB.DSN), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxOpenConns(cfg.DB.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.DB.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.DB.ConnMaxLifetime)
	return db, nil
}

func runMigrations(db *gorm.DB) error {
	return db.AutoMigrate(&model.User{}, &model.RefreshToken{})
}
