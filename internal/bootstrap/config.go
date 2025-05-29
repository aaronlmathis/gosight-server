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
	"log"
	"os"
	"path/filepath"

	"github.com/aaronlmathis/gosight-server/internal/config"
)

// LoadServerConfig loads the server configuration from a file, environment variables, and command line flags.
// It applies the following order of precedence:
// 1. Command line flags
// 2. Environment variables
// 3. Default configuration file
// 4. Hardcoded defaults
// The function returns a pointer to the loaded configuration.
func LoadServerConfig(configFlag *string) *config.Config {

	// Resolve config path
	configPath := resolvePath(*configFlag, "GOSIGHT_SERVER_CONFIG", "./config/config.yaml")

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	config.ApplyEnvOverrides(cfg)

	return cfg
}

// resolvePath determines the configuration file path using a priority order.
// This function implements the configuration resolution strategy by checking
// multiple sources in order of precedence: command line flags have highest
// priority, followed by environment variables, then the default fallback path.
//
// The resolution order ensures that:
// 1. Explicit command line arguments override all other sources
// 2. Environment variables provide deployment-specific configuration
// 3. Default paths enable zero-configuration startup for development
//
// All paths are converted to absolute paths to ensure consistent file access
// regardless of the current working directory.
//
// Parameters:
//   - flagVal: Command line flag value (highest priority)
//   - envVar: Environment variable name to check
//   - fallback: Default path to use if no other source provides a value
//
// Returns:
//   - string: Resolved absolute path to the configuration file
func resolvePath(flagVal, envVar, fallback string) string {
	if flagVal != "" {
		return absPath(flagVal)
	}
	if val := os.Getenv(envVar); val != "" {
		return absPath(val)
	}
	return absPath(fallback)
}

// absPath converts a relative or absolute path to an absolute path.
// This function ensures that all configuration file paths are resolved to
// absolute paths, preventing issues with relative path resolution when the
// application's working directory changes during execution.
//
// The function uses filepath.Abs() to handle platform-specific path resolution
// and will terminate the application if path resolution fails, as this indicates
// a fundamental filesystem issue that prevents proper configuration loading.
//
// Parameters:
//   - path: File path that may be relative or absolute
//
// Returns:
//   - string: Absolute path equivalent of the input path
//
// Panics:
//   - If filepath.Abs() fails, indicating a serious filesystem issue
func absPath(path string) string {
	abs, err := filepath.Abs(path)
	if err != nil {
		log.Fatalf("Failed to resolve path: %v", err)
	}
	return abs
}
