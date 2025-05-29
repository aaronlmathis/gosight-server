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

package routes

import (
	"net/http"

	"github.com/aaronlmathis/gosight-server/internal/api/handlers"
	gosightauth "github.com/aaronlmathis/gosight-server/internal/auth"
	"github.com/gorilla/mux"
)

// SetupLabelsRoutes configures label-related API routes
func SetupLabelsRoutes(r *mux.Router, labelsHandler *handlers.LabelsHandler, withAccessLog func(http.Handler) http.Handler) {
	// Configure middleware
	withAuth := gosightauth.AuthMiddleware(labelsHandler.Sys.Stores.Users)

	// Helper function to create secure handler with permission check
	secure := func(permission string, handler http.Handler) http.Handler {
		return withAccessLog(withAuth(gosightauth.RequirePermission(permission, handler, labelsHandler.Sys.Stores.Users)))
	}

	// Label value routes
	r.Handle("/labels/values",
		secure("gosight:api:labels:view", http.HandlerFunc(labelsHandler.HandleLabelValues))).
		Methods("GET")
}
