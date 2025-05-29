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
// It organizes API endpoints by functional domain and applies appropriate
// middleware for authentication, authorization, and logging.
package routes

import (
	"net/http"

	"github.com/aaronlmathis/gosight-server/internal/api/handlers"
	gosightauth "github.com/aaronlmathis/gosight-server/internal/auth"
	"github.com/gorilla/mux"
)

// SetupAuthRoutes configures authentication-related API routes.
// It sets up both public endpoints (login, providers, OAuth callbacks) and protected endpoints
// (logout, user profile) with appropriate middleware for logging and authentication.
//
// Public routes:
//   - GET /auth/providers - List available authentication providers
//   - POST /auth/login - Authenticate user credentials
//   - POST /auth/mfa/verify - Verify multi-factor authentication
//   - GET /login/start - Start OAuth authentication flow
//   - GET /callback - OAuth callback handler
//   - POST /callback - OAuth callback handler (alternative method)
//
// Protected routes:
//   - POST /auth/logout - End user session
//   - GET /auth/me - Get current user information
func SetupAuthRoutes(router *mux.Router, authHandler *handlers.AuthHandler, withAccessLog func(http.Handler) http.Handler) {
	// Configure middleware
	withAuth := gosightauth.AuthMiddleware(authHandler.Sys.Stores.Users)

	// Public authentication endpoints
	router.Handle("/auth/providers",
		withAccessLog(http.HandlerFunc(authHandler.HandleAPIAuthProviders))).
		Methods("GET")

	router.Handle("/auth/login",
		withAccessLog(http.HandlerFunc(authHandler.HandleAPILogin))).
		Methods("POST")

	router.Handle("/auth/mfa/verify",
		withAccessLog(http.HandlerFunc(authHandler.HandleAPIMFAVerify))).
		Methods("POST")

	// OAuth callback routes - these need to be handled by the server, not SvelteKit
	router.Handle("/auth/login/start",
		withAccessLog(http.HandlerFunc(authHandler.HandleLoginStart))).
		Methods("GET")

	router.Handle("/callback",
		withAccessLog(http.HandlerFunc(authHandler.HandleCallback))).
		Methods("GET", "POST")

	// Protected authentication endpoints
	router.Handle("/auth/logout",
		withAccessLog(withAuth(http.HandlerFunc(authHandler.HandleAPILogout)))).
		Methods("POST")

	router.Handle("/auth/me",
		withAccessLog(withAuth(http.HandlerFunc(authHandler.HandleCurrentUser)))).
		Methods("GET")
}
