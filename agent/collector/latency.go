package collector

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	latencyTimeout       = 5 * time.Second
	latencyProbeCount    = 3
	latencyProbeInterval = 200 * time.Millisecond
)

var hostnameLabelPattern = regexp.MustCompile(`^[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?$`)

type LatencyCheckResult struct {
	LatencyMS    float64 `json:"latency_ms"`
	Status       string  `json:"status"`
	ErrorMessage string  `json:"error_message"`
}

func CheckLatency(protocol, target string) LatencyCheckResult {
	switch strings.ToLower(strings.TrimSpace(protocol)) {
	case "tcp":
		return checkTCPLatency(target)
	case "icmp":
		return checkICMPLatency(target)
	case "http":
		return checkHTTPLatency(target)
	default:
		return LatencyCheckResult{Status: "error", ErrorMessage: "unsupported protocol"}
	}
}

func checkTCPLatency(target string) LatencyCheckResult {
	address, err := validateTCPTarget(target)
	if err != nil {
		return LatencyCheckResult{Status: "error", ErrorMessage: err.Error()}
	}

	latencyMS, err := sampleLatency(func() (float64, error) {
		return measureTCPLatency(address)
	})
	if err != nil {
		return LatencyCheckResult{Status: "error", ErrorMessage: err.Error()}
	}

	return LatencyCheckResult{LatencyMS: latencyMS, Status: "ok"}
}

func checkHTTPLatency(target string) LatencyCheckResult {
	requestURL, err := validateHTTPTarget(target)
	if err != nil {
		return LatencyCheckResult{Status: "error", ErrorMessage: err.Error()}
	}

	client := &http.Client{Timeout: latencyTimeout}
	latencyMS, err := sampleLatency(func() (float64, error) {
		return measureHTTPLatency(client, requestURL)
	})
	if err != nil {
		return LatencyCheckResult{Status: "error", ErrorMessage: err.Error()}
	}

	return LatencyCheckResult{LatencyMS: latencyMS, Status: "ok"}
}

func checkICMPLatency(target string) LatencyCheckResult {
	host, err := validateICMPTarget(target)
	if err != nil {
		return LatencyCheckResult{Status: "error", ErrorMessage: err.Error()}
	}

	if runtime.GOOS == "windows" {
		return checkWindowsICMPLatency(host)
	}

	command := exec.Command("ping", "-c", "1", "-W", "5", host)
	output, err := command.CombinedOutput()
	if err != nil {
		message := strings.TrimSpace(string(output))
		if message == "" {
			message = err.Error()
		}
		return LatencyCheckResult{Status: "error", ErrorMessage: message}
	}

	latencyMS, parsed := parsePingLatency(output)
	if parsed {
		return LatencyCheckResult{
			LatencyMS: latencyMS,
			Status:    "ok",
		}
	}

	return LatencyCheckResult{Status: "error", ErrorMessage: "failed to parse icmp response time"}
}

func checkWindowsICMPLatency(target string) LatencyCheckResult {
	command := exec.Command(
		"powershell",
		"-NoProfile",
		"-Command",
		fmt.Sprintf(`$result = Test-Connection -Count 1 -TargetName '%s' -ErrorAction Stop; Write-Output $result.ResponseTime`, escapePowerShellSingleQuoted(target)),
	)

	output, err := command.CombinedOutput()
	if err != nil {
		message := strings.TrimSpace(string(output))
		if message == "" {
			message = err.Error()
		}
		return LatencyCheckResult{Status: "error", ErrorMessage: message}
	}

	value, err := strconv.ParseFloat(strings.TrimSpace(string(output)), 64)
	if err != nil {
		return LatencyCheckResult{Status: "error", ErrorMessage: "failed to parse icmp response time"}
	}

	return LatencyCheckResult{
		LatencyMS: value,
		Status:    "ok",
	}
}

var (
	windowsAverageLatencyRegex = regexp.MustCompile(`Average = (\d+)ms`)
	commonPingLatencyRegex     = regexp.MustCompile(`time[=<]\s*([0-9.]+)\s*ms`)
)

func parsePingLatency(output []byte) (float64, bool) {
	text := string(output)

	if matches := windowsAverageLatencyRegex.FindStringSubmatch(text); len(matches) == 2 {
		value, err := strconv.ParseFloat(matches[1], 64)
		if err == nil {
			return value, true
		}
	}

	if matches := commonPingLatencyRegex.FindStringSubmatch(text); len(matches) == 2 {
		value, err := strconv.ParseFloat(matches[1], 64)
		if err == nil {
			return value, true
		}
	}

	return 0, false
}

func validateTCPTarget(target string) (string, error) {
	trimmed := strings.TrimSpace(target)
	if trimmed == "" {
		return "", fmt.Errorf("tcp target is required")
	}

	host, port, err := net.SplitHostPort(trimmed)
	if err != nil {
		return "", fmt.Errorf("tcp target must be host:port, for example example.com:443")
	}
	if !isValidLatencyHost(host) {
		return "", fmt.Errorf("invalid tcp host")
	}
	if !isValidPort(port) {
		return "", fmt.Errorf("invalid tcp port")
	}

	return trimmed, nil
}

func validateHTTPTarget(target string) (string, error) {
	trimmed := strings.TrimSpace(target)
	parsed, err := url.Parse(trimmed)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "", fmt.Errorf("http target must be a full URL, for example https://example.com/health")
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", fmt.Errorf("http target must start with http:// or https://")
	}

	return trimmed, nil
}

func validateICMPTarget(target string) (string, error) {
	trimmed := strings.TrimSpace(target)
	if trimmed == "" {
		return "", fmt.Errorf("icmp target is required")
	}
	if !isValidLatencyHost(trimmed) {
		return "", fmt.Errorf("icmp target must be a valid IP or hostname")
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
		if !hostnameLabelPattern.MatchString(label) {
			return false
		}
	}

	return true
}

func isValidPort(port string) bool {
	value, err := strconv.Atoi(port)
	return err == nil && value >= 1 && value <= 65535
}

func sampleLatency(measure func() (float64, error)) (float64, error) {
	total := 0.0
	for i := 0; i < latencyProbeCount; i++ {
		latencyMS, err := measure()
		if err != nil {
			return 0, err
		}
		total += latencyMS
		if i < latencyProbeCount-1 {
			time.Sleep(latencyProbeInterval)
		}
	}

	return total / float64(latencyProbeCount), nil
}

func measureTCPLatency(address string) (float64, error) {
	start := time.Now()
	conn, err := net.DialTimeout("tcp", address, latencyTimeout)
	if err != nil {
		return 0, err
	}
	_ = conn.Close()
	return durationToMilliseconds(time.Since(start)), nil
}

func measureHTTPLatency(client *http.Client, target string) (float64, error) {
	start := time.Now()
	resp, err := client.Get(target)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return 0, fmt.Errorf("http status %d", resp.StatusCode)
	}

	return durationToMilliseconds(time.Since(start)), nil
}

func escapePowerShellSingleQuoted(value string) string {
	return strings.ReplaceAll(value, `'`, `''`)
}

func durationToMilliseconds(duration time.Duration) float64 {
	return float64(duration.Microseconds()) / 1000
}
