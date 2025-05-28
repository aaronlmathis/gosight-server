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
	"io"
	"time"

	"github.com/aaronlmathis/gosight-server/internal/events"
	"github.com/aaronlmathis/gosight-server/internal/sys"
	"github.com/aaronlmathis/gosight-shared/model"
	pb "github.com/aaronlmathis/gosight-shared/proto"
	"github.com/aaronlmathis/gosight-shared/utils"
	"github.com/google/uuid"
	collogpb "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	commonpb "go.opentelemetry.io/proto/otlp/common/v1"
	logpb "go.opentelemetry.io/proto/otlp/logs/v1"
	resourcepb "go.opentelemetry.io/proto/otlp/resource/v1"
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

func (h *LogsHandler) SubmitStream(stream pb.LogService_SubmitStreamServer) error {
	utils.Info("Log SubmitStream started...")

	for {
		pbPayload, err := stream.Recv()
		if err == io.EOF {
			utils.Info("Log stream closed cleanly by client.")
			return nil
		}
		if err != nil {
			utils.Error("Log stream receive error: %v", err)
			return err
		}
		SafeHandlePayload(func() {
			converted := ConvertToModelLogPayload(pbPayload)

			// Tag enrichment from in-memory cache
			if converted.Meta != nil && converted.Meta.EndpointID != "" {
				tags := h.Sys.Cache.Tags.GetFlattenedTagsForEndpoint(converted.Meta.EndpointID)
				if len(tags) > 0 {
					if converted.Meta.Tags == nil {
						converted.Meta.Tags = make(map[string]string)
					}
					for k, v := range tags {
						if _, exists := converted.Meta.Tags[k]; !exists {
							converted.Meta.Tags[k] = v
						}
					}
				}
			}
			// Evaluate severity level of logs and act accordingly
			h.EvaluateSeverityLevel(&converted)

			// Check rulesrunner
			h.Sys.Tele.Evaluator.EvaluateLogs(h.Sys.Ctx, converted.Logs, converted.Meta)

			// Broadcast to hub.LogHub Websocket
			h.Sys.WSHub.Logs.Broadcast(converted)

			// Write to BufferedLog store, fallback to writing to logstore directly.
			if h.Sys.Buffers == nil || h.Sys.Buffers.Metrics == nil {
				utils.Warn("[stream] Logs buffer not configured — writing directly to store")
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
			// Tag enrichment from in-memory cache (PRESERVED)
			if converted.Meta != nil && converted.Meta.EndpointID != "" {
				tags := h.Sys.Cache.Tags.GetFlattenedTagsForEndpoint(converted.Meta.EndpointID)
				if len(tags) > 0 {
					if converted.Meta.Tags == nil {
						converted.Meta.Tags = make(map[string]string)
					}
					for k, v := range tags {
						if _, exists := converted.Meta.Tags[k]; !exists {
							converted.Meta.Tags[k] = v
						}
					}
				}
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

// convertOTLPToModelLogPayloads converts OTLP logs to GoSight model.LogPayload format
func convertOTLPToModelLogPayloads(req *collogpb.ExportLogsServiceRequest) []model.LogPayload {
	var payloads []model.LogPayload

	for _, resourceLogs := range req.ResourceLogs {
		// Extract resource attributes (host_id, agent_id, etc.)
		meta := extractMetaFromResource(resourceLogs.Resource)

		for _, scopeLogs := range resourceLogs.ScopeLogs {
			var logs []model.LogEntry

			for _, logRecord := range scopeLogs.LogRecords {
				logEntry := model.LogEntry{
					Timestamp: time.Unix(0, int64(logRecord.TimeUnixNano)),
					Level:     convertSeverityToLevel(logRecord.SeverityNumber),
					Message:   logRecord.Body.GetStringValue(),
					Source:    scopeLogs.Scope.GetName(),
					Fields:    convertAttributesToMap(logRecord.Attributes),
				}
				logs = append(logs, logEntry)
			}

			if len(logs) > 0 {
				payload := model.LogPayload{
					AgentID:    meta.AgentID,
					HostID:     meta.HostID,
					Hostname:   meta.Hostname,
					EndpointID: meta.EndpointID,
					Timestamp:  time.Now(),
					Logs:       logs,
					Meta:       &meta,
				}
				payloads = append(payloads, payload)
			}
		}
	}

	return payloads
}

// extractMetaFromResource extracts GoSight meta information from OTLP resource attributes
func extractMetaFromResource(resource *resourcepb.Resource) model.Meta {
	meta := model.Meta{}

	if resource != nil {
		for _, attr := range resource.Attributes {
			switch attr.Key {
			case "host.id":
				meta.HostID = attr.Value.GetStringValue()
			case "service.instance.id":
				meta.AgentID = attr.Value.GetStringValue()
			case "host.name":
				meta.Hostname = attr.Value.GetStringValue()
			case "gosight.endpoint.id":
				meta.EndpointID = attr.Value.GetStringValue()
			}
		}
	}

	return meta
}

// convertSeverityToLevel converts OTLP severity numbers to GoSight log levels
func convertSeverityToLevel(severity logpb.SeverityNumber) string {
	switch {
	case severity >= logpb.SeverityNumber_SEVERITY_NUMBER_FATAL:
		return "critical"
	case severity >= logpb.SeverityNumber_SEVERITY_NUMBER_ERROR:
		return "error"
	case severity >= logpb.SeverityNumber_SEVERITY_NUMBER_WARN:
		return "warning"
	case severity >= logpb.SeverityNumber_SEVERITY_NUMBER_INFO:
		return "info"
	default:
		return "debug"
	}
}

// convertAttributesToMap converts OTLP attributes to string map
func convertAttributesToMap(attributes []*commonpb.KeyValue) map[string]string {
	result := make(map[string]string)
	for _, attr := range attributes {
		result[attr.Key] = attr.Value.GetStringValue()
	}
	return result
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
