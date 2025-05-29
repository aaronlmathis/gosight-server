// File: server/internal/api/handlers/logs.go
// Description: This file contains the logs handlers for the GoSight server.

package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/aaronlmathis/gosight-server/internal/sys"
	"github.com/aaronlmathis/gosight-shared/model"
	"github.com/aaronlmathis/gosight-shared/utils"
)

// LogsHandler provides handlers for logs API endpoints
type LogsHandler struct {
	Sys *sys.SystemContext
}

// NewLogsHandler creates a new LogsHandler
func NewLogsHandler(sys *sys.SystemContext) *LogsHandler {
	return &LogsHandler{
		Sys: sys,
	}
}

// LogQueryParams represents the query parameters for log requests
type LogQueryParams struct {
	Level     string
	Source    string
	Host      string
	Container string
	Contains  string
	Start     *time.Time
	End       *time.Time
	Limit     int
	Cursor    string
}

// LogResponse represents the response structure for log API calls
type LogResponse struct {
	Logs       []model.LogEntry `json:"logs"`
	NextCursor string           `json:"next_cursor,omitempty"`
	HasMore    bool             `json:"has_more"`
	Count      int              `json:"count"`
}

// parseLogQueryParams parses query parameters into LogQueryParams
func (h *LogsHandler) parseLogQueryParams(r *http.Request) LogQueryParams {
	q := r.URL.Query()

	params := LogQueryParams{
		Level:     q.Get("level"),
		Source:    q.Get("source"),
		Host:      q.Get("host"),
		Container: q.Get("container"),
		Contains:  q.Get("contains"),
		Cursor:    q.Get("cursor"),
		Limit:     100, // default limit
	}

	if limitStr := q.Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 && limit <= 1000 {
			params.Limit = limit
		}
	}

	if startStr := q.Get("start"); startStr != "" {
		if start, err := time.Parse(time.RFC3339, startStr); err == nil {
			params.Start = &start
		}
	}

	if endStr := q.Get("end"); endStr != "" {
		if end, err := time.Parse(time.RFC3339, endStr); err == nil {
			params.End = &end
		}
	}

	return params
}

// HandleLogAPI handles the HTTP request for the log API.
// It retrieves the logs from the log store, applies any filters specified
// in the query parameters, and returns the logs as a JSON response.
// The function uses the LogQueryParams struct to parse the query parameters
// and filter the logs. It handles errors and returns appropriate HTTP status
// codes and messages. It also supports pagination using cursor-based pagination.
func (h *LogsHandler) HandleLogAPI(w http.ResponseWriter, r *http.Request) {
	filter := parseLogFilterFromQuery(r)

	// Ensure a sane default and request one extra record to determine if there are more
	if filter.Limit <= 0 || filter.Limit > 1000 {
		filter.Limit = 100
	}
	originalLimit := filter.Limit
	filter.Limit = originalLimit + 1 // Request one extra to determine if there are more

	logs, err := h.Sys.Stores.Logs.GetLogs(filter)
	if err != nil {
		utils.Error("log query failed: %v", err)
		http.Error(w, "log query failed", http.StatusInternalServerError)
		return
	}

	// Pagination logic (cursor-based)
	hasMore := len(logs) > originalLimit
	var nextCursor string

	if hasMore {
		// Remove the extra record we requested
		logs = logs[:originalLimit]
		// Use the timestamp of the last visible log as the next cursor
		nextCursor = logs[originalLimit-1].Timestamp.Format(time.RFC3339Nano)
	}

	resp := LogResponse{
		Logs:       logs,
		NextCursor: nextCursor,
		HasMore:    hasMore,
		Count:      len(logs),
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

// parseLogFilterFromQuery parses the query parameters from the HTTP request
// into a model.LogFilter struct. It extracts parameters like start time,
// end time, cursor, limit, order, endpoint ID, target, level, category,
// source, contains string, unit, app name, service, event ID, user,
// container ID, container name, platform, Label, fields, and meta.
func parseLogFilterFromQuery(r *http.Request) model.LogFilter {
	q := r.URL.Query()

	parseTime := func(key string) time.Time {
		str := q.Get(key)
		if str == "" {
			return time.Time{}
		}
		t, err := time.Parse(time.RFC3339, str)
		if err != nil {
			return time.Time{}
		}
		return t
	}
	parseInt := func(key string, def int) int {
		val := q.Get(key)
		if val == "" {
			return def
		}
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
		return def
	}

	// Map extract helper
	extractPrefixed := func(prefix string) map[string]string {
		out := make(map[string]string)
		for key, values := range q {
			if strings.HasPrefix(key, prefix) && len(values) > 0 {
				k := strings.TrimPrefix(key, prefix)
				out[k] = values[0]
			}
		}
		return out
	}

	filter := model.LogFilter{
		Start:         parseTime("start"),
		End:           parseTime("end"),
		Cursor:        parseTime("cursor"),
		Limit:         parseInt("limit", 100),
		Order:         q.Get("order"),
		EndpointID:    q.Get("endpoint_id"),
		Target:        q.Get("target"),
		Level:         q.Get("level"),
		Category:      q.Get("category"),
		Source:        q.Get("source"),
		Contains:      q.Get("contains"),
		Unit:          q.Get("unit"),
		AppName:       q.Get("app_name"),
		Service:       q.Get("service"),
		EventID:       q.Get("event_id"),
		User:          q.Get("user"),
		ContainerID:   q.Get("container_id"),
		ContainerName: q.Get("container_name"),
		Platform:      q.Get("platform"),

		Labels: extractPrefixed("label_"),
		Fields: extractPrefixed("field_"),
		Meta:   extractPrefixed("meta_"),
	}

	return filter
}

// matchesSearch checks if a log entry matches the search keyword.
// It checks if the keyword is present in the message, source, category,
// or any of the meta fields, Logs, or fields of the log entry.
// The function is case-insensitive and uses strings.Contains to check for matches.
// It returns true if the log entry matches the search keyword, false otherwise.
func matchesSearch(log model.LogEntry, keyword string) bool {
	kw := strings.ToLower(keyword)

	// Base fields
	if strings.Contains(strings.ToLower(log.Message), kw) ||
		strings.Contains(strings.ToLower(log.Source), kw) ||
		strings.Contains(strings.ToLower(log.Category), kw) ||
		strings.Contains(strings.ToLower(log.Level), kw) {
		return true
	}

	// Log.Meta fields
	if log.Meta != nil {
		metaFields := []string{
			log.Meta.Platform,
			log.Meta.AppName,
			log.Meta.AppVersion,
			log.Meta.ContainerID,
			log.Meta.ContainerName,
			log.Meta.Unit,
			log.Meta.Service,
			log.Meta.EventID,
			log.Meta.User,
			log.Meta.Executable,
			log.Meta.Path,
		}
		for _, f := range metaFields {
			if strings.Contains(strings.ToLower(f), kw) {
				return true
			}
		}

		for k, v := range log.Meta.Extra {
			if strings.Contains(strings.ToLower(k), kw) || strings.Contains(strings.ToLower(v), kw) {
				return true
			}
		}
	}

	// Logs and Fields
	for k, v := range log.Labels {
		if strings.Contains(strings.ToLower(k), kw) || strings.Contains(strings.ToLower(v), kw) {
			return true
		}
	}
	for k, v := range log.Fields {
		if strings.Contains(strings.ToLower(k), kw) || strings.Contains(strings.ToLower(v), kw) {
			return true
		}
	}

	return false
}
