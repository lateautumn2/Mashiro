package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"backend/db"
	"backend/models"
)

const (
	telegramAPI       = "https://api.telegram.org"
	trafficThreshold  = 0.85 // 85%
	expiryWarningDays = 15
	notifyCooldown    = 30 * time.Minute
)

// config holds cached Telegram configuration loaded from AppConfig.
type config struct {
	BotToken string
	ChatID   string
}

func loadConfig() (config, error) {
	var rows []models.AppConfig
	if err := db.DB.Where("key IN ?", []string{"tg_bot_token", "tg_chat_id"}).Find(&rows).Error; err != nil {
		return config{}, fmt.Errorf("failed to load notification config: %w", err)
	}

	cfg := config{}
	for _, row := range rows {
		switch row.Key {
		case "tg_bot_token":
			cfg.BotToken = strings.TrimSpace(row.Value)
		case "tg_chat_id":
			cfg.ChatID = strings.TrimSpace(row.Value)
		}
	}

	if cfg.BotToken == "" || cfg.ChatID == "" {
		return config{}, fmt.Errorf("telegram notification not configured (missing bot_token or chat_id)")
	}

	return cfg, nil
}

func sendTelegram(cfg config, message string) error {
	url := fmt.Sprintf("%s/bot%s/sendMessage", telegramAPI, cfg.BotToken)

	body := map[string]string{
		"chat_id":    cfg.ChatID,
		"text":       message,
		"parse_mode": "HTML",
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal telegram payload: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to create telegram request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send telegram message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("telegram API returned status %d", resp.StatusCode)
	}

	return nil
}

// SendSilent sends a notification message.  If the notification method is not
// configured, the error is silently swallowed (graceful fallback).
func SendSilent(message string) {
	cfg, err := loadConfig()
	if err != nil {
		return
	}

	if err := sendTelegram(cfg, message); err != nil {
		// Logging just to stderr, not crashing the caller.
		fmt.Printf("[notifier] failed to send message: %v\n", err)
	}
}

// formatOnlineMsg builds an HTML-formatted online notification message.
func formatOnlineMsg(serverName, serverIP, region string) string {
	return fmt.Sprintf(
		"🟢 <b>%s</b> is now <b>online</b>\nIP: %s\nRegion: %s\nTime: %s",
		escapeHTML(serverName),
		serverIP,
		region,
		time.Now().Format("2006-01-02 15:04:05"),
	)
}

// formatOfflineMsg builds an HTML-formatted offline notification message.
func formatOfflineMsg(serverName, serverIP, region string) string {
	return fmt.Sprintf(
		"🔴 <b>%s</b> went <b>offline</b>\nIP: %s\nRegion: %s\nTime: %s",
		escapeHTML(serverName),
		serverIP,
		region,
		time.Now().Format("2006-01-02 15:04:05"),
	)
}

// formatTrafficMsg builds a traffic warning message.
func formatTrafficMsg(serverName, serverIP string, usedGB, limitGB float64, pct int) string {
	return fmt.Sprintf(
		"⚠️ <b>%s</b> traffic warning\nIP: %s\nUsage: %.2f GB / %.2f GB (%d%%)\nTime: %s",
		escapeHTML(serverName),
		serverIP,
		usedGB,
		limitGB,
		pct,
		time.Now().Format("2006-01-02 15:04:05"),
	)
}

// formatExpiryMsg builds an expiry warning message.
func formatExpiryMsg(serverName, serverIP string, expiryDate string, daysLeft int) string {
	return fmt.Sprintf(
		"⌛ <b>%s</b> expires soon\nIP: %s\nExpiry: %s (%d days left)\nTime: %s",
		escapeHTML(serverName),
		serverIP,
		expiryDate,
		daysLeft,
		time.Now().Format("2006-01-02 15:04:05"),
	)
}

func escapeHTML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}

// NotifyOnline sends an online notification for the given server if preferences allow.
func NotifyOnline(server models.Server) {
	var pref models.NotificationPref
	if err := db.DB.First(&pref, server.ID).Error; err != nil {
		// No preference record = treat as default (enabled).
	}
	if !pref.NotifyOnline {
		return
	}
	SendSilent(formatOnlineMsg(server.Name, server.IP, server.Region))
}

// NotifyOffline sends an offline notification for the given server if preferences allow.
func NotifyOffline(server models.Server) {
	var pref models.NotificationPref
	if err := db.DB.First(&pref, server.ID).Error; err != nil {
	}
	if !pref.NotifyOffline {
		return
	}
	SendSilent(formatOfflineMsg(server.Name, server.IP, server.Region))
}

// NotifyTrafficIfNeeded checks traffic usage and sends warning if >85%.
func NotifyTrafficIfNeeded(server models.Server, trafficIn, trafficOut uint64) {
	var pref models.NotificationPref
	if err := db.DB.Where("server_id = ?", server.ID).First(&pref).Error; err != nil {
		return
	}
	if !pref.NotifyTraffic {
		return
	}

	if server.TrafficLimit == 0 {
		return
	}

	used := trafficIn + trafficOut
	ratio := float64(used) / float64(server.TrafficLimit)
	if ratio < trafficThreshold {
		// Traffic dropped below threshold; clear the warning sent flag.
		if pref.TrafficWarningSent {
			db.DB.Model(&pref).Update("traffic_warning_sent", false)
		}
		return
	}

	if pref.TrafficWarningSent {
		return
	}

	usedGB := float64(used) / 1024 / 1024 / 1024
	limitGB := float64(server.TrafficLimit) / 1024 / 1024 / 1024
	pct := int(ratio * 100)

	SendSilent(formatTrafficMsg(server.Name, server.IP, usedGB, limitGB, pct))
	db.DB.Model(&pref).Update("traffic_warning_sent", true)
}

// NotifyExpiryIfNeeded checks server expiry and sends warning if within 15 days.
func NotifyExpiryIfNeeded(server models.Server) {
	if server.ExpiryTime.IsZero() {
		return
	}

	var pref models.NotificationPref
	if err := db.DB.Where("server_id = ?", server.ID).First(&pref).Error; err != nil {
		return
	}
	if !pref.NotifyExpiry {
		return
	}

	now := time.Now()
	daysLeft := int(server.ExpiryTime.Sub(now).Hours() / 24)
	if daysLeft > expiryWarningDays {
		return
	}
	if daysLeft < 0 {
		return // already expired
	}

	// Only notify once every 24 hours.
	if now.Sub(pref.LastExpiryNotifiedAt) < 24*time.Hour {
		return
	}

	expiryStr := server.ExpiryTime.Format("2006-01-02")
	SendSilent(formatExpiryMsg(server.Name, server.IP, expiryStr, daysLeft))
	db.DB.Model(&pref).Update("last_expiry_notified_at", now)
}

// ResetTrafficWarning clears the traffic warning flag for a server (called on traffic reset).
func ResetTrafficWarning(serverID uint) {
	db.DB.Model(&models.NotificationPref{}).Where("server_id = ?", serverID).Update("traffic_warning_sent", false)
}
