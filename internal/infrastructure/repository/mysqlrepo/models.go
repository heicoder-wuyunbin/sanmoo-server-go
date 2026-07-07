package mysqlrepo

import (
	"time"

	"sync"

	"gorm.io/gorm"
)

type Repository struct {
	db              *gorm.DB
	uploadStrategy  string
	uploadLocalDir  string
	uploadURLPrefix string
	qiniuAccessKey  string
	qiniuSecretKey  string
	qiniuBucket     string
	qiniuDomain     string
	mu              sync.RWMutex
}

func New(db *gorm.DB, uploadLocalDir, uploadURLPrefix string, qiniuAccessKey, qiniuSecretKey, qiniuBucket, qiniuDomain string) *Repository {
	return &Repository{
		db:              db,
		uploadStrategy:  "LOCAL",
		uploadLocalDir:  uploadLocalDir,
		uploadURLPrefix: uploadURLPrefix,
		qiniuAccessKey:  qiniuAccessKey,
		qiniuSecretKey:  qiniuSecretKey,
		qiniuBucket:     qiniuBucket,
		qiniuDomain:     qiniuDomain,
	}
}

func (r *Repository) getUploadStrategy() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.uploadStrategy
}

func (r *Repository) getUploadLocalDir() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.uploadLocalDir
}

func (r *Repository) getUploadURLPrefix() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.uploadURLPrefix
}

func (r *Repository) getQiniuConfig() (accessKey, secretKey, bucket, domain string) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.qiniuAccessKey, r.qiniuSecretKey, r.qiniuBucket, r.qiniuDomain
}

func (r *Repository) setUploadStorageConfig(uploadStrategy, localDir, urlPrefix, qiniuBucket, qiniuDomain, qiniuAccessKey, qiniuSecretKey string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if uploadStrategy != "" {
		r.uploadStrategy = uploadStrategy
	}
	if localDir != "" {
		r.uploadLocalDir = localDir
	}
	if urlPrefix != "" {
		r.uploadURLPrefix = urlPrefix
	}
	if qiniuBucket != "" {
		r.qiniuBucket = qiniuBucket
	}
	if qiniuDomain != "" {
		r.qiniuDomain = qiniuDomain
	}
	if qiniuAccessKey != "" {
		r.qiniuAccessKey = qiniuAccessKey
	}
	if qiniuSecretKey != "" {
		r.qiniuSecretKey = qiniuSecretKey
	}
}

type tUser struct {
	ID                uint64     `gorm:"column:id;primaryKey"`
	Username          string     `gorm:"column:username"`
	PasswordHash      string     `gorm:"column:password_hash"`
	Email             string     `gorm:"column:email"`
	Nickname          string     `gorm:"column:nickname"`
	Avatar            string     `gorm:"column:avatar"`
	Status            string     `gorm:"column:status;default:'ENABLED'"`
	LastLoginTime     *time.Time `gorm:"column:last_login_time"`
	LastLoginIp       string     `gorm:"column:last_login_ip"`
	LoginFailureCount uint       `gorm:"column:login_failure_count"`
	LockedUntil       *time.Time `gorm:"column:locked_until"`
	CreateTime        time.Time  `gorm:"column:create_time"`
	UpdateTime        time.Time  `gorm:"column:update_time"`
}

func (tUser) TableName() string { return "t_user" }

type tRole struct {
	ID          uint64    `gorm:"column:id;primaryKey"`
	Name        string    `gorm:"column:name"`
	Description string    `gorm:"column:description"`
	Status      int8      `gorm:"column:status"`
	SortOrder   int       `gorm:"column:sort_order"`
	CreateTime  time.Time `gorm:"column:create_time;autoCreateTime"`
	UpdateTime  time.Time `gorm:"column:update_time;autoUpdateTime"`
}

func (tRole) TableName() string { return "t_role" }

type tUserRole struct {
	ID     uint64 `gorm:"column:id;primaryKey"`
	UserID uint64 `gorm:"column:user_id"`
	RoleID uint64 `gorm:"column:role_id"`
}

func (tUserRole) TableName() string { return "t_user_role" }

type tPermission struct {
	ID          uint64    `gorm:"column:id;primaryKey"`
	PermKey     string    `gorm:"column:perm_key"`
	Name        string    `gorm:"column:name"`
	Module      string    `gorm:"column:module"`
	Type        string    `gorm:"column:type"`
	Description string    `gorm:"column:description"`
	FrontPath   string    `gorm:"column:front_path"`
	Icon        string    `gorm:"column:icon"`
	SortOrder   int       `gorm:"column:sort_order"`
	Status      int8      `gorm:"column:status"`
	CreateTime  time.Time `gorm:"column:create_time;autoCreateTime"`
	UpdateTime  time.Time `gorm:"column:update_time;autoUpdateTime"`
}

func (tPermission) TableName() string { return "t_permission" }

type tRolePermission struct {
	ID         uint64    `gorm:"column:id;primaryKey"`
	RoleID     uint64    `gorm:"column:role_id"`
	PermKey    string    `gorm:"column:perm_key"`
	CreateTime time.Time `gorm:"column:create_time;autoCreateTime"`
}

func (tRolePermission) TableName() string { return "t_role_permission" }

type tTag struct {
	ID         uint64     `gorm:"column:id;primaryKey"`
	Name       string     `gorm:"column:name"`
	CreateTime time.Time  `gorm:"column:create_time;autoCreateTime"`
	UpdateTime time.Time  `gorm:"column:update_time;autoUpdateTime"`
	DeletedAt  *time.Time `gorm:"column:deleted_at"`
}

func (tTag) TableName() string { return "t_tag" }

type tCategory struct {
	ID         uint64     `gorm:"column:id;primaryKey"`
	Name       string     `gorm:"column:name"`
	CreateTime time.Time  `gorm:"column:create_time;autoCreateTime"`
	UpdateTime time.Time  `gorm:"column:update_time;autoUpdateTime"`
	DeletedAt  *time.Time `gorm:"column:deleted_at"`
}

func (tCategory) TableName() string { return "t_category" }

type tArticle struct {
	ID          uint64     `gorm:"column:id;primaryKey"`
	Title       string     `gorm:"column:title"`
	Slug        string     `gorm:"column:slug"`
	TitleImage  string     `gorm:"column:title_image"`
	Description string     `gorm:"column:description"`
	ReadNum     int        `gorm:"column:read_num"`
	ShareNum    int        `gorm:"column:share_num"`
	LikeNum     int        `gorm:"column:like_num;default:0"`
	IsTop       bool       `gorm:"column:is_top"`
	IsPublished bool       `gorm:"column:is_published"`
	PublishTime *time.Time `gorm:"column:publish_time"`
	CreateTime  time.Time  `gorm:"column:create_time;autoCreateTime"`
	UpdateTime  time.Time  `gorm:"column:update_time;autoUpdateTime"`
}

func (tArticle) TableName() string { return "t_article" }

type tArticleContent struct {
	ID        uint64 `gorm:"column:id;primaryKey"`
	ArticleID uint64 `gorm:"column:article_id"`
	Content   string `gorm:"column:content"`
}

func (tArticleContent) TableName() string { return "t_article_content" }

type tArticleCategoryRel struct {
	ID         uint64 `gorm:"column:id;primaryKey"`
	ArticleID  uint64 `gorm:"column:article_id"`
	CategoryID uint64 `gorm:"column:category_id"`
}

func (tArticleCategoryRel) TableName() string { return "t_article_category_rel" }

type tArticleTagRel struct {
	ID        uint64 `gorm:"column:id;primaryKey"`
	ArticleID uint64 `gorm:"column:article_id"`
	TagID     uint64 `gorm:"column:tag_id"`
}

func (tArticleTagRel) TableName() string { return "t_article_tag_rel" }

type tPV struct {
	ArticleID uint64    `gorm:"column:article_id"`
	PVDate    time.Time `gorm:"column:pv_date"`
	PVCount   int64     `gorm:"column:pv_count"`
}

func (tPV) TableName() string { return "t_statistics_article_pv" }

type tMPUser struct {
	ID             uint64    `gorm:"column:id;primaryKey"`
	OpenID         string    `gorm:"column:openid"`
	Nickname       string    `gorm:"column:nickname"`
	Avatar         string    `gorm:"column:avatar"`
	Status         uint8     `gorm:"column:status"`
	FirstLoginTime time.Time `gorm:"column:first_login_time"`
	LastLoginTime  time.Time `gorm:"column:last_login_time"`
	CreateTime     time.Time `gorm:"column:create_time"`
	UpdateTime     time.Time `gorm:"column:update_time"`
}

func (tMPUser) TableName() string { return "t_mp_user" }

type tMPUserBehavior struct {
	ID          uint64    `gorm:"column:id;primaryKey"`
	OpenID      string    `gorm:"column:openid"`
	ArticleID   uint64    `gorm:"column:article_id"`
	EventType   string    `gorm:"column:event_type"`
	StaySeconds int       `gorm:"column:stay_seconds"`
	Scene       string    `gorm:"column:scene"`
	Strategy    string    `gorm:"column:strategy"`
	EventTime   time.Time `gorm:"column:event_time"`
}

func (tMPUserBehavior) TableName() string { return "t_mp_user_behavior" }

type tMPUserInterest struct {
	ID            uint64    `gorm:"column:id;primaryKey"`
	OpenID        string    `gorm:"column:openid"`
	DimensionType string    `gorm:"column:dimension_type"`
	DimensionID   uint64    `gorm:"column:dimension_id"`
	Score         float64   `gorm:"column:score"`
	UpdateTime    time.Time `gorm:"column:update_time"`
}

func (tMPUserInterest) TableName() string { return "t_mp_user_interest" }

type tMPFavorite struct {
	ID         uint64    `gorm:"column:id;primaryKey"`
	OpenID     string    `gorm:"column:openid"`
	ArticleID  uint64    `gorm:"column:article_id"`
	CreateTime time.Time `gorm:"column:create_time"`
	UpdateTime time.Time `gorm:"column:update_time"`
}

func (tMPFavorite) TableName() string { return "t_mp_user_favorite" }

type tMPUserTag struct {
	ID          uint64    `gorm:"column:id;primaryKey"`
	OpenID      string    `gorm:"column:openid"`
	TagName     string    `gorm:"column:tag_name"`
	TagCategory string    `gorm:"column:tag_category"`
	Score       float64   `gorm:"column:score"`
	Source      string    `gorm:"column:source"`
	CreateTime  time.Time `gorm:"column:create_time"`
	UpdateTime  time.Time `gorm:"column:update_time"`
}

func (tMPUserTag) TableName() string { return "t_mp_user_tag" }

type tMPUserProfile struct {
	ID         uint64    `gorm:"column:id;primaryKey"`
	OpenID     string    `gorm:"column:openid"`
	Dimension  string    `gorm:"column:dimension"`
	Score      float64   `gorm:"column:score"`
	CreateTime time.Time `gorm:"column:create_time"`
	UpdateTime time.Time `gorm:"column:update_time"`
}

func (tMPUserProfile) TableName() string { return "t_mp_user_profile" }

type TMPUserSubscribe struct {
	ID         uint64    `gorm:"column:id;primaryKey"`
	OpenID     string    `gorm:"column:openid;size:128;uniqueIndex"`
	Subscribe  bool      `gorm:"column:subscribe"`
	CreateTime time.Time `gorm:"column:create_time;autoCreateTime"`
	UpdateTime time.Time `gorm:"column:update_time;autoUpdateTime"`
}

func (TMPUserSubscribe) TableName() string { return "t_mp_user_subscribe" }

type tTopic struct {
	ID          uint64    `gorm:"column:id;primaryKey"`
	Name        string    `gorm:"column:name"`
	Description string    `gorm:"column:description"`
	CoverImage  string    `gorm:"column:cover_image"`
	Status      uint8     `gorm:"column:status"`
	CreateTime  time.Time `gorm:"column:create_time;autoCreateTime"`
	UpdateTime  time.Time `gorm:"column:update_time;autoUpdateTime"`
}

func (tTopic) TableName() string { return "t_topic" }

type tArticleTopicRel struct {
	ID         uint64    `gorm:"column:id;primaryKey"`
	ArticleID  uint64    `gorm:"column:article_id"`
	TopicID    uint64    `gorm:"column:topic_id"`
	SortOrder  int       `gorm:"column:sort_order"`
	CreateTime time.Time `gorm:"column:create_time;autoCreateTime"`
	UpdateTime time.Time `gorm:"column:update_time;autoUpdateTime"`
}

func (tArticleTopicRel) TableName() string { return "t_article_topic_rel" }

type tSearchHistory struct {
	ID         uint64    `gorm:"column:id;primaryKey"`
	Keyword    string    `gorm:"column:keyword"`
	SearchTime time.Time `gorm:"column:search_time"`
}

func (tSearchHistory) TableName() string { return "t_search_history" }

type TAccessLog struct {
	ID             uint64    `gorm:"column:id;primaryKey"`
	TraceID        string    `gorm:"column:trace_id"`
	RequestMethod  string    `gorm:"column:request_method"`
	RequestURL     string    `gorm:"column:request_url"`
	RequestPath    string    `gorm:"column:request_path"`
	RequestQuery   string    `gorm:"column:request_query"`
	RequestParams  string    `gorm:"column:request_params"`
	RequestBody    string    `gorm:"column:request_body"`
	VisitorUserID  uint64    `gorm:"column:visitor_user_id"`
	VisitorName    string    `gorm:"column:visitor_name"`
	IPAddress      []byte    `gorm:"column:ip_address"`
	RequestTime    time.Time `gorm:"column:request_time"`
	ResponseTime   int       `gorm:"column:response_time"`
	ResponseStatus int       `gorm:"column:response_status"`
	ResponseBody   string    `gorm:"column:response_body"`
	UserAgent      string    `gorm:"column:user_agent"`
	RequestSource  string    `gorm:"column:request_source"`
	IsError        bool      `gorm:"column:is_error"`
	ErrorID        *uint64   `gorm:"column:error_id"`
	CreateTime     time.Time `gorm:"column:create_time"`
}

func (TAccessLog) TableName() string { return "t_access_log" }

type TErrorLog struct {
	ID            uint64    `gorm:"column:id;primaryKey"`
	AccessLogID   uint64    `gorm:"column:access_log_id"`
	TraceID       string    `gorm:"column:trace_id"`
	ErrorCode     string    `gorm:"column:error_code"`
	ErrorMessage  string    `gorm:"column:error_message"`
	ErrorDetail   string    `gorm:"column:error_detail"`
	StackTrace    string    `gorm:"column:stack_trace"`
	RequestURL    string    `gorm:"column:request_url"`
	RequestMethod string    `gorm:"column:request_method"`
	RequestParams string    `gorm:"column:request_params"`
	RequestBody   string    `gorm:"column:request_body"`
	ResponseBody  string    `gorm:"column:response_body"`
	IPAddress     []byte    `gorm:"column:ip_address"`
	UserAgent     string    `gorm:"column:user_agent"`
	CreateTime    time.Time `gorm:"column:create_time"`
}

func (TErrorLog) TableName() string { return "t_error_log" }

type tLink struct {
	ID          uint64    `gorm:"column:id;primaryKey"`
	Name        string    `gorm:"column:name"`
	Url         string    `gorm:"column:url"`
	Description string    `gorm:"column:description"`
	Icon        string    `gorm:"column:icon"`
	SortOrder   int       `gorm:"column:sort_order"`
	IsActive    bool      `gorm:"column:is_active"`
	CreateTime  time.Time `gorm:"column:create_time;autoCreateTime"`
	UpdateTime  time.Time `gorm:"column:update_time;autoUpdateTime"`
}

func (tLink) TableName() string { return "t_link" }
