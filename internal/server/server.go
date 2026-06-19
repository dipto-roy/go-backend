package server

import (
	"fmt"
	"strings"

	_ "github.com/dip-roy/go-backend/docs"
	"github.com/dip-roy/go-backend/internal/config"
	"github.com/dip-roy/go-backend/internal/handler"
	"github.com/dip-roy/go-backend/internal/middleware"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"
)

func New(cfg *config.Config, db *gorm.DB, h *handler.Handler, health *handler.HealthHandler) *gin.Engine {
	if cfg.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()

	// Global middleware
	r.Use(middleware.Recovery())
	r.Use(middleware.RequestID())
	r.Use(middleware.Logger())
	r.Use(middleware.CORS(getAllowedOrigins(cfg)))
	r.Use(middleware.RateLimit(cfg.RateLimit.RequestsPerSecond, cfg.RateLimit.Burst))

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Health routes (unauthenticated)
	r.GET("/health", health.Health)
	r.GET("/health/db", health.HealthDB)

	v1 := r.Group("/api/v1")

	// Auth routes
	auth := v1.Group("/auth")
	{
		auth.POST("/register", h.Register)
		auth.POST("/login", h.Login)
		auth.POST("/refresh", h.Refresh)
		auth.POST("/logout", middleware.Auth(cfg.JWT.Secret), h.Logout)
	}

	// Protected user routes
	users := v1.Group("/users", middleware.Auth(cfg.JWT.Secret))
	{
		users.GET("/me", h.GetMe)
		users.PUT("/me", h.UpdateMe)
		users.PUT("/me/password", h.ChangePassword)
		users.DELETE("/me", h.DeleteMe)
	}

	return r
}

func getAllowedOrigins(cfg *config.Config) []string {
	raw := cfg.App.Env
	if raw == "" {
		return nil
	}
	// In production you'd read CORS_ORIGINS from config
	if raw == "production" {
		origins := fmt.Sprintf("%s", "")
		if origins == "" {
			return nil
		}
		return strings.Split(origins, ",")
	}
	return nil // dev: allow all
}

func Addr(cfg *config.Config) string {
	return ":" + cfg.App.Port
}
