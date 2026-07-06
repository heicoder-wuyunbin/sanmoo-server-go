package auth

type RefreshStore interface {
	Save(userID uint64, refreshToken string) error
	Validate(userID uint64, refreshToken string) bool
}
