package permission

import (
	"time"

	apperr "sanmoo-server-go/internal/shared/errors"
	"sanmoo-server-go/internal/shared/validator"
)

type Permission struct {
	ID          uint64    `json:"id"`
	PermKey     string    `json:"permKey"`
	Name        string    `json:"name"`
	Module      string    `json:"module"`
	Type        string    `json:"type"`
	Description string    `json:"description"`
	FrontPath   string    `json:"frontPath"` // 前端菜单路径
	Icon        string    `json:"icon"`        // 前端菜单图标
	SortOrder   int       `json:"sortOrder"`
	Status      int8      `json:"status"`
	CreateTime  time.Time `json:"createTime"`
	UpdateTime  time.Time `json:"updateTime"`
}

func (p *Permission) Validate() error {
	if err := validator.RequireNonBlank(p.PermKey, "permKey"); err != nil {
		return err
	}
	if len(p.PermKey) > 128 {
		return apperr.New(apperr.ErrInvalidParam.Code, "permKey 长度不能超过 128")
	}
	if err := validator.RequireNonBlank(p.Name, "name"); err != nil {
		return err
	}
	if len(p.Name) > 64 {
		return apperr.New(apperr.ErrInvalidParam.Code, "name 长度不能超过 64")
	}
	if err := validator.RequireNonBlank(p.Module, "module"); err != nil {
		return err
	}
	if p.Type != "api" && p.Type != "menu" && p.Type != "button" {
		return apperr.New(apperr.ErrInvalidParam.Code, "type 必须是 api、menu 或 button")
	}
	return nil
}

type PermissionTree struct {
	Module     string         `json:"module"`
	ModuleName string         `json:"moduleName"`
	Children   []PermissionTreeItem `json:"children"`
}

type PermissionTreeItem struct {
	ID          uint64 `json:"id"`
	PermKey     string `json:"permKey"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
}

// UserMenuItem 用户菜单项（前端动态菜单渲染用）
type UserMenuItem struct {
	ID        uint64 `json:"id"`
	PermKey   string `json:"permKey"`
	Name      string `json:"name"`
	Module    string `json:"module"`
	FrontPath string `json:"frontPath"`
	Icon      string `json:"icon"`
	SortOrder int    `json:"sortOrder"`
}
