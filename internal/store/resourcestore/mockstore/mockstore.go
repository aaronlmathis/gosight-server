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

package mockstore

import (
	"context"
	"sync"
	"time"

	"github.com/aaronlmathis/gosight-shared/model"
)

// Call represents a method call with its arguments
type Call struct {
	Method    string
	Arguments []interface{}
	Timestamp time.Time
}

// CreateCall represents a Create method call
type CreateCall struct {
	Resource *model.Resource
	Error    error
}

// UpdateCall represents an Update method call
type UpdateCall struct {
	Resource *model.Resource
	Error    error
}

// UpdateBatchCall represents an UpdateBatch method call
type UpdateBatchCall struct {
	Resources []*model.Resource
	Error     error
}

// GetCall represents a Get method call
type GetCall struct {
	ID       string
	Resource *model.Resource
	Found    bool
	Error    error
}

// MockResourceStore implements resourcestore.ResourceStore for testing
type MockResourceStore struct {
	mu sync.RWMutex

	// Storage for resources
	resources map[string]*model.Resource

	// Call tracking
	calls            []Call
	createCalls      []CreateCall
	updateCalls      []UpdateCall
	updateBatchCalls []UpdateBatchCall
	getCalls         []GetCall

	// Configurable behaviors
	createError      error
	updateError      error
	updateBatchError error
	getError         error
}

// NewMockResourceStore creates a new mock resource store
func NewMockResourceStore() *MockResourceStore {
	return &MockResourceStore{
		resources:        make(map[string]*model.Resource),
		calls:            make([]Call, 0),
		createCalls:      make([]CreateCall, 0),
		updateCalls:      make([]UpdateCall, 0),
		updateBatchCalls: make([]UpdateBatchCall, 0),
		getCalls:         make([]GetCall, 0),
	}
}

// Create implements resourcestore.ResourceStore
func (m *MockResourceStore) Create(ctx context.Context, resource *model.Resource) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Track the call
	call := Call{
		Method:    "Create",
		Arguments: []interface{}{resource},
		Timestamp: time.Now(),
	}
	m.calls = append(m.calls, call)

	createCall := CreateCall{
		Resource: resource,
		Error:    m.createError,
	}
	m.createCalls = append(m.createCalls, createCall)

	if m.createError != nil {
		return m.createError
	}

	// Store the resource
	if resource != nil {
		resourceCopy := *resource
		m.resources[resource.ID] = &resourceCopy
	}

	return nil
}

// Update implements resourcestore.ResourceStore
func (m *MockResourceStore) Update(ctx context.Context, resource *model.Resource) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Track the call
	call := Call{
		Method:    "Update",
		Arguments: []interface{}{resource},
		Timestamp: time.Now(),
	}
	m.calls = append(m.calls, call)

	updateCall := UpdateCall{
		Resource: resource,
		Error:    m.updateError,
	}
	m.updateCalls = append(m.updateCalls, updateCall)

	if m.updateError != nil {
		return m.updateError
	}

	// Store the resource
	if resource != nil {
		resourceCopy := *resource
		m.resources[resource.ID] = &resourceCopy
	}

	return nil
}

// UpdateBatch implements resourcestore.ResourceStore
func (m *MockResourceStore) UpdateBatch(ctx context.Context, resources []*model.Resource) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Track the call
	call := Call{
		Method:    "UpdateBatch",
		Arguments: []interface{}{resources},
		Timestamp: time.Now(),
	}
	m.calls = append(m.calls, call)

	updateBatchCall := UpdateBatchCall{
		Resources: resources,
		Error:     m.updateBatchError,
	}
	m.updateBatchCalls = append(m.updateBatchCalls, updateBatchCall)

	if m.updateBatchError != nil {
		return m.updateBatchError
	}

	// Store all resources
	for _, resource := range resources {
		if resource != nil {
			resourceCopy := *resource
			m.resources[resource.ID] = &resourceCopy
		}
	}

	return nil
}

// Get implements resourcestore.ResourceStore
func (m *MockResourceStore) Get(ctx context.Context, id string) (*model.Resource, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Track the call
	call := Call{
		Method:    "Get",
		Arguments: []interface{}{id},
		Timestamp: time.Now(),
	}
	m.calls = append(m.calls, call)

	if m.getError != nil {
		getCall := GetCall{
			ID:    id,
			Error: m.getError,
		}
		m.getCalls = append(m.getCalls, getCall)
		return nil, m.getError
	}

	resource, found := m.resources[id]
	getCall := GetCall{
		ID:       id,
		Resource: resource,
		Found:    found,
	}
	m.getCalls = append(m.getCalls, getCall)

	if !found {
		return nil, nil
	}

	// Return a copy to prevent external modification
	resourceCopy := *resource
	return &resourceCopy, nil
}

// CreateBatch implements resourcestore.ResourceStore
func (m *MockResourceStore) CreateBatch(ctx context.Context, resources []*model.Resource) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Track the call
	call := Call{
		Method:    "CreateBatch",
		Arguments: []interface{}{resources},
		Timestamp: time.Now(),
	}
	m.calls = append(m.calls, call)

	// Store all resources
	for _, resource := range resources {
		if resource != nil {
			resourceCopy := *resource
			m.resources[resource.ID] = &resourceCopy
		}
	}

	return nil
}

// List implements resourcestore.ResourceStore
func (m *MockResourceStore) List(ctx context.Context, filter *model.ResourceFilter, limit, offset int) ([]*model.Resource, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*model.Resource
	for _, resource := range m.resources {
		matches := true

		if filter != nil {
			// Check kinds filter
			if len(filter.Kinds) > 0 {
				kindMatch := false
				for _, kind := range filter.Kinds {
					if resource.Kind == kind {
						kindMatch = true
						break
					}
				}
				if !kindMatch {
					matches = false
				}
			}

			// Check groups filter
			if matches && len(filter.Groups) > 0 {
				groupMatch := false
				for _, group := range filter.Groups {
					if resource.Group == group {
						groupMatch = true
						break
					}
				}
				if !groupMatch {
					matches = false
				}
			}

			// Check status filter
			if matches && len(filter.Status) > 0 {
				statusMatch := false
				for _, status := range filter.Status {
					if resource.Status == status {
						statusMatch = true
						break
					}
				}
				if !statusMatch {
					matches = false
				}
			}
		}

		if matches {
			resourceCopy := *resource
			result = append(result, &resourceCopy)
		}
	}

	// Apply pagination
	if offset >= len(result) {
		return []*model.Resource{}, nil
	}

	end := offset + limit
	if limit <= 0 || end > len(result) {
		end = len(result)
	}

	return result[offset:end], nil
}

// Delete implements resourcestore.ResourceStore
func (m *MockResourceStore) Delete(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.resources, id)
	return nil
}

// GetByKind implements resourcestore.ResourceStore
func (m *MockResourceStore) GetByKind(ctx context.Context, kind string) ([]*model.Resource, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*model.Resource
	for _, resource := range m.resources {
		if resource.Kind == kind {
			resourceCopy := *resource
			result = append(result, &resourceCopy)
		}
	}

	return result, nil
}

// GetByGroup implements resourcestore.ResourceStore
func (m *MockResourceStore) GetByGroup(ctx context.Context, group string) ([]*model.Resource, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*model.Resource
	for _, resource := range m.resources {
		if resource.Group == group {
			resourceCopy := *resource
			result = append(result, &resourceCopy)
		}
	}

	return result, nil
}

// GetByLabels implements resourcestore.ResourceStore
func (m *MockResourceStore) GetByLabels(ctx context.Context, labels map[string]string) ([]*model.Resource, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*model.Resource
	for _, resource := range m.resources {
		matches := true
		for key, value := range labels {
			if resource.Labels[key] != value {
				matches = false
				break
			}
		}
		if matches {
			resourceCopy := *resource
			result = append(result, &resourceCopy)
		}
	}

	return result, nil
}

// GetByTags implements resourcestore.ResourceStore
func (m *MockResourceStore) GetByTags(ctx context.Context, tags map[string]string) ([]*model.Resource, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*model.Resource
	for _, resource := range m.resources {
		matches := true
		for key, value := range tags {
			if resource.Tags[key] != value {
				matches = false
				break
			}
		}
		if matches {
			resourceCopy := *resource
			result = append(result, &resourceCopy)
		}
	}

	return result, nil
}

// GetByParent implements resourcestore.ResourceStore
func (m *MockResourceStore) GetByParent(ctx context.Context, parentID string) ([]*model.Resource, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*model.Resource
	for _, resource := range m.resources {
		if resource.ParentID == parentID {
			resourceCopy := *resource
			result = append(result, &resourceCopy)
		}
	}

	return result, nil
}

// GetStaleResources implements resourcestore.ResourceStore
func (m *MockResourceStore) GetStaleResources(ctx context.Context, threshold time.Duration) ([]*model.Resource, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	cutoff := time.Now().Add(-threshold)
	var result []*model.Resource

	for _, resource := range m.resources {
		if resource.LastSeen.Before(cutoff) {
			resourceCopy := *resource
			result = append(result, &resourceCopy)
		}
	}

	return result, nil
}

// Helper methods for test assertions

// GetCalls returns all Get method calls
func (m *MockResourceStore) GetCalls() []GetCall {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]GetCall, len(m.getCalls))
	copy(result, m.getCalls)
	return result
}

// CreateCalls returns all Create method calls
func (m *MockResourceStore) CreateCalls() []CreateCall {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]CreateCall, len(m.createCalls))
	copy(result, m.createCalls)
	return result
}

// UpdateCalls returns all Update method calls
func (m *MockResourceStore) UpdateCalls() []UpdateCall {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]UpdateCall, len(m.updateCalls))
	copy(result, m.updateCalls)
	return result
}

// UpdateBatchCalls returns all UpdateBatch method calls
func (m *MockResourceStore) UpdateBatchCalls() []UpdateBatchCall {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]UpdateBatchCall, len(m.updateBatchCalls))
	copy(result, m.updateBatchCalls)
	return result
}

// AllCalls returns all method calls
func (m *MockResourceStore) AllCalls() []Call {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]Call, len(m.calls))
	copy(result, m.calls)
	return result
}

// ClearCalls clears all tracked calls
func (m *MockResourceStore) ClearCalls() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.calls = make([]Call, 0)
	m.createCalls = make([]CreateCall, 0)
	m.updateCalls = make([]UpdateCall, 0)
	m.updateBatchCalls = make([]UpdateBatchCall, 0)
	m.getCalls = make([]GetCall, 0)
}

// SetCreateError configures Create to return an error
func (m *MockResourceStore) SetCreateError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.createError = err
}

// SetUpdateError configures Update to return an error
func (m *MockResourceStore) SetUpdateError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.updateError = err
}

// SetUpdateBatchError configures UpdateBatch to return an error
func (m *MockResourceStore) SetUpdateBatchError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.updateBatchError = err
}

// SetGetError configures Get to return an error
func (m *MockResourceStore) SetGetError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.getError = err
}

// Count implements resourcestore.ResourceStore
func (m *MockResourceStore) Count(ctx context.Context, filter *model.ResourceFilter) (int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// For simplicity, just return total count if no filter, otherwise apply basic filtering
	if filter == nil {
		return len(m.resources), nil
	}

	count := 0
	for _, resource := range m.resources {
		matches := true

		// Check kinds filter
		if len(filter.Kinds) > 0 {
			kindMatch := false
			for _, kind := range filter.Kinds {
				if resource.Kind == kind {
					kindMatch = true
					break
				}
			}
			if !kindMatch {
				matches = false
			}
		}

		// Check groups filter
		if matches && len(filter.Groups) > 0 {
			groupMatch := false
			for _, group := range filter.Groups {
				if resource.Group == group {
					groupMatch = true
					break
				}
			}
			if !groupMatch {
				matches = false
			}
		}

		// Check status filter
		if matches && len(filter.Status) > 0 {
			statusMatch := false
			for _, status := range filter.Status {
				if resource.Status == status {
					statusMatch = true
					break
				}
			}
			if !statusMatch {
				matches = false
			}
		}

		if matches {
			count++
		}
	}

	return count, nil
}

// Search implements resourcestore.ResourceStore
func (m *MockResourceStore) Search(ctx context.Context, query *model.ResourceSearchQuery) ([]*model.Resource, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// For simplicity, just return all resources (a real implementation would do proper searching)
	result := make([]*model.Resource, 0, len(m.resources))
	for _, resource := range m.resources {
		resourceCopy := *resource
		result = append(result, &resourceCopy)
	}

	return result, nil
}

// GetChildren implements resourcestore.ResourceStore
func (m *MockResourceStore) GetChildren(ctx context.Context, parentID string) ([]*model.Resource, error) {
	return m.GetByParent(ctx, parentID)
}

// GetParent implements resourcestore.ResourceStore
func (m *MockResourceStore) GetParent(ctx context.Context, childID string) (*model.Resource, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	child, found := m.resources[childID]
	if !found || child.ParentID == "" {
		return nil, nil
	}

	parent, found := m.resources[child.ParentID]
	if !found {
		return nil, nil
	}

	parentCopy := *parent
	return &parentCopy, nil
}

// UpdateLabels implements resourcestore.ResourceStore
func (m *MockResourceStore) UpdateLabels(ctx context.Context, resourceID string, labels map[string]string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	resource, found := m.resources[resourceID]
	if !found {
		return nil // Resource not found
	}

	resourceCopy := *resource
	resourceCopy.Labels = make(map[string]string)
	for k, v := range labels {
		resourceCopy.Labels[k] = v
	}
	m.resources[resourceID] = &resourceCopy

	return nil
}

// UpdateTags implements resourcestore.ResourceStore
func (m *MockResourceStore) UpdateTags(ctx context.Context, resourceID string, tags map[string]string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	resource, found := m.resources[resourceID]
	if !found {
		return nil // Resource not found
	}

	resourceCopy := *resource
	resourceCopy.Tags = make(map[string]string)
	for k, v := range tags {
		resourceCopy.Tags[k] = v
	}
	m.resources[resourceID] = &resourceCopy

	return nil
}

// UpdateStatus implements resourcestore.ResourceStore
func (m *MockResourceStore) UpdateStatus(ctx context.Context, resourceID string, status string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	resource, found := m.resources[resourceID]
	if !found {
		return nil // Resource not found
	}

	resourceCopy := *resource
	resourceCopy.Status = status
	m.resources[resourceID] = &resourceCopy

	return nil
}

// UpdateLastSeen implements resourcestore.ResourceStore
func (m *MockResourceStore) UpdateLastSeen(ctx context.Context, resourceID string, lastSeen time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	resource, found := m.resources[resourceID]
	if !found {
		return nil // Resource not found
	}

	resourceCopy := *resource
	resourceCopy.LastSeen = lastSeen
	m.resources[resourceID] = &resourceCopy

	return nil
}

// GetResourceSummary implements resourcestore.ResourceStore
func (m *MockResourceStore) GetResourceSummary(ctx context.Context) (map[string]int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	summary := make(map[string]int)
	for _, resource := range m.resources {
		summary[resource.Kind]++
	}

	return summary, nil
}

// GetResourcesByKind implements resourcestore.ResourceStore
func (m *MockResourceStore) GetResourcesByKind(ctx context.Context, kind string) ([]*model.Resource, error) {
	return m.GetByKind(ctx, kind)
}

// GetStoredResource returns a resource from internal storage (for testing)
func (m *MockResourceStore) GetStoredResource(id string) (*model.Resource, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	resource, found := m.resources[id]
	if !found {
		return nil, false
	}

	resourceCopy := *resource
	return &resourceCopy, true
}
