-- ============================================
-- 清理重复文章（安全版：临时表 + 手动验证）
-- 执行时间：2026-07-17
-- ============================================

-- 步骤1：创建临时表，存储待删除的重复文章ID（保留最小ID）
CREATE TEMPORARY TABLE _dup_article_ids AS
SELECT a.id FROM t_article a
INNER JOIN (
    SELECT title, MIN(id) AS min_id
    FROM t_article
    GROUP BY title
    HAVING COUNT(*) > 1
) keep ON a.title = keep.title AND a.id > keep.min_id;

-- 步骤2：验证待删除的ID（必须执行！）
-- SELECT COUNT(*) AS should_delete FROM _dup_article_ids;
-- SELECT a.id, a.title FROM t_article a WHERE a.id IN (SELECT id FROM _dup_article_ids) ORDER BY a.title, a.id;

-- 步骤3：确认无误后，逐条执行以下删除（观察影响行数）
-- DELETE FROM t_article_content WHERE article_id IN (SELECT id FROM _dup_article_ids);
-- DELETE FROM t_article_category_rel WHERE article_id IN (SELECT id FROM _dup_article_ids);
-- DELETE FROM t_article_tag_rel WHERE article_id IN (SELECT id FROM _dup_article_ids);
-- DELETE FROM t_article_topic_rel WHERE article_id IN (SELECT id FROM _dup_article_ids);
-- DELETE FROM t_article WHERE id IN (SELECT id FROM _dup_article_ids);

-- 步骤4：清理临时表
-- DROP TEMPORARY TABLE _dup_article_ids;

-- 验证：应该没有重复标题
-- SELECT title, COUNT(*) AS cnt FROM t_article GROUP BY title HAVING cnt > 1;
-- SELECT COUNT(*) AS total, COUNT(DISTINCT title) AS unique_titles FROM t_article;