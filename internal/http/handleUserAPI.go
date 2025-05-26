// File: server/internal/http/handleUserAPI.go
// Description: This file contains the user profile and settings API handlers for the GoSight server.

package httpserver

import (
	"encoding/json"
	"net/http"
	"strings"

	"golang.org/x/crypto/bcrypt"

	gosightauth "github.com/aaronlmathis/gosight-server/internal/auth"
	"github.com/aaronlmathis/gosight-server/internal/contextutil"
	"github.com/aaronlmathis/gosight-server/internal/usermodel"
	"github.com/aaronlmathis/gosight-shared/utils"
)

// HandleUpdateUserProfile handles PUT /users/profile requests to update user profile
func (s *HttpServer) HandleUpdateUserProfile(w http.ResponseWriter, r *http.Request) {
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

	// Parse request body
	var profileRequest usermodel.ProfileUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&profileRequest); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Invalid request body",
		})
		return
	}

	// Get existing profile or create new one
	profile, err := s.Sys.Stores.Users.GetUserProfile(ctx, userID)
	if err != nil {
		utils.Error("Failed to get user profile %s: %v", userID, err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Failed to get user profile",
		})
		return
	}

	// Update profile fields
	profile.UserID = userID
	profile.FullName = profileRequest.FullName
	profile.Phone = profileRequest.Phone

	// Save or update the profile
	if profile.AvatarURL == "" && profile.FullName == "" && profile.Phone == "" {
		// This is a new profile
		err = s.Sys.Stores.Users.CreateUserProfile(ctx, profile)
	} else {
		err = s.Sys.Stores.Users.UpdateUserProfile(ctx, profile)
	}

	if err != nil {
		utils.Error("Failed to save user profile %s: %v", userID, err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Failed to save user profile",
		})
		return
	}

	// Return success response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Profile updated successfully",
		"profile": profile,
	})
}

// HandleUpdateUserPassword handles PUT /users/password requests to update user password
func (s *HttpServer) HandleUpdateUserPassword(w http.ResponseWriter, r *http.Request) {
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

	// Parse request body
	var passwordRequest usermodel.PasswordChangeRequest
	if err := json.NewDecoder(r.Body).Decode(&passwordRequest); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Invalid request body",
		})
		return
	}

	// Validate password requirements
	if passwordRequest.CurrentPassword == "" || passwordRequest.NewPassword == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Current password and new password are required",
		})
		return
	}

	if passwordRequest.NewPassword != passwordRequest.ConfirmPassword {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "New password and confirm password do not match",
		})
		return
	}

	// Get current user to verify current password
	user, err := s.Sys.Stores.Users.GetUserByID(ctx, userID)
	if err != nil {
		utils.Error("Failed to get user %s: %v", userID, err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Failed to get user data",
		})
		return
	}

	// Verify current password
	if !gosightauth.CheckPasswordHash(passwordRequest.CurrentPassword, user.PasswordHash) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Current password is incorrect",
		})
		return
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(passwordRequest.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		utils.Error("Failed to hash password: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Failed to process new password",
		})
		return
	}

	// Update password in database
	err = s.Sys.Stores.Users.UpdateUserPassword(ctx, userID, string(hashedPassword))
	if err != nil {
		utils.Error("Failed to update password for user %s: %v", userID, err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Failed to update password",
		})
		return
	}

	// Return success response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Password updated successfully",
	})
}

// HandleGetUserPreferences handles GET /users/preferences requests to get user settings
func (s *HttpServer) HandleGetUserPreferences(w http.ResponseWriter, r *http.Request) {
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

	// Get user settings
	settings, err := s.Sys.Stores.Users.GetUserSettings(ctx, userID)
	if err != nil {
		utils.Error("Failed to get user settings %s: %v", userID, err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Failed to get user preferences",
		})
		return
	}

	// Transform flattened settings back to nested structure for frontend
	response := make(map[string]interface{})
	notifications := make(map[string]interface{})
	dashboard := make(map[string]interface{})

	for key, rawValue := range settings {
		var value interface{}
		if err := json.Unmarshal(rawValue, &value); err != nil {
			utils.Warn("Failed to unmarshal setting %s: %v", key, err)
			continue
		}

		switch {
		case key == "theme":
			response["theme"] = value
		case strings.HasPrefix(key, "notifications."):
			subKey := strings.TrimPrefix(key, "notifications.")
			switch subKey {
			case "emailAlerts":
				notifications["email_alerts"] = value
			case "pushNotifications":
				notifications["push_alerts"] = value
			case "alertFrequency":
				notifications["alert_frequency"] = value
			}
		case strings.HasPrefix(key, "dashboard."):
			subKey := strings.TrimPrefix(key, "dashboard.")
			switch subKey {
			case "refreshInterval":
				dashboard["refresh_interval"] = value
			case "defaultTimeRange":
				dashboard["default_time_range"] = value
			case "showSystemMetrics":
				dashboard["show_system_metrics"] = value
			}
		}
	}

	// Only include nested objects if they have data
	if len(notifications) > 0 {
		response["notifications"] = notifications
	}
	if len(dashboard) > 0 {
		response["dashboard"] = dashboard
	}

	// Return structured settings
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// HandleUpdateUserPreferences handles PUT /users/preferences requests to update user settings
func (s *HttpServer) HandleUpdateUserPreferences(w http.ResponseWriter, r *http.Request) {
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

	// Parse request body
	var preferencesRequest usermodel.UserPreferencesRequest
	if err := json.NewDecoder(r.Body).Decode(&preferencesRequest); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Invalid request body",
		})
		return
	}

	// Flatten preferences into individual key-value pairs
	flatPreferences := make(map[string]interface{})

	// Add theme
	if preferencesRequest.Theme != "" {
		flatPreferences["theme"] = preferencesRequest.Theme
	}

	// Flatten notifications - handle detailed notification structure
	flatPreferences["notifications.emailAlerts"] = preferencesRequest.Notifications.EmailAlerts
	flatPreferences["notifications.pushNotifications"] = preferencesRequest.Notifications.PushAlerts
	flatPreferences["notifications.alertFrequency"] = preferencesRequest.Notifications.AlertFrequency

	// Flatten dashboard settings
	if preferencesRequest.Dashboard != nil {
		for key, value := range preferencesRequest.Dashboard {
			switch key {
			case "refresh_interval":
				flatPreferences["dashboard.refreshInterval"] = value
			case "default_time_range":
				flatPreferences["dashboard.defaultTimeRange"] = value
			case "show_system_metrics":
				flatPreferences["dashboard.showSystemMetrics"] = value
			}
		}
	}

	// Save each flattened preference
	for key, value := range flatPreferences {
		if value != nil {
			valueBytes, err := json.Marshal(value)
			if err != nil {
				utils.Error("Failed to marshal preference %s: %v", key, err)
				continue
			}

			err = s.Sys.Stores.Users.SetUserSetting(ctx, userID, key, valueBytes)
			if err != nil {
				utils.Error("Failed to save preference %s for user %s: %v", key, userID, err)
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"error": "Failed to save preferences",
				})
				return
			}
		}
	}

	// Return success response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Preferences updated successfully",
	})
}

// HandleGetCompleteUser handles GET /users/me requests to get complete user data with profile and settings
func (s *HttpServer) HandleGetCompleteUser(w http.ResponseWriter, r *http.Request) {
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

	// Get complete user data
	completeUser, err := s.Sys.Stores.Users.GetCompleteUser(ctx, userID)
	if err != nil {
		utils.Error("Failed to get complete user data %s: %v", userID, err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Failed to get user data",
		})
		return
	}

	// Return user data (excluding sensitive info)
	response := map[string]interface{}{
		"id":          completeUser.User.ID,
		"username":    completeUser.User.Username,
		"email":       completeUser.User.Email,
		"first_name":  completeUser.User.FirstName,
		"last_name":   completeUser.User.LastName,
		"permissions": gosightauth.FlattenPermissions(completeUser.User.Roles),
		"profile":     completeUser.Profile,
		"settings":    completeUser.Settings,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
