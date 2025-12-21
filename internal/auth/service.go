package auth

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	authv1 "github.com/vantutran2k1/rwe/gen/go/auth/v1"
	sqlc "github.com/vantutran2k1/rwe/internal/auth/db"
	"github.com/vantutran2k1/rwe/internal/common/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Service struct {
	pool       *pgxpool.Pool
	querier    sqlc.Querier
	tokenMaker TokenMaker
	authv1.UnimplementedAuthServiceServer
}

func NewService(pool *pgxpool.Pool, tokenKey string, tokenDurationHours int32) *Service {
	duration := time.Duration(tokenDurationHours) * time.Hour
	tokenMaker, _ := NewPasetoMaker(tokenKey, duration)

	return &Service{
		pool:       pool,
		querier:    sqlc.New(pool),
		tokenMaker: tokenMaker,
	}
}

func (s *Service) ValidateApiKey(ctx context.Context, req *authv1.ValidateApiKeyRequest) (*authv1.ValidateApiKeyResponse, error) {
	if req.ApiKey == "" || !strings.HasPrefix(req.ApiKey, keyPrefix) {
		return &authv1.ValidateApiKeyResponse{Valid: false}, nil
	}

	inputHash := HashKey(req.ApiKey)

	key, err := s.querier.GetApiKeyByHash(ctx, inputHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &authv1.ValidateApiKeyResponse{Valid: false}, nil
		}
		return nil, status.Errorf(codes.Internal, "validation error: %v", err)
	}

	if key.Revoked.Bool {
		return &authv1.ValidateApiKeyResponse{Valid: false}, nil
	}

	return &authv1.ValidateApiKeyResponse{
		Valid:    true,
		TenantId: utils.PgUUIDToString(key.TenantID),
	}, nil
}

func (s *Service) IssueApiKey(ctx context.Context, req *authv1.IssueApiKeyRequest) (*authv1.IssueApiKeyResponse, error) {
	rawKey, hashedKey, prefix, err := GenerateAPIKey()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate key: %v", err)
	}

	tenantId, err := uuid.Parse(req.TenantId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant id: %v", err)
	}

	var row sqlc.CreateApiKeyRow
	if err := s.execTx(ctx, func(querier sqlc.Querier) error {
		row, err = s.querier.CreateApiKey(ctx, sqlc.CreateApiKeyParams{
			TenantID:  utils.UUIDToPgUUID(tenantId),
			KeyHash:   hashedKey,
			KeyPrefix: prefix,
			Name:      utils.StringToPgText(req.Name),
		})

		return err
	}); err != nil {
		return nil, status.Errorf(codes.Internal, "error issuing api key: %v", err)
	}

	return &authv1.IssueApiKeyResponse{
		RawApiKey: rawKey,
		Id:        utils.PgUUIDToString(row.ID),
	}, nil
}

func (s *Service) RevokeApiKey(ctx context.Context, req *authv1.RevokeApiKeyRequest) (*authv1.RevokeApiKeyResponse, error) {
	keyId, err := uuid.Parse(req.Id)
	if err != nil {
		return &authv1.RevokeApiKeyResponse{Success: false}, status.Errorf(codes.InvalidArgument, "invalid key id: %v", err)
	}

	if err := s.querier.RevokeApiKey(ctx, utils.UUIDToPgUUID(keyId)); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to revoke key: %v", err)
	}

	return &authv1.RevokeApiKeyResponse{Success: true}, nil
}

func (s *Service) ListApiKeys(ctx context.Context, req *authv1.ListApiKeysRequest) (*authv1.ListApiKeysResponse, error) {
	tenantId, err := uuid.Parse(req.TenantId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant id: %v", err)
	}

	rows, err := s.querier.ListApiKeys(ctx, utils.UUIDToPgUUID(tenantId))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error listing api keys: %v", err)
	}

	results := make([]*authv1.ApiKeyMetadata, 0, len(rows))
	var lastUsed *timestamppb.Timestamp
	var expires *timestamppb.Timestamp
	for _, r := range rows {
		if r.LastUsedAt.Valid {
			lastUsed = timestamppb.New(r.LastUsedAt.Time)
		}

		if r.ExpiresAt.Valid {
			expires = timestamppb.New(r.ExpiresAt.Time)
		}

		meta := &authv1.ApiKeyMetadata{
			Id:         utils.PgUUIDToString(r.ID),
			Name:       r.Name.String,
			Prefix:     r.KeyPrefix,
			Revoked:    r.Revoked.Bool,
			CreatedAt:  timestamppb.New(r.CreatedAt.Time),
			LastUsedAt: lastUsed,
			ExpiresAt:  expires,
		}

		results = append(results, meta)
	}

	return &authv1.ListApiKeysResponse{Keys: results}, nil
}

func (s *Service) Register(ctx context.Context, req *authv1.RegisterRequest) (*authv1.RegisterResponse, error) {
	if err := ValidateEmail(req.Email); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	passwordHash, err := HashPassword(req.Password)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error hashing password: %v", err)
	}

	var reqErr error
	var userId pgtype.UUID
	if err := s.execTx(ctx, func(querier sqlc.Querier) error {
		userId, err = querier.CreateUser(ctx, sqlc.CreateUserParams{
			Email:        req.Email,
			PasswordHash: passwordHash,
			FullName:     utils.StringToPgText(req.FullName),
		})
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.ConstraintName == "users_email_key" {
				reqErr = errors.New("duplicate email")
				return reqErr
			}
		}

		return err
	}); err != nil {
		if reqErr != nil {
			return nil, status.Errorf(codes.AlreadyExists, "duplicate email: %s", req.Email)
		}

		return nil, status.Errorf(codes.Internal, "error registering user: %v", err)
	}

	return &authv1.RegisterResponse{UserId: utils.PgUUIDToString(userId)}, nil
}

func (s *Service) Login(ctx context.Context, req *authv1.LoginRequest) (*authv1.LoginResponse, error) {
	row, err := s.querier.GetUserByEmailWithPassword(ctx, req.Email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, status.Error(codes.Unauthenticated, "invalid email or password")
		}

		return nil, status.Errorf(codes.Internal, "error getting user: %v", err)
	}

	if err := CheckPassword(req.Password, row.PasswordHash); err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid email or password")
	}

	userID, err := uuid.Parse(utils.PgUUIDToString(row.ID))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error parsing user id: %v", err)
	}

	token, payload, err := s.tokenMaker.CreateToken(req.Email, userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error generating token: %v", err)
	}

	return &authv1.LoginResponse{
		AccessToken: token,
		ExpiresAt:   timestamppb.New(payload.ExpiredAt),
	}, nil
}

func (s *Service) execTx(ctx context.Context, fn func(sqlc.Querier) error) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	qtx := sqlc.New(tx)

	if err := fn(qtx); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
