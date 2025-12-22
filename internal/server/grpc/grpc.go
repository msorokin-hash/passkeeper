package grpc

import (
	"fmt"
	"net"

	"github.com/msorokin-hash/passkeeper/internal/server/handler"
	"github.com/msorokin-hash/passkeeper/internal/server/jwt"
	"github.com/msorokin-hash/passkeeper/internal/storage"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"

	interceptors "github.com/msorokin-hash/passkeeper/internal/server/middleware"

	proto "github.com/msorokin-hash/passkeeper/internal/protobuf"
)

// GrpcTransport represents the gRPC server transport layer of the application.
// It holds the gRPC server instance, the configured storage backend, the user
// service handler, and a logger used for reporting server events and errors.
type GrpcTransport struct {
	grpcServerAddr string
	storage        storage.Storage
	grpcServer     *grpc.Server
	userHandler    *handler.GRPCUserHandler
	vaultHandler   *handler.GRPCVaultHandler
	log            *logrus.Entry
}

// GrpcTransportConfig defines configuration settings required to create
// a gRPC transport instance. It includes the server address, master key
// for decrypting user secrets, and an optional TLS enablement flag.
type GrpcTransportConfig struct {
	MasterKey     string
	ServerAddress string
	UseTLS        bool
}

// NewGRPCTransport initializes and returns a new gRPC transport server.
// It configures authentication and logging interceptors, sets up TLS when
// enabled, registers service handlers, and enables server reflection.
// Any failure during initialization results in an error being returned.
func NewGRPCTransport(
	cfg GrpcTransportConfig,
	storage storage.Storage,
	tokenService jwt.TokenService,
	logger *logrus.Logger,
) (*GrpcTransport, error) {
	var opts []grpc.ServerOption

	log := logger.WithField("instance", "grpcTransport")
	auth := interceptors.NewAuthInterceptor(cfg.MasterKey, storage, tokenService, log)

	opts = append(opts, grpc.ChainUnaryInterceptor(
		interceptors.LoggerInterceptor(log),
		auth.AuthenticateUser,
	))

	if cfg.UseTLS {
		creds, err := credentials.NewServerTLSFromFile("ssl/server-cert.pem", "ssl/server-key.pem")
		if err != nil {
			return nil, err
		}
		opts = append(opts, grpc.Creds(creds))
	}

	grpcServer := grpc.NewServer(opts...)

	userHandler := handler.NewGRPCUserHandler(cfg.MasterKey, storage, tokenService, log)
	vaultHandler := handler.NewGRPCVaultHandler(cfg.MasterKey, storage, log)

	grpcTransport := &GrpcTransport{
		grpcServerAddr: cfg.ServerAddress,
		grpcServer:     grpcServer,
		storage:        storage,
		userHandler:    userHandler,
		vaultHandler:   vaultHandler,
		log:            log,
	}

	proto.RegisterUserServiceServer(grpcTransport.grpcServer, grpcTransport.userHandler)
	proto.RegisterVaultServiceServer(grpcTransport.grpcServer, grpcTransport.vaultHandler)
	reflection.Register(grpcTransport.grpcServer)

	return grpcTransport, nil
}

// Run starts the gRPC server and begins listening for incoming requests.
// The method blocks until the server is stopped or an unrecoverable error occurs.
func (s *GrpcTransport) Run() error {
	listener, err := net.Listen("tcp", s.grpcServerAddr)
	if err != nil {
		return fmt.Errorf("listen tcp has failed: %w", err)
	}

	return s.grpcServer.Serve(listener)
}

// Stop gracefully shuts down the gRPC server. It waits for ongoing requests
// to finish before returning control to the caller.
func (s *GrpcTransport) Stop() error {
	s.grpcServer.GracefulStop()
	return nil
}
