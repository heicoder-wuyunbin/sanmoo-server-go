package category

import (
	"time"

	apperr "sanmoo-server-go/internal/shared/errors"
	"sanmoo-server-go/internal/shared/validator"
)

type Category struct {
	ID           uint64    `json:"id"`
	Name         string    `json:"name"`
	ArticleCount int       `json:"articleCount"`
	CreateTime   time.Time `json:"createTime"`
	UpdateTime   time.Time `json:"updateTime"`
}

func (c *Category) Validate() error {
	if err := validator.RequireNonBlank(c.Name, "name"); err != nil {
		return err
	}
	if len(c.Name) > 60 {
		return apperr.New(apperr.ErrInvalidParam.Code, "category name 长度不能超过 60")
	}
	return nil
}
