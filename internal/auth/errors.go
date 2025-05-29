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

import "errors"

// Authentication and authorization error definitions.
// These errors provide standardized error messages for common authentication
// and authorization failures throughout the GoSight application.
var (
	// ErrInvalidPassword is returned when a provided password does not match
	// the stored password hash during authentication attempts.
	ErrInvalidPassword = errors.New("invalid password")

	// ErrInvalidTOTP is returned when a Time-based One-Time Password (TOTP)
	// code provided for multi-factor authentication is incorrect or expired.
	ErrInvalidTOTP = errors.New("invalid TOTP code")

	// ErrUserNotFound is returned when attempting to authenticate or authorize
	// a user that does not exist in the system.
	ErrUserNotFound = errors.New("user not found")

	// ErrUnauthorized is returned when a user lacks the necessary permissions
	// to access a resource or perform an action.
	ErrUnauthorized = errors.New("unauthorized")
)
