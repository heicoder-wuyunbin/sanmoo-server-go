package setting

import "context"

type Repository interface {
	Get(ctx context.Context) (*BlogSettings, error)
	Update(ctx context.Context, s *BlogSettings, operator string) error
}
