package service

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/msorokin-hash/passkeeper/internal/client/grpc"
	"github.com/msorokin-hash/passkeeper/internal/entity"
	proto "github.com/msorokin-hash/passkeeper/internal/protobuf"
	mc "github.com/msorokin-hash/passkeeper/internal/protobuf/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestDeleteData(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		returnResp *proto.DeleteDataResponse
		returnErr  error
		wantErr    error
	}{
		{
			name:       "success",
			id:         "1",
			returnResp: &proto.DeleteDataResponse{},
			returnErr:  nil,
			wantErr:    nil,
		},
		{
			name:      "grpc_status_error",
			id:        "2",
			returnErr: status.Error(codes.NotFound, "data not found"),
			wantErr:   errors.New("failed to delete data: data not found"),
		},
		{
			name:      "native_error",
			id:        "3",
			returnErr: errors.New("connection lost"),
			wantErr:   errors.New("connection lost"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockVault := mc.NewMockVaultServiceClient(ctrl)
			req := &proto.DeleteDataRequest{Id: tt.id}

			mockVault.
				EXPECT().
				DeleteData(gomock.Any(), req).
				Return(tt.returnResp, tt.returnErr)

			grpcClient := &grpc.Client{
				VaultClient: mockVault,
			}

			svc := NewService(grpcClient, "")

			err := svc.DeleteData(context.Background(), tt.id)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.wantErr.Error())
				return
			}

			assert.NoError(t, err)
		})
	}
}

func TestService_GetAllByType(t *testing.T) {
	type want struct {
		errContains string
	}

	tests := []struct {
		name     string
		dataType entity.DataType
		response *proto.GetAllByTypeResponse
		grpcErr  error
		want     want
		assertFn func(t *testing.T, items []entity.ItemInfo)
	}{
		{
			name:     "success_password_items",
			dataType: entity.Password,
			response: &proto.GetAllByTypeResponse{
				Items: []*proto.GetAllByTypeResponse_TypeItem{
					{
						Id:   "1",
						Meta: `{"resource":"example.com","login":"user1","comment":"first"}`,
					},
					{
						Id:   "2",
						Meta: `{"resource":"example.org","login":"user2","comment":"second"}`,
					},
				},
			},
			grpcErr: nil,
			want:    want{errContains: ""},
			assertFn: func(t *testing.T, items []entity.ItemInfo) {
				assert.Len(t, items, 2)

				meta1, ok := items[0].Meta.(*entity.PasswordMeta)
				assert.True(t, ok, "expected PasswordMeta for first item")
				assert.Equal(t, "1", items[0].ID)
				assert.Equal(t, "example.com", meta1.Resource)
				assert.Equal(t, "user1", meta1.Login)
				assert.Equal(t, "first", meta1.Comment)

				meta2, ok := items[1].Meta.(*entity.PasswordMeta)
				assert.True(t, ok, "expected PasswordMeta for second item")
				assert.Equal(t, "2", items[1].ID)
				assert.Equal(t, "example.org", meta2.Resource)
				assert.Equal(t, "user2", meta2.Login)
				assert.Equal(t, "second", meta2.Comment)
			},
		},
		{
			name:     "grpc_status_error",
			dataType: entity.Text,
			response: nil,
			grpcErr:  status.Error(codes.Internal, "internal error"),
			want: want{
				errContains: "failed to get data: internal error",
			},
		},
		{
			name:     "native_error",
			dataType: entity.Text,
			response: nil,
			grpcErr:  errors.New("connection lost"),
			want: want{
				errContains: "connection lost",
			},
		},
		{
			name:     "unknown_data_type",
			dataType: entity.DataType("unknown"),
			response: &proto.GetAllByTypeResponse{
				Items: []*proto.GetAllByTypeResponse_TypeItem{
					{
						Id:   "10",
						Meta: `{}`,
					},
				},
			},
			grpcErr: nil,
			want: want{
				errContains: "unknown data type: unknown",
			},
		},
		{
			name:     "invalid_json_meta",
			dataType: entity.Password,
			response: &proto.GetAllByTypeResponse{
				Items: []*proto.GetAllByTypeResponse_TypeItem{
					{
						Id:   "broken",
						Meta: `{"resource":"example","login":"user",`,
					},
				},
			},
			grpcErr: nil,
			want: want{
				errContains: `failed to unmarshal metadata for item "broken"`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockVault := mc.NewMockVaultServiceClient(ctrl)

			req := &proto.GetAllByTypeRequest{
				RecordType: string(tt.dataType),
			}

			mockVault.
				EXPECT().
				GetAllByType(gomock.Any(), req).
				Return(tt.response, tt.grpcErr)

			grpcClient := &grpc.Client{
				VaultClient: mockVault,
			}

			svc := NewService(grpcClient, "")

			items, err := svc.GetAllByType(context.Background(), tt.dataType)

			if tt.want.errContains != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.want.errContains)
				assert.Nil(t, items)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, items)

			if tt.assertFn != nil {
				tt.assertFn(t, items)
			}
		})
	}
}

func TestService_GetData(t *testing.T) {
	type want struct {
		dataType    entity.DataType
		errContains string
	}

	tests := []struct {
		name     string
		id       string
		response *proto.GetDataResponse
		grpcErr  error
		want     want
		assertFn func(t *testing.T, dataType entity.DataType, data any, workDir string)
	}{
		{
			name: "success_password",
			id:   "1",
			response: &proto.GetDataResponse{
				Id: "1",
				Item: &proto.Item{
					RecordType: string(entity.Password),
					Meta:       `{"resource":"example.com","login":"user1","comment":"test"}`,
					Data:       []byte("super-secret-password"),
				},
			},
			want: want{dataType: entity.Password},
			assertFn: func(t *testing.T, dt entity.DataType, data any, _ string) {
				assert.Equal(t, entity.Password, dt)
				item, ok := data.(*entity.PasswordItem)
				assert.True(t, ok)
				assert.Equal(t, "example.com", item.Meta.Resource)
				assert.Equal(t, "user1", item.Meta.Login)
				assert.Equal(t, "test", item.Meta.Comment)
				assert.Equal(t, "super-secret-password", item.Data)
			},
		},
		{
			name: "success_text",
			id:   "2",
			response: &proto.GetDataResponse{
				Id: "2",
				Item: &proto.Item{
					RecordType: string(entity.Text),
					Meta:       `{"name":"note","comment":"short"}`,
					Data:       []byte("hello world"),
				},
			},
			want: want{dataType: entity.Text},
			assertFn: func(t *testing.T, dt entity.DataType, data any, _ string) {
				assert.Equal(t, entity.Text, dt)
				item, ok := data.(*entity.TextItem)
				assert.True(t, ok)
				assert.Equal(t, "note", item.Meta.Name)
				assert.Equal(t, "short", item.Meta.Comment)
				assert.Equal(t, "hello world", item.Data)
			},
		},
		{
			name: "success_file",
			id:   "3",
			response: &proto.GetDataResponse{
				Id: "3",
				Item: &proto.Item{
					RecordType: string(entity.File),
					Meta:       `{"name":"data.bin","comment":"file comment"}`,
					Data:       []byte("file-bytes"),
				},
			},
			want: want{dataType: entity.File},
			assertFn: func(t *testing.T, dt entity.DataType, data any, workDir string) {
				assert.Equal(t, entity.File, dt)

				item, ok := data.(*entity.FileItem)
				assert.True(t, ok)
				assert.Equal(t, "data.bin", item.Meta.Name)
				assert.Equal(t, "file comment", item.Meta.Comment)

				path := filepath.Join(workDir, "data.bin")
				content, err := os.ReadFile(path)
				assert.NoError(t, err)
				assert.Equal(t, []byte("file-bytes"), content)
			},
		},
		{
			name:    "grpc_status_error",
			id:      "4",
			grpcErr: status.Error(codes.NotFound, "not found"),
			want:    want{errContains: "failed to get data: not found"},
		},
		{
			name:    "native_error",
			id:      "5",
			grpcErr: errors.New("connection lost"),
			want:    want{errContains: "failed to get data: connection lost"},
		},
		{
			name: "empty_item",
			id:   "6",
			response: &proto.GetDataResponse{
				Id:   "6",
				Item: nil,
			},
			want: want{errContains: "empty item in server response"},
		},
		{
			name: "unknown_type",
			id:   "7",
			response: &proto.GetDataResponse{
				Id: "7",
				Item: &proto.Item{
					RecordType: "UNKNOWN",
					Meta:       `{}`,
					Data:       []byte{},
				},
			},
			want: want{errContains: "unknown data type: UNKNOWN"},
		},
		{
			name: "invalid_meta_json",
			id:   "8",
			response: &proto.GetDataResponse{
				Id: "8",
				Item: &proto.Item{
					RecordType: string(entity.Password),
					Meta:       `{"resource":`,
					Data:       []byte("secret"),
				},
			},
			want: want{errContains: "failed to decode metadata"},
		},
		{
			name: "invalid_bankcard_data_json",
			id:   "9",
			response: &proto.GetDataResponse{
				Id: "9",
				Item: &proto.Item{
					RecordType: string(entity.BankCard),
					Meta:       `{"bank":"test","comment":"card"}`,
					Data:       []byte(`{"number": "1234",`),
				},
			},
			want: want{errContains: "failed to decode data"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockVault := mc.NewMockVaultServiceClient(ctrl)

			mockVault.
				EXPECT().
				GetData(gomock.Any(), &proto.GetDataRequest{Id: tt.id}).
				Return(tt.response, tt.grpcErr)

			workDir := t.TempDir()
			defer os.RemoveAll(workDir)

			client := &grpc.Client{
				VaultClient: mockVault,
			}

			svc := NewService(client, workDir)

			gotType, gotData, err := svc.GetData(context.Background(), tt.id)

			if tt.want.errContains != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.want.errContains)
				assert.Empty(t, gotType)
				assert.Nil(t, gotData)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want.dataType, gotType)
			assert.NotNil(t, gotData)

			if tt.assertFn != nil {
				tt.assertFn(t, gotType, gotData, workDir)
			}
		})
	}
}

func TestService_AddText(t *testing.T) {
	type want struct {
		err string
	}

	tests := []struct {
		name   string
		data   string
		meta   entity.TextMeta
		addErr error
		want   want
	}{
		{
			name: "success",
			data: "hello world",
			meta: entity.TextMeta{
				Name:    "note 1",
				Comment: "test comment",
			},
			addErr: nil,
			want: want{
				err: "",
			},
		},
		{
			name: "grpc_status_error",
			data: "some data",
			meta: entity.TextMeta{
				Name:    "note 2",
				Comment: "another comment",
			},
			addErr: status.Error(codes.Internal, "internal error"),
			want: want{
				err: "failed to save data: internal error",
			},
		},
		{
			name: "native_error",
			data: "some data",
			meta: entity.TextMeta{
				Name:    "note 3",
				Comment: "third comment",
			},
			addErr: errors.New("connection lost"),
			want: want{
				err: "failed to save data: connection lost",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockVault := mc.NewMockVaultServiceClient(ctrl)

			mockVault.
				EXPECT().
				AddData(gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, req *proto.AddDataRequest, _ ...interface{}) (*proto.Empty, error) {
					assert.Equal(t, []byte(tt.data), req.GetItem().GetData())
					assert.Equal(t, string(entity.Text), req.GetItem().GetRecordType())

					var gotMeta entity.TextMeta
					err := json.Unmarshal([]byte(req.GetItem().GetMeta()), &gotMeta)
					assert.NoError(t, err)
					assert.Equal(t, tt.meta, gotMeta)

					return &proto.Empty{}, tt.addErr
				})

			grpcClient := &grpc.Client{
				VaultClient: mockVault,
			}

			svc := NewService(grpcClient, "")

			err := svc.AddText(context.Background(), tt.data, tt.meta)

			if tt.want.err != "" {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.want.err)
				return
			}

			assert.NoError(t, err)
		})
	}
}

func TestService_AddPassword(t *testing.T) {
	type want struct {
		err string
	}

	tests := []struct {
		name   string
		data   string
		meta   entity.PasswordMeta
		addErr error
		want   want
	}{
		{
			name: "success",
			data: "p@ssw0rd",
			meta: entity.PasswordMeta{
				Resource: "example.com",
				Login:    "user1",
				Comment:  "main account",
			},
			addErr: nil,
			want:   want{err: ""},
		},
		{
			name: "grpc_status_error",
			data: "123456",
			meta: entity.PasswordMeta{
				Resource: "example.org",
				Login:    "user2",
				Comment:  "secondary",
			},
			addErr: status.Error(codes.Internal, "internal error"),
			want: want{
				err: "failed to save data: internal error",
			},
		},
		{
			name: "native_error",
			data: "qwerty",
			meta: entity.PasswordMeta{
				Resource: "service.local",
				Login:    "user3",
				Comment:  "local",
			},
			addErr: errors.New("connection lost"),
			want: want{
				err: "failed to save data: connection lost",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockVault := mc.NewMockVaultServiceClient(ctrl)

			mockVault.
				EXPECT().
				AddData(gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, req *proto.AddDataRequest, _ ...interface{}) (*proto.Empty, error) {
					item := req.GetItem()
					assert.NotNil(t, item, "Item must not be nil")
					assert.Equal(t, []byte(tt.data), item.GetData())
					assert.Equal(t, string(entity.Password), item.GetRecordType())

					var gotMeta entity.PasswordMeta
					err := json.Unmarshal([]byte(item.GetMeta()), &gotMeta)
					assert.NoError(t, err)
					assert.Equal(t, tt.meta, gotMeta)

					return &proto.Empty{}, tt.addErr
				})

			grpcClient := &grpc.Client{
				VaultClient: mockVault,
			}

			svc := NewService(grpcClient, "")

			err := svc.AddPassword(context.Background(), tt.data, tt.meta)

			if tt.want.err != "" {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.want.err)
				return
			}

			assert.NoError(t, err)
		})
	}
}

func TestService_AddBankCard(t *testing.T) {
	type want struct {
		err string
	}

	tests := []struct {
		name   string
		data   entity.BankCardData
		meta   entity.BankCardMeta
		addErr error
		want   want
	}{
		{
			name: "success",
			data: entity.BankCardData{
				Number:     "4111111111111111",
				Holder:     "JOHN DOE",
				CSV:        "123",
				ValidMonth: 12,
				ValidYear:  2030,
			},
			meta: entity.BankCardMeta{
				Bank:    "Test Bank",
				Comment: "main card",
			},
			addErr: nil,
			want:   want{err: ""},
		},
		{
			name: "grpc_status_error",
			data: entity.BankCardData{
				Number:     "5555555555554444",
				Holder:     "ALICE",
				CSV:        "999",
				ValidMonth: 1,
				ValidYear:  2028,
			},
			meta: entity.BankCardMeta{
				Bank:    "Other Bank",
				Comment: "backup",
			},
			addErr: status.Error(codes.Internal, "internal error"),
			want: want{
				err: "failed to save data: internal error",
			},
		},
		{
			name: "native error",
			data: entity.BankCardData{
				Number:     "4000000000000002",
				Holder:     "BOB",
				CSV:        "777",
				ValidMonth: 5,
				ValidYear:  2027,
			},
			meta: entity.BankCardMeta{
				Bank:    "Local Bank",
				Comment: "test card",
			},
			addErr: errors.New("connection lost"),
			want: want{
				err: "failed to save data: connection lost",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockVault := mc.NewMockVaultServiceClient(ctrl)

			mockVault.
				EXPECT().
				AddData(gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, req *proto.AddDataRequest, _ ...interface{}) (*proto.Empty, error) {
					item := req.GetItem()
					assert.NotNil(t, item, "Item must not be nil")
					assert.Equal(t, string(entity.BankCard), item.GetRecordType())

					var gotData entity.BankCardData
					assert.NoError(t, json.Unmarshal(item.GetData(), &gotData))
					assert.Equal(t, tt.data, gotData)

					var gotMeta entity.BankCardMeta
					assert.NoError(t, json.Unmarshal([]byte(item.GetMeta()), &gotMeta))
					assert.Equal(t, tt.meta, gotMeta)

					return &proto.Empty{}, tt.addErr
				})

			grpcClient := &grpc.Client{
				VaultClient: mockVault,
			}

			svc := NewService(grpcClient, "")

			err := svc.AddBankCard(context.Background(), tt.data, tt.meta)

			if tt.want.err != "" {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.want.err)
				return
			}

			assert.NoError(t, err)
		})
	}
}

func TestService_AddFile(t *testing.T) {
	type want struct {
		errContains string
	}

	tests := []struct {
		name       string
		setupFile  func(t *testing.T, tempDir string) string
		comment    string
		addErr     error
		want       want
		expectCall bool
	}{
		{
			name: "file_not_exists",
			setupFile: func(t *testing.T, tempDir string) string {
				return filepath.Join(tempDir, "no_file.bin")
			},
			comment:    "test",
			expectCall: false,
			want: want{
				errContains: "file does not exist",
			},
		},
		{
			name: "success",
			setupFile: func(t *testing.T, tempDir string) string {
				path := filepath.Join(tempDir, "photo.jpg")
				err := os.WriteFile(path, []byte("file-bytes"), 0o644)
				assert.NoError(t, err)
				return path
			},
			comment:    "my comment",
			expectCall: true,
			addErr:     nil,
			want:       want{errContains: ""},
		},
		{
			name: "grpc_status_error",
			setupFile: func(t *testing.T, tempDir string) string {
				path := filepath.Join(tempDir, "doc.pdf")
				err := os.WriteFile(path, []byte("hello"), 0o644)
				assert.NoError(t, err)
				return path
			},
			comment:    "status",
			expectCall: true,
			addErr:     status.Error(codes.Internal, "internal fail"),
			want: want{
				errContains: "failed to save data: internal fail",
			},
		},
		{
			name: "native_error",
			setupFile: func(t *testing.T, tempDir string) string {
				path := filepath.Join(tempDir, "xx.mp4")
				err := os.WriteFile(path, []byte("123"), 0o644)
				assert.NoError(t, err)
				return path
			},
			comment:    "native",
			expectCall: true,
			addErr:     errors.New("connection lost"),
			want: want{
				errContains: "failed to save data: connection lost",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockVault := mc.NewMockVaultServiceClient(ctrl)

			tempDir := t.TempDir()
			filePath := tt.setupFile(t, tempDir)
			defer os.RemoveAll(tempDir)

			if tt.expectCall {
				mockVault.
					EXPECT().
					AddData(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, req *proto.AddDataRequest, _ ...interface{}) (*proto.Empty, error) {
						item := req.GetItem()
						assert.NotNil(t, item, "Item must not be nil")

						content, err := os.ReadFile(filePath)
						assert.NoError(t, err)
						assert.Equal(t, content, item.GetData())
						assert.Equal(t, string(entity.File), item.GetRecordType())

						var gotMeta entity.FileMeta
						err = json.Unmarshal([]byte(item.GetMeta()), &gotMeta)
						assert.NoError(t, err)

						assert.Equal(t, filepath.Base(filePath), gotMeta.Name)
						assert.Equal(t, filepath.Ext(filePath), gotMeta.Extension)
						assert.Equal(t, tt.comment, gotMeta.Comment)

						return &proto.Empty{}, tt.addErr
					})
			} else {
				mockVault.EXPECT().AddData(gomock.Any(), gomock.Any()).Times(0)
			}

			svc := NewService(
				&grpc.Client{VaultClient: mockVault},
				"",
			)

			err := svc.AddFile(context.Background(), filePath, tt.comment)

			if tt.want.errContains != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.want.errContains)
				return
			}

			assert.NoError(t, err)
		})
	}
}
