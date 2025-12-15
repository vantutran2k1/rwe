package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/vantutran2k1/rwe/config"
	workflowv1 "github.com/vantutran2k1/rwe/gen/go/workflow/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	ctx := context.Background()

	cfg, err := config.Load(".")
	if err != nil {
		panic("failed to load config: " + err.Error())
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	mux := runtime.NewServeMux()

	grpcEndpoint := "localhost" + cfg.Server.GRPCPort
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	err = workflowv1.RegisterWorkflowServiceHandlerFromEndpoint(ctx, mux, grpcEndpoint, opts)
	if err != nil {
		logger.Error("failed to register gateway", "error", err)
		os.Exit(1)
	}

	logger.Info("starting rest gateway", "port", cfg.Server.HTTPPort)

	if err := http.ListenAndServe(cfg.Server.HTTPPort, mux); err != nil {
		logger.Error("gateway server failed", "error", err)
		os.Exit(1)
	}
}
