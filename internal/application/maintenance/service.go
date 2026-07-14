package maintenance

import (
	"context"
	"fmt"
	"time"

	"sanmoo-server-go/internal/infrastructure/logger"

	"gorm.io/gorm"
)

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

type CleanupResult struct {
	TableName  string `json:"tableName"`
	Deleted    int64  `json:"deleted"`
	RetainDays int    `json:"retainDays"`
	CutoffDate string `json:"cutoffDate"`
}

type CleanupReport struct {
	Tables       []CleanupResult `json:"tables"`
	TotalDeleted int64           `json:"totalDeleted"`
	Duration     int64           `json:"duration"` // milliseconds
	Success      bool            `json:"success"`
	Message      string          `json:"message"`
}

type TableStats struct {
	TableName string `json:"tableName"`
	RowCount  int64  `json:"rowCount"`
	DataSize  string `json:"dataSize"`
	IndexSize string `json:"indexSize"`
}

type MaintenanceStats struct {
	Tables    []TableStats `json:"tables"`
	TotalRows int64        `json:"totalRows"`
}

var retentionPolicies = map[string]int{
	"t_access_log":            30,
	"t_error_log":             90,
	"t_search_history":        30,
	"t_mp_browse_history":     30,
	"t_mp_reco_exposure":      30,
	"t_statistics_article_pv": 365,
}

func (s *Service) CleanupExpiredLogs(ctx context.Context) (*CleanupReport, error) {
	start := time.Now()
	var results []CleanupResult
	var totalDeleted int64

	for tableName, retainDays := range retentionPolicies {
		cutoff := time.Now().AddDate(0, 0, -retainDays)
		cutoffStr := cutoff.Format("2006-01-02 15:04:05")

		var deleted int64
		timeField := "create_time"
		if tableName == "t_search_history" {
			timeField = "search_time"
		} else if tableName == "t_mp_reco_exposure" {
			timeField = "exposed_at"
		} else if tableName == "t_statistics_article_pv" {
			timeField = "pv_date"
		}

		result := s.db.WithContext(ctx).
			Table(tableName).
			Where(fmt.Sprintf("%s < ?", timeField), cutoff).
			Delete(nil)

		if result.Error != nil {
			logger.Warnf("清理表 %s 失败: %v", tableName, result.Error)
			continue
		}

		deleted = result.RowsAffected
		if deleted > 0 {
			totalDeleted += deleted
			results = append(results, CleanupResult{
				TableName:  tableName,
				Deleted:    deleted,
				RetainDays: retainDays,
				CutoffDate: cutoffStr,
			})
			logger.Infof("清理表 %s: 删除 %d 条记录（保留 %d 天，截止 %s）", tableName, deleted, retainDays, cutoffStr)
		}
	}

	duration := time.Since(start).Milliseconds()
	msg := fmt.Sprintf("清理完成，共删除 %d 条记录，耗时 %d ms", totalDeleted, duration)

	return &CleanupReport{
		Tables:       results,
		TotalDeleted: totalDeleted,
		Duration:     duration,
		Success:      true,
		Message:      msg,
	}, nil
}

func (s *Service) GetMaintenanceStats(ctx context.Context) (*MaintenanceStats, error) {
	var stats []TableStats
	var totalRows int64

	for tableName := range retentionPolicies {
		var rowCount int64
		if err := s.db.WithContext(ctx).Table(tableName).Count(&rowCount).Error; err != nil {
			logger.Warnf("统计表 %s 行数失败: %v", tableName, err)
			continue
		}

		var dataSize, indexSize string
		row := s.db.WithContext(ctx).Raw(`
			SELECT
				COALESCE(data_length, 0) as data_length,
				COALESCE(index_length, 0) as index_length
			FROM information_schema.tables
			WHERE table_schema = DATABASE() AND table_name = ?
		`, tableName).Row()

		if err := row.Scan(&dataSize, &indexSize); err != nil {
			logger.Warnf("查询表 %s 大小失败: %v", tableName, err)
			dataSize = "0"
			indexSize = "0"
		}

		dataSizeMB := float64(parseBytes(dataSize)) / 1024 / 1024
		indexSizeMB := float64(parseBytes(indexSize)) / 1024 / 1024

		stats = append(stats, TableStats{
			TableName: tableName,
			RowCount:  rowCount,
			DataSize:  fmt.Sprintf("%.2f MB", dataSizeMB),
			IndexSize: fmt.Sprintf("%.2f MB", indexSizeMB),
		})
		totalRows += rowCount
	}

	return &MaintenanceStats{
		Tables:    stats,
		TotalRows: totalRows,
	}, nil
}

func parseBytes(s string) int64 {
	if s == "" || s == "0" {
		return 0
	}
	var b int64
	fmt.Sscanf(s, "%d", &b)
	return b
}

func (s *Service) StartDailyCleanup() {
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				now := time.Now()
				if now.Hour() == 3 && now.Minute() == 0 {
					logger.Infof("定时清理任务开始执行（每日凌晨 3:00）")
					ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
					report, err := s.CleanupExpiredLogs(ctx)
					cancel()
					if err != nil {
						logger.Warnf("定时清理任务失败: %v", err)
					} else {
						logger.Infof("定时清理任务完成: %s", report.Message)
					}
				}
			}
		}
	}()
}
