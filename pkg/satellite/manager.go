package satellite

import (
	"sync"
	"time"

	"github.com/manu/octo/pkg/config"
)

// Status represents the current state of a satellite
type Status string

const (
	StatusOnline  Status = "online"
	StatusOffline Status = "offline"
)

// SatelliteState holds the runtime state of a satellite
type SatelliteState struct {
	ID            string                 `json:"id"`
	LastHeartbeat time.Time              `json:"last_heartbeat"`
	Status        Status                 `json:"status"`
	Config        config.SatelliteConfig `json:"config"`
}

// Manager handles the lifecycle and state of satellites
type Manager struct {
	mu         sync.RWMutex
	satellites map[string]*SatelliteState
	cfgMgr     *config.Manager
}

// NewManager creates a new satellite manager
func NewManager(cfgMgr *config.Manager) *Manager {
	m := &Manager{
		satellites: make(map[string]*SatelliteState),
		cfgMgr:     cfgMgr,
	}
	m.refreshFromConfig()
	return m
}

// refreshFromConfig syncs the in-memory state with the config
func (m *Manager) refreshFromConfig() {
	cfg := m.cfgMgr.GetConfig()
	m.mu.Lock()
	defer m.mu.Unlock()

	// Add new satellites from config
	for _, satCfg := range cfg.Satellites {
		if _, exists := m.satellites[satCfg.ID]; !exists {
			m.satellites[satCfg.ID] = &SatelliteState{
				ID:     satCfg.ID,
				Status: StatusOffline,
				Config: satCfg,
			}
		} else {
			// Update config if changed
			m.satellites[satCfg.ID].Config = satCfg
		}
	}
}

// GetSatellite returns the state of a specific satellite
func (m *Manager) GetSatellite(id string) (*SatelliteState, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	sat, ok := m.satellites[id]
	return sat, ok
}

// GetAllSatellites returns all satellites
func (m *Manager) GetAllSatellites() []*SatelliteState {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var sats []*SatelliteState
	for _, s := range m.satellites {
		sats = append(sats, s)
	}
	return sats
}

// RegisterHeartbeat updates the last heartbeat time for a satellite
func (m *Manager) RegisterHeartbeat(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if sat, exists := m.satellites[id]; exists {
		sat.LastHeartbeat = time.Now()
		sat.Status = StatusOnline
	}
}

// MarkOffline marks a satellite as offline (e.g. missed heartbeats)
// This could be called by a background ticker
func (m *Manager) CheckOffline(timeout time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	for _, sat := range m.satellites {
		if now.Sub(sat.LastHeartbeat) > timeout {
			sat.Status = StatusOffline
		}
	}
}
