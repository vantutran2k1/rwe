package token

import (
	"context"
)

type contextKey string

const (
	AuthorizationHeader = "authorization"
	AuthorizationBearer = "bearer"
	PayloadContextKey   = contextKey("authorization_payload")
)

func GetTokenPayload(ctx context.Context) *Payload {
	payload, ok := ctx.Value(PayloadContextKey).(*Payload)
	if !ok {
		return nil
	}

	return payload
}
