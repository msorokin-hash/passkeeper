package middleware

import (
	"context"
	"encoding/hex"
	"strings"

	"github.com/msorokin-hash/passkeeper/internal/server/jwt"
	"github.com/msorokin-hash/passkeeper/internal/storage"
	"github.com/msorokin-hash/passkeeper/pkg/utils"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	cx "github.com/msorokin-hash/passkeeper/internal/server/context"
)

// AuthInterceptor provides a gRPC interceptor responsible for handling
// authentication and user context initialization. It validates incoming JWT
// tokens, loads the corresponding user from storage, decrypts the user's
// secret using the master key, and injects user data into the request context.
type AuthInterceptor struct {
	masterKey    string
	storage      storage.Storage
	tokenService jwt.TokenService
	log          *logrus.Entry
}

// NewAuthInterceptor creates a new AuthInterceptor instance. It requires a
// master key for decrypting user secrets, a user storage implementation,
// a JWT token service for claim extraction, and a logger for error tracking.
func NewAuthInterceptor(
	masterKey string,
	storage storage.Storage,
	tokenService jwt.TokenService,
	log *logrus.Entry,
) *AuthInterceptor {
	return &AuthInterceptor{
		masterKey:    masterKey,
		storage:      storage,
		tokenService: tokenService,
		log:          log,
	}
}

// AuthenticateUser is a gRPC unary interceptor that enforces authentication
// for protected VaultService methods. It extracts the JWT token from the
// "Authorization" header, validates it, retrieves the associated user from
// storage, decrypts the user's secret, and attaches the resolved user context
// to the request. If authentication fails, an appropriate gRPC error is
// returned. Non-protected methods are passed through without checks.
func (i *AuthInterceptor) AuthenticateUser(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {

	needTokenMethods := map[string]bool{
		"/gophkeeper.v1.VaultService/AddData":      true,
		"/gophkeeper.v1.VaultService/GetData":      true,
		"/gophkeeper.v1.VaultService/DeleteData":   true,
		"/gophkeeper.v1.VaultService/UpdateData":   true,
		"/gophkeeper.v1.VaultService/GetAllByType": true,
	}

	if !needTokenMethods[info.FullMethod] {
		return handler(ctx, req)
	}

	var tokenStr string

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing metadata")
	}

	if authHeaders := md.Get("authorization"); len(authHeaders) > 0 {
		if strings.HasPrefix(authHeaders[0], "Bearer ") {
			tokenStr = strings.TrimPrefix(authHeaders[0], "Bearer ")
		}
	}

	if tokenStr == "" {
		return nil, status.Error(codes.Unauthenticated, "token is empty")
	}

	userData, err := i.tokenService.GetClaims(tokenStr)
	if err != nil {
		i.log.WithError(err).Error("failed to get claims from jwt")
		return nil, status.Error(codes.Unauthenticated, "invalid token")
	}

	user, err := i.storage.GetUserByLogin(ctx, userData.Login)
	if err != nil {
		i.log.WithError(err).Error("error while getting user by login")
		return nil, status.Error(codes.Internal, "error while getting user by login")
	}

	userSecretBytes, err := hex.DecodeString(user.EncryptedSecret)
	if err != nil {
		i.log.WithError(err).Error("failed to decode secret")
		return nil, status.Error(codes.Internal, "failed to decode secret")
	}

	decodedUserSecret, err := utils.Decrypt(userSecretBytes, i.masterKey)
	if err != nil {
		i.log.WithError(err).Error("failed to decrypt secret")
		return nil, status.Error(codes.Internal, "failed to decrypt secret")
	}

	ctx = cx.WrapContextWithUser(ctx, &cx.UserContext{
		ID:     user.ID,
		Login:  user.Login,
		Secret: hex.EncodeToString(decodedUserSecret),
	})

	return handler(ctx, req)
}
