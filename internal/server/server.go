package server

import (
	"context"
	"fmt"

	"github.com/msorokin-hash/passkeeper/internal/server/config"
	"github.com/msorokin-hash/passkeeper/internal/server/grpc"
	"github.com/msorokin-hash/passkeeper/internal/server/jwt"
	"github.com/msorokin-hash/passkeeper/internal/storage"
	"github.com/msorokin-hash/passkeeper/internal/storage/postgres"
	"github.com/sirupsen/logrus"
)

// Server represents the main application server. It encapsulates the storage
// layer, the gRPC transport, and the logger used for reporting server activity.
type Server struct {
	storage       storage.Storage
	grpcTransport *grpc.GrpcTransport
	log           *logrus.Entry
}

// NewServer initializes and returns a new Server instance. It establishes the
// PostgreSQL storage connection, configures the JWT token service, and creates
// the gRPC transport based on the provided application configuration. If any
// initialization step fails, the function returns an error.
func NewServer(ctx context.Context, cfg *config.Config, logger *logrus.Logger) (*Server, error) {
	log := logger.WithField("instance", "server")

	storage, err := postgres.NewPGStorage(ctx, cfg.Database.PGConnectionString(), logger)
	if err != nil {
		return nil, fmt.Errorf("failed to run storage: %w", err)
	}

	tokenService := jwt.NewTokenDataService(cfg.Server.TokenKey, cfg.Server.TokenLifetime)
	grpcTransport, err := grpc.NewGRPCTransport(grpc.GrpcTransportConfig{
		MasterKey:     cfg.Server.MasterKey,
		ServerAddress: cfg.Server.Address,
		UseTLS:        cfg.Server.UseTLS,
	}, storage, tokenService, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create grpc transport: %w", err)
	}

	return &Server{
		storage:       storage,
		grpcTransport: grpcTransport,
		log:           log,
	}, nil
}

// Run starts the gRPC transport layer and begins handling incoming requests.
// It blocks until the server is stopped or encounters an unrecoverable error.
func (s *Server) Run() error {
	return s.grpcTransport.Run()
}

// Stop gracefully shuts down the server. It stops the gRPC transport,
// logs shutdown events, and closes the underlying storage connection.
// Any errors during shutdown are logged but not returned to the caller.
func (s *Server) Stop() error {
	if err := s.grpcTransport.Stop(); err != nil {
		s.log.Errorf("an error occurred during grpc server shutdown: %v", err)
	}
	s.log.Info("gRPC server stopped")
	s.storage.Close()
	s.log.Info("storage closed")
	return nil
}
