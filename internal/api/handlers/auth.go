// File: server/internal/api/handlers/auth.go
// Description: This file contains the authentication handlers for the GoSight server.

package handlers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	gosightauth "github.com/aaronlmathis/gosight-server/internal/auth"
	"github.com/aaronlmathis/gosight-server/internal/contextutil"
	"github.com/aaronlmathis/gosight-server/internal/sys"
	"github.com/aaronlmathis/gosight-server/internal/usermodel"
	"github.com/aaronlmathis/gosight-shared/model"
	"github.com/aaronlmathis/gosight-shared/utils"
)

// AuthHandler provides handlers for authentication API endpoints
type AuthHandler struct {
	Sys *sys.SystemContext
}

// NewAuthHandler creates a new AuthHandler
func NewAuthHandler(sys *sys.SystemContext) *AuthHandler {
	return &AuthHandler{
		Sys: sys,
	}
}

// HandleAPILogin handles login requests from the SvelteKit frontend (JSON API)
func (h *AuthHandler) HandleAPILogin(w http.ResponseWriter, r *http.Request) {
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
	handler, ok := h.Sys.Auth["local"]
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
		h.Sys.Tele.Emitter.Emit(r.Context(), model.EventEntry{
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
	user, err = h.Sys.Stores.Users.GetUserWithPermissions(r.Context(), user.ID)
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
	ctx := injectUserContext(r.Context(), user)
	h.Sys.Tele.Emitter.Emit(ctx, model.EventEntry{
		Timestamp: time.Now(),
		Type:      "user.login.success",
		Level:     "info",
		Message:   fmt.Sprintf("User %s %s (%s) signed in via API", user.FirstName, user.LastName, user.Email),
		Target:    user.ID,
		Category:  "auth",
		Source:    "auth:local",
		Scope:     "user",
		Meta:      h.buildAuthEventMeta(user, r),
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
func (h *AuthHandler) HandleAPIMFAVerify(w http.ResponseWriter, r *http.Request) {
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
	user, err := h.Sys.Stores.Users.GetUserByID(r.Context(), userID)
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
		h.Sys.Tele.Emitter.Emit(ctx, model.EventEntry{
			Timestamp: time.Now(),
			Type:      "user.mfa.failed",
			Level:     "warning",
			Message:   fmt.Sprintf("MFA failed for user ID %s via API", userID),
			Target:    user.ID,
			Category:  "auth",
			Source:    "auth:local",
			Scope:     "user",
			Meta:      h.buildAuthEventMeta(user, r),
		})

		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Invalid MFA code",
		})
		return
	}

	// Load full user with permissions
	user, err = h.Sys.Stores.Users.GetUserWithPermissions(r.Context(), user.ID)
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
	ctx := injectUserContext(r.Context(), user)
	h.Sys.Tele.Emitter.Emit(ctx, model.EventEntry{
		Timestamp: time.Now(),
		Type:      "user.login.success",
		Level:     "info",
		Message:   fmt.Sprintf("User %s %s (%s) signed in via API after MFA", user.FirstName, user.LastName, user.Email),
		Target:    user.ID,
		Category:  "auth",
		Source:    "auth:local",
		Scope:     "user",
		Meta:      h.buildAuthEventMeta(user, r),
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
func (h *AuthHandler) HandleAPILogout(w http.ResponseWriter, r *http.Request) {
	// Set content type for JSON response
	w.Header().Set("Content-Type", "application/json")

	ctx := r.Context()
	userID, _ := contextutil.GetUserID(ctx)

	// Emit logout event
	h.Sys.Tele.Emitter.Emit(ctx, model.EventEntry{
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
func (h *AuthHandler) HandleCurrentUser(w http.ResponseWriter, r *http.Request) {
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
	user, err := h.Sys.Stores.Users.GetUserWithPermissions(ctx, userID)
	if err != nil {
		utils.Error("Failed to load user %s: %v", userID, err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Failed to load user data",
		})
		return
	}

	// Get user profile information (optional - don't fail if it doesn't exist)
	profile, err := h.Sys.Stores.Users.GetUserProfile(ctx, userID)
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
func (h *AuthHandler) HandleAPIAuthProviders(w http.ResponseWriter, r *http.Request) {
	// Set content type for JSON response
	w.Header().Set("Content-Type", "application/json")

	// Get the configured auth providers
	providers := make([]map[string]interface{}, 0)

	for name := range h.Sys.Auth {
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

// HandleLoginStart handles the login/start route for various auth providers
func (h *AuthHandler) HandleLoginStart(w http.ResponseWriter, r *http.Request) {
	provider := r.URL.Query().Get("provider")
	if handler, ok := h.Sys.Auth[provider]; ok {
		handler.StartLogin(w, r)
		return
	}
	http.Error(w, "invalid provider", http.StatusBadRequest)
}

// HandleCallback handles the callback route for various auth providers
func (h *AuthHandler) HandleCallback(w http.ResponseWriter, r *http.Request) {
	provider := r.URL.Query().Get("provider")

	handler, ok := h.Sys.Auth[provider]

	if !ok {
		// For SvelteKit frontend, redirect to login with error parameter
		http.Redirect(w, r, "/auth/login?error=invalid_provider", http.StatusSeeOther)
		return
	}

	user, err := handler.HandleCallback(w, r)
	if err != nil {
		h.Sys.Tele.Emitter.Emit(r.Context(), model.EventEntry{
			Timestamp: time.Now(),
			Type:      "user.login.failed",
			Level:     "warning",
			Category:  "auth",
			Source:    fmt.Sprintf("auth:%s", provider),
			Scope:     "user",
			Target:    "", // no user yet
			Message:   fmt.Sprintf("Login failed via %s", provider),
			Meta: map[string]string{
				"provider":   provider,
				"ip":         utils.GetClientIP(r),
				"user_agent": r.UserAgent(),
				"reason":     "SSO callback error",
				"timestamp":  time.Now().Format(time.RFC3339),
			},
		})
		h.setFlash(w, "Invalid username or password")
		http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
		return
	}

	user, err = h.Sys.Stores.Users.GetUserWithPermissions(r.Context(), user.ID)
	if err != nil {
		utils.Error("Failed to load roles for user %s: %v", user.Email, err)
		// For SvelteKit frontend, redirect to login with error parameter
		http.Redirect(w, r, "/auth/login?error=user_load_failed", http.StatusSeeOther)
		return
	}

	// Handle 2FA for local accounts
	if provider == "local" && user.TOTPSecret != "" {
		if gosightauth.CheckRememberMFA(r, user.ID) {
			ctx := h.injectUserContext(r.Context(), user)
			h.createFinalSessionAndRedirect(w, r.WithContext(ctx), user)

			// Always emit login success event here (centralized)
			h.Sys.Tele.Emitter.Emit(r.Context(), model.EventEntry{
				Timestamp: time.Now(),
				Type:      "user.login.success",
				Level:     "info",
				Message:   fmt.Sprintf("User %s %s (%s) signed in", user.FirstName, user.LastName, user.Email),
				Target:    user.ID,
				Category:  "auth",
				Source:    "auth.session",
				Scope:     "user",
				Meta:      h.buildAuthEventMeta(user, r),
			})
			return
		}

		gosightauth.SavePendingMFA(user.ID, w)

		// Optional: persist 'next' state
		if state := r.URL.Query().Get("state"); state != "" {
			http.SetCookie(w, &http.Cookie{
				Name:     "pending_next",
				Value:    state,
				Path:     "/",
				MaxAge:   300,
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
				Secure:   true, // TODO: Use config for HTTPS
			})
		}

		http.Redirect(w, r, "/auth/mfa", http.StatusSeeOther)
		return
	}

	// Normal Login flow for SSO provider login
	ctx := h.injectUserContext(r.Context(), user)
	utils.Debug("%s user: %s", provider, user.Email)
	utils.Debug("Roles: %v", user.Roles)
	utils.Debug("Permissions: %v", gosightauth.FlattenPermissions(user.Roles))

	// Set final session information and redirect user.
	h.createFinalSessionAndRedirect(w, r.WithContext(ctx), user)

	// Build Auth Event Entry
	eventEntry := model.EventEntry{
		Timestamp: time.Now(),
		Type:      "user.login.success",
		Level:     "info",
		Message:   fmt.Sprintf("User %s %s (%s) signed in", user.FirstName, user.LastName, user.Email),
		Target:    user.ID,
		Category:  "auth",
		Source:    fmt.Sprintf("auth:%s", provider),
		Scope:     "user",
		Meta:      h.buildAuthEventMeta(user, r),
	}
	// Emit the auth event
	h.Sys.Tele.Emitter.Emit(ctx, eventEntry)
}

// Helper methods for HandleCallback

// setFlash sets a flash message cookie
func (h *AuthHandler) setFlash(w http.ResponseWriter, message string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "flash",
		Value:    message,
		Path:     "/",
		MaxAge:   300,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   false, // TODO: Use config for HTTPS
	})
}

// injectUserContext injects user information into the request context
func (h *AuthHandler) injectUserContext(ctx context.Context, user *usermodel.User) context.Context {
	ctx = contextutil.SetUserID(ctx, user.ID)
	ctx = contextutil.SetUserRoles(ctx, gosightauth.ExtractRoleNames(user.Roles))
	ctx = contextutil.SetUserPermissions(ctx, gosightauth.FlattenPermissions(user.Roles))
	return ctx
}

// createFinalSessionAndRedirect creates the final session token and redirects the user
func (h *AuthHandler) createFinalSessionAndRedirect(w http.ResponseWriter, r *http.Request, user *usermodel.User) {
	traceID, _ := contextutil.GetTraceID(r.Context())
	roles := gosightauth.ExtractRoleNames(user.Roles)

	token, err := gosightauth.GenerateToken(user.ID, roles, traceID)
	if err != nil {
		utils.Error("Failed to generate session token: %v", err)
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	gosightauth.SetSessionCookie(w, token)

	// Clear any pending MFA/session leftovers
	gosightauth.ClearCookie(w, "pending_mfa")
	http.SetCookie(w, &http.Cookie{
		Name:     "pending_next",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
	})

	// handle redirect after login
	next := "/"
	if state := r.URL.Query().Get("state"); state != "" {
		if decoded, err := base64.URLEncoding.DecodeString(state); err == nil {
			next = string(decoded)
		}
	}
	if c, err := r.Cookie("pending_next"); err == nil && c.Value != "" {
		next = c.Value
	}

	// Validate that the next URL is safe and local
	next = strings.ReplaceAll(next, "\\", "/") // Normalize backslashes
	if !strings.HasPrefix(next, "/") || strings.HasPrefix(next, "//") ||
		strings.HasPrefix(next, "/\\") || strings.Contains(next, "..") {
		next = "/"
	}

	http.Redirect(w, r, next, http.StatusSeeOther)
}

// Utility functions

// injectUserContext injects user context from usermodel.User
func injectUserContext(ctx context.Context, user *usermodel.User) context.Context {
	ctx = contextutil.SetUserID(ctx, user.ID)
	ctx = contextutil.SetUserRoles(ctx, gosightauth.ExtractRoleNames(user.Roles))
	ctx = contextutil.SetUserPermissions(ctx, gosightauth.FlattenPermissions(user.Roles))
	return ctx
}

// buildAuthEventMeta builds authentication event metadata
func (h *AuthHandler) buildAuthEventMeta(user *usermodel.User, r *http.Request) map[string]string {
	return map[string]string{
		"user_id":    user.ID,
		"username":   user.Username,
		"email":      user.Email,
		"ip":         utils.GetClientIP(r),
		"user_agent": r.UserAgent(),
		"timestamp":  time.Now().Format(time.RFC3339),
	}
}

// getProviderDisplayName returns a user-friendly display name for the provider
func getProviderDisplayName(name string) string {
	switch name {
	case "local":
		return "Local Account"
	case "google":
		return "Google"
	case "github":
		return "GitHub"
	case "azure":
		return "Microsoft Azure"
	case "aws":
		return "Amazon Web Services"
	default:
		return name
	}
}
