package alerting

import (
	"context"

	"github.com/manu/octo/pkg/checker"
	"github.com/manu/octo/pkg/config"
)

// Provider defines the interface for an alert provider (e.g., Webhook, Email, Slack)
type Provider interface {
	// Send sends an alert using the provided configuration and result data
	Send(ctx context.Context, channel config.AlertChannel, rule config.AlertRule, endpoint config.EndpointConfig, result *checker.Result) error
}
