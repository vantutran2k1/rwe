package tenancy

import "context"

type ctxKey struct{}

type Tenant struct {
	ID   string
	Name string
}

func NewContext(ctx context.Context, t Tenant) context.Context {
	return context.WithValue(ctx, ctxKey{}, t)
}

func FromContext(ctx context.Context) (Tenant, bool) {
	t, ok := ctx.Value(ctxKey{}).(Tenant)
	return t, ok
}
