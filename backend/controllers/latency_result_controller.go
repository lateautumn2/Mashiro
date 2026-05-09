package controllers

import (
	"net/http"
	"time"

	"backend/db"
	"backend/models"

	"github.com/gin-gonic/gin"
)

type AgentLatencyResultInput struct {
	TaskID        uint    `json:"task_id"`
	TaskName      string  `json:"task_name"`
	Type          string  `json:"type"`
	Target        string  `json:"target"`
	LatencyMS     float64 `json:"latency_ms"`
	Status        string  `json:"status"`
	ErrorMessage  string  `json:"error_message"`
}

type AgentLatencyReportInput struct {
	Results []AgentLatencyResultInput `json:"results"`
}

func AgentReportLatencyResults(c *gin.Context) {
	token := c.GetString("agent_token")
	var server models.Server
	if err := db.DB.Where("auth_token = ?", token).First(&server).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid server token"})
		return
	}

	var input AgentLatencyReportInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	now := time.Now().UTC()
	for _, item := range input.Results {
		if item.TaskID == 0 {
			continue
		}

		record := models.LatencyResult{
			TaskID:       item.TaskID,
			ServerID:     server.ID,
			TaskName:     item.TaskName,
			Type:         item.Type,
			Target:       item.Target,
			LatencyMS:    item.LatencyMS,
			Status:       item.Status,
			ErrorMessage: item.ErrorMessage,
			CheckedAt:    now,
		}

		if err := db.DB.Where("task_id = ? AND server_id = ?", item.TaskID, server.ID).Assign(record).FirstOrCreate(&models.LatencyResult{}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save latency result"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "success"})
}
