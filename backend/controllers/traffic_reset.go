package controllers

import (
	"time"

	"backend/models"
	"backend/notifier"
)

const defaultTrafficResetDay = 1

func normalizeTrafficResetDay(day int) int {
	if day < 1 {
		return defaultTrafficResetDay
	}
	if day > 31 {
		return 31
	}
	return day
}

func syncTrafficResetState(server *models.Server, currentNetIn, currentNetOut uint64, now time.Time) bool {
	modified := false
	resetDay := normalizeTrafficResetDay(server.TrafficResetDay)
	if server.TrafficResetDay != resetDay {
		server.TrafficResetDay = resetDay
		modified = true
	}

	windowStart := currentTrafficWindowStart(now.UTC(), resetDay)
	if server.TrafficResetAt.IsZero() || server.TrafficResetAt.Before(windowStart) {
		server.TrafficResetAt = windowStart
		server.TrafficResetNetInBase = currentNetIn
		server.TrafficResetNetOutBase = currentNetOut
		notifier.ResetTrafficWarning(server.ID)
		return true
	}

	if currentNetIn < server.TrafficResetNetInBase || currentNetOut < server.TrafficResetNetOutBase {
		server.TrafficResetAt = now.UTC()
		server.TrafficResetNetInBase = currentNetIn
		server.TrafficResetNetOutBase = currentNetOut
		notifier.ResetTrafficWarning(server.ID)
		return true
	}

	return modified
}

func currentTrafficUsage(server models.Server, latest models.AgentData) (uint64, uint64) {
	netIn := uint64(0)
	netOut := uint64(0)
	if latest.NetIn >= server.TrafficResetNetInBase {
		netIn = latest.NetIn - server.TrafficResetNetInBase
	}
	if latest.NetOut >= server.TrafficResetNetOutBase {
		netOut = latest.NetOut - server.TrafficResetNetOutBase
	}
	return netIn, netOut
}

func currentTrafficWindowStart(now time.Time, resetDay int) time.Time {
	year, month, _ := now.Date()
	location := time.UTC

	currentMonthReset := time.Date(year, month, clampDay(year, month, resetDay), 0, 0, 0, 0, location)
	if !now.Before(currentMonthReset) {
		return currentMonthReset
	}

	previousMonth := currentMonthReset.AddDate(0, -1, 0)
	prevYear, prevMonth, _ := previousMonth.Date()
	return time.Date(prevYear, prevMonth, clampDay(prevYear, prevMonth, resetDay), 0, 0, 0, 0, location)
}

func clampDay(year int, month time.Month, day int) int {
	lastDay := time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
	if day > lastDay {
		return lastDay
	}
	return day
}
