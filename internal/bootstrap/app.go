package bootstrap

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"sanmoo-server-go/internal/application/article"
	"sanmoo-server-go/internal/application/auth"
	cacheapp "sanmoo-server-go/internal/application/cache"
	"sanmoo-server-go/internal/application/category"
	"sanmoo-server-go/internal/application/dashboard"
	"sanmoo-server-go/internal/application/file"
	linkapp "sanmoo-server-go/internal/application/link"
	mpuserapp "sanmoo-server-go/internal/application/mpuser"
	"sanmoo-server-go/internal/application/scheduler"
	"sanmoo-server-go/internal/application/setting"
	"sanmoo-server-go/internal/application/tag"
	"sanmoo-server-go/internal/application/topic"
	"sanmoo-server-go/internal/infrastructure/cache"
	"sanmoo-server-go/internal/infrastructure/config"
	"sanmoo-server-go/internal/infrastructure/db"
	"sanmoo-server-go/internal/infrastructure/email"
	"sanmoo-server-go/internal/infrastructure/logger"
	"sanmoo-server-go/internal/infrastructure/repository/mysqlrepo"
	"sanmoo-server-go/internal/infrastructure/security"
	"sanmoo-server-go/internal/interfaces/http/handler"
	"sanmoo-server-go/internal/interfaces/http/middleware"
	"sanmoo-server-go/internal/interfaces/http/router"

	"github.com/gin-gonic/gin"
)

type App struct {
	cfg    *config.Config
	server *http.Server
}

func New(cfg *config.Config) (*App, error) {
	logger.Infof("开始初始化数据库连接...")
	logger.Infof("数据库DSN: %s", cfg.MySQLDSN)
	database, err := db.NewMySQL(cfg.MySQLDSN)
	if err != nil {
		logger.Errorf("初始化数据库失败: %v", err)
		return nil, fmt.Errorf("init mysql failed: %w", err)
	}
	if cfg.ServerMode == "debug" {
		db.SetGormLogMode(database, "debug")
	}
	logger.Infof("数据库连接初始化成功")

	logger.Infof("初始化仓库...")
	repo := mysqlrepo.New(database, "uploads", "/uploads/", "", "", "", "")
	// 启动时从数据库加载存储配置，确保重启后七牛云等配置不丢失
	if err := repo.LoadStorageConfig(context.Background()); err != nil {
		logger.Warnf("加载存储配置失败（使用默认配置）: %v", err)
	} else {
		logger.Infof("存储配置加载成功")
	}
	logger.Infof("仓库初始化成功")

	logger.Infof("初始化JWT管理器...")
	jwtMgr := security.NewJWTManager(cfg.JWT.AccessSecret, cfg.JWT.RefreshSecret, cfg.JWT.AccessTTLSeconds, cfg.JWT.RefreshTTLSeconds)
	logger.Infof("JWT管理器初始化成功")

	logger.Infof("初始化Redis客户端...")
	redisClient := cache.NewRedis(cfg.RedisAddr)
	logger.Infof("Redis客户端初始化成功")

	logger.Infof("初始化业务缓存服务...")
	bizCache := cache.NewBusinessCache(redisClient)
	logger.Infof("业务缓存服务初始化成功")

	logger.Infof("初始化验证码服务与邮件服务...")
	verificationService := cache.NewVerificationService(redisClient)
	emailService := email.NewEmailService()
	// 从数据库加载邮件配置（用于验证码发送/后台登录 MFA）
	if st, err := repo.Get(context.Background()); err == nil && st != nil && st.EmailConfig != nil {
		emailService.UpdateConfig(st.EmailConfig)
	}

	logger.Infof("初始化微信配置动态提供者...")
	wechatProvider := config.NewWechatConfigProvider(redisClient, func(ctx context.Context) (*config.WechatMPConfig, error) {
		wechatCfg, err := repo.GetWechatConfig(ctx)
		if err != nil {
			return nil, err
		}
		return config.LoadWechatConfigFromDB(ctx, wechatCfg), nil
	})
	// 启动时从数据库加载微信配置到缓存，确保重启后配置不丢失
	if _, err := wechatProvider.Get(context.Background()); err != nil {
		logger.Warnf("加载微信配置失败（使用默认空配置）: %v", err)
	} else {
		logger.Infof("微信配置加载成功")
	}
	logger.Infof("微信配置动态提供者初始化成功")

	logger.Infof("初始化应用服务...")
	authSvc := auth.NewService(repo, jwtMgr, verificationService, emailService, wechatProvider)
	tagSvc := tag.NewService(repo, bizCache)
	categorySvc := category.NewService(repo, bizCache)

	articleSvc := article.NewService(repo, bizCache)
	topicSvc := topic.NewService(repo)
	settingSvc := setting.NewService(repo, emailService, verificationService, bizCache)
	fileSvc := file.NewService(repo)
	dashboardSvc := dashboard.NewService(repo, bizCache)
	mpUserSvc := mpuserapp.NewService(repo, repo)
	linkRepo := mysqlrepo.NewLinkRepo(database)
	linkSvc := linkapp.NewLinkService(linkRepo)
	cacheSvc := cacheapp.NewService(bizCache)
	logger.Infof("应用服务初始化成功")

	logger.Infof("初始化HTTP处理器...")
	h := handler.New(handler.Services{
		Auth:      authSvc,
		Tag:       tagSvc,
		Category:  categorySvc,
		Article:   articleSvc,
		Topic:     topicSvc,
		Setting:   settingSvc,
		File:      fileSvc,
		Dashboard: dashboardSvc,
		MPUser:    mpUserSvc,
		Cache:     cacheSvc,
		Link:      linkSvc,
	})
	logger.Infof("HTTP处理器初始化成功")

	logger.Infof("初始化Gin引擎...")
	ginMode := gin.ReleaseMode
	if cfg.ServerMode == "debug" {
		ginMode = gin.DebugMode
	}
	gin.SetMode(ginMode)
	engine := gin.New()
	engine.Use(gin.Recovery()) // 仅保留 panic 恢复，不启用 Gin 内置日志（使用自定义 Logger 中间件）
	engine.MaxMultipartMemory = 8 << 20
	engine.Static("/uploads", "./uploads")

	// 注册访问日志中间件（记录所有请求到 t_access_log）
	logger.Infof("注册访问日志中间件...")
	engine.Use(middleware.AccessLogMiddleware(database))

	logger.Infof("注册路由...")
	router.Register(engine, h, jwtMgr, repo, redisClient)
	logger.Infof("路由注册成功")

	// 启动定时发布调度器
	logger.Infof("启动定时发布调度器...")
	publishScheduler := scheduler.NewScheduledPublishScheduler(repo, bizCache)
	publishScheduler.Start()
	logger.Infof("定时发布调度器启动成功")

	logger.Infof("初始化HTTP服务器，端口: %s", cfg.ServerPort)
	server := &http.Server{
		Addr:         ":" + cfg.ServerPort,
		Handler:      engine,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	return &App{cfg: cfg, server: server}, nil
}

func (a *App) Run() error {
	return a.server.ListenAndServe()
}

func (a *App) Shutdown(ctx context.Context) error {
	ctx2, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	return a.server.Shutdown(ctx2)
}
