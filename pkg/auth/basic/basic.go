package basic

import (
	"golang.org/x/crypto/bcrypt"

	"github.com/manu/octo/pkg/auth"
)

// UserCredential stores the username, bcrypt hash, and role
type UserCredential struct {
	Username     string
	PasswordHash string
	Role         string
}

// Provider implements auth.Provider for basic local auth
type Provider struct {
	users []UserCredential
}

// NewProvider creates a new basic auth provider with a set of users
func NewProvider(users []UserCredential) *Provider {
	return &Provider{
		users: users,
	}
}

// Authenticate checks the username and password against bcrypt hashes
func (p *Provider) Authenticate(username, password string) (*auth.User, error) {
	for _, u := range p.users {
		if u.Username == username {
			err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
			if err == nil {
				return &auth.User{
					Username: u.Username,
					Role:     u.Role,
				}, nil
			}
			return nil, auth.ErrInvalidCredentials
		}
	}
	return nil, auth.ErrInvalidCredentials
}
