package context

import (
	"context"
	"testing"
)

func TestGetUserFromContext(t *testing.T) {
	tests := []struct {
		name     string
		ctx      context.Context
		expected *UserContext
	}{
		{
			name:     "user exists in context",
			ctx:      context.WithValue(context.Background(), userContextKey, &UserContext{ID: "123", Login: "test", Secret: "secret"}),
			expected: &UserContext{ID: "123", Login: "test", Secret: "secret"},
		},
		{
			name:     "user is nil in context",
			ctx:      context.WithValue(context.Background(), userContextKey, nil),
			expected: nil,
		},
		{
			name:     "wrong type in context",
			ctx:      context.WithValue(context.Background(), userContextKey, "wrong type"),
			expected: nil,
		},
		{
			name:     "no user in context",
			ctx:      context.Background(),
			expected: nil,
		},
		{
			name:     "different key in context",
			ctx:      context.WithValue(context.Background(), "differentKey", &UserContext{ID: "123"}),
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetUserFromContext(tt.ctx)

			if tt.expected == nil && result != nil {
				t.Errorf("Expected nil, got %v", result)
				return
			}

			if tt.expected != nil && result == nil {
				t.Errorf("Expected %v, got nil", tt.expected)
				return
			}

			if tt.expected != nil && result != nil {
				if tt.expected.ID != result.ID ||
					tt.expected.Login != result.Login ||
					tt.expected.Secret != result.Secret {
					t.Errorf("Expected %+v, got %+v", tt.expected, result)
				}
			}
		})
	}
}
