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
// This file contains debug and development related routes and handlers.
package routes

import (
	"net/http"

	"github.com/aaronlmathis/gosight-server/internal/api/handlers"
	gosightauth "github.com/aaronlmathis/gosight-server/internal/auth"
	"github.com/gorilla/mux"
)

// SetupDebugRoutes configures debug and development API routes.
// These routes should only be enabled in development or testing environments.
// They provide access to internal server state, debugging information, and test utilities
// with appropriate middleware for authentication, authorization, and logging.
//
// Protected routes:
//   - GET /debug/health - Server health check (requires gosight:api:debug:health permission)
//   - GET /debug/metrics - Internal metrics (requires gosight:api:debug:metrics permission)
//   - GET /debug/pprof - Performance profiling (requires gosight:api:debug:pprof permission)
//   - GET /debug/config - Server configuration (requires gosight:api:debug:config permission)
//   - POST /debug/test - Test endpoint (requires gosight:api:debug:test permission)
//   - GET /debug/version - Version information (requires gosight:api:debug:info permission)
//   - GET /debug/status - System status (requires gosight:api:debug:status permission)
func SetupDebugRoutes(router *mux.Router, debugHandler *handlers.DebugHandler, withAccessLog func(http.Handler) http.Handler) {
	// Configure middleware
	withAuth := gosightauth.AuthMiddleware(debugHandler.Sys.Stores.Users)

	// Helper function to create secure handler with permission check
	secure := func(permission string, handler http.Handler) http.Handler {
		return withAccessLog(withAuth(gosightauth.RequirePermission(permission, handler, debugHandler.Sys.Stores.Users)))
	}

	// Debug information endpoints
	router.Handle("/debug/health",
		secure("gosight:api:debug:health", http.HandlerFunc(debugHandler.HandleAPIHealthCheck))).
		Methods("GET")

	router.Handle("/debug/metrics",
		secure("gosight:api:debug:metrics", http.HandlerFunc(debugHandler.HandleAPIDebugMetrics))).
		Methods("GET")

	router.Handle("/debug/pprof",
		secure("gosight:api:debug:pprof", http.HandlerFunc(debugHandler.HandleAPIDebugPprof))).
		Methods("GET")

	router.Handle("/debug/config",
		secure("gosight:api:debug:config", http.HandlerFunc(debugHandler.HandleAPIDebugConfig))).
		Methods("GET")

	router.Handle("/debug/test",
		secure("gosight:api:debug:test", http.HandlerFunc(debugHandler.HandleAPIDebugTest))).
		Methods("POST")

	router.Handle("/debug/version",
		secure("gosight:api:debug:info", http.HandlerFunc(debugHandler.HandleAPIVersion))).
		Methods("GET")

	router.Handle("/debug/status",
		secure("gosight:api:debug:status", http.HandlerFunc(debugHandler.HandleAPIStatus))).
		Methods("GET")
}
