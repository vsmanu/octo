package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/manu/octo/pkg/agent"
)

func main() {
	// 1. Load Configuration from Env
	masterURL := os.Getenv("MASTER_URL")
	if masterURL == "" {
		log.Fatal("MASTER_URL is required")
	}

	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		log.Fatal("API_KEY is required")
	}

	satelliteID := os.Getenv("SATELLITE_ID")
	if satelliteID == "" {
		log.Fatal("SATELLITE_ID is required")
	}

	// 2. Initialize Agent
	agt := agent.NewAgent(satelliteID, masterURL, apiKey)

	// 3. Start Agent
	log.Printf("Starting Satellite Agent %s connecting to %s", satelliteID, masterURL)
	go agt.Start()

	// 4. Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down satellite...")
	agt.Stop()
	log.Println("Satellite exited")
}
