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
// This file contains endpoint management related routes and handlers.
package routes

import (
	"net/http"

	"github.com/aaronlmathis/gosight-server/internal/api/handlers"
	gosightauth "github.com/aaronlmathis/gosight-server/internal/auth"
	"github.com/gorilla/mux"
)

// SetupEndpointsRoutes configures endpoint management API routes.
// It sets up endpoints for managing monitoring endpoints, endpoint configuration,
// and endpoint status with appropriate middleware for authentication, authorization, and logging.
//
// Protected routes:
//   - GET /endpoints - List all endpoints (requires gosight:api:endpoints:view permission)
//   - POST /endpoints - Create new endpoint (requires gosight:api:endpoints:create permission)
//   - GET /endpoints/{id} - Get endpoint by ID (requires gosight:api:endpoints:view permission)
//   - PUT /endpoints/{id} - Update endpoint (requires gosight:api:endpoints:update permission)
//   - DELETE /endpoints/{id} - Delete endpoint (requires gosight:api:endpoints:delete permission)
//   - POST /endpoints/{id}/test - Test endpoint connection (requires gosight:api:endpoints:test permission)
//   - GET /endpoints/{id}/status - Get endpoint status (requires gosight:api:endpoints:view permission)
//   - POST /endpoints/{id}/enable - Enable endpoint (requires gosight:api:endpoints:manage permission)
//   - POST /endpoints/{id}/disable - Disable endpoint (requires gosight:api:endpoints:manage permission)
func SetupEndpointsRoutes(router *mux.Router, endpointsHandler *handlers.EndpointsHandler, withAccessLog func(http.Handler) http.Handler) {
	// Configure middleware
	withAuth := gosightauth.AuthMiddleware(endpointsHandler.Sys.Stores.Users)
	withLog := withAccessLog

	// Helper function to create secure handler with permission check
	secure := func(permission string, handler http.Handler) http.Handler {
		return withLog(withAuth(gosightauth.RequirePermission(permission, handler, endpointsHandler.Sys.Stores.Users)))
	}

	// Endpoint CRUD operations
	router.Handle("/endpoints",
		secure("gosight:api:endpoints:view", http.HandlerFunc(endpointsHandler.HandleAPIEndpoints))).
		Methods("GET")

	router.Handle("/endpoints/{endpointType}",
		secure("gosight:api:endpoints:view", http.HandlerFunc(endpointsHandler.HandleAPIEndpointsByType))).
		Methods("GET")

	router.Handle("/endpoints/{id}",
		secure("gosight:api:endpoints:view", http.HandlerFunc(endpointsHandler.HandleAPIEndpoint))).
		Methods("GET")

	router.Handle("/endpoints/{id}",
		secure("gosight:api:endpoints:delete", http.HandlerFunc(endpointsHandler.HandleAPIEndpointDelete))).
		Methods("DELETE")

	// Endpoint management operations
	router.Handle("/endpoints/{id}/test",
		secure("gosight:api:endpoints:test", http.HandlerFunc(endpointsHandler.HandleAPIEndpointTest))).
		Methods("POST")

	router.Handle("/endpoints/{id}/status",
		secure("gosight:api:endpoints:view", http.HandlerFunc(endpointsHandler.HandleAPIEndpointStatus))).
		Methods("GET")

	router.Handle("/endpoints/{id}/enable",
		secure("gosight:api:endpoints:manage", http.HandlerFunc(endpointsHandler.HandleAPIEndpointEnable))).
		Methods("POST")

	router.Handle("/endpoints/{id}/disable",
		secure("gosight:api:endpoints:manage", http.HandlerFunc(endpointsHandler.HandleAPIEndpointDisable))).
		Methods("POST")
}
