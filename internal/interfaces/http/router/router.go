package router

import (
	rolesvc "sanmoo-server-go/internal/application/role"
	"sanmoo-server-go/internal/infrastructure/repository/mysqlrepo"
	"sanmoo-server-go/internal/infrastructure/security"
	"sanmoo-server-go/internal/interfaces/http/handler"
	"sanmoo-server-go/internal/interfaces/http/middleware"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func Register(e *gin.Engine, h *handler.Handler, jwt *security.JWTManager, repo *mysqlrepo.Repository, roleSvc *rolesvc.Service, redisClient *redis.Client) {
	// 添加全局日志中间件
	e.Use(middleware.Logger())

	// 健康检查接口，通常用于容器探活。
	e.GET("/health", handler.Health)
	e.GET("/sitemap.xml", h.Sitemap)
	e.GET("/rss.xml", h.RSS)
	e.GET("/swagger", handler.SwaggerPage)
	e.GET("/swagger/", func(c *gin.Context) { c.Redirect(302, "/swagger") })
	e.GET("/swagger/openapi.json", handler.SwaggerSpec)

	// 认证相关接口（无需登录）。添加严格限流（5次/分钟，防暴力破解）
	auth := e.Group("/auth")
	auth.Use(middleware.RateLimit(redisClient, middleware.AuthRateLimit))
	{
		auth.POST("/login", h.Login)
		auth.POST("/refresh", h.Refresh)
		auth.POST("/send-verification-code", h.SendLoginVerificationCode)
		auth.POST("/verify-verification-code", h.VerifyLoginVerificationCode)
	}

	// 用户相关接口（需要登录）。
	user := e.Group("/user")
	user.Use(middleware.JWTAuth(jwt, repo))
	{
		user.PUT("/password", h.ChangePassword)
	}

	// 管理端接口统一要求 JWT 鉴权。添加宽松限流（300次/分钟）
	admin := e.Group("/admin")
	admin.Use(middleware.JWTAuth(jwt, repo))
	admin.Use(middleware.RateLimit(redisClient, middleware.AdminRateLimit))
	{
		// 用户管理
		admin.GET("/users", middleware.RequirePerm(roleSvc, "user:list"), h.GetUsers)
		admin.GET("/users/export", middleware.RequirePerm(roleSvc, "user:export"), h.ExportUsers)
		admin.POST("/users", middleware.RequirePerm(roleSvc, "user:create"), h.CreateUser)
		admin.PUT("/users/:id", middleware.RequirePerm(roleSvc, "user:update"), h.UpdateUser)
		admin.DELETE("/users/:id", middleware.RequirePerm(roleSvc, "user:delete"), h.DeleteUser)
		admin.PUT("/users/:id/password", middleware.RequirePerm(roleSvc, "user:password:reset"), h.UpdateUserPassword)
		admin.PUT("/users/:id/status", middleware.RequirePerm(roleSvc, "user:status"), h.ToggleUserStatus)
		admin.DELETE("/users/batch-delete", middleware.RequirePerm(roleSvc, "user:delete"), h.BatchDeleteUsers)
		admin.PUT("/users/:id/roles", middleware.RequirePerm(roleSvc, "user:role"), h.AssignUserRoles)

		// 标签管理
		admin.GET("/tags", middleware.RequirePerm(roleSvc, "tag:list"), h.GetTags)
		admin.POST("/tags", middleware.RequirePerm(roleSvc, "tag:create"), h.CreateTag)
		admin.PUT("/tags/:id", middleware.RequirePerm(roleSvc, "tag:update"), h.UpdateTag)
		admin.DELETE("/tags/:id", middleware.RequirePerm(roleSvc, "tag:delete"), h.DeleteTag)
		admin.DELETE("/tags/batch-delete", middleware.RequirePerm(roleSvc, "tag:delete"), h.BatchDeleteTags)

		// 分类管理
		admin.GET("/categories", middleware.RequirePerm(roleSvc, "category:list"), h.GetCategories)
		admin.POST("/categories", middleware.RequirePerm(roleSvc, "category:create"), h.CreateCategory)
		admin.PUT("/categories/:id", middleware.RequirePerm(roleSvc, "category:update"), h.UpdateCategory)
		admin.DELETE("/categories/:id", middleware.RequirePerm(roleSvc, "category:delete"), h.DeleteCategory)
		admin.DELETE("/categories/batch-delete", middleware.RequirePerm(roleSvc, "category:delete"), h.BatchDeleteCategories)

		// 友情链接
		admin.GET("/links", middleware.RequirePerm(roleSvc, "link:list"), h.GetLinks)
		admin.POST("/links", middleware.RequirePerm(roleSvc, "link:create"), h.CreateLink)
		admin.PUT("/links/:id", middleware.RequirePerm(roleSvc, "link:update"), h.UpdateLink)
		admin.DELETE("/links/:id", middleware.RequirePerm(roleSvc, "link:delete"), h.DeleteLink)
		admin.DELETE("/links/batch-delete", middleware.RequirePerm(roleSvc, "link:delete"), h.BatchDeleteLinks)

		// 专题管理
		admin.GET("/topics", middleware.RequirePerm(roleSvc, "topic:list"), h.GetTopics)
		admin.POST("/topics", middleware.RequirePerm(roleSvc, "topic:create"), h.CreateTopic)
		admin.PUT("/topics/:id", middleware.RequirePerm(roleSvc, "topic:update"), h.UpdateTopic)
		admin.DELETE("/topics/:id", middleware.RequirePerm(roleSvc, "topic:delete"), h.DeleteTopic)
		admin.DELETE("/topics/batch-delete", middleware.RequirePerm(roleSvc, "topic:delete"), h.BatchDeleteTopics)
		admin.GET("/topics/:id/articles", middleware.RequirePerm(roleSvc, "topic:articles"), h.GetTopicArticles)
		admin.GET("/articles/published-options", middleware.RequirePerm(roleSvc, "article:list"), h.GetPublishedArticleOptions)

		// 文章管理
		admin.GET("/articles", middleware.RequirePerm(roleSvc, "article:list"), h.GetArticles)
		admin.GET("/articles/export", middleware.RequirePerm(roleSvc, "article:export"), h.ExportArticles)
		admin.GET("/articles/:id", middleware.RequirePerm(roleSvc, "article:detail"), h.AdminArticleDetail)
		admin.POST("/articles", middleware.RequirePerm(roleSvc, "article:create"), h.CreateArticle)
		admin.PUT("/articles/:id", middleware.RequirePerm(roleSvc, "article:update"), h.UpdateArticle)
		admin.PUT("/articles/:id/status", middleware.RequirePerm(roleSvc, "article:status"), h.UpdateArticleStatus)
		admin.PUT("/articles/batch-status", middleware.RequirePerm(roleSvc, "article:status"), h.BatchUpdateArticleStatus)
		admin.DELETE("/articles/:id", middleware.RequirePerm(roleSvc, "article:delete"), h.DeleteArticle)
		admin.DELETE("/articles/batch-delete", middleware.RequirePerm(roleSvc, "article:delete"), h.BatchDeleteArticles)

		// 设置
		admin.GET("/settings", middleware.RequirePerm(roleSvc, "setting:read"), h.AdminGetSettings)
		admin.PUT("/settings", middleware.RequirePerm(roleSvc, "setting:update"), h.AdminUpdateSettings)
		admin.POST("/settings/email/send-code", middleware.RequirePerm(roleSvc, "setting:email"), h.AdminSendEmailVerificationCode)
		admin.POST("/settings/email/verify-code", middleware.RequirePerm(roleSvc, "setting:email"), h.AdminVerifyEmailVerificationCode)
		admin.GET("/settings/export", middleware.RequirePerm(roleSvc, "setting:export"), h.AdminExportSettings)
		admin.POST("/settings/import", middleware.RequirePerm(roleSvc, "setting:import"), h.AdminImportSettings)
		admin.POST("/search/sync", middleware.RequirePerm(roleSvc, "setting:search"), h.AdminSyncMeiliSearch)

		// 独立配置接口
		admin.GET("/settings/core", middleware.RequirePerm(roleSvc, "setting:core:read"), h.AdminGetCoreConfig)
		admin.PUT("/settings/core", middleware.RequirePerm(roleSvc, "setting:core:update"), h.AdminUpdateCoreConfig)
		admin.GET("/settings/privacy", middleware.RequirePerm(roleSvc, "setting:privacy:read"), h.AdminGetPrivacyConfig)
		admin.PUT("/settings/privacy", middleware.RequirePerm(roleSvc, "setting:privacy:update"), h.AdminUpdatePrivacyConfig)
		admin.GET("/settings/social", middleware.RequirePerm(roleSvc, "setting:social:read"), h.AdminGetSocialConfig)
		admin.PUT("/settings/social", middleware.RequirePerm(roleSvc, "setting:social:update"), h.AdminUpdateSocialConfig)
		admin.GET("/settings/search", middleware.RequirePerm(roleSvc, "setting:search:read"), h.AdminGetSearchConfig)
		admin.PUT("/settings/search", middleware.RequirePerm(roleSvc, "setting:search:update"), h.AdminUpdateSearchConfig)
		admin.GET("/settings/storage", middleware.RequirePerm(roleSvc, "setting:storage:read"), h.AdminGetStorageConfig)
		admin.PUT("/settings/storage", middleware.RequirePerm(roleSvc, "setting:storage:update"), h.AdminUpdateStorageConfig)
		admin.GET("/settings/email", middleware.RequirePerm(roleSvc, "setting:email:read"), h.AdminGetEmailConfig)
		admin.PUT("/settings/email", middleware.RequirePerm(roleSvc, "setting:email:update"), h.AdminUpdateEmailConfig)

		// 文件管理
		admin.GET("/files", middleware.RequirePerm(roleSvc, "file:list"), h.GetFiles)
		admin.POST("/files/upload", middleware.RequirePerm(roleSvc, "file:upload"), h.UploadFile)
		admin.DELETE("/files/:id", middleware.RequirePerm(roleSvc, "file:delete"), h.DeleteFile)

		// 仪表盘
		admin.GET("/dashboard", middleware.RequirePerm(roleSvc, "dashboard:read"), h.Dashboard)
		admin.GET("/dashboard/visitors", middleware.RequirePerm(roleSvc, "dashboard:visitors"), h.VisitorRecords)
		admin.DELETE("/dashboard/visitors/:id", middleware.RequirePerm(roleSvc, "dashboard:visitors:delete"), h.DeleteVisitorRecord)
		admin.DELETE("/dashboard/visitors/batch-delete", middleware.RequirePerm(roleSvc, "dashboard:visitors:delete"), h.BatchDeleteVisitorRecords)
		admin.DELETE("/dashboard/visitors/clear", middleware.RequirePerm(roleSvc, "dashboard:visitors:delete"), h.ClearAllVisitorRecords)
		admin.GET("/dashboard/errors", middleware.RequirePerm(roleSvc, "dashboard:errors"), h.ErrorLogRecords)
		admin.DELETE("/dashboard/errors/:id", middleware.RequirePerm(roleSvc, "dashboard:errors:delete"), h.DeleteErrorLog)
		admin.DELETE("/dashboard/errors/batch-delete", middleware.RequirePerm(roleSvc, "dashboard:errors:delete"), h.BatchDeleteErrorLogs)
		admin.DELETE("/dashboard/errors/clear", middleware.RequirePerm(roleSvc, "dashboard:errors:delete"), h.ClearAllErrorLogs)
		admin.POST("/dashboard/errors/import", middleware.RequirePerm(roleSvc, "dashboard:errors"), h.ImportErrorLogs)
		admin.GET("/dashboard/errors/export", middleware.RequirePerm(roleSvc, "dashboard:errors"), h.ExportErrorLogs)
		admin.GET("/dashboard/pv", middleware.RequirePerm(roleSvc, "dashboard:pv"), h.PV)
		admin.GET("/dashboard/tag-statistics", middleware.RequirePerm(roleSvc, "dashboard:statistics"), h.TagStatistics)
		admin.GET("/dashboard/category-statistics", middleware.RequirePerm(roleSvc, "dashboard:statistics"), h.CategoryStatistics)
		admin.GET("/dashboard/heatmap", middleware.RequirePerm(roleSvc, "dashboard:statistics"), h.ArticlePublishHeatmap)
		admin.GET("/dashboard/topic-statistics", middleware.RequirePerm(roleSvc, "dashboard:statistics"), h.TopicStatistics)
		admin.GET("/dashboard/mp-user-growth", middleware.RequirePerm(roleSvc, "dashboard:statistics"), h.MpUserGrowth)
		admin.GET("/dashboard/article-read-statistics", middleware.RequirePerm(roleSvc, "dashboard:statistics"), h.ArticleReadStatistics)
		admin.GET("/dashboard/category-read-statistics", middleware.RequirePerm(roleSvc, "dashboard:statistics"), h.CategoryReadStatistics)
		admin.GET("/dashboard/tag-read-statistics", middleware.RequirePerm(roleSvc, "dashboard:statistics"), h.TagReadStatistics)
		admin.GET("/dashboard/content-trend", middleware.RequirePerm(roleSvc, "dashboard:statistics"), h.ContentTrend)

		// 缓存管理
		admin.POST("/cache/clear", middleware.RequirePerm(roleSvc, "cache:clear"), h.ClearCache)
		admin.POST("/cache/warmup", middleware.RequirePerm(roleSvc, "cache:warmup"), h.WarmupCache)
		admin.GET("/cache/stats", middleware.RequirePerm(roleSvc, "cache:stats"), h.CacheStats)

		// 备份管理
		admin.POST("/backup/export", middleware.RequirePerm(roleSvc, "backup:export"), h.ExportData)
		admin.GET("/backup/list", middleware.RequirePerm(roleSvc, "backup:list"), h.ListBackups)
		admin.GET("/backup/download/:fileName", middleware.RequirePerm(roleSvc, "backup:download"), h.DownloadBackup)
		admin.DELETE("/backup/:fileName", middleware.RequirePerm(roleSvc, "backup:delete"), h.DeleteBackup)
		admin.GET("/backup/stats", middleware.RequirePerm(roleSvc, "backup:stats"), h.GetBackupStats)

		// 微信用户管理
		admin.GET("/mp-users", middleware.RequirePerm(roleSvc, "mpuser:list"), h.AdminMPUsers)
		admin.GET("/mp-users/:openid", middleware.RequirePerm(roleSvc, "mpuser:detail"), h.AdminMPUserDetail)
		admin.GET("/mp-users/:openid/profile", middleware.RequirePerm(roleSvc, "mpuser:profile"), h.AdminMPUserProfile)
		admin.POST("/mp-users/:openid/profile", middleware.RequirePerm(roleSvc, "mpuser:profile"), h.AdminMPUserGenerateProfile)
		admin.POST("/mp-users/:openid/tags/generate", middleware.RequirePerm(roleSvc, "mpuser:tags"), h.AdminMPUserGenerateTags)
		admin.GET("/mp-users/:openid/tags", middleware.RequirePerm(roleSvc, "mpuser:tags"), h.AdminMPUserTags)
		admin.DELETE("/mp-users/:openid/tags/:tagId", middleware.RequirePerm(roleSvc, "mpuser:tags"), h.AdminMPUserDeleteTag)
		admin.POST("/mp-users/:openid/radar/refresh", middleware.RequirePerm(roleSvc, "mpuser:profile"), h.AdminMPUserRefreshRadar)

		// 权限管理
		admin.GET("/permissions", middleware.RequirePerm(roleSvc, "permission:list"), h.GetPermissions)
		admin.GET("/permissions/tree", middleware.RequirePerm(roleSvc, "permission:list"), h.GetPermissionTree)
		admin.GET("/permissions/:id", middleware.RequirePerm(roleSvc, "permission:list"), h.GetPermission)
		admin.POST("/permissions", middleware.RequirePerm(roleSvc, "permission:list"), h.CreatePermission)
		admin.PUT("/permissions/:id", middleware.RequirePerm(roleSvc, "permission:list"), h.UpdatePermission)
		admin.DELETE("/permissions/:id", middleware.RequirePerm(roleSvc, "permission:list"), h.DeletePermission)

		// 角色管理
		admin.GET("/roles", middleware.RequirePerm(roleSvc, "role:list"), h.GetRoles)
		admin.GET("/roles/all", middleware.RequirePerm(roleSvc, "role:list"), h.GetAllRoles)
		admin.GET("/roles/:id", middleware.RequirePerm(roleSvc, "role:list"), h.GetRole)
		admin.POST("/roles", middleware.RequirePerm(roleSvc, "role:create"), h.CreateRole)
		admin.PUT("/roles/:id", middleware.RequirePerm(roleSvc, "role:update"), h.UpdateRole)
		admin.DELETE("/roles/:id", middleware.RequirePerm(roleSvc, "role:delete"), h.DeleteRole)
		admin.PUT("/roles/:id/permissions", middleware.RequirePerm(roleSvc, "role:permission"), h.AssignRolePermissions)
		admin.GET("/roles/:id/permissions", middleware.RequirePerm(roleSvc, "role:list"), h.GetRolePermissions)

		// 当前用户权限
		admin.GET("/user/permissions", h.GetUserPermissions)
		// 当前用户菜单（前端动态菜单渲染用）
		admin.GET("/user/menus", h.GetUserMenus)

		// 数据维护
		admin.GET("/maintenance/stats", middleware.RequirePerm(roleSvc, "maintenance:stats"), h.AdminMaintenanceStats)
		admin.POST("/maintenance/cleanup-logs", middleware.RequirePerm(roleSvc, "maintenance:cleanup"), h.AdminMaintenanceCleanupLogs)
	}

	// 门户端接口（公开访问）。添加公开读接口限流（60次/分钟）
	web := e.Group("/web")
	web.Use(middleware.RateLimit(redisClient, middleware.PublicReadRateLimit))
	{
		web.GET("/settings", h.WebSettings)
		web.GET("/categories", h.WebCategories)
		web.GET("/tags", h.WebTags)
		web.GET("/articles", h.WebArticles)
		web.GET("/articles/:id", h.WebArticleDetail)
		web.GET("/articles/slug/:slug", h.WebArticleBySlug) // 新增：slug 查询路由（SEO 友好）
		web.GET("/categories/:id/articles", h.WebArticlesByCategory)
		web.GET("/tags/:id/articles", h.WebArticlesByTag)
		web.GET("/archives", h.WebArchives)
		web.GET("/topics", h.WebTopics)
		web.GET("/topics/:id", h.WebTopicDetail)
		web.GET("/topics/:id/articles", h.WebTopicArticles)
		web.GET("/search/hot", h.WebHotSearches)
		web.GET("/search", h.WebSearch) // 搜索接口单独限流
		web.GET("/articles/:id/related", h.WebRelatedArticles)
		web.GET("/articles/hot", h.WebHotArticles)
		// 点赞接口：先过写接口限流（10次/分钟），再过点赞防刷（24小时内只能点1次）
		web.POST("/articles/:id/like", middleware.LikeRateLimit(redisClient), h.WebLikeArticle)
		web.GET("/articles/random", h.WebRandomArticle)
		web.GET("/links", h.GetActiveLinks)
	}

	// 小程序端接口：按约定与 /web 保持 1:1 对应。添加公开读接口限流（60次/分钟）
	mp := e.Group("/mp")
	mp.Use(middleware.RateLimit(redisClient, middleware.PublicReadRateLimit))
	{
		mp.POST("/auth/session", h.MpAuthSession)
		mp.GET("/user/profile", h.MpUserProfile)
		mp.POST("/user/profile", h.MpUpdateUserProfile)
		mp.POST("/behavior", h.MpReportBehavior)
		mp.GET("/settings", h.MpSettings)
		mp.GET("/categories", h.MpCategories)
		mp.GET("/tags", h.MpTags)
		mp.GET("/topics", h.MpTopics)
		mp.GET("/topics/:id", h.MpTopicDetail)
		mp.GET("/topics/:id/articles", h.MpTopicArticles)
		mp.GET("/articles", h.MpArticles)
		mp.GET("/recommendations", h.MpRecommendations)
		mp.GET("/articles/:id", h.MpArticleDetail)
		mp.POST("/favorites/:articleId", h.MpAddFavorite)
		mp.DELETE("/favorites/:articleId", h.MpRemoveFavorite)
		mp.GET("/favorites/:articleId/status", h.MpFavoriteStatus)
		mp.GET("/favorites", h.MpFavorites)
		mp.POST("/history/:articleId", h.MpAddBrowseHistory)
		mp.DELETE("/history", h.MpClearBrowseHistory)
		mp.GET("/history", h.MpBrowseHistory)
		mp.GET("/categories/:id/articles", h.MpArticlesByCategory)
		mp.GET("/tags/:id/articles", h.MpArticlesByTag)
		mp.GET("/archives", h.MpArchives)
		mp.GET("/search/hot", h.WebHotSearches)
		mp.GET("/privacy-policy", h.MpPrivacyPolicy)
		mp.DELETE("/user", h.MpDeleteUser)
		mp.POST("/subscribe", h.MpSubscribe)
		mp.GET("/subscribe/status", h.MpSubscribeStatus)
	}
}
