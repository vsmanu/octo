package alerting

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/manu/octo/pkg/checker"
	"github.com/manu/octo/pkg/config"
)

// AlertManager handles the evaluation of alert rules and triggering of notifications
type Manager struct {
	cfgManager *config.Manager
	providers  map[string]Provider
	// State to track firing alerts (to avoid spamming)
	// Key: endpointID + ruleName
	activeAlerts map[string]bool
	mu           sync.RWMutex
}

// NewManager creates a new AlertManager
func NewManager(cfgMgr *config.Manager) *Manager {
	return &Manager{
		cfgManager:   cfgMgr,
		providers:    make(map[string]Provider), // In future we can support multiple types map[type]Provider
		activeAlerts: make(map[string]bool),
	}
}

// RegisterProvider registers a provider implementation
// For now we only have "webhook", but this allows extension
func (m *Manager) RegisterProvider(providerType string, p Provider) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.providers[providerType] = p
}

// Evaluate checks if the result triggers any alerts
func (m *Manager) Evaluate(ctx context.Context, endpoint config.EndpointConfig, result *checker.Result) {
	cfg := m.cfgManager.GetConfig()

	for _, rule := range cfg.AlertRules {
		// 1. Check Tags
		if !m.matchTags(endpoint.Tags, rule.Tags) {
			continue
		}

		// 2. Check Condition
		triggered := m.checkCondition(rule.Condition, result)
		alertKey := fmt.Sprintf("%s-%s", endpoint.ID, rule.Name)

		m.mu.Lock()
		wasActive := m.activeAlerts[alertKey]
		m.mu.Unlock()

		if triggered {
			if !wasActive {
				// Fire Alert (New)
				log.Printf("Alert Triggered: %s for %s", rule.Name, endpoint.Name)
				m.triggerChannels(ctx, rule, endpoint, result, cfg.AlertChannels)

				m.mu.Lock()
				m.activeAlerts[alertKey] = true
				m.mu.Unlock()
			}
			// Else: Already active, maybe implementing repeat intervals later
		} else {
			if wasActive {
				// Resolve Alert
				log.Printf("Alert Resolved: %s for %s", rule.Name, endpoint.Name)
				// Optional: Send resolution notification

				m.mu.Lock()
				delete(m.activeAlerts, alertKey)
				m.mu.Unlock()
			}
		}
	}
}

// matchTags checks if the endpoint has all the tags defined in the rule
func (m *Manager) matchTags(endpointTags, ruleTags map[string]string) bool {
	if len(ruleTags) == 0 {
		return true // No tags in rule matches everything (or maybe nothing? implied "global")
		// Let's assume matches everything for now, or use specific "global" tag
	}
	for k, v := range ruleTags {
		if val, ok := endpointTags[k]; !ok || val != v {
			return false
		}
	}
	return true
}

// checkCondition evaluates the condition string against the result
// Supported: "success == false", "duration > 5s" (simple parsing)
func (m *Manager) checkCondition(condition string, result *checker.Result) bool {
	// Very basic parser for MVP
	// In a real system, use an expression engine
	cond := strings.TrimSpace(condition)

	if cond == "success == false" {
		return !result.Success
	}

	// TODO: Add more conditions (duration, etc.)
	return false
}

func (m *Manager) triggerChannels(ctx context.Context, rule config.AlertRule, endpoint config.EndpointConfig, result *checker.Result, channels []config.AlertChannel) {
	// Map channel names to config
	channelMap := make(map[string]config.AlertChannel)
	for _, ch := range channels {
		channelMap[ch.Name] = ch
	}

	for _, chName := range rule.Channels {
		chConfig, ok := channelMap[chName]
		if !ok {
			log.Printf("Warning: Alert channel '%s' not found", chName)
			continue
		}

		// Find provider logic
		// For now we assume "webhook" for everything or switch based on type
		// If we had multiple types, we'd lookup m.providers[chConfig.Type]

		// Since we only have webhook provider implemented and registered (eventually):
		provider := m.providers[chConfig.Type]
		if provider == nil {
			log.Printf("Warning: No provider registered for type '%s'", chConfig.Type)
			continue
		}

		go func(p Provider, c config.AlertChannel) {
			childCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			if err := p.Send(childCtx, c, rule, endpoint, result); err != nil {
				log.Printf("Failed to send alert to %s: %v", c.Name, err)
			}
		}(provider, chConfig)
	}
}
