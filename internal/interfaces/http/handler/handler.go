package handler

import (
	"net/http"

	"sanmoo-server-go/internal/application/article"
	"sanmoo-server-go/internal/application/auth"
	cacheapp "sanmoo-server-go/internal/application/cache"
	"sanmoo-server-go/internal/application/category"
	"sanmoo-server-go/internal/application/dashboard"
	"sanmoo-server-go/internal/application/file"
	linkapp "sanmoo-server-go/internal/application/link"
	mpuserapp "sanmoo-server-go/internal/application/mpuser"
	"sanmoo-server-go/internal/application/setting"
	"sanmoo-server-go/internal/application/tag"
	"sanmoo-server-go/internal/application/topic"

	"github.com/gin-gonic/gin"
)

// Handler 聚合所有 HTTP 请求处理函数。
// 各领域 handler 已拆分至独立文件：
//   - auth_handler.go          认证（登录/验证码/刷新令牌）
//   - admin_user_handler.go    后台用户管理
//   - admin_tag_handler.go     后台标签管理
//   - admin_category_handler.go 后台分类管理
//   - admin_topic_handler.go   后台专题管理
//   - admin_article_handler.go 后台文章管理
//   - admin_dashboard_handler.go 后台仪表盘/统计
//   - admin_setting_handler.go 后台设置/邮件验证/搜索同步
//   - admin_file_handler.go    后台文件管理
//   - web_handler.go           门户端公开接口
//   - mp_handler.go            小程序接口
//   - mp_user_admin_handler.go 后台微信用户管理
//   - cache_handler.go         缓存管理
//   - helpers.go               公共辅助函数

// Services 聚合所有应用层服务，作为 Handler 的依赖。
type Services struct {
	Auth      *auth.Service
	Tag       *tag.Service
	Category  *category.Service
	Article   *article.Service
	Topic     *topic.Service
	Setting   *setting.Service
	File      *file.Service
	Dashboard *dashboard.Service
	MPUser    *mpuserapp.Service
	Cache     *cacheapp.Service
	Link      *linkapp.LinkService
}

type Handler struct {
	svc Services
}

func New(svc Services) *Handler {
	return &Handler{svc: svc}
}

func Health(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) }
