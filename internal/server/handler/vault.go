package handler

import (
	"context"
	"errors"

	"github.com/msorokin-hash/passkeeper/internal/entity"
	errorscustom "github.com/msorokin-hash/passkeeper/internal/errors"
	proto "github.com/msorokin-hash/passkeeper/internal/protobuf"
	cx "github.com/msorokin-hash/passkeeper/internal/server/context"
	"github.com/msorokin-hash/passkeeper/internal/storage"
	"github.com/msorokin-hash/passkeeper/pkg/utils"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var validDataTypes = map[entity.DataType]bool{
	entity.Password: true,
	entity.Text:     true,
	entity.BankCard: true,
	entity.File:     true,
}

// GRPCVaultHandler handles gRPC requests for the secrets vault service.
// It provides operations for adding, retrieving, updating, and deleting
// encrypted user data.
type GRPCVaultHandler struct {
	proto.UnimplementedVaultServiceServer
	masterKey string
	storage   storage.Storage
	log       *logrus.Entry
}

// NewGRPCVaultHandler creates and returns a new gRPC handler
// for interacting with the secrets vault.
func NewGRPCVaultHandler(
	masterKey string,
	storage storage.Storage,
	log *logrus.Entry,
) *GRPCVaultHandler {
	return &GRPCVaultHandler{
		masterKey: masterKey,
		storage:   storage,
		log:       log,
	}
}

// GetData retrieves encrypted user data, decrypts it,
// and returns the resulting vault item.
// Returns errors when the ID is empty, the user is unauthenticated,
// data is not found, or decryption fails.
func (h *GRPCVaultHandler) GetData(ctx context.Context, in *proto.GetDataRequest) (*proto.GetDataResponse, error) {
	if in.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "id is empty")
	}

	user := cx.GetUserFromContext(ctx)
	if user == nil {
		return nil, status.Error(codes.Unauthenticated, "user is not authenticated")
	}

	data, err := h.storage.GetItem(ctx, in.GetId(), user.ID)
	if err != nil {
		switch {
		case errors.Is(err, errorscustom.ErrNoData):
			return nil, status.Error(codes.NotFound, "data not found")
		default:
			h.log.WithError(err).Error("error while getting data")
			return nil, status.Error(codes.Internal, "error while getting data")
		}
	}

	dec, err := utils.Decrypt(data.EncryptedData, user.Secret)
	if err != nil {
		h.log.WithError(err).Error("error while decrypting data")
		return nil, status.Error(codes.Internal, "error while decrypting data")
	}

	response := &proto.GetDataResponse{
		Id: data.ID,
		Item: &proto.Item{
			Data:       dec,
			RecordType: string(data.Type),
			Meta:       data.Meta,
		},
	}

	return response, nil
}

// GetAllByType retrieves all vault items of a specific type for the authenticated user.
// Returns an error if the user is unauthenticated or the storage query fails.
func (h *GRPCVaultHandler) GetAllByType(ctx context.Context, in *proto.GetAllByTypeRequest) (*proto.GetAllByTypeResponse, error) {
	dataType := in.GetRecordType()
	if !isValidDataType(dataType) {
		return nil, status.Error(
			codes.InvalidArgument,
			"invalid data type",
		)
	}

	user := cx.GetUserFromContext(ctx)
	if user == nil {
		return nil, status.Error(codes.Unauthenticated, "user is not authenticated")
	}

	items, err := h.storage.GetItemsByType(ctx, dataType, user.ID)
	if err != nil {
		h.log.WithError(err).Error("Error while deleting item")
		return nil, status.Error(codes.Internal, "Internal server error")
	}

	var responseItems []*proto.GetAllByTypeResponse_TypeItem
	for _, item := range items {
		responseItems = append(responseItems, &proto.GetAllByTypeResponse_TypeItem{
			Id:   item.ID,
			Meta: item.Meta,
		})
	}
	return &proto.GetAllByTypeResponse{
		Items: responseItems,
	}, nil
}

// DeleteData removes a stored vault item by its ID.
// Returns errors for missing IDs, unauthenticated users,
// missing items, or storage failures.
func (h *GRPCVaultHandler) DeleteData(ctx context.Context, in *proto.DeleteDataRequest) (*proto.Empty, error) {
	if in.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "id is empty")
	}

	user := cx.GetUserFromContext(ctx)
	if user == nil {
		return nil, status.Error(codes.Unauthenticated, "user is not authenticated")
	}

	err := h.storage.DeleteItem(ctx, in.GetId(), user.ID)
	if err != nil {
		switch {
		case errors.Is(err, errorscustom.ErrNoData):
			return nil, status.Error(codes.NotFound, "data not found")
		default:
			h.log.WithError(err).Error("error while deleting data")
			return nil, status.Error(codes.Internal, "error while deleting data")
		}
	}

	return &proto.Empty{}, nil
}

// AddData encrypts the provided data and stores it as a new vault item.
// Returns errors for missing input, unauthenticated users,
// encryption failures, or storage issues.
func (h *GRPCVaultHandler) AddData(ctx context.Context, in *proto.AddDataRequest) (*proto.Empty, error) {
	item := in.GetItem()
	if item == nil {
		return nil, status.Error(codes.InvalidArgument, "no data stored")
	}

	user := cx.GetUserFromContext(ctx)
	if user == nil {
		return nil, status.Error(codes.Unauthenticated, "user is not authenticated")
	}

	typ := in.Item.GetRecordType()

	enc, err := utils.Encrypt(item.GetData(), user.Secret)
	if err != nil {
		h.log.WithError(err).Error("error while encrypting data")
		return nil, status.Error(codes.Internal, "error while encrypting data")
	}

	vault := &entity.VaultItem{
		UserID:        user.ID,
		Meta:          item.GetMeta(),
		Type:          entity.DataType(typ),
		EncryptedData: enc,
	}

	_, err = h.storage.CreateItem(ctx, user.ID, vault)
	if err != nil {
		h.log.WithError(err).Error("error while creating vault secret")
		return nil, status.Error(codes.Internal, "error while creating vault secret")
	}

	return &proto.Empty{}, nil
}

// UpdateData encrypts and updates an existing vault item by ID.
// Returns errors for missing IDs, unauthenticated users,
// encryption failures, missing items, or storage errors.
func (h *GRPCVaultHandler) UpdateData(ctx context.Context, in *proto.UpdateDataRequest) (*proto.Empty, error) {
	if in.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "id is empty")
	}

	user := cx.GetUserFromContext(ctx)
	if user == nil {
		return nil, status.Error(codes.Unauthenticated, "user is not authenticated")
	}

	enc, err := utils.Encrypt(in.GetData(), user.Secret)
	if err != nil {
		h.log.WithError(err).Error("error while encrypting data")
		return nil, status.Error(codes.Internal, "error while encrypting data")
	}

	item := &entity.VaultItem{
		ID:            in.GetId(),
		UserID:        user.ID,
		Meta:          in.GetMeta(),
		EncryptedData: enc,
	}

	err = h.storage.UpdateItem(ctx, in.GetId(), user.ID, item)
	if err != nil {
		switch {
		case errors.Is(err, errorscustom.ErrNoData):
			return nil, status.Error(codes.NotFound, "data not found")
		default:
			h.log.WithError(err).Error("error while updating data")
			return nil, status.Error(codes.Internal, "error while updating data")
		}
	}

	return &proto.Empty{}, nil
}

// isValidDataType checks if the given string represents a valid data type.
// It converts the string to DataType and checks against predefined valid types.
// Valid types are: Password, Text, BankCard, and File.
// Returns true if the type is valid, false otherwise.
func isValidDataType(dataType string) bool {
	return validDataTypes[entity.DataType(dataType)]
}
