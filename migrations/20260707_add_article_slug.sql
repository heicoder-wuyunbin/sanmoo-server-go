-- ============================================================
-- 文章 Slug 字段迁移脚本
-- 日期: 2026-07-07
-- 说明: 新增 slug 字段支持伪静态 URL，提升 SEO 效果
-- ============================================================

-- --------------------------------------------------------
-- 1. 新增 slug 字段（可空，唯一索引）
-- --------------------------------------------------------
ALTER TABLE t_article ADD COLUMN slug VARCHAR(100) DEFAULT NULL COMMENT '文章 URL 别名（SEO 友好）' AFTER title;

-- 添加唯一索引
ALTER TABLE t_article ADD UNIQUE INDEX uk_slug (slug);

-- --------------------------------------------------------
-- 2. 为现有文章自动生成 slug（基于 id）
-- --------------------------------------------------------
-- 注意：此处仅简单用 id 作为初始 slug，后续可由后端服务自动生成更友好的 slug
-- MySQL 8.4 不支持在同一语句中引用同一表，故用子查询方式
UPDATE t_article SET slug = CONCAT('article-', id) WHERE slug IS NULL;

-- 验证
SELECT COUNT(*) as total, COUNT(slug) as with_slug FROM t_article;