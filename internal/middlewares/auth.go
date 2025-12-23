package middlewares

import (
	"context"
	"strings"

	"github.com/vantutran2k1/rwe/internal/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type AuthInterceptor struct {
	tokenMaker auth.TokenMaker
	// TODO: add RBAC
}

func NewAuthInterceptor(tokenMaker auth.TokenMaker) *AuthInterceptor {
	return &AuthInterceptor{
		tokenMaker: tokenMaker,
	}
}

func (i *AuthInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		if isPublicEndpoint(info.FullMethod) {
			return handler(ctx, req)
		}

		payload, err := i.authorize(ctx)
		if err != nil {
			return nil, err
		}

		newCtx := context.WithValue(ctx, payloadContextKey, payload)

		return handler(newCtx, req)
	}
}

func (i *AuthInterceptor) authorize(ctx context.Context) (*auth.TokenPayload, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing metadata")
	}

	values := md.Get(authorizationHeader)
	if len(values) == 0 {
		return nil, status.Error(codes.Unauthenticated, "authorization header is required")
	}

	authHeader := values[0]
	fields := strings.Fields(authHeader)
	if len(fields) < 2 {
		return nil, status.Error(codes.Unauthenticated, "invalid authorization header")
	}

	authType := strings.ToLower(fields[0])
	if authType != authorizationBearer {
		return nil, status.Errorf(codes.Unauthenticated, "unsupported authorization type: %s", authType)
	}

	accessToken := fields[1]
	payload, err := i.tokenMaker.VerifyToken(accessToken)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid access token: %v", err)
	}

	return payload, nil
}

func isPublicEndpoint(method string) bool {
	publicPaths := map[string]bool{
		"/auth.v1.AuthService/Login":    true,
		"/auth.v1.AuthService/Register": true,
		"/grpc.health.v1.Health/Check":  true,
	}

	return publicPaths[method]
}
