package handler

import (
	"fmt"
	"net/http"
	"time"

	"sanmoo-server-go/internal/infrastructure/search"
	"sanmoo-server-go/internal/interfaces/http/dto"
	"sanmoo-server-go/internal/interfaces/http/middleware"
	apperr "sanmoo-server-go/internal/shared/errors"
	"sanmoo-server-go/internal/shared/response"

	"github.com/gin-gonic/gin"
)

// ======================== Admin Settings Handlers ========================

func (h *Handler) AdminGetSettings(c *gin.Context) {
	out, err := h.svc.Setting.GetSettings(c.Request.Context())
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, out)
}

func (h *Handler) AdminUpdateSettings(c *gin.Context) {
	body := map[string]any{}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, apperr.ErrInvalidParam)
		return
	}
	op, _ := c.Get(middleware.CtxUsernameKey)
	operator, _ := op.(string)
	if operator == "" {
		operator = "system"
	}
	if err := h.svc.Setting.UpdateSettings(c.Request.Context(), body, operator); err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, dto.EmptyResponse{})
}

func (h *Handler) AdminSendEmailVerificationCode(c *gin.Context) {
	var req dto.SendEmailVerificationCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, apperr.ErrInvalidParam)
		return
	}
	identifier, err := h.svc.Setting.SendEmailVerificationCode(c.Request.Context(), req.EmailConfig)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, dto.EmailVerificationResponse{Identifier: identifier})
}

func (h *Handler) AdminVerifyEmailVerificationCode(c *gin.Context) {
	var req dto.VerifyEmailVerificationCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, apperr.ErrInvalidParam)
		return
	}
	if err := h.svc.Setting.VerifyEmailVerificationCode(c.Request.Context(), req.Email, req.Code); err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, dto.EmptyResponse{})
}

func (h *Handler) AdminExportSettings(c *gin.Context) {
	out, err := h.svc.Setting.ExportSettings(c.Request.Context())
	if err != nil {
		response.Fail(c, err)
		return
	}
	c.Header("Content-Disposition", "attachment; filename=settings.json")
	c.Header("Content-Type", "application/json")
	c.JSON(http.StatusOK, out)
}

func (h *Handler) AdminImportSettings(c *gin.Context) {
	var body map[string]any
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, apperr.ErrInvalidParam)
		return
	}
	op, _ := c.Get(middleware.CtxUsernameKey)
	operator, _ := op.(string)
	if operator == "" {
		operator = "system"
	}
	if err := h.svc.Setting.ImportSettings(c.Request.Context(), body, operator); err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, dto.EmptyResponse{})
}

func (h *Handler) AdminSyncMeiliSearch(c *gin.Context) {
	st, err := h.svc.Setting.GetSettings(c.Request.Context())
	if err != nil {
		response.Fail(c, err)
		return
	}

	host := ""
	apiKey := ""
	index := "articles"

	if st != nil && st.UIConfig != nil {
		if h, ok := st.UIConfig["meilisearchHost"].(string); ok {
			host = h
		}
		if key, ok := st.UIConfig["meilisearchApiKey"].(string); ok {
			apiKey = key
		}
		if idx, ok := st.UIConfig["meilisearchIndex"].(string); ok {
			index = idx
		}
	}

	if host == "" {
		response.Fail(c, apperr.New("400", "请先在博客设置中配置 MeiliSearch 地址后再同步"))
		return
	}

	articles, err := h.svc.Article.ListAllPublishedArticles(c.Request.Context())
	if err != nil {
		response.Fail(c, err)
		return
	}

	searchClient := search.NewMeiliSearchClient(host, apiKey, index)
	if err := searchClient.CreateIndexIfNotExists(c.Request.Context()); err != nil {
		response.Fail(c, err)
		return
	}

	type doc struct {
		ID          uint64 `json:"id"`
		Title       string `json:"title"`
		Description string `json:"description"`
		Content     string `json:"content"`
	}

	docs := make([]interface{}, 0, len(articles))
	for _, article := range articles {
		docs = append(docs, doc{
			ID:          article.ID,
			Title:       article.Title,
			Description: article.Description,
			Content:     article.Content,
		})
	}

	if err := searchClient.AddDocuments(c.Request.Context(), docs); err != nil {
		response.Fail(c, err)
		return
	}

	if st != nil {
		_ = h.svc.Setting.SaveMeiliSearchSyncTime(c.Request.Context(), time.Now().Format("2006-01-02 15:04:05"))
	}

	response.Ok(c, map[string]any{
		"count": len(articles),
		"msg":   "同步完成",
	})
}

func (h *Handler) AdminGetMeiliSearchStats(c *gin.Context) {
	st, err := h.svc.Setting.GetSettings(c.Request.Context())
	if err != nil {
		response.Fail(c, err)
		return
	}

	host := ""
	apiKey := ""
	index := "articles"
	lastSyncTime := "-"

	if st != nil && st.UIConfig != nil {
		if h, ok := st.UIConfig["meilisearchHost"].(string); ok {
			host = h
		}
		if key, ok := st.UIConfig["meilisearchApiKey"].(string); ok {
			apiKey = key
		}
		if idx, ok := st.UIConfig["meilisearchIndex"].(string); ok {
			index = idx
		}
		if tVal, ok := st.UIConfig["meilisearchLastSyncTime"]; ok {
			switch v := tVal.(type) {
			case string:
				lastSyncTime = v
			case time.Time:
				lastSyncTime = v.Format("2006-01-02 15:04:05")
			case []uint8:
				lastSyncTime = string(v)
			default:
				lastSyncTime = fmt.Sprintf("%v", v)
			}
		}
	}

	if host == "" {
		response.Ok(c, map[string]any{
			"articleCount": 0,
			"indexStatus":  "unknown",
			"lastSyncTime": lastSyncTime,
		})
		return
	}

	searchClient := search.NewMeiliSearchClient(host, apiKey, index)
	count, err := searchClient.GetIndexStats(c.Request.Context())
	if err != nil {
		response.Ok(c, map[string]any{
			"articleCount": 0,
			"indexStatus":  "error",
			"lastSyncTime": lastSyncTime,
		})
		return
	}

	response.Ok(c, map[string]any{
		"articleCount": count,
		"indexStatus":  "healthy",
		"lastSyncTime": lastSyncTime,
	})
}

func keys(m map[string]any) []string {
	k := make([]string, 0, len(m))
	for key := range m {
		k = append(k, key)
	}
	return k
}

func (h *Handler) AdminGetCoreConfig(c *gin.Context) {
	out, err := h.svc.Setting.GetCoreConfig(c.Request.Context())
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, out)
}

func (h *Handler) AdminUpdateCoreConfig(c *gin.Context) {
	body := map[string]any{}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, apperr.ErrInvalidParam)
		return
	}
	op, _ := c.Get(middleware.CtxUsernameKey)
	operator, _ := op.(string)
	if operator == "" {
		operator = "system"
	}
	if err := h.svc.Setting.UpdateCoreConfig(c.Request.Context(), body, operator); err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, dto.EmptyResponse{})
}

func (h *Handler) AdminGetSocialConfig(c *gin.Context) {
	out, err := h.svc.Setting.GetSocialConfig(c.Request.Context())
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, out)
}

func (h *Handler) AdminUpdateSocialConfig(c *gin.Context) {
	body := map[string]any{}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, apperr.ErrInvalidParam)
		return
	}
	op, _ := c.Get(middleware.CtxUsernameKey)
	operator, _ := op.(string)
	if operator == "" {
		operator = "system"
	}
	if err := h.svc.Setting.UpdateSocialConfig(c.Request.Context(), body, operator); err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, dto.EmptyResponse{})
}

func (h *Handler) AdminGetSearchConfig(c *gin.Context) {
	out, err := h.svc.Setting.GetSearchConfig(c.Request.Context())
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, out)
}

func (h *Handler) AdminUpdateSearchConfig(c *gin.Context) {
	body := map[string]any{}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, apperr.ErrInvalidParam)
		return
	}
	op, _ := c.Get(middleware.CtxUsernameKey)
	operator, _ := op.(string)
	if operator == "" {
		operator = "system"
	}
	if err := h.svc.Setting.UpdateSearchConfig(c.Request.Context(), body, operator); err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, dto.EmptyResponse{})
}

func (h *Handler) AdminGetPrivacyConfig(c *gin.Context) {
	out, err := h.svc.Setting.GetPrivacyConfig(c.Request.Context())
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, out)
}

func (h *Handler) AdminUpdatePrivacyConfig(c *gin.Context) {
	body := map[string]any{}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, apperr.ErrInvalidParam)
		return
	}
	op, _ := c.Get(middleware.CtxUsernameKey)
	operator, _ := op.(string)
	if operator == "" {
		operator = "system"
	}
	if err := h.svc.Setting.UpdatePrivacyConfig(c.Request.Context(), body, operator); err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, dto.EmptyResponse{})
}

func (h *Handler) AdminGetStorageConfig(c *gin.Context) {
	out, err := h.svc.Setting.GetStorageConfig(c.Request.Context())
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, out)
}

func (h *Handler) AdminUpdateStorageConfig(c *gin.Context) {
	body := map[string]any{}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, apperr.ErrInvalidParam)
		return
	}
	op, _ := c.Get(middleware.CtxUsernameKey)
	operator, _ := op.(string)
	if operator == "" {
		operator = "system"
	}
	if err := h.svc.Setting.UpdateStorageConfig(c.Request.Context(), body, operator); err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, dto.EmptyResponse{})
}

func (h *Handler) AdminGetEmailConfig(c *gin.Context) {
	out, err := h.svc.Setting.GetEmailConfig(c.Request.Context())
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, out)
}

func (h *Handler) AdminUpdateEmailConfig(c *gin.Context) {
	body := map[string]any{}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, apperr.ErrInvalidParam)
		return
	}
	op, _ := c.Get(middleware.CtxUsernameKey)
	operator, _ := op.(string)
	if operator == "" {
		operator = "system"
	}
	if err := h.svc.Setting.UpdateEmailConfig(c.Request.Context(), body, operator); err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, dto.EmptyResponse{})
}
