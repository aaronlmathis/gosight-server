package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/aaronlmathis/gosight-server/internal/sys"
	"github.com/aaronlmathis/gosight-server/internal/usermodel"
	"github.com/gorilla/mux"
)

type RoleHandler struct {
	Sys *sys.SystemContext
}

func NewRoleHandler(sys *sys.SystemContext) *RoleHandler {
	return &RoleHandler{Sys: sys}
}

type CreateRoleRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type UpdateRoleRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type AssignPermissionRequest struct {
	PermissionIDs []string `json:"permission_ids"`
}

// GetRoles returns all roles
func (h *RoleHandler) GetRoles(w http.ResponseWriter, r *http.Request) {
	roles, err := h.Sys.Stores.Users.GetRoles(r.Context())
	if err != nil {
		http.Error(w, "Failed to get roles", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(roles)
}

// GetRole returns a specific role by ID
func (h *RoleHandler) GetRole(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roleID := vars["id"]

	role, err := h.Sys.Stores.Users.GetRole(r.Context(), roleID)
	if err != nil {
		http.Error(w, "Failed to get role", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(role)
}

// CreateRole creates a new role
func (h *RoleHandler) CreateRole(w http.ResponseWriter, r *http.Request) {
	var req CreateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	role := &usermodel.Role{
		Name:        req.Name,
		Description: req.Description,
	}

	if err := h.Sys.Stores.Users.CreateRole(r.Context(), role); err != nil {
		http.Error(w, "Failed to create role", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(role)
}

// UpdateRole updates an existing role
func (h *RoleHandler) UpdateRole(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roleID := vars["id"]

	var req UpdateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	role := &usermodel.Role{
		ID:          roleID,
		Name:        req.Name,
		Description: req.Description,
	}

	if err := h.Sys.Stores.Users.UpdateRole(r.Context(), role); err != nil {
		http.Error(w, "Failed to update role", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(role)
}

// DeleteRole deletes a role
func (h *RoleHandler) DeleteRole(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roleID := vars["id"]

	if err := h.Sys.Stores.Users.DeleteRole(r.Context(), roleID); err != nil {
		http.Error(w, "Failed to delete role", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetRolePermissions returns permissions for a specific role
func (h *RoleHandler) GetRolePermissions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roleID := vars["id"]

	permissions, err := h.Sys.Stores.Users.GetRolePermissions(r.Context(), roleID)
	if err != nil {
		http.Error(w, "Failed to get role permissions", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(permissions)
}

// AssignPermissionsToRole assigns permissions to a role
func (h *RoleHandler) AssignPermissionsToRole(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roleID := vars["id"]

	var req AssignPermissionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.Sys.Stores.Users.AssignPermissionsToRole(r.Context(), roleID, req.PermissionIDs); err != nil {
		http.Error(w, "Failed to assign permissions to role", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// GetUsersWithRole returns users that have a specific role
func (h *RoleHandler) GetUsersWithRole(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roleID := vars["id"]

	users, err := h.Sys.Stores.Users.GetUsersWithRole(r.Context(), roleID)
	if err != nil {
		http.Error(w, "Failed to get users with role", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}
