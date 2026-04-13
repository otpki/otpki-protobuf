package authmap

import (
	"context"
)

var AuthRequiredByProcedure = map[string]bool{}

type authRequiredCtxKey struct{}

func WithAuthRequired(ctx context.Context, required bool) context.Context {
	return context.WithValue(ctx, authRequiredCtxKey{}, required)
}

func AuthRequiredFromContext(ctx context.Context) (bool, bool) {
	v := ctx.Value(authRequiredCtxKey{})
	if v == nil {
		return false, false
	}
	required, ok := v.(bool)
	return required, ok
}

func IsAuthRequiredProcedure(procedure string) bool {
	required, ok := AuthRequiredByProcedure[procedure]
	if !ok {
		return true
	}
	return required
}
