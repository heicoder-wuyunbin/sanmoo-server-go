# Sanmoo Blog 后端接口域清单

## 1. 接口域分类总览

| 域 | 路径前缀 | 认证要求 | 限流策略 | 说明 |
|----|----------|----------|----------|------|
| 健康检查 | `/health`, `/sitemap.xml`, `/rss.xml` | 无需认证 | 无 | 公开探活与静态资源 |
| 认证 | `/auth` | 无需认证 | 5次/分钟 | 登录、刷新、验证码 |
| 用户 | `/user` | JWT | 无 | 密码修改 |
| 管理端 | `/admin` | JWT | 300次/分钟 | 后台管理全部接口 |
| 门户端 | `/web` | 无需认证 | 60次/分钟 | 公开访问接口 |
| 小程序端 | `/mp` | 无需认证 | 60次/分钟 | 小程序专用接口 |

---

## 2. 门户端接口（/web）

### 2.1 内容浏览（保留）

| 接口 | 方法 | 说明 | 策略 |
|------|------|------|------|
| `/web/settings` | GET | 获取站点配置 | 保留 |
| `/web/categories` | GET | 获取分类列表 | 保留 |
| `/web/tags` | GET | 获取标签列表 | 保留 |
| `/web/articles` | GET | 获取文章列表 | 保留 |
| `/web/articles/:id` | GET | 获取文章详情 | 保留 |
| `/web/articles/slug/:slug` | GET | 通过slug获取文章 | 保留 |
| `/web/categories/:id/articles` | GET | 获取分类下文章 | 保留 |
| `/web/tags/:id/articles` | GET | 获取标签下文章 | 保留 |
| `/web/archives` | GET | 获取归档列表 | 保留 |
| `/web/topics` | GET | 获取专题列表 | 保留 |
| `/web/topics/:id` | GET | 获取专题详情 | 保留 |
| `/web/topics/:id/articles` | GET | 获取专题文章 | 保留 |
| `/web/search/hot` | GET | 获取热门搜索 | 保留 |
| `/web/search` | GET | 全文搜索 | 保留 |
| `/web/articles/:id/related` | GET | 获取相关文章 | 保留 |
| `/web/articles/hot` | GET | 获取热门文章 | 保留 |
| `/web/articles/random` | GET | 获取随机文章 | 保留 |
| `/web/links` | GET | 获取友情链接 | 保留 |

### 2.2 互动（保留）

| 接口 | 方法 | 说明 | 策略 |
|------|------|------|------|
| `/web/articles/:id/like` | POST | 文章点赞 | 保留 |

---

## 3. 小程序端接口（/mp）

### 3.1 核心阅读能力（保留）

| 接口 | 方法 | 说明 | 策略 |
|------|------|------|------|
| `/mp/auth/session` | POST | 微信登录 | 保留 |
| `/mp/settings` | GET | 获取小程序配置 | 保留 |
| `/mp/categories` | GET | 获取分类列表 | 保留 |
| `/mp/tags` | GET | 获取标签列表 | 保留 |
| `/mp/topics` | GET | 获取专题列表 | 保留 |
| `/mp/topics/:id` | GET | 获取专题详情 | 保留 |
| `/mp/topics/:id/articles` | GET | 获取专题文章 | 保留 |
| `/mp/articles` | GET | 获取文章列表 | 保留 |
| `/mp/articles/:id` | GET | 获取文章详情 | 保留 |
| `/mp/categories/:id/articles` | GET | 获取分类文章 | 保留 |
| `/mp/tags/:id/articles` | GET | 获取标签文章 | 保留 |
| `/mp/archives` | GET | 获取归档列表 | 保留 |
| `/mp/search/hot` | GET | 获取热门搜索 | 保留 |

### 3.2 用户基础能力（保留）

| 接口 | 方法 | 说明 | 策略 |
|------|------|------|------|
| `/mp/user/profile` | GET | 获取用户信息 | 保留 |
| `/mp/user/profile` | POST | 更新用户信息 | 保留 |
| `/mp/favorites/:articleId` | POST | 添加收藏 | 保留 |
| `/mp/favorites/:articleId` | DELETE | 取消收藏 | 保留 |
| `/mp/favorites/:articleId/status` | GET | 收藏状态 | 保留 |
| `/mp/favorites` | GET | 收藏列表 | 保留 |
| `/mp/history/:articleId` | POST | 添加浏览历史 | 保留 |
| `/mp/history` | DELETE | 清空浏览历史 | 保留 |
| `/mp/history` | GET | 浏览历史列表 | 保留 |
| `/mp/subscribe` | POST | 订阅/取消订阅 | 保留 |
| `/mp/subscribe/status` | GET | 订阅状态 | 保留 |
| `/mp/user` | DELETE | 删除用户数据 | 保留 |
| `/mp/privacy-policy` | GET | 获取隐私政策 | 保留 |

### 3.3 推荐与行为（弱化）

| 接口 | 方法 | 说明 | 策略 |
|------|------|------|------|
| `/mp/recommendations` | GET | 获取推荐文章 | 弱化（仅保留规则推荐） |
| `/mp/behavior` | POST | 上报行为日志 | 弱化（仅用于内容优化） |

---

## 4. 管理端接口（/admin）

### 4.1 内容管理（核心保留）

| 接口 | 方法 | 说明 | 策略 |
|------|------|------|------|
| `/admin/articles` | GET | 获取文章列表 | 保留 |
| `/admin/articles/export` | GET | 导出文章 | 保留 |
| `/admin/articles/:id` | GET | 文章详情 | 保留 |
| `/admin/articles` | POST | 创建文章 | 保留 |
| `/admin/articles/:id` | PUT | 更新文章 | 保留 |
| `/admin/articles/:id/status` | PUT | 更新文章状态 | 保留 |
| `/admin/articles/batch-status` | PUT | 批量更新状态 | 保留 |
| `/admin/articles/:id` | DELETE | 删除文章 | 保留 |
| `/admin/articles/batch-delete` | DELETE | 批量删除文章 | 保留 |
| `/admin/tags` | GET | 获取标签列表 | 保留 |
| `/admin/tags` | POST | 创建标签 | 保留 |
| `/admin/tags/:id` | PUT | 更新标签 | 保留 |
| `/admin/tags/:id` | DELETE | 删除标签 | 保留 |
| `/admin/tags/batch-delete` | DELETE | 批量删除标签 | 保留 |
| `/admin/categories` | GET | 获取分类列表 | 保留 |
| `/admin/categories` | POST | 创建分类 | 保留 |
| `/admin/categories/:id` | PUT | 更新分类 | 保留 |
| `/admin/categories/:id` | DELETE | 删除分类 | 保留 |
| `/admin/categories/batch-delete` | DELETE | 批量删除分类 | 保留 |
| `/admin/topics` | GET | 获取专题列表 | 保留 |
| `/admin/topics` | POST | 创建专题 | 保留 |
| `/admin/topics/:id` | PUT | 更新专题 | 保留 |
| `/admin/topics/:id` | DELETE | 删除专题 | 保留 |
| `/admin/topics/batch-delete` | DELETE | 批量删除专题 | 保留 |
| `/admin/topics/:id/articles` | GET | 专题文章 | 保留 |
| `/admin/articles/published-options` | GET | 已发布文章选项 | 保留 |
| `/admin/links` | GET | 获取友情链接 | 保留 |
| `/admin/links` | POST | 创建友情链接 | 保留 |
| `/admin/links/:id` | PUT | 更新友情链接 | 保留 |
| `/admin/links/:id` | DELETE | 删除友情链接 | 保留 |
| `/admin/links/batch-delete` | DELETE | 批量删除友情链接 | 保留 |

### 4.2 文件管理（保留）

| 接口 | 方法 | 说明 | 策略 |
|------|------|------|------|
| `/admin/files` | GET | 获取文件列表 | 保留 |
| `/admin/files/upload` | POST | 上传文件 | 保留 |
| `/admin/files/:id` | DELETE | 删除文件 | 保留 |

### 4.3 配置管理（保留并重构）

| 接口 | 方法 | 说明 | 策略 |
|------|------|------|------|
| `/admin/settings` | GET | 获取全部设置 | 保留（待重构） |
| `/admin/settings` | PUT | 更新全部设置 | 保留（待重构） |
| `/admin/settings/core` | GET | 获取核心配置 | 保留 |
| `/admin/settings/core` | PUT | 更新核心配置 | 保留 |
| `/admin/settings/privacy` | GET | 获取隐私配置 | 保留 |
| `/admin/settings/privacy` | PUT | 更新隐私配置 | 保留 |
| `/admin/settings/social` | GET | 获取社交配置 | 保留 |
| `/admin/settings/social` | PUT | 更新社交配置 | 保留 |
| `/admin/settings/search` | GET | 获取搜索配置 | 保留 |
| `/admin/settings/search` | PUT | 更新搜索配置 | 保留 |
| `/admin/settings/storage` | GET | 获取存储配置 | 保留 |
| `/admin/settings/storage` | PUT | 更新存储配置 | 保留 |
| `/admin/settings/email` | GET | 获取邮件配置 | 保留 |
| `/admin/settings/email` | PUT | 更新邮件配置 | 保留 |
| `/admin/settings/email/send-code` | POST | 发送验证邮件 | 保留 |
| `/admin/settings/email/verify-code` | POST | 验证邮件验证码 | 保留 |
| `/admin/search/sync` | POST | 同步MeiliSearch | 保留 |
| `/admin/search/stats` | GET | MeiliSearch统计 | 保留 |
| `/admin/settings/export` | POST | 导出配置 | 冻结 |
| `/admin/settings/import` | POST | 导入配置 | 冻结 |

### 4.4 数据看板（保留并聚焦）

| 接口 | 方法 | 说明 | 策略 |
|------|------|------|------|
| `/admin/dashboard` | GET | 仪表盘总览 | 保留 |
| `/admin/dashboard/pv` | GET | 页面浏览量 | 保留 |
| `/admin/dashboard/tag-statistics` | GET | 标签统计 | 保留 |
| `/admin/dashboard/category-statistics` | GET | 分类统计 | 保留 |
| `/admin/dashboard/heatmap` | GET | 发布热力图 | 保留 |
| `/admin/dashboard/topic-statistics` | GET | 专题统计 | 保留 |
| `/admin/dashboard/article-read-statistics` | GET | 文章阅读统计 | 保留 |
| `/admin/dashboard/category-read-statistics` | GET | 分类阅读统计 | 保留 |
| `/admin/dashboard/tag-read-statistics` | GET | 标签阅读统计 | 保留 |
| `/admin/dashboard/content-trend` | GET | 内容趋势 | 保留 |
| `/admin/dashboard/visitors` | GET | 访客记录 | 弱化 |
| `/admin/dashboard/visitors/:id` | DELETE | 删除访客记录 | 冻结 |
| `/admin/dashboard/visitors/batch-delete` | DELETE | 批量删除访客 | 冻结 |
| `/admin/dashboard/visitors/clear` | DELETE | 清空访客记录 | 冻结 |
| `/admin/dashboard/errors` | GET | 错误日志 | 弱化 |
| `/admin/dashboard/errors/:id` | DELETE | 删除错误日志 | 冻结 |
| `/admin/dashboard/errors/batch-delete` | DELETE | 批量删除错误 | 冻结 |
| `/admin/dashboard/errors/clear` | DELETE | 清空错误日志 | 冻结 |
| `/admin/dashboard/errors/import` | POST | 导入错误日志 | 冻结 |
| `/admin/dashboard/errors/export` | GET | 导出错误日志 | 冻结 |
| `/admin/dashboard/mp-user-growth` | GET | 小程序用户增长 | 弱化 |

### 4.5 用户管理（冻结）

| 接口 | 方法 | 说明 | 策略 |
|------|------|------|------|
| `/admin/users` | GET | 获取用户列表 | 冻结 |
| `/admin/users/export` | GET | 导出用户 | 冻结 |
| `/admin/users` | POST | 创建用户 | 冻结 |
| `/admin/users/:id` | PUT | 更新用户 | 冻结 |
| `/admin/users/:id` | DELETE | 删除用户 | 冻结 |
| `/admin/users/:id/password` | PUT | 重置密码 | 冻结 |
| `/admin/users/:id/status` | PUT | 切换用户状态 | 冻结 |
| `/admin/users/batch-delete` | DELETE | 批量删除用户 | 冻结 |
| `/admin/users/:id/roles` | PUT | 分配角色 | 冻结 |

### 4.6 权限管理（冻结）

| 接口 | 方法 | 说明 | 策略 |
|------|------|------|------|
| `/admin/permissions` | GET | 获取权限列表 | 冻结 |
| `/admin/permissions/tree` | GET | 获取权限树 | 冻结 |
| `/admin/permissions/:id` | GET | 获取权限详情 | 冻结 |
| `/admin/permissions` | POST | 创建权限 | 冻结 |
| `/admin/permissions/:id` | PUT | 更新权限 | 冻结 |
| `/admin/permissions/:id` | DELETE | 删除权限 | 冻结 |

### 4.7 角色管理（冻结）

| 接口 | 方法 | 说明 | 策略 |
|------|------|------|------|
| `/admin/roles` | GET | 获取角色列表 | 冻结 |
| `/admin/roles/all` | GET | 获取所有角色 | 冻结 |
| `/admin/roles/:id` | GET | 获取角色详情 | 冻结 |
| `/admin/roles` | POST | 创建角色 | 冻结 |
| `/admin/roles/:id` | PUT | 更新角色 | 冻结 |
| `/admin/roles/:id` | DELETE | 删除角色 | 冻结 |
| `/admin/roles/:id/permissions` | PUT | 分配权限 | 冻结 |
| `/admin/roles/:id/permissions` | GET | 获取角色权限 | 冻结 |

### 4.8 当前用户（保留）

| 接口 | 方法 | 说明 | 策略 |
|------|------|------|------|
| `/admin/user/permissions` | GET | 当前用户权限 | 保留 |
| `/admin/user/menus` | GET | 当前用户菜单 | 保留 |

### 4.9 缓存管理（冻结）

| 接口 | 方法 | 说明 | 策略 |
|------|------|------|------|
| `/admin/cache/clear` | POST | 清理缓存 | 冻结 |
| `/admin/cache/warmup` | POST | 预热缓存 | 冻结 |
| `/admin/cache/stats` | GET | 缓存统计 | 冻结 |

### 4.10 备份管理（冻结）

| 接口 | 方法 | 说明 | 策略 |
|------|------|------|------|
| `/admin/backup/export` | POST | 导出备份 | 冻结 |
| `/admin/backup/list` | GET | 备份列表 | 冻结 |
| `/admin/backup/download/:fileName` | GET | 下载备份 | 冻结 |
| `/admin/backup/:fileName` | DELETE | 删除备份 | 冻结 |
| `/admin/backup/stats` | GET | 备份统计 | 冻结 |

### 4.11 小程序用户管理（冻结）

| 接口 | 方法 | 说明 | 策略 |
|------|------|------|------|
| `/admin/mp-users` | GET | 小程序用户列表 | 冻结 |
| `/admin/mp-users/:openid` | GET | 小程序用户详情 | 冻结 |
| `/admin/mp-users/:openid/profile` | GET | 用户画像 | 冻结 |
| `/admin/mp-users/:openid/profile` | POST | 生成画像 | 冻结 |
| `/admin/mp-users/:openid/tags/generate` | POST | 生成标签 | 冻结 |
| `/admin/mp-users/:openid/tags` | GET | 用户标签 | 冻结 |
| `/admin/mp-users/:openid/tags/:tagId` | DELETE | 删除标签 | 冻结 |
| `/admin/mp-users/:openid/radar/refresh` | POST | 刷新雷达图 | 冻结 |

### 4.12 数据维护（冻结）

| 接口 | 方法 | 说明 | 策略 |
|------|------|------|------|
| `/admin/maintenance/stats` | GET | 维护统计 | 冻结 |
| `/admin/maintenance/cleanup-logs` | POST | 清理日志 | 冻结 |

---

## 5. 认证接口（/auth）

| 接口 | 方法 | 说明 | 策略 |
|------|------|------|------|
| `/auth/login` | POST | 用户登录 | 保留 |
| `/auth/refresh` | POST | 刷新Token | 保留 |
| `/auth/send-verification-code` | POST | 发送验证码 | 保留 |
| `/auth/verify-verification-code` | POST | 验证验证码 | 保留 |

---

## 6. 用户接口（/user）

| 接口 | 方法 | 说明 | 策略 |
|------|------|------|------|
| `/user/password` | PUT | 修改密码 | 保留 |

---

## 7. 策略说明

### 7.1 保留

接口持续维护，功能稳定，符合业务定位。

### 7.2 弱化

接口保持可用，但不再主动扩展功能，前端可选择隐藏相关入口。

### 7.3 冻结

接口仍存在，但标记为不建议使用，前端应隐藏相关入口，后续可考虑移除。

---

## 8. 文档版本

| 版本 | 日期 | 说明 |
|------|------|------|
| v1.0 | 2026-07-08 | 初始版本，基于 router.go 梳理 |
