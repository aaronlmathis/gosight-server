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

package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/aaronlmathis/gosight-server/internal/cache"
	"github.com/aaronlmathis/gosight-shared/model"
	"github.com/go-redis/redis/v8"
)

type RedisResourceCache struct {
	client *redis.Client
	prefix string
	ttl    time.Duration
}

func NewRedisResourceCache(client *redis.Client, prefix string, ttl time.Duration) cache.ResourceCache {
	return &RedisResourceCache{
		client: client,
		prefix: prefix,
		ttl:    ttl,
	}
}

func (r *RedisResourceCache) UpsertResource(resource *model.Resource) {
	ctx := context.Background()
	key := r.resourceKey(resource.ID)

	data, err := json.Marshal(resource)
	if err != nil {
		// Log error
		return
	}

	// Store resource
	r.client.Set(ctx, key, data, r.ttl)

	// Update indexes
	r.updateIndexes(resource)
}

func (r *RedisResourceCache) GetResource(id string) (*model.Resource, bool) {
	ctx := context.Background()
	key := r.resourceKey(id)

	data, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return nil, false
	}

	var resource model.Resource
	if err := json.Unmarshal([]byte(data), &resource); err != nil {
		return nil, false
	}

	return &resource, true
}

func (r *RedisResourceCache) DeleteResource(id string) bool {
	ctx := context.Background()
	key := r.resourceKey(id)

	// Get resource first to clean up indexes
	if resource, exists := r.GetResource(id); exists {
		r.removeIndexes(resource)
	}

	result := r.client.Del(ctx, key)
	return result.Val() > 0
}

func (r *RedisResourceCache) GetResourcesByKind(kind string) []*model.Resource {
	ctx := context.Background()
	indexKey := r.kindIndexKey(kind)

	ids, err := r.client.SMembers(ctx, indexKey).Result()
	if err != nil {
		return nil
	}

	return r.getResourcesByIds(ids)
}

func (r *RedisResourceCache) GetResourcesByGroup(group string) []*model.Resource {
	ctx := context.Background()
	indexKey := r.groupIndexKey(group)

	ids, err := r.client.SMembers(ctx, indexKey).Result()
	if err != nil {
		return nil
	}

	return r.getResourcesByIds(ids)
}

func (r *RedisResourceCache) GetResourcesByLabels(labels map[string]string) []*model.Resource {
	if len(labels) == 0 {
		return nil
	}

	ctx := context.Background()
	var keys []string

	for key, value := range labels {
		indexKey := r.labelIndexKey(key, value)
		keys = append(keys, indexKey)
	}

	// Intersect all label sets
	var ids []string
	if len(keys) == 1 {
		ids, _ = r.client.SMembers(ctx, keys[0]).Result()
	} else {
		ids, _ = r.client.SInter(ctx, keys...).Result()
	}

	return r.getResourcesByIds(ids)
}

func (r *RedisResourceCache) GetResourcesByTags(tags map[string]string) []*model.Resource {
	if len(tags) == 0 {
		return nil
	}

	ctx := context.Background()
	var keys []string

	for key, value := range tags {
		indexKey := r.tagIndexKey(key, value)
		keys = append(keys, indexKey)
	}

	// Intersect all tag sets
	var ids []string
	if len(keys) == 1 {
		ids, _ = r.client.SMembers(ctx, keys[0]).Result()
	} else {
		ids, _ = r.client.SInter(ctx, keys...).Result()
	}

	return r.getResourcesByIds(ids)
}

func (r *RedisResourceCache) GetResourcesByParent(parentID string) []*model.Resource {
	ctx := context.Background()
	indexKey := r.parentIndexKey(parentID)

	ids, err := r.client.SMembers(ctx, indexKey).Result()
	if err != nil {
		return nil
	}

	return r.getResourcesByIds(ids)
}

func (r *RedisResourceCache) UpdateLastSeen(id string, lastSeen time.Time) {
	if resource, exists := r.GetResource(id); exists {
		resource.LastSeen = lastSeen
		resource.Updated = true
		r.UpsertResource(resource)
	}
}

func (r *RedisResourceCache) UpdateStatus(id string, status string) {
	if resource, exists := r.GetResource(id); exists {
		resource.Status = status
		resource.Updated = true
		r.UpsertResource(resource)
	}
}

func (r *RedisResourceCache) GetStaleResources(threshold time.Duration) []*model.Resource {
	// This would require scanning all resources, which is expensive in Redis
	// In practice, you might want to maintain a separate sorted set by LastSeen
	// For now, return empty slice
	return []*model.Resource{}
}

func (r *RedisResourceCache) GetResourceCount() int {
	ctx := context.Background()
	pattern := r.prefix + ":resource:*"

	keys, err := r.client.Keys(ctx, pattern).Result()
	if err != nil {
		return 0
	}

	return len(keys)
}

func (r *RedisResourceCache) GetResourceCountByKind() map[string]int {
	ctx := context.Background()
	pattern := r.prefix + ":index:kind:*"

	keys, err := r.client.Keys(ctx, pattern).Result()
	if err != nil {
		return nil
	}

	counts := make(map[string]int)
	for _, key := range keys {
		// Extract kind from key
		parts := strings.Split(key, ":")
		if len(parts) >= 4 {
			kind := parts[3]
			count, _ := r.client.SCard(ctx, key).Result()
			counts[kind] = int(count)
		}
	}

	return counts
}

func (r *RedisResourceCache) Clear() {
	ctx := context.Background()
	pattern := r.prefix + ":*"

	keys, err := r.client.Keys(ctx, pattern).Result()
	if err != nil {
		return
	}

	if len(keys) > 0 {
		r.client.Del(ctx, keys...)
	}
}

func (r *RedisResourceCache) Stop() {
	// Redis client doesn't need explicit stopping for cache operations
}

func (r *RedisResourceCache) RemoveResource(id string) bool {
	return r.DeleteResource(id)
}

func (r *RedisResourceCache) GetSummary() map[string]interface{} {
	summary := make(map[string]interface{})
	summary["total_resources"] = r.GetResourceCount()
	summary["resource_count_by_kind"] = r.GetResourceCountByKind()
	summary["cache_type"] = "redis"
	return summary
}

func (r *RedisResourceCache) GetKinds() []string {
	ctx := context.Background()
	pattern := fmt.Sprintf("%s:index:kind:*", r.prefix)
	keys, err := r.client.Keys(ctx, pattern).Result()
	if err != nil {
		return []string{}
	}
	
	kinds := make([]string, 0, len(keys))
	prefixLen := len(fmt.Sprintf("%s:index:kind:", r.prefix))
	
	for _, key := range keys {
		if len(key) > prefixLen {
			kind := key[prefixLen:]
			kinds = append(kinds, kind)
		}
	}
	
	return kinds
}

// Helper methods

func (r *RedisResourceCache) resourceKey(id string) string {
	return fmt.Sprintf("%s:resource:%s", r.prefix, id)
}

func (r *RedisResourceCache) kindIndexKey(kind string) string {
	return fmt.Sprintf("%s:index:kind:%s", r.prefix, kind)
}

func (r *RedisResourceCache) groupIndexKey(group string) string {
	return fmt.Sprintf("%s:index:group:%s", r.prefix, group)
}

func (r *RedisResourceCache) labelIndexKey(key, value string) string {
	return fmt.Sprintf("%s:index:label:%s:%s", r.prefix, key, value)
}

func (r *RedisResourceCache) tagIndexKey(key, value string) string {
	return fmt.Sprintf("%s:index:tag:%s:%s", r.prefix, key, value)
}

func (r *RedisResourceCache) parentIndexKey(parentID string) string {
	return fmt.Sprintf("%s:index:parent:%s", r.prefix, parentID)
}

func (r *RedisResourceCache) getResourcesByIds(ids []string) []*model.Resource {
	if len(ids) == 0 {
		return nil
	}

	var resources []*model.Resource
	for _, id := range ids {
		if resource, exists := r.GetResource(id); exists {
			resources = append(resources, resource)
		}
	}

	return resources
}

func (r *RedisResourceCache) updateIndexes(resource *model.Resource) {
	ctx := context.Background()

	// Kind index
	kindKey := r.kindIndexKey(resource.Kind)
	r.client.SAdd(ctx, kindKey, resource.ID)
	r.client.Expire(ctx, kindKey, r.ttl)

	// Group index
	if resource.Group != "" {
		groupKey := r.groupIndexKey(resource.Group)
		r.client.SAdd(ctx, groupKey, resource.ID)
		r.client.Expire(ctx, groupKey, r.ttl)
	}

	// Label indexes
	for key, value := range resource.Labels {
		labelKey := r.labelIndexKey(key, value)
		r.client.SAdd(ctx, labelKey, resource.ID)
		r.client.Expire(ctx, labelKey, r.ttl)
	}

	// Tag indexes
	for key, value := range resource.Tags {
		tagKey := r.tagIndexKey(key, value)
		r.client.SAdd(ctx, tagKey, resource.ID)
		r.client.Expire(ctx, tagKey, r.ttl)
	}

	// Parent index
	if resource.ParentID != "" {
		parentKey := r.parentIndexKey(resource.ParentID)
		r.client.SAdd(ctx, parentKey, resource.ID)
		r.client.Expire(ctx, parentKey, r.ttl)
	}
}

func (r *RedisResourceCache) removeIndexes(resource *model.Resource) {
	ctx := context.Background()

	// Kind index
	kindKey := r.kindIndexKey(resource.Kind)
	r.client.SRem(ctx, kindKey, resource.ID)

	// Group index
	if resource.Group != "" {
		groupKey := r.groupIndexKey(resource.Group)
		r.client.SRem(ctx, groupKey, resource.ID)
	}

	// Label indexes
	for key, value := range resource.Labels {
		labelKey := r.labelIndexKey(key, value)
		r.client.SRem(ctx, labelKey, resource.ID)
	}

	// Tag indexes
	for key, value := range resource.Tags {
		tagKey := r.tagIndexKey(key, value)
		r.client.SRem(ctx, tagKey, resource.ID)
	}

	// Parent index
	if resource.ParentID != "" {
		parentKey := r.parentIndexKey(resource.ParentID)
		r.client.SRem(ctx, parentKey, resource.ID)
	}
}
