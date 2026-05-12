package controllers

import (
	"net/http"
	"time"

	"backend/db"
	"backend/models"
	"backend/notifier"

	"github.com/gin-gonic/gin"
)

type AgentReportInput struct {
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

func ReportAgentData(c *gin.Context) {
	serverID := c.GetUint("agent_server_id")
	if serverID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid server identity"})
		return
	}

	var input AgentReportInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var server models.Server
	if err := db.DB.First(&server, serverID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Server not found"})
		return
	}

	// Detect online event for notification (before overwriting status).
	wasOffline := server.Status == "offline"

	// Update server metadata from latest agent report.
	server.Status = "online"
	server.System = input.System
	server.HasIPv4 = input.HasIPv4
	server.HasIPv6 = input.HasIPv6
	nextIP := ""
	if input.PublicIPv4 != "" {
		nextIP = input.PublicIPv4
	} else if input.PublicIPv6 != "" {
		nextIP = input.PublicIPv6
	}
	ipChanged := nextIP != "" && nextIP != server.IP
	if ipChanged {
		server.IP = nextIP
		if region := lookupRegionByIP(server.IP); region != "" {
			server.Region = region
		}
	}
	syncTrafficResetState(&server, input.NetIn, input.NetOut, time.Now())
	db.DB.Save(&server)

	if wasOffline {
		notifier.NotifyOnline(server)
	}

	// Save agent data
	agentData := models.AgentData{
		ServerID:    server.ID,
		CPUCores:    input.CPUCores,
		CPUUsage:    input.CPUUsage,
		MemTotal:    input.MemTotal,
		MemUsed:     input.MemUsed,
		DiskTotal:   input.DiskTotal,
		DiskUsed:    input.DiskUsed,
		NetIn:       input.NetIn,
		NetOut:      input.NetOut,
		NetInSpeed:  input.NetInSpeed,
		NetOutSpeed: input.NetOutSpeed,
		Latency:     input.Latency,
		PacketLoss:  input.PacketLoss,
		Uptime:      input.Uptime,
		ReportTime:  time.Now(),
	}

	if err := db.DB.Create(&agentData).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save agent data"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "success"})
}
