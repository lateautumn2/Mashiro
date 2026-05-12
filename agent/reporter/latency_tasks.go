package reporter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"agent/collector"
	"agent/signer"
)

type LatencyTask struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Target      string `json:"target"`
	IntervalSec int    `json:"interval_sec"`
}

type LatencyTaskResult struct {
	TaskID       uint    `json:"task_id"`
	TaskName     string  `json:"task_name"`
	Type         string  `json:"type"`
	Target       string  `json:"target"`
	LatencyMS    float64 `json:"latency_ms"`
	Status       string  `json:"status"`
	ErrorMessage string  `json:"error_message"`
}

func (r *Reporter) FetchLatencyTasks() ([]LatencyTask, error) {
	req, err := http.NewRequest("GET", r.latencyTasksURL(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create latency task request: %w", err)
	}
	req.Header.Set("X-Agent-Token", r.AgentID)
	sig, ts := signer.SignRequest("GET", req.URL.Path, nil, r.AgentID)
	req.Header.Set("X-Signature", sig)
	req.Header.Set("X-Timestamp", ts)

	resp, err := r.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch latency tasks: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("latency task request returned status: %d", resp.StatusCode)
	}

	var tasks []LatencyTask
	if err := json.NewDecoder(resp.Body).Decode(&tasks); err != nil {
		return nil, fmt.Errorf("failed to decode latency tasks: %w", err)
	}

	return tasks, nil
}

func (r *Reporter) ReportLatencyResults(results []LatencyTaskResult) error {
	data, err := json.Marshal(map[string]any{"results": results})
	if err != nil {
		return fmt.Errorf("failed to marshal latency results: %w", err)
	}

	req, err := http.NewRequest("POST", r.latencyResultsURL(), bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to create latency result request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Agent-Token", r.AgentID)
	sig, ts := signer.SignRequest("POST", req.URL.Path, data, r.AgentID)
	req.Header.Set("X-Signature", sig)
	req.Header.Set("X-Timestamp", ts)

	resp, err := r.Client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send latency results: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("latency result request returned status: %d", resp.StatusCode)
	}

	return nil
}

func (r *Reporter) RunLatencyTasks(now time.Time, lastRuns map[uint]time.Time) error {
	tasks, err := r.FetchLatencyTasks()
	if err != nil {
		return err
	}

	results := make([]LatencyTaskResult, 0, len(tasks))
	for _, task := range tasks {
		interval := time.Duration(max(task.IntervalSec, 5)) * time.Second
		if lastRun, exists := lastRuns[task.ID]; exists && now.Sub(lastRun) < interval {
			continue
		}

		checkResult := collector.CheckLatency(task.Type, task.Target)
		results = append(results, LatencyTaskResult{
			TaskID:       task.ID,
			TaskName:     task.Name,
			Type:         task.Type,
			Target:       task.Target,
			LatencyMS:    checkResult.LatencyMS,
			Status:       checkResult.Status,
			ErrorMessage: checkResult.ErrorMessage,
		})
		lastRuns[task.ID] = now
	}

	if len(results) == 0 {
		return nil
	}

	return r.ReportLatencyResults(results)
}

func (r *Reporter) latencyTasksURL() string {
	return strings.TrimSuffix(r.ServerURL, "/report") + "/latency/tasks"
}

func (r *Reporter) latencyResultsURL() string {
	return strings.TrimSuffix(r.ServerURL, "/report") + "/latency/results"
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
