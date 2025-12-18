package middleware

import (
	"context"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

func TestLoggerInterceptor(t *testing.T) {
	logger, hook := test.NewNullLogger()
	entry := logrus.NewEntry(logger)

	interceptor := LoggerInterceptor(entry)

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "response", nil
	}

	info := &grpc.UnaryServerInfo{
		FullMethod: "/gophkeeper.v1.VaultService/GetData",
	}

	resp, err := interceptor(context.Background(), "request", info, handler)

	assert.Equal(t, "response", resp)
	assert.NoError(t, err)
	assert.Nil(t, err, "Expected no error")

	found := false
	for _, entry := range hook.AllEntries() {
		if entry.Message == "gRPC request" {
			found = true
			assert.Equal(t, "/gophkeeper.v1.VaultService/GetData", entry.Data["method"])
			assert.Equal(t, codes.OK.String(), entry.Data["code"])
			duration, ok := entry.Data["duration"].(float64)
			assert.True(t, ok)
			assert.GreaterOrEqual(t, duration, 0.0)
		}
	}

	assert.True(t, found, "Expected log entry not found")
}
