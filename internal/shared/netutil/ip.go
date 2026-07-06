package netutil

import (
	"net"
	"strings"

	"github.com/gin-gonic/gin"
)

// RealIP 从常见反向代理头中提取真实客户端 IP。
// 优先级：X-Forwarded-For -> X-Real-IP -> RemoteAddr(ClientIP)。
func RealIP(c *gin.Context) string {
	// 1) X-Forwarded-For: "client, proxy1, proxy2"
	if xff := strings.TrimSpace(c.GetHeader("X-Forwarded-For")); xff != "" {
		parts := strings.Split(xff, ",")
		for _, p := range parts {
			ip := strings.TrimSpace(p)
			if ip == "" || strings.EqualFold(ip, "unknown") {
				continue
			}
			return ip
		}
	}

	// 2) X-Real-IP
	if xri := strings.TrimSpace(c.GetHeader("X-Real-IP")); xri != "" && !strings.EqualFold(xri, "unknown") {
		return xri
	}

	// 3) gin 内置（会处理部分代理逻辑）
	ip := strings.TrimSpace(c.ClientIP())
	if ip != "" {
		return ip
	}

	// 4) RemoteAddr 兜底
	host, _, err := net.SplitHostPort(strings.TrimSpace(c.Request.RemoteAddr))
	if err == nil && host != "" {
		return host
	}
	return strings.TrimSpace(c.Request.RemoteAddr)
}
