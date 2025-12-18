package grpc

import (
	"fmt"

	"github.com/msorokin-hash/passkeeper/internal/client/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	interceptors "github.com/msorokin-hash/passkeeper/internal/client/middleware"
	proto "github.com/msorokin-hash/passkeeper/internal/protobuf"
)

// Client represents a gRPC client for interacting with the GophKeeper server.
// It provides typed API clients for user authentication and vault operations.
type Client struct {
	UserClient  proto.UserServiceClient
	VaultClient proto.VaultServiceClient
}

// NewGRPCConnection initializes and returns a new gRPC client based on the
// provided configuration. It configures TLS if enabled, otherwise a plaintext
// insecure connection is used.
// Returns a Client containing initialized RPC stubs or an error
// if the connection could not be established.
func NewGRPCConnection(cfg *config.ClientConfig) (*Client, error) {
	var opts []grpc.DialOption

	opts = append(opts, grpc.WithChainUnaryInterceptor(
		interceptors.TokenInterceptor(cfg.WorkDir),
	))

	if cfg.UseTLS {
		creds, err := credentials.NewClientTLSFromFile("ssl/server-cert.pem", "")
		if err != nil {
			return nil, fmt.Errorf("load tls cert: %w", err)
		}
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	conn, err := grpc.NewClient(cfg.ServerAddress, opts...)
	if err != nil {
		return nil, fmt.Errorf("grpc new client: %w", err)
	}

	return &Client{
		UserClient:  proto.NewUserServiceClient(conn),
		VaultClient: proto.NewVaultServiceClient(conn),
	}, nil
}
