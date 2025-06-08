// SPDX-License-Identifier: GPL-3.0-or-later

// Copyright (C) 2025 Aaron Mathis <aaron.mathis@gmail.com>

// This file is part of GoSight.

// GoSight is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// GoSight is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with GoSight. If not, see https://www.gnu.org/licenses/.
//

package otel

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/aaronlmathis/gosight-server/internal/sys"
	"github.com/aaronlmathis/gosight-shared/utils"
	"github.com/gorilla/mux"
	logpb "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	metricpb "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
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
	o.Router.Handle("/v1/traces", http.HandlerFunc(o.handleTraceIngest)).Methods("POST")
}

// handleMetricIngest processes incoming OTLP metrics requests.
func (o *OTelReceiver) handleMetricIngest(w http.ResponseWriter, r *http.Request) {

	ct := r.Header.Get("Content-Type")
	if r.Method != http.MethodPost || !strings.HasPrefix(ct, "application/x-protobuf") {
		http.Error(w, "Unsupported method or content type", http.StatusUnsupportedMediaType)
		return
	}

	const maxMetricPayloadSize = 10 << 20 // 10 MiB
	defer r.Body.Close()
	body, err := io.ReadAll(io.LimitReader(r.Body, maxMetricPayloadSize))
	if err != nil {
		http.Error(w, "Request body too large", http.StatusRequestEntityTooLarge)
		return
	}

	var req metricpb.ExportMetricsServiceRequest
	if err := proto.Unmarshal(body, &req); err != nil {
		http.Error(w, fmt.Sprintf("Failed to unmarshal trace request: %v", err), http.StatusBadRequest)
		return
	}

	metrics := o.OTLPToMetrics(&req)
	// Save trace to store here, for now log it
	utils.Debug("Received %d metrics in request", len(metrics))
	for _, metric := range metrics {
		utils.Debug("Metric: %s", metric.Name)

		// DO SOMETHING WITH THE SPAN
	}

	// After processing:
	resp := &tracepb.ExportTraceServiceResponse{} // from go.opentelemetry.io/proto/otlp/collector/logs/v1
	data, _ := proto.Marshal(resp)
	w.Header().Set("Content-Type", "application/x-protobuf")
	w.WriteHeader(http.StatusAccepted)
	w.Write(data)
	return

}

// handleLogIngest processes incoming OTLP logs requests.
func (o *OTelReceiver) handleLogIngest(w http.ResponseWriter, r *http.Request) {
	ct := r.Header.Get("Content-Type")
	if r.Method != http.MethodPost || !strings.HasPrefix(ct, "application/x-protobuf") {
		http.Error(w, "Unsupported method or content type", http.StatusUnsupportedMediaType)
		return
	}

	const maxLogPayloadSize = 10 << 20 // 10 MiB
	defer r.Body.Close()
	body, err := io.ReadAll(io.LimitReader(r.Body, maxLogPayloadSize))
	if err != nil {
		http.Error(w, "Request body too large", http.StatusRequestEntityTooLarge)
		return
	}


	var req logpb.ExportLogsServiceRequest
	if err := proto.Unmarshal(body, &req); err != nil {
		http.Error(w, fmt.Sprintf("Failed to unmarshal log request: %v", err), http.StatusBadRequest)
		return
	}

	logEntries := o.OTLPToLogEntries(&req)
	// Save logs to store here, for now log it
	utils.Debug("Received %d log entries in log request", len(logEntries))
	for _, entry := range logEntries {
		utils.Debug("Log Entry: %s - %s - %v", entry.Name, entry.TraceID, entry)

		// DO SOMETHING WITH THE LOG ENTRY
	}
	// After processing:
	resp := &logpb.ExportLogsServiceResponse{} // from go.opentelemetry.io/proto/otlp/collector/logs/v1
	data, _ := proto.Marshal(resp)
	w.Header().Set("Content-Type", "application/x-protobuf")
	w.WriteHeader(http.StatusAccepted)
	w.Write(data)
	return
}

// handleTraceIngest processes incoming OTLP traces requests.
func (o *OTelReceiver) handleTraceIngest(w http.ResponseWriter, r *http.Request) {
	ct := r.Header.Get("Content-Type")
	if r.Method != http.MethodPost || !strings.HasPrefix(ct, "application/x-protobuf") {
		http.Error(w, "Unsupported method or content type", http.StatusUnsupportedMediaType)
		return
	}

	const maxLogPayloadSize = 10 << 20 // 10 MiB
	defer r.Body.Close()
	body, err := io.ReadAll(io.LimitReader(r.Body, maxLogPayloadSize))
	if err != nil {
		http.Error(w, "Request body too large", http.StatusRequestEntityTooLarge)
		return
	}

	var req tracepb.ExportTraceServiceRequest
	if err := proto.Unmarshal(body, &req); err != nil {
		http.Error(w, fmt.Sprintf("Failed to unmarshal trace request: %v", err), http.StatusBadRequest)
		return
	}

	spans := o.OTLPToTraceSpans(&req)
	// Save trace to store here, for now log it
	utils.Debug("Received %d spans in trace request", len(spans))
	for _, span := range spans {
		utils.Debug("Span: %s - %s - %v", span.Name, span.TraceID, span)

		// DO SOMETHING WITH THE SPAN
	}

	// After processing:
	resp := &tracepb.ExportTraceServiceResponse{} // from go.opentelemetry.io/proto/otlp/collector/logs/v1
	data, _ := proto.Marshal(resp)
	w.Header().Set("Content-Type", "application/x-protobuf")
	w.WriteHeader(http.StatusAccepted)
	w.Write(data)
	return
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
