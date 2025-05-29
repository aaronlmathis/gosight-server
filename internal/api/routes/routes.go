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
// This package organizes API endpoints by functional domain into separate files
// for better maintainability and code organization.
//
// The main entry point is SetupAllRoutes which registers all route handlers
// with the provided router. Individual route setup functions are organized
// by domain:
//   - Authentication: auth.go
//   - User Management: users.go
//   - File Operations: files.go
//   - Metrics & Queries: metrics.go
//   - Alerts & Events: alerts.go
//   - Endpoint Management: endpoints.go
//   - Log Management: logs.go
//   - Tag Management: tags.go
//   - Device Management: devices.go
//   - Search & Commands: search.go
//   - Telemetry Data: telemetry.go
//   - Resource Management: resources.go
//   - Debug Utilities: debug.go
//   - WebSocket Connections: websockets.go
package routes

import (
	"net/http"
	"strings"

	"github.com/aaronlmathis/gosight-server/internal/api/handlers"
	"github.com/aaronlmathis/gosight-server/internal/sys"
	"github.com/gorilla/mux"
)

// APIVersion represents an API version configuration (for compatibility)
type APIVersion struct {
	Version    string
	IsDefault  bool
	IsEnabled  bool
}

// GetSupportedAPIVersionsFromConfig converts config API versions to legacy format
func GetSupportedAPIVersionsFromConfig(sys *sys.SystemContext) []APIVersion {
	apiConfig := sys.Cfg.API
	
	// If no API config is provided, use defaults
	if len(apiConfig.SupportedVersions) == 0 {
		return []APIVersion{
			{Version: "v1", IsDefault: true, IsEnabled: true},
		}
	}

	var versions []APIVersion
	for _, versionConfig := range apiConfig.SupportedVersions {
		if versionConfig.Enabled {
			isDefault := versionConfig.Version == apiConfig.DefaultVersion
			versions = append(versions, APIVersion{
				Version:   versionConfig.Version,
				IsDefault: isDefault,
				IsEnabled: versionConfig.Enabled,
			})
		}
	}

	// Ensure at least one version is marked as default
	if len(versions) > 0 {
		hasDefault := false
		for _, v := range versions {
			if v.IsDefault {
				hasDefault = true
				break
			}
		}
		if !hasDefault {
			versions[0].IsDefault = true
		}
	}

	return versions
}

// SetupAllRoutes configures all API routes for the GoSight server.
// This function serves as the main entry point for route configuration,
// organizing routes by functional domain and setting up appropriate middleware.
//
// The routes are organized into the following groups:
//   - Authentication and session management
//   - User management and profile operations
//   - File upload and management
//   - Metrics querying and metadata discovery
//   - Alert and event management
//   - Monitoring endpoint configuration
//   - Log querying and streaming
//   - Tag management and assignment
//   - Network device discovery and monitoring
//   - Search functionality and command execution
//   - Telemetry data ingestion and processing
//   - Resource management and discovery
//   - Debug and development utilities
//   - WebSocket connections for real-time data
//
// Parameters:
//   - router: The main mux.Router instance to register all routes with
//   - sys: The SystemContext instance containing all system dependencies
//   - withAccessLog: Access logging middleware function
func SetupAllRoutes(router *mux.Router, sys *sys.SystemContext, withAccessLog func(http.Handler) http.Handler) {
	// Create API subrouter for all API endpoints
	apiRouter := router.PathPrefix("/api").Subrouter()

	// Get supported API versions from configuration
	supportedVersions := GetSupportedAPIVersionsFromConfig(sys)

	// Setup routes for each enabled API version
	for _, version := range supportedVersions {
		if version.IsEnabled {
			setupVersionedRoutes(apiRouter, version.Version, sys, withAccessLog)
		}
	}

	// Setup default version redirect if enabled in config
	if sys.Cfg.API.EnableVersionRedirect {
		setupDefaultVersionRedirect(apiRouter, sys.Cfg.API.DefaultVersion)
	}

	// Setup WebSocket routes (these go on the main router, not API subrouter)
	SetupWebSocketRoutes(router, sys)
}

// setupVersionedRoutes configures routes for a specific API version
func setupVersionedRoutes(apiRouter *mux.Router, version string, sys *sys.SystemContext, withAccessLog func(http.Handler) http.Handler) {
	// Create version-specific subrouter
	versionRouter := apiRouter.PathPrefix("/" + version).Subrouter()

	// Initialize handlers with system context
	// Initialize handlers with system context
	authHandler := &handlers.AuthHandler{Sys: sys}
	usersHandler := &handlers.UsersHandler{Sys: sys}
	alertsHandler := &handlers.AlertsHandler{Sys: sys}
	endpointsHandler := &handlers.EndpointsHandler{Sys: sys}
	//agentsHandler := &handlers.AgentsHandler{Sys: sys}
	logsHandler := &handlers.LogsHandler{Sys: sys}
	metricsHandler := &handlers.MetricsHandler{Sys: sys}
	//searchHandler := &handlers.SearchHandler{Sys: sys}
	commandsHandler := &handlers.CommandsHandler{Sys: sys}
	tagsHandler := &handlers.TagsHandler{Sys: sys}
	labelsHandler := &handlers.LabelsHandler{Sys: sys}
	debugHandler := &handlers.DebugHandler{Sys: sys}
	telemetryHandler := &handlers.TelemetryHandler{Sys: sys}

	// Setup routes based on API version
	switch version {
	case "v1":
		setupV1Routes(versionRouter, sys, authHandler, usersHandler, alertsHandler, 
			endpointsHandler, logsHandler, metricsHandler, commandsHandler, 
			tagsHandler, labelsHandler, debugHandler, telemetryHandler, withAccessLog)
	default:
		// For future versions, we can add more cases here
		setupV1Routes(versionRouter, sys, authHandler, usersHandler, alertsHandler, 
			endpointsHandler, logsHandler, metricsHandler, commandsHandler, 
			tagsHandler, labelsHandler, debugHandler, telemetryHandler, withAccessLog)
	}
}

// setupV1Routes configures all routes for API version 1
func setupV1Routes(
	router *mux.Router, 
	sys *sys.SystemContext,
	authHandler *handlers.AuthHandler,
	usersHandler *handlers.UsersHandler,
	alertsHandler *handlers.AlertsHandler,
	endpointsHandler *handlers.EndpointsHandler,
	logsHandler *handlers.LogsHandler,
	metricsHandler *handlers.MetricsHandler,
	commandsHandler *handlers.CommandsHandler,
	tagsHandler *handlers.TagsHandler,
	labelsHandler *handlers.LabelsHandler,
	debugHandler *handlers.DebugHandler,
	telemetryHandler *handlers.TelemetryHandler,
	withAccessLog func(http.Handler) http.Handler,
) {
	// Setup authentication routes (login, logout, etc.)
	SetupAuthRoutes(router, authHandler, withAccessLog)

	// Setup user management routes
	SetupUserRoutes(router, usersHandler, withAccessLog)

	// Setup file upload and management routes
	SetupFileRoutes(router, sys, withAccessLog)

	// Setup metrics and query routes
	SetupMetricsRoutes(router, metricsHandler, withAccessLog)

	// Setup alerts and events routes
	SetupAlertsRoutes(router, alertsHandler, withAccessLog)

	// Setup endpoint management routes
	SetupEndpointsRoutes(router, endpointsHandler, withAccessLog)

	// Setup log management routes
	SetupLogsRoutes(router, logsHandler, withAccessLog)

	// Setup tag management routes
	SetupTagsRoutes(router, tagsHandler, withAccessLog)

	// Setup device management routes (uses agents handler)
	//SetupDevicesRoutes(router, agentsHandler, withAccessLog)

	// Setup search and command routes
	//SetupSearchRoutes(router, searchHandler, commandsHandler, withAccessLog)

	// Setup telemetry data ingestion routes
	SetupTelemetryRoutes(router, telemetryHandler, withAccessLog)

	// Setup resource management routes
	SetupResourceRoutes(router, sys, withAccessLog)

	// Setup labels routes
	SetupLabelsRoutes(router, labelsHandler, withAccessLog)

	// Setup commands routes
	SetupCommandsRoutes(router, commandsHandler, withAccessLog)

	// Setup debug routes (should be conditional based on environment)
	if sys.Cfg.Debug.Enabled {
		SetupDebugRoutes(router, debugHandler, withAccessLog)
	}
}

// setupDefaultVersionRedirect sets up a redirect from /api/ to /api/{defaultVersion}/
func setupDefaultVersionRedirect(apiRouter *mux.Router, defaultVersion string) {
	// Redirect unversioned API calls to the default version
	apiRouter.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only redirect if the path doesn't already contain a version
		path := r.URL.Path
		if path == "/api/" || path == "/api" {
			http.Redirect(w, r, "/api/"+defaultVersion+"/", http.StatusMovedPermanently)
			return
		}
		
		// For other paths under /api/ that don't start with a version, redirect
		if !isVersionedPath(path) {
			newPath := "/api/" + defaultVersion + path[4:] // Remove "/api" and add version
			http.Redirect(w, r, newPath, http.StatusMovedPermanently)
			return
		}
		
		// If it's already versioned, let it fall through
		http.NotFound(w, r)
	})
}

// isVersionedPath checks if the API path already contains a version
func isVersionedPath(path string) bool {
	// Remove /api/ prefix and check if next segment looks like a version (v1, v2, etc.)
	if len(path) < 5 || !strings.HasPrefix(path, "/api/") {
		return false
	}
	
	remainder := path[5:] // Remove "/api/"
	parts := strings.Split(remainder, "/")
	if len(parts) > 0 && strings.HasPrefix(parts[0], "v") {
		return true
	}
	return false
}
