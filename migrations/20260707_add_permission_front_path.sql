-- ============================================================
-- 权限表增加前端字段迁移脚本
-- 日期: 2026-07-07
-- 说明: 为权限表增加前端路径和图标字段，支持动态菜单渲染
-- ============================================================

-- 增加前端路径字段
ALTER TABLE t_permission ADD COLUMN front_path VARCHAR(128) DEFAULT NULL COMMENT '前端菜单路径' AFTER description;

-- 增加图标字段
ALTER TABLE t_permission ADD COLUMN icon VARCHAR(64) DEFAULT NULL COMMENT '前端菜单图标' AFTER front_path;

-- 更新现有菜单类型权限的前端路径（基于已有权限数据）
UPDATE t_permission SET front_path = '/admin' WHERE perm_key = 'dashboard:read';
UPDATE t_permission SET front_path = '/admin/articles' WHERE perm_key = 'article:list';
UPDATE t_permission SET front_path = '/admin/categories' WHERE perm_key = 'category:list';
UPDATE t_permission SET front_path = '/admin/tags' WHERE perm_key = 'tag:list';
UPDATE t_permission SET front_path = '/admin/topics' WHERE perm_key = 'topic:list';
UPDATE t_permission SET front_path = '/admin/links' WHERE perm_key = 'link:list';
UPDATE t_permission SET front_path = '/admin/files' WHERE perm_key = 'file:list';
UPDATE t_permission SET front_path = '/admin/users' WHERE perm_key = 'user:list';
UPDATE t_permission SET front_path = '/admin/mp-users' WHERE perm_key = 'mpuser:list';
UPDATE t_permission SET front_path = '/admin/roles' WHERE perm_key = 'role:list';
UPDATE t_permission SET front_path = '/admin/permissions' WHERE perm_key = 'permission:list';
UPDATE t_permission SET front_path = '/admin/visitors' WHERE perm_key = 'dashboard:visitors';
UPDATE t_permission SET front_path = '/admin/errors' WHERE perm_key = 'dashboard:errors';
UPDATE t_permission SET front_path = '/admin/settings' WHERE perm_key = 'setting:read';

-- 验证
SELECT perm_key, name, type, front_path, icon FROM t_permission WHERE type = 'menu' ORDER BY sort_order;