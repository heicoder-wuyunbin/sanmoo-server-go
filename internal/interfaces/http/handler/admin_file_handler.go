package handler

import (
	"encoding/base64"
	"strings"

	"sanmoo-server-go/internal/interfaces/http/dto"
	apperr "sanmoo-server-go/internal/shared/errors"
	"sanmoo-server-go/internal/shared/response"

	"github.com/gin-gonic/gin"
)

// ======================== Admin File Management Handlers ========================

func (h *Handler) GetFiles(c *gin.Context) {
	var q dto.PageQuery
	_ = c.ShouldBindQuery(&q)
	out, err := h.svc.File.ListFiles(c.Request.Context(), q.Page, q.Size, q.Keyword)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, out)
}

func (h *Handler) UploadFile(c *gin.Context) {
	file, err := c.FormFile("file")
	if err == nil {
		out, upErr := h.svc.File.UploadFile(c.Request.Context(), file)
		if upErr != nil {
			response.Fail(c, upErr)
			return
		}
		response.Ok(c, out)
		return
	}

	var jsonReq struct {
		File     string `json:"file"`
		Filename string `json:"filename"`
		Name     string `json:"name"`
	}
	if bindErr := c.ShouldBindJSON(&jsonReq); bindErr != nil {
		if file2, err2 := c.FormFile("upload"); err2 == nil {
			out, upErr := h.svc.File.UploadFile(c.Request.Context(), file2)
			if upErr != nil {
				response.Fail(c, upErr)
				return
			}
			response.Ok(c, out)
			return
		} else {
			response.Fail(c, apperr.ErrInvalidParam)
			return
		}
	}

	if strings.TrimSpace(jsonReq.File) == "" {
		response.Fail(c, apperr.ErrInvalidParam)
		return
	}
	raw := strings.TrimSpace(jsonReq.File)
	if idx := strings.Index(raw, ","); idx > 0 && strings.Contains(raw[:idx], "base64") {
		raw = raw[idx+1:]
	}
	data, decodeErr := base64.StdEncoding.DecodeString(raw)
	if decodeErr != nil {
		data, decodeErr = base64.RawStdEncoding.DecodeString(raw)
		if decodeErr != nil {
			response.Fail(c, apperr.ErrInvalidParam)
			return
		}
	}
	filename := strings.TrimSpace(jsonReq.Filename)
	if filename == "" {
		filename = strings.TrimSpace(jsonReq.Name)
	}
	if filename == "" {
		filename = "upload.bin"
	}
	out, upErr := h.svc.File.UploadFileBytes(c.Request.Context(), filename, data)
	if upErr != nil {
		response.Fail(c, upErr)
		return
	}
	response.Ok(c, out)
}

func (h *Handler) DeleteFile(c *gin.Context) {
	if err := h.svc.File.DeleteFile(c.Request.Context(), c.Param("id")); err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, dto.EmptyResponse{})
}

// ProxyImage 代理图片请求，生成临时 URL 并 302 重定向，避免 URL 过期问题
func (h *Handler) ProxyImage(c *gin.Context) {
	filePath := c.Param("path")
	if filePath == "" {
		response.Fail(c, apperr.ErrInvalidParam)
		return
	}
	// 去掉开头的斜杠
	filePath = strings.TrimPrefix(filePath, "/")

	proxyURL, err := h.svc.File.GetProxyURL(c.Request.Context(), filePath)
	if err != nil {
		response.Fail(c, err)
		return
	}
	c.Redirect(302, proxyURL)
}
