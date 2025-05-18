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

// LogCache is an interface that defines methods for adding and retrieving log entries.
// It is used to cache log entries for quick access and retrieval.
// The cache is thread-safe and uses a mutex to synchronize access to the underlying data structure.
type LogCache interface {
	Add(batch []*model.StoredLog)
	Get(logID string) (*model.LogEntry, bool)
	GetLogs() []*model.StoredLog
}

// logCache is a struct that implements the LogCache interface.
// It uses a map to store log entries, where the key is the log ID and the value is the log entry.
// The cache is protected by a mutex to ensure thread safety.
// It also maintains a set of endpoints to track the endpoints associated with the log entries.
type logCache struct {
	mu        sync.RWMutex
	store     map[string]*model.StoredLog
	endpoints map[string]struct{}
}

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
		if storedLog.LogID == "" {
			utils.Warn("log entry found with no LogID")
			continue
		}
		utils.Debug("Adding logcache: %v", storedLog.Meta.EndpointID)
		c.store[storedLog.LogID] = storedLog
		c.endpoints[storedLog.Meta.EndpointID] = struct{}{}

	}

}

func (c *logCache) Get(logID string) (*model.LogEntry, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	entry, ok := c.store[logID]
	if !ok {
		return nil, false
	}
	return &entry.Log, true
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
