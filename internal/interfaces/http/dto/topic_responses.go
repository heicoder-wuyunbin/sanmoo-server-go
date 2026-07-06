package dto

// TopicItem 小程序专题基本信息。
type TopicItem struct {
	ID           uint64 `json:"id"`
	Name         string `json:"title"`
	Description  string `json:"description"`
	CoverImage   string `json:"cover"`
	CreateTime   any    `json:"createTime"`
	ArticleCount int64  `json:"articleCount"`
}

type TopicDetailResponse struct {
	Topic TopicItem `json:"topic"`
}
