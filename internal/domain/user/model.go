package user

import (
	"time"

	apperr "sanmoo-server-go/internal/shared/errors"
	"sanmoo-server-go/internal/shared/validator"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID                uint64     `json:"id"`
	Username          string     `json:"username"`
	PasswordHash      string     `json:"-"`
	Email             string     `json:"email"`
	Nickname          string     `json:"nickname"`
	Avatar            string     `json:"avatar"`
	Status            string     `json:"status"`
	LastLoginTime     *time.Time `json:"lastLoginTime,omitempty"`
	LastLoginIp       string     `json:"lastLoginIp"`
	LoginFailureCount uint       `json:"loginFailureCount"`
	LockedUntil       *time.Time `json:"lockedUntil,omitempty"`
	RoleID            uint64     `json:"roleId"`
	RoleName          string     `json:"roleName"`
	CreateTime        time.Time  `json:"createTime"`
	UpdateTime        time.Time  `json:"updateTime"`
}

func (u *User) ValidateForCreate() error {
	if err := validator.RequireNonBlank(u.Username, "username"); err != nil {
		return err
	}
	if len(u.Username) > 60 {
		return apperr.New(apperr.ErrInvalidParam.Code, "username 长度不能超过 60")
	}
	if err := validator.RequireNonBlank(u.Email, "email"); err != nil {
		return err
	}
	if !validator.IsEmail(u.Email) {
		return apperr.New(apperr.ErrInvalidParam.Code, "email 格式不正确")
	}
	if len(u.Email) > 100 {
		return apperr.New(apperr.ErrInvalidParam.Code, "email 长度不能超过 100")
	}
	if len(u.Nickname) > 60 {
		return apperr.New(apperr.ErrInvalidParam.Code, "nickname 长度不能超过 60")
	}
	if len(u.Avatar) > 160 {
		return apperr.New(apperr.ErrInvalidParam.Code, "avatar 长度不能超过 160")
	}
	return nil
}

func (u *User) SetPassword(plain string) error {
	if len(plain) < 6 {
		return apperr.New(apperr.ErrInvalidParam.Code, "password 长度至少 6")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		return apperr.ErrInternal
	}
	u.PasswordHash = string(hash)
	return nil
}

func (u *User) VerifyPassword(plain string) bool {
	return bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(plain)) == nil
}
