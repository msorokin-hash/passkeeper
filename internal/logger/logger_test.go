package logger_test

import (
	"testing"

	"github.com/msorokin-hash/passkeeper/internal/logger"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestNewLogger(t *testing.T) {
	tests := []struct {
		name      string
		level     string
		wantLevel logrus.Level
		wantErr   bool
	}{
		{
			name:      "default level when empty",
			level:     "",
			wantLevel: logrus.InfoLevel,
			wantErr:   false,
		},
		{
			name:      "valid level debug",
			level:     "debug",
			wantLevel: logrus.DebugLevel,
			wantErr:   false,
		},
		{
			name:      "valid level warn",
			level:     "warn",
			wantLevel: logrus.WarnLevel,
			wantErr:   false,
		},
		{
			name:    "invalid level",
			level:   "invalid_level",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l, err := logger.NewLogger(tt.level)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, l)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, l)
			assert.Equal(t, tt.wantLevel, l.GetLevel())

			_, ok := l.Formatter.(*logrus.JSONFormatter)
			assert.True(t, ok, "logger formatter must be JSONFormatter")
		})
	}
}
