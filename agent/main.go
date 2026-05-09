package main

import (
	"log"
	"time"

	"agent/config"
	"agent/reporter"
)

func main() {
	cfg := config.LoadConfig()
	log.Printf("Starting Mashiro Agent (ID: %s)", cfg.AgentID)
	log.Printf("Reporting to: %s every %v", cfg.ServerURL, cfg.ReportInterval)

	rep := reporter.NewReporter(cfg.ServerURL, cfg.AgentID)
	latencyTaskRuns := make(map[uint]time.Time)

	ticker := time.NewTicker(cfg.ReportInterval)
	defer ticker.Stop()

	for range ticker.C {
		err := rep.Report()
		if err != nil {
			log.Printf("Error reporting data: %v", err)
		} else {
			log.Printf("Successfully reported data to server")
		}

		if err := rep.RunLatencyTasks(time.Now(), latencyTaskRuns); err != nil {
			log.Printf("Error running latency tasks: %v", err)
		}
	}
}
