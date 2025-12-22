package service

import (
	"github.com/msorokin-hash/passkeeper/internal/client/grpc"
)

// Service represents the client-side application service layer.
// It wraps the gRPC client and provides access to functionality
// that requires the user's working directory (workDir) such as
// local token storage or file-based configuration.
type Service struct {
	grpcClient *grpc.Client
	workDir    string
}

// NewService creates and returns a new Service instance.
// Returns Service that is ready to be used as a high-level interface to the
// application's remote and local operations.
func NewService(grpcClient *grpc.Client, workDir string) *Service {
	return &Service{
		grpcClient: grpcClient,
		workDir:    workDir,
	}
}
