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

func filterValidFields(m map[string]any, validFields []string) map[string]any {
	res := make(map[string]any)
	for _, field := range validFields {
		if v, ok := m[field]; ok {
			res[field] = v
		}
	}
	return res
}

func (r *Repository) Get(ctx context.Context) (*domsetting.BlogSettings, error) {
	coreRaw := map[string]any{}
	socialRaw := map[string]any{}
	searchRaw := map[string]any{}
	privacyRaw := map[string]any{}
	storageRaw := map[string]any{}
	emailCfg := map[string]any{}

	if err := r.db.WithContext(ctx).Table("t_blog_core_config").Where("id=1").Take(&coreRaw).Error; err != nil {
		return nil, err
	}
	if err := r.db.WithContext(ctx).Table("t_blog_social_config").Where("id=1").Take(&socialRaw).Error; err != nil {
		return nil, err
	}
	if err := r.db.WithContext(ctx).Table("t_blog_search_config").Where("id=1").Take(&searchRaw).Error; err != nil {
		return nil, err
	}
	if err := r.db.WithContext(ctx).Table("t_blog_privacy_config").Where("id=1").Take(&privacyRaw).Error; err != nil {
		return nil, err
	}
	if err := r.db.WithContext(ctx).Table("t_blog_storage_config").Where("id=1").Take(&storageRaw).Error; err != nil {
		return nil, err
	}
	var emailRow struct {
		ConfigJSON string `gorm:"column:config_json"`
	}
	if err := r.db.WithContext(ctx).Table("t_blog_email_config").Select("config_json").Where("id=1").Take(&emailRow).Error; err == nil {
		_ = json.Unmarshal([]byte(emailRow.ConfigJSON), &emailCfg)
	}

	core := convertKeysSnakeToCamel(coreRaw)
	social := convertKeysSnakeToCamel(socialRaw)
	search := convertKeysSnakeToCamel(searchRaw)
	privacy := convertKeysSnakeToCamel(privacyRaw)
	storage := convertKeysSnakeToCamel(storageRaw)

	uiConfig := map[string]any{}
	for k, v := range social {
		uiConfig[k] = v
	}
	for k, v := range search {
		uiConfig[k] = v
	}

	coreConfig := map[string]any{}
	for k, v := range core {
		coreConfig[k] = v
	}
	for k, v := range privacy {
		coreConfig[k] = v
	}

	return &domsetting.BlogSettings{
		CoreConfig:    coreConfig,
		UIConfig:      uiConfig,
		StorageConfig: storage,
		EmailConfig:   emailCfg,
	}, nil
}

func (r *Repository) Update(ctx context.Context, s *domsetting.BlogSettings, operator string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if len(s.CoreConfig) > 0 {
			core := convertKeysCamelToSnake(s.CoreConfig)
			core["updated_by"] = operator

			privacyPart := map[string]any{}
			if v, ok := core["privacy_policy"]; ok {
				privacyPart["privacy_policy"] = v
				privacyPart["updated_by"] = operator
			}

			coreFields := []string{"blog_name", "author", "introduction", "avatar", "rss_enabled", "updated_by"}
			filteredCore := filterValidFields(core, coreFields)
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
			if len(privacyPart) > 0 {
				if err := tx.Exec(`
INSERT INTO t_blog_privacy_config(id, privacy_policy, updated_by)
VALUES(1, ?, ?)
ON DUPLICATE KEY UPDATE privacy_policy = VALUES(privacy_policy), updated_by = VALUES(updated_by), update_time = CURRENT_TIMESTAMP
`, privacyPart["privacy_policy"], operator).Error; err != nil {
					return err
				}
			}
		}
		if len(s.UIConfig) > 0 {
			ui := convertKeysCamelToSnake(s.UIConfig)
			ui["updated_by"] = operator

			socialFields := []string{"github_home", "csdn_home", "gitee_home", "zhihu_home", "github_show", "csdn_show", "gitee_show", "zhihu_show", "updated_by"}
			searchFields := []string{"recommend_strategy", "search_engine", "hot_search_words", "hot_search_mode", "meilisearch_host", "meilisearch_api_key", "meilisearch_index", "updated_by"}
			allUIFields := append(append([]string{}, socialFields...), searchFields...)

			filteredAll := filterValidFields(ui, allUIFields)
			if len(filteredAll) > 0 {
				if err := tx.Table("t_blog_ui_config").Where("id=1").Updates(filteredAll).Error; err != nil {
					return err
				}
			}

			filteredSocial := filterValidFields(ui, socialFields)
			for _, field := range []string{"github_show", "csdn_show", "gitee_show", "zhihu_show"} {
				if v, ok := filteredSocial[field]; ok {
					switch val := v.(type) {
					case bool:
						if val {
							filteredSocial[field] = 1
						} else {
							filteredSocial[field] = 0
						}
					}
				}
			}
			if len(filteredSocial) > 0 {
				if err := tx.Table("t_blog_social_config").Where("id=1").Updates(filteredSocial).Error; err != nil {
					return err
				}
			}

			filteredSearch := filterValidFields(ui, searchFields)
			if v, ok := filteredSearch["hot_search_mode"]; ok {
				switch val := v.(type) {
				case string:
					if val == "REAL" || val == "true" {
						filteredSearch["hot_search_mode"] = 1
					} else {
						filteredSearch["hot_search_mode"] = 0
					}
				case bool:
					if val {
						filteredSearch["hot_search_mode"] = 1
					} else {
						filteredSearch["hot_search_mode"] = 0
					}
				}
			}
			if len(filteredSearch) > 0 {
				if err := tx.Table("t_blog_search_config").Where("id=1").Updates(filteredSearch).Error; err != nil {
					return err
				}
			}
		}
		if len(s.StorageConfig) > 0 {
			storage := convertKeysCamelToSnake(s.StorageConfig)
			storage["updated_by"] = operator
			validStorageFields := []string{"upload_strategy", "upload_local_dir", "upload_local_url_prefix", "upload_qiniu_bucket", "upload_qiniu_domain", "upload_qiniu_access_key", "upload_qiniu_secret_key", "upload_aliyun_endpoint", "upload_aliyun_bucket", "upload_aliyun_domain", "config_version", "updated_by"}
			filteredStorage := make(map[string]any)
			for _, field := range validStorageFields {
				if v, ok := storage[field]; ok {
					filteredStorage[field] = v
				}
			}
			if len(filteredStorage) > 0 {
				if err := tx.Table("t_blog_storage_config").Where("id=1").Updates(filteredStorage).Error; err != nil {
					return err
				}
			}
			strategy, _ := storage["upload_strategy"].(string)
			dir, _ := storage["upload_local_dir"].(string)
			prefix, _ := storage["upload_local_url_prefix"].(string)
			qiniuBucket, _ := storage["upload_qiniu_bucket"].(string)
			qiniuDomain, _ := storage["upload_qiniu_domain"].(string)
			qiniuAccessKey, _ := storage["upload_qiniu_access_key"].(string)
			qiniuSecretKey, _ := storage["upload_qiniu_secret_key"].(string)
			r.setUploadStorageConfig(strings.TrimSpace(strategy), strings.TrimSpace(dir), strings.TrimSpace(prefix), strings.TrimSpace(qiniuBucket), strings.TrimSpace(qiniuDomain), strings.TrimSpace(qiniuAccessKey), strings.TrimSpace(qiniuSecretKey))
		}
		if len(s.EmailConfig) > 0 {
			b, err := json.Marshal(s.EmailConfig)
			if err != nil {
				return err
			}
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

func (r *Repository) GetCoreConfig(ctx context.Context) (domsetting.CoreConfig, error) {
	coreRaw := map[string]any{}
	if err := r.db.WithContext(ctx).Table("t_blog_core_config").Where("id=1").Take(&coreRaw).Error; err != nil {
		return nil, err
	}
	core := convertKeysSnakeToCamel(coreRaw)
	delete(core, "privacyPolicy")
	delete(core, "id")
	delete(core, "configVersion")
	delete(core, "createTime")
	delete(core, "updateTime")
	delete(core, "createdBy")
	delete(core, "updatedBy")
	return domsetting.CoreConfig(core), nil
}

func (r *Repository) UpdateCoreConfig(ctx context.Context, cfg domsetting.CoreConfig, operator string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		core := convertKeysCamelToSnake(map[string]any(cfg))
		core["updated_by"] = operator
		validCoreFields := []string{"blog_name", "author", "introduction", "avatar", "rss_enabled", "updated_by"}
		filteredCore := filterValidFields(core, validCoreFields)
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
		return nil
	})
}

func (r *Repository) GetSocialConfig(ctx context.Context) (domsetting.SocialConfig, error) {
	socialRaw := map[string]any{}
	if err := r.db.WithContext(ctx).Table("t_blog_social_config").Where("id=1").Take(&socialRaw).Error; err != nil {
		return nil, err
	}
	social := convertKeysSnakeToCamel(socialRaw)
	delete(social, "id")
	delete(social, "configVersion")
	delete(social, "createTime")
	delete(social, "updateTime")
	delete(social, "createdBy")
	delete(social, "updatedBy")
	return domsetting.SocialConfig(social), nil
}

func (r *Repository) UpdateSocialConfig(ctx context.Context, cfg domsetting.SocialConfig, operator string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		social := convertKeysCamelToSnake(map[string]any(cfg))
		social["updated_by"] = operator
		validSocialFields := []string{"github_home", "csdn_home", "gitee_home", "zhihu_home", "github_show", "csdn_show", "gitee_show", "zhihu_show", "updated_by"}
		filteredSocial := filterValidFields(social, validSocialFields)
		for _, field := range []string{"github_show", "csdn_show", "gitee_show", "zhihu_show"} {
			if v, ok := filteredSocial[field]; ok {
				switch val := v.(type) {
				case bool:
					if val {
						filteredSocial[field] = 1
					} else {
						filteredSocial[field] = 0
					}
				}
			}
		}
		if len(filteredSocial) > 0 {
			if err := tx.Table("t_blog_social_config").Where("id=1").Updates(filteredSocial).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *Repository) GetSearchConfig(ctx context.Context) (domsetting.SearchConfig, error) {
	searchRaw := map[string]any{}
	if err := r.db.WithContext(ctx).Table("t_blog_search_config").Where("id=1").Take(&searchRaw).Error; err != nil {
		return nil, err
	}
	search := convertKeysSnakeToCamel(searchRaw)
	delete(search, "id")
	delete(search, "configVersion")
	delete(search, "createTime")
	delete(search, "updateTime")
	delete(search, "createdBy")
	delete(search, "updatedBy")
	return domsetting.SearchConfig(search), nil
}

func (r *Repository) UpdateSearchConfig(ctx context.Context, cfg domsetting.SearchConfig, operator string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		search := convertKeysCamelToSnake(map[string]any(cfg))
		search["updated_by"] = operator
		validSearchFields := []string{"recommend_strategy", "search_engine", "hot_search_mode", "hot_search_words", "meilisearch_host", "meilisearch_api_key", "meilisearch_index", "updated_by"}
		filteredSearch := filterValidFields(search, validSearchFields)
		if v, ok := filteredSearch["hot_search_mode"]; ok {
			switch val := v.(type) {
			case string:
				if val == "REAL" || val == "true" {
					filteredSearch["hot_search_mode"] = 1
				} else {
					filteredSearch["hot_search_mode"] = 0
				}
			case bool:
				if val {
					filteredSearch["hot_search_mode"] = 1
				} else {
					filteredSearch["hot_search_mode"] = 0
				}
			}
		}
		if len(filteredSearch) > 0 {
			if err := tx.Table("t_blog_search_config").Where("id=1").Updates(filteredSearch).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *Repository) GetPrivacyConfig(ctx context.Context) (domsetting.PrivacyConfig, error) {
	privacyRaw := map[string]any{}
	if err := r.db.WithContext(ctx).Table("t_blog_privacy_config").Where("id=1").Take(&privacyRaw).Error; err != nil {
		return nil, err
	}
	privacy := convertKeysSnakeToCamel(privacyRaw)
	delete(privacy, "id")
	delete(privacy, "configVersion")
	delete(privacy, "createTime")
	delete(privacy, "updateTime")
	delete(privacy, "createdBy")
	delete(privacy, "updatedBy")
	return domsetting.PrivacyConfig(privacy), nil
}

func (r *Repository) UpdatePrivacyConfig(ctx context.Context, cfg domsetting.PrivacyConfig, operator string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		privacy := convertKeysCamelToSnake(map[string]any(cfg))
		privacy["updated_by"] = operator
		validPrivacyFields := []string{"privacy_policy", "updated_by"}
		filteredPrivacy := filterValidFields(privacy, validPrivacyFields)
		if len(filteredPrivacy) > 0 {
			if err := tx.Table("t_blog_privacy_config").Where("id=1").Updates(filteredPrivacy).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *Repository) GetStorageConfig(ctx context.Context) (domsetting.StorageConfig, error) {
	storageRaw := map[string]any{}
	if err := r.db.WithContext(ctx).Table("t_blog_storage_config").Where("id=1").Take(&storageRaw).Error; err != nil {
		return nil, err
	}
	storage := convertKeysSnakeToCamel(storageRaw)
	delete(storage, "id")
	delete(storage, "configVersion")
	delete(storage, "createTime")
	delete(storage, "updateTime")
	delete(storage, "createdBy")
	delete(storage, "updatedBy")
	return domsetting.StorageConfig(storage), nil
}

func (r *Repository) UpdateStorageConfig(ctx context.Context, cfg domsetting.StorageConfig, operator string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		storage := convertKeysCamelToSnake(map[string]any(cfg))
		storage["updated_by"] = operator
		validStorageFields := []string{"upload_strategy", "upload_local_dir", "upload_local_url_prefix", "upload_qiniu_bucket", "upload_qiniu_domain", "upload_qiniu_access_key", "upload_qiniu_secret_key", "upload_aliyun_endpoint", "upload_aliyun_bucket", "upload_aliyun_domain", "config_version", "updated_by"}
		filteredStorage := filterValidFields(storage, validStorageFields)
		if len(filteredStorage) > 0 {
			if err := tx.Table("t_blog_storage_config").Where("id=1").Updates(filteredStorage).Error; err != nil {
				return err
			}
		}
		strategy, _ := storage["upload_strategy"].(string)
		dir, _ := storage["upload_local_dir"].(string)
		prefix, _ := storage["upload_local_url_prefix"].(string)
		qiniuBucket, _ := storage["upload_qiniu_bucket"].(string)
		qiniuDomain, _ := storage["upload_qiniu_domain"].(string)
		qiniuAccessKey, _ := storage["upload_qiniu_access_key"].(string)
		qiniuSecretKey, _ := storage["upload_qiniu_secret_key"].(string)
		r.setUploadStorageConfig(strings.TrimSpace(strategy), strings.TrimSpace(dir), strings.TrimSpace(prefix), strings.TrimSpace(qiniuBucket), strings.TrimSpace(qiniuDomain), strings.TrimSpace(qiniuAccessKey), strings.TrimSpace(qiniuSecretKey))
		return nil
	})
}

func (r *Repository) GetEmailConfig(ctx context.Context) (domsetting.EmailConfig, error) {
	emailCfg := map[string]any{}
	var emailRow struct {
		ConfigJSON string `gorm:"column:config_json"`
	}
	if err := r.db.WithContext(ctx).Table("t_blog_email_config").Select("config_json").Where("id=1").Take(&emailRow).Error; err != nil {
		return nil, err
	}
	_ = json.Unmarshal([]byte(emailRow.ConfigJSON), &emailCfg)
	return domsetting.EmailConfig(emailCfg), nil
}

func (r *Repository) UpdateEmailConfig(ctx context.Context, cfg domsetting.EmailConfig, operator string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		b, err := json.Marshal(cfg)
		if err != nil {
			return err
		}
		if err := tx.Exec(`
INSERT INTO t_blog_email_config(id, config_json, updated_by)
VALUES(1, ?, ?)
ON DUPLICATE KEY UPDATE config_json = VALUES(config_json), updated_by = VALUES(updated_by), update_time = CURRENT_TIMESTAMP
`, string(b), operator).Error; err != nil {
			return err
		}
		return nil
	})
}
