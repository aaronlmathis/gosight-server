// Copyright (C) 2025 Aaron Mathis
// This file is part of GoSight Server.
//
// GoSight Server is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// GoSight Server is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with GoSight Server.  If not, see <https://www.gnu.org/licenses/>.

// Package handlers provides HTTP handlers for the GoSight Server REST API.
// This file contains handlers for managing permissions and their associations
// with roles in the system. It supports CRUD operations for permissions and
// querying roles by permission.
package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/aaronlmathis/gosight-server/internal/sys"
	"github.com/aaronlmathis/gosight-server/internal/usermodel"
	"github.com/gorilla/mux"
)

// PermissionHandler handles HTTP requests related to permissions.
type PermissionHandler struct {
	Sys *sys.SystemContext
}

// NewPermissionHandler creates a new PermissionHandler with the given system context.
func NewPermissionHandler(sys *sys.SystemContext) *PermissionHandler {
	return &PermissionHandler{Sys: sys}
}

// CreatePermissionRequest represents the payload for creating a new permission.
type CreatePermissionRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// UpdatePermissionRequest represents the payload for updating an existing permission.
type UpdatePermissionRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// GetPermissions handles GET /permissions and returns all permissions.
// Responds with a JSON array of Permission objects.
func (h *PermissionHandler) GetPermissions(w http.ResponseWriter, r *http.Request) {
	permissions, err := h.Sys.Stores.Users.GetPermissions(r.Context())
	if err != nil {
		http.Error(w, "Failed to get permissions", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(permissions)
}

// GetPermission handles GET /permissions/{id} and returns a specific permission by ID.
// Responds with a JSON Permission object if found.
func (h *PermissionHandler) GetPermission(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	permissionID := vars["id"]

	permission, err := h.Sys.Stores.Users.GetPermission(r.Context(), permissionID)
	if err != nil {
		http.Error(w, "Failed to get permission", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(permission)
}

// CreatePermission handles POST /permissions and creates a new permission.
// Expects a JSON body with name and description fields.
// Responds with the created Permission object.
func (h *PermissionHandler) CreatePermission(w http.ResponseWriter, r *http.Request) {
	var req CreatePermissionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	permission := &usermodel.Permission{
		Name:        req.Name,
		Description: req.Description,
	}

	if err := h.Sys.Stores.Users.CreatePermission(r.Context(), permission); err != nil {
		http.Error(w, "Failed to create permission", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(permission)
}

// UpdatePermission handles PUT /permissions/{id} and updates an existing permission.
// Expects a JSON body with name and description fields.
// Responds with the updated Permission object.
func (h *PermissionHandler) UpdatePermission(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	permissionID := vars["id"]

	var req UpdatePermissionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	permission := &usermodel.Permission{
		ID:          permissionID,
		Name:        req.Name,
		Description: req.Description,
	}

	if err := h.Sys.Stores.Users.UpdatePermission(r.Context(), permission); err != nil {
		http.Error(w, "Failed to update permission", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(permission)
}

// DeletePermission handles DELETE /permissions/{id} and deletes a permission.
// Responds with HTTP 204 No Content on success.
func (h *PermissionHandler) DeletePermission(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	permissionID := vars["id"]

	if err := h.Sys.Stores.Users.DeletePermission(r.Context(), permissionID); err != nil {
		http.Error(w, "Failed to delete permission", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetRolesWithPermission handles GET /permissions/{id}/roles and returns roles with a specific permission.
// Responds with a JSON array of Role objects.
func (h *PermissionHandler) GetRolesWithPermission(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	permissionID := vars["id"]

	roles, err := h.Sys.Stores.Users.GetRolesWithPermission(r.Context(), permissionID)
	if err != nil {
		http.Error(w, "Failed to get roles with permission", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(roles)
}
