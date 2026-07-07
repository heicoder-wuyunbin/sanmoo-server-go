-- ============================================================
-- 拆分管理后台设置配置表
-- 将混在一起的配置拆分为独立的表
-- ============================================================

-- ------------------------------------------------------------
-- 1. 社交链接配置表（从 t_blog_ui_config 拆分）
-- ------------------------------------------------------------
DROP TABLE IF EXISTS `t_blog_social_config`;
CREATE TABLE `t_blog_social_config` (
  `id` tinyint unsigned NOT NULL DEFAULT '1' COMMENT 'id，固定为1',
  `github_home` varchar(120) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'GitHub主页',
  `csdn_home` varchar(120) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'CSDN主页',
  `gitee_home` varchar(120) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'Gitee主页',
  `zhihu_home` varchar(120) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '知乎主页',
  `github_show` tinyint(1) NOT NULL DEFAULT '1' COMMENT '是否显示GitHub',
  `csdn_show` tinyint(1) NOT NULL DEFAULT '1' COMMENT '是否显示CSDN',
  `gitee_show` tinyint(1) NOT NULL DEFAULT '1' COMMENT '是否显示Gitee',
  `zhihu_show` tinyint(1) NOT NULL DEFAULT '1' COMMENT '是否显示知乎',
  `config_version` int unsigned NOT NULL DEFAULT '1' COMMENT '配置版本号',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `created_by` varchar(60) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '创建人',
  `updated_by` varchar(60) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '更新人',
  PRIMARY KEY (`id`),
  CONSTRAINT `chk_social_single_row` CHECK ((`id` = 1))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='博客社交链接配置表';

-- 迁移社交链接数据
INSERT INTO `t_blog_social_config` (
  `id`, `github_home`, `csdn_home`, `gitee_home`, `zhihu_home`,
  `github_show`, `csdn_show`, `gitee_show`, `zhihu_show`,
  `config_version`, `created_by`, `updated_by`
)
SELECT
  1, `github_home`, `csdn_home`, `gitee_home`, `zhihu_home`,
  `github_show`, `csdn_show`, `gitee_show`, `zhihu_show`,
  `config_version`, `created_by`, `updated_by`
FROM `t_blog_ui_config`
WHERE `id` = 1
ON DUPLICATE KEY UPDATE
  `github_home` = VALUES(`github_home`),
  `csdn_home` = VALUES(`csdn_home`),
  `gitee_home` = VALUES(`gitee_home`),
  `zhihu_home` = VALUES(`zhihu_home`),
  `github_show` = VALUES(`github_show`),
  `csdn_show` = VALUES(`csdn_show`),
  `gitee_show` = VALUES(`gitee_show`),
  `zhihu_show` = VALUES(`zhihu_show`),
  `updated_by` = VALUES(`updated_by`);

-- ------------------------------------------------------------
-- 2. 搜索配置表（从 t_blog_ui_config 拆分）
-- ------------------------------------------------------------
DROP TABLE IF EXISTS `t_blog_search_config`;
CREATE TABLE `t_blog_search_config` (
  `id` tinyint unsigned NOT NULL DEFAULT '1' COMMENT 'id，固定为1',
  `recommend_strategy` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'rule' COMMENT '推荐策略：rule/weighted/cf',
  `search_engine` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'NONE' COMMENT '搜索引擎: NONE/MEILISEARCH',
  `hot_search_mode` tinyint(1) NOT NULL DEFAULT '0' COMMENT '热门搜索模式：0=伪热门(FAKE) 1=真热门(REAL)',
  `hot_search_words` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '热门搜索词JSON数组',
  `meilisearch_host` varchar(200) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'MeiliSearch主机地址',
  `meilisearch_api_key` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'MeiliSearch API Key',
  `meilisearch_index` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'articles' COMMENT 'MeiliSearch索引名称',
  `config_version` int unsigned NOT NULL DEFAULT '1' COMMENT '配置版本号',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `created_by` varchar(60) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '创建人',
  `updated_by` varchar(60) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '更新人',
  PRIMARY KEY (`id`),
  CONSTRAINT `chk_search_single_row` CHECK ((`id` = 1))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='博客搜索与推荐配置表';

-- 迁移搜索配置数据
INSERT INTO `t_blog_search_config` (
  `id`, `recommend_strategy`, `search_engine`, `hot_search_mode`, `hot_search_words`,
  `meilisearch_host`, `meilisearch_api_key`, `meilisearch_index`,
  `config_version`, `created_by`, `updated_by`
)
SELECT
  1, `recommend_strategy`, `search_engine`, `hot_search_mode`, `hot_search_words`,
  `meilisearch_host`, `meilisearch_api_key`, `meilisearch_index`,
  `config_version`, `created_by`, `updated_by`
FROM `t_blog_ui_config`
WHERE `id` = 1
ON DUPLICATE KEY UPDATE
  `recommend_strategy` = VALUES(`recommend_strategy`),
  `search_engine` = VALUES(`search_engine`),
  `hot_search_mode` = VALUES(`hot_search_mode`),
  `hot_search_words` = VALUES(`hot_search_words`),
  `meilisearch_host` = VALUES(`meilisearch_host`),
  `meilisearch_api_key` = VALUES(`meilisearch_api_key`),
  `meilisearch_index` = VALUES(`meilisearch_index`),
  `updated_by` = VALUES(`updated_by`);

-- ------------------------------------------------------------
-- 3. 隐私政策配置表（从 t_blog_core_config 拆分）
-- ------------------------------------------------------------
DROP TABLE IF EXISTS `t_blog_privacy_config`;
CREATE TABLE `t_blog_privacy_config` (
  `id` tinyint unsigned NOT NULL DEFAULT '1' COMMENT 'id，固定为1',
  `privacy_policy` text COLLATE utf8mb4_unicode_ci COMMENT '隐私政策内容',
  `config_version` int unsigned NOT NULL DEFAULT '1' COMMENT '配置版本号',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `created_by` varchar(60) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '创建人',
  `updated_by` varchar(60) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '更新人',
  PRIMARY KEY (`id`),
  CONSTRAINT `chk_privacy_single_row` CHECK ((`id` = 1))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='博客隐私政策配置表';

-- 迁移隐私政策数据
INSERT INTO `t_blog_privacy_config` (
  `id`, `privacy_policy`, `config_version`, `created_by`, `updated_by`
)
SELECT
  1, `privacy_policy`, `config_version`, `created_by`, `updated_by`
FROM `t_blog_core_config`
WHERE `id` = 1
ON DUPLICATE KEY UPDATE
  `privacy_policy` = VALUES(`privacy_policy`),
  `updated_by` = VALUES(`updated_by`);

-- ------------------------------------------------------------
-- 4. 新增设置相关权限
-- ------------------------------------------------------------
INSERT INTO t_permission (perm_key, name, module, type, description, front_path, icon, sort_order, status, create_time, update_time) VALUES
('setting:core:read',    '查看核心配置',     'setting', 'menu',   '查看核心配置',      '/admin/settings/core',         'SettingOutlined', 81, 1, NOW(), NOW()),
('setting:core:update',  '编辑核心配置',     'setting', 'button', '修改核心配置',      '',                             '',                82, 1, NOW(), NOW()),
('setting:privacy:read', '查看隐私政策',     'setting', 'menu',   '查看隐私政策',      '/admin/settings/privacy',      'FileTextOutlined', 83, 1, NOW(), NOW()),
('setting:privacy:update','编辑隐私政策',    'setting', 'button', '修改隐私政策',      '',                             '',                84, 1, NOW(), NOW()),
('setting:social:read',  '查看社交链接',     'setting', 'menu',   '查看社交链接配置',  '/admin/settings/social',       'GlobalOutlined',   85, 1, NOW(), NOW()),
('setting:social:update','编辑社交链接',     'setting', 'button', '修改社交链接配置',  '',                             '',                86, 1, NOW(), NOW()),
('setting:search:read',  '查看搜索配置',     'setting', 'menu',   '查看搜索配置',      '/admin/settings/search',       'SearchOutlined',   87, 1, NOW(), NOW()),
('setting:search:update','编辑搜索配置',     'setting', 'button', '修改搜索配置',      '',                             '',                88, 1, NOW(), NOW()),
('setting:storage:read', '查看存储配置',     'setting', 'menu',   '查看存储配置',      '/admin/settings/storage',      'CloudServerOutlined', 89, 1, NOW(), NOW()),
('setting:storage:update','编辑存储配置',    'setting', 'button', '修改存储配置',      '',                             '',                90, 1, NOW(), NOW()),
('setting:email:read',   '查看邮件配置',     'setting', 'menu',   '查看邮件配置',      '/admin/settings/email',        'MailOutlined',     91, 1, NOW(), NOW()),
('setting:cache:read',   '查看缓存管理',     'setting', 'menu',   '查看缓存管理',      '/admin/settings/cache',        'ThunderboltOutlined', 92, 1, NOW(), NOW()),
('setting:maintenance:read', '查看数据维护', 'setting', 'menu',  '查看数据维护',      '/admin/settings/maintenance',  'DatabaseOutlined', 93, 1, NOW(), NOW());

-- 给 admin 角色赋予所有新权限（admin 角色自动拥有所有权限，这里仅做记录）
-- 注意：根据项目约定，name='admin' 的角色自动拥有所有权限，无需显式分配
