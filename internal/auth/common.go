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

package gosightauth

import "github.com/aaronlmathis/gosight-server/internal/usermodel"

// FlattenPermissions extracts all unique permission names from a collection of roles.
// This function is used to create a flat list of permissions that a user has across
// all their assigned roles, removing duplicates. This is useful for authorization
// checks where you need to know if a user has a specific permission regardless of
// which role grants it.
//
// Parameters:
//   - roles: Slice of roles containing permissions to flatten
//
// Returns:
//   - []string: Unique permission names from all roles
func FlattenPermissions(roles []usermodel.Role) []string {
	perms := map[string]struct{}{}
	for _, role := range roles {
		for _, p := range role.Permissions {
			perms[p.Name] = struct{}{}
		}
	}
	var result []string
	for p := range perms {
		result = append(result, p)
	}
	return result
}

// ExtractRoleNames extracts the names from a collection of role objects.
// This is a utility function commonly used when you need just the role names
// for logging, token claims, or authorization checks without the full role data.
//
// Parameters:
//   - roles: Slice of role objects to extract names from
//
// Returns:
//   - []string: Role names in the same order as the input roles
func ExtractRoleNames(roles []usermodel.Role) []string {
	names := make([]string, 0, len(roles))
	for _, r := range roles {
		names = append(names, r.Name)
	}
	return names
}
