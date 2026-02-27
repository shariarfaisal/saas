package contextkey

type contextKey string

const (
	TenantKey contextKey = "tenant"
	UserKey   contextKey = "user"
)
