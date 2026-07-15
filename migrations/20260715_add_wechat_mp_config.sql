-- ============================================================
-- Sanmoo Blog - 微信小程序配置字段
-- ============================================================
-- 日期: 2026-07-15
-- 说明: 在 t_blog_infrastructure_config 中添加微信开发/生产环境配置
--       支持后台动态切换开发/生产环境，替代原有的配置文件方式
-- ============================================================

SET NAMES utf8mb4 COLLATE utf8mb4_unicode_ci;

-- 添加微信小程序开发环境 AppID
ALTER TABLE `t_blog_infrastructure_config`
    ADD COLUMN IF NOT EXISTS `wx_dev_app_id` varchar(100) NOT NULL DEFAULT '' COMMENT '微信小程序开发环境 AppID';

-- 添加微信小程序开发环境 Secret
ALTER TABLE `t_blog_infrastructure_config`
    ADD COLUMN IF NOT EXISTS `wx_dev_app_secret` varchar(200) NOT NULL DEFAULT '' COMMENT '微信小程序开发环境 Secret';

-- 添加微信小程序生产环境 AppID
ALTER TABLE `t_blog_infrastructure_config`
    ADD COLUMN IF NOT EXISTS `wx_prod_app_id` varchar(100) NOT NULL DEFAULT '' COMMENT '微信小程序生产环境 AppID';

-- 添加微信小程序生产环境 Secret
ALTER TABLE `t_blog_infrastructure_config`
    ADD COLUMN IF NOT EXISTS `wx_prod_app_secret` varchar(200) NOT NULL DEFAULT '' COMMENT '微信小程序生产环境 Secret';

-- 添加微信环境模式：0=开发环境，1=生产环境
ALTER TABLE `t_blog_infrastructure_config`
    ADD COLUMN IF NOT EXISTS `wx_env_mode` tinyint(1) NOT NULL DEFAULT '0' COMMENT '微信环境模式：0=开发环境 1=生产环境';

-- ============================================================
-- 迁移脚本执行完成
-- ============================================================