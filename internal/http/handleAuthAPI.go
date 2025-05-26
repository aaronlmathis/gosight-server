// File: server/internal/http/handleAuthAPI.go
// Description: This file contains the API authentication handlers for the GoSight server that return JSON responses.

package httpserver

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	gosightauth "github.com/aaronlmathis/gosight-server/internal/auth"
	"github.com/aaronlmathis/gosight-server/internal/contextutil"
	"github.com/aaronlmathis/gosight-shared/model"
	"github.com/aaronlmathis/gosight-shared/utils"
)

// HandleAPILogin handles login requests from the SvelteKit frontend (JSON API)
func (s *HttpServer) HandleAPILogin(w http.ResponseWriter, r *http.Request) {
	// Set content type for JSON response
	w.Header().Set("Content-Type", "application/json")

	// Parse request body for credentials
	var credentials struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Invalid request body",
		})
		return
	}

	// Validate credentials are provided
	if credentials.Username == "" || credentials.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Username and password are required",
		})
		return
	}

	// Use the local auth provider to authenticate the user
	handler, ok := s.Sys.Auth["local"]
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Local authentication not configured",
		})
		return
	}

	// Create a new request with form data for the auth handler
	r.PostForm = map[string][]string{
		"username": {credentials.Username},
		"password": {credentials.Password},
	}

	// Authenticate the user using the existing auth handler
	user, err := handler.HandleCallback(w, r)
	if err != nil {
		// Emit login failure event
		s.Sys.Tele.Emitter.Emit(r.Context(), model.EventEntry{
			Timestamp: time.Now(),
			Type:      "user.login.failed",
			Level:     "warning",
			Category:  "auth",
			Source:    "auth:local",
			Scope:     "user",
			Target:    "",
			Message:   "Login failed via API",
			Meta: map[string]string{
				"provider":   "local",
				"ip":         utils.GetClientIP(r),
				"user_agent": r.UserAgent(),
				"reason":     "invalid credentials",
				"timestamp":  time.Now().Format(time.RFC3339),
			},
		})

		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Invalid username or password",
		})
		return
	}

	// Load full user with permissions
	user, err = s.Sys.Stores.Users.GetUserWithPermissions(r.Context(), user.ID)
	if err != nil {
		utils.Error("Failed to load roles for user %s: %v", user.Email, err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Failed to load user permissions",
		})
		return
	}

	// Check if user has MFA enabled
	if user.TOTPSecret != "" {
		// For API login with MFA, we return a special response indicating MFA is required
		// The frontend should then call the MFA verification endpoint
		gosightauth.SavePendingMFA(user.ID, w)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":      false,
			"mfa_required": true,
			"message":      "MFA verification required",
		})
		return
	}

	// Generate JWT token for successful login without MFA
	traceID, _ := contextutil.GetTraceID(r.Context())
	roles := gosightauth.ExtractRoleNames(user.Roles)
	token, err := gosightauth.GenerateToken(user.ID, roles, traceID)
	if err != nil {
		utils.Error("Failed to generate session token: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Failed to generate session token",
		})
		return
	}

	// Set session cookie
	gosightauth.SetSessionCookie(w, token)

	// Emit successful login event
	ctx := InjectUserContext(r.Context(), user)
	s.Sys.Tele.Emitter.Emit(ctx, model.EventEntry{
		Timestamp: time.Now(),
		Type:      "user.login.success",
		Level:     "info",
		Message:   fmt.Sprintf("User %s %s (%s) signed in via API", user.FirstName, user.LastName, user.Email),
		Target:    user.ID,
		Category:  "auth",
		Source:    "auth:local",
		Scope:     "user",
		Meta:      s.BuildAuthEventMeta(user, r),
	})

	// Return success response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Login successful",
		"user": map[string]interface{}{
			"id":          user.ID,
			"username":    user.Username,
			"email":       user.Email,
			"first_name":  user.FirstName,
			"last_name":   user.LastName,
			"permissions": gosightauth.FlattenPermissions(user.Roles),
		},
	})
}

// HandleAPIMFAVerify handles MFA verification requests from the SvelteKit frontend (JSON API)
func (s *HttpServer) HandleAPIMFAVerify(w http.ResponseWriter, r *http.Request) {
	// Set content type for JSON response
	w.Header().Set("Content-Type", "application/json")

	// Parse request body for MFA code
	var mfaRequest struct {
		Code     string `json:"code"`
		Remember bool   `json:"remember"`
	}

	if err := json.NewDecoder(r.Body).Decode(&mfaRequest); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Invalid request body",
		})
		return
	}

	// Validate MFA code is provided
	if mfaRequest.Code == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "MFA code is required",
		})
		return
	}

	// Load pending MFA user ID
	userID, err := gosightauth.LoadPendingMFA(r)
	if err != nil || userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "No pending MFA verification found",
		})
		return
	}

	// Get user and validate TOTP code
	user, err := s.Sys.Stores.Users.GetUserByID(r.Context(), userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Failed to load user data",
		})
		return
	}

	if !gosightauth.ValidateTOTP(user.TOTPSecret, mfaRequest.Code) {
		// Emit MFA failure event
		ctx := r.Context()
		s.Sys.Tele.Emitter.Emit(ctx, model.EventEntry{
			Timestamp: time.Now(),
			Type:      "user.mfa.failed",
			Level:     "warning",
			Message:   fmt.Sprintf("MFA failed for user ID %s via API", userID),
			Target:    user.ID,
			Category:  "auth",
			Source:    "auth:local",
			Scope:     "user",
			Meta:      s.BuildAuthEventMeta(user, r),
		})

		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Invalid MFA code",
		})
		return
	}

	// Load full user with permissions
	user, err = s.Sys.Stores.Users.GetUserWithPermissions(r.Context(), user.ID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Failed to load user permissions",
		})
		return
	}

	// Set remember MFA cookie if requested
	if mfaRequest.Remember {
		gosightauth.SetRememberMFA(w, user.ID, r)
	}

	// Generate JWT token for successful MFA verification
	traceID, _ := contextutil.GetTraceID(r.Context())
	roles := gosightauth.ExtractRoleNames(user.Roles)
	token, err := gosightauth.GenerateToken(user.ID, roles, traceID)
	if err != nil {
		utils.Error("Failed to generate session token: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Failed to generate session token",
		})
		return
	}

	// Set session cookie
	gosightauth.SetSessionCookie(w, token)

	// Clear pending MFA cookie
	gosightauth.ClearCookie(w, "pending_mfa")

	// Emit successful login event
	ctx := InjectUserContext(r.Context(), user)
	s.Sys.Tele.Emitter.Emit(ctx, model.EventEntry{
		Timestamp: time.Now(),
		Type:      "user.login.success",
		Level:     "info",
		Message:   fmt.Sprintf("User %s %s (%s) signed in via API after MFA", user.FirstName, user.LastName, user.Email),
		Target:    user.ID,
		Category:  "auth",
		Source:    "auth:local",
		Scope:     "user",
		Meta:      s.BuildAuthEventMeta(user, r),
	})

	// Return success response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "MFA verification successful",
		"user": map[string]interface{}{
			"id":          user.ID,
			"username":    user.Username,
			"email":       user.Email,
			"first_name":  user.FirstName,
			"last_name":   user.LastName,
			"permissions": gosightauth.FlattenPermissions(user.Roles),
		},
	})
}

// HandleAPILogout handles logout requests from the SvelteKit frontend (JSON API)
func (s *HttpServer) HandleAPILogout(w http.ResponseWriter, r *http.Request) {
	// Set content type for JSON response
	w.Header().Set("Content-Type", "application/json")

	ctx := r.Context()
	userID, _ := contextutil.GetUserID(ctx)

	// Emit logout event
	s.Sys.Tele.Emitter.Emit(ctx, model.EventEntry{
		Timestamp: time.Now(),
		Type:      "user.logout",
		Level:     "info",
		Category:  "auth",
		Source:    "auth:logout",
		Scope:     "user",
		Target:    userID,
		Message:   fmt.Sprintf("User %s logged out via API", userID),
		Meta: map[string]string{
			"user_id":    userID,
			"ip":         utils.GetClientIP(r),
			"user_agent": r.UserAgent(),
		},
	})

	// Clear session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "gosight_session",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   false, // TODO: Use config for HTTPS
	})

	// Return success response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Logout successful",
	})
}

// HandleCurrentUser handles requests to get the current authenticated user (JSON API)
func (s *HttpServer) HandleCurrentUser(w http.ResponseWriter, r *http.Request) {
	// Set content type for JSON response
	w.Header().Set("Content-Type", "application/json")

	ctx := r.Context()
	userID, ok := contextutil.GetUserID(ctx)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Not authenticated",
		})
		return
	}

	// Get user with permissions
	user, err := s.Sys.Stores.Users.GetUserWithPermissions(ctx, userID)
	if err != nil {
		utils.Error("Failed to load user %s: %v", userID, err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Failed to load user data",
		})
		return
	}

	// Get user profile information (optional - don't fail if it doesn't exist)
	profile, err := s.Sys.Stores.Users.GetUserProfile(ctx, userID)
	profileData := map[string]interface{}{}
	if err == nil && profile != nil {
		profileData = map[string]interface{}{
			"full_name":  profile.FullName,
			"phone":      profile.Phone,
			"avatar_url": profile.AvatarURL,
		}
	}

	// Return user data
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":          user.ID,
		"username":    user.Username,
		"email":       user.Email,
		"first_name":  user.FirstName,
		"last_name":   user.LastName,
		"permissions": gosightauth.FlattenPermissions(user.Roles),
		"profile":     profileData,
	})
}

// HandleAPIAuthProviders returns the list of available authentication providers (JSON API)
func (s *HttpServer) HandleAPIAuthProviders(w http.ResponseWriter, r *http.Request) {
	// Set content type for JSON response
	w.Header().Set("Content-Type", "application/json")

	// Get the configured auth providers
	providers := make([]map[string]interface{}, 0)

	for name := range s.Sys.Auth {
		provider := map[string]interface{}{
			"name":         name,
			"display_name": getProviderDisplayName(name),
		}
		providers = append(providers, provider)
	}

	// Return providers list
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"providers": providers,
	})
}

// getProviderDisplayName returns a user-friendly display name for the provider
func getProviderDisplayName(name string) string {
	switch name {
	case "local":
		return "Username/Password"
	case "google":
		return "Google"
	case "github":
		return "GitHub"
	case "azure":
		return "Microsoft"
	case "aws":
		return "AWS"
	default:
		return strings.Title(name)
	}
}
