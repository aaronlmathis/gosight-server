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
// This file contains permission management related routes and handlers.
package routes

import (
	"net/http"

	"github.com/aaronlmathis/gosight-server/internal/api/handlers"
	gosightauth "github.com/aaronlmathis/gosight-server/internal/auth"
	"github.com/gorilla/mux"
)

// SetupPermissionRoutes configures permission management API routes.
// It sets up endpoints for permission CRUD operations
// with appropriate middleware for authentication, authorization, and logging.
//
// Protected routes:
//   - GET /permissions - List all permissions (requires gosight:api:permissions:list permission)
//   - POST /permissions - Create a new permission (requires gosight:api:permissions:create permission)
//   - GET /permissions/{id} - Get permission by ID (requires gosight:api:permissions:view permission)
//   - PUT /permissions/{id} - Update permission (requires gosight:api:permissions:update permission)
//   - DELETE /permissions/{id} - Delete permission (requires gosight:api:permissions:delete permission)
//   - GET /permissions/{id}/roles - Get roles with permission (requires gosight:api:permissions:roles:view permission)
func SetupPermissionRoutes(router *mux.Router, permissionHandler *handlers.PermissionHandler, withAccessLog func(http.Handler) http.Handler) {
	// Configure middleware
	withAuth := gosightauth.AuthMiddleware(permissionHandler.Sys.Stores.Users)

	// Permission management routes
	permissionRouter := router.PathPrefix("/permissions").Subrouter()

	// List all permissions - GET /permissions
	permissionRouter.Handle("", withAccessLog(withAuth(
		gosightauth.RequirePermission("gosight:api:permissions:list",
			http.HandlerFunc(permissionHandler.GetPermissions),
			permissionHandler.Sys.Stores.Users,
		),
	))).Methods("GET")

	// Create new permission - POST /permissions
	permissionRouter.Handle("", withAccessLog(withAuth(
		gosightauth.RequirePermission("gosight:api:permissions:create",
			http.HandlerFunc(permissionHandler.CreatePermission),
			permissionHandler.Sys.Stores.Users,
		),
	))).Methods("POST")

	// Get specific permission - GET /permissions/{id}
	permissionRouter.Handle("/{id}", withAccessLog(withAuth(
		gosightauth.RequirePermission("gosight:api:permissions:view",
			http.HandlerFunc(permissionHandler.GetPermission),
			permissionHandler.Sys.Stores.Users,
		),
	))).Methods("GET")

	// Update permission - PUT /permissions/{id}
	permissionRouter.Handle("/{id}", withAccessLog(withAuth(
		gosightauth.RequirePermission("gosight:api:permissions:update",
			http.HandlerFunc(permissionHandler.UpdatePermission),
			permissionHandler.Sys.Stores.Users,
		),
	))).Methods("PUT")

	// Delete permission - DELETE /permissions/{id}
	permissionRouter.Handle("/{id}", withAccessLog(withAuth(
		gosightauth.RequirePermission("gosight:api:permissions:delete",
			http.HandlerFunc(permissionHandler.DeletePermission),
			permissionHandler.Sys.Stores.Users,
		),
	))).Methods("DELETE")

	// Get roles with permission - GET /permissions/{id}/roles
	permissionRouter.Handle("/{id}/roles", withAccessLog(withAuth(
		gosightauth.RequirePermission("gosight:api:permissions:roles:view",
			http.HandlerFunc(permissionHandler.GetRolesWithPermission),
			permissionHandler.Sys.Stores.Users,
		),
	))).Methods("GET")
}
