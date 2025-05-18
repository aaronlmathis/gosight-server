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

// Package contextutil provides safe helpers for working with context values
package contextutil

import (
	"context"
)

// ctxKey private key type to avoid collisions
type ctxKey string

const (
	userIDKey     ctxKey = "user_id"
	roleKey       ctxKey = "user_roles"
	permissionKey ctxKey = "user_permissions"
	traceIDKey    ctxKey = "trace_id"
	forbiddenKey  ctxKey = "forbidden"
)

// SetUserID returns a new context with the user ID stored
func SetUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

// GetUserID retrieves the user ID from the context
func GetUserID(ctx context.Context) (string, bool) {
	val := ctx.Value(userIDKey)
	if id, ok := val.(string); ok {
		return id, true
	}
	return "", false
}

// SetUserRoles stores a slice of role names in the context
func SetUserRoles(ctx context.Context, roles []string) context.Context {
	return context.WithValue(ctx, roleKey, roles)
}

// GetUserRoles retrieves the roles slice from context
func GetUserRoles(ctx context.Context) ([]string, bool) {
	val := ctx.Value(roleKey)
	if roles, ok := val.([]string); ok {
		return roles, true
	}
	return nil, false
}

// SetUserPermissions stores a slice of permission strings in the context
func SetUserPermissions(ctx context.Context, perms []string) context.Context {
	return context.WithValue(ctx, permissionKey, perms)
}

// GetUserPermissions retrieves the permissions slice from context
func GetUserPermissions(ctx context.Context) ([]string, bool) {
	val := ctx.Value(permissionKey)
	if perms, ok := val.([]string); ok {
		return perms, true
	}
	return nil, false
}

// SetTraceID stores a trace ID for request correlation
func SetTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceIDKey, traceID)
}

// GetTraceID retrieves the trace ID if set
func GetTraceID(ctx context.Context) (string, bool) {
	val := ctx.Value(traceIDKey)
	if id, ok := val.(string); ok {
		return id, true
	}
	return "", false
}

// SetUserScopes stores a map of user scopes in the context for the user
func SetUserScopes(ctx context.Context, scopes map[string][]string) context.Context {
	return context.WithValue(ctx, ctxKey("user_scopes"), scopes)
}

// GetUserScopes retrieves the user scopes from the context
func GetUserScopes(ctx context.Context) (map[string][]string, bool) {
	val := ctx.Value(ctxKey("user_scopes"))
	if scopes, ok := val.(map[string][]string); ok {
		return scopes, true
	}
	return nil, false
}

// SetForbidden sets a flag in the context to indicate forbidden access
func SetForbidden(ctx context.Context) context.Context {
	return context.WithValue(ctx, forbiddenKey, true)
}

// IsForbidden checks if the context has the forbidden flag set
func IsForbidden(ctx context.Context) bool {
	v, ok := ctx.Value(forbiddenKey).(bool)
	return ok && v
}
