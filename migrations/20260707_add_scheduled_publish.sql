-- ============================================================
-- 文章定时发布字段迁移脚本
-- 日期: 2026-07-07
-- 说明: 新增 publish_time 字段，支持文章定时发布功能
-- ============================================================

-- --------------------------------------------------------
-- 1. 新增 publish_time 字段（可空）
-- --------------------------------------------------------
ALTER TABLE t_article ADD COLUMN publish_time DATETIME DEFAULT NULL COMMENT '定时发布时间' AFTER is_published;

-- 验证
SELECT COUNT(*) as total, COUNT(publish_time) as scheduled FROM t_article;