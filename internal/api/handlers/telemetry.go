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

	"github.com/aaronlmathis/gosight-server/internal/sys"
	"github.com/aaronlmathis/gosight-shared/model"
	"github.com/aaronlmathis/gosight-shared/utils"
)

// TelemetryHandler handles telemetry data ingestion endpoints
type TelemetryHandler struct {
	Sys *sys.SystemContext
}

// NewTelemetryHandler creates a new TelemetryHandler
func NewTelemetryHandler(sys *sys.SystemContext) *TelemetryHandler {
	return &TelemetryHandler{
		Sys: sys,
	}
}

// HandleMetrics handles POST /telemetry/metrics
func (h *TelemetryHandler) HandleMetrics(w http.ResponseWriter, r *http.Request) {
	var payload model.MetricPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		utils.Error("Failed to decode metric payload: %v", err)
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	// Store metrics using the metric store
	if err := h.Sys.Buffers.Metrics.WriteAny([]model.MetricPayload{payload}); err != nil {
		utils.Error("Failed to store metrics: %v", err)
		http.Error(w, "Failed to store metrics", http.StatusInternalServerError)
		return
	}

	// Update metric index for search capabilities
	if payload.Meta != nil {
		for _, metric := range payload.Metrics {
			h.Sys.Tele.Index.Add(metric.Namespace, metric.SubNamespace, metric.Name, metric.Dimensions)
		}
	}

	w.WriteHeader(http.StatusAccepted)
}

// HandleLogs handles POST /telemetry/logs
func (h *TelemetryHandler) HandleLogs(w http.ResponseWriter, r *http.Request) {
	var payload model.LogPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		utils.Error("Failed to decode log payload: %v", err)
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	// Process resource discovery first
	h.Sys.Tele.ResourceDiscovery.ProcessLogPayload(&payload)

	// Store logs
	if err := h.Sys.Buffers.Logs.WriteAny(&payload); err != nil {
		utils.Error("Failed to store logs: %v", err)
		http.Error(w, "Failed to store logs", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

// HandleTraces handles POST /telemetry/traces
func (h *TelemetryHandler) HandleTraces(w http.ResponseWriter, r *http.Request) {
	var payload model.TracePayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		utils.Error("Failed to decode trace payload: %v", err)
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	// Process resource discovery first
	h.Sys.Tele.ResourceDiscovery.ProcessTracePayload(&payload)

	// Store traces (if trace storage is implemented)
	// if err := h.Sys.Stores.Traces.StoreTraces(r.Context(), &payload); err != nil {
	//     utils.Error("Failed to store traces: %v", err)
	//     http.Error(w, "Failed to store traces", http.StatusInternalServerError)
	//     return
	// }

	// For now, just accept traces without storing
	w.WriteHeader(http.StatusAccepted)
}
