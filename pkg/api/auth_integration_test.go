package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/manu/octo/pkg/checker"
	"github.com/manu/octo/pkg/config"
	"github.com/manu/octo/pkg/storage"
)

// MockStorage implements storage.Provider
type MockStorage struct{}

func (m *MockStorage) WriteResult(result checker.Result) error {
	return nil
}

func (m *MockStorage) QueryHistory(ctx context.Context, endpointID string, from, to time.Time) ([]storage.Metric, error) {
	return []storage.Metric{}, nil
}

func (m *MockStorage) Close() {}

func TestAuthWorkflow(t *testing.T) {
	// 1. Create temporary config file
	configContent := `
global:
  check_interval: 60s
  request_timeout: 10s
auth:
  enabled: true
  provider: basic
  secret: "test-secret"
  basic:
    username: "testuser"
    password: "testpassword"
endpoints: []
satellites: []
alert_channels: []
alert_rules: []
`
	tmpConfig, err := os.CreateTemp("", "config-*.yml")
	if err != nil {
		t.Fatalf("Failed to create temp config: %v", err)
	}
	defer os.Remove(tmpConfig.Name())

	if _, err := tmpConfig.Write([]byte(configContent)); err != nil {
		t.Fatalf("Failed to write to temp config: %v", err)
	}
	if err := tmpConfig.Close(); err != nil {
		t.Fatalf("Failed to close temp config: %v", err)
	}

	// 2. Initialize Config Manager
	cfgMgr, err := config.NewManager(tmpConfig.Name())
	if err != nil {
		t.Fatalf("Failed to load config manager: %v", err)
	}

	// 3. Initialize Server with Mock Storage
	mockStorage := &MockStorage{}
	server := NewServer(cfgMgr, mockStorage, nil, nil) // frontendFS is nil for API tests

	// 4. Test Login (Success)
	loginPayload := map[string]string{
		"username": "testuser",
		"password": "testpassword",
	}
	payloadBytes, _ := json.Marshal(loginPayload)
	req := httptest.NewRequest("POST", "/api/v1/login", bytes.NewBuffer(payloadBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d. Body: %s", w.Code, w.Body.String())
	}

	var loginResp LoginResponse
	if err := json.NewDecoder(w.Body).Decode(&loginResp); err != nil {
		t.Fatalf("Failed to decode login response: %v", err)
	}

	if loginResp.Token == "" {
		t.Error("Expected token in response, got empty string")
	}

	token := loginResp.Token

	// 5. Test Login (Failure)
	badPayload := map[string]string{
		"username": "testuser",
		"password": "wrongpassword",
	}
	payloadBytes, _ = json.Marshal(badPayload)
	req = httptest.NewRequest("POST", "/api/v1/login", bytes.NewBuffer(payloadBytes))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	server.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 Unauthorized for bad password, got %d", w.Code)
	}

	// 6. Test Protected Route (Success with Token)
	req = httptest.NewRequest("GET", "/api/v1/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()

	server.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200 OK for valid token, got %d. Body: %s", w.Code, w.Body.String())
	}

	// 7. Test Protected Route (Failure without Token)
	req = httptest.NewRequest("GET", "/api/v1/auth/me", nil)
	w = httptest.NewRecorder()

	server.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 Unauthorized for missing token, got %d", w.Code)
	}
}
