package middleware

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	proto "github.com/msorokin-hash/passkeeper/internal/protobuf"
	"github.com/msorokin-hash/passkeeper/pkg/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// TokenInterceptor returns a gRPC unary client interceptor that injects a JWT
// stored under workDir into the Authorization header of outgoing requests:
//
//	Authorization: Bearer <token>
//
// The token is added to all requests except those belonging to UserService
// (e.g., login or registration), since those endpoints must be called without
// authentication.
//
// If the server responds with codes.Unauthenticated, the interceptor removes
// the token file from disk, effectively logging the user out.
func TokenInterceptor(workDir string) grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req interface{},
		reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		tokenFilePath := filepath.Join(workDir, ".token.jwt")

		token, err := utils.ReadFile(tokenFilePath)
		if err != nil {
			log.Printf("cannot load token from workdir %q: %v", workDir, err)
		}

		token = strings.TrimSpace(token)
		if token != "" && !isUserServiceMethod(method) {
			authHeaderValue := fmt.Sprintf("Bearer %s", token)
			ctx = metadata.AppendToOutgoingContext(ctx, "authorization", authHeaderValue)
		}

		err = invoker(ctx, method, req, reply, cc, opts...)
		if err != nil {
			if s, ok := status.FromError(err); ok && s.Code() == codes.Unauthenticated {
				log.Println("Unauthenticated received, clearing token")

				if rmErr := os.Remove(tokenFilePath); rmErr != nil && !os.IsNotExist(rmErr) {
					log.Printf("error removing token file %q: %v", tokenFilePath, rmErr)
				}
			}
		}

		return err
	}
}

// isUserServiceMethod reports whether the invoked gRPC method belongs
// to UserService. The method name has the form:
//
//	/package.Service/Method
func isUserServiceMethod(fullMethod string) bool {
	parts := strings.Split(fullMethod, "/")
	if len(parts) < 2 {
		return false
	}

	service := parts[1]

	return service == proto.UserService_ServiceDesc.ServiceName
}
