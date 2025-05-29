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

package telemetry

import (
	"context"
	"time"

	"github.com/aaronlmathis/gosight-server/internal/events"
	"github.com/aaronlmathis/gosight-server/internal/sys"
	"github.com/aaronlmathis/gosight-shared/model"
	"github.com/aaronlmathis/gosight-shared/utils"
	"github.com/google/uuid"
	collogpb "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type LogsHandler struct {
	Sys *sys.SystemContext
	collogpb.UnimplementedLogsServiceServer
}

func NewLogsHandler(sys *sys.SystemContext) *LogsHandler {
	utils.Debug("LogsHandler initialized with store: %T", sys.Stores.Logs)
	return &LogsHandler{
		Sys: sys,
	}
}

// Export implements the OTLP LogsService Export method (unary, not streaming)
func (h *LogsHandler) Export(ctx context.Context, req *collogpb.ExportLogsServiceRequest) (*collogpb.ExportLogsServiceResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	utils.Info("OTLP logs export received with %d resource logs", len(req.ResourceLogs))

	// Convert OTLP request to model.LogPayload(s)
	logPayloads := convertOTLPToModelLogPayloads(req)

	// Process each converted payload (preserving existing business logic)
	for _, converted := range logPayloads {
		SafeHandlePayload(func() {
			// Resource discovery and payload enrichment
			enrichedPayload := h.Sys.Tele.ResourceDiscovery.ProcessLogPayload(&converted)
			if enrichedPayload != nil {
				converted = *enrichedPayload
			}

			// Evaluate severity level of logs and act accordingly (PRESERVED)
			h.EvaluateSeverityLevel(&converted)

			// Check rulesrunner (PRESERVED)
			h.Sys.Tele.Evaluator.EvaluateLogs(h.Sys.Ctx, converted.Logs, converted.Meta)

			// Broadcast to hub.LogHub Websocket (PRESERVED)
			h.Sys.WSHub.Logs.Broadcast(converted)

			// Write to BufferedLog store, fallback to writing to logstore directly (PRESERVED)
			if h.Sys.Buffers == nil || h.Sys.Buffers.Logs == nil { // Fixed: was checking Metrics
				utils.Warn("[otlp] Logs buffer not configured — writing directly to store")
				// Fall back to writing directly to store
				if err := h.Sys.Stores.Logs.Write([]model.LogPayload{converted}); err != nil {
					utils.Warn("Failed to write logs directly to store: %v", err)
				}
			} else {
				if err := h.Sys.Buffers.Logs.WriteAny(converted); err != nil {
					utils.Warn("Failed to buffer LogPayload: %v", err)
				}
			}
		})
	}

	// Return OTLP success response
	return &collogpb.ExportLogsServiceResponse{}, nil
}

// EvaluateSeverityLevel evaluates the severity level of logs based on thresholds defined in the system.
// Based on that severity, different actions can be taken such as generating events that can trigger alerts.
// (COMPLETELY PRESERVED - no changes needed)
func (h *LogsHandler) EvaluateSeverityLevel(logPayload *model.LogPayload) {
	// This function can be used to evaluate the severity level of logs
	// based on thresholds defined in the system.

	// Iterate through Log Entries in logPayload
	for _, logEntry := range logPayload.Logs {
		// Example logic to evaluate severity level
		// This is a placeholder and should be replaced with actual logic
		switch logEntry.Level {
		case "critical", "error":
			// Trigger alert or event for error logs
			utils.Info("Error log detected: %s", logEntry.Message)
			// h.Sys.EventManager.TriggerAlert(logEntry) // Example function call
			evt := model.EventEntry{
				ID:         uuid.NewString(),
				Timestamp:  time.Now(),
				Type:       "log.critical",
				Level:      "critical",
				Category:   "logs",
				Message:    utils.Truncate(logEntry.Message, 256),
				Source:     "logs." + logEntry.Source,
				Scope:      "endpoint",
				Target:     logPayload.EndpointID,
				EndpointID: logPayload.EndpointID,
				Meta:       events.BuildLogEventMeta(&logEntry, logPayload),
			}
			h.Sys.Tele.Emitter.Emit(h.Sys.Ctx, evt)
		case "warning":
			// Handle warning logs if needed
			//utils.Info("Warning log detected: %s", logEntry.Message)
		case "info":
			// Info logs can be ignored or logged as needed
			//utils.Debug("Info log: %s", logEntry.Message)
		default:
			utils.Debug("Log with unknown level: %s", logEntry.Message)
		}
	}
}
