package notifier

import (
	"time"

	"backend/db"
	"backend/models"
)

// StartChecker launches a periodic goroutine that checks all servers for
// traffic warnings and expiry warnings every minute.
func StartChecker() {
	go func() {
		for {
			time.Sleep(60 * time.Second)
			checkAll()
		}
	}()
}

func checkAll() {
	var servers []models.Server
	if err := db.DB.Find(&servers).Error; err != nil {
		return
	}

	for _, s := range servers {
		// Find the latest agent data for traffic calculation.
		var latest models.AgentData
		if err := db.DB.Where("server_id = ?", s.ID).Order("report_time desc").First(&latest).Error; err != nil {
			if s.ExpiryTime.IsZero() {
				continue
			}
			// No agent data but check expiry.
			NotifyExpiryIfNeeded(s)
			continue
		}

		trafficIn := latest.NetIn - s.TrafficResetNetInBase
		if latest.NetIn < s.TrafficResetNetInBase {
			trafficIn = 0
		}
		trafficOut := latest.NetOut - s.TrafficResetNetOutBase
		if latest.NetOut < s.TrafficResetNetOutBase {
			trafficOut = 0
		}

		NotifyTrafficIfNeeded(s, trafficIn, trafficOut)
		NotifyExpiryIfNeeded(s)
	}
}
