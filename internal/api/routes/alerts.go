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
// This file contains alerts and events related routes and handlers.
package routes

import (
	"net/http"

	"github.com/aaronlmathis/gosight-server/internal/api/handlers"
	gosightauth "github.com/aaronlmathis/gosight-server/internal/auth"
	"github.com/gorilla/mux"
)

// SetupAlertsRoutes configures alerts and events API routes.
// It sets up endpoints for alert management, event handling, and alert context
// with appropriate middleware for authentication, authorization, and logging.
//
// Protected routes:
//   - GET /alerts - List alerts (requires gosight:api:events:view permission)
//   - POST /alerts - Create alert (requires gosight:api:events:create permission)
//   - GET /alerts/{id} - Get alert by ID (requires gosight:api:events:view permission)
//   - PUT /alerts/{id} - Update alert (requires gosight:api:events:update permission)
//   - DELETE /alerts/{id} - Delete alert (requires gosight:api:events:delete permission)
//   - GET /alerts/{id}/context - Get alert context (requires gosight:api:events:view permission)
//   - GET /events - List events (requires gosight:api:events:view permission)
//   - POST /events - Create event (requires gosight:api:events:create permission)
//   - GET /events/{id} - Get event by ID (requires gosight:api:events:view permission)
func SetupAlertsRoutes(router *mux.Router, alertsHandler *handlers.AlertsHandler, withAccessLog func(http.Handler) http.Handler) {
	// Configure middleware
	withAuth := gosightauth.AuthMiddleware(alertsHandler.Sys.Stores.Users)

	// Helper function to create secure handler with permission check
	secure := func(permission string, handler http.Handler) http.Handler {
		return withAccessLog(withAuth(gosightauth.RequirePermission(permission, handler, alertsHandler.Sys.Stores.Users)))
	}

	// Alert management endpoints
	router.Handle("/alerts",
		secure("gosight:api:alerts:view", http.HandlerFunc(alertsHandler.HandleAlertsAPI))).
		Methods("GET")

	router.Handle("/alerts/active",
		secure("gosight:api:alerts:view", http.HandlerFunc(alertsHandler.HandleActiveAlertsAPI))).
		Methods("GET")

	router.Handle("/alerts/rules",
		secure("gosight:api:alerts:view", http.HandlerFunc(alertsHandler.HandleAlertRulesAPI))).
		Methods("GET")

	router.Handle("/alerts/rules",
		secure("gosight:api:alerts:create", http.HandlerFunc(alertsHandler.HandleCreateAlertRuleAPI))).
		Methods("POST")

	router.Handle("/alerts/summary",
		secure("gosight:api:alerts:view", http.HandlerFunc(alertsHandler.HandleAlertsSummaryAPI))).
		Methods("GET")

	router.Handle("/alerts/{id}/context",
		secure("gosight:api:alerts:view", http.HandlerFunc(alertsHandler.HandleAlertContext))).
		Methods("GET")

	// Event management endpoints (placeholder for future implementation)
	// Events are currently handled by the alerts system but may be separated later
}
