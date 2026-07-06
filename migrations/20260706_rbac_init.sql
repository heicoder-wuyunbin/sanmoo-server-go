-- ============================================================
-- RBAC 权限管理系统初始化脚本
-- 日期: 2026-07-06
-- 说明: 五表 RBAC 模型（用户-角色-权限），支持多角色、权限到方法级
-- ============================================================

-- --------------------------------------------------------
-- 1. 扩展 t_role 表（补充字段）
-- --------------------------------------------------------
ALTER TABLE t_role ADD COLUMN description VARCHAR(255) DEFAULT NULL COMMENT '角色描述' AFTER name;
ALTER TABLE t_role ADD COLUMN status TINYINT NOT NULL DEFAULT 1 COMMENT '状态 1启用 0禁用' AFTER description;
ALTER TABLE t_role ADD COLUMN sort_order INT NOT NULL DEFAULT 0 COMMENT '排序' AFTER status;
ALTER TABLE t_role ADD COLUMN create_time DATETIME DEFAULT NULL AFTER sort_order;
ALTER TABLE t_role ADD COLUMN update_time DATETIME DEFAULT NULL AFTER create_time;

-- 补充 t_role 已有数据的时间字段
UPDATE t_role SET create_time = NOW(), update_time = NOW() WHERE create_time IS NULL;

-- --------------------------------------------------------
-- 2. 权限表 t_permission
-- --------------------------------------------------------
CREATE TABLE IF NOT EXISTS t_permission (
  id           BIGINT PRIMARY KEY AUTO_INCREMENT,
  perm_key     VARCHAR(128) NOT NULL COMMENT '权限标识，如 article:create',
  name         VARCHAR(64)  NOT NULL COMMENT '权限名称',
  module       VARCHAR(64)  NOT NULL COMMENT '所属模块',
  type         VARCHAR(32)  NOT NULL DEFAULT 'api' COMMENT '类型：api / menu / button',
  description  VARCHAR(255) DEFAULT NULL COMMENT '描述',
  sort_order   INT          NOT NULL DEFAULT 0 COMMENT '排序',
  status       TINYINT      NOT NULL DEFAULT 1 COMMENT '1启用 0禁用',
  create_time  DATETIME     NOT NULL,
  update_time  DATETIME     NOT NULL,
  UNIQUE KEY uk_perm_key (perm_key),
  KEY idx_module (module),
  KEY idx_type (type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='权限表';

-- --------------------------------------------------------
-- 3. 角色-权限关联表 t_role_permission
-- --------------------------------------------------------
CREATE TABLE IF NOT EXISTS t_role_permission (
  id          BIGINT PRIMARY KEY AUTO_INCREMENT,
  role_id     BIGINT       NOT NULL,
  perm_key    VARCHAR(128) NOT NULL,
  create_time DATETIME     NOT NULL,
  UNIQUE KEY uk_role_perm (role_id, perm_key),
  KEY idx_role_id (role_id),
  KEY idx_perm_key (perm_key)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='角色-权限关联表';

-- --------------------------------------------------------
-- 4. 种子数据：权限初始化（按模块分类）
-- --------------------------------------------------------

-- 仪表盘模块
INSERT INTO t_permission (perm_key, name, module, type, description, sort_order, status, create_time, update_time) VALUES
('dashboard:read',       '查看仪表盘',     'dashboard', 'menu',   '访问仪表盘页面',          1, 1, NOW(), NOW()),
('dashboard:visitors',   '访问记录',       'dashboard', 'menu',   '查看访问记录',            2, 1, NOW(), NOW()),
('dashboard:errors',     '错误日志',       'dashboard', 'menu',   '查看错误日志',            3, 1, NOW(), NOW()),
('dashboard:pv',         'PV统计',         'dashboard', 'button', '查看PV统计数据',          4, 1, NOW(), NOW()),
('dashboard:statistics', '统计图表',       'dashboard', 'button', '查看各类统计图表',        5, 1, NOW(), NOW()),
('dashboard:visitors:delete', '删除访问记录', 'dashboard', 'button', '删除访问记录',         6, 1, NOW(), NOW()),
('dashboard:errors:delete',   '删除错误日志', 'dashboard', 'button', '删除错误日志',          7, 1, NOW(), NOW());

-- 文章模块
INSERT INTO t_permission (perm_key, name, module, type, description, sort_order, status, create_time, update_time) VALUES
('article:list',     '文章列表',   'article', 'menu',   '查看文章列表',     10, 1, NOW(), NOW()),
('article:detail',   '文章详情',   'article', 'button', '查看文章详情',     11, 1, NOW(), NOW()),
('article:create',   '创建文章',   'article', 'button', '创建新文章',       12, 1, NOW(), NOW()),
('article:update',   '编辑文章',   'article', 'button', '编辑文章',         13, 1, NOW(), NOW()),
('article:delete',   '删除文章',   'article', 'button', '删除文章',         14, 1, NOW(), NOW()),
('article:status',   '文章状态',   'article', 'button', '发布/下架文章',    15, 1, NOW(), NOW());

-- 分类模块
INSERT INTO t_permission (perm_key, name, module, type, description, sort_order, status, create_time, update_time) VALUES
('category:list',   '分类列表',   'category', 'menu',   '查看分类列表',   20, 1, NOW(), NOW()),
('category:create', '创建分类',   'category', 'button', '创建新分类',     21, 1, NOW(), NOW()),
('category:update', '编辑分类',   'category', 'button', '编辑分类',       22, 1, NOW(), NOW()),
('category:delete', '删除分类',   'category', 'button', '删除分类',       23, 1, NOW(), NOW());

-- 标签模块
INSERT INTO t_permission (perm_key, name, module, type, description, sort_order, status, create_time, update_time) VALUES
('tag:list',   '标签列表',   'tag', 'menu',   '查看标签列表',   30, 1, NOW(), NOW()),
('tag:create', '创建标签',   'tag', 'button', '创建新标签',     31, 1, NOW(), NOW()),
('tag:update', '编辑标签',   'tag', 'button', '编辑标签',       32, 1, NOW(), NOW()),
('tag:delete', '删除标签',   'tag', 'button', '删除标签',       33, 1, NOW(), NOW());

-- 专题模块
INSERT INTO t_permission (perm_key, name, module, type, description, sort_order, status, create_time, update_time) VALUES
('topic:list',     '专题列表',   'topic', 'menu',   '查看专题列表',     40, 1, NOW(), NOW()),
('topic:create',   '创建专题',   'topic', 'button', '创建新专题',       41, 1, NOW(), NOW()),
('topic:update',   '编辑专题',   'topic', 'button', '编辑专题',         42, 1, NOW(), NOW()),
('topic:delete',   '删除专题',   'topic', 'button', '删除专题',         43, 1, NOW(), NOW()),
('topic:articles', '专题文章',   'topic', 'button', '管理专题文章',     44, 1, NOW(), NOW());

-- 友情链接模块
INSERT INTO t_permission (perm_key, name, module, type, description, sort_order, status, create_time, update_time) VALUES
('link:list',   '友链列表',   'link', 'menu',   '查看友情链接',   50, 1, NOW(), NOW()),
('link:create', '创建友链',   'link', 'button', '创建友情链接',   51, 1, NOW(), NOW()),
('link:update', '编辑友链',   'link', 'button', '编辑友情链接',   52, 1, NOW(), NOW()),
('link:delete', '删除友链',   'link', 'button', '删除友情链接',   53, 1, NOW(), NOW());

-- 文件模块
INSERT INTO t_permission (perm_key, name, module, type, description, sort_order, status, create_time, update_time) VALUES
('file:list',   '文件列表',   'file', 'menu',   '查看文件列表',   60, 1, NOW(), NOW()),
('file:upload', '上传文件',   'file', 'button', '上传文件',       61, 1, NOW(), NOW()),
('file:delete', '删除文件',   'file', 'button', '删除文件',       62, 1, NOW(), NOW());

-- 用户模块
INSERT INTO t_permission (perm_key, name, module, type, description, sort_order, status, create_time, update_time) VALUES
('user:list',          '用户列表',       'user', 'menu',   '查看用户列表',           70, 1, NOW(), NOW()),
('user:create',        '创建用户',       'user', 'button', '创建新用户',             71, 1, NOW(), NOW()),
('user:update',        '编辑用户',       'user', 'button', '编辑用户信息',           72, 1, NOW(), NOW()),
('user:delete',        '删除用户',       'user', 'button', '删除用户',               73, 1, NOW(), NOW()),
('user:password:reset','重置密码',       'user', 'button', '重置用户密码',           74, 1, NOW(), NOW()),
('user:status',        '用户状态',       'user', 'button', '启用/禁用用户',          75, 1, NOW(), NOW()),
('user:role',          '角色分配',       'user', 'button', '分配用户角色',           76, 1, NOW(), NOW());

-- 设置模块
INSERT INTO t_permission (perm_key, name, module, type, description, sort_order, status, create_time, update_time) VALUES
('setting:read',    '查看设置',     'setting', 'menu',   '查看系统设置',      80, 1, NOW(), NOW()),
('setting:update',  '编辑设置',     'setting', 'button', '修改系统设置',      81, 1, NOW(), NOW()),
('setting:email',   '邮箱配置',     'setting', 'button', '配置邮箱服务',      82, 1, NOW(), NOW()),
('setting:import',  '导入设置',     'setting', 'button', '导入配置',          83, 1, NOW(), NOW()),
('setting:export',  '导出设置',     'setting', 'button', '导出配置',          84, 1, NOW(), NOW()),
('setting:search',  '搜索同步',     'setting', 'button', '同步搜索索引',      85, 1, NOW(), NOW());

-- 缓存模块
INSERT INTO t_permission (perm_key, name, module, type, description, sort_order, status, create_time, update_time) VALUES
('cache:clear',   '清除缓存',   'cache', 'button', '清除所有缓存',    90, 1, NOW(), NOW()),
('cache:warmup',  '缓存预热',   'cache', 'button', '预热缓存',        91, 1, NOW(), NOW()),
('cache:stats',   '缓存统计',   'cache', 'button', '查看缓存统计',    92, 1, NOW(), NOW());

-- 备份模块
INSERT INTO t_permission (perm_key, name, module, type, description, sort_order, status, create_time, update_time) VALUES
('backup:export',   '数据导出',   'backup', 'button', '导出数据备份',    100, 1, NOW(), NOW()),
('backup:list',     '备份列表',   'backup', 'button', '查看备份列表',    101, 1, NOW(), NOW()),
('backup:download', '下载备份',   'backup', 'button', '下载备份文件',    102, 1, NOW(), NOW()),
('backup:delete',   '删除备份',   'backup', 'button', '删除备份文件',    103, 1, NOW(), NOW()),
('backup:stats',    '备份统计',   'backup', 'button', '查看备份统计',    104, 1, NOW(), NOW());

-- 微信用户模块
INSERT INTO t_permission (perm_key, name, module, type, description, sort_order, status, create_time, update_time) VALUES
('mpuser:list',     '小程序用户列表', 'mpuser', 'menu',   '查看小程序用户列表',    110, 1, NOW(), NOW()),
('mpuser:detail',   '用户详情',       'mpuser', 'button', '查看用户详情',          111, 1, NOW(), NOW()),
('mpuser:profile',  '用户画像',       'mpuser', 'button', '生成/查看用户画像',     112, 1, NOW(), NOW()),
('mpuser:tags',     '用户标签',       'mpuser', 'button', '管理用户标签',          113, 1, NOW(), NOW());

-- 权限模块
INSERT INTO t_permission (perm_key, name, module, type, description, sort_order, status, create_time, update_time) VALUES
('permission:list',  '权限列表',  'permission', 'menu',   '查看所有权限',    120, 1, NOW(), NOW());

-- 角色模块
INSERT INTO t_permission (perm_key, name, module, type, description, sort_order, status, create_time, update_time) VALUES
('role:list',       '角色列表',    'role', 'menu',   '查看角色列表',       130, 1, NOW(), NOW()),
('role:create',     '创建角色',    'role', 'button', '创建新角色',         131, 1, NOW(), NOW()),
('role:update',     '编辑角色',    'role', 'button', '编辑角色信息',       132, 1, NOW(), NOW()),
('role:delete',     '删除角色',    'role', 'button', '删除角色',           133, 1, NOW(), NOW()),
('role:permission', '分配权限',    'role', 'button', '给角色分配权限',     134, 1, NOW(), NOW());

-- --------------------------------------------------------
-- 5. 种子数据：默认角色
-- --------------------------------------------------------

-- 确保 admin 角色存在（admin 拥有所有权限，无需关联 t_role_permission）
INSERT INTO t_role (id, name, description, status, sort_order, create_time, update_time)
VALUES (1, 'admin', '超级管理员，拥有全部权限', 1, 1, NOW(), NOW())
ON DUPLICATE KEY UPDATE name = VALUES(name), description = VALUES(description);

-- 编辑器角色
INSERT INTO t_role (id, name, description, status, sort_order, create_time, update_time)
VALUES (2, 'editor', '内容编辑，可管理文章/分类/标签/专题', 1, 2, NOW(), NOW())
ON DUPLICATE KEY UPDATE name = VALUES(name), description = VALUES(description);

-- 访客角色（只读）
INSERT INTO t_role (id, name, description, status, sort_order, create_time, update_time)
VALUES (3, 'viewer', '只读访客，仅可查看数据', 1, 3, NOW(), NOW())
ON DUPLICATE KEY UPDATE name = VALUES(name), description = VALUES(description);

-- --------------------------------------------------------
-- 6. 种子数据：editor 角色权限（内容编辑相关）
-- --------------------------------------------------------
INSERT IGNORE INTO t_role_permission (role_id, perm_key, create_time) VALUES
-- 仪表盘
(2, 'dashboard:read', NOW()),
(2, 'dashboard:visitors', NOW()),
(2, 'dashboard:errors', NOW()),
(2, 'dashboard:pv', NOW()),
(2, 'dashboard:statistics', NOW()),
-- 文章
(2, 'article:list', NOW()),
(2, 'article:detail', NOW()),
(2, 'article:create', NOW()),
(2, 'article:update', NOW()),
(2, 'article:delete', NOW()),
(2, 'article:status', NOW()),
-- 分类
(2, 'category:list', NOW()),
(2, 'category:create', NOW()),
(2, 'category:update', NOW()),
(2, 'category:delete', NOW()),
-- 标签
(2, 'tag:list', NOW()),
(2, 'tag:create', NOW()),
(2, 'tag:update', NOW()),
(2, 'tag:delete', NOW()),
-- 专题
(2, 'topic:list', NOW()),
(2, 'topic:create', NOW()),
(2, 'topic:update', NOW()),
(2, 'topic:delete', NOW()),
(2, 'topic:articles', NOW()),
-- 友链
(2, 'link:list', NOW()),
(2, 'link:create', NOW()),
(2, 'link:update', NOW()),
(2, 'link:delete', NOW()),
-- 文件
(2, 'file:list', NOW()),
(2, 'file:upload', NOW()),
(2, 'file:delete', NOW()),
-- 设置（只读）
(2, 'setting:read', NOW()),
-- 小程序用户
(2, 'mpuser:list', NOW()),
(2, 'mpuser:detail', NOW());

-- --------------------------------------------------------
-- 7. 种子数据：viewer 角色权限（只读）
-- --------------------------------------------------------
INSERT IGNORE INTO t_role_permission (role_id, perm_key, create_time) VALUES
(3, 'dashboard:read', NOW()),
(3, 'dashboard:visitors', NOW()),
(3, 'dashboard:errors', NOW()),
(3, 'dashboard:pv', NOW()),
(3, 'dashboard:statistics', NOW()),
(3, 'article:list', NOW()),
(3, 'article:detail', NOW()),
(3, 'category:list', NOW()),
(3, 'tag:list', NOW()),
(3, 'topic:list', NOW()),
(3, 'link:list', NOW()),
(3, 'file:list', NOW()),
(3, 'setting:read', NOW()),
(3, 'mpuser:list', NOW()),
(3, 'mpuser:detail', NOW());
