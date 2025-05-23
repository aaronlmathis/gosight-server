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

// Package sys provides the system context for the GoSight application.
// It contains the SystemContext struct which holds references to various subsystems
// and modules, allowing for easy access and management of the application's state.
// The SystemContext is passed to all subsystems, providing full access to config, state, and interfaces.
package sys

import (
	"context"

	gosightauth "github.com/aaronlmathis/gosight-server/internal/auth"
	"github.com/aaronlmathis/gosight-server/internal/cache"
	"github.com/aaronlmathis/gosight-server/internal/config"
	"github.com/aaronlmathis/gosight-server/internal/syncmanager"
	"github.com/aaronlmathis/gosight-server/internal/tracker"
	"github.com/aaronlmathis/gosight-server/internal/websocket"
)

// SystemContext is passed to all subsystems, providing full access to config, state, and interfaces.
type SystemContext struct {
	Ctx     context.Context
	Cfg     *config.Config
	Tracker *tracker.EndpointTracker // Tracks endpoint state, uptime, heartbeat
	WSHub   *websocket.HubManager
	Auth    map[string]gosightauth.AuthProvider
	Stores  *StoreModule
	Tele    *TelemetryModule
	Cache   *cache.Cache
	Buffers *BufferModule
	SyncMgr *syncmanager.SyncManager
}

// NewSystemContext creates a new SystemContext with the provided parameters.
// It initializes the context, configuration, tracker, websocket hub, authentication providers,
// stores, telemetry, caches, buffers, and synchronization manager.
// This function is typically called during the initialization phase of the application.
func NewSystemContext(
	ctx context.Context,
	cfg *config.Config,
	tracker *tracker.EndpointTracker,
	wsHub *websocket.HubManager,
	authProviders map[string]gosightauth.AuthProvider,
	stores *StoreModule,
	telemetry *TelemetryModule,
	caches *cache.Cache,
	buffers *BufferModule,
	syncMgr *syncmanager.SyncManager,

) *SystemContext {
	return &SystemContext{
		Ctx:     ctx,
		Cfg:     cfg,
		Tracker: tracker,
		WSHub:   wsHub,
		Auth:    authProviders,
		Stores:  stores,
		Tele:    telemetry,
		Cache:   caches,
		Buffers: buffers,
		SyncMgr: syncMgr,
	}
}
