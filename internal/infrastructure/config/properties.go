package config

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	ServerPort string
	ServerMode string // debug 或 release
	MySQLDSN   string
	RedisAddr  string
	JWT        JWTConfig
}

type JWTConfig struct {
	AccessSecret      string
	AccessTTLSeconds  int
	RefreshSecret     string
	RefreshTTLSeconds int
}

type WechatMPConfig struct {
	AppID  string
	Secret string
}

// LoadFromProperties 读取 application.properties 配置文件
func LoadFromProperties(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open properties file failed: %w", err)
	}
	defer f.Close()

	// 读取配置文件到映射
	configMap := make(map[string]string)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		configMap[key] = value
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read properties file failed: %w", err)
	}

	// 解析数据库配置
	mysqlConfig := parseMySQLConfig(configMap)

	// 解析Redis配置
	redisAddr := parseRedisConfig(configMap)

	// 解析JWT配置
	jwtConfig := parseJWTConfig(configMap)

	// 构建配置对象
	cfg := &Config{
		ServerPort: getConfigValue(configMap, "server.port", "28080"),
		ServerMode: getConfigValue(configMap, "server.mode", "release"),
		MySQLDSN:   mysqlConfig,
		RedisAddr:  redisAddr,
		JWT:        jwtConfig,
	}

	return cfg, nil
}

// parseMySQLConfig 解析数据库配置并构建DSN
func parseMySQLConfig(configMap map[string]string) string {
	// 从环境变量或配置文件中读取数据库配置
	host := getEnvOrConfig(configMap, "MYSQL_HOST", "mysql.host", "localhost")
	port := getEnvOrConfig(configMap, "MYSQL_PORT", "mysql.port", "3306")
	dbName := getEnvOrConfig(configMap, "MYSQL_DATABASE", "mysql.database", "sanmoo_blog")
	user := getEnvOrConfig(configMap, "MYSQL_USERNAME", "mysql.username", "root")
	password := getEnvOrConfig(configMap, "MYSQL_PASSWORD", "mysql.password", "root")

	// 构建MySQL DSN
	// timeout: 连接超时 5s, readTimeout: 读超时 10s, writeTimeout: 写超时 5s
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local&timeout=5s&readTimeout=10s&writeTimeout=5s",
		user, password, host, port, dbName)
}

// parseRedisConfig 解析Redis配置
func parseRedisConfig(configMap map[string]string) string {
	host := getEnvOrConfig(configMap, "REDIS_HOST", "redis.host", "localhost")
	port := getEnvOrConfig(configMap, "REDIS_PORT", "redis.port", "6379")
	return fmt.Sprintf("%s:%s", host, port)
}

// parseJWTConfig 解析JWT配置
func parseJWTConfig(configMap map[string]string) JWTConfig {
	accessSecret := getEnvOrConfig(configMap, "JWT_ACCESS_SECRET", "jwt.access-secret", "6Z3kXyH8n3u0YqkQh0+J5O4o6DkP8o7gXwV5d2x8y1A=")
	refreshSecret := getEnvOrConfig(configMap, "JWT_REFRESH_SECRET", "jwt.refresh-secret", "7s9ZqYF2xH6uK8Lk2mB4P5Qw3eR9T1aU6vC0dJgH2N8=")

	accessTTL, _ := strconv.Atoi(getEnvOrConfig(configMap, "JWT_ACCESS_TTL", "jwt.access-ttl", "7200"))
	refreshTTL, _ := strconv.Atoi(getEnvOrConfig(configMap, "JWT_REFRESH_TTL", "jwt.refresh-ttl", "604800"))

	return JWTConfig{
		AccessSecret:      accessSecret,
		AccessTTLSeconds:  accessTTL,
		RefreshSecret:     refreshSecret,
		RefreshTTLSeconds: refreshTTL,
	}
}

// getEnvOrConfig 优先从环境变量读取配置，其次从配置文件读取
func getEnvOrConfig(configMap map[string]string, envKey, configKey, defaultValue string) string {
	// 优先从环境变量读取
	if value := os.Getenv(envKey); value != "" {
		return value
	}

	// 从配置文件读取
	if value := configMap[configKey]; value != "" {
		return value
	}

	// 返回默认值
	return defaultValue
}

// getConfigValue 从配置文件读取值，如果不存在则返回默认值
func getConfigValue(configMap map[string]string, key, defaultValue string) string {
	if value := configMap[key]; value != "" {
		return value
	}
	return defaultValue
}
