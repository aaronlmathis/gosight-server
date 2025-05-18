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

const snapshotRetention = 30 * time.Minute

// ProcessCache stores recent process snapshots per endpoint.
type ProcessCache interface {
	Add(snapshot model.ProcessSnapshot)
	Get(endpointID string) []model.ProcessSnapshot
	Prune()
}

// processCache is a struct that implements the ProcessCache interface.
// It uses a map to store process snapshots, where the key is the endpoint ID
// and the value is a slice of process snapshots.
type processCache struct {
	mu        sync.RWMutex
	snapshots map[string][]model.ProcessSnapshot // key: endpointID
}

// NewProcessCache creates a new instance of ProcessCache.
// It initializes the cache with an empty map for storing process snapshots.
// The cache is protected by a mutex to ensure thread safety.
func NewProcessCache() ProcessCache {
	return &processCache{
		snapshots: make(map[string][]model.ProcessSnapshot),
	}
}

// Add adds a new process snapshot to the cache.
// It appends the snapshot to the slice of snapshots for the given endpoint ID.
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

// Get retrieves the process snapshots for a given endpoint ID.
// It returns a slice of process snapshots for the specified endpoint ID.
func (c *processCache) Get(endpointID string) []model.ProcessSnapshot {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.snapshots[endpointID]
}

// Prune removes old snapshots from the cache.
// It iterates through the snapshots and keeps only those that are within the retention period.
// The retention period is defined by the snapshotRetention constant.
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
