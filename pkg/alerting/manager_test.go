package alerting

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/manu/octo/pkg/checker"
	"github.com/manu/octo/pkg/config"
)

// MockProvider for testing
type MockProvider struct {
	SentCount int
	LastRule  config.AlertRule
	Done      chan bool
}

func (m *MockProvider) Send(ctx context.Context, channel config.AlertChannel, rule config.AlertRule, endpoint config.EndpointConfig, result *checker.Result) error {
	m.SentCount++
	m.LastRule = rule
	if m.Done != nil {
		m.Done <- true
	}
	return nil
}

func TestManager_Evaluate(t *testing.T) {
	// ... (setup code remains same until provider init) ...
	tmpConfigFile, err := os.CreateTemp("", "config-*.yml")
	if err != nil {
		t.Fatalf("Failed to create temp config: %v", err)
	}
	defer os.Remove(tmpConfigFile.Name())

	// Write initial config
	initialConfig := `
alert_channels:
  - name: "test-webhook"
    type: "webhook"
    url: "http://localhost"

alert_rules:
  - name: "Production Down"
    condition: "success == false"
    channels: ["test-webhook"]
    tags:
      env: "prod"
`
	if _, err := tmpConfigFile.WriteString(initialConfig); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}
	tmpConfigFile.Close()

	// 2. Init Managers
	cfgMgr, err := config.NewManager(tmpConfigFile.Name())
	if err != nil {
		t.Fatalf("Failed to create config manager: %v", err)
	}

	am := NewManager(cfgMgr)
	mockProvider := &MockProvider{Done: make(chan bool, 1)}
	am.RegisterProvider("webhook", mockProvider)

	// 3. Test Case 1: Matching Tag & Condition
	endpoint1 := config.EndpointConfig{
		ID:   "ep1",
		Name: "Prod Check",
		Tags: map[string]string{"env": "prod"},
	}
	result1 := &checker.Result{Success: false}

	am.Evaluate(context.Background(), endpoint1, result1)

	select {
	case <-mockProvider.Done:
		// continued
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for alert")
	}

	if mockProvider.SentCount != 1 {
		t.Errorf("Expected 1 alert, got %d", mockProvider.SentCount)
	}

	// 4. Test Case 2: Matching Tag but Condition False (Success)
	result2 := &checker.Result{Success: true}
	// Reset active alerts manually? Or just rely on logic.
	// If success is true, it should RESOLVE the alert (if logic implemented) or just do nothing.
	// Current impl:
	// else { if wasActive { delete } }

	am.Evaluate(context.Background(), endpoint1, result2)
	// SentCount should NOT increase
	if mockProvider.SentCount != 1 {
		t.Errorf("Expected sent count to remain 1, got %d", mockProvider.SentCount)
	}

	// 5. Test Case 3: Mismatch Tag
	endpoint2 := config.EndpointConfig{
		ID:   "ep2",
		Name: "Dev Check",
		Tags: map[string]string{"env": "dev"},
	}
	result3 := &checker.Result{Success: false}

	am.Evaluate(context.Background(), endpoint2, result3)
	if mockProvider.SentCount != 1 {
		t.Errorf("Expected sent count to remain 1 (tag mismatch), got %d", mockProvider.SentCount)
	}
}
