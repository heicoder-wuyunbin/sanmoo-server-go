-- ============================================================
-- 数据导出权限迁移脚本
-- 日期: 2026-07-07
-- 说明: 为文章和用户模块增加导出权限
-- ============================================================

INSERT INTO t_permission (perm_key, name, module, type, description, sort_order, status, create_time, update_time) VALUES
('article:export', '导出文章', 'article', 'button', '导出文章列表为CSV', 16, 1, NOW(), NOW()),
('user:export',    '导出用户', 'user',    'button', '导出用户列表为CSV', 77, 1, NOW(), NOW());
