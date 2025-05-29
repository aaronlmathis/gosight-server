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

package memcache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aaronlmathis/gosight-server/internal/cache"
	"github.com/aaronlmathis/gosight-shared/model"
	"github.com/bradfitz/gomemcache/memcache"
)

type MemcacheResourceCache struct {
	client *memcache.Client
	prefix string
	ttl    int32 // TTL in seconds
}

func NewMemcacheResourceCache(client *memcache.Client, prefix string, ttl time.Duration) cache.ResourceCache {
	return &MemcacheResourceCache{
		client: client,
		prefix: prefix,
		ttl:    int32(ttl.Seconds()),
	}
}

func (m *MemcacheResourceCache) UpsertResource(resource *model.Resource) {
	key := m.resourceKey(resource.ID)

	data, err := json.Marshal(resource)
	if err != nil {
		// Log error
		return
	}

	item := &memcache.Item{
		Key:        key,
		Value:      data,
		Expiration: m.ttl,
	}

	m.client.Set(item)

	// Update indexes (simplified for memcache - would need external index store)
	m.updateIndexes(resource)
}

func (m *MemcacheResourceCache) GetResource(id string) (*model.Resource, bool) {
	key := m.resourceKey(id)

	item, err := m.client.Get(key)
	if err != nil {
		return nil, false
	}

	var resource model.Resource
	if err := json.Unmarshal(item.Value, &resource); err != nil {
		return nil, false
	}

	return &resource, true
}

func (m *MemcacheResourceCache) DeleteResource(id string) bool {
	key := m.resourceKey(id)

	// Get resource first to clean up indexes
	if resource, exists := m.GetResource(id); exists {
		m.removeIndexes(resource)
	}

	err := m.client.Delete(key)
	return err == nil
}

func (m *MemcacheResourceCache) GetResourcesByKind(kind string) []*model.Resource {
	indexKey := m.kindIndexKey(kind)

	item, err := m.client.Get(indexKey)
	if err != nil {
		return nil
	}

	var ids []string
	if err := json.Unmarshal(item.Value, &ids); err != nil {
		return nil
	}

	return m.getResourcesByIds(ids)
}

func (m *MemcacheResourceCache) GetResourcesByGroup(group string) []*model.Resource {
	indexKey := m.groupIndexKey(group)

	item, err := m.client.Get(indexKey)
	if err != nil {
		return nil
	}

	var ids []string
	if err := json.Unmarshal(item.Value, &ids); err != nil {
		return nil
	}

	return m.getResourcesByIds(ids)
}

func (m *MemcacheResourceCache) GetResourcesByLabels(labels map[string]string) []*model.Resource {
	if len(labels) == 0 {
		return nil
	}

	// For memcache, we need to manually intersect the sets
	var candidateIds map[string]bool
	first := true

	for key, value := range labels {
		indexKey := m.labelIndexKey(key, value)

		item, err := m.client.Get(indexKey)
		if err != nil {
			return nil
		}

		var ids []string
		if err := json.Unmarshal(item.Value, &ids); err != nil {
			return nil
		}

		if first {
			candidateIds = make(map[string]bool)
			for _, id := range ids {
				candidateIds[id] = true
			}
			first = false
		} else {
			// Intersect with existing candidates
			newCandidates := make(map[string]bool)
			for _, id := range ids {
				if candidateIds[id] {
					newCandidates[id] = true
				}
			}
			candidateIds = newCandidates
		}

		if len(candidateIds) == 0 {
			return nil
		}
	}

	var finalIds []string
	for id := range candidateIds {
		finalIds = append(finalIds, id)
	}

	return m.getResourcesByIds(finalIds)
}

func (m *MemcacheResourceCache) GetResourcesByTags(tags map[string]string) []*model.Resource {
	if len(tags) == 0 {
		return nil
	}

	// Similar to labels, manually intersect
	var candidateIds map[string]bool
	first := true

	for key, value := range tags {
		indexKey := m.tagIndexKey(key, value)

		item, err := m.client.Get(indexKey)
		if err != nil {
			return nil
		}

		var ids []string
		if err := json.Unmarshal(item.Value, &ids); err != nil {
			return nil
		}

		if first {
			candidateIds = make(map[string]bool)
			for _, id := range ids {
				candidateIds[id] = true
			}
			first = false
		} else {
			// Intersect with existing candidates
			newCandidates := make(map[string]bool)
			for _, id := range ids {
				if candidateIds[id] {
					newCandidates[id] = true
				}
			}
			candidateIds = newCandidates
		}

		if len(candidateIds) == 0 {
			return nil
		}
	}

	var finalIds []string
	for id := range candidateIds {
		finalIds = append(finalIds, id)
	}

	return m.getResourcesByIds(finalIds)
}

func (m *MemcacheResourceCache) GetResourcesByParent(parentID string) []*model.Resource {
	indexKey := m.parentIndexKey(parentID)

	item, err := m.client.Get(indexKey)
	if err != nil {
		return nil
	}

	var ids []string
	if err := json.Unmarshal(item.Value, &ids); err != nil {
		return nil
	}

	return m.getResourcesByIds(ids)
}

func (m *MemcacheResourceCache) UpdateLastSeen(id string, lastSeen time.Time) {
	if resource, exists := m.GetResource(id); exists {
		resource.LastSeen = lastSeen
		resource.Updated = true
		m.UpsertResource(resource)
	}
}

func (m *MemcacheResourceCache) UpdateStatus(id string, status string) {
	if resource, exists := m.GetResource(id); exists {
		resource.Status = status
		resource.Updated = true
		m.UpsertResource(resource)
	}
}

func (m *MemcacheResourceCache) GetStaleResources(threshold time.Duration) []*model.Resource {
	// This would require scanning all resources, which is expensive in Memcache
	// Return empty slice for now
	return []*model.Resource{}
}

func (m *MemcacheResourceCache) GetResourceCount() int {
	// Memcache doesn't have a native way to count keys by pattern
	// Would need to maintain a separate counter
	return 0
}

func (m *MemcacheResourceCache) GetResourceCountByKind() map[string]int {
	// Similar issue - would need separate counters per kind
	return make(map[string]int)
}

func (m *MemcacheResourceCache) Clear() {
	// Memcache doesn't support pattern deletion
	// This would require keeping track of all keys separately
	// For now, just flush all (if supported by the client setup)
	m.client.FlushAll()
}

func (m *MemcacheResourceCache) Stop() {
	// Memcache client doesn't need explicit stopping
}

func (m *MemcacheResourceCache) RemoveResource(id string) bool {
	return m.DeleteResource(id)
}

func (m *MemcacheResourceCache) GetSummary() map[string]interface{} {
	summary := make(map[string]interface{})
	summary["total_resources"] = m.GetResourceCount()
	summary["resource_count_by_kind"] = m.GetResourceCountByKind()
	summary["cache_type"] = "memcache"
	return summary
}

func (m *MemcacheResourceCache) GetKinds() []string {
	// For memcache, we can't efficiently get all kinds without scanning
	// This is a limitation of the memcache implementation
	// Return empty slice for now
	return []string{}
}

// Helper methods

func (m *MemcacheResourceCache) resourceKey(id string) string {
	return fmt.Sprintf("%s:resource:%s", m.prefix, id)
}

func (m *MemcacheResourceCache) kindIndexKey(kind string) string {
	return fmt.Sprintf("%s:index:kind:%s", m.prefix, kind)
}

func (m *MemcacheResourceCache) groupIndexKey(group string) string {
	return fmt.Sprintf("%s:index:group:%s", m.prefix, group)
}

func (m *MemcacheResourceCache) labelIndexKey(key, value string) string {
	return fmt.Sprintf("%s:index:label:%s:%s", m.prefix, key, value)
}

func (m *MemcacheResourceCache) tagIndexKey(key, value string) string {
	return fmt.Sprintf("%s:index:tag:%s:%s", m.prefix, key, value)
}

func (m *MemcacheResourceCache) parentIndexKey(parentID string) string {
	return fmt.Sprintf("%s:index:parent:%s", m.prefix, parentID)
}

func (m *MemcacheResourceCache) getResourcesByIds(ids []string) []*model.Resource {
	if len(ids) == 0 {
		return nil
	}

	var resources []*model.Resource
	for _, id := range ids {
		if resource, exists := m.GetResource(id); exists {
			resources = append(resources, resource)
		}
	}

	return resources
}

func (m *MemcacheResourceCache) updateIndexes(resource *model.Resource) {
	// Update kind index
	m.updateListIndex(m.kindIndexKey(resource.Kind), resource.ID, true)

	// Update group index
	if resource.Group != "" {
		m.updateListIndex(m.groupIndexKey(resource.Group), resource.ID, true)
	}

	// Update label indexes
	for key, value := range resource.Labels {
		m.updateListIndex(m.labelIndexKey(key, value), resource.ID, true)
	}

	// Update tag indexes
	for key, value := range resource.Tags {
		m.updateListIndex(m.tagIndexKey(key, value), resource.ID, true)
	}

	// Update parent index
	if resource.ParentID != "" {
		m.updateListIndex(m.parentIndexKey(resource.ParentID), resource.ID, true)
	}
}

func (m *MemcacheResourceCache) removeIndexes(resource *model.Resource) {
	// Remove from kind index
	m.updateListIndex(m.kindIndexKey(resource.Kind), resource.ID, false)

	// Remove from group index
	if resource.Group != "" {
		m.updateListIndex(m.groupIndexKey(resource.Group), resource.ID, false)
	}

	// Remove from label indexes
	for key, value := range resource.Labels {
		m.updateListIndex(m.labelIndexKey(key, value), resource.ID, false)
	}

	// Remove from tag indexes
	for key, value := range resource.Tags {
		m.updateListIndex(m.tagIndexKey(key, value), resource.ID, false)
	}

	// Remove from parent index
	if resource.ParentID != "" {
		m.updateListIndex(m.parentIndexKey(resource.ParentID), resource.ID, false)
	}
}

func (m *MemcacheResourceCache) updateListIndex(indexKey, id string, add bool) {
	item, err := m.client.Get(indexKey)
	var ids []string

	if err == nil {
		json.Unmarshal(item.Value, &ids)
	}

	if add {
		// Add id if not already present
		found := false
		for _, existingId := range ids {
			if existingId == id {
				found = true
				break
			}
		}
		if !found {
			ids = append(ids, id)
		}
	} else {
		// Remove id
		var newIds []string
		for _, existingId := range ids {
			if existingId != id {
				newIds = append(newIds, existingId)
			}
		}
		ids = newIds
	}

	// Store updated list
	data, _ := json.Marshal(ids)
	updatedItem := &memcache.Item{
		Key:        indexKey,
		Value:      data,
		Expiration: m.ttl,
	}
	m.client.Set(updatedItem)
}

// WarmCache is a placeholder implementation for Memcache cache warming.
// TODO: Implement cache warming for Memcache by adding store parameter to constructor
func (m *MemcacheResourceCache) WarmCache(ctx context.Context) error {
	// Memcache cache warming would require a store reference
	// For now, return nil (no-op)
	return nil
}
