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
	"errors"

	"github.com/aaronlmathis/gosight-server/internal/config"
	"github.com/aaronlmathis/gosight-server/internal/store/rulestore"
)

// InitRuleStore initializes the rule store component for the GoSight server.
// The rule store manages alerting rules, evaluation criteria, and monitoring
// configurations. It provides persistent storage for rules that define when
// alerts should be triggered based on metric thresholds, log patterns, or
// system conditions.
//
// Rule types supported:
//   - Metric-based alerting rules with thresholds and conditions
//   - Log-based rules for pattern matching and anomaly detection
//   - Composite rules combining multiple data sources
//   - Scheduled evaluation rules with time-based triggers
//   - Resource-specific rules for targeted monitoring
//
// Currently supported storage engines:
//   - yaml: File-based YAML storage for human-readable rule definitions
//   - memory: In-memory storage for testing and development (commented out)
//
// The YAML engine allows for easy rule management, version control, and
// collaborative rule development through standard file-based workflows.
//
// Parameters:
//   - cfg: Configuration containing rule store settings including engine type and file path
//
// Returns:
//   - rulestore.RuleStore: Initialized rule store interface implementation
//   - error: If rule store initialization fails or unsupported engine is specified
func InitRuleStore(cfg *config.Config) (rulestore.RuleStore, error) {
	switch cfg.RuleStore.Engine {
	case "yaml":
		return rulestore.NewYAMLStore(cfg.RuleStore.Path)
	//case "memory":
	//return rulestore.NewMemoryStore(), nil
	default:
		return nil, errors.New("unsupported rule store engine: " + cfg.RuleStore.Engine)
	}

}
