package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"sanmoo-server-go/internal/bootstrap"
	"sanmoo-server-go/internal/infrastructure/config"
	"sanmoo-server-go/internal/infrastructure/logger"
)

// main 负责应用启动与优雅停机。
func main() {
	// 设置时区为东八区（UTC+8）
	os.Setenv("TZ", "Asia/Shanghai")
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		println("设置时区失败:", err.Error())
	} else {
		time.Local = loc
		println("时区设置成功: Asia/Shanghai")
	}

	// 先打印简单的启动信息，确保至少能看到输出
	println("开始启动应用...")

	// 尝试加载配置
	cfg, err := config.LoadFromProperties("application.properties")
	if err != nil {
		println("加载配置失败:", err.Error())
		panic(err)
	}
	println("配置加载成功")

	// 根据配置设置日志模式
	logger.SetMode(cfg.ServerMode)

	// 现在初始化日志系统
	println("初始化日志系统...")
	logger.Infof("配置加载成功，服务器端口: %s，模式: %s", cfg.ServerPort, cfg.ServerMode)

	// 初始化应用
	println("初始化应用...")
	app, err := bootstrap.New(cfg)
	if err != nil {
		println("初始化应用失败:", err.Error())
		logger.Fatalf("初始化应用失败: %v", err)
	}
	println("应用初始化成功")
	logger.Infof("应用初始化成功")

	go func() {
		println("服务器开始运行...")
		logger.Infof("服务器开始运行")
		if err := app.Run(); err != nil {
			println("服务器停止:", err.Error())
			logger.Errorf("服务器停止: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	_ = app.Shutdown(context.Background())
}
