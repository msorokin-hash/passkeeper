package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		valid    bool
	}{
		{
			name:     "only digits",
			password: "12345678",
			valid:    false,
		},
		{
			name:     "only letters",
			password: "password",
			valid:    false,
		},
		{
			name:     "letters and digits",
			password: "pass1234",
			valid:    true,
		},
		{
			name:     "mixed case letters and digits",
			password: "Pass1234",
			valid:    true,
		},
		{
			name:     "contains space",
			password: "Pass 1234",
			valid:    false,
		},
		{
			name:     "too short",
			password: "Pa123",
			valid:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePassword(tt.password)

			if tt.valid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}
