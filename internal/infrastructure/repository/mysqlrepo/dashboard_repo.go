package mysqlrepo

import (
	"context"
	"net"
	"time"

	domdashboard "sanmoo-server-go/internal/domain/dashboard"
)

func (r *Repository) Statistics(ctx context.Context) (*domdashboard.Statistics, error) {
	out := &domdashboard.Statistics{}

	// 统计文章总数
	if err := r.db.WithContext(ctx).Table("t_article").Count(&out.ArticleCount).Error; err != nil {
		return nil, err
	}

	// 统计已发布文章数量
	if err := r.db.WithContext(ctx).Table("t_article").Where("is_published=?", true).Count(&out.PublishedArticleCount).Error; err != nil {
		return nil, err
	}

	// 计算未发布文章数量
	out.UnpublishedArticleCount = out.ArticleCount - out.PublishedArticleCount

	// 统计分类数量
	if err := r.db.WithContext(ctx).Table("t_category").Count(&out.CategoryCount).Error; err != nil {
		return nil, err
	}

	// 统计标签数量
	if err := r.db.WithContext(ctx).Table("t_tag").Count(&out.TagCount).Error; err != nil {
		return nil, err
	}

	// 统计用户数量
	if err := r.db.WithContext(ctx).Table("t_user").Count(&out.UserCount).Error; err != nil {
		return nil, err
	}

	// 统计今日独立访客数（根据IP去重）
	today := time.Now().Format("2006-01-02")
	if err := r.db.WithContext(ctx).Table("t_access_log").
		Where("DATE(request_time)=?", today).
		Distinct("ip_address").
		Count(&out.VisitorCount).Error; err != nil {
		return nil, err
	}

	// 统计今日 PV
	if err := r.db.WithContext(ctx).Table("t_statistics_article_pv").Where("DATE(pv_date)=?", today).Select("COALESCE(SUM(pv_count), 0)").Scan(&out.TodayPv).Error; err != nil {
		return nil, err
	}

	// 统计专题数量
	if err := r.db.WithContext(ctx).Table("t_topic").Count(&out.TopicCount).Error; err != nil {
		return nil, err
	}

	// 统计小程序用户数量（仅启用状态）
	if err := r.db.WithContext(ctx).Table("t_mp_user").Where("status = ?", 1).Count(&out.MpUserCount).Error; err != nil {
		return nil, err
	}

	// 统计总阅读量
	if err := r.db.WithContext(ctx).Table("t_article").Select("COALESCE(SUM(read_num), 0)").Scan(&out.TotalReads).Error; err != nil {
		return nil, err
	}

	// 统计昨日 PV
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	if err := r.db.WithContext(ctx).Table("t_statistics_article_pv").Where("DATE(pv_date)=?", yesterday).Select("COALESCE(SUM(pv_count), 0)").Scan(&out.YesterdayPv).Error; err != nil {
		return nil, err
	}

	// 统计今日新增小程序用户（仅启用状态）
	if err := r.db.WithContext(ctx).Table("t_mp_user").
		Where("status = ? AND DATE(create_time)=?", 1, today).
		Count(&out.TodayMpUserCount).Error; err != nil {
		return nil, err
	}

	// 统计昨日新增小程序用户（仅启用状态）
	if err := r.db.WithContext(ctx).Table("t_mp_user").
		Where("status = ? AND DATE(create_time)=?", 1, yesterday).
		Count(&out.YesterdayMpUserCount).Error; err != nil {
		return nil, err
	}

	return out, nil
}

func (r *Repository) VisitorList(ctx context.Context, page, size int, keyword string) ([]domdashboard.VisitorRecord, int64, error) {
	q := r.db.WithContext(ctx).Model(&TAccessLog{})
	if keyword != "" {
		q = q.Where("request_url LIKE ? OR ip_address LIKE ? OR visitor_name LIKE ? OR trace_id LIKE ?",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []TAccessLog
	if err := q.Order("id desc").Offset((page - 1) * size).Limit(size).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	res := make([]domdashboard.VisitorRecord, 0, len(rows))
	for _, row := range rows {
		res = append(res, domdashboard.VisitorRecord{
			ID:             row.ID,
			TraceID:        row.TraceID,
			RequestMethod:  row.RequestMethod,
			RequestURL:     row.RequestURL,
			VisitorUserID:  row.VisitorUserID,
			VisitorName:    row.VisitorName,
			IPAddress:      toIPString(row.IPAddress),
			VisitTime:      row.RequestTime.Format("2006-01-02 15:04:05"),
			ResponseTime:   row.ResponseTime,
			ResponseStatus: row.ResponseStatus,
			RequestSource:  row.RequestSource,
			IsError:        row.IsError,
			UserAgent:      row.UserAgent,
			RequestParams:  row.RequestParams,
			RequestBody:    row.RequestBody,
			ResponseBody:   row.ResponseBody,
		})
	}
	return res, total, nil
}

func (r *Repository) ErrorLogList(ctx context.Context, page, size int, keyword string) ([]domdashboard.ErrorLogRecord, int64, error) {
	q := r.db.WithContext(ctx).Model(&TErrorLog{})
	if keyword != "" {
		q = q.Where("error_code LIKE ? OR error_message LIKE ? OR request_url LIKE ? OR ip_address LIKE ? OR trace_id LIKE ?",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []TErrorLog
	if err := q.Order("id desc").Offset((page - 1) * size).Limit(size).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	res := make([]domdashboard.ErrorLogRecord, 0, len(rows))
	for _, row := range rows {
		res = append(res, domdashboard.ErrorLogRecord{
			ID:            row.ID,
			AccessLogID:   row.AccessLogID,
			TraceID:       row.TraceID,
			ErrorCode:     row.ErrorCode,
			ErrorMessage:  row.ErrorMessage,
			ErrorDetail:   row.ErrorDetail,
			StackTrace:    row.StackTrace,
			RequestURL:    row.RequestURL,
			RequestMethod: row.RequestMethod,
			RequestParams: row.RequestParams,
			RequestBody:   row.RequestBody,
			ResponseBody:  row.ResponseBody,
			IPAddress:     toIPString(row.IPAddress),
			UserAgent:     row.UserAgent,
			CreateTime:    row.CreateTime.Format("2006-01-02 15:04:05"),
		})
	}
	return res, total, nil
}

func (r *Repository) DeleteVisitorRecord(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&TAccessLog{}, id).Error
}

func (r *Repository) BatchDeleteVisitorRecords(ctx context.Context, ids []uint64) error {
	return r.db.WithContext(ctx).Delete(&TAccessLog{}, ids).Error
}

func (r *Repository) ClearAllVisitorRecords(ctx context.Context) error {
	return r.db.WithContext(ctx).Exec("DELETE FROM t_access_log").Error
}

func (r *Repository) DeleteErrorLog(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&TErrorLog{}, id).Error
}

func (r *Repository) BatchDeleteErrorLogs(ctx context.Context, ids []uint64) error {
	return r.db.WithContext(ctx).Delete(&TErrorLog{}, ids).Error
}

func (r *Repository) ClearAllErrorLogs(ctx context.Context) error {
	return r.db.WithContext(ctx).Exec("DELETE FROM t_error_log").Error
}

func (r *Repository) ImportErrorLogs(ctx context.Context, logs []domdashboard.ErrorLogRecord) (int64, error) {
	if len(logs) == 0 {
		return 0, nil
	}
	rows := make([]TErrorLog, 0, len(logs))
	now := time.Now()
	for _, l := range logs {
		rows = append(rows, TErrorLog{
			TraceID:       l.TraceID,
			ErrorCode:     l.ErrorCode,
			ErrorMessage:  l.ErrorMessage,
			ErrorDetail:   l.ErrorDetail,
			StackTrace:    l.StackTrace,
			RequestURL:    l.RequestURL,
			RequestMethod: l.RequestMethod,
			RequestParams: l.RequestParams,
			RequestBody:   l.RequestBody,
			ResponseBody:  l.ResponseBody,
			IPAddress:     net.ParseIP(l.IPAddress).To4(),
			UserAgent:     l.UserAgent,
			CreateTime:    now,
		})
	}
	result := r.db.WithContext(ctx).Create(&rows)
	return result.RowsAffected, result.Error
}

func (r *Repository) ExportErrorLogs(ctx context.Context) ([]domdashboard.ErrorLogRecord, error) {
	var rows []TErrorLog
	if err := r.db.WithContext(ctx).Order("id desc").Find(&rows).Error; err != nil {
		return nil, err
	}
	res := make([]domdashboard.ErrorLogRecord, 0, len(rows))
	for _, row := range rows {
		res = append(res, domdashboard.ErrorLogRecord{
			ID:            row.ID,
			AccessLogID:   row.AccessLogID,
			TraceID:       row.TraceID,
			ErrorCode:     row.ErrorCode,
			ErrorMessage:  row.ErrorMessage,
			ErrorDetail:   row.ErrorDetail,
			StackTrace:    row.StackTrace,
			RequestURL:    row.RequestURL,
			RequestMethod: row.RequestMethod,
			RequestParams: row.RequestParams,
			RequestBody:   row.RequestBody,
			ResponseBody:  row.ResponseBody,
			IPAddress:     toIPString(row.IPAddress),
			UserAgent:     row.UserAgent,
			CreateTime:    row.CreateTime.Format("2006-01-02 15:04:05"),
		})
	}
	return res, nil
}

func toIPString(raw []byte) string {
	if len(raw) == 0 {
		return ""
	}
	ip := net.IP(raw)
	if len(raw) == 16 {
		if v4 := ip.To4(); v4 != nil {
			return v4.String()
		}
	}
	return ip.String()
}

// TagStatistics 统计每个标签的文章数量
func (r *Repository) TagStatistics(ctx context.Context) ([]domdashboard.NameValue, error) {
	type TagStat struct {
		TagID   uint64
		TagName string
		Count   int64
	}

	var stats []TagStat

	// 通过关联查询统计每个标签的文章数量
	err := r.db.WithContext(ctx).Table("t_tag t").
		Select("t.id as tag_id, t.name as tag_name, COUNT(atr.article_id) as count").
		Joins("left join t_article_tag_rel atr on atr.tag_id = t.id").
		Group("t.id, t.name").
		Order("count desc").
		Scan(&stats).Error

	if err != nil {
		return nil, err
	}

	// 转换为前端需要的格式
	result := make([]domdashboard.NameValue, 0, len(stats))
	for _, stat := range stats {
		result = append(result, domdashboard.NameValue{Name: stat.TagName, Value: stat.Count})
	}

	return result, nil
}

// CategoryStatistics 统计每个分类的文章数量
func (r *Repository) CategoryStatistics(ctx context.Context) ([]domdashboard.NameValue, error) {
	type CategoryStat struct {
		CategoryID   uint64
		CategoryName string
		Count        int64
	}

	var stats []CategoryStat

	// 通过关联查询统计每个分类的文章数量
	err := r.db.WithContext(ctx).Table("t_category c").
		Select("c.id as category_id, c.name as category_name, COUNT(acr.article_id) as count").
		Joins("left join t_article_category_rel acr on acr.category_id = c.id").
		Group("c.id, c.name").
		Order("count desc").
		Scan(&stats).Error

	if err != nil {
		return nil, err
	}

	// 转换为前端需要的格式
	result := make([]domdashboard.NameValue, 0, len(stats))
	for _, stat := range stats {
		result = append(result, domdashboard.NameValue{Name: stat.CategoryName, Value: stat.Count})
	}

	return result, nil
}

// ArticlePublishHeatmap 统计近30天每天的文章发布数量
func (r *Repository) ArticlePublishHeatmap(ctx context.Context) ([]domdashboard.DateCount, error) {
	type HeatmapData struct {
		Date  string
		Count int64
	}

	var data []HeatmapData

	// 统计近30天每天的文章发布数量
	thirtyDaysAgo := time.Now().AddDate(0, 0, -29).Format("2006-01-02")
	err := r.db.WithContext(ctx).Table("t_article").
		Select("DATE(create_time) as date, COUNT(*) as count").
		Where("create_time >= ?", thirtyDaysAgo).
		Group("DATE(create_time)").
		Order("date").
		Scan(&data).Error

	if err != nil {
		return nil, err
	}

	// 转换为前端需要的格式
	result := make([]domdashboard.DateCount, 0, len(data))
	for _, item := range data {
		result = append(result, domdashboard.DateCount{Date: item.Date, Count: item.Count})
	}

	return result, nil
}

// TopicStatistics 统计每个专题的文章数量
func (r *Repository) TopicStatistics(ctx context.Context) ([]domdashboard.NameValue, error) {
	type TopicStat struct {
		TopicID   uint64
		TopicName string
		Count     int64
	}

	var stats []TopicStat

	err := r.db.WithContext(ctx).Table("t_topic t").
		Select("t.id as topic_id, t.name as topic_name, COUNT(atr.article_id) as count").
		Joins("left join t_article_topic_rel atr on atr.topic_id = t.id").
		Group("t.id, t.name").
		Order("count desc").
		Scan(&stats).Error

	if err != nil {
		return nil, err
	}

	result := make([]domdashboard.NameValue, 0, len(stats))
	for _, stat := range stats {
		result = append(result, domdashboard.NameValue{Name: stat.TopicName, Value: stat.Count})
	}

	return result, nil
}

// MpUserGrowth 统计近N天每天的小程序用户新增数量（仅启用状态）
func (r *Repository) MpUserGrowth(ctx context.Context, days int) ([]domdashboard.DateCount, error) {
	type GrowthData struct {
		Date  string
		Count int64
	}

	var data []GrowthData

	startDate := time.Now().AddDate(0, 0, -(days - 1)).Format("2006-01-02")
	err := r.db.WithContext(ctx).Table("t_mp_user").
		Select("DATE(create_time) as date, COUNT(*) as count").
		Where("status = ? AND create_time >= ?", 1, startDate).
		Group("DATE(create_time)").
		Order("date").
		Scan(&data).Error

	if err != nil {
		return nil, err
	}

	result := make([]domdashboard.DateCount, 0, len(data))
	for _, item := range data {
		result = append(result, domdashboard.DateCount{Date: item.Date, Count: item.Count})
	}

	return result, nil
}

// ArticleReadStatistics 统计文章阅读量（按阅读量排序）
func (r *Repository) ArticleReadStatistics(ctx context.Context, page, size int) ([]domdashboard.ArticleReadStat, int64, error) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 20
	}

	var total int64
	if err := r.db.WithContext(ctx).Table("t_article").Count(&total).Error; err != nil {
		return nil, 0, err
	}

	type row struct {
		ID          uint64
		Title       string
		ReadNum     int
		CategoryID  uint64
		Category    string
		CreateTime  time.Time
	}

	var rows []row
	err := r.db.WithContext(ctx).Table("t_article a").
		Select("a.id, a.title, a.read_num, a.create_time, acr.category_id, c.name as category").
		Joins("left join (select article_id, min(category_id) as category_id from t_article_category_rel group by article_id) acr on acr.article_id = a.id").
		Joins("left join t_category c on c.id = acr.category_id").
		Order("a.read_num desc").
		Offset((page - 1) * size).
		Limit(size).
		Scan(&rows).Error

	if err != nil {
		return nil, 0, err
	}

	result := make([]domdashboard.ArticleReadStat, 0, len(rows))
	for _, item := range rows {
		result = append(result, domdashboard.ArticleReadStat{
			ID:         item.ID,
			Title:      item.Title,
			ReadNum:    item.ReadNum,
			CategoryID: item.CategoryID,
			Category:   item.Category,
			CreateTime: item.CreateTime.Format("2006-01-02"),
		})
	}

	return result, total, nil
}

// CategoryReadStatistics 统计各分类的阅读量
func (r *Repository) CategoryReadStatistics(ctx context.Context) ([]domdashboard.CategoryReadStat, error) {
	type row struct {
		ID          uint64
		Name        string
		ArticleCount int64
		TotalReads  int64
	}

	var rows []row
	err := r.db.WithContext(ctx).Table("t_category c").
		Select("c.id, c.name, COUNT(DISTINCT acr.article_id) as article_count, COALESCE(SUM(a.read_num), 0) as total_reads").
		Joins("left join t_article_category_rel acr on acr.category_id = c.id").
		Joins("left join t_article a on a.id = acr.article_id").
		Group("c.id, c.name").
		Order("total_reads desc").
		Scan(&rows).Error

	if err != nil {
		return nil, err
	}

	result := make([]domdashboard.CategoryReadStat, 0, len(rows))
	for _, item := range rows {
		result = append(result, domdashboard.CategoryReadStat{
			ID:          item.ID,
			Name:        item.Name,
			ArticleCount: item.ArticleCount,
			TotalReads:  item.TotalReads,
		})
	}

	return result, nil
}

// TagReadStatistics 统计各标签的阅读量
func (r *Repository) TagReadStatistics(ctx context.Context) ([]domdashboard.TagReadStat, error) {
	type row struct {
		ID          uint64
		Name        string
		ArticleCount int64
		TotalReads  int64
	}

	var rows []row
	err := r.db.WithContext(ctx).Table("t_tag t").
		Select("t.id, t.name, COUNT(DISTINCT atr.article_id) as article_count, COALESCE(SUM(a.read_num), 0) as total_reads").
		Joins("left join t_article_tag_rel atr on atr.tag_id = t.id").
		Joins("left join t_article a on a.id = atr.article_id").
		Group("t.id, t.name").
		Order("total_reads desc").
		Scan(&rows).Error

	if err != nil {
		return nil, err
	}

	result := make([]domdashboard.TagReadStat, 0, len(rows))
	for _, item := range rows {
		result = append(result, domdashboard.TagReadStat{
			ID:          item.ID,
			Name:        item.Name,
			ArticleCount: item.ArticleCount,
			TotalReads:  item.TotalReads,
		})
	}

	return result, nil
}

// ContentTrend 统计近N天的内容热度趋势
func (r *Repository) ContentTrend(ctx context.Context, days int) ([]domdashboard.ContentTrend, error) {
	if days <= 0 {
		days = 30
	}

	startDate := time.Now().AddDate(0, 0, -(days - 1)).Format("2006-01-02")

	type pvRow struct {
		Date    string
		PVCount int64
	}
	var pvData []pvRow
	err := r.db.WithContext(ctx).Table("t_statistics_article_pv").
		Select("DATE(pv_date) as date, SUM(pv_count) as pv_count").
		Where("pv_date >= ?", startDate).
		Group("DATE(pv_date)").
		Order("date").
		Scan(&pvData).Error
	if err != nil {
		return nil, err
	}

	pvMap := make(map[string]int64)
	for _, item := range pvData {
		pvMap[item.Date] = item.PVCount
	}

	type articleRow struct {
		Date  string
		Count int64
	}
	var articleData []articleRow
	err = r.db.WithContext(ctx).Table("t_article").
		Select("DATE(create_time) as date, COUNT(*) as count").
		Where("create_time >= ?", startDate).
		Group("DATE(create_time)").
		Order("date").
		Scan(&articleData).Error
	if err != nil {
		return nil, err
	}

	articleMap := make(map[string]int64)
	for _, item := range articleData {
		articleMap[item.Date] = item.Count
	}

	result := make([]domdashboard.ContentTrend, 0, days)
	for i := days - 1; i >= 0; i-- {
		date := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
		result = append(result, domdashboard.ContentTrend{
			Date:        date,
			TotalPV:     pvMap[date],
			ArticlePV:   pvMap[date],
			NewArticles: articleMap[date],
		})
	}

	return result, nil
}
