package handler

import (
	"time"

	"github.com/dip-roy/go-backend/pkg/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var startTime = time.Now()

type HealthHandler struct {
	db *gorm.DB
}

func NewHealthHandler(db *gorm.DB) *HealthHandler {
	return &HealthHandler{db: db}
}

// Health godoc
// @Summary Health check
// @Tags system
// @Produce json
// @Success 200 {object} response.Response
// @Router /health [get]
func (h *HealthHandler) Health(c *gin.Context) {
	response.OK(c, gin.H{
		"status":  "ok",
		"uptime":  time.Since(startTime).String(),
		"version": "1.0.0",
	})
}

// HealthDB godoc
// @Summary Database health check
// @Tags system
// @Produce json
// @Success 200 {object} response.Response
// @Router /health/db [get]
func (h *HealthHandler) HealthDB(c *gin.Context) {
	sqlDB, err := h.db.DB()
	if err != nil {
		response.Err(c, err)
		return
	}
	if err := sqlDB.PingContext(c.Request.Context()); err != nil {
		response.Err(c, err)
		return
	}
	response.OK(c, gin.H{"status": "ok", "database": "connected"})
}
