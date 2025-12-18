package handler

import (
	"context"
	"encoding/hex"
	"errors"

	"github.com/msorokin-hash/passkeeper/internal/entity"
	"github.com/msorokin-hash/passkeeper/internal/server/jwt"
	"github.com/msorokin-hash/passkeeper/internal/storage"
	"github.com/msorokin-hash/passkeeper/pkg/utils"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	errorscustom "github.com/msorokin-hash/passkeeper/internal/errors"
	proto "github.com/msorokin-hash/passkeeper/internal/protobuf"
)

// GRPCUserHandler handles gRPC requests related to user registration
// and authentication.
type GRPCUserHandler struct {
	proto.UnimplementedUserServiceServer
	masterKey    string
	storage      storage.Storage
	tokenService jwt.TokenService
	log          *logrus.Entry
}

// NewGRPCUserHandler creates and returns a new gRPC handler
// for user-related operations.
func NewGRPCUserHandler(
	masterKey string,
	storage storage.Storage,
	tokenService jwt.TokenService,
	log *logrus.Entry,
) *GRPCUserHandler {
	return &GRPCUserHandler{
		masterKey:    masterKey,
		storage:      storage,
		tokenService: tokenService,
		log:          log,
	}
}

// Register creates a new user and returns an authentication token.
// Returns an error if the login is taken or the input is invalid.
func (h *GRPCUserHandler) Register(ctx context.Context, in *proto.RegisterRequest) (*proto.AuthResponse, error) {
	if in.Login == "" || in.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "login or password is empty")
	}

	hash, err := utils.GeneratePasswordHash(in.GetPassword())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%s", err.Error())
	}

	randomSecret, err := utils.GenerateRandomSecretKey()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%s", err.Error())
	}

	encryptedSecret, err := utils.Encrypt(randomSecret, h.masterKey)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%s", err.Error())
	}

	user := &entity.User{
		Login:           in.GetLogin(),
		PasswordHash:    hash,
		EncryptedSecret: hex.EncodeToString(encryptedSecret),
	}

	userID, err := h.storage.CreateUser(ctx, user)
	if err != nil {
		switch {
		case errors.Is(err, errorscustom.ErrLoginTaken):
			return nil, status.Error(codes.InvalidArgument, "user already exists")
		default:
			h.log.WithError(err).Error("error while creating user")
			return nil, status.Errorf(codes.Internal, "%s", err.Error())
		}
	}

	user.ID = userID

	token, err := h.tokenService.CreateToken(user)
	if err != nil {
		h.log.WithError(err).Error("error while creating token")
		return nil, status.Errorf(codes.Internal, "%s", err.Error())
	}

	respose := &proto.AuthResponse{Token: token}

	return respose, nil
}

// Login authenticates a user and returns a new authentication token.
// Returns an error if the credentials are invalid or the user is not found.
func (h *GRPCUserHandler) Login(ctx context.Context, in *proto.LoginRequest) (*proto.AuthResponse, error) {
	if in.Login == "" || in.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "login or password is empty")
	}

	user, err := h.storage.GetUserByLogin(ctx, in.GetLogin())
	if err != nil {
		switch {
		case errors.Is(err, errorscustom.ErrNoUser):
			return nil, status.Error(codes.NotFound, "user already exists")
		default:
			h.log.WithError(err).Error("error while getting user by login")
			return nil, status.Errorf(codes.Internal, "%s", err.Error())
		}
	}

	ok := utils.ComparePwdAndHash(in.GetPassword(), user.PasswordHash)
	if !ok {
		return nil, status.Error(codes.InvalidArgument, "invalid user credentials")
	}

	token, err := h.tokenService.CreateToken(user)
	if err != nil {
		h.log.WithError(err).Error("error while creating token")
		return nil, status.Errorf(codes.Internal, "%s", err.Error())
	}

	respose := &proto.AuthResponse{Token: token}

	return respose, nil
}
