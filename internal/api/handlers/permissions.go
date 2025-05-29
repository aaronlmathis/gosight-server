package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/aaronlmathis/gosight-server/internal/sys"
	"github.com/aaronlmathis/gosight-server/internal/usermodel"
	"github.com/gorilla/mux"
)

type PermissionHandler struct {
	Sys *sys.SystemContext
}

func NewPermissionHandler(sys *sys.SystemContext) *PermissionHandler {
	return &PermissionHandler{Sys: sys}
}

type CreatePermissionRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type UpdatePermissionRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// GetPermissions returns all permissions
func (h *PermissionHandler) GetPermissions(w http.ResponseWriter, r *http.Request) {
	permissions, err := h.Sys.Stores.Users.GetPermissions(r.Context())
	if err != nil {
		http.Error(w, "Failed to get permissions", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(permissions)
}

// GetPermission returns a specific permission by ID
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

// CreatePermission creates a new permission
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

// UpdatePermission updates an existing permission
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

// DeletePermission deletes a permission
func (h *PermissionHandler) DeletePermission(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	permissionID := vars["id"]

	if err := h.Sys.Stores.Users.DeletePermission(r.Context(), permissionID); err != nil {
		http.Error(w, "Failed to delete permission", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetRolesWithPermission returns roles that have a specific permission
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
