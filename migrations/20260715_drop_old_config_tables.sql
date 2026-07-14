-- ============================================================
-- Sanmoo Blog L3 配置域重构 - 删除旧配置表
-- ============================================================
-- 日期: 2026-07-15
-- 说明: 删除已迁移到新配置表的旧表，完成配置域重构收尾
-- 执行顺序: 3/3
-- 依赖: 
--   - 20260709_create_new_config_tables.sql (L1 创建新表)
--   - 20260709_migrate_config_data.sql (L2 数据迁移)
--   - 后端代码已迁移到新表读写
-- ============================================================

SET NAMES utf8mb4 COLLATE utf8mb4_unicode_ci;

-- ------------------------------------------------------------
-- 删除旧配置表
-- ------------------------------------------------------------
-- 说明: 以下旧表的数据已迁移到新表，代码已不再引用这些表
-- ------------------------------------------------------------

-- 1. 删除核心配置表（已迁移到 t_blog_brand_config）
DROP TABLE IF EXISTS `t_blog_core_config`;

-- 2. 删除 UI 配置表（已拆分到 t_blog_channel_config 和 t_blog_infrastructure_config）
DROP TABLE IF EXISTS `t_blog_ui_config`;

-- 3. 删除隐私配置表（已迁移到 t_blog_compliance_config）
DROP TABLE IF EXISTS `t_blog_privacy_config`;

-- 4. 删除社交配置表（已迁移到 t_blog_channel_config）
DROP TABLE IF EXISTS `t_blog_social_config`;

-- 5. 删除搜索配置表（已迁移到 t_blog_infrastructure_config）
DROP TABLE IF EXISTS `t_blog_search_config`;

-- 6. 删除存储配置表（已迁移到 t_blog_infrastructure_config）
DROP TABLE IF EXISTS `t_blog_storage_config`;

-- 7. 删除邮件配置表（已迁移到 t_blog_infrastructure_config）
DROP TABLE IF EXISTS `t_blog_email_config`;

-- ============================================================
-- 迁移完成验证
-- ============================================================
-- 验证新表存在且有数据
-- SELECT COUNT(*) FROM t_blog_brand_config WHERE id = 1;
-- SELECT COUNT(*) FROM t_blog_compliance_config WHERE id = 1;
-- SELECT COUNT(*) FROM t_blog_channel_config WHERE id = 1;
-- SELECT COUNT(*) FROM t_blog_infrastructure_config WHERE id = 1;

-- ============================================================
-- 迁移脚本执行完成
-- ============================================================
-- P1 阶段配置域重构正式完成
-- 新配置模型：
--   - t_blog_brand_config: 站点品牌配置
--   - t_blog_compliance_config: 合规配置
--   - t_blog_channel_config: 渠道配置
--   - t_blog_infrastructure_config: 基础设施配置
-- ============================================================
