package collector

import (
	"log"
	"os"
	"runtime"
	"time"

	gopsutilcpu "github.com/shirou/gopsutil/v3/cpu"
	gopsutildisk "github.com/shirou/gopsutil/v3/disk"
	gopsutilhost "github.com/shirou/gopsutil/v3/host"
	gopsutilmem "github.com/shirou/gopsutil/v3/mem"
)

type SystemInfo struct {
	CPUUsage    float64 `json:"cpu_usage"`    // percentage 0-100
	CPUCores    int     `json:"cpu_cores"`    // logical cores
	MemTotal    uint64  `json:"mem_total"`    // bytes
	MemUsed     uint64  `json:"mem_used"`     // bytes
	DiskTotal   uint64  `json:"disk_total"`   // bytes
	DiskUsed    uint64  `json:"disk_used"`    // bytes
	Uptime      uint64  `json:"uptime"`       // seconds
}

const defaultWindowsDiskPath = `C:\`

type memoryStats struct {
	total uint64
	used  uint64
}

type diskStats struct {
	total uint64
	used  uint64
}

// CollectSystemInfo returns real system information.
func CollectSystemInfo() SystemInfo {
	memInfo := collectMemoryInfo()
	diskInfo := collectDiskInfo()

	return SystemInfo{
		CPUUsage:  collectCPUUsage(),
		CPUCores:  runtime.NumCPU(),
		MemTotal:  memInfo.total,
		MemUsed:   memInfo.used,
		DiskTotal: diskInfo.total,
		DiskUsed:  diskInfo.used,
		Uptime:    collectUptime(),
	}
}

func collectCPUUsage() float64 {
	usage, err := gopsutilcpu.Percent(200*time.Millisecond, false)
	if err != nil || len(usage) == 0 {
		log.Printf("collect cpu usage failed: %v", err)
		return 0
	}

	return usage[0]
}

func collectMemoryInfo() memoryStats {
	info, err := gopsutilmem.VirtualMemory()
	if err != nil {
		log.Printf("collect memory info failed: %v", err)
		return memoryStats{}
	}

	return memoryStats{
		total: info.Total,
		used:  info.Used,
	}
}

func collectDiskInfo() diskStats {
	info, err := gopsutildisk.Usage(getSystemDiskPath())
	if err != nil {
		log.Printf("collect disk info failed: %v", err)
		return diskStats{}
	}

	return diskStats{
		total: info.Total,
		used:  info.Used,
	}
}

func collectUptime() uint64 {
	uptime, err := gopsutilhost.Uptime()
	if err != nil {
		log.Printf("collect uptime failed: %v", err)
		return 0
	}

	return uptime
}

func getSystemDiskPath() string {
	if runtime.GOOS != "windows" {
		return "/"
	}

	systemDrive := os.Getenv("SystemDrive")
	if systemDrive == "" {
		return defaultWindowsDiskPath
	}

	return systemDrive + `\`
}
