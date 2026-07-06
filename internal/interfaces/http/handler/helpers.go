package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

const (
	defaultPage     = 1
	defaultPageSize = 10
	maxPageSize      = 200
)

func parseIntDefault(raw string, def int) int {
	if raw == "" {
		return def
	}
	v, err := strconv.Atoi(raw)
	if err != nil || v <= 0 {
		return def
	}
	return v
}

func parseUintDefault(raw string, def uint64) uint64 {
	if raw == "" {
		return def
	}
	v, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		return def
	}
	return v
}

func parsePagination(c *gin.Context) (page, size int) {
	page = parseIntDefault(c.Query("page"), defaultPage)
	size = parseIntDefault(c.Query("size"), defaultPageSize)
	if size > maxPageSize {
		size = maxPageSize
	}
	return
}

func getOpenID(c *gin.Context) string {
	openID := c.Query("openid")
	if openID == "" {
		openID = c.GetHeader("X-MP-OPENID")
	}
	return openID
}
