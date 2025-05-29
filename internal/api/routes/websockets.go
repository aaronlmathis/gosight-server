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
// This file contains WebSocket connection related routes and handlers.
package routes

import (
	"net/http"

	"github.com/aaronlmathis/gosight-server/internal/api/handlers"
	gosightauth "github.com/aaronlmathis/gosight-server/internal/auth"
	"github.com/aaronlmathis/gosight-server/internal/sys"
	"github.com/gorilla/mux"
)

// SetupWebSocketRoutes configures WebSocket connection API routes.
// It sets up endpoints for real-time data streaming over WebSocket connections
// with appropriate middleware for authentication and logging.
//
// Protected routes:
//   - GET /ws/metrics - WebSocket for metrics streaming (requires authentication)
//   - GET /ws/alerts - WebSocket for alerts streaming (requires authentication)
//   - GET /ws/events - WebSocket for events streaming (requires authentication)
//   - GET /ws/logs - WebSocket for logs streaming (requires authentication)
//   - GET /ws/command - WebSocket for command execution (requires authentication)
//   - GET /ws/process - WebSocket for process monitoring (requires authentication)
//
// All WebSocket endpoints require authentication but use the withAuth middleware
// rather than permission-based authorization, as the specific permissions are
// handled within the WebSocket handlers themselves based on the requested data.
func SetupWebSocketRoutes(router *mux.Router, sys *sys.SystemContext) {
	// Create handler
	wsHandler := handlers.NewWebSocketsHandler(sys)

	// Configure middleware - WebSocket routes use auth middleware directly
	withAuth := gosightauth.AuthMiddleware(sys.Stores.Users)

	// WebSocket endpoints for real-time data streaming
	router.Handle("/ws/metrics",
		withAuth(http.HandlerFunc(wsHandler.HandleMetricsWS))).
		Methods("GET")

	router.Handle("/ws/alerts",
		withAuth(http.HandlerFunc(wsHandler.HandleAlertsWS))).
		Methods("GET")

	router.Handle("/ws/events",
		withAuth(http.HandlerFunc(wsHandler.HandleEventsWS))).
		Methods("GET")

	router.Handle("/ws/logs",
		withAuth(http.HandlerFunc(wsHandler.HandleLogsWS))).
		Methods("GET")

	router.Handle("/ws/command",
		withAuth(http.HandlerFunc(wsHandler.HandleCommandWS))).
		Methods("GET")

	router.Handle("/ws/process",
		withAuth(http.HandlerFunc(wsHandler.HandleProcessWS))).
		Methods("GET")
}
