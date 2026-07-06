package response

import (
	"errors"
	"time"

	apperr "sanmoo-server-go/internal/shared/errors"

	"github.com/gin-gonic/gin"
)

// CtxErrorDetailKey 用于在 gin.Context 中存储原始错误详情的键名。
const CtxErrorDetailKey = "errorDetail"

// Result 对应 OpenAPI 中的统一外层响应结构。
type Result struct {
	Success      bool        `json:"success"`
	Data         interface{} `json:"data"`
	ErrorCode    string      `json:"errorCode,omitempty"`
	ErrorMessage string      `json:"errorMessage,omitempty"`
	Timestamp    int64       `json:"timestamp"`
}

func Ok(c *gin.Context, data interface{}) {
	c.JSON(200, Result{
		Success:   true,
		Data:      data,
		Timestamp: time.Now().UnixMilli(),
	})
}

func Fail(c *gin.Context, err error) {
	code := apperr.ErrInternal.Code
	msg := apperr.ErrInternal.Message
	var ae *apperr.AppError
	if errors.As(err, &ae) {
		code = ae.Code
		msg = ae.Message
	}
	// 将原始错误信息存入上下文，供 access log middleware 写入 errorDetail
	c.Set(CtxErrorDetailKey, err.Error())
	// OpenAPI 只声明了 200 响应，因此错误也用统一的 200 + success=false 结构返回。
	c.JSON(200, Result{
		Success:      false,
		Data:         struct{}{},
		ErrorCode:    code,
		ErrorMessage: msg,
		Timestamp:    time.Now().UnixMilli(),
	})
}
