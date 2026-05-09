package controllers

import (
	"net/http"
	"time"

	"backend/db"
	"backend/models"

	"github.com/gin-gonic/gin"
)

type ServerDashboardResponse struct {
	ID          uint    `json:"id"`
	Name        string  `json:"name"`
	Status      string  `json:"status"`
	Region      string  `json:"region"`
	System      string  `json:"system"`
	HasIPv4     bool    `json:"has_ipv4"`
	HasIPv6     bool    `json:"has_ipv6"`
	TrafficLimit uint64 `json:"traffic_limit"`
	ExpiryTime  string  `json:"expiry_time"`
	Uptime      uint64  `json:"uptime"`
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
	LatencyResults []LatencyResultResponse `json:"latency_results"`
}

func GetDashboardStats(c *gin.Context) {
	var servers []models.Server
	db.DB.Find(&servers)
	now := time.Now().UTC()

	onlineCount := 0
	offlineCount := 0
	var totalTrafficUsed uint64
	var totalSpeedIn uint64
	var totalSpeedOut uint64

	for i := range servers {
		s := &servers[i]
		latestData, hasLatest := latestAgentDataForServer(s.ID)
		syncServerStatus(s, latestData, hasLatest, now)

		if s.Status == "online" {
			onlineCount++
		} else {
			offlineCount++
		}

		if hasLatest {
			netIn, netOut := currentTrafficUsage(*s, latestData)
			totalTrafficUsed += netIn + netOut
			if s.Status == "online" {
				totalSpeedIn += latestData.NetInSpeed
				totalSpeedOut += latestData.NetOutSpeed
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"total_servers":      len(servers),
		"online":             onlineCount,
		"offline":            offlineCount,
		"total_traffic_used": totalTrafficUsed,
		"total_speed_in":     totalSpeedIn,
		"total_speed_out":    totalSpeedOut,
	})
}

func GetServersList(c *gin.Context) {
	var servers []models.Server
	db.DB.Find(&servers)
	now := time.Now().UTC()

	var res []ServerDashboardResponse

	for i := range servers {
		s := &servers[i]
		latestData, hasLatest := latestAgentDataForServer(s.ID)
		syncServerStatus(s, latestData, hasLatest, now)
		netIn, netOut := currentTrafficUsage(*s, latestData)
		expiryTime := ""
		if !s.ExpiryTime.IsZero() {
			expiryTime = s.ExpiryTime.Format(time.RFC3339)
		}
		latencyResults, err := loadServerLatencyResults(s.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load latency results"})
			return
		}
		latestLatency := 0.0
		if len(latencyResults) > 0 {
			latestLatency = latencyResults[0].LatencyMS
		}

		res = append(res, ServerDashboardResponse{
			ID:          s.ID,
			Name:        s.Name,
			Status:      s.Status,
			Region:      s.Region,
			System:      s.System,
			HasIPv4:     s.HasIPv4,
			HasIPv6:     s.HasIPv6,
			TrafficLimit: s.TrafficLimit,
			ExpiryTime:  expiryTime,
			Uptime:      latestData.Uptime,
			CPUCores:    latestData.CPUCores,
			CPUUsage:    latestData.CPUUsage,
			MemTotal:    latestData.MemTotal,
			MemUsed:     latestData.MemUsed,
			DiskTotal:   latestData.DiskTotal,
			DiskUsed:    latestData.DiskUsed,
			NetIn:       netIn,
			NetOut:      netOut,
			NetInSpeed:  latestData.NetInSpeed,
			NetOutSpeed: latestData.NetOutSpeed,
			Latency:     latestLatency,
			LatencyResults: latencyResults,
		})
	}

	c.JSON(http.StatusOK, res)
}

func loadServerLatencyResults(serverID uint) ([]LatencyResultResponse, error) {
	var items []models.LatencyResult
	if err := db.DB.Where("server_id = ?", serverID).Order("checked_at desc").Find(&items).Error; err != nil {
		return nil, err
	}

	results := make([]LatencyResultResponse, 0, len(items))
	for _, item := range items {
		checkedAt := ""
		if !item.CheckedAt.IsZero() {
			checkedAt = item.CheckedAt.Format(time.RFC3339)
		}
		results = append(results, LatencyResultResponse{
			TaskID:       item.TaskID,
			ServerID:     item.ServerID,
			TaskName:     item.TaskName,
			Type:         item.Type,
			Target:       item.Target,
			LatencyMS:    item.LatencyMS,
			Status:       item.Status,
			ErrorMessage: item.ErrorMessage,
			CheckedAt:    checkedAt,
		})
	}

	return results, nil
}
