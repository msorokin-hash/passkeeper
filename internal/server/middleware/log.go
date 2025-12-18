package middleware

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// LoggerInterceptor returns a gRPC unary interceptor that logs incoming requests.
// For each call, it records the method name, execution duration, and resulting
// gRPC status code. The interceptor logs this information using the provided
// logrus Entry instance. Logging occurs after the handler finishes processing
// the request, regardless of whether it succeeded or failed.
func LoggerInterceptor(
	log *logrus.Entry,
) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler,
	) (interface{}, error) {
		start := time.Now()
		resp, err := handler(ctx, req)
		duration := time.Since(start).Seconds()
		status, _ := status.FromError(err)

		log.WithFields(logrus.Fields{
			"method":   info.FullMethod,
			"duration": duration,
			"code":     status.Code().String(),
		}).Info("gRPC request")

		return resp, err
	}
}
