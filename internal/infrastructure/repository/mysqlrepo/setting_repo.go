package mysqlrepo

import (
	"context"
	"encoding/json"
	"strings"
	"unicode"

	domsetting "sanmoo-server-go/internal/domain/setting"

	"gorm.io/gorm"
)

func snakeToCamel(s string) string {
	parts := strings.Split(s, "_")
	for i := 1; i < len(parts); i++ {
		parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
	}
	return strings.Join(parts, "")
}

func camelToSnake(s string) string {
	var res []rune
	for i, r := range s {
		// 将驼峰转为下划线，并仅对大写字母做小写化处理
		if i > 0 && unicode.IsUpper(r) {
			res = append(res, '_')
		}
		if unicode.IsUpper(r) {
			res = append(res, unicode.ToLower(r))
		} else {
			res = append(res, r)
		}
	}
	return string(res)
}

func convertKeysSnakeToCamel(m map[string]any) map[string]any {
	res := make(map[string]any, len(m))
	for k, v := range m {
		res[snakeToCamel(k)] = v
	}
	return res
}

func convertKeysCamelToSnake(m map[string]any) map[string]any {
	res := make(map[string]any, len(m))
	for k, v := range m {
		res[camelToSnake(k)] = v
	}
	return res
}

func (r *Repository) Get(ctx context.Context) (*domsetting.BlogSettings, error) {
	coreRaw := map[string]any{}
	uiRaw := map[string]any{}
	storageRaw := map[string]any{}
	emailCfg := map[string]any{}
	if err := r.db.WithContext(ctx).Table("t_blog_core_config").Where("id=1").Take(&coreRaw).Error; err != nil {
		return nil, err
	}
	if err := r.db.WithContext(ctx).Table("t_blog_ui_config").Where("id=1").Take(&uiRaw).Error; err != nil {
		return nil, err
	}
	if err := r.db.WithContext(ctx).Table("t_blog_storage_config").Where("id=1").Take(&storageRaw).Error; err != nil {
		return nil, err
	}
	// 邮件配置：允许不存在（首次初始化）
	var emailRow struct {
		ConfigJSON string `gorm:"column:config_json"`
	}
	if err := r.db.WithContext(ctx).Table("t_blog_email_config").Select("config_json").Where("id=1").Take(&emailRow).Error; err == nil {
		_ = json.Unmarshal([]byte(emailRow.ConfigJSON), &emailCfg)
	}
	core := convertKeysSnakeToCamel(coreRaw)
	ui := convertKeysSnakeToCamel(uiRaw)
	storage := convertKeysSnakeToCamel(storageRaw)
	return &domsetting.BlogSettings{
		CoreConfig:    core,
		UIConfig:      ui,
		StorageConfig: storage,
		EmailConfig:   emailCfg,
	}, nil
}

func (r *Repository) Update(ctx context.Context, s *domsetting.BlogSettings, operator string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if len(s.CoreConfig) > 0 {
			core := convertKeysCamelToSnake(s.CoreConfig)
			core["updated_by"] = operator
			validCoreFields := []string{"blog_name", "author", "introduction", "avatar", "privacy_policy", "rss_enabled", "updated_by"}
			filteredCore := make(map[string]any)
			for _, field := range validCoreFields {
				if v, ok := core[field]; ok {
					filteredCore[field] = v
				}
			}
			if v, ok := filteredCore["rss_enabled"]; ok {
				switch val := v.(type) {
				case bool:
					if val {
						filteredCore["rss_enabled"] = 1
					} else {
						filteredCore["rss_enabled"] = 0
					}
				}
			}
			if len(filteredCore) > 0 {
				if err := tx.Table("t_blog_core_config").Where("id=1").Updates(filteredCore).Error; err != nil {
					return err
				}
			}
		}
		if len(s.UIConfig) > 0 {
			ui := convertKeysCamelToSnake(s.UIConfig)
			ui["updated_by"] = operator
			// 只保留数据库表中存在的字段，避免更新不存在的字段导致错误
			validUIFields := []string{"github_home", "csdn_home", "gitee_home", "zhihu_home", "github_show", "csdn_show", "gitee_show", "zhihu_show", "recommend_strategy", "search_engine", "hot_search_words", "hot_search_mode", "meilisearch_host", "meilisearch_api_key", "meilisearch_index", "updated_by"}
			filteredUI := make(map[string]any)
			for _, field := range validUIFields {
				if v, ok := ui[field]; ok {
					filteredUI[field] = v
				}
			}
			// hot_search_mode 是 tinyint(1)，兼容前端可能传入字符串的情况
			if v, ok := filteredUI["hot_search_mode"]; ok {
				switch val := v.(type) {
				case string:
					if val == "REAL" || val == "true" {
						filteredUI["hot_search_mode"] = 1
					} else {
						filteredUI["hot_search_mode"] = 0
					}
				case bool:
					if val {
						filteredUI["hot_search_mode"] = 1
					} else {
						filteredUI["hot_search_mode"] = 0
					}
				}
			}
			if len(filteredUI) > 0 {
				if err := tx.Table("t_blog_ui_config").Where("id=1").Updates(filteredUI).Error; err != nil {
					return err
				}
			}
		}
		if len(s.StorageConfig) > 0 {
			storage := convertKeysCamelToSnake(s.StorageConfig)
			storage["updated_by"] = operator
			if err := tx.Table("t_blog_storage_config").Where("id=1").Updates(storage).Error; err != nil {
				return err
			}
			dir, _ := storage["upload_local_dir"].(string)
			prefix, _ := storage["upload_local_url_prefix"].(string)
			r.setUploadStorageConfig(strings.TrimSpace(dir), strings.TrimSpace(prefix))
		}
		if len(s.EmailConfig) > 0 {
			b, err := json.Marshal(s.EmailConfig)
			if err != nil {
				return err
			}
			// id=1 单行配置
			if err := tx.Exec(`
INSERT INTO t_blog_email_config(id, config_json, updated_by)
VALUES(1, ?, ?)
ON DUPLICATE KEY UPDATE config_json = VALUES(config_json), updated_by = VALUES(updated_by), update_time = CURRENT_TIMESTAMP
`, string(b), operator).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
