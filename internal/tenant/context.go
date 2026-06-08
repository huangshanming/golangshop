package tenant

import "context"

type ctxKey string

const tenantIDKey ctxKey = "tenant_id"

func WithTenantID(ctx context.Context, tenantID uint64) context.Context {
	return context.WithValue(ctx, tenantIDKey, tenantID)
}

func IDFromContext(ctx context.Context) (uint64, bool) {
	v, ok := ctx.Value(tenantIDKey).(uint64)
	return v, ok
}
