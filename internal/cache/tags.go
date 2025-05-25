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
	"strings"
	"sync"
	"time"

	"github.com/aaronlmathis/gosight-server/internal/store/datastore"
	"github.com/aaronlmathis/gosight-shared/model"
)

// Add adds a value to the StringSet
func (s StringSet) Add(val string) {
	if val != "" {
		s[val] = struct{}{}
	}
}

// TagCache is an in-memory index of endpoint tags
// built from ingested payloads and optionally flushed to a datastore

type TagCache interface {
	Add(payload *model.MetricPayload)
	GetTagsForEndpoint(endpointID string) map[string]StringSet
	GetFlattenedTagsForEndpoint(endpointID string) map[string]string
	GetTagKeys() []string
	GetTagValues(key string) StringSet
	Flush(dataStore datastore.DataStore)
	LoadFromStore(tags []model.Tag)
	Prune()

	// For debugtools
	GetAllEndpoints() map[string]map[string]StringSet
}

// tagCache is a thread-safe in-memory cache for tags
// It uses a map of maps to store tags for each endpoint
// and a reverse index to quickly find endpoints by tag
// It also tracks the last seen time for each endpoint
// and allows for flushing to a datastore
type tagCache struct {
	mu sync.RWMutex

	Endpoints      map[string]map[string]StringSet // endpointID -> key -> values
	TagKeys        map[string]struct{}             // key -> exists
	TagValues      map[string]StringSet            // key -> all seen values
	LastSeen       map[string]int64                // endpointID -> unix timestamp
	TagToEndpoints map[string]map[string]struct{}  // key:value -> set of endpointIDs
	dirty          map[string]struct{}             // endpointIDs that changed
}

// NewTagCache creates a new TagCache
// It initializes the maps to avoid nil checks later
// It is not thread-safe and should only be called once at startup
func NewTagCache() TagCache {
	return &tagCache{
		Endpoints:      make(map[string]map[string]StringSet),
		TagKeys:        make(map[string]struct{}),
		TagValues:      make(map[string]StringSet),
		LastSeen:       make(map[string]int64),
		TagToEndpoints: make(map[string]map[string]struct{}),
		dirty:          make(map[string]struct{}),
	}
}

// Add adds tags from a MetricPayload to the cache at ingestion (telemetry/stream.go)
// It uses the endpointID as the key and the tags as the value
// It also tracks the last seen time for each endpoint
// It does not check for duplicates or empty values
func (c *tagCache) Add(payload *model.MetricPayload) {
	if payload == nil || payload.Meta == nil || payload.Meta.EndpointID == "" {
		return
	}
	endpointID := payload.Meta.EndpointID
	metaTags := payload.Meta.Tags
	if len(metaTags) == 0 {
		return
	}

	now := time.Now().Unix()

	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.Endpoints[endpointID]; !ok {
		c.Endpoints[endpointID] = make(map[string]StringSet)
	}

	for k, v := range metaTags {
		if v == "" {
			continue
		}

		// Set endpoint-level tag
		if _, ok := c.Endpoints[endpointID][k]; !ok {
			c.Endpoints[endpointID][k] = make(StringSet)
		}
		c.Endpoints[endpointID][k][v] = struct{}{}

		// Set global key
		c.TagKeys[k] = struct{}{}

		// Track all values
		if _, ok := c.TagValues[k]; !ok {
			c.TagValues[k] = make(StringSet)
		}
		c.TagValues[k][v] = struct{}{}

		// Reverse index
		rev := k + ":" + v
		if _, ok := c.TagToEndpoints[rev]; !ok {
			c.TagToEndpoints[rev] = make(map[string]struct{})
		}
		c.TagToEndpoints[rev][endpointID] = struct{}{}
	}

	// Track update
	c.LastSeen[endpointID] = now
	c.dirty[endpointID] = struct{}{}
}

// LoadFromStore populates the cache from existing tags in the datastore during cache init/startup
func (c *tagCache) LoadFromStore(tags []model.Tag) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, t := range tags {
		if t.EndpointID == "" || t.Key == "" || t.Value == "" {
			continue
		}

		if _, ok := c.Endpoints[t.EndpointID]; !ok {
			c.Endpoints[t.EndpointID] = make(map[string]StringSet)
		}
		if _, ok := c.Endpoints[t.EndpointID][t.Key]; !ok {
			c.Endpoints[t.EndpointID][t.Key] = make(StringSet)
		}
		c.Endpoints[t.EndpointID][t.Key][t.Value] = struct{}{}

		c.TagKeys[t.Key] = struct{}{}

		if _, ok := c.TagValues[t.Key]; !ok {
			c.TagValues[t.Key] = make(StringSet)
		}
		c.TagValues[t.Key][t.Value] = struct{}{}

		rev := t.Key + ":" + t.Value
		if _, ok := c.TagToEndpoints[rev]; !ok {
			c.TagToEndpoints[rev] = make(map[string]struct{})
		}
		c.TagToEndpoints[rev][t.EndpointID] = struct{}{}
	}
}

// GetTagsForEndpoint retrieves a copy of the tags for a specific endpoint
func (c *tagCache) GetTagsForEndpoint(endpointID string) map[string]StringSet {
	c.mu.RLock()
	defer c.mu.RUnlock()

	source := c.Endpoints[endpointID]
	clone := make(map[string]StringSet)
	for k, set := range source {
		newSet := make(StringSet)
		for val := range set {
			newSet[val] = struct{}{}
		}
		clone[k] = newSet
	}
	return clone
}

// GetFlattenedTagsForEndpoint retrieves a flattened copy of the tags for a specific endpoint
// Keys are normalized for VictoriaMetrics: lowercase and spaces replaced with underscores
func (c *tagCache) GetFlattenedTagsForEndpoint(endpointID string) map[string]string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	source := c.Endpoints[endpointID]
	flat := make(map[string]string)

	for k, set := range source {
		for val := range set {
			// Normalize key for VictoriaMetrics: lowercase and replace spaces with underscores
			normalizedKey := strings.ToLower(strings.ReplaceAll(k, " ", "_"))
			flat[normalizedKey] = val // take first value found
			break
		}
	}
	return flat
}

// GetTagKeys retrieves a copy of the tag keys
// It returns a slice of strings containing all the keys
func (c *tagCache) GetTagKeys() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	keys := make([]string, 0, len(c.TagKeys))
	for k := range c.TagKeys {
		keys = append(keys, k)
	}
	return keys
}

// getTagValues retrieves a copy of the tag values for a specific key
// It returns a StringSet containing all the values for the given key
// It is a copy of the original set
func (c *tagCache) GetTagValues(key string) StringSet {
	c.mu.RLock()
	defer c.mu.RUnlock()

	copySet := make(StringSet)
	if orig, ok := c.TagValues[key]; ok {
		for v := range orig {
			copySet[v] = struct{}{}
		}
	}
	return copySet
}
func (c *tagCache) GetAllEndpoints() map[string]map[string]StringSet {
	c.mu.RLock()
	defer c.mu.RUnlock()

	clone := make(map[string]map[string]StringSet, len(c.Endpoints))
	for eid, tags := range c.Endpoints {
		tagCopy := make(map[string]StringSet, len(tags))
		for k, set := range tags {
			setCopy := make(StringSet)
			for val := range set {
				setCopy[val] = struct{}{}
			}
			tagCopy[k] = setCopy
		}
		clone[eid] = tagCopy
	}
	return clone
}
func (c *tagCache) Flush(dataStore datastore.DataStore) {
	// Placeholder for flushing to DB
	// You will call DataStore.UpsertTags(endpointID, tags) for each dirty endpoint
}

func (c *tagCache) Prune() {
	// Optional: remove endpoints not seen in N minutes based on LastSeen
}
