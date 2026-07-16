package mpuser

import "time"

// MPUserSummary 微信用户列表摘要（轻运营：仅保留基础字段）。
type MPUserSummary struct {
	ID             uint64    `json:"id"`
	OpenID         string    `json:"openid"`
	Nickname       string    `json:"nickname"`
	Avatar         string    `json:"avatar"`
	Status         uint8     `json:"status"`
	FirstLoginTime time.Time `json:"firstLoginTime"`
	LastLoginTime  time.Time `json:"lastLoginTime"`
	ViewCount      int64     `json:"viewCount"`
	FavoriteCount  int64     `json:"favoriteCount"`
	CreateTime     time.Time `json:"createTime"`
}