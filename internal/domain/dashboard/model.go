package dashboard

type Statistics struct {
	ArticleCount            int64 `json:"articleCount"`
	PublishedArticleCount   int64 `json:"publishedArticleCount"`
	UnpublishedArticleCount int64 `json:"unpublishedArticleCount"`
	CategoryCount           int64 `json:"categoryCount"`
	TagCount                int64 `json:"tagCount"`
	UserCount               int64 `json:"userCount"`
	VisitorCount            int64 `json:"visitorCount"`
	TodayPv                 int64 `json:"todayPv"`
	TopicCount              int64 `json:"topicCount"`
	MpUserCount             int64 `json:"mpUserCount"`
	TotalReads              int64 `json:"totalReads"`
	YesterdayPv             int64 `json:"yesterdayPv"`
	TodayMpUserCount        int64 `json:"todayMpUserCount"`
	YesterdayMpUserCount    int64 `json:"yesterdayMpUserCount"`
}

type VisitorRecord struct {
	ID             uint64 `json:"id"`
	TraceID        string `json:"traceId"`
	RequestMethod  string `json:"requestMethod"`
	RequestURL     string `json:"requestUrl"`
	VisitorUserID  uint64 `json:"visitorUserId"`
	VisitorName    string `json:"visitorName"`
	IPAddress      string `json:"ipAddress"`
	VisitTime      string `json:"visitTime"`
	ResponseTime   int    `json:"responseTime"`
	ResponseStatus int    `json:"responseStatus"`
	RequestSource  string `json:"requestSource"`
	IsError        bool   `json:"isError"`
	UserAgent      string `json:"userAgent"`
	RequestParams  string `json:"requestParams"`
	RequestBody    string `json:"requestBody"`
	ResponseBody   string `json:"responseBody"`
}

type ErrorLogRecord struct {
	ID            uint64 `json:"id"`
	AccessLogID   uint64 `json:"accessLogId"`
	TraceID       string `json:"traceId"`
	ErrorCode     string `json:"errorCode"`
	ErrorMessage  string `json:"errorMessage"`
	ErrorDetail   string `json:"errorDetail"`
	StackTrace    string `json:"stackTrace"`
	RequestURL    string `json:"requestUrl"`
	RequestMethod string `json:"requestMethod"`
	RequestParams string `json:"requestParams"`
	RequestBody   string `json:"requestBody"`
	ResponseBody  string `json:"responseBody"`
	IPAddress     string `json:"ipAddress"`
	UserAgent     string `json:"userAgent"`
	CreateTime    string `json:"createTime"`
}

// NameValue 用于饼图等统计输出结构。
type NameValue struct {
	Name  string `json:"name"`
	Value int64  `json:"value"`
}

// DateCount 用于热力图等按日期统计结构。
type DateCount struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

// ArticleReadStat 文章阅读量统计
type ArticleReadStat struct {
	ID          uint64 `json:"id"`
	Title       string `json:"title"`
	ReadNum     int    `json:"readNum"`
	CategoryID  uint64 `json:"categoryId"`
	Category    string `json:"category"`
	CreateTime  string `json:"createTime"`
}

// CategoryReadStat 分类阅读量统计
type CategoryReadStat struct {
	ID          uint64 `json:"id"`
	Name        string `json:"name"`
	ArticleCount int64 `json:"articleCount"`
	TotalReads  int64 `json:"totalReads"`
}

// TagReadStat 标签阅读量统计
type TagReadStat struct {
	ID          uint64 `json:"id"`
	Name        string `json:"name"`
	ArticleCount int64 `json:"articleCount"`
	TotalReads  int64 `json:"totalReads"`
}

// ContentTrend 内容热度趋势
type ContentTrend struct {
	Date        string `json:"date"`
	TotalPV     int64  `json:"totalPv"`
	ArticlePV   int64  `json:"articlePv"`
	NewArticles int64  `json:"newArticles"`
}
