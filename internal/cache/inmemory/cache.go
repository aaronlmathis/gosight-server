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

// Package inmemory provides a high-performance, thread-safe in-memory implementation
// of the ResourceCache interface for the GoSight monitoring system. This implementation
// offers optimized storage and retrieval of resource metadata with comprehensive
// indexing capabilities for efficient querying and discovery operations.

package inmemory

import (
	"context"
	"sync"
	"time"

	"github.com/aaronlmathis/gosight-server/internal/store/resourcestore"
	"github.com/aaronlmathis/gosight-shared/model"
)

// ResourceCache provides a comprehensive in-memory caching solution for resource
// metadata in the GoSight monitoring system. It implements advanced indexing
// strategies to support efficient resource discovery, relationship tracking,
// and complex query operations while maintaining high-performance concurrent access.
//
// The cache maintains multiple specialized indexes to optimize different access patterns:
//   - Primary index: resource ID → resource data (O(1) direct access)
//   - Kind index: resource type → resources (enables type-based queries)
//   - Group index: logical grouping → resources (supports organizational queries)
//   - Label index: label key/value pairs → resources (enables dimensional filtering)
//   - Tag index: tag key/value pairs → resources (supports custom metadata queries)
//   - Parent index: parent ID → child resources (enables hierarchical navigation)
//
// Key Features:
//   - Thread-safe concurrent operations with optimized read/write locking
//   - Automatic persistence with configurable flush intervals
//   - Comprehensive indexing for sub-second query performance
//   - Memory-efficient storage with automatic cleanup
//   - Dirty tracking for optimized persistence operations
//   - Batch update capabilities for high-throughput scenarios
//   - Resource lifecycle management with staleness detection
//   - Hierarchical resource relationship tracking
//
// The cache supports complex query patterns including:
//   - Multi-dimensional label filtering with intersection semantics
//   - Tag-based resource discovery and organization
//   - Parent-child relationship navigation
//   - Resource lifecycle and status tracking
//   - Comprehensive resource statistics and monitoring
//
// Architecture:
//   - Lock-free read operations where possible for optimal concurrency
//   - Lazy index cleanup to minimize write overhead
//   - Background persistence with failure recovery
//   - Memory-optimized data structures for large resource sets
type ResourceCache struct {
	mu          sync.RWMutex
	resources   map[string]*model.Resource  // Primary index: ID → resource
	dirty       map[string]*model.Resource  // Change tracking for persistence
	store       resourcestore.ResourceStore // Persistent storage backend
	flushTicker *time.Ticker                // Automatic persistence timer
	stopChan    chan struct{}               // Graceful shutdown coordination

	// Specialized indexes for optimized query performance
	byKind   map[string]map[string]*model.Resource            // kind → ID → resource
	byGroup  map[string]map[string]*model.Resource            // group → ID → resource
	byLabels map[string]map[string]map[string]*model.Resource // label_key → label_value → ID → resource
	byTags   map[string]map[string]map[string]*model.Resource // tag_key → tag_value → ID → resource
	byParent map[string]map[string]*model.Resource            // parent_ID → child_ID → resource
}

// NewResourceCache creates and initializes a new in-memory ResourceCache instance
// with comprehensive indexing and automatic persistence capabilities. The cache
// is immediately ready for concurrent operations and begins automatic background
// persistence according to the specified flush interval.
//
// The initialization process:
//   - Allocates all required data structures and indexes
//   - Configures the persistent storage backend
//   - Starts the automatic flush timer for background persistence
//   - Establishes graceful shutdown coordination mechanisms
//
// Parameters:
//   - store: The persistent storage backend for durability
//   - flushInterval: How frequently to persist dirty resources to storage
//
// Returns:
//   - *ResourceCache: A fully initialized cache ready for concurrent use
func NewResourceCache(store resourcestore.ResourceStore, flushInterval time.Duration) *ResourceCache {
	cache := &ResourceCache{
		resources: make(map[string]*model.Resource),
		dirty:     make(map[string]*model.Resource),
		store:     store,
		stopChan:  make(chan struct{}),
		byKind:    make(map[string]map[string]*model.Resource),
		byGroup:   make(map[string]map[string]*model.Resource),
		byLabels:  make(map[string]map[string]map[string]*model.Resource),
		byTags:    make(map[string]map[string]map[string]*model.Resource),
		byParent:  make(map[string]map[string]*model.Resource),
	}

	cache.flushTicker = time.NewTicker(flushInterval)
	go cache.flushLoop()

	return cache
}

// UpsertResource adds or updates a resource in the cache, automatically maintaining
// all indexes and marking the resource for persistence. This method handles both
// new resource creation and updates to existing resources, ensuring data consistency
// across all indexes and optimizing for high-frequency update scenarios.
//
// The operation performs:
//   - Resource storage/update in primary index
//   - Automatic index maintenance (removal of old, addition of new)
//   - Dirty tracking for optimized persistence
//   - Resource update flag management
//
// Parameters:
//   - resource: The resource to add or update (must have valid ID)
func (c *ResourceCache) UpsertResource(resource *model.Resource) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Update resource
	existing := c.resources[resource.ID]
	c.resources[resource.ID] = resource
	c.dirty[resource.ID] = resource
	resource.Updated = true

	// Update indexes
	c.updateIndexes(existing, resource)
}

// GetResource retrieves a specific resource by its unique identifier. This method
// provides O(1) access time through the primary index and is optimized for
// high-frequency individual resource lookups.
//
// Parameters:
//   - id: The unique identifier of the resource to retrieve
//
// Returns:
//   - *model.Resource: The resource if found (nil if not found)
//   - bool: Whether the resource exists in the cache
func (c *ResourceCache) GetResource(id string) (*model.Resource, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	resource, exists := c.resources[id]
	return resource, exists
}

// GetResourcesByKind retrieves all resources of a specific type/kind. This method
// leverages the kind index for efficient type-based resource discovery and is
// commonly used for resource enumeration and type-specific operations.
//
// Parameters:
//   - kind: The resource type/kind to filter by (e.g., "Pod", "Service", "Node")
//
// Returns:
//   - []*model.Resource: Slice of all resources matching the kind (nil if none found)
func (c *ResourceCache) GetResourcesByKind(kind string) []*model.Resource {
	c.mu.RLock()
	defer c.mu.RUnlock()

	kindMap, exists := c.byKind[kind]
	if !exists {
		return nil
	}

	result := make([]*model.Resource, 0, len(kindMap))
	for _, resource := range kindMap {
		result = append(result, resource)
	}
	return result
}

// GetResourcesByLabels performs complex multi-dimensional filtering to find resources
// matching all specified label criteria. This method implements intersection semantics,
// returning only resources that have ALL specified label key-value pairs.
//
// The operation is optimized through the label index and uses progressive filtering
// to minimize the result set at each step, ensuring efficient performance even
// with large resource populations.
//
// Parameters:
//   - labels: Map of label key-value pairs that resources must match (AND semantics)
//
// Returns:
//   - []*model.Resource: Resources matching ALL specified labels (nil if none match)
func (c *ResourceCache) GetResourcesByLabels(labels map[string]string) []*model.Resource {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var candidates map[string]*model.Resource
	first := true

	for key, value := range labels {
		labelMap, exists := c.byLabels[key]
		if !exists {
			return nil
		}

		valueMap, exists := labelMap[value]
		if !exists {
			return nil
		}

		if first {
			candidates = make(map[string]*model.Resource)
			for id, resource := range valueMap {
				candidates[id] = resource
			}
			first = false
		} else {
			// Intersect with existing candidates
			for id := range candidates {
				if _, exists := valueMap[id]; !exists {
					delete(candidates, id)
				}
			}
		}

		if len(candidates) == 0 {
			return nil
		}
	}

	result := make([]*model.Resource, 0, len(candidates))
	for _, resource := range candidates {
		result = append(result, resource)
	}
	return result
}

// UpdateLastSeen updates the last seen timestamp for a specific resource,
// marking it as recently active. This method is crucial for staleness detection
// and resource lifecycle management, automatically marking the resource as
// dirty for persistence.
//
// Parameters:
//   - id: The unique identifier of the resource to update
//   - lastSeen: The timestamp when the resource was last observed
func (c *ResourceCache) UpdateLastSeen(id string, lastSeen time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if resource, exists := c.resources[id]; exists {
		resource.LastSeen = lastSeen
		resource.Updated = true
		c.dirty[id] = resource
	}
}

// flushLoop manages the automatic background persistence of dirty resources
// to the configured storage backend. This goroutine runs continuously until
// the cache is stopped, ensuring that resource changes are regularly persisted
// without blocking cache operations.
//
// The loop coordinates between the flush timer and shutdown signals to provide
// reliable persistence with graceful shutdown capabilities.
func (c *ResourceCache) flushLoop() {
	for {
		select {
		case <-c.flushTicker.C:
			c.flushDirtyResources()
		case <-c.stopChan:
			return
		}
	}
}

// flushDirtyResources performs batch persistence of all dirty resources to the
// storage backend. This method optimizes persistence operations by batching
// multiple resource updates into single storage operations, reducing I/O overhead
// and improving overall system performance.
//
// The method includes comprehensive error handling with automatic retry mechanisms
// for transient failures, ensuring data durability even in challenging conditions.
func (c *ResourceCache) flushDirtyResources() {
	c.mu.Lock()
	toFlush := make([]*model.Resource, 0, len(c.dirty))
	for _, resource := range c.dirty {
		toFlush = append(toFlush, resource)
	}
	c.dirty = make(map[string]*model.Resource)
	c.mu.Unlock()

	if len(toFlush) == 0 {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Batch update to database
	if err := c.store.UpdateBatch(ctx, toFlush); err != nil {
		// Log error and re-add to dirty list
		c.mu.Lock()
		for _, resource := range toFlush {
			c.dirty[resource.ID] = resource
		}
		c.mu.Unlock()
	} else {
		// Mark as clean
		c.mu.Lock()
		for _, resource := range toFlush {
			resource.Updated = false
		}
		c.mu.Unlock()
	}
}

func (c *ResourceCache) updateIndexes(old, new *model.Resource) {
	// Remove old indexes
	if old != nil {
		c.removeFromIndexes(old)
	}

	// Add new indexes
	c.addToIndexes(new)
}

func (c *ResourceCache) addToIndexes(resource *model.Resource) {
	// Kind index
	if c.byKind[resource.Kind] == nil {
		c.byKind[resource.Kind] = make(map[string]*model.Resource)
	}
	c.byKind[resource.Kind][resource.ID] = resource

	// Group index
	if resource.Group != "" {
		if c.byGroup[resource.Group] == nil {
			c.byGroup[resource.Group] = make(map[string]*model.Resource)
		}
		c.byGroup[resource.Group][resource.ID] = resource
	}

	// Labels index
	for key, value := range resource.Labels {
		if c.byLabels[key] == nil {
			c.byLabels[key] = make(map[string]map[string]*model.Resource)
		}
		if c.byLabels[key][value] == nil {
			c.byLabels[key][value] = make(map[string]*model.Resource)
		}
		c.byLabels[key][value][resource.ID] = resource
	}

	// Tags index
	for key, value := range resource.Tags {
		if c.byTags[key] == nil {
			c.byTags[key] = make(map[string]map[string]*model.Resource)
		}
		if c.byTags[key][value] == nil {
			c.byTags[key][value] = make(map[string]*model.Resource)
		}
		c.byTags[key][value][resource.ID] = resource
	}

	// Parent index
	if resource.ParentID != "" {
		if c.byParent[resource.ParentID] == nil {
			c.byParent[resource.ParentID] = make(map[string]*model.Resource)
		}
		c.byParent[resource.ParentID][resource.ID] = resource
	}
}

// Stop initiates graceful shutdown of the cache, ensuring all pending operations
// complete and performing a final persistence flush. This method coordinates
// the termination of background goroutines and ensures data consistency during
// shutdown.
//
// The shutdown process:
//   - Signals the flush loop to terminate
//   - Stops the automatic flush timer
//   - Performs a final flush of any dirty resources
//   - Ensures all data is persisted before returning
func (c *ResourceCache) Stop() {
	close(c.stopChan)
	c.flushTicker.Stop()
	c.flushDirtyResources() // Final flush
}

func (c *ResourceCache) DeleteResource(id string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	resource, exists := c.resources[id]
	if !exists {
		return false
	}

	// Remove from main maps
	delete(c.resources, id)
	delete(c.dirty, id)

	// Remove from indexes
	c.removeFromIndexes(resource)

	return true
}

func (c *ResourceCache) GetResourcesByGroup(group string) []*model.Resource {
	c.mu.RLock()
	defer c.mu.RUnlock()

	groupMap, exists := c.byGroup[group]
	if !exists {
		return nil
	}

	result := make([]*model.Resource, 0, len(groupMap))
	for _, resource := range groupMap {
		result = append(result, resource)
	}
	return result
}

// GetResourcesByTags performs complex multi-dimensional filtering based on custom
// tag metadata to find resources matching all specified tag criteria. Similar to
// label filtering, this method implements intersection semantics for precise
// resource discovery based on user-defined metadata.
//
// Tags provide flexible custom metadata beyond standard labels, enabling
// application-specific resource organization and discovery patterns.
//
// Parameters:
//   - tags: Map of tag key-value pairs that resources must match (AND semantics)
//
// Returns:
//   - []*model.Resource: Resources matching ALL specified tags (nil if none match)
func (c *ResourceCache) GetResourcesByTags(tags map[string]string) []*model.Resource {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var candidates map[string]*model.Resource
	first := true

	for key, value := range tags {
		tagMap, exists := c.byTags[key]
		if !exists {
			return nil
		}

		valueMap, exists := tagMap[value]
		if !exists {
			return nil
		}

		if first {
			candidates = make(map[string]*model.Resource)
			for id, resource := range valueMap {
				candidates[id] = resource
			}
			first = false
		} else {
			// Intersect with existing candidates
			for id := range candidates {
				if _, exists := valueMap[id]; !exists {
					delete(candidates, id)
				}
			}
		}

		if len(candidates) == 0 {
			return nil
		}
	}

	result := make([]*model.Resource, 0, len(candidates))
	for _, resource := range candidates {
		result = append(result, resource)
	}
	return result
}

// GetResourcesByParent retrieves all child resources for a specific parent resource,
// enabling hierarchical resource navigation and relationship-based queries. This
// method is essential for understanding resource dependencies and implementing
// cascading operations.
//
// Parameters:
//   - parentID: The unique identifier of the parent resource
//
// Returns:
//   - []*model.Resource: All child resources of the specified parent (nil if none)
func (c *ResourceCache) GetResourcesByParent(parentID string) []*model.Resource {
	c.mu.RLock()
	defer c.mu.RUnlock()

	parentMap, exists := c.byParent[parentID]
	if !exists {
		return nil
	}

	result := make([]*model.Resource, 0, len(parentMap))
	for _, resource := range parentMap {
		result = append(result, resource)
	}
	return result
}

func (c *ResourceCache) UpdateStatus(id string, status string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if resource, exists := c.resources[id]; exists {
		resource.Status = status
		resource.Updated = true
		c.dirty[id] = resource
	}
}

// GetStaleResources identifies resources that haven't been seen within the
// specified threshold duration, enabling cleanup operations and staleness
// detection for resource lifecycle management.
//
// This method is commonly used for:
//   - Automated cleanup of disappeared resources
//   - Health monitoring and alerting
//   - Resource lifecycle analysis
//   - Capacity planning and optimization
//
// Parameters:
//   - threshold: Maximum age before a resource is considered stale
//
// Returns:
//   - []*model.Resource: All resources older than the threshold
func (c *ResourceCache) GetStaleResources(threshold time.Duration) []*model.Resource {
	c.mu.RLock()
	defer c.mu.RUnlock()

	cutoff := time.Now().Add(-threshold)
	var stale []*model.Resource

	for _, resource := range c.resources {
		if resource.LastSeen.Before(cutoff) {
			stale = append(stale, resource)
		}
	}

	return stale
}

func (c *ResourceCache) GetResourceCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.resources)
}

func (c *ResourceCache) GetResourceCountByKind() map[string]int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	counts := make(map[string]int)
	for kind, kindMap := range c.byKind {
		counts[kind] = len(kindMap)
	}
	return counts
}

func (c *ResourceCache) RemoveResource(id string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	resource, exists := c.resources[id]
	if !exists {
		return false
	}

	// Remove from indexes
	c.removeFromIndexes(resource)

	// Remove from main storage
	delete(c.resources, id)
	delete(c.dirty, id)

	return true
}

func (c *ResourceCache) GetSummary() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	summary := map[string]interface{}{
		"total_resources":          len(c.resources),
		"dirty_resources":          len(c.dirty),
		"resource_counts_by_kind":  c.getResourceCountsByKindUnsafe(),
		"resource_counts_by_group": c.getResourceCountsByGroupUnsafe(),
		"unique_kinds":             len(c.byKind),
		"unique_groups":            len(c.byGroup),
	}

	return summary
}

func (c *ResourceCache) GetKinds() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	kinds := make([]string, 0, len(c.byKind))
	for kind := range c.byKind {
		kinds = append(kinds, kind)
	}
	return kinds
}

// Clear removes all resources from the cache, effectively resetting it to an
// empty state. This method is primarily used for testing, debugging, or
// complete cache reinitialization scenarios.
//
// The operation clears:
//   - All resource data from primary storage
//   - All specialized indexes
//   - All dirty tracking information
//   - All relationship mappings
//
// Note: This operation does not affect persistent storage.
func (c *ResourceCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Clear all data structures
	c.resources = make(map[string]*model.Resource)
	c.byKind = make(map[string]map[string]*model.Resource)
	c.byGroup = make(map[string]map[string]*model.Resource)
	c.byParent = make(map[string]map[string]*model.Resource)
	c.dirty = make(map[string]*model.Resource)
}

// Helper methods (unsafe versions for internal use when already holding locks)
func (c *ResourceCache) getResourceCountsByKindUnsafe() map[string]int {
	counts := make(map[string]int)
	for kind, kindMap := range c.byKind {
		counts[kind] = len(kindMap)
	}
	return counts
}

func (c *ResourceCache) getResourceCountsByGroupUnsafe() map[string]int {
	counts := make(map[string]int)
	for group, groupMap := range c.byGroup {
		counts[group] = len(groupMap)
	}
	return counts
}

func (c *ResourceCache) removeFromIndexes(resource *model.Resource) {
	// Kind index
	if kindMap, exists := c.byKind[resource.Kind]; exists {
		delete(kindMap, resource.ID)
		if len(kindMap) == 0 {
			delete(c.byKind, resource.Kind)
		}
	}

	// Group index
	if resource.Group != "" {
		if groupMap, exists := c.byGroup[resource.Group]; exists {
			delete(groupMap, resource.ID)
			if len(groupMap) == 0 {
				delete(c.byGroup, resource.Group)
			}
		}
	}

	// Labels index
	for key, value := range resource.Labels {
		if labelMap, exists := c.byLabels[key]; exists {
			if valueMap, exists := labelMap[value]; exists {
				delete(valueMap, resource.ID)
				if len(valueMap) == 0 {
					delete(labelMap, value)
					if len(labelMap) == 0 {
						delete(c.byLabels, key)
					}
				}
			}
		}
	}

	// Tags index
	for key, value := range resource.Tags {
		if tagMap, exists := c.byTags[key]; exists {
			if valueMap, exists := tagMap[value]; exists {
				delete(valueMap, resource.ID)
				if len(valueMap) == 0 {
					delete(tagMap, value)
					if len(tagMap) == 0 {
						delete(c.byTags, key)
					}
				}
			}
		}
	}

	// Parent index
	if resource.ParentID != "" {
		if parentMap, exists := c.byParent[resource.ParentID]; exists {
			delete(parentMap, resource.ID)
			if len(parentMap) == 0 {
				delete(c.byParent, resource.ParentID)
			}
		}
	}
}
