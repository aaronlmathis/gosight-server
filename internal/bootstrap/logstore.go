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
	"fmt"

	"github.com/aaronlmathis/gosight-server/internal/cache"
	"github.com/aaronlmathis/gosight-server/internal/config"
	"github.com/aaronlmathis/gosight-server/internal/store/logstore"
	"github.com/aaronlmathis/gosight-shared/utils"
)

// InitLogStore initializes the log store component for the GoSight server.
// The log store provides high-performance storage and retrieval for application
// logs, system logs, and other structured log data. It serves as the primary
// backend for log aggregation, search, and analysis capabilities.
//
// The log store supports multiple storage engines optimized for different
// use cases and deployment scenarios:
//   - elasticsearch: Full-text search and analytics
//   - victorialogs: High-performance log storage with LogsQL
//   - file: Simple file-based storage for development
//   - postgres: Database storage for structured logs
//
// The function integrates with the log cache for improved query performance
// and delegates engine-specific initialization to the logstore package.
//
// Parameters:
//   - ctx: Context for initialization and cancellation
//   - cfg: Configuration containing log store settings and engine selection
//   - logCache: Cache component for improved log query performance
//
// Returns:
//   - logstore.LogStore: Initialized log store interface implementation
//   - error: If log store initialization fails for the specified engine
func InitLogStore(ctx context.Context, cfg *config.Config, logCache cache.LogCache) (logstore.LogStore, error) {
	engine := cfg.LogStore.Engine
	utils.Info("Initializing log store engine: %s", engine)

	s, err := logstore.InitLogStore(ctx, cfg, logCache)
	if err != nil {
		return nil, fmt.Errorf("failed to init log store: %w", err)
	}

	utils.Info("Log store [%s] initialized successfully", engine)
	return s, nil
}
