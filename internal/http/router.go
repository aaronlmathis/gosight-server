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

package httpserver

import (
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/aaronlmathis/gosight-server/internal/api/routes"
	gosightauth "github.com/aaronlmathis/gosight-server/internal/auth"
)

// withAuth returns the authentication middleware bound to the current user store.
func (s *HttpServer) withAuth() func(http.Handler) http.Handler {
	return gosightauth.AuthMiddleware(s.Sys.Stores.Users)
}

// withAccessLog wraps a handler with access logging middleware.
func (s *HttpServer) withAccessLog(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("ACCESS LOG: %s %s\n", r.Method, r.URL.Path)
		h.ServeHTTP(w, r)
	})
}

// HandleSvelteKitApp serves the SvelteKit application without permission requirements
func (s *HttpServer) HandleSvelteKitApp(w http.ResponseWriter, r *http.Request) {
	// Serve the SvelteKit app's index.html directly from UI/build
	buildDir := "web/build"
	indexPath := filepath.Join(buildDir, "index.html")
	http.ServeFile(w, r, indexPath)
}

// HandleSvelteKitAppWithPermission serves the SvelteKit application with permission check
func (s *HttpServer) HandleSvelteKitAppWithPermission(permission string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Serve the SvelteKit app's index.html directly from UI/build
		buildDir := "web/build"
		indexPath := filepath.Join(buildDir, "index.html")
		http.ServeFile(w, r, indexPath)
	}
}

// setupRoutes sets up the routes for the HTTP server.
// It includes routes for static files, API, websockets, and a catch-all SvelteKit handler.
func (s *HttpServer) setupRoutes() {
	s.setupAPIRoutes()
	s.setupWebSocketRoutes()
	s.setupSvelteKitRoutes() // Register SvelteKit routes before static routes
	s.setupStaticRoutes()    // Static routes must come after SvelteKit routes to avoid conflicts
}

// setupStaticRoutes sets up the static routes for the HTTP server.
// It includes routes for serving SvelteKit build assets from _app/ directory.
func (s *HttpServer) setupStaticRoutes() {
	// Serve SvelteKit build assets from UI/build/_app/
	buildDir := "web/build"
	staticFS := http.FileServer(http.Dir(buildDir))

	// Serve SvelteKit build assets like _app/immutable/ and other static files
	s.Router.PathPrefix("/_app/").Handler(http.StripPrefix("/", staticFS))

	// Serve favicon and other root-level static files from the build directory
	s.Router.Handle("/favicon.png", http.StripPrefix("/", staticFS)).Methods("GET")

	// Also serve any static images if they exist from the legacy web directory
	if s.Sys.Cfg.Web.StaticDir != "" {
		legacyStaticFS := http.FileServer(http.Dir(s.Sys.Cfg.Web.StaticDir))
		s.Router.PathPrefix("/images/").Handler(http.StripPrefix("/images/", legacyStaticFS))
	}

	// Serve uploaded files (avatars, etc.)
	uploadsFS := http.FileServer(http.Dir("uploads"))
	s.Router.PathPrefix("/uploads/").Handler(http.StripPrefix("/uploads/", uploadsFS))
}

// setupSvelteKitRoutes sets up routes for the SvelteKit application.
// This handles all frontend routes and lets SvelteKit's client-side router handle internal routing.
// Auth routes are publicly accessible, while dashboard routes require authentication and permissions.
func (s *HttpServer) setupSvelteKitRoutes() {
	withAuth := s.withAuth()
	withLog := s.withAccessLog

	// Public auth routes - no authentication required
	fmt.Println("Registering /auth/ route")
	s.Router.PathPrefix("/auth/").Handler(withLog(http.HandlerFunc(s.HandleSvelteKitApp))).Methods("GET")

	// Root route - redirect to dashboard if authenticated, otherwise to login
	fmt.Println("Registering root route")
	s.Router.Handle("/", withLog(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if user is authenticated
		token, err := gosightauth.GetSessionToken(r)
		if err != nil || token == "" {
			http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
			return
		}

		// Validate token
		_, err = gosightauth.ValidateToken(token)
		if err != nil {
			http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
			return
		}

		// User is authenticated, serve the main app
		s.HandleSvelteKitApp(w, r)
	}))).Methods("GET")

	// Protected dashboard routes - require authentication and permissions
	// Use more specific patterns to avoid catching everything
	fmt.Println("Registering dashboard routes")
	dashboardPaths := []string{"/dashboard", "/settings", "/alerts", "/metrics", "/logs", "/users", "/agents", "/system"}
	for _, path := range dashboardPaths {
		s.Router.PathPrefix(path).Handler(withAuth(
			gosightauth.RequirePermission("gosight:dashboard:view",
				withLog(http.HandlerFunc(s.HandleSvelteKitAppWithPermission("gosight:dashboard:view"))),
				s.Sys.Stores.Users,
			),
		)).Methods("GET")
	}
}

// setupAPIRoutes sets up the API routes for the HTTP server using the modular routes package.
// This replaces the monolithic route definition with organized, domain-specific route files.
// Routes are now organized by functional area for better maintainability.
func (s *HttpServer) setupAPIRoutes() {
	// Use the new modular route setup from the routes package
	routes.SetupAllRoutes(s.Router, s.Sys, s.withAccessLog)
}

// setupWebSocketRoutes sets up the websocket routes for the HTTP server.
// This is now handled by the modular routes package but we keep this function
// for backward compatibility and clear separation of concerns.
func (s *HttpServer) setupWebSocketRoutes() {
	// WebSocket routes are now handled in the routes.SetupAllRoutes function
	// This function is kept for clarity but the actual setup is done there
}
