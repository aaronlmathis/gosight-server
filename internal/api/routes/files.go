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
// This file contains file upload and management related routes and handlers.
package routes

import (
	"net/http"

	"github.com/aaronlmathis/gosight-server/internal/api/handlers"
	gosightauth "github.com/aaronlmathis/gosight-server/internal/auth"
	"github.com/aaronlmathis/gosight-server/internal/sys"
	"github.com/gorilla/mux"
)

// SetupFileRoutes configures file upload and management API routes.
// It sets up endpoints for file uploads with appropriate middleware
// for authentication, authorization, and logging.
//
// Protected routes:
//   - POST /upload - Upload a file (requires gosight:api:files:upload permission)
//   - GET /files - List uploaded files (requires gosight:api:files:list permission)
//   - GET /files/{id} - Get file by ID (requires gosight:api:files:view permission)
//   - DELETE /files/{id} - Delete file (requires gosight:api:files:delete permission)
//   - GET /files/{id}/download - Download file (requires gosight:api:files:download permission)
func SetupFileRoutes(router *mux.Router, sys *sys.SystemContext, withAccessLog func(http.Handler) http.Handler) {
	// Create handler
	filesHandler := handlers.NewFilesHandler(sys)

	// Configure middleware
	withAuth := gosightauth.AuthMiddleware(sys.Stores.Users)


	// Helper function to create secure handler with permission check
	secure := func(permission string, handler http.Handler) http.Handler {
		return withAccessLog(withAuth(gosightauth.RequirePermission(permission, handler, sys.Stores.Users)))
	}

	// File upload endpoint
	router.Handle("/upload",
		secure("gosight:api:files:upload", http.HandlerFunc(filesHandler.HandleAPIFileUpload))).
		Methods("POST")

	// File management endpoints
	router.Handle("/files",
		secure("gosight:api:files:list", http.HandlerFunc(filesHandler.HandleAPIFiles))).
		Methods("GET")

	router.Handle("/files/{id}",
		secure("gosight:api:files:view", http.HandlerFunc(filesHandler.HandleAPIFile))).
		Methods("GET")

	router.Handle("/files/{id}",
		secure("gosight:api:files:delete", http.HandlerFunc(filesHandler.HandleAPIFileDelete))).
		Methods("DELETE")

	router.Handle("/files/{id}/download",
		secure("gosight:api:files:download", http.HandlerFunc(filesHandler.HandleAPIFileDownload))).
		Methods("GET")
}
