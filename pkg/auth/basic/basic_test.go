package basic

import (
	"testing"

	"github.com/manu/octo/pkg/auth"
	"golang.org/x/crypto/bcrypt"
)

func TestBasicProvider(t *testing.T) {
	// Generate a bcrypt hash for "secret"
	hash, err := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	users := []UserCredential{
		{
			Username:     "admin",
			PasswordHash: string(hash),
			Role:         "admin",
		},
		{
			Username:     "viewer",
			PasswordHash: string(hash),
			Role:         "viewer",
		},
	}

	p := NewProvider(users)

	tests := []struct {
		name        string
		user        string
		pass        string
		expectedErr error
	}{
		{"ValidAdmin", "admin", "secret", nil},
		{"ValidViewer", "viewer", "secret", nil},
		{"InvalidUser", "unknown", "secret", auth.ErrInvalidCredentials},
		{"InvalidPass", "admin", "wrong", auth.ErrInvalidCredentials},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := p.Authenticate(tc.user, tc.pass)
			if err != tc.expectedErr {
				t.Fatalf("Expected err %v, got %v", tc.expectedErr, err)
			}
		})
	}
}
