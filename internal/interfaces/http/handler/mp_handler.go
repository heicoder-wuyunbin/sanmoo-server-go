package handler

import (
	"strconv"

	"sanmoo-server-go/internal/interfaces/http/dto"
	apperr "sanmoo-server-go/internal/shared/errors"
	"sanmoo-server-go/internal/shared/markdown"
	"sanmoo-server-go/internal/shared/response"

	"github.com/gin-gonic/gin"
)

// ======================== Mini-Program Handlers ========================

// MpSettings 小程序设置接口：返回轻量字段，避免把后台审计字段全部下发。
func (h *Handler) MpSettings(c *gin.Context) {
	out, err := h.svc.Setting.GetSettings(c.Request.Context())
	if err != nil {
		response.Fail(c, err)
		return
	}
	core := out.CoreConfig
	ui := out.UIConfig
	storage := out.StorageConfig
	var resp dto.MPSettingsResponse
	resp.CoreConfig.BlogName = core["blogName"]
	resp.CoreConfig.Author = core["author"]
	resp.CoreConfig.Introduction = core["introduction"]
	resp.CoreConfig.Avatar = core["avatar"]
	resp.CoreConfig.Poster = core["poster"]
	resp.UIConfig.GithubHome = ui["githubHome"]
	resp.UIConfig.CsdnHome = ui["csdnHome"]
	resp.UIConfig.GiteeHome = ui["giteeHome"]
	resp.UIConfig.ZhihuHome = ui["zhihuHome"]
	resp.UIConfig.GithubShow = ui["githubShow"]
	resp.UIConfig.CsdnShow = ui["csdnShow"]
	resp.UIConfig.GiteeShow = ui["giteeShow"]
	resp.UIConfig.ZhihuShow = ui["zhihuShow"]
	resp.UIConfig.RecommendStrategy = ui["recommendStrategy"]
	resp.StorageConfig.UploadStrategy = storage["uploadStrategy"]
	resp.StorageConfig.UploadLocalUrlPrefix = storage["uploadLocalUrlPrefix"]
	response.Ok(c, resp)
}

func (h *Handler) MpCompliance(c *gin.Context) {
	out, err := h.svc.Setting.GetPublicCompliance(c.Request.Context())
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, out)
}

func (h *Handler) MpRecommendations(c *gin.Context) {
	articleID := parseUintDefault(c.Query("articleId"), 0)
	size := parseIntDefault(c.Query("size"), 10)
	out, err := h.svc.Article.RecommendArticlesForMP(c.Request.Context(), articleID, size)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, out)
}

func (h *Handler) MpAuthSession(c *gin.Context) {
	var req dto.MPAuthSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, apperr.ErrInvalidParam)
		return
	}
	out, err := h.svc.Auth.MPAuthSession(c.Request.Context(), req.Code)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, out)
}

func (h *Handler) MpUserProfile(c *gin.Context) {
	out, err := h.svc.MPUser.MPGetUserProfile(c.Request.Context(), getOpenID(c))
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, out)
}

func (h *Handler) MpUpdateUserProfile(c *gin.Context) {
	var req dto.MPUserProfileUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, apperr.ErrInvalidParam)
		return
	}
	openID := req.OpenID
	if openID == "" {
		openID = c.GetHeader("X-MP-OPENID")
	}
	if err := h.svc.MPUser.MPUpdateUserProfile(c.Request.Context(), openID, req.NickName, req.AvatarUrl); err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, dto.EmptyResponse{})
}

func (h *Handler) MpAddFavorite(c *gin.Context) {
	articleID, err := strconv.ParseUint(c.Param("articleId"), 10, 64)
	if err != nil {
		response.Fail(c, apperr.ErrInvalidParam)
		return
	}
	if err := h.svc.MPUser.MPAddFavorite(c.Request.Context(), getOpenID(c), articleID); err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, dto.EmptyResponse{})
}

func (h *Handler) MpRemoveFavorite(c *gin.Context) {
	articleID, err := strconv.ParseUint(c.Param("articleId"), 10, 64)
	if err != nil {
		response.Fail(c, apperr.ErrInvalidParam)
		return
	}
	if err := h.svc.MPUser.MPRemoveFavorite(c.Request.Context(), getOpenID(c), articleID); err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, dto.EmptyResponse{})
}

func (h *Handler) MpFavoriteStatus(c *gin.Context) {
	articleID, err := strconv.ParseUint(c.Param("articleId"), 10, 64)
	if err != nil {
		response.Fail(c, apperr.ErrInvalidParam)
		return
	}
	out, err := h.svc.MPUser.MPFavoriteStatus(c.Request.Context(), getOpenID(c), articleID)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, out)
}

func (h *Handler) MpFavorites(c *gin.Context) {
	page := parseIntDefault(c.Query("page"), 1)
	size := parseIntDefault(c.Query("size"), 20)
	out, err := h.svc.MPUser.MPFavoriteList(c.Request.Context(), getOpenID(c), page, size)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, out)
}

func (h *Handler) MpAddBrowseHistory(c *gin.Context) {
	articleID, err := strconv.ParseUint(c.Param("articleId"), 10, 64)
	if err != nil {
		response.Fail(c, apperr.ErrInvalidParam)
		return
	}
	err = h.svc.MPUser.AddMPBrowseHistory(c.Request.Context(), getOpenID(c), articleID)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, dto.EmptyResponse{})
}

func (h *Handler) MpClearBrowseHistory(c *gin.Context) {
	err := h.svc.MPUser.ClearMPBrowseHistory(c.Request.Context(), getOpenID(c))
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, dto.EmptyResponse{})
}

func (h *Handler) MpBrowseHistory(c *gin.Context) {
	page := parseIntDefault(c.Query("page"), 1)
	size := parseIntDefault(c.Query("size"), 20)
	out, err := h.svc.MPUser.MPBrowseHistoryList(c.Request.Context(), getOpenID(c), page, size)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, out)
}

func (h *Handler) MpCategories(c *gin.Context) {
	out, err := h.svc.Category.ListCategories(c.Request.Context())
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, out)
}

func (h *Handler) MpTags(c *gin.Context) {
	out, err := h.svc.Tag.ListTags(c.Request.Context(), 0, 0, "")
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, out)
}

// MpTopics 小程序专题列表
func (h *Handler) MpTopics(c *gin.Context) {
	page := parseIntDefault(c.Query("page"), 1)
	size := parseIntDefault(c.Query("size"), 20)
	out, err := h.svc.Topic.ListTopics(c.Request.Context(), page, size)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, out)
}

// MpTopicDetail 小程序专题详情
func (h *Handler) MpTopicDetail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		response.Fail(c, apperr.ErrInvalidParam)
		return
	}
	out, err := h.svc.Topic.GetTopicDetail(c.Request.Context(), id)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, out)
}

// MpTopicArticles 小程序专题文章列表
func (h *Handler) MpTopicArticles(c *gin.Context) {
	topicID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || topicID == 0 {
		response.Fail(c, apperr.ErrInvalidParam)
		return
	}
	page := parseIntDefault(c.Query("page"), 1)
	size := parseIntDefault(c.Query("size"), 20)
	out, err := h.svc.Topic.ListTopicArticles(c.Request.Context(), topicID, page, size)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, out)
}

// MpArticles 小程序文章列表：默认 size=20，并裁剪为移动端常用字段。
func (h *Handler) MpArticles(c *gin.Context) {
	page := parseIntDefault(c.Query("page"), 1)
	size := parseIntDefault(c.Query("size"), 20)
	keyword := c.Query("keyword")
	categoryID := parseUintDefault(c.Query("categoryId"), 0)
	tagID := parseUintDefault(c.Query("tagId"), 0)
	one := 1
	out, err := h.svc.Article.ListArticles(c.Request.Context(), page, size, keyword, categoryID, tagID, &one)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, out)
}

func (h *Handler) MpArticleDetail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, apperr.ErrInvalidParam)
		return
	}
	out, err := h.svc.Article.GetArticleDetail(c.Request.Context(), id, true)
	if err != nil {
		response.Fail(c, err)
		return
	}

	if c.Query("share") == "1" {
		go func() {
			_ = h.svc.Article.IncrementShareCount(c.Request.Context(), id)
		}()
	}
	a := out.Article
	prev := out.PrevArticle
	next := out.NextArticle
	if a == nil {
		response.Fail(c, apperr.ErrNotFound)
		return
	}
	resp := dto.MPArticleDetailResponse{
		ID:           a.ID,
		Title:        a.Title,
		TitleImage:   a.TitleImage,
		Description:  a.Description,
		CreateTime:   a.CreateTime,
		UpdateTime:   a.UpdateTime,
		ReadNum:      a.ReadNum,
		CategoryID:   a.CategoryID,
		CategoryName: a.Category,
		Tags:         a.Tags,
		PrevArticle:  prev,
		NextArticle:  next,
		IsFavorited:  false,
	}

	openID := getOpenID(c)
	if openID != "" {
		if status, favErr := h.svc.MPUser.MPFavoriteStatus(c.Request.Context(), openID, id); favErr == nil {
			resp.IsFavorited = status.IsFavorited
		}
	}
	// 后端统一将 Markdown 转为结构化 HTML，小程序端直接渲染。
	if html, mErr := markdown.ToHTML(a.Content); mErr == nil {
		resp.ContentHtml = html
	}
	response.Ok(c, resp)
}

func (h *Handler) MpArticlesByCategory(c *gin.Context) {
	categoryID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, apperr.ErrInvalidParam)
		return
	}
	page := parseIntDefault(c.Query("page"), 1)
	size := parseIntDefault(c.Query("size"), 20)
	one := 1
	out, err := h.svc.Article.ListArticles(c.Request.Context(), page, size, "", categoryID, 0, &one)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, out)
}

func (h *Handler) MpArticlesByTag(c *gin.Context) {
	tagID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, apperr.ErrInvalidParam)
		return
	}
	page := parseIntDefault(c.Query("page"), 1)
	size := parseIntDefault(c.Query("size"), 20)
	one := 1
	out, err := h.svc.Article.ListArticles(c.Request.Context(), page, size, "", 0, tagID, &one)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, out)
}

func (h *Handler) MpArchives(c *gin.Context) {
	out, err := h.svc.Article.Archives(c.Request.Context())
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, out)
}

func (h *Handler) MpPrivacyPolicy(c *gin.Context) {
	compliance, err := h.svc.Setting.GetPublicCompliance(c.Request.Context())
	if err != nil {
		response.Fail(c, err)
		return
	}
	out, _ := compliance["privacyPolicy"].(string)
	contentHtml := out
	if out != "" {
		if html, mErr := markdown.ToHTML(out); mErr == nil {
			contentHtml = html
		}
	}
	response.Ok(c, map[string]any{
		"content":              contentHtml,
		"dataRetentionPolicy":  compliance["dataRetentionPolicy"],
		"accountDeletionGuide": compliance["accountDeletionGuide"],
	})
}

func (h *Handler) MpDeleteUser(c *gin.Context) {
	openID := getOpenID(c)
	if openID == "" {
		response.Fail(c, apperr.ErrInvalidParam)
		return
	}
	if err := h.svc.MPUser.MPDeleteUser(c.Request.Context(), openID); err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, dto.EmptyResponse{})
}

func (h *Handler) MpSubscribe(c *gin.Context) {
	openID := getOpenID(c)
	if openID == "" {
		response.Fail(c, apperr.ErrInvalidParam)
		return
	}
	var req struct {
		Subscribe bool `json:"subscribe"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, apperr.ErrInvalidParam)
		return
	}
	if err := h.svc.MPUser.MPSetSubscribe(c.Request.Context(), openID, req.Subscribe); err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, map[string]any{"subscribe": req.Subscribe})
}

func (h *Handler) MpSubscribeStatus(c *gin.Context) {
	openID := getOpenID(c)
	if openID == "" {
		response.Fail(c, apperr.ErrInvalidParam)
		return
	}
	status, err := h.svc.MPUser.MPGetSubscribe(c.Request.Context(), openID)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, map[string]any{"subscribe": status})
}
