package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

var (
	log       *logrus.Logger
	logFile   *os.File
	debugMode bool
)

func init() {
	// 创建日志目录
	logDir := "./logs"
	if err := os.MkdirAll(logDir, 0755); err != nil {
		fmt.Printf("创建日志目录失败: %v\n", err)
	}

	// 创建日志文件
	logPath := filepath.Join(logDir, fmt.Sprintf("app-%s.log", time.Now().Format("2006-01-02")))
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("打开日志文件失败: %v\n", err)
		file = os.Stdout
	}
	logFile = file

	// 初始化 logrus
	log = logrus.New()
	log.SetOutput(logFile)
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
		CallerPrettyfier: func(frame *runtime.Frame) (function string, file string) {
			funcName := filepath.Base(frame.Function)
			fileName := filepath.Base(frame.File)
			return funcName, fmt.Sprintf("%s:%d", fileName, frame.Line)
		},
	})
	log.SetReportCaller(true)

	// 默认日志级别为 info
	log.SetLevel(logrus.InfoLevel)

	// 检查环境变量，支持提前设置 debug 模式
	if strings.ToLower(os.Getenv("SERVER_MODE")) == "debug" {
		enableConsoleOutput()
	}
}

// SetMode 根据运行模式设置日志输出方式。
// debug 模式下同时输出到文件和控制台。
func SetMode(mode string) {
	if strings.ToLower(mode) == "debug" {
		enableConsoleOutput()
		log.SetLevel(logrus.DebugLevel)
	}
}

func enableConsoleOutput() {
	log.SetOutput(io.MultiWriter(logFile, os.Stdout))
	debugMode = true
}

// IsDebug 返回当前是否为调试模式。
func IsDebug() bool {
	return debugMode
}

// SetLevel 设置日志级别
func SetLevel(level string) {
	switch strings.ToLower(level) {
	case "debug":
		log.SetLevel(logrus.DebugLevel)
	case "info":
		log.SetLevel(logrus.InfoLevel)
	case "warn":
		log.SetLevel(logrus.WarnLevel)
	case "error":
		log.SetLevel(logrus.ErrorLevel)
	default:
		log.SetLevel(logrus.InfoLevel)
	}
}

// Debug 记录调试级别日志
func Debug(args ...interface{}) {
	log.Debug(args...)
}

// Debugf 记录格式化的调试级别日志
func Debugf(format string, args ...interface{}) {
	log.Debugf(format, args...)
}

// Info 记录信息级别日志
func Info(args ...interface{}) {
	log.Info(args...)
}

// Infof 记录格式化的信息级别日志
func Infof(format string, args ...interface{}) {
	log.Infof(format, args...)
}

// Warn 记录警告级别日志
func Warn(args ...interface{}) {
	log.Warn(args...)
}

// Warnf 记录格式化的警告级别日志
func Warnf(format string, args ...interface{}) {
	log.Warnf(format, args...)
}

// Error 记录错误级别日志
func Error(args ...interface{}) {
	log.Error(args...)
}

// Errorf 记录格式化的错误级别日志
func Errorf(format string, args ...interface{}) {
	log.Errorf(format, args...)
}

// Fatal 记录致命级别日志并退出程序
func Fatal(args ...interface{}) {
	log.Fatal(args...)
}

// Fatalf 记录格式化的致命级别日志并退出程序
func Fatalf(format string, args ...interface{}) {
	log.Fatalf(format, args...)
}

// GetLogger 获取日志实例
func GetLogger() *logrus.Logger {
	return log
}
