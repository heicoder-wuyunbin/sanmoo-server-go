package mysqlrepo

import (
	"context"
	"encoding/json"
	"errors"
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
	brandRaw := map[string]any{}
	channelRaw := map[string]any{}
	infraRaw := map[string]any{}
	privacyRaw := map[string]any{}

	if err := r.db.WithContext(ctx).Table("t_blog_brand_config").Where("id=1").Take(&brandRaw).Error; err != nil {
		return nil, err
	}
	if err := r.db.WithContext(ctx).Table("t_blog_channel_config").Where("id=1").Take(&channelRaw).Error; err != nil {
		return nil, err
	}
	if err := r.db.WithContext(ctx).Table("t_blog_infrastructure_config").Where("id=1").Take(&infraRaw).Error; err != nil {
		return nil, err
	}
	if err := r.db.WithContext(ctx).Table("t_blog_compliance_config").Where("id=1").Take(&privacyRaw).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		// t_blog_compliance_config 表可能没有初始数据，此时使用空 map 继续执行
		privacyRaw = map[string]any{}
	}

	brand := convertKeysSnakeToCamel(brandRaw)
	channel := convertKeysSnakeToCamel(channelRaw)
	infra := convertKeysSnakeToCamel(infraRaw)
	privacy := convertKeysSnakeToCamel(privacyRaw)

	// 组装 UIConfig = channel(social) + infra(search)
	uiConfig := map[string]any{}
	for k, v := range channel {
		uiConfig[k] = v
	}
	searchFieldsInUI := []string{"searchEngine", "recommendStrategy", "hotSearchWords", "hotSearchMode", "meilisearchHost", "meilisearchApiKey", "meilisearchIndex", "meilisearchLastSyncTime"}
	for _, field := range searchFieldsInUI {
		if v, ok := infra[field]; ok {
			uiConfig[field] = v
		}
	}

	// 组装 CoreConfig = brand + privacy
	coreConfig := map[string]any{}
	for k, v := range brand {
		coreConfig[k] = v
	}
	for k, v := range privacy {
		coreConfig[k] = v
	}

	// 从 infra 中提取存储字段
	storage := map[string]any{}
	storageFields := []string{"uploadStrategy", "uploadLocalDir", "uploadLocalUrlPrefix", "uploadQiniuBucket", "uploadQiniuDomain", "uploadQiniuRegion", "uploadQiniuAccessKey", "uploadQiniuSecretKey", "uploadAliyunEndpoint", "uploadAliyunBucket", "uploadAliyunDomain", "uploadAliyunAccessKey", "uploadAliyunSecretKey"}
	for _, field := range storageFields {
		if v, ok := infra[field]; ok {
			storage[field] = v
		}
	}

	// 从 infra 中提取邮件配置
	emailCfg := map[string]any{}
	if emailJSON, ok := infra["emailConfigJson"].(string); ok && emailJSON != "" {
		_ = json.Unmarshal([]byte(emailJSON), &emailCfg)
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

			brandFields := []string{"blog_name", "author", "introduction", "avatar", "site_url", "rss_enabled", "updated_by"}
			filteredBrand := filterValidFields(core, brandFields)
			if v, ok := filteredBrand["rss_enabled"]; ok {
				switch val := v.(type) {
				case bool:
					if val {
						filteredBrand["rss_enabled"] = 1
					} else {
						filteredBrand["rss_enabled"] = 0
					}
				}
			}

			if len(filteredBrand) > 0 {
				if err := tx.Table("t_blog_brand_config").Where("id=1").Updates(filteredBrand).Error; err != nil {
					return err
				}
			}
			if len(privacyPart) > 0 {
				if err := tx.Exec(`
INSERT INTO t_blog_compliance_config(id, privacy_policy, updated_by)
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

			socialFields := []string{"github_home", "csdn_home", "gitee_home", "zhihu_home", "github_show", "csdn_show", "gitee_show", "zhihu_show", "web_enabled", "mp_enabled", "updated_by"}
			searchFields := []string{"recommend_strategy", "search_engine", "hot_search_words", "hot_search_mode", "meilisearch_host", "meilisearch_api_key", "meilisearch_index", "updated_by"}

			filteredSocial := filterValidFields(ui, socialFields)
			for _, field := range []string{"github_show", "csdn_show", "gitee_show", "zhihu_show", "web_enabled", "mp_enabled"} {
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
				if err := tx.Table("t_blog_channel_config").Where("id=1").Updates(filteredSocial).Error; err != nil {
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
				if err := tx.Table("t_blog_infrastructure_config").Where("id=1").Updates(filteredSearch).Error; err != nil {
					return err
				}
			}
		}
		if len(s.StorageConfig) > 0 {
			storage := convertKeysCamelToSnake(s.StorageConfig)
			storage["updated_by"] = operator
			validStorageFields := []string{"upload_strategy", "upload_local_dir", "upload_local_url_prefix", "upload_qiniu_bucket", "upload_qiniu_domain", "upload_qiniu_region", "upload_qiniu_access_key", "upload_qiniu_secret_key", "upload_aliyun_endpoint", "upload_aliyun_bucket", "upload_aliyun_domain", "upload_aliyun_access_key", "upload_aliyun_secret_key", "updated_by"}
			filteredStorage := filterValidFields(storage, validStorageFields)
			if len(filteredStorage) > 0 {
				if err := tx.Table("t_blog_infrastructure_config").Where("id=1").Updates(filteredStorage).Error; err != nil {
					return err
				}
			}
			strategy, _ := storage["upload_strategy"].(string)
			dir, _ := storage["upload_local_dir"].(string)
			prefix, _ := storage["upload_local_url_prefix"].(string)
			qiniuBucket, _ := storage["upload_qiniu_bucket"].(string)
			qiniuDomain, _ := storage["upload_qiniu_domain"].(string)
			qiniuRegion, _ := storage["upload_qiniu_region"].(string)
			qiniuAccessKey, _ := storage["upload_qiniu_access_key"].(string)
			qiniuSecretKey, _ := storage["upload_qiniu_secret_key"].(string)
			aliyunEndpoint, _ := storage["upload_aliyun_endpoint"].(string)
			aliyunBucket, _ := storage["upload_aliyun_bucket"].(string)
			aliyunDomain, _ := storage["upload_aliyun_domain"].(string)
			aliyunAccessKey, _ := storage["upload_aliyun_access_key"].(string)
			aliyunSecretKey, _ := storage["upload_aliyun_secret_key"].(string)
			r.setUploadStorageConfig(strings.TrimSpace(strategy), strings.TrimSpace(dir), strings.TrimSpace(prefix), strings.TrimSpace(qiniuBucket), strings.TrimSpace(qiniuDomain), strings.TrimSpace(qiniuRegion), strings.TrimSpace(qiniuAccessKey), strings.TrimSpace(qiniuSecretKey), strings.TrimSpace(aliyunEndpoint), strings.TrimSpace(aliyunBucket), strings.TrimSpace(aliyunDomain), strings.TrimSpace(aliyunAccessKey), strings.TrimSpace(aliyunSecretKey))
		}
		if len(s.EmailConfig) > 0 {
			b, err := json.Marshal(s.EmailConfig)
			if err != nil {
				return err
			}
			if err := tx.Exec(`
INSERT INTO t_blog_infrastructure_config(id, email_config_json, updated_by)
VALUES(1, ?, ?)
ON DUPLICATE KEY UPDATE email_config_json = VALUES(email_config_json), updated_by = VALUES(updated_by), update_time = CURRENT_TIMESTAMP
`, string(b), operator).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *Repository) GetCoreConfig(ctx context.Context) (domsetting.CoreConfig, error) {
	brandRaw := map[string]any{}
	if err := r.db.WithContext(ctx).Table("t_blog_brand_config").Where("id=1").Take(&brandRaw).Error; err != nil {
		return nil, err
	}
	brand := convertKeysSnakeToCamel(brandRaw)
	delete(brand, "id")
	delete(brand, "configVersion")
	delete(brand, "createTime")
	delete(brand, "updateTime")
	delete(brand, "createdBy")
	delete(brand, "updatedBy")
	return domsetting.CoreConfig(brand), nil
}

func (r *Repository) UpdateCoreConfig(ctx context.Context, cfg domsetting.CoreConfig, operator string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		core := convertKeysCamelToSnake(map[string]any(cfg))
		core["updated_by"] = operator
		validBrandFields := []string{"blog_name", "author", "introduction", "avatar", "site_url", "rss_enabled", "contact_email", "updated_by"}
		filteredBrand := filterValidFields(core, validBrandFields)
		if v, ok := filteredBrand["rss_enabled"]; ok {
			switch val := v.(type) {
			case bool:
				if val {
					filteredBrand["rss_enabled"] = 1
				} else {
					filteredBrand["rss_enabled"] = 0
				}
			}
		}
		if len(filteredBrand) > 0 {
			if err := tx.Table("t_blog_brand_config").Where("id=1").Updates(filteredBrand).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *Repository) GetSocialConfig(ctx context.Context) (domsetting.SocialConfig, error) {
	channelRaw := map[string]any{}
	if err := r.db.WithContext(ctx).Table("t_blog_channel_config").Where("id=1").Take(&channelRaw).Error; err != nil {
		return nil, err
	}
	channel := convertKeysSnakeToCamel(channelRaw)
	delete(channel, "id")
	delete(channel, "configVersion")
	delete(channel, "createTime")
	delete(channel, "updateTime")
	delete(channel, "createdBy")
	delete(channel, "updatedBy")
	return domsetting.SocialConfig(channel), nil
}

func (r *Repository) UpdateSocialConfig(ctx context.Context, cfg domsetting.SocialConfig, operator string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		social := convertKeysCamelToSnake(map[string]any(cfg))
		social["updated_by"] = operator
		validSocialFields := []string{"github_home", "csdn_home", "gitee_home", "zhihu_home", "github_show", "csdn_show", "gitee_show", "zhihu_show", "web_enabled", "mp_enabled", "updated_by"}
		filteredSocial := filterValidFields(social, validSocialFields)
		for _, field := range []string{"github_show", "csdn_show", "gitee_show", "zhihu_show", "web_enabled", "mp_enabled"} {
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
			if err := tx.Table("t_blog_channel_config").Where("id=1").Updates(filteredSocial).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *Repository) GetSearchConfig(ctx context.Context) (domsetting.SearchConfig, error) {
	infraRaw := map[string]any{}
	if err := r.db.WithContext(ctx).Table("t_blog_infrastructure_config").Where("id=1").Take(&infraRaw).Error; err != nil {
		return nil, err
	}
	infra := convertKeysSnakeToCamel(infraRaw)
	// 只保留搜索相关字段
	search := map[string]any{}
	searchFields := []string{"searchEngine", "recommendStrategy", "hotSearchWords", "hotSearchMode", "meilisearchHost", "meilisearchApiKey", "meilisearchIndex", "meilisearchLastSyncTime"}
	for _, field := range searchFields {
		if v, ok := infra[field]; ok {
			search[field] = v
		}
	}
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
			if err := tx.Table("t_blog_infrastructure_config").Where("id=1").Updates(filteredSearch).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *Repository) GetPrivacyConfig(ctx context.Context) (domsetting.PrivacyConfig, error) {
	privacyRaw := map[string]any{}
	if err := r.db.WithContext(ctx).Table("t_blog_compliance_config").Where("id=1").Take(&privacyRaw).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		// t_blog_compliance_config 表可能没有初始数据，返回空配置
		return domsetting.PrivacyConfig{}, nil
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
		validPrivacyFields := []string{"privacy_policy", "filing_info", "contact_info", "data_retention_policy", "account_deletion_guide"}
		filteredPrivacy := filterValidFields(privacy, validPrivacyFields)
		if len(filteredPrivacy) > 0 {
			filteredPrivacy["updated_by"] = operator
			result := tx.Table("t_blog_compliance_config").Where("id=1").Updates(filteredPrivacy)
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected == 0 {
				filteredPrivacy["id"] = 1
				if err := tx.Table("t_blog_compliance_config").Create(filteredPrivacy).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func (r *Repository) GetStorageConfig(ctx context.Context) (domsetting.StorageConfig, error) {
	infraRaw := map[string]any{}
	if err := r.db.WithContext(ctx).Table("t_blog_infrastructure_config").Where("id=1").Take(&infraRaw).Error; err != nil {
		return nil, err
	}
	infra := convertKeysSnakeToCamel(infraRaw)
	// 只保留存储相关字段
	storage := map[string]any{}
	storageFields := []string{"uploadStrategy", "uploadLocalDir", "uploadLocalUrlPrefix", "uploadQiniuBucket", "uploadQiniuDomain", "uploadQiniuRegion", "uploadQiniuAccessKey", "uploadQiniuSecretKey", "uploadAliyunEndpoint", "uploadAliyunBucket", "uploadAliyunDomain", "uploadAliyunAccessKey", "uploadAliyunSecretKey"}
	for _, field := range storageFields {
		if v, ok := infra[field]; ok {
			storage[field] = v
		}
	}

	return domsetting.StorageConfig(storage), nil
}

// LoadStorageConfig 从数据库加载存储配置到内存，启动时调用
func (r *Repository) LoadStorageConfig(ctx context.Context) error {
	cfg, err := r.GetStorageConfig(ctx)
	if err != nil {
		return err
	}
	strategy, _ := cfg["uploadStrategy"].(string)
	dir, _ := cfg["uploadLocalDir"].(string)
	prefix, _ := cfg["uploadLocalUrlPrefix"].(string)
	qiniuBucket, _ := cfg["uploadQiniuBucket"].(string)
	qiniuDomain, _ := cfg["uploadQiniuDomain"].(string)
	qiniuRegion, _ := cfg["uploadQiniuRegion"].(string)
	qiniuAccessKey, _ := cfg["uploadQiniuAccessKey"].(string)
	qiniuSecretKey, _ := cfg["uploadQiniuSecretKey"].(string)
	aliyunEndpoint, _ := cfg["uploadAliyunEndpoint"].(string)
	aliyunBucket, _ := cfg["uploadAliyunBucket"].(string)
	aliyunDomain, _ := cfg["uploadAliyunDomain"].(string)
	aliyunAccessKey, _ := cfg["uploadAliyunAccessKey"].(string)
	aliyunSecretKey, _ := cfg["uploadAliyunSecretKey"].(string)
	r.setUploadStorageConfig(
		strings.TrimSpace(strategy),
		strings.TrimSpace(dir),
		strings.TrimSpace(prefix),
		strings.TrimSpace(qiniuBucket),
		strings.TrimSpace(qiniuDomain),
		strings.TrimSpace(qiniuRegion),
		strings.TrimSpace(qiniuAccessKey),
		strings.TrimSpace(qiniuSecretKey),
		strings.TrimSpace(aliyunEndpoint),
		strings.TrimSpace(aliyunBucket),
		strings.TrimSpace(aliyunDomain),
		strings.TrimSpace(aliyunAccessKey),
		strings.TrimSpace(aliyunSecretKey),
	)
	return nil
}

func (r *Repository) UpdateStorageConfig(ctx context.Context, cfg domsetting.StorageConfig, operator string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		storage := convertKeysCamelToSnake(map[string]any(cfg))
		storage["updated_by"] = operator
		validStorageFields := []string{"upload_strategy", "upload_local_dir", "upload_local_url_prefix", "upload_qiniu_bucket", "upload_qiniu_domain", "upload_qiniu_region", "upload_qiniu_access_key", "upload_qiniu_secret_key", "upload_aliyun_endpoint", "upload_aliyun_bucket", "upload_aliyun_domain", "upload_aliyun_access_key", "upload_aliyun_secret_key", "updated_by"}
		filteredStorage := filterValidFields(storage, validStorageFields)
		if len(filteredStorage) > 0 {
			if err := tx.Table("t_blog_infrastructure_config").Where("id=1").Updates(filteredStorage).Error; err != nil {
				return err
			}
		}
		strategy, _ := storage["upload_strategy"].(string)
		dir, _ := storage["upload_local_dir"].(string)
		prefix, _ := storage["upload_local_url_prefix"].(string)
		qiniuBucket, _ := storage["upload_qiniu_bucket"].(string)
		qiniuDomain, _ := storage["upload_qiniu_domain"].(string)
		qiniuRegion, _ := storage["upload_qiniu_region"].(string)
		qiniuAccessKey, _ := storage["upload_qiniu_access_key"].(string)
		qiniuSecretKey, _ := storage["upload_qiniu_secret_key"].(string)
		aliyunEndpoint, _ := storage["upload_aliyun_endpoint"].(string)
		aliyunBucket, _ := storage["upload_aliyun_bucket"].(string)
		aliyunDomain, _ := storage["upload_aliyun_domain"].(string)
		aliyunAccessKey, _ := storage["upload_aliyun_access_key"].(string)
		aliyunSecretKey, _ := storage["upload_aliyun_secret_key"].(string)
		r.setUploadStorageConfig(strings.TrimSpace(strategy), strings.TrimSpace(dir), strings.TrimSpace(prefix), strings.TrimSpace(qiniuBucket), strings.TrimSpace(qiniuDomain), strings.TrimSpace(qiniuRegion), strings.TrimSpace(qiniuAccessKey), strings.TrimSpace(qiniuSecretKey), strings.TrimSpace(aliyunEndpoint), strings.TrimSpace(aliyunBucket), strings.TrimSpace(aliyunDomain), strings.TrimSpace(aliyunAccessKey), strings.TrimSpace(aliyunSecretKey))
		return nil
	})
}

func (r *Repository) GetEmailConfig(ctx context.Context) (domsetting.EmailConfig, error) {
	emailCfg := map[string]any{}
	var infraRow struct {
		EmailConfigJSON string `gorm:"column:email_config_json"`
	}
	if err := r.db.WithContext(ctx).Table("t_blog_infrastructure_config").Select("email_config_json").Where("id=1").Take(&infraRow).Error; err != nil {
		return nil, err
	}
	_ = json.Unmarshal([]byte(infraRow.EmailConfigJSON), &emailCfg)
	return domsetting.EmailConfig(emailCfg), nil
}

func (r *Repository) UpdateEmailConfig(ctx context.Context, cfg domsetting.EmailConfig, operator string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		b, err := json.Marshal(cfg)
		if err != nil {
			return err
		}
		if err := tx.Exec(`
INSERT INTO t_blog_infrastructure_config(id, email_config_json, updated_by)
VALUES(1, ?, ?)
ON DUPLICATE KEY UPDATE email_config_json = VALUES(email_config_json), updated_by = VALUES(updated_by), update_time = CURRENT_TIMESTAMP
`, string(b), operator).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *Repository) GetWechatConfig(ctx context.Context) (domsetting.WechatConfig, error) {
	infraRaw := map[string]any{}
	if err := r.db.WithContext(ctx).Table("t_blog_infrastructure_config").Where("id=1").Take(&infraRaw).Error; err != nil {
		return nil, err
	}
	infra := convertKeysSnakeToCamel(infraRaw)
	wechat := map[string]any{}
	wechatFields := []string{"wxDevAppId", "wxDevAppSecret", "wxProdAppId", "wxProdAppSecret", "wxEnvMode"}
	for _, field := range wechatFields {
		if v, ok := infra[field]; ok {
			wechat[field] = v
		}
	}
	return domsetting.WechatConfig(wechat), nil
}

func (r *Repository) UpdateWechatConfig(ctx context.Context, cfg domsetting.WechatConfig, operator string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		wechat := convertKeysCamelToSnake(map[string]any(cfg))
		wechat["updated_by"] = operator
		validWechatFields := []string{"wx_dev_app_id", "wx_dev_app_secret", "wx_prod_app_id", "wx_prod_app_secret", "wx_env_mode", "updated_by"}
		filteredWechat := filterValidFields(wechat, validWechatFields)
		// 处理 wx_env_mode 布尔值转换
		if v, ok := filteredWechat["wx_env_mode"]; ok {
			switch val := v.(type) {
			case bool:
				if val {
					filteredWechat["wx_env_mode"] = 1
				} else {
					filteredWechat["wx_env_mode"] = 0
				}
			}
		}
		if len(filteredWechat) > 0 {
			if err := tx.Table("t_blog_infrastructure_config").Where("id=1").Updates(filteredWechat).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *Repository) SaveMeiliSearchSyncTime(ctx context.Context, syncTime string) error {
	return r.db.WithContext(ctx).Table("t_blog_infrastructure_config").Where("id=1").Update("meilisearch_last_sync_time", syncTime).Error
}
