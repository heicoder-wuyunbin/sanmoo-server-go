package link

import (
	"time"

	apperr "sanmoo-server-go/internal/shared/errors"
	"sanmoo-server-go/internal/shared/validator"
)

type Link struct {
	ID          uint64    `json:"id"`
	Name        string    `json:"name"`
	Url         string    `json:"url"`
	Description string    `json:"description"`
	Icon        string    `json:"icon"`
	SortOrder   int       `json:"sortOrder"`
	IsActive    bool      `json:"isActive"`
	CreateTime  time.Time `json:"createTime"`
	UpdateTime  time.Time `json:"updateTime"`
}

func (l *Link) Validate() error {
	if err := validator.RequireNonBlank(l.Name, "name"); err != nil {
		return err
	}
	if len(l.Name) > 100 {
		return apperr.New(apperr.ErrInvalidParam.Code, "link name 长度不能超过 100")
	}
	if err := validator.RequireNonBlank(l.Url, "url"); err != nil {
		return err
	}
	if len(l.Description) > 500 {
		return apperr.New(apperr.ErrInvalidParam.Code, "link description 长度不能超过 500")
	}
	return nil
}