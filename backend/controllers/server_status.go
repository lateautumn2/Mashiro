package controllers

import (
	"time"

	"backend/db"
	"backend/models"
)

const serverOfflineTimeout = 20 * time.Second

func latestAgentDataForServer(serverID uint) (models.AgentData, bool) {
	var latest models.AgentData
	if err := db.DB.Where("server_id = ?", serverID).Order("report_time desc").First(&latest).Error; err != nil {
		return models.AgentData{}, false
	}

	return latest, true
}

func syncServerStatus(server *models.Server, latest models.AgentData, hasLatest bool, now time.Time) {
	desiredStatus := "offline"
	if hasLatest && !latest.ReportTime.IsZero() && now.Sub(latest.ReportTime.UTC()) <= serverOfflineTimeout {
		desiredStatus = "online"
	}

	if server.Status == desiredStatus {
		return
	}

	server.Status = desiredStatus
	db.DB.Model(server).Update("status", desiredStatus)
}
