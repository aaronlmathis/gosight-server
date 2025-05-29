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

// Package cache provides a unified cache for the GoSight server.
// It includes caches for processes, metrics, tags, logs, and other components.
package cache

import (
	"strings"
	"sync"
	"time"

	"github.com/aaronlmathis/gosight-shared/model"
	"github.com/aaronlmathis/gosight-shared/utils"
)

// MetricCache provides high-performance in-memory caching for time-series metrics
// and their associated metadata. It serves as the primary interface for storing,
// retrieving, and querying metric data in the GoSight monitoring system.
//
// The cache is designed to handle high-throughput metric ingestion while providing
// fast query capabilities for dashboards, alerting, and analytics. It maintains
// comprehensive indexes for metrics, namespaces, dimensions, and tags to enable
// efficient data discovery and filtering operations.
//
// Key Features:
//   - Thread-safe concurrent operations with optimized locking strategies
//   - Hierarchical namespace organization (namespace -> subnamespace -> metric)
//   - Multi-dimensional tag indexing for complex filtering queries
//   - Automatic dimension discovery and metadata extraction
//   - Time-based data retention and cleanup policies
//   - Memory-efficient storage with configurable limits
//   - Fast namespace and metric name enumeration
//   - Advanced tag-based metric discovery and filtering
//
// The cache supports the full metric lifecycle from ingestion to query, providing
// the foundation for real-time monitoring and analytics capabilities.
//
// Architecture:
//   - Namespace hierarchy enables logical grouping of related metrics
//   - Tag indexes support complex multi-dimensional queries
//   - Dimension tracking enables dynamic dashboard and query building
//   - Memory management prevents unbounded cache growth
type MetricCache interface {
	// Add ingests a metric payload into the cache, extracting and indexing all
	// metadata including namespace, subnamespace, metric names, dimensions, and tags.
	// This method is optimized for high-throughput ingestion scenarios.
	//
	// Parameters:
	//   - payload: Complete metric payload containing metrics and metadata
	Add(payload *model.MetricPayload)

	// GetNamespaces returns all known metric namespaces in the cache. Namespaces
	// provide the top-level organization for metrics (e.g., "System", "Application").
	//
	// Returns:
	//   - []string: Slice of all namespace names
	GetNamespaces() []string

	// GetSubNamespaces retrieves all subnamespaces within a specific namespace.
	// Subnamespaces provide secondary organization (e.g., "CPU", "Memory" under "System").
	//
	// Parameters:
	//   - nameSpace: The parent namespace to query
	//
	// Returns:
	//   - []string: Slice of subnamespace names within the namespace
	GetSubNamespaces(nameSpace string) []string

	// GetMetricNames returns all metric names within a specific namespace and subnamespace.
	// This enables discovery of available metrics for dashboard configuration.
	//
	// Parameters:
	//   - nameSpace: The namespace containing the metrics
	//   - subNamespace: The subnamespace containing the metrics
	//
	// Returns:
	//   - []string: Slice of metric names in the specified hierarchy
	GetMetricNames(nameSpace, subNamespace string) []string

	// GetAllMetricNames returns every metric name known to the cache across all
	// namespaces. Useful for global metric discovery and validation.
	//
	// Returns:
	//   - []string: Complete list of all metric names
	GetAllMetricNames() []string

	// GetAvailableDimensions returns all known dimensions (labels) organized by
	// dimension key. This enables dynamic query building and filter construction.
	//
	// Returns:
	//   - map[string][]string: Map of dimension keys to their possible values
	GetAvailableDimensions() map[string][]string

	// GetMetricDimensions returns all dimension keys associated with a specific metric.
	// This helps in understanding what dimensions are available for filtering.
	//
	// Parameters:
	//   - metricName: The metric to query for dimensions
	//
	// Returns:
	//   - []string: Slice of dimension keys for the metric
	GetMetricDimensions(metricName string) []string

	// GetAllTagKeys returns all tag keys present in the cache. Tags provide
	// additional metadata beyond standard dimensions.
	//
	// Returns:
	//   - []string: Complete list of all tag keys
	GetAllTagKeys() []string

	// GetAllTagValuesForKey returns all known values for a specific tag key.
	// This supports tag-based filtering and validation.
	//
	// Parameters:
	//   - key: The tag key to query
	//
	// Returns:
	//   - []string: All values associated with the tag key
	GetAllTagValuesForKey(key string) []string

	// GetAllKnownLabelValues retrieves all values for a specified label key across
	// both dimensions and tags. Optionally filters results containing a substring.
	//
	// Parameters:
	//   - label: The label key to search for
	//   - contains: Optional substring filter (empty string returns all values)
	//
	// Returns:
	//   - []string: All label values matching the criteria
	GetAllKnownLabelValues(label, contains string) []string

	// GetLabelValues returns all values for a known label key with optional filtering.
	// This method focuses on established labels with confirmed existence.
	//
	// Parameters:
	//   - label: The established label key to query
	//   - contains: Optional substring filter for value matching
	//
	// Returns:
	//   - []string: Filtered label values for the specified key
	GetLabelValues(label, contains string) []string
	GetMetricsWithLabels(filters map[string]string) []string // Get all metric names that match a given label filter

	Prune()

	GetAllEntries() []*MetricEntry
}

// MetricEntry represents a metric entry in the cache.
// It contains information about the metric, including its namespace, subnamespace,
// name, unit, type, dimensions, labels, tags, and emitters.
// The dimensions and labels are stored as maps of string sets, allowing for efficient
// membership testing and uniqueness.
type MetricEntry struct {
	Namespace string
	SubNS     string
	Name      string
	Unit      string
	Type      string

	Dimensions map[string]StringSet // Dimensions are metric specific key value pairs, added by collectors
	Labels     map[string]StringSet // Labels are a superset of dimensions, including custom user-defined tags and meta fields
	Tags       map[string]StringSet // Tags are user-defined key value pairs, added by the user in the Meta
	Emitters   map[string]struct{}  // endpoint IDs
}

// metricCache is a struct that implements the MetricCache interface.
// It uses a map to store metric entries, where the key is the full metric name
// (namespace + subnamespace + name) and the value is the MetricEntry.
// The cache is protected by a mutex to ensure thread safety.
type metricCache struct {
	mu sync.RWMutex

	Namespaces    map[string]struct{}
	SubNamespaces map[string]map[string]struct{} // ns → subns
	MetricEntries map[string]*MetricEntry        // fullName → entry
	LabelValues   map[string]StringSet           // label key → all values
	EndpointMeta  map[string]*model.Meta         // endpoint_id → latest Meta
	LastSeen      map[string]int64
}

// NewMetricCache creates a new instance of MetricCache.
// It initializes the cache with empty maps for namespaces, subnamespaces,
// metric entries, label values, endpoint metadata, and last seen timestamps.
// The cache is designed to be thread-safe and allows concurrent access.
func NewMetricCache() MetricCache {
	return &metricCache{

		Namespaces:    make(map[string]struct{}),
		SubNamespaces: make(map[string]map[string]struct{}),
		MetricEntries: make(map[string]*MetricEntry),
		LabelValues:   make(map[string]StringSet),
		EndpointMeta:  make(map[string]*model.Meta),
		LastSeen:      make(map[string]int64),
	}
}

// Add adds a new metric payload to the cache.
// It updates the metric entries, namespaces, subnamespaces, label values,
// and endpoint metadata based on the provided payload.
func (c *metricCache) Add(payload *model.MetricPayload) {
	c.mu.Lock()
	defer c.mu.Unlock()

	eid := payload.EndpointID
	if payload.Meta != nil {
		c.EndpointMeta[eid] = payload.Meta
	}
	c.LastSeen[eid] = time.Now().Unix()
	for _, m := range payload.Metrics {

		normalNS := strings.ToLower(m.Namespace)
		normalSN := strings.ToLower(m.SubNamespace)
		normalMN := strings.ToLower(m.Name)

		fullName := normalNS + "." + normalSN + "." + normalMN

		if _, ok := c.Namespaces[normalNS]; !ok {
			c.Namespaces[normalNS] = struct{}{}
		}
		if _, ok := c.SubNamespaces[normalNS]; !ok {
			c.SubNamespaces[normalNS] = make(map[string]struct{})
			utils.Debug("Adding new subnamespace: %s.%s", normalNS, normalSN)
		}
		normalSn := strings.ToLower(normalSN)
		c.SubNamespaces[normalNS][normalSn] = struct{}{}

		entry := c.MetricEntries[fullName]
		if entry == nil {
			entry = &MetricEntry{
				Namespace:  normalNS,
				SubNS:      normalSN,
				Name:       normalMN,
				Unit:       strings.ToLower(m.Unit),
				Type:       strings.ToLower(m.Type),
				Dimensions: make(map[string]StringSet),
				Tags:       make(map[string]StringSet),
				Labels:     make(map[string]StringSet),
				Emitters:   make(map[string]struct{}),
			}
			c.MetricEntries[fullName] = entry
		}

		entry.Emitters[eid] = struct{}{}

		// Index dimensions
		for k, v := range m.Dimensions {
			addToSet(entry.Dimensions, k, v)
			addToSet(entry.Labels, k, v)
			addToSet(c.LabelValues, k, v)
		}

		// Index tags / Meta into Labels as well.
		AddMetaFieldsToLabels(payload.Meta, entry.Labels)

	}
}

// GetNamespaces returns a slice of all known namespaces in the cache.
// It iterates over the namespaces map and collects the keys into a slice.
// The function is thread-safe and uses a read lock to ensure concurrent access.
func (c *metricCache) GetNamespaces() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var out []string
	for ns := range c.Namespaces {
		out = append(out, ns)
	}
	return out
}

// GetMetricNames returns a slice of all metric names for a given namespace and subnamespace.
// It checks if the namespace and subnamespace exist in the metric entries map
// and collects the metric names into a slice. The function is thread-safe and uses
// a read lock to ensure concurrent access.
func (c *metricCache) GetMetricNames(namespace, subNamespace string) []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var names []string
	for _, entry := range c.MetricEntries {
		if entry.Namespace == namespace && entry.SubNS == subNamespace {
			names = append(names, entry.Name)
		}
	}
	return names
}

// GetSubNamespaces returns a slice of all known subnamespaces for a given namespace.
// It checks if the namespace exists in the subnamespaces map and collects
// the keys into a slice. The function is thread-safe and uses a read lock
// to ensure concurrent access.
func (c *metricCache) GetSubNamespaces(namespace string) []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	subnsMap, ok := c.SubNamespaces[namespace]
	if !ok {
		return []string{}
	}

	var out []string
	for sub := range subnsMap {
		out = append(out, sub)
	}
	return out
}

// GetAllMetricNames returns a slice of all known metric names in the cache.
// It iterates over the metric entries map and collects the full names into a slice.
func (c *metricCache) GetAllMetricNames() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var names []string
	for fullName := range c.MetricEntries {
		names = append(names, fullName)
	}
	return names
}

// GetAvailableDimensions returns a map of all available dimensions in the cache.
// It iterates over the label values map and collects the keys and their corresponding
// values into a new map. The function only includes keys that are used as dimensions
// (not tags only). The function is thread-safe and uses a read lock to ensure concurrent access.
func (c *metricCache) GetAvailableDimensions() map[string][]string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	out := make(map[string][]string)
	for key, valueSet := range c.LabelValues {
		// Only include keys that are used as dimensions (not tags only)
		isDimension := false
		for _, entry := range c.MetricEntries {
			if _, ok := entry.Dimensions[key]; ok {
				isDimension = true
				break
			}
		}
		if isDimension {
			var values []string
			for v := range valueSet {
				values = append(values, v)
			}
			out[key] = values
		}
	}
	return out
}

// GetMetricDimensions returns a slice of all dimension keys for a given metric name.
// It checks if the metric name exists in the metric entries map and collects
// the dimension keys into a slice. The function is thread-safe and uses
// a read lock to ensure concurrent access.
func (c *metricCache) GetMetricDimensions(metricName string) []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, ok := c.MetricEntries[metricName]
	if !ok {
		return nil
	}

	var keys []string
	for k := range entry.Dimensions {
		keys = append(keys, k)
	}
	return keys
}

// GetAllTagKeys returns a slice of all known tag keys in the cache.
// It iterates over the EndpointMeta map and collects the keys into a slice.
// The function is thread-safe and uses a read lock to ensure concurrent access.
// It also ensures that the keys are unique by using a set.
func (c *metricCache) GetAllTagKeys() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Collect tag keys directly from EndpointMeta
	tagKeySet := make(map[string]struct{})

	for _, meta := range c.EndpointMeta {
		for k := range meta.Labels {
			tagKeySet[k] = struct{}{}
		}
	}

	var keys []string
	for k := range tagKeySet {
		keys = append(keys, k)
	}
	return keys
}

// GetAllTagValuesForKey returns a slice of all known tag values for a given key.
// It iterates over the EndpointMeta map and collects the values into a slice.
// The function is thread-safe and uses a read lock to ensure concurrent access.
func (c *metricCache) GetAllTagValuesForKey(key string) []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	valueSet := make(map[string]struct{})

	for _, meta := range c.EndpointMeta {
		if val, ok := meta.Labels[key]; ok {
			valueSet[val] = struct{}{}
		}
	}

	var values []string
	for v := range valueSet {
		values = append(values, v)
	}
	return values
}

// GetAllKnownLabelValues returns a slice of all known label values for a given label key.
// It iterates over the label values map and collects the values into a slice.
func (c *metricCache) GetAllKnownLabelValues(label, contains string) []string {
	return c.GetLabelValues(label, contains)
}

// GetLabelValues returns a slice of all label values for a given label key.
// It checks if the label key exists in the label values map and collects
// the values into a slice. The function also allows for filtering based on
// a substring match. If the contains parameter is empty, all values are returned.
func (c *metricCache) GetLabelValues(label, contains string) []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	values, ok := c.LabelValues[label]
	if !ok {
		return nil
	}

	var out []string
	for v := range values {
		if contains == "" || contains == v || containsMatch(v, contains) {
			out = append(out, v)
		}
	}
	return out
}

// GetMetricsWithLabels returns a slice of all metric names that match a given label filter.
// It iterates over the metric entries map and checks if the labels match the filters.
// The function allows for filtering based on multiple label key-value pairs.
func (c *metricCache) GetMetricsWithLabels(filters map[string]string) []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var matched []string

	for fullName, entry := range c.MetricEntries {
		match := true

		for key, expected := range filters {
			values, exists := entry.Labels[key]
			if !exists {
				match = false
				break
			}
			if _, valueExists := values[expected]; !valueExists {
				match = false
				break
			}
		}

		if match {
			matched = append(matched, fullName)
		}
	}

	return matched
}

// Prune removes stale entries from the cache.
// It identifies endpoints that have not been seen for a certain period
// (e.g., 10 minutes) and removes them from the cache.
func (c *metricCache) Prune() {
	c.mu.Lock()
	defer c.mu.Unlock()

	cutoff := time.Now().Add(-1440 * time.Minute).Unix() // e.g. 10 minutes (one day 1440 minutes)

	// Identify stale endpoints
	stale := make(map[string]struct{})
	for eid, ts := range c.LastSeen {
		if ts < cutoff {
			stale[eid] = struct{}{}
			delete(c.EndpointMeta, eid)
			delete(c.LastSeen, eid)
		}
	}

	// Remove emitters from metrics
	for fullName, entry := range c.MetricEntries {
		for eid := range stale {
			delete(entry.Emitters, eid)
		}
		// Optional: remove metrics with no active emitters
		if len(entry.Emitters) == 0 {
			delete(c.MetricEntries, fullName)
		}
	}
}

// GetAllEntries returns a slice of all metric entries in the cache.
// It iterates over the metric entries map and collects the entries into a slice.
func (c *metricCache) GetAllEntries() []*MetricEntry {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entries := make([]*MetricEntry, 0, len(c.MetricEntries))
	for _, entry := range c.MetricEntries {
		entries = append(entries, entry)
	}
	return entries
}

// AddMetaFieldsToLabels adds metadata fields to the labels map.
// It takes a pointer to a model.Meta object and a map of labels.
// The function adds various metadata fields (e.g., agent ID, host ID, etc.)
// to the labels map, ensuring that the keys and values are unique.
func AddMetaFieldsToLabels(meta *model.Meta, labels map[string]StringSet) {
	if meta == nil || labels == nil {
		return
	}

	add := func(key, val string) {
		if val == "" {
			return
		}
		if _, exists := labels[key]; !exists {
			labels[key] = StringSet{}
		}
		labels[key][val] = struct{}{}
	}

	// Agent Info
	add("agent_id", meta.AgentID)
	add("agent_version", meta.AgentVersion)

	// Host Info
	add("host_id", meta.HostID)
	add("endpoint_id", meta.EndpointID)
	add("hostname", meta.Hostname)
	add("ip_address", meta.IPAddress)
	add("os", meta.OS)
	add("os_version", meta.OSVersion)
	add("platform", meta.Platform)
	add("platform_family", meta.PlatformFamily)
	add("platform_version", meta.PlatformVersion)
	add("kernel_architecture", meta.KernelArchitecture)
	add("kernel_version", meta.KernelVersion)
	add("architecture", meta.Architecture)
	add("virtualization_system", meta.VirtualizationSystem)
	add("virtualization_role", meta.VirtualizationRole)

	// Cloud Info
	add("cloud_provider", meta.CloudProvider)
	add("region", meta.Region)
	add("availability_zone", meta.AvailabilityZone)
	add("instance_id", meta.InstanceID)
	add("instance_type", meta.InstanceType)
	add("account_id", meta.AccountID)
	add("project_id", meta.ProjectID)
	add("resource_group", meta.ResourceGroup)
	add("vpc_id", meta.VPCID)
	add("subnet_id", meta.SubnetID)
	add("image_id", meta.ImageID)
	add("service_id", meta.ServiceID)

	// Container / K8s
	add("container_id", meta.ContainerID)
	add("container_name", meta.ContainerName)
	add("pod_name", meta.PodName)
	add("namespace", meta.Namespace)
	add("cluster_name", meta.ClusterName)
	add("node_name", meta.NodeName)
	add("container_image_id", meta.ContainerImageID)
	add("container_image", meta.ContainerImageName)

	// App Info
	add("application", meta.Application)
	add("environment", meta.Environment)
	add("service", meta.Service)
	add("version", meta.Version)
	add("deployment_id", meta.DeploymentID)

	// Network Info
	add("public_ip", meta.PublicIP)
	add("private_ip", meta.PrivateIP)
	add("mac_address", meta.MACAddress)
	add("network_interface", meta.NetworkInterface)

	// Custom labels
	for k, v := range meta.Labels {
		if _, exists := labels[k]; !exists {
			labels[k] = StringSet{}
		}
		labels[k][v] = struct{}{}
	}
}
