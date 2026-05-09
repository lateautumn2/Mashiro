package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username string `gorm:"uniqueIndex;not null" json:"username"`
	Password string `gorm:"not null" json:"-"`
}

type Server struct {
	gorm.Model
	Name                  string    `json:"name"`
	IP                    string    `json:"ip"`
	Region                string    `json:"region"`
	System                string    `json:"system"`
	HasIPv4               bool      `json:"has_ipv4"`
	HasIPv6               bool      `json:"has_ipv6"`
	Status                string    `json:"status"` // "online", "offline"
	AuthToken             string    `gorm:"uniqueIndex" json:"auth_token"`
	TrafficLimit          uint64    `json:"traffic_limit"` // in bytes
	TrafficResetDay       int       `json:"traffic_reset_day"`
	TrafficResetAt        time.Time `json:"traffic_reset_at"`
	TrafficResetNetInBase uint64    `json:"traffic_reset_net_in_base"`
	TrafficResetNetOutBase uint64   `json:"traffic_reset_net_out_base"`
	ExpiryTime            time.Time `json:"expiry_time"`
}

type AgentData struct {
	gorm.Model
	ServerID    uint      `json:"server_id"`
	CPUCores    int       `json:"cpu_cores"`
	CPUUsage    float64   `json:"cpu_usage"`
	MemTotal    uint64    `json:"mem_total"`
	MemUsed     uint64    `json:"mem_used"`
	DiskTotal   uint64    `json:"disk_total"`
	DiskUsed    uint64    `json:"disk_used"`
	NetIn       uint64    `json:"net_in"`
	NetOut      uint64    `json:"net_out"`
	NetInSpeed  uint64    `json:"net_in_speed"`
	NetOutSpeed uint64    `json:"net_out_speed"`
	Latency     float64   `json:"latency"`
	PacketLoss  float64   `json:"packet_loss"`
	Uptime      uint64    `json:"uptime"`
	ReportTime time.Time `json:"report_time"`
}

type LatencyTask struct {
	gorm.Model
	Name        string              `json:"name"`
	Type        string              `json:"type"`
	Target      string              `json:"target"`
	IntervalSec int                 `json:"interval_sec"`
	Enabled     bool                `json:"enabled"`
	Servers     []LatencyTaskServer `gorm:"foreignKey:TaskID;references:ID" json:"servers"`
}

type LatencyTaskServer struct {
	gorm.Model
	TaskID   uint `gorm:"index:idx_latency_task_server,unique" json:"task_id"`
	ServerID uint `gorm:"index:idx_latency_task_server,unique" json:"server_id"`
}

type LatencyResult struct {
	gorm.Model
	TaskID        uint      `gorm:"index:idx_latency_result_task_server,unique" json:"task_id"`
	ServerID      uint      `gorm:"index:idx_latency_result_task_server,unique" json:"server_id"`
	TaskName      string    `json:"task_name"`
	Type          string    `json:"type"`
	Target        string    `json:"target"`
	LatencyMS     float64   `json:"latency_ms"`
	Status        string    `json:"status"`
	ErrorMessage  string    `json:"error_message"`
	CheckedAt     time.Time `json:"checked_at"`
}

type AppConfig struct {
	gorm.Model
	Key   string `gorm:"uniqueIndex" json:"key"`
	Value string `json:"value"`
}
