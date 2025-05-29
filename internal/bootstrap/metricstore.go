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

package bootstrap

import (
	"context"
	"fmt"

	"github.com/aaronlmathis/gosight-server/internal/cache"
	"github.com/aaronlmathis/gosight-server/internal/config"
	"github.com/aaronlmathis/gosight-server/internal/store/metastore"
	"github.com/aaronlmathis/gosight-server/internal/store/metricindex"
	"github.com/aaronlmathis/gosight-server/internal/store/metricstore"
	"github.com/aaronlmathis/gosight-shared/utils"
)

// InitMetricIndex initializes the metric index component for the GoSight server.
// The metric index provides fast lookup and organization of metric metadata,
// enabling efficient metric discovery, search, and retrieval operations.
// It maintains an in-memory index of metric names, labels, and relationships
// for optimal query performance.
//
// Returns:
//   - *metricindex.MetricIndex: Initialized metric index for fast metric lookups
//   - error: Currently always nil, reserved for future error conditions
func InitMetricIndex() (*metricindex.MetricIndex, error) {

	metricIndex := metricindex.NewMetricIndex()

	return metricIndex, nil
}

// InitMetricStore initializes the metric store component for the GoSight server.
// The metric store provides persistent storage for time-series metric data,
// supporting various storage engines optimized for different deployment scenarios
// and performance requirements.
//
// Supported engines include:
//   - victoriametrics: High-performance time-series database
//   - prometheus: Compatible with Prometheus remote storage
//   - influxdb: InfluxDB time-series database
//   - memory: In-memory storage for testing and development
//
// The metric store integrates with the metric cache to provide fast access
// to frequently queried metrics and reduce load on the underlying storage.
//
// Parameters:
//   - ctx: Context for initialization and cancellation
//   - cfg: Configuration containing metric store settings and engine selection
//   - metricCache: Cache component for improved metric query performance
//
// Returns:
//   - metricstore.MetricStore: Initialized metric store interface implementation
//   - error: If metric store initialization fails for the specified engine
func InitMetricStore(ctx context.Context, cfg *config.Config, metricCache cache.MetricCache) (metricstore.MetricStore, error) {
	engine := cfg.MetricStore.Engine
	utils.Info("Initializing metric store engine: %s", engine)

	s, err := metricstore.InitMetricStore(ctx, cfg, metricCache)
	if err != nil {
		return nil, fmt.Errorf("failed to init metric store: %w", err)
	}

	utils.Info("Metric store [%s] initialized successfully", engine)
	return s, nil
}

// InitMetaStore initializes the metadata tracking component for the GoSight server.
// The meta store (MetaTracker) maintains metadata about system resources,
// including agents, endpoints, services, and their relationships. This metadata
// is essential for resource discovery, monitoring, and system topology mapping.
//
// The MetaTracker provides:
//   - Resource registration and discovery
//   - Metadata caching and indexing
//   - Resource relationship tracking
//   - System topology maintenance
//
// Returns:
//   - *metastore.MetaTracker: Initialized metadata tracker for resource management
func InitMetaStore() *metastore.MetaTracker {
	return metastore.NewMetaTracker()
}
