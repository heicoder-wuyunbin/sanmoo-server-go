-- ============================================================================
-- Sanmoo Blog 数据库迁移脚本
-- 基准: sanmoo_blog_schema.sql (2026-07-16)
-- 状态: 当前在线数据库与基准 schema 完全一致，无需结构性变更
-- 说明: 本脚本为幂等脚本，可重复执行，用于未来增量迁移
-- ============================================================================

-- 验证: 确认所有新配置表已存在（P0/P1 配置域重构产物）
SELECT 'check: t_blog_brand_config' AS check_point, COUNT(*) AS row_count
FROM information_schema.TABLES
WHERE TABLE_SCHEMA = 'sanmoo_blog' AND TABLE_NAME = 't_blog_brand_config';

SELECT 'check: t_blog_channel_config' AS check_point, COUNT(*) AS row_count
FROM information_schema.TABLES
WHERE TABLE_SCHEMA = 'sanmoo_blog' AND TABLE_NAME = 't_blog_channel_config';

SELECT 'check: t_blog_compliance_config' AS check_point, COUNT(*) AS row_count
FROM information_schema.TABLES
WHERE TABLE_SCHEMA = 'sanmoo_blog' AND TABLE_NAME = 't_blog_compliance_config';

SELECT 'check: t_blog_infrastructure_config' AS check_point, COUNT(*) AS row_count
FROM information_schema.TABLES
WHERE TABLE_SCHEMA = 'sanmoo_blog' AND TABLE_NAME = 't_blog_infrastructure_config';

-- 验证: 确认旧配置表已清理（不应存在）
SELECT 'check: old t_blog_core_config' AS check_point, COUNT(*) AS row_count
FROM information_schema.TABLES
WHERE TABLE_SCHEMA = 'sanmoo_blog' AND TABLE_NAME = 't_blog_core_config';

SELECT 'check: old t_blog_privacy_config' AS check_point, COUNT(*) AS row_count
FROM information_schema.TABLES
WHERE TABLE_SCHEMA = 'sanmoo_blog' AND TABLE_NAME = 't_blog_privacy_config';

SELECT 'check: old t_blog_ui_config' AS check_point, COUNT(*) AS row_count
FROM information_schema.TABLES
WHERE TABLE_SCHEMA = 'sanmoo_blog' AND TABLE_NAME = 't_blog_ui_config';

SELECT 'check: old t_blog_social_config' AS check_point, COUNT(*) AS row_count
FROM information_schema.TABLES
WHERE TABLE_SCHEMA = 'sanmoo_blog' AND TABLE_NAME = 't_blog_social_config';

SELECT 'check: old t_blog_search_config' AS check_point, COUNT(*) AS row_count
FROM information_schema.TABLES
WHERE TABLE_SCHEMA = 'sanmoo_blog' AND TABLE_NAME = 't_blog_search_config';

SELECT 'check: old t_blog_storage_config' AS check_point, COUNT(*) AS row_count
FROM information_schema.TABLES
WHERE TABLE_SCHEMA = 'sanmoo_blog' AND TABLE_NAME = 't_blog_storage_config';

SELECT 'check: old t_blog_email_config' AS check_point, COUNT(*) AS row_count
FROM information_schema.TABLES
WHERE TABLE_SCHEMA = 'sanmoo_blog' AND TABLE_NAME = 't_blog_email_config';

SELECT 'check: old t_blog_wechat_config' AS check_point, COUNT(*) AS row_count
FROM information_schema.TABLES
WHERE TABLE_SCHEMA = 'sanmoo_blog' AND TABLE_NAME = 't_blog_wechat_config';

-- 验证: 确认 t_visitor_record 已删除
SELECT 'check: old t_visitor_record' AS check_point, COUNT(*) AS row_count
FROM information_schema.TABLES
WHERE TABLE_SCHEMA = 'sanmoo_blog' AND TABLE_NAME = 't_visitor_record';

-- 验证: 所有关键唯一约束已存在
SELECT 'check: t_article_content.uk_article_id' AS check_point, COUNT(*) AS row_count
FROM information_schema.TABLE_CONSTRAINTS
WHERE TABLE_SCHEMA = 'sanmoo_blog' AND TABLE_NAME = 't_article_content' AND CONSTRAINT_NAME = 'uk_article_id';

SELECT 'check: t_article_tag_rel.uk_article_tag' AS check_point, COUNT(*) AS row_count
FROM information_schema.TABLE_CONSTRAINTS
WHERE TABLE_SCHEMA = 'sanmoo_blog' AND TABLE_NAME = 't_article_tag_rel' AND CONSTRAINT_NAME = 'uk_article_tag';

SELECT 'check: t_article_category_rel.uk_article_category' AS check_point, COUNT(*) AS row_count
FROM information_schema.TABLE_CONSTRAINTS
WHERE TABLE_SCHEMA = 'sanmoo_blog' AND TABLE_NAME = 't_article_category_rel' AND CONSTRAINT_NAME = 'uk_article_category';

SELECT 'check: t_article_topic_rel.uk_article_topic' AS check_point, COUNT(*) AS row_count
FROM information_schema.TABLE_CONSTRAINTS
WHERE TABLE_SCHEMA = 'sanmoo_blog' AND TABLE_NAME = 't_article_topic_rel' AND CONSTRAINT_NAME = 'uk_article_topic';

SELECT 'check: t_mp_user_favorite.uk_openid_article' AS check_point, COUNT(*) AS row_count
FROM information_schema.TABLE_CONSTRAINTS
WHERE TABLE_SCHEMA = 'sanmoo_blog' AND TABLE_NAME = 't_mp_user_favorite' AND CONSTRAINT_NAME = 'uk_openid_article';

-- ============================================================================
-- 验证通过则说明数据库与 sanmoo_blog_schema.sql 完全一致，无需变更
-- 所有 check 结果应为 1（表存在/约束存在），旧表检查结果应为 0（不存在）
-- ============================================================================