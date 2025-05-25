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

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/receiver/otlpreceiver"

	"github.com/aaronlmathis/gosight-server/internal/sys"
	"github.com/aaronlmathis/gosight/shared/model"
)

// OTelReceiver encapsulates OTLP ingestion receivers for metrics and logs.
// It uses the OpenTelemetry Collector receiver factory under the hood.
// Receivers are started via Start() and stopped via Shutdown().
type OTelReceiver struct {
	sysCtx      *sys.SystemContext
	metricsRcvr component.MetricsReceiver
	logsRcvr    component.LogsReceiver
}

// NewOTelReceiver constructs a new OTelReceiver based on sysCtx.Cfg.OpenTelemetry.
func NewOTelReceiver(sysCtx *sys.SystemContext) (*OTelReceiver, error) {
	cfg := &sysCtx.Cfg.OpenTelemetry
	if cfg == nil || !cfg.Enabled {
		return nil, fmt.Errorf("OpenTelemetry ingestion disabled or config missing")
	}

	// Build protocol settings
	protocols := &otlpreceiver.Protocols{}
	if cfg.ReceiverProtocols.GRPC.Enabled && cfg.ReceiverProtocols.GRPC.Endpoint != "" {
		protocols.GRPC = &otlpreceiver.GRPCServerSettings{
			Endpoint: cfg.ReceiverProtocols.GRPC.Endpoint,
		}
	}
	if cfg.ReceiverProtocols.HTTP.Enabled && cfg.ReceiverProtocols.HTTP.Endpoint != "" {
		protocols.HTTP = &otlpreceiver.HTTPServerSettings{
			Endpoint: cfg.ReceiverProtocols.HTTP.Endpoint,
		}
	}
	if protocols.GRPC == nil && protocols.HTTP == nil {
		return nil, fmt.Errorf("no OTLP protocols configured for receiver")
	}

	// OTLP receiver configuration
	rcvCfg := &otlpreceiver.Config{Protocols: *protocols}

	// Create metrics receiver
	metricsFactory := otlpreceiver.NewFactory()
	metricsRcvr, err := metricsFactory.CreateMetricsReceiver(
		sysCtx.Ctx,
		component.ReceiverCreateSettings{},
		rcvCfg,
		&metricConsumer{handle: sysCtx.Tele.MetricIngest},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP metrics receiver: %w", err)
	}

	// Create logs receiver
	logsFactory := otlpreceiver.NewFactory()
	logsRcvr, err := logsFactory.CreateLogsReceiver(
		sysCtx.Ctx,
		component.ReceiverCreateSettings{},
		rcvCfg,
		&logConsumer{handle: sysCtx.Tele.LogIngest},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP logs receiver: %w", err)
	}

	return &OTelReceiver{
		sysCtx:      sysCtx,
		metricsRcvr: metricsRcvr,
		logsRcvr:    logsRcvr,
	}, nil
}

// Start launches the metrics and logs receivers concurrently.
func (r *OTelReceiver) Start() error {
	host := componenttest.NewNopHost()
	if err := r.metricsRcvr.Start(r.sysCtx.Ctx, host); err != nil {
		return fmt.Errorf("metrics receiver start failed: %w", err)
	}
	if err := r.logsRcvr.Start(r.sysCtx.Ctx, host); err != nil {
		return fmt.Errorf("logs receiver start failed: %w", err)
	}
	return nil
}

// Shutdown gracefully stops both receivers.
func (r *OTelReceiver) Shutdown() error {
	if err := r.metricsRcvr.Shutdown(r.sysCtx.Ctx); err != nil {
		return fmt.Errorf("metrics receiver shutdown failed: %w", err)
	}
	if err := r.logsRcvr.Shutdown(r.sysCtx.Ctx); err != nil {
		return fmt.Errorf("logs receiver shutdown failed: %w", err)
	}
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
