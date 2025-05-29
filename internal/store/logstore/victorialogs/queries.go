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

package victorialogs

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/aaronlmathis/gosight-shared/model"
	"github.com/aaronlmathis/gosight-shared/utils"
)

// vlResponse represents the JSON response structure from VictoriaLogs.
// VictoriaLogs returns query results in this format when using the
// /select/logsql endpoint.
type vlResponse struct {
	Data []vlLogEntry `json:"data"` // Array of log entries matching the query
}

// vlLogEntry represents a single log entry in VictoriaLogs response format.
// The structure uses inline field mapping to capture all log fields
// as a flat map alongside the standard _time and _msg fields.
type vlLogEntry struct {
	Time   string            `json:"_time"`   // RFC3339Nano timestamp
	Msg    string            `json:"_msg"`    // Original log message
	Fields map[string]string `json:",inline"` // All other fields as key-value pairs
}

// GetLogs retrieves log entries from VictoriaLogs based on the provided filter.
// It constructs a LogsQL query from the filter parameters, executes it against
// VictoriaLogs, and converts the results back to the GoSight log format.
//
// Query Process:
//  1. Converts LogFilter to LogsQL query syntax
//  2. Sets time range and pagination parameters
//  3. Executes query against VictoriaLogs /select/logsql endpoint
//  4. Parses response and converts to model.LogEntry format
//  5. Applies additional client-side filtering if needed
//  6. Handles cursor-based pagination and sorting
//
// Parameters:
//   - filter: LogFilter specifying query criteria, time range, and pagination
//
// Returns:
//   - []model.LogEntry: Slice of log entries matching the filter criteria
//   - error: Any query execution or parsing error
//
// Example usage:
//
//	filter := model.LogFilter{
//	    Level: "error",
//	    Start: time.Now().Add(-1 * time.Hour),
//	    End: time.Now(),
//	    Limit: 100,
//	    Tags: map[string]string{"service": "auth"},
//	    Contains: "login failed",
//	}
//	logs, err := store.GetLogs(filter)
func (v *VictoriaLogsStore) GetLogs(filter model.LogFilter) ([]model.LogEntry, error) {
	// Build LogsQL query
	query := v.buildLogsQLQuery(filter)

	// Set time range
	start := filter.Start
	end := filter.End
	if start.IsZero() {
		start = time.Now().Add(-24 * time.Hour) // Default to last 24 hours
	}
	if end.IsZero() {
		end = time.Now()
	}

	// Prepare query parameters
	params := url.Values{}
	params.Set("query", query)
	params.Set("start", start.Format(time.RFC3339))
	params.Set("end", end.Format(time.RFC3339))

	if filter.Limit > 0 {
		params.Set("limit", strconv.Itoa(filter.Limit))
	} else {
		params.Set("limit", "1000") // Default limit
	}

	// Build request URL
	reqURL := fmt.Sprintf("%s/select/logsql?%s", v.url, params.Encode())

	utils.Debug("VictoriaLogs query: %s", reqURL)

	// Make request
	resp, err := v.client.Get(reqURL)
	if err != nil {
		return nil, fmt.Errorf("failed to query VictoriaLogs: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("VictoriaLogs query failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var vlResp vlResponse
	if err := json.NewDecoder(resp.Body).Decode(&vlResp); err != nil {
		return nil, fmt.Errorf("failed to decode VictoriaLogs response: %w", err)
	}

	// Convert VictoriaLogs entries to model.LogEntry
	var result []model.LogEntry
	for _, vlEntry := range vlResp.Data {
		logEntry, err := v.convertVLEntryToLogEntry(vlEntry)
		if err != nil {
			utils.Warn("Failed to convert VictoriaLogs entry: %v", err)
			continue
		}

		// Apply additional client-side filtering if needed
		if v.matchesFilter(logEntry, filter) {
			result = append(result, *logEntry)
		}
	}

	// Apply cursor-based pagination if needed
	if !filter.Cursor.IsZero() {
		result = v.applyCursorFilter(result, filter)
	}

	// Sort results
	if filter.Order == "asc" {
		// VictoriaLogs returns newest first by default, so reverse for asc
		for i := len(result)/2 - 1; i >= 0; i-- {
			opp := len(result) - 1 - i
			result[i], result[opp] = result[opp], result[i]
		}
	}

	return result, nil
}

// buildLogsQLQuery constructs a LogsQL query string from the provided filter.
// LogsQL is VictoriaLogs' query language, similar to PromQL but designed for logs.
//
// Query Construction Rules:
//   - Field filters use exact matching: field_name:"value"
//   - Tag filters are prefixed: tag_name:"value"
//   - Meta filters are prefixed: meta_name:"value"
//   - Text search uses wildcard matching: _msg:*search_term*
//   - Multiple conditions are joined with AND
//   - Empty filter returns "*" (match all)
//
// Parameters:
//   - filter: LogFilter containing search criteria
//
// Returns:
//   - string: LogsQL query string
//
// Example queries generated:
//   - level:"error" AND service:"auth" AND _msg:*login*
//   - tag_env:"prod" AND field_user_id:"123"
//   - endpoint_id:"endpoint-1" AND meta_datacenter:"us-east"
func (v *VictoriaLogsStore) buildLogsQLQuery(filter model.LogFilter) string {
	var conditions []string

	// Add basic filters
	if filter.Level != "" {
		conditions = append(conditions, fmt.Sprintf("level:%q", filter.Level))
	}
	if filter.Source != "" {
		conditions = append(conditions, fmt.Sprintf("source:%q", filter.Source))
	}
	if filter.Category != "" {
		conditions = append(conditions, fmt.Sprintf("category:%q", filter.Category))
	}
	if filter.EndpointID != "" {
		conditions = append(conditions, fmt.Sprintf("endpoint_id:%q", filter.EndpointID))
	}
	if filter.Unit != "" {
		conditions = append(conditions, fmt.Sprintf("unit:%q", filter.Unit))
	}
	if filter.AppName != "" {
		conditions = append(conditions, fmt.Sprintf("app_name:%q", filter.AppName))
	}
	if filter.Service != "" {
		conditions = append(conditions, fmt.Sprintf("service:%q", filter.Service))
	}
	if filter.EventID != "" {
		conditions = append(conditions, fmt.Sprintf("event_id:%q", filter.EventID))
	}
	if filter.User != "" {
		conditions = append(conditions, fmt.Sprintf("user:%q", filter.User))
	}
	if filter.ContainerID != "" {
		conditions = append(conditions, fmt.Sprintf("container_id:%q", filter.ContainerID))
	}
	if filter.ContainerName != "" {
		conditions = append(conditions, fmt.Sprintf("container_name:%q", filter.ContainerName))
	}
	if filter.Platform != "" {
		conditions = append(conditions, fmt.Sprintf("platform:%q", filter.Platform))
	}

	// Add tag filters
	for k, v := range filter.Tags {
		conditions = append(conditions, fmt.Sprintf("tag_%s:%q", k, v))
	}

	// Add field filters
	for k, v := range filter.Fields {
		conditions = append(conditions, fmt.Sprintf("field_%s:%q", k, v))
	}

	// Add meta filters
	for k, v := range filter.Meta {
		conditions = append(conditions, fmt.Sprintf("meta_%s:%q", k, v))
	}

	// Add text search
	if filter.Contains != "" {
		conditions = append(conditions, fmt.Sprintf("_msg:*%s*", filter.Contains))
	}

	// Build final query
	query := "*"
	if len(conditions) > 0 {
		query = strings.Join(conditions, " AND ")
	}

	return query
}

// convertVLEntryToLogEntry converts a VictoriaLogs response entry to GoSight's LogEntry format.
// This function handles the reverse mapping from VictoriaLogs' flat field structure
// back to GoSight's structured log entry with separate fields, tags, and metadata.
//
// Conversion Process:
//  1. Parse timestamp from _time field
//  2. Extract message from _msg field
//  3. Map standard fields (level, source, category, etc.)
//  4. Extract prefixed fields (field_, tag_, meta_)
//  5. Reconstruct LogMeta structure from individual fields
//  6. Handle type conversions (e.g., string to int for PID)
//
// Parameters:
//   - vlEntry: VictoriaLogs log entry from query response
//
// Returns:
//   - *model.LogEntry: Converted log entry in GoSight format
//   - error: Any conversion or parsing error
func (v *VictoriaLogsStore) convertVLEntryToLogEntry(vlEntry vlLogEntry) (*model.LogEntry, error) {
	// Parse timestamp
	timestamp, err := time.Parse(time.RFC3339Nano, vlEntry.Time)
	if err != nil {
		// Try alternative formats
		if timestamp, err = time.Parse(time.RFC3339, vlEntry.Time); err != nil {
			return nil, fmt.Errorf("failed to parse timestamp: %w", err)
		}
	}

	// Create log entry
	logEntry := &model.LogEntry{
		Timestamp: timestamp,
		Message:   vlEntry.Msg,
		Fields:    make(map[string]string),
		Tags:      make(map[string]string),
	}

	// Parse fields from VictoriaLogs response
	for k, v := range vlEntry.Fields {
		switch k {
		case "_time", "_msg":
			// Skip these as they're already processed
			continue
		case "level":
			logEntry.Level = v
		case "source":
			logEntry.Source = v
		case "category":
			logEntry.Category = v
		case "pid":
			if pid, err := strconv.Atoi(v); err == nil {
				logEntry.PID = pid
			}
		default:
			// Handle prefixed fields
			if strings.HasPrefix(k, "field_") {
				logEntry.Fields[strings.TrimPrefix(k, "field_")] = v
			} else if strings.HasPrefix(k, "tag_") {
				logEntry.Tags[strings.TrimPrefix(k, "tag_")] = v
			} else if strings.HasPrefix(k, "meta_") {
				// Initialize Meta if needed
				if logEntry.Meta == nil {
					logEntry.Meta = &model.LogMeta{
						Extra: make(map[string]string),
					}
				}
				logEntry.Meta.Extra[strings.TrimPrefix(k, "meta_")] = v
			} else {
				// Handle other meta fields
				if logEntry.Meta == nil {
					logEntry.Meta = &model.LogMeta{
						Extra: make(map[string]string),
					}
				}
				switch k {
				case "platform":
					logEntry.Meta.Platform = v
				case "app_name":
					logEntry.Meta.AppName = v
				case "app_version":
					logEntry.Meta.AppVersion = v
				case "container_id":
					logEntry.Meta.ContainerID = v
				case "container_name":
					logEntry.Meta.ContainerName = v
				case "unit":
					logEntry.Meta.Unit = v
				case "service":
					logEntry.Meta.Service = v
				case "event_id":
					logEntry.Meta.EventID = v
				case "user":
					logEntry.Meta.User = v
				case "executable":
					logEntry.Meta.Executable = v
				case "path":
					logEntry.Meta.Path = v
				default:
					// Add as tag
					logEntry.Tags[k] = v
				}
			}
		}
	}

	return logEntry, nil
}

// matchesFilter performs additional client-side filtering for complex conditions
// that cannot be easily expressed in LogsQL or need case-insensitive matching.
//
// Currently handles:
//   - Case-insensitive text search in log messages
//   - Additional validation of filter criteria
//
// Parameters:
//   - entry: Log entry to check against filter
//   - filter: Filter criteria to apply
//
// Returns:
//   - bool: True if entry matches filter criteria
func (v *VictoriaLogsStore) matchesFilter(entry *model.LogEntry, filter model.LogFilter) bool {
	// Additional client-side filtering for complex conditions
	if filter.Contains != "" && !strings.Contains(strings.ToLower(entry.Message), strings.ToLower(filter.Contains)) {
		return false
	}

	return true
}

// applyCursorFilter applies cursor-based pagination to the result set.
// This enables efficient pagination through large log datasets by using
// timestamp-based cursors rather than offset-based pagination.
//
// Cursor Logic:
//   - For ascending order: return entries after cursor timestamp
//   - For descending order: return entries before cursor timestamp
//   - Cursor timestamp is exclusive (not included in results)
//
// Parameters:
//   - entries: Slice of log entries to filter
//   - filter: Filter containing cursor and order information
//
// Returns:
//   - []model.LogEntry: Filtered entries based on cursor position
func (v *VictoriaLogsStore) applyCursorFilter(entries []model.LogEntry, filter model.LogFilter) []model.LogEntry {
	if filter.Cursor.IsZero() {
		return entries
	}

	var filtered []model.LogEntry
	for _, entry := range entries {
		if filter.Order == "asc" {
			if entry.Timestamp.After(filter.Cursor) {
				filtered = append(filtered, entry)
			}
		} else {
			if entry.Timestamp.Before(filter.Cursor) {
				filtered = append(filtered, entry)
			}
		}
	}

	return filtered
}
