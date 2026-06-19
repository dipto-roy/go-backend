package middleware

import (
	"net/http"

	"github.com/dip-roy/go-backend/pkg/logger"
	"github.com/dip-roy/go-backend/pkg/response"
	"github.com/gin-gonic/gin"
)

func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				logger.Error().
					Interface("panic", r).
					Str("path", c.Request.URL.Path).
					Msg("panic recovered")
				c.AbortWithStatusJSON(http.StatusInternalServerError, response.Response{
					Success: false,
					Error:   &response.ErrorBody{Code: "INTERNAL_ERROR", Message: "internal server error"},
				})
			}
		}()
		c.Next()
	}
}
