package role

import (
	"time"

	apperr "sanmoo-server-go/internal/shared/errors"
	"sanmoo-server-go/internal/shared/validator"
)

type Role struct {
	ID          uint64    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Status      int8      `json:"status"`
	SortOrder   int       `json:"sortOrder"`
	CreateTime  time.Time `json:"createTime"`
	UpdateTime  time.Time `json:"updateTime"`
}

type RoleWithPermissions struct {
	Role
	PermKeys []string `json:"permKeys"`
}

func (r *Role) Validate() error {
	if err := validator.RequireNonBlank(r.Name, "name"); err != nil {
		return err
	}
	if len(r.Name) > 60 {
		return apperr.New(apperr.ErrInvalidParam.Code, "name 长度不能超过 60")
	}
	return nil
}
