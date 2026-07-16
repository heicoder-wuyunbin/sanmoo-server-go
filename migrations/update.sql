-- ============================================================================
-- Sanmoo Blog 数据库迁移脚本
-- 基准: sanmoo_blog_schema.sql (2026-07-16)
-- 说明: 本脚本为幂等脚本，可重复执行，用于未来增量迁移
-- ============================================================================

-- ============================================================================
-- 迁移 001: t_statistics_article_pv 添加 pv_date 索引
-- 原因: Dashboard 按日期范围查询 PV 统计时，当前仅靠联合主键 (article_id, pv_date)
--       无法高效利用 pv_date 进行范围扫描，需单独索引
-- 执行日期: 2026-07-16
-- ============================================================================
ALTER TABLE `t_statistics_article_pv`
    ADD INDEX `idx_pv_date` (`pv_date`) USING BTREE;

-- ============================================================================
-- 迁移 002: P0 - 简化 RBAC 为 is_admin 单字段
-- 原因: 个人博客仅 1 个管理员用户，五表 RBAC 严重过度设计
-- 执行日期: 2026-07-16
-- ============================================================================

-- Step 1: 新增 is_admin 字段
ALTER TABLE `t_user`
    ADD COLUMN `is_admin` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否管理员：0否 1是'
    AFTER `updated_by`;

-- Step 2: 将现有用户设为管理员
UPDATE `t_user` SET `is_admin` = 1 WHERE `id` = 1;

-- Step 3: 删除 RBAC 四表（数据已无依赖）
DROP TABLE IF EXISTS `t_role_permission`;
DROP TABLE IF EXISTS `t_user_role`;
DROP TABLE IF EXISTS `t_permission`;
DROP TABLE IF EXISTS `t_role`;

-- ============================================================================
-- 迁移 003: P3 - 删除用户画像相关表
-- 原因: 个人博客用户量小，画像/标签/兴趣/推荐曝光无实际价值
-- 执行日期: 2026-07-16
-- ============================================================================
DROP TABLE IF EXISTS `t_mp_user_behavior`;
DROP TABLE IF EXISTS `t_mp_user_interest`;
DROP TABLE IF EXISTS `t_mp_user_tag`;
DROP TABLE IF EXISTS `t_mp_user_profile`;
DROP TABLE IF EXISTS `t_mp_reco_exposure`;
DROP TABLE IF EXISTS `t_mp_user_subscribe`;

-- ============================================================================
-- 迁移 004: P4 - 清理无用表
-- 原因: 从未使用或数据量极少
-- 执行日期: 2026-07-16
-- ============================================================================
DROP TABLE IF EXISTS `t_search_history`;
DROP TABLE IF EXISTS `t_third_party_log`;

-- ============================================================================
-- 以下为验证查询，确认迁移结果
-- ============================================================================

-- 验证: 确认所有新配置表已存在
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

-- 验证: 确认旧配置表已清理
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

-- 验证: 确认 RBAC 四表已删除
SELECT 'check: removed t_role_permission' AS check_point, COUNT(*) AS row_count
FROM information_schema.TABLES
WHERE TABLE_SCHEMA = 'sanmoo_blog' AND TABLE_NAME = 't_role_permission';

SELECT 'check: removed t_user_role' AS check_point, COUNT(*) AS row_count
FROM information_schema.TABLES
WHERE TABLE_SCHEMA = 'sanmoo_blog' AND TABLE_NAME = 't_user_role';

SELECT 'check: removed t_permission' AS check_point, COUNT(*) AS row_count
FROM information_schema.TABLES
WHERE TABLE_SCHEMA = 'sanmoo_blog' AND TABLE_NAME = 't_permission';

SELECT 'check: removed t_role' AS check_point, COUNT(*) AS row_count
FROM information_schema.TABLES
WHERE TABLE_SCHEMA = 'sanmoo_blog' AND TABLE_NAME = 't_role';

-- 验证: 确认用户画像表已删除
SELECT 'check: removed t_mp_user_behavior' AS check_point, COUNT(*) AS row_count
FROM information_schema.TABLES
WHERE TABLE_SCHEMA = 'sanmoo_blog' AND TABLE_NAME = 't_mp_user_behavior';

SELECT 'check: removed t_mp_user_interest' AS check_point, COUNT(*) AS row_count
FROM information_schema.TABLES
WHERE TABLE_SCHEMA = 'sanmoo_blog' AND TABLE_NAME = 't_mp_user_interest';

SELECT 'check: removed t_mp_user_tag' AS check_point, COUNT(*) AS row_count
FROM information_schema.TABLES
WHERE TABLE_SCHEMA = 'sanmoo_blog' AND TABLE_NAME = 't_mp_user_tag';

SELECT 'check: removed t_mp_user_profile' AS check_point, COUNT(*) AS row_count
FROM information_schema.TABLES
WHERE TABLE_SCHEMA = 'sanmoo_blog' AND TABLE_NAME = 't_mp_user_profile';

SELECT 'check: removed t_mp_reco_exposure' AS check_point, COUNT(*) AS row_count
FROM information_schema.TABLES
WHERE TABLE_SCHEMA = 'sanmoo_blog' AND TABLE_NAME = 't_mp_reco_exposure';

SELECT 'check: removed t_mp_user_subscribe' AS check_point, COUNT(*) AS row_count
FROM information_schema.TABLES
WHERE TABLE_SCHEMA = 'sanmoo_blog' AND TABLE_NAME = 't_mp_user_subscribe';

-- 验证: 确认无用表已删除
SELECT 'check: removed t_search_history' AS check_point, COUNT(*) AS row_count
FROM information_schema.TABLES
WHERE TABLE_SCHEMA = 'sanmoo_blog' AND TABLE_NAME = 't_search_history';

SELECT 'check: removed t_third_party_log' AS check_point, COUNT(*) AS row_count
FROM information_schema.TABLES
WHERE TABLE_SCHEMA = 'sanmoo_blog' AND TABLE_NAME = 't_third_party_log';

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

-- 验证: 确认 t_user.is_admin 字段已添加
SELECT 'check: t_user.is_admin' AS check_point, COUNT(*) AS row_count
FROM information_schema.COLUMNS
WHERE TABLE_SCHEMA = 'sanmoo_blog' AND TABLE_NAME = 't_user' AND COLUMN_NAME = 'is_admin';

-- 验证: 确认 t_statistics_article_pv.idx_pv_date 索引已创建
SELECT 'check: t_statistics_article_pv.idx_pv_date' AS check_point, COUNT(*) AS row_count
FROM information_schema.STATISTICS
WHERE TABLE_SCHEMA = 'sanmoo_blog' AND TABLE_NAME = 't_statistics_article_pv' AND INDEX_NAME = 'idx_pv_date';

-- ============================================================================
-- 验证通过则说明数据库迁移完成
-- 删除的表 check 结果应为 0（不存在），其他 check 结果应为 1（存在）
-- ============================================================================