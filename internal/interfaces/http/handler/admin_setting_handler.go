package handler

import (
	"net/http"

	"sanmoo-server-go/internal/interfaces/http/dto"
	"sanmoo-server-go/internal/interfaces/http/middleware"
	"sanmoo-server-go/internal/infrastructure/search"
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
	if err := h.svc.Setting.SendEmailVerificationCode(c.Request.Context(), req.EmailConfig); err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, dto.EmptyResponse{})
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

	response.Ok(c, map[string]any{
		"count": len(articles),
		"msg":   "同步完成",
	})
}
