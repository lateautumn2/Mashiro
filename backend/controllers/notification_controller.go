package controllers

import (
	"net/http"

	"backend/db"
	"backend/models"

	"github.com/gin-gonic/gin"
)

type NotificationPrefResponse struct {
	ServerID       uint   `json:"server_id"`
	ServerName     string `json:"server_name"`
	NotifyOnline   bool   `json:"notify_online"`
	NotifyOffline  bool   `json:"notify_offline"`
	NotifyTraffic  bool   `json:"notify_traffic"`
	NotifyExpiry   bool   `json:"notify_expiry"`
}

type UpdateNotificationPrefInput struct {
	NotifyOnline  *bool `json:"notify_online"`
	NotifyOffline *bool `json:"notify_offline"`
	NotifyTraffic *bool `json:"notify_traffic"`
	NotifyExpiry  *bool `json:"notify_expiry"`
}

func GetNotificationPrefs(c *gin.Context) {
	var servers []models.Server
	db.DB.Order("id asc").Find(&servers)

	prefMap := make(map[uint]models.NotificationPref)
	var prefs []models.NotificationPref
	db.DB.Find(&prefs)
	for _, p := range prefs {
		prefMap[p.ServerID] = p
	}

	response := make([]NotificationPrefResponse, 0, len(servers))
	for _, s := range servers {
		p, exists := prefMap[s.ID]
		resp := NotificationPrefResponse{
			ServerID:      s.ID,
			ServerName:    s.Name,
			NotifyOnline:  true,
			NotifyOffline: true,
			NotifyTraffic: true,
			NotifyExpiry:  true,
		}
		if exists {
			resp.NotifyOnline = p.NotifyOnline
			resp.NotifyOffline = p.NotifyOffline
			resp.NotifyTraffic = p.NotifyTraffic
			resp.NotifyExpiry = p.NotifyExpiry
		}
		response = append(response, resp)
	}

	c.JSON(http.StatusOK, response)
}

func UpdateNotificationPref(c *gin.Context) {
	serverID := c.Param("server_id")

	var server models.Server
	if err := db.DB.First(&server, serverID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Server not found"})
		return
	}

	var input UpdateNotificationPrefInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var pref models.NotificationPref
	result := db.DB.Where("server_id = ?", server.ID).First(&pref)

	if result.Error != nil {
		// Create with defaults then apply updates.
		pref = models.NotificationPref{ServerID: server.ID}
	}

	if input.NotifyOnline != nil {
		pref.NotifyOnline = *input.NotifyOnline
	}
	if input.NotifyOffline != nil {
		pref.NotifyOffline = *input.NotifyOffline
	}
	if input.NotifyTraffic != nil {
		pref.NotifyTraffic = *input.NotifyTraffic
	}
	if input.NotifyExpiry != nil {
		pref.NotifyExpiry = *input.NotifyExpiry
	}

	if err := db.DB.Save(&pref).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save notification preferences"})
		return
	}

	c.JSON(http.StatusOK, NotificationPrefResponse{
		ServerID:      server.ID,
		ServerName:    server.Name,
		NotifyOnline:  pref.NotifyOnline,
		NotifyOffline: pref.NotifyOffline,
		NotifyTraffic: pref.NotifyTraffic,
		NotifyExpiry:  pref.NotifyExpiry,
	})
}
