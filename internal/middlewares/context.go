package middlewares

import (
	"context"

	"github.com/vantutran2k1/rwe/internal/auth"
)

type contextKey string

const (
	authorizationHeader = "authorization"
	authorizationBearer = "bearer"
	payloadContextKey   = contextKey("authorization_payload")
)

func GetTokenPayload(ctx context.Context) *auth.TokenPayload {
	payload, ok := ctx.Value(payloadContextKey).(*auth.TokenPayload)
	if !ok {
		return nil
	}

	return payload
}
