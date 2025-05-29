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
	"net/http"

	"github.com/aaronlmathis/gosight-server/internal/store/userstore"
	"github.com/aaronlmathis/gosight-server/internal/usermodel"
	"github.com/aaronlmathis/gosight-shared/utils"
)

// LocalAuth implements the AuthProvider interface for username/password authentication.
// This provider handles traditional credential-based authentication where users
// provide a username and password that are validated against stored password hashes.
type LocalAuth struct {
	// Store provides access to user data for authentication verification
	Store userstore.UserStore
}

// StartLogin initiates local authentication flow.
// Unlike OAuth providers, local authentication doesn't require redirects.
// The UI handles rendering the login form directly.
//
// Parameters:
//   - w: HTTP response writer
//   - r: HTTP request (unused for local auth)
func (l *LocalAuth) StartLogin(w http.ResponseWriter, r *http.Request) {
	// Local login doesn't require redirect — handled via UI
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Local login requested. UI should render login form."))
}

// HandleCallback processes local authentication credentials.
// Validates the provided username and password against stored user data.
// This method extracts credentials from form data, looks up the user,
// and verifies the password hash.
//
// Parameters:
//   - w: HTTP response writer (unused for local auth)
//   - r: HTTP request containing form data with username and password
//
// Returns:
//   - *usermodel.User: Authenticated user if credentials are valid
//   - error: ErrUserNotFound if user doesn't exist, ErrInvalidPassword if password is wrong
func (l *LocalAuth) HandleCallback(w http.ResponseWriter, r *http.Request) (*usermodel.User, error) {
	ctx := r.Context()
	username := r.FormValue("username")
	password := r.FormValue("password")

	user, err := l.Store.GetUserByUsername(ctx, username)
	if err != nil {
		utils.Debug("User not found: %s", username)
		return nil, err
	}
	if !CheckPasswordHash(password, user.PasswordHash) {
		utils.Debug("Password Hash ain't right: ")
		return nil, ErrInvalidPassword
	}

	return user, nil
}
