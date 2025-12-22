package handler

import (
	"context"
	"testing"
	"time"

	"github.com/msorokin-hash/passkeeper/internal/entity"
	errorscustom "github.com/msorokin-hash/passkeeper/internal/errors"
	"github.com/msorokin-hash/passkeeper/internal/logger"
	proto "github.com/msorokin-hash/passkeeper/internal/protobuf"
	"github.com/msorokin-hash/passkeeper/internal/server/jwt"
	"github.com/msorokin-hash/passkeeper/internal/storage/mocks"
	"github.com/msorokin-hash/passkeeper/pkg/utils"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var lifeTime = 3600
var tokenKey = "secret123456789"
var masterKey = "1d0e95ed9e11b59ba42200720c252f98d4cd440412926a0c15b6a95e03ab4480"

func TestGRPCUserHandler_Register(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	tokenService := jwt.NewTokenDataService(tokenKey, time.Duration(lifeTime))
	log, err := logger.NewLogger("info")

	assert.NoError(t, err)

	handler := NewGRPCUserHandler(masterKey, mockStorage, tokenService, log.WithField("instance", "grpcTransport"))

	type Store struct {
		err    error
		userID string
	}

	tests := []struct {
		name    string
		request *proto.RegisterRequest
		store   *Store
		wantErr bool
		errCode codes.Code
	}{
		{
			name: "invalid password",
			request: &proto.RegisterRequest{
				Login:    "user",
				Password: "password",
			},
			store:   nil,
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "create user with success result",
			request: &proto.RegisterRequest{
				Login:    "user",
				Password: "password1234",
			},
			store: &Store{
				err:    nil,
				userID: "1",
			},
			wantErr: false,
		},
		{
			name: "create user that already exists",
			request: &proto.RegisterRequest{
				Login:    "user",
				Password: "password1234",
			},
			store: &Store{
				err: errorscustom.ErrLoginTaken,
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "internal database error",
			request: &proto.RegisterRequest{
				Login:    "user",
				Password: "password1234",
			},
			store: &Store{
				err: errorscustom.ErrInternalDatabase,
			},
			wantErr: true,
			errCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.store != nil {
				mockStorage.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Times(1).Return(tt.store.userID, tt.store.err)
			} else {
				mockStorage.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Times(0)
			}

			response, err := handler.Register(context.Background(), tt.request)
			if !tt.wantErr {
				assert.NoError(t, err)
				userData, err := tokenService.GetClaims(response.GetToken())
				assert.NoError(t, err)
				assert.Equal(t, tt.request.GetLogin(), userData.Login)
			} else {
				code, _ := status.FromError(err)
				assert.Equal(t, tt.errCode, code.Code())
			}
		})
	}
}

func TestGRPCUserHandler_Login(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	tokenService := jwt.NewTokenDataService(tokenKey, time.Duration(lifeTime))
	log, err := logger.NewLogger("info")

	assert.NoError(t, err)

	handler := NewGRPCUserHandler(masterKey, mockStorage, tokenService, log.WithField("instance", "grpcTransport"))

	type Store struct {
		err      error
		login    string
		user     *entity.User
		calcHash bool
	}

	tests := []struct {
		name    string
		request *proto.LoginRequest
		store   *Store
		wantErr bool
		errCode codes.Code
	}{
		{
			name: "success user login",
			request: &proto.LoginRequest{
				Login:    "user",
				Password: "password1234",
			},
			store: &Store{
				err:   nil,
				login: "user",
				user: &entity.User{
					ID:              "1",
					Login:           "user",
					EncryptedSecret: "secret",
				},
				calcHash: true,
			},
			wantErr: false,
		},
		{
			name: "failed user login with incorrect password",
			request: &proto.LoginRequest{
				Login:    "user",
				Password: "password1234",
			},
			store: &Store{
				err:   nil,
				login: "user",
				user: &entity.User{
					ID:              "1",
					Login:           "user",
					EncryptedSecret: "secret",
				},
				calcHash: false,
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "failed user login with incorrect username",
			request: &proto.LoginRequest{
				Login:    "user",
				Password: "password1234",
			},
			store: &Store{
				err:      errorscustom.ErrNoUser,
				login:    "user",
				calcHash: false,
			},
			wantErr: true,
			errCode: codes.NotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.store != nil {
				if tt.store.calcHash {
					hash, err := utils.GeneratePasswordHash(tt.request.GetPassword())
					assert.NoError(t, err)
					tt.store.user.PasswordHash = hash
				}
				mockStorage.EXPECT().GetUserByLogin(gomock.Any(), tt.store.login).Times(1).Return(tt.store.user, tt.store.err)
			} else {
				mockStorage.EXPECT().GetUserByLogin(gomock.Any(), gomock.Any()).Times(0)
			}

			response, err := handler.Login(context.Background(), tt.request)
			if !tt.wantErr {
				assert.NoError(t, err)
				userData, err := tokenService.GetClaims(response.GetToken())
				assert.NoError(t, err)
				assert.Equal(t, tt.request.GetLogin(), userData.Login)
			} else {
				code, _ := status.FromError(err)
				assert.Equal(t, tt.errCode, code.Code())
			}
		})
	}
}
