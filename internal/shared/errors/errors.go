package errors

import "fmt"

// AppError 是统一业务错误对象。
// 约定：应用层返回 AppError，HTTP 层统一转成 OpenAPI 的 ResultObject。
type AppError struct {
	Code    string
	Message string
}

func (e *AppError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func New(code, message string) *AppError {
	return &AppError{Code: code, Message: message}
}

var (
	ErrInvalidParam  = New("INVALID_PARAM", "请求参数不合法")
	ErrUnauthorized  = New("UNAUTHORIZED", "未登录或登录已过期")
	ErrForbidden     = New("FORBIDDEN", "无权限访问")
	ErrNotFound      = New("NOT_FOUND", "资源不存在")
	ErrConflict      = New("CONFLICT", "资源冲突")
	ErrInternal      = New("INTERNAL_ERROR", "服务器内部错误")
	ErrBadCredential = New("BAD_CREDENTIAL", "用户名或密码错误")
	ErrBadVerifyCode = New("BAD_VERIFY_CODE", "验证码错误")
	ErrMFARequired   = New("MFA_REQUIRED", "需要邮箱验证码")
	ErrEmailNotVerified = New("EMAIL_NOT_VERIFIED", "请先完成邮箱验证码验证")
	ErrUserDisabled  = New("USER_DISABLED", "用户已被封禁")
)
