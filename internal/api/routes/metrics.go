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
// This file contains metrics and query related routes and handlers.
package routes

import (
	"net/http"

	"github.com/aaronlmathis/gosight-server/internal/api/handlers"
	gosightauth "github.com/aaronlmathis/gosight-server/internal/auth"
	"github.com/gorilla/mux"
)

// SetupMetricsRoutes configures metrics and query API routes.
// It sets up endpoints for metrics queries, metadata discovery, and data export
// with appropriate middleware for authentication, authorization, and logging.
//
// Protected routes:
//   - GET /query - Execute metrics query (requires gosight:api:metrics:query permission)
//   - GET /exportquery - Export query results (requires gosight:api:metrics:export permission)
//   - GET /metrics - Get metric namespaces (requires gosight:api:metrics:meta permission)
//   - GET /metrics/{namespace} - Get sub-namespaces (requires gosight:api:metrics:meta permission)
//   - GET /metrics/{namespace}/{sub} - Get metric names (requires gosight:api:metrics:meta permission)
//   - GET /metrics/{namespace}/{sub}/{metric}/dimensions - Get metric dimensions (requires gosight:api:metrics:meta permission)
//   - GET /metrics/{namespace}/{sub}/{metric}/labels - Get metric labels (requires gosight:api:metrics:meta permission)
//   - GET /metrics/{namespace}/{sub}/{metric}/data - Get metric data (requires gosight:api:metrics:read permission)
//   - GET /metrics/{namespace}/{sub}/{metric}/latest - Get latest metric value (requires gosight:api:metrics:read permission)
func SetupMetricsRoutes(router *mux.Router, metricsHandler *handlers.MetricsHandler, withAccessLog func(http.Handler) http.Handler) {
	// Configure middleware
	withAuth := gosightauth.AuthMiddleware(metricsHandler.Sys.Stores.Users)
	withLog := withAccessLog

	// Helper function to create secure handler with permission check
	secure := func(permission string, handler http.Handler) http.Handler {
		return withLog(withAuth(gosightauth.RequirePermission(permission, handler, metricsHandler.Sys.Stores.Users)))
	}

	// Metrics query endpoints
	router.Handle("/query",
		secure("gosight:api:metrics:query", http.HandlerFunc(metricsHandler.HandleAPIQuery))).
		Methods("GET")

	router.Handle("/exportquery",
		secure("gosight:api:metrics:export", http.HandlerFunc(metricsHandler.HandleExportQuery))).
		Methods("GET")

	// Metadata discovery endpoints
	router.Handle("/metrics",
		secure("gosight:api:metrics:meta", http.HandlerFunc(metricsHandler.GetNamespaces))).
		Methods("GET")

	router.Handle("/metrics/{namespace}",
		secure("gosight:api:metrics:meta", http.HandlerFunc(metricsHandler.GetSubNamespaces))).
		Methods("GET")

	router.Handle("/metrics/{namespace}/{sub}",
		secure("gosight:api:metrics:meta", http.HandlerFunc(metricsHandler.GetMetricNames))).
		Methods("GET")

	router.Handle("/metrics/{namespace}/{sub}/{metric}/dimensions",
		secure("gosight:api:metrics:meta", http.HandlerFunc(metricsHandler.GetMetricDimensions))).
		Methods("GET")

	router.Handle("/metrics/{namespace}/{sub}/{metric}/labels",
		secure("gosight:api:metrics:meta", http.HandlerFunc(metricsHandler.GetMetricDimensions))).
		Methods("GET")

	router.Handle("/metrics/{namespace}/{sub}/{metric}/data",
		secure("gosight:api:metrics:read", http.HandlerFunc(metricsHandler.GetMetricData))).
		Methods("GET")

	router.Handle("/metrics/{namespace}/{sub}/{metric}/latest",
		secure("gosight:api:metrics:read", http.HandlerFunc(metricsHandler.GetMetricLatest))).
		Methods("GET")
}
