-- ============================================================
-- Sanmoo Blog L1 配置域重构 - 创建新配置表
-- ============================================================
-- 日期: 2026-07-09
-- 说明: 创建四个新配置表，与旧表共存，不影响现有功能
-- 执行顺序: 1/2
-- ============================================================

SET NAMES utf8mb4 COLLATE utf8mb4_unicode_ci;

-- ------------------------------------------------------------
-- 1. 站点品牌配置表 (t_blog_brand_config)
-- ------------------------------------------------------------
-- 包含: 博客名、作者、介绍、头像、站点地址、SEO基础项
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
-- 2. 合规配置表 (t_blog_compliance_config)
-- ------------------------------------------------------------
-- 包含: 隐私政策、数据保留说明、账号注销说明、备案信息展示
-- ------------------------------------------------------------

CREATE TABLE IF NOT EXISTS `t_blog_compliance_config` (
  `id` tinyint unsigned NOT NULL DEFAULT '1' COMMENT '主键，固定为1',
  `privacy_policy` text COMMENT '隐私政策内容',
  `filing_info` varchar(500) NOT NULL DEFAULT '' COMMENT '备案信息（JSON格式：{\"icpCode\":\"\",\"filingUrl\":\"\",\"recordType\":\"个人\"}）',
  `contact_info` varchar(500) NOT NULL DEFAULT '' COMMENT '联系方式（JSON格式：{\"email\":\"\",\"wechat\":\"\",\"github\":\"\"}）',
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
-- 3. 渠道配置表 (t_blog_channel_config)
-- ------------------------------------------------------------
-- 包含: Web渠道配置、小程序渠道配置、社交链接展示开关
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
-- 4. 基础设施配置表 (t_blog_infrastructure_config)
-- ------------------------------------------------------------
-- 包含: 搜索、存储、邮件、缓存等技术配置
-- ------------------------------------------------------------

CREATE TABLE IF NOT EXISTS `t_blog_infrastructure_config` (
  `id` tinyint unsigned NOT NULL DEFAULT '1' COMMENT '主键，固定为1',
  `search_engine` varchar(20) NOT NULL DEFAULT 'NONE' COMMENT '搜索引擎: NONE/MEILISEARCH',
  `hot_search_mode` tinyint(1) NOT NULL DEFAULT '0' COMMENT '热搜索模式：0=伪热门(FAKE) 1=真热门(REAL)',
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
  `upload_aliyun_endpoint` varchar(200) NOT NULL DEFAULT '' COMMENT '阿里云OSS Endpoint',
  `upload_aliyun_bucket` varchar(120) NOT NULL DEFAULT '' COMMENT '阿里云OSS Bucket',
  `upload_aliyun_domain` varchar(200) NOT NULL DEFAULT '' COMMENT '阿里云OSS访问域名',
  `email_config_json` longtext COMMENT '邮件配置JSON',
  `config_version` int unsigned NOT NULL DEFAULT '1' COMMENT '配置版本号',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `created_by` varchar(60) NOT NULL DEFAULT '' COMMENT '创建人',
  `updated_by` varchar(60) NOT NULL DEFAULT '' COMMENT '更新人',
  PRIMARY KEY (`id`),
  CONSTRAINT `chk_infrastructure_single_row` CHECK ((`id` = 1))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='博客基础设施配置表';

-- ============================================================
-- 迁移脚本执行完成
-- ============================================================
-- 下一步: 执行 20260709_migrate_config_data.sql 迁移数据
-- ============================================================
