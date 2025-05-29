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
	"strings"

	"github.com/aaronlmathis/gosight-server/internal/sys"
	"github.com/aaronlmathis/gosight-shared/utils"
	"github.com/gorilla/mux"
)

// TagsHandler handles tag-related API endpoints
type TagsHandler struct {
	Sys *sys.SystemContext
}

// NewTagsHandler creates a new TagsHandler
func NewTagsHandler(sys *sys.SystemContext) *TagsHandler {
	return &TagsHandler{
		Sys: sys,
	}
}

// HandleGetTags returns all tags for an endpoint
func (h *TagsHandler) HandleGetTags(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	endpointID := vars["endpointID"]
	utils.Debug("HandleGetTags() for endpoint ID: %s", endpointID)

	tags := h.Sys.Cache.Tags.GetFlattenedTagsForEndpoint(endpointID) // returns map[string]string

	if tags == nil {
		tags = make(map[string]string) // Return an empty map if no tags exist
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(tags)
}

// HandleSetTags replaces all tags for an endpoint
func (h *TagsHandler) HandleSetTags(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	endpointID := vars["endpointID"]

	var tags map[string]string
	if err := json.NewDecoder(r.Body).Decode(&tags); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	err := h.Sys.Stores.Data.SetTags(r.Context(), endpointID, tags)
	if err != nil {
		http.Error(w, "Failed to set tags", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// HandlePatchTags updates or adds individual tags
func (h *TagsHandler) HandlePatchTags(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	endpointID := vars["endpointID"]

	var updates map[string]string
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	existing, err := h.Sys.Stores.Data.GetTags(r.Context(), endpointID)
	if err != nil {
		http.Error(w, "Failed to fetch existing tags", http.StatusInternalServerError)
		return
	}

	for k, v := range updates {
		existing[k] = v
	}

	err = h.Sys.Stores.Data.SetTags(r.Context(), endpointID, existing)
	if err != nil {
		http.Error(w, "Failed to update tags", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// HandleDeleteTag deletes a specific tag key
func (h *TagsHandler) HandleDeleteTag(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	endpointID := vars["endpointID"]
	tagKey := vars["key"]

	tagKey = strings.TrimSpace(tagKey)
	if tagKey == "" {
		http.Error(w, "Tag key required", http.StatusBadRequest)
		return
	}

	err := h.Sys.Stores.Data.DeleteTag(r.Context(), endpointID, tagKey)
	if err != nil {
		http.Error(w, "Failed to delete tag", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// HandleTagKeys returns all known tag keys
func (h *TagsHandler) HandleTagKeys(w http.ResponseWriter, r *http.Request) {
	keys := h.Sys.Cache.Tags.GetTagKeys()
	utils.JSON(w, http.StatusOK, keys)
}

// HandleTagValues returns all values for a given key
func (h *TagsHandler) HandleTagValues(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	if key == "" {
		http.Error(w, "Missing key parameter", http.StatusBadRequest)
		return
	}

	valueSet := h.Sys.Cache.Tags.GetTagValues(key)
	values := make([]string, 0, len(valueSet))
	for v := range valueSet {
		values = append(values, v)
	}

	// Optional fuzzy filter
	query := strings.ToLower(r.URL.Query().Get("contains"))
	if query != "" {
		filtered := make([]string, 0, len(values))
		for _, val := range values {
			if strings.Contains(strings.ToLower(val), query) {
				filtered = append(filtered, val)
			}
		}
		values = filtered
	}

	utils.JSON(w, http.StatusOK, values)
}

// HandleAPITags handles GET /tags - List all tags
func (h *TagsHandler) HandleAPITags(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement general tag listing functionality
	// This would require a data structure for managing general tags, not just endpoint tags
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

// HandleAPITagCreate handles POST /tags - Create new tag
func (h *TagsHandler) HandleAPITagCreate(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement general tag creation functionality
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

// HandleAPITag handles GET /tags/{id} - Get tag by ID
func (h *TagsHandler) HandleAPITag(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement get tag by ID functionality
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

// HandleAPITagUpdate handles PUT /tags/{id} - Update tag
func (h *TagsHandler) HandleAPITagUpdate(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement tag update functionality
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

// HandleAPITagDelete handles DELETE /tags/{id} - Delete tag
func (h *TagsHandler) HandleAPITagDelete(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement tag deletion functionality
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

// HandleAPITagSearch handles GET /tags/search - Search tags
func (h *TagsHandler) HandleAPITagSearch(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement tag search functionality
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

// HandleAPITagAssign handles POST /tags/{id}/assign - Assign tag to resource
func (h *TagsHandler) HandleAPITagAssign(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement tag assignment functionality
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

// HandleAPITagUnassign handles DELETE /tags/{id}/assign - Remove tag from resource
func (h *TagsHandler) HandleAPITagUnassign(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement tag unassignment functionality
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}
