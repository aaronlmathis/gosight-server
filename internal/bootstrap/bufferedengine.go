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
	"time"

	"github.com/aaronlmathis/gosight-server/internal/bufferengine"
	"github.com/aaronlmathis/gosight-server/internal/config"
	"github.com/aaronlmathis/gosight-server/internal/sys"
	"github.com/aaronlmathis/gosight-shared/utils"
)

// InitBufferEngine initializes the buffered storage system for the GoSight server.
// The buffer engine provides high-performance data ingestion by buffering writes
// to backend storage systems, reducing latency and improving throughput for
// high-volume data streams like metrics, logs, and events.
//
// Key features:
//   - Configurable buffering for different data types (metrics, logs, data, alerts)
//   - Independent flush intervals and buffer sizes per data type
//   - Worker-based parallel processing for optimal performance
//   - Graceful degradation when backends are unavailable
//   - Memory management and overflow protection
//
// The function creates buffered wrappers around existing storage backends based on
// configuration settings. Each buffer type can be independently enabled/disabled
// and configured with specific parameters:
//   - Buffer size: Maximum items to buffer before forcing a flush
//   - Flush interval: Time-based automatic flushing
//   - Workers: Parallel processing threads for the buffer engine
//
// Parameters:
//   - ctx: Context for buffer engine lifecycle management
//   - cfg: Buffer engine configuration with per-type settings
//   - stores: Backend storage modules to wrap with buffering
//
// Returns:
//   - *sys.BufferModule: Container with all configured buffered stores
func InitBufferEngine(ctx context.Context, cfg *config.BufferEngineConfig, stores *sys.StoreModule) *sys.BufferModule {
	interval := cfg.FlushInterval
	if interval == 0 {
		interval = 5 * time.Second
	}

	workers := cfg.MaxWorkers
	if workers == 0 {
		workers = 2
	}

	buffers := sys.BufferModule{}
	e := bufferengine.NewBufferEngine(ctx, interval, workers)
	utils.Info("InitBufferEngine: Metrics buffering enabled = %v", cfg.Metrics.Enabled)
	if cfg.Metrics.Enabled && stores.Metrics != nil {
		if cfg.Metrics.FlushInterval > 0 {
			interval = cfg.Metrics.FlushInterval
		}
		metricBuffer := bufferengine.NewBufferedMetricStore("metrics", stores.Metrics, cfg.Metrics.BufferSize, interval)
		buffers.Metrics = metricBuffer
		e.RegisterStore(metricBuffer)
	}
	utils.Info("InitBufferEngine: Log buffering enabled = %v", cfg.Logs.Enabled)
	if cfg.Logs.Enabled && stores.Logs != nil {
		if cfg.Logs.FlushInterval > 0 {
			interval = cfg.Logs.FlushInterval
		}
		logBuffer := bufferengine.NewBufferedLogStore("logs", stores.Logs, cfg.Logs.BufferSize, interval)
		buffers.Logs = logBuffer
		e.RegisterStore(logBuffer)
	}

	if cfg.Data.Enabled && stores.Data != nil {
		if cfg.Data.FlushInterval > 0 {
			interval = cfg.Data.FlushInterval
		}
		dataBuffer := bufferengine.NewBufferedDataStore(ctx, "data", stores.Data, cfg.Logs.BufferSize, interval)
		buffers.Data = dataBuffer
		e.RegisterStore(dataBuffer)
	}

	if cfg.Alerts.Enabled && stores.Alerts != nil {
		// Implement alert buffering if needed
	}

	e.Start()
	return &buffers
}
