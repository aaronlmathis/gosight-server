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

// Package cache provides a unified caching system for the GoSight server application.
// It implements high-performance in-memory caching for various types of telemetry
// data including processes, metrics, tags, logs, and resources.
//
// The cache system is designed to optimize data retrieval and reduce latency
// in the monitoring pipeline. It supports multiple backend implementations
// including in-memory, Redis, and Memcache for different deployment scenarios.
//
// Key Features:
//   - Process caching for system process monitoring data
//   - Metric caching with tag-based indexing and fast lookups
//   - Log caching with timestamp-based retention policies
//   - Resource caching with hierarchical relationships and metadata
//   - Tag caching for efficient label and dimension queries
//   - Thread-safe operations with concurrent read/write support
//   - Configurable TTL (time-to-live) policies per cache type
//   - Memory usage monitoring and automatic cleanup
//   - Health check endpoints for cache status monitoring
//
// Cache Backends:
//   - InMemory: High-performance local caching with LRU eviction
//   - Redis: Distributed caching with persistence and clustering support
//   - Memcache: Distributed memory caching for large-scale deployments
//
// The cache system integrates seamlessly with the GoSight telemetry pipeline,
// providing fast data access for dashboards, alerting, and analytics workloads.
package cache

import (
	"time"

	"github.com/aaronlmathis/gosight-shared/model"
)

// Cache represents the unified caching system for the GoSight server.
// It provides centralized access to all cache types used throughout the application,
// including process monitoring data, metrics, logs, tags, and resource metadata.
//
// The Cache struct acts as a facade that organizes different cache implementations
// by data type, enabling efficient data retrieval and reducing database load.
// Each cache component is optimized for its specific data patterns and access requirements.
//
// Cache Components:
//   - Processes: Caches system process information and resource usage
//   - Metrics: Stores time-series metrics with tag-based indexing
//   - Tags: Provides fast lookups for metric labels and dimensions
//   - Logs: Buffers log entries with timestamp-based organization
//   - Resources: Caches resource hierarchy and metadata information
//
// Future expansion may include additional cache types for agents, endpoints,
// alerts, and events as the system evolves.
type Cache struct {
	Processes ProcessCache
	Metrics   MetricCache
	Tags      TagCache
	Logs      LogCache
	Resources ResourceCache
	/*
		Agents    AgentCache
		Endpoints EndpointCache
		Alerts    AlertCache
		Events    EventCache
	*/
}

// ResourceCache defines the interface for caching resource objects in the GoSight system.
// Resources represent monitored entities such as agents, hosts, containers, services,
// and endpoints in the infrastructure. The cache provides fast access to resource
// metadata, relationships, and status information.
//
// The interface supports both basic CRUD operations and advanced query capabilities
// including hierarchical lookups, label-based filtering, and status monitoring.
// This enables efficient resource discovery, dependency mapping, and health tracking
// across the entire monitored infrastructure.
//
// Key Operations:
//   - Core CRUD: Create, read, update, and delete resource entries
//   - Query by Kind: Retrieve resources by type (host, container, service, etc.)
//   - Query by Group: Group-based resource organization and filtering
//   - Label Filtering: Find resources matching specific label criteria
//   - Tag Searching: Advanced tag-based resource discovery
//   - Hierarchy Navigation: Parent-child relationship traversal
//   - Status Tracking: Monitor resource health and availability
//   - Staleness Detection: Identify resources that haven't reported recently
//
// The cache implementation handles resource lifecycle management, automatic
// cleanup of stale entries, and maintains indexes for efficient querying.
type ResourceCache interface {
	// Core operations
	UpsertResource(resource *model.Resource)
	GetResource(id string) (*model.Resource, bool)
	DeleteResource(id string) bool

	// Query operations
	GetResourcesByKind(kind string) []*model.Resource
	GetResourcesByGroup(group string) []*model.Resource
	GetResourcesByLabels(labels map[string]string) []*model.Resource
	GetResourcesByTags(tags map[string]string) []*model.Resource
	GetResourcesByParent(parentID string) []*model.Resource

	// Status operations
	UpdateLastSeen(id string, lastSeen time.Time)
	UpdateStatus(id string, status string)

	// Health operations
	GetStaleResources(threshold time.Duration) []*model.Resource
	GetResourceCount() int
	GetResourceCountByKind() map[string]int

	// Cache management
	RemoveResource(id string) bool
	GetSummary() map[string]interface{}
	GetKinds() []string
	Clear()
	Stop()
}
