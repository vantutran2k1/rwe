package tenant

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/vantutran2k1/rwe/internal/common/utils"

	"github.com/gosimple/slug"
	"github.com/jackc/pgx/v5/pgxpool"
	tenantv1 "github.com/vantutran2k1/rwe/gen/go/tenant/v1"
	"github.com/vantutran2k1/rwe/internal/middlewares"
	sqlc "github.com/vantutran2k1/rwe/internal/tenant/db"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Service struct {
	pool    *pgxpool.Pool
	querier sqlc.Querier
	tenantv1.UnimplementedTenantServiceServer
}

func NewService(pool *pgxpool.Pool) *Service {
	return &Service{
		pool:    pool,
		querier: sqlc.New(pool),
	}
}

func (s *Service) CreateTenant(ctx context.Context, req *tenantv1.CreateTenantRequest) (*tenantv1.CreateTenantResponse, error) {
	tokenPayload := middlewares.GetTokenPayload(ctx)
	if tokenPayload == nil {
		return nil, status.Error(codes.Unauthenticated, "missing user authentication")
	}

	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "tenant name must not be empty")
	}

	sl := slug.Make(req.Name)
	_, err := s.querier.GetTenantBySlug(ctx, sl)
	if err == nil {
		return nil, status.Errorf(codes.AlreadyExists, "tenant name %s already exists", req.Name)
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return nil, status.Errorf(codes.Internal, "error getting tenant: %v", err)
	}

	if !isValidRegion(req.Region) {
		return nil, status.Errorf(codes.InvalidArgument, "invalid region for tenant")
	}

	var tenant sqlc.Tenant
	if err := s.execTx(ctx, func(querier sqlc.Querier) error {
		tenant, err = querier.CreateTenant(ctx, sqlc.CreateTenantParams{
			Name:         req.Name,
			Slug:         sl,
			ContactEmail: utils.StringToPgText(tokenPayload.Email),
			Tier:         utils.StringToPgText(string(tenantTierFree)),
			Region:       utils.StringToPgText(req.Region),
		})

		return err
	}); err != nil {
		return nil, status.Errorf(codes.Internal, "error creating tenant: %v", err)
	}

	return &tenantv1.CreateTenantResponse{
		Name:   tenant.Name,
		Slug:   tenant.Slug,
		Tier:   tenant.Tier.String,
		Region: tenant.Region.String,
		Status: string(tenant.Status.TenantStatus),
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
