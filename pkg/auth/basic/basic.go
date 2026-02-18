package basic

import (
	"crypto/subtle"

	"github.com/manu/octo/pkg/auth"
)

// Provider implements auth.Provider for basic auth
type Provider struct {
	username string
	password string
}

// NewProvider creates a new basic auth provider
func NewProvider(username, password string) *Provider {
	return &Provider{
		username: username,
		password: password,
	}
}

// Authenticate checks the username and password
func (p *Provider) Authenticate(username, password string) (*auth.User, error) {
	// constant time comparison to prevent timing attacks
	if subtle.ConstantTimeCompare([]byte(username), []byte(p.username)) == 1 &&
		subtle.ConstantTimeCompare([]byte(password), []byte(p.password)) == 1 {
		return &auth.User{
			Username: username,
			Role:     "admin", // Default to admin for basic auth
		}, nil
	}
	return nil, auth.ErrInvalidCredentials
}
