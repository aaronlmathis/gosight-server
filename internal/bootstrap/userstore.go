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
	"github.com/aaronlmathis/gosight-server/internal/store/userstore"
	"github.com/aaronlmathis/gosight-server/internal/store/userstore/pgstore"
	"github.com/aaronlmathis/gosight-shared/utils"
	_ "github.com/lib/pq"
)

// InitUserStore initializes the user store component for the GoSight server.
// The user store provides persistent storage for user accounts, authentication data,
// roles, permissions, and access control information. It serves as the foundation
// for the authentication and authorization system.
//
// User data managed includes:
//   - User accounts and profiles
//   - Password hashes and authentication credentials
//   - User roles and role assignments
//   - Permissions and access control rules
//   - Multi-factor authentication settings
//   - Session and security information
//
// Currently supported storage engines:
//   - postgres: PostgreSQL database backend for production deployments
//
// The function establishes a secure database connection, validates connectivity,
// and returns a configured user store ready for authentication operations.
//
// Parameters:
//   - cfg: Configuration containing user store settings including engine type and DSN
//
// Returns:
//   - userstore.UserStore: Initialized user store interface implementation
//   - error: If database connection fails or unsupported engine is specified
func InitUserStore(cfg *config.Config) (userstore.UserStore, error) {
	userStoreEngine := cfg.UserStore.Engine

	utils.Info("Initializing user store type: %s", userStoreEngine)
	var userStore userstore.UserStore
	switch cfg.UserStore.Engine {
	case "postgres":
		db, err := sql.Open("postgres", cfg.UserStore.DSN)
		if err != nil {
			return nil, err
		}
		// Optionally test the connection
		if err := db.Ping(); err != nil {
			return nil, err
		}
		userStore = pgstore.New(db)

	default:
		return nil, fmt.Errorf("unsupported userstore type: %s", cfg.UserStore.Engine)
	}

	utils.Info("User store [%s] initialized successfully", userStoreEngine)
	return userStore, nil
}
