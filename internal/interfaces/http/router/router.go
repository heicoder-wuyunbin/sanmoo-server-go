package router

import (
	"sanmoo-server-go/internal/infrastructure/repository/mysqlrepo"
	"sanmoo-server-go/internal/infrastructure/security"
	"sanmoo-server-go/internal/interfaces/http/handler"
	"sanmoo-server-go/internal/interfaces/http/middleware"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func Register(e *gin.Engine, h *handler.Handler, jwt *security.JWTManager, repo *mysqlrepo.Repository, redisClient *redis.Client) {
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
		auth.POST("/check-mfa", h.CheckMFA)
		auth.POST("/send-verification-code", h.SendLoginVerificationCode)
		auth.POST("/verify-verification-code", h.VerifyLoginVerificationCode)
	}

	// 用户相关接口（需要登录）。
	user := e.Group("/user")
	user.Use(middleware.JWTAuth(jwt, repo))
	{
		user.PUT("/password", h.ChangePassword)
	}

	// 管理端接口统一要求 JWT 鉴权 + Admin 权限。添加宽松限流（300次/分钟）
	admin := e.Group("/admin")
	admin.Use(middleware.JWTAuth(jwt, repo))
	admin.Use(middleware.RequireAdmin(repo))
	admin.Use(middleware.RateLimit(redisClient, middleware.AdminRateLimit))
	{
		// 标签管理
		admin.GET("/tags", h.GetTags)
		admin.POST("/tags", h.CreateTag)
		admin.PUT("/tags/:id", h.UpdateTag)
		admin.DELETE("/tags/:id", h.DeleteTag)
		admin.DELETE("/tags/batch-delete", h.BatchDeleteTags)

		// 分类管理
		admin.GET("/categories", h.GetCategories)
		admin.POST("/categories", h.CreateCategory)
		admin.PUT("/categories/:id", h.UpdateCategory)
		admin.DELETE("/categories/:id", h.DeleteCategory)
		admin.DELETE("/categories/batch-delete", h.BatchDeleteCategories)

		// 友情链接
		admin.GET("/links", h.GetLinks)
		admin.POST("/links", h.CreateLink)
		admin.PUT("/links/:id", h.UpdateLink)
		admin.DELETE("/links/:id", h.DeleteLink)
		admin.DELETE("/links/batch-delete", h.BatchDeleteLinks)

		// 专题管理
		admin.GET("/topics", h.GetTopics)
		admin.POST("/topics", h.CreateTopic)
		admin.PUT("/topics/:id", h.UpdateTopic)
		admin.DELETE("/topics/:id", h.DeleteTopic)
		admin.DELETE("/topics/batch-delete", h.BatchDeleteTopics)
		admin.GET("/topics/:id/articles", h.GetTopicArticles)
		admin.GET("/articles/published-options", h.GetPublishedArticleOptions)

		// 文章管理
		admin.GET("/articles", h.GetArticles)
		admin.GET("/articles/export", h.ExportArticles)
		admin.GET("/articles/:id", h.AdminArticleDetail)
		admin.POST("/articles", h.CreateArticle)
		admin.PUT("/articles/:id", h.UpdateArticle)
		admin.PUT("/articles/:id/status", h.UpdateArticleStatus)
		admin.PUT("/articles/batch-status", h.BatchUpdateArticleStatus)
		admin.DELETE("/articles/:id", h.DeleteArticle)
		admin.DELETE("/articles/batch-delete", h.BatchDeleteArticles)

		// 设置
		admin.GET("/settings", h.AdminGetSettings)
		admin.PUT("/settings", h.AdminUpdateSettings)
		admin.POST("/settings/email/send-code", h.AdminSendEmailVerificationCode)
		admin.POST("/settings/email/verify-code", h.AdminVerifyEmailVerificationCode)
		admin.GET("/settings/export", h.AdminExportSettings)
		admin.POST("/settings/import", h.AdminImportSettings)
		admin.POST("/search/sync", h.AdminSyncMeiliSearch)
		admin.GET("/search/stats", h.AdminGetMeiliSearchStats)

		// 独立配置接口
		admin.GET("/settings/core", h.AdminGetCoreConfig)
		admin.PUT("/settings/core", h.AdminUpdateCoreConfig)
		admin.GET("/settings/privacy", h.AdminGetPrivacyConfig)
		admin.PUT("/settings/privacy", h.AdminUpdatePrivacyConfig)
		admin.GET("/settings/social", h.AdminGetSocialConfig)
		admin.PUT("/settings/social", h.AdminUpdateSocialConfig)
		admin.GET("/settings/search", h.AdminGetSearchConfig)
		admin.PUT("/settings/search", h.AdminUpdateSearchConfig)
		admin.GET("/settings/storage", h.AdminGetStorageConfig)
		admin.PUT("/settings/storage", h.AdminUpdateStorageConfig)
		admin.GET("/settings/email", h.AdminGetEmailConfig)
		admin.PUT("/settings/email", h.AdminUpdateEmailConfig)
		admin.GET("/settings/wechat", h.AdminGetWechatConfig)
		admin.PUT("/settings/wechat", h.AdminUpdateWechatConfig)

		// 文件管理
		admin.GET("/files", h.GetFiles)
		admin.POST("/files/upload", h.UploadFile)
		admin.DELETE("/files/:id", h.DeleteFile)
		// 图片代理：302 重定向到临时 URL，避免过期（无需认证，因为文件路径本身就是唯一标识）
		e.GET("/admin/files/image/*path", h.ProxyImage)

		// 仪表盘
		admin.GET("/dashboard", h.Dashboard)
		admin.GET("/dashboard/visitors", h.VisitorRecords)
		admin.DELETE("/dashboard/visitors/:id", h.DeleteVisitorRecord)
		admin.DELETE("/dashboard/visitors/batch-delete", h.BatchDeleteVisitorRecords)
		admin.DELETE("/dashboard/visitors/clear", h.ClearAllVisitorRecords)
		admin.GET("/dashboard/errors", h.ErrorLogRecords)
		admin.DELETE("/dashboard/errors/:id", h.DeleteErrorLog)
		admin.DELETE("/dashboard/errors/batch-delete", h.BatchDeleteErrorLogs)
		admin.DELETE("/dashboard/errors/clear", h.ClearAllErrorLogs)
		admin.POST("/dashboard/errors/import", h.ImportErrorLogs)
		admin.GET("/dashboard/errors/export", h.ExportErrorLogs)
		admin.GET("/dashboard/pv", h.PV)
		admin.GET("/dashboard/tag-statistics", h.TagStatistics)
		admin.GET("/dashboard/category-statistics", h.CategoryStatistics)
		admin.GET("/dashboard/heatmap", h.ArticlePublishHeatmap)
		admin.GET("/dashboard/topic-statistics", h.TopicStatistics)
		admin.GET("/dashboard/mp-user-growth", h.MpUserGrowth)
		admin.GET("/dashboard/article-read-statistics", h.ArticleReadStatistics)
		admin.GET("/dashboard/category-read-statistics", h.CategoryReadStatistics)
		admin.GET("/dashboard/tag-read-statistics", h.TagReadStatistics)
		admin.GET("/dashboard/content-trend", h.ContentTrend)
	}

	// 门户端接口（公开访问）。添加公开读接口限流（60次/分钟）
	web := e.Group("/web")
	web.Use(middleware.RateLimit(redisClient, middleware.PublicReadRateLimit))
	{
		web.GET("/settings", h.WebSettings)
		web.GET("/compliance", h.WebCompliance)
		web.GET("/categories", h.WebCategories)
		web.GET("/tags", h.WebTags)
		web.GET("/articles", h.WebArticles)
		web.GET("/articles/:id", h.WebArticleDetail)
		web.GET("/articles/slug/:slug", h.WebArticleBySlug)
		web.GET("/categories/:id/articles", h.WebArticlesByCategory)
		web.GET("/tags/:id/articles", h.WebArticlesByTag)
		web.GET("/archives", h.WebArchives)
		web.GET("/topics", h.WebTopics)
		web.GET("/topics/:id", h.WebTopicDetail)
		web.GET("/topics/:id/articles", h.WebTopicArticles)
		web.GET("/search/hot", h.WebHotSearches)
		web.GET("/search", h.WebSearch)
		web.GET("/articles/:id/related", h.WebRelatedArticles)
		web.GET("/articles/hot", h.WebHotArticles)
		web.POST("/articles/:id/like", middleware.LikeRateLimit(redisClient), h.WebLikeArticle)
		web.GET("/articles/random", h.WebRandomArticle)
		web.GET("/links", h.GetActiveLinks)
	}

	// 小程序端接口：按约定与 /web 保持 1:1 对应。添加公开读接口限流（60次/分钟）
	mp := e.Group("/mp")
	mp.Use(middleware.RateLimit(redisClient, middleware.PublicReadRateLimit))
	{
		mp.POST("/auth/session", h.MpAuthSession)
		mp.GET("/settings", h.MpSettings)
		mp.GET("/compliance", h.MpCompliance)
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
	}
}