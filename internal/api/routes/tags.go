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
// This file contains tag management related routes and handlers.
package routes

import (
	"net/http"

	"github.com/aaronlmathis/gosight-server/internal/api/handlers"
	gosightauth "github.com/aaronlmathis/gosight-server/internal/auth"
	"github.com/gorilla/mux"
)

// SetupTagsRoutes configures tag management API routes.
// It sets up endpoints for tag CRUD operations, tag assignment, and tag filtering
// with appropriate middleware for authentication, authorization, and logging.
//
// Protected routes:
//   - GET /tags - List all tags (requires gosight:api:tags:view permission)
//   - POST /tags - Create new tag (requires gosight:api:tags:create permission)
//   - GET /tags/{id} - Get tag by ID (requires gosight:api:tags:view permission)
//   - PUT /tags/{id} - Update tag (requires gosight:api:tags:update permission)
//   - DELETE /tags/{id} - Delete tag (requires gosight:api:tags:delete permission)
//   - GET /tags/search - Search tags (requires gosight:api:tags:search permission)
//   - POST /tags/{id}/assign - Assign tag to resource (requires gosight:api:tags:assign permission)
//   - DELETE /tags/{id}/assign - Remove tag from resource (requires gosight:api:tags:assign permission)
func SetupTagsRoutes(router *mux.Router, tagsHandler *handlers.TagsHandler, withAccessLog func(http.Handler) http.Handler) {

	// Configure middleware
	withAuth := gosightauth.AuthMiddleware(tagsHandler.Sys.Stores.Users)

	// Helper function to create secure handler with permission check
	secure := func(permission string, handler http.Handler) http.Handler {
		return withAccessLog(withAuth(gosightauth.RequirePermission(permission, handler, tagsHandler.Sys.Stores.Users)))
	}

	// Tag CRUD operations
	router.Handle("/tags",
		secure("gosight:api:tags:view", http.HandlerFunc(tagsHandler.HandleAPITags))).
		Methods("GET")

	router.Handle("/tags",
		secure("gosight:api:tags:create", http.HandlerFunc(tagsHandler.HandleAPITagCreate))).
		Methods("POST")

	router.Handle("/tags/{id}",
		secure("gosight:api:tags:view", http.HandlerFunc(tagsHandler.HandleAPITag))).
		Methods("GET")

	router.Handle("/tags/{id}",
		secure("gosight:api:tags:update", http.HandlerFunc(tagsHandler.HandleAPITagUpdate))).
		Methods("PUT")

	router.Handle("/tags/{id}",
		secure("gosight:api:tags:delete", http.HandlerFunc(tagsHandler.HandleAPITagDelete))).
		Methods("DELETE")

	// Tag search and assignment operations
	router.Handle("/tags/search",
		secure("gosight:api:tags:search", http.HandlerFunc(tagsHandler.HandleAPITagSearch))).
		Methods("GET")

	router.Handle("/tags/{id}/assign",
		secure("gosight:api:tags:assign", http.HandlerFunc(tagsHandler.HandleAPITagAssign))).
		Methods("POST")

	router.Handle("/tags/{id}/assign",
		secure("gosight:api:tags:assign", http.HandlerFunc(tagsHandler.HandleAPITagUnassign))).
		Methods("DELETE")
}
