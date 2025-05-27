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
	"net/http"
	"path/filepath"

	gosightauth "github.com/aaronlmathis/gosight-server/internal/auth"
)

// withAuth returns the authentication middleware bound to the current user store.
func (s *HttpServer) withAuth() func(http.Handler) http.Handler {
	return gosightauth.AuthMiddleware(s.Sys.Stores.Users)
}

// withAccessLog wraps a handler with access logging middleware.
func (s *HttpServer) withAccessLog(h http.Handler) http.Handler {
	return gosightauth.AccessLogMiddleware(h)
}

// HandleSvelteKitApp serves the SvelteKit application without permission requirements
func (s *HttpServer) HandleSvelteKitApp(w http.ResponseWriter, r *http.Request) {
	// Serve the SvelteKit app's index.html directly from UI/build
	buildDir := "UI/build"
	indexPath := filepath.Join(buildDir, "index.html")
	http.ServeFile(w, r, indexPath)
}

// HandleSvelteKitAppWithPermission serves the SvelteKit application with permission check
func (s *HttpServer) HandleSvelteKitAppWithPermission(permission string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Serve the SvelteKit app's index.html directly from UI/build
		buildDir := "UI/build"
		indexPath := filepath.Join(buildDir, "index.html")
		http.ServeFile(w, r, indexPath)
	}
}

// setupRoutes sets up the routes for the HTTP server.
// It includes routes for static files, authentication, API, websockets, and a catch-all SvelteKit handler.
func (s *HttpServer) setupRoutes() {
	s.setupStaticRoutes()
	s.setupAuthRoutes()
	s.setupAPIRoutes()
	s.setupWebSocketRoutes()
	s.setupSvelteKitRoutes() // Catch-all for dashboard routes - must be last
}

// setupStaticRoutes sets up the static routes for the HTTP server.
// It includes routes for serving SvelteKit build assets from _app/ directory.
func (s *HttpServer) setupStaticRoutes() {
	// Serve SvelteKit build assets from UI/build/_app/
	buildDir := "UI/build"
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

// setupAuthRoutes sets up the authentication routes for the HTTP server.
// It includes only the callback routes for OAuth providers since SvelteKit handles the auth UI.
// The routes are protected by middleware that checks for injects context and trace identifiers.
// The routes are also logged for access control.
func (s *HttpServer) setupAuthRoutes() {
	withLog := s.withAccessLog

	// OAuth callback routes - these need to be handled by the server, not SvelteKit
	s.Router.Handle("/login/start", withLog(http.HandlerFunc(s.HandleLoginStart))).Methods("GET")
	s.Router.Handle("/callback", withLog(http.HandlerFunc(s.HandleCallback))).Methods("GET", "POST")

	// Note: /login, /mfa, /logout UI routes are now handled by SvelteKit
	// The actual auth logic is handled by the API routes in setupAPIRoutes()
}

// setupSvelteKitRoutes sets up routes for the SvelteKit application.
// This handles all frontend routes and lets SvelteKit's client-side router handle internal routing.
// Auth routes are publicly accessible, while dashboard routes require authentication and permissions.
func (s *HttpServer) setupSvelteKitRoutes() {
	withAuth := s.withAuth()
	withLog := s.withAccessLog

	// Public auth routes - no authentication required
	// Handle all /auth/* routes publicly
	s.Router.PathPrefix("/auth/").Handler(withLog(http.HandlerFunc(s.HandleSvelteKitApp))).Methods("GET")

	// Protected dashboard routes - require authentication and permissions
	// Handle all other routes (including /) with authentication
	s.Router.PathPrefix("/").Handler(withAuth(
		gosightauth.RequirePermission("gosight:dashboard:view",
			withLog(http.HandlerFunc(s.HandleSvelteKitAppWithPermission("gosight:dashboard:view"))),
			s.Sys.Stores.Users,
		),
	)).Methods("GET")
}

// setupAPIRoutes sets up the API routes for the HTTP server.
// It includes routes for fetching namespaces, sub-namespaces, metric names, dimensions,
// and metric data.
func (s *HttpServer) setupAPIRoutes() {
	api := s.Router.PathPrefix("/api/v1").Subrouter()

	withAuth := gosightauth.AuthMiddleware(s.Sys.Stores.Users)

	secure := func(permission string, handler http.HandlerFunc) http.Handler {
		return withAuth(gosightauth.RequirePermission(permission, handler, s.Sys.Stores.Users))
	}

	// Auth API routes for SvelteKit frontend
	withLog := s.withAccessLog
	api.Handle("/auth/providers", withLog(http.HandlerFunc(s.HandleAPIAuthProviders))).Methods("GET")
	api.Handle("/auth/login", withLog(http.HandlerFunc(s.HandleAPILogin))).Methods("POST")
	api.Handle("/auth/mfa/verify", withLog(http.HandlerFunc(s.HandleAPIMFAVerify))).Methods("POST")
	api.Handle("/auth/logout", withLog(withAuth(http.HandlerFunc(s.HandleAPILogout)))).Methods("POST")
	api.Handle("/auth/me", withLog(withAuth(http.HandlerFunc(s.HandleCurrentUser)))).Methods("GET")

	// User profile and settings endpoints
	api.Handle("/users/profile", withLog(withAuth(http.HandlerFunc(s.HandleUpdateUserProfile)))).Methods("PUT")
	api.Handle("/users/password", withLog(withAuth(http.HandlerFunc(s.HandleUpdateUserPassword)))).Methods("PUT")
	api.Handle("/users/preferences", withLog(withAuth(http.HandlerFunc(s.HandleGetUserPreferences)))).Methods("GET")
	api.Handle("/users/preferences", withLog(withAuth(http.HandlerFunc(s.HandleUpdateUserPreferences)))).Methods("PUT")
	api.Handle("/users/me", withLog(withAuth(http.HandlerFunc(s.HandleGetCompleteUser)))).Methods("GET")

	// File upload endpoints
	api.Handle("/users/avatar", withLog(withAuth(http.HandlerFunc(s.HandleUploadAvatar)))).Methods("POST")
	api.Handle("/users/avatar", withLog(withAuth(http.HandlerFunc(s.HandleDeleteAvatar)))).Methods("DELETE")
	api.Handle("/users/avatar/crop", withLog(withAuth(http.HandlerFunc(s.HandleCropAvatar)))).Methods("POST")
	api.Handle("/upload/limits", withLog(withAuth(http.HandlerFunc(s.HandleGetUploadLimits)))).Methods("GET")

	api.Handle("/network-devices", secure("gosight:dashboard:view", http.HandlerFunc(s.HandleNetworkDevicesAPI))).Methods("GET", "POST")
	api.Handle("/network-devices/{id}", secure("gosight:dashboard:view", http.HandlerFunc(s.HandleDeleteNetworkDeviceAPI))).Methods("DELETE")
	api.Handle("/network-devices/{id}", secure("gosight:dashboard:view", http.HandlerFunc(s.HandleUpdateNetworkDeviceAPI))).Methods("PUT")
	api.Handle("/network-devices/{id}/toggle", secure("gosight:dashboard:view", http.HandlerFunc(s.HandleToggleNetworkDeviceStatusAPI))).Methods("POST")

	api.Handle("/debug/cache", secure("gosight:dashboard:view", http.HandlerFunc(s.HandleCacheAudit))).Methods("GET")

	// Search
	api.Handle("/search", secure("gosight:api:search", http.HandlerFunc(s.HandleGlobalSearchAPI))).Methods("GET")

	api.Handle("/command", secure("gosight:api:command:execute", http.HandlerFunc(s.HandleCommandsAPI))).Methods("POST")

	api.Handle("/labels/values", secure("gosight:api:tags:view", http.HandlerFunc(s.HandleLabelValues))).Methods("GET")
	// Tags
	// Tag management
	api.Handle("/tags/keys", secure("gosight:api:tags:view", http.HandlerFunc(s.HandleTagKeys))).Methods("GET")

	api.Handle("/tags/values", secure("gosight:api:tags:view", http.HandlerFunc(s.HandleTagValues))).Methods("GET")
	api.Handle("/tags/{endpointID}", secure("gosight:api:tags:view", http.HandlerFunc(s.HandleGetTags))).Methods("GET")
	api.Handle("/tags/{endpointID}", secure("gosight:api:tags:set", http.HandlerFunc(s.HandleSetTags))).Methods("POST")
	api.Handle("/tags/{endpointID}", secure("gosight:api:tags:patch", http.HandlerFunc(s.HandlePatchTags))).Methods("PATCH")
	api.Handle("/tags/{endpointID}/{key}", secure("gosight:api:tags:delete", http.HandlerFunc(s.HandleDeleteTag))).Methods("DELETE")

	// Endpoint APIs
	api.Handle("/endpoints", secure("gosight:api:endpoints:view", http.HandlerFunc(s.HandleEndpointsAPI))).Methods("GET")
	api.Handle("/endpoints/{endpointType}", secure("gosight:api:endpoints:view", http.HandlerFunc(s.HandleEndpointsByTypeAPI))).Methods("GET")

	// Logs
	api.Handle("/logs", secure("gosight:api:logs:view", http.HandlerFunc(s.HandleLogAPI))).Methods("GET")
	api.Handle("/logs/latest", secure("gosight:api:logs:view", http.HandlerFunc(s.HandleRecentLogs))).Methods("GET")

	// Events and Alerts
	api.Handle("/events", secure("gosight:api:events:view", http.HandlerFunc(s.HandleEventsAPI))).Methods("GET")

	api.Handle("/alerts", secure("gosight:api:events:view", http.HandlerFunc(s.HandleCreateAlertRuleAPI))).Methods("POST") // TODO: Permissions
	api.Handle("/alerts/summary", secure("gosight:api:events:view", http.HandlerFunc(s.HandleAlertsSummaryAPI))).Methods("GET")
	api.Handle("/alerts/rules", secure("gosight:api:events:view", http.HandlerFunc(s.HandleAlertRulesAPI))).Methods("GET")
	api.Handle("/alerts/active", secure("gosight:api:events:view", http.HandlerFunc(s.HandleActiveAlertsAPI))).Methods("GET")
	api.Handle("/alerts", secure("gosight:api:events:view", http.HandlerFunc(s.HandleAlertsAPI))).Methods("GET")
	api.Handle("/alerts/{id}/context", secure("gosight:api:events:view", http.HandlerFunc(s.HandleAlertContext))).Methods("GET")

	// Metrics and queries
	api.Handle("/query", secure("gosight:api:metrics:query", http.HandlerFunc(s.HandleAPIQuery))).Methods("GET")
	api.Handle("/exportquery", secure("gosight:api:metrics:export", http.HandlerFunc(s.HandleExportQuery))).Methods("GET")

	// Metadata discovery endpoints
	api.Handle("/metrics", secure("gosight:api:metrics:meta", http.HandlerFunc(s.GetNamespaces))).Methods("GET")
	api.Handle("/metrics/{namespace}/{sub}/{metric}/dimensions", secure("gosight:api:metrics:meta", http.HandlerFunc(s.GetMetricDimensions))).Methods("GET")
	api.Handle("/metrics/{namespace}/{sub}/{metric}/labels", secure("gosight:api:metrics:meta", http.HandlerFunc(s.GetMetricDimensions))).Methods("GET")
	api.Handle("/metrics/{namespace}/{sub}/{metric}/data", secure("gosight:api:metrics:read", http.HandlerFunc(s.GetMetricData))).Methods("GET")
	api.Handle("/metrics/{namespace}/{sub}/{metric}/latest", secure("gosight:api:metrics:read", http.HandlerFunc(s.GetMetricLatest))).Methods("GET")

	api.Handle("/metrics/{namespace}/{sub}", secure("gosight:api:metrics:meta", http.HandlerFunc(s.GetMetricNames))).Methods("GET")
	api.Handle("/metrics/{namespace}", secure("gosight:api:metrics:meta", http.HandlerFunc(s.GetSubNamespaces))).Methods("GET")

}

// setupWebSocketRoutes sets up the websocket routes for the HTTP server.
// It includes routes for websocket connections for metrics, alerts, events, logs, commands, and processes.
func (s *HttpServer) setupWebSocketRoutes() {
	withAuth := s.withAuth()

	s.Router.Handle("/ws/metrics", withAuth(http.HandlerFunc(s.Sys.WSHub.Metrics.ServeWS)))
	s.Router.Handle("/ws/alerts", withAuth(http.HandlerFunc(s.Sys.WSHub.Alerts.ServeWS)))
	s.Router.Handle("/ws/events", withAuth(http.HandlerFunc(s.Sys.WSHub.Events.ServeWS)))
	s.Router.Handle("/ws/logs", withAuth(http.HandlerFunc(s.Sys.WSHub.Logs.ServeWS)))
	s.Router.Handle("/ws/command", withAuth(http.HandlerFunc(s.Sys.WSHub.Commands.ServeWS)))
	s.Router.Handle("/ws/process", withAuth(http.HandlerFunc(s.Sys.WSHub.Processes.ServeWS)))
}
