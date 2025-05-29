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
// This file contains user management related routes and handlers.
package routes

import (
	"net/http"

	"github.com/aaronlmathis/gosight-server/internal/api/handlers"
	gosightauth "github.com/aaronlmathis/gosight-server/internal/auth"
	"github.com/gorilla/mux"
)

// SetupUserRoutes configures user management API routes.
// It sets up endpoints for user CRUD operations, user settings, and profile management
// with appropriate middleware for authentication, authorization, and logging.
//
// Protected routes:
//   - GET /users - List all users (requires gosight:api:users:list permission)
//   - POST /users - Create a new user (requires gosight:api:users:create permission)
//   - GET /users/{id} - Get user by ID (requires gosight:api:users:view permission)
//   - PUT /users/{id} - Update user (requires gosight:api:users:update permission)
//   - DELETE /users/{id} - Delete user (requires gosight:api:users:delete permission)
//   - POST /users/{id}/password - Change user password (requires gosight:api:users:password permission)
//   - GET /users/{id}/settings - Get user settings (requires gosight:api:users:settings:view permission)
//   - PUT /users/{id}/settings - Update user settings (requires gosight:api:users:settings:update permission)
//   - GET /users/preferences - Get user preferences (requires gosight:api:users:preferences:view permission)
//   - PUT /users/preferences - Update user preferences (requires gosight:api:users:preferences:update permission)
func SetupUserRoutes(router *mux.Router, usersHandler *handlers.UsersHandler, withAccessLog func(http.Handler) http.Handler) {
	// Configure middleware
	withAuth := gosightauth.AuthMiddleware(usersHandler.Sys.Stores.Users)

	// Helper function to create secure handler with permission check
	secure := func(permission string, handler http.Handler) http.Handler {
		return withAccessLog(withAuth(gosightauth.RequirePermission(permission, handler, usersHandler.Sys.Stores.Users)))
	}

	// User CRUD operations
	router.Handle("/users",
		secure("gosight:api:users:list", http.HandlerFunc(usersHandler.HandleAPIUsers))).
		Methods("GET")

	router.Handle("/users",
		secure("gosight:api:users:create", http.HandlerFunc(usersHandler.HandleAPIUserCreate))).
		Methods("POST")

	router.Handle("/users/{id}",
		secure("gosight:api:users:view", http.HandlerFunc(usersHandler.HandleAPIUser))).
		Methods("GET")

	router.Handle("/users/{id}",
		secure("gosight:api:users:update", http.HandlerFunc(usersHandler.HandleAPIUserUpdate))).
		Methods("PUT")

	router.Handle("/users/{id}",
		secure("gosight:api:users:delete", http.HandlerFunc(usersHandler.HandleAPIUserDelete))).
		Methods("DELETE")

	// User password management
	router.Handle("/users/{id}/password",
		secure("gosight:api:users:password", http.HandlerFunc(usersHandler.HandleAPIUserPasswordChange))).
		Methods("POST")

	// User settings management
	router.Handle("/users/{id}/settings",
		secure("gosight:api:users:settings:view", http.HandlerFunc(usersHandler.HandleAPIUserSettings))).
		Methods("GET")

	router.Handle("/users/{id}/settings",
		secure("gosight:api:users:settings:update", http.HandlerFunc(usersHandler.HandleAPIUserSettingsUpdate))).
		Methods("PUT")

	// User profile management (for current user)
	router.Handle("/users/profile",
		secure("gosight:api:users:profile:update", http.HandlerFunc(usersHandler.HandleUpdateUserProfile))).
		Methods("PUT")

	// User password management (for current user)
	router.Handle("/users/password",
		secure("gosight:api:users:password", http.HandlerFunc(usersHandler.HandleUpdateUserPassword))).
		Methods("PUT")

	// User avatar management (for current user)
	router.Handle("/users/avatar",
		secure("gosight:api:users:avatar:upload", http.HandlerFunc(usersHandler.HandleUploadAvatar))).
		Methods("POST")

	router.Handle("/users/avatar/crop",
		secure("gosight:api:users:avatar:crop", http.HandlerFunc(usersHandler.HandleCropAvatar))).
		Methods("POST")

	router.Handle("/users/avatar",
		secure("gosight:api:users:avatar:delete", http.HandlerFunc(usersHandler.HandleDeleteAvatar))).
		Methods("DELETE")

	// Upload limits (for current user)
	router.Handle("/upload/limits",
		secure("gosight:api:upload:limits", http.HandlerFunc(usersHandler.HandleGetUploadLimits))).
		Methods("GET")

	// User preferences management (for current user)
	router.Handle("/users/preferences",
		secure("gosight:api:users:preferences:view", http.HandlerFunc(usersHandler.HandleGetUserPreferences))).
		Methods("GET")

	router.Handle("/users/preferences",
		secure("gosight:api:users:preferences:update", http.HandlerFunc(usersHandler.HandleUpdateUserPreferences))).
		Methods("PUT")
}
