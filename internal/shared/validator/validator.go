package validator

import (
	"regexp"
	"strings"

	apperr "sanmoo-server-go/internal/shared/errors"
)

func RequireNonBlank(value string, field string) error {
	if strings.TrimSpace(value) == "" {
		return apperr.New(apperr.ErrInvalidParam.Code, field+" 不能为空")
	}
	return nil
}

func IsEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}
