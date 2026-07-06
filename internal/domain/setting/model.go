package setting

type BlogSettings struct {
	CoreConfig    map[string]any `json:"coreConfig"`
	UIConfig      map[string]any `json:"uiConfig"`
	StorageConfig map[string]any `json:"storageConfig"`
	EmailConfig   map[string]any `json:"emailConfig"`
}
