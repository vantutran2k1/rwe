package main

import (
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/vantutran2k1/rwe/config"
	authv1 "github.com/vantutran2k1/rwe/gen/go/auth/v1"
	tenantv1 "github.com/vantutran2k1/rwe/gen/go/tenant/v1"
	workflowv1 "github.com/vantutran2k1/rwe/gen/go/workflow/v1"
	"github.com/vantutran2k1/rwe/internal/auth"
	"github.com/vantutran2k1/rwe/internal/common/db"
	"github.com/vantutran2k1/rwe/internal/middlewares"
	"github.com/vantutran2k1/rwe/internal/tenant"
	"github.com/vantutran2k1/rwe/internal/workflow"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	cfg, err := config.Load(".")
	if err != nil {
		panic("failed to load config: " + err.Error())
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	pool, err := db.New(cfg.Database.URL)
	if err != nil {
		logger.Error("failed to connect to DB", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	tokenDuration := time.Duration(cfg.Auth.TokenDurationHours) * time.Hour
	tokenMaker, _ := auth.NewPasetoMaker(cfg.Auth.TokenSymmetricKey, tokenDuration)

	authInterceptor := middlewares.NewAuthInterceptor(tokenMaker)

	workflowSvc := workflow.NewService(pool)
	authSvc := auth.NewService(pool, tokenMaker)
	tenantSvc := tenant.NewService(pool)

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(authInterceptor.Unary()),
	)

	workflowv1.RegisterWorkflowServiceServer(grpcServer, workflowSvc)
	authv1.RegisterAuthServiceServer(grpcServer, authSvc)
	tenantv1.RegisterTenantServiceServer(grpcServer, tenantSvc)

	reflection.Register(grpcServer)

	lis, err := net.Listen("tcp", cfg.Server.GRPCPort)
	if err != nil {
		logger.Error("failed to listen", "error", err)
		os.Exit(1)
	}

	go func() {
		logger.Info("starting grpc server", "port", cfg.Server.GRPCPort)
		if err := grpcServer.Serve(lis); err != nil {
			logger.Error("grpc server failed", "error", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down grpc Server...")
	grpcServer.GracefulStop()
	logger.Info("server exited")
}
