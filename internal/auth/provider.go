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

// Package gosightauth provides authentication and authorization functionality for GoSight.
// This package implements various authentication providers including local authentication,
// OAuth providers (Google, GitHub, Azure, AWS), multi-factor authentication (MFA),
// WebAuthn support, and session management.
package gosightauth

import (
	"net/http"

	"github.com/aaronlmathis/gosight-server/internal/usermodel"
)

// AuthProvider defines the interface that all authentication providers must implement.
// This interface provides a consistent way to handle different authentication methods
// including local username/password authentication and various OAuth providers.
//
// Implementations include:
//   - Local authentication (username/password)
//   - Google OAuth
//   - GitHub OAuth
//   - Azure OAuth
//   - AWS OAuth
//
// Each provider handles the specific OAuth flow or authentication method while
// providing a uniform interface for the application.
type AuthProvider interface {
	// StartLogin initiates the authentication process for the provider.
	// For OAuth providers, this typically redirects to the provider's authorization endpoint.
	// For local authentication, this may validate credentials directly.
	StartLogin(w http.ResponseWriter, r *http.Request)

	// HandleCallback processes the authentication response from the provider.
	// For OAuth providers, this handles the callback with authorization code.
	// Returns the authenticated user or an error if authentication failed.
	HandleCallback(w http.ResponseWriter, r *http.Request) (*usermodel.User, error)
}
