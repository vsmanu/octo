package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/manu/octo/pkg/auth"
	"github.com/manu/octo/pkg/auth/basic"
)

type contextKey string

const (
	UserContextKey contextKey = "user"
	CookieName     string     = "octo_session"
)

// LoginRequest represents the login payload
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse represents the login response
type LoginResponse struct {
	Token string `json:"token"`
}

// handleLogin handles user authentication
func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	cfg := s.configManager.GetConfig()
	if !cfg.Auth.Enabled {
		http.Error(w, "Authentication is disabled", http.StatusBadRequest)
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Determine provider
	// For MVP, we only support basic auth provider defined in config
	var provider auth.Provider
	if cfg.Auth.Provider == "basic" {
		provider = basic.NewProvider(cfg.Auth.Basic.Username, cfg.Auth.Basic.Password)
	} else {
		http.Error(w, "Unknown auth provider", http.StatusInternalServerError)
		return
	}

	user, err := provider.Authenticate(req.Username, req.Password)
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Generate JWT
	token, err := s.generateToken(user, cfg.Auth.Secret)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	// Set Cookie
	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Value:    token,
		Path:     "/",
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(LoginResponse{Token: token})
}

// handleLogout clears the session cookie
func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Value:    "",
		Path:     "/",
		Expires:  time.Now().Add(-1 * time.Hour),
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
	w.WriteHeader(http.StatusOK)
}

// AuthMiddleware authenticates requests via Cookie or Bearer Token
func (s *Server) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg := s.configManager.GetConfig()
		if !cfg.Auth.Enabled {
			next.ServeHTTP(w, r)
			return
		}

		// Public routes that don't need auth (besides those handled outside this middleware)
		// Since we wrap /api/v1/, we need to exclude login if it falls under this tree
		if r.URL.Path == "/api/v1/login" {
			next.ServeHTTP(w, r)
			return
		}

		// Check Cookie first
		tokenString := ""
		cookie, err := r.Cookie(CookieName)
		if err == nil {
			tokenString = cookie.Value
		}

		// Fallback to Bearer Token
		if tokenString == "" {
			authHeader := r.Header.Get("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") {
				tokenString = strings.TrimPrefix(authHeader, "Bearer ")
			}
		}

		if tokenString == "" {
			s.handleUnauthenticated(w, r)
			return
		}

		// Validate Token
		claims := jwt.MapClaims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(cfg.Auth.Secret), nil
		})

		if err != nil || !token.Valid {
			s.handleUnauthenticated(w, r)
			return
		}

		// Add user to context
		ctx := context.WithValue(r.Context(), UserContextKey, claims["sub"])
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (s *Server) handleUnauthenticated(w http.ResponseWriter, r *http.Request) {
	// If it's an API request, return 401
	if strings.HasPrefix(r.URL.Path, "/api/") {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	// If it's a browser request (page load), we might want to let the frontend handle the redirect
	// But since we are serving the SPA from the same server, we can serve index.html
	// and let the React app check /api/v1/me or handle the 401
	// Actually, the middleware wraps the API routes.
	// For static assets/SPA, we might validly want to serve them without auth
	// so the login page can load!
	// So we should probably NOT apply this middleware to the static file server part in server.go
	http.Error(w, "Unauthorized", http.StatusUnauthorized)
}

func (s *Server) generateToken(user *auth.User, secret string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  user.Username,
		"role": user.Role,
		"exp":  time.Now().Add(24 * time.Hour).Unix(),
	})

	// If secret is empty, use a default (INSECURE for prod, but prevents crash)
	// Ideally we should enforce secret presence
	if secret == "" {
		secret = "default-insecure-secret-change-me"
	}

	return token.SignedString([]byte(secret))
}

// handleMe returns current user info
func (s *Server) handleMe(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(UserContextKey)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"username": user,
	})
}
