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
// This file contains log management related routes and handlers.
package routes

import (
	"net/http"

	"github.com/aaronlmathis/gosight-server/internal/api/handlers"
	gosightauth "github.com/aaronlmathis/gosight-server/internal/auth"
	"github.com/gorilla/mux"
)

// SetupLogsRoutes configures log management API routes.
// It sets up endpoints for log querying, log streaming, and log configuration
// with appropriate middleware for authentication, authorization, and logging.
//
// Protected routes:
//   - GET /logs - Query logs (requires gosight:api:logs:view permission)
//   - GET /logs/stream - Stream logs (requires gosight:api:logs:stream permission)
//   - GET /logs/search - Search logs (requires gosight:api:logs:search permission)
//   - GET /logs/sources - Get log sources (requires gosight:api:logs:view permission)
//   - POST /logs/export - Export logs (requires gosight:api:logs:export permission)
//   - GET /logs/stats - Get log statistics (requires gosight:api:logs:view permission)
func SetupLogsRoutes(router *mux.Router, logsHandler *handlers.LogsHandler, withAccessLog func(http.Handler) http.Handler) {
	// Configure middleware
	withAuth := gosightauth.AuthMiddleware(logsHandler.Sys.Stores.Users)
	withLog := withAccessLog

	// Helper function to create secure handler with permission check
	secure := func(permission string, handler http.Handler) http.Handler {
		return withLog(withAuth(gosightauth.RequirePermission(permission, handler, logsHandler.Sys.Stores.Users)))
	}

	// Log query and viewing endpoints
	router.Handle("/logs",
		secure("gosight:api:logs:view", http.HandlerFunc(logsHandler.HandleLogAPI))).
		Methods("GET")

	router.Handle("/logs/stream",
		secure("gosight:api:logs:stream", http.HandlerFunc(logsHandler.HandleLogAPI))).
		Methods("GET")

	router.Handle("/logs/search",
		secure("gosight:api:logs:search", http.HandlerFunc(logsHandler.HandleLogAPI))).
		Methods("GET")

	router.Handle("/logs/sources",
		secure("gosight:api:logs:view", http.HandlerFunc(logsHandler.HandleLogAPI))).
		Methods("GET")

	router.Handle("/logs/export",
		secure("gosight:api:logs:export", http.HandlerFunc(logsHandler.HandleLogAPI))).
		Methods("POST")

	router.Handle("/logs/stats",
		secure("gosight:api:logs:view", http.HandlerFunc(logsHandler.HandleLogAPI))).
		Methods("GET")
}
