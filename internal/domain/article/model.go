package article

import (
	"time"

	apperr "sanmoo-server-go/internal/shared/errors"
	"sanmoo-server-go/internal/shared/validator"
)

type TagRef struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
}

type TopicRef struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
}

type Article struct {
	ID          uint64     `json:"id"`
	Title       string     `json:"title"`
	TitleImage  string     `json:"titleImage"`
	Description string     `json:"description"`
	Content     string     `json:"content"`
	ReadNum     int        `json:"readNum"`
	ShareNum    int        `json:"shareNum"`
	LikeNum     int        `json:"likeNum"`
	IsTop       bool       `json:"isTop"`
	IsPublished bool       `json:"isPublished"`
	CategoryID  uint64     `json:"categoryId"`
	Category    string     `json:"categoryName"`
	Tags        []TagRef   `json:"tags"`
	Topics      []TopicRef `json:"topics"`
	CreateTime  time.Time  `json:"createTime"`
	UpdateTime  time.Time  `json:"updateTime"`
}

func (a *Article) Validate() error {
	if err := validator.RequireNonBlank(a.Title, "title"); err != nil {
		return err
	}
	if err := validator.RequireNonBlank(a.Description, "description"); err != nil {
		return err
	}
	if err := validator.RequireNonBlank(a.Content, "content"); err != nil {
		return err
	}
	if a.CategoryID == 0 {
		return apperr.New(apperr.ErrInvalidParam.Code, "categoryId 不能为空")
	}
	if len(a.Tags) == 0 {
		return apperr.New(apperr.ErrInvalidParam.Code, "tagIds 不能为空")
	}
	return nil
}
