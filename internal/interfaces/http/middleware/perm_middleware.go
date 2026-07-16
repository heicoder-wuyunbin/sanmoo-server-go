package middleware

import (
	"sanmoo-server-go/internal/infrastructure/repository/mysqlrepo"
	apperr "sanmoo-server-go/internal/shared/errors"
	"sanmoo-server-go/internal/shared/response"

	"github.com/gin-gonic/gin"
)

const (
	CtxRolesKey   = "ctx_roles"
	CtxPermSetKey = "ctx_perm_set"
	CtxIsAdminKey = "ctx_is_admin"
)

// RequireAdmin ensures the user is an admin.
func RequireAdmin(repo *mysqlrepo.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := c.Get(CtxUserIDKey)
		if !ok {
			response.Fail(c, apperr.ErrUnauthorized)
			c.Abort()
			return
		}
		uid := userID.(uint64)
		u, err := repo.FindByIDUser(c.Request.Context(), uid)
		if err != nil {
			response.Fail(c, apperr.ErrUnauthorized)
			c.Abort()
			return
		}
		if !u.IsAdmin {
			response.Fail(c, apperr.ErrForbidden)
			c.Abort()
			return
		}
		c.Set(CtxIsAdminKey, true)
		c.Next()
	}
}