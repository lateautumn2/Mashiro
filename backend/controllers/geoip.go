package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

const geoCacheTTL = 12 * time.Hour

type geoLookupResult struct {
	Region string
}

type geoCacheEntry struct {
	result    geoLookupResult
	expiresAt time.Time
}

type ipwhoisResponse struct {
	Success     bool   `json:"success"`
	CountryCode string `json:"country_code"`
	Country     string `json:"country"`
	City        string `json:"city"`
}

var (
	geoCache   = map[string]geoCacheEntry{}
	geoCacheMu sync.Mutex
)

func lookupRegionByIP(ip string) string {
	normalized := strings.TrimSpace(ip)
	if normalized == "" {
		return ""
	}

	geoCacheMu.Lock()
	if entry, ok := geoCache[normalized]; ok && time.Now().Before(entry.expiresAt) {
		geoCacheMu.Unlock()
		return entry.result.Region
	}
	geoCacheMu.Unlock()

	client := &http.Client{Timeout: 4 * time.Second}
	resp, err := client.Get(fmt.Sprintf("https://ipwho.is/%s", normalized))
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return ""
	}

	var payload ipwhoisResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return ""
	}

	if !payload.Success {
		return ""
	}

	parts := make([]string, 0, 3)
	if payload.Country != "" {
		parts = append(parts, payload.Country)
	}
	if payload.City != "" {
		parts = append(parts, payload.City)
	}

	region := strings.Join(parts, " · ")
	geoCacheMu.Lock()
	geoCache[normalized] = geoCacheEntry{
		result:    geoLookupResult{Region: region},
		expiresAt: time.Now().Add(geoCacheTTL),
	}
	geoCacheMu.Unlock()

	return region
}
