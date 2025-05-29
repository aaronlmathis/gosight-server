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

package handlers

import (
	"net/http"

	"github.com/aaronlmathis/gosight-server/internal/sys"
)

// WebSocketsHandler handles WebSocket connection endpoints
type WebSocketsHandler struct {
	Sys *sys.SystemContext
}

// NewWebSocketsHandler creates a new WebSocketsHandler
func NewWebSocketsHandler(sys *sys.SystemContext) *WebSocketsHandler {
	return &WebSocketsHandler{
		Sys: sys,
	}
}

// HandleMetricsWS handles WebSocket connections for metrics streaming
func (h *WebSocketsHandler) HandleMetricsWS(w http.ResponseWriter, r *http.Request) {
	h.Sys.WSHub.Metrics.ServeWS(w, r)
}

// HandleAlertsWS handles WebSocket connections for alerts streaming
func (h *WebSocketsHandler) HandleAlertsWS(w http.ResponseWriter, r *http.Request) {
	h.Sys.WSHub.Alerts.ServeWS(w, r)
}

// HandleEventsWS handles WebSocket connections for events streaming
func (h *WebSocketsHandler) HandleEventsWS(w http.ResponseWriter, r *http.Request) {
	h.Sys.WSHub.Events.ServeWS(w, r)
}

// HandleLogsWS handles WebSocket connections for logs streaming
func (h *WebSocketsHandler) HandleLogsWS(w http.ResponseWriter, r *http.Request) {
	h.Sys.WSHub.Logs.ServeWS(w, r)
}

// HandleCommandWS handles WebSocket connections for command execution
func (h *WebSocketsHandler) HandleCommandWS(w http.ResponseWriter, r *http.Request) {
	h.Sys.WSHub.Commands.ServeWS(w, r)
}

// HandleProcessWS handles WebSocket connections for process monitoring
func (h *WebSocketsHandler) HandleProcessWS(w http.ResponseWriter, r *http.Request) {
	h.Sys.WSHub.Processes.ServeWS(w, r)
}
