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
// This file contains telemetry data ingestion routes and handlers.
package routes

import (
	"net/http"

	"github.com/aaronlmathis/gosight-server/internal/api/handlers"
	gosightauth "github.com/aaronlmathis/gosight-server/internal/auth"
	"github.com/gorilla/mux"
)

// SetupTelemetryRoutes configures telemetry data ingestion API routes.
// It sets up endpoints for receiving metrics, logs, and traces from agents
// with appropriate middleware for authentication, authorization, and logging.
//
// Protected routes:
//   - POST /telemetry/metrics - Ingest metrics data (requires gosight:api:telemetry:metrics permission)
//   - POST /telemetry/logs - Ingest log data (requires gosight:api:telemetry:logs permission)
//   - POST /telemetry/traces - Ingest trace data (requires gosight:api:telemetry:traces permission)
func SetupTelemetryRoutes(router *mux.Router, telemetryHandler *handlers.TelemetryHandler, withAccessLog func(http.Handler) http.Handler) {
	// Configure middleware
	withAuth := gosightauth.AuthMiddleware(telemetryHandler.Sys.Stores.Users)

	// Helper function to create secure handler with permission check
	secure := func(permission string, handler http.Handler) http.Handler {
		return withAccessLog(withAuth(gosightauth.RequirePermission(permission, handler, telemetryHandler.Sys.Stores.Users)))
	}

	// Telemetry data ingestion endpoints
	router.Handle("/telemetry/metrics",
		secure("gosight:api:telemetry:metrics", http.HandlerFunc(telemetryHandler.HandleMetrics))).
		Methods("POST")

	router.Handle("/telemetry/logs",
		secure("gosight:api:telemetry:logs", http.HandlerFunc(telemetryHandler.HandleLogs))).
		Methods("POST")

	router.Handle("/telemetry/traces",
		secure("gosight:api:telemetry:traces", http.HandlerFunc(telemetryHandler.HandleTraces))).
		Methods("POST")
}
