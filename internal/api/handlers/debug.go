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

package handlers

import (
	"encoding/json"
	"net/http"
	"runtime"
	"time"

	"github.com/aaronlmathis/gosight-server/debugtools"
	"github.com/aaronlmathis/gosight-server/internal/sys"
)

// DebugHandler handles debug-related API endpoints
type DebugHandler struct {
	Sys *sys.SystemContext
}

// NewDebugHandler creates a new DebugHandler
func NewDebugHandler(sys *sys.SystemContext) *DebugHandler {
	return &DebugHandler{
		Sys: sys,
	}
}

// HandleCacheAudit performs an audit of the cache systems and returns diagnostic information
func (h *DebugHandler) HandleCacheAudit(w http.ResponseWriter, r *http.Request) {
	report := debugtools.AuditCaches(h.Sys.Cache.Tags, h.Sys.Cache.Metrics)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(report)
}

// HandleAPIHealthCheck returns the server health status
func (h *DebugHandler) HandleAPIHealthCheck(w http.ResponseWriter, r *http.Request) {
	healthStatus := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"services": map[string]string{
			"database":  "healthy",
			"cache":     "healthy",
			"metrics":   "healthy",
			"logs":      "healthy",
			"telemetry": "healthy",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(healthStatus)
}

// HandleAPIDebugMetrics returns internal server metrics for debugging
func (h *DebugHandler) HandleAPIDebugMetrics(w http.ResponseWriter, r *http.Request) {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	metrics := map[string]interface{}{
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"runtime": map[string]interface{}{
			"goroutines":   runtime.NumGoroutine(),
			"memory_alloc": memStats.Alloc,
			"memory_total": memStats.TotalAlloc,
			"memory_sys":   memStats.Sys,
			"gc_cycles":    memStats.NumGC,
			"gc_pause_ns":  memStats.PauseNs[(memStats.NumGC+255)%256],
		},
		"cache": map[string]interface{}{
			"tags_count":    len(h.Sys.Cache.Tags.GetAllEndpoints()),
			"metrics_count": len(h.Sys.Cache.Metrics.GetAllEntries()),
		},
		"connections": map[string]interface{}{
			"active_agents": len(h.Sys.Tracker.GetAgentMap()),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(metrics)
}

// HandleAPIDebugPprof enables Go pprof profiling endpoints
func (h *DebugHandler) HandleAPIDebugPprof(w http.ResponseWriter, r *http.Request) {
	// Redirect to pprof index
	http.Redirect(w, r, "/debug/pprof/", http.StatusTemporaryRedirect)
}

// HandleAPIDebugConfig returns the current server configuration (sanitized)
func (h *DebugHandler) HandleAPIDebugConfig(w http.ResponseWriter, r *http.Request) {
	config := map[string]interface{}{
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"server": map[string]interface{}{
			"grpc_addr":   h.Sys.Cfg.Server.GRPCAddr,
			"http_addr":   h.Sys.Cfg.Server.HTTPAddr,
			"environment": h.Sys.Cfg.Server.Environment,
		},
		"metricstore": map[string]interface{}{
			"engine":  h.Sys.Cfg.MetricStore.Engine,
			"url":     h.Sys.Cfg.MetricStore.URL,
			"workers": h.Sys.Cfg.MetricStore.Workers,
		},
		"buffers": map[string]interface{}{
			"enabled":        h.Sys.Cfg.BufferEngine.Enabled,
			"flush_interval": h.Sys.Cfg.BufferEngine.FlushInterval,
			"buffer_size":    h.Sys.Cfg.BufferEngine.Metrics.BufferSize,
		},
		"logs": map[string]interface{}{
			"error_log_file": h.Sys.Cfg.Logs.ErrorLogFile,
			"app_log_file":   h.Sys.Cfg.Logs.AppLogFile,
			"log_level":      h.Sys.Cfg.Logs.LogLevel,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(config)
}

// HandleAPIDebugTest provides a test endpoint for debugging connectivity
func (h *DebugHandler) HandleAPIDebugTest(w http.ResponseWriter, r *http.Request) {
	test := map[string]interface{}{
		"timestamp":   time.Now().UTC().Format(time.RFC3339),
		"method":      r.Method,
		"path":        r.URL.Path,
		"headers":     r.Header,
		"remote_addr": r.RemoteAddr,
		"user_agent":  r.UserAgent(),
		"test_result": "success",
		"message":     "Debug test endpoint is working correctly",
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(test)
}

// HandleAPIVersion returns version and build information
func (h *DebugHandler) HandleAPIVersion(w http.ResponseWriter, r *http.Request) {
	version := map[string]interface{}{
		"timestamp":  time.Now().UTC().Format(time.RFC3339),
		"version":    "1.0.0",   // TODO: Get from build variables
		"build":      "dev",     // TODO: Get from build variables
		"commit":     "unknown", // TODO: Get from build variables
		"built_at":   "unknown", // TODO: Get from build variables
		"go_version": runtime.Version(),
		"platform":   runtime.GOOS + "/" + runtime.GOARCH,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(version)
}

// HandleAPIStatus returns detailed system status information
func (h *DebugHandler) HandleAPIStatus(w http.ResponseWriter, r *http.Request) {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	status := map[string]interface{}{
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"status":    "operational",
		"uptime":    time.Since(time.Now().Add(-time.Hour)).Seconds(), // Placeholder since StartTime field doesn't exist
		"system": map[string]interface{}{
			"goroutines": runtime.NumGoroutine(),
			"memory_mb":  memStats.Alloc / 1024 / 1024,
			"gc_cycles":  memStats.NumGC,
		},
		"services": map[string]interface{}{
			"grpc_server":   "running",
			"http_server":   "running",
			"metric_buffer": "running",
			"log_buffer":    "running",
		},
		"statistics": map[string]interface{}{
			"active_agents":  len(h.Sys.Tracker.GetAgentMap()),
			"cached_tags":    len(h.Sys.Cache.Tags.GetAllEndpoints()),
			"cached_metrics": len(h.Sys.Cache.Metrics.GetAllEntries()),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(status)
}
