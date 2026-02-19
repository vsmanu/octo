package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/manu/octo/pkg/checker"
	"github.com/manu/octo/pkg/config"
)

type Agent struct {
	ID        string
	MasterURL string
	APIKey    string
	Client    *http.Client
	Checker   *checker.Checker
	StopChan  chan struct{}
}

func NewAgent(id, masterURL, apiKey string) *Agent {
	return &Agent{
		ID:        id,
		MasterURL: masterURL,
		APIKey:    apiKey,
		Client:    &http.Client{Timeout: 10 * time.Second},
		Checker:   checker.NewChecker(),
		StopChan:  make(chan struct{}),
	}
}

func (a *Agent) Start() {
	// 1. Heartbeat Loop
	go a.heartbeatLoop()

	// 2. Poll Config & Run Checks Loop
	// Logic: Poll config, if new checks, update.
	// For simplicity in MVP: Poll config every 30s, run checks immediately.
	// A better approach is separate scheduler, but let's keep it simple.
	go a.executionLoop()
}

func (a *Agent) Stop() {
	close(a.StopChan)
}

func (a *Agent) heartbeatLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			a.sendHeartbeat()
		case <-a.StopChan:
			return
		}
	}
}

func (a *Agent) sendHeartbeat() {
	req, err := http.NewRequest("POST", a.MasterURL+"/api/v1/satellites/heartbeat", nil)
	if err != nil {
		log.Printf("Error creating heartbeat request: %v", err)
		return
	}
	req.Header.Set("X-Satellite-ID", a.ID)
	// req.Header.Set("Authorization", "Bearer " + a.APIKey) // TODO

	resp, err := a.Client.Do(req)
	if err != nil {
		log.Printf("Error sending heartbeat: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Heartbeat failed with status: %d", resp.StatusCode)
	}
}

func (a *Agent) executionLoop() {
	// Initial delay
	time.Sleep(2 * time.Second)

	ticker := time.NewTicker(60 * time.Second) // Check interval default
	defer ticker.Stop()

	// Run once immediately
	a.runChecks()

	for {
		select {
		case <-ticker.C:
			a.runChecks()
		case <-a.StopChan:
			return
		}
	}
}

func (a *Agent) runChecks() {
	// 1. Get Config
	endpoints, err := a.fetchConfig()
	if err != nil {
		log.Printf("Failed to fetch config: %v", err)
		return
	}

	if len(endpoints) == 0 {
		log.Println("Fetched 0 endpoints from master. Nothing to do.")
		return
	}

	log.Printf("Running %d checks...", len(endpoints))

	// 2. Execute Checks (Sequential for MVP, could be parallel)
	var results []checker.Result
	ctx := context.Background()

	for _, ep := range endpoints {
		res := a.Checker.Check(ctx, ep)
		res.SatelliteID = a.ID // Tag result with our ID
		results = append(results, res)
	}

	// 3. Push Results
	if len(results) > 0 {
		if err := a.pushResults(results); err != nil {
			log.Printf("Failed to push results: %v", err)
		} else {
			log.Printf("Pushed %d results", len(results))
		}
	}
}

func (a *Agent) fetchConfig() ([]config.EndpointConfig, error) {
	req, err := http.NewRequest("GET", a.MasterURL+"/api/v1/satellites/config", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Satellite-ID", a.ID)

	resp, err := a.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, log.Output(2, "Config fetch failed: "+resp.Status) // simplified error
	}

	var endpoints []config.EndpointConfig
	if err := json.NewDecoder(resp.Body).Decode(&endpoints); err != nil {
		return nil, err
	}
	return endpoints, nil
}

func (a *Agent) pushResults(results []checker.Result) error {
	data, err := json.Marshal(results)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", a.MasterURL+"/api/v1/satellites/results", bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	req.Header.Set("X-Satellite-ID", a.ID)
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return log.Output(2, "Push results failed: "+resp.Status)
	}
	return nil
}
