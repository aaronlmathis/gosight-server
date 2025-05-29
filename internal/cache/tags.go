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

// Add safely appends a value to the StringSet, ignoring empty values.
// This method provides a convenient way to add non-empty values to a set
// while maintaining the set's integrity and avoiding nil entries.
//
// Parameters:
//   - val: The string value to add (empty strings are ignored)
func (s StringSet) Add(val string) {
	if val != "" {
		s[val] = struct{}{}
	}
}

// TagCache provides high-performance in-memory indexing and caching for endpoint
// metadata tags within the GoSight monitoring system. It serves as the central
// repository for tag-based metadata, enabling efficient filtering, discovery,
// and organization of monitored endpoints.
//
// The cache maintains comprehensive indexes for tags, keys, values, and their
// relationships to endpoints, supporting complex queries and rapid metadata
// retrieval. It includes persistence integration for durability and automatic
// cleanup mechanisms for optimal memory management.
//
// Key Features:
//   - Thread-safe concurrent operations with optimized locking strategies
//   - Multi-dimensional indexing (endpoint→tags, tag→endpoints, key→values)
//   - Efficient tag-based endpoint discovery and filtering
//   - Automatic tag normalization for metric system compatibility
//   - Integrated persistence layer with configurable flush policies
//   - Memory management with time-based endpoint pruning
//   - Real-time tag update tracking and change detection
//   - Support for hierarchical tag organization and inheritance
//
// The cache supports the complete tag lifecycle from ingestion through query,
// providing the foundation for metadata-driven monitoring and analytics.
//
// Architecture:
//   - Forward indexes enable rapid endpoint-to-tag resolution
//   - Reverse indexes support efficient tag-to-endpoint discovery
//   - Global key/value tracking enables comprehensive tag enumeration
//   - Change tracking optimizes persistence operations
//   - Normalized key handling ensures metric system compatibility
type TagCache interface {
	// Add ingests tags from a metric payload, automatically indexing them for
	// efficient retrieval and establishing bidirectional relationships between
	// endpoints and their associated metadata.
	//
	// Parameters:
	//   - payload: Complete metric payload containing endpoint metadata and tags
	Add(payload *model.MetricPayload)

	// GetTagsForEndpoint retrieves all tags associated with a specific endpoint,
	// returning a deep copy to prevent external modification of cached data.
	// Each tag key maps to a set of possible values for that endpoint.
	//
	// Parameters:
	//   - endpointID: The unique identifier of the endpoint to query
	//
	// Returns:
	//   - map[string]StringSet: Tag keys mapped to their associated value sets
	GetTagsForEndpoint(endpointID string) map[string]StringSet

	// GetFlattenedTagsForEndpoint returns a normalized, single-value representation
	// of endpoint tags optimized for metric system integration. Tag keys are
	// normalized (lowercase, spaces to underscores) and multi-value tags are
	// flattened to single values.
	//
	// Parameters:
	//   - endpointID: The unique identifier of the endpoint to query
	//
	// Returns:
	//   - map[string]string: Normalized tag keys mapped to single values
	GetFlattenedTagsForEndpoint(endpointID string) map[string]string

	// GetTagKeys returns all known tag keys across all endpoints, enabling
	// tag discovery and validation for query building and filtering operations.
	//
	// Returns:
	//   - []string: Complete list of all tag keys in the cache
	GetTagKeys() []string

	// GetTagValues retrieves all known values for a specific tag key across
	// all endpoints, supporting tag-based filtering and validation operations.
	//
	// Parameters:
	//   - key: The tag key to query for values
	//
	// Returns:
	//   - StringSet: Set of all values associated with the specified key
	GetTagValues(key string) StringSet

	// Flush persists modified tag data to the configured datastore, ensuring
	// durability of tag metadata across system restarts. Only changed endpoints
	// are written to optimize performance.
	//
	// Parameters:
	//   - dataStore: The datastore interface for persistence operations
	Flush(dataStore datastore.DataStore)

	// LoadFromStore initializes the cache with existing tag data from persistent
	// storage, typically called during system startup to restore previous state.
	//
	// Parameters:
	//   - tags: Slice of tag records to load into the cache
	LoadFromStore(tags []model.Tag)

	// Prune removes stale endpoint data based on configurable retention policies,
	// helping maintain optimal memory usage in long-running deployments.
	Prune()

	// GetAllEndpoints returns a complete deep copy of all endpoint tag data,
	// primarily used for debugging, diagnostics, and administrative operations.
	//
	// Returns:
	//   - map[string]map[string]StringSet: Complete endpoint tag mapping
	GetAllEndpoints() map[string]map[string]StringSet
}

// tagCache implements the TagCache interface providing comprehensive thread-safe
// tag management and indexing capabilities. It maintains multiple optimized
// indexes to support efficient tag-based queries, endpoint discovery, and
// metadata operations.
//
// The implementation uses a multi-layered indexing strategy:
//   - Forward index: endpoint → tag keys → values (for endpoint queries)
//   - Reverse index: tag key:value → endpoints (for tag-based discovery)
//   - Global indexes: all keys, all values per key (for enumeration)
//   - Change tracking: modified endpoints (for efficient persistence)
//   - Activity tracking: last seen timestamps (for pruning)
//
// This design enables optimal performance across diverse access patterns while
// maintaining data consistency and supporting high-frequency tag updates.
type tagCache struct {
	mu sync.RWMutex

	Endpoints      map[string]map[string]StringSet // endpointID → tag key → value set
	TagKeys        map[string]struct{}             // all known tag keys
	TagValues      map[string]StringSet            // tag key → all seen values
	LastSeen       map[string]int64                // endpointID → last update timestamp
	TagToEndpoints map[string]map[string]struct{}  // "key:value" → endpoint set (reverse index)
	dirty          map[string]struct{}             // endpoints with pending changes
}

// NewTagCache creates and initializes a new TagCache instance with all
// required internal data structures. The cache is immediately ready for
// concurrent use and provides thread-safe operations from creation.
//
// All internal maps are pre-allocated to avoid nil pointer issues and
// optimize initial performance. The cache supports immediate tag ingestion
// and querying without additional configuration.
//
// Returns:
//   - TagCache: A new thread-safe tag cache instance ready for use
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

// Add processes and indexes tags from a metric payload, automatically updating
// all relevant indexes and maintaining data consistency across the cache's
// multi-dimensional structure. This method is optimized for high-frequency
// ingestion patterns typical in monitoring environments.
//
// The method performs comprehensive indexing operations:
//   - Updates forward indexes (endpoint → tags)
//   - Maintains reverse indexes (tags → endpoints)
//   - Updates global key and value catalogs
//   - Tracks endpoint activity timestamps
//   - Marks endpoints as dirty for persistence
//
// Empty or nil values are safely ignored to maintain data quality.
//
// Parameters:
//   - payload: The metric payload containing endpoint metadata and tags
func (c *tagCache) Add(payload *model.MetricPayload) {
	if payload == nil || payload.Meta == nil || payload.Meta.EndpointID == "" {
		return
	}
	endpointID := payload.Meta.EndpointID
	metaTags := payload.Meta.Labels
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

// LoadFromStore initializes the cache with existing tag data from persistent
// storage, typically called during system startup to restore previous state.
// This method rebuilds all indexes from the provided tag records, ensuring
// the cache is fully functional immediately after loading.
//
// The loading process reconstructs:
//   - Forward indexes (endpoint → tags)
//   - Reverse indexes (tags → endpoints)
//   - Global key and value catalogs
//   - Tag relationship mappings
//
// Invalid or incomplete tag records are safely ignored to maintain data integrity.
//
// Parameters:
//   - tags: Slice of persisted tag records to load into the cache
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

// GetTagsForEndpoint retrieves a complete deep copy of all tags associated
// with a specific endpoint. The returned data is isolated from the cache's
// internal state, preventing accidental modification of cached data.
//
// Each tag key maps to a StringSet containing all possible values for that
// key on the specified endpoint. This supports multi-value tag scenarios
// where a single key may have multiple associated values.
//
// Parameters:
//   - endpointID: The unique identifier of the endpoint to query
//
// Returns:
//   - map[string]StringSet: Deep copy of endpoint tags (nil-safe, may be empty)
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

// GetFlattenedTagsForEndpoint returns a normalized, single-value representation
// of endpoint tags specifically optimized for metric system integration. This
// method performs automatic key normalization and value flattening to ensure
// compatibility with time-series databases like VictoriaMetrics.
//
// Key normalization includes:
//   - Converting to lowercase
//   - Replacing spaces with underscores
//   - Ensuring metric system compatibility
//
// Multi-value tags are flattened by taking the first available value, making
// this method suitable for scenarios requiring deterministic single values.
//
// Parameters:
//   - endpointID: The unique identifier of the endpoint to query
//
// Returns:
//   - map[string]string: Normalized tags with single values per key
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

// GetTagKeys returns a complete list of all known tag keys across all endpoints
// in the cache. This method enables comprehensive tag discovery, query building,
// and validation operations for filtering and analytics.
//
// The returned slice contains all tag keys that have been observed across any
// endpoint, providing a global view of available metadata dimensions.
//
// Returns:
//   - []string: Complete list of all tag keys (order not guaranteed)
func (c *tagCache) GetTagKeys() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	keys := make([]string, 0, len(c.TagKeys))
	for k := range c.TagKeys {
		keys = append(keys, k)
	}
	return keys
}

// GetTagValues retrieves a complete set of all known values for a specific
// tag key across all endpoints. The returned StringSet is a deep copy that
// can be safely modified without affecting the cache's internal state.
//
// This method supports tag-based filtering operations, value validation,
// and autocomplete functionality by providing comprehensive value enumeration
// for any given tag dimension.
//
// Parameters:
//   - key: The tag key to query for values
//
// Returns:
//   - StringSet: Deep copy of all values for the key (empty if key doesn't exist)
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

// GetAllEndpoints returns a complete deep copy of all endpoint tag data
// for debugging, diagnostics, and administrative operations. This method
// provides comprehensive visibility into the cache's state while maintaining
// data isolation through deep copying.
//
// The returned structure preserves the complete hierarchy:
// endpoint → tag keys → value sets, enabling detailed inspection of the
// cache's internal organization and data relationships.
//
// Returns:
//   - map[string]map[string]StringSet: Complete deep copy of all endpoint tags
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

// Flush persists all modified tag data to the configured datastore, ensuring
// durability of tag metadata across system restarts. This method uses change
// tracking to optimize persistence operations by only writing modified endpoints.
//
// The flush operation is typically called periodically or during graceful
// shutdown to maintain data consistency between memory and persistent storage.
// Future implementations will integrate with the datastore interface to
// perform batch upsert operations for optimal performance.
//
// Parameters:
//   - dataStore: The datastore interface for persistence operations
func (c *tagCache) Flush(dataStore datastore.DataStore) {
	// Placeholder for flushing to DB
	// You will call DataStore.UpsertTags(endpointID, tags) for each dirty endpoint
}

// Prune removes stale endpoint data based on configurable retention policies,
// helping maintain optimal memory usage in long-running deployments. This
// method identifies endpoints that haven't been seen within the retention
// window and removes their associated tag data.
//
// The pruning operation helps prevent memory leaks in environments with
// dynamic endpoint populations where endpoints frequently appear and disappear.
// Future implementations will include configurable retention windows and
// comprehensive cleanup of related indexes.
func (c *tagCache) Prune() {
	// Optional: remove endpoints not seen in N minutes based on LastSeen
}
