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
	"database/sql"
	"fmt"

	"github.com/aaronlmathis/gosight-server/internal/cache"
	"github.com/aaronlmathis/gosight-server/internal/config"
	"github.com/aaronlmathis/gosight-server/internal/store/resourcestore"
	"github.com/aaronlmathis/gosight-server/internal/store/resourcestore/pgstore"
	"github.com/aaronlmathis/gosight-server/internal/telemetry"
	"github.com/aaronlmathis/gosight-shared/utils"
)

// InitResourceStore initializes the resource store component for the GoSight server.
// The resource store provides persistent storage for system resources including
// agents, endpoints, services, containers, and their metadata. It serves as the
// central repository for resource information used throughout the monitoring system.
//
// Resources stored include:
//   - Agent information and registration data
//   - Service endpoints and health status
//   - Container and process metadata
//   - Resource tags and labels
//   - Resource relationships and dependencies
//
// Currently supported storage engines:
//   - postgres: PostgreSQL database backend for production deployments
//
// The function establishes a database connection, validates connectivity,
// and returns a configured resource store ready for use.
//
// Parameters:
//   - cfg: Configuration containing resource store settings including engine type and DSN
//
// Returns:
//   - resourcestore.ResourceStore: Initialized resource store interface implementation
//   - error: If database connection fails or unsupported engine is specified
func InitResourceStore(cfg *config.Config) (resourcestore.ResourceStore, error) {
	utils.Info("Initializing resource store")

	resourceStoreType := cfg.ResourceStore.Engine

	utils.Info("Initializing resource store type: %s", resourceStoreType)
	var resourceStore resourcestore.ResourceStore
	switch cfg.ResourceStore.Engine {
	case "postgres":
		db, err := sql.Open("postgres", cfg.ResourceStore.DSN) // TODO 	Fix this to be more generic
		if err != nil {
			return nil, err
		}

		if err := db.Ping(); err != nil {
			return nil, err
		}
		resourceStore = pgstore.NewPGResourceStore(db)

	default:
		return nil, fmt.Errorf("unsupported resource store type: %s", resourceStoreType)
	}

	utils.Info("Resource store [%s] initialized successfully", resourceStoreType)
	return resourceStore, nil
}

// InitResourceDiscovery initializes the resource discovery component for the GoSight server.
// Resource discovery automatically detects, catalogs, and monitors system resources
// across the infrastructure. It integrates with the resource cache for fast
// access and automatic persistence through the cache's background flushing.
//
// The resource discovery system:
//   - Automatically detects new agents and services
//   - Maintains an up-to-date inventory of system resources
//   - Tracks resource health and availability
//   - Provides resource topology and relationship mapping
//   - Enables dynamic monitoring configuration
//
// Parameters:
//   - resourceCache: Cache component for fast resource access and automatic persistence
//
// Returns:
//   - *telemetry.ResourceDiscovery: Initialized resource discovery service
//   - error: Currently always nil, reserved for future error conditions
func InitResourceDiscovery(resourceCache cache.ResourceCache) (*telemetry.ResourceDiscovery, error) {
	utils.Info("Initializing resource discovery")

	resourceDiscovery := telemetry.NewResourceDiscovery(resourceCache)
	return resourceDiscovery, nil
}
