package handler

import (
	"encoding/csv"
	"fmt"
	"strconv"
	"strings"
	"time"

	domarticle "sanmoo-server-go/internal/domain/article"
	"sanmoo-server-go/internal/interfaces/http/dto"
	apperr "sanmoo-server-go/internal/shared/errors"
	"sanmoo-server-go/internal/shared/response"

	"github.com/gin-gonic/gin"
)

// ======================== Admin Article Management Handlers ========================

func toArticle(req dto.AdminArticleCreateRequest) domarticle.Article {
	tags := make([]domarticle.TagRef, 0, len(req.TagIDs))
	for _, id := range req.TagIDs {
		tags = append(tags, domarticle.TagRef{ID: id})
	}
	topics := make([]domarticle.TopicRef, 0, len(req.TopicIDs))
	for _, id := range req.TopicIDs {
		topics = append(topics, domarticle.TopicRef{ID: id})
	}
	return domarticle.Article{
		Title:       req.Title,
		TitleImage:  req.TitleImage,
		Description: req.Description,
		Content:     req.Content,
		CategoryID:  req.CategoryID,
		Tags:        tags,
		Topics:      topics,
		IsTop:       req.IsTop == 1,
		IsPublished: req.IsPublished == 1,
	}
}

func (h *Handler) GetArticles(c *gin.Context) {
	var q dto.ArticleListQuery
	_ = c.ShouldBindQuery(&q)
	out, err := h.svc.Article.ListArticlesNoCache(c.Request.Context(), q.Page, q.Size, q.Keyword, q.CategoryID, q.TagID, q.IsPublished)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, out)
}

func (h *Handler) CreateArticle(c *gin.Context) {
	var req dto.AdminArticleCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, apperr.ErrInvalidParam)
		return
	}
	id, err := h.svc.Article.CreateArticle(c.Request.Context(), toArticle(req))
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, dto.IDResponse{ID: id})
}

func (h *Handler) UpdateArticle(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, apperr.ErrInvalidParam)
		return
	}
	var req dto.AdminArticleUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, apperr.ErrInvalidParam)
		return
	}

	// 获取现有文章，仅更新请求中明确传入的字段
	existing, err := h.svc.Article.GetArticleDetail(c.Request.Context(), id, false)
	if err != nil {
		response.Fail(c, err)
		return
	}
	merged := mergeArticleUpdate(existing.Article, req)

	if err := h.svc.Article.UpdateArticle(c.Request.Context(), id, merged); err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, dto.EmptyResponse{})
}

// mergeArticleUpdate 将请求中非 nil 的字段合并到现有文章上
func mergeArticleUpdate(existing *domarticle.Article, req dto.AdminArticleUpdateRequest) domarticle.Article {
	merged := *existing
	if req.Title != nil {
		merged.Title = *req.Title
	}
	if req.TitleImage != nil {
		merged.TitleImage = *req.TitleImage
	}
	if req.Description != nil {
		merged.Description = *req.Description
	}
	if req.Content != nil {
		merged.Content = *req.Content
	}
	if req.CategoryID != nil {
		merged.CategoryID = *req.CategoryID
	}
	if req.TagIDs != nil {
		merged.Tags = make([]domarticle.TagRef, 0, len(req.TagIDs))
		for _, tid := range req.TagIDs {
			merged.Tags = append(merged.Tags, domarticle.TagRef{ID: tid})
		}
	}
	if req.TopicIDs != nil {
		merged.Topics = make([]domarticle.TopicRef, 0, len(req.TopicIDs))
		for _, tid := range req.TopicIDs {
			merged.Topics = append(merged.Topics, domarticle.TopicRef{ID: tid})
		}
	}
	if req.IsTop != nil {
		merged.IsTop = *req.IsTop == 1
	}
	if req.IsPublished != nil {
		merged.IsPublished = *req.IsPublished == 1
	}
	return merged
}

func (h *Handler) UpdateArticleStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, apperr.ErrInvalidParam)
		return
	}
	var req struct {
		IsPublished *bool `json:"isPublished"`
		IsTop       *bool `json:"isTop"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, apperr.ErrInvalidParam)
		return
	}
	if err := h.svc.Article.UpdateArticleStatus(c.Request.Context(), id, req.IsPublished, req.IsTop); err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, dto.EmptyResponse{})
}

func (h *Handler) BatchUpdateArticleStatus(c *gin.Context) {
	var req struct {
		IDs         []uint64 `json:"ids" binding:"required"`
		IsPublished *bool    `json:"isPublished"`
		IsTop       *bool    `json:"isTop"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, apperr.ErrInvalidParam)
		return
	}
	if err := h.svc.Article.BatchUpdateArticleStatus(c.Request.Context(), req.IDs, req.IsPublished, req.IsTop); err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, dto.EmptyResponse{})
}

func (h *Handler) DeleteArticle(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, apperr.ErrInvalidParam)
		return
	}
	if err := h.svc.Article.DeleteArticle(c.Request.Context(), id); err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, dto.EmptyResponse{})
}

func (h *Handler) BatchDeleteArticles(c *gin.Context) {
	var req dto.BatchDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, apperr.ErrInvalidParam)
		return
	}
	if err := h.svc.Article.BatchDeleteArticles(c.Request.Context(), req.IDs); err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, dto.EmptyResponse{})
}

func (h *Handler) AdminArticleDetail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, apperr.ErrInvalidParam)
		return
	}
	out, err := h.svc.Article.GetArticleDetail(c.Request.Context(), id, false)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, out)
}

func (h *Handler) ExportArticles(c *gin.Context) {
	var q dto.ArticleListQuery
	_ = c.ShouldBindQuery(&q)
	// 导出时不分页，取大量数据
	out, err := h.svc.Article.ListArticlesNoCache(c.Request.Context(), 1, 10000, q.Keyword, q.CategoryID, q.TagID, q.IsPublished)
	if err != nil {
		response.Fail(c, err)
		return
	}

	list, ok := out.List.([]domarticle.Article)
	if !ok {
		response.Fail(c, apperr.ErrInternal)
		return
	}

	filename := fmt.Sprintf("articles_%s.csv", time.Now().Format("20060102"))
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))

	// 写入 UTF-8 BOM，确保 Excel 打开中文不乱码
	c.Writer.Write([]byte("\xEF\xBB\xBF"))

	writer := csv.NewWriter(c.Writer)
	defer writer.Flush()

	// 表头
	header := []string{"ID", "标题", "Slug", "分类", "标签", "阅读量", "点赞量", "置顶", "发布状态", "创建时间", "更新时间"}
	if err := writer.Write(header); err != nil {
		return
	}

	for _, item := range list {
		tagNames := make([]string, 0, len(item.Tags))
		for _, t := range item.Tags {
			tagNames = append(tagNames, t.Name)
		}
		top := "否"
		if item.IsTop {
			top = "是"
		}
		published := "未发布"
		if item.IsPublished {
			published = "已发布"
		}
		row := []string{
			fmt.Sprintf("%d", item.ID),
			item.Title,
			item.Slug,
			item.Category,
			strings.Join(tagNames, ","),
			fmt.Sprintf("%d", item.ReadNum),
			fmt.Sprintf("%d", item.LikeNum),
			top,
			published,
			item.CreateTime.Format("2006-01-02 15:04:05"),
			item.UpdateTime.Format("2006-01-02 15:04:05"),
		}
		writer.Write(row)
	}
}

// RefreshArticleSlug 刷新单篇文章的 slug
func (h *Handler) RefreshArticleSlug(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, apperr.ErrInvalidParam)
		return
	}
	if err := h.svc.Article.RefreshArticleSlug(c.Request.Context(), id); err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, dto.EmptyResponse{})
}

// BatchRefreshArticleSlugs 批量刷新所有文章的 slug
func (h *Handler) BatchRefreshArticleSlugs(c *gin.Context) {
	count, err := h.svc.Article.BatchRefreshArticleSlugs(c.Request.Context())
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, map[string]any{"count": count})
}

// ImportArticles 批量导入文章（逐个导入，标题去重）
func (h *Handler) ImportArticles(c *gin.Context) {
	var req dto.BatchImportArticlesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, apperr.ErrInvalidParam)
		return
	}

	results := make([]dto.BatchImportResult, 0, len(req.Articles))
	successCount := 0
	failedCount := 0
	skippedCount := 0

	for i, item := range req.Articles {
		art := domarticle.Article{
			Title:       item.Title,
			Description: item.Description,
			Content:     item.Content,
			CategoryID:  item.CategoryID,
			Tags:        make([]domarticle.TagRef, 0, len(item.TagIDs)),
			Topics:      make([]domarticle.TopicRef, 0, len(item.TopicIDs)),
			IsTop:       item.IsTop == 1,
			IsPublished: item.IsPublished == 1,
		}
		for _, id := range item.TagIDs {
			art.Tags = append(art.Tags, domarticle.TagRef{ID: id})
		}
		for _, id := range item.TopicIDs {
			art.Topics = append(art.Topics, domarticle.TopicRef{ID: id})
		}

		id, err := h.svc.Article.CreateArticle(c.Request.Context(), art)
		if err != nil {
			errMsg := err.Error()
			status := "failed"
			// 标题已存在视为"跳过"，而非"失败"
			if strings.Contains(errMsg, "文章标题已存在") {
				status = "skipped"
				skippedCount++
			} else {
				failedCount++
			}
			results = append(results, dto.BatchImportResult{
				Index:   i,
				Title:   item.Title,
				Success: false,
				Status:  status,
				Error:   errMsg,
			})
		} else {
			results = append(results, dto.BatchImportResult{
				Index:   i,
				Title:   item.Title,
				Success: true,
				Status:  "success",
				ID:      id,
			})
			successCount++
		}
	}

	response.Ok(c, dto.BatchImportResponse{
		Total:    len(req.Articles),
		Success:  successCount,
		Failed:   failedCount,
		Skipped:  skippedCount,
		Results:  results,
	})
}

// ExportArticlesSelected 导出指定文章（或全部）
func (h *Handler) ExportArticlesSelected(c *gin.Context) {
	var req dto.BatchExportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// 如果请求体为空，导出全部
		req.IDs = nil
	}

	var list []domarticle.Article
	if len(req.IDs) > 0 {
		// 导出指定文章
		for _, id := range req.IDs {
			out, err := h.svc.Article.GetArticleDetail(c.Request.Context(), id, false)
			if err != nil {
				continue
			}
			if out.Article != nil {
				list = append(list, *out.Article)
			}
		}
	} else {
		// 导出全部
		out, err := h.svc.Article.ListArticlesNoCache(c.Request.Context(), 1, 10000, "", 0, 0, nil)
		if err != nil {
			response.Fail(c, err)
			return
		}
		if l, ok := out.List.([]domarticle.Article); ok {
			list = l
		}
	}

	filename := fmt.Sprintf("articles_%s.csv", time.Now().Format("20060102"))
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))

	c.Writer.Write([]byte("\xEF\xBB\xBF"))

	writer := csv.NewWriter(c.Writer)
	defer writer.Flush()

	header := []string{"ID", "标题", "Slug", "分类", "标签", "阅读量", "点赞量", "置顶", "发布状态", "创建时间", "更新时间"}
	if err := writer.Write(header); err != nil {
		return
	}

	for _, item := range list {
		tagNames := make([]string, 0, len(item.Tags))
		for _, t := range item.Tags {
			tagNames = append(tagNames, t.Name)
		}
		top := "否"
		if item.IsTop {
			top = "是"
		}
		published := "未发布"
		if item.IsPublished {
			published = "已发布"
		}
		row := []string{
			fmt.Sprintf("%d", item.ID),
			item.Title,
			item.Slug,
			item.Category,
			strings.Join(tagNames, ","),
			fmt.Sprintf("%d", item.ReadNum),
			fmt.Sprintf("%d", item.LikeNum),
			top,
			published,
			item.CreateTime.Format("2006-01-02 15:04:05"),
			item.UpdateTime.Format("2006-01-02 15:04:05"),
		}
		writer.Write(row)
	}
}
