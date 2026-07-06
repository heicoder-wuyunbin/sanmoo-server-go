package email

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"os"
	"strings"
	"sync"
	"time"

	"sanmoo-server-go/internal/infrastructure/logger"
)

// EmailService 邮件服务
type EmailService struct {
	mu         sync.RWMutex
	from       string
	host       string
	port       string
	username   string
	password   string
	mfaEnabled bool
	// 配置来源：1. 数据库配置 2. 环境变量 3. 默认值
}

type smtpConfig struct {
	from     string
	host     string
	port     string
	username string
	password string
}

// NewEmailService 创建邮件服务实例
func NewEmailService() *EmailService {
	return &EmailService{
		from:     getEnv("EMAIL_FROM", "no-reply@example.com"),
		host:     getEnv("EMAIL_HOST", "smtp.example.com"),
		port:     getEnv("EMAIL_PORT", "587"),
		username: getEnv("EMAIL_USERNAME", ""),
		password: getEnv("EMAIL_PASSWORD", ""),
	}
}

// UpdateConfig 从数据库配置更新邮件服务配置
func (s *EmailService) UpdateConfig(config map[string]any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if host, ok := config["host"].(string); ok && host != "" {
		s.host = host
	}
	if port, ok := config["port"].(string); ok && port != "" {
		s.port = port
	}
	if username, ok := config["username"].(string); ok && username != "" {
		s.username = username
	}
	if password, ok := config["password"].(string); ok && password != "" {
		s.password = password
	}
	if from, ok := config["from"].(string); ok && from != "" {
		s.from = from
	}
	if enabled, ok := config["loginMfaEnabled"].(bool); ok {
		s.mfaEnabled = enabled
	}
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func (s *EmailService) currentConfig() smtpConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return smtpConfig{
		from:     s.from,
		host:     s.host,
		port:     s.port,
		username: s.username,
		password: s.password,
	}
}

func smtpConfigFromMap(config map[string]any) smtpConfig {
	getString := func(key string) string {
		if value, ok := config[key].(string); ok {
			return strings.TrimSpace(value)
		}
		return ""
	}

	return smtpConfig{
		from:     getString("from"),
		host:     getString("host"),
		port:     getString("port"),
		username: getString("username"),
		password: getString("password"),
	}
}

// shouldLogVerificationCode 仅在非生产环境输出验证码，方便本地开发调试。
func shouldLogVerificationCode() bool {
	envValues := []string{
		os.Getenv("APP_ENV"),
		os.Getenv("GO_ENV"),
		os.Getenv("ENV"),
		os.Getenv("GIN_MODE"),
	}

	hasExplicitEnv := false
	for _, value := range envValues {
		mode := strings.ToLower(strings.TrimSpace(value))
		if mode == "" {
			continue
		}
		hasExplicitEnv = true
		switch mode {
		case "prod", "production", "release", "staging", "stage":
			return false
		case "dev", "development", "local", "debug", "test":
			return true
		}
	}

	// 未显式配置环境时，默认按本地开发处理。
	return !hasExplicitEnv
}

// SendVerificationCode 发送验证码邮件
func (s *EmailService) SendVerificationCode(to string, code string) error {
	return s.sendVerificationCode(s.currentConfig(), to, code)
}

// SendVerificationCodeWithConfig 使用临时 SMTP 配置发送验证码，不污染全局服务状态。
func (s *EmailService) SendVerificationCodeWithConfig(config map[string]any, to string, code string) error {
	return s.sendVerificationCode(smtpConfigFromMap(config), to, code)
}

func (s *EmailService) sendVerificationCode(cfg smtpConfig, to string, code string) error {
	if shouldLogVerificationCode() {
		logger.Infof("开发环境登录验证码: email=%s code=%s", to, code)
	}

	// 构建邮件内容
	subject := "登录验证码"
	body := fmt.Sprintf(`
		<!DOCTYPE html>
		<html>
		<head>
			<meta charset="UTF-8">
			<title>登录验证码</title>
		</head>
		<body>
			<h1>登录验证码</h1>
			<p>您好，您正在尝试登录系统，以下是您的验证码：</p>
			<p style="font-size: 24px; font-weight: bold; margin: 20px 0;">%s</p>
			<p>验证码有效期为 15 分钟，请及时使用。</p>
			<p>如果您没有发起此操作，请忽略此邮件。</p>
			<p>发送时间：%s</p>
		</body>
		</html>
	`, code, time.Now().Format("2006-01-02 15:04:05"))

	// 构建邮件头
	message := fmt.Sprintf("From: %s\r\n", cfg.from)
	message += fmt.Sprintf("To: %s\r\n", to)
	message += fmt.Sprintf("Subject: %s\r\n", subject)
	message += "MIME-Version: 1.0\r\n"
	message += "Content-Type: text/html; charset=UTF-8\r\n"
	message += "\r\n"
	message += body

	return sendMail(cfg, to, message)
}

func sendMail(cfg smtpConfig, to string, message string) error {
	// 认证信息
	auth := smtp.PlainAuth("", cfg.username, cfg.password, cfg.host)

	addr := fmt.Sprintf("%s:%s", cfg.host, cfg.port)
	// QQ 邮箱常用 465(SSL) 或 587(STARTTLS)。net/smtp 的 SendMail 不支持 465 隐式 TLS，
	// 这里按端口做兼容处理。
	if cfg.port == "465" {
		conn, err := tls.Dial("tcp", addr, &tls.Config{ServerName: cfg.host})
		if err != nil {
			return err
		}
		c, err := smtp.NewClient(conn, cfg.host)
		if err != nil {
			return err
		}
		defer c.Close()
		if err := c.Auth(auth); err != nil {
			return err
		}
		if err := c.Mail(cfg.from); err != nil {
			return err
		}
		if err := c.Rcpt(to); err != nil {
			return err
		}
		w, err := c.Data()
		if err != nil {
			return err
		}
		if _, err := w.Write([]byte(message)); err != nil {
			_ = w.Close()
			return err
		}
		_ = w.Close()
		return c.Quit()
	}

	// 非 465：优先走 STARTTLS（如果服务端支持）
	c, err := smtp.Dial(addr)
	if err != nil {
		return err
	}
	defer c.Close()
	_ = c.Hello("localhost")
	if ok, _ := c.Extension("STARTTLS"); ok {
		if err := c.StartTLS(&tls.Config{ServerName: cfg.host}); err != nil {
			return err
		}
	}
	if err := c.Auth(auth); err != nil {
		return err
	}
	if err := c.Mail(cfg.from); err != nil {
		return err
	}
	if err := c.Rcpt(to); err != nil {
		return err
	}
	w, err := c.Data()
	if err != nil {
		return err
	}
	if _, err := w.Write([]byte(message)); err != nil {
		_ = w.Close()
		return err
	}
	_ = w.Close()
	return c.Quit()
}

// SendTestEmail 发送测试邮件，用于验证 SMTP 参数是否配置正确。
func (s *EmailService) SendTestEmail(to string) error {
	cfg := s.currentConfig()

	subject := "SMTP 配置测试"
	body := fmt.Sprintf(`
		<!DOCTYPE html>
		<html>
		<head>
			<meta charset="UTF-8">
			<title>SMTP 配置测试</title>
		</head>
		<body>
			<h1>SMTP 配置测试</h1>
			<p>如果你收到这封邮件，说明后台 SMTP 配置已生效。</p>
			<p>发送时间：%s</p>
		</body>
		</html>
	`, time.Now().Format("2006-01-02 15:04:05"))

	message := fmt.Sprintf("From: %s\r\n", cfg.from)
	message += fmt.Sprintf("To: %s\r\n", to)
	message += fmt.Sprintf("Subject: %s\r\n", subject)
	message += "MIME-Version: 1.0\r\n"
	message += "Content-Type: text/html; charset=UTF-8\r\n"
	message += "\r\n"
	message += body

	return sendMail(cfg, to, message)
}

// IsConfigured 检查邮件服务是否配置正确
func (s *EmailService) IsConfigured() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.host != "" && s.port != "" && s.username != "" && s.password != ""
}

// IsMFAEnabled 是否启用后台登录邮箱验证码。
func (s *EmailService) IsMFAEnabled() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.mfaEnabled
}
