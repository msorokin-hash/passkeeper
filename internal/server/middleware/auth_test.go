package middleware

import (
	"context"
	"errors"
	"testing"

	tmock "github.com/msorokin-hash/passkeeper/internal/server/jwt/mocks"
	"github.com/msorokin-hash/passkeeper/internal/storage/mocks"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func Test_AuthenticateUser_NoAuthNeeded_MethodPassesThrough(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	mockTokenService := tmock.NewMockTokenService(ctrl)
	logger, _ := test.NewNullLogger()
	entry := logrus.NewEntry(logger)
	masterKey := "1d0e95ed9e11b59ba42200720c252f98d4cd440412926a0c15b6a95e03ab4480"

	interceptor := NewAuthInterceptor(masterKey, mockStorage, mockTokenService, entry)

	info := &grpc.UnaryServerInfo{FullMethod: "/gophkeeper.v1.UserService/Register"}

	called := false
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		called = true
		return "ok", nil
	}

	ctx := context.Background()
	resp, err := interceptor.AuthenticateUser(ctx, "req", info, handler)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp != "ok" {
		t.Fatalf("expected response 'ok', got %v", resp)
	}
	if !called {
		t.Fatalf("handler was not called")
	}
}

func Test_AuthenticateUser_MissingToken_ReturnsUnauthenticated(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	mockTokenService := tmock.NewMockTokenService(ctrl)
	logger, _ := test.NewNullLogger()
	entry := logrus.NewEntry(logger)
	masterKey := "1d0e95ed9e11b59ba42200720c252f98d4cd440412926a0c15b6a95e03ab4480"

	interceptor := NewAuthInterceptor(masterKey, mockStorage, mockTokenService, entry)

	info := &grpc.UnaryServerInfo{FullMethod: "/gophkeeper.v1.VaultService/GetData"}

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "should-not-be-called", nil
	}

	ctx := context.Background()

	_, err := interceptor.AuthenticateUser(ctx, "req", info, handler)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if status.Code(err) != codes.Unauthenticated {
		t.Fatalf("expected Unauthenticated, got %v (err=%v)", status.Code(err), err)
	}
}

func Test_AuthenticateUser_BadToken_ReturnsUnauthenticated(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	mockTokenService := tmock.NewMockTokenService(ctrl)
	logger, _ := test.NewNullLogger()
	entry := logrus.NewEntry(logger)
	masterKey := "1d0e95ed9e11b59ba42200720c252f98d4cd440412926a0c15b6a95e03ab4480"

	interceptor := NewAuthInterceptor(masterKey, mockStorage, mockTokenService, entry)

	info := &grpc.UnaryServerInfo{FullMethod: "/gophkeeper.v1.VaultService/AddData"}

	mockTokenService.
		EXPECT().
		GetClaims("XXX").
		Return(nil, errors.New("bad token"))

	md := metadata.New(map[string]string{"authorization": "Bearer XXX"})
	ctx := metadata.NewIncomingContext(context.Background(), md)

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "should-not-be-called", nil
	}

	_, err := interceptor.AuthenticateUser(ctx, "req", info, handler)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if status.Code(err) != codes.Unauthenticated {
		t.Fatalf("expected Unauthenticated, got %v (err=%v)", status.Code(err), err)
	}
}

func Test_AuthenticateUser_MissingMetadata_ReturnsUnauthenticated(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	mockTokenService := tmock.NewMockTokenService(ctrl)
	logger, _ := test.NewNullLogger()
	entry := logrus.NewEntry(logger)
	masterKey := "1d0e95ed9e11b59ba42200720c252f98d4cd440412926a0c15b6a95e03ab4480"

	interceptor := NewAuthInterceptor(masterKey, mockStorage, mockTokenService, entry)

	info := &grpc.UnaryServerInfo{
		FullMethod: "/gophkeeper.v1.VaultService/AddData",
	}

	ctx := context.Background()

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		t.Fatalf("handler should not be called")
		return nil, nil
	}

	_, err := interceptor.AuthenticateUser(ctx, "req", info, handler)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if status.Code(err) != codes.Unauthenticated {
		t.Fatalf(
			"expected Unauthenticated, got %v (err=%v)",
			status.Code(err),
			err,
		)
	}
}
