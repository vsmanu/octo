package checker

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/manu/octo/pkg/config"
)

func TestChecker_Check_SSL(t *testing.T) {
	// Create a test server with TLS
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	c := NewChecker()
	// Hack to allow self-signed certs for testing since we can't easily inject config yet
	transport := c.client.Transport.(*http.Transport)
	transport.TLSClientConfig.InsecureSkipVerify = true

	endpoint := config.EndpointConfig{
		ID:     "test-ssl",
		URL:    ts.URL,
		Method: "GET",
		Validation: config.ValidationConfig{
			StatusCodes: []int{200},
		},
	}

	result := c.Check(context.Background(), endpoint)

	if !result.Success {
		t.Errorf("Expected success, got failure: %s", result.Error)
	}
	if result.Error != "" {
		t.Errorf("Expected no error, got: %s", result.Error)
	}
	if result.CertExpiry.IsZero() {
		t.Error("Expected CertExpiry to be set")
	}
	if result.CertIssuer == "" {
		t.Error("Expected CertIssuer to be set")
	}
	if result.CertSubject == "" {
		t.Error("Expected CertSubject to be set")
	}

	// Basic check that dates are reasonable
	if !result.CertNotAfter.After(time.Now()) {
		t.Error("Expected CertNotAfter to be in the future")
	}
	// httptest certs are usually valid starting now or slightly before
	if !result.CertNotBefore.Before(time.Now().Add(time.Hour)) {
		t.Error("Expected CertNotBefore to be reasonable")
	}
}
