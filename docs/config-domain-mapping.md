# Sanmoo Blog 配置域映射文档

## 1. 配置域重构目标

将现有混杂的配置域整理为四类：

1. **站点品牌配置**：博客名、作者、介绍、头像、站点地址、SEO基础项
2. **合规配置**：隐私政策、数据保留说明、账号注销说明、备案信息展示
3. **渠道配置**：Web渠道配置、小程序渠道配置、社交链接展示开关
4. **基础设施配置**：搜索、存储、邮件、缓存等技术配置

---

## 2. 现有配置表字段归类

### 2.1 当前配置表

| 表名 | 用途 | 问题 |
|------|------|------|
| `t_blog_core_config` | 核心配置 | 包含隐私政策，边界不清 |
| `t_blog_ui_config` | UI展示配置 | 混入社交、搜索、推荐等跨域字段 |
| `t_blog_privacy_config` | 隐私配置 | 与 core_config 重复 |
| `t_blog_social_config` | 社交配置 | 归属渠道配置 |
| `t_blog_search_config` | 搜索配置 | 归属基础设施配置 |
| `t_blog_storage_config` | 存储配置 | 归属基础设施配置 |
| `t_blog_email_config` | 邮件配置 | 归属基础设施配置 |

### 2.2 字段归类表

#### 站点品牌配置

| 字段名 | 当前表 | 说明 | 迁移目标 |
|--------|--------|------|----------|
| `blog_name` | `t_blog_core_config` | 博客名称 | 站点品牌 |
| `author` | `t_blog_core_config` | 作者名 | 站点品牌 |
| `introduction` | `t_blog_core_config` | 介绍语 | 站点品牌 |
| `avatar` | `t_blog_core_config` | 作者头像 | 站点品牌 |
| `rss_enabled` | `t_blog_core_config` | RSS开关 | 站点品牌 |
| `config_version` | `t_blog_core_config` | 配置版本号 | 通用 |

#### 合规配置

| 字段名 | 当前表 | 说明 | 迁移目标 |
|--------|--------|------|----------|
| `privacy_policy` | `t_blog_core_config` | 隐私政策内容 | 合规配置 |
| `privacy_policy` | `t_blog_privacy_config` | 隐私政策内容（重复） | 合规配置 |
| `filing_info` | 新增 | 备案信息 | 合规配置 |
| `contact_info` | 新增 | 联系方式 | 合规配置 |
| `data_retention_policy` | 新增 | 数据保留说明 | 合规配置 |
| `account_deletion_guide` | 新增 | 账号注销说明 | 合规配置 |

#### 渠道配置

| 字段名 | 当前表 | 说明 | 迁移目标 |
|--------|--------|------|----------|
| `github_home` | `t_blog_social_config` | GitHub主页 | 渠道配置 |
| `csdn_home` | `t_blog_social_config` | CSDN主页 | 渠道配置 |
| `gitee_home` | `t_blog_social_config` | Gitee主页 | 渠道配置 |
| `zhihu_home` | `t_blog_social_config` | 知乎主页 | 渠道配置 |
| `github_show` | `t_blog_social_config` | 显示GitHub | 渠道配置 |
| `csdn_show` | `t_blog_social_config` | 显示CSDN | 渠道配置 |
| `gitee_show` | `t_blog_social_config` | 显示Gitee | 渠道配置 |
| `zhihu_show` | `t_blog_social_config` | 显示知乎 | 渠道配置 |
| `github_home` | `t_blog_ui_config` | GitHub主页（重复） | 渠道配置 |
| `csdn_home` | `t_blog_ui_config` | CSDN主页（重复） | 渠道配置 |
| `gitee_home` | `t_blog_ui_config` | Gitee主页（重复） | 渠道配置 |
| `zhihu_home` | `t_blog_ui_config` | 知乎主页（重复） | 渠道配置 |
| `github_show` | `t_blog_ui_config` | 显示GitHub（重复） | 渠道配置 |
| `csdn_show` | `t_blog_ui_config` | 显示CSDN（重复） | 渠道配置 |
| `gitee_show` | `t_blog_ui_config` | 显示Gitee（重复） | 渠道配置 |
| `zhihu_show` | `t_blog_ui_config` | 显示知乎（重复） | 渠道配置 |

#### 基础设施配置

| 字段名 | 当前表 | 说明 | 迁移目标 |
|--------|--------|------|----------|
| `recommend_strategy` | `t_blog_search_config` | 推荐策略 | 基础设施 |
| `search_engine` | `t_blog_search_config` | 搜索引擎 | 基础设施 |
| `hot_search_mode` | `t_blog_search_config` | 热搜索模式 | 基础设施 |
| `hot_search_words` | `t_blog_search_config` | 热搜索词 | 基础设施 |
| `meilisearch_host` | `t_blog_search_config` | MeiliSearch主机 | 基础设施 |
| `meilisearch_api_key` | `t_blog_search_config` | MeiliSearch API Key | 基础设施 |
| `meilisearch_index` | `t_blog_search_config` | MeiliSearch索引 | 基础设施 |
| `meilisearch_last_sync_time` | `t_blog_search_config` | 最后同步时间 | 基础设施 |
| `upload_strategy` | `t_blog_storage_config` | 上传策略 | 基础设施 |
| `upload_local_dir` | `t_blog_storage_config` | 本地上传目录 | 基础设施 |
| `upload_local_url_prefix` | `t_blog_storage_config` | 本地访问前缀 | 基础设施 |
| `upload_qiniu_bucket` | `t_blog_storage_config` | 七牛Bucket | 基础设施 |
| `upload_qiniu_domain` | `t_blog_storage_config` | 七牛域名 | 基础设施 |
| `upload_aliyun_endpoint` | `t_blog_storage_config` | 阿里云Endpoint | 基础设施 |
| `upload_aliyun_bucket` | `t_blog_storage_config` | 阿里云Bucket | 基础设施 |
| `upload_aliyun_domain` | `t_blog_storage_config` | 阿里云域名 | 基础设施 |
| `config_json` | `t_blog_email_config` | 邮件配置JSON | 基础设施 |
| `recommend_strategy` | `t_blog_ui_config` | 推荐策略（重复） | 基础设施 |
| `search_engine` | `t_blog_ui_config` | 搜索引擎（重复） | 基础设施 |
| `hot_search_words` | `t_blog_ui_config` | 热搜索词（重复） | 基础设施 |
| `meilisearch_host` | `t_blog_ui_config` | MeiliSearch主机（重复） | 基础设施 |
| `meilisearch_api_key` | `t_blog_ui_config` | MeiliSearch API Key（重复） | 基础设施 |
| `meilisearch_index` | `t_blog_ui_config` | MeiliSearch索引（重复） | 基础设施 |
| `hot_search_mode` | `t_blog_ui_config` | 热搜索模式（重复） | 基础设施 |

---

## 3. 重构后的配置域模型

### 3.1 站点品牌配置表

**表名**: `t_blog_brand_config`

| 字段名 | 类型 | 说明 |
|--------|------|------|
| `id` | tinyint unsigned | 主键，固定为1 |
| `blog_name` | varchar(60) | 博客名称 |
| `author` | varchar(60) | 作者名 |
| `introduction` | varchar(200) | 介绍语 |
| `avatar` | varchar(160) | 作者头像 |
| `site_url` | varchar(200) | 站点地址 |
| `rss_enabled` | tinyint(1) | RSS开关 |
| `config_version` | int unsigned | 配置版本号 |
| `create_time` | datetime | 创建时间 |
| `update_time` | datetime | 更新时间 |
| `created_by` | varchar(60) | 创建人 |
| `updated_by` | varchar(60) | 更新人 |

### 3.2 合规配置表

**表名**: `t_blog_compliance_config`

| 字段名 | 类型 | 说明 |
|--------|------|------|
| `id` | tinyint unsigned | 主键，固定为1 |
| `privacy_policy` | text | 隐私政策内容 |
| `filing_info` | varchar(500) | 备案信息（JSON格式） |
| `contact_info` | varchar(500) | 联系方式（JSON格式） |
| `data_retention_policy` | text | 数据保留说明 |
| `account_deletion_guide` | text | 账号注销说明 |
| `config_version` | int unsigned | 配置版本号 |
| `create_time` | datetime | 创建时间 |
| `update_time` | datetime | 更新时间 |
| `created_by` | varchar(60) | 创建人 |
| `updated_by` | varchar(60) | 更新人 |

### 3.3 渠道配置表

**表名**: `t_blog_channel_config`

| 字段名 | 类型 | 说明 |
|--------|------|------|
| `id` | tinyint unsigned | 主键，固定为1 |
| `github_home` | varchar(120) | GitHub主页 |
| `csdn_home` | varchar(120) | CSDN主页 |
| `gitee_home` | varchar(120) | Gitee主页 |
| `zhihu_home` | varchar(120) | 知乎主页 |
| `github_show` | tinyint(1) | 是否显示GitHub |
| `csdn_show` | tinyint(1) | 是否显示CSDN |
| `gitee_show` | tinyint(1) | 是否显示Gitee |
| `zhihu_show` | tinyint(1) | 是否显示知乎 |
| `web_enabled` | tinyint(1) | Web渠道启用 |
| `mp_enabled` | tinyint(1) | 小程序渠道启用 |
| `config_version` | int unsigned | 配置版本号 |
| `create_time` | datetime | 创建时间 |
| `update_time` | datetime | 更新时间 |
| `created_by` | varchar(60) | 创建人 |
| `updated_by` | varchar(60) | 更新人 |

### 3.4 基础设施配置表

**表名**: `t_blog_infrastructure_config`

| 字段名 | 类型 | 说明 |
|--------|------|------|
| `id` | tinyint unsigned | 主键，固定为1 |
| `search_engine` | varchar(20) | 搜索引擎 |
| `hot_search_mode` | tinyint(1) | 热搜索模式 |
| `hot_search_words` | text | 热搜索词JSON |
| `meilisearch_host` | varchar(200) | MeiliSearch主机 |
| `meilisearch_api_key` | varchar(100) | MeiliSearch API Key |
| `meilisearch_index` | varchar(100) | MeiliSearch索引 |
| `meilisearch_last_sync_time` | datetime | 最后同步时间 |
| `recommend_strategy` | varchar(20) | 推荐策略 |
| `upload_strategy` | varchar(20) | 上传策略 |
| `upload_local_dir` | varchar(200) | 本地上传目录 |
| `upload_local_url_prefix` | varchar(80) | 本地访问前缀 |
| `upload_qiniu_bucket` | varchar(120) | 七牛Bucket |
| `upload_qiniu_domain` | varchar(200) | 七牛域名 |
| `upload_aliyun_endpoint` | varchar(200) | 阿里云Endpoint |
| `upload_aliyun_bucket` | varchar(120) | 阿里云Bucket |
| `upload_aliyun_domain` | varchar(200) | 阿里云域名 |
| `email_config_json` | longtext | 邮件配置JSON |
| `config_version` | int unsigned | 配置版本号 |
| `create_time` | datetime | 创建时间 |
| `update_time` | datetime | 更新时间 |
| `created_by` | varchar(60) | 创建人 |
| `updated_by` | varchar(60) | 更新人 |

---

## 4. 兼容改造方案

### 4.1 迁移步骤

1. **创建新表**：创建 `t_blog_brand_config`、`t_blog_compliance_config`、`t_blog_channel_config`、`t_blog_infrastructure_config`
2. **数据迁移**：从现有表迁移数据到新表
3. **代码适配**：修改 Service 和 Repository 层支持新旧表读写
4. **接口兼容**：对外接口保持不变，内部使用新表
5. **逐步淘汰**：后续版本移除旧表

### 4.2 接口兼容策略

| 阶段 | 说明 |
|------|------|
| 阶段1（L1） | 创建新表并迁移数据，代码同时支持新旧表 |
| 阶段2（L2） | 代码优先使用新表，旧表作为备份 |
| 阶段3（L3） | 移除旧表相关代码 |

### 4.3 SQL变更方案

详细SQL变更方案见 `migrations/` 目录下的迁移脚本。

---

## 5. 文档版本

| 版本 | 日期 | 说明 |
|------|------|------|
| v1.0 | 2026-07-08 | 初始版本，基于 plan.md 设计 |
