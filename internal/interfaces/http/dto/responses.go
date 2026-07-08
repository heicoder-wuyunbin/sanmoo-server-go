package dto

import domarticle "sanmoo-server-go/internal/domain/article"
import domdashboard "sanmoo-server-go/internal/domain/dashboard"

// EmptyResponse 用于无返回体的成功响应（替代 map[string]any{}）。
type EmptyResponse struct{}

type IDResponse struct {
	ID uint64 `json:"id"`
}

type BoolResponse struct {
	Value bool `json:"value"`
}

// PageResponse 通用分页响应结构。
type PageResponse[T any] struct {
	List  []T   `json:"list"`
	Total int64 `json:"total"`
	Page  int   `json:"page"`
	Size  int   `json:"size"`
}

// ListResponse 通用列表响应结构。
type ListResponse[T any] struct {
	List []T `json:"list"`
}

// SettingsResponse 后台/前台设置（保持字段命名与前端一致）。
// 注意：各 config 内部字段较多且动态变化，因此用 map 承载。
type SettingsResponse struct {
	CoreConfig    map[string]any `json:"coreConfig"`
	UIConfig      map[string]any `json:"uiConfig"`
	StorageConfig map[string]any `json:"storageConfig"`
	EmailConfig   map[string]any `json:"emailConfig,omitempty"`
}

type FileUploadResponse struct {
	URL  string `json:"url"`
	Size int64  `json:"size"`
}

// ArticleDetailResponse Web 端文章详情。ContentHtml 为后端统一渲染的 HTML。
type ArticleDetailResponse struct {
	Article     *domarticle.Article `json:"article"`
	ContentHtml string              `json:"contentHtml"`
	TOC         []TOCItem           `json:"toc"`
	PrevArticle *domarticle.Article `json:"prevArticle"`
	NextArticle *domarticle.Article `json:"nextArticle"`
}

// TOCItem 目录项
type TOCItem struct {
	Level int    `json:"level"`
	Text  string `json:"text"`
	ID    string `json:"id"`
}

type MPUserProfileResponse struct {
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}

type MPFavoriteStatusResponse struct {
	IsFavorited bool `json:"isFavorited"`
}

type DashboardResponse struct {
	Dashboard *domdashboard.Statistics `json:"dashboard"`
}

// MPSettingsResponse 小程序端设置（字段裁剪版本）
type MPSettingsResponse struct {
	CoreConfig struct {
		BlogName     any `json:"blogName"`
		Author       any `json:"author"`
		Introduction any `json:"introduction"`
		Avatar       any `json:"avatar"`
		Poster       any `json:"poster"`
	} `json:"coreConfig"`
	UIConfig struct {
		GithubHome        any `json:"githubHome"`
		CsdnHome          any `json:"csdnHome"`
		GiteeHome         any `json:"giteeHome"`
		ZhihuHome         any `json:"zhihuHome"`
		GithubShow        any `json:"githubShow"`
		CsdnShow          any `json:"csdnShow"`
		GiteeShow         any `json:"giteeShow"`
		ZhihuShow         any `json:"zhihuShow"`
		RecommendStrategy any `json:"recommendStrategy"`
	} `json:"uiConfig"`
	StorageConfig struct {
		UploadStrategy       any `json:"uploadStrategy"`
		UploadLocalUrlPrefix any `json:"uploadLocalUrlPrefix"`
	} `json:"storageConfig"`
}

// MPArticleDetailResponse 小程序端文章详情。
// ContentHtml 为后端统一渲染的 HTML（小程序端直接渲染此字段）。
type MPArticleDetailResponse struct {
	ID           uint64              `json:"id"`
	Title        string              `json:"title"`
	TitleImage   string              `json:"titleImage"`
	Description  string              `json:"description"`
	ContentHtml  string              `json:"contentHtml"`
	CreateTime   any                 `json:"createTime"`
	UpdateTime   any                 `json:"updateTime"`
	ReadNum      int                 `json:"readNum"`
	CategoryID   uint64              `json:"categoryId"`
	CategoryName string              `json:"categoryName"`
	Tags         []domarticle.TagRef `json:"tags"`
	PrevArticle  *domarticle.Article `json:"prevArticle,omitempty"`
	NextArticle  *domarticle.Article `json:"nextArticle,omitempty"`
	IsFavorited  bool                `json:"isFavorited"`
}

type ArticleOption struct {
	ID    uint64 `json:"id"`
	Title string `json:"title"`
}

type TopicArticleIDsResponse struct {
	ArticleIDs []uint64 `json:"articleIds"`
}

// MapResponse 通用 map 响应结构。
type MapResponse struct {
	Data map[string]interface{} `json:"data"`
}

type EmailVerificationResponse struct {
	Identifier string `json:"identifier"`
}
