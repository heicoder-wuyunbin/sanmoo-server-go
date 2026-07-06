package dto

// 通用请求 DTO
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	// Code 邮箱验证码（当开启后台登录邮箱二次验证时必填）
	Code string `json:"code,omitempty"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken"`
}

// 验证码相关请求 DTO
type SendVerificationCodeRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type VerifyVerificationCodeRequest struct {
	UserID uint64 `json:"userId"`
	Code   string `json:"code"`
}

type SendTestEmailRequest struct {
	To string `json:"to"`
}

type EmailConfigRequest struct {
	Host            string `json:"host"`
	Port            string `json:"port"`
	Username        string `json:"username"`
	Password        string `json:"password"`
	From            string `json:"from"`
	LoginMfaEnabled bool   `json:"loginMfaEnabled"`
}

type SendEmailVerificationCodeRequest struct {
	EmailConfig EmailConfigRequest `json:"emailConfig"`
}

type VerifyEmailVerificationCodeRequest struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

type AdminUserCreateRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	RoleID   uint64 `json:"roleId"`
	Email    string `json:"email"`
}

type AdminUserUpdateRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	RoleID   uint64 `json:"roleId"`
	Email    string `json:"email"`
}

type AdminUserUpdatePasswordRequest struct {
	Password string `json:"password"`
}

type UpdatePasswordRequest struct {
	OldPassword string `json:"oldPassword"`
	NewPassword string `json:"newPassword"`
}

type AdminTagCreateRequest struct {
	Name string `json:"name"`
}
type AdminTagUpdateRequest struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
}

type AdminCategoryCreateRequest struct {
	Name string `json:"name"`
}
type AdminCategoryUpdateRequest struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
}

type AdminLinkCreateRequest struct {
	Name        string `json:"name"`
	Url         string `json:"url"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	SortOrder   int    `json:"sortOrder"`
}

type AdminLinkUpdateRequest struct {
	Name        string `json:"name"`
	Url         string `json:"url"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	SortOrder   int    `json:"sortOrder"`
	IsActive    bool   `json:"isActive"`
}

type AdminTopicCreateRequest struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	CoverImage  string   `json:"coverImage"`
	ArticleIDs  []uint64 `json:"articleIds"`
}
type AdminTopicUpdateRequest struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	CoverImage  string   `json:"coverImage"`
	ArticleIDs  []uint64 `json:"articleIds"`
}

type AdminArticleCreateRequest struct {
	Title       string   `json:"title"`
	TitleImage  string   `json:"titleImage"`
	Description string   `json:"description"`
	Content     string   `json:"content"`
	CategoryID  uint64   `json:"categoryId"`
	TagIDs      []uint64 `json:"tagIds"`
	TopicIDs    []uint64 `json:"topicIds"`
	IsTop       int      `json:"isTop"`
	IsPublished int      `json:"isPublished"`
}

type AdminArticleUpdateRequest = AdminArticleCreateRequest

type BatchDeleteRequest struct {
	IDs []uint64 `json:"ids"`
}

type ImportErrorLogsRequest struct {
	Logs []ImportErrorLogItem `json:"logs"`
}

type ImportErrorLogItem struct {
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
}

type ArticleListQuery struct {
	Page        int    `form:"page"`
	Size        int    `form:"size"`
	Keyword     string `form:"keyword"`
	CategoryID  uint64 `form:"categoryId"`
	TagID       uint64 `form:"tagId"`
	IsPublished *int   `form:"isPublished"`
}

type PageQuery struct {
	Page    int    `form:"page"`
	Size    int    `form:"size"`
	Keyword string `form:"keyword"`
}

type MPAuthSessionRequest struct {
	Code   string `json:"code"`
	OpenID string `json:"openid"`
}

type MPBehaviorRequest struct {
	OpenID      string `json:"openid"`
	ArticleID   uint64 `json:"articleId"`
	EventType   string `json:"eventType"`
	StaySeconds int    `json:"staySeconds"`
	Scene       string `json:"scene"`
	Strategy    string `json:"strategy"`
}

type MPUserProfileUpdateRequest struct {
	OpenID    string `json:"openid"`
	NickName  string `json:"nickName"`
	AvatarUrl string `json:"avatarUrl"`
}

// MP 用户管理相关请求 DTO
type MPUserListQuery struct {
	Page    int    `form:"page"`
	Size    int    `form:"size"`
	Keyword string `form:"keyword"`
	TagName string `form:"tagName"`
}

// 权限管理相关请求 DTO
type AdminPermissionCreateRequest struct {
	PermKey     string `json:"permKey"`
	Name        string `json:"name"`
	Module      string `json:"module"`
	Type        string `json:"type"`
	Description string `json:"description"`
	SortOrder   int    `json:"sortOrder"`
}

type AdminPermissionUpdateRequest struct {
	Name        string `json:"name"`
	Module      string `json:"module"`
	Type        string `json:"type"`
	Description string `json:"description"`
	SortOrder   int    `json:"sortOrder"`
	Status      int    `json:"status"`
}

// 角色管理相关请求 DTO
type AdminRoleCreateRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	SortOrder   int    `json:"sortOrder"`
}

type AdminRoleUpdateRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	SortOrder   int    `json:"sortOrder"`
	Status      int    `json:"status"`
}

type AssignRolePermissionsRequest struct {
	PermKeys []string `json:"permKeys"`
}

type AssignUserRolesRequest struct {
	RoleIDs []uint64 `json:"roleIds"`
}
