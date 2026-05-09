package config

import (
	"os"
	"time"
)

type Config struct {
	ServerURL     string
	AgentID       string
	ReportInterval time.Duration
}

func LoadConfig() Config {
	serverURL := os.Getenv("MASHIRO_SERVER_URL")
	if serverURL == "" {
		serverURL = "http://localhost:8080/api/agent/report"
	}

	agentID := os.Getenv("MASHIRO_AGENT_ID")
	if agentID == "" {
		agentID = "default-agent-1"
	}

	return Config{
		ServerURL:     serverURL,
		AgentID:       agentID,
		ReportInterval: 5 * time.Second,
	}
}
