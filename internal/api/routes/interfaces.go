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
// This file defines interfaces to avoid circular dependencies.
package routes

import (
	"net/http"

	"github.com/aaronlmathis/gosight-server/internal/sys"
)

// ServerInterface defines the interface that route handlers need from the HTTP server.
// This interface helps avoid circular dependencies between the routes package and the httpserver package.
type ServerInterface interface {
	// WithAccessLog provides access to the access logging middleware
	WithAccessLog(h http.Handler) http.Handler

	// Sys provides access to the system context including stores, config, and telemetry
	GetSys() *sys.SystemContext

	// Handler methods for alerts and events
	HandleAlertsAPI(w http.ResponseWriter, r *http.Request)
	HandleActiveAlertsAPI(w http.ResponseWriter, r *http.Request)
	HandleAlertRulesAPI(w http.ResponseWriter, r *http.Request)
	HandleCreateAlertRuleAPI(w http.ResponseWriter, r *http.Request)
	HandleAlertsSummaryAPI(w http.ResponseWriter, r *http.Request)
	HandleAlertContext(w http.ResponseWriter, r *http.Request)

	// Handler methods for user management
	HandleAPIUsers(w http.ResponseWriter, r *http.Request)
	HandleAPIUserCreate(w http.ResponseWriter, r *http.Request)
	HandleAPIUser(w http.ResponseWriter, r *http.Request)
	HandleAPIUserUpdate(w http.ResponseWriter, r *http.Request)
	HandleAPIUserDelete(w http.ResponseWriter, r *http.Request)
	HandleAPIUserPasswordChange(w http.ResponseWriter, r *http.Request)
	HandleAPIUserSettings(w http.ResponseWriter, r *http.Request)
	HandleAPIUserSettingsUpdate(w http.ResponseWriter, r *http.Request)
	HandleGetUserPreferences(w http.ResponseWriter, r *http.Request)
	HandleUpdateUserPreferences(w http.ResponseWriter, r *http.Request)

	// Handler methods for endpoint management
	HandleAPIEndpoints(w http.ResponseWriter, r *http.Request)
	HandleAPIEndpointCreate(w http.ResponseWriter, r *http.Request)
	HandleAPIEndpoint(w http.ResponseWriter, r *http.Request)
	HandleAPIEndpointUpdate(w http.ResponseWriter, r *http.Request)
	HandleAPIEndpointDelete(w http.ResponseWriter, r *http.Request)
	HandleAPIEndpointTest(w http.ResponseWriter, r *http.Request)
	HandleAPIEndpointStatus(w http.ResponseWriter, r *http.Request)
	HandleAPIEndpointEnable(w http.ResponseWriter, r *http.Request)
	HandleAPIEndpointDisable(w http.ResponseWriter, r *http.Request)

	// Additional handler methods for other route files can be added here as needed
}
