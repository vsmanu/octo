package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"gopkg.in/yaml.v3"
)

// Config represents the top-level configuration structure
type Config struct {
	Global        GlobalConfig      `yaml:"global" json:"global"`
	Endpoints     []EndpointConfig  `yaml:"endpoints" json:"endpoints"`
	AlertChannels []AlertChannel    `yaml:"alert_channels" json:"alert_channels"`
	AlertRules    []AlertRule       `yaml:"alert_rules" json:"alert_rules"`
	Satellites    []SatelliteConfig `yaml:"satellites" json:"satellites"`
}

type GlobalConfig struct {
	CheckInterval  time.Duration `yaml:"check_interval" json:"check_interval"`
	RequestTimeout time.Duration `yaml:"request_timeout" json:"request_timeout"`
}

type EndpointConfig struct {
	ID         string            `yaml:"id" json:"id"`
	Name       string            `yaml:"name" json:"name"`
	URL        string            `yaml:"url" json:"url"`
	Method     string            `yaml:"method" json:"method"`
	Interval   time.Duration     `yaml:"interval" json:"interval"`
	Timeout    time.Duration     `yaml:"timeout" json:"timeout"`
	Headers    map[string]string `yaml:"headers" json:"headers"`
	Validation ValidationConfig  `yaml:"validation" json:"validation"`
	SSL        SSLConfig         `yaml:"ssl" json:"ssl"`
	Tags       map[string]string `yaml:"tags" json:"tags"`
}

type ValidationConfig struct {
	StatusCodes  []int        `yaml:"status_codes" json:"status_codes"`
	ContentMatch ContentMatch `yaml:"content_match" json:"content_match"`
}

type ContentMatch struct {
	Type    string `yaml:"type" json:"type"` // "regex" or "exact"
	Pattern string `yaml:"pattern" json:"pattern"`
}

type SatelliteConfig struct {
	ID   string `yaml:"id" json:"id"`
	Name string `yaml:"name" json:"name"`
	URL  string `yaml:"url,omitempty" json:"url,omitempty"`
}

type SSLConfig struct {
	ExpirationAlertDays []int `yaml:"expiration_alert_days" json:"expiration_alert_days"`
}

type AlertChannel struct {
	Name    string            `yaml:"name" json:"name"`
	Type    string            `yaml:"type" json:"type"` // "webhook"
	URL     string            `yaml:"url" json:"url"`
	Headers map[string]string `yaml:"headers" json:"headers"`
	Body    string            `yaml:"body" json:"body"` // Template string
}

type AlertRule struct {
	Name      string            `yaml:"name" json:"name"`
	Condition string            `yaml:"condition" json:"condition"`
	Severity  string            `yaml:"severity" json:"severity"`
	Channels  []string          `yaml:"channels" json:"channels"`
	Tags      map[string]string `yaml:"tags" json:"tags"`
}

// Manager handles concurrent access to configuration and file watching
type Manager struct {
	mu         sync.RWMutex
	config     *Config
	configPath string
	watcher    *fsnotify.Watcher
	onChange   func(*Config)
}

// NewManager creates a new configuration manager
func NewManager(path string) (*Manager, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	mgr := &Manager{
		configPath: absPath,
	}

	// Initial load
	if err := mgr.Load(); err != nil {
		return nil, err
	}

	return mgr, nil
}

// Load reads the configuration from the file
func (m *Manager) Load() error {
	file, err := os.Open(m.configPath)
	if err != nil {
		return fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	var cfg Config
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&cfg); err != nil {
		return fmt.Errorf("failed to decode config file: %w", err)
	}

	// Set defaults if needed
	if cfg.Global.CheckInterval == 0 {
		cfg.Global.CheckInterval = 60 * time.Second
	}
	if cfg.Global.RequestTimeout == 0 {
		cfg.Global.RequestTimeout = 10 * time.Second
	}

	m.mu.Lock()
	m.config = &cfg
	m.mu.Unlock()

	return nil
}

// Save writes the current configuration to the file
func (m *Manager) Save() error {
	m.mu.RLock()
	cfg := m.config
	m.mu.RUnlock()

	// Atomic write: write to temp file then rename
	dir := filepath.Dir(m.configPath)
	tmpFile, err := os.CreateTemp(dir, "config-*.yml")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name()) // Clean up on error

	encoder := yaml.NewEncoder(tmpFile)
	encoder.SetIndent(2)
	if err := encoder.Encode(cfg); err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to encode config: %w", err)
	}
	tmpFile.Close()

	// Rename temp file to actual config file
	if err := os.Rename(tmpFile.Name(), m.configPath); err != nil {
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	return nil
}

// GetConfig returns a copy of the current configuration
func (m *Manager) GetConfig() Config {
	m.mu.RLock()
	defer m.mu.RUnlock()
	// Return a copy to avoid race conditions if caller modifies it (shallow copy is usually fine for read-only access)
	// But to be safe, we rely on the caller not mutating the returned pointer if we returned *Config.
	// Here we return struct which does shallow copy. Deep copy would be safer but more expensive.
	return *m.config
}

// UpdateConfig updates the configuration and saves it to disk
func (m *Manager) UpdateConfig(updater func(*Config) error) error {
	m.mu.Lock()
	// Create a deep copy or just modify the existing struct (holding the lock)
	// Since we are holding the lock, we can modify directly.
	// However, if the updater fails, we might leave config in inconsistent state.
	// Better to clone, update, then swap.
	// For simplicity in MVP, we modify in place but rollback on save error?
	// Actually, simpler: just let updater modify.
	err := updater(m.config)
	m.mu.Unlock()

	if err != nil {
		// Config might be partially modified, which is bad. Ideally we re-load from disk to reset.
		// Reloading from disk to restore state
		_ = m.Load()
		return err
	}

	return m.Save()
}

// Watch starts watching the config file for changes
func (m *Manager) Watch(onChange func(*Config)) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	m.watcher = watcher
	m.onChange = onChange

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Has(fsnotify.Write) || event.Has(fsnotify.Rename) || event.Has(fsnotify.Chmod) {
					// Add a small delay to avoid reading partial writes
					time.Sleep(100 * time.Millisecond)

					// If renamed (e.g. atomic write), we might need to re-add watch?
					// fsnotify behavior varies by platform.
					// Usually Rename means the file we were watching is gone/renamed.
					// But if it was replaced by atomic write (rename temp to target), the target inode changes.
					// We need to re-add the watch on the file path.

					if err := m.Load(); err == nil {
						if m.onChange != nil {
							// Run callback in goroutine to not block watcher
							go func(cfg Config) {
								m.onChange(&cfg)
							}(m.GetConfig())
						}
					} else {
						// Log error?
						fmt.Fprintf(os.Stderr, "Error reloading config: %v\n", err)
					}

					// Re-add watch if needed (especially for Rename/Remove events on atomic writes)
					// The file path persists, so we just add it again to be safe.
					watcher.Add(m.configPath)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				fmt.Fprintf(os.Stderr, "Config watcher error: %v\n", err)
			}
		}
	}()

	return watcher.Add(m.configPath)
}

// Close stops the watcher and cleans up
func (m *Manager) Close() error {
	if m.watcher != nil {
		return m.watcher.Close()
	}
	return nil
}
