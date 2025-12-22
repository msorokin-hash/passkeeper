package middleware

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	proto "github.com/msorokin-hash/passkeeper/internal/protobuf"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func getAuthFromContext(ctx context.Context) []string {
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		return nil
	}
	return md["authorization"]
}

func TestTokenInterceptor_AddsAuthorizationHeader(t *testing.T) {
	t.Helper()

	workDir := t.TempDir()
	tokenFilePath := filepath.Join(workDir, ".token.jwt")

	const tokenValue = "test-token-123"
	if err := os.WriteFile(tokenFilePath, []byte(tokenValue+"\n"), 0o600); err != nil {
		t.Fatalf("failed to write token file: %v", err)
	}

	interceptor := TokenInterceptor(workDir)

	var capturedAuth []string

	fakeInvoker := func(
		ctx context.Context,
		method string,
		req interface{},
		reply interface{},
		cc *grpc.ClientConn,
		opts ...grpc.CallOption,
	) error {
		capturedAuth = getAuthFromContext(ctx)
		return nil
	}

	fullMethod := "/somepkg.VaultService/GetSecret"

	err := interceptor(context.Background(), fullMethod, nil, nil, nil, fakeInvoker)
	if err != nil {
		t.Fatalf("interceptor returned unexpected error: %v", err)
	}

	if len(capturedAuth) != 1 {
		t.Fatalf("expected 1 Authorization header, got %d (%v)", len(capturedAuth), capturedAuth)
	}

	expected := "Bearer " + tokenValue
	if capturedAuth[0] != expected {
		t.Fatalf("expected Authorization %q, got %q", expected, capturedAuth[0])
	}
}

func TestTokenInterceptor_NoTokenFile_NoAuthHeader(t *testing.T) {
	t.Helper()

	workDir := t.TempDir()
	tokenFilePath := filepath.Join(workDir, ".token.jwt")

	if err := os.Remove(tokenFilePath); err != nil && !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("failed to ensure token file does not exist: %v", err)
	}

	interceptor := TokenInterceptor(workDir)

	var capturedAuth []string

	fakeInvoker := func(
		ctx context.Context,
		method string,
		req interface{},
		reply interface{},
		cc *grpc.ClientConn,
		opts ...grpc.CallOption,
	) error {
		capturedAuth = getAuthFromContext(ctx)
		return nil
	}

	fullMethod := "/somepkg.VaultService/GetSecret"

	err := interceptor(context.Background(), fullMethod, nil, nil, nil, fakeInvoker)
	if err != nil {
		t.Fatalf("interceptor returned unexpected error: %v", err)
	}

	if len(capturedAuth) != 0 {
		t.Fatalf("expected no Authorization header, got %v", capturedAuth)
	}
}

func TestTokenInterceptor_UserServiceMethod_NoAuthHeader(t *testing.T) {
	t.Helper()

	workDir := t.TempDir()
	tokenFilePath := filepath.Join(workDir, ".token.jwt")

	const tokenValue = "user-service-token"
	if err := os.WriteFile(tokenFilePath, []byte(tokenValue), 0o600); err != nil {
		t.Fatalf("failed to write token file: %v", err)
	}

	interceptor := TokenInterceptor(workDir)

	var capturedAuth []string

	fakeInvoker := func(
		ctx context.Context,
		method string,
		req interface{},
		reply interface{},
		cc *grpc.ClientConn,
		opts ...grpc.CallOption,
	) error {
		capturedAuth = getAuthFromContext(ctx)
		return nil
	}

	userServiceName := proto.UserService_ServiceDesc.ServiceName
	fullMethod := "/" + userServiceName + "/Login"

	err := interceptor(context.Background(), fullMethod, nil, nil, nil, fakeInvoker)
	if err != nil {
		t.Fatalf("interceptor returned unexpected error: %v", err)
	}

	if len(capturedAuth) != 0 {
		t.Fatalf("expected no Authorization header for UserService method, got %v", capturedAuth)
	}
}

func TestTokenInterceptor_Unauthenticated_RemovesTokenFile(t *testing.T) {
	t.Helper()

	workDir := t.TempDir()
	tokenFilePath := filepath.Join(workDir, ".token.jwt")

	const tokenValue = "expired-or-invalid-token"
	if err := os.WriteFile(tokenFilePath, []byte(tokenValue), 0o600); err != nil {
		t.Fatalf("failed to write token file: %v", err)
	}

	interceptor := TokenInterceptor(workDir)

	fakeInvoker := func(
		ctx context.Context,
		method string,
		req interface{},
		reply interface{},
		cc *grpc.ClientConn,
		opts ...grpc.CallOption,
	) error {
		return status.Error(codes.Unauthenticated, "invalid token")
	}

	fullMethod := "/somepkg.VaultService/GetSecret"

	err := interceptor(context.Background(), fullMethod, nil, nil, nil, fakeInvoker)
	if status.Code(err) != codes.Unauthenticated {
		t.Fatalf("expected Unauthenticated error to be returned, got: %v", err)
	}

	_, statErr := os.Stat(tokenFilePath)
	if !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("expected token file %q to be removed, stat err: %v", tokenFilePath, statErr)
	}
}

func TestIsUserServiceMethod(t *testing.T) {
	t.Helper()

	userServiceName := proto.UserService_ServiceDesc.ServiceName

	tests := []struct {
		name       string
		fullMethod string
		want       bool
	}{
		{
			name:       "UserService method",
			fullMethod: "/" + userServiceName + "/Login",
			want:       true,
		},
		{
			name:       "different service",
			fullMethod: "/otherpkg.OtherService/Get",
			want:       false,
		},
		{
			name:       "malformed name",
			fullMethod: "no-leading-slash",
			want:       false,
		},
		{
			name:       "empty string",
			fullMethod: "",
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isUserServiceMethod(tt.fullMethod)
			if got != tt.want {
				t.Fatalf("isUserServiceMethod(%q) = %v, want %v", tt.fullMethod, got, tt.want)
			}
		})
	}
}
