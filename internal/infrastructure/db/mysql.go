package db

import (
	"time"

	"sanmoo-server-go/internal/infrastructure/logger"
	"sanmoo-server-go/internal/infrastructure/repository/mysqlrepo"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func NewMySQL(dsn string) (*gorm.DB, error) {
	logger.Infof("正在连接数据库，DSN: %s", maskDSN(dsn))
	lg := gormlogger.New(
		&gormLogAdapter{},
		gormlogger.Config{
			LogLevel:                  gormlogger.Warn,
			IgnoreRecordNotFoundError: true,
		},
	)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{Logger: lg})
	if err != nil {
		logger.Errorf("数据库连接失败: %v", err)
		return nil, err
	}

	// 配置连接池以适配 2C2G 服务器资源
	sqlDB, err := db.DB()
	if err != nil {
		logger.Errorf("获取数据库连接池失败: %v", err)
		return nil, err
	}
	sqlDB.SetMaxOpenConns(15)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)
	sqlDB.SetConnMaxIdleTime(2 * time.Minute)

	logger.Infof("数据库连接成功")

	if err := db.AutoMigrate(&mysqlrepo.TMPUserSubscribe{}); err != nil {
		logger.Errorf("自动迁移 t_mp_user_subscribe 表失败: %v", err)
		return nil, err
	}
	logger.Infof("t_mp_user_subscribe 表自动迁移完成")

	return db, nil
}

// SetGormLogMode 允许外部设置 GORM 日志级别（debug 模式下启用 SQL 日志）。
func SetGormLogMode(db *gorm.DB, mode string) {
	level := gormlogger.Warn
	if mode == "debug" {
		level = gormlogger.Info
	}
	db.Logger = gormlogger.New(
		&gormLogAdapter{},
		gormlogger.Config{
			LogLevel:                  level,
			IgnoreRecordNotFoundError: true,
		},
	)
	logger.Infof("GORM 日志级别设置为: %s", mode)
}

// maskDSN 遮蔽DSN中的敏感信息
func maskDSN(dsn string) string {
	return "[MASKED]"
}

// gormLogAdapter 实现gorm logger.Writer接口，将gorm日志重定向到我们的日志系统
type gormLogAdapter struct{}

func (l *gormLogAdapter) Printf(format string, args ...interface{}) {
	logger.Infof("[GORM] "+format, args...)
}
