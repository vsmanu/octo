package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/manu/octo/pkg/api"
	"github.com/manu/octo/pkg/config"
	"github.com/manu/octo/pkg/satellite"
	"github.com/manu/octo/pkg/scheduler"
	"github.com/manu/octo/pkg/storage/postgres"
	"github.com/manu/octo/web"
)

func main() {
	// 1. Load Configuration
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config/config.yml"
	}

	cfgMgr, err := config.NewManager(configPath)
	if err != nil {
		log.Printf("Failed to load config from %s: %v. Creating default if not exists...", configPath, err)
		if os.IsNotExist(err) {
			log.Fatalf("Config file not found at %s. Please copy config.example.yml to %s", configPath, configPath)
		}
		log.Fatalf("Failed to load config: %v", err)
	}

	// 2. Initialize Storage (PostgreSQL/TimescaleDB)
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}
	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		dbPort = "5432"
	}
	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		dbUser = "postgres"
	}
	dbPassword := os.Getenv("DB_PASSWORD")
	if dbPassword == "" {
		dbPassword = "password123"
	}
	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "octo"
	}

	ctx := context.Background()
	var store *postgres.PostgresStorage
	var dbErr error

	// Retry loop for database connection
	maxRetries := 30
	for i := 0; i < maxRetries; i++ {
		store, dbErr = postgres.NewPostgresStorage(ctx, dbHost, dbPort, dbUser, dbPassword, dbName)
		if dbErr == nil {
			break
		}
		log.Printf("Failed to initialize database (attempt %d/%d): %v. Retrying in 2s...", i+1, maxRetries, dbErr)
		time.Sleep(2 * time.Second)
	}

	if dbErr != nil {
		log.Fatalf("Failed to initialize database after %d attempts: %v", maxRetries, dbErr)
	}
	defer store.Close()

	// 3. Initialize Scheduler
	sched := scheduler.NewScheduler(cfgMgr, store)

	// 4. Start Scheduler
	go sched.Start()
	defer sched.Stop()

	// 5. Initialize Satellite Manager (after Config Manager)
	satMgr := satellite.NewManager(cfgMgr)

	// 6. Start API Server
	distFS, err := web.GetDistFS()
	if err != nil {
		log.Fatalf("Failed to get embedded frontend: %v", err)
	}

	apiServer := api.NewServer(cfgMgr, store, satMgr, distFS)
	srv := &http.Server{
		Addr:    ":8080",
		Handler: apiServer.Handler(),
	}

	go func() {
		log.Println("Starting HTTP API on :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("API server failed: %v", err)
		}
	}()

	// 6. Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
}
