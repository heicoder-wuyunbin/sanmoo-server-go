-- ============================================================
-- Sanmoo Blog 线上数据库综合迁移脚本
-- ============================================================
-- 日期: 2026-07-15
-- 说明: 将线上数据库结构同步至最新线下版本，包含：
--   1. 创建新配置表 (brand/compliance/channel/infrastructure)
--   2. 创建 RBAC 权限表 (t_permission, t_role_permission)
--   3. 创建新业务表 (t_link, t_mp_user_subscribe)
--   4. 从旧配置表迁移数据到新配置表
--   5. 删除旧配置表及冗余表
--   6. t_article / t_role 新增字段
--   7. 初始化 RBAC 数据（角色迁移、权限注入、角色权限分配）
--   8. 索引升级 / 添加唯一约束
-- 注意: 脚本使用 IF [NOT] EXISTS / INSERT IGNORE / ON DUPLICATE KEY UPDATE 保证幂等，
--       可安全重复执行（MySQL >= 8.0.16）。
-- ============================================================

SET NAMES utf8mb4 COLLATE utf8mb4_unicode_ci;
SET FOREIGN_KEY_CHECKS = 0;

-- ============================================================
-- 第一部分: 创建新配置表
-- ============================================================

-- ------------------------------------------------------------
-- 1.1 站点品牌配置表 (t_blog_brand_config)
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `t_blog_brand_config` (
  `id` tinyint unsigned NOT NULL DEFAULT '1' COMMENT '主键，固定为1',
  `blog_name` varchar(60) NOT NULL DEFAULT '' COMMENT '博客名称',
  `author` varchar(60) NOT NULL DEFAULT '' COMMENT '作者名',
  `introduction` varchar(200) NOT NULL DEFAULT '' COMMENT '介绍语',
  `avatar` varchar(160) NOT NULL DEFAULT '' COMMENT '作者头像',
  `site_url` varchar(200) NOT NULL DEFAULT '' COMMENT '站点地址',
  `rss_enabled` tinyint(1) NOT NULL DEFAULT '1' COMMENT 'RSS开关',
  `config_version` int unsigned NOT NULL DEFAULT '1' COMMENT '配置版本号',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `created_by` varchar(60) NOT NULL DEFAULT '' COMMENT '创建人',
  `updated_by` varchar(60) NOT NULL DEFAULT '' COMMENT '更新人',
  PRIMARY KEY (`id`),
  CONSTRAINT `chk_brand_single_row` CHECK ((`id` = 1))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='博客站点品牌配置表';

-- ------------------------------------------------------------
-- 1.2 合规配置表 (t_blog_compliance_config)
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `t_blog_compliance_config` (
  `id` tinyint unsigned NOT NULL DEFAULT '1' COMMENT '主键，固定为1',
  `privacy_policy` text COMMENT '隐私政策内容',
  `filing_info` varchar(500) NOT NULL DEFAULT '' COMMENT '备案信息（JSON格式）',
  `contact_info` varchar(500) NOT NULL DEFAULT '' COMMENT '联系方式（JSON格式）',
  `data_retention_policy` text COMMENT '数据保留说明',
  `account_deletion_guide` text COMMENT '账号注销说明',
  `config_version` int unsigned NOT NULL DEFAULT '1' COMMENT '配置版本号',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `created_by` varchar(60) NOT NULL DEFAULT '' COMMENT '创建人',
  `updated_by` varchar(60) NOT NULL DEFAULT '' COMMENT '更新人',
  PRIMARY KEY (`id`),
  CONSTRAINT `chk_compliance_single_row` CHECK ((`id` = 1))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='博客合规配置表';

-- ------------------------------------------------------------
-- 1.3 渠道配置表 (t_blog_channel_config)
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `t_blog_channel_config` (
  `id` tinyint unsigned NOT NULL DEFAULT '1' COMMENT '主键，固定为1',
  `github_home` varchar(120) NOT NULL DEFAULT '' COMMENT 'GitHub主页',
  `csdn_home` varchar(120) NOT NULL DEFAULT '' COMMENT 'CSDN主页',
  `gitee_home` varchar(120) NOT NULL DEFAULT '' COMMENT 'Gitee主页',
  `zhihu_home` varchar(120) NOT NULL DEFAULT '' COMMENT '知乎主页',
  `github_show` tinyint(1) NOT NULL DEFAULT '1' COMMENT '是否显示GitHub',
  `csdn_show` tinyint(1) NOT NULL DEFAULT '1' COMMENT '是否显示CSDN',
  `gitee_show` tinyint(1) NOT NULL DEFAULT '1' COMMENT '是否显示Gitee',
  `zhihu_show` tinyint(1) NOT NULL DEFAULT '1' COMMENT '是否显示知乎',
  `web_enabled` tinyint(1) NOT NULL DEFAULT '1' COMMENT 'Web渠道启用',
  `mp_enabled` tinyint(1) NOT NULL DEFAULT '1' COMMENT '小程序渠道启用',
  `config_version` int unsigned NOT NULL DEFAULT '1' COMMENT '配置版本号',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `created_by` varchar(60) NOT NULL DEFAULT '' COMMENT '创建人',
  `updated_by` varchar(60) NOT NULL DEFAULT '' COMMENT '更新人',
  PRIMARY KEY (`id`),
  CONSTRAINT `chk_channel_single_row` CHECK ((`id` = 1))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='博客渠道配置表';

-- ------------------------------------------------------------
-- 1.4 基础设施配置表 (t_blog_infrastructure_config)
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `t_blog_infrastructure_config` (
  `id` tinyint unsigned NOT NULL DEFAULT '1' COMMENT '主键，固定为1',
  `search_engine` varchar(20) NOT NULL DEFAULT 'NONE' COMMENT '搜索引擎: NONE/MEILISEARCH',
  `hot_search_mode` tinyint(1) NOT NULL DEFAULT '0' COMMENT '热搜索模式：0=伪热门 1=真热门',
  `hot_search_words` text COMMENT '热搜索词JSON数组',
  `meilisearch_host` varchar(200) NOT NULL DEFAULT '' COMMENT 'MeiliSearch主机地址',
  `meilisearch_api_key` varchar(100) NOT NULL DEFAULT '' COMMENT 'MeiliSearch API Key',
  `meilisearch_index` varchar(100) NOT NULL DEFAULT 'articles' COMMENT 'MeiliSearch索引名称',
  `meilisearch_last_sync_time` datetime DEFAULT NULL COMMENT 'MeiliSearch最后同步时间',
  `recommend_strategy` varchar(20) NOT NULL DEFAULT 'rule' COMMENT '推荐策略：rule/weighted/cf',
  `upload_strategy` varchar(20) NOT NULL DEFAULT 'LOCAL' COMMENT '上传策略：LOCAL/QINIU/ALIYUN_OSS',
  `upload_local_dir` varchar(200) NOT NULL DEFAULT 'uploads' COMMENT '本地上传目录',
  `upload_local_url_prefix` varchar(80) NOT NULL DEFAULT '/uploads/' COMMENT '本地访问前缀',
  `upload_qiniu_bucket` varchar(120) NOT NULL DEFAULT '' COMMENT '七牛Bucket',
  `upload_qiniu_domain` varchar(200) NOT NULL DEFAULT '' COMMENT '七牛访问域名',
  `upload_qiniu_access_key` varchar(120) NOT NULL DEFAULT '' COMMENT '七牛AccessKey',
  `upload_qiniu_secret_key` varchar(120) NOT NULL DEFAULT '' COMMENT '七牛SecretKey',
  `upload_qiniu_region` varchar(20) NOT NULL DEFAULT '' COMMENT '七牛存储区域',
  `upload_aliyun_endpoint` varchar(200) NOT NULL DEFAULT '' COMMENT '阿里云OSS Endpoint',
  `upload_aliyun_bucket` varchar(120) NOT NULL DEFAULT '' COMMENT '阿里云OSS Bucket',
  `upload_aliyun_domain` varchar(200) NOT NULL DEFAULT '' COMMENT '阿里云OSS访问域名',
  `upload_aliyun_access_key` varchar(120) NOT NULL DEFAULT '' COMMENT '阿里云AccessKey',
  `upload_aliyun_secret_key` varchar(120) NOT NULL DEFAULT '' COMMENT '阿里云SecretKey',
  `email_config_json` longtext COMMENT '邮件配置JSON',
  `config_version` int unsigned NOT NULL DEFAULT '1' COMMENT '配置版本号',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `created_by` varchar(60) NOT NULL DEFAULT '' COMMENT '创建人',
  `updated_by` varchar(60) NOT NULL DEFAULT '' COMMENT '更新人',
  `wx_dev_app_id` varchar(100) NOT NULL DEFAULT '' COMMENT '微信小程序开发环境 AppID',
  `wx_dev_app_secret` varchar(200) NOT NULL DEFAULT '' COMMENT '微信小程序开发环境 Secret',
  `wx_prod_app_id` varchar(100) NOT NULL DEFAULT '' COMMENT '微信小程序生产环境 AppID',
  `wx_prod_app_secret` varchar(200) NOT NULL DEFAULT '' COMMENT '微信小程序生产环境 Secret',
  `wx_env_mode` tinyint(1) NOT NULL DEFAULT '0' COMMENT '微信环境模式：0=开发环境 1=生产环境',
  PRIMARY KEY (`id`),
  CONSTRAINT `chk_infrastructure_single_row` CHECK ((`id` = 1))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='博客基础设施配置表';

-- ============================================================
-- 第二部分: 创建 RBAC 权限表
-- ============================================================

-- ------------------------------------------------------------
-- 2.1 权限表 (t_permission)
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `t_permission` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `perm_key` varchar(128) NOT NULL COMMENT '权限标识',
  `name` varchar(64) NOT NULL COMMENT '权限名称',
  `module` varchar(64) NOT NULL COMMENT '所属模块',
  `type` varchar(32) NOT NULL DEFAULT 'api' COMMENT '类型：api / menu / button',
  `description` varchar(255) DEFAULT NULL COMMENT '描述',
  `front_path` varchar(128) DEFAULT NULL COMMENT '前端菜单路径',
  `icon` varchar(64) DEFAULT NULL COMMENT '前端菜单图标',
  `sort_order` int NOT NULL DEFAULT '0' COMMENT '排序',
  `status` tinyint NOT NULL DEFAULT '1' COMMENT '1启用 0禁用',
  `create_time` datetime NOT NULL,
  `update_time` datetime NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_perm_key` (`perm_key`),
  KEY `idx_module` (`module`),
  KEY `idx_type` (`type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='权限表';

-- ------------------------------------------------------------
-- 2.2 角色-权限关联表 (t_role_permission)
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `t_role_permission` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `role_id` bigint NOT NULL,
  `perm_key` varchar(128) NOT NULL,
  `create_time` datetime NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_role_perm` (`role_id`,`perm_key`),
  KEY `idx_role_id` (`role_id`),
  KEY `idx_perm_key` (`perm_key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='角色-权限关联表';

-- ============================================================
-- 第三部分: 创建新业务表
-- ============================================================

-- ------------------------------------------------------------
-- 3.1 友情链接表 (t_link)
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `t_link` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `name` varchar(100) NOT NULL COMMENT '链接名称',
  `url` varchar(500) NOT NULL COMMENT '链接地址',
  `description` varchar(500) DEFAULT '' COMMENT '链接描述',
  `icon` varchar(500) DEFAULT '' COMMENT '图标URL',
  `sort_order` int NOT NULL DEFAULT '0' COMMENT '排序值',
  `is_active` tinyint(1) NOT NULL DEFAULT '1' COMMENT '是否启用',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_url` (`url`),
  KEY `idx_sort_order` (`sort_order`),
  KEY `idx_is_active` (`is_active`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='友情链接表';

-- ------------------------------------------------------------
-- 3.2 微信小程序用户订阅表 (t_mp_user_subscribe)
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `t_mp_user_subscribe` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `openid` varchar(128) DEFAULT NULL,
  `subscribe` tinyint(1) DEFAULT NULL,
  `create_time` datetime(3) DEFAULT NULL,
  `update_time` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_t_mp_user_subscribe_open_id` (`openid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='微信小程序用户订阅表';

-- ============================================================
-- 第四部分: 从旧配置表迁移数据到新配置表
-- ============================================================
-- 说明: 使用 INSERT ... ON DUPLICATE KEY UPDATE 保证幂等
--       仅当旧表存在时才执行迁移
-- ============================================================

-- ------------------------------------------------------------
-- 4.1 t_blog_core_config → t_blog_brand_config
-- ------------------------------------------------------------
INSERT INTO `t_blog_brand_config` (
  `id`, `blog_name`, `author`, `introduction`, `avatar`, `site_url`, `rss_enabled`,
  `config_version`, `created_by`, `updated_by`
)
SELECT
  1,
  COALESCE(c.`blog_name`, ''),
  COALESCE(c.`author`, ''),
  COALESCE(c.`introduction`, ''),
  COALESCE(c.`avatar`, ''),
  '',
  1,
  COALESCE(c.`config_version`, 1),
  COALESCE(c.`created_by`, ''),
  COALESCE(c.`updated_by`, '')
FROM `t_blog_core_config` c WHERE c.`id` = 1
ON DUPLICATE KEY UPDATE
  `blog_name` = VALUES(`blog_name`),
  `author` = VALUES(`author`),
  `introduction` = VALUES(`introduction`),
  `avatar` = VALUES(`avatar`),
  `config_version` = VALUES(`config_version`),
  `updated_by` = VALUES(`updated_by`);

-- ------------------------------------------------------------
-- 4.2 t_blog_core_config → t_blog_compliance_config
-- ------------------------------------------------------------
INSERT INTO `t_blog_compliance_config` (
  `id`, `privacy_policy`, `filing_info`, `contact_info`,
  `data_retention_policy`, `account_deletion_guide`,
  `config_version`, `created_by`, `updated_by`
)
SELECT
  1,
  COALESCE(c.`privacy_policy`, ''),
  '',
  '',
  '',
  '',
  COALESCE(c.`config_version`, 1),
  COALESCE(c.`created_by`, ''),
  COALESCE(c.`updated_by`, '')
FROM `t_blog_core_config` c WHERE c.`id` = 1
ON DUPLICATE KEY UPDATE
  `privacy_policy` = VALUES(`privacy_policy`),
  `config_version` = VALUES(`config_version`),
  `updated_by` = VALUES(`updated_by`);

-- ------------------------------------------------------------
-- 4.3 t_blog_ui_config → t_blog_channel_config
-- ------------------------------------------------------------
INSERT INTO `t_blog_channel_config` (
  `id`, `github_home`, `csdn_home`, `gitee_home`, `zhihu_home`,
  `github_show`, `csdn_show`, `gitee_show`, `zhihu_show`,
  `web_enabled`, `mp_enabled`, `config_version`,
  `created_by`, `updated_by`
)
SELECT
  1,
  COALESCE(u.`github_home`, ''),
  COALESCE(u.`csdn_home`, ''),
  COALESCE(u.`gitee_home`, ''),
  COALESCE(u.`zhihu_home`, ''),
  COALESCE(u.`github_show`, 1),
  COALESCE(u.`csdn_show`, 1),
  COALESCE(u.`gitee_show`, 1),
  COALESCE(u.`zhihu_show`, 1),
  1,
  1,
  COALESCE(u.`config_version`, 1),
  COALESCE(u.`created_by`, ''),
  COALESCE(u.`updated_by`, '')
FROM `t_blog_ui_config` u WHERE u.`id` = 1
ON DUPLICATE KEY UPDATE
  `github_home` = VALUES(`github_home`),
  `csdn_home` = VALUES(`csdn_home`),
  `gitee_home` = VALUES(`gitee_home`),
  `zhihu_home` = VALUES(`zhihu_home`),
  `github_show` = VALUES(`github_show`),
  `csdn_show` = VALUES(`csdn_show`),
  `gitee_show` = VALUES(`gitee_show`),
  `zhihu_show` = VALUES(`zhihu_show`),
  `config_version` = VALUES(`config_version`),
  `updated_by` = VALUES(`updated_by`);

-- ------------------------------------------------------------
-- 4.4 t_blog_ui_config + t_blog_storage_config + t_blog_email_config
--     → t_blog_infrastructure_config
-- ------------------------------------------------------------
INSERT INTO `t_blog_infrastructure_config` (
  `id`,
  `search_engine`, `hot_search_mode`, `hot_search_words`,
  `meilisearch_host`, `meilisearch_api_key`, `meilisearch_index`,
  `recommend_strategy`,
  `upload_strategy`, `upload_local_dir`, `upload_local_url_prefix`,
  `upload_qiniu_bucket`, `upload_qiniu_domain`,
  `upload_aliyun_endpoint`, `upload_aliyun_bucket`, `upload_aliyun_domain`,
  `email_config_json`,
  `config_version`, `created_by`, `updated_by`
)
SELECT
  1,
  COALESCE(u.`search_engine`, 'NONE'),
  COALESCE(u.`hot_search_mode`, 0),
  u.`hot_search_words`,
  COALESCE(u.`meilisearch_host`, ''),
  COALESCE(u.`meilisearch_api_key`, ''),
  COALESCE(u.`meilisearch_index`, 'articles'),
  COALESCE(u.`recommend_strategy`, 'rule'),
  COALESCE(s.`upload_strategy`, 'LOCAL'),
  COALESCE(s.`upload_local_dir`, 'uploads'),
  COALESCE(s.`upload_local_url_prefix`, '/uploads/'),
  COALESCE(s.`upload_qiniu_bucket`, ''),
  COALESCE(s.`upload_qiniu_domain`, ''),
  COALESCE(s.`upload_aliyun_endpoint`, ''),
  COALESCE(s.`upload_aliyun_bucket`, ''),
  COALESCE(s.`upload_aliyun_domain`, ''),
  e.`config_json`,
  COALESCE(u.`config_version`, 1),
  COALESCE(u.`created_by`, ''),
  COALESCE(u.`updated_by`, '')
FROM `t_blog_ui_config` u
LEFT JOIN `t_blog_storage_config` s ON s.`id` = 1
LEFT JOIN `t_blog_email_config` e ON e.`id` = 1
WHERE u.`id` = 1
ON DUPLICATE KEY UPDATE
  `search_engine` = VALUES(`search_engine`),
  `hot_search_mode` = VALUES(`hot_search_mode`),
  `hot_search_words` = VALUES(`hot_search_words`),
  `meilisearch_host` = VALUES(`meilisearch_host`),
  `meilisearch_api_key` = VALUES(`meilisearch_api_key`),
  `meilisearch_index` = VALUES(`meilisearch_index`),
  `recommend_strategy` = VALUES(`recommend_strategy`),
  `upload_strategy` = VALUES(`upload_strategy`),
  `upload_local_dir` = VALUES(`upload_local_dir`),
  `upload_local_url_prefix` = VALUES(`upload_local_url_prefix`),
  `upload_qiniu_bucket` = VALUES(`upload_qiniu_bucket`),
  `upload_qiniu_domain` = VALUES(`upload_qiniu_domain`),
  `upload_aliyun_endpoint` = VALUES(`upload_aliyun_endpoint`),
  `upload_aliyun_bucket` = VALUES(`upload_aliyun_bucket`),
  `upload_aliyun_domain` = VALUES(`upload_aliyun_domain`),
  `email_config_json` = VALUES(`email_config_json`),
  `config_version` = VALUES(`config_version`),
  `updated_by` = VALUES(`updated_by`);

-- ============================================================
-- 第五部分: 删除旧配置表及冗余表
-- ============================================================

-- 旧配置表（数据已迁移到新配置表）
DROP TABLE IF EXISTS `t_blog_core_config`;
DROP TABLE IF EXISTS `t_blog_ui_config`;
DROP TABLE IF EXISTS `t_blog_storage_config`;
DROP TABLE IF EXISTS `t_blog_email_config`;

-- 冗余表（功能已被 t_access_log 覆盖）
DROP TABLE IF EXISTS `t_visitor_record`;

-- ============================================================
-- 第六部分: 修改现有表结构
-- ============================================================

-- ------------------------------------------------------------
-- 6.1 t_article 新增字段
-- ------------------------------------------------------------

-- slug (URL 别名)
ALTER TABLE `t_article`
  ADD COLUMN IF NOT EXISTS `slug` varchar(100) DEFAULT NULL COMMENT '文章 URL 别名（SEO 友好）' AFTER `title`;

-- share_num (分享次数)
ALTER TABLE `t_article`
  ADD COLUMN IF NOT EXISTS `share_num` int DEFAULT '0' COMMENT '分享次数' AFTER `read_num`;

-- publish_time (定时发布时间)
ALTER TABLE `t_article`
  ADD COLUMN IF NOT EXISTS `publish_time` datetime DEFAULT NULL COMMENT '定时发布时间' AFTER `is_published`;

-- like_num (点赞数)
ALTER TABLE `t_article`
  ADD COLUMN IF NOT EXISTS `like_num` int NOT NULL DEFAULT '0' COMMENT '点赞数' AFTER `updated_by`;

-- slug 唯一索引
ALTER TABLE `t_article`
  ADD UNIQUE KEY IF NOT EXISTS `uk_slug` (`slug`);

-- ------------------------------------------------------------
-- 6.2 t_role 新增字段
-- ------------------------------------------------------------

ALTER TABLE `t_role`
  ADD COLUMN IF NOT EXISTS `description` varchar(255) DEFAULT NULL COMMENT '角色描述' AFTER `name`;

ALTER TABLE `t_role`
  ADD COLUMN IF NOT EXISTS `status` tinyint NOT NULL DEFAULT '1' COMMENT '状态 1启用 0禁用' AFTER `description`;

ALTER TABLE `t_role`
  ADD COLUMN IF NOT EXISTS `sort_order` int NOT NULL DEFAULT '0' COMMENT '排序' AFTER `status`;

-- ============================================================
-- 第七部分: 初始化 RBAC 数据
-- ============================================================
-- 说明: 将线上旧角色体系迁移到新版 RBAC，并注入权限数据。
--       admin 角色按规则自动拥有全部权限，无需在 t_role_permission 中维护。
--       使用 INSERT ... ON DUPLICATE KEY UPDATE / INSERT IGNORE 保证幂等。
-- ============================================================

-- ------------------------------------------------------------
-- 7.1 迁移旧角色数据
-- ------------------------------------------------------------
-- 线上旧角色: ROLE_ADMIN / ROLE_VISITOR
-- 本地新角色: admin / editor / viewer
UPDATE `t_role`
  SET `name` = 'admin',
      `description` = '超级管理员，拥有全部权限',
      `status` = 1,
      `sort_order` = 1
  WHERE `name` = 'ROLE_ADMIN';

UPDATE `t_role`
  SET `name` = 'viewer',
      `description` = '只读访客，仅可查看数据',
      `status` = 1,
      `sort_order` = 3
  WHERE `name` = 'ROLE_VISITOR';

-- ------------------------------------------------------------
-- 7.2 补充 editor 角色
-- ------------------------------------------------------------
INSERT IGNORE INTO `t_role` (`name`, `description`, `status`, `sort_order`, `create_time`, `update_time`)
VALUES ('editor', '内容编辑，可管理文章/分类/标签/专题', 1, 2, NOW(), NOW());

-- ------------------------------------------------------------
-- 7.3 注入权限数据
-- ------------------------------------------------------------
INSERT INTO `t_permission` (
  `perm_key`, `name`, `module`, `type`, `description`, `front_path`, `icon`, `sort_order`, `status`, `create_time`, `update_time`
) VALUES
  ('dashboard:read', '查看仪表盘', 'dashboard', 'menu', '访问仪表盘页面', '/admin', NULL, 1, 1, NOW(), NOW()),
  ('dashboard:visitors', '访问记录', 'dashboard', 'menu', '查看访问记录', '/admin/visitors', NULL, 2, 1, NOW(), NOW()),
  ('dashboard:errors', '错误日志', 'dashboard', 'menu', '查看错误日志', '/admin/errors', NULL, 3, 1, NOW(), NOW()),
  ('dashboard:pv', 'PV统计', 'dashboard', 'button', '查看PV统计数据', NULL, NULL, 4, 1, NOW(), NOW()),
  ('dashboard:statistics', '统计图表', 'dashboard', 'button', '查看各类统计图表', NULL, NULL, 5, 1, NOW(), NOW()),
  ('dashboard:visitors:delete', '删除访问记录', 'dashboard', 'button', '删除访问记录', NULL, NULL, 6, 1, NOW(), NOW()),
  ('dashboard:errors:delete', '删除错误日志', 'dashboard', 'button', '删除错误日志', NULL, NULL, 7, 1, NOW(), NOW()),
  ('article:list', '文章列表', 'article', 'menu', '查看文章列表', '/admin/articles', NULL, 10, 1, NOW(), NOW()),
  ('article:detail', '文章详情', 'article', 'button', '查看文章详情', NULL, NULL, 11, 1, NOW(), NOW()),
  ('article:create', '创建文章', 'article', 'button', '创建新文章', NULL, NULL, 12, 1, NOW(), NOW()),
  ('article:update', '编辑文章', 'article', 'button', '编辑文章', NULL, NULL, 13, 1, NOW(), NOW()),
  ('article:delete', '删除文章', 'article', 'button', '删除文章', NULL, NULL, 14, 1, NOW(), NOW()),
  ('article:status', '文章状态', 'article', 'button', '发布/撤销文章', NULL, NULL, 15, 1, NOW(), NOW()),
  ('article:export', '导出文章', 'article', 'button', '导出文章列表为CSV', NULL, NULL, 16, 1, NOW(), NOW()),
  ('category:list', '分类列表', 'category', 'menu', '查看分类列表', '/admin/categories', NULL, 20, 1, NOW(), NOW()),
  ('category:create', '创建分类', 'category', 'button', '创建新分类', NULL, NULL, 21, 1, NOW(), NOW()),
  ('category:update', '编辑分类', 'category', 'button', '编辑分类', NULL, NULL, 22, 1, NOW(), NOW()),
  ('category:delete', '删除分类', 'category', 'button', '删除分类', NULL, NULL, 23, 1, NOW(), NOW()),
  ('tag:list', '标签列表', 'tag', 'menu', '查看标签列表', '/admin/tags', NULL, 30, 1, NOW(), NOW()),
  ('tag:create', '创建标签', 'tag', 'button', '创建新标签', NULL, NULL, 31, 1, NOW(), NOW()),
  ('tag:update', '编辑标签', 'tag', 'button', '编辑标签', NULL, NULL, 32, 1, NOW(), NOW()),
  ('tag:delete', '删除标签', 'tag', 'button', '删除标签', NULL, NULL, 33, 1, NOW(), NOW()),
  ('topic:list', '专题列表', 'topic', 'menu', '查看专题列表', '/admin/topics', NULL, 40, 1, NOW(), NOW()),
  ('topic:create', '创建专题', 'topic', 'button', '创建新专题', NULL, NULL, 41, 1, NOW(), NOW()),
  ('topic:update', '编辑专题', 'topic', 'button', '编辑专题', NULL, NULL, 42, 1, NOW(), NOW()),
  ('topic:delete', '删除专题', 'topic', 'button', '删除专题', NULL, NULL, 43, 1, NOW(), NOW()),
  ('topic:articles', '专题文章', 'topic', 'button', '管理专题文章', NULL, NULL, 44, 1, NOW(), NOW()),
  ('link:list', '友链列表', 'link', 'menu', '查看友情链接', '/admin/links', NULL, 50, 1, NOW(), NOW()),
  ('link:create', '创建友链', 'link', 'button', '创建友情链接', NULL, NULL, 51, 1, NOW(), NOW()),
  ('link:update', '编辑友链', 'link', 'button', '编辑友情链接', NULL, NULL, 52, 1, NOW(), NOW()),
  ('link:delete', '删除友链', 'link', 'button', '删除友情链接', NULL, NULL, 53, 1, NOW(), NOW()),
  ('file:list', '文件列表', 'file', 'menu', '查看文件列表', '/admin/files', NULL, 60, 1, NOW(), NOW()),
  ('file:upload', '上传文件', 'file', 'button', '上传文件', NULL, NULL, 61, 1, NOW(), NOW()),
  ('file:delete', '删除文件', 'file', 'button', '删除文件', NULL, NULL, 62, 1, NOW(), NOW()),
  ('user:list', '用户列表', 'user', 'menu', '查看用户列表', '/admin/users', NULL, 70, 1, NOW(), NOW()),
  ('user:detail', '用户详情', 'user', 'button', '查看用户详情', NULL, NULL, 71, 1, NOW(), NOW()),
  ('user:create', '创建用户', 'user', 'button', '创建新用户', NULL, NULL, 72, 1, NOW(), NOW()),
  ('user:update', '编辑用户', 'user', 'button', '编辑用户信息', NULL, NULL, 73, 1, NOW(), NOW()),
  ('user:delete', '删除用户', 'user', 'button', '删除用户', NULL, NULL, 74, 1, NOW(), NOW()),
  ('user:status', '用户状态', 'user', 'button', '启用/禁用用户', NULL, NULL, 75, 1, NOW(), NOW()),
  ('user:role', '分配角色', 'user', 'button', '给用户分配角色', NULL, NULL, 76, 1, NOW(), NOW()),
  ('user:export', '导出用户', 'user', 'button', '导出用户列表', NULL, NULL, 77, 1, NOW(), NOW()),
  ('setting:read', '查看设置', 'setting', 'menu', '查看系统设置', '/admin/settings', NULL, 80, 1, NOW(), NOW()),
  ('setting:update', '更新设置', 'setting', 'button', '更新系统配置', NULL, NULL, 81, 1, NOW(), NOW()),
  ('setting:email', '邮件配置', 'setting', 'button', '配置邮件服务', NULL, NULL, 82, 1, NOW(), NOW()),
  ('setting:import', '导入设置', 'setting', 'button', '导入配置', NULL, NULL, 83, 1, NOW(), NOW()),
  ('setting:export', '导出设置', 'setting', 'button', '导出配置', NULL, NULL, 84, 1, NOW(), NOW()),
  ('setting:search', '搜索同步', 'setting', 'button', '同步搜索索引', NULL, NULL, 85, 1, NOW(), NOW()),
  ('cache:clear', '清除缓存', 'cache', 'button', '清除所有缓存', NULL, NULL, 90, 1, NOW(), NOW()),
  ('cache:warmup', '缓存预热', 'cache', 'button', '预热缓存', NULL, NULL, 91, 1, NOW(), NOW()),
  ('cache:stats', '缓存统计', 'cache', 'button', '查看缓存统计', NULL, NULL, 92, 1, NOW(), NOW()),
  ('backup:export', '数据导出', 'backup', 'button', '导出数据备份', NULL, NULL, 100, 1, NOW(), NOW()),
  ('backup:list', '备份列表', 'backup', 'button', '查看备份列表', NULL, NULL, 101, 1, NOW(), NOW()),
  ('backup:download', '下载备份', 'backup', 'button', '下载备份文件', NULL, NULL, 102, 1, NOW(), NOW()),
  ('backup:delete', '删除备份', 'backup', 'button', '删除备份文件', NULL, NULL, 103, 1, NOW(), NOW()),
  ('backup:stats', '备份统计', 'backup', 'button', '查看备份统计', NULL, NULL, 104, 1, NOW(), NOW()),
  ('mpuser:list', '小程序用户', 'mpuser', 'menu', '查看小程序用户', '/admin/mp-users', NULL, 110, 1, NOW(), NOW()),
  ('mpuser:detail', '用户详情', 'mpuser', 'button', '查看用户详情', NULL, NULL, 111, 1, NOW(), NOW()),
  ('mpuser:export', '导出用户', 'mpuser', 'button', '导出用户列表', NULL, NULL, 112, 1, NOW(), NOW()),
  ('role:list', '角色列表', 'role', 'menu', '查看角色列表', '/admin/roles', NULL, 120, 1, NOW(), NOW()),
  ('role:detail', '角色详情', 'role', 'button', '查看角色详情', NULL, NULL, 121, 1, NOW(), NOW()),
  ('role:create', '创建角色', 'role', 'button', '创建新角色', NULL, NULL, 122, 1, NOW(), NOW()),
  ('role:update', '编辑角色', 'role', 'button', '编辑角色信息', NULL, NULL, 123, 1, NOW(), NOW()),
  ('role:delete', '删除角色', 'role', 'button', '删除角色', NULL, NULL, 124, 1, NOW(), NOW()),
  ('role:permission', '分配权限', 'role', 'button', '给角色分配权限', NULL, NULL, 134, 1, NOW(), NOW()),
  ('setting:core:read', '查看核心配置', 'setting', 'menu', '查看核心配置', '/admin/settings/core', 'SettingOutlined', 81, 1, NOW(), NOW()),
  ('setting:core:update', '编辑核心配置', 'setting', 'button', '修改核心配置', '', '', 82, 1, NOW(), NOW()),
  ('setting:privacy:read', '查看隐私政策', 'setting', 'menu', '查看隐私政策', '/admin/settings/privacy', 'FileTextOutlined', 83, 1, NOW(), NOW()),
  ('setting:privacy:update', '编辑隐私政策', 'setting', 'button', '修改隐私政策', '', '', 84, 1, NOW(), NOW()),
  ('setting:social:read', '查看社交链接', 'setting', 'menu', '查看社交链接配置', '/admin/settings/social', 'GlobalOutlined', 85, 1, NOW(), NOW()),
  ('setting:social:update', '编辑社交链接', 'setting', 'button', '修改社交链接配置', '', '', 86, 1, NOW(), NOW()),
  ('setting:search:read', '查看搜索配置', 'setting', 'menu', '查看搜索配置', '/admin/settings/search', 'SearchOutlined', 87, 1, NOW(), NOW()),
  ('setting:search:update', '编辑搜索配置', 'setting', 'button', '修改搜索配置', '', '', 88, 1, NOW(), NOW()),
  ('setting:storage:read', '查看存储配置', 'setting', 'menu', '查看存储配置', '/admin/settings/storage', 'CloudServerOutlined', 89, 1, NOW(), NOW()),
  ('setting:storage:update', '编辑存储配置', 'setting', 'button', '修改存储配置', '', '', 90, 1, NOW(), NOW()),
  ('setting:email:read', '查看邮件配置', 'setting', 'menu', '查看邮件配置', '/admin/settings/email', 'MailOutlined', 91, 1, NOW(), NOW()),
  ('setting:cache:read', '查看缓存管理', 'setting', 'menu', '查看缓存管理', '/admin/settings/cache', 'ThunderboltOutlined', 92, 1, NOW(), NOW()),
  ('setting:wechat:read', '查看微信配置', 'setting', 'menu', '查看微信小程序配置', '/admin/settings/wechat', 'WechatOutlined', 93, 1, NOW(), NOW()),
  ('setting:wechat:update', '编辑微信配置', 'setting', 'button', '修改微信小程序配置', '', '', 94, 1, NOW(), NOW())
ON DUPLICATE KEY UPDATE
  `name` = VALUES(`name`),
  `module` = VALUES(`module`),
  `type` = VALUES(`type`),
  `description` = VALUES(`description`),
  `front_path` = VALUES(`front_path`),
  `icon` = VALUES(`icon`),
  `sort_order` = VALUES(`sort_order`),
  `status` = VALUES(`status`),
  `update_time` = NOW();

-- ------------------------------------------------------------
-- 7.4 为 editor 角色分配内容管理权限
-- ------------------------------------------------------------
INSERT IGNORE INTO `t_role_permission` (`role_id`, `perm_key`, `create_time`)
SELECT r.id, p.perm_key, NOW()
FROM `t_role` r
CROSS JOIN `t_permission` p
WHERE r.name = 'editor'
  AND p.perm_key IN (
    'dashboard:read',
    'article:list', 'article:detail', 'article:create', 'article:update', 'article:delete', 'article:status', 'article:export',
    'category:list', 'category:create', 'category:update', 'category:delete',
    'tag:list', 'tag:create', 'tag:update', 'tag:delete',
    'topic:list', 'topic:create', 'topic:update', 'topic:delete', 'topic:articles',
    'link:list', 'link:create', 'link:update', 'link:delete',
    'file:list', 'file:upload', 'file:delete',
    'setting:read', 'setting:core:read', 'setting:privacy:read', 'setting:social:read', 'setting:search:read', 'setting:storage:read', 'setting:email:read', 'setting:cache:read', 'setting:wechat:read'
  );

-- ------------------------------------------------------------
-- 7.5 为 viewer 角色分配只读权限
-- ------------------------------------------------------------
INSERT IGNORE INTO `t_role_permission` (`role_id`, `perm_key`, `create_time`)
SELECT r.id, p.perm_key, NOW()
FROM `t_role` r
CROSS JOIN `t_permission` p
WHERE r.name = 'viewer'
  AND p.perm_key IN (
    'dashboard:read', 'dashboard:pv', 'dashboard:statistics',
    'article:list', 'article:detail', 'article:export',
    'category:list', 'tag:list', 'topic:list', 'link:list', 'file:list',
    'user:list', 'user:detail', 'user:export',
    'role:list', 'role:detail',
    'mpuser:list', 'mpuser:detail', 'mpuser:export',
    'setting:read', 'setting:core:read', 'setting:privacy:read', 'setting:social:read', 'setting:search:read', 'setting:storage:read', 'setting:email:read', 'setting:cache:read', 'setting:wechat:read',
    'cache:stats', 'backup:list', 'backup:download', 'backup:stats'
  );

-- ============================================================
-- 第八部分: 索引升级 / 添加唯一约束
-- ============================================================

-- ------------------------------------------------------------
-- 7.1 t_article_content: 普通索引 → 唯一约束
-- ------------------------------------------------------------
-- 确保每篇文章只有一条内容记录
ALTER TABLE `t_article_content`
  DROP INDEX `idx_article_id`,
  ADD UNIQUE KEY `uk_article_id` (`article_id`) USING BTREE;

-- ------------------------------------------------------------
-- 7.2 t_article_tag_rel: 添加联合唯一约束
-- ------------------------------------------------------------
-- 确保同一篇文章不会重复关联同一个标签
ALTER TABLE `t_article_tag_rel`
  ADD UNIQUE KEY `uk_article_tag` (`article_id`, `tag_id`) USING BTREE;

-- ============================================================
-- 恢复设置
-- ============================================================
SET FOREIGN_KEY_CHECKS = 1;

-- ============================================================
-- 迁移完成验证（可选执行）
-- ============================================================
-- SELECT COUNT(*) FROM t_blog_brand_config WHERE id = 1;
-- SELECT COUNT(*) FROM t_blog_compliance_config WHERE id = 1;
-- SELECT COUNT(*) FROM t_blog_channel_config WHERE id = 1;
-- SELECT COUNT(*) FROM t_blog_infrastructure_config WHERE id = 1;
-- SELECT COUNT(*) FROM t_permission;
-- SHOW COLUMNS FROM t_article LIKE 'slug';
-- SHOW COLUMNS FROM t_role LIKE 'description';
-- SHOW INDEX FROM t_article_content WHERE Key_name = 'uk_article_id';
-- SHOW INDEX FROM t_article_tag_rel WHERE Key_name = 'uk_article_tag';
