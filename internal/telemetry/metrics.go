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

// Package telemetry provides OTLP metrics handling for the GoSight server.
// This module implements the OpenTelemetry Protocol (OTLP) metrics collection
// service, converting incoming OTLP metrics to GoSight's internal model format
// and processing them through the complete telemetry pipeline including rule
// evaluation, indexing, broadcasting, and storage.

package telemetry

import (
	"context"

	"github.com/aaronlmathis/gosight-server/internal/sys"
	"github.com/aaronlmathis/gosight-shared/model"
	"github.com/aaronlmathis/gosight-shared/utils"
	colmetricpb "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// MetricsHandler implements the OTLP MetricsService for processing metric telemetry data.
// This handler receives OTLP metrics via unary gRPC calls, converts them to GoSight's
// internal model format, and processes them through the complete telemetry pipeline.
// It handles tag enrichment, rule evaluation, agent tracking, broadcasting to WebSocket
// clients, buffering/storage, metric indexing, and caching.
type MetricsHandler struct {
	Sys *sys.SystemContext
	colmetricpb.UnimplementedMetricsServiceServer
}

// NewMetricsHandler creates a new OTLP metrics handler with the provided system context.
// The handler initializes with access to the complete GoSight system including stores,
// buffers, caches, WebSocket hubs, rule evaluators, and metric indexing systems.
// It logs the initialization with details about the configured metric store type.
func NewMetricsHandler(sys *sys.SystemContext) *MetricsHandler {
	utils.Debug("MetricsHandler initialized with store: %T", sys.Stores.Metrics)
	return &MetricsHandler{
		Sys: sys,
	}
}

// Export implements the OTLP MetricsService Export method for receiving metric telemetry.
// This method handles incoming OTLP ExportMetricsServiceRequest messages, converts them
// to GoSight's internal MetricPayload format, and processes them through the complete
// telemetry pipeline. The processing includes:
//
// - Tag enrichment from endpoint-specific tag cache
// - Rule evaluation for alerting and event generation
// - Agent and container information tracking
// - Real-time broadcasting to WebSocket clients
// - Buffered storage with fallback to direct store writes
// - Metric indexing for search and discovery
// - In-memory caching for performance optimization
//
// The method returns an OTLP-compliant success response or an error status if the
// request is invalid or processing fails. All processing is wrapped in SafeHandlePayload
// to ensure robust error handling and prevent service disruption.
func (h *MetricsHandler) Export(ctx context.Context, req *colmetricpb.ExportMetricsServiceRequest) (*colmetricpb.ExportMetricsServiceResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	// Convert OTLP request to model.MetricPayload(s) using comprehensive conversion
	metricPayloads := convertOTLPToModelMetricPayloads(req)
	utils.Info("OTLP metrics export received with %d resource metrics", len(metricPayloads))
	// Process each converted payload using existing business logic
	for _, converted := range metricPayloads {
		utils.Debug("Processing converted MetricPayload with %d metrics", len(converted.Metrics))
		SafeHandlePayload(func() {
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

			// Evaluate rules
			h.Sys.Tele.Evaluator.EvaluateMetric(h.Sys.Ctx, converted.Metrics, converted.Meta)

			// Update agent/container info
			h.Sys.Tracker.UpdateAgent(converted.Meta)
			if converted.Meta.ContainerID != "" {
				h.Sys.Tracker.UpdateContainer(converted.Meta)
			}

			// Broadcast + store metrics
			h.Sys.WSHub.Metrics.Broadcast(converted)

			if h.Sys.Buffers == nil || h.Sys.Buffers.Metrics == nil {
				utils.Warn("[otlp] Metrics buffer not configured — writing directly to store")
				err := h.Sys.Stores.Metrics.Write([]model.MetricPayload{converted})
				if err != nil {
					utils.Warn("Failed to store MetricPayload: %v", err)
				}
			} else {
				err := h.Sys.Buffers.Metrics.WriteAny(converted)
				if err != nil {
					utils.Warn("Failed to buffer MetricPayload: %v", err)
				}
			}

			// Metric indexing
			if converted.Meta != nil && converted.Meta.EndpointID != "" {
				h.Sys.Tele.Meta.Set(converted.Meta.EndpointID, *converted.Meta)
			}
			for _, m := range converted.Metrics {
				merged := MergeDimensionsWithMeta(m.Dimensions, converted.Meta)
				h.Sys.Tele.Index.Add(m.Namespace, m.SubNamespace, m.Name, merged)
			}

			// Add to cache
			h.Sys.Cache.Metrics.Add(&converted)
		})
	}

	// Return OTLP success response
	return &colmetricpb.ExportMetricsServiceResponse{}, nil
}
