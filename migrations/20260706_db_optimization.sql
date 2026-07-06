-- ============================================================
-- 数据库优化迭代脚本
-- 日期: 2026-07-06
-- 说明: 基于数据库设计分析报告，添加缺失索引、约束，清理冗余表
-- ============================================================

-- --------------------------------------------------------
-- 1. 删除未使用的 t_visitor_record 表
--    功能已被 t_access_log 完全覆盖
-- --------------------------------------------------------
DROP TABLE IF EXISTS t_visitor_record;

-- --------------------------------------------------------
-- 2. t_article_content 添加唯一约束
--    防止同一文章出现多条内容记录
-- --------------------------------------------------------
ALTER TABLE t_article_content ADD UNIQUE KEY uk_article_id (article_id);

-- --------------------------------------------------------
-- 3. t_article_tag_rel 添加索引和唯一约束
-- --------------------------------------------------------
ALTER TABLE t_article_tag_rel ADD INDEX idx_article_id (article_id);
ALTER TABLE t_article_tag_rel ADD INDEX idx_tag_id (tag_id);
ALTER TABLE t_article_tag_rel ADD UNIQUE KEY uk_article_tag (article_id, tag_id);

-- --------------------------------------------------------
-- 4. t_article_category_rel 添加索引和唯一约束
-- --------------------------------------------------------
ALTER TABLE t_article_category_rel ADD INDEX idx_article_id (article_id);
ALTER TABLE t_article_category_rel ADD INDEX idx_category_id (category_id);
ALTER TABLE t_article_category_rel ADD UNIQUE KEY uk_article_category (article_id, category_id);

-- --------------------------------------------------------
-- 5. t_article_topic_rel 添加索引和唯一约束
-- --------------------------------------------------------
ALTER TABLE t_article_topic_rel ADD INDEX idx_article_id (article_id);
ALTER TABLE t_article_topic_rel ADD INDEX idx_topic_id (topic_id);
ALTER TABLE t_article_topic_rel ADD UNIQUE KEY uk_article_topic (article_id, topic_id);

-- --------------------------------------------------------
-- 6. t_statistics_article_pv 添加索引
--    IncreaseReadAndPV 使用 article_id + pv_date 查询
-- --------------------------------------------------------
ALTER TABLE t_statistics_article_pv ADD UNIQUE KEY uk_article_date (article_id, pv_date);
ALTER TABLE t_statistics_article_pv ADD INDEX idx_pv_date (pv_date);

-- --------------------------------------------------------
-- 7. t_mp_user_favorite 添加索引和唯一约束
--    防止同一用户重复收藏同一文章
-- --------------------------------------------------------
ALTER TABLE t_mp_user_favorite ADD INDEX idx_openid (openid);
ALTER TABLE t_mp_user_favorite ADD UNIQUE KEY uk_openid_article (openid, article_id);

-- --------------------------------------------------------
-- 8. t_search_history 添加索引
--    优化热搜关键词查询（按 search_time 范围 + keyword 分组）
-- --------------------------------------------------------
ALTER TABLE t_search_history ADD INDEX idx_search_time (search_time);
ALTER TABLE t_search_history ADD INDEX idx_keyword (keyword);

-- --------------------------------------------------------
-- 9. t_mp_user_behavior 添加索引
--    常用查询：按 openid + event_type 查询用户行为
-- --------------------------------------------------------
ALTER TABLE t_mp_user_behavior ADD INDEX idx_openid (openid);
ALTER TABLE t_mp_user_behavior ADD INDEX idx_openid_event (openid, event_type);
ALTER TABLE t_mp_user_behavior ADD INDEX idx_article_id (article_id);

-- --------------------------------------------------------
-- 10. t_mp_user_interest 添加索引
--     常用查询：按 openid 查询用户兴趣
-- --------------------------------------------------------
ALTER TABLE t_mp_user_interest ADD INDEX idx_openid (openid);
ALTER TABLE t_mp_user_interest ADD INDEX idx_openid_dimension (openid, dimension_type);

-- --------------------------------------------------------
-- 11. t_mp_user_tag 添加索引
--     常用查询：按 openid 查询用户标签
-- --------------------------------------------------------
ALTER TABLE t_mp_user_tag ADD INDEX idx_openid (openid);

-- --------------------------------------------------------
-- 12. t_mp_user_profile 添加索引
--     常用查询：按 openid 查询用户画像
-- --------------------------------------------------------
ALTER TABLE t_mp_user_profile ADD INDEX idx_openid (openid);

-- --------------------------------------------------------
-- 13. t_access_log 添加索引
--     优化访问日志查询（按 request_time + ip_address）
-- --------------------------------------------------------
ALTER TABLE t_access_log ADD INDEX idx_request_time (request_time);
ALTER TABLE t_access_log ADD INDEX idx_ip_address (ip_address);
ALTER TABLE t_access_log ADD INDEX idx_is_error (is_error);

-- --------------------------------------------------------
-- 14. t_error_log 添加索引
--     优化错误日志查询
-- --------------------------------------------------------
ALTER TABLE t_error_log ADD INDEX idx_create_time (create_time);
ALTER TABLE t_error_log ADD INDEX idx_access_log_id (access_log_id);

-- --------------------------------------------------------
-- 15. t_article 添加索引
--     优化文章列表查询（按 is_published + create_time）
-- --------------------------------------------------------
ALTER TABLE t_article ADD INDEX idx_is_published_create_time (is_published, create_time);
