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

package cache

import (
	"sync"
	"time"

	"github.com/aaronlmathis/gosight-shared/model"
)

// snapshotRetention defines the time duration for which process snapshots
// are retained in the cache before being pruned. Snapshots older than this
// duration are automatically removed during cleanup operations.
const snapshotRetention = 30 * time.Minute

// ProcessCache provides high-performance caching for process lifecycle data
// collected from monitored endpoints. It maintains time-ordered snapshots of
// process states, enabling historical analysis, trend detection, and process
// monitoring capabilities within the GoSight monitoring system.
//
// The cache is optimized for high-frequency process data ingestion while
// providing efficient retrieval mechanisms for dashboards, alerting, and
// analytics. It automatically manages memory usage through time-based
// retention policies and maintains data consistency across concurrent
// operations.
//
// Key Features:
//   - Thread-safe concurrent access with optimized read/write locking
//   - Automatic time-based data retention and cleanup
//   - Per-endpoint data organization for efficient retrieval
//   - Memory-efficient storage with configurable retention periods
//   - Real-time process state tracking and historical analysis
//   - Support for process lifecycle monitoring and trend analysis
//
// The cache supports the full process monitoring workflow from data ingestion
// to query, providing the foundation for process-based monitoring and analytics.
type ProcessCache interface {
	// Add ingests a process snapshot into the cache, organizing it by endpoint
	// and maintaining temporal ordering for historical analysis. The snapshot
	// is automatically indexed by timestamp for efficient time-based queries.
	//
	// Parameters:
	//   - snapshot: Complete process snapshot containing process states and metadata
	Add(snapshot model.ProcessSnapshot)

	// Get retrieves all process snapshots for a specific endpoint within the
	// retention window. Snapshots are returned in chronological order, enabling
	// time-series analysis and process lifecycle tracking.
	//
	// Parameters:
	//   - endpointID: The unique identifier of the endpoint to query
	//
	// Returns:
	//   - []model.ProcessSnapshot: Time-ordered snapshots for the endpoint
	Get(endpointID string) []model.ProcessSnapshot

	// Prune removes expired process snapshots from the cache based on the
	// configured retention policy. This method is typically called periodically
	// to maintain optimal memory usage and prevent unbounded cache growth.
	Prune()
}

// processCache implements the ProcessCache interface providing thread-safe
// storage and retrieval of process snapshots organized by endpoint. It uses
// efficient time-based indexing to support both real-time monitoring and
// historical analysis requirements.
//
// The implementation maintains separate snapshot buffers for each endpoint,
// automatically managing retention through timestamp-based pruning. This
// design enables optimal memory usage while preserving data locality for
// per-endpoint queries.
//
// Architecture:
//   - Per-endpoint snapshot buffers for data locality
//   - Time-based automatic cleanup and retention management
//   - Optimized read/write locking for concurrent access patterns
//   - Memory-efficient storage with minimal overhead
type processCache struct {
	mu        sync.RWMutex
	snapshots map[string][]model.ProcessSnapshot // key: endpointID, value: time-ordered snapshots
}

// NewProcessCache creates and initializes a new ProcessCache instance with
// empty storage structures. The cache is immediately ready for concurrent
// use and provides thread-safe operations from the moment of creation.
//
// The returned cache implements automatic memory management through time-based
// retention policies and supports high-frequency data ingestion patterns
// typical in process monitoring scenarios.
//
// Returns:
//   - ProcessCache: A new thread-safe process cache instance
func NewProcessCache() ProcessCache {
	return &processCache{
		snapshots: make(map[string][]model.ProcessSnapshot),
	}
}

// Add ingests a process snapshot into the cache, automatically organizing it
// by endpoint and maintaining temporal ordering for efficient time-based queries.
// The method includes built-in retention management, immediately pruning expired
// snapshots to maintain optimal memory usage.
//
// The snapshot is appended to the endpoint's buffer and old entries beyond the
// retention window are automatically removed. This ensures that the cache
// maintains a rolling window of recent process data without unbounded growth.
//
// Parameters:
//   - snapshot: The process snapshot to add, containing process states and metadata
func (c *processCache) Add(snapshot model.ProcessSnapshot) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Append snapshot to ring buffer per endpoint
	ep := snapshot.EndpointID
	c.snapshots[ep] = append(c.snapshots[ep], snapshot)

	// Trim old snapshots right after append
	cutoff := time.Now().Add(-snapshotRetention)
	buf := c.snapshots[ep]
	for i := 0; i < len(buf); i++ {
		if buf[i].Timestamp.After(cutoff) {
			c.snapshots[ep] = buf[i:]
			return
		}
	}
	c.snapshots[ep] = nil // all entries too old
}

// Get retrieves all process snapshots for a specific endpoint that are within
// the configured retention window. The snapshots are returned in chronological
// order, making them suitable for time-series analysis and process lifecycle
// tracking.
//
// This method provides read-only access to the cached data and is optimized
// for concurrent access patterns. The returned slice contains all available
// snapshots for the endpoint, from oldest to newest within the retention period.
//
// Parameters:
//   - endpointID: The unique identifier of the endpoint to query
//
// Returns:
//   - []model.ProcessSnapshot: Time-ordered snapshots for the endpoint (may be nil if no data exists)
func (c *processCache) Get(endpointID string) []model.ProcessSnapshot {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.snapshots[endpointID]
}

// Prune performs comprehensive cleanup of expired process snapshots across all
// endpoints based on the configured retention policy. This method systematically
// removes snapshots older than the retention window while preserving recent data.
//
// The pruning operation is optimized to minimize memory allocation and provides
// efficient cleanup for large numbers of endpoints. It should be called
// periodically (e.g., via a scheduled task) to maintain optimal cache performance
// and prevent memory leaks in long-running processes.
//
// The method processes all endpoints in a single pass, making it suitable for
// batch cleanup operations in high-throughput monitoring environments.
func (c *processCache) Prune() {
	c.mu.Lock()
	defer c.mu.Unlock()

	cutoff := time.Now().Add(-snapshotRetention)
	for ep, buf := range c.snapshots {
		for i := 0; i < len(buf); i++ {
			if buf[i].Timestamp.After(cutoff) {
				c.snapshots[ep] = buf[i:]
				break
			}
		}
	}
}
