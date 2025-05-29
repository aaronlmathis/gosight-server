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

// Package routes provides HTTP route configuration for the GoSight API server.
// This file contains role management related routes and handlers.
package routes

import (
	"net/http"

	"github.com/aaronlmathis/gosight-server/internal/api/handlers"
	gosightauth "github.com/aaronlmathis/gosight-server/internal/auth"
	"github.com/gorilla/mux"
)

// SetupRoleRoutes configures role management API routes.
// It sets up endpoints for role CRUD operations and role-permission assignments
// with appropriate middleware for authentication, authorization, and logging.
//
// Protected routes:
//   - GET /roles - List all roles (requires gosight:api:roles:list permission)
//   - POST /roles - Create a new role (requires gosight:api:roles:create permission)
//   - GET /roles/{id} - Get role by ID (requires gosight:api:roles:view permission)
//   - PUT /roles/{id} - Update role (requires gosight:api:roles:update permission)
//   - DELETE /roles/{id} - Delete role (requires gosight:api:roles:delete permission)
//   - GET /roles/{id}/permissions - Get role permissions (requires gosight:api:roles:permissions:view permission)
//   - POST /roles/{id}/permissions - Assign permissions to role (requires gosight:api:roles:permissions:assign permission)
//   - GET /roles/{id}/users - Get users with role (requires gosight:api:roles:users:view permission)
func SetupRoleRoutes(router *mux.Router, roleHandler *handlers.RoleHandler, withAccessLog func(http.Handler) http.Handler) {
	// Configure middleware
	withAuth := gosightauth.AuthMiddleware(roleHandler.Sys.Stores.Users)

	// Role management routes
	roleRouter := router.PathPrefix("/roles").Subrouter()

	// List all roles - GET /roles
	roleRouter.Handle("", withAccessLog(withAuth(
		gosightauth.RequirePermission("gosight:api:roles:list",
			http.HandlerFunc(roleHandler.GetRoles),
			roleHandler.Sys.Stores.Users,
		),
	))).Methods("GET")

	// Create new role - POST /roles
	roleRouter.Handle("", withAccessLog(withAuth(
		gosightauth.RequirePermission("gosight:api:roles:create",
			http.HandlerFunc(roleHandler.CreateRole),
			roleHandler.Sys.Stores.Users,
		),
	))).Methods("POST")

	// Get specific role - GET /roles/{id}
	roleRouter.Handle("/{id}", withAccessLog(withAuth(
		gosightauth.RequirePermission("gosight:api:roles:view",
			http.HandlerFunc(roleHandler.GetRole),
			roleHandler.Sys.Stores.Users,
		),
	))).Methods("GET")

	// Update role - PUT /roles/{id}
	roleRouter.Handle("/{id}", withAccessLog(withAuth(
		gosightauth.RequirePermission("gosight:api:roles:update",
			http.HandlerFunc(roleHandler.UpdateRole),
			roleHandler.Sys.Stores.Users,
		),
	))).Methods("PUT")

	// Delete role - DELETE /roles/{id}
	roleRouter.Handle("/{id}", withAccessLog(withAuth(
		gosightauth.RequirePermission("gosight:api:roles:delete",
			http.HandlerFunc(roleHandler.DeleteRole),
			roleHandler.Sys.Stores.Users,
		),
	))).Methods("DELETE")

	// Get role permissions - GET /roles/{id}/permissions
	roleRouter.Handle("/{id}/permissions", withAccessLog(withAuth(
		gosightauth.RequirePermission("gosight:api:roles:permissions:view",
			http.HandlerFunc(roleHandler.GetRolePermissions),
			roleHandler.Sys.Stores.Users,
		),
	))).Methods("GET")

	// Assign permissions to role - POST /roles/{id}/permissions
	roleRouter.Handle("/{id}/permissions", withAccessLog(withAuth(
		gosightauth.RequirePermission("gosight:api:roles:permissions:assign",
			http.HandlerFunc(roleHandler.AssignPermissionsToRole),
			roleHandler.Sys.Stores.Users,
		),
	))).Methods("POST")

	// Get users with role - GET /roles/{id}/users
	roleRouter.Handle("/{id}/users", withAccessLog(withAuth(
		gosightauth.RequirePermission("gosight:api:roles:users:view",
			http.HandlerFunc(roleHandler.GetUsersWithRole),
			roleHandler.Sys.Stores.Users,
		),
	))).Methods("GET")
}
