package reporter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"agent/collector"
	"agent/signer"
)

type ReportPayload struct {
	CPUCores    int     `json:"cpu_cores"`
	CPUUsage    float64 `json:"cpu_usage"`
	MemTotal    uint64  `json:"mem_total"`
	MemUsed     uint64  `json:"mem_used"`
	DiskTotal   uint64  `json:"disk_total"`
	DiskUsed    uint64  `json:"disk_used"`
	NetIn       uint64  `json:"net_in"`
	NetOut      uint64  `json:"net_out"`
	NetInSpeed  uint64  `json:"net_in_speed"`
	NetOutSpeed uint64  `json:"net_out_speed"`
	Latency     float64 `json:"latency"`
	PacketLoss  float64 `json:"packet_loss"`
	Uptime      uint64  `json:"uptime"`
	System      string  `json:"system"`
	PublicIPv4  string  `json:"public_ipv4"`
	PublicIPv6  string  `json:"public_ipv6"`
	HasIPv4     bool    `json:"has_ipv4"`
	HasIPv6     bool    `json:"has_ipv6"`
}

type Reporter struct {
	ServerURL string
	AgentID   string
	Client    *http.Client
}

func NewReporter(serverURL, agentID string) *Reporter {
	return &Reporter{
		ServerURL: serverURL,
		AgentID:   agentID,
		Client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (r *Reporter) Report() error {
	systemInfo := collector.CollectSystemInfo()
	networkInfo := collector.CollectNetworkInfo()
	agentMeta := collector.CollectAgentMeta()

	payload := ReportPayload{
		CPUCores:    systemInfo.CPUCores,
		CPUUsage:    systemInfo.CPUUsage,
		MemTotal:    systemInfo.MemTotal,
		MemUsed:     systemInfo.MemUsed,
		DiskTotal:   systemInfo.DiskTotal,
		DiskUsed:    systemInfo.DiskUsed,
		NetIn:       networkInfo.BytesRecv,
		NetOut:      networkInfo.BytesSent,
		NetInSpeed:  uint64(networkInfo.SpeedRecv),
		NetOutSpeed: uint64(networkInfo.SpeedSend),
		Latency:     networkInfo.Latency,
		PacketLoss:  networkInfo.PacketLoss,
		Uptime:      systemInfo.Uptime,
		System:      agentMeta.SystemName,
		PublicIPv4:  agentMeta.PublicIPv4,
		PublicIPv6:  agentMeta.PublicIPv6,
		HasIPv4:     agentMeta.HasIPv4,
		HasIPv6:     agentMeta.HasIPv6,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest("POST", r.ServerURL, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Agent-Token", r.AgentID)

	// Sign the request with HMAC to prevent tampering and replay.
	sig, ts := signer.SignRequest("POST", req.URL.Path, data, r.AgentID)
	req.Header.Set("X-Signature", sig)
	req.Header.Set("X-Timestamp", ts)

	resp, err := r.Client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send report: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("server returned non-success status: %d", resp.StatusCode)
	}

	return nil
}
