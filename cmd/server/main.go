package main

import (
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/vantutran2k1/rwe/config"
	workflowv1 "github.com/vantutran2k1/rwe/gen/go/workflow/v1"
	"github.com/vantutran2k1/rwe/internal/common/db"
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

	workflowSvc := workflow.NewService(pool)

	grpcServer := grpc.NewServer()

	workflowv1.RegisterWorkflowServiceServer(grpcServer, workflowSvc)

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

	logger.Info("shutting down gRPC Server...")
	grpcServer.GracefulStop()
	logger.Info("server exited")
}
