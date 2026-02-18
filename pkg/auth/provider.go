package auth

import "errors"

// User represents an authenticated user
type User struct {
	Username string   `json:"username"`
	Role     string   `json:"role"` // e.g., "admin", "viewer"
	Groups   []string `json:"groups,omitempty"`
}

// Provider defines the interface for authentication providers
type Provider interface {
	// Authenticate validates credentials and returns a user
	Authenticate(username, password string) (*User, error)
}

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotFound       = errors.New("user not found")
)
