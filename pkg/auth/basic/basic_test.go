package basic

import (
	"testing"

	"github.com/manu/octo/pkg/auth"
)

func TestProvider_Authenticate(t *testing.T) {
	provider := NewProvider("admin", "secret")

	tests := []struct {
		name          string
		username      string
		password      string
		expectedError error
	}{
		{
			name:          "Valid credentials",
			username:      "admin",
			password:      "secret",
			expectedError: nil,
		},
		{
			name:          "Invalid username",
			username:      "wrong",
			password:      "secret",
			expectedError: auth.ErrInvalidCredentials,
		},
		{
			name:          "Invalid password",
			username:      "admin",
			password:      "wrong",
			expectedError: auth.ErrInvalidCredentials,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := provider.Authenticate(tt.username, tt.password)
			if err != tt.expectedError {
				t.Errorf("expected error %v, got %v", tt.expectedError, err)
			}
			if err == nil {
				if user.Username != tt.username {
					t.Errorf("expected username %s, got %s", tt.username, user.Username)
				}
				if user.Role != "admin" {
					t.Errorf("expected role admin, got %s", user.Role)
				}
			}
		})
	}
}
