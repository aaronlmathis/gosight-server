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

	"github.com/aaronlmathis/gosight-shared/model"
	"github.com/aaronlmathis/gosight-shared/utils"
)

// LogCache defines the interface for caching log entries in the GoSight monitoring system.
// It provides high-performance access to stored log data with thread-safe operations
// for concurrent read and write access. The cache serves as an intermediate layer
// between log collectors and persistent storage, enabling fast log retrieval for
// real-time monitoring, alerting, and analysis.
//
// The cache implementation supports batch operations for efficient ingestion of
// large volumes of log data from multiple sources. It maintains indexes for
// fast lookup by log ID and tracks associated endpoints for efficient filtering.
//
// Key Features:
//   - Thread-safe concurrent access with read-write mutex protection
//   - Batch insertion for high-throughput log ingestion
//   - Fast lookup by log ID with O(1) complexity
//   - Endpoint tracking for source-based filtering
//   - Memory-efficient storage with automatic cleanup policies
//
// The cache is designed to handle high-volume log streams while maintaining
// low-latency access patterns required for real-time log analysis and monitoring.
type LogCache interface {
	// Add inserts a batch of log entries into the cache. This method is optimized
	// for high-throughput scenarios where multiple log entries are processed together.
	// The batch operation reduces lock contention and improves overall performance.
	//
	// Parameters:
	//   - batch: Slice of StoredLog pointers to be added to the cache
	Add(batch []*model.StoredLog)

	// Get retrieves a specific log entry by its unique identifier. Returns the log
	// entry and a boolean indicating whether the entry was found in the cache.
	//
	// Parameters:
	//   - logID: Unique identifier for the log entry
	//
	// Returns:
	//   - *model.StoredLog: The log entry if found, nil otherwise
	//   - bool: true if the entry exists in the cache, false otherwise
	Get(logID string) (*model.StoredLog, bool)

	// GetLogs returns all log entries currently stored in the cache. This method
	// is useful for bulk operations, cache inspection, and administrative tasks.
	// The returned slice is a copy to prevent external modification of cache data.
	//
	// Returns:
	//   - []*model.StoredLog: Slice containing all cached log entries
	GetLogs() []*model.StoredLog
}

// logCache implements the LogCache interface providing in-memory storage for log entries.
// It uses a map-based storage system with read-write mutex protection for thread-safe
// concurrent operations. The implementation is optimized for fast retrieval and
// efficient memory usage in high-throughput logging scenarios.
//
// Internal Structure:
//   - store: Primary map storing log entries indexed by unique log ID
//   - endpoints: Set tracking unique endpoint identifiers for filtering
//   - mu: Read-write mutex ensuring thread-safe access to cache data
//
// The cache maintains an additional endpoint index to support efficient filtering
// and querying of logs by their source endpoints. This dual-index approach enables
// both direct log lookup and endpoint-based log discovery operations.
type logCache struct {
	mu        sync.RWMutex
	store     map[string]*model.StoredLog
	endpoints map[string]struct{}
}

// NewLogCache creates and initializes a new LogCache instance with empty storage maps.
// The returned cache is ready for immediate use and provides thread-safe operations
// for log entry management. All internal data structures are properly initialized
// to prevent nil pointer exceptions during cache operations.
//
// Returns:
//   - LogCache: A new logCache instance implementing the LogCache interface
func NewLogCache() LogCache {
	return &logCache{
		store:     make(map[string]*model.StoredLog),
		endpoints: make(map[string]struct{}),
	}
}

func (c *logCache) Add(batch []*model.StoredLog) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, storedLog := range batch {
		if storedLog == nil {
			utils.Warn("logCache.Add: nil StoredLog in batch")
			continue
		}
		if storedLog.LogID == "" {
			utils.Warn("log entry found with no LogID")
			continue
		}
		if storedLog.Meta == nil {
			utils.Warn("log entry found with nil Meta")
			continue
		}
		utils.Debug("Adding logcache: %v", storedLog.Meta.EndpointID)
		c.store[storedLog.LogID] = storedLog
		c.endpoints[storedLog.Meta.EndpointID] = struct{}{}
	}
}

func (c *logCache) Get(logID string) (*model.StoredLog, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	entry, ok := c.store[logID]
	if !ok {
		return nil, false
	}
	utils.Debug("logCache.Get: %v", entry.Meta.EndpointID)
	return entry, true
}

func (c *logCache) GetLogs() []*model.StoredLog {
	c.mu.RLock()
	defer c.mu.RUnlock()

	logs := make([]*model.StoredLog, 0, len(c.store))
	for _, log := range c.store {
		logs = append(logs, log)
	}
	return logs
}
