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

package bootstrap

import (
	"context"

	"github.com/aaronlmathis/gosight-server/internal/store/metastore"
	"github.com/aaronlmathis/gosight-server/internal/websocket"
	"github.com/aaronlmathis/gosight-shared/utils"
)

// InitWebSocketHub initializes the WebSocket hub manager for the GoSight server.
// The WebSocket hub provides real-time communication between the server and web clients,
// enabling live updates for monitoring dashboards, alerts, logs, and system events.
//
// The hub manager coordinates multiple specialized WebSocket hubs:
//   - Alert Hub: Real-time alert notifications and status updates
//   - Log Hub: Live log streaming and search results
//   - Event Hub: System events and audit log streaming
//   - Metric Hub: Real-time metric data and dashboard updates
//
// Key features:
//   - Multi-hub architecture for organized data streaming
//   - Client connection management and authentication
//   - Message broadcasting and targeted delivery
//   - Connection lifecycle management
//   - Integration with metadata tracking for resource-aware streaming
//
// The function creates the hub manager, starts all individual hubs in separate
// goroutines, and returns the manager for integration with HTTP handlers.
//
// Parameters:
//   - ctx: Context for hub lifecycle management and graceful shutdown
//   - metaStore: Metadata tracker for resource-aware message routing
//
// Returns:
//   - *websocket.HubManager: Initialized WebSocket hub manager with all hubs running
func InitWebSocketHub(ctx context.Context, metaStore *metastore.MetaTracker) *websocket.HubManager {

	// Create a new WebSocket hub manager
	hubManager := websocket.NewHubManager(metaStore)

	// Start all websocket hubs (alerts, logs, events, metrics) in separate goroutines
	hubManager.StartAll(ctx)
	utils.Info("WebSocket hub initialized and running")

	return hubManager

}
