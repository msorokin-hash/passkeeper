package service

import (
	"context"
	"fmt"
	"path/filepath"

	"google.golang.org/grpc/status"

	proto "github.com/msorokin-hash/passkeeper/internal/protobuf"
	"github.com/msorokin-hash/passkeeper/pkg/utils"
)

// Login performs user authentication via gRPC.
// It sends the provided login and password to the UserService.Login endpoint,
// and on success stores the returned JWT token on disk under the service's workDir.
// The token is written to the same file that is later used by the gRPC interceptor.
//
// On success, it returns the JWT token. On failure, it wraps gRPC status errors
// into a user-friendly message and returns a non-nil error.
func (s *Service) Login(ctx context.Context, login string, password string) (token string, err error) {
	res, err := s.grpcClient.UserClient.Login(ctx, &proto.LoginRequest{
		Login: login, Password: password,
	})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			return "", fmt.Errorf("user authentication failed: %s", st.Message())
		}
		return "", err
	}

	tokenFilePath := filepath.Join(s.workDir, "token.jwt")
	err = utils.AppendToFile(tokenFilePath, res.GetToken())
	if err != nil {
		return "", fmt.Errorf("failed to write token to file: %w", err)
	}

	return res.GetToken(), nil
}

// Register performs user registration via gRPC.
// It sends the provided login and password to the UserService.Register endpoint,
// and on success stores the returned JWT token on disk under the service's workDir.
// The token is written to the same file that is later used by the gRPC interceptor.
//
// On success, it returns the JWT token. On failure, it wraps gRPC status errors
// into a user-friendly message and returns a non-nil error.
func (s *Service) Register(ctx context.Context, login string, password string) (token string, err error) {
	res, err := s.grpcClient.UserClient.Register(ctx, &proto.RegisterRequest{
		Login: login, Password: password,
	})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			return "", fmt.Errorf("user registration failed: %s", st.Message())
		}
		return "", err
	}

	tokenFilePath := filepath.Join(s.workDir, "token.jwt")
	err = utils.AppendToFile(tokenFilePath, res.GetToken())
	if err != nil {
		return "", fmt.Errorf("failed to write token to file: %w", err)
	}

	return res.GetToken(), nil
}
