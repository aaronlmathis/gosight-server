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

	"github.com/aaronlmathis/gosight-server/internal/config"
	"github.com/aaronlmathis/gosight-server/internal/store/datastore"
	pgstore "github.com/aaronlmathis/gosight-server/internal/store/datastore/pgsql"

	"github.com/aaronlmathis/gosight-shared/utils"
	_ "github.com/lib/pq"
)

// InitDataStore initializes the data store component for the GoSight server.
// The data store provides persistent storage for various data types including
// metrics, logs, and system data. It abstracts the underlying storage engine
// to allow different implementations while maintaining a consistent interface.
//
// Currently supported engines:
//   - postgres: PostgreSQL database backend for relational data storage
//
// The function establishes a database connection, tests connectivity, and
// returns a configured data store instance ready for use by the application.
//
// Parameters:
//   - cfg: Configuration containing data store settings including engine type and DSN
//
// Returns:
//   - datastore.DataStore: Initialized data store interface implementation
//   - error: If database connection fails or unsupported engine is specified
func InitDataStore(cfg *config.Config) (datastore.DataStore, error) {
	dataStoreType := cfg.DataStore.Engine

	utils.Info("Initializing user store type: %s", dataStoreType)
	var dataStore datastore.DataStore
	switch cfg.DataStore.Engine {
	case "postgres":
		db, err := sql.Open("postgres", cfg.DataStore.DSN) // TODO 	Fix this to be more generic
		if err != nil {
			return nil, err
		}

		if err := db.Ping(); err != nil {
			return nil, err
		}
		dataStore = pgstore.NewPGDataStore(db)

	default:
		return nil, fmt.Errorf("unsupported datastore type: %s", dataStoreType)
	}

	utils.Info("Data store [%s] initialized successfully", dataStoreType)
	return dataStore, nil
}
