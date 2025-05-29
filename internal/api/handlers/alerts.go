/*
SPDX-License-Identifier: GPL-3.0-or-later

Copyright (C) 2025 Aaron Mathis aaron.mathis@gmail.com

This file is part of GoSight.

GoSight is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

GoSight is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with GoSight. If not, see https://www.gnu.org/licenses/.
*/

// Package handlers provides HTTP handlers for the GoSight API server.
package handlers

import (
	"encoding/json"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/aaronlmathis/gosight-server/internal/sys"
	"github.com/aaronlmathis/gosight-shared/model"
	"github.com/aaronlmathis/gosight-shared/utils"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// AlertsHandler provides handlers for alerts API endpoints
type AlertsHandler struct {
	Sys *sys.SystemContext
}

// NewAlertsHandler creates a new AlertsHandler
func NewAlertsHandler(sys *sys.SystemContext) *AlertsHandler {
	return &AlertsHandler{
		Sys: sys,
	}
}

// HandleAlertsAPI handles requests to the /api/alerts endpoint
// It retrieves all alerts from the database and returns them as JSON.
func (h *AlertsHandler) HandleAlertsAPI(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	query := r.URL.Query()

	limit := utils.ParseIntOrDefault(query.Get("limit"), 100)
	page := utils.ParseIntOrDefault(query.Get("page"), 1)

	alertQuery := model.AlertQuery{
		RuleID: strings.TrimSpace(query.Get("rule_id")),
		State:  strings.TrimSpace(query.Get("state")),
		Level:  strings.TrimSpace(query.Get("level")),
		Target: strings.TrimSpace(query.Get("target")),
		Scope:  strings.TrimSpace(query.Get("scope")),
		Sort:   strings.TrimSpace(query.Get("sort")),
		Order:  strings.TrimSpace(query.Get("order")),
		Limit:  limit,
		Offset: (page - 1) * limit,
	}

	// Total count
	total, err := h.Sys.Stores.Alerts.CountAlertsFiltered(ctx, alertQuery)
	if err != nil {
		http.Error(w, "failed to count alerts", http.StatusInternalServerError)
		return
	}

	// Filtered + paged alerts
	alerts, err := h.Sys.Stores.Alerts.ListAlertsFiltered(ctx, alertQuery)
	if err != nil {
		http.Error(w, "failed to query alerts", http.StatusInternalServerError)
		return
	}

	tags := make(map[string]string)
	for _, raw := range query["tag"] {
		parts := strings.SplitN(raw, ":", 2)
		if len(parts) == 2 {
			tags[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}
	alertQuery.Tags = tags

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Total-Count", strconv.Itoa(total))
	if err := json.NewEncoder(w).Encode(alerts); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

// HandleActiveAlertsAPI handles requests to the /api/alerts/active endpoint
func (h *AlertsHandler) HandleActiveAlertsAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	alerts := h.Sys.Tele.Alerts.ListActive()
	if alerts == nil {
		alerts = []model.AlertInstance{}
	}

	if err := json.NewEncoder(w).Encode(alerts); err != nil {
		http.Error(w, "failed to encode alerts", http.StatusInternalServerError)
		return
	}
}

// HandleAlertRulesAPI handles requests to the /api/alerts/rules endpoint
func (h *AlertsHandler) HandleAlertRulesAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	alertRules, err := h.Sys.Stores.Rules.ListRules(r.Context())
	if err != nil {
		http.Error(w, "failed to load alert rules", http.StatusInternalServerError)
		return
	}

	if alertRules == nil {
		alertRules = []model.AlertRule{}
	}

	if err := json.NewEncoder(w).Encode(alertRules); err != nil {
		http.Error(w, "failed to encode alert rules", http.StatusInternalServerError)
		return
	}
}

// HandleAlertsSummaryAPI handles requests to the /api/alerts/summary endpoint
// It returns a summary of alerts grouped by rule_id, showing the latest state and last change time.
func (h *AlertsHandler) HandleAlertsSummaryAPI(w http.ResponseWriter, r *http.Request) {
	alerts, err := h.Sys.Stores.Alerts.ListAlerts(r.Context())
	if err != nil {
		http.Error(w, "failed to fetch alerts", http.StatusInternalServerError)
		return
	}

	// Build summary: latest state per rule_id
	summaryMap := make(map[string]model.AlertSummary)

	for _, a := range alerts {
		existing, exists := summaryMap[a.RuleID]
		if !exists || a.LastFired.After(existing.LastChange) {
			summaryMap[a.RuleID] = model.AlertSummary{
				RuleID:     a.RuleID,
				State:      a.State,
				LastChange: a.LastFired,
			}
		}
	}

	var summaries []model.AlertSummary
	for _, v := range summaryMap {
		summaries = append(summaries, v)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(summaries); err != nil {
		http.Error(w, "failed to encode summary", http.StatusInternalServerError)
		return
	}
}

// HandleCreateAlertRuleAPI handles POST /api/v1/alerts
func (h *AlertsHandler) HandleCreateAlertRuleAPI(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var rule model.AlertRule
	if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
		utils.Error("Failed to decode alert rule: %v", err)
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	// Generate ID if missing
	if strings.TrimSpace(rule.ID) == "" {
		rule.ID = uuid.NewString()
	}

	// Name must be unique
	existing, err := h.Sys.Stores.Rules.GetRuleByName(ctx, rule.Name)
	if err == nil && existing.ID != "" {
		utils.Warn("Duplicate rule name: %s", rule.Name)
		http.Error(w, "rule name already exists", http.StatusConflict)
		return
	}

	if err := h.Sys.Stores.Rules.AddRule(ctx, rule); err != nil {
		utils.Error("Failed to save rule: %v", err)
		http.Error(w, "failed to save rule", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
		"id":     rule.ID,
	}); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

// AlertContextResponse represents the response structure for alert context
type AlertContextResponse struct {
	Alert  model.AlertInstance `json:"alert"`
	Logs   []model.LogEntry    `json:"logs"`
	Events []model.EventEntry  `json:"events"`
}

// HandleAlertContext handles requests to /api/alerts/{id}/context
func (h *AlertsHandler) HandleAlertContext(w http.ResponseWriter, r *http.Request) {
	// Extract the alert ID from the URL parameters
	id := mux.Vars(r)["id"]
	utils.Debug("Fetching context for alert ID: %s", id)

	// Retrieve the alert by ID
	alert, err := h.Sys.Stores.Alerts.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, "alert not found", http.StatusNotFound)
		return
	}

	// Default window duration is 30 minutes if not provided
	windowStr := r.URL.Query().Get("window")
	if windowStr == "" {
		windowStr = "30m"
	}

	// Parse the window duration
	window, err := time.ParseDuration(windowStr)
	if err != nil {
		http.Error(w, "invalid time window", http.StatusBadRequest)
		return
	}

	// Set the start and end time for the alert context window based on when the alert was fired
	start := alert.FirstFired.Add(-window) // Start time is 30 minutes before the first fired time
	end := alert.FirstFired.Add(window)    // End time is 30 minutes after the first fired time

	// Create a log filter based on the alert's context
	filter := model.LogFilter{
		Start:      start,
		End:        end,
		EndpointID: alert.EndpointID,
		Target:     alert.Target,
		Limit:      1000,  // Limit to 1000 logs
		Order:      "asc", // Order logs in ascending order by timestamp
	}

	// Fetch logs based on the filter
	logs, err := h.Sys.Stores.Logs.GetLogs(filter)
	if err != nil {
		http.Error(w, "failed to fetch logs", http.StatusInternalServerError)
		return
	}

	// Create an event filter based on the alert's context
	eventFilter := model.EventFilter{
		Start:      &start,
		End:        &end,
		EndpointID: alert.EndpointID,
		Target:     alert.Target,
		Limit:      1000,  // Limit to 1000 events
		SortOrder:  "asc", // Order events in ascending order by timestamp
	}

	// Fetch events based on the filter
	events, err := h.Sys.Stores.Events.GetRecentEvents(r.Context(), eventFilter)
	if err != nil {
		http.Error(w, "failed to fetch events", http.StatusInternalServerError)
		return
	}

	// Build the response structure
	resp := AlertContextResponse{
		Alert:  alert,  // Include the alert instance in the response
		Logs:   logs,   // Include the fetched logs in the response
		Events: events, // Include the fetched events in the response
	}

	// Send the response as JSON
	utils.JSON(w, 200, resp)
}

// SortBy sorts alerts by the specified field and order
func SortBy(alerts *[]model.AlertInstance, field, order string) {
	slice := *alerts
	field = strings.ToLower(strings.TrimSpace(field))
	order = strings.ToLower(strings.TrimSpace(order))

	less := func(i, j int) bool { return true }

	switch field {
	case "first_fired":
		less = func(i, j int) bool { return slice[i].FirstFired.Before(slice[j].FirstFired) }
	case "last_ok":
		less = func(i, j int) bool { return slice[i].LastOK.Before(slice[j].LastOK) }
	case "rule_id":
		less = func(i, j int) bool { return slice[i].RuleID < slice[j].RuleID }
	case "target":
		less = func(i, j int) bool { return slice[i].Target < slice[j].Target }
	case "scope":
		less = func(i, j int) bool { return slice[i].Scope < slice[j].Scope }
	case "state":
		less = func(i, j int) bool { return slice[i].State < slice[j].State }
	}

	if order == "desc" {
		originalLess := less
		less = func(i, j int) bool { return !originalLess(i, j) }
	}

	sort.Slice(slice, less)
}
