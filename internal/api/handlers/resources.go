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

package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/aaronlmathis/gosight-server/internal/sys"
	"github.com/aaronlmathis/gosight-shared/model"
	"github.com/gorilla/mux"
)

// ResourcesHandler handles resource management endpoints
type ResourcesHandler struct {
	Sys *sys.SystemContext
}

// NewResourcesHandler creates a new ResourcesHandler
func NewResourcesHandler(sys *sys.SystemContext) *ResourcesHandler {
	return &ResourcesHandler{
		Sys: sys,
	}
}

// ListResources handles GET /resources
func (h *ResourcesHandler) ListResources(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit == 0 {
		limit = 100
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	filter := &model.ResourceFilter{}

	// Parse filter parameters
	if kinds := r.URL.Query().Get("kinds"); kinds != "" {
		filter.Kinds = strings.Split(kinds, ",")
	}
	if groups := r.URL.Query().Get("groups"); groups != "" {
		filter.Groups = strings.Split(groups, ",")
	}
	if status := r.URL.Query().Get("status"); status != "" {
		filter.Status = strings.Split(status, ",")
	}

	resources, err := h.Sys.Stores.Resources.List(r.Context(), filter, limit, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"resources": resources,
		"limit":     limit,
		"offset":    offset,
	})
}

// CreateResource handles POST /resources
func (h *ResourcesHandler) CreateResource(w http.ResponseWriter, r *http.Request) {
	var resource model.Resource
	if err := json.NewDecoder(r.Body).Decode(&resource); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.Sys.Stores.Resources.Create(r.Context(), &resource); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Update cache
	h.Sys.Cache.Resources.UpsertResource(&resource)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resource)
}

// GetResource handles GET /resources/{id}
func (h *ResourcesHandler) GetResource(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// Try cache first
	if resource, exists := h.Sys.Cache.Resources.GetResource(id); exists {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resource)
		return
	}

	// Fall back to store
	resource, err := h.Sys.Stores.Resources.Get(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resource)
}

// UpdateResource handles PUT /resources/{id}
func (h *ResourcesHandler) UpdateResource(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var resource model.Resource
	if err := json.NewDecoder(r.Body).Decode(&resource); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resource.ID = id
	if err := h.Sys.Stores.Resources.Update(r.Context(), &resource); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Update cache
	h.Sys.Cache.Resources.UpsertResource(&resource)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resource)
}

// DeleteResource handles DELETE /resources/{id}
func (h *ResourcesHandler) DeleteResource(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if err := h.Sys.Stores.Resources.Delete(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Remove from cache
	h.Sys.Cache.Resources.RemoveResource(id)

	w.WriteHeader(http.StatusNoContent)
}

// UpdateResourceTags handles PUT /resources/{id}/tags
func (h *ResourcesHandler) UpdateResourceTags(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var tags map[string]string
	if err := json.NewDecoder(r.Body).Decode(&tags); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.Sys.Stores.Resources.UpdateTags(r.Context(), id, tags); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Update cache
	if resource, exists := h.Sys.Cache.Resources.GetResource(id); exists {
		resource.Tags = tags
		resource.Updated = true
		h.Sys.Cache.Resources.UpsertResource(resource)
	}

	w.WriteHeader(http.StatusOK)
}

// UpdateResourceStatus handles PUT /resources/{id}/status
func (h *ResourcesHandler) UpdateResourceStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var statusUpdate struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&statusUpdate); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.Sys.Stores.Resources.UpdateStatus(r.Context(), id, statusUpdate.Status); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Update cache
	if resource, exists := h.Sys.Cache.Resources.GetResource(id); exists {
		resource.Status = statusUpdate.Status
		resource.Updated = true
		h.Sys.Cache.Resources.UpsertResource(resource)
	}

	w.WriteHeader(http.StatusOK)
}

// GetResourceSummary handles GET /resources/summary
func (h *ResourcesHandler) GetResourceSummary(w http.ResponseWriter, r *http.Request) {
	summary := h.Sys.Cache.Resources.GetSummary()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summary)
}

// GetResourceKinds handles GET /resources/kinds
func (h *ResourcesHandler) GetResourceKinds(w http.ResponseWriter, r *http.Request) {
	kinds := h.Sys.Cache.Resources.GetKinds()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(kinds)
}

// GetResourcesByKind handles GET /resources/kinds/{kind}
func (h *ResourcesHandler) GetResourcesByKind(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	kind := vars["kind"]

	resources := h.Sys.Cache.Resources.GetResourcesByKind(kind)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"kind":      kind,
		"resources": resources,
		"count":     len(resources),
	})
}

// SearchResources handles POST /resources/search
func (h *ResourcesHandler) SearchResources(w http.ResponseWriter, r *http.Request) {
	var searchQuery model.ResourceSearchQuery
	if err := json.NewDecoder(r.Body).Decode(&searchQuery); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resources, err := h.Sys.Stores.Resources.Search(r.Context(), &searchQuery)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"resources": resources,
		"query":     searchQuery,
		"count":     len(resources),
	})
}

// GetResourcesByLabels handles GET /resources/labels
func (h *ResourcesHandler) GetResourcesByLabels(w http.ResponseWriter, r *http.Request) {
	labelParams := r.URL.Query()
	labels := make(map[string]string)
	for key, values := range labelParams {
		if len(values) > 0 {
			labels[key] = values[0]
		}
	}

	resources := h.Sys.Cache.Resources.GetResourcesByLabels(labels)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"labels":    labels,
		"resources": resources,
		"count":     len(resources),
	})
}

// GetResourcesByTags handles GET /resources/tags
func (h *ResourcesHandler) GetResourcesByTags(w http.ResponseWriter, r *http.Request) {
	tagParams := r.URL.Query()
	tags := make(map[string]string)
	for key, values := range tagParams {
		if len(values) > 0 {
			tags[key] = values[0]
		}
	}

	resources := h.Sys.Cache.Resources.GetResourcesByTags(tags)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"tags":      tags,
		"resources": resources,
		"count":     len(resources),
	})
}
