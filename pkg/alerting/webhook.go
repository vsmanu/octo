package alerting

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"text/template"
	"time"

	"github.com/manu/octo/pkg/checker"
	"github.com/manu/octo/pkg/config"
)

// WebhookProvider implements the Provider interface for generic webhooks
type WebhookProvider struct {
	client *http.Client
}

// NewWebhookProvider creates a new WebhookProvider
func NewWebhookProvider() *WebhookProvider {
	return &WebhookProvider{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// AlertPayload is the data available to the template
type AlertPayload struct {
	Endpoint config.EndpointConfig
	Result   *checker.Result
	Rule     config.AlertRule
}

// Send sends an alert using the provided configuration
func (p *WebhookProvider) Send(ctx context.Context, channel config.AlertChannel, rule config.AlertRule, endpoint config.EndpointConfig, result *checker.Result) error {
	// 1. Prepare Payload
	payload := AlertPayload{
		Endpoint: endpoint,
		Result:   result,
		Rule:     rule,
	}

	// 2. Render Body Template
	tmpl, err := template.New("alert").Parse(channel.Body)
	if err != nil {
		return fmt.Errorf("failed to parse alert template: %w", err)
	}

	var body bytes.Buffer
	if err := tmpl.Execute(&body, payload); err != nil {
		return fmt.Errorf("failed to execute alert template: %w", err)
	}

	// 3. Create Request
	req, err := http.NewRequestWithContext(ctx, "POST", channel.URL, &body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// 4. Set Headers
	for k, v := range channel.Headers {
		req.Header.Set(k, v)
	}
	// Default Content-Type if not set
	if req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	// 5. Send Request
	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("webhook failed with status: %d", resp.StatusCode)
	}

	return nil
}
