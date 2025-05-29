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

import (
	"context"

	"github.com/aaronlmathis/gosight-server/internal/contextutil"
	"github.com/aaronlmathis/gosight-server/internal/usermodel"
	"github.com/aaronlmathis/gosight-shared/utils"
)

// InjectSessionContext enriches a context with authenticated user information.
// This function takes user data and injects the user ID, roles, and permissions
// into the context for use throughout the request lifecycle. This enables
// authorization checks and audit logging without repeatedly querying the database.
//
// The function:
// 1. Sets the user ID in the context
// 2. Extracts and sets role names
// 3. Flattens permissions from all roles and removes duplicates
// 4. Logs the injected information for debugging
//
// Parameters:
//   - ctx: Base context to enrich with user information
//   - user: Authenticated user with roles and permissions loaded
//
// Returns:
//   - context.Context: Enhanced context containing user ID, roles, and permissions
func InjectSessionContext(ctx context.Context, user *usermodel.User) context.Context {
	ctx = contextutil.SetUserID(ctx, user.ID)

	roleNames := ExtractRoleNames((user.Roles))
	ctx = contextutil.SetUserRoles(ctx, roleNames)

	var permNames []string
	seen := map[string]struct{}{}
	for _, role := range user.Roles {
		for _, perm := range role.Permissions {
			if _, exists := seen[perm.Name]; !exists {
				permNames = append(permNames, perm.Name)
				seen[perm.Name] = struct{}{}
			}
		}
	}
	ctx = contextutil.SetUserPermissions(ctx, permNames)
	utils.Debug("Injected user: %s", user.ID)
	utils.Debug("Roles: %v", roleNames)
	utils.Debug("Permissions: %v", permNames)
	return ctx
}
