package middleware

import (
	"bytes"
	"io"
	"strings"
	"time"

	"sanmoo-server-go/internal/infrastructure/logger"

	"github.com/gin-gonic/gin"
)

const maxBodyLogSize = 4096 // 最多捕获 4KB 的 body，防止日志膨胀

// bodyLogWriter 包装 gin.ResponseWriter，截获响应体内容。
type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *bodyLogWriter) Write(b []byte) (int, error) {
	// 限制捕获大小
	if w.body.Len() < maxBodyLogSize {
		remaining := maxBodyLogSize - w.body.Len()
		if len(b) < remaining {
			w.body.Write(b)
		} else {
			w.body.Write(b[:remaining])
		}
	}
	return w.ResponseWriter.Write(b)
}

// Logger 日志中间件。debug 模式下输出控制台 + 日志文件，并记录请求体与响应体。
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 跳过健康检查请求，避免频繁日志产生 CPU 开销
		if c.Request.URL.Path == "/health" || c.Request.URL.Path == "/actuator/health" {
			c.Next()
			return
		}

		start := time.Now()
		path := c.Request.URL.Path
		rawQuery := c.Request.URL.RawQuery
		method := c.Request.Method
		ip := c.ClientIP()
		contentType := c.Request.Header.Get("Content-Type")

		var (
			requestBody  string
			responseBody string
			blw          *bodyLogWriter
		)

		if logger.IsDebug() && isJSONContent(contentType) {
			// 读取请求体并恢复
			bodyBytes, err := io.ReadAll(c.Request.Body)
			if err == nil && len(bodyBytes) > 0 {
				requestBody = truncate(string(bodyBytes), maxBodyLogSize)
			}
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

			// 包装响应写入器以截获响应体
			blw = &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
			c.Writer = blw
		}

		c.Next()

		end := time.Now()
		latency := end.Sub(start)
		status := c.Writer.Status()

		if logger.IsDebug() {
			if blw != nil {
				responseBody = blw.body.String()
			}
			q := ""
			if rawQuery != "" {
				q = " query=" + rawQuery
			}
			req := ""
			if requestBody != "" {
				req = " req=" + requestBody
			}
			resp := ""
			if responseBody != "" {
				resp = " resp=" + responseBody
			}
			logger.Debugf("[HTTP] %s %s%s%s%s | %d | %v | %s", method, path, q, req, resp, status, latency, ip)
			return
		}

		// 生产模式保持原有行为
		if status >= 500 {
			logger.Errorf("[HTTP] %s %s %s %d %v", method, path, ip, status, latency)
		} else if status >= 400 {
			logger.Warnf("[HTTP] %s %s %s %d %v", method, path, ip, status, latency)
		} else {
			logger.Infof("[HTTP] %s %s %s %d %v", method, path, ip, status, latency)
		}
	}
}

func isJSONContent(ct string) bool {
	return ct == "" || strings.Contains(ct, "json")
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "...[truncated]"
}
