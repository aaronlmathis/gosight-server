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
	"github.com/aaronlmathis/gosight-server/internal/config"
	"github.com/aaronlmathis/gosight-server/internal/store/routestore"
)

// InitRouteStore initializes the route store component for the GoSight server.
// The route store manages routing configurations for data flow, load balancing,
// and service mesh integration. It stores and retrieves routing rules that
// determine how requests and data are directed throughout the system.
//
// Route store capabilities:
//   - Service routing and load balancing rules
//   - Data pipeline routing configurations
//   - Traffic routing and failover rules
//   - Service mesh integration settings
//   - Dynamic routing rule updates
//
// The route store uses file-based storage for routing configurations,
// allowing for easy management and version control of routing rules.
//
// Parameters:
//   - cfg: Configuration containing route store settings including file path
//
// Returns:
//   - *routestore.RouteStore: Initialized route store for routing management
//   - error: If route store initialization or file access fails
func InitRouteStore(cfg *config.Config) (*routestore.RouteStore, error) {
	// Initialize the route store
	routeStore, err := routestore.NewRouteStore(cfg.RouteStore.Path)
	if err != nil {
		return nil, err
	}
	return routeStore, nil
}
