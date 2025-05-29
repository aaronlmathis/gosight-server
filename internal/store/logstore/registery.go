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

package logstore

import (
	"context"
	"fmt"

	"github.com/aaronlmathis/gosight-server/internal/cache"
	"github.com/aaronlmathis/gosight-server/internal/config"
	"github.com/aaronlmathis/gosight-server/internal/store/logstore/filestore"
	"github.com/aaronlmathis/gosight-server/internal/store/logstore/victorialogs"
	victorialogstore "github.com/aaronlmathis/gosight-server/internal/store/logstore/victoriametrics"

	"github.com/aaronlmathis/gosight-shared/utils"
)

// InitLogStore initializes and returns a LogStore implementation based on the
// configured engine type. It acts as a factory function for creating different
// types of log storage backends.
//
// Supported Engines:
//   - "file": JSON file-based storage for development and small deployments
//   - "victoriametrics": VictoriaMetrics-based storage (logs stored as metrics)
//   - "victorialogs": VictoriaLogs-based storage (native log format)
//
// The function automatically selects the appropriate implementation based on
// the engine configuration and initializes it with the provided parameters.
//
// Parameters:
//   - ctx: Context for initialization operations
//   - cfg: Configuration containing logstore settings
//   - logCache: Optional cache implementation for improved performance
//
// Returns:
//   - LogStore: Initialized log store implementation
//   - error: Any initialization error
//
// Configuration Examples:
//
//	# File-based storage
//	logstore:
//	  engine: "file"
//	  dir: "/var/lib/gosight/logs"
//
//	# VictoriaMetrics storage
//	logstore:
//	  engine: "victoriametrics"
//	  url: "http://localhost:8428"
//	  dir: "/var/lib/gosight/logs"
//
//	# VictoriaLogs storage
//	logstore:
//	  engine: "victorialogs"
//	  url: "http://localhost:9428"
//
// Engine Selection Guide:
//   - Use "file" for development, testing, or small deployments
//   - Use "victoriametrics" when you already have VictoriaMetrics and want unified storage
//   - Use "victorialogs" for optimal log storage performance and native log querying
func InitLogStore(ctx context.Context, cfg *config.Config, logCache cache.LogCache) (LogStore, error) {
	engine := cfg.LogStore.Engine
	if engine == "" {
		engine = "file"
	}
	utils.Debug("InitLogStore selected engine: %s", engine)

	switch engine {
	case "file":
		utils.Debug("Bootstrapping JSON File LogStore.")
		s := filestore.New(cfg.LogStore.Dir)
		utils.Debug("Returning JSON Filestore at: %p", s)
		return s, nil

	case "victoriametrics":
		utils.Debug("Bootstrapping VictoriaMetrics LogStore.")
		s, err := victorialogstore.NewVictoriaLogStore(cfg.LogStore.Url, cfg.LogStore.Dir, logCache)
		if err != nil {
			return nil, fmt.Errorf("failed to create VictoriaMetrics LogStore: %w", err)
		}
		utils.Debug("Returning VictoriaMetric LogStore at: %p", s)
		return s, nil

	case "victorialogs":
		utils.Debug("Bootstrapping VictoriaLogs LogStore.")
		s, err := victorialogs.NewVictoriaLogsStore(cfg.LogStore.Url, logCache)
		if err != nil {
			return nil, fmt.Errorf("failed to create VictoriaLogs LogStore: %w", err)
		}
		utils.Debug("Returning VictoriaLogs LogStore at: %p", s)
		return s, nil

	default:
		return nil, fmt.Errorf("unsupported storage engine: %s", engine)
	}
}
