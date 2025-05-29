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
	"errors"

	"github.com/aaronlmathis/gosight-server/internal/config"
	"github.com/aaronlmathis/gosight-server/internal/store/eventstore"
	"github.com/aaronlmathis/gosight-server/internal/store/eventstore/pgeventstore"
	"github.com/aaronlmathis/gosight-shared/utils"
)

// InitEventStore initializes the event store component for the GoSight server.
// The event store provides persistent storage and retrieval of system events,
// audit logs, and operational activities. Events are critical for monitoring,
// debugging, and maintaining system observability.
//
// Supported storage engines:
//   - json: File-based JSON storage for development and small deployments
//   - postgres: PostgreSQL database backend for production environments
//
// The JSON backend stores events in local files and is suitable for testing
// or single-node deployments. The PostgreSQL backend provides ACID compliance,
// concurrent access, and better performance for production workloads.
//
// Parameters:
//   - cfg: Configuration containing event store settings including engine type and connection details
//
// Returns:
//   - eventstore.EventStore: Initialized event store interface implementation
//   - error: If storage initialization fails or unsupported engine is specified
func InitEventStore(cfg *config.Config) (eventstore.EventStore, error) {
	utils.Info("Initializing user store type: %s", cfg.EventStore.Engine)

	switch cfg.EventStore.Engine {
	case "json":
		store, err := eventstore.NewJSONEventStore(cfg.EventStore.Path)
		if err != nil {
			return nil, err
		}
		return store, nil

	case "postgres":

		db, err := sql.Open("postgres", cfg.EventStore.DSN) // TODO 	Fix this to be more generic
		if err != nil {
			return nil, err
		}

		// Optionally test the connection
		if err := db.Ping(); err != nil {
			return nil, err
		}
		return pgeventstore.NewPGEventStore(db), nil

	default:
		return nil, errors.New("invalid event backend")
	}
}
