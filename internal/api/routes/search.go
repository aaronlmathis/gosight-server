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
// This file contains search and command execution related routes and handlers.
package routes

import (
	"net/http"

	"github.com/aaronlmathis/gosight-server/internal/api/handlers"
	gosightauth "github.com/aaronlmathis/gosight-server/internal/auth"
	"github.com/aaronlmathis/gosight-server/internal/sys"
	"github.com/gorilla/mux"
)

// SetupSearchRoutes configures search and command execution API routes.
// It sets up endpoints for global search, command execution, and search indexing
// with appropriate middleware for authentication, authorization, and logging.
//
// Protected routes:
//   - GET /search - Global search across all resources (requires gosight:api:search:query permission)
//   - POST /search/index - Rebuild search index (requires gosight:api:search:admin permission)
//   - GET /search/suggestions - Get search suggestions (requires gosight:api:search:query permission)
//   - POST /commands - Execute command (requires gosight:api:commands:execute permission)
//   - GET /commands - List available commands (requires gosight:api:commands:view permission)
//   - GET /commands/{id} - Get command by ID (requires gosight:api:commands:view permission)
//   - GET /commands/{id}/status - Get command execution status (requires gosight:api:commands:view permission)
func SetupSearchRoutes(router *mux.Router, sys *sys.SystemContext) {
	// Create handler
	searchHandler := handlers.NewSearchHandler(sys)

	// Configure middleware
	withAuth := gosightauth.AuthMiddleware(sys.Stores.Users)
	withLog := func(handler http.Handler) http.Handler {
		// TODO: Access logging middleware
		return handler
	}

	// Helper function to create secure handler with permission check
	secure := func(permission string, handler http.Handler) http.Handler {
		return withLog(withAuth(gosightauth.RequirePermission(permission, handler, sys.Stores.Users)))
	}

	// Search endpoints
	router.Handle("/search",
		secure("gosight:api:search:query", http.HandlerFunc(searchHandler.HandleAPISearch))).
		Methods("GET")

	router.Handle("/search/index",
		secure("gosight:api:search:admin", http.HandlerFunc(searchHandler.HandleAPISearchIndex))).
		Methods("POST")

	router.Handle("/search/suggestions",
		secure("gosight:api:search:query", http.HandlerFunc(searchHandler.HandleAPISearchSuggestions))).
		Methods("GET")

	// Command execution endpoints
	router.Handle("/commands",
		secure("gosight:api:commands:execute", http.HandlerFunc(searchHandler.HandleAPICommandExecute))).
		Methods("POST")

	router.Handle("/commands",
		secure("gosight:api:commands:view", http.HandlerFunc(searchHandler.HandleAPICommands))).
		Methods("GET")

	router.Handle("/commands/{id}",
		secure("gosight:api:commands:view", http.HandlerFunc(searchHandler.HandleAPICommand))).
		Methods("GET")

	router.Handle("/commands/{id}/status",
		secure("gosight:api:commands:view", http.HandlerFunc(searchHandler.HandleAPICommandStatus))).
		Methods("GET")
}
