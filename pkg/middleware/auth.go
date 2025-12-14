package middleware

import (
	"context"
	"errors"
	"strings"

	"github.com/vantutran2k1/rwe/internal/tenancy"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type Authenticator interface {
	Authenticate(ctx context.Context, apiKey string) (*tenancy.Tenant, error)
}

func UnaryAuthInterceptor(auth Authenticator) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, errors.New("missing metadata")
		}

		authz := md.Get("authorization")
		if len(authz) == 0 {
			return nil, errors.New("missing authorization")
		}

		token := strings.TrimPrefix(authz[0], "Bearer ")
		tenant, err := auth.Authenticate(ctx, token)
		if err != nil {
			return nil, err
		}

		ctx = tenancy.NewContext(ctx, *tenant)
		return handler(ctx, req)
	}
}
