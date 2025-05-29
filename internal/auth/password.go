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

import "golang.org/x/crypto/bcrypt"

// HashPassword creates a bcrypt hash of the provided password.
// Uses bcrypt cost factor of 14 for enhanced security. This provides
// strong protection against brute force attacks while maintaining
// reasonable performance for authentication operations.
//
// Parameters:
//   - password: Plain text password to hash
//
// Returns:
//   - string: Base64-encoded bcrypt hash
//   - error: If hashing fails due to memory or other system constraints
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

// CheckPasswordHash verifies a plain text password against a bcrypt hash.
// This function is used during authentication to validate user credentials
// without storing or comparing plain text passwords.
//
// Parameters:
//   - password: Plain text password to verify
//   - hash: Stored bcrypt hash to compare against
//
// Returns:
//   - bool: true if password matches the hash, false otherwise
func CheckPasswordHash(password, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}
