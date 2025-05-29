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

package resourcecache

import (
	"time"

	"github.com/aaronlmathis/gosight-server/internal/cache"
	"github.com/aaronlmathis/gosight-server/internal/cache/inmemory"
	"github.com/aaronlmathis/gosight-server/internal/store/resourcestore"
)

// ResourceCache is a type alias for the cache.ResourceCache interface
// This allows other packages to reference the cache interface without
// importing the entire cache package
type ResourceCache = cache.ResourceCache

// NewInMemoryResourceCache creates a new in-memory resource cache implementation
func NewInMemoryResourceCache(store resourcestore.ResourceStore, flushInterval time.Duration) ResourceCache {
	return inmemory.NewResourceCache(store, flushInterval)
}

// NewRedisResourceCache creates a new Redis-based resource cache implementation
func NewRedisResourceCache(client interface{}, prefix string, ttl time.Duration) ResourceCache {
	// Import Redis package dynamically to avoid dependency issues
	// This would require proper Redis client type assertion
	return nil // Placeholder - would need proper Redis client
}

// NewMemcacheResourceCache creates a new Memcache-based resource cache implementation
func NewMemcacheResourceCache(client interface{}, prefix string, ttl time.Duration) ResourceCache {
	// Import Memcache package dynamically to avoid dependency issues
	// This would require proper Memcache client type assertion
	return nil // Placeholder - would need proper Memcache client
}
