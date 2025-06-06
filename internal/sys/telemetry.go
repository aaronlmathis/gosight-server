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

// gosight/agent/internal/sys/sys.go
// Package sys provides the system context for the GoSight application.

package sys

import (
	"github.com/aaronlmathis/gosight-server/internal/alerts"
	"github.com/aaronlmathis/gosight-server/internal/core/events/dispatcher"
	"github.com/aaronlmathis/gosight-server/internal/events"
	"github.com/aaronlmathis/gosight-server/internal/rules"
	"github.com/aaronlmathis/gosight-server/internal/store/metastore"
	"github.com/aaronlmathis/gosight-server/internal/store/metricindex"
	"github.com/aaronlmathis/gosight-shared/model"
)

// ResourceDiscoverer defines the interface for resource discovery functionality
type ResourceDiscoverer interface {
	ProcessMetricPayload(payload *model.MetricPayload) *model.MetricPayload
	ProcessLogPayload(payload *model.LogPayload) *model.LogPayload
	ProcessTracePayload(payload *model.TracePayload) *model.TracePayload
}

// TelemetryModule encapsulates telemetry-related state and processing.
type TelemetryModule struct {
	Index             *metricindex.MetricIndex // Metric name/dimension catalog
	Meta              *metastore.MetaTracker   // Tracks source metadata (labels, tags, endpoint info)
	Evaluator         *rules.Evaluator         // Rule evaluator (metrics → match?)
	Alerts            *alerts.Manager          // Tracks alert state per rule/endpoint
	Emitter           *events.Emitter          // Emits events (alerts, system actions)
	Dispatcher        *dispatcher.Dispatcher   // Routes alert events to actions
	ResourceDiscovery ResourceDiscoverer       // Discovers and tracks resources
}

// NewTelemetryModule creates a new TelemetryModule with the provided components.

func NewTelemetryModule(
	index *metricindex.MetricIndex,
	meta *metastore.MetaTracker,
	evaluator *rules.Evaluator,
	alerts *alerts.Manager,
	emitter *events.Emitter,
	dispatcher *dispatcher.Dispatcher,
	resourceDiscovery ResourceDiscoverer,
) *TelemetryModule {
	return &TelemetryModule{
		Index:             index,
		Meta:              meta,
		Evaluator:         evaluator,
		Alerts:            alerts,
		Emitter:           emitter,
		Dispatcher:        dispatcher,
		ResourceDiscovery: resourceDiscovery,
	}
}
