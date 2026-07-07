package setting

type BlogSettings struct {
	CoreConfig    map[string]any `json:"coreConfig"`
	UIConfig      map[string]any `json:"uiConfig"`
	StorageConfig map[string]any `json:"storageConfig"`
	EmailConfig   map[string]any `json:"emailConfig"`
}

type CoreConfig map[string]any
type SocialConfig map[string]any
type SearchConfig map[string]any
type PrivacyConfig map[string]any
type StorageConfig map[string]any
type EmailConfig map[string]any
