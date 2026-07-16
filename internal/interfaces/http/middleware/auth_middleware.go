package middleware

import (
	"strings"

	"sanmoo-server-go/internal/infrastructure/repository/mysqlrepo"
	"sanmoo-server-go/internal/infrastructure/security"
	apperr "sanmoo-server-go/internal/shared/errors"
	"sanmoo-server-go/internal/shared/response"

	"github.com/gin-gonic/gin"
)

const (
	CtxUserIDKey   = "ctx_user_id"
	CtxUsernameKey = "ctx_username"
)

func JWTAuth(jwtMgr *security.JWTManager, repo *mysqlrepo.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if auth == "" || !strings.HasPrefix(strings.ToLower(auth), "bearer ") {
			response.Fail(c, apperr.ErrUnauthorized)
			c.Abort()
			return
		}
		token := strings.TrimSpace(auth[7:])
		claims, err := jwtMgr.ParseAccessToken(token)
		if err != nil {
			response.Fail(c, apperr.ErrUnauthorized)
			c.Abort()
			return
		}
		u, err := repo.FindByIDUser(c.Request.Context(), claims.UserID)
		if err != nil {
			response.Fail(c, apperr.ErrUnauthorized)
			c.Abort()
			return
		}
		if strings.ToUpper(u.Status) != "ENABLED" {
			response.Fail(c, apperr.ErrUserDisabled)
			c.Abort()
			return
		}
		c.Set(CtxUserIDKey, claims.UserID)
		c.Set(CtxUsernameKey, claims.Username)
		c.Next()
	}
}
