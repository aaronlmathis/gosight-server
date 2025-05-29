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
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// session.go: Handles JWT token generation, validation, and session management for GoSight.
// This file provides functionality for creating secure session tokens, validating them,
// managing session cookies, and extracting user information from authenticated requests.

// jwtSecret holds the JWT signing secret key used for token generation and validation.
// This should be initialized with a secure random key at application startup.
var jwtSecret []byte

// SessionClaims represents the JWT claims structure used for GoSight sessions.
// It includes user identification, roles, tracing information, and standard JWT claims.
type SessionClaims struct {
	// UserID is the unique identifier for the authenticated user
	UserID string `json:"sub"`
	// Roles contains the list of roles assigned to the user for authorization
	Roles []string `json:"roles,omitempty"`
	// TraceID is used for request tracing and debugging
	TraceID string `json:"trace_id,omitempty"`
	// RolesRefreshedAt tracks when user roles were last updated (Unix timestamp)
	RolesRefreshedAt int64 `json:"roles_refreshed_at"`
	// RegisteredClaims embeds standard JWT claims (exp, iat, nbf, etc.)
	jwt.RegisteredClaims
}

// InitJWTSecret initializes the JWT signing secret from a base64-encoded string.
// The secret must be base64-encoded and at least 32 bytes long for security.
// This function should be called once at application startup with a secure random key.
//
// Parameters:
//   - encoded: Base64-encoded secret key string
//
// Returns:
//   - error: If the secret is invalid, not base64-encoded, or less than 32 bytes
func InitJWTSecret(encoded string) error {
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil || len(decoded) < 32 {
		return fmt.Errorf("invalid JWT secret (must be base64 and 32+ bytes)")
	}
	jwtSecret = decoded
	return nil
}

// GenerateToken creates a new JWT token for an authenticated user.
// The token includes user identification, roles, and expiration information.
// Tokens are valid for 2 hours from generation time.
//
// Parameters:
//   - userID: Unique identifier for the user
//   - roles: List of roles assigned to the user for authorization
//   - traceID: Request trace ID for debugging and monitoring
//
// Returns:
//   - string: Signed JWT token
//   - error: If token generation fails
func GenerateToken(userID string, roles []string, traceID string) (string, error) {
	claims := SessionClaims{
		UserID:           userID,
		Roles:            roles,
		TraceID:          traceID,
		RolesRefreshedAt: time.Now().Unix(),
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(2 * time.Hour)),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// ValidateToken parses and validates a JWT token string.
// Verifies the token signature, expiration, and extracts the claims.
//
// Parameters:
//   - tokenStr: JWT token string to validate
//
// Returns:
//   - *SessionClaims: Parsed and validated token claims
//   - error: If token is invalid, expired, or malformed
func ValidateToken(tokenStr string) (*SessionClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &SessionClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*SessionClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}

// SetSessionCookie sets a secure HTTP-only session cookie with the provided JWT token.
// The cookie is configured with security best practices including HttpOnly, Secure,
// and SameSite attributes. Expires in 2 hours to match token expiration.
//
// Parameters:
//   - w: HTTP response writer to set the cookie on
//   - token: JWT token string to store in the cookie
func SetSessionCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "gosight_session",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(2 * time.Hour),
	})
}

// ErrNoSession is returned when no session token is found in the request.
// This can occur when the session cookie is missing or the Authorization header is not provided.
var ErrNoSession = errors.New("no session token found")

// GetSessionToken retrieves the session token from either a cookie or Authorization header.
// It first checks for a "gosight_session" cookie, then falls back to checking for a
// "Bearer" token in the Authorization header. This provides flexibility for both
// web browser clients (using cookies) and API clients (using Authorization headers).
//
// Parameters:
//   - r: HTTP request to extract the session token from
//
// Returns:
//   - string: Session token if found
//   - error: ErrNoSession if no token is found in either location
func GetSessionToken(r *http.Request) (string, error) {
	if cookie, err := r.Cookie("gosight_session"); err == nil && cookie.Value != "" {
		return cookie.Value, nil
	}

	authHeader := r.Header.Get("Authorization")
	if len(authHeader) > 7 && strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer "), nil
	}

	return "", ErrNoSession
}

// GetSessionClaims retrieves and validates the session claims from the request.
// This is a convenience function that combines token extraction and validation
// in a single call. It extracts the token from the request and validates it.
//
// Parameters:
//   - r: HTTP request containing the session token
//
// Returns:
//   - *SessionClaims: Validated session claims containing user information
//   - error: If no token is found or validation fails
func GetSessionClaims(r *http.Request) (*SessionClaims, error) {
	token, err := GetSessionToken(r)
	if err != nil {
		return nil, err
	}
	return ValidateToken(token)
}

// GetSessionUserID is a convenience function to extract just the user ID from a request.
// This is commonly used in handlers that only need the user ID for authorization
// or logging purposes without requiring the full session claims.
//
// Parameters:
//   - r: HTTP request containing the session token
//
// Returns:
//   - string: User ID from the session token
//   - error: If no token is found, validation fails, or user ID is empty
func GetSessionUserID(r *http.Request) (string, error) {
	claims, err := GetSessionClaims(r)
	if err != nil {
		return "", err
	}
	return claims.UserID, nil
}

// ClearCookie clears a browser cookie by setting it to expire immediately.
// This is used during logout to ensure the session cookie is removed from the client.
// The cookie is set with security attributes matching those used when setting it.
//
// Parameters:
//   - w: HTTP response writer to set the expired cookie on
//   - name: Name of the cookie to clear
func ClearCookie(w http.ResponseWriter, name string) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   true, // set to false if not using HTTPS in dev
		SameSite: http.SameSiteLaxMode,
	})
}
