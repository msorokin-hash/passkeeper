package handler

import (
	"context"
	"errors"
	"testing"

	"github.com/msorokin-hash/passkeeper/internal/entity"
	errorscustom "github.com/msorokin-hash/passkeeper/internal/errors"
	"github.com/msorokin-hash/passkeeper/internal/logger"
	proto "github.com/msorokin-hash/passkeeper/internal/protobuf"
	cx "github.com/msorokin-hash/passkeeper/internal/server/context"
	"github.com/msorokin-hash/passkeeper/internal/storage/mocks"
	"github.com/msorokin-hash/passkeeper/pkg/utils"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestGRPCVaultHandler_GetData(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	log, err := logger.NewLogger("info")

	assert.NoError(t, err)

	handler := NewGRPCVaultHandler(masterKey, mockStorage, log.WithField("instance", "grpcTransport"))

	type Store struct {
		err    error
		result *entity.VaultItem
	}

	tests := []struct {
		name    string
		user    *cx.UserContext
		request *proto.GetDataRequest
		data    []byte
		store   *Store
		wantErr bool
		errCode codes.Code
	}{
		{
			name: "success getting data",
			user: &cx.UserContext{
				ID:     "1",
				Login:  "password",
				Secret: masterKey,
			},
			request: &proto.GetDataRequest{
				Id: "1",
			},
			data: []byte("some stored data"),
			store: &Store{
				err: nil,
				result: &entity.VaultItem{
					ID:     "1",
					UserID: "1",
					Meta:   "some data meta",
					Type:   "PASSWORD",
				},
			},
			wantErr: false,
		},
		{
			name:    "failed with no id in request",
			request: &proto.GetDataRequest{},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "failed get user data from context",
			request: &proto.GetDataRequest{
				Id: "1",
			},
			wantErr: true,
			errCode: codes.Unauthenticated,
		},
		{
			name: "failed with no data found",
			user: &cx.UserContext{
				ID:     "1",
				Login:  "password",
				Secret: masterKey,
			},
			request: &proto.GetDataRequest{
				Id: "1",
			},
			store: &Store{
				err: errorscustom.ErrNoData,
			},
			wantErr: true,
			errCode: codes.NotFound,
		},
		{
			name: "failed with database error",
			user: &cx.UserContext{
				ID:     "1",
				Login:  "password",
				Secret: masterKey,
			},
			request: &proto.GetDataRequest{
				Id: "1",
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
				mockStorage.EXPECT().GetItem(gomock.Any(), tt.request.GetId(), tt.user.ID).DoAndReturn(
					func(ctx context.Context, id string, userID string) (*entity.VaultItem, error) {
						if tt.store.err != nil {
							return nil, tt.store.err
						}
						enc, err := utils.Encrypt(tt.data, tt.user.Secret)
						assert.NoError(t, err)
						tt.store.result.EncryptedData = enc
						return tt.store.result, nil
					},
				)
			}

			ctx := context.Background()
			if tt.user != nil {
				ctx = cx.WrapContextWithUser(ctx, tt.user)
			}

			response, err := handler.GetData(ctx, tt.request)
			if !tt.wantErr {
				assert.NoError(t, err)
				assert.Equal(t, tt.data, response.GetItem().GetData())
				assert.Equal(t, tt.store.result.Meta, response.GetItem().GetMeta())
			} else {
				code, _ := status.FromError(err)
				assert.Equal(t, tt.errCode, code.Code())
			}
		})
	}
}

func TestGRPCVaultHandler_DeleteData(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	log, err := logger.NewLogger("info")

	assert.NoError(t, err)

	handler := NewGRPCVaultHandler(masterKey, mockStorage, log.WithField("instance", "grpcTransport"))

	type Store struct {
		err error
	}

	tests := []struct {
		name    string
		user    *cx.UserContext
		request *proto.DeleteDataRequest
		store   *Store
		wantErr bool
		errCode codes.Code
	}{
		{
			name: "success deleting data",
			user: &cx.UserContext{
				ID:    "1",
				Login: "user",
			},
			request: &proto.DeleteDataRequest{
				Id: "1",
			},
			store: &Store{
				err: nil,
			},
			wantErr: false,
		},
		{
			name: "failed with no id in request",
			user: &cx.UserContext{
				ID:    "1",
				Login: "user",
			},
			request: &proto.DeleteDataRequest{},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "failed get user data from context",
			request: &proto.DeleteDataRequest{
				Id: "1",
			},
			wantErr: true,
			errCode: codes.Unauthenticated,
		},
		{
			name: "failed with no data found",
			user: &cx.UserContext{
				ID:    "1",
				Login: "password",
			},
			request: &proto.DeleteDataRequest{
				Id: "1",
			},
			store: &Store{
				err: errorscustom.ErrNoData,
			},
			wantErr: true,
			errCode: codes.NotFound,
		},
		{
			name: "failed with database error",
			user: &cx.UserContext{
				ID:    "1",
				Login: "password",
			},
			request: &proto.DeleteDataRequest{
				Id: "1",
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
				mockStorage.EXPECT().DeleteItem(gomock.Any(), tt.request.GetId(), tt.user.ID).Times(1).Return(tt.store.err)
			}

			ctx := context.Background()
			if tt.user != nil {
				ctx = cx.WrapContextWithUser(ctx, tt.user)
			}

			_, err = handler.DeleteData(ctx, tt.request)
			if !tt.wantErr {
				assert.NoError(t, err)
			} else {
				code, _ := status.FromError(err)
				assert.Equal(t, tt.errCode, code.Code())
			}
		})
	}
}

func TestGRPCVaultHandler_AddData(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	log, err := logger.NewLogger("info")

	assert.NoError(t, err)

	handler := NewGRPCVaultHandler(masterKey, mockStorage, log.WithField("instance", "grpcTransport"))

	type Store struct {
		err error
	}

	tests := []struct {
		name    string
		user    *cx.UserContext
		request *proto.AddDataRequest
		store   *Store
		wantErr bool
		errCode codes.Code
	}{
		{
			name: "success adding data",
			user: &cx.UserContext{
				ID:     "1",
				Login:  "user",
				Secret: masterKey,
			},
			request: &proto.AddDataRequest{
				Item: &proto.Item{
					Data:       []byte("123"),
					RecordType: string(entity.Password),
					Meta:       "some meta info",
				},
			},
			store: &Store{
				err: nil,
			},
			wantErr: false,
		},
		{
			name: "failed with no data found",
			user: &cx.UserContext{
				ID:     "1",
				Login:  "user",
				Secret: masterKey,
			},
			request: &proto.AddDataRequest{
				Item: nil,
			},
			store:   nil,
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "failed get user data from context",
			user: nil,
			request: &proto.AddDataRequest{
				Item: &proto.Item{
					Data:       []byte("123"),
					RecordType: string(entity.Password),
					Meta:       "some meta info",
				},
			},
			store:   nil,
			wantErr: true,
			errCode: codes.Unauthenticated,
		},
		{
			name: "failed with database error",
			user: &cx.UserContext{
				ID:     "1",
				Login:  "user",
				Secret: masterKey,
			},
			request: &proto.AddDataRequest{
				Item: &proto.Item{
					Data:       []byte("123"),
					RecordType: string(entity.Password),
					Meta:       "some meta info",
				},
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
				mockStorage.EXPECT().CreateItem(gomock.Any(), tt.user.ID, gomock.Any()).DoAndReturn(
					func(_ context.Context, userID string, item *entity.VaultItem) (string, error) {
						if len(item.EncryptedData) == 0 {
							t.Errorf("encrypted data is empty")
						}
						if userID != tt.user.ID ||
							item.Meta != tt.request.GetItem().GetMeta() ||
							item.Type != entity.DataType(tt.request.GetItem().GetRecordType()) {
							t.Errorf("unexpected vaul item data: got %+v", item)
						}
						return "1", tt.store.err
					},
				)
			}

			ctx := context.Background()
			if tt.user != nil {
				ctx = cx.WrapContextWithUser(ctx, tt.user)
			}

			_, err = handler.AddData(ctx, tt.request)
			if !tt.wantErr {
				assert.NoError(t, err)
			} else {
				code, _ := status.FromError(err)
				assert.Equal(t, tt.errCode, code.Code())
			}
		})
	}
}

func TestGRPCVaultHandler_UpdateData(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	log, err := logger.NewLogger("info")

	assert.NoError(t, err)

	handler := NewGRPCVaultHandler(masterKey, mockStorage, log.WithField("instance", "grpcTransport"))

	type Store struct {
		err error
	}

	tests := []struct {
		name    string
		user    *cx.UserContext
		request *proto.UpdateDataRequest
		store   *Store
		wantErr bool
		errCode codes.Code
	}{
		{
			name: "success updating data",
			user: &cx.UserContext{
				ID:     "1",
				Login:  "user",
				Secret: masterKey,
			},
			request: &proto.UpdateDataRequest{
				Id:   "1",
				Data: []byte("some data"),
				Meta: "some meta",
			},
			store: &Store{
				err: nil,
			},
			wantErr: false,
		},
		{
			name: "failed with no id in request",
			user: &cx.UserContext{
				ID:    "1",
				Login: "user",
			},
			request: &proto.UpdateDataRequest{},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "failed get user data from context",
			user: nil,
			request: &proto.UpdateDataRequest{
				Id:   "1",
				Data: []byte("some data"),
				Meta: "some meta",
			},
			wantErr: true,
			errCode: codes.Unauthenticated,
		},
		{
			name: "failed with no data found",
			user: &cx.UserContext{
				ID:     "1",
				Login:  "password",
				Secret: masterKey,
			},
			request: &proto.UpdateDataRequest{
				Id:   "1",
				Data: []byte("some data"),
				Meta: "some meta",
			},
			store: &Store{
				err: errorscustom.ErrNoData,
			},
			wantErr: true,
			errCode: codes.NotFound,
		},
		{
			name: "failed with database error",
			user: &cx.UserContext{
				ID:     "1",
				Login:  "password",
				Secret: masterKey,
			},
			request: &proto.UpdateDataRequest{
				Id:   "1",
				Data: []byte("some data"),
				Meta: "some meta",
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
				mockStorage.EXPECT().UpdateItem(gomock.Any(), tt.request.GetId(), tt.user.ID, gomock.Any()).DoAndReturn(
					func(ctx context.Context, id string, userID string, item *entity.VaultItem) error {
						if len(item.EncryptedData) == 0 {
							t.Errorf("encrypted data is empty")
						}
						if userID != tt.user.ID ||
							item.Meta != tt.request.GetMeta() {
							t.Errorf("unexpected vaul item data: got %+v", item)
						}
						return tt.store.err
					},
				)
			}

			ctx := context.Background()
			if tt.user != nil {
				ctx = cx.WrapContextWithUser(ctx, tt.user)
			}

			_, err = handler.UpdateData(ctx, tt.request)
			if !tt.wantErr {
				assert.NoError(t, err)
			} else {
				code, _ := status.FromError(err)
				assert.Equal(t, tt.errCode, code.Code())
			}
		})
	}
}

func TestGRPCVaultHandler_GetAllByType(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	log, err := logger.NewLogger("info")

	assert.NoError(t, err)

	handler := NewGRPCVaultHandler(masterKey, mockStorage, log.WithField("instance", "grpcTransport"))

	type Store struct {
		err error
		res []entity.VaultItem
	}

	tests := []struct {
		name    string
		user    *cx.UserContext
		request *proto.GetAllByTypeRequest
		store   *Store
		wantErr bool
		errCode codes.Code
	}{
		{
			name: "success getting data",
			user: &cx.UserContext{
				ID:    "1",
				Login: "user",
			},
			request: &proto.GetAllByTypeRequest{
				RecordType: string(entity.Password),
			},
			store: &Store{
				err: nil,
				res: []entity.VaultItem{
					{
						ID:   "1",
						Meta: "some meta",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "failed getting user from context",
			request: &proto.GetAllByTypeRequest{
				RecordType: string(entity.Password),
			},
			wantErr: true,
			errCode: codes.Unauthenticated,
		},
		{
			name: "failed with database error",
			user: &cx.UserContext{
				ID:    "1",
				Login: "password",
			},
			request: &proto.GetAllByTypeRequest{
				RecordType: string(entity.Password),
			},
			store: &Store{
				err: errors.New("db error"),
			},
			wantErr: true,
			errCode: codes.Internal,
		},
		{
			name: "invalid data type",
			request: &proto.GetAllByTypeRequest{
				RecordType: "INVALID",
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.store != nil {
				mockStorage.EXPECT().GetItemsByType(gomock.Any(), tt.request.GetRecordType(), tt.user.ID).
					Times(1).Return(tt.store.res, tt.store.err)
			}

			ctx := context.Background()
			if tt.user != nil {
				ctx = cx.WrapContextWithUser(ctx, tt.user)
			}

			response, err := handler.GetAllByType(ctx, tt.request)
			if !tt.wantErr {
				assert.NoError(t, err)
				for i, item := range response.GetItems() {
					assert.Equal(t, tt.store.res[i].ID, item.GetId())
					assert.Equal(t, tt.store.res[i].Meta, item.GetMeta())
				}
			} else {
				code, _ := status.FromError(err)
				assert.Equal(t, tt.errCode, code.Code())
			}
		})
	}
}
