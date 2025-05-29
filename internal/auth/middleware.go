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
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/aaronlmathis/gosight-server/internal/contextutil"
	"github.com/aaronlmathis/gosight-server/internal/store/userstore"
	"github.com/aaronlmathis/gosight-server/internal/usermodel"
	"github.com/aaronlmathis/gosight-shared/utils"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// RequirePermission is a middleware that ensures the authenticated user has a specific permission.
// If the user lacks the required permission, it returns 403 Forbidden for API requests
// or redirects to /unauthorized for web requests. The middleware can refresh permissions
// from the database if they're missing from the context.
//
// Parameters:
//   - required: The permission name that the user must have
//   - next: The next HTTP handler to call if permission check passes
//   - userStore: User store for refreshing permissions from database
//
// Returns:
//   - http.Handler: Middleware handler that enforces the permission requirement
func RequirePermission(required string, next http.Handler, userStore userstore.UserStore) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Permissions missing from context?
		perms, ok := contextutil.GetUserPermissions(ctx)
		if !ok || len(perms) == 0 {
			// Try to fetch fresh perms from DB
			userID, ok := contextutil.GetUserID(ctx)
			if ok {
				user, err := userStore.GetUserWithPermissions(ctx, userID)
				if err == nil {
					perms = FlattenPermissions(user.Roles)
					ctx = contextutil.SetUserPermissions(ctx, perms)
					r = r.WithContext(ctx) // update request context
				}
			} else {
				// User is not authenticated
				// Set forbidden flag in context
				ctx = contextutil.SetForbidden(ctx)
				r = r.WithContext(ctx)

			}
		}

		if !HasPermission(ctx, required) {
			//utils.Debug("RequirePermission: missing %s", required)

			if isAPIRequest(r) {
				http.Error(w, "forbidden", http.StatusForbidden)
			} else {
				http.Redirect(w, r, "/unauthorized", http.StatusSeeOther)
			}
			return
		}

		next.ServeHTTP(w, r)
	})
}

// RequireAnyPermissionWithStore creates a middleware that requires any one of the specified permissions.
// This is useful for endpoints that can be accessed by users with different permission levels.
// Returns a middleware function that can be used with gorilla/mux router.
//
// Parameters:
//   - store: User store for refreshing permissions from database
//   - required: Variable number of permission names, user needs at least one
//
// Returns:
//   - func(http.Handler) http.Handler: Middleware factory function
func RequireAnyPermissionWithStore(store userstore.UserStore, required ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			perms, ok := contextutil.GetUserPermissions(ctx)

			if !ok || len(perms) == 0 {
				if userID, ok := contextutil.GetUserID(ctx); ok {
					user, err := store.GetUserWithPermissions(ctx, userID)
					if err == nil {
						perms = FlattenPermissions(user.Roles)
						ctx = contextutil.SetUserPermissions(ctx, perms)
						r = r.WithContext(ctx)
					}
				}
			}

			if !HasAnyPermission(ctx, required...) {
				//utils.Debug("RequireAnyPermission: missing one of %v", required)

				if isAPIRequest(r) {
					http.Error(w, "forbidden", http.StatusForbidden)
				} else {
					http.Redirect(w, r, "/unauthorized", http.StatusSeeOther)
				}
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// AuthMiddleware validates JWT tokens and injects user context into requests.
// This middleware:
// 1. Extracts session tokens from cookies or Authorization headers
// 2. Validates JWT tokens and extracts claims
// 3. Refreshes user roles/permissions if they're stale (>10 minutes old)
// 4. Injects user ID, roles, permissions, and trace ID into request context
// 5. Redirects unauthenticated users to login or returns 401 for API requests
//
// Parameters:
//   - userStore: User store for refreshing user data from database
//
// Returns:
//   - mux.MiddlewareFunc: Middleware function compatible with gorilla/mux
func AuthMiddleware(userStore userstore.UserStore) mux.MiddlewareFunc {
	//utils.Debug("AuthMiddleware: Injecting user context")
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, err := GetSessionToken(r)
			if err != nil {
				//utils.Debug("AuthMiddleWare: GetSessionToken failed: %v", err)
				if isAPIRequest(r) {
					http.Error(w, "unauthorized", http.StatusUnauthorized)
				} else {
					http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
				}
				return
			}
			//utils.Debug("AuthMiddleware: Token found: %s", token)
			claims, err := ValidateToken(token)
			if err != nil {
				//utils.Debug("AuthMiddleWare: ValidateToken failed: %v", err)
				if isAPIRequest(r) {
					http.Error(w, "unauthorized", http.StatusUnauthorized)
				} else {
					http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
				}
				return
			}
			//utils.Debug("AuthMiddleware: Token claims: %v", claims)
			userID := claims.UserID
			traceID := claims.TraceID

			var user *usermodel.User
			var roleNames []string
			var perms []string
			rolesTTL := 10 * time.Minute
			refreshedAt := time.Unix(claims.RolesRefreshedAt, 0)

			// Revalidate if roles are stale or missing
			if len(claims.Roles) == 0 || time.Since(refreshedAt) > rolesTTL {
				user, err = userStore.GetUserWithPermissions(r.Context(), userID)
				if err != nil {
					//utils.Debug("Failed to reload user roles: %v", err)
					http.Error(w, "unauthorized", http.StatusUnauthorized)
					return
				}
				roleNames = ExtractRoleNames(user.Roles)
				perms = FlattenPermissions(user.Roles)
				//utils.Debug("Flattened permissions from DB: %v", perms)
			} else {
				// Roles are fresh, use from token
				roleNames = claims.Roles

			}

			// Inject trace ID
			ctx := r.Context()
			if traceID != "" {
				ctx = contextutil.SetTraceID(ctx, traceID)
			} else if _, ok := contextutil.GetTraceID(ctx); !ok {
				ctx = contextutil.SetTraceID(ctx, uuid.New().String())
			}

			ctx = contextutil.SetUserID(ctx, userID)
			ctx = contextutil.SetUserRoles(ctx, roleNames)

			if user != nil && len(user.Roles) > 0 {
				perms = FlattenPermissions(user.Roles)
				ctx = contextutil.SetUserPermissions(ctx, perms)
				//utils.Debug("Revalidated user: %s, permissions: %v", user.ID, perms)
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// AccessLogMiddleware provides structured access logging for HTTP requests.
// This middleware captures comprehensive request information including:
// - Request timing and duration
// - HTTP method, path, and status code
// - User identification and permissions
// - Trace ID for request correlation
// - User agent and IP address
//
// The middleware ensures every request has a trace ID for debugging and monitoring.
// Log entries are formatted as JSON for easy parsing by log aggregation systems.
//
// Parameters:
//   - next: The next HTTP handler in the chain
//
// Returns:
//   - http.Handler: Middleware handler that logs request details
func AccessLogMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Use existing X-Trace-ID or generate one
		traceID := r.Header.Get("X-Trace-ID")
		if traceID == "" {
			traceID = uuid.NewString()
		}

		// Set trace ID into context and response header
		ctx := contextutil.SetTraceID(r.Context(), traceID)
		w.Header().Set("X-Trace-ID", traceID)

		// Wrap writer to capture status code
		rr := &statusRecorder{ResponseWriter: w, status: 200}

		// Call next handler
		next.ServeHTTP(rr, r.WithContext(ctx))

		duration := time.Since(start)

		// Structured fields
		userID, _ := contextutil.GetUserID(ctx)
		roles, _ := contextutil.GetUserRoles(ctx)
		perms, _ := contextutil.GetUserPermissions(ctx)

		if userID == "" {
			userID = "anonymous"
		}

		entry := map[string]interface{}{
			"timestamp":   time.Now().Format(time.RFC3339),
			"trace_id":    traceID,
			"method":      r.Method,
			"path":        r.URL.Path,
			"status":      rr.status,
			"duration_ms": duration.Milliseconds(),
			"user_id":     userID,
			"roles":       roles,
			"permissions": perms,
			"user_agent":  r.UserAgent(),
			"ip":          r.RemoteAddr,
		}

		logJSON, _ := json.Marshal(entry)
		utils.Access(string(logJSON)) // or send to file/syslog/etc
	})
}

// statusRecorder wraps http.ResponseWriter to capture the HTTP status code.
// This is used by AccessLogMiddleware to log response status codes since
// the standard ResponseWriter doesn't provide access to the status after writing.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

// WriteHeader captures the status code before delegating to the wrapped ResponseWriter.
// This allows the middleware to log the actual HTTP status code returned by handlers.
func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

// isAPIRequest determines if an HTTP request is targeting an API endpoint.
// Used by authentication middleware to decide between JSON error responses
// (for API requests) and HTML redirects (for web requests).
//
// Parameters:
//   - r: HTTP request to check
//
// Returns:
//   - bool: true if the request path starts with "/api/", false otherwise
func isAPIRequest(r *http.Request) bool {
	return strings.HasPrefix(r.URL.Path, "/api/")
}
