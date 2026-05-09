package collector

import (
	"log"
	"sync"
	"time"

	gopsutilnet "github.com/shirou/gopsutil/v3/net"
)

type NetworkInfo struct {
	BytesSent   uint64  `json:"bytes_sent"`   // Total bytes sent
	BytesRecv   uint64  `json:"bytes_recv"`   // Total bytes received
	SpeedSend   float64 `json:"speed_send"`   // Bytes per second
	SpeedRecv   float64 `json:"speed_recv"`   // Bytes per second
	Latency     float64 `json:"latency"`      // milliseconds
	PacketLoss  float64 `json:"packet_loss"`  // percentage 0-100
}

var (
	networkSnapshotMu sync.Mutex
	lastSnapshot      networkSnapshot
)

type networkSnapshot struct {
	bytesSent  uint64
	bytesRecv  uint64
	capturedAt time.Time
}

// CollectNetworkInfo returns real network byte counters and transfer speed.
// Latency and packet loss require active probing, so they are left at zero here.
func CollectNetworkInfo() NetworkInfo {
	counters, err := gopsutilnet.IOCounters(false)
	if err != nil || len(counters) == 0 {
		log.Printf("collect network info failed: %v", err)
		return NetworkInfo{}
	}

	current := networkSnapshot{
		bytesSent:  counters[0].BytesSent,
		bytesRecv:  counters[0].BytesRecv,
		capturedAt: time.Now(),
	}
	speedSend, speedRecv := calculateNetworkSpeed(current)

	return NetworkInfo{
		BytesSent:  current.bytesSent,
		BytesRecv:  current.bytesRecv,
		SpeedSend:  speedSend,
		SpeedRecv:  speedRecv,
		Latency:    0,
		PacketLoss: 0,
	}
}

func calculateNetworkSpeed(current networkSnapshot) (float64, float64) {
	networkSnapshotMu.Lock()
	defer networkSnapshotMu.Unlock()

	if lastSnapshot.capturedAt.IsZero() {
		lastSnapshot = current
		return 0, 0
	}

	elapsed := current.capturedAt.Sub(lastSnapshot.capturedAt).Seconds()
	if elapsed <= 0 {
		lastSnapshot = current
		return 0, 0
	}

	speedSend := float64(current.bytesSent-lastSnapshot.bytesSent) / elapsed
	speedRecv := float64(current.bytesRecv-lastSnapshot.bytesRecv) / elapsed
	lastSnapshot = current

	return speedSend, speedRecv
}
