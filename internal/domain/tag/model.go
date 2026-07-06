package tag

import (
	"time"

	apperr "sanmoo-server-go/internal/shared/errors"
	"sanmoo-server-go/internal/shared/validator"
)

type Tag struct {
	ID           uint64    `json:"id"`
	Name         string    `json:"name"`
	ArticleCount int       `json:"articleCount"`
	CreateTime   time.Time `json:"createTime"`
	UpdateTime   time.Time `json:"updateTime"`
}

func (t *Tag) Validate() error {
	if err := validator.RequireNonBlank(t.Name, "name"); err != nil {
		return err
	}
	if len(t.Name) > 60 {
		return apperr.New(apperr.ErrInvalidParam.Code, "tag name 长度不能超过 60")
	}
	return nil
}
