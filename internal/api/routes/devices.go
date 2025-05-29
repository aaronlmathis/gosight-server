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
// This file contains network device management related routes and handlers.
package routes

import (
	"net/http"

	"github.com/aaronlmathis/gosight-server/internal/api/handlers"
	gosightauth "github.com/aaronlmathis/gosight-server/internal/auth"
	"github.com/aaronlmathis/gosight-server/internal/sys"
	"github.com/gorilla/mux"
)

// SetupDevicesRoutes configures network device management API routes.
// It sets up endpoints for device discovery, device configuration, and device monitoring
// with appropriate middleware for authentication, authorization, and logging.
//
// Protected routes:
//   - GET /devices - List all devices (requires gosight:api:devices:view permission)
//   - POST /devices - Add new device (requires gosight:api:devices:create permission)
//   - GET /devices/{id} - Get device by ID (requires gosight:api:devices:view permission)
//   - PUT /devices/{id} - Update device (requires gosight:api:devices:update permission)
//   - DELETE /devices/{id} - Delete device (requires gosight:api:devices:delete permission)
//   - POST /devices/discover - Discover devices (requires gosight:api:devices:discover permission)
//   - GET /devices/{id}/status - Get device status (requires gosight:api:devices:view permission)
//   - POST /devices/{id}/poll - Poll device (requires gosight:api:devices:poll permission)
//   - GET /devices/{id}/interfaces - Get device interfaces (requires gosight:api:devices:view permission)
func SetupDevicesRoutes(router *mux.Router, sys *sys.SystemContext) {
	// Create handler
	devicesHandler := handlers.NewDevicesHandler(sys)

	// Configure middleware
	withAuth := gosightauth.AuthMiddleware(sys.Stores.Users)
	withLog := func(handler http.Handler) http.Handler {
		// TODO: Access logging middleware
		return handler
	}

	// Helper function to create secure handler with permission check
	secure := func(permission string, handler http.Handler) http.Handler {
		return withLog(withAuth(gosightauth.RequirePermission(permission, handler, sys.Stores.Users)))
	}

	// Device CRUD operations
	router.Handle("/devices",
		secure("gosight:api:devices:view", http.HandlerFunc(devicesHandler.HandleAPIDevices))).
		Methods("GET")

	router.Handle("/devices",
		secure("gosight:api:devices:create", http.HandlerFunc(devicesHandler.HandleAPIDeviceCreate))).
		Methods("POST")

	router.Handle("/devices/{id}",
		secure("gosight:api:devices:view", http.HandlerFunc(devicesHandler.HandleAPIDevice))).
		Methods("GET")

	router.Handle("/devices/{id}",
		secure("gosight:api:devices:update", http.HandlerFunc(devicesHandler.HandleAPIDeviceUpdate))).
		Methods("PUT")

	router.Handle("/devices/{id}",
		secure("gosight:api:devices:delete", http.HandlerFunc(devicesHandler.HandleAPIDeviceDelete))).
		Methods("DELETE")

	// Device discovery and management operations
	router.Handle("/devices/discover",
		secure("gosight:api:devices:discover", http.HandlerFunc(devicesHandler.HandleAPIDeviceDiscover))).
		Methods("POST")

	router.Handle("/devices/{id}/status",
		secure("gosight:api:devices:view", http.HandlerFunc(devicesHandler.HandleAPIDeviceStatus))).
		Methods("GET")

	router.Handle("/devices/{id}/poll",
		secure("gosight:api:devices:poll", http.HandlerFunc(devicesHandler.HandleAPIDevicePoll))).
		Methods("POST")

	router.Handle("/devices/{id}/interfaces",
		secure("gosight:api:devices:view", http.HandlerFunc(devicesHandler.HandleAPIDeviceInterfaces))).
		Methods("GET")
}
