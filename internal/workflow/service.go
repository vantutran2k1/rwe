package workflow

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	workflowv1 "github.com/vantutran2k1/rwe/gen/go/workflow/v1"
	"github.com/vantutran2k1/rwe/internal/common/db"
	"github.com/vantutran2k1/rwe/internal/common/utils"
	sqlc "github.com/vantutran2k1/rwe/internal/workflow/db"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Service struct {
	pool    *pgxpool.Pool
	querier sqlc.Querier
	workflowv1.UnimplementedWorkflowServiceServer
}

func NewService(pool *pgxpool.Pool) *Service {
	return &Service{
		pool:    pool,
		querier: sqlc.New(pool),
	}
}

func (s *Service) CreateWorkflow(ctx context.Context, req *workflowv1.CreateWorkflowRequest) (*workflowv1.CreateWorkflowResponse, error) {
	// TODO: get tenant id from auth
	tenantID, err := uuid.Parse(req.TenantId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant id: %v", err)
	}

	// TODO: check for unique name per tenant

	definition, err := protojson.Marshal(req.Definition)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid definition: %v", err)
	}

	// TODO: validate definition

	var row sqlc.CreateWorkflowRow
	if err := s.execTx(ctx, func(querier sqlc.Querier) error {
		row, err = querier.CreateWorkflow(ctx, sqlc.CreateWorkflowParams{
			TenantID:   utils.UUIDToPgUUID(tenantID),
			Name:       req.Name,
			Definition: definition,
		})

		return err
	}); err != nil {
		return nil, status.Errorf(codes.Internal, "error creating workflow: %v", err)
	}

	return &workflowv1.CreateWorkflowResponse{
		Id:       utils.PgUUIDToString(row.ID),
		TenantId: utils.PgUUIDToString(row.TenantID),
		Name:     row.Name,
	}, nil
}

func (s *Service) GetWorkflow(ctx context.Context, req *workflowv1.GetWorkflowRequest) (*workflowv1.GetWorkflowResponse, error) {
	workflowID, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid workflow id: %v", err)
	}

	row, err := s.querier.GetWorkflowByID(ctx, utils.UUIDToPgUUID(workflowID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, status.Errorf(codes.NotFound, "workflow with id %s not found", workflowID)
		}

		return nil, status.Errorf(codes.Internal, "error getting workflow: %v", err)
	}

	// TODO: check for tenant id

	var definition structpb.Struct
	if err := protojson.Unmarshal(row.Definition, &definition); err != nil {
		return nil, status.Errorf(codes.Internal, "error parsing definition: %v", err)
	}

	return &workflowv1.GetWorkflowResponse{
		Id:         utils.PgUUIDToString(row.ID),
		TenantId:   utils.PgUUIDToString(row.TenantID),
		Name:       row.Name,
		Version:    row.Version.Int32,
		Definition: &definition,
		CreatedAt:  timestamppb.New(row.CreatedAt.Time),
		UpdatedAt:  timestamppb.New(row.UpdatedAt.Time),
		Archived:   row.Archived.Bool,
	}, nil
}

func (s *Service) GetWorkflows(ctx context.Context, req *workflowv1.GetWorkflowsRequest) (*workflowv1.GetWorkflowsResponse, error) {
	c, err := db.DecodeCursor(req.Token)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid token: %v", err)
	}

	pageSize := int32(20)
	if req.PageSize > pageSize {
		req.PageSize = pageSize
	}

	// TODO: get tenant id from auth
	uid, _ := uuid.Parse("22222222-2222-2222-2222-222222222222")
	rows, err := s.querier.ListWorkflowsByTenantID(ctx, sqlc.ListWorkflowsByTenantIDParams{
		TenantID:  utils.UUIDToPgUUID(uid),
		UpdatedAt: utils.TimeToPgTimestamptz(c.LastUpdatedAt),
		ID:        utils.UUIDToPgUUID(c.LastID),
		Limit:     pageSize,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error getting workflows: %v", err)
	}

	wfs := make([]*workflowv1.Workflow, 0, len(rows))
	var definition structpb.Struct
	for _, row := range rows {
		if err := protojson.Unmarshal(row.Definition, &definition); err != nil {
			return nil, status.Errorf(codes.Internal, "error parsing definition: %v", err)
		}

		wf := workflowv1.Workflow{
			Id:         utils.PgUUIDToString(row.ID),
			TenantId:   utils.PgUUIDToString(row.TenantID),
			Name:       row.Name,
			Version:    row.Version.Int32,
			Definition: &definition,
			CreatedAt:  timestamppb.New(row.CreatedAt.Time),
			UpdatedAt:  timestamppb.New(row.UpdatedAt.Time),
			Archived:   row.Archived.Bool,
		}
		wfs = append(wfs, &wf)
	}

	return &workflowv1.GetWorkflowsResponse{Workflows: wfs}, nil
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
