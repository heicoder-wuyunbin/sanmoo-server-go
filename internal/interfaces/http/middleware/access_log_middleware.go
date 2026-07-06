package middleware

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"sanmoo-server-go/internal/infrastructure/logger"
	"sanmoo-server-go/internal/infrastructure/repository/mysqlrepo"
	"sanmoo-server-go/internal/shared/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const accessLogMaxBodySize = 4096

type accessLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *accessLogWriter) Write(b []byte) (int, error) {
	if w.body.Len() < accessLogMaxBodySize {
		remaining := accessLogMaxBodySize - w.body.Len()
		if len(b) <= remaining {
			w.body.Write(b)
		} else {
			w.body.Write(b[:remaining])
		}
	}
	return w.ResponseWriter.Write(b)
}

func (w *accessLogWriter) WriteString(s string) (int, error) {
	if w.body.Len() < accessLogMaxBodySize {
		remaining := accessLogMaxBodySize - w.body.Len()
		if len(s) <= remaining {
			w.body.WriteString(s)
		} else {
			w.body.WriteString(s[:remaining])
		}
	}
	return w.ResponseWriter.WriteString(s)
}

// AccessLogMiddleware records each HTTP request into t_access_log and writes
// failed responses or panics into t_error_log.
func AccessLogMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 跳过健康检查和管理后台请求，只记录用户访问
		if c.Request.URL.Path == "/health" ||
			c.Request.URL.EscapedPath() == "/actuator/health" ||
			strings.HasPrefix(c.Request.URL.Path, "/admin") {
			c.Next()
			return
		}

		start := time.Now()
		traceID := newTraceID()
		c.Set("traceID", traceID)
		c.Header("X-Trace-ID", traceID)

		requestBody := readRequestBody(c)
		responseWriter := &accessLogWriter{
			ResponseWriter: c.Writer,
			body:           bytes.NewBufferString(""),
		}
		c.Writer = responseWriter

		var panicValue any
		var stackTrace string

		defer func() {

			duration := time.Since(start)
			responseTime := int(duration.Milliseconds())
			responseStatus := c.Writer.Status()
			ipAddress := getRealIP(c)
			userAgent := truncateLog(c.GetHeader("User-Agent"), 500)
			requestSource := truncateLog(identifyRequestSource(userAgent), 100)

			// 识别访问者身份：已登录用户使用用户名，未登录标记为游客
			var visitorUserID uint64
			visitorName := "游客"
			if uid, ok := c.Get(CtxUserIDKey); ok {
				if v, ok2 := uid.(uint64); ok2 {
					visitorUserID = v
				}
			}
			if uname, ok := c.Get(CtxUsernameKey); ok {
				if v, ok2 := uname.(string); ok2 && v != "" {
					visitorName = v
				}
			}

			responseBody := truncateLog(responseWriter.body.String(), accessLogMaxBodySize)
			errorCode, errorMessage, errorDetail := extractResponseError(responseStatus, responseBody)

			// 优先使用 handler 通过 response.Fail 存入上下文的原始错误信息
			if ctxDetail, ok := c.Get(response.CtxErrorDetailKey); ok {
				if s, ok2 := ctxDetail.(string); ok2 && s != "" {
					errorDetail = s
				}
			}

			if panicValue != nil {
				errorCode = "PANIC"
				errorMessage = fmt.Sprint(panicValue)
				errorDetail = stackTrace
			}

			isError := panicValue != nil || responseStatus >= http.StatusBadRequest || errorCode != ""
			requestURL := truncateLog(sanitizeText(c.Request.URL.RequestURI()), 500)
			requestQuery := sanitizeText(c.Request.URL.RawQuery)
			requestParams := buildRequestParams(requestQuery, requestBody)

			accessLog := mysqlrepo.TAccessLog{
				TraceID:        traceID,
				RequestMethod:  truncateLog(c.Request.Method, 12),
				RequestURL:     requestURL,
				RequestPath:    truncateLog(c.Request.URL.Path, 300),
				RequestQuery:   requestQuery,
				RequestParams:  requestParams,
				RequestBody:    requestBody,
				VisitorUserID:  visitorUserID,
				VisitorName:    visitorName,
				IPAddress:      parseIPBytes(ipAddress),
				RequestTime:    start,
				ResponseTime:   responseTime,
				ResponseStatus: responseStatus,
				ResponseBody:   responseBody,
				UserAgent:      userAgent,
				RequestSource:  requestSource,
				IsError:        isError,
				CreateTime:     time.Now(),
			}

			go func() {
				if err := db.Create(&accessLog).Error; err != nil {
					logger.Warnf("保存访问日志失败: %v", err)
					return
				}

				if !isError {
					return
				}

				errorLog := mysqlrepo.TErrorLog{
					AccessLogID:   accessLog.ID,
					TraceID:       traceID,
					ErrorCode:     errorCode,
					ErrorMessage:  truncateLog(errorMessage, 1000),
					ErrorDetail:   errorDetail,
					StackTrace:    stackTrace,
					RequestURL:    requestURL,
					RequestMethod: truncateLog(c.Request.Method, 12),
					RequestParams: requestParams,
					RequestBody:   requestBody,
					ResponseBody:  responseBody,
					IPAddress:     parseIPBytes(ipAddress),
					UserAgent:     userAgent,
					CreateTime:    time.Now(),
				}
				if err := db.Create(&errorLog).Error; err != nil {
					logger.Warnf("保存错误日志失败: %v", err)
					return
				}
				_ = db.Model(&mysqlrepo.TAccessLog{}).Where("id = ?", accessLog.ID).Update("error_id", errorLog.ID).Error
			}()

			if panicValue != nil {
				panic(panicValue)
			}
		}()

		c.Next()
	}
}

func readRequestBody(c *gin.Context) string {
	if c.Request.Body == nil || isMultipartContent(c.ContentType()) {
		return ""
	}

	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.Request.Body = io.NopCloser(bytes.NewBuffer(nil))
		return ""
	}
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	if len(bodyBytes) == 0 {
		return ""
	}
	return sanitizeBody(string(bodyBytes), c.ContentType())
}

func isMultipartContent(contentType string) bool {
	return strings.Contains(strings.ToLower(contentType), "multipart/form-data")
}

func sanitizeBody(body, contentType string) string {
	body = truncateLog(body, accessLogMaxBodySize)
	if strings.Contains(strings.ToLower(contentType), "json") ||
		strings.HasPrefix(strings.TrimSpace(body), "{") ||
		strings.HasPrefix(strings.TrimSpace(body), "[") {
		return sanitizeJSON(body)
	}
	return sanitizeText(body)
}

func sanitizeJSON(raw string) string {
	var v any
	if err := json.Unmarshal([]byte(raw), &v); err != nil {
		return sanitizeText(raw)
	}
	maskSensitiveJSON(v)
	b, err := json.Marshal(v)
	if err != nil {
		return sanitizeText(raw)
	}
	return truncateLog(string(b), accessLogMaxBodySize)
}

func maskSensitiveJSON(v any) {
	switch value := v.(type) {
	case map[string]any:
		for key, item := range value {
			if isSensitiveKey(key) {
				value[key] = "***"
				continue
			}
			maskSensitiveJSON(item)
		}
	case []any:
		for _, item := range value {
			maskSensitiveJSON(item)
		}
	}
}

func sanitizeText(raw string) string {
	out := raw
	for _, key := range []string{"password", "token", "accessToken", "refreshToken", "code", "authorization"} {
		out = maskTextValue(out, key)
	}
	return truncateLog(out, accessLogMaxBodySize)
}

func maskTextValue(raw, key string) string {
	lower := strings.ToLower(raw)
	lowerKey := strings.ToLower(key)
	idx := strings.Index(lower, lowerKey+"=")
	for idx >= 0 {
		start := idx + len(key) + 1
		end := start
		for end < len(raw) && raw[end] != '&' && raw[end] != ' ' && raw[end] != '\n' && raw[end] != '\r' {
			end++
		}
		raw = raw[:start] + "***" + raw[end:]
		lower = strings.ToLower(raw)
		idx = strings.Index(lower[start:], lowerKey+"=")
		if idx >= 0 {
			idx += start
		}
	}
	return raw
}

func isSensitiveKey(key string) bool {
	k := strings.ToLower(key)
	return strings.Contains(k, "password") ||
		strings.Contains(k, "token") ||
		strings.Contains(k, "secret") ||
		strings.Contains(k, "authorization") ||
		k == "code"
}

func buildRequestParams(query, body string) string {
	switch {
	case query != "" && body != "":
		return "query=" + query + "\nbody=" + body
	case query != "":
		return "query=" + query
	default:
		return body
	}
}

func extractResponseError(status int, responseBody string) (string, string, string) {
	if status >= http.StatusBadRequest {
		return fmt.Sprintf("HTTP_%d", status), http.StatusText(status), responseBody
	}
	if responseBody == "" {
		return "", "", ""
	}
	var result struct {
		Success      *bool  `json:"success"`
		ErrorCode    string `json:"errorCode"`
		ErrorMessage string `json:"errorMessage"`
	}
	if err := json.Unmarshal([]byte(responseBody), &result); err != nil {
		return "", "", ""
	}
	if result.Success != nil && !*result.Success {
		return result.ErrorCode, result.ErrorMessage, responseBody
	}
	return "", "", ""
}

func parseIPBytes(ipAddress string) []byte {
	ip := net.ParseIP(ipAddress)
	if ip == nil {
		return []byte(net.ParseIP("127.0.0.1"))
	}
	return []byte(ip)
}

func newTraceID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b[:])
}

func truncateLog(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "...[truncated]"
}

func getRealIP(c *gin.Context) string {
	if xff := c.GetHeader("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			ip := strings.TrimSpace(parts[0])
			if ip != "" && !strings.EqualFold(ip, "unknown") {
				return ip
			}
		}
	}

	if xri := c.GetHeader("X-Real-IP"); xri != "" && !strings.EqualFold(xri, "unknown") {
		return xri
	}

	if ip := c.ClientIP(); ip != "" {
		return ip
	}

	host, _, err := net.SplitHostPort(strings.TrimSpace(c.Request.RemoteAddr))
	if err == nil && host != "" {
		return host
	}

	return strings.TrimSpace(c.Request.RemoteAddr)
}

func identifyRequestSource(userAgent string) string {
	if userAgent == "" {
		return "Unknown"
	}

	ua := strings.ToLower(userAgent)

	if strings.Contains(ua, "micromessenger") {
		if strings.Contains(ua, "miniprogram") {
			return "WeChat MiniProgram"
		}
		return "WeChat"
	}

	if strings.Contains(ua, "mobile") || strings.Contains(ua, "android") || strings.Contains(ua, "iphone") {
		if strings.Contains(ua, "android") {
			return "Android Mobile"
		}
		if strings.Contains(ua, "iphone") || strings.Contains(ua, "ipad") {
			return "iOS Device"
		}
		return "Mobile Browser"
	}

	if strings.Contains(ua, "edge") {
		return "Edge Browser"
	}
	if strings.Contains(ua, "chrome") {
		return "Chrome Browser"
	}
	if strings.Contains(ua, "firefox") {
		return "Firefox Browser"
	}
	if strings.Contains(ua, "safari") {
		return "Safari Browser"
	}
	if strings.Contains(ua, "msie") || strings.Contains(ua, "trident") {
		return "Internet Explorer"
	}

	if strings.Contains(ua, "bot") || strings.Contains(ua, "spider") || strings.Contains(ua, "crawler") {
		return "Search Engine Bot"
	}

	return "Other Browser"
}
