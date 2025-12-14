package auth

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type TenantRecord struct {
	ID   string
	Name string
}

type Repository interface {
	FindTenantByAPIKey(ctx context.Context, apiKey string) (*TenantRecord, error)
}

type repo struct {
	db pgx.Tx
}
