package mpuser

import "time"

// MPUserSummary 微信用户列表摘要（后端返回给管理端）。
type MPUserSummary struct {
	ID            uint64    `json:"id"`
	OpenID        string    `json:"openid"`
	Nickname      string    `json:"nickname"`
	Avatar        string    `json:"avatar"`
	Status        uint8     `json:"status"`
	FirstLoginTime time.Time `json:"firstLoginTime"`
	LastLoginTime time.Time `json:"lastLoginTime"`
	TagCount      int       `json:"tagCount"`
	ViewCount     int64     `json:"viewCount"`
	FavoriteCount int64     `json:"favoriteCount"`
	CreateTime    time.Time `json:"createTime"`
}

// MPUserDetail 微信用户详情。
type MPUserDetail struct {
	ID            uint64     `json:"id"`
	OpenID        string     `json:"openid"`
	Nickname      string     `json:"nickname"`
	Avatar        string     `json:"avatar"`
	Status        uint8      `json:"status"`
	FirstLoginTime time.Time  `json:"firstLoginTime"`
	LastLoginTime time.Time  `json:"lastLoginTime"`
	ViewCount     int64      `json:"viewCount"`
	FavoriteCount int64      `json:"favoriteCount"`
	TotalStaySeconds int64   `json:"totalStaySeconds"`
	Tags          []UserTag  `json:"tags"`
	Interests     []UserInterest `json:"interests"`
	Profile       *UserProfile    `json:"profile"`
	CreateTime    time.Time  `json:"createTime"`
}

// UserTag 用户标签。
type UserTag struct {
	ID          uint64    `json:"id"`
	TagName     string    `json:"tagName"`
	TagCategory string    `json:"tagCategory"`
	Score       float64   `json:"score"`
	Source      string    `json:"source"`
	CreateTime  time.Time `json:"createTime"`
}

// UserInterest 用户兴趣维度（来自 t_mp_user_interest）。
type UserInterest struct {
	DimensionType string  `json:"dimensionType"`
	DimensionID   uint64  `json:"dimensionId"`
	DimensionName string  `json:"dimensionName"`
	Score         float64 `json:"score"`
}

// ProfileDimension 六边形画像单个维度。
type ProfileDimension struct {
	Dimension string  `json:"dimension"`
	Score     float64 `json:"score"`
}

// UserProfile 用户六边形画像。
type UserProfile struct {
	Dimensions []ProfileDimension `json:"dimensions"`
	UpdatedAt  *time.Time         `json:"updatedAt,omitempty"`
}

// RadarData 雷达图刷新结果（行为标签 + 兴趣维度 + 六边形画像）。
type RadarData struct {
	Tags      []UserTag      `json:"tags"`
	Interests []UserInterest `json:"interests"`
	Profile   *UserProfile   `json:"profile"`
}