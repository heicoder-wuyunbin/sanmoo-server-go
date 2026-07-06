package middleware

import (
	rolesvc "sanmoo-server-go/internal/application/role"
	apperr "sanmoo-server-go/internal/shared/errors"
	"sanmoo-server-go/internal/shared/response"

	"github.com/gin-gonic/gin"
)

const (
	CtxRolesKey    = "ctx_roles"
	CtxPermSetKey  = "ctx_perm_set"
	CtxIsAdminKey  = "ctx_is_admin"
)

func RequirePerm(roleSvc *rolesvc.Service, permKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := c.Get(CtxUserIDKey)
		if !ok {
			response.Fail(c, apperr.ErrUnauthorized)
			c.Abort()
			return
		}
		uid := userID.(uint64)

		roles, err := roleSvc.GetUserRoles(c.Request.Context(), uid)
		if err != nil {
			response.Fail(c, apperr.ErrInternal)
			c.Abort()
			return
		}
		c.Set(CtxRolesKey, roles)

		isAdmin := false
		for _, r := range roles {
			if rolesvc.IsAdminRoleName(r.Name) {
				isAdmin = true
				break
			}
		}
		c.Set(CtxIsAdminKey, isAdmin)

		if isAdmin {
			c.Next()
			return
		}

		has, err := roleSvc.HasPermission(c.Request.Context(), uid, permKey)
		if err != nil {
			response.Fail(c, apperr.ErrInternal)
			c.Abort()
			return
		}
		if !has {
			response.Fail(c, apperr.ErrForbidden)
			c.Abort()
			return
		}
		c.Next()
	}
}
