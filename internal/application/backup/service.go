package backup

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

type BackupResult struct {
	FileName string `json:"fileName"`
	FilePath string `json:"filePath"`
	Size     int64  `json:"size"`
	TableCount int   `json:"tableCount"`
	Success   bool  `json:"success"`
	Message  string `json:"message"`
}

type TableData struct {
	TableName  string        `json:"tableName"`
	ColumnNames []string     `json:"columnNames"`
	Rows       []interface{} `json:"rows"`
	RowCount   int           `json:"rowCount"`
}

func (s *Service) ExportAllData(ctx context.Context) (*BackupResult, error) {
	tables, err := s.getTableNames(ctx)
	if err != nil {
		return nil, err
	}

	var tableDataList []TableData
	for _, table := range tables {
		data, err := s.exportTable(ctx, table)
		if err != nil {
			logger.Warnf("导出表 %s 失败: %v", table, err)
			continue
		}
		tableDataList = append(tableDataList, data)
	}

	backupData := struct {
		ExportTime string       `json:"exportTime"`
		Version    string       `json:"version"`
		TableCount int          `json:"tableCount"`
		Tables     []TableData  `json:"tables"`
	}{
		ExportTime: time.Now().Format("2006-01-02 15:04:05"),
		Version:    "1.0",
		TableCount: len(tableDataList),
		Tables:     tableDataList,
	}

	data, err := json.MarshalIndent(backupData, "", "  ")
	if err != nil {
		return nil, err
	}

	backupDir := "./backups"
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return nil, err
	}

	fileName := fmt.Sprintf("backup_%s.json", time.Now().Format("20060102_150405"))
	filePath := filepath.Join(backupDir, fileName)

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return nil, err
	}

	info, _ := os.Stat(filePath)
	size := int64(0)
	if info != nil {
		size = info.Size()
	}

	return &BackupResult{
		FileName:   fileName,
		FilePath:   filePath,
		Size:       size,
		TableCount: len(tableDataList),
		Success:    true,
		Message:    fmt.Sprintf("成功导出 %d 张表", len(tableDataList)),
	}, nil
}

func (s *Service) getTableNames(ctx context.Context) ([]string, error) {
	var tables []string
	result := s.db.WithContext(ctx).Raw(`
		SELECT table_name 
		FROM information_schema.tables 
		WHERE table_schema = DATABASE() 
		ORDER BY table_name
	`).Scan(&tables)
	if result.Error != nil {
		return nil, result.Error
	}
	return tables, nil
}

func (s *Service) exportTable(ctx context.Context, tableName string) (TableData, error) {
	var rows []map[string]interface{}
	result := s.db.WithContext(ctx).Table(tableName).Find(&rows)
	if result.Error != nil {
		return TableData{}, result.Error
	}

	var columnNames []string
	if len(rows) > 0 {
		for key := range rows[0] {
			columnNames = append(columnNames, key)
		}
	}

	var interfaceRows []interface{}
	for _, row := range rows {
		interfaceRows = append(interfaceRows, row)
	}

	return TableData{
		TableName:  tableName,
		ColumnNames: columnNames,
		Rows:       interfaceRows,
		RowCount:   len(rows),
	}, nil
}

func (s *Service) ListBackups(ctx context.Context) ([]BackupResult, error) {
	backupDir := "./backups"
	entries, err := os.ReadDir(backupDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []BackupResult{}, nil
		}
		return nil, err
	}

	var results []BackupResult
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		results = append(results, BackupResult{
			FileName: entry.Name(),
			FilePath: filepath.Join(backupDir, entry.Name()),
			Size:     info.Size(),
			Success:  true,
		})
	}

	return results, nil
}

func (s *Service) DeleteBackup(ctx context.Context, fileName string) error {
	backupDir := "./backups"
	filePath := filepath.Join(backupDir, fileName)

	if !strings.HasSuffix(fileName, ".json") {
		return fmt.Errorf("invalid file type")
	}

	if filepath.Base(filePath) != fileName {
		return fmt.Errorf("invalid file path")
	}

	return os.Remove(filePath)
}
