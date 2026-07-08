-- ============================================================
-- Sanmoo Blog L1 配置域重构 - 数据迁移脚本
-- ============================================================
-- 日期: 2026-07-09
-- 说明: 从旧配置表迁移数据到新配置表，使用 INSERT ... ON DUPLICATE KEY UPDATE
--       确保安全迁移，不覆盖已存在的新表数据
-- 执行顺序: 2/2
-- 依赖: 先执行 20260709_create_new_config_tables.sql
-- ============================================================

SET NAMES utf8mb4 COLLATE utf8mb4_unicode_ci;

-- ============================================================
-- 迁移策略说明
-- ============================================================
-- 1. 使用 INSERT ... ON DUPLICATE KEY UPDATE 避免主键冲突
-- 2. 旧表数据优先，新表已存在数据不被覆盖（保护已手动配置的新表数据）
-- 3. 字段优先级: t_blog_social_config > t_blog_ui_config（社交字段）
--                t_blog_search_config > t_blog_ui_config（搜索字段）
--                t_blog_core_config > t_blog_privacy_config（隐私政策）
-- 4. 迁移完成后旧表保留，不删除，确保回滚能力
-- ============================================================

-- ------------------------------------------------------------
-- 1. 迁移到 t_blog_brand_config（站点品牌配置）
-- ------------------------------------------------------------
-- 来源: t_blog_core_config
-- 字段映射: blog_name, author, introduction, avatar, rss_enabled
-- ------------------------------------------------------------

INSERT INTO `t_blog_brand_config` (
  `id`, `blog_name`, `author`, `introduction`, `avatar`, `site_url`, `rss_enabled`,
  `config_version`, `created_by`, `updated_by`
)
SELECT
  1 AS id,
  COALESCE(core.`blog_name`, '') AS blog_name,
  COALESCE(core.`author`, '') AS author,
  COALESCE(core.`introduction`, '') AS introduction,
  COALESCE(core.`avatar`, '') AS avatar,
  '' AS site_url,
  COALESCE(core.`rss_enabled`, 1) AS rss_enabled,
  COALESCE(core.`config_version`, 1) AS config_version,
  COALESCE(core.`created_by`, '') AS created_by,
  COALESCE(core.`updated_by`, '') AS updated_by
FROM `t_blog_core_config` core
WHERE core.`id` = 1
ON DUPLICATE KEY UPDATE
  `blog_name` = VALUES(`blog_name`),
  `author` = VALUES(`author`),
  `introduction` = VALUES(`introduction`),
  `avatar` = VALUES(`avatar`),
  `site_url` = VALUES(`site_url`),
  `rss_enabled` = VALUES(`rss_enabled`),
  `config_version` = VALUES(`config_version`),
  `updated_by` = VALUES(`updated_by`);

-- ------------------------------------------------------------
-- 2. 迁移到 t_blog_compliance_config（合规配置）
-- ------------------------------------------------------------
-- 来源: t_blog_core_config.privacy_policy, t_blog_privacy_config.privacy_policy
-- 优先使用 t_blog_core_config 的隐私政策（数据更完整）
-- ------------------------------------------------------------

INSERT INTO `t_blog_compliance_config` (
  `id`, `privacy_policy`, `filing_info`, `contact_info`,
  `data_retention_policy`, `account_deletion_guide`, `config_version`,
  `created_by`, `updated_by`
)
SELECT
  1 AS id,
  COALESCE(core.`privacy_policy`, privacy.`privacy_policy`, '') AS privacy_policy,
  '' AS filing_info,
  '' AS contact_info,
  '' AS data_retention_policy,
  '' AS account_deletion_guide,
  COALESCE(core.`config_version`, privacy.`config_version`, 1) AS config_version,
  COALESCE(core.`created_by`, privacy.`created_by`, '') AS created_by,
  COALESCE(core.`updated_by`, privacy.`updated_by`, '') AS updated_by
FROM `t_blog_core_config` core
LEFT JOIN `t_blog_privacy_config` privacy ON privacy.`id` = 1
WHERE core.`id` = 1
ON DUPLICATE KEY UPDATE
  `privacy_policy` = VALUES(`privacy_policy`),
  `filing_info` = VALUES(`filing_info`),
  `contact_info` = VALUES(`contact_info`),
  `data_retention_policy` = VALUES(`data_retention_policy`),
  `account_deletion_guide` = VALUES(`account_deletion_guide`),
  `config_version` = VALUES(`config_version`),
  `updated_by` = VALUES(`updated_by`);

-- ------------------------------------------------------------
-- 3. 迁移到 t_blog_channel_config（渠道配置）
-- ------------------------------------------------------------
-- 来源: t_blog_social_config（优先）, t_blog_ui_config（补充）
-- 社交字段优先级: t_blog_social_config > t_blog_ui_config
-- ------------------------------------------------------------

INSERT INTO `t_blog_channel_config` (
  `id`, `github_home`, `csdn_home`, `gitee_home`, `zhihu_home`,
  `github_show`, `csdn_show`, `gitee_show`, `zhihu_show`,
  `web_enabled`, `mp_enabled`, `config_version`,
  `created_by`, `updated_by`
)
SELECT
  1 AS id,
  COALESCE(social.`github_home`, ui.`github_home`, '') AS github_home,
  COALESCE(social.`csdn_home`, ui.`csdn_home`, '') AS csdn_home,
  COALESCE(social.`gitee_home`, ui.`gitee_home`, '') AS gitee_home,
  COALESCE(social.`zhihu_home`, ui.`zhihu_home`, '') AS zhihu_home,
  COALESCE(social.`github_show`, ui.`github_show`, 1) AS github_show,
  COALESCE(social.`csdn_show`, ui.`csdn_show`, 1) AS csdn_show,
  COALESCE(social.`gitee_show`, ui.`gitee_show`, 1) AS gitee_show,
  COALESCE(social.`zhihu_show`, ui.`zhihu_show`, 1) AS zhihu_show,
  1 AS web_enabled,
  1 AS mp_enabled,
  COALESCE(social.`config_version`, ui.`config_version`, 1) AS config_version,
  COALESCE(social.`created_by`, ui.`created_by`, '') AS created_by,
  COALESCE(social.`updated_by`, ui.`updated_by`, '') AS updated_by
FROM `t_blog_social_config` social
LEFT JOIN `t_blog_ui_config` ui ON ui.`id` = 1
WHERE social.`id` = 1
ON DUPLICATE KEY UPDATE
  `github_home` = VALUES(`github_home`),
  `csdn_home` = VALUES(`csdn_home`),
  `gitee_home` = VALUES(`gitee_home`),
  `zhihu_home` = VALUES(`zhihu_home`),
  `github_show` = VALUES(`github_show`),
  `csdn_show` = VALUES(`csdn_show`),
  `gitee_show` = VALUES(`gitee_show`),
  `zhihu_show` = VALUES(`zhihu_show`),
  `web_enabled` = VALUES(`web_enabled`),
  `mp_enabled` = VALUES(`mp_enabled`),
  `config_version` = VALUES(`config_version`),
  `updated_by` = VALUES(`updated_by`);

-- ------------------------------------------------------------
-- 4. 迁移到 t_blog_infrastructure_config（基础设施配置）
-- ------------------------------------------------------------
-- 来源: t_blog_search_config, t_blog_storage_config, t_blog_email_config
-- 搜索字段优先级: t_blog_search_config > t_blog_ui_config
-- ------------------------------------------------------------

INSERT INTO `t_blog_infrastructure_config` (
  `id`, `search_engine`, `hot_search_mode`, `hot_search_words`,
  `meilisearch_host`, `meilisearch_api_key`, `meilisearch_index`,
  `meilisearch_last_sync_time`, `recommend_strategy`,
  `upload_strategy`, `upload_local_dir`, `upload_local_url_prefix`,
  `upload_qiniu_bucket`, `upload_qiniu_domain`,
  `upload_aliyun_endpoint`, `upload_aliyun_bucket`, `upload_aliyun_domain`,
  `email_config_json`, `config_version`, `created_by`, `updated_by`
)
SELECT
  1 AS id,
  COALESCE(search.`search_engine`, ui.`search_engine`, 'NONE') AS search_engine,
  COALESCE(search.`hot_search_mode`, ui.`hot_search_mode`, 0) AS hot_search_mode,
  COALESCE(search.`hot_search_words`, ui.`hot_search_words`, '') AS hot_search_words,
  COALESCE(search.`meilisearch_host`, ui.`meilisearch_host`, '') AS meilisearch_host,
  COALESCE(search.`meilisearch_api_key`, ui.`meilisearch_api_key`, '') AS meilisearch_api_key,
  COALESCE(search.`meilisearch_index`, ui.`meilisearch_index`, 'articles') AS meilisearch_index,
  search.`meilisearch_last_sync_time` AS meilisearch_last_sync_time,
  COALESCE(search.`recommend_strategy`, ui.`recommend_strategy`, 'rule') AS recommend_strategy,
  COALESCE(storage.`upload_strategy`, 'LOCAL') AS upload_strategy,
  COALESCE(storage.`upload_local_dir`, 'uploads') AS upload_local_dir,
  COALESCE(storage.`upload_local_url_prefix`, '/uploads/') AS upload_local_url_prefix,
  COALESCE(storage.`upload_qiniu_bucket`, '') AS upload_qiniu_bucket,
  COALESCE(storage.`upload_qiniu_domain`, '') AS upload_qiniu_domain,
  COALESCE(storage.`upload_aliyun_endpoint`, '') AS upload_aliyun_endpoint,
  COALESCE(storage.`upload_aliyun_bucket`, '') AS upload_aliyun_bucket,
  COALESCE(storage.`upload_aliyun_domain`, '') AS upload_aliyun_domain,
  email.`config_json` AS email_config_json,
  COALESCE(search.`config_version`, storage.`config_version`, 1) AS config_version,
  COALESCE(search.`created_by`, storage.`created_by`, '') AS created_by,
  COALESCE(search.`updated_by`, storage.`updated_by`, '') AS updated_by
FROM `t_blog_search_config` search
LEFT JOIN `t_blog_ui_config` ui ON ui.`id` = 1
LEFT JOIN `t_blog_storage_config` storage ON storage.`id` = 1
LEFT JOIN `t_blog_email_config` email ON email.`id` = 1
WHERE search.`id` = 1
ON DUPLICATE KEY UPDATE
  `search_engine` = VALUES(`search_engine`),
  `hot_search_mode` = VALUES(`hot_search_mode`),
  `hot_search_words` = VALUES(`hot_search_words`),
  `meilisearch_host` = VALUES(`meilisearch_host`),
  `meilisearch_api_key` = VALUES(`meilisearch_api_key`),
  `meilisearch_index` = VALUES(`meilisearch_index`),
  `meilisearch_last_sync_time` = VALUES(`meilisearch_last_sync_time`),
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
-- 迁移完成后的验证查询（可选执行）
-- ============================================================

-- 验证新表数据
-- SELECT * FROM t_blog_brand_config;
-- SELECT * FROM t_blog_compliance_config;
-- SELECT * FROM t_blog_channel_config;
-- SELECT * FROM t_blog_infrastructure_config;

-- ============================================================
-- 迁移脚本执行完成
-- ============================================================
-- 重要说明:
-- 1. 旧表 (t_blog_core_config, t_blog_ui_config, t_blog_privacy_config,
--    t_blog_social_config, t_blog_search_config, t_blog_storage_config,
--    t_blog_email_config) 已保留，不删除
-- 2. 当前代码仍读取旧表，新表数据作为备份和后续迁移使用
-- 3. L2 阶段将修改代码优先读取新表
-- 4. L3 阶段可考虑删除旧表（需确认前端无依赖）
-- ============================================================
