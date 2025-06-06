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
along with GoBright. If not, see https://www.gnu.org/licenses/.
*/

// server/internal/store/logstore/victoriametrics/queries.go

package victorialogstore

import (
	"bufio"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/aaronlmathis/gosight-shared/model"
	"github.com/aaronlmathis/gosight-shared/utils"
)

type logRef struct {
	LogID     string
	Timestamp time.Time
}

type vmExport struct {
	Metric     map[string]string `json:"metric"`
	Timestamps []int64           `json:"timestamps"`
	Values     []float64         `json:"values"`
}

func (v *VictoriaLogStore) queryMatchingLogIDs(filter model.LogFilter) ([]logRef, error) {
	var matchers []string

	add := func(k, v string) {
		if v != "" {
			matchers = append(matchers, fmt.Sprintf(`%s="%s"`, k, v))
		}
	}
	add("endpoint_id", filter.EndpointID)
	add("level", filter.Level)
	add("category", filter.Category)
	add("source", filter.Source)
	add("unit", filter.Unit)
	add("app_name", filter.AppName)
	add("service", filter.Service)
	add("event_id", filter.EventID)
	add("user", filter.User)
	add("container_id", filter.ContainerID)
	add("container_name", filter.ContainerName)
	add("platform", filter.Platform)
	for k, v := range filter.Labels {
		add(k, v)
	}

	query := "gosight.logs.entry"
	if len(matchers) > 0 {
		query += "{" + strings.Join(matchers, ",") + "}"
	}

	var start, end int64

	if filter.Start.IsZero() {
		start = time.Now().Add(-30 * 24 * time.Hour).UnixMilli()
	} else {
		start = filter.Start.UnixMilli()
	}

	if filter.End.IsZero() {
		end = time.Now().UnixMilli()
	} else {
		end = filter.End.UnixMilli()
	}

	reqURL := fmt.Sprintf("%s/api/v1/export?match[]=%s&start=%d&end=%d", v.url, url.QueryEscape(query), start, end)

	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("build request failed: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := v.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("VM query failed: %w", err)
	}
	defer resp.Body.Close()

	var refs []logRef
	scanner := bufio.NewScanner(resp.Body)

	for scanner.Scan() {
		var result vmExport
		line := scanner.Bytes()

		if err := json.Unmarshal(line, &result); err != nil {
			utils.Warn("failed to parse VM export line: %v", err)
			continue
		}

		logID := result.Metric["log_id"]
		if logID == "" {
			utils.Warn("no log_id found in line: %v", result.Metric)
		}
		if len(result.Timestamps) == 0 {
			utils.Warn("no timestamps found for log_id=%s", logID)
		}
		refs = append(refs, logRef{
			LogID:     logID,
			Timestamp: time.UnixMilli(result.Timestamps[0]),
		})
	}

	return refs, nil
}

func (v *VictoriaLogStore) GetLogs(filter model.LogFilter) ([]model.LogEntry, error) {
	refs, err := v.queryMatchingLogIDs(filter)
	if err != nil {
		return nil, err
	}

	// Sort by time before cursor/limit
	sort.Slice(refs, func(i, j int) bool {
		if filter.Order == "asc" {
			return refs[i].Timestamp.Before(refs[j].Timestamp)
		}
		return refs[i].Timestamp.After(refs[j].Timestamp)
	})

	// Apply cursor-based pagination
	if !filter.Cursor.IsZero() {
		// For descending order (newest to oldest), keep logs older than cursor
		// For ascending order (oldest to newest), keep logs newer than cursor
		cursorTime := filter.Cursor
		filteredRefs := make([]logRef, 0, len(refs))
		for _, ref := range refs {
			if filter.Order == "asc" {
				if ref.Timestamp.After(cursorTime) {
					filteredRefs = append(filteredRefs, ref)
				}
			} else {
				if ref.Timestamp.Before(cursorTime) {
					filteredRefs = append(filteredRefs, ref)
				}
			}
		}
		refs = filteredRefs
	}

	// Apply limit
	if filter.Limit > 0 && len(refs) > filter.Limit {
		refs = refs[:filter.Limit]
	}

	var result []model.LogEntry
	for _, ref := range refs {
		entry, err := v.GetLogByID(ref.LogID)
		if err != nil {
			utils.Debug("GetLogs: GetLogByID failed: log_id=%s err=%v", ref.LogID, err)
			continue
		}

		if filter.Contains != "" && !strings.Contains(strings.ToLower(entry.Message), strings.ToLower(filter.Contains)) {
			utils.Debug("GetLogs: skipped due to filter.Contains on log_id=%s", ref.LogID)
			continue
		}

		result = append(result, *entry)
	}

	return result, nil
}

func (v *VictoriaLogStore) GetLogByID(logID string) (*model.LogEntry, error) {

	if entry, ok := v.cache.Get(logID); ok {
		utils.Debug("Log was found in cache: %v", entry.Log.Message)
		if entry.Log.Labels == nil {
			entry.Log.Labels = make(map[string]string)
		}
		entry.Log.Labels["endpoint_id"] = entry.Meta.EndpointID
		entry.Log.Labels["agent_id"] = entry.Meta.AgentID
		entry.Log.Labels["host_id"] = entry.Meta.HostID
		entry.Log.Labels["hostname"] = entry.Meta.Hostname

		for k, v := range entry.Meta.Labels {
			entry.Log.Labels[k] = v
		}
		return &entry.Log, nil
	}

	files, err := filepath.Glob(filepath.Join(v.logsPath, "logs", "*", "*", "*", "*.json.gz"))
	if err != nil {
		return nil, fmt.Errorf("glob failed: %w", err)
	}

	for _, file := range files {
		f, err := os.Open(file)
		if err != nil {
			continue
		}
		gz, err := gzip.NewReader(f)
		if err != nil {
			f.Close()
			continue
		}

		dec := json.NewDecoder(gz)
		for dec.More() {
			var entry model.StoredLog
			if err := dec.Decode(&entry); err != nil {
				break
			}
			if entry.LogID == logID {
				_ = gz.Close()
				_ = f.Close()

				if entry.Log.Labels == nil {
					entry.Log.Labels = make(map[string]string)
				}
				entry.Log.Labels["endpoint_id"] = entry.Meta.EndpointID
				entry.Log.Labels["agent_id"] = entry.Meta.AgentID
				entry.Log.Labels["host_id"] = entry.Meta.HostID
				entry.Log.Labels["hostname"] = entry.Meta.Hostname

				for k, v := range entry.Meta.Labels {
					entry.Log.Labels[k] = v
				}
				return &entry.Log, nil
			}
		}

		_ = gz.Close()
		_ = f.Close()
	}

	return nil, fmt.Errorf("log_id %s not found", logID)
}
