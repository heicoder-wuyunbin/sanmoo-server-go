-- 数据维护模块权限初始化
INSERT INTO t_permission (perm_key, name, module, type, description, sort_order, status, create_time, update_time) VALUES
('maintenance:stats',    '维护统计',    'maintenance', 'button', '查看日志表统计',       140, 1, NOW(), NOW()),
('maintenance:cleanup',  '清理日志',    'maintenance', 'button', '清理过期日志数据',     141, 1, NOW(), NOW());

-- admin 角色自动拥有全部权限（无需手动关联）
-- editor 角色（可查看统计，但不能清理）
INSERT IGNORE INTO t_role_permission (role_id, perm_key, create_time) VALUES
(2, 'maintenance:stats', NOW());

-- viewer 角色（只读）
INSERT IGNORE INTO t_role_permission (role_id, perm_key, create_time) VALUES
(3, 'maintenance:stats', NOW());