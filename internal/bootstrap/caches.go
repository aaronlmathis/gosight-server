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
	"fmt"

	"github.com/aaronlmathis/gosight-server/internal/cache"
	"github.com/aaronlmathis/gosight-server/internal/cache/resourcecache"
	"github.com/aaronlmathis/gosight-server/internal/config"
	"github.com/aaronlmathis/gosight-server/internal/store/resourcestore"
)

// InitCaches initializes all caching components for the GoSight system.
// This function sets up various cache types including metric cache, tag cache,
// process cache, log cache, and resource cache based on the configuration.
//
// The caches improve system performance by storing frequently accessed data
// in memory, reducing the need for expensive database queries. Different cache
// types serve specific purposes:
//   - Metric Cache: Stores metric metadata and time-series data
//   - Tag Cache: Caches resource tags for fast filtering and search
//   - Process Cache: Stores process information for monitoring
//   - Log Cache: Buffers log entries for efficient processing
//   - Resource Cache: Caches resource metadata and relationships
//
// Parameters:
//   - ctx: Context for the initialization process
//   - cfg: Configuration containing cache settings and engine type
//   - resourceStore: Resource store for cache population and refresh
//
// Returns:
//   - *cache.Cache: Initialized cache container with all cache types
//   - error: If cache initialization fails for any component
func InitCaches(ctx context.Context, cfg *config.Config, resourceStore resourcestore.ResourceStore) (*cache.Cache, error) {
	caches := &cache.Cache{}

	// Initialize Metric Cache
	metricCache := cache.NewMetricCache()
	caches.Metrics = metricCache

	// Initialize tag cache
	tagCache := cache.NewTagCache()

	// Warm fill tag cache
	//tags, err := resourceStore.GetAllTags(ctx)
	//if err != nil {
	//	return nil, fmt.Errorf("failed to load tags for cache: %w", err)
	//}

	//tagCache.LoadFromStore(tags)

	caches.Tags = tagCache

	// Initialize Process Cache
	processCache := cache.NewProcessCache()
	caches.Processes = processCache

	// Log Cache

	logCache := cache.NewLogCache()
	caches.Logs = logCache

	switch cfg.Cache.Engine {
	//case "redis":
	//	resourceCache = resourcecache.NewRedisResourceCache()
	//case "memcache":
	//	resourceCache = resourcecache.NewMemcacheResourceCache()
	case "memory":
		caches.Resources = resourcecache.NewInMemoryResourceCache(resourceStore, cfg.Cache.ResourceFlushInterval)
	default:
		return nil, fmt.Errorf("unsupported cache provider: %s", cfg.Cache.Engine)
	}

	// Warm the resource cache with existing data from store
	if err := caches.Resources.WarmCache(ctx); err != nil {
		return nil, fmt.Errorf("failed to warm resource cache: %w", err)
	}

	return caches, nil
}
