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

	"github.com/aaronlmathis/gosight-server/internal/events"
	"github.com/aaronlmathis/gosight-server/internal/store/datastore"
	"github.com/aaronlmathis/gosight-server/internal/tracker"
)

// InitTracker initializes the unified endpoint tracker for the GoSight server.
// The endpoint tracker monitors and manages the lifecycle of all system endpoints
// including agents, containers, and services. It provides real-time tracking of
// endpoint status, health, and availability while emitting lifecycle events for
// system monitoring and alerting.
//
// Key responsibilities:
//   - Tracking agent registration and heartbeats
//   - Monitoring container lifecycle events
//   - Detecting endpoint failures and recoveries
//   - Emitting lifecycle events for monitoring
//   - Maintaining endpoint metadata and status
//   - Providing endpoint discovery and inventory
//
// The tracker integrates with the event emitter to publish endpoint changes
// and uses the data store for persistent tracking state.
//
// Parameters:
//   - ctx: Context for tracker operations and lifecycle management
//   - dataStore: Persistent storage for tracking state and endpoint data
//   - emitter: Event emitter for publishing endpoint lifecycle events
//
// Returns:
//   - *tracker.EndpointTracker: Initialized endpoint tracker ready for monitoring
func InitTracker(ctx context.Context, dataStore datastore.DataStore, emitter *events.Emitter) *tracker.EndpointTracker {
	t := tracker.NewEndpointTracker(ctx, emitter, dataStore)

	return t
}
