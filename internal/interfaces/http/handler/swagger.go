package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const swaggerHTML = `<!doctype html>
<html lang="en">
<head>
  <meta charset="UTF-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1.0" />
  <title>Sanmoo API Docs</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css" />
  <style>
    html, body, #swagger-ui { height: 100%; margin: 0; }
    body { background: #fafafa; }
  </style>
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    window.ui = SwaggerUIBundle({
      url: '/swagger/openapi.json',
      dom_id: '#swagger-ui',
      deepLinking: true,
      presets: [SwaggerUIBundle.presets.apis],
      layout: 'BaseLayout',
    });
  </script>
</body>
</html>`

// SwaggerPage 返回 Swagger UI 页面。
func SwaggerPage(c *gin.Context) {
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, swaggerHTML)
}

// 路由信息结构
type RouteInfo struct {
	Method      string
	Path        string
	Summary     string
	Description string
}

// 路由信息列表
var routes = []RouteInfo{
	{"GET", "/health", "健康检查", "用于容器探活的健康检查接口"},
	{"POST", "/auth/login", "用户登录", "用户登录接口"},
	{"POST", "/auth/refresh", "刷新令牌", "刷新访问令牌"},
	{"POST", "/auth/send-verification-code", "发送验证码", "发送登录验证码"},
	{"POST", "/auth/verify-verification-code", "验证验证码", "验证登录验证码"},
	{"PUT", "/user/password", "修改密码", "修改用户密码"},
	{"GET", "/admin/users", "获取用户列表", "获取所有用户列表"},
	{"POST", "/admin/users", "创建用户", "创建新用户"},
	{"PUT", "/admin/users/:id", "更新用户", "更新用户信息"},
	{"DELETE", "/admin/users/:id", "删除用户", "删除用户"},
	{"PUT", "/admin/users/:id/password", "更新用户密码", "更新用户密码"},
	{"PUT", "/admin/users/:id/status", "切换用户状态", "切换用户启用/禁用状态"},
	{"DELETE", "/admin/users/batch-delete", "批量删除用户", "批量删除用户"},
	{"GET", "/admin/tags", "获取标签列表", "获取所有标签列表"},
	{"POST", "/admin/tags", "创建标签", "创建新标签"},
	{"PUT", "/admin/tags/:id", "更新标签", "更新标签信息"},
	{"DELETE", "/admin/tags/:id", "删除标签", "删除标签"},
	{"DELETE", "/admin/tags/batch-delete", "批量删除标签", "批量删除标签"},
	{"GET", "/admin/categories", "获取分类列表", "获取所有分类列表"},
	{"POST", "/admin/categories", "创建分类", "创建新分类"},
	{"PUT", "/admin/categories/:id", "更新分类", "更新分类信息"},
	{"DELETE", "/admin/categories/:id", "删除分类", "删除分类"},
	{"DELETE", "/admin/categories/batch-delete", "批量删除分类", "批量删除分类"},
	{"GET", "/admin/articles", "获取文章列表", "获取所有文章列表"},
	{"POST", "/admin/articles", "创建文章", "创建新文章"},
	{"PUT", "/admin/articles/:id", "更新文章", "更新文章信息"},
	{"DELETE", "/admin/articles/:id", "删除文章", "删除文章"},
	{"DELETE", "/admin/articles/batch-delete", "批量删除文章", "批量删除文章"},
	{"GET", "/admin/settings", "获取设置", "获取系统设置"},
	{"PUT", "/admin/settings", "更新设置", "更新系统设置"},
	{"POST", "/admin/staticize/incremental", "增量静态化", "执行增量静态化"},
	{"POST", "/admin/staticize/full", "全量静态化", "执行全量静态化"},
	{"DELETE", "/admin/staticize", "清理静态文件", "清理静态化文件"},
	{"GET", "/admin/files", "获取文件列表", "获取所有文件列表"},
	{"POST", "/admin/files/upload", "上传文件", "上传文件"},
	{"DELETE", "/admin/files/:id", "删除文件", "删除文件"},
	{"GET", "/admin/dashboard", "获取仪表盘", "获取仪表盘数据"},
	{"GET", "/admin/dashboard/visitors", "获取访问记录", "获取访问记录"},
	{"GET", "/admin/dashboard/pv", "获取访问量", "获取页面访问量"},
	{"GET", "/admin/dashboard/tag-statistics", "获取标签统计", "获取标签使用统计"},
	{"GET", "/admin/dashboard/category-statistics", "获取分类统计", "获取分类使用统计"},
	{"GET", "/admin/dashboard/heatmap", "获取发布热图", "获取文章发布热图"},
	{"GET", "/web/settings", "获取前端设置", "获取前端系统设置"},
	{"GET", "/web/categories", "获取前端分类", "获取前端分类列表"},
	{"GET", "/web/tags", "获取前端标签", "获取前端标签列表"},
	{"GET", "/web/articles", "获取前端文章", "获取前端文章列表"},
	{"GET", "/web/articles/:id", "获取前端文章详情", "获取前端文章详细信息"},
	{"GET", "/web/categories/:id/articles", "获取分类文章", "获取指定分类的文章列表"},
	{"GET", "/web/tags/:id/articles", "获取标签文章", "获取指定标签的文章列表"},
	{"GET", "/web/archives", "获取归档", "获取文章归档"},
	{"POST", "/mp/auth/session", "小程序登录", "小程序登录接口"},
	{"GET", "/mp/user/profile", "获取小程序用户资料", "根据 openid 获取小程序用户头像昵称（首次需客户端授权后写入）"},
	{"POST", "/mp/user/profile", "更新小程序用户资料", "保存小程序用户头像昵称（按 openid 覆盖）"},
	{"GET", "/mp/settings", "获取小程序设置", "获取小程序系统设置"},
	{"GET", "/mp/categories", "获取小程序分类", "获取小程序分类列表"},
	{"GET", "/mp/tags", "获取小程序标签", "获取小程序标签列表"},
	{"GET", "/mp/articles", "获取小程序文章", "获取小程序文章列表"},
	{"GET", "/mp/recommendations", "获取推荐文章", "获取推荐文章列表"},
	{"GET", "/mp/articles/:id", "获取小程序文章详情", "获取小程序文章详细信息"},
	{"POST", "/mp/favorites/:articleId", "添加收藏", "添加文章收藏"},
	{"DELETE", "/mp/favorites/:articleId", "取消收藏", "取消文章收藏"},
	{"GET", "/mp/favorites/:articleId/status", "获取收藏状态", "获取文章收藏状态"},
	{"GET", "/mp/favorites", "获取收藏列表", "获取用户收藏列表"},
	{"GET", "/mp/categories/:id/articles", "获取小程序分类文章", "获取小程序指定分类的文章列表"},
	{"GET", "/mp/tags/:id/articles", "获取小程序标签文章", "获取小程序指定标签的文章列表"},
	{"GET", "/mp/archives", "获取小程序归档", "获取小程序文章归档"},
}

// SwaggerSpec 返回 OpenAPI JSON 文件。
func SwaggerSpec(c *gin.Context) {
	c.Header("Content-Type", "application/json; charset=utf-8")

	// 按路径分组路由
	pathGroups := make(map[string][]RouteInfo)
	for _, route := range routes {
		pathGroups[route.Path] = append(pathGroups[route.Path], route)
	}

	// 构建 OpenAPI 规范
	openapi := `{
  "openapi": "3.0.0",
  "info": {
    "title": "Sanmoo API",
    "description": "Sanmoo Blog API Documentation",
    "version": "1.0.0"
  },
  "paths": {
`

	// 遍历路径组，生成路径定义
	i := 0
	for path, routeGroup := range pathGroups {
		// 生成路径定义
		openapi += `    "` + path + `": {
`

		// 遍历路径下的所有方法
		for j, route := range routeGroup {
			// 将 HTTP 方法转换为小写
			method := route.Method
			if method == "GET" {
				method = "get"
			} else if method == "POST" {
				method = "post"
			} else if method == "PUT" {
				method = "put"
			} else if method == "DELETE" {
				method = "delete"
			}

			// 生成方法定义
			openapi += `      "` + method + `": {
        "summary": "` + route.Summary + `",
        "description": "` + route.Description + `",
        "responses": {
          "200": {
            "description": "成功"
          }
        }
      }`

			// 添加逗号（除了最后一个方法）
			if j < len(routeGroup)-1 {
				openapi += `,`
			}

			openapi += `
`
		}

		openapi += `    }`

		// 添加逗号（除了最后一个路径）
		if i < len(pathGroups)-1 {
			openapi += `,`
		}

		openapi += `
`
		i++
	}

	// 结束 OpenAPI 规范
	openapi += `  }
}`

	c.String(http.StatusOK, openapi)
}
