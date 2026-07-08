package setting

import "context"

type Repository interface {
	Get(ctx context.Context) (*BlogSettings, error)
	Update(ctx context.Context, s *BlogSettings, operator string) error

	GetCoreConfig(ctx context.Context) (CoreConfig, error)
	UpdateCoreConfig(ctx context.Context, cfg CoreConfig, operator string) error

	GetSocialConfig(ctx context.Context) (SocialConfig, error)
	UpdateSocialConfig(ctx context.Context, cfg SocialConfig, operator string) error

	GetSearchConfig(ctx context.Context) (SearchConfig, error)
	UpdateSearchConfig(ctx context.Context, cfg SearchConfig, operator string) error

	GetPrivacyConfig(ctx context.Context) (PrivacyConfig, error)
	UpdatePrivacyConfig(ctx context.Context, cfg PrivacyConfig, operator string) error

	GetStorageConfig(ctx context.Context) (StorageConfig, error)
	UpdateStorageConfig(ctx context.Context, cfg StorageConfig, operator string) error

	GetEmailConfig(ctx context.Context) (EmailConfig, error)
	UpdateEmailConfig(ctx context.Context, cfg EmailConfig, operator string) error

	SaveMeiliSearchSyncTime(ctx context.Context, syncTime string) error
}
