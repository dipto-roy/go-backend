package middleware

import (
	"strings"

	"github.com/dip-roy/go-backend/pkg/apperror"
	"github.com/dip-roy/go-backend/pkg/response"
	"github.com/dip-roy/go-backend/pkg/token"
	"github.com/gin-gonic/gin"
)

const (
	AuthUserIDKey = "auth_user_id"
	AuthEmailKey  = "auth_email"
)

func Auth(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			response.Err(c, apperror.ErrUnauthorized)
			c.Abort()
			return
		}

		tokenStr := strings.TrimPrefix(header, "Bearer ")
		claims, err := token.Verify(tokenStr, jwtSecret)
		if err != nil {
			if err == token.ErrExpiredToken {
				response.Err(c, apperror.ErrTokenExpired)
			} else {
				response.Err(c, apperror.ErrTokenInvalid)
			}
			c.Abort()
			return
		}

		c.Set(AuthUserIDKey, claims.UserID)
		c.Set(AuthEmailKey, claims.Email)
		c.Next()
	}
}
