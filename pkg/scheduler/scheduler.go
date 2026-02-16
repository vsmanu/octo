package scheduler

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/manu/octo/pkg/alerting"
	"github.com/manu/octo/pkg/checker"
	"github.com/manu/octo/pkg/config"
	"github.com/manu/octo/pkg/storage"
)

type Scheduler struct {
	cfgManager   *config.Manager
	checker      *checker.Checker
	storage      storage.Provider
	alertManager *alerting.Manager
	ctx          context.Context
	cancel       context.CancelFunc
	wg           sync.WaitGroup
}

func NewScheduler(cfgMgr *config.Manager, store storage.Provider) *Scheduler {
	ctx, cancel := context.WithCancel(context.Background())
	am := alerting.NewManager(cfgMgr)
	am.RegisterProvider("webhook", alerting.NewWebhookProvider())

	return &Scheduler{
		cfgManager:   cfgMgr,
		checker:      checker.NewChecker(),
		storage:      store,
		alertManager: am,
		ctx:          ctx,
		cancel:       cancel,
	}
}

func (s *Scheduler) Start() {
	// Start initial set of workers
	s.restartWorkers()

	// Watch for config changes
	s.cfgManager.Watch(func(cfg *config.Config) {
		log.Println("Config changed, restarting scheduler...")
		s.Stop()
		s.ctx, s.cancel = context.WithCancel(context.Background())
		s.restartWorkers()
	})
}

func (s *Scheduler) Stop() {
	s.cancel()
	s.wg.Wait()
}

func (s *Scheduler) restartWorkers() {
	cfg := s.cfgManager.GetConfig()

	for _, endpoint := range cfg.Endpoints {
		s.wg.Add(1)
		go s.runWorker(endpoint)
	}
}

func (s *Scheduler) runWorker(endpoint config.EndpointConfig) {
	defer s.wg.Done()

	interval := endpoint.Interval
	if interval == 0 {
		interval = 60 * time.Second // Default
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Run immediately
	s.executeCheck(endpoint)

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.executeCheck(endpoint)
		}
	}
}

func (s *Scheduler) executeCheck(endpoint config.EndpointConfig) {
	ctx, cancel := context.WithTimeout(s.ctx, endpoint.Timeout)
	if endpoint.Timeout == 0 {
		ctx, cancel = context.WithTimeout(s.ctx, 10*time.Second)
	}
	defer cancel()

	result := s.checker.Check(ctx, endpoint)

	// Log result
	if result.Success {
		log.Printf("Check passed: %s (%s) - %v", endpoint.Name, endpoint.URL, result.Duration)
	} else {
		log.Printf("Check failed: %s (%s) - %s", endpoint.Name, endpoint.URL, result.Error)
	}

	if err := s.storage.WriteResult(result); err != nil {
		log.Printf("Failed to write result to InfluxDB: %v", err)
	}

	// Evaluate Alerts
	s.alertManager.Evaluate(ctx, endpoint, &result)
}
