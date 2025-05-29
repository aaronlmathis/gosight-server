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

package telemetry

import (
	"testing"
	"time"

	"github.com/aaronlmathis/gosight-server/internal/cache/inmemory"
	"github.com/aaronlmathis/gosight-server/internal/store/resourcestore/mockstore"
	"github.com/aaronlmathis/gosight-shared/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResourceDiscovery_CacheOnlyApproach(t *testing.T) {
	// Create mock store to verify no direct store calls are made
	mockStore := mockstore.NewMockResourceStore()

	// Create in-memory cache with short flush interval for testing
	cache := inmemory.NewResourceCache(mockStore, 100*time.Millisecond)
	defer cache.Stop()

	// Create resource discovery with cache-only approach
	rd := NewResourceDiscovery(cache)

	t.Run("ProcessMetricPayload creates resource in cache", func(t *testing.T) {
		payload := &model.MetricPayload{
			Meta: &model.Meta{
				AgentID:      "agent-123",
				HostID:       "host-456",
				Hostname:     "test-host",
				Kind:         model.ResourceKindHost,
				Environment:  "production",
				IPAddress:    "192.168.1.100",
				OS:           "linux",
				Platform:     "ubuntu",
				Architecture: "amd64",
			},
		}

		// Process the payload
		rd.ProcessMetricPayload(payload)

		// Verify resource was created in cache
		require.NotEmpty(t, payload.Meta.ResourceID, "ResourceID should be set after processing")

		resource, exists := cache.GetResource(payload.Meta.ResourceID)
		require.True(t, exists, "Resource should exist in cache")
		assert.Equal(t, model.ResourceKindHost, resource.Kind)
		assert.Equal(t, "test-host", resource.Name)
		assert.Equal(t, "production", resource.Environment)
		assert.Equal(t, "192.168.1.100", resource.IPAddress)
		assert.Equal(t, model.ResourceStatusOnline, resource.Status)

		// Verify no direct store calls were made during processing
		assert.Empty(t, mockStore.GetCalls(), "No direct store Get calls should be made")
		assert.Empty(t, mockStore.CreateCalls(), "No direct store Create calls should be made")
		assert.Empty(t, mockStore.UpdateCalls(), "No direct store Update calls should be made")
	})

	t.Run("UpdateExistingResource only uses cache", func(t *testing.T) {
		// Create initial resource
		initialMeta := &model.Meta{
			AgentID:     "agent-789",
			HostID:      "host-789",
			Hostname:    "existing-host",
			Kind:        model.ResourceKindHost,
			Environment: "staging",
			IPAddress:   "10.0.0.1",
		}

		payload1 := &model.MetricPayload{Meta: initialMeta}
		rd.ProcessMetricPayload(payload1)

		// Get the created resource
		resourceID := payload1.Meta.ResourceID
		originalResource, exists := cache.GetResource(resourceID)
		require.True(t, exists)
		originalLastSeen := originalResource.LastSeen

		// Wait a moment to ensure timestamp difference
		time.Sleep(10 * time.Millisecond)

		// Update with new metadata
		updateMeta := &model.Meta{
			AgentID:     "agent-789",
			HostID:      "host-789",
			Hostname:    "existing-host",
			Kind:        model.ResourceKindHost,
			Environment: "production", // Changed environment
			IPAddress:   "10.0.0.2",   // Changed IP
		}

		payload2 := &model.MetricPayload{Meta: updateMeta}
		rd.ProcessMetricPayload(payload2)

		// Verify resource was updated in cache
		assert.Equal(t, resourceID, payload2.Meta.ResourceID, "Same resource ID should be used")

		updatedResource, exists := cache.GetResource(resourceID)
		require.True(t, exists)
		assert.Equal(t, "production", updatedResource.Environment, "Environment should be updated")
		assert.Equal(t, "10.0.0.2", updatedResource.IPAddress, "IP address should be updated")
		assert.True(t, updatedResource.LastSeen.After(originalLastSeen), "LastSeen should be updated")

		// Verify still no direct store calls
		assert.Empty(t, mockStore.GetCalls(), "No direct store Get calls should be made")
		assert.Empty(t, mockStore.CreateCalls(), "No direct store Create calls should be made")
		assert.Empty(t, mockStore.UpdateCalls(), "No direct store Update calls should be made")
	})

	t.Run("ParentResourceCreation uses cache", func(t *testing.T) {
		// Create container resource that should trigger parent host creation
		containerMeta := &model.Meta{
			AgentID:       "agent-456",
			HostID:        "host-parent-123",
			Hostname:      "parent-host",
			ContainerID:   "container-123",
			ContainerName: "web-app",
			Kind:          model.ResourceKindContainer,
			Environment:   "production",
		}

		payload := &model.MetricPayload{Meta: containerMeta}
		rd.ProcessMetricPayload(payload)

		// Verify container resource was created
		containerResourceID := payload.Meta.ResourceID
		containerResource, exists := cache.GetResource(containerResourceID)
		require.True(t, exists)
		assert.Equal(t, model.ResourceKindContainer, containerResource.Kind)
		assert.Equal(t, "web-app", containerResource.Name)
		assert.NotEmpty(t, containerResource.ParentID, "Container should have parent ID")

		// Verify parent host resource was also created in cache
		parentResource, exists := cache.GetResource(containerResource.ParentID)
		require.True(t, exists, "Parent host resource should be created")
		assert.Equal(t, model.ResourceKindHost, parentResource.Kind)
		assert.Equal(t, "parent-host", parentResource.Name)

		// Verify no direct store calls for parent creation
		assert.Empty(t, mockStore.GetCalls(), "No direct store Get calls should be made")
		assert.Empty(t, mockStore.CreateCalls(), "No direct store Create calls should be made")
		assert.Empty(t, mockStore.UpdateCalls(), "No direct store Update calls should be made")
	})

	t.Run("CacheAutomaticallyPersists", func(t *testing.T) {
		// Clear previous calls
		mockStore.ClearCalls()

		// Create a resource
		meta := &model.Meta{
			AgentID:     "agent-persist",
			HostID:      "host-persist",
			Hostname:    "persist-host",
			Kind:        model.ResourceKindHost,
			Environment: "test",
		}

		payload := &model.MetricPayload{Meta: meta}
		rd.ProcessMetricPayload(payload)

		// Wait for cache flush interval to trigger persistence
		time.Sleep(150 * time.Millisecond)

		// Verify that the cache automatically persisted to store
		updateBatchCalls := mockStore.UpdateBatchCalls()
		assert.NotEmpty(t, updateBatchCalls, "Cache should automatically persist resources via UpdateBatch")

		// Verify the persisted resource
		if len(updateBatchCalls) > 0 {
			persistedResources := updateBatchCalls[0].Resources
			assert.NotEmpty(t, persistedResources, "Persisted resources should not be empty")

			found := false
			for _, resource := range persistedResources {
				if resource.Name == "persist-host" {
					found = true
					assert.Equal(t, model.ResourceKindHost, resource.Kind)
					assert.Equal(t, "test", resource.Environment)
					break
				}
			}
			assert.True(t, found, "The created resource should be persisted")
		}
	})
}

func TestResourceDiscovery_ResourceIDGeneration(t *testing.T) {
	mockStore := mockstore.NewMockResourceStore()
	cache := inmemory.NewResourceCache(mockStore, time.Hour) // Long interval to avoid automatic flush
	defer cache.Stop()

	rd := NewResourceDiscovery(cache)

	t.Run("ConsistentResourceIDGeneration", func(t *testing.T) {
		meta1 := &model.Meta{
			AgentID:  "agent-123",
			HostID:   "host-123",
			Hostname: "test-host",
			Kind:     model.ResourceKindHost,
		}

		meta2 := &model.Meta{
			AgentID:  "agent-123",
			HostID:   "host-123",
			Hostname: "test-host",
			Kind:     model.ResourceKindHost,
		}

		payload1 := &model.MetricPayload{Meta: meta1}
		payload2 := &model.MetricPayload{Meta: meta2}

		rd.ProcessMetricPayload(payload1)
		rd.ProcessMetricPayload(payload2)

		// Both should generate the same resource ID
		assert.Equal(t, payload1.Meta.ResourceID, payload2.Meta.ResourceID,
			"Identical metadata should generate the same resource ID")
	})

	t.Run("DifferentKindsGenerateDifferentIDs", func(t *testing.T) {
		baseMeta := &model.Meta{
			AgentID:  "agent-456",
			HostID:   "host-456",
			Hostname: "multi-host",
		}

		hostMeta := &model.Meta{
			AgentID:  baseMeta.AgentID,
			HostID:   baseMeta.HostID,
			Hostname: baseMeta.Hostname,
			Kind:     model.ResourceKindHost,
		}

		agentMeta := &model.Meta{
			AgentID:  baseMeta.AgentID,
			HostID:   baseMeta.HostID,
			Hostname: baseMeta.Hostname,
			Kind:     model.ResourceKindAgent,
		}

		payload1 := &model.MetricPayload{Meta: hostMeta}
		payload2 := &model.MetricPayload{Meta: agentMeta}

		rd.ProcessMetricPayload(payload1)
		rd.ProcessMetricPayload(payload2)

		assert.NotEqual(t, payload1.Meta.ResourceID, payload2.Meta.ResourceID,
			"Different kinds should generate different resource IDs")
	})
}

func TestResourceDiscovery_ErrorHandling(t *testing.T) {
	mockStore := mockstore.NewMockResourceStore()
	cache := inmemory.NewResourceCache(mockStore, time.Hour)
	defer cache.Stop()

	rd := NewResourceDiscovery(cache)

	t.Run("NilPayloadHandling", func(t *testing.T) {
		// Should not panic
		assert.NotPanics(t, func() {
			rd.ProcessMetricPayload(nil)
		})
	})

	t.Run("NilMetaHandling", func(t *testing.T) {
		payload := &model.MetricPayload{Meta: nil}

		// Should not panic
		assert.NotPanics(t, func() {
			rd.ProcessMetricPayload(payload)
		})
	})

	t.Run("EmptyMetaHandling", func(t *testing.T) {
		payload := &model.MetricPayload{
			Meta: &model.Meta{}, // Empty meta
		}

		rd.ProcessMetricPayload(payload)

		// Should still work and create some kind of resource
		// The resource ID might be generated from minimal available data
		if payload.Meta.ResourceID != "" {
			resource, exists := cache.GetResource(payload.Meta.ResourceID)
			if exists {
				assert.NotEmpty(t, resource.Kind, "Resource should have some kind assigned")
			}
		}
	})
}
