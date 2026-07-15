-- ============================================================
-- Sanmoo Blog - 微信配置权限
-- ============================================================
-- 日期: 2026-07-15
-- 说明: 插入微信配置相关的权限记录，供后台菜单显示和接口鉴权使用
-- ============================================================

SET NAMES utf8mb4 COLLATE utf8mb4_unicode_ci;

-- 插入微信配置查看权限（菜单项）
INSERT INTO `t_permission` (`perm_key`, `name`, `module`, `type`, `description`, `front_path`, `icon`, `sort_order`, `status`, `create_time`, `update_time`)
VALUES ('setting:wechat:read', '查看微信配置', 'setting', 'menu', '查看微信小程序配置', '/admin/settings/wechat', 'WechatOutlined', 93, 1, NOW(), NOW())
ON DUPLICATE KEY UPDATE `status` = 1, `update_time` = NOW();

-- 插入微信配置编辑权限（按钮）
INSERT INTO `t_permission` (`perm_key`, `name`, `module`, `type`, `description`, `front_path`, `icon`, `sort_order`, `status`, `create_time`, `update_time`)
VALUES ('setting:wechat:update', '编辑微信配置', 'setting', 'button', '修改微信小程序配置', '', '', 94, 1, NOW(), NOW())
ON DUPLICATE KEY UPDATE `status` = 1, `update_time` = NOW();

-- ============================================================
-- 迁移脚本执行完成
-- ============================================================