ALTER TABLE t_blog_storage_config
ADD COLUMN `upload_qiniu_access_key` varchar(200) NOT NULL DEFAULT '' COMMENT '七牛AccessKey' AFTER `upload_qiniu_domain`,
ADD COLUMN `upload_qiniu_secret_key` varchar(200) NOT NULL DEFAULT '' COMMENT '七牛SecretKey' AFTER `upload_qiniu_access_key`;