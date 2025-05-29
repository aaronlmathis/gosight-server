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

// Package victorialogs provides a VictoriaLogs implementation of the LogStore interface.
// VictoriaLogs is a fast, cost-effective log management system that provides efficient
// log storage and querying capabilities using LogsQL query language.
//
// This implementation uses VictoriaLogs' native JSON log ingestion endpoint (/insert/jsonl)
// and LogsQL query endpoint (/select/logsql) for optimal performance with log data.
//
// Features:
//   - Native log format storage (not converted to metrics)
//   - LogsQL query language support for powerful log filtering
//   - Structured field and tag handling
//   - Metadata preservation for logs and their sources
//   - Optional log caching integration
//   - Batch log ingestion for improved performance
//
// Configuration example:
//
//	logstore:
//	  engine: "victorialogs"
//	  url: "http://localhost:9428"
package victorialogs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/aaronlmathis/gosight-server/internal/cache"
	"github.com/aaronlmathis/gosight-shared/model"
	"github.com/aaronlmathis/gosight-shared/utils"
)

// VictoriaLogsStore implements the LogStore interface for VictoriaLogs.
// It provides log storage and retrieval functionality using VictoriaLogs'
// native JSON log format and LogsQL query language.
//
// The store handles:
//   - Batch log ingestion via JSONL format
//   - LogsQL-based log querying and filtering
//   - Structured field and tag preservation
//   - Metadata extraction and storage
//   - Optional caching integration
type VictoriaLogsStore struct {
	url    string         // VictoriaLogs server URL
	client *http.Client   // HTTP client for API requests
	cache  cache.LogCache // Optional log cache for performance

}

// NewVictoriaLogsStore creates a new VictoriaLogs store instance.
// It initializes the HTTP client with appropriate timeouts and configures
// the connection to the specified VictoriaLogs server.
//
// Parameters:
//   - url: VictoriaLogs server URL (e.g., "http://localhost:9428")
//   - logCache: Optional cache implementation for improved query performance
//
// Returns:
//   - *VictoriaLogsStore: Configured store instance
//   - error: Any initialization error
//
// Example:
//
//	store, err := NewVictoriaLogsStore("http://localhost:9428", myCache)
//	if err != nil {
//	    log.Fatal("Failed to create VictoriaLogs store:", err)
//	}
func NewVictoriaLogsStore(url string, logCache cache.LogCache) (*VictoriaLogsStore, error) {
	return &VictoriaLogsStore{
		url:    strings.TrimSuffix(url, "/"),
		client: &http.Client{Timeout: 30 * time.Second},
		cache:  logCache,
	}, nil
}

// Name returns the display name of the log store implementation.
// This is used for identification and logging purposes.
func (v *VictoriaLogsStore) Name() string {
	return "VictoriaLogs Store"
}

// Write stores a batch of log payloads in VictoriaLogs.
// It converts log entries to VictoriaLogs' JSONL format and sends them
// to the /insert/jsonl endpoint. The method also updates the cache if available.
//
// The conversion process:
//   - Maps log fields to VictoriaLogs field format
//   - Preserves metadata as searchable fields
//   - Handles structured fields with prefixes (field_, tag_, meta_)
//   - Maintains timestamp precision using RFC3339Nano format
//
// Parameters:
//   - batch: Slice of LogPayload containing logs and their metadata
//
// Returns:
//   - error: Any write operation error
//
// Example log entry conversion:
//
//	LogEntry{
//	    Timestamp: time.Now(),
//	    Message: "User login failed",
//	    Level: "error",
//	    Source: "auth-service",
//	    Fields: {"user_id": "123", "ip": "192.168.1.1"},
//	    Labels: {"env": "prod", "service": "auth"}
//	}
//
//	Becomes VictoriaLogs entry:
//	{
//	    "_time": "2025-01-01T12:00:00.000000000Z",
//	    "_msg": "User login failed",
//	    "level": "error",
//	    "source": "auth-service",
//	    "field_user_id": "123",
//	    "field_ip": "192.168.1.1",
//	    "tag_env": "prod",
//	    "tag_service": "auth"
//	}
func (v *VictoriaLogsStore) Write(batch []model.LogPayload) error {
	if len(batch) == 0 {
		return nil
	}

	// Convert to VictoriaLogs JSON format
	var logEntries []map[string]interface{}

	for _, payload := range batch {
		for _, logEntry := range payload.Logs {
			// Create VictoriaLogs entry
			entry := map[string]interface{}{
				"_time":  logEntry.Timestamp.Format(time.RFC3339Nano),
				"_msg":   logEntry.Message,
				"level":  logEntry.Level,
				"source": logEntry.Source,
			}

			// Add optional fields
			if logEntry.Category != "" {
				entry["category"] = logEntry.Category
			}
			if logEntry.PID != 0 {
				entry["pid"] = strconv.Itoa(logEntry.PID)
			}

			// Add meta information
			if payload.Meta != nil {
				entry["endpoint_id"] = payload.Meta.EndpointID
				entry["agent_id"] = payload.Meta.AgentID
				entry["host_id"] = payload.Meta.HostID
				entry["hostname"] = payload.Meta.Hostname

				// Add meta labels
				for k, v := range payload.Meta.Labels {
					entry[k] = v
				}
			}

			// Add log-specific meta
			if logEntry.Meta != nil {
				if logEntry.Meta.Platform != "" {
					entry["platform"] = logEntry.Meta.Platform
				}
				if logEntry.Meta.AppName != "" {
					entry["app_name"] = logEntry.Meta.AppName
				}
				if logEntry.Meta.AppVersion != "" {
					entry["app_version"] = logEntry.Meta.AppVersion
				}
				if logEntry.Meta.ContainerID != "" {
					entry["container_id"] = logEntry.Meta.ContainerID
				}
				if logEntry.Meta.ContainerName != "" {
					entry["container_name"] = logEntry.Meta.ContainerName
				}
				if logEntry.Meta.Unit != "" {
					entry["unit"] = logEntry.Meta.Unit
				}
				if logEntry.Meta.Service != "" {
					entry["service"] = logEntry.Meta.Service
				}
				if logEntry.Meta.EventID != "" {
					entry["event_id"] = logEntry.Meta.EventID
				}
				if logEntry.Meta.User != "" {
					entry["user"] = logEntry.Meta.User
				}
				if logEntry.Meta.Executable != "" {
					entry["executable"] = logEntry.Meta.Executable
				}
				if logEntry.Meta.Path != "" {
					entry["path"] = logEntry.Meta.Path
				}

				// Add extra meta fields
				for k, v := range logEntry.Meta.Extra {
					entry["meta_"+k] = v
				}
			}

			// Add structured fields
			for k, v := range logEntry.Fields {
				entry["field_"+k] = v
			}

			// Add labels
			for k, v := range logEntry.Labels {
				entry["tag_"+k] = v
			}

			logEntries = append(logEntries, entry)
		}
	}

	// Cache entries if cache is available
	if v.cache != nil {
		// Convert to StoredLog format for caching
		var storedLogs []*model.StoredLog
		for _, payload := range batch {
			for _, logEntry := range payload.Logs {
				logID := generateLogID(logEntry)
				storedLogs = append(storedLogs, &model.StoredLog{
					LogID: logID,
					Log:   logEntry,
					Meta:  payload.Meta,
				})
			}
		}
		v.cache.Add(storedLogs)
	}

	// Send to VictoriaLogs
	return v.sendToVictoriaLogs(logEntries)
}

// sendToVictoriaLogs sends log entries to VictoriaLogs OSS using NDJSON format.
//
// This function marshals each log entry as a single line of JSON (newline-delimited)
// and sends the batch to the /insert endpoint, as required by VictoriaLogs OSS (2024+).
// Note: OSS VictoriaLogs does not support named tables—there is a single default log store.
//
// Parameters:
//   - entries: slice of log entry maps to send to VictoriaLogs.
//
// Returns:
//   - error: Any HTTP or VictoriaLogs API error encountered.
//
// Example NDJSON sent:
//
//	{"timestamp":1716943056,"level":"info","message":"starting up"}
//	{"timestamp":1716943057,"level":"error","message":"something broke"}
func (v *VictoriaLogsStore) sendToVictoriaLogs(entries []map[string]interface{}) error {
	// Convert entries to NDJSON (newline-delimited JSON)
	var buf bytes.Buffer
	for _, entry := range entries {
		entryBytes, err := json.Marshal(entry)
		if err != nil {
			utils.Warn("Failed to marshal log entry: %v", err)
			continue // skip this entry, continue with others
		}
		buf.Write(entryBytes)
		buf.WriteByte('\n')
	}

	// Send to VictoriaLogs /insert/jsonline endpoint (no table name)
	req, err := http.NewRequest("POST", v.url+"/insert/jsonline", &buf)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-ndjson")

	resp, err := v.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send logs to VictoriaLogs: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("VictoriaLogs responded with status %d: %s", resp.StatusCode, string(body))
	}

	utils.Debug("Successfully sent %d log entries to VictoriaLogs", len(entries))
	return nil
}

// Close cleans up any resources used by the store.
// Currently, this implementation doesn't require cleanup, but the method
// is provided for interface compliance and future extensibility.
//
// Returns:
//   - error: Always returns nil for this implementation
func (v *VictoriaLogsStore) Close() error {
	// No resources to clean up
	return nil
}

// generateLogID creates a unique identifier for a log entry.
// The ID is generated using the timestamp and a portion of the message
// to ensure uniqueness while maintaining deterministic behavior for
// identical log entries.
//
// Parameters:
//   - entry: The log entry to generate an ID for
//
// Returns:
//   - string: Unique identifier for the log entry
func generateLogID(entry model.LogEntry) string {
	return fmt.Sprintf("%d_%s", entry.Timestamp.UnixNano(), entry.Message[:min(20, len(entry.Message))])
}

// min returns the smaller of two integers.
// This utility function is used to safely truncate strings for ID generation.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
