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

package bootstrap

import (
	"fmt"

	gosightauth "github.com/aaronlmathis/gosight-server/internal/auth"
	"github.com/aaronlmathis/gosight-server/internal/config"
	"github.com/aaronlmathis/gosight-server/internal/store/userstore"
	"github.com/aaronlmathis/gosight-shared/utils"
)

// InitAuth initializes the complete authentication system for the GoSight server.
// This function sets up cryptographic secrets, multi-factor authentication,
// and all configured authentication providers to enable secure user access.
//
// The initialization process includes:
//
// 1. JWT Secret Setup:
//   - Decodes and validates the base64-encoded JWT signing secret
//   - Ensures the secret meets minimum security requirements (32+ bytes)
//   - Stores the secret for session token generation and validation
//
// 2. MFA Secret Setup:
//   - Initializes the multi-factor authentication encryption key
//   - Enables TOTP (Time-based One-Time Password) functionality
//   - Provides secure storage for MFA device registrations
//
// 3. Authentication Provider Configuration:
//   - Builds provider instances based on server configuration
//   - Configures OAuth settings for external providers (Google, GitHub, etc.)
//   - Sets up local authentication with password hashing
//   - Validates all provider configurations
//
// The function performs critical security validations and will fail if secrets
// are malformed or insufficient for production security requirements.
//
// Parameters:
//   - cfg: Server configuration containing authentication settings and secrets
//   - userStore: User storage interface for authentication data persistence
//
// Returns:
//   - map[string]gosightauth.AuthProvider: Configured authentication providers by name
//   - error: If secret initialization or provider configuration fails
func InitAuth(cfg *config.Config, userStore userstore.UserStore) (map[string]gosightauth.AuthProvider, error) {
	// Decode and store JWTSecret
	err := gosightauth.InitJWTSecret(cfg.Auth.JWTSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to decode JWT secret: %v", err)
	}

	// Decode and store MFASecret and JWTSecret
	err = gosightauth.InitMFAKey(cfg.Auth.MFASecret)
	if err != nil {
		return nil, fmt.Errorf("failed to decode MFA secret: %v", err)
	}

	// Build Auth Providers
	authProviders, err := buildAuthProviders(cfg, userStore)
	if err != nil {
		return nil, fmt.Errorf("failed to build auth providers: %v", err)
	}

	return authProviders, nil
}

// buildAuthProviders constructs authentication provider instances based on configuration.
// This function creates and configures the actual AuthProvider implementations for
// each authentication method specified in the server configuration. Each provider
// is initialized with its specific OAuth configuration and user store access.
//
// Supported authentication providers:
//   - local: Username/password authentication with bcrypt password hashing
//   - google: Google OAuth 2.0 authentication
//   - github: GitHub OAuth 2.0 authentication
//   - azure: Microsoft Azure OAuth 2.0 authentication
//   - aws: Amazon Web Services OAuth 2.0 authentication
//
// OAuth providers use the ToOAuthConfig() helper methods to convert GoSight
// configuration structures into standard oauth2.Config objects. All providers
// receive access to the user store for user lookup and creation during the
// authentication callback process.
//
// The function validates that all configured providers are supported and
// returns an error if an unknown provider is specified in the configuration.
//
// Parameters:
//   - cfg: Server configuration containing auth provider settings and OAuth credentials
//   - store: User store interface for user data access during authentication
//
// Returns:
//   - map[string]gosightauth.AuthProvider: Map of provider names to configured instances
//   - error: If an unsupported provider is configured or initialization fails
func buildAuthProviders(cfg *config.Config, store userstore.UserStore) (map[string]gosightauth.AuthProvider, error) {
	providers := make(map[string]gosightauth.AuthProvider)

	for _, name := range cfg.Web.AuthProviders {
		switch name {
		case "local":
			providers["local"] = &gosightauth.LocalAuth{Store: store}

		case "google":
			providers["google"] = &gosightauth.GoogleAuth{
				OAuthConfig: cfg.Auth.Google.ToOAuthConfig(), // helper to build *oauth2.Config
				Store:       store,
			}
		case "github":
			providers["github"] = &gosightauth.GitHubAuth{
				OAuthConfig: cfg.Auth.GitHub.ToOAuthConfig(), // helper to build *oauth2.Config
				Store:       store,
			}
		case "azure":
			providers["azure"] = &gosightauth.AzureAuth{
				OAuthConfig: cfg.Auth.Azure.ToOAuthConfig(), // helper to build *oauth2.Config
				Store:       store,
			}
		case "aws":
			providers["aws"] = &gosightauth.AWSAuth{
				OAuthConfig: cfg.Auth.AWS.ToOAuthConfig(), // helper to build *oauth2.Config
				Store:       store,
			}
		default:
			return nil, fmt.Errorf("unsupported auth provider: %s", name)
		}
	}
	utils.Debug("Auth providers configured: %v", cfg.Web.AuthProviders)
	utils.Debug("Auth providers initialized: %v", providers)
	return providers, nil
}
