package security

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID   uint64 `json:"userId"`
	Username string `json:"username"`
	RoleID   uint64 `json:"roleId"`
	RoleName string `json:"roleName"`
	jwt.RegisteredClaims
}

type JWTManager struct {
	accessSecret  []byte
	refreshSecret []byte
	accessTTL     time.Duration
	refreshTTL    time.Duration
}

func NewJWTManager(accessSecret, refreshSecret string, accessTTLSeconds, refreshTTLSeconds int) *JWTManager {
	return &JWTManager{
		accessSecret:  []byte(accessSecret),
		refreshSecret: []byte(refreshSecret),
		accessTTL:     time.Duration(accessTTLSeconds) * time.Second,
		refreshTTL:    time.Duration(refreshTTLSeconds) * time.Second,
	}
}

func (m *JWTManager) GenerateAccessToken(userID uint64, username string, roleID uint64, roleName string) (string, error) {
	claims := Claims{UserID: userID, Username: username, RoleID: roleID, RoleName: roleName,
		RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.accessTTL)), IssuedAt: jwt.NewNumericDate(time.Now())}}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.accessSecret)
}

func (m *JWTManager) GenerateRefreshToken(userID uint64, username string, roleID uint64, roleName string) (string, error) {
	claims := Claims{UserID: userID, Username: username, RoleID: roleID, RoleName: roleName,
		RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.refreshTTL)), IssuedAt: jwt.NewNumericDate(time.Now())}}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.refreshSecret)
}

func (m *JWTManager) ParseAccessToken(tokenStr string) (*Claims, error) {
	return parseToken(tokenStr, m.accessSecret)
}

func (m *JWTManager) ParseRefreshToken(tokenStr string) (*Claims, error) {
	return parseToken(tokenStr, m.refreshSecret)
}

func parseToken(tokenStr string, secret []byte) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, fmt.Errorf("unexpected sign method")
		}
		return secret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return claims, nil
}
