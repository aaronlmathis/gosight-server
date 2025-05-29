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

package config

import (
	"os"
	"path/filepath"
)

// defaultServerYAML contains the minimal default configuration for GoSight server.
// This configuration provides basic server settings that allow the server to start
// with minimal setup, using safe defaults for development and testing environments.
//
// Default configuration includes:
//   - HTTP server listening on port 8080
//   - SQLite storage engine for simplicity
//   - Local database file in ./data directory
//
// This configuration is automatically created when no configuration file exists,
// enabling zero-configuration startup for development and evaluation purposes.
// Production deployments should use a comprehensive configuration file with
// appropriate security, performance, and reliability settings.
const defaultServerYAML = `listen: ":8080"
storage: "sqlite"
database_path: "./data/gobright.db"
`

// EnsureDefaultConfig checks if the default configuration file exists at the specified path
// and creates it with minimal default settings if it doesn't exist.
//
// This function provides automatic configuration file generation for first-time
// installations and development environments, reducing setup complexity while
// maintaining security through explicit file creation.
//
// Parameters:
//   - path: Target path for the configuration file (absolute or relative)
//
// Returns:
//   - error: File creation or directory creation errors
//
// Behavior:
//   - Returns nil immediately if configuration file already exists
//   - Creates parent directories if they don't exist (with 0755 permissions)
//   - Writes default configuration with restrictive permissions (0644)
//   - Preserves existing configuration files without modification
//
// Error conditions:
//   - Permission denied when creating directories or files
//   - Disk space insufficient for file creation
//   - Invalid path or filesystem errors
//
// Security considerations:
//   - Creates files with 0644 permissions (readable by owner and group)
//   - Creates directories with 0755 permissions (standard directory permissions)
//   - Does not overwrite existing configuration files
//
// Example usage:
//
//	if err := EnsureDefaultConfig("/etc/gosight/server.yaml"); err != nil {
//		log.Fatalf("Failed to create default config: %v", err)
//	}
//
// The generated configuration file should be reviewed and customized
// for production use, particularly security and storage settings.
func EnsureDefaultConfig(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return err
		}
		return os.WriteFile(path, []byte(defaultServerYAML), 0644)
	}
	return nil
}
