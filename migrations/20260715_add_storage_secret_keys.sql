-- ============================================================
-- 为 t_blog_infrastructure_config 添加存储相关字段
-- ============================================================
-- 日期: 2026-07-15
-- 说明: 添加七牛云和阿里云的 access_key / secret_key 列，
--       以及七牛云 region 列，使存储配置可以持久化到数据库
-- ============================================================

ALTER TABLE `t_blog_infrastructure_config`
  ADD COLUMN `upload_qiniu_access_key` varchar(120) NOT NULL DEFAULT '' COMMENT '七牛AccessKey' AFTER `upload_qiniu_domain`,
  ADD COLUMN `upload_qiniu_secret_key` varchar(120) NOT NULL DEFAULT '' COMMENT '七牛SecretKey' AFTER `upload_qiniu_access_key`,
  ADD COLUMN `upload_qiniu_region` varchar(20) NOT NULL DEFAULT '' COMMENT '七牛存储区域' AFTER `upload_qiniu_secret_key`,
  ADD COLUMN `upload_aliyun_access_key` varchar(120) NOT NULL DEFAULT '' COMMENT '阿里云AccessKey' AFTER `upload_aliyun_domain`,
  ADD COLUMN `upload_aliyun_secret_key` varchar(120) NOT NULL DEFAULT '' COMMENT '阿里云SecretKey' AFTER `upload_aliyun_access_key`;
