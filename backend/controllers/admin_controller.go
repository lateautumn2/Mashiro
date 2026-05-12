package controllers

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	"backend/db"
	"backend/models"

	"github.com/gin-gonic/gin"
)

type AddServerInput struct {
	Name         string `json:"name" binding:"required"`
	IP           string `json:"ip"`
	Region       string `json:"region"`
	System       string `json:"system"`
	HasIPv4      bool   `json:"has_ipv4"`
	HasIPv6      bool   `json:"has_ipv6"`
	TrafficLimit uint64 `json:"traffic_limit"`
	TrafficResetDay int  `json:"traffic_reset_day"`
	ExpiryTime   string `json:"expiry_time"`
}

type AdminServerResponse struct {
	ID           uint   `json:"id"`
	Name         string `json:"name"`
	IP           string `json:"ip"`
	Region       string `json:"region"`
	System       string `json:"system"`
	HasIPv4      bool   `json:"has_ipv4"`
	HasIPv6      bool   `json:"has_ipv6"`
	Status       string `json:"status"`
	TrafficLimit uint64 `json:"traffic_limit"`
	TrafficResetDay int  `json:"traffic_reset_day"`
	ExpiryTime   string `json:"expiry_time"`
}

func toAdminServerResponse(server models.Server) AdminServerResponse {
	expiryTime := ""
	if !server.ExpiryTime.IsZero() {
		expiryTime = server.ExpiryTime.Format(time.RFC3339)
	}

	return AdminServerResponse{
		ID:           server.ID,
		Name:         server.Name,
		IP:           server.IP,
		Region:       server.Region,
		System:       server.System,
		HasIPv4:      server.HasIPv4,
		HasIPv6:      server.HasIPv6,
		Status:       server.Status,
		TrafficLimit: server.TrafficLimit,
		TrafficResetDay: normalizeTrafficResetDay(server.TrafficResetDay),
		ExpiryTime:   expiryTime,
	}
}

func parseExpiryTime(value string) (time.Time, error) {
	if value == "" {
		return time.Time{}, nil
	}

	return time.Parse(time.RFC3339, value)
}

func AdminGetServers(c *gin.Context) {
	var servers []models.Server
	db.DB.Order("id desc").Find(&servers)
	now := time.Now().UTC()

	response := make([]AdminServerResponse, 0, len(servers))
	for i := range servers {
		server := &servers[i]
		latestData, hasLatest := latestAgentDataForServer(server.ID)
		syncServerStatus(server, latestData, hasLatest, now)
		response = append(response, toAdminServerResponse(*server))
	}

	c.JSON(http.StatusOK, response)
}

func AdminAddServer(c *gin.Context) {
	var input AddServerInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	name := strings.TrimSpace(input.Name)
	ip := strings.TrimSpace(input.IP)
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}

	authToken, err := generateAuthToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate server token"})
		return
	}

	expiryTime, err := parseExpiryTime(input.ExpiryTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "expiry_time must be RFC3339 format"})
		return
	}

	server := models.Server{
		Name:            name,
		IP:              ip,
		Region:          strings.TrimSpace(input.Region),
		System:          strings.TrimSpace(input.System),
		HasIPv4:         input.HasIPv4,
		HasIPv6:         input.HasIPv6,
		Status:          "offline",
		AuthToken:       authToken,
		TrafficLimit:    input.TrafficLimit,
		TrafficResetDay: normalizeTrafficResetDay(input.TrafficResetDay),
		ExpiryTime:      expiryTime,
	}

	if err := db.DB.Create(&server).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create server"})
		return
	}

	c.JSON(http.StatusOK, toAdminServerResponse(server))
}

func AdminUpdateServer(c *gin.Context) {
	id := c.Param("id")

	var server models.Server
	if err := db.DB.First(&server, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Server not found"})
		return
	}

	var input AddServerInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	name := strings.TrimSpace(input.Name)
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}

	expiryTime, err := parseExpiryTime(input.ExpiryTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "expiry_time must be RFC3339 format"})
		return
	}

	server.Name = name
	server.IP = strings.TrimSpace(input.IP)
	server.TrafficLimit = input.TrafficLimit
	server.TrafficResetDay = normalizeTrafficResetDay(input.TrafficResetDay)
	server.ExpiryTime = expiryTime

	if err := db.DB.Save(&server).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update server"})
		return
	}

	c.JSON(http.StatusOK, toAdminServerResponse(server))
}

func AdminDeleteServer(c *gin.Context) {
	id := c.Param("id")
	if err := db.DB.Where("server_id = ?", id).Delete(&models.AgentData{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete server metrics"})
		return
	}

	result := db.DB.Delete(&models.Server{}, id)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete server"})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Server not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "success"})
}

func getRequestBaseURL(c *gin.Context) string {
	scheme := c.GetHeader("X-Forwarded-Proto")
	if scheme == "" {
		if c.Request.TLS != nil {
			scheme = "https"
		} else {
			scheme = "http"
		}
	}

	host := c.GetHeader("X-Forwarded-Host")
	if host == "" {
		host = c.Request.Host
	}

	return fmt.Sprintf("%s://%s", scheme, host)
}

func AdminGetDeployCommand(c *gin.Context) {
	id := c.Param("id")
	var server models.Server
	if err := db.DB.First(&server, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Server not found"})
		return
	}

	c.JSON(http.StatusOK, buildDeployCommands(c, server))
}

// generateAuthToken creates a cryptographically random 64-character hex token (256 bits).
func generateAuthToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}
