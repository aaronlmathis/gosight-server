/*
SPDX-License-Identifier: GPL-3.0-or-later

Copyright (C) 2025 Aaron Mathis <aaron.mathis@gmail.com>

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

package otel

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/aaronlmathis/gosight-server/internal/sys"
	"github.com/aaronlmathis/gosight-shared/utils"
	"github.com/aaronlmathis/gosight/shared/model"
	"github.com/gorilla/mux"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	tracepb "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	"google.golang.org/protobuf/proto"
)

// OTelReceiver encapsulates OTLP ingestion receivers for metrics and logs.
// It uses the OpenTelemetry Collector receiver factory under the hood.
// Receivers are started via Start() and stopped via Shutdown().
type OTelReceiver struct {
	sysCtx     *sys.SystemContext
	httpServer *http.Server // HTTP server for OTLP ingestion
	Router     *mux.Router
}

// NewOTelReceiver constructs a new OTelReceiver based on sysCtx.Cfg.OpenTelemetry.
func NewOTelReceiver(sysCtx *sys.SystemContext) (*OTelReceiver, error) {

	// Create a new router for the OTLP receiver
	router := mux.NewRouter()
	// Create an HTTP server for OTLP ingestion
	httpServer := &http.Server{
		Addr:    sysCtx.Cfg.OpenTelemetry.HTTP.Addr,
		Handler: router, // Use the new router for OTLP handlers
	}
	// Create the OTLP metrics receiver
	return &OTelReceiver{
		sysCtx:     sysCtx,
		httpServer: httpServer,
		Router:     router,
	}, nil
}

// Start launches the metrics and logs receivers concurrently.
func (o *OTelReceiver) Start() error {
	o.setupRoutes()

	// Start the HTTP server for OTLP ingestion
	utils.Info("Starting OTEL HTTP server at %s", o.httpServer.Addr)
	if err := o.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("failed to start OTEL HTTP server: %w", err)
	}
	return nil
}

// setupRoutes configures the HTTP routes for OTLP ingestion.
func (o *OTelReceiver) setupRoutes() {
	// Register OTLP routes
	o.Router.Handle("/v1/metrics", http.HandlerFunc(o.handleMetricIngest)).Methods("POST")
	o.Router.Handle("/v1/logs", http.HandlerFunc(o.handleLogIngest)).Methods("POST")
	o.Router.Handle("/v1/trace", http.HandlerFunc(o.handleTraceIngest)).Methods("POST")
}

// handleMetricIngest processes incoming OTLP metrics requests.
func (o *OTelReceiver) handleMetricIngest(w http.ResponseWriter, r *http.Request) {

}

// handleLogIngest processes incoming OTLP logs requests.
func (o *OTelReceiver) handleLogIngest(w http.ResponseWriter, r *http.Request) {

}

// handleTraceIngest processes incoming OTLP traces requests.
func (o *OTelReceiver) handleTraceIngest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost || r.Header.Get("Content-Type") != "application/x-protobuf" {
		http.Error(w, "Unsupported method or content type", http.StatusMethodNotAllowed)
		return
	}
	body, _ := io.ReadAll(r.Body)
	defer r.Body.Close()

	var req tracepb.ExportTraceServiceRequest
	if err := proto.Unmarshal(body, &req); err != nil {
		http.Error(w, fmt.Sprintf("Failed to unmarshal trace request: %v", err), http.StatusBadRequest)
		return
	}

	spans := o.OTLPToTraceSpans(&req)
	// Save trace to store here, for now log it
	utils.Debug("Received %d spans in trace request", len(spans))
	for _, span := range spans {
		utils.Debug("Span: %s - %s", span.Name, span.TraceID)
	}
	// Respond with success
	w.WriteHeader(http.StatusAccepted)
}

// Shutdown gracefully stops both receivers.
func (o *OTelReceiver) Shutdown() error {
	utils.Info("Shutting down OTEL HTTP server...")

	if err := o.httpServer.Shutdown(o.sysCtx.Ctx); err != nil {
		utils.Error("otel HTTP shutdown error: %v", err)
		return err
	}

	utils.Info("OTEL HTTP server shut down cleanly")
	return nil
}

// metricConsumer adapts OTLP metrics to GoSight's internal model.
type metricConsumer struct {
	handle func([]model.MetricPayload)
}

// ConsumeMetrics implements consumer.Metrics interface.
func (c *metricConsumer) ConsumeMetrics(ctx context.Context, md pmetric.Metrics) error {
	// TODO: convert md into []model.MetricPayload and call c.handle(payloads)
	return nil
}

// logConsumer adapts OTLP logs to GoSight's internal model.
type logConsumer struct {
	handle func([]model.LogEntry)
}

// ConsumeLogs implements consumer.Logs interface.
func (c *logConsumer) ConsumeLogs(ctx context.Context, ld plog.Logs) error {
	// TODO: convert ld into []model.LogEntry and call c.handle(entries)
	return nil
}
