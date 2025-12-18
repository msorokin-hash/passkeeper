package service

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/msorokin-hash/passkeeper/internal/client/grpc"
	proto "github.com/msorokin-hash/passkeeper/internal/protobuf"
	mc "github.com/msorokin-hash/passkeeper/internal/protobuf/mocks"
	"github.com/msorokin-hash/passkeeper/pkg/utils"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestService_Register(t *testing.T) {
	type want struct {
		err error
		jwt string
	}

	tests := []struct {
		name     string
		login    string
		password string
		response *proto.AuthResponse
		resErr   error
		want     want
	}{
		{
			name:     "success",
			login:    "user1",
			password: "pass1",
			response: &proto.AuthResponse{Token: "jwt-token-123"},
			resErr:   nil,
			want: want{
				err: nil,
				jwt: "jwt-token-123",
			},
		},
		{
			name:     "grpc_status_error",
			login:    "user2",
			password: "pass2",
			response: nil,
			resErr:   status.Error(codes.AlreadyExists, "user already exists"),
			want: want{
				err: errors.New("user registration failed: user already exists"),
				jwt: "",
			},
		},
		{
			name:     "native_error",
			login:    "user3",
			password: "pass3",
			response: nil,
			resErr:   errors.New("network timeout"),
			want: want{
				err: errors.New("network timeout"),
				jwt: "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUserClient := mc.NewMockUserServiceClient(ctrl)

			mockUserClient.
				EXPECT().
				Register(gomock.Any(), gomock.Any()).
				Return(tt.response, tt.resErr)

			grpcClient := &grpc.Client{
				UserClient: mockUserClient,
			}

			workDir := t.TempDir()
			defer os.RemoveAll(workDir)

			svc := NewService(grpcClient, workDir)
			jwt, err := svc.Register(context.Background(), tt.login, tt.password)

			if tt.want.err != nil {
				assert.EqualError(t, err, tt.want.err.Error())
				assert.Empty(t, jwt)

				_, readErr := utils.ReadFile(filepath.Join(workDir, "token.jwt"))
				assert.Error(t, readErr)

				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want.jwt, jwt)

			tokenData, readErr := utils.ReadFile(filepath.Join(workDir, "token.jwt"))
			assert.NoError(t, readErr)
			assert.Equal(t, tt.want.jwt, string(tokenData))
		})
	}
}

func TestService_Login(t *testing.T) {
	type want struct {
		err string
		jwt string
	}

	tests := []struct {
		name     string
		login    string
		password string
		response *proto.AuthResponse
		resErr   error
		want     want
	}{
		{
			name:     "success",
			login:    "user1",
			password: "pass1",
			response: &proto.AuthResponse{Token: "jwt-login-123"},
			resErr:   nil,
			want: want{
				err: "",
				jwt: "jwt-login-123",
			},
		},
		{
			name:     "grpc_status_error",
			login:    "user2",
			password: "pass2",
			response: nil,
			resErr:   status.Error(codes.Unauthenticated, "invalid credentials"),
			want: want{
				err: "user authentication failed: invalid credentials",
				jwt: "",
			},
		},
		{
			name:     "native_error",
			login:    "user3",
			password: "pass3",
			response: nil,
			resErr:   errors.New("network timeout"),
			want: want{
				err: "network timeout",
				jwt: "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUserClient := mc.NewMockUserServiceClient(ctrl)

			mockUserClient.
				EXPECT().
				Login(gomock.Any(), gomock.Any()).
				Return(tt.response, tt.resErr)

			workDir := t.TempDir()
			defer os.RemoveAll(workDir)

			grpcClient := &grpc.Client{
				UserClient: mockUserClient,
			}

			svc := NewService(grpcClient, workDir)

			jwt, err := svc.Login(context.Background(), tt.login, tt.password)

			if tt.want.err != "" {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.want.err)
				assert.Empty(t, jwt)

				tokenFilePath := filepath.Join(workDir, "token.jwt")
				_, readErr := os.ReadFile(tokenFilePath)
				assert.Error(t, readErr)

				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want.jwt, jwt)

			tokenFilePath := filepath.Join(workDir, "token.jwt")
			data, readErr := os.ReadFile(tokenFilePath)
			assert.NoError(t, readErr)
			assert.Equal(t, tt.want.jwt, string(data))
		})
	}
}
