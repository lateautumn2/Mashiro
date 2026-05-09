package controllers

import (
	"fmt"
	"net"
	"net/url"
	"regexp"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"backend/db"
	"backend/models"

	"github.com/gin-gonic/gin"
)

const defaultLatencyIntervalSec = 60

var hostnameLabelRegex = regexp.MustCompile(`^[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?$`)

type LatencyTaskInput struct {
	Name        string `json:"name" binding:"required"`
	Type        string `json:"type" binding:"required"`
	Target      string `json:"target" binding:"required"`
	IntervalSec int    `json:"interval_sec"`
	ServerIDs   []uint `json:"server_ids"`
}

type LatencyTaskAgentResponse struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Target      string `json:"target"`
	IntervalSec int    `json:"interval_sec"`
}

type LatencyTaskServerResponse struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type LatencyResultResponse struct {
	TaskID       uint    `json:"task_id"`
	ServerID     uint    `json:"server_id"`
	TaskName     string  `json:"task_name"`
	Type         string  `json:"type"`
	Target       string  `json:"target"`
	LatencyMS    float64 `json:"latency_ms"`
	Status       string  `json:"status"`
	ErrorMessage string  `json:"error_message"`
	CheckedAt    string  `json:"checked_at"`
}

type LatencyTaskResponse struct {
	ID          uint                        `json:"id"`
	Name        string                      `json:"name"`
	Type        string                      `json:"type"`
	Target      string                      `json:"target"`
	IntervalSec int                         `json:"interval_sec"`
	ServerIDs   []uint                      `json:"server_ids"`
	Servers     []LatencyTaskServerResponse `json:"servers"`
	Results     []LatencyResultResponse     `json:"results"`
}

func AdminListLatencyTasks(c *gin.Context) {
	tasks, err := loadLatencyTasks()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load latency tasks"})
		return
	}

	c.JSON(http.StatusOK, tasks)
}

func AdminCreateLatencyTask(c *gin.Context) {
	var input LatencyTaskInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	task, err := buildLatencyTaskModel(input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := db.DB.Create(&task).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create latency task"})
		return
	}

	if err := replaceLatencyTaskServers(task.ID, input.ServerIDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save latency task servers"})
		return
	}

	respondLatencyTaskByID(c, task.ID)
}

func AdminUpdateLatencyTask(c *gin.Context) {
	var task models.LatencyTask
	if err := db.DB.First(&task, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Latency task not found"})
		return
	}

	var input LatencyTaskInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updated, err := buildLatencyTaskModel(input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	task.Name = updated.Name
	task.Type = updated.Type
	task.Target = updated.Target
	task.IntervalSec = updated.IntervalSec
	task.Enabled = true

	if err := db.DB.Save(&task).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update latency task"})
		return
	}

	if err := replaceLatencyTaskServers(task.ID, input.ServerIDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update latency task servers"})
		return
	}

	respondLatencyTaskByID(c, task.ID)
}

func AdminDeleteLatencyTask(c *gin.Context) {
	id := c.Param("id")
	db.DB.Unscoped().Where("task_id = ?", id).Delete(&models.LatencyTaskServer{})
	db.DB.Unscoped().Where("task_id = ?", id).Delete(&models.LatencyResult{})

	if err := db.DB.Delete(&models.LatencyTask{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete latency task"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "success"})
}

func AgentGetLatencyTasks(c *gin.Context) {
	token := c.GetString("agent_token")
	var server models.Server
	if err := db.DB.Where("auth_token = ?", token).First(&server).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid server token"})
		return
	}

	var tasks []models.LatencyTask
	db.DB.
		Joins("JOIN latency_task_servers ON latency_task_servers.task_id = latency_tasks.id").
		Where("latency_task_servers.server_id = ? AND latency_tasks.enabled = ?", server.ID, true).
		Order("latency_tasks.id desc").
		Find(&tasks)

	response := make([]LatencyTaskAgentResponse, 0, len(tasks))
	for _, task := range tasks {
		response = append(response, LatencyTaskAgentResponse{
			ID:          task.ID,
			Name:        task.Name,
			Type:        task.Type,
			Target:      task.Target,
			IntervalSec: task.IntervalSec,
		})
	}

	c.JSON(http.StatusOK, response)
}

func loadLatencyTasks() ([]LatencyTaskResponse, error) {
	var tasks []models.LatencyTask
	if err := db.DB.Order("id desc").Find(&tasks).Error; err != nil {
		return nil, err
	}

	results := make([]LatencyTaskResponse, 0, len(tasks))
	for _, task := range tasks {
		taskResponse, err := buildLatencyTaskResponse(task)
		if err != nil {
			return nil, err
		}
		results = append(results, taskResponse)
	}

	return results, nil
}

func respondLatencyTaskByID(c *gin.Context, id uint) {
	var task models.LatencyTask
	if err := db.DB.First(&task, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Latency task not found"})
		return
	}

	response, err := buildLatencyTaskResponse(task)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to build latency task response"})
		return
	}

	c.JSON(http.StatusOK, response)
}

func buildLatencyTaskModel(input LatencyTaskInput) (models.LatencyTask, error) {
	taskType := strings.ToLower(strings.TrimSpace(input.Type))
	if !slices.Contains([]string{"tcp", "icmp", "http"}, taskType) {
		return models.LatencyTask{}, fmt.Errorf("type must be one of tcp, icmp, http")
	}

	name := strings.TrimSpace(input.Name)
	target := strings.TrimSpace(input.Target)
	if name == "" || target == "" {
		return models.LatencyTask{}, fmt.Errorf("name and target are required")
	}

	normalizedTarget, err := validateLatencyTarget(taskType, target)
	if err != nil {
		return models.LatencyTask{}, err
	}

	return models.LatencyTask{
		Name:        name,
		Type:        taskType,
		Target:      normalizedTarget,
		IntervalSec: normalizeLatencyInterval(input.IntervalSec),
		Enabled:     true,
	}, nil
}

func normalizeLatencyInterval(interval int) int {
	if interval < 5 {
		return defaultLatencyIntervalSec
	}
	return interval
}

func validateLatencyTarget(taskType, target string) (string, error) {
	switch taskType {
	case "tcp":
		return validateTCPTarget(target)
	case "icmp":
		return validateICMPTarget(target)
	case "http":
		return validateHTTPTarget(target)
	default:
		return "", fmt.Errorf("unsupported latency task type")
	}
}

func validateTCPTarget(target string) (string, error) {
	trimmed := strings.TrimSpace(target)
	host, port, err := net.SplitHostPort(trimmed)
	if err != nil {
		return "", fmt.Errorf("TCP 目标必须填写为 host:port，例如 example.com:443")
	}
	if !isValidLatencyHost(host) {
		return "", fmt.Errorf("TCP 目标主机无效")
	}
	if !isValidLatencyPort(port) {
		return "", fmt.Errorf("TCP 端口必须是 1-65535 之间的数字")
	}
	return trimmed, nil
}

func validateICMPTarget(target string) (string, error) {
	trimmed := strings.TrimSpace(target)
	if !isValidLatencyHost(trimmed) {
		return "", fmt.Errorf("ICMP 目标必须是合法的 IP 或主机名")
	}
	return trimmed, nil
}

func validateHTTPTarget(target string) (string, error) {
	trimmed := strings.TrimSpace(target)
	parsed, err := url.Parse(trimmed)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "", fmt.Errorf("HTTP 目标必须是完整 URL，例如 https://example.com/health")
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", fmt.Errorf("HTTP 目标必须以 http:// 或 https:// 开头")
	}
	return trimmed, nil
}

func isValidLatencyHost(host string) bool {
	trimmed := strings.Trim(strings.TrimSpace(host), "[]")
	if trimmed == "" || strings.ContainsAny(trimmed, "/\\ \t\r\n") {
		return false
	}
	if ip := net.ParseIP(trimmed); ip != nil {
		return true
	}

	labels := strings.Split(trimmed, ".")
	for _, label := range labels {
		if !hostnameLabelRegex.MatchString(label) {
			return false
		}
	}
	return true
}

func isValidLatencyPort(port string) bool {
	value, err := strconv.Atoi(port)
	return err == nil && value >= 1 && value <= 65535
}

func replaceLatencyTaskServers(taskID uint, serverIDs []uint) error {
	normalized := uniqueServerIDs(serverIDs)
	if err := db.DB.Unscoped().Where("task_id = ?", taskID).Delete(&models.LatencyTaskServer{}).Error; err != nil {
		return err
	}

	assignments := make([]models.LatencyTaskServer, 0, len(normalized))
	for _, serverID := range normalized {
		assignments = append(assignments, models.LatencyTaskServer{
			TaskID:   taskID,
			ServerID: serverID,
		})
	}

	if len(assignments) == 0 {
		return nil
	}

	return db.DB.Create(&assignments).Error
}

func uniqueServerIDs(serverIDs []uint) []uint {
	seen := make(map[uint]struct{}, len(serverIDs))
	result := make([]uint, 0, len(serverIDs))
	for _, id := range serverIDs {
		if id == 0 {
			continue
		}
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result
}

func buildLatencyTaskResponse(task models.LatencyTask) (LatencyTaskResponse, error) {
	serverItems, serverIDs, err := loadLatencyTaskServers(task.ID)
	if err != nil {
		return LatencyTaskResponse{}, err
	}

	results, err := loadLatencyTaskResults(task.ID)
	if err != nil {
		return LatencyTaskResponse{}, err
	}

	return LatencyTaskResponse{
		ID:          task.ID,
		Name:        task.Name,
		Type:        task.Type,
		Target:      task.Target,
		IntervalSec: task.IntervalSec,
		ServerIDs:   serverIDs,
		Servers:     serverItems,
		Results:     results,
	}, nil
}

func loadLatencyTaskServers(taskID uint) ([]LatencyTaskServerResponse, []uint, error) {
	var assignments []models.LatencyTaskServer
	if err := db.DB.Where("task_id = ?", taskID).Find(&assignments).Error; err != nil {
		return nil, nil, err
	}

	serverIDs := make([]uint, 0, len(assignments))
	for _, item := range assignments {
		serverIDs = append(serverIDs, item.ServerID)
	}

	var servers []models.Server
	if len(serverIDs) > 0 {
		if err := db.DB.Where("id IN ?", serverIDs).Order("id asc").Find(&servers).Error; err != nil {
			return nil, nil, err
		}
	}

	serverItems := make([]LatencyTaskServerResponse, 0, len(servers))
	for _, server := range servers {
		serverItems = append(serverItems, LatencyTaskServerResponse{
			ID:   server.ID,
			Name: server.Name,
		})
	}

	return serverItems, serverIDs, nil
}

func loadLatencyTaskResults(taskID uint) ([]LatencyResultResponse, error) {
	var items []models.LatencyResult
	if err := db.DB.Where("task_id = ?", taskID).Order("server_id asc").Find(&items).Error; err != nil {
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
