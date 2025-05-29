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
// This file contains resource management routes and handlers.
package routes

import (
	"net/http"

	"github.com/aaronlmathis/gosight-server/internal/api/handlers"
	gosightauth "github.com/aaronlmathis/gosight-server/internal/auth"
	"github.com/aaronlmathis/gosight-server/internal/sys"
	"github.com/gorilla/mux"
)

// SetupResourceRoutes configures resource management API routes.
// It sets up endpoints for resource CRUD operations, resource discovery,
// and resource querying with appropriate middleware for authentication, authorization, and logging.
//
// Protected routes:
//   - GET /resources - List resources (requires gosight:api:resources:view permission)
//   - POST /resources - Create resource (requires gosight:api:resources:create permission)
//   - GET /resources/{id} - Get resource by ID (requires gosight:api:resources:view permission)
//   - PUT /resources/{id} - Update resource (requires gosight:api:resources:update permission)
//   - DELETE /resources/{id} - Delete resource (requires gosight:api:resources:delete permission)
//   - PUT /resources/{id}/tags - Update resource tags (requires gosight:api:resources:update permission)
//   - PUT /resources/{id}/status - Update resource status (requires gosight:api:resources:update permission)
//   - GET /resources/summary - Get resource summary (requires gosight:api:resources:view permission)
//   - GET /resources/kinds - Get resource kinds (requires gosight:api:resources:view permission)
//   - GET /resources/kinds/{kind} - Get resources by kind (requires gosight:api:resources:view permission)
//   - POST /resources/search - Search resources (requires gosight:api:resources:search permission)
//   - GET /resources/labels - Get resources by labels (requires gosight:api:resources:view permission)
//   - GET /resources/tags - Get resources by tags (requires gosight:api:resources:view permission)
func SetupResourceRoutes(router *mux.Router, sys *sys.SystemContext, withAccessLog func(http.Handler) http.Handler) {
	// Configure middleware
	withAuth := gosightauth.AuthMiddleware(sys.Stores.Users)


	// Helper function to create secure handler with permission check
	secure := func(permission string, handler http.Handler) http.Handler {
		return withAccessLog(withAuth(gosightauth.RequirePermission(permission, handler, sys.Stores.Users)))
	}

	// Create resource handler
	resourceHandler := handlers.NewResourcesHandler(sys)

	// Resource CRUD operations
	router.Handle("/resources",
		secure("gosight:api:resources:view", http.HandlerFunc(resourceHandler.ListResources))).
		Methods("GET")

	router.Handle("/resources",
		secure("gosight:api:resources:create", http.HandlerFunc(resourceHandler.CreateResource))).
		Methods("POST")

	router.Handle("/resources/{id}",
		secure("gosight:api:resources:view", http.HandlerFunc(resourceHandler.GetResource))).
		Methods("GET")

	router.Handle("/resources/{id}",
		secure("gosight:api:resources:update", http.HandlerFunc(resourceHandler.UpdateResource))).
		Methods("PUT")

	router.Handle("/resources/{id}",
		secure("gosight:api:resources:delete", http.HandlerFunc(resourceHandler.DeleteResource))).
		Methods("DELETE")

	router.Handle("/resources/{id}/tags",
		secure("gosight:api:resources:update", http.HandlerFunc(resourceHandler.UpdateResourceTags))).
		Methods("PUT")

	router.Handle("/resources/{id}/status",
		secure("gosight:api:resources:update", http.HandlerFunc(resourceHandler.UpdateResourceStatus))).
		Methods("PUT")

	// Aggregation endpoints
	router.Handle("/resources/summary",
		secure("gosight:api:resources:view", http.HandlerFunc(resourceHandler.GetResourceSummary))).
		Methods("GET")

	router.Handle("/resources/kinds",
		secure("gosight:api:resources:view", http.HandlerFunc(resourceHandler.GetResourceKinds))).
		Methods("GET")

	router.Handle("/resources/kinds/{kind}",
		secure("gosight:api:resources:view", http.HandlerFunc(resourceHandler.GetResourcesByKind))).
		Methods("GET")

	// Search endpoints
	router.Handle("/resources/search",
		secure("gosight:api:resources:search", http.HandlerFunc(resourceHandler.SearchResources))).
		Methods("POST")

	router.Handle("/resources/labels",
		secure("gosight:api:resources:view", http.HandlerFunc(resourceHandler.GetResourcesByLabels))).
		Methods("GET")

	router.Handle("/resources/tags",
		secure("gosight:api:resources:view", http.HandlerFunc(resourceHandler.GetResourcesByTags))).
		Methods("GET")
}
