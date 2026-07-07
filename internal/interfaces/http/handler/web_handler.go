package handler

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	domarticle "sanmoo-server-go/internal/domain/article"
	"sanmoo-server-go/internal/infrastructure/search"
	"sanmoo-server-go/internal/interfaces/http/dto"
	apperr "sanmoo-server-go/internal/shared/errors"
	"sanmoo-server-go/internal/shared/markdown"
	"sanmoo-server-go/internal/shared/response"

	"github.com/gin-gonic/gin"
)

// ======================== Web (Public Portal) Handlers ========================

func (h *Handler) WebSettings(c *gin.Context) { h.AdminGetSettings(c) }

func (h *Handler) WebCategories(c *gin.Context) { h.GetCategories(c) }

func (h *Handler) WebTags(c *gin.Context) {
	out, err := h.svc.Tag.ListTags(c.Request.Context(), 0, 0, "")
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, out)
}

func (h *Handler) WebArticles(c *gin.Context) {
	var q dto.ArticleListQuery
	_ = c.ShouldBindQuery(&q)
	one := 1
	out, err := h.svc.Article.ListArticles(c.Request.Context(), q.Page, q.Size, q.Keyword, q.CategoryID, q.TagID, &one)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, out)
}

func (h *Handler) WebArticleDetail(c *gin.Context) {
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
	if out.Article != nil {
		if html, toc, mErr := markdown.ToHTMLWithTOC(out.Article.Content); mErr == nil {
			out.ContentHtml = html
			tocItems := make([]dto.TOCItem, 0, len(toc))
			for _, item := range toc {
				tocItems = append(tocItems, dto.TOCItem{
					Level: item.Level,
					Text:  item.Text,
					ID:    item.ID,
				})
			}
			out.TOC = tocItems
		}
	}
	response.Ok(c, out)
}

// WebArticleBySlug 根据 slug 获取文章详情（SEO 友好 URL）
func (h *Handler) WebArticleBySlug(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		response.Fail(c, apperr.ErrInvalidParam)
		return
	}

	// 通过 slug 获取文章
	article, err := h.svc.Article.GetArticleBySlug(c.Request.Context(), slug)
	if err != nil {
		response.Fail(c, err)
		return
	}

	// 如果有 slug，则从 id 访问应重定向到 slug URL（SEO 权重传递）
	out := &dto.ArticleDetailResponse{Article: article}
	if article != nil {
		if html, toc, mErr := markdown.ToHTMLWithTOC(article.Content); mErr == nil {
			out.ContentHtml = html
			tocItems := make([]dto.TOCItem, 0, len(toc))
			for _, item := range toc {
				tocItems = append(tocItems, dto.TOCItem{
					Level: item.Level,
					Text:  item.Text,
					ID:    item.ID,
				})
			}
			out.TOC = tocItems
		}
	}
	response.Ok(c, out)
}

func (h *Handler) WebArticlesByCategory(c *gin.Context) {
	categoryID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, apperr.ErrInvalidParam)
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	one := 1
	out, err := h.svc.Article.ListArticles(c.Request.Context(), page, size, "", categoryID, 0, &one)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, out)
}

func (h *Handler) WebArticlesByTag(c *gin.Context) {
	tagID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, apperr.ErrInvalidParam)
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	one := 1
	out, err := h.svc.Article.ListArticles(c.Request.Context(), page, size, "", 0, tagID, &one)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, out)
}

func (h *Handler) WebArchives(c *gin.Context) {
	out, err := h.svc.Article.Archives(c.Request.Context())
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, out)
}

func (h *Handler) WebHotSearches(c *gin.Context) {
	ctx := c.Request.Context()
	st, err := h.svc.Setting.GetSettings(ctx)
	if err != nil {
		st = nil
	}

	useRealHot := false
	if st != nil && st.UIConfig != nil {
		if modeBool, ok := st.UIConfig["hotSearchMode"].(bool); ok {
			useRealHot = modeBool
		} else if modeStr, ok := st.UIConfig["hotSearchMode"].(string); ok {
			useRealHot = modeStr == "REAL" || modeStr == "true" || modeStr == "TRUE"
		} else if modeInt, ok := st.UIConfig["hotSearchMode"].(int); ok {
			useRealHot = modeInt == 1
		} else if modeInt64, ok := st.UIConfig["hotSearchMode"].(int64); ok {
			useRealHot = modeInt64 == 1
		} else if modeUint, ok := st.UIConfig["hotSearchMode"].(uint); ok {
			useRealHot = modeUint == 1
		} else if modeUint8, ok := st.UIConfig["hotSearchMode"].(uint8); ok {
			useRealHot = modeUint8 == 1
		}
	}

	var hotSearches []string
	if useRealHot {
		hotSearches, err = h.svc.Article.GetRealHotSearches(ctx, 10)
		if err != nil || len(hotSearches) == 0 {
			hotSearches, _ = h.svc.Setting.GetHotSearches(ctx)
		}
	} else {
		hotSearches, _ = h.svc.Setting.GetHotSearches(ctx)
	}

	response.Ok(c, hotSearches)
}

func (h *Handler) WebHotArticles(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "6"))
	if limit < 1 || limit > 20 {
		limit = 6
	}

	isPublished := 1
	pagedList, err := h.svc.Article.ListArticles(c.Request.Context(), 1, 50, "", 0, 0, &isPublished)
	if err != nil {
		response.Fail(c, err)
		return
	}

	items, ok := pagedList.List.([]domarticle.Article)
	if !ok {
		response.Fail(c, apperr.ErrInternal)
		return
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].ReadNum > items[j].ReadNum
	})

	if len(items) > limit {
		items = items[:limit]
	}

	response.Ok(c, items)
}

// ======================== Web Topics ========================

// WebTopics PC 端专题列表
func (h *Handler) WebTopics(c *gin.Context) {
	page := parseIntDefault(c.Query("page"), 1)
	size := parseIntDefault(c.Query("size"), 12)
	out, err := h.svc.Topic.ListTopics(c.Request.Context(), page, size)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, out)
}

// WebTopicDetail PC 端专题详情
func (h *Handler) WebTopicDetail(c *gin.Context) {
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

// WebTopicArticles PC 端专题文章列表
func (h *Handler) WebTopicArticles(c *gin.Context) {
	topicID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || topicID == 0 {
		response.Fail(c, apperr.ErrInvalidParam)
		return
	}
	page := parseIntDefault(c.Query("page"), 1)
	size := parseIntDefault(c.Query("size"), 10)
	out, err := h.svc.Topic.ListTopicArticles(c.Request.Context(), topicID, page, size)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, out)
}

func (h *Handler) Sitemap(c *gin.Context) {
	ctx := c.Request.Context()

	st, err := h.svc.Setting.GetSettings(ctx)
	if err != nil {
		st = nil
	}

	baseURL := "https://backendart.com"
	if st != nil && st.CoreConfig != nil {
		if url, ok := st.CoreConfig["siteUrl"].(string); ok && url != "" {
			baseURL = url
		}
	}

	now := time.Now()

	var xmlBuilder strings.Builder
	xmlBuilder.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
	xmlBuilder.WriteString(`<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9" xmlns:image="http://www.google.com/schemas/sitemap-image/1.1">`)

	xmlBuilder.WriteString(`<url>`)
	xmlBuilder.WriteString(`<loc>` + baseURL + `</loc>`)
	xmlBuilder.WriteString(`<lastmod>` + now.Format("2006-01-02") + `</lastmod>`)
	xmlBuilder.WriteString(`<changefreq>daily</changefreq>`)
	xmlBuilder.WriteString(`<priority>1.0</priority>`)
	xmlBuilder.WriteString(`</url>`)

	xmlBuilder.WriteString(`<url>`)
	xmlBuilder.WriteString(`<loc>` + baseURL + `/archives</loc>`)
	xmlBuilder.WriteString(`<lastmod>` + now.Format("2006-01-02") + `</lastmod>`)
	xmlBuilder.WriteString(`<changefreq>weekly</changefreq>`)
	xmlBuilder.WriteString(`<priority>0.6</priority>`)
	xmlBuilder.WriteString(`</url>`)

	xmlBuilder.WriteString(`<url>`)
	xmlBuilder.WriteString(`<loc>` + baseURL + `/categories</loc>`)
	xmlBuilder.WriteString(`<lastmod>` + now.Format("2006-01-02") + `</lastmod>`)
	xmlBuilder.WriteString(`<changefreq>weekly</changefreq>`)
	xmlBuilder.WriteString(`<priority>0.6</priority>`)
	xmlBuilder.WriteString(`</url>`)

	xmlBuilder.WriteString(`<url>`)
	xmlBuilder.WriteString(`<loc>` + baseURL + `/tags</loc>`)
	xmlBuilder.WriteString(`<lastmod>` + now.Format("2006-01-02") + `</lastmod>`)
	xmlBuilder.WriteString(`<changefreq>weekly</changefreq>`)
	xmlBuilder.WriteString(`<priority>0.6</priority>`)
	xmlBuilder.WriteString(`</url>`)

	xmlBuilder.WriteString(`<url>`)
	xmlBuilder.WriteString(`<loc>` + baseURL + `/topics</loc>`)
	xmlBuilder.WriteString(`<lastmod>` + now.Format("2006-01-02") + `</lastmod>`)
	xmlBuilder.WriteString(`<changefreq>weekly</changefreq>`)
	xmlBuilder.WriteString(`<priority>0.6</priority>`)
	xmlBuilder.WriteString(`</url>`)

	categories, err := h.svc.Category.ListCategories(ctx)
	if err == nil && categories != nil {
		for _, cat := range categories.List {
			xmlBuilder.WriteString(`<url>`)
			xmlBuilder.WriteString(`<loc>` + baseURL + `/categories/` + fmt.Sprintf("%d", cat.ID) + `</loc>`)
			xmlBuilder.WriteString(`<changefreq>weekly</changefreq>`)
			xmlBuilder.WriteString(`<priority>0.5</priority>`)
			xmlBuilder.WriteString(`</url>`)
		}
	}

	tags, err := h.svc.Tag.ListAllTags(ctx)
	if err == nil && tags != nil {
		for _, t := range tags.List {
			xmlBuilder.WriteString(`<url>`)
			xmlBuilder.WriteString(`<loc>` + baseURL + `/tags/` + fmt.Sprintf("%d", t.ID) + `</loc>`)
			xmlBuilder.WriteString(`<changefreq>weekly</changefreq>`)
			xmlBuilder.WriteString(`<priority>0.5</priority>`)
			xmlBuilder.WriteString(`</url>`)
		}
	}

	topics, err := h.svc.Topic.ListAllTopicsWithCount(ctx)
	if err == nil && topics != nil {
		for _, t := range topics.List {
			xmlBuilder.WriteString(`<url>`)
			xmlBuilder.WriteString(`<loc>` + baseURL + `/topics/` + fmt.Sprintf("%d", t.ID) + `</loc>`)
			if t.CreateTime != nil {
				switch ct := t.CreateTime.(type) {
				case string:
					if len(ct) >= 10 {
						xmlBuilder.WriteString(`<lastmod>` + ct[:10] + `</lastmod>`)
					}
				case []uint8:
					s := string(ct)
					if len(s) >= 10 {
						xmlBuilder.WriteString(`<lastmod>` + s[:10] + `</lastmod>`)
					}
				}
			}
			xmlBuilder.WriteString(`<changefreq>monthly</changefreq>`)
			xmlBuilder.WriteString(`<priority>0.5</priority>`)
			xmlBuilder.WriteString(`</url>`)
		}
	}

	articles, err := h.svc.Article.ListAllPublishedArticlesForSitemap(ctx)
	if err != nil {
		response.Fail(c, err)
		return
	}

	for _, article := range articles {
		xmlBuilder.WriteString(`<url>`)
		xmlBuilder.WriteString(`<loc>` + baseURL + `/article/` + fmt.Sprintf("%d", article.ID) + `</loc>`)
		if !article.UpdateTime.IsZero() {
			xmlBuilder.WriteString(`<lastmod>` + article.UpdateTime.Format("2006-01-02") + `</lastmod>`)
		}

		var changefreq string
		daysSinceUpdate := int(now.Sub(article.UpdateTime).Hours() / 24)
		if daysSinceUpdate <= 7 {
			changefreq = "daily"
		} else if daysSinceUpdate <= 30 {
			changefreq = "weekly"
		} else if daysSinceUpdate <= 90 {
			changefreq = "monthly"
		} else {
			changefreq = "yearly"
		}
		xmlBuilder.WriteString(`<changefreq>` + changefreq + `</changefreq>`)

		var priority string
		if article.IsTop {
			priority = "1.0"
		} else if article.ReadNum >= 1000 {
			priority = "0.9"
		} else if article.ReadNum >= 500 {
			priority = "0.85"
		} else if daysSinceUpdate <= 30 {
			priority = "0.85"
		} else {
			priority = "0.8"
		}
		xmlBuilder.WriteString(`<priority>` + priority + `</priority>`)

		if article.TitleImage != "" {
			imageURL := article.TitleImage
			if !strings.HasPrefix(imageURL, "http://") && !strings.HasPrefix(imageURL, "https://") {
				if strings.HasPrefix(imageURL, "/") {
					imageURL = baseURL + imageURL
				} else {
					imageURL = baseURL + "/" + imageURL
				}
			}
			xmlBuilder.WriteString(`<image:image>`)
			xmlBuilder.WriteString(`<image:loc>` + imageURL + `</image:loc>`)
			xmlBuilder.WriteString(`<image:title>` + escapeXML(article.Title) + `</image:title>`)
			xmlBuilder.WriteString(`</image:image>`)
		}

		xmlBuilder.WriteString(`</url>`)
	}

	xmlBuilder.WriteString(`</urlset>`)

	c.Header("Content-Type", "application/xml; charset=utf-8")
	c.String(http.StatusOK, xmlBuilder.String())
}

func (h *Handler) WebRelatedArticles(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, apperr.ErrInvalidParam)
		return
	}
	size := 6
	if s, err := strconv.Atoi(c.DefaultQuery("size", "6")); err == nil && s > 0 {
		size = s
	}
	out, err := h.svc.Article.RecommendRelatedArticles(c.Request.Context(), id, size)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, out)
}

func (h *Handler) WebLikeArticle(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, apperr.ErrInvalidParam)
		return
	}
	likeNum, err := h.svc.Article.LikeArticle(c.Request.Context(), id)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, map[string]int{"likeNum": likeNum})
}

func (h *Handler) WebRandomArticle(c *gin.Context) {
	excludeID, _ := strconv.ParseUint(c.Query("exclude"), 10, 64)
	article, err := h.svc.Article.RandomArticle(c.Request.Context(), excludeID)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, article)
}

func (h *Handler) RSS(c *gin.Context) {
	ctx := c.Request.Context()

	st, err := h.svc.Setting.GetSettings(ctx)
	if err != nil {
		st = nil
	}

	enabled := true
	if st != nil && st.CoreConfig != nil {
		if val, ok := st.CoreConfig["rssEnabled"].(bool); ok {
			enabled = val
		} else if val, ok := st.CoreConfig["rssEnabled"].(int); ok {
			enabled = val != 0
		} else if val, ok := st.CoreConfig["rssEnabled"].(float64); ok {
			enabled = val != 0
		}
	}

	if !enabled {
		c.Status(http.StatusNotFound)
		return
	}

	articles, err := h.svc.Article.ListAllPublishedArticlesForSitemap(ctx)
	if err != nil {
		response.Fail(c, err)
		return
	}

	baseURL := "https://backendart.com"
	siteTitle := "Sanmoo Blog"
	siteDescription := "Sanmoo 的个人博客"
	if st != nil && st.CoreConfig != nil {
		if url, ok := st.CoreConfig["siteUrl"].(string); ok && url != "" {
			baseURL = url
		}
		if title, ok := st.CoreConfig["siteName"].(string); ok && title != "" {
			siteTitle = title
		}
		if desc, ok := st.CoreConfig["siteDescription"].(string); ok && desc != "" {
			siteDescription = desc
		}
	}

	var xmlBuilder strings.Builder
	xmlBuilder.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
	xmlBuilder.WriteString(`<rss version="2.0" xmlns:atom="http://www.w3.org/2005/Atom">`)
	xmlBuilder.WriteString(`<channel>`)
	xmlBuilder.WriteString(`<title>` + siteTitle + `</title>`)
	xmlBuilder.WriteString(`<link>` + baseURL + `</link>`)
	xmlBuilder.WriteString(`<description>` + siteDescription + `</description>`)
	xmlBuilder.WriteString(`<atom:link href="` + baseURL + `/rss.xml" rel="self" type="application/rss+xml"/>`)

	for _, article := range articles {
		xmlBuilder.WriteString(`<item>`)
		xmlBuilder.WriteString(`<title>` + escapeXML(article.Title) + `</title>`)
		xmlBuilder.WriteString(`<link>` + baseURL + `/article/` + fmt.Sprintf("%d", article.ID) + `</link>`)
		if article.Description != "" {
			xmlBuilder.WriteString(`<description>` + escapeXML(article.Description) + `</description>`)
		}
		if !article.CreateTime.IsZero() {
			xmlBuilder.WriteString(`<pubDate>` + article.CreateTime.Format("Mon, 02 Jan 2006 15:04:05 -0700") + `</pubDate>`)
		}
		xmlBuilder.WriteString(`<guid>` + baseURL + `/article/` + fmt.Sprintf("%d", article.ID) + `</guid>`)
		xmlBuilder.WriteString(`</item>`)
	}

	xmlBuilder.WriteString(`</channel></rss>`)

	c.Header("Content-Type", "application/rss+xml; charset=utf-8")
	c.String(http.StatusOK, xmlBuilder.String())
}

func escapeXML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	return s
}

func (h *Handler) WebSearch(c *gin.Context) {
	var q dto.ArticleListQuery
	_ = c.ShouldBindQuery(&q)

	st, err := h.svc.Setting.GetSettings(c.Request.Context())
	if err != nil {
		st = nil
	}

	useMeiliSearch := false
	if st != nil && st.UIConfig != nil {
		if modeBool, ok := st.UIConfig["hotSearchMode"].(bool); ok {
			useMeiliSearch = modeBool
		} else if modeStr, ok := st.UIConfig["hotSearchMode"].(string); ok {
			useMeiliSearch = modeStr == "REAL"
		} else if modeNum, ok := st.UIConfig["hotSearchMode"].(float64); ok {
			useMeiliSearch = modeNum != 0
		}
	}

	if q.Keyword != "" {
		_ = h.svc.Article.RecordSearchHistory(c.Request.Context(), q.Keyword)
	}

	if useMeiliSearch && q.Keyword != "" {
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

		if host != "" {
			searchClient := search.NewMeiliSearchClient(host, apiKey, index)
			results, err := searchClient.Search(c.Request.Context(), q.Keyword, q.Size)
			if err == nil && len(results) > 0 {
				ids := make([]uint64, 0, len(results))
				for _, r := range results {
					ids = append(ids, uint64(r.ID))
				}
				list, err := h.svc.Article.ListArticlesByIDs(c.Request.Context(), ids)
				if err == nil && len(list) > 0 {
					response.Ok(c, map[string]any{
						"page":  1,
						"size":  len(list),
						"total": len(list),
						"list":  list,
					})
					return
				}
			}
		}
	}

	one := 1
	out, err := h.svc.Article.ListArticles(c.Request.Context(), q.Page, q.Size, q.Keyword, q.CategoryID, q.TagID, &one)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, out)
}
